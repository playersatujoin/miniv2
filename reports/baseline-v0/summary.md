# Laporan soak — baseline-v0

- Dibuat: 2026-10-05 21:15:42
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 120 m (utuh) | 1 | 147 | 34 | 1 | 6 | 4 | 52,2 | 0,94 | 40,0 | 1,1 | 7,8 | 44,5 | 76 | 0,57 | 259 | 0,71 |
| 2 | 120 m (utuh) | 1 | 147 | 36 | 1 | 4 | 14 | 51,4 | 0,91 | 40,7 | 0,6 | 5,7 | 49,6 | 64 | 0,60 | 272 | 0,73 |
| 3 | 15 m | 4 | 147 | 26 | 1 | 6 | 12 | 38,6 | 0,83 | 29,3 | 3,0 | 6,2 | 28,9 | 26 | 0,42 | 452 | 0,36 |
| 4 | 8 m | 2 | 149 | 30 | 1 | 6 | 7 | 49,9 | 0,95 | 37,0 | 0,6 | 5,2 | 48,5 | 66 | 0,53 | 136 | 0,81 |
| 5 | 38 m | 3 | 144 | 25 | 1 | 4 | 5 | 36,6 | 0,87 | 25,3 | 4,2 | 6,6 | 21,7 | 26 | 0,37 | 361 | 0,25 |
| 6 | 7 m | 3 | 144 | 35 | 1 | 6 | 7 | 42,4 | 0,95 | 29,1 | 2,4 | 5,6 | 35,3 | 26 | 0,54 | 218 | 0,46 |
| 7 | 13 m | 2 | 149 | 32 | 1 | 6 | 11 | 41,0 | 0,85 | 31,6 | 1,6 | 6,9 | 36,1 | 26 | 0,50 | 589 | 0,69 |
| 8 | 120 m (utuh) | 1 | 144 | 33 | 1 | 6 | 12 | 54,7 | 0,98 | 40,7 | 0,4 | 4,9 | 50,7 | 70 | 0,62 | 109 | 0,86 |
| **Median** | 26 m | | 147 | 32 | 1 | 6 | 9 | 46,1 | 0,92 | 34,3 | 1,4 | 5,9 | 40,3 | 45 | 0,54 | 266 | 0,70 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **3 dari 8** dunia (selang kepercayaan 95%: 14–69%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 46,1 tahun | 21–37 | ↑ di atas | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,92 | 0,44–0,73 | ↑ di atas | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 34,3 tahun | 28–43 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 45 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 1,4 anak | 5–7 | ↓ di bawah | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 5,9 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 40,3 tahun | 18–20 | ↑ di atas | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,54 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian bayi (q0) | 0,000 |
| Rasio kelamin (♂ per 100 ♀) | 105 |
| Anggota per rumah | 7,8 |
| Pembunuhan per 100.000 tahun-orang | 266 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian |
| --- | ---: | ---: |
| Kelaparan | 358 | 28% |
| Kehausan | 282 | 22% |
| Usia tua | 454 | 36% |
| Dibunuh | 175 | 14% |

Catatan: simulasi belum punya penyakit (Fase 3), sedangkan di masyarakat nyata penyakit menyebabkan lebih dari separuh kematian. Perbedaan ini temuan, bukan galat.
