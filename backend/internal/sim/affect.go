package sim

import (
	"math"
)

// Affect: moods that rise with what happens to a person and ebb back to their
// own resting level, after RAGE's ped motivation (Peds/PedIntelligence/
// PedMotivation.h). Moods are senses (inFear, inAnger, inJoy, inGrief); they
// never block or choose an action. The observer sees them on faces.
//
// As in CPedMotivation each mood is 0–1 with a resting level it drifts back
// to, things that happen raise it by an amount times a personal rate, and a
// mood shows from 0.3 up (its fear threshold). Here the personal part is
// inherited: three temperament genes (Traits) set how strongly things move a
// person (Reactivity), how fast their moods ebb (Recovery) and how cheerful
// they are at rest (Cheer).
//
// What moves the moods:
//   - fear follows what is frightening now: the stimuli perceived (as
//     frightening as their kind), being struck, a predator in sight, a body
//     close to death; and jumps when assaulted;
//   - anger jumps when robbed or assaulted, and simmers with hunger (the
//     frustration–aggression link: low blood sugar goes with irritability
//     and aggression, Bushman et al. 2014; hunger with anger and
//     irritability in daily life, Swami et al. 2022). Bumping into things is
//     left out: here it is mostly a body steering round the crowd;
//   - joy rises with a meal or a drink when hungry or thirsty (the hungrier,
//     the more), a gift, a baby conceived or born, and whatever other
//     systems report through feel (a talk, a good trade);
//   - grief comes with the death of family (mourn), most for a child, then a
//     spouse, a parent, a brother or sister.
//
// How long moods last, on the scale of actions (a blow a second): fear is
// the shortest-lived, then anger, then joy, following the order of emotion
// durations measured by Verduyn & Lavrijsen (2015); grief, the longest-lived
// there, is on the scale of a life: it eases over months and is mostly gone
// within a year or two, as acute grief is for most bereaved people (Bonanno
// et al. 2002; prolonged grief is diagnosed only after a year, DSM-5-TR).

// Mood is one of the moods a person feels.
type Mood int

const (
	moodFear Mood = iota
	moodAnger
	moodJoy
	moodGrief
	numMoods
)

var moodNames = [numMoods]string{"fear", "anger", "joy", "grief"}

// Affect is a person's moods, each 0–1.
type Affect struct {
	Fear  float64 `json:"fear,omitempty"`
	Anger float64 `json:"anger,omitempty"`
	Joy   float64 `json:"joy,omitempty"`
	Grief float64 `json:"grief,omitempty"`
	// At is the tick the moods were last brought up to date (0 never: they
	// start calm), and Energy and Water how fed and watered the body was
	// then, so that a meal since can lift the mood.
	At     int64   `json:"at,omitempty"`
	Energy float64 `json:"energy,omitempty"`
	Water  float64 `json:"water,omitempty"`

	// How frightening what the senses perceived at tick sensedAt was
	// (stimuli.go): the moods use it when brought up to date in the same
	// tick, so the stimuli are looked at once. Only read in that tick, so it
	// need not be saved.
	sensedAt   int64
	sensedFear float64
}

// MoodView is a person's moods for the inspector.
type MoodView struct {
	Fear     float64 `json:"fear"`
	Anger    float64 `json:"anger"`
	Joy      float64 `json:"joy"`
	Grief    float64 `json:"grief"`
	Dominant string  `json:"dominant"` // "" when calm
	// The temperament genes (about 1 each).
	Reactivity float64 `json:"reactivity"`
	Recovery   float64 `json:"recovery"`
	Cheer      float64 `json:"cheer"`
}

