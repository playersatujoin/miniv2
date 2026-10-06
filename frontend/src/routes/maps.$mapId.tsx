import { useQuery, useSuspenseQuery } from '@tanstack/react-query'
import { Link, createFileRoute, useBlocker, type ErrorComponentProps } from '@tanstack/react-router'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { ApiError, type GameMap, type Point, type TileSet } from '../api/client'
import { geologyQuery, mapQuery, tilesQuery, useSaveMap } from '../api/queries'
import { Palette } from '../components/Palette'
import { ObserverPanel } from '../components/sim/ObserverPanel'
import { SimStatus } from '../components/sim/SimStatus'
import { GameCanvas } from '../game/GameCanvas'
import { GeologyLegend, useGeologyOverlay } from '../game/GeologyLegend'
import type { Brush, EngineEvents, GameEngine, HoverInfo, Mode, Tool } from '../game/engine'
import { minedOutQuery } from '../sim/api'
import { CROPS, PLOT_FLAG, type FieldPlot } from '../sim/protocol'
import { useSimStream, type StreamStatus } from '../sim/stream'

type MapSearch = {
  mode: Mode
  /** Observed creature in watch mode, kept in the URL so reloads and links keep it. */
  creature?: number
}

const MODES: Mode[] = ['watch', 'play', 'edit']

export const Route = createFileRoute('/maps/$mapId')({
  validateSearch: (search: Record<string, unknown>): MapSearch => {
    const creature = Number(search.creature)
    return {
      mode: MODES.includes(search.mode as Mode) ? (search.mode as Mode) : 'watch',
      creature: Number.isInteger(creature) && creature > 0 ? creature : undefined,
    }
  },
  loader: ({ context: { queryClient }, params }) =>
    Promise.all([
      queryClient.ensureQueryData(mapQuery(params.mapId)),
      queryClient.ensureQueryData(tilesQuery),
    ]),
  pendingComponent: () => <main className="page center muted">Memuat peta…</main>,
  errorComponent: MapError,
  component: MapPage,
})

const STREAM_LABELS: Record<StreamStatus, string> = {
  connecting: 'Menghubungkan…',
  live: 'Langsung',
  error: 'Terputus',
}

function MapPage() {
  const { mapId } = Route.useParams()
  const { mode, creature } = Route.useSearch()
  const { data: map } = useSuspenseQuery(mapQuery(mapId))
  const { data: tiles } = useSuspenseQuery(tilesQuery)
  // Keyed so local editor state resets when navigating to another map.
  return <MapEditor key={mapId} map={map} tiles={tiles} mode={mode} selectedId={creature ?? null} />
}

type EditorProps = { map: GameMap; tiles: TileSet; mode: Mode; selectedId: number | null }

