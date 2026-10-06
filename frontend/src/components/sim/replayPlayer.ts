// The replay's playback controller (after RAGE's replay controller and its
// buffer markers: a cursor over a recorded window that plays, pauses, steps
// and jumps, with a preloader reading ahead of the cursor). It only reads the
// server's recording and hands frames to the page's views; nothing is ever
// sent to the simulation, which keeps running live on the server.
//
// No React and no DOM: the loaders, the clock and the sinks come from outside,
// so node tests can drive it with fakes.

import type { ReplayIndex, SimFrame, StructureFrame, StructuresMessage } from '../../sim/protocol'
import { REPLAY_SPEEDS, evictable, frameAtTick, frameDelay, prefetchWindow, type Frames, type ReplaySpeed } from './replayTimeline'

export type ReplayFrameData = { frame: SimFrame; structures: number }

export type ReplayLoaders = {
  /** The recorded frame at (or just before) a tick, and the building version it shows. */
  frame: (tick: number) => Promise<ReplayFrameData>
  /** Buildings as they stood in a recorded version. */
  structures: (version: number) => Promise<StructuresMessage>
}

export type ReplaySinks = {
  /**
   * A recorded frame to show. `interval` (ms) is when the next one is due, so
   * people glide between them at the replay's pace; `jump` after a seek: place
   * everyone where the frame says instead of sliding them there.
   */
  show: (frame: SimFrame, interval: number, jump: boolean) => void
  /** The buildings the recorded frames show (a new version arrived). */
  structures: (list: StructureFrame[], version: number) => void
}

export type ReplayClock = {
  setTimeout: (fn: () => void, ms: number) => unknown
  clearTimeout: (id: unknown) => void
}

export type ReplayState = {
  /** Cursor: index into the current recording's frames (-1 before anything is shown). */
  index: number
  /** Tick and simulated seconds of the frame on screen (-1 / 0 before the first). */
  tick: number
  time: number
  playing: boolean
  speed: ReplaySpeed
  /** The next frame is still loading. */
  waiting: boolean
  /** Played to the newest recorded frame (and paused there). */
  atEnd: boolean
  error: string | null
}

/** At most this many frame requests in flight at once. */
const MAX_INFLIGHT = 4

const defaultClock: ReplayClock = {
  setTimeout: (fn, ms) => globalThis.setTimeout(fn, ms),
  clearTimeout: (id) => globalThis.clearTimeout(id as ReturnType<typeof setTimeout>),
}

export class ReplayPlayer {
  private frames: Frames = []
  /** Frame index by tick (the first of a run of equal ticks). */
  private indexOf = new Map<number, number>()
  /** Frames requested or held, by tick. */
  private cache = new Map<number, Promise<ReplayFrameData>>()
  private ready = new Set<number>()
  private inflight = 0
  private buildings = new Map<number, Promise<StructuresMessage>>()
  private shownBuildings = 0
  private wantedBuildings = 0
  /** Bumped by every seek (and teardown), so a late frame from an earlier seek is never shown. */
  private seekGen = 0
  /** Bumped by every seek, pause and teardown: the play loop that was running stops. */
  private playGen = 0
  /** A seek's frame is still loading. */
  private seeking = false
  private timer: unknown = null
  private listeners = new Set<() => void>()
  private destroyed = false
  private state: ReplayState = { index: -1, tick: -1, time: 0, playing: false, speed: 1, waiting: false, atEnd: false, error: null }

  private readonly load: ReplayLoaders
  private readonly sinks: ReplaySinks
  private readonly clock: ReplayClock

  constructor(load: ReplayLoaders, sinks: ReplaySinks, clock: ReplayClock = defaultClock) {
    this.load = load
    this.sinks = sinks
    this.clock = clock
  }

  // --- For React (useSyncExternalStore) ------------------------------------------

  subscribe = (fn: () => void) => {
    this.listeners.add(fn)
    return () => this.listeners.delete(fn)
  }

