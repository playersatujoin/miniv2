package chem

import "slices"

// FarmingTech is learned by whoever first harvests a crop that was planted;
// HerdingTech by whoever first tames a wild animal. Neither needs a recipe
// or a building: people stumble on them while living off the land.
const (
	FarmingTech = "pertanian"
	HerdingTech = "peternakan"
)

// Crop is a food plant people can grow. It is planted from, and harvested
// as, the same item (a grain of padi, a taro corm, a banana or sago sucker, a
// coconut), so eating the whole harvest leaves nothing to plant.
//
// Growing times follow the real plants; yields and fertility use are
// compressed like the rest of the world. Every crop here is grown in
// Indonesia and all but rice grow wild in the region: taro, yams, bananas
// and sago were domesticated in Island Southeast Asia and New Guinea, rice
// arrived with Austronesian farmers.
type Crop struct {
	Item ItemID `json:"item"`
	Name string `json:"name"`
	// GrowYears from planting to the first harvest. Perennials then bear
	// every RepeatYears for LifeYears; annuals are harvested once.
	GrowYears   float64 `json:"growYears"`
	RepeatYears float64 `json:"repeatYears,omitempty"`
	LifeYears   float64 `json:"lifeYears,omitempty"`
	// Yield is units per harvest on fertile ground in good hands. A person
	// eats about one unit a year, so a tile of crops feeds one to three
	// people, ten to forty times what the same tile of wild forest gives:
	// in step with how much denser farmers lived than foragers.
	Yield int `json:"yield"`
	// Water is the soil moisture (0–1) it needs to grow at full speed; far
	// below it the plants wither.
	Water float64 `json:"water"`
	// Drain is how much fertility one harvest takes out of the soil.
	Drain float64 `json:"drain"`
	// Ground lists the ground tile keys it can be planted on.
	Ground []string `json:"ground"`
	// Wet crops only grow beside fresh water or on irrigated land; coastal
	// ones only near the sea.
	Wet     bool `json:"wet,omitempty"`
	Coastal bool `json:"coastal,omitempty"`
	// Wild says where the wild plant grows (ground keys, same Wet/Coastal
	// rules) and on what share of such tiles.
	Wild       []string `json:"wild"`
	WildChance float64  `json:"wildChance"`
}

var crops = []Crop{
	// Rice: about four months to harvest; rain-fed upland rice (padi gogo)
	// in the wet season, all year on irrigated sawah, which also keeps the
	// soil fertile. Wild rice grows in wetlands.
	{Item: "padi", Name: "Padi", GrowYears: 0.35, Yield: 5, Water: 0.75, Drain: 0.12,
		Ground: []string{"grass", "dirt", "forest_floor"}, Wild: []string{"grass"}, WildChance: 0.04},
	// Taro: about eight months, likes wet soil.
	{Item: "talas", Name: "Talas", GrowYears: 0.7, Yield: 5, Water: 0.65, Drain: 0.1,
		Ground: []string{"grass", "dirt", "forest_floor"}, Wild: []string{"forest_floor"}, WildChance: 0.05},
	// Yams (uwi, gembili): about nine months and drought-hardy, a famine food.
	{Item: "ubi", Name: "Ubi", GrowYears: 0.75, Yield: 5, Water: 0.3, Drain: 0.08,
		Ground: []string{"grass", "dirt", "forest_floor", "batuan_vulkanik"}, Wild: []string{"forest_floor"}, WildChance: 0.04},
	// Bananas: fruit after a year, then a bunch every half year or so from
	// new suckers.
	{Item: "pisang", Name: "Pisang", GrowYears: 1, RepeatYears: 0.6, LifeYears: 8, Yield: 2, Water: 0.5, Drain: 0.05,
		Ground: []string{"grass", "dirt", "forest_floor", "batuan_vulkanik"}, Wild: []string{"forest_floor", "grass"}, WildChance: 0.02},
	// Coconut palms bear after about six years, then all year round for decades.
	{Item: "kelapa", Name: "Kelapa", GrowYears: 6, RepeatYears: 0.5, LifeYears: 60, Yield: 2, Water: 0.3, Drain: 0.02,
		Ground: []string{"sand", "grass"}, Coastal: true, Wild: []string{"sand"}, WildChance: 0.08},
	// Sago palms are felled for their starch after about eight to twelve
	// years; suckers grow into the next trunk.
	{Item: "sagu", Name: "Sagu", GrowYears: 8, RepeatYears: 4, LifeYears: 40, Yield: 15, Water: 0.8, Drain: 0.02,
		Ground: []string{"grass", "forest_floor"}, Wet: true, Wild: []string{"grass", "forest_floor"}, WildChance: 0.05},
}

// Crops returns every crop in a stable order.
func Crops() []Crop {
	out := slices.Clone(crops)
	for i := range out {
		out[i].Ground = slices.Clone(out[i].Ground)
		out[i].Wild = slices.Clone(out[i].Wild)
	}
	return out
}

// CropByItem returns the crop grown from item.
func CropByItem(id ItemID) (Crop, bool) {
	for _, c := range crops {
		if c.Item == id {
			return c, true
		}
	}
	return Crop{}, false
}
