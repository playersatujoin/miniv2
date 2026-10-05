import { DEATH_LABELS, type DeathCause, type SimInfo } from '../../sim/protocol'
import { SEX_COLOR, nf, nf1 } from './format'

function Stat({ label, value, mark, title }: { label: string; value: string; mark?: string; title?: string }) {
  return (
    <div className="obs-stat" title={title}>
      <span className="obs-stat-label">
        {mark && <i className="obs-dot" style={{ background: mark }} />}
        {label}
      </span>
      <strong>{value}</strong>
    </div>
  )
}

const n = (v: number | undefined) => (v == null ? '—' : nf.format(v))

/** Per-year rate when the server reports one (easier to read than huge totals), else the total. */
function rate(label: string, perYear: number | undefined, total: number | undefined) {
  return perYear == null
    ? { label, value: n(total) }
    : { label: `${label}/thn`, value: nf1.format(perYear), title: `Total sepanjang masa: ${n(total)}` }
}

export function PopulationStats({ info }: { info: SimInfo }) {
  const causes = Object.entries(info.deathsByCause ?? {}) as [DeathCause, number][]

  return (
    <div className="obs-stats">
      <div className="obs-stat-grid">
        <Stat label="Populasi" value={`${n(info.population)} / ${n(info.capacity)}`} />
        <Stat label="♀ Perempuan" value={n(info.females)} mark={SEX_COLOR.female} />
        <Stat label="♂ Laki-laki" value={n(info.males)} mark={SEX_COLOR.male} />
        <Stat label="Generasi maks" value={n(info.maxGeneration)} />
        <Stat label="Gen. rata-rata" value={info.avgGeneration == null ? '—' : nf1.format(info.avgGeneration)} />
        <Stat label="⚗ Unsur" value={`${n(info.elementsDiscovered)} / ${n(info.elementsTotal ?? 118)}`} />
        <Stat label="🏠 Rumah" value={n(info.houses)} />
        <Stat {...rate('🤝 Kebaikan', info.kindnessPerYear, info.kindness)} />
        <Stat {...rate('🗡 Kejahatan', info.crimesPerYear, info.crimes)} />
        {info.killsPerYear != null && <Stat {...rate('☠ Pembunuhan', info.killsPerYear, info.kills)} />}
        {info.avgBrainSize != null && <Stat label="🧠 Otak rata-rata" value={`${nf1.format(info.avgBrainSize)} neuron`} />}
        {info.knowledgeLost != null && (
          <Stat label="📉 Ilmu hilang" value={n(info.knowledgeLost)} title="Berapa kali sebuah keahlian mati bersama pemegang terakhirnya" />
        )}
      </div>
      {info.crimesPerYear != null && <p className="small muted obs-rate-note">Laju "/thn" = rata-rata per tahun selama 50 tahun terakhir.</p>}
      <p className="obs-vitals small">
        <span>
          Kelahiran <strong>{n(info.births)}</strong>
        </span>
        <span>
          Kematian <strong>{n(info.deaths)}</strong>
        </span>
        {info.kills > 0 && (
          <span>
            Pembunuhan <strong>{n(info.kills)}</strong>
          </span>
        )}
      </p>
      {info.deaths > 0 && (
        <p className="obs-causes small muted">
          {causes
            .filter(([, count]) => count > 0)
            .map(([cause, count]) => (
              <span key={cause}>
                {DEATH_LABELS[cause] ?? cause} {nf.format(count)}
              </span>
            ))}
        </p>
      )}
    </div>
  )
}
