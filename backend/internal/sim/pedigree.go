package sim

import "math"

// The pedigree remembers who the parents of everyone who ever lived were, as
// far back as anyone alive can descend through, and computes kinship from it
// (Fase 3c). The inbreeding coefficient F of a child is the kinship of its
// parents: the chance that the two copies of a gene it inherits come from
// the same ancestor (Wright 1922). Kinship follows the classic recursion
// f(x, y) = ½·[f(mother of x, y) + f(father of x, y)] for x not an ancestor
// of y, and f(x, x) = ½·(1 + F_x), over the pedigree truncated kinDepth
// generations above each of the two; ancestors beyond that count as
// unrelated founders, as genealogies of real populations do.

const (
	// kinDepth is how many generations above each of two people the search
	// for common ancestors goes, so a child's F counts its ancestors up to
	// kinDepth+1 generations back (great-great-great-great-great-grandparents).
	kinDepth = 6
	// keepDepth: generations of ancestors remembered above everyone alive
	// (deeper than kinDepth, for the family tree).
	keepDepth = 10
	// pruneEvery: how often (ticks) people nobody alive descends from within
	// keepDepth generations are forgotten.
	pruneEvery = 10 * SecondsPerYear * TicksPerSecond
	// maxSpan bounds the id range the pedigree holds (a guard; real worlds
	// hold some tens of thousands).
	maxSpan = 1 << 20
	// pairCacheMax: remembered relatedness pairs before the cache is emptied.
	pairCacheMax = 1 << 17
)

// pedEntry is one person in the pedigree.
type pedEntry struct {
	mother, father int64 // ids; 0 unknown (a founder)
	born, died     int64 // ticks; died 0 while alive
	f              float32
	gen            int32
	sex            Sex
	known          bool
	name           string
	// Scratch for kinship queries: the query stamp and the node index.
	mark uint32
	node int32
}

// pedigree holds entries by id: entries[id-base]. Ids are handed out in
// order of birth, so a parent's id is always smaller than its child's.
type pedigree struct {
	base    int64
	entries []pedEntry
	calc    kinCalc
}

// slotAt is the entry slot for id, known or not, if it lies in range.
func (p *pedigree) slotAt(id int64) *pedEntry {
	i := id - p.base
	if id <= 0 || i < 0 || i >= int64(len(p.entries)) {
		return nil
	}
	return &p.entries[i]
}

// entry is id's entry if the pedigree remembers them.
func (p *pedigree) entry(id int64) *pedEntry {
	if e := p.slotAt(id); e != nil && e.known {
		return e
	}
	return nil
}

// grow makes room for id, returning its slot (nil if out of bounds).
func (p *pedigree) grow(id int64) *pedEntry {
	if id <= 0 {
		return nil
	}
	if len(p.entries) == 0 {
		p.base = id
		p.entries = append(p.entries[:0], pedEntry{})
		return &p.entries[0]
	}
	if id < p.base {
		extra := p.base - id
		if extra+int64(len(p.entries)) > maxSpan {
			return nil
		}
		grown := make([]pedEntry, int(extra)+len(p.entries), int(extra)+cap(p.entries))
		copy(grown[extra:], p.entries)
		p.entries, p.base = grown, id
	}
	i := id - p.base
	if i >= maxSpan {
		return nil
	}
	for int64(len(p.entries)) <= i {
		p.entries = append(p.entries, pedEntry{})
	}
	return &p.entries[i]
}

// add records c with its inbreeding coefficient f.
func (p *pedigree) add(c *Creature, f float64) {
	e := p.grow(c.ID)
	if e == nil {
		return
	}
	*e = pedEntry{born: c.BornTick, f: float32(f), gen: int32(c.Generation), sex: c.Sex, known: true, name: c.Name, mark: e.mark, node: e.node}
	if c.Mother != nil {
		e.mother = c.Mother.ID
	}
	if c.Father != nil {
		e.father = c.Father.ID
	}
}

