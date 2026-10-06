import { useQuery } from '@tanstack/react-query'
import { ApiError } from '../../api/client'
import { villagesQuery } from '../../sim/api'
import type { EngineInfo, SimInfo, TradeRecord } from '../../sim/protocol'
import { windArrow, windCompass, windWord } from '../../game/wind'
import { FoodSharing, VillageStores } from './FoodSharing'
import { PersonLink } from './Inspector'
import { formatClockShort, nf, useSecondsPerYear } from './format'
import {
  attitudeLabel,
  leaderTrust,
  sortAttitudes,
  sortVillages,
  trustWord,
  villageCamera,
  yearOf,
  yearsBetween,
  type VillageDetail,
} from './villageList'

/** A missing endpoint (older server) won't appear by retrying; other errors might. */
const retry = (count: number, err: Error) => !(err instanceof ApiError && err.status === 404) && count < 2

// Diverging pair as the Inspector's reputation: blue = trusted, red = distrusted.
const TRUST = { pos: '#3987e5', neg: '#e66767' }

export type VillagesProps = {
  mapId: string
  info: SimInfo | undefined
  onSelect: (id: number) => void
  /** Points the camera (2D or 3D) at a spot, `across` tiles wide. */
  onLocate?: (view: { x: number; y: number; across: number }) => void
}

/** The Desa tab: how the island's people live together, then each village. */
export function Villages({ mapId, info, onSelect, onLocate }: VillagesProps) {
  const q = useQuery({ ...villagesQuery(mapId), retry })
  const villages = sortVillages(q.data as VillageDetail[] | undefined)
  const missing = q.isError && q.error instanceof ApiError && q.error.status === 404

  return (
    <div className="vil-grid">
      <SocietySummary engine={info?.engine} />
      <FoodSharing engine={info?.engine} />
      {info?.engine?.exchange?.recent?.length ? <RecentTrades exchange={info.engine.exchange} onSelect={onSelect} /> : null}
      {missing ? (
        <p className="small muted">Server ini belum mengenal desa.</p>
      ) : q.isError ? (
        <p className="obs-error small">Gagal memuat desa: {q.error.message}</p>
      ) : !q.data ? (
        <p className="small muted">Memuat desa…</p>
      ) : villages.length === 0 ? (
        <p className="obs-hint small">
          Belum ada desa. Sebuah desa terbentuk sendiri bila beberapa rumah berdiri berdekatan; warganya lalu memilih orang yang paling
          mereka percayai sebagai pemimpin.
        </p>
      ) : (
        villages.map((v) => <VillageCard key={v.id} v={v} now={info?.time ?? 0} onSelect={onSelect} onLocate={onLocate} />)
      )}
    </div>
  )
}

function Stat({ label, value, title }: { label: string; value: string; title?: string }) {
  return (
    <div className="obs-stat" title={title}>
      <span className="obs-stat-label">{label}</span>
      <strong>{value}</strong>
    </div>
  )
}

const n = (v: number | undefined | null) => (v == null ? '—' : nf.format(v))

const STIMULUS_LABELS: Record<string, string> = {
  corpse: 'jenazah',
  fight: 'perkelahian',
  theft: 'pencurian',
  predator: 'serangan hewan',
  fire: 'api',
  collapse: 'bangunan runtuh',
  drowning: 'orang tenggelam',
  fall: 'orang jatuh',
  carcass: 'bangkai buruan',
}

