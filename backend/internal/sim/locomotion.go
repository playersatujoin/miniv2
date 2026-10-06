package sim

import (
	"math"

	"miniv2/backend/internal/chem"
)

// Locomotion: how a body gets over the land and the water, after RAGE's
// navigation capabilities and water physics (Peds/NavCapabilities.h,
// physics/Floater.h, task/Motion/Locomotion/TaskInWater.h, task/Movement/
// Climbing/): wading in the shallows, swimming out of one's depth (tiring,
// and drowning when spent), floating on a raft, climbing steep ground and
// boulders slowly and sometimes falling. The brain only steers and sets the
// pace; this file applies the body's limits.
//
// What a body may enter is its navigation capability (navCaps, after
// CPedNavCapabilityInfo's FLAG_MAY_ENTER_WATER and FLAG_MAY_CLIMB): a small
// child only walks, an older one wades, from about six anyone swims and
// climbs. Where it is decides the medium; the medium sets the speed, the
// effort, how fast it tires and what can go wrong, like CTaskMotionSwimming's
// breath and struggle timers and the climb detector's slips.

// Locomotion is a body's state on land or water. The zero value is on foot
// and rested, which is what an older save loads with.
type Locomotion struct {
	Medium string `json:"medium,omitempty"` // "" on foot, wading, swimming, rafting, climbing
	// Fatigue is how spent the body is from swimming or climbing: 0 rested,
	// 1 spent (a spent swimmer struggles and drowns). It wears off on land.
	Fatigue float64 `json:"fatigue,omitempty"`
	// Down is how many seconds a fall still keeps them on the ground.
	Down float64 `json:"down,omitempty"`
	// Turn and Speed are the body's turn rate (rad/s) and pace (tiles/s)
	// now; they ease towards what the brain wants (see act).
	Turn  float64 `json:"turn,omitempty"`
	Speed float64 `json:"speed,omitempty"`
}

// WaterInfo summarises travel over water and steep ground for the world info.
type WaterInfo struct {
	Wading     int `json:"wading"`
	Swimming   int `json:"swimming"`
	Rafting    int `json:"rafting"`
	Climbing   int `json:"climbing"`
	Struggling int `json:"struggling"` // swimmers who are spent
	Fallen     int `json:"fallen"`     // down after a fall just now
	Rafts      int `json:"rafts"`      // people carrying a raft
	Drowned    int `json:"drowned"`    // all drownings so far
	Falls      int `json:"falls"`      // all deaths from falls so far
}

// How a body gets about, as saved in Locomotion.Medium.
const (
	onFoot   = ""
	wading   = "wading"
	swimming = "swimming"
	rafting  = "rafting"
	climbing = "climbing"
)

// The raft (chem/items.go): logs or bamboo lashed with rope. People reached
// Sahul over open sea at least 50,000 years ago (O'Connell et al. 2018;
// Clarkson et al. 2017 argue 65,000), and hominins were on Flores about a
// million years ago (Brumm et al. 2010), so it is a stone-age invention.
const (
	raftItem chem.ItemID = "rakit"
	raftTech             = "pelayaran"
)

