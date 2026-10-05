# Miniv2 — Peta 2D Top-down

Map 2D untuk game top-down, dihuni makhluk hidup yang berevolusi sendiri. Backend **Go** membuat peta secara prosedural dan menjalankan simulasi kehidupan; frontend **TanStack Router + TanStack Query** (React + Vite) merender peta di canvas, dengan tiga mode:

- **Amati** (default): kamu hanya pengamat, seperti melihat akuarium.
- **Main**: jalan-jalan dengan collision.
- **Edit**: cat tile, lalu simpan ke server.

```
backend/    Go 1.22+ (stdlib saja, tanpa dependency)
frontend/   React 19 + TanStack Router (file-based) + TanStack Query + Vite
```

## Menjalankan (development)

Terminal 1 — backend di `:8080`:

```bash
cd backend
go run .            # data peta disimpan di backend/data/*.json
```

Terminal 2 — frontend di `:5173` (request `/api` di-proxy ke `:8080`):

```bash
cd frontend
npm install
npm run dev
```

Buka http://localhost:5173. Saat pertama jalan, backend otomatis membuat peta "Starter Island".

## Production (satu binary)

```bash
cd frontend && npm run build
cd ../backend && go build -o mapserver . && ./mapserver -static ../frontend/dist
```

Semua (UI + API) dilayani di http://localhost:8080. Flag: `-addr`, `-data`, `-static`.

## Akuarium: peradaban dari Adam & Hawa

Setiap peta punya dunianya sendiri yang **terus berjalan di server Go** — juga saat browser ditutup — dan disimpan ke `backend/data/sims/<mapId>.json.gz` (tiap menit dan saat server berhenti).

- **Awal**: setiap dunia dimulai hanya dari **Adam ♂ & Hawa ♀**. Tidak ada pendatang. Jika semua manusia punah, Adam & Hawa baru memulai **era** berikutnya, dibiakkan dari genom paling sukses era sebelumnya; pengetahuan (teknologi, unsur) dunia tetap tersimpan.
- **Otak**: jaringan saraf rekuren milik setiap makhluk — 55 indra → 24 neuron (dengan memori) → 12 keputusan: belok, gerak, makan, minum, kawin, istirahat, **kumpulkan, buat, bangun, beri, curi, serang**. Indranya: 6 "sinar" penglihatan (rintangan, makanan, air, lawan jenis, sesama jenis, sumber daya) + kondisi diri (energi, hidrasi, kesehatan, rumah, keluarga dekat, orang asing, reputasi, diserang, bawaan, …) + input acak `acak (kehendak)`.
- **Kehendak sendiri**: tidak ada skrip perilaku — setiap tindakan berasal dari output otaknya; kode hanya menentukan detail (resep mana, target terdekat). Kebaikan (memberi makan yang lapar), kejahatan (mencuri, menyerang, membunuh) dan bertahan hidup semuanya pilihan otak. Anak-anak tidak bisa mencuri atau menyerang. Generasi pertama punya refleks bawaan lemah (bobot awal) yang bisa berubah lewat evolusi; mencuri dan menyerang awalnya ditekan, tapi evolusi bebas mengubahnya.
- **Keluarga**: pasangan terbentuk saat anak pertama; anak ikut rumah ibu (atau ayah). **Kepala keluarga** = pemilik rumah (siapa pun yang membangunnya). Saat ia meninggal, rumah diwarisi pasangan → anak tertua → kosong (bisa diklaim orang lain).
- **Evolusi**: anak mewarisi persilangan otak kedua orang tua (per neuron) + mutasi, juga sifat bawaan (ukuran, kecepatan, jarak pandang, metabolisme, umur, laju mutasi, warna keturunan).
- **Ekonomi & teknologi**: kumpulkan bahan → buat bahan baru (arang, tali, beliung, bata, tembaga, perunggu, besi, baja, kaca, semen, beton, …) → bangun gubuk / rumah kayu / rumah bata, ladang, sumur, dan stasiun riset. Setiap stasiun membuka zaman baru: Zaman Batu → Logam (tungku) → Kimia (laboratorium) → Listrik → Spektroskopi → Radiasi → Nuklir (reaktor) → Partikel (akselerator).
- **118 unsur**: seluruh tabel periodik ada. Unsur ditemukan dengan mengumpulkan (C, S, Cu, Ag, Au), meneliti mineral di stasiun sesuai zamannya, atau mensintesis unsur buatan di reaktor/akselerator.

