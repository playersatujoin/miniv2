package sim

import (
	"math"
	"slices"
	"strconv"
)

// Stimuli: striking things that stay in the world for a while and can be
// seen or heard by whoever is near, after RAGE's shocking events
// (game/event/EventShocking.h, ShockingEvents.h): a body, a fight, a theft,
// a predator's kill, a fire, a drowning, a fall, a house falling in. Each
// kind has its own shock level, sight and hearing range and lifetime; near
// duplicates merge. People only sense them (inShock, inShockSide); what to do
// about them is up to their brains.
//
// Like RAGE's global event queue (fwevent/EventQueue.h) the list is bounded:
// a new stimulus close to a live one of the same kind refreshes it instead
// of adding another, and when the list is full the least striking kind gives
// way, or the newcomer is dropped if nothing ranks below it. A grid of
// cells, each listing the stimuli that can reach into it, keeps sensing
// cheap: a person looks only at those listed for where they stand.

// Stimulus kinds other files raise.
const (
	stimCorpse   = "corpse"   // a dead body (die)
	stimFight    = "fight"    // an assault (perceiveEvent)
	stimTheft    = "theft"    // a theft (perceiveEvent)
	stimPredator = "predator" // an animal attacking a person
	stimFire     = "fire"     // a burning tile (fire.go)
	stimCollapse = "collapse" // a building destroyed (fire.go)
	stimDrowning = "drowning" // someone in trouble in the water (locomotion.go)
	stimFall     = "fall"     // someone falling on a slope (locomotion.go)
	stimCarcass  = "carcass"  // meat left at a kill, "bangkai buruan" (sharing.go)
)

