import { useQuery } from '@tanstack/react-query'
import { useEffect, useMemo, useRef } from 'react'
import { ApiError } from '../../api/client'
import { mapQuery, tilesQuery } from '../../api/queries'
import { healthQuery } from '../../sim/api'
import { DEATH_LABELS, type DiseaseKey, type EpiPoint, type HealthMap } from '../../sim/protocol'
import { islandBackdrop, layerCanvas, type HealthLayer } from './healthMaps'
import { Meter } from './Ecology'
import { LineChart, type LineSeries } from './LineChart'
import { formatClock, formatClockShort, nf, nf1, pct, useSecondsPerYear } from './format'

/** A missing endpoint (older server) won't appear by retrying; other errors might. */
const retry = (count: number, err: Error) => !(err instanceof ApiError && err.status === 404) && count < 2

// Validated categorical slots on the panel surface, in the same order everywhere.
const DISEASE_COLORS: Record<DiseaseKey, string> = { diarrhea: '#c98500', malaria: '#d55181', respiratory: '#3987e5' }

const EPI_SERIES: LineSeries<EpiPoint>[] = [
  { key: 'diarrhea', label: 'Diare', color: DISEASE_COLORS.diarrhea, value: (p) => p.diarrhea },
  { key: 'malaria', label: 'Malaria', color: DISEASE_COLORS.malaria, value: (p) => p.malaria },
  { key: 'respiratory', label: 'ISPA', color: DISEASE_COLORS.respiratory, value: (p) => p.respiratory },
]

const ROUTE: Record<DiseaseKey, string> = {
  diarrhea: 'Air dan halaman tercemar kotoran. Paling berbahaya bagi balita yang mulai disapih.',
  malaria: 'Gigitan nyamuk dari genangan, sawah, dan rawa dataran rendah. Kekebalan terbentuk perlahan setelah sering terkena.',
  respiratory: 'Menular lewat kedekatan. Galur baru datang sesekali dan menyebar cepat di permukiman padat.',
}

/** The island's health: who is ill, the epidemic curve, and what spreads each disease. */
export function Health({ mapId }: { mapId: string }) {
  const q = useQuery({ ...healthQuery(mapId), retry })
  const spy = useSecondsPerYear()
  if (q.isError) {
    return <p className="obs-error small">Data kesehatan belum tersedia: {q.error.message}</p>
  }
  const h = q.data
  if (!h) return <p className="small muted">Memuat data kesehatan…</p>
  if (h.off) return <p className="small muted">Penyakit dimatikan di dunia ini (uji A/B).</p>

  const pop = Math.max(1, h.now.population)
  const ill = h.now.diarrhea + h.now.malaria + h.now.respiratory
  const deaths = h.diseases.reduce((n, d) => n + d.deaths, 0)
  return (
    <>
      <section className="obs-section" aria-labelledby="health-now-title">
        <div className="obs-section-head">
          <h3 className="obs-title" id="health-now-title">
            Kesehatan pulau
          </h3>
          <span className="small muted">
            {nf.format(ill)} sakit dari {nf.format(h.now.population)} orang
          </span>
        </div>
        <ul className="health-diseases">
          {h.diseases.map((d) => (
            <li key={d.key}>
              <div className="health-row">
                <span className="health-name">
                  <i className="obs-swatch-sm" style={{ background: DISEASE_COLORS[d.key] }} aria-hidden /> {DEATH_LABELS[d.key] ?? d.name}
                </span>
                <strong className="health-num">{nf.format(d.ill)} sakit</strong>
              </div>
              <div className="health-row small">
                <span className="muted" title="Episode baru per orang per tahun (50 tahun terakhir)">
                  {d.incidence == null ? 'belum ada data' : `${nf1.format(d.incidence)} episode/orang/tahun`}
                </span>
                <span className="muted" title="Kematian sepanjang masa">
                  {nf.format(d.deaths)} wafat
                </span>
              </div>
              <span className="health-route small muted">{ROUTE[d.key]}</span>
            </li>
          ))}
        </ul>
        <div className="obs-meters">
          <Meter label="Sedang sakit" value={ill / pop} text={pct(ill / pop)} color="#e66767" />
          <Meter
            label="Pembawa malaria tanpa gejala"
            value={h.now.carriers / pop}
            text={`${nf.format(h.now.carriers)} orang`}
            color="#d55181"
          />
          <Meter
            label="Cacingan (rata-rata)"
            value={h.now.worms}
            text={h.now.worms < 0.05 ? 'hampir tidak ada' : pct(h.now.worms)}
            color="#9b7f4c"
          />
        </div>
        <dl className="obs-kv">
          <dt>🚽 Jamban</dt>
          <dd>{nf.format(h.latrines)}</dd>
          <dt>🪣 Sumur</dt>
          <dd>{nf.format(h.wells)}</dd>
          <dt title="Rata-rata gen bawaan; lebih kuat berarti butuh energi lebih banyak">🧬 Gen kekebalan</dt>
          <dd>{h.immunity.toFixed(2)}×</dd>
          <dt>Galur ISPA</dt>
          <dd>{nf.format(h.strains)} sejauh ini</dd>
          <dt>Wafat karena penyakit</dt>
          <dd>{nf.format(deaths)}</dd>
        </dl>
      </section>

      <section className="obs-section" aria-labelledby="health-curve-title">
        <div className="obs-section-head">
          <h3 className="obs-title" id="health-curve-title">
            Kurva wabah
          </h3>
        </div>
        <LineChart
          title="Orang yang sedang sakit"
          points={h.history}
          x={(p) => p.time}
          xLabel={(p) => formatClockShort(p.time, spy)}
          tooltipTitle={(p) => formatClock(p.time, spy)}
          series={EPI_SERIES}
          fmt={(v) => nf.format(Math.round(v))}
          floor={5}
          extra={(p) => (
            <span>
              {nf.format(p.population)} orang · {p.deaths > 0 ? `${nf.format(p.deaths)} wafat · ` : ''}gigitan nyamuk {nf1.format(p.mosquitoes)}
            </span>
          )}
          empty="Kurva muncul setelah beberapa detik simulasi berjalan."
        />
        <p className="small muted eco-note">
          Tidak ada yang tahu soal kuman. Diare menular lewat air dan halaman yang tercemar, malaria lewat nyamuk, ISPA lewat
          kedekatan, dan cacing lewat tanah. Bayi dan lansia paling rentan, dan kurang gizi memperparah semuanya. Yang membantu
          hanyalah menyusui, keluarga yang merawat, jamban, dan sumur, tanpa ada yang tahu alasannya. Di peta, air yang tercemar tampak
          keruh dan kawanan nyamuk beterbangan di atas tempat mereka berkembang biak, terutama saat senja dan malam.
        </p>
      </section>

      <HealthMaps mapId={mapId} data={h.map} />
    </>
  )
}

