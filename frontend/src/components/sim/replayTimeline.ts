// Timeline maths for the replay (after RAGE's replay controller: a cursor over
// a recorded window, jumps to markers, and a preloader that reads ahead of the
// cursor). Pure functions over the server's ReplayIndex, so node tests can
// check them. The timeline's axis is real time: the server keeps a frame every
// REPLAY_FRAME_SECONDS whatever the simulation's speed, so position i of n
// frames is i × REPLAY_FRAME_SECONDS after the oldest.

/** Real seconds between recorded frames (the server keeps one in five streamed frames, ten a second). */
export const REPLAY_FRAME_SECONDS = 0.5

/** Replay speeds offered, relative to the recording's own pace. */
export const REPLAY_SPEEDS = [0.5, 1, 2, 4] as const
export type ReplaySpeed = (typeof REPLAY_SPEEDS)[number]

/** [tick, simulated seconds] of each recorded frame, oldest first (ReplayIndex.frames). */
export type Frames = readonly (readonly [number, number])[]

/** The fields of a ReplayMark the timeline needs. */
export type Mark = { tick: number; importance: number }

/** Index of the frame at or just before tick (the oldest when tick is earlier still); -1 when nothing is recorded. */
export function frameAtTick(frames: Frames, tick: number): number {
  if (!frames.length) return -1
  let lo = 0
  let hi = frames.length - 1
  if (tick < frames[0][0]) return 0
  // The last frame whose tick is ≤ tick (ticks repeat while the world is paused: take the first of a run).
  while (lo < hi) {
    const mid = (lo + hi + 1) >> 1
    if (frames[mid][0] <= tick) lo = mid
    else hi = mid - 1
  }
  while (lo > 0 && frames[lo - 1][0] === frames[lo][0]) lo--
  return lo
}

/**
 * Where a tick falls on the timeline, 0 (oldest frame) … 1 (newest), between
 * frames in proportion to their ticks. Outside the window it clamps.
 */
export function tickToPosition(frames: Frames, tick: number): number {
  const n = frames.length
  if (n < 2) return n ? 1 : 0
  if (tick <= frames[0][0]) return 0
  if (tick >= frames[n - 1][0]) return 1
  const i = frameAtTick(frames, tick)
  const [t0] = frames[i]
  let j = i + 1
  while (j < n - 1 && frames[j][0] === t0) j++
  const t1 = frames[j][0]
  const f = t1 > t0 ? (tick - t0) / (t1 - t0) : 0
  return (i + (j - i) * f) / (n - 1)
}

/** The frame index nearest to a timeline position 0–1. */
export function positionToIndex(frames: Frames, position: number): number {
  const n = frames.length
  if (!n) return -1
  return Math.round(Math.min(1, Math.max(0, position)) * (n - 1))
}

/** The timeline position 0–1 of frame index i. */
export function indexToPosition(frames: Frames, i: number): number {
  const n = frames.length
  return n < 2 ? 1 : Math.min(1, Math.max(0, i / (n - 1)))
}

/**
 * The marker under a timeline position: of those within `tolerance` (a
 * fraction of the timeline, e.g. 6 px / width), the most important, then the
 * nearest (the timeline draws the most important of each slice, so that is the
 * one the pointer is on).
 */
export function nearestMark<M extends Mark>(frames: Frames, marks: readonly M[], position: number, tolerance: number): M | null {
  let best: M | null = null
  let bestScore = Infinity
  for (const m of marks) {
    const d = Math.abs(tickToPosition(frames, m.tick) - position)
    if (d > tolerance) continue
    const score = d / tolerance - m.importance * 2
    if (score < bestScore) {
      best = m
      bestScore = score
    }
  }
  return best
}

/**
 * Marks grouped into `buckets` equal slices of the timeline, keeping the most
 * important (then latest) of each slice and how many it stands for, so a
 * thousand events still draw as a readable row of ticks.
 */
