package sim

import (
	"math"
	"testing"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/world"
)

// mobilityWorld is a flat grass field ringed by deep sea, laid out by paint,
// with the fake chemistry and nobody in it yet.
func mobilityWorld(t *testing.T, size int, paint func(set func(x, y, ground, object int))) (*Sim, *world.Map) {
	t.Helper()
	m := testMap(t, size)
	set := func(x, y, ground, object int) {
		m.Layers.Ground[y*size+x], m.Layers.Objects[y*size+x] = ground, object
	}
	for y := range size {
		for x := range size {
			g := world.Grass
			if x < 2 || y < 2 || x >= size-2 || y >= size-2 {
				g = world.DeepWater
			}
			set(x, y, g, world.None)
		}
	}
	if paint != nil {
		paint(set)
	}
	s := newSimWith(m, 1, fakeCatalog())
	for _, c := range s.creatures {
		delete(s.byID, c.ID)
	}
	s.creatures = nil
	return s, m
}

// aged makes c the given number of years old.
func aged(s *Sim, c *Creature, years float64) *Creature {
	c.BornTick = s.tick - int64(math.Ceil(years*SecondsPerYear*TicksPerSecond))
	return c
}

// stepFor runs c's body (not its brain) for the given seconds, pushing on at
// full effort without getting anywhere.
func stepFor(s *Sim, c *Creature, seconds float64) {
	for range int(seconds * TicksPerSecond) {
		s.tick++
		c.output[outMove] = 1
		s.locomotionStep(c)
		if c.Health <= 0 {
			return
		}
	}
}

func TestWadingIsAllowedAndSlow(t *testing.T) {
	s, _ := mobilityWorld(t, 32, func(set func(x, y, g, o int)) {
		for y := 2; y < 30; y++ {
			set(16, y, world.Water, world.None)
		}
	})
	c := person(s, Male, "Wader", 15, 10)
	c.Heading = 0 // east, into the shallows
	if f := s.locomotionFactor(c); f != 1 {
		t.Fatalf("on dry ground the factor is %v", f)
	}
	for range 20 {
		s.move(c, 0.1)
	}
	if s.mediumAt(c) != wading {
		t.Fatalf("an adult should wade into the shallows, stands at %.2f", c.X)
	}
	s.locomotionStep(c)
	if c.Loco.Medium != wading || s.locomotionFactor(c) != wadeFactor {
		t.Fatalf("wading: medium %q factor %v", c.Loco.Medium, s.locomotionFactor(c))
	}
	// However hard one pushes, the water holds one back.
	x := c.X
	s.move(c, 1)
	if got := c.X - x; got > c.Genome.Traits.MaxSpeed*wadeFactor*dt+1e-9 {
		t.Fatalf("waded %.3f tiles in a tick", got)
	}

	s.opts.NoMobility = true
	d := person(s, Female, "Landbound", 15, 20)
	d.Heading = 0
	for range 20 {
		s.move(d, 0.1)
	}
	if s.terrain.blockedAt(d.X, d.Y) || d.X > 16 {
		t.Fatal("without mobility water is solid")
	}
}

func TestDeepWaterAndSlopesAreForTheOldEnough(t *testing.T) {
	s, _ := mobilityWorld(t, 32, func(set func(x, y, g, o int)) {
		set(10, 10, world.Water, world.None)
		set(11, 10, world.DeepWater, world.None)
		set(12, 10, world.Mountain, world.None)
		set(13, 10, world.Grass, world.Boulder)
		set(14, 10, world.Grass, world.Tree)
	})
	at := func(x int) (float64, float64) { return float64(x) + 0.5, 10.5 }
	cases := []struct {
		years                        float64
		wade, deep, climb, tree, off bool
	}{
		{2.5, false, false, false, false, false},
		{4, true, false, false, false, false},
		{7, true, true, true, false, false},
		{30, true, true, true, false, false},
		{30, false, false, false, false, true},
	}
	for _, tc := range cases {
		s.opts.NoMobility = tc.off
		c := aged(s, person(s, Female, "C", 5, 5), tc.years)
		for x, want := range map[int]bool{10: tc.wade, 11: tc.deep, 12: tc.climb, 13: tc.climb, 14: tc.tree} {
			px, py := at(x)
			if got := s.standable(c, px, py); got != want {
				t.Errorf("age %v (off %v) at tile %d: standable %v, want %v", tc.years, tc.off, x, got, want)
			}
		}
	}
}

