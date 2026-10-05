package sim

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/world"
)

const (
	invCapacity       = 20  // item units a creature can carry
	gatherRate        = 0.5 // units per second without tools
	stationReach      = 3.0 // tiles to a station for crafting and experiments
	homeReach         = 2.0 // tiles to one's house to use its storage
	wellReach         = 1.5
	houseSpacing      = 3.0
	localStationRange = 10.0 // a family builds its own copy of a station if none is this close
	amenityRange      = 4.0  // farms and wells count as "near the house" within this
	experimentSeconds = 4.0
	synthesisSeconds  = 8.0
	foodReserve       = 3 // food units kept on hand when storing at home
	farmFood          = 1.0
	farmRegrow        = 0.02
)

// Stock is a bag of items.
type Stock map[chem.ItemID]int

func (st Stock) count() int {
	n := 0
	for _, q := range st {
		n += q
	}
	return n
}

func (st *Stock) add(id chem.ItemID, n int) {
	if n <= 0 {
		return
	}
	if *st == nil {
		*st = Stock{}
	}
	(*st)[id] += n
}

// take removes up to n units and reports how many were removed.
func (st Stock) take(id chem.ItemID, n int) int {
	have := st[id]
	n = min(n, have)
	if n == have {
		delete(st, id)
	} else {
		st[id] -= n
	}
	return n
}

