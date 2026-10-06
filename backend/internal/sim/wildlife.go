package sim

import (
	"fmt"
	"math"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/ecology"
)

// Hunting, taming and keeping animals (Fase 2). Hunting is its own act
// ("buru"), apart from attacking people: killing a deer is not a crime.

const (
	huntRange       = 1.3
	spearRange      = 2.2 // a thrown spear reaches further
	huntDamage      = 0.35
	fightBack       = 0.5 // chance a fierce animal turns on its hunter
	livestockHungry = 0.5 // livestock this thin is "lapar"
	feedRange       = 1.5
	cullAbove       = 6 // a family slaughters animals beyond this many
	settledRange    = 3.0
	huntEventGap    = 20.0
)

var speciesInfo = ecology.SpeciesList()

func speciesByID(id string) (ecology.Species, bool) {
	for _, sp := range speciesInfo {
		if sp.ID == id {
			return sp, true
		}
	}
	return ecology.Species{}, false
}

// animalKill words a death by an animal, e.g. "diterkam harimau".
func animalKill(species string) string {
	if sp, ok := speciesByID(species); ok {
		if sp.Kill != "" {
			return sp.Kill
		}
		return "tewas diserang " + lowerName(sp.Name)
	}
	return "tewas diserang hewan"
}

func lowerName(s string) string {
	b := []byte(s)
	for i, ch := range b {
		if ch >= 'A' && ch <= 'Z' {
			b[i] = ch + 'a' - 'A'
		}
	}
	return string(b)
}

// hunt strikes the nearest wild animal in reach. A kill gives meat; a hurt
// boar, buffalo or tiger may turn on the hunter.
func (s *Sim) hunt(c *Creature) bool {
	reach := huntRange
	if s.weaponBonus(c) > 0 {
		reach = spearRange
	}
	var target *ecology.Animal
	bestD := reach
	s.eco.Near(c.X, c.Y, reach, func(a *ecology.Animal) {
		if a.Owner != 0 {
			return
		}
		if d := math.Hypot(a.X-c.X, a.Y-c.Y); d <= bestD {
			target, bestD = a, d
		}
	})
	if target == nil {
		return false
	}
	sp := target.Kind()
	c.Energy -= attackCost
	s.flash(c, fxHunt)
	c.action = ActHunt
	if s.eco.Strike(target, huntDamage*c.Genome.Traits.Size*(1+s.weaponBonus(c))) {
		c.Deeds.Hunted++
		meat := sp.Meat
		room := invCapacity - c.Inventory.count()
		c.Inventory.add("daging", min(meat, room))
		if s.time()-s.lastEcoEvt["hunt"] >= huntEventGap {
			s.lastEcoEvt["hunt"] = s.time()
			s.event("hunt", fmt.Sprintf("%s berburu %s", c.Name, lowerName(sp.Name)), c.ID)
		}
		return true
	}
	if sp.Fierce && s.rng.Float64() < fightBack {
		s.Maul(c.ID, sp.Bite, sp.ID)
	}
	return true
}

// feedAnimal gives a tameable animal or a hungry head of the family's
// livestock a bite, when there is nobody to share with.
func (s *Sim) feedAnimal(c *Creature, item chem.ItemID) bool {
	var target *ecology.Animal
	best := -1.0
	s.eco.Near(c.X, c.Y, feedRange, func(a *ecology.Animal) {
		d := math.Hypot(a.X-c.X, a.Y-c.Y)
		own := a.Owner != 0 && a.Owner == c.HouseID
		if d > feedRange || !(own && a.Energy < 0.7 || a.Owner == 0 && a.Tameable()) {
			return
		}
		if score := 1 - a.Energy - d/feedRange*0.2; score > best {
			target, best = a, score
		}
	})
	if target == nil {
		return false
	}
	c.Inventory.take(item, 1)
	s.eco.Feed(target, s.cat.item(item).Food, c.HouseID, s)
	s.flash(c, fxGive)
	c.action = ActGive
	return true
}

