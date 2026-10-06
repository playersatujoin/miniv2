# Laporan soak — v13: gizi + genetika + berbagi pangan

- Dibuat: 2026-10-07 00:34:55
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 20 menit simulasi per dunia (≈ 150 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 20 m (utuh) | 1 | 52 | 8 | 0 | 1 | 2 | 30,6 | 0,65 | 31,3 | 6,2 | 3,6 | 19,4 | 20 | 0,25 | 52 | 0,22 |
| 2 | 20 m (utuh) | 1 | 14 | 7 | 0 | 1 | 4 | 53,6 | 1,00 | 38,6 | 2,0 | 3,4 | – | – | 0,37 | 0 | 0,18 |
| 3 | 20 m (utuh) | 1 | 43 | 7 | 0 | 2 | 2 | 39,6 | 0,77 | 35,9 | 6,5 | 3,9 | 19,5 | 60 | 0,37 | 702 | 0,20 |
| 4 | 20 m (utuh) | 1 | 106 | 9 | 1 | 4 | 3 | 36,3 | 0,70 | 35,3 | 5,4 | 3,9 | 18,9 | 24 | 0,41 | 152 | 0,36 |
| 5 | 20 m (utuh) | 1 | 120 | 7 | 1 | 2 | 3 | 40,3 | 0,85 | 31,9 | 5,1 | 4,2 | 19,6 | 28 | 0,45 | 49 | 0,29 |
| 6 | 20 m (utuh) | 1 | 33 | 8 | 1 | 3 | 1 | 26,8 | 0,67 | 24,1 | 3,5 | 3,1 | 18,8 | 26 | 0,39 | 147 | 0,20 |
| 7 | 20 m (utuh) | 1 | 88 | 8 | 1 | 1 | 5 | 34,6 | 0,73 | 32,1 | 7,0 | 2,9 | 18,7 | 26 | 0,40 | 27 | 0,29 |
| 8 | 10 m | 2 | 35 | 5 | 0 | 3 | 3 | 45,8 | 0,85 | 38,7 | 7,7 | 4,0 | 17,1 | – | 0,23 | 0 | 0,15 |
| **Median** | 20 m | | 48 | 8 | 0 | 2 | 3 | 38,0 | 0,75 | 33,7 | 5,8 | 3,7 | 18,9 | 26 | 0,38 | 50 | 0,21 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **7 dari 8** dunia (selang kepercayaan 95%: 53–98%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 38,0 tahun | 21–37 | ↑ di atas | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,75 | 0,44–0,73 | ↑ di atas | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 33,7 tahun | 28–43 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 26 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 5,8 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,7 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 18,9 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,38 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,155 | 0,13–0,41 | ✓ dalam rentang | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,203 |
| Diare per orang per tahun | 0,30 |
| Malaria per orang per tahun | 0,74 |
| ISPA per orang per tahun | 0,00 |
| Rasio kelamin (♂ per 100 ♀) | 127 |
| Anggota per rumah | 18,5 |
| Pembunuhan per 100.000 tahun-orang | 50 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 49 | 10% | 0 | 0% |
| Kehausan | 105 | 22% | 0 | 0% |
| Usia tua | 119 | 25% | 0 | 0% |
| Dibunuh | 24 | 5% | 4 | 3% |
| Diterkam hewan | 4 | 1% | 0 | 0% |
| Melahirkan | 7 | 1% | 0 | 0% |
| Neonatal (minggu pertama) | 100 | 21% | 100 | 70% |
| Diare | 30 | 6% | 17 | 12% |
| Malaria | 34 | 7% | 21 | 15% |
| Radang paru (ISPA) | 1 | 0% | 0 | 0% |
| Terbakar | 0 | 0% | 0 | 0% |
| Tenggelam | 0 | 0% | 0 | 0% |
| Terjatuh | 0 | 0% | 0 | 0% |
| **Penyakit menular** | 65 | 14% | 38 | 27% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | – | – | 0 | 9 | 32,5 | 0,57 |
| 2 | – | – | 0 | 6 | 40,6 | 0,79 |
| 3 | – | – | 0 | 5 | 38,2 | 0,96 |
| 4 | 5 | – | 1 | 6 | 39,6 | 0,86 |
| 5 | 3 | – | 1 | 1 | 34,0 | 0,84 |
| 6 | 6 | – | 1 | 15 | 35,5 | 0,76 |
| 7 | 4 | – | 1 | 3 | 34,7 | 0,81 |
| 8 | – | – | 0 | 4 | 33,4 | 0,51 |
| **Median** | 4 | – | 0 | 6 | 35,1 | 0,80 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,80 | 213 |
| 15–29 | 0,81 | 136 |
| 30–44 | 0,83 | 69 |
| 45–59 | 0,87 | 44 |
| 60+ | 0,87 | 29 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 35,3 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 52 | 0 | 7 | 16 | 0,98 | 87 / 167 / 83 / 27 / 0 | 0 | 4 / 3 |
| 2 | 20 | 0 | 5 | 44 | 0,99 | 91 / 213 / 55 / 26 / 3 | 5 | 2 / 2 |
| 3 | 43 | 0 | 7 | 49 | 0,98 | 95 / 111 / 65 / 53 / 3 | 0 | 3 / 3 |
| 4 | 106 | 0 | 8 | 26 | 0,97 | 113 / 117 / 115 / 8 / 3 | 0 | 4 / 4 |
| 5 | 120 | 0 | 4 | 68 | 0,97 | 134 / 91 / 71 / 0 / 0 | 0 | 8 / 6 |
| 6 | 33 | 0 | 12 | 14 | 0,99 | 122 / 176 / 6 / 9 / 2 | 0 | 3 / 3 |
| 7 | 88 | 0 | 11 | 42 | 0,97 | 108 / 142 / 18 / 30 / 0 | 4 | 6 / 5 |
| 8 | 35 | 0 | 9 | 4 | 0,99 | 114 / 139 / 229 / 3 / 0 | 0 | 6 / 5 |
| **Median** | 48 | 0 | 8 | 34 | 0,98 | | | 4 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 100% | 8/8 |
| Ayam Hutan | 100% | 100% | 8/8 |
| Kerbau Liar | 100% | 55% | 7/8 |
| Harimau | 50% | 40% | 4/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 99% | 0% | 0% | 1% | 0,07 |
| 2 | 90% | 10% | 0% | 0% | 0,95 |
| 3 | 66% | 33% | 0% | 0% | 1,21 |
| 4 | 86% | 14% | 0% | 1% | 0,26 |
| 5 | 77% | 22% | 0% | 1% | 0,51 |
| 6 | 93% | 4% | 2% | 1% | 0,16 |
| 7 | 87% | 11% | 0% | 1% | 0,23 |
| 8 | 97% | 1% | 2% | 0% | 0,01 |
| **Median** | | 10% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 102 | 2,7 | 4,1 | 42,5 | 35,4 |
| Netral | 355 | 3,2 | 3,0 | 37,1 | 37,3 |
| La Niña | 160 | 2,5 | 2,2 | 37,6 | 41,5 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 54 | ×1,00 | ×1,56 | 13% |
| La Niña | 79 | ×0,89 | ×0,83 | 15% |

## Engine II: pertukaran, desa, api, air

| Seed | Bicara | Barter | Kabar | Desa | Warga desa | Ganti pemimpin | Kebakaran | Bangunan terbakar | Rakit dibawa | Tenggelam | Terjatuh | Stimulus aktif |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 849 | 127 | 2720 | 1 | 52 | 6 | 9 | 0 | 0 | 0 | 0 | 1 |
| 2 | 514 | 66 | 1408 | 1 | 7 | 7 | 10 | 0 | 0 | 0 | 0 | 0 |
| 3 | 1104 | 186 | 3775 | 1 | 36 | 6 | 7 | 0 | 5 | 0 | 0 | 2 |
| 4 | 2247 | 401 | 7581 | 1 | 105 | 9 | 18 | 0 | 2 | 0 | 0 | 7 |
| 5 | 2789 | 218 | 9109 | 1 | 119 | 8 | 11 | 1 | 0 | 0 | 0 | 4 |
| 6 | 788 | 167 | 2384 | 0 | 0 | 9 | 6 | 0 | 0 | 0 | 0 | 0 |
| 7 | 3022 | 173 | 9728 | 1 | 7 | 11 | 6 | 0 | 0 | 0 | 0 | 5 |
| 8 | 432 | 45 | 1140 | 1 | 24 | 2 | 9 | 0 | 0 | 0 | 0 | 1 |

## Gizi (Fase 3d)

Bagian penduduk, dirata-rata atas setiap menit simulasi. Stunting: tinggi badan sekitar 2 SD di bawah anak sebaya yang cukup gizi.

| Seed | Kurang protein/mikro | Berat | Gizi kurang/buruk | Kurus | Kurang protein | Kurang mikro | Balita pendek | Anak <15 pendek | Dewasa pendek | Ibu hamil/menyusui kurang | Diet protein | Diet mikro |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 9% | 2% | 9% | 1% | 8% | 6% | 0% | 0% | 0% | 33% | 1,14 | 1,12 |
| 2 | 9% | 2% | 9% | 0% | 8% | 4% | 0% | 0% | 0% | 12% | 1,14 | 1,13 |
| 3 | 19% | 6% | 19% | 1% | 16% | 14% | 2% | 6% | 9% | 41% | 1,08 | 1,09 |
| 4 | 16% | 3% | 17% | 2% | 14% | 11% | 0% | 4% | 4% | 31% | 1,08 | 1,07 |
| 5 | 21% | 6% | 22% | 2% | 17% | 14% | 2% | 8% | 8% | 52% | 1,08 | 1,04 |
| 6 | 4% | 0% | 5% | 2% | 3% | 1% | 0% | 0% | 0% | 6% | 1,34 | 1,23 |
| 7 | 13% | 3% | 15% | 2% | 12% | 9% | 0% | 2% | 3% | 39% | 1,08 | 1,08 |
| 8 | 1% | 0% | 1% | 1% | 0% | 1% | 0% | 0% | 0% | 6% | 1,39 | 1,23 |
| **Median** | 11% | 3% | 12% | | | | 0% | 1% | 1% | | | |

Acuan: stunting balita di dunia 26% pada 2011, 30–50% di Asia Selatan dan Afrika sub-Sahara (Black dkk. 2013, Lancet 382:427); petani awal lebih pendek dan lebih sering kurang gizi daripada pemburu-peramu (Cohen & Armelagos 1984).

## Genetika: inbreeding dan pemilihan pasangan

F = koefisien inbreeding (1/16 anak sepupu, 1/4 anak saudara kandung), dihitung dari silsilah 7 generasi ke atas. Kerabat dekat = F ≥ 1/8; sedarah = F ≥ 1/64 (sepupu dua kali atau lebih dekat).

| Seed | Kelahiran | F rata-rata (hidup) | Sedarah (hidup) | Kerabat dekat (hidup) | Varian dibawa /orang | Punya kelainan | Bayi wafat karena kelainan |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 156 | 0,3809 | 100,0% | 100,0% | 2,08 | 13,5% | 34 |
| 2 | 49 | 0,4337 | 100,0% | 100,0% | 0,07 | 0,0% | 2 |
| 3 | 105 | 0,3995 | 100,0% | 100,0% | 1,33 | 25,6% | 10 |
| 4 | 282 | 0,3369 | 100,0% | 100,0% | 1,02 | 4,7% | 2 |
| 5 | 251 | 0,3692 | 100,0% | 100,0% | 0,08 | 0,0% | 0 |
| 6 | 126 | 0,4360 | 100,0% | 100,0% | 0,76 | 9,1% | 24 |
| 7 | 264 | 0,3558 | 100,0% | 100,0% | 1,82 | 9,1% | 40 |
| 8 | 50 | 0,3565 | 97,1% | 97,1% | 0,20 | 2,9% | 0 |

### Menurut generasi anak (semua dunia)

| Generasi | Kelahiran | F rata-rata | Sedarah | Kerabat dekat |
| --- | ---: | ---: | ---: | ---: |
| 1 | 78 | 0,0000 | 0,0% | 0,0% |
| 2 | 114 | 0,2500 | 100,0% | 100,0% |
| 3 | 134 | 0,2952 | 100,0% | 100,0% |
| 4–5 | 434 | 0,3546 | 100,0% | 100,0% |
| 6–8 | 522 | 0,3683 | 100,0% | 100,0% |
| 9–12 | 1 | 0,3339 | 100,0% | 100,0% |

### Nasib anak menurut F (semua dunia)

| F | Kelahiran | Lahir dengan kelainan resesif | Wafat minggu pertama karena kelainan | Wafat < 1 th | Wafat < 15 th | Wafat < 15 per kelahiran |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 (tak sekerabat) | 78 | 1 | 1 | 7 | 12 | 0,154 |
| < 1/64 (kerabat jauh) | 0 | 0 | 0 | 0 | 0 | – |
| 1/64–1/16 (sepupu dua kali) | 0 | 0 | 0 | 0 | 0 | – |
| 1/16–1/8 (sepupu) | 0 | 0 | 0 | 0 | 0 | – |
| 1/8–1/4 (saudara tiri, paman–keponakan) | 5 | 0 | 0 | 0 | 1 | 0,200 |
| ≥ 1/4 (saudara kandung, orang tua–anak) | 1200 | 229 | 111 | 223 | 312 | 0,260 |

Anak yang masih kecil di akhir run belum sempat wafat, jadi angka per kelahiran adalah batas bawah; bandingkan antar-baris dan antar-konfigurasi, bukan dengan l15.

### Dari waktu ke waktu (semua dunia, per 25 tahun)

Keinginan kawin = bagian saat orang dewasa yang sedang melihat calon pasangan dewasa memilih *kawin*, bila calon itu orang tua, anak, atau saudara (kandung atau tiri) — kerabat yang tumbuh bersama, sasaran efek Westermarck — dan bila orang lain. r lain = rata-rata koefisien kekerabatan dengan calon lain itu.

| Tahun | Kelahiran | F rata-rata | Anak kerabat dekat | Keinginan kawin: keluarga inti | : orang lain | r lain | Sampel keluarga / lain |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1–25 | 78 | 0,0385 | 15,4% | 57,3% | 80,1% | 0,00 | 239 / 738 |
| 26–50 | 130 | 0,2399 | 93,1% | 71,3% | 69,9% | 0,31 | 2299 / 519 |
| 51–75 | 167 | 0,2985 | 98,8% | 57,9% | 63,2% | 0,42 | 3238 / 2108 |
| 76–100 | 219 | 0,3413 | 99,5% | 60,1% | 56,1% | 0,45 | 2678 / 4293 |
| 101–125 | 286 | 0,3520 | 100,0% | 57,8% | 54,8% | 0,47 | 3370 / 7239 |
| 126–150 | 402 | 0,3850 | 100,0% | 56,3% | 56,7% | 0,50 | 4186 / 8231 |
| 151–175 | 1 | 0,3339 | 100,0% | 0,0% | 23,7% | 0,52 | 24 / 55 |

### Frekuensi varian resesif di akhir run (median antar-dunia)

| Kelainan | Frekuensi alel |
| --- | ---: |
| kelainan bawaan berat | 0,0188 |
| otot dan tulang lemah | 0,0000 |
| kekebalan lemah | 0,0186 |
| kurang subur | 0,0000 |
| talasemia mayor | 0,0000 |
