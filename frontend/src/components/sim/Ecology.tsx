import { useQuery } from '@tanstack/react-query'
import { ApiError } from '../../api/client'
import { ecologyQuery } from '../../sim/api'
import {
  ENSO_LABELS,
  MONTH_NAMES,
  SEASON_LABELS,
  SPECIES,
  type CropView,
  type Diet,
  type EcoPoint,
  type Enso,
  type Season,
  type SpeciesView,
} from '../../sim/protocol'
import { EcologyYears } from './EcologyYears'
import { LineChart, type LineSeries } from './LineChart'
import { formatClock, formatClockShort, nf, nf1, pct, useSecondsPerYear } from './format'

/** A missing endpoint (older server) won't appear by retrying; other errors might. */
const retry = (count: number, err: Error) => !(err instanceof ApiError && err.status === 404) && count < 2

const ENSO_HINT: Record<Enso, string> = {
  netral: 'Tahun biasa.',
  el_nino: 'El Niño: kemarau panjang dan kering, tanaman mudah layu dan gagal panen.',
  la_nina: 'La Niña: tahun lebih basah dari biasa, rawan banjir di bulan Februari.',
}

/** "🌧 Musim hujan" and, in an El Niño or La Niña year, its badge. Text carries the meaning. */
export function SeasonBadges({ season, enso, title }: { season: Season | undefined; enso: Enso | undefined; title?: string }) {
  return (
    <>
      {season && (
        <span className={`eco-badge eco-season-${season}`} title={title}>
          <span aria-hidden="true">{season === 'hujan' ? '🌧' : '☀'}</span> {SEASON_LABELS[season] ?? season}
        </span>
      )}
      {enso && enso !== 'netral' && (
        <span className={`eco-badge eco-enso-${enso}`} title={ENSO_HINT[enso]}>
          {ENSO_LABELS[enso] ?? enso}
        </span>
      )}
    </>
  )
}

const DIET_LABELS: Record<Diet, string> = {
  grazer: 'pemakan rumput dan daun',
  forager: 'pemakan buah, umbi dan biji',
  predator: 'pemangsa',
}

// Five lines in the first five categorical slots, in species order. Validated
// on the panel surface (#172234, dark, adjacent pairs): worst CVD ΔE 8.4,
// normal vision 19.3, all ≥ 3:1. The legend carries names and latest counts.
const ANIMAL_COLORS = ['#3987e5', '#d95926', '#199e70', '#c98500', '#d55181']
const ANIMAL_SHORT = ['Rusa', 'Babi', 'Ayam', 'Kerbau', 'Harimau']

const ANIMAL_SERIES: LineSeries<EcoPoint>[] = SPECIES.map((sp, i) => ({
  key: sp.id,
  label: ANIMAL_SHORT[i] ?? sp.name,
  color: ANIMAL_COLORS[i],
  value: (p: EcoPoint) => p.animals?.[i] ?? 0,
}))

const whole = (v: number) => nf.format(Math.round(v))

/** The land: seasons and weather, food, animals and crops. */
export function Ecology({ mapId }: { mapId: string }) {
  const q = useQuery({ ...ecologyQuery(mapId), retry })
  const spy = useSecondsPerYear()

  if (!q.data) {
    const missing = q.error instanceof ApiError && q.error.status === 404
    return (
      <section className="obs-section">
        <h3 className="obs-title">Ekologi</h3>
        <p className={`small ${q.isError && !missing ? 'obs-error' : 'muted'}`}>
          {q.isPending
            ? 'Memuat ekologi…'
            : missing
              ? 'Data ekologi belum tersedia di server ini.'
              : `Gagal memuat ekologi: ${q.error?.message ?? ''}`}
        </p>
      </section>
    )
  }

  const e = q.data
  const month = MONTH_NAMES[Math.min(11, Math.max(0, (e.month ?? 1) - 1))]
  const food = e.food ?? { carried: 0, stored: 0, granary: 0 }
  const species = e.species ?? []
  const wild = species.reduce((s, sp) => s + sp.wild, 0)
  const tame = species.reduce((s, sp) => s + sp.tame, 0)

  return (
    <>
      <section className="obs-section" aria-labelledby="eco-title">
        <div className="obs-section-head">
          <h3 className="obs-title" id="eco-title">
            Musim &amp; cuaca
          </h3>
          <span className="small muted">
            {month} · tahun {nf.format(e.year)}
          </span>
        </div>
        <p className="eco-badges">
          <SeasonBadges season={e.season} enso={e.enso} />
          {e.enso === 'netral' && <span className="eco-badge">ENSO netral</span>}
        </p>
        <div className="obs-meters">
          <Meter
            label="Curah hujan"
            value={e.rain / 2}
            mark={0.5}
            text={`×${nf1.format(e.rain)} dari rata-rata`}
            color="#3987e5"
          />
          <Meter label="Kelembapan tanah" value={e.moisture} text={pct(e.moisture)} color="#199e70" />
          <Meter label="Hutan tersisa" value={e.forest} text={pct(e.forest)} color="#008300" />
        </div>
        <p className="small muted eco-note">
          {ENSO_HINT[e.enso] ?? ''} Musim hujan sekitar Oktober–April, kemarau April–Oktober. Garis putih pada curah hujan =
          rata-rata setahun.
        </p>
      </section>

      <section className="obs-section" aria-labelledby="eco-food-title">
        <div className="obs-section-head">
          <h3 className="obs-title" id="eco-food-title">
            Persediaan pangan
          </h3>
          <span className="small muted">{nf.format(food.carried + food.stored + food.granary)} satuan</span>
        </div>
        <div className="obs-stat-grid">
          <div className="obs-stat">
            <span className="obs-stat-label">🎒 Dibawa</span>
            <strong>{nf.format(food.carried)}</strong>
          </div>
          <div className="obs-stat">
            <span className="obs-stat-label">🏠 Di rumah</span>
            <strong>{nf.format(food.stored)}</strong>
          </div>
          <div className="obs-stat">
            <span className="obs-stat-label">🌾 Di lumbung</span>
            <strong>{nf.format(food.granary)}</strong>
          </div>
        </div>
        <p className="small muted eco-note">Satu satuan pangan kira-kira cukup untuk makan satu orang selama setahun.</p>
      </section>

      <section className="obs-section" aria-labelledby="eco-animals-title">
        <div className="obs-section-head">
          <h3 className="obs-title" id="eco-animals-title">
            Hewan
          </h3>
          <span className="small muted">
            {nf.format(wild)} liar · {nf.format(tame)} jinak
          </span>
        </div>
        <SpeciesTable species={species} />
        <LineChart
          title="Jumlah hewan dari waktu ke waktu"
          points={e.history ?? []}
          x={(p) => p.time}
          xLabel={(p) => formatClockShort(p.time, spy)}
          tooltipTitle={(p) => formatClock(p.time, spy)}
          series={ANIMAL_SERIES}
          fmt={whole}
          floor={10}
          extra={(p) => (
            <span>
              {nf.format(p.tame)} di antaranya jinak · hutan {pct(p.forest)}
            </span>
          )}
          empty="Grafik muncul setelah beberapa detik simulasi berjalan."
        />
        <p className="small muted eco-note">Liar dan jinak dijumlahkan; kerbau, babi dan ayam yang jinak dipelihara keluarga.</p>
      </section>

      <section className="obs-section" aria-labelledby="eco-crops-title">
        <div className="obs-section-head">
          <h3 className="obs-title" id="eco-crops-title">
            Tanaman
          </h3>
          <span className="small muted">
            {nf.format(e.plots)} petak · {nf.format(e.ripe)} siap panen
          </span>
        </div>
        <CropTable crops={e.crops ?? []} />
        {e.plots === 0 && (
          <p className="small muted eco-note">
            Belum ada yang menanam. Pertanian ditemukan saat seseorang pertama kali memanen tanaman yang sengaja ia tanam.
          </p>
        )}
      </section>

      <EcologyYears years={e.years ?? []} />
    </>
  )
}

