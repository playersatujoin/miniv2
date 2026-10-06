package sim

import "math"

// Perception uses the simulation's terrain, never the observer's camera.
// Water blocks feet but not sight. Solid objects and land block sight.
func (s *Sim) lineOfSight(ax, ay, bx, by float64) bool {
	// Grid traversal visits each crossed tile once, rather than sampling the
	// same tile repeatedly. No allocation and no dependence on frame rate.
	x, y, ex, ey := int(math.Floor(ax)), int(math.Floor(ay)), int(math.Floor(bx)), int(math.Floor(by))
	dx, dy := bx-ax, by-ay
	sx, sy := 1, 1
	if dx < 0 {
		sx = -1
	}
	if dy < 0 {
		sy = -1
	}
	dtx, dty := math.Inf(1), math.Inf(1)
	tx, ty := math.Inf(1), math.Inf(1)
	if dx != 0 {
		dtx = 1 / math.Abs(dx)
		if sx > 0 {
			tx = (float64(x+1) - ax) * dtx
		} else {
			tx = (ax - float64(x)) * dtx
		}
	}
	if dy != 0 {
		dty = 1 / math.Abs(dy)
		if sy > 0 {
			ty = (float64(y+1) - ay) * dty
		} else {
			ty = (ay - float64(y)) * dty
		}
	}
	for n := absInt(ex-x) + absInt(ey-y) + 2; n > 0; n-- {
		if x == ex && y == ey {
			return true
		}
		if tx < ty {
			x += sx
			tx += dtx
		} else {
			y += sy
			ty += dty
		}
		if x == ex && y == ey {
			return true
		}
		i, ok := s.terrain.index(x, y)
		if !ok || s.terrain.blocked[i] && !s.terrain.water[i] {
			return false
		}
	}
	return false
}

func (s *Sim) sees(c *Creature, x, y float64) bool {
	dx, dy := x-c.X, y-c.Y
	d := math.Hypot(dx, dy)
	vision := c.Genome.Traits.Vision * (0.55 + 0.45*s.eco.Light())
	if d > vision {
		return false
	}
	// Touch is perceived even outside the forward field of view.
	if d > 0.8 && dx*math.Cos(c.Heading)+dy*math.Sin(c.Heading) < d*0.2588190451 {
		return false
	}
	return s.lineOfSight(c.X, c.Y, x, y)
}

const maxMemories = 12

// PerceivedEvent is private knowledge. A sound carries no actor identity.
type PerceivedEvent struct {
	Kind       string  `json:"kind"`
	Actor      *Ref    `json:"actor,omitempty"`
	Mode       string  `json:"mode"` // direct, seen, heard
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Tick       int64   `json:"tick"`
	Confidence float64 `json:"confidence"`
}

// witnessMargin (tiles) covers how far anyone may have moved since the
// spatial grid was rebuilt at the start of the step: a tick's walk at the
// top speed is 0.15 tiles, a shove 0.06, a baby rides with its mother.
const witnessMargin = 1.5

// perceiveEvent lets everyone who notices an act remember it: its target
// directly, onlookers who see the actor (and so know who it was), and those
// within earshot of a loud one (who only know something happened there).
// The act also leaves a stimulus (a fight, a theft) others may notice for a
// while, and moves the target's moods: a wrong angers, an assault also
// frightens, a gift gladdens.
//
// Witnesses are found through the spatial grid, after RAGE's entity scanner
// and crime witnesses (ai/EntityScanner.h, game/witness.h), within the
// farthest anyone can see or the act be heard, rather than by asking
// everyone; the checks on each are exact, so the witnesses are the same as
// if everyone were asked. Each witness keeps only their own memories, so the
// order they are visited in changes nothing.
func (s *Sim) perceiveEvent(actor, target *Creature, kind string, value, loudness float64) {
	switch kind {
	case "attack":
		x, y := actor.X, actor.Y
		if target != nil {
			x, y = (x+target.X)/2, (y+target.Y)/2
		}
		s.addStimulus(stimFight, x, y, 1, actor)
	case "steal":
		s.addStimulus(stimTheft, actor.X, actor.Y, 1, actor)
	}
	if target != nil && target != actor && target.Health > 0 {
		switch {
		case value < 0:
			s.feel(target, moodAnger, -value)
		case value > 0:
			s.feel(target, moodJoy, giftJoy*value)
		}
		if kind == "attack" {
			s.feel(target, moodFear, assaultFear)
		}
	}
	if s.opts.NoPerception {
		return
	}
	s.witnesses(actor, target, loudness, func(observer *Creature, mode string, confidence float64) {
		var identity *Ref
		if mode != "heard" {
			identity = &Ref{actor.ID, actor.Name}
			s.rememberRelation(observer, actor, value*confidence)
		}
		ev := PerceivedEvent{kind, identity, mode, actor.X, actor.Y, s.tick, confidence}
		observer.Memories = append(observer.Memories, ev)
		if len(observer.Memories) > maxMemories {
			copy(observer.Memories, observer.Memories[len(observer.Memories)-maxMemories:])
			observer.Memories = observer.Memories[:maxMemories]
		}
	})
}

// witnesses calls fn for everyone alive who perceives an act of actor's:
// the target first ("direct"), then whoever sees the actor or hears it.
func (s *Sim) witnesses(actor, target *Creature, loudness float64, fn func(o *Creature, mode string, confidence float64)) {
	if target != nil && target != actor && target.Health > 0 {
		fn(target, "direct", 1)
	}
	look := func(o *Creature) {
		if o == actor || o == target || o.Health <= 0 {
			return
		}
		if mode, confidence := s.witnessed(o, actor, loudness); mode != "" {
			fn(o, mode, confidence)
		}
	}
	if len(s.grid.cells) == 0 {
		// No grid yet (a world that hasn't stepped): ask everyone.
		for _, o := range s.creatures {
			look(o)
		}
		return
	}
	r := math.Max(s.maxVision()*(0.55+0.45*s.eco.Light()), loudness) + witnessMargin
	s.grid.near(actor.X, actor.Y, r, look)
}

// witnessed tells how observer perceives an act of actor's, if at all.
func (s *Sim) witnessed(observer, actor *Creature, loudness float64) (string, float64) {
	if s.sees(observer, actor.X, actor.Y) {
		return "seen", 0.65
	}
	if loudness <= 0 {
		return "", 0
	}
	if d := math.Hypot(observer.X-actor.X, observer.Y-actor.Y); d < loudness {
		confidence := 0.4 * (1 - d/loudness)
		if !s.lineOfSight(observer.X, observer.Y, actor.X, actor.Y) {
			confidence *= 0.4
		}
		return "heard", confidence
	}
	return "", 0
}

// maxVision is the farthest anyone alive can see by day, worked out once a
// tick (the vision gene is at most 9 tiles, but nothing here relies on that).
func (s *Sim) maxVision() float64 {
	w := s.stims()
	if w.visionTick != s.tick || w.vision == 0 {
		w.visionTick, w.vision = s.tick, 0
		for _, c := range s.creatures {
			w.vision = math.Max(w.vision, c.Genome.Traits.Vision)
		}
	}
	return w.vision
}

func (s *Sim) alarm(c *Creature) float64 {
	if s.opts.NoPerception {
		return 0
	}
	strength := 0.0
	for _, ev := range c.Memories {
		if ev.Kind == "attack" {
			age := float64(s.tick-ev.Tick) * dt
			strength = math.Max(strength, ev.Confidence*math.Max(0, 1-age/3))
		}
	}
	return strength
}