function MapEditor({ map, tiles, mode, selectedId }: EditorProps) {
  const navigate = Route.useNavigate()
  const save = useSaveMap(map.id)
  const geology = useQuery(geologyQuery(map.id, map.updatedAt))
  const [showGeology, setShowGeology] = useGeologyOverlay()
  const geologyShown = showGeology && !!geology.data
  // Old pits only exist in the living world, so only ask while watching it.
  const mined = useQuery({ ...minedOutQuery(map.id), enabled: mode === 'watch' })
  const engineRef = useRef<GameEngine | null>(null)

  const [tool, setTool] = useState<Tool>(() => ({
    kind: 'paint',
    layer: 'ground',
    id: tiles.ground.find((t) => t.key === 'grass')?.id ?? 0,
  }))
  const [brush, setBrush] = useState<Brush>(1)
  const [name, setName] = useState(map.name)
  const [tilesDirty, setTilesDirty] = useState(false)
  const [history, setHistory] = useState({ canUndo: false, canRedo: false })
  const [hover, setHover] = useState<HoverInfo | null>(null)
  const [playerTile, setPlayerTile] = useState<Point | null>(null)
  const [following, setFollowing] = useState(selectedId !== null)

  const dirty = tilesDirty || name.trim() !== map.name

  const streamStatus = useSimStream(map.id, mode === 'watch', {
    onFrame: (frame) => engineRef.current?.setCreatureFrame(frame),
    onStructures: (msg) => engineRef.current?.setStructures(msg.structures),
    onFields: (msg) => engineRef.current?.setFields(msg.plots),
  })

  // Selecting a creature (on the map or in the panel) also points the camera at it.
  const select = useCallback(
    (id: number | null) => {
      setFollowing(id !== null)
      navigate({ search: (prev) => ({ ...prev, creature: id ?? undefined }), replace: true })
    },
    [navigate],
  )

  useBlocker({
    // Switching modes only changes search params, so only block real page changes.
    shouldBlockFn: ({ current, next }) =>
      current.pathname !== next.pathname && !confirm('Ada perubahan yang belum disimpan. Tinggalkan halaman ini?'),
    enableBeforeUnload: () => dirty,
    disabled: !dirty,
  })

  const events = useMemo<EngineEvents>(
    () => ({
      onDirtyChange: setTilesDirty,
      onHistoryChange: (canUndo, canRedo) => setHistory({ canUndo, canRedo }),
      onHover: setHover,
      onPlayerTile: setPlayerTile,
      onSelectCreature: select,
      onFollowChange: setFollowing,
    }),
    [select],
  )

  const onSave = useCallback(async () => {
    const engine = engineRef.current
    if (!engine || save.isPending) return
    const snap = engine.snapshot()
    try {
      const saved = await save.mutateAsync({ name: name.trim(), spawn: snap.spawn, layers: snap.layers })
      engine.markSaved(snap.marker, saved.updatedAt)
      setName(saved.name)
    } catch {
      // Shown via save.error.
    }
  }, [name, save])

  const onDiscard = () => {
    if (!confirm('Buang semua perubahan yang belum disimpan?')) return
    engineRef.current?.setMap(map)
    setName(map.name)
    save.reset()
  }

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.code === 'KeyS') {
        e.preventDefault()
        if (mode === 'edit' && dirty) onSave()
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [mode, dirty, onSave])

  const setMode = (next: Mode) => navigate({ search: (prev) => ({ ...prev, mode: next }), replace: true })

  return (
    <main className={`map-page mode-${mode}`}>
      <div className="toolbar">
        <Link to="/" className="button ghost">
          ← Peta
        </Link>
        {mode === 'edit' ? (
          <input className="name-input" value={name} maxLength={64} onChange={(e) => setName(e.target.value)} />
        ) : (
          <h1 className="map-title">{map.name}</h1>
        )}

        <div className="segmented">
          <button type="button" className={mode === 'watch' ? 'active' : ''} onClick={() => setMode('watch')}>
            👁 Amati
          </button>
          <button type="button" className={mode === 'play' ? 'active' : ''} onClick={() => setMode('play')}>
            ▶ Main
          </button>
          <button type="button" className={mode === 'edit' ? 'active' : ''} onClick={() => setMode('edit')}>
            ✎ Edit
          </button>
        </div>

        {mode === 'watch' && (
          <>
            <span
              className={`stream-badge ${streamStatus}`}
              title={`Status aliran simulasi dari server: ${STREAM_LABELS[streamStatus]}`}
            >
              <span aria-hidden="true">●</span>
              <span className="stream-text"> {STREAM_LABELS[streamStatus]}</span>
            </span>
            <SimStatus mapId={map.id} />
          </>
        )}

        <div className="spacer" />

        <button
          type="button"
          className={`geo-toggle ${geologyShown ? 'active' : ''}`}
          aria-pressed={geologyShown}
          disabled={!geology.data}
          title={geology.data ? 'Tampilkan peta geologi: jenis batuan, gunung api, sungai, dan endapan' : 'Data geologi belum tersedia'}
          onClick={() => setShowGeology(!showGeology)}
        >
          🪨 <span className="btn-label">Geologi</span>
        </button>

        <div className="segmented">
          <button type="button" title="Perkecil (-)" onClick={() => engineRef.current?.zoomBy(-1)}>
            −
          </button>
          <button type="button" title="Perbesar (+)" onClick={() => engineRef.current?.zoomBy(1)}>
            +
          </button>
        </div>

        {mode === 'watch' ? null : mode === 'play' ? (
          <button type="button" onClick={() => engineRef.current?.respawn()}>
            ⟲ Respawn
          </button>
        ) : (
          <>
            <div className="segmented">
              <button
                type="button"
                title="Undo (Ctrl+Z)"
                disabled={!history.canUndo}
                onClick={() => engineRef.current?.undo()}
              >
                ↶
              </button>
              <button
                type="button"
                title="Redo (Ctrl+Y)"
                disabled={!history.canRedo}
                onClick={() => engineRef.current?.redo()}
              >
                ↷
              </button>
            </div>
            <button type="button" className="ghost" disabled={!dirty} onClick={onDiscard}>
              Buang
            </button>
            <button type="button" className="primary" disabled={!dirty || save.isPending} onClick={onSave}>
              {save.isPending ? 'Menyimpan…' : dirty ? 'Simpan' : 'Tersimpan ✓'}
            </button>
          </>
        )}
      </div>

      {save.error && <div className="banner error">Gagal menyimpan: {save.error.message}</div>}

      <div className="workspace">
        <div className="stage-wrap">
          <GameCanvas
            map={map}
            geology={geology.data}
            geologyOverlay={geologyShown}
            minedOut={mode === 'watch' ? mined.data : undefined}
            tiles={tiles}
            mode={mode}
            tool={tool}
            brush={brush}
            selectedId={selectedId}
            following={following}
            engineRef={engineRef}
            events={events}
          />
          <div className="hint">
            {mode === 'watch' ? (
              <>
                <kbd>klik</kbd> makhluk atau bangunan untuk mengamati · <kbd>drag</kbd>/<kbd>WASD</kbd> geser ·{' '}
                <kbd>scroll</kbd> zoom
              </>
            ) : mode === 'play' ? (
              <>
                <kbd>WASD</kbd>/<kbd>←↑↓→</kbd> jalan · <kbd>Shift</kbd> lari · <kbd>scroll</kbd> zoom
              </>
            ) : (
              <>
                <kbd>klik kiri</kbd> cat · <kbd>klik kanan</kbd> geser · <kbd>WASD</kbd> kamera · <kbd>Ctrl+Z</kbd>/
                <kbd>Ctrl+Y</kbd> undo/redo · <kbd>Ctrl+S</kbd> simpan
              </>
            )}
          </div>
          {geologyShown && geology.data && <GeologyLegend geology={geology.data} />}
          <div className="hud">
            {hover ? (
              <span>
                ({hover.x}, {hover.y}) {hover.ground?.name}
                {hover.object && hover.object.id !== 0 && <> · {hover.object.name}</>}
                {hover.rock && hover.rock.id !== 0 && <span className="hud-rock"> · {hover.rock.name}</span>}
                {hover.deposits.map(({ item, model }) => (
                  <span key={item.id} className="hud-deposit">
                    {' '}
                    · ⛏ {item.name}
                    {item.formula && <small> {item.formula}</small>}
                    {model && <small> — {model.name}</small>}
                  </span>
                ))}
                {hover.minedOut && <span className="hud-mined"> · bekas tambang (habis digali)</span>}
                {hover.stump && <span className="hud-mined"> · tunggul (pohon ditebang)</span>}
                {hover.plot && <span className="hud-plot"> · 🌱 {plotText(hover.plot)}</span>}
                {(hover.ground?.solid || hover.object?.solid) && <em> — blok</em>}
              </span>
            ) : mode === 'play' && playerTile ? (
              <span>
                Posisi: ({playerTile.x}, {playerTile.y})
              </span>
            ) : (
              <span>
                {map.width}×{map.height} tile · seed {map.seed}
              </span>
            )}
          </div>
        </div>
        {mode === 'edit' && (
          <Palette tiles={tiles} tool={tool} brush={brush} onTool={setTool} onBrush={setBrush} />
        )}
        {mode === 'watch' && (
          <ObserverPanel
            mapId={map.id}
            selectedId={selectedId}
            following={following}
            onSelect={select}
            onFollow={setFollowing}
          />
        )}
      </div>
    </main>
  )
}

const PLOT_STAGES = ['bibit', 'tumbuh', 'besar', 'siap panen']

/** "Padi (siap panen, sawah)" for the HUD. */
function plotText(p: FieldPlot) {
  const notes = [PLOT_STAGES[p.stage] ?? '']
  if (p.flags & PLOT_FLAG.withered) notes.push('layu kekeringan')
  if (p.flags & PLOT_FLAG.irrigated) notes.push(p.crop === 0 ? 'sawah' : 'diairi')
  if (p.flags & PLOT_FLAG.farmland) notes.push('ladang')
  if (p.flags & PLOT_FLAG.manured) notes.push('berpupuk')
  return `${CROPS[p.crop]?.name ?? 'Tanaman'} (${notes.join(', ')})`
}

function MapError({ error }: ErrorComponentProps) {
  const notFound = error instanceof ApiError && error.status === 404
  return (
    <main className="page center">
      <h2>{notFound ? 'Peta tidak ditemukan' : 'Gagal memuat peta'}</h2>
      {!notFound && <p className="error">{error instanceof Error ? error.message : String(error)}</p>}
      <Link to="/">Kembali ke daftar peta</Link>
    </main>
  )
}
