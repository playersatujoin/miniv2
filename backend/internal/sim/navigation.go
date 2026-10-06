package sim

import (
	"container/heap"
	"math"
	"slices"
)

type Waypoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Travel is an executor for an action the neural network still requests.
// It never chooses needs or starts a job. Paths are short, local and persisted.
type Travel struct {
	Action  Action     `json:"action"`
	Phase   string     `json:"phase"` // approach, blocked, reached
	Target  Waypoint   `json:"target"`
	Path    []Waypoint `json:"path,omitempty"`
	Step    int        `json:"step"`
	Planned int64      `json:"planned"`
}

type ExecutionView struct {
	Action    Action    `json:"action"`
	Phase     string    `json:"phase"`
	Target    *Waypoint `json:"target,omitempty"`
	Remaining float64   `json:"remaining,omitempty"`
	Waypoints int       `json:"waypoints"`
}

func (s *Sim) executionView(c *Creature) *ExecutionView {
	if c.Job != nil {
		return &ExecutionView{Action: c.Job.action(), Phase: "perform", Remaining: r3(c.Job.Remaining)}
	}
	if n := c.Travel; n != nil {
		p := n.Target
		return &ExecutionView{Action: n.Action, Phase: n.Phase, Target: &p, Waypoints: max(0, len(n.Path)-n.Step)}
	}
	return nil
}

func (s *Sim) travelIntent(c *Creature) Action {
	if s.opts.NoNavigation || c.resting || c.Job != nil || c.output[outMove] <= 0.1 {
		return ""
	}
	best, action := 0.5, Action("")
	for _, choice := range []struct {
		a    Action
		out  int
		want bool
	}{
		{ActDrink, outDrink, c.Hydration < 0.85 && !s.canDrink(c)},
		{ActEat, outEat, c.Energy < 0.85 && !c.ate},
		{ActGather, outGather, s.adult(c) && c.Inventory.count() < invCapacity && c.Gathering == 0 && c.action == ActExplore},
	} {
		if choice.want && c.output[choice.out] > best {
			action, best = choice.a, c.output[choice.out]
		}
	}
	return action
}

// travelTarget only considers currently visible terrain, plus the existing
// memory of drinking water. It cannot reveal food behind a wall or across a map.
func (s *Sim) travelTarget(c *Creature, action Action) (Waypoint, bool) {
	r := int(math.Ceil(c.Genome.Traits.Vision))
	cx, cy := int(c.X), int(c.Y)
	best := math.Inf(1)
	var goal Waypoint
	for y := max(0, cy-r); y <= min(s.terrain.h-1, cy+r); y++ {
		for x := max(0, cx-r); x <= min(s.terrain.w-1, cx+r); x++ {
			i := y*s.terrain.w + x
			px, py := float64(x)+0.5, float64(y)+0.5
			d := math.Hypot(px-c.X, py-c.Y)
			if d < 0.8 || d >= best || s.terrain.blocked[i] || !s.sees(c, px, py) {
				continue
			}
			good := false
			switch action {
			case ActDrink:
				good = s.waterAt(px, py)
			case ActEat:
				good = s.eco.Forage(i) > 0.15
			case ActGather:
				good = s.resourceAt(c, i, s.toolBonus(c) > 0) > 0.15
			}
			if good {
				best, goal = d, Waypoint{px, py}
			}
		}
	}
	if math.IsInf(best, 1) && action == ActDrink && (c.WaterX != 0 || c.WaterY != 0) && math.Hypot(c.WaterX-c.X, c.WaterY-c.Y) > 1 {
		return Waypoint{c.WaterX, c.WaterY}, true
	}
	// No food in sight: the remembered place most worth the walk (forage.go).
	if math.IsInf(best, 1) && action == ActEat {
		if p, v := s.bestFoodPlace(c); v > foodPlaceMin {
			return Waypoint{p.X, p.Y}, true
		}
	}
	return goal, !math.IsInf(best, 1)
}

