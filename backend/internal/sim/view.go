package sim

import (
	"math"
	"slices"
	"strconv"

	"miniv2/backend/internal/chem"
)

// Flags in stream frames; keep in sync with FLAG in frontend/src/sim/protocol.ts.
const (
	flagPregnant  = 1
	flagEating    = 2
	flagDrinking  = 4
	flagWantsMate = 8
	flagResting   = 16
	flagChild     = 32
	flagAttacking = 64
	flagStealing  = 128
	flagGiving    = 256
	flagBuilding  = 512
	flagCrafting  = 1024
	flagGathering = 2048
	flagHead      = 4096
	flagCarrying  = 8192
	flagTeaching  = 16384
)

var fxFlags = [numFX]int{fxAttack: flagAttacking, fxSteal: flagStealing, fxGive: flagGiving,
	fxBuild: flagBuilding, fxCraft: flagCrafting, fxGather: flagGathering, fxTeach: flagTeaching}

type Info struct {
	MapID              string         `json:"mapId"`
	Tick               int64          `json:"tick"`
	Time               float64        `json:"time"`
	SecondsPerYear     float64        `json:"secondsPerYear"`
	Year               int            `json:"year"`
	Speed              int            `json:"speed"`
	Era                int            `json:"era"`
	Tier               int            `json:"tier"`
	TierName           string         `json:"tierName"`
	Population         int            `json:"population"`
	Females            int            `json:"females"`
	Males              int            `json:"males"`
	Capacity           int            `json:"capacity"`
	Births             int            `json:"births"`
	Deaths             int            `json:"deaths"`
	DeathsByCause      deathCounts    `json:"deathsByCause"`
	MaxGeneration      int            `json:"maxGeneration"`
	AvgGeneration      float64        `json:"avgGeneration"`
	ElementsDiscovered int            `json:"elementsDiscovered"`
	ElementsTotal      int            `json:"elementsTotal"`
	Houses             int            `json:"houses"`
	Structures         int            `json:"structures"`
	Crimes             int            `json:"crimes"`
	Kindness           int            `json:"kindness"`
	Kills              int            `json:"kills"`
	CrimesPerYear      float64        `json:"crimesPerYear"`
	KindnessPerYear    float64        `json:"kindnessPerYear"`
	KillsPerYear       float64        `json:"killsPerYear"`
	AvgBrainSize       float64        `json:"avgBrainSize"`
	KnowledgeLost      int            `json:"knowledgeLost"`
	History            []HistoryPoint `json:"history"`
	Events             []Event        `json:"events"`
}

func (s *Sim) Info() Info {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, m, maxGen, avgGen := s.census()
	crimes, kindness, kills := s.stats.deedRates()
	return Info{
		MapID:              s.mapID,
		Tick:               s.tick,
		Time:               s.time(),
		SecondsPerYear:     SecondsPerYear,
		Year:               s.year(),
		Speed:              s.speed,
		Era:                s.era,
		Tier:               s.tier,
		TierName:           s.cat.tierLabel(s.tier),
		Population:         len(s.creatures),
		Females:            f,
		Males:              m,
		Capacity:           s.capacity,
		Births:             s.births,
		Deaths:             s.deaths,
		DeathsByCause:      s.deathsBy,
		MaxGeneration:      maxGen,
		AvgGeneration:      math.Round(avgGen*100) / 100,
		ElementsDiscovered: len(s.elements),
		ElementsTotal:      s.cat.totalElements(),
		Houses:             s.houseCount(),
		Structures:         len(s.structures),
		Crimes:             s.crimes,
		Kindness:           s.kindness,
		Kills:              s.kills,
		CrimesPerYear:      crimes,
		KindnessPerYear:    kindness,
		KillsPerYear:       kills,
		AvgBrainSize:       math.Round(s.avgBrainSize()*10) / 10,
		KnowledgeLost:      s.knowledgeLost,
		History:            append([]HistoryPoint{}, s.history...), // never null in JSON
		Events:             append([]Event{}, s.events...),
	}
}

