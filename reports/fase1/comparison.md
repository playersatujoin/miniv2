# Fase 1 — Perbandingan A/B: belajar & budaya

- Dunia: peta Starter Island 128×128 (seed peta 1337), **16 dunia per konfigurasi** (seed 1–16), masing-masing **120 menit simulasi ≈ 900 tahun ≈ 34 generasi** (1 tahun = 8 detik).
- Laporan lengkap per konfigurasi: `on/`, `off/`, `no-plasticity/`, `no-culture/` (masing-masing `summary.md` + `summary.json`).
- Dibuat dengan `go run ./cmd/soak -seeds 1-16 -minutes 120 [-off …]`.

| Konfigurasi | Arti |
| --- | --- |
| **on** | Fase 1 penuh: otak belajar selama hidup (plastisitas) + keahlian per orang, mengajar, belajar di rumah, mengamati, perpustakaan |
| **off** (`-off learning`) | Model lama: otak tetap seumur hidup; apa pun yang pernah ditemukan siapa pun langsung bisa dilakukan semua orang |
| **no-plasticity** | Budaya per orang saja, otak tidak belajar |
| **no-culture** | Otak belajar, tapi pengetahuan tetap milik semua orang |

## Hasil utama (median 16 dunia)

| | on | off | no-plasticity | no-culture |
| --- | ---: | ---: | ---: | ---: |
| Era 1 bertahan sampai akhir | **11/16** (44–86%) | 9/16 (33–77%) | 7/16 (23–67%) | 11/16 (44–86%) |
| Dunia yang mencapai Zaman Logam | 7/16 | **16/16** | 4/16 | **16/16** |
| Generasi saat Zaman Logam tercapai | 10 | 11,5 | 14 | **7** |
| Dunia yang mencapai Zaman Kimia | 0/16 | 0/16 | 0/16 | 0/16 |
| Unsur ditemukan | 5 | 6 | 4,5 | 6 |
| Populasi akhir | 146 | 147 | 145 | 146 |
| Kejadian "pengetahuan hilang" per dunia | 32 | – | 35 | – |
| Rata-rata keahlian terbaik orang dewasa | 0,44 | – | 0,45 | – |
| Rata-rata neuron tersembunyi | 19,9 | 20,0 | 19,8 | 19,9 |

Angka dalam kurung adalah selang kepercayaan 95% (Wilson). Dengan 16 dunia, selisih kelangsungan hidup antar-konfigurasi belum signifikan secara statistik.

## Apa artinya

1. **Belajar selama hidup membantu.**
   - Dengan plastisitas, era 1 bertahan di 11/16 dunia, dibanding 9/16 (model lama) dan 7/16 (budaya tanpa plastisitas).
   - Jika pengetahuan milik semua orang, Zaman Logam tercapai **4,5 generasi lebih cepat** (generasi 7 vs 11,5).
   - Arahnya konsisten di semua perbandingan, tapi masih dalam selang kepercayaan.
2. **Budaya per orang memperlambat teknologi, dan itu realistis.**
   - Ketika pengetahuan harus diajarkan dan bisa mati bersama pemiliknya, hanya 7/16 dunia yang sampai Zaman Logam (vs 16/16 di model lama, yang secara tidak realistis membagikan semua pengetahuan gratis ke semua orang).
   - Di populasi kecil (dibatasi sekitar 148 orang), pengetahuan sering hilang (median 32 kejadian per dunia). Ini sejalan dengan teori evolusi budaya bahwa populasi kecil yang terisolasi bisa kehilangan teknologi (Henrich 2004, kasus Tasmania).
3. **Keahlian naik seiring umur**, tapi landai, karena anak sudah banyak belajar dari keluarganya:

   | Umur | 0–14 | 15–29 | 30–44 | 45–59 | 60+ |
   | --- | ---: | ---: | ---: | ---: | ---: |
   | Median keahlian terbaik (on) | 0,37 | 0,43 | 0,45 | 0,44 | 0,44 |

   Sumber keahlian (dari uji awal): terbanyak lewat **diajar** dan **belajar dari keluarga di rumah**, lalu latihan; menemukan sendiri jarang.
4. **Ukuran otak tidak berubah** dalam 34 generasi: rata-rata tetap sekitar 20 neuron di semua konfigurasi dan semua kelompok generasi (0–9: 20,0 · 10–19: 19,9 · 20–29: 19,8 · 30–39: 20,0). Tekanan seleksinya terlalu lemah untuk terlihat dalam waktu sesingkat ini. Ini temuan jujur, bukan bukti bahwa ukuran otak tidak penting.

## Kriteria penerimaan (PLAN.md, Fase 1)

| Kriteria | Hasil |
| --- | --- |
| Median keahlian naik seiring umur | ✓ (0,37 → 0,45), landai |
| Kejadian "pengetahuan hilang" terjadi | ✓ (median 32 per dunia) |
| A/B: zaman ≥ 2 lebih cepat dengan belajar (target: Zaman Kimia di ≥ 4/8 dunia dalam 30 generasi) | ✗ **Tidak tercapai.** Tidak satu pun konfigurasi mencapai Zaman Kimia, termasuk model lama. Penghambat utamanya ada di luar Fase 1: batas populasi keras sekitar 148 orang dalam satu permukiman (Fase 2), dan rantai bahan laboratorium (kaca + besi + bata) yang panjang. Dibanding model lama, budaya per orang justru memperlambat Zaman Logam; plastisitas mempercepatnya |
| Distribusi ukuran otak dilaporkan per generasi | ✓ (tidak berubah) |
| Performa ≤ 1 core pada 20× | ✓ 0,61 ms/tick untuk 140 makhluk (0,49 tanpa belajar) ≈ 24% satu core pada 20× |

## Demografi dibanding baseline-v0

| Indikator | baseline-v0 | on | off |
| --- | ---: | ---: | ---: |
| Era 1 bertahan | ~50% (24/48) | 11/16 | 9/16 |
| e0 (harapan hidup saat lahir) | 46,1 | 47,1 | 47,1 |
| l15 (hidup sampai 15) | 0,92 | 0,94 | 0,90 |
| TFR | 1,4 | 1,8 | 1,4 |
| Gini | 0,54 | 0,51 | 0,48 |
| Pembunuhan per 100.000 tahun-orang | 266 | 103 | 239 |

Konfigurasi "off" tidak identik dengan baseline-v0: otaknya sekarang punya 60 indra dan 13 keputusan, dan sebagian metabolisme basal kini dihitung sebagai biaya otak. Angka pembunuhan lebih rendah dengan budaya (103 vs 239) teramati tapi belum dianalisis penyebabnya.

## Batasan

- 16 dunia per konfigurasi masih menghasilkan selang kepercayaan lebar. Selisih kecil bisa kebetulan.
- Imbalan belajar hanya dari tubuh (energi, air, kesehatan, rasa sakit), sengaja tanpa imbalan tugas. Akibatnya plastisitas hanya membantu hal yang langsung terasa di tubuh.
- Mengajar sepenuhnya keputusan otak (output `ajar`). Generasi pertama punya refleks naluri yang lemah untuk mengajar saat ada yang bisa diajari.
