package sim

import (
	"math"
	"testing"
)

func TestFecundityByAge(t *testing.T) {
	if fecundability(25, 50) <= fecundability(37, 50) || fecundability(37, 50) <= fecundability(44, 50) {
		t.Error("fecundability should fall from the twenties to the forties")
	}
	if fecundability(51, 50) != 0 || fecundability(14, 50) != 0 {
		t.Error("no conceptions before 15 or after menopause")
	}
	if p := fecundability(16, 50); p >= fecundPeak {
		t.Errorf("adolescents should be subfecund: %.3f", p)
	}
}

func TestOldAgeHazardDoublesEveryEightYears(t *testing.T) {
	a, b := senescenceHazard(60, refLife), senescenceHazard(68, refLife)
	if r := b / a; math.Abs(r-2) > 0.05 {
		t.Errorf("hazard grew %.2f× in 8 years, want about 2×", r)
	}
	// The modal age of adult death is where the hazard equals its growth rate (about 72–76).
	mode := math.Log(senB/senA) / senB
	if mode < 70 || mode > 78 {
		t.Errorf("modal age of adult death %.0f, want in the seventies", mode)
	}
}

func TestNursingDelaysTheNextConception(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	mother := person(s, Female, "Ibu", x, y)
	mother.BornTick = s.tick - int64(25*SecondsPerYear/dt)
	baby := person(s, Male, "Bayi", x, y)
	baby.BornTick = s.tick
	mother.NursingID = baby.ID
	if s.fertile(mother) {
		t.Fatal("a mother nursing a newborn should not be able to conceive")
	}
	baby.BornTick = s.tick - int64(2*SecondsPerYear/dt)
	if !s.fertile(mother) || s.nursingBlock(mother) != 0.4 {
		t.Fatalf("a two-year-old is half weaned: fertile %v, block %.1f", s.fertile(mother), s.nursingBlock(mother))
	}
	baby.Health = 0
	if s.nursingBlock(mother) != 0 {
		t.Error("a mother whose baby died should soon be fertile again")
	}
}
