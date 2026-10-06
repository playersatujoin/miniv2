# Acuan Demografi Pra-Modern

Angka-angka ini dipakai untuk membandingkan statistik simulasi dengan dunia nyata (Fase 0 di `PLAN.md`). Semuanya untuk **masyarakat pra-modern skala kecil** (pemburu-peramu dan peladang). Dunia simulasi dimulai dari Adam & Hawa dan baru sampai Zaman Batu/Logam, jadi itulah pembanding yang paling cocok.

Skala waktu simulasi: **1 tahun = 8 detik simulasi**.

| Indikator | Acuan | Sumber | Catatan |
| --- | --- | --- | --- |
| Harapan hidup saat lahir (e0) | **21–37 tahun** | Gurven & Kaplan (2007), teks hlm. "the average life expectancy at birth (e0) varies from 21 to 37 years" | Rendah karena kematian bayi dan anak tinggi, bukan karena orang dewasa mati muda |
| Peluang bayi bertahan sampai umur 15 (l15) | **rata-rata 0,57**; rentang antar-populasi **0,44–0,73** | Gurven & Kaplan (2007), Tabel 2–3 (kolom l15) | 0,64 untuk peramu-peladang; 0,67 untuk pemburu-peramu yang sudah berakulturasi |
| Sisa harapan hidup pada umur 15 (e15) | **sekitar 28–43 tahun lagi** | Gurven & Kaplan (2007), Tabel 3 (kolom e15) | Kelompok yang berakulturasi bisa sampai sekitar 52 tahun |
| Kematian bayi (q0, mati sebelum umur 1) | **rata-rata 0,27 (SD 0,07)**; rentang tampilan **0,13–0,41** (±2 SD) | Volk & Atkinson (2013), 20 populasi pemburu-peramu | Kematian sebelum pubertas rata-rata 0,49 (SD 0,06), sejalan dengan l15 di atas |
| Usia kematian dewasa yang paling umum (modus) | **68–78 tahun** | Gurven & Kaplan (2007), Tabel 4 | Kesimpulan penulis: tubuh manusia "dirancang" berfungsi baik sekitar tujuh dekade |
| Angka kelahiran total (TFR) | **sekitar 5–7 anak per perempuan** (rata-rata 6,2) | Kompilasi 5 populasi (Ache, Agta, Hadza, Hiwi, !Kung); lihat arXiv:2601.13442 | Kompilasi sekunder. Rata-rata etnografis lebih luas sekitar 5,4 (rentang 0,8–8,5); perlu dicek ke sumber primer |
| Jarak antar-kelahiran | **2,8–3,3 tahun** (rata-rata 3,1) | Kompilasi yang sama | Panjang karena menyusui. Sejak Fase 2, bayi di bawah 2 tahun digendong dan disusui ibunya (`docs/reference-ecology.md`); masa tidak subur 2 tahun setelah melahirkan tetap dipakai sebagai efek menyusui pada ovulasi |
| Umur ibu saat anak pertama | **sekitar 18–20 tahun** | Kompilasi yang sama; Baka: rata-rata 18 tahun (Ramirez Rozzi 2018) | — |
| Ketimpangan kekayaan (Gini) | **0,25 ± 0,04** (pemburu-peramu); 0,27 (peladang); 0,42 (penggembala); 0,48 (petani) | Borgerhoff Mulder dkk. (2009), Tabel 2, rata-rata semua jenis kekayaan berbobot | Di simulasi hanya kekayaan material (bawaan + simpanan rumah) yang dihitung, jadi angkanya lebih bisa dibandingkan dengan kolom "material" |
| Penyebab kematian | **Penyakit > 50% kematian** di hampir semua kelompok; kekerasan sangat bervariasi | Gurven & Kaplan (2007), Tabel 5 | Sejak Fase 3b ada diare, malaria, ISPA, cacingan, dan kematian neonatal. Kelaparan dan kehausan dewasa masih terlalu sering (berbagi pangan dan pindah kemah belum ada, Fase 4). Itu temuan, bukan galat |
| Angka pembunuhan | **Tidak diberi rentang** | — | Sangat bervariasi antar-masyarakat (rendah pada Hadza dan Tsimane, tinggi pada Ache dan Yanomamo). Dilaporkan tanpa penilaian |

## Definisi di simulasi

- **Jendela:** indikator dihitung dari 50 tahun simulasi terakhir (400 detik), dengan *period life table* berkelompok umur 0, 1–4, 5–9, …, 80+.
- **e0, e15, l15, q0:** dari tabel hidup periode tersebut (angka kematian per kelompok umur = kematian ÷ tahun-orang).
- **TFR:** 5 × jumlah angka kelahiran per kelompok umur ibu 15–49 tahun.
- **Gini:** kekayaan setiap orang dewasa = nilai barang bawaan + bagian rata dari simpanan rumahnya.
- **Angka pembunuhan:** kematian karena dibunuh per 100.000 tahun-orang.

## Batasan pembanding

- Masyarakat nyata punya penyakit, menyusui, menopause, dan budaya. Simulasi baru sebagian (budaya di Fase 1, menyusui sederhana di Fase 2; penyakit dan menopause di Fase 3). Selisih dengan acuan menunjukkan apa yang belum dimodelkan.
- Populasi simulasi kecil (sekitar 150 orang), jadi angkanya berfluktuasi. Bandingkan median beberapa seed dari runner soak, bukan satu dunia.

## Sumber

- Gurven, M. & Kaplan, H. (2007). Longevity among hunter-gatherers: a cross-cultural examination. *Population and Development Review* 33(2), 321–365. [PDF](https://gurven.anth.ucsb.edu/sites/secure.lsit.ucsb.edu.anth.d7_gurven/files/sitefiles/papers/GurvenKaplan2007pdr.pdf)
- Borgerhoff Mulder, M. dkk. (2009). Intergenerational wealth transmission and the dynamics of inequality in small-scale societies. *Science* 326, 682–688. [PDF](https://www.gurven.anth.ucsb.edu/sites/secure.lsit.ucsb.edu.anth.d7_gurven/files/sitefiles/papers/borgerhoffmulderetal2009.pdf)
- Ramirez Rozzi, F. V. (2018). Reproduction in the Baka pygmies and drop in their fertility with the arrival of alcohol. *PNAS*. [PMC6142234](https://pmc.ncbi.nlm.nih.gov/articles/PMC6142234)
- Volk, A. A. & Atkinson, J. A. (2013). Infant and child death in the human environment of evolutionary adaptation. *Evolution and Human Behavior* 34, 182–192.
- Kompilasi TFR dan jarak kelahiran 5 populasi: [arXiv:2601.13442](https://arxiv.org/pdf/2601.13442) (sumber sekunder; perlu dicek ke sumber primer).