const (
	// Ages (years) from which a child's body may do more. Children walk
	// steadily by two or three; in societies living by the water they swim
	// and scramble up rocks by five or six (estimates).
	wadeAge = 3.0
	swimAge = 6.0
	// staminaAge is when a child has an adult's endurance in the water.
	staminaAge = 14.0

	// Speed in each medium relative to walking. Self-chosen walking speed in
	// waist-deep water is about half that on land or less (Barela et al.
	// 2006). A recreational breaststroke, about 0.5 m/s, is a quarter of a
	// brisk walk. Paddled rafts in voyage reconstructions made 2–3 km/h
	// (Kaifu et al. 2020), about half a walking pace; there is no wind help
	// yet. Tobler's hiking function gives a third of the flat speed up a
	// slope of about 17°. A raft carried overland (logs or bamboo, tens of
	// kilograms) slows its bearer.
	wadeFactor      = 0.4
	swimFactor      = 0.25
	raftFactor      = 0.5
	climbFactor     = 0.3
	struggleFactor  = 0.5 // a spent swimmer makes little headway
	raftCarryFactor = 0.8

	// Effort relative to walking at the same exertion, from the Compendium
	// of Physical Activities (Ainsworth et al. 2011; brisk walking 3.5 MET):
	// water walking 4.5, leisurely swimming 6–8, light canoeing 2.8–5.8,
	// climbing hills and rock 6.3–7.5 MET.
	wadeEffort  = 1.3
	swimEffort  = 2.0
	raftEffort  = 1.15
	climbEffort = 2.1
	// treadEffort is the least effort that keeps a swimmer afloat (treading
	// water is about half as hard as swimming).
	treadEffort = 0.4
	// Water draws heat from a body some 25 times faster than air of the same
	// temperature, and even tropical water (26–30 °C) is below the 33–35 °C a
	// still body is comfortable in, so time in it costs energy even standing
	// still (Tipton & Bradford 2014): a share of the basal cost, more for a
	// body in the water to the neck than one in it to the waist.
	wadeChill = 0.3
	swimChill = 0.6

	// Seconds a fit adult can swim or climb flat out before spent, and the
	// fatigue shed per second on land (three times as fast resting), in the
	// shallows or on a raft at half that. Estimates on the world's
	// compressed clock: an adult swims across a few tiles of deep water and
	// back, not out to sea.
	swimEndurance  = 14.0
	climbEndurance = 40.0
	recoverRate    = 0.2
	// drownRate is the health a spent swimmer loses per second: a few
	// seconds of struggling, which is all a drowning takes.
	drownRate = 0.25

	// slipBase is the chance per second of climbing flat out that a fit
	// young adult slips; age, illness, wounds, frailty, a load and fatigue
	// raise it (an estimate: accidents, falls among them, are a few percent
	// of deaths among foragers, Gurven & Kaplan 2007, and falls are the
	// leading injury death in old age, WHO 2021). A fall hurts by
	// fallMin + fallSpread·u³: mostly bruises, now and then a broken body.
	slipBase   = 0.01
	fallMin    = 0.15
	fallSpread = 0.9
	fallDown   = 1.0 // seconds a fall keeps someone down
	raftLoad   = 0.5 // a raft weighs on a climber like half a full load

	// climbLook is how far ahead (tiles) steep ground is sensed.
	climbLook = 3.0
)

// Salts for hash01 so slips and the harm they do are independent draws.
const (
	slipSalt = 0x5115_0001
	hurtSalt = 0x5115_0002
)

// navCaps is what a body may do to get somewhere.
type navCaps uint8

const (
	mayWade navCaps = 1 << iota
	maySwim
	mayClimb
	// For planning routes only (routeCaps): carries a raft, so deep water is
	// plain travel; fit enough to swim the longer gaps.
	mayRaft
	maySwimFar
)

// navCaps reports what c's body may enter: nothing but dry ground in a world
// without mobility.
func (s *Sim) navCaps(c *Creature) navCaps {
	if s.opts.NoMobility {
		return 0
	}
	age := s.ageYears(c)
	var k navCaps
	if age >= wadeAge {
		k |= mayWade
	}
	if age >= swimAge {
		k |= maySwim | mayClimb
	}
	return k
}

// routeCaps adds the raft to c's capabilities, for planning routes.
func (s *Sim) routeCaps(c *Creature) navCaps {
	k := s.navCaps(c)
	if k&maySwim != 0 && c.Inventory[raftItem] > 0 {
		k |= mayRaft
	}
	return k
}

// allows reports whether a body with these capabilities may be on a tile.
func (k navCaps) allows(m mobility) bool {
	switch m {
	case mobWalk:
		return true
	case mobWade:
		return k&mayWade != 0
	case mobDeep:
		return k&(maySwim|mayRaft) != 0
	case mobClimb:
		return k&mayClimb != 0
	}
	return false
}

