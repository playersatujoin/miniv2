package chem

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	"miniv2/backend/internal/world"
)

// Each tile has two deposit slots: one in the ground itself (bedrock, ore,
// soil, sand, water) and one for what grows or lies on it (a tree, a boulder,
// grass, seaweed).
const slotsPerTile = 2

// Inexhaustible deposits report this amount and never shrink.
const endless = 1000

// Starting amounts in item units, scaled roughly to real-world commonness:
// iron and coal plentiful, copper moderate, tin less, gold scarce.
var depositAmounts = map[ItemID]float32{
	"kayu": 8, "serat": 6, "rumput_laut": 4,
	"batu": 40, "pasir": 80, "tanah_liat": 40, "batu_kapur": 60, "dolomit": 40,
	"batu_bara": 50, "belerang": 20,
	"hematit": 40, "limonit": 30, "pasir_besi": 30, "laterit_nikel": 40, "bauksit": 40,
	"kalkopirit": 30, "malakit": 15, "tembaga_alam": 6,
	"kasiterit": 15, "galena": 20, "sfalerit": 20,
	"bijih_emas": 6,
	"gipsum":     30, "halit": 30, "silvit": 20, "boraks": 20, "selestin": 15,
}

const defaultMineralAmount = 15

// Renewable deposits regrow, in units per second: plants, and sulphur that
// fumaroles keep depositing on a crater rim.
var depositRegrow = map[ItemID]float32{"kayu": 0.02, "serat": 0.02, "rumput_laut": 0.01, "belerang": 0.02}

type slot struct {
	item  uint8 // index into items + 1; 0 = empty
	model uint8 // index into modelSpecs + 1
	tool  bool
}

// Geology knows which resources lie on which tiles. It follows the map's
// geological model (world.GeoModel), is deterministic for a map, and keeps
// remaining amounts, which the simulation persists.
type Geology struct {
	w, h    int
	geo     *world.GeoModel
	slots   []slot
	amounts []float32
	caps    []float32
	regrow  []float32
	growing []int32 // renewable deposits currently below their cap
	queued  []bool
}

func NewGeology(m *world.Map) *Geology {
	n := m.Width * m.Height * slotsPerTile
	g := &Geology{
		w:       m.Width,
		h:       m.Height,
		geo:     world.BuildGeoModel(m),
		slots:   make([]slot, n),
		amounts: make([]float32, n),
		caps:    make([]float32, n),
		regrow:  make([]float32, n),
		queued:  make([]bool, n),
	}
	l := newLay(m, g.geo)
	g.placeZones(l)
	g.placePlacers(l)
	g.placeSurface(l)
	g.relocate(l)
	return g
}

// GeoModel returns the geological model the deposits were placed from.
func (g *Geology) GeoModel() *world.GeoModel { return g.geo }

func (g *Geology) markGrowing(i int) {
	if g.regrow[i] > 0 && !g.queued[i] && g.amounts[i] < g.caps[i] {
		g.queued[i] = true
		g.growing = append(g.growing, int32(i))
	}
}

func (g *Geology) requeue() {
	g.growing = g.growing[:0]
	clear(g.queued)
	for i := range g.amounts {
		g.markGrowing(i)
	}
}

// set places a deposit of id in slot i; scale shrinks thin deposits (placers).
func (g *Geology) set(i int, id ItemID, model string, scale float32) {
	it := items[itemIndex[id]]
	amount := depositAmounts[id]
	if amount == 0 {
		amount = defaultMineralAmount
	}
	amount *= scale
	if id == "air" || id == "air_laut" {
		amount = endless
	}
	g.slots[i] = slot{
		item:  uint8(itemIndex[id] + 1),
		model: uint8(modelIndex[model] + 1),
		// Every mineral has to be dug out; so does quarried limestone.
		tool: it.Kind == kindMineral || id == "batu_kapur",
	}
	g.amounts[i] = amount
	g.caps[i] = amount
	g.regrow[i] = depositRegrow[id]
}

func (g *Geology) itemAt(i int) (ItemID, bool) {
	if s := g.slots[i]; s.item != 0 {
		return items[s.item-1].ID, true
	}
	return "", false
}

func (g *Geology) modelAt(i int) string {
	if s := g.slots[i]; s.model != 0 {
		return modelSpecs[s.model-1].Key
	}
	return ""
}

func (g *Geology) isMineral(i int) bool {
	id, ok := g.itemAt(i)
	return ok && items[itemIndex[id]].Kind == kindMineral
}

