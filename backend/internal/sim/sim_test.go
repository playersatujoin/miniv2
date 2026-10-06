package sim

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"miniv2/backend/internal/chem"
	"miniv2/backend/internal/world"
)

var (
	soak        = flag.Bool("soak", false, "run the long evolution soak test")
	soakSeed    = flag.Uint64("soakseed", 42, "world seed for the soak test")
	soakMinutes = flag.Int("soakminutes", 120, "simulated minutes for the soak test")
)

func testMap(t *testing.T, size int) *world.Map {
	t.Helper()
	m := world.Generate(world.GenOptions{Name: "test", Width: size, Height: size, Seed: 1337})
	m.ID = "0123456789abcdef"
	return m
}

// --- a tiny hand-made chemistry, so each rule can be checked in isolation ---

type fakeGeo struct{ at map[[2]int][]chem.Source }

func (g *fakeGeo) put(x, y int, item chem.ItemID, amount float64, tool bool) {
	k := [2]int{x, y}
	g.at[k] = append(g.at[k], chem.Source{Item: item, X: x, Y: y, Amount: amount, NeedsTool: tool})
}

func (g *fakeGeo) Sources(x, y int) []chem.Source {
	var out []chem.Source
	for _, d := range [][2]int{{0, 0}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		out = append(out, g.at[[2]int{x + d[0], y + d[1]}]...)
	}
	return out
}

func (g *fakeGeo) Take(x, y int, item chem.ItemID, amount float64) float64 {
	srcs := g.at[[2]int{x, y}]
	for i := range srcs {
		if srcs[i].Item == item {
			took := math.Min(amount, srcs[i].Amount)
			srcs[i].Amount -= took
			return took
		}
	}
	return 0
}

func (g *fakeGeo) Regrow(float64)                 {}
func (g *fakeGeo) Update(*world.Map)              {}
func (g *fakeGeo) Amounts() []float32             { return nil }
func (g *fakeGeo) SetAmounts([]float32) error     { return nil }
func (g *fakeGeo) MinedOut() [][2]int             { return nil }
func (g *fakeGeo) Woodland() ([]float32, float64) { return nil, 0 }
func (g *fakeGeo) Logged() [][2]int               { return nil }

func fakeCatalog() *catalog {
	type in = map[chem.ItemID]int
	items := []chem.Item{
		{ID: chem.Food, Name: "Makanan", Kind: "makanan", Food: 0.4, Value: 1},
		{ID: "kayu", Name: "Kayu", Elements: []string{"C"}, Value: 1},
		{ID: "serat", Name: "Serat", Value: 0.5},
		{ID: "batu", Name: "Batu", Value: 1},
		{ID: "tali", Name: "Tali", Value: 1},
		{ID: "bijih", Name: "Bijih", Kind: "mineral", Elements: []string{"Fe", "H"}, Value: 3},
		{ID: "besi", Name: "Besi", Kind: "logam", Elements: []string{"Fe"}, Value: 5},
		{ID: "beliung", Name: "Beliung", Kind: "alat", Value: 2, Gather: 1.5},
		{ID: "pedang", Name: "Pedang", Kind: "senjata", Value: 8, Damage: 1},
		{ID: chem.SynthesisFuel, Name: "Uranium", Value: 10},
	}
	recipes := []chem.Recipe{
		{ID: "tali", Name: "Memilin tali", Inputs: in{"serat": 2}, Outputs: in{"tali": 1}, Seconds: 2},
		{ID: "beliung", Name: "Membuat beliung", Inputs: in{"batu": 1, "kayu": 1, "tali": 1}, Outputs: in{"beliung": 1}, Teaches: "alat", Seconds: 3},
		{ID: "lebur", Name: "Melebur bijih", Inputs: in{"bijih": 2}, Outputs: in{"besi": 1}, Station: "tungku", Discovers: []string{"Fe"}, Seconds: 4},
	}
	structures := []chem.StructureKind{
		{ID: "gubuk", Name: "Gubuk", Cost: in{"kayu": 2}, House: true, Level: 1, Storage: 10},
		{ID: "rumah", Name: "Rumah", Cost: in{"kayu": 4}, Tech: "alat", House: true, Level: 2, Upgrades: "gubuk", Storage: 20},
		{ID: "tungku", Name: "Tungku", Cost: in{"batu": 2}, Teaches: "lebur", Tier: 1},
		{ID: "reaktor", Name: "Reaktor", Cost: in{"besi": 1}, Tech: "lebur", Tier: 6},
	}
	techs := []chem.Tech{{ID: "alat", Name: "Alat"}, {ID: "lebur", Name: "Peleburan", Tier: 1}}
	elements := []chem.Element{
		{Z: 1, Symbol: "H", Name: "Hidrogen", Natural: true, Tier: 2},
		{Z: 6, Symbol: "C", Name: "Karbon", Natural: true, Tier: 0},
		{Z: 26, Symbol: "Fe", Name: "Besi", Natural: true, Tier: 1},
		{Z: 43, Symbol: "Tc", Name: "Teknesium", Tier: 6},
	}
	c := newCatalog(items, recipes, structures, techs, elements)
	c.tierName = func(t int) string { return fmt.Sprintf("Zaman %d", t) }
	c.discoverable = func(id chem.ItemID, tier int, known func(string) bool) []string {
		var out []string
		for _, sym := range c.item(id).Elements {
			if e := c.element[sym]; e.Natural && e.Tier <= tier && !known(sym) {
				out = append(out, sym)
			}
		}
		slices.SortFunc(out, func(a, b string) int { return c.element[a].Z - c.element[b].Z })
		return out
	}
	c.synthesizable = func(tier int, known func(string) bool) []string {
		var out []string
		for _, e := range c.elements {
			if !e.Natural && e.Tier <= tier && !known(e.Symbol) {
				out = append(out, e.Symbol)
			}
		}
		return out
	}
	c.newGeology = func(*world.Map) geology { return &fakeGeo{at: map[[2]int][]chem.Source{}} }
	return c
}

