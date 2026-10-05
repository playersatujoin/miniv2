package world

import (
	"container/heap"
	"math"
	"math/rand/v2"
	"slices"
)

type GenOptions struct {
	Name   string
	Width  int
	Height int
	Seed   uint32
}

// Generate builds an island map deterministically from opts.Seed.
// The caller is responsible for ID and timestamps.
func Generate(opts GenOptions) *Map {
	w, h := opts.Width, opts.Height
	m := &Map{
		Name:     opts.Name,
		Width:    w,
		Height:   h,
		TileSize: DefaultTileSize,
		Seed:     opts.Seed,
		Layers:   Layers{Ground: make([]int, w*h), Objects: make([]int, w*h)},
	}
	rng := rand.New(rand.NewPCG(uint64(opts.Seed), 0x9e3779b97f4a7c15))

	elev, moist := terrainFields(w, h, uint64(opts.Seed))
	paintTerrain(m, elev, moist)
	geo := buildGeoModel(w, h, uint64(opts.Seed), elev, moist)
	applyGeology(m, geo)
	scatterObjects(m, elev, rng)
	rims := craterRims(m, geo)
	m.Spawn = findSpawn(m)
	clearAround(m, m.Spawn, 2)
	carvePaths(m, rng, rims)
	connectRegions(m)
	return m
}

// applyGeology shapes the terrain after the geological story: crater lakes
// and young volcanic cones, karst outcrops, salt flats and river channels.
func applyGeology(m *Map, g *GeoModel) {
	w := m.Width
	for _, v := range g.Volcanoes {
		g.disk(v.X, v.Y, v.Apron, func(x, y int, d float64) {
			i := y*w + x
			if !g.Land(i) {
				return
			}
			if d <= v.Crater {
				m.Layers.Ground[i] = Crater
				return
			}
			switch m.Layers.Ground[i] {
			case Mountain, StoneFloor, Grass, ForestFloor, Dirt:
				m.Layers.Ground[i] = VolcanicRock
			}
		})
	}
	for i, r := range g.Rock {
		x, y := i%w, i/w
		switch r {
		case RockLimestone:
			switch m.Layers.Ground[i] {
			case StoneFloor:
				m.Layers.Ground[i] = Limestone
			case Grass, ForestFloor:
				if hash2(uint64(m.Seed)^0x11ce, x, y) < 0.45 {
					m.Layers.Ground[i] = Limestone
				}
			}
		case RockEvaporite:
			switch m.Layers.Ground[i] {
			case Grass, ForestFloor, StoneFloor:
				m.Layers.Ground[i] = Sand // salt flat
			}
		}
	}
	for i, r := range g.RiverOf {
		if r >= 0 && m.Layers.Ground[i] != Crater {
			m.Layers.Ground[i] = Water
		}
	}
}

// craterRims clears the ground just outside each crater and returns one rim
// tile per volcano, so a trail can lead up to it.
func craterRims(m *Map, g *GeoModel) []Point {
	var rims []Point
	for _, v := range g.Volcanoes {
		var rim *Point
		g.disk(v.X, v.Y, v.Crater+1.6, func(x, y int, d float64) {
			i := y*m.Width + x
			if d <= v.Crater || groundSolid(m.Layers.Ground[i]) {
				return
			}
			m.Layers.Objects[i] = None
			if rim == nil {
				rim = &Point{x, y}
			}
		})
		if rim != nil {
			rims = append(rims, *rim)
		}
	}
	return rims
}

// connectRegions links every sizeable walkable area to the spawn's with a
// trail (bridging rivers where needed), so no one is stranded by a river.
func connectRegions(m *Map) {
	for range 16 {
		comp, sizes := walkableComponents(m)
		home := comp[m.Spawn.Y*m.Width+m.Spawn.X]
		target, best := -1, math.Inf(1)
		for i, c := range comp {
			if c < 0 || c == home || sizes[c] < 12 {
				continue
			}
			if d := dist(Point{i % m.Width, i / m.Width}, m.Spawn); d < best {
				target, best = i, d
			}
		}
		if target < 0 {
			return
		}
		path := findPath(m, m.Spawn, Point{target % m.Width, target / m.Width})
		if path == nil {
			return
		}
		for _, p := range path {
			pave(m, p)
		}
	}
}

func walkableComponents(m *Map) (comp []int, sizes []int) {
	comp = make([]int, m.Width*m.Height)
	for i := range comp {
		comp[i] = -1
	}
	for start := range comp {
		if comp[start] >= 0 || !m.Walkable(start%m.Width, start/m.Width) {
			continue
		}
		id := len(sizes)
		sizes = append(sizes, 0)
		stack := []int{start}
		comp[start] = id
		for len(stack) > 0 {
			i := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			sizes[id]++
			x, y := i%m.Width, i/m.Width
			for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+d[0], y+d[1]
				if n := ny*m.Width + nx; m.inBounds(nx, ny) && comp[n] < 0 && m.Walkable(nx, ny) {
					comp[n] = id
					stack = append(stack, n)
				}
			}
		}
	}
	return comp, sizes
}

