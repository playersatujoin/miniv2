import { DEATH_LABELS, type DeathCause, type SimInfo } from '../../sim/protocol'
import { SEX_COLOR, nf, nf1 } from './format'

function Stat({ label, value, mark }: { label: string; value: string; mark?: string }) {
  return (
    <div className="obs-stat">
      <span className="obs-stat-label">
        {mark && <i className="obs-dot" style={{ background: mark }} />}
        {label}
      </span>
      <strong>{value}</strong>
    </div>
  )
}

const n = (v: number | undefined) => (v == null ? '—' : nf.format(v))

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
        <Stat label="🤝 Kebaikan" value={n(info.kindness)} />
        <Stat label="🗡 Kejahatan" value={n(info.crimes)} />
      </div>
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