// ids returns the item ids in a stable order.
func (st Stock) ids() []chem.ItemID {
	ids := make([]chem.ItemID, 0, len(st))
	for id := range st {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}

// Structure is a building on a tile. Creatures can walk over it.
type Structure struct {
	ID        int64   `json:"id"`
	Kind      string  `json:"kind"`
	X         int     `json:"x"`
	Y         int     `json:"y"`
	Owner     int64   `json:"owner,omitempty"` // kepala keluarga or builder; 0 = abandoned
	OwnerName string  `json:"ownerName,omitempty"`
	Hue       float64 `json:"hue"`
	Storage   Stock   `json:"storage,omitempty"`
	BuiltAt   float64 `json:"builtAt"`

	kind chem.StructureKind
}

func (st *Structure) house() bool { return st.kind.House }

func (st *Structure) dist(x, y float64) float64 {
	return math.Hypot(float64(st.X)+0.5-x, float64(st.Y)+0.5-y)
}

// Job is ongoing work that finishes after Remaining seconds.
type Job struct {
	Kind      string      `json:"kind"` // craft | experiment | synthesis | build | upgrade
	Recipe    string      `json:"recipe,omitempty"`
	Item      chem.ItemID `json:"item,omitempty"`
	Structure string      `json:"structure,omitempty"`
	Target    int64       `json:"target,omitempty"` // house being upgraded
	X         int         `json:"x,omitempty"`
	Y         int         `json:"y,omitempty"`
	Remaining float64     `json:"remaining"`
}

func (j *Job) building() bool { return j.Kind == "build" || j.Kind == "upgrade" }

func (j *Job) output() int {
	if j.building() {
		return outBuild
	}
	return outCraft
}

func (j *Job) fx() int {
	if j.building() {
		return fxBuild
	}
	return fxCraft
}

func (j *Job) action() Action {
	if j.building() {
		return ActBuild
	}
	return ActCraft
}

// noGeology stands in when chem provides none.
type noGeology struct{}

func (noGeology) Sources(int, int) []chem.Source              { return nil }
func (noGeology) Take(int, int, chem.ItemID, float64) float64 { return 0 }
func (noGeology) Regrow(float64)                              {}
func (noGeology) Update(*world.Map)                           {}
func (noGeology) Amounts() []float32                          { return nil }
func (noGeology) SetAmounts([]float32) error                  { return nil }

// --- knowledge --------------------------------------------------------------

func (s *Sim) known(symbol string) bool { return s.elements[symbol] != nil }

func (s *Sim) techKnown(id string) bool { return id == "" || s.techs[id] != nil }

func (s *Sim) discover(c *Creature, symbol, source string) {
	if s.known(symbol) {
		return
	}
	s.elements[symbol] = &Discovery{Time: s.time(), By: Ref{c.ID, c.Name}, Era: s.era, Source: source}
	c.Deeds.Discoveries++
	name := symbol
	if e, ok := s.cat.element[symbol]; ok {
		name = e.Name
	}
	if source == synthesisSource {
		s.event("discovery", fmt.Sprintf("Unsur sintetis dibuat: %s (%s) oleh %s", name, symbol, c.Name), c.ID)
	} else {
		s.event("discovery", fmt.Sprintf("Unsur baru ditemukan: %s (%s) oleh %s dari %s", name, symbol, c.Name, source), c.ID)
	}
}

const synthesisSource = "Sintesis"

func (s *Sim) learn(c *Creature, tech string) {
	if s.techKnown(tech) {
		return
	}
	s.techs[tech] = &Discovery{Time: s.time(), By: Ref{c.ID, c.Name}, Era: s.era}
	c.Deeds.Discoveries++
	name := tech
	for _, t := range s.cat.techs {
		if t.ID == tech {
			name = t.Name
		}
	}
	s.event("discovery", fmt.Sprintf("Teknologi baru: %s — %s", name, c.Name), c.ID)
}

// --- what is worth gathering -------------------------------------------------

// computeInterest rates each item by rarity (the share of walkable tiles from
// which it cannot be gathered) times how useful it is early on.
func (s *Sim) computeInterest() {
	t := s.terrain
	counts := map[chem.ItemID]int{}
	for _, i := range t.walkable {
		seen := map[chem.ItemID]bool{}
		for _, src := range s.geo.Sources(int(i)%t.w, int(i)/t.w) {
			if !seen[src.Item] {
				seen[src.Item] = true
				counts[src.Item]++
			}
		}
	}
	s.interest = map[chem.ItemID]float64{}
	n := float64(max(1, len(t.walkable)))
	for id, k := range counts {
		s.interest[id] = clamp(1-float64(k)/n, 0.05, 1) * s.cat.useful(id)
	}
}

const (
	tileSlots   = 4
	tileToolBit = 0x80
)

// refreshResources rebuilds the per-tile table of what can still be gathered
// there, most interesting first, for the "sumber daya" senses.
func (s *Sim) refreshResources() {
	t := s.terrain
	if len(s.tileItems) != len(t.blocked)*tileSlots {
		s.tileItems = make([]uint8, len(t.blocked)*tileSlots)
	}
	clear(s.tileItems)
	for _, i := range t.walkable {
		srcs := s.geo.Sources(int(i)%t.w, int(i)/t.w)
		slices.SortStableFunc(srcs, func(a, b chem.Source) int {
			return cmp.Compare(s.interest[b.Item], s.interest[a.Item])
		})
		n := 0
		for _, src := range srcs {
			idx := s.cat.itemIndex[src.Item]
			if n == tileSlots || idx == 0 || src.Amount < 1 || s.interest[src.Item] < 0.1 {
				continue
			}
			if src.NeedsTool {
				idx |= tileToolBit
			}
			s.tileItems[int(i)*tileSlots+n] = idx
			n++
		}
	}
}

// resourceAt is how appealing tile i is to c: what can be gathered there
// (mostly what it wants, a little what is rare and useful), a station when
// it carries something to analyse, and home when its hands are full.
func (s *Sim) resourceAt(c *Creature, i int, hasTool bool) float64 {
	if st := s.structAt[i]; st != nil {
		if c.sample && st.kind.Tier > 0 || st.ID == c.HouseID && c.Inventory.count() >= invCapacity/2 {
			return 1
		}
	}
	best := float32(0)
	for _, idx := range s.tileItems[i*tileSlots : (i+1)*tileSlots] {
		if idx == 0 {
			break
		}
		if idx&tileToolBit != 0 {
			if !hasTool {
				continue
			}
			idx &^= tileToolBit
		}
		best = max(best, c.wantVec[idx])
	}
	return float64(best)
}

// updateWantVec turns the creature's wants into the per-item appeal its
// senses use.
func (s *Sim) updateWantVec(c *Creature) {
	if len(c.wantVec) != len(s.cat.itemIDs) {
		c.wantVec = make([]float32, len(s.cat.itemIDs))
	}
	want := s.wantsOf(c)
	for idx, id := range s.cat.itemIDs {
		if idx > 0 {
			c.wantVec[idx] = float32(clamp(0.5*want[id]+0.3*s.interest[id], 0, 1))
		}
	}
	c.sample = false
	for _, id := range s.samples() {
		c.sample = c.sample || c.Inventory[id] > 0
	}
}

// --- inventory and storage ---------------------------------------------------

func (s *Sim) houseOf(c *Creature) *Structure {
	if c.HouseID == 0 {
		return nil
	}
	return s.structByID[c.HouseID]
}

func (s *Sim) atHome(c *Creature) bool {
	h := s.houseOf(c)
	return h != nil && h.dist(c.X, c.Y) <= homeReach
}

// available counts an item in hand plus, at home, in the house's storage.
func (s *Sim) available(c *Creature, id chem.ItemID) int {
	n := c.Inventory[id]
	if s.atHome(c) {
		n += s.houseOf(c).Storage[id]
	}
	return n
}

// holding counts an item in hand and in the family's storage, wherever they are.
func (s *Sim) holding(c *Creature, id chem.ItemID) int {
	n := c.Inventory[id]
	if h := s.houseOf(c); h != nil {
		n += h.Storage[id]
	}
	return n
}

func (s *Sim) affordable(c *Creature, cost map[chem.ItemID]int) bool {
	for id, n := range cost {
		if s.available(c, id) < n {
			return false
		}
	}
	return true
}

// consume takes a cost from hand first, then from home storage.
func (s *Sim) consume(c *Creature, cost map[chem.ItemID]int) bool {
	return s.consumeFrom(s.funds(c, false), cost)
}

// funds lists where a creature can take materials from: its hands, its home
// storage when there, and for building (gotong royong) the hands of household
// members who are at the house too.
func (s *Sim) funds(c *Creature, building bool) []Stock {
	out := []Stock{c.Inventory}
	if !s.atHome(c) {
		return out
	}
	h := s.houseOf(c)
	out = append(out, h.Storage)
	if building {
		for _, o := range s.creatures {
			if o != c && o.HouseID == h.ID && o.Health > 0 && h.dist(o.X, o.Y) <= homeReach {
				out = append(out, o.Inventory)
			}
		}
	}
	return out
}

func fundsHave(funds []Stock, id chem.ItemID) int {
	n := 0
	for _, st := range funds {
		n += st[id]
	}
	return n
}

func (s *Sim) consumeFrom(funds []Stock, cost map[chem.ItemID]int) bool {
	for id, n := range cost {
		if fundsHave(funds, id) < n {
			return false
		}
	}
	for id, n := range cost {
		for _, st := range funds {
			if n <= 0 {
				break
			}
			n -= st.take(id, n)
		}
	}
	return true
}

func (s *Sim) canBuildWith(c *Creature, cost map[chem.ItemID]int) bool {
	funds := s.funds(c, true)
	for id, n := range cost {
		if fundsHave(funds, id) < n {
			return false
		}
	}
	return true
}

// receive puts items in hand, overflowing into home storage when at home.
func (s *Sim) receive(c *Creature, id chem.ItemID, n int) {
	room := invCapacity - c.Inventory.count()
	inHand := min(n, room)
	c.Inventory.add(id, inHand)
	if rest := n - inHand; rest > 0 && s.atHome(c) {
		h := s.houseOf(c)
		h.Storage.add(id, min(rest, h.kind.Storage-h.Storage.count()))
	}
}

func (s *Sim) toolBonus(c *Creature) float64 {
	best := 0.0
	for id := range c.Inventory {
		best = math.Max(best, s.cat.item(id).Gather)
	}
	return best
}

func (s *Sim) weaponBonus(c *Creature) float64 {
	best := 0.0
	for id := range c.Inventory {
		best = math.Max(best, s.cat.item(id).Damage)
	}
	return best
}

// keepHouse is the automatic part of home life: leave surplus in storage and
// take food from it when hungry.
func (s *Sim) keepHouse(c *Creature) {
	h := s.houseOf(c)
	tool, weapon := s.bestTool(c), s.bestWeapon(c)
	for _, id := range c.Inventory.ids() {
		keep := 0
		switch id {
		case chem.Food:
			keep = foodReserve
		case tool, weapon:
			keep = 1
		}
		if surplus := c.Inventory[id] - keep; surplus > 0 {
			n := min(surplus, h.kind.Storage-h.Storage.count())
			h.Storage.add(id, c.Inventory.take(id, n))
		}
	}
	if c.Energy < 0.5 && c.Inventory[chem.Food] == 0 {
		c.Inventory.add(chem.Food, h.Storage.take(chem.Food, 2))
	}
}

func (s *Sim) bestTool(c *Creature) chem.ItemID {
	return s.bestBy(c, func(it chem.Item) float64 { return it.Gather })
}

func (s *Sim) bestWeapon(c *Creature) chem.ItemID {
	return s.bestBy(c, func(it chem.Item) float64 { return it.Damage })
}

func (s *Sim) bestBy(c *Creature, score func(chem.Item) float64) chem.ItemID {
	var best chem.ItemID
	bestScore := 0.0
	for _, id := range c.Inventory.ids() {
		if v := score(s.cat.item(id)); v > bestScore {
			best, bestScore = id, v
		}
	}
	return best
}

// --- gathering -------------------------------------------------------------------

func (s *Sim) gather(c *Creature) bool {
	x, y := int(math.Floor(c.X)), int(math.Floor(c.Y))
	if _, ok := s.terrain.index(x, y); !ok {
		return false
	}
	foodCost := float32(s.cat.foodEnergy() / foodValue)
	// Decide what to gather at the start of each unit.
	if c.Gathering == 0 || c.GatherPick == nil && !c.GatherFood {
		c.GatherPick, c.GatherFood = s.chooseGather(c, x, y, foodCost)
		if c.GatherPick == nil && !c.GatherFood {
			return false
		}
	}
	if c.Inventory.count() >= invCapacity && !s.makeRoom(c) {
		c.GatherPick, c.GatherFood = nil, false
		return false
	}

	c.Gathering += gatherRate * math.Max(1, s.toolBonus(c)) * dt
	s.flash(c, fxGather)
	c.action = ActGather
	if c.Gathering < 1 {
		return true
	}
	c.Gathering = 0
	pick, food := c.GatherPick, c.GatherFood
	c.GatherPick, c.GatherFood = nil, false
	if food {
		if s.takeFoodAround(x, y, foodCost) {
			c.Inventory.add(chem.Food, 1)
		}
		return true
	}
	if s.geo.Take(pick.X, pick.Y, pick.Item, 1) < 1 {
		return true
	}
	c.Inventory.add(pick.Item, 1)
	if s.cat.discoverable != nil {
		for _, sym := range s.cat.discoverable(pick.Item, 0, s.known) {
			s.discover(c, sym, s.cat.itemName(pick.Item))
		}
	}
	return true
}

// chooseGather picks the most wanted thing within reach, or food from the
// ground when that is what is needed.
func (s *Sim) chooseGather(c *Creature, x, y int, foodCost float32) (*chem.Source, bool) {
	wants := s.wantsOf(c)
	tool := s.toolBonus(c)
	bestScore, food := 0.2, false
	if s.foodAround(x, y) >= foodCost && c.Inventory[chem.Food] < 4 {
		bestScore, food = 0.6+wants[chem.Food]*3, true
	}
	var pick *chem.Source
	sources := s.geo.Sources(x, y)
	for i := range sources {
		src := &sources[i]
		if src.Amount < 1 || src.NeedsTool && tool == 0 {
			continue
		}
		score := wants[src.Item]*3 + 0.5*s.interest[src.Item] - 0.3*float64(s.holding(c, src.Item))
		if score > bestScore {
			pick, bestScore, food = src, score, false
		}
	}
	return pick, food
}

// foodAround is the food on the tile and its 8 neighbours, which can be
// picked to carry.
func (s *Sim) foodAround(x, y int) float32 {
	t := s.terrain
	var sum float32
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if i, ok := t.index(x+dx, y+dy); ok {
				sum += t.food[i]
			}
		}
	}
	return sum
}

