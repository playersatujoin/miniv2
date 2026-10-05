import { useMemo, useState, type KeyboardEvent, type PointerEvent } from 'react'
import type { DemographyMetrics, DemographyPoint, MetricRef } from '../../sim/protocol'
import { fmtGini, nf, nf1, pct } from './format'
import { useElementWidth } from './useElementWidth'

type HKey = 'lifeExpectancy' | 'tfr' | 'survivalTo15' | 'gini' | 'homicideRate'

const OPTIONS: { key: HKey; tab: string; title: string; fmt: (v: number) => string }[] = [
  { key: 'lifeExpectancy', tab: 'Harapan hidup', title: 'Harapan hidup saat lahir (tahun)', fmt: (v) => nf1.format(v) },
  { key: 'tfr', tab: 'TFR', title: 'Anak per perempuan (TFR)', fmt: (v) => nf1.format(v) },
  { key: 'survivalTo15', tab: 'Hidup 15 th', title: 'Bayi bertahan sampai 15 tahun', fmt: pct },
  { key: 'gini', tab: 'Gini', title: 'Ketimpangan kekayaan (Gini)', fmt: fmtGini },
  { key: 'homicideRate', tab: 'Pembunuhan', title: 'Pembunuhan per 100.000 per tahun', fmt: (v) => nf.format(Math.round(v)) },
]

// One series at a time, so one validated slot; the reference band is a neutral wash.
const LINE = '#c98500'
const SURFACE = '#172234'
const GRID = '#24324a'
const AXIS = '#33476a'
const BAND = '#8ea0bb'
const TEXT = '#e6edf7'

const HEIGHT = 140
const M = { top: 12, right: 36, bottom: 22, left: 34 }

function niceCeil(v: number) {
  if (v <= 0) return 1
  const step = 10 ** Math.floor(Math.log10(v))
  for (const m of [1, 2, 2.5, 5, 10]) if (m * step >= v) return m * step
  return 10 * step
}

type Props = {
  history: DemographyPoint[]
  reference: Partial<Record<keyof DemographyMetrics, MetricRef>>
}