type TraitsView struct {
	Hue          float64 `json:"hue"`
	Size         float64 `json:"size"`
	MaxSpeed     float64 `json:"maxSpeed"`
	Vision       float64 `json:"vision"`
	Metabolism   float64 `json:"metabolism"`
	MutationRate float64 `json:"mutationRate"`
}

type BrainView struct {
	InputLabels  []string    `json:"inputLabels"`
	OutputLabels []string    `json:"outputLabels"`
	Size         int         `json:"size"`
	LearningRate float64     `json:"learningRate"`
	Learned      []float64   `json:"learned"`
	Input        []float64   `json:"input"`
	Hidden       []float64   `json:"hidden"`
	Output       []float64   `json:"output"`
	WIn          [][]float64 `json:"wIn"`
	WRec         [][]float64 `json:"wRec"`
	WOut         [][]float64 `json:"wOut"`
	BOut         []float64   `json:"bOut"`
}

type StackView struct {
	Item string `json:"item"`
	Name string `json:"name"`
	Qty  int    `json:"qty"`
}

type HouseView struct {
	ID       int64       `json:"id"`
	Kind     string      `json:"kind"`
	Name     string      `json:"name"`
	Level    int         `json:"level"`
	X        int         `json:"x"`
	Y        int         `json:"y"`
	Head     *Ref        `json:"head"`
	Members  int         `json:"members"`
	Storage  []StackView `json:"storage"`
	Capacity int         `json:"capacity"`
}

type CreatureDetail struct {
	SecondsPerYear float64     `json:"secondsPerYear"`
	ID             int64       `json:"id"`
	Name           string      `json:"name"`
	Sex            string      `json:"sex"`
	Generation     int         `json:"generation"`
	Mother         *Ref        `json:"mother"`
	Father         *Ref        `json:"father"`
	Spouse         *Ref        `json:"spouse"`
	BornAt         float64     `json:"bornAt"`
	Age            float64     `json:"age"`
	Lifespan       float64     `json:"lifespan"`
	Adult          bool        `json:"adult"`
	Energy         float64     `json:"energy"`
	Hydration      float64     `json:"hydration"`
	Health         float64     `json:"health"`
	Reputation     float64     `json:"reputation"`
	Role           string      `json:"role"`
	House          *HouseView  `json:"house"`
	Inventory      []StackView `json:"inventory"`
	Deeds          Deeds       `json:"deeds"`
	Skills         []SkillView `json:"skills"`
	Teacher        *Ref        `json:"teacher"`
	Taught         int         `json:"taught"`
	Pregnant       bool        `json:"pregnant"`
	Gestation      float64     `json:"gestation"`
	Children       int         `json:"children"`
	X              float64     `json:"x"`
	Y              float64     `json:"y"`
	Action         Action      `json:"action"`
	Traits         TraitsView  `json:"traits"`
	Brain          BrainView   `json:"brain"`
}

func r3(v float64) float64 { return math.Round(v*1000) / 1000 }

func round3[T float32 | float64](vs []T) []float64 {
	out := make([]float64, len(vs))
	for i, v := range vs {
		out[i] = r3(float64(v))
	}
	return out
}

// matrix reshapes a flat row-major weight slice into rows of `cols`.
func matrix(flat weights, cols int) [][]float64 {
	rows := make([][]float64, len(flat)/cols)
	for i := range rows {
		rows[i] = round3(flat[i*cols : (i+1)*cols])
	}
	return rows
}

// stacks lists a stock for display, largest first.
func (s *Sim) stacks(st Stock) []StackView {
	out := []StackView{}
	for _, id := range st.ids() {
		out = append(out, StackView{Item: string(id), Name: s.cat.itemName(id), Qty: st[id]})
	}
	slices.SortStableFunc(out, func(a, b StackView) int { return b.Qty - a.Qty })
	return out
}

