package sim

import (
	"cmp"
	"math"
	"slices"

	"miniv2/backend/internal/chem"
)

// Two-person exchanges, after RAGE's chat task (task/Default/TaskChat.h):
// one person turns to another and waits a moment; the exchange happens only
// if the other also wants it while facing them. Talking passes on what each
// remembers about third people (gossip, after game/witness.h and
// WitnessInformation.h: second-hand, less certain; see gossip.go); trading
// swaps goods each values more than what they give (trade.go).
//
// Nothing here chooses for anyone. A request starts only when the brain's
// "bicara" or "tukar" output is above one half, the other learns of it only
// through the inTalkNear / inTradeNear senses, and the exchange takes place
// only when the other's own brain wants the same, at the same time. Nobody
// is turned or moved: who faces whom is the brains' doing.
//
// A Chat goes through three states, all saved with the person:
//   - waiting (Ended 0): asked someone and waiting for an answer, at most
//     chatWait; the one asked senses it;
//   - exchanged (Ended set, OK): talked or traded at tick Ended; for chatRest
//     they are busy (shown as talking or trading) and can't be asked again;
//   - given up (Ended set, not OK): nobody answered, or a meeting to trade
//     found nothing worth swapping; for chatRest they don't ask again, but
//     may still answer someone else.

const (
	ActTalk  Action = "talk"
	ActTrade Action = "trade"
)

// Kinds of exchange (Chat.Kind).
const (
	chatTalk  = "talk"
	chatTrade = "trade"
)

const (
	// exchangeRange: close enough to hand something over, as for a gift.
	exchangeRange = giveRange
	// exchangeFacing is the cosine of the widest angle between where someone
	// faces and the person they address or answer: ±60°. RAGE's chat turns
	// a ped to within 5° of its partner before chatting, but here nobody is
	// turned for them. People in conversation stand face to face or at an
	// angle (Kendon's F-formations: vis-à-vis, L-shaped, side by side);
	// ±60° admits the first two.
	exchangeFacing = 0.5
	// chatWait is how long an unanswered request stands: long enough for the
	// other to notice it and decide (seven or eight of their decisions at
	// 5 Hz, see cognition.go), short enough that a request no one wants
	// doesn't linger. RAGE waits up to 30 s of real time (CTaskChat
	// m_MaxWaitTime); everything here is compressed (1 year = 8 s).
	chatWait = int64(1.5 * TicksPerSecond)
	// chatRest is how long after an exchange the two are busy with it (and
	// shown talking or trading), and how long someone whose request went
	// unanswered waits before asking again.
	chatRest = int64(fxSeconds * TicksPerSecond)
	// talkAge: children speak in short sentences by two to three years.
	talkAge = 3 * SecondsPerYear
	// tradeAge: children share food from early on, but exchanging goods of
	// their own with non-kin starts around puberty: Ju/'hoansi parents set up
	// a child's hxaro exchange partners, and the young take them over in
	// their teens (Wiessner 1982). Puberty is 10 here (act.go).
	tradeAge = pubertyAge
	// exchangeSlack widens grid queries: the grid is built at the start of a
	// tick, before anyone has moved.
	exchangeSlack = 0.5
)

// Chat is a two-person exchange someone has begun and is waiting on, or
// has just had (see the states above).
type Chat struct {
	With  Ref    `json:"with"`
	Kind  string `json:"kind"`  // talk | trade
	Since int64  `json:"since"` // tick it began
	// Ended is the tick the request was answered or given up, 0 while
	// waiting; OK says the exchange took place.
	Ended int64 `json:"ended,omitempty"`
	OK    bool  `json:"ok,omitempty"`
}

func (ch *Chat) waiting() bool { return ch != nil && ch.Ended == 0 }

// ExchangeView is someone's exchange under way and their record, for the inspector.
type ExchangeView struct {
	Chat   *Chat `json:"chat,omitempty"`
	Talks  int   `json:"talks"`
	Trades int   `json:"trades"`
	Told   int   `json:"told"` // memories they only have from what others told them
}

