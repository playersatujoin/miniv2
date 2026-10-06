import { useQuery } from '@tanstack/react-query'
import { useEffect, useLayoutEffect, useRef, useState, useSyncExternalStore } from 'react'
import { ApiError } from '../../api/client'
import { fetchReplayFrame, fetchReplayStructures, replayIndexQuery } from '../../sim/api'
import { ReplayPlayer, type ReplaySinks, type ReplayState } from './replayPlayer'

/** Frames back from the newest where the replay opens: ten seconds (it then plays). */
const START_BACK = 20

const IDLE: ReplayState = { index: -1, tick: -1, time: 0, playing: false, speed: 1, waiting: false, atEnd: false, error: null }
const idle = () => IDLE
const noSubscribe = () => () => {}

/**
 * The replay while `active`: its index (refreshed as the live world records
 * on), a player that feeds recorded frames to `sinks`, and the player's state.
 * The live stream is the caller's to pause and resume.
 */
export function useReplay(mapId: string, active: boolean, sinks: ReplaySinks) {
  const index = useQuery({ ...replayIndexQuery(mapId), enabled: active })
  const sinksRef = useRef(sinks)
  useLayoutEffect(() => {
    sinksRef.current = sinks
  })
  const [player, setPlayer] = useState<ReplayPlayer | null>(null)

  useEffect(() => {
    if (!active) return
    const p = new ReplayPlayer(
      { frame: (tick) => fetchReplayFrame(mapId, tick), structures: (v) => fetchReplayStructures(mapId, v) },
      {
        show: (frame, interval, jump) => sinksRef.current.show(frame, interval, jump),
        structures: (list, version) => sinksRef.current.structures(list, version),
      },
    )
    setPlayer(p)
    return () => {
      p.destroy()
      setPlayer(null)
    }
  }, [active, mapId])

  // Each fresh index slides the window; the first one starts playback a little before now.
  const started = useRef<ReplayPlayer | null>(null)
  useEffect(() => {
    if (!player || !index.data) return
    player.setIndex(index.data)
    if (started.current !== player && index.data.frames.length) {
      started.current = player
      player.seek(index.data.frames.length - 1 - START_BACK)
      player.play()
    }
  }, [player, index.data])

  const state = useSyncExternalStore(player?.subscribe ?? noSubscribe, player?.getState ?? idle)
  const missing = index.error instanceof ApiError && index.error.status === 404
  return {
    player,
    state,
    frames: index.data?.frames ?? [],
    marks: index.data?.marks ?? [],
    loading: active && index.isPending,
    /** The server keeps no replay (an older one). */
    missing,
    error: index.isError && !missing ? index.error.message : null,
  }
}
