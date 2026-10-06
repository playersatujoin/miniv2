package ecology

import (
	"bytes"
	"encoding/json"
	"math"
	"math/rand/v2"
	"testing"

	"miniv2/backend/internal/world"
)

// grassLand is a w×h meadow of grass with nothing else on it.
func grassLand(w, h int) Land {
	n := w * h
	l := Land{W: w, H: h, Ground: make([]int, n), Objects: make([]int, n), Blocked: make([]bool, n),
		Water: make([]bool, n), NearWater: make([]bool, n), Fresh: make([]bool, n)}
	for i := range n {
		l.Ground[i] = world.Grass
	}
	return l
}

// dryMeadow is a grass meadow in a dry season with a steady wind.
func dryMeadow(t *testing.T, opts Options, windDir, wind float64) *Ecology {
	t.Helper()
	return dryLand(t, grassLand(61, 61), opts, windDir, wind)
}

// dryPatch is a 15×15 patch of grass in the middle of a sandy 61×61 plain.
func dryPatch(t *testing.T, windDir, wind float64) *Ecology {
	t.Helper()
	l := grassLand(61, 61)
	for i := range l.Ground {
		if x, y := i%61, i/61; x < 23 || x > 37 || y < 23 || y > 37 {
			l.Ground[i] = world.Sand
		}
	}
	return dryLand(t, l, Options{}, windDir, wind)
}

func dryLand(t *testing.T, l Land, opts Options, windDir, wind float64) *Ecology {
	t.Helper()
	opts.NoFauna = true
	e := New(l, rand.New(rand.NewPCG(1, 2)), opts, 0)
	e.clim.Moisture, e.clim.Slow, e.clim.Rain = 0.12, 0.3, 0.15
	e.clim.WindDir, e.clim.WindSpeed = windDir, wind
	return e
}

// burnFor runs only the fire for seconds, the weather held as it is.
func burnFor(e *Ecology, seconds float64) {
	for range int(seconds / (fireEvery * tick)) {
		e.fire.tick += fireEvery
		e.burn(fireEvery * tick)
	}
}

// reach is how far the scorched ground runs east and west of column x0 along row y0.
func reach(e *Ecology, x0, y0 int) (east, west int) {
	w := e.land.W
	for x := x0; x < w && e.fire.scorch[y0*w+x] > 0; x++ {
		east = x - x0
	}
	for x := x0; x >= 0 && e.fire.scorch[y0*w+x] > 0; x-- {
		west = x0 - x
	}
	return east, west
}

func TestWindTurnsSmoothlyWithTheMonsoon(t *testing.T) {
	a := New(starterLand(32), rand.New(rand.NewPCG(1, 2)), Options{NoFauna: true}, 0)
	b := New(starterLand(32), rand.New(rand.NewPCG(1, 2)), Options{NoFauna: true}, 0)
	other := New(starterLand(32), rand.New(rand.NewPCG(3, 4)), Options{NoFauna: true}, 0)
	var wet, dry [2]float64
	differs := false
	prev := a.clim.WindDir
	for k := 1; k <= int(30*SecondsPerYear/tick); k++ {
		tm := float64(k) * tick
		for _, e := range []*Ecology{a, b, other} {
			e.clim.step(e, tm, tick)
			e.clim.ENSO = Neutral
		}
		ca := &a.clim
		if ca.WindDir != b.clim.WindDir || ca.WindSpeed != b.clim.WindSpeed || ca.Storm != b.clim.Storm {
			t.Fatalf("tick %d: the same world's wind differs", k)
		}
		differs = differs || ca.WindDir != other.clim.WindDir
		if turn := math.Abs(wrapAngle(ca.WindDir - prev)); turn > windTurn*tick+1e-9 {
			t.Fatalf("tick %d: the wind turned %.3f rad at once", k, turn)
		}
		prev = ca.WindDir
		if ca.WindSpeed < 0 || ca.WindSpeed > 1 {
			t.Fatalf("wind speed %v", ca.WindSpeed)
		}
		v := [2]float64{ca.WindSpeed * math.Cos(ca.WindDir), ca.WindSpeed * math.Sin(ca.WindDir)}
		if a.Wet() {
			wet[0], wet[1] = wet[0]+v[0], wet[1]+v[1]
		} else {
			dry[0], dry[1] = dry[0]+v[0], dry[1]+v[1]
		}
	}
	if !differs {
		t.Fatal("two worlds have the same weather")
	}
	// The west monsoon blows towards the east in the wet season, the east
	// monsoon towards the north-west in the dry season.
	if wet[0] <= 0 || dry[0] >= 0 || dry[1] >= 0 {
		t.Fatalf("mean wind wet season %v, dry season %v: no monsoon", wet, dry)
	}
}

