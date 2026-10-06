# Laporan soak — f3b-nosan

- Dibuat: 2026-10-06 17:09:51
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: sanitation
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 116 m | 2 | 6 | 2 | 0 | 1 | 1 | 27,8 | 0,59 | 27,7 | – | 2,8 | – | – | 0,34 | 0 | 0,42 |
| 2 | 120 m (utuh) | 1 | 80 | 42 | 0 | 3 | 6 | 28,5 | 0,72 | 22,2 | 7,5 | 2,9 | 18,1 | 18 | 0,40 | 64 | 0,78 |
| 3 | 65 m | 2 | 49 | 20 | 0 | 1 | 3 | 24,4 | 0,70 | 17,4 | 8,7 | 3,7 | 17,1 | 18 | 0,57 | 319 | 0,72 |
| 4 | 120 m (utuh) | 1 | 126 | 44 | 0 | 1 | 1 | 24,6 | 0,74 | 16,4 | 6,1 | 3,7 | 17,7 | 18 | 0,45 | 68 | 0,60 |
| 5 | 120 m (utuh) | 1 | 465 | 44 | 0 | 3 | 5 | 22,9 | 0,65 | 17,5 | 8,6 | 3,0 | 17,2 | 18 | 0,53 | 119 | 2,37 |
| 6 | 42 m | 7 | 5 | 1 | 0 | 2 | 1 | – | – | – | 0,9 | – | – | – | 0,04 | – | 0,29 |
| 7 | 120 m (utuh) | 1 | 16 | 39 | 0 | 1 | 2 | 27,5 | 0,82 | 17,4 | 5,0 | 4,6 | 17,6 | 18 | 0,69 | 82 | 0,46 |
| 8 | 29 m | 5 | 12 | 5 | 0 | 2 | 2 | 20,8 | 0,59 | 15,3 | 7,7 | 3,2 | 18,1 | – | 0,37 | 1189 | 0,32 |
| 9 | 120 m (utuh) | 1 | 23 | 41 | 0 | 1 | 3 | 20,1 | 0,57 | 15,3 | 7,5 | 3,1 | 17,2 | 18 | 0,31 | 93 | 0,54 |
| 10 | 13 m | 9 | 28 | 7 | 0 | 1 | 1 | 40,6 | 0,83 | 32,9 | 2,7 | 3,3 | 21,3 | – | 0,28 | 0 | 0,28 |
| 11 | 120 m (utuh) | 1 | 137 | 44 | 0 | 3 | 3 | 21,6 | 0,57 | 18,8 | 7,9 | 3,2 | 17,9 | 18 | 0,43 | 538 | 1,20 |
| 12 | 120 m (utuh) | 1 | 99 | 41 | 0 | 3 | 8 | 20,4 | 0,52 | 17,4 | 8,0 | 3,5 | 17,8 | 18 | 0,40 | 147 | 0,80 |
| 13 | 120 m (utuh) | 1 | 118 | 43 | 1 | 3 | 1 | 24,2 | 0,73 | 16,1 | 6,8 | 3,4 | 18,0 | 18 | 0,49 | 143 | 0,94 |
| 14 | 40 m | 6 | 52 | 10 | 0 | 4 | 1 | 25,9 | 0,70 | 19,2 | 4,5 | 4,2 | 19,3 | 22 | 0,39 | 163 | 0,38 |
| 15 | 27 m | 6 | 11 | 4 | 0 | 2 | 1 | 20,4 | 0,46 | 18,2 | 6,7 | 2,9 | 19,8 | – | 0,42 | 712 | 0,31 |
| 16 | 20 m | 4 | 66 | 20 | 0 | 5 | 2 | 25,7 | 0,70 | 19,5 | 5,6 | 3,3 | 17,3 | 18 | 0,49 | 111 | 0,59 |
| **Median** | 118 m | | 50 | 30 | 0 | 2 | 2 | 24,4 | 0,70 | 17,5 | 6,8 | 3,3 | 17,9 | 18 | 0,41 | 119 | 0,56 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **8 dari 16** dunia (selang kepercayaan 95%: 28–72%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 24,4 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,70 | 0,44–0,73 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 17,5 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 18 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 6,8 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,3 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 17,9 tahun | 18–20 | ↓ di bawah | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,41 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,096 | 0,13–0,41 | ↓ di bawah | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,170 |
| Diare per orang per tahun | 0,65 |
| Malaria per orang per tahun | 0,64 |
| ISPA per orang per tahun | 0,01 |
| Rasio kelamin (♂ per 100 ♀) | 107 |
| Anggota per rumah | 12,1 |
| Pembunuhan per 100.000 tahun-orang | 119 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 959 | 36% | 0 | 0% |
| Kehausan | 588 | 22% | 9 | 2% |
| Usia tua | 233 | 9% | 0 | 0% |
| Dibunuh | 118 | 4% | 30 | 6% |
| Diterkam hewan | 0 | 0% | 0 | 0% |
| Melahirkan | 48 | 2% | 0 | 0% |
| Neonatal (minggu pertama) | 178 | 7% | 178 | 33% |
| Diare | 310 | 12% | 164 | 30% |
| Malaria | 226 | 8% | 142 | 26% |
| Radang paru (ISPA) | 26 | 1% | 22 | 4% |
| **Penyakit menular** | 562 | 21% | 328 | 60% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | – | – | 0 | 5 | 31,7 | 0,00 |
| 2 | – | – | 0 | 17 | 31,6 | 0,88 |
| 3 | – | – | 0 | 25 | 25,2 | 0,94 |
| 4 | – | – | 0 | 8 | 25,7 | 0,95 |
| 5 | – | – | 0 | 11 | 25,1 | 0,96 |
| 6 | – | – | 0 | 18 | 21,0 | 0,00 |
| 7 | – | – | 0 | 16 | 28,3 | 0,51 |
| 8 | – | – | 0 | 16 | 30,5 | 0,06 |
| 9 | – | – | 0 | 14 | 34,1 | 0,64 |
| 10 | – | – | 0 | 33 | 29,9 | 0,56 |
| 11 | – | – | 0 | 13 | 21,6 | 0,93 |
| 12 | – | – | 0 | 12 | 30,2 | 0,87 |
| 13 | 17 | – | 1 | 15 | 25,3 | 0,95 |
| 14 | – | – | 0 | 16 | 29,3 | 0,93 |
| 15 | – | – | 0 | 25 | 35,9 | 0,47 |
| 16 | – | – | 0 | 27 | 27,5 | 0,95 |
| **Median** | 17 | – | 0 | 16 | 28,8 | 0,87 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,88 | 653 |
| 15–29 | 0,89 | 363 |
| 30–44 | 0,88 | 134 |
| 45–59 | 0,88 | 100 |
| 60+ | 0,91 | 43 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 31,0 |
| 10–19 | 30,2 |
| 20–29 | 28,1 |
| 30–39 | 27,5 |
| 40–49 | 25,8 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 29 | 0 | 15 | 15 | 1,00 | 193 / 161 / 31 / 0 / 0 | 0 | 45 / 43 |
| 2 | 82 | 0 | 7 | 39 | 0,98 | 98 / 196 / 74 / 6 / 3 | 0 | 30 / 30 |
| 3 | 96 | 0 | 11 | 43 | 0,99 | 155 / 149 / 27 / 0 / 0 | 0 | 42 / 40 |
| 4 | 126 | 0 | 4 | 80 | 0,98 | 81 / 165 / 7 / 35 / 0 | 0 | 37 / 36 |
| 5 | 477 | 0 | 11 | 152 | 0,97 | 95 / 171 / 28 / 36 / 4 | 0 | 27 / 27 |
| 6 | 21 | 0 | 60 | 4 | 1,00 | 105 / 165 / 64 / 26 / 0 | 0 | 26 / 25 |
| 7 | 65 | 0 | 14 | 33 | 0,99 | 85 / 166 / 38 / 20 / 0 | 0 | 54 / 53 |
| 8 | 23 | 0 | 5 | 17 | 0,99 | 132 / 213 / 41 / 2 / 0 | 0 | 41 / 39 |
| 9 | 58 | 0 | 23 | 27 | 0,99 | 159 / 158 / 0 / 3 / 0 | 0 | 51 / 49 |
| 10 | 28 | 0 | 8 | 12 | 0,99 | 152 / 143 / 28 / 29 / 0 | 0 | 32 / 31 |
| 11 | 185 | 0 | 29 | 100 | 0,98 | 151 / 182 / 59 / 2 / 3 | 0 | 43 / 43 |
| 12 | 157 | 0 | 8 | 57 | 0,97 | 117 / 8 / 45 / 13 / 0 | 0 | 43 / 42 |
| 13 | 149 | 0 | 12 | 55 | 0,98 | 189 / 164 / 14 / 3 / 0 | 0 | 44 / 43 |
| 14 | 84 | 0 | 27 | 74 | 1,00 | 141 / 140 / 49 / 4 / 0 | 0 | 48 / 47 |
| 15 | 36 | 0 | 29 | 41 | 0,99 | 154 / 176 / 55 / 3 / 3 | 0 | 41 / 41 |
| 16 | 105 | 0 | 1 | 63 | 0,99 | 129 / 184 / 22 / 0 / 5 | 0 | 48 / 47 |
| **Median** | 83 | 0 | 12 | 42 | 0,99 | | | 42 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 16/16 |
| Babi Hutan | 100% | 100% | 16/16 |
| Ayam Hutan | 92% | 65% | 15/16 |
| Kerbau Liar | 78% | 57% | 13/16 |
| Harimau | 52% | 41% | 5/16 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 100% | 0% | 0% | 0% | 0,00 |
| 2 | 65% | 34% | 0% | 1% | 0,30 |
| 3 | 37% | 63% | 0% | 0% | 0,83 |
| 4 | 36% | 63% | 0% | 1% | 0,98 |
| 5 | 16% | 84% | 0% | 0% | 0,93 |
| 6 | 100% | 0% | 0% | 0% | 0,00 |
| 7 | 100% | 0% | 0% | 0% | 0,00 |
| 8 | 89% | 0% | 11% | 0% | 0,00 |
| 9 | 60% | 40% | 0% | 0% | 0,21 |
| 10 | 94% | 6% | 0% | 0% | 0,12 |
| 11 | 38% | 62% | 0% | 0% | 0,84 |
| 12 | 46% | 54% | 0% | 0% | 0,38 |
| 13 | 43% | 56% | 0% | 0% | 0,65 |
| 14 | 62% | 37% | 0% | 2% | 0,82 |
| 15 | 89% | 0% | 6% | 4% | 0,01 |
| 16 | 46% | 54% | 0% | 0% | 1,08 |
| **Median** | | 39% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 1128 | 15,1 | 16,5 | 43,1 | 67,9 |
| Netral | 4106 | 15,1 | 15,1 | 42,5 | 66,2 |
| La Niña | 1983 | 15,0 | 14,6 | 42,7 | 64,5 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 1034 | ×1,08 | ×1,22 | 27% |
| La Niña | 1782 | ×1,11 | ×1,10 | 23% |
