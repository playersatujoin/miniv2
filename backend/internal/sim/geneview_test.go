package sim

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"slices"
	"testing"
)

func TestFamilyTree(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	adam := person(s, Male, "Adam", x, y)
	hawa := person(s, Female, "Hawa", x, y)
	s.foundGenes(adam, hawa)
	birth := func(mother, father *Creature, name string, sex Sex) *Creature {
		c := s.newChild(mother, &Pregnancy{Father: Ref{father.ID, father.Name}, FatherGenome: father.Genome})
		c.Name, c.Sex, c.fate = name, sex, ""
		s.genetics.ped.entry(c.ID).name, s.genetics.ped.entry(c.ID).sex = name, sex
		s.add(c)
		return c
	}
	son := birth(hawa, adam, "Kabil", Male)
	daughter := birth(hawa, adam, "Iqlima", Female)
	grandchild := birth(daughter, son, "Cucu", Female)
	if !same(grandchild.Heredity.F, 0.25) {
		t.Fatalf("child of siblings: F %.3f", grandchild.Heredity.F)
	}
	s.tick = 100
	adam.Health = 0
	s.reap()

	v, ok := s.Family(grandchild.ID, 3)
	if !ok {
		t.Fatal("no family tree")
	}
	if len(v.Ancestors) != 2 || !slices.Equal(v.Ancestors[0], []int64{son.ID, daughter.ID}) ||
		!slices.Equal(v.Ancestors[1], []int64{adam.ID, hawa.ID, adam.ID, hawa.ID}) {
		t.Fatalf("ancestors %v", v.Ancestors)
	}
	if len(v.Descendants) != 0 || v.Genetics == nil || !same(v.Genetics.F, 0.25) || !same(v.ParentsRelated, 0.5) {
		t.Fatalf("tree %+v", v)
	}
	byID := map[int64]FamilyPerson{}
	for _, p := range v.People {
		byID[p.ID] = p
	}
	if p := byID[adam.ID]; !p.Founder || p.Alive || p.Died == nil || p.Name != "Adam" {
		t.Fatalf("Adam in the tree: %+v", p)
	}
	if p := byID[son.ID]; !p.Alive || p.Mother != hawa.ID || p.Father != adam.ID || p.Generation != 1 {
		t.Fatalf("son in the tree: %+v", p)
	}
	down, _ := s.Family(hawa.ID, 5)
	if len(down.Descendants) != 2 || !slices.Equal(down.Descendants[0], []int64{son.ID, daughter.ID}) ||
		!slices.Equal(down.Descendants[1], []int64{grandchild.ID}) || len(down.Ancestors) != 0 {
		t.Fatalf("Hawa's descendants %v, ancestors %v", down.Descendants, down.Ancestors)
	}
	if _, ok := s.Family(99999, 3); ok {
		t.Fatal("a tree for nobody")
	}
	if !same(s.Relatedness(son.ID, daughter.ID), 0.5) || !same(s.Inbreeding(grandchild.ID), 0.25) {
		t.Fatal("public kinship accessors")
	}
}

// A world's genetics survive a save exactly, and statistics fill in as
// people are born.
func TestGeneticsSaveRestore(t *testing.T) {
	m := testMap(t, 48)
	a := New(m, 3)
	a.Advance(int(60 * SecondsPerYear * TicksPerSecond))
	info := a.Genetics()
	if a.births == 0 || info.Remembered < 3 {
		t.Skipf("no births in this world (%d remembered)", info.Remembered)
	}
	born := 0
	for _, g := range info.ByGeneration {
		born += g.Births
	}
	if born != a.births {
		t.Fatalf("statistics count %d births, the world %d", born, a.births)
	}
	data, err := a.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	b, err := Restore(m, data)
	if err != nil {
		t.Fatal(err)
	}
	ja, _ := json.Marshal(a.saveGenetics())
	jb, _ := json.Marshal(b.saveGenetics())
	if !bytes.Equal(ja, jb) {
		t.Fatal("genetics changed in a save round trip")
	}
	for i, c := range a.creatures {
		d := b.creatures[i]
		if c.Heredity.F != d.Heredity.F || (c.Genome.Loci == nil) != (d.Genome.Loci == nil) ||
			c.Genome.Loci != nil && *c.Genome.Loci != *d.Genome.Loci {
			t.Fatalf("%s's genes changed in a save round trip", c.Name)
		}
	}
	a.Advance(10 * TicksPerSecond)
	b.Advance(10 * TicksPerSecond)
	ja, _ = json.Marshal(a.Genetics())
	jb, _ = json.Marshal(b.Genetics())
	if !bytes.Equal(ja, jb) {
		t.Fatal("restored genetics diverged")
	}
}

// A world saved before Fase 3c (no pedigree, no genes) enters everyone alive
// into the pedigree with their parents by id.
func TestOldWorldJoinsThePedigree(t *testing.T) {
	m := testMap(t, 48)
	a := New(m, 3)
	a.Advance(int(40 * SecondsPerYear * TicksPerSecond))
	a.genetics = genetics{}
	for _, c := range a.creatures {
		c.Genome.Loci, c.Heredity = nil, Heredity{}
	}
	data, err := a.MarshalState()
	if err != nil {
		t.Fatal(err)
	}
	var st map[string]json.RawMessage
	raw := gunzip(t, data)
	if err := json.Unmarshal(raw, &st); err != nil {
		t.Fatal(err)
	}
	delete(st, "genetics")
	raw, _ = json.Marshal(st)
	b, err := Restore(m, gzipBytes(t, raw))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range b.creatures {
		e := b.genetics.ped.entry(c.ID)
		if e == nil {
			t.Fatalf("%s is not in the pedigree", c.Name)
		}
		if c.Mother != nil && e.mother != c.Mother.ID {
			t.Fatal("parents lost")
		}
		if b.vigor(c) > 1 || b.geneticsView(c).Known {
			t.Fatal("old genes treated as known")
		}
	}
	b.Advance(20 * TicksPerSecond)
}

func gunzip(t *testing.T, data []byte) []byte {
	t.Helper()
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func gzipBytes(t *testing.T, raw []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(raw); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
