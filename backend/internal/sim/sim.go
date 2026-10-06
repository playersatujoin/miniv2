// Package sim runs an artificial-life world on top of a map: creatures with
// their own recurrent neural networks that eat, drink, mate, gather, craft,
// build homes, help or harm each other and die, passing a mix of both
// parents' brains on to their children. Humanity starts from Adam and Hawa.
package sim

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"sync"
	"time"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/ecology"
	"miniv2/backend/internal/world"
)

const (
	TicksPerSecond = 20
	dt             = 1.0 / TicksPerSecond

	// SecondsPerYear is the world's time scale: one simulated year lasts 8
	// simulated seconds. Ages, lifespans, seasons and demographic rates use it.
	SecondsPerYear = ecology.SecondsPerYear

	archiveSize   = 20
	founderNoise  = 0.5 // the first couple's brains are mostly instinct
	historyEvery  = 5 * TicksPerSecond
	historyMax    = 720
	eventsMax     = 80
	foodEvery     = 10                  // regrow deposits in batches every N ticks
	resourceEvery = 10 * TicksPerSecond // refresh the "sumber daya" map
	abilityEvery  = TicksPerSecond      // re-check "bisa membuat/membangun" once a second
)

// Body and metabolism. Rates are per simulated second.
const (
	bodyRadius       = 0.25
	maxTurnRate      = math.Pi
	slowWhileFeeding = 0.2

	// The brain is part of the basal cost (see brainCost): with the initial
	// 20 hidden neurons the total is the same 0.006 as before brains had a price.
	basalCost     = 0.0052
	moveCost      = 0.003
	visionCost    = 0.0004
	pregnancyCost = 0.002
	thirstRate    = 0.004
	eatRate       = 0.6
	foodValue     = 0.8
	drinkRate     = 0.25

	healRate      = 0.01
	reputationAge = 300.0 // seconds for reputation to fade by ~63 %

	// Fase 2: bodies burn energy faster than before (3×, eased to 2.5× in
	// Fase 3 once births follow nursing and conception is a monthly chance),
	// so a person with nothing to eat lasts a few years, and a dry season or
	// an El Niño drought matters. Faster (4–6×) starves the first families
	// outright on wild food; see PLAN.md, Hasil Fase 2.
	// Real fasting lasts about two months; like everything here it is
	// compressed, but now on the scale of the seasons.
	hungerScale = 2.5
	// Only rivers and lakes can be drunk from, so how long a body lasts dry
	// decides how far people can range from them. At 8× a person lasts
	// about four years without water, a little longer than without food on
	// the move. In life thirst kills far sooner; faster thirst than this
	// kills the first families before they find their way around.
	thirstScale = 8.0
)

// Reproduction.
const (
	adultAge     = 15 * SecondsPerYear // grown up at 15
	mateRange    = 1.2
	spouseRange  = 3.0
	partnerRange = 1.5
	mateEnergy   = 0.4
	conceiveCost = 0.08
	gestation    = 0.75 * SecondsPerYear // about nine months
	birthCost    = 0.15
	childEnergy  = 0.8
	maleCooldown = cycle // a father may try again next month
	// How soon a mother conceives again is up to her nursing (see body.go).
)

var validSpeeds = []int{0, 1, 2, 5, 10, 20}

var ErrInvalidSpeed = errors.New("speed must be one of 0, 1, 2, 5, 10, 20")

type Sex uint8

const (
	Female Sex = iota
	Male
)

func (s Sex) String() string {
	if s == Male {
		return "male"
	}
	return "female"
}

func (s Sex) symbol() string {
	if s == Male {
		return "♂"
	}
	return "♀"
}

type Action string

const (
	ActExplore Action = "explore"
	ActEat     Action = "eat"
	ActDrink   Action = "drink"
	ActMate    Action = "mate"
	ActRest    Action = "rest"
	ActGather  Action = "gather"
	ActCraft   Action = "craft"
	ActBuild   Action = "build"
	ActGive    Action = "give"
	ActSteal   Action = "steal"
	ActAttack  Action = "attack"
	ActTeach   Action = "teach"
	ActPlant   Action = "plant"
	ActHunt    Action = "hunt"
	ActHarvest Action = "harvest"
	ActFish    Action = "fish"
)

type Ref struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Pregnancy struct {
	Remaining        float64 `json:"remaining"`
	Father           Ref     `json:"father"`
	FatherGenome     *Genome `json:"fatherGenome"`
	FatherGeneration int     `json:"fatherGeneration"`
}

type Deeds struct {
	Kindness    int `json:"kindness"`
	Crimes      int `json:"crimes"`
	Kills       int `json:"kills"`
	Built       int `json:"built"`
	Crafted     int `json:"crafted"`
	Discoveries int `json:"discoveries"`
	Planted     int `json:"planted"`
	Harvested   int `json:"harvested"`
	Hunted      int `json:"hunted"`
	Tamed       int `json:"tamed"`
	// Two-person exchanges (interact.go).
	Talks  int `json:"talks,omitempty"`
	Trades int `json:"trades,omitempty"`
}

// Short-lived visual effects shown in stream frames, indexed into Creature.fx.
const (
	fxAttack = iota
	fxSteal
	fxGive
	fxBuild
	fxCraft
	fxGather
	fxTeach
	fxPlant
	fxHunt
	numFX
)

