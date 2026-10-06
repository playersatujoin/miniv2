package main

import (
	"fmt"
	"strings"

	"miniv2/backend/internal/sim"
)

// nutritionFinal sums up a run's nutrition (Fase 3d): shares averaged over
// every sampled minute (pooled, so a minute with more people weighs more),
// and the state at the end. Shares are nil in a world without nutrition.
type nutritionFinal struct {
	Deficient    *float64          `json:"deficient"`    // short of protein or micronutrients
	Severe       *float64          `json:"severe"`       // severely
	Malnourished *float64          `json:"malnourished"` // "kurang" or "buruk" for any reason, thinness included
	Thin         *float64          `json:"thin"`
	ProteinShort *float64          `json:"proteinShort"`
	MicroShort   *float64          `json:"microShort"`
	Stunted5     *float64          `json:"stunted5"`     // of children under five
	StuntedChild *float64          `json:"stuntedChild"` // of children under fifteen
	StuntedAdult *float64          `json:"stuntedAdult"` // of adults
	MothersShort *float64          `json:"mothersShort"` // of pregnant and nursing women
	DietProtein  *float64          `json:"dietProtein"`  // mean diet against need
	DietMicro    *float64          `json:"dietMicro"`
	End          sim.NutritionInfo `json:"end"`
}

// nutritionTally adds up the minutes of a run.
type nutritionTally struct {
	sum sim.NutritionInfo
	// Mean diets are averaged per person: weighted sums.
	dietP, dietM float64
}

func (t *nutritionTally) add(v sim.NutritionInfo) {
	s := &t.sum
	s.People += v.People
	s.Deficient += v.Deficient
	s.Severe += v.Severe
	s.ProteinShort += v.ProteinShort
	s.MicroShort += v.MicroShort
	s.Thin += v.Thin
	s.Malnourished += v.Malnourished
	s.Under5 += v.Under5
	s.Stunted5 += v.Stunted5
	s.Children += v.Children
	s.StuntedChild += v.StuntedChild
	s.Adults += v.Adults
	s.StuntedAdults += v.StuntedAdults
	s.PregnantOrNurs += v.PregnantOrNurs
	s.PregnantShort += v.PregnantShort
	t.dietP += v.DietProtein * float64(v.People)
	t.dietM += v.DietMicro * float64(v.People)
}

func (t *nutritionTally) final(end sim.NutritionInfo) nutritionFinal {
	s := t.sum
	share := func(n, of int) *float64 {
		if of == 0 {
			return nil
		}
		return num(float64(n) / float64(of))
	}
	f := nutritionFinal{
		Deficient:    share(s.Deficient, s.People),
		Severe:       share(s.Severe, s.People),
		Malnourished: share(s.Malnourished, s.People),
		Thin:         share(s.Thin, s.People),
		ProteinShort: share(s.ProteinShort, s.People),
		MicroShort:   share(s.MicroShort, s.People),
		Stunted5:     share(s.Stunted5, s.Under5),
		StuntedChild: share(s.StuntedChild, s.Children),
		StuntedAdult: share(s.StuntedAdults, s.Adults),
		MothersShort: share(s.PregnantShort, s.PregnantOrNurs),
		End:          end,
	}
	if s.People > 0 {
		f.DietProtein = num(t.dietP / float64(s.People))
		f.DietMicro = num(t.dietM / float64(s.People))
	}
	return f
}

var nutritionMetrics = []metric{
	{"deficientShare", "Kurang protein/zat gizi mikro", "", 2, func(f final) *float64 { return f.Nutrition.Deficient }},
	{"severeShare", "Kekurangan berat", "", 2, func(f final) *float64 { return f.Nutrition.Severe }},
	{"malnourishedShare", "Gizi kurang atau buruk (termasuk kurus)", "", 2, func(f final) *float64 { return f.Nutrition.Malnourished }},
	{"stunted5Share", "Balita pendek (stunting)", "", 2, func(f final) *float64 { return f.Nutrition.Stunted5 }},
	{"stuntedChildShare", "Anak <15 pendek", "", 2, func(f final) *float64 { return f.Nutrition.StuntedChild }},
	{"stuntedAdultShare", "Dewasa pendek", "", 2, func(f final) *float64 { return f.Nutrition.StuntedAdult }},
}

// nutritionSection reports nutrition per world (Fase 3d).
func nutritionSection(b *strings.Builder, rep report) {
	b.WriteString("\n## Gizi (Fase 3d)\n\n")
	b.WriteString("Bagian penduduk, dirata-rata atas setiap menit simulasi. Stunting: tinggi badan sekitar 2 SD di bawah anak sebaya yang cukup gizi.\n\n")
	b.WriteString("| Seed | Kurang protein/mikro | Berat | Gizi kurang/buruk | Kurus | Kurang protein | Kurang mikro | Balita pendek | Anak <15 pendek | Dewasa pendek | Ibu hamil/menyusui kurang | Diet protein | Diet mikro |\n")
	b.WriteString("| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	pct := func(v *float64) string {
		if v == nil {
			return "–"
		}
		return fmtPtr(num(*v*100), 0) + "%"
	}
	for _, r := range rep.Runs {
		n := r.Final.Nutrition
		fmt.Fprintf(b, "| %d | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n", r.Seed,
			pct(n.Deficient), pct(n.Severe), pct(n.Malnourished), pct(n.Thin), pct(n.ProteinShort), pct(n.MicroShort),
			pct(n.Stunted5), pct(n.StuntedChild), pct(n.StuntedAdult), pct(n.MothersShort),
			fmtPtr(n.DietProtein, 2), fmtPtr(n.DietMicro, 2))
	}
	md := rep.Median
	fmt.Fprintf(b, "| **Median** | %s | %s | %s | | | | %s | %s | %s | | | |\n",
		pct(md["deficientShare"]), pct(md["severeShare"]), pct(md["malnourishedShare"]),
		pct(md["stunted5Share"]), pct(md["stuntedChildShare"]), pct(md["stuntedAdultShare"]))
	b.WriteString("\nAcuan: stunting balita di dunia 26% pada 2011, 30–50% di Asia Selatan dan Afrika sub-Sahara (Black dkk. 2013, Lancet 382:427); petani awal lebih pendek dan lebih sering kurang gizi daripada pemburu-peramu (Cohen & Armelagos 1984).\n")
}
