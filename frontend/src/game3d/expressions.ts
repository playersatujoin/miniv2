// Faces and body language from the simulation's moods, after RAGE's facial idle clips
// (FacialData.h: one idle face per mood — normal, stressed, angry, happy — blended in and
// out, with variations on a timer) and PedMotivation.h (moods that rise with events and
// ebb back towards rest). Pure math, no three.js: the crowd turns these numbers into
// vertex offsets on the face and small turns of the spine, head and arms.
//
// Moods are indexed by MOOD (0 calm, 1 fear, 2 anger, 3 joy, 4 grief).

export const MOODS = 5

/** Face channels, each about −1…1 (jaw, cheek and blink 0…1). */
export const FACE = {
  /** Inner ends of the brows: raised (+) or lowered and drawn together (−). */
  browInner: 0,
  /** Outer ends of the brows: raised (+) or drooping (−). */
  browOuter: 1,
  /** Eyelids: wide open (+) or narrowed (−). */
  eyes: 2,
  /** Mouth corners: up in a smile (+) or down (−). */
  smile: 3,
  /** Jaw: how far the mouth falls open, 0…1. */
  jaw: 4,
  /** Lips: stretched sideways (+) or pressed together (−). */
  stretch: 5,
  /** Cheeks raised, crinkling the lower lids, 0…1. */
  cheek: 6,
  /** A blink closing both eyes, 0…1. */
  blink: 7,
} as const
export const FACE_CHANNELS = 8

/** The full-strength face of each mood (FACE order, without the blink). */
export const MOOD_FACE: readonly (readonly number[])[] = [
  // Calm: the rest face.
  [0, 0, 0, 0, 0, 0, 0],
  // Fear: inner brows up and outer a little, eyes wide, lips stretched back, mouth slightly open.
  [1, 0.5, 1, -0.25, 0.4, 0.85, 0],
  // Anger: brows lowered and drawn in, eyes narrowed, lips pressed, corners down.
  [-1, 0.3, -0.55, -0.35, 0, -0.95, 0],
  // Joy: a smile raising the cheeks and crinkling the eyes.
  [0.2, 0.15, -0.35, 1, 0.2, 0.3, 1],
  // Grief: inner brows up and outer down, heavy lids, mouth corners down.
  [0.9, -0.65, -0.6, -0.9, 0.06, -0.15, 0],
]

/** Body-language channels, 0…1 each. */
export const POSTURE = {
  /** Fear: shoulders up, back rounded, arms drawn in. */
  hunch: 0,
  /** Anger: chest out, chin down, arms held away from the body. */
  chest: 1,
  /** Grief: head and shoulders slumped, arms hanging. */
  slump: 2,
  /** Joy: upright, chin up, a lighter step. */
  lift: 3,
  /** Fear: how much of the time the head darts about (quick glances). */
  glance: 4,
} as const
export const POSTURE_CHANNELS = 5

/** How fast faces follow a mood change (per second): a few tenths of a second, never a pop. */
export const MOOD_RATE = 3

const clamp = (v: number, lo: number, hi: number) => (v < lo ? lo : v > hi ? hi : v)

/** What each mood weight is heading for: the current mood at its strength, the others at 0. */
export function moodTarget(mood: number | undefined, strength: number | undefined, out: Float32Array) {
  out.fill(0)
  const m = mood ?? 0
  if (m > 0 && m < MOODS) {
    const s = Number.isFinite(strength) ? clamp(strength as number, 0, 1) : 0.6
    out[m] = s
  }
  return out
}

/**
 * Eases the mood weights (one per MOOD; weight 0 unused) towards the current mood. `k` is the
 * blend for this frame, 1 − exp(−dt·rate), computed once for everyone. Weights stay within 0…1
 * and move at most k of the way per step, so faces change continuously.
 */
export function stepMood(w: Float32Array, mood: number | undefined, strength: number | undefined, k: number) {
  const m = mood ?? 0
  const s = m > 0 && m < MOODS ? (Number.isFinite(strength) ? clamp(strength as number, 0, 1) : 0.6) : 0
  const kk = clamp(k, 0, 1)
  for (let i = 1; i < MOODS; i++) {
    const target = i === m ? s : 0
    w[i] += (target - w[i]) * kk
  }
  w[0] = 0
  return w
}

/** The blend for one frame of `dt` seconds at `rate` per second. */
export const blendFactor = (dt: number, rate = MOOD_RATE) => 1 - Math.exp(-Math.max(0, dt) * rate)

