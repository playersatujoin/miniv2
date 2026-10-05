package sim

import (
	"encoding/json"
	"math"
	"math/rand/v2"
	"strings"
	"testing"
)

func near(t *testing.T, what string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %.4f, want %.4f ± %.4f", what, got, want, tol)
	}
}

// With the same death rate at every age, survival is exponential: e0 = 1/μ
// and l(15) = e^(−15μ).
func TestLifeTableConstantHazard(t *testing.T) {
	const mu = 0.03
	var exposure, deaths [numBands]float64
	for b := range numBands {
		exposure[b] = 10000
		deaths[b] = mu * exposure[b]
	}
	lt, ok := lifeTable(exposure, deaths)
	if !ok {
		t.Fatal("no life table")
	}
	near(t, "e0", lt.e0, 1/mu, 0.03/mu)
	near(t, "l15", lt.l15, math.Exp(-15*mu), 0.01)
	near(t, "q0", lt.q0, 1-math.Exp(-mu), 0.002)
	near(t, "e15", lt.e15, 1/mu, 0.03/mu)
}

func TestLifeTableEndsWhereNobodyWasSeen(t *testing.T) {
	var exposure, deaths [numBands]float64
	for b := range bandOf(20) + 1 { // nobody older than 24 in the window
		exposure[b] = 1000
		deaths[b] = 10
	}
	lt, ok := lifeTable(exposure, deaths)
	if !ok || lt.e0 <= 15 || lt.e0 > 27 {
		t.Fatalf("e0 = %.1f (ok=%v), want it cut off in the late twenties", lt.e0, ok)
	}
	if _, ok := lifeTable([numBands]float64{}, deaths); ok {
		t.Fatal("a table without infants should not be reported")
	}
}

func TestTFRAndGini(t *testing.T) {
	var births, female [numBands]float64
	for b := band15; b <= band45; b++ {
		female[b] = 100
		births[b] = 30 // 0.3 children per woman-year
	}
	v, ok := tfr(births, female)
	if !ok {
		t.Fatal("tfr not reported")
	}
	near(t, "tfr", v, 0.3*5*7, 1e-9)

	if g := gini([]float64{1, 1, 1, 1}); g != 0 {
		t.Errorf("equal wealth: gini %.3f", g)
	}
	near(t, "gini one has all", gini([]float64{0, 0, 0, 10}), 0.75, 1e-9)
	near(t, "gini", gini([]float64{1, 2, 3, 4}), 0.25, 1e-9)
}

func TestDemographyRecordsBirthsDeathsAndExposure(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	mother := person(s, Female, "Ibu", x, y)
	person(s, Male, "Ayah", x, y)
	// The mother is 15 at spawn; give her a first child at 20 and a second at 23.
	s.tick += int64(5 * SecondsPerYear * TicksPerSecond)
	s.stats.delivery(s, mother, 1)
	mother.Children = 1
	s.tick += int64(3 * SecondsPerYear * TicksPerSecond)
	s.stats.delivery(s, mother, 2)

	y0 := s.stats.current(s)
	if y0.Births[bandOf(23)] != 2 || y0.IntervalCount != 1 {
		t.Fatalf("births %v intervals %d", y0.Births, y0.IntervalCount)
	}
	near(t, "interval", y0.Intervals, 3, 1e-9)
	first := s.stats.Years[0]
	if first.FirstBirths != 1 {
		t.Fatalf("first births %d", first.FirstBirths)
	}
	near(t, "first birth age", first.FirstBirthAges, 20, 1e-9)

	s.stats.expose(s)
	if got := y0.Exposure[Female][bandOf(23)] + y0.Exposure[Male][bandOf(23)]; math.Abs(got-2/SecondsPerYear) > 1e-9 {
		t.Fatalf("one second of two lives should be 2/8 person-years, got %v", got)
	}

	age := s.ageYears(mother)
	s.die(mother, "killed")
	if y0.Causes.Killed != 1 || y0.Deaths[Female][bandOf(age)] != 1 || y0.AdultDeathAges[int((age-15)/2)] != 1 {
		t.Fatalf("death not recorded: %+v", y0.Causes)
	}
	if text := s.events[len(s.events)-1].Text; !strings.HasSuffix(text, "pada usia 23 tahun") {
		t.Fatalf("death event should give the age in years: %q", text)
	}
}