const (
	// affectEvery: moods are brought up to date five times a second, in step
	// with thinking (cognition.go), so the senses are fresh when they are.
	affectEvery = 4
	// affectGap caps the time one update may cover (seconds).
	affectGap = 1.0
	// moodShown is the level from which a mood shows on the face
	// (CPedMotivation's fear threshold).
	moodShown = 0.3

	// Time constants (seconds to cover about two thirds of the way to what
	// a mood follows): rising is quick; ebbing back takes each mood its own
	// time (above), divided by the Recovery gene.
	riseTime  = 0.25
	fearTime  = 2.0
	angerTime = 3.0
	joyTime   = 4.0
	griefTime = SecondsPerYear / 2

	// joyRest is the resting joy of an average temperament (times Cheer);
	// calm is otherwise free of fear, anger and grief.
	joyRest = 0.15

	// Fear while hurt (struck or robbed: the "diserang" sense), with a
	// predator in plain sight close by, and with a body this close to death
	// (health below frailHealth).
	hurtFear     = 0.5
	predatorFear = 0.7
	frailHealth  = 0.3
	frailFear    = 0.7
	// Anger simmering with hunger, from hungerEnergy down to nothing to eat.
	hungerAnger  = 0.35
	hungerEnergy = 0.3

	// Jumps. A direct wrong angers by its weight in perceiveEvent (0.5 an
	// assault, 0.3 a theft), an assault also frightens; a gift gladdens
	// giftJoy times its weight (0.29 for food handed over).
	assaultFear = 0.3
	giftJoy     = 1.6
	// A meal or a drink: joy per unit of energy or water gained, times how
	// much it was needed.
	mealJoy     = 1.0
	birthJoy    = 0.6
	conceiveJoy = 0.3

	// Grief at a death in the family: losing a child is the hardest
	// bereavement (Rogers et al. 2008), then a spouse, a parent, a brother
	// or sister (a half-sibling less).
	griefChild       = 0.9
	griefSpouse      = 0.8
	griefParent      = 0.6
	griefSibling     = 0.45
	griefHalfSibling = 0.3
)

// Ranges of the temperament genes.
const (
	reactivityMin, reactivityMax = 0.3, 2.0
	recoveryMin, recoveryMax     = 0.3, 3.0
	cheerMin, cheerMax           = 0.2, 2.5
)

// gene reads a temperament gene: 0 (a genome built by hand) means average.
func gene(v float64) float64 {
	if v > 0 {
		return v
	}
	return 1
}

func (a *Affect) mood(m Mood) *float64 {
	switch m {
	case moodFear:
		return &a.Fear
	case moodAnger:
		return &a.Anger
	case moodJoy:
		return &a.Joy
	case moodGrief:
		return &a.Grief
	}
	return nil
}

// feel raises (or, negative, eases) one of c's moods.
func (s *Sim) feel(c *Creature, m Mood, amount float64) {
	if s.opts.NoAffect || c == nil || c.Genome == nil || math.IsNaN(amount) {
		return
	}
	if v := c.Affect.mood(m); v != nil {
		*v = clamp(*v+amount*gene(c.Genome.Traits.Reactivity), 0, 1)
	}
}

