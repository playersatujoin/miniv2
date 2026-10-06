# Fase 2 — Perbandingan A/B: ekologi, pertanian, iklim

- Dunia: peta Starter Island 128×128 (seed peta 1337). **16 dunia per konfigurasi** (seed 1–16), masing-masing **120 menit simulasi ≈ 900 tahun**. Tanpa manusia: 8 dunia × 240 menit ≈ 1.800 tahun.
- Laporan lengkap per konfigurasi: `on/`, `no-farming/`, `no-climate/`, `no-humans/` (masing-masing `summary.md` + data lengkap `summary.json.gz`).
- Dibuat dengan `go run ./cmd/soak -seeds 1-16 -minutes 120 [-off farming|climate]` dan `go run ./cmd/soak -seeds 1-8 -minutes 240 -off humans`.
- Definisi setiap angka: `docs/reference-ecology.md` §8 dan `docs/reference-demography.md`.

| Konfigurasi | Arti |
| --- | --- |
| **on** | Fase 2 penuh: musim, El Niño/La Niña, tumbuhan, satwa, ikan, pertanian, peternakan, air tawar |
| **no-farming** (`-off farming`) | Tidak ada yang bisa menanam; pemburu-peramu saja |
| **no-climate** (`-off climate`) | Hujan rata sepanjang tahun: tanpa musim kemarau, El Niño, La Niña, dan banjir |
| **no-humans** (`-off humans`) | Satwa dan tumbuhan saja, tanpa Adam & Hawa |

## Hasil utama (median 16 dunia)

| | on | no-farming | no-climate |
| --- | ---: | ---: | ---: |
| Era 1 bertahan sampai akhir | 9/16 (33–77%) | 7/16 (23–67%) | 7/16 (23–67%) |
| Populasi puncak | **100** | 56 | 148 |
| Populasi puncak tertinggi | 994 | 120 | 1.285 |
| Populasi rata-rata seperempat akhir | **59** | 21 | 121 |
| Dunia dengan populasi akhir lebih tinggi dari no-farming | **12/16** | – | – |
| Bagian energi dari ladang (100 tahun terakhir) | 37% | 0% | 73% |
| Pertanian ditemukan (menit) | 16 | – | 21,5 |
| Batas teknis tersentuh (kehamilan batal) | **0** | 0 | 0 |
| Mati kelaparan per 1.000 orang per tahun | 35 | 23 | 38 |
| Harapan hidup saat lahir (e0) | 22,7 | 23,4 | 20,2 |
| Bertahan sampai 15 tahun (l15) | 0,56 | 0,49 | 0,51 |
| Angka kelahiran total (TFR) | 8,9 | 5,5 | 11,9 |
| Rasio kelamin (♂ per 100 ♀) | 114 | 100 | 111 |
| Gini kekayaan | 0,50 | 0,31 | 0,51 |
| ms/tick median (tertinggi) | 0,33 (6,51) | 0,33 (0,71) | 0,91 (8,32) |

Angka dalam kurung pada baris era 1 adalah selang kepercayaan 95% (Wilson). Antar-putaran soak dengan 16 dunia, angka seperti era 1 dan populasi puncak bisa bergeser cukup jauh (misalnya era 1 dengan pertanian 9–12/16, puncak 100–214), jadi bandingkan arah dan besarnya, bukan angka persisnya.

## Apa artinya

1. **Pertanian menaikkan daya dukung sekitar 2–3×.**
   - Populasi puncak median 100 lawan 56, populasi akhir 59 lawan 21, dan lebih tinggi di 12 dari 16 dunia.
   - Di dunia yang pertaniannya berkembang penuh, ladang memberi 60–90% energi, dan populasi mencapai 300–994 orang.
   - Kelangsungan era 1 tidak berbeda nyata (9/16 lawan 7/16).