// lay bundles what deposit placement needs to know about the map.
type lay struct {
	m    *world.Map
	geo  *world.GeoModel
	seed uint64
	sea  []bool // sea water, as opposed to fresh rivers and lakes
}

func newLay(m *world.Map, geo *world.GeoModel) lay {
	l := lay{m: m, geo: geo, seed: uint64(m.Seed), sea: make([]bool, m.Width*m.Height)}
	// The sea is deep water plus the shallows connected to it, never a river channel.
	var stack []int
	for i, gr := range m.Layers.Ground {
		if gr == world.DeepWater {
			l.sea[i] = true
			stack = append(stack, i)
		}
	}
	for len(stack) > 0 {
		i := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		x, y := i%m.Width, i/m.Width
		for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := x+d[0], y+d[1]
			if !l.inside(nx, ny) {
				continue
			}
			j := ny*m.Width + nx
			if !l.sea[j] && m.Layers.Ground[j] == world.Water && geo.RiverOf[j] < 0 {
				l.sea[j] = true
				stack = append(stack, j)
			}
		}
	}
	return l
}

func (l lay) ground(x, y int) int  { return l.m.Layers.Ground[y*l.m.Width+x] }
func (l lay) object(x, y int) int  { return l.m.Layers.Objects[y*l.m.Width+x] }
func (l lay) inside(x, y int) bool { return x >= 0 && y >= 0 && x < l.m.Width && y < l.m.Height }

func (l lay) walkable(x, y int) bool {
	return l.inside(x, y) && !world.GroundTiles[l.ground(x, y)].Solid && !world.ObjectTiles[l.object(x, y)].Solid
}

// reachable reports whether a creature can gather from (x, y): standing on it
// or on one of its 4 neighbours.
func (l lay) reachable(x, y int) bool {
	return l.walkable(x, y) || l.walkable(x+1, y) || l.walkable(x-1, y) || l.walkable(x, y+1) || l.walkable(x, y-1)
}

func (l lay) vegetated(i int) bool {
	gr := l.m.Layers.Ground[i]
	return gr == world.Grass || gr == world.ForestFloor
}

func (l lay) rocky(i int) bool {
	switch l.m.Layers.Ground[i] {
	case world.Mountain, world.StoneFloor, world.VolcanicRock, world.Limestone:
		return true
	}
	return false
}

// distanceFrom returns every tile's 4-neighbour distance to the nearest tile
// matching from, capped at limit+1.
func (l lay) distanceFrom(from func(i int) bool, limit int) []int {
	n := l.m.Width * l.m.Height
	d := make([]int, n)
	var queue []int
	for i := range d {
		d[i] = limit + 1
		if from(i) {
			d[i] = 0
			queue = append(queue, i)
		}
	}
	for len(queue) > 0 {
		i := queue[0]
		queue = queue[1:]
		if d[i] >= limit {
			continue
		}
		x, y := i%l.m.Width, i/l.m.Width
		for _, dd := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := x+dd[0], y+dd[1]
			if j := ny*l.m.Width + nx; l.inside(nx, ny) && d[j] > d[i]+1 {
				d[j] = d[i] + 1
				queue = append(queue, j)
			}
		}
	}
	return d
}

// stripe reports whether (x, y) lies on one of a set of parallel, slightly
// wavy bands: ore veins, dykes or coal seams.
func stripe(seed uint64, x, y int, angle, period, width float64, salt int) bool {
	s := float64(x)*math.Cos(angle) + float64(y)*math.Sin(angle) + 1.5*valueNoise(seed, float64(x)/4, float64(y)/4, salt)
	f := s/period - math.Floor(s/period)
	return f < width
}

