import { useQuery } from '@tanstack/react-query'
import { ApiError } from '../../api/client'
import { brainsQuery, creatureQuery } from '../../sim/api'
import { SEX_SYMBOL, type Brain, type BrainsSummary } from '../../sim/protocol'
import { BrainView } from './BrainView'
import { LineChart, type LineSeries } from './LineChart'
import { SEX_COLOR, formatClock, formatClockShort, nf, nf1, pct, useSecondsPerYear } from './format'

type Point = NonNullable<BrainsSummary['history']>[number]

// Palette slots validated on the panel surface (as in the other charts).
const SERIES: LineSeries<Point>[] = [
  { key: 'max', label: 'Otak terbesar', color: '#8ea0bb', value: (p) => p.max },
  { key: 'total', label: 'Rata-rata (bawaan + tumbuh)', color: '#c98500', value: (p) => p.avgTotal },
  { key: 'grown', label: 'Rata-rata tumbuh', color: '#a78bfa', value: (p) => p.avgGrown },
]

const one = (v: number) => nf1.format(v)

function Stat({ label, value, title }: { label: string; value: string; title?: string }) {
  return (
    <div className="obs-stat" title={title}>
      <span className="obs-stat-label">{label}</span>
      <strong>{value}</strong>
    </div>
  )
}

/** People by brain size, in bins of a few neurons. */
function Histogram({ b }: { b: BrainsSummary }) {
  const bins = b.histogram
  const last = Math.max(0, bins.findLastIndex((n) => n > 0))
  const shown = bins.slice(0, Math.max(last + 1, Math.ceil(64 / b.histogramBin)))
  const top = Math.max(1, ...shown)
  return (
    <section className="obs-section">
      <h3 className="obs-title">Sebaran ukuran otak</h3>
      <div className="neuron-hist" role="img" aria-label="Jumlah orang menurut jumlah neuron">
        {shown.map((n, i) => (
          <div key={i} className="neuron-hist-bar" title={`${i * b.histogramBin}–${(i + 1) * b.histogramBin - 1} neuron: ${nf.format(n)} orang`}>
            <span style={{ height: `${(n / top) * 100}%` }} />
            <small>{i % 2 === 0 ? i * b.histogramBin : ''}</small>
          </div>
        ))}
      </div>
      <p className="small muted">
        Jumlah orang menurut banyaknya neuron (bawaan + tumbuh). Tidak ada batas: yang membatasi hanya energi, karena setiap neuron makan energi.
      </p>
    </section>
  )
}

/** When in this life each of its grown neurons appeared, on an age axis. */
function GrowthTimeline({ brain, age }: { brain: Brain; age: number }) {
  const born = brain.grownBorn ?? []
  const use = brain.grownUse ?? []
  const span = Math.max(1, age)
  return (
    <div className="neuron-timeline" aria-label="Kapan neuron tumbuh">
      <div className="neuron-timeline-track">
        {born.map((b, k) => (
          <i
            key={k}
            style={{ left: `${(b / span) * 100}%`, opacity: 0.45 + 0.55 * Math.min(1, (use[k] ?? 0) * 8) }}
            title={`g${k + 1}: tumbuh pada usia ${one(b)} tahun · dipakai ${(use[k] ?? 0).toFixed(3).replace('.', ',')}`}
          />
        ))}
      </div>
      <div className="neuron-timeline-axis small muted">
        <span>lahir</span>
        <span>usia {one(age)} tahun</span>
      </div>
    </div>
  )
}

function SelectedBrain({ mapId, id }: { mapId: string; id: number }) {
  const q = useQuery(creatureQuery(mapId, id))
  const spy = useSecondsPerYear()
  if (!q.data) {
    const gone = q.error instanceof ApiError && q.error.status === 404
    return <p className="small muted">{gone ? 'Makhluk ini telah tiada.' : 'Memuat otaknya…'}</p>
  }
  const c = q.data
  const b = c.brain
  const age = c.age / spy
  const grown = b.grown ?? 0
  return (
    <section className="obs-section neuron-selected">
      <div className="obs-section-head">
        <h3 className="obs-title">
          Otak <span style={{ color: SEX_COLOR[c.sex] }}>{SEX_SYMBOL[c.sex]}</span> {c.name}
        </h3>
        <span className="small muted">usia {one(age)} tahun</span>
      </div>
      <div className="obs-stat-grid">
        <Stat label="Neuron total" value={nf.format(b.hidden.length + grown)} />
        <Stat label="Bawaan" value={nf.format(b.hidden.length)} title="Diwarisi dari kedua induk, bertambah/berkurang lewat evolusi" />
        <Stat label="Tumbuh" value={nf.format(grown)} title="Tumbuh selama hidup karena pengalaman; tidak diwariskan" />
        {b.grew != null && <Stat label="Pernah tumbuh / dipangkas" value={`${nf.format(b.grew)} / ${nf.format(b.pruned ?? 0)}`} />}
        {b.novelty != null && (
          <Stat
            label="Rasa kejutan"
            value={b.novelty.toFixed(2).replace('.', ',')}
            title="Seberapa sering hidup tak sesuai harapannya belakangan ini; inilah yang memicu neuron baru"
          />
        )}
        {b.neurogenesis != null && (
          <Stat label="Gen neurogenesis" value={b.neurogenesis.toFixed(2).replace('.', ',')} title="Seberapa mudah neuron baru tumbuh (diwariskan)" />
        )}
      </div>
      {grown > 0 || (b.grew ?? 0) > 0 ? (
        <GrowthTimeline brain={b} age={age} />
      ) : (
        <p className="small muted">Belum ada neuron yang tumbuh di otak ini.</p>
      )}
      <BrainView brain={b} />
    </section>
  )
}