// navigate returns the fraction of normal speed while turning onto the route.
func (s *Sim) navigate(c *Creature) float64 {
	action := s.travelIntent(c)
	if action == "" {
		c.Travel = nil
		return 1
	}
	if c.Travel != nil && c.Travel.Action != action {
		c.Travel = nil
	}
	// Limit path work even if the brain changes its mind repeatedly.
	if s.tick >= c.NextRoute && (c.Travel == nil || c.Bumped || s.tick-c.Travel.Planned >= 2*TicksPerSecond) {
		c.NextRoute = s.tick + TicksPerSecond
		if goal, ok := s.travelTarget(c, action); ok {
			path := s.routePath(c, goal.X, goal.Y, min(14, int(math.Ceil(c.Genome.Traits.Vision))+3))
			phase := "approach"
			if len(path) == 0 {
				phase = "blocked"
			}
			c.Travel = &Travel{Action: action, Phase: phase, Target: goal, Path: path, Planned: s.tick}
		} else {
			c.Travel = nil
		}
	}
	n := c.Travel
	if n == nil || n.Phase != "approach" {
		return 1
	}
	for n.Step < len(n.Path) && math.Hypot(n.Path[n.Step].X-c.X, n.Path[n.Step].Y-c.Y) < 0.2 {
		n.Step++
	}
	if n.Step >= len(n.Path) {
		n.Phase = "reached"
		return 1
	}
	p := n.Path[n.Step]
	if !s.standable(c, p.X, p.Y) {
		n.Phase = "blocked"
		return 1
	}
	delta := normAngle(math.Atan2(p.Y-c.Y, p.X-c.X) - c.Heading)
	c.Heading = normAngle(c.Heading + clamp(delta, -maxTurnRate*dt, maxTurnRate*dt))
	return math.Max(0.1, math.Cos(delta))
}

// Route costs per tile relative to walking, like the climb cost modifier of
// a ped's navigation capabilities: the time a medium takes (1/speed factor)
// and, where a body can come to harm, a premium for the risk. Routes so keep
// to dry land, then the shallows, then slopes, and swim only short gaps.
const (
	wadeCost  = 1 / wadeFactor
	raftCost  = 1 / raftFactor
	climbCost = 2 / climbFactor
	swimCost  = 2.5 / swimFactor
	// A swimmer plans across deep water no farther than this from a place to
	// stand (a gap of about twice as many tiles), a weaker one half as far.
	swimReach = 2
)

// fitSwimmer is the stamina (see stamina) a swimmer needs for the longer gaps.
const fitSwimmer = 0.7

// routeCapsFor is routeCaps plus whether c is fit for the longer swims now.
func (s *Sim) routeCapsFor(c *Creature) navCaps {
	k := s.routeCaps(c)
	if k&maySwim != 0 && s.stamina(c)*(1-c.Loco.Fatigue) >= fitSwimmer {
		k |= maySwimFar
	}
	return k
}

// costOf is what crossing ground of medium m (deep water reach tiles from a
// place to stand) costs a body with caps k; 0 where it may not go.
func costOf(m mobility, reach uint8, k navCaps) float64 {
	switch m {
	case mobWalk:
		return 1
	case mobWade:
		if k&mayWade != 0 {
			return wadeCost
		}
	case mobClimb:
		if k&mayClimb != 0 {
			return climbCost
		}
	case mobDeep:
		if k&mayRaft != 0 {
			return raftCost
		}
		limit := uint8(swimReach / 2)
		if k&maySwimFar != 0 {
			limit = swimReach
		}
		if k&maySwim != 0 && reach <= limit {
			return swimCost
		}
	}
	return 0
}

// tileCost is costOf for tile i.
func (t *terrain) tileCost(i int, k navCaps) float64 {
	if !t.blocked[i] {
		return 1
	}
	if k == 0 || t.rough == nil {
		return 0
	}
	var reach uint8
	if t.deepReach != nil {
		reach = t.deepReach[i]
	}
	return costOf(t.rough[i], reach, k)
}

// routePath plans c's way towards (bx, by) within the local window: on foot,
// through water or over slopes as its body allows. A destination beyond the
// window is approached along the coarse route (see navGraph) rather than in
// a straight line.
func (s *Sim) routePath(c *Creature, bx, by float64, radius int) []Waypoint {
	if s.opts.NoMobility {
		return s.localPath(c.X, c.Y, bx, by, radius)
	}
	k := s.routeCapsFor(c)
	cx, cy := int(math.Floor(c.X)), int(math.Floor(c.Y))
	if absInt(int(math.Floor(bx))-cx) > radius || absInt(int(math.Floor(by))-cy) > radius {
		if wx, wy, ok := s.corridorGoal(k, c.X, c.Y, bx, by, radius); ok {
			if p := s.searchPath(k, c.X, c.Y, wx, wy, radius); len(p) > 0 {
				return p
			}
		}
	}
	return s.searchPath(k, c.X, c.Y, bx, by, radius)
}

