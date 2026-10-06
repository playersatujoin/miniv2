package sim

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/world"
)

// years runs c's nutrition (and growth) on for n years, ageing the world as it goes.
func years(s *Sim, c *Creature, n float64) {
	yr := float64(nutritionEvery) * dt / SecondsPerYear
	for t := 0.0; t < n; t += yr {
		s.tick += nutritionEvery
		s.nourishBody(c, yr)
	}
}

// child puts a child of age years with a blank brain at (x, y).
func child(s *Sim, name string, x, y int, age float64) *Creature {
	c := person(s, Female, name, x, y)
	c.BornTick = s.tick - int64(age*SecondsPerYear/dt)
	return c
}

func TestEveryEdibleItemHasANutrientProfile(t *testing.T) {
	cat := defaultCatalog()
	for id, it := range cat.items {
		if it.Food <= 0 {
			continue
		}
		n, ok := chem.NutrientsOf(id)
		if !ok || n.Protein <= 0 || n.Micro <= 0 {
			t.Errorf("edible %s has no usable profile: %+v", id, n)
		}
	}
	// Wild food everywhere people can stand has a profile too.
	s := New(testMap(t, 48), 1)
	for _, i := range s.terrain.walkable {
		if n := s.eco.ForageNutrients(int(i)); n.Protein <= 0 || n.Micro <= 0 || math.IsNaN(n.Protein+n.Micro) {
			t.Fatalf("tile %d: %+v", i, n)
		}
	}
}

func TestMealsMixIntoTheReserveByEnergy(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Male, "Adi", x, y)
	c.Energy = 0.5
	if p, m := c.Nutrition.diet(); p != 1 || m != 1 {
		t.Fatalf("an unknown diet counts as adequate: %.2f %.2f", p, m)
	}
	s.eatFood(c, "padi", 0.35)
	rice, _ := chem.NutrientsOf("padi")
	w := 0.35 / 0.85
	if want := 1 + (rice.Protein-1)*w; math.Abs(c.Nutrition.DietProtein-want) > 1e-9 {
		t.Errorf("rice into a half-full body: protein %.3f, want %.3f", c.Nutrition.DietProtein, want)
	}
	// An empty stomach takes on almost all of what it eats.
	c.Energy = 0
	s.eatFood(c, "ikan", 0.4)
	fish, _ := chem.NutrientsOf("ikan")
	if c.Nutrition.DietProtein < 0.85*fish.Protein {
		t.Errorf("fish into an empty body: %.2f", c.Nutrition.DietProtein)
	}
	// Food without a profile counts as mixed wild food; nothing eaten, nothing changes.
	before := c.Nutrition
	s.eatFood(c, "kayu", 0)
	if c.Nutrition != before {
		t.Error("eating nothing changed the diet")
	}
}

func TestStoresDrainOnAPoorDietAndRecover(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Male, "Adi", x, y)
	c.Energy = 0.8
	c.Nutrition.DietProtein, c.Nutrition.DietMicro = 0.02, 0.1 // sago only
	years(s, c, 1.0/12)
	month := c.Nutrition.deficit()
	years(s, c, 2.0/12)
	quarter := c.Nutrition.deficit()
	years(s, c, 1.75)
	dp, dm := c.Nutrition.deficits()
	if month <= 0.1 || month >= badFrom || quarter < badFrom {
		t.Errorf("on sago alone an adult becomes deficient over months: %.2f after one, %.2f after three", month, quarter)
	}
	if dp < 0.99 || dm < 0.99 {
		t.Errorf("two years of sago: protein %.2f, micro %.2f, want severe", dp, dm)
	}
	// Fish and wild greens bring it back over months, and build a small reserve.
	c.Nutrition.DietProtein, c.Nutrition.DietMicro = 2, 1.5
	years(s, c, 0.25)
	if d := c.Nutrition.deficit(); d < 0.05 || d > 0.95 {
		t.Errorf("recovery should take months: %.2f after three", d)
	}
	years(s, c, 3)
	if c.Nutrition.deficit() != 0 || c.Nutrition.ProteinLack > -0.2 || c.Nutrition.ProteinLack < -reserveMax {
		t.Errorf("a rich diet should leave a reserve: %+v", c.Nutrition)
	}
	// A child's stores turn over faster than an adult's.
	k := child(s, "Kecil", x, y, 6)
	a := person(s, Female, "Ibu", x, y)
	for _, p := range []*Creature{k, a} {
		p.Energy = 0.8
		p.Nutrition.DietProtein, p.Nutrition.DietMicro = 0.4, 0.5
	}
	years(s, k, 0.2)
	years(s, a, 0.2)
	if k.Nutrition.MicroLack <= a.Nutrition.MicroLack {
		t.Errorf("child %.2f should fall behind faster than adult %.2f", k.Nutrition.MicroLack, a.Nutrition.MicroLack)
	}
}