// affectStep lets moods follow the body and ebb; called from act every tick.
func (s *Sim) affectStep(c *Creature) {
	a := &c.Affect
	if s.opts.NoAffect {
		if *a != (Affect{}) {
			*a = Affect{}
		}
		return
	}
	if (s.tick+c.ID)%affectEvery != 0 {
		return
	}
	if a.At == 0 || a.At >= s.tick {
		// The first look at a newborn: its parents rejoice.
		if a.At == 0 && s.tick-c.BornTick <= affectEvery {
			for _, p := range []*Ref{c.Mother, c.Father} {
				if p != nil {
					s.feel(s.living(p.ID), moodJoy, birthJoy)
				}
			}
		}
		a.At, a.Energy, a.Water = s.tick, c.Energy, c.Hydration
		return
	}
	el := math.Min(float64(s.tick-a.At)*dt, affectGap)
	tr := &c.Genome.Traits
	react, recovery := gene(tr.Reactivity), gene(tr.Recovery)

	// A child conceived since the last look.
	if p := c.Pregnancy; p != nil && gestation-p.Remaining < (float64(s.tick-a.At)+0.5)*dt {
		s.feel(c, moodJoy, conceiveJoy)
		s.feel(s.living(p.Father.ID), moodJoy, conceiveJoy)
	}
	// A meal or a drink since, the sweeter the hungrier or thirstier.
	meal := math.Max(0, c.Energy-a.Energy)*clamp(1-a.Energy, 0, 1) +
		math.Max(0, c.Hydration-a.Water)*clamp(1-a.Water, 0, 1)
	if meal > 0 {
		s.feel(c, moodJoy, mealJoy*meal)
	}

	fear := a.sensedFear
	if a.sensedAt != s.tick {
		_, _, fear = s.perceivedShock(c)
	}
	if c.Hurt > 0 {
		fear = math.Max(fear, hurtFear)
	}
	fear = math.Max(fear, c.input[inPredatorNear]*predatorFear)
	if c.Health < frailHealth {
		fear = math.Max(fear, (frailHealth-math.Max(0, c.Health))/frailHealth*frailFear)
	}
	anger := hungerAnger * clamp((hungerEnergy-c.Energy)/hungerEnergy, 0, 1)

	follow := func(v *float64, rest, drive, ebb float64) {
		target := math.Max(rest, math.Min(1, drive*react))
		// A leaky integrator: k/(1+k) of the gap closes per update.
		k := el / riseTime
		if *v > target {
			k = el * recovery / ebb
		}
		*v = clamp(*v+(target-*v)*k/(1+k), 0, 1)
	}
	follow(&a.Fear, 0, fear, fearTime)
	follow(&a.Anger, 0, anger, angerTime)
	follow(&a.Joy, clamp(joyRest*gene(tr.Cheer), 0, 1), 0, joyTime)
	follow(&a.Grief, 0, 0, griefTime)
	a.At, a.Energy, a.Water = s.tick, c.Energy, c.Hydration
}

// senseAffect fills inFear, inAnger, inJoy and inGrief.
func (s *Sim) senseAffect(c *Creature, in *[NumInputs]float64) {
	if s.opts.NoAffect {
		return
	}
	a := &c.Affect
	in[inFear], in[inAnger], in[inJoy], in[inGrief] = a.Fear, a.Anger, a.Joy, a.Grief
}

// mourn lets those who loved the dead grieve (called from die).
func (s *Sim) mourn(dead *Creature) {
	if s.opts.NoAffect {
		return
	}
	grieve := func(o *Creature, amount float64) {
		if o != nil && o != dead && o.Health > 0 {
			s.feel(o, moodGrief, amount)
		}
	}
	if dead.Spouse != nil {
		if sp := s.living(dead.Spouse.ID); sp != nil && sp.Spouse != nil && sp.Spouse.ID == dead.ID {
			grieve(sp, griefSpouse)
		}
	}
	for _, p := range []*Ref{dead.Mother, dead.Father} {
		if p != nil {
			grieve(s.living(p.ID), griefChild)
		}
	}
	// Children and brothers and sisters: deaths are rare enough to look
	// through everyone (a few a second among a thousand people).
	for _, o := range s.creatures {
		if o == dead || o.Health <= 0 {
			continue
		}
		if isParent(dead, o) {
			grieve(o, griefParent)
			continue
		}
		mother := dead.Mother != nil && o.Mother != nil && dead.Mother.ID == o.Mother.ID
		father := dead.Father != nil && o.Father != nil && dead.Father.ID == o.Father.ID
		switch {
		case mother && father:
			grieve(o, griefSibling)
		case mother || father:
			grieve(o, griefHalfSibling)
		}
	}
}

// dominantMood is the strongest mood that shows (≥ moodShown), or -1 if c
// looks calm.
func dominantMood(a *Affect) (Mood, float64) {
	best, top := Mood(-1), moodShown
	for m := range numMoods {
		if v := *a.mood(m); v >= top && (best < 0 || v > top) {
			best, top = m, v
		}
	}
	if best < 0 {
		return -1, 0
	}
	return best, top
}

