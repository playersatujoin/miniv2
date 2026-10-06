package sim

import (
	"encoding/json"
	"math"
	"os"
	"slices"
	"strconv"
	"testing"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/world"
)

func TestFrailtyIsWorstForInfantsAndTheOld(t *testing.T) {
	if !(frailty(0.5) > frailty(3) && frailty(3) > frailty(7) && frailty(7) > frailty(10)) {
		t.Errorf("frailty should fall through childhood: %v %v %v %v", frailty(0.5), frailty(3), frailty(7), frailty(10))
	}
	for _, age := range []float64{0, 2, 5, 15, 25, 40, 60} {
		if frailty(age) < frailty(10) {
			t.Errorf("frailty should be lowest around ten, but at %v it is %v < %v", age, frailty(age), frailty(10))
		}
	}
	if !(frailty(70) > frailty(50) && frailty(50) >= frailty(30)) {
		t.Errorf("frailty should rise in old age: %v %v %v", frailty(30), frailty(50), frailty(70))
	}
}

func TestLatrineCoversItsNeighbourhood(t *testing.T) {
	s := New(testMap(t, 64), 1)
	k, ok := s.cat.structure["jamban"]
	if !ok || !k.Latrine {
		t.Fatal("no latrine in the catalogue")
	}
	x, y := 30, 30
	for i := range s.terrain.blocked {
		if !s.terrain.blocked[i] && s.structAt[i] == nil {
			x, y = i%s.terrain.w, i/s.terrain.w
			break
		}
	}
	s.addStructure(k, x, y, nil)
	s.refreshSanitation()
	in, _ := s.terrain.index(x+2, y+1)
	out, _ := s.terrain.index(min(s.terrain.w-1, x+latrineReach+3), y)
	if !s.sanitary(in) {
		t.Error("a tile two steps from the latrine should use it")
	}
	if s.sanitary(out) {
		t.Error("a tile far from the latrine should not")
	}
	s.opts.NoSanitation = true
	s.epi.sanVersion = -1
	s.refreshSanitation()
	if s.sanitary(in) {
		t.Error("with sanitation off no tile is covered")
	}
}

func TestIllnessRunsItsCourse(t *testing.T) {
	s := New(testMap(t, 64), 3)
	c := s.creatures[0]
	c.Health, c.Energy, c.Hydration = 1, 1, 1
	s.fallIll(c, Respiratory)
	c.Ill.Severity = 0.3
	for range 200 {
		s.course(c, 0.2, 0.2/SecondsPerYear)
	}
	if c.Ill != nil {
		t.Fatal("the illness should be over")
	}
	if c.immunity(Respiratory) <= 0 {
		t.Error("recovery should leave immunity")
	}
	if c.Health < 0.95 {
		t.Errorf("health should come back after the illness, got %.2f", c.Health)
	}
}

