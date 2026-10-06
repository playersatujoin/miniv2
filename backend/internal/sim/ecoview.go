package sim

import (
	"encoding/base64"
	"math"
	"slices"
	"strconv"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/ecology"
)

// EcoPoint is the land every 5 simulated seconds.
type EcoPoint struct {
	Time     float64                   `json:"time"`
	Animals  [ecology.SpeciesCount]int `json:"animals"` // wild and tame
	Tame     int                       `json:"tame"`
	Forest   float64                   `json:"forest"` // share of the trees still standing
	Plots    int                       `json:"plots"`
	Ripe     int                       `json:"ripe"`
	Food     int                       `json:"food"` // food units people carry and store
	Wild     float64                   `json:"wild"` // wild food standing on the land, units
	Moisture float64                   `json:"moisture"`
	ENSO     int                       `json:"enso"`
	// Fresh water: shares of river and lake tiles running (or full), standing
	// in pools, and dry; the median groundwater level.
	Running     float64 `json:"running"`
	Pools       float64 `json:"pools"`
	DryBeds     float64 `json:"dryBeds"`
	Groundwater float64 `json:"groundwater"`
}

func (s *Sim) sampleEcology() {
	wild, tame := s.eco.Counts()
	p := EcoPoint{Time: s.time(), Forest: r3(1 - s.eco.Deforested()), Moisture: r3(s.eco.Climate().Moisture), ENSO: s.eco.Climate().ENSO}
	for i := range p.Animals {
		p.Animals[i] = wild[i] + tame[i]
		p.Tame += tame[i]
	}
	for _, pl := range s.eco.Plots() {
		p.Plots++
		if pl.IsRipe() {
			p.Ripe++
		}
	}
	p.Food = s.foodStock().total()
	f, _, _ := s.eco.Totals()
	p.Wild = math.Round(f)
	if w := s.eco.Water(); w.Flowing+w.Pools+w.Under+w.Dry > 0 {
		total := float64(w.Flowing + w.Pools + w.Under + w.Dry)
		p.Running, p.Pools = r3(float64(w.Flowing)/total), r3(float64(w.Pools)/total)
		p.DryBeds = r3(float64(w.Under+w.Dry) / total)
		p.Groundwater = r3(w.Groundwater)
	}
	s.ecoHistory = append(s.ecoHistory, p)
	if n := len(s.ecoHistory) - historyMax; n > 0 {
		s.ecoHistory = slices.Delete(s.ecoHistory, 0, n)
	}
}

// WaterPoint is the island's fresh water at one moment: shares of river and
// lake tiles running, in pools and dry, and the median groundwater.
type WaterPoint struct {
	Time        float64 `json:"time"`
	Running     float64 `json:"running"`
	Pools       float64 `json:"pools"`
	DryBeds     float64 `json:"dryBeds"`
	Groundwater float64 `json:"groundwater"`
}

const (
	waterEvery      = 7   // ticks: about twice a simulated month
	waterHistoryMax = 460 // about twenty years
)

func (s *Sim) sampleWater() {
	w := s.eco.Water()
	total := float64(w.Flowing + w.Pools + w.Under + w.Dry)
	if total == 0 {
		return
	}
	s.waterHistory = append(s.waterHistory, WaterPoint{Time: s.time(), Running: r3(float64(w.Flowing) / total),
		Pools: r3(float64(w.Pools) / total), DryBeds: r3(float64(w.Under+w.Dry) / total), Groundwater: r3(w.Groundwater)})
	if n := len(s.waterHistory) - waterHistoryMax; n > 0 {
		s.waterHistory = slices.Delete(s.waterHistory, 0, n)
	}
}

// FoodStock is the food people hold, in units.
type FoodStock struct {
	Carried int `json:"carried"`
	Stored  int `json:"stored"`  // in houses
	Granary int `json:"granary"` // in granaries
}

func (f FoodStock) total() int { return f.Carried + f.Stored + f.Granary }

