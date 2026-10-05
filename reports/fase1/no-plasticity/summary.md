# Laporan soak — fase1 tanpa plastisitas (budaya saja)

- Dibuat: 2026-10-05 22:07:45
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: plasticity
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 120 m (utuh) | 1 | 148 | 38 | 1 | 5 | 26 | 43,4 | 0,94 | 30,6 | 2,5 | 6,9 | 30,7 | 26 | 0,46 | 96 | 1,21 |
| 2 | 30 m | 2 | 145 | 33 | 0 | 3 | 14 | 35,0 | 0,90 | 22,6 | 3,2 | 5,7 | 22,5 | 26 | 0,43 | 238 | 0,94 |
| 3 | 18 m | 6 | 141 | 19 | 0 | 2 | 3 | 39,3 | 0,88 | 28,3 | 3,2 | 6,4 | 24,1 | 26 | 0,43 | 370 | 0,42 |
| 4 | 120 m (utuh) | 1 | 148 | 38 | 0 | 2 | 10 | 38,7 | 0,92 | 26,2 | 2,9 | 6,8 | 30,3 | 26 | 0,44 | 301 | 1,07 |
| 5 | 12 m | 3 | 148 | 34 | 0 | 5 | 11 | 32,7 | 0,80 | 23,7 | 4,0 | 6,3 | 22,0 | 26 | 0,51 | 586 | 1,15 |
| 6 | 16 m | 5 | 146 | 26 | 0 | 5 | 9 | 45,1 | 0,96 | 31,5 | 2,2 | 8,5 | 31,9 | 26 | 0,48 | 96 | 0,61 |
| 7 | 5 m | 4 | 146 | 36 | 0 | 4 | 19 | 45,8 | 0,84 | 37,1 | 1,6 | 8,8 | 35,6 | 26 | 0,37 | 68 | 0,77 |
| 8 | 120 m (utuh) | 1 | 144 | 32 | 1 | 6 | 23 | 54,8 | 0,92 | 43,5 | 0,8 | 6,9 | 41,7 | 76 | 0,47 | 82 | 1,45 |
| 9 | 70 m | 2 | 147 | 16 | 0 | 5 | 10 | 45,7 | 0,89 | 35,1 | 0,9 | 5,6 | 47,0 | 64 | 0,47 | 14 | 0,84 |
| 10 | 13 m | 3 | 143 | 32 | 1 | 2 | 14 | 40,8 | 0,90 | 29,1 | 3,3 | 5,8 | 28,1 | 26 | 0,46 | 0 | 0,72 |
| 11 | 120 m (utuh) | 1 | 139 | 44 | 1 | 6 | 6 | 34,7 | 0,84 | 24,4 | 4,1 | 5,6 | 22,0 | 26 | 0,59 | 207 | 0,77 |
| 12 | 10 m | 8 | 90 | 12 | 0 | 1 | 5 | 35,5 | 0,92 | 22,5 | 4,2 | 6,2 | 23,4 | 26 | 0,55 | 0 | 0,21 |
| 13 | 11 m | 7 | 1 | 7 | 0 | 2 | 1 | – | – | – | – | – | – | – | – | – | 0,16 |
| 14 | 120 m (utuh) | 1 | 145 | 41 | 0 | 5 | 19 | 47,6 | 0,93 | 35,5 | 1,8 | 9,2 | 34,2 | 64 | 0,52 | 96 | 1,13 |
| 15 | 120 m (utuh) | 1 | 148 | 42 | 0 | 5 | 16 | 37,7 | 0,93 | 24,6 | 2,9 | 7,1 | 24,6 | 26 | 0,52 | 220 | 1,07 |
| 16 | 120 m (utuh) | 1 | 145 | 43 | 0 | 2 | 17 | 37,3 | 0,85 | 27,2 | 2,8 | 7,3 | 27,4 | 26 | 0,44 | 425 | 1,12 |
| **Median** | 50 m | | 145 | 34 | 0 | 4 | 12 | 39,3 | 0,90 | 28,3 | 2,9 | 6,8 | 28,1 | 26 | 0,47 | 96 | 0,89 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **7 dari 16** dunia (selang kepercayaan 95%: 23–67%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 39,3 tahun | 21–37 | ↑ di atas | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,90 | 0,44–0,73 | ↑ di atas | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 28,3 tahun | 28–43 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 26 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 2,9 anak | 5–7 | ↓ di bawah | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 6,8 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 28,1 tahun | 18–20 | ↑ di atas | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,47 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian bayi (q0) | 0,000 |
| Rasio kelamin (♂ per 100 ♀) | 106 |
| Anggota per rumah | 7,5 |
| Pembunuhan per 100.000 tahun-orang | 96 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian |
| --- | ---: | ---: |
| Kelaparan | 953 | 37% |
| Kehausan | 822 | 32% |
| Usia tua | 607 | 23% |
| Dibunuh | 203 | 8% |

Catatan: simulasi belum punya penyakit (Fase 3), sedangkan di masyarakat nyata penyakit menyebabkan lebih dari separuh kematian. Perbedaan ini temuan, bukan galat.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 3 | – | 1 | 37 | 19,4 | 0,45 |
| 2 | – | – | 0 | 37 | 20,3 | 0,51 |
| 3 | – | – | 0 | 23 | 19,6 | 0,14 |
| 4 | – | – | 0 | 26 | 21,1 | 0,53 |
| 5 | – | – | 0 | 38 | 18,9 | 0,52 |
| 6 | – | – | 0 | 38 | 19,7 | 0,04 |
| 7 | – | – | 0 | 45 | 20,1 | 0,22 |
| 8 | 8 | – | 1 | 58 | 20,3 | 0,62 |
| 9 | – | – | 0 | 39 | 19,5 | 0,65 |
| 10 | 28 | – | 1 | 33 | 19,8 | 0,39 |
| 11 | 20 | – | 1 | 30 | 20,4 | 0,59 |
| 12 | – | – | 0 | 31 | 19,6 | 0,01 |
| 13 | – | – | 0 | 25 | 20,0 | 0,00 |
| 14 | – | – | 0 | 39 | 19,1 | 0,56 |
| 15 | – | – | 0 | 26 | 19,9 | 0,46 |
| 16 | – | – | 0 | 32 | 19,6 | 0,43 |
| **Median** | 14 | – | 0 | 35 | 19,8 | 0,45 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,41 | 745 |
| 15–29 | 0,44 | 597 |
| 30–44 | 0,46 | 365 |
| 45–59 | 0,46 | 297 |
| 60+ | 0,47 | 120 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 20,0 |
| 10–19 | 20,0 |
| 20–29 | 19,9 |
| 30–39 | 19,9 |
| 40–49 | 20,0 |
