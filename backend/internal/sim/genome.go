package sim

import (
	"fmt"
	"math"
	"math/rand/v2"
)

const (
	numRays     = 5
	numChannels = 6
	numInternal = 25

	NumInputs  = numRays*numChannels + numInternal
	NumHidden  = 24
	NumOutputs = 12

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
		"rumah orang lain dekat", "bisa membuat", "bisa membangun", "jam internal", "acak (kehendak)", "bias")
}()

var OutputLabels = []string{"belok", "gerak", "makan", "minum", "kawin", "istirahat",
	"kumpulkan", "buat", "bangun", "beri", "curi", "serang"}

type Traits struct {
	Hue          float64 `json:"hue"`
	Size         float64 `json:"size"`
	MaxSpeed     float64 `json:"maxSpeed"`
	Vision       float64 `json:"vision"`
	Metabolism   float64 `json:"metabolism"`
	Lifespan     float64 `json:"lifespan"`
	MutationRate float64 `json:"mutationRate"`
}

// Genome is everything a creature inherits. Genomes are never mutated after
// a creature is born, so parents and children can share pointers safely.
type Genome struct {
	Traits Traits    `json:"traits"`
	WIn    []float64 `json:"wIn"`  // [input*NumHidden + hidden]
	WRec   []float64 `json:"wRec"` // [from*NumHidden + to]
	WOut   []float64 `json:"wOut"` // [hidden*NumOutputs + output]
	BOut   []float64 `json:"bOut"`
}

func emptyGenome() *Genome {
	return &Genome{
		WIn:  make([]float64, NumInputs*NumHidden),
		WRec: make([]float64, NumHidden*NumHidden),
		WOut: make([]float64, NumHidden*NumOutputs),
		BOut: make([]float64, NumOutputs),
	}
}

func (g *Genome) clone() *Genome {
	c := emptyGenome()
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

func randomGenomeWith(rng *rand.Rand, noise float64, instincts bool) *Genome {
	g := emptyGenome()
	g.Traits = Traits{
		Hue:          rng.Float64() * 360,
		Size:         0.85 + rng.Float64()*0.3,
		MaxSpeed:     1.5 + rng.Float64(),
		Vision:       4 + rng.Float64()*3,
		Metabolism:   0.9 + rng.Float64()*0.2,
		Lifespan:     lifespanInitLo + rng.Float64()*lifespanInitSpr,
		MutationRate: 0.05,
	}
	for i := range g.WIn {
		g.WIn[i] = rng.NormFloat64() * 0.3 * noise
	}
	for i := range g.WRec {
		g.WRec[i] = rng.NormFloat64() * 0.2 * noise
	}
	for i := range g.WOut {
		g.WOut[i] = rng.NormFloat64() * 0.3 * noise
	}
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
	reflex := func(h int, inputs map[int]float64, output int) {
		for in, w := range inputs {
			g.WIn[in*NumHidden+h] += w
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

	// Steering: turn towards water, food, the opposite sex and resources, away from walls.
	steer := func(h, channel int, w float64) {
		for r, d := range rayDegrees {
			if d != 0 {
				g.WIn[(channel+r)*NumHidden+h] += w * math.Copysign(1, d)
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
	for _, ws := range [][]float64{g.WIn, g.WOut} {
		for i, w := range ws {
			ws[i] = clamp(w, -maxWeight, maxWeight)
		}
	}
}

func sigmoid(x float64) float64 { return 1 / (1 + math.Exp(-x)) }

// think runs one step of the recurrent network, updating hidden in place.
func (g *Genome) think(in *[NumInputs]float64, hidden *[NumHidden]float64, out *[NumOutputs]float64) {
	var next [NumHidden]float64
	for i, v := range in {
		if v == 0 {
			continue
		}
		row := g.WIn[i*NumHidden : (i+1)*NumHidden]
		for h, w := range row {
			next[h] += v * w
		}
	}
	for j, v := range hidden {
		row := g.WRec[j*NumHidden : (j+1)*NumHidden]
		for h, w := range row {
			next[h] += v * w
		}
	}
	for h := range next {
		next[h] = math.Tanh(next[h])
	}
	*hidden = next

	for o := range out {
		sum := g.BOut[o]
		for h, v := range next {
			sum += v * g.WOut[h*NumOutputs+o]
		}
		if o == outTurn {
			out[o] = math.Tanh(sum)
		} else {
			out[o] = sigmoid(sum)
		}
	}
}

// crossover builds a child brain neuron by neuron: each hidden neuron keeps
// all its incoming and outgoing weights from one parent, so working circuits
// survive inheritance.
func crossover(a, b *Genome, rng *rand.Rand) *Genome {
	pick := func() *Genome {
		if rng.IntN(2) == 0 {
			return a
		}
		return b
	}
	c := emptyGenome()
	for h := range NumHidden {
		p := pick()
		for i := range NumInputs {
			c.WIn[i*NumHidden+h] = p.WIn[i*NumHidden+h]
		}
		for j := range NumHidden {
			c.WRec[j*NumHidden+h] = p.WRec[j*NumHidden+h]
		}
		copy(c.WOut[h*NumOutputs:(h+1)*NumOutputs], p.WOut[h*NumOutputs:(h+1)*NumOutputs])
	}
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

// mutate perturbs weights and traits. The mutation rate is itself inherited
// and mutated, so lineages can evolve how fast they evolve.
func (g *Genome) mutate(rng *rand.Rand) {
	t := &g.Traits
	t.MutationRate = clamp(t.MutationRate*math.Exp(rng.NormFloat64()*0.2), 0.01, 0.2)
	rate := t.MutationRate
	perturb := func(ws []float64) {
		for i := range ws {
			if rng.Float64() >= rate {
				continue
			}
			if rng.Float64() < 0.05 {
				ws[i] = rng.NormFloat64()
			} else {
				ws[i] += rng.NormFloat64() * 0.4
			}
			ws[i] = clamp(ws[i], -maxWeight, maxWeight)
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
}
