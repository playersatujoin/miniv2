# Laporan soak — Fase 3c genetika aktif (900 tahun)

- Dibuat: 2026-10-07 00:11:33
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 120 m (utuh) | 1 | 42 | 43 | 1 | 5 | 5 | 26,8 | 0,60 | 26,7 | 6,5 | 3,8 | 19,4 | 26 | 0,39 | 1434 | 0,26 |
| 2 | 15 m | 2 | 48 | 39 | 1 | 3 | 8 | 27,6 | 0,73 | 21,0 | 4,2 | 4,1 | 19,0 | 18 | 0,39 | 395 | 0,33 |
| 3 | 76 m | 4 | 1 | 6 | 1 | 5 | 1 | – | – | – | – | – | – | – | – | – | 0,18 |
| 4 | 120 m (utuh) | 1 | 112 | 47 | 0 | 5 | 9 | 29,2 | 0,69 | 24,5 | 6,6 | 3,6 | 18,6 | 20 | 0,51 | 346 | 0,48 |
| 5 | 120 m (utuh) | 1 | 1 | 40 | 1 | 6 | 1 | 6,1 | 0,04 | 14,6 | – | – | – | – | – | 2533 | 0,37 |
| 6 | 14 m | 2 | 266 | 40 | 0 | 4 | 10 | 26,7 | 0,73 | 19,1 | 6,6 | 3,8 | 18,5 | 18 | 0,43 | 131 | 0,56 |
| 7 | 120 m (utuh) | 1 | 439 | 48 | 1 | 6 | 13 | 22,7 | 0,66 | 16,0 | 7,8 | 3,3 | 17,9 | 18 | 0,46 | 61 | 1,33 |
| 8 | 9 m | 2 | 136 | 44 | 1 | 4 | 10 | 24,1 | 0,70 | 16,8 | 6,8 | 3,5 | 17,7 | 22 | 0,39 | 68 | 0,42 |
| **Median** | 98 m | | 80 | 42 | 1 | 5 | 8 | 26,7 | 0,69 | 19,1 | 6,6 | 3,7 | 18,6 | 19 | 0,41 | 346 | 0,39 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **4 dari 8** dunia (selang kepercayaan 95%: 22–78%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 26,7 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,69 | 0,44–0,73 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 19,1 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 19 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 6,6 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,7 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 18,6 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,41 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,104 | 0,13–0,41 | ↓ di bawah | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,158 |
| Diare per orang per tahun | 0,17 |
| Malaria per orang per tahun | 0,66 |
| ISPA per orang per tahun | 0,01 |
| Rasio kelamin (♂ per 100 ♀) | 106 |
| Anggota per rumah | 9,4 |
| Pembunuhan per 100.000 tahun-orang | 346 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 678 | 35% | 0 | 0% |
| Kehausan | 532 | 28% | 4 | 1% |
| Usia tua | 185 | 10% | 0 | 0% |
| Dibunuh | 93 | 5% | 23 | 7% |
| Diterkam hewan | 0 | 0% | 0 | 0% |
| Melahirkan | 20 | 1% | 0 | 0% |
| Neonatal (minggu pertama) | 158 | 8% | 158 | 51% |
| Diare | 86 | 4% | 35 | 11% |
| Malaria | 171 | 9% | 83 | 27% |
| Radang paru (ISPA) | 5 | 0% | 4 | 1% |
| Terbakar | 0 | 0% | 0 | 0% |
| Tenggelam | 0 | 0% | 0 | 0% |
| Terjatuh | 0 | 0% | 0 | 0% |
| **Penyakit menular** | 262 | 14% | 122 | 40% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 4 | – | 1 | 70 | 35,0 | 0,76 |
| 2 | 12 | – | 1 | 37 | 29,1 | 0,93 |
| 3 | 6 | – | 1 | 31 | 88,0 | 0,20 |
| 4 | – | – | 0 | 33 | 35,6 | 0,89 |
| 5 | 11 | – | 1 | 67 | 31,0 | 0,51 |
| 6 | – | – | 0 | 34 | 30,9 | 0,95 |
| 7 | 5 | – | 1 | 14 | 30,3 | 0,94 |
| 8 | 7 | – | 1 | 15 | 38,1 | 0,87 |
| **Median** | 6 | – | 1 | 34 | 33,0 | 0,88 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,90 | 547 |
| 15–29 | 0,91 | 256 |
| 30–44 | 0,93 | 149 |
| 45–59 | 0,91 | 61 |
| 60+ | 0,90 | 32 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 37,9 |
| 10–19 | 36,0 |
| 20–29 | 35,7 |
| 30–39 | 34,2 |
| 40–49 | 34,2 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 84 | 0 | 9 | 71 | 0,99 | 166 / 142 / 85 / 0 / 5 | 0 | 34 / 33 |
| 2 | 165 | 0 | 2 | 140 | 0,97 | 155 / 122 / 0 / 22 / 3 | 0 | 39 / 38 |
| 3 | 57 | 0 | 7 | 56 | 1,00 | 112 / 187 / 29 / 1 / 4 | 0 | 36 / 36 |
| 4 | 158 | 0 | 6 | 98 | 0,97 | 129 / 183 / 58 / 2 / 8 | 0 | 34 / 34 |
| 5 | 210 | 0 | 4 | 103 | 1,00 | 133 / 58 / 41 / 2 / 4 | 0 | 48 / 47 |
| 6 | 266 | 0 | 19 | 229 | 0,95 | 123 / 180 / 67 / 0 / 2 | 0 | 44 / 42 |
| 7 | 533 | 0 | 3 | 204 | 0,96 | 115 / 140 / 88 / 19 / 5 | 4 | 32 / 32 |
| 8 | 161 | 0 | 21 | 105 | 0,97 | 196 / 135 / 46 / 0 / 6 | 0 | 42 / 41 |
| **Median** | 163 | 0 | 6 | 104 | 0,97 | | | 38 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 99% | 8/8 |
| Ayam Hutan | 99% | 77% | 7/8 |
| Kerbau Liar | 69% | 35% | 5/8 |
| Harimau | 50% | 44% | 8/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 96% | 3% | 0% | 1% | 0,12 |
| 2 | 77% | 23% | 0% | 0% | 0,74 |
| 3 | 99% | 0% | 1% | 0% | 0,00 |
| 4 | 64% | 36% | 0% | 0% | 0,34 |
| 5 | 93% | 7% | 0% | 0% | 0,28 |
| 6 | 51% | 49% | 0% | 0% | 0,66 |
| 7 | 28% | 72% | 0% | 0% | 0,78 |
| 8 | 57% | 40% | 4% | 0% | 0,43 |
| **Median** | | 29% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 992 | 13,4 | 14,6 | 41,0 | 60,0 |
| Netral | 3409 | 12,9 | 12,6 | 40,0 | 59,0 |
| La Niña | 1577 | 13,2 | 12,9 | 40,8 | 57,4 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 869 | ×1,22 | ×1,43 | 28% |
| La Niña | 1359 | ×1,03 | ×0,98 | 21% |

## Engine II: pertukaran, desa, api, air

| Seed | Bicara | Barter | Kabar | Desa | Warga desa | Ganti pemimpin | Kebakaran | Bangunan terbakar | Rakit dibawa | Tenggelam | Terjatuh | Stimulus aktif |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 7651 | 1042 | 23390 | 1 | 42 | 107 | 63 | 1 | 2 | 0 | 0 | 4 |
| 2 | 23004 | 1666 | 76504 | 1 | 36 | 70 | 71 | 0 | 0 | 0 | 1 | 2 |
| 3 | 5242 | 830 | 17211 | 0 | 0 | 49 | 62 | 0 | 0 | 0 | 0 | 0 |
| 4 | 33713 | 2520 | 111181 | 1 | 109 | 129 | 96 | 7 | 6 | 0 | 1 | 11 |
| 5 | 23660 | 1848 | 77054 | 0 | 0 | 113 | 143 | 15 | 0 | 0 | 0 | 0 |
| 6 | 54065 | 1669 | 179527 | 1 | 191 | 96 | 70 | 4 | 0 | 0 | 1 | 14 |
| 7 | 173095 | 4460 | 648391 | 2 | 186 | 131 | 109 | 5 | 10 | 0 | 0 | 22 |
| 8 | 42416 | 847 | 148625 | 1 | 104 | 90 | 49 | 2 | 7 | 4 | 0 | 8 |

## Genetika: inbreeding dan pemilihan pasangan

F = koefisien inbreeding (1/16 anak sepupu, 1/4 anak saudara kandung), dihitung dari silsilah 7 generasi ke atas. Kerabat dekat = F ≥ 1/8; sedarah = F ≥ 1/64 (sepupu dua kali atau lebih dekat).

| Seed | Kelahiran | F rata-rata (hidup) | Sedarah (hidup) | Kerabat dekat (hidup) | Varian dibawa /orang | Punya kelainan | Bayi wafat karena kelainan |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1269 | 0,5029 | 100,0% | 100,0% | 0,91 | 40,5% | 48 |
| 2 | 2220 | 0,2532 | 100,0% | 97,9% | 1,04 | 27,1% | 38 |
| 3 | 587 | 0,5469 | 100,0% | 100,0% | 3,00 | 100,0% | 34 |
| 4 | 2888 | 0,2473 | 100,0% | 98,2% | 0,66 | 11,6% | 103 |
| 5 | 2442 | 0,4620 | 100,0% | 100,0% | 0,00 | 100,0% | 111 |
| 6 | 4092 | 0,1575 | 97,4% | 55,6% | 0,88 | 13,5% | 43 |
| 7 | 10551 | 0,1691 | 95,2% | 54,7% | 1,73 | 54,2% | 206 |
| 8 | 3073 | 0,1422 | 96,3% | 45,6% | 1,21 | 25,0% | 144 |

### Menurut generasi anak (semua dunia)

| Generasi | Kelahiran | F rata-rata | Sedarah | Kerabat dekat |
| --- | ---: | ---: | ---: | ---: |
| 1 | 91 | 0,0000 | 0,0% | 0,0% |
| 2 | 118 | 0,2500 | 100,0% | 100,0% |
| 3 | 162 | 0,3075 | 100,0% | 100,0% |
| 4–5 | 473 | 0,3473 | 100,0% | 100,0% |
| 6–8 | 1045 | 0,3929 | 100,0% | 100,0% |
| 9–12 | 2542 | 0,4389 | 100,0% | 100,0% |
| 13–20 | 5653 | 0,2780 | 99,7% | 86,5% |
| 21–35 | 11106 | 0,1909 | 98,0% | 66,3% |
| 36+ | 5932 | 0,1706 | 96,3% | 55,8% |

### Nasib anak menurut F (semua dunia)

| F | Kelahiran | Lahir dengan kelainan resesif | Wafat minggu pertama karena kelainan | Wafat < 1 th | Wafat < 15 th | Wafat < 15 per kelahiran |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 (tak sekerabat) | 153 | 19 | 1 | 13 | 41 | 0,268 |
| < 1/64 (kerabat jauh) | 393 | 140 | 9 | 43 | 121 | 0,308 |
| 1/64–1/16 (sepupu dua kali) | 2072 | 717 | 25 | 237 | 649 | 0,313 |
| 1/16–1/8 (sepupu) | 4594 | 1602 | 51 | 463 | 1427 | 0,311 |
| 1/8–1/4 (saudara tiri, paman–keponakan) | 8327 | 2872 | 144 | 997 | 2569 | 0,309 |
| ≥ 1/4 (saudara kandung, orang tua–anak) | 11583 | 4321 | 497 | 1518 | 3391 | 0,293 |

Anak yang masih kecil di akhir run belum sempat wafat, jadi angka per kelahiran adalah batas bawah; bandingkan antar-baris dan antar-konfigurasi, bukan dengan l15.

### Dari waktu ke waktu (semua dunia, per 25 tahun)

Keinginan kawin = bagian saat orang dewasa yang sedang melihat calon pasangan dewasa memilih *kawin*, bila calon itu orang tua, anak, atau saudara (kandung atau tiri) — kerabat yang tumbuh bersama, sasaran efek Westermarck — dan bila orang lain. r lain = rata-rata koefisien kekerabatan dengan calon lain itu.

| Tahun | Kelahiran | F rata-rata | Anak kerabat dekat | Keinginan kawin: keluarga inti | : orang lain | r lain | Sampel keluarga / lain |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1–25 | 70 | 0,0286 | 11,4% | 69,2% | 74,5% | 0,00 | 201 / 1047 |
| 26–50 | 80 | 0,2446 | 95,0% | 61,9% | 45,5% | 0,18 | 2078 / 356 |
| 51–75 | 108 | 0,2812 | 96,3% | 56,4% | 58,3% | 0,43 | 1993 / 1762 |
| 76–100 | 190 | 0,3101 | 100,0% | 57,7% | 52,7% | 0,45 | 2188 / 3528 |
| 101–125 | 226 | 0,3232 | 96,0% | 51,8% | 53,9% | 0,44 | 2286 / 6244 |
| 126–150 | 276 | 0,3568 | 97,8% | 59,5% | 57,2% | 0,48 | 3002 / 7119 |
| 151–175 | 375 | 0,3838 | 100,0% | 56,7% | 58,8% | 0,53 | 3768 / 8723 |
| 176–200 | 485 | 0,3997 | 100,0% | 63,9% | 61,0% | 0,54 | 3139 / 10909 |
| 201–225 | 598 | 0,4160 | 100,0% | 61,3% | 59,3% | 0,55 | 3657 / 13015 |
| 226–250 | 738 | 0,4265 | 100,0% | 63,3% | 60,3% | 0,57 | 5159 / 15488 |
| 251–275 | 623 | 0,3985 | 99,7% | 60,5% | 60,4% | 0,53 | 4279 / 15887 |
| 276–300 | 684 | 0,3805 | 98,8% | 60,0% | 60,1% | 0,49 | 4098 / 16179 |
| 301–325 | 879 | 0,3616 | 96,5% | 62,5% | 62,0% | 0,47 | 5137 / 20575 |
| 326–350 | 827 | 0,3559 | 92,7% | 57,4% | 61,0% | 0,46 | 4827 / 21682 |
| 351–375 | 727 | 0,3196 | 87,9% | 63,5% | 60,8% | 0,42 | 4231 / 19789 |
| 376–400 | 978 | 0,2869 | 88,4% | 61,6% | 59,5% | 0,38 | 5108 / 22780 |
| 401–425 | 1176 | 0,2471 | 83,9% | 61,8% | 61,2% | 0,35 | 6225 / 26785 |
| 426–450 | 1147 | 0,2430 | 81,5% | 62,2% | 58,9% | 0,33 | 5785 / 29975 |
| 451–475 | 1010 | 0,2201 | 76,6% | 63,0% | 60,0% | 0,30 | 4783 / 28173 |
| 476–500 | 886 | 0,2091 | 77,1% | 67,1% | 62,4% | 0,28 | 4573 / 22628 |
| 501–525 | 823 | 0,1958 | 75,9% | 61,5% | 63,1% | 0,27 | 4087 / 20012 |
| 526–550 | 868 | 0,2013 | 73,9% | 63,1% | 64,7% | 0,27 | 5126 / 21987 |
| 551–575 | 893 | 0,1954 | 68,9% | 63,9% | 65,9% | 0,25 | 4640 / 21366 |
| 576–600 | 857 | 0,1800 | 66,4% | 64,5% | 65,6% | 0,24 | 4139 / 18862 |
| 601–625 | 915 | 0,1805 | 61,1% | 66,9% | 67,5% | 0,23 | 5500 / 22020 |
| 626–650 | 904 | 0,1796 | 62,5% | 66,0% | 68,1% | 0,24 | 5117 / 20442 |
| 651–675 | 959 | 0,1636 | 52,6% | 67,7% | 69,3% | 0,22 | 3928 / 21487 |
| 676–700 | 966 | 0,1653 | 53,2% | 70,0% | 70,0% | 0,22 | 4296 / 21931 |
| 701–725 | 1041 | 0,1677 | 55,2% | 69,1% | 71,5% | 0,22 | 4949 / 23540 |
| 726–750 | 1031 | 0,1722 | 56,2% | 68,2% | 70,6% | 0,22 | 4358 / 23260 |
| 751–775 | 892 | 0,1745 | 57,0% | 66,2% | 69,2% | 0,21 | 5159 / 21618 |
| 776–800 | 909 | 0,1786 | 58,8% | 64,6% | 68,8% | 0,21 | 5713 / 19257 |
| 801–825 | 960 | 0,1770 | 56,1% | 65,3% | 70,1% | 0,21 | 6466 / 21018 |
| 826–850 | 980 | 0,1671 | 50,3% | 68,0% | 71,9% | 0,20 | 6371 / 20254 |
| 851–875 | 994 | 0,1747 | 59,1% | 66,9% | 68,7% | 0,23 | 5240 / 22233 |
| 876–900 | 1046 | 0,1852 | 61,9% | 68,7% | 67,5% | 0,23 | 6162 / 22663 |
| 901–925 | 1 | 0,1616 | 100,0% | 0,0% | 52,4% | 0,22 | 35 / 124 |

### Frekuensi varian resesif di akhir run (median antar-dunia)

| Kelainan | Frekuensi alel |
| --- | ---: |
| kelainan bawaan berat | 0,0040 |
| otot dan tulang lemah | 0,0588 |
| kekebalan lemah | 0,0333 |
| kurang subur | 0,0321 |
| talasemia mayor | 0,0000 |
