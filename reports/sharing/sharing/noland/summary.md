# Laporan soak — tanpa pengetahuan lahan

- Dibuat: 2026-10-06 23:48:53
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 20 menit simulasi per dunia (≈ 150 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 20 m (utuh) | 1 | 45 | 8 | 1 | 2 | 7 | 31,5 | 0,79 | 23,1 | 4,6 | 3,5 | 18,6 | 18 | 0,30 | 272 | 0,32 |
| 2 | 20 m (utuh) | 1 | 40 | 8 | 1 | 2 | 2 | 42,2 | 0,79 | 38,0 | 5,7 | 4,6 | 19,2 | 70 | 0,27 | 0 | 0,26 |
| 3 | 20 m (utuh) | 1 | 74 | 8 | 0 | 1 | 5 | 33,0 | 0,77 | 26,2 | 5,3 | 4,2 | 19,7 | 20 | 0,35 | 173 | 0,30 |
| 4 | 20 m (utuh) | 1 | 101 | 8 | 1 | 1 | 2 | 37,4 | 0,84 | 28,8 | 5,2 | 3,9 | 19,0 | 18 | 0,42 | 209 | 0,43 |
| 5 | 20 m (utuh) | 1 | 107 | 8 | 1 | 1 | 5 | 27,8 | 0,79 | 18,7 | 5,0 | 4,0 | 18,8 | 18 | 0,40 | 0 | 0,45 |
| 6 | 20 m (utuh) | 1 | 30 | 7 | 1 | 6 | 3 | 38,6 | 0,78 | 32,3 | 5,0 | 4,7 | 18,3 | 58 | 0,40 | 429 | 0,25 |
| 7 | 20 m (utuh) | 1 | 297 | 9 | 0 | 5 | 10 | 29,3 | 0,76 | 22,1 | 8,9 | 2,9 | 17,5 | 18 | 0,57 | 36 | 0,63 |
| 8 | 20 m (utuh) | 1 | 114 | 8 | 1 | 1 | 5 | 31,4 | 0,79 | 23,6 | 5,5 | 3,6 | 17,7 | 18 | 0,40 | 0 | 0,41 |
| **Median** | 20 m | | 88 | 8 | 1 | 2 | 5 | 32,2 | 0,79 | 24,9 | 5,2 | 3,9 | 18,7 | 18 | 0,40 | 104 | 0,37 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **8 dari 8** dunia (selang kepercayaan 95%: 68–100%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 32,2 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,79 | 0,44–0,73 | ↑ di atas | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 24,9 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 18 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 5,2 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,9 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 18,7 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,40 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,080 | 0,13–0,41 | ↓ di bawah | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,127 |
| Diare per orang per tahun | 0,31 |
| Malaria per orang per tahun | 0,66 |
| ISPA per orang per tahun | 0,01 |
| Rasio kelamin (♂ per 100 ♀) | 93 |
| Anggota per rumah | 14,1 |
| Pembunuhan per 100.000 tahun-orang | 104 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 306 | 31% | 0 | 0% |
| Kehausan | 254 | 26% | 0 | 0% |
| Usia tua | 133 | 14% | 0 | 0% |
| Dibunuh | 28 | 3% | 4 | 2% |
| Diterkam hewan | 0 | 0% | 0 | 0% |
| Melahirkan | 17 | 2% | 0 | 0% |
| Neonatal (minggu pertama) | 84 | 9% | 84 | 50% |
| Diare | 74 | 8% | 35 | 21% |
| Malaria | 74 | 8% | 45 | 27% |
| Radang paru (ISPA) | 2 | 0% | 1 | 1% |
| Terbakar | 0 | 0% | 0 | 0% |
| Tenggelam | 0 | 0% | 0 | 0% |
| Terjatuh | 1 | 0% | 0 | 0% |
| **Penyakit menular** | 150 | 15% | 81 | 48% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 4 | – | 1 | 2 | 36,3 | 0,62 |
| 2 | 5 | – | 1 | 1 | 36,9 | 0,78 |
| 3 | – | – | 0 | 4 | 33,8 | 0,83 |
| 4 | 4 | – | 1 | 3 | 35,9 | 0,92 |
| 5 | 7 | – | 1 | 5 | 32,2 | 0,92 |
| 6 | 5 | – | 1 | 8 | 36,1 | 0,59 |
| 7 | – | – | 0 | 7 | 31,9 | 0,90 |
| 8 | 6 | – | 1 | 10 | 33,3 | 0,94 |
| **Median** | 5 | – | 1 | 4 | 34,8 | 0,87 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,88 | 391 |
| 15–29 | 0,92 | 239 |
| 30–44 | 0,91 | 96 |
| 45–59 | 0,85 | 52 |
| 60+ | 0,80 | 30 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 34,3 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 62 | 0 | – | 4 | 0,97 | 77 / 167 / 168 / 0 / 0 | 0 | 7 / 5 |
| 2 | 40 | 0 | 8 | 28 | 0,99 | 118 / 201 / 77 / 16 / 5 | 0 | 3 / 3 |
| 3 | 74 | 0 | 7 | 43 | 0,98 | 104 / 173 / 126 / 25 / 0 | 0 | 3 / 2 |
| 4 | 104 | 0 | 5 | 38 | 0,97 | 102 / 144 / 73 / 12 / 0 | 0 | 2 / 1 |
| 5 | 123 | 0 | 6 | 85 | 0,97 | 112 / 141 / 67 / 0 / 4 | 0 | 5 / 4 |
| 6 | 32 | 0 | 11 | 5 | 0,99 | 111 / 191 / 29 / 0 / 0 | 0 | 4 / 2 |
| 7 | 297 | 0 | 6 | 77 | 0,95 | 97 / 169 / 82 / 0 / 0 | 0 | 6 / 4 |
| 8 | 114 | 0 | 7 | 100 | 0,97 | 96 / 135 / 101 / 17 / 0 | 0 | 4 / 3 |
| **Median** | 89 | 0 | 7 | 40 | 0,97 | | | 4 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 100% | 8/8 |
| Ayam Hutan | 100% | 100% | 8/8 |
| Kerbau Liar | 98% | 30% | 4/8 |
| Harimau | 55% | 40% | 2/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 100% | 0% | 0% | 0% | 0,00 |
| 2 | 89% | 10% | 0% | 1% | 0,19 |
| 3 | 88% | 11% | 0% | 1% | 0,31 |
| 4 | 76% | 23% | 0% | 1% | 0,48 |
| 5 | 62% | 38% | 0% | 1% | 0,66 |
| 6 | 97% | 0% | 2% | 0% | 0,02 |
| 7 | 51% | 49% | 0% | 0% | 0,52 |
| 8 | 69% | 30% | 0% | 2% | 0,47 |
| **Median** | | 17% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 143 | 7,6 | 8,3 | 43,2 | 38,5 |
| Netral | 455 | 7,2 | 7,4 | 39,2 | 41,9 |
| La Niña | 204 | 7,3 | 6,6 | 42,3 | 45,8 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 101 | ×0,98 | ×0,98 | 19% |
| La Niña | 134 | ×0,94 | ×0,99 | 20% |

## Engine II: pertukaran, desa, api, air

| Seed | Bicara | Barter | Kabar | Desa | Warga desa | Ganti pemimpin | Kebakaran | Bangunan terbakar | Rakit dibawa | Tenggelam | Terjatuh | Stimulus aktif |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1347 | 235 | 3917 | 1 | 38 | 8 | 12 | 0 | 4 | 0 | 0 | 0 |
| 2 | 893 | 109 | 2907 | 1 | 13 | 11 | 8 | 0 | 0 | 0 | 0 | 1 |
| 3 | 1194 | 134 | 4165 | 1 | 73 | 12 | 11 | 0 | 0 | 0 | 0 | 2 |
| 4 | 2008 | 346 | 7333 | 1 | 101 | 12 | 10 | 0 | 1 | 0 | 0 | 4 |
| 5 | 4050 | 197 | 15108 | 1 | 64 | 11 | 8 | 0 | 3 | 0 | 0 | 2 |
| 6 | 988 | 70 | 2433 | 1 | 21 | 13 | 10 | 0 | 0 | 0 | 0 | 3 |
| 7 | 8496 | 312 | 31204 | 2 | 264 | 11 | 11 | 3 | 5 | 0 | 1 | 9 |
| 8 | 3883 | 136 | 14457 | 1 | 114 | 11 | 5 | 0 | 5 | 0 | 0 | 2 |
