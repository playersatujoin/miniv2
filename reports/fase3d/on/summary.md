# Laporan soak — Fase 3d gizi aktif

- Dibuat: 2026-10-06 23:49:56
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 20 menit simulasi per dunia (≈ 150 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 20 m (utuh) | 1 | 79 | 8 | 1 | 3 | 3 | 28,7 | 0,71 | 23,1 | 4,8 | 4,1 | 20,2 | 20 | 0,38 | 25 | 0,30 |
| 2 | 20 m (utuh) | 1 | 63 | 8 | 1 | 2 | 7 | 31,0 | 0,77 | 24,2 | 6,3 | 3,4 | 19,6 | 20 | 0,37 | 46 | 0,24 |
| 3 | 20 m (utuh) | 1 | 68 | 8 | 1 | 2 | 1 | 27,6 | 0,75 | 20,7 | 6,3 | 4,2 | 18,8 | 20 | 0,39 | 110 | 0,32 |
| 4 | 20 m (utuh) | 1 | 75 | 8 | 1 | 6 | 5 | 28,9 | 0,70 | 23,9 | 4,4 | 4,5 | 19,9 | 20 | 0,34 | 66 | 0,46 |
| 5 | 20 m (utuh) | 1 | 56 | 7 | 0 | 1 | 3 | 36,7 | 0,75 | 31,7 | 5,6 | 3,7 | 20,2 | 24 | 0,27 | 178 | 0,24 |
| 6 | 20 m (utuh) | 1 | 63 | 8 | 0 | 3 | 4 | 24,7 | 0,73 | 17,5 | 6,1 | 3,9 | 18,7 | 18 | 0,34 | 78 | 0,27 |
| 7 | 20 m (utuh) | 1 | 152 | 8 | 1 | 2 | 2 | 41,2 | 0,78 | 36,5 | 6,5 | 3,5 | 18,1 | 18 | 0,44 | 306 | 0,46 |
| 8 | 20 m (utuh) | 1 | 99 | 9 | 1 | 3 | 8 | 26,7 | 0,63 | 24,1 | 7,5 | 3,7 | 18,8 | 20 | 0,36 | 70 | 0,34 |
| **Median** | 20 m | | 72 | 8 | 1 | 2 | 4 | 28,8 | 0,74 | 24,0 | 6,2 | 3,8 | 19,2 | 20 | 0,37 | 74 | 0,31 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **8 dari 8** dunia (selang kepercayaan 95%: 68–100%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 28,8 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,74 | 0,44–0,73 | ↑ di atas | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 24,0 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 20 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 6,2 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,8 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 19,2 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,37 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,123 | 0,13–0,41 | ↓ di bawah | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,176 |
| Diare per orang per tahun | 0,46 |
| Malaria per orang per tahun | 0,60 |
| ISPA per orang per tahun | 0,01 |
| Rasio kelamin (♂ per 100 ♀) | 105 |
| Anggota per rumah | 12,7 |
| Pembunuhan per 100.000 tahun-orang | 74 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 246 | 27% | 0 | 0% |
| Kehausan | 224 | 25% | 0 | 0% |
| Usia tua | 131 | 15% | 0 | 0% |
| Dibunuh | 33 | 4% | 8 | 4% |
| Diterkam hewan | 0 | 0% | 0 | 0% |
| Melahirkan | 20 | 2% | 0 | 0% |
| Neonatal (minggu pertama) | 101 | 11% | 101 | 54% |
| Diare | 73 | 8% | 39 | 21% |
| Malaria | 71 | 8% | 39 | 21% |
| Radang paru (ISPA) | 1 | 0% | 1 | 1% |
| Terbakar | 0 | 0% | 0 | 0% |
| Tenggelam | 0 | 0% | 0 | 0% |
| Terjatuh | 0 | 0% | 0 | 0% |
| **Penyakit menular** | 145 | 16% | 79 | 42% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 8 | – | 1 | 6 | 36,0 | 0,54 |
| 2 | 4 | – | 1 | 14 | 34,8 | 0,81 |
| 3 | 8 | – | 1 | 1 | 33,7 | 0,96 |
| 4 | 4 | – | 1 | 5 | 37,2 | 0,59 |
| 5 | – | – | 0 | 6 | 36,1 | 0,50 |
| 6 | – | – | 0 | 3 | 36,7 | 0,69 |
| 7 | 6 | – | 1 | 6 | 33,0 | 0,77 |
| 8 | 6 | – | 1 | 14 | 32,4 | 0,55 |
| **Median** | 6 | – | 1 | 6 | 35,4 | 0,64 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,57 | 304 |
| 15–29 | 0,59 | 164 |
| 30–44 | 0,62 | 96 |
| 45–59 | 0,61 | 54 |
| 60+ | 0,61 | 37 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 34,1 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 94 | 0 | – | 0 | 0,98 | 122 / 148 / 41 / 3 / 4 | 0 | 6 / 6 |
| 2 | 63 | 0 | 6 | 39 | 0,97 | 109 / 138 / 101 / 2 / 8 | 0 | 5 / 5 |
| 3 | 79 | 0 | 10 | 97 | 0,99 | 161 / 161 / 0 / 9 / 0 | 0 | 5 / 3 |
| 4 | 104 | 0 | 6 | 7 | 0,96 | 115 / 157 / 95 / 3 / 0 | 0 | 6 / 5 |
| 5 | 56 | 0 | 10 | 5 | 0,98 | 153 / 132 / 66 / 0 / 5 | 0 | 8 / 7 |
| 6 | 63 | 0 | 11 | 13 | 0,98 | 131 / 159 / 75 / 0 / 5 | 0 | 6 / 5 |
| 7 | 152 | 0 | 3 | 44 | 0,97 | 122 / 160 / 101 / 25 / 0 | 0 | 5 / 4 |
| 8 | 108 | 0 | 16 | 4 | 0,96 | 80 / 140 / 74 / 2 / 1 | 0 | 5 / 5 |
| **Median** | 86 | 0 | 10 | 10 | 0,98 | | | 6 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 100% | 8/8 |
| Ayam Hutan | 100% | 85% | 7/8 |
| Kerbau Liar | 90% | 55% | 6/8 |
| Harimau | 52% | 40% | 5/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 99% | 0% | 0% | 1% | 0,00 |
| 2 | 89% | 10% | 0% | 1% | 0,31 |
| 3 | 71% | 28% | 0% | 1% | 0,67 |
| 4 | 98% | 1% | 0% | 1% | 0,02 |
| 5 | 99% | 0% | 0% | 1% | 0,02 |
| 6 | 95% | 3% | 2% | 0% | 0,12 |
| 7 | 66% | 33% | 0% | 0% | 0,69 |
| 8 | 100% | 0% | 0% | 0% | 0,00 |
| **Median** | | 2% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 144 | 7,0 | 7,8 | 41,0 | 23,3 |
| Netral | 517 | 7,5 | 7,4 | 38,7 | 23,2 |
| La Niña | 212 | 6,8 | 6,7 | 39,1 | 20,4 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 124 | ×1,07 | ×1,12 | 22% |
| La Niña | 162 | ×0,98 | ×0,81 | 19% |

## Engine II: pertukaran, desa, api, air

| Seed | Bicara | Barter | Kabar | Desa | Warga desa | Ganti pemimpin | Kebakaran | Bangunan terbakar | Rakit dibawa | Tenggelam | Terjatuh | Stimulus aktif |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1697 | 328 | 6106 | 1 | 79 | 5 | 6 | 0 | 4 | 0 | 0 | 0 |
| 2 | 1862 | 148 | 6459 | 1 | 37 | 10 | 20 | 0 | 10 | 0 | 0 | 5 |
| 3 | 2236 | 136 | 8144 | 0 | 0 | 0 | 6 | 0 | 0 | 0 | 0 | 4 |
| 4 | 2076 | 357 | 7354 | 1 | 11 | 4 | 7 | 0 | 6 | 0 | 1 | 2 |
| 5 | 1438 | 162 | 4327 | 1 | 56 | 7 | 10 | 0 | 0 | 0 | 0 | 0 |
| 6 | 2360 | 162 | 6782 | 1 | 26 | 4 | 3 | 0 | 1 | 0 | 0 | 3 |
| 7 | 4626 | 215 | 15365 | 1 | 152 | 11 | 7 | 0 | 0 | 0 | 0 | 4 |
| 8 | 3730 | 272 | 11797 | 1 | 52 | 11 | 5 | 0 | 1 | 0 | 0 | 3 |

## Gizi (Fase 3d)

Bagian penduduk, dirata-rata atas setiap menit simulasi. Stunting: tinggi badan sekitar 2 SD di bawah anak sebaya yang cukup gizi.

| Seed | Kurang protein/mikro | Berat | Gizi kurang/buruk | Kurus | Kurang protein | Kurang mikro | Balita pendek | Anak <15 pendek | Dewasa pendek | Ibu hamil/menyusui kurang | Diet protein | Diet mikro |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 11% | 2% | 13% | 5% | 9% | 7% | 1% | 1% | 2% | 31% | 1,13 | 1,09 |
| 2 | 15% | 3% | 16% | 3% | 13% | 9% | 1% | 3% | 13% | 53% | 1,10 | 1,06 |
| 3 | 29% | 11% | 30% | 4% | 26% | 20% | 3% | 11% | 20% | 56% | 1,00 | 1,02 |
| 4 | 14% | 3% | 16% | 4% | 12% | 8% | 0% | 1% | 3% | 40% | 1,09 | 1,06 |
| 5 | 10% | 1% | 10% | 1% | 8% | 6% | 0% | 0% | 0% | 35% | 1,14 | 1,10 |
| 6 | 5% | 1% | 8% | 3% | 4% | 4% | 1% | 1% | 0% | 13% | 1,32 | 1,17 |
| 7 | 33% | 12% | 33% | 1% | 21% | 28% | 16% | 21% | 12% | 61% | 1,04 | 0,96 |
| 8 | 12% | 2% | 14% | 4% | 9% | 8% | 0% | 2% | 3% | 40% | 1,12 | 1,08 |
| **Median** | 13% | 3% | 15% | | | | 1% | 2% | 3% | | | |

Acuan: stunting balita di dunia 26% pada 2011, 30–50% di Asia Selatan dan Afrika sub-Sahara (Black dkk. 2013, Lancet 382:427); petani awal lebih pendek dan lebih sering kurang gizi daripada pemburu-peramu (Cohen & Armelagos 1984).
