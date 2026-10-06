package ecology

import (
	"fmt"
	"math"
)

// The monsoon year. Indonesia has a wet season from about October to April
// and a dry season from about April to October; rain peaks around January.
// Every year in May the ocean decides the coming twelve months: an El Niño
// makes the dry season long and parched (as in 1982, 1997 and 2015), a La
// Niña makes the year wetter, with floods. They recur irregularly, every two
// to seven years (BMKG; McBride, Haylock & Nicholls 2003).
const (
	rainPeak     = 0.04      // fraction of the year when rain peaks (mid January)
	rainSwing    = 0.85      // seasonal amplitude around the yearly mean
	ensoAt       = 4.0 / 12  // May: the ENSO state for the next 12 months is set
	floodAt      = 1.0 / 12  // February: the height of the floods
	lateRainsAt  = 10.0 / 12 // El Niño: the dry season runs May–October …
	lateRainsEnd = 2.0 / 12  // … and the rains stay thin until about March
	moistTau     = 0.6       // seconds (≈ a month) for the soil to follow the rain
	slowTau      = 8.0       // seconds (a year) for rivers and groundwater
	moistFull    = 1.45      // rain at which the soil is saturated
	dayLength    = 2.0       // seconds of the abstract day–night rhythm
	elNinoDry    = 0.3       // rain multiplier in an El Niño dry season
	elNinoLate   = 0.4       // … in the months the rains should have come
	elNinoWet    = 0.8       // … and in the rest of the wet season
	laNinaRain   = 1.35      // rain multiplier in a La Niña year
	floodLaNina  = 0.6       // chance of a flood in a La Niña February
	floodNeutral = 0.08      // … in an ordinary one
)

// ENSO states.
const (
	LaNina  = -1
	Neutral = 0
	ElNino  = 1
)

// Climate is the weather of the whole island.
type Climate struct {
	Off      bool    `json:"off,omitempty"`
	ENSO     int     `json:"enso"`     // for the twelve months from the last May
	ENSOYear int     `json:"ensoYear"` // the year (0-based) of that May
	Moisture float64 `json:"moisture"` // soil moisture 0–1
	Slow     float64 `json:"slow"`     // about a year's mean moisture: rivers, groundwater
	Rain     float64 `json:"rain"`     // rainfall now, 1 = the yearly average
	Year     int     `json:"year"`     // current year, 0-based
	Phase    float64 `json:"phase"`    // fraction of the year, 0 = 1 January
	Light    float64 `json:"light"`
	RainSum  float64 `json:"rainSum"`  // rain×seconds so far this year
	LastRain float64 `json:"lastRain"` // last year's mean rain, 1 = average
	Floods   int     `json:"floods"`   // last year checked for floods

	// Wind and storms (see stepWind). Zero in older saves: a calm to start
	// with, which the wind leaves within a few ticks.
	WindDir    float64 `json:"windDir,omitempty"`    // radians, where the wind blows towards (0 = +x, east)
	WindSpeed  float64 `json:"wind,omitempty"`       // 0–1; 1 ≈ 25 m/s, a storm
	WindAim    float64 `json:"windAim,omitempty"`    // the direction it is turning towards
	WindLevel  float64 `json:"windLevel,omitempty"`  // where the wind sits between the weather's least and most, 0–1
	WindTarget float64 `json:"windTarget,omitempty"` // … and the level it drifts towards
	Gust       float64 `json:"gust,omitempty"`       // a gust on top of the wind, fading fast
	Storm      float64 `json:"storm,omitempty"`      // how hard a storm blows now, 0–1 (0 = no storm)
	StormPeak  float64 `json:"stormPeak,omitempty"`  // the storm's strength at its height
	StormLen   float64 `json:"stormLen,omitempty"`   // seconds it lasts in all …
	StormLeft  float64 `json:"stormLeft,omitempty"`  // … and still to go
	StormDry   bool    `json:"stormDry,omitempty"`   // a dry thunderstorm: wind and lightning, little rain
	StormFires int     `json:"stormFires,omitempty"` // fires its lightning has started
	Storms     int     `json:"storms,omitempty"`     // storms so far
	// Seed makes this world's weather its own without drawing from the
	// shared random generator (see weatherSeed).
	Seed uint64 `json:"seed,omitempty"`
}

func yearAndPhase(t float64) (int, float64) {
	y := math.Floor(t / SecondsPerYear)
	return int(y), t/SecondsPerYear - y
}

func (c *Climate) init(t float64, off bool) {
	c.Off = off
	c.Year, c.Phase = yearAndPhase(t)
	c.ENSOYear = c.Year
	if c.Phase < ensoAt {
		c.ENSOYear--
	}
	c.Floods = c.Year
	if c.Phase < floodAt {
		c.Floods--
	}
	c.Rain = c.rainAt(c.Phase)
	c.Moisture = c.target()
	c.Slow = 1 / moistFull
	c.LastRain = 1
	c.Light = lightAt(t)
	c.initWind()
}

