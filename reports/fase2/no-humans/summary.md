# Laporan soak — Fase 2 — tanpa manusia

- Dibuat: 2026-10-06 11:53:08
- Peta: 128×128, seed peta 1337 · seed dunia: 1, 2, 3, 4, 5, 6, 7, 8
- Durasi: 240 menit simulasi per dunia (≈ 1800 tahun; 1 tahun = 8 detik simulasi)
- Aturan dimatikan: humans
- Indikator demografi dihitung dari 50 tahun simulasi terakhir (tabel hidup periode). Acuan dan definisinya: `docs/reference-demography.md`.

## Per dunia

| Seed | Era 1 bertahan | Era | Populasi | Gen. maks | Zaman | Unsur | Rumah | e0 | l15 | e15 | TFR | Jarak lahir | Ibu pertama | Modus mati dewasa | Gini | Pembunuhan /100rb | ms/tick |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 240 m | 0 | 0 | 0 | 0 | 0 | 0 | – | – | – | – | – | – | – | – | – | 0,16 |
| 2 | 240 m | 0 | 0 | 0 | 0 | 0 | 0 | – | – | – | – | – | – | – | – | – | 0,14 |
| 3 | 240 m | 0 | 0 | 0 | 0 | 0 | 0 | – | – | – | – | – | – | – | – | – | 0,16 |
| 4 | 240 m | 0 | 0 | 0 | 0 | 0 | 0 | – | – | – | – | – | – | – | – | – | 0,16 |
| 5 | 240 m | 0 | 0 | 0 | 0 | 0 | 0 | – | – | – | – | – | – | – | – | – | 0,17 |
| 6 | 240 m | 0 | 0 | 0 | 0 | 0 | 0 | – | – | – | – | – | – | – | – | – | 0,17 |
| 7 | 240 m | 0 | 0 | 0 | 0 | 0 | 0 | – | – | – | – | – | – | – | – | – | 0,14 |
| 8 | 240 m | 0 | 0 | 0 | 0 | 0 | 0 | – | – | – | – | – | – | – | – | – | 0,16 |
| **Median** | 240 m | | 0 | 0 | 0 | 0 | 0 | – | – | – | – | – | – | – | – | – | 0,16 |

Era 1 (keturunan Adam & Hawa pertama) bertahan sampai akhir di **0 dari 8** dunia (selang kepercayaan 95%: 0–32%). Hasil per dunia sangat dipengaruhi kebetulan, jadi bandingkan konfigurasi dengan banyak seed.

## Dibanding acuan pra-modern

| Indikator | Median simulasi | Acuan | Status | Sumber |
| --- | ---: | ---: | --- | --- |
| Harapan hidup saat lahir (e0) | – tahun | 21–37 | – | Gurven & Kaplan (2007), pemburu-peramu |
| Peluang hidup sampai umur 15 (l15) | – | 0,44–0,73 | – | Gurven & Kaplan (2007), Tabel 2–3 — rata-rata 0,57 |
| Sisa harapan hidup pada umur 15 (e15) | – tahun | 28–43 | – | Gurven & Kaplan (2007), Tabel 3 |
| Modus usia kematian dewasa | – tahun | 68–78 | – | Gurven & Kaplan (2007), Tabel 4 |
| Angka kelahiran total (TFR) | – anak | 5–7 | – | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 6,2; sumber sekunder |
| Jarak antar-kelahiran | – tahun | 2,8–3,3 | – | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); arXiv:2601.13442 — rata-rata 3,1 tahun |
| Umur ibu saat anak pertama | – tahun | 18–20 | – | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) |
| Ketimpangan kekayaan (Gini) | – | 0,21–0,29 | – | Borgerhoff Mulder dkk. (2009), Tabel 2 — pemburu-peramu 0,25 ± 0,04; simulasi hanya menghitung kekayaan material |

Keterangan: ✓ dalam rentang · ↑ di atas · ↓ di bawah · – belum cukup data.

## Indikator lain (tanpa rentang acuan)

| Indikator | Median simulasi |
| --- | ---: |
| Kematian bayi (q0) | – |
| Rasio kelamin (♂ per 100 ♀) | – |
| Anggota per rumah | – |
| Pembunuhan per 100.000 tahun-orang | – |

## Penyebab kematian (semua dunia, 50 tahun terakhir)

| Penyebab | Kematian | Bagian |
| --- | ---: | ---: |
| Kelaparan | 0 | 0% |
| Kehausan | 0 | 0% |
| Usia tua | 0 | 0% |
| Dibunuh | 0 | 0% |
| Diterkam hewan | 0 | 0% |

Catatan: simulasi belum punya penyakit (Fase 3), sedangkan di masyarakat nyata penyakit menyebabkan lebih dari separuh kematian. Perbedaan ini temuan, bukan galat.

## Budaya dan otak

