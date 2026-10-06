package ecology

import (
	"fmt"
	"math"

	"miniv2/backend/internal/world"
)

// Fire on the land, after RAGE's fire manager (game/vfx/misc/Fire.h and
// Fire.cpp: CFireManager and CFire). There a fire's strength grows to a peak
// set by the flammability and fuel of what burns, holds, and burns out; every
// so often it tries to spread, with a chance scaled by flammability and fuel;
// water puts it out (ExtinguishArea); and a map status grid remembers
// burnt-out ground so it doesn't catch again at once.
//
// Here any tile of vegetation or trees may burn. Its fuel is what the ecology
// already grows there (grass and leaves, wild food plants, a standing crop,
// the canopy of the trees around), dampened by the soil moisture: fuel in
// soil wetter than fireWetLimit doesn't carry fire, so fires belong to the
// dry season, and to El Niño years above all. A burning tile tries every step
// to set its neighbours alight, far more readily downwind and uphill, and in a
// strong wind its embers carry a few tiles ahead. It eats its fuel and
// blackens the ground, then dies down; rain damps it and a storm's rain puts
// it out. The blackened ground (the "status grid") greens again as plants
// regrow, and its ash leaves the soil a little richer for a year or two.
//
// Work is bounded by what burns: an active list of burning tiles, and a list
// of scorched ones that fade every growth step. Nothing burning costs nothing.
// Every chance is hashed from the tile, the tick and the world's weather seed.
const (
	fireEvery    = 2    // ticks between fire steps
	fireCell     = 8    // side of the cells indexing burning tiles, in tiles
	fireWetLimit = 0.45 // soil moisture from which fuel no longer burns (an estimate after the moisture of extinction of fine dead fuels, 25–40 %)
	fireHeat     = 1.3  // dry grass in a fresh breeze burns at about 0.75, in an El Niño gale at full strength
	fireOut      = 0.04 // intensity below which a tile has burnt out
	startHeat    = 0.15 // a newly caught tile, before it takes hold
	fireRamp     = 3.0  // per second a fire grows towards its peak …
	fireFade     = 2.0  // … and dies down once its fuel is spent
	// A fire burns the cured grass and the litter, but grass tillers and
	// tubers survive in the ground: a fierce fire leaves about half the
	// grass and more of the food plants to sprout again.
	burnUp     = 0.8 // share of the grass burnt per second at full intensity (half that of the food plants)
	scorchRate = 0.8 // how fast the ground blackens per second at full intensity …
	smoulder   = 0.4 // … and at least this fast while anything burns, so every fire burns out
	// Spread. Time is compressed (a simulated second is some six weeks), so
	// these are chosen for a fire to cross a few tiles a second in dry grass
	// and to burn out in one to two seconds: estimates, not physics. The wind
	// factor gives a head fire some 7 times the calm rate in a fresh breeze
	// (speed 0.5) and a backing fire a seventh of it, near the 10–60 : 1
	// head-to-back ratios measured in grass fires (Cheney & Sullivan 2008);
	// fire runs uphill faster (McArthur: about twice as fast per 10° of slope).
	spreadRate  = 1.2  // per second, at full intensity into fuel that catches at once
	spreadMin   = 0.12 // a fire this weak no longer spreads
	windSpread  = 4.0  // spread downwind × e^(windSpread·speed), upwind ÷ the same
	slopeSpread = 25.0 // spread uphill × e^(slopeSpread·rise), within e^±1
	spotWind    = 0.4  // wind from which embers carry ahead of the front …
	spotRate    = 3.0  // … at most this often per second from a fully burning tile
	// Rain: it damps the flames, and wet fuel is hard to light.
	rainFrom   = 0.4 // rain (1 = the yearly average) above which it tells
	rainQuench = 1.0 // intensity lost per second per unit of rain above rainFrom
	rainDamp   = 2.0 // catching and a fire's peak ÷ (1 + rainDamp × (rain − rainFrom))
	plotBurn   = 3.0 // crop health lost per second at full intensity
	// Recovery. Burnt grassland greens within a wet season; forest and the
	// trees themselves take years.
	scorchFade = 0.3 // per second in moist soil, for grass
	// Slash and burn: the ash of the burnt vegetation releases phosphorus,
	// potassium and calcium and lowers the soil's acidity, so the first crops
	// after a burn do better; the flush is gone after one or two harvests
	// (Nye & Greenland 1960; Juo & Manu 1996).
	ashBoost = 0.12 // fertility the ash of a fully burnt tile adds
	ashYears = 1.5  // years the ash's nutrients last
	// Sizes for the observer.
	bigFire = 25 // tiles a fire must burn before its end is reported
)

