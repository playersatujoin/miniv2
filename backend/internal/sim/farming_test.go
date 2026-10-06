package sim

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/ecology"
)

// calmWorld is a world with the real chemistry, steady rain and no wild
// animals, plus a tile where yams grow well.
func calmWorld(t *testing.T) (*Sim, int, int) {
	t.Helper()
	m := testMap(t, 48)
	s := newSimWith(m, 1, defaultCatalog())
	s.eco = ecology.New(ecology.LandFromMap(m), s.rng, ecology.Options{NoClimate: true, NoFauna: true}, 0)
	for _, i := range s.terrain.walkable {
		if s.eco.SuitsHere(int(i), "ubi") > 0.5 && !s.terrain.nearWater[i] {
			return s, int(i) % s.terrain.w, int(i) / s.terrain.w
		}
	}
	t.Fatal("nowhere to grow yams on the test map")
	return nil, 0, 0
}

// growFor advances only the land, so blank-brained people stay put.
func growFor(s *Sim, seconds float64) {
	for range int(seconds * TicksPerSecond) {
		s.tick++
		s.eco.Tick(s.tick, s.time(), dt, s)
	}
}

func TestPlantingAndTheFirstHarvestInventFarming(t *testing.T) {
	s, x, y := calmWorld(t)
	c := person(s, Female, "Tani", x, y)
	c.Inventory.add("ubi", 2)
	if id, _ := s.plantChoice(c); id != "ubi" {
		t.Fatalf("a yam suits this ground but plantChoice = %q", id)
	}
	if !s.plant(c) || c.Inventory["ubi"] != 1 || c.Deeds.Planted != 1 {
		t.Fatalf("planting failed: inventory %v deeds %+v", c.Inventory, c.Deeds)
	}
	tile, _ := s.terrain.index(x, y)
	if p := s.eco.PlotAt(tile); p == nil || p.Crop != "ubi" {
		t.Fatal("the yam should go in where she stands")
	}
	// The next one goes into free ground next to her.
	if !s.plant(c) || c.Inventory["ubi"] != 0 {
		t.Fatal("a second yam should go in next to the first")
	}
	plots := 0
	for k := range 9 {
		if i, ok := s.terrain.index(x+k%3-1, y+k/3-1); ok && s.eco.PlotAt(i) != nil {
			plots++
		}
	}
	if plots != 2 {
		t.Fatalf("%d plots around her, want 2", plots)
	}
	growFor(s, 0.75*SecondsPerYear*2)
	if p := s.eco.PlotAt(tile); p == nil || !p.IsRipe() {
		t.Fatalf("yams not ripe after a year and a half: %+v", p)
	}
	if got, ok := s.ripeInReach(c, x, y); !ok || got != tile {
		t.Fatal("the planter can't reach her own ripe field")
	}
	s.harvest(c, tile)
	if c.Inventory["ubi"] == 0 || c.Deeds.Harvested == 0 {
		t.Fatalf("harvest gave nothing: %v", c.Inventory)
	}
	if !s.techKnown(chem.FarmingTech) || !strings.Contains(lastEvent(s)+eventsText(s), "Pertanian") {
		t.Fatal("the first harvest of a planted crop should invent farming")
	}
}

func eventsText(s *Sim) string {
	var b strings.Builder
	for _, e := range s.events {
		b.WriteString(e.Text + "\n")
	}
	return b.String()
}

func TestOthersMayOnlyStealAFamilysHarvest(t *testing.T) {
	s, x, y := calmWorld(t)
	owner := person(s, Female, "Tani", x, y)
	owner.Inventory.add("ubi", 1)
	s.plant(owner)
	tile, _ := s.terrain.index(x, y)
	growFor(s, 0.75*SecondsPerYear*2)
	thief := person(s, Male, "Maling", x, y)
	if _, ok := s.ripeInReach(thief, x, y); ok {
		t.Fatal("a stranger may harvest someone else's field")
	}
	if !s.steal(thief) || thief.Inventory["ubi"] == 0 || s.crimes != 1 {
		t.Fatalf("stealing the harvest failed: %v, crimes %d", thief.Inventory, s.crimes)
	}
	if !strings.Contains(eventsText(s), "panen dari ladang") {
		t.Fatal("crop theft not reported")
	}
	if p := s.eco.PlotAt(tile); p != nil && p.Left > 0 && s.mayHarvest(thief, p) {
		t.Fatal("theft made the field the thief's")
	}
}

