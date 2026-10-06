import { useMemo, useState, type KeyboardEvent, type PointerEvent } from 'react'
import type { SimHistoryPoint } from '../../sim/protocol'
import { formatClock, formatClockShort, nf, nf1, pct, useSecondsPerYear } from './format'
import { useElementWidth } from './useElementWidth'

type NumKey = 'population' | 'females' | 'males' | 'elements' | 'houses' | 'avgBrainSize' | 'avgSkill'
type Metric = 'population' | 'elements' | 'houses' | 'brain' | 'skill'
type Series = { key: NumKey; label: string; color: string }
type MetricConfig = {
  tab: string
  title: string
  series: Series[]
  drawOrder: NumKey[]
  headline: NumKey
  /** Lowest y-axis maximum, so small values don't fill the whole chart. */
  floor: number
  fmt: (v: number) => string
}

const whole = (v: number) => nf.format(Math.round(v))

// Colours are slots from the dataviz palette validated on the panel surface
// (#172234, dark): the population trio passes all-pairs; the single-series
// metrics each use one validated slot. Each metric is its own chart with one axis.
const METRICS: Record<Metric, MetricConfig> = {
  population: {
    tab: 'Populasi',
    title: 'Populasi dari waktu ke waktu',
    series: [
      { key: 'population', label: 'Total', color: '#c98500' },
      { key: 'females', label: '♀ Perempuan', color: '#d55181' },
      { key: 'males', label: '♂ Laki-laki', color: '#3987e5' },
    ],
    // Draw the total last so it sits on top.
    drawOrder: ['females', 'males', 'population'],
    headline: 'population',
    // Food limits the population now, so the axis follows the data, not the technical ceiling.
    floor: 10,
    fmt: whole,
  },
  elements: {
    tab: 'Unsur',
    title: 'Unsur ditemukan dari waktu ke waktu',
    series: [{ key: 'elements', label: 'Unsur', color: '#9085e9' }],
    drawOrder: ['elements'],
    headline: 'elements',
    floor: 10,
    fmt: whole,
  },
  houses: {
    tab: 'Rumah',
    title: 'Jumlah rumah dari waktu ke waktu',
    series: [{ key: 'houses', label: 'Rumah', color: '#199e70' }],
    drawOrder: ['houses'],
    headline: 'houses',
    floor: 5,
    fmt: whole,
  },
  brain: {
    tab: 'Otak',
    title: 'Rata-rata ukuran otak (neuron tersembunyi)',
    series: [{ key: 'avgBrainSize', label: 'Neuron', color: '#d95926' }],
    drawOrder: ['avgBrainSize'],
    headline: 'avgBrainSize',
    floor: 16,
    fmt: (v) => nf1.format(v),
  },
  skill: {
    tab: 'Keahlian',
    title: 'Rata-rata keahlian terbaik orang dewasa',
    series: [{ key: 'avgSkill', label: 'Keahlian', color: '#008300' }],
    drawOrder: ['avgSkill'],
    headline: 'avgSkill',
    floor: 1,
    fmt: pct,
  },
}

const SURFACE = '#172234'
const GRID = '#24324a'
const AXIS = '#33476a'
const MUTED = '#8ea0bb'
const TEXT = '#e6edf7'

const HEIGHT = 156
const M = { top: 12, right: 32, bottom: 22, left: 30 }

function niceCeil(v: number) {
  if (v <= 0) return 10
  const step = 10 ** Math.floor(Math.log10(v))
  for (const m of [1, 2, 2.5, 5, 10]) if (m * step >= v) return m * step
  return 10 * step
}

const val = (h: SimHistoryPoint, key: NumKey) => (h[key] as number | undefined) ?? 0

type Props = { history: SimHistoryPoint[] }

