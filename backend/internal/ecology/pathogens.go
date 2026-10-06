package ecology

import "math"

// Germs and mosquitoes (Fase 3b): what the land and its water do with them.
// The people side (who falls ill, who is immune) is in sim/disease.go.
//
// Faeces left near the water wash into it, mostly with the rain, and the
// germs in them (the bacteria, viruses and protozoa of diarrhoeal disease)
// travel with the river: a camp upstream fouls the water of everyone below
// it. Germs die off in the water within weeks (sunlight and time), flowing
// water carries them away, and a shrinking dry-season pool concentrates
// them, which is why drinking from the last pools is so dangerous. A well
// draws filtered groundwater; a hole dug in a dry bed lets the sand filter
// most of them out.
//
// Mosquitoes that carry malaria (Anopheles) lay their eggs in still, sunlit
// water: the pools a drying river leaves, the weedy shallows of lakes, the
// puddles of the rainy season and, above all, flooded rice fields. Fast
// streams and deep water are poor nurseries, and few breed in the cool
// uplands.

const (
	germLife = 0.04 // years for most germs in water to die (a couple of weeks)
	// germDilute is the water (units) that a node's germs are mixed into
	// besides its own, so a nearly dry pool doesn't turn into pure poison.
	germDilute = 0.3
	washReach  = 3.0 // tiles over which filth on the ground is washed into the water

	// FoulFull is the germ level shown as the foulest water.
	FoulFull = 2.0
)

// carryGerms moves the germs in the water on by yr years: those washed in
// since the last step join, a share dies, and what leaves with the outflow
// (share of each node's water that left, nil if it isn't tracked) joins the
// node downstream.
func (e *Ecology) carryGerms(yr float64, outflow []float64) {
	h := &e.hydro
	if len(h.P) == 0 {
		return
	}
	survive := math.Exp(-yr / germLife)
	carried := make([]float64, len(h.nodes))
	for _, id := range h.order {
		nd := &h.nodes[id]
		p := (h.P[id]+h.Pin[id])*survive + carried[id]
		h.Pin[id] = 0
		out := 0.0
		switch {
		case outflow != nil:
			out = outflow[id]
		case !nd.lake:
			// Rivers kept full (no water cycle) always run.
			out = 1 - math.Exp(-yr/kChannel)
		}
		if nd.down >= 0 && out > 0 {
			carried[nd.down] += p * out
		}
		p *= 1 - out
		if p < 1e-9 {
			p = 0
		}
		h.P[id] = p
	}
}

// Pollute leaves amount of filth on tile i. What lies within a few tiles of
// a river or lake washes into it, most when it rains.
func (e *Ecology) Pollute(i int, amount float64) {
	h := &e.hydro
	if i < 0 || i >= len(h.nearest) || len(h.Pin) == 0 {
		return
	}
	id := h.nearest[i]
	if id < 0 {
		return
	}
	wash := math.Exp(-float64(h.reach[i])/washReach) * (0.25 + 0.75*clamp(e.clim.Rain/moistFull, 0, 1))
	h.Pin[id] += amount * wash
}

// Germs is how foul the water of fresh water tile i is (0 = clean; about 1
// for a pool shared by a small camp without latrines).
func (e *Ecology) Germs(i int) float64 {
	h := &e.hydro
	if i < 0 || i >= len(h.nodeOf) || len(h.P) == 0 {
		return 0
	}
	id := h.nodeOf[i]
	if id < 0 {
		return 0
	}
	return h.P[id] / (h.S[id] + germDilute)
}

// Breeding is how good a nursery for mosquito larvae tile i is now (0–1).
func (e *Ecology) Breeding(i int) float64 {
	h := &e.hydro
	if i < 0 || i >= len(e.fresh) {
		return 0
	}
	b := 0.0
	switch {
	case e.fresh[i] && len(h.nodeOf) > 0 && h.nodeOf[i] >= 0:
		nd := &h.nodes[h.nodeOf[i]]
		switch h.state[i] {
		case WaterPools:
			b = 1
		case WaterFlowing:
			if !nd.lake {
				b = 0.08 // quiet margins of a running stream
				break
			}
			// Weedy shallows near the shore, not the open water.
			k := tileIndex(nd.tiles, int32(i))
			if k >= 0 && nd.shallow[k] > 0.6 {
				b = 0.5
			} else {
				b = 0.05
			}
		}
	case e.fresh[i]:
		b = 0.3 // standing fresh water outside the water cycle
	case e.land.Water[i] || e.land.Blocked[i]:
		return 0
	default:
		if p := e.plotAt[i]; p != 0 && e.manage[i]&Irrigated != 0 {
			// A flooded rice field is the best nursery of all.
			wet := float64(h.supplyAt(i))
			if e.plots[p-1].Crop == "padi" {
				b = 0.8 * wet
			} else {
				b = 0.25 * wet
			}
		}
		// Puddles of the rainy season, on the flat land by the rivers.
		if e.nearFresh[i] {
			b = math.Max(b, 0.12*smoothstep(1.1, 1.7, e.clim.Rain))
		}
	}
	if a := e.land.Altitude; a != nil {
		b *= 1 - smoothstep(0.3, 0.75, float64(a[i]))
	}
	return b
}

// supplyAt is how well irrigated tile i is watered (1 where the water cycle is off).
func (h *hydrology) supplyAt(i int) float32 {
	if i < len(h.supply) {
		return h.supply[i]
	}
	return 1
}

// tileIndex finds t in the sorted tiles of a node (-1 if absent).
func tileIndex(tiles []int32, t int32) int {
	lo, hi := 0, len(tiles)
	for lo < hi {
		m := (lo + hi) / 2
		if tiles[m] < t {
			lo = m + 1
		} else {
			hi = m
		}
	}
	if lo < len(tiles) && tiles[lo] == t {
		return lo
	}
	return -1
}

// Rain is the rainfall now (1 = the yearly average).
func (e *Ecology) Rain() float64 { return e.clim.Rain }
