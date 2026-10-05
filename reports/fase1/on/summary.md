# Laporan soak — fase1-on (belajar menyala)

- Dibuat: 2026-10-05 22:01:09
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 120 m (utuh) | 1 | 146 | 32 | 1 | 6 | 20 | 54,7 | 0,95 | 41,7 | 0,5 | 6,6 | 52,2 | 72 | 0,59 | 55 | 1,86 |
| 2 | 120 m (utuh) | 1 | 147 | 36 | 0 | 5 | 16 | 47,4 | 0,94 | 34,5 | 2,1 | 6,8 | 35,4 | 62 | 0,57 | 96 | 1,71 |
| 3 | 10 m | 2 | 146 | 34 | 1 | 6 | 27 | 43,5 | 0,91 | 31,7 | 2,4 | 7,4 | 28,4 | 26 | 0,51 | 151 | 1,68 |
| 4 | 26 m | 3 | 147 | 27 | 0 | 5 | 16 | 47,9 | 0,90 | 36,9 | 1,3 | 8,8 | 43,6 | 70 | 0,42 | 218 | 1,28 |
| 5 | 10 m | 4 | 146 | 26 | 0 | 5 | 25 | 42,6 | 0,93 | 30,1 | 2,3 | 9,4 | 30,3 | 68 | 0,42 | 68 | 1,07 |
| 6 | 120 m (utuh) | 1 | 148 | 39 | 0 | 4 | 18 | 47,2 | 0,94 | 34,7 | 1,8 | 9,0 | 33,4 | 26 | 0,51 | 27 | 1,47 |
| 7 | 120 m (utuh) | 1 | 148 | 34 | 1 | 4 | 7 | 48,5 | 0,94 | 35,8 | 1,2 | 5,1 | 44,9 | 72 | 0,65 | 14 | 1,97 |
| 8 | 120 m (utuh) | 1 | 146 | 40 | 0 | 5 | 24 | 42,0 | 0,94 | 29,1 | 3,0 | 7,5 | 26,0 | 26 | 0,54 | 150 | 1,62 |
| 9 | 120 m (utuh) | 1 | 146 | 36 | 0 | 3 | 21 | 46,9 | 0,93 | 34,7 | 1,1 | 6,3 | 38,9 | 66 | 0,55 | 27 | 1,98 |
| 10 | 12 m | 2 | 148 | 34 | 1 | 6 | 23 | 45,9 | 0,94 | 33,3 | 1,8 | 9,8 | 33,3 | 26 | 0,47 | 150 | 1,24 |
| 11 | 120 m (utuh) | 1 | 145 | 34 | 0 | 5 | 14 | 47,2 | 0,92 | 35,2 | 1,2 | 7,6 | 39,8 | 26 | 0,53 | 150 | 1,96 |
| 12 | 120 m (utuh) | 1 | 148 | 40 | 0 | 5 | 20 | 40,7 | 0,92 | 28,5 | 2,6 | 6,5 | 24,8 | 26 | 0,47 | 137 | 1,40 |
| 13 | 120 m (utuh) | 1 | 146 | 29 | 0 | 3 | 15 | 44,2 | 0,86 | 34,7 | 1,0 | 4,8 | 50,4 | 82 | 0,55 | 355 | 2,15 |
| 14 | 7 m | 2 | 147 | 40 | 1 | 4 | 13 | 45,7 | 0,98 | 31,4 | 2,0 | 8,5 | 32,2 | 26 | 0,38 | 109 | 1,37 |
| 15 | 120 m (utuh) | 1 | 145 | 34 | 1 | 6 | 30 | 48,3 | 0,94 | 35,8 | 1,8 | 8,3 | 34,4 | 26 | 0,45 | 14 | 1,57 |
| 16 | 120 m (utuh) | 1 | 146 | 34 | 1 | 6 | 13 | 57,7 | 0,96 | 44,7 | 0,7 | 6,8 | 49,1 | 74 | 0,45 | 41 | 1,96 |
| **Median** | 120 m | | 146 | 34 | 0 | 5 | 19 | 47,0 | 0,94 | 34,7 | 1,8 | 7,4 | 34,9 | 44 | 0,51 | 102 | 1,65 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **11 dari 16** dunia (selang kepercayaan 95%: 44–86%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 47,0 tahun | 21–37 | ↑ di atas | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,94 | 0,44–0,73 | ↑ di atas | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 34,7 tahun | 28–43 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 44 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 1,8 anak | 5–7 | ↓ di bawah | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 7,4 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 34,9 tahun | 18–20 | ↑ di atas | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,51 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian bayi (q0) | 0,000 |
| Rasio kelamin (♂ per 100 ♀) | 93 |
| Anggota per rumah | 6,7 |
| Pembunuhan per 100.000 tahun-orang | 102 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian |
| --- | ---: | ---: |
| Kelaparan | 853 | 34% |
| Kehausan | 686 | 27% |
| Usia tua | 860 | 34% |
| Dibunuh | 129 | 5% |

Catatan: simulasi belum punya penyakit (Fase 3), sedangkan di masyarakat nyata penyakit menyebabkan lebih dari separuh kematian. Perbedaan ini temuan, bukan galat.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 6 | – | 1 | 4 | 20,1 | 0,00 |
| 2 | – | – | 0 | 41 | 19,7 | 0,48 |
| 3 | 10 | – | 1 | 50 | 19,8 | 0,54 |
| 4 | – | – | 0 | 48 | 20,0 | 0,46 |
| 5 | – | – | 0 | 32 | 19,9 | 0,46 |
| 6 | – | – | 0 | 32 | 20,5 | 0,33 |
| 7 | 28 | – | 1 | 47 | 19,0 | 0,31 |
| 8 | – | – | 0 | 34 | 19,3 | 0,41 |
| 9 | – | – | 0 | 30 | 19,5 | 0,35 |
| 10 | 6 | – | 1 | 19 | 20,2 | 0,01 |
| 11 | – | – | 0 | 23 | 20,8 | 0,61 |
| 12 | – | – | 0 | 39 | 19,8 | 0,48 |
| 13 | – | – | 0 | 32 | 19,7 | 0,53 |
| 14 | 15 | – | 1 | 29 | 21,3 | 0,46 |
| 15 | 6 | – | 1 | 41 | 18,2 | 0,04 |
| 16 | 17 | – | 1 | 26 | 20,5 | 0,38 |
| **Median** | 10 | – | 0 | 32 | 19,9 | 0,43 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,37 | 727 |
| 15–29 | 0,43 | 621 |
| 30–44 | 0,45 | 477 |
| 45–59 | 0,44 | 337 |
| 60+ | 0,44 | 183 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 20,0 |
| 10–19 | 19,9 |
| 20–29 | 19,8 |
| 30–39 | 20,0 |
