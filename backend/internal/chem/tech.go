package chem

import (
	"maps"
	"slices"
)

var techs = []Tech{
	{ID: "api", Name: "Api", Description: "Membakar kayu menjadi arang; dasar memasak dan membakar tanah liat.", Tier: 0},
	{ID: "alat_batu", Name: "Alat Batu", Description: "Beliung dan tombak dari batu, kayu dan tali. Membuka penambangan bijih.", Tier: 0},
	{ID: "tembikar", Name: "Tembikar", Description: "Membakar tanah liat menjadi bata.", Tier: 0, Requires: []string{"api"}},
	{ID: "pertanian", Name: "Pertanian", Description: "Menanam dan memanen padi, talas, ubi, pisang, kelapa dan sagu. Ditemukan saat panen pertama dari tanaman yang sengaja ditanam.", Tier: 0},
	{ID: "peternakan", Name: "Peternakan", Description: "Menjinakkan dan memelihara ayam, babi dan kerbau. Ditemukan saat hewan liar pertama menjadi jinak.", Tier: 0},
	{ID: "pengawetan", Name: "Pengawetan Pangan", Description: "Mengasap daging dan mengasinkan ikan agar tahan bertahun-tahun.", Tier: 0, Requires: []string{"api"}},
	{ID: "tulisan", Name: "Tulisan", Description: "Menulis di lempeng tanah liat agar pengetahuan tidak hilang bersama pemiliknya.", Tier: 1, Requires: []string{"tembikar"}},
	{ID: "peleburan", Name: "Peleburan", Description: "Tungku yang cukup panas untuk melebur bijih menjadi logam.", Tier: 1, Requires: []string{"tembikar"}},
	{ID: "perunggu", Name: "Perunggu", Description: "Paduan tembaga dan timah yang lebih keras dari keduanya.", Tier: 1, Requires: []string{"peleburan"}},
	{ID: "besi", Name: "Pengolahan Besi", Description: "Melebur hematit dengan arang menjadi besi.", Tier: 1, Requires: []string{"peleburan"}},
	{ID: "baja", Name: "Baja", Description: "Besi yang dicampur karbon menjadi lebih kuat.", Tier: 1, Requires: []string{"besi"}},
	{ID: "kaca", Name: "Kaca", Description: "Melelehkan pasir dengan batu kapur menjadi kaca bening.", Tier: 1, Requires: []string{"peleburan"}},
	{ID: "semen", Name: "Semen dan Beton", Description: "Membakar batu kapur dan tanah liat menjadi semen, lalu beton.", Tier: 1, Requires: []string{"peleburan"}},
	{ID: "kimia", Name: "Kimia", Description: "Laboratorium untuk memurnikan dan meneliti mineral.", Tier: 2, Requires: []string{"kaca", "besi"}},
	{ID: "listrik", Name: "Listrik", Description: "Arus listrik untuk elektrolisis garam dan bijih.", Tier: 3, Requires: []string{"kimia"}},
	{ID: "spektroskopi", Name: "Spektroskopi", Description: "Membaca garis cahaya unsur dan mencairkan udara.", Tier: 4, Requires: []string{"listrik"}},
	{ID: "radiokimia", Name: "Radiokimia", Description: "Memisahkan unsur radioaktif dari bijih uranium.", Tier: 5, Requires: []string{"spektroskopi", "semen"}},
	{ID: "fisika_nuklir", Name: "Fisika Nuklir", Description: "Reaktor yang menyinari uranium untuk membuat unsur baru.", Tier: 6, Requires: []string{"radiokimia", "baja"}},
	{ID: "fisika_partikel", Name: "Fisika Partikel", Description: "Akselerator yang menumbukkan inti atom menjadi unsur superberat.", Tier: 7, Requires: []string{"fisika_nuklir"}},
}