// fireSpec is how a vegetation class burns: its fuel when untouched, how
// readily it catches and carries fire, and how fast the burn scar fades.
type fireSpec struct {
	load, flam, regrow float32
}

var fireSpecs = [numCovers]fireSpec{
	// Alang-alang (Imperata) and other grassland: the most flammable fuel on
	// the islands, burnt nearly every dry season in Nusa Tenggara.
	coverGrass: {load: 0.9, flam: 1, regrow: 1},
	// Closed rain forest keeps a damp microclimate and seldom burns; logged,
	// opened forest dries out and burns readily (Siegert et al. 2001; Cochrane
	// 2003), so flammability rises as the woodland is cut (see fuel).
	coverForest: {load: 0.6, flam: 0.35, regrow: 0.25},
	coverBush:   {load: 0.8, flam: 0.7, regrow: 0.6},
	coverMeadow: {load: 0.7, flam: 0.8, regrow: 1},
	coverSand:   {load: 0.1, flam: 0.1, regrow: 0.5},
	coverBare:   {load: 0.2, flam: 0.3, regrow: 0.5},
}

// treeFire is the canopy of a stand of trees (tree tiles, solid to walk).
var treeFire = fireSpec{load: 0.7, flam: 0.35, regrow: 0.15}

// FireCause is what started a fire.
type FireCause uint8

const (
	FireSpread    FireCause = iota // from a fire already burning nearby
	FireLightning                  // a new fire, struck by lightning
	FireHearth                     // a new fire, from a spark from a hearth or a kiln
)

// FireStats counts fires all-time.
type FireStats struct {
	Started   int `json:"started"`             // fires started
	Lightning int `json:"lightning,omitempty"` // … by lightning
	Hearth    int `json:"hearth,omitempty"`    // … by sparks from hearths and kilns
	Tiles     int `json:"tiles"`               // tiles that caught
	Crops     int `json:"crops,omitempty"`     // plots burnt
	// Filled in by the simulation.
	Buildings int `json:"buildings,omitempty"` // buildings burnt down
}

type fireField struct {
	heat    []float32 // intensity per tile, 0 = not burning
	scorch  []float32 // how blackened the ground is, 0–1
	ash     []float32 // fertility the ash added that the soil has not yet lost
	scorchQ []uint8   // scorch in tenths as last versioned (see scorchLevel)
	active  []int32   // burning tiles, in the order they caught
	marked  []int32   // tiles with scorch or ash, in the order they were first burnt
	cells   [][]int32 // burning tiles per fireCell² cell
	w       int       // the land's width
	cw, ch  int
	ver     int64 // changes when the scorched land does
	episode int   // tiles that caught since nothing burned
	stats   FireStats
	tick    int64 // the current tick, for hashed chances
}

func (f *fireField) init(w, h int) {
	n := w * h
	f.heat = make([]float32, n)
	f.scorch = make([]float32, n)
	f.ash = make([]float32, n)
	f.scorchQ = make([]uint8, n)
	f.w = w
	f.cw, f.ch = (w+fireCell-1)/fireCell, (h+fireCell-1)/fireCell
	f.cells = make([][]int32, f.cw*f.ch)
}

// Salts for the fire's hashed chances.
const (
	saltSpread uint64 = 0xF17E + iota
	saltSpot
	saltSpotFar
	saltSpotSide
	saltSpotCatch
)

func (e *Ecology) fireRoll(a, b int, salt uint64) float64 {
	return hash01(a, b, salt^e.clim.Seed)
}