// exchangeState is the world's saved exchange record. Only the event clocks
// affect the future (which trades and rumours are written in the log); the
// rest is for the observer.
type exchangeState struct {
	Talks    int                 `json:"talks,omitempty"`
	Trades   int                 `json:"trades,omitempty"`
	Rumors   int                 `json:"rumors,omitempty"`   // memories passed on
	Opinions int                 `json:"opinions,omitempty"` // opinions of third people passed on
	Units    int                 `json:"units,omitempty"`    // item units that changed hands in trades
	ByItem   map[chem.ItemID]int `json:"byItem,omitempty"`   // … by item
	Recent   []TradeRecord       `json:"recent,omitempty"`   // the latest trades, oldest first
	// When a trade and a rumour were last written in the event log.
	TradeEvent float64 `json:"tradeEvent,omitempty"`
	RumorEvent float64 `json:"rumorEvent,omitempty"`

	// asked indexes the waiting requests by whom they are to, for the
	// senses. Rebuilt after every exchange step and on restore, from the
	// saved chats alone; not saved.
	asked []request
}

// request is a waiting request to target, by someone.
type request struct {
	target int64
	by     *Creature
}

// ExchangeInfo summarises exchanges for the world info.
type ExchangeInfo struct {
	Talks    int `json:"talks"`
	Trades   int `json:"trades"`
	Rumors   int `json:"rumors"`
	Opinions int `json:"opinions"`
	Units    int `json:"units"`
	Waiting  int `json:"waiting"` // people waiting for an answer now
	Talking  int `json:"talking"` // people in a talk now
	Trading  int `json:"trading"` // people in a trade now
	// The latest trades (newest last) and the most traded items.
	Recent []TradeRecord `json:"recent,omitempty"`
	Items  []ItemTraded  `json:"items,omitempty"`
}

// ItemTraded is how many units of an item have changed hands.
type ItemTraded struct {
	Item  chem.ItemID `json:"item"`
	Name  string      `json:"name"`
	Units int         `json:"units"`
}

// exchange pairs up people who want to talk or trade and carries out the
// exchanges both want; called once per step after everyone has acted. It
// visits people in their stable order, so who meets whom is deterministic.
func (s *Sim) exchange() {
	if s.opts.NoExchange {
		return
	}
	for _, c := range s.creatures {
		if c.Health > 0 {
			s.exchangeStep(c)
		}
	}
	s.indexRequests()
}

// indexRequests lists the waiting requests by whom they are to. Chats change
// only in exchange, so the list holds until the next one; someone who has
// died since (and is gone from the world, as after a restore) is skipped
// where it is used.
func (s *Sim) indexRequests() {
	e := &s.exch
	clear(e.asked)
	e.asked = e.asked[:0]
	for _, c := range s.creatures {
		if c.Chat.waiting() && c.Health > 0 {
			e.asked = append(e.asked, request{c.Chat.With.ID, c})
		}
	}
	// Requests to the same person may come out in any order: the senses
	// take the closest of them.
	slices.SortFunc(e.asked, func(a, b request) int { return cmp.Compare(a.target, b.target) })
}

// exchangeStep keeps c's request going, starts one or lets it lapse, and
// holds the exchange if the one asked wants it too.
func (s *Sim) exchangeStep(c *Creature) {
	ch := c.Chat
	if ch != nil && ch.Ended > 0 {
		if s.tick < ch.Ended+chatRest {
			return
		}
		c.Chat, ch = nil, nil
	}
	if ch != nil {
		o := s.living(ch.With.ID)
		switch {
		case o == nil || c.output[chatOutput(ch.Kind)] <= 0.5 || !s.mayExchange(c, ch.Kind) ||
			distance(c, o) > exchangeRange:
			// Changed their mind, or the other has gone: perhaps a new request below.
			c.Chat, ch = nil, nil
		case s.tick-ch.Since >= chatWait:
			ch.Ended = s.tick // nobody answered
			return
		}
	}
	if ch == nil {
		// A new request is a decision, made at the brain's own rhythm.
		kind := ""
		if s.decisionDue(c) {
			kind = s.exchangeWish(c)
		}
		if kind == "" {
			return
		}
		o := s.addressee(c, kind)
		if o == nil {
			return
		}
		ch = &Chat{With: Ref{o.ID, o.Name}, Kind: kind, Since: s.tick}
		c.Chat = ch
	}
	if o := s.byID[ch.With.ID]; s.answers(o, c, ch.Kind) {
		s.meet(c, o, ch.Kind)
	}
}