func TestHuntingAKillGivesMeatWithoutACrime(t *testing.T) {
	s, x, y := calmWorld(t)
	c := person(s, Male, "Pemburu", x, y)
	a := s.eco.AddAnimal("rusa", c.X+0.6, c.Y, s.time())
	for range 20 {
		if a.Dead() {
			break
		}
		if !s.hunt(c) {
			t.Fatal("no prey in reach")
		}
	}
	if !a.Dead() || c.Inventory["daging"] == 0 || c.Deeds.Hunted != 1 {
		t.Fatalf("hunt failed: dead %v meat %d", a.Dead(), c.Inventory["daging"])
	}
	if s.crimes != 0 || c.Reputation < 0 {
		t.Fatal("hunting an animal must not count as a crime")
	}
}

func TestFierceAnimalsFightBackAndCanKill(t *testing.T) {
	s, x, y := calmWorld(t)
	c := person(s, Male, "Nekat", x, y)
	c.Health = 0.1
	tiger := s.eco.AddAnimal("harimau", c.X+0.5, c.Y, s.time())
	for range 50 {
		if c.Health <= 0 || tiger.Dead() {
			break
		}
		s.hunt(c)
	}
	if c.Health > 0 {
		t.Skip("the tiger never turned on the hunter in 50 tries (chance)")
	}
	s.reap()
	if s.deathsBy.Animal != 1 || !strings.Contains(eventsText(s), "diterkam harimau") {
		t.Fatalf("death by a tiger not recorded: %+v\n%s", s.deathsBy, eventsText(s))
	}
}

func TestBabiesAreCarriedAndNursed(t *testing.T) {
	s, x, y := calmWorld(t)
	mother := person(s, Female, "Ibu", x, y)
	baby := person(s, Male, "Bayi", x+2, y)
	baby.BornTick = s.tick
	baby.Mother = &Ref{mother.ID, mother.Name}
	baby.Energy = 0.5
	s.act(baby)
	if math.Hypot(baby.X-mother.X, baby.Y-mother.Y) > 0.5 {
		t.Fatal("a baby should ride with its mother")
	}
	before := mother.Energy
	for range TicksPerSecond {
		s.act(baby)
	}
	if baby.Energy <= 0.5 || mother.Energy >= before {
		t.Fatalf("nursing: baby %.2f, mother %.2f→%.2f", baby.Energy, before, mother.Energy)
	}
	baby.BornTick = s.tick - int64(3*SecondsPerYear*TicksPerSecond)
	if s.carrier(baby) != nil {
		t.Fatal("a three-year-old walks on its own")
	}
}

func TestFoodKeepsLongerInAGranary(t *testing.T) {
	s, x, y := calmWorld(t)
	c := person(s, Female, "Simpan", x, y)
	c.Inventory = Stock{chem.Food: 200}
	lumbung := s.addStructure(s.cat.structure["lumbung"], x+1, y, c)
	lumbung.Storage.add(chem.Food, 200)
	for range 8 {
		s.spoil()
	}
	if c.Inventory[chem.Food] >= lumbung.Storage[chem.Food] {
		t.Fatalf("after a year: %d left in hand, %d in the granary", c.Inventory[chem.Food], lumbung.Storage[chem.Food])
	}
	// A year is one half-life in the open and a quarter of one in a granary.
	keeps := s.cat.item(chem.Food).Keeps
	wantHand := 200 * math.Exp2(-1/keeps)
	wantGranary := 200 * math.Exp2(-1/(keeps*spoilGranary))
	if math.Abs(float64(c.Inventory[chem.Food])-wantHand) > 15 || math.Abs(float64(lumbung.Storage[chem.Food])-wantGranary) > 15 {
		t.Fatalf("spoilage off: hand %d (want ≈%.0f), granary %d (want ≈%.0f) of 200",
			c.Inventory[chem.Food], wantHand, lumbung.Storage[chem.Food], wantGranary)
	}
}

func TestFeedingTamesAnimalsForTheHousehold(t *testing.T) {
	s, x, y := calmWorld(t)
	c := person(s, Female, "Peternak", x, y)
	s.moveIn(c, s.addStructure(s.cat.structure["gubuk"], x, y, c))
	hen := s.eco.AddAnimal("ayam_hutan", c.X+0.5, c.Y, s.time())
	c.Inventory.add(chem.Food, 10)
	for range 10 {
		if hen.Owner != 0 {
			break
		}
		if !s.give(c) {
			t.Fatal("with nobody to share with, food should go to the animal")
		}
	}
	if hen.Owner != c.HouseID || !s.techKnown(chem.HerdingTech) || c.Deeds.Tamed != 1 {
		t.Fatalf("taming failed: owner %d, herding known %v", hen.Owner, s.techKnown(chem.HerdingTech))
	}
	if lv := s.houseView(s.houseOf(c)).Livestock; len(lv) != 1 || lv[0].Name != "Ayam" {
		t.Fatalf("house livestock %+v", lv)
	}
}

func TestNoHardPopulationCap(t *testing.T) {
	s := New(testMap(t, 48), 1)
	if s.capacity < 300 || float64(s.capacity) <= float64(len(s.terrain.walkable))/60 {
		t.Fatalf("technical ceiling %d looks like the old hard cap", s.capacity)
	}
}