func (s *Sim) foodStock() FoodStock {
	var f FoodStock
	for _, c := range s.creatures {
		f.Carried += s.cat.foodUnits(c.Inventory)
	}
	for _, st := range s.structures {
		switch {
		case st.kind.Granary:
			f.Granary += s.cat.foodUnits(st.Storage)
		case st.house():
			f.Stored += s.cat.foodUnits(st.Storage)
		}
	}
	return f
}

type SpeciesView struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Domestic string `json:"domestic,omitempty"` // its name once tamed
	Diet     string `json:"diet"`               // grazer | forager | predator
	Wild     int    `json:"wild"`
	Tame     int    `json:"tame"`
	Start    int    `json:"start"`
	Hunted   int    `json:"hunted"` // killed by people, all-time
	Extinct  bool   `json:"extinct"`
}

type CropView struct {
	Item      string  `json:"item"`
	Name      string  `json:"name"`
	GrowYears float64 `json:"growYears"`
	Perennial bool    `json:"perennial"`
	Plots     int     `json:"plots"`
	Ripe      int     `json:"ripe"`
}

// EcologyView is the land for the observer's Ekologi panel.
type EcologyView struct {
	SecondsPerYear float64              `json:"secondsPerYear"`
	Year           int                  `json:"year"`
	Phase          float64              `json:"phase"` // fraction of the year, 0 = 1 January
	Month          int                  `json:"month"` // 1–12
	Season         string               `json:"season"`
	ENSO           string               `json:"enso"`
	Rain           float64              `json:"rain"` // 1 = the yearly average
	Moisture       float64              `json:"moisture"`
	Light          float64              `json:"light"`
	Forest         float64              `json:"forest"`
	Species        []SpeciesView        `json:"species"`
	Crops          []CropView           `json:"crops"`
	Plots          int                  `json:"plots"`
	Ripe           int                  `json:"ripe"`
	Food           FoodStock            `json:"food"`
	History        []EcoPoint           `json:"history"`
	Years          []ecology.YearRecord `json:"years"`
	Water          ecology.WaterSummary `json:"water"`
	WaterHistory   []WaterPoint         `json:"waterHistory"`
	// Wells, and how many still reach the groundwater.
	Wells    int `json:"wells"`
	WellsDry int `json:"wellsDry"`
}

func seasonName(e *ecology.Ecology) string {
	if e.Wet() {
		return "hujan"
	}
	return "kemarau"
}

var dietNames = map[ecology.Diet]string{ecology.Grazer: "grazer", ecology.Forager: "forager", ecology.Predator: "predator"}

// Ecology reports the state of the land, its animals and fields.
func (s *Sim) Ecology() EcologyView {
	s.mu.Lock()
	defer s.mu.Unlock()
	cl := s.eco.Climate()
	v := EcologyView{
		SecondsPerYear: SecondsPerYear,
		Year:           s.year(),
		Phase:          r3(cl.Phase),
		Month:          int(cl.Phase*12) + 1,
		Season:         seasonName(s.eco),
		ENSO:           ecology.ENSOName(cl.ENSO),
		Rain:           r3(cl.Rain),
		Moisture:       r3(cl.Moisture),
		Light:          r3(cl.Light),
		Forest:         r3(1 - s.eco.Deforested()),
		Species:        []SpeciesView{},
		Crops:          []CropView{},
		Food:           s.foodStock(),
		History:        append([]EcoPoint{}, s.ecoHistory...),
		Years:          append([]ecology.YearRecord{}, s.eco.Years()...),
		Water:          s.eco.Water(),
		WaterHistory:   append([]WaterPoint{}, s.waterHistory...),
	}
	for _, st := range s.structures {
		if !st.kind.Well {
			continue
		}
		v.Wells++
		if i, ok := s.terrain.index(st.X, st.Y); ok && s.eco.Groundwater(i) < ecology.WellMin {
			v.WellsDry++
		}
	}
	wild, tame := s.eco.Counts()
	start, extinct, hunted := s.eco.StartCounts(), s.eco.Extinct(), s.eco.HuntedTotal()
	for i, sp := range speciesInfo {
		v.Species = append(v.Species, SpeciesView{ID: sp.ID, Name: sp.Name, Domestic: sp.Tame, Diet: dietNames[sp.Diet],
			Wild: wild[i], Tame: tame[i], Start: start[i], Hunted: hunted[i], Extinct: extinct[i]})
	}
	byCrop := map[chem.ItemID]*CropView{}
	for _, c := range chem.Crops() {
		v.Crops = append(v.Crops, CropView{Item: string(c.Item), Name: c.Name, GrowYears: c.GrowYears, Perennial: c.RepeatYears > 0})
	}
	for i := range v.Crops {
		byCrop[chem.ItemID(v.Crops[i].Item)] = &v.Crops[i]
	}
	for _, p := range s.eco.Plots() {
		v.Plots++
		cv := byCrop[p.Crop]
		if cv != nil {
			cv.Plots++
		}
		if p.IsRipe() {
			v.Ripe++
			if cv != nil {
				cv.Ripe++
			}
		}
	}
	return v
}

