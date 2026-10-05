package sim

import (
	"math"

	"miniv2/backend/internal/chem"
)

const (
	// Seconds before the next give, theft or blow. A theft needs a new
	// opportunity (the thief lies low); handing things over takes a moment.
	blowCooldown  = 1.0
	stealCooldown = 12.0
	giveCooldown  = 4.0
	idleRetry     = 1.0 // seconds before looking for work again after finding none
	fxSeconds     = 1.0
	hurtSeconds   = 3.0
)

// act lets the brain decide and applies the consequences.
func (s *Sim) act(c *Creature) {
	out := &c.output
	think(c.Genome, c.Mind.wIn, c.Mind.wOut, &c.input, c.Hidden, out)
	if (s.tick+c.ID)%learnEvery == 0 {
		s.learnStep(c)
	}
	tr := &c.Genome.Traits
	t := s.terrain
	here, _ := t.indexAt(c.X, c.Y)

	c.resting = out[outRest] > 0.5
	c.ate, c.drank = false, false
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
			if t.food[here] > 0.01 {
				take := min(float64(t.food[here]), eatRate*dt, (1-c.Energy)/foodValue)
				t.food[here] -= float32(take)
				c.Energy += take * foodValue
				c.ate = true
			} else if c.Energy < 0.9 && c.Inventory[chem.Food] > 0 && (s.tick+c.ID)%TicksPerSecond == 0 {
				// Carried food is eaten a unit at a time.
				c.Inventory.take(chem.Food, 1)
				c.Energy = math.Min(1, c.Energy+s.cat.foodEnergy())
				c.ate = true
			}
		}
		if out[outDrink] > 0.5 && c.Hydration < 0.98 && s.canDrink(c) {
			c.Hydration = math.Min(1, c.Hydration+drinkRate*dt)
			c.drank = true
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
	c.Bumped = false
	if speed > 0 {
		c.Bumped = !s.move(c, speed*dt)
	}

	cost := basalCost * tr.Size * tr.Metabolism
	if c.resting {
		cost *= 0.5
	}
	cost += moveCost*speed*speed*tr.Size + visionCost*tr.Vision + brainCost*float64(c.Genome.Hidden)
	if c.Pregnancy != nil {
		cost += pregnancyCost
	}
	c.Energy -= cost * dt
	c.Hydration -= thirstRate * dt
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
	if (s.tick+c.ID)%TicksPerSecond == 0 && s.atHome(c) {
		s.keepHouse(c)
	}
}

// work runs at most one deliberate activity this tick, in the order
// attack > steal > give > teach > build > craft > gather, each only if the
// brain wants it. It returns how much the creature may still move (1 = freely).
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
	if out[outGather] > 0.5 && s.gather(c) {
		return slowWhileFeeding
	}
	return 1
}

func (s *Sim) flash(c *Creature, fx int) {
	c.fx[fx] = s.tick + int64(fxSeconds*TicksPerSecond)
}

// canDrink reports whether water or a well is within reach.
func (s *Sim) canDrink(c *Creature) bool {
	if i, ok := s.terrain.indexAt(c.X, c.Y); ok && s.terrain.nearWater[i] {
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
