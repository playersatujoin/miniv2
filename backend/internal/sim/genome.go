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
	numChannels = 7
	numInternal = 64

	// legacyInputs is how many senses brains had before Fase 3 (saves up to
	// version 7); their weights are padded with zeros for the senses added since.
	legacyInputs = numRays*numChannels + 36
	healthInputs = numRays*numChannels + 42 // saves v8 and v9
	alarmInputs  = numRays*numChannels + 43 // saves v10
	engineInputs = numRays*numChannels + 61 // saves v11

	NumInputs  = numRays*numChannels + numInternal
	NumOutputs = 17
	// legacyOutputs is how many actions brains had up to save v10; the
	// actions added since start with no weights.
	legacyOutputs = 15

	// Brain size is inherited and evolves: hidden neurons per genome. More
	// grow during a life (see growth.go). There is no ceiling: every neuron
	// costs energy (brainCost), and that is what limits brains, as in life,
	// where the brain burns about a fifth of a resting body's energy (Raichle
	// & Gusnard 2002). brainSanity only rejects corrupt saves.
	minHidden     = 8
	initialHidden = 20 + driveNeurons
	brainSanity   = 1 << 14
	// stackNeurons: brains up to this size think without allocating.
	stackNeurons = 128

	// A neuron's time constant, in ticks: 1 follows its inputs at once, more
	// lets it hold on to what it saw (a working memory).
	minTau = 1
	maxTau = 20

	maxWeight = 4
)

// Input layout: seven ray channels (channel-major), then internal senses.
const (
	inObstacle = iota * numRays
	inFood
	inWater
	inMate
	inSame
	inResource
	inAnimal
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
	inSeason
	inLight
	inCanPlant
	inCropReady
	inLivestockHungry
	inPredatorNear
	inClock
	inNoise
	inBias
	// Fase 3, appended so older brains keep their wiring.
	inChildHungry  // a hungry child of the family nearby
	inNursing      // breastfeeding her youngest
	inSick         // feverish (Fase 3b)
	inMalnourished // short of protein or vitamins (Fase 3d)
	inMateKin      // the nearest possible mate is a close relative (Fase 3c)
	inMateHealth   // … and how healthy they look
	inAlarm        // a recent perceived assault, without prescribing a reaction
	// Engine adaptation II (save v11), appended likewise. Each says what is
	// there, never what to do about it.
	inShock      // the most striking thing seen or heard lately: a body, a fire, a fight (stimuli.go)
	inShockSide  // where it is: -1 on the left … 1 on the right
	inFear       // moods that rise with what happens and ebb back (affect.go)
	inAnger      //
	inJoy        //
	inGrief      //
	inTalkNear   // someone close by, facing me, wants to talk (interact.go)
	inTradeNear  // someone close by, facing me, offers to trade
	inOwnVillage // standing in my village's land (village.go)
	inLeaderNear // the person my village trusts most, how close
	inLeaderSide // … and on which side
	inIsLeader   // I am that person
	inFireNear   // fire in sight, how close (fire.go)
	inWind       // how hard the wind blows
	inInWater    // 0 on dry land, 0.5 wading, 1 out of my depth (locomotion.go)
	inOnRaft     // afloat on a raft
	inSeaAhead   // deep water straight ahead, how close
	inClimbAhead // steep ground to climb straight ahead, how close
	// Save v12 (forage.go): what they know of food beyond sight, and water carried.
	inFoodMemory     // the remembered food place most worth the walk, how much
	inFoodMemorySide // … and on which side
	inCarryWater     // water carried in a tube, share of what the tubes hold
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
	outPlant
	outHunt
	// Save v11: two-person exchanges, which only happen when both want them
	// (interact.go).
	outTalk
	outTrade
)

var rayDegrees = [numRays]float64{-60, -30, 0, 30, 60}

