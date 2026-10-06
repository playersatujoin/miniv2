package chem

import (
	"maps"
	"slices"
)

type in = map[ItemID]int

var recipes = []Recipe{
	// Stone age: no station.
	{ID: "arang", Name: "Membakar arang", Inputs: in{"kayu": 2}, Outputs: in{"arang": 1}, Teaches: "api", Discovers: el("C"), Seconds: 3},
	{ID: "tali", Name: "Memilin tali", Inputs: in{"serat": 2}, Outputs: in{"tali": 1}, Seconds: 2},
	{ID: "beliung_batu", Name: "Membuat beliung batu", Inputs: in{"batu": 2, "kayu": 1, "tali": 1}, Outputs: in{"beliung_batu": 1}, Teaches: "alat_batu", Seconds: 4},
	{ID: "tombak_batu", Name: "Membuat tombak batu", Inputs: in{"kayu": 2, "batu": 1, "tali": 1}, Outputs: in{"tombak_batu": 1}, Teaches: "alat_batu", Seconds: 4},
	{ID: "bata", Name: "Membakar bata", Inputs: in{"tanah_liat": 2, "arang": 1}, Outputs: in{"bata": 2}, Tech: "api", Teaches: "tembikar", Seconds: 4},
	{ID: "garam", Name: "Menguapkan air laut", Inputs: in{"air_laut": 3, "arang": 1}, Outputs: in{"garam": 1}, Tech: "api", Seconds: 4},
	{ID: "masak_rumput_laut", Name: "Memasak rumput laut", Inputs: in{"rumput_laut": 3, "arang": 1}, Outputs: in{Food: 2}, Tech: "api", Seconds: 3},
	// Smoke and salt keep meat and fish for years instead of days.
	{ID: "asap_daging", Name: "Mengasap daging", Inputs: in{"daging": 3, "kayu": 1}, Outputs: in{"daging_asap": 3}, Tech: "api", Teaches: "pengawetan", Seconds: 3},
	{ID: "asin_ikan", Name: "Mengasinkan ikan", Inputs: in{"ikan": 3, "garam": 1}, Outputs: in{"ikan_asin": 3}, Tech: "api", Teaches: "pengawetan", Seconds: 3},
	{ID: "garam_batu", Name: "Menggerus garam batu", Inputs: in{"halit": 2}, Outputs: in{"garam": 2}, Seconds: 2},
	// Native copper was hammered and annealed long before ores were smelted.
	{ID: "tempa_tembaga_alam", Name: "Menempa tembaga alam", Inputs: in{"tembaga_alam": 2, "arang": 1}, Outputs: in{"tembaga": 1}, Tech: "api", Discovers: el("Cu"), Seconds: 4},

	// Furnace: smelting and alloys.
	{ID: "lebur_tembaga", Name: "Melebur tembaga", Inputs: in{"kalkopirit": 2, "arang": 1}, Outputs: in{"tembaga": 1}, Station: "tungku", Tech: "peleburan", Discovers: el("Cu"), Seconds: 5},
	{ID: "lebur_timah", Name: "Melebur timah", Inputs: in{"kasiterit": 2, "arang": 1}, Outputs: in{"timah": 1}, Station: "tungku", Tech: "peleburan", Discovers: el("Sn"), Seconds: 5},
	{ID: "perunggu", Name: "Membuat perunggu", Inputs: in{"tembaga": 2, "timah": 1}, Outputs: in{"perunggu": 3}, Station: "tungku", Tech: "peleburan", Teaches: "perunggu", Seconds: 5},
	{ID: "lebur_malakit", Name: "Melebur malakit", Inputs: in{"malakit": 2, "arang": 1}, Outputs: in{"tembaga": 1}, Station: "tungku", Tech: "peleburan", Discovers: el("Cu"), Seconds: 5},
	{ID: "lebur_besi", Name: "Melebur besi", Inputs: in{"hematit": 2, "arang": 2}, Outputs: in{"besi": 1}, Station: "tungku", Tech: "peleburan", Teaches: "besi", Discovers: el("Fe"), Seconds: 6},
	{ID: "lebur_limonit", Name: "Melebur limonit", Inputs: in{"limonit": 3, "arang": 2}, Outputs: in{"besi": 1}, Station: "tungku", Tech: "peleburan", Teaches: "besi", Discovers: el("Fe"), Seconds: 6},
	{ID: "lebur_pasir_besi", Name: "Melebur pasir besi", Inputs: in{"pasir_besi": 3, "arang": 2}, Outputs: in{"besi": 1}, Station: "tungku", Tech: "peleburan", Teaches: "besi", Discovers: el("Fe"), Seconds: 6},
	{ID: "baja", Name: "Membuat baja", Inputs: in{"besi": 2, "arang": 1}, Outputs: in{"baja": 1}, Station: "tungku", Tech: "besi", Teaches: "baja", Seconds: 6},
	{ID: "lebur_timbal", Name: "Melebur timbal", Inputs: in{"galena": 2, "arang": 1}, Outputs: in{"timbal": 1}, Station: "tungku", Tech: "peleburan", Discovers: el("Pb"), Seconds: 5},
	{ID: "pemurnian_perak", Name: "Memurnikan perak", Inputs: in{"galena": 3, "arang": 1}, Outputs: in{"perak": 1, "timbal": 1}, Station: "tungku", Tech: "peleburan", Discovers: el("Ag"), Seconds: 6},
	{ID: "lebur_seng", Name: "Melebur seng", Inputs: in{"sfalerit": 2, "arang": 2}, Outputs: in{"seng": 1}, Station: "tungku", Tech: "peleburan", Discovers: el("Zn"), Seconds: 5},
	{ID: "kuningan", Name: "Membuat kuningan", Inputs: in{"tembaga": 2, "seng": 1}, Outputs: in{"kuningan": 3}, Station: "tungku", Tech: "peleburan", Seconds: 5},
	{ID: "raksa", Name: "Memanggang sinabar", Inputs: in{"sinabar": 2, "arang": 1}, Outputs: in{"raksa": 1}, Station: "tungku", Tech: "peleburan", Discovers: el("Hg"), Seconds: 5},
	{ID: "lebur_emas", Name: "Melebur emas", Inputs: in{"bijih_emas": 2, "arang": 1}, Outputs: in{"emas": 1}, Station: "tungku", Tech: "peleburan", Discovers: el("Au"), Seconds: 5},
	{ID: "perhiasan", Name: "Membuat perhiasan", Inputs: in{"emas": 1, "perak": 1}, Outputs: in{"perhiasan": 1}, Station: "tungku", Tech: "peleburan", Seconds: 6},
	{ID: "kokas", Name: "Membakar batu bara", Inputs: in{"batu_bara": 2}, Outputs: in{"arang": 2}, Station: "tungku", Tech: "peleburan", Seconds: 3},
	{ID: "kaca", Name: "Meniup kaca", Inputs: in{"pasir": 2, "batu_kapur": 1, "arang": 1}, Outputs: in{"kaca": 1}, Station: "tungku", Tech: "peleburan", Teaches: "kaca", Seconds: 5},
	{ID: "semen", Name: "Membakar semen", Inputs: in{"batu_kapur": 2, "tanah_liat": 1, "arang": 1}, Outputs: in{"semen": 2}, Station: "tungku", Tech: "peleburan", Teaches: "semen", Seconds: 5},
	{ID: "beton", Name: "Mencampur beton", Inputs: in{"semen": 1, "pasir": 1, "batu": 1}, Outputs: in{"beton": 2}, Tech: "semen", Seconds: 4},

	// Metal tools and weapons.
	{ID: "beliung_perunggu", Name: "Menempa beliung perunggu", Inputs: in{"perunggu": 2, "kayu": 1}, Outputs: in{"beliung_perunggu": 1}, Station: "tungku", Tech: "perunggu", Seconds: 5},
	{ID: "pedang_perunggu", Name: "Menempa pedang perunggu", Inputs: in{"perunggu": 3, "kayu": 1}, Outputs: in{"pedang_perunggu": 1}, Station: "tungku", Tech: "perunggu", Seconds: 6},
	{ID: "beliung_besi", Name: "Menempa beliung besi", Inputs: in{"besi": 2, "kayu": 1}, Outputs: in{"beliung_besi": 1}, Station: "tungku", Tech: "besi", Seconds: 5},
	{ID: "pedang_besi", Name: "Menempa pedang besi", Inputs: in{"besi": 3, "kayu": 1}, Outputs: in{"pedang_besi": 1}, Station: "tungku", Tech: "besi", Seconds: 6},
	{ID: "beliung_baja", Name: "Menempa beliung baja", Inputs: in{"baja": 2, "kayu": 1}, Outputs: in{"beliung_baja": 1}, Station: "tungku", Tech: "baja", Seconds: 6},
	{ID: "pedang_baja", Name: "Menempa pedang baja", Inputs: in{"baja": 3, "kayu": 1}, Outputs: in{"pedang_baja": 1}, Station: "tungku", Tech: "baja", Seconds: 7},

	// Electrochemistry and beyond.
	{ID: "aluminium", Name: "Elektrolisis bauksit", Inputs: in{"bauksit": 3, "arang": 1}, Outputs: in{"aluminium": 1}, Station: "pembangkit_listrik", Tech: "listrik", Discovers: el("Al"), Seconds: 6},
	{ID: "titanium", Name: "Memurnikan titanium", Inputs: in{"ilmenit": 3, "arang": 2}, Outputs: in{"titanium": 1}, Station: "pembangkit_listrik", Tech: "listrik", Discovers: el("Ti"), Seconds: 7},
	{ID: "niobium", Name: "Memurnikan niobium", Inputs: in{"kolumbit": 3, "arang": 1}, Outputs: in{"niobium": 1}, Station: "pembangkit_listrik", Tech: "listrik", Discovers: el("Nb"), Seconds: 7},
	{ID: "niobium_pirokhlor", Name: "Memurnikan niobium dari pirokhlor", Inputs: in{"pirokhlor": 3, "arang": 1}, Outputs: in{"niobium": 1}, Station: "pembangkit_listrik", Tech: "listrik", Discovers: el("Nb"), Seconds: 7},
	{ID: "uranium", Name: "Memurnikan uranium", Inputs: in{"uraninit": 3, "arang": 1}, Outputs: in{SynthesisFuel: 1}, Station: "lab_radiasi", Tech: "radiokimia", Discovers: el("U"), Seconds: 8},
}

