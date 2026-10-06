package ecology

import (
	"math/rand/v2"
	"testing"
)

// TestGermsFlowDownstreamAndDie: filth at a river upstream fouls the water
// below it; once it stops, the germs die off and wash away.
func TestGermsFlowDownstreamAndDie(t *testing.T) {
	e := New(LandFromMap(starterIsland()), rand.New(rand.NewPCG(5, 6)), Options{NoFauna: true}, 0)
	h := &e.hydro
	// A river node with at least two nodes below it.
	src := int32(-1)
	for _, id := range h.order {
		nd := &h.nodes[id]
		if nd.lake || nd.down < 0 || h.nodes[nd.down].down < 0 {
			continue
		}
		src = id
		break
	}
	if src < 0 {
		t.Skip("no river long enough on this map")
	}
	below := h.nodes[src].down
	tile := int(h.nodes[src].tiles[0])
	const dt = 1.0 / 20
	tick := int64(0)
	for range 40 {
		tick++
		e.Pollute(tile, 0.5)
		e.Tick(tick, float64(tick)*dt, dt, nil)
	}
	if h.P[below] <= 0 {
		t.Fatalf("the node below the filth should carry germs, got %v (source %v)", h.P[below], h.P[src])
	}
	foulBelow := e.Germs(int(h.nodes[below].tiles[0]))
	// Two simulated years without filth: they die off and wash out.
	for range int(2 * SecondsPerYear / dt) {
		tick++
		e.Tick(tick, float64(tick)*dt, dt, nil)
	}
	if after := e.Germs(int(h.nodes[below].tiles[0])); after >= foulBelow*0.01 {
		t.Errorf("germs should die off once the filth stops: %v then %v", foulBelow, after)
	}
}

func TestBreedingFavoursStillLowlandWater(t *testing.T) {
	e := New(LandFromMap(starterIsland()), rand.New(rand.NewPCG(5, 6)), Options{NoFauna: true}, 0)
	h := &e.hydro
	pools, flowing := -1, -1
	for i, id := range h.nodeOf {
		if id < 0 || h.nodes[id].lake {
			continue
		}
		if a := e.land.Altitude; a != nil && a[i] > 0.3 {
			continue
		}
		switch h.state[i] {
		case WaterFlowing:
			flowing = i
		}
	}
	if flowing < 0 {
		t.Skip("no lowland river")
	}
	// Turn the same tile into a pool: it becomes a far better nursery.
	b0 := e.Breeding(flowing)
	h.state[flowing] = WaterPools
	pools = flowing
	if b1 := e.Breeding(pools); b1 <= b0*3 {
		t.Errorf("a pool should breed far more mosquitoes than running water: %v vs %v", b1, b0)
	}
}
