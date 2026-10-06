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
| Era 1 bertahan sampai akhir | 11/16 (44–86%) | 12/16 (51–90%) | 8/16 (28–72%) |
| Populasi puncak | **175** | 54 | 428 |
| Populasi puncak tertinggi | 444 | 153 | 1.452 |
| Populasi rata-rata seperempat akhir | **93** | 31 | 314 |
| Dunia dengan populasi akhir lebih tinggi dari no-farming | **11/16** | – | – |
| Bagian energi dari ladang (100 tahun terakhir) | 50% | 0% | 84% |
| Pertanian ditemukan (menit) | 14,5 | – | 13 |
| Batas teknis tersentuh (kehamilan batal) | **0** | 0 | 36.746 |
| Mati kelaparan per 1.000 orang per tahun | 31 | 25 | 40 |
| Harapan hidup saat lahir (e0) | 21,5 | 24,0 | 19,1 |
| Bertahan sampai 15 tahun (l15) | 0,52 | 0,51 | 0,49 |
| Angka kelahiran total (TFR) | 8,4 | 5,3 | 11,9 |
| Rasio kelamin (♂ per 100 ♀) | 93 | 73 | 104 |
| Gini kekayaan | 0,45 | 0,43 | 0,53 |
| ms/tick median (tertinggi) | 0,59 (2,29) | 0,33 (0,77) | 1,86 (12,25) |

Angka dalam kurung pada baris era 1 adalah selang kepercayaan 95% (Wilson).

## Apa artinya

1. **Pertanian menaikkan daya dukung sekitar 3×.**
   - Populasi puncak median 175 lawan 54, dan populasi akhir lebih tinggi di 11 dari 16 dunia.
   - Di dunia yang pertaniannya berkembang penuh, ladang memberi 60–90% energi, dan populasi mencapai 300–444 orang.
   - Kelangsungan era 1 tidak berbeda (11/16 lawan 12/16). Pada putaran soak sebelumnya, dengan aturan yang hampir sama, angkanya 12/16 lawan 7/16. Dengan 16 dunia, selisih seperti itu masih kebetulan.
2. **Petani lebih subur**, sesuai catatan arkeologi.
   - TFR petani 8,4 lawan 5,3 pada pemburu-peramu. Ini "transisi demografi Neolitikum" (Bocquet-Appel 2011). Angka petani masih di atas data nyata (5–7) karena belum ada penyakit dan peluang kembar masih 20%.
   - Gini petani sedikit lebih tinggi (0,45 lawan 0,43). Masyarakat pertanian nyata jauh lebih timpang (Borgerhoff Mulder dkk. 2009: pemburu-peramu 0,25, petani 0,48).
3. **Iklim ikut membatasi populasi.**
   - Tanpa musim kemarau dan El Niño, ladang selalu cukup air, sehingga populasi petani 2,4× lebih besar dan menyentuh batas teknis (36.746 kehamilan batal).
   - Dengan iklim, batas itu tidak pernah tersentuh.
4. **El Niño terlihat sebagai lonjakan kelaparan,** dibanding tahun netral ± 10 tahun di dunia yang sama:

| | Tahun El Niño | Tahun sesudahnya | Kejadian dengan lonjakan > 1,5× |
| --- | ---: | ---: | ---: |
| Petani (on) | ×1,10 | **×1,29** | 31% |
| Pemburu-peramu (no-farming) | ×1,25 | **×1,49** | 37% |
| La Niña, petani (pembanding) | ×1,10 | ×1,07 | 24% |

   - Di tahun El Niño, ladang yang mati kekeringan naik dari 0,01 menjadi 0,24 petak per tahun. Panen per orang setahun sesudahnya turun 17% dibanding tahun sesudah tahun netral.
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
6. **Perburuan hampir tidak terjadi.**
   - Total dalam 16 dunia × 900 tahun: 10 rusa, 16 babi, 211 ayam, 1 kerbau, 2 harimau. Daging memberi 0–1% energi.
   - Keberadaan satwa dengan dan tanpa manusia hampir sama, jadi kepunahan lokal karena perburuan berlebihan **belum** terlihat.
   - Penyebabnya, hewan lari begitu melihat manusia (dalam 1,5–3 tile, peluang 20% per pengamatan), sedangkan pukulan tangan kosong butuh sekitar 5 kali untuk merobohkan rusa yang larinya lebih cepat dari manusia.