var InputLabels = func() []string {
	var out []string
	for _, ch := range []string{"rintangan", "makanan", "air", "lawan jenis", "sesama jenis", "sumber daya", "hewan"} {
		for _, d := range rayDegrees {
			out = append(out, fmt.Sprintf("%s %g°", ch, d))
		}
	}
	return append(out, "energi", "hidrasi", "kesehatan", "usia", "subur", "hamil", "makanan di sini",
		"air di dekat", "sumber di sini", "menabrak", "pasangan dekat", "keluarga dekat", "orang asing dekat",
		"kepercayaanku pada orang dekat", "diserang", "bawa makanan", "bawa bahan", "bawa senjata", "rumah sendiri dekat",
		"rumah orang lain dekat", "bisa membuat", "bisa membangun", "guru dekat", "murid dekat",
		"keahlian tertinggi", "imbalan terakhir", "perpustakaan dekat", "musim", "cahaya", "bisa menanam",
		"tanaman siap panen", "ternak lapar", "pemangsa dekat", "jam internal", "acak (kehendak)", "bias",
		"anak lapar dekat", "menyusui", "sakit", "kurang gizi", "calon pasangan kerabat", "kesehatan calon pasangan", "bahaya yang terdengar atau terlihat",
		"kejadian mengejutkan", "arah kejadian", "takut", "marah", "senang", "duka",
		"ada yang mengajak bicara", "ada yang menawarkan tukar", "di wilayah desa sendiri", "pemimpin dekat",
		"arah pemimpin", "aku pemimpin", "api terlihat", "angin", "di dalam air", "di atas rakit",
		"laut di depan", "tanjakan di depan", "makanan yang diingat", "arah makanan yang diingat", "bawa air")
}()

var OutputLabels = []string{"belok", "gerak", "makan", "minum", "kawin", "istirahat",
	"kumpulkan", "buat", "bangun", "beri", "curi", "serang", "ajar", "tanam", "buru", "bicara", "tukar"}

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

	// Neurogenesis is how readily new neurons grow during life when the world
	// keeps surprising (0 = never; see growth.go).
	Neurogenesis float64 `json:"neurogenesis"`
	// Menopause is the age (years) at which a woman's fertility ends; men
	// carry the gene too and pass it on.
	Menopause float64 `json:"menopause"`
	// Immunity is how strong the body's defences against infection are
	// (about 1); stronger ones cost energy (see disease.go).
	Immunity float64 `json:"immunity"`
	// Temperament (save v11, see affect.go), about 1 each: how strongly
	// what happens moves the moods, how fast they ebb back, and how
	// cheerful the resting mood is.
	Reactivity float64 `json:"reactivity"`
	Recovery   float64 `json:"recovery"`
	Cheer      float64 `json:"cheer"`
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
	// Per hidden neuron: its bias and its time constant (see minTau). Saves
	// from before version 7 have neither; they mean bias 0 and tau 1.
	Bias weights `json:"bias,omitempty"`
	Tau  weights `json:"tau,omitempty"`
	// Loci: the Mendelian loci, two copies each (genetics.go); nil in
	// genomes from before Fase 3c.
	Loci *Loci `json:"loci,omitempty"`
}

func emptyGenome(hidden int) *Genome {
	return &Genome{
		Hidden: hidden,
		WIn:    make(weights, NumInputs*hidden),
		WRec:   make(weights, hidden*hidden),
		WOut:   make(weights, hidden*NumOutputs),
		BOut:   make(weights, NumOutputs),
		Bias:   make(weights, hidden),
		Tau:    filled(hidden, minTau),
	}
}

func filled(n int, v float32) weights {
	w := make(weights, n)
	for i := range w {
		w[i] = v
	}
	return w
}

// valid reports whether the weight arrays match the brain size.
func (g *Genome) valid() bool {
	h := g.Hidden
	return h >= 1 && h <= brainSanity && len(g.WIn) == NumInputs*h && len(g.WRec) == h*h &&
		len(g.WOut) == h*NumOutputs && len(g.BOut) == NumOutputs && len(g.Bias) == h && len(g.Tau) == h
}

// upgrade fills in what a save from before version 7 lacks: neuron biases
// and time constants, and the neurogenesis gene.
func (g *Genome) upgrade(version int) {
	if g.Bias == nil {
		g.Bias = make(weights, g.Hidden)
	}
	if g.Tau == nil {
		g.Tau = filled(g.Hidden, minTau)
	}
	if version < 7 {
		g.Traits.Neurogenesis = neurogenesisInit
	}
	if n := len(g.WIn); n == legacyInputs*g.Hidden || n == healthInputs*g.Hidden || n == alarmInputs*g.Hidden || n == engineInputs*g.Hidden {
		g.WIn = padInputs(g.WIn, g.Hidden)
	}
	if len(g.WOut) == legacyOutputs*g.Hidden {
		g.WOut = padRows(g.WOut, legacyOutputs, NumOutputs)
	}
	if len(g.BOut) == legacyOutputs {
		g.BOut = padRows(g.BOut, legacyOutputs, NumOutputs)
	}
	if g.Traits.Menopause == 0 {
		g.Traits.Menopause = menopauseInit + 2
	}
	if g.Traits.Immunity == 0 {
		g.Traits.Immunity = 1
	}
	g.Traits.upgradeTemperament()
}

