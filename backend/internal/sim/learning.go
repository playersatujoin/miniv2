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

	// Neurons grown during this life (see growth.go), neuron by neuron.
	Grown int       `json:"grown,omitempty"`
	GIn   weights   `json:"gIn,omitempty"`   // [k*NumInputs + input]
	GRec  weights   `json:"gRec,omitempty"`  // [k*Hidden + inherited neuron]
	GOut  weights   `json:"gOut,omitempty"`  // [k*NumOutputs + output]
	GBias weights   `json:"gBias,omitempty"` // per grown neuron
	GTau  weights   `json:"gTau,omitempty"`
	GBorn weights   `json:"gBorn,omitempty"` // age in years when it grew
	GUse  weights   `json:"gUse,omitempty"`  // how much it has been used lately
	EGIn  weights   `json:"eGIn,omitempty"`  // eligibility traces
	EGOut weights   `json:"eGOut,omitempty"`
	GAct  []float64 `json:"gAct,omitempty"` // current activations
	// Novelty is the felt surprise (how unpredictable life has been lately);
	// Grew and Pruned count the neurons this life has gained and lost.
	Novelty float64 `json:"novelty,omitempty"`
	Grew    int     `json:"grew,omitempty"`
	Pruned  int     `json:"pruned,omitempty"`

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

// upgrade widens a mind saved with fewer senses or actions (see padInputs
// and padRows); what was learned keeps its place.
func (m *Mind) upgrade(g *Genome) {
	H := g.Hidden
	for _, w := range []*weights{&m.DIn, &m.EIn} {
		if n := len(*w); n == legacyInputs*H || n == healthInputs*H || n == alarmInputs*H || n == engineInputs*H {
			*w = padInputs(*w, H)
		}
	}
	for _, w := range []*weights{&m.DOut, &m.EOut} {
		if len(*w) == legacyOutputs*H {
			*w = padRows(*w, legacyOutputs, NumOutputs)
		}
	}
	if m.Grown > 0 {
		for _, w := range []*weights{&m.GIn, &m.EGIn} {
			for _, old := range []int{legacyInputs, healthInputs, alarmInputs, engineInputs} {
				if len(*w) == old*m.Grown {
					*w = padRows(*w, old, NumInputs)
					break
				}
			}
		}
		for _, w := range []*weights{&m.GOut, &m.EGOut} {
			if len(*w) == legacyOutputs*m.Grown {
				*w = padRows(*w, legacyOutputs, NumOutputs)
			}
		}
	}
}

func (m *Mind) valid(g *Genome) bool {
	return len(m.DIn) == len(g.WIn) && len(m.EIn) == len(g.WIn) &&
		len(m.DOut) == len(g.WOut) && len(m.EOut) == len(g.WOut) && m.growthValid(g.Hidden)
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
	s.grow(c, surprise)
	if s.noPlasticity() || tr.LearningRate <= 0 || tr.Plasticity <= 0 {
		return
	}

	H := g.Hidden
	decay := float32(tr.TraceDecay)
	step := float32(tr.LearningRate * surprise)
	limit := float32(tr.Plasticity)
	keep := float32(1 - forgetting)
	var hidden [stackNeurons]float32
	var post []float32
	if H <= stackNeurons {
		post = hidden[:H]
	} else {
		post = make([]float32, H)
	}
	for h := range post {
		post[h] = float32(c.Hidden[h])
	}
	for i, v := range c.input {
		pre, row := float32(v), i*H
		e, d, eff, base := m.EIn[row:row+H], m.DIn[row:row+H], m.wIn[row:row+H], g.WIn[row:row+H]
		for h, q := range post {
			e[h] = decay*e[h] + pre*q
			d[h] = clamp32((d[h]+step*e[h])*keep, -limit, limit)
			eff[h] = clamp32(base[h]+d[h], -maxWeight, maxWeight)
		}
	}
	var acts [NumOutputs]float32
	for o, act := range c.output {
		if o != outTurn {
			act = (act - 0.5) * 2 // sigmoid outputs centred on "undecided"
		}
		acts[o] = float32(act)
	}
	for h, q := range post {
		row := h * NumOutputs
		e, d, eff, base := m.EOut[row:row+NumOutputs], m.DOut[row:row+NumOutputs], m.wOut[row:row+NumOutputs], g.WOut[row:row+NumOutputs]
		for o, a := range acts {
			e[o] = decay*e[o] + q*a
			d[o] = clamp32((d[o]+step*e[o])*keep, -limit, limit)
			eff[o] = clamp32(base[o]+d[o], -maxWeight, maxWeight)
		}
	}
	m.learnGrown(c, decay, step, keep, limit)
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
