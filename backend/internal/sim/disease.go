package sim

import (
	"fmt"
	"math"
)

// Infectious disease (Fase 3b). Among foragers and early farmers illness,
// not hunger or violence, killed most people, and above all the young: in
// the pre-modern world about one child in five died before its first
// birthday and as many again before fifteen (Gurven & Kaplan 2007; Volk &
// Atkinson 2013). Three diseases do most of that killing here, each spread
// the way it really spreads, plus the worms that sap everyone a little:
//
//   - Diarrhoea is caught from water fouled by faeces (see
//     ecology/pathogens.go) and from a fouled yard. It is worst for children
//     just being weaned (babies at the breast are mostly spared), in the last
//     pools of the dry season, and downstream of a camp; the chance saturates
//     with the dose, as germs' dose-response does. A latrine keeps a family's
//     filth out of the water and the yard; a lined well draws clean
//     groundwater. It kills by draining the body of water.
//   - Malaria comes from mosquitoes that breed in still water (pools, lake
//     shallows, rice paddies, rain puddles) in the lowlands. Following Ross
//     and Macdonald, the share of mosquitoes carrying the parasite rises with
//     the share of people they bite who carry it, and each person is bitten
//     more where there are more mosquitoes for each person. Repeated
//     infections build up a partial immunity, so in a malarious place the
//     adults mostly just carry the parasite while the children die of it.
//   - Respiratory infections (colds turning into pneumonia) pass between
//     people close together. Recovery gives immunity, but new strains keep
//     appearing (more often where more people live) and the old immunity
//     only partly protects: so epidemics come and go, and they burn through
//     a crowded village where they would fizzle out in a scattered band.
//   - Intestinal worms come from soil fouled near where people live. They
//     rarely kill, but they eat into their host's food and slow recovery.
//
// How badly an infection goes depends on age (infants and the old are
// frail), on being underfed (malnutrition and infection feed each other),
// on worms, on immunity and on an inherited Immunity gene, whose stronger
// defences cost energy. A sick person who rests and has family close by
// does better. Nobody knows about germs: only the latrine, the well and the
// family's care help, and nobody chooses them for that reason.
//
// Time is compressed like everything else (a year is 8 seconds): a bout of
// illness lasts a couple of weeks (a few tenths of a second).

type Disease uint8

const (
	Diarrhea Disease = iota
	Malaria
	Respiratory
	numDiseases
)

// diseaseKey names each disease in saves, the API and death causes.
var diseaseKey = [numDiseases]string{"diarrhea", "malaria", "respiratory"}

// diseaseName is how the event log names each disease.
var diseaseName = [numDiseases]string{"diare", "malaria", "radang paru (ISPA)"}

func (d Disease) String() string {
	if d < numDiseases {
		return diseaseKey[d]
	}
	return "unknown"
}

// Illness is a bout of disease.
type Illness struct {
	Disease Disease `json:"disease"`
	Left    float64 `json:"left"`   // seconds until it is over
	Length  float64 `json:"length"` // seconds it lasts in all
	// Severity is the health it would take over its whole course with
	// nothing to ease it; above what the body has left, it kills.
	Severity float64 `json:"severity"`
	Taken    float64 `json:"taken,omitempty"` // health it has taken so far
}

