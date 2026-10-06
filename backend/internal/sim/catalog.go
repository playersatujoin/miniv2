package sim

import (
	"math"
	"sync"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/world"
)

// geology is what the simulation needs from chem.Geology; tests use fakes.
type geology interface {
	Sources(x, y int) []chem.Source
	Take(x, y int, item chem.ItemID, amount float64) float64
	Regrow(seconds float64)
	Update(m *world.Map)
	Amounts() []float32
	SetAmounts(a []float32) error
	MinedOut() [][2]int
	// Woodland is the tree cover left around each tile (1 = untouched) and
	// the share of all trees cut.
	Woodland() ([]float32, float64)
	// Logged lists trees cut down to stumps.
	Logged() [][2]int
}

// catalog indexes chem's static data for quick lookups. Tests swap in small
// hand-made catalogs so behaviour rules can be checked in isolation.
type catalog struct {
	items      map[chem.ItemID]chem.Item
	recipes    []chem.Recipe
	structures []chem.StructureKind
	structure  map[string]chem.StructureKind
	techs      []chem.Tech
	elements   []chem.Element
	element    map[string]chem.Element

	tierName      func(int) string
	discoverable  func(item chem.ItemID, tier int, known func(string) bool) []string
	synthesizable func(tier int, known func(string) bool) []string
	newGeology    func(m *world.Map) geology

	// usefulness rates how early an item is needed: 1 for the first tiers,
	// less for later ones, a little for things nothing is made from.
	usefulness map[chem.ItemID]float64

	// Small integer ids (1-based) so per-tile tables stay compact.
	itemIndex map[chem.ItemID]uint8
	itemIDs   []chem.ItemID
}

func newCatalog(items []chem.Item, recipes []chem.Recipe, structures []chem.StructureKind,
	techs []chem.Tech, elements []chem.Element) *catalog {
	c := &catalog{
		items:      map[chem.ItemID]chem.Item{},
		recipes:    recipes,
		structures: structures,
		structure:  map[string]chem.StructureKind{},
		techs:      techs,
		elements:   elements,
		element:    map[string]chem.Element{},
	}
	c.itemIndex = map[chem.ItemID]uint8{}
	c.itemIDs = []chem.ItemID{""}
	for _, it := range items {
		c.items[it.ID] = it
		if len(c.itemIDs) < tileToolBit {
			c.itemIndex[it.ID] = uint8(len(c.itemIDs))
			c.itemIDs = append(c.itemIDs, it.ID)
		}
	}
	for _, k := range structures {
		c.structure[k.ID] = k
	}
	for _, e := range elements {
		c.element[e.Symbol] = e
	}
	c.rateUsefulness()
	return c
}

func (k *catalog) rateUsefulness() {
	techTier := map[string]int{}
	for _, t := range k.techs {
		techTier[t.ID] = t.Tier
	}
	weight := func(tier int) float64 {
		switch {
		case tier <= 1:
			return 1
		case tier <= 3:
			return 0.6
		default:
			return 0.3
		}
	}
	k.usefulness = map[chem.ItemID]float64{chem.Food: 1}
	use := func(items map[chem.ItemID]int, tier int) {
		for id := range items {
			k.usefulness[id] = max(k.usefulness[id], weight(tier))
		}
	}
	for _, r := range k.recipes {
		tier := techTier[r.Tech]
		if st, ok := k.structure[r.Station]; ok {
			tier = max(tier, st.Tier)
		}
		use(r.Inputs, tier)
	}
	for _, st := range k.structures {
		use(st.Cost, max(st.Tier, techTier[st.Tech]))
	}
}

func (k *catalog) useful(id chem.ItemID) float64 {
	if u, ok := k.usefulness[id]; ok {
		return u
	}
	return 0.1
}

var (
	chemOnce    sync.Once
	chemCatalog *catalog
)

// defaultCatalog is the real chemistry from package chem.
func defaultCatalog() *catalog {
	chemOnce.Do(func() {
		c := newCatalog(chem.Items(), chem.Recipes(), chem.Structures(), chem.Techs(), chem.Elements())
		c.tierName = chem.TierName
		c.discoverable = chem.Discoverable
		c.synthesizable = chem.Synthesizable
		c.newGeology = func(m *world.Map) geology { return chem.NewGeology(m) }
		chemCatalog = c
	})
	return chemCatalog
}

func (k *catalog) item(id chem.ItemID) chem.Item {
	if it, ok := k.items[id]; ok {
		return it
	}
	return chem.Item{ID: id, Name: string(id)}
}

func (k *catalog) itemName(id chem.ItemID) string { return k.item(id).Name }

// foodEnergy is the energy one carried unit of food gives.
func (k *catalog) foodEnergy() float64 {
	if f := k.item(chem.Food).Food; f > 0 {
		return f
	}
	return 0.3
}

func (k *catalog) totalElements() int {
	if n := len(k.elements); n > 0 {
		return n
	}
	return 118
}

func (k *catalog) tierLabel(tier int) string {
	if k.tierName != nil {
		if n := k.tierName(tier); n != "" {
			return n
		}
	}
	return "Zaman " + string(rune('0'+tier))
}

// edible picks what to eat from st: the food that spoils soonest (a ripe
// banana before dried sago), or "" if there is nothing to eat. keep, if not
// nil, says how many units of an item must stay (seed kept back for sowing).
func (k *catalog) edible(st Stock, keep func(chem.ItemID) int) chem.ItemID {
	var best chem.ItemID
	bestKeeps := math.MaxFloat64
	for _, id := range st.ids() {
		it := k.item(id)
		if it.Food <= 0 || st[id] <= 0 || keep != nil && st[id] <= keep(id) {
			continue
		}
		keeps := it.Keeps
		if keeps <= 0 {
			keeps = math.MaxFloat64 / 2 // never spoils: eat it last
		}
		if keeps < bestKeeps {
			best, bestKeeps = id, keeps
		}
	}
	return best
}

// isFood reports whether an item can be eaten.
func (k *catalog) isFood(id chem.ItemID) bool { return k.item(id).Food > 0 }

// foodUnits counts the edible units in st.
func (k *catalog) foodUnits(st Stock) int {
	n := 0
	for id, q := range st {
		if k.isFood(id) {
			n += q
		}
	}
	return n
}
