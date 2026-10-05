import { useState, type CSSProperties } from 'react'
import type { ElementInfo, Knowledge } from '../../sim/protocol'
import { ELEMENT_GROUPS, groupOf } from './categories'
import { formatAbundance, formatClockShort, nf, useSecondsPerYear } from './format'

type Props = { knowledge: Knowledge; onSelect: (id: number) => void }

const PHASE: Record<string, string> = { padat: 'Padat', cair: 'Cair', gas: 'Gas' }
const F_ROW = { lanthanide: 9, actinide: 10 }

/** Grid position from the atomic number alone (standard 18-column layout). */
function position(z: number): { row: number; col: number } {
  if (z >= 57 && z <= 71) return { row: F_ROW.lanthanide, col: 3 + (z - 57) }
  if (z >= 89 && z <= 103) return { row: F_ROW.actinide, col: 3 + (z - 89) }
  if (z === 1) return { row: 1, col: 1 }
  if (z === 2) return { row: 1, col: 18 }
  const short = (start: number, row: number) => {
    const i = z - start
    return { row, col: i < 2 ? i + 1 : i + 11 }
  }
  if (z <= 10) return short(3, 2)
  if (z <= 18) return short(11, 3)
  if (z <= 36) return { row: 4, col: z - 18 }
  if (z <= 54) return { row: 5, col: z - 36 }
  const long = (start: number, dBlock: number, row: number) =>
    z < start + 2 ? { row, col: z - start + 1 } : { row, col: z - dBlock + 4 }
  if (z <= 86) return long(55, 72, 6)
  return long(87, 104, 7)
}

function tierLabel(k: Knowledge, tier: number) {
  return k.tierNames?.[tier] ?? `Tier ${tier}`
}

function ariaLabel(e: ElementInfo, k: Knowledge) {
  const base = `${e.name} (${e.symbol}), nomor atom ${e.z}`
  return e.discovered
    ? `${base}, ditemukan${e.discoveredBy ? ` oleh ${e.discoveredBy.name}` : ''}`
    : `${base}, belum ditemukan, butuh ${tierLabel(k, e.tier)}`
}

export function PeriodicTable({ knowledge, onSelect }: Props) {
  const elements = knowledge.elements ?? []
  const [activeZ, setActiveZ] = useState<number | null>(null)

  const found = elements.filter((e) => e.discovered)
  const latest = found.reduce<ElementInfo | null>(
    (a, e) => (!a || (e.discoveredAt ?? 0) > (a.discoveredAt ?? 0) ? e : a),
    null,
  )
  const active = elements.find((e) => e.z === activeZ) ?? latest
  const total = elements.length || 118
  // Categories the colour scheme doesn't know fall back to grey; list that only if it occurs.
  const other = elements.map((e) => groupOf(e.category)).find((g) => !ELEMENT_GROUPS.includes(g))
  const legend = other ? [...ELEMENT_GROUPS, other] : ELEMENT_GROUPS

  if (elements.length === 0) {
    return <p className="small muted">Memuat tabel periodik…</p>
  }

  return (
    <section className="obs-section pt">
      <div className="pt-progress">
        <div className="obs-meter-head">
          <span>Unsur ditemukan</span>
          <span className="obs-meter-value">
            <strong>{nf.format(found.length)}</strong> / {nf.format(total)}
          </span>
        </div>
        <div className="obs-meter-track" style={{ background: 'color-mix(in srgb, var(--accent) 18%, transparent)' }}>
          <div className="obs-meter-fill" style={{ width: `${(found.length / total) * 100}%`, background: 'var(--accent)' }} />
        </div>
        <p className="small muted pt-now">
          {tierLabel(knowledge, knowledge.tier ?? 0)}: unsur hingga tier {knowledge.tier ?? 0} bisa dipisahkan.
        </p>
      </div>

      <div className="pt-grid" role="group" aria-label="Tabel periodik">
        {elements.map((e) => {
          const { row, col } = position(e.z)
          const g = groupOf(e.category)
          const style = { gridRow: row, gridColumn: col, '--c': g.color } as CSSProperties
          return (
            <button
              key={e.z}
              type="button"
              className={`pt-cell ${e.discovered ? 'is-found' : ''} ${active?.z === e.z ? 'is-active' : ''}`}
              style={style}
              aria-label={ariaLabel(e, knowledge)}
              aria-pressed={activeZ === e.z}
              onMouseEnter={() => setActiveZ(e.z)}
              onFocus={() => setActiveZ(e.z)}
              onClick={() => setActiveZ(e.z)}
            >
              <span className="pt-z">{e.z}</span>
              <span className="pt-sym">{e.symbol}</span>
            </button>
          )
        })}
        <span className="pt-ref" style={{ gridRow: 6, gridColumn: 3 }} aria-hidden>
          57–71
        </span>
        <span className="pt-ref" style={{ gridRow: 7, gridColumn: 3 }} aria-hidden>
          89–103
        </span>
      </div>

      {active ? <ElementCard e={active} k={knowledge} onSelect={onSelect} /> : (
        <p className="obs-hint small">Belum ada unsur yang ditemukan. Arahkan ke kotak untuk melihat detailnya.</p>
      )}

      <ul className="pt-legend">
        {legend.map((g) => (
          <li key={g.id}>
            <i className="pt-key" style={{ background: g.color }} />
            {g.label}
          </li>
        ))}
        <li>
          <i className="pt-key pt-key-unknown" />
          Belum ditemukan
        </li>
      </ul>
    </section>
  )
}

function ElementCard({ e, k, onSelect }: { e: ElementInfo; k: Knowledge; onSelect: (id: number) => void }) {
  const g = groupOf(e.category)
  const spy = useSecondsPerYear()
  return (
    <div className="pt-card" aria-live="polite">
      <div className={`pt-card-sym ${e.discovered ? 'is-found' : ''}`} style={{ '--c': g.color } as CSSProperties}>
        <span className="pt-z">{e.z}</span>
        <strong>{e.symbol}</strong>
      </div>
      <div className="pt-card-body">
        <h4>{e.name}</h4>
        <p className="small muted">
          {e.category || g.label} · {PHASE[e.phase] ?? e.phase}
          {e.group ? ` · golongan ${e.group}` : ''} · periode {e.period}
        </p>
        <dl className="obs-kv pt-kv">
          <dt>Kelimpahan</dt>
          <dd>{formatAbundance(e.abundance)}</dd>
          <dt>Butuh</dt>
          <dd>
            {tierLabel(k, e.tier)} (tier {e.tier})
          </dd>
          <dt>Status</dt>
          <dd>
            {e.discovered ? (
              <>
                ✓ Ditemukan
                {e.discoveredBy && (
                  <>
                    {' '}
                    oleh{' '}
                    <button type="button" className="obs-link" onClick={() => onSelect(e.discoveredBy!.id)}>
                      {e.discoveredBy.name}
                    </button>
                  </>
                )}
              </>
            ) : (
              <span className="muted">Belum ditemukan</span>
            )}
          </dd>
          {e.discovered && (
            <>
              <dt>Kapan</dt>
              <dd>
                {e.era ? `Era ${e.era} · ` : ''}
                {e.discoveredAt != null ? formatClockShort(e.discoveredAt, spy) : '—'}
              </dd>
              {e.source && (
                <>
                  <dt>Sumber</dt>
                  <dd>{e.source}</dd>
                </>
              )}
            </>
          )}
        </dl>
      </div>
    </div>
  )
}
