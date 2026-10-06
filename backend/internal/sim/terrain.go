package sim

import (
	"math"
	"math/rand/v2"

	"miniv2/backend/internal/world"
)

// terrain is the simulation's view of a map: what blocks movement and where
// the water is. What grows on it lives in the ecology. Only fresh water
// (rivers and lakes, see world.FreshWater) can be drunk.
//
// blocked means "not dry, walkable ground": houses, fields, food and the
// ecology only ever use dry ground. How a body may still get onto a blocked
// tile (wading, swimming, climbing) is the mobility layer, rough.
type terrain struct {
	w, h      int
	blocked   []bool
	rough     []mobility // per blocked tile: wade, deep, climb or never (see mobilityAt)
	deepReach []uint8    // per deep-water tile: tiles to the nearest place to stand or wade
	water     []bool
	fresh     []bool // drinkable water
	nearWater []bool // next to any water (fishing)
	nearFresh []bool // next to fresh water (drinking)
	walkable  []int32
	shore     []int32 // walkable tiles next to water
	riverbank []int32 // walkable tiles next to fresh water

	nav    *navGraph   // the coarse route layer (navigation.go), built on first use
	search *navScratch // working memory for local route searches
}

// mobility is how a body can be on a tile, after RAGE's per-ped navigation
// capabilities (Peds/NavCapabilities.h: may enter water, may climb): walk on
// dry ground, wade in the shallows, swim or float out of one's depth, climb
// steep ground and boulders, or never (walls, trees, a crater). The zero
// value is "never", so a tile a test or an edit marks blocked is solid.
type mobility uint8

const (
	mobNever mobility = iota
	mobWalk
	mobWade
	mobDeep
	mobClimb
)

// roughOf classifies a tile that is not dry, walkable ground. A boulder can
// be scrambled over wherever it lies (except in a crater's lake); trees and
// walls are never passable; shallow water is waded, deep water swum and
// mountain ground climbed.
func roughOf(ground, object int) mobility {
	switch {
	case object == world.Boulder && ground != world.Crater:
		return mobClimb
	case world.ObjectTiles[object].Solid:
		return mobNever
	case ground == world.Water:
		return mobWade
	case ground == world.DeepWater:
		return mobDeep
	case ground == world.Mountain:
		return mobClimb
	}
	return mobNever
}

// mobilityAt is how a body can be on tile i.
func (t *terrain) mobilityAt(i int) mobility {
	if !t.blocked[i] {
		return mobWalk
	}
	if t.rough == nil {
		return mobNever
	}
	return t.rough[i]
}

func newTerrain(m *world.Map) *terrain {
	n := m.Width * m.Height
	t := &terrain{
		w:         m.Width,
		h:         m.Height,
		blocked:   make([]bool, n),
		rough:     make([]mobility, n),
		water:     make([]bool, n),
		nearWater: make([]bool, n),
		nearFresh: make([]bool, n),
	}
	for i := range n {
		g, o := m.Layers.Ground[i], m.Layers.Objects[i]
		t.water[i] = g == world.Water || g == world.DeepWater
		t.blocked[i] = world.GroundTiles[g].Solid || world.ObjectTiles[o].Solid
		if t.blocked[i] {
			t.rough[i] = roughOf(g, o)
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
	t.deepReach = t.measureDeep()
	return t
}

// measureDeep finds, for every deep-water tile, how many tiles it lies from
// the nearest tile where a body can stand or wade (255 at most): a swimmer
// only routes across narrow stretches of deep water (see navigation.go).
func (t *terrain) measureDeep() []uint8 {
	reach := make([]uint8, len(t.blocked))
	var queue []int
	for i := range reach {
		switch t.mobilityAt(i) {
		case mobDeep:
			reach[i] = 255
		case mobWalk, mobWade, mobClimb:
			queue = append(queue, i)
		}
	}
	for len(queue) > 0 {
		i := queue[0]
		queue = queue[1:]
		x, y := i%t.w, i/t.w
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				j, ok := t.index(x+dx, y+dy)
				if !ok || reach[j] != 255 || t.mobilityAt(j) != mobDeep {
					continue
				}
				reach[j] = min(254, reach[i]+1)
				queue = append(queue, j)
			}
		}
	}
	return reach
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
	w, h int
	// cell is the side of a cell in tiles (gridCell if unset).
	cell  float64
	cells [][]*Creature
}

const gridCell = 3.0

func (g *grid) size() float64 {
	if g.cell == 0 {
		return gridCell
	}
	return g.cell
}

func (g *grid) rebuild(mapW, mapH int, cs []*Creature) {
	w, h := int(math.Ceil(float64(mapW)/g.size())), int(math.Ceil(float64(mapH)/g.size()))
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
	cx := min(g.w-1, max(0, int(x/g.size())))
	cy := min(g.h-1, max(0, int(y/g.size())))
	return cy*g.w + cx
}

func (g *grid) near(x, y, r float64, fn func(*Creature)) {
	x0, x1 := max(0, int((x-r)/g.size())), min(g.w-1, int((x+r)/g.size()))
	y0, y1 := max(0, int((y-r)/g.size())), min(g.h-1, int((y+r)/g.size()))
	for cy := y0; cy <= y1; cy++ {
		for cx := x0; cx <= x1; cx++ {
			for _, c := range g.cells[cy*g.w+cx] {
				fn(c)
			}
		}
	}
}
