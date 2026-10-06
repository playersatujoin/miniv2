import { useEffect } from 'react'
import { ReplayBar, replayClock } from './ReplayBar'
import type { ReplaySinks } from './replayPlayer'
import { formatClockSpan, secondsBehind } from './replayTimeline'
import { useReplay } from './useReplay'

type Props = {
  mapId: string
  /** Where recorded frames and buildings go (the page's LiveWorld and views). */
  sinks: ReplaySinks
  secondsPerYear: number
  onExit: () => void
  /** The server keeps no replay (an older one). */
  onMissing: () => void
}

/**
 * Everything on the stage while replaying: a badge saying so (with the
 * recorded time) and the controls. Mounted only while replaying; it re-renders
 * with every recorded frame, the page around it doesn't.
 */
export function ReplayOverlay({ mapId, sinks, secondsPerYear: spy, onExit, onMissing }: Props) {
  const replay = useReplay(mapId, true, sinks)
  useEffect(() => {
    if (replay.missing) onMissing()
  }, [replay.missing, onMissing])
  const s = replay.state

  return (
    <>
      <div className="replay-frame" aria-hidden />
      <div className="replay-top">
        <div className="replay-badge" role="status">
          <span aria-hidden="true">⏪</span> <strong>Tayangan ulang</strong>
          {s.index >= 0 && (
            <>
              {' · '}
              {replayClock(s.time, spy)}
              <span className="replay-badge-behind">
                {' · '}
                {s.atEnd ? 'akhir rekaman' : `${formatClockSpan(secondsBehind(replay.frames, s.index))} sebelum siaran langsung`}
              </span>
            </>
          )}
        </div>
      </div>
      <ReplayBar
        state={s}
        frames={replay.frames}
        marks={replay.marks}
        secondsPerYear={spy}
        loading={replay.loading}
        error={replay.error ?? s.error}
        onSeek={(i) => replay.player?.seek(i)}
        onToggle={() => replay.player?.toggle()}
        onStep={(dir) => replay.player?.step(dir)}
        onSpeed={(v) => replay.player?.setSpeed(v)}
        onExit={onExit}
      />
    </>
  )
}
