# Laporan soak — Fase 2 — tanpa iklim

- Dibuat: 2026-10-06 11:00:43
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: climate
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 115 m | 2 | 20 | 3 | 0 | 2 | 0 | 34,5 | 0,67 | 30,2 | 9,3 | 3,4 | 16,2 | – | 0,38 | 0 | 0,12 |
| 2 | 120 m (utuh) | 1 | 16 | 38 | 0 | 3 | 3 | 20,3 | 0,49 | 17,2 | 8,4 | 3,8 | 18,4 | 16 | 0,25 | 1626 | 0,37 |
| 3 | 120 m (utuh) | 1 | 543 | 41 | 0 | 5 | 5 | 19,2 | 0,51 | 13,2 | 12,5 | 3,3 | 16,4 | 18 | 0,49 | 332 | 2,11 |
| 4 | 120 m (utuh) | 1 | 162 | 45 | 0 | 2 | 4 | 19,8 | 0,43 | 16,1 | 13,1 | 3,1 | 16,4 | 16 | 0,54 | 0 | 1,05 |
| 5 | 22 m | 4 | 1269 | 23 | 0 | 2 | 6 | 20,2 | 0,54 | 12,8 | 11,9 | 3,2 | 16,1 | 18 | 0,57 | 8 | 3,41 |
| 6 | 120 m (utuh) | 1 | 300 | 45 | 1 | 6 | 2 | 19,8 | 0,49 | 13,6 | 8,9 | 3,6 | 16,5 | 18 | 0,59 | 263 | 1,49 |
| 7 | 11 m | 2 | 144 | 41 | 0 | 4 | 2 | 18,8 | 0,48 | 12,6 | 12,1 | 3,2 | 16,5 | 18 | 0,51 | 13 | 0,63 |
| 8 | 16 m | 3 | 1074 | 32 | 0 | 5 | 8 | 19,0 | 0,48 | 13,0 | 11,7 | 3,2 | 16,3 | 18 | 0,61 | 124 | 5,00 |
| 9 | 10 m | 2 | 83 | 38 | 0 | 1 | 4 | 20,0 | 0,58 | 11,6 | 6,8 | 4,9 | 17,5 | 18 | 0,57 | 119 | 0,40 |
| 10 | 120 m (utuh) | 1 | 1096 | 46 | 1 | 6 | 12 | 18,2 | 0,35 | 16,0 | 12,4 | 3,1 | 16,4 | 18 | 0,55 | 203 | 12,25 |
| 11 | 13 m | 2 | 772 | 43 | 0 | 2 | 6 | 18,8 | 0,52 | 11,4 | 10,5 | 3,4 | 16,5 | 18 | 0,52 | 115 | 3,61 |
| 12 | 120 m (utuh) | 1 | 1315 | 46 | 1 | 6 | 8 | 18,9 | 0,47 | 12,4 | 12,3 | 3,2 | 16,1 | 18 | 0,54 | 431 | 8,27 |
| 13 | 19 m | 3 | 74 | 29 | 0 | 3 | 5 | 21,1 | 0,53 | 16,0 | 7,4 | 4,5 | 18,6 | 20 | 0,50 | 262 | 0,34 |
| 14 | 16 m | 2 | 257 | 42 | 1 | 4 | 3 | 18,0 | 0,43 | 11,4 | 13,0 | 3,2 | 16,3 | 16 | 0,48 | 143 | 1,60 |
| 15 | 120 m (utuh) | 1 | 1136 | 47 | 1 | 6 | 9 | 18,5 | 0,38 | 14,9 | 12,0 | 3,1 | 16,6 | 16 | 0,57 | 92 | 8,75 |
| 16 | 120 m (utuh) | 1 | 290 | 47 | 1 | 3 | 3 | 19,0 | 0,45 | 13,9 | 12,3 | 3,2 | 16,2 | 18 | 0,46 | 51 | 2,56 |
| **Median** | 118 m | | 295 | 42 | 0 | 4 | 4 | 19,1 | 0,49 | 13,4 | 11,9 | 3,2 | 16,4 | 18 | 0,53 | 122 | 1,86 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **8 dari 16** dunia (selang kepercayaan 95%: 28–72%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 19,1 tahun | 21–37 | ↓ di bawah | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,49 | 0,44–0,73 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 13,4 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 18 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 11,9 anak | 5–7 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,2 tahun | 2,8–3,3 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 16,4 tahun | 18–20 | ↓ di bawah | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,53 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian bayi (q0) | 0,000 |
| Rasio kelamin (♂ per 100 ♀) | 104 |
| Anggota per rumah | 65,7 |
| Pembunuhan per 100.000 tahun-orang | 122 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian |
| --- | ---: | ---: |
| Kelaparan | 18089 | 83% |
| Kehausan | 2007 | 9% |
| Usia tua | 946 | 4% |
| Dibunuh | 749 | 3% |
| Diterkam hewan | 0 | 0% |

Catatan: simulasi belum punya penyakit (Fase 3), sedangkan di masyarakat nyata penyakit menyebabkan lebih dari separuh kematian. Perbedaan ini temuan, bukan galat.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | – | – | 0 | 16 | 20,3 | 0,03 |
| 2 | – | – | 0 | 17 | 21,0 | 0,88 |
| 3 | – | – | 0 | 11 | 20,5 | 0,94 |
| 4 | – | – | 0 | 6 | 18,7 | 0,92 |
| 5 | – | – | 0 | 21 | 20,1 | 0,93 |
| 6 | 14 | – | 1 | 28 | 19,4 | 0,95 |
| 7 | – | – | 0 | 20 | 19,3 | 0,97 |
| 8 | – | – | 0 | 21 | 19,8 | 0,91 |
| 9 | – | – | 0 | 14 | 20,3 | 0,93 |
| 10 | 6 | – | 1 | 3 | 19,9 | 0,91 |
| 11 | – | – | 0 | 10 | 21,4 | 0,92 |
| 12 | 7 | – | 1 | 26 | 19,4 | 0,92 |
| 13 | – | – | 0 | 18 | 20,0 | 0,94 |
| 14 | 4 | – | 1 | 15 | 20,2 | 0,92 |
| 15 | 8 | – | 1 | 4 | 18,9 | 0,92 |
| 16 | 7 | – | 1 | 7 | 19,2 | 0,95 |
| **Median** | 7 | – | 0 | 16 | 19,9 | 0,92 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,92 | 5670 |
| 15–29 | 0,93 | 1663 |
| 30–44 | 0,93 | 664 |
| 45–59 | 0,93 | 394 |
| 60+ | 0,92 | 160 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 20,1 |
| 10–19 | 20,0 |
| 20–29 | 20,1 |
| 30–39 | 20,0 |
| 40–49 | 19,3 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 40 | 0 | 24 | 14 | 0,99 | 168 / 200 / 16 / 1 / 0 | 0 | 49 / 48 |
| 2 | 58 | 0 | 30 | 54 | 0,99 | 150 / 188 / 38 / 40 / 5 | 0 | 23 / 23 |
| 3 | 643 | 0 | 11 | 119 | 0,96 | 182 / 167 / 37 / 0 / 6 | 0 | 37 / 36 |
| 4 | 264 | 0 | 18 | 33 | 0,98 | 186 / 206 / 14 / 3 / 1 | 0 | 50 / 50 |
| 5 | 1352 | 0 | 15 | 172 | 0,92 | 120 / 192 / 66 / 51 / 2 | 0 | 25 / 25 |
| 6 | 438 | 0 | 17 | 106 | 0,96 | 100 / 153 / 53 / 96 / 3 | 0 | 24 / 24 |
| 7 | 168 | 0 | 2 | 38 | 0,99 | 138 / 139 / 59 / 47 / 5 | 0 | 21 / 21 |
| 8 | 1144 | 0 | 4 | 174 | 0,94 | 180 / 145 / 11 / 0 / 0 | 0 | 47 / 45 |
| 9 | 117 | 0 | 26 | 40 | 0,98 | 139 / 133 / 28 / 35 / 0 | 0 | 31 / 30 |
| 10 | 1401 | 0 | 8 | 184 | 0,95 | 120 / 147 / 25 / 20 / 3 | 0 | 42 / 42 |
| 11 | 993 | 0 | 27 | 171 | 0,94 | 179 / 97 / 13 / 1 / 0 | 0 | 48 / 47 |
| 12 | 1452 | 36746 | 11 | 227 | 0,94 | 138 / 149 / 12 / 46 / 0 | 0 | 27 / 26 |
| 13 | 128 | 0 | 42 | 84 | 0,98 | 152 / 134 / 11 / 16 / 0 | 0 | 33 / 32 |
| 14 | 315 | 0 | 3 | 51 | 0,99 | 167 / 132 / 0 / 10 / 0 | 0 | 42 / 40 |
| 15 | 1332 | 0 | 4 | 173 | 0,96 | 198 / 138 / 0 / 0 / 0 | 0 | 35 / 32 |
| 16 | 417 | 0 | 5 | 71 | 0,97 | 155 / 133 / 50 / 30 / 4 | 0 | 40 / 40 |
| **Median** | 428 | 0 | 13 | 95 | 0,97 | | | 36 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 16/16 |
| Babi Hutan | 100% | 98% | 16/16 |
| Ayam Hutan | 92% | 84% | 14/16 |
| Kerbau Liar | 82% | 59% | 13/16 |
| Harimau | 55% | 43% | 8/16 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 100% | 0% | 0% | 0% | 0,84 |
| 2 | 79% | 21% | 0% | 0% | 0,60 |
| 3 | 14% | 86% | 0% | 0% | 0,43 |
| 4 | 19% | 80% | 1% | 0% | 0,35 |
| 5 | 15% | 84% | 0% | 0% | 0,31 |
| 6 | 27% | 73% | 0% | 0% | 0,54 |
| 7 | 12% | 88% | 0% | 0% | 0,77 |
| 8 | 14% | 86% | 0% | 0% | 0,35 |
| 9 | 40% | 60% | 0% | 0% | 0,57 |
| 10 | 14% | 86% | 0% | 0% | 0,35 |
| 11 | 14% | 86% | 0% | 0% | 0,39 |
| 12 | 12% | 88% | 0% | 0% | 0,28 |
| 13 | 72% | 28% | 0% | 0% | 0,56 |
| 14 | 14% | 83% | 2% | 0% | 0,26 |
| 15 | 11% | 89% | 0% | 0% | 0,31 |
| 16 | 19% | 80% | 1% | 0% | 0,35 |
| **Median** | | 84% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 0 | – | – | – | – |
| Netral | 10615 | 40,2 | 40,2 | 51,2 | 39,1 |
| La Niña | 0 | – | – | – | – |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 0 | – | – | – |
| La Niña | 0 | – | – | – |