  getState = () => this.state

  private set(patch: Partial<ReplayState>) {
    this.state = { ...this.state, ...patch }
    for (const fn of this.listeners) fn()
  }

  // --- The recording -------------------------------------------------------------

  /** The recorded frames, oldest first. */
  get recording(): Frames {
    return this.frames
  }

  /**
   * Takes a fresh index from the server (it slides as the live world records).
   * The cursor stays on the same tick; if that has dropped out of the window it
   * moves to the oldest frame still held.
   */
  setIndex(index: Pick<ReplayIndex, 'frames'> | null | undefined) {
    this.frames = index?.frames ?? []
    this.indexOf.clear()
    for (let i = this.frames.length - 1; i >= 0; i--) this.indexOf.set(this.frames[i][0], i)
    if (this.state.index >= 0) {
      const i = this.frames.length ? frameAtTick(this.frames, this.state.tick) : -1
      const atEnd = this.state.atEnd && i >= this.frames.length - 1
      if (i !== this.state.index || atEnd !== this.state.atEnd) this.set({ index: i, atEnd })
    }
    this.prefetch()
  }

  // --- Controls -------------------------------------------------------------------

  /** Shows frame i (clamped to the recording) and, if playing, plays on from there. */
  seek(i: number) {
    if (this.destroyed || !this.frames.length) return
    const index = Math.min(this.frames.length - 1, Math.max(0, Math.round(i)))
    const gen = ++this.seekGen
    this.playGen++
    this.cancelTimer()
    const tick = this.frames[index][0]
    this.seeking = true
    this.set({ index, tick, time: this.frames[index][1], waiting: true, atEnd: false, error: null })
    this.fetch(tick).then(
      (d) => {
        if (gen !== this.seekGen || this.destroyed) return
        this.seeking = false
        this.present(d, frameDelay(this.state.speed), true)
        this.set({ waiting: false })
        this.prefetch()
        if (this.state.playing) this.schedule(this.playGen)
      },
      (err) => {
        if (gen !== this.seekGen) return
        this.seeking = false
        this.fail(err)
      },
    )
    this.prefetch()
  }

  /** Shows the frame at (or just before) a tick. */
  seekTick(tick: number) {
    if (this.frames.length) this.seek(frameAtTick(this.frames, tick))
  }

  /** One frame forward or back (pauses). */
  step(dir: number) {
    this.pause()
    this.seek(Math.max(0, this.state.index) + Math.sign(dir) * Math.max(1, Math.abs(Math.round(dir))))
  }

  play() {
    if (this.destroyed || !this.frames.length || this.state.playing) return
    // From the end, start again from the beginning of the recording.
    if (this.state.atEnd || this.state.index >= this.frames.length - 1) {
      this.set({ playing: true })
      this.seek(0)
      return
    }
    this.set({ playing: true, atEnd: false })
    if (this.state.index < 0) this.seek(0)
    // (While a seek is loading, it starts the loop once its frame is shown.)
    else if (!this.state.waiting) this.schedule(this.playGen)
  }

  pause() {
    if (!this.state.playing) return
    this.playGen++
    this.cancelTimer()
    // A frame the play loop was waiting for is dropped; one a seek is loading still shows.
    this.set({ playing: false, waiting: this.seeking })
  }

  toggle() {
    if (this.state.playing) this.pause()
    else this.play()
  }

  setSpeed(speed: ReplaySpeed) {
    if (!REPLAY_SPEEDS.includes(speed) || speed === this.state.speed) return
    this.set({ speed })
    this.prefetch()
  }

  /** Stops everything; nothing is shown after this. */
  destroy() {
    this.destroyed = true
    this.seekGen++
    this.playGen++
    this.cancelTimer()
    this.listeners.clear()
    this.cache.clear()
    this.ready.clear()
    this.buildings.clear()
  }

  // --- Playback -------------------------------------------------------------------