const (
	diseaseEvery = 4 // ticks between disease checks (a multiple of the water cycle's step)
	cellTiles    = 4 // tiles across a cell of the mosquito and soil maps

	// Each disease's course, in years, and how severe it is for a healthy,
	// well-fed young adult without immunity (before age and luck).
	diarrheaLength    = 0.04
	malariaLength     = 0.06
	respiratoryLength = 0.05
	diarrheaSeverity  = 0.10
	malariaSeverity   = 0.13
	respiratorySev    = 0.12
	severitySpread    = 1.0 // log-normal spread of how badly a bout goes

	// Diarrhoea: chance of falling ill for each unit of germs swallowed with
	// a unit of hydration, and the filth each person leaves (per year) and a
	// sick one (times as much).
	diarrheaDose = 20.0
	diarrheaMax  = 6.0 // most bouts a year, however foul the water and the yard
	filthPerYear = 4.0
	sickFilth    = 10.0
	// It also passes by hand and food in a fouled yard (the filth on the
	// soil, shared with the worm eggs), which crawling children who put
	// everything in their mouths catch most: bouts per year per unit of
	// filth on a tile.
	yardDose = 15.0
	// A hole in a dry bed: the sand filters most germs out.
	sandFilter = 0.25
	// dehydration is the share of a diarrhoea's severity also lost as water.
	dehydration = 0.5

	// Malaria: bites (per year, per mosquito per person) that pass the
	// parasite on, other blood the mosquitoes feed on (counted as people),
	// how long a mosquito lives (years), how readily it picks the parasite
	// up, and a trickle of infected mosquitoes from outside (monkeys,
	// birds), so the disease can start somewhere.
	biteRate      = 15.0
	mosquitoScale = 25.0 // mosquitoes for each unit of breeding site
	otherHosts    = 2.0
	mosquitoLife  = 0.04
	mosquitoCatch = 3.0 / mosquitoLife
	mosquitoSeed  = 0.004
	carrierYears  = 0.4 // years parasites linger after a bout (or a silent infection)
	carrierShare  = 0.3 // how infectious a carrier is to mosquitoes, against a sick person

	// Respiratory: chance per second of catching it from a sick person
	// right next to you (less further off, none beyond contactRange), and
	// how often a new strain turns up (a year).
	breathRate   = 2.0
	contactRange = 1.5
	crowd        = 4.0   // closeness at which crowding halves the chance of catching it
	newStrain    = 0.5   // most new strains a year (in a big population) …
	strainPeople = 200.0 // … and half that among this many people
	crossImmune  = 0.7   // what immunity to the old strains still does against a new one

	// Immunity (0–1): what one bout adds, and the years it takes to fade.
	diarrheaImmune    = 0.35
	diarrheaFade      = 3.0
	malariaImmune     = 0.2
	malariaFade       = 4.0
	respiratoryFade   = 2.5
	immuneMin         = 0.6 // range of the Immunity gene
	immuneMax         = 1.4
	immuneCost        = 0.15 // extra basal energy per unit of the gene above 1
	immunityInit      = 0.9  // founders: 0.9 + up to 0.2
	wormEggLife       = 0.5  // years worm eggs stay alive in the soil
	wormUptake        = 0.6  // most worm load picked up in a year, where eggs are thickest …
	wormHalf          = 0.3  // … and half that where this many eggs lie on a tile
	wormLife          = 1.5  // years worms live
	wormHunger        = 0.2  // extra energy a full worm load eats
	wormFrailty       = 0.3  // how much worse a full load makes any illness
	convalesceYears   = 0.08 // years to get back the health an illness took
	restEase          = 0.7  // how much rest eases an illness
	careEase          = 0.8  // … and family close by
	latrineReach      = 4    // tiles around a latrine whose people use it
	latrineCatch      = 0.9  // share of their filth it keeps out of water and soil
	feverCost         = 0.2  // extra basal energy burnt while ill
	sicknessSlows     = 0.5  // how much a bad illness slows someone down
	epiEvery          = TicksPerSecond
	epiHistoryMax     = 480
	outbreakEvery     = 3 * SecondsPerYear // between reports of an outbreak of the same disease
	outbreakMinCases  = 5
	outbreakMinShare  = 0.06
	malnourishedBelow = 0.5 // energy below which infections go worse
)

var diseaseLength = [numDiseases]float64{diarrheaLength, malariaLength, respiratoryLength}
var diseaseSeverity = [numDiseases]float64{diarrheaSeverity, malariaSeverity, respiratorySev}

// epidemiology is the world's side of disease: the mosquito and worm-egg
// maps, the epidemic curve, and scratch space for each check.
type epidemiology struct {
	cw, ch   int
	mosquito []float32 // per cell: mosquitoes (sum of breeding sites, spread by flight)
	infected []float32 // per cell: share of the mosquitoes carrying malaria
	soil     []float32 // per cell: worm eggs in the soil
	strains  int       // respiratory strains so far
	history  []EpiPoint
	deaths   int // disease deaths since the last history point

	// Scratch, rebuilt as needed (never carried across a save).
	people, biters []float32 // per cell: people, and how infectious to mosquitoes they are
	blur           []float32
	sanitary       []bool // per tile: people here use a latrine
	sanVersion     int64
}

