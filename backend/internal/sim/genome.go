package sim

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
)

const (
	numRays     = 5
	numChannels = 6
	numInternal = 30

	NumInputs  = numRays*numChannels + numInternal
	NumOutputs = 13

	// Brain size is inherited and evolves: hidden neurons per genome.
	minHidden     = 8
	maxHidden     = 64
	initialHidden = 20

	maxWeight = 4
)

// Input layout: six ray channels (channel-major), then internal senses.
const (
	inObstacle = iota * numRays
	inFood
	inWater
	inMate
	inSame
	inResource
)

const (
	inEnergy = numRays*numChannels + iota
	inHydration
	inHealth
	inAge
	inFertile
	inPregnant
	inFoodHere
	inWaterNear
	inSourceHere
	inBumped
	inPartnerNear
	inFamilyNear
	inStrangerNear
	inNearRep
	inAttacked
	inCarryFood
	inCarryMaterial
	inCarryWeapon
	inOwnHouseNear
	inOtherHouseNear
	inCanCraft
	inCanBuild
	inTeacherNear
	inStudentNear
	inBestSkill
	inReward
	inLibraryNear
	inClock
	inNoise
	inBias
)

const (
	outTurn = iota
	outMove
	outEat
	outDrink
	outMate
	outRest
	outGather
	outCraft
	outBuild
	outGive
	outSteal
	outAttack
	outTeach
)

var rayDegrees = [numRays]float64{-60, -30, 0, 30, 60}

var InputLabels = func() []string {
	var out []string
	for _, ch := range []string{"rintangan", "makanan", "air", "lawan jenis", "sesama jenis", "sumber daya"} {
		for _, d := range rayDegrees {
			out = append(out, fmt.Sprintf("%s %g°", ch, d))
		}
	}
	return append(out, "energi", "hidrasi", "kesehatan", "usia", "subur", "hamil", "makanan di sini",
		"air di dekat", "sumber di sini", "menabrak", "pasangan dekat", "keluarga dekat", "orang asing dekat",
		"reputasi orang dekat", "diserang", "bawa makanan", "bawa bahan", "bawa senjata", "rumah sendiri dekat",
		"rumah orang lain dekat", "bisa membuat", "bisa membangun", "guru dekat", "murid dekat",
		"keahlian tertinggi", "imbalan terakhir", "perpustakaan dekat", "jam internal", "acak (kehendak)", "bias")
}()

var OutputLabels = []string{"belok", "gerak", "makan", "minum", "kawin", "istirahat",
	"kumpulkan", "buat", "bangun", "beri", "curi", "serang", "ajar"}

type Traits struct {
	Hue          float64 `json:"hue"`
	Size         float64 `json:"size"`
	MaxSpeed     float64 `json:"maxSpeed"`
	Vision       float64 `json:"vision"`
	Metabolism   float64 `json:"metabolism"`
	Lifespan     float64 `json:"lifespan"`
	MutationRate float64 `json:"mutationRate"`

	// Lifetime learning (see learning.go): how fast experience changes the
	// brain, how long a decision stays "eligible" for credit, and how far any
	// one weight may drift from what was inherited.
	LearningRate float64 `json:"learningRate"`
	TraceDecay   float64 `json:"traceDecay"`
	Plasticity   float64 `json:"plasticity"`
}

// weights are stored as float32: half the memory, and saved compactly as
// base64 of little-endian bytes, which round-trips exactly.
type weights []float32

func (w weights) MarshalJSON() ([]byte, error) {
	buf := make([]byte, 4*len(w))
	for i, v := range w {
		binary.LittleEndian.PutUint32(buf[4*i:], math.Float32bits(v))
	}
	return json.Marshal(base64.StdEncoding.EncodeToString(buf))
}

func (w *weights) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	buf, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	if len(buf)%4 != 0 {
		return fmt.Errorf("weights: %d bytes is not a whole number of float32s", len(buf))
	}
	*w = make(weights, len(buf)/4)
	for i := range *w {
		(*w)[i] = math.Float32frombits(binary.LittleEndian.Uint32(buf[4*i:]))
	}
	return nil
}

