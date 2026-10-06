package sim

import (
	"fmt"
	"math"
	"slices"
)

// Leaders. In RAGE a group has a leader its members follow, and a new one is
// appointed when the leader is gone (PedGroup/PedGroup.h,
// CPedGroupMembership::AppointNewLeader; event/EventLeader.h tells members
// what the leader does). Here nobody appoints anyone and nobody follows: the
// leader is simply the adult the village's other adults trust most, counted
// from their own relations (relations.go), and people can only sense where
// that person is. Leadership in small-scale societies works this way: a
// headman or "big man" holds influence, not command, for as long as people
// trust him, and loses it when they stop (Service 1962; Glowacki & von Rueden
// 2015, Phil. Trans. R. Soc. B 370: 20150010).

const (
	// supportTrust: trust that makes someone a supporter, about one act of
	// kindness seen first hand (give is +0.18, seen +0.12; perception.go).
	supportTrust = 0.1
	// A leader needs at least two adult supporters besides net trust: one
	// admirer is a friend, not a following. Estimate. (Requiring support
	// from another household as well left the live world's big clan houses,
	// dozens of people each, without any leader.)
	minSupporters = 2
	// leaderMargin and leaderLead: a rival displaces a leader who still has
	// support only with more than leaderLead times their trust plus
	// leaderMargin (about two kind acts seen first hand). Standing, once
	// gained, carries weight of its own: headmen and "big men" usually keep
	// it for many years, often until old age or death, rather than changing
	// hands with every shift of favour (Service 1962; Glowacki & von Rueden
	// 2015). With a 0.2 margin alone the live world changed leader every
	// seven months on average. Estimates.
	leaderMargin = 0.2
	leaderLead   = 1.5
	// standingYears: a leader's standing follows the trust they hold with a
	// memory of about this long, so they keep office through a bad season
	// or the death of a few supporters and lose it only when support stays
	// gone. Before this, nine in ten leaders of the live world lost office
	// within about a year, mostly when two of a handful of supporters died.
	// Estimate.
	standingYears = 3.0
	// leaderEventGap: one leadership report per village per two years at most.
	leaderEventGap = 2 * SecondsPerYear
)

// tally is the trust one candidate holds among a village's adults.
type tally struct {
	trust      float64
	supporters int
}

func (t tally) eligible() bool {
	return t.supporters >= minSupporters && t.trust > 0
}

// keeps reports whether a sitting leader still has some standing left.
func keeps(standing float64) bool { return standing > 0 }

// tallyTrust counts, for every adult of v, the trust the other adults of v
// have in them, into vs.tally.
func (s *Sim) tallyTrust(v *village) {
	vs := s.vil()
	if vs.tally == nil {
		vs.tally = map[int64]tally{}
	}
	clear(vs.tally)
	if s.opts.NoPerception {
		// Without personal memories everyone sees the same reputation, as
		// opinion() does: each adult is trusted alike by all the others.
		adults := 0
		for _, c := range v.members {
			if s.adult(c) {
				adults++
			}
		}
		for _, c := range v.members {
			if !s.adult(c) {
				continue
			}
			n := adults - 1
			t := tally{trust: c.Reputation * float64(n)}
			if c.Reputation >= supportTrust {
				t.supporters = n
			}
			vs.tally[c.ID] = t
		}
		return
	}
	for _, voter := range v.members {
		if !s.adult(voter) {
			continue // children don't choose leaders
		}
		for _, r := range voter.Relations {
			cand := s.byID[r.Person.ID]
			if cand == nil || cand == voter || cand.Health <= 0 || cand.VillageID != v.ID || !s.adult(cand) {
				continue
			}
			w := s.relationTrust(r)
			t := vs.tally[cand.ID]
			t.trust += w
			if w >= supportTrust {
				t.supporters++
			}
			vs.tally[cand.ID] = t
		}
	}
}

// elect finds v's leader from v.members: the incumbent while they keep
// enough support and nobody clearly outranks them, else the most trusted
// eligible adult (the lower ID on a tie), else nobody.
func (s *Sim) elect(v *village) {
	vs := s.vil()
	s.tallyTrust(v)
	var best *Creature
	var bestT tally
	for _, c := range v.members {
		t, ok := vs.tally[c.ID]
		if !ok || !t.eligible() {
			continue
		}
		if best == nil || t.trust > bestT.trust || t.trust == bestT.trust && c.ID < best.ID {
			best, bestT = c, t
		}
	}
	if inc := s.living(v.Leader); inc != nil && inc.VillageID == v.ID && s.adult(inc) {
		// Elections run every villageEvery; the standing follows the trust held now.
		t := vs.tally[inc.ID]
		v.Standing += (t.trust - v.Standing) * math.Min(1, villageEvery*dt/(standingYears*SecondsPerYear))
		if keeps(v.Standing) && (best == nil || bestT.trust <= math.Max(v.Standing, t.trust)*leaderLead+leaderMargin) {
			best, bestT = inc, t
		}
	}
	v.Trust, v.Supporters = r3(bestT.trust), bestT.supporters
	s.setLeader(v, best)
}

