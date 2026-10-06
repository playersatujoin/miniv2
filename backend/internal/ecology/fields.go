package ecology

import (
	"fmt"
	"math"

	"miniv2/backend/internal/chem"
)

// Plot is one tile of planted crop.
type Plot struct {
	Tile     int32       `json:"tile"`
	Crop     chem.ItemID `json:"crop"`
	Owner    int64       `json:"owner"`           // who planted it
	House    int64       `json:"house,omitempty"` // their household
	Planted  float64     `json:"planted"`         // simulated time
	Growth   float32     `json:"growth"`          // 0–1 towards the next harvest
	Health   float32     `json:"health"`          // 1 thriving, 0 dead
	Ripe     float32     `json:"ripe,omitempty"`  // seconds it has stood ripe
	Left     int16       `json:"left,omitempty"`  // units still to harvest
	Harvests int16       `json:"harvests,omitempty"`
	Skill    float32     `json:"skill"` // the planter's farming skill, 0–1

	ci    int
	stage uint8
}

const (
	ripeHold       = 0.5 * SecondsPerYear // a ripe crop waits half a year (tubers keep in the ground) …
	overripeLoss   = 1.5                  // … then loses a unit every so many seconds to rot and birds
	witherRate     = 0.4                  // health lost per second in parched soil (dead in a few months)
	perennialHardy = 0.3                  // deep-rooted palms and bananas wither slower
	recoverRate    = 0.02                 // health regained per second in good soil
	farmlandGrowth = 1.25
	farmlandYield  = 1.3
	sawahYield     = 1.6  // irrigated rice
	sawahDrain     = 0.25 // irrigated rice paddies keep their fertility
	manureYield    = 1.15
	farmlandDrain  = 0.8
	ripeCell       = 8
)

// Plot stages as streamed to the observer.
const (
	StageSeedling = iota
	StageYoung
	StageGrown
	StageRipe
)

func (e *Ecology) crop(p *Plot) *chem.Crop { return &e.crops[p.ci] }

func (e *Ecology) cropIndex(id chem.ItemID) (int, bool) {
	for i := range e.crops {
		if e.crops[i].Item == id {
			return i, true
		}
	}
	return 0, false
}

// Plantable reports whether item is a crop.
func (e *Ecology) Plantable(id chem.ItemID) bool {
	_, ok := e.cropIndex(id)
	return ok
}

// CanPlant reports whether crop item can be planted on tile i: walkable,
// free of other plants, the right ground, wet crops by fresh water or on
// irrigated land, coastal ones by the sea.
func (e *Ecology) CanPlant(i int, id chem.ItemID) bool {
	ci, ok := e.cropIndex(id)
	if !ok || i < 0 || i >= len(e.plotAt) || e.land.Blocked[i] || e.plotAt[i] != 0 {
		return false
	}
	c := &e.crops[ci]
	if !hasKey(c.Ground, groundKey(e.land.Ground[i])) {
		return false
	}
	if c.Wet && !e.nearFresh[i] && e.manage[i]&Irrigated == 0 {
		return false
	}
	return !c.Coastal || e.nearSea[i]
}

// SuitsHere rates how well crop item would grow on tile i now (0–1): soil
// moisture against what it needs, and fertility.
func (e *Ecology) SuitsHere(i int, id chem.ItemID) float64 {
	ci, ok := e.cropIndex(id)
	if !ok || !e.CanPlant(i, id) {
		return 0
	}
	c := &e.crops[ci]
	m := e.moisture(i)
	return clamp(m/c.Water, 0, 1) * (0.5 + 0.5*float64(e.fert[i]))
}

// Plant puts crop item into tile i.
func (e *Ecology) Plant(i int, id chem.ItemID, owner, house int64, skill, t float64) bool {
	if !e.CanPlant(i, id) {
		return false
	}
	ci, _ := e.cropIndex(id)
	p := &Plot{Tile: int32(i), Crop: id, Owner: owner, House: house, Planted: t, Health: 1, Skill: float32(skill), ci: ci}
	e.plots = append(e.plots, p)
	e.plotAt[i] = int32(len(e.plots))
	e.plotVer++
	e.stats.cur.Planted++
	return true
}

func (e *Ecology) removePlot(k int) {
	p := e.plots[k]
	e.plotAt[p.Tile] = 0
	last := len(e.plots) - 1
	if k != last {
		e.plots[k] = e.plots[last]
		e.plotAt[e.plots[k].Tile] = int32(k + 1)
	}
	e.plots[last] = nil
	e.plots = e.plots[:last]
	e.plotVer++
}

