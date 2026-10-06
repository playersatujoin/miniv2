# Laporan soak — f3b-off

- Dibuat: 2026-10-06 17:09:12
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: disease
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 47 m | 6 | 27 | 10 | 0 | 1 | 3 | 29,0 | 0,66 | 23,7 | 5,0 | 4,3 | 22,1 | 22 | 0,21 | 653 | 0,36 |
| 2 | 120 m (utuh) | 1 | 52 | 44 | 0 | 1 | 2 | 24,3 | 0,77 | 13,6 | 5,7 | 3,3 | 16,5 | 16 | 0,45 | 506 | 0,71 |
| 3 | 120 m (utuh) | 1 | 95 | 44 | 0 | 3 | 3 | 23,6 | 0,73 | 13,2 | 7,0 | 3,4 | 17,4 | 18 | 0,54 | 299 | 1,24 |
| 4 | 120 m (utuh) | 1 | 27 | 44 | 0 | 1 | 2 | 36,2 | 0,81 | 27,0 | 9,0 | 3,1 | 17,0 | – | 0,32 | 476 | 0,82 |
| 5 | 120 m (utuh) | 1 | 59 | 41 | 1 | 1 | 3 | 31,9 | 0,84 | 21,1 | 5,8 | 4,4 | 18,5 | 18 | 0,47 | 155 | 0,83 |
| 6 | 31 m | 4 | 36 | 22 | 0 | 2 | 1 | 33,0 | 0,74 | 25,2 | 4,6 | 4,2 | 17,6 | 16 | 0,49 | 0 | 0,38 |
| 7 | 9 m | 5 | 17 | 10 | 0 | 2 | 4 | 28,4 | 0,79 | 18,1 | 5,3 | 3,7 | 19,5 | 18 | 0,49 | 472 | 0,31 |
| 8 | 78 m | 2 | 20 | 15 | 0 | 1 | 2 | 29,8 | 0,56 | 29,2 | 4,4 | 3,7 | 23,1 | – | 0,46 | 1157 | 0,39 |
| 9 | 120 m (utuh) | 1 | 30 | 43 | 0 | 3 | 3 | 24,8 | 0,72 | 16,6 | 5,8 | 3,9 | 17,5 | 16 | 0,38 | 743 | 0,73 |
| 10 | 11 m | 4 | 27 | 29 | 0 | 2 | 2 | 24,9 | 0,56 | 20,7 | 7,0 | 3,9 | 17,0 | 18 | 0,30 | 118 | 0,31 |
| 11 | 60 m | 5 | 12 | 4 | 0 | 2 | 2 | 23,2 | 0,42 | 24,5 | 5,0 | 4,8 | 20,1 | – | 0,37 | 0 | 0,36 |
| 12 | 24 m | 4 | 24 | 7 | 0 | 3 | 3 | 26,8 | 0,57 | 23,8 | 5,6 | 4,6 | 18,4 | 16 | 0,59 | 0 | 0,42 |
| 13 | 120 m (utuh) | 1 | 235 | 47 | 1 | 3 | 8 | 28,0 | 0,74 | 19,2 | 8,8 | 3,0 | 17,3 | 18 | 0,61 | 49 | 1,67 |
| 14 | 67 m | 2 | 25 | 19 | 1 | 5 | 2 | 36,5 | 0,91 | 23,9 | 6,0 | 5,2 | 22,5 | 16 | 0,29 | 0 | 0,53 |
| 15 | 43 m | 6 | 50 | 5 | 1 | 4 | 2 | 30,8 | 0,77 | 21,8 | 7,4 | 2,9 | 17,4 | 18 | 0,38 | 51 | 0,30 |
| 16 | 54 m | 2 | 16 | 23 | 0 | 3 | 2 | 27,5 | 0,72 | 18,8 | 6,7 | 2,8 | 18,8 | 18 | 0,47 | 323 | 0,47 |
| **Median** | 64 m | | 27 | 22 | 0 | 2 | 2 | 28,2 | 0,73 | 21,5 | 5,8 | 3,8 | 18,0 | 18 | 0,45 | 227 | 0,45 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **6 dari 16** dunia (selang kepercayaan 95%: 18–61%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 28,2 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,73 | 0,44–0,73 | ↑ di atas | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 21,5 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 18 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 5,8 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,8 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 18,0 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,45 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,000 | 0,13–0,41 | ↓ di bawah | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,005 |
| Diare per orang per tahun | 0,00 |
| Malaria per orang per tahun | 0,00 |
| ISPA per orang per tahun | 0,00 |
| Rasio kelamin (♂ per 100 ♀) | 101 |
| Anggota per rumah | 9,0 |
| Pembunuhan per 100.000 tahun-orang | 227 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 497 | 41% | 0 | 0% |
| Kehausan | 445 | 37% | 0 | 0% |
| Usia tua | 172 | 14% | 0 | 0% |
| Dibunuh | 84 | 7% | 21 | 100% |
| Diterkam hewan | 0 | 0% | 0 | 0% |
| Melahirkan | 17 | 1% | 0 | 0% |
| Neonatal (minggu pertama) | 0 | 0% | 0 | 0% |
| Diare | 0 | 0% | 0 | 0% |
| Malaria | 0 | 0% | 0 | 0% |
| Radang paru (ISPA) | 0 | 0% | 0 | 0% |
| **Penyakit menular** | 0 | 0% | 0 | 0% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | – | – | 0 | 20 | 31,3 | 0,61 |
| 2 | – | – | 0 | 21 | 27,1 | 0,93 |
| 3 | – | – | 0 | 30 | 21,9 | 0,95 |
| 4 | – | – | 0 | 8 | 22,9 | 0,75 |
| 5 | 13 | – | 1 | 18 | 18,3 | 0,78 |
| 6 | – | – | 0 | 13 | 25,4 | 0,82 |
| 7 | – | – | 0 | 38 | 25,7 | 0,90 |
| 8 | – | – | 0 | 29 | 25,1 | 0,69 |
| 9 | – | – | 0 | 10 | 28,2 | 0,93 |
| 10 | – | – | 0 | 19 | 27,4 | 0,88 |
| 11 | – | – | 0 | 25 | 30,4 | 0,48 |
| 12 | – | – | 0 | 18 | 30,8 | 0,56 |
| 13 | 6 | – | 1 | 18 | 25,8 | 0,96 |
| 14 | 5 | – | 1 | 20 | 32,4 | 0,94 |
| 15 | 5 | – | 1 | 7 | 27,1 | 0,84 |
| 16 | – | – | 0 | 12 | 30,9 | 0,46 |
| **Median** | 6 | – | 0 | 18 | 27,1 | 0,83 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,84 | 403 |
| 15–29 | 0,85 | 200 |
| 30–44 | 0,81 | 90 |
| 45–59 | 0,85 | 39 |
| 60+ | 0,78 | 20 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 29,8 |
| 10–19 | 30,0 |
| 20–29 | 27,7 |
| 30–39 | 24,4 |
| 40–49 | 25,7 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 37 | 0 | 21 | 12 | 0,99 | 131 / 145 / 61 / 0 / 0 | 0 | 48 / 46 |
| 2 | 96 | 0 | 6 | 53 | 0,98 | 134 / 148 / 30 / 22 / 0 | 0 | 33 / 32 |
| 3 | 128 | 0 | 11 | 75 | 0,98 | 166 / 137 / 22 / 0 / 0 | 0 | 27 / 25 |
| 4 | 112 | 0 | 11 | 38 | 0,99 | 108 / 143 / 67 / 11 / 1 | 0 | 33 / 33 |
| 5 | 126 | 0 | 3 | 64 | 0,98 | 162 / 124 / 64 / 15 / 0 | 0 | 36 / 35 |
| 6 | 80 | 0 | 76 | 13 | 0,99 | 144 / 169 / 34 / 2 / 5 | 0 | 50 / 50 |
| 7 | 33 | 0 | 34 | 16 | 0,99 | 153 / 184 / 19 / 1 / 4 | 0 | 56 / 56 |
| 8 | 60 | 0 | 13 | 23 | 0,99 | 86 / 147 / 15 / 15 / 0 | 0 | 39 / 38 |
| 9 | 76 | 0 | 45 | 75 | 0,99 | 149 / 178 / 8 / 26 / 2 | 0 | 25 / 25 |
| 10 | 41 | 0 | 9 | 16 | 0,99 | 113 / 152 / 28 / 1 / 0 | 0 | 52 / 51 |
| 11 | 207 | 0 | 15 | 38 | 0,99 | 112 / 116 / 63 / 1 / 5 | 0 | 33 / 33 |
| 12 | 57 | 0 | 6 | 36 | 0,99 | 110 / 176 / 26 / 22 / 3 | 0 | 34 / 34 |
| 13 | 235 | 0 | 3 | 80 | 0,97 | 143 / 222 / 114 / 1 / 1 | 0 | 36 / 36 |
| 14 | 63 | 0 | 12 | 43 | 0,99 | 148 / 115 / 0 / 18 / 0 | 0 | 27 / 25 |
| 15 | 50 | 0 | 11 | 21 | 1,00 | 126 / 119 / 7 / 25 / 4 | 0 | 34 / 34 |
| 16 | 50 | 0 | 4 | 21 | 0,99 | 101 / 170 / 35 / 2 / 1 | 0 | 38 / 38 |
| **Median** | 70 | 0 | 11 | 37 | 0,99 | | | 35 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 16/16 |
| Babi Hutan | 100% | 100% | 16/16 |
| Ayam Hutan | 97% | 76% | 15/16 |
| Kerbau Liar | 74% | 45% | 14/16 |
| Harimau | 54% | 45% | 9/16 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 93% | 6% | 0% | 0% | 0,05 |
| 2 | 52% | 47% | 0% | 1% | 0,74 |
| 3 | 42% | 58% | 0% | 1% | 0,62 |
| 4 | 76% | 24% | 0% | 0% | 0,18 |
| 5 | 94% | 6% | 0% | 0% | 0,09 |
| 6 | 61% | 38% | 0% | 0% | 0,18 |
| 7 | 70% | 28% | 2% | 0% | 0,79 |
| 8 | 92% | 8% | 0% | 0% | 0,14 |
| 9 | 85% | 15% | 0% | 0% | 0,41 |
| 10 | 40% | 55% | 5% | 0% | 0,36 |
| 11 | 92% | 0% | 8% | 0% | 0,01 |
| 12 | 98% | 2% | 0% | 0% | 0,03 |
| 13 | 34% | 66% | 0% | 0% | 0,71 |
| 14 | 82% | 18% | 0% | 0% | 0,43 |
| 15 | 70% | 30% | 0% | 0% | 0,32 |
| 16 | 96% | 1% | 3% | 0% | 0,05 |
| **Median** | | 21% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 1270 | 17,7 | 20,1 | 36,3 | 45,2 |
| Netral | 4400 | 17,6 | 17,1 | 37,8 | 46,1 |
| La Niña | 2015 | 15,8 | 15,7 | 37,0 | 44,7 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 1167 | ×1,20 | ×1,41 | 31% |
| La Niña | 1841 | ×1,10 | ×1,04 | 23% |