// canopy reports whether tile i is a stand of trees (solid to walk, but it burns).
func (e *Ecology) canopy(i int) bool {
	l := &e.land
	if i >= len(l.Objects) || i >= len(l.Ground) || l.Water[i] {
		return false
	}
	o, g := l.Objects[i], l.Ground[i]
	return (o == world.Tree || o == world.Pine) && g >= 0 && g < len(world.GroundTiles) && !world.GroundTiles[g].Solid
}

// fuel is the fuel standing on tile i (0–1), how readily it burns, and how
// dry it is (0 at fireWetLimit soil moisture or wetter, 1 bone dry).
func (e *Ecology) fuel(i int) (load, flam float32, dry float64, ok bool) {
	cv := e.cover[i]
	if cv == coverNone {
		if !e.canopy(i) {
			return 0, 0, 0, false
		}
		// Under the trees: the forest's shade keeps the ground damp.
		c := &e.clim
		m := c.Moisture * float64(covers[coverForest].wet)
		if e.nearFresh[i] {
			m = math.Max(m, (riverbankWater*c.Slow-riverbankDry)*e.riverWetness(i))
		}
		w := e.woodland[i]
		return treeFire.load * (0.4 + 0.6*w), treeFire.flam + 0.45*(1-w), dryness(m), true
	}
	m := e.moisture(i)
	fs := &fireSpecs[cv]
	cs := &covers[cv]
	// What is left of what the land holds now: grass that browns in the dry
	// season is still fuel (cured grass burns best), grass eaten or burnt is not.
	g := cs.drought*float32(m) + (1 - cs.drought)
	held := cs.graze*(0.25+0.75*g) + 0.3*cs.forage*(0.45+0.55*g)
	stand := float32(1)
	if held > 0 {
		stand = clamp32((e.graze[i]+0.3*e.forage[i])/held, 0, 1)
	}
	load, flam = fs.load*stand, fs.flam
	if cv == coverForest {
		flam += 0.45 * (1 - e.woodland[i])
	}
	if k := e.plotAt[i]; k != 0 {
		load = min(1, load+0.3*e.plots[k-1].Growth) // a standing crop, dry at harvest time
	}
	return load, flam, dryness(m), true
}

// dryness is how dry fuel is in soil of moisture m: 0 at fireWetLimit or
// wetter, 1 bone dry.
func dryness(m float64) float64 {
	return clamp((fireWetLimit-m)/fireWetLimit, 0, 1)
}

// wetting is how much a downpour damps fire now: 1 in dry weather, less
// the harder it rains.
func (e *Ecology) wetting() float64 {
	return 1 / (1 + rainDamp*math.Max(0, e.clim.Downpour()-rainFrom))
}

// catch is how readily tile i catches fire now, 0–1.
func (e *Ecology) catch(i int) float64 {
	load, flam, d, ok := e.fuel(i)
	if !ok || load <= 0 || d <= 0 {
		return 0
	}
	return float64(flam*load) * d * (1 - 0.8*float64(e.fire.scorch[i])) * e.wetting()
}

// Ignite lets a flame of strength heat (0–1) touch tile i: it catches with
// chance heat × how readily the tile burns now, judged against roll, a
// number in [0, 1) the caller draws. It reports whether the tile caught.
func (e *Ecology) Ignite(i int, heat, roll float64, cause FireCause) bool {
	f := &e.fire
	if e.opts.NoFire || i < 0 || i >= len(f.heat) || f.heat[i] > 0 || roll >= heat*e.catch(i) {
		return false
	}
	e.light(i, startHeat)
	switch cause {
	case FireLightning:
		f.stats.Started++
		f.stats.Lightning++
	case FireHearth:
		f.stats.Started++
		f.stats.Hearth++
	}
	return true
}

// Flare makes tile i burn at least as hot as heat, whatever grows there: a
// building on fire. It reports whether the tile can hold a fire.
func (e *Ecology) Flare(i int, heat float64) bool {
	f := &e.fire
	if e.opts.NoFire || i < 0 || i >= len(f.heat) || e.cover[i] == coverNone && !e.canopy(i) {
		return false
	}
	h := float32(clamp(heat, 0, 1))
	if f.heat[i] > 0 {
		f.heat[i] = max(f.heat[i], h)
		return true
	}
	e.light(i, h)
	return true
}

