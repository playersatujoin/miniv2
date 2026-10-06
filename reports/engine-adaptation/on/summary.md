# Laporan soak — engine-adaptation

- Dibuat: 2026-10-06 17:59:25
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 20 menit simulasi per dunia (≈ 150 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: tidak ada
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 20 m (utuh) | 1 | 23 | 8 | 0 | 1 | 3 | 28,8 | 0,56 | 31,5 | 3,6 | 5,7 | 18,7 | 18 | 0,29 | 1092 | 0,16 |
| 2 | 20 m (utuh) | 1 | 73 | 8 | 0 | 1 | 5 | 27,1 | 0,74 | 19,0 | 6,5 | 3,6 | 18,3 | 18 | 0,37 | 47 | 0,30 |
| 3 | 19 m | 2 | 5 | 1 | 0 | 1 | 0 | – | – | – | – | – | – | – | 0,21 | – | 0,11 |
| 4 | 20 m (utuh) | 1 | 38 | 8 | 0 | 1 | 2 | 29,2 | 0,67 | 25,8 | 5,3 | 3,3 | 20,3 | 18 | 0,23 | 0 | 0,19 |
| 5 | 20 m (utuh) | 1 | 11 | 6 | 0 | 1 | 2 | 43,3 | 0,80 | 38,6 | 2,5 | 4,1 | – | – | 0,16 | 135 | 0,14 |
| 6 | 20 m (utuh) | 1 | 56 | 8 | 0 | 2 | 2 | 31,8 | 0,79 | 23,9 | 7,0 | 3,6 | 18,4 | 20 | 0,33 | 52 | 0,17 |
| 7 | 20 m (utuh) | 1 | 65 | 8 | 1 | 1 | 6 | 26,4 | 0,68 | 20,3 | 8,3 | 3,1 | 17,1 | 20 | 0,50 | 26 | 0,28 |
| 8 | 20 m (utuh) | 1 | 6 | 5 | 1 | 2 | 2 | – | – | – | 0,0 | – | – | 70 | 0,41 | 261 | 0,17 |
| **Median** | 20 m | | 30 | 8 | 0 | 1 | 2 | 29,0 | 0,71 | 24,9 | 5,3 | 3,6 | 18,4 | 19 | 0,31 | 52 | 0,17 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **7 dari 8** dunia (selang kepercayaan 95%: 53–98%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 29,0 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,71 | 0,44–0,73 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 24,9 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 19 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 5,3 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,6 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 18,4 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,31 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,090 | 0,13–0,41 | ↓ di bawah | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,139 |
| Diare per orang per tahun | 0,35 |
| Malaria per orang per tahun | 0,55 |
| ISPA per orang per tahun | 0,01 |
| Rasio kelamin (♂ per 100 ♀) | 116 |
| Anggota per rumah | 3,6 |
| Pembunuhan per 100.000 tahun-orang | 52 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 205 | 41% | 0 | 0% |
| Kehausan | 75 | 15% | 0 | 0% |
| Usia tua | 97 | 19% | 0 | 0% |
| Dibunuh | 20 | 4% | 4 | 6% |
| Diterkam hewan | 0 | 0% | 0 | 0% |
| Melahirkan | 9 | 2% | 0 | 0% |
| Neonatal (minggu pertama) | 28 | 6% | 28 | 43% |
| Diare | 20 | 4% | 12 | 18% |
| Malaria | 46 | 9% | 21 | 32% |
| Radang paru (ISPA) | 0 | 0% | 0 | 0% |
| **Penyakit menular** | 66 | 13% | 33 | 51% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | – | – | 0 | 5 | 34,2 | 0,64 |
| 2 | – | – | 0 | 2 | 28,2 | 0,89 |
| 3 | – | – | 0 | 6 | 24,0 | 0,29 |
| 4 | – | – | 0 | 6 | 28,7 | 0,67 |
| 5 | – | – | 0 | 0 | 39,6 | 0,91 |
| 6 | – | – | 0 | 5 | 30,4 | 0,92 |
| 7 | 5 | – | 1 | 1 | 30,1 | 0,93 |
| 8 | 4 | – | 1 | 3 | 53,5 | 0,58 |
| **Median** | 4 | – | 0 | 4 | 30,2 | 0,78 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,85 | 115 |
| 15–29 | 0,89 | 70 |
| 30–44 | 0,91 | 36 |
| 45–59 | 0,89 | 31 |
| 60+ | 0,87 | 25 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 30,8 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 29 | 0 | 15 | 4 | 0,99 | 99 / 222 / 125 / 6 / 0 | 0 | 8 / 7 |
| 2 | 108 | 0 | 5 | 48 | 0,98 | 138 / 138 / 59 / 11 / 0 | 0 | 5 / 4 |
| 3 | 8 | 0 | 7 | 11 | 1,00 | 108 / 150 / 111 / 9 / 0 | 0 | 5 / 4 |
| 4 | 50 | 0 | 7 | 20 | 0,99 | 122 / 154 / 96 / 0 / 0 | 0 | 8 / 6 |
| 5 | 19 | 0 | 5 | 36 | 0,99 | 86 / 141 / 111 / 20 / 0 | 0 | 6 / 5 |
| 6 | 56 | 0 | 15 | 43 | 0,98 | 147 / 210 / 59 / 3 / 2 | 0 | 6 / 6 |
| 7 | 91 | 0 | 8 | 42 | 0,99 | 136 / 123 / 110 / 0 / 5 | 0 | 6 / 5 |
| 8 | 36 | 0 | 9 | 9 | 0,99 | 88 / 81 / 69 / 19 / 6 | 0 | 3 / 3 |
| **Median** | 43 | 0 | 8 | 28 | 0,99 | | | 6 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 100% | 8/8 |
| Ayam Hutan | 100% | 100% | 8/8 |
| Kerbau Liar | 88% | 40% | 6/8 |
| Harimau | 62% | 35% | 3/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 99% | 0% | 0% | 1% | 0,01 |
| 2 | 57% | 43% | 0% | 0% | 0,72 |
| 3 | 98% | 2% | 0% | 0% | 1,54 |
| 4 | 99% | 1% | 0% | 0% | 0,04 |
| 5 | 91% | 9% | 0% | 0% | 0,66 |
| 6 | 84% | 13% | 2% | 1% | 0,32 |
| 7 | 57% | 43% | 0% | 0% | 0,70 |
| 8 | 100% | 0% | 0% | 0% | 0,09 |
| **Median** | | 5% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 88 | 15,6 | 12,9 | 40,3 | 36,5 |
| Netral | 354 | 12,6 | 12,4 | 37,1 | 41,7 |
| La Niña | 175 | 11,3 | 13,2 | 38,8 | 40,4 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 82 | ×1,15 | ×1,45 | 22% |
| La Niña | 159 | ×1,41 | ×1,07 | 26% |
