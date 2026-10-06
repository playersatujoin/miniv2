# Acuan Ekologi (Fase 2)

Dokumen ini mencatat dasar setiap aturan ekologi di simulasi: musim dan El Niño, tumbuhan, tanaman pangan, hewan, ikan, pembusukan pangan, dan tubuh manusia. Ada juga **apa yang dikompres** dan **bagaimana setiap angka di laporan soak didefinisikan**. Angka yang hanya perkiraan diberi label.

Skala waktu tetap **1 tahun = 8 detik simulasi**. Ruang dan jumlah juga dikompres: pulau 128×128 tile jauh lebih kecil dari pulau sungguhan. Yang dijaga adalah **urutan sebab-akibat dan perbandingan** (hutan lawan ladang, kemarau biasa lawan El Niño, hewan kecil lawan besar), bukan angka absolut.

## 1. Musim dan iklim

| Aturan | Nilai di simulasi | Dasar |
| --- | --- | --- |
| Pola hujan tahunan | Kurva kosinus, puncak pertengahan Januari, terendah pertengahan Juli; rata-rata tahunan = 1 | Pola monsun wilayah selatan Indonesia (Jawa, Bali, Nusa Tenggara): hujan puncak Desember–Februari, kering Juni–Agustus (Aldrian & Susanto 2003) |
| Musim | "Hujan" kira-kira Oktober–April, "kemarau" April–Oktober (masing-masing 4 detik) | Sama |
| Kelembapan tanah | Mengikuti hujan dengan jeda sekitar sebulan; sungai dan air tanah mengikuti rata-rata sekitar setahun | Penyederhanaan model ember (*bucket model*) |
| Air tanah tepi sungai | Dalam 2 tile dari air tawar, kelembapan minimal 1,6 × debit setahun − 0,35: sekitar 0,63 di tahun biasa, di bawah 0,5 selama El Niño, sekitar 0,2 di akhir kemarau terparah. Padi tadah hujan di tepi sungai bisa layu; ubi bertahan; sawah beririgasi paling aman | Saat El Niño, debit sungai dan muka air tanah di Jawa turun tajam, dan padi tadah hujan paling terpukul (Naylor dkk. 2007). Rumus lama (0,35 + 0,5 × debit) membuat tepi sungai hampir tidak pernah kering, sehingga panen di tahun El Niño sama sekali tidak turun |
| ENSO | Ditentukan setiap Mei untuk 12 bulan. Peluang El Niño 25% sesudah tahun netral, 5% sesudah El Niño atau La Niña; La Niña 15% sesudah netral, 35% sesudah El Niño, 45% sesudah La Niña | Kejadian ENSO berulang setiap sekitar 2–7 tahun. El Niño sering disusul La Niña, dan La Niña sering berlangsung lebih dari setahun (McBride, Haylock & Nicholls 2003). Hasil uji: satu El Niño setiap sekitar 4–5 tahun |
| El Niño | Hujan ×0,3 pada Mei–Oktober, ×0,4 pada November–Februari (hujan terlambat datang), ×0,8 sisanya | Di Indonesia, El Niño memperkering musim kemarau dan menunda awal musim hujan (Hendon 2003; Naylor dkk. 2007) |
| La Niña | Hujan ×1,35 sepanjang tahun | Tahun La Niña lebih basah |
| Banjir | Diperiksa setiap Februari. Peluang 60% di tahun La Niña, 8% di tahun netral, 0 saat El Niño. Keparahan ×(1 + bagian hutan yang ditebang). Merusak tanaman di tepi sungai, mengurangi pangan liar di sana, dan menambah kesuburan (lumpur) | Penebangan hutan memperbesar limpasan dan banjir di hilir. Endapan banjir menyuburkan dataran aluvial |
| Siang/malam | Ritme abstrak 2 detik; malam mengurangi jarak pandang sampai sekitar setengahnya | Satu hari nyata hanya sekitar 0,02 detik, terlalu cepat untuk dimodelkan langsung |

