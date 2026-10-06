package sim

import (
	"math"
)

const (
	// Seconds before the next give, theft or blow. A theft needs a new
	// opportunity (the thief lies low); handing things over takes a moment.
	blowCooldown  = 1.0
	stealCooldown = 12.0
	giveCooldown  = 4.0
	huntCooldown  = 1.0

	// Babies ride with their mother and are breastfed until about two
	// (weaning in foraging societies is around two to three years).
	carryAge    = 2 * SecondsPerYear
	nurseRate   = 0.05 // energy and water a baby can take per second
	nurseCost   = 1.2  // a mother spends a little more than the baby gets
	nurseFloor  = 0.25 // a starving mother's milk dries up
	childBody   = 0.35 // a newborn's needs relative to an adult
	femaleBody  = 0.78 // a grown woman's needs relative to a man's
	pubertyAge  = 10 * SecondsPerYear
	idleRetry   = 1.0 // seconds before looking for work again after finding none
	fxSeconds   = 1.0
	hurtSeconds = 3.0
)

// act lets the brain decide and applies the consequences.
func (s *Sim) act(c *Creature) {
	out := &c.output
	think(c.Genome, c.Mind.wIn, c.Mind.wOut, &c.input, c.Hidden, out)
	if (s.tick+c.ID)%learnEvery == 0 {
		s.learnStep(c)
	}
	tr := &c.Genome.Traits
	c.ate, c.drank = false, false
	if m := s.carrier(c); m != nil {
		s.beCarried(c, m)
		return
	}
	c.resting = out[outRest] > 0.5
	c.action = ActExplore
	c.wantsMate = out[outMate] > 0.5
	if c.wantsMate && s.fertile(c) {
		c.action = ActMate
	}

	moveFactor := 1.0
	if c.resting {
		c.action = ActRest
		c.Job = nil
	} else {
		// A sated creature doesn't eat or drink, however much it wants to.
		if out[outEat] > 0.5 && c.Energy < 0.98 {
			if reach := s.foodAround(int(math.Floor(c.X)), int(math.Floor(c.Y))); reach > 0.01 {
				// Fruit, tubers and greens within arm's reach.
				want := float32(min(float64(reach), eatRate*dt, (1-c.Energy)/foodValue))
				s.takeFoodAround(int(math.Floor(c.X)), int(math.Floor(c.Y)), want)
				c.Energy += float64(want) * foodValue
				s.eco.Current().FoodWild += float64(want) * foodValue
				c.ate = true
			} else if c.Energy < 0.9 && (s.tick+c.ID)%(TicksPerSecond/2) == 0 {
				// Carried food is eaten a unit at a time, what spoils first first.
				if id := s.cat.edible(c.Inventory, s.handKeep(c)); id != "" {
					c.Inventory.take(id, 1)
					f := s.cat.item(id).Food
					c.Energy = math.Min(1, c.Energy+f)
					s.ateFood(id, f)
					c.ate = true
				}
			}
		}
		if out[outDrink] > 0.5 && c.Hydration < 0.98 && s.canDrink(c) {
			c.Hydration = math.Min(1, c.Hydration+drinkRate*dt)
			c.drank = true
			c.WaterX, c.WaterY = c.X, c.Y
		}
		if c.ate || c.drank {
			moveFactor = slowWhileFeeding
			c.action = ActEat
			if !c.ate {
				c.action = ActDrink
			}
		}
		moveFactor = math.Min(moveFactor, s.work(c))
	}

	c.Heading = normAngle(c.Heading + out[outTurn]*maxTurnRate*dt)
	speed := 0.0
	if !c.resting {
		speed = out[outMove] * tr.MaxSpeed * moveFactor
	}
	c.Pace = speed / tr.MaxSpeed
	c.Bumped = false
	if speed > 0 {
		c.Bumped = !s.move(c, speed*dt)
	}

	s.metabolize(c, speed)
	if (s.tick+c.ID)%TicksPerSecond == 0 && s.atHome(c) {
		s.keepHouse(c)
	}
}

// metabolize spends a tick's energy and water (children, being smaller,
// need less) and lets time heal and fade.
func (s *Sim) metabolize(c *Creature, speed float64) {
	tr := &c.Genome.Traits
	cost := basalCost * tr.Size * tr.Metabolism
	if c.resting {
		cost *= 0.5
	}
	cost += moveCost*speed*speed*tr.Size + visionCost*tr.Vision + brainCost*float64(c.Genome.Hidden)
	if c.Pregnancy != nil {
		cost += pregnancyCost
	}
	body := s.bodyScale(c)
	c.Energy -= cost * hungerScale * body * dt
	c.Hydration -= thirstRate * thirstScale * body * dt
	if c.Cooldown > 0 {
		c.Cooldown -= dt
	}
	if c.Hurt > 0 {
		c.Hurt -= dt
	}
	c.Reputation *= 1 - dt/reputationAge

	if c.Energy > 0.3 && c.Hydration > 0.3 && c.Health < 1 {
		heal := healRate
		if c.resting {
			heal *= 3
		}
		if s.atHome(c) {
			heal *= 2
		}
		c.Health = math.Min(1, c.Health+heal*dt)
	}
}

