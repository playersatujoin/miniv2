package ecology

import (
	"fmt"
	"math/rand/v2"
	"os"
	"testing"

	"miniv2/backend/internal/world"
)

func starterIsland() *world.Map {
	m := world.Generate(world.GenOptions{Name: "Starter Island", Width: 128, Height: 128, Seed: 1337})
	m.ID = "0000000000000000"
	return m
}

// TestHydrologyReport prints, month by month, how much of the island's fresh
// water runs, stands in pools or has dried up (HYDRO_REPORT=1 to run).
func TestHydrologyReport(t *testing.T) {
	if os.Getenv("HYDRO_REPORT") == "" {
		t.Skip("set HYDRO_REPORT=1 to print the water calendar")
	}
	years := 14
	e := New(LandFromMap(starterIsland()), rand.New(rand.NewPCG(5, 6)), Options{NoFauna: true}, 0)
	h := &e.hydro
	rivers, lakes := 0, 0
	for _, nd := range h.nodes {
		if nd.lake {
			lakes++
		} else {
			rivers++
		}
	}
	fmt.Printf("nodes: %d river tiles, %d lakes\n", rivers, lakes)
	const dt = 1.0 / 20
	const perYear = int64(SecondsPerYear / dt)
	tick := int64(0)
	for y := 0; y < years; y++ {
		line := fmt.Sprintf("year %2d ", y+1)
		for mo := 0; mo < 12; mo++ {
			for tick < int64(y)*perYear+int64(mo+1)*perYear/12 {
				tick++
				e.Tick(tick, float64(tick)*dt, dt, nil)
			}
			w := e.Water()
			total := float64(w.Flowing + w.Pools + w.Under + w.Dry)
			line += fmt.Sprintf(" %3.0f/%2.0f/%2.0f", 100*float64(w.Flowing)/total, 100*float64(w.Pools)/total, 100*float64(w.Dry)/total)
			if mo == 8 {
				line += fmt.Sprintf("[gw%.2f lk%.2f]", w.Groundwater, w.Lakes)
			}
		}
		enso := map[int]string{-1: "La Niña", 0: "netral", 1: "El Niño"}[e.clim.ENSO]
		fmt.Println(line, enso)
	}
}

func TestRiversShrinkInTheDrySeasonAndRecover(t *testing.T) {
	e := New(LandFromMap(starterIsland()), rand.New(rand.NewPCG(5, 6)), Options{NoFauna: true}, 0)
	const dt = 1.0 / 20
	running := func() float64 {
		w := e.Water()
		return float64(w.Flowing) / float64(w.Flowing+w.Pools+w.Under+w.Dry)
	}
	// The fullest February and the emptiest September of three years (an El
	// Niño delays the rains, so not every February is full).
	wetSeason, drySeason := 0.0, 1.0
	tick := int64(0)
	for range int(3 * SecondsPerYear / dt) {
		tick++
		e.Tick(tick, float64(tick)*dt, dt, nil)
		switch phase := e.clim.Phase; {
		case phase > 0.08 && phase < 0.09:
			wetSeason = max(wetSeason, running())
		case phase > 0.70 && phase < 0.71:
			drySeason = min(drySeason, running())
		}
	}
	if wetSeason < 0.9 {
		t.Errorf("only %.0f%% of the fresh water runs in February", 100*wetSeason)
	}
	if drySeason >= wetSeason-0.1 {
		t.Errorf("rivers don't shrink in the dry season: %.0f%% running in September vs %.0f%% in February", 100*drySeason, 100*wetSeason)
	}
}

func TestDrinkingEmptiesPools(t *testing.T) {
	e := New(LandFromMap(starterIsland()), rand.New(rand.NewPCG(5, 6)), Options{NoFauna: true}, 0)
	h := &e.hydro
	// A river tile standing as a pool, no rain, a crowd drinking from it.
	var id int32 = -1
	for k, nd := range h.nodes {
		if !nd.lake {
			id = int32(k)
			break
		}
	}
	nd := &h.nodes[id]
	h.S[id] = nd.dead * 0.9
	for k := range h.S {
		if int32(k) != id {
			h.S[k] = 0
		}
	}
	h.G[id] = 0
	nd.area = 0
	e.clim.Off, e.clim.Rain, e.clim.Moisture = true, 0, 0
	tile := int(nd.tiles[0])
	e.refreshWater()
	if e.WaterState(tile) != WaterPools {
		t.Fatalf("set-up: tile state %d, want pools", e.WaterState(tile))
	}
	for range 200 {
		e.Drink(tile, 0.05)
		e.stepHydrology(0.1)
	}
	if st := e.WaterState(tile); st >= WaterPools {
		t.Errorf("a crowd drank for weeks and the pool is still there (state %d, %.3f units)", st, h.S[id])
	}
}
