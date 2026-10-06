import { useQuery } from '@tanstack/react-query'
import { lazy, Suspense, useState } from 'react'
import { ApiError } from '../../api/client'
import { creatureQuery } from '../../sim/api'
import {
  ACTION_LABELS,
  DISEASE_LABELS,
  LOCOMOTION_LABELS,
  MOOD_LABELS,
  ROLE_LABELS,
  SEX_SYMBOL,
  STAGE_LABELS,
  type BodyView,
  type CreatureDetail,
  type CreatureRef,
  type Deeds,
  type ExchangeView,
  type HouseDetail,
  type MoodView,
  type Skill,
} from '../../sim/protocol'
import { Chips } from './Chips'
import { SEX_COLOR, SEX_LABEL, formatYears, hueColor, nf, reputationWord, useSecondsPerYear } from './format'
import { MOOD_BAR_LABELS, MOOD_COLOR, MOOD_KEYS, dominantMood, temperWord } from './moods'
import { NutritionSection } from './NutritionSection'
import { FamilyTree } from './FamilyTree'

type Props = {
  mapId: string
  id: number
  following: boolean
  onSelect: (id: number | null) => void
  onFollow: (follow: boolean) => void
  /** Opens the Neuron tab, where this brain is shown whole. */
  onShowBrain?: () => void
}

const STATUS = { warning: '#fab219', critical: '#d03b3b' }
const CharacterPortrait = lazy(() => import('../../game3d/CharacterPortrait'))
// Diverging pair (dataviz reference, dark steps): red = disliked, blue = respected.
const REPUTATION = { neg: '#e66767', pos: '#3987e5' }