  private cancelTimer() {
    if (this.timer !== null) this.clock.clearTimeout(this.timer)
    this.timer = null
  }

  /** The next frame after the replay's pace. */
  private schedule(gen: number) {
    this.cancelTimer()
    this.timer = this.clock.setTimeout(() => {
      this.timer = null
      if (gen === this.playGen) this.advance(gen)
    }, frameDelay(this.state.speed))
  }

  private advance(gen: number) {
    const next = this.state.index + 1
    if (next >= this.frames.length) {
      // The newest recorded frame: pause there (the live world is just ahead).
      this.playGen++
      this.set({ playing: false, waiting: false, atEnd: true })
      return
    }
    const tick = this.frames[next][0]
    const time = this.frames[next][1]
    const waiting = !this.ready.has(tick)
    if (waiting) this.set({ waiting: true })
    this.fetch(tick).then(
      (d) => {
        if (gen !== this.playGen || this.destroyed) return
        this.present(d, frameDelay(this.state.speed), false)
        this.set({ index: next, tick, time, waiting: false })
        this.prefetch()
        this.schedule(gen)
      },
      (err) => {
        if (gen === this.playGen) this.fail(err)
      },
    )
  }

  private present(d: ReplayFrameData, interval: number, jump: boolean) {
    this.sinks.show(d.frame, interval, jump)
    this.state = { ...this.state, tick: d.frame.tick, time: d.frame.time }
    const v = d.structures
    if (!v || v === this.shownBuildings || v === this.wantedBuildings) return
    this.wantedBuildings = v
    let p = this.buildings.get(v)
    if (!p) {
      p = this.load.structures(v)
      this.buildings.set(v, p)
      // Old versions are few and small, but don't keep them all.
      if (this.buildings.size > 8) this.buildings.delete(this.buildings.keys().next().value as number)
    }
    p.then(
      (msg) => {
        if (this.destroyed || this.wantedBuildings !== v) return
        this.shownBuildings = v
        this.sinks.structures(msg.structures, v)
      },
      () => {
        this.buildings.delete(v)
        if (this.wantedBuildings === v) this.wantedBuildings = this.shownBuildings
      },
    )
  }

  private fail(err: unknown) {
    if (this.destroyed) return
    this.playGen++
    this.cancelTimer()
    this.set({ playing: false, waiting: false, error: err instanceof Error ? err.message : String(err) })
  }

  // --- Loading ahead --------------------------------------------------------------

  private fetch(tick: number): Promise<ReplayFrameData> {
    let p = this.cache.get(tick)
    if (!p) {
      this.inflight++
      p = this.load.frame(tick).then(
        (d) => {
          this.inflight--
          if (this.cache.get(tick) === p) this.ready.add(tick)
          return d
        },
        (err) => {
          this.inflight--
          if (this.cache.get(tick) === p) this.cache.delete(tick)
          throw err
        },
      )
      this.cache.set(tick, p)
      // A failed prefetch is retried when it's needed; don't report it twice.
      p.catch(() => {})
    }
    return p
  }

  /** Reads ahead of the cursor (further at higher speeds) and forgets frames far from it. */
  private prefetch() {
    if (this.destroyed || !this.frames.length) return
    const index = Math.max(0, this.state.index)
    const want = prefetchWindow(index, this.state.speed, this.frames.length, (i) => this.cache.has(this.frames[i][0]))
    for (const i of want) {
      if (this.inflight >= MAX_INFLIGHT) break
      this.fetch(this.frames[i][0])
    }
    if (this.cache.size > 64) {
      const held: number[] = []
      for (const tick of this.cache.keys()) held.push(this.indexOf.get(tick) ?? -1e9)
      const drop = new Set(evictable(held, index))
      for (const tick of [...this.cache.keys()]) {
        if (drop.has(this.indexOf.get(tick) ?? -1e9)) {
          this.cache.delete(tick)
          this.ready.delete(tick)
        }
      }
    }
  }
}
