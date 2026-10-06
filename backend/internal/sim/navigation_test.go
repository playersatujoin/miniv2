package sim

import (
	"container/heap"
	"math"
	"math/rand/v2"
	"slices"
	"testing"

	"miniv2/backend/internal/world"
)

// crosses reports whether a path enters ground of medium m.
func crosses(s *Sim, path []Waypoint, m mobility) bool {
	for _, p := range path {
		if i, ok := s.terrain.indexAt(p.X, p.Y); ok && s.terrain.mobilityAt(i) == m {
			return true
		}
	}
	return false
}

// A wall of rock from north to south (x = 16) with openings: routes take dry
// land over a ford when it is about as short, a ford over a slope, and
// nobody is routed where their body can't go.
func TestRoutesPreferLandThenWadingThenClimbing(t *testing.T) {
	wall := func(openings map[int][2]int) func(set func(x, y, g, o int)) {
		return func(set func(x, y, g, o int)) {
			for y := 2; y < 30; y++ {
				set(16, y, world.Grass, world.Wall)
			}
			for y, tile := range openings {
				set(16, y, tile[0], tile[1])
			}
		}
	}
	ford := [2]int{world.Water, world.None}
	slope := [2]int{world.Mountain, world.None}
	land := [2]int{world.Grass, world.None}
	route := func(s *Sim, c *Creature) []Waypoint {
		return s.searchPath(s.routeCapsFor(c), c.X, c.Y, 20.5, 15.5, 12)
	}

	// A ford on the straight line, dry land a tile aside.
	s, _ := mobilityWorld(t, 32, wall(map[int][2]int{15: ford, 16: land}))
	c := person(s, Male, "A", 12, 15)
	if p := route(s, c); len(p) == 0 || crosses(s, p, mobWade) {
		t.Fatalf("the route should keep its feet dry: %v", p)
	}

	// A ford and a slope, each three tiles off the straight line.
	s, _ = mobilityWorld(t, 32, wall(map[int][2]int{12: ford, 18: slope}))
	c = person(s, Male, "A", 12, 15)
	p := route(s, c)
	if len(p) == 0 || !crosses(s, p, mobWade) || crosses(s, p, mobClimb) {
		t.Fatalf("the route should wade rather than climb: %v", p)
	}
	if p[len(p)-1] != (Waypoint{20.5, 15.5}) {
		t.Fatal("the route did not arrive")
	}

	// Only a slope: grown-ups climb, a four-year-old can't, nor can anyone
	// in a world without mobility.
	s, _ = mobilityWorld(t, 32, wall(map[int][2]int{18: slope}))
	c = person(s, Male, "A", 12, 15)
	if p := route(s, c); len(p) == 0 || !crosses(s, p, mobClimb) {
		t.Fatal("a grown-up should climb when there is no other way")
	}
	kid := aged(s, person(s, Female, "Kid", 12, 15), 4)
	if p := route(s, kid); len(p) != 0 {
		t.Fatal("a small child was routed up a slope")
	}
	s.opts.NoMobility = true
	if p := s.routePath(c, 20.5, 15.5, 12); len(p) != 0 {
		t.Fatal("without mobility the slope is solid")
	}
}

// Deep water: a swimmer is routed across a narrow channel, not across a
// wide one, and anyone with a raft goes where they like.
func TestRoutesSwimOnlyShortGaps(t *testing.T) {
	channel := func(width int) func(set func(x, y, g, o int)) {
		return func(set func(x, y, g, o int)) {
			for y := 0; y < 40; y++ {
				for x := 18; x < 18+width; x++ {
					set(x, y, world.DeepWater, world.None)
				}
			}
		}
	}
	s, _ := mobilityWorld(t, 40, channel(3))
	c := person(s, Male, "A", 15, 20)
	if p := s.routePath(c, 23.5, 20.5, 12); len(p) == 0 || !crosses(s, p, mobDeep) {
		t.Fatal("a fit swimmer should cross a narrow channel")
	}
	s, _ = mobilityWorld(t, 40, channel(8))
	c = person(s, Male, "A", 13, 20)
	if p := s.routePath(c, 28.5, 20.5, 16); len(p) != 0 && crosses(s, p, mobDeep) {
		t.Fatal("a swimmer was routed across a wide channel")
	}
	c.Inventory.add(raftItem, 1)
	if p := s.routePath(c, 28.5, 20.5, 16); len(p) == 0 || p[len(p)-1] != (Waypoint{28.5, 20.5}) {
		t.Fatal("a rafter should cross the wide channel")
	}
}

// A long wall with its only gap far to the east: a person following the
// coarse route reaches remembered water on the far side; without it the
// local search runs into the wall and stays there.
func TestCoarseRouteReachesRememberedWaterAroundAWall(t *testing.T) {
	layout := func(set func(x, y, g, o int)) {
		for x := 0; x < 50; x++ {
			set(x, 24, world.Grass, world.Wall)
		}
		for x := 50; x < 54; x++ {
			set(x, 24, world.Grass, world.None)
		}
	}
	walk := func(s *Sim, c *Creature, goal Waypoint, radius int) (int, bool) {
		for i := 1; i <= 60; i++ {
			p := s.routePath(c, goal.X, goal.Y, radius)
			if len(p) == 0 {
				return i, false
			}
			last := p[len(p)-1]
			c.X, c.Y = last.X, last.Y
			if math.Hypot(c.X-goal.X, c.Y-goal.Y) < 1 {
				return i, true
			}
		}
		return 60, false
	}
	water := Waypoint{10.5, 16.5}

	s, _ := mobilityWorld(t, 64, layout)
	c := person(s, Female, "Pulang", 10, 32)
	if n, ok := walk(s, c, water, 8); !ok {
		t.Fatalf("did not reach the water around the wall (at %.1f, %.1f after %d plans)", c.X, c.Y, n)
	}

	s, _ = mobilityWorld(t, 64, layout)
	s.opts.NoMobility = true
	c = person(s, Female, "Pulang", 10, 32)
	if _, ok := walk(s, c, water, 8); ok || c.Y < 24 {
		t.Fatal("the plain local search should not find the far gap")
	}
}

