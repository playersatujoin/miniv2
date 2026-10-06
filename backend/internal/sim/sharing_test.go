package sim

import (
	"bytes"
	"math"
	"strings"
	"testing"

	"miniv2/backend/internal/chem"
)

// fill gives c a load of firewood, leaving room hands free.
func fill(c *Creature, room int) {
	c.Inventory.add("kayu", invCapacity-c.Inventory.count()-room)
}

func TestMeatAHunterCannotCarryStaysAsACarcass(t *testing.T) {
	for _, off := range []bool{false, true} {
		s, x, y := calmWorld(t)
		s.opts.NoSharing = off
		c := person(s, Male, "Pemburu", x, y)
		fill(c, 2)
		a := s.eco.AddAnimal("rusa", c.X+0.6, c.Y, s.time())
		for range 40 {
			if a.Dead() || !s.hunt(c) {
				break
			}
		}
		if !a.Dead() || c.Inventory[meatItem] != 2 {
			t.Fatalf("off=%v: the hunter should carry 2 units: dead %v, carries %d", off, a.Dead(), c.Inventory[meatItem])
		}
		meat := a.Kind().Meat
		if off {
			if len(s.share.Carcasses) != 0 {
				t.Fatal("without sharing the rest of the kill should vanish as before")
			}
			continue
		}
		if len(s.share.Carcasses) != 1 || s.carcassMeat() != meat-2 || s.share.Stats.MeatLeft != meat-2 {
			t.Fatalf("want one carcass with %d meat, got %d carcasses, %d meat", meat-2, len(s.share.Carcasses), s.carcassMeat())
		}
		k := s.share.Carcasses[0]
		if math.Hypot(k.X-a.X, k.Y-a.Y) > 1e-9 || k.Species != "rusa" {
			t.Fatalf("carcass at (%.2f, %.2f) %q, the deer fell at (%.2f, %.2f)", k.X, k.Y, k.Species, a.X, a.Y)
		}
		if info := s.stimuliInfo(); info.ByKind[stimCarcass] != 1 {
			t.Fatalf("the carcass should be a lasting stimulus: %+v", info)
		}
	}
}

func TestMeatNobodyCanCarryOrStoreIsLeftWhereItIs(t *testing.T) {
	s, x, y := calmWorld(t)
	c := person(s, Male, "Jagal", x, y)
	fill(c, 1)
	s.receive(c, meatItem, 6)
	s.receive(c, "batu", 3) // anything else is lost as before
	if c.Inventory[meatItem] != 1 || s.carcassMeat() != 5 {
		t.Fatalf("hand %d, left %d; want 1 and 5", c.Inventory[meatItem], s.carcassMeat())
	}
	// At home the house takes what it has room for first.
	h := s.addStructure(s.cat.structure["gubuk"], x, y, c)
	s.moveIn(c, h)
	h.Storage.add("kayu", h.kind.Storage-2)
	s.receive(c, meatItem, 4)
	if h.Storage[meatItem] != 2 || s.carcassMeat() != 7 {
		t.Fatalf("house %d, left %d; want 2 and 7", h.Storage[meatItem], s.carcassMeat())
	}
}

