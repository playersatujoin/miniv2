package sim

import (
	"math"
	"slices"
)

const maxRelations = 16

// Relation is an individual's experience, not a world-wide reputation.
type Relation struct {
	Person Ref     `json:"person"`
	Trust  float64 `json:"trust"`
	Tick   int64   `json:"tick"`
	Met    int     `json:"met"`
}

func (s *Sim) relationTrust(r Relation) float64 {
	return r.Trust * math.Exp(-math.Max(0, float64(s.tick-r.Tick)*dt)/(30*SecondsPerYear))
}

func (s *Sim) rememberRelation(c, other *Creature, change float64) {
	s.shiftRelation(c, other, change, 1)
}

// shiftRelation changes c's trust in other by change. met is 1 for
// something c went through or saw with them, 0 for what c only heard about
// them: hearsay is not a meeting, and it never pushes out of a full table
// someone c feels more strongly about.
func (s *Sim) shiftRelation(c, other *Creature, change float64, met int) {
	for i := range c.Relations {
		r := &c.Relations[i]
		if r.Person.ID == other.ID {
			r.Trust = clamp(s.relationTrust(*r)+change, -1, 1)
			r.Tick, r.Met = s.tick, r.Met+met
			return
		}
	}
	r := Relation{Ref{other.ID, other.Name}, clamp(change, -1, 1), s.tick, met}
	if len(c.Relations) < maxRelations {
		c.Relations = append(c.Relations, r)
		return
	}
	weakest := 0
	for i := 1; i < len(c.Relations); i++ {
		if math.Abs(s.relationTrust(c.Relations[i])) < math.Abs(s.relationTrust(c.Relations[weakest])) {
			weakest = i
		}
	}
	if met == 0 && math.Abs(r.Trust) <= math.Abs(s.relationTrust(c.Relations[weakest])) {
		return
	}
	c.Relations[weakest] = r
}

// heardOpinion moves c's trust in other a share (weight) of the way towards
// view, what someone else told c of them: opinions pooled with those of
// others, as in DeGroot (1974), so hearing the same view again converges on
// it instead of piling up.
func (s *Sim) heardOpinion(c, other *Creature, view, weight float64) {
	s.shiftRelation(c, other, (view-s.opinion(c, other))*weight, 0)
}

// strongestOpinion is the living person, other than c and skip, whom c
// trusts or distrusts most, if that feeling is at least floor; it is what c
// would bring up about others when talking with skip.
func (s *Sim) strongestOpinion(c, skip *Creature, floor float64) (*Creature, float64) {
	var who *Creature
	view := 0.0
	for _, r := range c.Relations {
		if r.Person.ID == skip.ID || r.Person.ID == c.ID {
			continue
		}
		t := s.relationTrust(r)
		if math.Abs(t) < floor || math.Abs(t) <= math.Abs(view) {
			continue
		}
		if o := s.living(r.Person.ID); o != nil {
			who, view = o, t
		}
	}
	return who, view
}

func (s *Sim) opinion(c, other *Creature) float64 {
	if s.opts.NoPerception {
		return other.Reputation
	}
	for _, r := range c.Relations {
		if r.Person.ID == other.ID {
			return s.relationTrust(r)
		}
	}
	return 0 // an unknown person is not known to be kind or dangerous
}

func (s *Sim) relationViews(c *Creature) []Relation {
	out := append([]Relation{}, c.Relations...)
	for i := range out {
		out[i].Trust = r3(s.relationTrust(out[i]))
	}
	slices.SortStableFunc(out, func(a, b Relation) int {
		if math.Abs(a.Trust) > math.Abs(b.Trust) {
			return -1
		}
		if math.Abs(a.Trust) < math.Abs(b.Trust) {
			return 1
		}
		return 0
	})
	return out
}
