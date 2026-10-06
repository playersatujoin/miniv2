# Laporan soak — Fase 2 — on

- Dibuat: 2026-10-06 12:07:26
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 120 m (utuh) | 1 | 61 | 43 | 0 | 2 | 3 | 16,5 | 0,38 | 10,5 | 9,4 | 3,4 | 17,2 | 18 | 0,52 | 328 | 0,97 |
| 2 | 120 m (utuh) | 1 | 375 | 45 | 1 | 4 | 8 | 21,5 | 0,55 | 15,0 | 12,6 | 3,3 | 16,5 | 18 | 0,51 | 262 | 2,38 |
| 3 | 43 m | 2 | 47 | 26 | 0 | 1 | 2 | 26,6 | 0,58 | 22,7 | 7,0 | 3,5 | 19,3 | 16 | 0,38 | 0 | 0,30 |
| 4 | 120 m (utuh) | 1 | 70 | 43 | 0 | 1 | 5 | 22,0 | 0,54 | 16,2 | 8,9 | 3,5 | 16,3 | 16 | 0,55 | 81 | 0,67 |
| 5 | 120 m (utuh) | 1 | 15 | 40 | 0 | 5 | 1 | 24,1 | 0,55 | 19,5 | 6,5 | 3,9 | 24,9 | – | 0,31 | 0 | 0,37 |
| 6 | 18 m | 2 | 17 | 34 | 0 | 3 | 3 | 29,1 | 0,57 | 28,3 | 2,0 | 6,2 | – | – | 0,53 | 478 | 0,17 |
| 7 | 71 m | 2 | 43 | 17 | 0 | 2 | 1 | 21,4 | 0,57 | 14,6 | 11,3 | 4,2 | 17,2 | 16 | 0,39 | 53 | 0,17 |
| 8 | 120 m (utuh) | 1 | 12 | 36 | 0 | 1 | 2 | 20,3 | 0,61 | 11,8 | 9,4 | 4,1 | 17,1 | 18 | 0,28 | 942 | 0,16 |
| 9 | 120 m (utuh) | 1 | 44 | 39 | 0 | 4 | 3 | 24,2 | 0,56 | 19,9 | 3,7 | 6,1 | 24,1 | 18 | 0,26 | 489 | 0,30 |
| 10 | 23 m | 2 | 126 | 31 | 0 | 5 | 5 | 24,2 | 0,57 | 18,6 | 8,9 | 3,9 | 16,8 | 16 | 0,55 | 57 | 0,70 |
| 11 | 81 m | 3 | 23 | 9 | 0 | 1 | 1 | 23,6 | 0,46 | 24,8 | 4,8 | 5,2 | 27,0 | 62 | 0,21 | 0 | 0,29 |
| 12 | 120 m (utuh) | 1 | 315 | 45 | 0 | 2 | 12 | 20,1 | 0,50 | 14,1 | 9,8 | 3,5 | 17,6 | 18 | 0,51 | 97 | 1,38 |
| 13 | 120 m (utuh) | 1 | 802 | 48 | 1 | 5 | 11 | 19,1 | 0,46 | 12,8 | 12,3 | 3,2 | 16,4 | 18 | 0,54 | 42 | 6,51 |
| 14 | 10 m | 5 | 37 | 23 | 0 | 4 | 3 | 26,7 | 0,66 | 20,7 | 7,0 | 4,6 | 17,5 | 20 | 0,57 | 585 | 0,26 |
| 15 | 16 m | 4 | 38 | 18 | 0 | 2 | 1 | 21,2 | 0,60 | 13,6 | 12,9 | 3,5 | 16,3 | 16 | 0,39 | 145 | 0,18 |
| 16 | 120 m (utuh) | 1 | 152 | 41 | 0 | 5 | 8 | 23,3 | 0,53 | 19,1 | 7,2 | 4,7 | 19,2 | 16 | 0,48 | 36 | 1,56 |
| **Median** | 120 m | | 46 | 38 | 0 | 2 | 3 | 22,6 | 0,56 | 17,4 | 8,9 | 3,9 | 17,2 | 18 | 0,50 | 89 | 0,33 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **9 dari 16** dunia (selang kepercayaan 95%: 33–77%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 22,6 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,56 | 0,44–0,73 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 17,4 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 18 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 8,9 anak | 5–7 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,9 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 17,2 tahun | 18–20 | ↓ di bawah | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,50 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian bayi (q0) | 0,000 |
| Rasio kelamin (♂ per 100 ♀) | 114 |
| Anggota per rumah | 12,6 |
| Pembunuhan per 100.000 tahun-orang | 89 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian |
| --- | ---: | ---: |
| Kelaparan | 4365 | 78% |
| Kehausan | 718 | 13% |
| Usia tua | 357 | 6% |
| Dibunuh | 151 | 3% |
| Diterkam hewan | 0 | 0% |

Catatan: simulasi belum punya penyakit (Fase 3), sedangkan di masyarakat nyata penyakit menyebabkan lebih dari separuh kematian. Perbedaan ini temuan, bukan galat.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | – | – | 0 | 6 | 19,3 | 0,88 |
| 2 | 13 | – | 1 | 25 | 19,4 | 0,95 |
| 3 | – | – | 0 | 23 | 21,1 | 0,64 |
| 4 | – | – | 0 | 16 | 19,2 | 0,85 |
| 5 | – | – | 0 | 13 | 19,3 | 0,53 |
| 6 | – | – | 0 | 18 | 19,7 | 0,62 |
| 7 | – | – | 0 | 20 | 20,8 | 0,95 |
| 8 | – | – | 0 | 25 | 18,9 | 0,94 |
| 9 | – | – | 0 | 18 | 19,6 | 0,86 |
| 10 | – | – | 0 | 17 | 20,3 | 0,84 |
| 11 | – | – | 0 | 7 | 20,0 | 0,00 |
| 12 | – | – | 0 | 14 | 19,3 | 0,94 |
| 13 | 8 | – | 1 | 6 | 20,5 | 0,95 |
| 14 | – | – | 0 | 28 | 20,3 | 0,48 |
| 15 | – | – | 0 | 16 | 20,2 | 0,94 |
| 16 | – | – | 0 | 23 | 19,9 | 0,83 |
| **Median** | 10 | – | 0 | 18 | 19,8 | 0,86 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,87 | 1336 |
| 15–29 | 0,87 | 473 |
| 30–44 | 0,92 | 176 |
| 45–59 | 0,89 | 139 |
| 60+ | 0,85 | 53 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 20,1 |
| 10–19 | 20,1 |
| 20–29 | 19,8 |
| 30–39 | 19,5 |
| 40–49 | 20,1 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 261 | 0 | 23 | 68 | 0,98 | 105 / 147 / 33 / 0 / 6 | 0 | 37 / 36 |
| 2 | 414 | 0 | 5 | 91 | 0,97 | 142 / 115 / 23 / 4 / 0 | 0 | 51 / 50 |
| 3 | 57 | 0 | 6 | 10 | 0,98 | 148 / 142 / 0 / 0 / 4 | 0 | 39 / 37 |
| 4 | 109 | 0 | 15 | 31 | 0,98 | 140 / 141 / 0 / 21 / 0 | 0 | 31 / 29 |
| 5 | 122 | 0 | 6 | 35 | 0,99 | 164 / 163 / 13 / 0 / 5 | 0 | 58 / 57 |
| 6 | 34 | 0 | 81 | 12 | 0,99 | 105 / 151 / 62 / 20 / 4 | 0 | 27 / 27 |
| 7 | 44 | 0 | 4 | 16 | 0,99 | 120 / 167 / 16 / 53 / 0 | 0 | 32 / 31 |
| 8 | 36 | 0 | 21 | 29 | 0,99 | 130 / 209 / 15 / 3 / 0 | 0 | 36 / 35 |
| 9 | 78 | 0 | 40 | 53 | 0,98 | 133 / 154 / 34 / 6 / 0 | 0 | 30 / 29 |
| 10 | 216 | 0 | 17 | 41 | 0,98 | 139 / 135 / 22 / 19 / 3 | 0 | 55 / 55 |
| 11 | 91 | 0 | 24 | 23 | 0,99 | 119 / 199 / 85 / 3 / 0 | 0 | 38 / 37 |
| 12 | 430 | 0 | 7 | 122 | 0,96 | 114 / 65 / 29 / 39 / 4 | 0 | 43 / 43 |
| 13 | 994 | 0 | 3 | 174 | 0,96 | 153 / 136 / 64 / 0 / 7 | 0 | 41 / 40 |
| 14 | 78 | 0 | 47 | 3 | 0,99 | 103 / 143 / 103 / 12 / 5 | 0 | 37 / 37 |
| 15 | 70 | 0 | 17 | 23 | 0,99 | 100 / 148 / 10 / 10 / 0 | 0 | 42 / 41 |
| 16 | 415 | 0 | 4 | 74 | 0,96 | 137 / 90 / 7 / 27 / 1 | 0 | 33 / 33 |
| **Median** | 100 | 0 | 16 | 33 | 0,98 | | | 38 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 16/16 |
| Babi Hutan | 100% | 100% | 16/16 |
| Ayam Hutan | 94% | 77% | 14/16 |
| Kerbau Liar | 71% | 56% | 12/16 |
| Harimau | 52% | 41% | 9/16 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 35% | 65% | 0% | 0% | 0,31 |
| 2 | 18% | 82% | 0% | 0% | 0,76 |
| 3 | 99% | 1% | 0% | 0% | 0,01 |
| 4 | 51% | 49% | 0% | 0% | 0,29 |
| 5 | 100% | 0% | 0% | 0% | 0,00 |
| 6 | 99% | 0% | 1% | 0% | 0,00 |
| 7 | 49% | 47% | 4% | 0% | 0,64 |
| 8 | 77% | 23% | 0% | 0% | 0,43 |
| 9 | 67% | 33% | 0% | 0% | 0,45 |
| 10 | 61% | 39% | 0% | 0% | 0,19 |
| 11 | 100% | 0% | 0% | 0% | 0,00 |
| 12 | 26% | 73% | 0% | 0% | 0,70 |
| 13 | 12% | 88% | 0% | 0% | 0,62 |
| 14 | 100% | 0% | 0% | 0% | 0,01 |
| 15 | 65% | 35% | 0% | 0% | 0,36 |
| 16 | 49% | 50% | 0% | 0% | 0,22 |
| **Median** | | 37% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 1659 | 34,6 | 36,7 | 46,8 | 52,8 |
| Netral | 5528 | 35,2 | 34,5 | 47,1 | 53,5 |
| La Niña | 2529 | 35,4 | 35,8 | 46,7 | 52,0 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 1616 | ×1,10 | ×1,23 | 27% |
| La Niña | 2444 | ×1,09 | ×1,09 | 23% |