// classify decides which deposit model, if any, owns the ground slot of each
// land tile, following where each deposit type forms in nature.
func (g *Geology) classify(l lay) []int {
	geo := l.geo
	w := l.m.Width
	zone := make([]int, w*l.m.Height)
	for i := range zone {
		zone[i] = -1
	}
	rock := geo.Rock

	porphyry := func(i int) bool {
		p := geo.PlutonOf[i]
		return p >= 0 && geo.Plutons[p].Porphyry
	}
	granitic := func(i int) bool { return rock[i] == world.RockGranite || rock[i] == world.RockPegmatite }
	inPorphyry := l.distanceFrom(func(i int) bool { return !porphyry(i) }, 3)
	nearPorphyry := l.distanceFrom(porphyry, 3)
	nearGranite := l.distanceFrom(granitic, 3)
	nearLimestone := l.distanceFrom(func(i int) bool { return rock[i] == world.RockLimestone }, 2)
	nearArcRock := l.distanceFrom(func(i int) bool { return granitic(i) || rock[i] == world.RockVolcanic }, 4)
	nearCrater := l.distanceFrom(func(i int) bool { return l.m.Layers.Ground[i] == world.Crater }, 1)
	nearRiver := l.distanceFrom(func(i int) bool { return geo.RiverOf[i] >= 0 }, 1)

	epiReach := 3 + geo.Size()/32
	veinAngle := hash(l.seed, 1, 2, 401) * math.Pi
	tinAngle := hash(l.seed, 3, 4, 402) * math.Pi
	lowland := func(i int) bool { return geo.Elevation[i] < geo.HighLevel }

	for i := range zone {
		gr := l.m.Layers.Ground[i]
		if gr == world.Water || gr == world.DeepWater || gr == world.Crater || gr == world.Bridge {
			continue
		}
		x, y := i%w, i/w
		set := func(model string) { zone[i] = modelIndex[model] }

		if nearCrater[i] == 1 {
			set("solfatara")
			continue
		}
		switch rock[i] {
		case world.RockEvaporite:
			set("evaporit")
		case world.RockCarbonatite:
			set("karbonatit")
		case world.RockPegmatite:
			set("pegmatit")
		case world.RockUltramafic:
			if l.vegetated(i) && region(l.seed, x, y, 3, 411) > 0.2 {
				set("laterit_nikel")
			} else {
				set("ofiolit")
			}
		case world.RockGranite:
			p := geo.PlutonOf[i]
			switch {
			case porphyry(i) && inPorphyry[i] <= 2:
				set("porfiri")
			case nearLimestone[i] <= 1:
				set("skarn")
			case p >= 0 && geo.Plutons[p].Tin && stripe(l.seed, x, y, tinAngle, 3, 0.4, 421):
				set("granit_timah")
			case l.vegetated(i) && lowland(i) && region(l.seed, x, y, 5, 431) < 0.1:
				set("bauksit")
			}
		case world.RockLimestone:
			switch {
			case nearGranite[i] <= 2:
				set("skarn")
			case region(l.seed, x, y, 4, 441) < 0.15:
				set("mvt")
			default:
				set("karbonat")
			}
		case world.RockVolcanic:
			if geo.VolcanoAt(x, y) >= 0 {
				break // the young cone itself is fresh, unaltered rock
			}
			beyond, angle := math.Inf(1), veinAngle
			for k, v := range geo.Volcanoes {
				if d := math.Hypot(float64(x-v.X), float64(y-v.Y)) - v.Apron; d < beyond {
					beyond, angle = d, veinAngle+float64(k)*0.7
				}
			}
			switch {
			case nearPorphyry[i] <= 2:
				set("porfiri")
			case beyond <= epiReach && stripe(l.seed, x, y, angle, 4, 0.18, 451):
				set("epitermal")
			case region(l.seed, x, y, 3, 461) < 0.006:
				set("tembaga_alam")
			case l.vegetated(i) && lowland(i) && region(l.seed, x, y, 5, 431) < 0.08:
				set("bauksit")
			case l.vegetated(i) && region(l.seed, x, y, 4, 471) < 0.08:
				set("laterit_besi")
			}
		case world.RockSedimentary:
			seam := stripe(l.seed, x, y, geo.ArcAngle(), 6, 0.2, 481)
			// River cuts expose more of each seam.
			cut := nearRiver[i] <= 1 && stripe(l.seed, x, y, geo.ArcAngle(), 6, 0.35, 481)
			switch {
			case seam || cut:
				set("batubara")
			case nearArcRock[i] <= 4 && region(l.seed, x, y, 3, 491) < 0.12:
				set("uranium_sedimen")
			case l.vegetated(i) && region(l.seed, x, y, 4, 471) < 0.04:
				set("laterit_besi") // ironstone
			}
		case world.RockMetamorphic:
			switch {
			case region(l.seed, x, y, 3, 501) < 0.06:
				set("mangan")
			case l.vegetated(i) && region(l.seed, x, y, 4, 471) < 0.08:
				set("laterit_besi")
			}
		}
	}
	return zone
}

