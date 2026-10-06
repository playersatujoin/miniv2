package ecology

import "slices"

// SpeciesCount is the number of wild animal species; per-species tallies are
// arrays of this length.
const SpeciesCount = 5

func init() {
	if len(species) != SpeciesCount {
		panic("ecology: SpeciesCount out of date")
	}
}

const yearLogMax = 400

// YearRecord is one simulated year of the land: its weather, its harvests
// and its animals. The simulation adds what happened to people.
type YearRecord struct {
	Year          int               `json:"year"` // 1-based
	ENSO          int               `json:"enso"`
	Rain          float64           `json:"rain"` // mean rainfall, 1 = an average year
	Flood         bool              `json:"flood,omitempty"`
	Planted       int               `json:"planted"`
	Harvest       int               `json:"harvest"`  // units harvested
	CropLoss      int               `json:"cropLoss"` // plots lost to drought, floods or raiders
	Withered      int               `json:"withered"` // … of which to drought
	Wasted        int               `json:"wasted"`   // plots left to rot unharvested
	Fished        int               `json:"fished"`
	Hunted        [SpeciesCount]int `json:"hunted"`
	Predated      [SpeciesCount]int `json:"predated"` // killed by tigers
	Slaughtered   int               `json:"slaughtered"`
	Tamed         int               `json:"tamed"`
	AnimalBirths  [SpeciesCount]int `json:"animalBirths"`
	AnimalStarved [SpeciesCount]int `json:"animalStarved"`
	Animals       [SpeciesCount]int `json:"animals"` // alive at the end of the year
	Extinct       []string          `json:"extinct,omitempty"`
	Arrived       []string          `json:"arrived,omitempty"`

	// Filled in by the simulation.
	Population int `json:"population"`
	Births     int `json:"births"`
	Starved    int `json:"starved"` // people who starved
	Deaths     int `json:"deaths"`
	Rotten     int `json:"rotten"` // food units that went off
	// Energy people got from each kind of food (an adult burns roughly half
	// a unit of energy a year): wild plants, harvested crops, fish, meat.
	FoodWild  float64 `json:"foodWild,omitempty"`
	FoodCrops float64 `json:"foodCrops,omitempty"`
	FoodFish  float64 `json:"foodFish,omitempty"`
	FoodMeat  float64 `json:"foodMeat,omitempty"`
}

// Stats is the ecology's record: this year so far, closed years, and
// all-time hunting tallies.
type Stats struct {
	Cur    YearRecord        `json:"cur"`
	Log    []YearRecord      `json:"log"`
	Hunted [SpeciesCount]int `json:"hunted"`

	cur *YearRecord // alias of Cur, kept for brevity
}

func (s *Stats) ready() { s.cur = &s.Cur }

func (s *Stats) endYear(e *Ecology, year, enso int, rain float64) {
	s.ready()
	r := s.Cur
	r.Year = year + 1
	r.ENSO = enso
	r.Rain = rain
	wild, tame := e.Counts()
	for i := range r.Animals {
		r.Animals[i] = wild[i] + tame[i]
	}
	s.Log = append(s.Log, r)
	if over := len(s.Log) - yearLogMax; over > 0 {
		s.Log = slices.Delete(s.Log, 0, over)
	}
	if r.Withered >= 5 {
		e.event("climate", e.describeHarvestFailure(r.Withered))
	}
	s.Cur = YearRecord{Population: r.Population}
	e.recolonise(float64(year+1) * SecondsPerYear)
}

// Current is this year's record so far; the simulation adds to it.
func (e *Ecology) Current() *YearRecord {
	e.stats.ready()
	return e.stats.cur
}

// Years are the closed years, oldest first (at most 400).
func (e *Ecology) Years() []YearRecord { return e.stats.Log }

// HuntedTotal is how many of each species people have killed, all-time.
func (e *Ecology) HuntedTotal() [SpeciesCount]int { return e.stats.Hunted }

// StartCounts are how many of each species lived on the land at the start.
func (e *Ecology) StartCounts() []int { return slices.Clone(e.fauna.start) }

// Extinct reports which species have died out on the island.
func (e *Ecology) Extinct() []bool { return slices.Clone(e.fauna.extinct) }
