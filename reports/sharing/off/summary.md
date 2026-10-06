# Laporan soak — berbagi pangan (mati)

- Dibuat: 2026-10-06 23:45:38
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 20 menit simulasi per dunia (≈ 150 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: sharing
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 20 m (utuh) | 1 | 71 | 8 | 1 | 1 | 2 | 25,3 | 0,67 | 19,4 | 6,4 | 3,6 | 17,7 | 18 | 0,35 | 338 | 0,27 |
| 2 | 20 m (utuh) | 1 | 49 | 9 | 1 | 6 | 2 | 46,6 | 0,85 | 39,7 | 3,4 | 3,7 | 19,2 | 52 | 0,35 | 57 | 0,20 |
| 3 | 20 m (utuh) | 1 | 78 | 8 | 1 | 1 | 6 | 33,3 | 0,83 | 24,2 | 6,7 | 4,0 | 18,5 | 18 | 0,39 | 61 | 0,27 |
| 4 | 20 m (utuh) | 1 | 87 | 8 | 1 | 5 | 3 | 39,8 | 0,81 | 33,4 | 5,4 | 4,0 | 19,8 | 18 | 0,47 | 59 | 0,33 |
| 5 | 20 m (utuh) | 1 | 75 | 8 | 1 | 3 | 4 | 35,9 | 0,79 | 28,9 | 5,2 | 3,6 | 17,9 | 22 | 0,42 | 76 | 0,22 |
| 6 | 20 m (utuh) | 1 | 16 | 6 | 1 | 5 | 2 | 33,8 | 0,67 | 32,2 | 3,3 | 2,9 | 18,9 | – | 0,18 | 124 | 0,18 |
| 7 | 20 m (utuh) | 1 | 130 | 9 | 0 | 1 | 8 | 25,2 | 0,72 | 17,1 | 6,1 | 3,6 | 17,9 | 18 | 0,45 | 0 | 0,47 |
| 8 | 20 m (utuh) | 1 | 102 | 9 | 1 | 2 | 6 | 33,0 | 0,81 | 24,9 | 7,2 | 3,2 | 17,9 | 18 | 0,35 | 0 | 0,29 |
| **Median** | 20 m | | 76 | 8 | 1 | 2 | 4 | 33,5 | 0,80 | 26,9 | 5,8 | 3,6 | 18,2 | 18 | 0,37 | 60 | 0,27 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **8 dari 8** dunia (selang kepercayaan 95%: 68–100%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 33,5 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,80 | 0,44–0,73 | ↑ di atas | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 26,9 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 18 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 5,8 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,6 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 18,2 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,37 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,081 | 0,13–0,41 | ↓ di bawah | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,132 |
| Diare per orang per tahun | 0,29 |
| Malaria per orang per tahun | 0,59 |
| ISPA per orang per tahun | 0,01 |
| Rasio kelamin (♂ per 100 ♀) | 111 |
| Anggota per rumah | 13,4 |
| Pembunuhan per 100.000 tahun-orang | 60 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 296 | 36% | 0 | 0% |
| Kehausan | 186 | 22% | 2 | 2% |
| Usia tua | 133 | 16% | 0 | 0% |
| Dibunuh | 19 | 2% | 3 | 2% |
| Diterkam hewan | 0 | 0% | 0 | 0% |
| Melahirkan | 18 | 2% | 0 | 0% |
| Neonatal (minggu pertama) | 70 | 8% | 70 | 56% |
| Diare | 53 | 6% | 34 | 27% |
| Malaria | 49 | 6% | 13 | 10% |
| Radang paru (ISPA) | 5 | 1% | 3 | 2% |
| Terbakar | 0 | 0% | 0 | 0% |
| Tenggelam | 0 | 0% | 0 | 0% |
| Terjatuh | 0 | 0% | 0 | 0% |
| **Penyakit menular** | 107 | 13% | 50 | 40% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 6 | – | 1 | 4 | 31,7 | 0,54 |
| 2 | 5 | – | 1 | 4 | 34,3 | 0,84 |
| 3 | 7 | – | 1 | 5 | 32,4 | 0,90 |
| 4 | 5 | – | 1 | 2 | 36,8 | 0,95 |
| 5 | 7 | – | 1 | 7 | 33,1 | 0,78 |
| 6 | 4 | – | 1 | 9 | 38,3 | 0,54 |
| 7 | – | – | 0 | 11 | 33,1 | 0,88 |
| 8 | 6 | – | 1 | 7 | 33,2 | 0,81 |
| **Median** | 6 | – | 1 | 6 | 33,2 | 0,82 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,84 | 287 |
| 15–29 | 0,85 | 158 |
| 30–44 | 0,87 | 85 |
| 45–59 | 0,85 | 48 |
| 60+ | 0,77 | 30 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 34,3 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 74 | 0 | 16 | 1 | 0,98 | 52 / 173 / 129 / 18 / 0 | 0 | 8 / 7 |
| 2 | 49 | 0 | 8 | 74 | 0,99 | 118 / 107 / 30 / 23 / 0 | 0 | 4 / 3 |
| 3 | 78 | 0 | 8 | 42 | 0,97 | 138 / 166 / 75 / 3 / 0 | 0 | 3 / 2 |
| 4 | 87 | 0 | 5 | 63 | 0,98 | 110 / 202 / 117 / 10 / 3 | 0 | 2 / 2 |
| 5 | 75 | 0 | 6 | 32 | 0,98 | 115 / 157 / 119 / 19 / 7 | 0 | 4 / 4 |
| 6 | 24 | 0 | 11 | 8 | 0,99 | 134 / 192 / 48 / 0 / 0 | 0 | 5 / 3 |
| 7 | 170 | 0 | 6 | 69 | 0,96 | 121 / 137 / 88 / 6 / 1 | 0 | 7 / 7 |
| 8 | 103 | 0 | 7 | 30 | 0,97 | 105 / 169 / 84 / 45 / 0 | 6 | 4 / 3 |
| **Median** | 76 | 0 | 8 | 37 | 0,98 | | | 4 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 100% | 8/8 |
| Ayam Hutan | 100% | 100% | 8/8 |
| Kerbau Liar | 98% | 35% | 7/8 |
| Harimau | 52% | 30% | 3/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 100% | 0% | 0% | 0% | 0,00 |
| 2 | 94% | 5% | 0% | 1% | 0,59 |
| 3 | 69% | 31% | 0% | 0% | 0,59 |
| 4 | 64% | 36% | 0% | 0% | 1,02 |
| 5 | 85% | 14% | 0% | 0% | 0,16 |
| 6 | 100% | 0% | 0% | 0% | 0,01 |
| 7 | 49% | 50% | 0% | 0% | 0,66 |
| 8 | 89% | 10% | 0% | 1% | 0,24 |
| **Median** | | 12% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 128 | 10,1 | 12,5 | 36,1 | 49,4 |
| Netral | 451 | 8,9 | 8,2 | 39,3 | 49,0 |
| La Niña | 197 | 7,4 | 7,6 | 38,1 | 47,4 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 93 | ×1,51 | ×1,34 | 26% |
| La Niña | 147 | ×0,92 | ×0,90 | 18% |

## Engine II: pertukaran, desa, api, air

| Seed | Bicara | Barter | Kabar | Desa | Warga desa | Ganti pemimpin | Kebakaran | Bangunan terbakar | Rakit dibawa | Tenggelam | Terjatuh | Stimulus aktif |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 2077 | 98 | 6661 | 1 | 71 | 7 | 6 | 0 | 0 | 0 | 0 | 6 |
| 2 | 980 | 74 | 3027 | 1 | 49 | 9 | 13 | 0 | 0 | 0 | 0 | 2 |
| 3 | 1460 | 340 | 5360 | 1 | 62 | 10 | 6 | 0 | 4 | 0 | 0 | 2 |
| 4 | 1974 | 564 | 7439 | 1 | 87 | 11 | 7 | 0 | 0 | 0 | 0 | 3 |
| 5 | 2150 | 143 | 7469 | 1 | 15 | 9 | 7 | 0 | 3 | 0 | 0 | 1 |
| 6 | 846 | 63 | 2268 | 0 | 0 | 10 | 9 | 0 | 1 | 0 | 0 | 0 |
| 7 | 8290 | 324 | 31417 | 1 | 74 | 10 | 12 | 0 | 0 | 0 | 0 | 1 |
| 8 | 3269 | 162 | 11303 | 2 | 98 | 10 | 14 | 1 | 0 | 0 | 0 | 1 |
