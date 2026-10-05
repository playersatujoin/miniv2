package sim

import (
	"math"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"miniv2/backend/internal/chem"
)

// --- learning brains -----------------------------------------------------------

func TestPlasticityIsBoundedAndOnlyWhenLearning(t *testing.T) {
	for _, off := range []bool{false, true} {
		s, _, x, y := fakeWorld(t)
		s.opts.NoLearning = off
		c := person(s, Female, "Sari", x, y)
		c.Genome = randomGenome(rand.New(rand.NewPCG(3, 3))) // a blank brain has no activity to learn from
		s.giveBrain(c)
		c.Genome.Traits.LearningRate, c.Genome.Traits.TraceDecay, c.Genome.Traits.Plasticity = 0.3, 0.9, 0.5
		rng := rand.New(rand.NewPCG(9, 9))
		for step := range 400 {
			for i := range c.input {
				c.input[i] = rng.Float64()*2 - 1
			}
			think(c.Genome, c.Mind.wIn, c.Mind.wOut, &c.input, c.Hidden, &c.output)
			// Alternate good and bad times so the surprise keeps flipping.
			if step%2 == 0 {
				c.Energy = 1
			} else {
				c.Energy = 0.2
			}
			s.learnStep(c)
		}
		changed := false
		for i, d := range c.Mind.DIn {
			if math.Abs(float64(d)) > 0.5+1e-6 {
				t.Fatalf("learned change %v beyond the plasticity limit", d)
			}
			if w := c.Mind.wIn[i]; w < -maxWeight || w > maxWeight {
				t.Fatalf("effective weight %v escaped the clamp", w)
			}
			changed = changed || d != 0
		}
		if changed == off {
			t.Fatalf("NoLearning=%v but learned anything=%v", off, changed)
		}
		if c.Mind.Reward == 0 {
			t.Fatal("reward was never felt")
		}
	}
}

func TestBrainGrowsNeutrallyAndShrinks(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 6))
	g := randomGenome(rng)
	var in [NumInputs]float64
	for i := range in {
		in[i] = rng.Float64()*2 - 1
	}
	run := func(g *Genome) [NumOutputs]float64 {
		var out [NumOutputs]float64
		h := make([]float64, g.Hidden)
		for range 3 {
			g.think(&in, h, &out)
		}
		return out
	}
	before := run(g)
	big := g.withNeuron(rng)
	if big.Hidden != g.Hidden+1 || !big.valid() {
		t.Fatalf("grown brain has %d neurons, valid=%v", big.Hidden, big.valid())
	}
	if after := run(big); after != before {
		t.Fatalf("a new silent neuron changed behaviour: %v → %v", before, after)
	}
	small := big.withoutNeuron(big.weakestNeuron())
	if small.Hidden != g.Hidden || !small.valid() {
		t.Fatalf("shrunk brain has %d neurons", small.Hidden)
	}

	// Parents of different sizes still make a working child.
	a, b := g, g.clone()
	for range 15 {
		b = b.withNeuron(rng)
	}
	for range 20 {
		c := crossover(a, b, rng)
		if !c.valid() || c.Hidden < a.Hidden || c.Hidden > b.Hidden {
			t.Fatalf("child of %d and %d neurons has %d (valid=%v)", a.Hidden, b.Hidden, c.Hidden, c.valid())
		}
	}

	// Over many mutations sizes stay within the allowed range.
	m := g
	for range 2000 {
		m = m.clone().mutate(rng)
		if m.Hidden < minHidden || m.Hidden > maxHidden || !m.valid() {
			t.Fatalf("mutation produced %d neurons", m.Hidden)
		}
	}
}

func TestBiggerBrainsCostEnergy(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	small := person(s, Female, "Sari", x, y)
	big := person(s, Female, "Ayu", x, y)
	big.Genome = big.Genome.clone()
	for big.Genome.Hidden < maxHidden {
		big.Genome = big.Genome.withNeuron(rand.New(rand.NewPCG(1, 1)))
	}
	s.giveBrain(big)
	for range 100 {
		s.act(small)
		s.act(big)
	}
	if big.Energy >= small.Energy {
		t.Fatalf("a %d-neuron brain cost no more than a %d-neuron one (%.4f vs %.4f)",
			big.Genome.Hidden, small.Genome.Hidden, big.Energy, small.Energy)
	}
}

