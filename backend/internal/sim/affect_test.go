package sim

import (
	"math"
	"math/rand/v2"
	"testing"
)

// moodTicks runs c's mood updates for the given simulated seconds (only the
// moods, not the rest of the body).
func moodTicks(s *Sim, c *Creature, seconds float64) {
	for range int(seconds * TicksPerSecond) {
		s.tick++
		s.affectStep(c)
	}
}

func TestMoodsRiseAndEbbToTheirRestingLevel(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Female, "Rasa", x, y)
	moodTicks(s, c, 0.5)
	if c.Affect.At == 0 {
		t.Fatal("moods never brought up to date")
	}

	// Calm drifts to the resting joy; nothing else.
	moodTicks(s, c, 20)
	a := c.Affect
	if math.Abs(a.Joy-joyRest) > 0.01 || a.Fear != 0 || a.Anger != 0 || a.Grief != 0 {
		t.Fatalf("resting moods %+v", a)
	}

	// Struck: fear rises within a second towards hurtFear, and ebbs once
	// the pain is over.
	c.Hurt = hurtSeconds
	moodTicks(s, c, 1)
	if c.Affect.Fear < 0.9*hurtFear || c.Affect.Fear > hurtFear+1e-9 {
		t.Fatalf("fear while struck %.3f", c.Affect.Fear)
	}
	c.Hurt = 0
	moodTicks(s, c, fearTime)
	afraid := c.Affect.Fear
	if afraid > 0.6*hurtFear || afraid <= 0 {
		t.Fatalf("fear %.3f after %gs", afraid, fearTime)
	}
	moodTicks(s, c, 10)
	if c.Affect.Fear > 0.01 {
		t.Fatalf("fear never ebbed: %.3f", c.Affect.Fear)
	}

	// A jump of joy ebbs back down to the resting level, not below.
	s.feel(c, moodJoy, 0.8)
	moodTicks(s, c, joyTime)
	if c.Affect.Joy < joyRest+0.1 || c.Affect.Joy > 0.8 {
		t.Fatalf("joy %.3f after %gs", c.Affect.Joy, joyTime)
	}
	moodTicks(s, c, 30)
	if math.Abs(c.Affect.Joy-joyRest) > 0.01 {
		t.Fatalf("joy %.3f, resting %.3f", c.Affect.Joy, joyRest)
	}

	// Hunger angers mildly; a meal when hungry gladdens.
	c.Energy = 0.05
	moodTicks(s, c, 2)
	if c.Affect.Anger < 0.5*hungerAnger {
		t.Fatalf("anger when starving %.3f", c.Affect.Anger)
	}
	before := c.Affect.Joy
	for range 10 {
		c.Energy += 0.05
		moodTicks(s, c, 0.2)
	}
	if c.Affect.Joy < before+0.2 {
		t.Fatalf("a meal when hungry: joy %.3f → %.3f", before, c.Affect.Joy)
	}

	// Grief eases over months and is mostly gone within a year or two.
	s.feel(c, moodGrief, 1)
	moodTicks(s, c, SecondsPerYear/4)
	g0 := c.Affect.Grief
	moodTicks(s, c, SecondsPerYear*3/4)
	g1 := c.Affect.Grief
	moodTicks(s, c, SecondsPerYear)
	if g0 < 0.6 || g1 < 0.05 || g1 > 0.25 || c.Affect.Grief > 0.03 {
		t.Fatalf("grief %.2f after three months, %.2f after a year, %.2f after two", g0, g1, c.Affect.Grief)
	}
}

func TestTemperamentSetsHowMoodsMove(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	calm := person(s, Female, "Tenang", x, y)
	touchy := person(s, Female, "Peka", x+1, y)
	calm.Genome.Traits.Reactivity, calm.Genome.Traits.Recovery, calm.Genome.Traits.Cheer = 0.5, 2, 2
	touchy.Genome.Traits.Reactivity, touchy.Genome.Traits.Recovery, touchy.Genome.Traits.Cheer = 1.5, 0.5, 0.5
	for _, c := range []*Creature{calm, touchy} {
		moodTicks(s, c, 30)
		s.feel(c, moodAnger, 0.4)
	}
	if calm.Affect.Anger >= touchy.Affect.Anger {
		t.Fatal("reactivity did not scale a jump")
	}
	if calm.Affect.Joy <= touchy.Affect.Joy || math.Abs(calm.Affect.Joy-2*joyRest) > 0.02 {
		t.Fatalf("resting joy %.3f vs %.3f", calm.Affect.Joy, touchy.Affect.Joy)
	}
	calm.Affect.Anger, touchy.Affect.Anger = 0.5, 0.5
	moodTicks(s, calm, 1)
	moodTicks(s, touchy, 1)
	if calm.Affect.Anger >= touchy.Affect.Anger {
		t.Fatal("recovery did not speed the ebb")
	}
}

