package sim

import (
	"fmt"
	"math"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/ecology"
)

// Farming (Fase 2): anyone may push a seed, tuber or sucker they carry into
// ground where it can grow; the first harvest of a planted crop invents
// farming, and practice makes better harvests. Fields belong to the family
// that planted them; taking someone else's harvest is theft.

const (
	harvestUnits    = 3   // units taken per completed harvest action
	plantSuits      = 0.3 // how well a crop must suit the ground to be planted
	gardenRange     = 8.0 // a family with a house plants within this many tiles of it
	irrigationWater = 3   // a channel must be this close to a river or lake …
	irrigationReach = 2   // … and waters fields this far around it
	irrigationSite  = 6   // how far from the house a channel may be dug
	snareSite       = 4   // how far from the house a family sets its snare
	manureReach     = 2
	seedKeep        = 4    // crop units a farming household keeps back in store for sowing …
	seedCarry       = 2    // … and in hand, taken along to plant
	seedHunger      = 0.25 // below this energy even the seed is eaten
	spoilHouse      = 1.5  // food keeps this much longer in a house …
	spoilGranary    = 4.0  // … and in a granary
	farmEventGap    = 30.0
)

// mayHarvest reports whether c may take the harvest of plot p: its own
// field or its family's, or one whose family is gone.
func (s *Sim) mayHarvest(c *Creature, p *ecology.Plot) bool {
	if p.Owner == c.ID || p.House != 0 && p.House == c.HouseID {
		return true
	}
	if o := s.living(p.Owner); o != nil {
		return s.kin(c, o)
	}
	if h := s.structByID[p.House]; h != nil && h.Owner != 0 {
		return s.kinHouse(c, h)
	}
	return true // nobody is left to claim it
}

// ripeInReach finds a ripe plot c may harvest on its tile or next to it.
func (s *Sim) ripeInReach(c *Creature, x, y int) (int, bool) {
	for k := range 9 {
		i, ok := s.terrain.index(x+k%3-1, y+k/3-1)
		if !ok {
			continue
		}
		if p := s.eco.PlotAt(i); p != nil && p.IsRipe() && s.mayHarvest(c, p) {
			return i, true
		}
	}
	return 0, false
}

// fishInReach finds the water tile next to (x, y) with the most fish.
func (s *Sim) fishInReach(x, y int) (int, bool) {
	best, bestFish := 0, float32(1)
	found := false
	for k := range 9 {
		i, ok := s.terrain.index(x+k%3-1, y+k/3-1)
		if ok && s.terrain.water[i] && s.eco.Fish(i) >= bestFish {
			best, bestFish, found = i, s.eco.Fish(i), true
		}
	}
	return best, found
}

// harvest takes what c can carry from the ripe plot on tile i.
func (s *Sim) harvest(c *Creature, i int) {
	room := invCapacity - c.Inventory.count()
	id, n := s.eco.Harvest(i, min(harvestUnits, room), s.time())
	if n == 0 {
		return
	}
	c.Inventory.add(id, n)
	c.Deeds.Harvested += n
	s.learn(c, chem.FarmingTech)
	if s.time()-s.lastEcoEvt["harvest"] >= farmEventGap {
		s.lastEcoEvt["harvest"] = s.time()
		s.event("farming", fmt.Sprintf("%s memanen %d %s", c.Name, n, s.cat.itemName(id)), c.ID)
	}
}

// plantChoice is the crop c carries that would grow best where it stands,
// or failing that on a free tile next to it, and that tile, if any crop
// would grow well enough there.
func (s *Sim) plantChoice(c *Creature) (chem.ItemID, int) {
	if s.opts.NoFarming || !s.adult(c) || len(c.Inventory) == 0 {
		return "", 0
	}
	// Settled families keep their gardens by the house, where they will be
	// around to harvest them; the homeless plant wherever they roam.
	if h := s.houseOf(c); h != nil && h.dist(c.X, c.Y) > gardenRange {
		return "", 0
	}
	// Runs every tick for everyone (the "bisa menanam" sense), so it walks
	// the inventory without sorting; ties go to the first id by name and
	// then the first tile, so worlds stay deterministic.
	carries := false
	for id := range c.Inventory {
		carries = carries || s.eco.Plantable(id)
	}
	if !carries {
		return "", 0
	}
	x, y := int(math.Floor(c.X)), int(math.Floor(c.Y))
	for _, k := range [9]int{4, 0, 1, 2, 3, 5, 6, 7, 8} { // here first
		i, ok := s.terrain.index(x+k%3-1, y+k/3-1)
		if !ok || s.structAt[i] != nil {
			continue
		}
		var best chem.ItemID
		bestScore := plantSuits
		for id := range c.Inventory {
			if !s.eco.Plantable(id) {
				continue
			}
			if v := s.eco.SuitsHere(i, id); v > bestScore || v == bestScore && best != "" && id < best {
				best, bestScore = id, v
			}
		}
		if best != "" {
			return best, i
		}
	}
	return "", 0
}