// fakeWorld is an empty fake-chemistry world plus a walkable tile far from water.
func fakeWorld(t *testing.T) (*Sim, *fakeGeo, int, int) {
	t.Helper()
	s := newSimWith(testMap(t, 48), 1, fakeCatalog())
	for _, i := range s.terrain.walkable {
		x, y := int(i)%s.terrain.w, int(i)/s.terrain.w
		ok := true
		for dy := -4; dy <= 4 && ok; dy++ {
			for dx := -4; dx <= 4; dx++ {
				if j, in := s.terrain.index(x+dx, y+dy); !in || s.terrain.blocked[j] || s.terrain.nearWater[j] {
					ok = false
					break
				}
			}
		}
		if ok {
			return s, s.geo.(*fakeGeo), x, y
		}
	}
	t.Fatal("no open area on the test map")
	return nil, nil, 0, 0
}

// person puts an adult with a blank brain at the centre of tile (x, y).
func person(s *Sim, sex Sex, name string, x, y int) *Creature {
	g := emptyGenome(initialHidden)
	g.Traits = Traits{Hue: 120, Size: 1, MaxSpeed: 2, Vision: 5, Metabolism: 1, Lifespan: 500, MutationRate: 0.05}
	return s.spawnAdult(g, sex, name, float64(x)+0.5, float64(y)+0.5)
}

func lastEvent(s *Sim) string {
	if len(s.events) == 0 {
		return ""
	}
	return s.events[len(s.events)-1].Text
}

// --- brain ------------------------------------------------------------------

func TestBrainShapes(t *testing.T) {
	if len(InputLabels) != NumInputs || NumInputs != 71 {
		t.Fatalf("got %d input labels, NumInputs=%d", len(InputLabels), NumInputs)
	}
	if len(OutputLabels) != NumOutputs || NumOutputs != 15 {
		t.Fatalf("got %d output labels", len(OutputLabels))
	}
	checks := map[int]string{
		inBias: "bias", inNoise: "acak (kehendak)", inFood + 2: "makanan 0°", inResource: "sumber daya -60°",
		inEnergy: "energi", inHealth: "kesehatan", inCanBuild: "bisa membangun", inOtherHouseNear: "rumah orang lain dekat",
		inTeacherNear: "guru dekat", inStudentNear: "murid dekat", inBestSkill: "keahlian tertinggi",
		inReward: "imbalan terakhir", inLibraryNear: "perpustakaan dekat", inClock: "jam internal",
		inAnimal: "hewan -60°", inSeason: "musim", inLight: "cahaya", inCanPlant: "bisa menanam",
		inCropReady: "tanaman siap panen", inLivestockHungry: "ternak lapar", inPredatorNear: "pemangsa dekat",
	}
	for i, want := range checks {
		if InputLabels[i] != want {
			t.Errorf("input %d = %q, want %q", i, InputLabels[i], want)
		}
	}
	if OutputLabels[outGather] != "kumpulkan" || OutputLabels[outAttack] != "serang" || OutputLabels[outTeach] != "ajar" ||
		OutputLabels[outPlant] != "tanam" || OutputLabels[outHunt] != "buru" {
		t.Fatalf("outputs out of order: %v", OutputLabels)
	}
}

