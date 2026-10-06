package sim

import (
	"math"

	"miniv2/backend/internal/ecology"
)

const socialRange = 3.0 // tiles within which family, strangers and reputation are felt

// childHungry is the energy (or hydration) below which a child's hunger (or
// thirst) is felt by its family.
const childHungry = 0.6

// sense fills the creature's input vector from what it can see and feel.
func (s *Sim) sense(c *Creature) {
	in := &c.input
	*in = [NumInputs]float64{}
	t := s.terrain
	// The abstract night halves how far anyone can see.
	vision := c.Genome.Traits.Vision * (0.55 + 0.45*s.eco.Light())
	hasTool := s.toolBonus(c) > 0
	// Someone wading, swimming or climbing looks out from the water or rock
	// they stand in; only what lies beyond it can hide the land.
	own, onLand := t.indexAt(c.X, c.Y)
	inside := onLand && t.blocked[own]
	// The richest food in sight, to remember (forage.go).
	seen, seenAt := 0.0, -1

	for r := range numRays {
		a := c.Heading + rayDegrees[r]*math.Pi/180
		dx, dy := math.Cos(a), math.Sin(a)
		food, res := 0.0, 0.0
		for d := 0.5; d <= vision; d += 0.5 {
			near := 1 - d/vision
			i, ok := t.indexAt(c.X+dx*d, c.Y+dy*d)
			if inside && ok && i == own {
				continue
			}
			if ok && t.fresh[i] {
				// Open water is plain to see; a dry riverbed may still hold
				// water in its sand; an empty one is just a ditch.
				switch s.eco.WaterState(i) {
				case ecology.WaterFlowing, ecology.WaterPools:
					in[inWater+r] = near
				case ecology.WaterUnder:
					in[inWater+r] = near * 0.5
				default:
					in[inObstacle+r] = near
				}
				break
			}
			if !ok || t.blocked[i] {
				in[inObstacle+r] = near
				break
			}
			f := math.Max(float64(s.eco.Forage(i)), s.meatFood(i)) // meat left at a kill (sharing.go)
			if p := s.eco.PlotAt(i); p != nil && p.IsRipe() && s.mayHarvest(c, p) {
				f = 1 // a ripe field is the most food there is
			}
			food = math.Max(food, f*near)
			if f > seen {
				seen, seenAt = f, i
			}
			res = math.Max(res, s.resourceAt(c, i, hasTool)*near)
		}
		in[inFood+r] = food
		in[inResource+r] = res
	}
	s.senseRememberedWater(c)

	partner, family, stranger, teacher, student := 0.0, 0.0, 0.0, 0.0, 0.0
	nearestD, nearestRep := socialRange, 0.0
	childHunger := 0.0
	var mate *Creature // the possible mate inPartnerNear feels (genetics.go)
	mateD := partnerRange
	lookX, lookY := math.Cos(c.Heading), math.Sin(c.Heading)
	farSq := vision * vision * 1.000001 // beyond this, d > vision whatever the rounding
	s.grid.near(c.X, c.Y, vision, func(o *Creature) {
		if o == c || o.Health <= 0 {
			return
		}
		dx, dy := o.X-c.X, o.Y-c.Y
		// Most neighbours are out of sight; tell cheaply before measuring.
		d2 := dx*dx + dy*dy
		if d2 > farSq || !s.opts.NoPerception && d2 > 0.6401 && dx*lookX+dy*lookY < 0 {
			return
		}
		d := math.Hypot(dx, dy)
		if d > vision || d == 0 {
			return
		}
		if !s.opts.NoPerception {
			if d > 0.8 && dx*lookX+dy*lookY < d*0.2588190451 {
				return
			}
			if !s.lineOfSight(c.X, c.Y, o.X, o.Y) {
				return
			}
		}
		if o.Sex != c.Sex && d < partnerRange {
			partner = math.Max(partner, 1-d/partnerRange)
			if d < mateD {
				mate, mateD = o, d
			}
		}
		if d < socialRange {
			if s.kin(c, o) {
				family = math.Max(family, 1-d/socialRange)
				// A hungry or thirsty child of the family calls on the grown-ups.
				if o != c && !s.adult(o) && s.adult(c) && s.carrier(o) == nil {
					childHunger = math.Max(childHunger, (1-d/socialRange)*childNeed(o))
				}
			} else {
				stranger = math.Max(stranger, 1-d/socialRange)
			}
			if d < nearestD {
				nearestD, nearestRep = d, s.opinion(c, o)
			}
			if t, st := s.teacherOrStudentNear(c, o); t || st {
				near := 1 - d/socialRange
				if t {
					teacher = math.Max(teacher, near)
				}
				if st {
					student = math.Max(student, near)
				}
			}
		}
		// Each ray covers ±15° around its angle, so the rays tile -75°..75°.
		rel := normAngle(math.Atan2(dy, dx)-c.Heading) * 180 / math.Pi
		if rel < -75 || rel > 75 {
			return
		}
		r := min(numRays-1, int((rel+75)/30))
		ch := inSame
		if o.Sex != c.Sex {
			ch = inMate
		}
		in[ch+r] = math.Max(in[ch+r], 1-d/vision)
	})

	// Animals, seen like people; a hunting tiger nearby is felt as danger.
	predator, hungry := 0.0, 0.0
	s.eco.Near(c.X, c.Y, vision, func(a *ecology.Animal) {
		dx, dy := a.X-c.X, a.Y-c.Y
		d := math.Hypot(dx, dy)
		if d > vision || d == 0 {
			return
		}
		if !s.opts.NoPerception && !s.sees(c, a.X, a.Y) {
			return
		}
		near := 1 - d/vision
		if a.Kind().Diet == ecology.Predator {
			predator = math.Max(predator, near)
		}
		if a.Owner != 0 && a.Owner == c.HouseID && a.Energy < livestockHungry {
			hungry = math.Max(hungry, near)
		}
		rel := normAngle(math.Atan2(dy, dx)-c.Heading) * 180 / math.Pi
		if rel < -75 || rel > 75 {
			return
		}
		r := min(numRays-1, int((rel+75)/30))
		in[inAnimal+r] = math.Max(in[inAnimal+r], near)
	})

	here, _ := t.indexAt(c.X, c.Y)
	carried := c.Inventory.count()
	in[inEnergy] = c.Energy
	in[inHydration] = c.Hydration
	in[inHealth] = c.Health
	in[inAge] = s.age(c) / c.Genome.Traits.Lifespan
	in[inFertile] = b2f(s.fertile(c))
	in[inPregnant] = b2f(c.Pregnancy != nil)
	in[inChildHungry] = childHunger
	in[inNursing] = b2f(s.nursing(c))
	in[inSick] = sickness(c)
	in[inFoodHere] = math.Max(math.Min(1, float64(s.foodAround(int(math.Floor(c.X)), int(math.Floor(c.Y))))), s.meatHere(c))
	in[inWaterNear] = b2f(s.canDrink(c))
	in[inSourceHere] = s.resourceAt(c, here, hasTool)
	in[inBumped] = b2f(c.Bumped)
	in[inPartnerNear] = partner
	in[inFamilyNear] = family
	in[inStrangerNear] = stranger
	in[inNearRep] = nearestRep
	in[inAttacked] = b2f(c.Hurt > 0)
	food := s.cat.foodUnits(c.Inventory)
	in[inCarryFood] = math.Min(1, float64(s.mealsInHand(c))/5)
	in[inCarryMaterial] = float64(carried-food) / invCapacity
	in[inCarryWeapon] = b2f(s.weaponBonus(c) > 0)
	in[inOwnHouseNear], in[inOtherHouseNear] = s.houseCloseness(c, vision)
	in[inCanCraft] = b2f(c.CanCraft)
	in[inCanBuild] = b2f(c.CanBuild)
	in[inTeacherNear] = teacher
	in[inStudentNear] = student
	in[inBestSkill] = s.bestSkill(c)
	in[inReward] = c.Mind.Reward
	if lib := s.libraryNear(c.X, c.Y, vision); lib != nil {
		in[inLibraryNear] = 1 - lib.dist(c.X, c.Y)/vision
	}
	in[inSeason] = s.eco.Season()
	in[inLight] = s.eco.Light()
	if id, _ := s.plantChoice(c); id != "" {
		in[inCanPlant] = 1
	}
	if _, d, ok := s.eco.RipeNear(c.X, c.Y, vision, func(p *ecology.Plot) bool { return s.mayHarvest(c, p) }); ok {
		in[inCropReady] = 1 - d/vision
	}
	in[inLivestockHungry] = hungry
	in[inPredatorNear] = predator
	in[inClock] = math.Sin(2*math.Pi*s.time()/8 + float64(c.ID))
	in[inNoise] = s.rng.Float64()*2 - 1
	in[inBias] = 1
	in[inAlarm] = s.alarm(c)
	s.senseMate(c, in, mate)
	s.senseStimuli(c, in)
	s.senseAffect(c, in)
	s.senseExchange(c, in)
	s.senseVillage(c, in)
	s.senseFire(c, in)
	s.senseTerrain(c, in)
	s.senseNutrition(c, in)
	if seenAt >= 0 {
		s.noteFood(c, float64(seenAt%t.w)+0.5, float64(seenAt/t.w)+0.5, seen)
	} else {
		s.noteFood(c, 0, 0, 0)
	}
	s.senseFoodMemory(c, in)
	if room := s.waterRoom(c); room > 0 {
		in[inCarryWater] = math.Min(1, c.WaterCarried/room)
	}
}