## 2. Tumbuhan liar dan tanah

Setiap tile yang bisa dilalui punya dua "kolam" makanan:

- **Pangan liar** (buah, umbi, pucuk, sayur, kerang di pantai): bisa dimakan manusia, babi hutan, dan dipetik untuk dibawa.
- **Rumput dan daun**: untuk hewan perumput (rusa, kerbau, ayam hutan).

Manusia **tidak bisa mencerna rumput**. Sebelum Fase 2, rumput di seluruh pulau dianggap makanan manusia. Itulah yang membuat daya dukung alami tak terbatas (uji tanpa batas populasi: 715 jiwa dan masih tumbuh).

| Jenis tutupan | Pangan liar (kapasitas, laju) | Rumput (kapasitas, laju) | Kepekaan kemarau | Kesuburan alami |
| --- | --- | --- | --- | --- |
| Hutan | 1,0, 0,024/s | 0,15, 0,003/s | 0,7 | 0,75 |
| Semak | 1,5, 0,036/s | 0,25, 0,005/s | 0,8 | 0,65 |
| Padang bunga | 0,6, 0,014/s | 0,35, 0,007/s | 0,9 | 0,6 |
| Rumput | 0,15, 0,004/s | 0,4, 0,008/s | 1,0 | 0,6 |
| Pasir pantai | 0,12, 0,003/s (kerang, kepiting) | 0,05 | 0,3 | 0,25 |
| Tanah terbuka/batu | 0,02 | 0,08 | 0,9 | 0,4 |

- **Musim:** laju tumbuh × (1 − kepekaan + kepekaan × kelembapan). Kapasitas ikut menyusut di musim kemarau, jadi tanaman layu atau mengering.
- **Hutan yang ditebang:** pangan liar di lantai hutan ikut berkurang (pohon buah dan naungan hilang), dan kesuburan tanah turun sampai 30% (erosi).
- **Kesuburan:** pulih ke tingkat alaminya dengan waktu karakteristik sekitar 10 tahun (bera). Pemulihan 4× lebih lambat jika ada tanaman di atasnya. Ditambah lumpur banjir dan kotoran ternak.
  - Batuan vulkanik bernilai 0,9 (andosol, salah satu tanah tersubur).
  - Tepi sungai mendapat +0,15 (aluvium).

Angka kapasitas dan laju adalah **kalibrasi**, bukan pengukuran. Patokannya: ladang harus 10–100× lebih produktif per tile daripada hutan liar, sejalan dengan kepadatan petani dibanding pemburu-peramu.

## 3. Tanaman pangan dan pertanian

| Tanaman | Umur panen | Berbuah ulang | Kebutuhan air | Tumbuh liar di | Dasar |
| --- | --- | --- | --- | --- | --- |
| Padi | sekitar 4 bulan (0,35 th) | — (semusim) | tinggi (0,75) | rumput di lahan basah | Padi 110–150 hari; dibawa penutur Austronesia ke Asia Tenggara Kepulauan |
| Talas | sekitar 8 bulan | — | 0,65 | lantai hutan dekat air | Talas dipanen 7–12 bulan; didomestikasi di Nugini (Denham dkk. 2003) |
| Ubi (uwi, gembili; *Dioscorea*) | sekitar 9 bulan | — | rendah (0,3), tahan kering | lantai hutan | *Dioscorea* asli Asia Tenggara. Ubi jalar baru datang dari Amerika abad ke-16, jadi sengaja **tidak** dipakai |
| Pisang | 1 tahun | setiap sekitar 0,6 tahun selama 8 tahun | 0,5 | hutan, rumput | *Musa* liar asli wilayah ini |
| Kelapa | 6 tahun | setiap sekitar 0,5 tahun selama 60 tahun | 0,3 | pasir pantai | Kelapa mulai berbuah 6–10 tahun |
| Sagu | 8 tahun | setiap 4 tahun (tunas baru) | 0,8, harus di tepi air tawar atau lahan irigasi | rawa dan tepi sungai | Sagu dipanen pada umur 8–15 tahun (Flach 1997) |

