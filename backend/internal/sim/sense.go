package sim

import (
	"math"

	"miniv2/backend/internal/chem"
)

const socialRange = 3.0 // tiles within which family, strangers and reputation are felt

// sense fills the creature's input vector from what it can see and feel.
func (s *Sim) sense(c *Creature) {
	in := &c.input
	*in = [NumInputs]float64{}
	t := s.terrain
	vision := c.Genome.Traits.Vision
	if c.wantVec == nil {
		s.updateWantVec(c)
	}
	hasTool := s.toolBonus(c) > 0

	for r := range numRays {
		a := c.Heading + rayDegrees[r]*math.Pi/180
		dx, dy := math.Cos(a), math.Sin(a)
		food, res := 0.0, 0.0
		for d := 0.5; d <= vision; d += 0.5 {
			near := 1 - d/vision
			i, ok := t.indexAt(c.X+dx*d, c.Y+dy*d)
			if ok && t.water[i] {
				in[inWater+r] = near
				break
			}
			if !ok || t.blocked[i] {
				in[inObstacle+r] = near
				break
			}
			food = math.Max(food, float64(t.food[i])*near)
			res = math.Max(res, s.resourceAt(c, i, hasTool)*near)
		}
		in[inFood+r] = food
		in[inResource+r] = res
	}

	partner, family, stranger := 0.0, 0.0, 0.0
	nearestD, nearestRep := socialRange, 0.0
	s.grid.near(c.X, c.Y, vision, func(o *Creature) {
		if o == c || o.Health <= 0 {
			return
		}
		dx, dy := o.X-c.X, o.Y-c.Y
		d := math.Hypot(dx, dy)
		if d > vision || d == 0 {
			return
		}
		if o.Sex != c.Sex && d < partnerRange {
			partner = math.Max(partner, 1-d/partnerRange)
		}
		if d < socialRange {
			if s.kin(c, o) {
				family = math.Max(family, 1-d/socialRange)
			} else {
				stranger = math.Max(stranger, 1-d/socialRange)
			}
			if d < nearestD {
				nearestD, nearestRep = d, o.Reputation
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

	here, _ := t.indexAt(c.X, c.Y)
	carried := c.Inventory.count()
	in[inEnergy] = c.Energy
	in[inHydration] = c.Hydration
	in[inHealth] = c.Health
	in[inAge] = s.age(c) / c.Genome.Traits.Lifespan
	in[inFertile] = b2f(s.fertile(c))
	in[inPregnant] = b2f(c.Pregnancy != nil)
	in[inFoodHere] = float64(t.food[here])
	in[inWaterNear] = b2f(s.canDrink(c))
	in[inSourceHere] = s.resourceAt(c, here, hasTool)
	in[inBumped] = b2f(c.Bumped)
	in[inPartnerNear] = partner
	in[inFamilyNear] = family
	in[inStrangerNear] = stranger
	in[inNearRep] = nearestRep
	in[inAttacked] = b2f(c.Hurt > 0)
	in[inCarryFood] = math.Min(1, float64(c.Inventory[chem.Food])/5)
	in[inCarryMaterial] = float64(carried-c.Inventory[chem.Food]) / invCapacity
	in[inCarryWeapon] = b2f(s.weaponBonus(c) > 0)
	in[inOwnHouseNear], in[inOtherHouseNear] = s.houseCloseness(c, vision)
	in[inCanCraft] = b2f(c.CanCraft)
	in[inCanBuild] = b2f(c.CanBuild)
	in[inClock] = math.Sin(2*math.Pi*s.time()/8 + float64(c.ID))
	in[inNoise] = s.rng.Float64()*2 - 1
	in[inBias] = 1
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