/**
 * The Neuron tab: brains across the island (how big, how fast they grow and
 * are pruned, the biggest alive) and, for the person being watched, the
 * whole brain with the neurons it grew in its own life.
 */
export function NeuronView({ mapId, selectedId, onSelect }: { mapId: string; selectedId: number | null; onSelect: (id: number) => void }) {
  const q = useQuery(brainsQuery(mapId))
  const spy = useSecondsPerYear()
  const b = q.data
  return (
    <div className="neuron-view">
      <div className="neuron-side">
        {!b ? (
          <p className={`small ${q.isError ? 'obs-error' : 'muted'}`}>
            {q.isError
              ? q.error instanceof ApiError && q.error.status === 404
                ? 'Data otak belum tersedia di server ini.'
                : `Gagal memuat data otak: ${q.error.message}`
              : 'Memuat data otak…'}
          </p>
        ) : (
          <>
            <div className="obs-stats">
              <div className="obs-stat-grid">
                <Stat label="🧠 Rata-rata" value={`${one(b.avgTotal)} neuron`} />
                <Stat label="Bawaan" value={one(b.avgInherited)} title="Rata-rata neuron yang diwarisi" />
                <Stat label="Tumbuh" value={one(b.avgGrown)} title="Rata-rata neuron yang tumbuh selama hidup" />
                <Stat label="Terbesar" value={nf.format(b.max)} />
                <Stat label="Tumbuh/thn" value={one(b.grownPerYear)} title={`Total sepanjang masa: ${nf.format(b.grown)}`} />
                <Stat label="Dipangkas/thn" value={one(b.prunedPerYear)} title={`Total sepanjang masa: ${nf.format(b.pruned)}`} />
                <Stat label="Gen neurogenesis" value={b.avgNeurogenesis.toFixed(2).replace('.', ',')} title="Rata-rata populasi; evolusi menaikkannya bila otak yang tumbuh menguntungkan" />
                <Stat label="Neuron lambat" value={pct(b.slowNeurons)} title="Bagian neuron bawaan yang cukup lambat untuk menyimpan ingatan (τ ≥ 3)" />
              </div>
            </div>
            <LineChart
              title="Ukuran otak dari waktu ke waktu"
              points={b.history ?? []}
              x={(p) => p.time}
              xLabel={(p) => formatClockShort(p.time, spy)}
              tooltipTitle={(p) => formatClock(p.time, spy)}
              series={SERIES}
              fmt={one}
              floor={30}
              empty="Grafik muncul setelah beberapa detik simulasi berjalan."
            />
            <Histogram b={b} />
            {b.biggest?.length ? (
              <section className="obs-section">
                <h3 className="obs-title">Otak terbesar</h3>
                <ol className="neuron-top">
                  {b.biggest.map((e) => (
                    <li key={e.id}>
                      <button type="button" className={e.id === selectedId ? 'active' : ''} onClick={() => onSelect(e.id)}>
                        <span style={{ color: SEX_COLOR[e.sex] }}>{SEX_SYMBOL[e.sex]}</span> {e.name}
                        <span className="muted small"> · {one(e.age)} th</span>
                        <strong>
                          {e.inherited + e.grown}
                          {e.grown > 0 && <em className="neuron-grown"> +{e.grown}</em>}
                        </strong>
                      </button>
                    </li>
                  ))}
                </ol>
              </section>
            ) : null}
          </>
        )}
        <p className="small muted neuron-note">
          Setiap otak mewarisi neuron dari kedua induk. Selama hidup, ketika hidup terus memberi kejutan (lebih baik atau lebih
          buruk dari yang diharapkan), neuron baru tumbuh — lebih cepat pada anak-anak, dan hanya saat perut cukup kenyang.
          Neuron baru menyalin keadaan saat itu (menjadi "pengenal" situasi tersebut) lalu belajar apa yang harus dilakukan.
          Yang tak pernah berguna dipangkas lagi, dan otak yang kelaparan menyusut. Setiap neuron butuh energi.
        </p>
      </div>
      <div className="neuron-main">
        {selectedId !== null ? (
          <SelectedBrain mapId={mapId} id={selectedId} />
        ) : (
          <p className="small muted">Klik seseorang di peta atau di daftar "Otak terbesar" untuk melihat seluruh otaknya.</p>
        )}
      </div>
    </div>
  )
}
