package sim

import (
	"math"
	"testing"
)

// outputsOf runs a fresh genome's brain to rest on fixed senses.
func outputsOf(g *Genome, in [NumInputs]float64) [NumOutputs]float64 {
	hidden := make([]float64, g.Hidden)
	var out [NumOutputs]float64
	for range 30 {
		think(g, g.WIn, g.WOut, nil, &in, hidden, &out)
	}
	return out
}

func TestDriveIsQuietWhenFedAndSendsTheHungryOut(t *testing.T) {
	g := emptyGenome(driveNeurons)
	g.addDrive(0)
	senses := func(energy, memory, side float64) [NumInputs]float64 {
		var in [NumInputs]float64
		in[inEnergy], in[inHydration], in[inBias] = energy, 1, 1
		in[inFoodMemory], in[inFoodMemorySide] = memory, side
		return in
	}
	fed := outputsOf(g, senses(1, 0.8, 1))
	if math.Abs(fed[outMove]-0.5) > 0.03 || math.Abs(fed[outEat]-0.5) > 0.02 || math.Abs(fed[outDrink]-0.5) > 0.03 || math.Abs(fed[outTurn]) > 0.1 {
		t.Fatalf("a fed body is pushed: move %.3f eat %.3f turn %.3f", fed[outMove], fed[outEat], fed[outTurn])
	}
	hungry := outputsOf(g, senses(0.3, 0.8, 1))
	if hungry[outMove] < 0.8 || hungry[outEat] < 0.7 {
		t.Fatalf("hunger does not send them out: move %.2f eat %.2f", hungry[outMove], hungry[outEat])
	}
	right, left := hungry[outTurn], outputsOf(g, senses(0.3, 0.8, -1))[outTurn]
	if right < 0.5 || left > -0.5 {
		t.Fatalf("does not turn towards remembered food: right %.2f left %.2f", right, left)
	}
}

func TestOldBrainsGainTheDriveAndStaySound(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	a := person(s, Female, "Sari", x, y)
	b := person(s, Male, "Budi", x+1, y)
	b.Genome = a.Genome // shared, as parents' and children's often are
	s.giveBrain(b)
	H := a.Genome.Hidden
	a.Mind.addGrown(a, 20)
	seen := map[*Genome]*Genome{}
	giveDrive(a, seen)
	giveDrive(b, seen)
	for _, c := range []*Creature{a, b} {
		if c.Genome.Hidden != H+driveNeurons || len(c.Hidden) != c.Genome.Hidden || !c.Genome.valid() || !c.Mind.valid(c.Genome) {
			t.Fatalf("%s: malformed brain after the drive", c.Name)
		}
		c.Mind.rebuild(c.Genome)
		s.decide(c) // thinks without panicking
	}
	if a.Genome != b.Genome {
		t.Fatal("a shared genome was copied twice")
	}
	if a.Genome.WIn[inEnergy*a.Genome.Hidden+H] != -4 {
		t.Fatal("drive neuron not wired")
	}
}

func TestFoodPlacesArePassedOnInTalk(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	a := person(s, Female, "Sari", x, y)
	b := person(s, Male, "Budi", x+1, y)
	a.FoodPlaces = []FoodPlace{{X: a.X + 15, Y: a.Y, Rich: 1, Tick: s.tick}}
	s.tellFood(a, b, 0.9)
	if len(b.FoodPlaces) != 1 || b.FoodPlaces[0].Rich >= 1 || b.FoodPlaces[0].X != a.X+15 {
		t.Fatalf("listener's places %+v", b.FoodPlaces)
	}
	s.tellFood(a, b, 0.9) // already known
	if len(b.FoodPlaces) != 1 {
		t.Fatal("told twice")
	}
}

func TestThirstComesBeforeHunger(t *testing.T) {
	g := emptyGenome(driveNeurons)
	g.addDrive(0)
	var in [NumInputs]float64
	in[inEnergy], in[inHydration], in[inBias] = 0.3, 0.3, 1
	both := outputsOf(g, in)
	if both[outDrink] <= both[outEat] || both[outMove] < 0.8 {
		t.Fatalf("hungry and thirsty: drink %.2f eat %.2f move %.2f", both[outDrink], both[outEat], both[outMove])
	}
	in[inWaterNear] = 1 // at the river: drinking is up to the old reflexes
	at := outputsOf(g, in)
	if at[outDrink] > both[outDrink]-0.1 {
		t.Fatalf("the thirst drive still pushes at the water's edge: %.2f", at[outDrink])
	}
}