const MAPS: { layer: HealthLayer; title: string; note: string }[] = [
  { layer: 'mosquito', title: '🦟 Nyamuk', note: 'kuning: banyak nyamuk · merah: banyak yang membawa malaria' },
  { layer: 'water', title: '💧 Air tercemar', note: 'hijau keruh sampai cokelat: makin banyak kuman' },
  { layer: 'worms', title: '🪱 Telur cacing di tanah', note: 'di sekitar tempat tinggal tanpa jamban' },
]

/** Where each disease lurks, over a dimmed picture of the island (observation only). */
function HealthMaps({ mapId, data }: { mapId: string; data: HealthMap }) {
  const map = useQuery(mapQuery(mapId))
  const tiles = useQuery(tilesQuery)
  const backdrop = useMemo(() => (map.data && tiles.data ? islandBackdrop(map.data, tiles.data) : null), [map.data, tiles.data])
  if (!backdrop || !map.data) return null
  return (
    <section className="obs-section" aria-labelledby="health-maps-title">
      <div className="obs-section-head">
        <h3 className="obs-title" id="health-maps-title">
          Sebaran penyakit
        </h3>
        <span className="small muted">diperbarui tiap 2 detik</span>
      </div>
      <div className="health-maps">
        {MAPS.map((m) => (
          <figure key={m.layer}>
            <MapCanvas backdrop={backdrop} layer={m.layer} data={data} width={map.data.width} height={map.data.height} label={m.title} />
            <figcaption>
              <strong>{m.title}</strong>
              <span className="small muted">{m.note}</span>
            </figcaption>
          </figure>
        ))}
      </div>
    </section>
  )
}

function MapCanvas({
  backdrop,
  layer,
  data,
  width,
  height,
  label,
}: {
  backdrop: HTMLCanvasElement
  layer: HealthLayer
  data: HealthMap
  width: number
  height: number
  label: string
}) {
  const ref = useRef<HTMLCanvasElement>(null)
  useEffect(() => {
    const canvas = ref.current
    if (!canvas) return
    const size = 2 * Math.max(width, height)
    canvas.width = size
    canvas.height = size
    const ctx = canvas.getContext('2d')!
    ctx.imageSmoothingEnabled = false
    ctx.drawImage(backdrop, 0, 0, size, size)
    const over = layerCanvas(layer, data, width, height)
    if (over) {
      ctx.imageSmoothingEnabled = layer !== 'water'
      ctx.drawImage(over, 0, 0, size, size)
    }
  }, [backdrop, layer, data, width, height])
  return <canvas ref={ref} className="health-map" role="img" aria-label={`Peta ${label}`} />
}