2. **Petani lebih subur dan lebih timpang**, sesuai catatan arkeologi dan antropologi.
   - TFR petani 8,9 lawan 5,5 pada pemburu-peramu. Ini "transisi demografi Neolitikum" (Bocquet-Appel 2011). Angka petani masih di atas data nyata (5–7) karena belum ada penyakit dan peluang kembar masih 20%.
   - Gini 0,50 lawan 0,31. Pemburu-peramu nyata sekitar 0,25 dan petani 0,48 (Borgerhoff Mulder dkk. 2009).
3. **Iklim ikut membatasi populasi.** Tanpa musim kemarau dan El Niño, ladang selalu cukup air, sehingga populasi akhir 2× lebih besar (121 lawan 59) dan puncak tertinggi 1.285.
4. **El Niño terlihat sebagai lonjakan kelaparan,** dibanding tahun netral ± 10 tahun di dunia yang sama:

| | Tahun El Niño | Tahun sesudahnya | Kejadian dengan lonjakan > 1,5× |
| --- | ---: | ---: | ---: |
| Petani (on) | ×1,10 | **×1,23** | 27% |
| Pemburu-peramu (no-farming) | ×1,20 | **×1,42** | 37% |
| La Niña, petani (pembanding) | ×1,09 | ×1,09 | 23% |

   - Di tahun El Niño, ladang yang mati kekeringan naik dari 0,02 menjadi 1,07 petak per tahun. Panen per orang setahun sesudahnya turun 18% dibanding tahun sesudah tahun netral.
   - Dampaknya baru terasa setahun kemudian, karena cadangan tubuh bertahan beberapa tahun.
5. **Satwa tanpa manusia berosilasi dan bertahan, kecuali harimau.**

| Spesies | Ada di pulau (median waktu) | Terendah | Masih ada di akhir |
| --- | ---: | ---: | ---: |
| Rusa | 100% | 100% | 8/8 |
| Babi hutan | 100% | 100% | 8/8 |
| Ayam hutan | 95% | 86% | 7/8 |
| Kerbau liar | 76% | 61% | 7/8 |
| Harimau | 49% | 45% | 4/8 |

   - Pulau ini hanya menampung 2–5 harimau. Populasi sekecil itu pasti punah karena kebetulan, berapa pun aturan teritorinya (sudah dicoba: 1, 2, 3 ekor dewasa per teritori). Harimau hidup sebagai populasi yang punah lalu datang lagi dari seberang laut (5% per tahun).
   - Kerbau dan ayam hutan juga kadang punah lalu datang lagi.
6. **Perburuan tidak memunahkan satwa.**
   - Sejak Fase 2b ada jerat, cara mengendap, buruan yang terluka melambat, dan tombak yang lebih mematikan. Dalam 16 dunia × 900 tahun tertangkap 986 rusa, 361 babi, 375 ayam, dan 6 kerbau (sebelumnya 10 rusa dan 16 babi). Sebagian besar lewat jerat.
   - Daging tetap di bawah 0,5% energi: daging segar membusuk dalam beberapa tahun simulasi (waktu paruh 0,5 tahun), dan satu rusa (6 unit) sudah setara makanan satu orang selama sekitar 12 tahun.
   - Rusa dan babi tetap ada 100% waktu di 16/16 dunia, jadi kepunahan karena perburuan berlebihan **belum** terlihat. Lihat `docs/reference-ecology.md` §9.

## Perubahan yang membuat hasil ini (dibanding laporan Fase 2 sebelumnya)