// EpiPoint is the island's health at one moment, for the epidemic chart.
type EpiPoint struct {
	Time        float64 `json:"time"`
	Population  int     `json:"population"`
	Diarrhea    int     `json:"diarrhea"` // ill now
	Malaria     int     `json:"malaria"`
	Respiratory int     `json:"respiratory"`
	Carriers    int     `json:"carriers"` // carrying malaria without being ill
	Worms       float64 `json:"worms"`    // mean worm load (0–1)
	// Mosquitoes is the mosquito pressure where people are (bites per
	// person, relative), and Deaths the deaths from disease since the last point.
	Mosquitoes float64 `json:"mosquitoes"`
	Deaths     int     `json:"deaths"`
}

func (e *epidemiology) init(w, h int) {
	e.cw, e.ch = (w+cellTiles-1)/cellTiles, (h+cellTiles-1)/cellTiles
	n := e.cw * e.ch
	e.mosquito = make([]float32, n)
	e.infected = make([]float32, n)
	e.soil = make([]float32, n)
	e.people = make([]float32, n)
	e.biters = make([]float32, n)
	e.blur = make([]float32, n)
	e.sanVersion = -1
}

func (e *epidemiology) cellOf(x, y float64) int {
	cx := min(e.cw-1, max(0, int(x)/cellTiles))
	cy := min(e.ch-1, max(0, int(y)/cellTiles))
	return cy*e.cw + cx
}

// spread averages a cell map over each cell's neighbours (how far
// mosquitoes fly) into dst.
func (e *epidemiology) spread(src, dst []float32) {
	for cy := range e.ch {
		for cx := range e.cw {
			var sum, wsum float32
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					x, y := cx+dx, cy+dy
					if x < 0 || y < 0 || x >= e.cw || y >= e.ch {
						continue
					}
					w := float32(1)
					if dx != 0 && dy != 0 {
						w = 0.25
					} else if dx != 0 || dy != 0 {
						w = 0.5
					}
					sum += src[y*e.cw+x] * w
					wsum += w
				}
			}
			dst[cy*e.cw+cx] = sum / wsum
		}
	}
}

// --- the yearly-ish clock of each body ---------------------------------------

func (s *Sim) noDisease() bool { return s.opts.NoDisease }

// sickness is how ill c is now (0 well … 1 gravely).
func sickness(c *Creature) float64 {
	if c.Ill == nil {
		return 0
	}
	return clamp(0.25+1.5*c.Ill.Severity, 0, 1)
}

// frailty is how much worse an infection goes at an age (years) than in a
// young adult: infants worst, then the old; least of all around ten, the
// age at which people everywhere are least likely to die.
func frailty(age float64) float64 {
	switch {
	case age < 1:
		return 2.6
	case age < 5:
		return 2.6 - 1.6*(age-1)/4
	case age < 10:
		return 1 - 0.55*(age-5)/5
	case age < 15:
		return 0.45 + 0.1*(age-10)/5
	case age < 45:
		return 0.55
	}
	return math.Min(3, 0.55*math.Exp((age-45)/18))
}

func (c *Creature) immunity(d Disease) float64 {
	if int(d) < len(c.Immune) {
		return c.Immune[d]
	}
	return 0
}

func (c *Creature) setImmunity(d Disease, v float64) {
	if len(c.Immune) < int(numDiseases) {
		c.Immune = append(c.Immune, make([]float64, int(numDiseases)-len(c.Immune))...)
	}
	c.Immune[d] = clamp(v, 0, 1)
}

