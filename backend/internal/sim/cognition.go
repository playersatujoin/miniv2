package sim

// Cognition keeps the last sensory sample and intent across motor ticks and
// save/restore. Every creature thinks at 5 Hz, staggered by id; physics, bodily
// costs and work still run at 20 Hz. Camera distance never changes this budget.
type Cognition struct {
	Input  [NumInputs]float64  `json:"input"`
	Output [NumOutputs]float64 `json:"output"`
}

func (s *Sim) decisionDue(c *Creature) bool {
	return s.opts.NoAIBudget || c.Cognition == nil || (s.tick+c.ID)%4 == 0
}

func (s *Sim) decide(c *Creature) {
	if !s.decisionDue(c) {
		c.input, c.output = c.Cognition.Input, c.Cognition.Output
		return
	}
	s.sense(c)
	think(c.Genome, c.Mind.wIn, c.Mind.wOut, c.Mind, &c.input, c.Hidden, &c.output)
	if c.Cognition == nil {
		c.Cognition = &Cognition{}
	}
	c.Cognition.Input, c.Cognition.Output = c.input, c.output
}
