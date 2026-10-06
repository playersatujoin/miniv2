package ecology

import (
	"fmt"
	"math"
	"slices"
)

// The water cycle. Rain falls on the land; the soil hands much of it back to
// the air, and what is left either runs off at once (quickflow) or soaks down
// to the groundwater, which leaks back out into the streams for months after
// the rain has stopped (baseflow). This is the textbook "linear reservoir"
// picture of a catchment (as in the HBV model). Each river tile and each lake
// is a store of water fed by the land that drains to it and by the river
// above it, and losing water to the air (open water evaporates about 5 mm a
// day), into its bed, to whoever drinks from it and to the paddies it
// waters; what is left flows on downstream.
//
// So rivers swell in the wet season and shrink in the dry. Small streams,
// with little groundwater behind them, stop flowing first; the flow breaks
// up into pools, which the sun, the sand and the thirsty use up in weeks;
// then the bed is dry and stays dry until the rains return. Big rivers, fed
// by the slow drain of a large aquifer, keep running except in the worst
// droughts. Lakes shrink from their shallow edges inwards. Under a dry bed
// there is often still water in the sand, which people dig for (the belik of
// Java), and wells reach the groundwater until it too sinks out of reach, as
// in an El Niño year.
//
// One unit of water is 16 m³: about a metre over a tile of some 4 × 4 m.

const (
	hydroEvery = 2 // ticks between steps of the water cycle

	rainDepth = 2.5   // units per year falling on a tile at average rain (≈ 2,500 mm)
	etLand    = 1.6   // units per year a well-watered tile gives back to the air
	evapOpen  = 1.8   // units per year leaving a tile of open water (≈ 5 mm a day)
	seepBed   = 1.5   // units per year a wet river tile loses into its bed (≈ 4 mm a day)
	seepLake  = 0.3   // … and a lake into its clay bottom
	kChannel  = 0.004 // years a river tile holds its flowing water (a day or two)
	bankfull  = 3.0   // units per year per upstream tile a channel carries when full
	poolMin   = 0.15  // units a river tile keeps in pools when the flow stops: shallow riffles …
	poolMax   = 0.7   // … to deep holes
	lakeDeep  = 2.5   // units the deepest tile of a lake holds
	lakeSill  = 0.8   // share of a lake's volume below its outlet: above it, it drains out …
	kLake     = 0.06  // … over about three weeks

	// Groundwater: how many years the aquifer behind a stretch of river takes
	// to drain, longer behind big rivers.
	gwBase  = 0.08
	gwScale = 0.06
	gwArea  = 20.0
	// gwRef is the recharge (units per tile per year) a full aquifer is measured against.
	gwRef = 1.0

	// WellMin and SoakMin: groundwater (a share of its usual wet-season level)
	// below which wells and holes dug in a dry riverbed find no more water.
	WellMin = 0.12
	SoakMin = 0.3

	// irrigNeed is what a paddy's crop drinks in a year (units per tile) when
	// no rain falls on it.
	irrigNeed = 1.6
)

// Water states of a fresh water tile.
const (
	WaterDry     uint8 = iota // nothing to drink
	WaterUnder                // a dry bed with water in the sand below
	WaterPools                // the flow has stopped; pools remain
	WaterFlowing              // running (low or full)
)

type hnode struct {
	tiles []int32
	lake  bool
	down  int32   // the node this one flows into, -1 for the sea (or nowhere)
	area  float64 // land tiles draining straight here
	up    float64 // land and water tiles draining here in all, upstream included
	dead  float64 // water a river tile keeps in its pools; for a lake, its whole volume
	cap   float64 // water when full to the banks
	kb    float64 // years the aquifer takes to drain
	// Lakes: how shallow each tile is (0 the deepest … 1 the shore), in tiles order.
	shallow []float32
}

