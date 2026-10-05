import type { Knowledge, SimInfo, StructureKindInfo, TechInfo } from '../../sim/protocol'
import { Chips } from './Chips'
import { formatClockShort, nf, useSecondsPerYear } from './format'

type Props = {
  knowledge: Knowledge | undefined
  info: SimInfo | undefined
  error: string | null
  onSelect: (id: number) => void
}

const TIER_COUNT = 8

export function CivilizationView({ knowledge, info, error, onSelect }: Props) {
  if (!knowledge) {
    return <p className={`small ${error ? 'obs-error' : 'muted'}`}>{error ? `Gagal memuat peradaban: ${error}` : 'Memuat peradaban…'}</p>
  }

  const tier = knowledge.tier ?? 0
  const kinds = knowledge.structureKinds ?? []
  const techs = knowledge.techs ?? []
  const stationFor = (t: number) => kinds.find((k) => k.tier === t && !k.house)
  const tierName = (t: number) => knowledge.tierNames?.[t] ?? `Tier ${t}`
  const techName = new Map(techs.map((t) => [t.id, t.name]))
  const knownCount = techs.filter((t) => t.known).length

  return (
    <div className="civ">
      <section className="obs-section">
        <h3 className="obs-title">Tingkat peradaban</h3>
        <ol className="civ-ladder">
          {Array.from({ length: TIER_COUNT }, (_, t) => {
            const status = t < tier ? 'reached' : t === tier ? 'current' : 'locked'
            const station = stationFor(t)
            return (
              <li key={t} className={`civ-tier is-${status}`} aria-current={status === 'current' ? 'step' : undefined}>
                <span className="civ-step">{t}</span>
                <div className="civ-tier-body">
                  <strong>{tierName(t)}</strong>
                  <span className="small muted">
                    {t === 0 ? 'Awal mula — tangan kosong' : `Bangun ${station?.name ?? 'stasiun'}`}
                    {station && station.count > 0 ? ` · ${nf.format(station.count)} berdiri` : ''}
                  </span>
                </div>
                <span className="civ-status small">
                  {status === 'reached' ? '✓ tercapai' : status === 'current' ? '● sekarang' : 'terkunci'}
                </span>
              </li>
            )
          })}
        </ol>
      </section>

      {info && (
        <section className="obs-section">
          <h3 className="obs-title">Masyarakat</h3>
          <div className="obs-stat-grid">
            <Tile label="🏠 Rumah" value={info.houses} />
            <Tile label="🏗 Bangunan" value={info.structures} />
            <Tile label="⚗ Unsur" value={info.elementsDiscovered} suffix={` / ${info.elementsTotal ?? 118}`} />
            <Tile label="🤝 Kebaikan" value={info.kindness} />
            <Tile label="🗡 Kejahatan" value={info.crimes} />
            <Tile label="☠ Pembunuhan" value={info.kills} />
          </div>
        </section>
      )}

      <section className="obs-section">
        <div className="obs-section-head">
          <h3 className="obs-title">Teknologi</h3>
          <span className="small muted">
            {nf.format(knownCount)} / {nf.format(techs.length)} dikuasai
          </span>
        </div>
        {techs.length === 0 ? (
          <p className="small muted">Belum ada data teknologi.</p>
        ) : (
          Array.from({ length: TIER_COUNT }, (_, t) => techs.filter((x) => x.tier === t))
            .map((list, t) =>
              list.length === 0 ? null : (
                <div key={t} className="civ-group">
                  <h4 className="civ-group-title small muted">{tierName(t)}</h4>
                  <ul className="civ-techs">
                    {list.map((x) => (
                      <TechRow key={x.id} tech={x} techName={techName} onSelect={onSelect} />
                    ))}
                  </ul>
                </div>
              ),
            )
        )}
      </section>

      <section className="obs-section">
        <h3 className="obs-title">Bangunan</h3>
        <StructureGroup title="Rumah keluarga" kinds={kinds.filter((k) => k.house)} />
        <StructureGroup title="Stasiun ilmu" kinds={kinds.filter((k) => !k.house && k.tier > 0)} />
        <StructureGroup title="Lainnya" kinds={kinds.filter((k) => !k.house && !k.tier)} />
      </section>
    </div>
  )
}

function Tile({ label, value, suffix = '' }: { label: string; value: number | undefined; suffix?: string }) {
  return (
    <div className="obs-stat">
      <span className="obs-stat-label">{label}</span>
      <strong>
        {value == null ? '—' : nf.format(value)}
        {value != null && suffix && <small className="muted">{suffix}</small>}
      </strong>
    </div>
  )
}

function TechRow({
  tech,
  techName,
  onSelect,
}: {
  tech: TechInfo
  techName: Map<string, string>
  onSelect: (id: number) => void
}) {
  const spy = useSecondsPerYear()
  return (
    <li className={`civ-tech ${tech.known ? 'is-known' : ''}`} title={tech.description}>
      <span className="civ-tech-mark" aria-hidden>
        {tech.known ? '✓' : '○'}
      </span>
      <div className="civ-tech-body">
        <strong>{tech.name}</strong>
        {tech.known ? (
          <span className="small muted">
            {tech.learnedBy ? (
              <>
                oleh{' '}
                <button type="button" className="obs-link" onClick={() => onSelect(tech.learnedBy!.id)}>
                  {tech.learnedBy.name}
                </button>
              </>
            ) : (
              'dikuasai'
            )}
            {tech.learnedAt != null && ` · ${formatClockShort(tech.learnedAt, spy)}`}
          </span>
        ) : (
          <span className="small muted">
            {tech.requires?.length
              ? `butuh: ${tech.requires.map((r) => techName.get(r) ?? r).join(', ')}`
              : tech.description || 'belum ditemukan'}
          </span>
        )}
      </div>
    </li>
  )
}

function StructureGroup({ title, kinds }: { title: string; kinds: StructureKindInfo[] }) {
  if (kinds.length === 0) return null
  return (
    <div className="civ-group">
      <h4 className="civ-group-title small muted">{title}</h4>
      <ul className="civ-structs">
        {kinds.map((k) => (
          <li key={k.id} className={k.count > 0 ? 'is-built' : ''}>
            <div className="civ-struct-head">
              <strong>
                {k.name}
                {k.house && k.level ? <span className="muted"> · Lv {k.level}</span> : null}
                {!k.house && k.tier ? <span className="muted"> · tier {k.tier}</span> : null}
              </strong>
              <span className="obs-pill civ-count">{nf.format(k.count)}</span>
            </div>
            <Chips stacks={k.cost} empty="tanpa bahan" />
          </li>
        ))}
      </ul>
    </div>
  )
}
