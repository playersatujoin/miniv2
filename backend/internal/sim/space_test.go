package sim

import (
	"math"
	"testing"
)

func TestPersonalSpace(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	// Five people on exactly one spot, and two more nearly so.
	var crowd []*Creature
	for range 5 {
		crowd = append(crowd, person(s, Female, "Ana", x, y))
	}
	a := person(s, Male, "Budi", x+2, y)
	b := person(s, Male, "Cahya", x+2, y)
	b.X += 0.05
	crowd = append(crowd, a, b)
	for range 4 * TicksPerSecond {
		s.personalSpace()
	}
	for i, c := range crowd {
		for _, o := range crowd[i+1:] {
			if d, want := math.Hypot(c.X-o.X, c.Y-o.Y), 2*bodyRadius; d < want*0.97 {
				t.Errorf("%s #%d and %s #%d still %.2f apart, want at least %.2f", c.Name, c.ID, o.Name, o.ID, d, want)
			}
		}
	}
	// Nobody was pushed far: the five spread into a small ring.
	for _, c := range crowd[:5] {
		if d := math.Hypot(c.X-(float64(x)+0.5), c.Y-(float64(y)+0.5)); d > 1 {
			t.Errorf("%s pushed %.2f tiles away", c.Name, d)
		}
	}
}

func TestPersonalSpaceLeavesBabiesInArms(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	mother := person(s, Female, "Ibu", x, y)
	baby := person(s, Male, "Bayi", x, y)
	baby.Mother = &Ref{mother.ID, mother.Name}
	baby.BornTick = s.tick
	baby.X, baby.Y = mother.X+0.1, mother.Y
	for range TicksPerSecond {
		s.personalSpace()
	}
	if baby.X != mother.X+0.1 || baby.Y != mother.Y {
		t.Errorf("baby in arms was pushed: (%.3f,%.3f) vs mother (%.3f,%.3f)", baby.X, baby.Y, mother.X, mother.Y)
	}
}

func TestPersonalSpaceOff(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	s.opts.NoPersonalSpace = true
	a := person(s, Female, "Ana", x, y)
	b := person(s, Female, "Ani", x, y)
	s.personalSpace()
	if a.X != b.X || a.Y != b.Y {
		t.Error("bodies moved with personal space switched off")
	}
}