func TestFrameCarriesWeatherAnimalsAndFields(t *testing.T) {
	s := New(testMap(t, 48), 4)
	for range 40 {
		s.step()
	}
	var raw struct {
		W []float64   `json:"w"`
		A [][]float64 `json:"a"`
		C [][]any     `json:"c"`
	}
	if err := json.Unmarshal(s.encodeFrame(), &raw); err != nil {
		t.Fatal(err)
	}
	if len(raw.W) != 5 || raw.W[0] < 0 || raw.W[0] >= 1 || len(raw.A) == 0 || len(raw.A[0]) != 6 {
		t.Fatalf("frame weather %v, %d animals", raw.W, len(raw.A))
	}
	c := s.creatures[0]
	c.Inventory.add("ubi", 1)
	tile, _ := s.terrain.indexAt(c.X, c.Y)
	s.eco.Plant(tile, "ubi", c.ID, 0, 0.5, s.time())
	data, _ := s.Fields()
	var f struct {
		V int64     `json:"v"`
		P [][]int64 `json:"p"`
	}
	if err := json.Unmarshal(data, &f); err != nil || len(f.P) == 0 || len(f.P[0]) != 5 {
		t.Fatalf("fields message %s (%v)", data, err)
	}
	if v := s.Ecology(); len(v.Species) != ecology.SpeciesCount || len(v.Crops) != len(chem.Crops()) || v.Plots != 1 {
		t.Fatalf("ecology view: %d species, %d crops, %d plots", len(v.Species), len(v.Crops), v.Plots)
	}
}

func farmer(s *Sim, name string, x, y int) *Creature {
	c := person(s, Female, name, x, y)
	c.Skills = map[string]float64{chem.FarmingTech: 0.5}
	return c
}

func TestFarmersKeepSeedUntilFamine(t *testing.T) {
	s, x, y := calmWorld(t)
	c := farmer(s, "Tani", x, y)
	c.Inventory = Stock{"padi": 2}
	c.Energy = 0.5
	if id := s.cat.edible(c.Inventory, s.handKeep(c)); id != "" {
		t.Fatalf("a farmer ate her seed (%s) while only peckish", id)
	}
	if s.mealsInHand(c) != 0 {
		t.Fatalf("seed counted as a meal: %d", s.mealsInHand(c))
	}
	c.Inventory.add(chem.Food, 1)
	if id := s.cat.edible(c.Inventory, s.handKeep(c)); id != chem.Food {
		t.Fatalf("she should eat the wild food, not the seed: %q", id)
	}
	c.Energy = seedHunger / 2
	c.Inventory.take(chem.Food, 1)
	if id := s.cat.edible(c.Inventory, s.handKeep(c)); id != "padi" {
		t.Fatal("in a famine the seed is eaten too")
	}
	// Someone who doesn't farm keeps nothing back.
	o := person(s, Male, "Peramu", x, y)
	o.Inventory = Stock{"padi": 2}
	o.Energy = 0.5
	if s.cat.edible(o.Inventory, s.handKeep(o)) != "padi" {
		t.Fatal("a forager has no reason to keep seed")
	}
}

func TestFarmersTakeSeedFromTheStoreAndLeaveSomeBehind(t *testing.T) {
	s, x, y := calmWorld(t)
	c := farmer(s, "Tani", x, y)
	home := s.addStructure(s.cat.structure["gubuk"], x, y, c)
	s.moveIn(c, home)
	lumbung := s.addStructure(s.cat.structure["lumbung"], x+1, y, c)
	lumbung.Storage.add("ubi", 7)
	c.Energy = 0.9
	s.keepHouse(c)
	if c.Inventory["ubi"] != seedCarry || lumbung.Storage["ubi"] != 7-seedCarry {
		t.Fatalf("seed not taken out to sow: hand %v, granary %v", c.Inventory, lumbung.Storage)
	}
	// Back home with a harvest: two units stay in hand to sow, the rest is stored.
	c.Inventory.add("ubi", 4)
	s.keepHouse(c)
	if c.Inventory["ubi"] != seedCarry {
		t.Fatalf("hand %v after storing a harvest", c.Inventory)
	}
	// A hungry child at home eats from the store but leaves the seed.
	kid := person(s, Male, "Anak", x, y)
	kid.BornTick = s.tick // a newborn: not a sower
	kid.Mother = &Ref{c.ID, c.Name}
	s.moveIn(kid, home)
	kid.Energy = 0.4
	lumbung.Storage = Stock{"ubi": seedKeep + 1}
	s.keepHouse(kid)
	if kid.Inventory["ubi"] != 1 || lumbung.Storage["ubi"] != seedKeep {
		t.Fatalf("child took %v, granary left %v", kid.Inventory, lumbung.Storage)
	}
}

