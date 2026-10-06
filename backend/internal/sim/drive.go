package sim

import "math"

// The foraging drive: four inborn neurons that make hunger send a body out
// to look for food and turn it towards food it remembers (forage.go), and
// make thirst, which kills sooner, send it to water first.
// Hunger-driven foraging is innate in every animal that forages, humans
// included; what is learned is where and how. The first couple carried a
// similar reflex ("search when hungry with no food in reach", see
// addInstincts), but over two hundred generations it had all but vanished:
// in the live world most people who starved wanted to eat yet stood still
// with food a few tiles off (PLAN.md, "Dorongan mencari makan"). Save v12
// gives every brain these neurons once; after that they are ordinary
// neurons that crossover passes on and mutation may change or remove.

// driveNeurons is how many neurons the drive takes.
const driveNeurons = 4

// addDrive wires the drive into neurons at … at+3 of g, on top of whatever
// they hold.
func (g *Genome) addDrive(at int) {
	H := g.Hidden
	in := func(h, input int, w float64) { g.WIn[input*H+h] += float32(w) }
	out := func(h, output int, w float64) { g.WOut[h*NumOutputs+output] += float32(w) }

	// Hungry with nothing at hand: walk, and look to eat; more so when a
	// remembered place calls. Weights stay within maxWeight, so the neuron
	// is held near -1 by a full stomach (input about -2 or less) and opens
	// as energy falls below about 0.6; the two outputs' biases are raised
	// by what it takes away when shut, so being full changes nothing that
	// was there before.
	// Thirst quietens it: the hydration term is zero for a body with water
	// enough and closes the neuron as water runs short.
	f := at
	in(f, inEnergy, -4)
	in(f, inFoodHere, -4)
	in(f, inCarryFood, -3)
	in(f, inFoodMemory, 1.5)
	in(f, inHydration, 2)
	in(f, inBias, -1.5)
	out(f, outMove, 2)
	out(f, outEat, 1.5)
	g.BOut[outMove] += 2
	g.BOut[outEat] += 1.5

	// Turn towards remembered food: a pair that sees its side with opposite
	// signs behind a shared hunger gate. Fed, the gate holds both near -1
	// and their pulls cancel; below about 60% energy it opens.
	for i, sign := range []float64{1, -1} {
		h := at + 1 + i
		in(h, inEnergy, -4)
		in(h, inHydration, 2)
		in(h, inBias, -1.5)
		in(h, inFoodMemorySide, 1.5*sign)
		out(h, outTurn, 1.5*sign)
	}

	// Thirsty with no water at hand or in a tube: walk, and want to drink,
	// more than hunger wants to eat, so a body that is both makes for water
	// first (thirst kills in days, hunger in weeks). Shut near -1 above
	// about 60% hydration, with the biases raised to match as above.
	w := at + 3
	in(w, inHydration, -4)
	in(w, inWaterNear, -4)
	in(w, inCarryWater, -3)
	in(w, inBias, 2)
	out(w, outMove, 2)
	out(w, outDrink, 2.5)
	g.BOut[outMove] += 2
	g.BOut[outDrink] += 2.5
	for _, ws := range []weights{g.WIn, g.WOut, g.BOut} {
		for i, w := range ws {
			ws[i] = clamp32(w, -maxWeight, maxWeight)
		}
	}
}

// withNeurons returns g with k silent neurons appended: no weights in or
// out, no bias, time constant 1.
func (g *Genome) withNeurons(k int) *Genome {
	H, nH := g.Hidden, g.Hidden+k
	n := &Genome{
		Traits: g.Traits,
		Hidden: nH,
		WIn:    padRows(g.WIn, H, nH),
		WRec:   append(padRows(g.WRec, H, nH), make(weights, k*nH)...),
		WOut:   append(slices32(g.WOut), make(weights, k*NumOutputs)...),
		BOut:   slices32(g.BOut),
		Bias:   append(slices32(g.Bias), make(weights, k)...),
		Tau:    append(slices32(g.Tau), filled(k, minTau)...),
		Loci:   g.Loci,
	}
	return n
}

// withNeurons widens a mind for k neurons appended to its genome (whose
// hidden size was H): nothing learned about them yet.
func (m *Mind) withNeurons(H, k int) {
	nH := H + k
	m.DIn, m.EIn = padRows(m.DIn, H, nH), padRows(m.EIn, H, nH)
	m.DOut = append(m.DOut, make(weights, k*NumOutputs)...)
	m.EOut = append(m.EOut, make(weights, k*NumOutputs)...)
	if m.Grown > 0 {
		m.GRec = padRows(m.GRec, H, nH)
	}
}

// giveDrive is the save-v12 upgrade of one person: three drive neurons in
// their genome (shared genomes stay shared, via seen) and room for them in
// their mind and activity.
func giveDrive(c *Creature, seen map[*Genome]*Genome) {
	g := c.Genome
	H := g.Hidden
	c.Genome = droveGenome(g, seen)
	c.Mind.withNeurons(H, driveNeurons)
	c.Hidden = append(c.Hidden, make([]float64, driveNeurons)...)
	if p := c.Pregnancy; p != nil && p.FatherGenome != nil {
		p.FatherGenome = droveGenome(p.FatherGenome, seen)
	}
}

func droveGenome(g *Genome, seen map[*Genome]*Genome) *Genome {
	if d, ok := seen[g]; ok {
		return d
	}
	d := g.withNeurons(driveNeurons)
	d.addDrive(g.Hidden)
	seen[g] = d
	return d
}

// --- telling where food is -------------------------------------------------------

// tellFood passes on the teller's best food place in a talk, believed as
// much as the teller is (see belief), unless the listener knows it already.
// Foragers tell each other where food is; Hadza and Ju/'hoansi camps share
// news of ripe trees and game this way (Marlowe 2010; Wiessner 2014).
func (s *Sim) tellFood(teller, listener *Creature, belief float64) {
	p, v := s.bestFoodPlace(teller)
	if v <= 0 {
		return
	}
	for _, q := range listener.FoodPlaces {
		if math.Hypot(q.X-p.X, q.Y-p.Y) < foodPlaceMerge {
			return
		}
	}
	// Second-hand it counts for less, the more so the less the teller is
	// trusted; only a good place is remembered (see rememberFood).
	s.rememberFood(listener, p.X, p.Y, s.foodPlaceWorth(p)*(0.5+0.5*belief))
}