// rainAt is the rainfall at a point in the year, 1 = the yearly average.
func (c *Climate) rainAt(phase float64) float64 {
	if c.Off {
		return 1
	}
	r := 1 + rainSwing*math.Cos(2*math.Pi*(phase-rainPeak))
	switch c.ENSO {
	case ElNino:
		// A parched dry season, then the rains come two or three months late.
		switch {
		case phase >= ensoAt && phase < lateRainsAt:
			r *= elNinoDry
		case phase >= lateRainsAt || phase < lateRainsEnd:
			r *= elNinoLate
		default:
			r *= elNinoWet
		}
	case LaNina:
		r *= laNinaRain
	}
	return r
}

func (c *Climate) target() float64 { return clamp(c.Rain/moistFull, 0, 1) }

func lightAt(t float64) float64 {
	return clamp(0.5+0.7*math.Sin(2*math.Pi*t/dayLength), 0, 1)
}

func (c *Climate) step(e *Ecology, t, dt float64) {
	year, phase := yearAndPhase(t)
	if year != c.Year {
		c.LastRain = c.RainSum / SecondsPerYear
		e.stats.endYear(e, c.Year, c.ENSO, c.LastRain)
		c.RainSum = 0
		c.Year = year
	}
	c.Phase = phase
	if !c.Off && (year > c.ENSOYear+1 || year == c.ENSOYear+1 && phase >= ensoAt) {
		c.ENSOYear = year
		if phase < ensoAt {
			c.ENSOYear--
		}
		c.nextENSO(e)
	}
	if !c.Off && (year > c.Floods+1 || year == c.Floods+1 && phase >= floodAt) {
		c.Floods = year
		if phase < floodAt {
			c.Floods--
		}
		chance := floodNeutral
		switch c.ENSO {
		case LaNina:
			chance = floodLaNina
		case ElNino:
			chance = 0
		}
		if e.rng.Float64() < chance {
			e.flood(0.3 + 0.7*e.rng.Float64())
		}
	}
	c.stepWind(e, t, dt)
	c.Rain = c.rainAt(phase)
	c.RainSum += c.Rain * dt
	c.Moisture += (c.target() - c.Moisture) * math.Min(1, dt/moistTau)
	c.Slow += (c.Moisture - c.Slow) * math.Min(1, dt/slowTau)
	c.Light = lightAt(t)
}

// nextENSO draws the coming twelve months. An El Niño is often followed by
// a La Niña, and La Niñas tend to last more than one year; neither repeats
// as an El Niño does. Over many years about one year in five is an El Niño.
func (c *Climate) nextENSO(e *Ecology) {
	pNino, pNina := 0.25, 0.15
	switch c.ENSO {
	case ElNino:
		pNino, pNina = 0.05, 0.35
	case LaNina:
		pNino, pNina = 0.05, 0.45
	}
	prev := c.ENSO
	switch r := e.rng.Float64(); {
	case r < pNino:
		c.ENSO = ElNino
	case r < pNino+pNina:
		c.ENSO = LaNina
	default:
		c.ENSO = Neutral
	}
	if c.ENSO == prev {
		return
	}
	switch c.ENSO {
	case ElNino:
		e.event("climate", fmt.Sprintf("El Niño datang (tahun %d): kemarau akan panjang dan kering", c.Year+1))
	case LaNina:
		e.event("climate", fmt.Sprintf("La Niña (tahun %d): hujan lebih lebat dari biasa, waspada banjir", c.Year+1))
	default:
		e.event("climate", fmt.Sprintf("Iklim kembali normal (tahun %d)", c.Year+1))
	}
}

// Season is +1 at the height of the wet season and −1 at the height of the
// dry season, following the calendar rather than this year's actual rain.
func (e *Ecology) Season() float64 {
	if e.clim.Off {
		return 0
	}
	return math.Cos(2 * math.Pi * (e.clim.Phase - rainPeak))
}

// Wet reports whether it is the wet season.
func (e *Ecology) Wet() bool { return e.Season() >= 0 }

// Light is daylight in the abstract day–night rhythm, 0 (night) to 1.
func (e *Ecology) Light() float64 { return e.clim.Light }

// Climate returns the current weather.
func (e *Ecology) Climate() Climate { return e.clim }

// ENSOName is the observer-facing name of an ENSO state.
func ENSOName(s int) string {
	switch s {
	case ElNino:
		return "el_nino"
	case LaNina:
		return "la_nina"
	}
	return "netral"
}

