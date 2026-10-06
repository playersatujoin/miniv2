package ecology

import (
	"fmt"
	"math"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/world"
)

// Vegetation classes of walkable tiles.
const (
	coverNone = iota // water, mountains, trees and other solid tiles
	coverGrass
	coverForest
	coverBush
	coverMeadow
	coverSand
	coverBare
	numCovers
)

// coverSpec is how a vegetation class grows. forage is what people can eat
// from the wild (fruit, tubers, shoots, greens); people can't digest grass,
// so grassland feeds grazing animals far better than people. Rates are per
// second at full moisture on soil of average fertility.
type coverSpec struct {
	forage, forageRate float32
	graze, grazeRate   float32
	drought            float32 // 1 = growth stops in dry soil, 0 = unaffected
	fert               float32 // natural fertility
	wet                float32 // how well the soil holds water (forest shade keeps it moist)
}

var covers = [numCovers]coverSpec{
	coverGrass:  {forage: 0.15, forageRate: 0.004, graze: 0.4, grazeRate: 0.008, drought: 1.0, fert: 0.6, wet: 1},
	coverForest: {forage: 1.0, forageRate: 0.024, graze: 0.15, grazeRate: 0.003, drought: 0.7, fert: 0.75, wet: 1.15},
	coverBush:   {forage: 1.5, forageRate: 0.036, graze: 0.25, grazeRate: 0.005, drought: 0.8, fert: 0.65, wet: 1},
	coverMeadow: {forage: 0.6, forageRate: 0.014, graze: 0.35, grazeRate: 0.007, drought: 0.9, fert: 0.6, wet: 1},
	// Sand by the sea: shellfish, crabs and beach plants.
	coverSand: {forage: 0.12, forageRate: 0.003, graze: 0.05, grazeRate: 0.001, drought: 0.3, fert: 0.25, wet: 0.6},
	coverBare: {forage: 0.02, forageRate: 0.0005, graze: 0.08, grazeRate: 0.0015, drought: 0.9, fert: 0.4, wet: 0.9},
}

const (
	dieBack        = 0.15   // share per second of growth above the season's cap that withers
	fallowYears    = 10     // years for exhausted soil to recover under fallow
	manureBoost    = 0.25   // fertility a pen's manure adds around it
	wildCropOdds   = 0.35   // chance foraging a wild crop's tile yields the crop itself
	fishGrowth     = 0.1    // logistic growth rate of fish, per second
	fishDrift      = 0.0005 // fish moving in from deeper water, per second
	floodSilt      = 0.05   // fertility a flood leaves on the riverbanks
	riparian       = 2      // tiles from fresh water that count as riverbank
	riverbankWater = 1.6    // riverbank soil moisture = riverbankWater × river flow − riverbankDry
	riverbankDry   = 0.35
	coastal        = 3 // tiles from the sea that count as coast
	fallowRate     = 1 / (fallowYears * SecondsPerYear)
)

// deriveLand classifies tiles: vegetation, sea or fresh water, riverbanks and coasts.
func (e *Ecology) deriveLand() {
	l := &e.land
	n := l.W * l.H
	e.cover = make([]uint8, n)
	e.sea = make([]bool, n)
	e.fresh = make([]bool, n)
	e.walkable = e.walkable[:0]
	for i := range n {
		if l.Blocked[i] {
			continue
		}
		e.walkable = append(e.walkable, int32(i))
		switch l.Objects[i] {
		case world.Bush:
			e.cover[i] = coverBush
			continue
		case world.Flowers:
			e.cover[i] = coverMeadow
			continue
		}
		switch l.Ground[i] {
		case world.Grass:
			e.cover[i] = coverGrass
		case world.ForestFloor:
			e.cover[i] = coverForest
		case world.Sand:
			e.cover[i] = coverSand
		default:
			e.cover[i] = coverBare
		}
	}
	// Rivers and lakes are fresh (world.FreshWater); the rest is the sea.
	// A Land built by hand without that map counts every lake and river
	// that doesn't reach the edge of the map as fresh.
	if l.Fresh == nil {
		e.markSeaFromEdges()
	}
	for i := range n {
		if l.Fresh != nil {
			e.sea[i] = l.Water[i] && !l.Fresh[i]
		}
		e.fresh[i] = l.Water[i] && !e.sea[i]
	}
	e.nearFresh = e.within(e.fresh, riparian)
	e.nearSea = e.within(e.sea, coastal)
	e.buildHydrology()
}

// markSeaFromEdges marks the water that reaches the edge of the map as sea.
func (e *Ecology) markSeaFromEdges() {
	l := &e.land
	var queue []int
	for i := range l.W * l.H {
		x, y := i%l.W, i/l.W
		if l.Water[i] && (x == 0 || y == 0 || x == l.W-1 || y == l.H-1) {
			e.sea[i] = true
			queue = append(queue, i)
		}
	}
	for len(queue) > 0 {
		i := queue[0]
		queue = queue[1:]
		x, y := i%l.W, i/l.W
		for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			if j, ok := l.index(x+d[0], y+d[1]); ok && l.Water[j] && !e.sea[j] {
				e.sea[j] = true
				queue = append(queue, j)
			}
		}
	}
}