/** Talks, trades, rumours, fire, water and wind across the island right now (SimInfo.engine). */
function SocietySummary({ engine: e }: { engine: EngineInfo | undefined }) {
  if (!e) {
    return (
      <section className="obs-section soc-summary">
        <h3 className="obs-title">Masyarakat</h3>
        <p className="small muted">Server ini belum melaporkan obrolan, barter, api, dan desa.</p>
      </section>
    )
  }
  const x = e.exchange ?? { talks: 0, trades: 0, rumors: 0 }
  const f = e.fire ?? { burning: 0 }
  const w = e.water ?? { swimming: 0, rafting: 0, climbing: 0 }
  const st = e.stimuli ?? { active: 0 }
  const byKind = Object.entries(st.byKind ?? {})
    .filter(([, c]) => c > 0)
    .sort((a, b) => b[1] - a[1])
    .map(([k, c]) => `${STIMULUS_LABELS[k] ?? k} ${nf.format(c)}`)
    .join(', ')
  const blowing = f.wind != null && (f.storm || f.wind >= 0.1)
  const wind = f.wind != null ? `${blowing ? `${windArrow(f.windDir ?? 0)} ` : ''}${windWord(f.wind, f.storm)}` : '—'
  const allTime = [
    w.drowned ? `${nf.format(w.drowned)} tenggelam` : '',
    w.falls ? `${nf.format(w.falls)} terjatuh` : '',
    f.burned ? `${nf.format(f.burned)} terbakar` : '',
  ].filter(Boolean)
  const fires = [
    f.started ? `${nf.format(f.started)} kebakaran` : '',
    f.lightning ? `${nf.format(f.lightning)} karena petir` : '',
    f.hearth ? `${nf.format(f.hearth)} dari tungku` : '',
    f.buildings ? `${nf.format(f.buildings)} bangunan terbakar` : '',
    f.crops ? `${nf.format(f.crops)} petak ladang hangus` : '',
    f.storms ? `${nf.format(f.storms)} badai` : '',
  ].filter(Boolean)

  return (
    <section className="obs-section soc-summary" aria-label="Ringkasan masyarakat">
      <h3 className="obs-title">Masyarakat</h3>
      <div className="obs-stat-grid">
        <Stat label="🗣 Obrolan" value={n(x.talks)} title="Obrolan dua orang sepanjang masa (keduanya mau)" />
        <Stat
          label="🔁 Barter"
          value={n(x.trades)}
          title={x.units ? `${nf.format(x.units)} satuan barang berpindah tangan sepanjang masa` : 'Tukar-menukar barang sepanjang masa'}
        />
        <Stat
          label="📣 Kabar"
          value={n(x.rumors)}
          title={`Ingatan yang diteruskan dalam obrolan${x.opinions ? `, ditambah ${nf.format(x.opinions)} pendapat tentang orang ketiga` : ''}`}
        />
        <Stat
          label="🔥 Api menyala"
          value={n(f.burning)}
          title={f.scorched ? `${nf.format(f.scorched)} petak tanah hangus masih terlihat` : 'Petak yang sedang terbakar'}
        />
        <Stat
          label="🌬 Angin"
          value={wind}
          title={`${blowing ? `Bertiup ke ${windCompass(f.windDir ?? 0)}` : 'Hampir tidak berangin'}${f.wind != null ? ` (kekuatan ${Math.round(f.wind * 100)}%)` : ''}. Angin yang sama menyebarkan api.`}
        />
        <Stat
          label="⚡ Rangsangan"
          value={n(st.active)}
          title={byKind ? `Hal mencolok yang masih bisa dilihat orang: ${byKind}` : 'Hal mencolok yang masih bisa dilihat orang'}
        />
        <Stat label="🌊 Mengarungi" value={n(w.wading)} title="Orang yang berjalan di air dangkal" />
        <Stat
          label="🏊 Berenang"
          value={n(w.swimming)}
          title={w.struggling ? `${nf.format(w.struggling)} sedang kepayahan` : 'Orang yang berenang di air dalam'}
        />
        <Stat
          label="🛶 Berakit"
          value={n(w.rafting)}
          title={w.rafts ? `${nf.format(w.rafts)} rakit pernah dibuat` : 'Orang di atas rakit'}
        />
        <Stat
          label="🧗 Memanjat"
          value={n(w.climbing)}
          title={w.fallen ? `${nf.format(w.fallen)} baru saja terjatuh` : 'Orang yang memanjat lereng curam'}
        />
        {e.villages && (
          <Stat
            label="🏘 Desa"
            value={n(e.villages.villages)}
            title={e.villages.villagers != null ? `${nf.format(e.villages.villagers)} orang tinggal di desa` : undefined}
          />
        )}
        {e.villages?.leaderChanges != null && (
          <Stat label="🚩 Pergantian" value={n(e.villages.leaderChanges)} title="Berapa kali desa berganti pemimpin" />
        )}
      </div>
      {(x.talking != null || x.trading != null || x.waiting != null) && (
        <p className="obs-vitals small">
          <span>
            Sedang mengobrol <strong>{n(x.talking)}</strong>
          </span>
          <span>
            bertukar barang <strong>{n(x.trading)}</strong>
          </span>
          <span>
            menunggu jawaban <strong>{n(x.waiting)}</strong>
          </span>
        </p>
      )}
      {fires.length > 0 && <p className="small muted obs-note">Api sepanjang masa: {fires.join(' · ')}.</p>}
      {allTime.length > 0 && <p className="small muted obs-note">Meninggal karena air, lereng, dan api: {allTime.join(' · ')}.</p>}
    </section>
  )
}

