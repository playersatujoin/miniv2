# Laporan soak — Fase 3c tanpa genetika (900 tahun)

- Dibuat: 2026-10-07 00:29:29
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: genetics
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 120 m (utuh) | 1 | 129 | 46 | 1 | 5 | 12 | 25,3 | 0,66 | 20,2 | 6,1 | 3,8 | 18,5 | 18 | 0,54 | 140 | 0,48 |
| 2 | 120 m (utuh) | 1 | 804 | 46 | 1 | 6 | 7 | 21,5 | 0,60 | 16,7 | 8,7 | 3,0 | 17,0 | 20 | 0,58 | 91 | 1,58 |
| 3 | 120 m (utuh) | 1 | 413 | 46 | 1 | 5 | 15 | 19,1 | 0,63 | 11,5 | 7,9 | 3,2 | 17,5 | 18 | 0,45 | 72 | 1,62 |
| 4 | 120 m (utuh) | 1 | 171 | 46 | 1 | 5 | 25 | 28,0 | 0,72 | 21,1 | 4,7 | 4,7 | 19,7 | 18 | 0,41 | 150 | 0,55 |
| 5 | 120 m (utuh) | 1 | 160 | 46 | 1 | 5 | 9 | 27,6 | 0,71 | 20,8 | 5,8 | 4,4 | 18,6 | 18 | 0,44 | 15 | 0,64 |
| 6 | 120 m (utuh) | 1 | 48 | 44 | 1 | 6 | 4 | 22,9 | 0,69 | 15,7 | 7,5 | 3,0 | 17,7 | 20 | 0,61 | 109 | 0,28 |
| 7 | 120 m (utuh) | 1 | 301 | 46 | 0 | 5 | 13 | 27,8 | 0,77 | 19,6 | 6,7 | 3,9 | 17,9 | 18 | 0,47 | 71 | 1,02 |
| 8 | 120 m (utuh) | 1 | 192 | 46 | 1 | 6 | 19 | 27,1 | 0,73 | 19,6 | 5,6 | 4,2 | 18,8 | 18 | 0,53 | 134 | 0,66 |
| **Median** | 120 m | | 182 | 46 | 1 | 5 | 12 | 26,2 | 0,70 | 19,6 | 6,4 | 3,8 | 18,2 | 18 | 0,50 | 100 | 0,65 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **8 dari 8** dunia (selang kepercayaan 95%: 68–100%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 26,2 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,70 | 0,44–0,73 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 19,6 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 18 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 6,4 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,8 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 18,2 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,50 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,091 | 0,13–0,41 | ↓ di bawah | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,145 |
| Diare per orang per tahun | 0,24 |
| Malaria per orang per tahun | 0,90 |
| ISPA per orang per tahun | 0,01 |
| Rasio kelamin (♂ per 100 ♀) | 94 |
| Anggota per rumah | 9,3 |
| Pembunuhan per 100.000 tahun-orang | 100 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 2304 | 46% | 5 | 1% |
| Kehausan | 989 | 20% | 5 | 1% |
| Usia tua | 418 | 8% | 0 | 0% |
| Dibunuh | 104 | 2% | 42 | 5% |
| Diterkam hewan | 0 | 0% | 0 | 0% |
| Melahirkan | 73 | 1% | 0 | 0% |
| Neonatal (minggu pertama) | 350 | 7% | 350 | 44% |
| Diare | 319 | 6% | 177 | 22% |
| Malaria | 454 | 9% | 201 | 25% |
| Radang paru (ISPA) | 37 | 1% | 16 | 2% |
| Terbakar | 0 | 0% | 0 | 0% |
| Tenggelam | 0 | 0% | 0 | 0% |
| Terjatuh | 1 | 0% | 0 | 0% |
| **Penyakit menular** | 810 | 16% | 394 | 49% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 6 | – | 1 | 69 | 27,8 | 0,73 |
| 2 | 5 | – | 1 | 24 | 28,6 | 0,97 |
| 3 | 7 | – | 1 | 34 | 28,8 | 0,94 |
| 4 | 5 | – | 1 | 5 | 27,4 | 0,32 |
| 5 | 7 | – | 1 | 79 | 29,9 | 0,87 |
| 6 | 4 | – | 1 | 64 | 29,6 | 0,62 |
| 7 | – | – | 0 | 42 | 34,5 | 0,92 |
| 8 | 6 | – | 1 | 71 | 33,8 | 0,75 |
| **Median** | 6 | – | 1 | 53 | 29,2 | 0,81 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,84 | 1158 |
| 15–29 | 0,91 | 611 |
| 30–44 | 0,87 | 235 |
| 45–59 | 0,86 | 121 |
| 60+ | 0,89 | 93 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 34,0 |
| 10–19 | 32,1 |
| 20–29 | 31,7 |
| 30–39 | 31,0 |
| 40–49 | 30,2 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 228 | 0 | 16 | 49 | 0,94 | 148 / 189 / 0 / 0 / 5 | 0 | 40 / 38 |
| 2 | 895 | 0 | 8 | 261 | 0,95 | 155 / 170 / 18 / 0 / 3 | 0 | 37 / 36 |
| 3 | 727 | 0 | 8 | 256 | 0,95 | 133 / 130 / 52 / 12 / 0 | 0 | 34 / 33 |
| 4 | 192 | 0 | 5 | 65 | 0,96 | 95 / 188 / 32 / 20 / 2 | 0 | 18 / 18 |
| 5 | 307 | 0 | 6 | 198 | 0,96 | 143 / 100 / 26 / 20 / 0 | 0 | 30 / 29 |
| 6 | 90 | 0 | 11 | 61 | 0,98 | 116 / 145 / 8 / 18 / 0 | 0 | 38 / 37 |
| 7 | 337 | 0 | 6 | 346 | 0,94 | 135 / 110 / 16 / 0 / 0 | 0 | 32 / 30 |
| 8 | 217 | 0 | 7 | 145 | 0,95 | 159 / 153 / 3 / 16 / 0 | 0 | 25 / 24 |
| **Median** | 268 | 0 | 8 | 172 | 0,95 | | | 33 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 100% | 8/8 |
| Ayam Hutan | 99% | 61% | 7/8 |
| Kerbau Liar | 90% | 63% | 5/8 |
| Harimau | 53% | 42% | 3/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 70% | 29% | 0% | 0% | 0,24 |
| 2 | 14% | 85% | 0% | 0% | 1,05 |
| 3 | 17% | 82% | 0% | 0% | 1,08 |
| 4 | 98% | 0% | 1% | 0% | 0,01 |
| 5 | 52% | 47% | 0% | 0% | 0,53 |
| 6 | 95% | 0% | 5% | 0% | 0,01 |
| 7 | 55% | 44% | 1% | 1% | 0,72 |
| 8 | 78% | 21% | 1% | 0% | 0,42 |
| **Median** | | 36% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 1096 | 16,0 | 18,6 | 40,3 | 73,5 |
| Netral | 3838 | 15,8 | 15,4 | 41,3 | 73,4 |
| La Niña | 1775 | 15,7 | 15,0 | 41,0 | 71,3 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 1054 | ×1,19 | ×1,34 | 31% |
| La Niña | 1711 | ×1,07 | ×1,00 | 22% |

## Engine II: pertukaran, desa, api, air

| Seed | Bicara | Barter | Kabar | Desa | Warga desa | Ganti pemimpin | Kebakaran | Bangunan terbakar | Rakit dibawa | Tenggelam | Terjatuh | Stimulus aktif |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 47831 | 2442 | 151440 | 2 | 104 | 93 | 117 | 6 | 8 | 0 | 4 | 9 |
| 2 | 125388 | 7062 | 471925 | 2 | 412 | 131 | 63 | 0 | 5 | 0 | 2 | 39 |
| 3 | 74266 | 8003 | 271878 | 2 | 166 | 205 | 156 | 3 | 1 | 0 | 0 | 19 |
| 4 | 24201 | 3684 | 81883 | 2 | 130 | 99 | 34 | 0 | 0 | 0 | 1 | 10 |
| 5 | 66240 | 2833 | 241753 | 1 | 55 | 99 | 74 | 2 | 6 | 0 | 0 | 10 |
| 6 | 20724 | 585 | 60309 | 1 | 32 | 129 | 43 | 0 | 4 | 0 | 0 | 2 |
| 7 | 101110 | 3051 | 375975 | 2 | 276 | 118 | 130 | 11 | 8 | 0 | 1 | 18 |
| 8 | 50553 | 1979 | 164887 | 2 | 143 | 108 | 69 | 4 | 8 | 0 | 1 | 14 |

## Genetika: inbreeding dan pemilihan pasangan

F = koefisien inbreeding (1/16 anak sepupu, 1/4 anak saudara kandung), dihitung dari silsilah 7 generasi ke atas. Kerabat dekat = F ≥ 1/8; sedarah = F ≥ 1/64 (sepupu dua kali atau lebih dekat).

| Seed | Kelahiran | F rata-rata (hidup) | Sedarah (hidup) | Kerabat dekat (hidup) | Varian dibawa /orang | Punya kelainan | Bayi wafat karena kelainan |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 3375 | 0,1403 | 99,2% | 39,5% | 1,91 | 96,1% | 0 |
| 2 | 11791 | 0,1175 | 96,6% | 33,5% | 1,29 | 98,3% | 0 |
| 3 | 10151 | 0,1143 | 93,0% | 30,5% | 1,25 | 17,4% | 0 |
| 4 | 3198 | 0,1930 | 93,0% | 62,6% | 0,68 | 2,9% | 0 |
| 5 | 4109 | 0,2645 | 96,9% | 75,0% | 0,52 | 23,8% | 0 |
| 6 | 1637 | 0,3634 | 100,0% | 95,8% | 0,23 | 0,0% | 0 |
| 7 | 7507 | 0,1681 | 94,7% | 50,8% | 1,83 | 37,9% | 0 |
| 8 | 3534 | 0,2502 | 98,4% | 79,7% | 0,30 | 0,5% | 0 |

### Menurut generasi anak (semua dunia)

| Generasi | Kelahiran | F rata-rata | Sedarah | Kerabat dekat |
| --- | ---: | ---: | ---: | ---: |
| 1 | 67 | 0,0000 | 0,0% | 0,0% |
| 2 | 95 | 0,2500 | 100,0% | 100,0% |
| 3 | 185 | 0,2868 | 100,0% | 100,0% |
| 4–5 | 583 | 0,3427 | 100,0% | 100,0% |
| 6–8 | 1327 | 0,3800 | 100,0% | 100,0% |
| 9–12 | 2104 | 0,4206 | 100,0% | 99,7% |
| 13–20 | 5416 | 0,3012 | 99,3% | 85,9% |
| 21–35 | 18883 | 0,2621 | 98,2% | 77,4% |
| 36+ | 16642 | 0,1561 | 96,4% | 49,1% |

### Nasib anak menurut F (semua dunia)

| F | Kelahiran | Lahir dengan kelainan resesif | Wafat minggu pertama karena kelainan | Wafat < 1 th | Wafat < 15 th | Wafat < 15 per kelahiran |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 (tak sekerabat) | 200 | 23 | 0 | 18 | 50 | 0,250 |
| < 1/64 (kerabat jauh) | 840 | 373 | 0 | 81 | 261 | 0,311 |
| 1/64–1/16 (sepupu dua kali) | 4067 | 1835 | 0 | 420 | 1357 | 0,334 |
| 1/16–1/8 (sepupu) | 8463 | 3820 | 0 | 889 | 2883 | 0,341 |
| 1/8–1/4 (saudara tiri, paman–keponakan) | 12979 | 5893 | 0 | 1291 | 3987 | 0,307 |
| ≥ 1/4 (saudara kandung, orang tua–anak) | 18753 | 8646 | 0 | 1810 | 5397 | 0,288 |

Anak yang masih kecil di akhir run belum sempat wafat, jadi angka per kelahiran adalah batas bawah; bandingkan antar-baris dan antar-konfigurasi, bukan dengan l15.

### Dari waktu ke waktu (semua dunia, per 25 tahun)

Keinginan kawin = bagian saat orang dewasa yang sedang melihat calon pasangan dewasa memilih *kawin*, bila calon itu orang tua, anak, atau saudara (kandung atau tiri) — kerabat yang tumbuh bersama, sasaran efek Westermarck — dan bila orang lain. r lain = rata-rata koefisien kekerabatan dengan calon lain itu.

| Tahun | Kelahiran | F rata-rata | Anak kerabat dekat | Keinginan kawin: keluarga inti | : orang lain | r lain | Sampel keluarga / lain |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1–25 | 82 | 0,0488 | 19,5% | 67,3% | 84,5% | 0,00 | 281 / 976 |
| 26–50 | 130 | 0,2683 | 99,2% | 62,7% | 56,9% | 0,38 | 2359 / 441 |
| 51–75 | 237 | 0,3155 | 100,0% | 59,8% | 57,7% | 0,44 | 3112 / 2880 |
| 76–100 | 367 | 0,3342 | 100,0% | 59,4% | 59,4% | 0,46 | 3300 / 7455 |
| 101–125 | 460 | 0,3620 | 100,0% | 67,2% | 60,8% | 0,49 | 3724 / 10787 |
| 126–150 | 538 | 0,3805 | 100,0% | 63,4% | 61,4% | 0,51 | 3614 / 11684 |
| 151–175 | 545 | 0,3922 | 100,0% | 58,4% | 60,2% | 0,53 | 3835 / 14375 |
| 176–200 | 589 | 0,4094 | 100,0% | 61,9% | 60,6% | 0,54 | 3954 / 16673 |
| 201–225 | 596 | 0,4244 | 100,0% | 60,5% | 58,2% | 0,54 | 5258 / 15118 |
| 226–250 | 642 | 0,4166 | 98,9% | 64,5% | 60,8% | 0,52 | 4336 / 16119 |
| 251–275 | 716 | 0,3829 | 97,1% | 64,3% | 60,1% | 0,47 | 4512 / 17025 |
| 276–300 | 802 | 0,3161 | 90,4% | 64,8% | 61,2% | 0,40 | 4489 / 21667 |
| 301–325 | 798 | 0,3054 | 87,9% | 63,6% | 59,8% | 0,36 | 5114 / 21658 |
| 326–350 | 829 | 0,2835 | 82,2% | 63,7% | 62,9% | 0,34 | 4657 / 20423 |
| 351–375 | 874 | 0,2742 | 84,2% | 63,6% | 62,5% | 0,34 | 5742 / 22263 |
| 376–400 | 844 | 0,2827 | 80,2% | 63,2% | 61,6% | 0,32 | 5397 / 21282 |
| 401–425 | 776 | 0,2821 | 78,8% | 62,2% | 61,5% | 0,33 | 4936 / 19354 |
| 426–450 | 939 | 0,2651 | 75,2% | 63,2% | 63,7% | 0,32 | 6170 / 24093 |
| 451–475 | 1087 | 0,2715 | 71,9% | 66,8% | 63,8% | 0,30 | 6067 / 24445 |
| 476–500 | 1089 | 0,2979 | 80,6% | 67,2% | 66,6% | 0,33 | 6954 / 22905 |
| 501–525 | 1347 | 0,3175 | 81,0% | 68,1% | 68,4% | 0,37 | 7944 / 26233 |
| 526–550 | 1384 | 0,3135 | 82,1% | 67,4% | 67,6% | 0,39 | 8691 / 31490 |
| 551–575 | 1521 | 0,2951 | 82,7% | 65,8% | 67,5% | 0,37 | 8279 / 35857 |
| 576–600 | 1844 | 0,2777 | 80,4% | 70,6% | 68,9% | 0,36 | 7669 / 42448 |
| 601–625 | 1738 | 0,2643 | 79,8% | 71,1% | 69,3% | 0,34 | 8219 / 41581 |
| 626–650 | 1705 | 0,2443 | 78,1% | 70,8% | 68,4% | 0,32 | 8644 / 40263 |
| 651–675 | 1727 | 0,2346 | 78,9% | 68,5% | 68,4% | 0,32 | 8628 / 40126 |
| 676–700 | 1967 | 0,2223 | 72,5% | 70,4% | 68,7% | 0,28 | 9732 / 42025 |
| 701–725 | 2091 | 0,2114 | 70,2% | 70,0% | 66,5% | 0,27 | 8972 / 45123 |
| 726–750 | 2236 | 0,1878 | 63,3% | 69,9% | 68,1% | 0,25 | 9548 / 50035 |
| 751–775 | 2350 | 0,1684 | 57,7% | 73,2% | 68,6% | 0,22 | 10278 / 53015 |
| 776–800 | 2504 | 0,1635 | 54,9% | 72,5% | 70,2% | 0,21 | 9697 / 54809 |
| 801–825 | 2479 | 0,1528 | 47,3% | 75,6% | 72,3% | 0,21 | 9384 / 59877 |
| 826–850 | 2487 | 0,1499 | 42,7% | 74,2% | 73,1% | 0,20 | 9205 / 56869 |
| 851–875 | 2543 | 0,1429 | 40,3% | 74,2% | 72,6% | 0,20 | 9846 / 54475 |
| 876–900 | 2437 | 0,1503 | 44,5% | 76,4% | 75,1% | 0,19 | 10535 / 53563 |
| 901–925 | 2 | 0,0243 | 0,0% | 0,0% | 65,2% | 0,19 | 42 / 293 |

### Frekuensi varian resesif di akhir run (median antar-dunia)

| Kelainan | Frekuensi alel |
| --- | ---: |
| kelainan bawaan berat | 0,0163 |
| otot dan tulang lemah | 0,0137 |
| kekebalan lemah | 0,0302 |
| kurang subur | 0,0040 |
| talasemia mayor | 0,0000 |