// livestockView lists a household's animals by kind.
func (s *Sim) livestockView(house int64) []StackView {
	out := []StackView{}
	count := map[uint8]int{}
	for _, a := range s.eco.Livestock(house) {
		count[a.Species]++
	}
	for i, sp := range speciesInfo {
		if n := count[uint8(i)]; n > 0 {
			out = append(out, StackView{Item: sp.ID, Name: sp.Tame, Qty: n})
		}
	}
	return out
}

// Plot flags in the fields stream; keep in sync with PLOT_FLAG in frontend/src/sim/protocol.ts.
const (
	plotWithered = 1
	plotIrrigate = 2
	plotFarmland = 4
	plotManured  = 8
)

// Fields returns the encoded `fields` stream message (every planted plot)
// and its version, which changes whenever a plot appears, goes or changes look.
func (s *Sim) Fields() ([]byte, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	plots := s.eco.Plots()
	ver := s.eco.PlotVersion()
	b := make([]byte, 0, 32+len(plots)*20)
	b = append(b, `{"v":`...)
	b = strconv.AppendInt(b, ver, 10)
	b = append(b, `,"p":[`...)
	w := s.terrain.w
	for k, p := range plots {
		if k > 0 {
			b = append(b, ',')
		}
		flags := 0
		if p.Health < 0.5 {
			flags |= plotWithered
		}
		mg := s.eco.Management(int(p.Tile))
		if mg&ecology.Irrigated != 0 {
			flags |= plotIrrigate
		}
		if mg&ecology.Farmland != 0 {
			flags |= plotFarmland
		}
		if mg&ecology.Manured != 0 {
			flags |= plotManured
		}
		b = append(b, '[')
		b = strconv.AppendInt(b, int64(int(p.Tile)%w), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(int(p.Tile)/w), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(s.eco.CropIndex(p.Crop)), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(p.Stage()), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(flags), 10)
		b = append(b, ']')
	}
	return append(b, "]}"...), ver
}

// FieldsVersion is cheap to poll to see whether Fields changed.
// Water encodes the state of every river and lake tile for the stream:
// {"v":version,"t":[tile indices],"s":"states","l":"levels","g":"germs"},
// states, levels and germs (0–255) as base64 bytes in the order of the tiles.
func (s *Sim) Water() ([]byte, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tiles, state, level, foul := s.eco.WaterTiles()
	ver := s.eco.WaterVersion()
	b := make([]byte, 0, 64+len(tiles)*8)
	b = append(b, `{"v":`...)
	b = strconv.AppendInt(b, ver, 10)
	b = append(b, `,"t":[`...)
	for k, t := range tiles {
		if k > 0 {
			b = append(b, ',')
		}
		b = strconv.AppendInt(b, int64(t), 10)
	}
	b = append(b, `],"s":"`...)
	b = base64.StdEncoding.AppendEncode(b, state)
	b = append(b, `","l":"`...)
	b = base64.StdEncoding.AppendEncode(b, level)
	if !s.noDisease() {
		b = append(b, `","g":"`...)
		b = base64.StdEncoding.AppendEncode(b, foul)
	}
	b = append(b, `"}`...)
	return b, ver
}

