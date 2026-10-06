package sim

import (
	"math"
	"math/bits"
	"testing"
)

// Two carriers have children in the Mendelian ratio 1 : 2 : 1, and every
// locus assorts on its own.
func TestMendelianSegregation(t *testing.T) {
	const locus, other = 4, 9
	carrier := &Loci{Mat: 1<<locus | 1<<other}
	const n = 40000
	var counts [3]int
	together := 0
	for i := range n {
		r := newGeneRand(int64(i+1), 77, saltGamete)
		child := Loci{Mat: gamete(carrier, &r), Pat: gamete(carrier, &r)}
		copies := int(child.Mat>>locus&1 + child.Pat>>locus&1)
		counts[copies]++
		if child.Mat>>locus&1 == child.Mat>>other&1 {
			together++
		}
	}
	for copies, want := range []float64{0.25, 0.5, 0.25} {
		got := float64(counts[copies]) / n
		// ±4 standard errors of a binomial share.
		if tol := 4 * math.Sqrt(want*(1-want)/n); math.Abs(got-want) > tol {
			t.Fatalf("%d copies in %.4f of children, want %.2f (counts %v)", copies, got, want, counts)
		}
	}
	// Unlinked loci: a gamete carries both variants or neither half the time.
	if got := float64(together) / n; math.Abs(got-0.5) > 0.02 {
		t.Fatalf("loci %d and %d inherited together in %.3f of gametes", locus, other, got)
	}
}

func TestMutationAddsVariantsRarely(t *testing.T) {
	healthy := &Loci{}
	const n = 20000
	mutations := 0
	for i := range n {
		r := newGeneRand(int64(i+1), 5, saltGamete)
		mutations += bits.OnesCount32(gamete(healthy, &r))
	}
	want := n * numLoci * geneMutation
	if math.Abs(float64(mutations)-want) > 5*math.Sqrt(want) {
		t.Fatalf("%d new variants in %d gametes, want about %.0f", mutations, n, want)
	}
}

func TestFoundersCarryButDontSuffer(t *testing.T) {
	carried, total := 0, 400
	for i := range total {
		r := newGeneRand(int64(i), 0, saltFounder)
		l := founderLoci(&r)
		if l.expressed() != 0 {
			t.Fatal("a founder starts with a recessive disorder")
		}
		carried += bits.OnesCount32(l.carried())
	}
	// About 1.6 variants per founder (see defectClasses).
	expect := 0.0
	for _, d := range defectClasses {
		expect += 2 * float64(d.loci) * d.founderQ
	}
	if mean := float64(carried) / float64(total); math.Abs(mean-expect) > 0.4 {
		t.Fatalf("founders carry %.2f variants each, want about %.2f", mean, expect)
	}
	// A real world's first couple gets genes and a place in the pedigree,
	// and different worlds differ.
	s1, s2 := New(testMap(t, 48), 1), New(testMap(t, 48), 2)
	for _, s := range []*Sim{s1, s2} {
		for _, c := range s.creatures {
			if c.Genome.Loci == nil || s.genetics.ped.entry(c.ID) == nil || c.Genome.Loci.expressed() != 0 {
				t.Fatalf("founder %s without genes or pedigree entry", c.Name)
			}
		}
	}
	if *s1.creatures[0].Genome.Loci == *s2.creatures[0].Genome.Loci && *s1.creatures[1].Genome.Loci == *s2.creatures[1].Genome.Loci {
		t.Fatal("two worlds' founders have the same genes")
	}
}

// couple makes a man and a woman with the given genotypes, she pregnant by him.
func couple(t *testing.T, mat, pat Loci) (*Sim, *Creature, *Creature, *Pregnancy) {
	s, _, x, y := fakeWorld(t)
	m := person(s, Female, "Ibu", x, y)
	f := person(s, Male, "Ayah", x, y)
	m.Genome.Loci, f.Genome.Loci = &mat, &pat
	s.genetics.ped.add(m, 0)
	s.genetics.ped.add(f, 0)
	p := &Pregnancy{Father: Ref{f.ID, f.Name}, FatherGenome: f.Genome}
	return s, m, f, p
}

