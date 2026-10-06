# Laporan soak — f3b-on

- Dibuat: 2026-10-06 17:11:09
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16
- Durasi: 120 menit simulasi per dunia (≈ 900 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 120 m (utuh) | 1 | 23 | 40 | 0 | 1 | 2 | 39,3 | 0,85 | 30,5 | 7,2 | 3,3 | 17,8 | 28 | 0,61 | 98 | 0,43 |
| 2 | 33 m | 4 | 18 | 8 | 0 | 5 | 2 | 22,7 | 0,61 | 16,8 | 7,1 | 3,2 | 16,9 | 16 | 0,40 | 102 | 0,32 |
| 3 | 120 m (utuh) | 1 | 749 | 45 | 0 | 3 | 7 | 20,7 | 0,65 | 13,0 | 9,4 | 3,0 | 16,9 | 18 | 0,53 | 57 | 2,91 |
| 4 | 120 m (utuh) | 1 | 43 | 41 | 0 | 3 | 3 | 33,6 | 0,73 | 28,9 | 6,1 | 2,7 | 17,4 | 22 | 0,47 | 0 | 0,53 |
| 5 | 120 m (utuh) | 1 | 5 | 39 | 0 | 3 | 2 | 28,0 | 0,36 | 57,4 | – | – | – | – | 0,30 | 0 | 0,99 |
| 6 | 42 m | 5 | 23 | 13 | 0 | 2 | 4 | 26,9 | 0,73 | 19,3 | 6,4 | 3,9 | 18,6 | 18 | 0,35 | 155 | 0,33 |
| 7 | 120 m (utuh) | 1 | 20 | 40 | 0 | 1 | 3 | 22,5 | 0,60 | 17,9 | 5,2 | 4,3 | 20,3 | 18 | 0,21 | 251 | 0,46 |
| 8 | 29 m | 5 | 12 | 5 | 0 | 2 | 2 | 20,8 | 0,59 | 15,3 | 7,7 | 3,2 | 18,1 | – | 0,37 | 1189 | 0,33 |
| 9 | 120 m (utuh) | 1 | 101 | 43 | 0 | 1 | 1 | 21,8 | 0,67 | 14,5 | 6,5 | 3,6 | 17,9 | 18 | 0,46 | 123 | 0,70 |
| 10 | 13 m | 6 | 2 | 5 | 1 | 2 | 1 | 33,4 | 0,75 | 25,2 | – | – | – | – | – | 462 | 0,30 |
| 11 | 120 m (utuh) | 1 | 88 | 44 | 0 | 3 | 2 | 28,9 | 0,69 | 24,0 | 7,9 | 3,5 | 17,6 | 18 | 0,51 | 340 | 0,95 |
| 12 | 120 m (utuh) | 1 | 90 | 45 | 0 | 3 | 2 | 24,7 | 0,71 | 17,1 | 7,1 | 3,1 | 17,1 | 20 | 0,46 | 251 | 1,08 |
| 13 | 120 m (utuh) | 1 | 189 | 45 | 0 | 5 | 6 | 25,6 | 0,70 | 18,6 | 7,3 | 3,3 | 17,4 | 18 | 0,43 | 92 | 1,89 |
| 14 | 66 m | 3 | 13 | 6 | 0 | 5 | 3 | 19,6 | 0,50 | 19,5 | 9,7 | 3,1 | 17,8 | 20 | 0,36 | 1455 | 0,32 |
| 15 | 27 m | 6 | 4 | 1 | 0 | 2 | 1 | 38,4 | 0,87 | 27,2 | 2,8 | 2,4 | – | – | 0,17 | 0 | 0,30 |
| 16 | 20 m | 5 | 8 | 7 | 0 | 4 | 2 | 46,8 | 0,89 | 35,9 | 2,8 | 2,1 | – | – | 0,18 | 0 | 0,35 |
| **Median** | 120 m | | 22 | 40 | 0 | 3 | 2 | 26,2 | 0,69 | 19,4 | 7,1 | 3,2 | 17,7 | 18 | 0,40 | 112 | 0,45 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **9 dari 16** dunia (selang kepercayaan 95%: 33–77%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 26,2 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,69 | 0,44–0,73 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 19,4 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 18 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 7,1 anak | 5–7 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,2 tahun | 2,8–3,3 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 17,7 tahun | 18–20 | ↓ di bawah | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,40 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,074 | 0,13–0,41 | ↓ di bawah | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,135 |
| Diare per orang per tahun | 0,23 |
| Malaria per orang per tahun | 0,60 |
| ISPA per orang per tahun | 0,01 |
| Rasio kelamin (♂ per 100 ♀) | 103 |
| Anggota per rumah | 4,5 |
| Pembunuhan per 100.000 tahun-orang | 112 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 1067 | 38% | 0 | 0% |
| Kehausan | 752 | 27% | 6 | 1% |
| Usia tua | 225 | 8% | 0 | 0% |
| Dibunuh | 83 | 3% | 20 | 5% |
| Diterkam hewan | 0 | 0% | 0 | 0% |
| Melahirkan | 57 | 2% | 0 | 0% |
| Neonatal (minggu pertama) | 218 | 8% | 218 | 51% |
| Diare | 153 | 5% | 63 | 15% |
| Malaria | 205 | 7% | 104 | 24% |
| Radang paru (ISPA) | 31 | 1% | 17 | 4% |
| **Penyakit menular** | 389 | 14% | 184 | 43% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | – | – | 0 | 11 | 25,0 | 0,49 |
| 2 | – | – | 0 | 24 | 24,3 | 0,88 |
| 3 | – | – | 0 | 10 | 24,2 | 0,96 |
| 4 | – | – | 0 | 8 | 27,1 | 0,92 |
| 5 | – | – | 0 | 12 | 61,0 | 0,92 |
| 6 | – | – | 0 | 19 | 25,7 | 0,17 |
| 7 | – | – | 0 | 16 | 30,0 | 0,90 |
| 8 | – | – | 0 | 16 | 30,5 | 0,06 |
| 9 | – | – | 0 | 19 | 23,9 | 0,95 |
| 10 | 16 | – | 1 | 35 | 24,0 | 0,35 |
| 11 | – | – | 0 | 15 | 22,6 | 0,88 |
| 12 | – | – | 0 | 17 | 29,5 | 0,94 |
| 13 | – | – | 0 | 18 | 25,0 | 0,95 |
| 14 | – | – | 0 | 17 | 22,9 | 0,52 |
| 15 | – | – | 0 | 24 | 48,3 | 0,66 |
| 16 | – | – | 0 | 33 | 34,1 | 0,70 |
| **Median** | 16 | – | 0 | 17 | 25,4 | 0,88 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,87 | 802 |
| 15–29 | 0,91 | 348 |
| 30–44 | 0,90 | 135 |
| 45–59 | 0,91 | 67 |
| 60+ | 0,94 | 36 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 31,7 |
| 10–19 | 31,1 |
| 20–29 | 26,3 |
| 30–39 | 27,0 |
| 40–49 | 25,7 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 35 | 0 | 15 | 5 | 0,99 | 165 / 136 / 51 / 4 / 0 | 0 | 56 / 55 |
| 2 | 36 | 0 | 7 | 19 | 0,99 | 126 / 124 / 20 / 20 / 2 | 0 | 40 / 40 |
| 3 | 749 | 0 | 11 | 197 | 0,96 | 131 / 122 / 31 / 3 / 4 | 0 | 29 / 29 |
| 4 | 73 | 0 | 4 | 40 | 0,99 | 65 / 135 / 43 / 12 / 1 | 0 | 35 / 35 |
| 5 | 104 | 0 | 11 | 95 | 1,00 | 149 / 120 / 29 / 3 / 6 | 0 | 33 / 33 |
| 6 | 31 | 0 | 60 | 4 | 0,99 | 57 / 94 / 63 / 2 / 3 | 0 | 28 / 28 |
| 7 | 65 | 0 | 14 | 33 | 0,99 | 120 / 129 / 34 / 0 / 0 | 0 | 62 / 60 |
| 8 | 23 | 0 | 5 | 17 | 0,99 | 132 / 213 / 41 / 2 / 0 | 0 | 41 / 39 |
| 9 | 128 | 0 | 23 | 38 | 0,98 | 128 / 161 / 18 / 3 / 0 | 0 | 36 / 35 |
| 10 | 24 | 0 | 8 | 12 | 1,00 | 181 / 169 / 0 / 0 / 3 | 0 | 46 / 44 |
| 11 | 154 | 0 | 29 | 40 | 0,98 | 125 / 112 / 60 / 0 / 0 | 0 | 42 / 40 |
| 12 | 141 | 0 | 8 | 74 | 0,99 | 122 / 172 / 25 / 15 / 4 | 0 | 33 / 33 |
| 13 | 240 | 0 | 12 | 141 | 0,97 | 128 / 163 / 29 / 0 / 2 | 0 | 40 / 38 |
| 14 | 24 | 0 | 27 | 28 | 0,99 | 95 / 103 / 12 / 11 / 4 | 0 | 40 / 40 |
| 15 | 21 | 0 | 29 | 11 | 1,00 | 116 / 165 / 68 / 9 / 1 | 0 | 29 / 29 |
| 16 | 33 | 0 | 1 | 20 | 1,00 | 122 / 145 / 49 / 0 / 6 | 0 | 41 / 40 |
| **Median** | 50 | 0 | 12 | 30 | 0,99 | | | 40 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 16/16 |
| Babi Hutan | 100% | 100% | 16/16 |
| Ayam Hutan | 95% | 72% | 15/16 |
| Kerbau Liar | 76% | 42% | 11/16 |
| Harimau | 54% | 37% | 11/16 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 99% | 0% | 0% | 1% | 0,00 |
| 2 | 79% | 21% | 0% | 0% | 0,27 |
| 3 | 13% | 86% | 0% | 0% | 0,88 |
| 4 | 70% | 30% | 0% | 0% | 0,62 |
| 5 | 35% | 65% | 0% | 0% | 1,64 |
| 6 | 99% | 0% | 1% | 0% | 0,00 |
| 7 | 76% | 24% | 0% | 0% | 0,35 |
| 8 | 89% | 0% | 11% | 0% | 0,00 |
| 9 | 37% | 63% | 0% | 0% | 0,62 |
| 10 | 100% | 0% | 0% | 0% | 0,00 |
| 11 | 61% | 39% | 0% | 0% | 0,22 |
| 12 | 36% | 64% | 0% | 1% | 0,69 |
| 13 | 31% | 69% | 0% | 0% | 0,96 |
| 14 | 99% | 0% | 1% | 0% | 0,00 |
| 15 | 100% | 0% | 0% | 0% | 0,41 |
| 16 | 97% | 3% | 0% | 0% | 0,17 |
| **Median** | | 23% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 1091 | 16,5 | 17,0 | 44,2 | 72,3 |
| Netral | 3840 | 15,4 | 15,4 | 43,6 | 71,0 |
| La Niña | 1725 | 16,0 | 15,8 | 42,5 | 69,1 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 1006 | ×1,12 | ×1,33 | 28% |
| La Niña | 1557 | ×1,27 | ×1,13 | 24% |
