# Rencana Pengembangan — Menuju Simulasi yang Lebih Nyata

Dokumen ini berisi rencana untuk membuat simulasi peradaban Miniv2 lebih realistis: apa yang akan ditambahkan, **kenapa** (dasar ilmiahnya), **bagaimana** (desain teknis di kode yang ada), dan **bagaimana kita tahu berhasil** (kriteria yang bisa diukur).

Status: **Fase 0, 0.5, dan 1 selesai** (5 Oktober 2026; lihat [hasil Fase 0](#hasil-fase-0-baseline-v0) dan [hasil Fase 1](#hasil-fase-1)). Berikutnya disarankan **Fase 2**, karena batas populasi keras adalah penghambat utama yang muncul di kedua uji. Keputusan sudah dijawab (lihat [Keputusan](#keputusan-sudah-dijawab)). Urutan fase lain bisa diubah, tapi perhatikan [dependensi](#urutan-dependensi-dan-milestone).

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
5. **Simulasi nyata: semuanya selalu aktif.** Setiap mekanisme alam (penyakit, nyamuk, air tercemar, siklus air, dan seterusnya) selalu berjalan dan tampil alami di dunia, misalnya air keruh, kawanan nyamuk, dan orang yang tampak sakit. Tidak ada tombol atau lapisan di UI untuk menyalakan atau mematikannya. Peta dan grafik analisis hanya ada di panel pengamat. Sakelar `-off` hanya untuk uji A/B di runner soak.
6. **Dunia tetap hidup.** Setiap perubahan format simpanan punya strategi migrasi atau reset yang jelas, dan cadangan dibuat dulu.
7. **Anggaran performa.** Kecepatan 20× dengan populasi target harus tetap di bawah 1 core CPU, kecuali fase yang khusus menaikkan skala.

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

Status Fase 0.5: #1 ✓ (`/sim/mined`, bekas tambang di peta) · #2 ✓ (cache GeoModel: 78 → 33 ms) · #3 ✓ (simpanan v5, bobot float32: 3,7 → 2,6 MB walau data otak 3×) · #4 sebagian (era 1 naik ke 11/16 dengan belajar; lanjut di Fase 2–3) · #5 ✓ (legenda responsif) · #6 ✓ (peringatan `-data` relatif) · #7 ✓ (laju per tahun) · #8 ✓ (filter kekerasan).

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

### Hasil Fase 1

Selesai: plastisitas otak (aturan tiga faktor, imbalan hanya dari tubuh), ukuran otak sebagai gen (8–64 neuron, dengan biaya energi), keahlian per orang (ambang 0,3), mengajar / belajar di rumah / mengamati, tulisan + `perpustakaan`, peristiwa "pengetahuan hilang", sakelar `-off learning|plasticity|culture`. Laporan A/B (16 dunia × 900 tahun per konfigurasi): `reports/fase1/comparison.md`.

| Kriteria penerimaan | Hasil |
| --- | --- |
| Median keahlian naik seiring umur | ✓ 0,37 (0–14 th) → 0,45 (30–44 th), landai |
| Peristiwa "pengetahuan hilang" terjadi | ✓ median 32 per dunia (contoh nyata: "Pengetahuan hilang: Api — pemegang terakhir, Fiyana, meninggal", lalu "ditemukan kembali") |
| Zaman Kimia di ≥ 4/8 dunia dalam 30 generasi | ✗ 0/16 di **semua** konfigurasi, termasuk model lama. Penghambat utamanya batas populasi keras (Fase 2) dan rantai bahan laboratorium |
| Ukuran otak dilaporkan per generasi | ✓ tetap sekitar 20 neuron; seleksi terlalu lemah untuk terlihat dalam 34 generasi |
| Performa ≤ 1 core pada 20× | ✓ 0,61 ms/tick (≈ 24% satu core) |

Temuan:
- **Belajar membantu bertahan hidup:** era 1 bertahan 11/16 (vs 9/16 model lama), dan Zaman Logam tercapai di generasi 7 bila pengetahuan dibagi bebas (vs 11,5).
- **Budaya per orang memperlambat teknologi:** hanya 7/16 dunia sampai Zaman Logam, vs 16/16 di model lama yang membagikan semua pengetahuan gratis. Ini realistis untuk populasi kecil (Henrich 2004).
- **Angka pembunuhan turun** menjadi 103 per 100.000 (vs 239 model lama); belum dianalisis penyebabnya.

Penyesuaian di luar rencana: plastisitas awal diturunkan (yang lebih tinggi merusak refleks bawaan), ditambah belajar dari keluarga di rumah (tanpa itu pengetahuan hampir tak menyebar), dan lupa diperlambat ke paruh waktu sekitar 170 tahun.

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

### Hasil Fase 2

Selesai:
- Paket `backend/internal/ecology`: iklim monsun dengan El Niño, La Niña, dan banjir; tumbuhan liar dan kesuburan tanah; enam tanaman pangan lokal; satwa (rusa, babi hutan, ayam hutan, kerbau liar, harimau); ikan; ternak.
- Di simulasi: menanam, memanen, menyimpan benih, berburu (terpisah dari menyerang), memancing, domestikasi; bangunan ladang, saluran irigasi, lumbung, kandang, dan sumur.
- Pangan membusuk, dan pengawetan (asap, garam) memperlambatnya.
- Hanya air tawar yang bisa diminum (`world.FreshWater`: sungai sampai muaranya dan danau), dan orang mengingat tempat ia biasa minum.
- Kebutuhan energi menurut jenis kelamin; bayi digendong dan disusui.
- Batas populasi keras dihapus.
- Simpanan v6. Frame stream membawa cuaca dan hewan, event `fields`, endpoint ekologi, dan panel Ekologi.
- Laporan soak: asal pangan, El Niño berpasangan, opsi `-render`.

Laporan A/B (16 dunia × 900 tahun per konfigurasi; tanpa manusia 8 × 1.800 tahun): `reports/fase2/comparison.md`.

| Kriteria penerimaan | Hasil |
| --- | --- |
| Tanpa manusia, hewan berosilasi tapi bertahan ≥ 4 jam di 8/8 seed | ◐ Rusa dan babi hutan ada 100% waktu di 8/8. Ayam hutan (95%) dan kerbau (76%) kadang punah lalu datang lagi. Harimau hanya 49%: pulau ini cuma menampung 2–5 ekor, sehingga harimau hidup sebagai populasi yang punah lalu datang lagi (`TestWildlifePersists -long` kini menguji keberadaan sepanjang waktu, bukan satu titik di akhir) |
| Perburuan berlebihan menyebabkan kepunahan lokal di sebagian seed | ✗ Sejak Fase 2b (jerat, mengendap, buruan terluka melambat, tombak) tertangkap 986 rusa dan 361 babi per 16 dunia × 900 tahun, tapi rusa dan babi tetap ada 100% waktu. Daging membusuk cepat dan satu rusa setara makanan sekitar 12 tahun, jadi perburuan tak pernah jadi sumber pangan utama (< 0,5%) |
| Pertanian menaikkan daya dukung (A/B) | ✓ Puncak median 100 lawan 56, populasi akhir 59 lawan 21; lebih tinggi di 12/16 dunia; ladang memberi 37% energi (60–90% di dunia yang pertaniannya berkembang penuh) |
| Kemarau panjang terlihat sebagai lonjakan kematian karena kelaparan | ✓ Dibanding tahun netral di sekitarnya, kelaparan setahun sesudah El Niño ×1,23 pada petani dan ×1,42 pada pemburu-peramu (La Niña ×1,09). Petak layu di tahun El Niño (1,07 lawan 0,02 per tahun), dan panen tahun berikutnya −18% |
| Tidak ada batas populasi keras yang aktif dalam kondisi normal | ✓ Batas teknis tidak tersentuh sekali pun. Tanpa iklim, populasi petani 2× lebih besar |
| Performa ≤ 1 core pada 20× | ◐ Median 0,33 ms/tick (≈ 13% satu core); dunia petani berpenduduk ≈ 1.000 butuh 6,5 ms/tick |

Temuan:
- **Pertanian gagal karena perilaku, bukan angka hasil panen.** Hasil panen masuk lumbung dan tidak pernah ditanam lagi, dan benih di tangan ikut dimakan. Aturan rumah tangga "simpan benih" menaikkan porsi ladang dari ≈ 1% menjadi 53%.
- **Air laut yang bisa diminum** membuat seluruh pulau layak huni, dengan daya dukung ≈ 1.500 pemburu-peramu. Menurunkan pangan liar bukan jalan keluar: laju tumbuh kembali ×0,8 saja sudah menurunkan kelangsungan pendiri dari 30/32 ke 24/32. Air tawar yang realistis menurunkan daya dukung tanpa melukai pendiri, dan membuat permukiman tumbuh di tepi sungai.
- **Rasio kelamin 152** disebabkan tubuh perempuan yang dihitung sebesar laki-laki, ditambah biaya hamil dan menyusui. Dengan 78% kebutuhan laki-laki sejak pubertas (di antara rasio energi 71% dan massa tubuh 84% suku Hadza, Pontzer dkk. 2012), rasionya menjadi 100–114.
- **Petani lebih subur dan lebih timpang** (TFR 8,9 lawan 5,5; Gini 0,50 lawan 0,31), seperti pada transisi demografi Neolitikum dan data kekayaan masyarakat kecil.
- **Determinisme:** seed yang sama menghasilkan dunia yang identik, juga setelah disimpan dan dimuat ulang di tengah jalan (`TestSameSeedAndSaveRestoreReplayExactly`). Tes ini menemukan dua bug yang sudah diperbaiki: bayi yang digendong kadang diletakkan di atas air, lalu dipindah saat dimuat ulang; dan jeda antar-peristiwa tidak ikut disimpan.

Penyesuaian di luar rencana:
- Laju lapar 3× dan laju haus 8×, dengan ingatan sumber air.
- Sumur dipindah ke teknologi pertanian.
- Air tanah tepi sungai mengikuti debit sungai setahun.
- Bayi digendong dan disusui (dimajukan dari Fase 3).
- Perbandingan keahlian antar-tetangga lewat array: `teacherOrStudentNear` turun dari 30% ke ≈ 1% waktu CPU.

Masalah yang diwariskan:
- **Fase 2b (sudah):** jerat, mengendap, buruan terluka melambat, tombak lebih mematikan, dan faktor perempuan 78%. Kriteria perburuan tetap belum terpenuhi karena energetika daging (lihat di atas). Yang dicoba lalu dibuang: memaksa pemburu berjalan pelan, refleks berburu yang lebih luas, dan keinginan membuat senjata. Itu menambah buruan tanpa menambah makanan, dan menurunkan kelangsungan dunia.
- **Ke Fase 3:**
  - Kematian dewasa hampir seluruhnya karena lapar, dengan laju ≈ 5% per tahun di semua umur, sehingga modus usia kematian dewasa 18 tahun dan e15 16–23 tahun.
  - Kematian bayi 0, dan peluang kembar 20% (nyata 1–2%).
- **Ke optimasi:** dunia petani besar masih di atas anggaran performa.

### Fase 2c — air yang nyata, ruang pribadi, otak yang tumbuh (Okt 2026)

Dikerjakan setelah Fase 2b atas permintaan pengamat:
- **Siklus air** (`ecology/hydrology.go`). Model reservoir linear per petak sungai dan danau, bergaya HBV: hujan, penguapan tanah, limpasan cepat, air tanah, dan aliran dasar. Air hilang lewat penguapan air terbuka (~5 mm/hari), rembesan dasar sungai (~4 mm/hari), minum (~4 L/orang/hari), dan irigasi sawah (~1,6 m/tahun); sisanya mengalir ke hilir. Satuan air 16 m³ (1 m di atas satu tile ~4×4 m).
  - Akibatnya, sungai surut di kemarau. Anak sungai kecil berhenti mengalir dan tinggal genangan, lalu kering. Danau menyusut dari tepinya. Di bawah dasar yang kering sering masih ada air yang bisa digali (*belik*). Sumur bertahan selama air tanah belum turun terlalu dalam.
  - Hujan pertama diserap tanah kering dulu, sehingga sungai baru mengalir lagi Oktober–November.
  - Saat El Niño, Oktober–Desember tinggal 16–22% sungai mengalir dan 23–31% dasar sungai kering. Kemarau sesudah El Niño ikut berat karena air tanah belum pulih.
  - Orang hanya melihat air yang masih ada, melupakan sumber yang kering, minum dari belik (lebih lambat) dan sumur, dan membangun sumur bila sungai di dekat rumahnya kering. Tepi sungai mengering bila air tanahnya turun, sawah irigasi hanya basah selama sungainya berair, dan ikan sungai mati bila sungainya kering.
  - Event stream `water`, panel "Air tawar" di tab Ekologi, dan sakelar soak `-off water`.
- **Ruang pribadi** (`sim/space.go`): tubuh tidak lagi saling menembus. Kerumunan menyebar membentuk lingkaran; bayi yang digendong tetap pada ibunya. Sakelar `-off space`.
- **Neurogenesis** (`sim/growth.go`): neuron baru tumbuh selama hidup ketika hidup terus memberi kejutan (lebih cepat pada anak-anak dan hanya saat cukup makan), lalu dipangkas bila tak berguna. Tiap neuron punya bias dan konstanta waktu sendiri; neuron lambat menjadi memori kerja. Gen `neurogenesis` diwariskan dan bisa berevolusi. Batas 128 neuron per otak. Tab "🧠 Neuron" di panel bawah. Sakelar `-off neurogenesis`. Simpanan v7 (dunia v6 tetap terbaca).
- **Tampilan 3D** untuk mode Amati (three.js): manusia berkerangka dengan pakaian per jenis kelamin dan animasi kerja, sungai mengalir dengan buih dan kaustik, rumput, burung, dan peralihan 2D⇄3D tanpa jeda.

Hasil A/B (16 dunia × 900 tahun):
- Ruang pribadi: era 1 bertahan 11/16 lawan 9/16 tanpanya.
- Neurogenesis: rata-rata otak 24,8 lawan 19,7 neuron. Puncak populasi lebih rendah (median 127 lawan 196). Eksperimen "neuron diam" (puncak 187) menunjukkan penyebabnya keputusan yang dipelajari neuron baru, bukan biaya energinya. Gennya bisa berevolusi turun bila tak menguntungkan.
- Siklus air: lihat `reports/fase2c/`.

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
  - Grafik wabah dan peta sebaran (nyamuk, air tercemar, telur cacing) di tab Kesehatan. Di peta dunia semuanya tampil alami (air keruh, kawanan nyamuk, orang sakit), bukan sebagai lapisan yang bisa dinyalakan.

### Kriteria penerimaan (dibandingkan data acuan Fase 0)

- Persentase anak yang bertahan sampai "15 tahun" berada dalam rentang acuan pra-modern (sekitar 50–70%).
- TFR sekitar 4–6 dan jarak antar-kelahiran sekitar 3–4 tahun sim, sebelum teknologi kesehatan.
- Wabah muncul lebih sering dan lebih parah di permukiman padat.
- Teknologi sanitasi menurunkan kematian karena diare secara terukur (A/B).
- Koefisien inbreeding tinggi pada generasi awal setelah Adam & Hawa, lalu menurun seiring populasi membesar.

Ukuran: **L**.

### Hasil Fase 3a — tahap hidup, kesuburan, kelahiran (Okt 2026)

- **Penuaan** (`sim/body.go`): risiko mati karena usia tua mengikuti Gompertz (bagian senesen model Siler untuk pemburu-peramu, Gurven & Kaplan 2007: a₃ ≈ 1,5·10⁻⁴, b₃ ≈ 0,086/tahun), digeser gen `lifespan`; tidak ada yang melewati 110 tahun. Lansia melemah (tenaga turun sampai 50%).
- **Kesuburan**: peluang hamil per bulan menurut umur (Wood 1994; puncak 25% di umur 20–30, separuhnya di umur 40), menopause sebagai gen (40–56 tahun), dan dikalikan gizi ibu. **Menyusui** menahan ovulasi penuh selama setahun dan sebagian sampai disapih (~2,4 tahun); bila bayinya meninggal, kesuburan kembali.
- **Kelahiran**: kematian ibu sekitar 1% per persalinan (lebih tinggi bila ibu sangat muda, tua, kurang gizi, atau mengandung kembar), kembar 1,5%, dan kematian neonatal 5% (lebih tinggi bila ibu kurang gizi atau terlalu muda/tua, dan pada bayi kembar).
- **Anak**: kemampuan makan sendiri naik bertahap dari umur 2 sampai 12 (Crittenden 2013). Anak di bawah 12 tahun tetap di dekat pengasuhnya, yaitu ibu, ayah, atau kerabat dewasa (Hewlett & Lamb 2005). Orang dewasa yang memilih **beri** dengan tangan kosong memetikkan makanan dan mengambilkan air untuk anak keluarganya yang lapar atau haus.
  - Diagnosis menemukan masalahnya: 107 dari 141 anak yang mati kelaparan atau kehausan sedang berada di samping orang tua yang kenyang, karena air tidak bisa diberikan dan orang tua hanya bisa memberi barang bawaan. Setelah perbaikan, kematian ini turun ke 80.
- Otak: input `anak lapar/haus dekat`, `menyusui`, `sakit`, `kurang gizi`, `calon pasangan kerabat`, dan `kesehatan calon pasangan` (simpanan v8, otak lama diperlebar otomatis).

### Hasil Fase 3b — penyakit menular (Okt 2026)

Desain (`sim/disease.go`, `ecology/pathogens.go`). Tidak ada yang tahu soal kuman; penyakit menyebar lewat jalur nyatanya:
- **Diare**: kotoran di dekat air terbawa hujan ke sungai, lalu kumannya hanyut ke hilir. Kuman mati dalam beberapa minggu dan pekat di genangan musim kemarau. Penularan juga terjadi lewat halaman yang kotor, terutama pada balita yang mulai disapih. Peluang sakit jenuh terhadap dosis (dosis–respons beta-Poisson). Belik menyaring sebagian kuman, dan sumur menyaring air tanah.
- **Malaria**: nyamuk Anopheles berkembang biak di genangan, tepi danau yang dangkal, sawah padi, dan kubangan musim hujan di dataran rendah. Bagian nyamuk yang membawa parasit mengikuti model Ross–Macdonald, dan gigitan per orang = nyamuk / (orang + hewan lain). Kekebalan klinis terbentuk setelah sering terinfeksi, jadi orang dewasa di daerah malaria kebanyakan hanya menjadi pembawa.
- **ISPA**: menular lewat kedekatan, dengan kejenuhan di keramaian. Galur baru muncul lebih sering di populasi besar (Black 1975), dan kekebalan lama hanya melindungi sebagian.
- **Cacingan**: telur di tanah sekitar tempat tinggal tanpa jamban. Jarang membunuh, tetapi ikut memakan makanan inangnya dan memperparah penyakit lain.
- **Keparahan** dipengaruhi umur (kerentanan terendah sekitar umur 10), kurang gizi, cacing, kekebalan, dan gen `immunity` (lebih kuat berarti lebih boros energi). Istirahat dan kerabat yang merawat meringankan. Demam membakar energi, diare menguras cairan, dan orang sakit bergerak lebih lambat dan jarang hamil.
- **Jamban** (`jamban`, butuh pertanian; kayu dan serat) menahan 90% kotoran orang di sekitarnya agar tidak mencemari air dan tanah.
- **Dunia lama** (simpanan sebelum v9) mendapat kekebalan menurut umur saat dimuat. Tanpa itu, dunia berpenduduk 748 orang turun ke sekitar 380 karena wabah "tanah perawan".
- **Tampilan**:
  - Di peta 2D dan 3D, air tercemar tampak keruh, kawanan nyamuk beterbangan di atas tempat berkembang biaknya (lebih ramai saat senja dan malam), orang sakit diberi ikon termometer (2D) atau berjalan membungkuk (3D), dan jamban tampil sebagai bangunan.
  - Tab **🩺 Kesehatan** berisi kasus per penyakit, kurva wabah, dan peta sebaran (nyamuk, air tercemar, telur cacing).
  - Inspector memuat bagian "Tubuh & kesehatan".
  - Stream mengirim event `mosquitoes` dan kekeruhan air ikut dalam event `water`. Endpoint baru `GET /sim/health`.
  - Sakelar soak: `-off disease|sanitation|attachment`. Simpanan v9.

Hasil (Starter Island; pendiri 32 dunia × 150 tahun; jangka panjang 16 dunia × 900 tahun; dunia padat = dunia uji 748 orang dan salinan dunia live 1.260 orang). Laporan: `reports/fase3/`.

| Kriteria | Hasil |
| --- | --- |
| l15 dalam rentang pra-modern | ✓ 0,68 (pendiri), 0,69 (900 tahun); acuan 0,44–0,73 |
| TFR 4–6, jarak lahir 3–4 tahun | TFR 6,3 (pendiri) ✓ dan 7,1 (900 tahun, sedikit di atas; Ache sekitar 8); jarak 3,2–3,5 ✓ |
| Wabah lebih sering dan parah di permukiman padat | ✓ Kelompok pendiri hampir tak terkena ISPA (0–1% kematian); di dunia padat ISPA memuncak 6–13% penduduk sakit sekaligus dan menjadi penyakit paling mematikan. Diare di dunia padat sekitar 6% sakit sekaligus |
| Sanitasi menurunkan kematian karena diare (A/B) | ✓ Dunia padat 60 tahun: kematian karena diare −20% dan insiden −17% hanya dengan 5 jamban untuk sekitar 800 orang. Jangka panjang (16 dunia × 900 tahun, 51 jamban dibangun sendiri): insiden diare 0,23 lawan 0,65 per orang per tahun dan kematian karena diare 2,4 lawan 4,9 per 1.000 orang-tahun (**−51%**) |
| Inbreeding | Fase 3c (belum) |

- Kematian bayi q0 0,10 (acuan Volk & Atkinson 2013: rata-rata 0,27, rentang 0,13–0,41), sedikit di bawah rentang. Balita 5q0 sekitar 0,13–0,16.
- Penyakit menyebabkan 14–24% kematian di dunia kecil, dan sekitar 50% di dunia padat (acuan Gurven & Kaplan: lebih dari 50%). Di dunia padat, kematian karena penyakit sebagian menggantikan kematian karena kelaparan (salinan dunia live: kelaparan 29 → 19 per tahun, penduduk sekitar 15% lebih rendah).
- Era 1 bertahan: 25/32 dengan penyakit lawan 26/32 tanpa penyakit (pendiri), 9/16 lawan 6/16 (900 tahun; selisih ini masih dalam derau 16 dunia).
- **Belum realistis**: orang dewasa masih terlalu sering mati kelaparan dan kehausan (modus umur wafat dewasa 18). Diagnosis menunjukkan 233 dari 245 orang dewasa itu mati di dekat keluarganya, di tempat yang makanan atau airnya sudah habis: keluarga menetap sampai sumber di sekitarnya habis, tidak pindah dan tidak berbagi. Ini bagian Fase 4 (berbagi pangan) dan gizi Fase 3d.

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
## Adaptasi subsistem engine — 6 Oktober 2026

Sasaran implementasi: enam kandidat pada pemeriksaan Gameporject1, disesuaikan
dengan backend Go, frontend Three.js, dan kehendak makhluk yang berasal dari otak.
Ini tidak menandai seluruh Fase 4–7 selesai.

| Pola referensi lokal | Implementasi Miniv2 | Batas saat ini |
| --- | --- | --- |
| `Peds/PedIntelligence/PedPerception`, `event/EventGroup` | `sim/perception.go`, `relations.go`: FOV, occlusion, suara, ingatan terbatas, penilaian pribadi | Belum ada gosip, komunikasi bermakna, norma atau lembaga |
| `pathserver/PathServer_PathSearch` | `sim/navigation.go`: A* lokal, tanpa pemotongan sudut, tujuan yang diketahui, pencarian ulang | Grid 2D; belum navmesh 3D, berenang, perahu, atau navigasi antar-pulau |
| `task/System/TaskComplex` | Pelaksana rute + `Job` yang ada; pembatalan berdasarkan output otak, status di Inspector | Pendekatan tujuan untuk makan/minum/kumpulkan; craft/build memakai pekerjaan lokal yang ada |
| `Peds/PedIntelligence/PedAILod` dan `streaming` | `cognition.go`: 5 Hz berpikir / 20 Hz tubuh; SSE viewport; mesh terrain per 32×32 tile | Kamera tidak mengubah simulasi; peta/grid, vegetasi dan layer ekologi belum di-stream sebagai aset jaringan per wilayah |
| `timecycle/TimeCycle` | `game3d/timecycle.ts`, integrasi cuaca dan jam pada `World3D` | Hari abstrak 2 detik; pencahayaan dirata-ratakan pada playback cepat |
| `ik/solvers`, `Peds/PedMoveBlend` | Karakter prosedural semi-realistis; kontak kaki dua ruas, arah pandang, blending aksi, LOD dan portrait | Belum motion capture, jari berartikulasi, ekspresi wajah, ragdoll atau simulasi kain |

Semua implementasi ditulis untuk struktur Miniv2. Kode C++ RAGE, SDK platform,
model dan tekstur dari arsip tidak menjadi dependensi aplikasi.

### Validasi dan pengukuran

- Go: unit/integration, replay seed yang sama dan save/restore, serta `go test -race ./...`.
- Kasus baru: saksi terhalang dinding, suara tanpa identitas, batas ingatan, rute
  menghindari dinding dan tidak menembus area tertutup, pembatalan niat, migrasi
  bobot neuron lama/tumbuh, viewport tidak mengubah dunia, backup sebelum migrasi.
- Frontend: typecheck/build; uji proporsi usia, solver kaki dengan target terjangkau
  dan tidak terjangkau, kurva cahaya; pemeriksaan WebGL melalui browser.
- Salinan dunia lama berhasil dimuat untuk pengamatan, dengan ratusan penduduk.
  Dunia yang dipakai profiling mulai dari 908 penduduk; aturan baru menghabiskan
  sekitar **4,47 ms/tick** untuk 20 detik simulasi di mesin ini (biaya restore tidak
  dimasukkan). Target ≤1 core pada 20× memerlukan ≤2,5 ms/tick, jadi target tersebut
  **belum tercapai pada populasi ini**. Hasil bukan jaminan performa perangkat lain.
- Perbaikan lanjutan (6 Okt, sore): simpanan v10 yang dipulihkan bisa langsung
  *panic* karena tiga dari empat makhluk bertindak dengan keputusan tersimpan sebelum
  mengindra lagi, sementara cache minat barang (`wantVec`) tidak ikut disimpan.
  Peluangnya tergantung isi simpanan: salinan simpanan live pukul 18.17 mematikan
  server lama seketika. Cache kini dibangun ulang dari `Want` yang tersimpan, dan
  `Sample` ikut disimpan (tambahan JSON, tetap v10). Ada tes regresinya.
- Optimasi yang **tidak mengubah hasil**: dunia live disalin, lalu dijalankan
  60 detik dengan kode lama dan baru, dan hash simpanannya identik. Isinya:
  mengajar/belajar di rumah hanya memeriksa skill yang benar-benar dimiliki;
  tetangga yang jelas di luar jangkauan atau di belakang ditolak sebelum `Hypot`;
  loop belajar jaringan saraf dibuat per baris. Dunia live (~930–1.000 orang),
  satu core, median 3 run: **4,54 → 3,94 ms/tick**. Target 2,5 ms/tick masih
  belum tercapai. Penghematan berikutnya memerlukan time-slicing pemindaian
  (pola `ai/ExpensiveProcess`, `ai/EntityScanner`) yang mengubah perilaku dan
  karena itu perlu soak A/B.

Soak delapan seed ×20 menit simulasi, map seed 1337, dijalankan satu dunia per
giliran. Kontrol mematikan `perception,navigation,ai_budget` pada versi kode yang
sama; ini membandingkan paket mekanisme, bukan mengisolasi setiap mekanisme.

| Ukuran | Paket aktif | Kontrol |
| --- | ---: | ---: |
| Era pertama bertahan sampai akhir | 7/8 | 8/8 |
| Median populasi akhir | 30,5 | 21,5 |
| Dunia mencapai Zaman Logam | 2/8 | 0/8 |
| Median harapan hidup saat lahir | 29,0 tahun | 32,6 tahun |
| Median pembunuhan /100.000/tahun | 52 | 12,5 |
| Median ms/tick sepanjang run | 0,168 | 0,164 |

Hasil mentah: `reports/engine-adaptation/on/` dan `off/`. Paket aktif tidak lebih
baik pada semua ukuran; seed 3 memulai era baru. Sampel pendek ini membuktikan
mekanisme berjalan, bukan realisme demografi atau keseimbangan jangka panjang.
Parameter kepercayaan, pendengaran, bentuk tubuh dan animasi adalah pendekatan
untuk simulasi/visual yang harus dikalibrasi terpisah jika dipakai untuk klaim ilmiah.

Simpanan v10 menambahkan ingatan, rute, dan cache keputusan. Kanal indra baru
ditambahkan di belakang layout lama. File asli v6–v9 diarsipkan sebelum penulisan
v10, dan kegagalan backup mencegah penimpaan simpanan lama.

## Adaptasi subsistem engine II — 6 Oktober 2026 (malam)

Semua sistem yang belum diadaptasi pada survei `clonegame` dikerjakan oleh lima
agen backend dan tiga agen frontend paralel, dengan kontrak tertulis (stub per
subsistem, kepemilikan file) dan penggabungan oleh pemimpin. Pola yang diambil,
bukan kode C++-nya:

| Pola referensi `clonegame/src/dev_ng/game` | Implementasi Miniv2 | Batas saat ini |
| --- | --- | --- |
| `event/EventShocking.h`, `ShockingEvents.h`, `ai/EntityScanner.h` | `sim/stimuli.go`: daftar terbatas (256) kejadian per jenis (mayat 4 s, perkelahian, pencurian, pemangsa, api, runtuh, tenggelam, jatuh), penggabungan, grid indra; `perceiveEvent` memakai grid | Mayat hanya terlihat 4 detik simulasi (setengah tahun) |
| `Peds/PedIntelligence/PedMotivation.h`, `animation/FacialData.h` | `sim/affect.go` (takut/marah/senang/duka, gen `Reactivity`/`Recovery`/`Cheer`); `game3d/expressions.ts` | Suasana hati hanya indra; tidak memengaruhi imbalan |
| `game/witness.h`, `WitnessInformation.h`, `task/Default/TaskChat.h` | `sim/interact.go`, `gossip.go`, `trade.go`: jabat tangan dua pihak, kabar "diceritakan", barter nilai subjektif | Tanpa uang, pasar, atau kebohongan |
| `vfx/misc/Fire.h`, `game/wind.h`, `game/weather.h` | `ecology/fire.go`, `ecology/climate.go` (muson, badai, petir), `sim/fire.go` (bangunan, orang, tungku) | Hewan tidak lari dari api |
| `Peds/NavCapabilities.h`, `physics/Floater.h`, `task/Movement/Climbing/`, `pathserver/PathServer_Hierarchical.cpp` | `sim/terrain.go` lapisan mobilitas, `locomotion.go`, `navigation.go` (A* berbiaya medium + wilayah 8×8), item `rakit` | Otak lama belum memakai laut/lereng, jadi berenang/memanjat masih jarang |
| `game/MapZones.h`, `Peds/Relationships.h`, `PedGroup/PedGroup.h` | `sim/village*.go`: klaster rumah, hull, nama, pemimpin berdasar kepercayaan dengan wibawa ±3 tahun | Belum ada norma, warisan atau penyerbuan |
| `control/replay/ReplayController.h`, `ReplayBufferMarker.h` | `sim/replay.go` (±5 menit frame terkompresi + penanda), `ReplayBar` | Inspector/desa saat replay masih keadaan live |
| `camera/cinematic/CinematicDirector.h` | `game3d/cinematic.ts` | Belum memeriksa penghalang pandangan |
| `naturalmotion/`, `cloth/`, `ik/solvers/QuadLegSolver.h` | `game3d/ragdoll.ts`, `cloth.ts`, `quadruped.ts` | Ragdoll tidak bertabrakan antar-tubuh |

### Validasi

- Go: `go test ./...` dan `go test -race ./...` lulus; replay seed sama dan
  save/restore identik; migrasi v10→v11 diuji (bobot lama tetap di tempatnya).
- Frontend: typecheck, 61 tes node, `vite build`; diperiksa di browser (2D, 3D,
  tab Desa, kamera otomatis menyorot mayat dengan ragdoll).
- Soak A/B 8 seed × 20 menit (`reports/engine-adaptation-2/on` dan `off`):

| Ukuran | Engine II aktif | Dimatikan |
| --- | ---: | ---: |
| Era pertama bertahan | 7/8 | 6/8 |
| Median populasi akhir | 44 | 22 |
| Median e0 | 32,4 tahun | 29,8 tahun |
| Median Gini | 0,41 | 0,30 |
| Median pembunuhan /100.000 | 77 | 0 |
| Desa terbentuk | 7/8 dunia | – |
| Barter per dunia | 28–168 | – |
| Kebakaran per dunia (150 tahun) | 3–14 | – |
| ms/tick dunia kecil | 0,20 | 0,23 |

  Sampel kecil; perbedaan populasi dan Gini belum boleh dianggap efek pasti.
- Dunia live (~900 orang): sekitar +13% ms/tick dibanding versi sebelumnya
  (6,13 vs 5,43 ms/tick di bawah beban yang sama). Target 20× pada satu core
  tetap belum tercapai untuk populasi ini.
- Pemimpin desa: sebelum wibawa diingat, masa jabatan rata-rata ±0,6 tahun;
  sesudahnya, di salinan dunia live, sekitar 8 tahun per pemimpin.

## Dorongan mencari makan, ingatan pangan, tabung air (6 Oktober 2026, malam)

**Diagnosis** (dunia live, ±900 orang, 60 detik simulasi; alat `/tmp/mv-diag`):
orang dewasa hanya bergerak 20% waktu; 89% berdesakan di tepi sungai (±76
orang dalam radius 3 petak); "berkeliling" yang terlihat adalah dorongan
tubuh di kerumunan (14% waktu) dan langkah zig-zag (arah belok berbalik pada
24% keputusan, kelurusan jalan 0,27). Refleks sederhana masih baik (89% mau
makan bila lapar dan ada makanan dalam jangkauan). Masalahnya: dari 183 yang
mati kelaparan, 152 sedang ingin makan dan makanan liar ada 3–8 petak dari
mereka, di luar pandangan; mereka tidak bergerak karena dorongan bawaan
pendiri (refleks "cari makan saat lapar") sudah luntur selama ±200 generasi.

**Perubahan** (simpanan v12):
- Gerak halus (`act.go`, pola `PedMoveBlend`): belok dan kecepatan mengikuti
  keinginan otak dengan konstanta waktu ±0,25 s / 0,4 s.
- Ingatan tempat makanan (`forage.go`): 5 tempat, memudar dalam ±2 tahun,
  dilupakan bila ternyata habis; indra `makanan yang diingat` + arahnya;
  eksekutor rute menuju tempat yang diingat; tempat makanan diceritakan saat
  mengobrol.
- Tabung air bambu (`tabung_air`, teknologi `wadah_air`): diisi saat minum,
  diminum saat jauh dari air, membawa kuman sumbernya.
- Dorongan bawaan (`drive.go`): empat neuron di gen setiap orang (lapar →
  jalan dan makan, belok ke makanan yang diingat; haus → ke air lebih dulu),
  netral saat kenyang. Dimasukkan sekali saat migrasi; selanjutnya biasa
  diwariskan, dimutasi, atau dihapus oleh evolusi.
- Otak tanpa batas ukuran: hanya biaya energi per neuron yang membatasi.

**Hasil:** dunia live yang sama, 60 detik: mati kelaparan 183 → 132; bergerak
20% → 26,5%; mau makan saat lapar 89% → 94%; namun mati kehausan 20 → 53
(orang mulai meninggalkan sungai). Dunia baru 8 seed × 150 tahun
(`reports/forage-drive/on` vs `reports/engine-adaptation-2/on`): era pertama
bertahan 8/8 (sebelumnya 7/8), median populasi 76 (44), e0 33,5 (32,4);
kelaparan + kehausan tetap ±58% kematian. Dunia masih dibatasi pangan: populasi
tumbuh sampai pangan habis. Sasaran berikutnya: berbagi pangan, cadangan, gizi
(3d), genetika (3c).

## Gizi (3d), genetika (3c), berbagi pangan (Fase 4) — simpanan v13 (7 Oktober 2026)

Dikerjakan tiga agen paralel, digabung pemimpin. Semua aturan dunia; otak tetap
memutuskan.

- **Fase 3d — gizi** (`chem/nutrition.go`, `ecology/nutrition.go`,
  `sim/nutrition.go`): tiap makanan punya kepadatan protein (USDA × skor mutu
  protein ÷ kebutuhan WHO/FAO/UNU 2007) dan mikronutrien (vit. A, besi, seng,
  yodium, vit. C, folat; FAO/WHO 2004). Tubuh menyimpan status protein dan
  mikronutrien yang naik-turun dalam hitungan bulan; kebutuhan anak, hamil dan
  menyusui berbeda. Kurang gizi: anak kerdil (tinggi tertinggal, mengejar
  sebagian), infeksi lebih berat, kesuburan turun, risiko melahirkan dan
  neonatal naik, luka lambat sembuh. Indra `kurang gizi` kini terisi.
  Soak (`reports/fase3d`): q0 0,08 → 0,12 (masuk acuan), l15 0,80 → 0,74.
  Di dunia yang bergantung pada padi, sebagian besar orang kekurangan
  mikronutrien — pola monokultur beras.
- **Fase 3c — genetika** (`sim/genetics.go`, `pedigree.go`, `geneview.go`):
  31 lokus diploid (13 kelainan resesif mematikan bayi, 6 otot/tulang lemah,
  6 kekebalan lemah, 5 kesuburan rendah, 1 talasemia yang juga melindungi dari
  malaria); muatan pendiri ±0,3 lethal equivalent per gamet (Gao dkk. 2015).
  Koefisien inbreeding F dari silsilah (7 generasi). Indra `calon pasangan
  kerabat` dan `kesehatan calon pasangan` kini terisi. Pohon keluarga di
  Inspector (`/sim/creatures/{id}/family`). Soak (`reports/fase3c`): F 0,25 di
  generasi 2, puncak ±0,44 (generasi 9–12), turun ke ±0,17 setelah generasi
  36; anak dengan F ≥ 1/4 mati sebelum umur 1 tahun 14,5% (vs 8,2% tanpa
  genetika); seleksi membersihkan alel letal (0,022 → 0,007 dalam 900 tahun).
  Penghindaran kerabat belum berevolusi dalam 900 tahun.
- **Fase 4 — berbagi pangan** (`sim/sharing.go`): bangkai buruan tertinggal di
  tempat dan bisa dimakan siapa saja; *demand sharing* (orang lapar makan satu
  satuan dari bekal kerabat yang lebih kenyang atau orang yang membawa banyak;
  Peterson 1993, Blurton Jones 1984); *bawon* (siapa pun boleh membantu panen,
  membawa pulang 1/6; Collier dkk. 1974); lumbung desa; orang lapar tanpa
  ingatan tempat makan mengenal blok lahan terdekat yang paling banyak
  makanannya. Dunia lama 60 detik: mati kelaparan 157 → 55, simpanan lumbung
  0 → 304.

**Gabungan v13** (8 seed × 150 tahun, `reports/v13/on`): kelaparan 37% → 10%
dari kematian, kehausan 22%, usia tua 25%; e0 38,0; e15 33,7; modus umur wafat
dewasa 26 (sebelumnya 18–20); TFR 5,8. Dunia lama (±900 orang, 60 detik):
kelaparan 183 → 52, populasi akhir 892 → 1.080, ISPA menjadi pembunuh utama di
permukiman padat.

Masih belum: isyarat/komunikasi bermakna, uang dan pasar, norma dan hukuman,
kepemilikan tanah dan warisan, aliansi/penyerbuan (sisa Fase 4); Fase 5 (dunia
besar), Fase 6 (bencana), Fase 7 (fisika material).