// light sets tile i alight.
func (e *Ecology) light(i int, heat float32) {
	f := &e.fire
	f.heat[i] = max(heat, fireOut*2)
	f.active = append(f.active, int32(i))
	c := f.cellOf(i)
	f.cells[c] = append(f.cells[c], int32(i))
	f.stats.Tiles++
	f.episode++
}

// neighbours are the eight tiles around one, and how far each is.
var neighbours = [8]struct {
	dx, dy int
	d      float64
}{{1, 0, 1}, {-1, 0, 1}, {0, 1, 1}, {0, -1, 1}, {1, 1, math.Sqrt2}, {1, -1, math.Sqrt2}, {-1, 1, math.Sqrt2}, {-1, -1, math.Sqrt2}}

// burn moves every fire on by dt seconds: it grows or dies down, eats its
// fuel, burns crops, blackens the ground and spreads.
func (e *Ecology) burn(dt float64) {
	f := &e.fire
	if e.opts.NoFire || len(f.active) == 0 {
		return
	}
	c := &e.clim
	l := &e.land
	wind := c.WindSpeed
	wx, wy := math.Cos(c.WindDir), math.Sin(c.WindDir)
	// The pull of the wind on each direction of spread, the same everywhere.
	var pull [8]float64
	for k, nb := range neighbours {
		along := (float64(nb.dx)*wx + float64(nb.dy)*wy) / nb.d
		pull[k] = math.Exp(windSpread*wind*along) / nb.d
	}
	quench := rainQuench * math.Max(0, c.Downpour()-rainFrom)
	fan := (0.75 + 0.5*wind) * e.wetting() * fireHeat
	tick := int(f.tick)
	n := len(f.active)
	for k := range n {
		i := int(f.active[k])
		heat := float64(f.heat[i])
		load, _, dry, _ := e.fuel(i)
		peak := clamp(float64(load)*math.Sqrt(dry)*(1-float64(f.scorch[i]))*fan, 0, 1)
		if peak > heat {
			heat += (peak - heat) * math.Min(1, fireRamp*dt)
		} else {
			heat += (peak - heat) * math.Min(1, fireFade*dt)
		}
		heat -= quench * dt
		if heat < fireOut {
			e.burntOut(i)
			continue
		}
		f.heat[i] = float32(heat)
		// The fire eats the grass, the undergrowth and any crop.
		if e.cover[i] != coverNone {
			e.graze[i] *= float32(1 - math.Min(1, burnUp*heat*dt))
			e.forage[i] *= float32(1 - math.Min(1, 0.5*burnUp*heat*dt))
		}
		if heat > 0.2 {
			if p := e.PlotAt(i); p != nil {
				if p.Health -= float32(plotBurn * heat * dt); p.Health <= 0 {
					e.removePlot(int(e.plotAt[i] - 1))
					e.stats.cur.CropLoss++
					f.stats.Crops++
				}
			}
		}
		if f.scorch[i] == 0 && f.ash[i] == 0 {
			f.marked = append(f.marked, int32(i))
		}
		f.setScorch(i, min(1, f.scorch[i]+float32(scorchRate*math.Max(heat, smoulder)*dt)))
		if heat < spreadMin {
			continue
		}
		x, y := i%l.W, i/l.W
		for d, nb := range neighbours {
			j, ok := l.index(x+nb.dx, y+nb.dy)
			if !ok || f.heat[j] > 0 {
				continue
			}
			// The chance is p × how readily j catches (at most 1), so most
			// rolls are settled before working that out.
			p := spreadRate * heat * pull[d] * e.slope(i, j) * dt
			if r := e.fireRoll(j, tick*8+d, saltSpread); r < p && r < p*e.catch(j) {
				e.light(j, startHeat)
			}
		}
		// Embers: in a strong wind a hot fire throws sparks a few tiles ahead.
		if wind > spotWind && heat > 0.5 && e.fireRoll(i, tick, saltSpot) < spotRate*(wind-spotWind)/(1-spotWind)*heat*dt {
			far := 2 + 3*e.fireRoll(i, tick, saltSpotFar)
			side := 2*e.fireRoll(i, tick, saltSpotSide) - 1
			tx := x + int(math.Round(far*wx-side*wy))
			ty := y + int(math.Round(far*wy+side*wx))
			if j, ok := l.index(tx, ty); ok && f.heat[j] == 0 && e.fireRoll(j, tick, saltSpotCatch) < e.catch(j) {
				e.light(j, startHeat)
			}
		}
	}
	// Drop the tiles that went out, keeping the order.
	kept := f.active[:0]
	for _, t := range f.active {
		if f.heat[t] > 0 {
			kept = append(kept, t)
		}
	}
	f.active = kept
	f.indexCells()
	if len(f.active) == 0 {
		if f.episode >= bigFire {
			e.event("climate", fmt.Sprintf("Kebakaran padam (tahun %d) setelah menghanguskan %d petak", c.Year+1, f.episode))
		}
		f.episode = 0
	}
}

