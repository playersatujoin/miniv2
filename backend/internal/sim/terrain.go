package sim

import (
	"math"
	"math/rand/v2"

	"miniv2/backend/internal/world"
)

// terrain is the simulation's view of a map: what blocks movement and where
// the water is. What grows on it lives in the ecology. Only fresh water
// (rivers and lakes, see world.FreshWater) can be drunk.
type terrain struct {
	w, h      int
	blocked   []bool
	water     []bool
	fresh     []bool // drinkable water
	nearWater []bool // next to any water (fishing)
	nearFresh []bool // next to fresh water (drinking)
	walkable  []int32
	shore     []int32 // walkable tiles next to water
	riverbank []int32 // walkable tiles next to fresh water
}

func newTerrain(m *world.Map) *terrain {
	n := m.Width * m.Height
	t := &terrain{
		w:         m.Width,
		h:         m.Height,
		blocked:   make([]bool, n),
		water:     make([]bool, n),
		nearWater: make([]bool, n),
		nearFresh: make([]bool, n),
	}
	for i := range n {
		g, o := m.Layers.Ground[i], m.Layers.Objects[i]
		t.water[i] = g == world.Water || g == world.DeepWater
		t.blocked[i] = world.GroundTiles[g].Solid || world.ObjectTiles[o].Solid
		if t.blocked[i] {
			continue
		}
		t.walkable = append(t.walkable, int32(i))
	}

	t.fresh = world.FreshWater(m)
	for _, i := range t.walkable {
		x, y := int(i)%t.w, int(i)/t.w
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if j, ok := t.index(x+dx, y+dy); ok && t.water[j] {
					t.nearWater[i] = true
					t.nearFresh[i] = t.nearFresh[i] || t.fresh[j]
				}
			}
		}
		if t.nearWater[i] {
			t.shore = append(t.shore, i)
		}
		if t.nearFresh[i] {
			t.riverbank = append(t.riverbank, i)
		}
	}
	return t
}

func (t *terrain) index(x, y int) (int, bool) {
	if x < 0 || y < 0 || x >= t.w || y >= t.h {
		return 0, false
	}
	return y*t.w + x, true
}

func (t *terrain) indexAt(x, y float64) (int, bool) {
	return t.index(int(math.Floor(x)), int(math.Floor(y)))
}

func (t *terrain) blockedAt(x, y float64) bool {
	i, ok := t.indexAt(x, y)
	return !ok || t.blocked[i]
}

// randomSpot returns a point on a random walkable tile, preferring the banks
// of rivers and lakes, then the shore.
func (t *terrain) randomSpot(rng *rand.Rand) (x, y float64, ok bool) {
	pool := t.riverbank
	if len(pool) == 0 {
		pool = t.shore
	}
	if len(pool) == 0 {
		pool = t.walkable
	}
	if len(pool) == 0 {
		return 0, 0, false
	}
	i := int(pool[rng.IntN(len(pool))])
	return float64(i%t.w) + 0.3 + rng.Float64()*0.4, float64(i/t.w) + 0.3 + rng.Float64()*0.4, true
}

// spotNear returns a walkable point within radius tiles of (cx, cy), or false.
func (t *terrain) spotNear(rng *rand.Rand, cx, cy, radius float64) (x, y float64, ok bool) {
	for range 50 {
		x = cx + (rng.Float64()*2-1)*radius
		y = cy + (rng.Float64()*2-1)*radius
		if !t.blockedAt(x, y) {
			return x, y, true
		}
	}
	return 0, 0, false
}

// nearestFree finds the closest walkable tile centre to (x, y) by BFS.
func (t *terrain) nearestFree(x, y float64) (float64, float64, bool) {
	start, ok := t.indexAt(x, y)
	if !ok {
		start = t.clampIndex(x, y)
	}
	seen := make([]bool, len(t.blocked))
	queue := []int{start}
	seen[start] = true
	for len(queue) > 0 {
		i := queue[0]
		queue = queue[1:]
		if !t.blocked[i] {
			return float64(i%t.w) + 0.5, float64(i/t.w) + 0.5, true
		}
		cx, cy := i%t.w, i/t.w
		for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			if j, ok := t.index(cx+d[0], cy+d[1]); ok && !seen[j] {
				seen[j] = true
				queue = append(queue, j)
			}
		}
	}
	return 0, 0, false
}

func (t *terrain) clampIndex(x, y float64) int {
	cx := int(clamp(math.Floor(x), 0, float64(t.w-1)))
	cy := int(clamp(math.Floor(y), 0, float64(t.h-1)))
	return cy*t.w + cx
}

// grid is a uniform spatial hash for neighbour queries.
type grid struct {
	w, h  int
	cells [][]*Creature
}

const gridCell = 3.0

func (g *grid) rebuild(mapW, mapH int, cs []*Creature) {
	w, h := int(math.Ceil(float64(mapW)/gridCell)), int(math.Ceil(float64(mapH)/gridCell))
	if g.w != w || g.h != h {
		g.w, g.h = w, h
		g.cells = make([][]*Creature, w*h)
	}
	for i := range g.cells {
		clear(g.cells[i])
		g.cells[i] = g.cells[i][:0]
	}
	for _, c := range cs {
		g.cells[g.cellOf(c.X, c.Y)] = append(g.cells[g.cellOf(c.X, c.Y)], c)
	}
}

func (g *grid) cellOf(x, y float64) int {
	cx := min(g.w-1, max(0, int(x/gridCell)))
	cy := min(g.h-1, max(0, int(y/gridCell)))
	return cy*g.w + cx
}

func (g *grid) near(x, y, r float64, fn func(*Creature)) {
	x0, x1 := max(0, int((x-r)/gridCell)), min(g.w-1, int((x+r)/gridCell))
	y0, y1 := max(0, int((y-r)/gridCell)), min(g.h-1, int((y+r)/gridCell))
	for cy := y0; cy <= y1; cy++ {
		for cx := x0; cx <= x1; cx++ {
			for _, c := range g.cells[cy*g.w+cx] {
				fn(c)
			}
		}
	}
}