func TestLethalDisordersKillSomeNewborns(t *testing.T) {
	const lethal = 0 // the first locus is of the lethal class
	both := Loci{Mat: 1 << lethal, Pat: 1 << lethal}
	s, m, _, p := couple(t, both, both)
	died, n := 0, 2000
	for range n {
		c := s.newChild(m, p)
		if c.Genome.Loci.expressed()&(1<<lethal) == 0 {
			t.Fatal("two homozygous parents had a child without the disorder")
		}
		if c.fate == "neonatal" {
			died++
		}
	}
	if got := float64(died) / float64(n); math.Abs(got-locusDeath[lethal]) > 0.04 {
		t.Fatalf("%.3f of affected newborns died, want %.2f", got, locusDeath[lethal])
	}
	if s.genetics.stats.Lethal != died || s.genetics.stats.ByF[0].Lethal != died {
		t.Fatalf("statistics count %d lethal births, want %d", s.genetics.stats.Lethal, died)
	}
	// Without genetics nobody dies of it, but the genes are still inherited.
	s.opts.NoGenetics = true
	for range 200 {
		if c := s.newChild(m, p); c.fate != "" || c.Genome.Loci.expressed() == 0 {
			t.Fatal("a recessive disorder killed or vanished with genetics off")
		}
	}
}

func TestDisordersWeakenTheBody(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	c := person(s, Female, "Sakit", x, y)
	well := person(s, Female, "Sehat", x, y)
	firstOf := func(class int) int {
		for i, k := range locusClass {
			if int(k) == class {
				return i
			}
		}
		t.Fatalf("no locus of class %d", class)
		return 0
	}
	frail, immune, subfertile := firstOf(1), firstOf(2), firstOf(3)
	hom := uint32(1<<frail | 1<<immune | 1<<subfertile)
	c.Genome.Loci = &Loci{Mat: hom | 1<<thalLocus, Pat: hom}
	well.Genome.Loci = &Loci{Mat: 1 << frail} // a carrier only

	if got := s.vigor(c); !same(got, defectClasses[1].vigor) {
		t.Fatalf("vigor with weak muscles %.3f", got)
	}
	if s.vigor(well) != 1 || s.fertilityGene(well) != 1 || s.diseaseGene(well, Diarrhea) != 1 {
		t.Fatal("a healthy carrier suffers")
	}
	if got := s.diseaseGene(c, Diarrhea); !same(got, defectClasses[2].severity) {
		t.Fatalf("diarrhoea severity ×%.2f with weak immunity", got)
	}
	// Thalassaemia carriers get milder malaria.
	if got := s.diseaseGene(c, Malaria); !same(got, defectClasses[2].severity*thalCarrierMalaria) {
		t.Fatalf("malaria severity ×%.2f for an immune-deficient thalassaemia carrier", got)
	}
	if got := s.fertilityGene(c); !same(got, defectClasses[3].fertility) {
		t.Fatalf("fertility ×%.2f", got)
	}
	c.Energy, well.Energy = 1, 1
	if a, b := s.conceptionChance(c), s.conceptionChance(well); !same(a, b*defectClasses[3].fertility) || b == 0 {
		t.Fatalf("conception chance %.3f vs %.3f", a, b)
	}
	view := s.geneticsView(c)
	if len(view.Defects) != 3 || view.Carried != 1 || !view.MalariaShield || !view.Known {
		t.Fatalf("genetics view %+v", view)
	}
	s.opts.NoGenetics = true
	if s.vigor(c) != 1 || s.fertilityGene(c) != 1 || s.diseaseGene(c, Malaria) != 1 {
		t.Fatal("disorders still harm with genetics off")
	}
}