func TestStormsComeWithTheRains(t *testing.T) {
	count := func(enso int) (wet, dry, lightningYears float64) {
		e := New(starterLand(32), rand.New(rand.NewPCG(5, 6)), Options{NoFauna: true, NoFire: true}, 0)
		inStorm := false
		for k := 1; k <= int(200*SecondsPerYear/tick); k++ {
			e.clim.ENSO = enso
			e.clim.step(e, float64(k)*tick, tick)
			on := e.clim.StormLeft > 0
			if on && !inStorm {
				if e.Wet() {
					wet++
				} else {
					dry++
				}
			}
			inStorm = on
		}
		return wet, dry, 0
	}
	wet, dry, _ := count(Neutral)
	t.Logf("neutral: %.0f storms in the wet seasons, %.0f in the dry, of 200 years", wet, dry)
	if wet <= 1.5*dry {
		t.Fatalf("%.0f storms in the wet seasons, %.0f in the dry", wet, dry)
	}
	nWet, nDry, _ := count(LaNina)
	eWet, eDry, _ := count(ElNino)
	if nWet+nDry <= wet+dry || eWet+eDry >= wet+dry {
		t.Fatalf("storms: La Niña %.0f, neutral %.0f, El Niño %.0f", nWet+nDry, wet+dry, eWet+eDry)
	}
	// A storm blows harder than the ordinary weather.
	e := New(starterLand(32), rand.New(rand.NewPCG(7, 8)), Options{NoFauna: true}, 0)
	var calm, storm, nc, ns float64
	for k := 1; k <= int(50*SecondsPerYear/tick); k++ {
		e.clim.step(e, float64(k)*tick, tick)
		if _, w, s := e.Wind(); s {
			storm, ns = storm+w, ns+1
		} else {
			calm, nc = calm+w, nc+1
		}
	}
	if ns == 0 || storm/ns < 2*calm/nc {
		t.Fatalf("mean wind %.2f in storms, %.2f otherwise", storm/max(ns, 1), calm/nc)
	}
}

func TestFireRunsDownwind(t *testing.T) {
	e := dryMeadow(t, Options{}, 0, 0.5) // blowing east
	c := 30*61 + 30
	if !e.Ignite(c, 1, 0, FireLightning) {
		t.Fatal("dry grass did not catch")
	}
	burnFor(e, 2)
	east, west := reach(e, 30, 30)
	t.Logf("after 2 s: %d tiles east, %d west; %d burning, %d caught", east, west, len(e.fire.active), e.fire.stats.Tiles)
	if east < 2*west+3 {
		t.Fatalf("the fire ran %d tiles downwind and %d upwind", east, west)
	}
	if e.fire.stats.Started != 1 || e.fire.stats.Lightning != 1 {
		t.Fatalf("fire stats %+v", e.fire.stats)
	}
	// In still air it spreads about evenly.
	still := dryMeadow(t, Options{}, 0, 0)
	still.Ignite(c, 1, 0, FireLightning)
	burnFor(still, 2)
	if e2, w2 := reach(still, 30, 30); e2 > 2*w2+2 || w2 > 2*e2+2 {
		t.Fatalf("without wind the fire ran %d east and %d west", e2, w2)
	}
}

func TestWetFuelDoesNotBurn(t *testing.T) {
	e := dryMeadow(t, Options{}, 0, 0.3)
	e.clim.Moisture = 0.8 // the wet season
	c := 30*61 + 30
	if e.catch(c) != 0 || e.Ignite(c, 1, 0, FireLightning) {
		t.Fatal("wet grass caught fire")
	}
	// Even a fire set going (a burning house) dies out without spreading.
	e.Flare(c, 1)
	burnFor(e, 2)
	if len(e.fire.active) != 0 || e.fire.stats.Tiles != 1 {
		t.Fatalf("on wet land: %d burning, %d caught", len(e.fire.active), e.fire.stats.Tiles)
	}
}