type hydrology struct {
	nodes []hnode
	order []int32 // upstream before downstream
	// Per tile: its node for fresh water tiles, and for land the node it
	// drains to (its groundwater) and the nearest node (what a channel taps).
	nodeOf  []int32
	drainTo []int32
	nearest []int32

	// State.
	S []float64 // water in each node
	G []float64 // groundwater behind each node
	W []float64 // drunk or drawn from each node since the last step
	U []float64 // … and from the groundwater (wells, dug holes)
	// Germs in each node's water, and those washed in since the last step
	// (see pathogens.go).
	P   []float64
	Pin []float64
	// reach is how many tiles each land tile lies from its nearest node.
	reach []uint8

	// Per tile, refreshed every step: state and level (0–255) of fresh water,
	// and how well irrigated land is supplied (0–1).
	state  []uint8
	level  []uint8
	foul   []uint8 // germs in the water, 0–255 (see pathogens.go)
	supply []float32
	// Island totals, for the observer and for events.
	flowing, pools, under, dry int
	version                    int64
	worst                      float64 // lowest share of running water this dry season
	warned                     int     // what the last event said: 0 nothing, 1 drying, 2 dry, 3 running again
}

// buildHydrology works out the river network: which tiles feed which river
// stretch or lake, and how they connect on the way to the sea.
func (e *Ecology) buildHydrology() {
	l := &e.land
	n := l.W * l.H
	h := &e.hydro
	h.nodes = h.nodes[:0]
	h.nodeOf = make([]int32, n)
	h.drainTo = make([]int32, n)
	h.nearest = make([]int32, n)
	h.state = make([]uint8, n)
	h.level = make([]uint8, n)
	h.foul = make([]uint8, n)
	h.supply = make([]float32, n)
	h.reach = make([]uint8, n)
	for i := range n {
		h.nodeOf[i], h.drainTo[i], h.nearest[i] = -1, -1, -1
	}
	river := func(i int) bool { return l.River != nil && l.River[i] }
	for i := range n {
		if e.fresh[i] && river(i) {
			h.nodeOf[i] = int32(len(h.nodes))
			h.nodes = append(h.nodes, hnode{tiles: []int32{int32(i)}})
		}
	}
	// Lakes: what is left of the fresh water, in connected pieces.
	for i := range n {
		if !e.fresh[i] || h.nodeOf[i] >= 0 {
			continue
		}
		id := int32(len(h.nodes))
		nd := hnode{lake: true}
		h.nodeOf[i] = id
		queue := []int32{int32(i)}
		for len(queue) > 0 {
			j := int(queue[0])
			queue = queue[1:]
			nd.tiles = append(nd.tiles, int32(j))
			x, y := j%l.W, j/l.W
			for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				if k, ok := l.index(x+d[0], y+d[1]); ok && e.fresh[k] && h.nodeOf[k] < 0 && !river(k) {
					h.nodeOf[k] = id
					queue = append(queue, int32(k))
				}
			}
		}
		slices.Sort(nd.tiles)
		h.nodes = append(h.nodes, nd)
	}

	// Where each land tile's water goes: down the slope to the first river or
	// lake it meets (or the sea). Without a slope map, to the nearest one.
	if l.Down != nil {
		for i := range n {
			if e.fresh[i] || l.Water[i] || h.drainTo[i] >= 0 {
				continue
			}
			var path []int32
			j, to := i, int32(-1)
			for steps := 0; steps < n; steps++ {
				if h.nodeOf[j] >= 0 {
					to = h.nodeOf[j]
					break
				}
				if l.Water[j] || (h.drainTo[j] >= 0 && j != i) {
					to = h.drainTo[j]
					break
				}
				path = append(path, int32(j))
				next := l.Down[j]
				if next < 0 || int(next) == j {
					break
				}
				j = int(next)
			}
			for _, p := range path {
				h.drainTo[p] = to
			}
		}
	}
	// The nearest node within reach of every tile (BFS from the water), also
	// the fallback for where land drains.
	dist := make([]int16, n)
	for i := range dist {
		dist[i] = -1
	}
	var queue []int32
	for i := range n {
		if h.nodeOf[i] >= 0 {
			h.nearest[i], dist[i] = h.nodeOf[i], 0
			queue = append(queue, int32(i))
		}
	}
	const reach = 12
	for len(queue) > 0 {
		i := int(queue[0])
		queue = queue[1:]
		if dist[i] >= reach {
			continue
		}
		x, y := i%l.W, i/l.W
		for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			if k, ok := l.index(x+d[0], y+d[1]); ok && dist[k] < 0 && !l.Water[k] {
				dist[k] = dist[i] + 1
				h.nearest[k] = h.nearest[i]
				h.reach[k] = uint8(dist[k])
				queue = append(queue, int32(k))
			}
		}
	}
	if l.Down == nil {
		for i := range n {
			if !l.Water[i] {
				h.drainTo[i] = h.nearest[i]
			}
		}
	}
	for i := range n {
		if to := h.drainTo[i]; to >= 0 && !l.Water[i] {
			h.nodes[to].area++
		}
	}

	// Downstream: from a river tile, or from a lake's outlet, follow the slope
	// to the next node.
	for id := range h.nodes {
		nd := &h.nodes[id]
		nd.down = -1
		if l.Down == nil {
			continue
		}
		for _, t := range nd.tiles {
			j := int(l.Down[t])
			for steps := 0; steps < n && j >= 0; steps++ {
				if o := h.nodeOf[j]; o >= 0 && o != int32(id) {
					nd.down = o
					break
				}
				if o := h.nodeOf[j]; o == int32(id) {
					break // still inside this lake
				}
				if l.Water[j] {
					break // the sea
				}
				if to := h.drainTo[j]; to >= 0 && to != int32(id) {
					nd.down = to
					break
				}
				next := int(l.Down[j])
				if next == j {
					break
				}
				j = next
			}
			if nd.down >= 0 || !nd.lake {
				break
			}
		}
	}
	// Upstream before downstream (Kahn); a loop, which a hand-edited map might
	// make, is cut.
	indeg := make([]int, len(h.nodes))
	for _, nd := range h.nodes {
		if nd.down >= 0 {
			indeg[nd.down]++
		}
	}
	h.order = h.order[:0]
	for id := range h.nodes {
		if indeg[id] == 0 {
			h.order = append(h.order, int32(id))
		}
	}
	for k := 0; k < len(h.order); k++ {
		if d := h.nodes[h.order[k]].down; d >= 0 {
			if indeg[d]--; indeg[d] == 0 {
				h.order = append(h.order, d)
			}
		}
	}
	if len(h.order) < len(h.nodes) {
		seen := make([]bool, len(h.nodes))
		for _, id := range h.order {
			seen[id] = true
		}
		for id := range h.nodes {
			if !seen[id] {
				h.nodes[id].down = -1
				h.order = append(h.order, int32(id))
			}
		}
	}
	for _, id := range h.order {
		nd := &h.nodes[id]
		nd.up += nd.area + float64(len(nd.tiles))
		if nd.down >= 0 {
			h.nodes[nd.down].up += nd.up
		}
	}
	for id := range h.nodes {
		nd := &h.nodes[id]
		nd.kb = gwBase + gwScale*math.Log(1+nd.up/gwArea)
		if nd.lake {
			e.shapeLake(nd)
			continue
		}
		t := int(nd.tiles[0])
		nd.dead = poolMin + (poolMax-poolMin)*hash01(t%l.W, t/l.W, 71)
		nd.cap = nd.dead + kChannel*bankfull*nd.up
	}
	h.S = make([]float64, len(h.nodes))
	h.G = make([]float64, len(h.nodes))
	h.W = make([]float64, len(h.nodes))
	h.U = make([]float64, len(h.nodes))
	h.P = make([]float64, len(h.nodes))
	h.Pin = make([]float64, len(h.nodes))
}