func TestCoarseGraphLinksEveryReachablePatch(t *testing.T) {
	s, _ := mobilityWorld(t, 48, func(set func(x, y, g, o int)) {
		for y := 10; y < 30; y++ {
			set(20, y, world.Water, world.None)
			set(30, y, world.Mountain, world.None)
		}
	})
	g := s.terrain.navGraph()
	for i, n := range g.node {
		if (n < 0) != (s.terrain.mobilityAt(i) == mobNever) {
			t.Fatalf("tile %d: patch %d for medium %d", i, n, s.terrain.mobilityAt(i))
		}
		if n >= 0 && g.nodes[n].mob != s.terrain.mobilityAt(i) {
			t.Fatalf("tile %d in a patch of another medium", i)
		}
	}
	for a, n := range g.nodes {
		if s.terrain.mobilityAt(int(n.y)*s.terrain.w+int(n.x)) != n.mob || g.node[int(n.y)*s.terrain.w+int(n.x)] != int32(a) {
			t.Fatalf("patch %d's middle lies outside it", a)
		}
		for _, b := range n.links {
			back := false
			for _, l := range g.nodes[b].links {
				back = back || l == int32(a)
			}
			if !back {
				t.Fatalf("link %d→%d is one-way", a, b)
			}
		}
	}
	// The same ends and body give the same (cached) route.
	from, to := g.node[12*48+12], g.node[40*48+40]
	r1 := g.route(mayWade|maySwim|mayClimb, from, to)
	r2 := g.route(mayWade|maySwim|mayClimb, from, to)
	if len(r1) < 2 || &r1[0] != &r2[0] || r1[0] != from || r1[len(r1)-1] != to {
		t.Fatalf("bad coarse route %v", r1)
	}
}

// legacyLocalPath is the route search as it was before the mobility layer
// (maps instead of reusable arrays, solid tiles only), kept to check that a
// body confined to dry ground is routed exactly as before.
func legacyLocalPath(s *Sim, ax, ay, bx, by float64, radius int) []Waypoint {
	t := s.terrain
	start, ok := t.indexAt(ax, ay)
	if !ok {
		return nil
	}
	goal, ok := t.indexAt(bx, by)
	if !ok || t.blocked[goal] {
		return nil
	}
	cx, cy := start%t.w, start/t.w
	heuristic := func(i int) float64 { return math.Hypot(float64(i%t.w)+0.5-bx, float64(i/t.w)+0.5-by) }
	cost := map[int]float64{start: 0}
	prev := map[int]int{}
	queue := &navQueue{{start, 0, heuristic(start)}}
	best := start
	for count := 0; queue.Len() > 0 && count < 512; count++ {
		n := heap.Pop(queue).(navNode)
		if n.g > cost[n.index] {
			continue
		}
		if heuristic(n.index) < heuristic(best) {
			best = n.index
		}
		if n.index == goal {
			best = goal
			break
		}
		x, y := n.index%t.w, n.index/t.w
		for _, d := range [][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}, {1, 1}, {-1, 1}, {-1, -1}, {1, -1}} {
			nx, ny := x+d[0], y+d[1]
			if absInt(nx-cx) > radius || absInt(ny-cy) > radius {
				continue
			}
			j, in := t.index(nx, ny)
			if !in || t.blocked[j] {
				continue
			}
			if d[0] != 0 && d[1] != 0 && (t.blocked[y*t.w+nx] || t.blocked[ny*t.w+x]) {
				continue
			}
			g := n.g + math.Hypot(float64(d[0]), float64(d[1]))
			if old, seen := cost[j]; seen && old <= g {
				continue
			}
			cost[j], prev[j] = g, n.index
			heap.Push(queue, navNode{j, g, g + heuristic(j)})
		}
	}
	if best != goal && absInt(goal%t.w-cx) <= radius && absInt(goal/t.w-cy) <= radius {
		return nil
	}
	var reversed []Waypoint
	for i := best; i != start; {
		reversed = append(reversed, Waypoint{float64(i%t.w) + 0.5, float64(i/t.w) + 0.5})
		p, ok := prev[i]
		if !ok {
			return nil
		}
		i = p
	}
	path := make([]Waypoint, len(reversed))
	for i := range reversed {
		path[i] = reversed[len(reversed)-1-i]
	}
	return path
}

// On dry ground (and in a world without mobility) routes are exactly what
// they were before water and slopes became passable.
func TestDryRoutesMatchTheOldSearch(t *testing.T) {
	s := newSimWith(testMap(t, 64), 1, fakeCatalog())
	rng := rand.New(rand.NewPCG(7, 7))
	for range 3000 {
		ax, ay := rng.Float64()*64, rng.Float64()*64
		bx, by := ax+rng.Float64()*40-20, ay+rng.Float64()*40-20
		radius := 4 + rng.IntN(11)
		want := legacyLocalPath(s, ax, ay, bx, by, radius)
		if got := s.localPath(ax, ay, bx, by, radius); !slices.Equal(got, want) {
			t.Fatalf("route (%.1f,%.1f)→(%.1f,%.1f) r%d: %v, was %v", ax, ay, bx, by, radius, got, want)
		}
	}
}
