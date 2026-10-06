# Laporan soak — Fase 2 — tanpa pertanian

- Dibuat: 2026-10-06 10:28:19
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: farming
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 120 m (utuh) | 1 | 22 | 40 | 0 | 1 | 3 | 25,6 | 0,64 | 19,2 | 5,0 | 5,2 | 17,5 | 16 | 0,45 | 273 | 0,28 |
| 2 | 120 m (utuh) | 1 | 21 | 34 | 0 | 1 | 3 | 30,4 | 0,59 | 28,8 | 3,7 | 6,1 | 20,2 | 76 | 0,38 | 75 | 0,33 |
| 3 | 13 m | 2 | 50 | 34 | 0 | 4 | 5 | 25,3 | 0,50 | 24,7 | 4,6 | 4,5 | 22,2 | 16 | 0,46 | 34 | 0,52 |
| 4 | 28 m | 4 | 85 | 18 | 0 | 1 | 1 | 24,4 | 0,46 | 26,3 | 6,7 | 4,2 | 19,3 | 16 | 0,48 | 74 | 0,33 |
| 5 | 120 m (utuh) | 1 | 133 | 39 | 0 | 5 | 12 | 25,8 | 0,55 | 22,1 | 6,1 | 5,1 | 22,1 | 18 | 0,53 | 60 | 0,77 |
| 6 | 120 m (utuh) | 1 | 57 | 35 | 0 | 2 | 3 | 19,6 | 0,37 | 19,6 | 9,0 | 4,2 | 18,2 | 16 | 0,43 | 0 | 0,29 |
| 7 | 120 m (utuh) | 1 | 22 | 36 | 0 | 1 | 1 | 23,7 | 0,52 | 21,3 | 7,3 | 3,3 | 16,0 | – | 0,25 | 0 | 0,29 |
| 8 | 120 m (utuh) | 1 | 1 | 35 | 0 | 1 | 0 | 12,0 | 0,27 | 10,8 | 0,5 | – | – | – | – | 4607 | 0,34 |
| 9 | 120 m (utuh) | 1 | 25 | 33 | 0 | 1 | 1 | 21,7 | 0,29 | 32,5 | 5,7 | 4,5 | – | 76 | 0,42 | 0 | 0,37 |
| 10 | 20 m | 2 | 41 | 32 | 0 | 2 | 3 | 22,8 | 0,49 | 21,0 | 5,2 | 6,2 | 22,0 | 16 | 0,26 | 463 | 0,28 |
| 11 | 120 m (utuh) | 1 | 77 | 38 | 1 | 6 | 5 | 24,1 | 0,47 | 24,9 | 5,3 | 4,8 | 22,0 | 16 | 0,34 | 106 | 0,43 |
| 12 | 120 m (utuh) | 1 | 6 | 32 | 1 | 5 | 2 | 24,0 | 0,66 | 17,4 | 1,5 | 5,2 | – | 72 | 0,40 | 830 | 0,46 |
| 13 | 120 m (utuh) | 1 | 20 | 35 | 0 | 1 | 3 | 17,6 | 0,32 | 20,3 | 6,7 | 4,7 | 18,7 | 16 | 0,46 | 1976 | 0,43 |
| 14 | 120 m (utuh) | 1 | 24 | 39 | 1 | 7 | 2 | 30,2 | 0,68 | 24,2 | 5,3 | 5,0 | 16,9 | – | 0,33 | 113 | 0,24 |
| 15 | 15 m | 2 | 21 | 30 | 0 | 3 | 1 | 21,4 | 0,58 | 13,6 | 9,2 | 4,2 | 18,5 | – | 0,47 | 246 | 0,23 |
| 16 | 120 m (utuh) | 1 | 67 | 39 | 0 | 1 | 8 | 24,0 | 0,55 | 20,5 | 4,1 | 5,7 | 21,2 | 16 | 0,72 | 510 | 0,53 |
| **Median** | 120 m | | 24 | 35 | 0 | 2 | 3 | 24,0 | 0,51 | 21,1 | 5,3 | 4,8 | 19,3 | 16 | 0,43 | 110 | 0,33 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **12 dari 16** dunia (selang kepercayaan 95%: 51–90%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 24,0 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,51 | 0,44–0,73 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 21,1 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 16 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 5,3 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 4,8 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 19,3 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,43 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian bayi (q0) | 0,000 |
| Rasio kelamin (♂ per 100 ♀) | 73 |
| Anggota per rumah | 9,5 |
| Pembunuhan per 100.000 tahun-orang | 110 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian |
| --- | ---: | ---: |
| Kelaparan | 898 | 62% |
| Kehausan | 271 | 19% |
| Usia tua | 176 | 12% |
| Dibunuh | 95 | 7% |
| Diterkam hewan | 0 | 0% |

Catatan: simulasi belum punya penyakit (Fase 3), sedangkan di masyarakat nyata penyakit menyebabkan lebih dari separuh kematian. Perbedaan ini temuan, bukan galat.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | – | – | 0 | 6 | 19,5 | 0,67 |
| 2 | – | – | 0 | 13 | 18,6 | 0,67 |
| 3 | – | – | 0 | 22 | 20,2 | 0,49 |
| 4 | – | – | 0 | 14 | 20,0 | 0,35 |
| 5 | – | – | 0 | 17 | 19,4 | 0,47 |
| 6 | – | – | 0 | 10 | 19,2 | 0,43 |
| 7 | – | – | 0 | 10 | 19,4 | 0,33 |
| 8 | – | – | 0 | 8 | 20,0 | 0,35 |
| 9 | – | – | 0 | 14 | 19,2 | 0,14 |
| 10 | – | – | 0 | 5 | 18,6 | 0,43 |
| 11 | 26 | – | 1 | 22 | 19,0 | 0,55 |
| 12 | 3 | – | 1 | 10 | 20,7 | 0,24 |
| 13 | – | – | 0 | 10 | 20,9 | 0,00 |
| 14 | 10 | – | 1 | 9 | 20,5 | 0,41 |
| 15 | – | – | 0 | 9 | 18,0 | 0,31 |
| 16 | – | – | 0 | 16 | 20,3 | 0,00 |
| **Median** | 10 | – | 0 | 10 | 19,4 | 0,38 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,40 | 357 |
| 15–29 | 0,41 | 134 |
| 30–44 | 0,40 | 74 |
| 45–59 | 0,39 | 78 |
| 60+ | 0,42 | 29 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 20,0 |
| 10–19 | 20,1 |
| 20–29 | 19,8 |
| 30–39 | 19,7 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 33 | 0 | – | 0 | 0,99 | 27 / 186 / 35 / 86 / 0 | 0 | 29 / 28 |
| 2 | 39 | 0 | – | 0 | 0,99 | 153 / 198 / 2 / 1 / 5 | 0 | 50 / 49 |
| 3 | 93 | 0 | – | 0 | 0,98 | 141 / 171 / 29 / 3 / 0 | 0 | 47 / 46 |
| 4 | 91 | 0 | – | 0 | 0,97 | 92 / 116 / 0 / 32 / 2 | 0 | 31 / 30 |
| 5 | 153 | 0 | – | 0 | 0,96 | 132 / 179 / 11 / 21 / 0 | 0 | 22 / 21 |
| 6 | 57 | 0 | – | 0 | 0,98 | 80 / 200 / 23 / 29 / 7 | 0 | 32 / 32 |
| 7 | 50 | 0 | – | 0 | 0,99 | 137 / 159 / 57 / 31 / 5 | 0 | 30 / 30 |
| 8 | 48 | 0 | – | 0 | 1,00 | 117 / 161 / 135 / 11 / 0 | 0 | 27 / 26 |
| 9 | 50 | 0 | – | 0 | 1,00 | 145 / 174 / 17 / 7 / 8 | 0 | 45 / 45 |
| 10 | 41 | 0 | – | 0 | 0,99 | 161 / 178 / 4 / 24 / 0 | 0 | 34 / 33 |
| 11 | 87 | 0 | – | 0 | 0,97 | 126 / 204 / 65 / 0 / 12 | 0 | 38 / 37 |
| 12 | 71 | 0 | – | 0 | 1,00 | 115 / 158 / 45 / 12 / 5 | 0 | 37 / 37 |
| 13 | 72 | 0 | – | 0 | 0,99 | 113 / 196 / 18 / 1 / 0 | 0 | 46 / 45 |
| 14 | 28 | 0 | – | 0 | 0,99 | 85 / 191 / 50 / 23 / 7 | 0 | 34 / 34 |
| 15 | 25 | 0 | – | 0 | 0,99 | 93 / 192 / 18 / 23 / 7 | 0 | 39 / 39 |
| 16 | 85 | 0 | – | 0 | 0,99 | 111 / 128 / 68 / 10 / 1 | 0 | 22 / 22 |
| **Median** | 54 | 0 | – | 0 | 0,99 | | | 34 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 16/16 |
| Babi Hutan | 100% | 100% | 16/16 |
| Ayam Hutan | 95% | 76% | 15/16 |
| Kerbau Liar | 83% | 57% | 15/16 |
| Harimau | 52% | 44% | 10/16 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 100% | 0% | 0% | 0% | 0,00 |
| 2 | 99% | 1% | 0% | 0% | 0,00 |
| 3 | 100% | 0% | 0% | 0% | 0,00 |
| 4 | 99% | 1% | 0% | 0% | 0,00 |
| 5 | 100% | 0% | 0% | 0% | 0,00 |
| 6 | 96% | 0% | 3% | 0% | 0,00 |
| 7 | 100% | 0% | 0% | 0% | 0,00 |
| 8 | 100% | 0% | 0% | 0% | 0,00 |
| 9 | 100% | 0% | 0% | 0% | 0,00 |
| 10 | 94% | 0% | 6% | 0% | 0,00 |
| 11 | 100% | 0% | 0% | 0% | 0,00 |
| 12 | 100% | 0% | 0% | 0% | 0,00 |
| 13 | 100% | 0% | 0% | 0% | 0,00 |
| 14 | 94% | 1% | 6% | 0% | 0,00 |
| 15 | 92% | 0% | 8% | 0% | 0,00 |
| 16 | 100% | 0% | 0% | 0% | 0,00 |
| **Median** | | 0% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 1512 | 26,2 | 31,0 | 42,2 | 0,0 |
| Netral | 5375 | 24,8 | 24,2 | 41,7 | 0,0 |
| La Niña | 2459 | 23,2 | 22,4 | 42,0 | 0,0 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 1467 | ×1,25 | ×1,49 | 37% |
| La Niña | 2372 | ×1,16 | ×1,12 | 27% |
