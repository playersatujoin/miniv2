import { useMemo, useState, type KeyboardEvent, type PointerEvent } from 'react'
import { ENSO_LABELS, SPECIES, ensoOf, type EcoYear, type Enso } from '../../sim/protocol'
import { LineChart, type LineSeries } from './LineChart'
import { nf, nf1 } from './format'
import { useElementWidth } from './useElementWidth'

// ENSO is polarity, so it takes the diverging pair: blue (La Niña, wetter) ↔
// red (El Niño, drier) with a neutral grey for ordinary years. Validated
// all-pairs on the panel surface #172234: worst CVD ΔE 11.7, normal 21.8; the
// grey is below 3:1 on purpose (a neutral recedes), relieved by the tooltip
// and the table view. Starvation is a separate panel with one series in the
// violet slot, so it never reads as an ENSO colour.
export const ENSO_COLOR: Record<Enso, string> = { la_nina: '#3987e5', netral: '#5f5e5a', el_nino: '#e66767' }
const ENSO_ORDER: Enso[] = ['la_nina', 'netral', 'el_nino']
const STARVED = '#9085e9'

const GRID = '#24324a'
const AXIS = '#33476a'
const AVERAGE = '#8ea0bb'
const TEXT = '#e6edf7'

const RAIN_H = 96
const GAP = 26 // room for the second panel's title
const STARVE_H = 56
const M = { top: 12, right: 8, bottom: 20, left: 34 }
const HEIGHT = M.top + RAIN_H + GAP + STARVE_H + M.bottom

const WINDOWS = [
  { years: 50, label: '50 th' },
  { years: 150, label: '150 th' },
  { years: 0, label: 'Semua' },
]

function niceCeil(v: number) {
  if (v <= 0) return 1
  const step = 10 ** Math.floor(Math.log10(v))
  for (const m of [1, 2, 2.5, 5, 10]) if (m * step >= v) return m * step
  return 10 * step
}

/** A vertical bar from the baseline up to `top`, rounded only at its data end. */
function barPath(x: number, w: number, base: number, top: number) {
  const h = base - top
  if (h < 0.5 || w <= 0) return ''
  const r = Math.min(3, w / 2, h)
  return `M${x},${base}L${x},${top + r}Q${x},${top} ${x + r},${top}L${x + w - r},${top}Q${x + w},${top} ${x + w},${top + r}L${x + w},${base}Z`
}

const huntedTotal = (y: EcoYear) => (y.hunted ?? []).reduce((a, b) => a + b, 0)

/** Species ids (as logged per year) to their names. */
const speciesNames = (ids: string[] | undefined) =>
  (ids ?? []).map((id) => SPECIES.find((s) => s.id === id)?.name ?? id).join(', ')

// Three series: the first three categorical slots, which validate all-pairs.
const FOOD_SERIES: LineSeries<EcoYear>[] = [
  { key: 'harvest', label: 'Panen', color: '#3987e5', value: (y) => y.harvest ?? 0 },
  { key: 'fished', label: 'Ikan', color: '#d95926', value: (y) => y.fished ?? 0 },
  { key: 'hunted', label: 'Hewan diburu', color: '#199e70', value: huntedTotal },
]

/** Closed years: rainfall by ENSO above starvation deaths, then where food came from. */
export function EcologyYears({ years }: { years: EcoYear[] }) {
  const [windowYears, setWindowYears] = useState(50)
  const shown = useMemo(() => (windowYears ? years.slice(-windowYears) : years), [years, windowYears])

  return (
    <section className="obs-section" aria-labelledby="eco-years-title">
      <div className="obs-section-head">
        <h3 className="obs-title" id="eco-years-title">
          Tahun demi tahun
        </h3>
        <div className="segmented eco-window" role="group" aria-label="Rentang tahun">
          {WINDOWS.map((w) => (
            <button
              key={w.years}
              type="button"
              className={windowYears === w.years ? 'active' : ''}
              aria-pressed={windowYears === w.years}
              onClick={() => setWindowYears(w.years)}
            >
              {w.label}
            </button>
          ))}
        </div>
      </div>
      {years.length === 0 ? (
        <p className="obs-empty small muted">Data tahunan muncul setelah tahun pertama selesai.</p>
      ) : (
        <>
          <ClimateBars years={shown} />
          <LineChart
            title="Pangan per tahun: panen, ikan dan buruan"
            points={shown}
            x={(y) => y.year}
            xLabel={(y) => `Thn ${nf.format(y.year)}`}
            tooltipTitle={(y) => `Tahun ${nf.format(y.year)}`}
            series={FOOD_SERIES}
            fmt={(v) => nf.format(Math.round(v))}
            floor={5}
            extra={(y) => (
              <span>
                Gagal panen {nf.format(y.cropLoss ?? 0)} petak · busuk {nf.format(y.rotten ?? 0)} satuan
              </span>
            )}
            empty="Grafik muncul setelah dua tahun simulasi."
          />
          <p className="small muted eco-note">
            Panen dan ikan dalam satuan pangan; hewan diburu dalam ekor (seekor rusa ≈ 6 satuan daging).
          </p>
        </>
      )}
    </section>
  )
}

