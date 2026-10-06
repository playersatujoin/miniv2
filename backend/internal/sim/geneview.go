package sim

import (
	"math"
	"math/bits"
	"slices"
	"strconv"
)

// genetics is the world's genetic state (Fase 3c): the pedigree, the salt
// that makes genetic draws differ between worlds, and statistics.
type genetics struct {
	ped   pedigree
	salt  uint64
	stats geneStats
	// pairs remembers relatedness per pair of ids; a cache, not saved.
	pairs map[uint64]float64
}

// --- statistics ---------------------------------------------------------------

// Generation bands (of the child) and inbreeding bands for the statistics.
var (
	genBandStart = [...]int{1, 2, 3, 4, 6, 9, 13, 21, 36}
	fBandStart   = [...]float64{0, 1e-9, 1.0/64 - 1e-9, 1.0/16 - 1e-9, 1.0/8 - 1e-9, 1.0/4 - 1e-9}
	fBandLabel   = [...]string{"0 (tak sekerabat)", "< 1/64 (kerabat jauh)", "1/64–1/16 (sepupu dua kali)",
		"1/16–1/8 (sepupu)", "1/8–1/4 (saudara tiri, paman–keponakan)", "≥ 1/4 (saudara kandung, orang tua–anak)"}
)

const (
	numGenBands = len(genBandStart)
	numFBands   = len(fBandStart)
	periodYears = 25 // years per period in the time series
	periodsMax  = 400
	// closeKinF: a child of close kin (half-siblings, uncle and niece and
	// closer); consanguineousF: second cousins or closer, the usual
	// definition of a consanguineous union (Bittles & Black 2010).
	closeKinF       = 1.0/8 - 1e-9
	consanguineousF = 1.0/64 - 1e-9
)

func genBand(gen int) int {
	for b := numGenBands - 1; b > 0; b-- {
		if gen >= genBandStart[b] {
			return b
		}
	}
	return 0
}

func fBand(f float64) int {
	for b := numFBands - 1; b > 0; b-- {
		if f >= fBandStart[b] {
			return b
		}
	}
	return 0
}

// fTally sums the inbreeding of births.
type fTally struct {
	Births int     `json:"births,omitempty"`
	SumF   float64 `json:"sumF,omitempty"`
	Consan int     `json:"consan,omitempty"` // F ≥ 1/64
	Close  int     `json:"close,omitempty"`  // F ≥ 1/8
}

func (t *fTally) add(f float64) {
	t.Births++
	t.SumF += f
	if f >= consanguineousF {
		t.Consan++
	}
	if f >= closeKinF {
		t.Close++
	}
}

// fOutcome follows the children born in one inbreeding band.
type fOutcome struct {
	Births   int `json:"births,omitempty"`
	Affected int `json:"affected,omitempty"` // born with two copies of a recessive variant
	Lethal   int `json:"lethal,omitempty"`   // of whom died in their first weeks of it
	Under1   int `json:"under1,omitempty"`   // deaths before 1, all causes
	Under15  int `json:"under15,omitempty"`  // deaths before 15, all causes
}

// periodTally is periodYears of births and of the mate-desire sample.
type periodTally struct {
	Year int `json:"year"` // first year
	fTally
	// Once a second, adults looking at a possible adult mate: how many
	// times the mate was a parent, child or (half-)sibling (the relatives
	// people grow up with, whom the Westermarck effect concerns) or anyone
	// else, how often they wanted to mate then, and how related the others
	// were (the background kinship of a small population).
	KinSeen   int     `json:"kinSeen,omitempty"`
	KinWant   int     `json:"kinWant,omitempty"`
	OtherSeen int     `json:"otherSeen,omitempty"`
	OtherWant int     `json:"otherWant,omitempty"`
	OtherR    float64 `json:"otherR,omitempty"` // summed relatedness of the others
}

type geneStats struct {
	ByGen   [numGenBands]fTally `json:"byGen"`
	ByF     [numFBands]fOutcome `json:"byF"`
	Periods []periodTally       `json:"periods,omitempty"`
	Lethal  int                 `json:"lethal,omitempty"`
}

