package sim

import "math"

// Neurogenesis: brains grow during life. A world that keeps surprising a
// creature (rewards better or worse than it expected) makes new neurons
// grow, faster in childhood and only on a full enough stomach. A new neuron
// is recruited Hebbian-fashion: its incoming weights copy what the senses and
// the inherited neurons are doing at that moment, so it becomes a detector of
// that situation; it starts silent towards the outputs and learns what to do
// about it with the same reward-driven rule as the rest of the brain. Grown
// neurons that stay unused are pruned again, and a starving body sheds the
// least useful one first. They are not inherited (the gene for how readily
// they grow is), and every neuron, inherited or grown, costs energy.
//
// Grown neurons sit after the inherited ones. Each listens to the senses and
// to the inherited neurons' new state (a second, deeper layer) and speaks to
// the outputs; none feeds back into the inherited circuits, so growth can't
// upset the reflexes they carry. Their weights are stored neuron by neuron,
// so a neuron is added or removed as one block.
//
// Growing and pruning draw no numbers from the world's random stream (they
// hash the creature and the tick instead), so they don't shift anything else.

const (
	neurogenesisInit = 0.3
	neurogenesisMax  = 1.5
	// growScale turns gene × surprise × development into a chance per learning step.
	growScale = 0.2
	// growFloor: no new neurons below this much energy.
	growFloor = 0.6
	// noveltyRate is how fast the felt surprise follows the actual.
	noveltyRate = 0.02
	// A grown neuron older than pruneAge seconds whose use stays under pruneBelow is pruned.
	pruneAge   = 2 * SecondsPerYear
	pruneBelow = 0.01
	useRate    = 0.01
	// starveShed: below this energy the brain sheds its least useful grown neuron now and then.
	starveShed = 0.12
)

// thinkGrown runs the grown neurons on the senses and the inherited neurons'
// new state, adding their say to the outputs (still before the squashing).
func (m *Mind) thinkGrown(in *[NumInputs]float64, inherited []float64, out *[NumOutputs]float64) {
	G, H := m.Grown, len(inherited)
	for k := range G {
		a := float64(m.GBias[k])
		gin := m.GIn[k*NumInputs : (k+1)*NumInputs]
		for i, v := range in {
			if v != 0 {
				a += v * float64(gin[i])
			}
		}
		rec := m.GRec[k*H : (k+1)*H]
		for j, v := range inherited {
			a += v * float64(rec[j])
		}
		act := m.GAct[k] + (math.Tanh(a)-m.GAct[k])/float64(m.GTau[k])
		m.GAct[k] = act
		if act == 0 {
			continue
		}
		gout := m.GOut[k*NumOutputs : (k+1)*NumOutputs]
		for o := range out {
			out[o] += act * float64(gout[o])
		}
	}
}

// growthValid reports whether the grown neurons' arrays agree with their count.
func (m *Mind) growthValid(H int) bool {
	G := m.Grown
	return G >= 0 && G <= brainSanity && len(m.GIn) == G*NumInputs && len(m.GRec) == G*H &&
		len(m.GOut) == G*NumOutputs && len(m.GBias) == G && len(m.GTau) == G && len(m.GBorn) == G &&
		len(m.GUse) == G && len(m.EGIn) == G*NumInputs && len(m.EGOut) == G*NumOutputs && len(m.GAct) == G
}

// brainSize is how many neurons a creature thinks with: inherited and grown.
func brainSize(c *Creature) int {
	n := c.Genome.Hidden
	if c.Mind != nil {
		n += c.Mind.Grown
	}
	return n
}

// hash01 is a well-mixed number in [0, 1) from a creature, a tick and a salt.
func hash01(id, tick int64, salt uint64) float64 {
	h := uint64(id)*0x9E3779B97F4A7C15 ^ uint64(tick)*0xC2B2AE3D27D4EB4F ^ salt*0x165667B19E3779F9
	h ^= h >> 31
	h *= 0xD6E8FEB86659FD93
	h ^= h >> 32
	return float64(h>>11) / (1 << 53)
}

// development is how much faster a young brain grows: three times at birth,
// easing to once by fifteen.
func (s *Sim) development(c *Creature) float64 {
	years := s.age(c) / SecondsPerYear
	return 1 + 2*math.Max(0, 1-years/15)
}

