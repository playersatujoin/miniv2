package sim

import (
	"math"

	"miniv2/backend/internal/chem"
)

// Nutrition beyond calories (Fase 3d). Energy says how much someone eats;
// this says what: protein and micronutrients (vitamins and minerals, one
// pool; see chem/nutrition.go for the food values and their basis).
//
//   - The body lives on its energy reserve, and that reserve is made of what
//     was eaten: each meal mixes its protein and micronutrient density into
//     the reserve's in proportion to the energy it adds. A unit of food here
//     feeds a person for a year or two (energy is compressed far less than
//     the calendar), so its nutrients reach the body as that energy is
//     spent, as if the unit were eaten in daily portions.
//   - The body's protein and micronutrient status follow what that diet
//     supplies against what this body needs, as first-order pools do (intake
//     in, a fixed share turned over): within a couple of months in a child,
//     longer in an adult whose stores are larger, and they recover likewise.
//     A rich diet builds up a small reserve.
//   - Needs per unit of energy: young children need about two thirds of an
//     adult's protein density (their energy needs are high; WHO/FAO/UNU 2007,
//     Table 47) but much denser micronutrients (iron and zinc above all;
//     Dewey & Brown 2003, Food Nutr Bull 24:5); pregnancy and nursing raise
//     both (IOM Dietary Reference Intakes: iron 27 against 18 mg a day,
//     iodine 220–290 against 150 µg, protein +25 g, for some 15% more
//     energy); the old need a little more protein. Breast milk meets a
//     baby's needs, but its vitamin A and iodine depend on the mother's
//     stores (Dror & Allen 2018, Adv Nutr 9:278S).
//   - A child short of protein, micronutrients or food, or ill, grows less;
//     the lost height (stunting) is made up only in part later (Victora et
//     al. 2010, Pediatrics 125:e473: most faltering is before two; Prentice
//     et al. 2013, Am J Clin Nutr 97:911: partial catch-up in adolescence;
//     Checkley et al. 2008: repeated diarrhoea stunts). Babies of deficient
//     or short mothers are born small (Black et al. 2013, Lancet 382:427).
//   - What it does, kept modest: infections go worse, and diarrhoea lasts longer
//     (underweight and deficient children die of diarrhoea, pneumonia and
//     malaria two to several times as often, Black et al. 2008, Lancet
//     371:243; Caulfield et al. 2004; zinc shortens diarrhoea, Lazzerini &
//     Wanzira 2016), a deficient woman conceives a little less often and a
//     stunted girl later (Bongaarts 1980, Science 208:564: the effect is
//     modest), childbirth is riskier for an anaemic or short mother (Daru et
//     al. 2018, Lancet Glob Health 6:e548) and her baby more often dies
//     (Kozuki et al. 2015; Black et al. 2013), and wounds heal slower
//     (Stechmiller 2010, Nutr Clin Pract 25:61). Malnutrition is not a cause
//     of death of its own: as in Black et al. it kills through infection,
//     starvation and childbirth.
//
// Nobody knows any of this. The brain feels "kurang gizi" (inMalnourished)
// and may learn to seek other food; nothing makes it.

// Nutrition is a body's protein and micronutrient state. The zero value is a
// well-nourished body on an adequate diet (an old save, or nutrition switched
// off), so it costs nothing in saves until it matters.
type Nutrition struct {
	// DietProtein and DietMicro are the densities of what the body lives on
	// now (its energy reserve), relative to an adult's need; 0 = not yet
	// known, counted as just what this body needs.
	DietProtein float64 `json:"dietProtein,omitempty"`
	DietMicro   float64 `json:"dietMicro,omitempty"`
	// ProteinLack and MicroLack are how far the body's status has fallen
	// below adequate (0 adequate … 1 empty; negative: a reserve).
	ProteinLack float64 `json:"proteinLack,omitempty"`
	MicroLack   float64 `json:"microLack,omitempty"`
	// Stunt is the height a child has lost to want and illness, as a share
	// of what a well-fed child of that age would have; it stays for life.
	Stunt float64 `json:"stunt,omitempty"`
}