type Creature struct {
	ID         int64            `json:"id"`
	Name       string           `json:"name"`
	Sex        Sex              `json:"sex"`
	Generation int              `json:"generation"`
	Mother     *Ref             `json:"mother"`
	Father     *Ref             `json:"father"`
	Spouse     *Ref             `json:"spouse,omitempty"`
	BornTick   int64            `json:"bornTick"`
	X          float64          `json:"x"`
	Y          float64          `json:"y"`
	Heading    float64          `json:"heading"`
	Energy     float64          `json:"energy"`
	Hydration  float64          `json:"hydration"`
	Health     float64          `json:"health"`
	Reputation float64          `json:"reputation"`
	Pregnancy  *Pregnancy       `json:"pregnancy,omitempty"`
	Cooldown   float64          `json:"cooldown"`
	Children   int              `json:"children"`
	LastBirth  int64            `json:"lastBirth,omitempty"` // tick of her latest delivery
	NursingID  int64            `json:"nursingId,omitempty"` // her youngest, whom she breastfeeds until weaning
	HouseID    int64            `json:"houseId,omitempty"`
	Inventory  Stock            `json:"inventory,omitempty"`
	Deeds      Deeds            `json:"deeds"`
	Job        *Job             `json:"job,omitempty"`
	Relations  []Relation       `json:"relations,omitempty"`
	Memories   []PerceivedEvent `json:"memories,omitempty"`
	Travel     *Travel          `json:"travel,omitempty"`
	NextRoute  int64            `json:"nextRoute,omitempty"`
	Gathering  float64          `json:"gathering,omitempty"`   // progress towards the next gathered unit
	ActCD      float64          `json:"actCooldown,omitempty"` // seconds until the next give/steal/attack
	Hurt       float64          `json:"hurt,omitempty"`        // seconds the "diserang" sense stays lit
	Offender   *Ref             `json:"offender,omitempty"`    // who last attacked or robbed them
	Mauled     string           `json:"mauled,omitempty"`      // the animal species that last hurt them
	// Health (Fase 3b, see disease.go): the illness they have now, their
	// immunity to each disease (0–1), seconds still carrying malaria
	// parasites, worm load (0–1), health an illness took that is coming
	// back, and the germs swallowed with water since the last check.
	Ill        *Illness  `json:"ill,omitempty"`
	Immune     []float64 `json:"immune,omitempty"`
	Carrier    float64   `json:"carrier,omitempty"`
	Worms      float64   `json:"worms,omitempty"`
	Convalesce float64   `json:"convalesce,omitempty"`
	Swallowed  float64   `json:"swallowed,omitempty"`
	contact    float64   // closeness to people with a respiratory infection, this check
	struck     int64     // tick of the last blow from a person or an animal
	// fate is a death decided this tick (old age, childbirth), for reap.
	fate string
	// Pace is how fast it moved last tick relative to its top speed; animals
	// notice movement.
	Pace float64 `json:"pace,omitempty"`
	// Where it last drank: people remember the way back to the river.
	WaterX    float64    `json:"waterX,omitempty"`
	WaterY    float64    `json:"waterY,omitempty"`
	Genome    *Genome    `json:"genome"`
	Mind      *Mind      `json:"mind"`
	Hidden    []float64  `json:"hidden"`
	Cognition *Cognition `json:"cognition,omitempty"`
	Bumped    bool       `json:"bumped,omitempty"` // fed back as an input next tick

	// Engine adaptation II (v11): moods (affect.go), a two-person exchange
	// under way (interact.go), how the body gets about (locomotion.go) and
	// the village they belong to (village.go).
	Affect    Affect     `json:"affect"`
	Chat      *Chat      `json:"chat,omitempty"`
	Loco      Locomotion `json:"loco"`
	VillageID int64      `json:"villageId,omitempty"`
	// Save v12 (forage.go): food places they remember, and water carried in
	// tubes with the germs it was drawn with.
	FoodPlaces   []FoodPlace `json:"foodPlaces,omitempty"`
	WaterCarried float64     `json:"waterCarried,omitempty"`
	WaterGerms   float64     `json:"waterGerms,omitempty"`
	// Fase 3d (nutrition.go): protein and micronutrient status, and growth lost.
	Nutrition Nutrition `json:"nutrition,omitzero"`
	// Fase 3c (genetics.go): inbreeding at birth.
	Heredity Heredity `json:"heredity"`

	// Culture: what this creature knows how to do, who is teaching it now,
	// and whom it has taught.
	Skills       map[string]float64 `json:"skills,omitempty"`
	Teacher      *Ref               `json:"teacher,omitempty"`
	TeacherUntil int64              `json:"teacherUntil,omitempty"`
	Students     []int64            `json:"students,omitempty"`

	// Cached decisions, saved so a restored world resumes exactly.
	CanCraft   bool                    `json:"canCraft,omitempty"`
	CanBuild   bool                    `json:"canBuild,omitempty"`
	IdleWork   float64                 `json:"idleWork,omitempty"` // seconds before looking for a new job after finding none
	Want       map[chem.ItemID]float64 `json:"want,omitempty"`
	GatherPick *chem.Source            `json:"gatherPick,omitempty"` // the deposit being gathered, or
	GatherWhat string                  `json:"gatherWhat,omitempty"` // "food" from the ground, "harvest" or "fish"
	GatherTile int                     `json:"gatherTile,omitempty"` // the field or water being harvested or fished
	Sample     bool                    `json:"sample,omitempty"`     // carries something a station could reveal an element from

	// Per-tick state, recomputed every step.
	input     [NumInputs]float64
	output    [NumOutputs]float64
	wantsMate bool
	ate       bool
	drank     bool
	resting   bool
	action    Action
	fx        [numFX]int64 // tick until which each effect shows
	wantVec   []float32    // Want indexed like tileItems, for the senses; see appeal
	skills    skillVec     // Skills by technology index, refreshed each tick
}