// placeZones fills each deposit model's zone. The first tiles of a zone
// (reachable ones first) get one of each mineral of the model's association,
// so everything that belongs there is present; the rest follow the
// association's proportions, with barren host rock in between.
func (g *Geology) placeZones(l lay) {
	zone := g.classify(l)
	w := l.m.Width
	byModel := make([][]int, len(modelSpecs))
	for i, z := range zone {
		if z >= 0 {
			byModel[z] = append(byModel[z], i)
		}
	}
	for mi, tiles := range byModel {
		spec := modelSpecs[mi]
		if len(spec.minerals) == 0 || len(tiles) == 0 {
			continue
		}
		l.sortReachableFirst(tiles, 600+mi)
		listFor := func(i int) []weighted {
			if spec.surface != nil && l.walkable(i%w, i/w) {
				return spec.surface
			}
			return spec.minerals
		}
		done := map[int]bool{}
		// Rarest first, so a cramped zone still holds its rare minerals; the
		// common ones fill the rest anyway.
		wants := slices.Concat(spec.surface, spec.minerals)
		slices.SortStableFunc(wants, func(a, b weighted) int { return cmp.Compare(a.weight, b.weight) })
		for _, want := range wants {
			if g.hasIn(tiles, want.item) {
				continue
			}
			for _, i := range tiles {
				if done[i] || !slices.ContainsFunc(listFor(i), func(c weighted) bool { return c.item == want.item }) {
					continue
				}
				g.set(i*slotsPerTile, want.item, spec.Key, 1)
				done[i] = true
				break
			}
		}
		for _, i := range tiles {
			if done[i] {
				continue
			}
			x, y := i%w, i/w
			if spec.barren > 0 && hash(l.seed, x, y, 700+mi) < spec.barren {
				g.set(i*slotsPerTile, spec.host, spec.Key, 1)
				continue
			}
			g.set(i*slotsPerTile, pick(listFor(i), hash(l.seed, x, y, 800+mi)), spec.Key, 1)
		}
	}

	// Everything else gets soil, sand, stone or water according to the ground.
	for i, gr := range l.m.Layers.Ground {
		si := i * slotsPerTile
		if g.slots[si].item != 0 {
			continue
		}
		x, y := i%w, i/w
		rock := l.geo.Rock[i]
		switch {
		case gr == world.DeepWater || (gr == world.Water && l.sea[i]):
			g.set(si, "air_laut", "laut", 1)
		case gr == world.Water:
			g.set(si, "air", "air", 1)
		case gr == world.Crater || gr == world.Bridge:
		case gr == world.Sand:
			g.set(si, "pasir", "pasir", 1)
		case gr == world.Dirt:
			g.set(si, "tanah_liat", "lempung", 1)
		case rock == world.RockLimestone && l.rocky(i):
			g.set(si, "batu_kapur", "karbonat", 1)
		case l.rocky(i):
			g.set(si, "batu", "batu", 1)
		case rock == world.RockAlluvium,
			rock == world.RockVolcanic && hash(l.seed, x, y, 901) < 0.3,
			rock == world.RockSedimentary && hash(l.seed, x, y, 902) < 0.2:
			g.set(si, "tanah_liat", "lempung", 1)
		}
	}
}

// sortReachableFirst orders tiles so reachable ones come first, each group in
// a stable scattered order.
func (l lay) sortReachableFirst(tiles []int, salt int) {
	w := l.m.Width
	slices.SortFunc(tiles, func(a, b int) int {
		ra, rb := l.reachable(a%w, a/w), l.reachable(b%w, b/w)
		if ra != rb {
			if ra {
				return -1
			}
			return 1
		}
		return cmp.Compare(hash(l.seed, a%w, a/w, salt), hash(l.seed, b%w, b/w, salt))
	})
}

// hasIn reports whether any of tiles already holds item in its ground slot.
func (g *Geology) hasIn(tiles []int, item ItemID) bool {
	for _, i := range tiles {
		if id, ok := g.itemAt(i * slotsPerTile); ok && id == item {
			return true
		}
	}
	return false
}