| Seed | Generasi saat Zaman Logam | Generasi saat Zaman Kimia | Zaman akhir | Pengetahuan hilang | Neuron (rata-rata) | Keahlian dewasa |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | – | – | 0 | 0 | 0,0 | 0,00 |
| 2 | – | – | 0 | 0 | 0,0 | 0,00 |
| 3 | – | – | 0 | 0 | 0,0 | 0,00 |
| 4 | – | – | 0 | 0 | 0,0 | 0,00 |
| 5 | – | – | 0 | 0 | 0,0 | 0,00 |
| 6 | – | – | 0 | 0 | 0,0 | 0,00 |
| 7 | – | – | 0 | 0 | 0,0 | 0,00 |
| 8 | – | – | 0 | 0 | 0,0 | 0,00 |
| **Median** | – | – | 0 | 0 | 0,0 | 0,00 |

Generasi dihitung sebagai generasi tertinggi yang hidup saat zaman itu pertama tercapai (– = tidak tercapai).

### Keahlian menurut umur (median antar-dunia dari median keahlian terbaik)

| Umur | Keahlian | Orang (total) |
| --- | ---: | ---: |
| 0–14 | – | 0 |
| 15–29 | – | 0 |
| 30–44 | – | 0 |
| 45–59 | – | 0 |
| 60+ | – | 0 |

### Ukuran otak menurut generasi (rata-rata neuron tersembunyi, semua dunia)

| Generasi rata-rata | Neuron |
| --- | ---: |
| 0–9 | 0,0 |

## Ekologi

| Seed | Populasi puncak | Batas teknis tersentuh | Pertanian (menit) | Petak maks | Hutan tersisa | Hewan akhir (Rusa / Babi Hutan / Ayam Hutan / Kerbau Liar / Harimau) | Ternak | Punah lokal / datang lagi |
| ---: | ---: | ---: | ---: | ---: | ---: | --- | ---: | ---: |
| 1 | 0 | 0 | – | 0 | 1,00 | 113 / 158 / 0 / 18 / 0 | 0 | 74 / 72 |
| 2 | 0 | 0 | – | 0 | 1,00 | 141 / 115 / 53 / 2 / 2 | 0 | 66 / 66 |
| 3 | 0 | 0 | – | 0 | 1,00 | 88 / 205 / 28 / 23 / 3 | 0 | 58 / 58 |
| 4 | 0 | 0 | – | 0 | 1,00 | 130 / 177 / 21 / 27 / 0 | 0 | 65 / 64 |
| 5 | 0 | 0 | – | 0 | 1,00 | 128 / 168 / 1 / 28 / 0 | 0 | 79 / 78 |
| 6 | 0 | 0 | – | 0 | 1,00 | 137 / 191 / 115 / 13 / 4 | 0 | 94 / 94 |
| 7 | 0 | 0 | – | 0 | 1,00 | 123 / 164 / 112 / 0 / 5 | 0 | 76 / 75 |
| 8 | 0 | 0 | – | 0 | 1,00 | 94 / 192 / 65 / 6 / 0 | 0 | 75 / 74 |
| **Median** | 0 | 0 | – | 0 | 1,00 | | | 74 |

### Keberadaan satwa (bagian waktu spesies itu ada di pulau)

| Spesies | Median | Terendah | Dunia yang masih punya di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi Hutan | 100% | 100% | 8/8 |
| Ayam Hutan | 95% | 86% | 7/8 |
| Kerbau Liar | 76% | 61% | 7/8 |
| Harimau | 49% | 45% | 4/8 |

### Asal pangan (bagian energi yang dimakan, 100 tahun terakhir)

| Seed | Liar | Ladang | Ikan | Daging | Panen per orang per tahun |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | – | – | – | – | – |
| 2 | – | – | – | – | – |
| 3 | – | – | – | – | – |
| 4 | – | – | – | – | – |
| 5 | – | – | – | – | – |
| 6 | – | – | – | – | – |
| 7 | – | – | – | – | – |
| 8 | – | – | – | – | – |

### Iklim dan kelaparan

Tahun dengan ≥ 20 penduduk, semua dunia digabung. Kelaparan dan kelahiran per 1.000 penduduk per tahun.

| Tahun | Jumlah tahun | Mati kelaparan | Mati kelaparan tahun berikutnya | Kelahiran | Panen per 100 orang |
| --- | ---: | ---: | ---: | ---: | ---: |
| El Niño | 0 | – | – | – | – |
| Netral | 0 | – | – | – | – |
| La Niña | 0 | – | – | – | – |

Dibanding tahun netral di sekitarnya (± 10 tahun) di dunia yang sama:

| Tahun | Kejadian | Kelaparan tahun itu | Kelaparan tahun berikutnya | Tahun berikutnya > 1,5× |
| --- | ---: | ---: | ---: | ---: |
| El Niño | 0 | – | – | – |
| La Niña | 0 | – | – | – |