type DeathCounts struct {
	Starvation int `json:"starvation"`
	Thirst     int `json:"thirst"`
	OldAge     int `json:"oldAge"`
	Killed     int `json:"killed"`
	Animal     int `json:"animal"`               // killed by a wild animal
	Childbirth int `json:"childbirth,omitempty"` // mothers who died giving birth
	Neonatal   int `json:"neonatal,omitempty"`   // babies who died in their first weeks
	// Infectious disease (Fase 3b).
	Diarrhea    int `json:"diarrhea,omitempty"`
	Malaria     int `json:"malaria,omitempty"`
	Respiratory int `json:"respiratory,omitempty"`
	// Engine adaptation II: fire, water and steep ground.
	Burned  int `json:"burned,omitempty"`
	Drowned int `json:"drowned,omitempty"`
	Fall    int `json:"fall,omitempty"`
}

// add counts one death of cause.
func (d *DeathCounts) add(cause string) {
	switch cause {
	case "starvation":
		d.Starvation++
	case "thirst":
		d.Thirst++
	case "killed":
		d.Killed++
	case "animal":
		d.Animal++
	case "childbirth":
		d.Childbirth++
	case "neonatal":
		d.Neonatal++
	case "diarrhea":
		d.Diarrhea++
	case "malaria":
		d.Malaria++
	case "respiratory":
		d.Respiratory++
	case "burned":
		d.Burned++
	case "drowned":
		d.Drowned++
	case "fall":
		d.Fall++
	default:
		d.OldAge++
	}
}

// plus adds up two tallies.
func (d DeathCounts) plus(o DeathCounts) DeathCounts {
	d.Starvation += o.Starvation
	d.Thirst += o.Thirst
	d.OldAge += o.OldAge
	d.Killed += o.Killed
	d.Animal += o.Animal
	d.Childbirth += o.Childbirth
	d.Neonatal += o.Neonatal
	d.Diarrhea += o.Diarrhea
	d.Malaria += o.Malaria
	d.Respiratory += o.Respiratory
	d.Burned += o.Burned
	d.Drowned += o.Drowned
	d.Fall += o.Fall
	return d
}

// disease is the deaths from infectious disease.
func (d DeathCounts) disease() int { return d.Diarrhea + d.Malaria + d.Respiratory }

type HistoryPoint struct {
	Time          float64 `json:"time"`
	Population    int     `json:"population"`
	Females       int     `json:"females"`
	Males         int     `json:"males"`
	AvgGeneration float64 `json:"avgGeneration"`
	MaxGeneration int     `json:"maxGeneration"`
	Elements      int     `json:"elements"`
	Houses        int     `json:"houses"`
	AvgBrainSize  float64 `json:"avgBrainSize"`
	AvgSkill      float64 `json:"avgSkill"`
	// Neurons grown during life: the average per person, the biggest brain,
	// and the running totals of neurons grown and pruned (v7).
	AvgGrown float64 `json:"avgGrown,omitempty"`
	MaxBrain int     `json:"maxBrain,omitempty"`
	Grown    int64   `json:"grown,omitempty"`
	Pruned   int64   `json:"pruned,omitempty"`
}

type Event struct {
	ID         int64   `json:"id"`
	Time       float64 `json:"time"`
	Kind       string  `json:"kind"`
	Text       string  `json:"text"`
	CreatureID int64   `json:"creatureId,omitempty"`
	Violent    bool    `json:"violent,omitempty"` // assaults and killings, for the log filter
}

// Discovery records when and by whom an element or technology became known.
type Discovery struct {
	Time   float64 `json:"time"`
	By     Ref     `json:"by"`
	Era    int     `json:"era"`
	Source string  `json:"source,omitempty"`
}

type archived struct {
	Genome  *Genome `json:"genome"`
	Fitness float64 `json:"fitness"`
}

// Sim is one living world. Exported methods are safe for concurrent use.
type Sim struct {
	mu       sync.Mutex
	mapID    string
	terrain  *terrain
	cat      *catalog
	geo      geology
	src      *rand.PCG
	rng      *rand.Rand
	capacity int

	// What is worth gathering: per item (rare = 1) and per tile.
	interest    map[chem.ItemID]float64
	tileItems   []uint8 // per tile, tileSlots item indexes gatherable there (tileToolBit = needs a tool)
	sampleKey   [2]int
	sampleItems []chem.ItemID

	tick      int64
	speed     int
	nextID    int64
	creatures []*Creature
	byID      map[int64]*Creature
	grid      grid
	// space finds who touches whom (personal space), on a finer grid; shove is its scratch.
	space grid
	shove [][2]float64
	// Neurons grown and pruned in all lives so far.
	grown, pruned int64

	structures    []*Structure
	structByID    map[int64]*Structure
	structAt      []*Structure // by tile index
	nextStructID  int64
	structVersion int64
	tier          int

	// Fase 2: the living land, and how often the technical population limit
	// stopped a conception (it should never bite in a normal world).
	eco        *ecology.Ecology
	ecoHistory []EcoPoint
	// waterHistory: the rivers twice a simulated month, for the last twenty
	// years (the seasons need a finer clock than ecoHistory).
	waterHistory []WaterPoint
	capHits      int
	lastEcoEvt   map[string]float64 // when each kind of farming/hunting event was last reported

	elements map[string]*Discovery
	techs    map[string]*Discovery

	// Culture: who can practise what (recounted every second), techs whose
	// last holder is gone, and who held them last.
	holders        map[string]int
	lost           map[string]bool
	lastHolder     map[string]Ref
	knowledgeLost  int
	lastLearnEvent float64
	learnedVia     map[string]int // how people became able to practise something, all-time
	techIndex      map[string]int // technology → position in skillVec

	births         int
	deaths         int
	deathsBy       DeathCounts
	crimes         int
	kindness       int
	kills          int
	lastKindEvent  float64
	lastCrimeEvent float64
	history        []HistoryPoint
	events         []Event
	nextEvent      int64
	archive        []archived
	milestone      int
	// era counts how many times humanity began from an Adam and Hawa.
	era int

	opts  Options
	stats *demography
	// onDeath, if set, hears of every death (for tests and diagnostics).
	onDeath func(c *Creature, cause string)
	// Fase 3b: mosquitoes, worm eggs and the epidemic curve.
	epi epidemiology
	// Talks, rumours and trades so far (interact.go).
	exch exchangeState

	frames broadcaster
	// replay records the last minutes as streamed (replay.go); not saved.
	replay recorder
	// Engine adaptation II state that lives on the world.
	village villageSys // village.go
	stimuli stimWorld  // stimuli.go: lasting stimuli and bodies
	share   sharing    // sharing.go: carcasses and what was shared
	// Fase 3c: the pedigree, inbreeding and its statistics (geneview.go).
	genetics genetics
	// Source format is only used to archive a save before its first migration.
	loadedVersion     int
	migrationOriginal []byte
}