func TestCarcassesAreEatenCutSensedAndRot(t *testing.T) {
	s, x, y := calmWorld(t)
	c := person(s, Female, "Lapar", x, y)
	c.Energy = 0.3
	s.leaveCarcass(c.X+1, c.Y, 40, "kerbau")
	if s.meatHere(c) != 1 {
		t.Fatalf("a carcass next to her should feel like plenty of food, got %.2f", s.meatHere(c))
	}
	i, _ := s.terrain.index(x+1, y)
	if s.meatFood(i) != 1 || s.meatFood(i+1) != 0 {
		t.Fatal("the rays should see meat on the carcass's tile only")
	}
	if !s.eatCarcass(c) || c.Energy <= 0.3 || s.carcassMeat() != 39 || s.share.Stats.MeatTaken != 1 {
		t.Fatalf("eating from the carcass: energy %.2f, meat left %d", c.Energy, s.carcassMeat())
	}
	c.Want = map[chem.ItemID]float64{chem.Food: 0.5} // nothing else on her mind
	if _, what, tile := s.chooseGather(c, x, y, float32(s.cat.foodEnergy()/foodValue)); what != "carcass" || tile != i {
		t.Fatalf("a hungry gatherer by a carcass should cut meat, chose %q", what)
	}
	s.cutMeat(c, i)
	if c.Inventory[meatItem] != 1 || s.carcassMeat() != 38 {
		t.Fatalf("cutting meat: carries %d, left %d", c.Inventory[meatItem], s.carcassMeat())
	}
	// A half-life is a quarter of a year: two seconds.
	for range int(carcassKeeps * SecondsPerYear) {
		s.tick += TicksPerSecond
		s.rotCarcasses()
	}
	if left := s.carcassMeat(); left < 14 || left > 24 {
		t.Fatalf("after one half-life %d of 38 units are left", left)
	}
	for range 40 {
		s.tick += TicksPerSecond
		s.rotCarcasses()
	}
	if len(s.share.Carcasses) != 0 || s.carcassOn(i) != nil || s.meatHere(c) != 0 {
		t.Fatal("a rotten carcass should be gone, from the list and the tiles")
	}
	if st := s.share.Stats; st.MeatLeft != st.MeatTaken+st.MeatRotted {
		t.Fatalf("meat not accounted for: %+v", st)
	}
}

func TestCarcassesMergeAndStayBounded(t *testing.T) {
	s, x, y := calmWorld(t)
	s.leaveCarcass(float64(x)+0.2, float64(y)+0.2, 3, "")
	s.leaveCarcass(float64(x)+0.8, float64(y)+0.7, 2, "")
	if len(s.share.Carcasses) != 1 || s.carcassMeat() != 5 {
		t.Fatal("meat left on the same tile should join the carcass there")
	}
	n := 0
	for _, i := range s.terrain.walkable {
		if n == maxCarcasses+5 {
			break
		}
		s.leaveCarcass(float64(int(i)%s.terrain.w)+0.5, float64(int(i)/s.terrain.w)+0.5, 1+n%3, "")
		n++
	}
	if len(s.share.Carcasses) > maxCarcasses {
		t.Fatalf("%d carcasses, bound %d", len(s.share.Carcasses), maxCarcasses)
	}
	for j, k := range s.share.Carcasses {
		if i, _ := s.terrain.indexAt(k.X, k.Y); s.carcassOn(i) != s.share.Carcasses[j] {
			t.Fatal("the tile index is out of step with the list")
		}
	}
}