func (s *Sim) takeFoodAround(x, y int, amount float32) bool {
	if s.foodAround(x, y) < amount {
		return false
	}
	t := s.terrain
	for dy := -1; dy <= 1 && amount > 0; dy++ {
		for dx := -1; dx <= 1 && amount > 0; dx++ {
			if i, ok := t.index(x+dx, y+dy); ok {
				take := min(t.food[i], amount)
				t.food[i] -= take
				amount -= take
			}
		}
	}
	return true
}

// makeRoom drops one unit of the least useful thing the creature doesn't
// want, so a full pair of hands can still pick up what it is after.
func (s *Sim) makeRoom(c *Creature) bool {
	if !c.GatherFood && c.GatherPick != nil && s.wantsOf(c)[c.GatherPick.Item] == 0 {
		return false
	}
	want := s.wantsOf(c)
	var drop chem.ItemID
	worst := math.Inf(1)
	for _, id := range c.Inventory.ids() {
		it := s.cat.item(id)
		if id == chem.Food || want[id] > 0 || it.Gather > 0 || it.Damage > 0 {
			continue
		}
		if v := s.interest[id] + it.Value; v < worst {
			drop, worst = id, v
		}
	}
	if drop == "" {
		return false
	}
	c.Inventory.take(drop, 1)
	return true
}

