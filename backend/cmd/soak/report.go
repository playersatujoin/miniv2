package main

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"miniv2/backend/internal/sim"
)

// metric reads one number from a run's final state; nil when not available.
type metric struct {
	key   string
	label string
	unit  string
	digit int
	get   func(f final) *float64
}

func num(v float64) *float64 { return &v }

func intNum(v int) *float64 { return num(float64(v)) }

var demographyMetrics = []metric{
	{"lifeExpectancy", "Harapan hidup saat lahir (e0)", "tahun", 1, func(f final) *float64 { return f.Demography.LifeExpectancy }},
	{"survivalTo15", "Peluang hidup sampai umur 15 (l15)", "", 2, func(f final) *float64 { return f.Demography.SurvivalTo15 }},
	{"lifeExpectancy15", "Sisa harapan hidup pada umur 15 (e15)", "tahun", 1, func(f final) *float64 { return f.Demography.LifeExpectancy15 }},
	{"modalAgeAdultDeath", "Modus usia kematian dewasa", "tahun", 0, func(f final) *float64 { return f.Demography.ModalAgeAdultDeath }},
	{"tfr", "Angka kelahiran total (TFR)", "anak", 1, func(f final) *float64 { return f.Demography.TFR }},
	{"meanBirthInterval", "Jarak antar-kelahiran", "tahun", 1, func(f final) *float64 { return f.Demography.MeanBirthInterval }},
	{"meanAgeFirstBirth", "Umur ibu saat anak pertama", "tahun", 1, func(f final) *float64 { return f.Demography.MeanAgeFirstBirth }},
	{"gini", "Ketimpangan kekayaan (Gini)", "", 2, func(f final) *float64 { return f.Demography.Gini }},
	{"infantMortality", "Kematian bayi (q0)", "", 3, func(f final) *float64 { return f.Demography.InfantMortality }},
	{"sexRatio", "Rasio kelamin (♂ per 100 ♀)", "", 0, func(f final) *float64 { return f.Demography.SexRatio }},
	{"householdSize", "Anggota per rumah", "orang", 1, func(f final) *float64 { return f.Demography.HouseholdSize }},
	{"homicideRate", "Pembunuhan per 100.000 tahun-orang", "", 0, func(f final) *float64 { return f.Demography.HomicideRate }},
}

var worldMetrics = []metric{
	{"era1LastedMinutes", "Era 1 bertahan (menit sim)", "", 0, func(f final) *float64 { return intNum(f.Era1LastedMinutes) }},
	{"firstHouseMinute", "Rumah pertama (menit sim)", "", 0, func(f final) *float64 {
		if f.FirstHouseMinute == nil {
			return nil
		}
		return intNum(*f.FirstHouseMinute)
	}},
	{"population", "Populasi", "", 0, func(f final) *float64 { return intNum(f.Population) }},
	{"maxGeneration", "Generasi maks", "", 0, func(f final) *float64 { return intNum(f.MaxGeneration) }},
	{"tier", "Zaman", "", 0, func(f final) *float64 { return intNum(f.Tier) }},
	{"elements", "Unsur", "", 0, func(f final) *float64 { return intNum(f.Elements) }},
	{"houses", "Rumah", "", 0, func(f final) *float64 { return intNum(f.Houses) }},
	{"crimes", "Kejahatan", "", 0, func(f final) *float64 { return intNum(f.Crimes) }},
	{"kindness", "Kebaikan", "", 0, func(f final) *float64 { return intNum(f.Kindness) }},
	{"kills", "Pembunuhan", "", 0, func(f final) *float64 { return intNum(f.Kills) }},
	{"msPerTick", "ms/tick", "", 2, func(f final) *float64 { return num(f.MsPerTick) }},
}

func median(vs []float64) *float64 {
	if len(vs) == 0 {
		return nil
	}
	s := slices.Clone(vs)
	slices.Sort(s)
	n := len(s)
	if n%2 == 1 {
		return num(s[n/2])
	}
	return num((s[n/2-1] + s[n/2]) / 2)
}

