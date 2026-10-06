package ecology

import "miniv2/backend/internal/chem"

// What wild food is made of, by where it grows (Fase 3d; see chem.Nutrients
// for the scale: 1 = a diet of only this just meets an adult's needs).
// Foragers' plant food is tubers, fruit, nuts, shoots and greens, eaten with
// the insects, grubs, eggs and small animals found alongside; forest and bush
// give the most varied of it. Grassland gives grass seeds, roots and
// grasshoppers, poor in vitamins; bare ground little but roots. The beach
// gives shellfish and crabs, rich in protein, zinc and iodine (coastal
// foragers lived well on them), and riverbanks snails, prawns and frogs.
// These are estimates in the spirit of the USDA-based staple values in
// chem/nutrition.go; forager diets usually meet their protein and
// micronutrient needs (Cordain et al. 2000, Am J Clin Nutr 71:682), mostly
// thanks to the animal part, which here comes from hunting and fishing.
var coverNutrients = [numCovers]chem.Nutrients{
	coverNone:   chem.MixedWild,
	coverGrass:  {Protein: 0.7, Micro: 0.9},
	coverForest: {Protein: 1.0, Micro: 1.3},
	coverBush:   {Protein: 0.9, Micro: 1.2},
	coverMeadow: {Protein: 0.9, Micro: 1.1},
	coverSand:   {Protein: 2.5, Micro: 1.6},
	coverBare:   {Protein: 0.4, Micro: 0.5},
}

const (
	riverbankProtein = 0.3 // snails, prawns, frogs at the water's edge
	riverbankMicro   = 0.1
	coastProtein     = 0.3 // shellfish gleaned on the shore
	coastMicro       = 0.2 // … and the iodine of the sea
)

// ForageNutrients is the nutrient profile of the wild food on tile i: by its
// vegetation, the water nearby, and the wild crop growing there (which is
// part of what is found, as in ForageFind).
func (e *Ecology) ForageNutrients(i int) chem.Nutrients {
	if i < 0 || i >= len(e.cover) {
		return chem.MixedWild
	}
	n := coverNutrients[e.cover[i]]
	if e.nearFresh[i] {
		n.Protein += riverbankProtein
		n.Micro += riverbankMicro
	}
	if e.nearSea[i] && e.cover[i] != coverSand {
		n.Protein += coastProtein
		n.Micro += coastMicro
	}
	if w := e.wild[i]; w != 0 {
		if crop, ok := chem.NutrientsOf(e.crops[w-1].Item); ok {
			n = n.Blend(crop, wildCropOdds)
		}
	}
	return n
}
