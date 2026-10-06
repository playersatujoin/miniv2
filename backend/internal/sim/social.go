package sim

import (
	"fmt"
	"math"

	"miniv2/backend/internal/chem"
)

const (
	giveRange       = 1.5
	stealRange      = 1.2
	houseStealRange = 1.5
	attackRange     = 1.0
	attackDamage    = 0.15 // health per blow, scaled by size and weapon
	attackCost      = 0.02 // energy per blow
	crimeEventGap   = 5.0  // seconds between reported thefts and assaults
	kindEventGap    = 20.0 // seconds between reported kindnesses
)

func isParent(p, c *Creature) bool {
	return c.Mother != nil && c.Mother.ID == p.ID || c.Father != nil && c.Father.ID == p.ID
}

// kin is family: same household, spouses, parents and children, or siblings.
func (s *Sim) kin(a, b *Creature) bool {
	switch {
	case a.HouseID != 0 && a.HouseID == b.HouseID,
		a.Spouse != nil && a.Spouse.ID == b.ID,
		b.Spouse != nil && b.Spouse.ID == a.ID,
		isParent(a, b), isParent(b, a):
		return true
	}
	return a.Mother != nil && b.Mother != nil && a.Mother.ID == b.Mother.ID ||
		a.Father != nil && b.Father != nil && a.Father.ID == b.Father.ID
}

// kinHouse reports whether the house belongs to c's family.
func (s *Sim) kinHouse(c *Creature, st *Structure) bool {
	if st.ID == c.HouseID {
		return true
	}
	owner := s.byID[st.Owner]
	return owner != nil && s.kin(c, owner)
}

func (s *Sim) livingSpouse(c *Creature) *Creature {
	if c.Spouse == nil {
		return nil
	}
	return s.living(c.Spouse.ID)
}

// bond makes a couple spouses on their first child together, if both are free.
func (s *Sim) bond(f, m *Creature) {
	if s.livingSpouse(f) != nil || s.livingSpouse(m) != nil {
		return
	}
	f.Spouse = &Ref{m.ID, m.Name}
	m.Spouse = &Ref{f.ID, f.Name}
	s.event("family", fmt.Sprintf("%s & %s menjadi pasangan", m.Name, f.Name), f.ID)
	// The one without a home of their own moves in with the one who has one.
	switch {
	case s.ownedHouse(f) != nil && s.ownedHouse(m) == nil:
		m.HouseID = f.HouseID
	case s.ownedHouse(m) != nil && s.ownedHouse(f) == nil:
		f.HouseID = m.HouseID
	}
}

// moveIn makes c the head of house st; a spouse without a home of their own follows.
func (s *Sim) moveIn(c *Creature, st *Structure) {
	c.HouseID = st.ID
	if sp := s.livingSpouse(c); sp != nil && s.ownedHouse(sp) == nil {
		sp.HouseID = st.ID
	}
}

// claim takes over an abandoned house.
func (s *Sim) claim(c *Creature, st *Structure) {
	st.Owner, st.OwnerName, st.Hue = c.ID, c.Name, c.Genome.Traits.Hue
	s.moveIn(c, st)
	s.structVersion++
	s.event("family", fmt.Sprintf("%s menempati %s yang tak bertuan", c.Name, st.kind.Name), c.ID)
}

// inherit passes the dead creature's house to its spouse, else its eldest
// living child in the household; with no heir the house is abandoned.
func (s *Sim) inherit(dead *Creature) {
	for _, st := range s.structures {
		if st.Owner != dead.ID {
			continue
		}
		s.structVersion++
		if !st.house() {
			// Fields, granaries, pens and stations stay in the family: the
			// spouse, else the eldest child, takes them over.
			heir := s.livingSpouse(dead)
			if heir == nil {
				for _, c := range s.creatures {
					if c.Health > 0 && isParent(dead, c) && (heir == nil || c.BornTick < heir.BornTick) {
						heir = c
					}
				}
			}
			st.Owner, st.OwnerName = 0, ""
			if heir != nil {
				st.Owner, st.OwnerName = heir.ID, heir.Name
			}
			continue
		}
		var heir *Creature
		if sp := s.livingSpouse(dead); sp != nil && sp.HouseID == st.ID {
			heir = sp
		} else {
			for _, c := range s.creatures {
				if c.HouseID == st.ID && isParent(dead, c) && (heir == nil || c.BornTick < heir.BornTick) {
					heir = c
				}
			}
		}
		if heir != nil {
			st.Owner, st.OwnerName, st.Hue = heir.ID, heir.Name, heir.Genome.Traits.Hue
			s.event("family", fmt.Sprintf("Rumah keluarga %s diwarisi oleh %s", dead.Name, heir.Name), heir.ID)
			continue
		}
		st.Owner, st.OwnerName = 0, ""
		for _, c := range s.creatures {
			if c.HouseID == st.ID {
				c.HouseID = 0
			}
		}
		s.event("family", fmt.Sprintf("Rumah keluarga %s kini kosong", dead.Name), 0)
	}
}

// leaveBelongings hands a dead creature's things to its killer, or to its home.
func (s *Sim) leaveBelongings(c *Creature) {
	if c.Health <= 0 && c.Offender != nil {
		if killer := s.living(c.Offender.ID); killer != nil {
			s.loot(killer, c.Inventory)
		}
	}
	if s.atHome(c) {
		h := s.houseOf(c)
		for _, id := range c.Inventory.ids() {
			h.Storage.add(id, c.Inventory.take(id, h.kind.Storage-h.Storage.count()))
		}
	}
	c.Inventory = nil
}