/** Rain per year coloured by ENSO, with starvation deaths in a second panel on the same years. */
function ClimateBars({ years }: { years: EcoYear[] }) {
  const [wrapRef, width] = useElementWidth<HTMLDivElement>()
  const [active, setActive] = useState<number | null>(null)
  const [showTable, setShowTable] = useState(false)
  const n = years.length

  const geo = useMemo(() => {
    if (!width || !n) return null
    const plotW = width - M.left - M.right
    const bw = plotW / n
    const gap = bw >= 6 ? 2 : bw >= 3 ? 1 : 0
    const rainMax = Math.max(2, Math.ceil(Math.max(...years.map((y) => y.rain)) * 2) / 2)
    const starveMax = niceCeil(Math.max(4, ...years.map((y) => y.starved ?? 0)))
    const rainTop = M.top
    const rainBase = M.top + RAIN_H
    const starveTop = rainBase + GAP
    const starveBase = starveTop + STARVE_H
    const ry = (v: number) => rainBase - (Math.min(v, rainMax) / rainMax) * RAIN_H
    const sy = (v: number) => starveBase - (Math.min(v, starveMax) / starveMax) * STARVE_H
    return { plotW, bw, gap, rainMax, starveMax, rainTop, rainBase, starveTop, starveBase, ry, sy }
  }, [width, n, years])

  const onPointerMove = (e: PointerEvent<SVGSVGElement>) => {
    if (!geo) return
    const px = e.clientX - e.currentTarget.getBoundingClientRect().left - M.left
    const i = Math.floor(px / geo.bw)
    setActive(i >= 0 && i < n ? i : null)
  }

  const onKeyDown = (e: KeyboardEvent<SVGSVGElement>) => {
    if (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight') return
    e.preventDefault()
    const cur = active ?? n - 1
    setActive(Math.min(n - 1, Math.max(0, cur + (e.key === 'ArrowLeft' ? -1 : 1))))
  }

  const counts = { el_nino: 0, la_nina: 0, netral: 0 } as Record<Enso, number>
  for (const y of years) counts[ensoOf(y.enso)]++
  const starvedTotal = years.reduce((s, y) => s + (y.starved ?? 0), 0)
  const floods = years.filter((y) => y.flood).length
  const point = active !== null ? years[active] : null
  const summary =
    `Curah hujan dan kematian karena kelaparan, ${nf.format(n)} tahun terakhir: ` +
    `${counts.el_nino} tahun El Niño, ${counts.la_nina} tahun La Niña, ${floods} banjir; ` +
    `${nf.format(starvedTotal)} orang mati kelaparan. Gunakan panah kiri/kanan untuk menelusuri.`

  return (
    <div className="dm-block">
      <div className="obs-section-head">
        <h4 className="dm-subtitle">Curah hujan per tahun &amp; kelaparan</h4>
        <button type="button" className="obs-mini" onClick={() => setShowTable((v) => !v)}>
          {showTable ? 'Grafik' : 'Tabel'}
        </button>
      </div>
      <ul className="obs-legend">
        {ENSO_ORDER.map((e) => (
          <li key={e}>
            <i className="eco-swatch" style={{ background: ENSO_COLOR[e] }} />
            {ENSO_LABELS[e]}
          </li>
        ))}
        <li>
          <span className="eco-flood-key" aria-hidden="true">
            ≈
          </span>
          banjir
        </li>
        <li>
          <i className="eco-dash-key" aria-hidden="true" />
          ×1 = rata-rata
        </li>
      </ul>

      <div ref={wrapRef} className="obs-chart-wrap">
        {showTable ? (
          <table className="obs-table">
            <thead>
              <tr>
                <th>Tahun</th>
                <th>ENSO</th>
                <th>Hujan</th>
                <th>Panen</th>
                <th>Kelaparan</th>
              </tr>
            </thead>
            <tbody>
              {years
                .slice(-12)
                .reverse()
                .map((y) => (
                  <tr key={y.year}>
                    <td>{nf.format(y.year)}</td>
                    <td>
                      {ENSO_LABELS[ensoOf(y.enso)]}
                      {y.flood ? ' · banjir' : ''}
                    </td>
                    <td>×{nf1.format(y.rain)}</td>
                    <td>{nf.format(y.harvest ?? 0)}</td>
                    <td>{nf.format(y.starved ?? 0)}</td>
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
                aria-label={summary}
                onPointerMove={onPointerMove}
                onPointerLeave={() => setActive(null)}
                onFocus={() => setActive(n - 1)}
                onBlur={() => setActive(null)}
                onKeyDown={onKeyDown}
              >
                {active !== null && (
                  <rect
                    x={M.left + active * geo.bw}
                    width={Math.max(1, geo.bw)}
                    y={geo.rainTop}
                    height={geo.starveBase - geo.rainTop}
                    fill="#ffffff"
                    fillOpacity={0.06}
                  />
                )}

                {/* Rain panel: 0, the yearly average (1) and the top. */}
                {[0, 1, geo.rainMax].map((v) => (
                  <g key={`r${v}`}>
                    <line
                      x1={M.left}
                      x2={M.left + geo.plotW}
                      y1={geo.ry(v)}
                      y2={geo.ry(v)}
                      stroke={v === 0 ? AXIS : v === 1 ? AVERAGE : GRID}
                      strokeOpacity={v === 1 ? 0.6 : 1}
                      strokeDasharray={v === 1 ? '3 3' : undefined}
                      shapeRendering="crispEdges"
                    />
                    <text x={M.left - 5} y={geo.ry(v) + 3} textAnchor="end" className="obs-axis-text">
                      ×{nf1.format(v)}
                    </text>
                  </g>
                ))}
                {years.map((y, i) => {
                  const x = M.left + i * geo.bw + geo.gap / 2
                  const w = geo.bw - geo.gap
                  const top = geo.ry(y.rain)
                  return (
                    <g key={y.year}>
                      <path d={barPath(x, w, geo.rainBase, top)} fill={ENSO_COLOR[ensoOf(y.enso)]} />
                      {y.flood && geo.bw >= 4 && (
                        <text x={x + w / 2} y={top - 3} textAnchor="middle" className="eco-flood-mark" fill={TEXT}>
                          ≈
                        </text>
                      )}
                    </g>
                  )
                })}

                {/* Starvation panel, same years. */}
                <text x={M.left} y={geo.starveTop - 8} className="obs-axis-text eco-panel-title">
                  Mati kelaparan (orang)
                </text>
                {[0, geo.starveMax].map((v) => (
                  <g key={`s${v}`}>
                    <line
                      x1={M.left}
                      x2={M.left + geo.plotW}
                      y1={geo.sy(v)}
                      y2={geo.sy(v)}
                      stroke={v === 0 ? AXIS : GRID}
                      shapeRendering="crispEdges"
                    />
                    <text x={M.left - 5} y={geo.sy(v) + 3} textAnchor="end" className="obs-axis-text">
                      {nf.format(v)}
                    </text>
                  </g>
                ))}
                {years.map((y, i) => (
                  <path
                    key={y.year}
                    d={barPath(M.left + i * geo.bw + geo.gap / 2, geo.bw - geo.gap, geo.starveBase, geo.sy(y.starved ?? 0))}
                    fill={STARVED}
                  />
                ))}

                <text x={M.left} y={HEIGHT - 5} className="obs-axis-text">
                  Thn {nf.format(years[0].year)}
                </text>
                <text x={M.left + geo.plotW} y={HEIGHT - 5} textAnchor="end" className="obs-axis-text">
                  Thn {nf.format(years[n - 1].year)}
                </text>
              </svg>

              {point && (
                <div
                  className="obs-tooltip"
                  style={{ left: Math.min(width - 178, Math.max(0, M.left + (active! + 1) * geo.bw + 8)), top: 4 }}
                >
                  <div className="obs-tooltip-title">
                    Tahun {nf.format(point.year)} · {ENSO_LABELS[ensoOf(point.enso)]}
                    {point.flood ? ' · banjir' : ''}
                  </div>
                  <div className="obs-tooltip-row">
                    <i className="obs-line-key" style={{ background: ENSO_COLOR[ensoOf(point.enso)] }} />
                    <strong>×{nf1.format(point.rain)}</strong>
                    <span>curah hujan</span>
                  </div>
                  <div className="obs-tooltip-row">
                    <i className="obs-line-key" style={{ background: STARVED }} />
                    <strong>{nf.format(point.starved ?? 0)}</strong>
                    <span>mati kelaparan</span>
                  </div>
                  <div className="obs-tooltip-row obs-tooltip-extra">
                    <span>
                      Panen {nf.format(point.harvest ?? 0)} · layu {nf.format(point.withered ?? 0)} petak · penduduk{' '}
                      {nf.format(point.population ?? 0)}
                    </span>
                  </div>
                  {!!(point.extinct?.length || point.arrived?.length) && (
                    <div className="obs-tooltip-row obs-tooltip-extra">
                      <span>
                        {[
                          point.extinct?.length ? `Punah: ${speciesNames(point.extinct)}` : '',
                          point.arrived?.length ? `Datang dari seberang: ${speciesNames(point.arrived)}` : '',
                        ]
                          .filter(Boolean)
                          .join(' · ')}
                      </span>
                    </div>
                  )}
                </div>
              )}
            </>
          )
        )}
      </div>
      <p className="small muted eco-note">
        Batang = curah hujan rata-rata setahun (×1 = tahun biasa). El Niño mengeringkan kemarau, La Niña membawa hujan
        dan banjir.
      </p>
    </div>
  )
}