// bodyScale is how much food and water a body needs relative to a grown
// man: a newborn about a third, growing to the full amount at 15. Women,
// being smaller, need less. Among the Hadza they spend 71 % of the energy
// men do, but partly because the men walk twice as far a day, while women
// weigh 84 % of what men do (Pontzer et al. 2012); here both sexes move
// alike, so the value lies between. Girls and boys need much the same
// until the sexes grow apart at puberty, 10 to 15. With a man's budget
// women starved twice as often as men; at 71 % men starved more, at 84 %
// women did again.
func (s *Sim) bodyScale(c *Creature) float64 {
	age := s.age(c)
	scale := childBody + (1-childBody)*math.Min(1, age/adultAge)
	if c.Sex == Female {
		puberty := clamp((age-pubertyAge)/(adultAge-pubertyAge), 0, 1)
		scale *= 1 - (1-femaleBody)*puberty
	}
	return scale
}

// carrier is the mother carrying and nursing c, if c is still a baby and
// she is alive (anticipating Fase 3's dependent childhood).
func (s *Sim) carrier(c *Creature) *Creature {
	if c.Mother == nil || s.age(c) >= carryAge {
		return nil
	}
	return s.living(c.Mother.ID)
}

// beCarried: a baby rides with its mother and is nursed by her, as long as
// she has enough to give. Its own brain still runs (and learns) but does
// not move it.
func (s *Sim) beCarried(c, m *Creature) {
	c.X, c.Y = m.X+0.15*math.Cos(float64(c.ID)), m.Y+0.15*math.Sin(float64(c.ID))
	if s.terrain.blockedAt(c.X, c.Y) {
		// Never over water or rock, where a restored world would move it.
		c.X, c.Y = m.X, m.Y
	}
	c.Heading = m.Heading
	c.resting, c.wantsMate = true, false
	c.Pace = 0
	c.action = ActRest
	c.Bumped = false
	s.metabolize(c, 0)
	if m.Energy > nurseFloor {
		give := math.Min(math.Max(0, 0.95-c.Energy), nurseRate*dt)
		c.Energy += give
		m.Energy -= give * nurseCost
	}
	if m.Hydration > nurseFloor {
		give := math.Min(math.Max(0, 0.95-c.Hydration), nurseRate*dt)
		c.Hydration += give
		m.Hydration -= give
	}
}

// work runs at most one deliberate activity this tick, in the order
// attack > steal > hunt > give > teach > build > craft > plant > gather, each
// only if the brain wants it. It returns how much the creature may still
// move (1 = freely).
func (s *Sim) work(c *Creature) float64 {
	out := &c.output
	if c.ActCD > 0 {
		c.ActCD -= dt
	}
	if c.IdleWork > 0 {
		c.IdleWork -= dt
	}
	if j := c.Job; j != nil {
		if out[j.output()] <= 0.5 {
			c.Job = nil // changed its mind; nothing is consumed until the work is done
		} else {
			s.flash(c, j.fx())
			c.action = j.action()
			if j.Remaining -= dt; j.Remaining <= 0 {
				c.Job = nil
				s.finishJob(c, j)
			}
			return 0
		}
	}
	if c.ActCD <= 0 {
		// Children may share, but only grown-ups fight and steal (unless crime is switched off).
		grown := s.adult(c) && !s.opts.NoCrime
		switch {
		case grown && out[outAttack] > 0.5 && s.attack(c):
			c.ActCD = blowCooldown
			return 1
		case grown && out[outSteal] > 0.5 && s.steal(c):
			c.ActCD = stealCooldown
			return 1
		case s.adult(c) && out[outHunt] > 0.5 && s.hunt(c):
			c.ActCD = huntCooldown
			return 1
		case out[outGive] > 0.5 && s.give(c):
			c.ActCD = giveCooldown
			return 1
		}
	}
	if out[outTeach] > 0.5 && s.adult(c) && s.teach(c) {
		return 0 // teaching (or writing) holds still
	}
	if c.IdleWork <= 0 && s.adult(c) {
		if out[outBuild] > 0.5 && s.startBuild(c) || out[outCraft] > 0.5 && s.startCraft(c) {
			return 0
		}
		if out[outBuild] > 0.5 || out[outCraft] > 0.5 {
			c.IdleWork = idleRetry
		}
	}
	if out[outPlant] > 0.5 && s.plant(c) {
		return slowWhileFeeding
	}
	if out[outGather] > 0.5 && s.gather(c) {
		return slowWhileFeeding
	}
	return 1
}

func (s *Sim) flash(c *Creature, fx int) {
	c.fx[fx] = s.tick + int64(fxSeconds*TicksPerSecond)
}

// canDrink reports whether fresh water or a well is within reach. Sea water
// can't be drunk.
func (s *Sim) canDrink(c *Creature) bool {
	if i, ok := s.terrain.indexAt(c.X, c.Y); ok && s.terrain.nearFresh[i] {
		return true
	}
	return s.structureNear(c.X, c.Y, wellReach, func(st *Structure) bool { return st.kind.Well })
}

// move steps the creature forward, sliding along walls. It reports whether
// the full move succeeded.
func (s *Sim) move(c *Creature, dist float64) bool {
	dx, dy := math.Cos(c.Heading)*dist, math.Sin(c.Heading)*dist
	r := bodyRadius * c.Genome.Traits.Size
	if s.free(c.X+dx, c.Y+dy, dx, dy, r) {
		c.X += dx
		c.Y += dy
		return true
	}
	if s.free(c.X+dx, c.Y, dx, 0, r) {
		c.X += dx
	} else if s.free(c.X, c.Y+dy, 0, dy, r) {
		c.Y += dy
	}
	return false
}

// free checks the destination centre and the body's leading edge.
func (s *Sim) free(x, y, dx, dy, r float64) bool {
	l := math.Hypot(dx, dy)
	if l == 0 {
		return !s.terrain.blockedAt(x, y)
	}
	return !s.terrain.blockedAt(x, y) && !s.terrain.blockedAt(x+dx/l*r, y+dy/l*r)
}