const (
	nutritionEvery = 10 // ticks between updates of a body (staggered by id)

	// Turnover (years) of the protein and micronutrient pools: how fast
	// status follows the diet, in a child (under 10) and an adult.
	proteinTurnoverChild = 0.15
	proteinTurnoverAdult = 0.3
	microTurnoverChild   = 0.3
	microTurnoverAdult   = 0.75
	// reserveMax: the most a rich diet stores beyond adequate.
	reserveMax = 0.25
	// deficitRange: status this far below adequate is a severe deficiency.
	deficitRange = 0.6
	// lowReserve: below this energy the reserve that supplies the nutrients
	// is running out with it.
	lowReserve = 0.2

	// Labels: deficits from poorFrom are "kurang", from badFrom "buruk".
	poorFrom = 0.25
	badFrom  = 0.6
	// thinBelow, wastedBelow: energy below which a body is thin or wasted.
	thinBelow   = 0.3
	wastedBelow = 0.15

	// Breast milk, relative to an adult's need per unit of energy: enough
	// protein and micronutrients for a baby when the mother is well fed.
	milkProtein = 0.7
	milkMicro   = 1.5

	// Growth: share of a child's growth lost at the worst, how readily lost
	// height is made up when well fed (relative to the growth rate), the most
	// that can be lost, and the height lost below which a child counts as
	// stunted (about −2 SD of height-for-age, 7–8% of the median height;
	// WHO Child Growth Standards 2006).
	stuntLoss     = 0.35
	catchUp       = 0.5
	stuntMax      = 0.25
	stuntedFrom   = 0.075
	stuntAtBirth  = 0.03 // height a baby of a severely deficient mother is born without …
	stuntInherits = 0.2  // … plus this share of a short mother's lost height
	childHungerAt = 0.45 // a child's energy below which growth falters …
	childHungerTo = 0.3  // … fully this much lower

	// Effects at a full deficit (each scaled by the deficit).
	sickWorse      = 0.5  // infections this much more severe …
	stuntSickWorse = 1.0  // … plus this times the height lost, under five
	sickLonger     = 0.25 // and diarrhoea this much longer
	fewerConceived = 0.15 // a month's chance of conceiving lower
	lateMaturing   = 3.0  // a stunted adolescent's fecundity, lower by this times her lost height
	anaemicBirth   = 0.6  // risk of dying in childbirth higher with a micronutrient deficit …
	shortBirth     = 3.0  // … and by this times a short mother's lost height
	smallBaby      = 0.6  // a newborn's risk higher with the mother's deficit …
	slowHealing    = 0.5  // wounds heal this much slower
)

func (s *Sim) noNutrition() bool { return s.opts.NoNutrition }

// known reports whether the body's diet has been seen yet.
func (n *Nutrition) known() bool { return n.DietProtein != 0 || n.DietMicro != 0 }

// diet is what the body is living on (an adult's adequate diet when not yet known).
func (n *Nutrition) diet() (protein, micro float64) {
	if !n.known() {
		return 1, 1
	}
	return n.DietProtein, n.DietMicro
}

// deficits are the protein and micronutrient deficits, 0 well … 1 severe.
func (n *Nutrition) deficits() (protein, micro float64) {
	return clamp(n.ProteinLack/deficitRange, 0, 1), clamp(n.MicroLack/deficitRange, 0, 1)
}

// deficit is the worse of the two.
func (n *Nutrition) deficit() float64 {
	p, m := n.deficits()
	return math.Max(p, m)
}

// mix takes energy (the amount the body gains) of food with profile f into
// the reserve, which held before.
func (n *Nutrition) mix(f chem.Nutrients, before, energy float64) {
	if energy <= 0 {
		return
	}
	p, m := n.diet()
	w := energy / (math.Max(before, 0.05) + energy)
	n.DietProtein = p + (f.Protein-p)*w
	n.DietMicro = m + (f.Micro-m)*w
}

// --- eating ----------------------------------------------------------------------

// meal is the start of c's next meal: a body whose diet isn't known yet (a
// world saved before nutrition) has been living on just what it needs.
func (s *Sim) meal(c *Creature) *Nutrition {
	n := &c.Nutrition
	if !n.known() {
		n.DietProtein, n.DietMicro = s.nutrientNeed(c)
	}
	return n
}

// eatWild: c takes energy from the wild food around tile (x, y).
func (s *Sim) eatWild(c *Creature, x, y int, energy float64) {
	if s.noNutrition() {
		return
	}
	i, ok := s.terrain.index(x, y)
	if !ok {
		i = -1
	}
	s.meal(c).mix(s.eco.ForageNutrients(i), c.Energy, energy)
}

// eatFood: c takes energy from a carried unit of id.
func (s *Sim) eatFood(c *Creature, id chem.ItemID, energy float64) {
	if s.noNutrition() {
		return
	}
	f, ok := chem.NutrientsOf(id)
	if !ok {
		f = chem.MixedWild
	}
	s.meal(c).mix(f, c.Energy, energy)
}