func (s *Sim) wantsOf(c *Creature) map[chem.ItemID]float64 {
	if c.Want == nil {
		c.Want = s.wants(c)
	}
	return c.Want
}

// wants weighs items the creature is short of for its next goals: food, the
// building it would put up next, recipes that teach something new, a tool,
// and samples that could reveal new elements at the stations that exist.
func (s *Sim) wants(c *Creature) map[chem.ItemID]float64 {
	w := map[chem.ItemID]float64{}
	if c.Inventory[chem.Food] < 2 {
		w[chem.Food] = 0.5
	}
	if k, ok := s.buildGoal(c); ok {
		for id, n := range k.Cost {
			if short := n - s.holding(c, id); short > 0 {
				w[id] += 1
				// One step back: inputs of recipes that make what is missing.
				for _, r := range s.cat.recipes {
					if r.Outputs[id] > 0 && s.techKnown(r.Tech) {
						for in := range r.Inputs {
							w[in] += 0.6
						}
					}
				}
			}
		}
	}
	for _, r := range s.cat.recipes {
		if r.Teaches != "" && !s.techKnown(r.Teaches) && s.techKnown(r.Tech) {
			for in := range r.Inputs {
				w[in] += 1
			}
		}
	}
	// Without a pick minerals can't be mined at all, so the first tool matters.
	if s.toolBonus(c) == 0 {
		for _, r := range s.cat.recipes {
			for out := range r.Outputs {
				if s.cat.item(out).Gather > 0 && s.techKnown(r.Tech) && r.Station == "" {
					for in := range r.Inputs {
						w[in] += 1
					}
				}
			}
		}
	}
	for _, id := range s.samples() {
		w[id] += 0.7
	}
	return w
}