func TestWomenNeedLessFoodThanMen(t *testing.T) {
	s, x, y := calmWorld(t)
	w, m := person(s, Female, "Hawa", x, y), person(s, Male, "Adam", x, y)
	if r := s.bodyScale(w) / s.bodyScale(m); math.Abs(r-femaleBody) > 1e-9 {
		t.Fatalf("a grown woman needs %.2f of a man's food, want %.2f", r, femaleBody)
	}
	girl, boy := person(s, Female, "Bayi", x, y), person(s, Male, "Bayu", x, y)
	girl.BornTick, boy.BornTick = s.tick, s.tick
	if s.bodyScale(girl) != s.bodyScale(boy) {
		t.Fatal("newborn girls and boys need the same")
	}
	nine := s.tick - int64(9*SecondsPerYear*TicksPerSecond)
	girl.BornTick, boy.BornTick = nine, nine
	if s.bodyScale(girl) != s.bodyScale(boy) {
		t.Fatal("girls and boys need the same before puberty")
	}
}

func TestOnlyFreshWaterQuenchesThirstAndPeopleRememberIt(t *testing.T) {
	m := testMap(t, 96)
	s := newSimWith(m, 1, defaultCatalog())
	tr := s.terrain
	river, coast := -1, -1
	for _, i := range tr.walkable {
		switch {
		case tr.nearFresh[i] && river < 0:
			river = int(i)
		case tr.nearWater[i] && !tr.nearFresh[i] && coast < 0:
			coast = int(i)
		}
	}
	if river < 0 || coast < 0 {
		t.Fatal("test map needs a riverbank and a sea shore")
	}
	c := person(s, Female, "Haus", coast%tr.w, coast/tr.w)
	if s.canDrink(c) {
		t.Fatal("drank sea water")
	}
	// Far from any water, the place she last drank pulls on the water rays.
	far := -1
	for _, i := range tr.walkable {
		x, y := int(i)%tr.w, int(i)/tr.w
		if math.Hypot(float64(x-river%tr.w), float64(y-river/tr.w)) > 20 && !tr.nearWater[i] {
			far = int(i)
			break
		}
	}
	c.X, c.Y = float64(far%tr.w)+0.5, float64(far/tr.w)+0.5
	c.WaterX, c.WaterY = float64(river%tr.w)+0.5, float64(river/tr.w)+0.5
	c.Heading = math.Atan2(c.WaterY-c.Y, c.WaterX-c.X) // facing it
	s.sense(c)
	if c.input[inWater+2] != waterMemory {
		t.Fatalf("remembered river not felt straight ahead: water rays %v", c.input[inWater:inWater+numRays])
	}
	c.Heading += math.Pi // now it lies behind her
	s.sense(c)
	if c.input[inWater] == 0 && c.input[inWater+numRays-1] == 0 {
		t.Fatalf("a river behind should pull on an outer ray: %v", c.input[inWater:inWater+numRays])
	}
	c.X, c.Y = c.WaterX, c.WaterY
	if !s.canDrink(c) {
		t.Fatal("can't drink by the river")
	}
}

func TestFamiliesSetSnaresAndCollectTheCatch(t *testing.T) {
	s, x, y := calmWorld(t)
	c := person(s, Male, "Pemburu", x, y)
	home := s.addStructure(s.cat.structure["gubuk"], x, y, c)
	s.moveIn(c, home)
	k := s.cat.structure["jerat"]
	if !s.amenityWanted(c, home, k) {
		t.Fatal("a family without a snare should want one")
	}
	plan := s.placeSnare(home, k)
	if plan == nil || math.Hypot(float64(plan.job.X-x), float64(plan.job.Y-y)) > snareSite {
		t.Fatalf("snare placed badly: %+v", plan)
	}
	sn := s.addStructure(k, plan.job.X, plan.job.Y, c)
	tile, _ := s.terrain.index(sn.X, sn.Y)
	if s.eco.Management(tile)&ecology.SnareSet == 0 {
		t.Fatal("a new snare should be set")
	}
	if s.amenityWanted(c, home, k) {
		t.Fatal("one snare per house is enough")
	}
	s.Snared(tile, "rusa", 6)
	s.applyFarms()
	if sn.Storage["daging"] != 6 || s.eco.Management(tile)&ecology.SnareSet != 0 {
		t.Fatalf("catch not held, or snare not sprung: %v", sn.Storage)
	}
	s.keepHouse(c)
	if s.holding(c, "daging") != 6 || len(sn.Storage) != 0 || s.eco.Management(tile)&ecology.SnareSet == 0 {
		t.Fatalf("catch not collected or snare not reset: holding %d, snare %v", s.holding(c, "daging"), sn.Storage)
	}
}