func (g *geneStats) period(s *Sim) *periodTally {
	year := int(s.time()/SecondsPerYear) / periodYears * periodYears
	if n := len(g.Periods); n > 0 && g.Periods[n-1].Year == year {
		return &g.Periods[n-1]
	}
	g.Periods = append(g.Periods, periodTally{Year: year})
	if over := len(g.Periods) - periodsMax; over > 0 {
		g.Periods = slices.Delete(g.Periods, 0, over)
	}
	return &g.Periods[len(g.Periods)-1]
}

func (g *geneStats) birth(s *Sim, c *Creature, affected bool) {
	f := c.Heredity.F
	g.ByGen[genBand(c.Generation)].add(f)
	o := &g.ByF[fBand(f)]
	o.Births++
	if affected {
		o.Affected++
	}
	g.period(s).add(f)
}

func (g *geneStats) lethal(c *Creature) {
	g.Lethal++
	g.ByF[fBand(c.Heredity.F)].Lethal++
}

// geneticsDeath notes a death in the pedigree and the statistics.
func (s *Sim) geneticsDeath(c *Creature) {
	if e := s.genetics.ped.entry(c.ID); e != nil {
		e.died = s.tick
	}
	if c.Generation == 0 {
		return // a founder, not a birth
	}
	o := &s.genetics.stats.ByF[fBand(c.Heredity.F)]
	age := s.ageYears(c)
	if age < 1 {
		o.Under1++
	}
	if age < 15 {
		o.Under15++
	}
}

// geneticsTick runs once a simulated second: the mate-desire sample, and
// every pruneEvery ticks forgetting the long dead.
func (s *Sim) geneticsTick() {
	gs := &s.genetics
	var kinSeen, kinWant, otherSeen, otherWant int
	otherR := 0.0
	for _, c := range s.creatures {
		h := &c.Heredity
		if h.mateID == 0 || !s.adult(c) {
			continue
		}
		m := s.living(h.mateID)
		if m == nil || !s.adult(m) {
			continue
		}
		if h.mateR < 0 {
			h.mateR = gs.relatedness(c.ID, m.ID)
		}
		want := c.output[outMate] > 0.5
		if firstDegree(c, m) {
			kinSeen++
			if want {
				kinWant++
			}
		} else {
			otherSeen++
			otherR += h.mateR
			if want {
				otherWant++
			}
		}
	}
	if kinSeen+otherSeen > 0 {
		p := gs.stats.period(s)
		p.KinSeen += kinSeen
		p.KinWant += kinWant
		p.OtherSeen += otherSeen
		p.OtherWant += otherWant
		p.OtherR += otherR
	}
	if s.tick%pruneEvery == 0 {
		s.prunePedigree()
		// Pairs with someone dead are never asked for again.
		for key := range gs.pairs {
			if s.living(int64(key>>32)) == nil || s.living(int64(key&(1<<32-1))) == nil {
				delete(gs.pairs, key)
			}
		}
	}
}

// firstDegree reports whether a and b are parent and child or share a parent.
func firstDegree(a, b *Creature) bool {
	return isParent(a, b) || isParent(b, a) ||
		a.Mother != nil && b.Mother != nil && a.Mother.ID == b.Mother.ID ||
		a.Father != nil && b.Father != nil && a.Father.ID == b.Father.ID
}

// --- saving -------------------------------------------------------------------

type geneticsSave struct {
	Salt     uint64       `json:"salt,omitempty"`
	Pedigree pedigreeSave `json:"pedigree"`
	Stats    geneStats    `json:"stats"`
}

func (s *Sim) saveGenetics() *geneticsSave {
	gs := &s.genetics
	return &geneticsSave{Salt: gs.salt, Pedigree: gs.ped.save(), Stats: gs.stats}
}

// restoreGenetics runs after the creatures are restored. A world saved
// before Fase 3c has no pedigree: everyone alive enters it with their
// parents known by id, and nobody's genes are known until they are needed
// (lociOf).
func (s *Sim) restoreGenetics(st *geneticsSave) {
	if st != nil {
		gs := &s.genetics
		gs.salt = st.Salt
		gs.ped.restore(st.Pedigree)
		gs.stats = st.Stats
	}
	s.ensurePedigree()
}

// --- views ----------------------------------------------------------------------

