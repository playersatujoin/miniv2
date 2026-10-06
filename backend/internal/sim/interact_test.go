package sim

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"strconv"
	"testing"
	"time"

	"miniv2/backend/internal/world"
)

// facePair puts two grown-ups a tile apart, facing each other.
func facePair(t *testing.T) (*Sim, *Creature, *Creature) {
	t.Helper()
	s, _, x, y := fakeWorld(t)
	a := person(s, Female, "Sari", x, y)
	b := person(s, Male, "Budi", x+1, y)
	a.Heading, b.Heading = 0, math.Pi
	s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures)
	return s, a, b
}

// exchangeTick moves the clock on a tick and runs only the exchanges, with
// whatever outputs the test has set.
func exchangeTick(s *Sim) {
	s.tick++
	s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures)
	s.exchange()
}

func exchangeSense(s *Sim, c *Creature) (talk, trade float64) {
	var in [NumInputs]float64
	s.senseExchange(c, &in)
	return in[inTalkNear], in[inTradeNear]
}

func TestExchangeNeedsBothToWant(t *testing.T) {
	s, a, b := facePair(t)
	a.output[outTalk] = 0.9
	exchangeTick(s)
	if !a.Chat.waiting() || a.Chat.With.ID != b.ID || a.Chat.Kind != chatTalk || a.Chat.Since != s.tick {
		t.Fatalf("Sari should ask Budi to talk: %+v", a.Chat)
	}
	if a.Deeds.Talks != 0 || b.Deeds.Talks != 0 || s.exch.Talks != 0 || s.exchangeFlags(a) != 0 {
		t.Fatal("a talk took place with only one wanting it")
	}
	if talk, trade := exchangeSense(s, b); math.Abs(talk-(1-1/exchangeRange)) > 1e-9 || trade != 0 {
		t.Fatalf("Budi should sense the request a tile away: talk %v trade %v", talk, trade)
	}
	if talk, _ := exchangeSense(s, a); talk != 0 {
		t.Fatal("the one asking senses no request to them")
	}

	// Budi wants something else: still nothing.
	b.output[outTrade] = 0.9
	exchangeTick(s)
	if a.Deeds.Talks+b.Deeds.Talks+a.Deeds.Trades+b.Deeds.Trades != 0 {
		t.Fatal("a talk and a trade request are not the same exchange")
	}
	if _, trade := exchangeSense(s, a); trade <= 0 {
		t.Fatal("Sari should sense Budi's trade request")
	}

	// Now both want to talk.
	b.output[outTrade], b.output[outTalk] = 0, 0.9
	exchangeTick(s)
	if a.Deeds.Talks != 1 || b.Deeds.Talks != 1 || s.exch.Talks != 1 {
		t.Fatalf("no talk: %+v %+v", a.Deeds, b.Deeds)
	}
	if a.action != ActTalk || b.action != ActTalk || s.exchangeFlags(a) != flagTalking || s.exchangeFlags(b) != flagTalking {
		t.Fatal("the two should be seen talking")
	}
	if !a.Chat.OK || a.Chat.With.ID != b.ID || !b.Chat.OK || b.Chat.With.ID != a.ID {
		t.Fatalf("chats not closed: %+v %+v", a.Chat, b.Chat)
	}
	// Talking builds a little trust both ways.
	if s.opinion(a, b) <= 0 || s.opinion(b, a) <= 0 {
		t.Fatal("no trust from talking")
	}
	if info := s.exchangeInfo(); info.Talks != 1 || info.Talking != 2 {
		t.Fatalf("info %+v", info)
	}
	// Busy for a moment: no second talk until the rest is over, then shown no more.
	for range chatRest - 1 {
		exchangeTick(s)
	}
	if a.Deeds.Talks != 1 || s.exchangeFlags(a) != flagTalking {
		t.Fatal("talked again while still busy, or stopped showing too early")
	}
	exchangeTick(s)
	if a.Deeds.Talks != 2 {
		t.Fatalf("after the rest the two may talk again: %d", a.Deeds.Talks)
	}
}

func TestChatRequestExpires(t *testing.T) {
	s, a, b := facePair(t)
	a.output[outTalk] = 0.9
	exchangeTick(s)
	since := a.Chat.Since
	for s.tick < since+chatWait-1 {
		exchangeTick(s)
	}
	if !a.Chat.waiting() || a.Chat.Since != since {
		t.Fatalf("request lapsed early: %+v", a.Chat)
	}
	exchangeTick(s)
	if a.Chat.waiting() || a.Chat.OK || a.Chat.Ended != s.tick {
		t.Fatalf("unanswered request should lapse after %d ticks: %+v", chatWait, a.Chat)
	}
	if talk, _ := exchangeSense(s, b); talk != 0 {
		t.Fatal("a lapsed request is not sensed")
	}
	for range chatRest - 1 {
		exchangeTick(s)
		if a.Chat.waiting() {
			t.Fatal("asked again straight after giving up")
		}
	}
	exchangeTick(s)
	if !a.Chat.waiting() || a.Chat.Since != s.tick {
		t.Fatalf("should ask again after a rest: %+v", a.Chat)
	}
	if b.Deeds.Talks != 0 {
		t.Fatal("Budi never wanted to talk")
	}

	// Walking away, or no longer wanting to, ends a request at once.
	a.output[outTalk] = 0.2
	exchangeTick(s)
	if a.Chat != nil {
		t.Fatalf("request kept after changing her mind: %+v", a.Chat)
	}
}