func TestHungryPeopleEatFromFamilyAndFromBigLoads(t *testing.T) {
	s, x, y := calmWorld(t)
	c := person(s, Male, "Lapar", x, y)
	c.Energy = 0.3
	mother := person(s, Female, "Ibu", x, y)
	c.Mother = &Ref{mother.ID, mother.Name}
	mother.Inventory.add("ubi", 2)
	mother.Energy = 0.8
	regrid := func() { s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures) }
	regrid()
	if !s.scrounge(c) || mother.Inventory["ubi"] != 1 || s.share.Stats.FromKin != 1 || c.Energy <= 0.3 {
		t.Fatalf("a hungry son should eat from his better-fed mother's food: %v", mother.Inventory)
	}
	c.Energy = 0.3
	if s.scrounge(c) {
		t.Fatal("family keep their last meal")
	}
	mother.Inventory.add("ubi", 3)
	mother.Energy = c.Energy - 0.1
	if s.scrounge(c) {
		t.Fatal("a hungrier mother's food is hers")
	}
	mother.Inventory = nil

	// A stranger's load is only shared beyond what they can use.
	other := person(s, Female, "Kaya", x, y)
	other.Energy = 1
	regrid()
	other.Inventory.add(meatItem, bigPackage)
	if s.scrounge(c) {
		t.Fatal("a neighbour's food is theirs up to a big package")
	}
	other.Inventory.add(meatItem, 2)
	if !s.scrounge(c) || other.Inventory[meatItem] != bigPackage+1 || s.share.Stats.FromOthers != 1 {
		t.Fatalf("a load beyond %d should be shared: %v", bigPackage, other.Inventory)
	}
	c.Energy = 0.3
	if s.crimes != 0 || c.Deeds.Crimes != 0 {
		t.Fatal("eating from what is shared is not theft")
	}
	if s.opinion(c, other) <= 0 {
		t.Fatal("whoever was fed should think better of the one who fed them")
	}
	// Seed held back for sowing is never eaten by others.
	other.Inventory = nil
	sower := farmer(s, "Petani", x, y)
	sower.Inventory.add("ubi", seedCarry)
	regrid()
	c.Mother = &Ref{sower.ID, sower.Name}
	if s.scrounge(c) {
		t.Fatal("a sower's seed went to the pot")
	}
	// Not when fed, not out of reach, not with sharing off.
	sower.Inventory.add("talas", 4)
	c.Energy = shareHungry
	if s.scrounge(c) {
		t.Fatal("someone fed enough doesn't ask")
	}
	c.Energy = 0.3
	sower.X += giveRange + 0.5
	regrid()
	if s.scrounge(c) {
		t.Fatal("food out of reach was eaten")
	}
	sower.X -= giveRange + 0.5
	regrid()
	s.opts.NoSharing = true
	if s.scrounge(c) {
		t.Fatal("no sharing with NoSharing")
	}
}

// ripeField plants a yam field of owner's household on (x, y) and grows it.
func ripeField(t *testing.T, s *Sim, owner *Creature, x, y int) int {
	t.Helper()
	owner.Inventory.add("ubi", 1)
	tile, _ := s.terrain.index(x, y)
	if !s.eco.Plant(tile, "ubi", owner.ID, owner.HouseID, 0.5, s.time()) {
		t.Fatal("couldn't plant")
	}
	growFor(s, 0.75*SecondsPerYear*2)
	if p := s.eco.PlotAt(tile); p == nil || !p.IsRipe() {
		t.Fatal("the field didn't ripen")
	}
	return tile
}

func TestAnyoneMayJoinAHarvestForAShareOfIt(t *testing.T) {
	s, x, y := calmWorld(t)
	owner := person(s, Female, "Tani", x+2, y)
	h := s.addStructure(s.cat.structure["gubuk"], x+2, y, owner)
	s.moveIn(owner, h)
	tile := ripeField(t, s, owner, x, y)
	p := s.eco.PlotAt(tile)
	if p.House != h.ID {
		t.Fatal("the field should be the household's")
	}
	worker := person(s, Male, "Buruh", x, y)
	if _, ok := s.ripeInReach(worker, x, y); ok {
		t.Fatal("the field isn't his to harvest for himself")
	}
	if got, ok := s.bawonInReach(worker, x, y); !ok || got != tile {
		t.Fatal("he should be able to join the harvest")
	}
	worker.Want = map[chem.ItemID]float64{chem.Food: 0.5}
	if _, what, _ := s.chooseGather(worker, x, y, float32(s.cat.foodEnergy()/foodValue)); what != "bawon" {
		t.Fatalf("a hungry man by a ripe field should join its harvest, chose %q", what)
	}
	// A big field: the share comes to about a sixth.
	p.Left = 120
	s.lastEcoEvt["bawon"] = -farmEventGap
	h.Storage = nil
	h.kind.Storage = 200
	cut := 0
	for p.Left > 0 {
		s.tick++
		before := p.Left
		s.bawonHarvest(worker, tile)
		cut += int(before - p.Left)
		if worker.Inventory.count() >= invCapacity {
			worker.Inventory = nil
		}
	}
	st := s.share.Stats
	if cut != 120 || st.Bawon+st.BawonShare != 120 || h.Storage["ubi"] != st.Bawon {
		t.Fatalf("cut %d: stored %d (house %d), kept %d", cut, st.Bawon, h.Storage["ubi"], st.BawonShare)
	}
	if share := float64(st.BawonShare) / 120; math.Abs(share-bawonShare) > 0.08 {
		t.Fatalf("harvesters kept %.2f of the harvest, want about %.2f", share, bawonShare)
	}
	if s.crimes != 0 || !strings.Contains(eventsText(s), "bawon") {
		t.Fatal("joining a harvest is no crime, and is reported")
	}
}