type navNode struct {
	index int
	g, f  float64
}
type navQueue []navNode

func (q navQueue) Len() int { return len(q) }
func (q navQueue) Less(i, j int) bool {
	if q[i].f == q[j].f {
		return q[i].index < q[j].index
	}
	return q[i].f < q[j].f
}
func (q navQueue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }
func (q *navQueue) Push(v any)   { *q = append(*q, v.(navNode)) }
func (q *navQueue) Pop() any     { a := *q; n := a[len(a)-1]; *q = a[:len(a)-1]; return n }

// navScratch is reusable working memory for searches on one terrain: costs
// and predecessors per tile, valid where stamp equals the current search.
type navScratch struct {
	cost  []float64
	prev  []int32
	stamp []uint32
	gen   uint32
	queue navQueue
}

func (t *terrain) scratch() *navScratch {
	if t.search == nil {
		n := len(t.blocked)
		t.search = &navScratch{cost: make([]float64, n), prev: make([]int32, n), stamp: make([]uint32, n)}
	}
	sc := t.search
	sc.gen++
	if sc.gen == 0 { // wrapped: forget every stamp
		clear(sc.stamp)
		sc.gen = 1
	}
	sc.queue = sc.queue[:0]
	return sc
}

// Bounded A*: the local terrain is motor knowledge. A remembered distant
// destination gets a frontier waypoint; inaccessible destinations never
// teleport. localPath plans for dry ground only.
func (s *Sim) localPath(ax, ay, bx, by float64, radius int) []Waypoint {
	return s.searchPath(0, ax, ay, bx, by, radius)
}

// searchPath is the bounded A* for a body with capabilities k: each tile
// costs its medium (tileCost), and diagonals never cut a corner between two
// tiles the body can't enter.
func (s *Sim) searchPath(k navCaps, ax, ay, bx, by float64, radius int) []Waypoint {
	t := s.terrain
	start, ok := t.indexAt(ax, ay)
	if !ok {
		return nil
	}
	goal, ok := t.indexAt(bx, by)
	if !ok || t.tileCost(goal, k) == 0 {
		return nil
	}
	cx, cy := start%t.w, start/t.w
	heuristic := func(i int) float64 { return math.Hypot(float64(i%t.w)+0.5-bx, float64(i/t.w)+0.5-by) }
	sc := t.scratch()
	seen := func(i int) bool { return sc.stamp[i] == sc.gen }
	sc.stamp[start], sc.cost[start], sc.prev[start] = sc.gen, 0, -1
	sc.queue = append(sc.queue, navNode{start, 0, heuristic(start)})
	queue := &sc.queue
	best := start
	for count := 0; queue.Len() > 0 && count < 512; count++ {
		n := heap.Pop(queue).(navNode)
		if n.g > sc.cost[n.index] {
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
			if !in {
				continue
			}
			step := t.tileCost(j, k)
			if step == 0 {
				continue
			}
			// Diagonals cannot cut a corner between two solid tiles.
			if d[0] != 0 && d[1] != 0 && (t.tileCost(y*t.w+nx, k) == 0 || t.tileCost(ny*t.w+x, k) == 0) {
				continue
			}
			g := n.g + math.Hypot(float64(d[0]), float64(d[1]))*step
			if seen(j) && sc.cost[j] <= g {
				continue
			}
			sc.stamp[j], sc.cost[j], sc.prev[j] = sc.gen, g, int32(n.index)
			heap.Push(queue, navNode{j, g, g + heuristic(j)})
		}
	}
	// Only use a frontier if the destination lies outside this local search.
	if best != goal && absInt(goal%t.w-cx) <= radius && absInt(goal/t.w-cy) <= radius {
		return nil
	}
	var reversed []Waypoint
	for i := best; i != start; {
		reversed = append(reversed, Waypoint{float64(i%t.w) + 0.5, float64(i/t.w) + 0.5})
		if !seen(i) || sc.prev[i] < 0 {
			return nil
		}
		i = int(sc.prev[i])
	}
	path := make([]Waypoint, len(reversed))
	for i := range reversed {
		path[i] = reversed[len(reversed)-1-i]
	}
	return path
}

