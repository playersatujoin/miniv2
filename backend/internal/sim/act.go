package sim

import (
	"math"

	"miniv2/backend/internal/ecology"
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

// Movement eases (see act): the turn rate follows the brain's wish with a
// time constant of a quarter second, and the pace rises to full speed or
// falls to a stop in about paceRise seconds, as a walker's does (Winter
// 1991: two or three steps to reach a steady gait).
var turnEase = 1 - math.Exp(-dt/0.25)

const paceRise = 0.4

// act lets the brain decide and applies the consequences.
func (s *Sim) act(c *Creature) {
	out := &c.output
	s.decide(c)
	if (s.tick+c.ID)%learnEvery == 0 {
		s.learnStep(c)
	}
	s.affectStep(c)
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
				// A small child picks and eats slowly (see childSkill).
				want := float32(min(float64(reach), eatRate*dt*s.childSkill(c), (1-c.Energy)/foodValue))
				s.takeFoodAround(int(math.Floor(c.X)), int(math.Floor(c.Y)), want)
				s.eatWild(c, int(math.Floor(c.X)), int(math.Floor(c.Y)), float64(want)*foodValue)
				c.Energy += float64(want) * foodValue
				s.eco.Current().FoodWild += float64(want) * foodValue
				c.ate = true
				s.rememberFood(c, c.X, c.Y, float64(reach)/2)
			} else if c.Energy < 0.9 && (s.tick+c.ID)%(TicksPerSecond/2) == 0 && !s.eatCarcass(c) {
				// Carried food is eaten a unit at a time, what spoils first
				// first (after meat lying in reach); with none, what family
				// or a well-off neighbour close by can spare (sharing.go).
				if id := s.cat.edible(c.Inventory, s.handKeep(c)); id != "" {
					c.Inventory.take(id, 1)
					f := s.cat.item(id).Food
					s.eatFood(c, id, math.Min(f, 1-c.Energy))
					c.Energy = math.Min(1, c.Energy+f)
					s.ateFood(id, f)
					c.ate = true
				} else {
					s.scrounge(c)
				}
			}
		}
		if out[outDrink] > 0.5 && (c.Hydration < 0.98 || c.WaterCarried < s.waterRoom(c)) {
			if kind, tile := s.waterSource(c); kind == drinkNone {
				// Away from water: the tubes, if they hold any.
				c.drank = c.Hydration < 0.98 && s.drinkCarried(c)
			} else {
				s.fillWater(c, kind, tile)
				if c.Hydration < 0.98 {
					rate := drinkRate
					if kind == drinkDug {
						rate *= digRate
					}
					gain := math.Min(1-c.Hydration, rate*dt)
					c.Hydration += gain
					// The water comes out of the river, the pool, the sand or the aquifer.
					if kind == drinkWell {
						s.eco.DrawGroundwater(tile, gain*waterPerHydration)
					} else {
						s.eco.Drink(tile, gain*waterPerHydration)
					}
					if !s.noDisease() {
						c.Swallowed += s.germsDrunk(kind, tile) * gain
					}
					c.drank = true
					c.WaterX, c.WaterY = c.X, c.Y
				}
			}
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

	// The network supplies intent; its executor only handles the route. A
	// body doesn't swing round or stop dead the moment the mind changes: the
	// turn and the pace ease towards what is wanted (after RAGE's
	// Peds/PedMoveBlend), so a brain that wavers between left and right at
	// each thought walks a smooth, if wandering, line.
	turn := c.Heading
	moveFactor *= s.navigate(c)
	if c.Travel == nil || c.Travel.Phase != "approach" {
		c.Loco.Turn += (out[outTurn]*maxTurnRate - c.Loco.Turn) * turnEase
		c.Heading = normAngle(turn + c.Loco.Turn*dt)
	} else {
		c.Loco.Turn = 0
	}
	speed := 0.0
	if !c.resting {
		speed = out[outMove] * tr.MaxSpeed * moveFactor * s.vigor(c) * (1 - sicknessSlows*sickness(c)) * s.locomotionFactor(c)
	}
	speed = s.stayClose(c, speed)
	step := tr.MaxSpeed / paceRise * dt
	c.Loco.Speed += clamp(speed-c.Loco.Speed, -step, step)
	speed = math.Min(speed, c.Loco.Speed)
	c.Pace = speed / tr.MaxSpeed
	c.Bumped = false
	if speed > 0 {
		c.Bumped = !s.move(c, speed*dt)
	}
	s.locomotionStep(c)

	s.metabolize(c, speed)
	if (s.tick+c.ID)%TicksPerSecond == 0 && s.atHome(c) {
		s.keepHouse(c)
	}
}