// setLeader makes c (nil for nobody) v's leader and reports the change.
func (s *Sim) setLeader(v *village, c *Creature) {
	var id int64
	if c != nil {
		id = c.ID
	}
	if id == v.Leader {
		return
	}
	vs := s.vil()
	prev, prevName := v.Leader, v.LeaderName
	prevGone := prev != 0 && s.living(prev) == nil
	v.Leader, v.LeaderName, v.LeaderSince, v.Standing = 0, "", 0, 0
	if c == nil {
		v.Trust, v.Supporters = 0, 0
		if prev != 0 && s.villageNews(v, false) {
			s.event("village", fmt.Sprintf("Desa %s kini tanpa pemimpin", v.Name), 0)
		}
		return
	}
	v.Leader, v.LeaderName, v.LeaderSince = c.ID, c.Name, s.time()
	v.Standing = vs.tally[c.ID].trust
	v.Leaders++
	vs.leaderChanges++
	if !s.villageNews(v, prevGone) {
		return
	}
	switch {
	case prev != 0 && prevGone:
		s.event("village", fmt.Sprintf("%s menjadi pemimpin Desa %s, menggantikan mendiang %s", c.Name, v.Name, prevName), c.ID)
	case prev != 0:
		s.event("village", fmt.Sprintf("%s menggantikan %s sebagai pemimpin Desa %s", c.Name, prevName, v.Name), c.ID)
	default:
		s.event("village", fmt.Sprintf("%s menjadi pemimpin Desa %s", c.Name, v.Name), c.ID)
	}
}

// villageNews reports whether a leadership change in v may be reported now:
// at most one per village every leaderEventGap (a leader's death is always
// news), within the check's budget.
func (s *Sim) villageNews(v *village, death bool) bool {
	if !death && v.LastNews != 0 && s.time()-v.LastNews < leaderEventGap {
		return false
	}
	if !s.spendNews() {
		return false
	}
	v.LastNews = s.time()
	return true
}

// succession replaces leaders who died or moved out since the last count.
func (s *Sim) succession() {
	for _, v := range s.vil().list {
		if v.Leader == 0 {
			continue
		}
		if l := s.living(v.Leader); l == nil || l.VillageID != v.ID {
			s.elect(v)
		}
	}
}

// senseVillage fills inOwnVillage, inLeaderNear, inLeaderSide and inIsLeader.
// The leader is felt like everyone else close by: by how near (1 at arm's
// length, fading to 0 at the edge of sight, as the other "near" senses) and
// on which side (the sine of the bearing, -1 left … 1 right, 0 ahead or
// behind), when in sight distance with nothing in between. There is no field
// of view: a familiar voice and figure are noticed all around.
func (s *Sim) senseVillage(c *Creature, in *[NumInputs]float64) {
	if s.opts.NoVillages || c.VillageID == 0 {
		return
	}
	vs := s.vil()
	v := vs.byID[c.VillageID]
	if v == nil {
		return
	}
	if i, ok := s.terrain.indexAt(c.X, c.Y); ok && i < len(vs.land) && int64(vs.land[i]) == v.ID {
		in[inOwnVillage] = 1
	}
	switch v.Leader {
	case 0:
		return
	case c.ID:
		in[inIsLeader] = 1
		return
	}
	l := s.byID[v.Leader]
	if l == nil || l.Health <= 0 {
		return
	}
	dx, dy := l.X-c.X, l.Y-c.Y
	vision := c.Genome.Traits.Vision * (0.55 + 0.45*s.eco.Light())
	if math.Abs(dx) >= vision || math.Abs(dy) >= vision {
		return
	}
	d := math.Hypot(dx, dy)
	if d >= vision || !s.opts.NoPerception && !s.lineOfSight(c.X, c.Y, l.X, l.Y) {
		return
	}
	in[inLeaderNear] = 1 - d/vision
	if d > 0 {
		sin, cos := math.Sincos(c.Heading)
		in[inLeaderSide] = (dy*cos - dx*sin) / d
	}
}

// isLeader reports whether c leads a village.
func (s *Sim) isLeader(c *Creature) bool {
	if c.VillageID == 0 {
		return false
	}
	v := s.vil().byID[c.VillageID]
	return v != nil && v.Leader == c.ID
}

// Attitudes between villages, for the observer only, after RAGE's
// relationship groups (Peds/Relationships.h, RelationshipEnums.h: respect,
// like, ignore, dislike, hate between groups): here the mean trust one
// village's adults have in the people of another they know.
func attitudeLabel(trust float64, known int) string {
	switch {
	case known == 0:
		return "asing"
	case trust >= 0.5:
		return "hormat"
	case trust >= 0.15:
		return "suka"
	case trust > -0.15:
		return "acuh"
	case trust > -0.5:
		return "tidak suka"
	}
	return "benci"
}

// attitudes lists how v's adults regard the other villages.
func (s *Sim) attitudes(v *village, members []*Creature) []VillageAttitude {
	vs := s.vil()
	sum, n := map[int64]float64{}, map[int64]int{}
	for _, c := range members {
		if !s.adult(c) {
			continue
		}
		for _, r := range c.Relations {
			o := s.byID[r.Person.ID]
			if o == nil || o.VillageID == 0 || o.VillageID == v.ID {
				continue
			}
			sum[o.VillageID] += s.relationTrust(r)
			n[o.VillageID]++
		}
	}
	out := []VillageAttitude{}
	for _, w := range vs.list {
		if w == v {
			continue
		}
		mean := 0.0
		if n[w.ID] > 0 {
			mean = sum[w.ID] / float64(n[w.ID])
		}
		out = append(out, VillageAttitude{ID: w.ID, Name: w.Name, Trust: r3(mean), Known: n[w.ID], Attitude: attitudeLabel(mean, n[w.ID])})
	}
	slices.SortStableFunc(out, func(a, b VillageAttitude) int { return b.Known - a.Known })
	return out
}