func TestNeedsAreHigherInPregnancyNursingAndChildhood(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	w := person(s, Female, "Sari", x, y)
	w.BornTick = s.tick - int64(25*SecondsPerYear/dt)
	p0, m0 := s.nutrientNeed(w)
	w.Pregnancy = &Pregnancy{Remaining: gestation}
	p1, m1 := s.nutrientNeed(w)
	if p0 != 1 || m0 != 1 || p1 <= p0 || m1 <= m0 {
		t.Errorf("pregnancy should raise needs: %.2f/%.2f → %.2f/%.2f", p0, m0, p1, m1)
	}
	w.Pregnancy = nil
	baby := child(s, "Bayi", x, y, 0.3)
	w.NursingID = baby.ID
	if p, m := s.nutrientNeed(w); p <= 1 || m <= 1 {
		t.Errorf("nursing should raise needs: %.2f/%.2f", p, m)
	}
	toddler := child(s, "Balita", x, y, 3)
	if p, m := s.nutrientNeed(toddler); m <= 1.2 || p >= 1 {
		t.Errorf("a toddler needs denser micronutrients but less protein per unit of energy: %.2f/%.2f", p, m)
	}
}

func TestDeficiencyWorsensInfectionWithinBounds(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Male, "Adi", x, y)
	if s.malnutritionSeverity(c) != 1 || s.malnutritionLength(c, Diarrhea) != 1 {
		t.Fatal("a well-fed body should be unaffected")
	}
	c.Nutrition.MicroLack = 1
	sev, long := s.malnutritionSeverity(c), s.malnutritionLength(c, Diarrhea)
	if sev <= 1.2 || sev > 2 {
		t.Errorf("severe deficiency: infections %.2f× as severe, want 1.2–2", sev)
	}
	if long <= 1 || long > 1.5 || s.malnutritionLength(c, Respiratory) != 1 {
		t.Errorf("diarrhoea %.2f× as long; pneumonia unaffected", long)
	}
	// The same bout, drawn from the same random state, goes worse.
	c.Genome.Traits.Immunity = 1
	state, _ := s.src.MarshalBinary()
	worse := s.severityOf(c, Diarrhea)
	s.src.UnmarshalBinary(state)
	c.Nutrition.MicroLack = 0
	well := s.severityOf(c, Diarrhea)
	if r := worse / well; math.Abs(r-sev) > 1e-9 {
		t.Errorf("severity ratio %.3f, want %.3f", r, sev)
	}
	// Stunted under-fives are frailer still; adults are not.
	k := child(s, "Kecil", x, y, 2)
	k.Nutrition.Stunt = 0.1
	if m := s.malnutritionSeverity(k); m <= 1 || m > 1.2 {
		t.Errorf("stunted toddler: %.2f", m)
	}
	c.Nutrition.Stunt = 0.1
	if s.malnutritionSeverity(c) != 1 {
		t.Error("an adult's lost height doesn't make infections worse")
	}
	s.fallIll(k, Diarrhea)
	if k.Ill.Length <= diarrheaLength*SecondsPerYear*0.999 {
		t.Error("the bout should have its normal length or more")
	}
}

func TestDeficiencyLowersFertilityAndRaisesBirthRisks(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	f := person(s, Female, "Sari", x, y)
	f.BornTick = s.tick - int64(25*SecondsPerYear/dt)
	f.Energy = 0.9
	base, cb, nn := s.conceptionChance(f), s.childbirthRisk(f, 1), s.neonatalDeathRisk(f, 1)
	f.Nutrition.ProteinLack, f.Nutrition.MicroLack = 1, 1
	if r := s.conceptionChance(f) / base; r >= 1 || r < 0.8 {
		t.Errorf("a severely deficient woman conceives %.2f× as often, want 0.8–1", r)
	}
	if r := s.childbirthRisk(f, 1) / cb; r <= 1.2 || r > 2 {
		t.Errorf("childbirth risk %.2f×, want 1.2–2", r)
	}
	if r := s.neonatalDeathRisk(f, 1) / nn; r <= 1.2 || r > 2 {
		t.Errorf("neonatal risk %.2f×, want 1.2–2", r)
	}
	// A short mother: more obstructed labour; a stunted girl matures later.
	f.Nutrition = Nutrition{Stunt: 0.1}
	if r := s.childbirthRisk(f, 1) / cb; r <= 1.1 || r > 1.6 {
		t.Errorf("short mother: childbirth risk %.2f×", r)
	}
	girl := child(s, "Remaja", x, y, 16.5)
	girl.Energy = 0.9
	young := s.conceptionChance(girl)
	girl.Nutrition.Stunt = 0.1
	if r := s.conceptionChance(girl) / young; r >= 0.9 || r < 0.5 {
		t.Errorf("stunted girl conceives %.2f× as often, want 0.5–0.9", r)
	}
	// Wounds heal slower.
	if h := s.nutritionHealing(f); h != 1 {
		t.Errorf("stunting alone shouldn't slow healing: %.2f", h)
	}
	f.Nutrition.ProteinLack = 1
	if h := s.nutritionHealing(f); h >= 1 || h < 0.4 {
		t.Errorf("protein-deficient healing %.2f, want 0.4–1", h)
	}
}