// severityOf draws how badly a new bout of d goes for c.
func (s *Sim) severityOf(c *Creature, d Disease) float64 {
	m := diseaseSeverity[d] * frailty(s.ageYears(c))
	if c.Energy < malnourishedBelow {
		m *= 1 + 2*(malnourishedBelow-c.Energy)
	}
	m *= 1 + wormFrailty*c.Worms
	m *= s.malnutritionSeverity(c)
	m *= (2 - c.Genome.Traits.Immunity) * s.diseaseGene(c, d)
	switch d {
	case Diarrhea:
		m *= 1 - 0.5*c.immunity(d)
	case Malaria:
		m *= 1 - 0.8*c.immunity(d)
	}
	return m * math.Exp(severitySpread*s.rng.NormFloat64()-severitySpread*severitySpread/2)
}

// fallIll starts a bout of d.
func (s *Sim) fallIll(c *Creature, d Disease) {
	length := diseaseLength[d] * SecondsPerYear * s.malnutritionLength(c, d)
	c.Ill = &Illness{Disease: d, Left: length, Length: length, Severity: s.severityOf(c, d)}
	s.stats.current(s).Cases[d]++
}

// recover ends c's illness and leaves its immunity behind.
func (s *Sim) recover(c *Creature) {
	ill := c.Ill
	c.Ill = nil
	c.Convalesce += ill.Taken
	switch ill.Disease {
	case Diarrhea:
		c.setImmunity(Diarrhea, c.immunity(Diarrhea)+diarrheaImmune)
	case Malaria:
		c.setImmunity(Malaria, c.immunity(Malaria)+malariaImmune)
		c.Carrier = carrierYears * SecondsPerYear
	case Respiratory:
		c.setImmunity(Respiratory, 1)
	}
}

// --- the check every few ticks -----------------------------------------------

// disease runs a check of everyone's health: what they caught, how their
// illness goes, and the mosquitoes, germs and worms around them.
func (s *Sim) disease() {
	if s.noDisease() || s.tick%diseaseEvery != 0 {
		return
	}
	e := &s.epi
	sec := float64(diseaseEvery) * dt
	yr := sec / SecondsPerYear
	s.refreshSanitation()

	// Where people are, and how infectious to mosquitoes.
	clear(e.people)
	clear(e.biters)
	for _, c := range s.creatures {
		k := e.cellOf(c.X, c.Y)
		e.people[k]++
		switch {
		case c.Ill != nil && c.Ill.Disease == Malaria:
			e.biters[k]++
		case c.Carrier > 0:
			e.biters[k] += carrierShare
		}
	}
	e.spread(e.people, e.blur)
	if s.tick%TicksPerSecond == 0 {
		s.breedMosquitoes()
	}

	// Filth: into the water and onto the soil.
	eggs := float32(math.Exp(-yr / wormEggLife))
	for k := range e.soil {
		e.soil[k] *= eggs
	}
	for _, c := range s.creatures {
		i, ok := s.terrain.indexAt(c.X, c.Y)
		if !ok {
			continue
		}
		filth := s.bodyScale(c) * yr
		if s.sanitary(i) {
			filth *= 1 - latrineCatch
		}
		e.soil[e.cellOf(c.X, c.Y)] += float32(filth)
		if c.Ill != nil && c.Ill.Disease == Diarrhea {
			filth *= sickFilth
		}
		s.eco.Pollute(i, filth*filthPerYear)
	}

	// Coughs: everyone near someone with a respiratory infection.
	for _, c := range s.creatures {
		c.contact = 0
	}
	for _, c := range s.creatures {
		if c.Ill == nil || c.Ill.Disease != Respiratory || c.Health <= 0 {
			continue
		}
		s.grid.near(c.X, c.Y, contactRange, func(o *Creature) {
			if o == c {
				return
			}
			if d := math.Hypot(o.X-c.X, o.Y-c.Y); d < contactRange {
				o.contact += 1 - d/contactRange
			}
		})
	}
	// Babies share what their mother drinks, more as they are weaned.
	for _, c := range s.creatures {
		if m := s.carrier(c); m != nil {
			c.Swallowed += m.Swallowed * s.weaned(c)
		}
	}

	// A new respiratory strain now and then, caught by someone somewhere: from
	// animals or by mutation, so more often the more people there are. An
	// isolated band of a dozen rarely meets one (Black 1975: crowd diseases
	// need crowds).
	if n := float64(len(s.creatures)); n > 0 && s.rng.Float64() < 1-math.Exp(-newStrain*n/(n+strainPeople)*yr) {
		e.strains++
		for _, c := range s.creatures {
			if v := c.immunity(Respiratory); v > 0 {
				c.setImmunity(Respiratory, v*crossImmune)
			}
		}
		if c := s.creatures[s.rng.IntN(len(s.creatures))]; c.Ill == nil {
			s.fallIll(c, Respiratory)
		}
	}

	for _, c := range s.creatures {
		if c.Health <= 0 || c.fate != "" {
			continue
		}
		s.course(c, sec, yr)
		if c.Ill == nil {
			s.catchDisease(c, yr)
		}
		c.Swallowed = 0
	}
	if s.tick%epiEvery == 0 {
		s.sampleEpidemic()
	}
}