func TestThink(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	g := randomGenome(rng)
	var in [NumInputs]float64
	hidden := make([]float64, g.Hidden)
	var out [NumOutputs]float64
	for i := range in {
		in[i] = rng.Float64()*2 - 1
	}
	g.think(&in, hidden, &out)
	for h, v := range hidden {
		if v < -1 || v > 1 {
			t.Fatalf("hidden[%d]=%v outside tanh range", h, v)
		}
	}
	if out[outTurn] < -1 || out[outTurn] > 1 {
		t.Fatalf("turn %v outside [-1,1]", out[outTurn])
	}
	for o := outMove; o < NumOutputs; o++ {
		if out[o] <= 0 || out[o] >= 1 {
			t.Fatalf("output %d = %v outside (0,1)", o, out[o])
		}
	}

	// The recurrent state makes the same input produce a different hidden state.
	before := slices.Clone(hidden)
	g.think(&in, hidden, &out)
	if slices.Equal(before, hidden) {
		t.Fatal("recurrent hidden state had no effect")
	}

	// The first humans start peaceful.
	var calm [NumInputs]float64
	calm[inBias] = 1
	clear(hidden)
	g.think(&calm, hidden, &out)
	if out[outAttack] > 0.5 || out[outSteal] > 0.5 {
		t.Fatalf("a fresh brain wants to attack (%.2f) or steal (%.2f)", out[outAttack], out[outSteal])
	}
}

func TestCrossoverAndMutation(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	a, b := randomGenome(rng), randomGenome(rng)
	c := crossover(a, b, rng)

	fromA, fromB := 0, 0
	for h := range c.Hidden {
		col := func(g *Genome) []float32 {
			var v []float32
			for i := range NumInputs {
				v = append(v, g.WIn[i*g.Hidden+h])
			}
			return append(v, g.WOut[h*NumOutputs:(h+1)*NumOutputs]...)
		}
		switch {
		case slices.Equal(col(c), col(a)):
			fromA++
		case slices.Equal(col(c), col(b)):
			fromB++
		default:
			t.Fatalf("hidden neuron %d mixes parents", h)
		}
	}
	if fromA == 0 || fromB == 0 {
		t.Fatalf("expected neurons from both parents, got %d/%d", fromA, fromB)
	}

	before := c.clone()
	c.Traits.MutationRate = 0.2
	c = c.mutate(rng)
	if c.Hidden == before.Hidden && slices.Equal(before.WIn, c.WIn) {
		t.Fatal("mutation changed no input weights")
	}
	for _, w := range c.WIn {
		if w < -4 || w > 4 {
			t.Fatalf("weight %v escaped clamp", w)
		}
	}
	tr := c.Traits
	if tr.Size < 0.7 || tr.Size > 1.3 || tr.Vision < 3 || tr.Vision > 9 || tr.Hue < 0 || tr.Hue >= 360 {
		t.Fatalf("traits out of range: %+v", tr)
	}
}

// --- life cycle ---------------------------------------------------------------

// willingGenome sits still and always wants to mate.
func willingGenome(rng *rand.Rand) *Genome {
	g := randomGenome(rng)
	clear(g.WOut)
	clear(g.BOut)
	g.BOut[outMove], g.BOut[outRest], g.BOut[outMate] = -20, -20, 20
	for _, o := range []int{outGather, outCraft, outBuild, outGive, outSteal, outAttack} {
		g.BOut[o] = -20
	}
	return g
}

func TestReproductionAndMarriage(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	f := s.spawnAdult(willingGenome(s.rng), Female, "Hawa", float64(x)+0.5, float64(y)+0.5)
	m := s.spawnAdult(willingGenome(s.rng), Male, "Adam", float64(x)+0.6, float64(y)+0.5)
	home := s.addStructure(s.cat.structure["gubuk"], x, y, f)
	s.moveIn(f, home)

	for range int(gestation*TicksPerSecond) + 10 {
		s.step()
	}

	var child *Creature
	for _, c := range s.creatures {
		if c.Mother != nil && c.Mother.ID == f.ID {
			child = c
		}
	}
	if child == nil {
		t.Fatalf("no child born; births=%d, mother pregnant=%v", s.births, f.Pregnancy != nil)
	}
	if child.Father == nil || child.Father.ID != m.ID || child.Generation != 1 {
		t.Fatalf("bad lineage: father=%+v gen=%d", child.Father, child.Generation)
	}
	if f.Children == 0 || m.Children == 0 {
		t.Fatalf("parents' child counts not updated: %d/%d", f.Children, m.Children)
	}
	if f.Spouse == nil || f.Spouse.ID != m.ID || m.Spouse == nil || m.Spouse.ID != f.ID {
		t.Fatal("parents did not become spouses")
	}
	if m.HouseID != home.ID || child.HouseID != home.ID {
		t.Fatalf("husband and child should live in Hawa's house: %d/%d, want %d", m.HouseID, child.HouseID, home.ID)
	}
	if !s.kin(child, m) || !s.kin(f, m) {
		t.Fatal("family not recognised as kin")
	}
}

