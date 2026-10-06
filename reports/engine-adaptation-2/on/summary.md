# Laporan soak — engine II aktif

- Dibuat: 2026-10-06 20:04:39
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 20 menit simulasi per dunia (≈ 150 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 20 m (utuh) | 1 | 47 | 8 | 1 | 3 | 4 | 33,4 | 0,69 | 29,9 | 6,8 | 3,9 | 18,3 | 16 | 0,47 | 0 | 0,20 |
| 2 | 11 m | 2 | 32 | 4 | 0 | 1 | 1 | 50,9 | 0,86 | 43,8 | 7,7 | 2,6 | 17,3 | – | 0,33 | 0 | 0,12 |
| 3 | 20 m (utuh) | 1 | 70 | 8 | 1 | 6 | 8 | 31,4 | 0,73 | 26,0 | 6,1 | 4,3 | 18,5 | 20 | 0,49 | 172 | 0,24 |
| 4 | 20 m (utuh) | 1 | 53 | 8 | 0 | 1 | 6 | 38,2 | 0,77 | 32,5 | 5,4 | 3,7 | 20,0 | 20 | 0,38 | 0 | 0,19 |
| 5 | 20 m (utuh) | 1 | 33 | 7 | 1 | 2 | 7 | 27,1 | 0,60 | 25,6 | 3,5 | 4,9 | 21,2 | 24 | 0,32 | 462 | 0,20 |
| 6 | 20 m (utuh) | 1 | 42 | 8 | 1 | 4 | 3 | 30,1 | 0,69 | 26,4 | 6,0 | 3,2 | 19,3 | 20 | 0,45 | 307 | 0,19 |
| 7 | 20 m (utuh) | 1 | 58 | 7 | 0 | 1 | 6 | 29,5 | 0,71 | 24,2 | 5,2 | 3,4 | 17,5 | 22 | 0,70 | 33 | 0,25 |
| 8 | 20 m (utuh) | 1 | 42 | 8 | 0 | 1 | 6 | 37,1 | 0,79 | 29,7 | 3,5 | 4,9 | 18,6 | 18 | 0,32 | 121 | 0,22 |
| **Median** | 20 m | | 44 | 8 | 0 | 2 | 6 | 32,4 | 0,72 | 28,0 | 5,7 | 3,8 | 18,6 | 20 | 0,41 | 77 | 0,20 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **7 dari 8** dunia (selang kepercayaan 95%: 53–98%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 32,4 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,72 | 0,44–0,73 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 28,0 tahun | 28–43 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 20 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 5,7 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,8 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 18,6 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,41 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,076 | 0,13–0,41 | ↓ di bawah | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,137 |
| Diare per orang per tahun | 0,40 |
| Malaria per orang per tahun | 0,58 |
| ISPA per orang per tahun | 0,01 |
| Rasio kelamin (♂ per 100 ♀) | 102 |
| Anggota per rumah | 4,5 |
| Pembunuhan per 100.000 tahun-orang | 77 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 189 | 37% | 0 | 0% |
| Kehausan | 76 | 15% | 0 | 0% |
| Usia tua | 115 | 23% | 0 | 0% |
| Dibunuh | 23 | 5% | 8 | 11% |
| Diterkam hewan | 0 | 0% | 0 | 0% |
| Melahirkan | 5 | 1% | 0 | 0% |
| Neonatal (minggu pertama) | 38 | 7% | 38 | 51% |
| Diare | 27 | 5% | 11 | 15% |
| Malaria | 32 | 6% | 15 | 20% |
| Radang paru (ISPA) | 2 | 0% | 2 | 3% |
| Terbakar | 0 | 0% | 0 | 0% |
| Tenggelam | 0 | 0% | 0 | 0% |
| Terjatuh | 0 | 0% | 0 | 0% |
| **Penyakit menular** | 61 | 12% | 28 | 38% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 4 | – | 1 | 4 | 26,2 | 0,59 |
| 2 | – | – | 0 | 3 | 29,3 | 0,92 |
| 3 | 4 | – | 1 | 7 | 29,7 | 0,71 |
| 4 | – | – | 0 | 1 | 32,9 | 0,51 |
| 5 | 7 | – | 1 | 2 | 31,5 | 0,51 |
| 6 | 6 | – | 1 | 2 | 30,9 | 0,57 |
| 7 | – | – | 0 | 5 | 31,1 | 0,44 |
| 8 | – | – | 0 | 2 | 31,6 | 0,47 |
| **Median** | 5 | – | 0 | 2 | 31,0 | 0,54 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,50 | 161 |
| 15–29 | 0,55 | 96 |
| 30–44 | 0,56 | 53 |
| 45–59 | 0,53 | 45 |
| 60+ | 0,54 | 22 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 30,5 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 47 | 0 | 7 | 3 | 0,98 | 129 / 156 / 59 / 21 / 1 | 0 | 3 / 3 |
| 2 | 32 | 0 | 4 | 22 | 0,99 | 114 / 120 / 123 / 0 / 0 | 0 | 6 / 4 |
| 3 | 70 | 0 | 10 | 30 | 0,98 | 98 / 83 / 73 / 26 / 0 | 0 | 7 / 6 |
| 4 | 53 | 0 | 9 | 19 | 0,98 | 148 / 138 / 64 / 2 / 2 | 0 | 9 / 8 |
| 5 | 48 | 0 | 7 | 5 | 0,98 | 158 / 195 / 59 / 0 / 6 | 0 | 5 / 4 |
| 6 | 42 | 0 | 16 | 5 | 0,98 | 93 / 115 / 81 / 23 / 4 | 0 | 5 / 5 |
| 7 | 76 | 0 | 18 | 1 | 0,99 | 136 / 206 / 79 / 7 / 1 | 0 | 8 / 8 |
| 8 | 56 | 0 | 4 | 18 | 0,98 | 143 / 127 / 33 / 0 / 3 | 0 | 7 / 6 |
| **Median** | 50 | 0 | 8 | 12 | 0,98 | | | 6 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 100% | 8/8 |
| Ayam Hutan | 100% | 100% | 8/8 |
| Kerbau Liar | 85% | 45% | 5/8 |
| Harimau | 55% | 25% | 6/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 98% | 2% | 0% | 0% | 0,03 |
| 2 | 70% | 30% | 0% | 0% | 1,08 |
| 3 | 92% | 6% | 0% | 1% | 0,19 |
| 4 | 98% | 2% | 0% | 0% | 0,07 |
| 5 | 99% | 0% | 0% | 1% | 0,02 |
| 6 | 97% | 0% | 2% | 0% | 0,01 |
| 7 | 99% | 0% | 0% | 1% | 0,00 |
| 8 | 94% | 5% | 0% | 0% | 0,05 |
| **Median** | | 2% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 132 | 11,1 | 11,9 | 38,9 | 5,3 |
| Netral | 447 | 12,6 | 11,0 | 35,0 | 6,9 |
| La Niña | 217 | 7,3 | 10,6 | 33,3 | 5,4 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 119 | ×0,79 | ×0,97 | 25% |
| La Niña | 187 | ×0,55 | ×0,87 | 22% |

## Engine II: pertukaran, desa, api, air

| Seed | Bicara | Barter | Kabar | Desa | Warga desa | Ganti pemimpin | Kebakaran | Bangunan terbakar | Rakit dibawa | Tenggelam | Terjatuh | Stimulus aktif |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 1036 | 119 | 3101 | 1 | 39 | 7 | 12 | 0 | 1 | 0 | 0 | 0 |
| 2 | 427 | 28 | 1612 | 0 | 0 | 0 | 11 | 0 | 0 | 0 | 0 | 0 |
| 3 | 3781 | 151 | 12749 | 1 | 29 | 10 | 12 | 1 | 6 | 0 | 0 | 1 |
| 4 | 2914 | 53 | 9060 | 1 | 32 | 8 | 13 | 1 | 0 | 0 | 0 | 0 |
| 5 | 1456 | 168 | 4347 | 1 | 23 | 6 | 14 | 0 | 1 | 0 | 0 | 1 |
| 6 | 1009 | 167 | 2971 | 1 | 35 | 9 | 3 | 0 | 3 | 0 | 0 | 2 |
| 7 | 3043 | 80 | 10180 | 1 | 10 | 12 | 8 | 0 | 0 | 0 | 0 | 1 |
| 8 | 1431 | 109 | 4466 | 1 | 12 | 11 | 6 | 0 | 1 | 0 | 0 | 1 |