export function bucketMarks<M extends Mark>(
  frames: Frames,
  marks: readonly M[],
  buckets: number,
): { mark: M; position: number; count: number }[] {
  const n = Math.max(1, Math.floor(buckets))
  const slots: ({ mark: M; position: number; count: number } | undefined)[] = new Array(n)
  for (const m of marks) {
    const position = tickToPosition(frames, m.tick)
    const k = Math.min(n - 1, Math.floor(position * n))
    const slot = slots[k]
    if (!slot) slots[k] = { mark: m, position, count: 1 }
    else {
      slot.count++
      if (m.importance >= slot.mark.importance) {
        slot.mark = m
        slot.position = position
      }
    }
  }
  return slots.filter((s): s is { mark: M; position: number; count: number } => !!s)
}

/** The next (dir 1) or previous (dir −1) mark of at least `importance` strictly after/before tick, or null. */
export function adjacentMark<M extends Mark>(marks: readonly M[], tick: number, dir: 1 | -1, importance = 3): M | null {
  let best: M | null = null
  for (const m of marks) {
    if (m.importance < importance) continue
    if (dir > 0 ? m.tick > tick && (!best || m.tick < best.tick) : m.tick < tick && (!best || m.tick > best.tick)) best = m
  }
  return best
}

/** Real milliseconds between frames shown at a replay speed. */
export function frameDelay(speed: number) {
  return (REPLAY_FRAME_SECONDS * 1000) / Math.max(0.1, speed)
}

/** Real seconds the frame at index i is behind the newest recorded one. */
export function secondsBehind(frames: Frames, i: number) {
  return Math.max(0, frames.length - 1 - i) * REPLAY_FRAME_SECONDS
}

/**
 * Frame indices to fetch ahead of the cursor, nearest first: enough to keep
 * `lookahead` real seconds of playback in hand at this speed (so 4× reads
 * further ahead), one frame behind (stepping back), skipping those already
 * held. At most `max`.
 */
export function prefetchWindow(
  index: number,
  speed: number,
  count: number,
  held: (i: number) => boolean,
  lookahead = 3,
  max = 24,
): number[] {
  if (index < 0 || count <= 0) return []
  const ahead = Math.min(max, Math.ceil((lookahead * Math.max(0.5, speed)) / REPLAY_FRAME_SECONDS) + 1)
  const want: number[] = []
  for (let i = index; i <= Math.min(count - 1, index + ahead); i++) if (!held(i)) want.push(i)
  if (index > 0 && !held(index - 1)) want.push(index - 1)
  return want.slice(0, max)
}

/** Which held frames to forget: those further than `behind` frames back or `ahead` frames forward of the cursor. */
export function evictable(heldIndices: Iterable<number>, index: number, behind = 12, ahead = 48): number[] {
  const out: number[] = []
  for (const i of heldIndices) if (i < index - behind || i > index + ahead) out.push(i)
  return out
}

/** "1:05" for a number of real seconds. */
export function formatClockSpan(seconds: number) {
  const s = Math.max(0, Math.round(seconds))
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`
}

/**
 * How prominent each mark is drawn: its importance, one step lower for a kind
 * that is common in the recording (more often than once every `every` seconds,
 * e.g. deaths in a famine, births in a boom, a run of climate news) so that
 * rare important events stand out; never below 1. The server's importance is
 * kept as `base`.
 */
export function rankMarks<M extends { kind: string; importance: number }>(
  marks: readonly M[],
  recordedSeconds: number,
  every = 15,
): (M & { base: number })[] {
  const counts = new Map<string, number>()
  for (const m of marks) counts.set(m.kind, (counts.get(m.kind) ?? 0) + 1)
  const limit = Math.max(1, recordedSeconds / Math.max(0.5, every))
  return marks.map((m) => {
    const common = (counts.get(m.kind) ?? 0) > limit
    const level = Math.min(3, Math.max(1, Math.round(m.importance) - (common ? 1 : 0)))
    return { ...m, importance: level, base: m.importance }
  })
}
