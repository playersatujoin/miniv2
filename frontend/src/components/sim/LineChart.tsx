import { useMemo, useState, type KeyboardEvent, type PointerEvent, type ReactNode } from 'react'
import { useElementWidth } from './useElementWidth'

/** One line: its colour is a validated palette slot, assigned in order. */
export type LineSeries<P> = { key: string; label: string; color: string; value: (p: P) => number }

type Props<P> = {
  title: string
  points: P[]
  /** Position along the x axis (time or year). */
  x: (p: P) => number
  /** Short x label for the axis ends and the table ("Thn 12"). */
  xLabel: (p: P) => string
  /** Tooltip heading ("Tahun 12"). */
  tooltipTitle: (p: P) => string
  series: LineSeries<P>[]
  fmt: (v: number) => string
  /** Lowest y-axis maximum, so small values don't fill the whole chart. */
  floor?: number
  /** Extra tooltip line under the series values. */
  extra?: (p: P) => ReactNode
  empty: string
  /** First column header of the table view. */
  tableHead?: string
}

const SURFACE = '#172234'
const GRID = '#24324a'
const AXIS = '#33476a'
const MUTED = '#8ea0bb'

const HEIGHT = 148
const M = { top: 12, right: 14, bottom: 22, left: 34 }

function niceCeil(v: number) {
  if (v <= 0) return 1
  const step = 10 ** Math.floor(Math.log10(v))
  for (const m of [1, 2, 2.5, 5, 10]) if (m * step >= v) return m * step
  return 10 * step
}

/**
 * Several series over time on one axis, in the look of the population chart:
 * a legend with the latest values (so identity never rests on colour alone),
 * a crosshair tooltip, arrow-key browsing and a table view.
 */