func Recipes() []Recipe {
	out := slices.Clone(recipes)
	for i := range out {
		out[i].Inputs = maps.Clone(out[i].Inputs)
		out[i].Outputs = maps.Clone(out[i].Outputs)
		out[i].Discovers = slices.Clone(out[i].Discovers)
	}
	return out
}

// Discoverable lists the elements contained in item that are not yet known and
// can be isolated at the given tier, ordered by atomic number.
func Discoverable(item ItemID, tier int, known func(symbol string) bool) []string {
	it, ok := ItemByID(item)
	if !ok {
		return nil
	}
	var out []string
	for _, sym := range it.Elements {
		e, ok := ElementBySymbol(sym)
		if ok && e.Natural && e.Tier <= tier && !known(sym) {
			out = append(out, sym)
		}
	}
	slices.SortFunc(out, byZ)
	return out
}

// Synthesizable lists synthetic (non-mineral) elements that are not yet known
// and can be made at the given tier, ordered by atomic number.
func Synthesizable(tier int, known func(symbol string) bool) []string {
	var out []string
	for _, e := range elements {
		if !e.Natural && e.Tier <= tier && !known(e.Symbol) {
			out = append(out, e.Symbol)
		}
	}
	return out
}

func byZ(a, b string) int {
	return elementBySy[a] - elementBySy[b]
}