// within marks tiles no more than r tiles (in any direction) from a marked one.
func (e *Ecology) within(src []bool, r int) []bool {
	l := &e.land
	out := make([]bool, len(src))
	for i, on := range src {
		if !on {
			continue
		}
		x, y := i%l.W, i/l.W
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if j, ok := l.index(x+dx, y+dy); ok {
					out[j] = true
				}
			}
		}
	}
	return out
}

func groundKey(id int) string {
	if id >= 0 && id < len(world.GroundTiles) {
		return world.GroundTiles[id].Key
	}
	return ""
}

func hasKey(keys []string, k string) bool {
	for _, v := range keys {
		if v == k {
			return true
		}
	}
	return false
}

// seedVegetation fills plants and fish to what the land holds and scatters
// wild crops.
func (e *Ecology) seedVegetation() {
	e.baseVegetation()
	l := &e.land
	for _, i32 := range e.walkable {
		i := int(i32)
		cs := covers[e.cover[i]]
		e.fert[i] = e.fertBase[i]
		e.forage[i] = cs.forage
		e.graze[i] = cs.graze
		// Wild relatives of the crops, each where it naturally grows.
		key := groundKey(l.Ground[i])
		for _, ci := range e.rng.Perm(len(e.crops)) {
			c := e.crops[ci]
			if hasKey(c.Wild, key) && (!c.Wet || e.nearFresh[i]) && (!c.Coastal || e.nearSea[i]) && e.rng.Float64() < c.WildChance {
				e.wild[i] = uint8(ci + 1)
				break
			}
		}
	}
	copy(e.fish, e.fishCap)
}

// moisture is the soil moisture of tile i: the island's soil moisture as the
// ground holds it, kept up along rivers by groundwater and on irrigated land
// by the channels, which run low in a long drought. Riverbank groundwater
// follows the rivers' flow over the past year: about 0.63 in an ordinary
// year, under 0.5 through an El Niño and near 0.2 at the end of the worst,
// when rain-fed rice on the banks withers and only yams and the irrigated
// paddies come through.
func (e *Ecology) moisture(i int) float64 {
	c := &e.clim
	m := c.Moisture * float64(covers[e.cover[i]].wet)
	if e.nearFresh[i] {
		// Banks stay damp while the river's groundwater lasts.
		m = math.Max(m, (riverbankWater*c.Slow-riverbankDry)*e.riverWetness(i))
	}
	if e.manage[i]&Irrigated != 0 {
		// A channel waters only as well as its river can give.
		m = math.Max(m, 0.9*clamp(0.4+0.9*c.Slow, 0.4, 1)*e.irrigation(i))
	}
	return clamp(m, 0, 1)
}

// grow advances plants, soil, fish and fields by dt seconds.
func (e *Ecology) grow(t, dt float64) {
	fd := float32(dt)
	for _, i32 := range e.walkable {
		i := int(i32)
		cs := &covers[e.cover[i]]
		m := float32(e.moisture(i))
		e.moistNow[i] = m
		f := (1 - cs.drought) + cs.drought*m
		fertF := 0.5 + e.fert[i]
		// In the dry season plants die back: what the land holds shrinks.
		fc := cs.forage * (0.45 + 0.55*f)
		if e.cover[i] == coverForest {
			fc *= 0.4 + 0.6*e.woodland[i] // fruit trees and shade go with the forest
		}
		gc := cs.graze * (0.25 + 0.75*f) // grass browns off in the dry season
		e.forage[i] = regrow(e.forage[i], fc, cs.forageRate*f*fertF, fd)
		e.graze[i] = regrow(e.graze[i], gc, cs.grazeRate*f*fertF, fd)

		// Fertility drifts back to what the land naturally holds (a fallow),
		// slowly under crops; cut forest loses topsoil, manure adds to it.
		target := e.fertBase[i] * (0.7 + 0.3*e.woodland[i])
		if e.manage[i]&Manured != 0 {
			target = min(1, target+manureBoost)
		}
		rate := float32(fallowRate)
		if e.plotAt[i] != 0 {
			rate /= 4
		}
		e.fert[i] += (target - e.fert[i]) * min(1, rate*fd)
	}
	// Rivers and lakes shrink in a drought and hold fewer fish (none in a dry
	// bed); the sea doesn't.
	for i, c := range e.fishCap {
		if c > 0 {
			if e.fresh[i] {
				switch e.WaterState(i) {
				case WaterFlowing:
					c *= 0.4 + 0.6*float32(e.hydro.level[i])/255
				case WaterPools:
					c *= 0.3
				default:
					c = 0
				}
			}
			f := e.fish[i]
			if f > c {
				e.fish[i] = f - (f-c)*min(1, dieBack*fd)
				continue
			}
			e.fish[i] = min(c, f+(fishGrowth*f*(1-f/c)+fishDrift)*fd)
		}
	}
	e.growPlots(t, dt)
}

