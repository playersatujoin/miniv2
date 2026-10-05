package chem

import (
	"slices"

	"miniv2/backend/internal/world"
)

// DepositModel is a real-world type of ore or resource deposit.
type DepositModel struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

type weighted struct {
	item   ItemID
	weight float64
}

// modelSpec says what a deposit model yields. Minerals apply to every tile of
// its zone; Surface, when set, replaces them on walkable ground (the weathered
// top of a deposit). Barren tiles yield only the host rock.
type modelSpec struct {
	DepositModel
	minerals []weighted
	surface  []weighted
	barren   float64
	host     ItemID
	rocks    []world.Rock // host rock units, for relocating deposits on edited maps
}

var modelSpecs = []modelSpec{
	{
		DepositModel: DepositModel{"porfiri", "Porfiri Cu-Au-Mo",
			"Tembaga, emas dan molibdenum tersebar halus di dalam dan di sekitar stok intrusi di bawah gunung api. Di permukaan, bagian yang lapuk teroksidasi menjadi malakit hijau — sumber tembaga pertama manusia.",
			"Grasberg (Papua), Batu Hijau (Sumbawa), Tujuh Bukit (Jawa Timur)"},
		minerals: []weighted{{"kalkopirit", 60}, {"molibdenit", 25}, {"bijih_emas", 8}},
		surface:  []weighted{{"malakit", 55}, {"kalkopirit", 25}, {"molibdenit", 10}, {"bijih_emas", 5}},
		barren:   0.25, host: "batu",
		rocks: []world.Rock{world.RockGranite, world.RockVolcanic},
	},
	{
		DepositModel: DepositModel{"epitermal", "Epitermal Au-Ag",
			"Urat kuarsa berisi emas dan perak dari fluida panas dangkal di sekitar gunung api; raksa (sinabar), antimon (stibnit) dan arsen (arsenopirit) menyertainya.",
			"Pongkor (Jawa Barat), Martabe (Sumatra Utara), Gosowong (Halmahera)"},
		minerals: []weighted{{"bijih_emas", 20}, {"arsenopirit", 30}, {"stibnit", 30}, {"sinabar", 20}},
		barren:   0.15, host: "batu",
		rocks: []world.Rock{world.RockVolcanic},
	},
	{
		DepositModel: DepositModel{"solfatara", "Belerang kawah (solfatara)",
			"Belerang murni mengendap dari gas fumarol di dinding kawah gunung api aktif dan terus terbentuk kembali; para penambang memecah dan memikulnya keluar kawah.",
			"Kawah Ijen (Jawa Timur)"},
		minerals: []weighted{{"belerang", 1}},
	},
	{
		DepositModel: DepositModel{"skarn", "Skarn Fe-Cu-Zn-W",
			"Terbentuk di kontak intrusi dengan batu gamping: magnetit-hematit, tembaga, seng-timbal dan wolfram (scheelit), sering bersama fluorit.",
			"Ertsberg dan Big Gossan (Papua)"},
		minerals: []weighted{{"hematit", 40}, {"kalkopirit", 20}, {"sfalerit", 12}, {"galena", 10}, {"scheelit", 10}, {"fluorit", 8}},
		barren:   0.2, host: "batu",
		rocks: []world.Rock{world.RockLimestone, world.RockGranite},
	},
	{
		DepositModel: DepositModel{"granit_timah", "Granit timah (urat Sn-W)",
			"Granit tipe-S yang membawa urat kuarsa berisi kasiterit dan wolframit, dengan bismut, molibdenum dan fluorit.",
			"Sabuk timah Asia Tenggara: Bangka-Belitung, Semenanjung Malaysia"},
		minerals: []weighted{{"kasiterit", 50}, {"wolframit", 20}, {"fluorit", 15}, {"bismutinit", 8}, {"molibdenit", 7}},
		barren:   0.2, host: "batu",
		rocks: []world.Rock{world.RockGranite},
	},
	{
		DepositModel: DepositModel{"pegmatit", "Pegmatit logam langka",
			"Retas pegmatit berbutir kasar dari sisa magma granit: jenis LCT (litium, sesium, tantalum) dan NYF (niobium, itrium, unsur tanah jarang, uranium).",
			"Greenbushes (Australia), Tanco (Kanada)"},
		minerals: []weighted{{"spodumen", 14}, {"beril", 12}, {"kolumbit", 12}, {"kasiterit", 10}, {"polusit", 8}, {"uraninit", 8}, {"monasit", 8}, {"xenotim", 8}, {"zirkon", 6}, {"torvetit", 5}},
		barren:   0.3, host: "batu",
		rocks: []world.Rock{world.RockPegmatite},
	},
	{
		DepositModel: DepositModel{"ofiolit", "Ofiolit: kromit & sulfida Ni-Co",
			"Kantong kromit di batuan mantel yang terangkat, sedikit sulfida nikel-kobalt-platinoid, ilmenit dari gabro, dan kobaltit di zona serpentinit hidrotermal.",
			"Kromit: Kempirsai (Kazakhstan); kobalt-arsen: Bou Azzer (Maroko)"},
		minerals: []weighted{{"kromit", 40}, {"pentlandit", 25}, {"kobaltit", 15}, {"ilmenit", 10}},
		barren:   0.4, host: "batu",
		rocks: []world.Rock{world.RockUltramafic},
	},
	{
		DepositModel: DepositModel{"laterit_nikel", "Laterit nikel",
			"Tanah merah hasil pelapukan batuan ultrabasa di iklim tropis basah: lapisan limonit (besi) di atas saprolit yang kaya nikel dan kobalt, juga skandium.",
			"Sorowako (Sulawesi Selatan), Pomalaa (Sulawesi Tenggara), Weda Bay (Halmahera)"},
		minerals: []weighted{{"laterit_nikel", 65}, {"limonit", 35}},
		barren:   0.1, host: "tanah_liat",
		rocks: []world.Rock{world.RockUltramafic},
	},
	{
		DepositModel: DepositModel{"laterit_besi", "Laterit besi (kerak limonit)",
			"Kerak besi limonit-goetit yang terbentuk ketika batuan lapuk di iklim tropis dan unsur lain tercuci habis. Sangat umum di dataran rendah dan mudah digali; dilebur sejak zaman kuno.",
			"Peleburan besi laterit kuno di Afrika Barat dan India"},
		minerals: []weighted{{"limonit", 1}},
		rocks:    []world.Rock{world.RockVolcanic, world.RockMetamorphic, world.RockSedimentary},
	},
	{
		DepositModel: DepositModel{"bauksit", "Bauksit laterit",
			"Hasil pelapukan intensif granit dan batuan vulkanik di dataran tropis yang berdrainase baik; aluminium hidroksida tertinggal dan terkayakan.",
			"Kalimantan Barat, Pulau Bintan (Kepulauan Riau)"},
		minerals: []weighted{{"bauksit", 1}},
		rocks:    []world.Rock{world.RockGranite, world.RockVolcanic},
	},
	{
		DepositModel: DepositModel{"tembaga_alam", "Tembaga alam dalam basal",
			"Tembaga murni mengisi rongga dan retakan lava basal. Langka, tetapi bisa ditempa tanpa dilebur.",
			"Semenanjung Keweenaw (Amerika Serikat)"},
		minerals: []weighted{{"tembaga_alam", 1}},
		rocks:    []world.Rock{world.RockVolcanic},
	},
	{
		DepositModel: DepositModel{"batubara", "Batu bara",
			"Lapisan sisa tumbuhan rawa purba yang terkubur di cekungan sedimen; tersingkap di tepi cekungan dan di tebing sungai.",
			"Cekungan Kutai (Kalimantan Timur), Cekungan Sumatra Selatan"},
		minerals: []weighted{{"batu_bara", 1}},
		rocks:    []world.Rock{world.RockSedimentary},
	},
	{
		DepositModel: DepositModel{"karbonat", "Batu gamping & dolomit",
			"Batuan karbonat dari terumbu dan paparan laut dangkal; dolomit terbentuk ketika magnesium menggantikan sebagian kalsium; fosfat (apatit) terkumpul di gua karst.",
			"Gunung Sewu (Gunung Kidul), Karst Maros-Pangkep (Sulawesi Selatan)"},
		minerals: []weighted{{"batu_kapur", 70}, {"dolomit", 25}, {"apatit", 5}},
		rocks:    []world.Rock{world.RockLimestone},
	},
	{
		DepositModel: DepositModel{"mvt", "Pb-Zn dalam batuan karbonat",
			"Galena dan sfalerit mengisi rongga batuan karbonat dari fluida cekungan bersuhu rendah, bersama barit dan fluorit; vanadinit di zona oksidasinya.",
			"Mississippi Valley (Amerika Serikat), Pine Point (Kanada)"},
		minerals: []weighted{{"galena", 30}, {"sfalerit", 30}, {"barit", 15}, {"fluorit", 15}, {"vanadinit", 10}},
		barren:   0.3, host: "batu_kapur",
		rocks: []world.Rock{world.RockLimestone},
	},
	{
		DepositModel: DepositModel{"evaporit", "Evaporit",
			"Garam yang mengendap saat air danau atau laguna tertutup menguap di iklim kering: gipsum, halit, silvit, boraks dan selestin. Jarang di Indonesia yang beriklim basah.",
			"Searles Lake (boraks, AS), Zechstein (kalium, Jerman)"},
		minerals: []weighted{{"gipsum", 30}, {"halit", 25}, {"silvit", 15}, {"boraks", 15}, {"selestin", 15}},
		rocks:    []world.Rock{world.RockEvaporite},
	},
	{
		DepositModel: DepositModel{"uranium_sedimen", "Uranium dalam batu pasir",
			"Uranium dan vanadium yang terlarut dari granit dan batuan vulkanik mengendap di batu pasir cekungan pada front reduksi.",
			"Colorado Plateau (Amerika Serikat)"},
		minerals: []weighted{{"uraninit", 50}, {"karnotit", 50}},
		rocks:    []world.Rock{world.RockSedimentary},
	},
	{
		DepositModel: DepositModel{"karbonatit", "Karbonatit REE-Nb",
			"Intrusi karbonatit: sumber utama unsur tanah jarang (bastnasit, monasit) dan niobium (pirokhlor), disertai apatit, barit dan fluorit.",
			"Mountain Pass (AS), Bayan Obo (Tiongkok), Araxá (Brasil)"},
		minerals: []weighted{{"bastnasit", 25}, {"pirokhlor", 20}, {"monasit", 15}, {"apatit", 15}, {"fluorit", 15}, {"barit", 10}},
		barren:   0.15, host: "batu",
		rocks: []world.Rock{world.RockCarbonatite},
	},
	{
		DepositModel: DepositModel{"mangan", "Mangan sedimen",
			"Lapisan oksida mangan di dalam rijang dan sedimen laut dalam yang terangkat bersama kompleks akresi.",
			"Pulau Timor (Nusa Tenggara Timur)"},
		minerals: []weighted{{"pirolusit", 1}},
		rocks:    []world.Rock{world.RockMetamorphic},
	},
	{
		DepositModel: DepositModel{"plaser", "Plaser sungai & pantai",
			"Mineral berat yang tahan pelapukan terbawa sungai dari batuan sumbernya lalu terkumpul di dasar sungai, delta dan pantai; diambil dengan menggali dan mendulang kerikil.",
			"Timah aluvial Bangka-Belitung, emas aluvial Kalimantan, pasir besi pantai selatan Jawa"},
		rocks: []world.Rock{world.RockAlluvium},
	},
	{DepositModel: DepositModel{"lempung", "Lempung", "Tanah liat dari endapan sungai dan pelapukan batuan.", ""}},
	{DepositModel: DepositModel{"pasir", "Pasir", "Pasir kuarsa di pantai dan gosong sungai.", ""}},
	{DepositModel: DepositModel{"batu", "Batu", "Batuan yang tersingkap; bongkah lepasnya bisa dipungut dengan tangan.", ""}},
	{DepositModel: DepositModel{"kayu", "Kayu", "Pohon yang bisa ditebang dan tumbuh kembali.", ""}},
	{DepositModel: DepositModel{"serat", "Serat tumbuhan", "Rumput dan semak untuk tali dan atap.", ""}},
	{DepositModel: DepositModel{"air", "Air tawar", "Sungai dan danau.", ""}},
	{DepositModel: DepositModel{"laut", "Laut", "Air laut dan rumput laut di perairan dangkal.", ""}},
}