func TestChildrenStuntOnlyWhenShortOfFood(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	poor := child(s, "Kurang", x, y, 0.5)
	well := child(s, "Cukup", x, y, 0.5)
	hungry := child(s, "Lapar", x, y, 0.5)
	adult := person(s, Male, "Adi", x, y)
	for _, c := range []*Creature{poor, well, hungry, adult} {
		c.Energy = 0.8
	}
	hungry.Energy = 0.3
	well.Nutrition.DietProtein, well.Nutrition.DietMicro = 1.5, 1.6
	poor.Nutrition.DietProtein, poor.Nutrition.DietMicro = 0.4, 0.4 // taro and little else
	adult.Nutrition.DietProtein, adult.Nutrition.DietMicro = 0.4, 0.4
	for _, c := range []*Creature{poor, well, hungry, adult} {
		save := s.tick
		years(s, c, 2)
		s.tick = save
	}
	if well.Nutrition.Stunt != 0 {
		t.Errorf("a well-fed child shouldn't stunt: %.3f", well.Nutrition.Stunt)
	}
	if poor.Nutrition.Stunt < stuntedFrom || poor.Nutrition.Stunt > 0.15 {
		t.Errorf("a deficient toddler should be stunted by two and a half, moderately: %.3f", poor.Nutrition.Stunt)
	}
	if hungry.Nutrition.Stunt < 0.03 {
		t.Errorf("a hungry toddler should fall behind: %.3f", hungry.Nutrition.Stunt)
	}
	if adult.Nutrition.Stunt != 0 {
		t.Error("grown-ups don't grow")
	}
	// Well fed again, a child makes up only part of what it lost.
	lost := poor.Nutrition.Stunt
	poor.Nutrition.DietProtein, poor.Nutrition.DietMicro = 1.5, 1.5
	years(s, poor, 10)
	if got := poor.Nutrition.Stunt; got >= lost || got < 0.4*lost {
		t.Errorf("catch-up from %.3f to %.3f, want partial", lost, got)
	}
}

func TestMothersPassOnTheirNutrition(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	m := person(s, Female, "Ibu", x, y)
	m.Energy = 0.9
	m.Nutrition = Nutrition{ProteinLack: 0.5, MicroLack: 0.6, Stunt: 0.1}
	baby := child(s, "Bayi", x, y, 0)
	baby.Mother = &Ref{m.ID, m.Name}
	s.bornNourished(baby, m)
	if baby.Nutrition.Stunt <= 0.02 || baby.Nutrition.Stunt > 0.06 || baby.Nutrition.MicroLack <= 0 {
		t.Errorf("a deficient, short mother's baby is born smaller with fewer stores: %+v", baby.Nutrition)
	}
	// Her milk carries fewer micronutrients than a well-fed mother's.
	baby.Energy = 0.3
	s.suckle(baby, m, 0.3)
	poorMilk := baby.Nutrition.DietMicro
	baby.Nutrition, baby.Energy = Nutrition{}, 0.3
	m.Nutrition = Nutrition{}
	s.suckle(baby, m, 0.3)
	if poorMilk >= baby.Nutrition.DietMicro {
		t.Errorf("milk micronutrients %.2f (deficient mother) vs %.2f", poorMilk, baby.Nutrition.DietMicro)
	}
	// A well-fed mother's milk meets her baby's needs.
	years(s, baby, 0.5)
	if baby.Nutrition.deficit() != 0 {
		t.Errorf("breastfed baby of a well-fed mother: %+v", baby.Nutrition)
	}
}

func TestMalnourishedSense(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Male, "Adi", x, y)
	s.sense(c)
	if v := c.input[inMalnourished]; v != 0 {
		t.Errorf("well fed: %.2f", v)
	}
	for _, lack := range []float64{0.1, 0.3, 0.6, 2} {
		c.Nutrition.ProteinLack = lack
		s.sense(c)
		if v := c.input[inMalnourished]; v < 0 || v > 1 || v != c.Nutrition.deficit() {
			t.Errorf("lack %.1f: sense %.2f", lack, v)
		}
	}
	if c.input[inMalnourished] != 1 {
		t.Error("a severe deficiency should be felt fully")
	}
	s.opts.NoNutrition = true
	s.sense(c)
	if c.input[inMalnourished] != 0 {
		t.Error("without nutrition the sense stays 0")
	}
}