// metabolize spends a tick's energy and water (children, being smaller,
// need less) and lets time heal and fade.
func (s *Sim) metabolize(c *Creature, speed float64) {
	tr := &c.Genome.Traits
	// Stronger immune defences and a gut full of worms both cost food.
	cost := basalCost * tr.Size * tr.Metabolism * (1 + immuneCost*(tr.Immunity-1) + wormHunger*c.Worms)
	if c.resting {
		cost *= 0.5
	}
	cost += moveCost*speed*speed*tr.Size + visionCost*tr.Vision + brainCost*float64(brainSize(c))
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

	// Wounds don't mend while an illness is taking its toll.
	if c.Energy > 0.3 && c.Hydration > 0.3 && c.Health < 1 && c.Ill == nil {
		heal := healRate
		if c.resting {
			heal *= 3
		}
		if s.atHome(c) {
			heal *= 2
		}
		heal *= s.nutritionHealing(c)
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
		s.suckle(c, m, give)
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

// Where a drink comes from.
const (
	drinkNone = iota
	drinkOpen // a river, a pool or a lake
	drinkDug  // a hole dug in a dry riverbed, seeping slowly
	drinkWell // a well that still reaches the groundwater
)

// digRate is how much slower water seeps into a hole dug in a dry riverbed
// than it can be scooped from a river.
const digRate = 0.4

// drinkPerYear is what one person drinks in a year in the water cycle's
// units (about 4 litres a day, 1.5 m³ a year; see ecology/hydrology.go).
const drinkPerYear = 0.094

// waterPerHydration turns the hydration a drink restores into water taken
// from the river: a year's worth of thirst is a year's drinking.
const waterPerHydration = drinkPerYear / (thirstRate * thirstScale * SecondsPerYear)

// waterSource finds the water within reach: open fresh water first, then
// water in the sand of a dry riverbed, then a well whose groundwater hasn't
// sunk out of reach. Sea water can't be drunk. tile is where it is taken from.
func (s *Sim) waterSource(c *Creature) (kind, tile int) {
	x, y := int(math.Floor(c.X)), int(math.Floor(c.Y))
	dug := -1
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			j, ok := s.terrain.index(x+dx, y+dy)
			if !ok || !s.terrain.fresh[j] {
				continue
			}
			switch s.eco.WaterState(j) {
			case ecology.WaterFlowing, ecology.WaterPools:
				return drinkOpen, j
			case ecology.WaterUnder:
				dug = j
			}
		}
	}
	if dug >= 0 {
		return drinkDug, dug
	}
	for _, st := range s.structures {
		if !st.kind.Well || st.dist(c.X, c.Y) > wellReach {
			continue
		}
		if i, ok := s.terrain.index(st.X, st.Y); ok && s.eco.Groundwater(i) >= ecology.WellMin {
			return drinkWell, i
		}
	}
	return drinkNone, -1
}

// riverRunsBy reports whether open fresh water lies next to tile i now.
func (s *Sim) riverRunsBy(i int) bool {
	x, y := i%s.terrain.w, i/s.terrain.w
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if j, ok := s.terrain.index(x+dx, y+dy); ok && s.terrain.fresh[j] && s.eco.WaterState(j) >= ecology.WaterPools {
				return true
			}
		}
	}
	return false
}

// waterAt reports whether someone standing at (x, y) would find water to drink.
func (s *Sim) waterAt(x, y float64) bool {
	kind, _ := s.waterSource(&Creature{X: x, Y: y})
	return kind != drinkNone
}

// canDrink reports whether there is water to drink within reach.
func (s *Sim) canDrink(c *Creature) bool {
	kind, _ := s.waterSource(c)
	return kind != drinkNone
}