// --- the coarse route layer ------------------------------------------------------

// The coarse layer, after RAGE's hierarchical path server
// (pathserver/PathServer_Hierarchical.cpp): the map is cut into regions of
// regionSize² tiles, each split into patches of one medium (dry ground,
// shallows, deep water, slope) that hang together, and patches that touch
// are linked. A route to a place beyond the local window is first found over
// these patches; the local A* then heads for the farthest patch on it that
// lies within reach, so travel follows the corridor around a lake or a
// mountain instead of running into it. The graph is derived from the map
// alone (rebuilt when the map changes) and routes are cached as a pure
// function of their ends and the body's capabilities, so nothing about it
// needs saving.
const (
	regionSize    = 8
	routeCacheMax = 4096 // cached coarse routes before the cache is emptied
)

type coarseNode struct {
	mob   mobility
	x, y  int32   // the patch's tile nearest its middle
	reach uint8   // deep water: farthest any of its tiles lies from a place to stand
	links []int32 // patches it touches
}

type navGraph struct {
	node   []int32 // per tile: its patch, -1 where nobody can go
	nodes  []coarseNode
	routes map[uint64][]int32
	// Working memory for the coarse search.
	cost  []float64
	prev  []int32
	stamp []uint32
	gen   uint32
	queue navQueue
}

// navGraph returns the terrain's coarse route layer, building it on first use.
func (t *terrain) navGraph() *navGraph {
	if t.nav == nil {
		t.nav = buildNavGraph(t)
	}
	return t.nav
}

func buildNavGraph(t *terrain) *navGraph {
	g := &navGraph{node: make([]int32, len(t.blocked)), routes: map[uint64][]int32{}}
	for i := range g.node {
		g.node[i] = -1
	}
	var stack, members []int
	for ry := 0; ry < t.h; ry += regionSize {
		for rx := 0; rx < t.w; rx += regionSize {
			x1, y1 := min(t.w, rx+regionSize), min(t.h, ry+regionSize)
			for y := ry; y < y1; y++ {
				for x := rx; x < x1; x++ {
					i := y*t.w + x
					m := t.mobilityAt(i)
					if m == mobNever || g.node[i] >= 0 {
						continue
					}
					// Flood the patch: tiles of this medium joined side by side
					// within the region.
					id := int32(len(g.nodes))
					g.node[i] = id
					stack = append(stack[:0], i)
					members = members[:0]
					sx, sy := 0, 0
					var reach uint8
					for len(stack) > 0 {
						j := stack[len(stack)-1]
						stack = stack[:len(stack)-1]
						members = append(members, j)
						jx, jy := j%t.w, j/t.w
						sx, sy = sx+jx, sy+jy
						if m == mobDeep && t.deepReach != nil {
							reach = max(reach, t.deepReach[j])
						}
						for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
							kx, ky := jx+d[0], jy+d[1]
							if kx < rx || ky < ry || kx >= x1 || ky >= y1 {
								continue
							}
							if k := ky*t.w + kx; g.node[k] < 0 && t.mobilityAt(k) == m {
								g.node[k] = id
								stack = append(stack, k)
							}
						}
					}
					mx, my := float64(sx)/float64(len(members)), float64(sy)/float64(len(members))
					rep, repD := members[0], math.Inf(1)
					for _, j := range members {
						if d := math.Hypot(float64(j%t.w)-mx, float64(j/t.w)-my); d < repD || d == repD && j < rep {
							rep, repD = j, d
						}
					}
					g.nodes = append(g.nodes, coarseNode{mob: m, x: int32(rep % t.w), y: int32(rep / t.w), reach: reach})
				}
			}
		}
	}
	link := func(a, b int32) {
		for _, l := range g.nodes[a].links {
			if l == b {
				return
			}
		}
		g.nodes[a].links = append(g.nodes[a].links, b)
		g.nodes[b].links = append(g.nodes[b].links, a)
	}
	for i, a := range g.node {
		if a < 0 {
			continue
		}
		x, y := i%t.w, i/t.w
		if x+1 < t.w {
			if b := g.node[i+1]; b >= 0 && b != a {
				link(a, b)
			}
		}
		if y+1 < t.h {
			if b := g.node[i+t.w]; b >= 0 && b != a {
				link(a, b)
			}
		}
	}
	n := len(g.nodes)
	g.cost, g.prev, g.stamp = make([]float64, n), make([]int32, n), make([]uint32, n)
	return g
}

