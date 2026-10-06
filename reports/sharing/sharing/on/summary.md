# Laporan soak — berbagi pangan (aktif)

- Dibuat: 2026-10-06 23:45:34
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 20 menit simulasi per dunia (≈ 150 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 20 m (utuh) | 1 | 21 | 7 | 0 | 1 | 5 | 48,2 | 0,83 | 41,7 | 2,7 | 3,7 | 20,6 | 76 | 0,42 | 71 | 0,21 |
| 2 | 20 m (utuh) | 1 | 56 | 8 | 1 | 6 | 2 | 41,0 | 0,83 | 33,8 | 5,5 | 4,4 | 18,6 | 38 | 0,39 | 43 | 0,22 |
| 3 | 20 m (utuh) | 1 | 104 | 9 | 1 | 1 | 3 | 35,4 | 0,83 | 27,0 | 6,6 | 3,2 | 18,9 | 20 | 0,43 | 0 | 0,25 |
| 4 | 20 m (utuh) | 1 | 135 | 8 | 1 | 3 | 5 | 41,5 | 0,79 | 35,8 | 6,2 | 3,4 | 19,4 | 20 | 0,36 | 21 | 0,37 |
| 5 | 20 m (utuh) | 1 | 84 | 8 | 1 | 2 | 5 | 37,3 | 0,81 | 29,6 | 6,7 | 3,5 | 19,2 | 16 | 0,36 | 159 | 0,27 |
| 6 | 20 m (utuh) | 1 | 88 | 9 | 0 | 4 | 4 | 29,7 | 0,81 | 20,1 | 5,2 | 3,4 | 19,9 | 20 | 0,43 | 0 | 0,25 |
| 7 | 20 m (utuh) | 1 | 136 | 9 | 0 | 3 | 12 | 30,4 | 0,76 | 23,1 | 5,9 | 4,2 | 18,8 | 18 | 0,35 | 51 | 0,37 |
| 8 | 20 m (utuh) | 1 | 146 | 9 | 1 | 3 | 5 | 33,5 | 0,82 | 25,1 | 8,0 | 3,4 | 17,6 | 20 | 0,42 | 0 | 0,32 |
| **Median** | 20 m | | 96 | 8 | 1 | 3 | 5 | 36,3 | 0,82 | 28,3 | 6,1 | 3,5 | 19,0 | 20 | 0,40 | 32 | 0,26 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **8 dari 8** dunia (selang kepercayaan 95%: 68–100%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 36,3 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,82 | 0,44–0,73 | ↑ di atas | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 28,3 tahun | 28–43 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 20 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 6,1 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,5 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 19,0 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,40 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,064 | 0,13–0,41 | ↓ di bawah | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,105 |
| Diare per orang per tahun | 0,31 |
| Malaria per orang per tahun | 0,69 |
| ISPA per orang per tahun | 0,00 |
| Rasio kelamin (♂ per 100 ♀) | 119 |
| Anggota per rumah | 24,4 |
| Pembunuhan per 100.000 tahun-orang | 32 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 148 | 20% | 0 | 0% |
| Kehausan | 251 | 34% | 0 | 0% |
| Usia tua | 149 | 20% | 0 | 0% |
| Dibunuh | 11 | 2% | 1 | 1% |
| Diterkam hewan | 0 | 0% | 0 | 0% |
| Melahirkan | 14 | 2% | 0 | 0% |
| Neonatal (minggu pertama) | 54 | 7% | 54 | 47% |
| Diare | 42 | 6% | 21 | 18% |
| Malaria | 62 | 8% | 39 | 34% |
| Radang paru (ISPA) | 0 | 0% | 0 | 0% |
| Terbakar | 0 | 0% | 0 | 0% |
| Tenggelam | 0 | 0% | 0 | 0% |
| Terjatuh | 0 | 0% | 0 | 0% |
| **Penyakit menular** | 104 | 14% | 60 | 52% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | – | – | 0 | 5 | 37,0 | 0,65 |
| 2 | 5 | – | 1 | 2 | 36,8 | 0,92 |
| 3 | 8 | – | 1 | 8 | 33,5 | 0,93 |
| 4 | 4 | – | 1 | 1 | 35,0 | 0,93 |
| 5 | 4 | – | 1 | 7 | 34,7 | 0,73 |
| 6 | – | – | 0 | 11 | 31,3 | 0,81 |
| 7 | – | – | 0 | 14 | 33,9 | 0,55 |
| 8 | 5 | – | 1 | 6 | 33,1 | 0,92 |
| **Median** | 5 | – | 1 | 6 | 34,3 | 0,86 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,92 | 382 |
| 15–29 | 0,93 | 199 |
| 30–44 | 0,91 | 93 |
| 45–59 | 0,93 | 59 |
| 60+ | 0,92 | 37 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 34,6 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 32 | 0 | 11 | 12 | 0,98 | 131 / 212 / 55 / 17 / 6 | 0 | 4 / 4 |
| 2 | 56 | 0 | 8 | 69 | 0,99 | 104 / 129 / 87 / 17 / 4 | 0 | 5 / 5 |
| 3 | 104 | 0 | 11 | 39 | 0,98 | 115 / 213 / 61 / 7 / 3 | 0 | 3 / 3 |
| 4 | 137 | 0 | 5 | 52 | 0,97 | 136 / 173 / 46 / 13 / 0 | 0 | 3 / 2 |
| 5 | 84 | 0 | 7 | 25 | 0,97 | 107 / 108 / 157 / 4 / 0 | 0 | 4 / 3 |
| 6 | 88 | 0 | 10 | 13 | 0,98 | 157 / 140 / 22 / 1 / 0 | 0 | 6 / 5 |
| 7 | 136 | 0 | 6 | 21 | 0,96 | 174 / 202 / 24 / 15 / 4 | 0 | 5 / 5 |
| 8 | 146 | 0 | 10 | 73 | 0,97 | 99 / 202 / 85 / 20 / 0 | 0 | 4 / 3 |
| **Median** | 96 | 0 | 9 | 32 | 0,98 | | | 4 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 100% | 8/8 |
| Ayam Hutan | 100% | 100% | 8/8 |
| Kerbau Liar | 100% | 75% | 8/8 |
| Harimau | 55% | 30% | 4/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 98% | 2% | 0% | 1% | 0,13 |
| 2 | 78% | 20% | 0% | 1% | 0,47 |
| 3 | 65% | 35% | 0% | 1% | 0,57 |
| 4 | 64% | 35% | 0% | 0% | 0,42 |
| 5 | 85% | 14% | 0% | 2% | 0,32 |
| 6 | 85% | 12% | 2% | 1% | 0,17 |
| 7 | 99% | 0% | 0% | 1% | 0,02 |
| 8 | 71% | 28% | 0% | 1% | 0,38 |
| **Median** | | 17% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 129 | 5,6 | 5,9 | 41,7 | 30,8 |
| Netral | 456 | 4,3 | 4,2 | 38,2 | 26,9 |
| La Niña | 224 | 4,8 | 4,9 | 39,5 | 30,1 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 89 | ×1,03 | ×0,96 | 22% |
| La Niña | 133 | ×1,04 | ×0,91 | 23% |

## Engine II: pertukaran, desa, api, air

| Seed | Bicara | Barter | Kabar | Desa | Warga desa | Ganti pemimpin | Kebakaran | Bangunan terbakar | Rakit dibawa | Tenggelam | Terjatuh | Stimulus aktif |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 834 | 67 | 2604 | 1 | 11 | 7 | 6 | 0 | 0 | 0 | 0 | 4 |
| 2 | 1405 | 123 | 4936 | 1 | 56 | 13 | 7 | 0 | 0 | 0 | 0 | 0 |
| 3 | 1467 | 146 | 5389 | 1 | 104 | 11 | 7 | 0 | 0 | 0 | 0 | 5 |
| 4 | 2528 | 319 | 9333 | 1 | 134 | 11 | 8 | 0 | 8 | 0 | 0 | 5 |
| 5 | 1954 | 109 | 6509 | 1 | 84 | 5 | 12 | 0 | 1 | 0 | 0 | 2 |
| 6 | 2403 | 125 | 8482 | 1 | 88 | 11 | 12 | 0 | 1 | 0 | 0 | 2 |
| 7 | 4520 | 178 | 14361 | 2 | 100 | 10 | 6 | 0 | 7 | 0 | 0 | 2 |
| 8 | 2973 | 120 | 10783 | 1 | 140 | 11 | 13 | 0 | 0 | 0 | 0 | 1 |
