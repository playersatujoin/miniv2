package ecology

import (
	"math/rand/v2"
	"testing"

	"miniv2/backend/internal/chem"
)

func TestForageNutrientsFollowTheLand(t *testing.T) {
	e := New(starterLand(64), rand.New(rand.NewPCG(1, 2)), Options{}, 0)
	seen := map[uint8]bool{}
	for _, i32 := range e.walkable {
		i := int(i32)
		n := e.ForageNutrients(i)
		if n.Protein <= 0 || n.Micro <= 0 || n.Protein > 5 || n.Micro > 5 {
			t.Fatalf("tile %d (cover %d): %+v", i, e.cover[i], n)
		}
		seen[e.cover[i]] = true
		if e.wild[i] != 0 {
			bare := coverNutrients[e.cover[i]]
			crop, _ := chem.NutrientsOf(e.crops[e.wild[i]-1].Item)
			if !e.nearFresh[i] && !e.nearSea[i] && crop.Protein < bare.Protein && n.Protein >= bare.Protein {
				t.Errorf("tile %d: a wild staple should dilute the protein of what is found there", i)
			}
		}
	}
	if len(seen) < 3 {
		t.Errorf("only %d vegetation classes on the test island", len(seen))
	}
	if coverNutrients[coverSand].Protein <= coverNutrients[coverForest].Protein ||
		coverNutrients[coverForest].Micro <= coverNutrients[coverGrass].Micro ||
		coverNutrients[coverBare].Protein >= coverNutrients[coverGrass].Protein {
		t.Error("shellfish beaches > forest > grass > bare ground")
	}
	if n := e.ForageNutrients(-1); n != chem.MixedWild {
		t.Errorf("off the map: %+v", n)
	}
}
