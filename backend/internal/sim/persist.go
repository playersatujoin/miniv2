package sim

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/ecology"
	"miniv2/backend/internal/world"
)

// v6 (Fase 2): the living land (climate, plants, fields, fish, animals) and
// brains with senses for animals, seasons and crops. Older saves can't be
// read; the world is backed up and starts over.
const stateVersion = 6

// state is the on-disk form of a Sim.
type state struct {
	Version       int                   `json:"version"`
	Width         int                   `json:"width"`
	Height        int                   `json:"height"`
	Tick          int64                 `json:"tick"`
	Speed         int                   `json:"speed"`
	NextID        int64                 `json:"nextId"`
	NextEvent     int64                 `json:"nextEvent"`
	NextStructID  int64                 `json:"nextStructId"`
	StructVersion int64                 `json:"structVersion"`
	Births        int                   `json:"births"`
	Deaths        int                   `json:"deaths"`
	Era           int                   `json:"era"`
	DeathsBy      deathCounts           `json:"deathsBy"`
	Crimes        int                   `json:"crimes"`
	Kindness      int                   `json:"kindness"`
	Kills         int                   `json:"kills"`
	Milestone     int                   `json:"milestone"`
	Rng           []byte                `json:"rng"`
	Creatures     []*Creature           `json:"creatures"`
	Structures    []*Structure          `json:"structures"`
	Elements      map[string]*Discovery `json:"elements"`
	Techs         map[string]*Discovery `json:"techs"`
	Eco           *ecology.State        `json:"eco"`
	EcoHistory    []EcoPoint            `json:"ecoHistory,omitempty"`
	CapHits       int                   `json:"capHits,omitempty"`
	Deposits      []float32             `json:"deposits"`
	TileItems     []byte                `json:"tileItems"`
	History       []HistoryPoint        `json:"history"`
	Events        []Event               `json:"events"`
	Archive       []archived            `json:"archive"`
	Options       Options               `json:"options"`
	Stats         *demography           `json:"stats,omitempty"`
	Lost          map[string]bool       `json:"lost,omitempty"`
	LastHolder    map[string]Ref        `json:"lastHolder,omitempty"`
	KnowledgeLost int                   `json:"knowledgeLost,omitempty"`
	EventClock    *eventClock           `json:"eventClock,omitempty"`
}

// eventClock is when each throttled kind of event was last reported, so a
// restored world writes exactly the log the original would have.
type eventClock struct {
	Eco   map[string]float64 `json:"eco,omitempty"`
	Learn float64            `json:"learn"`
	Kind  float64            `json:"kind"`
	Crime float64            `json:"crime"`
}