// GeneticsView is one person's genes as the Inspector shows them.
type GeneticsView struct {
	// F is the inbreeding coefficient (0 unrelated parents … 1/4 siblings).
	F float64 `json:"f"`
	// Carried: loci with one copy of a recessive variant (healthy carrier).
	Carried int `json:"carried"`
	// Defects: the recessive disorders they have (two copies).
	Defects []string `json:"defects"`
	// MalariaShield: carrier of thalassaemia, so malaria goes milder.
	MalariaShield bool `json:"malariaShield"`
	// Known is false for people born before genetics: their variants are a guess.
	Known bool `json:"known"`
}

func (s *Sim) geneticsView(c *Creature) *GeneticsView {
	l := s.lociOf(c.Genome, c.ID)
	return &GeneticsView{
		F:             r4(c.Heredity.F),
		Carried:       bits.OnesCount32(l.carried()),
		Defects:       defectNames(l.expressed()),
		MalariaShield: l.carried()&(1<<thalLocus) != 0,
		Known:         c.Genome.Loci != nil,
	}
}

func r4(v float64) float64 { return math.Round(v*10000) / 10000 }

// defectNames lists the disorders of the expressed loci, one per class.
func defectNames(hom uint32) []string {
	out := []string{}
	var seen [len(defectClasses)]bool
	for ; hom != 0; hom &= hom - 1 {
		k := locusClass[bits.TrailingZeros32(hom)]
		if !seen[k] {
			seen[k] = true
			out = append(out, defectClasses[k].name)
		}
	}
	return out
}

// GeneticsInfo summarises the island's genetics: inbreeding now, by
// generation and over time, its outcomes, and how often adults wanted to
// mate with close kin and with others.
type GeneticsInfo struct {
	Remembered int     `json:"remembered"` // people in the pedigree
	MeanF      float64 `json:"meanF"`      // of the living
	Inbred     float64 `json:"inbred"`     // share of the living born to second cousins or closer
	CloseKin   float64 `json:"closeKin"`   // share of the living born to close kin (F ≥ 1/8)
	Carried    float64 `json:"carried"`    // recessive variants carried (one copy) per living person
	Affected   float64 `json:"affected"`   // share of the living with a recessive disorder
	Lethal     int     `json:"lethal"`     // babies lost to recessive disorders, all time
	// Variants: per kind of disorder, the share of the living's chromosomes
	// that carry a variant (mean over its loci).
	Variants     []VariantView    `json:"variants"`
	ByGeneration []GenerationBand `json:"byGeneration"`
	ByF          []InbreedingBand `json:"byF"`
	Periods      []GeneticsPeriod `json:"periods"`
}

type VariantView struct {
	Name      string  `json:"name"`
	Frequency float64 `json:"frequency"`
}

// GenerationBand is the births of one band of generations.
type GenerationBand struct {
	Label          string  `json:"label"`
	Births         int     `json:"births"`
	MeanF          float64 `json:"meanF"`
	Consanguineous float64 `json:"consanguineous"` // share born to second cousins or closer
	CloseKin       float64 `json:"closeKin"`       // share born to close kin
}

// InbreedingBand is what became of the children born in one band of F.
type InbreedingBand struct {
	Label    string `json:"label"`
	Births   int    `json:"births"`
	Affected int    `json:"affected"`
	Lethal   int    `json:"lethal"`
	Under1   int    `json:"under1"`
	Under15  int    `json:"under15"`
	// Q15 is deaths before 15 per birth (children still young are not yet
	// counted, so it is a lower bound for recent births).
	Q15 *float64 `json:"q15"`
}

// GeneticsPeriod is periodYears of births and mate choice.
type GeneticsPeriod struct {
	Year     int     `json:"year"`
	Births   int     `json:"births"`
	MeanF    float64 `json:"meanF"`
	CloseKin float64 `json:"closeKin"`
	// The share of moments an adult looking at a possible mate wanted to
	// mate, when the mate was a parent, child or (half-)sibling and when
	// anyone else (nil: too few), and how related the others were on average.
	KinDesire   *float64 `json:"kinDesire"`
	OtherDesire *float64 `json:"otherDesire"`
	KinSeen     int      `json:"kinSeen"`
	OtherSeen   int      `json:"otherSeen"`
	OtherR      float64  `json:"otherR"`
}

func share(n, of int) float64 {
	if of == 0 {
		return 0
	}
	return float64(n) / float64(of)
}

func shareOrNil(n, of, least int) *float64 {
	if of < least {
		return nil
	}
	return ptr(share(n, of), 3)
}