func (s *Sim) houseView(st *Structure) *HouseView {
	members := 0
	for _, c := range s.creatures {
		if c.HouseID == st.ID {
			members++
		}
	}
	var head *Ref
	if st.Owner != 0 {
		head = &Ref{st.Owner, st.OwnerName}
	}
	return &HouseView{
		ID:       st.ID,
		Kind:     st.Kind,
		Name:     st.kind.Name,
		Level:    st.kind.Level,
		X:        st.X,
		Y:        st.Y,
		Head:     head,
		Members:  members,
		Storage:  s.stacks(st.Storage),
		Capacity: st.kind.Storage,
	}
}

// Creature returns a live snapshot of one creature, or false if it is dead
// or never existed.
func (s *Sim) Creature(id int64) (CreatureDetail, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.byID[id]
	if !ok {
		return CreatureDetail{}, false
	}
	g := c.Genome
	tr := g.Traits
	progress := 0.0
	if c.Pregnancy != nil {
		progress = 1 - c.Pregnancy.Remaining/gestation
	}
	role := "none"
	var house *HouseView
	if h := s.houseOf(c); h != nil {
		role = "member"
		if h.Owner == c.ID {
			role = "head"
		}
		house = s.houseView(h)
	}
	var spouse *Ref
	if sp := s.livingSpouse(c); sp != nil {
		spouse = &Ref{sp.ID, sp.Name}
	}
	action := c.action
	if action == "" {
		action = ActExplore
	}
	var teacher *Ref
	if c.Teacher != nil && s.tick <= c.TeacherUntil {
		teacher = c.Teacher
	}
	H := g.Hidden
	return CreatureDetail{
		SecondsPerYear: SecondsPerYear,
		ID:             c.ID,
		Name:           c.Name,
		Sex:            c.Sex.String(),
		Generation:     c.Generation,
		Mother:         c.Mother,
		Father:         c.Father,
		Spouse:         spouse,
		BornAt:         r3(math.Max(0, float64(c.BornTick)*dt)),
		Age:            r3(s.age(c)),
		Lifespan:       r3(tr.Lifespan),
		Adult:          s.adult(c),
		Energy:         r3(clamp(c.Energy, 0, 1)),
		Hydration:      r3(clamp(c.Hydration, 0, 1)),
		Health:         r3(clamp(c.Health, 0, 1)),
		Reputation:     r3(c.Reputation),
		Role:           role,
		House:          house,
		Inventory:      s.stacks(c.Inventory),
		Deeds:          c.Deeds,
		Skills:         s.skillViews(c),
		Teacher:        teacher,
		Taught:         len(c.Students),
		Pregnant:       c.Pregnancy != nil,
		Gestation:      r3(progress),
		Children:       c.Children,
		X:              r3(c.X),
		Y:              r3(c.Y),
		Action:         action,
		Traits: TraitsView{
			Hue:          r3(tr.Hue),
			Size:         r3(tr.Size),
			MaxSpeed:     r3(tr.MaxSpeed),
			Vision:       r3(tr.Vision),
			Metabolism:   r3(tr.Metabolism),
			MutationRate: r3(tr.MutationRate),
		},
		Brain: BrainView{
			InputLabels:  InputLabels,
			OutputLabels: OutputLabels,
			Size:         H,
			LearningRate: r3(tr.LearningRate),
			Learned:      learnedPerNeuron(c),
			Input:        round3(c.input[:]),
			Hidden:       round3(c.Hidden),
			Output:       round3(c.output[:]),
			WIn:          matrix(c.Mind.wIn, H),
			WRec:         matrix(g.WRec, H),
			WOut:         matrix(c.Mind.wOut, NumOutputs),
			BOut:         round3(g.BOut),
		},
	}, true
}