// move steps the creature forward, sliding along walls. It reports whether
// the full move succeeded.
func (s *Sim) move(c *Creature, dist float64) bool {
	if !s.opts.NoMobility {
		// However hard a child hurries after its mother, water and slopes
		// hold it back like everyone else, and a fallen body stays put.
		if f := s.locomotionFactor(c); f < 1 {
			dist = math.Min(dist, c.Genome.Traits.MaxSpeed*f*dt)
			if dist <= 0 {
				return true
			}
		}
	}
	dx, dy := math.Cos(c.Heading)*dist, math.Sin(c.Heading)*dist
	r := bodyRadius * c.Genome.Traits.Size
	if s.free(c, c.X+dx, c.Y+dy, dx, dy, r) {
		c.X += dx
		c.Y += dy
		return true
	}
	if s.free(c, c.X+dx, c.Y, dx, 0, r) {
		c.X += dx
	} else if s.free(c, c.X, c.Y+dy, 0, dy, r) {
		c.Y += dy
	}
	return false
}

// free checks whether c's body fits at (x, y): the centre and the leading
// edge when moving by (dx, dy).
func (s *Sim) free(c *Creature, x, y, dx, dy, r float64) bool {
	if !s.standable(c, x, y) {
		return false
	}
	l := math.Hypot(dx, dy)
	return l == 0 || s.standable(c, x+dx/l*r, y+dy/l*r)
}

// standable reports whether c's body may be at (x, y): on dry ground, or in
// water or on a slope its body can manage.
func (s *Sim) standable(c *Creature, x, y float64) bool {
	i, ok := s.terrain.indexAt(x, y)
	if !ok {
		return false
	}
	if !s.terrain.blocked[i] {
		return true
	}
	return !s.opts.NoMobility && s.navCaps(c).allows(s.terrain.rough[i])
}

// shoveFree is free for personal space: people are eased apart onto dry
// ground, or within the water or the slope they are already in, never
// pushed off the bank or up a rock.
func (s *Sim) shoveFree(c *Creature, x, y, dx, dy, r float64) bool {
	if !s.free(c, x, y, dx, dy, r) {
		return false
	}
	if s.opts.NoMobility {
		return true
	}
	there, _ := s.terrain.indexAt(x, y)
	to := s.terrain.mobilityAt(there)
	if to == mobWalk {
		return true
	}
	here, ok := s.terrain.indexAt(c.X, c.Y)
	return ok && s.terrain.mobilityAt(here) == to
}

// mediumAt is how c gets about where it stands now.
func (s *Sim) mediumAt(c *Creature) string {
	i, ok := s.terrain.indexAt(c.X, c.Y)
	if !ok {
		return onFoot
	}
	switch s.terrain.mobilityAt(i) {
	case mobWade:
		return wading
	case mobDeep:
		if c.Inventory[raftItem] > 0 {
			return rafting
		}
		return swimming
	case mobClimb:
		return climbing
	}
	return onFoot
}

// locomotionFactor scales c's speed for the ground or water underfoot.
func (s *Sim) locomotionFactor(c *Creature) float64 {
	if s.opts.NoMobility {
		return 1
	}
	if c.Loco.Down > 0 {
		return 0
	}
	switch s.mediumAt(c) {
	case wading:
		return wadeFactor
	case swimming:
		if c.Loco.Fatigue >= 1 {
			return swimFactor * struggleFactor
		}
		return swimFactor
	case rafting:
		return raftFactor
	case climbing:
		return climbFactor
	}
	if c.Inventory[raftItem] > 0 {
		return raftCarryFactor
	}
	return 1
}

// effortOf is how much harder than walking each medium is.
func effortOf(medium string) float64 {
	switch medium {
	case wading:
		return wadeEffort
	case swimming:
		return swimEffort
	case rafting:
		return raftEffort
	case climbing:
		return climbEffort
	}
	return 0
}

