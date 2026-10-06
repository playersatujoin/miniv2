package sim

import "math"

// Personal space: bodies don't stand inside one another. After everyone has
// moved, anyone closer to another than their two body radii is eased apart,
// a share of the overlap each tick and never more than a short step, and
// never off dry ground into water or onto rock (those already wading,
// swimming or climbing are eased apart where they are; see shoveFree). A crowd round a well or a hearth so spreads into
// a ring instead of piling onto one spot. Every shove is worked out from
// where everyone stood before any of them moves, so the order of the
// creatures doesn't matter and the world stays deterministic. A baby in arms
// rides with its mother and is left out.

const (
	// spaceCell is the side (tiles) of the fine grid used to find who touches whom.
	spaceCell = 1.0
	// spaceEase is the share of an overlap undone per tick, spaceStep the most
	// anyone is shoved in one tick (tiles).
	spaceEase = 0.5
	spaceStep = 0.06
	// spaceReach covers the largest two bodies touching (body size is clamped to 1.3; a little spare).
	spaceReach = 2 * bodyRadius * 1.5
)

func (s *Sim) personalSpace() {
	if s.opts.NoPersonalSpace || len(s.creatures) < 2 {
		return
	}
	s.space.cell = spaceCell
	s.space.rebuild(s.terrain.w, s.terrain.h, s.creatures)
	if cap(s.shove) < len(s.creatures) {
		s.shove = make([][2]float64, len(s.creatures))
	}
	shove := s.shove[:len(s.creatures)]
	for i, c := range s.creatures {
		shove[i] = [2]float64{}
		if !s.solid(c) {
			continue
		}
		r := bodyRadius * c.Genome.Traits.Size
		s.space.near(c.X, c.Y, spaceReach, func(o *Creature) {
			if o == c || !s.solid(o) {
				return
			}
			want := r + bodyRadius*o.Genome.Traits.Size
			dx, dy := c.X-o.X, c.Y-o.Y
			d := math.Hypot(dx, dy)
			if d >= want {
				return
			}
			if d < 1e-9 {
				// Exactly on top of each other: part along a direction fixed by the pair.
				a := pairAngle(c.ID, o.ID)
				shove[i][0] += math.Cos(a) * want * spaceEase * 0.5
				shove[i][1] += math.Sin(a) * want * spaceEase * 0.5
				return
			}
			k := (want - d) * spaceEase * 0.5 / d
			shove[i][0] += dx * k
			shove[i][1] += dy * k
		})
	}
	for i, c := range s.creatures {
		px, py := shove[i][0], shove[i][1]
		l := math.Hypot(px, py)
		if l == 0 {
			continue
		}
		if l > spaceStep {
			px, py = px/l*spaceStep, py/l*spaceStep
		}
		r := bodyRadius * c.Genome.Traits.Size
		switch {
		case s.shoveFree(c, c.X+px, c.Y+py, px, py, r):
			c.X += px
			c.Y += py
		case s.shoveFree(c, c.X+px, c.Y, px, 0, r):
			c.X += px
		case s.shoveFree(c, c.X, c.Y+py, 0, py, r):
			c.Y += py
		}
	}
}

// solid reports whether c takes up room: alive and on its own feet.
func (s *Sim) solid(c *Creature) bool {
	return c.Health > 0 && s.carrier(c) == nil
}

// pairAngle is a direction fixed by two ids, pointing from b to a (and the
// opposite way round for the pair seen from b).
func pairAngle(a, b int64) float64 {
	lo, hi := min(a, b), max(a, b)
	h := uint64(lo)*0x9E3779B97F4A7C15 ^ uint64(hi)*0xC2B2AE3D27D4EB4F
	h ^= h >> 29
	angle := float64(h%6283) / 1000
	if a > b {
		angle += math.Pi
	}
	return angle
}