func TestGenesisStartsWithAdamAndHawa(t *testing.T) {
	s := New(testMap(t, 48), 21)
	if len(s.creatures) != 2 || s.era != 1 {
		t.Fatalf("got %d creatures in era %d, want 2 in era 1", len(s.creatures), s.era)
	}
	adam, hawa := s.creatures[0], s.creatures[1]
	if adam.Name != "Adam" || adam.Sex != Male || hawa.Name != "Hawa" || hawa.Sex != Female {
		t.Fatalf("first couple is %s/%v and %s/%v", adam.Name, adam.Sex, hawa.Name, hawa.Sex)
	}
	if adam.Mother != nil || hawa.Father != nil || adam.Generation != 0 || !s.adult(adam) || !s.adult(hawa) {
		t.Fatal("first couple should be parentless adults of generation 0")
	}
	if d := math.Hypot(adam.X-hawa.X, adam.Y-hawa.Y); d > 1.5 {
		t.Fatalf("Adam and Hawa start %.1f tiles apart", d)
	}

	// Nobody else ever arrives while someone is alive...
	hawa.Energy = -1
	for range 200 {
		s.step()
	}
	if len(s.creatures) != 1 || s.era != 1 {
		t.Fatalf("after Hawa died: %d creatures in era %d, want only Adam in era 1", len(s.creatures), s.era)
	}
	// ...but after an extinction a new Adam and Hawa begin the next era.
	s.creatures[0].Energy = -1
	s.step()
	if len(s.creatures) != 2 || s.era != 2 || s.creatures[0].Name != "Adam" || s.creatures[1].Name != "Hawa" {
		t.Fatalf("after extinction: %d creatures in era %d", len(s.creatures), s.era)
	}
}

// --- economy ----------------------------------------------------------------------

func gatherOne(s *Sim, c *Creature) bool {
	before := c.Inventory.count()
	for range 100 {
		if !s.gather(c) {
			return false
		}
		if c.Inventory.count() > before {
			return true
		}
	}
	return false
}

func TestGatherNeedsToolsAndDiscovers(t *testing.T) {
	s, geo, x, y := fakeWorld(t)
	geo.put(x, y, "kayu", 5, false)
	geo.put(x+1, y, "bijih", 5, true)
	s.computeInterest()
	s.interest["bijih"] = 1 // rarer than wood
	c := person(s, Male, "Budi", x, y)
	c.Inventory.add(chem.Food, 4) // not hungry for food from the ground

	if !gatherOne(s, c) || c.Inventory["kayu"] != 1 || c.Inventory["bijih"] != 0 {
		t.Fatalf("without a tool expected wood, got %v", c.Inventory)
	}
	if !s.known("C") || !strings.Contains(lastEvent(s), "Karbon (C) oleh Budi dari Kayu") {
		t.Fatalf("gathering wood should reveal carbon; last event %q", lastEvent(s))
	}

	c.Inventory.add("beliung", 1)
	c.HouseID = 1 // housed: wood is no longer wanted for a first home
	s.learn(c, "alat")
	s.interest["kayu"] = 0.2
	c.Want = nil
	if !gatherOne(s, c) || c.Inventory["bijih"] != 1 {
		t.Fatalf("with a pick expected ore, got %v", c.Inventory)
	}

	c.Inventory.add("batu", invCapacity)
	if s.gather(c) {
		t.Fatal("a full creature should not gather")
	}
}

func TestCraftingTeachesAndUsesStations(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Female, "Sari", x, y)
	c.Inventory = Stock{"serat": 2, "batu": 1, "kayu": 1}

	j := s.chooseCraft(c)
	if j == nil || j.Recipe != "tali" {
		t.Fatalf("expected to twist rope first, got %+v", j)
	}
	s.finishJob(c, j)
	if c.Inventory["tali"] != 1 || c.Inventory["serat"] != 0 {
		t.Fatalf("rope not made: %v", c.Inventory)
	}

	j = s.chooseCraft(c)
	if j == nil || j.Recipe != "beliung" {
		t.Fatalf("expected the tech-teaching pick next, got %+v", j)
	}
	s.finishJob(c, j)
	if c.Inventory["beliung"] != 1 || !s.techKnown("alat") || c.Deeds.Crafted != 2 {
		t.Fatalf("pick/tech not made: %v known=%v", c.Inventory, s.techKnown("alat"))
	}

	// Smelting needs a furnace within reach; at one, experimenting comes first.
	c.Inventory = Stock{"bijih": 3}
	if j := s.chooseCraft(c); j != nil {
		t.Fatalf("no station yet, but chose %+v", j)
	}
	s.addStructure(s.cat.structure["tungku"], x+1, y, c)
	s.recomputeTier()
	s.learn(c, "lebur")
	j = s.chooseCraft(c)
	if j == nil || j.Kind != "experiment" || j.Item != "bijih" {
		t.Fatalf("expected an experiment on ore, got %+v", j)
	}
	s.finishJob(c, j)
	if !s.known("Fe") || s.known("H") || c.Inventory["bijih"] != 2 {
		t.Fatalf("experiment at tier 1 should reveal Fe only: Fe=%v H=%v inv=%v", s.known("Fe"), s.known("H"), c.Inventory)
	}
	s.moveIn(c, s.addStructure(s.cat.structure["gubuk"], x-1, y, c))
	c.Want = nil // housed, iron is now wanted for the reactor
	j = s.chooseCraft(c)
	if j == nil || j.Recipe != "lebur" {
		t.Fatalf("expected smelting once nothing is left to discover, got %+v", j)
	}
}