## Perubahan yang membuat hasil ini (dibanding laporan Fase 2 sebelumnya)

| Perubahan | Sebelum | Sesudah |
| --- | --- | --- |
| Keluarga petani menyimpan benih dan membawanya dari lumbung; menanam di tile sebelah juga boleh | Ladang ≈ 1% energi; puncak dengan pertanian 406 < tanpa 444 | Ladang 50%; puncak 175 > 54 |
| Hanya air tawar yang bisa diminum, sungai dihitung sampai muaranya, haus 8×, ingatan sumber air, sumur butuh pertanian | Pulau menampung ≈ 1.500 pemburu-peramu; populasi menempel di batas teknis | Pemburu-peramu ≈ 30–150 orang, hidup di tepi sungai |
| Perempuan dewasa butuh 72% energi dan air laki-laki (mulai pubertas) | Rasio kelamin 152 | 73–104 (lihat di bawah) |
| Air tanah tepi sungai turun tajam saat kemarau panjang | Panen tidak turun di tahun El Niño | Petak layu di tahun El Niño; panen tahun berikutnya −17% |
| Perbandingan keahlian antar-tetangga lewat array, bukan map | 30% waktu CPU di `teacherOrStudentNear` | ≈ 1% |
| Bayi yang digendong tidak lagi diletakkan di atas air; jeda antar-peristiwa ikut disimpan | Dunia yang disimpan lalu dimuat ulang menyimpang dari aslinya | Identik (`TestSameSeedAndSaveRestoreReplayExactly`) |

## Belum terpenuhi / masalah yang diketahui

- **Perburuan terlalu jarang** (lihat butir 6). Perlu teknik berburu yang lebih realistis (mengendap, tombak sekali tusuk, berburu berkelompok) sebelum kepunahan karena perburuan bisa diuji.
- **Kematian dewasa hampir seluruhnya karena lapar** dengan laju sekitar 5% per tahun di semua umur. Karena itu modus usia kematian dewasa 18 tahun (acuan 68–78), dan e15 16–23 tahun (acuan 28–43). Di masyarakat nyata, populasi lebih banyak diatur lewat kesuburan dan penyakit → **Fase 3**.
- **Kematian bayi 0** dan peluang kembar 20% (nyata 1–2%) membuat TFR petani terlalu tinggi → Fase 3.
- **Haus masih sedikit lebih lambat dari lapar** (sekitar 4 lawan 3 tahun). Haus yang lebih cepat menggagalkan pendiri sebelum mereka mengenal jalan ke sungai.
- **Rasio kelamin pemburu-peramu 73** (petani 93): kini laki-laki yang lebih sering mati. Angka 72% dari Pontzer dkk. (2012) sebagian berasal dari laki-laki Hadza yang berjalan dua kali lebih jauh, padahal di simulasi kedua jenis kelamin bergerak sama banyak. Rasio massa tubuh (Hadza: 43 lawan 51 kg ≈ 0,84) kemungkinan lebih tepat.
- **Performa:** median 0,59 ms/tick (≈ 24% satu core pada 20×) dan tertinggi 2,29 ms/tick, masih di bawah anggaran 2,5 ms/tick. Tanpa iklim, dunia berpenduduk 1.450 butuh 12 ms/tick. Hambatan berikutnya: otak (`think`, 17%) dan pemrosesan tetangga dalam `sense` (≈ 30%).
- **Teknologi melambat:** Zaman Logam hanya tercapai di 2/16 dunia (Fase 1: 7/16), karena populasi kini jauh lebih kecil. Ini sejalan dengan teori bahwa populasi kecil sulit mempertahankan teknologi (Henrich 2004).

## Referensi tambahan

- Bocquet-Appel, J.-P. (2011). When the world's population took off: the springboard of the Neolithic Demographic Transition. *Science* 333.
- Borgerhoff Mulder, M. dkk. (2009). Intergenerational wealth transmission and the dynamics of inequality in small-scale societies. *Science* 326.
- Henrich, J. (2004). Demography and cultural evolution: how adaptive cultural processes can produce maladaptive losses — the Tasmanian case. *American Antiquity* 69.
- Naylor, R. L. dkk. (2007). Assessing risks of climate variability and climate change for Indonesian rice agriculture. *PNAS* 104.
- Pontzer, H. dkk. (2012). Hunter-gatherer energetics and human obesity. *PLoS ONE* 7(7): e40503.