// cullLivestock slaughters an animal when the family keeps more than it can
// use, or when someone is hungry and the stores are bare.
func (s *Sim) cullLivestock(c *Creature, h *Structure) {
	if h == nil || !s.adult(c) {
		return
	}
	herd := s.eco.Livestock(h.ID)
	hungry := c.Energy < 0.4 && s.cat.edible(h.Storage, s.storeKeep(c, h)) == "" && len(herd) >= 2
	if len(herd) <= cullAbove && !hungry {
		return
	}
	// The oldest male goes first; breeding females last.
	var pick *ecology.Animal
	for _, a := range herd {
		switch {
		case pick == nil,
			!a.Female && pick.Female,
			a.Female == pick.Female && a.Born < pick.Born:
			pick = a
		}
	}
	meat := s.eco.Slaughter(pick)
	s.receive(c, "daging", meat)
	if s.time()-s.lastEcoEvt["cull"] >= huntEventGap {
		s.lastEcoEvt["cull"] = s.time()
		s.event("hunt", fmt.Sprintf("Keluarga %s menyembelih seekor %s", h.OwnerName, lowerName(pick.Kind().Tame)), c.ID)
	}
}

// --- what the ecology may ask about people (ecology.Humans) -----------------

// Nearest returns the closest living person within r of (x, y).
func (s *Sim) Nearest(x, y, r float64) (int64, float64, float64, bool) {
	var best *Creature
	bestD := r
	s.grid.near(x, y, r, func(c *Creature) {
		if c.Health <= 0 {
			return
		}
		if d := math.Hypot(c.X-x, c.Y-y); d <= bestD {
			best, bestD = c, d
		}
	})
	if best == nil {
		return 0, 0, 0, false
	}
	return best.ID, best.X, best.Y, true
}

// Crowd counts living people within r of (x, y).
func (s *Sim) Crowd(x, y, r float64) int {
	n := 0
	s.grid.near(x, y, r, func(c *Creature) {
		if c.Health > 0 && math.Hypot(c.X-x, c.Y-y) <= r {
			n++
		}
	})
	return n
}

// Maul hurts a person; if it kills them, the death is the animal's doing.
func (s *Sim) Maul(id int64, damage float64, species string) {
	c := s.living(id)
	if c == nil {
		return
	}
	c.Health -= damage
	c.Hurt = hurtSeconds
	c.Mauled = species
	c.Offender = nil
}

// Home is where household house keeps its livestock: its pen, or the house.
func (s *Sim) Home(house int64) (float64, float64, bool, bool) {
	h := s.structByID[house]
	if h == nil || !h.house() || h.Owner == 0 {
		return 0, 0, false, false
	}
	if pen := s.penOf(h); pen != nil {
		return float64(pen.X) + 0.5, float64(pen.Y) + 0.5, true, true
	}
	return float64(h.X) + 0.5, float64(h.Y) + 0.5, false, true
}

// Settled returns a lived-in house within a few tiles of (x, y).
func (s *Sim) Settled(x, y float64) (int64, bool) {
	for _, st := range s.structures {
		if st.house() && st.Owner != 0 && math.Abs(float64(st.X)+0.5-x) <= settledRange &&
			math.Abs(float64(st.Y)+0.5-y) <= settledRange {
			return st.ID, true
		}
	}
	return 0, false
}

// Tamed: an animal now belongs to a household. Its head learns husbandry
// (the first time anywhere, that is the invention of herding).
func (s *Sim) Tamed(house int64, species string) {
	h := s.structByID[house]
	if h == nil {
		return
	}
	sp, _ := speciesByID(species)
	owner := s.living(h.Owner)
	if owner != nil {
		owner.Deeds.Tamed++
		s.learn(owner, chem.HerdingTech)
	}
	if s.time()-s.lastEcoEvt["tame"] >= huntEventGap {
		s.lastEcoEvt["tame"] = s.time()
		s.event("hunt", fmt.Sprintf("Seekor %s menjadi jinak dan tinggal di rumah keluarga %s", lowerName(sp.Name), h.OwnerName), h.Owner)
	}
}
