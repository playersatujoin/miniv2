package chem

import "slices"

const (
	catAlkali         = "logam alkali"
	catAlkaline       = "logam alkali tanah"
	catLanthanide     = "lantanida"
	catActinide       = "aktinida"
	catTransition     = "logam transisi"
	catPostTransition = "logam pasca-transisi"
	catMetalloid      = "metaloid"
	catNonmetal       = "nonlogam"
	catHalogen        = "halogen"
	catNoble          = "gas mulia"
)

const (
	solid  = "padat"
	liquid = "cair"
	gas    = "gas"
)

// Tier needed to isolate each element. Tiers 6–7 are synthetic.
var elementTiers = [MaxTier + 1][]string{
	{"C", "S", "Cu", "Ag", "Au"},
	{"Fe", "Sn", "Pb", "Zn", "Hg", "As", "Sb", "Bi"},
	{"H", "N", "O", "Cl", "P", "Co", "Ni", "Mn", "Mo", "W", "Te", "Cr", "U", "Ti", "Y", "Be", "Zr", "Pt", "Pd", "Rh", "Os", "Ir", "Nb", "Ta", "Ce"},
	{"Na", "K", "Ca", "Mg", "Ba", "Sr", "B", "Li", "Al", "Si", "Cd", "I", "Br", "Se", "Th", "V", "La", "Ru", "Er", "Tb", "F"},
	{"Cs", "Rb", "Tl", "In", "Ga", "He", "Ar", "Ne", "Kr", "Xe", "Ge", "Sc", "Ho", "Tm", "Sm", "Gd", "Pr", "Nd", "Dy", "Yb", "Lu", "Eu"},
	{"Po", "Ra", "Rn", "Ac", "Pa", "Hf", "Re", "Fr"},
	{"Tc", "Pm", "At", "Np", "Pu", "Am", "Cm", "Bk", "Cf", "Es", "Fm"},
	{"Md", "No", "Lr", "Rf", "Db", "Sg", "Bh", "Hs", "Mt", "Ds", "Rg", "Cn", "Nh", "Fl", "Mc", "Lv", "Ts", "Og"},
}

type elementRow struct {
	symbol, name, category, phase string
	abundance                     float64 // ppm in Earth's crust
}

