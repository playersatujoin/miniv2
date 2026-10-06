# Laporan soak — tanpa bawon

- Dibuat: 2026-10-06 23:48:48
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 20 menit simulasi per dunia (≈ 150 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 20 m (utuh) | 1 | 30 | 6 | 1 | 1 | 7 | 50,4 | 0,86 | 43,0 | 3,8 | 3,1 | 23,0 | 76 | 0,39 | 0 | 0,28 |
| 2 | 20 m (utuh) | 1 | 79 | 8 | 0 | 1 | 2 | 38,6 | 0,79 | 32,7 | 7,7 | 3,2 | 18,1 | 18 | 0,33 | 48 | 0,27 |
| 3 | 20 m (utuh) | 1 | 144 | 8 | 0 | 5 | 3 | 46,5 | 0,87 | 38,3 | 6,5 | 3,6 | 18,6 | 18 | 0,48 | 0 | 0,34 |
| 4 | 20 m (utuh) | 1 | 104 | 8 | 1 | 6 | 6 | 38,0 | 0,86 | 28,1 | 5,4 | 3,5 | 19,3 | 18 | 0,37 | 46 | 0,45 |
| 5 | 20 m (utuh) | 1 | 81 | 8 | 0 | 2 | 4 | 32,9 | 0,75 | 27,1 | 4,8 | 3,3 | 19,3 | 18 | 0,39 | 25 | 0,30 |
| 6 | 20 m (utuh) | 1 | 104 | 8 | 0 | 4 | 3 | 29,7 | 0,82 | 20,1 | 6,4 | 3,6 | 20,3 | 22 | 0,39 | 32 | 0,33 |
| 7 | 20 m (utuh) | 1 | 114 | 9 | 0 | 5 | 10 | 29,1 | 0,76 | 21,2 | 5,5 | 4,0 | 18,9 | 18 | 0,32 | 33 | 0,41 |
| 8 | 20 m (utuh) | 1 | 154 | 9 | 1 | 1 | 6 | 32,0 | 0,79 | 23,8 | 6,4 | 3,6 | 18,4 | 18 | 0,36 | 19 | 0,46 |
| **Median** | 20 m | | 104 | 8 | 0 | 3 | 5 | 35,5 | 0,81 | 27,6 | 5,9 | 3,5 | 19,1 | 18 | 0,38 | 28 | 0,34 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **8 dari 8** dunia (selang kepercayaan 95%: 68–100%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 35,5 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,81 | 0,44–0,73 | ↑ di atas | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 27,6 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 18 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 5,9 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,5 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 19,1 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,38 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,070 | 0,13–0,41 | ↓ di bawah | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,111 |
| Diare per orang per tahun | 0,36 |
| Malaria per orang per tahun | 0,68 |
| ISPA per orang per tahun | 0,00 |
| Rasio kelamin (♂ per 100 ♀) | 97 |
| Anggota per rumah | 21,1 |
| Pembunuhan per 100.000 tahun-orang | 28 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 154 | 19% | 0 | 0% |
| Kehausan | 302 | 38% | 0 | 0% |
| Usia tua | 143 | 18% | 0 | 0% |
| Dibunuh | 8 | 1% | 4 | 3% |
| Diterkam hewan | 0 | 0% | 0 | 0% |
| Melahirkan | 15 | 2% | 0 | 0% |
| Neonatal (minggu pertama) | 61 | 8% | 61 | 46% |
| Diare | 46 | 6% | 29 | 22% |
| Malaria | 66 | 8% | 37 | 28% |
| Radang paru (ISPA) | 1 | 0% | 1 | 1% |
| Terbakar | 0 | 0% | 0 | 0% |
| Tenggelam | 0 | 0% | 0 | 0% |
| Terjatuh | 0 | 0% | 0 | 0% |
| **Penyakit menular** | 113 | 14% | 67 | 51% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 6 | – | 1 | 5 | 37,3 | 0,74 |
| 2 | – | – | 0 | 3 | 32,9 | 0,95 |
| 3 | – | – | 0 | 7 | 35,2 | 0,95 |
| 4 | 6 | – | 1 | 1 | 36,9 | 0,84 |
| 5 | – | – | 0 | 6 | 34,5 | 0,68 |
| 6 | – | – | 0 | 11 | 31,3 | 0,88 |
| 7 | – | – | 0 | 13 | 34,2 | 0,61 |
| 8 | 5 | – | 1 | 10 | 33,6 | 0,88 |
| **Median** | 6 | – | 0 | 6 | 34,4 | 0,86 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,85 | 399 |
| 15–29 | 0,89 | 215 |
| 30–44 | 0,87 | 107 |
| 45–59 | 0,87 | 64 |
| 60+ | 0,89 | 25 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 34,6 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 32 | 0 | 11 | 12 | 0,98 | 181 / 180 / 138 / 4 / 0 | 0 | 4 / 3 |
| 2 | 79 | 0 | 8 | 47 | 0,98 | 92 / 166 / 168 / 1 / 4 | 0 | 5 / 5 |
| 3 | 144 | 0 | 11 | 79 | 0,97 | 149 / 173 / 66 / 0 / 0 | 0 | 4 / 2 |
| 4 | 106 | 0 | 5 | 45 | 0,97 | 128 / 168 / 100 / 6 / 0 | 0 | 5 / 4 |
| 5 | 86 | 0 | 7 | 25 | 0,97 | 134 / 134 / 55 / 22 / 0 | 0 | 4 / 3 |
| 6 | 104 | 0 | 10 | 25 | 0,97 | 168 / 182 / 35 / 3 / 1 | 0 | 5 / 5 |
| 7 | 135 | 0 | 6 | 15 | 0,96 | 172 / 179 / 19 / 6 / 0 | 0 | 9 / 8 |
| 8 | 154 | 0 | 10 | 37 | 0,96 | 86 / 154 / 105 / 14 / 0 | 0 | 5 / 4 |
| **Median** | 105 | 0 | 9 | 31 | 0,97 | | | 5 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 100% | 8/8 |
| Ayam Hutan | 100% | 100% | 8/8 |
| Kerbau Liar | 90% | 55% | 7/8 |
| Harimau | 55% | 30% | 2/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 95% | 3% | 0% | 2% | 0,15 |
| 2 | 67% | 32% | 0% | 1% | 0,90 |
| 3 | 54% | 46% | 0% | 1% | 0,78 |
| 4 | 66% | 34% | 0% | 1% | 0,50 |
| 5 | 91% | 8% | 0% | 1% | 0,23 |
| 6 | 79% | 17% | 2% | 1% | 0,24 |
| 7 | 99% | 1% | 0% | 1% | 0,02 |
| 8 | 66% | 34% | 0% | 0% | 0,25 |
| **Median** | | 25% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 139 | 4,6 | 7,4 | 40,7 | 27,5 |
| Netral | 460 | 4,7 | 4,0 | 38,9 | 34,7 |
| La Niña | 201 | 4,3 | 3,9 | 41,6 | 31,8 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 77 | ×0,81 | ×1,12 | 27% |
| La Niña | 115 | ×0,67 | ×0,60 | 15% |

## Engine II: pertukaran, desa, api, air

| Seed | Bicara | Barter | Kabar | Desa | Warga desa | Ganti pemimpin | Kebakaran | Bangunan terbakar | Rakit dibawa | Tenggelam | Terjatuh | Stimulus aktif |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 897 | 74 | 2851 | 1 | 22 | 8 | 10 | 0 | 0 | 0 | 0 | 0 |
| 2 | 1560 | 104 | 5742 | 1 | 78 | 11 | 7 | 0 | 5 | 0 | 0 | 7 |
| 3 | 1880 | 131 | 7028 | 1 | 144 | 8 | 11 | 0 | 0 | 0 | 0 | 9 |
| 4 | 2075 | 339 | 7332 | 1 | 39 | 12 | 8 | 0 | 0 | 0 | 0 | 2 |
| 5 | 2463 | 113 | 7703 | 1 | 66 | 5 | 5 | 0 | 4 | 0 | 0 | 4 |
| 6 | 2759 | 146 | 9862 | 1 | 104 | 14 | 12 | 0 | 0 | 0 | 0 | 0 |
| 7 | 4614 | 214 | 14628 | 2 | 47 | 11 | 7 | 0 | 4 | 0 | 0 | 6 |
| 8 | 3165 | 153 | 11052 | 1 | 154 | 8 | 14 | 0 | 4 | 0 | 0 | 3 |
