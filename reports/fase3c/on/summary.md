# Laporan soak — Fase 3c genetika aktif (150 tahun)

- Dibuat: 2026-10-07 00:17:25
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 20 menit simulasi per dunia (≈ 150 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 20 m (utuh) | 1 | 67 | 8 | 1 | 6 | 5 | 32,9 | 0,78 | 26,2 | 5,6 | 4,2 | 19,0 | 22 | 0,44 | 0 | 0,25 |
| 2 | 15 m | 2 | 13 | 2 | 0 | 1 | 1 | 45,0 | 0,78 | 42,5 | 9,2 | 2,7 | – | – | 0,23 | 0 | 0,10 |
| 3 | 20 m (utuh) | 1 | 82 | 8 | 1 | 1 | 7 | 37,7 | 0,80 | 31,8 | 5,8 | 4,0 | 18,6 | 20 | 0,35 | 0 | 0,21 |
| 4 | 20 m (utuh) | 1 | 41 | 7 | 0 | 1 | 7 | 36,8 | 0,70 | 35,5 | 3,1 | 5,0 | 21,5 | 16 | 0,32 | 232 | 0,21 |
| 5 | 20 m (utuh) | 1 | 140 | 8 | 0 | 4 | 4 | 33,9 | 0,77 | 27,3 | 7,2 | 3,5 | 17,0 | 18 | 0,34 | 39 | 0,31 |
| 6 | 20 m (utuh) | 1 | 60 | 7 | 1 | 5 | 4 | 27,5 | 0,62 | 28,5 | 6,0 | 3,1 | 18,3 | 24 | 0,36 | 0 | 0,19 |
| 7 | 20 m (utuh) | 1 | 42 | 7 | 1 | 2 | 5 | 40,8 | 0,87 | 31,8 | 3,8 | 3,6 | 19,5 | 22 | 0,35 | 0 | 0,17 |
| 8 | 9 m | 2 | 45 | 5 | 0 | 3 | 6 | 37,1 | 0,73 | 34,6 | 7,2 | 3,2 | 19,2 | – | 0,37 | 0 | 0,12 |
| **Median** | 20 m | | 52 | 7 | 0 | 2 | 5 | 37,0 | 0,78 | 31,8 | 5,9 | 3,5 | 19,0 | 21 | 0,35 | 0 | 0,20 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **6 dari 8** dunia (selang kepercayaan 95%: 41–93%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 37,0 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,78 | 0,44–0,73 | ↑ di atas | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 31,8 tahun | 28–43 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 21 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 5,9 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,5 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 19,0 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,35 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,135 | 0,13–0,41 | ✓ dalam rentang | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,200 |
| Diare per orang per tahun | 0,35 |
| Malaria per orang per tahun | 0,63 |
| ISPA per orang per tahun | 0,00 |
| Rasio kelamin (♂ per 100 ♀) | 115 |
| Anggota per rumah | 11,9 |
| Pembunuhan per 100.000 tahun-orang | 0 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 74 | 15% | 0 | 0% |
| Kehausan | 144 | 29% | 0 | 0% |
| Usia tua | 110 | 22% | 0 | 0% |
| Dibunuh | 7 | 1% | 2 | 1% |
| Diterkam hewan | 0 | 0% | 0 | 0% |
| Melahirkan | 6 | 1% | 0 | 0% |
| Neonatal (minggu pertama) | 69 | 14% | 69 | 51% |
| Diare | 44 | 9% | 34 | 25% |
| Malaria | 41 | 8% | 29 | 21% |
| Radang paru (ISPA) | 1 | 0% | 1 | 1% |
| Terbakar | 0 | 0% | 0 | 0% |
| Tenggelam | 0 | 0% | 0 | 0% |
| Terjatuh | 0 | 0% | 0 | 0% |
| **Penyakit menular** | 86 | 17% | 64 | 47% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 7 | – | 1 | 1 | 39,0 | 0,92 |
| 2 | – | – | 0 | 3 | 34,5 | 0,41 |
| 3 | 6 | – | 1 | 2 | 33,3 | 0,95 |
| 4 | – | – | 0 | 5 | 45,6 | 0,52 |
| 5 | – | – | 0 | 3 | 34,5 | 0,94 |
| 6 | 6 | – | 1 | 1 | 37,0 | 0,88 |
| 7 | 4 | – | 1 | 2 | 33,3 | 0,88 |
| 8 | – | – | 0 | 4 | 32,8 | 0,52 |
| **Median** | 6 | – | 0 | 2 | 34,5 | 0,88 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,84 | 215 |
| 15–29 | 0,87 | 122 |
| 30–44 | 0,89 | 82 |
| 45–59 | 0,90 | 50 |
| 60+ | 0,88 | 21 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 35,3 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 72 | 0 | 7 | 38 | 0,98 | 62 / 166 / 138 / 34 / 0 | 0 | 3 / 2 |
| 2 | 13 | 0 | 2 | 12 | 0,99 | 125 / 188 / 64 / 15 / 4 | 0 | 5 / 5 |
| 3 | 82 | 0 | 7 | 66 | 0,98 | 145 / 118 / 83 / 1 / 0 | 0 | 5 / 4 |
| 4 | 50 | 0 | 15 | 24 | 0,98 | 152 / 146 / 70 / 0 / 3 | 0 | 5 / 4 |
| 5 | 141 | 0 | 4 | 73 | 0,97 | 108 / 106 / 63 / 0 / 4 | 0 | 9 / 8 |
| 6 | 60 | 0 | 14 | 42 | 0,98 | 100 / 167 / 82 / 0 / 5 | 1 | 8 / 7 |
| 7 | 42 | 0 | 3 | 28 | 0,99 | 145 / 137 / 43 / 0 / 0 | 0 | 6 / 4 |
| 8 | 45 | 0 | 18 | 1 | 0,99 | 163 / 127 / 46 / 2 / 5 | 0 | 8 / 8 |
| **Median** | 55 | 0 | 7 | 33 | 0,98 | | | 6 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 100% | 8/8 |
| Ayam Hutan | 100% | 100% | 8/8 |
| Kerbau Liar | 82% | 35% | 4/8 |
| Harimau | 60% | 30% | 5/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 78% | 21% | 0% | 0% | 0,54 |
| 2 | 100% | 0% | 0% | 0% | 0,15 |
| 3 | 59% | 40% | 0% | 0% | 1,43 |
| 4 | 97% | 2% | 0% | 1% | 0,04 |
| 5 | 46% | 53% | 0% | 1% | 0,92 |
| 6 | 72% | 24% | 4% | 0% | 0,46 |
| 7 | 86% | 14% | 0% | 0% | 1,16 |
| 8 | 99% | 0% | 1% | 0% | 0,00 |
| **Median** | | 18% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 95 | 5,7 | 5,1 | 38,6 | 67,3 |
| Netral | 334 | 3,3 | 3,8 | 38,1 | 75,2 |
| La Niña | 146 | 3,0 | 2,3 | 37,9 | 67,5 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 44 | ×2,54 | ×1,58 | 26% |
| La Niña | 64 | ×0,75 | ×1,12 | 15% |

## Engine II: pertukaran, desa, api, air

| Seed | Bicara | Barter | Kabar | Desa | Warga desa | Ganti pemimpin | Kebakaran | Bangunan terbakar | Rakit dibawa | Tenggelam | Terjatuh | Stimulus aktif |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1750 | 217 | 6397 | 1 | 65 | 9 | 8 | 0 | 0 | 0 | 0 | 2 |
| 2 | 211 | 7 | 563 | 0 | 0 | 0 | 9 | 0 | 0 | 0 | 0 | 0 |
| 3 | 1906 | 146 | 7294 | 1 | 75 | 8 | 12 | 0 | 0 | 0 | 0 | 2 |
| 4 | 570 | 387 | 1687 | 1 | 36 | 7 | 11 | 0 | 2 | 0 | 0 | 0 |
| 5 | 3847 | 338 | 14808 | 1 | 140 | 8 | 13 | 0 | 1 | 0 | 0 | 4 |
| 6 | 1342 | 123 | 4448 | 1 | 60 | 10 | 5 | 0 | 1 | 0 | 0 | 4 |
| 7 | 2030 | 75 | 7257 | 1 | 11 | 10 | 6 | 0 | 4 | 0 | 0 | 1 |
| 8 | 1037 | 34 | 3168 | 1 | 35 | 2 | 5 | 0 | 1 | 0 | 0 | 3 |

## Genetika: inbreeding dan pemilihan pasangan

F = koefisien inbreeding (1/16 anak sepupu, 1/4 anak saudara kandung), dihitung dari silsilah 7 generasi ke atas. Kerabat dekat = F ≥ 1/8; sedarah = F ≥ 1/64 (sepupu dua kali atau lebih dekat).

| Seed | Kelahiran | F rata-rata (hidup) | Sedarah (hidup) | Kerabat dekat (hidup) | Varian dibawa /orang | Punya kelainan | Bayi wafat karena kelainan |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 216 | 0,3591 | 100,0% | 100,0% | 1,64 | 23,9% | 19 |
| 2 | 19 | 0,0769 | 30,8% | 30,8% | 2,46 | 38,5% | 1 |
| 3 | 180 | 0,3707 | 100,0% | 100,0% | 1,56 | 7,3% | 15 |
| 4 | 140 | 0,3056 | 100,0% | 100,0% | 0,58 | 7,3% | 5 |
| 5 | 321 | 0,3404 | 100,0% | 100,0% | 0,19 | 0,0% | 1 |
| 6 | 168 | 0,4522 | 100,0% | 100,0% | 1,50 | 3,3% | 26 |
| 7 | 99 | 0,4019 | 100,0% | 100,0% | 1,24 | 26,2% | 0 |
| 8 | 77 | 0,3507 | 97,8% | 97,8% | 1,44 | 24,4% | 3 |

### Menurut generasi anak (semua dunia)

| Generasi | Kelahiran | F rata-rata | Sedarah | Kerabat dekat |
| --- | ---: | ---: | ---: | ---: |
| 1 | 75 | 0,0000 | 0,0% | 0,0% |
| 2 | 102 | 0,2500 | 100,0% | 100,0% |
| 3 | 139 | 0,2954 | 100,0% | 100,0% |
| 4–5 | 479 | 0,3572 | 100,0% | 100,0% |
| 6–8 | 425 | 0,3723 | 100,0% | 100,0% |

### Nasib anak menurut F (semua dunia)

| F | Kelahiran | Lahir dengan kelainan resesif | Wafat minggu pertama karena kelainan | Wafat < 1 th | Wafat < 15 th | Wafat < 15 per kelahiran |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 (tak sekerabat) | 75 | 4 | 1 | 9 | 19 | 0,253 |
| < 1/64 (kerabat jauh) | 0 | 0 | 0 | 0 | 0 | – |
| 1/64–1/16 (sepupu dua kali) | 0 | 0 | 0 | 0 | 0 | – |
| 1/16–1/8 (sepupu) | 0 | 0 | 0 | 0 | 0 | – |
| 1/8–1/4 (saudara tiri, paman–keponakan) | 9 | 0 | 0 | 0 | 1 | 0,111 |
| ≥ 1/4 (saudara kandung, orang tua–anak) | 1136 | 197 | 69 | 165 | 255 | 0,224 |

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
| 151–175 | 0 | – | – | 0,0% | 21,7% | 0,53 | 22 / 60 |

### Frekuensi varian resesif di akhir run (median antar-dunia)

| Kelainan | Frekuensi alel |
| --- | ---: |
| kelainan bawaan berat | 0,0187 |
| otot dan tulang lemah | 0,0113 |
| kekebalan lemah | 0,0150 |
| kurang subur | 0,0046 |
| talasemia mayor | 0,0000 |
