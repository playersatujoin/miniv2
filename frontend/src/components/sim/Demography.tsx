import { useQuery } from '@tanstack/react-query'
import { ApiError } from '../../api/client'
import { demographyQuery } from '../../sim/api'
import {
  DEATH_LABELS,
  type DeathCause,
  type DemographyMetrics,
  type MetricRef,
} from '../../sim/protocol'
import { AgePyramid } from './AgePyramid'
import { DemographyHistory } from './DemographyHistory'
import { fmtGini, nf, nf1, pct } from './format'

export type MetricKey = Exclude<keyof DemographyMetrics, 'windowYears' | 'personYears' | 'deathsByCause'>

type MetricSpec = {
  key: MetricKey
  label: string
  value: (v: number) => string
  range: (lo: number, hi: number) => string
}

const years = (v: number) => `${nf1.format(v)} tahun`
const yearsRange = (lo: number, hi: number) => `${nf1.format(lo)}–${nf1.format(hi)} tahun`

export const METRIC_SPECS: MetricSpec[] = [
  { key: 'lifeExpectancy', label: 'Harapan hidup saat lahir', value: years, range: yearsRange },
  { key: 'survivalTo15', label: 'Bayi bertahan sampai 15 tahun', value: pct, range: (lo, hi) => `${pct(lo)}–${pct(hi)}` },
  { key: 'lifeExpectancy15', label: 'Sisa harapan hidup di umur 15', value: (v) => `+${years(v)}`, range: yearsRange },
  { key: 'infantMortality', label: 'Kematian bayi (< 1 tahun)', value: pct, range: (lo, hi) => `${pct(lo)}–${pct(hi)}` },
  { key: 'modalAgeAdultDeath', label: 'Usia wafat terumum (dewasa)', value: years, range: yearsRange },
  {
    key: 'tfr',
    label: 'Anak per perempuan (TFR)',
    value: (v) => `${nf1.format(v)} anak`,
    range: (lo, hi) => `${nf1.format(lo)}–${nf1.format(hi)} anak`,
  },
  { key: 'meanAgeFirstBirth', label: 'Usia ibu saat anak pertama', value: years, range: yearsRange },
  { key: 'meanBirthInterval', label: 'Jarak antar-kelahiran', value: years, range: yearsRange },
  {
    key: 'sexRatio',
    label: 'Rasio kelamin',
    value: (v) => `${nf.format(Math.round(v))} ♂ per 100 ♀`,
    range: (lo, hi) => `${nf.format(lo)}–${nf.format(hi)}`,
  },
  {
    key: 'householdSize',
    label: 'Anggota per rumah',
    value: (v) => `${nf1.format(v)} orang`,
    range: (lo, hi) => `${nf1.format(lo)}–${nf1.format(hi)} orang`,
  },
  {
    key: 'homicideRate',
    label: 'Pembunuhan',
    value: (v) => `${nf.format(Math.round(v))} per 100.000/thn`,
    range: (lo, hi) => `${nf.format(lo)}–${nf.format(hi)} per 100.000/thn`,
  },
  {
    key: 'gini',
    label: 'Ketimpangan kekayaan (Gini)',
    value: fmtGini,
    range: (lo, hi) => `${fmtGini(lo)}–${fmtGini(hi)}`,
  },
]

type Standing = 'in' | 'above' | 'below'

const STANDING: Record<Standing, { icon: string; text: string }> = {
  in: { icon: '✓', text: 'sesuai acuan' },
  above: { icon: '↑', text: 'di atas acuan' },
  below: { icon: '↓', text: 'di bawah acuan' },
}

function standing(v: number, ref: MetricRef): Standing {
  if (v < ref.low) return 'below'
  if (v > ref.high) return 'above'
  return 'in'
}

const refTitle = (ref: MetricRef) => `${ref.source}${ref.note ? ` — ${ref.note}` : ''}`

/** A missing endpoint (older server) won't appear by retrying; other errors might. */
const retry = (count: number, err: Error) => !(err instanceof ApiError && err.status === 404) && count < 2

