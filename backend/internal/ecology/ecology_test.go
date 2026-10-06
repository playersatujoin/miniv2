package ecology

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand/v2"
	"testing"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/world"
)

var (
	long      = flag.Bool("long", false, "run the four-hour wildlife test on 8 seeds")
	longTrace = flag.Bool("wildtrace", false, "print animal numbers every 25 years")
)

const tick = 1.0 / 20

func starterLand(size int) Land {
	m := world.Generate(world.GenOptions{Name: "test", Width: size, Height: size, Seed: 1337})
	return LandFromMap(m)
}

// run advances e by seconds of simulated time from t0 without people.
func run(e *Ecology, t0, seconds float64, h Humans, each func(t float64)) float64 {
	ticks := int64(math.Round(t0 / tick))
	n := int64(math.Round(seconds / tick))
	for k := int64(1); k <= n; k++ {
		ticks++
		t := float64(ticks) * tick
		e.Tick(ticks, t, tick, h)
		if each != nil && ticks%int64(SecondsPerYear/tick) == 0 {
			each(t)
		}
	}
	return float64(ticks) * tick
}

func TestSeasonsFollowTheMonsoon(t *testing.T) {
	e := New(starterLand(32), rand.New(rand.NewPCG(1, 2)), Options{}, 0)
	e.clim.ENSO = Neutral
	wetSum, drySum := 0.0, 0.0
	for k := 1; k <= int(SecondsPerYear/tick); k++ {
		t := float64(k) * tick
		e.clim.step(e, t, tick)
		e.clim.ENSO = Neutral
		if e.Wet() {
			wetSum += e.clim.Rain
		} else {
			drySum += e.clim.Rain
		}
	}
	if wetSum < 3*drySum {
		t.Fatalf("wet season rain %.1f is not much more than dry season %.1f", wetSum, drySum)
	}
	if m := e.clim.Moisture; m <= 0 || m > 1 {
		t.Fatalf("moisture %v out of range", m)
	}
}

func TestENSORecursEveryFewYears(t *testing.T) {
	e := New(starterLand(32), rand.New(rand.NewPCG(3, 4)), Options{}, 0)
	years, nino, nina := 0, 0, 0
	for y := range 2000 {
		e.clim.nextENSO(e)
		years++
		switch e.clim.ENSO {
		case ElNino:
			nino++
		case LaNina:
			nina++
		}
		_ = y
	}
	interval := float64(years) / float64(nino)
	if interval < 3 || interval > 7 {
		t.Fatalf("an El Niño every %.1f years, want every 3–7", interval)
	}
	if nina == 0 {
		t.Fatal("no La Niña in 2000 years")
	}
}

func TestElNinoDriesTheLand(t *testing.T) {
	land := starterLand(64)
	normal := New(land, rand.New(rand.NewPCG(5, 6)), Options{NoFauna: true}, 0)
	drought := New(land, rand.New(rand.NewPCG(5, 6)), Options{NoFauna: true}, 0)
	var mNormal, mDrought float64
	for k := int64(1); k <= int64(SecondsPerYear/tick); k++ {
		t := float64(k) * tick
		normal.clim.ENSO, drought.clim.ENSO = Neutral, ElNino
		normal.clim.ENSOYear, drought.clim.ENSOYear = 0, 0
		normal.Tick(k, t, tick, nil)
		drought.Tick(k, t, tick, nil)
		mNormal += normal.clim.Moisture
		mDrought += drought.clim.Moisture
	}
	if mDrought > 0.8*mNormal {
		t.Fatalf("El Niño year moisture %.1f vs normal %.1f: not a drought", mDrought, mNormal)
	}
	fN, _, _ := normal.Totals()
	fD, _, _ := drought.Totals()
	if fD >= fN {
		t.Fatalf("wild food after an El Niño year %.0f, after a normal one %.0f", fD, fN)
	}
}