// MarshalState serialises the whole world, including every creature's genome
// and brain memory and the random generator, so it resumes exactly. The
// result is gzip-compressed JSON.
func (s *Sim) MarshalState() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rng, err := s.src.MarshalBinary()
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(state{
		Version:       stateVersion,
		Width:         s.terrain.w,
		Height:        s.terrain.h,
		Tick:          s.tick,
		Speed:         s.speed,
		NextID:        s.nextID,
		NextEvent:     s.nextEvent,
		NextStructID:  s.nextStructID,
		StructVersion: s.structVersion,
		Births:        s.births,
		Deaths:        s.deaths,
		Era:           s.era,
		DeathsBy:      s.deathsBy,
		Crimes:        s.crimes,
		Kindness:      s.kindness,
		Kills:         s.kills,
		Milestone:     s.milestone,
		Rng:           rng,
		Creatures:     s.creatures,
		Structures:    s.structures,
		Elements:      s.elements,
		Techs:         s.techs,
		Eco:           s.eco.State(),
		EcoHistory:    s.ecoHistory,
		CapHits:       s.capHits,
		Deposits:      s.geo.Amounts(),
		TileItems:     s.tileItems,
		History:       s.history,
		Events:        s.events,
		Archive:       s.archive,
		Options:       s.opts,
		Stats:         s.stats,
		Lost:          s.lost,
		LastHolder:    s.lastHolder,
		KnowledgeLost: s.knowledgeLost,
		EventClock:    &eventClock{Eco: s.lastEcoEvt, Learn: s.lastLearnEvent, Kind: s.lastKindEvent, Crime: s.lastCrimeEvent},
	})
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(data); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Restore rebuilds a Sim saved by MarshalState on top of the (possibly
// edited) map.
func Restore(m *world.Map, data []byte) (*Sim, error) {
	return restoreWith(m, data, defaultCatalog())
}

func restoreWith(m *world.Map, data []byte, cat *catalog) (*Sim, error) {
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("not a v%d world save: %w", stateVersion, err)
	}
	raw, err := io.ReadAll(zr)
	if err != nil {
		return nil, err
	}
	var st state
	if err := json.Unmarshal(raw, &st); err != nil {
		return nil, err
	}
	if st.Version != stateVersion {
		return nil, fmt.Errorf("unsupported sim state version %d", st.Version)
	}
	if st.Width != m.Width || st.Height != m.Height {
		return nil, fmt.Errorf("saved world is %dx%d, map is %dx%d", st.Width, st.Height, m.Width, m.Height)
	}

	s := newSimWith(m, 0, cat)
	if err := s.src.UnmarshalBinary(st.Rng); err != nil {
		return nil, err
	}
	if st.Eco == nil {
		return nil, fmt.Errorf("save has no ecology")
	}
	eco, err := ecology.Restore(ecology.LandFromMap(m), st.Eco, s.rng, st.Options.ecology())
	if err != nil {
		return nil, err
	}
	s.eco = eco
	s.ecoHistory = st.EcoHistory
	s.capHits = st.CapHits
	if len(st.Deposits) > 0 {
		if err := s.geo.SetAmounts(st.Deposits); err != nil {
			return nil, err
		}
	}
	for _, c := range st.Creatures {
		g := c.Genome
		if g == nil || !g.valid() || len(c.Hidden) != g.Hidden || c.Mind == nil || !c.Mind.valid(g) {
			return nil, fmt.Errorf("creature %d has a malformed brain", c.ID)
		}
		if p := c.Pregnancy; p != nil && (p.FatherGenome == nil || !p.FatherGenome.valid()) {
			return nil, fmt.Errorf("creature %d carries a malformed genome", c.ID)
		}
		c.Mind.rebuild(g)
		s.add(c)
	}
	for _, a := range st.Archive {
		if a.Genome == nil || !a.Genome.valid() {
			return nil, fmt.Errorf("archive holds a malformed genome")
		}
	}
	for _, b := range st.Structures {
		k, ok := cat.structure[b.Kind]
		if !ok {
			k = chem.StructureKind{ID: b.Kind, Name: b.Kind}
		}
		b.kind = k
		s.indexStructure(b)
	}
	if st.Elements != nil {
		s.elements = st.Elements
	}
	if st.Techs != nil {
		s.techs = st.Techs
	}
	s.tick = st.Tick
	s.speed = st.Speed
	s.nextID = st.NextID
	s.nextEvent = st.NextEvent
	s.nextStructID = st.NextStructID
	s.structVersion = st.StructVersion
	s.births = st.Births
	s.deaths = st.Deaths
	s.era = st.Era
	s.deathsBy = st.DeathsBy
	s.crimes = st.Crimes
	s.kindness = st.Kindness
	s.kills = st.Kills
	s.milestone = st.Milestone
	s.history = st.History
	s.events = st.Events
	s.archive = st.Archive
	s.opts = st.Options
	if st.Stats != nil {
		s.stats = st.Stats
	}
	if st.Lost != nil {
		s.lost = st.Lost
	}
	if st.LastHolder != nil {
		s.lastHolder = st.LastHolder
	}
	s.knowledgeLost = st.KnowledgeLost
	if ec := st.EventClock; ec != nil {
		if ec.Eco != nil {
			s.lastEcoEvt = ec.Eco
		}
		s.lastLearnEvent, s.lastKindEvent, s.lastCrimeEvent = ec.Learn, ec.Kind, ec.Crime
	}
	s.removeBlockedStructures()
	s.applyFarms()
	s.relocateStranded()
	if len(st.TileItems) == len(s.tileItems) {
		copy(s.tileItems, st.TileItems)
	} else {
		s.refreshResources()
	}
	s.recountHolders()
	s.frames.publish(s.encodeFrame())
	return s, nil
}