// placePlacers lays heavy minerals along each river below the rocks it has
// drained so far, on its delta, and on the beaches beside its mouth.
func (g *Geology) placePlacers(l lay) {
	geo := l.geo
	w := l.m.Width
	loose := func(i int) bool {
		switch l.m.Layers.Ground[i] {
		case world.Sand, world.Dirt, world.Grass, world.ForestFloor:
			return !g.isMineral(i * slotsPerTile) // never bury a primary ore
		}
		return false
	}
	const (
		fromGold = 1 << iota
		fromGranite
		fromVolcanic
	)
	poolFor := func(mask int) []weighted {
		var out []weighted
		if mask&fromGold != 0 {
			out = append(out, goldPlacers...)
		}
		if mask&fromGranite != 0 {
			out = append(out, granitePlacers...)
		}
		if mask&fromVolcanic != 0 {
			out = append(out, volcanicPlacers...)
		}
		return out
	}
	type spot struct{ i, mask int }
	for r, path := range geo.Rivers {
		mask := 0
		var spots []spot
		seen := map[int]bool{}
		addSpot := func(i, mask int, p float64, salt int) {
			if !seen[i] && loose(i) && hash(l.seed, i%w, i/w, salt) < p {
				seen[i] = true
				spots = append(spots, spot{i, mask})
			}
		}
		for k, p := range path {
			for dy := -3; dy <= 3; dy++ {
				for dx := -3; dx <= 3; dx++ {
					x, y := p.X+dx, p.Y+dy
					if !l.inside(x, y) {
						continue
					}
					j := y*w + x
					switch m := g.modelAt(j * slotsPerTile); {
					case m == "porfiri" || m == "epitermal":
						mask |= fromGold
					case geo.Rock[j] == world.RockGranite || geo.Rock[j] == world.RockPegmatite:
						mask |= fromGranite
					case geo.Rock[j] == world.RockVolcanic:
						mask |= fromVolcanic
					}
				}
			}
			if k < 2 || mask == 0 {
				continue
			}
			// A bar carries only what the river has drained so far.
			for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				if x, y := p.X+d[0], p.Y+d[1]; l.inside(x, y) && geo.RiverOf[y*w+x] < 0 {
					addSpot(y*w+x, mask, 0.4, 1000+r)
				}
			}
		}
		if mask == 0 {
			continue
		}

		mouth := path[0]
		for _, q := range path {
			if geo.RiverOf[q.Y*w+q.X] == int16(r) {
				mouth = q
			}
		}
		reach := 3 + geo.Size()/64
		far := 2.5 * reach
		for dy := -int(far); dy <= int(far); dy++ {
			for dx := -int(far); dx <= int(far); dx++ {
				x, y := mouth.X+dx, mouth.Y+dy
				if !l.inside(x, y) {
					continue
				}
				i := y*w + x
				switch d := math.Hypot(float64(dx), float64(dy)); {
				case d <= reach:
					addSpot(i, mask, 0.45, 1100+r)
				case d <= far && l.m.Layers.Ground[i] == world.Sand:
					addSpot(i, mask, 0.3, 1200+r)
				}
			}
		}
		slices.SortStableFunc(spots, func(a, b spot) int {
			ra, rb := l.reachable(a.i%w, a.i/w), l.reachable(b.i%w, b.i/w)
			if ra != rb {
				if ra {
					return -1
				}
				return 1
			}
			return cmp.Compare(hash(l.seed, a.i%w, a.i/w, 1300+r), hash(l.seed, b.i%w, b.i/w, 1300+r))
		})
		placed := map[ItemID]bool{}
		for _, s := range spots {
			pool := poolFor(s.mask)
			id := pick(pool, hash(l.seed, s.i%w, s.i/w, 1400+r))
			for _, c := range pool {
				if !placed[c.item] {
					id = c.item // one of each first
					break
				}
			}
			placed[id] = true
			g.set(s.i*slotsPerTile, id, "plaser", 0.8)
		}
	}
}

// placeSurface fills the second slot: trees, boulders, grass and seaweed.
func (g *Geology) placeSurface(l lay) {
	w := l.m.Width
	nearLimestone := l.distanceFrom(func(i int) bool { return l.geo.Rock[i] == world.RockLimestone }, 3)
	for i := range l.m.Layers.Ground {
		si := i*slotsPerTile + 1
		switch l.m.Layers.Objects[i] {
		case world.Tree, world.Pine:
			g.set(si, "kayu", "kayu", 1)
			continue
		case world.Boulder:
			g.set(si, "batu", "batu", 1)
			continue
		case world.Bush, world.Flowers:
			g.set(si, "serat", "serat", 1)
			continue
		}
		switch {
		case l.vegetated(i):
			g.set(si, "serat", "serat", 1)
		case l.m.Layers.Ground[i] == world.Water && l.sea[i]:
			// Seaweed grows on shallow coasts, most richly on reef flats.
			p := 0.2
			if nearLimestone[i] <= 3 {
				p = 0.45
			}
			if hash(l.seed, i%w, i/w, 41) < p {
				g.set(si, "rumput_laut", "laut", 1)
			}
		}
	}
}