// plant puts the best-suited crop c carries into the best free ground in reach.
func (s *Sim) plant(c *Creature) bool {
	id, tile := s.plantChoice(c)
	if id == "" {
		return false
	}
	skill := c.Skills[chem.FarmingTech]
	if s.noCulture() && s.techKnown(chem.FarmingTech) {
		skill = math.Max(skill, discoverSkill)
	}
	if !s.eco.Plant(tile, id, c.ID, c.HouseID, skill, s.time()) {
		return false
	}
	c.Inventory.take(id, 1)
	c.Deeds.Planted++
	s.flash(c, fxPlant)
	c.action = ActPlant
	if s.time()-s.lastEcoEvt["plant"] >= farmEventGap {
		s.lastEcoEvt["plant"] = s.time()
		s.event("farming", fmt.Sprintf("%s menanam %s", c.Name, s.eco.CropName(id)), c.ID)
	}
	return true
}

// --- seed ------------------------------------------------------------------------

// sows reports whether c keeps seed and sows it: a grown-up who knows how
// to farm. Farmers everywhere hold seed back from the pot; eating it is the
// last resort of a famine.
func (s *Sim) sows(c *Creature) bool {
	return !s.opts.NoFarming && s.adult(c) && s.canPractise(c, chem.FarmingTech)
}

// storeKeep says how many units of each item c leaves uneaten in the
// stores of house h: seedKeep of every crop if c, or the head of the house,
// farms and c is not yet desperate; nil keeps nothing.
func (s *Sim) storeKeep(c *Creature, h *Structure) func(chem.ItemID) int {
	if c.Energy < seedHunger {
		return nil
	}
	if !s.sows(c) {
		var head *Creature
		if h != nil {
			head = s.byID[h.Owner]
		}
		if head == nil || !s.sows(head) {
			return nil
		}
	}
	return func(id chem.ItemID) int {
		if s.eco.Plantable(id) {
			return seedKeep
		}
		return 0
	}
}

// handKeep says how many carried units of each item c leaves uneaten (or
// doesn't give away): the seed a sower carries, until hunger is desperate.
func (s *Sim) handKeep(c *Creature) func(chem.ItemID) int {
	seed := s.seedInHand(c)
	if seed == "" || c.Energy < seedHunger {
		return nil
	}
	return func(id chem.ItemID) int {
		if id == seed {
			return seedCarry
		}
		return 0
	}
}

// seedInHand is the crop a sower carries most of, if any.
func (s *Sim) seedInHand(c *Creature) chem.ItemID {
	if !s.sows(c) {
		return ""
	}
	var best chem.ItemID
	for id, n := range c.Inventory {
		if s.eco.Plantable(id) && (best == "" || n > c.Inventory[best] || n == c.Inventory[best] && id < best) {
			best = id
		}
	}
	return best
}

// mealsInHand counts the carried food c would eat: everything edible but the
// seed a sower carries.
func (s *Sim) mealsInHand(c *Creature) int {
	n := s.cat.foodUnits(c.Inventory)
	if id := s.seedInHand(c); id != "" && c.Energy >= seedHunger {
		n -= min(c.Inventory[id], seedCarry)
	}
	return n
}

// takeSeed lets a sower at home take seed from the family stores: the crop
// they hold most of, a couple of units to put in the ground nearby.
func (s *Sim) takeSeed(c *Creature, stores ...*Structure) {
	var best chem.ItemID
	have := func(id chem.ItemID) int {
		n := 0
		for _, st := range stores {
			if st != nil {
				n += st.Storage[id]
			}
		}
		return n
	}
	for _, st := range stores {
		if st == nil {
			continue
		}
		for id := range st.Storage {
			if s.eco.Plantable(id) && (best == "" || have(id) > have(best) || have(id) == have(best) && id < best) {
				best = id
			}
		}
	}
	if best == "" {
		return
	}
	want := min(seedCarry, invCapacity-c.Inventory.count())
	for _, st := range stores {
		if st != nil && want > 0 {
			n := st.Storage.take(best, want)
			c.Inventory.add(best, n)
			want -= n
		}
	}
}

