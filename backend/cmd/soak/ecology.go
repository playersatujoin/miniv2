package main

import (
	"fmt"
	"strings"

	"miniv2/backend/internal/ecology"
)

// ecologySection reports the land (Fase 2): wildlife, farming, and how the
// monsoon and El Niño show in hunger.
func ecologySection(b *strings.Builder, rep report) {
	species := ecology.SpeciesList()
	b.WriteString("\n## Ekologi\n\n")
	b.WriteString("| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (")
	for i, sp := range species {
		if i > 0 {
			b.WriteString(" / ")
		}
		b.WriteString(sp.Name)
	}
	b.WriteString(") | Ternak | Punah lokal / datang lagi |\n| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |\n")
	for _, r := range rep.Runs {
		e := r.Final.Ecology
		animals := make([]string, len(e.Animals))
		for i, n := range e.Animals {
			animals[i] = fmt.Sprint(n)
		}
		fmt.Fprintf(b, "| %d | %d | %d | %s | %d | %s | %s | %d | %d / %d |\n", r.Seed, e.PeakPopulation, e.CapacityHits,
			fmtPtr(minutePtr(e.FarmingMinute), 0), e.MaxPlots, fmtPtr(num(e.Forest), 2), strings.Join(animals, " / "),
			e.Livestock, e.Extinctions, e.Arrivals)
	}
	md := rep.Median
	fmt.Fprintf(b, "| **Median** | %s | %s | %s | %s | %s | | | %s |\n\n", fmtPtr(md["peakPopulation"], 0), fmtPtr(md["capacityHits"], 0),
		fmtPtr(md["farmingMinute"], 0), fmtPtr(md["maxPlots"], 0), fmtPtr(md["forest"], 2), fmtPtr(md["extinctions"], 0))

	// How much of the time each species was on the island.
	b.WriteString("### Keberadaan satwa (bagian waktu spesies itu ada di pulau)\n\n| Spesies | Median | Terendah | Dunia yang masih punya di akhir |\n| --- | ---: | ---: | ---: |\n")
	for i, sp := range species {
		var vs []float64
		lo, alive := 1.0, 0
		for _, r := range rep.Runs {
			p := r.Final.Ecology.Presence[i]
			vs = append(vs, p)
			lo = min(lo, p)
			if r.Final.Ecology.Animals[i] > 0 {
				alive++
			}
		}
		fmt.Fprintf(b, "| %s | %s%% | %s%% | %d/%d |\n", sp.Name, fmtPtr(num(*median(vs)*100), 0), fmtPtr(num(lo*100), 0), alive, len(rep.Runs))
	}

	// Where the food came from, over the last century of each world.
	b.WriteString("\n### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)\n\n| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |\n| ---: | ---: | ---: | ---: | ---: | ---: |\n")
	var crops []float64
	for _, r := range rep.Runs {
		sh, perHead := foodShares(r.Final.Ecology.Years, 100)
		if sh == nil {
			fmt.Fprintf(b, "| %d | – | – | – | – | – |\n", r.Seed)
			continue
		}
		crops = append(crops, sh[1])
		fmt.Fprintf(b, "| %d | %s%% | %s%% | %s%% | %s%% | %s |\n", r.Seed, fmtPtr(num(sh[0]*100), 0), fmtPtr(num(sh[1]*100), 0),
			fmtPtr(num(sh[2]*100), 0), fmtPtr(num(sh[3]*100), 0), fmtPtr(num(perHead), 2))
	}
	if len(crops) > 0 {
		fmt.Fprintf(b, "| **Median** | | %s%% | | | |\n", fmtPtr(num(*median(crops)*100), 0))
	}

	// Hunger by ENSO state, pooled over worlds, counting years with at least
	// 20 people (smaller bands are too noisy).
	type tally struct{ years, people, starved, nextPeople, nextStarved, births, harvest int }
	var by [3]tally
	for _, r := range rep.Runs {
		ys := r.Final.Ecology.Years
		for i, y := range ys {
			if y.Population < 20 {
				continue
			}
			t := &by[y.ENSO+1]
			t.years++
			t.people += y.Population
			t.starved += y.Starved
			t.births += y.Births
			t.harvest += y.Harvest
			if i+1 < len(ys) && ys[i+1].Population >= 20 {
				t.nextPeople += ys[i+1].Population
				t.nextStarved += ys[i+1].Starved
			}
		}
	}
	b.WriteString("\n### Iklim dan kelaparan\n\nTahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.\n\n")
	b.WriteString("| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |\n| --- | ---: | ---: | ---: | ---: | ---: |\n")
	per := func(n, d, scale int) *float64 {
		if d == 0 {
			return nil
		}
		return num(float64(n) * float64(scale) / float64(d))
	}
	for _, k := range []struct {
		name string
		i    int
	}{{"El Niño", 2}, {"Netral", 1}, {"La Niña", 0}} {
		t := by[k.i]
		fmt.Fprintf(b, "| %s | %d | %s | %s | %s | %s |\n", k.name, t.years, fmtPtr(per(t.starved, t.people, 1000), 1),
			fmtPtr(per(t.nextStarved, t.nextPeople, 1000), 1), fmtPtr(per(t.births, t.people, 1000), 1), fmtPtr(per(t.harvest, t.people, 100), 1))
	}

	// The same, paired: each ENSO year against the neutral years around it
	// in the same world, so slow changes in population don't hide the shock.
	b.WriteString("\nDibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:\n\n")
	b.WriteString("| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |\n| --- | ---: | ---: | ---: | ---: |\n")
	for _, k := range []struct {
		name string
		enso int
	}{{"El Niño", 1}, {"La Niña", -1}} {
		n, same, next, spikes := ensoShock(rep.Runs, k.enso)
		if n == 0 {
			fmt.Fprintf(b, "| %s | 0 | – | – | – |\n", k.name)
			continue
		}
		fmt.Fprintf(b, "| %s | %d | ×%s | ×%s | %s%% |\n", k.name, n, fmtPtr(num(same), 2), fmtPtr(num(next), 2), fmtPtr(num(spikes*100), 0))
	}
}