func TestNutritionLabels(t *testing.T) {
	cases := []struct {
		deficit, energy float64
		want            string
	}{{0, 0.8, "baik"}, {0.3, 0.8, "kurang"}, {0.7, 0.8, "buruk"}, {0, 0.25, "kurang"}, {0, 0.1, "buruk"}}
	for _, k := range cases {
		if got := nutritionLabel(k.deficit, k.energy); got != k.want {
			t.Errorf("deficit %.1f energy %.2f: %s, want %s", k.deficit, k.energy, got, k.want)
		}
	}
	s, _, x, y := fakeWorld(t)
	c := child(s, "Kecil", x, y, 3)
	c.Energy = 0.8
	c.Nutrition = Nutrition{DietProtein: 0.5, DietMicro: 0.6, MicroLack: 0.3, Stunt: 0.09}
	v := s.bodyView(c).Nutrition
	if v == nil || v.Label != "kurang" || !v.Stunted || v.Need != "balita" || math.Abs(v.Micro-0.7) > 1e-9 || v.DietMicro >= 0.5 {
		t.Errorf("view %+v", v)
	}
	info := s.nutritionInfo()
	if info.Under5 != 1 || info.Stunted5 != 1 || info.Deficient != 1 {
		t.Errorf("info %+v", info)
	}
	s.opts.NoNutrition = true
	if s.bodyView(c).Nutrition != nil || s.nutritionInfo().People != 0 {
		t.Error("nothing to show without nutrition")
	}
}

func TestNutritionSurvivesSaveAndRestore(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	a := person(s, Female, "Sari", x, y)
	k := child(s, "Kecil", x, y, 3)
	a.Nutrition = Nutrition{DietProtein: 0.513, DietMicro: 0.4471, ProteinLack: 0.21, MicroLack: 0.637, Stunt: 0.0412}
	k.Nutrition = Nutrition{DietProtein: 1.7, DietMicro: 0.3, ProteinLack: -0.1, MicroLack: 0.7, Stunt: 0.08}
	for _, c := range []*Creature{a, k} {
		c.Genome.Traits.Immunity = 1 // as a restore would make it
		c.Genome.Traits.upgradeTemperament()
	}
	s.Advance(43)
	data, err := s.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	r, err := restoreWith(testMap(t, 48), data, fakeCatalog())
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range s.creatures {
		if r.byID[c.ID].Nutrition != c.Nutrition {
			t.Fatalf("%s: %+v restored as %+v", c.Name, c.Nutrition, r.byID[c.ID].Nutrition)
		}
	}
	s.Advance(200)
	r.Advance(200)
	j1, _ := json.Marshal(s.creatures)
	j2, _ := json.Marshal(r.creatures)
	if !bytes.Equal(j1, j2) {
		t.Fatal("the restored world diverged")
	}
}

// Switched off, nutrition leaves no trace: nobody's state changes, every
// effect is neutral, the sense stays 0 and saves look as they did before.
func TestNoNutritionIsTheCaloriesOnlyWorld(t *testing.T) {
	if testing.Short() {
		t.Skip("runs a simulated minute")
	}
	m := world.Generate(world.GenOptions{Name: "Starter Island", Width: 96, Height: 96, Seed: 1337})
	m.ID = "0000000000000000"
	s := NewWithOptions(m, 3, Options{NoNutrition: true})
	s.Advance(60 * TicksPerSecond)
	if len(s.creatures) < 2 {
		t.Fatalf("only %d people", len(s.creatures))
	}
	for _, c := range s.creatures {
		if c.Nutrition != (Nutrition{}) || c.input[inMalnourished] != 0 {
			t.Fatalf("%s: %+v, sense %.2f", c.Name, c.Nutrition, c.input[inMalnourished])
		}
		c.Nutrition = Nutrition{ProteinLack: 1, MicroLack: 1, Stunt: 0.2}
		if s.malnutritionSeverity(c) != 1 || s.malnutritionLength(c, Diarrhea) != 1 || s.nutritionFertility(c) != 1 ||
			s.nutritionBirthRisk(c) != 1 || s.nutritionNeonatalRisk(c) != 1 || s.nutritionHealing(c) != 1 {
			t.Fatal("an effect is not neutral without nutrition")
		}
		c.Nutrition = Nutrition{}
	}
	data, _ := json.Marshal(s.creatures)
	if strings.Contains(string(data), `"nutrition"`) {
		t.Error("creatures saved with a nutrition field")
	}
	// The same world with nutrition on does keep track.
	on := New(m, 3)
	on.Advance(60 * TicksPerSecond)
	tracked := 0
	for _, c := range on.creatures {
		if c.Nutrition != (Nutrition{}) {
			tracked++
		}
	}
	if tracked == 0 {
		t.Error("with nutrition on, nobody's nutrition was tracked")
	}
}