- **Asal benih:** memetik pangan liar di tile yang ada tanaman liarnya memberi peluang 35% mendapatkan bahan tanam (gabah, umbi, anakan, butir kelapa). Barang yang sama bisa dimakan atau ditanam, jadi **makan habis berarti tak ada yang ditanam**.
- **Menanam:** keputusan otak (`tanam`), hanya untuk orang dewasa, di tanah yang cocok: tile tempat ia berdiri, atau tile kosong di sebelahnya (sejangkauan tangan). Orang yang punya rumah menanam dalam radius 8 tile dari rumahnya (kebun dekat permukiman).
- **Menyimpan benih:** orang dewasa yang bisa bertani (keahlian ≥ 0,3) tidak memakan 2 unit benih yang dibawanya, dan keluarga petani menyisakan 4 unit tiap jenis tanaman di rumah atau lumbung. Benih baru dimakan bila energi di bawah 0,25 (paceklik). Saat di rumah tanpa benih di tangan, petani mengambil 2 unit dari simpanan untuk ditanam di kebun. Ini aturan rumah tangga yang otomatis (seperti menyimpan hasil di rumah), bukan keputusan otak. Petani di mana pun menyisihkan benih, dan "memakan benih" adalah tanda kelaparan parah. Tanpa aturan ini, benih yang masuk lumbung tidak pernah ditanam lagi, dan pertanian hanya menyumbang sekitar 1% makanan.
- **Pertanian ditemukan** saat panen pertama dari tanaman yang ditanam. Keahlian bertani menaikkan hasil (×0,6 sampai ×1,4).
- **Hasil panen:** dipengaruhi kesehatan tanaman, kesuburan, dan keahlian. Bonus pengelolaan:
  - ladang (dibersihkan dan disiangi): tumbuh ×1,25, hasil ×1,3, menguras kesuburan ×0,8;
  - padi sawah beririgasi: hasil ×1,6, menguras kesuburan hanya ×0,25 (sawah memang tidak cepat mengurus tanah);
  - pupuk kandang: hasil ×1,15.
- **Kekeringan:** tanaman layu jika kelembapan < ½ kebutuhannya. Tanaman semusim mati dalam beberapa bulan; pisang dan palem 3× lebih tahan.
- **Panen matang** menunggu sekitar setengah tahun (umbi bisa "disimpan di tanah"), lalu mulai rusak dimakan burung atau busuk.
- **Pemilik:** ladang milik keluarga penanam. Orang lain hanya bisa mengambil panennya dengan mencuri (dicatat sebagai kejahatan). Ladang tanpa keluarga yang tersisa boleh dipanen siapa saja.
- **Hewan perusak:** rusa, babi, dan kerbau liar memakan dan menginjak ladang.

Hasil per panen (padi 5, talas 5, ubi 5, pisang 2, kelapa 2, sagu 15 unit) dikalibrasi terhadap kebutuhan satu orang, yaitu sekitar 1–1,5 unit per tahun (lihat §6).

## 4. Hewan

| Spesies | Makan | Dewasa / umur (th) | Anak per kelahiran / jarak (th) | Lari dari manusia | Catatan |
| --- | --- | --- | --- | --- | --- |
| Rusa | rumput | 1,5 / 14 | 1 / 1 | 3 tile | Mangsa harimau; merusak ladang |
| Babi hutan | pangan liar (bersaing dengan manusia) | 1 / 10 | sampai 4 / 1 | 2,5 | Melawan jika diserang; bisa dijinakkan → babi |
| Ayam hutan | biji dan serangga di semak/rumput | 0,4 / 5 | sampai 5 / 1 | 2 | Nenek moyang ayam, didomestikasi di Asia Tenggara (Wang dkk. 2020) |
| Kerbau liar | rumput (dekat air) | 3 / 20 | 1 / 2 | 1,5 | Melawan; bisa dijinakkan → kerbau |
| Harimau | hewan lain (jarang manusia) | 3 / 15 | sampai 3 / 2,5 | — | Pemangsa puncak |