// inbreeding is id's inbreeding coefficient as remembered (0 if unknown).
func (p *pedigree) inbreeding(id int64) float64 {
	if e := p.entry(id); e != nil {
		return float64(e.f)
	}
	return 0
}

// relationship is Wright's coefficient of relationship of a and b:
// 2·f(a, b) / √((1+F_a)(1+F_b)), between 0 and 1.
func (p *pedigree) relationship(a, b int64) float64 {
	if a == b {
		return 1
	}
	f := p.kinship(a, b)
	if f == 0 {
		return 0
	}
	return clamp(2*f/math.Sqrt((1+p.inbreeding(a))*(1+p.inbreeding(b))), 0, 1)
}

// --- kinship ---------------------------------------------------------------

// kinNode is a person in one kinship query's truncated pedigree.
type kinNode struct {
	id             int64
	f              float64
	mother, father int32 // node indexes, -1 outside the truncated pedigree
	side           uint8 // reached from a (1), from b (2)
	ready          bool  // ancestry bits computed
}

// kinCalc is the reusable scratch of kinship queries.
type kinCalc struct {
	stamp    uint32
	nodes    []kinNode
	foreign  map[int64]int32 // ids outside the pedigree's range
	frontier []int64
	next     []int64
	words    int
	bits     []uint64 // per node: itself and its ancestors in the query, as a bit set
	memoN    int
	memoAt   []uint32
	memo     []float64
}

func (k *kinCalc) begin(p *pedigree) {
	k.stamp++
	if k.stamp == 0 {
		for i := range p.entries {
			p.entries[i].mark = 0
		}
		clear(k.memoAt)
		k.stamp = 1
	}
	k.nodes = k.nodes[:0]
	if len(k.foreign) > 0 {
		clear(k.foreign)
	}
}

// find is the node of id in this query, or -1.
func (k *kinCalc) find(p *pedigree, id int64) int32 {
	if id <= 0 {
		return -1
	}
	if e := p.slotAt(id); e != nil {
		if e.mark == k.stamp {
			return e.node
		}
		return -1
	}
	if n, ok := k.foreign[id]; ok {
		return n
	}
	return -1
}

// nodeOf is the node of id in this query, made if new.
func (k *kinCalc) nodeOf(p *pedigree, id int64) int32 {
	if n := k.find(p, id); n >= 0 {
		return n
	}
	n := int32(len(k.nodes))
	node := kinNode{id: id, mother: -1, father: -1}
	if e := p.slotAt(id); e != nil {
		e.mark, e.node = k.stamp, n
		if e.known {
			node.f = float64(e.f)
		}
	} else {
		if k.foreign == nil {
			k.foreign = map[int64]int32{}
		}
		k.foreign[id] = n
	}
	k.nodes = append(k.nodes, node)
	return n
}

// climb gathers root's ancestors up to kinDepth generations, breadth first.
func (k *kinCalc) climb(p *pedigree, root int64, side uint8) int32 {
	r := k.nodeOf(p, root)
	if k.nodes[r].side&side != 0 {
		return r
	}
	k.nodes[r].side |= side
	k.frontier = append(k.frontier[:0], root)
	for d := 0; d < kinDepth && len(k.frontier) > 0; d++ {
		k.next = k.next[:0]
		for _, id := range k.frontier {
			e := p.entry(id)
			if e == nil {
				continue
			}
			for _, par := range [2]int64{e.mother, e.father} {
				if par <= 0 {
					continue
				}
				n := k.nodeOf(p, par)
				if k.nodes[n].side&side == 0 {
					k.nodes[n].side |= side
					k.next = append(k.next, par)
				}
			}
		}
		k.frontier, k.next = k.next, k.frontier
	}
	return r
}

// link connects every node to its parents within the query's pedigree.
func (k *kinCalc) link(p *pedigree) {
	for i := range k.nodes {
		if e := p.entry(k.nodes[i].id); e != nil {
			k.nodes[i].mother = k.find(p, e.mother)
			k.nodes[i].father = k.find(p, e.father)
		}
	}
}

