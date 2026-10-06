package sim

import "math"

// The body's life course (Fase 3a): ageing, menopause, fertility, nursing and
// childbirth, after what is known of small-scale societies without modern
// medicine.
//
//   - Ageing: the risk of dying of old age doubles every eight years or so
//     (Gompertz; the senescent part of the Siler model fitted to
//     hunter-gatherers by Gurven & Kaplan 2007: a3 ≈ 1.5e-4, b3 ≈ 0.086 a
//     year), so the most common age of an adult death is in the early
//     seventies. The Lifespan gene shifts that clock; nobody passes 110.
//   - Women are fertile from adolescence to menopause (about 50); the chance
//     of conceiving in a month (fecundability) is about one in five at 20–30
//     and halves by 35 (Wood 1994). A thin woman conceives less often.
//   - Nursing keeps a mother from conceiving: fully for about a year, partly
//     until weaning at about two and a half (lactational amenorrhoea; this is what spaces
//     births about three years apart among foragers). If the baby dies her
//     fertility soon returns.
//   - About 1 in 100 births kills the mother (more when she is very young,
//     older, starved or carrying twins); twins are about 1.5% of births.
//   - About 5 in 100 babies die in their first weeks (prematurity, birth
//     asphyxia, infection of the cord): neonatal mortality without modern
//     care is some 30–60 per 1,000 births (Lawn et al. 2005; Hill & Hurtado
//     1996 for the Ache). More when the mother is starved, very young or
//     old, and for twins.
//   - Children feed themselves only slowly, getting better as they grow (Hadza
//     children start foraging at about five, Crittenden 2013); until then they
//     live on what others give them. The old slow down.
//   - Young children don't roam alone: among foragers a child under about ten
//     is always within reach of its mother or another carer (attachment,
//     Bowlby; Hewlett & Lamb 2005). A child that strays goes back; so it is
//     where its family finds food and water. An orphan with no grown
//     relative near has nobody to stay with.

const (
	senA      = 1.47e-4 // yearly risk of dying of old age at age 0, for a Lifespan gene of refLife …
	senB      = 0.086   // … rising by this factor's exponent each year (doubling every ~8 years)
	refLife   = 70.0
	maxAge    = 110.0
	senEvery  = TicksPerSecond // ticks between ageing checks (once a second, staggered)
	senescent = 15.0           // no one younger dies of old age

	menopauseInit = 48.0 // founders' age at menopause: 48 + up to 4 years
	menopauseMin  = 40.0
	menopauseMax  = 56.0

	fecundPeak   = 0.25                // chance of conceiving in a month at 20–30
	cycle        = SecondsPerYear / 12 // a month: one try at conceiving
	postpartum   = 0.15 * SecondsPerYear
	nurseFull    = 1.0 * SecondsPerYear // nursing stops a mother conceiving this long …
	nursePartial = 2.4 * SecondsPerYear // … and cuts her chances until weaning

	maternalRisk = 0.01  // chance a birth kills the mother
	twinRate     = 0.015 // twins per birth
	neonatalRisk = 0.05  // chance a newborn dies within weeks

	followAge   = 12.0 // years until which a child stays with whoever looks after it
	followRange = 2.5  // tiles it may stray before going back
	followPace  = 0.8  // share of its top speed it hurries back at
)

// guardian is who looks after child c: its mother, else its father, else
// the nearest grown relative in sight. Nil for children old enough to roam,
// for orphans left alone and in worlds without attachment.
func (s *Sim) guardian(c *Creature) *Creature {
	if s.opts.NoAttachment || s.ageYears(c) >= followAge {
		return nil
	}
	if c.Mother != nil {
		if m := s.living(c.Mother.ID); m != nil {
			return m
		}
	}
	if c.Father != nil {
		if f := s.living(c.Father.ID); f != nil {
			return f
		}
	}
	var best *Creature
	bestD := c.Genome.Traits.Vision
	s.grid.near(c.X, c.Y, bestD, func(o *Creature) {
		if o == c || o.Health <= 0 || !s.adult(o) || !s.kin(c, o) {
			return
		}
		if d := math.Hypot(o.X-c.X, o.Y-c.Y); d < bestD {
			best, bestD = o, d
		}
	})
	return best
}

// stayClose turns a child that has strayed from its carer back towards
// them, at least at a walk. It returns the child's speed.
func (s *Sim) stayClose(c *Creature, speed float64) float64 {
	g := s.guardian(c)
	if g == nil {
		return speed
	}
	dx, dy := g.X-c.X, g.Y-c.Y
	if math.Hypot(dx, dy) <= followRange {
		return speed
	}
	turn := normAngle(math.Atan2(dy, dx) - c.Heading)
	step := maxTurnRate * 2 * dt
	c.Heading = normAngle(c.Heading + clamp(turn, -step, step))
	c.resting = false
	return math.Max(speed, followPace*c.Genome.Traits.MaxSpeed*s.vigor(c)*(1-sicknessSlows*sickness(c)))
}