// Options switch rules off for experiments (A/B soak runs). The zero value is
// the normal world.
type Options struct {
	// Experiments only: the normal world perceives local events and navigates.
	NoPerception bool `json:"noPerception,omitempty"`
	NoNavigation bool `json:"noNavigation,omitempty"`
	NoAIBudget   bool `json:"noAIBudget,omitempty"`
	// NoCrime stops anyone from stealing or attacking.
	NoCrime bool `json:"noCrime,omitempty"`
	// NoInstincts gives the first couple random brains without inborn reflexes.
	NoInstincts bool `json:"noInstincts,omitempty"`
	// NoLearning is the world before Fase 1: brains don't change during a
	// life, and whatever anyone ever discovered, everybody can do.
	NoLearning bool `json:"noLearning,omitempty"`
	// NoPlasticity and NoCulture switch off one half of NoLearning each, to
	// tell their effects apart.
	NoPlasticity bool `json:"noPlasticity,omitempty"`
	NoCulture    bool `json:"noCulture,omitempty"`

	// Fase 2 switches. NoHumans leaves the island to the animals (no Adam
	// and Hawa); NoFarming stops anyone from planting; NoClimate keeps the
	// rain at its yearly average (no seasons, no El Niño); NoFauna removes
	// wild animals.
	NoHumans  bool `json:"noHumans,omitempty"`
	NoFarming bool `json:"noFarming,omitempty"`
	NoClimate bool `json:"noClimate,omitempty"`
	NoFauna   bool `json:"noFauna,omitempty"`

	// NoPersonalSpace lets bodies overlap, as before personal space was added.
	NoPersonalSpace bool `json:"noPersonalSpace,omitempty"`
	// NoNeurogenesis keeps every brain at its inherited size for life.
	NoNeurogenesis bool `json:"noNeurogenesis,omitempty"`
	// NoWaterCycle keeps rivers and lakes full all year, as before the water cycle.
	NoWaterCycle bool `json:"noWaterCycle,omitempty"`

	// Fase 3 switches. NoDisease: nobody falls ill or gets worms.
	// NoSanitation: latrines keep nothing out of the water and the soil,
	// and wells are no cleaner than a hole dug by the river.
	NoDisease    bool `json:"noDisease,omitempty"`
	NoSanitation bool `json:"noSanitation,omitempty"`
	// NoAttachment lets young children roam alone instead of staying with
	// their mother or another carer.
	NoAttachment bool `json:"noAttachment,omitempty"`

	// Engine adaptation II switches: lasting stimuli, moods, two-person
	// exchanges (talk, gossip, trade), fire, travel over water and steep
	// ground, and villages with leaders.
	NoStimuli  bool `json:"noStimuli,omitempty"`
	NoAffect   bool `json:"noAffect,omitempty"`
	NoExchange bool `json:"noExchange,omitempty"`
	NoFire     bool `json:"noFire,omitempty"`
	NoMobility bool `json:"noMobility,omitempty"`
	NoVillages bool `json:"noVillages,omitempty"`

	// NoNutrition: food is only calories, as before Fase 3d.
	NoNutrition bool `json:"noNutrition,omitempty"`
	// NoSharing: meat a hunter can't carry vanishes, nobody eats from
	// others' food, only a field's family may harvest it, granaries serve
	// only their family, and nobody knows land beyond sight (sharing.go).
	NoSharing bool `json:"noSharing,omitempty"`
	// NoGenetics (Fase 3c): recessive disorders do no harm and the
	// mate-choice senses (kinship, looks) stay 0. Genes and the pedigree
	// are still tracked, for comparison.
	NoGenetics bool `json:"noGenetics,omitempty"`
}

func (o Options) ecology() ecology.Options {
	return ecology.Options{NoClimate: o.NoClimate, NoFauna: o.NoFauna, NoWaterCycle: o.NoWaterCycle, NoFire: o.NoFire}
}

func (s *Sim) noPlasticity() bool { return s.opts.NoLearning || s.opts.NoPlasticity }

func (s *Sim) noCulture() bool { return s.opts.NoLearning || s.opts.NoCulture }

func newSim(m *world.Map, seed uint64) *Sim {
	return newSimWith(m, seed, defaultCatalog())
}

func newSimWith(m *world.Map, seed uint64, cat *catalog) *Sim {
	src := rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)
	s := &Sim{
		mapID:        m.ID,
		cat:          cat,
		src:          src,
		rng:          rand.New(src),
		speed:        1,
		nextID:       1,
		nextStructID: 1,
		byID:         map[int64]*Creature{},
		structByID:   map[int64]*Structure{},
		elements:     map[string]*Discovery{},
		techs:        map[string]*Discovery{},
		holders:      map[string]int{},
		lost:         map[string]bool{},
		lastHolder:   map[string]Ref{},
		stats:        newDemography(),
		lastEcoEvt:   map[string]float64{},
		// The first learning in a world is always reported.
		lastLearnEvent: -learnEventGap,
	}
	s.setTerrain(newTerrain(m))
	s.epi.init(m.Width, m.Height)
	s.eco = ecology.New(ecology.LandFromMap(m), s.rng, ecology.Options{}, 0)
	if cat.newGeology != nil {
		s.geo = cat.newGeology(m)
	} else {
		s.geo = noGeology{}
	}
	s.computeInterest()
	s.refreshResources()
	return s
}

