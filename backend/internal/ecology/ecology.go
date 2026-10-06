// Package ecology is the living land under the simulation's people: the
// monsoon year with its wet and dry seasons, El Niño droughts and La Niña
// floods; wild plants that grow with the rain and the soil; fields that
// people plant and harvest; fish in rivers and the sea; and wild animals
// that graze, breed, hunt each other and flee from people.
//
// It knows nothing about people beyond the Humans interface, so the
// simulation stays in charge of creatures and the ecology can be tested on
// its own. All randomness comes from the simulation's generator, so a world
// replays exactly.
package ecology

import (
	"math"
	"math/rand/v2"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/world"
)

// SecondsPerYear is the world's time scale: one simulated year lasts 8
// simulated seconds. The simulation uses the same constant.
const SecondsPerYear = 8.0

// Land is what the ecology needs to know about the map.
type Land struct {
	W, H      int
	Ground    []int // world ground tile id per tile
	Objects   []int // world object tile id per tile
	Blocked   []bool
	Water     []bool
	Fresh     []bool // drinkable water: rivers and lakes, not the sea
	NearWater []bool // walkable and next to water
	// The lie of the land for the water cycle (nil in hand-made lands):
	// where each tile's water runs next (-1: the sea), and which tiles are
	// river channels rather than lakes.
	Down  []int32
	River []bool
	// Altitude of each tile from the shore (0) to the peaks (1); nil in
	// hand-made lands, which are taken as lowland.
	Altitude []float32
}

// LandFromMap derives a Land from a map.
func LandFromMap(m *world.Map) Land {
	n := m.Width * m.Height
	l := Land{W: m.Width, H: m.Height, Ground: m.Layers.Ground, Objects: m.Layers.Objects,
		Blocked: make([]bool, n), Water: make([]bool, n), NearWater: make([]bool, n), Fresh: world.FreshWater(m)}
	g := world.BuildGeoModel(m)
	if len(g.Down) == n && len(g.RiverOf) == n {
		l.Down = g.Down
		l.River = make([]bool, n)
		for i, r := range g.RiverOf {
			l.River[i] = r >= 0
		}
	}
	if len(g.Elevation) == n && g.PeakLevel > g.ShoreLevel {
		l.Altitude = make([]float32, n)
		for i, h := range g.Elevation {
			l.Altitude[i] = float32(clamp((h-g.ShoreLevel)/(g.PeakLevel-g.ShoreLevel), 0, 1))
		}
	}
	for i := range n {
		gr, o := m.Layers.Ground[i], m.Layers.Objects[i]
		l.Water[i] = gr == world.Water || gr == world.DeepWater
		l.Blocked[i] = world.GroundTiles[gr].Solid || world.ObjectTiles[o].Solid
	}
	for i := range n {
		if l.Blocked[i] {
			continue
		}
		x, y := i%l.W, i/l.W
		for dy := -1; dy <= 1 && !l.NearWater[i]; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if j, ok := l.index(x+dx, y+dy); ok && l.Water[j] {
					l.NearWater[i] = true
					break
				}
			}
		}
	}
	return l
}

func (l *Land) index(x, y int) (int, bool) {
	if x < 0 || y < 0 || x >= l.W || y >= l.H {
		return 0, false
	}
	return y*l.W + x, true
}

func (l *Land) indexAt(x, y float64) (int, bool) {
	return l.index(int(math.Floor(x)), int(math.Floor(y)))
}

// Options switch parts of the ecology off for experiments.
type Options struct {
	// NoClimate keeps the rain at its yearly average: no seasons, no El Niño.
	NoClimate bool
	// NoFauna leaves the land without wild animals.
	NoFauna bool
	// NoWaterCycle keeps every river and lake full all year (no drying up).
	NoWaterCycle bool
	// NoFire: nothing ever burns. The wind and the storms still blow.
	NoFire bool
}

// Humans is what the ecology may ask about, or do to, people.
type Humans interface {
	// Nearest returns the closest living person within r of (x, y).
	Nearest(x, y, r float64) (id int64, hx, hy float64, ok bool)
	// Crowd counts living people within r of (x, y).
	Crowd(x, y, r float64) int
	// Maul hurts a person; species names the animal, for the death record.
	Maul(id int64, damage float64, species string)
	// Home is where a household's livestock lives: its house, or its pen
	// if it has one nearby. ok is false once the house is gone.
	Home(house int64) (x, y float64, pen, ok bool)
	// Settled returns a house within a few tiles of (x, y), if any;
	// chickens and pigs scavenging there slowly lose their fear of people.
	Settled(x, y float64) (house int64, ok bool)
	// Tamed tells that an animal of species now belongs to household house.
	Tamed(house int64, species string)
	// Pace is how fast person id moved last, 0 (still) to 1 (flat out).
	// Animals notice movement: a hunter creeping up is seen late.
	Pace(id int64) float64
	// Snared tells that the snare on tile caught an animal of species,
	// worth meat units; the snare stays sprung until SetSnares re-arms it.
	Snared(tile int, species string, meat int)
}

// Event is something worth telling the observer.
type Event struct {
	Kind string // "climate", "ecology", "harvest"
	Text string
}