// TestDiseaseReport runs a world and prints what drives each disease (set
// DISEASE_REPORT=1; DISEASE_SEED and DISEASE_MINUTES choose the world).
func TestDiseaseReport(t *testing.T) {
	if os.Getenv("DISEASE_REPORT") == "" {
		t.Skip("set DISEASE_REPORT=1")
	}
	seed, minutes := uint64(5), 30
	if v, err := strconv.Atoi(os.Getenv("DISEASE_SEED")); err == nil {
		seed = uint64(v)
	}
	if v, err := strconv.Atoi(os.Getenv("DISEASE_MINUTES")); err == nil {
		minutes = v
	}
	m := world.Generate(world.GenOptions{Name: "Starter Island", Width: 128, Height: 128, Seed: 1337})
	s := New(m, seed)
	pct := func(v []float64, p float64) float64 {
		if len(v) == 0 {
			return 0
		}
		slices.Sort(v)
		return v[min(len(v)-1, int(p*float64(len(v))))]
	}
	t.Logf("%5s %4s | %6s %6s %6s | %5s %5s | %6s %6s %6s | %5s | %4s %4s %4s %4s",
		"year", "pop", "germ50", "germ90", "germMx", "nb50", "nb90", "mosq50", "mosq90", "z90", "worm", "D", "M", "R", "car")
	for minute := 0; minute < minutes; minute++ {
		for sec := 0; sec < 60; sec += 6 {
			s.Advance(6 * TicksPerSecond)
		}
		var germs, near, mosq, z []float64
		worms := 0.0
		var ill [numDiseases]int
		carriers := 0
		for _, c := range s.creatures {
			if c.WaterX != 0 || c.WaterY != 0 {
				if kind, tile := s.waterSource(&Creature{X: c.WaterX, Y: c.WaterY}); kind != drinkNone {
					germs = append(germs, s.germsDrunk(kind, tile))
				}
			}
			n := 0
			s.grid.near(c.X, c.Y, contactRange, func(o *Creature) {
				if o != c && math.Hypot(o.X-c.X, o.Y-c.Y) < contactRange {
					n++
				}
			})
			near = append(near, float64(n))
			k := s.epi.cellOf(c.X, c.Y)
			mosq = append(mosq, float64(s.epi.mosquito[k])*mosquitoScale/(float64(s.epi.blur[k])+otherHosts))
			z = append(z, float64(s.epi.infected[k]))
			worms += c.Worms
			if c.Ill != nil {
				ill[c.Ill.Disease]++
			} else if c.Carrier > 0 {
				carriers++
			}
		}
		n := float64(max(1, len(s.creatures)))
		t.Logf("%5d %4d | %6.3f %6.3f %6.3f | %5.1f %5.1f | %6.3f %6.3f %6.3f | %5.2f | %4d %4d %4d %4d",
			s.year(), len(s.creatures), pct(germs, 0.5), pct(germs, 0.9), pct(germs, 1), pct(near, 0.5), pct(near, 0.9),
			pct(mosq, 0.5), pct(mosq, 0.9), pct(z, 0.9), worms/n, ill[0], ill[1], ill[2], carriers)
	}
	d := s.deathsBy
	t.Logf("deaths: starvation %d thirst %d old %d killed %d animal %d childbirth %d | diarrhoea %d malaria %d respiratory %d",
		d.Starvation, d.Thirst, d.OldAge, d.Killed, d.Animal, d.Childbirth, d.Diarrhea, d.Malaria, d.Respiratory)
	met := s.stats.metrics(s)
	inc := func(p *float64) float64 {
		if p == nil {
			return -1
		}
		return *p
	}
	t.Logf("incidence/person-year: diarrhoea %.2f malaria %.2f respiratory %.2f; strains %d; latrines %d",
		inc(met.Incidence[0]), inc(met.Incidence[1]), inc(met.Incidence[2]), s.epi.strains,
		len(slices.DeleteFunc(slices.Clone(s.structures), func(st *Structure) bool { return !st.kind.Latrine })))
	_ = chem.ItemID("")
}

// TestDeathsByAge pools several worlds and prints deaths by age and cause
// (set DEATHS_REPORT=1; DEATHS_SEEDS and DEATHS_MINUTES choose the worlds).
func TestDeathsByAge(t *testing.T) {
	if os.Getenv("DEATHS_REPORT") == "" {
		t.Skip("set DEATHS_REPORT=1")
	}
	seeds, minutes := 8, 30
	if v, err := strconv.Atoi(os.Getenv("DEATHS_SEEDS")); err == nil {
		seeds = v
	}
	if v, err := strconv.Atoi(os.Getenv("DEATHS_MINUTES")); err == nil {
		minutes = v
	}
	bands := []float64{0, 1, 5, 10, 15, 20, 30, 40, 50, 60, 70, 200}
	causes := []string{"starvation", "thirst", "oldAge", "killed", "animal", "childbirth", "neonatal", "diarrhea", "malaria", "respiratory"}
	type tally map[string][]int
	results := make([]tally, seeds)
	m := world.Generate(world.GenOptions{Name: "Starter Island", Width: 128, Height: 128, Seed: 1337})
	done := make(chan int)
	for k := range seeds {
		go func() {
			tl := tally{}
			for _, c := range causes {
				tl[c] = make([]int, len(bands)-1)
			}
			tl["exposure"] = make([]int, len(bands)-1)
			s := New(m, uint64(k+1))
			s.onDeath = func(c *Creature, cause string) {
				age := s.ageYears(c)
				for b := range len(bands) - 1 {
					if age < bands[b+1] {
						tl[cause][b]++
						break
					}
				}
			}
			for range minutes * 60 {
				s.Advance(TicksPerSecond)
				for _, c := range s.creatures {
					age := s.ageYears(c)
					for b := range len(bands) - 1 {
						if age < bands[b+1] {
							tl["exposure"][b] += 1 // seconds; turned into years below
							break
						}
					}
				}
			}
			results[k] = tl
			done <- k
		}()
	}
	for range seeds {
		<-done
	}
	total := tally{}
	for _, c := range causes {
		total[c] = make([]int, len(bands)-1)
		for _, r := range results {
			for b, n := range r[c] {
				total[c][b] += n
			}
		}
	}
	head := "cause        "
	for b := range len(bands) - 1 {
		head += strconv.Itoa(int(bands[b])) + "-\t"
	}
	t.Log(head + "all")
	col := make([]int, len(bands)-1)
	for _, c := range causes {
		line, sum := c+"            "[:13-min(13, len(c))], 0
		for b, n := range total[c] {
			line += strconv.Itoa(n) + "\t"
			sum += n
			col[b] += n
		}
		t.Log(line + strconv.Itoa(sum))
	}
	line := "all          "
	for _, n := range col {
		line += strconv.Itoa(n) + "\t"
	}
	t.Log(line)
	// A pooled life table: death rates per band, then survival.
	var exp []float64
	for b := range len(bands) - 1 {
		e := 0
		for _, r := range results {
			e += r["exposure"][b]
		}
		exp = append(exp, float64(e)/SecondsPerYear)
	}
	l := 1.0
	line = "pooled: "
	for b := range len(bands) - 1 {
		if bands[b] >= 15 {
			break
		}
		n := bands[b+1] - bands[b]
		m := float64(col[b]) / math.Max(exp[b], 1e-9)
		q := math.Min(1, n*m/(1+n/2*m))
		l *= 1 - q
		line += "q" + strconv.Itoa(int(bands[b])) + "=" + strconv.FormatFloat(q, 'f', 3, 64) + " l" + strconv.Itoa(int(bands[b+1])) + "=" + strconv.FormatFloat(l, 'f', 3, 64) + "  "
	}
	t.Log(line)
	adult := 0.0
	deaths := 0
	for b := range len(bands) - 1 {
		if bands[b] >= 15 && bands[b] < 50 {
			adult += exp[b]
			deaths += col[b]
		}
	}
	t.Logf("adult (15–50) mortality %.2f%%/year", float64(deaths)/adult*100)
}