// slope speeds a fire running uphill from i to j and slows one running down.
func (e *Ecology) slope(i, j int) float64 {
	a := e.land.Altitude
	if a == nil {
		return 1
	}
	return math.Exp(clamp(slopeSpread*float64(a[j]-a[i]), -1, 1))
}

// burntOut puts tile i's fire out; its ash enriches the soil a little.
func (e *Ecology) burntOut(i int) {
	f := &e.fire
	f.heat[i] = 0
	if e.cover[i] == coverNone {
		return
	}
	add := min(ashBoost*f.scorch[i], 1-e.fert[i])
	if add > 0 {
		if f.scorch[i] == 0 && f.ash[i] == 0 {
			f.marked = append(f.marked, int32(i))
		}
		e.fert[i] += add
		f.ash[i] += add
	}
}

// indexCells rebuilds the cells of burning tiles.
func (f *fireField) indexCells() {
	for k := range f.cells {
		f.cells[k] = f.cells[k][:0]
	}
	for _, t := range f.active {
		f.cells[f.cellOf(int(t))] = append(f.cells[f.cellOf(int(t))], t)
	}
}

func (f *fireField) cellOf(i int) int {
	return (i/f.w/fireCell)*f.cw + i%f.w/fireCell
}

// scorchLevel is how burnt ground looks to the observer, in tenths (0 for
// less than 0.05), so the scorched land's version changes only when that does.
func scorchLevel(v float32) uint8 {
	if v < 0.05 {
		return 0
	}
	return 1 + uint8(min(v, 1)*10)
}

func (f *fireField) setScorch(i int, v float32) {
	f.scorch[i] = v
	if q := scorchLevel(v); q != f.scorchQ[i] {
		f.scorchQ[i] = q
		f.ver++
	}
}

// fadeScorch lets burnt ground green again as plants regrow over dt seconds,
// faster in moist soil, and the ash's nutrients wash away.
func (e *Ecology) fadeScorch(dt float64) {
	f := &e.fire
	if len(f.marked) == 0 {
		return
	}
	ashLoss := float32(math.Min(1, dt/(ashYears*SecondsPerYear)))
	kept := f.marked[:0]
	for _, t := range f.marked {
		i := int(t)
		if f.heat[i] > 0 {
			kept = append(kept, t)
			continue
		}
		spec, m := treeFire, float32(e.clim.Moisture)
		if cv := e.cover[i]; cv != coverNone {
			spec, m = fireSpecs[cv], e.moistNow[i]
		}
		if s := f.scorch[i]; s > 0 {
			f.setScorch(i, max(0, s-float32(scorchFade*dt)*spec.regrow*(0.15+0.85*m)))
		}
		if a := f.ash[i]; a > 0 {
			d := a * ashLoss
			if a-d < 0.002 {
				d = a // the last trace
			}
			f.ash[i] = a - d
			e.fert[i] = max(0.05, e.fert[i]-d)
		}
		if f.scorch[i] > 0 || f.ash[i] > 0 {
			kept = append(kept, t)
		}
	}
	clear(f.marked[len(kept):])
	f.marked = kept
}