// New starts a fresh world from a single couple, Adam and Hawa.
func New(m *world.Map, seed uint64) *Sim {
	return NewWithOptions(m, seed, Options{})
}

// NewWithOptions is New with some rules switched off.
func NewWithOptions(m *world.Map, seed uint64, opts Options) *Sim {
	s := newSim(m, seed)
	s.opts = opts
	if opts.ecology() != (ecology.Options{}) {
		s.eco = ecology.New(ecology.LandFromMap(m), s.rng, opts.ecology(), 0)
	}
	s.applyFarms()
	s.genesis()
	s.frames.publish(s.encodeFrame())
	return s
}

func (s *Sim) setTerrain(t *terrain) {
	s.terrain = t
	if len(s.structAt) != len(t.blocked) {
		s.structAt = make([]*Structure, len(t.blocked))
		for _, st := range s.structures {
			if i, ok := t.index(st.X, st.Y); ok {
				s.structAt[i] = st
			}
		}
	}
	// Not a carrying capacity: food, water and space set that. This is only
	// a technical ceiling so a runaway world can't stall the server.
	s.capacity = int(clamp(float64(len(t.walkable))/6, 300, 2000))
}

func (s *Sim) time() float64 { return float64(s.tick) * dt }

func (s *Sim) age(c *Creature) float64 { return float64(s.tick-c.BornTick) * dt }

func (s *Sim) adult(c *Creature) bool { return s.age(c) >= adultAge }

// fertile reports whether c could conceive (or father a child) now: a woman
// before her menopause who is neither pregnant nor fully nursing.
func (s *Sim) fertile(c *Creature) bool {
	if !s.adult(c) || c.Pregnancy != nil || c.Cooldown > 0 {
		return false
	}
	if c.Sex == Female {
		return s.ageYears(c) < c.Genome.Traits.Menopause && s.nursingBlock(c) < 1
	}
	return true
}

func (s *Sim) add(c *Creature) {
	s.creatures = append(s.creatures, c)
	s.byID[c.ID] = c
}

// living returns the creature with id if it is alive (health > 0).
func (s *Sim) living(id int64) *Creature {
	if c := s.byID[id]; c != nil && c.Health > 0 {
		return c
	}
	return nil
}

// spawnAdult creates a creature without parents at (x, y), just old enough to
// reproduce.
func (s *Sim) spawnAdult(g *Genome, sex Sex, name string, x, y float64) *Creature {
	c := &Creature{
		ID:        s.nextID,
		Name:      name,
		Sex:       sex,
		BornTick:  s.tick - int64(math.Ceil(adultAge*TicksPerSecond)),
		X:         x,
		Y:         y,
		Heading:   s.rng.Float64()*2*math.Pi - math.Pi,
		Energy:    1,
		Hydration: 1,
		Health:    1,
		Genome:    g,
		WaterX:    x, // the first people start by water and know it
		WaterY:    y,
	}
	s.giveBrain(c)
	s.nextID++
	s.add(c)
	return c
}

// Run advances the world in real time until ctx is cancelled: 20 times a
// second it runs `speed` steps, and every 100 ms it publishes a stream frame.
func (s *Sim) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second / TicksPerSecond)
	defer ticker.Stop()
	publish := false
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		s.mu.Lock()
		for range s.speed {
			s.step()
		}
		var frame, structures []byte
		var tick, structVersion int64
		var at float64
		keep := false
		if publish = !publish; publish {
			frame = s.encodeFrame()
			if keep = s.replay.replayDue(); keep {
				tick, at, structVersion = s.tick, s.time(), s.structVersion
				if !s.replay.knowsStructures(structVersion) {
					structures = s.encodeStructures()
				}
			}
		}
		s.mu.Unlock()
		if frame != nil {
			s.frames.publish(frame)
		}
		if keep {
			s.replay.record(tick, at, frame, structVersion, structures)
		}
	}
}

// Step advances the world by one tick.
func (s *Sim) Step() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.step()
}

// Advance runs n ticks under one lock, for headless runs.
func (s *Sim) Advance(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for range n {
		s.step()
	}
}

func (s *Sim) step() {
	s.tick++
	if s.tick%foodEvery == 0 {
		s.geo.Regrow(foodEvery * dt)
	}
	if s.tick%resourceEvery == 0 {
		s.refreshResources()
		s.updateWoodland()
	}
	s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures)
	if !s.noCulture() {
		s.refreshSkillVecs()
	}
	s.eco.Tick(s.tick, s.time(), dt, s)
	for _, ev := range s.eco.TakeEvents() {
		s.event(ev.Kind, ev.Text, 0)
	}
	s.fireEffects()
	for _, c := range s.creatures {
		if c.Health <= 0 {
			continue // killed earlier this tick
		}
		if (s.tick+c.ID)%abilityEvery == 0 {
			s.updateAbilities(c)
		}
		s.act(c)
	}
	s.exchange()
	s.personalSpace()
	s.disease()
	s.nourish()
	s.senesce()
	s.mate()
	s.gestate()
	s.reap()
	s.stimuliTick()
	if s.tick%TicksPerSecond == 0 {
		s.stats.expose(s)
		s.cultureTick()
		s.spoil()
		s.villageTick()
		s.shareSecond()
		s.geneticsTick()
		s.eco.Current().Population = len(s.creatures)
	}
	if len(s.creatures) == 0 && !s.opts.NoHumans {
		s.event("milestone", fmt.Sprintf("Manusia punah di era %d. Adam & Hawa baru memulai era %d.", s.era, s.era+1), 0)
		s.genesis()
	}
	if s.tick%historyEvery == 0 {
		s.sample()
		s.sampleEcology()
	}
	if s.tick%waterEvery == 0 {
		s.sampleWater()
	}
	if s.tick%demographyEvery == 0 {
		s.stats.sample(s)
	}
}

