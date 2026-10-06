package sim

import (
	"maps"
	"math"
	"slices"
	"strconv"
)

// Demography keeps what demographers need for a period life table: person-
// years lived, deaths and births by age band, over a rolling window of the
// most recent simulated years. Rates are computed from those on demand.

const (
	statsWindowYears  = 50
	demographyEvery   = 10 * SecondsPerYear * TicksPerSecond // a history point every 10 years
	demographyHistMax = 400
	minExposure       = 200.0 // person-years before rates are reported
	minEvents         = 5     // births or intervals before means are reported
	minAdultDeaths    = 20
	adultDeathBins    = 48 // 2-year bins of age at death from 15 to 111
)

// Life-table age bands: 0, 1–4, 5–9, …, 75–79, 80+.
var bandStart = [...]float64{0, 1, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80}

const (
	numBands  = len(bandStart)
	band15    = 4  // first band of 15–19
	band45    = 10 // last fertile band, 45–49
	openBand  = numBands - 1
	pyramidTo = 80 // the pyramid's last band is 80+
)

func bandOf(age float64) int {
	for b := numBands - 1; b > 0; b-- {
		if age >= bandStart[b] {
			return b
		}
	}
	return 0
}

func bandWidth(b int) float64 {
	if b >= openBand {
		return math.Inf(1)
	}
	return bandStart[b+1] - bandStart[b]
}

// yearStats is one simulated year of observations. Index 0/1 of the sex
// dimension is Female/Male.
type yearStats struct {
	Year           int64                `json:"year"`
	Exposure       [2][numBands]float64 `json:"exposure"` // person-years
	Deaths         [2][numBands]int     `json:"deaths"`
	Causes         DeathCounts          `json:"causes"`
	Births         [numBands]int        `json:"births"` // children, by the mother's age band
	FirstBirthAges float64              `json:"firstBirthAges"`
	FirstBirths    int                  `json:"firstBirths"`
	Intervals      float64              `json:"intervals"` // years between a mother's deliveries
	IntervalCount  int                  `json:"intervalCount"`
	AdultDeathAges [adultDeathBins]int  `json:"adultDeathAges"`
	Crimes         int                  `json:"crimes,omitempty"`
	Kindness       int                  `json:"kindness,omitempty"`
	// Fase 3b: deaths of children under five by cause, and new bouts of
	// each disease.
	Under5 DeathCounts      `json:"under5"`
	Cases  [numDiseases]int `json:"cases"`
}

type demography struct {
	Years   []*yearStats      `json:"years"` // oldest first
	History []DemographyPoint `json:"history"`
}

func newDemography() *demography { return &demography{} }

func (d *demography) current(s *Sim) *yearStats {
	year := int64(s.time() / SecondsPerYear)
	if n := len(d.Years); n > 0 && d.Years[n-1].Year == year {
		return d.Years[n-1]
	}
	d.Years = append(d.Years, &yearStats{Year: year})
	if over := len(d.Years) - (statsWindowYears + 1); over > 0 {
		d.Years = slices.Delete(d.Years, 0, over)
	}
	return d.Years[len(d.Years)-1]
}

func (s *Sim) ageYears(c *Creature) float64 { return s.age(c) / SecondsPerYear }

// expose adds one simulated second of life for everyone alive.
func (d *demography) expose(s *Sim) {
	y := d.current(s)
	for _, c := range s.creatures {
		y.Exposure[c.Sex][bandOf(s.ageYears(c))] += 1 / SecondsPerYear
	}
}

// delivery records a mother giving birth to n children (before her child
// count is updated).
func (d *demography) delivery(s *Sim, mother *Creature, n int) {
	y := d.current(s)
	age := s.ageYears(mother)
	y.Births[bandOf(age)] += n
	if mother.Children == 0 {
		y.FirstBirthAges += age
		y.FirstBirths++
	}
	if mother.LastBirth != 0 {
		y.Intervals += float64(s.tick-mother.LastBirth) * dt / SecondsPerYear
		y.IntervalCount++
	}
	mother.LastBirth = s.tick
}

func (d *demography) death(s *Sim, c *Creature, cause string) {
	y := d.current(s)
	age := s.ageYears(c)
	y.Deaths[c.Sex][bandOf(age)]++
	y.Causes.add(cause)
	if age < 5 {
		y.Under5.add(cause)
	}
	if age >= 15 {
		y.AdultDeathAges[min(int((age-15)/2), adultDeathBins-1)]++
	}
}