// --- skills ------------------------------------------------------------------------

func TestSkillsGateWhatOneCanDo(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	inventor := person(s, Female, "Sari", x, y)
	other := person(s, Male, "Budi", x+3, y)
	s.learn(inventor, "alat") // invented: she can practise it
	if !s.techKnown("alat") || !s.canPractise(inventor, "alat") || s.canPractise(other, "alat") {
		t.Fatalf("inventor %.2f, other %.2f", inventor.Skills["alat"], other.Skills["alat"])
	}
	// The house upgrade needs the tool know-how: only she may build it.
	h := s.addStructure(s.cat.structure["gubuk"], x, y, other)
	s.moveIn(other, h)
	if _, ok := s.upgradeFor(other, h); ok {
		t.Fatal("someone without the skill may upgrade")
	}
	other.Skills = map[string]float64{"alat": 0.35}
	if _, ok := s.upgradeFor(other, h); !ok {
		t.Fatal("with the skill the upgrade should be possible")
	}

	// Practice raises a skill; long disuse fades it.
	l := other.Skills["alat"]
	s.practise(other, "alat")
	if other.Skills["alat"] <= l {
		t.Fatal("practice did not help")
	}
	for range 600 {
		s.cultureTick()
	}
	if other.Skills["alat"] >= l {
		t.Fatalf("an unused skill did not fade: %.3f", other.Skills["alat"])
	}

	// In the old model whatever the world knows, everybody can do.
	s.opts.NoLearning = true
	if !s.canPractise(other, "alat") || s.teach(inventor) {
		t.Fatal("NoLearning should make know-how world-wide and teaching a no-op")
	}
}

func TestTeachingPassesSkillsOn(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	teacher := person(s, Female, "Sari", x, y)
	student := person(s, Male, "Budi", x, y)
	student.X += 1
	teacher.Skills = map[string]float64{"alat": 0.9}
	s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures)
	for range 20 * TicksPerSecond {
		if !s.teach(teacher) {
			t.Fatal("teacher found nobody to teach")
		}
		if student.Skills["alat"] >= skillPractise {
			break
		}
	}
	if student.Skills["alat"] < skillPractise {
		t.Fatalf("student only reached %.2f", student.Skills["alat"])
	}
	if student.Teacher == nil || student.Teacher.ID != teacher.ID || !slices.Equal(teacher.Students, []int64{student.ID}) {
		t.Fatalf("teacher %+v students %v", student.Teacher, teacher.Students)
	}
	if !strings.Contains(lastEvent(s), "Budi belajar Alat dari Sari") {
		t.Fatalf("no learning event, last: %q", lastEvent(s))
	}
	if s.flags(teacher)&flagTeaching == 0 {
		t.Fatal("teaching not shown in the stream flags")
	}
}

func TestWatchingTeachesALittle(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	worker := person(s, Female, "Sari", x, y)
	watcher := person(s, Male, "Budi", x+1, y)
	worker.Skills = map[string]float64{"alat": 0.9}
	worker.Job = &Job{Kind: "build", Structure: "rumah", Remaining: 100}
	s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures)
	s.cultureTick()
	if l := watcher.Skills["alat"]; l <= 0 || l > watchRate {
		t.Fatalf("watching one second gave %.4f skill", l)
	}
}