// Stimulus is one striking thing left in the world.
type Stimulus struct {
	ID    int64   `json:"id"`
	Kind  string  `json:"kind"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Tick  int64   `json:"tick"`  // when it happened
	Until int64   `json:"until"` // when it fades
	Level float64 `json:"level"` // how striking, 0–1
	Actor *Ref    `json:"actor,omitempty"`
	// Body is how a dead person looked, so the observer can draw them.
	Body *CorpseBody `json:"body,omitempty"`
}

// CorpseBody is what an observer needs to draw a body where it fell.
type CorpseBody struct {
	ID      int64   `json:"id"`
	Sex     Sex     `json:"sex"`
	Hue     int     `json:"hue"`
	Size    float64 `json:"size"`
	Age     float64 `json:"age"` // years
	Heading float64 `json:"heading"`
	Cause   string  `json:"cause"`
}

// stimulusState is the saved list of stimuli.
type stimulusState struct {
	List []*Stimulus `json:"list,omitempty"`
	Next int64       `json:"next,omitempty"`
}

// StimuliInfo summarises stimuli for the world info.
type StimuliInfo struct {
	Active  int            `json:"active"`
	Corpses int            `json:"corpses"`
	ByKind  map[string]int `json:"byKind,omitempty"`
}

// stimKind is what every stimulus of a kind shares, after the tunables of
// RAGE's shocking events (EventShocking.h).
type stimKind struct {
	shock   float64 // how striking at level 1, 0–1
	fear    float64 // how frightening that shock is, 0–1 (m_PedFearImpact)
	sight   float64 // tiles it can be noticed from by eye (m_VisualReactionRange); the eye's own reach limits it further
	hearing float64 // tiles it can be heard from, 0 if silent (m_AudioReactionRange)
	life    float64 // seconds it stays (m_LifeTime)
	fresh   float64 // seconds over which its shock fades to nothing (life if 0)
	merge   float64 // tiles within which a new one refreshes a live one (m_DuplicateDistanceCheck); 0 never merges
	// priority decides which kinds give way when the list is full, after
	// RAGE's shock levels (eSEventShockLevels: interesting 0, affects others
	// 1, potentially dangerous 2, dangerous 3, serious danger 4).
	priority int
	glow     bool // seen by its own light from as far as sight reaches, by night as by day
}

// The kinds. A tile is about 4 m (ecology/hydrology.go), so a fight is heard
// some 30 m off, as perceiveEvent hears an assault, screams and crashes
// further, a fire's crackle less, and the eye reaches 12–36 m. How long each
// lasts is an estimate on the scale of actions (a blow a second), kept short
// because a second is also some 45 days of the world's calendar: a theft is
// over in a moment, a brawl in a couple of blows.
//
// A body stays four seconds, half a simulated year. In the open tropics soft
// tissue is gone within weeks, but remains stay recognisable for months
// where nobody buries them, and four seconds is what an observer watching at
// 1× needs to notice one. Only a fresh body shocks: that fades over the first
// two seconds. A body upsets more than it frightens; a fight, a predator or
// a fire is what makes people afraid.
var stimKinds = map[string]*stimKind{
	stimCorpse:   {shock: 0.8, fear: 0.25, sight: 9, life: 4, fresh: 2, priority: 2},
	stimFight:    {shock: 0.7, fear: 0.6, sight: 9, hearing: 8, life: 2, merge: 2, priority: 3},
	stimTheft:    {shock: 0.35, fear: 0.1, sight: 6, life: 1, merge: 1.5, priority: 1},
	stimPredator: {shock: 1, fear: 1, sight: 9, hearing: 10, life: 2, merge: 3, priority: 4},
	stimFire:     {shock: 0.9, fear: 0.8, sight: 12, hearing: 4, life: 2, merge: 3, priority: 3, glow: true},
	stimCollapse: {shock: 0.7, fear: 0.5, sight: 9, hearing: 10, life: 2, merge: 2, priority: 2},
	stimDrowning: {shock: 0.8, fear: 0.5, sight: 9, hearing: 6, life: 2, merge: 2, priority: 3},
	stimFall:     {shock: 0.5, fear: 0.3, sight: 9, hearing: 5, life: 1, merge: 1.5, priority: 1},
	// A kill draws the eye (and the crows) without frightening anyone; it
	// stays about as long as most of its meat (sharing.go).
	stimCarcass: {shock: 0.3, sight: 9, life: 4, merge: 1.5, priority: 0},
}

func init() {
	for _, k := range stimKinds {
		if k.fresh <= 0 || k.fresh > k.life {
			k.fresh = k.life
		}
	}
}

const (
	// maxStimuli bounds the list. A live world of ~900 people keeps a few
	// dozen (bodies, brawls and thefts); a wide wildfire merges into one
	// stimulus every few tiles.
	maxStimuli = 256
	// stimCell is the side (tiles) of a cell of the sensing grid: small
	// enough that a crowded village's cells list only what can reach them.
	stimCell = 4.0
	// naturalShock is how striking a body is after a quiet death (old age,
	// illness, hunger) relative to a violent one.
	naturalShock = 0.75
)

// stimWorld is a world's stimuli and the sensing grid built over them.
type stimWorld struct {
	list []*Stimulus
	next int64
	// expiry is the earliest tick a stimulus fades (0: none to check).
	expiry int64

	// The sensing grid: entries[i] is list[i] laid out for a quick look,
	// cells[c] the entries reaching into cell c (in list order), touched the
	// cells in use. dirty means it must be rebuilt before it is used (after a
	// removal, a restore or a new map size).
	entries []stimEntry
	cells   [][]int32
	touched []int32
	cw, ch  int
	dirty   bool

	// The largest vision of anyone alive, worked out once a tick for the
	// witness search (perception.go).
	visionTick int64
	vision     float64
}

// stims returns s's stimuli.
func (s *Sim) stims() *stimWorld { return &s.stimuli }

// stimEntry is a stimulus as the sensing grid holds it.
type stimEntry struct {
	x, y   float64
	reach2 float64   // squared reach
	k      *stimKind // nil for a kind this version doesn't know
	st     *Stimulus
}

// reach is how far a stimulus of kind k can be perceived at all.
func (k *stimKind) reach() float64 { return math.Max(k.sight, k.hearing) }

// priorityOf is a stimulus's rank when the list is full.
func priorityOf(st *Stimulus) int {
	if k := stimKinds[st.Kind]; k != nil {
		return k.priority
	}
	return -1
}

// addStimulus leaves something striking of kind at (x, y); level scales the
// kind's usual shock (1 = as usual). actor may be nil.
func (s *Sim) addStimulus(kind string, x, y, level float64, actor *Creature) {
	s.putStimulus(kind, x, y, level, actor, nil)
}

// putStimulus adds (or merges) a stimulus and returns it, or nil if it was
// dropped.
func (s *Sim) putStimulus(kind string, x, y, level float64, actor *Creature, body *CorpseBody) *Stimulus {
	k := stimKinds[kind]
	if s.opts.NoStimuli || k == nil || !(level > 0) || math.IsNaN(x) || math.IsNaN(y) {
		return nil
	}
	w := s.stims()
	level = clamp(k.shock*level, 0, 1)
	until := s.tick + int64(math.Round(k.life*TicksPerSecond))
	var ref *Ref
	if actor != nil {
		ref = &Ref{actor.ID, actor.Name}
	}
	if k.merge > 0 && body == nil {
		if old := s.duplicate(w, kind, x, y, k.merge); old != nil {
			// Refreshed where it first happened (so the sensing grid stays
			// right); as fresh and as strong as the stronger of the two.
			old.Tick, old.Until = s.tick, until
			old.Level = math.Max(old.Level, level)
			if ref != nil {
				old.Actor = ref
			}
			w.expiry = minExpiry(w.expiry, until)
			return old
		}
	}
	if len(w.list) >= maxStimuli && !s.stimRoom(w, k.priority) {
		return nil
	}
	w.next++
	st := &Stimulus{ID: w.next, Kind: kind, X: x, Y: y, Tick: s.tick, Until: until, Level: level, Actor: ref, Body: body}
	w.list = append(w.list, st)
	w.expiry = minExpiry(w.expiry, until)
	if !w.dirty {
		s.register(w, st)
	}
	return st
}

func minExpiry(cur, until int64) int64 {
	if cur == 0 || until < cur {
		return until
	}
	return cur
}

// duplicate finds a live stimulus of the same kind within merge tiles of (x, y).
func (s *Sim) duplicate(w *stimWorld, kind string, x, y, merge float64) *Stimulus {
	s.indexStimuli(w)
	for _, i := range w.cells[s.stimCellOf(w, x, y)] {
		st := w.entries[i].st
		if st.Kind == kind && st.Until > s.tick && st.Body == nil &&
			math.Abs(st.X-x) <= merge && math.Abs(st.Y-y) <= merge && math.Hypot(st.X-x, st.Y-y) <= merge {
			return st
		}
	}
	return nil
}

// stimRoom frees a place in a full list: faded stimuli first, else the least
// striking kind (the oldest of them) if it ranks below priority.
func (s *Sim) stimRoom(w *stimWorld, priority int) bool {
	if s.expireStimuli(w) {
		return true
	}
	low, lowP := -1, 0
	for i, st := range w.list {
		if p := priorityOf(st); low < 0 || p < lowP {
			low, lowP = i, p
		}
	}
	if low < 0 || lowP >= priority {
		return false
	}
	w.list = slices.Delete(w.list, low, low+1)
	w.dirty = true
	return true
}

// expireStimuli drops faded stimuli and reports whether any went.
func (s *Sim) expireStimuli(w *stimWorld) bool {
	if w.expiry == 0 || s.tick < w.expiry {
		return false
	}
	kept := w.list[:0]
	w.expiry = 0
	for _, st := range w.list {
		if st.Until > s.tick {
			kept = append(kept, st)
			w.expiry = minExpiry(w.expiry, st.Until)
		}
	}
	gone := len(w.list) - len(kept)
	clear(w.list[len(kept):])
	w.list = kept
	if gone > 0 {
		w.dirty = true
	}
	return gone > 0
}

// stimuliTick expires old stimuli; called once per step.
func (s *Sim) stimuliTick() {
	w := s.stims()
	if s.opts.NoStimuli {
		if len(w.list) > 0 {
			clear(w.list)
			w.list, w.expiry, w.dirty = w.list[:0], 0, true
		}
		return
	}
	s.expireStimuli(w)
}

// --- the sensing grid --------------------------------------------------------

func (s *Sim) stimCellOf(w *stimWorld, x, y float64) int {
	cx := min(w.cw-1, max(0, int(x/stimCell)))
	cy := min(w.ch-1, max(0, int(y/stimCell)))
	return cy*w.cw + cx
}

// indexStimuli rebuilds the sensing grid if it is out of date.
func (s *Sim) indexStimuli(w *stimWorld) {
	cw := max(1, int(math.Ceil(float64(s.terrain.w)/stimCell)))
	ch := max(1, int(math.Ceil(float64(s.terrain.h)/stimCell)))
	if w.cw != cw || w.ch != ch {
		w.cw, w.ch = cw, ch
		w.cells = make([][]int32, cw*ch)
		w.touched = w.touched[:0]
		w.dirty = true
	}
	if !w.dirty {
		return
	}
	for _, c := range w.touched {
		w.cells[c] = w.cells[c][:0]
	}
	w.touched = w.touched[:0]
	clear(w.entries)
	w.entries = w.entries[:0]
	w.dirty = false
	for _, st := range w.list {
		s.register(w, st)
	}
}

// register adds the newest stimulus to the sensing grid, in every cell its
// reach touches.
func (s *Sim) register(w *stimWorld, st *Stimulus) {
	if w.cells == nil {
		w.dirty = true // no grid yet: built on first use
		return
	}
	k := stimKinds[st.Kind]
	e := stimEntry{x: st.X, y: st.Y, k: k, st: st}
	w.entries = append(w.entries, e)
	if k == nil {
		return // a kind this version doesn't know reaches nobody
	}
	i := int32(len(w.entries) - 1)
	r := k.reach()
	w.entries[i].reach2 = r * r
	x0, x1 := max(0, int((st.X-r)/stimCell)), min(w.cw-1, int((st.X+r)/stimCell))
	y0, y1 := max(0, int((st.Y-r)/stimCell)), min(w.ch-1, int((st.Y+r)/stimCell))
	for cy := y0; cy <= y1; cy++ {
		for cx := x0; cx <= x1; cx++ {
			c := cy*w.cw + cx
			if len(w.cells[c]) == 0 {
				w.touched = append(w.touched, int32(c))
			}
			w.cells[c] = append(w.cells[c], i)
		}
	}
}

// --- sensing -----------------------------------------------------------------

// perceivedShock is the strongest stimulus c perceives now: how striking it
// is (0–1), on which side (sin of its bearing from c's heading: -1 left …
// 1 right), and how frightening the most frightening one is (0–1).
//
// A stimulus is seen within the eye's reach by the light of the day (or by
// its own, for a fire) and the kind's sight range, inside the field of view
// and with nothing solid in between; it is heard within the kind's hearing
// range, half as clearly through a wall. Seen, it fades to half at the edge
// of sight; heard, to nothing at the edge of hearing; both fade as it ages.
// The one who caused it isn't struck by it.
func (s *Sim) perceivedShock(c *Creature) (shock, side, fear float64) {
	if s.opts.NoStimuli {
		return 0, 0, 0
	}
	w := s.stims()
	if len(w.list) == 0 {
		return 0, 0, 0
	}
	s.indexStimuli(w)
	cands := w.cells[s.stimCellOf(w, c.X, c.Y)]
	if len(cands) == 0 {
		return 0, 0, 0
	}
	looked := false
	var eye, lookX, lookY, bestX, bestY float64
	for _, i := range cands {
		e := &w.entries[i]
		dx, dy := e.x-c.X, e.y-c.Y
		d2 := dx*dx + dy*dy
		if d2 > e.reach2 {
			continue
		}
		st, k := e.st, e.k
		if st.Until <= s.tick || st.Actor != nil && st.Actor.ID == c.ID {
			continue
		}
		base := st.Level * (1 - float64(s.tick-st.Tick)/(k.fresh*TicksPerSecond))
		if base <= 0 || base <= shock && base*k.fear <= fear {
			continue // faded, or can't beat what is already perceived
		}
		if !looked {
			looked = true
			eye = c.Genome.Traits.Vision * (0.55 + 0.45*s.eco.Light())
			lookX, lookY = math.Cos(c.Heading), math.Sin(c.Heading)
		}
		see := k.sight
		if !k.glow {
			see = math.Min(see, eye)
		}
		d := math.Sqrt(d2)
		inView := d <= see && (s.opts.NoPerception || d <= 0.8 || dx*lookX+dy*lookY >= d*0.2588190451)
		seen, heard := 0.0, 0.0
		if inView {
			seen = base * (1 - 0.5*d/see)
		}
		if d < k.hearing {
			heard = base * (1 - d/k.hearing)
		}
		if top := math.Max(seen, heard); top <= 0 || top <= shock && top*k.fear <= fear {
			continue
		}
		// Sight and sound share one look along the line (skipped without
		// perception, as for every other sense).
		if !s.opts.NoPerception && !s.lineOfSight(c.X, c.Y, e.x, e.y) {
			seen, heard = 0, heard*0.5
		}
		strength := math.Max(seen, heard)
		if strength > shock {
			shock, bestX, bestY = strength, dx, dy
		}
		fear = math.Max(fear, strength*k.fear)
	}
	if shock > 0 {
		if d := math.Hypot(bestX, bestY); d > 1e-9 {
			side = (lookX*bestY - lookY*bestX) / d // sin of the bearing; +y is to the right of +x
		}
	}
	return shock, side, fear
}

// senseStimuli fills inShock and inShockSide. The fear of what was
// perceived is kept for the moods, which are brought up to date later in the
// same tick (affectStep).
func (s *Sim) senseStimuli(c *Creature, in *[NumInputs]float64) {
	if s.opts.NoStimuli {
		return
	}
	shock, side, fear := s.perceivedShock(c)
	in[inShock] = shock
	in[inShockSide] = side
	if !s.opts.NoAffect {
		c.Affect.sensedAt, c.Affect.sensedFear = s.tick, fear
	}
}

// --- bodies ------------------------------------------------------------------

// violentDeath reports whether a death leaves a body more striking than a
// quiet one.
func violentDeath(cause string) bool {
	switch cause {
	case "killed", "animal", "burned", "drowned", "fall":
		return true
	}
	return false
}

// leaveCorpse marks where c died (called from die before c is forgotten).
func (s *Sim) leaveCorpse(c *Creature, cause string) {
	if s.opts.NoStimuli {
		return
	}
	level := 1.0
	if !violentDeath(cause) {
		level = naturalShock
	}
	body := &CorpseBody{
		ID:      c.ID,
		Sex:     c.Sex,
		Hue:     int(int64(c.Genome.Traits.Hue) % 360),
		Size:    c.Genome.Traits.Size,
		Age:     s.age(c) / SecondsPerYear,
		Heading: c.Heading,
		Cause:   cause,
	}
	s.putStimulus(stimCorpse, c.X, c.Y, level, nil, body)
}

// appendCorpses writes the frame's bodies, `,"d":[[id,x,y,heading,sex,hue,size,age,"cause",seconds],…]`.
func (s *Sim) appendCorpses(b []byte, view *Viewport) []byte {
	if s.opts.NoStimuli {
		return b
	}
	first := true
	for _, st := range s.stims().list {
		body := st.Body
		if body == nil || st.Until <= s.tick || !view.contains(st.X, st.Y) {
			continue
		}
		if first {
			b = append(b, `,"d":[`...)
			first = false
		} else {
			b = append(b, ',')
		}
		b = append(b, '[')
		b = strconv.AppendInt(b, body.ID, 10)
		b = append(b, ',')
		b = strconv.AppendFloat(b, st.X, 'f', 2, 64)
		b = append(b, ',')
		b = strconv.AppendFloat(b, st.Y, 'f', 2, 64)
		b = append(b, ',')
		b = strconv.AppendFloat(b, body.Heading, 'f', 2, 64)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(body.Sex), 10)
		b = append(b, ',')
		b = strconv.AppendInt(b, int64(body.Hue), 10)
		b = append(b, ',')
		b = strconv.AppendFloat(b, body.Size, 'f', 2, 64)
		b = append(b, ',')
		b = strconv.AppendFloat(b, body.Age, 'f', 1, 64)
		b = append(b, ',')
		b = strconv.AppendQuote(b, body.Cause)
		b = append(b, ',')
		b = strconv.AppendFloat(b, float64(s.tick-st.Tick)*dt, 'f', 1, 64)
		b = append(b, ']')
	}
	if !first {
		b = append(b, ']')
	}
	return b
}

// --- save, restore, info -----------------------------------------------------

func (s *Sim) saveStimuli() *stimulusState {
	w := s.stims()
	if len(w.list) == 0 && w.next == 0 {
		return nil
	}
	return &stimulusState{List: w.list, Next: w.next}
}

func (s *Sim) restoreStimuli(st *stimulusState) {
	w := s.stims()
	clear(w.list)
	w.list, w.next, w.expiry, w.dirty = w.list[:0], 0, 0, true
	if st == nil {
		return
	}
	w.next = st.Next
	for _, x := range st.List {
		if x != nil {
			w.list = append(w.list, x)
			w.expiry = minExpiry(w.expiry, x.Until)
		}
	}
}

func (s *Sim) stimuliInfo() StimuliInfo {
	var info StimuliInfo
	for _, st := range s.stims().list {
		if st.Until <= s.tick {
			continue
		}
		info.Active++
		if st.Body != nil {
			info.Corpses++
		}
		if info.ByKind == nil {
			info.ByKind = map[string]int{}
		}
		info.ByKind[st.Kind]++
	}
	return info
}