func TestTiredSwimmerDrownsButRafterFloats(t *testing.T) {
	s, _ := mobilityWorld(t, 32, func(set func(x, y, g, o int)) {
		for y := 2; y < 30; y++ {
			for x := 12; x < 30; x++ {
				set(x, y, world.DeepWater, world.None)
			}
		}
	})
	swimmer := person(s, Male, "Swimmer", 20, 10)
	stepFor(s, swimmer, 8)
	if swimmer.Loco.Medium != swimming || swimmer.Loco.Fatigue <= 0 || swimmer.Loco.Fatigue >= 1 || swimmer.Health != 1 {
		t.Fatalf("after 8 s a fit swimmer tires but holds on: %+v health %v", swimmer.Loco, swimmer.Health)
	}
	if s.locomotionFlags(swimmer)&flagSwimming == 0 {
		t.Fatal("no swimming flag")
	}
	energy := swimmer.Energy
	stepFor(s, swimmer, 60)
	if swimmer.Health > 0 || swimmer.fate != "drowned" {
		t.Fatalf("a spent swimmer should drown: health %v fate %q", swimmer.Health, swimmer.fate)
	}
	if swimmer.Energy >= energy {
		t.Fatal("swimming cost nothing")
	}
	s.reap()
	if s.deathsBy.Drowned != 1 || s.waterInfo().Drowned != 1 {
		t.Fatalf("drowning not counted: %+v", s.deathsBy)
	}

	rafter := person(s, Female, "Rafter", 20, 20)
	rafter.Inventory.add(raftItem, 1)
	stepFor(s, rafter, 60)
	if rafter.Loco.Medium != rafting || rafter.Health != 1 || rafter.fate != "" || rafter.Loco.Fatigue != 0 {
		t.Fatalf("a rafter should float: %+v health %v fate %q", rafter.Loco, rafter.Health, rafter.fate)
	}
	if s.locomotionFactor(rafter) != raftFactor || s.locomotionFlags(rafter)&flagRafting == 0 {
		t.Fatal("rafting: wrong speed or flag")
	}
	if w := s.waterInfo(); w.Rafting != 1 || w.Rafts != 1 || w.Swimming != 0 {
		t.Fatalf("water info %+v", w)
	}

	// A swimmer who makes it back to shore recovers.
	back := person(s, Male, "Back", 20, 25)
	stepFor(s, back, 10)
	back.X = 8.5
	tired := back.Loco.Fatigue
	stepFor(s, back, 1)
	if back.Loco.Medium != onFoot || back.Loco.Fatigue >= tired {
		t.Fatalf("no recovery on land: %+v", back.Loco)
	}
}

