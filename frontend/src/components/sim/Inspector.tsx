import { useQuery } from '@tanstack/react-query'
import { ApiError } from '../../api/client'
import { creatureQuery } from '../../sim/api'
import {
  ACTION_LABELS,
  ROLE_LABELS,
  SEX_SYMBOL,
  type CreatureDetail,
  type CreatureRef,
  type Deeds,
  type HouseDetail,
  type Skill,
} from '../../sim/protocol'
import { BrainView } from './BrainView'
import { Chips } from './Chips'
import { SEX_COLOR, SEX_LABEL, formatYears, hueColor, nf, reputationWord, useSecondsPerYear } from './format'

type Props = {
  mapId: string
  id: number
  following: boolean
  onSelect: (id: number | null) => void
  onFollow: (follow: boolean) => void
}

const STATUS = { warning: '#fab219', critical: '#d03b3b' }
// Diverging pair (dataviz reference, dark steps): red = disliked, blue = respected.
const REPUTATION = { neg: '#e66767', pos: '#3987e5' }

/** Fill carries severity; the track is a faint step of the same colour. */
function Meter({ label, value, text, color, severity = false }: {
  label: string
  value: number
  text: string
  color: string
  severity?: boolean
}) {
  const v = Math.min(1, Math.max(0, value))
  const level = !severity ? null : v < 0.15 ? 'kritis' : v < 0.3 ? 'rendah' : null
  const fill = level === 'kritis' ? STATUS.critical : level === 'rendah' ? STATUS.warning : color
  return (
    <div className="obs-meter">
      <div className="obs-meter-head">
        <span>{label}</span>
        <span className="obs-meter-value">
          {text}
          {level && <em className={`obs-level obs-level-${level === 'kritis' ? 'critical' : 'warning'}`}> · {level}</em>}
        </span>
      </div>
      <div className="obs-meter-track" style={{ background: `color-mix(in srgb, ${fill} 22%, transparent)` }}>
        <div className="obs-meter-fill" style={{ width: `${v * 100}%`, background: fill }} />
      </div>
    </div>
  )
}

/** −1..1 bar growing out from the centre; the word carries the meaning, not the colour. */
function ReputationMeter({ value }: { value: number }) {
  const v = Math.min(1, Math.max(-1, value))
  const color = v >= 0 ? REPUTATION.pos : REPUTATION.neg
  return (
    <div className="obs-meter">
      <div className="obs-meter-head">
        <span>Reputasi</span>
        <span className="obs-meter-value">
          {reputationWord(v)} ({v >= 0 ? '+' : ''}
          {v.toFixed(2)})
        </span>
      </div>
      <div className="obs-meter-track obs-rep-track">
        <div
          className="obs-meter-fill obs-rep-fill"
          style={{ left: v >= 0 ? '50%' : `${50 + v * 50}%`, width: `${Math.abs(v) * 50}%`, background: color }}
        />
      </div>
    </div>
  )
}

function PersonLink({ who, none, onSelect }: { who: CreatureRef | null | undefined; none: string; onSelect: (id: number) => void }) {
  if (!who) return <span className="muted">{none}</span>
  // id 0 is not a creature (e.g. reading at an ownerless library), so nothing to select.
  if (who.id === 0) return <span>{who.name}</span>
  return (
    <button type="button" className="obs-link" onClick={() => onSelect(who.id)}>
      {who.name}
    </button>
  )
}

const pct = (v: number) => `${Math.round(v * 100)}%`

const ROLE_ICON: Record<CreatureDetail['role'], string> = { head: '👑', member: '🏠', none: '⛺' }