// route finds the cheapest chain of patches from one to another for a body
// with capabilities k, or nil if there is none.
func (g *navGraph) route(k navCaps, from, to int32) []int32 {
	key := uint64(k)<<48 | uint64(from)<<24 | uint64(to)
	if r, ok := g.routes[key]; ok {
		return r
	}
	r := g.search(k, from, to)
	if len(g.routes) >= routeCacheMax {
		clear(g.routes)
	}
	g.routes[key] = r
	return r
}

func (g *navGraph) nodeCost(n int32, k navCaps) float64 {
	c := &g.nodes[n]
	return costOf(c.mob, c.reach, k)
}

func (g *navGraph) search(k navCaps, from, to int32) []int32 {
	if g.nodeCost(from, k) == 0 || g.nodeCost(to, k) == 0 {
		return nil
	}
	g.gen++
	if g.gen == 0 {
		clear(g.stamp)
		g.gen = 1
	}
	goal := &g.nodes[to]
	dist := func(a, b *coarseNode) float64 { return math.Hypot(float64(a.x-b.x), float64(a.y-b.y)) }
	g.stamp[from], g.cost[from], g.prev[from] = g.gen, 0, -1
	g.queue = append(g.queue[:0], navNode{int(from), 0, dist(&g.nodes[from], goal)})
	found := false
	for count := 0; g.queue.Len() > 0 && count < 4*len(g.nodes); count++ {
		n := heap.Pop(&g.queue).(navNode)
		if n.g > g.cost[n.index] {
			continue
		}
		if int32(n.index) == to {
			found = true
			break
		}
		a := &g.nodes[n.index]
		ca := g.nodeCost(int32(n.index), k)
		for _, b := range a.links {
			cb := g.nodeCost(b, k)
			if cb == 0 {
				continue
			}
			nb := &g.nodes[b]
			cost := n.g + math.Max(1, dist(a, nb))*(ca+cb)/2
			if g.stamp[b] == g.gen && g.cost[b] <= cost {
				continue
			}
			g.stamp[b], g.cost[b], g.prev[b] = g.gen, cost, int32(n.index)
			heap.Push(&g.queue, navNode{int(b), cost, cost + dist(nb, goal)})
		}
	}
	if !found {
		return nil
	}
	var path []int32
	for n := to; n >= 0; n = g.prev[n] {
		path = append(path, n)
	}
	slices.Reverse(path)
	return path
}

// corridorGoal is where to head within the local window (radius tiles) on
// the coarse route from (ax, ay) to (bx, by): the farthest patch on it whose
// middle lies inside the window, else the next patch on it (outside, so the
// local search heads for its frontier).
func (s *Sim) corridorGoal(k navCaps, ax, ay, bx, by float64, radius int) (float64, float64, bool) {
	t := s.terrain
	a, okA := t.indexAt(ax, ay)
	b, okB := t.indexAt(bx, by)
	if !okA || !okB {
		return 0, 0, false
	}
	g := t.navGraph()
	from, to := g.node[a], g.node[b]
	if from < 0 || to < 0 || from == to {
		return 0, 0, false
	}
	path := g.route(k, from, to)
	if len(path) < 2 {
		return 0, 0, false
	}
	cx, cy := a%t.w, a/t.w
	for i := len(path) - 1; i > 0; i-- {
		n := &g.nodes[path[i]]
		if absInt(int(n.x)-cx) < radius && absInt(int(n.y)-cy) < radius {
			return float64(n.x) + 0.5, float64(n.y) + 0.5, true
		}
	}
	n := &g.nodes[path[1]]
	return float64(n.x) + 0.5, float64(n.y) + 0.5, true
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
