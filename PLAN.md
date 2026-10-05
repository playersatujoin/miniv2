# Rencana Pengembangan — Menuju Simulasi yang Lebih Nyata

Dokumen ini berisi rencana untuk membuat simulasi peradaban Miniv2 lebih realistis: apa yang akan ditambahkan, **kenapa** (dasar ilmiahnya), **bagaimana** (desain teknis di kode yang ada), dan **bagaimana kita tahu berhasil** (kriteria yang bisa diukur).

Status: **Fase 0 selesai** (5 Oktober 2026; lihat [hasilnya](#hasil-fase-0-baseline-v0)). Berikutnya **Fase 1**. Keputusan sudah dijawab (lihat [Keputusan](#keputusan-sudah-dijawab)). Urutan fase lain bisa diubah, tapi perhatikan [dependensi](#urutan-dependensi-dan-milestone).

---

## Daftar isi

1. [Prinsip](#prinsip)
2. [Kondisi saat ini (baseline)](#kondisi-saat-ini-baseline)
3. [Fase 0 — Alat ukur realisme & kalibrasi](#fase-0--alat-ukur-realisme--kalibrasi)
4. [Fase 0.5 — Utang teknis yang sudah diketahui](#fase-05--utang-teknis-yang-sudah-diketahui)
5. [Fase 1 — Belajar selama hidup & budaya](#fase-1--belajar-selama-hidup--budaya)
6. [Fase 2 — Ekologi: waktu, musim, tanaman, hewan, pertanian](#fase-2--ekologi-waktu-musim-tanaman-hewan-pertanian)
7. [Fase 3 — Tubuh & kesehatan](#fase-3--tubuh--kesehatan)
8. [Fase 4 — Masyarakat: hubungan, komunikasi, ekonomi, norma, kepemimpinan](#fase-4--masyarakat-hubungan-komunikasi-ekonomi-norma-kepemimpinan)
9. [Fase 5 — Dunia besar & banyak kelompok](#fase-5--dunia-besar--banyak-kelompok)
10. [Fase 6 — Bencana alam dari geologi](#fase-6--bencana-alam-dari-geologi)
11. [Fase 7 — Fisika material & pertambangan dalam](#fase-7--fisika-material--pertambangan-dalam)
12. [Lintas fase](#lintas-fase)
13. [Urutan, dependensi, dan milestone](#urutan-dependensi-dan-milestone)
14. [Risiko & mitigasi](#risiko--mitigasi)
15. [Batas realisme & di luar cakupan](#batas-realisme--di-luar-cakupan)
16. [Keputusan (sudah dijawab)](#keputusan-sudah-dijawab)
17. [Referensi](#referensi)

---

## Prinsip

Semua fase mengikuti prinsip yang sama:

1. **Muncul sendiri, bukan diskrip.** Setiap perilaku (belajar, berdagang, menghukum, berburu, mengungsi dari letusan) harus berasal dari keputusan otak makhluk. Kode hanya mengatur *hukum alam* (fisika, biologi, kimia) dan *detail pelaksanaan* (resep mana, target terdekat). Kehendak bebas tetap dijaga.
2. **Berdasar ilmu nyata.** Setiap mekanisme punya dasar dari biologi, demografi, geologi, ekonomi, atau sejarah. Kalau disederhanakan, penyederhanaannya ditulis terang-terangan.
3. **Realisme harus bisa diukur.** Setiap fase punya kriteria penerimaan berupa angka dari uji soak, dibandingkan dengan data dunia nyata bila ada.
4. **Pengamat tetap pengamat.** UI hanya boleh mengamati dan mengatur waktu, tidak memerintah makhluk.
5. **Dunia tetap hidup.** Setiap perubahan format simpanan punya strategi migrasi atau reset yang jelas, dan cadangan dibuat dulu.
6. **Anggaran performa.** Kecepatan 20× dengan populasi target harus tetap di bawah 1 core CPU, kecuali fase yang khusus menaikkan skala.

---

## Kondisi saat ini (baseline)

Ringkasan sistem yang sudah ada sebagai titik awal:

| Aspek | Sekarang |
| --- | --- |
| Awal dunia | Adam ♂ & Hawa ♀. Kalau semua punah, era baru dimulai dari genom terbaik |
| Otak | Rekuren 55 input → 24 neuron → 12 output, tetap sepanjang hidup; berubah hanya lewat evolusi (`backend/internal/sim/genome.go`) |
| Tindakan | belok, gerak, makan, minum, kawin, istirahat, kumpulkan, buat, bangun, beri, curi, serang |
| Pengetahuan | Global dan permanen untuk seluruh dunia (`sim` knowledge), tidak bisa hilang |
| Populasi | Batas keras = jumlah tile walkable / 60 (20–250); Starter Island sekitar 148 |
| Waktu | 20 tick/detik; umur 300–600 detik simulasi (akan diskalakan ulang di Fase 0: 1 tahun = 8 detik); belum ada siang/malam atau musim |
| Makanan | Tumbuh sendiri di tile rumput/hutan; ladang membuat area 3×3 subur |
| Kesehatan | Energi, hidrasi, kesehatan; belum ada penyakit, gizi, atau tahap hidup |
| Sosial | Reputasi berupa satu angka global; pasangan & rumah tangga; kejahatan tanpa hukuman sosial |
| Geologi | Pulau busur gunung api dengan model endapan nyata (`backend/internal/world`, `backend/internal/chem`); geologi statis, tanpa bencana |
| Material | Resep diskrit tanpa suhu, kadar bijih, atau keausan alat |
| Hasil soak (8 seed, 60 menit) | Zaman Logam di 8/8 seed; era 1 bertahan di 5/8; 2–6 unsur per jam; 0,34 ms/tick untuk ±145 makhluk |

Kesimpulan utama: **dunianya sudah kaya, tapi penghuninya belum mampu belajar dan bekerja sama** untuk memanfaatkannya. Karena itu kemajuan macet di Zaman Logam.

---

## Fase 0 — Alat ukur realisme & kalibrasi

**Tujuan.** Sebelum menambah fitur, kita butuh cara mengukur "seberapa nyata" simulasinya. Tanpa ini, setiap fase hanya bisa dinilai dari perasaan.

**Kenapa.** Simulasi ilmiah selalu divalidasi terhadap data. Demografi masyarakat pra-modern (harapan hidup, angka kelahiran, jarak kelahiran, penyebab kematian) sudah banyak diteliti, jadi bisa dijadikan patokan.

### Desain

1. **Skala waktu resmi: 1 tahun = 8 detik simulasi** (sudah diputuskan). Semua statistik umur dan laju dilaporkan dalam "tahun sim". Biologi yang ada perlu disesuaikan sedikit:

   | Besaran | Sekarang | Setelah disesuaikan (1 tahun = 8 s) |
   | --- | --- | --- |
   | Umur maksimum (gen) | 300–600 s (≈ 37–75 tahun) | sekitar 60–80 tahun = 480–640 s |
   | Dewasa / subur | 20% umur (60–120 s) | sekitar 15 tahun = 120 s |
   | Kehamilan | 10 s (≈ 1,25 tahun, terlalu lama) | sekitar 9 bulan ≈ 6 s |
   | Jarak kelahiran alami | tidak dimodelkan | sekitar 3–4 tahun ≈ 24–32 s (menyusui, Fase 3) |
   | Satu generasi | sekitar 2–3 menit sim | sekitar 25 tahun ≈ 200 s ≈ 3,3 menit sim (≈ 10 detik nyata pada 20×) |
   | Musim | belum ada | 2 musim (hujan/kemarau) masing-masing 4 s |
   | Siang/malam | belum ada | 1 hari ≈ 0,02 s, jadi hanya bisa sebagai **ritme abstrak** (kebutuhan tidur, cahaya) dengan periode tampilan yang terlihat |

   Konsekuensi: kecepatan evolusi kurang lebih tetap seperti sekarang, jadi perubahannya kecil. Pada kecepatan 20×, musim berganti 2–3 kali per detik nyata, jadi tampilan musim (warna rumput, hujan) harus dihaluskan atau hanya berupa indikator supaya layar tidak berkedip.

2. **Perekam statistik di sim.** Modul baru `backend/internal/sim/stats.go` mencatat setiap kelahiran dan kematian (umur, sebab, generasi, ibu, rumah), sehingga bisa menghitung:
   - harapan hidup saat lahir dan saat dewasa; persentase anak yang bertahan sampai "15 tahun";
   - angka kelahiran total (TFR, rata-rata anak per perempuan), umur ibu saat anak pertama, jarak antar-kelahiran;
   - rasio kelamin, piramida umur, ukuran rumah tangga;
   - distribusi penyebab kematian, angka pembunuhan per 100.000 orang per tahun sim;
   - ketimpangan kekayaan (koefisien Gini dari inventaris + simpanan rumah);
   - garis waktu teknologi (kapan setiap zaman dan unsur tercapai).
3. **Runner soak multi-seed.** Program baru `backend/cmd/soak` menjalankan N seed paralel selama M menit simulasi, lalu menulis `reports/<tanggal>/summary.json` dan `summary.md` (tabel per seed + median). Ini menggantikan pemakaian `go test -soak` manual.
4. **Data acuan.** `docs/reference-demography.md` berisi rentang dari literatur, lengkap dengan sumbernya, misalnya pemburu-peramu (Gurven & Kaplan 2007, perlu diverifikasi saat implementasi):
   - harapan hidup saat lahir sekitar 21–37 tahun;
   - sekitar 50–70% anak bertahan sampai umur 15;
   - umur kematian dewasa yang paling umum sekitar 68–78 tahun;
   - TFR sekitar 4–6, jarak antar-kelahiran sekitar 3–4 tahun.
5. **Uji A/B (ablasi).** Runner bisa menyalakan/mematikan satu fitur lewat flag (misalnya `-learning=off`), supaya dampak setiap fase terlihat jelas.
6. **UI pengamat.** Tambah bagian "Demografi" di tab Populasi: piramida umur, harapan hidup, TFR, dan penyebab kematian dari waktu ke waktu.

### Perubahan kode

`backend/internal/sim/stats.go` (baru), `backend/cmd/soak/` (baru), `backend/internal/sim/view.go` (SimInfo + demografi), `frontend/src/sim/protocol.ts`, `frontend/src/components/sim/` (komponen demografi), `docs/reference-demography.md` (baru).

### Kriteria penerimaan

- Runner menjalankan 8 seed × 2 jam sim (≈ 900 tahun sim, sekitar 36 generasi) dalam < 10 menit waktu nyata dan menghasilkan laporan.
- Biologi sudah disesuaikan ke 1 tahun = 8 detik, dan populasi tetap bertahan di uji soak (era 1 bertahan di ≥ 5/8 seed).
- Baseline saat ini tercatat sebagai "versi 0" untuk dibandingkan dengan fase-fase berikutnya.
- Setiap angka di laporan punya definisi tertulis.

Ukuran: **S–M**.

---

### Hasil Fase 0 (baseline-v0)

Selesai: skala waktu 1 tahun = 8 detik, `sim/stats.go` (tabel hidup periode), `GET /api/maps/{id}/sim/demography`, sakelar A/B `sim.Options` (`-off crime,instincts`), runner `backend/cmd/soak`, bagian Demografi di panel, dan semua waktu dalam tahun. Laporan: `reports/baseline-v0/summary.md`.

| Kriteria penerimaan | Hasil |
| --- | --- |
| Runner 8 seed × 2 jam sim < 10 menit | ✓ |
| Baseline tercatat | ✓ `reports/baseline-v0/` |
| Setiap angka punya definisi tertulis | ✓ `docs/reference-demography.md` |
| Era 1 bertahan di ≥ 5/8 seed | ✗ sekitar **50%**: 3/8 di baseline, 24/48 pada studi 48 seed (selang 95%: 36–64%). Biologi lama 30/48, selisihnya tidak signifikan. Penyebabnya keluarga pendiri menghabiskan makanan di sekitar tempat lahir lalu kelaparan karena tidak pindah. Ini perilaku (Fase 1–3), bukan parameter |

Median 8 dunia dibanding acuan pra-modern, dan fase mana yang diharapkan memperbaikinya:

| Indikator | Simulasi | Acuan | Penyebab selisih → fase |
| --- | ---: | ---: | --- |
| Harapan hidup saat lahir | 46,1 th ↑ | 21–37 | Belum ada penyakit, sehingga bayi dan anak hampir tak pernah mati → **Fase 3** |
| Bertahan sampai 15 | 0,92 ↑ | 0,44–0,73 | Sama → **Fase 3** |
| Sisa harapan hidup di 15 | 34,3 th ✓ | 28–43 | — |
| Modus usia wafat dewasa | 45 th ↓ | 68–78 | Kelaparan, kehausan, dan pembunuhan pada orang dewasa → **Fase 2 & 4** |
| TFR | 1,4 ↓ | 5–7 | Batas populasi keras (±148) menahan kehamilan → **Fase 2** (batas dihapus) |
| Jarak kelahiran | 5,9 th ↑ | 2,8–3,3 | Sama, ditambah masa tidak subur 2 tahun setelah melahirkan (pengganti menyusui) → **Fase 2–3** |
| Umur ibu saat anak pertama | 40,3 th ↑ | 18–20 | Sama → **Fase 2** |
| Gini | 0,54 ↑ | 0,21–0,29 | Barang menumpuk di simpanan rumah → **Fase 4** (berbagi, norma, pasar) |
| Pembunuhan | 266 / 100.000 | tanpa rentang | Dilaporkan saja → **Fase 4** (hukuman sosial) |

Penyesuaian di luar rencana: masa tidak subur setelah melahirkan dinaikkan dari 0,75 menjadi 2 tahun, sebagai pengganti sementara efek menyusui.

## Fase 0.5 — Utang teknis yang sudah diketahui

Masalah yang sudah terlihat dari uji sebelumnya. Kecil-kecil, tapi sebaiknya dibereskan sebelum fitur besar.

| # | Masalah | Rencana |
| --- | --- | --- |
| 1 | Lapisan geologi di peta statis: bijih yang sudah habis ditambang tetap tergambar | Endpoint `GET /api/maps/{id}/sim/geology` yang membaca sisa endapan dari sim; tile yang habis ditandai "bekas tambang" |
| 2 | `chem.NewGeology` sekitar 78 ms untuk peta 256×256 (membangun ulang model setiap kali) | Cache `GeoModel` per (seed, ukuran) di `world` |
| 3 | Ukuran simpanan dunia sekitar 3,7 MB (gzip) dan terus membesar | Bobot otak dalam `float32` biner + gzip; simpan bertahap |
| 4 | Era 1 hanya bertahan di 5/8 seed, dan beberapa dunia berulang kali punah (misalnya seed 6: 5 era dalam 1 jam) | Analisis penyebab (lokasi "Eden", rasio kelamin, kelaparan); kemungkinan besar membaik sendiri lewat Fase 1–3 |
| 5 | Kartu legenda geologi bisa menutupi HUD di layar sempit | Tata letak responsif |
| 6 | Server yang dijalankan dengan `-data` relatif menulis di direktori kerja yang salah | Peringatan saat start; contoh di README pakai path absolut |
| 7 | Angka kejahatan dan kebaikan sangat besar dan sulit dibaca | Tampilkan sebagai laju per tahun sim, bukan total |
| 8 | Log peristiwa tidak bisa disaring | **Filter kekerasan** (sudah diputuskan): pengamat bisa menyembunyikan peristiwa pembunuhan, serangan, dan wabah; visual tetap abstrak tanpa darah |

Ukuran: **S**.

---

## Fase 1 — Belajar selama hidup & budaya

**Tujuan.** Makhluk bisa **belajar dari pengalaman**, **meniru dan diajari** orang lain, dan **pengetahuan melekat pada orang**: ia bisa menyebar, tapi juga bisa hilang.

**Kenapa.**
- Peradaban manusia maju karena **akumulasi budaya**, yaitu pengetahuan yang diajarkan antar-generasi dan diperbaiki sedikit demi sedikit, bukan karena gen.
- Studi tentang Tasmania (Henrich 2004) menunjukkan populasi kecil yang terisolasi bisa **kehilangan** teknologi.
- Rantai panjang seperti "gali bijih → bawa ke tungku → lebur" hampir mustahil ditemukan oleh evolusi saja dalam waktu wajar. Inilah penyebab kemajuan macet sekarang.

### Desain

1. **Plastisitas otak (belajar dari pengalaman).**
   - Aturan belajar Hebbian yang dimodulasi imbalan, atau "three-factor rule" (Izhikevich 2007): `Δw = η · imbalan · jejak_eligibilitas`, dengan jejak = rata-rata (aktivitas pra × aktivitas pasca) yang meluruh.
   - **Imbalan berasal dari tubuh sendiri** (homeostasis): perubahan energi, hidrasi, dan kesehatan, ditambah rasa sakit saat diserang. Tidak ada tujuan buatan seperti "bangun tungku = +10", supaya kehendak bebas tetap terjaga.
   - Gen baru: laju belajar `η`, peluruhan jejak, dan seberapa besar bobot boleh berubah. Evolusi menentukan seberapa pandai suatu garis keturunan belajar (efek Baldwin).
   - **Bobot hasil belajar tidak diwariskan** (tidak Lamarckian); anak mulai dari bobot genetik. Yang diwariskan lewat ajaran adalah *keahlian* (lihat poin 2).
   - Demi performa, pembaruan dilakukan 1–2 kali per detik sim, bukan setiap tick.
2. **Keahlian dan pengetahuan per orang.**
   - `Creature.Skills map[tech]float32` (0–1) menggantikan pengetahuan global. Resep hanya bisa dikerjakan kalau pembuatnya punya keahlian di atas ambang; keahlian menentukan peluang berhasil, kecepatan, dan mutu hasil.
   - **Belajar sambil bekerja:** keahlian naik saat dipraktikkan, dan turun pelan kalau lama tidak dipakai.
   - **Penemuan** (eksperimen) memberi keahlian baru kepada penemunya saja.
   - Arsip dunia tetap mencatat "pernah ditemukan oleh X pada era Y" untuk UI, tapi juga menghitung **berapa orang hidup yang menguasainya**. Kalau jumlahnya jatuh ke 0, muncul peristiwa "Pengetahuan hilang: Peleburan besi".
3. **Pewarisan budaya.**
   - **Vertikal** (orang tua → anak), **horizontal** (teman sebaya), dan **miring** (guru → murid).
   - Output baru `ajar` dan input baru `guru dekat`, `murid dekat`, `keahlian tertinggi saya`. Kalau guru yang memutuskan `ajar` berada dekat dengan seseorang yang keahliannya lebih rendah, keahlian murid naik (lebih cepat jika sambil menonton guru bekerja).
   - **Peniruan:** makhluk yang melihat tetangga sukses (energi naik setelah suatu tindakan) mendapat dorongan kecil untuk meniru tindakan itu dalam konteks serupa.
4. **Tulisan (teknologi baru, Zaman Kimia ke atas).** Bangunan `perpustakaan` / `prasasti` menyimpan keahlian, jadi pengetahuan tidak ikut mati bersama pemiliknya. Ini meniru peran nyata penemuan tulisan.
5. **Perubahan otak.** Input sekitar 55 → 59 (`guru dekat`, `murid dekat`, `keahlian tertinggi`, `imbalan terakhir`), output 12 → 13 (`ajar`). Crossover per neuron tetap jalan.
6. **Ukuran otak ikut berevolusi, dengan biaya energi.** Sekarang jumlah neuron tersembunyi ditetapkan 24.
   - Usulan: jumlah neuron tersembunyi menjadi gen (misalnya 8–256). Mutasi bisa menambah atau membuang neuron dan koneksi (gaya NEAT: topologi tumbuh bertahap; neuron baru dimulai dengan bobot kecil supaya tidak merusak perilaku lama).
   - Setiap neuron menambah biaya metabolisme. Otak manusia memakai sekitar 20% energi tubuh, jadi otak besar harus "membayar dirinya sendiri".
   - Evolusi yang memutuskan apakah otak lebih besar sepadan dengan biayanya, bukan angka yang saya tetapkan.
   - Ini baru bermakna setelah ada pembelajaran selama hidup (poin 1). Tanpa itu, otak besar justru berkembang lebih lambat, karena mutasi acak harus "menebak" lebih banyak bobot.
   - **Batasan teknis:**
     - Koneksi rekuren tumbuh kuadratik (1.000 neuron ≈ 1 juta bobot per makhluk), jadi koneksi rekuren dibuat jarang (sparse), bukan penuh.
     - Batas atas neuron mengikuti anggaran performa.
     - BrainView di panel pengamat perlu mode ringkas (kelompok neuron atau heatmap) untuk otak besar, dan polling bobot penuh diganti dengan "bobot sekali + aktivasi berkala".
   - **Kriteria penerimaan:** distribusi ukuran otak tercatat per generasi di laporan soak. Pada dunia dengan pembelajaran menyala, ukuran rata-rata naik kalau manfaatnya lebih besar dari biaya energinya. Dengan pembelajaran mati (A/B), ukuran tetap kecil.

### Perubahan kode

- `sim/genome.go`: gen belajar, plastisitas.
- `sim/brain_learning.go` (baru): jejak, imbalan, pembaruan bobot.
- `sim/economy.go`: keahlian menggantikan pengetahuan global pada resep dan eksperimen.
- `sim/social.go`: ajar dan tiru.
- `sim/persist.go`: bobot fenotipe + keahlian; format simpanan v5.
- `chem`: struktur `perpustakaan`, tech `tulisan`.
- Frontend:
  - Inspector: keahlian berupa bar, guru/murid, dan perbedaan bobot hasil belajar disorot di BrainView.
  - Tab Peradaban: "dikuasai N orang", daftar pengetahuan yang hilang.

### Kriteria penerimaan (dengan runner Fase 0)

- Median keahlian naik seiring umur.
- Peristiwa "pengetahuan hilang" terjadi ketika pemegang terakhir mati tanpa murid.
- Uji A/B "belajar menyala vs mati": zaman ≥ 2 tercapai lebih cepat secara signifikan pada median 8 seed (target: Zaman Kimia di ≥ 4/8 seed dalam 30 generasi; target diukur per generasi supaya tidak bergantung pada skala waktu).
- Performa: ≤ 0,5 ms/tick untuk 150 makhluk.

**Risiko:** "reward hacking" (misalnya diam terus karena aman), ketidakstabilan bobot, dan beban CPU. **Mitigasi:** batasi besar perubahan bobot, beri peluruhan, uji A/B.

Ukuran: **L**.

---

## Fase 2 — Ekologi: waktu, musim, tanaman, hewan, pertanian

**Tujuan.** Makanan berasal dari ekosistem yang hidup: tanaman dengan siklusnya, hewan dengan populasinya, musim dan cuaca, serta pertanian dan peternakan yang sungguhan.

**Kenapa.**
- Revolusi pertanian adalah titik balik terbesar sejarah. **Surplus pangan** memungkinkan spesialisasi, dan spesialisasi memungkinkan teknologi.
- Di Indonesia, musim hujan dan kemarau serta kekeringan El Niño (berulang setiap sekitar 2–7 tahun) memengaruhi panen dan kelaparan secara nyata.
- Batas populasi yang sekarang "keras" (tile/60) tidak realistis. Di dunia nyata, populasi dibatasi oleh pangan, penyakit, dan ruang.

### Desain

1. **Waktu & musim.** Dengan 1 tahun = 8 detik: 2 musim (hujan/kemarau, sesuai iklim muson) masing-masing 4 detik. Siang/malam sebagai ritme abstrak (kebutuhan tidur, cahaya mengurangi jarak pandang) dengan periode tampilan yang terlihat, karena 1 hari nyata hanya sekitar 0,02 detik. Pada kecepatan tinggi, visual musim dihaluskan supaya tidak berkedip (lihat [Keputusan](#keputusan-sudah-dijawab)).
2. **Cuaca & iklim.**
   - Curah hujan per musim mengatur debit sungai, kelembapan tanah, dan pertumbuhan.
   - Tahun El Niño (kemarau panjang) dan La Niña (banjir) muncul secara stokastik dengan periode realistis.
3. **Tanaman.**
   - Jenis-jenis dengan siklus nyata: rumput, semak, pohon buah dengan musim berbuah, dan tanaman pangan lokal seperti padi, talas, ubi, sagu, pisang, kelapa.
   - Suksesi: lahan yang dibuka kembali menjadi semak, lalu hutan.
   - Kesuburan tanah per tile (disederhanakan menjadi N, P, K, dan bahan organik): habis karena panen, pulih lewat bera, kotoran ternak, atau abu vulkanik.
4. **Hewan (fauna).**
   - Agen sederhana dengan populasi realistis: rusa, babi hutan, ayam hutan (nenek moyang ayam peliharaan, didomestikasi di Asia Tenggara), kerbau, ikan di sungai dan laut, dan pemangsa (harimau, seperti di Sumatra dan Jawa dulu).
   - Dinamika pemangsa–mangsa (mirip Lotka–Volterra) yang stabil tanpa manusia; perburuan berlebihan bisa memunahkan hewan secara lokal.
   - **Domestikasi:** hewan yang berulang kali diberi makan dan dikandangkan lama-lama menjadi jinak dan bisa diternakkan.
5. **Pertanian & peternakan.**
   - Benih, tanam, tumbuh per musim, panen; sawah butuh **irigasi** (saluran dari sungai); hasil panen tergantung kesuburan dan hujan.
   - Bangunan baru: `lumbung` (penyimpanan, mengurangi pembusukan), `kandang`, `saluran_irigasi`.
   - **Pembusukan:** makanan basi seiring waktu; teknologi pengawetan (asap, garam) memperlambatnya.
6. **Hapus batas populasi keras** (sudah diputuskan). Daya dukung muncul sendiri dari pangan, air, penyakit (Fase 3), dan ruang. Batas teknis tetap ada demi performa, tapi jauh di atas daya dukung alami.
7. **Perubahan otak.**
   - Saluran sinar baru: `hewan`.
   - Input baru: `musim`, `cahaya`, `tanaman siap panen`, `lapar ternak`.
   - Output baru: `tanam`, `buru` (dipisah dari `serang`, supaya berburu hewan berbeda secara moral dari menyerang manusia).
8. **Dampak manusia.** Penebangan hutan menyebabkan erosi (kesuburan turun), banjir lebih sering di hilir, dan berkurangnya hewan.

### Perubahan kode

- `world`: iklim dan kesuburan awal per tile.
- `sim/terrain.go` → `sim/ecology/` (paket baru: tanaman, hewan, musim, cuaca).
- `sim/act.go`: tanam, buru, ternak.
- `chem`: item benih, hasil panen, daging, ternak; resep pengawetan; bangunan baru.
- Stream: hewan ikut dikirim di frame (event `fauna` terpisah atau tuple baru).
- Frontend:
  - Sprite hewan, tahap tumbuh tanaman, indikator musim dan cuaca (hujan), warna rumput mengikuti musim.
  - Panel "Ekologi": populasi hewan, tutupan hutan (%), stok pangan, dan grafik curah hujan.

### Kriteria penerimaan

- Tanpa manusia, populasi hewan berosilasi tapi bertahan ≥ 4 jam sim di 8/8 seed.
- Dengan manusia, perburuan berlebihan menyebabkan kepunahan lokal di sebagian seed (realistis).
- Pertanian menaikkan daya dukung: populasi puncak dengan pertanian > tanpa pertanian (A/B).
- Tahun kemarau panjang terlihat sebagai lonjakan kematian karena kelaparan.
- Tidak ada lagi batas populasi keras yang aktif dalam kondisi normal.

Ukuran: **L–XL**.

---

## Fase 3 — Tubuh & kesehatan

**Tujuan.** Tubuh makhluk mengikuti biologi manusia: tahap hidup, ketergantungan anak, gizi, penyakit menular, cedera, dan genetika berpasangan (diploid).

**Kenapa.**
- **Wabah** membentuk sejarah sama kuatnya dengan perang, dan menyebar lebih cepat di permukiman padat.
- **Bayi manusia bergantung bertahun-tahun.** Inilah alasan biologis adanya keluarga dan pembagian tugas.
- Dunia yang dimulai dari Adam & Hawa berarti **leher botol genetik**: perkawinan sedarah secara nyata meningkatkan risiko kelainan resesif (tekanan inbreeding). Menghindari kerabat sebagai pasangan (efek Westermarck) bisa berevolusi sendiri.

### Desain

1. **Tahap hidup.**
   - bayi (menyusu: harus dekat ibu, dan ibu mengeluarkan energi ekstra)
   - anak (harus diberi makan orang dewasa, tidak bisa bertahan sendiri)
   - remaja
   - dewasa
   - lansia (tenaga berkurang)
   - menopause (masa subur perempuan berakhir)
2. **Kehamilan & kelahiran.** Durasi realistis menurut skala waktu; risiko kematian ibu; jarak kelahiran alami karena menyusui.
3. **Gizi.** Energi (karbohidrat), protein, dan "zat gizi mikro" yang diabstraksikan. Makanan yang beragam itu penting. Kurang gizi menyebabkan pertumbuhan terhambat, kekebalan turun, dan kesuburan turun.
4. **Penyakit menular** (model SIR per individu).
   - Jenis-jenis:
     - **Diare/kolera:** sumber air tercemar dekat permukiman padat atau tanpa sanitasi.
     - **Malaria:** dataran rendah, air tergenang, musim hujan; kepadatan nyamuk sebagai medan per tile.
     - **Penyakit pernapasan:** menular lewat kedekatan.
     - **Parasit/cacingan.**
   - Kekebalan setelah sembuh; kematian bergantung umur (anak dan lansia paling rentan); gen kekebalan ikut berevolusi.
   - Teknologi yang menekan penyakit: sanitasi (jamban, sumur terlindung), obat herbal, lalu kina untuk malaria (sejarah nyata), dan seterusnya.
5. **Cedera.** Luka sembuh perlahan, risiko infeksi, dan cacat permanen (gerak lebih lambat).
6. **Genetika diploid.**
   - Dua alel per gen sifat, dengan dominan/resesif.
   - Beban alel resesif berbahaya: anak dari kerabat dekat lebih mungkin homozigot (kesehatan dan kesuburan turun).
   - Koefisien inbreeding dihitung dari silsilah.
   - Bobot otak tetap poligenik (crossover per neuron).
7. **Pemilihan pasangan.**
   - Input `pasangan potensial kerabat?`, `kesehatan calon`, `reputasi calon`.
   - Seleksi seksual muncul sendiri, termasuk kemungkinan menghindari perkawinan sedarah.
8. **Perubahan otak.** Input: `demam/sakit`, `lapar gizi`, `anak lapar dekat`, `menyusui`. Output `beri` dipakai untuk merawat dan memberi makan anak (sudah ada).

### Perubahan kode

- `sim/body.go` (baru): tahap hidup, gizi, cedera.
- `sim/disease.go` (baru): patogen, penularan, kekebalan.
- `sim/genetics.go` (baru): diploid, inbreeding.
- `chem`: makanan bergizi berbeda, obat, bangunan jamban.
- Frontend:
  - Inspector: tahap hidup, status penyakit, kekebalan, koefisien inbreeding.
  - **Pohon keluarga** (silsilah sampai Adam & Hawa).
  - Grafik wabah, dan overlay peta kepadatan nyamuk / sumber air tercemar.

### Kriteria penerimaan (dibandingkan data acuan Fase 0)

- Persentase anak yang bertahan sampai "15 tahun" berada dalam rentang acuan pra-modern (sekitar 50–70%).
- TFR sekitar 4–6 dan jarak antar-kelahiran sekitar 3–4 tahun sim, sebelum teknologi kesehatan.
- Wabah muncul lebih sering dan lebih parah di permukiman padat.
- Teknologi sanitasi menurunkan kematian karena diare secara terukur (A/B).
- Koefisien inbreeding tinggi pada generasi awal setelah Adam & Hawa, lalu menurun seiring populasi membesar.

Ukuran: **L**.

---

## Fase 4 — Masyarakat: hubungan, komunikasi, ekonomi, norma, kepemimpinan

**Tujuan.** Kehidupan sosial muncul dari interaksi pribadi: siapa teman dan siapa musuh, komunikasi, perdagangan, uang, norma dan hukuman, pemimpin, desa, serta hubungan antar-desa.

**Kenapa.**
- Di dunia nyata, kejahatan dibatasi oleh **konsekuensi sosial**. Eksperimen ekonomi menunjukkan manusia mau menghukum pelanggar walau merugikan diri sendiri (altruistic punishment; Fehr & Gächter 2002).
- **Perdagangan dan uang** memungkinkan spesialisasi.
- **Kepemimpinan dan lembaga** memungkinkan ratusan sampai ribuan orang bekerja sama.

### Desain

1. **Ingatan pribadi.**
   - Setiap makhluk menyimpan tabel hubungan terbatas (misalnya 16 orang terpenting) dengan skor kepercayaan, kasih sayang, dan dendam.
   - Skor berubah lewat interaksi: diberi makan (+), dicuri (−), diserang (−−), keluarga (+).
   - Input `reputasi orang dekat` diganti `hubunganku dengan orang terdekat`.
2. **Saksi & kabar (gosip).**
   - Kejahatan yang **terlihat** oleh orang lain (dalam jarak pandang) mengubah hubungan para saksi dengan pelaku.
   - Saat dua orang bertemu dan "berkomunikasi", mereka bertukar pendapat tentang orang ketiga, sehingga reputasi menyebar lewat jaringan sosial, bukan sebagai angka global.
3. **Komunikasi.**
   - Output baru `isyarat` (beberapa kanal 0–1) yang bisa ditangkap orang lain sebagai input (kekuatan dan arah).
   - Makna isyarat tidak ditentukan oleh kode; arti seperti "bahaya!" atau "ada makanan di sini" bisa muncul sendiri lewat evolusi dan pembelajaran.
4. **Perdagangan & uang.**
   - Output `dagang`: dua orang yang sama-sama mau bertukar barang sesuai kebutuhan masing-masing; nilai barang subjektif (kelangkaan bagi pemiliknya).
   - **Uang** muncul sebagai teknologi: koin dari emas, perak, atau tembaga yang sudah mereka gali, dicetak setelah mengenal peleburan.
   - Bangunan `pasar` menjadi tempat berdagang.
   - Harga tercatat sebagai deret waktu; ketimpangan diukur dengan koefisien Gini.
5. **Norma & hukuman.**
   - Output `hukum` (menegur, memukul, mengusir, atau mendenda pelanggar yang diketahui), dengan biaya bagi penghukum. Apakah hukuman muncul dan bertahan ditentukan oleh evolusi dan pembelajaran.
   - Pengusiran: orang yang diusir kehilangan akses ke rumah dan desanya.
6. **Kepemimpinan & desa.**
   - Kumpulan rumah yang berdekatan otomatis dikenali sebagai **desa** (punya nama, wilayah, dan penduduk).
   - **Pemimpin** adalah orang yang paling dipercaya banyak orang; input `arah pemimpin` dan `pemimpin sedang X` memungkinkan koordinasi (membangun bersama, bertahan).
   - Kepemilikan lahan di sekitar rumah dan ladang; aturan warisan per desa (patrilineal/matrilineal/bilateral) bisa muncul sebagai norma.
7. **Hubungan antar-desa.** Perdagangan, aliansi lewat perkawinan, dan konflik (penyerbuan) bila sumber daya langka.

### Perubahan kode

- `sim/relations.go` (baru), `sim/comm.go` (baru), `sim/trade.go` (baru), `sim/norms.go` (baru), `sim/village.go` (baru).
- `chem`: koin, `pasar`, tech `uang`.
- Frontend:
  - Inspector: jaringan sosial (teman/musuh, isyarat terakhir).
  - Panel "Desa": pemimpin, warga, norma, pasar.
  - Grafik harga dan Gini, serta log kejahatan dan hukuman.

### Kriteria penerimaan

- A/B "hukuman ada vs tidak": laju pencurian per tahun sim lebih rendah ketika hukuman berevolusi.
- Volume perdagangan naik seiring spesialisasi (keahlian Fase 1).
- Desa terbentuk sendiri di ≥ 6/8 seed; pemimpin berganti saat meninggal.
- Ada bukti isyarat yang bermakna: korelasi antara isyarat tertentu dan kehadiran makanan/bahaya di dekatnya.
- Angka pembunuhan per 100.000 per tahun sim dilaporkan dan dibandingkan dengan rentang masyarakat skala kecil (catatan: rentang nyatanya sangat lebar).

Ukuran: **XL** (sebaiknya dipecah menjadi 4a hubungan + gosip, 4b komunikasi, 4c dagang + uang, 4d norma + desa + pemimpin).

---

## Fase 5 — Dunia besar & banyak kelompok

**Tujuan.** Dunia berupa kepulauan dengan ribuan penduduk dan banyak desa, migrasi antar-pulau dengan perahu, dan budaya yang berbeda-beda.

**Kenapa.**
- Laju kemajuan teknologi dalam sejarah berhubungan erat dengan **jumlah orang yang terhubung** (Kremer 1993). Dengan sekitar 148 orang, masyarakat nyata pun tidak akan pernah sampai Zaman Listrik.
- Penyebaran Austronesia lewat laut adalah kisah nyata di wilayah ini, dan populasi yang terpisah akan membentuk budaya sendiri (efek pendiri).

### Desain

1. **Peta besar & kepulauan.**
   - `world.Generate` mendapat mode "kepulauan": beberapa pulau busur dengan selat.
   - `MaxSize` naik dari 256 ke 512 atau 1024.
   - Layer peta disimpan sebagai biner dengan RLE, bukan JSON angka.
2. **Perahu.** Teknologi dan bangunan `perahu` (kayu + tali + serat) untuk menyeberang; pelayaran punya risiko (cuaca, Fase 2).
   - **Satu Adam & Hawa saja** (sudah diputuskan): mereka mulai di satu pulau, dan pulau lain kosong sampai keturunannya membuat perahu dan menyeberang, seperti penyebaran Austronesia. Era baru setelah kepunahan juga dimulai dari satu pasangan.
3. **Skala populasi** (ribuan makhluk).
   - Otak "berpikir" 5 Hz dengan gerakan diinterpolasi, dan inferensi `float32` yang ramah SIMD.
   - Langkah paralel per wilayah (goroutine per partisi spasial) dengan penggabungan yang deterministik.
   - **Level of detail:** wilayah jauh dari pengamat disimulasikan lebih jarang. Pilihan lanjutan: model agregat tingkat desa, dengan konsekuensi determinisme yang dicatat.
4. **Streaming berbasis viewport lewat WebSocket** (dependensi sudah diizinkan). Klien mengirim area kamera, dan server hanya mengirim makhluk serta bangunan di area itu, dalam format biner ringkas. Kandidat pustaka: `github.com/coder/websocket` (kelanjutan `nhooyr.io/websocket` yang masih dirawat). SSE yang ada tetap dipertahankan sebagai cadangan untuk peta kecil.
5. **Simpanan per wilayah** dan simpanan bertahap.

### Perubahan kode

- `world` (kepulauan, format biner), `sim` (partisi, paralel, LOD), `api` (stream viewport), `store` (format biner).
- Frontend: chunk loading untuk peta besar, minimap kepulauan, stream viewport.

### Kriteria penerimaan

- Lebih dari 2.000 makhluk pada kecepatan 1× memakai ≤ 1 core; kecepatan 20× tetap lancar dengan LOD.
- Migrasi antar-pulau terjadi di ≥ 4/8 seed.
- Desa di pulau berbeda menunjukkan statistik budaya yang berbeda (norma, teknologi, isyarat).

Ukuran: **XL**. Bergantung pada Fase 1, 2, dan 4.

---

## Fase 6 — Bencana alam dari geologi

**Tujuan.** Gunung api, zona subduksi, dan sungai yang sudah ada menghasilkan **peristiwa nyata**: letusan, gempa, tsunami, banjir, longsor, kebakaran hutan.

**Kenapa.**
- Di Indonesia, letusan memusnahkan permukiman tapi meninggalkan **tanah vulkanik yang sangat subur**; ini salah satu alasan Pulau Jawa padat penduduk.
- Pilihan "tinggal dekat gunung yang subur tapi berbahaya" adalah keputusan nyata yang menarik untuk diamati.
- Gempa mengikuti statistik nyata (hukum Gutenberg–Richter: gempa kecil sering, gempa besar jarang).

### Desain

1. **Gunung api.**
   - Status bertahap: istirahat → waspada → letusan, dengan **tanda awal** yang bisa dirasakan makhluk (input `getaran tanah`, `bau belerang`). Mereka bisa belajar mengungsi.
   - Jenis bahaya:
     - **hujan abu:** merusak tanaman dan mengganggu pernapasan;
     - **awan panas:** mengalir di lembah dan mematikan;
     - **lahar:** mengalir di sungai, seperti di Merapi;
     - **aliran lava:** mengubah tile menjadi batuan baru.
   - Setelah letusan, abu menaikkan kesuburan tanah selama bertahun-tahun.
2. **Gempa.** Magnitudo mengikuti distribusi Gutenberg–Richter. Kerusakan bergantung jenis bangunan: gubuk ringan relatif aman, rumah bata tanpa tulangan rawan runtuh. Longsor di lereng curam.
3. **Tsunami.** Setelah gempa besar di lepas pantai, air surut dulu (tanda yang bisa dipelajari), lalu tile pantai rendah tergenang.
4. **Banjir.** Debit sungai di musim hujan (Fase 2); diperparah oleh penggundulan hutan di hulu.
5. **Kebakaran hutan.** Musim kemarau dan pembukaan lahan dengan api.
6. **Perubahan peta sementara/permanen.** Lapisan abu, lava yang menjadi batuan, alur sungai yang berpindah. Disimpan di sim sebagai "overlay medan", bukan mengubah peta editor.

### Perubahan kode

- `world/hazards.go` (status gunung api, katalog gempa).
- `sim/hazards.go` (dampak pada makhluk, bangunan, tanaman).
- `chem`: tanah abu, batuan lava.
- Frontend:
  - Animasi letusan, abu, lahar, dan banjir; spanduk peristiwa; sparkline seismograf.
  - Overlay **peta Kawasan Rawan Bencana (KRB)** seperti yang dipakai di Indonesia.

### Kriteria penerimaan

- Katalog gempa jangka panjang mengikuti distribusi Gutenberg–Richter (kemiringan nilai-b sekitar 1).
- Setelah letusan, kesuburan tanah di sekitarnya naik selama periode yang ditentukan.
- Ada bukti makhluk belajar merespons tanda awal (dengan Fase 1): proporsi yang menjauh dari gunung saat status naik > acak.
- Pola permukiman menunjukkan tarik-ulur antara kesuburan dan risiko (dicatat per seed).

Ukuran: **M–L**. Bisa dikerjakan paralel setelah Fase 2.

---

## Fase 7 — Fisika material & pertambangan dalam

**Tujuan.** Pengolahan material mengikuti fisika: suhu, bahan bakar, kadar bijih, keausan alat, dan tambang bawah tanah. Urutan zaman (perunggu sebelum besi) harus **muncul sendiri dari batasan fisika**, bukan dipaksa oleh tier.

**Kenapa.**
- Dalam sejarah, perunggu muncul sebelum besi karena **suhu**:
  - tembaga meleleh pada sekitar 1.085 °C dan timah pada sekitar 232 °C;
  - besi meleleh pada sekitar 1.538 °C; tungku bloomery kuno mereduksi besi pada sekitar 1.200 °C tanpa melelehkannya, sehingga butuh tungku dan teknik yang lebih maju.
- Produksi arang dalam jumlah besar untuk peleburan menyebabkan **penggundulan hutan** yang nyata dalam sejarah.

### Desain

1. **Model suhu tungku.** Suhu dari bahan bakar (kayu < arang < kokas) dan aliran udara (ububan/alat tiup). Jenis tungku: tungku tanah liat → tanur tiup → tanur listrik (Zaman Listrik). Setiap proses punya **suhu minimum**.
2. **Kadar bijih & neraca massa.**
   - Setiap endapan punya kadar khas model endapannya. Contoh: porfiri tembaga sekitar 0,5–1% Cu; emas epitermal dalam satuan gram per ton.
   - Hasil peleburan = massa × kadar × perolehan; sisanya terak.
   - Karena satuan di game diskalakan, yang dipertahankan adalah **rasionya**.
3. **Alat.** Daya tahan (aus dan rusak), kekerasan material (skala Mohs) menentukan kecepatan menambang, dan mutu (perunggu < besi < baja).
4. **Tambang bawah tanah (2,5D).**
   - Endapan punya kedalaman. Menambang membuat lubang/terowongan (bangunan `tambang` dengan tingkat kedalaman).
   - Kebutuhan tambang dalam: penyangga kayu, ventilasi, dan pompa di bawah muka air tanah (teknologi tinggi).
   - **Kecelakaan** (runtuh, banjir tambang) dengan peluang realistis.
   - Tambang terbuka untuk laterit dan batu bara.
5. **Energi & lingkungan.** Konsumsi bahan bakar dicatat; produksi arang menebang hutan (terhubung ke ekologi Fase 2).

### Perubahan kode

- `chem`: suhu proses, kadar per model, daya tahan alat, kekerasan.
- `sim/economy.go`: neraca massa, suhu tungku.
- `sim/mining.go` (baru): kedalaman, terowongan, kecelakaan.
- Frontend:
  - Inspector tungku (suhu, bahan bakar) dan lubang tambang di peta.
  - Overlay sisa endapan (menyambung utang teknis #1).

### Kriteria penerimaan

- Tanpa memaksa urutan tier, Zaman Perunggu mendahului Zaman Besi di ≥ 6/8 seed.
- Kebutuhan arang berkorelasi dengan berkurangnya tutupan hutan.
- Kecelakaan tambang terjadi dengan laju rendah dan tercatat sebagai penyebab kematian tersendiri.

Ukuran: **M–L**. Paling bermakna setelah Fase 1 (keahlian).

---

## Lintas fase

### Format simpanan & migrasi

- Setiap fase yang mengubah status menaikkan versi simpanan (v5, v6, …).
- Kebijakan default: dunia lama **dicadangkan** (`data/backups/<tanggal>/`) lalu diganti dunia baru dari Adam & Hawa, dengan pemberitahuan di log peristiwa. Migrasi otomatis hanya bila perubahannya kecil.

### Performa

- Setiap fase mencatat ms/tick di laporan soak.
- Anggaran sampai Fase 4: ≤ 1 core pada 20× untuk populasi target. Fase 5 punya anggaran sendiri.
- Profiling (`pprof`) wajib sebelum optimasi.

### UI pengamat

- Tab baru mengikuti fase: Demografi (0), Ekologi (2), Kesehatan / pohon keluarga (3), Desa / pasar (4), Bencana (6).
- Semua grafik mengikuti panduan warna yang sudah dipakai (diuji kontras di tema gelap).
- Tetap tanpa tombol yang memerintah makhluk.

### Pengujian

- Unit test untuk setiap aturan.
- Soak multi-seed + A/B untuk setiap fase.
- Tes determinisme: seed yang sama → hasil sama, termasuk setelah save/restore.

### Dokumentasi

README diperbarui per fase. Setiap dasar ilmiah dicatat di `docs/` beserta sumbernya.

### Cara kerja

Setiap fase dipecah ke beberapa agen paralel dengan kontrak tertulis dan pembagian file yang jelas (maksimal 5 agen bersamaan), seperti yang sudah dipakai sejauh ini. Pemimpin menggabungkan hasil dan menguji di browser.

---

## Urutan, dependensi, dan milestone

```
Fase 0 (alat ukur) ──┬──> Fase 0.5 (utang teknis)
                     │
                     ├──> Fase 1 (belajar & budaya) ──┬──> Fase 4 (masyarakat) ──┐
                     │                                 └──> Fase 7 (material)      │
                     │                                                            ├──> Fase 5 (dunia besar)
                     └──> Fase 2 (ekologi) ──┬──> Fase 3 (tubuh & kesehatan) ─────┘
                                             └──> Fase 6 (bencana)
```

| Milestone | Isi | Momen yang bisa diamati ("demo") |
| --- | --- | --- |
| M0 | Fase 0 + 0.5 | Laporan demografi pertama: harapan hidup, TFR, penyebab kematian dibanding data nyata |
| M1 | Fase 1 | "Pandai besi terakhir": pengetahuan peleburan hilang ketika pemegang terakhir mati tanpa murid |
| M2 | Fase 2 | Kemarau panjang El Niño memicu kelaparan; desa yang punya lumbung bertahan |
| M3 | Fase 3 | Wabah diare di desa padat; pohon keluarga sampai Adam & Hawa dengan koefisien inbreeding |
| M4 | Fase 4 | Pencuri diusir dari desa; muncul pemimpin dan pasar dengan koin emas |
| M5 | Fase 6 | Letusan dengan tanda awal: sebagian warga mengungsi, sebagian tidak; tanah subur sesudahnya |
| M6 | Fase 7 | Perunggu muncul sebelum besi tanpa dipaksa; hutan menipis karena arang |
| M7 | Fase 5 | Perahu pertama menyeberang ke pulau lain; budaya kedua pulau mulai berbeda |

Urutan yang disarankan: **M0 → M1 → M2 → M3 → M4 → (M5 dan M6 paralel) → M7**.

---

## Risiko & mitigasi

| Risiko | Dampak | Mitigasi |
| --- | --- | --- |
| Kompleksitas meledak, terlalu banyak sistem saling memengaruhi | Sulit di-debug, dan sulit dituning | Satu fase per waktu; uji A/B per fitur; flag untuk mematikan sistem |
| Dunia lebih sering punah karena tantangan baru (penyakit, musim) | Pengamat melihat era berulang terus | Pantau "lama era 1" di soak; tuning daya dukung; Adam & Hawa baru dari arsip tetap jadi jaring pengaman |
| Performa turun (belajar, hewan, penyakit) | Kecepatan 20× tersendat | Pembaruan berfrekuensi rendah, profiling, anggaran ms/tick per fase |
| Reward hacking pada pembelajaran | Perilaku aneh (misalnya diam terus) | Imbalan hanya dari tubuh, plastisitas dibatasi, uji ablasi |
| Klaim ilmiah keliru | Mengurangi kepercayaan pada "realisme" | Setiap angka acuan diberi sumber; label "perkiraan" bila belum diverifikasi |
| Simpanan membesar | Disk dan waktu simpan | `float32` biner, gzip, simpanan bertahap |
| Determinisme rusak oleh paralelisme (Fase 5) | Save/restore tidak identik | Penggabungan deterministik, tes determinisme wajib |
| Tema sensitif (kekerasan, wabah) | Tampilan yang tidak nyaman | Visual tetap abstrak dan tidak sadis; pengamat bisa menyaring log peristiwa |

---

## Batas realisme & di luar cakupan

- **Otak.** Makhluk sekarang punya 24 neuron tersembunyi. Dengan rencana di Fase 1 (ukuran otak berevolusi), jumlahnya bisa tumbuh sampai ratusan, sedangkan otak manusia sekitar 86 miliar neuron. "Kehendak bebas" di sini berarti keputusan sepenuhnya datang dari otak masing-masing, bukan kesadaran. Tujuan yang realistis adalah **pola besar** yang muncul sendiri, bukan meniru manusia individual persis.
- **Skala.** Waktu, ruang, dan jumlah bahan dikompres. Yang dijaga adalah **rasio dan urutan sebab-akibat**, bukan angka absolut.
- **Kimia.** 118 unsur dan mineral nyata ada, tapi reaksi kimia tetap berupa resep yang disederhanakan, bukan simulasi molekul.
- **Di luar cakupan untuk sekarang:**
  - agama dan kepercayaan (topik sensitif; tidak akan dimodelkan tanpa permintaan khusus);
  - penggambaran suku atau budaya nyata tertentu (nama tetap umum bergaya Indonesia, bukan meniru kelompok nyata);
  - grafik 3D.

---

## Keputusan (sudah dijawab)

Dijawab pada 5 Oktober 2026:

| # | Pertanyaan | Keputusan | Dampak ke rencana |
| --- | --- | --- | --- |
| 1 | Skala waktu | **1 tahun = 8 detik simulasi** (semula dipilih 30 detik, lalu diubah ke rekomendasi) | Fase 0 menyesuaikan biologi sedikit (umur sekitar 60–80 tahun = 480–640 s, dewasa sekitar 15 tahun = 120 s, kehamilan sekitar 6 s). Kecepatan evolusi kurang lebih tetap; musim 2 × 4 s dengan visual yang dihaluskan; siang/malam berupa ritme abstrak |
| 2 | Batas populasi keras | **Dihapus**, populasi dibatasi alam | Fase 2: daya dukung dari pangan, air, penyakit, dan ruang; hanya batas teknis demi performa |
| 3 | Tampilan kekerasan | **Abstrak + filter di log** | Fase 0.5 #8: filter peristiwa kekerasan; visual tanpa darah |
| 4 | Mulai dari | **Fase 0, lalu Fase 1** | Lihat urutan milestone |
| 5 | Dependensi WebSocket | **Diizinkan** | Fase 5: stream berbasis viewport lewat WebSocket (`github.com/coder/websocket`); SSE tetap sebagai cadangan |
| 6 | Adam & Hawa di kepulauan | **Satu pasangan, menyebar dengan perahu** | Fase 5: pulau lain kosong sampai diseberangi |
| 7 | Target perangkat | Mesin ini: AMD Ryzen 7 9800X3D (8 core / 16 thread), RAM 30 GB | Anggaran performa Fase 5 dihitung untuk mesin ini |

## Referensi

Sumber yang dipakai sebagai dasar. Angka spesifik perlu diverifikasi ulang saat implementasi.

- Gurven, M. & Kaplan, H. (2007). *Longevity among hunter-gatherers: a cross-cultural examination.* Population and Development Review. (Demografi pra-modern.)
- Henrich, J. (2004). *Demography and cultural evolution: how adaptive cultural processes can produce maladaptive losses — the Tasmanian case.* American Antiquity. (Hilangnya teknologi pada populasi kecil.)
- Kremer, M. (1993). *Population growth and technological change: one million B.C. to 1990.* Quarterly Journal of Economics. (Populasi dan laju teknologi.)
- Izhikevich, E. M. (2007). *Solving the distal reward problem through linkage of STDP and dopamine signaling.* Cerebral Cortex. (Belajar dengan imbalan tertunda.)
- Fehr, E. & Gächter, S. (2002). *Altruistic punishment in humans.* Nature. (Hukuman sosial.)
- Hukum Gutenberg–Richter (statistik frekuensi–magnitudo gempa).
- Azevedo, F. A. C. et al. (2009). Jumlah neuron otak manusia (sekitar 86 miliar).
- Data titik leleh standar: Cu sekitar 1.085 °C, Sn sekitar 232 °C, Pb sekitar 327 °C, Fe sekitar 1.538 °C.
- Peta Kawasan Rawan Bencana (KRB) gunung api Indonesia (PVMBG) sebagai contoh penyajian bahaya.