// Genetics reports the island's genetics.
func (s *Sim) Genetics() GeneticsInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.geneticsInfo()
}

func (s *Sim) geneticsInfo() GeneticsInfo {
	gs := &s.genetics
	info := GeneticsInfo{Lethal: gs.stats.Lethal, Variants: []VariantView{}, ByGeneration: []GenerationBand{},
		ByF: []InbreedingBand{}, Periods: []GeneticsPeriod{}}
	for i := range gs.ped.entries {
		if gs.ped.entries[i].known {
			info.Remembered++
		}
	}
	var counts [numLoci]int
	n, inbred, close, carried, affected := 0, 0, 0, 0, 0
	sumF := 0.0
	for _, c := range s.creatures {
		n++
		f := c.Heredity.F
		sumF += f
		if f >= consanguineousF {
			inbred++
		}
		if f >= closeKinF {
			close++
		}
		l := s.lociOf(c.Genome, c.ID)
		carried += bits.OnesCount32(l.carried())
		if l.expressed() != 0 {
			affected++
		}
		for i := range numLoci {
			counts[i] += int(l.Mat>>i&1 + l.Pat>>i&1)
		}
	}
	if n > 0 {
		info.MeanF = r4(sumF / float64(n))
		info.Inbred = r3(share(inbred, n))
		info.CloseKin = r3(share(close, n))
		info.Carried = r3(float64(carried) / float64(n))
		info.Affected = r3(share(affected, n))
	}
	i := 0
	for _, d := range defectClasses {
		sum := 0
		for range d.loci {
			sum += counts[i]
			i++
		}
		freq := 0.0
		if n > 0 {
			freq = float64(sum) / float64(2*n*d.loci)
		}
		info.Variants = append(info.Variants, VariantView{Name: d.name, Frequency: r4(freq)})
	}
	for b, t := range gs.stats.ByGen {
		label := strconv.Itoa(genBandStart[b])
		switch {
		case b == numGenBands-1:
			label += "+"
		case genBandStart[b+1]-1 > genBandStart[b]:
			label += "–" + strconv.Itoa(genBandStart[b+1]-1)
		}
		info.ByGeneration = append(info.ByGeneration, GenerationBand{Label: label, Births: t.Births,
			MeanF: r4(t.SumF / math.Max(1, float64(t.Births))), Consanguineous: r3(share(t.Consan, t.Births)), CloseKin: r3(share(t.Close, t.Births))})
	}
	for b, o := range gs.stats.ByF {
		info.ByF = append(info.ByF, InbreedingBand{Label: fBandLabel[b], Births: o.Births, Affected: o.Affected,
			Lethal: o.Lethal, Under1: o.Under1, Under15: o.Under15, Q15: shareOrNil(o.Under15, o.Births, 1)})
	}
	for _, p := range gs.stats.Periods {
		info.Periods = append(info.Periods, GeneticsPeriod{Year: p.Year, Births: p.Births,
			MeanF: r4(p.SumF / math.Max(1, float64(p.Births))), CloseKin: r3(share(p.Close, p.Births)),
			KinDesire: shareOrNil(p.KinWant, p.KinSeen, 20), OtherDesire: shareOrNil(p.OtherWant, p.OtherSeen, 20),
			KinSeen: p.KinSeen, OtherSeen: p.OtherSeen, OtherR: r3(p.OtherR / math.Max(1, float64(p.OtherSeen)))})
	}
	return info
}

// --- family tree ----------------------------------------------------------------

// FamilyPerson is someone in a family tree.
type FamilyPerson struct {
	ID         int64    `json:"id"`
	Name       string   `json:"name"`
	Sex        string   `json:"sex"`
	Generation int      `json:"generation"`
	Mother     int64    `json:"mother,omitempty"`
	Father     int64    `json:"father,omitempty"`
	F          float64  `json:"f"`
	Alive      bool     `json:"alive"`
	Born       float64  `json:"born"`           // year of birth
	Died       *float64 `json:"died,omitempty"` // year of death, if known
	// Founder: one of the first humans (an Adam or a Hawa). Forgotten: the
	// pedigree no longer remembers them (nobody alive descends from them
	// within keepDepth generations), so only the id and maybe a name are known.
	Founder   bool `json:"founder,omitempty"`
	Forgotten bool `json:"forgotten,omitempty"`
}

