# Laporan soak — fase1-off (model lama: tanpa belajar, pengetahuan milik semua)

- Dibuat: 2026-10-05 22:04:06
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: learning
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 11 m | 3 | 146 | 30 | 1 | 6 | 12 | 35,3 | 0,82 | 26,0 | 3,6 | 6,4 | 27,3 | 26 | 0,50 | 384 | 1,05 |
| 2 | 120 m (utuh) | 1 | 145 | 39 | 1 | 6 | 13 | 43,2 | 0,89 | 32,3 | 2,5 | 6,3 | 32,2 | 26 | 0,52 | 260 | 0,90 |
| 3 | 10 m | 2 | 146 | 33 | 1 | 6 | 14 | 44,7 | 0,89 | 34,0 | 1,9 | 6,4 | 36,3 | 64 | 0,56 | 287 | 1,06 |
| 4 | 120 m (utuh) | 1 | 148 | 35 | 1 | 7 | 14 | 47,4 | 0,91 | 36,3 | 2,4 | 6,2 | 34,9 | 68 | 0,47 | 218 | 1,07 |
| 5 | 21 m | 3 | 149 | 25 | 1 | 6 | 13 | 46,9 | 0,87 | 37,5 | 1,1 | 4,4 | 52,3 | 74 | 0,55 | 396 | 1,01 |
| 6 | 120 m (utuh) | 1 | 148 | 34 | 1 | 7 | 11 | 47,8 | 0,93 | 35,7 | 1,3 | 6,4 | 40,4 | 26 | 0,39 | 382 | 1,03 |
| 7 | 5 m | 2 | 144 | 36 | 1 | 5 | 7 | 39,2 | 0,89 | 27,6 | 2,5 | 5,8 | 35,6 | 26 | 0,41 | 137 | 1,05 |
| 8 | 120 m (utuh) | 1 | 148 | 33 | 1 | 6 | 10 | 50,1 | 0,88 | 40,4 | 0,9 | 6,2 | 47,7 | 70 | 0,52 | 245 | 1,16 |
| 9 | 120 m (utuh) | 1 | 148 | 37 | 1 | 7 | 13 | 42,9 | 0,93 | 30,6 | 2,8 | 5,5 | 35,2 | 26 | 0,46 | 396 | 1,06 |
| 10 | 8 m | 2 | 147 | 35 | 1 | 6 | 9 | 47,8 | 0,91 | 36,7 | 1,2 | 7,5 | 42,8 | 74 | 0,41 | 68 | 1,03 |
| 11 | 120 m (utuh) | 1 | 147 | 33 | 1 | 6 | 15 | 56,4 | 0,92 | 45,6 | 0,1 | 5,6 | 60,2 | 78 | 0,46 | 177 | 1,11 |
| 12 | 120 m (utuh) | 1 | 142 | 35 | 1 | 6 | 11 | 40,4 | 0,89 | 29,0 | 2,8 | 6,3 | 33,4 | 26 | 0,52 | 123 | 1,05 |
| 13 | 10 m | 2 | 147 | 28 | 1 | 8 | 12 | 52,0 | 0,92 | 40,7 | 0,7 | 6,9 | 47,7 | 84 | 0,37 | 272 | 1,02 |
| 14 | 28 m | 2 | 145 | 26 | 1 | 6 | 15 | 47,2 | 0,92 | 35,3 | 1,4 | 6,0 | 41,5 | 62 | 0,52 | 232 | 0,92 |
| 15 | 120 m (utuh) | 1 | 146 | 36 | 1 | 6 | 15 | 42,7 | 0,85 | 33,2 | 1,4 | 6,2 | 37,6 | 72 | 0,60 | 191 | 1,13 |
| 16 | 120 m (utuh) | 1 | 147 | 33 | 1 | 6 | 14 | 53,3 | 0,93 | 41,5 | 1,0 | 5,4 | 46,9 | 70 | 0,46 | 164 | 1,16 |
| **Median** | 120 m | | 147 | 34 | 1 | 6 | 13 | 47,0 | 0,90 | 35,5 | 1,4 | 6,2 | 39,0 | 66 | 0,48 | 238 | 1,05 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **9 dari 16** dunia (selang kepercayaan 95%: 33–77%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 47,0 tahun | 21–37 | ↑ di atas | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,90 | 0,44–0,73 | ↑ di atas | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 35,5 tahun | 28–43 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 66 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 1,4 anak | 5–7 | ↓ di bawah | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 6,2 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 39,0 tahun | 18–20 | ↑ di atas | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,48 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian bayi (q0) | 0,000 |
| Rasio kelamin (♂ per 100 ♀) | 106 |
| Anggota per rumah | 8,9 |
| Pembunuhan per 100.000 tahun-orang | 238 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian |
| --- | ---: | ---: |
| Kelaparan | 938 | 36% |
| Kehausan | 514 | 20% |
| Usia tua | 870 | 33% |
| Dibunuh | 288 | 11% |

Catatan: simulasi belum punya penyakit (Fase 3), sedangkan di masyarakat nyata penyakit menyebabkan lebih dari separuh kematian. Perbedaan ini temuan, bukan galat.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 12 | – | 1 | 0 | 21,6 | 0,00 |
| 2 | 24 | – | 1 | 0 | 20,4 | 0,00 |
| 3 | 12 | – | 1 | 0 | 19,4 | 0,00 |
| 4 | 14 | – | 1 | 0 | 20,2 | 0,00 |
| 5 | 3 | – | 1 | 0 | 19,5 | 0,00 |
| 6 | 17 | – | 1 | 0 | 20,7 | 0,00 |
| 7 | 10 | – | 1 | 0 | 19,8 | 0,00 |
| 8 | 4 | – | 1 | 0 | 20,1 | 0,00 |
| 9 | 11 | – | 1 | 0 | 19,8 | 0,00 |
| 10 | 9 | – | 1 | 0 | 20,4 | 0,00 |
| 11 | 15 | – | 1 | 0 | 20,2 | 0,00 |
| 12 | 12 | – | 1 | 0 | 20,1 | 0,00 |
| 13 | 11 | – | 1 | 0 | 18,4 | 0,00 |
| 14 | 15 | – | 1 | 0 | 18,6 | 0,00 |
| 15 | 8 | – | 1 | 0 | 18,9 | 0,00 |
| 16 | 5 | – | 1 | 0 | 19,0 | 0,00 |
| **Median** | 12 | – | 1 | 0 | 20,0 | 0,00 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,00 | 706 |
| 15–29 | 0,00 | 633 |
| 30–44 | 0,00 | 512 |
| 45–59 | 0,00 | 324 |
| 60+ | 0,00 | 168 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 19,9 |
| 10–19 | 19,8 |
| 20–29 | 19,9 |
| 30–39 | 20,0 |