const hash = (n: number, s: number) => {
  let h = (Math.floor(n) * 374761393 + Math.floor(s * 1e6) * 668265263) | 0
  h = Math.imul(h ^ (h >>> 13), 1274126177)
  return ((h ^ (h >>> 16)) >>> 0) / 4294967296
}

/**
 * Idle blinks: one every three or four seconds at an irregular moment, sometimes twice, a
 * little more often when afraid. 0 open … 1 shut. `seed` 0–1 keeps people out of step.
 */
export function blinkAt(t: number, seed: number, fear = 0) {
  if (!Number.isFinite(t)) return 0
  const period = 3.6 - 1.4 * clamp(fear, 0, 1)
  const u = t / period + seed * 7.31
  const n = Math.floor(u)
  const f = (u - n) * period
  const at = 0.15 + (period - 0.7) * hash(n, seed)
  const length = 0.17
  let b = 0
  for (const start of hash(n, seed + 0.5) > 0.82 ? [at, at + 0.3] : [at]) {
    const d = (f - start) / length
    if (d >= 0 && d <= 1) b = Math.max(b, 1 - Math.abs(2 * d - 1))
  }
  return b
}

/**
 * The face for the blended moods (weights indexed by MOOD), plus a blink and, while talking, the
 * jaw moving with speech. Every channel is a weighted sum of bounded mood faces, then clamped.
 */
export function faceFrom(w: ArrayLike<number>, blink: number, speak: number, out: Float32Array) {
  out.fill(0)
  for (let m = 1; m < MOODS; m++) {
    const k = w[m]
    if (!k) continue
    const face = MOOD_FACE[m]
    for (let c = 0; c < 7; c++) out[c] += face[c] * k
  }
  for (let c = 0; c < 7; c++) out[c] = clamp(out[c], c === FACE.jaw || c === FACE.cheek ? 0 : -1, 1)
  out[FACE.jaw] = clamp(out[FACE.jaw] + 0.45 * clamp(speak, 0, 1), 0, 1)
  out[FACE.blink] = clamp(blink, 0, 1)
  return out
}

/** Body language for the blended moods (POSTURE order), each 0…1. */
export function postureFrom(w: ArrayLike<number>, out: Float32Array) {
  out[POSTURE.hunch] = clamp(w[1] ?? 0, 0, 1)
  out[POSTURE.chest] = clamp(w[2] ?? 0, 0, 1)
  out[POSTURE.lift] = clamp(w[3] ?? 0, 0, 1)
  out[POSTURE.slump] = clamp(w[4] ?? 0, 0, 1)
  out[POSTURE.glance] = clamp((w[1] ?? 0) * 1.2, 0, 1)
  return out
}

/** Hand gestures while talking (after GestureManager's clip sets): kinds of beat a speaker makes. */
export const GESTURE = { none: 0, open: 1, both: 2, point: 3, chest: 4 } as const

export type Gesture = { kind: number; side: 1 | -1; env: number; nod: number; speak: number }

/**
 * Talking: every couple of seconds a speaker picks a gesture (or none, listening and nodding),
 * eases into it and out again; the jaw moves in bursts of speech. Deterministic in time and seed.
 */
export function gestureAt(t: number, seed: number, out: Gesture): Gesture {
  const slot = 2.1
  const u = t / slot + seed * 3.7
  const n = Math.floor(u)
  const f = u - n
  const r = hash(n, seed + 0.25)
  const speaking = r < 0.62
  out.kind = speaking ? 1 + Math.floor(hash(n, seed + 0.75) * 4) : GESTURE.none
  out.side = hash(n, seed + 0.9) < 0.5 ? 1 : -1
  // Ease in over the first fifth, hold, ease out over the last quarter.
  const rise = clamp(f / 0.2, 0, 1)
  const fall = clamp((1 - f) / 0.25, 0, 1)
  const smooth = (x: number) => x * x * (3 - 2 * x)
  out.env = speaking ? smooth(rise) * smooth(fall) : 0
  // Listeners nod; speakers' jaws move with their syllables.
  out.nod = speaking
    ? 0.25 * Math.sin(t * 7 + seed * 11) * out.env
    : Math.max(0, Math.sin(t * 6.5 + seed * 5)) * smooth(fall) * smooth(rise)
  out.speak = speaking ? out.env * (0.5 + 0.5 * Math.sin(t * 17 + seed * 13)) * (0.6 + 0.4 * Math.sin(t * 5.3 + seed)) : 0
  return out
}