// suckle: baby c takes energy of its mother m's milk.
func (s *Sim) suckle(c, m *Creature, energy float64) {
	if s.noNutrition() {
		return
	}
	s.meal(c).mix(milk(m), c.Energy, energy)
}

// milk is what mother m's breast milk carries: its protein whatever her
// state, its vitamin A and iodine less when her own stores are low.
func milk(m *Creature) chem.Nutrients {
	stores := clamp(1-m.Nutrition.MicroLack, 0, 1)
	return chem.Nutrients{Protein: milkProtein, Micro: milkMicro * (0.6 + 0.4*stores)}
}

// bornNourished gives a newborn what its mother's body could give it: it
// lives on her milk, and a deficient mother's baby is born with smaller
// stores, and smaller.
func (s *Sim) bornNourished(c, mother *Creature) {
	if s.noNutrition() {
		return
	}
	n := &mother.Nutrition
	m := milk(mother)
	c.Nutrition.DietProtein, c.Nutrition.DietMicro = m.Protein, m.Micro
	c.Nutrition.MicroLack = 0.5 * math.Max(0, n.MicroLack)
	c.Nutrition.Stunt = math.Min(stuntMax, stuntAtBirth*n.deficit()+stuntInherits*n.Stunt)
}

// --- the body's clock ------------------------------------------------------------

// nutrientNeed is c's need per unit of energy relative to an adult's.
func (s *Sim) nutrientNeed(c *Creature) (protein, micro float64) {
	switch age := s.ageYears(c); {
	case age < 2:
		protein, micro = 0.65, 1.4
	case age < 5:
		protein, micro = 0.65, 1.5
	case age < 10:
		protein, micro = 0.7, 1.25
	case age < 15:
		protein, micro = 0.85, 1.15
	case age >= 60:
		protein, micro = 1.15, 1
	default:
		protein, micro = 1, 1
	}
	switch {
	case c.Pregnancy != nil:
		protein, micro = math.Max(protein, 1.3), math.Max(micro, 1.4)
	case c.Sex == Female && s.nursing(c):
		protein, micro = math.Max(protein, 1.2), math.Max(micro, 1.3)
	}
	return protein, micro
}

// adequacy is what c's diet supplies against what c needs (1 = just enough;
// an unknown diet is taken to be just enough).
func (s *Sim) adequacy(c *Creature) (protein, micro float64) {
	if !c.Nutrition.known() {
		return 1, 1
	}
	needP, needM := s.nutrientNeed(c)
	return c.Nutrition.DietProtein / needP, c.Nutrition.DietMicro / needM
}

// nourish moves everyone's nutrition on, each body every nutritionEvery ticks.
func (s *Sim) nourish() {
	if s.noNutrition() {
		return
	}
	for _, c := range s.creatures {
		if (s.tick+c.ID)%nutritionEvery != 0 || c.Health <= 0 {
			continue
		}
		s.nourishBody(c, float64(nutritionEvery)*dt/SecondsPerYear)
	}
}

// nourishBody moves c's status towards what its diet supplies, and a
// child's growth with it, over yr years.
func (s *Sim) nourishBody(c *Creature, yr float64) {
	n := &c.Nutrition
	p, m := s.adequacy(c)
	// What a nearly spent reserve still supplies.
	supply := clamp(c.Energy/lowReserve, 0, 1)
	age := s.ageYears(c)
	tauP, tauM := proteinTurnoverAdult, microTurnoverAdult
	if age < 10 {
		tauP, tauM = proteinTurnoverChild, microTurnoverChild
	}
	towards := func(lack, adequacy, tau float64) float64 {
		target := clamp(1-adequacy, -reserveMax, 1)
		return lack + (target-lack)*(1-math.Exp(-yr/tau))
	}
	n.ProteinLack = towards(n.ProteinLack, supply*p, tauP)
	n.MicroLack = towards(n.MicroLack, supply*m, tauM)
	if age < adultAge/SecondsPerYear {
		s.growBody(c, age, yr)
	}
}

// growthRate is how fast a child grows at an age (years), as a share of its
// height a year: half again in the first year, then slower until the
// adolescent spurt (WHO Child Growth Standards; about 50 → 75 → 87 cm).
func growthRate(age float64) float64 {
	switch {
	case age < 1:
		return 0.45
	case age < 2:
		return 0.15
	case age < 5:
		return 0.08
	case age < 10:
		return 0.055
	}
	return 0.045
}