/** Fill carries severity; the track is a faint step of the same colour. */
function Meter({
  label,
  value,
  text,
  color,
  severity = false,
}: {
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

/** A person's name as a button that opens them in the Inspector (plain text for id 0, "none" when missing). */
export function PersonLink({ who, none, onSelect }: { who: CreatureRef | null | undefined; none: string; onSelect: (id: number) => void }) {
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

const LOCOMOTION_ICON: Record<NonNullable<CreatureDetail['locomotion']>, string> = {
  '': '',
  wading: '🌊',
  swimming: '🏊',
  rafting: '🛶',
  climbing: '🧗',
}

const MEMORY_MODE: Record<CreatureDetail['memories'][number]['mode'], string> = {
  direct: 'Mengalami',
  seen: 'Melihat',
  heard: 'Mendengar',
  told: 'Diceritakan',
}

export function Inspector({ mapId, id, following, onSelect, onFollow, onShowBrain }: Props) {
  const [portrait, setPortrait] = useState(false)
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
            {SEX_LABEL[c.sex]} · Generasi {c.generation} ·{' '}
            {c.body ? (STAGE_LABELS[c.body.stage] ?? c.body.stage) : c.adult ? 'Dewasa' : 'Anak'}
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
        {c.locomotion && (
          <span className="obs-pill obs-pill-move" title="Cara bergerak saat ini">
            {LOCOMOTION_ICON[c.locomotion]} {LOCOMOTION_LABELS[c.locomotion] ?? c.locomotion}
          </span>
        )}
        {c.village && (
          <span
            className={`obs-pill obs-pill-village ${c.leader ? 'leader' : ''}`}
            title={c.leader ? `Pemimpin desa ${c.village.name}: orang yang paling dipercaya warganya` : `Warga desa ${c.village.name}`}
          >
            {c.leader ? '🚩 Pemimpin' : '🏘'} {c.village.name}
          </span>
        )}
        {!c.village && c.leader && <span className="obs-pill obs-pill-village leader">🚩 Pemimpin desa</span>}
        {(c.waterRoom ?? 0) > 0 && (
          <span className="obs-pill" title="Air minum yang dibawa dalam tabung bambu">
            💧 Tabung {Math.round((100 * (c.waterCarried ?? 0)) / (c.waterRoom ?? 1))}%
          </span>
        )}
        {(c.foodPlaces?.length ?? 0) > 0 && (
          <span
            className="obs-pill"
            title={`Tempat makanan yang diingat:\n${c
              .foodPlaces!.map((p) => `(${Math.round(p.x)}, ${Math.round(p.y)}) · ${Math.round(p.rich * 100)}%`)
              .join('\n')}`}
          >
            🧺 Ingat {c.foodPlaces!.length} tempat makan
          </span>
        )}
        {c.pregnant && <span className="obs-pill obs-pill-pregnant">Hamil</span>}
        {c.body?.ill && <span className={`obs-pill obs-pill-ill ${c.body.ill.danger >= 1 ? 'grave' : ''}`}>🌡 {c.body.ill.name}</span>}
        <label className="obs-follow">
          <input type="checkbox" checked={following} onChange={(e) => onFollow(e.target.checked)} />
          Ikuti kamera
        </label>
      </div>

      <details className="obs-portrait" onToggle={(e) => setPortrait(e.currentTarget.open)}>
        <summary>Karakter 3D · lihat lebih dekat</summary>
        {portrait && (
          <Suspense fallback={<p className="small muted">Memuat karakter…</p>}>
            <CharacterPortrait person={c} />
          </Suspense>
        )}
        {portrait && <p className="small muted">Seret untuk memutar · gulir untuk memperbesar</p>}
      </details>

      {c.execution && (
        <div className="obs-execution">
          <strong>{ACTION_LABELS[c.execution.action] ?? c.execution.action}</strong>
          <span>
            {
              { approach: 'Menuju lokasi', blocked: 'Jalur terhalang', reached: 'Tiba di lokasi', perform: 'Sedang bekerja' }[
                c.execution.phase
              ]
            }
          </span>
          {c.execution.target && (
            <small className="muted">
              Tujuan {c.execution.target.x.toFixed(1)}, {c.execution.target.y.toFixed(1)} · {c.execution.waypoints} langkah rute
            </small>
          )}
          {c.execution.remaining != null && <small className="muted">Sisa {c.execution.remaining.toFixed(1)} detik simulasi</small>}
        </div>
      )}

      <details className="obs-private-memory">
        <summary>Hubungan & ingatan · {c.relations?.length ?? 0} orang</summary>
        <p className="small muted">Penilaian pribadi dari pengalaman dan kejadian yang disaksikan.</p>
        {(c.relations ?? []).map((r) => (
          <div className="obs-relation" key={r.person.id}>
            <PersonLink who={r.person} none="Tidak dikenal" onSelect={onSelect} />
            <span style={{ color: r.trust >= 0 ? REPUTATION.pos : REPUTATION.neg }}>
              {r.trust > 0.05 ? 'Percaya' : r.trust < -0.05 ? 'Tidak percaya' : 'Netral'} {r.trust.toFixed(2)}
            </span>
          </div>
        ))}
        {!c.relations?.length && <p className="small muted">Belum ada pengalaman sosial yang diingat.</p>}
        <ol className="obs-memories">
          {(c.memories ?? [])
            .slice(-5)
            .reverse()
            .map((m, i) => (
              <li
                key={`${m.tick}-${i}`}
                className={m.mode === 'told' ? 'obs-mem-told' : undefined}
                title={m.mode === 'told' ? 'Kabar dari orang lain (gosip): tidak disaksikan sendiri, jadi kurang diyakini' : undefined}
              >
                <span>
                  {m.mode === 'told' && <span aria-hidden="true">💬 </span>}
                  {MEMORY_MODE[m.mode] ?? m.mode} {{ give: 'pemberian', steal: 'pencurian', attack: 'pertengkaran' }[m.kind]}
                </span>
                {m.actor && (
                  <>
                    {' '}
                    · <PersonLink who={m.actor} none="" onSelect={onSelect} />
                  </>
                )}
                <small className="muted"> · keyakinan {Math.round(m.confidence * 100)}%</small>
              </li>
            ))}
        </ol>
      </details>

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

      {c.mood && <MoodCard mood={c.mood} />}
      {c.exchange && <ExchangeCard exchange={c.exchange} action={c.action} onSelect={onSelect} />}

      {c.body && <BodyCard body={c.body} female={c.sex === 'female'} adult={c.adult} onSelect={onSelect} />}

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
      {firstHumans && c.generation === 0 && <p className="obs-note small muted">Tanpa orang tua — awal dari seluruh umat di dunia ini.</p>}
      <h4 className="obs-subtitle">Pohon keluarga &amp; genetika</h4>
      <FamilyTree mapId={mapId} id={c.id} onSelect={onSelect} />

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
        Semua keputusannya — bekerja, berbagi, mengajar, mencuri, atau menyerang — berasal dari jaringan sarafnya sendiri, ditambah input{' '}
        <em>acak (kehendak)</em>. Kamu hanya mengamati.
      </p>
      <dl className="obs-kv">
        <dt>Neuron</dt>
        <dd>
          {nf.format(c.brain.hidden.length + (c.brain.grown ?? 0))}{' '}
          <span className="muted small">
            ({nf.format(c.brain.hidden.length)} bawaan
            {c.brain.grown != null && <> + {nf.format(c.brain.grown)} tumbuh selama hidup</>})
          </span>
        </dd>
        {c.brain.learningRate != null && (
          <>
            <dt>Laju belajar</dt>
            <dd>{c.brain.learningRate > 0 ? c.brain.learningRate.toPrecision(2).replace('.', ',') : '0 (tidak belajar)'}</dd>
          </>
        )}
      </dl>
      {onShowBrain && (
        <button type="button" className="obs-brain-open" onClick={onShowBrain}>
          🧠 Lihat seluruh otaknya di tab Neuron →
        </button>
      )}
    </section>
  )
}

/** Moods 0–1 each (engine adaptation II): four bars and the one that shows on their face. */
function MoodCard({ mood }: { mood: MoodView }) {
  const dominant = dominantMood(mood)
  return (
    <>
      <h4 className="obs-subtitle">Suasana hati</h4>
      <div className="obs-mood">
        <p className="obs-mood-now">
          <i
            className="obs-dot"
            style={dominant ? { background: MOOD_COLOR[dominant] } : { background: 'transparent', boxShadow: 'inset 0 0 0 2px #8ea0bb' }}
            aria-hidden
          />
          <strong>{MOOD_LABELS[dominant] ?? dominant}</strong>
          <span className="small muted">{dominant ? '— terlihat di wajahnya' : '— tidak ada perasaan yang menonjol'}</span>
        </p>
        <div className="obs-mood-bars">
          {MOOD_KEYS.map((k) => {
            const v = Math.min(1, Math.max(0, mood[k] ?? 0))
            return (
              <div key={k} className={`obs-mood-row ${k === dominant ? 'dominant' : ''}`}>
                <span>{MOOD_BAR_LABELS[k]}</span>
                <div
                  className="obs-meter-track"
                  role="meter"
                  aria-label={`${MOOD_BAR_LABELS[k]}: ${pct(v)}`}
                  aria-valuemin={0}
                  aria-valuemax={100}
                  aria-valuenow={Math.round(v * 100)}
                  style={{ background: `color-mix(in srgb, ${MOOD_COLOR[k]} 18%, transparent)` }}
                >
                  <div className="obs-meter-fill" style={{ width: `${v * 100}%`, background: MOOD_COLOR[k] }} />
                </div>
                <span className="obs-mood-val">{pct(v)}</span>
              </div>
            )
          })}
        </div>
        {(mood.reactivity != null || mood.recovery != null || mood.cheer != null) && (
          <p
            className="small muted obs-note"
            title="Gen watak bawaan (sekitar 1): seberapa kuat kejadian menggerakkan perasaannya, seberapa cepat reda, dan seberapa riang saat tenang"
          >
            Watak: {temperWord(mood.reactivity, 'tenang', 'peka')} · pulih {temperWord(mood.recovery, 'lambat', 'cepat')} ·{' '}
            {temperWord(mood.cheer, 'murung', 'riang')}
          </p>
        )}
      </div>
    </>
  )
}

const CHAT_VERB = { talk: 'mengobrol', trade: 'bertukar barang' } as const

/** The two-person exchange under way (or the last one) and how many they have had. */
function ExchangeCard({
  exchange: x,
  action,
  onSelect,
}: {
  exchange: ExchangeView
  action: CreatureDetail['action']
  onSelect: (id: number) => void
}) {
  const chat = x.chat
  const now = chat && (action === 'talk' || action === 'trade')
  return (
    <>
      <h4 className="obs-subtitle">Obrolan &amp; barter</h4>
      {chat ? (
        <p className={`obs-chat obs-chat-${chat.kind}`}>
          <span aria-hidden="true">{chat.kind === 'trade' ? '🔁' : '💬'} </span>
          {chat.ended == null ? (
            <>
              Mengajak <PersonLink who={chat.with} none="seseorang" onSelect={onSelect} /> {CHAT_VERB[chat.kind]}{' '}
              <span className="muted">— menunggu jawaban</span>
            </>
          ) : chat.ok ? (
            <>
              {now ? 'Sedang ' : 'Terakhir '}
              {CHAT_VERB[chat.kind]} dengan <PersonLink who={chat.with} none="seseorang" onSelect={onSelect} />
            </>
          ) : (
            <>
              Ajakan {CHAT_VERB[chat.kind]} kepada <PersonLink who={chat.with} none="seseorang" onSelect={onSelect} />{' '}
              <span className="muted">tidak berbalas</span>
            </>
          )}
        </p>
      ) : (
        <p className="small muted obs-note">
          {(x.talks ?? 0) + (x.trades ?? 0) > 0
            ? 'Tidak sedang mengobrol atau bertukar barang.'
            : 'Belum pernah mengobrol atau bertukar barang.'}
        </p>
      )}
      <dl className="obs-kv">
        <dt>Obrolan</dt>
        <dd>{nf.format(x.talks ?? 0)} kali</dd>
        <dt>Barter</dt>
        <dd>{nf.format(x.trades ?? 0)} kali</dd>
        {x.told != null && (
          <>
            <dt title="Ingatan tentang orang lain yang hanya ia dengar dari cerita (gosip)">Ingatan dari cerita</dt>
            <dd>{nf.format(x.told)}</dd>
          </>
        )}
      </dl>
    </>
  )
}

const IMMUNITY_HINT = 'Kekebalan dari sakit sebelumnya; memudar bila lama tidak terpapar'

/** Body and health (Fase 3): the illness now, immunity, worms, nursing and fertility. */
function BodyCard({
  body: b,
  female,
  adult,
  onSelect,
}: {
  body: BodyView
  female: boolean
  adult: boolean
  onSelect: (id: number) => void
}) {
  const ill = b.ill
  return (
    <>
      <h4 className="obs-subtitle">Tubuh &amp; kesehatan</h4>
      {ill ? (
        <div className={`obs-ill ${ill.danger >= 1 ? 'grave' : ill.danger >= 0.6 ? 'serious' : ''}`}>
          <strong>🌡 Sakit {ill.name}</strong>
          <span className="small">
            {pct(ill.progress)} perjalanan penyakit · {ill.danger >= 1 ? 'kritis — bisa meninggal' : ill.danger >= 0.6 ? 'berat' : 'ringan'}
          </span>
          <div className="obs-meter-track" aria-hidden>
            <div className="obs-meter-fill" style={{ width: `${ill.progress * 100}%`, background: '#e66767' }} />
          </div>
        </div>
      ) : (
        <p className="small muted obs-note">
          Sehat{b.carrier ? ' — tetapi masih membawa parasit malaria (bisa menular lewat nyamuk)' : ''}.
        </p>
      )}
      <NutritionSection nutrition={b.nutrition} />
      <dl className="obs-kv">
        <dt title={IMMUNITY_HINT}>Kekebalan</dt>
        <dd title={IMMUNITY_HINT}>
          {(Object.keys(DISEASE_LABELS) as (keyof typeof DISEASE_LABELS)[]).map((k, i) => (
            <span key={k}>
              {i > 0 && ' · '}
              {DISEASE_LABELS[k]} {pct(b.immunity?.[k] ?? 0)}
            </span>
          ))}
        </dd>
        <dt title="Gen bawaan: pertahanan tubuh lebih kuat, tetapi butuh energi lebih banyak">Gen kekebalan</dt>
        <dd>{b.gene.toFixed(2)}×</dd>
        <dt>Cacingan</dt>
        <dd>
          {b.worms < 0.05
            ? 'tidak'
            : b.worms < 0.3
              ? `ringan (${pct(b.worms)})`
              : b.worms < 0.6
                ? `sedang (${pct(b.worms)})`
                : `berat (${pct(b.worms)})`}
        </dd>
        {b.vigor < 0.999 && (
          <>
            <dt>Tenaga</dt>
            <dd>{pct(b.vigor)} (melemah karena usia)</dd>
          </>
        )}
        {b.carer && (
          <>
            <dt>Diasuh oleh</dt>
            <dd>
              <PersonLink who={b.carer} none="—" onSelect={onSelect} />
            </dd>
          </>
        )}
        {female && adult && (
          <>
            <dt>Menopause</dt>
            <dd>umur {Math.round(b.menopause)}</dd>
            <dt title="Peluang hamil dalam sebulan bila berhubungan, menurut umur, gizi, dan menyusui">Kesuburan</dt>
            <dd>{b.fertility > 0 ? `${pct(b.fertility)} per bulan` : b.nursing ? 'tertahan (menyusui)' : '—'}</dd>
          </>
        )}
        {b.nursing && (
          <>
            <dt>Menyusui</dt>
            <dd>ya</dd>
          </>
        )}
      </dl>
    </>
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
        Garis putih = batas {Math.round(PRACTICE_THRESHOLD * 100)}%: di atasnya keahlian bisa dipraktikkan. Keahlian naik karena latihan dan
        diajari, turun pelan bila tak dipakai, dan hilang bersama pemiliknya kalau tak diajarkan atau ditulis.
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
  { key: 'talks', label: '🗣 Mengobrol' },
  { key: 'trades', label: '🔁 Barter' },
]

function DeedsGrid({ deeds }: { deeds: Deeds }) {
  return (
    <>
      <h4 className="obs-subtitle">Perbuatan</h4>
      <div className="obs-stat-grid obs-deeds">
        {/* Talks and trades only from servers that count them. */}
        {DEEDS.filter((d) => deeds[d.key] != null || (d.key !== 'talks' && d.key !== 'trades')).map((d) => (
          <div key={d.key} className="obs-stat" title={d.label}>
            <span className="obs-stat-label">{d.label}</span>
            <strong>{nf.format(deeds[d.key] ?? 0)}</strong>
          </div>
        ))}
      </div>
    </>
  )
}