// locomotionStep runs after c moved this tick: breath, tiring in the water,
// drowning, slips and falls, the raft.
func (s *Sim) locomotionStep(c *Creature) {
	if s.opts.NoMobility {
		return
	}
	l := &c.Loco
	if l.Down > 0 {
		l.Down = math.Max(0, l.Down-dt)
	}
	l.Medium = s.mediumAt(c)
	effort := 0.0
	if !c.resting {
		effort = clamp(c.output[outMove], 0, 1)
	}
	if l.Medium == swimming {
		effort = math.Max(effort, treadEffort)
	}
	// metabolize charged the move as a walk at this pace; water and slopes
	// cost more for the same exertion, and water chills.
	if ratio := effortOf(l.Medium); ratio > 0 {
		tr := &c.Genome.Traits
		speed := c.Pace * tr.MaxSpeed
		extra := math.Max(0, moveCost*tr.Size*(ratio*effort*effort*tr.MaxSpeed*tr.MaxSpeed-speed*speed))
		switch l.Medium {
		case wading:
			extra += wadeChill * basalCost * tr.Size
		case swimming:
			extra += swimChill * basalCost * tr.Size
		}
		c.Energy -= extra * hungerScale * s.bodyScale(c) * dt
	}
	switch l.Medium {
	case swimming:
		l.Fatigue = math.Min(1, l.Fatigue+effort/(swimEndurance*s.stamina(c))*dt)
		if l.Fatigue >= 1 {
			s.struggle(c)
		}
	case climbing:
		l.Fatigue = math.Min(1, l.Fatigue+effort/(climbEndurance*s.stamina(c))*dt)
		if (s.tick+c.ID)%TicksPerSecond == 0 {
			s.maybeSlip(c, effort)
		}
	case wading, rafting:
		l.Fatigue = math.Max(0, l.Fatigue-recoverRate*0.5*dt)
	default:
		rate := recoverRate
		if c.resting {
			rate *= 3
		}
		l.Fatigue = math.Max(0, l.Fatigue-rate*dt)
	}
}

// stamina is how long c lasts in the water or on a slope relative to a fit
// adult: less for a child, the old, the sick, the hungry and the laden.
func (s *Sim) stamina(c *Creature) float64 {
	f := s.vigor(c) * (1 - 0.5*sickness(c)) * (0.4 + 0.6*clamp(c.Energy/0.5, 0, 1))
	if age := s.ageYears(c); age < staminaAge {
		f *= 0.4 + 0.6*clamp((age-swimAge)/(staminaAge-swimAge), 0, 1)
	}
	f *= 1 - 0.4*s.load(c)
	return math.Max(f, 0.05)
}

// load is how laden c is (0–1): what it carries, a raft weighing extra.
func (s *Sim) load(c *Creature) float64 {
	l := float64(c.Inventory.count()) / invCapacity
	if c.Inventory[raftItem] > 0 {
		l += raftLoad
	}
	return math.Min(1, l)
}

// struggle: a spent swimmer goes under. Health drains, those nearby see
// someone in trouble, and if it runs out the swimmer has drowned.
func (s *Sim) struggle(c *Creature) {
	c.Health -= drownRate * dt
	if (s.tick+c.ID)%TicksPerSecond == 0 {
		s.addStimulus(stimDrowning, c.X, c.Y, 1, c)
		s.feel(c, moodFear, 0.5)
	}
	if c.Health <= 0 {
		c.Health = 0
		if c.fate == "" {
			c.fate = "drowned"
		}
	}
}

// slipRisk scales slipBase for c, climbing at this effort.
func (s *Sim) slipRisk(c *Creature, effort float64) float64 {
	r := 0.3 + 0.7*effort // holding still on a slope is safer than scrambling
	switch age := s.ageYears(c); {
	case age < 10:
		r *= 1.5
	case age > 50:
		r *= 1 + 0.04*(age-50)
	}
	r *= (1 + 3*sickness(c)) * (1 + 2*(1-c.Health)) / s.vigor(c)
	r *= (1 + s.load(c)) * (1 + c.Loco.Fatigue)
	return r
}