// TestChildDeaths looks at children who starve or die of thirst: were they
// with a carer, and how was the carer doing? (CHILD_REPORT=1)
func TestChildDeaths(t *testing.T) {
	if os.Getenv("CHILD_REPORT") == "" {
		t.Skip("set CHILD_REPORT=1")
	}
	m := world.Generate(world.GenOptions{Name: "Starter Island", Width: 128, Height: 128, Seed: 1337})
	type row struct {
		cause                      string
		age                        float64
		hasCarer, near             bool
		carerE, carerW, dist, ownE float64
		water                      bool
	}
	var rows []row
	for seed := uint64(1); seed <= 6; seed++ {
		s := New(m, seed)
		s.onDeath = func(c *Creature, cause string) {
			age := s.ageYears(c)
			if age < 2 || age >= 15 || (cause != "starvation" && cause != "thirst") {
				return
			}
			r := row{cause: cause, age: age, ownE: c.Energy, water: c.WaterX != 0 || c.WaterY != 0}
			var g *Creature
			if c.Mother != nil {
				g = s.living(c.Mother.ID)
			}
			if g == nil && c.Father != nil {
				g = s.living(c.Father.ID)
			}
			if g != nil {
				r.hasCarer = true
				r.dist = math.Hypot(g.X-c.X, g.Y-c.Y)
				r.near = r.dist < 4
				r.carerE, r.carerW = g.Energy, g.Hydration
			}
			rows = append(rows, r)
		}
		s.Advance(30 * 60 * TicksPerSecond)
	}
	var orphans, far, near int
	var nearE, nearW float64
	byAge := map[string][3]int{}
	for _, r := range rows {
		k := r.cause
		v := byAge[k]
		switch {
		case !r.hasCarer:
			orphans++
			v[0]++
		case !r.near:
			far++
			v[1]++
		default:
			near++
			v[2]++
			nearE += r.carerE
			nearW += r.carerW
		}
		byAge[k] = v
	}
	t.Logf("child (2–15) starvation/thirst deaths: %d — orphans %d, carer far (>4 tiles) %d, carer near %d", len(rows), orphans, far, near)
	if near > 0 {
		t.Logf("carers near a dying child: mean energy %.2f, hydration %.2f", nearE/float64(near), nearW/float64(near))
	}
	for k, v := range byAge {
		t.Logf("%s: orphan %d far %d near %d", k, v[0], v[1], v[2])
	}
	young := 0
	for _, r := range rows {
		if r.age < 10 {
			young++
		}
	}
	t.Logf("under 10: %d of %d", young, len(rows))
}

