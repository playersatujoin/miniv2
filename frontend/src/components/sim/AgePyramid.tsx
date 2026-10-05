import { useState, type KeyboardEvent, type PointerEvent } from 'react'
import type { AgeBand } from '../../sim/protocol'
import { SEX_COLOR, nf } from './format'
import { useElementWidth } from './useElementWidth'

const ROW = 12
const GAP = 2
const LABEL_W = 44
const TOP = 20
const BOTTOM = 16
const GRID = '#24324a'
const AXIS = '#33476a'

function niceCeil(v: number) {
  if (v <= 0) return 1
  const step = 10 ** Math.floor(Math.log10(v))
  for (const m of [1, 2, 2.5, 5, 10]) if (m * step >= v) return m * step
  return 10 * step
}

/** Horizontal bar from the anchored end x0 to the data end x1, rounded only at the data end. */
function barPath(x0: number, x1: number, y: number, h: number) {
  const len = Math.abs(x1 - x0)
  if (len < 0.5) return ''
  const dir = x1 >= x0 ? 1 : -1
  const r = Math.min(4, h / 2, len)
  return `M${x0},${y}L${x1 - dir * r},${y}Q${x1},${y} ${x1},${y + r}L${x1},${y + h - r}Q${x1},${y + h} ${x1 - dir * r},${y + h}L${x0},${y + h}Z`
}