func TestDirectWrongsAndGiftsMoveTheTarget(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	a := person(s, Male, "A", x, y)
	b := person(s, Female, "B", x+1, y)
	onlooker := person(s, Female, "C", x-1, y)
	s.perceiveEvent(a, b, "attack", -0.5, 8)
	if math.Abs(b.Affect.Anger-0.5) > 1e-9 || math.Abs(b.Affect.Fear-assaultFear) > 1e-9 {
		t.Fatalf("assaulted: %+v", b.Affect)
	}
	if onlooker.Affect != (Affect{}) || a.Affect != (Affect{}) {
		t.Fatal("only the target's moods jump; onlookers feel the fight through their senses")
	}
	s.perceiveEvent(a, b, "give", 0.18, 0)
	if math.Abs(b.Affect.Joy-giftJoy*0.18) > 1e-9 {
		t.Fatalf("gift: joy %.3f", b.Affect.Joy)
	}
	// The onlooker is frightened by the fight she perceives.
	onlooker.Heading = 0
	s.tick = 3 // (tick+id)%affectEvery == 0 for id 1 … 3 on some tick
	for range 2 * affectEvery {
		s.tick++
		s.affectStep(onlooker)
	}
	if onlooker.Affect.Fear <= 0 {
		t.Fatal("perceived fight did not frighten")
	}

	// An animal's bite frightens and leaves a predator stimulus.
	s.Maul(b.ID, 0.1, "harimau")
	found := false
	for _, st := range s.stims().list {
		found = found || st.Kind == stimPredator
	}
	if !found || b.Affect.Fear <= assaultFear {
		t.Fatal("a mauling left no fright")
	}
}

func TestMourningHitsKinNotStrangers(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	mother := person(s, Female, "Ibu", x, y)
	father := person(s, Male, "Ayah", x+1, y)
	mother.Spouse, father.Spouse = &Ref{father.ID, father.Name}, &Ref{mother.ID, mother.Name}
	child := func(name string, m, f *Creature) *Creature {
		c := person(s, Female, name, x, y+1)
		if m != nil {
			c.Mother = &Ref{m.ID, m.Name}
		}
		if f != nil {
			c.Father = &Ref{f.ID, f.Name}
		}
		return c
	}
	other := person(s, Male, "Lain", x-1, y)
	dead := child("Anak", mother, father)
	sister := child("Kakak", mother, father)
	half := child("Tiri", mother, other)
	stranger := person(s, Male, "Asing", x+2, y)
	grandchild := child("Cucu", dead, nil)

	dead.Health = 0
	s.reap()
	cases := []struct {
		c    *Creature
		want float64
	}{
		{mother, griefChild}, {father, griefChild}, {sister, griefSibling}, {half, griefHalfSibling},
		{grandchild, griefParent}, {stranger, 0}, {other, 0},
	}
	for _, k := range cases {
		if math.Abs(k.c.Affect.Grief-k.want) > 1e-9 {
			t.Fatalf("%s grieves %.2f, want %.2f", k.c.Name, k.c.Affect.Grief, k.want)
		}
	}
	// A spouse grieves; with moods off nobody does.
	mother.Affect.Grief = 0
	s.mourn(father)
	if mother.Affect.Grief != griefSpouse || stranger.Affect.Grief != 0 {
		t.Fatalf("widow grieves %.2f", mother.Affect.Grief)
	}
	s.opts.NoAffect = true
	mother.Affect.Grief = 0
	s.mourn(father)
	if mother.Affect.Grief != 0 {
		t.Fatal("grief with moods off")
	}
}

func TestBirthAndConceptionGladdenParents(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	mother := person(s, Female, "Ibu", x, y)
	father := person(s, Male, "Ayah", x+1, y)
	moodTicks(s, mother, 1)
	moodTicks(s, father, 1)
	j0 := mother.Affect.Joy
	// Conceived this tick: mate sets the pregnancy, gestate counts it down.
	mother.Pregnancy = &Pregnancy{Remaining: gestation, Father: Ref{father.ID, father.Name}, FatherGenome: father.Genome}
	pregnant := func(seconds float64) {
		for range int(seconds * TicksPerSecond) {
			mother.Pregnancy.Remaining -= dt
			s.tick++
			s.affectStep(mother)
		}
	}
	pregnant(0.2)
	if mother.Affect.Joy < j0+0.8*conceiveJoy || father.Affect.Joy < j0+0.8*conceiveJoy {
		t.Fatalf("conception: joy %.3f / %.3f", mother.Affect.Joy, father.Affect.Joy)
	}
	j1 := mother.Affect.Joy
	pregnant(0.4)
	if mother.Affect.Joy > j1 {
		t.Fatal("conception gladdened twice")
	}
	p := mother.Pregnancy
	mother.Pregnancy = nil
	baby := s.newChild(mother, p)
	s.add(baby)
	j2 := father.Affect.Joy
	moodTicks(s, baby, 0.2)
	if father.Affect.Joy < j2+0.5*birthJoy {
		t.Fatalf("a birth: father's joy %.3f → %.3f", j2, father.Affect.Joy)
	}
}