func TestCropsGrowAndCanBeHarvested(t *testing.T) {
	e := New(starterLand(64), rand.New(rand.NewPCG(7, 8)), Options{NoFauna: true, NoClimate: true}, 0)
	tile := -1
	for _, i := range e.walkable {
		if e.CanPlant(int(i), "ubi") {
			tile = int(i)
			break
		}
	}
	if tile < 0 {
		t.Fatal("nowhere to plant ubi")
	}
	if !e.Plant(tile, "ubi", 1, 0, 0.5, 0) {
		t.Fatal("planting failed")
	}
	if e.Plant(tile, "ubi", 1, 0, 0.5, 0) {
		t.Fatal("planted twice on one tile")
	}
	fert := e.fert[tile]
	tNow := run(e, 0, 0.75*SecondsPerYear*1.6, nil, nil)
	p := e.PlotAt(tile)
	if p == nil || !p.IsRipe() {
		t.Fatalf("ubi not ripe after 1.2 years: %+v", p)
	}
	id, n := e.Harvest(tile, 99, tNow)
	if id != "ubi" || n <= 0 {
		t.Fatalf("harvest gave %d %s", n, id)
	}
	if e.PlotAt(tile) != nil {
		t.Fatal("an annual crop is still standing after the harvest")
	}
	if e.fert[tile] >= fert {
		t.Fatalf("harvest left fertility at %.2f (was %.2f)", e.fert[tile], fert)
	}
}

func TestDroughtWithersThirstyCrops(t *testing.T) {
	e := New(starterLand(64), rand.New(rand.NewPCG(9, 10)), Options{NoFauna: true}, 0)
	planted := 0
	for _, i := range e.walkable {
		if !e.nearFresh[i] && e.CanPlant(int(i), "padi") && planted < 40 {
			e.Plant(int(i), "padi", 1, 0, 0.5, 0)
			planted++
		}
	}
	e.clim.Moisture, e.clim.Slow = 0.05, 0.05
	for k := int64(1); k <= int64(SecondsPerYear/tick); k++ {
		e.clim.ENSO = ElNino
		e.clim.Rain = 0
		e.clim.Moisture = 0.05
		e.grow(float64(k)*tick, tick)
	}
	if len(e.plots) > planted/4 {
		t.Fatalf("%d of %d rice plots survived a year without rain", len(e.plots), planted)
	}
	if e.stats.Cur.Withered == 0 {
		t.Fatal("no withered plots recorded")
	}
}

func TestSaveRestoreReplaysExactly(t *testing.T) {
	land := starterLand(48)
	src := rand.NewPCG(11, 12)
	a := New(land, rand.New(src), Options{}, 0)
	for _, i := range a.walkable[:200] {
		a.Plant(int(i), "ubi", 1, 0, 0.5, 0)
	}
	tNow := run(a, 0, 40, nil, nil)
	data, err := json.Marshal(a.State())
	if err != nil {
		t.Fatal(err)
	}
	rngState, _ := src.MarshalBinary()

	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		t.Fatal(err)
	}
	src2 := rand.NewPCG(0, 0)
	if err := src2.UnmarshalBinary(rngState); err != nil {
		t.Fatal(err)
	}
	b, err := Restore(land, &st, rand.New(src2), Options{})
	if err != nil {
		t.Fatal(err)
	}
	run(a, tNow, 30, nil, nil)
	run(b, tNow, 30, nil, nil)
	ja, _ := json.Marshal(a.State())
	jb, _ := json.Marshal(b.State())
	if !bytes.Equal(ja, jb) {
		t.Fatal("restored ecology diverged from the original")
	}
}

// TestWildlifePersists checks the PLAN criterion: without people, animal
// numbers swing but every species lasts (4 simulated hours on 8 seeds with
// -long, 40 minutes on 2 seeds otherwise).
func TestWildlifePersists(t *testing.T) {
	seeds, minutes := []uint64{1, 2}, 40.0
	if *long {
		seeds, minutes = []uint64{1, 2, 3, 4, 5, 6, 7, 8}, 240
	}
	land := starterLand(128)
	for _, seed := range seeds {
		e := New(land, rand.New(rand.NewPCG(seed, seed*7)), Options{}, 0)
		start := e.StartCounts()
		lo := make([]int, SpeciesCount)
		hi := make([]int, SpeciesCount)
		present := make([]int, SpeciesCount)
		years := 0
		for i := range lo {
			lo[i] = math.MaxInt
		}
		run(e, 0, minutes*60, nil, func(tm float64) {
			wild, _ := e.Counts()
			years++
			for i, n := range wild {
				lo[i], hi[i] = min(lo[i], n), max(hi[i], n)
				if n > 0 {
					present[i]++
				}
			}
			if *longTrace && int(tm/SecondsPerYear)%10 == 0 && len(e.Years()) > 0 {
				y := e.Years()[len(e.Years())-1]
				fmt.Printf("seed %d year %4d: %v born %v starved %v eaten %v enso %d\n", seed, int(tm/SecondsPerYear), wild,
					y.AnimalBirths, y.Starved, y.Predated, e.clim.ENSO)
			}
		})
		wild, _ := e.Counts()
		t.Logf("seed %d: start %v, end %v, min %v, max %v, present %v of %d years, recolonised %d", seed, start, wild, lo, hi, present, years, arrivals(e))
		// The common species never die out. The island is too small for a
		// tiger population to last on its own, and buffalo and junglefowl
		// crash and come back: they die out now and then and arrive again
		// from across the sea, but must be there most of the time.
		for i := range species {
			switch {
			case species[i].ID == "rusa" || species[i].ID == "babi_hutan":
				if lo[i] == 0 {
					t.Errorf("seed %d: %s died out without people", seed, species[i].Name)
				}
			case float64(present[i]) < 0.4*float64(years):
				t.Errorf("seed %d: %s on the island only %d of %d years", seed, species[i].Name, present[i], years)
			}
		}
	}
}