/** Living population by age band, women on the left and men on the right. */
export function AgePyramid({ bands }: { bands: AgeBand[] }) {
  const [wrapRef, width] = useElementWidth<HTMLDivElement>()
  const [active, setActive] = useState<number | null>(null)
  const [showTable, setShowTable] = useState(false)

  // Oldest at the top, like a printed pyramid.
  const rows = [...bands].reverse()
  const females = bands.reduce((s, b) => s + b.female, 0)
  const males = bands.reduce((s, b) => s + b.male, 0)
  const max = niceCeil(Math.max(1, ...bands.flatMap((b) => [b.female, b.male])))
  const height = TOP + rows.length * (ROW + GAP) - GAP + BOTTOM
  const half = Math.max(0, (width - LABEL_W) / 2)
  const left = half // female bars grow leftwards from here
  const right = half + LABEL_W // male bars grow rightwards from here
  const len = (n: number) => (n / max) * Math.max(0, half - 2)
  const rowY = (i: number) => TOP + i * (ROW + GAP)

  const onPointerMove = (e: PointerEvent<SVGSVGElement>) => {
    const rect = e.currentTarget.getBoundingClientRect()
    const i = Math.floor((e.clientY - rect.top - TOP + GAP / 2) / (ROW + GAP))
    setActive(i >= 0 && i < rows.length ? i : null)
  }

  const onKeyDown = (e: KeyboardEvent<SVGSVGElement>) => {
    if (e.key !== 'ArrowUp' && e.key !== 'ArrowDown') return
    e.preventDefault()
    const cur = active ?? rows.length - 1
    setActive(Math.min(rows.length - 1, Math.max(0, cur + (e.key === 'ArrowUp' ? -1 : 1))))
  }

  const row = active !== null ? rows[active] : null
  const summary = `Piramida umur: ${nf.format(females)} perempuan, ${nf.format(males)} laki-laki. Gunakan panah atas/bawah untuk menelusuri kelompok umur.`

  return (
    <div className="dm-block">
      <div className="obs-section-head">
        <h4 className="dm-subtitle">Piramida umur</h4>
        {bands.length > 0 && (
          <button type="button" className="obs-mini" onClick={() => setShowTable((v) => !v)}>
            {showTable ? 'Grafik' : 'Tabel'}
          </button>
        )}
      </div>

      {/* Two series: the legend names them, and the column headers label them directly. */}
      <ul className="obs-legend">
        <li>
          <i className="obs-line-key" style={{ background: SEX_COLOR.female }} />♀ Perempuan {nf.format(females)}
        </li>
        <li>
          <i className="obs-line-key" style={{ background: SEX_COLOR.male }} />♂ Laki-laki {nf.format(males)}
        </li>
      </ul>

      <div ref={wrapRef} className="obs-chart-wrap">
        {bands.length === 0 ? (
          <p className="obs-empty small muted">Belum ada data umur.</p>
        ) : showTable ? (
          <table className="obs-table">
            <thead>
              <tr>
                <th>Umur</th>
                <th>♀</th>
                <th>♂</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((b) => (
                <tr key={b.label}>
                  <td>{b.label}</td>
                  <td>{nf.format(b.female)}</td>
                  <td>{nf.format(b.male)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          width > 0 && (
            <>
              <svg
                width={width}
                height={height}
                className="obs-chart-svg"
                tabIndex={0}
                role="img"
                aria-label={summary}
                onPointerMove={onPointerMove}
                onPointerLeave={() => setActive(null)}
                onFocus={() => setActive(rows.length - 1)}
                onBlur={() => setActive(null)}
                onKeyDown={onKeyDown}
              >
                <text x={2} y={12} className="obs-axis-text">
                  ♀ Perempuan
                </text>
                <text x={width - 2} y={12} textAnchor="end" className="obs-axis-text">
                  Laki-laki ♂
                </text>

                {/* Recessive guides at half and full scale on both sides. */}
                {[0.5, 1].map((f) => (
                  <g key={f}>
                    <line x1={left - len(max * f)} x2={left - len(max * f)} y1={TOP - 2} y2={height - BOTTOM + 2} stroke={GRID} />
                    <line x1={right + len(max * f)} x2={right + len(max * f)} y1={TOP - 2} y2={height - BOTTOM + 2} stroke={GRID} />
                  </g>
                ))}
                <line x1={left} x2={left} y1={TOP - 2} y2={height - BOTTOM + 2} stroke={AXIS} />
                <line x1={right} x2={right} y1={TOP - 2} y2={height - BOTTOM + 2} stroke={AXIS} />

                {rows.map((b, i) => {
                  const y = rowY(i)
                  const on = active === i
                  return (
                    <g key={b.label}>
                      {on && (
                        <rect x={0} y={y - 1} width={width} height={ROW + 2} fill="#ffffff" fillOpacity={0.05} />
                      )}
                      <path d={barPath(left, left - len(b.female), y, ROW)} fill={SEX_COLOR.female} />
                      <path d={barPath(right, right + len(b.male), y, ROW)} fill={SEX_COLOR.male} />
                      <text x={left + LABEL_W / 2} y={y + ROW - 2} textAnchor="middle" className="obs-axis-text">
                        {b.label}
                      </text>
                    </g>
                  )
                })}

                <text x={left - len(max)} y={height - 3} textAnchor="start" className="obs-axis-text">
                  {nf.format(max)}
                </text>
                <text x={right + len(max)} y={height - 3} textAnchor="end" className="obs-axis-text">
                  {nf.format(max)}
                </text>
                <text x={left + LABEL_W / 2} y={height - 3} textAnchor="middle" className="obs-axis-text">
                  0
                </text>
              </svg>

              {row && (
                <div
                  className="obs-tooltip"
                  style={{ left: Math.max(0, Math.min(width - 150, right - 40)), top: Math.max(0, rowY(active!) - 34) }}
                >
                  <div className="obs-tooltip-title">{row.label} tahun</div>
                  <div className="obs-tooltip-row">
                    <i className="obs-line-key" style={{ background: SEX_COLOR.female }} />
                    <strong>{nf.format(row.female)}</strong>
                    <span>♀ Perempuan</span>
                  </div>
                  <div className="obs-tooltip-row">
                    <i className="obs-line-key" style={{ background: SEX_COLOR.male }} />
                    <strong>{nf.format(row.male)}</strong>
                    <span>♂ Laki-laki</span>
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