func (d *demography) sample(s *Sim) {
	m := d.metrics(s)
	d.History = append(d.History, DemographyPoint{
		Year:           s.year(),
		Population:     len(s.creatures),
		LifeExpectancy: m.LifeExpectancy,
		SurvivalTo15:   m.SurvivalTo15,
		TFR:            m.TFR,
		Gini:           m.Gini,
		HomicideRate:   m.HomicideRate,
	})
	if over := len(d.History) - demographyHistMax; over > 0 {
		d.History = slices.Delete(d.History, 0, over)
	}
}

func (s *Sim) year() int { return int(s.time()/SecondsPerYear) + 1 }

// --- Period life table -------------------------------------------------------

type lifeTableResult struct {
	e0, e15, l5, l15, q0 float64
}

// lifeTable builds an abridged period life table from person-years and deaths
// per band (both sexes together). Death rates m become probabilities with
// q = n·m / (1 + (n−a)·m), where a is the average years lived in the band by
// those who die in it (half the band, but 0.3 for infants who mostly die
// early). The open last band lives 1/m years. If nobody was observed at some
// age, nobody survives past it. ok is false without any infant exposure.
func lifeTable(exposure, deaths [numBands]float64) (lifeTableResult, bool) {
	if exposure[0] <= 0 {
		return lifeTableResult{}, false
	}
	var l [numBands + 1]float64
	var L [numBands]float64
	l[0] = 1
	var r lifeTableResult
	for b := range numBands {
		if l[b] == 0 {
			break
		}
		n := bandWidth(b)
		if exposure[b] <= 0 {
			// Nobody this old was seen in the window: assume they die within it.
			L[b] = l[b] * math.Min(n, 5) / 2
			break
		}
		m := deaths[b] / exposure[b]
		if b == openBand {
			if m <= 0 {
				m = 0.1 // no deaths seen yet at this age: assume about 10 more years
			}
			L[b] = l[b] / m
			break
		}
		a := n / 2
		if b == 0 {
			a = 0.3
		}
		q := math.Min(1, n*m/(1+(n-a)*m))
		if b == 0 {
			r.q0 = q
		}
		l[b+1] = l[b] * (1 - q)
		L[b] = n*l[b+1] + a*(l[b]-l[b+1])
	}
	var T [numBands + 1]float64
	for b := numBands - 1; b >= 0; b-- {
		T[b] = T[b+1] + L[b]
	}
	r.e0 = T[0]
	r.l5 = l[2] // bands 0, 1–4, then 5
	r.l15 = l[band15]
	if l[band15] > 0 {
		r.e15 = T[band15] / l[band15]
	}
	return r, true
}

// tfr is the total fertility rate: births per woman-year in each 5-year band
// from 15 to 49, summed and multiplied by the band width.
func tfr(births [numBands]float64, femaleExposure [numBands]float64) (float64, bool) {
	total, exposed := 0.0, 0.0
	for b := band15; b <= band45; b++ {
		exposed += femaleExposure[b]
		if femaleExposure[b] > 0 {
			total += births[b] / femaleExposure[b] * bandWidth(b)
		}
	}
	return total, exposed >= minExposure/4
}

// gini is the Gini coefficient of non-negative values (0 = equal, 1 = one has all).
func gini(values []float64) float64 {
	v := slices.Clone(values)
	slices.Sort(v)
	sum, weighted := 0.0, 0.0
	for i, x := range v {
		sum += x
		weighted += float64(2*(i+1)-len(v)-1) * x
	}
	if sum == 0 || len(v) < 2 {
		return 0
	}
	return weighted / (float64(len(v)) * sum)
}

// --- Views ---------------------------------------------------------------

type AgeBand struct {
	Label  string `json:"label"`
	From   int    `json:"from"`
	To     *int   `json:"to"`
	Female int    `json:"female"`
	Male   int    `json:"male"`
}

type MetricRef struct {
	Low    float64 `json:"low"`
	High   float64 `json:"high"`
	Source string  `json:"source"`
	Note   string  `json:"note,omitempty"`
}