// samples lists gatherable items that could reveal a new element at the
// stations the world has. Cached until the tier or the known elements change.
func (s *Sim) samples() []chem.ItemID {
	key := [2]int{s.tier, len(s.elements)}
	if s.sampleKey == key && s.sampleItems != nil {
		return s.sampleItems
	}
	s.sampleKey = key
	s.sampleItems = []chem.ItemID{}
	if s.tier >= 1 && s.cat.discoverable != nil {
		for id := range s.interest {
			if len(s.cat.discoverable(id, s.tier, s.known)) > 0 {
				s.sampleItems = append(s.sampleItems, id)
			}
		}
		slices.Sort(s.sampleItems)
	}
	return s.sampleItems
}

// updateAbilities refreshes the "bisa membuat/membangun" senses.
func (s *Sim) updateAbilities(c *Creature) {
	c.Want = s.wants(c)
	s.updateWantVec(c)
	adult := s.adult(c)
	c.CanCraft = adult && s.chooseCraft(c) != nil
	c.CanBuild = adult && s.chooseBuild(c) != nil
}

// --- crafting ----------------------------------------------------------------------

// stationsNear lists station kinds within reach and the highest station tier.
func (s *Sim) stationsNear(x, y float64) (map[string]bool, int) {
	kinds := map[string]bool{}
	tier := 0
	for _, st := range s.structures {
		if st.kind.Tier > 0 && st.dist(x, y) <= stationReach {
			kinds[st.Kind] = true
			tier = max(tier, st.kind.Tier)
		}
	}
	return kinds, tier
}

func (s *Sim) recipeByID(id string) (chem.Recipe, bool) {
	for _, r := range s.cat.recipes {
		if r.ID == id {
			return r, true
		}
	}
	return chem.Recipe{}, false
}

func (s *Sim) recipeReady(c *Creature, r chem.Recipe, stations map[string]bool) bool {
	return s.techKnown(r.Tech) && (r.Station == "" || stations[r.Station]) && s.affordable(c, r.Inputs)
}

// heldItems lists items in hand and, at home, in storage.
func (s *Sim) heldItems(c *Creature) []chem.ItemID {
	ids := c.Inventory.ids()
	if s.atHome(c) {
		for _, id := range s.houseOf(c).Storage.ids() {
			if c.Inventory[id] == 0 {
				ids = append(ids, id)
			}
		}
		slices.Sort(ids)
	}
	return ids
}

// chooseCraft picks the most useful thing to make here: an experiment that
// reveals an element, then recipes that teach or discover something, then
// what the next building needs, then a missing tool or weapon.
func (s *Sim) chooseCraft(c *Creature) *Job {
	stations, tier := s.stationsNear(c.X, c.Y)
	if tier >= 1 && s.cat.discoverable != nil {
		for _, id := range s.heldItems(c) {
			if len(s.cat.discoverable(id, tier, s.known)) > 0 {
				return &Job{Kind: "experiment", Item: id, Remaining: experimentSeconds}
			}
		}
		if s.cat.synthesizable != nil && s.available(c, chem.SynthesisFuel) > 0 &&
			len(s.cat.synthesizable(tier, s.known)) > 0 {
			return &Job{Kind: "synthesis", Remaining: synthesisSeconds}
		}
	}

	wants := s.wantsOf(c)
	goal, _ := s.buildGoal(c)
	var best *chem.Recipe
	bestScore := 0.0
	for i := range s.cat.recipes {
		r := &s.cat.recipes[i]
		if !s.recipeReady(c, *r, stations) {
			continue
		}
		score := 0.0
		if r.Teaches != "" && !s.techKnown(r.Teaches) {
			score += 5
		}
		if slices.ContainsFunc(r.Discovers, func(sym string) bool { return !s.known(sym) }) {
			score += 5
		}
		for out := range r.Outputs {
			it := s.cat.item(out)
			have := s.holding(c, out)
			// Only make what is still short; a stockpile is enough.
			if have < max(2, goal.Cost[out]) {
				score += 3 * wants[out]
			}
			switch {
			case it.Gather > 0 && s.toolBonus(c) < it.Gather:
				score += 2
			case it.Damage > 0 && s.weaponBonus(c) < it.Damage:
				score += 1
			case have < 2:
				score += 0.3
			}
		}
		// Don't use up what the next building needs, unless the recipe is for
		// it or makes the first tool.
		firstTool := s.toolBonus(c) == 0 && slices.ContainsFunc(sortedKeys(r.Outputs), func(id chem.ItemID) bool {
			return s.cat.item(id).Gather > 0
		})
		for in, n := range r.Inputs {
			if need := goal.Cost[in]; need > 0 && !firstTool && goal.Cost[sortedKeys(r.Outputs)[0]] == 0 && s.holding(c, in)-n < need {
				score -= 2
			}
		}
		if score > bestScore {
			best, bestScore = r, score
		}
	}
	if best == nil || bestScore < 0.5 {
		return nil
	}
	return &Job{Kind: "craft", Recipe: best.ID, Remaining: math.Max(1, best.Seconds)}
}