func TestSynthesisAtReactor(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Male, "Dimas", x, y)
	c.Inventory = Stock{chem.SynthesisFuel: 1}
	s.addStructure(s.cat.structure["reaktor"], x, y+1, c)
	s.recomputeTier()
	if j := s.chooseCraft(c); j != nil {
		t.Fatalf("working a reactor needs its know-how, but chose %+v", j)
	}
	s.learn(c, "lebur")
	j := s.chooseCraft(c)
	if j == nil || j.Kind != "synthesis" {
		t.Fatalf("expected synthesis, got %+v", j)
	}
	s.finishJob(c, j)
	if d := s.elements["Tc"]; d == nil || d.Source != synthesisSource || c.Inventory.count() != 0 {
		t.Fatalf("Tc not synthesised from uranium: %+v %v", d, c.Inventory)
	}
}

func TestBuildingHomesAndStations(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Male, "Adam", x, y)
	wife := person(s, Female, "Hawa", x, y)
	s.bond(wife, c)
	c.Inventory = Stock{"kayu": 6, "batu": 2}

	p := s.chooseBuild(c)
	if p == nil || p.job == nil || p.job.Structure != "gubuk" || p.job.X != x || p.job.Y != y {
		t.Fatalf("expected a hut on the current tile, got %+v", p)
	}
	s.finishJob(c, p.job)
	home := s.houseOf(c)
	if home == nil || home.Owner != c.ID || wife.HouseID != home.ID || c.Deeds.Built != 1 {
		t.Fatalf("hut not built or family not moved in: %+v wife=%d", home, wife.HouseID)
	}
	if s.flags(c)&flagHead == 0 || s.flags(wife)&flagHead != 0 {
		t.Fatal("only the builder is kepala keluarga")
	}

	// Neighbours keep their distance.
	other := person(s, Male, "Budi", x+1, y)
	other.Inventory = Stock{"kayu": 2}
	if p := s.chooseBuild(other); p != nil {
		t.Fatalf("a second house right next door: %+v", p)
	}

	// With wood left over and stone, the station comes next and raises the tier.
	p = s.chooseBuild(c)
	if p == nil || p.job == nil || p.job.Structure != "tungku" {
		t.Fatalf("expected a furnace, got %+v", p)
	}
	s.finishJob(c, p.job)
	if s.tier != 1 || !s.techKnown("lebur") || !strings.Contains(lastEvent(s), "Zaman 1 dimulai!") {
		t.Fatalf("tier %d, lebur known %v, last event %q", s.tier, s.techKnown("lebur"), lastEvent(s))
	}

	// Learning the tool tech unlocks the house upgrade, done in place.
	s.learn(c, "alat")
	c.Inventory.add("kayu", 4)
	p = s.chooseBuild(c)
	if p == nil || p.job == nil || p.job.Kind != "upgrade" {
		t.Fatalf("expected an upgrade, got %+v", p)
	}
	s.finishJob(c, p.job)
	if home.Kind != "rumah" || home.kind.Storage != 20 {
		t.Fatalf("house not upgraded: %s", home.Kind)
	}
}

func TestKeepHouse(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Female, "Sari", x, y)
	s.moveIn(c, s.addStructure(s.cat.structure["gubuk"], x, y, c))
	c.Inventory = Stock{"kayu": 5, chem.Food: 5, "beliung": 1, "pedang": 1}
	s.keepHouse(c)
	h := s.houseOf(c)
	if c.Inventory[chem.Food] != foodReserve || c.Inventory["beliung"] != 1 || c.Inventory["pedang"] != 1 || c.Inventory["kayu"] != 0 {
		t.Fatalf("kept the wrong things: %v", c.Inventory)
	}
	if h.Storage["kayu"] != 5 || h.Storage[chem.Food] != 2 {
		t.Fatalf("storage %v", h.Storage)
	}
	c.Inventory.take(chem.Food, foodReserve)
	c.Energy = 0.3
	s.keepHouse(c)
	if c.Inventory[chem.Food] != 2 {
		t.Fatalf("hungry at home should take food from storage, has %v", c.Inventory)
	}
}