// DemographyMetrics matches DemographyMetrics in frontend/src/sim/protocol.ts;
// nil pointers are reported as null (not enough data yet).
type DemographyMetrics struct {
	WindowYears        int         `json:"windowYears"`
	PersonYears        float64     `json:"personYears"`
	LifeExpectancy     *float64    `json:"lifeExpectancy"`
	LifeExpectancy15   *float64    `json:"lifeExpectancy15"`
	SurvivalTo15       *float64    `json:"survivalTo15"`
	InfantMortality    *float64    `json:"infantMortality"`
	ModalAgeAdultDeath *float64    `json:"modalAgeAdultDeath"`
	TFR                *float64    `json:"tfr"`
	MeanAgeFirstBirth  *float64    `json:"meanAgeFirstBirth"`
	MeanBirthInterval  *float64    `json:"meanBirthInterval"`
	SexRatio           *float64    `json:"sexRatio"`
	HouseholdSize      *float64    `json:"householdSize"`
	HomicideRate       *float64    `json:"homicideRate"`
	Gini               *float64    `json:"gini"`
	DeathsByCause      DeathCounts `json:"deathsByCause"`
	// Fase 3b: the chance of dying before five (5q0), deaths under five by
	// cause, and new bouts of each disease per person-year.
	Under5Mortality *float64              `json:"under5Mortality"`
	Under5ByCause   DeathCounts           `json:"under5ByCause"`
	Incidence       [numDiseases]*float64 `json:"incidence"`
}

type DemographyPoint struct {
	Year           int      `json:"year"`
	Population     int      `json:"population"`
	LifeExpectancy *float64 `json:"lifeExpectancy"`
	SurvivalTo15   *float64 `json:"survivalTo15"`
	TFR            *float64 `json:"tfr"`
	Gini           *float64 `json:"gini"`
	HomicideRate   *float64 `json:"homicideRate"`
}

type Demography struct {
	SecondsPerYear float64              `json:"secondsPerYear"`
	Year           int                  `json:"year"`
	Pyramid        []AgeBand            `json:"pyramid"`
	Current        DemographyMetrics    `json:"current"`
	History        []DemographyPoint    `json:"history"`
	Reference      map[string]MetricRef `json:"reference"`
	// Genetics (Fase 3c): inbreeding by generation and over time.
	Genetics GeneticsInfo `json:"genetics"`
}

// Reference ranges for pre-modern small-scale societies; see
// docs/reference-demography.md for the sources and how to read them.
var demographyReference = map[string]MetricRef{
	"lifeExpectancy": {Low: 21, High: 37, Source: "Gurven & Kaplan (2007), pemburu-peramu"},
	"survivalTo15": {Low: 0.44, High: 0.73, Source: "Gurven & Kaplan (2007), Tabel 2–3",
		Note: "rata-rata 0,57"},
	"lifeExpectancy15": {Low: 28, High: 43, Source: "Gurven & Kaplan (2007), Tabel 3"},
	"infantMortality": {Low: 0.13, High: 0.41, Source: "Volk & Atkinson (2013), 20 populasi pemburu-peramu",
		Note: "rata-rata 0,27 (SD 0,07); rentang ±2 SD"},
	"modalAgeAdultDeath": {Low: 68, High: 78, Source: "Gurven & Kaplan (2007), Tabel 4"},
	"tfr": {Low: 5, High: 7, Source: "Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442",
		Note: "rata-rata 6,2; sumber sekunder"},
	"meanBirthInterval": {Low: 2.8, High: 3.3, Source: "Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442",
		Note: "rata-rata 3,1 tahun"},
	"meanAgeFirstBirth": {Low: 18, High: 20, Source: "Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018)"},
	"gini": {Low: 0.21, High: 0.29, Source: "Borgerhoff Mulder dkk. (2009), Tabel 2",
		Note: "pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material"},
}

func ptr(v float64, decimals int) *float64 {
	p := math.Pow(10, float64(decimals))
	r := math.Round(v*p) / p
	return &r
}

