package sim

import (
	"math"
	"math/bits"
)

// Diploid genetics (Fase 3c). Brains and the quantitative traits stay
// polygenic (crossover and blending in genome.go); beside them every
// genome carries a few Mendelian loci, two copies each, one from each parent.
// Most are recessive disorders: a carrier is healthy, a child who inherits
// the variant from both parents is not. Children of relatives are far more
// likely to inherit the same variant twice, which is inbreeding depression.
//
//   - Load. Every person carries a few recessive variants that would kill or
//     cripple when homozygous: about 0.6 lethal ones per person in a
//     genome-wide estimate (Gao et al. 2015), 1–2 lethal equivalents in older
//     ones. The founders draw theirs at the frequencies below: 0.57 lethal
//     variants and about one milder one each, some 0.3 lethal equivalents
//     per gamete, under the 0.7 implied by the excess deaths of first
//     cousins' children (4.4 points, Bittles & Neel 1994; 3.5 points,
//     Bittles & Black 2010). That keeps it modest, for a world begun by one
//     couple is inbred for centuries (F ≈ 0.3–0.4 for the first 300 years).
//   - Effects follow the kinds of disorder consanguinity raises (Fareed &
//     Afzal 2017 review): lethal disorders of infancy, weak muscles and
//     skeleton, weak immune defences, and reduced fertility. Each class
//     stands for many real genes, so its loci mutate faster than one gene
//     would (an estimate: 5·10⁻⁴ per locus and gamete; at that rate
//     selection holds lethal variants near √(μ/s) ≈ 0.025 per locus, the
//     founders' frequency; Crow 1999 for the genome-wide rate of new
//     deleterious mutations).
//   - Thalassaemia: carriers of a haemoglobin variant get milder malaria;
//     two copies are thalassaemia major, which without transfusion kills
//     most children (Weatherall & Clegg 2001). Carriers are 3–10 % of people
//     in much of Indonesia; in malarial lowlands the variant may spread by
//     itself (Haldane's balanced polymorphism).
//
// Nobody knows any of this. People see who looks healthy and may learn who
// their kin are (inMateKin, inMateHealth); avoiding relatives as mates
// (the Westermarck effect) can only evolve. No rule forbids kin mating.

const (
	numLoci = 31
	// geneMutation is the chance, per locus and gamete, that a healthy copy
	// turns into the recessive variant (estimate; see above).
	geneMutation = 5e-4
	// thalCarrierMalaria: how much milder malaria goes in a thalassaemia
	// carrier (estimate between α⁺-thalassaemia and HbE in Southeast Asia
	// and the sickle-cell trait in Africa; Taylor et al. 2012).
	thalCarrierMalaria = 0.5
)

// defectClass is a kind of recessive disorder and what two copies do.
type defectClass struct {
	name      string  // as the observer reads it
	loci      int     // loci of this class
	founderQ  float64 // share of the founders' chromosomes that carry the variant, per locus
	death     float64 // chance a homozygous baby dies in its first weeks
	vigor     float64 // share of strength left to the survivors
	severity  float64 // how much worse infections go
	fertility float64 // share of fertility left
}

var defectClasses = [...]defectClass{
	// Inborn errors of metabolism, spinal muscular atrophy and the like.
	{name: "kelainan bawaan berat", loci: 13, founderQ: 0.022, death: 0.8, vigor: 0.7, severity: 1, fertility: 1},
	// Recessive muscular dystrophies and skeletal dysplasias.
	{name: "otot dan tulang lemah", loci: 6, founderQ: 0.03, vigor: 0.8, severity: 1, fertility: 1},
	// Inborn immune deficiencies.
	{name: "kekebalan lemah", loci: 6, founderQ: 0.03, vigor: 1, severity: 1.5, fertility: 1},
	// E.g. the cystic fibrosis variants that leave men infertile.
	{name: "kurang subur", loci: 5, founderQ: 0.03, vigor: 1, severity: 1, fertility: 0.5},
	// The thalassaemia locus (see thalLocus).
	{name: "talasemia mayor", loci: 1, founderQ: 0.03, death: 0.5, vigor: 0.6, severity: 1.5, fertility: 1},
}

// thalLocus is the haemoglobin locus whose carriers resist malaria.
const thalLocus = numLoci - 1

// Per locus: its class and the effects of two copies.
var (
	locusClass                                         [numLoci]uint8
	locusDeath, locusVigor, locusSeverity, locusFertil [numLoci]float64
)

func init() {
	i := 0
	for k, d := range defectClasses {
		for range d.loci {
			locusClass[i] = uint8(k)
			locusDeath[i], locusVigor[i], locusSeverity[i], locusFertil[i] = d.death, d.vigor, d.severity, d.fertility
			i++
		}
	}
	if i != numLoci {
		panic("genetics: defect classes don't add up to numLoci")
	}
}

