# Fase 3 (3a + 3b) — Perbandingan A/B: tubuh, kelahiran, dan penyakit menular

- Dunia: peta Starter Island 128×128 (seed peta 1337).
  - **Jangka panjang**: 16 dunia per konfigurasi (seed 1–16), masing-masing 120 menit simulasi ≈ 900 tahun.
  - **Pendiri**: 32 dunia (seed 1–32) × 20 menit ≈ 150 tahun, untuk melihat apakah Adam & Hawa dan keturunannya bertahan.
- Laporan lengkap per konfigurasi (`summary.md` + data `summary.json.gz`):
  - jangka panjang: `on/`, `no-disease/`, `no-sanitation/`;
  - pendiri: `founders-on/`, `founders-no-disease/`.
- Dibuat dengan `go run ./cmd/soak -seeds 1-16 -minutes 120 [-off disease|sanitation]` dan `-seeds 1-32 -minutes 20`.
- Definisi angka: `docs/reference-demography.md`.

| Konfigurasi | Arti |
| --- | --- |
| **on** | Fase 3a + 3b penuh: penuaan Gompertz, kesuburan bulanan, menyusui, menopause, kematian ibu dan neonatal, pengasuhan anak, diare, malaria, ISPA, cacingan, jamban |
| **no-disease** (`-off disease`) | Tanpa penyakit menular dan kematian neonatal (tahap hidup tetap ada) |
| **no-sanitation** (`-off sanitation`) | Jamban tidak menahan kotoran, dan sumur tidak lebih bersih dari lubang di tepi sungai |

## Hasil utama (median)

| | on | no-disease | no-sanitation | Acuan pra-modern |
| --- | ---: | ---: | ---: | ---: |
| Era 1 bertahan 900 tahun | 9/16 | 6/16 | 8/16 | – |
| Era 1 bertahan 150 tahun (32 dunia) | 25/32 | 26/32 | – | – |
| Harapan hidup saat lahir (e0) | **26,3** ✓ | 28,2 | 24,4 | 21–37 |
| Bertahan sampai 15 (l15) | **0,69** ✓ | 0,73 | 0,70 | 0,44–0,73 |
| Kematian bayi (q0) | 0,074 | 0,000 | 0,096 | 0,13–0,41 |
| Kematian balita (5q0) | 0,135 | 0,005 | 0,170 | – |
| TFR | 7,1 | 5,8 | 6,8 | 5–7 |
| Jarak antar-kelahiran | **3,2** ✓ | 3,8 | 3,3 | 2,8–3,3 |
| Diare per orang per tahun | **0,23** | – | 0,65 | – |
| Kematian karena diare per 1.000 orang-tahun | **2,4** | – | 4,9 | – |
| Jamban dibangun (16 dunia) | 51 | – | 43 | – |
| Penyakit + neonatal, bagian kematian | 22% | 0% | 28% | > 50% |
| Bagian kematian balita di semua kematian | 15% | 2% | 20% | – |

Dengan 16 dunia, angka seperti era 1 bisa bergeser beberapa dunia antar-putaran. Bandingkan arah dan besarnya, bukan angka persisnya.

## Dunia padat

Dunia kecil di atas jarang lebih dari 50–80 orang. Karena itu model juga diuji di dua dunia padat yang sudah lama berjalan, masing-masing dimuat dari simpanan lalu dijalankan tanpa antarmuka:

| | Dunia uji (748 orang) | Salinan dunia live (1.260 orang) |
| --- | --- | --- |
| Penduduk setelah 40–60 tahun | sekitar 650–830 | sekitar 885–980 (tanpa penyakit sekitar 1.050–1.130) |
| Paling banyak sakit sekaligus | diare sekitar 6%, malaria sekitar 9%, ISPA 6–13% | sama |
| Penyakit sebagai bagian kematian | sekitar 50% | sekitar 50% |
| Kelaparan per tahun | – | 19 (tanpa penyakit 29) |
| A/B sanitasi, 60 tahun | kematian karena diare −20%, insiden −17% (5 jamban) | – |

## Apa artinya

1. **Kematian anak kini datang dari penyakit, seperti di dunia nyata.**
   - Tanpa penyakit, balita hampir tidak pernah mati (5q0 0,005) dan anak-anak mati kelaparan.
   - Dengan penyakit, 5q0 0,135, l15 0,69, dan sebagian besar kematian balita karena diare, malaria, dan kematian neonatal.
2. **Sanitasi bekerja tanpa ada yang tahu alasannya.** Jamban yang dibangun keluarga petani menahan kotoran dari air dan halaman, sehingga kematian karena diare turun separuhnya.
3. **Wabah bergantung kepadatan.** Kelompok pendiri hampir tidak pernah terkena ISPA, karena galur baru jarang muncul di kelompok kecil. Di dunia padat, ISPA adalah pembunuh terbesar di antara penyakit, dan penyakit sebagian menggantikan kelaparan sebagai pengendali jumlah penduduk.
4. **Yang belum realistis.** Orang dewasa masih terlalu sering mati kelaparan dan kehausan (modus umur wafat dewasa 18 tahun). Diagnosis menunjukkan mereka mati bersama keluarga di tempat yang makanannya habis, tanpa pindah atau berbagi. Ini pekerjaan Fase 3d (gizi) dan Fase 4 (berbagi pangan).
