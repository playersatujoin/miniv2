package sim

import (
	"fmt"
	"slices"

	"miniv2/backend/internal/chem"
)

// Barter: two people who both want to trade swap what each values more than
// what they give (PLAN.md, Fase 4 design point 4). Value is subjective: what
// an item means to its owner now (hunger, the house or tool they are working
// towards, see wantsOf) and how many they hold, the last unit being worth
// the most. With worth falling as 1/n for the n-th unit held (Bernoulli's
// logarithmic utility), surplus flows to where it is scarce, which is what
// makes barter worth the trouble. Needs are kept smooth on purpose: wantsOf
// stops wanting a building material once enough is in hand, and a value
// that drops to nothing the moment a swap is made would make the same two
// trade it straight back.
//
// One swap per meeting: the single item each hands over, and how many (up
// to tradeBundle each way, so two tubers can buy one axe), that leaves both
// better off by at least tradeMinGain, choosing the swap that maximises the
// product of the two gains (the Nash bargaining solution: neither is left
// with a sliver while the other takes the rest). No swap, no trade.

const (
	// tradeBundle is the most units of one item handed over in one swap.
	tradeBundle = 3
	// tradeMinGain is the least either must gain for a swap to be worth the
	// haggling: a transaction cost, about the worth of a spare stone.
	tradeMinGain = 0.05
	// tradeWorth is how much an item's general worth (chem's Value, 0–10)
	// counts for someone with no use for it: a little, enough to keep.
	tradeWorth = 0.05
	// toolNeed and weaponNeed: what a first gathering tool or weapon is
	// worth to someone of trading age (weapons only to grown-ups, who hunt
	// and fight), against 1.5 for the first meal to the starving.
	toolNeed   = 1.0
	weaponNeed = 0.5
	// foodNeed is what a first meal in hand is worth to the well fed (as
	// wantsOf's wish for food when short of two meals); hunger adds up to 1.
	foodNeed = 0.5
	// goalNeed: each material of the building they are working towards stays
	// wanted while they hold it, as wantsOf wants it while short.
	goalNeed = 1.0
	// tradeTrust: a fair exchange builds trust between the two (Molm et al.
	// 2000), more than a talk, less than a gift of food.
	tradeTrust = 0.05
	tradeJoy   = 0.1
	// tradeEventGap is the seconds between trades written in the event log.
	tradeEventGap = 10.0
	// recentTrades is how many of the latest trades the observer sees.
	recentTrades = 12
)

// TradeRecord is one barter, for the observer: A gave GaveN of Gave to B for
// GotN of Got.
type TradeRecord struct {
	Time  float64     `json:"time"`
	A     Ref         `json:"a"`
	Gave  chem.ItemID `json:"gave"`
	GaveN int         `json:"gaveN"`
	B     Ref         `json:"b"`
	Got   chem.ItemID `json:"got"`
	GotN  int         `json:"gotN"`
}

// valuer is what goods mean to one person at the moment of a trade.
type valuer struct {
	s      *Sim
	c      *Creature
	want   map[chem.ItemID]float64
	goal   map[chem.ItemID]int // what the building they are working towards takes
	hunger float64             // 0 sated … 1 starving
	adult  bool
	keep   func(chem.ItemID) int // units they won't part with (seed for sowing)
	// Units of food, gathering tools and weapons in hand and in the
	// family's stores: one food, tool or weapon stands in for another.
	food, tools, weapons int
}

func (s *Sim) valuer(c *Creature) *valuer {
	v := &valuer{s: s, c: c, want: s.wantsOf(c), hunger: clamp((0.9-c.Energy)/0.6, 0, 1),
		adult: s.adult(c), keep: s.handKeep(c)}
	if k, ok := s.buildGoal(c); ok {
		v.goal = k.Cost
	}
	v.count(c.Inventory)
	if h := s.houseOf(c); h != nil {
		v.count(h.Storage)
	}
	return v
}

func (v *valuer) count(st Stock) {
	for id, n := range st {
		switch it := v.s.cat.item(id); {
		case it.Food > 0:
			v.food += n
		case it.Gather > 0:
			v.tools += n
		case it.Damage > 0:
			v.weapons += n
		}
	}
}

// need is what one unit of id is worth to them if it were the only one they
// had.
func (v *valuer) need(id chem.ItemID) float64 {
	it := v.s.cat.item(id)
	n := tradeWorth * it.Value
	if id != chem.Food {
		w := v.want[id]
		if v.goal[id] > 0 {
			w = max(w, goalNeed)
		}
		n += w
	}
	switch {
	case it.Food > 0:
		// Any food will do: by how much a unit feeds compared with an
		// ordinary meal.
		n += (foodNeed + v.hunger) * it.Food / v.s.cat.foodEnergy()
	case it.Gather > 0:
		n += toolNeed
	case it.Damage > 0 && v.adult:
		n += weaponNeed
	}
	return n
}