// Genome is everything a creature inherits. Genomes are never mutated after
// a creature is born, so parents and children can share pointers safely.
type Genome struct {
	Traits Traits  `json:"traits"`
	Hidden int     `json:"hidden"` // number of hidden neurons
	WIn    weights `json:"wIn"`    // [input*Hidden + hidden]
	WRec   weights `json:"wRec"`   // [from*Hidden + to]
	WOut   weights `json:"wOut"`   // [hidden*NumOutputs + output]
	BOut   weights `json:"bOut"`
}

func emptyGenome(hidden int) *Genome {
	return &Genome{
		Hidden: hidden,
		WIn:    make(weights, NumInputs*hidden),
		WRec:   make(weights, hidden*hidden),
		WOut:   make(weights, hidden*NumOutputs),
		BOut:   make(weights, NumOutputs),
	}
}

// valid reports whether the weight arrays match the brain size.
func (g *Genome) valid() bool {
	h := g.Hidden
	return h >= 1 && h <= maxHidden && len(g.WIn) == NumInputs*h && len(g.WRec) == h*h &&
		len(g.WOut) == h*NumOutputs && len(g.BOut) == NumOutputs
}

func (g *Genome) clone() *Genome {
	c := emptyGenome(g.Hidden)
	c.Traits = g.Traits
	copy(c.WIn, g.WIn)
	copy(c.WRec, g.WRec)
	copy(c.WOut, g.WOut)
	copy(c.BOut, g.BOut)
	return c
}

func randomGenome(rng *rand.Rand) *Genome { return randomGenomeWithNoise(rng, 1) }

// randomGenomeWithNoise scales the random part of the brain; less noise
// leaves the instincts clearer.
func randomGenomeWithNoise(rng *rand.Rand, noise float64) *Genome {
	return randomGenomeWith(rng, noise, true)
}

// Lifespans in simulated seconds: newborn genomes live 60–80 years, and
// evolution may push that to between 55 and 90.
const (
	lifespanMin     = 55 * SecondsPerYear
	lifespanMax     = 90 * SecondsPerYear
	lifespanInitLo  = 60 * SecondsPerYear
	lifespanInitSpr = 20 * SecondsPerYear
)

// Ranges of the learning genes.
const (
	learningRateMax = 0.3
	traceDecayMin   = 0.3
	traceDecayMax   = 0.95
	plasticityMax   = 3
)

func randomGenomeWith(rng *rand.Rand, noise float64, instincts bool) *Genome {
	g := emptyGenome(initialHidden)
	g.Traits = Traits{
		Hue:          rng.Float64() * 360,
		Size:         0.85 + rng.Float64()*0.3,
		MaxSpeed:     1.5 + rng.Float64(),
		Vision:       4 + rng.Float64()*3,
		Metabolism:   0.9 + rng.Float64()*0.2,
		Lifespan:     lifespanInitLo + rng.Float64()*lifespanInitSpr,
		MutationRate: 0.05,
		// Learning starts as fine-tuning; evolution may turn it up if it pays.
		LearningRate: 0.005 + rng.Float64()*0.015,
		TraceDecay:   0.6 + rng.Float64()*0.3,
		Plasticity:   0.1 + rng.Float64()*0.4,
	}
	fill := func(ws weights, sd float64) {
		for i := range ws {
			ws[i] = float32(rng.NormFloat64() * sd * noise)
		}
	}
	fill(g.WIn, 0.3)
	fill(g.WRec, 0.2)
	fill(g.WOut, 0.3)
	if instincts {
		g.addInstincts()
	}
	return g
}