// relocate handles hand-edited maps: a mineral whose every deposit became
// unreachable moves to reachable tiles of compatible host rock. If no such
// rock is left, the mineral is simply absent from this map.
func (g *Geology) relocate(l lay) {
	w := l.m.Width
	reachable := map[ItemID]int{}
	for i, s := range g.slots {
		t := i / slotsPerTile
		if s.item != 0 && l.reachable(t%w, t/w) {
			reachable[items[s.item-1].ID]++
		}
	}
	for _, it := range items {
		if it.Kind != kindMineral || reachable[it.ID] > 0 {
			continue
		}
		placed := 0
		for _, mi := range hostModels(it.ID) {
			spec := modelSpecs[mi]
			var cands []int
			for i, r := range l.geo.Rock {
				if slices.Contains(spec.rocks, r) && l.reachable(i%w, i/w) && !g.isMineral(i*slotsPerTile) {
					cands = append(cands, i)
				}
			}
			l.sortReachableFirst(cands, 1500)
			for _, i := range cands[:min(2-placed, len(cands))] {
				g.set(i*slotsPerTile, it.ID, spec.Key, 1)
				placed++
			}
			if placed >= 2 {
				break
			}
		}
	}
}

func (g *Geology) inside(x, y int) bool { return x >= 0 && y >= 0 && x < g.w && y < g.h }

// Sources lists what a creature standing on tile (x, y) can gather: from its
// own tile and the 4 neighbours (ore from an adjacent mountain face, wood from
// an adjacent tree, water from the shore, …). Air is available everywhere.
// Food is not included; the simulation handles food on tiles itself.
func (g *Geology) Sources(x, y int) []Source {
	var out []Source
	for _, d := range [5][2]int{{0, 0}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		tx, ty := x+d[0], y+d[1]
		if !g.inside(tx, ty) {
			continue
		}
		base := (ty*g.w + tx) * slotsPerTile
		for s := range slotsPerTile {
			sl := g.slots[base+s]
			if sl.item == 0 || g.amounts[base+s] < 0.01 {
				continue
			}
			out = append(out, Source{
				Item: items[sl.item-1].ID, X: tx, Y: ty,
				Amount:    float64(g.amounts[base+s]),
				NeedsTool: sl.tool,
				Model:     g.modelAt(base + s),
			})
		}
	}
	if g.inside(x, y) {
		out = append(out, Source{Item: "udara", X: x, Y: y, Amount: endless})
	}
	return out
}

// Deposit is one resource deposit on a tile, for maps and overlays.
type Deposit struct {
	X, Y      int
	Item      ItemID
	Amount    float64
	NeedsTool bool
	// Surface is true for what grows or lies on the tile (a tree, a boulder,
	// grass, seaweed) rather than the ground itself.
	Surface bool
	Model   string // deposit model key, see DepositModels
}

// Deposits lists every deposit that still has something left, in tile order.
func (g *Geology) Deposits() []Deposit {
	var out []Deposit
	for i, sl := range g.slots {
		if sl.item == 0 || g.amounts[i] < 0.01 {
			continue
		}
		t := i / slotsPerTile
		out = append(out, Deposit{
			X: t % g.w, Y: t / g.w,
			Item:      items[sl.item-1].ID,
			Amount:    float64(g.amounts[i]),
			NeedsTool: sl.tool,
			Surface:   i%slotsPerTile == 1,
			Model:     g.modelAt(i),
		})
	}
	return out
}

// MinedOut lists tiles whose ground deposit of a mineral has been dug out:
// less than 1% of what was there is left. Renewables never count.
func (g *Geology) MinedOut() [][2]int {
	var out [][2]int
	for t := 0; t < g.w*g.h; t++ {
		i := t * slotsPerTile
		if g.caps[i] > 0 && g.regrow[i] == 0 && g.isMineral(i) && g.amounts[i] < g.caps[i]*0.01 {
			out = append(out, [2]int{t % g.w, t / g.w})
		}
	}
	return out
}