// Ecology is the living land of one world.
type Ecology struct {
	land  Land
	opts  Options
	rng   *rand.Rand
	crops []chem.Crop

	// Derived from the land.
	cover     []uint8 // vegetation class per tile
	sea       []bool  // water connected to the map edge
	fresh     []bool  // rivers and lakes
	nearFresh []bool  // within 2 tiles of fresh water (riparian)
	nearSea   []bool  // within 3 tiles of the sea
	walkable  []int32

	clim  Climate
	hydro hydrology

	// Vegetation per tile.
	forage    []float32 // food people can gather or eat: fruit, tubers, greens
	graze     []float32 // grass and leaves for grazing animals
	fert      []float32 // soil fertility 0–1
	fertBase  []float32 // what fertility recovers to under fallow
	wild      []uint8   // wild crop growing here (index into crops + 1)
	manage    []uint8   // farmland / irrigated / manured bits, set by the simulation
	woodland  []float32 // nearby tree cover 0–1 (1 = untouched), set by the simulation
	deforest  float64   // share of the island's trees that are gone
	fish      []float32 // fish in each water tile
	fishCap   []float32
	moistNow  []float32 // soil moisture per tile at the last growth step
	plots     []*Plot
	plotAt    []int32 // tile → index into plots + 1
	plotVer   int64
	ripeCells [][]int32 // per 8×8 cell, tiles of ripe plots

	fauna fauna
	fire  fireField // fire.go

	events []Event
	stats  Stats
}

// Management bits per tile.
const (
	Farmland  = 1 // inside a ladang: cleared and weeded
	Irrigated = 2 // watered by an irrigation channel
	Manured   = 4 // near a pen with livestock
	SnareSet  = 8 // a snare waits here
)

// New brings a land to life at simulated time t.
func New(l Land, rng *rand.Rand, opts Options, t float64) *Ecology {
	e := newEcology(l, rng, opts)
	e.clim.init(t, opts.NoClimate)
	e.fillHydrology()
	e.seedVegetation()
	if !opts.NoFauna {
		e.populate(t)
	}
	e.clim.Seed = e.weatherSeed()
	return e
}

// weatherSeed is a number of this world's own for the weather's hashed
// chances, taken from where wild crops and animals happened to be placed
// (which the simulation's random generator decided) so that it draws
// nothing more from it.
func (e *Ecology) weatherSeed() uint64 {
	h := uint64(0xCBF29CE484222325)
	mix := func(v uint64) { h = (h ^ v) * 0x100000001B3 }
	for i, w := range e.wild {
		if w != 0 {
			mix(uint64(i)<<8 | uint64(w))
		}
	}
	for _, a := range e.fauna.animals {
		mix(math.Float64bits(a.X))
		mix(math.Float64bits(a.Y))
	}
	return h | 1
}

func newEcology(l Land, rng *rand.Rand, opts Options) *Ecology {
	n := l.W * l.H
	e := &Ecology{
		land:     l,
		opts:     opts,
		rng:      rng,
		crops:    chem.Crops(),
		forage:   make([]float32, n),
		graze:    make([]float32, n),
		fert:     make([]float32, n),
		fertBase: make([]float32, n),
		wild:     make([]uint8, n),
		manage:   make([]uint8, n),
		woodland: make([]float32, n),
		fish:     make([]float32, n),
		fishCap:  make([]float32, n),
		moistNow: make([]float32, n),
		plotAt:   make([]int32, n),
	}
	for i := range e.woodland {
		e.woodland[i] = 1
	}
	e.fire.init(l.W, l.H)
	e.deriveLand()
	e.fauna.init(e)
	e.stats.ready()
	return e
}

// Tick advances the ecology by one simulation tick of dt seconds ending at
// time t. Animals move and fires burn every other tick; plants, fish and
// fields grow, and burnt land greens, every growEvery ticks.
func (e *Ecology) Tick(tick int64, t, dt float64, h Humans) {
	e.fire.tick = tick
	e.clim.step(e, t, dt)
	if tick%hydroEvery == 0 {
		e.stepHydrology(dt * hydroEvery)
		e.waterEvents()
	}
	if tick%faunaEvery == 0 && !e.opts.NoFauna {
		e.fauna.step(e, t, dt*faunaEvery, h)
	}
	if tick%growEvery == 0 {
		e.grow(t, dt*growEvery)
		e.fadeScorch(dt * growEvery)
	}
	if tick%fireEvery == 0 {
		e.burn(dt * fireEvery)
	}
}

const (
	faunaEvery = 2
	growEvery  = 10
)

// TakeEvents returns and clears what happened since the last call.
func (e *Ecology) TakeEvents() []Event {
	ev := e.events
	e.events = nil
	return ev
}

func (e *Ecology) event(kind, text string) {
	e.events = append(e.events, Event{Kind: kind, Text: text})
}

// SetManagement replaces the farmland / irrigation / manure bits.
func (e *Ecology) SetManagement(bits []uint8) {
	if len(bits) == len(e.manage) {
		copy(e.manage, bits)
		e.plotVer++
	}
}

// Management is tile i's farmland / irrigated / manured bits.
func (e *Ecology) Management(i int) uint8 { return e.manage[i] }

// SetWoodland tells the ecology how much of the tree cover around each tile
// is left (1 = untouched, 0 = clear-cut), and the share of all trees gone.
func (e *Ecology) SetWoodland(cover []float32, deforested float64) {
	if len(cover) == len(e.woodland) {
		copy(e.woodland, cover)
	}
	e.deforest = deforested
}

// Deforested is the share of the island's trees that have been cut.
func (e *Ecology) Deforested() float64 { return e.deforest }

func clamp(v, lo, hi float64) float64 { return math.Min(hi, math.Max(lo, v)) }

func clamp32(v, lo, hi float32) float32 { return min(hi, max(lo, v)) }