Panel kanan: tab **Populasi** (statistik, grafik, peristiwa), **Peradaban** (tangga zaman, teknologi, bangunan), **Unsur** (tabel periodik). Klik makhluk atau rumah untuk melihat status, rumah & isinya, bawaan, perbuatan baik/jahat, keluarga, dan aktivitas jaringan sarafnya.

## Geologi realistis

Peta dibentuk seperti **pulau busur gunung api di atas zona subduksi** (Sumatra/Jawa/Halmahera): gunung api dengan kawah asam, intrusi granit & porfiri, pegmatit, sabuk ofiolit (ultrabasa), cekungan sedimen, karst gamping, dataran garam, karbonatit, dan sungai yang mengalir ke laut membentuk delta.

Endapan **hanya muncul di batuan tempat ia terbentuk di dunia nyata**, mengikuti model endapan nyata, misalnya: porfiri Cu-Au-Mo (Grasberg, Batu Hijau), epitermal Au-Ag (Pongkor), belerang kawah (Ijen), granit timah (Bangka-Belitung), laterit nikel (Sorowako), bauksit laterit (Kalimantan Barat), batu bara (Kutai), karbonatit REE (Mountain Pass), plaser sungai (emas/timah/pasir besi hanya di hilir batuan sumbernya). **Semua mineral harus digali dengan alat**, termasuk emas. Tombol **🪨 Geologi** menampilkan peta batuan, nama fitur, dan legenda model endapan; arahkan kursor ke tile untuk melihat endapan dan batuannya.

Catatan jujur: ini tetap penyederhanaan (skala tile, waktu, kimia), tapi tidak ada endapan yang ditempatkan sembarangan. Tes membuktikan bahwa pada setiap peta hasil generator ≥ 48×48 semua bangunan bisa dibuat dan semua unsur bisa ditemukan; peta yang lebih kecil bisa kekurangan beberapa jenis endapan.

Skala waktu: **1 tahun = 8 detik simulasi** (umur 60–80 tahun, dewasa di umur 15, hamil ±9 bulan). Tab Populasi punya bagian **Demografi**: harapan hidup, kesuburan, ketimpangan, dan piramida penduduk, dibandingkan dengan data masyarakat pra-modern nyata (`docs/reference-demography.md`).

Uji keseimbangan banyak dunia sekaligus (laporan ke `reports/<waktu>/summary.md`):

```bash
cd backend
go run ./cmd/soak -seeds 1-8 -minutes 120              # baseline
go run ./cmd/soak -seeds 1-48 -minutes 20 -off crime   # A/B: dunia tanpa kejahatan
```

Rencana pengembangan lengkap: `PLAN.md`.

## Kontrol

| Mode | Input |
| --- | --- |
| Amati | klik makhluk atau rumah untuk mengamati (kamera mengikuti), drag/`WASD` geser, scroll zoom, klik minimap untuk lompat, 🪨 Geologi untuk peta batuan |
| Main | `WASD`/panah jalan, `Shift` lari, scroll atau `+`/`-` zoom |
| Edit | klik kiri cat, klik kanan/tengah drag geser, `WASD` geser kamera, klik minimap untuk lompat, `Ctrl+Z`/`Ctrl+Y` undo/redo, `Ctrl+S` simpan |