var structures = []StructureKind{
	{ID: "gubuk", Name: "Gubuk", Cost: map[ItemID]int{"kayu": 6, "serat": 4}, House: true, Level: 1, Storage: 40},
	{ID: "rumah_kayu", Name: "Rumah Kayu", Cost: map[ItemID]int{"kayu": 12, "batu": 6, "tali": 2}, Tech: "alat_batu", House: true, Level: 2, Upgrades: "gubuk", Storage: 100},
	{ID: "rumah_bata", Name: "Rumah Bata", Cost: map[ItemID]int{"bata": 16, "kayu": 6, "kaca": 2}, Tech: "kaca", House: true, Level: 3, Upgrades: "rumah_kayu", Storage: 200},
	// Farming: a cleared, weeded field, a ditch from the river, a granary on
	// stilts and a pen for livestock.
	{ID: "ladang", Name: "Ladang", Cost: map[ItemID]int{"serat": 4, "kayu": 2}, Tech: "pertanian", Farm: true},
	{ID: "saluran_irigasi", Name: "Saluran Irigasi", Cost: map[ItemID]int{"batu": 4, "kayu": 2, "tali": 1}, Tech: "pertanian", Irrigation: true},
	{ID: "lumbung", Name: "Lumbung", Cost: map[ItemID]int{"kayu": 10, "serat": 4, "tali": 2}, Tech: "pertanian", Granary: true, Storage: 120},
	{ID: "kandang", Name: "Kandang", Cost: map[ItemID]int{"kayu": 8, "tali": 2}, Tech: "peternakan", Pen: true},
	// A noose of plant cord on a bent stick, set on a game trail: foragers
	// across Southeast Asia snare pigs, deer and junglefowl this way, and
	// snaring needs no special knowledge.
	{ID: "jerat", Name: "Jerat", Cost: map[ItemID]int{"serat": 2, "kayu": 1}, Snare: true, Storage: 6},
	// Lined wells came with settled farming villages in the Neolithic;
	// foragers moved to the water instead.
	{ID: "sumur", Name: "Sumur", Cost: map[ItemID]int{"batu": 8, "tali": 2}, Tech: "pertanian", Well: true},
	// Clay tablets: writing began with fired clay, long before paper.
	{ID: "perpustakaan", Name: "Perpustakaan", Cost: map[ItemID]int{"bata": 8, "kayu": 4, "tanah_liat": 6}, Tech: "tembikar", Teaches: "tulisan", Library: true},
	{ID: "tungku", Name: "Tungku", Cost: map[ItemID]int{"batu": 8, "tanah_liat": 4, "bata": 2}, Tech: "tembikar", Teaches: "peleburan", Tier: 1},
	{ID: "laboratorium", Name: "Laboratorium", Cost: map[ItemID]int{"bata": 10, "kaca": 4, "besi": 4}, Tech: "kaca", Teaches: "kimia", Tier: 2},
	{ID: "pembangkit_listrik", Name: "Pembangkit Listrik", Cost: map[ItemID]int{"tembaga": 10, "seng": 4, "besi": 6, "kaca": 2}, Tech: "kimia", Teaches: "listrik", Tier: 3},
	{ID: "lab_spektroskopi", Name: "Lab Spektroskopi", Cost: map[ItemID]int{"kaca": 8, "kuningan": 4, "tembaga": 4, "bata": 6}, Tech: "listrik", Teaches: "spektroskopi", Tier: 4},
	{ID: "lab_radiasi", Name: "Lab Radiasi", Cost: map[ItemID]int{"timbal": 12, "baja": 8, "kaca": 4, "beton": 6}, Tech: "spektroskopi", Teaches: "radiokimia", Tier: 5},
	{ID: "reaktor_nuklir", Name: "Reaktor Nuklir", Cost: map[ItemID]int{"beton": 20, "baja": 20, "timbal": 10, SynthesisFuel: 4}, Tech: "radiokimia", Teaches: "fisika_nuklir", Tier: 6},
	{ID: "akselerator", Name: "Akselerator Partikel", Cost: map[ItemID]int{"baja": 30, "tembaga": 30, "beton": 20, "niobium": 4}, Tech: "fisika_nuklir", Teaches: "fisika_partikel", Tier: 7},
}

var structureIndex = map[string]int{}

func init() {
	for i, s := range structures {
		structureIndex[s.ID] = i
	}
}

func Techs() []Tech {
	out := slices.Clone(techs)
	for i := range out {
		out[i].Requires = slices.Clone(out[i].Requires)
	}
	return out
}

func Structures() []StructureKind {
	out := slices.Clone(structures)
	for i := range out {
		out[i].Cost = maps.Clone(out[i].Cost)
	}
	return out
}

func StructureByID(id string) (StructureKind, bool) {
	i, ok := structureIndex[id]
	if !ok {
		return StructureKind{}, false
	}
	s := structures[i]
	s.Cost = maps.Clone(s.Cost)
	return s, true
}