// stealHarvest takes the ripe harvest of another family's field next to c.
func (s *Sim) stealHarvest(c *Creature, room int) (Stock, string) {
	x, y := int(math.Floor(c.X)), int(math.Floor(c.Y))
	for k := range 9 {
		i, ok := s.terrain.index(x+k%3-1, y+k/3-1)
		if !ok {
			continue
		}
		p := s.eco.PlotAt(i)
		if p == nil || !p.IsRipe() || s.mayHarvest(c, p) {
			continue
		}
		owner := "orang lain"
		if o := s.byID[p.Owner]; o != nil {
			owner = "keluarga " + o.Name
		}
		id, n := s.eco.Harvest(i, min(2, room), s.time())
		if n == 0 {
			continue
		}
		return Stock{id: n}, "panen dari ladang " + owner
	}
	return nil, ""
}

// --- fields, channels, granaries and pens -----------------------------------

func (s *Sim) farmKind() (chem.StructureKind, bool) {
	for _, k := range s.cat.structures {
		if k.Farm {
			return k, true
		}
	}
	return chem.StructureKind{}, false
}

// amenityWanted reports whether c's family, living in h, should build k next
// to the house: a snare, a field once they farm, a well, an irrigation
// channel for a field by a river, a granary for a farming family, a pen for
// their animals.
func (s *Sim) amenityWanted(c *Creature, h *Structure, k chem.StructureKind) bool {
	if !(k.Farm || k.Well || k.Irrigation || k.Granary || k.Pen || k.Snare || k.Latrine) || !s.canPractise(c, k.Tech) || s.amenityNear(h, k) {
		return false
	}
	switch {
	case k.Snare, k.Latrine:
		return true
	case k.Farm, k.Well:
		return k.Farm && !s.opts.NoFarming || k.Well
	case k.Irrigation:
		farm, ok := s.farmKind()
		return ok && s.amenityNear(h, farm) && s.eco.FreshWithin(h.X, h.Y, irrigationSite+irrigationWater)
	case k.Granary:
		farm, ok := s.farmKind()
		return ok && s.amenityNear(h, farm)
	case k.Pen:
		return len(s.eco.Livestock(h.ID)) > 0
	}
	return false
}

// placeIrrigation finds a free tile near the house and near fresh water for
// a channel.
func (s *Sim) placeIrrigation(h *Structure, k chem.StructureKind) *buildPlan {
	bestD := math.Inf(1)
	var plan *buildPlan
	for dy := -irrigationSite; dy <= irrigationSite; dy++ {
		for dx := -irrigationSite; dx <= irrigationSite; dx++ {
			x, y := h.X+dx, h.Y+dy
			if !s.tileFree(x, y) || !s.eco.FreshWithin(x, y, irrigationWater) {
				continue
			}
			if d := math.Hypot(float64(dx), float64(dy)); d < bestD {
				bestD = d
				plan = &buildPlan{job: &Job{Kind: "build", Structure: k.ID, X: x, Y: y, Remaining: buildSeconds(k)}}
			}
		}
	}
	return plan
}

// placeSnare finds a free tile near the house where game comes to feed: the
// most wild food within snareSite tiles, the nearest of equals.
func (s *Sim) placeSnare(h *Structure, k chem.StructureKind) *buildPlan {
	best, bestD := -1.0, math.Inf(1)
	var plan *buildPlan
	for dy := -snareSite; dy <= snareSite; dy++ {
		for dx := -snareSite; dx <= snareSite; dx++ {
			x, y := h.X+dx, h.Y+dy
			i, ok := s.terrain.index(x, y)
			if !ok || !s.tileFree(x, y) {
				continue
			}
			v, d := float64(s.eco.Forage(i)), math.Hypot(float64(dx), float64(dy))
			if d > snareSite || v < best || v == best && d >= bestD {
				continue
			}
			best, bestD = v, d
			plan = &buildPlan{job: &Job{Kind: "build", Structure: k.ID, X: x, Y: y, Remaining: buildSeconds(k)}}
		}
	}
	return plan
}

// snareOf is the family snare near house h, if any.
func (s *Sim) snareOf(h *Structure) *Structure {
	return s.familyAmenity(h, func(k chem.StructureKind) bool { return k.Snare })
}