// Ordered by atomic number; Z is the index + 1.
var elementRows = [118]elementRow{
	{"H", "Hidrogen", catNonmetal, gas, 1400},
	{"He", "Helium", catNoble, gas, 0.008},
	{"Li", "Litium", catAlkali, solid, 20},
	{"Be", "Berilium", catAlkaline, solid, 2.8},
	{"B", "Boron", catMetalloid, solid, 10},
	{"C", "Karbon", catNonmetal, solid, 200},
	{"N", "Nitrogen", catNonmetal, gas, 19},
	{"O", "Oksigen", catNonmetal, gas, 461000},
	{"F", "Fluorin", catHalogen, gas, 585},
	{"Ne", "Neon", catNoble, gas, 0.005},
	{"Na", "Natrium", catAlkali, solid, 23600},
	{"Mg", "Magnesium", catAlkaline, solid, 23300},
	{"Al", "Aluminium", catPostTransition, solid, 82300},
	{"Si", "Silikon", catMetalloid, solid, 282000},
	{"P", "Fosfor", catNonmetal, solid, 1050},
	{"S", "Belerang", catNonmetal, solid, 350},
	{"Cl", "Klorin", catHalogen, gas, 145},
	{"Ar", "Argon", catNoble, gas, 3.5},
	{"K", "Kalium", catAlkali, solid, 20900},
	{"Ca", "Kalsium", catAlkaline, solid, 41500},
	{"Sc", "Skandium", catTransition, solid, 22},
	{"Ti", "Titanium", catTransition, solid, 5650},
	{"V", "Vanadium", catTransition, solid, 120},
	{"Cr", "Kromium", catTransition, solid, 102},
	{"Mn", "Mangan", catTransition, solid, 950},
	{"Fe", "Besi", catTransition, solid, 56300},
	{"Co", "Kobalt", catTransition, solid, 25},
	{"Ni", "Nikel", catTransition, solid, 84},
	{"Cu", "Tembaga", catTransition, solid, 60},
	{"Zn", "Seng", catTransition, solid, 70},
	{"Ga", "Galium", catPostTransition, solid, 19},
	{"Ge", "Germanium", catMetalloid, solid, 1.5},
	{"As", "Arsen", catMetalloid, solid, 1.8},
	{"Se", "Selenium", catNonmetal, solid, 0.05},
	{"Br", "Bromin", catHalogen, liquid, 2.4},
	{"Kr", "Kripton", catNoble, gas, 1e-4},
	{"Rb", "Rubidium", catAlkali, solid, 90},
	{"Sr", "Stronsium", catAlkaline, solid, 370},
	{"Y", "Itrium", catTransition, solid, 33},
	{"Zr", "Zirkonium", catTransition, solid, 165},
	{"Nb", "Niobium", catTransition, solid, 20},
	{"Mo", "Molibdenum", catTransition, solid, 1.2},
	{"Tc", "Teknesium", catTransition, solid, 0},
	{"Ru", "Rutenium", catTransition, solid, 0.001},
	{"Rh", "Rodium", catTransition, solid, 0.001},
	{"Pd", "Paladium", catTransition, solid, 0.015},
	{"Ag", "Perak", catTransition, solid, 0.075},
	{"Cd", "Kadmium", catTransition, solid, 0.15},
	{"In", "Indium", catPostTransition, solid, 0.25},
	{"Sn", "Timah", catPostTransition, solid, 2.3},
	{"Sb", "Antimon", catMetalloid, solid, 0.2},
	{"Te", "Telurium", catMetalloid, solid, 0.001},
	{"I", "Iodin", catHalogen, solid, 0.45},
	{"Xe", "Xenon", catNoble, gas, 3e-5},
	{"Cs", "Sesium", catAlkali, solid, 3},
	{"Ba", "Barium", catAlkaline, solid, 425},
	{"La", "Lantanum", catLanthanide, solid, 39},
	{"Ce", "Serium", catLanthanide, solid, 66.5},
	{"Pr", "Praseodimium", catLanthanide, solid, 9.2},
	{"Nd", "Neodimium", catLanthanide, solid, 41.5},
	{"Pm", "Prometium", catLanthanide, solid, 0},
	{"Sm", "Samarium", catLanthanide, solid, 7.05},
	{"Eu", "Europium", catLanthanide, solid, 2},
	{"Gd", "Gadolinium", catLanthanide, solid, 6.2},
	{"Tb", "Terbium", catLanthanide, solid, 1.2},
	{"Dy", "Disprosium", catLanthanide, solid, 5.2},
	{"Ho", "Holmium", catLanthanide, solid, 1.3},
	{"Er", "Erbium", catLanthanide, solid, 3.5},
	{"Tm", "Tulium", catLanthanide, solid, 0.52},
	{"Yb", "Iterbium", catLanthanide, solid, 3.2},
	{"Lu", "Lutesium", catLanthanide, solid, 0.8},
	{"Hf", "Hafnium", catTransition, solid, 3},
	{"Ta", "Tantalum", catTransition, solid, 2},
	{"W", "Wolfram", catTransition, solid, 1.25},
	{"Re", "Renium", catTransition, solid, 7e-4},
	{"Os", "Osmium", catTransition, solid, 0.0015},
	{"Ir", "Iridium", catTransition, solid, 0.001},
	{"Pt", "Platina", catTransition, solid, 0.005},
	{"Au", "Emas", catTransition, solid, 0.004},
	{"Hg", "Raksa", catTransition, liquid, 0.085},
	{"Tl", "Talium", catPostTransition, solid, 0.85},
	{"Pb", "Timbal", catPostTransition, solid, 14},
	{"Bi", "Bismut", catPostTransition, solid, 0.0085},
	{"Po", "Polonium", catPostTransition, solid, 2e-10},
	{"At", "Astatin", catHalogen, solid, 0},
	{"Rn", "Radon", catNoble, gas, 4e-13},
	{"Fr", "Fransium", catAlkali, solid, 1e-18},
	{"Ra", "Radium", catAlkaline, solid, 9e-7},
	{"Ac", "Aktinium", catActinide, solid, 5.5e-10},
	{"Th", "Torium", catActinide, solid, 9.6},
	{"Pa", "Protaktinium", catActinide, solid, 1.4e-6},
	{"U", "Uranium", catActinide, solid, 2.7},
	{"Np", "Neptunium", catActinide, solid, 0},
	{"Pu", "Plutonium", catActinide, solid, 0},
	{"Am", "Amerisium", catActinide, solid, 0},
	{"Cm", "Kurium", catActinide, solid, 0},
	{"Bk", "Berkelium", catActinide, solid, 0},
	{"Cf", "Kalifornium", catActinide, solid, 0},
	{"Es", "Einsteinium", catActinide, solid, 0},
	{"Fm", "Fermium", catActinide, solid, 0},
	{"Md", "Mendelevium", catActinide, solid, 0},
	{"No", "Nobelium", catActinide, solid, 0},
	{"Lr", "Lawrensium", catActinide, solid, 0},
	{"Rf", "Rutherfordium", catTransition, solid, 0},
	{"Db", "Dubnium", catTransition, solid, 0},
	{"Sg", "Seaborgium", catTransition, solid, 0},
	{"Bh", "Bohrium", catTransition, solid, 0},
	{"Hs", "Hasium", catTransition, solid, 0},
	{"Mt", "Meitnerium", catTransition, solid, 0},
	{"Ds", "Darmstadtium", catTransition, solid, 0},
	{"Rg", "Roentgenium", catTransition, solid, 0},
	{"Cn", "Kopernisium", catTransition, solid, 0},
	{"Nh", "Nihonium", catPostTransition, solid, 0},
	{"Fl", "Flerovium", catPostTransition, solid, 0},
	{"Mc", "Moskovium", catPostTransition, solid, 0},
	{"Lv", "Livermorium", catPostTransition, solid, 0},
	{"Ts", "Tenesin", catHalogen, solid, 0},
	{"Og", "Oganeson", catNoble, solid, 0},
}

