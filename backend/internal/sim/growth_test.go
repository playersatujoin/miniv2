package sim

import (
	"math"
	"testing"
)

// surprised gives a creature a lively life: well fed, things keep happening.
func surprised(s *Sim, x, y int) *Creature {
	c := person(s, Female, "Ana", x, y)
	c.Genome.Traits.Neurogenesis = 1
	c.Energy = 0.9
	c.Mind.Novelty = 0.6
	for i := range c.input {
		c.input[i] = float64(i%3) * 0.3
	}
	return c
}

func TestNeuronsGrowWhenSurprised(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := surprised(s, x, y)
	for range 400 {
		s.tick++
		s.grow(c, 0.6)
	}
	m := c.Mind
	if m.Grown == 0 {
		t.Fatal("no neuron grew in a surprising life")
	}
	if !m.growthValid(c.Genome.Hidden) {
		t.Fatalf("grown arrays out of step: %d neurons", m.Grown)
	}
	if brainSize(c) != c.Genome.Hidden+m.Grown {
		t.Errorf("brain size %d, want %d", brainSize(c), c.Genome.Hidden+m.Grown)
	}
	// A new neuron fires for the moment it was born in and thinks along at once.
	var out [NumOutputs]float64
	think(c.Genome, m.wIn, m.wOut, m, &c.input, c.Hidden, &out)
	fired := false
	for _, a := range m.GAct {
		if math.IsNaN(a) || math.IsInf(a, 0) {
			t.Fatal("grown neuron activation is not a number")
		}
		fired = fired || a > 0
	}
	if !fired {
		t.Error("no grown neuron recognises the situation it grew in")
	}
}

func TestNoGrowthWhenHungryOrSwitchedOff(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	hungry := surprised(s, x, y)
	hungry.Energy = growFloor - 0.05
	off := surprised(s, x+1, y)
	for range 400 {
		s.tick++
		s.grow(hungry, 0.6)
	}
	s.opts.NoNeurogenesis = true
	for range 400 {
		s.tick++
		s.grow(off, 0.6)
	}
	if hungry.Mind.Grown != 0 || off.Mind.Grown != 0 {
		t.Errorf("grew while hungry (%d) or switched off (%d)", hungry.Mind.Grown, off.Mind.Grown)
	}
}

func TestUnusedNeuronsArePrunedAndStarvingBrainsShrink(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := surprised(s, x, y)
	c.Mind.addGrown(c, 0)
	c.Mind.addGrown(c, 0)
	// Years later, one of them never learned to say anything.
	c.BornTick = s.tick - int64((pruneAge+SecondsPerYear)/dt)
	c.Mind.GUse[0] = 0
	c.Mind.GUse[1] = 1
	c.Genome.Traits.Neurogenesis = 0 // no new ones in the way
	s.tick++
	s.grow(c, 0)
	if c.Mind.Grown != 1 || c.Mind.Pruned != 1 {
		t.Fatalf("after pruning: %d grown, %d pruned; want 1 and 1", c.Mind.Grown, c.Mind.Pruned)
	}
	if c.Mind.GUse[0] < 0.5 {
		t.Error("the useful neuron was pruned instead of the idle one")
	}
	c.Energy = starveShed / 2
	for range 400 {
		s.tick++
		s.grow(c, 0)
	}
	if c.Mind.Grown != 0 {
		t.Errorf("a starving brain kept %d grown neurons", c.Mind.Grown)
	}
	if !c.Mind.growthValid(c.Genome.Hidden) {
		t.Error("arrays out of step after pruning")
	}
}

func TestBrainCostsGrowWithIt(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	small := surprised(s, x, y)
	big := surprised(s, x+1, y)
	for range 10 {
		big.Mind.addGrown(big, 0)
	}
	small.Energy, big.Energy = 1, 1
	s.metabolize(small, 0)
	s.metabolize(big, 0)
	if big.Energy >= small.Energy {
		t.Errorf("ten more neurons cost nothing: %.6f vs %.6f", big.Energy, small.Energy)
	}
}

func TestOldGenomesUpgrade(t *testing.T) {
	g := randomGenome(newSimWith(testMap(t, 48), 3, fakeCatalog()).rng)
	g.Bias, g.Tau = nil, nil
	g.Traits.Neurogenesis = 0
	g.upgrade(6)
	if !g.valid() {
		t.Fatal("upgraded genome is not valid")
	}
	if g.Tau[0] != minTau || g.Bias[0] != 0 || g.Traits.Neurogenesis != neurogenesisInit {
		t.Errorf("upgrade gave tau %v bias %v neurogenesis %v", g.Tau[0], g.Bias[0], g.Traits.Neurogenesis)
	}
	// A v7 genome that evolved the gene down to zero keeps it.
	g.Traits.Neurogenesis = 0
	g.upgrade(7)
	if g.Traits.Neurogenesis != 0 {
		t.Error("upgrade overwrote an evolved neurogenesis gene")
	}
}