// --- social -------------------------------------------------------------------------

func TestGiving(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	mother := person(s, Female, "Ibu", x, y)
	child := person(s, Male, "Anak", x, y)
	child.Mother = &Ref{mother.ID, mother.Name}
	child.Energy = 0.2
	stranger := person(s, Male, "Orang", x, y)
	stranger.Energy = 0.9
	mother.Inventory = Stock{chem.Food: 1}
	s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures)

	if !s.give(mother) || child.Inventory[chem.Food] != 1 || stranger.Inventory.count() != 0 {
		t.Fatalf("food should go to the hungry child: child=%v stranger=%v", child.Inventory, stranger.Inventory)
	}
	if mother.Deeds.Kindness != 1 || s.kindness != 1 || mother.Reputation <= 0 {
		t.Fatal("kindness not recorded")
	}
	if s.give(mother) {
		t.Fatal("nothing left to give")
	}
}

func TestStealing(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	thief := person(s, Male, "Maling", x, y)
	brother := person(s, Male, "Saudara", x, y)
	victim := person(s, Female, "Korban", x, y)
	thief.Mother = &Ref{999, "Ibu"}
	brother.Mother = &Ref{999, "Ibu"}
	brother.Inventory = Stock{"besi": 3}
	victim.Inventory = Stock{"kayu": 2, "besi": 3}
	s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures)

	if !s.steal(thief) {
		t.Fatal("theft failed")
	}
	if thief.Inventory["besi"] != 2 || victim.Inventory["besi"] != 1 || brother.Inventory["besi"] != 3 {
		t.Fatalf("should take the most valuable item from the stranger only: thief=%v victim=%v", thief.Inventory, victim.Inventory)
	}
	if victim.Hurt <= 0 || victim.Offender == nil || victim.Offender.ID != thief.ID || thief.Reputation >= 0 || s.crimes != 1 {
		t.Fatal("theft not recorded")
	}
	if !strings.Contains(lastEvent(s), "Maling mencuri 2 Besi dari Korban") {
		t.Fatalf("event %q", lastEvent(s))
	}

	// Another family's storage is fair game for a thief too.
	victim.Inventory = nil
	victim.X += 10
	house := s.addStructure(s.cat.structure["gubuk"], x+1, y, victim)
	house.Storage = Stock{"kayu": 4}
	s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures)
	if !s.steal(thief) || house.Storage["kayu"] != 2 {
		t.Fatalf("house not robbed: %v", house.Storage)
	}
}

func TestAttackAndKill(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	killer := person(s, Male, "Kain", x, y)
	victim := person(s, Male, "Habil", x, y)
	killer.Inventory = Stock{"pedang": 1}
	victim.Inventory = Stock{"besi": 2}
	s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures)

	blows := 0
	for victim.Health > 0 && blows < 20 {
		if !s.attack(killer) {
			t.Fatal("no one to attack")
		}
		blows++
	}
	if blows != 4 { // 0.15 × size 1 × (1 + sword 1) = 0.3 per blow
		t.Fatalf("took %d blows", blows)
	}
	if killer.Deeds.Kills != 1 || killer.Inventory["besi"] != 2 || killer.Reputation > -0.5 {
		t.Fatalf("killer not credited: %+v %v", killer.Deeds, killer.Inventory)
	}
	s.reap()
	if s.byID[victim.ID] != nil || s.deathsBy.Killed != 1 || !strings.Contains(lastEvent(s), "Habil ♂ dibunuh oleh Kain ♂") {
		t.Fatalf("victim not dead or event wrong: %q", lastEvent(s))
	}
}

func TestInheritance(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	head := person(s, Male, "Ayah", x, y)
	wife := person(s, Female, "Ibu", x, y)
	s.bond(wife, head)
	home := s.addStructure(s.cat.structure["gubuk"], x, y, head)
	s.moveIn(head, home)
	old := person(s, Female, "Sulung", x, y)
	young := person(s, Male, "Bungsu", x, y)
	for _, c := range []*Creature{old, young} {
		c.Mother, c.Father, c.HouseID = &Ref{wife.ID, wife.Name}, &Ref{head.ID, head.Name}, home.ID
	}
	young.BornTick = old.BornTick + 100

	head.Energy = 0
	s.reap()
	if home.Owner != wife.ID || !strings.Contains(lastEvent(s), "diwarisi oleh Ibu") {
		t.Fatalf("the widow should inherit: owner %d, event %q", home.Owner, lastEvent(s))
	}
	wife.Energy = 0
	s.reap()
	if home.Owner != old.ID {
		t.Fatalf("the eldest child should inherit next, owner %d", home.Owner)
	}
	old.Energy, young.Energy = 0, 0
	s.reap()
	if home.Owner != 0 {
		t.Fatalf("with no heir the house is abandoned, owner %d", home.Owner)
	}

	stranger := person(s, Female, "Pendatang", x, y)
	if !s.startBuild(stranger) && stranger.HouseID != home.ID {
		t.Fatal("a homeless creature should move into the empty house")
	}
	if home.Owner != stranger.ID {
		t.Fatalf("claim failed, owner %d", home.Owner)
	}
}