/** Demographic indicators of the living world next to pre-modern reference ranges. */
export function Demography({ mapId }: { mapId: string }) {
  const q = useQuery({ ...demographyQuery(mapId), retry })

  if (!q.data) {
    const missing = q.error instanceof ApiError && q.error.status === 404
    return (
      <section className="obs-section dm">
        <h3 className="obs-title">Demografi</h3>
        <p className={`small ${q.isError && !missing ? 'obs-error' : 'muted'}`}>
          {q.isPending
            ? 'Memuat demografi…'
            : missing
              ? 'Data demografi belum tersedia di server ini.'
              : `Gagal memuat demografi: ${q.error?.message ?? ''}`}
        </p>
      </section>
    )
  }

  const d = q.data
  const m = d.current
  const ref = d.reference ?? {}
  const sources = [...new Set(Object.values(ref).map((r) => r?.source).filter(Boolean))] as string[]

  return (
    <section className="obs-section dm" aria-labelledby="dm-title">
      <div className="obs-section-head">
        <h3 className="obs-title" id="dm-title">
          Demografi
        </h3>
        <span className="small muted">{nf.format(m?.windowYears ?? 50)} tahun terakhir</span>
      </div>
      <p className="dm-caveat small muted">
        Dibandingkan dengan masyarakat pra-modern. Penyakit, menyusui, dan menopause belum dimodelkan, jadi selisih
        dengan acuan adalah temuan, bukan galat.
        {m && m.personYears > 0 && <> Dasar hitungan: {nf.format(Math.round(m.personYears))} tahun-orang.</>}
      </p>

      <div className="dm-grid">
        {METRIC_SPECS.map((spec) => {
          const v = m?.[spec.key] ?? null
          const r = ref[spec.key]
          const st = v != null && r ? standing(v, r) : null
          return (
            <div key={spec.key} className="dm-tile">
              <span className="dm-label">{spec.label}</span>
              <strong className="dm-value">
                {v == null ? <span className="dm-na">belum cukup data</span> : spec.value(v)}
              </strong>
              {r && (
                <span className={`dm-ref ${st ? `dm-${st}` : ''}`} title={refTitle(r)}>
                  {st && (
                    <>
                      <i className="dm-ref-icon" aria-hidden="true">
                        {STANDING[st].icon}
                      </i>{' '}
                      {STANDING[st].text} ·{' '}
                    </>
                  )}
                  pra-modern {spec.range(r.low, r.high)}
                </span>
              )}
            </div>
          )
        })}
      </div>

      <AgePyramid bands={d.pyramid ?? []} />
      <CauseBars causes={m?.deathsByCause} windowYears={m?.windowYears ?? 50} />
      <DemographyHistory history={d.history ?? []} reference={ref} />

      {sources.length > 0 && (
        <details className="dm-sources small muted">
          <summary>Sumber acuan</summary>
          <ul>
            {sources.map((s) => (
              <li key={s}>{s}</li>
            ))}
          </ul>
          <p>Rincian dan batasannya: docs/reference-demography.md</p>
        </details>
      )}
    </section>
  )
}

/** Deaths in the window by cause: one hue, values and shares written out. */
function CauseBars({ causes, windowYears }: { causes: Record<DeathCause, number> | undefined; windowYears: number }) {
  const entries = (Object.keys(DEATH_LABELS) as DeathCause[]).map((c) => [c, causes?.[c] ?? 0] as const)
  const total = entries.reduce((sum, [, n]) => sum + n, 0)
  const max = Math.max(1, ...entries.map(([, n]) => n))

  return (
    <div className="dm-block">
      <h4 className="dm-subtitle">Penyebab kematian ({nf.format(windowYears)} tahun terakhir)</h4>
      {total === 0 ? (
        <p className="small muted">Belum ada kematian dalam jendela ini.</p>
      ) : (
        <ul className="dm-causes">
          {entries.map(([cause, n]) => (
            <li key={cause}>
              <span className="dm-cause-label">{DEATH_LABELS[cause] ?? cause}</span>
              <span className="dm-cause-track" aria-hidden="true">
                {n > 0 && <span className="dm-cause-bar" style={{ width: `${(n / max) * 100}%` }} />}
              </span>
              <span className="dm-cause-val">
                {nf.format(n)} · {Math.round((n / total) * 100)}%
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