func TestClimbingSlipsAreDeterministicAndCanKill(t *testing.T) {
	slope := func(set func(x, y, g, o int)) {
		for y := 2; y < 30; y++ {
			for x := 10; x < 20; x++ {
				set(x, y, world.Mountain, world.None)
			}
		}
	}
	// A frail old climber: every slip kills.
	fallTick := func() (int64, string) {
		s, _ := mobilityWorld(t, 32, slope)
		c := aged(s, person(s, Female, "Nenek", 15, 15), 70)
		c.Health = 0.12
		stepFor(s, c, 600)
		return s.tick, c.fate
	}
	tick1, fate1 := fallTick()
	tick2, fate2 := fallTick()
	if fate1 != "fall" {
		t.Fatalf("a frail climber never fell in 600 s (fate %q)", fate1)
	}
	if tick1 != tick2 || fate2 != fate1 {
		t.Fatalf("slips are not deterministic: tick %d vs %d", tick1, tick2)
	}

	// A fit adult slips now and then, is down for a moment, and lives.
	s, _ := mobilityWorld(t, 32, slope)
	c := person(s, Male, "Pemanjat", 15, 15)
	falls := 0
	for range 1000 {
		c.Health, c.Loco.Fatigue = 1, 0
		before := c.Health
		stepFor(s, c, 1)
		if c.Health < before {
			falls++
			if c.Loco.Down <= 0 || s.locomotionFactor(c) != 0 || s.locomotionFlags(c)&flagFallen == 0 {
				t.Fatal("a fall should leave the climber down")
			}
			c.Loco.Down = 0
		}
	}
	if falls < 3 || falls > 40 {
		t.Fatalf("%d slips in 1000 s of climbing; expected about 1 %%", falls)
	}
	if s.locomotionFlags(c)&flagClimbing == 0 || s.locomotionFactor(c) != climbFactor {
		t.Fatal("climbing: wrong flag or speed")
	}
}

func TestRestoreKeepsBodiesInWaterAndOnSlopes(t *testing.T) {
	s, m := mobilityWorld(t, 32, func(set func(x, y, g, o int)) {
		set(10, 10, world.Water, world.None)
		set(20, 10, world.Mountain, world.None)
		for x := 12; x < 18; x++ {
			set(x, 20, world.DeepWater, world.None)
		}
	})
	wader := person(s, Female, "Wader", 10, 10)
	climber := person(s, Male, "Climber", 20, 10)
	swimmer := person(s, Male, "Swimmer", 15, 20)
	stepFor(s, swimmer, 3)
	s.locomotionStep(wader)
	s.locomotionStep(climber)
	data, err := s.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	r, err := restoreWith(m, data, fakeCatalog())
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []*Creature{wader, climber, swimmer} {
		got := r.byID[c.ID]
		if got == nil || got.X != c.X || got.Y != c.Y || got.Loco != c.Loco {
			t.Fatalf("%s moved or changed on restore: %+v → %+v", c.Name, c, got)
		}
	}
	if r.byID[swimmer.ID].Loco.Medium != swimming || r.byID[swimmer.ID].Loco.Fatigue == 0 {
		t.Fatal("the swimmer's state was not saved")
	}
}

func TestPersonalSpaceKeepsPeopleOnTheBank(t *testing.T) {
	s, _ := mobilityWorld(t, 32, func(set func(x, y, g, o int)) {
		for y := 2; y < 30; y++ {
			set(16, y, world.Water, world.None)
		}
	})
	a := person(s, Male, "A", 15, 10)
	b := person(s, Female, "B", 15, 10)
	a.X, b.X = 15.9, 15.85 // a is pushed east, towards the water
	for range 40 {
		s.personalSpace()
	}
	if s.mediumAt(a) != onFoot || s.mediumAt(b) != onFoot {
		t.Fatal("personal space shoved someone off the bank")
	}
	// Two waders are eased apart where they stand.
	w1 := person(s, Male, "W1", 16, 20)
	w2 := person(s, Female, "W2", 16, 20)
	w1.Y = 20.4
	d := math.Abs(w1.Y - w2.Y)
	for range 20 {
		s.personalSpace()
	}
	if math.Abs(w1.Y-w2.Y) <= d || s.mediumAt(w1) != wading || s.mediumAt(w2) != wading {
		t.Fatal("waders were not eased apart in the water")
	}
}