// yardContact is how much of the yard's filth reaches c's mouth: most for
// toddlers, little for babies at the breast and for grown-ups.
func (s *Sim) yardContact(c *Creature) float64 {
	switch age := s.ageYears(c); {
	case age < 2:
		return s.weaned(c)
	case age < 5:
		return 1
	case age < 15:
		return 0.4
	}
	return 0.2
}

// weaned is how much of a baby's food is no longer its mother's milk:
// hardly any before six months, all of it by two.
func (s *Sim) weaned(c *Creature) float64 {
	return clamp((s.ageYears(c)-0.5)/1.5, 0.05, 1)
}

// catchDisease gives c, who is well, the chance to catch something.
func (s *Sim) catchDisease(c *Creature, yr float64) {
	e := &s.epi
	var risk [numDiseases]float64
	k := e.cellOf(c.X, c.Y)
	// Bouts a year the water and the yard would give; the chance saturates
	// (the beta-Poisson dose-response of enteric germs): however foul the
	// water, nobody falls ill more than diarrheaMax times a year.
	exposure := diarrheaDose*c.Swallowed/yr + float64(e.soil[k])/(cellTiles*cellTiles)*yardDose*s.yardContact(c)
	risk[Diarrhea] = diarrheaMax * exposure / (exposure + diarrheaMax) * yr * (1 - 0.85*c.immunity(Diarrhea))
	// In a crowd one is close to only so many people at once.
	contact := c.contact / (1 + c.contact/crowd)
	risk[Respiratory] = breathRate * contact * float64(diseaseEvery) * dt * (1 - 0.95*c.immunity(Respiratory))
	if m := float64(e.mosquito[k]) * mosquitoScale; m > 0 {
		risk[Malaria] = biteRate * m / (float64(e.blur[k]) + otherHosts) * (float64(e.infected[k]) + mosquitoSeed) * yr
	}
	total := risk[0] + risk[1] + risk[2]
	if total <= 0 {
		return
	}
	p := 1 - math.Exp(-total)
	u := s.rng.Float64()
	if u >= p {
		return
	}
	// Which one, in proportion to its share of the risk.
	v := u / p * total
	d := Diarrhea
	for d < numDiseases-1 && v > risk[d] {
		v -= risk[d]
		d++
	}
	if d == Malaria {
		// The more immune, the more likely an infection passes unnoticed
		// (but its parasites can still be passed on).
		immune := c.immunity(Malaria)
		c.setImmunity(Malaria, immune+malariaImmune*0.5)
		if s.rng.Float64() < 0.85*immune {
			c.Carrier = carrierYears * SecondsPerYear
			return
		}
	}
	s.fallIll(c, d)
}