func regrow(v, cap, rate, dt float32) float32 {
	if v < cap {
		return min(cap, v+rate*dt)
	}
	return v - (v-cap)*min(1, dieBack*dt)
}

// flood drowns crops on the riverbanks; the silt it leaves makes them more fertile.
func (e *Ecology) flood(severity float64) {
	sev := float32(severity * (1 + e.deforest))
	lost := 0
	for k := len(e.plots) - 1; k >= 0; k-- {
		p := e.plots[k]
		if !e.nearFresh[p.Tile] {
			continue
		}
		p.Health -= sev
		if p.Health <= 0 {
			e.removePlot(k)
			lost++
		}
	}
	for _, i32 := range e.walkable {
		if i := int(i32); e.nearFresh[i] {
			e.forage[i] *= 1 - 0.5*min(1, sev)
			e.fert[i] = min(1, e.fert[i]+floodSilt*min(1, sev))
		}
	}
	e.stats.cur.Flood = true
	e.stats.cur.CropLoss += lost
	e.plotVer++
	if lost > 0 {
		e.event("climate", fmt.Sprintf("Banjir di tepi sungai (tahun %d): %d petak tanaman hanyut; lumpurnya menyuburkan tanah", e.clim.Year+1, lost))
	} else {
		e.event("climate", fmt.Sprintf("Banjir di tepi sungai (tahun %d); lumpurnya menyuburkan tanah", e.clim.Year+1))
	}
}

// --- what the simulation reads and takes ------------------------------------

// Forage is the wild food on tile i that people can gather or eat.
func (e *Ecology) Forage(i int) float32 { return e.forage[i] }

// TakeForage removes up to amount of wild food from tile i and returns what was taken.
func (e *Ecology) TakeForage(i int, amount float32) float32 {
	took := min(e.forage[i], amount)
	e.forage[i] -= took
	return took
}

// WildCrop is the crop that grows wild on tile i, if any; foraging there
// sometimes yields the crop itself, which can then be planted.
func (e *Ecology) WildCrop(i int) (chem.ItemID, bool) {
	if w := e.wild[i]; w != 0 {
		return e.crops[w-1].Item, true
	}
	return "", false
}

// ForageFind picks what a forager brings home from tile i: now and then the
// wild crop growing there, otherwise mixed wild food.
func (e *Ecology) ForageFind(i int) chem.ItemID {
	if id, ok := e.WildCrop(i); ok && e.rng.Float64() < wildCropOdds {
		return id
	}
	return chem.Food
}

// Fish is the fish in water tile i.
func (e *Ecology) Fish(i int) float32 { return e.fish[i] }

// TakeFish catches one fish from water tile i if there is one.
func (e *Ecology) TakeFish(i int) bool {
	if e.fish[i] < 1 {
		return false
	}
	e.fish[i]--
	e.stats.cur.Fished++
	return true
}

// Fertility is the soil fertility of tile i, 0–1.
func (e *Ecology) Fertility(i int) float32 { return e.fert[i] }

// Moisture is the soil moisture of tile i now, 0–1.
func (e *Ecology) Moisture(i int) float64 { return e.moisture(i) }

// Fresh reports whether tile i is a river or lake; Sea whether it is the sea.
func (e *Ecology) Fresh(i int) bool { return e.fresh[i] }
func (e *Ecology) Sea(i int) bool   { return e.sea[i] }

// Riverbank reports whether tile i is within a couple of tiles of fresh water.
func (e *Ecology) Riverbank(i int) bool { return e.nearFresh[i] }

// FreshWithin reports whether fresh water lies within r tiles of (x, y).
func (e *Ecology) FreshWithin(x, y, r int) bool {
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if j, ok := e.land.index(x+dx, y+dy); ok && e.fresh[j] {
				return true
			}
		}
	}
	return false
}

// Production is what the land grows per second right now: wild food people
// can eat, and fish (both in food units).
func (e *Ecology) Production() (forage, fish float64) {
	for _, i32 := range e.walkable {
		i := int(i32)
		cs := &covers[e.cover[i]]
		m := e.moistNow[i]
		f := (1 - cs.drought) + cs.drought*m
		forage += float64(cs.forageRate * f * (0.5 + e.fert[i]))
	}
	for i, c := range e.fishCap {
		if c > 0 {
			f := e.fish[i]
			fish += float64(fishGrowth*f*(1-f/c) + fishDrift)
		}
	}
	return forage, fish
}

// Totals sums wild food and fish standing on the land now.
func (e *Ecology) Totals() (forage, graze, fish float64) {
	for _, i := range e.walkable {
		forage += float64(e.forage[i])
		graze += float64(e.graze[i])
	}
	for i, c := range e.fishCap {
		if c > 0 {
			fish += float64(e.fish[i])
		}
	}
	return
}
