package chem

import "slices"

const (
	kindFood      = "makanan"
	kindBasic     = "bahan"
	kindMineral   = "mineral"
	kindProcessed = "olahan"
	kindMetal     = "logam"
	kindTool      = "alat"
	kindWeapon    = "senjata"
)

func el(symbols ...string) []string { return symbols }

var items = []Item{
	// Food. Keeps is a half-life in years, compressed like everything else
	// against how fast bodies burn energy (a person lasts a few years without
	// food here, not a couple of months): fresh meat and fish go off first,
	// then fruit and greens, tubers last longer, dry grain and sago starch
	// many years, smoked or salted food longest.
	{ID: Food, Name: "Makanan", Kind: kindFood, Elements: el("C", "H", "O", "N"), Food: 0.4, Value: 1, Keeps: 1},
	{ID: "rumput_laut", Name: "Rumput Laut", Kind: kindFood, Elements: el("C", "H", "O", "N", "I", "Br", "K"), Food: 0.1, Value: 0.5, Keeps: 1},
	{ID: "padi", Name: "Padi (gabah)", Kind: kindFood, Elements: el("C", "H", "O", "N"), Food: 0.35, Value: 1.2, Keeps: 5},
	{ID: "talas", Name: "Talas", Kind: kindFood, Elements: el("C", "H", "O", "N"), Food: 0.4, Value: 1, Keeps: 1.5},
	{ID: "ubi", Name: "Ubi (uwi, gembili)", Kind: kindFood, Elements: el("C", "H", "O", "N"), Food: 0.4, Value: 1, Keeps: 2.5},
	{ID: "pisang", Name: "Pisang", Kind: kindFood, Elements: el("C", "H", "O", "N"), Food: 0.3, Value: 1, Keeps: 0.75},
	{ID: "kelapa", Name: "Kelapa", Kind: kindFood, Elements: el("C", "H", "O", "N"), Food: 0.45, Value: 1.2, Keeps: 3},
	{ID: "sagu", Name: "Sagu", Kind: kindFood, Elements: el("C", "H", "O"), Food: 0.5, Value: 1.2, Keeps: 5},
	{ID: "daging", Name: "Daging", Kind: kindFood, Elements: el("C", "H", "O", "N"), Food: 0.5, Value: 1.5, Keeps: 0.5},
	{ID: "ikan", Name: "Ikan", Kind: kindFood, Elements: el("C", "H", "O", "N"), Food: 0.4, Value: 1.2, Keeps: 0.5},
	{ID: "daging_asap", Name: "Daging Asap", Kind: kindFood, Elements: el("C", "H", "O", "N"), Food: 0.45, Value: 2.5, Keeps: 10},
	{ID: "ikan_asin", Name: "Ikan Asin", Kind: kindFood, Elements: el("C", "H", "O", "N", "Na", "Cl"), Food: 0.4, Value: 2.5, Keeps: 10},

	// Everyday materials.
	{ID: "kayu", Name: "Kayu", Kind: kindBasic, Formula: "(C6H10O5)n", Elements: el("C", "H", "O"), Value: 1},
	{ID: "serat", Name: "Serat Tumbuhan", Kind: kindBasic, Elements: el("C", "H", "O"), Value: 0.5},
	{ID: "batu", Name: "Batu", Kind: kindBasic, Elements: el("Si", "O", "Al"), Value: 0.5},
	{ID: "pasir", Name: "Pasir", Kind: kindBasic, Formula: "SiO2", Elements: el("Si", "O"), Value: 0.3},
	{ID: "tanah_liat", Name: "Tanah Liat", Kind: kindBasic, Formula: "Al2Si2O5(OH)4", Elements: el("Al", "Si", "O", "H"), Value: 0.5},
	{ID: "batu_kapur", Name: "Batu Kapur", Kind: kindBasic, Formula: "CaCO3", Elements: el("Ca", "C", "O"), Value: 0.8},
	{ID: "air", Name: "Air", Kind: kindBasic, Formula: "H2O", Elements: el("H", "O"), Value: 0.2},
	{ID: "air_laut", Name: "Air Laut", Kind: kindBasic, Elements: el("H", "O", "Na", "Cl", "Mg", "S", "Ca", "K", "Br", "Sr", "B", "Li", "I"), Value: 0.3},
	{ID: "udara", Name: "Udara", Kind: kindBasic, Elements: el("N", "O", "Ar", "C", "Ne", "He", "Kr", "H", "Xe"), Value: 0},
	{ID: "garam", Name: "Garam", Kind: kindProcessed, Formula: "NaCl", Elements: el("Na", "Cl"), Value: 2},

	// Minerals and ores.
	{ID: "belerang", Name: "Belerang", Kind: kindMineral, Formula: "S", Elements: el("S"), Value: 2},
	{ID: "batu_bara", Name: "Batu Bara", Kind: kindMineral, Elements: el("C", "H", "S", "N", "O"), Value: 1.5},
	{ID: "kalkopirit", Name: "Kalkopirit", Kind: kindMineral, Formula: "CuFeS2", Elements: el("Cu", "Fe", "S", "Se", "Te"), Value: 3},
	{ID: "kasiterit", Name: "Kasiterit", Kind: kindMineral, Formula: "SnO2", Elements: el("Sn", "O"), Value: 3},
	{ID: "hematit", Name: "Hematit", Kind: kindMineral, Formula: "Fe2O3", Elements: el("Fe", "O"), Value: 2},
	{ID: "galena", Name: "Galena", Kind: kindMineral, Formula: "PbS", Elements: el("Pb", "S", "Ag", "Tl"), Value: 3},
	{ID: "sfalerit", Name: "Sfalerit", Kind: kindMineral, Formula: "ZnS", Elements: el("Zn", "S", "Cd", "Ga", "Ge", "In"), Value: 3},
	{ID: "sinabar", Name: "Sinabar", Kind: kindMineral, Formula: "HgS", Elements: el("Hg", "S"), Value: 4},
	{ID: "bijih_emas", Name: "Bijih Emas", Kind: kindMineral, Formula: "Au", Elements: el("Au", "Ag", "Te"), Value: 10},
	{ID: "stibnit", Name: "Stibnit", Kind: kindMineral, Formula: "Sb2S3", Elements: el("Sb", "S"), Value: 3},
	{ID: "arsenopirit", Name: "Arsenopirit", Kind: kindMineral, Formula: "FeAsS", Elements: el("Fe", "As", "S"), Value: 2},
	{ID: "bismutinit", Name: "Bismutinit", Kind: kindMineral, Formula: "Bi2S3", Elements: el("Bi", "S"), Value: 3},
	{ID: "bauksit", Name: "Bauksit", Kind: kindMineral, Formula: "Al(OH)3", Elements: el("Al", "O", "H", "Ga"), Value: 2},
	{ID: "ilmenit", Name: "Ilmenit", Kind: kindMineral, Formula: "FeTiO3", Elements: el("Fe", "Ti", "O"), Value: 3},
	{ID: "kromit", Name: "Kromit", Kind: kindMineral, Formula: "FeCr2O4", Elements: el("Fe", "Cr", "O"), Value: 3},
	{ID: "pirolusit", Name: "Pirolusit", Kind: kindMineral, Formula: "MnO2", Elements: el("Mn", "O"), Value: 3},
	{ID: "pentlandit", Name: "Pentlandit", Kind: kindMineral, Formula: "(Fe,Ni)9S8", Elements: el("Ni", "Fe", "S", "Co", "Pt", "Pd", "Rh", "Ru", "Ir", "Os"), Value: 5},
	{ID: "molibdenit", Name: "Molibdenit", Kind: kindMineral, Formula: "MoS2", Elements: el("Mo", "S", "Re"), Value: 4},
	{ID: "wolframit", Name: "Wolframit", Kind: kindMineral, Formula: "(Fe,Mn)WO4", Elements: el("W", "Fe", "Mn", "O"), Value: 4},
	{ID: "zirkon", Name: "Zirkon", Kind: kindMineral, Formula: "ZrSiO4", Elements: el("Zr", "Hf", "Si", "O"), Value: 4},
	{ID: "beril", Name: "Beril", Kind: kindMineral, Formula: "Be3Al2Si6O18", Elements: el("Be", "Al", "Si", "O"), Value: 5},
	{ID: "spodumen", Name: "Spodumen", Kind: kindMineral, Formula: "LiAlSi2O6", Elements: el("Li", "Al", "Si", "O"), Value: 4},
	{ID: "polusit", Name: "Polusit", Kind: kindMineral, Formula: "(Cs,Na)2Al2Si4O12·2H2O", Elements: el("Cs", "Rb", "Na", "Al", "Si", "O", "H"), Value: 5},
	{ID: "monasit", Name: "Monasit", Kind: kindMineral, Formula: "(Ce,La,Nd,Th)PO4", Elements: el("Ce", "La", "Nd", "Pr", "Sm", "Gd", "Eu", "Th", "Y", "P", "O"), Value: 6},
	{ID: "xenotim", Name: "Xenotim", Kind: kindMineral, Formula: "YPO4", Elements: el("Y", "Dy", "Er", "Yb", "Ho", "Tm", "Lu", "Tb", "P", "O"), Value: 6},
	{ID: "torvetit", Name: "Torvetit", Kind: kindMineral, Formula: "(Sc,Y)2Si2O7", Elements: el("Sc", "Y", "Si", "O"), Value: 8},
	{ID: "uraninit", Name: "Uraninit", Kind: kindMineral, Formula: "UO2", Elements: el("U", "Th", "Ra", "Po", "Pa", "Ac", "Rn", "He", "Pb", "Fr", "O"), Value: 8},
	{ID: "fluorit", Name: "Fluorit", Kind: kindMineral, Formula: "CaF2", Elements: el("Ca", "F"), Value: 2},
	{ID: "apatit", Name: "Apatit", Kind: kindMineral, Formula: "Ca5(PO4)3(F,Cl,OH)", Elements: el("Ca", "P", "O", "F", "Cl", "H"), Value: 2},
	{ID: "gipsum", Name: "Gipsum", Kind: kindMineral, Formula: "CaSO4·2H2O", Elements: el("Ca", "S", "O", "H"), Value: 1},
	{ID: "barit", Name: "Barit", Kind: kindMineral, Formula: "BaSO4", Elements: el("Ba", "S", "O"), Value: 2},
	{ID: "selestin", Name: "Selestin", Kind: kindMineral, Formula: "SrSO4", Elements: el("Sr", "S", "O"), Value: 2},
	{ID: "boraks", Name: "Boraks", Kind: kindMineral, Formula: "Na2B4O7·10H2O", Elements: el("Na", "B", "O", "H"), Value: 2},
	{ID: "kolumbit", Name: "Kolumbit", Kind: kindMineral, Formula: "(Fe,Mn)(Nb,Ta)2O6", Elements: el("Nb", "Ta", "Fe", "Mn", "O"), Value: 6},
	{ID: "kobaltit", Name: "Kobaltit", Kind: kindMineral, Formula: "CoAsS", Elements: el("Co", "As", "S"), Value: 4},
	{ID: "silvit", Name: "Silvit", Kind: kindMineral, Formula: "KCl", Elements: el("K", "Cl"), Value: 2},
	{ID: "dolomit", Name: "Dolomit", Kind: kindMineral, Formula: "CaMg(CO3)2", Elements: el("Ca", "Mg", "C", "O"), Value: 1},
	{ID: "vanadinit", Name: "Vanadinit", Kind: kindMineral, Formula: "Pb5(VO4)3Cl", Elements: el("V", "Pb", "Cl", "O"), Value: 4},
	{ID: "limonit", Name: "Limonit (Laterit Besi)", Kind: kindMineral, Formula: "FeO(OH)·nH2O", Elements: el("Fe", "O", "H"), Value: 1.5},
	{ID: "pasir_besi", Name: "Pasir Besi", Kind: kindMineral, Formula: "Fe3O4·FeTiO3", Elements: el("Fe", "Ti", "V", "O"), Value: 1.5},
	{ID: "malakit", Name: "Malakit", Kind: kindMineral, Formula: "Cu2CO3(OH)2", Elements: el("Cu", "C", "O", "H"), Value: 3},
	{ID: "tembaga_alam", Name: "Tembaga Alam", Kind: kindMineral, Formula: "Cu", Elements: el("Cu", "Ag"), Value: 4},
	{ID: "laterit_nikel", Name: "Laterit Nikel", Kind: kindMineral, Formula: "(Ni,Mg)3Si2O5(OH)4 + FeO(OH)", Elements: el("Ni", "Co", "Fe", "Mg", "Si", "O", "H", "Sc", "Cr"), Value: 3},
	{ID: "halit", Name: "Halit (Garam Batu)", Kind: kindMineral, Formula: "NaCl", Elements: el("Na", "Cl"), Value: 1.5},
	{ID: "scheelit", Name: "Scheelit", Kind: kindMineral, Formula: "CaWO4", Elements: el("Ca", "W", "O"), Value: 4},
	{ID: "karnotit", Name: "Karnotit", Kind: kindMineral, Formula: "K2(UO2)2(VO4)2·3H2O", Elements: el("K", "U", "V", "O", "H"), Value: 6},
	{ID: "pirokhlor", Name: "Pirokhlor", Kind: kindMineral, Formula: "(Na,Ca)2Nb2O6(OH,F)", Elements: el("Na", "Ca", "Nb", "Ta", "O", "H", "F"), Value: 6},
	{ID: "bastnasit", Name: "Bastnasit", Kind: kindMineral, Formula: "(Ce,La)CO3F", Elements: el("Ce", "La", "Nd", "Pr", "C", "O", "F"), Value: 6},

	// Processed materials.
	{ID: "arang", Name: "Arang", Kind: kindProcessed, Formula: "C", Elements: el("C"), Value: 1},
	{ID: "tali", Name: "Tali", Kind: kindProcessed, Elements: el("C", "H", "O"), Value: 1},
	{ID: "bata", Name: "Bata", Kind: kindProcessed, Elements: el("Al", "Si", "O"), Value: 1.5},
	{ID: "kaca", Name: "Kaca", Kind: kindProcessed, Formula: "SiO2·Na2O·CaO", Elements: el("Si", "O", "Na", "Ca"), Value: 3},
	{ID: "semen", Name: "Semen", Kind: kindProcessed, Elements: el("Ca", "Si", "Al", "O"), Value: 2},
	{ID: "beton", Name: "Beton", Kind: kindProcessed, Elements: el("Ca", "Si", "Al", "O"), Value: 3},

	// Metals and alloys.
	{ID: "tembaga", Name: "Tembaga", Kind: kindMetal, Formula: "Cu", Elements: el("Cu"), Value: 5},
	{ID: "timah", Name: "Timah", Kind: kindMetal, Formula: "Sn", Elements: el("Sn"), Value: 5},
	{ID: "perunggu", Name: "Perunggu", Kind: kindMetal, Formula: "Cu+Sn", Elements: el("Cu", "Sn"), Value: 7},
	{ID: "besi", Name: "Besi", Kind: kindMetal, Formula: "Fe", Elements: el("Fe"), Value: 5},
	{ID: "baja", Name: "Baja", Kind: kindMetal, Formula: "Fe+C", Elements: el("Fe", "C"), Value: 8},
	{ID: "timbal", Name: "Timbal", Kind: kindMetal, Formula: "Pb", Elements: el("Pb"), Value: 4},
	{ID: "seng", Name: "Seng", Kind: kindMetal, Formula: "Zn", Elements: el("Zn"), Value: 4},
	{ID: "kuningan", Name: "Kuningan", Kind: kindMetal, Formula: "Cu+Zn", Elements: el("Cu", "Zn"), Value: 6},
	{ID: "emas", Name: "Emas", Kind: kindMetal, Formula: "Au", Elements: el("Au"), Value: 20},
	{ID: "perak", Name: "Perak", Kind: kindMetal, Formula: "Ag", Elements: el("Ag"), Value: 12},
	{ID: "raksa", Name: "Raksa", Kind: kindMetal, Formula: "Hg", Elements: el("Hg"), Value: 6},
	{ID: "aluminium", Name: "Aluminium", Kind: kindMetal, Formula: "Al", Elements: el("Al"), Value: 8},
	{ID: "titanium", Name: "Titanium", Kind: kindMetal, Formula: "Ti", Elements: el("Ti"), Value: 12},
	{ID: "niobium", Name: "Niobium", Kind: kindMetal, Formula: "Nb", Elements: el("Nb"), Value: 12},
	{ID: SynthesisFuel, Name: "Uranium", Kind: kindMetal, Formula: "U", Elements: el("U"), Value: 15},
	{ID: "perhiasan", Name: "Perhiasan", Kind: kindProcessed, Formula: "Au+Ag", Elements: el("Au", "Ag"), Value: 30},

	// Tools speed up gathering and are needed for ores.
	{ID: "beliung_batu", Name: "Beliung Batu", Kind: kindTool, Elements: el("Si", "O", "C"), Value: 2, Gather: 1.5},
	{ID: "beliung_perunggu", Name: "Beliung Perunggu", Kind: kindTool, Elements: el("Cu", "Sn", "C"), Value: 8, Gather: 2},
	{ID: "beliung_besi", Name: "Beliung Besi", Kind: kindTool, Elements: el("Fe", "C"), Value: 10, Gather: 2.5},
	{ID: "beliung_baja", Name: "Beliung Baja", Kind: kindTool, Elements: el("Fe", "C"), Value: 13, Gather: 3},

	// Weapons add to attack damage.
	{ID: "tombak_batu", Name: "Tombak Batu", Kind: kindWeapon, Elements: el("Si", "O", "C"), Value: 2, Damage: 0.5},
	{ID: "pedang_perunggu", Name: "Pedang Perunggu", Kind: kindWeapon, Elements: el("Cu", "Sn", "C"), Value: 9, Damage: 1},
	{ID: "pedang_besi", Name: "Pedang Besi", Kind: kindWeapon, Elements: el("Fe", "C"), Value: 11, Damage: 1.5},
	{ID: "pedang_baja", Name: "Pedang Baja", Kind: kindWeapon, Elements: el("Fe", "C"), Value: 14, Damage: 2},
}

var itemIndex = map[ItemID]int{}

func init() {
	for i, it := range items {
		itemIndex[it.ID] = i
	}
}

// Items returns every item in a stable order.
func Items() []Item { return slices.Clone(items) }

func ItemByID(id ItemID) (Item, bool) {
	i, ok := itemIndex[id]
	if !ok {
		return Item{}, false
	}
	return items[i], true
}