func normAngle(a float64) float64 {
	a = math.Mod(a+math.Pi, 2*math.Pi)
	if a < 0 {
		a += 2 * math.Pi
	}
	return a - math.Pi
}

func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func (s *Sim) unborn() int {
	n := 0
	for _, c := range s.creatures {
		if c.Pregnancy != nil {
			n++
		}
	}
	return n
}

// mate pairs willing, fertile females with the closest willing male in reach.
func (s *Sim) mate() {
	room := s.capacity - len(s.creatures) - s.unborn()
	for _, f := range s.creatures {
		if f.Sex != Female || !f.wantsMate || !s.fertile(f) || f.Energy < mateEnergy || f.Health <= 0 {
			continue
		}
		var best *Creature
		bestD := mateRange
		s.grid.near(f.X, f.Y, mateRange+0.5, func(m *Creature) {
			if m.Sex != Male || !m.wantsMate || !s.fertile(m) || m.Energy < mateEnergy || m.Health <= 0 {
				return
			}
			if d := math.Hypot(m.X-f.X, m.Y-f.Y); d <= bestD {
				best, bestD = m, d
			}
		})
		// A married couple shares a camp: being within a few tiles is enough.
		if sp := s.livingSpouse(f); best == nil && sp != nil && sp.wantsMate && s.fertile(sp) &&
			sp.Energy >= mateEnergy && math.Hypot(sp.X-f.X, sp.Y-f.Y) <= spouseRange {
			best = sp
		}
		if best == nil {
			continue
		}
		if room <= 0 {
			s.capHits++
			continue
		}
		// One try a month: conception is a chance, by her age and her body.
		if s.rng.Float64() >= s.conceptionChance(f)*s.fertilityGene(best) {
			f.Cooldown = cycle
			continue
		}
		f.Pregnancy = &Pregnancy{
			Remaining:        gestation,
			Father:           Ref{best.ID, best.Name},
			FatherGenome:     best.Genome,
			FatherGeneration: best.Generation,
		}
		f.Energy -= conceiveCost
		best.Energy -= conceiveCost
		best.Cooldown = maleCooldown
		s.bond(f, best)
		room--
	}
}

func (s *Sim) gestate() {
	var born []*Creature
	for _, f := range s.creatures {
		p := f.Pregnancy
		if p == nil || f.Health <= 0 {
			continue
		}
		if p.Remaining -= dt; p.Remaining > 0 {
			continue
		}
		f.Pregnancy = nil
		n := 1
		if s.rng.Float64() < twinRate {
			n = 2
		}
		s.stats.delivery(s, f, n)
		s.eco.Current().Births += n
		first := len(born)
		for range n {
			c := s.newChild(f, p)
			if !s.opts.NoDisease && s.rng.Float64() < s.neonatalDeathRisk(f, n) {
				c.fate = "neonatal"
			}
			born = append(born, c)
		}
		// She nurses the baby that lived (if one did).
		f.NursingID = born[first].ID
		for _, c := range born[first:] {
			if c.fate == "" {
				f.NursingID = c.ID
				break
			}
		}
		f.Energy -= birthCost * float64(n)
		f.Cooldown = postpartum
		// Childbirth can kill the mother; her babies are born all the same.
		if s.rng.Float64() < s.childbirthRisk(f, n) {
			f.fate = "childbirth"
		}
		f.Children += n
		if father := s.byID[p.Father.ID]; father != nil {
			father.Children += n
		}
	}
	for _, c := range born {
		s.add(c)
		s.births++
		s.event("birth", fmt.Sprintf("%s %s lahir dari %s & %s (gen %d)",
			c.Name, c.Sex.symbol(), c.Mother.Name, c.Father.Name, c.Generation), c.ID)
		if c.Generation > s.milestone {
			s.milestone = c.Generation
			if c.Generation <= 3 || c.Generation%5 == 0 {
				s.event("milestone", fmt.Sprintf("Generasi %d tercapai!", c.Generation), 0)
			}
		}
	}
}

func (s *Sim) newChild(mother *Creature, p *Pregnancy) *Creature {
	g := crossover(mother.Genome, p.FatherGenome, s.rng).mutate(s.rng)
	sex := Sex(s.rng.IntN(2))
	x, y := mother.X, mother.Y
	if nx, ny := x+s.rng.Float64()*0.6-0.3, y+s.rng.Float64()*0.6-0.3; !s.terrain.blockedAt(nx, ny) {
		x, y = nx, ny
	}
	c := &Creature{
		ID:         s.nextID,
		Name:       newName(s.rng, sex),
		Sex:        sex,
		Generation: max(mother.Generation, p.FatherGeneration) + 1,
		Mother:     &Ref{mother.ID, mother.Name},
		Father:     &Ref{p.Father.ID, p.Father.Name},
		BornTick:   s.tick,
		X:          x,
		Y:          y,
		Heading:    s.rng.Float64()*2*math.Pi - math.Pi,
		Energy:     childEnergy,
		Hydration:  0.8,
		Health:     1,
		Genome:     g,
	}
	s.giveBrain(c)
	s.bornNourished(c, mother)
	// Children grow up in their mother's home, else their father's, and
	// know where she fetches water.
	c.HouseID = mother.HouseID
	c.WaterX, c.WaterY = mother.WaterX, mother.WaterY
	if father := s.byID[p.Father.ID]; c.HouseID == 0 && father != nil {
		c.HouseID = father.HouseID
	}
	s.inheritGenes(c, mother, p)
	s.nextID++
	return c
}

