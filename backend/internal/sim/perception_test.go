package sim

import (
	"math"
	"testing"
)

func TestWitnessMustSeeActorAndHearingDoesNotRevealIdentity(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	actor := person(s, Male, "Actor", x, y)
	victim := person(s, Female, "Victim", x+1, y)
	witness := person(s, Female, "Witness", x+3, y)
	witness.Heading = math.Pi
	s.perceiveEvent(actor, victim, "give", 0.2, 0)
	if s.opinion(victim, actor) <= 0 || s.opinion(witness, actor) <= 0 {
		t.Fatal("visible gift not remembered")
	}
	witness.Relations, witness.Memories = nil, nil
	s.terrain.blocked[y*s.terrain.w+x+2] = true
	s.perceiveEvent(actor, victim, "attack", -0.5, 8)
	if s.opinion(victim, actor) >= 0 {
		t.Fatal("victim did not remember attacker")
	}
	if len(witness.Relations) != 0 {
		t.Fatal("heard sound revealed the attacker's identity")
	}
	if len(witness.Memories) != 1 || witness.Memories[0].Mode != "heard" || witness.Memories[0].Actor != nil {
		t.Fatal("expected unidentified sound behind obstruction")
	}
	if s.alarm(witness) <= 0 {
		t.Fatal("heard assault did not reach the senses")
	}
	s.tick += 4 * TicksPerSecond
	if s.alarm(witness) != 0 {
		t.Fatal("alarm did not expire")
	}
}

func TestPerceptionMemoryIsBoundedAndOpinionIsPersonal(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	a := person(s, Male, "A", x, y)
	b := person(s, Female, "B", x+1, y)
	for i := 0; i < 50; i++ {
		other := person(s, Male, "Other", x, y)
		s.perceiveEvent(other, a, "give", 0.1, 0)
	}
	if len(a.Relations) != maxRelations || len(a.Memories) != maxMemories {
		t.Fatal("unbounded social memory")
	}
	b.Reputation = -1
	if s.opinion(a, b) != 0 {
		t.Fatal("unknown global reputation leaked into private knowledge")
	}
}

func TestLocalRouteAvoidsWallsAndRefusesSealedDestination(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	for dy := -1; dy <= 1; dy++ {
		s.terrain.blocked[(y+dy)*s.terrain.w+x+1] = true
	}
	path := s.localPath(float64(x)+0.5, float64(y)+0.5, float64(x+3)+0.5, float64(y)+0.5, 4)
	if len(path) == 0 {
		t.Fatal("no detour around a short wall")
	}
	for _, p := range path {
		if s.terrain.blockedAt(p.X, p.Y) {
			t.Fatal("route enters solid terrain")
		}
	}
	goal := path[len(path)-1]
	if goal.X != float64(x+3)+0.5 || goal.Y != float64(y)+0.5 {
		t.Fatal("route did not reach target")
	}
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx != 0 || dy != 0 {
				s.terrain.blocked[(y+dy)*s.terrain.w+x+3+dx] = true
			}
		}
	}
	if p := s.localPath(float64(x)+0.5, float64(y)+0.5, goal.X, goal.Y, 5); len(p) != 0 {
		t.Fatal("route crosses sealed destination")
	}
}

func TestRestAndChangedIntentCancelTravel(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Male, "A", x, y)
	c.Travel = &Travel{Action: ActDrink, Phase: "approach", Target: Waypoint{1, 1}}
	c.resting = true
	s.navigate(c)
	if c.Travel != nil {
		t.Fatal("rest did not cancel travel")
	}
	c.resting = false
	c.Travel = &Travel{Action: ActDrink, Phase: "approach"}
	c.output[outDrink] = 0
	s.navigate(c)
	if c.Travel != nil {
		t.Fatal("withdrawn neural intent did not cancel travel")
	}
}

func TestV9BrainUpgradePreservesLearnedAndGrownWiring(t *testing.T) {
	g := emptyGenome(initialHidden)
	g.WIn = make(weights, healthInputs*g.Hidden)
	g.WIn[len(g.WIn)-1] = 1.25
	m := newMind(g)
	m.DIn = make(weights, len(g.WIn))
	m.EIn = make(weights, len(g.WIn))
	m.Grown = 2
	m.GIn = make(weights, 2*healthInputs)
	m.EGIn = make(weights, 2*healthInputs)
	m.GIn[healthInputs-1] = 0.5
	m.GIn[2*healthInputs-1] = 0.75
	g.upgrade(9)
	m.upgrade(g)
	if len(g.WIn) != NumInputs*g.Hidden || g.WIn[healthInputs*g.Hidden-1] != 1.25 || g.WIn[healthInputs*g.Hidden] != 0 {
		t.Fatal("inherited input wiring corrupted")
	}
	if m.GIn[healthInputs-1] != 0.5 || m.GIn[NumInputs+healthInputs-1] != 0.75 || m.GIn[healthInputs] != 0 {
		t.Fatal("grown neuron wiring corrupted")
	}
}