func TestTerrainSenses(t *testing.T) {
	s, _ := mobilityWorld(t, 32, func(set func(x, y, g, o int)) {
		set(10, 10, world.Water, world.None)
		for x := 14; x < 18; x++ {
			set(x, 10, world.DeepWater, world.None)
		}
		set(12, 20, world.Mountain, world.None)
	})
	c := person(s, Female, "Looker", 10, 10)
	c.Heading = 0
	var in [NumInputs]float64
	s.senseTerrain(c, &in)
	if in[inInWater] != 0.5 || in[inOnRaft] != 0 || in[inSeaAhead] <= 0 || in[inClimbAhead] != 0 {
		t.Fatalf("wading, sea ahead: %v %v %v %v", in[inInWater], in[inOnRaft], in[inSeaAhead], in[inClimbAhead])
	}
	c.X, c.Y = 10.5, 20.5
	in = [NumInputs]float64{}
	s.senseTerrain(c, &in)
	if in[inInWater] != 0 || in[inSeaAhead] != 0 || in[inClimbAhead] <= 0 {
		t.Fatalf("dry, slope ahead: %v %v %v", in[inInWater], in[inSeaAhead], in[inClimbAhead])
	}
	c.X, c.Y = 15.5, 10.5
	c.Inventory.add(raftItem, 1)
	s.locomotionStep(c)
	in = [NumInputs]float64{}
	s.senseTerrain(c, &in)
	if in[inInWater] != 1 || in[inOnRaft] != 1 {
		t.Fatalf("afloat on a raft: %v %v", in[inInWater], in[inOnRaft])
	}
	s.opts.NoMobility = true
	in = [NumInputs]float64{}
	s.senseTerrain(c, &in)
	if in[inInWater] != 0 || in[inOnRaft] != 0 || in[inSeaAhead] != 0 || in[inClimbAhead] != 0 {
		t.Fatal("senses must stay 0 without mobility")
	}
}

func TestLocomotionFlagsAndNames(t *testing.T) {
	s, _ := mobilityWorld(t, 32, nil)
	c := person(s, Male, "F", 10, 10)
	for medium, flag := range map[string]int{wading: flagWading, swimming: flagSwimming, rafting: flagRafting, climbing: flagClimbing} {
		c.Loco = Locomotion{Medium: medium}
		if s.locomotionFlags(c) != flag || s.locomotionName(c) != medium {
			t.Errorf("%s: flags %b", medium, s.locomotionFlags(c))
		}
	}
	c.Loco = Locomotion{Down: 0.5}
	if s.locomotionFlags(c) != flagFallen || s.locomotionName(c) != "" {
		t.Error("fallen flag")
	}
}

func TestRaftIsInTheRealCatalog(t *testing.T) {
	k := defaultCatalog()
	if it := k.item(raftItem); it.Name == "" || it.Value <= 0 {
		t.Fatal("no raft item")
	}
	taught := false
	for _, r := range k.recipes {
		if r.Outputs[raftItem] > 0 && r.Teaches == raftTech && r.Inputs["kayu"] > 0 && r.Inputs["tali"] > 0 && r.Station == "" {
			taught = true
		}
	}
	if !taught {
		t.Fatal("no stone-age recipe makes a raft and teaches seafaring")
	}
	found := false
	for _, tc := range k.techs {
		found = found || tc.ID == raftTech && tc.Tier == 0
	}
	if !found {
		t.Fatal("seafaring is not a stone-age technology")
	}
	if _, ok := chem.ItemByID(raftItem); !ok {
		t.Fatal("chem has no raft")
	}
}

func TestCarryingARaftSlowsOnLand(t *testing.T) {
	s, _ := mobilityWorld(t, 32, nil)
	c := person(s, Male, "Porter", 10, 10)
	c.Inventory.add(raftItem, 1)
	if s.locomotionFactor(c) != raftCarryFactor {
		t.Fatal("a raft carried overland should slow its bearer")
	}
	if s.raftWant(c) != 0 {
		t.Fatal("someone with a raft wants no other")
	}
}