// Take removes up to amount of item from the deposit at (x, y) and reports how much was taken.
func (g *Geology) Take(x, y int, item ItemID, amount float64) float64 {
	if !g.inside(x, y) || amount <= 0 {
		return 0
	}
	if item == "udara" {
		return amount
	}
	base := (y*g.w + x) * slotsPerTile
	for s := range slotsPerTile {
		i := base + s
		sl := g.slots[i]
		if sl.item == 0 || items[sl.item-1].ID != item {
			continue
		}
		if g.caps[i] >= endless {
			return amount
		}
		took := min(amount, float64(g.amounts[i]))
		g.amounts[i] -= float32(took)
		g.markGrowing(i)
		return took
	}
	return 0
}

// Regrow replenishes renewable deposits (wood, fibre, seaweed, sulphur) by seconds of growth.
func (g *Geology) Regrow(seconds float64) {
	keep := g.growing[:0]
	for _, i := range g.growing {
		a := min(g.caps[i], g.amounts[i]+g.regrow[i]*float32(seconds))
		g.amounts[i] = a
		if a < g.caps[i] {
			keep = append(keep, i)
		} else {
			g.queued[i] = false
		}
	}
	g.growing = keep
}

// Update re-derives deposits after the map was edited, keeping remaining amounts where possible.
func (g *Geology) Update(m *world.Map) {
	next := NewGeology(m)
	if next.w == g.w && next.h == g.h {
		for i, s := range next.slots {
			if s.item != 0 && s.item == g.slots[i].item {
				next.amounts[i] = min(g.amounts[i], next.caps[i])
			}
		}
		next.requeue()
	}
	*g = *next
}

// Amounts and SetAmounts expose the remaining amounts for persistence.
func (g *Geology) Amounts() []float32 { return slices.Clone(g.amounts) }

func (g *Geology) SetAmounts(a []float32) error {
	if len(a) != len(g.amounts) {
		return fmt.Errorf("geology has %d deposits, got %d amounts", len(g.amounts), len(a))
	}
	for i, v := range a {
		if math.IsNaN(float64(v)) || v < 0 {
			return fmt.Errorf("invalid amount %v at deposit %d", v, i)
		}
		g.amounts[i] = min(v, g.caps[i])
	}
	g.requeue()
	return nil
}

// pick chooses from a weighted list with r in [0, 1).
func pick(list []weighted, r float64) ItemID {
	total := 0.0
	for _, w := range list {
		total += w.weight
	}
	r *= total
	for _, w := range list {
		if r < w.weight {
			return w.item
		}
		r -= w.weight
	}
	return list[len(list)-1].item
}

// region returns a stable random number shared by tiles in the same
// irregular patch of roughly size×size tiles, so deposits form bodies.
func region(seed uint64, x, y int, size float64, salt int) float64 {
	fx, fy := float64(x), float64(y)
	wx := fx + 2.5*size/4*(valueNoise(seed, fx/5, fy/5, salt+1)*2-1)
	wy := fy + 2.5*size/4*(valueNoise(seed, fx/5, fy/5, salt+2)*2-1)
	return hash(seed, int(math.Floor(wx/size)), int(math.Floor(wy/size)), salt)
}

// hash maps a tile coordinate to [0, 1), stable for a seed and salt.
func hash(seed uint64, x, y, salt int) float64 {
	h := seed*0x9E3779B97F4A7C15 ^ uint64(int64(x))*0xBF58476D1CE4E5B9 ^ uint64(int64(y))*0x94D049BB133111EB ^ uint64(int64(salt))*0xD6E8FEB86659FD93
	h ^= h >> 30
	h *= 0xBF58476D1CE4E5B9
	h ^= h >> 27
	h *= 0x94D049BB133111EB
	h ^= h >> 31
	return float64(h>>11) / float64(1<<53)
}

func valueNoise(seed uint64, x, y float64, salt int) float64 {
	x0, y0 := math.Floor(x), math.Floor(y)
	ix, iy := int(x0), int(y0)
	sx, sy := x-x0, y-y0
	sx, sy = sx*sx*(3-2*sx), sy*sy*(3-2*sy)
	lerp := func(a, b, t float64) float64 { return a + (b-a)*t }
	top := lerp(hash(seed, ix, iy, salt), hash(seed, ix+1, iy, salt), sx)
	bottom := lerp(hash(seed, ix, iy+1, salt), hash(seed, ix+1, iy+1, salt), sx)
	return lerp(top, bottom, sy)
}