// Mosquitoes encodes the mosquito map for the stream:
// {"c":cell tiles,"w":cols,"h":rows,"m":"density","i":"infected"}, both
// 0–255 per cell as base64 bytes, row-major.
func (s *Sim) Mosquitoes() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := &s.epi
	m := make([]byte, len(e.mosquito))
	z := make([]byte, len(e.infected))
	for k := range e.mosquito {
		m[k] = toByte(float64(e.mosquito[k]) / mosquitoFull)
		z[k] = toByte(float64(e.infected[k]))
	}
	b := make([]byte, 0, 64+len(m)*3)
	b = append(b, `{"c":`...)
	b = strconv.AppendInt(b, cellTiles, 10)
	b = append(b, `,"w":`...)
	b = strconv.AppendInt(b, int64(e.cw), 10)
	b = append(b, `,"h":`...)
	b = strconv.AppendInt(b, int64(e.ch), 10)
	b = append(b, `,"m":"`...)
	b = base64.StdEncoding.AppendEncode(b, m)
	b = append(b, `","i":"`...)
	b = base64.StdEncoding.AppendEncode(b, z)
	b = append(b, `"}`...)
	return b
}

func (s *Sim) WaterVersion() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.eco.WaterVersion()
}

func (s *Sim) FieldsVersion() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.eco.PlotVersion()
}

// appendWeather writes the frame's weather tuple: year phase, daylight, soil
// moisture, ENSO state, rainfall, then the wind (see appendWindTail).
func (s *Sim) appendWeather(b []byte) []byte {
	cl := s.eco.Climate()
	b = append(b, `,"w":[`...)
	b = strconv.AppendFloat(b, cl.Phase, 'f', 3, 64)
	b = append(b, ',')
	b = strconv.AppendFloat(b, cl.Light, 'f', 2, 64)
	b = append(b, ',')
	b = strconv.AppendFloat(b, cl.Moisture, 'f', 2, 64)
	b = append(b, ',')
	b = strconv.AppendInt(b, int64(cl.ENSO), 10)
	b = append(b, ',')
	b = strconv.AppendFloat(b, cl.Rain, 'f', 2, 64)
	b = s.appendWindTail(b)
	return append(b, ']')
}

// appendAnimals writes the frame's animals: id, species, x, y, heading, flags.
func (s *Sim) appendAnimals(b []byte) []byte {
	return s.appendViewAnimals(b, nil)
}

func (s *Sim) appendViewAnimals(b []byte, view *Viewport) []byte {
	b = append(b, `,"a":[`...)
	t := s.time()
	first := true
	for _, a := range s.eco.Animals() {
		if a.Dead() || !view.contains(a.X, a.Y) {
			continue
		}
		if !first {
			b = append(b, ',')
		}
		first = false
		b = append(b, '[')
		b = strconv.AppendInt(b, a.ID, 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(a.Species), 10)
		b = append(b, ',')
		b = strconv.AppendFloat(b, a.X, 'f', 2, 64)
		b = append(b, ',')
		b = strconv.AppendFloat(b, a.Y, 'f', 2, 64)
		b = append(b, ',')
		b = strconv.AppendFloat(b, a.Heading, 'f', 2, 64)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(a.Flags(t)), 10)
		b = append(b, ']')
	}
	return append(b, ']')
}

// TechsKnown lists the technologies the world has invented so far.
func (s *Sim) TechsKnown() map[string]bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]bool, len(s.techs))
	for id := range s.techs {
		out[id] = true
	}
	return out
}