// growBody loses a child some height when it is short of protein,
// micronutrients or food, or ill, and makes some of it up when it is not.
func (s *Sim) growBody(c *Creature, age, yr float64) {
	n := &c.Nutrition
	dp, dm := n.deficits()
	hunger := clamp((childHungerAt-c.Energy)/(childHungerAt-childHungerTo), 0, 1)
	want := math.Min(1, math.Max(dp, math.Max(0.7*dm, 0.6*hunger))+0.5*sickness(c))
	v := growthRate(age) * yr
	n.Stunt += v * (stuntLoss*want - catchUp*(1-want)*n.Stunt)
	n.Stunt = clamp(n.Stunt, 0, stuntMax)
	if n.Stunt < 1e-6 {
		n.Stunt = 0
	}
}

// --- what it does --------------------------------------------------------------

// malnutritionSeverity is how much worse an infection goes in c's body.
func (s *Sim) malnutritionSeverity(c *Creature) float64 {
	if s.noNutrition() {
		return 1
	}
	m := 1 + sickWorse*c.Nutrition.deficit()
	if s.ageYears(c) < 5 {
		m += stuntSickWorse * c.Nutrition.Stunt
	}
	return m
}

// malnutritionLength is how much longer a bout of d lasts in c's body. Only
// diarrhoea: zinc deficiency prolongs it (Lazzerini & Wanzira 2016), while
// for pneumonia and malaria the trials show no such effect.
func (s *Sim) malnutritionLength(c *Creature, d Disease) float64 {
	if s.noNutrition() || d != Diarrhea {
		return 1
	}
	return 1 + sickLonger*c.Nutrition.deficit()
}

// nutritionFertility scales a woman's monthly chance of conceiving.
func (s *Sim) nutritionFertility(f *Creature) float64 {
	if s.noNutrition() {
		return 1
	}
	p := 1 - fewerConceived*f.Nutrition.deficit()
	if s.ageYears(f) < 18 {
		p *= math.Max(0.5, 1-lateMaturing*f.Nutrition.Stunt)
	}
	return p
}

// nutritionBirthRisk scales the risk that childbirth kills mother f.
func (s *Sim) nutritionBirthRisk(f *Creature) float64 {
	if s.noNutrition() {
		return 1
	}
	_, dm := f.Nutrition.deficits()
	return 1 + anaemicBirth*dm + shortBirth*f.Nutrition.Stunt
}

// nutritionNeonatalRisk scales the risk that a baby of mother f dies in its
// first weeks: small babies of deficient or short mothers die more often.
func (s *Sim) nutritionNeonatalRisk(f *Creature) float64 {
	if s.noNutrition() {
		return 1
	}
	return 1 + smallBaby*f.Nutrition.deficit() + shortBirth*f.Nutrition.Stunt
}

// nutritionHealing scales how fast c's wounds heal.
func (s *Sim) nutritionHealing(c *Creature) float64 {
	if s.noNutrition() {
		return 1
	}
	return 1 - slowHealing*c.Nutrition.deficit()
}

// senseNutrition fills inMalnourished: 0 well fed … 1 severely short of
// protein or micronutrients.
func (s *Sim) senseNutrition(c *Creature, in *[NumInputs]float64) {
	if s.noNutrition() {
		return
	}
	in[inMalnourished] = c.Nutrition.deficit()
}

// --- the observer ------------------------------------------------------------------

// NutritionView is a person's nutrition as the inspector shows it.
type NutritionView struct {
	// Label: "baik", "kurang" or "buruk" (the worse of the protein and
	// micronutrient deficit and how thin they are).
	Label string `json:"label"`
	// Protein and Micro: the body's status, 1 adequate (up to 1.25 with a
	// reserve, 0 empty).
	Protein float64 `json:"protein"`
	Micro   float64 `json:"micro"`
	// Deficit: what the brain feels as "kurang gizi", 0–1.
	Deficit float64 `json:"deficit"`
	// DietProtein and DietMicro: what their food gives against their own
	// need (1 enough).
	DietProtein float64 `json:"dietProtein"`
	DietMicro   float64 `json:"dietMicro"`
	// Need: why they need more than an adult does, if they do ("bayi",
	// "balita", "anak", "remaja", "hamil", "menyusui", "lansia" or "").
	Need string `json:"need"`
	// Stature: height against a well-fed person's of the same age (1 =
	// full); Stunted below about −2 SD.
	Stature float64 `json:"stature"`
	Stunted bool    `json:"stunted"`
	Thin    bool    `json:"thin"`
}