Kuas: 1×1, 3×3, 5×5, atau **Isi** (flood fill). Alat **Titik Spawn** memindahkan posisi awal pemain.

## Format peta

Dua layer, array row-major (`index = y * width + x`) berisi ID tile:

```jsonc
{
  "id": "8c7408c85e45e09c", "name": "Starter Island",
  "width": 128, "height": 128, "tileSize": 32, "seed": 1337,
  "spawn": { "x": 64, "y": 64 },
  "layers": { "ground": [0, 0, 1, ...], "objects": [0, 0, 0, ...] }
}
```

Definisi tile (ID, key, nama, warna, `solid`) ada di `backend/internal/world/tiles.go` dan dikirim lewat `GET /api/tiles`, jadi frontend tidak meng-hardcode ID. Untuk menambah tile: tambahkan di `tiles.go`, lalu (opsional) gambarnya di `frontend/src/game/render.ts` berdasarkan `key` — tanpa gambar khusus, tile tetap tampil memakai warnanya.

## API

| Method | Path | Keterangan |
| --- | --- | --- |
| GET | `/api/tiles` | Definisi tile ground & objek |
| GET | `/api/maps` | Daftar peta (ringkasan) |
| POST | `/api/maps` | Generate peta `{name, width, height, seed?}` (16–256) |
| GET | `/api/maps/{id}` | Peta lengkap |
| PUT | `/api/maps/{id}` | Simpan `{name, spawn, layers}` (divalidasi) |
| DELETE | `/api/maps/{id}` | Hapus peta |
| GET | `/api/maps/{id}/preview.png?scale=2` | Thumbnail PNG |
| GET | `/api/maps/{id}/sim` | Ringkasan dunia: populasi, statistik, riwayat, peristiwa |
| PUT | `/api/maps/{id}/sim/speed` | `{speed: 0\|1\|2\|5\|10\|20}` (0 = jeda) |
| GET | `/api/maps/{id}/sim/creatures/{cid}` | Detail makhluk + otaknya (bobot & aktivasi) |
| GET | `/api/maps/{id}/geology` | Peta geologi: batuan per tile, fitur, model endapan, endapan |
| GET | `/api/maps/{id}/sim/knowledge` | 118 unsur, teknologi, jenis bangunan |
| GET | `/api/maps/{id}/sim/demography` | Demografi: tabel hidup, kesuburan, Gini, piramida, acuan pra-modern |
| GET | `/api/maps/{id}/sim/stream` | Server-Sent Events: `frame` (semua makhluk 10×/detik) dan `structures` (bangunan) |

## Cara kerja singkat

- **Generator** (`backend/internal/world/`): value-noise fBm untuk ketinggian & kelembapan, falloff tepi supaya berbentuk pulau, model geologi busur gunung api (`BuildGeoModel`, hanya dari seed sehingga tetap sama walau peta diedit), sungai dengan priority-flood, kawah, karst, lalu jalan tanah dengan A* (otomatis jadi jembatan di atas air). Deterministik per seed.
- **Kimia** (`backend/internal/chem/`): 118 unsur, item & mineral dengan komposisi nyata, resep, teknologi, bangunan, dan penempatan endapan menurut model endapan.
- **Renderer** (`frontend/src/game/`): ground dirender sekali ke chunk 16×16 tile yang di-cache (LRU), objek tinggi (pohon, batu, tembok) digambar sebagai sprite dan di-sort berdasarkan Y bersama pemain supaya bisa jalan "di belakang" pohon. Semua grafis prosedural — tidak ada file aset.
- **TanStack**: loader route memanggil `queryClient.ensureQueryData`, search param `?mode=watch|play|edit` divalidasi lewat `validateSearch`, `useBlocker` mencegah keluar halaman saat ada perubahan belum disimpan, dan mutation memperbarui cache Query setelah simpan.

## Test

```bash
cd backend && go test ./...
cd frontend && npm run typecheck
```