// reap removes creatures that starved, died of thirst, of old age or were killed.
func (s *Sim) reap() {
	alive := s.creatures[:0]
	var dead []*Creature
	for _, c := range s.creatures {
		if c.Health <= 0 || c.Energy <= 0 || c.Hydration <= 0 || c.fate != "" {
			dead = append(dead, c)
		} else {
			alive = append(alive, c)
		}
	}
	clear(s.creatures[len(alive):])
	s.creatures = alive
	for _, c := range dead {
		cause := c.fate
		if cause == "" {
			cause = s.diseaseDeath(c)
		}
		switch {
		case cause != "":
		case c.Health <= 0 && c.Mauled != "":
			cause = "animal"
		case c.Health <= 0:
			cause = "killed"
		case c.Energy <= 0:
			cause = "starvation"
		case c.Hydration <= 0:
			cause = "thirst"
		}
		s.die(c, cause)
	}
}

func (s *Sim) die(c *Creature, cause string) {
	delete(s.byID, c.ID)
	s.deaths++
	age := s.age(c)
	s.stats.death(s, c, cause)
	s.geneticsDeath(c)
	if s.onDeath != nil {
		s.onDeath(c, cause)
	}
	s.eco.Current().Deaths++
	var how string
	s.deathsBy.add(cause)
	switch cause {
	case "starvation":
		s.eco.Current().Starved++
		how = "mati kelaparan"
	case "animal":
		how = animalKill(c.Mauled)
	case "childbirth":
		how = "meninggal saat melahirkan"
	case "neonatal":
		how = "meninggal beberapa hari setelah lahir"
	case "diarrhea", "malaria", "respiratory":
		how = "meninggal karena " + c.Ill.Disease.name()
		s.epi.deaths++
	case "thirst":
		how = "mati kehausan"
	case "burned":
		how = "tewas terbakar"
	case "drowned":
		how = "tenggelam"
	case "fall":
		how = "tewas terjatuh"
	case "killed":
		how = "dibunuh"
		if c.Offender != nil {
			how += " oleh " + c.Offender.Name
			if killer := s.byID[c.Offender.ID]; killer != nil {
				how += " " + killer.Sex.symbol()
			}
		}
	default:
		how = "meninggal karena usia tua"
	}
	s.addEvent("death", fmt.Sprintf("%s %s %s pada usia %s", c.Name, c.Sex.symbol(), how, fmtAge(age)), c.ID, cause == "killed" || cause == "animal")
	s.leaveCorpse(c, cause)
	s.mourn(c)
	s.leaveBelongings(c)
	s.inherit(c)
	d := c.Deeds
	s.remember(c.Genome, float64(c.Children)*2+age/c.Genome.Traits.Lifespan+0.5*float64(d.Built+d.Discoveries))
}

func fmtAge(sec float64) string {
	return fmt.Sprintf("%d tahun", int(sec/SecondsPerYear))
}

// remember keeps the best genomes seen so a new Adam and Hawa can be bred from them.
func (s *Sim) remember(g *Genome, fitness float64) {
	if len(s.archive) < archiveSize {
		s.archive = append(s.archive, archived{g, fitness})
		return
	}
	worst := 0
	for i, a := range s.archive {
		if a.Fitness < s.archive[worst].Fitness {
			worst = i
		}
	}
	if fitness > s.archive[worst].Fitness {
		s.archive[worst] = archived{g, fitness}
	}
}

// genesis places a new first couple, Adam and Hawa, side by side by fresh water.
// Every later human descends from them. After an extinction the next couple is
// bred from the most successful genomes of earlier eras, so progress carries over.
func (s *Sim) genesis() {
	if s.opts.NoHumans {
		return
	}
	x, y, ok := s.eden()
	if !ok {
		return
	}
	s.era++
	s.milestone = 0
	adam := s.spawnAdult(s.foundingGenome(), Male, "Adam", x, y)
	hx, hy := x, y
	if nx, ny, ok := s.terrain.spotNear(s.rng, x, y, 0.6); ok {
		hx, hy = nx, ny
	}
	hawa := s.spawnAdult(s.foundingGenome(), Female, "Hawa", hx, hy)
	s.foundGenes(adam, hawa)
	s.event("genesis", fmt.Sprintf("Era %d dimulai: Adam %s", s.era, adam.Sex.symbol()), adam.ID)
	s.event("genesis", fmt.Sprintf("Era %d dimulai: Hawa %s", s.era, hawa.Sex.symbol()), hawa.ID)
}

// eden picks the most fertile of a sample of spots by a river or lake (or the
// shore, on a map without fresh water): water and plenty of food within
// reach, a garden for the first couple.
func (s *Sim) eden() (float64, float64, bool) {
	t := s.terrain
	bx, by, ok := t.randomSpot(s.rng)
	best := -1.0
	for range 64 {
		x, y, found := t.randomSpot(s.rng)
		if !found {
			break
		}
		food := 0.0
		for dy := -3; dy <= 3; dy++ {
			for dx := -3; dx <= 3; dx++ {
				if i, in := t.index(int(x)+dx, int(y)+dy); in {
					food += float64(s.eco.Forage(i))
				}
			}
		}
		if food > best {
			bx, by, best = x, y, food
		}
	}
	return bx, by, ok
}