// PlotAt is the plot on tile i, or nil.
func (e *Ecology) PlotAt(i int) *Plot {
	if i < 0 || i >= len(e.plotAt) || e.plotAt[i] == 0 {
		return nil
	}
	return e.plots[e.plotAt[i]-1]
}

// Plots lists every plot; the slice must not be kept.
func (e *Ecology) Plots() []*Plot { return e.plots }

// PlotVersion changes whenever a plot appears, goes or looks different.
func (e *Ecology) PlotVersion() int64 { return e.plotVer }

// Stage is how a plot looks: seedling, young, grown or ripe. A palm or
// banana that has borne fruit stays a grown tree between harvests.
func (p *Plot) Stage() uint8 {
	switch {
	case p.Left > 0:
		return StageRipe
	case p.Growth >= 0.66 || p.Harvests > 0:
		return StageGrown
	case p.Growth >= 0.33:
		return StageYoung
	}
	return StageSeedling
}

// look is what the observer sees of a plot: its stage, and whether it wilts.
func (p *Plot) look() uint8 {
	l := p.Stage()
	if p.Health < 0.5 {
		l |= 8
	}
	return l
}

// IsRipe reports whether there is something to harvest.
func (p *Plot) IsRipe() bool { return p.Left > 0 }

func (e *Ecology) growPlots(t, dt float64) {
	fd := float32(dt)
	for k := len(e.plots) - 1; k >= 0; k-- {
		p := e.plots[k]
		c := e.crop(p)
		i := int(p.Tile)
		m := float64(e.moistNow[i])
		if p.Left == 0 {
			span := c.GrowYears
			if p.Harvests > 0 && c.RepeatYears > 0 {
				span = c.RepeatYears
			}
			rate := dt / (span * SecondsPerYear) * clamp(m/c.Water, 0, 1) * (0.6 + 0.4*float64(e.fert[i]))
			if e.manage[i]&Farmland != 0 {
				rate *= farmlandGrowth
			}
			p.Growth += float32(rate)
		}
		// Parched plants wither; healthy soil lets them recover.
		if dry := 0.5 * c.Water; m < dry {
			rate := witherRate
			if c.RepeatYears > 0 {
				rate *= perennialHardy
			}
			p.Health -= float32(rate * (1 - m/dry) * dt)
		} else {
			p.Health = min(1, p.Health+recoverRate*fd)
		}
		if p.Health <= 0 {
			e.stats.cur.CropLoss++
			e.stats.cur.Withered++
			e.removePlot(k)
			continue
		}
		if p.Growth >= 1 && p.Left == 0 {
			p.Growth = 1
			p.Ripe = 0
			if p.Left = int16(e.yield(p)); p.Left == 0 {
				e.stats.cur.CropLoss++
				e.removePlot(k)
				continue
			}
		}
		if p.Left > 0 {
			p.Ripe += fd
			if p.Ripe > ripeHold && e.rng.Float64() < dt/overripeLoss {
				p.Left--
				if p.Left == 0 {
					e.stats.cur.Wasted++
					if !e.nextCycle(k, t) {
						continue
					}
				}
			}
		}
		if look := p.look(); look != p.stage {
			p.stage = look
			e.plotVer++
		}
	}
	e.indexRipe()
}

// yield is what a ripe plot gives: its crop's yield, scaled by the plants'
// health, the soil, the planter's skill and how the field is kept.
func (e *Ecology) yield(p *Plot) int {
	c := e.crop(p)
	i := int(p.Tile)
	y := float64(c.Yield) * float64(p.Health) * (0.4 + 0.6*float64(e.fert[i])) * (0.6 + 0.8*float64(p.Skill))
	mg := e.manage[i]
	if mg&Farmland != 0 {
		y *= farmlandYield
	}
	if mg&Irrigated != 0 && c.Item == "padi" {
		y *= sawahYield
	}
	if mg&Manured != 0 {
		y *= manureYield
	}
	n := int(y)
	if e.rng.Float64() < y-float64(n) {
		n++
	}
	if n == 0 && p.Health > 0.3 {
		n = 1
	}
	return n
}

// nextCycle ends a plot's harvest: an annual is done, a perennial starts its
// next fruiting unless it is too old. It reports whether the plot remains.
func (e *Ecology) nextCycle(k int, t float64) bool {
	p := e.plots[k]
	c := e.crop(p)
	if c.RepeatYears == 0 || t-p.Planted > c.LifeYears*SecondsPerYear {
		e.removePlot(k)
		return false
	}
	p.Growth, p.Ripe = 0, 0
	p.Harvests++
	return true
}