// addInstincts wires a few reflexes into otherwise random brains so the first
// couple isn't hopeless. Each reflex is one hidden neuron that swings from
// "off" (about -2) to "on" (about +2), so it can switch its output either way.
// Evolution is free to keep, change or erase them.
func (g *Genome) addInstincts() {
	H := g.Hidden
	reflex := func(h int, inputs map[int]float64, output int) {
		for in, w := range inputs {
			g.WIn[in*H+h] += float32(w)
		}
		g.WOut[h*NumOutputs+output] += 2.5
	}
	reflex(0, map[int]float64{inFoodHere: 4, inCarryFood: 3, inEnergy: -4, inBias: 2}, outEat)
	reflex(1, map[int]float64{inWaterNear: 3, inHydration: -5, inBias: 2}, outDrink)
	reflex(2, map[int]float64{inPartnerNear: 2, inFertile: 2, inEnergy: 1, inBias: -2}, outMate)
	reflex(7, map[int]float64{inSourceHere: 2, inEnergy: 3, inCarryMaterial: -3, inBias: -2.5}, outGather)
	reflex(9, map[int]float64{inCanBuild: 3, inEnergy: 2, inBias: -3}, outBuild)
	reflex(10, map[int]float64{inCanCraft: 3, inEnergy: 2, inBias: -3}, outCraft)
	reflex(11, map[int]float64{inFamilyNear: 2.5, inCarryFood: 3, inEnergy: 1, inBias: -3}, outGive)
	// Needs drive: search when hungry or thirsty with nothing at hand, idle when sated.
	reflex(13, map[int]float64{inEnergy: -5, inHydration: -3, inFoodHere: -4, inWaterNear: -2, inBias: 6}, outMove)
	// Show a younger one what you know when you are well fed.
	reflex(14, map[int]float64{inStudentNear: 3, inBestSkill: 2, inEnergy: 1, inBias: -3}, outTeach)

	// Steering: turn towards water, food, the opposite sex and resources, away from walls.
	steer := func(h, channel int, w float64) {
		for r, d := range rayDegrees {
			if d != 0 {
				g.WIn[(channel+r)*H+h] += float32(w * math.Copysign(1, d))
			}
		}
		g.WOut[h*NumOutputs+outTurn] += 1.5
	}
	steer(3, inWater, 1.5)
	steer(4, inFood, 1.5)
	steer(5, inMate, 0.8)
	steer(6, inObstacle, -2)
	steer(12, inResource, 1.5)

	g.BOut[outMove] += 0.5
	g.BOut[outRest] -= 2
	// The first humans are peaceful; evolution may change that.
	g.BOut[outSteal] -= 3
	g.BOut[outAttack] -= 3
	for _, ws := range []weights{g.WIn, g.WOut} {
		for i, w := range ws {
			ws[i] = clamp32(w, -maxWeight, maxWeight)
		}
	}
}

func sigmoid(x float64) float64 { return 1 / (1 + math.Exp(-x)) }

// think runs one step of a recurrent network with input weights wIn and
// output weights wOut (the inherited ones plus anything learned), updating
// hidden in place.
func think(g *Genome, wIn, wOut weights, in *[NumInputs]float64, hidden []float64, out *[NumOutputs]float64) {
	H := g.Hidden
	var buf [maxHidden]float64
	next := buf[:H]
	for i, v := range in {
		if v == 0 {
			continue
		}
		row := wIn[i*H : (i+1)*H]
		for h, w := range row {
			next[h] += v * float64(w)
		}
	}
	for j, v := range hidden {
		if v == 0 {
			continue
		}
		row := g.WRec[j*H : (j+1)*H]
		for h, w := range row {
			next[h] += v * float64(w)
		}
	}
	for h := range next {
		next[h] = math.Tanh(next[h])
	}
	copy(hidden, next)

	for o := range out {
		sum := float64(g.BOut[o])
		for h, v := range next {
			sum += v * float64(wOut[h*NumOutputs+o])
		}
		if o == outTurn {
			out[o] = math.Tanh(sum)
		} else {
			out[o] = sigmoid(sum)
		}
	}
}

// think runs the genome's own (unlearned) network; tests and tools use it.
func (g *Genome) think(in *[NumInputs]float64, hidden []float64, out *[NumOutputs]float64) {
	think(g, g.WIn, g.WOut, in, hidden, out)
}

// neuronSrc says which parent and which of its hidden neurons a child's
// neuron comes from.
type neuronSrc struct {
	p   *Genome
	idx int
}