func (s *Sim) startCraft(c *Creature) bool {
	j := s.chooseCraft(c)
	if j == nil {
		return false
	}
	c.Job = j
	s.flash(c, fxCraft)
	c.action = ActCraft
	return true
}

func (s *Sim) finishJob(c *Creature, j *Job) {
	switch j.Kind {
	case "experiment":
		_, tier := s.stationsNear(c.X, c.Y)
		found := s.cat.discoverable(j.Item, tier, s.known)
		if len(found) == 0 || !s.consume(c, map[chem.ItemID]int{j.Item: 1}) {
			return
		}
		s.discover(c, found[0], s.cat.itemName(j.Item))
	case "synthesis":
		_, tier := s.stationsNear(c.X, c.Y)
		made := s.cat.synthesizable(tier, s.known)
		if len(made) == 0 || !s.consume(c, map[chem.ItemID]int{chem.SynthesisFuel: 1}) {
			return
		}
		s.discover(c, made[0], synthesisSource)
	case "craft":
		r, ok := s.recipeByID(j.Recipe)
		stations, _ := s.stationsNear(c.X, c.Y)
		if !ok || !s.recipeReady(c, r, stations) || !s.consume(c, r.Inputs) {
			return
		}
		for _, id := range sortedKeys(r.Outputs) {
			s.receive(c, id, r.Outputs[id])
		}
		c.Deeds.Crafted++
		s.learn(c, r.Teaches)
		for _, sym := range r.Discovers {
			src := r.Name
			if len(r.Inputs) > 0 {
				src = s.cat.itemName(sortedKeys(r.Inputs)[0])
			}
			s.discover(c, sym, src)
		}
	case "build", "upgrade":
		s.finishBuild(c, j)
	}
}