// Harvest takes up to max units from the ripe plot on tile i at time t. It
// returns the crop and how many units were taken. The soil gives up some
// fertility with every completed harvest.
func (e *Ecology) Harvest(i, max int, t float64) (chem.ItemID, int) {
	p := e.PlotAt(i)
	if p == nil || p.Left == 0 || max <= 0 {
		return "", 0
	}
	n := min(int(p.Left), max)
	p.Left -= int16(n)
	id := p.Crop
	e.stats.cur.Harvest += n
	if p.Left == 0 {
		c := e.crop(p)
		drain := c.Drain
		mg := e.manage[i]
		if mg&Irrigated != 0 && c.Item == "padi" {
			drain *= sawahDrain
		}
		if mg&Farmland != 0 {
			drain *= farmlandDrain
		}
		e.fert[i] = max32(0.05, e.fert[i]-float32(drain))
		e.nextCycle(int(e.plotAt[i]-1), t)
	}
	e.plotVer++
	e.indexRipe()
	return id, n
}

func max32(a, b float32) float32 { return max(a, b) }

// Trample lets an animal eat from or trample the plot on tile i.
func (e *Ecology) trample(i int, damage float32) bool {
	p := e.PlotAt(i)
	if p == nil {
		return false
	}
	p.Health -= damage
	if p.Health <= 0 {
		e.stats.cur.CropLoss++
		e.removePlot(int(e.plotAt[i] - 1))
	}
	return true
}

// indexRipe buckets ripe plots by 8×8 cell for quick "ripe crops nearby" queries.
func (e *Ecology) indexRipe() {
	cw, ch := (e.land.W+ripeCell-1)/ripeCell, (e.land.H+ripeCell-1)/ripeCell
	if len(e.ripeCells) != cw*ch {
		e.ripeCells = make([][]int32, cw*ch)
	}
	for k := range e.ripeCells {
		e.ripeCells[k] = e.ripeCells[k][:0]
	}
	for _, p := range e.plots {
		if p.Left > 0 {
			x, y := int(p.Tile)%e.land.W, int(p.Tile)/e.land.W
			c := (y/ripeCell)*cw + x/ripeCell
			e.ripeCells[c] = append(e.ripeCells[c], p.Tile)
		}
	}
}

// RipeNear finds the closest ripe plot within r tiles of (x, y) that may
// returns true for. It returns the plot's tile and distance.
func (e *Ecology) RipeNear(x, y, r float64, may func(*Plot) bool) (tile int, dist float64, ok bool) {
	if len(e.ripeCells) == 0 {
		e.indexRipe()
	}
	cw := (e.land.W + ripeCell - 1) / ripeCell
	ch := (e.land.H + ripeCell - 1) / ripeCell
	x0, x1 := max(0, int((x-r)/ripeCell)), min(cw-1, int((x+r)/ripeCell))
	y0, y1 := max(0, int((y-r)/ripeCell)), min(ch-1, int((y+r)/ripeCell))
	dist = r
	for cy := y0; cy <= y1; cy++ {
		for cx := x0; cx <= x1; cx++ {
			for _, t := range e.ripeCells[cy*cw+cx] {
				// The index is rebuilt at each growth step; a plot trampled
				// or flooded since then is skipped.
				p := e.PlotAt(int(t))
				if p == nil || !p.IsRipe() {
					continue
				}
				px, py := float64(int(t)%e.land.W)+0.5, float64(int(t)/e.land.W)+0.5
				d := math.Hypot(px-x, py-y)
				if d <= dist && may(p) {
					tile, dist, ok = int(t), d, true
				}
			}
		}
	}
	return
}

// CropName is the display name of a crop item.
func (e *Ecology) CropName(id chem.ItemID) string {
	if ci, ok := e.cropIndex(id); ok {
		return e.crops[ci].Name
	}
	return string(id)
}

// CropIndex is a crop's position in chem.Crops(), for compact streaming.
func (e *Ecology) CropIndex(id chem.ItemID) int {
	ci, _ := e.cropIndex(id)
	return ci
}

func (e *Ecology) describeHarvestFailure(n int) string {
	return fmt.Sprintf("Gagal panen (tahun %d): %d petak tanaman mati kekeringan", e.clim.Year+1, n)
}