func TestTemperamentGenesInherit(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	a, b := randomGenome(rng), randomGenome(rng)
	for _, g := range []*Genome{a, b} {
		tr := g.Traits
		if tr.Reactivity < 0.9 || tr.Reactivity > 1.1 || tr.Recovery < 0.9 || tr.Recovery > 1.1 || tr.Cheer < 0.9 || tr.Cheer > 1.1 {
			t.Fatalf("founding temperament %+v", tr)
		}
	}
	a.Traits.Reactivity, b.Traits.Reactivity = 0.5, 1.8
	a.Traits.Recovery, b.Traits.Recovery = 2.5, 0.4
	a.Traits.Cheer, b.Traits.Cheer = 0.3, 2.2
	spread := map[float64]bool{}
	for range 50 {
		c := crossover(a, b, rng)
		tr := c.Traits
		if tr.Reactivity < 0.5 || tr.Reactivity > 1.8 || tr.Recovery < 0.4 || tr.Recovery > 2.5 || tr.Cheer < 0.3 || tr.Cheer > 2.2 {
			t.Fatalf("child outside its parents: %+v", tr)
		}
		spread[math.Round(tr.Reactivity*10)] = true
		before := tr
		m := c.mutate(rng).Traits
		if math.Abs(m.Reactivity-before.Reactivity) > 0.3 || m.Reactivity < reactivityMin || m.Reactivity > reactivityMax {
			t.Fatalf("mutation jumped %.3f → %.3f", before.Reactivity, m.Reactivity)
		}
	}
	if len(spread) < 4 {
		t.Fatal("children all alike")
	}
	// The same draws give the same temperament.
	r1, r2 := rand.New(rand.NewPCG(5, 6)), rand.New(rand.NewPCG(5, 6))
	if c1, c2 := crossover(a, b, r1).mutate(r1), crossover(a, b, r2).mutate(r2); c1.Traits != c2.Traits {
		t.Fatal("temperament is not deterministic")
	}
	// Genomes from before save v11 are average.
	old := randomGenome(rng)
	old.Traits.Reactivity, old.Traits.Recovery, old.Traits.Cheer = 0, 0, 0
	old.upgrade(10)
	if old.Traits.Reactivity != 1 || old.Traits.Recovery != 1 || old.Traits.Cheer != 1 {
		t.Fatalf("upgraded temperament %+v", old.Traits)
	}
}

func TestSwitchesZeroTheSenses(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	s.opts.NoStimuli, s.opts.NoAffect = true, true
	c := facing(s, x, y, 0)
	other := person(s, Male, "B", x+1, y)
	s.addStimulus(stimFire, float64(x)+2.5, float64(y)+0.5, 1, nil)
	s.perceiveEvent(other, c, "attack", -0.5, 8)
	other.Health = 0
	s.reap()
	s.feel(c, moodJoy, 1)
	c.Hurt = hurtSeconds
	moodTicks(s, c, 2)
	s.sense(c)
	in := c.input
	for _, i := range []int{inShock, inShockSide, inFear, inAnger, inJoy, inGrief} {
		if in[i] != 0 {
			t.Fatalf("sense %s is %.3f with stimuli and moods off", InputLabels[i], in[i])
		}
	}
	if len(s.stims().list) != 0 || c.Affect != (Affect{}) || s.moodView(c) != nil {
		t.Fatal("state kept with the switches off")
	}
	if m, v := s.moodCode(c); m != 0 || v != 0 {
		t.Fatal("a face with moods off")
	}
	// Moods stuck from before the switch are cleared, never sensed.
	c.Affect.Fear = 0.9
	if s.sense(c); c.input[inFear] != 0 {
		t.Fatal("old fear sensed")
	}
	moodTicks(s, c, 0.2)
	if c.Affect != (Affect{}) {
		t.Fatal("old moods kept")
	}
}

func TestFaceShowsTheDominantMood(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Male, "Wajah", x, y)
	if m, v := s.moodCode(c); m != 0 || v != 0 || s.moodView(c).Dominant != "" {
		t.Fatal("a calm face shows a mood")
	}
	c.Affect = Affect{Fear: 0.2, Joy: 0.29}
	if m, _ := s.moodCode(c); m != 0 {
		t.Fatal("moods below the threshold shown")
	}
	c.Affect = Affect{Fear: 0.4, Anger: 0.7, Joy: 0.5, Grief: 0.69}
	if m, v := s.moodCode(c); m != 2 || v != 0.7 || s.moodView(c).Dominant != "anger" {
		t.Fatalf("got %d %.2f", m, v)
	}
	c.Affect = Affect{Grief: 0.8, Joy: 0.3}
	if m, _ := s.moodCode(c); m != 4 || s.moodView(c).Dominant != "grief" {
		t.Fatal("grief not shown")
	}
}