// shapeLake gives a lake its depths: deepest far from the shore.
func (e *Ecology) shapeLake(nd *hnode) {
	l := &e.land
	d := make([]int, len(nd.tiles))
	far := 1
	for k, t := range nd.tiles {
		x, y := int(t)%l.W, int(t)/l.W
		d[k] = 99
		for r := 1; r < 99 && d[k] == 99; r++ {
			for dy := -r; dy <= r && d[k] == 99; dy++ {
				for dx := -r; dx <= r; dx++ {
					if max(abs(dx), abs(dy)) != r {
						continue
					}
					if j, ok := l.index(x+dx, y+dy); !ok || !e.fresh[j] {
						d[k] = r
						break
					}
				}
			}
		}
		far = max(far, d[k])
	}
	nd.shallow = make([]float32, len(nd.tiles))
	nd.dead = 0
	for k, dk := range d {
		depth := float64(dk) / float64(far)
		nd.shallow[k] = float32(1 - depth)
		nd.dead += lakeDeep * (0.3 + 0.7*depth)
	}
	nd.cap = nd.dead
}

func smoothstep(lo, hi, v float64) float64 {
	t := clamp((v-lo)/(hi-lo), 0, 1)
	return t * t * (3 - 2*t)
}

// hash01 is a fixed number in [0, 1) for a tile and a salt: what the land is
// like there (how deep the pools in a riverbed are), the same every run.
func hash01(x, y int, salt uint64) float64 {
	h := uint64(x)*0x9E3779B97F4A7C15 ^ uint64(y)*0xC2B2AE3D27D4EB4F ^ salt*0x165667B19E3779F9
	h ^= h >> 31
	h *= 0xD6E8FEB86659FD93
	h ^= h >> 32
	return float64(h>>11) / (1 << 53)
}