// maybeSlip checks once a second whether a climber slips, and if so the
// fall: hurt, down for a moment, seen by those around, perhaps killed.
func (s *Sim) maybeSlip(c *Creature, effort float64) {
	if hash01(c.ID, s.tick, slipSalt) >= slipBase*s.slipRisk(c, effort) {
		return
	}
	u := hash01(c.ID, s.tick, hurtSalt)
	hurt := fallMin + fallSpread*u*u*u
	if age := s.ageYears(c); age > 50 {
		hurt *= 1 + 0.02*(age-50) // old bones break
	}
	c.Health -= hurt
	c.Loco.Down = fallDown
	s.addStimulus(stimFall, c.X, c.Y, clamp(2*hurt, 0.3, 1), c)
	s.feel(c, moodFear, 0.4)
	if c.Health <= 0 {
		c.Health = 0
		if c.fate == "" {
			c.fate = "fall"
		}
	}
}

// bodyOf is whose body carries c about: its mother's while it is a baby in
// arms, else its own.
func (s *Sim) bodyOf(c *Creature) *Creature {
	if m := s.carrier(c); m != nil {
		return m
	}
	return c
}

// senseTerrain fills inInWater, inOnRaft, inSeaAhead and inClimbAhead.
func (s *Sim) senseTerrain(c *Creature, in *[NumInputs]float64) {
	if s.opts.NoMobility {
		return
	}
	t := s.terrain
	here, ok := t.indexAt(c.X, c.Y)
	if !ok {
		return
	}
	switch t.mobilityAt(here) {
	case mobWade:
		in[inInWater] = 0.5
	case mobDeep:
		in[inInWater] = 1
	}
	if s.bodyOf(c).Loco.Medium == rafting {
		in[inOnRaft] = 1
	}
	// Look along the heading, as far as one can see, for deep water and for
	// steep ground close ahead. Walls, trees, rocks and slopes hide what lies
	// beyond them, as they do for the eyes (perception.go).
	vision := c.Genome.Traits.Vision * (0.55 + 0.45*s.eco.Light())
	dx, dy := math.Cos(c.Heading), math.Sin(c.Heading)
	for d := 0.5; d <= vision; d += 0.5 {
		i, ok := t.indexAt(c.X+dx*d, c.Y+dy*d)
		if !ok {
			return
		}
		if i == here {
			continue
		}
		switch t.mobilityAt(i) {
		case mobNever:
			return
		case mobClimb:
			if d <= climbLook {
				in[inClimbAhead] = 1 - d/(climbLook+0.5)
			}
			return
		case mobDeep:
			if in[inSeaAhead] == 0 {
				in[inSeaAhead] = 1 - d/vision
			}
			if d > climbLook {
				return
			}
		}
	}
}

// locomotionFlags are the stream flags for wading, swimming, rafting,
// climbing and having just fallen.
func (s *Sim) locomotionFlags(c *Creature) int {
	f := 0
	switch c.Loco.Medium {
	case wading:
		f = flagWading
	case swimming:
		f = flagSwimming
	case rafting:
		f = flagRafting
	case climbing:
		f = flagClimbing
	}
	if c.Loco.Down > 0 {
		f |= flagFallen
	}
	return f
}

// locomotionName names how c is getting about, for the inspector ("" on foot).
func (s *Sim) locomotionName(c *Creature) string { return c.Loco.Medium }

func (s *Sim) waterInfo() WaterInfo {
	w := WaterInfo{Drowned: s.deathsBy.Drowned, Falls: s.deathsBy.Fall}
	for _, c := range s.creatures {
		switch c.Loco.Medium {
		case wading:
			w.Wading++
		case swimming:
			w.Swimming++
			if c.Loco.Fatigue >= 1 {
				w.Struggling++
			}
		case rafting:
			w.Rafting++
		case climbing:
			w.Climbing++
		}
		if c.Loco.Down > 0 {
			w.Fallen++
		}
		if c.Inventory[raftItem] > 0 {
			w.Rafts++
		}
	}
	return w
}

// raftWant is how much c would like a raft of its own (for wants() in
// economy.go): someone who can lash one together, has none, and has deep
// water in view.
func (s *Sim) raftWant(c *Creature) float64 {
	if s.opts.NoMobility || c.Inventory[raftItem] > 0 || c.input[inSeaAhead] <= 0 || !s.canPractise(c, raftTech) {
		return 0
	}
	return 0.5
}
