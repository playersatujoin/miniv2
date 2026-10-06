package sim

import (
	"fmt"
	"math"
)

// Gossip: when two people talk, each passes on what they remember about
// others, after RAGE's crime witnesses (game/witness.h, CCrimeWitnesses,
// and the witness personalities of WitnessInformation.h), where those who
// saw a crime carry the news on. Here there are no police: news goes from
// person to person, so a reputation spreads through who talks with whom,
// never as a world-wide number (PLAN.md, Fase 4 design point 2). What is
// told is believed less than what was seen, and less still from someone
// the listener distrusts; retold again and again it fades out, as rumours
// lose detail and weight from one teller to the next (Allport & Postman
// 1947). People do pass on such news readily: in experiments gossip sways
// whom people help about as much as seeing for themselves (Sommerfeld et
// al. 2007).

const (
	// modeTold marks a memory someone was told rather than saw.
	modeTold = "told"
	// rumoursPerTalk is how many memories each passes on in one talk, the
	// most striking first.
	rumoursPerTalk = 2
	// rumourFloor: news believed less than this isn't worth passing on.
	rumourFloor = 0.05
	// opinionFloor: only a clear opinion of someone is brought up, and
	// opinionPull is how far the listener's view moves towards it, before
	// weighing how much they believe the teller.
	opinionFloor = 0.2
	opinionPull  = 0.25
	// talkTrust: talking is social grooming that builds bonds (Dunbar 1996),
	// a little at a time: less than a tenth of a meal given (social.go).
	talkTrust = 0.02
	talkJoy   = 0.05
	// rumorEventGap is the seconds between rumours written in the event log.
	rumorEventGap = 30.0
)

// deedValue is how a deed changes the trust of whoever witnesses it, as in
// social.go: a gift +0.18, a theft −0.3, an assault −0.5. Other memories are
// not passed on.
func deedValue(kind string) float64 {
	switch kind {
	case "give":
		return 0.18
	case "steal":
		return -0.3
	case "attack":
		return -0.5
	}
	return 0
}

// belief is how far listener believes what teller tells them: half for a
// stranger, more for someone trusted, little for someone distrusted.
func (s *Sim) belief(listener, teller *Creature) float64 {
	return clamp(0.5+0.4*s.opinion(listener, teller), 0.1, 0.9)
}

// talk is a conversation between a and b: each passes on news and an
// opinion of a third person, and both feel a little closer and happier.
func (s *Sim) talk(a, b *Creature) {
	// Everything is chosen before anything is heard, so neither echoes back
	// what the other has just said.
	newsA, nA := s.rumours(a, b)
	newsB, nB := s.rumours(b, a)
	aboutA, viewA := s.strongestOpinion(a, b, opinionFloor)
	aboutB, viewB := s.strongestOpinion(b, a, opinionFloor)
	beliefB, beliefA := s.belief(b, a), s.belief(a, b)
	s.tell(a, b, newsA[:nA], beliefB, aboutA, viewA)
	s.tell(b, a, newsB[:nB], beliefA, aboutB, viewB)
	s.tellFood(a, b, beliefB)
	s.tellFood(b, a, beliefA)
	s.rememberRelation(a, b, talkTrust)
	s.rememberRelation(b, a, talkTrust)
	s.feel(a, moodJoy, talkJoy)
	s.feel(b, moodJoy, talkJoy)
	s.exch.Talks++
}

// rumours picks what teller would pass on to listener: their most striking
// memories of what living third people did (not the listener, whom it
// concerns, nor the teller), that the listener doesn't already know and
// would believe enough to matter. Striking is how sure, how grave and how
// fresh (fading as trust does, see relationTrust).
func (s *Sim) rumours(teller, listener *Creature) ([rumoursPerTalk]PerceivedEvent, int) {
	var picked [rumoursPerTalk]PerceivedEvent
	var score [rumoursPerTalk]float64
	n := 0
	belief := s.belief(listener, teller)
	for _, ev := range teller.Memories {
		v := deedValue(ev.Kind)
		if v == 0 || ev.Actor == nil || ev.Actor.ID == listener.ID || ev.Actor.ID == teller.ID ||
			ev.Confidence*belief < rumourFloor || knows(listener, ev) || s.living(ev.Actor.ID) == nil {
			continue
		}
		age := math.Max(0, float64(s.tick-ev.Tick)*dt)
		sc := ev.Confidence * math.Abs(v) * math.Exp(-age/(30*SecondsPerYear))
		// Insert into the short list, best first; ties keep the earlier memory.
		k := n
		for k > 0 && sc > score[k-1] {
			k--
		}
		if k >= rumoursPerTalk {
			continue
		}
		if n < rumoursPerTalk {
			n++
		}
		copy(picked[k+1:n], picked[k:n-1])
		copy(score[k+1:n], score[k:n-1])
		picked[k], score[k] = ev, sc
	}
	return picked, n
}

// knows reports whether c already remembers the deed ev (same doer, kind
// and moment), however they learned of it.
func knows(c *Creature, ev PerceivedEvent) bool {
	for _, m := range c.Memories {
		if m.Actor != nil && m.Actor.ID == ev.Actor.ID && m.Kind == ev.Kind && m.Tick == ev.Tick {
			return true
		}
	}
	return false
}

// tell passes news and an opinion from teller to listener, who believes it
// as far as belief.
func (s *Sim) tell(teller, listener *Creature, news []PerceivedEvent, belief float64, about *Creature, view float64) {
	for _, ev := range news {
		actor := s.living(ev.Actor.ID)
		if actor == nil {
			continue
		}
		told := ev
		told.Actor = &Ref{actor.ID, actor.Name}
		told.Mode = modeTold
		told.Confidence = ev.Confidence * belief
		remember(listener, told)
		// Heard of, not met: see shiftRelation.
		s.shiftRelation(listener, actor, deedValue(ev.Kind)*told.Confidence, 0)
		s.exch.Rumors++
		s.rumourEvent(teller, listener, actor, ev.Kind)
	}
	if about != nil && about != listener {
		s.heardOpinion(listener, about, view, opinionPull*belief)
		s.exch.Opinions++
	}
}

// remember adds ev to c's memories, forgetting the oldest beyond maxMemories
// (as perceiveEvent does).
func remember(c *Creature, ev PerceivedEvent) {
	c.Memories = append(c.Memories, ev)
	if len(c.Memories) > maxMemories {
		copy(c.Memories, c.Memories[len(c.Memories)-maxMemories:])
		c.Memories = c.Memories[:maxMemories]
	}
}

// rumourEvent writes, now and then, a piece of news going round in the log.
func (s *Sim) rumourEvent(teller, listener, actor *Creature, kind string) {
	if s.time()-s.exch.RumorEvent < rumorEventGap && s.exch.RumorEvent > 0 {
		return
	}
	deed := map[string]string{"give": "suka berbagi", "steal": "pernah mencuri", "attack": "pernah menyerang orang"}[kind]
	s.exch.RumorEvent = s.time()
	s.event("rumor", fmt.Sprintf("%s mendengar dari %s bahwa %s %s", listener.Name, teller.Name, actor.Name, deed), listener.ID)
}