export function PopulationChart({ history }: Props) {
  const [wrapRef, width] = useElementWidth<HTMLDivElement>()
  const [active, setActive] = useState<number | null>(null)
  const [showTable, setShowTable] = useState(false)
  const [metric, setMetric] = useState<Metric>('population')
  const spy = useSecondsPerYear()
  const ready = history.length >= 2 && width > 0

  // Older worlds don't report every metric; only offer what the data has.
  const available = (Object.keys(METRICS) as Metric[]).filter(
    (m) => m === 'population' || history.some((h) => typeof h[METRICS[m].headline] === 'number'),
  )
  const cfg = METRICS[available.includes(metric) ? metric : 'population']
  const colorOf = (key: NumKey) => cfg.series.find((s) => s.key === key)!.color

  const geo = useMemo(() => {
    if (!ready) return null
    const t0 = history[0].time
    const t1 = history[history.length - 1].time
    const peak = Math.max(...history.map((h) => val(h, cfg.headline)))
    const yMax = niceCeil(Math.max(cfg.floor, peak))
    const plotW = width - M.left - M.right
    const plotH = HEIGHT - M.top - M.bottom
    const x = (t: number) => M.left + ((t - t0) / Math.max(1e-9, t1 - t0)) * plotW
    const y = (v: number) => M.top + plotH - (v / yMax) * plotH
    const paths = {} as Record<NumKey, string>
    for (const s of cfg.series) {
      paths[s.key] = history.map((h, i) => `${i ? 'L' : 'M'}${x(h.time).toFixed(1)},${y(val(h, s.key)).toFixed(1)}`).join('')
    }
    return { t0, t1, yMax, plotW, plotH, x, y, paths }
  }, [ready, history, width, cfg])

  const nearest = (px: number) => {
    if (!geo) return null
    let best = 0
    let bestD = Infinity
    history.forEach((h, i) => {
      const d = Math.abs(geo.x(h.time) - px)
      if (d < bestD) {
        bestD = d
        best = i
      }
    })
    return best
  }

  const onPointerMove = (e: PointerEvent<SVGSVGElement>) => {
    const rect = e.currentTarget.getBoundingClientRect()
    setActive(nearest(e.clientX - rect.left))
  }

  const onKeyDown = (e: KeyboardEvent<SVGSVGElement>) => {
    if (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight') return
    e.preventDefault()
    const cur = active ?? history.length - 1
    setActive(Math.min(history.length - 1, Math.max(0, cur + (e.key === 'ArrowLeft' ? -1 : 1))))
  }

  const last = history[history.length - 1]
  const point = active !== null ? history[active] : null
  const multi = cfg.series.length > 1

  return (
    <section className="obs-section obs-chart">
      <div className="obs-section-head">
        <h3 className="obs-title">{cfg.title}</h3>
        {history.length >= 2 && (
          <button type="button" className="obs-mini" onClick={() => setShowTable((v) => !v)}>
            {showTable ? 'Grafik' : 'Tabel'}
          </button>
        )}
      </div>

      {available.length > 1 && (
        <div className="segmented obs-metric" role="group" aria-label="Pilih data grafik">
          {available.map((m) => (
            <button
              key={m}
              type="button"
              className={cfg === METRICS[m] ? 'active' : ''}
              aria-pressed={cfg === METRICS[m]}
              onClick={() => setMetric(m)}
            >
              {METRICS[m].tab}
            </button>
          ))}
        </div>
      )}

      {/* A single series is named by the title; a legend only appears for several. */}
      {multi && (
        <ul className="obs-legend">
          {cfg.series.map((s) => (
            <li key={s.key}>
              <i className="obs-line-key" style={{ background: s.color }} />
              {s.label}
            </li>
          ))}
        </ul>
      )}

      <div ref={wrapRef} className="obs-chart-wrap">
        {!ready ? (
          <p className="obs-empty small muted">Grafik muncul setelah beberapa detik simulasi berjalan.</p>
        ) : showTable ? (
          <HistoryTable history={history} series={cfg.series} spy={spy} fmt={cfg.fmt} />
        ) : (
          geo && (
            <>
              <svg
                width={width}
                height={HEIGHT}
                className="obs-chart-svg"
                tabIndex={0}
                role="img"
                aria-label={`${cfg.title}: ${cfg.series.map((s) => `${s.label} ${cfg.fmt(val(last, s.key))}`).join(', ')}. Gunakan panah kiri/kanan untuk menelusuri.`}
                onPointerMove={onPointerMove}
                onPointerLeave={() => setActive(null)}
                onFocus={() => setActive(history.length - 1)}
                onBlur={() => setActive(null)}
                onKeyDown={onKeyDown}
              >
                {[0, geo.yMax / 2, geo.yMax].map((v) => (
                  <g key={v}>
                    <line
                      x1={M.left}
                      x2={M.left + geo.plotW}
                      y1={geo.y(v)}
                      y2={geo.y(v)}
                      stroke={v === 0 ? AXIS : GRID}
                      strokeWidth={1}
                      shapeRendering="crispEdges"
                    />
                    <text x={M.left - 5} y={geo.y(v) + 3} textAnchor="end" className="obs-axis-text">
                      {cfg.fmt(v)}
                    </text>
                  </g>
                ))}

                <text x={M.left} y={HEIGHT - 6} className="obs-axis-text">
                  {formatClockShort(geo.t0, spy)}
                </text>
                <text x={M.left + geo.plotW} y={HEIGHT - 6} textAnchor="end" className="obs-axis-text">
                  {formatClockShort(geo.t1, spy)}
                </text>

                {cfg.drawOrder.map((key) => (
                  <path
                    key={key}
                    d={geo.paths[key]}
                    fill="none"
                    stroke={colorOf(key)}
                    strokeWidth={2}
                    strokeLinejoin="round"
                    strokeLinecap="round"
                  />
                ))}

                {point ? (
                  <g>
                    <line
                      x1={geo.x(point.time)}
                      x2={geo.x(point.time)}
                      y1={M.top}
                      y2={M.top + geo.plotH}
                      stroke={MUTED}
                      strokeOpacity={0.6}
                      strokeWidth={1}
                    />
                    {cfg.drawOrder.map((key) => (
                      <circle
                        key={key}
                        cx={geo.x(point.time)}
                        cy={geo.y(val(point, key))}
                        r={4}
                        fill={colorOf(key)}
                        stroke={SURFACE}
                        strokeWidth={2}
                      />
                    ))}
                  </g>
                ) : (
                  <g>
                    {cfg.drawOrder.map((key) => (
                      <circle
                        key={key}
                        cx={geo.x(last.time)}
                        cy={geo.y(val(last, key))}
                        r={4}
                        fill={colorOf(key)}
                        stroke={SURFACE}
                        strokeWidth={2}
                      />
                    ))}
                    {/* Label only the headline series; the sexes converge, so they use legend + tooltip. */}
                    <text x={geo.x(last.time) + 7} y={geo.y(val(last, cfg.headline)) + 4} fill={TEXT} className="obs-end-label">
                      {cfg.fmt(val(last, cfg.headline))}
                    </text>
                  </g>
                )}
              </svg>

              {point && (
                <div
                  className="obs-tooltip"
                  style={{ left: Math.min(width - 168, Math.max(0, geo.x(point.time) + 10)), top: 4 }}
                >
                  <div className="obs-tooltip-title">{formatClock(point.time, spy)}</div>
                  {cfg.series.map((s) => (
                    <div key={s.key} className="obs-tooltip-row">
                      <i className="obs-line-key" style={{ background: s.color }} />
                      <strong>{cfg.fmt(val(point, s.key))}</strong>
                      <span>{s.label}</span>
                    </div>
                  ))}
                  {cfg.headline === 'population' && (
                    <div className="obs-tooltip-row obs-tooltip-extra">
                      <span>
                        Gen. maks {point.maxGeneration} · rata-rata {point.avgGeneration.toFixed(1)}
                      </span>
                    </div>
                  )}
                </div>
              )}
            </>
          )
        )}
      </div>
    </section>
  )
}

function HistoryTable({
  history,
  series,
  spy,
  fmt,
}: {
  history: SimHistoryPoint[]
  series: Series[]
  spy: number
  fmt: (v: number) => string
}) {
  const rows = history.slice(-12).reverse()
  return (
    <table className="obs-table">
      <thead>
        <tr>
          <th>Tahun</th>
          {series.map((s) => (
            <th key={s.key}>{s.label}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map((h) => (
          <tr key={h.time}>
            <td>{formatClockShort(h.time, spy)}</td>
            {series.map((s) => (
              <td key={s.key}>{fmt(val(h, s.key))}</td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  )
}