// --- persistence and stream -------------------------------------------------------

func TestPersistRoundTrip(t *testing.T) {
	m := testMap(t, 48)
	a := New(m, 11)
	for range 437 { // not a multiple of any refresh interval
		a.step()
	}
	c := a.creatures[0]
	c.Inventory.add("kayu", 3)
	k := a.cat.structure["gubuk"]
	a.moveIn(c, a.addStructure(k, int(c.X), int(c.Y), c))
	a.houseOf(c).Storage.add("batu", 2)
	a.discover(c, "Au", "Bijih Emas")
	c.Skills = map[string]float64{"api": 0.7, "tulisan": 0.4}
	lx, ly, ok := a.spotNear(int(c.X), int(c.Y), 4)
	if !ok {
		t.Fatal("no room for a library")
	}
	lib := a.addStructure(a.cat.structure["perpustakaan"], lx, ly, c)
	lib.Written = map[string]float64{"api": 0.5}
	a.lost["kaca"] = true
	a.knowledgeLost = 1

	data, err := a.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte{0x1f, 0x8b}) {
		t.Fatal("save is not gzip-compressed")
	}
	b, err := Restore(m, data)
	if err != nil {
		t.Fatal(err)
	}
	if b.tick != a.tick || len(b.creatures) != len(a.creatures) || b.nextID != a.nextID {
		t.Fatalf("restored tick/pop/nextID %d/%d/%d, want %d/%d/%d",
			b.tick, len(b.creatures), b.nextID, a.tick, len(a.creatures), a.nextID)
	}
	bc := b.byID[c.ID]
	if bc.Inventory["kayu"] != c.Inventory["kayu"] || b.houseOf(bc) == nil || b.houseOf(bc).Storage["batu"] != 2 || !b.known("Au") {
		t.Fatal("inventory, house or discovery lost in the save")
	}
	if bc.Skills["api"] != 0.7 || b.structByID[lib.ID].Written["api"] != 0.5 || !b.lost["kaca"] || b.knowledgeLost != 1 {
		t.Fatal("skills, library or lost knowledge lost in the save")
	}
	for i := range a.creatures {
		ca, cb := a.creatures[i], b.creatures[i]
		if !slices.Equal(ca.Hidden, cb.Hidden) || !slices.Equal(ca.Genome.WIn, cb.Genome.WIn) || !slices.Equal(ca.Mind.DIn, cb.Mind.DIn) {
			t.Fatalf("creature %d brain differs after restore", ca.ID)
		}
	}

	// With the RNG and deposits restored too, both worlds evolve identically.
	for range 200 {
		a.step()
		b.step()
	}
	if !bytes.Equal(a.encodeFrame(), b.encodeFrame()) {
		t.Fatal("restored world diverged")
	}
	sa, _ := a.Structures()
	sb, _ := b.Structures()
	if !bytes.Equal(sa, sb) {
		t.Fatal("restored buildings differ")
	}

	if _, err := Restore(testMap(t, 32), data); err == nil {
		t.Fatal("restore onto a different-sized map should fail")
	}
	if _, err := Restore(m, []byte(`{"version":2}`)); err == nil {
		t.Fatal("an old uncompressed save should be rejected")
	}
}

func TestFrameAndStructuresAreValidJSON(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Male, "Adam", x, y)
	s.moveIn(c, s.addStructure(s.cat.structure["gubuk"], x, y, c))
	s.step()
	var raw struct {
		T int64   `json:"t"`
		S float64 `json:"s"`
		C [][]any `json:"c"`
	}
	if err := json.Unmarshal(s.encodeFrame(), &raw); err != nil {
		t.Fatal(err)
	}
	if raw.T != 1 || len(raw.C) != 1 || len(raw.C[0]) != 12 || raw.C[0][11].(float64) != float64(c.HouseID) {
		t.Fatalf("unexpected frame: %+v", raw)
	}
	if int(raw.C[0][8].(float64))&flagHead == 0 {
		t.Fatal("head flag missing")
	}

	data, v := s.Structures()
	var st struct {
		V int64   `json:"v"`
		S [][]any `json:"s"`
	}
	if err := json.Unmarshal(data, &st); err != nil {
		t.Fatal(err)
	}
	if st.V != v || len(st.S) != 1 || st.S[0][1] != "gubuk" || st.S[0][5] != "Adam" || len(st.S[0]) != 8 {
		t.Fatalf("unexpected structures message: %s", data)
	}
}