// metrics computes the indicators over the rolling window.
func (d *demography) metrics(s *Sim) DemographyMetrics {
	var exposure, deaths, births, female [numBands]float64
	var causes, under5 DeathCounts
	var cases [numDiseases]int
	var adultAges [adultDeathBins]int
	firstAges, intervals := 0.0, 0.0
	firsts, intervalN := 0, 0
	total := 0.0
	for _, y := range d.Years {
		for b := range numBands {
			for sex := range 2 {
				exposure[b] += y.Exposure[sex][b]
				deaths[b] += float64(y.Deaths[sex][b])
				total += y.Exposure[sex][b]
			}
			female[b] += y.Exposure[Female][b]
			births[b] += float64(y.Births[b])
		}
		causes = causes.plus(y.Causes)
		under5 = under5.plus(y.Under5)
		for d, n := range y.Cases {
			cases[d] += n
		}
		firstAges += y.FirstBirthAges
		firsts += y.FirstBirths
		intervals += y.Intervals
		intervalN += y.IntervalCount
		for i, n := range y.AdultDeathAges {
			adultAges[i] += n
		}
	}

	m := DemographyMetrics{
		WindowYears:   len(d.Years),
		PersonYears:   math.Round(total*10) / 10,
		DeathsByCause: causes,
		Under5ByCause: under5,
	}
	if total >= minExposure {
		if lt, ok := lifeTable(exposure, deaths); ok {
			m.LifeExpectancy = ptr(lt.e0, 1)
			m.SurvivalTo15 = ptr(lt.l15, 3)
			m.InfantMortality = ptr(lt.q0, 3)
			m.Under5Mortality = ptr(1-lt.l5, 3)
			if lt.l15 > 0 {
				m.LifeExpectancy15 = ptr(lt.e15, 1)
			}
		}
		m.HomicideRate = ptr(float64(causes.Killed)/total*1e5, 0)
		for d, n := range cases {
			m.Incidence[d] = ptr(float64(n)/total, 2)
		}
	}
	if v, ok := tfr(births, female); ok {
		m.TFR = ptr(v, 2)
	}
	if firsts >= minEvents {
		m.MeanAgeFirstBirth = ptr(firstAges/float64(firsts), 1)
	}
	if intervalN >= minEvents {
		m.MeanBirthInterval = ptr(intervals/float64(intervalN), 2)
	}
	if n := sumInts(adultAges[:]); n >= minAdultDeaths {
		mode := 0
		for i, c := range adultAges {
			if c > adultAges[mode] {
				mode = i
			}
		}
		m.ModalAgeAdultDeath = ptr(15+2*float64(mode)+1, 0)
	}

	females, males := 0, 0
	members := map[int64]int{}
	for _, c := range s.creatures {
		if c.Sex == Female {
			females++
		} else {
			males++
		}
		if h := s.houseOf(c); h != nil {
			members[h.ID]++
		}
	}
	if females > 0 {
		m.SexRatio = ptr(float64(males)*100/float64(females), 1)
	}
	if len(members) > 0 {
		m.HouseholdSize = ptr(float64(sumInts(mapValues(members)))/float64(len(members)), 2)
	}
	var wealth []float64
	for _, c := range s.creatures {
		if !s.adult(c) {
			continue
		}
		w := s.stockValue(c.Inventory)
		if h := s.houseOf(c); h != nil && members[h.ID] > 0 {
			w += s.stockValue(h.Storage) / float64(members[h.ID])
		}
		wealth = append(wealth, w)
	}
	if len(wealth) >= 2 {
		m.Gini = ptr(gini(wealth), 3)
	}
	return m
}

func (s *Sim) stockValue(st Stock) float64 {
	v := 0.0
	for id, n := range st {
		v += s.cat.item(id).Value * float64(n)
	}
	return v
}

func sumInts(vs []int) int {
	n := 0
	for _, v := range vs {
		n += v
	}
	return n
}

func mapValues(m map[int64]int) []int {
	out := make([]int, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

func (s *Sim) pyramid() []AgeBand {
	out := make([]AgeBand, 0, pyramidTo/5+1)
	for from := 0; from <= pyramidTo; from += 5 {
		b := AgeBand{From: from}
		if from < pyramidTo {
			to := from + 4
			b.To = &to
			b.Label = strconv.Itoa(from) + "–" + strconv.Itoa(to)
		} else {
			b.Label = strconv.Itoa(from) + "+"
		}
		out = append(out, b)
	}
	for _, c := range s.creatures {
		i := min(int(s.ageYears(c))/5, len(out)-1)
		if c.Sex == Female {
			out[i].Female++
		} else {
			out[i].Male++
		}
	}
	return out
}

// Demography reports the world's demographic indicators.
func (s *Sim) Demography() Demography {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Demography{
		SecondsPerYear: SecondsPerYear,
		Year:           s.year(),
		Pyramid:        s.pyramid(),
		Current:        s.stats.metrics(s),
		History:        append([]DemographyPoint{}, s.stats.History...),
		Reference:      demographyReference,
		Genetics:       s.geneticsInfo(),
	}
}

// DemographyReference returns the pre-modern reference ranges.
func DemographyReference() map[string]MetricRef { return maps.Clone(demographyReference) }

// deedRates are thefts and assaults, kindnesses and killings per simulated
// year over the window (totals grow without bound and are hard to read).
func (d *demography) deedRates() (crimes, kindness, kills float64) {
	if len(d.Years) == 0 {
		return 0, 0, 0
	}
	for _, y := range d.Years {
		crimes += float64(y.Crimes)
		kindness += float64(y.Kindness)
		kills += float64(y.Causes.Killed)
	}
	n := float64(len(d.Years))
	return round1(crimes / n), round1(kindness / n), round1(kills / n)
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }
