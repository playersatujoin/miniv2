# Laporan soak — Fase 2 — tanpa pertanian

- Dibuat: 2026-10-06 11:53:19
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: farming
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 26 m | 5 | 7 | 4 | 0 | 2 | 1 | 20,8 | 0,59 | 13,8 | – | 4,0 | 16,4 | – | 0,24 | 821 | 0,27 |
| 2 | 120 m (utuh) | 1 | 28 | 41 | 1 | 2 | 2 | 23,6 | 0,55 | 20,2 | 6,2 | 5,1 | 18,2 | 20 | 0,49 | 881 | 0,32 |
| 3 | 64 m | 3 | 2 | 12 | 0 | 2 | 2 | 10,2 | 0,16 | 21,1 | 5,3 | 4,8 | – | – | 0,00 | 4847 | 0,25 |
| 4 | 120 m (utuh) | 1 | 84 | 39 | 0 | 5 | 7 | 22,7 | 0,44 | 22,8 | 5,2 | 4,7 | 20,9 | 18 | 0,57 | 20 | 0,54 |
| 5 | 120 m (utuh) | 1 | 40 | 41 | 0 | 1 | 4 | 25,5 | 0,65 | 18,5 | 8,2 | 4,1 | 16,9 | 16 | 0,28 | 347 | 0,39 |
| 6 | 18 m | 2 | 17 | 34 | 0 | 3 | 1 | 19,6 | 0,43 | 15,4 | 6,2 | 4,0 | 20,8 | 20 | 0,12 | 0 | 0,25 |
| 7 | 96 m | 3 | 11 | 4 | 0 | 1 | 1 | 19,2 | 0,49 | 15,6 | 12,1 | 3,0 | – | – | 0,10 | 1642 | 0,33 |
| 8 | 120 m (utuh) | 1 | 1 | 35 | 0 | 1 | 1 | 12,0 | 0,24 | 13,1 | 5,0 | 3,3 | – | – | – | 3168 | 0,37 |
| 9 | 120 m (utuh) | 1 | 40 | 37 | 0 | 2 | 3 | 25,5 | 0,51 | 24,3 | 5,2 | 3,9 | 19,0 | 18 | 0,34 | 311 | 0,54 |
| 10 | 23 m | 2 | 47 | 27 | 0 | 1 | 1 | 25,2 | 0,49 | 24,4 | 5,0 | 6,4 | 21,1 | 16 | 0,38 | 150 | 0,40 |
| 11 | 98 m | 2 | 1 | 4 | 0 | 1 | 0 | – | – | – | – | – | – | – | – | – | 0,26 |
| 12 | 120 m (utuh) | 1 | 10 | 33 | 0 | 2 | 1 | 28,0 | 0,49 | 30,4 | 4,4 | 4,5 | – | – | 0,07 | 0 | 0,28 |
| 13 | 120 m (utuh) | 1 | 108 | 40 | 1 | 5 | 10 | 25,9 | 0,61 | 20,3 | 4,5 | 5,7 | 25,2 | 20 | 0,47 | 0 | 0,71 |
| 14 | 10 m | 5 | 34 | 24 | 0 | 3 | 2 | 21,6 | 0,59 | 14,1 | 11,5 | 3,3 | 16,3 | 18 | 0,41 | 0 | 0,28 |
| 15 | 16 m | 3 | 43 | 10 | 0 | 5 | 1 | 23,4 | 0,58 | 18,0 | 5,7 | 5,5 | 23,4 | 18 | 0,26 | 0 | 0,33 |
| 16 | 107 m | 3 | 21 | 2 | 0 | 5 | 1 | 31,5 | 0,43 | 42,9 | 11,2 | 3,1 | – | – | 0,41 | 0 | 0,34 |
| **Median** | 102 m | | 24 | 30 | 0 | 2 | 1 | 23,4 | 0,49 | 20,2 | 5,5 | 4,1 | 19,9 | 18 | 0,31 | 150 | 0,33 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **7 dari 16** dunia (selang kepercayaan 95%: 23–67%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 23,4 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,49 | 0,44–0,73 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 20,2 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 18 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 5,5 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 4,1 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 19,9 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,31 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian bayi (q0) | 0,000 |
| Rasio kelamin (♂ per 100 ♀) | 100 |
| Anggota per rumah | 9,5 |
| Pembunuhan per 100.000 tahun-orang | 150 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian |
| --- | ---: | ---: |
| Kelaparan | 639 | 58% |
| Kehausan | 252 | 23% |
| Usia tua | 132 | 12% |
| Dibunuh | 70 | 6% |
| Diterkam hewan | 0 | 0% |

Catatan: simulasi belum punya penyakit (Fase 3), sedangkan di masyarakat nyata penyakit menyebabkan lebih dari separuh kematian. Perbedaan ini temuan, bukan galat.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | – | – | 0 | 23 | 21,3 | 0,31 |
| 2 | 18 | – | 1 | 15 | 19,8 | 0,38 |
| 3 | – | – | 0 | 18 | 22,5 | 0,00 |
| 4 | – | – | 0 | 14 | 20,5 | 0,52 |
| 5 | – | – | 0 | 16 | 20,4 | 0,64 |
| 6 | – | – | 0 | 13 | 19,9 | 0,34 |
| 7 | – | – | 0 | 20 | 20,0 | 0,32 |
| 8 | – | – | 0 | 15 | 19,0 | 0,30 |
| 9 | – | – | 0 | 24 | 20,0 | 0,63 |
| 10 | – | – | 0 | 25 | 19,8 | 0,54 |
| 11 | – | – | 0 | 18 | 20,0 | 0,33 |
| 12 | – | – | 0 | 14 | 20,8 | 0,00 |
| 13 | 31 | – | 1 | 20 | 19,7 | 0,42 |
| 14 | – | – | 0 | 25 | 18,6 | 0,42 |
| 15 | – | – | 0 | 19 | 20,6 | 0,00 |
| 16 | – | – | 0 | 18 | 20,5 | 0,31 |
| **Median** | 24 | – | 0 | 18 | 20,0 | 0,34 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,34 | 272 |
| 15–29 | 0,37 | 91 |
| 30–44 | 0,39 | 55 |
| 45–59 | 0,35 | 53 |
| 60+ | 0,42 | 23 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 20,2 |
| 10–19 | 20,1 |
| 20–29 | 19,9 |
| 30–39 | 20,1 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 55 | 0 | – | 0 | 0,99 | 108 / 198 / 51 / 5 / 0 | 0 | 37 / 36 |
| 2 | 42 | 0 | – | 0 | 1,00 | 129 / 119 / 35 / 18 / 0 | 0 | 34 / 33 |
| 3 | 43 | 0 | – | 0 | 1,00 | 152 / 126 / 78 / 24 / 2 | 0 | 31 / 30 |
| 4 | 112 | 0 | – | 0 | 0,97 | 137 / 122 / 0 / 11 / 0 | 0 | 43 / 41 |
| 5 | 83 | 0 | – | 0 | 0,98 | 142 / 189 / 37 / 34 / 0 | 0 | 29 / 28 |
| 6 | 34 | 0 | – | 0 | 0,99 | 102 / 123 / 18 / 34 / 0 | 0 | 28 / 27 |
| 7 | 62 | 0 | – | 0 | 0,99 | 130 / 198 / 45 / 20 / 0 | 0 | 42 / 41 |
| 8 | 56 | 0 | – | 0 | 1,00 | 116 / 102 / 38 / 0 / 0 | 0 | 29 / 27 |
| 9 | 90 | 0 | – | 0 | 0,99 | 90 / 130 / 81 / 36 / 5 | 0 | 31 / 31 |
| 10 | 72 | 0 | – | 0 | 0,99 | 120 / 154 / 35 / 1 / 0 | 0 | 43 / 42 |
| 11 | 29 | 0 | – | 0 | 1,00 | 138 / 186 / 18 / 0 / 0 | 0 | 31 / 29 |
| 12 | 44 | 0 | – | 0 | 0,99 | 125 / 197 / 11 / 0 / 5 | 0 | 39 / 38 |
| 13 | 120 | 0 | – | 0 | 0,96 | 138 / 137 / 23 / 1 / 2 | 0 | 43 / 42 |
| 14 | 47 | 0 | – | 0 | 0,98 | 122 / 159 / 31 / 35 / 3 | 0 | 36 / 36 |
| 15 | 63 | 0 | – | 0 | 0,98 | 123 / 175 / 19 / 18 / 5 | 0 | 49 / 49 |
| 16 | 52 | 0 | – | 0 | 0,99 | 118 / 202 / 10 / 25 / 0 | 0 | 28 / 27 |
| **Median** | 56 | 0 | – | 0 | 0,99 | | | 35 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 16/16 |
| Babi Hutan | 100% | 100% | 16/16 |
| Ayam Hutan | 97% | 72% | 15/16 |
| Kerbau Liar | 86% | 52% | 13/16 |
| Harimau | 52% | 40% | 6/16 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 100% | 0% | 0% | 0% | 0,00 |
| 2 | 100% | 0% | 0% | 0% | 0,00 |
| 3 | 100% | 0% | 0% | 0% | 0,00 |
| 4 | 100% | 0% | 0% | 0% | 0,00 |
| 5 | 99% | 0% | 0% | 0% | 0,00 |
| 6 | 100% | 0% | 0% | 0% | 0,00 |
| 7 | 100% | 0% | 0% | 0% | 0,00 |
| 8 | 100% | 0% | 0% | 0% | 0,00 |
| 9 | 100% | 0% | 0% | 0% | 0,00 |
| 10 | 100% | 0% | 0% | 0% | 0,00 |
| 11 | 100% | 0% | 0% | 0% | 0,00 |
| 12 | 99% | 1% | 0% | 0% | 0,00 |
| 13 | 99% | 0% | 0% | 1% | 0,00 |
| 14 | 94% | 0% | 5% | 1% | 0,00 |
| 15 | 100% | 0% | 0% | 0% | 0,00 |
| 16 | 100% | 0% | 0% | 0% | 0,00 |
| **Median** | | 0% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 1417 | 24,4 | 30,4 | 38,9 | 0,0 |
| Netral | 4909 | 23,3 | 23,1 | 39,3 | 0,0 |
| La Niña | 2327 | 23,2 | 20,4 | 39,2 | 0,0 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 1381 | ×1,20 | ×1,42 | 37% |
| La Niña | 2256 | ×1,17 | ×1,06 | 25% |
