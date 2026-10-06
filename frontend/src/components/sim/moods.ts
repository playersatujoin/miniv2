// Moods (engine adaptation II) as the observer shows them: one colour per mood,
// shared by the 2D map's emote dots and the Inspector's bars, and the words for
// a person's temperament genes. Pure (type imports only), so node tests can
// check it.

import type { MoodView } from '../../sim/protocol'

export type MoodKey = Exclude<MoodView['dominant'], ''>

/** Mood keys in MOOD code order (1 fear … 4 grief; 0 is calm). */
export const MOOD_KEYS: readonly MoodKey[] = ['fear', 'anger', 'joy', 'grief']

/**
 * Fear violet, anger red, joy aqua, grief blue: four palette hues that stay
 * apart for common colour blindness and pass 4.5:1 on the dark panel. The
 * words next to each bar carry the meaning; colour is a second cue.
 */
export const MOOD_COLOR: Record<MoodKey, string> = { fear: '#9085e9', anger: '#e66767', joy: '#199e70', grief: '#3987e5' }

/** By MOOD code (calm has none), for the map. */
export const MOOD_COLORS: readonly string[] = ['', ...MOOD_KEYS.map((k) => MOOD_COLOR[k])]

/** From this strength a mood shows on a face (as the backend's moodShown) and counts as felt. */
export const MOOD_SHOWN = 0.3

/** Noun labels for the four bars. */
export const MOOD_BAR_LABELS: Record<MoodKey, string> = { fear: 'Takut', anger: 'Marah', joy: 'Senang', grief: 'Duka' }

/**
 * The mood to name: the backend's dominant one, or (when it says calm but a
 * mood is strong anyway, e.g. an older server) the strongest above MOOD_SHOWN.
 */
export function dominantMood(m: Partial<Pick<MoodView, MoodKey | 'dominant'>> | null | undefined): MoodView['dominant'] {
  if (!m) return ''
  if (m.dominant) return m.dominant
  let best: MoodView['dominant'] = ''
  let bestV = MOOD_SHOWN
  for (const k of MOOD_KEYS) {
    const v = m[k] ?? 0
    if (v >= bestV) {
      best = k
      bestV = v
    }
  }
  return best
}

/** A temperament gene (about 1) in words: how far from the usual it is. */
export function temperWord(v: number | undefined, low: string, high: string) {
  if (v == null || !Number.isFinite(v)) return '—'
  if (v >= 1.15) return `sangat ${high}`
  if (v >= 1.05) return high
  if (v <= 0.85) return `sangat ${low}`
  if (v <= 0.95) return low
  return 'biasa'
}
