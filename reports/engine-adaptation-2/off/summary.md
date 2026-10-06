# Laporan soak — engine II mati

- Dibuat: 2026-10-06 20:04:53
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 20 menit simulasi per dunia (≈ 150 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: stimuli, affect, exchange, fire, mobility, villages
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 20 m (utuh) | 1 | 18 | 8 | 0 | 1 | 2 | 30,1 | 0,78 | 22,3 | 5,6 | 3,4 | 20,9 | – | 0,21 | 0 | 0,19 |
| 2 | 20 m (utuh) | 1 | 24 | 7 | 0 | 1 | 4 | 28,0 | 0,57 | 29,1 | 5,6 | 4,7 | 20,2 | 18 | 0,27 | 181 | 0,38 |
| 3 | 12 m | 2 | 7 | 4 | 0 | 1 | 0 | 25,9 | 0,71 | 18,2 | 5,4 | 3,4 | – | – | 0,36 | 0 | 0,20 |
| 4 | 20 m (utuh) | 1 | 25 | 7 | 1 | 4 | 2 | 32,4 | 0,83 | 23,1 | 6,6 | 3,5 | 20,8 | 18 | 0,31 | 0 | 0,22 |
| 5 | 18 m | 2 | 4 | 1 | 0 | 2 | 1 | – | – | – | – | – | – | – | 0,04 | – | 0,24 |
| 6 | 20 m (utuh) | 1 | 43 | 8 | 0 | 1 | 4 | 30,8 | 0,75 | 23,5 | 3,9 | 4,8 | 20,5 | 18 | 0,47 | 48 | 0,31 |
| 7 | 20 m (utuh) | 1 | 36 | 8 | 1 | 4 | 4 | 25,9 | 0,71 | 18,3 | 6,1 | 3,5 | 16,9 | 18 | 0,39 | 0 | 0,23 |
| 8 | 20 m (utuh) | 1 | 21 | 7 | 0 | 3 | 4 | 29,8 | 0,71 | 24,4 | 3,8 | 4,1 | 21,1 | 32 | 0,29 | 1165 | 0,32 |
| **Median** | 20 m | | 22 | 7 | 0 | 2 | 3 | 29,8 | 0,71 | 23,1 | 5,6 | 3,5 | 20,6 | 18 | 0,30 | 0 | 0,23 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **6 dari 8** dunia (selang kepercayaan 95%: 41–93%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | 29,8 tahun | 21–37 | ✓ dalam rentang | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | 0,71 | 0,44–0,73 | ✓ dalam rentang | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | 23,1 tahun | 28–43 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | 18 tahun | 68–78 | ↓ di bawah | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | 5,6 anak | 5–7 | ✓ dalam rentang | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | 3,5 tahun | 2,8–3,3 | ↑ di atas | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | 20,6 tahun | 18–20 | ↑ di atas | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | 0,30 | 0,21–0,29 | ↑ di atas | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |
| Kematian bayi (q0) | 0,081 | 0,13–0,41 | ↓ di bawah | Volk & Atkinson (2013), 20 populasi pemburu-peramu — rata-rata 0,27 (SD 0,07); rentang ±2 SD |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian balita (5q0) | 0,119 |
| Diare per orang per tahun | 0,38 |
| Malaria per orang per tahun | 1,99 |
| ISPA per orang per tahun | 0,01 |
| Rasio kelamin (♂ per 100 ♀) | 85 |
| Anggota per rumah | 7,5 |
| Pembunuhan per 100.000 tahun-orang | 0 |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian | Balita (<5) | Bagian |
| --- | ---: | ---: | ---: | ---: |
| Kelaparan | 86 | 28% | 0 | 0% |
| Kehausan | 78 | 26% | 3 | 8% |
| Usia tua | 53 | 17% | 0 | 0% |
| Dibunuh | 17 | 6% | 2 | 5% |
| Diterkam hewan | 0 | 0% | 0 | 0% |
| Melahirkan | 4 | 1% | 0 | 0% |
| Neonatal (minggu pertama) | 17 | 6% | 17 | 42% |
| Diare | 22 | 7% | 11 | 28% |
| Malaria | 25 | 8% | 7 | 18% |
| Radang paru (ISPA) | 1 | 0% | 0 | 0% |
| Terbakar | 0 | 0% | 0 | 0% |
| Tenggelam | 0 | 0% | 0 | 0% |
| Terjatuh | 0 | 0% | 0 | 0% |
| **Penyakit menular** | 48 | 16% | 18 | 45% |

Acuan: di masyarakat pemburu-peramu penyakit menyebabkan sekitar 70% kematian (Gurven & Kaplan 2007), terutama pada anak.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | – | – | 0 | 8 | 26,1 | 0,78 |
| 2 | – | – | 0 | 5 | 35,0 | 0,41 |
| 3 | – | – | 0 | 8 | 32,4 | 0,23 |
| 4 | 4 | – | 1 | 3 | 30,6 | 0,56 |
| 5 | – | – | 0 | 5 | 28,5 | 0,26 |
| 6 | – | – | 0 | 4 | 28,9 | 0,47 |
| 7 | 6 | – | 1 | 2 | 32,8 | 0,48 |
| 8 | – | – | 0 | 3 | 34,7 | 0,83 |
| **Median** | 5 | – | 0 | 4 | 31,5 | 0,47 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | 0,44 | 77 |
| 15–29 | 0,47 | 43 |
| 30–44 | 0,49 | 30 |
| 45–59 | 0,47 | 15 |
| 60+ | 0,49 | 13 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 30,4 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 18 | 0 | 5 | 12 | 0,99 | 126 / 116 / 93 / 18 / 2 | 0 | 4 / 3 |
| 2 | 41 | 0 | 11 | 5 | 0,99 | 123 / 183 / 96 / 34 / 0 | 0 | 6 / 5 |
| 3 | 18 | 0 | 17 | 5 | 0,99 | 76 / 199 / 114 / 26 / 4 | 0 | 4 / 4 |
| 4 | 25 | 0 | 12 | 3 | 0,99 | 130 / 93 / 66 / 0 / 0 | 0 | 9 / 7 |
| 5 | 18 | 0 | 14 | 6 | 1,00 | 156 / 175 / 46 / 2 / 7 | 0 | 2 / 2 |
| 6 | 49 | 0 | – | 0 | 0,99 | 48 / 205 / 105 / 54 / 3 | 0 | 3 / 3 |
| 7 | 43 | 0 | 10 | 7 | 0,99 | 136 / 184 / 62 / 3 / 0 | 0 | 8 / 7 |
| 8 | 40 | 0 | 9 | 29 | 0,99 | 119 / 202 / 70 / 9 / 0 | 0 | 4 / 3 |
| **Median** | 32 | 0 | 11 | 6 | 0,99 | | | 4 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 100% | 8/8 |
| Ayam Hutan | 100% | 100% | 8/8 |
| Kerbau Liar | 100% | 45% | 7/8 |
| Harimau | 68% | 45% | 4/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 87% | 12% | 0% | 0% | 0,36 |
| 2 | 99% | 0% | 0% | 1% | 0,00 |
| 3 | 97% | 2% | 0% | 1% | 0,02 |
| 4 | 88% | 0% | 10% | 2% | 0,01 |
| 5 | 97% | 0% | 3% | 0% | 0,28 |
| 6 | 100% | 0% | 0% | 0% | 0,00 |
| 7 | 89% | 4% | 4% | 3% | 0,04 |
| 8 | 87% | 11% | 0% | 2% | 0,41 |
| **Median** | | 1% | | | |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 64 | 12,9 | 10,3 | 37,3 | 10,4 |
| Netral | 266 | 10,9 | 11,4 | 37,3 | 6,2 |
| La Niña | 111 | 10,6 | 11,2 | 35,3 | 5,9 |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 50 | ×2,02 | ×1,33 | 31% |
| La Niña | 91 | ×1,36 | ×1,55 | 20% |

## Engine II: pertukaran, desa, api, air

| Seed | Bicara | Barter | Kabar | Desa | Warga desa | Ganti pemimpin | Kebakaran | Bangunan terbakar | Rakit dibawa | Tenggelam | Terjatuh | Stimulus aktif |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 2 | 0 | 0 | 0 |
| 2 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 4 | 0 | 0 | 0 |
| 3 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| 4 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 2 | 0 | 0 | 0 |
| 5 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| 6 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| 7 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| 8 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