func TestNoHarvestToJoinWithoutAStoreOrWithSharingOff(t *testing.T) {
	s, x, y := calmWorld(t)
	owner := person(s, Female, "Tani", x+2, y)
	tile := ripeField(t, s, owner, x, y) // she has no house
	worker := person(s, Male, "Buruh", x, y)
	if _, ok := s.bawonInReach(worker, x, y); ok {
		t.Fatal("a homeless planter's field has nowhere to carry the harvest to")
	}
	h := s.addStructure(s.cat.structure["gubuk"], x+2, y, owner)
	s.moveIn(owner, h)
	s.eco.PlotAt(tile).House = h.ID
	h.Storage.add("kayu", h.kind.Storage)
	if _, ok := s.bawonInReach(worker, x, y); ok {
		t.Fatal("a full store takes no more")
	}
	h.Storage = nil
	s.opts.NoSharing = true
	if _, ok := s.bawonInReach(worker, x, y); ok {
		t.Fatal("no open harvest with NoSharing")
	}
	if _, what, _ := s.chooseGather(worker, x, y, float32(s.cat.foodEnergy()/foodValue)); what == "bawon" {
		t.Fatal("no open harvest with NoSharing")
	}
}

// granaryVillage is a village of three households with a granary by the
// first house, kept by a farming family.
func granaryVillage(t *testing.T) (*Sim, *Structure, *Creature) {
	t.Helper()
	s := newSimWith(villageMap(t), 1, defaultCatalog())
	a, _ := home(s, "A", 10, 10)
	a.Skills = map[string]float64{chem.FarmingTech: 0.5}
	home(s, "B", 15, 10)
	home(s, "C", 10, 15)
	g := s.addStructure(s.cat.structure["lumbung"], 12, 10, a)
	recount(s)
	if len(villagesOf(s)) != 1 {
		t.Fatal("no village")
	}
	return s, g, a
}

func TestHungryVillagersDrawOnTheVillageGranary(t *testing.T) {
	s, g, _ := granaryVillage(t)
	g.Storage.add(chem.Food, 3)
	g.Storage.add("padi", seedKeep+1)
	v := villagesOf(s)[0]
	var b *Creature
	for _, c := range s.creatures {
		if c.Name == "B" {
			b = c
		}
	}
	b.X, b.Y = float64(g.X)+0.5, float64(g.Y)+1.5
	b.Energy = 0.4
	if b.VillageID != v.ID || s.kin(b, s.byID[g.Owner]) {
		t.Fatal("B should be a villager and no kin of the granary's family")
	}
	s.drawGranary(b)
	if s.mealsInHand(b) != granaryMeals || s.share.Stats.GranaryMeals != granaryMeals {
		t.Fatalf("a hungry villager at the granary took %d meals", s.mealsInHand(b))
	}
	// The family's seed stays: one unit of food and one of rice above the seed.
	b.Inventory = nil
	s.drawGranary(b)
	b.Inventory = nil
	s.drawGranary(b)
	if g.Storage["padi"] != seedKeep || g.Storage[chem.Food] != 0 {
		t.Fatalf("granary left with %v, want the seed kept", g.Storage)
	}
	g.Storage.add(chem.Food, 5)
	b.Inventory = nil
	b.Energy = granaryHungry
	s.drawGranary(b)
	if len(b.Inventory) != 0 {
		t.Fatal("someone not hungry took from the granary")
	}
	// A stranger without a house in the village has no claim on it.
	stranger := person(s, Male, "Asing", g.X, g.Y+1)
	stranger.Energy = 0.2
	s.drawGranary(stranger)
	if len(stranger.Inventory) != 0 {
		t.Fatal("an outsider drew on the village granary")
	}
	// Not from afar.
	b.Energy = 0.3
	b.X += homeReach + 1
	s.drawGranary(b)
	if len(b.Inventory) != 0 {
		t.Fatal("a villager drew on the granary from afar")
	}
}

