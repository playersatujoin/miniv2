package sim

import (
	"math"

	"miniv2/backend/internal/chem"
)

// Foraging knowledge and carried water. Foragers know their land: where the
// fruit trees, tuber patches and sago stands are and when they were last
// worth the walk (e.g. Hadza and Ju/'hoansi camps plan the day's foraging
// from such knowledge; Marlowe 2010, Lee 1979). Until now people here only
// remembered where they last drank, so once the food within sight was gone
// they starved two or three tiles from a full patch they had walked past.
// They now remember a few good places and sense the best one (how good and
// which way); walking there is up to them, the route executor only finds the
// way. They can also carry water in a bamboo tube or gourd, as foragers have
// done for tens of thousands of years (ostrich-eggshell flasks in southern
// Africa by about 60,000 years ago; Texier et al. 2010), so a day's
// foraging needn't stay within reach of the river.

const (
	// maxFoodPlaces is how many food places one person keeps in mind.
	maxFoodPlaces = 5
	// foodPlaceMerge: places closer than this are the same patch.
	foodPlaceMerge = 3.0
	// foodSeenRich: wild food on one tile worth remembering when seen
	// (a ripe field one may harvest counts as 1).
	foodSeenRich = 0.5
	// foodMemoryYears: how long a remembered place stays trusted; wild
	// plants fruit and fail with the seasons, so knowledge fades.
	foodMemoryYears = 2.0
	// foodMemoryReach: a place this far off (tiles) is felt half as strongly.
	foodMemoryReach = 15.0
	// foodVisit: within this distance a remembered place is checked again.
	foodVisit = 1.5
	// foodPlaceMin: a remembered place worth less than this isn't walked to.
	foodPlaceMin = 0.1

	waterTube     chem.ItemID = "tabung_air"
	waterTubeTech             = "wadah_air"
	// tubeHolds is the water one tube carries in hydration units: a quarter
	// of a full body's water, about a year of thirst at this world's time
	// scale. Two tubes at most are carried for water.
	tubeHolds = 0.25
	maxTubes  = 2
)

// FoodPlace is somewhere a person found plenty to eat: how plentiful (0–1)
// and when they last saw it.
type FoodPlace struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Rich float64 `json:"rich"`
	Tick int64   `json:"tick"`
}

// rememberFood notes a food place, merging it with one already known nearby.
func (s *Sim) rememberFood(c *Creature, x, y, rich float64) {
	rich = clamp(rich, 0, 1)
	for i := range c.FoodPlaces {
		p := &c.FoodPlaces[i]
		if math.Hypot(p.X-x, p.Y-y) < foodPlaceMerge {
			if rich >= p.Rich {
				p.X, p.Y = x, y
			}
			p.Rich, p.Tick = rich, s.tick
			return
		}
	}
	if rich < foodSeenRich {
		return // not worth remembering a new place
	}
	place := FoodPlace{x, y, rich, s.tick}
	if len(c.FoodPlaces) < maxFoodPlaces {
		c.FoodPlaces = append(c.FoodPlaces, place)
		return
	}
	worst := 0
	for i := range c.FoodPlaces {
		if s.foodPlaceWorth(c.FoodPlaces[i]) < s.foodPlaceWorth(c.FoodPlaces[worst]) {
			worst = i
		}
	}
	c.FoodPlaces[worst] = place
}

// foodPlaceWorth is how much a remembered place is still to be trusted.
func (s *Sim) foodPlaceWorth(p FoodPlace) float64 {
	years := float64(s.tick-p.Tick) * dt / SecondsPerYear
	return p.Rich * math.Exp(-years/foodMemoryYears)
}

// bestFoodPlace is the remembered place most worth walking to from where c
// stands, and how much (0 when none is).
func (s *Sim) bestFoodPlace(c *Creature) (FoodPlace, float64) {
	var best FoodPlace
	bestV := 0.0
	for _, p := range c.FoodPlaces {
		d := math.Hypot(p.X-c.X, p.Y-c.Y)
		if d < foodVisit {
			continue
		}
		if v := s.foodPlaceWorth(p) / (1 + d/foodMemoryReach); v > bestV {
			best, bestV = p, v
		}
	}
	return best, bestV
}

// noteFood updates c's food places from the senses: a remembered place now
// within reach is checked (found eaten bare, it is soon forgotten), and the
// best food in sight is remembered. Called from sense.
func (s *Sim) noteFood(c *Creature, seenX, seenY, seen float64) {
	for i := range c.FoodPlaces {
		p := &c.FoodPlaces[i]
		if math.Hypot(p.X-c.X, p.Y-c.Y) < foodVisit {
			p.Rich = math.Min(1, float64(s.foodAround(int(p.X), int(p.Y)))/2)
			p.Tick = s.tick
		}
	}
	keep := c.FoodPlaces[:0]
	for _, p := range c.FoodPlaces {
		if p.Rich > 0.05 {
			keep = append(keep, p)
		}
	}
	c.FoodPlaces = keep
	if seen >= foodSeenRich {
		s.rememberFood(c, seenX, seenY, seen)
	}
}

// senseFoodMemory fills inFoodMemory and inFoodMemorySide.
func (s *Sim) senseFoodMemory(c *Creature, in *[NumInputs]float64) {
	p, v := s.bestFoodPlace(c)
	if v <= 0 {
		return
	}
	in[inFoodMemory] = math.Min(1, v)
	in[inFoodMemorySide] = math.Sin(normAngle(math.Atan2(p.Y-c.Y, p.X-c.X) - c.Heading))
}

// --- carried water ------------------------------------------------------------------

// waterRoom is how much water c's tubes can hold.
func (s *Sim) waterRoom(c *Creature) float64 {
	return tubeHolds * float64(min(maxTubes, c.Inventory[waterTube]))
}

// fillWater tops up c's tubes from the source being drunk at, taking the
// water from the river or well and its germs along.
func (s *Sim) fillWater(c *Creature, kind, tile int) {
	room := s.waterRoom(c) - c.WaterCarried
	if room <= 0 {
		return
	}
	add := math.Min(room, drinkRate*dt)
	if kind == drinkWell {
		s.eco.DrawGroundwater(tile, add*waterPerHydration)
	} else {
		s.eco.Drink(tile, add*waterPerHydration)
	}
	germs := 0.0
	if !s.noDisease() {
		germs = s.germsDrunk(kind, tile)
	}
	c.WaterGerms = (c.WaterGerms*c.WaterCarried + germs*add) / (c.WaterCarried + add)
	c.WaterCarried += add
}

// drinkCarried lets c drink from its tubes; it reports whether c drank.
func (s *Sim) drinkCarried(c *Creature) bool {
	room := s.waterRoom(c)
	if c.WaterCarried > room { // a tube was lost or given away
		c.WaterCarried = room
	}
	if c.WaterCarried <= 0 {
		return false
	}
	gain := math.Min(math.Min(1-c.Hydration, drinkRate*dt), c.WaterCarried)
	c.Hydration += gain
	c.WaterCarried -= gain
	if !s.noDisease() {
		c.Swallowed += c.WaterGerms * gain
	}
	if c.WaterCarried <= 0 {
		c.WaterCarried, c.WaterGerms = 0, 0
	}
	return true
}

// tubeWant is how much c would like a water tube (for wants() in
// economy.go): an adult who knows how to make one and carries none.
func (s *Sim) tubeWant(c *Creature) float64 {
	if !s.adult(c) || c.Inventory[waterTube] > 0 || !s.canPractise(c, "alat_batu") {
		return 0
	}
	return 0.4
}
