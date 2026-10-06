package chem

import (
	"math"
	"testing"
)

func TestEveryFoodHasANutrientProfile(t *testing.T) {
	for _, it := range Items() {
		n, ok := NutrientsOf(it.ID)
		if it.Food <= 0 {
			if ok {
				t.Errorf("%s is not food but has a nutrient profile", it.ID)
			}
			continue
		}
		if !ok {
			t.Errorf("food %s has no nutrient profile", it.ID)
			continue
		}
		for name, v := range map[string]float64{"protein": n.Protein, "micro": n.Micro} {
			if math.IsNaN(v) || v <= 0 || v > 10 {
				t.Errorf("%s: %s density %.2f out of range", it.ID, name, v)
			}
		}
	}
	for id := range foodNutrients {
		if it, ok := ItemByID(id); !ok || it.Food <= 0 {
			t.Errorf("nutrient profile for %s, which is not a food item", id)
		}
	}
}

// Starchy staples are energy-rich but protein-poor; meat and fish are rich
// in protein and micronutrients; wild food and seaweed carry the vitamins.
func TestNutrientProfilesFollowFoodComposition(t *testing.T) {
	get := func(id ItemID) Nutrients {
		n, _ := NutrientsOf(id)
		return n
	}
	for _, id := range []ItemID{"padi", "talas", "ubi", "pisang", "kelapa", "sagu"} {
		if n := get(id); n.Protein >= 1 || n.Micro >= 1 {
			t.Errorf("staple %s should fall short of an adult's needs on its own: %+v", id, n)
		}
	}
	for _, id := range []ItemID{"daging", "ikan", "daging_asap", "ikan_asin"} {
		if n := get(id); n.Protein < 3 || n.Micro < 1 {
			t.Errorf("animal food %s should be rich in protein and micronutrients: %+v", id, n)
		}
	}
	if get("sagu").Protein >= get("padi").Protein {
		t.Error("sago starch has less protein than rice")
	}
	if MixedWild.Protein < 1 || MixedWild.Micro < 1 {
		t.Errorf("mixed wild food should meet an adult's needs: %+v", MixedWild)
	}
	if get("rumput_laut").Micro <= get("ikan").Micro {
		t.Error("seaweed is the richest in micronutrients per unit of energy")
	}
	// A little fish makes a rice diet adequate in protein, not in micronutrients.
	mix := get("padi").Blend(get("ikan"), 0.15)
	if mix.Protein < 1 || mix.Micro >= 1 {
		t.Errorf("rice with 15%% fish: %+v", mix)
	}
}