// ancestry fills node i's bit set: itself and all its ancestors.
func (k *kinCalc) ancestry(i int32) {
	n := &k.nodes[i]
	if n.ready {
		return
	}
	n.ready = true
	w := k.words
	row := k.bits[int(i)*w : (int(i)+1)*w]
	row[i/64] |= 1 << (i % 64)
	for _, par := range [2]int32{n.mother, n.father} {
		if par < 0 {
			continue
		}
		k.ancestry(par)
		for j, v := range k.bits[int(par)*w : (int(par)+1)*w] {
			row[j] |= v
		}
	}
}

// related reports whether nodes i and j share an ancestor (or are one).
func (k *kinCalc) related(i, j int32) bool {
	w := k.words
	a, b := k.bits[int(i)*w:(int(i)+1)*w], k.bits[int(j)*w:(int(j)+1)*w]
	for x := range a {
		if a[x]&b[x] != 0 {
			return true
		}
	}
	return false
}

// f is the kinship of nodes i and j.
func (k *kinCalc) f(i, j int32) float64 {
	if i < 0 || j < 0 {
		return 0
	}
	if i == j {
		return 0.5 * (1 + k.nodes[i].f)
	}
	if !k.related(i, j) {
		return 0
	}
	// Expand the younger: born later, it can't be the other's ancestor.
	if k.nodes[i].id < k.nodes[j].id {
		i, j = j, i
	}
	slot := int(i)*k.memoN + int(j)
	if k.memoAt[slot] == k.stamp {
		return k.memo[slot]
	}
	v := 0.5 * (k.f(k.nodes[i].mother, j) + k.f(k.nodes[i].father, j))
	k.memoAt[slot], k.memo[slot] = k.stamp, v
	return v
}

// kinship is the coefficient of kinship (coancestry) of a and b: the chance
// that a gene drawn from each is identical by descent.
func (p *pedigree) kinship(a, b int64) float64 {
	if a <= 0 || b <= 0 {
		return 0
	}
	if a == b {
		return 0.5 * (1 + p.inbreeding(a))
	}
	k := &p.calc
	k.begin(p)
	ia := k.climb(p, a, 1)
	ib := k.climb(p, b, 2)
	k.link(p)
	n := len(k.nodes)
	k.words = (n + 63) / 64
	k.bits = resetUint64(k.bits, n*k.words)
	k.ancestry(ia)
	k.ancestry(ib)
	if !k.related(ia, ib) {
		return 0
	}
	if k.memoN = n; len(k.memoAt) < n*n {
		k.memoAt = make([]uint32, n*n)
		k.memo = make([]float64, n*n)
	}
	// Stamps (not clearing) retire earlier queries' answers; a query's
	// layout depends on n, so a stale slot never carries this stamp.
	return k.f(ia, ib)
}

func resetUint64(s []uint64, n int) []uint64 {
	if cap(s) < n {
		return make([]uint64, n)
	}
	s = s[:n]
	clear(s)
	return s
}

// --- forgetting -------------------------------------------------------------