type ElementView struct {
	Z            int      `json:"z"`
	Symbol       string   `json:"symbol"`
	Name         string   `json:"name"`
	Category     string   `json:"category"`
	Period       int      `json:"period"`
	Group        int      `json:"group"`
	Phase        string   `json:"phase"`
	Abundance    float64  `json:"abundance"`
	Natural      bool     `json:"natural"`
	Tier         int      `json:"tier"`
	Discovered   bool     `json:"discovered"`
	DiscoveredAt *float64 `json:"discoveredAt,omitempty"`
	DiscoveredBy *Ref     `json:"discoveredBy,omitempty"`
	Era          *int     `json:"era,omitempty"`
	Source       string   `json:"source,omitempty"`
}

type TechView struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tier        int      `json:"tier"`
	Requires    []string `json:"requires"`
	Known       bool     `json:"known"`
	LearnedAt   *float64 `json:"learnedAt,omitempty"`
	LearnedBy   *Ref     `json:"learnedBy,omitempty"`
	Holders     int      `json:"holders"`
	Lost        bool     `json:"lost"`
	Written     bool     `json:"written"`
}

type StructureKindView struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	House bool        `json:"house"`
	Level int         `json:"level"`
	Tier  int         `json:"tier"`
	Cost  []StackView `json:"cost"`
	Count int         `json:"count"`
}

type Knowledge struct {
	Era            int                 `json:"era"`
	Tier           int                 `json:"tier"`
	TierName       string              `json:"tierName"`
	TierNames      []string            `json:"tierNames"`
	Elements       []ElementView       `json:"elements"`
	Techs          []TechView          `json:"techs"`
	StructureKinds []StructureKindView `json:"structureKinds"`
}

// Knowledge is everything the civilisation has found out, against everything there is.
func (s *Sim) Knowledge() Knowledge {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := Knowledge{
		Era:            s.era,
		Tier:           s.tier,
		TierName:       s.cat.tierLabel(s.tier),
		TierNames:      []string{},
		Elements:       []ElementView{},
		Techs:          []TechView{},
		StructureKinds: []StructureKindView{},
	}
	for t := 0; t <= chem.MaxTier; t++ {
		k.TierNames = append(k.TierNames, s.cat.tierLabel(t))
	}
	for _, e := range s.cat.elements {
		v := ElementView{Z: e.Z, Symbol: e.Symbol, Name: e.Name, Category: e.Category, Period: e.Period,
			Group: e.Group, Phase: e.Phase, Abundance: e.Abundance, Natural: e.Natural, Tier: e.Tier}
		if d := s.elements[e.Symbol]; d != nil {
			at, era, by := r3(d.Time), d.Era, d.By
			v.Discovered, v.DiscoveredAt, v.DiscoveredBy, v.Era, v.Source = true, &at, &by, &era, d.Source
		}
		k.Elements = append(k.Elements, v)
	}
	for _, t := range s.cat.techs {
		v := TechView{ID: t.ID, Name: t.Name, Description: t.Description, Tier: t.Tier,
			Requires: append([]string{}, t.Requires...)}
		if d := s.techs[t.ID]; d != nil {
			at, by := r3(d.Time), d.By
			v.Known, v.LearnedAt, v.LearnedBy = true, &at, &by
		}
		v.Holders = s.holderCount(t.ID)
		v.Lost = s.lost[t.ID] && !s.noCulture()
		v.Written = s.writtenLevel(t.ID) >= skillPractise
		k.Techs = append(k.Techs, v)
	}
	counts := map[string]int{}
	for _, st := range s.structures {
		counts[st.Kind]++
	}
	for _, sk := range s.cat.structures {
		cost := []StackView{}
		for _, id := range sortedKeys(sk.Cost) {
			cost = append(cost, StackView{Item: string(id), Name: s.cat.itemName(id), Qty: sk.Cost[id]})
		}
		k.StructureKinds = append(k.StructureKinds, StructureKindView{ID: sk.ID, Name: sk.Name, House: sk.House,
			Level: sk.Level, Tier: sk.Tier, Cost: cost, Count: counts[sk.ID]})
	}
	return k
}

