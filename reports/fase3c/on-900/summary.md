# Laporan soak — Fase 3c genetika aktif (900 tahun)

- Dibuat: 2026-10-07 00:22:49
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 120 m (utuh) | 1 | 104 | 43 | 1 | 6 | 4 | 27,8 | 0,71 | 22,5 | 6,5 | 3,8 | 18,9 | 18 | 0,44 | 478 | 0,38 |
| 2 | 15 m | 4 | 64 | 16 | 0 | 4 | 6 | 22,8 | 0,61 | 18,9 | 4,1 | 4,0 | 20,4 | 26 | 0,36 | 42 | 0,16 |
| 3 | 120 m (utuh) | 1 | 593 | 47 | 1 | 5 | 18 | 23,5 | 0,64 | 18,2 | 8,1 | 3,2 | 17,7 | 18 | 0,47 | 166 | 1,96 |
| 4 | 120 m (utuh) | 1 | 153 | 45 | 0 | 5 | 13 | 24,6 | 0,61 | 20,7 | 6,9 | 3,4 | 18,2 | 18 | 0,48 | 301 | 0,71 |
| 5 | 120 m (utuh) | 1 | 42 | 45 | 0 | 5 | 4 | 35,4 | 0,75 | 29,9 | 4,0 | 4,1 | 20,6 | 58 | 0,44 | 247 | 0,41 |
| 6 | 120 m (utuh) | 1 | 76 | 41 | 1 | 8 | 6 | 26,5 | 0,68 | 22,0 | 6,7 | 3,6 | 18,6 | 22 | 0,23 | 484 | 0,29 |
| 7 | 120 m (utuh) | 1 | 226 | 47 | 1 | 6 | 15 | 30,6 | 0,77 | 22,9 | 5,6 | 4,3 | 18,3 | 18 | 0,43 | 92 | 0,77 |
| 8 | 9 m | 2 | 177 | 44 | 1 | 4 | 6 | 24,9 | 0,65 | 20,6 | 7,8 | 3,0 | 17,9 | 18 | 0,53 | 35 | 0,47 |
| **Median** | 120 m | | 128 | 44 | 1 | 5 | 6 | 25,7 | 0,67 | 21,4 | 6,6 | 3,7 | 18,5 | 18 | 0,44 | 206 | 0,44 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **6 dari 8** dunia (selang kepercayaan 95%: 41–93%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 25,7 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,67 | 0,44–0,73 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 21,4 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 18 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 6,6 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,7 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 18,5 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,44 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,113 | 0,13–0,41 | ↓ di bawah | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,180 |
| Diare per orang per tahun | 0,18 |
| Malaria per orang per tahun | 1,01 |
| ISPA per orang per tahun | 0,01 |
| Rasio kelamin (♂ per 100 ♀) | 112 |
| Anggota per rumah | 11,0 |
| Pembunuhan per 100.000 tahun-orang | 206 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 872 | 35% | 2 | 0% |
| Kehausan | 610 | 25% | 0 | 0% |
| Usia tua | 262 | 11% | 0 | 0% |
| Dibunuh | 119 | 5% | 51 | 11% |
| Diterkam hewan | 1 | 0% | 0 | 0% |
| Melahirkan | 46 | 2% | 0 | 0% |
| Neonatal (minggu pertama) | 227 | 9% | 227 | 51% |
| Diare | 111 | 5% | 61 | 14% |
| Malaria | 210 | 9% | 102 | 23% |
| Radang paru (ISPA) | 5 | 0% | 3 | 1% |
| Terbakar | 0 | 0% | 0 | 0% |
| Tenggelam | 0 | 0% | 0 | 0% |
| Terjatuh | 1 | 0% | 0 | 0% |
| **Penyakit menular** | 326 | 13% | 166 | 37% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 7 | – | 1 | 28 | 29,3 | 0,85 |
| 2 | – | – | 0 | 18 | 37,8 | 0,88 |
| 3 | 6 | – | 1 | 55 | 29,5 | 0,95 |
| 4 | – | – | 0 | 34 | 42,5 | 0,88 |
| 5 | – | – | 0 | 56 | 45,5 | 0,85 |
| 6 | 6 | – | 1 | 20 | 37,5 | 0,89 |
| 7 | 4 | – | 1 | 47 | 27,1 | 0,88 |
| 8 | 14 | – | 1 | 50 | 31,0 | 0,92 |
| **Median** | 6 | – | 1 | 40 | 34,2 | 0,88 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,89 | 690 |
| 15–29 | 0,92 | 410 |
| 30–44 | 0,92 | 160 |
| 45–59 | 0,89 | 119 |
| 60+ | 0,88 | 56 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 35,2 |
| 10–19 | 35,0 |
| 20–29 | 33,6 |
| 30–39 | 33,6 |
| 40–49 | 34,1 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 133 | 0 | 7 | 58 | 0,96 | 116 / 123 / 0 / 6 / 5 | 0 | 34 / 33 |
| 2 | 74 | 0 | 2 | 63 | 0,98 | 92 / 204 / 24 / 15 / 2 | 0 | 28 / 27 |
| 3 | 802 | 0 | 7 | 250 | 0,94 | 122 / 146 / 26 / 28 / 4 | 0 | 29 / 29 |
| 4 | 348 | 0 | 15 | 136 | 0,95 | 110 / 161 / 12 / 18 / 0 | 0 | 41 / 40 |
| 5 | 148 | 0 | 4 | 120 | 0,98 | 111 / 214 / 36 / 9 / 0 | 0 | 38 / 37 |
| 6 | 76 | 0 | 14 | 52 | 0,99 | 121 / 177 / 99 / 0 / 5 | 0 | 43 / 42 |
| 7 | 314 | 0 | 3 | 252 | 0,95 | 148 / 169 / 26 / 0 / 0 | 4 | 52 / 50 |
| 8 | 214 | 0 | 18 | 53 | 0,96 | 140 / 155 / 42 / 0 / 0 | 0 | 39 / 37 |
| **Median** | 181 | 0 | 7 | 92 | 0,96 | | | 38 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 100% | 8/8 |
| Ayam Hutan | 98% | 74% | 7/8 |
| Kerbau Liar | 72% | 42% | 5/8 |
| Harimau | 52% | 36% | 4/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 78% | 22% | 0% | 0% | 0,19 |
| 2 | 55% | 40% | 5% | 0% | 0,89 |
| 3 | 19% | 81% | 0% | 0% | 1,03 |
| 4 | 65% | 35% | 0% | 0% | 0,32 |
| 5 | 79% | 21% | 0% | 0% | 0,39 |
| 6 | 84% | 12% | 4% | 0% | 0,40 |
| 7 | 66% | 33% | 1% | 0% | 0,62 |
| 8 | 42% | 54% | 4% | 0% | 0,35 |
| **Median** | | 34% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 1084 | 14,6 | 17,2 | 40,7 | 69,4 |
| Netral | 3488 | 14,1 | 13,6 | 41,1 | 71,6 |
| La Niña | 1614 | 14,8 | 14,2 | 42,0 | 72,6 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 1004 | ×1,21 | ×1,33 | 29% |
| La Niña | 1470 | ×1,14 | ×1,10 | 24% |

## Engine II: pertukaran, desa, api, air

| Seed | Bicara | Barter | Kabar | Desa | Warga desa | Ganti pemimpin | Kebakaran | Bangunan terbakar | Rakit dibawa | Tenggelam | Terjatuh | Stimulus aktif |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 21365 | 1964 | 68729 | 1 | 101 | 112 | 93 | 1 | 5 | 0 | 0 | 6 |
| 2 | 7792 | 630 | 26693 | 1 | 32 | 41 | 40 | 0 | 0 | 0 | 0 | 3 |
| 3 | 164113 | 6363 | 612978 | 1 | 392 | 208 | 100 | 5 | 13 | 0 | 0 | 29 |
| 4 | 54753 | 3774 | 193980 | 2 | 84 | 86 | 103 | 9 | 13 | 0 | 0 | 7 |
| 5 | 25420 | 2338 | 89557 | 1 | 36 | 109 | 87 | 2 | 3 | 0 | 0 | 3 |
| 6 | 13437 | 1033 | 41032 | 1 | 76 | 98 | 88 | 8 | 3 | 0 | 3 | 6 |
| 7 | 78130 | 2188 | 279423 | 1 | 128 | 116 | 111 | 10 | 7 | 0 | 1 | 9 |
| 8 | 49365 | 1465 | 177058 | 1 | 167 | 73 | 80 | 12 | 11 | 0 | 2 | 9 |

## Genetika: inbreeding dan pemilihan pasangan

F = koefisien inbreeding (1/16 anak sepupu, 1/4 anak saudara kandung), dihitung dari silsilah 7 generasi ke atas. Kerabat dekat = F ≥ 1/8; sedarah = F ≥ 1/64 (sepupu dua kali atau lebih dekat).

| Seed | Kelahiran | F rata-rata (hidup) | Sedarah (hidup) | Kerabat dekat (hidup) | Varian dibawa /orang | Punya kelainan | Bayi wafat karena kelainan |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 2321 | 0,1612 | 94,2% | 59,6% | 0,48 | 0,0% | 54 |
| 2 | 635 | 0,4253 | 100,0% | 100,0% | 0,53 | 6,3% | 16 |
| 3 | 14114 | 0,1437 | 90,4% | 44,5% | 1,59 | 22,4% | 157 |
| 4 | 4796 | 0,1933 | 96,7% | 66,0% | 0,26 | 1,3% | 9 |
| 5 | 2746 | 0,2497 | 100,0% | 88,1% | 0,76 | 7,1% | 49 |
| 6 | 1779 | 0,2870 | 100,0% | 86,8% | 0,97 | 2,6% | 142 |
| 7 | 5408 | 0,1698 | 99,1% | 49,1% | 0,89 | 14,2% | 65 |
| 8 | 3528 | 0,1501 | 97,7% | 66,7% | 1,38 | 10,7% | 149 |

### Menurut generasi anak (semua dunia)

| Generasi | Kelahiran | F rata-rata | Sedarah | Kerabat dekat |
| --- | ---: | ---: | ---: | ---: |
| 1 | 86 | 0,0000 | 0,0% | 0,0% |
| 2 | 123 | 0,2500 | 100,0% | 100,0% |
| 3 | 182 | 0,2919 | 100,0% | 100,0% |
| 4–5 | 632 | 0,3577 | 100,0% | 100,0% |
| 6–8 | 1292 | 0,4096 | 100,0% | 100,0% |
| 9–12 | 2182 | 0,4375 | 100,0% | 99,9% |
| 13–20 | 5552 | 0,3028 | 99,8% | 88,0% |
| 21–35 | 15403 | 0,1929 | 97,8% | 61,7% |
| 36+ | 9875 | 0,1697 | 95,1% | 55,1% |

### Nasib anak menurut F (semua dunia)

| F | Kelahiran | Lahir dengan kelainan resesif | Wafat minggu pertama karena kelainan | Wafat < 1 th | Wafat < 15 th | Wafat < 15 per kelahiran |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 (tak sekerabat) | 214 | 13 | 1 | 21 | 73 | 0,341 |
| < 1/64 (kerabat jauh) | 704 | 90 | 1 | 74 | 260 | 0,369 |
| 1/64–1/16 (sepupu dua kali) | 4412 | 610 | 17 | 436 | 1450 | 0,329 |
| 1/16–1/8 (sepupu) | 5754 | 915 | 43 | 614 | 1864 | 0,324 |
| 1/8–1/4 (saudara tiri, paman–keponakan) | 9672 | 1423 | 129 | 1061 | 3168 | 0,328 |
| ≥ 1/4 (saudara kandung, orang tua–anak) | 14571 | 2281 | 450 | 1831 | 4341 | 0,298 |

Anak yang masih kecil di akhir run belum sempat wafat, jadi angka per kelahiran adalah batas bawah; bandingkan antar-baris dan antar-konfigurasi, bukan dengan l15.

### Dari waktu ke waktu (semua dunia, per 25 tahun)

Keinginan kawin = bagian saat orang dewasa yang sedang melihat calon pasangan dewasa memilih *kawin*, bila calon itu orang tua, anak, atau saudara (kandung atau tiri) — kerabat yang tumbuh bersama, sasaran efek Westermarck — dan bila orang lain. r lain = rata-rata koefisien kekerabatan dengan calon lain itu.

| Tahun | Kelahiran | F rata-rata | Anak kerabat dekat | Keinginan kawin: keluarga inti | : orang lain | r lain | Sampel keluarga / lain |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1–25 | 66 | 0,0227 | 9,1% | 66,7% | 76,0% | 0,00 | 162 / 1043 |
| 26–50 | 88 | 0,2479 | 96,6% | 64,9% | 55,2% | 0,21 | 1814 / 272 |
| 51–75 | 135 | 0,2849 | 97,0% | 52,0% | 56,6% | 0,42 | 2089 / 1654 |
| 76–100 | 225 | 0,3333 | 100,0% | 61,7% | 52,9% | 0,44 | 2181 / 4209 |
| 101–125 | 294 | 0,3560 | 98,3% | 55,9% | 54,1% | 0,48 | 3091 / 7002 |
| 126–150 | 412 | 0,3750 | 99,3% | 56,2% | 56,2% | 0,50 | 3460 / 8887 |
| 151–175 | 497 | 0,3972 | 100,0% | 58,6% | 57,0% | 0,52 | 4327 / 12950 |
| 176–200 | 509 | 0,4193 | 100,0% | 56,2% | 54,6% | 0,54 | 4495 / 11867 |
| 201–225 | 546 | 0,4187 | 100,0% | 53,3% | 53,9% | 0,56 | 4012 / 13230 |
| 226–250 | 655 | 0,4253 | 100,0% | 55,0% | 53,9% | 0,56 | 4755 / 16156 |
| 251–275 | 660 | 0,4164 | 99,7% | 57,6% | 56,9% | 0,54 | 4946 / 18790 |
| 276–300 | 737 | 0,3840 | 96,8% | 60,5% | 59,8% | 0,48 | 4271 / 18539 |
| 301–325 | 811 | 0,3495 | 91,7% | 62,2% | 62,0% | 0,42 | 4935 / 18944 |
| 326–350 | 861 | 0,3192 | 87,2% | 61,0% | 61,6% | 0,41 | 4855 / 21022 |
| 351–375 | 830 | 0,2868 | 85,3% | 61,0% | 60,2% | 0,35 | 4544 / 20165 |
| 376–400 | 904 | 0,2688 | 81,5% | 64,1% | 58,7% | 0,35 | 5285 / 21528 |
| 401–425 | 974 | 0,2369 | 76,0% | 58,4% | 58,1% | 0,31 | 4837 / 25417 |
| 426–450 | 1103 | 0,2170 | 68,8% | 58,0% | 58,0% | 0,29 | 5328 / 28663 |
| 451–475 | 821 | 0,2308 | 76,0% | 62,7% | 60,5% | 0,31 | 3710 / 21404 |
| 476–500 | 808 | 0,2315 | 70,3% | 61,4% | 62,2% | 0,29 | 4434 / 20448 |
| 501–525 | 1202 | 0,2219 | 68,5% | 64,3% | 61,3% | 0,29 | 6302 / 27603 |
| 526–550 | 1384 | 0,1948 | 64,2% | 63,2% | 61,7% | 0,26 | 6390 / 33843 |
| 551–575 | 1515 | 0,1741 | 57,9% | 63,7% | 60,9% | 0,24 | 6899 / 34776 |
| 576–600 | 1440 | 0,1839 | 60,9% | 61,7% | 60,1% | 0,24 | 6581 / 34341 |
| 601–625 | 1418 | 0,1808 | 57,6% | 64,0% | 62,2% | 0,23 | 6307 / 31481 |
| 626–650 | 1583 | 0,1838 | 57,3% | 65,5% | 61,7% | 0,22 | 7983 / 33232 |
| 651–675 | 1396 | 0,1843 | 58,0% | 68,1% | 64,0% | 0,23 | 6076 / 31668 |
| 676–700 | 1584 | 0,1837 | 58,0% | 65,4% | 63,3% | 0,24 | 7721 / 35642 |
| 701–725 | 1452 | 0,1838 | 60,6% | 65,0% | 63,1% | 0,24 | 6321 / 32704 |
| 726–750 | 1626 | 0,1953 | 63,0% | 69,3% | 65,9% | 0,24 | 6578 / 35299 |
| 751–775 | 1589 | 0,1824 | 56,6% | 68,6% | 66,1% | 0,25 | 5882 / 37078 |
| 776–800 | 1601 | 0,1914 | 57,3% | 71,1% | 66,4% | 0,24 | 7310 / 35254 |
| 801–825 | 1689 | 0,1928 | 58,3% | 70,1% | 69,2% | 0,25 | 7878 / 37741 |
| 826–850 | 1213 | 0,1980 | 62,7% | 67,4% | 68,7% | 0,26 | 6031 / 29806 |
| 851–875 | 1276 | 0,1674 | 56,3% | 72,4% | 69,5% | 0,23 | 6691 / 30177 |
| 876–900 | 1422 | 0,1708 | 55,2% | 72,5% | 69,8% | 0,22 | 6707 / 31891 |
| 901–925 | 1 | 0,0833 | 0,0% | 27,5% | 62,5% | 0,23 | 51 / 205 |

### Frekuensi varian resesif di akhir run (median antar-dunia)

| Kelainan | Frekuensi alel |
| --- | ---: |
| kelainan bawaan berat | 0,0074 |
| otot dan tulang lemah | 0,0066 |
| kekebalan lemah | 0,0058 |
| kurang subur | 0,0202 |
| talasemia mayor | 0,0004 |