func TestAnOwnerlessGranaryGoesToTheHouseholdBesideIt(t *testing.T) {
	s, g, a := granaryVillage(t)
	far := s.addStructure(s.cat.structure["lumbung"], 40, 40, nil)
	g.Owner, g.OwnerName = 0, ""
	if s.granaryOf(s.houseOf(a)) != nil {
		t.Fatal("an ownerless granary was the family's before")
	}
	s.shareSecond()
	if g.Owner != a.ID || s.granaryOf(s.houseOf(a)) != g || s.share.Stats.Adopted != 1 {
		t.Fatalf("granary owner %d, want %d", g.Owner, a.ID)
	}
	if far.Owner != 0 {
		t.Fatal("a granary far from any house has nobody to take it")
	}
	g.Owner = 0
	s.opts.NoSharing = true
	s.shareSecond()
	if g.Owner != 0 {
		t.Fatal("NoSharing should leave ownerless granaries alone")
	}
}

// landWorld is a calm world with no wild food anywhere, and a riverbank
// tile where the river runs.
func landWorld(t *testing.T) (*Sim, int) {
	t.Helper()
	s, _, _ := calmWorld(t)
	clear(s.eco.State().Forage)
	for _, i := range s.terrain.riverbank {
		if s.riverRunsBy(int(i)) {
			return s, int(i)
		}
	}
	t.Fatal("no running river on the test map")
	return nil, 0
}

func TestHungryPeopleKnowBetterLandNearTheirWater(t *testing.T) {
	s, bank := landWorld(t)
	w := s.terrain.w
	wet := s.wetLand()
	// The candidates: land a few steps from water, and land far from any.
	var nearTile, dryTile = -1, -1
	for _, i := range s.terrain.walkable {
		d := math.Hypot(float64(int(i)%w-bank%w), float64(int(i)/w-bank/w))
		switch {
		case wet[i] && s.share.wetDist[i] >= 4 && d >= 9 && d <= 12 && nearTile < 0:
			nearTile = int(i)
		case !wet[i] && d <= 16 && dryTile < 0:
			dryTile = int(i)
		}
	}
	if nearTile < 0 {
		t.Fatal("no land near water to test with")
	}
	forage := s.eco.State().Forage
	forage[nearTile] = 1
	c := person(s, Female, "Peramu", bank%w, bank/w)
	c.WaterX, c.WaterY = c.X, c.Y
	c.Energy = 0.5
	s.knowLand()
	if len(c.FoodPlaces) != 1 || int(c.FoodPlaces[0].Y)*w+int(c.FoodPlaces[0].X) != nearTile {
		t.Fatalf("she should know of the food near the water: %+v", c.FoodPlaces)
	}
	if _, v := s.bestFoodPlace(c); v < foodPlaceMin {
		t.Fatal("the place should be worth the walk to her senses")
	}
	// Land too far from any water isn't part of anyone's range.
	if dryTile >= 0 {
		c.FoodPlaces = nil
		forage[nearTile] = 0
		forage[dryTile] = 1.5
		s.knowLand()
		if len(c.FoodPlaces) != 0 {
			t.Fatalf("land far from water was offered: %+v", c.FoodPlaces)
		}
		forage[dryTile] = 0
		forage[nearTile] = 1
	}
	// Not for the fed, nor for someone who doesn't know where to drink.
	c.FoodPlaces = nil
	c.Energy = landHungry
	s.knowLand()
	c.Energy = 0.5
	c.WaterX, c.WaterY = 0, 0
	s.knowLand()
	if len(c.FoodPlaces) != 0 {
		t.Fatal("knowledge of land given to the fed or the lost")
	}
	// A crowd there makes it worth less than where she stands.
	c.WaterX, c.WaterY = c.X, c.Y
	forage[bank] = 0.6
	for k := range 8 {
		person(s, Male, "Ramai", nearTile%w, nearTile/w).Energy = 1 + float64(k)
	}
	s.knowLand()
	if len(c.FoodPlaces) != 0 {
		t.Fatalf("crowded land was offered over her own: %+v", c.FoodPlaces)
	}
}