export function Inspector({ mapId, id, following, onSelect, onFollow }: Props) {
  const q = useQuery(creatureQuery(mapId, id))
  const spy = useSecondsPerYear()

  if (q.isError) {
    const gone = q.error instanceof ApiError && q.error.status === 404
    return (
      <section className="obs-section obs-inspector obs-gone">
        <p>{gone ? 'Makhluk ini telah tiada.' : `Gagal memuat makhluk: ${q.error.message}`}</p>
        <button type="button" onClick={() => onSelect(null)}>
          Tutup
        </button>
      </section>
    )
  }

  if (!q.data) {
    return (
      <section className="obs-section obs-inspector">
        <p className="muted small">Memuat makhluk…</p>
      </section>
    )
  }

  const c = q.data
  const t = c.traits
  const firstHumans = !c.mother && !c.father

  return (
    <section className="obs-section obs-inspector">
      <div className="obs-insp-head">
        <span className="obs-swatch" style={{ background: hueColor(t.hue) }} title={`Warna garis keturunan ${Math.round(t.hue)}°`} />
        <div className="obs-insp-title">
          <h3>
            {c.name}{' '}
            <span className="obs-sex" style={{ color: SEX_COLOR[c.sex] }} aria-label={SEX_LABEL[c.sex]}>
              {SEX_SYMBOL[c.sex]}
            </span>
          </h3>
          <p className="small muted">
            {SEX_LABEL[c.sex]} · Generasi {c.generation} · {c.adult ? 'Dewasa' : 'Anak'}
          </p>
        </div>
        <button type="button" className="ghost obs-close" aria-label="Tutup" title="Tutup" onClick={() => onSelect(null)}>
          ×
        </button>
      </div>

      <div className="obs-insp-status">
        <span className="obs-pill">{ACTION_LABELS[c.action] ?? c.action}</span>
        {c.role && (
          <span className={`obs-pill obs-role obs-role-${c.role}`}>
            {ROLE_ICON[c.role]} {ROLE_LABELS[c.role]}
          </span>
        )}
        {c.pregnant && <span className="obs-pill obs-pill-pregnant">Hamil</span>}
        <label className="obs-follow">
          <input type="checkbox" checked={following} onChange={(e) => onFollow(e.target.checked)} />
          Ikuti kamera
        </label>
      </div>

      <div className="obs-meters">
        <Meter
          label="Usia"
          value={c.age / Math.max(1, c.lifespan)}
          text={`${formatYears(c.age, c.secondsPerYear ?? spy)} / ${formatYears(c.lifespan, c.secondsPerYear ?? spy)}`}
          color="#8ea0bb"
        />
        <Meter label="Energi" value={c.energy} text={pct(c.energy)} color="#1baf7a" severity />
        <Meter label="Hidrasi" value={c.hydration} text={pct(c.hydration)} color="#3987e5" severity />
        {c.health != null && <Meter label="Kesehatan" value={c.health} text={pct(c.health)} color="#9085e9" severity />}
        {c.pregnant && <Meter label="Kehamilan" value={c.gestation} text={pct(c.gestation)} color="#d55181" />}
        {c.reputation != null && <ReputationMeter value={c.reputation} />}
      </div>

      <h4 className="obs-subtitle">Keluarga</h4>
      <dl className="obs-kv">
        <dt>Ibu</dt>
        <dd>
          <PersonLink who={c.mother} none="— (manusia pertama)" onSelect={onSelect} />
        </dd>
        <dt>Ayah</dt>
        <dd>
          <PersonLink who={c.father} none="— (manusia pertama)" onSelect={onSelect} />
        </dd>
        <dt>Pasangan</dt>
        <dd>
          <PersonLink who={c.spouse} none="—" onSelect={onSelect} />
        </dd>
        <dt>Anak</dt>
        <dd>{nf.format(c.children)}</dd>
        <dt>Posisi</dt>
        <dd>
          ({c.x.toFixed(1)}, {c.y.toFixed(1)})
        </dd>
      </dl>
      {firstHumans && c.generation === 0 && (
        <p className="obs-note small muted">Tanpa orang tua — awal dari seluruh umat di dunia ini.</p>
      )}

      {c.house ? (
        <HouseCard house={c.house} selfId={c.id} onSelect={onSelect} />
      ) : (
        c.role === 'none' && <p className="obs-note small muted">Belum punya rumah.</p>
      )}

      <h4 className="obs-subtitle">Bawaan</h4>
      <Chips stacks={c.inventory} empty="Tidak membawa apa-apa." />

      {c.skills && <SkillsCard creature={c} onSelect={onSelect} />}

      {c.deeds && <DeedsGrid deeds={c.deeds} />}

      <h4 className="obs-subtitle">Sifat bawaan</h4>
      <dl className="obs-kv">
        <dt>Warna keturunan</dt>
        <dd>
          <i className="obs-swatch-sm" style={{ background: hueColor(t.hue) }} /> {Math.round(t.hue)}°
        </dd>
        <dt>Ukuran tubuh</dt>
        <dd>{t.size.toFixed(2)}×</dd>
        <dt>Kecepatan maks</dt>
        <dd>{t.maxSpeed.toFixed(2)} tile/detik</dd>
        <dt>Jarak pandang</dt>
        <dd>{t.vision.toFixed(1)} tile</dd>
        <dt>Metabolisme</dt>
        <dd>{t.metabolism.toFixed(2)}×</dd>
        <dt>Laju mutasi</dt>
        <dd>{(t.mutationRate * 100).toFixed(1)}%</dd>
      </dl>

      <h4 className="obs-subtitle">Otak</h4>
      <p className="obs-note small muted">
        Jaringan saraf {c.brain.input.length}→{c.brain.hidden.length}→{c.brain.output.length} dengan memori, diwarisi dari
        kedua induk lalu bermutasi.
        {c.brain.learningRate != null &&
          ' Selama hidup, bobotnya ikut berubah oleh pengalaman (kenyang, haus, sakit) — hasil belajar ini tidak diwariskan.'}{' '}
        Semua keputusannya — bekerja, berbagi, mengajar, mencuri, atau menyerang — berasal dari jaringan ini sendiri,
        ditambah input <em>acak (kehendak)</em>. Kamu hanya mengamati.
      </p>
      {(c.brain.size != null || c.brain.learningRate != null) && (
        <dl className="obs-kv">
          <dt>Neuron tersembunyi</dt>
          <dd>
            {nf.format(c.brain.size ?? c.brain.hidden.length)}{' '}
            <span className="muted small">(diwarisi, bisa bertambah/berkurang lewat evolusi)</span>
          </dd>
          {c.brain.learningRate != null && (
            <>
              <dt>Laju belajar</dt>
              <dd>
                {c.brain.learningRate > 0 ? c.brain.learningRate.toPrecision(2).replace('.', ',') : '0 (tidak belajar)'}
              </dd>
            </>
          )}
        </dl>
      )}
      <BrainView brain={c.brain} />
    </section>
  )
}

