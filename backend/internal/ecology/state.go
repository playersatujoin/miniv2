package ecology

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"

	"miniv2/backend/internal/world"
)

// floats are saved as base64 of little-endian float32s: compact and exact.
type floats []float32

func (f floats) MarshalJSON() ([]byte, error) {
	buf := make([]byte, 4*len(f))
	for i, v := range f {
		binary.LittleEndian.PutUint32(buf[4*i:], math.Float32bits(v))
	}
	return json.Marshal(base64.StdEncoding.EncodeToString(buf))
}

func (f *floats) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	buf, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	if len(buf)%4 != 0 {
		return fmt.Errorf("ecology: %d bytes is not a whole number of float32s", len(buf))
	}
	*f = make(floats, len(buf)/4)
	for i := range *f {
		(*f)[i] = math.Float32frombits(binary.LittleEndian.Uint32(buf[4*i:]))
	}
	return nil
}

// State is everything about the land that changes, for saving a world.
type State struct {
	Climate Climate   `json:"climate"`
	Forage  floats    `json:"forage"`
	Graze   floats    `json:"graze"`
	Fert    floats    `json:"fert"`
	Fish    floats    `json:"fish"`
	Wood    floats    `json:"woodland"`
	Cut     float64   `json:"deforested"`
	Wild    []byte    `json:"wild"`
	Plots   []*Plot   `json:"plots"`
	PlotVer int64     `json:"plotVer"`
	Animals []*Animal `json:"animals"`
	NextID  int64     `json:"nextAnimal"`
	Start   []int     `json:"start"`
	Extinct []bool    `json:"extinct"`
	Stats   Stats     `json:"stats"`
	// The water cycle (missing in saves from before it; then filled afresh).
	Hydro *hydroState `json:"hydro,omitempty"`
	// Fires and burnt land (missing in saves from before fire: nothing burns).
	Fire *fireState `json:"fire,omitempty"`
}

// State captures the ecology for saving.
func (e *Ecology) State() *State {
	st := &State{
		Climate: e.clim,
		Forage:  e.forage,
		Graze:   e.graze,
		Fert:    e.fert,
		Fish:    e.fish,
		Wood:    e.woodland,
		Cut:     e.deforest,
		Wild:    e.wild,
		Plots:   e.plots,
		PlotVer: e.plotVer,
		Animals: e.fauna.animals,
		NextID:  e.fauna.nextID,
		Start:   e.fauna.start,
		Extinct: e.fauna.extinct,
		Stats:   e.stats,
		Hydro:   e.saveHydrology(),
		Fire:    e.saveFire(),
	}
	// An emptied slice saves like one that never held anything.
	if len(st.Plots) == 0 {
		st.Plots = nil
	}
	if len(st.Animals) == 0 {
		st.Animals = nil
	}
	return st
}