func TestChatHeadingTolerance(t *testing.T) {
	s, a, b := facePair(t)
	a.output[outTalk] = 0.9
	for _, c := range []struct {
		deg  float64
		asks bool
	}{{180, false}, {50, true}, {-50, true}, {70, false}, {-90, false}, {0, true}} {
		a.Chat = nil
		a.Heading = c.deg * math.Pi / 180
		exchangeTick(s)
		if got := a.Chat.waiting(); got != c.asks {
			t.Fatalf("heading %v° away from Budi: asks %v, want %v", c.deg, got, c.asks)
		}
	}

	// Budi wants to talk but looks the other way, with nobody in front of
	// him: he doesn't answer.
	a.Chat, a.Heading = nil, 0
	b.Heading = 0
	b.output[outTalk] = 0.9
	exchangeTick(s)
	if a.Deeds.Talks != 0 || b.Chat != nil {
		t.Fatalf("answered with his back turned: %+v", b.Chat)
	}
	b.Heading = math.Pi * 0.8 // ±36° off: close enough
	exchangeTick(s)
	if a.Deeds.Talks != 1 || b.Deeds.Talks != 1 {
		t.Fatal("should talk once he turns to her")
	}
}

func TestWhoCanExchange(t *testing.T) {
	s, a, b := facePair(t)
	// A five-year-old talks but doesn't trade.
	b.BornTick = s.tick - int64(5*SecondsPerYear*TicksPerSecond)
	a.output[outTrade], b.output[outTrade] = 0.9, 0.9
	exchangeTick(s)
	if a.Chat != nil || b.Chat != nil {
		t.Fatalf("a child of five trading: %+v %+v", a.Chat, b.Chat)
	}
	a.output[outTalk], b.output[outTalk] = 0.8, 0.8
	exchangeTick(s)
	if a.Deeds.Talks != 1 || a.Chat.Kind != chatTalk {
		t.Fatal("grown-up and child should talk")
	}

	// A toddler doesn't talk yet.
	s2, c, d := facePair(t)
	d.BornTick = s2.tick - int64(2.5*SecondsPerYear*TicksPerSecond)
	c.output[outTalk], d.output[outTalk] = 0.9, 0.9
	exchangeTick(s2)
	if c.Deeds.Talks+d.Deeds.Talks != 0 {
		t.Fatal("a two-year-old talked")
	}

	// Resting or working at a job, nobody exchanges.
	s3, e, f := facePair(t)
	e.output[outTalk], f.output[outTalk] = 0.9, 0.9
	f.resting = true
	exchangeTick(s3)
	if e.Deeds.Talks != 0 || e.Chat != nil {
		t.Fatal("talked with someone resting")
	}
	f.resting = false
	f.Job = &Job{Kind: "craft", Remaining: 3}
	exchangeTick(s3)
	if e.Deeds.Talks != 0 {
		t.Fatal("talked with someone at work")
	}
	f.Job = nil
	exchangeTick(s3)
	if e.Deeds.Talks != 1 {
		t.Fatal("should talk once free")
	}
}

// Three want to talk at once: the closest pair facing each other talks, the
// same way every time, and the one left over is not taken into it.
func TestExchangePairingIsDeterministic(t *testing.T) {
	run := func(cFirst bool) (*Sim, []*Creature) {
		s, _, x, y := fakeWorld(t)
		var c *Creature
		if cFirst {
			c = person(s, Male, "Cahyo", x, y)
		}
		a := person(s, Female, "Ani", x, y)
		b := person(s, Male, "Bayu", x, y)
		if !cFirst {
			c = person(s, Male, "Cahyo", x, y)
		}
		b.X, c.X = a.X+1, a.X+1.3
		a.Heading, b.Heading, c.Heading = 0, math.Pi, math.Pi
		for _, p := range []*Creature{a, b, c} {
			p.output[outTalk] = 0.9
		}
		s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures)
		exchangeTick(s)
		return s, []*Creature{a, b, c}
	}
	for _, cFirst := range []bool{false, true} {
		s, p := run(cFirst)
		a, b, c := p[0], p[1], p[2]
		if a.Deeds.Talks != 1 || b.Deeds.Talks != 1 || c.Deeds.Talks != 0 || a.Chat.With.ID != b.ID {
			t.Fatalf("cFirst %v: Ani and Bayu face each other and should talk: %+v %+v %+v", cFirst, a.Chat, b.Chat, c.Chat)
		}
		// Cahyo, behind Bayu's back, waits on Bayu (if he asked first) or
		// finds nobody free in front of him.
		if cFirst != c.Chat.waiting() {
			t.Fatalf("cFirst %v: Cahyo's chat %+v", cFirst, c.Chat)
		}
		s2, _ := run(cFirst)
		j1, _ := json.Marshal(s.creatures)
		j2, _ := json.Marshal(s2.creatures)
		if !bytes.Equal(j1, j2) {
			t.Fatal("the same scene paired differently")
		}
	}
}

