import type { Knowledge, SimInfo, StructureKindInfo, TechInfo } from '../../sim/protocol'
import { Chips } from './Chips'
import { formatClockShort, nf, nf1, useSecondsPerYear } from './format'

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
  // Newer servers track living holders; then "known" means someone can practise it now.
  const perPerson = techs.some((t) => t.holders != null)
  const knownCount = techs.filter((t) => (perPerson ? (t.holders ?? 0) > 0 || t.written : t.known)).length
  const lostCount = techs.filter((t) => t.lost).length

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
            <RateTile label="🤝 Kebaikan" rate={info.kindnessPerYear} total={info.kindness} />
            <RateTile label="🗡 Kejahatan" rate={info.crimesPerYear} total={info.crimes} />
            <RateTile label="☠ Pembunuhan" rate={info.killsPerYear} total={info.kills} />
            {info.avgBrainSize != null && <Tile label="🧠 Otak rata-rata" value={info.avgBrainSize} suffix=" neuron" fraction />}
            {info.knowledgeLost != null && <Tile label="📉 Ilmu hilang" value={info.knowledgeLost} suffix=" kali" />}
          </div>
          {info.kindnessPerYear != null && <p className="small muted">Laju per tahun selama 50 tahun terakhir; total sepanjang masa di tooltip.</p>}
        </section>
      )}

      <section className="obs-section">
        <div className="obs-section-head">
          <h3 className="obs-title">Teknologi</h3>
          <span className="small muted">
            {nf.format(knownCount)} / {nf.format(techs.length)} dikuasai
            {lostCount > 0 && <> · {nf.format(lostCount)} hilang</>}
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
        <StructureGroup title="Stasiun ilmu" kinds={kinds.filter((k) => !k.house && k.tier > 0 && !isLibrary(k))} />
        <StructureGroup title="Tulisan & ilmu" kinds={kinds.filter((k) => !k.house && isLibrary(k))} />
        <StructureGroup title="Lainnya" kinds={kinds.filter((k) => !k.house && !k.tier && !isLibrary(k))} />
      </section>
    </div>
  )
}

function Tile({
  label,
  value,
  suffix = '',
  fraction = false,
}: {
  label: string
  value: number | undefined
  suffix?: string
  fraction?: boolean
}) {
  return (
    <div className="obs-stat">
      <span className="obs-stat-label">{label}</span>
      <strong>
        {value == null ? '—' : (fraction ? nf1 : nf).format(value)}
        {value != null && suffix && <small className="muted">{suffix}</small>}
      </strong>
    </div>
  )
}

/** A rate per simulated year when the server reports one; otherwise the all-time total. */
function RateTile({ label, rate, total }: { label: string; rate: number | undefined; total: number | undefined }) {
  if (rate == null) return <Tile label={label} value={total} />
  return (
    <div className="obs-stat" title={total != null ? `Total sepanjang masa: ${nf.format(total)}` : undefined}>
      <span className="obs-stat-label">{label}</span>
      <strong>
        {nf1.format(rate)}
        <small className="muted">/thn</small>
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
  const tracked = tech.holders != null
  const holders = tech.holders ?? 0
  // Lost: discovered once, but nobody alive can practise it and nothing was written down.
  const lost = !!tech.lost
  const alive = tracked ? holders > 0 || !!tech.written : tech.known
  const state = lost ? 'is-lost' : alive ? 'is-known' : ''
  return (
    <li className={`civ-tech ${state}`} title={tech.description}>
      <span className="civ-tech-mark" aria-hidden>
        {lost ? '✕' : alive ? '✓' : '○'}
      </span>
      <div className="civ-tech-body">
        <strong>
          {tech.name}
          {lost && <span className="civ-tag civ-tag-lost">hilang</span>}
          {tech.written && <span className="civ-tag civ-tag-written">📜 tertulis</span>}
        </strong>
        {tracked && tech.known && (
          <span className="small muted">
            {holders > 0
              ? `dikuasai ${nf.format(holders)} orang`
              : tech.written
                ? 'tak ada yang menguasai — bisa dipelajari lagi dari perpustakaan'
                : 'tak ada lagi yang menguasainya'}
          </span>
        )}
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

/** Libraries keep written skills alive after their holders die. */
const isLibrary = (k: StructureKindInfo) => k.id === 'perpustakaan'

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
