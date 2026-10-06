package sim

import (
	"math"
	"strings"
	"testing"
)

func TestGossipPassesNewsLessSurely(t *testing.T) {
	s, a, b := facePair(t)
	thief := person(s, Male, "Maling", int(a.X)+4, int(a.Y))
	s.tick = 100
	a.Memories = []PerceivedEvent{
		{Kind: "steal", Actor: &Ref{thief.ID, thief.Name}, Mode: "seen", X: thief.X, Y: thief.Y, Tick: 90, Confidence: 0.65},
		// About Budi himself: not told to him.
		{Kind: "attack", Actor: &Ref{b.ID, b.Name}, Mode: "seen", Tick: 95, Confidence: 0.65},
	}
	a.output[outTalk], b.output[outTalk] = 0.9, 0.9
	exchangeTick(s)
	if a.Deeds.Talks != 1 {
		t.Fatal("no talk")
	}
	if len(b.Memories) != 1 {
		t.Fatalf("Budi should hear exactly the theft: %+v", b.Memories)
	}
	m := b.Memories[0]
	// A stranger is believed half: 0.65 × 0.5.
	if m.Kind != "steal" || m.Mode != modeTold || m.Actor.ID != thief.ID || m.Tick != 90 || math.Abs(m.Confidence-0.325) > 1e-9 {
		t.Fatalf("told memory %+v", m)
	}
	if m.Actor == a.Memories[0].Actor {
		t.Fatal("the told memory shares the teller's Ref")
	}
	trust := s.opinion(b, thief)
	if math.Abs(trust-(-0.3*0.325)) > 1e-9 {
		t.Fatalf("Budi's trust in the thief %v, want %v", trust, -0.3*0.325)
	}
	if r := b.Relations[0]; r.Person.ID != thief.ID || r.Met != 0 {
		t.Fatalf("hearsay is not a meeting: %+v", b.Relations)
	}
	if len(a.Memories) != 2 || s.exch.Rumors != 1 {
		t.Fatalf("Sari heard nothing new; one rumour passed: %d memories, %d rumours", len(a.Memories), s.exch.Rumors)
	}
	if !strings.Contains(eventsText(s), "Budi mendengar dari Sari bahwa Maling pernah mencuri") {
		t.Fatalf("first rumour not logged: %q", eventsText(s))
	}
	if v := s.exchangeView(b); v == nil || v.Told != 1 || v.Talks != 1 {
		t.Fatalf("view %+v", v)
	}

	// Talking again passes nothing twice.
	for range chatRest {
		exchangeTick(s)
	}
	if a.Deeds.Talks != 2 || len(b.Memories) != 1 || s.exch.Rumors != 1 {
		t.Fatalf("news repeated: talks %d, %d memories, %d rumours", a.Deeds.Talks, len(b.Memories), s.exch.Rumors)
	}
	if r := b.Relations[0]; r.Person.ID != thief.ID || math.Abs(r.Trust-trust) > 1e-12 {
		t.Fatalf("the same news changed his mind twice: %+v", r)
	}
}

// Retold from one to the next, news is believed less each time until it is
// no longer worth passing on.
func TestRumoursFade(t *testing.T) {
	s, a, b := facePair(t)
	thief := person(s, Male, "Maling", int(a.X)+4, int(a.Y))
	a.Memories = []PerceivedEvent{{Kind: "steal", Actor: &Ref{thief.ID, thief.Name}, Mode: "told", Tick: 1, Confidence: 0.09}}
	if _, n := s.rumours(a, b); n != 0 {
		t.Fatal("0.09 believed half is below the floor: not worth telling")
	}
	a.Memories[0].Confidence = 0.2
	if _, n := s.rumours(a, b); n != 1 {
		t.Fatal("0.2 believed half is worth telling")
	}
	// The dead are not talked about; nor are unknown doers (sounds).
	a.Memories = append(a.Memories, PerceivedEvent{Kind: "attack", Mode: "heard", Tick: 2, Confidence: 0.4})
	thief.Health = 0
	if _, n := s.rumours(a, b); n != 0 {
		t.Fatal("news of the dead or of nobody in particular passed on")
	}
}

func TestRumoursMostStrikingFirst(t *testing.T) {
	s, a, b := facePair(t)
	var people []*Creature
	for i, name := range []string{"Satu", "Dua", "Tiga", "Empat"} {
		people = append(people, person(s, Male, name, int(a.X)+3+i, int(a.Y)))
	}
	ref := func(c *Creature) *Ref { return &Ref{c.ID, c.Name} }
	a.Memories = []PerceivedEvent{
		{Kind: "give", Actor: ref(people[0]), Mode: "seen", Tick: 1, Confidence: 0.65},   // 0.117
		{Kind: "attack", Actor: ref(people[1]), Mode: "seen", Tick: 1, Confidence: 0.65}, // 0.325
		{Kind: "steal", Actor: ref(people[2]), Mode: "direct", Tick: 1, Confidence: 1},   // 0.3
		{Kind: "steal", Actor: ref(people[3]), Mode: "seen", Tick: 1, Confidence: 0.65},  // 0.195
	}
	news, n := s.rumours(a, b)
	if n != rumoursPerTalk || news[0].Actor.ID != people[1].ID || news[1].Actor.ID != people[2].ID {
		t.Fatalf("picked %d: %+v", n, news[:n])
	}
}

// What one thinks of a third person rubs off a little on the listener.
func TestOpinionsSpread(t *testing.T) {
	s, a, b := facePair(t)
	bully := person(s, Male, "Preman", int(a.X)+4, int(a.Y))
	friend := person(s, Female, "Teman", int(a.X)+4, int(a.Y)+2)
	s.rememberRelation(a, bully, -0.8)
	s.rememberRelation(a, friend, 0.3)
	a.output[outTalk], b.output[outTalk] = 0.9, 0.9
	exchangeTick(s)
	// Pulled a quarter of the way, believed half: −0.8 × 0.25 × 0.5.
	if got := s.opinion(b, bully); math.Abs(got-(-0.1)) > 1e-3 {
		t.Fatalf("Budi's view of the bully %v, want -0.1", got)
	}
	if s.opinion(b, friend) != 0 || s.exch.Opinions != 1 {
		t.Fatal("only the strongest opinion is passed on")
	}
	// Hearing it again draws closer to it but never past it.
	for range 40 * chatRest {
		exchangeTick(s)
	}
	if got := s.opinion(b, bully); got > -0.3 || got < -0.8 {
		t.Fatalf("after many talks Budi's view %v should approach Sari's", got)
	}
}

// Hearsay never pushes someone out of a full table of relations if it is
// weaker than everything there.
func TestHearsayKeepsKnownPeople(t *testing.T) {
	s, a, _ := facePair(t)
	for i := range maxRelations {
		o := person(s, Male, "Kenalan", int(a.X)+3, int(a.Y)+i%3)
		s.rememberRelation(a, o, 0.5)
	}
	stranger := person(s, Male, "Asing", int(a.X)+4, int(a.Y))
	s.heardOpinion(a, stranger, -0.4, 0.1)
	if s.opinion(a, stranger) != 0 || len(a.Relations) != maxRelations {
		t.Fatal("a weak rumour replaced someone known")
	}
	s.rememberRelation(a, stranger, -0.6) // met: does replace the weakest
	if s.opinion(a, stranger) == 0 {
		t.Fatal("an encounter should make room")
	}
}