func TestLibrariesKeepKnowledgeAlive(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	lib := chem.StructureKind{ID: "perpustakaan", Name: "Perpustakaan", Teaches: "tulisan", Library: true}
	s.cat = cloneCatalogWith(s.cat, lib, chem.Tech{ID: "tulisan", Name: "Tulisan"})
	writer := person(s, Female, "Sari", x, y)
	writer.Skills = map[string]float64{"alat": 0.9, literacyTech: 0.6}
	s.techs["alat"] = &Discovery{By: Ref{writer.ID, writer.Name}}
	st := s.addStructure(lib, x+1, y, writer)
	s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures)
	for range 30 * TicksPerSecond {
		if !s.teach(writer) { // nobody to teach: she writes
			t.Fatal("writer neither taught nor wrote")
		}
	}
	if st.Written["alat"] < skillPractise || s.writtenLevel("alat") < skillPractise {
		t.Fatalf("library holds %.2f of alat", st.Written["alat"])
	}

	// The writer dies; the know-how is not lost because it is written down.
	writer.Energy = 0
	s.reap()
	s.countHolders()
	if s.lost["alat"] {
		t.Fatal("written knowledge was marked lost")
	}

	// A literate reader can learn it back.
	reader := person(s, Male, "Budi", x, y)
	reader.Skills = map[string]float64{literacyTech: 0.5}
	for range 200 {
		s.cultureTick()
	}
	if reader.Skills["alat"] < skillPractise {
		t.Fatalf("reader only learned %.2f", reader.Skills["alat"])
	}
}

func TestKnowledgeIsLostAndRediscovered(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	sari := person(s, Female, "Sari", x, y)
	person(s, Male, "Budi", x+4, y) // someone has to outlive her
	s.learn(sari, "alat")
	s.countHolders()
	sari.Energy = 0
	s.reap()
	s.countHolders()
	if !s.lost["alat"] || s.knowledgeLost != 1 || !strings.Contains(lastEvent(s), "Pengetahuan hilang: Alat — pemegang terakhir, Sari, meninggal") {
		t.Fatalf("lost=%v count=%d last=%q", s.lost["alat"], s.knowledgeLost, lastEvent(s))
	}
	view := s.Knowledge()
	for _, tv := range view.Techs {
		if tv.ID == "alat" && (!tv.Known || !tv.Lost || tv.Holders != 0) {
			t.Fatalf("knowledge view: %+v", tv)
		}
	}
	budi := s.creatures[0]
	budi.Skills = map[string]float64{"alat": 0.25}
	s.practise(budi, "alat")
	if s.lost["alat"] || !strings.Contains(lastEvent(s), "ditemukan kembali") {
		t.Fatalf("re-learning not noticed: lost=%v last=%q", s.lost["alat"], lastEvent(s))
	}
}

func TestViolentEventsAreFlagged(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	a := person(s, Male, "Adam", x, y)
	b := person(s, Male, "Budi", x, y)
	b.X += 0.5
	s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures)
	if !s.attack(a) {
		t.Fatal("no attack")
	}
	e := s.events[len(s.events)-1]
	if e.Kind != "crime" || !e.Violent {
		t.Fatalf("assault event %+v", e)
	}
	crimes, _, _ := s.stats.deedRates()
	if crimes <= 0 {
		t.Fatal("assault not counted in the yearly rates")
	}
}

// cloneCatalogWith adds a structure kind and a tech to a test catalog.
func cloneCatalogWith(c *catalog, k chem.StructureKind, tech chem.Tech) *catalog {
	n := newCatalog(nil, c.recipes, append(slices.Clone(c.structures), k), append(slices.Clone(c.techs), tech), c.elements)
	n.items, n.itemIndex, n.itemIDs = c.items, c.itemIndex, c.itemIDs
	n.tierName, n.discoverable, n.synthesizable, n.newGeology = c.tierName, c.discoverable, c.synthesizable, c.newGeology
	return n
}

// BenchmarkStep150 measures one tick of a world with 150 grown-ups who know
// a few crafts. Run: go test ./internal/sim -bench Step150 -run X
func BenchmarkStep150(b *testing.B) {
	for _, off := range []bool{false, true} {
		name := "learning"
		if off {
			name = "no-learning"
		}
		b.Run(name, func(b *testing.B) {
			m := testMap(&testing.T{}, 128)
			s := NewWithOptions(m, 3, Options{NoLearning: off})
			for i := range 150 {
				x, y, _ := s.terrain.randomSpot(s.rng)
				c := s.spawnAdult(randomGenome(s.rng), Sex(i%2), "Orang", x, y)
				c.Skills = map[string]float64{"api": 0.6, "tembikar": 0.5, "alat_batu": 0.4}
			}
			b.ResetTimer()
			for range b.N {
				s.step()
			}
			b.ReportMetric(float64(len(s.creatures)), "creatures")
		})
	}
}
