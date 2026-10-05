package sim

import "math"

// Lifetime learning: a reward-modulated Hebbian ("three-factor") rule. Each
// plastic weight keeps an eligibility trace of how often its two neurons were
// active together; when the body's wellbeing changes better or worse than the
// creature has come to expect, the traced weights move in that direction.
// The reward comes only from the body (energy, water, health, pain), never
// from tasks, so what is learned is whatever helped this creature live.
// Learned changes are not inherited.

const (
	learnEvery   = TicksPerSecond / 2 // learning steps happen twice a second
	rewardScale  = 20.0               // wellbeing change per second → reward (≈ ±1 for a meal or a blow)
	painPenalty  = 0.5                // extra negative reward while being attacked
	baselineRate = 0.05               // how fast the expected reward follows the actual
	forgetting   = 0.0005             // learned changes fade by this share per step
	brainCost    = 0.00004            // energy per hidden neuron per second
)

// Mind is a creature's lifetime-learned state on top of its genome.
type Mind struct {
	DIn  weights `json:"dIn"`  // learned change of input→hidden weights
	DOut weights `json:"dOut"` // learned change of hidden→output weights
	EIn  weights `json:"eIn"`  // eligibility traces
	EOut weights `json:"eOut"`

	Wellbeing float64 `json:"wellbeing"` // energy + hydration + health at the last step
	Baseline  float64 `json:"baseline"`  // expected reward
	Reward    float64 `json:"reward"`    // last reward, felt as "imbalan terakhir"

	// Effective weights (inherited + learned), rebuilt from the above.
	wIn, wOut weights
}

func newMind(g *Genome) *Mind {
	m := &Mind{
		DIn:  make(weights, len(g.WIn)),
		DOut: make(weights, len(g.WOut)),
		EIn:  make(weights, len(g.WIn)),
		EOut: make(weights, len(g.WOut)),
	}
	m.rebuild(g)
	return m
}

func (m *Mind) valid(g *Genome) bool {
	return len(m.DIn) == len(g.WIn) && len(m.EIn) == len(g.WIn) &&
		len(m.DOut) == len(g.WOut) && len(m.EOut) == len(g.WOut)
}

// rebuild recomputes the effective weights.
func (m *Mind) rebuild(g *Genome) {
	m.wIn = make(weights, len(g.WIn))
	m.wOut = make(weights, len(g.WOut))
	for i, w := range g.WIn {
		m.wIn[i] = clamp32(w+m.DIn[i], -maxWeight, maxWeight)
	}
	for i, w := range g.WOut {
		m.wOut[i] = clamp32(w+m.DOut[i], -maxWeight, maxWeight)
	}
}

func wellbeing(c *Creature) float64 {
	return clamp(c.Energy, 0, 1) + clamp(c.Hydration, 0, 1) + clamp(c.Health, 0, 1)
}

// learnStep updates the reward and, if this world learns, the brain.
func (s *Sim) learnStep(c *Creature) {
	m, g := c.Mind, c.Genome
	wb := wellbeing(c)
	r := (wb - m.Wellbeing) / (learnEvery * dt) * rewardScale
	if c.Hurt > 0 {
		r -= painPenalty
	}
	m.Wellbeing = wb
	r = clamp(r, -1, 1)
	surprise := r - m.Baseline
	m.Baseline += (r - m.Baseline) * baselineRate
	m.Reward = r
	tr := g.Traits
	if s.noPlasticity() || tr.LearningRate <= 0 || tr.Plasticity <= 0 {
		return
	}

	H := g.Hidden
	decay := float32(tr.TraceDecay)
	step := float32(tr.LearningRate * surprise)
	limit := float32(tr.Plasticity)
	keep := float32(1 - forgetting)
	update := func(d, e, eff, base weights, i int, coactive float32) {
		e[i] = decay*e[i] + coactive
		d[i] = clamp32((d[i]+step*e[i])*keep, -limit, limit)
		eff[i] = clamp32(base[i]+d[i], -maxWeight, maxWeight)
	}
	for i, pre := range c.input {
		row := i * H
		p := float32(pre)
		for h := range H {
			update(m.DIn, m.EIn, m.wIn, g.WIn, row+h, p*float32(c.Hidden[h]))
		}
	}
	for h := range H {
		post := float32(c.Hidden[h])
		row := h * NumOutputs
		for o := range NumOutputs {
			act := c.output[o]
			if o != outTurn {
				act = (act - 0.5) * 2 // sigmoid outputs centred on "undecided"
			}
			update(m.DOut, m.EOut, m.wOut, g.WOut, row+o, post*float32(act))
		}
	}
}

// learnedPerNeuron tells how much each hidden neuron's weights changed
// through learning, relative to the most changed one (0–1).
func learnedPerNeuron(c *Creature) []float64 {
	g, m := c.Genome, c.Mind
	H := g.Hidden
	out := make([]float64, H)
	if m == nil {
		return out
	}
	top := 0.0
	for h := range H {
		sum := 0.0
		for i := range NumInputs {
			sum += math.Abs(float64(m.DIn[i*H+h]))
		}
		for o := range NumOutputs {
			sum += math.Abs(float64(m.DOut[h*NumOutputs+o]))
		}
		out[h] = sum
		top = math.Max(top, sum)
	}
	if top > 0 {
		for h := range out {
			out[h] = r3(out[h] / top)
		}
	}
	return out
}
