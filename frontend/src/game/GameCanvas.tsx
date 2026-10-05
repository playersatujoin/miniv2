import { useEffect, useLayoutEffect, useRef, type RefObject } from 'react'
import type { GameMap, MapGeology, TileSet } from '../api/client'
import type { MinedOut } from '../sim/protocol'
import { GameEngine, type Brush, type EngineEvents, type Mode, type Tool } from './engine'

type Props = {
  map: GameMap
  tiles: TileSet
  mode: Mode
  tool: Tool
  brush: Brush
  /** Watch mode: the observed creature and whether the camera follows it. */
  selectedId: number | null
  following: boolean
  /** Resource deposits to draw on the ground (loaded separately from the map). */
  geology?: MapGeology
  /** Show the geological map (rock units, feature labels) over the land. */
  geologyOverlay?: boolean
  /** Mined-out deposits of the living world (watch mode only). */
  minedOut?: MinedOut
  engineRef: RefObject<GameEngine | null>
  events: EngineEvents
}

/** Hosts the imperative GameEngine; React only feeds it props. */
export function GameCanvas({
  map,
  tiles,
  mode,
  tool,
  brush,
  selectedId,
  following,
  geology,
  geologyOverlay = false,
  minedOut,
  engineRef,
  events,
}: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const minimapRef = useRef<HTMLCanvasElement>(null)
  const latest = useRef({ mode, tool, brush, selectedId, following, geology, geologyOverlay, minedOut, events, map })
  useLayoutEffect(() => {
    latest.current = { mode, tool, brush, selectedId, following, geology, geologyOverlay, minedOut, events, map }
  })

  // Recreate the engine only when switching maps; data updates go through setMap below.
  useEffect(() => {
    const ev = () => latest.current.events
    const engine = new GameEngine(canvasRef.current!, minimapRef.current, tiles, latest.current.map, {
      onDirtyChange: (d) => ev().onDirtyChange?.(d),
      onHistoryChange: (u, r) => ev().onHistoryChange?.(u, r),
      onHover: (h) => ev().onHover?.(h),
      onPlayerTile: (p) => ev().onPlayerTile?.(p),
      onSelectCreature: (id) => ev().onSelectCreature?.(id),
      onFollowChange: (f) => ev().onFollowChange?.(f),
    })
    engine.setMode(latest.current.mode)
    engine.setTool(latest.current.tool)
    engine.setBrush(latest.current.brush)
    engine.setSelected(latest.current.selectedId)
    engine.setFollow(latest.current.following)
    engine.setGeology(latest.current.geology)
    engine.setGeologyOverlay(latest.current.geologyOverlay)
    engine.setMinedOut(latest.current.minedOut)
    engineRef.current = engine
    canvasRef.current!.focus()
    return () => {
      engine.destroy()
      engineRef.current = null
    }
  }, [map.id, tiles, engineRef])

  useEffect(() => engineRef.current?.setMode(mode), [mode, engineRef])
  useEffect(() => engineRef.current?.setTool(tool), [tool, engineRef])
  useEffect(() => engineRef.current?.setBrush(brush), [brush, engineRef])
  useEffect(() => engineRef.current?.setGeology(geology), [geology, engineRef])
  useEffect(() => engineRef.current?.setGeologyOverlay(geologyOverlay), [geologyOverlay, engineRef])
  // Geology rebakes the chunks, which already read the mined-out set; this only patches changes.
  useEffect(() => engineRef.current?.setMinedOut(minedOut), [minedOut, engineRef])
  // Selection first, so following a newly selected creature takes effect.
  useEffect(() => {
    engineRef.current?.setSelected(selectedId)
    engineRef.current?.setFollow(following)
  }, [selectedId, following, engineRef])

  // Pick up a newer server version (e.g. refetch) unless there are unsaved edits.
  useEffect(() => {
    const engine = engineRef.current
    if (engine && !engine.isDirty() && engine.version !== map.updatedAt) engine.setMap(map)
  }, [map, engineRef])

  return (
    <div className="stage">
      <canvas ref={canvasRef} className="stage-canvas" tabIndex={0} />
      <canvas
        ref={minimapRef}
        className="minimap"
        style={{ aspectRatio: `${map.width} / ${map.height}` }}
        title={mode !== 'play' ? 'Klik untuk memindahkan kamera' : undefined}
      />
    </div>
  )
}