func terrainFields(w, h int, seed uint64) (elev, moist []float64) {
	en := noise2D{seed}
	mn := noise2D{seed ^ 0xA5A5A5A55A5A5A5A}
	elev = make([]float64, w*h)
	moist = make([]float64, w*h)
	for y := range h {
		for x := range w {
			// Square-bump falloff: 0 in the centre, 1 at the edges, so the map is an island.
			nx := 2*float64(x)/float64(w-1) - 1
			ny := 2*float64(y)/float64(h-1) - 1
			d := 1 - (1-nx*nx)*(1-ny*ny)

			i := y*w + x
			elev[i] = en.fbm(float64(x)/28, float64(y)/28, 5) - 0.75*d
			moist[i] = mn.fbm(float64(x)/18, float64(y)/18, 4)
		}
	}
	return elev, moist
}

func quantiles(vals []float64, ps ...float64) []float64 {
	sorted := slices.Clone(vals)
	slices.Sort(sorted)
	out := make([]float64, len(ps))
	for i, p := range ps {
		out[i] = sorted[int(p*float64(len(sorted)-1))]
	}
	return out
}

// paintTerrain thresholds by quantile rather than absolute height, so every
// seed gets roughly the same land/water balance.
func paintTerrain(m *Map, elev, moist []float64) {
	t := quantiles(elev, 0.20, 0.32, 0.37, 0.87, 0.94)
	forest := quantiles(moist, 0.62)[0]
	for i, e := range elev {
		var g int
		switch {
		case e < t[0]:
			g = DeepWater
		case e < t[1]:
			g = Water
		case e < t[2]:
			g = Sand
		case e < t[3]:
			g = Grass
			if moist[i] > forest {
				g = ForestFloor
			}
		case e < t[4]:
			g = StoneFloor
		default:
			g = Mountain
		}
		m.Layers.Ground[i] = g
	}
}

func scatterObjects(m *Map, elev []float64, rng *rand.Rand) {
	pineLine := quantiles(elev, 0.75)[0]
	for i, g := range m.Layers.Ground {
		r := rng.Float64()
		obj := None
		switch g {
		case ForestFloor:
			switch {
			case r < 0.32:
				obj = Tree
				if elev[i] > pineLine {
					obj = Pine
				}
			case r < 0.38:
				obj = Bush
			}
		case Grass:
			switch {
			case r < 0.03:
				obj = Tree
			case r < 0.055:
				obj = Bush
			case r < 0.10:
				obj = Flowers
			}
		case Sand:
			if r < 0.015 {
				obj = Boulder
			}
		case StoneFloor:
			if r < 0.09 {
				obj = Boulder
			}
		case VolcanicRock:
			switch {
			case r < 0.04:
				obj = Pine
			case r < 0.07:
				obj = Boulder
			}
		case Limestone:
			switch {
			case r < 0.05:
				obj = Tree
			case r < 0.11:
				obj = Bush
			case r < 0.15:
				obj = Boulder
			}
		}
		m.Layers.Objects[i] = obj
	}
}

// findSpawn picks the open tile closest to the centre of the map.
func findSpawn(m *Map) Point {
	cx, cy := m.Width/2, m.Height/2
	best, bestDist := Point{cx, cy}, math.MaxInt
	for y := range m.Height {
		for x := range m.Width {
			g := m.Layers.Ground[y*m.Width+x]
			if g != Grass && g != Sand && g != Dirt {
				continue
			}
			if d := (x-cx)*(x-cx) + (y-cy)*(y-cy); d < bestDist {
				best, bestDist = Point{x, y}, d
			}
		}
	}
	if bestDist == math.MaxInt {
		m.Layers.Ground[cy*m.Width+cx] = Grass
	}
	return best
}

func clearAround(m *Map, c Point, r int) {
	for y := c.Y - r; y <= c.Y+r; y++ {
		for x := c.X - r; x <= c.X+r; x++ {
			if m.inBounds(x, y) {
				m.Layers.Objects[y*m.Width+x] = None
			}
		}
	}
}