// fillHydrology sets the stores as they would be after an ordinary year of
// the current weather (a new world, or one saved before the water cycle).
func (e *Ecology) fillHydrology() {
	h := &e.hydro
	wet := clamp(e.clim.Slow*moistFull, 0.2, 1.5)
	for id := range h.nodes {
		nd := &h.nodes[id]
		h.G[id] = (nd.area + 1) * gwRef * nd.kb * wet
		if nd.lake {
			h.S[id] = nd.cap * clamp(0.6*wet, 0.1, 1)
		} else {
			h.S[id] = nd.dead + (nd.cap-nd.dead)*clamp(0.4*wet, 0, 1)
		}
	}
	e.refreshWater()
}

// stepHydrology moves the water on by dt seconds.
func (e *Ecology) stepHydrology(dt float64) {
	h := &e.hydro
	if len(h.nodes) == 0 {
		return
	}
	if e.opts.NoWaterCycle {
		for id := range h.nodes {
			nd := &h.nodes[id]
			h.S[id], h.G[id] = nd.cap, (nd.area+1)*gwRef*nd.kb
			h.W[id], h.U[id] = 0, 0
		}
		for i := range h.supply {
			h.supply[i] = 1
		}
		e.carryGerms(dt/SecondsPerYear, nil)
		e.refreshWater()
		return
	}
	c := &e.clim
	yr := dt / SecondsPerYear
	// Hot, dry months draw more water into the air; an El Niño more still.
	pet := clamp(1.3-0.35*c.Rain, 0.7, 1.3)
	if c.ENSO == ElNino {
		pet *= 1.1
	}
	rain := rainDepth * c.Rain
	m := c.Moisture
	et := etLand * pet * math.Min(1, 0.2+0.9*m)
	// The first rains after a dry season soak into the parched soil; only a
	// wet soil lets water run off or sink through to the groundwater.
	excess := math.Max(0, rain-et) * smoothstep(0.35, 0.8, m)
	quick := excess * (0.25 + 0.45*m)
	recharge := excess - quick

	// Paddies draw what the rain doesn't give them from their channel's river.
	e.irrigate(rain, pet, yr)

	inflow := make([]float64, len(h.nodes))
	outflow := make([]float64, len(h.nodes))
	for _, id := range h.order {
		nd := &h.nodes[id]
		// Groundwater fills with the soaked-in rain and drains out as baseflow.
		g := h.G[id] + (nd.area+1)*recharge*yr - h.U[id]
		base := math.Max(0, g) * (1 - math.Exp(-yr/nd.kb))
		h.G[id] = math.Max(0, g-base)
		h.U[id] = 0
		s := h.S[id] + inflow[id] + nd.area*quick*yr + base + float64(len(nd.tiles))*rain*yr
		// Open water evaporates and seeps into the bed; people and fields take their share.
		wetTiles := e.wetTiles(nd, s)
		loss := wetTiles * evapOpen * pet * yr
		seep := 0.0
		if nd.lake {
			seep = wetTiles * seepLake * yr
		} else if s > 0 {
			seep = seepBed * yr
		}
		seep = math.Min(seep, math.Max(0, s-loss))
		h.G[id] += seep
		s = math.Max(0, s-loss-seep-h.W[id])
		h.W[id] = 0
		// What stands above the pools (or above a full lake) flows on.
		var out float64
		if nd.lake {
			if sill := lakeSill * nd.cap; s > sill {
				out = (s - sill) * (1 - math.Exp(-yr/kLake))
			}
			out += math.Max(0, s-out-nd.cap) // and a flood spills over at once
		} else if s > nd.dead {
			out = (s - nd.dead) * (1 - math.Exp(-yr/kChannel))
		}
		s -= out
		if nd.down >= 0 {
			inflow[nd.down] += out
		}
		h.S[id] = s
		// The share of the water that left, for the germs it carries.
		outflow[id] = out / math.Max(s+out, 1e-9)
	}
	e.carryGerms(yr, outflow)
	e.refreshWater()
}