func (s *Sim) flags(c *Creature) int {
	f := 0
	if c.Pregnancy != nil {
		f |= flagPregnant
	}
	if c.ate {
		f |= flagEating
	}
	if c.drank {
		f |= flagDrinking
	}
	if c.wantsMate && s.fertile(c) {
		f |= flagWantsMate
	}
	if c.resting {
		f |= flagResting
	}
	if !s.adult(c) {
		f |= flagChild
	}
	for i, until := range c.fx {
		if until > s.tick {
			f |= fxFlags[i]
		}
	}
	if h := s.houseOf(c); h != nil && h.Owner == c.ID {
		f |= flagHead
	}
	if c.Inventory.count() >= invCapacity/2 {
		f |= flagCarrying
	}
	return f
}

// encodeFrame writes the compact stream frame by hand: it runs 10 times a
// second for every world, so it avoids reflection.
func (s *Sim) encodeFrame() []byte {
	b := make([]byte, 0, 64+len(s.creatures)*84)
	b = append(b, `{"t":`...)
	b = strconv.AppendInt(b, s.tick, 10)
	b = append(b, `,"s":`...)
	b = strconv.AppendFloat(b, s.time(), 'f', 2, 64)
	b = append(b, `,"c":[`...)
	for i, c := range s.creatures {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, '[')
		b = strconv.AppendInt(b, c.ID, 10)
		b = append(b, ',', '"')
		b = append(b, c.Name...) // names are plain ASCII letters
		b = append(b, '"', ',')
		b = strconv.AppendFloat(b, c.X, 'f', 2, 64)
		b = append(b, ',')
		b = strconv.AppendFloat(b, c.Y, 'f', 2, 64)
		b = append(b, ',')
		b = strconv.AppendFloat(b, c.Heading, 'f', 2, 64)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(c.Sex), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(c.Genome.Traits.Hue)%360, 10)
		b = append(b, ',')
		b = strconv.AppendFloat(b, c.Genome.Traits.Size, 'f', 2, 64)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(s.flags(c)), 10)
		b = append(b, ',')
		b = strconv.AppendFloat(b, clamp(c.Energy, 0, 1), 'f', 2, 64)
		b = append(b, ',')
		b = strconv.AppendFloat(b, clamp(c.Health, 0, 1), 'f', 2, 64)
		b = append(b, ',')
		b = strconv.AppendInt(b, c.HouseID, 10)
		b = append(b, ']')
	}
	return append(b, "]}"...)
}

// Structures returns the encoded `structures` stream message and its version,
// which changes whenever any building does.
func (s *Sim) Structures() ([]byte, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b := make([]byte, 0, 32+len(s.structures)*48)
	b = append(b, `{"v":`...)
	b = strconv.AppendInt(b, s.structVersion, 10)
	b = append(b, `,"s":[`...)
	for i, st := range s.structures {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, '[')
		b = strconv.AppendInt(b, st.ID, 10)
		b = append(b, ',')
		b = strconv.AppendQuote(b, st.Kind)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(st.X), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(st.Y), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, st.Owner, 10)
		b = append(b, ',', '"')
		if st.Owner != 0 {
			b = append(b, st.OwnerName...) // names are plain ASCII letters
		}
		b = append(b, '"', ',')
		b = strconv.AppendInt(b, int64(st.kind.Level), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(st.Hue)%360, 10)
		b = append(b, ']')
	}
	return append(b, "]}"...), s.structVersion
}

// StructureVersion is cheap to poll to see whether Structures changed.
func (s *Sim) StructureVersion() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.structVersion
}

// MinedOut lists tiles whose mineral deposit has been dug out, and a version
// that changes whenever the list does.
func (s *Sim) MinedOut() (version int, tiles [][2]int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tiles = s.geo.MinedOut()
	if tiles == nil {
		tiles = [][2]int{}
	}
	return len(tiles), tiles
}
