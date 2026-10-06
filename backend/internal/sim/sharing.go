package sim

import (
	"fmt"
	"math"

	"miniv2/backend/internal/chem"
)

// Food sharing, storage and moving on (Fase 4, berbagi pangan). Hunters,
// gatherers and early farmers outlive bad seasons less by luck than by
// pooling: big game is shared out across the camp, a hungry person takes
// what a better-fed relative or a neighbour with more than they can use
// does not defend, harvests are carried home and kept for the lean months,
// and when the land around a camp is eaten bare people know where it is
// better and move. Diagnosis on the live world (PLAN.md, Hasil Fase 3b and
// "Berbagi pangan") found the opposite: most deaths were from hunger, beside
// people who carried food and fields that rotted unharvested, with the
// stores empty and 85% of adults crowded on the riverbanks while the land a
// few tiles inland went uneaten.
//
// None of this hands anyone food or sends anyone anywhere. It changes the
// rules of the world:
//   - meat a hunter can't carry stays where the animal fell, for anyone who
//     comes to eat or cut a share (carcasses);
//   - someone who wants to eat and has nothing may eat from the food of
//     family within arm's reach who are better fed, or from a neighbour's
//     load that is more than they can use (demand sharing, tolerated theft);
//   - anyone may help harvest a household's ripe field, keeping a share, the
//     rest going to that household's store (bawon);
//   - villagers may draw on their village's granary when hungry;
//   - people know where the land within their home range feeds more mouths
//     than where they stand (one new piece of knowledge, in the food memory).
// What they do with any of it is up to their brains.

const (
	// meatItem is what a kill yields.
	meatItem chem.ItemID = "daging"

	// carcassKeeps is the half-life (years) of meat left in the open. Cut,
	// carried and cooked meat keeps half a year here (chem: daging); a
	// carcass lying out in the tropics goes about twice as fast to flies,
	// heat and scavengers (dogs, monitor lizards, crows, tigers), which
	// take an unattended kill within days (Blumenschine 1986, Serengeti:
	// carcass completeness falls by half within a day or two).
	carcassKeeps = 0.25
	// maxCarcasses bounds the list; a full one gives up the least meat.
	maxCarcasses = 64

	// shareHungry is the energy below which a person asks for food, as
	// the "lapar" of a child of the family (childHungry).
	shareHungry = childHungry
	// bigPackage is how many meals in hand are more than one can eat
	// before they spoil: five units are some two years of a grown man's
	// food here, and more than half of a load of meat rots (half-life half
	// a year) before it can be eaten. Beyond this, by Blurton Jones's
	// (1984) argument, the last units are worth less to the holder than a
	// quarrel with the hungry, and a neighbour's share is not defended.
	bigPackage = 5
	// scroungeTrust is the gratitude a fed person feels for whoever fed
	// them (a gift is worth 0.18 to its witnesses; see give).
	scroungeTrust = 0.05

	// bawonShare is the harvester's share of a household's harvest. In
	// the Javanese bawon, harvest was open to all of the village (and
	// often outsiders), each harvester taking home a share of what they
	// cut: a sixth to a tenth before the 1970s (Collier et al. 1974;
	// Hayami & Kikuchi 1981); the rest went to the household's barn.
	bawonShare = 1.0 / 6
	// bawonScore ranks joining a harvest among gathering choices: above
	// wild food (0.6) and below meat on the ground and one's own harvest.
	bawonScore   = 0.62
	carcassScore = 0.8

	// granaryHungry: below this energy, with nothing in hand, a villager
	// may take from the village granary. Communal reserve granaries for
	// the lean season and the needy are an old institution in the
	// archipelago: the Minangkabau rangkiang sitangguang lapa ("bearer of
	// hunger"), the Sundanese leuit, the Javanese lumbung desa, which lent
	// paddy to villagers in the paceklik (Hüsken & White 1989).
	granaryHungry = 0.5
	granaryMeals  = 2

	// Home-range knowledge: the land is known in landCell-sized blocks,
	// within landReach blocks of where one stands (about 24 tiles). Foragers
	// know their range far beyond sight, and talk at camp about where food
	// is (Lee 1979; Marlowe 2010).
	landCell   = 8
	landReach  = 3
	landHungry = 0.75 // energy below which one thinks of where to go next
	landBetter = 1.5  // food per head must be this many times better than here
	// landWater: only land within this many steps of a river or lake that
	// holds water now counts. A forager's range is the land around water
	// (camps sit by it and the day's walk returns to it; Lee 1979, Marlowe
	// 2010); here a body lasts some 30 seconds dry and walks 1.5–2.5 tiles
	// a second, so eight tiles out and back leaves a wide margin. Without
	// this bound the hungry learned of the rich interior and died of
	// thirst on the way (live copy: thirst deaths 27 → 61 in 60 s).
	landWater = 8
	// landRange: the places one knows lie within this many tiles of the
	// water one drinks at, the centre of one's range; the walk there and
	// back is then one's own well-trodden way (see senseRememberedWater).
	landRange = 1.5 * landWater

	saltBawon   = 0x6261776f6e
	saltCarcass = 0x6361726361
)