- **Kepadatan:** peluang berkembang biak turun seiring jumlah hewan dewasa sejenis di wilayahnya (ketergantungan kepadatan: teritorialitas dan stres). Betina juga harus dalam kondisi baik dan ada jantan dalam 25 tile.
- **Pemangsa:** harimau berburu saat lapar. Ia mengendap mendekat, mangsa mungkin menyadarinya, lalu menerkam dengan peluang berhasil 15–60% sesuai ukuran mangsa. Energi yang didapat sebanding dengan besar mangsa, dan anak harimau diberi makan induknya. Harimau yang kelaparan sangat jarang menerkam orang yang sendirian.
- **Relung:** babi hutan memakan pangan liar, ayam hutan memakan biji dan serangga. Pada percobaan awal keduanya berebut makanan yang sama, dan ayam (kebutuhan per ekor paling rendah) menyingkirkan babi, sesuai teori R* (Tilman 1982).
- **Kepunahan lokal dan kedatangan ulang:** spesies yang punah di pulau punya peluang 5% per tahun untuk datang lagi sepasang dari seberang laut (biogeografi pulau, MacArthur & Wilson 1967). Laporan soak mencatat berapa kali ini terjadi.
- **Berburu** (`buru`) terpisah dari menyerang manusia, dan **bukan kejahatan**.
  - **Mengendap:** hewan menyadari orang dalam jarak larinya dengan peluang 10% per pengamatan (10 kali per detik) bila orang itu berjalan biasa. Peluangnya turun sampai 20% dari itu bila orang itu diam. Pemburu yang otaknya memutuskan `buru` dan melihat hewan bergerak paling cepat setengah kecepatannya. Mangsa memang lebih peka pada gerakan, dan pemburu nyata mendekat perlahan.
  - **Tombak** menambah jangkauan (dilempar) dan jauh lebih mematikan untuk berburu: rusa roboh dalam sekitar 2 tusukan tombak batu, babi 3, kerbau 5. Dengan tangan kosong, rusa butuh sekitar 5 pukulan.
  - **Buruan yang terluka melambat:** kecepatan lari dikali sisa kesehatannya (paling lambat 25%), sehingga pemburu bisa menyusul dan menyelesaikannya.
  - **Jerat** (`jerat`: 2 serat + 1 kayu, tanpa syarat teknologi): setiap rumah memasang satu jerat dalam 4 tile, di tempat yang paling banyak makanan liarnya. Rusa, babi, atau ayam hutan yang berdiri di atasnya tertangkap dengan laju 1,5 per detik. Kerbau terlalu besar dan harimau tidak ikut. Jerat yang berisi tangkapan tidak aktif sampai keluarganya mengambil daging saat berada di rumah, lalu terpasang lagi. Tangkapan dihitung sebagai hewan yang diburu. Pemburu-peramu Asia Tenggara (Punan, Batek, Agta) memakai jerat untuk babi, rusa, dan ayam hutan.
  - Babi, kerbau, dan harimau bisa melawan. Kematian karenanya dicatat sebagai "diterkam/diseruduk/ditanduk".
- **Domestikasi:** ayam dan babi yang sering berada di sekitar rumah perlahan kehilangan rasa takutnya (jalur komensal, Zeder 2012). Memberi makan hewan mempercepatnya. Saat cukup jinak, hewan itu menjadi milik rumah tangga dan **peternakan** ditemukan.
  - Ternak tinggal di dekat rumah, atau di **kandang** (aman dari harimau; kotorannya menyuburkan tanah).
  - Keluarga menyembelih ternak yang berlebih (lebih dari 6), atau saat lapar dan lumbung kosong.