// moodCode is the mood that shows on c's face (0 calm, 1 fear, 2 anger,
// 3 joy, 4 grief) and how strongly, for stream frames.
func (s *Sim) moodCode(c *Creature) (int, float64) {
	if s.opts.NoAffect {
		return 0, 0
	}
	m, v := dominantMood(&c.Affect)
	if m < 0 {
		return 0, 0
	}
	return int(m) + 1, v
}

func (s *Sim) moodView(c *Creature) *MoodView {
	if s.opts.NoAffect {
		return nil
	}
	a := &c.Affect
	tr := &c.Genome.Traits
	v := &MoodView{Fear: r3(a.Fear), Anger: r3(a.Anger), Joy: r3(a.Joy), Grief: r3(a.Grief),
		Reactivity: r3(gene(tr.Reactivity)), Recovery: r3(gene(tr.Recovery)), Cheer: r3(gene(tr.Cheer))}
	if m, _ := dominantMood(a); m >= 0 {
		v.Dominant = moodNames[m]
	}
	return v
}

// --- temperament genes (genome.go calls these) --------------------------------

// Temperament is inherited like any trait (blended from both parents, a
// little mutation each generation; temperament is moderately heritable in
// twin studies, Saudino 2005), but the random numbers come from a hash of
// traits the world's generator has just drawn, not from the generator: adding
// these genes shifts no other draw, so a world grows exactly as it did
// before, except for the moods.

// traitSeed mixes a genome's freshly drawn traits into a seed.
func traitSeed(t *Traits) uint64 {
	h := uint64(0x243F6A8885A308D3)
	for _, v := range []float64{t.Hue, t.Size, t.MaxSpeed, t.Vision, t.Lifespan, t.Metabolism} {
		h = (h ^ math.Float64bits(v)) * 0x100000001B3
		h ^= h >> 29
	}
	return h
}

// traitUniform is a well-mixed number in [0, 1) from a seed and a salt.
func traitUniform(seed, salt uint64) float64 {
	return hash01(int64(seed), int64(salt), 0xA5)
}

// traitNormal is about normally distributed (mean 0, deviation 1): the sum of
// four uniforms, rescaled.
func traitNormal(seed, salt uint64) float64 {
	sum := 0.0
	for i := range uint64(4) {
		sum += traitUniform(seed, salt*8+i)
	}
	return (sum - 2) * math.Sqrt(3)
}

// initTemperament gives a founding genome a temperament near average.
func (t *Traits) initTemperament() {
	seed := traitSeed(t)
	t.Reactivity = 0.9 + 0.2*traitUniform(seed, 1)
	t.Recovery = 0.9 + 0.2*traitUniform(seed, 2)
	t.Cheer = 0.9 + 0.2*traitUniform(seed, 3)
}

// blendTemperament sets a child's temperament between its parents'.
func (t *Traits) blendTemperament(a, b *Traits) {
	seed := traitSeed(t)
	blend := func(x, y float64, salt uint64) float64 {
		x, y = gene(x), gene(y)
		return x + (y-x)*traitUniform(seed, salt)
	}
	t.Reactivity = blend(a.Reactivity, b.Reactivity, 4)
	t.Recovery = blend(a.Recovery, b.Recovery, 5)
	t.Cheer = blend(a.Cheer, b.Cheer, 6)
}

// mutateTemperament drifts the temperament a little, after the other traits
// have mutated.
func (t *Traits) mutateTemperament() {
	seed := traitSeed(t)
	t.Reactivity = clamp(gene(t.Reactivity)+traitNormal(seed, 7)*0.05, reactivityMin, reactivityMax)
	t.Recovery = clamp(gene(t.Recovery)+traitNormal(seed, 8)*0.05, recoveryMin, recoveryMax)
	t.Cheer = clamp(gene(t.Cheer)+traitNormal(seed, 9)*0.05, cheerMin, cheerMax)
}

// upgradeTemperament gives genomes from before save v11 an average temperament.
func (t *Traits) upgradeTemperament() {
	t.Reactivity, t.Recovery, t.Cheer = gene(t.Reactivity), gene(t.Recovery), gene(t.Cheer)
}