/** At or above this level a skill can be put to work (crafting, building). */
const PRACTICE_THRESHOLD = 0.3

function SkillsCard({ creature: c, onSelect }: { creature: CreatureDetail; onSelect: (id: number) => void }) {
  const skills = [...(c.skills ?? [])].sort((a, b) => b.level - a.level)
  return (
    <>
      <h4 className="obs-subtitle">Keahlian & belajar</h4>
      <dl className="obs-kv">
        <dt>Belajar dari</dt>
        <dd>
          <PersonLink who={c.teacher} none="— (tidak sedang belajar)" onSelect={onSelect} />
        </dd>
        <dt>Pernah mengajar</dt>
        <dd>{c.taught ? `${nf.format(c.taught)} orang` : 'belum pernah'}</dd>
      </dl>
      {skills.length === 0 ? (
        <p className="obs-note small muted">Belum menguasai keahlian apa pun.</p>
      ) : (
        <ul className="obs-skills" aria-label="Keahlian">
          {skills.map((sk) => (
            <SkillBar key={sk.tech} skill={sk} />
          ))}
        </ul>
      )}
      <p className="obs-note small muted">
        Garis putih = batas {Math.round(PRACTICE_THRESHOLD * 100)}%: di atasnya keahlian bisa dipraktikkan. Keahlian naik
        karena latihan dan diajari, turun pelan bila tak dipakai, dan hilang bersama pemiliknya kalau tak diajarkan atau
        ditulis.
      </p>
    </>
  )
}