/** An item id as the exchange's list names it, else the id with a capital. */
function itemName(id: string, names: Map<string, string>) {
  return names.get(id) ?? (id ? id[0].toUpperCase() + id.slice(1).replace(/_/g, ' ') : '?')
}

/** The latest barters (newest first) and the most traded goods. */
function RecentTrades({ exchange: x, onSelect }: { exchange: EngineInfo['exchange']; onSelect: (id: number) => void }) {
  const spy = useSecondsPerYear()
  const names = new Map((x.items ?? []).map((it) => [it.item, it.name]))
  const recent: TradeRecord[] = [...(x.recent ?? [])].reverse().slice(0, 8)
  const items = [...(x.items ?? [])].sort((a, b) => b.units - a.units).slice(0, 8)
  return (
    <section className="obs-section soc-trades-card">
      <h3 className="obs-title">Barter terakhir</h3>
      <ol className="soc-trades">
        {recent.map((t, i) => (
          <li key={`${t.time}-${t.a.id}-${i}`}>
            <span>
              <PersonLink who={t.a} none="?" onSelect={onSelect} /> memberi{' '}
              <strong>
                {nf.format(t.gaveN)} {itemName(t.gave, names)}
              </strong>{' '}
              kepada <PersonLink who={t.b} none="?" onSelect={onSelect} />, mendapat{' '}
              <strong>
                {nf.format(t.gotN)} {itemName(t.got, names)}
              </strong>
            </span>
            <small className="muted">{formatClockShort(t.time, spy)}</small>
          </li>
        ))}
      </ol>
      {items.length > 0 && (
        <>
          <h4 className="obs-subtitle">Paling sering ditukar</h4>
          <ul className="obs-chips">
            {items.map((it) => (
              <li key={it.item}>
                {it.name || it.item} <strong>×{nf.format(it.units)}</strong>
              </li>
            ))}
          </ul>
        </>
      )}
    </section>
  )
}