var (
	elements    []Element
	elementBySy = map[string]int{} // symbol -> index into elements
)

func init() {
	tierOf := map[string]int{}
	for tier, syms := range elementTiers {
		for _, s := range syms {
			tierOf[s] = tier
		}
	}
	elements = make([]Element, len(elementRows))
	for i, r := range elementRows {
		z := i + 1
		period, group := position(z)
		elements[i] = Element{
			Z:         z,
			Symbol:    r.symbol,
			Name:      r.name,
			Category:  r.category,
			Period:    period,
			Group:     group,
			Phase:     r.phase,
			Abundance: r.abundance,
			// Technetium, promethium, astatine and everything past uranium only
			// exist in usable amounts when made in a reactor or accelerator.
			Natural: z <= 92 && z != 43 && z != 61 && z != 85,
			Tier:    tierOf[r.symbol],
		}
		elementBySy[r.symbol] = i
	}
}

// position returns the period and group of element z in the standard
// 18-column table; lanthanides and actinides get group 0.
func position(z int) (period, group int) {
	starts := [...]int{1, 3, 11, 19, 37, 55, 87, 119}
	for p := 0; p < 7; p++ {
		if z >= starts[p+1] {
			continue
		}
		period = p + 1
		i := z - starts[p]
		size := starts[p+1] - starts[p]
		switch {
		case period == 1:
			if z == 1 {
				return period, 1
			}
			return period, 18
		case size == 8: // periods 2–3: s-block then p-block
			if i < 2 {
				return period, i + 1
			}
			return period, i + 11
		case size == 18:
			return period, i + 1
		default: // periods 6–7 with the f-block
			switch {
			case i < 2:
				return period, i + 1
			case i < 17:
				return period, 0
			default:
				return period, i - 13
			}
		}
	}
	return 0, 0
}

// Elements returns all 118 elements ordered by atomic number.
func Elements() []Element { return slices.Clone(elements) }

func ElementBySymbol(symbol string) (Element, bool) {
	i, ok := elementBySy[symbol]
	if !ok {
		return Element{}, false
	}
	return elements[i], true
}
