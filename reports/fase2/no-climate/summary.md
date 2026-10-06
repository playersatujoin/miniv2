# Laporan soak — Fase 2 — tanpa iklim

- Dibuat: 2026-10-06 12:14:26
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: climate
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 120 m (utuh) | 1 | 40 | 41 | 0 | 2 | 2 | 19,4 | 0,42 | 14,9 | 12,1 | 3,9 | 17,2 | 18 | 0,26 | 55 | 0,21 |
| 2 | 25 m | 2 | 64 | 30 | 0 | 1 | 4 | 33,0 | 0,65 | 29,4 | 4,1 | 6,4 | 17,9 | 16 | 0,49 | 73 | 0,26 |
| 3 | 120 m (utuh) | 1 | 38 | 36 | 0 | 1 | 4 | 32,1 | 0,53 | 36,5 | 6,4 | 3,9 | 19,7 | 68 | 0,64 | 0 | 0,36 |
| 4 | 120 m (utuh) | 1 | 130 | 44 | 0 | 2 | 1 | 19,0 | 0,54 | 11,4 | 11,9 | 3,4 | 16,2 | 18 | 0,40 | 64 | 1,25 |
| 5 | 120 m (utuh) | 1 | 137 | 39 | 0 | 2 | 1 | 19,9 | 0,51 | 13,3 | 12,8 | 3,3 | 15,9 | 18 | 0,47 | 78 | 0,88 |
| 6 | 120 m (utuh) | 1 | 779 | 44 | 1 | 4 | 12 | 20,2 | 0,55 | 12,8 | 11,7 | 3,2 | 16,4 | 18 | 0,53 | 453 | 2,36 |
| 7 | 11 m | 2 | 44 | 32 | 0 | 1 | 5 | 21,7 | 0,51 | 17,5 | 6,1 | 4,4 | 21,2 | 16 | 0,51 | 229 | 0,35 |
| 8 | 9 m | 3 | 1055 | 40 | 1 | 5 | 8 | 18,9 | 0,42 | 14,8 | 12,4 | 3,2 | 16,1 | 18 | 0,60 | 114 | 6,78 |
| 9 | 120 m (utuh) | 1 | 1118 | 44 | 0 | 5 | 10 | 19,7 | 0,50 | 13,3 | 12,8 | 3,1 | 16,1 | 18 | 0,54 | 410 | 8,32 |
| 10 | 120 m (utuh) | 1 | 359 | 44 | 0 | 4 | 1 | 20,7 | 0,54 | 14,0 | 12,8 | 3,1 | 16,1 | 18 | 0,52 | 110 | 2,43 |
| 11 | 49 m | 3 | 504 | 23 | 0 | 2 | 2 | 21,1 | 0,52 | 16,1 | 11,6 | 3,4 | 16,6 | 18 | 0,47 | 219 | 0,94 |
| 12 | 78 m | 2 | 30 | 15 | 0 | 3 | 3 | 21,2 | 0,52 | 15,6 | 9,3 | 3,8 | 16,6 | 16 | 0,22 | 0 | 0,21 |
| 13 | 27 m | 2 | 728 | 35 | 0 | 5 | 3 | 19,1 | 0,48 | 13,3 | 12,0 | 3,2 | 16,3 | 16 | 0,53 | 49 | 3,50 |
| 14 | 94 m | 3 | 11 | 3 | 0 | 2 | 1 | 26,1 | 0,44 | 30,8 | 6,8 | 3,8 | – | – | 0,23 | 0 | 0,14 |
| 15 | 10 m | 6 | 1 | 0 | 0 | 2 | 0 | – | – | – | – | – | – | – | – | – | 0,13 |
| 16 | 18 m | 2 | 1138 | 41 | 0 | 5 | 13 | 19,1 | 0,39 | 16,2 | 12,1 | 3,2 | 16,2 | 16 | 0,56 | 186 | 7,86 |
| **Median** | 86 m | | 134 | 38 | 0 | 2 | 3 | 20,2 | 0,51 | 14,9 | 11,9 | 3,4 | 16,4 | 18 | 0,51 | 78 | 0,91 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **7 dari 16** dunia (selang kepercayaan 95%: 23–67%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 20,2 tahun | 21–37 | ↓ di bawah | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,51 | 0,44–0,73 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 14,9 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 18 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 11,9 anak | 5–7 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,4 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 16,4 tahun | 18–20 | ↓ di bawah | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,51 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian bayi (q0) | 0,000 |
| Rasio kelamin (♂ per 100 ♀) | 111 |
| Anggota per rumah | 31,5 |
| Pembunuhan per 100.000 tahun-orang | 78 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian |
| --- | ---: | ---: |
| Kelaparan | 12075 | 81% |
| Kehausan | 1506 | 10% |
| Usia tua | 678 | 5% |
| Dibunuh | 665 | 4% |
| Diterkam hewan | 0 | 0% |

Catatan: simulasi belum punya penyakit (Fase 3), sedangkan di masyarakat nyata penyakit menyebabkan lebih dari separuh kematian. Perbedaan ini temuan, bukan galat.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | – | – | 0 | 12 | 20,8 | 0,96 |
| 2 | – | – | 0 | 14 | 19,7 | 0,29 |
| 3 | – | – | 0 | 14 | 21,8 | 0,65 |
| 4 | – | – | 0 | 17 | 20,2 | 0,93 |
| 5 | – | – | 0 | 17 | 21,5 | 0,94 |
| 6 | 19 | – | 1 | 28 | 19,6 | 0,94 |
| 7 | – | – | 0 | 14 | 18,3 | 0,57 |
| 8 | 8 | – | 1 | 23 | 20,8 | 0,89 |
| 9 | – | – | 0 | 15 | 20,3 | 0,93 |
| 10 | – | – | 0 | 9 | 19,1 | 0,95 |
| 11 | – | – | 0 | 18 | 20,4 | 0,89 |
| 12 | – | – | 0 | 21 | 19,7 | 0,89 |
| 13 | – | – | 0 | 8 | 18,7 | 0,93 |
| 14 | – | – | 0 | 22 | 18,7 | 0,07 |
| 15 | – | – | 0 | 30 | 20,0 | 0,31 |
| 16 | – | – | 0 | 25 | 20,5 | 0,91 |
| **Median** | 14 | – | 0 | 17 | 20,1 | 0,90 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,90 | 4053 |
| 15–29 | 0,92 | 1217 |
| 30–44 | 0,92 | 460 |
| 45–59 | 0,91 | 291 |
| 60+ | 0,93 | 155 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 19,9 |
| 10–19 | 20,0 |
| 20–29 | 20,0 |
| 30–39 | 20,4 |
| 40–49 | 19,8 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 73 | 0 | 23 | 28 | 0,99 | 187 / 179 / 69 / 1 / 2 | 0 | 35 / 35 |
| 2 | 87 | 0 | 9 | 33 | 0,98 | 154 / 183 / 38 / 29 / 1 | 0 | 16 / 16 |
| 3 | 67 | 0 | 29 | 11 | 0,99 | 168 / 207 / 10 / 35 / 0 | 7 | 23 / 22 |
| 4 | 149 | 0 | 22 | 37 | 0,98 | 158 / 211 / 7 / 16 / 0 | 0 | 43 / 42 |
| 5 | 147 | 0 | 54 | 18 | 0,98 | 188 / 150 / 1 / 1 / 0 | 0 | 34 / 33 |
| 6 | 779 | 0 | 27 | 170 | 0,94 | 133 / 194 / 47 / 15 / 4 | 0 | 28 / 28 |
| 7 | 59 | 0 | 2 | 7 | 0,99 | 144 / 125 / 7 / 0 / 6 | 0 | 41 / 40 |
| 8 | 1154 | 0 | 3 | 173 | 0,96 | 139 / 180 / 0 / 69 / 5 | 0 | 36 / 35 |
| 9 | 1285 | 0 | 7 | 186 | 0,95 | 178 / 155 / 8 / 6 / 0 | 0 | 25 / 24 |
| 10 | 420 | 0 | 37 | 59 | 0,98 | 196 / 164 / 24 / 0 / 0 | 0 | 31 / 29 |
| 11 | 504 | 0 | 27 | 122 | 0,97 | 128 / 96 / 37 / 55 / 0 | 0 | 43 / 42 |
| 12 | 70 | 0 | 21 | 28 | 0,99 | 136 / 192 / 13 / 23 / 0 | 0 | 36 / 35 |
| 13 | 894 | 0 | 41 | 168 | 0,95 | 138 / 141 / 20 / 53 / 2 | 0 | 27 / 27 |
| 14 | 35 | 0 | 4 | 12 | 0,99 | 161 / 113 / 55 / 0 / 0 | 0 | 31 / 29 |
| 15 | 33 | 0 | 4 | 23 | 1,00 | 201 / 144 / 0 / 4 / 3 | 0 | 34 / 33 |
| 16 | 1265 | 0 | 12 | 173 | 0,95 | 179 / 189 / 8 / 31 / 0 | 1 | 31 / 30 |
| **Median** | 148 | 0 | 22 | 35 | 0,98 | | | 32 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 16/16 |
| Babi Hutan | 100% | 100% | 16/16 |
| Ayam Hutan | 92% | 62% | 14/16 |
| Kerbau Liar | 96% | 60% | 13/16 |
| Harimau | 51% | 42% | 7/16 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 60% | 35% | 4% | 2% | 0,51 |
| 2 | 100% | 0% | 0% | 0% | 0,00 |
| 3 | 94% | 5% | 0% | 0% | 0,15 |
| 4 | 24% | 75% | 1% | 0% | 0,32 |
| 5 | 26% | 72% | 2% | 0% | 0,31 |
| 6 | 16% | 83% | 0% | 1% | 0,50 |
| 7 | 99% | 0% | 0% | 1% | 0,00 |
| 8 | 11% | 88% | 0% | 1% | 0,35 |
| 9 | 13% | 87% | 0% | 1% | 0,31 |
| 10 | 18% | 82% | 0% | 0% | 0,31 |
| 11 | 20% | 80% | 0% | 0% | 0,42 |
| 12 | 51% | 49% | 0% | 0% | 0,69 |
| 13 | 13% | 86% | 0% | 1% | 0,38 |
| 14 | 100% | 0% | 0% | 0% | 0,00 |
| 15 | 100% | 0% | 0% | 0% | 0,00 |
| 16 | 13% | 86% | 0% | 1% | 0,34 |
| **Median** | | 73% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 0 | – | – | – | – |
| Netral | 9626 | 38,2 | 38,2 | 50,2 | 38,8 |
| La Niña | 0 | – | – | – | – |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 0 | – | – | – |
| La Niña | 0 | – | – | – |