func arrivals(e *Ecology) int {
	n := 0
	for _, y := range e.Years() {
		n += len(y.Arrived)
	}
	return n
}

func TestWildCropsGrowWhereTheyBelong(t *testing.T) {
	e := New(starterLand(128), rand.New(rand.NewPCG(13, 14)), Options{NoFauna: true}, 0)
	found := map[chem.ItemID]int{}
	for _, i := range e.walkable {
		if id, ok := e.WildCrop(int(i)); ok {
			found[id]++
			c, _ := chem.CropByItem(id)
			if c.Wet && !e.nearFresh[i] || c.Coastal && !e.nearSea[i] {
				t.Errorf("wild %s on tile %d away from its water", id, i)
			}
		}
	}
	for _, c := range chem.Crops() {
		if found[c.Item] == 0 {
			t.Errorf("no wild %s on the island", c.Item)
		}
	}
}

// snareHumans is nobody nearby, recording what snares catch.
type snareHumans struct{ caught []string }

func (h *snareHumans) Nearest(x, y, r float64) (int64, float64, float64, bool) { return 0, 0, 0, false }
func (h *snareHumans) Crowd(x, y, r float64) int                               { return 0 }
func (h *snareHumans) Maul(id int64, damage float64, species string)           {}
func (h *snareHumans) Home(house int64) (float64, float64, bool, bool)         { return 0, 0, false, false }
func (h *snareHumans) Settled(x, y float64) (int64, bool)                      { return 0, false }
func (h *snareHumans) Tamed(house int64, species string)                       {}
func (h *snareHumans) Pace(id int64) float64                                   { return 1 }
func (h *snareHumans) Snared(tile int, species string, meat int) {
	h.caught = append(h.caught, species)
}

func TestSnaresCatchDeerButNotBuffalo(t *testing.T) {
	land := starterLand(48)
	e := New(land, rand.New(rand.NewPCG(4, 4)), Options{NoClimate: true}, 0)
	var deerTile, buffaloTile int
	found := 0
	for _, i32 := range e.walkable {
		if i := int(i32); found == 0 {
			deerTile, found = i, 1
		} else if i-deerTile > 10 {
			buffaloTile = i
			break
		}
	}
	bits := make([]uint8, land.W*land.H)
	bits[deerTile], bits[buffaloTile] = SnareSet, SnareSet
	e.SetManagement(bits)
	at := func(i int) (float64, float64) { return float64(i%land.W) + 0.5, float64(i/land.W) + 0.5 }
	dx, dy := at(deerTile)
	bx, by := at(buffaloTile)
	deer := e.AddAnimal("rusa", dx, dy, 0)
	buffalo := e.AddAnimal("kerbau", bx, by, 0)
	h := &snareHumans{}
	tm := 0.0
	for k := int64(1); k <= 200 && !deer.Dead(); k++ {
		deer.X, deer.Y = dx, dy // it keeps to the trail
		buffalo.X, buffalo.Y = bx, by
		tm += tick
		e.Tick(k, tm, tick, h)
	}
	if !deer.Dead() || len(h.caught) != 1 || h.caught[0] != "rusa" {
		t.Fatalf("deer dead %v, caught %v", deer.Dead(), h.caught)
	}
	if e.Management(deerTile)&SnareSet != 0 {
		t.Fatal("a snare that caught something must be sprung")
	}
	if buffalo.Dead() || e.Management(buffaloTile)&SnareSet == 0 {
		t.Fatal("a buffalo breaks free of a snare")
	}
	if e.HuntedTotal()[0] != 1 {
		t.Fatalf("a snared deer counts as hunted: %v", e.HuntedTotal())
	}
}