// fecundability is a woman's chance of conceiving in a month of trying, by
// age (years), before her nutrition is counted.
func fecundability(age, menopause float64) float64 {
	switch {
	case age < 15 || age >= menopause:
		return 0
	case age < 18:
		return fecundPeak * (0.5 + 0.5*(age-15)/3) // adolescent subfecundity
	case age < 30:
		return fecundPeak
	case age < 40:
		return fecundPeak * (1 - 0.55*(age-30)/10)
	default:
		end := math.Max(menopause, 41)
		return fecundPeak * 0.45 * math.Max(0, 1-(age-40)/(end-40))
	}
}

// nursingBlock tells how much a mother's nursing keeps her from conceiving:
// 0 not at all, 0.5 partly, 1 fully.
func (s *Sim) nursingBlock(f *Creature) float64 {
	if f.Sex != Female || f.NursingID == 0 {
		return 0
	}
	baby := s.byID[f.NursingID]
	if baby == nil || baby.Health <= 0 {
		return 0
	}
	switch age := s.age(baby); {
	case age < nurseFull:
		return 1
	case age < nursePartial:
		return 0.4
	}
	return 0
}

// nursing reports whether f is breastfeeding (her youngest is not yet weaned).
func (s *Sim) nursing(f *Creature) bool { return s.nursingBlock(f) > 0 }

// conceptionChance is a month's chance that f conceives now.
func (s *Sim) conceptionChance(f *Creature) float64 {
	p := fecundability(s.ageYears(f), f.Genome.Traits.Menopause)
	p *= 1 - s.nursingBlock(f)
	// Thin women conceive less (energy here stands in for body fat), and a
	// fever all but stops it.
	return p * clamp((f.Energy-0.25)/0.35, 0, 1) * (1 - sickness(f)) * s.nutritionFertility(f) * s.fertilityGene(f)
}

// childbirthRisk is the chance that delivering n babies kills the mother.
func (s *Sim) childbirthRisk(f *Creature, n int) float64 {
	r := maternalRisk
	switch age := s.ageYears(f); {
	case age < 18:
		r *= 1.7
	case age >= 40:
		r *= 1.5
	}
	if f.Energy < 0.35 || f.Health < 0.5 {
		r *= 2
	}
	if n > 1 {
		r *= 2
	}
	return r * s.nutritionBirthRisk(f)
}

// neonatalDeathRisk is the chance that a baby of mother f, one of n born
// together, dies in its first weeks.
func (s *Sim) neonatalDeathRisk(f *Creature, n int) float64 {
	r := neonatalRisk
	switch age := s.ageYears(f); {
	case age < 18:
		r *= 1.5
	case age >= 40:
		r *= 1.4
	}
	if f.Energy < 0.35 || f.Health < 0.5 {
		r *= 2
	}
	if n > 1 {
		r *= 3
	}
	return r * s.nutritionNeonatalRisk(f)
}

// senescenceHazard is the yearly risk of dying of old age.
func senescenceHazard(age, lifespan float64) float64 {
	return senA * math.Exp(senB*(age+refLife-lifespan))
}

// senesce gives everyone, once a second, the chance to die of old age.
func (s *Sim) senesce() {
	for _, c := range s.creatures {
		if (s.tick+c.ID)%senEvery != 0 || c.Health <= 0 {
			continue
		}
		age := s.ageYears(c)
		if age < senescent {
			continue
		}
		h := senescenceHazard(age, c.Genome.Traits.Lifespan/SecondsPerYear)
		p := 1 - math.Exp(-h*float64(senEvery)*dt/SecondsPerYear)
		if age >= maxAge || s.rng.Float64() < p {
			c.fate = "oldAge"
		}
	}
}

// vigor is how much of their strength someone has left with age (1 until
// 50), and with a recessive disorder (genetics.go).
func (s *Sim) vigor(c *Creature) float64 {
	return (1 - clamp((s.ageYears(c)-50)*0.012, 0, 0.5)) * s.vigorGene(c)
}

// childSkill is how well a child feeds itself: about a third as well as an
// adult at two, nearly as well by twelve.
func (s *Sim) childSkill(c *Creature) float64 {
	age := s.ageYears(c)
	if age >= 12 {
		return 1
	}
	t := clamp((age-2)/10, 0, 1)
	return 0.35 + 0.65*t*t*(3-2*t)
}
