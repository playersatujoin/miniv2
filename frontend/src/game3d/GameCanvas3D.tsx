import { useEffect, useLayoutEffect, useRef, type RefObject } from 'react'
import type { GameMap, MapRelief, TileSet } from '../api/client'
import type { MinedOut } from '../sim/protocol'
import { World3D, type World3DEvents } from './World3D'

type Props = {
  map: GameMap
  tiles: TileSet
  relief?: MapRelief
  minedOut?: MinedOut
  selectedId: number | null
  following: boolean
  worldRef: RefObject<World3D | null>
  events: World3DEvents
  /** Called once the view is set up, e.g. to hand it the 2D map's camera and the stream so far. */
  onReady?: (world: World3D) => void
}

/** Hosts the 3D view of the living island; React only feeds it props. */
export function GameCanvas3D({ map, tiles, relief, minedOut, selectedId, following, worldRef, events, onReady }: Props) {
  const stageRef = useRef<HTMLDivElement>(null)
  const labelsRef = useRef<HTMLDivElement>(null)
  const latest = useRef({ map, relief, minedOut, selectedId, following, events, onReady })
  useLayoutEffect(() => {
    latest.current = { map, relief, minedOut, selectedId, following, events, onReady }
  })

  useEffect(() => {
    const ev = () => latest.current.events
    const world = new World3D(stageRef.current!, labelsRef.current!, latest.current.map, tiles, {
      onSelectCreature: (id) => ev().onSelectCreature?.(id),
      onFollowChange: (f) => ev().onFollowChange?.(f),
      onCinematicSubject: (subject) => ev().onCinematicSubject?.(subject),
    })
    world.setRelief(latest.current.relief)
    world.setMinedOut(latest.current.minedOut)
    world.setSelected(latest.current.selectedId)
    world.setFollow(latest.current.following)
    worldRef.current = world
    latest.current.onReady?.(world)
    stageRef.current!.querySelector('canvas')?.focus()
    return () => {
      world.destroy()
      worldRef.current = null
    }
  }, [map.id, tiles, worldRef])

  useEffect(() => worldRef.current?.setRelief(relief), [relief, worldRef])
  useEffect(() => worldRef.current?.setMinedOut(minedOut), [minedOut, worldRef])
  useEffect(() => {
    worldRef.current?.setSelected(selectedId)
    worldRef.current?.setFollow(following)
  }, [selectedId, following, worldRef])
  useEffect(() => worldRef.current?.setMap(map), [map, worldRef])

  return (
    <div className="stage" ref={stageRef}>
      <div className="label-layer" ref={labelsRef} />
    </div>
  )
}