// TestAdultDeaths looks at grown-ups (15–50) who starve or die of thirst:
// was there food or water close by, family, anyone? (ADULT_REPORT=1)
func TestAdultDeaths(t *testing.T) {
	if os.Getenv("ADULT_REPORT") == "" {
		t.Skip("set ADULT_REPORT=1")
	}
	m := world.Generate(world.GenOptions{Name: "Starter Island", Width: 128, Height: 128, Seed: 1337})
	var n, foodNear, waterNear, waterKnown, kinNear, anyoneNear, sick, inHand, atHome int
	var thirstWet, thirstWetResting, thirstWetWants, starveFed, starveFedResting, starveFedWants int
	var sumE, sumW, sumAge float64
	byCause := map[string]int{}
	for seed := uint64(1); seed <= 6; seed++ {
		s := New(m, seed)
		s.onDeath = func(c *Creature, cause string) {
			age := s.ageYears(c)
			if age < 15 || age >= 50 || (cause != "starvation" && cause != "thirst") {
				return
			}
			n++
			byCause[cause]++
			sumAge += age
			x, y := int(math.Floor(c.X)), int(math.Floor(c.Y))
			if s.foodAround(x, y) > 0.05 {
				foodNear++
			}
			if s.canDrink(c) {
				waterNear++
				if cause == "thirst" {
					thirstWet++
					if c.resting {
						thirstWetResting++
					}
					if c.output[outDrink] > 0.5 {
						thirstWetWants++
					}
				}
			}
			if cause == "starvation" && s.foodAround(x, y) > 0.05 {
				starveFed++
				if c.resting {
					starveFedResting++
				}
				if c.output[outEat] > 0.5 {
					starveFedWants++
				}
			}
			if c.WaterX != 0 || c.WaterY != 0 {
				waterKnown++
			}
			if c.Ill != nil {
				sick++
			}
			if s.mealsInHand(c) > 0 {
				inHand++
			}
			if s.atHome(c) {
				atHome++
			}
			kin, other := false, false
			s.grid.near(c.X, c.Y, 4, func(o *Creature) {
				if o == c || o.Health <= 0 || math.Hypot(o.X-c.X, o.Y-c.Y) > 4 {
					return
				}
				other = true
				if s.kin(c, o) {
					kin = true
				}
			})
			if kin {
				kinNear++
			}
			if other {
				anyoneNear++
			}
			sumE += c.Energy
			sumW += c.Hydration
		}
		s.Advance(30 * 60 * TicksPerSecond)
	}
	f := float64(max(1, n))
	t.Logf("adult starvation/thirst deaths: %d %v, mean age %.0f", n, byCause, sumAge/f)
	t.Logf("food within reach %d, water within reach %d, knew where water was %d, ill %d, food in hand %d, at home %d",
		foodNear, waterNear, waterKnown, sick, inHand, atHome)
	t.Logf("family within 4 tiles %d, anyone within 4 tiles %d; energy %.2f hydration %.2f at death", kinNear, anyoneNear, sumE/f, sumW/f)
	t.Logf("died of thirst with water in reach %d (resting %d, wanting to drink %d); starved with food in reach %d (resting %d, wanting to eat %d)",
		thirstWet, thirstWetResting, thirstWetWants, starveFed, starveFedResting, starveFedWants)
}

// TestDenseWorldHealth restores a saved world (DENSE_MAP, DENSE_STATE) and
// prints its health every few years (DENSE_YEARS, default 60).
func TestDenseWorldHealth(t *testing.T) {
	mapPath, statePath := os.Getenv("DENSE_MAP"), os.Getenv("DENSE_STATE")
	if mapPath == "" || statePath == "" {
		t.Skip("set DENSE_MAP and DENSE_STATE")
	}
	years := 60
	if v, err := strconv.Atoi(os.Getenv("DENSE_YEARS")); err == nil {
		years = v
	}
	raw, err := os.ReadFile(mapPath)
	if err != nil {
		t.Fatal(err)
	}
	var m world.Map
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	s, err := Restore(&m, data)
	if err != nil {
		t.Fatal(err)
	}
	switch os.Getenv("DENSE_OFF") {
	case "disease":
		s.opts.NoDisease = true
	case "sanitation":
		s.opts.NoSanitation = true
	}
	var peak [numDiseases]int
	for y := 0; y < years; y += 5 {
		var maxIll [numDiseases]int
		for range 5 * 8 {
			s.Advance(TicksPerSecond)
			var ill [numDiseases]int
			for _, c := range s.creatures {
				if c.Ill != nil {
					ill[c.Ill.Disease]++
				}
			}
			for d := range numDiseases {
				maxIll[d] = max(maxIll[d], ill[d])
				peak[d] = max(peak[d], ill[d])
			}
		}
		met := s.stats.metrics(s)
		inc := func(p *float64) float64 {
			if p == nil {
				return -1
			}
			return *p
		}
		d := s.deathsBy
		latrines := 0
		for _, st := range s.structures {
			if st.kind.Latrine {
				latrines++
			}
		}
		t.Logf("latrines %d", latrines)
		t.Logf("year %d pop %d | most ill at once D %d M %d R %d | incidence D %.2f M %.2f R %.2f | deaths D %d M %d R %d starve %d thirst %d",
			s.year(), len(s.creatures), maxIll[0], maxIll[1], maxIll[2], inc(met.Incidence[0]), inc(met.Incidence[1]), inc(met.Incidence[2]),
			d.Diarrhea, d.Malaria, d.Respiratory, d.Starvation, d.Thirst)
	}
	_ = peak
}
