# Laporan soak — Fase 3c genetika aktif (150 tahun)

- Dibuat: 2026-10-07 00:06:50
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 20 menit simulasi per dunia (≈ 150 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 20 m (utuh) | 1 | 29 | 7 | 1 | 3 | 6 | 43,1 | 0,82 | 36,7 | 3,7 | 4,9 | 25,4 | 18 | 0,34 | 152 | 0,18 |
| 2 | 15 m | 2 | 9 | 2 | 0 | 1 | 1 | 39,3 | 0,91 | 28,3 | – | 2,7 | – | – | 0,26 | 0 | 0,10 |
| 3 | 20 m (utuh) | 1 | 28 | 8 | 1 | 1 | 3 | 25,8 | 0,62 | 24,0 | 4,1 | 6,2 | 19,2 | 18 | 0,42 | 201 | 0,21 |
| 4 | 20 m (utuh) | 1 | 103 | 8 | 0 | 3 | 7 | 31,4 | 0,74 | 26,2 | 5,6 | 4,1 | 18,6 | 20 | 0,34 | 50 | 0,27 |
| 5 | 20 m (utuh) | 1 | 72 | 7 | 0 | 1 | 4 | 36,1 | 0,65 | 39,3 | 6,1 | 3,5 | 20,2 | 70 | 0,43 | 38 | 0,20 |
| 6 | 14 m | 2 | 12 | 2 | 0 | 3 | 2 | 61,6 | 1,00 | 46,6 | 5,4 | 4,0 | – | – | 0,39 | 0 | 0,11 |
| 7 | 20 m (utuh) | 1 | 49 | 8 | 1 | 6 | 7 | 34,7 | 0,78 | 28,7 | 4,6 | 4,5 | 20,2 | 18 | 0,33 | 161 | 0,22 |
| 8 | 9 m | 2 | 39 | 5 | 0 | 1 | 4 | 33,9 | 0,69 | 34,2 | 9,1 | 2,7 | 18,1 | – | 0,37 | 0 | 0,12 |
| **Median** | 20 m | | 34 | 7 | 0 | 2 | 4 | 35,4 | 0,76 | 31,5 | 5,4 | 4,1 | 19,7 | 18 | 0,36 | 44 | 0,19 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **5 dari 8** dunia (selang kepercayaan 95%: 31–86%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 35,4 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,76 | 0,44–0,73 | ↑ di atas | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 31,5 tahun | 28–43 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 18 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 5,4 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 4,1 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 19,7 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,36 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,136 | 0,13–0,41 | ✓ dalam rentang | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,181 |
| Diare per orang per tahun | 0,36 |
| Malaria per orang per tahun | 0,60 |
| ISPA per orang per tahun | 0,00 |
| Rasio kelamin (♂ per 100 ♀) | 112 |
| Anggota per rumah | 6,2 |
| Pembunuhan per 100.000 tahun-orang | 44 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 66 | 16% | 0 | 0% |
| Kehausan | 95 | 24% | 0 | 0% |
| Usia tua | 89 | 22% | 0 | 0% |
| Dibunuh | 13 | 3% | 3 | 3% |
| Diterkam hewan | 2 | 0% | 0 | 0% |
| Melahirkan | 5 | 1% | 0 | 0% |
| Neonatal (minggu pertama) | 69 | 17% | 69 | 64% |
| Diare | 22 | 6% | 12 | 11% |
| Malaria | 39 | 10% | 24 | 22% |
| Radang paru (ISPA) | 0 | 0% | 0 | 0% |
| Terbakar | 0 | 0% | 0 | 0% |
| Tenggelam | 0 | 0% | 0 | 0% |
| Terjatuh | 0 | 0% | 0 | 0% |
| **Penyakit menular** | 61 | 15% | 36 | 33% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 4 | – | 1 | 6 | 39,9 | 0,51 |
| 2 | – | – | 0 | 3 | 34,6 | 0,41 |
| 3 | 6 | – | 1 | 3 | 47,0 | 0,62 |
| 4 | – | – | 0 | 7 | 37,9 | 0,92 |
| 5 | – | – | 0 | 2 | 39,6 | 0,92 |
| 6 | – | – | 0 | 6 | 38,3 | 0,66 |
| 7 | 5 | – | 1 | 1 | 35,8 | 0,89 |
| 8 | – | – | 0 | 3 | 32,2 | 0,52 |
| **Median** | 5 | – | 0 | 3 | 38,1 | 0,64 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,59 | 142 |
| 15–29 | 0,62 | 85 |
| 30–44 | 0,68 | 63 |
| 45–59 | 0,61 | 28 |
| 60+ | 0,74 | 23 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 36,0 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 36 | 0 | 9 | 3 | 0,99 | 92 / 211 / 64 / 32 / 0 | 0 | 5 / 4 |
| 2 | 10 | 0 | 2 | 12 | 0,99 | 115 / 163 / 68 / 9 / 0 | 0 | 5 / 4 |
| 3 | 57 | 0 | 7 | 25 | 0,98 | 130 / 207 / 120 / 12 / 0 | 0 | 3 / 2 |
| 4 | 103 | 0 | 6 | 50 | 0,97 | 118 / 189 / 39 / 0 / 1 | 0 | 7 / 6 |
| 5 | 72 | 0 | 4 | 54 | 0,98 | 167 / 45 / 7 / 0 / 0 | 0 | 6 / 4 |
| 6 | 12 | 0 | 19 | 8 | 0,99 | 144 / 181 / 89 / 2 / 0 | 0 | 8 / 6 |
| 7 | 59 | 0 | 3 | 42 | 0,99 | 116 / 184 / 83 / 22 / 0 | 5 | 4 / 3 |
| 8 | 39 | 0 | – | 0 | 0,99 | 112 / 115 / 77 / 0 / 0 | 0 | 7 / 5 |
| **Median** | 48 | 0 | 6 | 18 | 0,99 | | | 6 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 100% | 8/8 |
| Ayam Hutan | 100% | 100% | 8/8 |
| Kerbau Liar | 98% | 40% | 5/8 |
| Harimau | 57% | 35% | 1/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 100% | 0% | 0% | 0% | 0,01 |
| 2 | 100% | 0% | 0% | 0% | 0,16 |
| 3 | 98% | 2% | 0% | 0% | 0,07 |
| 4 | 74% | 26% | 0% | 0% | 0,47 |
| 5 | 62% | 37% | 0% | 0% | 0,93 |
| 6 | 97% | 2% | 0% | 1% | 0,11 |
| 7 | 75% | 23% | 0% | 2% | 0,84 |
| 8 | 98% | 0% | 1% | 1% | 0,00 |
| **Median** | | 2% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 78 | 4,0 | 2,3 | 34,1 | 58,1 |
| Netral | 308 | 3,9 | 3,3 | 34,0 | 43,4 |
| La Niña | 160 | 3,1 | 4,9 | 39,4 | 45,5 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 34 | ×1,02 | ×0,79 | 15% |
| La Niña | 86 | ×0,69 | ×1,43 | 20% |

## Engine II: pertukaran, desa, api, air

| Seed | Bicara | Barter | Kabar | Desa | Warga desa | Ganti pemimpin | Kebakaran | Bangunan terbakar | Rakit dibawa | Tenggelam | Terjatuh | Stimulus aktif |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 758 | 81 | 2478 | 1 | 20 | 8 | 11 | 0 | 0 | 0 | 0 | 1 |
| 2 | 216 | 7 | 565 | 0 | 0 | 0 | 9 | 0 | 0 | 0 | 0 | 0 |
| 3 | 1397 | 122 | 4771 | 1 | 23 | 13 | 8 | 0 | 1 | 0 | 0 | 1 |
| 4 | 2042 | 354 | 7470 | 1 | 102 | 8 | 12 | 1 | 0 | 0 | 0 | 5 |
| 5 | 1724 | 100 | 6347 | 1 | 66 | 10 | 11 | 0 | 0 | 0 | 0 | 0 |
| 6 | 226 | 12 | 283 | 1 | 8 | 2 | 2 | 1 | 0 | 0 | 0 | 0 |
| 7 | 2284 | 185 | 8093 | 1 | 29 | 10 | 4 | 0 | 0 | 0 | 0 | 1 |
| 8 | 998 | 25 | 3105 | 1 | 19 | 3 | 5 | 0 | 0 | 0 | 0 | 0 |

## Genetika: inbreeding dan pemilihan pasangan

F = koefisien inbreeding (1/16 anak sepupu, 1/4 anak saudara kandung), dihitung dari silsilah 7 generasi ke atas. Kerabat dekat = F ≥ 1/8; sedarah = F ≥ 1/64 (sepupu dua kali atau lebih dekat).

| Seed | Kelahiran | F rata-rata (hidup) | Sedarah (hidup) | Kerabat dekat (hidup) | Varian dibawa /orang | Punya kelainan | Bayi wafat karena kelainan |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 107 | 0,3202 | 100,0% | 100,0% | 1,72 | 10,3% | 14 |
| 2 | 16 | 0,0556 | 22,2% | 22,2% | 3,67 | 33,3% | 0 |
| 3 | 143 | 0,3344 | 100,0% | 100,0% | 1,71 | 32,1% | 8 |
| 4 | 269 | 0,3779 | 100,0% | 100,0% | 1,86 | 29,1% | 19 |
| 5 | 175 | 0,3338 | 100,0% | 100,0% | 1,82 | 18,1% | 25 |
| 6 | 23 | 0,1042 | 41,7% | 41,7% | 1,75 | 8,3% | 4 |
| 7 | 141 | 0,3246 | 100,0% | 100,0% | 2,08 | 44,9% | 3 |
| 8 | 76 | 0,3646 | 100,0% | 100,0% | 3,18 | 33,3% | 10 |

### Menurut generasi anak (semua dunia)

| Generasi | Kelahiran | F rata-rata | Sedarah | Kerabat dekat |
| --- | ---: | ---: | ---: | ---: |
| 1 | 85 | 0,0000 | 0,0% | 0,0% |
| 2 | 107 | 0,2500 | 100,0% | 100,0% |
| 3 | 124 | 0,3044 | 100,0% | 100,0% |
| 4–5 | 317 | 0,3221 | 100,0% | 100,0% |
| 6–8 | 317 | 0,3653 | 100,0% | 100,0% |

### Nasib anak menurut F (semua dunia)

| F | Kelahiran | Lahir dengan kelainan resesif | Wafat minggu pertama karena kelainan | Wafat < 1 th | Wafat < 15 th | Wafat < 15 per kelahiran |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 (tak sekerabat) | 85 | 8 | 1 | 8 | 17 | 0,200 |
| < 1/64 (kerabat jauh) | 0 | 0 | 0 | 0 | 0 | – |
| 1/64–1/16 (sepupu dua kali) | 0 | 0 | 0 | 0 | 0 | – |
| 1/16–1/8 (sepupu) | 0 | 0 | 0 | 0 | 0 | – |
| 1/8–1/4 (saudara tiri, paman–keponakan) | 1 | 0 | 0 | 0 | 0 | 0,000 |
| ≥ 1/4 (saudara kandung, orang tua–anak) | 864 | 296 | 82 | 150 | 217 | 0,251 |

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
| 151–175 | 0 | – | – | 0,0% | 0,0% | 0,49 | 24 / 29 |

### Frekuensi varian resesif di akhir run (median antar-dunia)

| Kelainan | Frekuensi alel |
| --- | ---: |
| kelainan bawaan berat | 0,0290 |
| otot dan tulang lemah | 0,0476 |
| kekebalan lemah | 0,0697 |
| kurang subur | 0,0268 |
| talasemia mayor | 0,0000 |