function VillageCard({
  v,
  now,
  onSelect,
  onLocate,
}: {
  v: VillageDetail
  now: number
  onSelect: (id: number) => void
  onLocate?: VillagesProps['onLocate']
}) {
  const spy = useSecondsPerYear()
  const trust = leaderTrust(v)
  const attitudes = sortAttitudes(v.attitudes)
  const color = `hsl(${Math.round(v.hue)} 70% 62%)`
  const age = yearsBetween(v.founded, now, spy)
  const shown = trust.trusting + trust.distrusting
  return (
    <section className="obs-section vil-card" style={{ borderTopColor: color }} aria-label={`Desa ${v.name}`}>
      <div className="vil-head">
        <i className="vil-swatch" style={{ background: color }} aria-hidden />
        <h3>{v.name}</h3>
        {onLocate && (
          <button
            type="button"
            className="obs-mini vil-locate"
            title={`Arahkan kamera ke desa ${v.name}`}
            aria-label={`Lihat desa ${v.name} di peta`}
            onClick={() => onLocate(villageCamera(v))}
          >
            📍 Lihat di peta
          </button>
        )}
      </div>
      <div className="obs-stat-grid">
        <Stat label="👥 Warga" value={n(v.people)} />
        <Stat label="🏠 Rumah" value={n(v.houses)} />
        {v.fields != null && <Stat label="🌾 Ladang" value={n(v.fields)} title="Petak yang ditanami keluarga desa ini" />}
      </div>
      <dl className="obs-kv">
        <dt>Pemimpin</dt>
        <dd>
          {v.leader ? (
            <>
              🚩 <PersonLink who={v.leader} none="—" onSelect={onSelect} />
              {v.leaderSince != null && <span className="muted"> · sejak {formatClockShort(v.leaderSince, spy)}</span>}
            </>
          ) : (
            <span className="muted">belum ada</span>
          )}
        </dd>
        <dt>Berdiri</dt>
        <dd>
          Tahun {nf.format(yearOf(v.founded, spy))}
          <span className="muted"> · {age > 0 ? `${nf.format(age)} tahun lalu` : 'baru saja'}</span>
        </dd>
        {v.area != null && (
          <>
            <dt>Luas wilayah</dt>
            <dd>{nf.format(v.area)} petak</dd>
          </>
        )}
        {v.leaders != null && (
          <>
            <dt title="Berapa orang yang pernah memimpin desa ini">Pernah dipimpin</dt>
            <dd>{nf.format(v.leaders)} orang</dd>
          </>
        )}
        {v.supporters != null && (
          <>
            <dt title="Warga dewasa yang paling mempercayai pemimpinnya dibanding orang lain">Pendukung</dt>
            <dd>{nf.format(v.supporters)} orang</dd>
          </>
        )}
      </dl>
      {v.leader && v.residents && trust.adults > 0 && (
        <div className="vil-trust">
          <h4 className="obs-subtitle">Kepercayaan warga pada pemimpin</h4>
          <div
            className="vil-trust-bar"
            role="img"
            aria-label={`${trust.trusting} percaya, ${trust.distrusting} curiga, ${trust.adults - shown} netral atau belum kenal, dari ${trust.adults} warga dewasa`}
          >
            <span style={{ width: `${(trust.trusting / trust.adults) * 100}%`, background: TRUST.pos }} />
            <span style={{ width: `${(trust.distrusting / trust.adults) * 100}%`, background: TRUST.neg }} />
          </div>
          <p className="small obs-note">
            <span style={{ color: TRUST.pos }}>■</span> {nf.format(trust.trusting)} percaya · <span style={{ color: TRUST.neg }}>■</span>{' '}
            {nf.format(trust.distrusting)} curiga · {nf.format(trust.adults - shown)} netral atau belum kenal
            {trust.mean != null && (
              <span className="muted">
                {' '}
                — rata-rata {trustWord(trust.mean)} ({trust.mean >= 0 ? '+' : ''}
                {trust.mean.toFixed(2)})
              </span>
            )}
          </p>
        </div>
      )}
      {attitudes.length > 0 && (
        <div className="vil-attitudes">
          <h4 className="obs-subtitle">Sikap terhadap desa lain</h4>
          <ul>
            {attitudes.map((a) => (
              <li key={a.id}>
                <span>{a.name}</span>
                <span
                  style={{ color: a.known === 0 ? undefined : a.trust >= 0 ? TRUST.pos : TRUST.neg }}
                  className={a.known === 0 ? 'muted' : undefined}
                >
                  {attitudeLabel(a.attitude)}
                  {a.known > 0 && (
                    <small className="muted">
                      {' '}
                      ({a.trust >= 0 ? '+' : ''}
                      {a.trust.toFixed(2)}, kenal {nf.format(a.known)} orang)
                    </small>
                  )}
                </span>
              </li>
            ))}
          </ul>
        </div>
      )}
      <VillageStores v={v} />
    </section>
  )
}