/** A 0–1 bar like the inspector's, with an optional reference tick (e.g. the yearly average). */
function Meter({ label, value, text, color, mark }: { label: string; value: number; text: string; color: string; mark?: number }) {
  const v = Math.min(1, Math.max(0, value))
  return (
    <div className="obs-meter">
      <div className="obs-meter-head">
        <span>{label}</span>
        <span className="obs-meter-value">{text}</span>
      </div>
      <div
        className="obs-meter-track eco-meter-track"
        style={{ background: `color-mix(in srgb, ${color} 22%, transparent)` }}
        role="meter"
        aria-label={`${label}: ${text}`}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={Math.round(v * 100)}
      >
        <div className="obs-meter-fill" style={{ width: `${v * 100}%`, background: color }} />
        {mark != null && <i className="obs-skill-threshold" style={{ left: `calc(${mark * 100}% - 1px)` }} aria-hidden />}
      </div>
    </div>
  )
}

function SpeciesTable({ species }: { species: SpeciesView[] }) {
  if (species.length === 0) return <p className="small muted">Belum ada data hewan.</p>
  return (
    <table className="obs-table eco-table">
      <thead>
        <tr>
          <th>Hewan</th>
          <th title="Hidup liar di pulau">Liar</th>
          <th title="Dipelihara keluarga">Jinak</th>
          <th title="Jumlah saat dunia dimulai">Awal</th>
          <th title="Dibunuh manusia, sepanjang masa">Diburu</th>
        </tr>
      </thead>
      <tbody>
        {species.map((sp) => (
          <tr key={sp.id} className={sp.extinct ? 'eco-extinct' : undefined}>
            <td title={DIET_LABELS[sp.diet] ?? sp.diet}>
              {sp.name}
              {sp.extinct && <span className="civ-tag civ-tag-lost">punah</span>}
            </td>
            <td>{nf.format(sp.wild)}</td>
            <td title={sp.domestic ? `Jinak disebut ${sp.domestic}` : 'Tidak bisa dijinakkan'}>
              {sp.domestic ? nf.format(sp.tame) : '—'}
            </td>
            <td>{nf.format(sp.start)}</td>
            <td>{nf.format(sp.hunted)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}

function CropTable({ crops }: { crops: CropView[] }) {
  if (crops.length === 0) return <p className="small muted">Belum ada data tanaman.</p>
  return (
    <table className="obs-table eco-table">
      <thead>
        <tr>
          <th>Tanaman</th>
          <th>Petak</th>
          <th>Siap panen</th>
          <th title="Tahun dari tanam sampai panen pertama">Umur panen</th>
        </tr>
      </thead>
      <tbody>
        {crops.map((c) => (
          <tr key={c.item}>
            <td>
              {c.name}
              {c.perennial && (
                <span className="civ-tag eco-tag-perennial" title="Berbuah berulang kali setelah panen pertama">
                  menahun
                </span>
              )}
            </td>
            <td>{nf.format(c.plots)}</td>
            <td>{nf.format(c.ripe)}</td>
            <td>{nf1.format(c.growYears)} th</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
