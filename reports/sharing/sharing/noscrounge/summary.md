# Laporan soak — tanpa makan bersama

- Dibuat: 2026-10-06 23:48:49
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 20 menit simulasi per dunia (≈ 150 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 20 m (utuh) | 1 | 64 | 8 | 1 | 2 | 8 | 29,9 | 0,77 | 22,6 | 5,0 | 4,8 | 18,9 | 18 | 0,35 | 142 | 0,39 |
| 2 | 20 m (utuh) | 1 | 71 | 8 | 1 | 6 | 3 | 49,6 | 0,93 | 38,2 | 4,9 | 3,6 | 21,2 | 24 | 0,28 | 54 | 0,27 |
| 3 | 20 m (utuh) | 1 | 31 | 8 | 1 | 2 | 4 | 33,0 | 0,72 | 27,9 | 4,4 | 3,9 | 21,1 | 22 | 0,39 | 57 | 0,28 |
| 4 | 20 m (utuh) | 1 | 102 | 7 | 1 | 3 | 2 | 37,6 | 0,82 | 29,9 | 5,4 | 4,5 | 19,1 | 28 | 0,42 | 45 | 0,43 |
| 5 | 20 m (utuh) | 1 | 83 | 7 | 1 | 3 | 6 | 30,9 | 0,70 | 26,2 | 5,1 | 4,0 | 20,6 | 20 | 0,35 | 0 | 0,35 |
| 6 | 20 m (utuh) | 1 | 71 | 8 | 0 | 4 | 3 | 26,2 | 0,73 | 18,5 | 4,8 | 3,9 | 20,6 | 18 | 0,42 | 37 | 0,31 |
| 7 | 20 m (utuh) | 1 | 191 | 9 | 0 | 3 | 8 | 30,0 | 0,77 | 22,5 | 6,3 | 3,8 | 18,6 | 18 | 0,40 | 46 | 0,56 |
| 8 | 20 m (utuh) | 1 | 94 | 9 | 1 | 5 | 8 | 34,5 | 0,80 | 27,1 | 5,9 | 3,7 | 19,7 | 18 | 0,47 | 57 | 0,36 |
| **Median** | 20 m | | 77 | 8 | 1 | 3 | 5 | 31,9 | 0,77 | 26,6 | 5,0 | 3,9 | 20,1 | 19 | 0,39 | 50 | 0,35 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **8 dari 8** dunia (selang kepercayaan 95%: 68–100%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 31,9 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,77 | 0,44–0,73 | ↑ di atas | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 26,6 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 19 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 5,0 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,9 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 20,1 tahun | 18–20 | ↑ di atas | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,39 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,083 | 0,13–0,41 | ↓ di bawah | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,131 |
| Diare per orang per tahun | 0,42 |
| Malaria per orang per tahun | 0,67 |
| ISPA per orang per tahun | 0,01 |
| Rasio kelamin (♂ per 100 ♀) | 113 |
| Anggota per rumah | 15,7 |
| Pembunuhan per 100.000 tahun-orang | 50 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 273 | 31% | 3 | 2% |
| Kehausan | 247 | 28% | 1 | 1% |
| Usia tua | 148 | 17% | 0 | 0% |
| Dibunuh | 16 | 2% | 6 | 4% |
| Diterkam hewan | 1 | 0% | 0 | 0% |
| Melahirkan | 16 | 2% | 0 | 0% |
| Neonatal (minggu pertama) | 67 | 8% | 67 | 47% |
| Diare | 60 | 7% | 34 | 24% |
| Malaria | 57 | 6% | 32 | 22% |
| Radang paru (ISPA) | 2 | 0% | 1 | 1% |
| Terbakar | 0 | 0% | 0 | 0% |
| Tenggelam | 0 | 0% | 0 | 0% |
| Terjatuh | 1 | 0% | 0 | 0% |
| **Penyakit menular** | 119 | 13% | 67 | 47% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 5 | – | 1 | 7 | 35,5 | 0,59 |
| 2 | 8 | – | 1 | 6 | 35,9 | 0,90 |
| 3 | 6 | – | 1 | 12 | 41,8 | 0,56 |
| 4 | 7 | – | 1 | 1 | 36,9 | 0,94 |
| 5 | 4 | – | 1 | 7 | 33,9 | 0,48 |
| 6 | – | – | 0 | 11 | 32,1 | 0,57 |
| 7 | – | – | 0 | 9 | 31,8 | 0,84 |
| 8 | 6 | – | 1 | 4 | 33,6 | 0,54 |
| **Median** | 6 | – | 1 | 7 | 34,7 | 0,58 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,57 | 328 |
| 15–29 | 0,58 | 183 |
| 30–44 | 0,59 | 106 |
| 45–59 | 0,59 | 56 |
| 60+ | 0,58 | 34 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 34,0 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 76 | 0 | 10 | 5 | 0,97 | 137 / 157 / 160 / 3 / 1 | 0 | 8 / 8 |
| 2 | 71 | 0 | 8 | 41 | 0,98 | 112 / 140 / 202 / 2 / 0 | 0 | 6 / 5 |
| 3 | 39 | 0 | – | 1 | 0,98 | 114 / 148 / 37 / 39 / 6 | 0 | 3 / 3 |
| 4 | 102 | 0 | 5 | 70 | 0,99 | 106 / 128 / 39 / 2 / 4 | 0 | 4 / 3 |
| 5 | 83 | 0 | 7 | 9 | 0,97 | 157 / 137 / 91 / 2 / 5 | 0 | 4 / 3 |
| 6 | 77 | 0 | 10 | 3 | 0,98 | 160 / 155 / 17 / 0 / 0 | 0 | 6 / 4 |
| 7 | 200 | 0 | 6 | 75 | 0,96 | 82 / 209 / 123 / 0 / 0 | 0 | 8 / 6 |
| 8 | 94 | 0 | 11 | 10 | 0,97 | 72 / 123 / 145 / 57 / 0 | 0 | 4 / 3 |
| **Median** | 80 | 0 | 8 | 10 | 0,98 | | | 5 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 100% | 8/8 |
| Ayam Hutan | 100% | 100% | 8/8 |
| Kerbau Liar | 95% | 60% | 6/8 |
| Harimau | 55% | 20% | 4/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 99% | 0% | 0% | 1% | 0,01 |
| 2 | 72% | 27% | 0% | 2% | 0,52 |
| 3 | 100% | 0% | 0% | 0% | 0,00 |
| 4 | 64% | 36% | 0% | 0% | 0,96 |
| 5 | 98% | 1% | 0% | 1% | 0,02 |
| 6 | 97% | 0% | 2% | 1% | 0,00 |
| 7 | 63% | 35% | 0% | 1% | 0,54 |
| 8 | 98% | 1% | 0% | 1% | 0,01 |
| **Median** | | 1% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 128 | 7,7 | 9,1 | 40,1 | 36,4 |
| Netral | 467 | 7,3 | 7,3 | 38,5 | 29,6 |
| La Niña | 225 | 7,1 | 6,3 | 36,0 | 33,7 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 108 | ×0,76 | ×1,62 | 25% |
| La Niña | 173 | ×1,06 | ×0,95 | 21% |

## Engine II: pertukaran, desa, api, air

| Seed | Bicara | Barter | Kabar | Desa | Warga desa | Ganti pemimpin | Kebakaran | Bangunan terbakar | Rakit dibawa | Tenggelam | Terjatuh | Stimulus aktif |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1688 | 264 | 5088 | 1 | 37 | 11 | 9 | 0 | 0 | 0 | 0 | 3 |
| 2 | 1205 | 120 | 4294 | 1 | 62 | 10 | 10 | 0 | 10 | 0 | 0 | 4 |
| 3 | 1193 | 224 | 3952 | 1 | 21 | 11 | 6 | 0 | 10 | 0 | 0 | 2 |
| 4 | 2316 | 240 | 8724 | 1 | 101 | 11 | 5 | 0 | 0 | 0 | 1 | 3 |
| 5 | 1876 | 164 | 5699 | 1 | 45 | 7 | 8 | 0 | 0 | 0 | 0 | 1 |
| 6 | 2395 | 100 | 8154 | 1 | 71 | 14 | 11 | 0 | 4 | 0 | 0 | 2 |
| 7 | 6471 | 328 | 23150 | 1 | 185 | 7 | 10 | 0 | 1 | 0 | 0 | 7 |
| 8 | 2526 | 126 | 8539 | 1 | 86 | 7 | 10 | 0 | 0 | 0 | 0 | 0 |