// wetTiles is how many of a node's tiles hold water with s in it.
func (e *Ecology) wetTiles(nd *hnode, s float64) float64 {
	if s <= 1e-9 {
		return 0
	}
	if !nd.lake {
		return 1
	}
	return math.Max(1, float64(len(nd.tiles))*math.Pow(math.Min(1, s/nd.cap), 0.7))
}

// irrigate draws each irrigated field's need from the river its channel taps,
// as far as the river can give it, and remembers how well it was served.
func (e *Ecology) irrigate(rain, pet, yr float64) {
	h := &e.hydro
	need := math.Max(0, irrigNeed*pet-0.8*rain) * yr
	want := make(map[int32]float64)
	for _, p := range e.plots {
		i := int(p.Tile)
		if e.manage[i]&Irrigated == 0 {
			continue
		}
		if to := h.nearest[i]; to >= 0 {
			want[to] += need
		}
	}
	got := make(map[int32]float64, len(want))
	for to, w := range want {
		nd := &h.nodes[to]
		spare := h.S[to]
		if !nd.lake {
			spare = math.Max(0, h.S[to]-0.3*nd.dead) // a channel can't drain the last puddles
		}
		take := math.Min(w, spare)
		h.S[to] -= take
		got[to] = 1
		if w > 0 {
			got[to] = take / w
		}
	}
	for _, p := range e.plots {
		i := int(p.Tile)
		if e.manage[i]&Irrigated == 0 {
			continue
		}
		served := float32(0)
		if to := h.nearest[i]; to >= 0 {
			served = float32(got[to])
			if need == 0 {
				served = 1
			}
		}
		// The soil holds water for a while: the supply eases rather than snaps.
		h.supply[i] += (served - h.supply[i]) * 0.2
	}
}

// refreshWater works out every fresh tile's state and level, and the island's totals.
func (e *Ecology) refreshWater() {
	h := &e.hydro
	changed := false
	h.flowing, h.pools, h.under, h.dry = 0, 0, 0, 0
	for id := range h.nodes {
		nd := &h.nodes[id]
		s := h.S[id]
		gw := e.groundwaterOf(int32(id))
		foul := uint8(0)
		if len(h.P) > 0 {
			foul = uint8(math.Round(clamp(h.P[id]/(s+germDilute)/FoulFull, 0, 1) * 255))
		}
		for k, t := range nd.tiles {
			st, lv := WaterDry, 0.0
			switch {
			case nd.lake:
				fill := math.Min(1, s/nd.cap)
				// The shallow edges fall dry first.
				if s > 1e-6 && math.Pow(fill, 0.7) > float64(nd.shallow[k])*0.98 {
					st, lv = WaterFlowing, 0.35+0.65*fill
				}
			case s > nd.dead*1.02:
				st, lv = WaterFlowing, 0.3+0.7*math.Min(1, (s-nd.dead)/(nd.cap-nd.dead))
			case s > 0.06*nd.dead:
				st, lv = WaterPools, 0.3*s/nd.dead
			}
			if st == WaterDry && gw >= SoakMin {
				st = WaterUnder
			}
			lvb := uint8(math.Round(clamp(lv, 0, 1) * 255))
			if h.state[t] != st || h.level[t]>>4 != lvb>>4 || h.foul[t]>>5 != foul>>5 {
				changed = true
			}
			h.state[t], h.level[t], h.foul[t] = st, lvb, foul
			switch st {
			case WaterFlowing:
				h.flowing++
			case WaterPools:
				h.pools++
			case WaterUnder:
				h.under++
			default:
				h.dry++
			}
		}
	}
	if changed {
		h.version++
	}
}

// groundwaterOf is how full the aquifer behind a node is, against an
// ordinary wet season (about 1 then; lower as a dry season wears on).
func (e *Ecology) groundwaterOf(id int32) float64 {
	nd := &e.hydro.nodes[id]
	return e.hydro.G[id] / ((nd.area + 1) * gwRef * nd.kb)
}