// People from before Fase 3c (no loci) are healthy carriers, the same ones
// every time; their children get real genes.
func TestUnknownGenesAreDrawnFromTheId(t *testing.T) {
	s, m, f, p := couple(t, Loci{}, Loci{})
	m.Genome.Loci, f.Genome.Loci = nil, nil
	a, b := s.lociOf(m.Genome, m.ID), s.lociOf(m.Genome, m.ID)
	if *a != *b || a.expressed() != 0 {
		t.Fatal("unknown genes not stable or not healthy")
	}
	if s.vigor(m) != 1 || s.geneticsView(m).Known {
		t.Fatal("unknown genes treated as known")
	}
	c := s.newChild(m, p)
	if c.Genome.Loci == nil {
		t.Fatal("a child of unknown genes has none")
	}
	if extra := c.Genome.Loci.Mat &^ (a.Mat | a.Pat); bits.OnesCount32(extra) > 1 {
		t.Fatal("the child's maternal copy doesn't come from the mother")
	}
}

// The mate senses: how related the nearest possible mate in sight is, and
// how healthy they look; both 0 with nobody there or genetics off.
func TestMateSenses(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	mother := person(s, Female, "Ibu", x-3, y-3)
	father := person(s, Male, "Ayah", x-3, y+3)
	stranger := person(s, Female, "Asing", x-3, y)
	for _, c := range []*Creature{mother, father, stranger} {
		s.genetics.ped.add(c, 0)
	}
	sister := facing(s, x, y, 0)
	brother := person(s, Male, "Kakak", x, y)
	brother.X = sister.X + 1
	for _, c := range []*Creature{sister, brother} {
		c.Mother, c.Father = &Ref{mother.ID, mother.Name}, &Ref{father.ID, father.Name}
		s.genetics.ped.add(c, 0)
	}
	look := func(c *Creature) (float64, float64) {
		s.grid.rebuild(s.terrain.w, s.terrain.h, s.creatures)
		s.sense(c)
		return c.input[inMateKin], c.input[inMateHealth]
	}
	kin, health := look(sister)
	if !same(kin, 0.5) || !same(health, 1) || c0(sister.input[inPartnerNear]) {
		t.Fatalf("looking at her brother: kin %.3f, health %.3f", kin, health)
	}
	brother.Health = 0.5
	brother.Ill = &Illness{Disease: Respiratory, Severity: 0.2, Left: 1, Length: 2}
	if _, health := look(sister); !same(health, 0.5*(1-sickness(brother))) {
		t.Fatalf("a sick brother looks %.3f healthy", health)
	}
	// An unrelated man in front instead.
	brother.X = sister.X - 3
	other := person(s, Male, "Tetangga", x, y)
	other.X = sister.X + 1
	s.genetics.ped.add(other, 0)
	if kin, health := look(sister); kin != 0 || !same(health, 1) {
		t.Fatalf("a stranger: kin %.3f, health %.3f", kin, health)
	}
	// Nobody of the other sex in reach.
	other.X = sister.X - 3
	if kin, health := look(sister); kin != 0 || health != 0 {
		t.Fatal("mate senses lit with nobody there")
	}
	// A child is nobody's possible mate.
	brother.X, brother.Health, brother.Ill = sister.X+1, 1, nil
	grown := brother.BornTick
	brother.BornTick = s.tick
	if kin, health := look(sister); kin != 0 || health != 0 || sister.input[inPartnerNear] == 0 {
		t.Fatal("mate senses lit for a child")
	}
	brother.BornTick = grown
	// Genetics off: no senses, though the candidate is still noted for statistics.
	brother.X, brother.Health, brother.Ill = sister.X+1, 1, nil
	s.opts.NoGenetics = true
	if kin, health := look(sister); kin != 0 || health != 0 || sister.Heredity.mateID != brother.ID {
		t.Fatal("mate senses lit with genetics off")
	}
	s.geneticsTick()
	if s.genetics.stats.Periods[0].KinSeen != 1 {
		t.Fatalf("the desire sample missed a sibling pair: %+v", s.genetics.stats.Periods)
	}
}

func c0(v float64) bool { return v == 0 }
