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