func TestDemographyOfAGrowingWorld(t *testing.T) {
	s := New(testMap(t, 64), 3)
	for range int(60 * SecondsPerYear * TicksPerSecond) { // 60 years
		s.step()
	}
	d := s.Demography()
	if d.SecondsPerYear != 8 || d.Year != 61 || len(d.Pyramid) != 17 || d.Pyramid[16].Label != "80+" || d.Pyramid[16].To != nil {
		t.Fatalf("bad demography header: spy=%v year=%d bands=%d", d.SecondsPerYear, d.Year, len(d.Pyramid))
	}
	alive := 0
	for _, b := range d.Pyramid {
		alive += b.Female + b.Male
	}
	if alive != len(s.creatures) {
		t.Fatalf("pyramid counts %d, population %d", alive, len(s.creatures))
	}
	if d.Current.WindowYears != statsWindowYears+1 || d.Current.PersonYears <= 0 {
		t.Fatalf("window %d years, %v person-years", d.Current.WindowYears, d.Current.PersonYears)
	}
	if len(d.History) != 6 {
		t.Fatalf("history points %d, want one per decade", len(d.History))
	}
	for _, key := range []string{"lifeExpectancy", "survivalTo15", "tfr", "gini"} {
		if _, ok := d.Reference[key]; !ok {
			t.Errorf("no reference for %s", key)
		}
	}
}

func TestStatsAndOptionsSurviveSaving(t *testing.T) {
	m := testMap(t, 48)
	a := NewWithOptions(m, 11, Options{NoCrime: true})
	for range int(12 * SecondsPerYear * TicksPerSecond) {
		a.step()
	}
	data, err := a.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	b, err := Restore(m, data)
	if err != nil {
		t.Fatal(err)
	}
	if !b.opts.NoCrime {
		t.Fatal("options lost")
	}
	ja, _ := json.Marshal(a.Demography())
	jb, _ := json.Marshal(b.Demography())
	if string(ja) != string(jb) {
		t.Fatal("demography differs after restore")
	}
}

func TestCrimeCanBeSwitchedOff(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	thief := person(s, Male, "Maling", x, y)
	victim := person(s, Female, "Korban", x, y)
	victim.Inventory = Stock{"besi": 3}
	s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures)
	thief.output[outSteal] = 1
	thief.output[outAttack] = 1

	s.opts.NoCrime = true
	s.work(thief)
	if s.crimes != 0 || victim.Health < 1 || victim.Inventory["besi"] != 3 {
		t.Fatal("crime happened although it is switched off")
	}
	s.opts.NoCrime = false
	s.work(thief)
	if s.crimes != 1 {
		t.Fatal("crime should happen when allowed")
	}
}

func TestFoundersWithoutInstincts(t *testing.T) {
	plain := randomGenomeWith(rand.New(rand.NewPCG(1, 2)), 0, false)
	for _, w := range plain.WOut {
		if w != 0 {
			t.Fatal("a noiseless genome without instincts should be blank")
		}
	}
	if randomGenomeWith(rand.New(rand.NewPCG(1, 2)), 0, true).WOut[outEat] == 0 {
		t.Fatal("instincts should wire the eating reflex")
	}

	m := testMap(t, 48)
	with, without := New(m, 5), NewWithOptions(m, 5, Options{NoInstincts: true})
	reflex := func(s *Sim) float64 { return math.Abs(float64(s.creatures[0].Genome.WOut[outEat])) }
	if reflex(with) < 2 || reflex(without) > 1 {
		t.Fatalf("eating reflex %.2f with instincts, %.2f without", reflex(with), reflex(without))
	}
}

func TestBiologyTimeScale(t *testing.T) {
	g := randomGenome(rand.New(rand.NewPCG(9, 9)))
	if g.Traits.Lifespan < 60*SecondsPerYear || g.Traits.Lifespan > 80*SecondsPerYear {
		t.Fatalf("newborn lifespan %.0f s, want 60–80 years", g.Traits.Lifespan)
	}
	if adultAge != 120 || gestation != 6 {
		t.Fatalf("adult at %v s, gestation %v s", adultAge, gestation)
	}
	s := New(testMap(t, 32), 1)
	if age := s.ageYears(s.creatures[0]); math.Abs(age-15) > 0.01 {
		t.Fatalf("Adam starts at %.2f years, want 15", age)
	}
}