// waterMemory is how strongly a remembered river pulls when none is in
// sight: as strongly as water seen close by, since people know where their
// river is.
const waterMemory = 1.0

// senseRememberedWater: with no fresh water in sight, the way back to where
// c last drank shows faintly on the water rays (on the outermost ray when it
// lies behind), so the brain's pull towards water can bring it home.
func (s *Sim) senseRememberedWater(c *Creature) {
	in := &c.input
	if c.WaterX == 0 && c.WaterY == 0 {
		return
	}
	for r := range numRays {
		if in[inWater+r] > 0 {
			return
		}
	}
	dx, dy := c.WaterX-c.X, c.WaterY-c.Y
	if math.Abs(dx)+math.Abs(dy) < 1.5 && !s.waterAt(c.WaterX, c.WaterY) {
		// Back where it last drank, and the water there has gone: so has the
		// memory, and it must look for water elsewhere.
		c.WaterX, c.WaterY = 0, 0
		return
	}
	if math.Abs(dx)+math.Abs(dy) < 1 {
		return
	}
	rel := normAngle(math.Atan2(dy, dx)-c.Heading) * 180 / math.Pi
	r := min(numRays-1, max(0, int((rel+75)/30)))
	in[inWater+r] = waterMemory
}

// houseCloseness reports how close the creature's own house and the nearest
// house of another family with something in storage are (0 = out of sight).
func (s *Sim) houseCloseness(c *Creature, vision float64) (own, other float64) {
	for _, st := range s.structures {
		if !st.house() {
			continue
		}
		dx, dy := float64(st.X)+0.5-c.X, float64(st.Y)+0.5-c.Y
		if math.Abs(dx) > vision || math.Abs(dy) > vision {
			continue
		}
		near := 1 - math.Hypot(dx, dy)/vision
		if near <= 0 {
			continue
		}
		switch {
		case st.ID == c.HouseID:
			own = near
		case st.Owner != 0 && len(st.Storage) > 0 && !s.kinHouse(c, st):
			other = math.Max(other, near)
		}
	}
	return own, other
}