// ensoShock compares starvation in years of the given ENSO state, and in
// the year after, with the mean of the neutral years within ten years in
// the same world (years with at least 20 people). It returns the number of
// events, the mean ratios, and the share of events whose following year
// starved more than 1.5 times the usual.
func ensoShock(runs []run, enso int) (n int, same, next, spikes float64) {
	const window, minPop = 10, 20
	nNext := 0
	for _, r := range runs {
		ys := r.Final.Ecology.Years
		rate := func(i int) (float64, bool) {
			if i < 0 || i >= len(ys) || ys[i].Population < minPop {
				return 0, false
			}
			return float64(ys[i].Starved) / float64(ys[i].Population), true
		}
		for i, y := range ys {
			v, ok := rate(i)
			if y.ENSO != enso || !ok {
				continue
			}
			base, k := 0.0, 0
			for j := i - window; j <= i+window; j++ {
				if j == i || j == i+1 || j < 0 || j >= len(ys) || ys[j].ENSO != 0 {
					continue
				}
				if w, ok := rate(j); ok {
					base += w
					k++
				}
			}
			if k < 4 || base == 0 {
				continue
			}
			base /= float64(k)
			n++
			same += v / base
			if w, ok := rate(i + 1); ok {
				nNext++
				next += w / base
				if w > 1.5*base {
					spikes++
				}
			}
		}
	}
	if n > 0 {
		same /= float64(n)
	}
	if nNext > 0 {
		next /= float64(nNext)
		spikes /= float64(nNext)
	}
	return n, same, next, spikes
}

// foodShares splits the energy eaten over the last n years into wild,
// crops, fish and meat, and gives the units harvested per person-year.
func foodShares(ys []ecology.YearRecord, n int) ([]float64, float64) {
	ys = ys[max(0, len(ys)-n):]
	var sum [4]float64
	harvest, people := 0, 0
	for _, y := range ys {
		sum[0] += y.FoodWild
		sum[1] += y.FoodCrops
		sum[2] += y.FoodFish
		sum[3] += y.FoodMeat
		harvest += y.Harvest
		people += y.Population
	}
	total := sum[0] + sum[1] + sum[2] + sum[3]
	if total == 0 || people == 0 {
		return nil, 0
	}
	return []float64{sum[0] / total, sum[1] / total, sum[2] / total, sum[3] / total}, float64(harvest) / float64(people)
}