// grow runs after each learning step: it follows the felt surprise, keeps
// track of how much each grown neuron is used, and grows or prunes one.
func (s *Sim) grow(c *Creature, surprise float64) {
	m := c.Mind
	m.Novelty += (math.Abs(surprise) - m.Novelty) * noveltyRate
	for k := range m.Grown {
		reach := 0.0
		for _, w := range m.GOut[k*NumOutputs : (k+1)*NumOutputs] {
			reach += math.Abs(float64(w))
		}
		u := float64(m.GUse[k])
		m.GUse[k] = float32(u + (math.Abs(m.GAct[k])*reach-u)*useRate)
	}
	if s.opts.NoNeurogenesis {
		return
	}
	years := s.age(c) / SecondsPerYear
	// Prune the weakest grown neuron once it has had its chance, or when starving.
	if m.Grown > 0 {
		weakest := 0
		for k := 1; k < m.Grown; k++ {
			if m.GUse[k] < m.GUse[weakest] {
				weakest = k
			}
		}
		matured := (years - float64(m.GBorn[weakest])) * SecondsPerYear
		starving := c.Energy < starveShed && hash01(c.ID, s.tick, 2) < 0.05
		if starving || matured > pruneAge && float64(m.GUse[weakest]) < pruneBelow {
			m.removeGrown(weakest, c.Genome.Hidden)
			s.pruned++
			return
		}
	}
	tr := c.Genome.Traits
	if tr.Neurogenesis <= 0 || c.Energy < growFloor {
		return
	}
	chance := tr.Neurogenesis * m.Novelty * s.development(c) * growScale
	if hash01(c.ID, s.tick, 1) < chance {
		m.addGrown(c, years)
		s.grown++
	}
}

// addGrown recruits a neuron tuned to the present: its incoming weights are
// the current senses and inherited activity (normalised), with a bias that
// keeps it quiet in situations unlike this one; it starts with nothing to say.
func (m *Mind) addGrown(c *Creature, years float64) {
	H := c.Genome.Hidden
	G := m.Grown
	norm := 0.0
	for _, v := range c.input {
		norm += v * v
	}
	inner := 0.0
	for _, v := range c.Hidden {
		inner += v * v
	}
	norm, inner = math.Sqrt(norm), math.Sqrt(inner)
	gin := make(weights, NumInputs)
	for i, v := range c.input {
		if norm > 0 {
			gin[i] = float32(1.2 * v / norm)
		}
	}
	rec := make(weights, H)
	for j, v := range c.Hidden {
		if inner > 0 {
			rec[j] = float32(0.6 * v / inner)
		}
	}
	m.GIn = append(m.GIn, gin...)
	m.GRec = append(m.GRec, rec...)
	m.GOut = append(m.GOut, make(weights, NumOutputs)...)
	// Fires (above zero) only when what it sees resembles the moment it was born in.
	m.GBias = append(m.GBias, float32(-0.6*(1.2*norm+0.6*inner)))
	m.GTau = append(m.GTau, float32(1+3*hash01(c.ID, int64(G), 3)))
	m.GBorn = append(m.GBorn, float32(years))
	m.GUse = append(m.GUse, 0.05) // a grace start, so it isn't pruned before it has learned anything
	m.EGIn = append(m.EGIn, make(weights, NumInputs)...)
	m.EGOut = append(m.EGOut, make(weights, NumOutputs)...)
	m.GAct = append(m.GAct, 0)
	m.Grown++
	m.Grew++
}

// removeGrown takes grown neuron k out.
func (m *Mind) removeGrown(k, H int) {
	cut := func(w weights, size int) weights { return append(w[:k*size], w[(k+1)*size:]...) }
	m.GIn = cut(m.GIn, NumInputs)
	m.GRec = cut(m.GRec, H)
	m.GOut = cut(m.GOut, NumOutputs)
	m.GBias = cut(m.GBias, 1)
	m.GTau = cut(m.GTau, 1)
	m.GBorn = cut(m.GBorn, 1)
	m.GUse = cut(m.GUse, 1)
	m.EGIn = cut(m.EGIn, NumInputs)
	m.EGOut = cut(m.EGOut, NumOutputs)
	m.GAct = append(m.GAct[:k], m.GAct[k+1:]...)
	m.Grown--
	m.Pruned++
}

// learnGrown applies the learning rule to the grown neurons: their incoming
// weights sharpen (or blur) what they detect, their outgoing weights learn
// what to do about it. What they may say is held to the same bound as what
// learning may change in the inherited brain (the plasticity gene), so a
// crowd of new neurons can't shout down the instincts.
func (m *Mind) learnGrown(c *Creature, decay, step, keep, limit float32) {
	for k := range m.Grown {
		post := float32(m.GAct[k])
		row := k * NumInputs
		for i, pre := range c.input {
			e := decay*m.EGIn[row+i] + float32(pre)*post
			m.EGIn[row+i] = e
			m.GIn[row+i] = clamp32(m.GIn[row+i]+step*e, -maxWeight, maxWeight)
		}
		row = k * NumOutputs
		for o := range NumOutputs {
			act := c.output[o]
			if o != outTurn {
				act = (act - 0.5) * 2
			}
			e := decay*m.EGOut[row+o] + post*float32(act)
			m.EGOut[row+o] = e
			m.GOut[row+o] = clamp32((m.GOut[row+o]+step*e)*keep, -limit, limit)
		}
	}
}
