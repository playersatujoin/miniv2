import { useMemo, useRef, useState, type KeyboardEvent, type PointerEvent as ReactPointerEvent } from 'react'
import { MONTH_NAMES, type ReplayMark } from '../../sim/protocol'
import { isViolent, kindStyle, violenceHidden } from './eventKinds'
import { formatClock, formatClockShort, nf } from './format'
import type { ReplayState } from './replayPlayer'
import {
  REPLAY_SPEEDS,
  adjacentMark,
  bucketMarks,
  formatClockSpan,
  frameAtTick,
  indexToPosition,
  nearestMark,
  positionToIndex,
  rankMarks,
  REPLAY_FRAME_SECONDS,
  secondsBehind,
  type Frames,
  type ReplaySpeed,
} from './replayTimeline'
import { useElementWidth } from './useElementWidth'

/** Marker colour and height by importance (height repeats the colour's meaning). */
export const MARK_STYLE: Record<1 | 2 | 3, { color: string; height: number; label: string }> = {
  1: { color: '#6b7a93', height: 6, label: 'kecil' },
  2: { color: '#3987e5', height: 10, label: 'sedang' },
  3: { color: '#ffd166', height: 15, label: 'penting' },
}

/** Pixels: a marker this close to the pointer is the one it is on. */
const HIT_PX = 6
/** Jumping to a marker lands this many frames (half a second each) before it, to see it coming. */
const LEAD_FRAMES = 2

type Props = {
  state: ReplayState
  frames: Frames
  marks: readonly ReplayMark[]
  secondsPerYear: number
  loading?: boolean
  error?: string | null
  onSeek: (index: number) => void
  onToggle: () => void
  onStep: (dir: number) => void
  onSpeed: (speed: ReplaySpeed) => void
  onExit: () => void
}

/** "Tahun 3.637, Maret" for a simulated time. */
export function replayClock(time: number, spy: number) {
  const month = MONTH_NAMES[Math.floor(((Math.max(0, time) / spy) % 1) * 12)] ?? ''
  return `${formatClock(time, spy)}${month ? `, ${month}` : ''}`
}

/**
 * The replay's controls under the map: play/pause, step, speed, and a timeline
 * across the recorded window with the events marked on it (click to jump, hover
 * to read). Watching only: nothing here reaches the simulation.
 */