// Loci are the two copies of every Mendelian locus, as bit sets: bit i of
// Mat (from the mother) or Pat (from the father) is set when that copy of
// locus i carries the recessive variant. Genomes from before Fase 3c have
// none (nil): their carriers are drawn from the person's id when needed.
type Loci struct {
	Mat uint32 `json:"m"`
	Pat uint32 `json:"p"`
}

// expressed are the loci where both copies carry the variant.
func (l *Loci) expressed() uint32 {
	if l == nil {
		return 0
	}
	return l.Mat & l.Pat
}

// carried are the loci where one copy carries it: healthy carriers.
func (l *Loci) carried() uint32 {
	if l == nil {
		return 0
	}
	return l.Mat ^ l.Pat
}

// Heredity is what a person's birth decided about their genes.
type Heredity struct {
	// F is the inbreeding coefficient: the chance that both copies of a
	// gene come from the same ancestor (0 unrelated parents, 1/16 first
	// cousins, 1/4 siblings), counting kindredGens generations back.
	F float64 `json:"f,omitempty"`

	// The possible mate last seen (the inMateKin sense) and how related; a
	// cache of a pure function of the pedigree, so it isn't saved.
	mateID int64
	mateR  float64
}

// geneRand is a splitmix64 stream for genetic draws. Seeded from ids and
// ticks, it never touches the world's random generator, so genes don't
// shift any other draw.
type geneRand struct{ s uint64 }

func newGeneRand(id, tick int64, salt uint64) geneRand {
	return geneRand{uint64(id)*0x9E3779B97F4A7C15 ^ uint64(tick)*0xC2B2AE3D27D4EB4F ^ salt*0x165667B19E3779F9}
}

func (r *geneRand) next() uint64 {
	r.s += 0x9E3779B97F4A7C15
	z := r.s
	z = (z ^ z>>30) * 0xBF58476D1CE4E5B9
	z = (z ^ z>>27) * 0x94D049BB133111EB
	return z ^ z>>31
}

func (r *geneRand) float() float64 { return float64(r.next()>>11) / (1 << 53) }

// Salts that keep the different draws apart.
const (
	saltFounder = 0x5eed_f0
	saltVirtual = 0x5eed_f1
	saltGamete  = 0x5eed_f2
	saltNeonate = 0x5eed_f3
)

// founderLoci draws a healthy founder's genotype: each copy of each locus
// carries the variant at its class's frequency, and nobody starts with two.
func founderLoci(r *geneRand) *Loci {
	var l Loci
	for i := range numLoci {
		q := defectClasses[locusClass[i]].founderQ
		if r.float() < q {
			l.Mat |= 1 << i
		}
		if r.float() < q {
			l.Pat |= 1 << i
		}
	}
	l.Pat &^= l.Mat
	return &l
}

// lociOf is a genome's genotype; a genome from before Fase 3c gets one drawn
// from its owner's id, the same every time it is asked for.
func (s *Sim) lociOf(g *Genome, id int64) *Loci {
	if g != nil && g.Loci != nil {
		return g.Loci
	}
	r := newGeneRand(id, 0, saltVirtual^s.genetics.salt)
	return founderLoci(&r)
}

// gamete is one parent's contribution: each locus from either copy at
// random (the loci lie far apart and assort independently), now and then
// mutated into the recessive variant.
func gamete(l *Loci, r *geneRand) uint32 {
	pick := uint32(r.next())
	g := l.Mat&^pick | l.Pat&pick
	for i := range numLoci {
		if r.float() < geneMutation {
			g |= 1 << i
		}
	}
	return g
}

// foundGenes gives a new first couple their genes and their place at the
// root of the pedigree. Their genomes are fresh (foundingGenome), so writing
// the loci shares nothing.
func (s *Sim) foundGenes(founders ...*Creature) {
	gs := &s.genetics
	for _, c := range founders {
		if gs.salt == 0 {
			// The world's first founders seed every later genetic draw, so
			// worlds of different seeds differ although ids and ticks repeat.
			gs.salt = math.Float64bits(c.Genome.Traits.Hue) ^ math.Float64bits(c.Genome.Traits.Size)<<1 | 1
		}
		r := newGeneRand(c.ID, s.tick, saltFounder^gs.salt)
		c.Genome.Loci = founderLoci(&r)
		gs.ped.add(c, 0)
	}
}