function SkillBar({ skill }: { skill: Skill }) {
  const v = Math.min(1, Math.max(0, skill.level))
  const can = v >= PRACTICE_THRESHOLD
  return (
    <li>
      <div className="obs-skill-head">
        <span>{skill.name}</span>
        <span>
          {pct(v)} <em>· {can ? 'bisa dipraktikkan' : 'masih belajar'}</em>
        </span>
      </div>
      <div
        className="obs-skill-track"
        role="meter"
        aria-label={`${skill.name}: ${pct(v)}`}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={Math.round(v * 100)}
      >
        <div className="obs-skill-fill" style={{ width: `${v * 100}%`, opacity: can ? 1 : 0.6 }} />
        <i className="obs-skill-threshold" style={{ left: `calc(${PRACTICE_THRESHOLD * 100}% - 1px)` }} aria-hidden />
      </div>
    </li>
  )
}

function HouseCard({ house, selfId, onSelect }: { house: HouseDetail; selfId: number; onSelect: (id: number) => void }) {
  const used = (house.storage ?? []).reduce((n, s) => n + s.qty, 0)
  return (
    <div className="obs-house">
      <div className="obs-house-head">
        <strong>
          🏠 {house.name}
          {house.level ? <span className="muted"> · Lv {house.level}</span> : null}
        </strong>
        <span className="small muted">
          ({house.x}, {house.y})
        </span>
      </div>
      <dl className="obs-kv">
        <dt>Kepala keluarga</dt>
        <dd>
          {house.head?.id === selfId ? (
            <span>dirinya sendiri</span>
          ) : (
            <PersonLink who={house.head} none="— (terbengkalai)" onSelect={onSelect} />
          )}
        </dd>
        <dt>Anggota</dt>
        <dd>{nf.format(house.members)}</dd>
        <dt>Gudang</dt>
        <dd>
          {nf.format(Math.round(used))} / {nf.format(house.capacity)}
        </dd>
      </dl>
      <Chips stacks={house.storage} empty="Gudang kosong." />
      {house.granary?.length > 0 && (
        <>
          <h5 className="obs-house-sub">🌾 Lumbung</h5>
          <Chips stacks={house.granary} empty="Lumbung kosong." />
        </>
      )}
      {house.livestock?.length > 0 && (
        <>
          <h5 className="obs-house-sub">🐔 Ternak</h5>
          <Chips stacks={house.livestock} empty="Belum punya ternak." />
        </>
      )}
    </div>
  )
}

const DEEDS: { key: keyof Deeds; label: string }[] = [
  { key: 'kindness', label: '🤝 Kebaikan' },
  { key: 'crimes', label: '🗡 Kejahatan' },
  { key: 'kills', label: '☠ Pembunuhan' },
  { key: 'built', label: '🔨 Dibangun' },
  { key: 'crafted', label: '⚒ Dibuat' },
  { key: 'discoveries', label: '⚗ Penemuan' },
  { key: 'planted', label: '🌱 Menanam' },
  { key: 'harvested', label: '🌾 Memanen' },
  { key: 'hunted', label: '🏹 Berburu' },
  { key: 'tamed', label: '🐔 Menjinakkan' },
]

function DeedsGrid({ deeds }: { deeds: Deeds }) {
  return (
    <>
      <h4 className="obs-subtitle">Perbuatan</h4>
      <div className="obs-stat-grid obs-deeds">
        {DEEDS.map((d) => (
          <div key={d.key} className="obs-stat" title={d.label}>
            <span className="obs-stat-label">{d.label}</span>
            <strong>{nf.format(deeds[d.key] ?? 0)}</strong>
          </div>
        ))}
      </div>
    </>
  )
}