export function ReplayBar({
  state,
  frames,
  marks: serverMarks,
  secondsPerYear: spy,
  loading,
  error,
  onSeek,
  onToggle,
  onStep,
  onSpeed,
  onExit,
}: Props) {
  const [trackRef, width] = useElementWidth<HTMLDivElement>()
  // Common kinds (a famine's deaths) step down a level, so rare important events stand out.
  const marks = useMemo(() => rankMarks(serverMarks, frames.length * REPLAY_FRAME_SECONDS), [serverMarks, frames.length])
  const [hover, setHover] = useState<{ x: number; position: number; mark: (typeof marks)[number] | null; count: number } | null>(null)
  const drag = useRef<{ last: number; at: number } | null>(null)
  const n = frames.length
  const index = Math.max(0, Math.min(n - 1, state.index))
  const position = state.index >= 0 ? indexToPosition(frames, index) : 1
  const buckets = width > 0 ? bucketMarks(frames, marks, Math.max(20, Math.floor(width / 5))) : []
  const behind = secondsBehind(frames, index)

  const positionAt = (clientX: number) => {
    const r = trackRef.current?.getBoundingClientRect()
    if (!r || r.width <= 0) return 0
    return Math.min(1, Math.max(0, (clientX - r.left) / r.width))
  }
  const markAt = (p: number) => (width > 0 ? nearestMark(frames, marks, p, HIT_PX / width) : null)
  const jumpToMark = (m: ReplayMark) => onSeek(Math.max(0, frameAtTick(frames, m.tick) - LEAD_FRAMES))

  const onPointerDown = (e: ReactPointerEvent<HTMLDivElement>) => {
    if (!n || e.button !== 0) return
    e.currentTarget.setPointerCapture(e.pointerId)
    const p = positionAt(e.clientX)
    const m = markAt(p)
    if (m) {
      jumpToMark(m)
      return
    }
    const i = positionToIndex(frames, p)
    drag.current = { last: i, at: performance.now() }
    onSeek(i)
  }
  const onPointerMove = (e: ReactPointerEvent<HTMLDivElement>) => {
    const p = positionAt(e.clientX)
    const d = drag.current
    if (d) {
      // Scrubbing: a new frame at most every 80 ms.
      const i = positionToIndex(frames, p)
      const now = performance.now()
      if (i !== d.last && now - d.at > 80) {
        drag.current = { last: i, at: now }
        onSeek(i)
      }
    }
    const r = trackRef.current?.getBoundingClientRect()
    const m = d ? null : markAt(p)
    const count = m ? (buckets.find((b) => b.mark === m)?.count ?? 1) : 0
    setHover({ x: p * (r?.width ?? 0), position: p, mark: m, count })
  }
  const onPointerUp = (e: ReactPointerEvent<HTMLDivElement>) => {
    const d = drag.current
    drag.current = null
    if (d) {
      const i = positionToIndex(frames, positionAt(e.clientX))
      if (i !== d.last) onSeek(i)
    }
  }

  const onKeyDown = (e: KeyboardEvent<HTMLDivElement>) => {
    if (!n) return
    const tick = frames[index]?.[0] ?? 0
    let handled = true
    if (e.key === 'ArrowLeft') onStep(e.shiftKey ? -10 : -1)
    else if (e.key === 'ArrowRight') onStep(e.shiftKey ? 10 : 1)
    else if (e.key === 'Home') onSeek(0)
    else if (e.key === 'End') onSeek(n - 1)
    else if (e.key === 'PageUp') {
      const m = adjacentMark(marks, tick, -1, 2)
      if (m) jumpToMark(m)
    } else if (e.key === 'PageDown') {
      const m = adjacentMark(marks, tick, 1, 2)
      if (m) jumpToMark(m)
    } else if (e.key === ' ' || e.key === 'k') onToggle()
    else handled = false
    if (handled) e.preventDefault()
  }

  const hoverIndex = hover ? positionToIndex(frames, hover.position) : -1
  const tipStyle = hover ? { left: `clamp(120px, ${hover.x}px, calc(100% - 120px))` } : undefined

  return (
    <div className="replay-bar" role="group" aria-label="Kontrol tayangan ulang">
      <div className="replay-buttons">
        <button
          type="button"
          className="replay-btn"
          title="Mundur satu langkah (←)"
          aria-label="Mundur satu langkah"
          disabled={!n}
          onClick={() => onStep(-1)}
        >
          ⏮
        </button>
        <button
          type="button"
          className="replay-btn replay-play"
          title={state.playing ? 'Jeda tayangan ulang (spasi)' : 'Putar tayangan ulang (spasi)'}
          aria-label={state.playing ? 'Jeda' : 'Putar'}
          disabled={!n}
          onClick={onToggle}
        >
          {state.playing ? '⏸' : '▶'}
        </button>
        <button
          type="button"
          className="replay-btn"
          title="Maju satu langkah (→)"
          aria-label="Maju satu langkah"
          disabled={!n}
          onClick={() => onStep(1)}
        >
          ⏭
        </button>
      </div>
      <div className="segmented replay-speeds" role="group" aria-label="Kecepatan tayangan ulang">
        {REPLAY_SPEEDS.map((s) => (
          <button
            key={s}
            type="button"
            className={state.speed === s ? 'active' : ''}
            aria-pressed={state.speed === s}
            title={`Putar ${s === 0.5 ? 'setengah' : `${s}×`} kecepatan rekaman`}
            onClick={() => onSpeed(s)}
          >
            {s === 0.5 ? '½×' : `${s}×`}
          </button>
        ))}
      </div>

      <div className="replay-timeline">
        <div
          ref={trackRef}
          className="replay-track"
          role="slider"
          tabIndex={0}
          aria-label="Lini masa tayangan ulang"
          aria-valuemin={0}
          aria-valuemax={Math.max(0, n - 1)}
          aria-valuenow={index}
          aria-valuetext={n ? `${replayClock(state.time, spy)}, ${formatClockSpan(behind)} sebelum siaran langsung` : 'Belum ada rekaman'}
          aria-keyshortcuts="ArrowLeft ArrowRight Home End PageUp PageDown Space"
          onPointerDown={onPointerDown}
          onPointerMove={onPointerMove}
          onPointerUp={onPointerUp}
          onPointerCancel={() => (drag.current = null)}
          onPointerLeave={() => setHover(null)}
          onKeyDown={onKeyDown}
        >
          <div className="replay-played" style={{ width: `${position * 100}%` }} />
          {buckets.map((b) => {
            const st = MARK_STYLE[Math.min(3, Math.max(1, b.mark.importance)) as 1 | 2 | 3]
            return (
              <i
                key={`${b.mark.tick}-${b.position}`}
                className="replay-mark"
                style={{ left: `${b.position * 100}%`, height: st.height, background: st.color }}
                aria-hidden
              />
            )
          })}
          <div className="replay-cursor" style={{ left: `${position * 100}%` }} aria-hidden />
          {hover && hoverIndex >= 0 && (
            <div className="replay-tip" style={tipStyle} role="tooltip">
              {hover.mark ? (
                <>
                  <span
                    className="replay-tip-kind"
                    style={{ color: MARK_STYLE[Math.min(3, Math.max(1, hover.mark.importance)) as 1 | 2 | 3].color }}
                  >
                    {kindStyle(hover.mark.kind).label} · {formatClockShort(hover.mark.time, spy)}
                  </span>
                  <span>{violenceHidden() && isViolent(hover.mark) ? 'Peristiwa kekerasan (disembunyikan)' : hover.mark.text}</span>
                  {hover.count > 1 && <small className="muted">dan {nf.format(hover.count - 1)} kejadian lain di sekitar sini</small>}
                  <small className="muted">Klik untuk melompat ke sini</small>
                </>
              ) : (
                <>
                  <span>{replayClock(frames[hoverIndex][1], spy)}</span>
                  <small className="muted">{formatClockSpan(secondsBehind(frames, hoverIndex))} sebelum siaran langsung</small>
                </>
              )}
            </div>
          )}
        </div>
        <ReplayLegend />
      </div>

      <div className="replay-time" aria-live="off">
        <strong>{n ? replayClock(state.time, spy) : '—'}</strong>
        <small className="muted">
          {error ? (
            <span className="obs-error">{error}</span>
          ) : loading || !n ? (
            'memuat rekaman…'
          ) : state.waiting ? (
            'memuat…'
          ) : state.atEnd ? (
            'akhir rekaman'
          ) : (
            `−${formatClockSpan(behind)} dari siaran langsung`
          )}
        </small>
      </div>
      <button type="button" className="replay-exit" onClick={onExit} title="Kembali ke siaran langsung (Esc)">
        ● Kembali ke siaran langsung
      </button>
    </div>
  )
}

/** The legend for the timeline's markers. */
function ReplayLegend() {
  return (
    <span className="replay-legend small muted" aria-hidden>
      {([3, 2, 1] as const).map((k) => (
        <span key={k}>
          <i style={{ background: MARK_STYLE[k].color, height: MARK_STYLE[k].height }} /> {MARK_STYLE[k].label}
        </span>
      ))}
    </span>
  )
}