export function LineChart<P>({
  title,
  points,
  x,
  xLabel,
  tooltipTitle,
  series,
  fmt,
  floor = 1,
  extra,
  empty,
  tableHead = 'Tahun',
}: Props<P>) {
  const [wrapRef, width] = useElementWidth<HTMLDivElement>()
  const [active, setActive] = useState<number | null>(null)
  const [showTable, setShowTable] = useState(false)
  const ready = points.length >= 2 && width > 0

  const geo = useMemo(() => {
    if (!ready) return null
    const x0 = x(points[0])
    const x1 = x(points[points.length - 1])
    const peak = Math.max(...points.flatMap((p) => series.map((s) => s.value(p))))
    const yMax = niceCeil(Math.max(floor, peak))
    const plotW = width - M.left - M.right
    const plotH = HEIGHT - M.top - M.bottom
    const sx = (v: number) => M.left + ((v - x0) / Math.max(1e-9, x1 - x0)) * plotW
    const sy = (v: number) => M.top + plotH - (v / yMax) * plotH
    const paths = series.map((s) =>
      points.map((p, i) => `${i ? 'L' : 'M'}${sx(x(p)).toFixed(1)},${sy(s.value(p)).toFixed(1)}`).join(''),
    )
    return { yMax, plotW, plotH, sx, sy, paths }
  }, [ready, points, series, x, floor, width])

  const onPointerMove = (e: PointerEvent<SVGSVGElement>) => {
    if (!geo) return
    const px = e.clientX - e.currentTarget.getBoundingClientRect().left
    let best = 0
    let bestD = Infinity
    points.forEach((p, i) => {
      const d = Math.abs(geo.sx(x(p)) - px)
      if (d < bestD) {
        bestD = d
        best = i
      }
    })
    setActive(best)
  }

  const onKeyDown = (e: KeyboardEvent<SVGSVGElement>) => {
    if (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight') return
    e.preventDefault()
    const cur = active ?? points.length - 1
    setActive(Math.min(points.length - 1, Math.max(0, cur + (e.key === 'ArrowLeft' ? -1 : 1))))
  }

  const last = points[points.length - 1]
  const point = active !== null ? points[active] : null

  return (
    <div className="dm-block">
      <div className="obs-section-head">
        <h4 className="dm-subtitle">{title}</h4>
        {points.length >= 2 && (
          <button type="button" className="obs-mini" onClick={() => setShowTable((v) => !v)}>
            {showTable ? 'Grafik' : 'Tabel'}
          </button>
        )}
      </div>

      {/* The legend carries the latest values, so every line is named without crowding the plot. */}
      {series.length > 1 && last && (
        <ul className="obs-legend">
          {series.map((s) => (
            <li key={s.key}>
              <i className="obs-line-key" style={{ background: s.color }} />
              {s.label} <strong className="eco-legend-value">{fmt(s.value(last))}</strong>
            </li>
          ))}
        </ul>
      )}

      <div ref={wrapRef} className="obs-chart-wrap">
        {!ready ? (
          <p className="obs-empty small muted">{empty}</p>
        ) : showTable ? (
          <table className="obs-table">
            <thead>
              <tr>
                <th>{tableHead}</th>
                {series.map((s) => (
                  <th key={s.key}>{s.label}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {points
                .slice(-12)
                .reverse()
                .map((p) => (
                  <tr key={x(p)}>
                    <td>{xLabel(p)}</td>
                    {series.map((s) => (
                      <td key={s.key}>{fmt(s.value(p))}</td>
                    ))}
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
                aria-label={`${title}: ${series.map((s) => `${s.label} ${fmt(s.value(last))}`).join(', ')}. Gunakan panah kiri/kanan untuk menelusuri.`}
                onPointerMove={onPointerMove}
                onPointerLeave={() => setActive(null)}
                onFocus={() => setActive(points.length - 1)}
                onBlur={() => setActive(null)}
                onKeyDown={onKeyDown}
              >
                {[0, geo.yMax / 2, geo.yMax].map((v) => (
                  <g key={v}>
                    <line
                      x1={M.left}
                      x2={M.left + geo.plotW}
                      y1={geo.sy(v)}
                      y2={geo.sy(v)}
                      stroke={v === 0 ? AXIS : GRID}
                      strokeWidth={1}
                      shapeRendering="crispEdges"
                    />
                    <text x={M.left - 5} y={geo.sy(v) + 3} textAnchor="end" className="obs-axis-text">
                      {fmt(v)}
                    </text>
                  </g>
                ))}

                <text x={M.left} y={HEIGHT - 6} className="obs-axis-text">
                  {xLabel(points[0])}
                </text>
                <text x={M.left + geo.plotW} y={HEIGHT - 6} textAnchor="end" className="obs-axis-text">
                  {xLabel(last)}
                </text>

                {series.map((s, i) => (
                  <path
                    key={s.key}
                    d={geo.paths[i]}
                    fill="none"
                    stroke={s.color}
                    strokeWidth={2}
                    strokeLinejoin="round"
                    strokeLinecap="round"
                  />
                ))}

                {(point ?? last) && (
                  <g>
                    {point && (
                      <line
                        x1={geo.sx(x(point))}
                        x2={geo.sx(x(point))}
                        y1={M.top}
                        y2={M.top + geo.plotH}
                        stroke={MUTED}
                        strokeOpacity={0.6}
                        strokeWidth={1}
                      />
                    )}
                    {series.map((s) => (
                      <circle
                        key={s.key}
                        cx={geo.sx(x(point ?? last))}
                        cy={geo.sy(s.value(point ?? last))}
                        r={4}
                        fill={s.color}
                        stroke={SURFACE}
                        strokeWidth={2}
                      />
                    ))}
                  </g>
                )}
              </svg>

              {point && (
                <div
                  className="obs-tooltip"
                  style={{ left: Math.min(width - 168, Math.max(0, geo.sx(x(point)) + 10)), top: 4 }}
                >
                  <div className="obs-tooltip-title">{tooltipTitle(point)}</div>
                  {series.map((s) => (
                    <div key={s.key} className="obs-tooltip-row">
                      <i className="obs-line-key" style={{ background: s.color }} />
                      <strong>{fmt(s.value(point))}</strong>
                      <span>{s.label}</span>
                    </div>
                  ))}
                  {extra && <div className="obs-tooltip-row obs-tooltip-extra">{extra(point)}</div>}
                </div>
              )}
            </>
          )
        )}
      </div>
    </div>
  )
}