// crossover builds a child brain neuron by neuron: each hidden neuron keeps
// all its incoming and outgoing weights from one parent, so working circuits
// survive inheritance. Brains of different sizes are aligned by position; a
// neuron only one parent has is inherited half the time.
func crossover(a, b *Genome, rng *rand.Rand) *Genome {
	pick := func() *Genome {
		if rng.IntN(2) == 0 {
			return a
		}
		return b
	}
	var src []neuronSrc
	for h := range max(a.Hidden, b.Hidden) {
		switch {
		case h < a.Hidden && h < b.Hidden:
			src = append(src, neuronSrc{pick(), h})
		case rng.IntN(2) == 0:
			// Only the larger parent has this neuron.
			if h < a.Hidden {
				src = append(src, neuronSrc{a, h})
			} else {
				src = append(src, neuronSrc{b, h})
			}
		}
	}
	for len(src) < minHidden {
		big := a
		if b.Hidden > a.Hidden {
			big = b
		}
		src = append(src, neuronSrc{big, len(src) % big.Hidden})
	}
	c := assemble(src)
	for o := range c.BOut {
		c.BOut[o] = pick().BOut[o]
	}

	ta, tb := a.Traits, b.Traits
	blend := func(x, y float64) float64 { return x + (y-x)*rng.Float64() }
	c.Traits = Traits{
		Hue:          blendHue(ta.Hue, tb.Hue, rng.Float64()),
		Size:         blend(ta.Size, tb.Size),
		MaxSpeed:     blend(ta.MaxSpeed, tb.MaxSpeed),
		Vision:       blend(ta.Vision, tb.Vision),
		Metabolism:   blend(ta.Metabolism, tb.Metabolism),
		Lifespan:     blend(ta.Lifespan, tb.Lifespan),
		MutationRate: blend(ta.MutationRate, tb.MutationRate),
		LearningRate: blend(ta.LearningRate, tb.LearningRate),
		TraceDecay:   blend(ta.TraceDecay, tb.TraceDecay),
		Plasticity:   blend(ta.Plasticity, tb.Plasticity),
	}
	return c
}

// assemble builds a genome whose hidden neurons are copies of the given
// source neurons. A recurrent weight into neuron k comes from k's parent; it
// is the weight from the same position there, or 0 if that parent is smaller.
func assemble(src []neuronSrc) *Genome {
	H := len(src)
	c := emptyGenome(H)
	for k, s := range src {
		p, ph := s.p, s.p.Hidden
		for i := range NumInputs {
			c.WIn[i*H+k] = p.WIn[i*ph+s.idx]
		}
		copy(c.WOut[k*NumOutputs:(k+1)*NumOutputs], p.WOut[s.idx*NumOutputs:(s.idx+1)*NumOutputs])
		for j, sj := range src {
			from := sj.idx
			if sj.p != p && from >= ph {
				continue
			}
			c.WRec[j*H+k] = p.WRec[from*ph+s.idx]
		}
	}
	return c
}

// blendHue interpolates two hues around the colour wheel.
func blendHue(a, b, t float64) float64 {
	ra, rb := a*math.Pi/180, b*math.Pi/180
	x := (1-t)*math.Cos(ra) + t*math.Cos(rb)
	y := (1-t)*math.Sin(ra) + t*math.Sin(rb)
	return math.Mod(math.Atan2(y, x)*180/math.Pi+360, 360)
}

func clamp(v, lo, hi float64) float64 { return math.Min(hi, math.Max(lo, v)) }

func clamp32(v, lo, hi float32) float32 { return min(hi, max(lo, v)) }

// Chances per child that the brain gains or loses a hidden neuron.
const (
	growChance   = 0.05
	shrinkChance = 0.05
)

