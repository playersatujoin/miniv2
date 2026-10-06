package main

import (
	"fmt"
	"strings"
)

// geneticsSection reports inbreeding (Fase 3c): by world, by generation and
// over time pooled over all worlds, what became of the inbred, and whether
// adults came to want close kin as mates less than others.
func geneticsSection(b *strings.Builder, rep report) {
	if len(rep.Runs) == 0 || len(rep.Runs[0].Final.Genetics.ByGeneration) == 0 {
		return
	}
	b.WriteString("\n## Genetika: inbreeding dan pemilihan pasangan\n\n")
	b.WriteString("F = koefisien inbreeding (1/16 anak sepupu, 1/4 anak saudara kandung), dihitung dari silsilah 7 generasi ke atas. " +
		"Kerabat dekat = F ≥ 1/8; sedarah = F ≥ 1/64 (sepupu dua kali atau lebih dekat).\n\n")
	b.WriteString("| Seed | Kelahiran | F rata-rata (hidup) | Sedarah (hidup) | Kerabat dekat (hidup) | Varian dibawa /orang | Punya kelainan | Bayi wafat karena kelainan |\n")
	b.WriteString("| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for _, r := range rep.Runs {
		g := r.Final.Genetics
		births := 0
		for _, band := range g.ByGeneration {
			births += band.Births
		}
		fmt.Fprintf(b, "| %d | %d | %s | %s | %s | %s | %s | %d |\n", r.Seed, births, fmtPtr(num(g.MeanF), 4),
			pct(g.Inbred), pct(g.CloseKin), fmtPtr(num(g.Carried), 2), pct(g.Affected), g.Lethal)
	}

	b.WriteString("\n### Menurut generasi anak (semua dunia)\n\n| Generasi | Kelahiran | F rata-rata | Sedarah | Kerabat dekat |\n| --- | ---: | ---: | ---: | ---: |\n")
	for i, band := range rep.Runs[0].Final.Genetics.ByGeneration {
		births, sumF, consan, close := 0, 0.0, 0.0, 0.0
		for _, r := range rep.Runs {
			gb := r.Final.Genetics.ByGeneration[i]
			births += gb.Births
			sumF += gb.MeanF * float64(gb.Births)
			consan += gb.Consanguineous * float64(gb.Births)
			close += gb.CloseKin * float64(gb.Births)
		}
		if births == 0 {
			continue
		}
		n := float64(births)
		fmt.Fprintf(b, "| %s | %d | %s | %s | %s |\n", band.Label, births, fmtPtr(num(sumF/n), 4), pct(consan/n), pct(close/n))
	}

	b.WriteString("\n### Nasib anak menurut F (semua dunia)\n\n" +
		"| F | Kelahiran | Lahir dengan kelainan resesif | Wafat minggu pertama karena kelainan | Wafat < 1 th | Wafat < 15 th | Wafat < 15 per kelahiran |\n" +
		"| --- | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for i, band := range rep.Runs[0].Final.Genetics.ByF {
		var births, affected, lethal, u1, u15 int
		for _, r := range rep.Runs {
			o := r.Final.Genetics.ByF[i]
			births += o.Births
			affected += o.Affected
			lethal += o.Lethal
			u1 += o.Under1
			u15 += o.Under15
		}
		q := "–"
		if births > 0 {
			q = fmtPtr(num(float64(u15)/float64(births)), 3)
		}
		fmt.Fprintf(b, "| %s | %d | %d | %d | %d | %d | %s |\n", band.Label, births, affected, lethal, u1, u15, q)
	}
	b.WriteString("\nAnak yang masih kecil di akhir run belum sempat wafat, jadi angka per kelahiran adalah batas bawah; " +
		"bandingkan antar-baris dan antar-konfigurasi, bukan dengan l15.\n")

	// Over time: pool the periods of every world by their first year.
	type acc struct {
		births                                         int
		sumF, close                                    float64
		kinSeen, kinWant, otherSeen, otherWant, otherR float64
	}
	byYear := map[int]*acc{}
	top := -1
	for _, r := range rep.Runs {
		for _, p := range r.Final.Genetics.Periods {
			a := byYear[p.Year]
			if a == nil {
				a = &acc{}
				byYear[p.Year] = a
			}
			a.births += p.Births
			a.sumF += p.MeanF * float64(p.Births)
			a.close += p.CloseKin * float64(p.Births)
			a.kinSeen += float64(p.KinSeen)
			a.otherSeen += float64(p.OtherSeen)
			a.otherR += p.OtherR * float64(p.OtherSeen)
			if p.KinDesire != nil {
				a.kinWant += *p.KinDesire * float64(p.KinSeen)
			}
			if p.OtherDesire != nil {
				a.otherWant += *p.OtherDesire * float64(p.OtherSeen)
			}
			top = max(top, p.Year)
		}
	}
	b.WriteString("\n### Dari waktu ke waktu (semua dunia, per 25 tahun)\n\n" +
		"Keinginan kawin = bagian saat orang dewasa yang sedang melihat calon pasangan dewasa memilih *kawin*, " +
		"bila calon itu orang tua, anak, atau saudara (kandung atau tiri) — kerabat yang tumbuh bersama, sasaran efek Westermarck — " +
		"dan bila orang lain. r lain = rata-rata koefisien kekerabatan dengan calon lain itu.\n\n" +
		"| Tahun | Kelahiran | F rata-rata | Anak kerabat dekat | Keinginan kawin: keluarga inti | : orang lain | r lain | Sampel keluarga / lain |\n" +
		"| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	for y := 0; y <= top; y += 25 {
		a := byYear[y]
		if a == nil {
			continue
		}
		meanF, close := "–", "–"
		if a.births > 0 {
			meanF, close = fmtPtr(num(a.sumF/float64(a.births)), 4), pct(a.close/float64(a.births))
		}
		kin, other, r := "–", "–", "–"
		if a.kinSeen >= 20 {
			kin = pct(a.kinWant / a.kinSeen)
		}
		if a.otherSeen >= 20 {
			other, r = pct(a.otherWant/a.otherSeen), fmtPtr(num(a.otherR/a.otherSeen), 2)
		}
		fmt.Fprintf(b, "| %d–%d | %d | %s | %s | %s | %s | %s | %.0f / %.0f |\n", y+1, y+25, a.births, meanF, close, kin, other, r, a.kinSeen, a.otherSeen)
	}

	b.WriteString("\n### Frekuensi varian resesif di akhir run (median antar-dunia)\n\n| Kelainan | Frekuensi alel |\n| --- | ---: |\n")
	for i, v := range rep.Runs[0].Final.Genetics.Variants {
		var vs []float64
		for _, r := range rep.Runs {
			if r.Final.Population > 0 && i < len(r.Final.Genetics.Variants) {
				vs = append(vs, r.Final.Genetics.Variants[i].Frequency)
			}
		}
		fmt.Fprintf(b, "| %s | %s |\n", v.Name, fmtPtr(median(vs), 4))
	}
}

func pct(v float64) string { return fmtPtr(num(v*100), 1) + "%" }