// waterEvents tells the observer when the rivers fail and when they run again.
func (e *Ecology) waterEvents() {
	h := &e.hydro
	total := h.flowing + h.pools + h.under + h.dry
	if total == 0 {
		return
	}
	running := float64(h.flowing) / float64(total)
	wet := float64(h.flowing+h.pools) / float64(total)
	year := e.clim.Year + 1
	switch {
	case h.warned < 1 && running < 0.5:
		h.warned = 1
		e.event("climate", fmt.Sprintf("Sungai-sungai menyusut (tahun %d): separuh alirannya sudah berhenti, tinggal genangan", year))
	case h.warned < 2 && wet < 0.25:
		h.warned = 2
		e.event("climate", fmt.Sprintf("Kekeringan (tahun %d): sebagian besar sungai kering; orang menggali dasar sungai dan bergantung pada sumur", year))
	case h.warned >= 1 && h.warned < 3 && running > 0.7:
		h.warned = 3
		e.event("climate", fmt.Sprintf("Hujan kembali (tahun %d): sungai-sungai mengalir lagi", year))
	case h.warned == 3 && running > 0.85:
		h.warned = 0
	}
}

// --- what the simulation reads and does ---------------------------------------

// WaterState is the state of fresh water tile i (WaterDry for anything else).
func (e *Ecology) WaterState(i int) uint8 {
	if i < 0 || i >= len(e.hydro.state) || e.hydro.nodeOf[i] < 0 {
		return WaterDry
	}
	return e.hydro.state[i]
}

// Drink takes units of water from fresh water tile i: from the river or lake
// while it holds water, else from the sand under its dry bed.
func (e *Ecology) Drink(i int, units float64) {
	h := &e.hydro
	if i < 0 || i >= len(h.nodeOf) || h.nodeOf[i] < 0 {
		return
	}
	id := h.nodeOf[i]
	if h.state[i] >= WaterPools {
		h.W[id] += units
	} else {
		h.U[id] += units
	}
}

// Groundwater is how full the aquifer under tile i is (1 ≈ an ordinary wet
// season), as a well dug there would find it.
func (e *Ecology) Groundwater(i int) float64 {
	h := &e.hydro
	if i < 0 || i >= len(h.drainTo) {
		return 0
	}
	id := h.drainTo[i]
	if id < 0 {
		id = h.nearest[i]
	}
	if id < 0 {
		// The coast: a lens of fresh water under the sand, following the year's rain.
		return clamp(e.clim.Slow*moistFull, 0, 1.5)
	}
	return e.groundwaterOf(id)
}

// DrawGroundwater takes units of water from the aquifer under tile i (a well).
func (e *Ecology) DrawGroundwater(i int, units float64) {
	h := &e.hydro
	if i < 0 || i >= len(h.drainTo) {
		return
	}
	id := h.drainTo[i]
	if id < 0 {
		id = h.nearest[i]
	}
	if id >= 0 {
		h.U[id] += units
	}
}

// WaterVersion changes whenever a fresh water tile changes state or level noticeably.
func (e *Ecology) WaterVersion() int64 { return e.hydro.version }

// WaterTiles lists the fresh water tiles in index order, with their state,
// level and how foul they are (0–255), for the observer's map.
func (e *Ecology) WaterTiles() (tiles []int32, state, level, foul []uint8) {
	h := &e.hydro
	for i, id := range h.nodeOf {
		if id >= 0 {
			tiles = append(tiles, int32(i))
			state = append(state, h.state[i])
			level = append(level, h.level[i])
			foul = append(foul, h.foul[i])
		}
	}
	return
}

// WaterSummary is the island's fresh water at a glance.
type WaterSummary struct {
	Flowing int `json:"flowing"` // river and lake tiles with running or standing open water
	Pools   int `json:"pools"`
	Under   int `json:"under"` // dry beds with water in the sand
	Dry     int `json:"dry"`
	// Groundwater is the median aquifer level (1 ≈ an ordinary wet season).
	Groundwater float64 `json:"groundwater"`
	// Lakes is how full the lakes are together (0–1); -1 without lakes.
	Lakes float64 `json:"lakes"`
}