// chatOutput is the brain output that wants an exchange of kind.
func chatOutput(kind string) int {
	if kind == chatTrade {
		return outTrade
	}
	return outTalk
}

func distance(a, b *Creature) float64 { return math.Hypot(b.X-a.X, b.Y-a.Y) }

// exchangeWish is the exchange c's brain wants now and c is old enough for:
// the stronger wish if both, "" if none.
func (s *Sim) exchangeWish(c *Creature) string {
	talk := c.output[outTalk] > 0.5 && s.mayExchange(c, chatTalk)
	trade := c.output[outTrade] > 0.5 && s.mayExchange(c, chatTrade)
	switch {
	case trade && (!talk || c.output[outTrade] > c.output[outTalk]):
		return chatTrade
	case talk:
		return chatTalk
	}
	return ""
}

// mayExchange reports whether c is able to take part in an exchange of kind
// now: old enough, and not resting (babies being carried rest too), working
// at a job, or in the middle of a theft or a blow.
func (s *Sim) mayExchange(c *Creature, kind string) bool {
	minAge := talkAge
	if kind == chatTrade {
		minAge = tradeAge
	}
	return c.Health > 0 && !c.resting && c.Job == nil && c.action != ActAttack && c.action != ActSteal &&
		s.age(c) >= minAge
}

// busy reports whether c is in an exchange that just took place.
func (s *Sim) busy(c *Creature) bool {
	ch := c.Chat
	return ch != nil && ch.OK && s.tick < ch.Ended+chatRest
}

// faces reports whether b stands within reach in front of a (within
// exchangeFacing of a's heading).
func faces(a, b *Creature) bool {
	dx, dy := b.X-a.X, b.Y-a.Y
	return dx*dx+dy*dy <= exchangeRange*exchangeRange && inFront(dx, dy, math.Cos(a.Heading), math.Sin(a.Heading))
}

// inFront reports whether the offset (dx, dy) lies within exchangeFacing of
// the direction (lookX, lookY). Two bodies at the very same spot touch;
// there is no "in front" then.
func inFront(dx, dy, lookX, lookY float64) bool {
	d2 := dx*dx + dy*dy
	ahead := dx*lookX + dy*lookY
	return d2 == 0 || ahead >= 0 && ahead*ahead >= d2*exchangeFacing*exchangeFacing
}

// addressee is whom c turns to: the closest person in reach in front of c,
// in sight, who could take part in an exchange of kind and isn't busy with
// another.
func (s *Sim) addressee(c *Creature, kind string) *Creature {
	var best *Creature
	bestD2 := math.Inf(1)
	lookX, lookY := math.Cos(c.Heading), math.Sin(c.Heading)
	s.grid.near(c.X, c.Y, exchangeRange+exchangeSlack, func(o *Creature) {
		dx, dy := o.X-c.X, o.Y-c.Y
		d2 := dx*dx + dy*dy
		if o == c || d2 > exchangeRange*exchangeRange || d2 >= bestD2 || o.Health <= 0 || !inFront(dx, dy, lookX, lookY) ||
			!s.mayExchange(o, kind) || s.busy(o) || !s.lineOfSight(c.X, c.Y, o.X, o.Y) {
			return
		}
		best, bestD2 = o, d2
	})
	return best
}

// answers reports whether o, asked by c, wants the same exchange with c now:
// o's brain wants that kind, o is free and in reach, and o's attention is on
// c (o's own request is to c, or o faces c).
func (s *Sim) answers(o, c *Creature, kind string) bool {
	if o == nil || o.Health <= 0 || o.output[chatOutput(kind)] <= 0.5 || !s.mayExchange(o, kind) || s.busy(o) {
		return false
	}
	if oc := o.Chat; oc.waiting() && oc.With.ID == c.ID {
		return distance(o, c) <= exchangeRange
	}
	return faces(o, c)
}