// inheritGenes runs at a birth: the child's two copies of every locus, its
// inbreeding coefficient and its entry in the pedigree; and, when it
// inherited a lethal disorder from both sides, perhaps an early death.
func (s *Sim) inheritGenes(c, mother *Creature, p *Pregnancy) {
	gs := &s.genetics
	r := newGeneRand(c.ID, s.tick, saltGamete^gs.salt)
	c.Genome.Loci = &Loci{
		Mat: gamete(s.lociOf(mother.Genome, mother.ID), &r),
		Pat: gamete(s.lociOf(p.FatherGenome, p.Father.ID), &r),
	}
	c.Heredity.F = gs.ped.kinship(mother.ID, p.Father.ID)
	gs.ped.add(c, c.Heredity.F)
	hom := c.Genome.Loci.expressed()
	gs.stats.birth(s, c, hom != 0)
	if s.opts.NoGenetics || hom == 0 {
		return
	}
	survive := 1.0
	for h := hom; h != 0; h &= h - 1 {
		survive *= 1 - locusDeath[bits.TrailingZeros32(h)]
	}
	if survive < 1 {
		u := newGeneRand(c.ID, s.tick, saltNeonate^gs.salt)
		if u.float() >= survive {
			c.fate = "neonatal"
			gs.stats.lethal(c)
		}
	}
}

// --- effects --------------------------------------------------------------

// vigorGene is the share of strength a person's genes leave them: 1 for
// nearly everyone.
func (s *Sim) vigorGene(c *Creature) float64 {
	hom := c.Genome.Loci.expressed()
	if hom == 0 || s.opts.NoGenetics {
		return 1
	}
	v := 1.0
	for ; hom != 0; hom &= hom - 1 {
		v *= locusVigor[bits.TrailingZeros32(hom)]
	}
	return v
}

// fertilityGene is the share of fertility a person's genes leave them.
func (s *Sim) fertilityGene(c *Creature) float64 {
	hom := c.Genome.Loci.expressed()
	if hom == 0 || s.opts.NoGenetics {
		return 1
	}
	v := 1.0
	for ; hom != 0; hom &= hom - 1 {
		v *= locusFertil[bits.TrailingZeros32(hom)]
	}
	return v
}

// diseaseGene is how much worse (above 1) or milder an infection of kind d
// goes because of a person's genes.
func (s *Sim) diseaseGene(c *Creature, d Disease) float64 {
	l := c.Genome.Loci
	if l == nil || s.opts.NoGenetics {
		return 1
	}
	v := 1.0
	if d == Malaria && l.carried()&(1<<thalLocus) != 0 {
		v = thalCarrierMalaria
	}
	for hom := l.expressed(); hom != 0; hom &= hom - 1 {
		v *= locusSeverity[bits.TrailingZeros32(hom)]
	}
	return v
}

// --- mate-choice senses -----------------------------------------------------

// senseMate fills the mate-choice senses for the possible mate the brain is
// looking at (the nearest person of the other sex within partnerRange, as
// for inPartnerNear; nil if none): how related they are, by the pedigree,
// and how healthy they look, from what shows (health, strength, illness),
// never from the genes they carry. Nothing tells the brain what to make
// of either. Children are nobody's possible mates and have none: both
// senses stay 0 unless both are grown up.
func (s *Sim) senseMate(c *Creature, in *[NumInputs]float64, mate *Creature) {
	h := &c.Heredity
	if mate == nil || !s.adult(c) || !s.adult(mate) {
		h.mateID = 0
		return
	}
	if s.opts.NoGenetics {
		if h.mateID != mate.ID {
			h.mateID, h.mateR = mate.ID, -1 // remembered for the statistics only
		}
		return
	}
	in[inMateKin] = s.mateRelatedness(c, mate)
	in[inMateHealth] = clamp(mate.Health, 0, 1) * s.vigor(mate) * (1 - sickness(mate))
}

// mateRelatedness is c's coefficient of relationship to mate, remembered
// while c keeps looking at the same person.
func (s *Sim) mateRelatedness(c, mate *Creature) float64 {
	h := &c.Heredity
	if h.mateID == mate.ID && h.mateR >= 0 {
		return h.mateR
	}
	h.mateID, h.mateR = mate.ID, s.genetics.relatedness(c.ID, mate.ID)
	return h.mateR
}

// Relatedness is Wright's coefficient of relationship between two people,
// 0–1: 1/2 for parent and child or full siblings, 1/4 for half-siblings,
// grandparents, uncles and aunts, 1/8 for first cousins. It counts
// ancestors up to kindredGens generations above either.
func (s *Sim) Relatedness(a, b int64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.genetics.ped.relationship(a, b)
}

// Inbreeding is a person's inbreeding coefficient (0 if unknown).
func (s *Sim) Inbreeding(id int64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c := s.byID[id]; c != nil {
		return c.Heredity.F
	}
	if e := s.genetics.ped.entry(id); e != nil {
		return float64(e.f)
	}
	return 0
}
