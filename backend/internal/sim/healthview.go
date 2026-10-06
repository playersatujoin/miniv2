package sim

import (
	"math"

	"miniv2/backend/internal/ecology"
)

// HealthView is the island's health for the observer (GET /sim/health):
// who is ill with what, the epidemic curve, and maps of the mosquitoes, the
// worm eggs in the soil and the fouled water.
type HealthView struct {
	SecondsPerYear float64       `json:"secondsPerYear"`
	Year           int           `json:"year"`
	Now            EpiPoint      `json:"now"`
	History        []EpiPoint    `json:"history"`
	Diseases       []DiseaseView `json:"diseases"`
	Latrines       int           `json:"latrines"`
	Wells          int           `json:"wells"`
	Strains        int           `json:"strains"`  // respiratory strains so far
	Immunity       float64       `json:"immunity"` // mean Immunity gene
	Map            HealthMap     `json:"map"`
	Off            bool          `json:"off,omitempty"` // disease switched off in this world
}

// DiseaseView is one disease: ill now, all-time deaths, and new bouts per
// person-year over the demography window.
type DiseaseView struct {
	Key       string   `json:"key"`
	Name      string   `json:"name"`
	Ill       int      `json:"ill"`
	Deaths    int      `json:"deaths"`
	Incidence *float64 `json:"incidence"`
}

// HealthMap holds the map overlays. Mosquito, Infected and Soil are per
// cell (Cell tiles square, Cols × Rows, row-major, 0–255); Foul lists the
// fresh water tiles with germs in them and FoulLevel how foul (0–255).
type HealthMap struct {
	Cell      int     `json:"cell"`
	Cols      int     `json:"cols"`
	Rows      int     `json:"rows"`
	Mosquito  []byte  `json:"mosquito"`
	Infected  []byte  `json:"infected"`
	Soil      []byte  `json:"soil"`
	Foul      []int32 `json:"foul"`
	FoulLevel []byte  `json:"foulLevel"`
}

// Scales of the overlays: the value shown at full strength.
const (
	mosquitoFull = 2.0  // breeding sites per cell (an eighth of a cell's tiles in pools)
	soilFull     = 3.0  // worm eggs per cell
	foulShown    = 0.05 // less than this is shown as clean
)

func toByte(v float64) byte { return byte(math.Round(clamp(v, 0, 1) * 255)) }

// Health reports the island's health.
func (s *Sim) Health() HealthView {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := &s.epi
	v := HealthView{
		SecondsPerYear: SecondsPerYear,
		Year:           s.year(),
		History:        append([]EpiPoint{}, e.history...),
		Strains:        e.strains,
		Off:            s.noDisease(),
		Map:            HealthMap{Cell: cellTiles, Cols: e.cw, Rows: e.ch, Foul: []int32{}, FoulLevel: []byte{}},
	}
	if n := len(e.history); n > 0 {
		v.Now = e.history[n-1]
	}
	// The current tallies, fresh rather than as of the last history point.
	var ill [numDiseases]int
	gene := 0.0
	for _, c := range s.creatures {
		if c.Ill != nil {
			ill[c.Ill.Disease]++
		}
		gene += c.Genome.Traits.Immunity
	}
	v.Now.Population = len(s.creatures)
	v.Now.Diarrhea, v.Now.Malaria, v.Now.Respiratory = ill[Diarrhea], ill[Malaria], ill[Respiratory]
	if len(s.creatures) > 0 {
		v.Immunity = r3(gene / float64(len(s.creatures)))
	}
	m := s.stats.metrics(s)
	deaths := [numDiseases]int{s.deathsBy.Diarrhea, s.deathsBy.Malaria, s.deathsBy.Respiratory}
	for d := range numDiseases {
		v.Diseases = append(v.Diseases, DiseaseView{Key: diseaseKey[d], Name: diseaseName[d], Ill: ill[d], Deaths: deaths[d], Incidence: m.Incidence[d]})
	}
	for _, st := range s.structures {
		switch {
		case st.kind.Latrine:
			v.Latrines++
		case st.kind.Well:
			v.Wells++
		}
	}
	mp := &v.Map
	mp.Mosquito = make([]byte, len(e.mosquito))
	mp.Infected = make([]byte, len(e.infected))
	mp.Soil = make([]byte, len(e.soil))
	for k := range e.mosquito {
		mp.Mosquito[k] = toByte(float64(e.mosquito[k]) / mosquitoFull)
		mp.Infected[k] = toByte(float64(e.infected[k]))
		mp.Soil[k] = toByte(float64(e.soil[k]) / soilFull)
	}
	for i, fresh := range s.terrain.fresh {
		if !fresh {
			continue
		}
		if g := s.eco.Germs(i); g >= foulShown {
			mp.Foul = append(mp.Foul, int32(i))
			mp.FoulLevel = append(mp.FoulLevel, toByte(g/ecology.FoulFull))
		}
	}
	return v
}