// lightning strikes the land now and then while a storm rages. A strike
// sets dry fuel alight; in a wet storm the rain soon puts most out again.
func (e *Ecology) lightning(t, dt float64) {
	c := &e.clim
	if e.opts.NoFire || c.Storm <= 0 || c.roll(t, saltStrike) >= strikeRate*c.Storm*dt {
		return
	}
	l := &e.land
	x := min(l.W-1, int(c.roll(t, saltStrikeX)*float64(l.W)))
	y := min(l.H-1, int(c.roll(t, saltStrikeY)*float64(l.H)))
	i := y*l.W + x
	if !e.Ignite(i, 1, c.roll(t, saltStrikeCatch), FireLightning) {
		return
	}
	if c.StormFires++; c.StormFires == 1 {
		e.event("climate", fmt.Sprintf("Petir menyambar dan membakar %s (tahun %d)", e.placeName(i), c.Year+1))
	}
}

// placeName is what grows on tile i, for the observer.
func (e *Ecology) placeName(i int) string {
	switch e.cover[i] {
	case coverGrass:
		return "padang rumput"
	case coverForest, coverNone:
		return "hutan"
	case coverBush:
		return "semak belukar"
	case coverMeadow:
		return "padang bunga"
	case coverSand:
		return "tumbuhan pantai"
	}
	return "tanah kering"
}

// fireLandChanged puts out fires on tiles a map edit made water or rock.
func (e *Ecology) fireLandChanged() {
	f := &e.fire
	kept := f.active[:0]
	for _, t := range f.active {
		if i := int(t); e.cover[i] != coverNone || e.canopy(i) {
			kept = append(kept, t)
		} else {
			f.heat[i] = 0
		}
	}
	f.active = kept
	f.indexCells()
}

// --- what the simulation reads ------------------------------------------------

// Burning lists the tiles alight now, in the order they caught; the slice
// must not be kept or changed.
func (e *Ecology) Burning() []int32 { return e.fire.active }

// Heat is how fiercely tile i burns, 0 (not at all) to 1.
func (e *Ecology) Heat(i int) float32 {
	if i < 0 || i >= len(e.fire.heat) {
		return 0
	}
	return e.fire.heat[i]
}

// Scorch is how blackened tile i is by fire, 0–1.
func (e *Ecology) Scorch(i int) float32 {
	if i < 0 || i >= len(e.fire.scorch) {
		return 0
	}
	return e.fire.scorch[i]
}

// Canopy reports whether tile i is a stand of trees, which can burn.
func (e *Ecology) Canopy(i int) bool {
	return i >= 0 && i < len(e.cover) && e.cover[i] == coverNone && e.canopy(i)
}

// FireNear calls fn for every burning tile in the cells around (x, y) within
// r; the caller measures the distance itself.
func (e *Ecology) FireNear(x, y, r float64, fn func(tile int, heat float32)) {
	f := &e.fire
	if len(f.active) == 0 {
		return
	}
	x0, x1 := max(0, int((x-r)/fireCell)), min(f.cw-1, int((x+r)/fireCell))
	y0, y1 := max(0, int((y-r)/fireCell)), min(f.ch-1, int((y+r)/fireCell))
	for cy := y0; cy <= y1; cy++ {
		for cx := x0; cx <= x1; cx++ {
			for _, t := range f.cells[cy*f.cw+cx] {
				if h := f.heat[t]; h > 0 {
					fn(int(t), h)
				}
			}
		}
	}
}

// FireClusters calls fn once for every cell with fire in it, with its
// hottest tile, in a fixed order.
func (e *Ecology) FireClusters(fn func(tile int, heat float32)) {
	f := &e.fire
	if len(f.active) == 0 {
		return
	}
	for _, cell := range f.cells {
		best, hot := -1, float32(0)
		for _, t := range cell {
			if h := f.heat[t]; h > hot {
				best, hot = int(t), h
			}
		}
		if best >= 0 {
			fn(best, hot)
		}
	}
}