// meet carries out an exchange both want. A trade that finds nothing both
// would gain from ends without one.
func (s *Sim) meet(c, o *Creature, kind string) {
	ok, act := true, ActTalk
	switch kind {
	case chatTrade:
		ok, act = s.barter(c, o), ActTrade
	default:
		s.talk(c, o)
	}
	since := s.tick
	if oc := o.Chat; oc.waiting() && oc.With.ID == c.ID && oc.Kind == kind {
		since = oc.Since
	}
	c.Chat = &Chat{With: Ref{o.ID, o.Name}, Kind: kind, Since: c.Chat.Since, Ended: s.tick, OK: ok}
	o.Chat = &Chat{With: Ref{c.ID, c.Name}, Kind: kind, Since: since, Ended: s.tick, OK: ok}
	if !ok {
		return
	}
	c.action, o.action = act, act
	if kind == chatTrade {
		c.Deeds.Trades++
		o.Deeds.Trades++
	} else {
		c.Deeds.Talks++
		o.Deeds.Talks++
	}
}

// senseExchange fills inTalkNear and inTradeNear: how close (1 touching …
// 0 at the edge of reach) the nearest person is whose waiting request is to
// c, for each kind.
func (s *Sim) senseExchange(c *Creature, in *[NumInputs]float64) {
	if s.opts.NoExchange {
		return
	}
	asked := s.exch.asked
	i, _ := slices.BinarySearchFunc(asked, c.ID, func(r request, id int64) int { return cmp.Compare(r.target, id) })
	for ; i < len(asked) && asked[i].target == c.ID; i++ {
		o := asked[i].by
		ch := o.Chat
		if !ch.waiting() || ch.With.ID != c.ID || s.living(o.ID) != o {
			continue
		}
		d := distance(c, o)
		if d > exchangeRange {
			continue
		}
		k := inTalkNear
		if ch.Kind == chatTrade {
			k = inTradeNear
		}
		in[k] = math.Max(in[k], 1-d/exchangeRange)
	}
}

// exchangeFlags are the stream flags for someone talking or trading.
func (s *Sim) exchangeFlags(c *Creature) int {
	if s.opts.NoExchange || !s.busy(c) {
		return 0
	}
	if c.Chat.Kind == chatTrade {
		return flagTrading
	}
	return flagTalking
}

func (s *Sim) exchangeView(c *Creature) *ExchangeView {
	if s.opts.NoExchange {
		return nil
	}
	v := &ExchangeView{Talks: c.Deeds.Talks, Trades: c.Deeds.Trades}
	if c.Chat != nil {
		ch := *c.Chat
		v.Chat = &ch
	}
	for _, m := range c.Memories {
		if m.Mode == modeTold {
			v.Told++
		}
	}
	return v
}

func (s *Sim) saveExchange() *exchangeState {
	st := s.exch
	return &st
}

// restoreExchange takes back the saved record; a world saved before
// exchanges existed starts with none.
func (s *Sim) restoreExchange(st *exchangeState) {
	s.exch = exchangeState{}
	if st != nil {
		s.exch = *st
	}
	s.indexRequests()
}

func (s *Sim) exchangeInfo() ExchangeInfo {
	e := &s.exch
	info := ExchangeInfo{Talks: e.Talks, Trades: e.Trades, Rumors: e.Rumors, Opinions: e.Opinions, Units: e.Units,
		Recent: slices.Clone(e.Recent)}
	if s.opts.NoExchange {
		return info
	}
	for _, c := range s.creatures {
		switch {
		case c.Chat.waiting():
			info.Waiting++
		case s.busy(c) && c.Chat.Kind == chatTrade:
			info.Trading++
		case s.busy(c):
			info.Talking++
		}
	}
	for id, n := range e.ByItem {
		info.Items = append(info.Items, ItemTraded{id, s.cat.itemName(id), n})
	}
	slices.SortFunc(info.Items, func(a, b ItemTraded) int {
		return cmp.Or(b.Units-a.Units, cmp.Compare(a.Item, b.Item))
	})
	if len(info.Items) > 8 {
		info.Items = info.Items[:8]
	}
	return info
}