func TestUpdateMapRelocatesAndRemovesBuildings(t *testing.T) {
	m := testMap(t, 32)
	s := New(m, 9)
	c := s.creatures[0]
	cx, cy := int(math.Floor(c.X)), int(math.Floor(c.Y))
	s.moveIn(c, s.addStructure(s.cat.structure["gubuk"], cx, cy, c))

	edited := *m
	edited.Layers.Objects = slices.Clone(m.Layers.Objects)
	edited.Layers.Objects[cy*m.Width+cx] = world.Wall
	s.UpdateMap(&edited)
	if s.terrain.blockedAt(c.X, c.Y) {
		t.Fatal("creature left standing inside a wall")
	}
	if len(s.structures) != 0 || c.HouseID != 0 {
		t.Fatal("the house under the new wall should be gone")
	}
}

func TestRealChemistryCatalog(t *testing.T) {
	k := defaultCatalog()
	if len(k.elements) != 118 || len(k.recipes) == 0 || k.foodEnergy() <= 0 {
		t.Fatalf("chem catalog incomplete: %d elements, %d recipes", len(k.elements), len(k.recipes))
	}
	if _, ok := k.structure["gubuk"]; !ok {
		t.Fatal("no first house in chem")
	}
	s := New(testMap(t, 64), 3)
	if len(s.interest) == 0 || slices.Max(s.tileItems) == 0 {
		t.Fatal("geology gave nothing to gather")
	}
}

// --- soak ---------------------------------------------------------------------------

// TestSoak runs the starter map from Adam and Hawa and reports how humanity
// fares. Run several seeds in parallel processes:
//
//	go test ./internal/sim -run Soak -v -soak -soakseed N [-soakminutes 120]
func TestSoak(t *testing.T) {
	if !*soak {
		t.Skip("pass -soak to run")
	}
	m := world.Generate(world.GenOptions{Name: "Starter Island", Width: 128, Height: 128, Seed: 1337})
	m.ID = "0123456789abcdef"
	s := New(m, *soakSeed)
	t.Logf("seed %d: capacity %d, walkable %d", *soakSeed, s.capacity, len(s.terrain.walkable))

	firstHouse, era1Lasted := -1, -1
	for minute := 1; minute <= *soakMinutes; minute++ {
		for range 60 * TicksPerSecond {
			s.step()
		}
		if firstHouse < 0 && len(s.structures) > 0 {
			firstHouse = minute
		}
		if era1Lasted < 0 && s.era > 1 {
			era1Lasted = minute
		}
		if minute%10 == 0 || minute <= 2 {
			f, ml, maxGen, _ := s.census()
			t.Logf("t=%3dm era=%d pop=%3d (♀%d ♂%d) births=%4d deaths=%4d [starve %d thirst %d old %d killed %d] gen=%d houses=%d structs=%d tier=%d elems=%d techs=%d crimes=%d kind=%d kills=%d",
				minute, s.era, len(s.creatures), f, ml, s.births, s.deaths,
				s.deathsBy.Starvation, s.deathsBy.Thirst, s.deathsBy.OldAge, s.deathsBy.Killed,
				maxGen, s.houseCount(), len(s.structures), s.tier, len(s.elements), len(s.techs), s.crimes, s.kindness, s.kills)
		}
	}
	if era1Lasted < 0 {
		era1Lasted = *soakMinutes
	}
	var elems []string
	for sym := range s.elements {
		elems = append(elems, sym)
	}
	slices.Sort(elems)
	var techs []string
	for id := range s.techs {
		techs = append(techs, id)
	}
	slices.Sort(techs)
	kinds := map[string]int{}
	held := Stock{}
	for _, st := range s.structures {
		kinds[st.Kind]++
		for id, n := range st.Storage {
			held.add(id, n)
		}
	}
	for _, c := range s.creatures {
		for id, n := range c.Inventory {
			held.add(id, n)
		}
	}
	t.Logf("SUMMARY seed=%d era1Lasted=%dm eras=%d pop=%d firstHouse=%dm tier=%d elements=%d %v crimes=%d kindness=%d kills=%d",
		*soakSeed, era1Lasted, s.era, len(s.creatures), firstHouse, s.tier, len(s.elements), elems, s.crimes, s.kindness, s.kills)
	t.Logf("DETAIL techs=%v buildings=%v held=%v", techs, kinds, held)
}
