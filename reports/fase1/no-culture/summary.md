# Laporan soak — fase1 tanpa budaya per orang (plastisitas saja)

- Dibuat: 2026-10-05 22:10:50
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: culture
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 120 m (utuh) | 1 | 145 | 31 | 1 | 6 | 8 | 38,2 | 0,85 | 28,2 | 3,2 | 6,0 | 30,2 | 16 | 0,52 | 493 | 1,24 |
| 2 | 120 m (utuh) | 1 | 145 | 35 | 1 | 3 | 14 | 47,5 | 0,91 | 36,4 | 1,4 | 6,4 | 38,0 | 66 | 0,53 | 205 | 1,21 |
| 3 | 10 m | 2 | 144 | 28 | 1 | 6 | 18 | 46,9 | 0,86 | 37,9 | 1,0 | 5,3 | 49,0 | 76 | 0,46 | 259 | 1,18 |
| 4 | 16 m | 4 | 147 | 24 | 1 | 5 | 10 | 51,7 | 0,95 | 38,9 | 1,5 | 5,1 | 36,7 | 72 | 0,43 | 123 | 0,61 |
| 5 | 120 m (utuh) | 1 | 146 | 36 | 1 | 6 | 12 | 55,7 | 0,95 | 42,9 | 0,6 | 5,1 | 50,6 | 66 | 0,50 | 191 | 1,10 |
| 6 | 120 m (utuh) | 1 | 148 | 35 | 1 | 8 | 4 | 52,3 | 0,95 | 39,3 | 0,8 | 5,8 | 47,3 | 72 | 0,60 | 354 | 1,14 |
| 7 | 120 m (utuh) | 1 | 147 | 32 | 1 | 7 | 7 | 43,5 | 0,89 | 32,3 | 2,6 | 5,7 | 33,0 | 64 | 0,63 | 14 | 1,11 |
| 8 | 120 m (utuh) | 1 | 145 | 31 | 1 | 6 | 17 | 51,1 | 0,92 | 39,4 | 0,9 | 5,4 | 46,4 | 26 | 0,45 | 191 | 1,17 |
| 9 | 120 m (utuh) | 1 | 141 | 38 | 1 | 6 | 24 | 47,6 | 0,92 | 35,7 | 1,4 | 5,9 | 40,9 | 62 | 0,46 | 218 | 1,15 |
| 10 | 12 m | 3 | 148 | 30 | 1 | 6 | 18 | 46,9 | 0,92 | 35,0 | 2,0 | 6,2 | 38,8 | 74 | 0,45 | 82 | 0,94 |
| 11 | 120 m (utuh) | 1 | 149 | 29 | 1 | 6 | 12 | 57,3 | 0,94 | 45,2 | 0,5 | 5,7 | 56,8 | 76 | 0,49 | 136 | 1,24 |
| 12 | 10 m | 4 | 146 | 25 | 1 | 6 | 13 | 59,1 | 0,98 | 44,9 | 0,1 | 5,3 | 53,0 | 68 | 0,59 | 14 | 0,89 |
| 13 | 120 m (utuh) | 1 | 149 | 31 | 1 | 6 | 20 | 52,2 | 0,91 | 41,5 | 0,7 | 5,8 | 51,3 | 80 | 0,40 | 436 | 1,23 |
| 14 | 7 m | 4 | 145 | 27 | 1 | 6 | 4 | 43,1 | 0,85 | 33,7 | 2,1 | 6,5 | 36,4 | 72 | 0,60 | 82 | 0,89 |
| 15 | 120 m (utuh) | 1 | 147 | 32 | 1 | 6 | 11 | 50,9 | 0,86 | 42,1 | 0,3 | 5,6 | 50,7 | 72 | 0,52 | 191 | 1,28 |
| 16 | 120 m (utuh) | 1 | 144 | 36 | 1 | 6 | 11 | 46,1 | 0,85 | 37,0 | 1,4 | 7,1 | 38,9 | 68 | 0,48 | 82 | 1,19 |
| **Median** | 120 m | | 146 | 31 | 1 | 6 | 12 | 49,2 | 0,92 | 38,4 | 1,2 | 5,8 | 43,6 | 70 | 0,50 | 191 | 1,16 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **11 dari 16** dunia (selang kepercayaan 95%: 44–86%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 49,2 tahun | 21–37 | ↑ di atas | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,92 | 0,44–0,73 | ↑ di atas | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 38,4 tahun | 28–43 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 70 tahun | 68–78 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 1,2 anak | 5–7 | ↓ di bawah | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 5,8 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 43,6 tahun | 18–20 | ↑ di atas | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,50 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian bayi (q0) | 0,000 |
| Rasio kelamin (♂ per 100 ♀) | 104 |
| Anggota per rumah | 6,4 |
| Pembunuhan per 100.000 tahun-orang | 191 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian |
| --- | ---: | ---: |
| Kelaparan | 819 | 34% |
| Kehausan | 380 | 16% |
| Usia tua | 1018 | 42% |
| Dibunuh | 225 | 9% |

Catatan: simulasi belum punya penyakit (Fase 3), sedangkan di masyarakat nyata penyakit menyebabkan lebih dari separuh kematian. Perbedaan ini temuan, bukan galat.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 8 | – | 1 | 0 | 19,9 | 0,00 |
| 2 | 7 | – | 1 | 0 | 19,4 | 0,00 |
| 3 | 7 | – | 1 | 0 | 20,0 | 0,00 |
| 4 | 22 | – | 1 | 0 | 19,8 | 0,00 |
| 5 | 23 | – | 1 | 0 | 21,3 | 0,00 |
| 6 | 6 | – | 1 | 0 | 19,3 | 0,00 |
| 7 | 6 | – | 1 | 0 | 19,9 | 0,00 |
| 8 | 6 | – | 1 | 0 | 20,1 | 0,00 |
| 9 | 8 | – | 1 | 0 | 19,7 | 0,00 |
| 10 | 2 | – | 1 | 0 | 19,8 | 0,00 |
| 11 | 6 | – | 1 | 0 | 20,0 | 0,00 |
| 12 | 11 | – | 1 | 0 | 19,2 | 0,00 |
| 13 | 5 | – | 1 | 0 | 19,7 | 0,00 |
| 14 | 7 | – | 1 | 0 | 19,9 | 0,00 |
| 15 | 5 | – | 1 | 0 | 19,3 | 0,00 |
| 16 | 8 | – | 1 | 0 | 20,3 | 0,00 |
| **Median** | 7 | – | 1 | 0 | 19,9 | 0,00 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,00 | 719 |
| 15–29 | 0,00 | 572 |
| 30–44 | 0,00 | 470 |
| 45–59 | 0,00 | 344 |
| 60+ | 0,00 | 231 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 20,0 |
| 10–19 | 20,0 |
| 20–29 | 19,9 |
| 30–39 | 20,1 |