func TestRainPutsFiresOut(t *testing.T) {
	dryRun := dryMeadow(t, Options{}, 0, 0.3)
	wetRun := dryMeadow(t, Options{}, 0, 0.3)
	for _, e := range []*Ecology{dryRun, wetRun} {
		for dx := -2; dx <= 2; dx++ {
			e.Flare(30*61+30+dx, 0.8)
		}
		burnFor(e, 0.3)
	}
	// A storm's downpour breaks over one of them.
	wetRun.clim.Storm, wetRun.clim.StormLeft, wetRun.clim.Rain = 1, 1, 0.6
	burnFor(dryRun, 1)
	burnFor(wetRun, 1)
	if len(wetRun.fire.active) != 0 {
		t.Fatalf("%d tiles still burn in the downpour", len(wetRun.fire.active))
	}
	if len(dryRun.fire.active) == 0 || dryRun.fire.stats.Tiles <= 2*wetRun.fire.stats.Tiles {
		t.Fatalf("dry: %d burning, %d caught; storm: %d caught", len(dryRun.fire.active), dryRun.fire.stats.Tiles, wetRun.fire.stats.Tiles)
	}
}

func TestFireBurnsOutAndTheLandGreensAgain(t *testing.T) {
	e := dryPatch(t, 0, 0.3)
	c := 30*61 + 30
	fert, graze := e.fert[c], e.graze[c]
	v0 := e.BurntVersion()
	e.Ignite(c, 1, 0, FireHearth)
	burnFor(e, 30)
	if len(e.fire.active) != 0 {
		t.Fatalf("%d tiles still burn after 30 s", len(e.fire.active))
	}
	if e.fire.stats.Tiles < 100 || e.fire.stats.Tiles > 15*15+40 || e.fire.stats.Hearth != 1 {
		t.Fatalf("fire stats %+v", e.fire.stats)
	}
	if e.fire.scorch[c] < 0.5 || e.graze[c] > graze/2 || e.fert[c] <= fert || e.fire.ash[c] <= 0 {
		t.Fatalf("burnt tile: scorch %.2f, graze %.2f (was %.2f), fertility %.2f (was %.2f)", e.fire.scorch[c], e.graze[c], graze, e.fert[c], fert)
	}
	if e.BurntVersion() == v0 || e.ScorchedCount() > e.fire.stats.Tiles || e.ScorchedCount() < 15*15*3/4 {
		t.Fatalf("version %d → %d, %d scorched of %d burnt", v0, e.BurntVersion(), e.ScorchedCount(), e.fire.stats.Tiles)
	}
	n := 0
	e.Scorched(func(tile int, level float32) {
		if level < 0.05 || level > 1 {
			t.Fatalf("scorch level %v", level)
		}
		n++
	})
	if n != e.ScorchedCount() {
		t.Fatal("Scorched and ScorchedCount disagree")
	}
	// The rains come: the scar fades within a year or so, and the ash's
	// nutrients wash out over a few.
	e.clim.Moisture = 1
	for i := range e.moistNow {
		e.moistNow[i] = 1
	}
	for k := 1; k <= int(8*SecondsPerYear/(growEvery*tick)); k++ {
		e.fadeScorch(growEvery * tick)
		switch years := float64(k) * growEvery * tick / SecondsPerYear; {
		case years == 0.25 && e.ScorchedCount() == 0:
			t.Fatal("the scar was gone within three months")
		case years == 1 && e.ScorchedCount() != 0:
			t.Fatalf("%d tiles still scorched after a wet year", e.ScorchedCount())
		case years == 1 && e.fert[c] <= fert:
			t.Fatal("the ash's nutrients were gone within a year")
		}
	}
	if e.fire.scorch[c] != 0 || e.fire.ash[c] != 0 || len(e.fire.marked) != 0 || e.ScorchedCount() != 0 {
		t.Fatalf("after eight wet years: scorch %v, ash %v, %d marked", e.fire.scorch[c], e.fire.ash[c], len(e.fire.marked))
	}
	if math.Abs(float64(e.fert[c]-fert)) > 1e-3 {
		t.Fatalf("fertility %.4f after the ash washed out, was %.4f", e.fert[c], fert)
	}
}