Angka-angka riwayat hidup dibulatkan dari buku acuan (Corbet & Hill 1992; Keuling dkk. 2013 untuk babi hutan; Collias & Collias 1967 untuk ayam hutan; Sunquist 1981 dan Seidensticker 1986 untuk harimau). Laju makan, energi, dan batas kepadatan adalah **kalibrasi**, supaya tanpa manusia populasi berayun tetapi bertahan.

## 5. Ikan

- Ikan hidup di setiap tile air dan tumbuh logistik.
  - Laut dangkal berkapasitas 1,5, air tawar 1,0.
  - Ikan sungai dan danau ikut menyusut saat kemarau panjang (debit turun); ikan laut tidak.
- Memancing dilakukan dari tile di tepi air, dan tombak mempercepatnya.
- Penangkapan berlebih menguras petak pesisir. Ikan perlahan kembali dari perairan dalam.

## 6. Tubuh manusia (yang berubah di Fase 2)

| Aturan | Nilai | Dasar dan alasan |
| --- | --- | --- |
| Laju lapar | **3× lebih cepat** dari Fase 0–1. Tanpa makanan bertahan sekitar 3 tahun sim jika bergerak, sekitar 8 tahun jika diam | Manusia nyata bertahan tanpa makan sekitar 1–2 bulan (perkiraan). Dengan laju lama (sekitar 10 tahun), musim dan El Niño tidak pernah terasa. Uji 16 dunia: 4–6× membuat keluarga pertama mati kelaparan sebelum sempat berkembang (era 1 bertahan 0–4/16); 2,5× membuat populasi terlalu besar untuk anggaran performa |
| Makanan dibutuhkan | Rata-rata sekitar 0,19 energi per orang per tahun (semua umur, termasuk anak dan waktu istirahat), atau sekitar 0,5 unit pangan liar | Akibat laju lapar di atas; diukur dari tabel "Asal pangan" soak |
| Air minum | **Hanya air tawar**: sungai (sampai ke muaranya) dan danau. Air laut tidak bisa diminum | Sungai diambil dari model geologi peta; air yang tidak terhubung ke tepi peta dihitung danau. Di peta uji 128×128 hanya 31% daratan yang berjarak ≤ 10 tile dari air tawar, jadi permukiman tumbuh di tepi sungai seperti di dunia nyata |
| Laju haus | **8× lebih cepat** dari Fase 0–1: tanpa minum bertahan sekitar 4 tahun sim | Di dunia nyata haus membunuh dalam beberapa hari, lapar dalam beberapa minggu. Uji 32 seed × 20 menit: 12× membuat pendiri bertahan 25/32, 8× 29/32 (sama dengan sebelum ada aturan air). Jadi urutannya belum sepenuhnya realistis |
| Ingatan sumber air | Setiap orang mengingat tempat terakhir ia minum. Saat tak ada air terlihat, arah ke tempat itu terasa di sinar "air" sekuat air yang terlihat dekat. Anak mewarisi ingatan ibunya | Orang tahu di mana sungainya. Tanpa ingatan ini, haus 8× membuat pendiri bertahan 23/32; dengan ingatan 29/32 |
| Sumur | Butuh teknologi **pertanian** (semula alat batu). Orang bisa minum dalam 1,5 tile dari sumur | Sumur berdinding muncul bersama desa petani Neolitikum. Pemburu-peramu berpindah ke air. Dengan begitu, pertanian juga membuka lahan jauh dari sungai |
| Bayi | Di bawah 2 tahun digendong dan **disusui** ibu; ibu yang kelaparan berhenti menyusui | Penyapihan pada masyarakat pemburu-peramu sekitar 2–3 tahun (Sellen 2007). Bagian dari "ketergantungan anak" Fase 3 yang dimajukan karena perlu |
| Kebutuhan anak | Bayi baru lahir sekitar 35% kebutuhan orang dewasa, penuh pada umur 15 | Tubuh kecil butuh lebih sedikit |
| Kebutuhan perempuan | Perempuan dewasa butuh **78%** makanan dan air laki-laki. Anak perempuan dan laki-laki butuh sama banyak sampai umur 10; selisihnya tumbuh selama pubertas (10–15 tahun) | Data suku Hadza (Pontzer dkk. 2012): pengeluaran energi perempuan 71% laki-laki (1.877 lawan 2.649 kkal/hari), massa tubuh 84% (43 lawan 51 kg). Angka energi itu sebagian karena laki-laki Hadza berjalan dua kali lebih jauh per hari, sedangkan di simulasi kedua jenis kelamin bergerak sama banyak, jadi nilainya diambil di antara keduanya (kalibrasi). Hasil soak 16 dunia: tubuh perempuan sebesar laki-laki → rasio kelamin 152 (perempuan dewasa mati kelaparan dua kali lebih sering); 72% → 73–104 (laki-laki lebih sering mati); 84% → 110–120 (perempuan lagi). Bila selisihnya dimulai sejak lahir, anak laki-laki yang lebih sering mati (rasio 82–86) |
| Pasangan | Suami-istri yang berkemah dalam jarak 3 tile bisa hamil | Pasangan berbagi tempat bermalam |
| Batas populasi | **Tidak ada batas keras.** Hanya batas teknis = tile walkable ÷ 6 (300–2000) agar server tidak macet; laporan mencatat berapa kali batas ini tersentuh | Keputusan #2 di `PLAN.md` |