// --- wind and storms ----------------------------------------------------------
//
// The wind follows RAGE's CWind and CWeather (game/wind.h, game/weather.h).
// Its direction turns at a limited rate towards a target that is redrawn now
// and then; its strength sits between the weather's least and most wind at a
// level that drifts towards a target redrawn at random; and calm weather
// blends into a storm as CWeather blends one weather type into the next, so
// a storm brings gale-force wind, downpours (see Downpour) and lightning. Gusts come and go
// on top. Every chance is hashed from the time and the world's weather seed,
// so the weather never shifts the random draws of anything else.
//
// Directions are in map coordinates with x to the east and y to the south
// (north is up on the map): 0 blows towards the east, π/2 towards the south.
const (
	// The monsoons (BMKG; Aldrian & Susanto 2003): in the wet season the
	// west monsoon blows from the west-north-west across Java and Nusa
	// Tenggara towards the east-south-east; in the dry season the east
	// monsoon blows from Australia, from the south-east towards the
	// north-west. At the turn of the monsoons (the pancaroba, about April
	// and October) the wind is light and comes from anywhere.
	westMonsoon = math.Pi / 8      // towards the east-south-east
	eastMonsoon = -3 * math.Pi / 4 // towards the north-west
	// Strength, 1 ≈ 25 m/s. Surface winds over the islands are mostly light,
	// 2–6 m/s, a little stronger and steadier at a monsoon's height; squalls
	// and storms bring 12–25 m/s.
	calmMin      = 0.04
	calmMax      = 0.26
	monsoonWind  = 0.1 // added to the most wind at a monsoon's height
	stormWindMin = 0.5
	stormWindMax = 1.0
	windAverage  = 0.15 // NoClimate: a steady breeze from the west
	// How the wind wanders (estimates; a simulated second is about six
	// weeks, so this is the run of the weather, not a single day's breeze).
	windLevelStep = 0.8 // the level moves this much per second towards its target …
	windRedraw    = 1.5 // … and once there draws a new one at this rate per second
	windTurn      = 2.0 // radians per second the wind turns …
	windAimRedraw = 1.0 // … and the rate per second at which it takes a new heading
	windWander    = 0.8 // at a monsoon's height it strays up to (1 − this)·π from it
	gustTau       = 0.15
	gustRate      = 4.0 // gusts per second in a full wind
	gustiness     = 0.4 // a gust adds up to this share of the wind
	// Storms: thunderstorms with squalls come with the rains, most of all at
	// the turn of the monsoons, when BMKG counts the most thunderstorm days;
	// La Niña's heavier rains bring more of them and El Niño fewer. A storm
	// here is a stormy spell, from a week or two to a couple of months.
	stormRate     = 0.12 // storms per second at the wet season's usual rain
	stormTurnRate = 0.12 // more at the turn of the monsoons
	stormShortest = 0.25 // seconds
	stormSpread   = 0.75 // seconds a storm may last beyond the shortest
	stormRamp     = 0.12 // seconds to build up and to die down
	stormRainWet  = 2.0  // a storm's downpour at its height, on top of the season's rain …
	stormRainDry  = 0.15 // … but a dry thunderstorm brings little
	// Dry thunderstorms, with lightning but little rain reaching the ground,
	// come over parched land: the build-up storms at the end of a long dry
	// season are what lightning fires start from.
	dryStormMax = 0.65
	stormShows  = 0.2  // strength from which the observer sees a storm
	strikeRate  = 40.0 // cloud-to-ground strikes per second over the map in a full storm (an estimate: a stormy spell of weeks over the island)
)

// Salts for the weather's hashed chances.
const (
	saltAim uint64 = 0x5717 + iota
	saltAimTo
	saltLevel
	saltLevelTo
	saltGust
	saltGustSize
	saltStorm
	saltStormLen
	saltStormPeak
	saltStormDry
	saltStrike
	saltStrikeX
	saltStrikeY
	saltStrikeCatch
)

// roll is a fixed number in [0, 1) for a moment and a salt, different in
// every world.
func (c *Climate) roll(t float64, salt uint64) float64 {
	return hash01(int(math.Float64bits(t)), int(c.Seed), salt)
}

// prevailing is the monsoon's direction now and how steady it is: 1 at the
// height of either monsoon, 0 at their turn.
func (c *Climate) prevailing() (dir, steady float64) {
	s := math.Cos(2 * math.Pi * (c.Phase - rainPeak))
	if s >= 0 {
		return westMonsoon, s
	}
	return eastMonsoon, -s
}

// windRange is the least and most wind of the weather now: calm weather
// blended into a storm by how hard the storm blows.
func (c *Climate) windRange(steady float64) (lo, hi float64) {
	lo = calmMin + (stormWindMin-calmMin)*c.Storm
	most := calmMax + monsoonWind*steady
	hi = most + (stormWindMax-most)*c.Storm
	return lo, hi
}