func medians(runs []run) map[string]*float64 {
	out := map[string]*float64{}
	for _, m := range append(slices.Clone(worldMetrics), demographyMetrics...) {
		var vs []float64
		for _, r := range runs {
			if v := m.get(r.Final); v != nil {
				vs = append(vs, *v)
			}
		}
		out[m.key] = median(vs)
	}
	return out
}

// fmtPtr formats a number with Indonesian decimal commas, or "–" for nil.
func fmtPtr(v *float64, digits int) string {
	if v == nil {
		return "–"
	}
	return strings.Replace(strconv.FormatFloat(*v, 'f', digits, 64), ".", ",", 1)
}

func status(v *float64, ref sim.MetricRef) string {
	switch {
	case v == nil:
		return "–"
	case *v < ref.Low:
		return "↓ di bawah"
	case *v > ref.High:
		return "↑ di atas"
	}
	return "✓ dalam rentang"
}

func markdown(rep report) string {
	var b strings.Builder
	title := rep.Label
	if title == "" {
		title = rep.GeneratedAt.Format("2006-01-02 15:04")
	}
	cfg := rep.Config
	fmt.Fprintf(&b, "# Laporan soak — %s\n\n", title)
	fmt.Fprintf(&b, "- Dibuat: %s\n", rep.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "- Peta: %d×%d, seed peta %d · seed dunia: %s\n", cfg.MapSize, cfg.MapSize, cfg.MapSeed, seedList(cfg.Seeds))
	fmt.Fprintf(&b, "- Durasi: %d menit simulasi per dunia (≈ %.0f tahun; 1 tahun = %.0f detik simulasi)\n",
		cfg.Minutes, cfg.SimYears, cfg.SecondsPerYear)
	off := "tidak ada"
	if len(cfg.Off) > 0 {
		off = strings.Join(cfg.Off, ", ")
	}
	fmt.Fprintf(&b, "- Aturan dimatikan: %s\n", off)
	b.WriteString("- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.\n\n")

	b.WriteString("## Per dunia\n\n")
	b.WriteString("| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |\n")
	b.WriteString("| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for _, r := range rep.Runs {
		f, d := r.Final, r.Final.Demography
		era1 := fmt.Sprintf("%d m", f.Era1LastedMinutes)
		if f.Eras == 1 {
			era1 += " (utuh)"
		}
		fmt.Fprintf(&b, "| %d | %s | %d | %d | %d | %d | %d | %d | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			r.Seed, era1, f.Eras, f.Population, f.MaxGeneration, f.Tier, f.Elements, f.Houses,
			fmtPtr(d.LifeExpectancy, 1), fmtPtr(d.SurvivalTo15, 2), fmtPtr(d.LifeExpectancy15, 1), fmtPtr(d.TFR, 1),
			fmtPtr(d.MeanBirthInterval, 1), fmtPtr(d.MeanAgeFirstBirth, 1), fmtPtr(d.ModalAgeAdultDeath, 0),
			fmtPtr(d.Gini, 2), fmtPtr(d.HomicideRate, 0), fmtPtr(num(f.MsPerTick), 2))
	}
	md := rep.Median
	fmt.Fprintf(&b, "| **Median** | %s m | | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n\n",
		fmtPtr(md["era1LastedMinutes"], 0), fmtPtr(md["population"], 0), fmtPtr(md["maxGeneration"], 0),
		fmtPtr(md["tier"], 0), fmtPtr(md["elements"], 0), fmtPtr(md["houses"], 0),
		fmtPtr(md["lifeExpectancy"], 1), fmtPtr(md["survivalTo15"], 2), fmtPtr(md["lifeExpectancy15"], 1),
		fmtPtr(md["tfr"], 1), fmtPtr(md["meanBirthInterval"], 1), fmtPtr(md["meanAgeFirstBirth"], 1),
		fmtPtr(md["modalAgeAdultDeath"], 0), fmtPtr(md["gini"], 2), fmtPtr(md["homicideRate"], 0), fmtPtr(md["msPerTick"], 2))

	survived := 0
	for _, r := range rep.Runs {
		if r.Final.Eras == 1 {
			survived++
		}
	}
	lo, hi := wilson(survived, len(rep.Runs))
	fmt.Fprintf(&b, "Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **%d dari %d** dunia "+
		"(selang kepercayaan 95%%: %s–%s%%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.\n\n",
		survived, len(rep.Runs), fmtPtr(num(lo*100), 0), fmtPtr(num(hi*100), 0))

	b.WriteString("## Dibanding acuan pra-modern\n\n")
	b.WriteString("| Indikator | Median simulasi | Acuan | Status | Sumber |\n| --- | ---: | ---: | --- | --- |\n")
	for _, m := range demographyMetrics {
		ref, ok := rep.Reference[m.key]
		if !ok {
			continue
		}
		unit := ""
		if m.unit != "" {
			unit = " " + m.unit
		}
		src := ref.Source
		if ref.Note != "" {
			src += " — " + ref.Note
		}
		fmt.Fprintf(&b, "| %s | %s%s | %s–%s | %s | %s |\n", m.label, fmtPtr(md[m.key], m.digit), unit,
			fmtPtr(num(ref.Low), -1), fmtPtr(num(ref.High), -1), status(md[m.key], ref), src)
	}
	b.WriteString("\nKeterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.\n\n")

	b.WriteString("## Indikator lain (tanpa rentang acuan)\n\n| Indikator | Median simulasi |\n| --- | ---: |\n")
	for _, m := range demographyMetrics {
		if _, ok := rep.Reference[m.key]; ok {
			continue
		}
		fmt.Fprintf(&b, "| %s | %s |\n", m.label, fmtPtr(md[m.key], m.digit))
	}

	var causes [4]int
	for _, r := range rep.Runs {
		c := r.Final.Demography.DeathsByCause
		causes[0] += c.Starvation
		causes[1] += c.Thirst
		causes[2] += c.OldAge
		causes[3] += c.Killed
	}
	total := causes[0] + causes[1] + causes[2] + causes[3]
	b.WriteString("\n## Penyebab kematian (semua dunia, 50 tahun terakhir)\n\n| Penyebab | Kematian | Bagian |\n| --- | ---: | ---: |\n")
	for i, name := range []string{"Kelaparan", "Kehausan", "Usia tua", "Dibunuh"} {
		share := 0.0
		if total > 0 {
			share = float64(causes[i]) * 100 / float64(total)
		}
		fmt.Fprintf(&b, "| %s | %d | %s%% |\n", name, causes[i], fmtPtr(num(share), 0))
	}
	b.WriteString("\nCatatan: simulasi belum punya penyakit (Fase 3), sedangkan di masyarakat nyata penyakit menyebabkan lebih dari separuh kematian. Perbedaan ini temuan, bukan galat.\n")
	return b.String()
}

// wilson is the 95 % Wilson score interval for k successes out of n.
func wilson(k, n int) (lo, hi float64) {
	if n == 0 {
		return 0, 1
	}
	const z = 1.96
	p, fn := float64(k)/float64(n), float64(n)
	centre := (p + z*z/(2*fn)) / (1 + z*z/fn)
	half := z * math.Sqrt(p*(1-p)/fn+z*z/(4*fn*fn)) / (1 + z*z/fn)
	return math.Max(0, centre-half), math.Min(1, centre+half)
}

func seedList(seeds []uint64) string {
	parts := make([]string, len(seeds))
	for i, s := range seeds {
		parts[i] = strconv.FormatUint(s, 10)
	}
	return strings.Join(parts, ", ")
}