// nutritionLabel is "baik", "kurang" or "buruk".
func nutritionLabel(deficit, energy float64) string {
	switch {
	case deficit >= badFrom || energy < wastedBelow:
		return "buruk"
	case deficit >= poorFrom || energy < thinBelow:
		return "kurang"
	}
	return "baik"
}

// needLabel names why c needs denser food than an adult.
func (s *Sim) needLabel(c *Creature) string {
	switch age := s.ageYears(c); {
	case c.Pregnancy != nil:
		return "hamil"
	case c.Sex == Female && s.nursing(c):
		return "menyusui"
	case age < 2:
		return "bayi"
	case age < 5:
		return "balita"
	case age < 10:
		return "anak"
	case age < 15:
		return "remaja"
	case age >= 60:
		return "lansia"
	}
	return ""
}

func (s *Sim) nutritionView(c *Creature) *NutritionView {
	if s.noNutrition() {
		return nil
	}
	n := &c.Nutrition
	d := n.deficit()
	p, m := s.adequacy(c)
	return &NutritionView{
		Label:       nutritionLabel(d, c.Energy),
		Protein:     r3(1 - n.ProteinLack),
		Micro:       r3(1 - n.MicroLack),
		Deficit:     r3(d),
		DietProtein: r3(p),
		DietMicro:   r3(m),
		Need:        s.needLabel(c),
		Stature:     r3(1 - n.Stunt),
		Stunted:     n.Stunt >= stuntedFrom,
		Thin:        c.Energy < thinBelow,
	}
}

// NutritionInfo counts the island's nutrition now, for the soak runner and
// the health view.
type NutritionInfo struct {
	People       int `json:"people"`
	Deficient    int `json:"deficient"` // short of protein or micronutrients ("kurang" or worse)
	Severe       int `json:"severe"`    // severely ("buruk")
	ProteinShort int `json:"proteinShort"`
	MicroShort   int `json:"microShort"`
	Thin         int `json:"thin"`         // energy reserve low
	Malnourished int `json:"malnourished"` // labelled "kurang" or "buruk" for any reason
	// Stunting, among children under five, children under fifteen and adults.
	Under5         int `json:"under5"`
	Stunted5       int `json:"stunted5"`
	Children       int `json:"children"`
	StuntedChild   int `json:"stuntedChild"`
	Adults         int `json:"adults"`
	StuntedAdults  int `json:"stuntedAdults"`
	PregnantShort  int `json:"pregnantShort"` // pregnant or nursing women short of protein or micronutrients …
	PregnantOrNurs int `json:"pregnantOrNursing"`
	// Mean diet against need (1 enough), and mean height lost by adults.
	DietProtein float64 `json:"dietProtein"`
	DietMicro   float64 `json:"dietMicro"`
	AdultStunt  float64 `json:"adultStunt"`
}

// Nutrition reports the island's nutrition.
func (s *Sim) Nutrition() NutritionInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.nutritionInfo()
}

func (s *Sim) nutritionInfo() NutritionInfo {
	var v NutritionInfo
	if s.noNutrition() {
		return v
	}
	for _, c := range s.creatures {
		n := &c.Nutrition
		dp, dm := n.deficits()
		d := math.Max(dp, dm)
		v.People++
		if d >= poorFrom {
			v.Deficient++
		}
		if d >= badFrom {
			v.Severe++
		}
		if dp >= poorFrom {
			v.ProteinShort++
		}
		if dm >= poorFrom {
			v.MicroShort++
		}
		if c.Energy < thinBelow {
			v.Thin++
		}
		if nutritionLabel(d, c.Energy) != "baik" {
			v.Malnourished++
		}
		p, m := s.adequacy(c)
		v.DietProtein += p
		v.DietMicro += m
		stunted := n.Stunt >= stuntedFrom
		switch age := s.ageYears(c); {
		case age < 5:
			v.Under5++
			v.Children++
			if stunted {
				v.Stunted5++
				v.StuntedChild++
			}
		case age < 15:
			v.Children++
			if stunted {
				v.StuntedChild++
			}
		default:
			v.Adults++
			v.AdultStunt += n.Stunt
			if stunted {
				v.StuntedAdults++
			}
		}
		if c.Sex == Female && (c.Pregnancy != nil || s.nursing(c)) {
			v.PregnantOrNurs++
			if d >= poorFrom {
				v.PregnantShort++
			}
		}
	}
	if v.People > 0 {
		v.DietProtein = r3(v.DietProtein / float64(v.People))
		v.DietMicro = r3(v.DietMicro / float64(v.People))
	}
	if v.Adults > 0 {
		v.AdultStunt = r3(v.AdultStunt / float64(v.Adults))
	}
	return v
}
