package sim

import (
	"math"
	"math/rand/v2"

	"miniv2/backend/internal/world"
)

// Food capacity per tile and regrowth per simulated second.
const (
	grassFood    = 0.5
	grassRegrow  = 0.003
	forestFood   = 0.7
	forestRegrow = 0.005
	flowerFood   = 0.3
	flowerRegrow = 0.006
	bushFood     = 1.0
	bushRegrow   = 0.012
)

// terrain is the simulation's view of a map: what blocks movement, where the
// water is and how much food each tile holds.
type terrain struct {
	w, h      int
	blocked   []bool
	water     []bool
	nearWater []bool
	foodCap   []float32
	foodRate  []float32
	food      []float32
	walkable  []int32
	shore     []int32 // walkable tiles next to water
}

func newTerrain(m *world.Map) *terrain {
	n := m.Width * m.Height
	t := &terrain{
		w:         m.Width,
		h:         m.Height,
		blocked:   make([]bool, n),
		water:     make([]bool, n),
		nearWater: make([]bool, n),
		foodCap:   make([]float32, n),
		foodRate:  make([]float32, n),
		food:      make([]float32, n),
	}
	for i := range n {
		g, o := m.Layers.Ground[i], m.Layers.Objects[i]
		t.water[i] = g == world.Water || g == world.DeepWater
		t.blocked[i] = world.GroundTiles[g].Solid || world.ObjectTiles[o].Solid
		if t.blocked[i] {
			continue
		}
		t.walkable = append(t.walkable, int32(i))

		var cap, rate float64
		switch g {
		case world.Grass:
			cap, rate = grassFood, grassRegrow
		case world.ForestFloor:
			cap, rate = forestFood, forestRegrow
		}
		switch o {
		case world.Bush:
			cap, rate = bushFood, bushRegrow
		case world.Flowers:
			cap, rate = cap+flowerFood, rate+flowerRegrow
		}
		t.foodCap[i], t.foodRate[i] = float32(cap), float32(rate)
	}
	copy(t.food, t.foodCap)

	for _, i := range t.walkable {
		x, y := int(i)%t.w, int(i)/t.w
		for dy := -1; dy <= 1 && !t.nearWater[i]; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if j, ok := t.index(x+dx, y+dy); ok && t.water[j] {
					t.nearWater[i] = true
					break
				}
			}
		}
		if t.nearWater[i] {
			t.shore = append(t.shore, i)
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

func (t *terrain) regrow(seconds float64) {
	for i, f := range t.food {
		if c := t.foodCap[i]; f < c {
			t.food[i] = min(c, f+t.foodRate[i]*float32(seconds))
		}
	}
}

// randomSpot returns a point on a random walkable tile, preferring the shore.
func (t *terrain) randomSpot(rng *rand.Rand) (x, y float64, ok bool) {
	pool := t.shore
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