// FamilyView is a person's family tree, up and down.
type FamilyView struct {
	Root  int64 `json:"root"`
	Depth int   `json:"depth"`
	// Ancestors[k-1] are the 2^k places k generations up, in Ahnentafel
	// order (father before mother; 0 = unknown). One person in several
	// places is pedigree collapse: their descendants married each other.
	Ancestors [][]int64 `json:"ancestors"`
	// Descendants[k-1] are everyone k generations down, by id.
	Descendants [][]int64      `json:"descendants"`
	People      []FamilyPerson `json:"people"`
	// Genetics of the person, when alive.
	Genetics *GeneticsView `json:"genetics"`
	// Relatedness of the person's parents (coefficient of relationship).
	ParentsRelated float64 `json:"parentsRelated"`
}

// familyDepthMax bounds the generations a family tree shows each way.
const familyDepthMax = keepDepth

// Family returns id's family tree, depth generations up and down; false if
// the pedigree doesn't know them.
func (s *Sim) Family(id int64, depth int) (FamilyView, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := &s.genetics.ped
	if p.entry(id) == nil {
		return FamilyView{}, false
	}
	depth = max(1, min(depth, familyDepthMax))
	v := FamilyView{Root: id, Depth: depth, Ancestors: [][]int64{}, Descendants: [][]int64{}}
	people := map[int64]bool{id: true}
	level := []int64{id}
	for range depth {
		next := make([]int64, 2*len(level))
		any := false
		for i, x := range level {
			if e := p.entry(x); x != 0 && e != nil {
				next[2*i], next[2*i+1] = e.father, e.mother
				any = any || e.father != 0 || e.mother != 0
			}
		}
		if !any {
			break
		}
		for _, x := range next {
			if x != 0 {
				people[x] = true
			}
		}
		v.Ancestors = append(v.Ancestors, next)
		level = next
	}
	// Descendants, from an index of children built for this request.
	children := map[int64][]int64{}
	for i := range p.entries {
		e := &p.entries[i]
		if !e.known {
			continue
		}
		cid := p.base + int64(i)
		if e.mother != 0 {
			children[e.mother] = append(children[e.mother], cid)
		}
		if e.father != 0 {
			children[e.father] = append(children[e.father], cid)
		}
	}
	gen := []int64{id}
	for range depth {
		var next []int64
		for _, x := range gen {
			next = append(next, children[x]...)
		}
		slices.Sort(next)
		next = slices.Compact(next)
		if len(next) == 0 {
			break
		}
		for _, x := range next {
			people[x] = true
		}
		v.Descendants = append(v.Descendants, next)
		gen = next
	}
	ids := make([]int64, 0, len(people))
	for x := range people {
		ids = append(ids, x)
	}
	slices.Sort(ids)
	for _, x := range ids {
		v.People = append(v.People, s.familyPerson(x))
	}
	if c := s.living(id); c != nil {
		v.Genetics = s.geneticsView(c)
	}
	if e := p.entry(id); e.mother != 0 && e.father != 0 {
		v.ParentsRelated = r4(p.relationship(e.mother, e.father))
	}
	return v, true
}

func (s *Sim) familyPerson(id int64) FamilyPerson {
	e := s.genetics.ped.entry(id)
	if e == nil {
		fp := FamilyPerson{ID: id, Forgotten: true}
		// A living child may still know a forgotten parent's name.
		for _, c := range s.creatures {
			if c.Mother != nil && c.Mother.ID == id {
				fp.Name, fp.Sex = c.Mother.Name, Female.String()
				break
			}
			if c.Father != nil && c.Father.ID == id {
				fp.Name, fp.Sex = c.Father.Name, Male.String()
				break
			}
		}
		return fp
	}
	fp := FamilyPerson{ID: id, Name: e.name, Sex: e.sex.String(), Generation: int(e.gen), Mother: e.mother, Father: e.father,
		F: r4(float64(e.f)), Born: math.Floor(math.Max(0, float64(e.born)*dt/SecondsPerYear)) + 1,
		Founder: e.mother == 0 && e.father == 0 && e.gen == 0}
	if c := s.living(id); c != nil {
		fp.Alive = true
	} else if e.died != 0 {
		fp.Died = ptr(math.Floor(float64(e.died)*dt/SecondsPerYear)+1, 0)
	}
	return fp
}