## 7. Pembusukan pangan

Waktu paruh di tempat terbuka (dibawa), **dikompres terhadap laju lapar** seperti angka lainnya:

| Pangan | Waktu paruh (th) |
| --- | --- |
| Daging, ikan | 0,5 |
| Pisang | 0,75 |
| Makanan liar campuran | 1 |
| Talas | 1,5 |
| Ubi | 2,5 |
| Kelapa | 3 |
| Padi (gabah), sagu | 5 |
| Daging asap, ikan asin | 10 |

Di dalam rumah, pangan tahan **1,5×** lebih lama; di **lumbung** **4×** lebih lama. Pengasapan dan pengasinan (teknologi **Pengawetan**) membuat daging dan ikan tahan bertahun-tahun.

## 8. Definisi angka di laporan soak

- **Populasi puncak:** populasi tertinggi pada sampel tiap menit simulasi.
- **Batas teknis tersentuh:** jumlah kehamilan yang batal karena batas teknis. Nilai 0 berarti batas itu tidak pernah aktif.
- **Keberadaan satwa:** bagian dari sampel tiap menit di mana spesies itu ada (≥ 1 ekor) di pulau.
- **Punah lokal / datang lagi:** jumlah peristiwa spesies habis di pulau, dan pasangan baru yang tiba.
- **Iklim dan kelaparan:** semua tahun dengan ≥ 20 penduduk, dari semua dunia, dikelompokkan menurut ENSO tahun itu.
  - "Mati kelaparan" = kematian karena lapar ÷ penduduk × 1.000.
  - "Tahun berikutnya" memakai tahun sesudahnya, karena cadangan tubuh membuat dampak kemarau baru terasa belakangan.
  - **Pembanding berpasangan:** setiap tahun El Niño (atau La Niña) dibandingkan dengan rata-rata tahun netral dalam ± 10 tahun di dunia yang sama (tahun dengan ≥ 20 penduduk). Cara ini menghilangkan pengaruh naik-turunnya populasi jangka panjang. "> 1,5×" adalah bagian kejadian yang tahun berikutnya kelaparannya lebih dari 1,5 kali biasanya.
- **Asal pangan:** energi yang benar-benar dimakan selama 100 tahun terakhir tiap dunia, dibagi menurut asalnya. "Liar" mencakup buah, umbi, dan sayur yang dimakan di tempat maupun dibawa. "Ladang" adalah tanaman yang dipanen (atau benih liar yang dipetik lalu dimakan). Lalu ikan (termasuk ikan asin) dan daging (buruan, ternak, daging asap). "Panen per orang per tahun" dalam unit hasil panen.