func (s *Sim) loot(to *Creature, from Stock) {
	for _, id := range s.byValue(from) {
		room := invCapacity - to.Inventory.count()
		if room <= 0 {
			return
		}
		to.Inventory.add(id, from.take(id, room))
	}
}

// byValue orders a stock's items from most to least valuable.
func (s *Sim) byValue(st Stock) []chem.ItemID {
	ids := st.ids()
	for i := 1; i < len(ids); i++ {
		for j := i; j > 0 && s.cat.item(ids[j]).Value > s.cat.item(ids[j-1]).Value; j-- {
			ids[j], ids[j-1] = ids[j-1], ids[j]
		}
	}
	return ids
}

func (s *Sim) crimeEvent(text string, id int64, violent bool) {
	if s.time()-s.lastCrimeEvent >= crimeEventGap || s.crimes == 1 {
		s.lastCrimeEvent = s.time()
		s.addEvent("crime", text, id, violent)
	}
}

// give hands one unit of food to the hungriest creature in reach, family
// first. Kindness here is feeding someone who needs it.
func (s *Sim) give(c *Creature) bool {
	item := s.cat.edible(c.Inventory, s.handKeep(c))
	if item == "" {
		return false
	}
	var target *Creature
	best := -1.0
	s.grid.near(c.X, c.Y, giveRange, func(o *Creature) {
		d := math.Hypot(o.X-c.X, o.Y-c.Y)
		needy := o.Energy < 0.75 || !s.adult(o)
		if o == c || o.Health <= 0 || d > giveRange || o.Inventory.count() >= invCapacity || !needy {
			return
		}
		score := 1 - d/giveRange + b2f(s.kin(c, o)) + b2f(o.Energy < 0.5)
		if score > best {
			target, best = o, score
		}
	})
	if target == nil {
		// Nobody to share with: feed an animal instead.
		return s.feedAnimal(c, item)
	}
	target.Inventory.add(item, c.Inventory.take(item, 1))
	c.Deeds.Kindness++
	s.kindness++
	s.stats.current(s).Kindness++
	c.Reputation = math.Min(1, c.Reputation+0.05)
	s.flash(c, fxGive)
	c.action = ActGive
	if s.time()-s.lastKindEvent >= kindEventGap {
		s.lastKindEvent = s.time()
		s.event("kindness", fmt.Sprintf("%s berbagi %s dengan %s", c.Name, s.cat.itemName(item), target.Name), c.ID)
	}
	return true
}

// steal takes the most valuable things from a stranger in reach, or from the
// storage of another family's house.
func (s *Sim) steal(c *Creature) bool {
	room := invCapacity - c.Inventory.count()
	if room <= 0 {
		return false
	}
	var victim *Creature
	bestD := stealRange
	s.grid.near(c.X, c.Y, stealRange, func(o *Creature) {
		d := math.Hypot(o.X-c.X, o.Y-c.Y)
		if o != c && o.Health > 0 && d <= bestD && len(o.Inventory) > 0 && !s.kin(c, o) {
			victim, bestD = o, d
		}
	})

	var from Stock
	var text string
	switch {
	case victim != nil:
		from = victim.Inventory
		victim.Hurt = hurtSeconds
		victim.Offender = &Ref{c.ID, c.Name}
		text = "dari " + victim.Name
	default:
		for _, st := range s.structures {
			if (st.house() || st.kind.Granary) && st.Owner != 0 && len(st.Storage) > 0 && st.dist(c.X, c.Y) <= houseStealRange && !s.kinHouse(c, st) {
				from = st.Storage
				text = "dari " + lowerName(st.kind.Name) + " keluarga " + st.OwnerName
				break
			}
		}
		if from == nil {
			from, text = s.stealHarvest(c, room)
		}
	}
	if from == nil {
		return false
	}
	id := s.byValue(from)[0]
	n := from.take(id, min(2, room))
	c.Inventory.add(id, n)
	c.Deeds.Crimes++
	s.crimes++
	s.stats.current(s).Crimes++
	c.Reputation = math.Max(-1, c.Reputation-0.1)
	s.flash(c, fxSteal)
	c.action = ActSteal
	s.crimeEvent(fmt.Sprintf("%s mencuri %d %s %s", c.Name, n, s.cat.itemName(id), text), c.ID, false)
	return true
}

// attack strikes the nearest creature in reach. Whoever drops to zero health dies.
func (s *Sim) attack(c *Creature) bool {
	var target *Creature
	bestD := attackRange
	s.grid.near(c.X, c.Y, attackRange, func(o *Creature) {
		if d := math.Hypot(o.X-c.X, o.Y-c.Y); o != c && o.Health > 0 && d <= bestD {
			target, bestD = o, d
		}
	})
	if target == nil {
		return false
	}
	// Blows in quick succession on the same victim are one assault.
	newAssault := target.Hurt <= 0 || target.Offender == nil || target.Offender.ID != c.ID
	target.Health -= attackDamage * c.Genome.Traits.Size * (1 + s.weaponBonus(c))
	target.Hurt = hurtSeconds
	target.Offender = &Ref{c.ID, c.Name}
	c.Energy -= attackCost
	if newAssault {
		c.Deeds.Crimes++
		s.crimes++
		s.stats.current(s).Crimes++
		c.Reputation = math.Max(-1, c.Reputation-0.15)
	}
	s.flash(c, fxAttack)
	c.action = ActAttack
	if target.Health > 0 {
		s.crimeEvent(fmt.Sprintf("%s menyerang %s", c.Name, target.Name), c.ID, true)
		return true
	}
	c.Deeds.Kills++
	s.kills++
	c.Reputation = math.Max(-1, c.Reputation-0.5)
	s.loot(c, target.Inventory)
	return true
}