// padInputs widens input-major weights (one row of `width` per sense) from
// the legacy senses to today's, the new senses starting unwired.
func padInputs(w weights, width int) weights {
	out := make(weights, NumInputs*width)
	copy(out, w)
	return out
}

// padRows widens row-major weights from rows of `from` to rows of `to`
// columns, the new columns starting at zero: hidden-major output weights
// (one row of actions per neuron) and neuron-major sense weights alike.
func padRows(w weights, from, to int) weights {
	n := len(w) / from
	out := make(weights, n*to)
	for k := range n {
		copy(out[k*to:], w[k*from:(k+1)*from])
	}
	return out
}

func (g *Genome) clone() *Genome {
	c := emptyGenome(g.Hidden)
	c.Traits, c.Loci = g.Traits, g.Loci
	copy(c.WIn, g.WIn)
	copy(c.WRec, g.WRec)
	copy(c.WOut, g.WOut)
	copy(c.BOut, g.BOut)
	copy(c.Bias, g.Bias)
	copy(c.Tau, g.Tau)
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
		Neurogenesis: neurogenesisInit,
		Menopause:    menopauseInit + rng.Float64()*4,
		Immunity:     immunityInit + rng.Float64()*0.2,
	}
	g.Traits.initTemperament()
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
	reflex(7, map[int]float64{inSourceHere: 2, inCropReady: 2.5, inEnergy: 3, inCarryMaterial: -3, inBias: -2.5}, outGather)
	// Build and craft only on a full stomach: the work holds you still.
	reflex(9, map[int]float64{inCanBuild: 3, inEnergy: 4, inBias: -5.5}, outBuild)
	reflex(10, map[int]float64{inCanCraft: 3, inEnergy: 4, inBias: -5.5}, outCraft)
	reflex(11, map[int]float64{inFamilyNear: 2.5, inChildHungry: 3, inLivestockHungry: 2.5, inCarryFood: 3, inEnergy: 1, inBias: -3}, outGive)
	// Needs drive: search when hungry with no food in reach, or thirsty with
	// no water near; stay put to eat or drink, idle when sated.
	reflex(13, map[int]float64{inEnergy: -5, inFoodHere: -4, inBias: 3.5}, outMove)
	reflex(8, map[int]float64{inHydration: -6, inWaterNear: -4, inBias: 3}, outMove)
	// Show a younger one what you know, but only when well fed (above about
	// 0.7): teaching holds you still, and a hungry teacher starves.
	reflex(14, map[int]float64{inStudentNear: 3, inBestSkill: 2, inEnergy: 6, inBias: -7.8}, outTeach)
	// Put a seed or tuber in the ground when it would grow here.
	reflex(15, map[int]float64{inCanPlant: 4, inEnergy: 1, inBias: -2.5}, outPlant)
	// Hunt an animal in front when hungry, more readily with a spear.
	reflex(16, map[int]float64{inAnimal + 2: 3, inEnergy: -2.5, inCarryWeapon: 1.5, inBias: -1}, outHunt)

	// Steering: turn towards water, food, the opposite sex and resources, away from walls.
	steer := func(h, channel int, w float64) {
		for r, d := range rayDegrees {
			if d != 0 {
				g.WIn[(channel+r)*H+h] += float32(w * math.Copysign(1, d))
			}
		}
		g.WOut[h*NumOutputs+outTurn] += 1.5
	}
	// Water and food pull only when needed. Each pull is a pair of neurons
	// that see the rays with opposite signs and share a gate: when sated the
	// gate drives both to the same extreme and their pulls cancel; when
	// thirsty or hungry the gate opens and the pair steers.
	steerGated := func(h1, h2, channel int, w float64, need int, needW, open float64) {
		for r, d := range rayDegrees {
			if d != 0 {
				v := float32(w * math.Copysign(1, d))
				g.WIn[(channel+r)*H+h1] += v
				g.WIn[(channel+r)*H+h2] -= v
			}
		}
		for _, h := range []int{h1, h2} {
			g.WIn[need*H+h] += float32(needW)
			g.WIn[inBias*H+h] += float32(open)
		}
		g.WOut[h1*NumOutputs+outTurn] += 1.5
		g.WOut[h2*NumOutputs+outTurn] -= 1.5
	}
	steerGated(3, 18, inWater, 1.5, inHydration, -6, 2)
	steerGated(4, 19, inFood, 1.5, inEnergy, -6, 4)
	steer(5, inMate, 0.8)
	steer(6, inObstacle, -2)
	steer(12, inResource, 1.5)
	steer(17, inAnimal, 0.6)

	// Hunger sends them out foraging and towards food they remember (drive.go).
	g.addDrive(20)

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
// hidden in place. Each neuron is a leaky integrator: it moves a 1/tau share
// of the way to tanh(input + bias), so a slow one keeps a memory of the last
// moments. Neurons grown during life (m, may be nil) come after: they listen
// to the senses and to the inherited neurons' new state, and add to the
// outputs.
func think(g *Genome, wIn, wOut weights, m *Mind, in *[NumInputs]float64, hidden []float64, out *[NumOutputs]float64) {
	H := g.Hidden
	var buf [stackNeurons]float64
	var next []float64
	if H <= stackNeurons {
		next = buf[:H]
	} else {
		next = make([]float64, H)
	}
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
		a := math.Tanh(next[h] + float64(g.Bias[h]))
		next[h] = hidden[h] + (a-hidden[h])/float64(g.Tau[h])
	}
	copy(hidden, next)

	for o := range out {
		out[o] = float64(g.BOut[o])
	}
	for h, v := range next {
		for o, w := range wOut[h*NumOutputs : (h+1)*NumOutputs] {
			out[o] += v * float64(w)
		}
	}
	if m != nil && m.Grown > 0 {
		m.thinkGrown(in, next, out)
	}
	for o := range out {
		if o == outTurn {
			out[o] = math.Tanh(out[o])
		} else {
			out[o] = sigmoid(out[o])
		}
	}
}