func TestFireBurnsCropsAndRaisesNoFireInTheWetSeason(t *testing.T) {
	e := dryMeadow(t, Options{}, 0, 0.3)
	c := 30*61 + 30
	for dx := 1; dx <= 3; dx++ {
		if !e.Plant(c+dx, "ubi", 1, 0, 0.5, 0) {
			t.Fatal("could not plant")
		}
	}
	e.Flare(c, 1)
	burnFor(e, 5)
	if len(e.plots) != 0 || e.fire.stats.Crops != 3 || e.stats.Cur.CropLoss != 3 {
		t.Fatalf("%d plots left, %d burnt", len(e.plots), e.fire.stats.Crops)
	}
}

func TestLightningSetsDryLandAlight(t *testing.T) {
	strikeFires := func(moisture float64, opts Options) int {
		e := dryMeadow(t, opts, 0, 0.3)
		for k := 1; k <= 2000; k++ {
			tm := float64(k) * tick
			e.clim.Storm, e.clim.StormLeft, e.clim.StormDry = 1, 1, true
			e.clim.Moisture = moisture
			e.lightning(tm, tick)
		}
		return e.fire.stats.Lightning
	}
	if n := strikeFires(0.1, Options{}); n == 0 {
		t.Fatal("no lightning fire in a dry storm over parched grass")
	}
	if n := strikeFires(0.8, Options{}); n != 0 {
		t.Fatalf("%d lightning fires in the wet season", n)
	}
	if n := strikeFires(0.1, Options{NoFire: true}); n != 0 {
		t.Fatalf("%d lightning fires with fire switched off", n)
	}
}

func TestNoFireLeavesTheWind(t *testing.T) {
	e := dryMeadow(t, Options{NoFire: true}, 0, 0.5)
	c := 30*61 + 30
	if e.Ignite(c, 1, 0, FireLightning) || e.Flare(c, 1) || len(e.Burning()) != 0 || e.FireDanger() != 0 {
		t.Fatal("something burns with fire switched off")
	}
	run(e, 0, 2, nil, nil)
	if _, w, _ := e.Wind(); w <= 0 {
		t.Fatal("no wind with fire switched off")
	}
	calm := New(grassLand(16, 16), rand.New(rand.NewPCG(1, 2)), Options{NoFauna: true, NoClimate: true}, 0)
	for k := int64(1); k <= 400; k++ {
		calm.Tick(k, float64(k)*tick, tick, nil)
		if d, w, s := calm.Wind(); d != 0 || w != windAverage || s {
			t.Fatalf("without a climate the wind is %v, %v, storm %v", d, w, s)
		}
	}
}

func TestFireSaveRestoreContinuesExactly(t *testing.T) {
	land := starterLand(48)
	src := rand.NewPCG(13, 14)
	// A parched July in an El Niño year.
	t0 := 10*SecondsPerYear + 0.55*SecondsPerYear
	a := New(land, rand.New(src), Options{}, t0)
	a.clim.ENSO, a.clim.ENSOYear = ElNino, 10
	a.clim.Moisture, a.clim.Slow = 0.08, 0.2
	lit := 0
	for _, i := range a.walkable {
		if a.cover[i] == coverGrass && a.catch(int(i)) > 0.3 && lit < 6 {
			a.Flare(int(i), 0.8)
			lit++
		}
	}
	tNow := run(a, t0, 0.6, nil, nil)
	if len(a.fire.active) == 0 {
		t.Fatal("nothing burns when the world is saved")
	}
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
	if b.clim.WindDir != a.clim.WindDir || b.clim.Seed != a.clim.Seed || len(b.fire.active) != len(a.fire.active) {
		t.Fatal("wind or fire lost in saving")
	}
	run(a, tNow, 3, nil, nil)
	run(b, tNow, 3, nil, nil)
	ja, _ := json.Marshal(a.State())
	jb, _ := json.Marshal(b.State())
	if !bytes.Equal(ja, jb) {
		t.Fatal("the restored fire went differently")
	}
	if a.fire.stats.Tiles <= lit {
		t.Fatalf("the fire never spread (%d tiles)", a.fire.stats.Tiles)
	}
}