// prune forgets everyone no living person (nor the father of an unborn
// child) descends from within keepDepth generations, except the children of
// the living who died young, so trees still show them. What kinship needs
// (kinDepth generations above the living) is always kept, so forgetting
// never changes a kinship.
func (s *Sim) prunePedigree() {
	p := &s.genetics.ped
	if len(p.entries) == 0 {
		return
	}
	depth := make([]int8, len(p.entries)) // generations above the nearest root, +1; 0 = not reached
	var frontier, next []int64
	reach := func(id int64, d int8) {
		if e := p.slotAt(id); e != nil && e.known && depth[id-p.base] == 0 {
			depth[id-p.base] = d + 1
			frontier = append(frontier, id)
		}
	}
	for _, c := range s.creatures {
		reach(c.ID, 0)
		if c.Pregnancy != nil {
			reach(c.Pregnancy.Father.ID, 0)
		}
	}
	for d := int8(1); d <= keepDepth && len(frontier) > 0; d++ {
		frontier, next = next[:0], frontier
		for _, id := range next {
			e := p.entry(id)
			reach(e.mother, d)
			reach(e.father, d)
		}
	}
	lo, hi := -1, -1
	for i := range p.entries {
		e := &p.entries[i]
		if !e.known {
			continue
		}
		if depth[i] == 0 && s.living(e.mother) == nil && s.living(e.father) == nil {
			*e = pedEntry{mark: e.mark, node: e.node}
			continue
		}
		if lo < 0 {
			lo = i
		}
		hi = i
	}
	if lo < 0 {
		p.entries, p.base = nil, 0
		return
	}
	kept := make([]pedEntry, hi-lo+1)
	copy(kept, p.entries[lo:hi+1])
	p.entries, p.base = kept, p.base+int64(lo)
}

// ensurePedigree adds anyone alive the pedigree lacks: everyone in a world
// saved before Fase 3c. Their dead parents are known by id only.
func (s *Sim) ensurePedigree() {
	p := &s.genetics.ped
	for _, c := range s.creatures {
		if p.entry(c.ID) == nil {
			p.add(c, c.Heredity.F)
		}
	}
}

// relatedness is the coefficient of relationship of a and b, remembered per
// pair. Forgetting never changes a living pair's value, so neither the
// remembered pairs nor their loss on restore can change the world.
func (g *genetics) relatedness(a, b int64) float64 {
	if a == b {
		return 1
	}
	lo, hi := min(a, b), max(a, b)
	if lo <= 0 || hi >= 1<<32 {
		return g.ped.relationship(lo, hi)
	}
	key := uint64(lo)<<32 | uint64(hi)
	if v, ok := g.pairs[key]; ok {
		return v
	}
	if g.pairs == nil || len(g.pairs) >= pairCacheMax {
		g.pairs = make(map[uint64]float64, 1024)
	}
	v := g.ped.relationship(lo, hi)
	g.pairs[key] = v
	return v
}

// --- saving -----------------------------------------------------------------

// pedigreeSave is the pedigree as columns, in order of id.
type pedigreeSave struct {
	ID     []int64   `json:"id"`
	Mother []int64   `json:"mother"`
	Father []int64   `json:"father"`
	Born   []int64   `json:"born"`
	Died   []int64   `json:"died"`
	F      []float32 `json:"f"`
	Gen    []int32   `json:"gen"`
	Sex    []int8    `json:"sex"`
	Name   []string  `json:"name"`
}

func (p *pedigree) save() pedigreeSave {
	var out pedigreeSave
	for i := range p.entries {
		e := &p.entries[i]
		if !e.known {
			continue
		}
		out.ID = append(out.ID, p.base+int64(i))
		out.Mother = append(out.Mother, e.mother)
		out.Father = append(out.Father, e.father)
		out.Born = append(out.Born, e.born)
		out.Died = append(out.Died, e.died)
		out.F = append(out.F, e.f)
		out.Gen = append(out.Gen, e.gen)
		out.Sex = append(out.Sex, int8(e.sex))
		out.Name = append(out.Name, e.name)
	}
	return out
}

func (p *pedigree) restore(st pedigreeSave) {
	n := len(st.ID)
	if len(st.Mother) != n || len(st.Father) != n || len(st.Born) != n || len(st.Died) != n ||
		len(st.F) != n || len(st.Gen) != n || len(st.Sex) != n || len(st.Name) != n {
		return
	}
	for i, id := range st.ID {
		e := p.grow(id)
		if e == nil {
			continue
		}
		*e = pedEntry{mother: st.Mother[i], father: st.Father[i], born: st.Born[i], died: st.Died[i],
			f: st.F[i], gen: st.Gen[i], sex: Sex(st.Sex[i]), known: true, name: st.Name[i]}
	}
}