var modelIndex = map[string]int{}

func init() {
	for i, s := range modelSpecs {
		modelIndex[s.Key] = i
	}
}

// DepositModels lists every deposit model in a stable order.
func DepositModels() []DepositModel {
	out := make([]DepositModel, len(modelSpecs))
	for i, s := range modelSpecs {
		out[i] = s.DepositModel
	}
	return out
}

// Heavy minerals a river carries, by the rocks it drains.
var (
	goldPlacers     = []weighted{{"bijih_emas", 1}}
	granitePlacers  = []weighted{{"kasiterit", 50}, {"ilmenit", 12}, {"zirkon", 12}, {"monasit", 8}, {"xenotim", 5}}
	volcanicPlacers = []weighted{{"pasir_besi", 30}, {"ilmenit", 8}}
)

// hostModels lists the models each mineral can come from, for relocating it
// within compatible rock on hand-edited maps.
func hostModels(id ItemID) []int {
	var out []int
	for i, s := range modelSpecs {
		has := func(ws []weighted) bool {
			return slices.ContainsFunc(ws, func(w weighted) bool { return w.item == id })
		}
		if has(s.minerals) || has(s.surface) {
			out = append(out, i)
		}
	}
	if slices.ContainsFunc(slices.Concat(goldPlacers, granitePlacers, volcanicPlacers), func(w weighted) bool { return w.item == id }) {
		out = append(out, modelIndex["plaser"])
	}
	return out
}
