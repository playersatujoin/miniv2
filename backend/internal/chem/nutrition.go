package chem

// Nutrition beyond calories (Fase 3d). A food's Food value is its energy;
// Nutrients says what else comes with that energy: protein, and the
// vitamins and minerals lumped into one micronutrient pool. Both are
// densities per unit of energy relative to what a grown-up needs per unit of
// energy, so 1 means a diet of nothing but this food just meets an adult's
// need, however much or little of it is eaten; below 1 such a diet falls
// short. A mixed diet is the energy-weighted mean of its foods.
//
// How the numbers were made (a compressed, abstract model; rounded):
//
//   - Protein: the share of the food's energy that is protein (USDA
//     FoodData Central, SR Legacy, raw foods), times a quality score for
//     how well its amino acids are used (about 0.6 for rice, 0.75 for roots
//     and fruit, 1 for meat and fish; FAO 2013, Dietary protein quality
//     evaluation), divided by the safe protein:energy ratio of a moderately
//     active adult, about 0.10 (WHO/FAO/UNU 2007, Technical Report 935,
//     Table 47). So brown rice, 8.6% protein energy × 0.6 ≈ 0.5; taro and
//     yam about 5% × 0.75 ≈ 0.4; banana 0.35; coconut meat (mostly fat)
//     0.3; sago starch (0.2 g protein per 100 g) almost nothing; game and
//     fish, 55–80% protein energy, 5–8.
//   - Micronutrients: per 1000 kcal of the food, its vitamin A, iron, zinc,
//     iodine, vitamin C and folate against an adult's recommended intake per
//     1000 kcal (FAO/WHO 2004, Vitamin and mineral requirements in human
//     nutrition), each ratio capped at 3 so one abundant nutrient can't stand
//     in for the rest, then averaged. Starches are poor (sago 0.1, rice 0.3,
//     coconut 0.4), roots and fruit middling (vitamin C and folate but
//     little iron, zinc or vitamin A), meat eaten whole the way hunters eat
//     it (liver and other organs included) and fish rich in iron, zinc,
//     vitamin A and iodine, seaweed richest of all per unit of energy.
//
// One pool for all micronutrients rather than separate vitamin A, iron and
// zinc: in diets like these the deficiencies come together (a monotonous
// starchy diet lacks them all, and the same foods - animal foods, greens and
// fruit - relieve them), dietary diversity predicts their joint adequacy
// (Arimond et al. 2010, J Nutr 140:2059S), and the brain could do nothing
// different about each one. Fresh and preserved foods differ little per
// unit of energy; smoking and salting lose some vitamins.
type Nutrients struct {
	Protein float64 `json:"protein"`
	Micro   float64 `json:"micro"`
}

// MixedWild is the profile of mixed wild food gathered to carry home (fruit,
// nuts, tubers, shoots and greens, insects, eggs, shellfish): about what a
// forager's plant-and-small-animal diet gives, just enough protein for an
// adult and a little more than enough vitamins and minerals.
var MixedWild = Nutrients{Protein: 1.0, Micro: 1.2}

var foodNutrients = map[ItemID]Nutrients{
	Food:          MixedWild,
	"rumput_laut": {Protein: 1.1, Micro: 2.5}, // iodine, iron, folate; little energy
	"padi":        {Protein: 0.55, Micro: 0.35},
	"talas":       {Protein: 0.4, Micro: 0.55},
	"ubi":         {Protein: 0.4, Micro: 0.65}, // yams carry vitamin C
	"pisang":      {Protein: 0.35, Micro: 0.7},
	"kelapa":      {Protein: 0.3, Micro: 0.4},
	"sagu":        {Protein: 0.02, Micro: 0.1},
	"daging":      {Protein: 7.5, Micro: 1.5},
	"ikan":        {Protein: 5.5, Micro: 1.3},
	"daging_asap": {Protein: 6.5, Micro: 1.2},
	"ikan_asin":   {Protein: 7.0, Micro: 1.1},
}

// NutrientsOf is the nutrient profile of a food; ok is false for anything
// without one (not food, or a food added without a profile).
func NutrientsOf(id ItemID) (Nutrients, bool) {
	n, ok := foodNutrients[id]
	return n, ok
}

// Blend mixes share (0–1) of o into n, by energy.
func (n Nutrients) Blend(o Nutrients, share float64) Nutrients {
	return Nutrients{Protein: n.Protein + (o.Protein-n.Protein)*share, Micro: n.Micro + (o.Micro-n.Micro)*share}
}