// held is how many of id they have, counting their family's stores; foods,
// tools and weapons are each counted together.
func (v *valuer) held(id chem.ItemID) int {
	switch it := v.s.cat.item(id); {
	case it.Food > 0:
		return v.food
	case it.Gather > 0:
		return v.tools
	case it.Damage > 0:
		return v.weapons
	}
	return v.s.holding(v.c, id)
}

// spare is how many units of id they could hand over.
func (v *valuer) spare(id chem.ItemID) int {
	n := v.c.Inventory[id]
	if v.keep != nil {
		n -= v.keep(id)
	}
	return max(0, n)
}

// worth is what an item means to someone: what one unit would be worth if
// it were their only one, and how many they hold.
type worth struct {
	need float64
	held int
}

func (v *valuer) of(id chem.ItemID) worth { return worth{v.need(id), v.held(id)} }

// gain is what getting n more units is worth to them.
func (w worth) gain(n int) float64 {
	sum := 0.0
	for j := 1; j <= n; j++ {
		sum += w.need / float64(w.held+j)
	}
	return sum
}

// loss is what handing over n of their units costs them.
func (w worth) loss(n int) float64 {
	sum := 0.0
	for i := range n {
		sum += w.need / float64(max(1, w.held-i))
	}
	return sum
}

// offer is an item one of two could hand over: how many, and what it means
// to them (owner) and to the other (taker).
type offer struct {
	id           chem.ItemID
	spare        int
	food         bool
	owner, taker worth
}

// offers lists what from could hand over to to, by item id.
func (s *Sim) offers(from, to *valuer) []offer {
	var out []offer
	for _, id := range from.c.Inventory.ids() {
		if n := min(from.spare(id), tradeBundle); n > 0 {
			out = append(out, offer{id, n, s.cat.isFood(id), from.of(id), to.of(id)})
		}
	}
	return out
}

// barter finds and makes the best swap between a and b, if there is one
// both gain from; it reports whether they traded.
func (s *Sim) barter(a, b *Creature) bool {
	va, vb := s.valuer(a), s.valuer(b)
	roomA, roomB := invCapacity-a.Inventory.count(), invCapacity-b.Inventory.count()
	var x, y chem.ItemID
	nx, ny, best := 0, 0, 0.0
	fromB := s.offers(vb, va)
	for _, ox := range s.offers(va, vb) {
		for _, oy := range fromB {
			// One food for another isn't worth a trade: any food will do.
			if ox.id == oy.id || ox.food && oy.food {
				continue
			}
			for n := 1; n <= ox.spare; n++ {
				for m := 1; m <= oy.spare; m++ {
					// Nobody ends up carrying more than they can.
					if m-n > roomA || n-m > roomB {
						continue
					}
					ga := oy.taker.gain(m) - ox.owner.loss(n)
					gb := ox.taker.gain(n) - oy.owner.loss(m)
					if ga < tradeMinGain || gb < tradeMinGain {
						continue
					}
					if p := ga * gb; p > best {
						x, y, nx, ny, best = ox.id, oy.id, n, m, p
					}
				}
			}
		}
	}
	if best == 0 {
		return false
	}
	nx = a.Inventory.take(x, nx)
	ny = b.Inventory.take(y, ny)
	b.Inventory.add(x, nx)
	a.Inventory.add(y, ny)
	// What each wants has changed with what they now hold.
	s.updateAbilities(a)
	s.updateAbilities(b)
	s.rememberRelation(a, b, tradeTrust)
	s.rememberRelation(b, a, tradeTrust)
	s.feel(a, moodJoy, tradeJoy)
	s.feel(b, moodJoy, tradeJoy)
	s.recordTrade(a, x, nx, b, y, ny)
	return true
}

// recordTrade keeps the world's trade record and, now and then, logs one.
func (s *Sim) recordTrade(a *Creature, x chem.ItemID, nx int, b *Creature, y chem.ItemID, ny int) {
	e := &s.exch
	e.Trades++
	e.Units += nx + ny
	if e.ByItem == nil {
		e.ByItem = map[chem.ItemID]int{}
	}
	e.ByItem[x] += nx
	e.ByItem[y] += ny
	e.Recent = append(e.Recent, TradeRecord{s.time(), Ref{a.ID, a.Name}, x, nx, Ref{b.ID, b.Name}, y, ny})
	if over := len(e.Recent) - recentTrades; over > 0 {
		e.Recent = slices.Delete(e.Recent, 0, over)
	}
	if e.Trades == 1 || s.time()-e.TradeEvent >= tradeEventGap {
		e.TradeEvent = s.time()
		s.event("trade", fmt.Sprintf("%s menukar %d %s dengan %d %s milik %s",
			a.Name, nx, s.cat.itemName(x), ny, s.cat.itemName(y), b.Name), a.ID)
	}
}