// Carcass is meat left where an animal was killed: whoever is near may eat
// it or cut a piece to carry.
type Carcass struct {
	ID      int64   `json:"id"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Meat    int     `json:"meat"`
	Species string  `json:"species,omitempty"`
	Tick    int64   `json:"tick"` // when it was left
}

// SharingStats counts what was shared, all-time.
type SharingStats struct {
	MeatLeft     int `json:"meatLeft"`            // units left at kills
	MeatTaken    int `json:"meatTaken"`           // … eaten or cut from carcasses
	MeatRotted   int `json:"meatRotted"`          // … lost to rot and scavengers
	FromKin      int `json:"fromKin"`             // meals eaten from family's food
	FromOthers   int `json:"fromOthers"`          // … from a neighbour's surplus
	Bawon        int `json:"bawon"`               // units harvested into another household's store
	BawonShare   int `json:"bawonShare"`          // … and kept by the harvesters
	GranaryMeals int `json:"granaryMeals"`        // meals drawn from a village granary
	Adopted      int `json:"adopted,omitempty"`   // ownerless granaries taken over by a household
	LandKnown    int `json:"landKnown,omitempty"` // times someone learned of better land
}

// SharingInfo is sharing for the world info: the all-time counts, and the
// meat lying out and the food people hold now.
type SharingInfo struct {
	SharingStats
	Carcasses   int `json:"carcasses"`
	CarcassMeat int `json:"carcassMeat"`
	Granaries   int `json:"granaries"`
	GranaryFood int `json:"granaryFood"`
	HouseFood   int `json:"houseFood"`
	CarriedFood int `json:"carriedFood"`
}

// sharing is a world's sharing state; its exported fields are saved.
type sharing struct {
	Carcasses []*Carcass   `json:"carcasses,omitempty"`
	Next      int64        `json:"next,omitempty"`
	Stats     SharingStats `json:"stats"`

	tiles []int32     // per tile, 1 + the index of a carcass on it (0 none)
	land  []landPatch // home-range knowledge, rebuilt each time it is used
	// Scratch for wetLand: the tiles within landWater steps of open fresh
	// water, their distance, and the search queue.
	wet      []bool
	wetDist  []uint8
	wetQueue []int32
}

// landPatch is one block of land as people know it.
type landPatch struct {
	food   float32 // wild food on it
	best   float32 // on its richest tile …
	tile   int32   // … which is this one
	people int32
}

// --- carcasses ----------------------------------------------------------------

// leaveCarcass leaves n units of meat at (x, y), joining a carcass already
// on that tile.
func (s *Sim) leaveCarcass(x, y float64, n int, species string) {
	sh := &s.share
	if s.opts.NoSharing || n <= 0 {
		return
	}
	i, ok := s.terrain.indexAt(x, y)
	if !ok {
		return
	}
	sh.Stats.MeatLeft += n
	if k := s.carcassOn(i); k != nil {
		k.Meat += n
		k.Tick = s.tick
		return
	}
	if len(sh.Carcasses) >= maxCarcasses {
		// The least meat goes first: what is left of it the scavengers take.
		least := 0
		for j, k := range sh.Carcasses {
			if k.Meat < sh.Carcasses[least].Meat {
				least = j
			}
		}
		sh.Stats.MeatRotted += sh.Carcasses[least].Meat
		sh.Carcasses[least].Meat = 0
		s.dropEmptyCarcasses()
	}
	sh.Next++
	sh.Carcasses = append(sh.Carcasses, &Carcass{ID: sh.Next, X: x, Y: y, Meat: n, Species: species, Tick: s.tick})
	s.indexCarcasses()
	s.addStimulus(stimCarcass, x, y, 1, nil)
}

// indexCarcasses rebuilds the per-tile lookup.
func (s *Sim) indexCarcasses() {
	sh := &s.share
	if len(sh.tiles) != len(s.terrain.blocked) {
		sh.tiles = make([]int32, len(s.terrain.blocked))
	} else {
		clear(sh.tiles)
	}
	for j, k := range sh.Carcasses {
		if i, ok := s.terrain.indexAt(k.X, k.Y); ok {
			sh.tiles[i] = int32(j + 1)
		}
	}
}

// dropEmptyCarcasses removes carcasses with no meat left.
func (s *Sim) dropEmptyCarcasses() {
	sh := &s.share
	kept := sh.Carcasses[:0]
	for _, k := range sh.Carcasses {
		if k.Meat > 0 {
			kept = append(kept, k)
		}
	}
	clear(sh.Carcasses[len(kept):])
	sh.Carcasses = kept
	s.indexCarcasses()
}

// carcassOn is the carcass on tile i, if any.
func (s *Sim) carcassOn(i int) *Carcass {
	sh := &s.share
	if len(sh.Carcasses) == 0 || i < 0 || i >= len(sh.tiles) || sh.tiles[i] == 0 {
		return nil
	}
	return sh.Carcasses[sh.tiles[i]-1]
}

// meatFood is how much food a carcass on tile i is to the senses: like a
// ripe field, a carcass with a meal or two on it is as much food as there
// is (1). 0 without one; cheap enough for every ray step.
func (s *Sim) meatFood(i int) float64 {
	if len(s.share.Carcasses) == 0 {
		return 0
	}
	if k := s.carcassOn(i); k != nil {
		return math.Min(1, float64(k.Meat)*s.cat.item(meatItem).Food/foodValue)
	}
	return 0
}

// carcassNear is a carcass on tile (x, y) or next to it: within arm's reach,
// as wild food is (foodAround).
func (s *Sim) carcassNear(x, y int) (*Carcass, int) {
	if len(s.share.Carcasses) == 0 {
		return nil, 0
	}
	for _, k := range [9]int{4, 0, 1, 2, 3, 5, 6, 7, 8} { // here first
		if i, ok := s.terrain.index(x+k%3-1, y+k/3-1); ok {
			if c := s.carcassOn(i); c != nil && c.Meat > 0 {
				return c, i
			}
		}
	}
	return nil, 0
}

// meatHere is the sense of food in reach from a carcass (see meatFood).
func (s *Sim) meatHere(c *Creature) float64 {
	if len(s.share.Carcasses) == 0 {
		return 0
	}
	if k, i := s.carcassNear(int(math.Floor(c.X)), int(math.Floor(c.Y))); k != nil {
		return s.meatFood(i)
	}
	return 0
}

// takeMeat removes one unit from carcass k, dropping it when bare.
func (s *Sim) takeMeat(k *Carcass) {
	k.Meat--
	s.share.Stats.MeatTaken++
	if k.Meat <= 0 {
		s.dropEmptyCarcasses()
	}
}

// eatCarcass lets c, wanting to eat, eat a unit of meat lying in reach.
// Fresh meat on the ground goes before carried food that keeps.
func (s *Sim) eatCarcass(c *Creature) bool {
	k, _ := s.carcassNear(int(math.Floor(c.X)), int(math.Floor(c.Y)))
	if k == nil || s.opts.NoSharing {
		return false
	}
	f := s.cat.item(meatItem).Food
	c.Energy = math.Min(1, c.Energy+f)
	s.ateFood(meatItem, f)
	c.ate = true
	s.takeMeat(k)
	return true
}

// cutMeat is a finished "gather" at a carcass: a unit of meat to carry.
func (s *Sim) cutMeat(c *Creature, tile int) {
	k := s.carcassOn(tile)
	if k == nil || k.Meat <= 0 || invCapacity-c.Inventory.count() <= 0 {
		return
	}
	c.Inventory.add(meatItem, 1)
	s.takeMeat(k)
}

// rotCarcasses lets a second of heat and scavengers at the meat lying out.
func (s *Sim) rotCarcasses() {
	sh := &s.share
	if len(sh.Carcasses) == 0 {
		return
	}
	p := 1 - math.Exp2(-1/(carcassKeeps*SecondsPerYear))
	gone := false
	for _, k := range sh.Carcasses {
		n := min(k.Meat, int(float64(k.Meat)*p+hash01(k.ID, s.tick, saltCarcass^s.eco.Seed())))
		k.Meat -= n
		sh.Stats.MeatRotted += n
		gone = gone || k.Meat <= 0
	}
	if gone {
		s.dropEmptyCarcasses()
	}
}

// leftOver is what happens to n units of id that c could neither carry
// nor store: meat stays where it is for whoever comes (before sharing,
// it vanished).
func (s *Sim) leftOver(c *Creature, id chem.ItemID, n int) {
	if n > 0 && id == meatItem {
		s.leaveCarcass(c.X, c.Y, n, "")
	}
}

// --- eating from others' food ---------------------------------------------------

// spareFor is how many of holder o's meals c, hungry, may eat: family's
// food if o is better fed and keeps a meal, a neighbour's beyond what o
// can use (bigPackage). Seed kept for sowing is never touched.
func (s *Sim) spareFor(c, o *Creature) (int, bool) {
	meals := s.mealsInHand(o)
	if s.kin(c, o) {
		if o.Energy > c.Energy && meals > 1 {
			return meals - 1, true
		}
		return 0, true
	}
	return max(0, meals-bigPackage), false
}

// scrounge lets c, who wants to eat and has nothing in hand or in reach,
// eat a unit from the food of someone within arm's reach who can spare it:
// family first, then whoever has the most to spare. Among foragers most
// food changes hands this way, asked for rather than offered (demand
// sharing; Peterson 1993), and among farmers the household eats from one
// pot. The holder doesn't decide; it is the norm (and the cost of refusing)
// that lets the food go.
func (s *Sim) scrounge(c *Creature) bool {
	if s.opts.NoSharing || c.Energy >= shareHungry {
		return false
	}
	var from *Creature
	bestSpare, bestKin := 0, false
	s.grid.near(c.X, c.Y, giveRange, func(o *Creature) {
		if o == c || len(o.Inventory) == 0 || o.Health <= 0 || math.Hypot(o.X-c.X, o.Y-c.Y) > giveRange {
			return
		}
		spare, kin := s.spareFor(c, o)
		if spare <= 0 || from != nil && (bestKin && !kin || bestKin == kin && spare <= bestSpare) {
			return
		}
		from, bestSpare, bestKin = o, spare, kin
	})
	if from == nil {
		return false
	}
	id := s.cat.edible(from.Inventory, s.handKeep(from))
	if id == "" {
		return false
	}
	from.Inventory.take(id, 1)
	f := s.cat.item(id).Food
	c.Energy = math.Min(1, c.Energy+f)
	s.ateFood(id, f)
	c.ate = true
	s.rememberRelation(c, from, scroungeTrust)
	if bestKin {
		s.share.Stats.FromKin++
	} else {
		s.share.Stats.FromOthers++
	}
	if s.time()-s.lastEcoEvt["share"] >= kindEventGap {
		s.lastEcoEvt["share"] = s.time()
		text := fmt.Sprintf("%s ikut makan dari bekal %s, keluarganya", c.Name, from.Name)
		if !bestKin {
			text = fmt.Sprintf("%s ikut makan dari bekal %s yang berlimpah", c.Name, from.Name)
		}
		s.event("kindness", text, c.ID)
	}
	return true
}

// --- bawon: an open harvest ---------------------------------------------------------

// bawonStore is where the harvest of household houseID's field goes when
// others cut it: its granary, else its house, while either has room.
func (s *Sim) bawonStore(houseID int64) *Structure {
	h := s.structByID[houseID]
	if h == nil || !h.house() || h.Owner == 0 {
		return nil
	}
	if g := s.granaryOf(h); g != nil && g.Storage.count() < g.kind.Storage {
		return g
	}
	if h.Storage.count() < h.kind.Storage {
		return h
	}
	return nil
}

// bawonOpen reports whether household houseID has a store with room for a
// harvest: bawonStore, asking the cheap question (the house) first.
func (s *Sim) bawonOpen(houseID int64) bool {
	h := s.structByID[houseID]
	if h == nil || !h.house() || h.Owner == 0 {
		return false
	}
	return h.Storage.count() < h.kind.Storage || s.bawonStore(houseID) != nil
}

// bawonInReach finds a ripe field on c's tile or next to it that c may not
// harvest for itself but may help harvest: another household's, with a
// store to carry the harvest to.
func (s *Sim) bawonInReach(c *Creature, x, y int) (int, bool) {
	if s.opts.NoSharing {
		return 0, false
	}
	for k := range 9 {
		i, ok := s.terrain.index(x+k%3-1, y+k/3-1)
		if !ok {
			continue
		}
		if p := s.eco.PlotAt(i); p != nil && p.IsRipe() && p.House != 0 && !s.mayHarvest(c, p) && s.bawonOpen(p.House) {
			return i, true
		}
	}
	return 0, false
}

// bawonHarvest is a finished unit of harvest work on another household's
// field: what is cut goes to that household's store, but for the
// harvester's share (bawonShare on average, a whole unit at a time).
func (s *Sim) bawonHarvest(c *Creature, tile int) {
	p := s.eco.PlotAt(tile)
	if p == nil || !p.IsRipe() || s.mayHarvest(c, p) {
		return
	}
	store := s.bawonStore(p.House)
	if store == nil {
		return
	}
	room := store.kind.Storage - store.Storage.count()
	hands := invCapacity - c.Inventory.count()
	id, n := s.eco.Harvest(tile, min(harvestUnits, room+min(1, hands)), s.time())
	if n == 0 {
		return
	}
	mine := 0
	if hands > 0 && hash01(c.ID, s.tick, saltBawon^s.eco.Seed()) < float64(n)*bawonShare {
		mine = 1
	}
	mine = max(mine, n-room) // what the store can't take stays with the harvester
	c.Inventory.add(id, mine)
	store.Storage.add(id, n-mine)
	c.Deeds.Harvested += n
	s.share.Stats.Bawon += n - mine
	s.share.Stats.BawonShare += mine
	s.learn(c, chem.FarmingTech)
	if s.time()-s.lastEcoEvt["bawon"] >= farmEventGap {
		s.lastEcoEvt["bawon"] = s.time()
		owner := "orang lain"
		if h := s.structByID[p.House]; h != nil && h.OwnerName != "" {
			owner = "keluarga " + h.OwnerName
		}
		s.event("farming", fmt.Sprintf("%s ikut memanen %s di ladang %s dan membawa pulang bawon", c.Name, s.cat.itemName(id), owner), c.ID)
	}
}

// shareGather offers the gathering choices sharing adds (see chooseGather):
// cutting meat from a carcass, or joining another household's harvest,
// when c is short of food.
func (s *Sim) shareGather(c *Creature, x, y int, hungry bool, wants map[chem.ItemID]float64) (string, float64, int) {
	if s.opts.NoSharing || !hungry {
		return "", 0, 0
	}
	if k, i := s.carcassNear(x, y); k != nil {
		return "carcass", carcassScore + wants[chem.Food]*3, i
	}
	if i, ok := s.bawonInReach(c, x, y); ok {
		return "bawon", bawonScore + wants[chem.Food]*3, i
	}
	return "", 0, 0
}

// --- the village granary --------------------------------------------------------

// drawGranary lets a hungry villager with nothing in hand standing at one
// of the village's granaries take a couple of meals, leaving the seed of
// the family that keeps it.
func (s *Sim) drawGranary(c *Creature) {
	if c.VillageID == 0 || c.Energy >= granaryHungry || s.carrier(c) != nil || s.mealsInHand(c) > 0 {
		return
	}
	v := s.vil().byID[c.VillageID]
	if v == nil {
		return
	}
	for _, id := range v.Buildings {
		st := s.structByID[id]
		if st == nil || !st.kind.Granary || len(st.Storage) == 0 || st.dist(c.X, c.Y) > homeReach {
			continue
		}
		var keep func(chem.ItemID) int
		if o := s.byID[st.Owner]; o != nil {
			keep = s.storeKeep(c, s.houseOf(o))
		}
		took := 0
		for took < granaryMeals && c.Inventory.count() < invCapacity {
			food := s.cat.edible(st.Storage, keep)
			if food == "" {
				break
			}
			c.Inventory.add(food, st.Storage.take(food, 1))
			took++
		}
		if took == 0 {
			continue
		}
		s.share.Stats.GranaryMeals += took
		if s.time()-s.lastEcoEvt["lumbung"] >= farmEventGap {
			s.lastEcoEvt["lumbung"] = s.time()
			s.event("farming", fmt.Sprintf("%s mengambil bekal dari lumbung Desa %s", c.Name, v.Name), c.ID)
		}
		return
	}
}

// adoptGranaries gives a granary whose family died out to the household
// living beside it. A granary's builder may die without spouse or child
// while the house next to it passes to someone else (inherit); the
// granary then stood empty for good, because only its owner's family may
// use it and its presence stops the family next door building their own.
// A barn in a family's yard is theirs to use, as an abandoned house is
// anyone's to claim (claim).
func (s *Sim) adoptGranaries() {
	for _, st := range s.structures {
		if !st.kind.Granary || st.Owner != 0 {
			continue
		}
		var home *Structure
		bestD := amenityRange + 0.01
		for _, h := range s.structures {
			if h.house() && h.Owner != 0 {
				if d := h.dist(float64(st.X)+0.5, float64(st.Y)+0.5); d < bestD {
					home, bestD = h, d
				}
			}
		}
		if home == nil {
			continue
		}
		st.Owner, st.OwnerName, st.Hue = home.Owner, home.OwnerName, home.Hue
		s.structVersion++
		s.share.Stats.Adopted++
		s.event("family", fmt.Sprintf("Keluarga %s memakai %s tak bertuan di samping rumahnya", home.OwnerName, lowerName(st.kind.Name)), home.Owner)
	}
}

// --- knowing the land ---------------------------------------------------------------

// knowLand gives everyone getting hungry who remembers nowhere worth the
// walk the best land of their home range: the block within reach with the
// most wild food per head, if it is clearly better than where they stand.
// It goes into their food memory (forage.go) like a place they had seen,
// so the senses that feel remembered food point them there; whether they
// go is up to them. Crowding enters as the heads: a riverbank eaten bare
// by seventy people is worth less than the same land with five.
func (s *Sim) knowLand() {
	t := s.terrain
	cw, ch := (t.w+landCell-1)/landCell, (t.h+landCell-1)/landCell
	sh := &s.share
	if len(sh.land) != cw*ch {
		sh.land = make([]landPatch, cw*ch)
	}
	clear(sh.land)
	any := false
	for _, c := range s.creatures {
		if c.Energy < landHungry {
			any = true
		}
		if x, y := int(c.X)/landCell, int(c.Y)/landCell; x >= 0 && y >= 0 && x < cw && y < ch {
			sh.land[y*cw+x].people++
		}
	}
	if !any {
		return
	}
	wet := s.wetLand()
	for _, i := range t.walkable {
		if !wet[i] {
			continue
		}
		f := s.eco.Forage(int(i))
		p := &sh.land[(int(i)/t.w/landCell)*cw+int(i)%t.w/landCell]
		p.food += f
		if f > p.best {
			p.best, p.tile = f, i
		}
	}
	perHead := func(p *landPatch) float64 { return float64(p.food) / float64(1+p.people) }
	for _, c := range s.creatures {
		if c.Energy >= landHungry || s.carrier(c) != nil {
			continue
		}
		if _, v := s.bestFoodPlace(c); v >= foodPlaceMin {
			continue
		}
		// One's range is the land around the water one drinks at: someone
		// who doesn't know where to drink has that to find first.
		if c.WaterX == 0 && c.WaterY == 0 {
			continue
		}
		cx, cy := int(c.X)/landCell, int(c.Y)/landCell
		if cx < 0 || cy < 0 || cx >= cw || cy >= ch {
			continue
		}
		here := perHead(&sh.land[cy*cw+cx]) * landBetter
		var best *landPatch
		bestV := 0.0
		for y := max(0, cy-landReach); y <= min(ch-1, cy+landReach); y++ {
			for x := max(0, cx-landReach); x <= min(cw-1, cx+landReach); x++ {
				p := &sh.land[y*cw+x]
				if (x == cx && y == cy) || p.best < foodSeenRich {
					continue
				}
				tx, ty := float64(int(p.tile)%t.w)+0.5, float64(int(p.tile)/t.w)+0.5
				if math.Hypot(tx-c.WaterX, ty-c.WaterY) > landRange {
					continue
				}
				v := perHead(p) / (1 + math.Hypot(tx-c.X, ty-c.Y)/foodMemoryReach)
				if v > here && v > bestV {
					best, bestV = p, v
				}
			}
		}
		if best == nil {
			continue
		}
		s.rememberFood(c, float64(int(best.tile)%t.w)+0.5, float64(int(best.tile)/t.w)+0.5, math.Min(1, float64(best.best)))
		sh.Stats.LandKnown++
	}
}

// wetLand marks the walkable tiles within landWater steps (through
// walkable ground) of open fresh water as it is now: in the dry season a
// stream that has stopped running is no place to forage from.
func (s *Sim) wetLand() []bool {
	sh := &s.share
	t := s.terrain
	if len(sh.wet) != len(t.blocked) {
		sh.wet, sh.wetDist = make([]bool, len(t.blocked)), make([]uint8, len(t.blocked))
	} else {
		clear(sh.wet)
	}
	queue := sh.wetQueue[:0]
	for _, i := range t.riverbank {
		if s.riverRunsBy(int(i)) {
			sh.wet[i], sh.wetDist[i] = true, 0
			queue = append(queue, i)
		}
	}
	for k := 0; k < len(queue); k++ {
		i := queue[k]
		if sh.wetDist[i] >= landWater {
			continue
		}
		x, y := int(i)%t.w, int(i)/t.w
		for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			if j, ok := t.index(x+d[0], y+d[1]); ok && !t.blocked[j] && !sh.wet[j] {
				sh.wet[j], sh.wetDist[j] = true, sh.wetDist[i]+1
				queue = append(queue, int32(j))
			}
		}
	}
	sh.wetQueue = queue
	return sh.wet
}

// --- the world's clock -------------------------------------------------------------

// shareSecond runs once a simulated second: meat rots, villagers draw on
// their granaries, the hungry think of better land.
func (s *Sim) shareSecond() {
	if s.opts.NoSharing {
		if len(s.share.Carcasses) > 0 {
			s.share.Carcasses = nil
			s.indexCarcasses()
		}
		return
	}
	s.rotCarcasses()
	s.adoptGranaries()
	for _, c := range s.creatures {
		if c.Health > 0 {
			s.drawGranary(c)
		}
	}
	s.knowLand()
}

// --- saving and the observer -----------------------------------------------------------

func (s *Sim) saveSharing() *sharing {
	sh := &s.share
	if len(sh.Carcasses) == 0 && sh.Next == 0 && sh.Stats == (SharingStats{}) {
		return nil
	}
	return sh
}

func (s *Sim) restoreSharing(st *sharing) {
	s.share = sharing{}
	if st != nil {
		s.share.Next, s.share.Stats = st.Next, st.Stats
		for _, k := range st.Carcasses {
			if k != nil && k.Meat > 0 {
				s.share.Carcasses = append(s.share.Carcasses, k)
			}
		}
	}
	s.indexCarcasses()
}

// carcassMeat is the meat lying out now.
func (s *Sim) carcassMeat() int {
	n := 0
	for _, k := range s.share.Carcasses {
		n += k.Meat
	}
	return n
}

// sharingInfo summarises sharing and the food people hold, for the world info.
func (s *Sim) sharingInfo() SharingInfo {
	f := s.foodStock()
	info := SharingInfo{SharingStats: s.share.Stats, Carcasses: len(s.share.Carcasses), CarcassMeat: s.carcassMeat(),
		GranaryFood: f.Granary, HouseFood: f.Stored, CarriedFood: f.Carried}
	for _, st := range s.structures {
		if st.kind.Granary {
			info.Granaries++
		}
	}
	return info
}

// villageFood is the food in a village's houses and in its granaries.
func (s *Sim) villageFood(v *village) (houses, granaries int) {
	for _, id := range v.Houses {
		if st := s.structByID[id]; st != nil {
			houses += s.cat.foodUnits(st.Storage)
		}
	}
	for _, id := range v.Buildings {
		if st := s.structByID[id]; st != nil && st.kind.Granary {
			granaries += s.cat.foodUnits(st.Storage)
		}
	}
	return houses, granaries
}