func (c *Climate) initWind() {
	if c.Off {
		c.WindDir, c.WindSpeed = 0, windAverage
		return
	}
	dir, steady := c.prevailing()
	c.WindDir, c.WindAim = dir, dir
	c.WindLevel, c.WindTarget = 0.5, 0.5
	lo, hi := c.windRange(steady)
	c.WindSpeed = lo + (hi-lo)*c.WindLevel
}

// stepWind moves the wind and the storms on by dt seconds to time t.
func (c *Climate) stepWind(e *Ecology, t, dt float64) {
	if c.Off {
		c.WindDir, c.WindAim, c.WindSpeed, c.Gust = 0, 0, windAverage, 0
		c.Storm, c.StormLeft = 0, 0
		return
	}
	c.stepStorm(e, t, dt)
	dir, steady := c.prevailing()
	if math.Abs(wrapAngle(c.WindAim-c.WindDir)) < 1e-3 && c.roll(t, saltAim) < windAimRedraw*dt {
		c.WindAim = wrapAngle(dir + (2*c.roll(t, saltAimTo)-1)*math.Pi*(1-windWander*steady))
	}
	turn := windTurn * dt
	c.WindDir = wrapAngle(c.WindDir + clamp(wrapAngle(c.WindAim-c.WindDir), -turn, turn))
	if math.Abs(c.WindTarget-c.WindLevel) < 1e-3 && c.roll(t, saltLevel) < windRedraw*dt {
		c.WindTarget = c.roll(t, saltLevelTo)
	}
	step := windLevelStep * dt
	c.WindLevel += clamp(c.WindTarget-c.WindLevel, -step, step)
	lo, hi := c.windRange(steady)
	base := lo + (hi-lo)*c.WindLevel
	c.Gust *= math.Exp(-dt / gustTau)
	if c.roll(t, saltGust) < gustRate*base*dt {
		c.Gust = math.Max(c.Gust, gustiness*base*c.roll(t, saltGustSize))
	}
	c.WindSpeed = clamp(base+c.Gust, 0, 1)
}

// stormChance is how many storms a second break at this time of year.
func (c *Climate) stormChance() float64 {
	s := math.Cos(2 * math.Pi * (c.Phase - rainPeak))
	turn := math.Max(0, 1-math.Abs(s)/0.5)
	return stormRate*clamp(c.rainAt(c.Phase)/(1+rainSwing), 0, 1.5) + stormTurnRate*turn
}

// stepStorm lets a storm build, rage and die down, or a new one break.
func (c *Climate) stepStorm(e *Ecology, t, dt float64) {
	if c.StormLeft > 0 {
		c.StormLeft = math.Max(0, c.StormLeft-dt)
		since := c.StormLen - c.StormLeft
		c.Storm = c.StormPeak * clamp(math.Min(since, c.StormLeft)/stormRamp, 0, 1)
		e.lightning(t, dt)
		return
	}
	c.Storm = 0
	if c.roll(t, saltStorm) >= c.stormChance()*dt {
		return
	}
	c.StormLen = stormShortest + stormSpread*c.roll(t, saltStormLen)
	c.StormLeft = c.StormLen
	c.StormPeak = 0.4 + 0.6*c.roll(t, saltStormPeak)
	c.StormDry = c.roll(t, saltStormDry) < clamp(0.75-1.1*c.Moisture, 0, dryStormMax)
	c.StormFires = 0
	c.Storms++
	if c.StormPeak >= 0.85 {
		if c.StormDry {
			e.event("climate", fmt.Sprintf("Badai petir kering (tahun %d): angin kencang dan petir, hampir tanpa hujan", c.Year+1))
		} else {
			e.event("climate", fmt.Sprintf("Badai (tahun %d): angin kencang dan hujan lebat", c.Year+1))
		}
	}
}

// Downpour is how hard it rains right now (1 = the yearly average): the
// season's rain, and on top of it a storm's cloudburst (a dry thunderstorm
// brings little). Rain reaches the soil, the rivers and the breeding pools at
// the season's rate, which already counts its storms; a storm only gathers
// that rain into a few heavy downpours, and those are what put out a fire or
// keep it from catching.
func (c *Climate) Downpour() float64 {
	if c.StormDry {
		return c.Rain + stormRainDry*c.Storm
	}
	return c.Rain + stormRainWet*c.Storm
}

// Wind is where the wind blows towards (radians: 0 east, π/2 south), how
// hard (0 still – 1 a storm's 25 m/s), and whether a storm is raging.
func (e *Ecology) Wind() (dir, speed float64, storm bool) {
	c := &e.clim
	return c.WindDir, c.WindSpeed, c.Storm >= stormShows
}

// wrapAngle brings an angle into [−π, π).
func wrapAngle(a float64) float64 {
	a = math.Mod(a+math.Pi, 2*math.Pi)
	if a < 0 {
		a += 2 * math.Pi
	}
	return a - math.Pi
}