// Restore rebuilds a saved ecology on the same land. It draws nothing from
// rng, so a restored world goes on exactly as the saved one would have.
func Restore(l Land, st *State, rng *rand.Rand, opts Options) (*Ecology, error) {
	e := newEcology(l, rng, opts)
	e.baseVegetation()
	n := l.W * l.H
	for name, a := range map[string]int{"forage": len(st.Forage), "graze": len(st.Graze), "fert": len(st.Fert),
		"fish": len(st.Fish), "wild": len(st.Wild)} {
		if a != n {
			return nil, fmt.Errorf("ecology: saved %s has %d tiles, the land %d", name, a, n)
		}
	}
	copy(e.forage, st.Forage)
	copy(e.graze, st.Graze)
	copy(e.fert, st.Fert)
	copy(e.fish, st.Fish)
	copy(e.wild, st.Wild)
	if len(st.Wood) == n {
		copy(e.woodland, st.Wood)
	}
	e.deforest = st.Cut
	for i, w := range e.wild {
		if int(w) > len(e.crops) {
			e.wild[i] = 0
		}
	}
	e.clim = st.Climate
	e.clim.Off = opts.NoClimate
	e.restoreHydrology(st.Hydro)
	for _, p := range st.Plots {
		ci, ok := e.cropIndex(p.Crop)
		if !ok || p.Tile < 0 || int(p.Tile) >= n || e.plotAt[p.Tile] != 0 {
			continue
		}
		p.ci = ci
		p.stage = p.look()
		e.plots = append(e.plots, p)
		e.plotAt[p.Tile] = int32(len(e.plots))
	}
	e.plotVer = st.PlotVer
	f := &e.fauna
	for _, a := range st.Animals {
		if int(a.Species) >= len(species) || a.Dead() {
			continue
		}
		f.animals = append(f.animals, a)
		f.byID[a.ID] = a
	}
	f.nextID = max(1, st.NextID)
	if len(st.Start) == len(species) {
		copy(f.start, st.Start)
	}
	if len(st.Extinct) == len(species) {
		copy(f.extinct, st.Extinct)
	}
	e.stats = st.Stats
	e.stats.ready()
	f.rebuild(e)
	e.indexRipe()
	e.restoreFire(st.Fire)
	if e.clim.Seed == 0 {
		// A save from before the wind: the weather takes a seed of its own now.
		e.clim.Seed = e.weatherSeed()
	}
	return e, nil
}

// baseVegetation sets what doesn't change: natural fertility and how many
// fish each water holds.
func (e *Ecology) baseVegetation() {
	l := &e.land
	for _, i32 := range e.walkable {
		i := int(i32)
		f := covers[e.cover[i]].fert
		switch l.Ground[i] {
		case world.VolcanicRock:
			f = 0.9 // young volcanic soils (andosols) are among the richest
		case world.Limestone:
			f = 0.45
		}
		if e.nearFresh[i] {
			f += 0.15 // river silt
		}
		e.fertBase[i] = clamp32(f, 0.05, 1)
	}
	for i, w := range l.Water {
		switch {
		case !w:
			e.fishCap[i] = 0
		case e.sea[i]:
			e.fishCap[i] = 1.5
		default:
			e.fishCap[i] = 1
		}
	}
}

// UpdateLand applies an edited map: new land grows wild plants, plots and
// animals on tiles that became solid are lost or moved.
func (e *Ecology) UpdateLand(l Land) {
	if l.W != e.land.W || l.H != e.land.H {
		return
	}
	wasBlocked := e.land.Blocked
	e.land = l
	e.deriveLand()
	// The rivers may run differently now: fill them afresh.
	e.fillHydrology()
	e.baseVegetation()
	for _, i32 := range e.walkable {
		if i := int(i32); wasBlocked[i] {
			cs := covers[e.cover[i]]
			e.forage[i], e.graze[i], e.fert[i] = cs.forage, cs.graze, e.fertBase[i]
		}
	}
	for i := range e.fish {
		if e.fishCap[i] == 0 {
			e.fish[i] = 0
		} else {
			e.fish[i] = min(max(e.fish[i], e.fishCap[i]/2), e.fishCap[i])
		}
	}
	for k := len(e.plots) - 1; k >= 0; k-- {
		if l.Blocked[e.plots[k].Tile] {
			e.removePlot(k)
		}
	}
	for _, a := range e.fauna.animals {
		if !e.passable(a.X, a.Y) {
			if x, y, ok := e.nearestOpen(a.X, a.Y); ok {
				a.X, a.Y = x, y
			} else {
				a.Health = 0
			}
		}
	}
	e.fauna.rebuild(e)
	e.indexRipe()
	e.fireLandChanged()
}

func (e *Ecology) nearestOpen(x, y float64) (float64, float64, bool) {
	cx, cy := int(x), int(y)
	for r := 1; r < 20; r++ {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if max(abs(dx), abs(dy)) != r {
					continue
				}
				if i, ok := e.land.index(cx+dx, cy+dy); ok && !e.land.Blocked[i] {
					return float64(cx+dx) + 0.5, float64(cy+dy) + 0.5, true
				}
			}
		}
	}
	return 0, 0, false
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