// course moves c's illness, immunity, worms and recovery on by sec seconds
// (yr years).
func (s *Sim) course(c *Creature, sec, yr float64) {
	if ill := c.Ill; ill != nil {
		harm := ill.Severity * sec / ill.Length
		if c.resting {
			harm *= restEase
		}
		if s.caredFor(c) {
			harm *= careEase
		}
		c.Health -= harm
		ill.Taken += harm
		body := s.bodyScale(c)
		c.Energy -= feverCost * basalCost * hungerScale * body * sec
		if ill.Disease == Diarrhea {
			c.Hydration -= dehydration * harm
		}
		if ill.Left -= sec; ill.Left <= 0 && c.Health > 0 {
			s.recover(c)
		}
	} else if c.Convalesce > 0 && c.Energy > 0.3 && c.Hydration > 0.3 {
		give := math.Min(c.Convalesce, sec/(convalesceYears*SecondsPerYear)*(1-0.5*c.Worms))
		c.Convalesce -= give
		c.Health = math.Min(1, c.Health+give)
	}
	if c.Carrier > 0 {
		c.Carrier = math.Max(0, c.Carrier-sec)
	}
	for d, fade := range [numDiseases]float64{diarrheaFade, malariaFade, respiratoryFade} {
		if v := c.immunity(Disease(d)); v > 0 {
			v *= math.Exp(-yr / fade)
			if v < 1e-3 {
				v = 0
			}
			c.setImmunity(Disease(d), v)
		}
	}
	// Worms: picked up from fouled soil (children, playing in it, most), dying off in a year or two.
	e := &s.epi
	eggs := float64(e.soil[e.cellOf(c.X, c.Y)]) / (cellTiles * cellTiles)
	uptake := wormUptake * eggs / (eggs + wormHalf)
	if age := s.ageYears(c); age >= 2 && age < 15 {
		uptake *= 1.5
	}
	c.Worms = clamp(c.Worms+(uptake*(1-c.Worms)-c.Worms/wormLife)*yr, 0, 1)
	if c.Worms < 1e-4 {
		c.Worms = 0
	}
}

// caredFor reports whether a grown relative is at c's side.
func (s *Sim) caredFor(c *Creature) bool {
	cared := false
	s.grid.near(c.X, c.Y, contactRange, func(o *Creature) {
		if !cared && o != c && o.Health > 0 && s.adult(o) && s.kin(c, o) && math.Hypot(o.X-c.X, o.Y-c.Y) < contactRange {
			cared = true
		}
	})
	return cared
}

// --- mosquitoes and latrines ---------------------------------------------------

// breedMosquitoes renews the mosquito map from where they breed now, and
// how many of them carry malaria from the people they have bitten.
func (s *Sim) breedMosquitoes() {
	e := &s.epi
	t := s.terrain
	sites := make([]float32, len(e.mosquito))
	for i := range t.w * t.h {
		if b := s.eco.Breeding(i); b > 0 {
			sites[e.cellOf(float64(i%t.w), float64(i/t.w))] += float32(b)
		}
	}
	e.spread(sites, e.mosquito)
	// People and parasites, spread like the mosquitoes that carry them.
	biters := make([]float32, len(e.biters))
	e.spread(e.biters, biters)
	people := e.blur
	// A mosquito lives a couple of weeks, so a second (six weeks) is enough
	// for the infected share to settle where its bites put it.
	for k := range e.infected {
		x := 0.0
		if people[k] > 0 {
			x = float64(biters[k] / (people[k] + otherHosts))
		}
		e.infected[k] = float32(mosquitoCatch * x / (mosquitoCatch*x + 1/mosquitoLife))
	}
}

// refreshSanitation marks the tiles whose people use a latrine, when the buildings change.
func (s *Sim) refreshSanitation() {
	e := &s.epi
	if e.sanVersion == s.structVersion && e.sanitary != nil {
		return
	}
	e.sanVersion = s.structVersion
	t := s.terrain
	if len(e.sanitary) != t.w*t.h {
		e.sanitary = make([]bool, t.w*t.h)
	} else {
		clear(e.sanitary)
	}
	if s.opts.NoSanitation {
		return
	}
	for _, st := range s.structures {
		if !st.kind.Latrine {
			continue
		}
		for dy := -latrineReach; dy <= latrineReach; dy++ {
			for dx := -latrineReach; dx <= latrineReach; dx++ {
				if dx*dx+dy*dy > latrineReach*latrineReach {
					continue
				}
				if i, ok := t.index(st.X+dx, st.Y+dy); ok {
					e.sanitary[i] = true
				}
			}
		}
	}
}

func (s *Sim) sanitary(i int) bool {
	return i >= 0 && i < len(s.epi.sanitary) && s.epi.sanitary[i]
}

