package sim

import (
	"math"
	"testing"
)

func TestFoodPlacesAreRememberedMergedAndForgotten(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Female, "Sari", x, y)
	s.rememberFood(c, 10, 10, 0.3)
	if len(c.FoodPlaces) != 0 {
		t.Fatal("a poor patch was remembered as a new place")
	}
	s.rememberFood(c, 10, 10, 0.8)
	s.rememberFood(c, 11, 10.5, 0.9) // the same patch
	if len(c.FoodPlaces) != 1 || c.FoodPlaces[0].Rich != 0.9 || c.FoodPlaces[0].X != 11 {
		t.Fatalf("places %+v", c.FoodPlaces)
	}
	for i := range maxFoodPlaces + 3 {
		s.rememberFood(c, 20+float64(i)*5, 20, 0.6+0.01*float64(i))
	}
	if len(c.FoodPlaces) != maxFoodPlaces {
		t.Fatalf("kept %d places", len(c.FoodPlaces))
	}
	// Knowledge fades with the years.
	p := c.FoodPlaces[0]
	now := s.foodPlaceWorth(p)
	s.tick += int64(2 * foodMemoryYears * SecondsPerYear * TicksPerSecond)
	if later := s.foodPlaceWorth(p); later >= now*0.2 {
		t.Fatalf("worth %.3f → %.3f after two memory spans", now, later)
	}
}

func TestEatenBarePlaceIsForgottenOnArrival(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Male, "Budi", x, y)
	c.FoodPlaces = []FoodPlace{{X: c.X, Y: c.Y, Rich: 0.9, Tick: s.tick}}
	// Nothing grows where he stands in the test land.
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if i, ok := s.terrain.index(x+dx, y+dy); ok {
				s.eco.TakeForage(i, 1)
			}
		}
	}
	s.noteFood(c, 0, 0, 0)
	if len(c.FoodPlaces) != 0 {
		t.Fatalf("an eaten-bare place is still trusted: %+v", c.FoodPlaces)
	}
}

func TestRememberedFoodIsSensedAndWalkedTo(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Female, "Ayu", x, y)
	c.Heading = 0                                                              // facing +x
	c.FoodPlaces = []FoodPlace{{X: c.X, Y: c.Y + 12, Rich: 0.9, Tick: s.tick}} // +y: to the right
	var in [NumInputs]float64
	s.senseFoodMemory(c, &in)
	if in[inFoodMemory] <= 0.2 || in[inFoodMemorySide] < 0.99 {
		t.Fatalf("food memory %.2f side %.2f", in[inFoodMemory], in[inFoodMemorySide])
	}
	// Nothing in sight to eat, so the memory is what she has to go on.
	for dy := -10; dy <= 10; dy++ {
		for dx := -10; dx <= 10; dx++ {
			if i, ok := s.terrain.index(x+dx, y+dy); ok {
				s.eco.TakeForage(i, 1)
			}
		}
	}
	goal, ok := s.travelTarget(c, ActEat)
	if !ok || math.Hypot(goal.X-c.X, goal.Y-(c.Y+12)) > 0.01 {
		t.Fatalf("no route to remembered food: %v %+v", ok, goal)
	}
}

func TestWaterTubesFillAtTheRiverAndQuenchAway(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Female, "Sri", x, y)
	c.Inventory.add(waterTube, 3)
	if s.waterRoom(c) != tubeHolds*maxTubes {
		t.Fatalf("room %.2f", s.waterRoom(c))
	}
	// Fill from a source (kind open, any tile index of fresh water).
	fresh := -1
	for i, f := range s.terrain.fresh {
		if f {
			fresh = i
			break
		}
	}
	if fresh < 0 {
		t.Skip("no fresh water on the test land")
	}
	for range 200 {
		s.fillWater(c, drinkOpen, fresh)
	}
	if math.Abs(c.WaterCarried-s.waterRoom(c)) > 1e-9 {
		t.Fatalf("carried %.3f of %.3f", c.WaterCarried, s.waterRoom(c))
	}
	c.Hydration = 0.4
	before := c.WaterCarried
	if !s.drinkCarried(c) || c.Hydration <= 0.4 || c.WaterCarried >= before {
		t.Fatal("did not drink from the tube")
	}
	c.Inventory.take(waterTube, 3)
	if s.drinkCarried(c) || c.WaterCarried != 0 {
		t.Fatal("drank from tubes they no longer have")
	}
}

func TestTurningAndPaceEase(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Male, "Adi", x, y)
	c.Genome.BOut[outMove] = 4
	c.Genome.BOut[outRest] = -10
	c.Heading = 0
	// A wavering mind: hard left, hard right, every thought.
	var headings []float64
	for i := range 40 {
		c.Genome.BOut[outTurn] = float32(4 * (1 - 2*float64((i/4)%2)))
		s.tick++
		s.act(c)
		headings = append(headings, c.Heading)
	}
	maxStep := 0.0
	for i := 1; i < len(headings); i++ {
		maxStep = math.Max(maxStep, math.Abs(normAngle(headings[i]-headings[i-1])))
	}
	if maxStep >= maxTurnRate*dt*0.95 {
		t.Fatalf("heading jumped %.3f rad in a tick (limit %.3f)", maxStep, maxTurnRate*dt)
	}
	if c.Loco.Speed <= 0 || c.Loco.Speed > c.Genome.Traits.MaxSpeed {
		t.Fatalf("pace %.2f", c.Loco.Speed)
	}
}
