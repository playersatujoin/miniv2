package sim

import (
	"strings"
	"testing"

	"miniv2/backend/internal/chem"
)

// tradePair is facePair with both wanting to trade.
func tradePair(t *testing.T) (*Sim, *Creature, *Creature) {
	t.Helper()
	s, a, b := facePair(t)
	a.output[outTrade], b.output[outTrade] = 0.9, 0.9
	return s, a, b
}

func TestBarterWhenBothGain(t *testing.T) {
	s, a, b := tradePair(t)
	// Sari is starving with two picks; Budi is well fed with food to spare
	// and no tool.
	a.Energy = 0.3
	a.Inventory = Stock{"beliung": 2}
	b.Inventory = Stock{chem.Food: 6}
	va, vb := s.valuer(a), s.valuer(b)
	exchangeTick(s)
	if a.Deeds.Trades != 1 || b.Deeds.Trades != 1 || s.exch.Trades != 1 || a.action != ActTrade {
		t.Fatalf("no trade: %+v %v %v", a.Deeds, a.Inventory, b.Inventory)
	}
	// The best swap for both: one pick for three meals (see barter).
	if a.Inventory["beliung"] != 1 || a.Inventory[chem.Food] != 3 || b.Inventory["beliung"] != 1 || b.Inventory[chem.Food] != 3 {
		t.Fatalf("after the trade Sari %v, Budi %v", a.Inventory, b.Inventory)
	}
	if va.of(chem.Food).gain(3)-va.of("beliung").loss(1) < tradeMinGain || vb.of("beliung").gain(1)-vb.of(chem.Food).loss(3) < tradeMinGain {
		t.Fatal("one of them lost by it")
	}
	if !strings.Contains(eventsText(s), "Sari menukar 1 Beliung dengan 3 Makanan milik Budi") {
		t.Fatalf("trade not logged: %q", eventsText(s))
	}
	if s.exchangeFlags(a) != flagTrading || s.opinion(a, b) <= 0 {
		t.Fatal("should be seen trading, and trust each other a little more")
	}
	info := s.exchangeInfo()
	if info.Trades != 1 || info.Units != 4 || len(info.Recent) != 1 || info.Recent[0].GaveN != 1 || info.Recent[0].GotN != 3 ||
		len(info.Items) != 2 || info.Items[0].Item != chem.Food || info.Items[0].Units != 3 {
		t.Fatalf("info %+v", info)
	}
}

func TestNoBarterUnlessBothGain(t *testing.T) {
	s, a, b := tradePair(t)
	// Both hungry. Sari has only fibre nobody needs; Budi's two meals and his
	// pick are worth more to him than any amount of it.
	a.Energy, b.Energy = 0.3, 0.3
	a.Inventory = Stock{"serat": 3}
	b.Inventory = Stock{chem.Food: 2, "beliung": 1}
	exchangeTick(s)
	if a.Deeds.Trades+b.Deeds.Trades != 0 || s.exch.Trades != 0 || a.Inventory["serat"] != 3 || b.Inventory[chem.Food] != 2 {
		t.Fatalf("traded at a loss: Sari %v Budi %v", a.Inventory, b.Inventory)
	}
	// They met and found nothing to swap: both stand down for a moment.
	if a.Chat == nil || a.Chat.OK || a.Chat.Ended != s.tick || b.Chat == nil || b.Chat.OK || s.exchangeFlags(a) != 0 {
		t.Fatalf("chats after a failed trade: %+v %+v", a.Chat, b.Chat)
	}

	// Two who need neither don't swap stones for rope just for variety.
	// (Without tools, stone and rope would make a pick, and they would.)
	s2, c, d := tradePair(t)
	s2.techs["alat"] = &Discovery{}
	c.Inventory = Stock{"batu": 3, "beliung": 1}
	d.Inventory = Stock{"tali": 3, "beliung": 1}
	if s2.wantsOf(c)["tali"] != 0 || s2.wantsOf(d)["batu"] != 0 {
		t.Fatalf("they should want neither: %v %v", s2.wantsOf(c), s2.wantsOf(d))
	}
	exchangeTick(s2)
	if c.Deeds.Trades != 0 {
		t.Fatalf("traded without a need: %v %v", c.Inventory, d.Inventory)
	}
}

func TestBarterNeverOverfills(t *testing.T) {
	s, a, b := tradePair(t)
	a.Energy = 0.3
	a.Inventory = Stock{"beliung": 1, "serat": invCapacity - 1}
	b.Inventory = Stock{chem.Food: invCapacity}
	exchangeTick(s)
	if a.Deeds.Trades != 1 {
		t.Fatal("a one-for-one swap fits")
	}
	if a.Inventory.count() > invCapacity || b.Inventory.count() > invCapacity || a.Inventory[chem.Food] != 1 || b.Inventory["beliung"] != 1 {
		t.Fatalf("over capacity: Sari %d %v, Budi %d %v", a.Inventory.count(), a.Inventory, b.Inventory.count(), b.Inventory)
	}

	// With room on one side only, the full one can't take more than they give.
	s2, c, d := tradePair(t)
	c.Energy = 0.3
	c.Inventory = Stock{"beliung": 2}
	d.Inventory = Stock{chem.Food: invCapacity}
	exchangeTick(s2)
	if d.Inventory.count() > invCapacity || c.Inventory.count() > invCapacity {
		t.Fatalf("over capacity: %v %v", c.Inventory, d.Inventory)
	}
}

// Seed a sower keeps back is not for sale.
func TestBarterKeepsSeed(t *testing.T) {
	s, a, _ := tradePair(t)
	a.Inventory = Stock{"beliung": 3}
	v := s.valuer(a)
	v.keep = func(id chem.ItemID) int {
		if id == "beliung" {
			return 2
		}
		return 0
	}
	if v.spare("beliung") != 1 || v.spare("kayu") != 0 {
		t.Fatal("only what is not kept back is spare")
	}
}

// Someone who has gathered the wood for the hut they mean to build doesn't
// part with it for a few stones, though wantsOf no longer lists wood once
// enough is in hand; otherwise the same two would trade it straight back.
func TestBarterKeepsWhatTheGoalNeeds(t *testing.T) {
	s, a, b := tradePair(t)
	s.techs["alat"] = &Discovery{}
	a.Inventory = Stock{"kayu": 2, "beliung": 1}
	b.Inventory = Stock{"batu": 3, "beliung": 1}
	if s.wantsOf(a)["kayu"] != 0 || s.wantsOf(b)["kayu"] == 0 {
		t.Fatalf("Sari has the wood for a hut, Budi still needs it: %v %v", s.wantsOf(a), s.wantsOf(b))
	}
	exchangeTick(s)
	if a.Deeds.Trades != 0 || a.Inventory["kayu"] != 2 {
		t.Fatalf("sold the wood for her hut: %v %v", a.Inventory, b.Inventory)
	}
}