## 9. Batasan yang diketahui

- **Daging terlalu berharga dibanding kebutuhan.** Karena waktu dimampatkan, satu rusa (6 unit daging) setara makanan satu orang selama sekitar 12 tahun simulasi, padahal di dunia nyata sekitar 17 hari. Dengan populasi puluhan sampai ratusan orang, perburuan (termasuk jerat) hanya menyumbang di bawah 1% makanan, dan tidak pernah memunahkan satwa. Kepunahan karena perburuan butuh pemburu yang jauh lebih banyak atau daging per hewan yang jauh lebih kecil.
- Haus masih sedikit lebih lambat dari lapar (sekitar 4 tahun lawan 3 tahun saat bergerak). Di dunia nyata haus jauh lebih cepat.
- Satu tile mewakili lahan kecil, jadi kepadatan absolut tidak bermakna. Yang bermakna hanya perbandingan.
- Penyakit belum ada (Fase 3). Karena itu kelaparan menjadi penyebab utama kematian, padahal di masyarakat nyata penyakit lebih dominan.
- Pengetahuan "di mana makanan berada" tidak diingat. Makhluk hanya melihat sejauh pandangannya.

## Referensi

- Aldrian, E. & Susanto, R. D. (2003). Identification of three dominant rainfall regions within Indonesia and their relationship to sea surface temperature. *International Journal of Climatology* 23.
- Collias, N. E. & Collias, E. C. (1967). A field study of the red jungle fowl in north-central India. *The Condor* 69.
- Corbet, G. B. & Hill, J. E. (1992). *The Mammals of the Indomalayan Region.* Oxford University Press.
- Denham, T. P. dkk. (2003). Origins of agriculture at Kuk Swamp in the highlands of New Guinea. *Science* 301.
- Flach, M. (1997). *Sago palm.* IPGRI.
- Hendon, H. H. (2003). Indonesian rainfall variability: impacts of ENSO and local air–sea interaction. *Journal of Climate* 16.
- Institute of Medicine (2004). *Dietary Reference Intakes for Water, Potassium, Sodium, Chloride, and Sulfate.* National Academies Press.
- Keuling, O. dkk. (2013). Mortality rates of wild boar *Sus scrofa* L. in central Europe. *European Journal of Wildlife Research* 59.
- MacArthur, R. H. & Wilson, E. O. (1967). *The Theory of Island Biogeography.* Princeton University Press.
- McBride, J. L., Haylock, M. R. & Nicholls, N. (2003). Relationships between the Maritime Continent heat source and the El Niño–Southern Oscillation phenomenon. *Journal of Climate* 16.
- Naylor, R. L. dkk. (2007). Assessing risks of climate variability and climate change for Indonesian rice agriculture. *PNAS* 104.
- Pontzer, H. dkk. (2012). Hunter-gatherer energetics and human obesity. *PLoS ONE* 7(7): e40503.
- Seidensticker, J. (1986). Large carnivores and the consequences of habitat insularization: ecology and conservation of tigers in Indonesia and Bangladesh. Dalam *Cats of the World.* National Wildlife Federation.
- Sellen, D. W. (2007). Evolution of infant and young child feeding: implications for contemporary public health. *Annual Review of Nutrition* 27.
- Sunquist, M. E. (1981). The social organization of tigers (*Panthera tigris*) in Royal Chitawan National Park, Nepal. *Smithsonian Contributions to Zoology* 336.
- Tilman, D. (1982). *Resource Competition and Community Structure.* Princeton University Press.
- Wang, M.-S. dkk. (2020). 863 genomes reveal the origin and domestication of chicken. *Cell Research* 30.
- Zeder, M. A. (2012). Pathways to animal domestication. Dalam *Biodiversity in Agriculture.* Cambridge University Press.

Semua referensi perlu dicek ulang ke sumber primernya sebelum dikutip di tempat lain. Angka simulasi yang tidak tertera di tabel adalah kalibrasi.