/** How one indicator moved over the years, against its pre-modern range. */
export function DemographyHistory({ history, reference }: Props) {
  const [wrapRef, width] = useElementWidth<HTMLDivElement>()
  const [key, setKey] = useState<HKey>('lifeExpectancy')
  const [active, setActive] = useState<number | null>(null)
  const [showTable, setShowTable] = useState(false)
  const opt = OPTIONS.find((o) => o.key === key)!
  const ref = reference[key]
  const points = history.filter((h) => h[key] != null)
  const ready = points.length >= 2 && width > 0

  const geo = useMemo(() => {
    if (!ready) return null
    const y0 = history[0].year
    const y1 = history[history.length - 1].year
    const peak = Math.max(...points.map((h) => h[key] as number), ref?.high ?? 0)
    const yMax = niceCeil(peak)
    const plotW = width - M.left - M.right
    const plotH = HEIGHT - M.top - M.bottom
    const x = (yr: number) => M.left + ((yr - y0) / Math.max(1e-9, y1 - y0)) * plotW
    const y = (v: number) => M.top + plotH - (Math.min(v, yMax) / yMax) * plotH
    // Break the line where there wasn't enough data, instead of bridging the gap.
    let d = ''
    let pen = false
    for (const h of history) {
      const v = h[key]
      if (v == null) {
        pen = false
        continue
      }
      d += `${pen ? 'L' : 'M'}${x(h.year).toFixed(1)},${y(v).toFixed(1)}`
      pen = true
    }
    return { y0, y1, yMax, plotW, plotH, x, y, d }
  }, [ready, history, points, key, ref, width])

  const onPointerMove = (e: PointerEvent<SVGSVGElement>) => {
    if (!geo) return
    const px = e.clientX - e.currentTarget.getBoundingClientRect().left
    let best = 0
    let bestD = Infinity
    history.forEach((h, i) => {
      const dist = Math.abs(geo.x(h.year) - px)
      if (dist < bestD) {
        bestD = dist
        best = i
      }
    })
    setActive(best)
  }

  const onKeyDown = (e: KeyboardEvent<SVGSVGElement>) => {
    if (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight') return
    e.preventDefault()
    const cur = active ?? history.length - 1
    setActive(Math.min(history.length - 1, Math.max(0, cur + (e.key === 'ArrowLeft' ? -1 : 1))))
  }

  const point = active !== null ? history[active] : null
  const lastPoint = points[points.length - 1]

  return (
    <div className="dm-block">
      <div className="obs-section-head">
        <h4 className="dm-subtitle">{opt.title}</h4>
        {points.length >= 2 && (
          <button type="button" className="obs-mini" onClick={() => setShowTable((v) => !v)}>
            {showTable ? 'Grafik' : 'Tabel'}
          </button>
        )}
      </div>
      <div className="segmented obs-metric dm-metric" role="group" aria-label="Pilih indikator">
        {OPTIONS.map((o) => (
          <button
            key={o.key}
            type="button"
            className={o.key === key ? 'active' : ''}
            aria-pressed={o.key === key}
            onClick={() => {
              setKey(o.key)
              setActive(null)
            }}
          >
            {o.tab}
          </button>
        ))}
      </div>

      <div ref={wrapRef} className="obs-chart-wrap">
        {!ready ? (
          <p className="obs-empty small muted">Grafik muncul setelah beberapa puluh tahun simulasi.</p>
        ) : showTable ? (
          <table className="obs-table">
            <thead>
              <tr>
                <th>Tahun</th>
                <th>{opt.tab}</th>
                <th>Populasi</th>
              </tr>
            </thead>
            <tbody>
              {history
                .slice(-12)
                .reverse()
                .map((h) => (
                  <tr key={h.year}>
                    <td>{nf.format(h.year)}</td>
                    <td>{h[key] == null ? '—' : opt.fmt(h[key] as number)}</td>
                    <td>{nf.format(h.population)}</td>
                  </tr>
                ))}
            </tbody>
          </table>
        ) : (
          geo && (
            <>
              <svg
                width={width}
                height={HEIGHT}
                className="obs-chart-svg"
                tabIndex={0}
                role="img"
                aria-label={`${opt.title}${lastPoint ? `: terakhir ${opt.fmt(lastPoint[key] as number)}` : ''}${ref ? `, acuan pra-modern ${opt.fmt(ref.low)}–${opt.fmt(ref.high)}` : ''}. Gunakan panah kiri/kanan untuk menelusuri.`}
                onPointerMove={onPointerMove}
                onPointerLeave={() => setActive(null)}
                onFocus={() => setActive(history.length - 1)}
                onBlur={() => setActive(null)}
                onKeyDown={onKeyDown}
              >
                {ref && (
                  <g>
                    <rect
                      x={M.left}
                      width={geo.plotW}
                      y={geo.y(ref.high)}
                      height={Math.max(1, geo.y(ref.low) - geo.y(ref.high))}
                      fill={BAND}
                      fillOpacity={0.14}
                    />
                    <text x={M.left + 3} y={geo.y(ref.high) + 10} className="obs-axis-text">
                      acuan pra-modern
                    </text>
                  </g>
                )}

                {[0, geo.yMax / 2, geo.yMax].map((v) => (
                  <g key={v}>
                    <line
                      x1={M.left}
                      x2={M.left + geo.plotW}
                      y1={geo.y(v)}
                      y2={geo.y(v)}
                      stroke={v === 0 ? AXIS : GRID}
                      shapeRendering="crispEdges"
                    />
                    <text x={M.left - 5} y={geo.y(v) + 3} textAnchor="end" className="obs-axis-text">
                      {opt.fmt(v)}
                    </text>
                  </g>
                ))}

                <text x={M.left} y={HEIGHT - 6} className="obs-axis-text">
                  Thn {nf.format(geo.y0)}
                </text>
                <text x={M.left + geo.plotW} y={HEIGHT - 6} textAnchor="end" className="obs-axis-text">
                  Thn {nf.format(geo.y1)}
                </text>

                <path d={geo.d} fill="none" stroke={LINE} strokeWidth={2} strokeLinejoin="round" strokeLinecap="round" />

                {point && point[key] != null ? (
                  <g>
                    <line
                      x1={geo.x(point.year)}
                      x2={geo.x(point.year)}
                      y1={M.top}
                      y2={M.top + geo.plotH}
                      stroke={BAND}
                      strokeOpacity={0.6}
                    />
                    <circle cx={geo.x(point.year)} cy={geo.y(point[key] as number)} r={4} fill={LINE} stroke={SURFACE} strokeWidth={2} />
                  </g>
                ) : (
                  lastPoint &&
                  !point && (
                    <g>
                      <circle
                        cx={geo.x(lastPoint.year)}
                        cy={geo.y(lastPoint[key] as number)}
                        r={4}
                        fill={LINE}
                        stroke={SURFACE}
                        strokeWidth={2}
                      />
                      <text
                        x={geo.x(lastPoint.year) + 7}
                        y={geo.y(lastPoint[key] as number) + 4}
                        fill={TEXT}
                        className="obs-end-label"
                      >
                        {opt.fmt(lastPoint[key] as number)}
                      </text>
                    </g>
                  )
                )}
              </svg>

              {point && (
                <div className="obs-tooltip" style={{ left: Math.min(width - 160, Math.max(0, geo.x(point.year) + 10)), top: 4 }}>
                  <div className="obs-tooltip-title">Tahun {nf.format(point.year)}</div>
                  <div className="obs-tooltip-row">
                    <i className="obs-line-key" style={{ background: LINE }} />
                    <strong>{point[key] == null ? '—' : opt.fmt(point[key] as number)}</strong>
                    <span>{opt.tab}</span>
                  </div>
                  <div className="obs-tooltip-row obs-tooltip-extra">
                    <span>Populasi {nf.format(point.population)}</span>
                  </div>
                </div>
              )}
            </>
          )
        )}
      </div>
    </div>
  )
}