func TestNoSharingSwitchesItAllOff(t *testing.T) {
	s, x, y := calmWorld(t)
	s.opts.NoSharing = true
	s.leaveCarcass(float64(x)+0.5, float64(y)+0.5, 5, "rusa")
	c := person(s, Female, "Lapar", x, y)
	c.Energy = 0.3
	c.WaterX, c.WaterY = c.X, c.Y
	if len(s.share.Carcasses) != 0 || s.eatCarcass(c) || s.scrounge(c) {
		t.Fatal("sharing should be off")
	}
	s.shareSecond()
	if len(c.FoodPlaces) != 0 || s.share.Stats != (SharingStats{}) || s.saveSharing() != nil {
		t.Fatalf("nothing should be shared or known: %+v", s.share.Stats)
	}
	// A carcass left from before the switch goes.
	s.opts.NoSharing = false
	s.leaveCarcass(float64(x)+0.5, float64(y)+0.5, 5, "rusa")
	s.opts.NoSharing = true
	s.shareSecond()
	if len(s.share.Carcasses) != 0 || s.meatHere(c) != 0 {
		t.Fatal("carcasses should be cleared when sharing is off")
	}
}

func TestSharingSavesAndRestoresExactly(t *testing.T) {
	m := testMap(t, 48)
	s := New(m, 3)
	var spot int32 = -1
	for _, i := range s.terrain.walkable {
		spot = i
		break
	}
	w := s.terrain.w
	s.leaveCarcass(float64(int(spot)%w)+0.3, float64(int(spot)/w)+0.6, 9, "babi")
	s.share.Stats.FromKin, s.share.Stats.Bawon = 3, 7
	data, err := s.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	r, err := Restore(testMap(t, 48), data)
	if err != nil {
		t.Fatal(err)
	}
	if r.carcassMeat() != 9 || r.carcassOn(int(spot)) == nil || r.share.Next != s.share.Next || r.share.Stats != s.share.Stats {
		t.Fatalf("restored sharing differs: %+v vs %+v", r.share.Stats, s.share.Stats)
	}
	for _, x := range []*Sim{s, r} {
		x.Advance(30 * TicksPerSecond)
	}
	a, _ := s.MarshalState()
	b, _ := r.MarshalState()
	ja, jb := liveState(t, a), liveState(t, b)
	for k, v := range ja {
		if !bytes.Equal(v, jb[k]) {
			t.Fatalf("the restored world diverged in %s", k)
		}
	}
}

func TestVillagesReportTheirStores(t *testing.T) {
	s, g, a := granaryVillage(t)
	g.Storage.add("padi", 7)
	s.houseOf(a).Storage.add("ubi", 2)
	views := s.VillagesView()
	if len(views) != 1 || views[0].GranaryFood != 7 || views[0].HouseFood != 2 {
		t.Fatalf("village stores: %+v", views)
	}
	info := s.sharingInfo()
	if info.Granaries != 1 || info.GranaryFood != 7 || info.HouseFood != 2 {
		t.Fatalf("sharing info: %+v", info)
	}
}