func (e *Ecology) Water() WaterSummary {
	h := &e.hydro
	w := WaterSummary{Flowing: h.flowing, Pools: h.pools, Under: h.under, Dry: h.dry, Lakes: -1}
	if len(h.nodes) == 0 {
		return w
	}
	gws := make([]float64, len(h.nodes))
	var lakeS, lakeC float64
	for id := range h.nodes {
		gws[id] = e.groundwaterOf(int32(id))
		if h.nodes[id].lake {
			lakeS += h.S[id]
			lakeC += h.nodes[id].cap
		}
	}
	slices.Sort(gws)
	w.Groundwater = math.Round(gws[len(gws)/2]*1000) / 1000
	if lakeC > 0 {
		w.Lakes = math.Round(math.Min(1, lakeS/lakeC)*1000) / 1000
	}
	return w
}

// riverWetness is how well the river nearest tile i keeps its banks wet:
// 1 while its aquifer is healthy, down to 0.3 when it has sunk away.
func (e *Ecology) riverWetness(i int) float64 {
	h := &e.hydro
	if len(h.nodes) == 0 {
		return 1
	}
	id := h.nearest[i]
	if id < 0 {
		return 1
	}
	return clamp(0.3+0.7*e.groundwaterOf(id)/0.35, 0.3, 1)
}

// irrigation is how well an irrigated tile is watered (0–1): what its
// field got lately, or for an empty one whether its river has water to give.
func (e *Ecology) irrigation(i int) float64 {
	h := &e.hydro
	if len(h.nodes) == 0 {
		return 1
	}
	if e.plotAt[i] != 0 {
		return float64(h.supply[i])
	}
	id := h.nearest[i]
	if id < 0 {
		return 0
	}
	if nd := &h.nodes[id]; nd.lake || h.S[id] > 0.3*nd.dead {
		return 1
	}
	return 0
}

// hydroState is the saved water cycle.
type hydroState struct {
	Nodes int       `json:"nodes"`
	S     []float64 `json:"s"`
	G     []float64 `json:"g"`
	W     []float64 `json:"w,omitempty"`
	U     []float64 `json:"u,omitempty"`
	// Irrigation supply per irrigated tile, as tile → supply pairs.
	Supply map[int32]float32 `json:"supply,omitempty"`
	Worst  float64           `json:"worst,omitempty"`
	Warned int               `json:"warned,omitempty"`
	// Germs in the water (Fase 3b), and those washed in since the last step.
	P   []float64 `json:"p,omitempty"`
	Pin []float64 `json:"pin,omitempty"`
}

func (e *Ecology) saveHydrology() *hydroState {
	h := &e.hydro
	st := &hydroState{Nodes: len(h.nodes), S: h.S, G: h.G, Worst: h.worst, Warned: h.warned}
	if slices.ContainsFunc(h.W, func(v float64) bool { return v != 0 }) {
		st.W = h.W
	}
	if slices.ContainsFunc(h.U, func(v float64) bool { return v != 0 }) {
		st.U = h.U
	}
	if slices.ContainsFunc(h.P, func(v float64) bool { return v != 0 }) {
		st.P = h.P
	}
	if slices.ContainsFunc(h.Pin, func(v float64) bool { return v != 0 }) {
		st.Pin = h.Pin
	}
	for i, v := range h.supply {
		if v != 0 {
			if st.Supply == nil {
				st.Supply = map[int32]float32{}
			}
			st.Supply[int32(i)] = v
		}
	}
	return st
}

// restoreHydrology takes the saved water back, or fills the land afresh if
// the save has none (from before the water cycle) or the map has changed.
func (e *Ecology) restoreHydrology(st *hydroState) {
	h := &e.hydro
	if st == nil || st.Nodes != len(h.nodes) || len(st.S) != len(h.nodes) || len(st.G) != len(h.nodes) {
		e.fillHydrology()
		return
	}
	copy(h.S, st.S)
	copy(h.G, st.G)
	if len(st.W) == len(h.W) {
		copy(h.W, st.W)
	}
	if len(st.U) == len(h.U) {
		copy(h.U, st.U)
	}
	if len(st.P) == len(h.P) {
		copy(h.P, st.P)
	}
	if len(st.Pin) == len(h.Pin) {
		copy(h.Pin, st.Pin)
	}
	for i, v := range st.Supply {
		if int(i) < len(h.supply) {
			h.supply[i] = v
		}
	}
	h.worst, h.warned = st.Worst, st.Warned
	e.refreshWater()
}