// Scorched calls fn for every tile blackened by fire (level 0.05 or more),
// in a fixed order.
func (e *Ecology) Scorched(fn func(tile int, level float32)) {
	f := &e.fire
	for _, t := range f.marked {
		if s := f.scorch[t]; s >= 0.05 {
			fn(int(t), s)
		}
	}
}

// ScorchedCount is how many tiles are blackened by fire.
func (e *Ecology) ScorchedCount() int {
	n := 0
	for _, t := range e.fire.marked {
		if e.fire.scorch[t] >= 0.05 {
			n++
		}
	}
	return n
}

// BurntVersion changes whenever the scorched land looks different.
func (e *Ecology) BurntVersion() int64 { return e.fire.ver }

// FireStats counts fires all-time; the simulation adds the buildings lost.
func (e *Ecology) FireStats() *FireStats { return &e.fire.stats }

// Seed is this world's own number for hashed chances (see weatherSeed).
func (e *Ecology) Seed() uint64 { return e.clim.Seed }

// FireDanger is how dry the driest fuel on the island is now, 0 (nothing
// can burn anywhere) to 1: the simulation skips looking for sparks at 0.
func (e *Ecology) FireDanger() float64 {
	if e.opts.NoFire {
		return 0
	}
	return clamp((fireWetLimit-e.clim.Moisture*float64(covers[coverSand].wet))/fireWetLimit, 0, 1)
}

// --- saving -------------------------------------------------------------------

// fireState is the saved fire: the burning tiles in order with their heat,
// the scorched tiles with their scorch and ash, and the tallies.
type fireState struct {
	Burning []int32   `json:"burning,omitempty"`
	Heat    floats    `json:"heat,omitempty"`
	Marked  []int32   `json:"marked,omitempty"`
	Scorch  floats    `json:"scorch,omitempty"`
	Ash     floats    `json:"ash,omitempty"`
	Version int64     `json:"version,omitempty"`
	Episode int       `json:"episode,omitempty"`
	Stats   FireStats `json:"stats"`
}

func (e *Ecology) saveFire() *fireState {
	f := &e.fire
	st := &fireState{Version: f.ver, Episode: f.episode, Stats: f.stats}
	if len(f.active) > 0 {
		st.Burning = append([]int32(nil), f.active...)
		st.Heat = make(floats, len(f.active))
		for k, t := range f.active {
			st.Heat[k] = f.heat[t]
		}
	}
	if len(f.marked) > 0 {
		st.Marked = append([]int32(nil), f.marked...)
		st.Scorch = make(floats, len(f.marked))
		st.Ash = make(floats, len(f.marked))
		for k, t := range f.marked {
			st.Scorch[k], st.Ash[k] = f.scorch[t], f.ash[t]
		}
	}
	return st
}

// restoreFire puts a saved fire back (nothing burning and no scars in saves
// from before fire).
func (e *Ecology) restoreFire(st *fireState) {
	f := &e.fire
	if st == nil {
		return
	}
	f.ver, f.episode, f.stats = st.Version, st.Episode, st.Stats
	n := int32(len(f.heat))
	if len(st.Heat) == len(st.Burning) {
		for k, t := range st.Burning {
			if t >= 0 && t < n && st.Heat[k] > 0 && f.heat[t] == 0 {
				f.heat[t] = st.Heat[k]
				f.active = append(f.active, t)
			}
		}
	}
	if len(st.Scorch) == len(st.Marked) && len(st.Ash) == len(st.Marked) {
		for k, t := range st.Marked {
			if t >= 0 && t < n && (st.Scorch[k] > 0 || st.Ash[k] > 0) && f.scorch[t] == 0 && f.ash[t] == 0 {
				f.scorch[t], f.ash[t] = st.Scorch[k], st.Ash[k]
				f.scorchQ[t] = scorchLevel(st.Scorch[k])
				f.marked = append(f.marked, t)
			}
		}
	}
	f.indexCells()
}