func sortedKeys(m map[chem.ItemID]int) []chem.ItemID {
	ids := make([]chem.ItemID, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}

// --- building ----------------------------------------------------------------------

func costUnits(cost map[chem.ItemID]int) int {
	n := 0
	for _, q := range cost {
		n += q
	}
	return n
}

func buildSeconds(k chem.StructureKind) float64 {
	return clamp(float64(costUnits(k.Cost))*0.4, 4, 30)
}

func (s *Sim) houseCount() int {
	n := 0
	for _, st := range s.structures {
		if st.house() && st.Owner != 0 {
			n++
		}
	}
	return n
}

func (s *Sim) structureNear(x, y, r float64, pred func(*Structure) bool) bool {
	for _, st := range s.structures {
		if st.dist(x, y) <= r && pred(st) {
			return true
		}
	}
	return false
}

// firstHouse is the cheapest house that needs no existing home.
func (s *Sim) firstHouse() (chem.StructureKind, bool) {
	var best chem.StructureKind
	found := false
	for _, k := range s.cat.structures {
		if k.House && k.Upgrades == "" && s.techKnown(k.Tech) && (!found || costUnits(k.Cost) < costUnits(best.Cost)) {
			best, found = k, true
		}
	}
	return best, found
}

func (s *Sim) upgradeFor(h *Structure) (chem.StructureKind, bool) {
	for _, k := range s.cat.structures {
		if k.House && k.Upgrades == h.Kind && s.techKnown(k.Tech) {
			return k, true
		}
	}
	return chem.StructureKind{}, false
}

func (s *Sim) ownedHouse(c *Creature) *Structure {
	if h := s.houseOf(c); h != nil && h.Owner == c.ID {
		return h
	}
	return nil
}

// buildGoal is the building the creature works towards, affordable or not.
func (s *Sim) buildGoal(c *Creature) (chem.StructureKind, bool) {
	if c.HouseID == 0 {
		return s.firstHouse()
	}
	if k, ok := s.nextStation(); ok {
		return k, true
	}
	if h := s.ownedHouse(c); h != nil {
		if k, ok := s.upgradeFor(h); ok {
			return k, true
		}
	}
	if h := s.houseOf(c); h != nil {
		for _, k := range s.cat.structures {
			if (k.Farm || k.Well) && s.techKnown(k.Tech) && !s.amenityNear(h, k) {
				return k, true
			}
		}
	}
	return chem.StructureKind{}, false
}

// nextStation is the lowest station tier the world doesn't have yet, if its tech is known.
func (s *Sim) nextStation() (chem.StructureKind, bool) {
	var best chem.StructureKind
	found := false
	for _, k := range s.cat.structures {
		if k.Tier > s.tier && s.techKnown(k.Tech) && (!found || k.Tier < best.Tier) {
			best, found = k, true
		}
	}
	return best, found
}

func (s *Sim) amenityNear(h *Structure, k chem.StructureKind) bool {
	x, y := float64(h.X)+0.5, float64(h.Y)+0.5
	if k.Well {
		if i, ok := s.terrain.index(h.X, h.Y); ok && s.terrain.nearWater[i] {
			return true
		}
	}
	return s.structureNear(x, y, amenityRange, func(st *Structure) bool { return st.Kind == k.ID })
}

// buildPlan is what a creature would do on deciding to build right now.
type buildPlan struct {
	claim *Structure
	job   *Job
}

func (s *Sim) tileFree(x, y int) bool {
	i, ok := s.terrain.index(x, y)
	return ok && !s.terrain.blocked[i] && s.structAt[i] == nil
}

func (s *Sim) canPlaceHouse(x, y int) bool {
	if !s.tileFree(x, y) {
		return false
	}
	fx, fy := float64(x)+0.5, float64(y)+0.5
	return !s.structureNear(fx, fy, houseSpacing, func(st *Structure) bool { return st.house() })
}

// spotNear finds a free tile closest to (cx, cy) within r tiles.
func (s *Sim) spotNear(cx, cy, r int) (int, int, bool) {
	for d := 1; d <= r; d++ {
		for dy := -d; dy <= d; dy++ {
			for dx := -d; dx <= d; dx++ {
				if max(abs(dx), abs(dy)) == d && s.tileFree(cx+dx, cy+dy) {
					return cx + dx, cy + dy, true
				}
			}
		}
	}
	return 0, 0, false
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func (s *Sim) chooseBuild(c *Creature) *buildPlan {
	x, y := int(math.Floor(c.X)), int(math.Floor(c.Y))
	if c.HouseID == 0 {
		for _, st := range s.structures {
			if st.house() && st.Owner == 0 && st.dist(c.X, c.Y) <= homeReach {
				return &buildPlan{claim: st}
			}
		}
		if k, ok := s.firstHouse(); ok && s.canBuildWith(c, k.Cost) && s.canPlaceHouse(x, y) {
			return &buildPlan{job: &Job{Kind: "build", Structure: k.ID, X: x, Y: y, Remaining: buildSeconds(k)}}
		}
		return nil
	}

	home := s.houseOf(c)
	if h := s.ownedHouse(c); h != nil && s.atHome(c) {
		if k, ok := s.upgradeFor(h); ok && s.canBuildWith(c, k.Cost) {
			return &buildPlan{job: &Job{Kind: "upgrade", Structure: k.ID, Target: h.ID, X: h.X, Y: h.Y, Remaining: buildSeconds(k)}}
		}
	}
	cx, cy := x, y
	if home != nil && s.atHome(c) {
		cx, cy = home.X, home.Y
	}
	place := func(k chem.StructureKind) *buildPlan {
		if px, py, ok := s.spotNear(cx, cy, 3); ok {
			return &buildPlan{job: &Job{Kind: "build", Structure: k.ID, X: px, Y: py, Remaining: buildSeconds(k)}}
		}
		return nil
	}

	// The most advanced station the world lacks.
	var station *chem.StructureKind
	for i := range s.cat.structures {
		k := &s.cat.structures[i]
		if k.Tier > s.tier && s.techKnown(k.Tech) && s.canBuildWith(c, k.Cost) && (station == nil || k.Tier > station.Tier) {
			station = k
		}
	}
	if station != nil {
		return place(*station)
	}
	if home != nil {
		for _, k := range s.cat.structures {
			if (k.Farm || k.Well) && s.techKnown(k.Tech) && s.canBuildWith(c, k.Cost) && !s.amenityNear(home, k) {
				return place(k)
			}
		}
	}
	// A family far from existing stations builds its own.
	for _, k := range s.cat.structures {
		if k.Tier > 0 && k.Tier <= s.tier && s.techKnown(k.Tech) && s.canBuildWith(c, k.Cost) &&
			!s.structureNear(c.X, c.Y, localStationRange, func(st *Structure) bool { return st.Kind == k.ID }) {
			return place(k)
		}
	}
	return nil
}

func (s *Sim) startBuild(c *Creature) bool {
	p := s.chooseBuild(c)
	switch {
	case p == nil:
		return false
	case p.claim != nil:
		s.claim(c, p.claim)
		return false
	}
	c.Job = p.job
	s.flash(c, fxBuild)
	c.action = ActBuild
	return true
}

func (s *Sim) finishBuild(c *Creature, j *Job) {
	k, ok := s.cat.structure[j.Structure]
	if !ok || !s.techKnown(k.Tech) {
		return
	}
	if j.Kind == "upgrade" {
		h := s.structByID[j.Target]
		if h == nil || h.Owner != c.ID || !s.consumeFrom(s.funds(c, true), k.Cost) {
			return
		}
		h.Kind, h.kind = k.ID, k
		s.structVersion++
		c.Deeds.Built++
		s.learn(c, k.Teaches)
		s.event("build", fmt.Sprintf("%s memperbaiki rumahnya menjadi %s", c.Name, k.Name), c.ID)
		return
	}
	free := s.tileFree(j.X, j.Y)
	if k.House {
		free = c.HouseID == 0 && s.canPlaceHouse(j.X, j.Y)
	}
	if !free || !s.consumeFrom(s.funds(c, true), k.Cost) {
		return
	}
	st := s.addStructure(k, j.X, j.Y, c)
	c.Deeds.Built++
	s.learn(c, k.Teaches)
	if k.House {
		s.moveIn(c, st)
	}
	before := s.tier
	s.recomputeTier()
	if s.tier > before {
		s.event("build", fmt.Sprintf("%s dimulai! %s membangun %s", s.cat.tierLabel(s.tier), c.Name, k.Name), c.ID)
	} else {
		s.event("build", fmt.Sprintf("%s membangun %s", c.Name, k.Name), c.ID)
	}
}

func (s *Sim) addStructure(k chem.StructureKind, x, y int, owner *Creature) *Structure {
	st := &Structure{ID: s.nextStructID, Kind: k.ID, X: x, Y: y, BuiltAt: s.time(), kind: k}
	if owner != nil {
		st.Owner, st.OwnerName, st.Hue = owner.ID, owner.Name, owner.Genome.Traits.Hue
	}
	s.nextStructID++
	s.indexStructure(st)
	if k.Farm {
		s.applyFarms()
	}
	return st
}

func (s *Sim) indexStructure(st *Structure) {
	s.structures = append(s.structures, st)
	s.structByID[st.ID] = st
	if i, ok := s.terrain.index(st.X, st.Y); ok {
		s.structAt[i] = st
	}
	s.structVersion++
}

func (s *Sim) recomputeTier() {
	s.tier = 0
	for _, st := range s.structures {
		s.tier = max(s.tier, st.kind.Tier)
	}
}

// removeBlockedStructures drops buildings whose tile a map edit made solid.
func (s *Sim) removeBlockedStructures() {
	kept := s.structures[:0]
	for _, st := range s.structures {
		if i, ok := s.terrain.index(st.X, st.Y); ok && !s.terrain.blocked[i] {
			kept = append(kept, st)
			continue
		}
		delete(s.structByID, st.ID)
		for _, c := range s.creatures {
			if c.HouseID == st.ID {
				c.HouseID = 0
			}
		}
		s.structVersion++
	}
	clear(s.structures[len(kept):])
	s.structures = kept
	clear(s.structAt)
	for _, st := range s.structures {
		if i, ok := s.terrain.index(st.X, st.Y); ok {
			s.structAt[i] = st
		}
	}
	s.recomputeTier()
}

// applyFarms makes the tiles around every farm fertile.
func (s *Sim) applyFarms() {
	t := s.terrain
	for _, st := range s.structures {
		if !st.kind.Farm {
			continue
		}
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if i, ok := t.index(st.X+dx, st.Y+dy); ok && !t.blocked[i] {
					t.foodCap[i] = max(t.foodCap[i], farmFood)
					t.foodRate[i] = max(t.foodRate[i], farmRegrow)
				}
			}
		}
	}
}