// mutate perturbs weights and traits, and may grow or shrink the brain. The
// mutation rate is itself inherited and mutated, so lineages can evolve how
// fast they evolve. It returns the (possibly resized) genome.
func (g *Genome) mutate(rng *rand.Rand) *Genome {
	t := &g.Traits
	t.MutationRate = clamp(t.MutationRate*math.Exp(rng.NormFloat64()*0.2), 0.01, 0.2)
	rate := t.MutationRate
	perturb := func(ws weights) {
		for i := range ws {
			if rng.Float64() >= rate {
				continue
			}
			v := float64(ws[i])
			if rng.Float64() < 0.05 {
				v = rng.NormFloat64()
			} else {
				v += rng.NormFloat64() * 0.4
			}
			ws[i] = float32(clamp(v, -maxWeight, maxWeight))
		}
	}
	perturb(g.WIn)
	perturb(g.WRec)
	perturb(g.WOut)
	perturb(g.BOut)

	t.Hue = math.Mod(t.Hue+rng.NormFloat64()*6+360, 360)
	t.Size = clamp(t.Size+rng.NormFloat64()*0.03, 0.7, 1.3)
	t.MaxSpeed = clamp(t.MaxSpeed+rng.NormFloat64()*0.08, 1, 3)
	t.Vision = clamp(t.Vision+rng.NormFloat64()*0.25, 3, 9)
	t.Metabolism = clamp(t.Metabolism+rng.NormFloat64()*0.03, 0.7, 1.3)
	t.Lifespan = clamp(t.Lifespan+rng.NormFloat64()*12, lifespanMin, lifespanMax)
	t.LearningRate = clamp(t.LearningRate+rng.NormFloat64()*0.01, 0, learningRateMax)
	t.TraceDecay = clamp(t.TraceDecay+rng.NormFloat64()*0.03, traceDecayMin, traceDecayMax)
	t.Plasticity = clamp(t.Plasticity+rng.NormFloat64()*0.1, 0, plasticityMax)

	out := g
	if out.Hidden < maxHidden && rng.Float64() < growChance {
		out = out.withNeuron(rng)
	}
	if out.Hidden > minHidden && rng.Float64() < shrinkChance {
		out = out.withoutNeuron(out.weakestNeuron())
	}
	return out
}

// withNeuron adds a hidden neuron that listens to the inputs but says nothing
// yet (zero outgoing weights), so the brain behaves as before until mutation
// or learning gives it a voice.
func (g *Genome) withNeuron(rng *rand.Rand) *Genome {
	src := make([]neuronSrc, g.Hidden)
	for h := range src {
		src[h] = neuronSrc{g, h}
	}
	c := assemble(src)
	n := emptyGenome(g.Hidden + 1)
	n.Traits, n.BOut = g.Traits, slices32(g.BOut)
	H, nH := g.Hidden, g.Hidden+1
	for i := range NumInputs {
		copy(n.WIn[i*nH:i*nH+H], c.WIn[i*H:(i+1)*H])
		n.WIn[i*nH+H] = float32(rng.NormFloat64() * 0.2)
	}
	for j := range H {
		copy(n.WRec[j*nH:j*nH+H], c.WRec[j*H:(j+1)*H])
		n.WRec[j*nH+H] = float32(rng.NormFloat64() * 0.1) // others feed the new neuron
	}
	copy(n.WOut, c.WOut)
	return n
}

// withoutNeuron removes hidden neuron k.
func (g *Genome) withoutNeuron(k int) *Genome {
	src := make([]neuronSrc, 0, g.Hidden-1)
	for h := range g.Hidden {
		if h != k {
			src = append(src, neuronSrc{g, h})
		}
	}
	n := assemble(src)
	n.Traits, n.BOut = g.Traits, slices32(g.BOut)
	return n
}

// weakestNeuron is the hidden neuron with the least influence on the rest of
// the brain and on behaviour.
func (g *Genome) weakestNeuron() int {
	H := g.Hidden
	best, bestW := 0, math.Inf(1)
	for h := range H {
		w := 0.0
		for o := range NumOutputs {
			w += math.Abs(float64(g.WOut[h*NumOutputs+o]))
		}
		for k := range H {
			w += math.Abs(float64(g.WRec[h*H+k]))
		}
		if w < bestW {
			best, bestW = h, w
		}
	}
	return best
}

func slices32(w weights) weights { return append(weights(nil), w...) }