// germsDrunk is how foul the water c drinks from tile i is.
func (s *Sim) germsDrunk(kind, i int) float64 {
	switch kind {
	case drinkOpen:
		return s.eco.Germs(i)
	case drinkDug:
		return s.eco.Germs(i) * sandFilter
	case drinkWell:
		if s.opts.NoSanitation {
			// An unlined hole is no better than one in a riverbed.
			return s.eco.Germs(s.nearestFresh(i)) * sandFilter
		}
	}
	return 0
}

// nearestFresh is the fresh water tile closest to tile i within a few tiles, or -1.
func (s *Sim) nearestFresh(i int) int {
	t := s.terrain
	x, y := i%t.w, i/t.w
	for r := 0; r <= 3; r++ { // a well is within wellReach of where people stand
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if max(abs(dx), abs(dy)) != r {
					continue
				}
				if j, ok := t.index(x+dx, y+dy); ok && t.fresh[j] {
					return j
				}
			}
		}
	}
	return -1
}

// --- the record --------------------------------------------------------------

// diseaseDeath is the cause of death of c, if an illness killed them: they
// were ill and no blow (from a person or an animal) finished them this tick.
func (s *Sim) diseaseDeath(c *Creature) string {
	if c.Ill == nil || c.Health <= 0 && c.struck == s.tick {
		return ""
	}
	return c.Ill.Disease.String()
}

// sampleEpidemic adds a point to the epidemic curve and reports outbreaks.
func (s *Sim) sampleEpidemic() {
	e := &s.epi
	p := EpiPoint{Time: s.time(), Population: len(s.creatures), Deaths: e.deaths}
	e.deaths = 0
	var ill [numDiseases]int
	worms, bites := 0.0, 0.0
	for _, c := range s.creatures {
		if c.Ill != nil {
			ill[c.Ill.Disease]++
		} else if c.Carrier > 0 {
			p.Carriers++
		}
		worms += c.Worms
		k := e.cellOf(c.X, c.Y)
		bites += float64(e.mosquito[k]) * mosquitoScale / (float64(e.blur[k]) + otherHosts)
	}
	p.Diarrhea, p.Malaria, p.Respiratory = ill[Diarrhea], ill[Malaria], ill[Respiratory]
	if n := float64(len(s.creatures)); n > 0 {
		p.Worms = math.Round(worms/n*1000) / 1000
		p.Mosquitoes = math.Round(bites/n*1000) / 1000
	}
	e.history = append(e.history, p)
	if over := len(e.history) - epiHistoryMax; over > 0 {
		e.history = append(e.history[:0], e.history[over:]...)
	}
	for d := range numDiseases {
		n := ill[d]
		if n < outbreakMinCases || float64(n) < outbreakMinShare*float64(len(s.creatures)) {
			continue
		}
		key := "outbreak-" + diseaseKey[d]
		if last, ok := s.lastEcoEvt[key]; ok && s.time()-last < outbreakEvery {
			continue
		}
		s.lastEcoEvt[key] = s.time()
		s.event("disease", fmt.Sprintf("Wabah %s (tahun %d): %d orang sakit", diseaseName[d], s.year(), n), 0)
	}
}

// name is how the event log names d.
func (d Disease) name() string {
	if d < numDiseases {
		return diseaseName[d]
	}
	return "penyakit"
}

// seedImmunity gives the people of a world saved before disease existed
// what they would have had if it had been there all along: immunity built
// up over their years (more the older they are), malaria parasites in some
// adults and a light worm load, so the diseases arrive as old companions
// rather than as a "virgin soil" epidemic that would sweep away half of them.
func (s *Sim) seedImmunity() {
	for _, c := range s.creatures {
		age := s.ageYears(c)
		c.setImmunity(Diarrhea, clamp(age/8, 0, 0.8))
		c.setImmunity(Respiratory, clamp(age/20, 0, 0.6))
		c.setImmunity(Malaria, clamp(age/25, 0, 0.6))
		c.Worms = 0.2 * clamp(age/5, 0, 1)
	}
}