// think runs the genome's own (unlearned) network; tests and tools use it.
func (g *Genome) think(in *[NumInputs]float64, hidden []float64, out *[NumOutputs]float64) {
	think(g, g.WIn, g.WOut, nil, in, hidden, out)
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
		Neurogenesis: blend(ta.Neurogenesis, tb.Neurogenesis),
		Menopause:    blend(ta.Menopause, tb.Menopause),
		Immunity:     blend(ta.Immunity, tb.Immunity),
	}
	c.Traits.blendTemperament(&ta, &tb)
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
		c.Bias[k], c.Tau[k] = p.Bias[s.idx], p.Tau[s.idx]
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

// clamp32 limits v to [lo, hi]. Comparisons rather than min and max, which
// spend extra work on NaN and signed zeros: this runs for every weight on
// every learning step, and its bounds are never zero of the wrong sign.
func clamp32(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

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
	perturb(g.Bias)
	// Time constants drift on a log scale, so slow memories can evolve from fast reflexes.
	for i := range g.Tau {
		if rng.Float64() < rate {
			g.Tau[i] = float32(clamp(float64(g.Tau[i])*math.Exp(rng.NormFloat64()*0.3), minTau, maxTau))
		}
	}

	t.Hue = math.Mod(t.Hue+rng.NormFloat64()*6+360, 360)
	t.Size = clamp(t.Size+rng.NormFloat64()*0.03, 0.7, 1.3)
	t.MaxSpeed = clamp(t.MaxSpeed+rng.NormFloat64()*0.08, 1, 3)
	t.Vision = clamp(t.Vision+rng.NormFloat64()*0.25, 3, 9)
	t.Metabolism = clamp(t.Metabolism+rng.NormFloat64()*0.03, 0.7, 1.3)
	t.Lifespan = clamp(t.Lifespan+rng.NormFloat64()*12, lifespanMin, lifespanMax)
	t.LearningRate = clamp(t.LearningRate+rng.NormFloat64()*0.01, 0, learningRateMax)
	t.TraceDecay = clamp(t.TraceDecay+rng.NormFloat64()*0.03, traceDecayMin, traceDecayMax)
	t.Plasticity = clamp(t.Plasticity+rng.NormFloat64()*0.1, 0, plasticityMax)
	t.Neurogenesis = clamp(t.Neurogenesis+rng.NormFloat64()*0.05, 0, neurogenesisMax)
	t.Menopause = clamp(t.Menopause+rng.NormFloat64()*0.5, menopauseMin, menopauseMax)
	t.Immunity = clamp(t.Immunity+rng.NormFloat64()*0.03, immuneMin, immuneMax)
	t.mutateTemperament()

	out := g
	if rng.Float64() < growChance {
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
	n.Traits, n.BOut, n.Loci = g.Traits, slices32(g.BOut), g.Loci
	H, nH := g.Hidden, g.Hidden+1
	copy(n.Bias, c.Bias)
	copy(n.Tau, c.Tau)
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
	n.Traits, n.BOut, n.Loci = g.Traits, slices32(g.BOut), g.Loci
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