// granaryOf is the family granary near house h, if any.
func (s *Sim) granaryOf(h *Structure) *Structure {
	return s.familyAmenity(h, func(k chem.StructureKind) bool { return k.Granary })
}

// penOf is the family pen near house h, if any.
func (s *Sim) penOf(h *Structure) *Structure {
	return s.familyAmenity(h, func(k chem.StructureKind) bool { return k.Pen })
}

func (s *Sim) familyAmenity(h *Structure, is func(chem.StructureKind) bool) *Structure {
	if h == nil {
		return nil
	}
	x, y := float64(h.X)+0.5, float64(h.Y)+0.5
	head := s.byID[h.Owner]
	mine := func(st *Structure) bool {
		if st.Owner == h.Owner && h.Owner != 0 {
			return true
		}
		o := s.byID[st.Owner]
		return head != nil && o != nil && s.kin(head, o)
	}
	var best *Structure
	bestD := amenityRange + 0.01
	for _, st := range s.structures {
		if !is(st.kind) || !mine(st) {
			continue
		}
		if d := st.dist(x, y); d < bestD {
			best, bestD = st, d
		}
	}
	return best
}

// applyFarms tells the ecology which tiles are fields (around a ladang),
// irrigated (around a channel by fresh water), manured (around a pen with
// animals in it) or hold a set snare.
func (s *Sim) applyFarms() {
	t := s.terrain
	bits := make([]uint8, len(t.blocked))
	mark := func(cx, cy, r int, bit uint8) {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if i, ok := t.index(cx+dx, cy+dy); ok && !t.blocked[i] {
					bits[i] |= bit
				}
			}
		}
	}
	for _, st := range s.structures {
		switch {
		case st.kind.Farm:
			mark(st.X, st.Y, 1, ecology.Farmland)
		case st.kind.Irrigation:
			if s.eco.FreshWithin(st.X, st.Y, irrigationWater) {
				mark(st.X, st.Y, irrigationReach, ecology.Irrigated)
			}
		case st.kind.Snare:
			if len(st.Storage) == 0 {
				mark(st.X, st.Y, 0, ecology.SnareSet) // set; a sprung one waits to be emptied
			}
		case st.house() && st.Owner != 0:
			if pen := s.penOf(st); pen != nil && len(s.eco.Livestock(st.ID)) > 0 {
				mark(pen.X, pen.Y, manureReach, ecology.Manured)
			}
		}
	}
	s.eco.SetManagement(bits)
}

// --- food going off ------------------------------------------------------------

// ateFood books energy eaten from a carried unit of id under where the food
// came from: a crop, fish, meat or the wild.
func (s *Sim) ateFood(id chem.ItemID, energy float64) {
	y := s.eco.Current()
	switch {
	case s.eco.Plantable(id):
		y.FoodCrops += energy
	case id == "ikan" || id == "ikan_asin":
		y.FoodFish += energy
	case id == "daging" || id == "daging_asap":
		y.FoodMeat += energy
	default:
		y.FoodWild += energy
	}
}

// spoil lets a second pass on all food: each unit goes off with the chance
// its half-life gives, slower in houses and much slower in granaries.
func (s *Sim) spoil() {
	rotten := 0
	for _, c := range s.creatures {
		rotten += s.rot(c.Inventory, 1)
	}
	for _, st := range s.structures {
		switch {
		case len(st.Storage) == 0:
		case st.kind.Granary:
			rotten += s.rot(st.Storage, spoilGranary)
		default:
			rotten += s.rot(st.Storage, spoilHouse)
		}
	}
	if rotten > 0 {
		s.eco.Current().Rotten += rotten
	}
}

func (s *Sim) rot(st Stock, keepsLonger float64) int {
	if len(st) == 0 {
		return 0
	}
	lost := 0
	for _, id := range st.ids() {
		keeps := s.cat.item(id).Keeps
		if keeps <= 0 {
			continue
		}
		p := 1 - math.Exp2(-1/(keeps*SecondsPerYear*keepsLonger))
		n := st[id]
		k := min(n, int(float64(n)*p+s.rng.Float64()))
		lost += st.take(id, k)
	}
	return lost
}

// updateWoodland tells the ecology how much forest is left around each tile.
func (s *Sim) updateWoodland() {
	if cover, cut := s.geo.Woodland(); cover != nil {
		s.eco.SetWoodland(cover, cut)
	}
	s.applyFarms() // livestock may have come or gone
}