// A world saved while someone waits for an answer carries on exactly.
func TestExchangeSaveRestoreMidRequest(t *testing.T) {
	s, a, b := facePair(t)
	for _, c := range []*Creature{a, b} {
		c.Genome.BOut[outMove] = -10 // stand still
		c.Genome.BOut[outRest] = -10
		c.Genome.Traits.Immunity = 1 // as a restore would make it
		c.Genome.Traits.upgradeTemperament()
	}
	a.Genome.BOut[outTalk] = 3
	s.Advance(6)
	if !a.Chat.waiting() || a.Chat.With.ID != b.ID {
		t.Fatalf("Sari should be waiting on Budi: %+v", a.Chat)
	}
	data, err := s.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	r, err := restoreWith(testMap(t, 48), data, fakeCatalog())
	if err != nil {
		t.Fatal(err)
	}
	if rc := r.byID[a.ID].Chat; rc == nil || *rc != *a.Chat {
		t.Fatalf("request lost in the save: %+v", rc)
	}
	// Budi comes round to wanting a talk, in both worlds.
	for _, w := range []*Sim{s, r} {
		w.byID[b.ID].Genome.BOut[outTalk] = 3
		w.Advance(12)
	}
	if a.Deeds.Talks != 1 || s.exch.Talks != 1 {
		t.Fatalf("no talk after Budi wanted one: %+v", a.Deeds)
	}
	j1, _ := json.Marshal([]any{s.creatures, s.exch, s.exchangeInfo()})
	j2, _ := json.Marshal([]any{r.creatures, r.exch, r.exchangeInfo()})
	if !bytes.Equal(j1, j2) {
		t.Fatal("the restored world diverged")
	}
	if !bytes.Equal(s.encodeFrame(), r.encodeFrame()) {
		t.Fatal("frames differ after restore")
	}
}

func TestNoExchange(t *testing.T) {
	s, a, b := facePair(t)
	s.opts.NoExchange = true
	a.output[outTalk], b.output[outTalk] = 0.9, 0.9
	a.output[outTrade], b.output[outTrade] = 0.9, 0.9
	for range 5 {
		exchangeTick(s)
	}
	if a.Chat != nil || b.Chat != nil || a.Deeds.Talks+a.Deeds.Trades != 0 || s.exch.Talks != 0 {
		t.Fatal("exchanges with the switch off")
	}
	// Even a chat left over from elsewhere is neither sensed nor shown.
	a.Chat = &Chat{With: Ref{b.ID, b.Name}, Kind: chatTalk, Since: s.tick}
	if talk, trade := exchangeSense(s, b); talk != 0 || trade != 0 {
		t.Fatal("senses must stay 0")
	}
	a.Chat = &Chat{With: Ref{b.ID, b.Name}, Kind: chatTalk, Since: s.tick, Ended: s.tick, OK: true}
	if s.exchangeFlags(a) != 0 || s.exchangeView(a) != nil {
		t.Fatal("nothing to show with the switch off")
	}
}

// TestProfileExchange measures exchanges at their busiest on a saved world
// (PROFILE_MAP and PROFILE_STATE, as TestProfileSave): everyone's wish to
// talk and to trade is set at random, changing twice a second, so about
// half want each. Skipped otherwise.
func TestProfileExchange(t *testing.T) {
	mapPath, statePath := os.Getenv("PROFILE_MAP"), os.Getenv("PROFILE_STATE")
	if mapPath == "" || statePath == "" {
		t.Skip("set PROFILE_MAP and PROFILE_STATE")
	}
	raw, err := os.ReadFile(mapPath)
	if err != nil {
		t.Fatal(err)
	}
	var m world.Map
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	s, err := Restore(&m, data)
	if err != nil {
		t.Fatal(err)
	}
	s.Advance(TicksPerSecond)
	ticks := 30 * TicksPerSecond
	if v, err := strconv.Atoi(os.Getenv("PROFILE_SECONDS")); err == nil && v > 0 {
		ticks = v * TicksPerSecond
	}
	var pairing, sensing time.Duration
	for range ticks {
		s.tick++
		s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures)
		for _, c := range s.creatures {
			c.output[outTalk] = hash01(c.ID, s.tick/10, 1)
			c.output[outTrade] = hash01(c.ID, s.tick/10, 2)
		}
		start := time.Now()
		s.exchange()
		pairing += time.Since(start)
		start = time.Now()
		for _, c := range s.creatures {
			if (s.tick+c.ID)%4 == 0 {
				var in [NumInputs]float64
				s.senseExchange(c, &in)
			}
		}
		sensing += time.Since(start)
	}
	per := func(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 / float64(ticks) }
	info := s.exchangeInfo()
	t.Logf("%d people: pairing %.4f ms/tick, senses %.4f ms/tick; %d talks, %d trades, %d rumours, %d opinions",
		len(s.creatures), per(pairing), per(sensing), info.Talks, info.Trades, info.Rumors, info.Opinions)
}
