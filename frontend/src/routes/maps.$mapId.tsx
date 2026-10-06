import { useQuery, useSuspenseQuery } from '@tanstack/react-query'
import { Link, createFileRoute, useBlocker, type ErrorComponentProps } from '@tanstack/react-router'
import { Component, Suspense, lazy, useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { ApiError, type GameMap, type Point, type TileSet } from '../api/client'
import { geologyQuery, mapQuery, reliefQuery, tilesQuery, useSaveMap } from '../api/queries'
import { Palette } from '../components/Palette'
import { CinematicButton, CinematicSubjectChip, type CinematicSubject } from '../components/sim/CinematicButton'
import { ObserverPanel } from '../components/sim/ObserverPanel'
import { ReplayOverlay } from '../components/sim/ReplayOverlay'
import type { ReplaySinks } from '../components/sim/replayPlayer'
import { SimStatus } from '../components/sim/SimStatus'
import { DEFAULT_SECONDS_PER_YEAR } from '../components/sim/format'
import { GameCanvas } from '../game/GameCanvas'
import type { World3D, World3DEvents } from '../game3d/World3D'
import { GeologyLegend, useGeologyOverlay } from '../game/GeologyLegend'
import {
  snapAcross,
  type Brush,
  type EngineEvents,
  type GameEngine,
  type HoverInfo,
  type MapView,
  type Mode,
  type Tool,
} from '../game/engine'
import { minedOutQuery, simInfoQuery } from '../sim/api'
import { LiveWorld } from '../sim/live'
import { CROPS, PLOT_FLAG, type FieldPlot, type SimEventKind, type SimFrame, type StructureFrame } from '../sim/protocol'
import { useSimStream, type StreamStatus } from '../sim/stream'

// Three.js is big: load the 3D view only when watching (fetched quietly in the background, so
// switching to it is instant).
const load3d = () => import('../game3d/GameCanvas3D')
const GameCanvas3D = lazy(() => load3d().then((m) => ({ default: m.GameCanvas3D })))

/** Falls back to the 2D map when the 3D view cannot start (no WebGL 2, a blocked GPU, a failed download). */
class ViewBoundary extends Component<{ onError: () => void; children: ReactNode }, { failed: boolean }> {
  state = { failed: false }
  static getDerivedStateFromError() {
    return { failed: true }
  }
  componentDidCatch() {
    this.props.onError()
  }
  render() {
    return this.state.failed ? null : this.props.children
  }
}

const cameraKey = (mapId: string) => `miniv2.camera.${mapId}`

function loadCamera(mapId: string): MapView | null {
  try {
    const v = JSON.parse(localStorage.getItem(cameraKey(mapId)) ?? 'null') as MapView | null
    return v && Number.isFinite(v.x) && Number.isFinite(v.y) && v.across > 0 ? v : null
  } catch {
    return null
  }
}

function saveCamera(mapId: string, view: MapView) {
  const round = (n: number | undefined) => (n === undefined ? undefined : Math.round(n * 1000) / 1000)
  const { x, y, across, phi, theta } = view
  localStorage.setItem(
    cameraKey(mapId),
    JSON.stringify({ x: round(x), y: round(y), across: round(across), phi: round(phi), theta: round(theta) }),
  )
}

const reducedMotion = () => window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false

/** A still copy of a view's canvas, to cross-fade from while the other view takes over. */
function snapshot(src: HTMLCanvasElement | null | undefined) {
  if (!src || !src.width || !src.height) return null
  const copy = document.createElement('canvas')
  copy.width = src.width
  copy.height = src.height
  copy.getContext('2d')?.drawImage(src, 0, 0)
  return copy
}

type MapSearch = {
  mode: Mode
  /** Observed creature in watch mode, kept in the URL so reloads and links keep it. */
  creature?: number
}

const MODES: Mode[] = ['watch', 'play', 'edit']

/** Event-log moments the automatic camera may turn to (label for its chip, importance 1–3). */
const CINEMATIC_HINTS: Partial<Record<SimEventKind, { label: string; importance: number }>> = {
  birth: { label: 'Kelahiran', importance: 1.5 },
  village: { label: 'Kabar desa', importance: 2.5 },
  fire: { label: 'Kebakaran', importance: 3 },
}
export const Route = createFileRoute('/maps/$mapId')({
  validateSearch: (search: Record<string, unknown>): MapSearch => {
    const creature = Number(search.creature)
    return {
      mode: MODES.includes(search.mode as Mode) ? (search.mode as Mode) : 'watch',
      creature: Number.isInteger(creature) && creature > 0 ? creature : undefined,
    }
  },
  loader: ({ context: { queryClient }, params }) =>
    Promise.all([queryClient.ensureQueryData(mapQuery(params.mapId)), queryClient.ensureQueryData(tilesQuery)]),
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
  const worldRef = useRef<World3D | null>(null)
  // Watching can be in 3D; editing and playing stay in 2D. The choice is remembered.
  const [prefer3d, setPrefer3d] = useState(() => localStorage.getItem('miniv2.view3d') === '1')
  const view3d = mode === 'watch' && prefer3d
  // Fetched while watching in 2D too, so the 3D view starts with its hills in place.
  const relief = useQuery({ ...reliefQuery(map.id, map.updatedAt), enabled: mode === 'watch' })
  useEffect(() => {
    if (mode !== 'watch') return
    const id = setTimeout(() => void load3d(), 1200)
    return () => clearTimeout(id)
  }, [mode])

  // Switching views hands over the camera (the same spot, as wide) and cross-fades from a still of
  // the old view; the people and animals are shared (see LiveWorld), so they never jump.
  const stageWrapRef = useRef<HTMLDivElement>(null)
  const ghostRef = useRef<HTMLDivElement>(null)
  // The view to open with, kept (not used up) so a view mounted twice (StrictMode) still gets it:
  // first the camera saved last time (unless following someone or playing), then each switch's.
  const [saved] = useState(() => loadCamera(map.id))
  const handover = useRef<MapView | null>(mode !== 'play' && selectedId === null ? saved : null)
  // The 3D camera's last angle, so going back to 3D looks from where you left it.
  const angle = useRef<Pick<MapView, 'phi' | 'theta'>>({ phi: saved?.phi, theta: saved?.theta })
  const currentView = useCallback((): MapView | null => {
    const view = worldRef.current?.getView() ?? engineRef.current?.getView() ?? null
    return view && { ...angle.current, ...view }
  }, [])
  const [switching, setSwitching] = useState(false)
  const [view3dFailed, setView3dFailed] = useState(false)
  const fadeGhost = useCallback(() => {
    const host = ghostRef.current
    if (!host?.firstChild) return
    // Let the new view draw a couple of frames under it first.
    requestAnimationFrame(() =>
      requestAnimationFrame(() => {
        host.classList.add('fading')
        setTimeout(() => {
          if (host.classList.contains('fading')) host.replaceChildren()
          host.classList.remove('fading')
        }, 500)
      }),
    )
  }, [])
  const showGhost = (still: HTMLCanvasElement | null) => {
    const host = ghostRef.current
    if (!host || !still) return
    host.classList.remove('fading')
    host.replaceChildren(still)
    // Never left covering the stage, even if the new view fails to start.
    setTimeout(fadeGhost, 4000)
  }
  const setView3d = (on: boolean) => {
    localStorage.setItem('miniv2.view3d', on ? '1' : '0')
    setPrefer3d(on)
  }
  const toggle3d = () => {
    if (switching) return
    const canvas = stageWrapRef.current?.querySelector<HTMLCanvasElement>('.stage-canvas')
    if (!view3d) {
      setView3dFailed(false)
      handover.current = currentView()
      if (!reducedMotion()) showGhost(snapshot(canvas))
      setView3d(true)
      return
    }
    const world = worldRef.current
    if (!world) {
      setView3d(false)
      return
    }
    // Settle the 3D camera onto exactly what the 2D map will show, then swap.
    const { phi, theta } = world.getView()
    // (Not mid tilt-up, when it still looks nearly straight down.)
    if (phi !== undefined && phi > 0.2) angle.current = { phi, theta }
    setSwitching(true)
    world.flatten(snapAcross(world.getView().across, canvas?.clientWidth ?? 0), () => {
      handover.current = world.getView()
      if (!reducedMotion()) showGhost(snapshot(canvas))
      setSwitching(false)
      setView3d(false)
    })
  }

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

  // One smoothed copy of the stream, shared by whichever view is showing.
  const [live] = useState(() => new LiveWorld())
  // Replay (watching only): while it runs, the live stream stays connected but its frames and
  // buildings are only kept aside (the newest of each), and the views show recorded ones instead.
  const [replaying, setReplaying] = useState(false)
  const replayingRef = useRef(false)
  const liveFrame = useRef<SimFrame | null>(null)
  const liveStructures = useRef<StructureFrame[] | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const streamStatus = useSimStream(map.id, mode === 'watch', {
    viewport: () => {
      const v = currentView()
      return v ? { ...v, follow: following ? selectedId : null } : null
    },
    onFrame: (frame) => {
      if (replayingRef.current) {
        liveFrame.current = frame
        return
      }
      live.push(frame)
      engineRef.current?.onFrame(frame)
      worldRef.current?.onFrame(frame)
    },
    onStructures: (msg) => {
      liveStructures.current = msg.structures
      if (replayingRef.current) return
      live.structures = msg.structures
      engineRef.current?.setStructures(msg.structures)
      worldRef.current?.setStructures(msg.structures)
    },
    onFields: (msg) => {
      live.plots = msg.plots
      engineRef.current?.setFields(msg.plots)
      worldRef.current?.setFields(msg.plots)
    },
    onWater: (msg) => {
      live.water = msg
      engineRef.current?.setWater(msg)
      worldRef.current?.setWater(msg)
    },
    // Both views read the mosquitoes, villages and scorched land from the live world as they draw.
    onMosquitoes: (msg) => {
      live.mosquitoes = msg
    },
    onVillages: (msg) => {
      live.villages = msg.villages
    },
    onBurnt: (msg) => {
      live.burnt = msg
    },
  })
  // Stale people would slide across the map when the stream resumes.
  useEffect(() => {
    if (mode !== 'watch') live.clear()
  }, [mode, live])

  // (Both ignore a recorded frame that lands after going back to live, before the replay is torn down.)
  const replaySinks: ReplaySinks = {
    show: (frame, interval, jump) => {
      if (!replayingRef.current) return
      // After a jump in time everyone is placed, not slid across the map.
      if (jump) live.forgetMotion()
      live.push(frame, performance.now(), interval)
      engineRef.current?.onFrame(frame)
      worldRef.current?.onFrame(frame)
    },
    structures: (list) => {
      if (!replayingRef.current) return
      live.structures = list
      engineRef.current?.setStructures(list)
      worldRef.current?.setStructures(list)
    },
  }
  const startReplay = () => {
    if (replayingRef.current) return
    setNotice(null)
    liveFrame.current = live.frame
    liveStructures.current = live.structures
    replayingRef.current = true
    setReplaying(true)
  }
  // Back to the live stream: the buildings as they stand now and the newest frame, placed at once.
  const stopReplay = useCallback(() => {
    if (!replayingRef.current) return
    replayingRef.current = false
    setReplaying(false)
    live.forgetMotion()
    const structures = liveStructures.current
    if (structures) {
      live.structures = structures
      engineRef.current?.setStructures(structures)
      worldRef.current?.setStructures(structures)
    }
    const frame = liveFrame.current
    if (frame) {
      live.push(frame)
      engineRef.current?.onFrame(frame)
      worldRef.current?.onFrame(frame)
    }
  }, [live])
  useEffect(() => {
    if (mode !== 'watch') stopReplay()
  }, [mode, stopReplay])
  const replayMissing = useCallback(() => {
    stopReplay()
    setNotice('Server ini belum merekam tayangan ulang.')
  }, [stopReplay])
  useEffect(() => {
    if (!replaying) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== 'Escape' || (e.target as HTMLElement | null)?.closest?.('input, textarea, select')) return
      stopReplay()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [replaying, stopReplay])
  const info = useQuery({ ...simInfoQuery(map.id), enabled: mode === 'watch' })
  const spy = info.data?.secondsPerYear && info.data.secondsPerYear > 0 ? info.data.secondsPerYear : DEFAULT_SECONDS_PER_YEAR

  // A view created mid-stream (switching 2D ⇄ 3D, or between modes) picks up where the other left
  // off: its camera, and everything the stream has already sent.
  const attach = useCallback(
    (view: GameEngine | World3D) => {
      view.setLive(live)
      if (handover.current) view.setView(handover.current)
      if (mode === 'watch') {
        if (live.structures) view.setStructures(live.structures)
        if (live.plots) view.setFields(live.plots)
        if (live.water) view.setWater(live.water)
        if (live.frame) view.onFrame(live.frame)
      }
      fadeGhost()
    },
    [live, mode, fadeGhost],
  )

  // The camera is remembered per map, so a reload opens where you were looking.
  useEffect(() => {
    const save = () => {
      const view = currentView()
      // (A view whose canvas has just left the page measures nothing.)
      if (!view || !(view.across > 0) || !Number.isFinite(view.x)) return
      // Mid-switch the 3D camera looks straight down; keep the angle it had before.
      if (view.phi !== undefined && view.phi < 0.2) Object.assign(view, angle.current)
      saveCamera(map.id, view)
    }
    const id = setInterval(save, 2000)
    window.addEventListener('pagehide', save)
    return () => {
      clearInterval(id)
      window.removeEventListener('pagehide', save)
      save() // leaving through the app's own links fires no pagehide
    }
  }, [map.id, currentView])

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

  // The automatic camera (3D only). The world reports what it looks at, and null once it is off
  // (e.g. the user moved the camera), which turns the button off too.
  const [cinematic, setCinematic] = useState(false)
  const [cineSubject, setCineSubject] = useState<CinematicSubject | null>(null)
  const events3d = useMemo<World3DEvents>(
    () => ({
      ...events,
      onCinematicSubject: (subject) => {
        setCineSubject(subject)
        if (!subject) setCinematic(false)
      },
    }),
    [events],
  )
  const toggleCinematic = () => {
    const world = worldRef.current
    // (Not before the 3D view has loaded: there is no camera to hand over yet.)
    if (!world) return
    const next = !cinematic
    setCinematic(next)
    setCineSubject(null)
    world.setCinematic(next)
  }
  // A new 3D view starts with its own camera; the 2D map has no director.
  useEffect(() => {
    setCinematic(false)
    setCineSubject(null)
  }, [view3d])
  // The event log tells the director what frames alone don't: a birth, a new leader. Only events
  // newer than the moment it was switched on, about someone the stream still shows.
  const hinted = useRef(Infinity)
  const logged = info.data?.events
  useEffect(() => {
    const world = worldRef.current
    if (!cinematic || !world || !logged) {
      hinted.current = Infinity
      return
    }
    if (hinted.current === Infinity) hinted.current = logged.at(-1)?.id ?? 0
    for (const e of logged) {
      if (e.id <= hinted.current) continue
      hinted.current = e.id
      const hint = CINEMATIC_HINTS[e.kind]
      const c = e.creatureId ? live.creatures.get(e.creatureId) : undefined
      if (hint && c) world.cinematicHint({ id: c.id, x: c.rx, y: c.ry, ...hint })
    }
  }, [logged, cinematic, live])

  // Points whichever view is showing at a spot (a village from the Desa tab), letting go of anyone followed.
  const locate = useCallback((view: { x: number; y: number; across: number }) => {
    setFollowing(false)
    const target = { ...angle.current, ...view }
    if (worldRef.current) worldRef.current.setView(target)
    else engineRef.current?.setView(target)
  }, [])

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

  const setMode = (next: Mode) => {
    // Whenever the change swaps the 2D map for the 3D view or back, the camera carries over (and a
    // half-done switch is dropped).
    if (prefer3d && (mode === 'watch') !== (next === 'watch')) handover.current = currentView()
    setSwitching(false)
    navigate({ search: (prev) => ({ ...prev, mode: next }), replace: true })
  }

  return (
    <main className={`map-page mode-${mode}${replaying ? ' replaying' : ''}`}>
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
            {replaying ? (
              <span className="stream-badge replay" title="Menampilkan rekaman; simulasi tetap berjalan langsung di server">
                <span aria-hidden="true">⏪</span>
                <span className="stream-text"> Rekaman</span>
              </span>
            ) : (
              <span className={`stream-badge ${streamStatus}`} title={`Status aliran simulasi dari server: ${STREAM_LABELS[streamStatus]}`}>
                <span aria-hidden="true">●</span>
                <span className="stream-text"> {STREAM_LABELS[streamStatus]}</span>
              </span>
            )}
            <SimStatus mapId={map.id} />
          </>
        )}

        <div className="spacer" />

        {mode === 'watch' && (
          <button
            type="button"
            className={`geo-toggle replay-toggle ${replaying ? 'active' : ''}`}
            aria-pressed={replaying}
            aria-label="Tayangan ulang"
            title={
              replaying
                ? 'Kembali ke siaran langsung (Esc)'
                : 'Tayangan ulang: putar lagi beberapa menit terakhir (simulasi tetap berjalan)'
            }
            onClick={() => (replaying ? stopReplay() : startReplay())}
          >
            ⏪ <span className="btn-label">Tayangan ulang</span>
          </button>
        )}

        {mode === 'watch' && <CinematicButton available={view3d && !switching} on={cinematic} onToggle={toggleCinematic} />}

        {mode === 'watch' && (
          <button
            type="button"
            className={`geo-toggle ${view3d ? 'active' : ''}`}
            aria-pressed={view3d}
            title={view3d ? 'Kembali ke peta 2D' : 'Tampilkan pulau dalam 3D'}
            disabled={switching}
            onClick={toggle3d}
          >
            {view3d ? '🗺' : '🏝'} <span className="btn-label">{view3d ? '2D' : '3D'}</span>
          </button>
        )}

        <button
          type="button"
          className={`geo-toggle ${geologyShown ? 'active' : ''}`}
          aria-pressed={geologyShown}
          disabled={!geology.data || view3d}
          title={geology.data ? 'Tampilkan peta geologi: jenis batuan, gunung api, sungai, dan endapan' : 'Data geologi belum tersedia'}
          onClick={() => setShowGeology(!showGeology)}
        >
          🪨 <span className="btn-label">Geologi</span>
        </button>

        <div className="segmented">
          <button
            type="button"
            title="Perkecil (-)"
            onClick={() => (view3d ? worldRef.current?.zoomBy(-1) : engineRef.current?.zoomBy(-1))}
          >
            −
          </button>
          <button type="button" title="Perbesar (+)" onClick={() => (view3d ? worldRef.current?.zoomBy(1) : engineRef.current?.zoomBy(1))}>
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
              <button type="button" title="Undo (Ctrl+Z)" disabled={!history.canUndo} onClick={() => engineRef.current?.undo()}>
                ↶
              </button>
              <button type="button" title="Redo (Ctrl+Y)" disabled={!history.canRedo} onClick={() => engineRef.current?.redo()}>
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
      {notice && (
        <div className="banner error" role="status">
          {notice}{' '}
          <button type="button" className="ghost" onClick={() => setNotice(null)}>
            Tutup
          </button>
        </div>
      )}
      {view3dFailed && !view3d && (
        <div className="banner error">Tampilan 3D tidak bisa dimulai di browser ini, jadi pulau ditampilkan di peta 2D.</div>
      )}

      <div className="workspace">
        <div className="stage-wrap" ref={stageWrapRef}>
          {view3d ? (
            <ViewBoundary
              onError={() => {
                setView3dFailed(true)
                setView3d(false)
              }}
            >
              <Suspense fallback={<div className="stage center muted">Memuat tampilan 3D…</div>}>
                <GameCanvas3D
                  map={map}
                  tiles={tiles}
                  relief={relief.data}
                  minedOut={mined.data}
                  selectedId={selectedId}
                  following={following}
                  worldRef={worldRef}
                  events={events3d}
                  onReady={attach}
                />
              </Suspense>
            </ViewBoundary>
          ) : (
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
              onReady={attach}
            />
          )}
          <div className="stage-ghost" ref={ghostRef} aria-hidden />
          {mode === 'watch' && cinematic && view3d && (
            <div className="stage-top">
              <CinematicSubjectChip subject={cineSubject} onSelect={select} />
            </div>
          )}
          <div className="hint">
            {view3d ? (
              <>
                <kbd>klik</kbd> makhluk atau bangunan · <kbd>drag</kbd> putar · <kbd>klik kanan</kbd> geser · <kbd>scroll</kbd> zoom ·{' '}
                <kbd>WASD</kbd> jalan
              </>
            ) : mode === 'watch' ? (
              <>
                <kbd>klik</kbd> makhluk atau bangunan untuk mengamati · <kbd>drag</kbd>/<kbd>WASD</kbd> geser · <kbd>scroll</kbd> zoom
              </>
            ) : mode === 'play' ? (
              <>
                <kbd>WASD</kbd>/<kbd>←↑↓→</kbd> jalan · <kbd>Shift</kbd> lari · <kbd>scroll</kbd> zoom
              </>
            ) : (
              <>
                <kbd>klik kiri</kbd> cat · <kbd>klik kanan</kbd> geser · <kbd>WASD</kbd> kamera · <kbd>Ctrl+Z</kbd>/<kbd>Ctrl+Y</kbd>{' '}
                undo/redo · <kbd>Ctrl+S</kbd> simpan
              </>
            )}
          </div>
          {geologyShown && !view3d && geology.data && <GeologyLegend geology={geology.data} />}
          {replaying && (
            <ReplayOverlay mapId={map.id} sinks={replaySinks} secondsPerYear={spy} onExit={stopReplay} onMissing={replayMissing} />
          )}
          <div className={`hud ${replaying ? 'hud-hidden' : ''}`}>
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
        {mode === 'edit' && <Palette tiles={tiles} tool={tool} brush={brush} onTool={setTool} onBrush={setBrush} />}
        {mode === 'watch' && (
          <ObserverPanel
            mapId={map.id}
            selectedId={selectedId}
            following={following}
            onSelect={select}
            onFollow={setFollowing}
            onLocate={locate}
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
