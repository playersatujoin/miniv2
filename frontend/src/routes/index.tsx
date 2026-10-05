import { useSuspenseQuery } from '@tanstack/react-query'
import { Link, createFileRoute, useNavigate } from '@tanstack/react-router'
import { useState, type FormEvent } from 'react'
import { previewUrl, type MapSummary } from '../api/client'
import { mapsQuery, useCreateMap, useDeleteMap } from '../api/queries'

export const Route = createFileRoute('/')({
  loader: ({ context }) => context.queryClient.ensureQueryData(mapsQuery),
  component: HomePage,
})

function HomePage() {
  const { data: maps } = useSuspenseQuery(mapsQuery)

  return (
    <main className="page">
      <section className="hero">
        <div>
          <h1>Akuarium Kehidupan</h1>
          <p>
            Setiap dunia dimulai dari dua orang, Adam dan Hawa. Keturunan mereka punya otak jaringan saraf sendiri yang
            memutuskan segalanya: makan, minum, mencari pasangan, mengumpulkan bahan, membuat alat, membangun rumah,
            menemukan unsur-unsur bumi — juga berbagi, mencuri, atau menyerang. Anak mewarisi otak kedua orang tuanya
            plus mutasi, jadi peradaban mereka berevolusi sendiri di server Go, walau kamu tidak sedang melihat. Kamu
            cukup mengamati.
          </p>
        </div>
        <CreateMapForm />
      </section>

      <h2 className="section-title">Peta kamu ({maps.length})</h2>
      {maps.length === 0 ? (
        <p className="muted">Belum ada peta. Buat satu di atas.</p>
      ) : (
        <div className="map-grid">
          {maps.map((m) => (
            <MapCard key={m.id} map={m} />
          ))}
        </div>
      )}
    </main>
  )
}

const SIZES = [32, 64, 96, 128, 192, 256]

function CreateMapForm() {
  const navigate = useNavigate()
  const create = useCreateMap()
  const [name, setName] = useState('Pulau Baru')
  const [width, setWidth] = useState(96)
  const [height, setHeight] = useState(96)
  const [seed, setSeed] = useState('')

  const onSubmit = (e: FormEvent) => {
    e.preventDefault()
    create.mutate(
      { name, width, height, seed: seed === '' ? undefined : Number(seed) },
      { onSuccess: (m) => navigate({ to: '/maps/$mapId', params: { mapId: m.id }, search: { mode: 'watch' } }) },
    )
  }

  return (
    <form className="card create-form" onSubmit={onSubmit}>
      <h2>Buat peta baru</h2>
      <label>
        Nama
        <input value={name} onChange={(e) => setName(e.target.value)} maxLength={64} required />
      </label>
      <div className="row">
        <label>
          Lebar
          <select value={width} onChange={(e) => setWidth(Number(e.target.value))}>
            {SIZES.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
        </label>
        <label>
          Tinggi
          <select value={height} onChange={(e) => setHeight(Number(e.target.value))}>
            {SIZES.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
        </label>
      </div>
      <label>
        Seed <small className="muted">(kosong = acak)</small>
        <div className="row">
          <input
            inputMode="numeric"
            pattern="\d*"
            placeholder="acak"
            value={seed}
            onChange={(e) => setSeed(e.target.value.replace(/\D/g, '').slice(0, 9))}
          />
          <button type="button" className="ghost" onClick={() => setSeed(String(Math.floor(Math.random() * 1e9)))}>
            🎲
          </button>
        </div>
      </label>
      {create.error && <p className="error">{create.error.message}</p>}
      <button type="submit" className="primary" disabled={create.isPending}>
        {create.isPending ? 'Membuat…' : 'Generate peta'}
      </button>
    </form>
  )
}

const dateFmt = new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short' })

function MapCard({ map }: { map: MapSummary }) {
  const del = useDeleteMap()

  return (
    <article className="card map-card">
      <Link to="/maps/$mapId" params={{ mapId: map.id }} search={{ mode: 'watch' }} className="preview">
        <img src={previewUrl(map, 2)} alt={`Pratinjau ${map.name}`} loading="lazy" />
      </Link>
      <div className="map-card-body">
        <h3>{map.name}</h3>
        <p className="muted">
          {map.width}×{map.height} · seed {map.seed}
        </p>
        <p className="muted small">Diubah {dateFmt.format(new Date(map.updatedAt))}</p>
        <div className="actions">
          <Link to="/maps/$mapId" params={{ mapId: map.id }} search={{ mode: 'watch' }} className="button primary">
            👁 Amati
          </Link>
          <Link to="/maps/$mapId" params={{ mapId: map.id }} search={{ mode: 'play' }} className="button" title="Main" aria-label="Main">
            ▶
          </Link>
          <Link to="/maps/$mapId" params={{ mapId: map.id }} search={{ mode: 'edit' }} className="button" title="Edit" aria-label="Edit">
            ✎
          </Link>
          <button
            type="button"
            className="ghost danger"
            disabled={del.isPending}
            onClick={() => confirm(`Hapus "${map.name}"?`) && del.mutate(map.id)}
          >
            Hapus
          </button>
        </div>
      </div>
    </article>
  )
}