// carvePaths connects the spawn to a few landmark clearings, and to the extra
// targets (crater rims), with dirt roads, bridging water and tunnelling
// through mountains where needed.
func carvePaths(m *Map, rng *rand.Rand, extra []Point) {
	const landmarks = 4
	minDist := float64(min(m.Width, m.Height)) / 4

	var targets []Point
	for attempt := 0; attempt < 400 && len(targets) < landmarks; attempt++ {
		p := Point{rng.IntN(m.Width), rng.IntN(m.Height)}
		g := m.Layers.Ground[p.Y*m.Width+p.X]
		if g != Grass && g != ForestFloor && g != Sand {
			continue
		}
		if dist(p, m.Spawn) < minDist || slices.ContainsFunc(targets, func(q Point) bool { return dist(p, q) < minDist }) {
			continue
		}
		targets = append(targets, p)
	}

	for _, t := range targets {
		for _, p := range findPath(m, m.Spawn, t) {
			pave(m, p)
		}
		plaza(m, t)
	}
	for _, t := range extra {
		for _, p := range findPath(m, m.Spawn, t) {
			pave(m, p)
		}
	}
}

func dist(a, b Point) float64 { return math.Hypot(float64(a.X-b.X), float64(a.Y-b.Y)) }

func pave(m *Map, p Point) {
	i := p.Y*m.Width + p.X
	switch m.Layers.Ground[i] {
	case Water, DeepWater:
		m.Layers.Ground[i] = Bridge
	case Mountain:
		m.Layers.Ground[i] = StoneFloor
	case Bridge, StoneFloor, VolcanicRock, Limestone, Crater:
	default:
		m.Layers.Ground[i] = Dirt
	}
	m.Layers.Objects[i] = None
}

// plaza lays a small stone-centred clearing around a landmark.
func plaza(m *Map, c Point) {
	for y := c.Y - 2; y <= c.Y+2; y++ {
		for x := c.X - 2; x <= c.X+2; x++ {
			if !m.inBounds(x, y) {
				continue
			}
			i := y*m.Width + x
			if g := m.Layers.Ground[i]; groundSolid(g) || g == Bridge {
				continue
			}
			d := dist(Point{x, y}, c)
			switch {
			case d <= 1:
				m.Layers.Ground[i] = StoneFloor
			case d <= 2.3:
				m.Layers.Ground[i] = Dirt
			default:
				continue
			}
			m.Layers.Objects[i] = None
		}
	}
}

// stepCost is the A* cost of entering a tile. Existing roads are cheap so
// later paths merge into earlier ones; a little per-tile jitter makes roads wander.
func stepCost(m *Map, x, y int) float64 {
	i := y*m.Width + x
	var c float64
	switch m.Layers.Ground[i] {
	case Dirt, Bridge:
		c = 0.6
	case ForestFloor:
		c = 1.6
	case VolcanicRock:
		c = 1.3
	case Water:
		c = 5
	case Mountain:
		c = 9
	case DeepWater:
		c = 14
	default:
		c = 1
	}
	if objectSolid(m.Layers.Objects[i]) {
		c += 1
	}
	return c + 0.5*hash2(uint64(m.Seed)^0x51ab, x, y)
}

type pqNode struct {
	i int
	f float64
}

type pq []pqNode

func (q pq) Len() int           { return len(q) }
func (q pq) Less(a, b int) bool { return q[a].f < q[b].f }
func (q pq) Swap(a, b int)      { q[a], q[b] = q[b], q[a] }
func (q *pq) Push(x any)        { *q = append(*q, x.(pqNode)) }
func (q *pq) Pop() any {
	old := *q
	n := old[len(old)-1]
	*q = old[:len(old)-1]
	return n
}

func findPath(m *Map, from, to Point) []Point {
	w, h := m.Width, m.Height
	start, goal := from.Y*w+from.X, to.Y*w+to.X

	g := make([]float64, w*h)
	for i := range g {
		g[i] = math.Inf(1)
	}
	prev := make([]int, w*h)
	g[start] = 0

	// 0.6 is the cheapest step, so this heuristic never overestimates.
	heur := func(i int) float64 {
		x, y := i%w, i/w
		return 0.6 * float64(abs(x-to.X)+abs(y-to.Y))
	}

	open := &pq{{start, heur(start)}}
	dirs := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for open.Len() > 0 {
		cur := heap.Pop(open).(pqNode)
		if cur.i == goal {
			break
		}
		if cur.f > g[cur.i]+heur(cur.i)+1e-9 {
			continue // stale entry
		}
		cx, cy := cur.i%w, cur.i/w
		for _, d := range dirs {
			nx, ny := cx+d[0], cy+d[1]
			if nx < 0 || ny < 0 || nx >= w || ny >= h {
				continue
			}
			ni := ny*w + nx
			if m.Layers.Ground[ni] == Crater {
				continue
			}
			if ng := g[cur.i] + stepCost(m, nx, ny); ng < g[ni] {
				g[ni] = ng
				prev[ni] = cur.i
				heap.Push(open, pqNode{ni, ng + heur(ni)})
			}
		}
	}
	if math.IsInf(g[goal], 1) {
		return nil
	}

	var path []Point
	for i := goal; i != start; i = prev[i] {
		path = append(path, Point{i % w, i / w})
	}
	return append(path, from)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