| Perubahan | Sebelum | Sesudah |
| --- | --- | --- |
| Keluarga petani menyimpan benih dan membawanya dari lumbung; menanam di tile sebelah juga boleh | Ladang ≈ 1% energi; puncak dengan pertanian 406 < tanpa 444 | Ladang 50%; puncak 175 > 54 |
| Hanya air tawar yang bisa diminum, sungai dihitung sampai muaranya, haus 8×, ingatan sumber air, sumur butuh pertanian | Pulau menampung ≈ 1.500 pemburu-peramu; populasi menempel di batas teknis | Pemburu-peramu ≈ 30–150 orang, hidup di tepi sungai |
| Perempuan dewasa butuh 78% energi dan air laki-laki (mulai pubertas) | Rasio kelamin 152 | 100–114 (72%: 73–104; 84%: 110–120) |
| Air tanah tepi sungai turun tajam saat kemarau panjang | Panen tidak turun di tahun El Niño | Petak layu di tahun El Niño; panen tahun berikutnya −18% |
| Perbandingan keahlian antar-tetangga lewat array, bukan map | 30% waktu CPU di `teacherOrStudentNear` | ≈ 1% |
| Bayi yang digendong tidak lagi diletakkan di atas air; jeda antar-peristiwa ikut disimpan | Dunia yang disimpan lalu dimuat ulang menyimpang dari aslinya | Identik (`TestSameSeedAndSaveRestoreReplayExactly`) |
| Fase 2b: jerat, mengendap, buruan terluka melambat, tombak 3× lebih mematikan untuk berburu | 10 rusa diburu | 986 rusa diburu, daging tetap < 0,5% |

## Belum terpenuhi / masalah yang diketahui

- **Perburuan tidak memunahkan satwa** (lihat butir 6). Daging membusuk cepat dan sangat berharga dibanding kebutuhan karena waktu dimampatkan, sehingga perburuan tidak pernah menjadi sumber pangan utama. Untuk menguji kepunahan karena perburuan perlu daging per hewan yang jauh lebih kecil, atau pengawetan yang dipakai luas.
- **Kematian dewasa hampir seluruhnya karena lapar** dengan laju sekitar 5% per tahun di semua umur. Karena itu modus usia kematian dewasa 18 tahun (acuan 68–78), dan e15 16–23 tahun (acuan 28–43). Di masyarakat nyata, populasi lebih banyak diatur lewat kesuburan dan penyakit → **Fase 3**.
- **Kematian bayi 0** dan peluang kembar 20% (nyata 1–2%) membuat TFR petani terlalu tinggi → Fase 3.
- **Haus masih sedikit lebih lambat dari lapar** (sekitar 4 lawan 3 tahun). Haus yang lebih cepat menggagalkan pendiri sebelum mereka mengenal jalan ke sungai.
- **Rasio kelamin peka terhadap faktor kebutuhan perempuan:** 72% → 73–104, 78% → 100–114, 84% → 110–120. Nilai 78% dipakai, di antara rasio energi (71%) dan massa tubuh (84%) suku Hadza.
- **Performa:** median 0,33 ms/tick (≈ 13% satu core pada 20×), tapi dunia petani berpenduduk sekitar 1.000 butuh 6,5 ms/tick, di atas anggaran 2,5 ms/tick. Hambatan berikutnya: otak (`think`, 17%) dan pemrosesan tetangga dalam `sense` (≈ 30%).
- **Teknologi melambat:** Zaman Logam hanya tercapai di 2/16 dunia (Fase 1: 7/16), karena populasi kini jauh lebih kecil. Ini sejalan dengan teori bahwa populasi kecil sulit mempertahankan teknologi (Henrich 2004).

## Referensi tambahan

- Bocquet-Appel, J.-P. (2011). When the world's population took off: the springboard of the Neolithic Demographic Transition. *Science* 333.
- Borgerhoff Mulder, M. dkk. (2009). Intergenerational wealth transmission and the dynamics of inequality in small-scale societies. *Science* 326.
- Henrich, J. (2004). Demography and cultural evolution: how adaptive cultural processes can produce maladaptive losses — the Tasmanian case. *American Antiquity* 69.
- Naylor, R. L. dkk. (2007). Assessing risks of climate variability and climate change for Indonesian rice agriculture. *PNAS* 104.
- Pontzer, H. dkk. (2012). Hunter-gatherer energetics and human obesity. *PLoS ONE* 7(7): e40503.
