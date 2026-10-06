# Laporan soak — Fase 2 — on

- Dibuat: 2026-10-06 10:30:31
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 120 m (utuh) | 1 | 109 | 43 | 0 | 1 | 4 | 18,5 | 0,42 | 14,7 | 6,8 | 4,8 | 19,8 | 18 | 0,37 | 105 | 0,84 |
| 2 | 120 m (utuh) | 1 | 137 | 42 | 0 | 3 | 6 | 20,7 | 0,50 | 15,4 | 9,1 | 3,4 | 18,2 | 16 | 0,57 | 324 | 0,85 |
| 3 | 13 m | 3 | 18 | 21 | 0 | 3 | 2 | 20,5 | 0,56 | 12,8 | 7,4 | 3,4 | 22,6 | 16 | 0,60 | 398 | 0,19 |
| 4 | 25 m | 7 | 63 | 15 | 0 | 3 | 1 | 24,8 | 0,66 | 17,0 | 6,9 | 3,7 | 17,8 | 18 | 0,45 | 0 | 0,16 |
| 5 | 28 m | 3 | 400 | 25 | 0 | 1 | 6 | 20,3 | 0,57 | 12,7 | 9,2 | 4,0 | 17,5 | 18 | 0,47 | 205 | 1,06 |
| 6 | 120 m (utuh) | 1 | 52 | 38 | 0 | 2 | 4 | 21,1 | 0,55 | 15,7 | 7,8 | 4,3 | 19,7 | 20 | 0,45 | 808 | 0,33 |
| 7 | 120 m (utuh) | 1 | 130 | 37 | 0 | 5 | 6 | 24,8 | 0,63 | 17,7 | 6,7 | 4,1 | 20,3 | 16 | 0,45 | 54 | 0,63 |
| 8 | 120 m (utuh) | 1 | 38 | 34 | 0 | 3 | 2 | 25,7 | 0,54 | 23,8 | 6,3 | 5,1 | 17,8 | 16 | 0,45 | 261 | 0,24 |
| 9 | 120 m (utuh) | 1 | 21 | 37 | 0 | 3 | 4 | 22,8 | 0,49 | 22,4 | 9,0 | 4,4 | 18,9 | 18 | 0,42 | 1867 | 0,21 |
| 10 | 20 m | 2 | 151 | 34 | 0 | 2 | 2 | 26,5 | 0,69 | 18,5 | 9,3 | 3,5 | 16,4 | 16 | 0,45 | 58 | 0,55 |
| 11 | 120 m (utuh) | 1 | 67 | 41 | 0 | 5 | 1 | 17,3 | 0,39 | 12,6 | 9,2 | 3,9 | 16,9 | 18 | 0,50 | 675 | 0,62 |
| 12 | 120 m (utuh) | 1 | 370 | 39 | 1 | 6 | 6 | 21,9 | 0,63 | 13,8 | 10,2 | 3,5 | 16,2 | 16 | 0,56 | 734 | 1,08 |
| 13 | 120 m (utuh) | 1 | 122 | 44 | 0 | 5 | 6 | 19,4 | 0,48 | 13,3 | 9,4 | 3,7 | 16,7 | 16 | 0,55 | 247 | 1,16 |
| 14 | 120 m (utuh) | 1 | 21 | 43 | 0 | 4 | 2 | 24,8 | 0,50 | 24,6 | 5,5 | 6,0 | 28,8 | – | 0,23 | 1173 | 0,18 |
| 15 | 15 m | 3 | 15 | 9 | 0 | 4 | 2 | 26,8 | 0,48 | 27,8 | 6,2 | 3,9 | 25,2 | – | 0,14 | 0 | 0,14 |
| 16 | 120 m (utuh) | 1 | 286 | 44 | 1 | 5 | 8 | 18,7 | 0,41 | 14,3 | 9,5 | 3,5 | 17,3 | 18 | 0,41 | 91 | 2,29 |
| **Median** | 120 m | | 88 | 38 | 0 | 3 | 4 | 21,5 | 0,52 | 15,6 | 8,4 | 3,9 | 18,0 | 17 | 0,45 | 254 | 0,59 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **11 dari 16** dunia (selang kepercayaan 95%: 44–86%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 21,5 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,52 | 0,44–0,73 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 15,6 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 17 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 8,4 anak | 5–7 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,9 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 18,0 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,45 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian bayi (q0) | 0,000 |
| Rasio kelamin (♂ per 100 ♀) | 93 |
| Anggota per rumah | 10,8 |
| Pembunuhan per 100.000 tahun-orang | 254 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian |
| --- | ---: | ---: |
| Kelaparan | 3413 | 70% |
| Kehausan | 847 | 17% |
| Usia tua | 266 | 5% |
| Dibunuh | 348 | 7% |
| Diterkam hewan | 0 | 0% |

Catatan: simulasi belum punya penyakit (Fase 3), sedangkan di masyarakat nyata penyakit menyebabkan lebih dari separuh kematian. Perbedaan ini temuan, bukan galat.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | – | – | 0 | 5 | 20,5 | 0,92 |
| 2 | – | – | 0 | 25 | 18,5 | 0,88 |
| 3 | – | – | 0 | 22 | 20,2 | 0,90 |
| 4 | – | – | 0 | 23 | 19,0 | 0,95 |
| 5 | – | – | 0 | 14 | 20,8 | 0,93 |
| 6 | – | – | 0 | 7 | 21,1 | 0,83 |
| 7 | – | – | 0 | 22 | 21,1 | 0,91 |
| 8 | – | – | 0 | 15 | 22,3 | 0,40 |
| 9 | – | – | 0 | 10 | 19,9 | 0,49 |
| 10 | – | – | 0 | 6 | 19,1 | 0,93 |
| 11 | – | – | 0 | 21 | 19,8 | 0,89 |
| 12 | 6 | – | 1 | 21 | 21,0 | 0,96 |
| 13 | – | – | 0 | 23 | 20,6 | 0,85 |
| 14 | – | – | 0 | 26 | 18,9 | 0,23 |
| 15 | – | – | 0 | 19 | 20,1 | 0,50 |
| 16 | 16 | – | 1 | 36 | 20,0 | 0,90 |
| **Median** | 11 | – | 0 | 21 | 20,1 | 0,89 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,88 | 1275 |
| 15–29 | 0,90 | 384 |
| 30–44 | 0,90 | 185 |
| 45–59 | 0,89 | 102 |
| 60+ | 0,89 | 54 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 20,0 |
| 10–19 | 20,1 |
| 20–29 | 20,3 |
| 30–39 | 20,2 |
| 40–49 | 19,7 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 240 | 0 | 32 | 52 | 0,98 | 167 / 169 / 3 / 6 / 0 | 2 | 30 / 29 |
| 2 | 204 | 0 | 8 | 56 | 0,97 | 88 / 149 / 74 / 32 / 0 | 0 | 25 / 24 |
| 3 | 62 | 0 | 9 | 29 | 0,99 | 152 / 129 / 29 / 0 / 0 | 0 | 46 / 44 |
| 4 | 74 | 0 | 42 | 18 | 0,99 | 126 / 131 / 39 / 0 / 2 | 0 | 41 / 39 |
| 5 | 444 | 0 | 3 | 100 | 0,96 | 134 / 60 / 10 / 10 / 4 | 0 | 41 / 41 |
| 6 | 64 | 0 | 16 | 27 | 0,98 | 88 / 201 / 15 / 18 / 0 | 0 | 32 / 31 |
| 7 | 198 | 0 | 13 | 76 | 0,97 | 96 / 158 / 13 / 22 / 0 | 0 | 26 / 25 |
| 8 | 59 | 0 | 5 | 10 | 0,99 | 108 / 179 / 60 / 14 / 2 | 0 | 26 / 26 |
| 9 | 47 | 0 | 44 | 2 | 0,99 | 92 / 154 / 55 / 12 / 6 | 0 | 37 / 37 |
| 10 | 151 | 0 | 51 | 20 | 0,99 | 120 / 184 / 40 / 0 / 2 | 0 | 34 / 33 |
| 11 | 252 | 0 | 9 | 40 | 0,97 | 111 / 166 / 60 / 2 / 0 | 0 | 40 / 39 |
| 12 | 383 | 0 | 3 | 99 | 0,96 | 133 / 156 / 54 / 18 / 0 | 0 | 34 / 33 |
| 13 | 288 | 0 | 3 | 64 | 0,98 | 166 / 141 / 8 / 4 / 4 | 0 | 40 / 40 |
| 14 | 44 | 0 | 17 | 11 | 1,00 | 131 / 162 / 26 / 14 / 2 | 0 | 48 / 47 |
| 15 | 28 | 0 | 30 | 3 | 0,99 | 79 / 203 / 0 / 21 / 4 | 0 | 33 / 32 |
| 16 | 438 | 0 | 18 | 102 | 0,96 | 84 / 188 / 3 / 11 / 0 | 0 | 29 / 28 |
| **Median** | 174 | 0 | 14 | 34 | 0,98 | | | 34 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 98% | 16/16 |
| Babi Hutan | 100% | 100% | 16/16 |
| Ayam Hutan | 99% | 82% | 15/16 |
| Kerbau Liar | 79% | 58% | 13/16 |
| Harimau | 51% | 38% | 8/16 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 47% | 53% | 0% | 0% | 0,36 |
| 2 | 52% | 48% | 0% | 0% | 0,32 |
| 3 | 59% | 41% | 0% | 0% | 0,37 |
| 4 | 54% | 46% | 0% | 0% | 0,66 |
| 5 | 24% | 76% | 0% | 0% | 0,70 |
| 6 | 76% | 18% | 6% | 1% | 0,23 |
| 7 | 46% | 54% | 0% | 0% | 0,58 |
| 8 | 100% | 0% | 0% | 0% | 0,01 |
| 9 | 100% | 0% | 0% | 0% | 0,00 |
| 10 | 33% | 66% | 1% | 0% | 0,39 |
| 11 | 39% | 61% | 0% | 0% | 0,34 |
| 12 | 20% | 79% | 0% | 0% | 0,80 |
| 13 | 48% | 52% | 0% | 0% | 0,22 |
| 14 | 99% | 0% | 1% | 0% | 0,00 |
| 15 | 89% | 0% | 11% | 0% | 0,00 |
| 16 | 38% | 62% | 0% | 0% | 0,36 |
| **Median** | | 50% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 1567 | 31,1 | 33,3 | 47,7 | 38,5 |
| Netral | 5546 | 30,4 | 30,1 | 47,4 | 41,0 |
| La Niña | 2508 | 30,4 | 29,9 | 47,2 | 38,7 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 1524 | ×1,10 | ×1,29 | 31% |
| La Niña | 2448 | ×1,10 | ×1,07 | 24% |
