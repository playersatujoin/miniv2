# Laporan soak — without-engine-adaptation

- Dibuat: 2026-10-06 18:01:10
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 20 menit simulasi per dunia (≈ 150 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: perception, navigation, ai_budget
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 20 m (utuh) | 1 | 11 | 7 | 0 | 1 | 3 | 24,9 | 0,54 | 24,5 | 4,2 | 3,6 | 18,9 | – | 0,26 | 779 | 0,15 |
| 2 | 20 m (utuh) | 1 | 65 | 8 | 0 | 1 | 4 | 22,7 | 0,64 | 16,1 | 6,6 | 3,2 | 17,8 | 18 | 0,40 | 25 | 0,47 |
| 3 | 20 m (utuh) | 1 | 31 | 7 | 0 | 1 | 3 | 45,9 | 0,88 | 36,7 | 3,8 | 3,3 | 18,6 | – | 0,50 | 0 | 0,18 |
| 4 | 20 m (utuh) | 1 | 27 | 7 | 0 | 1 | 1 | 40,5 | 0,78 | 36,2 | 6,1 | 3,9 | 21,9 | – | 0,46 | 0 | 0,16 |
| 5 | 20 m (utuh) | 1 | 16 | 9 | 0 | 1 | 2 | 31,9 | 0,67 | 28,9 | 4,8 | 3,7 | 22,7 | 18 | 0,24 | 0 | 0,17 |
| 6 | 20 m (utuh) | 1 | 1 | 3 | 0 | 1 | 0 | – | – | – | – | – | – | – | – | – | 0,12 |
| 7 | 20 m (utuh) | 1 | 62 | 8 | 0 | 1 | 1 | 33,3 | 0,82 | 23,8 | 7,5 | 3,4 | 17,3 | 18 | 0,45 | 124 | 0,29 |
| 8 | 20 m (utuh) | 1 | 1 | 5 | 0 | 1 | 0 | – | – | – | 0,0 | – | – | – | – | – | 0,11 |
| **Median** | 20 m | | 22 | 7 | 0 | 1 | 2 | 32,6 | 0,72 | 26,7 | 4,8 | 3,5 | 18,8 | 18 | 0,43 | 12 | 0,16 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **8 dari 8** dunia (selang kepercayaan 95%: 68–100%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 32,6 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,72 | 0,44–0,73 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 26,7 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 18 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 4,8 anak | 5–7 | ↓ di bawah | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,5 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 18,8 tahun | 18–20 | ✓ dalam rentang | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,43 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,056 | 0,13–0,41 | ↓ di bawah | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,095 |
| Diare per orang per tahun | 0,34 |
| Malaria per orang per tahun | 0,57 |
| ISPA per orang per tahun | 0,01 |
| Rasio kelamin (♂ per 100 ♀) | 78 |
| Anggota per rumah | 5,3 |
| Pembunuhan per 100.000 tahun-orang | 12 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 143 | 41% | 0 | 0% |
| Kehausan | 81 | 23% | 0 | 0% |
| Usia tua | 41 | 12% | 0 | 0% |
| Dibunuh | 13 | 4% | 1 | 3% |
| Diterkam hewan | 0 | 0% | 0 | 0% |
| Melahirkan | 3 | 1% | 0 | 0% |
| Neonatal (minggu pertama) | 19 | 5% | 19 | 50% |
| Diare | 20 | 6% | 9 | 24% |
| Malaria | 26 | 7% | 9 | 24% |
| Radang paru (ISPA) | 1 | 0% | 0 | 0% |
| **Penyakit menular** | 47 | 14% | 18 | 47% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | – | – | 0 | 1 | 34,5 | 0,60 |
| 2 | – | – | 0 | 1 | 29,1 | 0,90 |
| 3 | – | – | 0 | 0 | 31,5 | 0,87 |
| 4 | – | – | 0 | 3 | 27,9 | 0,55 |
| 5 | – | – | 0 | 1 | 33,7 | 0,63 |
| 6 | – | – | 0 | 2 | 66,0 | 0,72 |
| 7 | – | – | 0 | 4 | 24,8 | 0,96 |
| 8 | – | – | 0 | 2 | 23,0 | 0,59 |
| **Median** | – | – | 0 | 2 | 30,3 | 0,68 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,70 | 95 |
| 15–29 | 0,76 | 52 |
| 30–44 | 0,72 | 40 |
| 45–59 | 0,74 | 13 |
| 60+ | 0,72 | 14 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 29,7 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 24 | 0 | – | 0 | 0,99 | 132 / 170 / 48 / 5 / 3 | 0 | 7 / 7 |
| 2 | 106 | 0 | 3 | 61 | 0,98 | 126 / 170 / 79 / 3 / 1 | 0 | 6 / 6 |
| 3 | 31 | 0 | 6 | 10 | 0,99 | 106 / 186 / 122 / 16 / 4 | 0 | 2 / 2 |
| 4 | 27 | 0 | 16 | 0 | 0,99 | 123 / 100 / 22 / 0 / 5 | 0 | 11 / 10 |
| 5 | 24 | 0 | 6 | 1 | 0,99 | 129 / 168 / 0 / 26 / 5 | 0 | 5 / 4 |
| 6 | 12 | 0 | – | 0 | 1,00 | 122 / 131 / 73 / 3 / 3 | 0 | 9 / 9 |
| 7 | 62 | 0 | 9 | 37 | 0,99 | 137 / 74 / 75 / 2 / 1 | 0 | 8 / 8 |
| 8 | 7 | 0 | – | 0 | 1,00 | 121 / 94 / 60 / 45 / 0 | 0 | 4 / 3 |
| **Median** | 26 | 0 | 6 | 0 | 0,99 | | | 6 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 100% | 8/8 |
| Ayam Hutan | 100% | 85% | 7/8 |
| Kerbau Liar | 75% | 45% | 7/8 |
| Harimau | 62% | 45% | 7/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 100% | 0% | 0% | 0% | 0,00 |
| 2 | 45% | 54% | 0% | 0% | 0,82 |
| 3 | 96% | 4% | 0% | 0% | 0,54 |
| 4 | 100% | 0% | 0% | 0% | 0,00 |
| 5 | 99% | 1% | 0% | 0% | 0,04 |
| 6 | 100% | 0% | 0% | 0% | 0,00 |
| 7 | 45% | 55% | 0% | 0% | 0,88 |
| 8 | 100% | 0% | 0% | 0% | 0,00 |
| **Median** | | 1% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 46 | 16,6 | 15,4 | 35,0 | 76,9 |
| Netral | 151 | 12,6 | 12,3 | 43,7 | 81,0 |
| La Niña | 58 | 13,3 | 16,0 | 44,4 | 71,3 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 29 | ×0,81 | ×1,10 | 31% |
| La Niña | 39 | ×1,20 | ×1,27 | 32% |
