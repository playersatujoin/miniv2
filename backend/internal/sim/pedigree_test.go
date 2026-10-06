package sim

import (
	"encoding/json"
	"math"
	"strconv"
	"testing"
)

// family builds a pedigree by hand: each person gets the next id (so parents
// always come first, as in a world) and the inbreeding their parents give.
type family struct {
	p    pedigree
	next int64
}

func (fm *family) born(mother, father int64) int64 {
	fm.next++
	c := &Creature{ID: fm.next, Name: "P" + strconv.FormatInt(fm.next, 10), BornTick: fm.next}
	if mother != 0 {
		c.Mother = &Ref{ID: mother}
	}
	if father != 0 {
		c.Father = &Ref{ID: father}
	}
	c.Heredity.F = fm.p.kinship(mother, father)
	fm.p.add(c, c.Heredity.F)
	return fm.next
}

func same(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestInbreedingOfKnownPedigrees(t *testing.T) {
	var fm family
	adam, hawa := fm.born(0, 0), fm.born(0, 0)
	son, daughter := fm.born(hawa, adam), fm.born(hawa, adam)
	stranger := fm.born(0, 0)

	cases := []struct {
		name     string
		mother   int64
		father   int64
		f        float64
		relation float64
	}{
		{"unrelated", hawa, stranger, 0, 0},
		{"full siblings", daughter, son, 0.25, 0.5},
		{"parent and child", hawa, son, 0.25, 0.5},
	}
	// Half-siblings and first cousins need more people.
	other := fm.born(0, 0)
	halfA, halfB := fm.born(hawa, adam), fm.born(other, adam)
	sp1, sp2 := fm.born(0, 0), fm.born(0, 0)
	cousinA, cousinB := fm.born(daughter, sp1), fm.born(sp2, son)
	sp3, sp4 := fm.born(0, 0), fm.born(0, 0)
	second1, second2 := fm.born(cousinA, sp3), fm.born(sp4, cousinB)
	cases = append(cases,
		struct {
			name     string
			mother   int64
			father   int64
			f        float64
			relation float64
		}{"half-siblings", halfA, halfB, 0.125, 0.25},
		struct {
			name     string
			mother   int64
			father   int64
			f        float64
			relation float64
		}{"first cousins", cousinA, cousinB, 0.0625, 0.125},
		struct {
			name     string
			mother   int64
			father   int64
			f        float64
			relation float64
		}{"second cousins", second1, second2, 1.0 / 64, 1.0 / 32},
	)
	for _, c := range cases {
		if got := fm.p.kinship(c.mother, c.father); !same(got, c.f) {
			t.Errorf("%s: child's F %.5f, want %.5f", c.name, got, c.f)
		}
		if got := fm.p.relationship(c.mother, c.father); !same(got, c.relation) {
			t.Errorf("%s: relationship %.5f, want %.5f", c.name, got, c.relation)
		}
		// Kinship is symmetric.
		if a, b := fm.p.kinship(c.mother, c.father), fm.p.kinship(c.father, c.mother); !same(a, b) {
			t.Errorf("%s: kinship not symmetric: %.6f vs %.6f", c.name, a, b)
		}
	}
	if got := fm.p.relationship(son, son); got != 1 {
		t.Fatalf("relationship with oneself %.3f", got)
	}
	if got := fm.p.kinship(son, son); !same(got, 0.5) {
		t.Fatalf("self-kinship of a non-inbred person %.3f, want 0.5", got)
	}
}

// Generations of brother–sister mating raise F along the classic series
// 1/4, 3/8, 1/2, 19/32 (Wright 1921), which needs the common ancestors' own
// inbreeding to be counted.
func TestRepeatedSibMatingSeries(t *testing.T) {
	var fm family
	m, f := fm.born(0, 0), fm.born(0, 0)
	want := []float64{0, 0.25, 0.375, 0.5, 0.59375}
	for gen, w := range want {
		a, b := fm.born(m, f), fm.born(m, f)
		if got := fm.p.inbreeding(a); !same(got, w) {
			t.Fatalf("generation %d: F %.5f, want %.5f", gen+1, got, w)
		}
		m, f = a, b
	}
	// Inbred siblings are more than half related: 2f/(1+F).
	r := fm.p.relationship(m, f)
	F := fm.p.inbreeding(m)
	if fk := fm.p.kinship(m, f); !same(r, 2*fk/(1+F)) || r <= 0.5 || r > 1 {
		t.Fatalf("relationship of inbred siblings %.4f", r)
	}
}

// Common ancestors further up than kinDepth generations are not counted.
func TestKinshipStopsAtDepth(t *testing.T) {
	var fm family
	root := fm.born(0, 0)
	lineA, lineB := root, root
	for range kinDepth + 1 {
		lineA = fm.born(lineA, fm.born(0, 0))
		lineB = fm.born(fm.born(0, 0), lineB)
	}
	if got := fm.p.kinship(lineA, lineB); got != 0 {
		t.Fatalf("an ancestor %d generations up still counted: %.6f", kinDepth+1, got)
	}
	// One generation less and it shows: (1/2)^(2·kinDepth+1).
	if got := fm.p.kinship(fm.p.entry(lineA).mother, fm.p.entry(lineB).father); !same(got, math.Pow(0.5, 2*kinDepth+1)) {
		t.Fatalf("an ancestor %d generations up: kinship %.8f", kinDepth, got)
	}
}

// Parents nobody remembers still count: siblings of forgotten parents are
// siblings.
func TestKinshipThroughUnknownParents(t *testing.T) {
	var p pedigree
	for _, id := range []int64{10, 11} {
		p.add(&Creature{ID: id, Mother: &Ref{ID: 3}, Father: &Ref{ID: 4}}, 0)
	}
	p.add(&Creature{ID: 12, Mother: &Ref{ID: 3}, Father: &Ref{ID: 5}}, 0)
	if got := p.relationship(10, 11); !same(got, 0.5) {
		t.Fatalf("siblings of forgotten parents: %.3f", got)
	}
	if got := p.relationship(10, 12); !same(got, 0.25) {
		t.Fatalf("half-siblings of forgotten parents: %.3f", got)
	}
	// Ids out of any range are just unknown people.
	if got := p.kinship(1<<40, 10); got != 0 {
		t.Fatal("a stranger is related")
	}
}

func TestPruneKeepsWhatKinshipNeeds(t *testing.T) {
	s, _, x, y := fakeWorld(t)
	// Twenty generations of a single line, then two cousins alive.
	var line []*Creature
	mother := person(s, Female, "Leluhur", x, y)
	s.genetics.ped.add(mother, 0)
	for g := range 20 {
		father := person(s, Male, "Ayah", x, y)
		s.genetics.ped.add(father, 0)
		child := person(s, Female, "Anak", x, y)
		child.Mother, child.Father, child.Generation = &Ref{mother.ID, mother.Name}, &Ref{father.ID, father.Name}, g+1
		s.genetics.ped.add(child, 0)
		line = append(line, mother, father)
		mother = child
	}
	uncle := person(s, Male, "Paman", x, y)
	s.genetics.ped.add(uncle, 0)
	a := person(s, Female, "A", x, y)
	b := person(s, Male, "B", x, y)
	a.Mother, a.Father = &Ref{mother.ID, ""}, &Ref{uncle.ID, ""}
	sib := person(s, Female, "Saudara", x, y)
	sib.Mother, sib.Father = &Ref{mother.ID, ""}, &Ref{uncle.ID, ""}
	s.genetics.ped.add(sib, 0)
	b.Mother, b.Father = &Ref{sib.ID, ""}, &Ref{uncle.ID + 1000, ""}
	s.genetics.ped.add(a, 0)
	s.genetics.ped.add(b, 0)
	before := s.genetics.ped.relationship(a.ID, b.ID)
	if !same(before, 0.25) {
		t.Fatalf("aunt and nephew: %.4f", before)
	}
	// Everyone but a, b and a dead-young child of a dies.
	baby := person(s, Male, "Bayi", x, y)
	baby.Mother = &Ref{a.ID, a.Name}
	s.genetics.ped.add(baby, 0)
	for _, c := range append(line, mother, uncle, sib, baby) {
		c.Health = 0
	}
	s.reap()
	s.prunePedigree()
	if got := s.genetics.ped.relationship(a.ID, b.ID); got != before {
		t.Fatalf("forgetting changed a kinship: %.4f, was %.4f", got, before)
	}
	if s.genetics.ped.entry(baby.ID) == nil {
		t.Fatal("a living mother's dead baby was forgotten")
	}
	if s.genetics.ped.entry(line[0].ID) != nil {
		t.Fatal("an ancestor 20 generations up is still remembered")
	}
	if s.genetics.ped.entry(mother.ID) == nil || s.genetics.ped.entry(line[len(line)-2*(keepDepth-2)].ID) == nil {
		t.Fatal("near ancestors were forgotten")
	}
	kept := 0
	for i := range s.genetics.ped.entries {
		if s.genetics.ped.entries[i].known {
			kept++
		}
	}
	if kept > 2*keepDepth+8 {
		t.Fatalf("%d remembered after forgetting", kept)
	}
}

func TestPedigreeSavesExactly(t *testing.T) {
	var fm family
	m, f := fm.born(0, 0), fm.born(0, 0)
	for range 6 {
		a, b := fm.born(m, f), fm.born(m, f)
		fm.born(a, fm.born(0, 0))
		m, f = a, b
	}
	fm.p.entry(m).died = 1234
	raw, err := json.Marshal(fm.p.save())
	if err != nil {
		t.Fatal(err)
	}
	var st pedigreeSave
	if err := json.Unmarshal(raw, &st); err != nil {
		t.Fatal(err)
	}
	var back pedigree
	back.restore(st)
	again, _ := json.Marshal(back.save())
	if string(again) != string(raw) {
		t.Fatal("pedigree changed in a save round trip")
	}
	for id := int64(1); id <= fm.next; id++ {
		for other := int64(1); other <= fm.next; other++ {
			if fm.p.kinship(id, other) != back.kinship(id, other) {
				t.Fatalf("kinship of %d and %d changed", id, other)
			}
		}
	}
}