// foundingGenome is random in the first era, and later the winner of a small
// tournament over the archive, mutated.
func (s *Sim) foundingGenome() *Genome {
	if len(s.archive) == 0 {
		if s.opts.NoInstincts {
			return randomGenomeWith(s.rng, founderNoise, false)
		}
		return randomGenomeWithNoise(s.rng, founderNoise)
	}
	best := s.archive[s.rng.IntN(len(s.archive))]
	for range 2 {
		if a := s.archive[s.rng.IntN(len(s.archive))]; a.Fitness > best.Fitness {
			best = a
		}
	}
	return best.Genome.clone().mutate(s.rng)
}

func (s *Sim) event(kind, text string, creatureID int64) { s.addEvent(kind, text, creatureID, false) }

func (s *Sim) addEvent(kind, text string, creatureID int64, violent bool) {
	s.nextEvent++
	s.events = append(s.events, Event{ID: s.nextEvent, Time: s.time(), Kind: kind, Text: text, CreatureID: creatureID, Violent: violent})
	s.replay.mark(ReplayMark{Tick: s.tick, Time: s.time(), Kind: kind, Text: text, CreatureID: creatureID, Importance: importance(kind, violent)})
	if n := len(s.events) - eventsMax; n > 0 {
		s.events = slices.Delete(s.events, 0, n)
	}
}

func (s *Sim) census() (females, males, maxGen int, avgGen float64) {
	sum := 0
	for _, c := range s.creatures {
		if c.Sex == Female {
			females++
		} else {
			males++
		}
		sum += c.Generation
		maxGen = max(maxGen, c.Generation)
	}
	if n := len(s.creatures); n > 0 {
		avgGen = float64(sum) / float64(n)
	}
	return
}

func (s *Sim) sample() {
	f, m, maxGen, avgGen := s.census()
	s.history = append(s.history, HistoryPoint{
		Time:          s.time(),
		Population:    len(s.creatures),
		Females:       f,
		Males:         m,
		AvgGeneration: math.Round(avgGen*100) / 100,
		MaxGeneration: maxGen,
		Elements:      len(s.elements),
		Houses:        s.houseCount(),
		AvgBrainSize:  math.Round(s.avgBrainSize()*10) / 10,
		AvgSkill:      math.Round(s.avgSkill()*1000) / 1000,
		AvgGrown:      math.Round(s.avgGrown()*10) / 10,
		MaxBrain:      s.maxBrain(),
		Grown:         s.grown,
		Pruned:        s.pruned,
	})
	if n := len(s.history) - historyMax; n > 0 {
		s.history = slices.Delete(s.history, 0, n)
	}
}

func (s *Sim) Speed() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.speed
}

func (s *Sim) SetSpeed(v int) error {
	if !slices.Contains(validSpeeds, v) {
		return ErrInvalidSpeed
	}
	s.mu.Lock()
	s.speed = v
	s.mu.Unlock()
	return nil
}

// UpdateMap swaps in edited terrain, keeping creatures, food and buildings where possible.
func (s *Sim) UpdateMap(m *world.Map) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.setTerrain(newTerrain(m))
	s.eco.UpdateLand(ecology.LandFromMap(m))
	s.geo.Update(m)
	s.removeBlockedStructures()
	s.applyFarms()
	s.relocateStranded()
	s.computeInterest()
	s.refreshResources()
}

// relocateStranded moves creatures standing on blocked tiles to the nearest
// walkable one. Someone wading, swimming or climbing (locomotion.go) is
// where a body can be and stays there, so a restored world carries on
// exactly; only walls, trees and the like (or the map's edge) displace.
func (s *Sim) relocateStranded() {
	for _, c := range s.creatures {
		if !s.terrain.blockedAt(c.X, c.Y) {
			continue
		}
		if i, ok := s.terrain.indexAt(c.X, c.Y); ok && !s.opts.NoMobility && s.terrain.mobilityAt(i) != mobNever {
			continue
		}
		if x, y, ok := s.terrain.nearestFree(c.X, c.Y); ok {
			c.X, c.Y = x, y
		}
	}
}

// broadcaster fans the latest encoded frame out to any number of streams.
type broadcaster struct {
	mu    sync.Mutex
	frame []byte
	next  chan struct{}
}

func (b *broadcaster) publish(frame []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.frame = frame
	if b.next != nil {
		close(b.next)
	}
	b.next = make(chan struct{})
}

// Frame returns the latest stream frame and a channel that is closed when the
// next one is published.
func (s *Sim) Frame() ([]byte, <-chan struct{}) {
	b := &s.frames
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.next == nil {
		b.next = make(chan struct{})
	}
	return b.frame, b.next
}

// giveBrain sets up a newborn's working brain: empty memory and nothing learned yet.
func (s *Sim) giveBrain(c *Creature) {
	c.Hidden = make([]float64, c.Genome.Hidden)
	c.Mind = newMind(c.Genome)
	c.Mind.Wellbeing = wellbeing(c)
}

// avgBrainSize is the mean number of neurons people think with, inherited and grown.
func (s *Sim) avgBrainSize() float64 {
	if len(s.creatures) == 0 {
		return 0
	}
	sum := 0
	for _, c := range s.creatures {
		sum += brainSize(c)
	}
	return float64(sum) / float64(len(s.creatures))
}

func (s *Sim) avgGrown() float64 {
	if len(s.creatures) == 0 {
		return 0
	}
	sum := 0
	for _, c := range s.creatures {
		sum += c.Mind.Grown
	}
	return float64(sum) / float64(len(s.creatures))
}

func (s *Sim) maxBrain() int {
	top := 0
	for _, c := range s.creatures {
		top = max(top, brainSize(c))
	}
	return top
}

// avgSkill is the mean of each adult's best skill.
func (s *Sim) avgSkill() float64 {
	n, sum := 0, 0.0
	for _, c := range s.creatures {
		if s.adult(c) {
			n++
			sum += s.bestSkill(c)
		}
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}
