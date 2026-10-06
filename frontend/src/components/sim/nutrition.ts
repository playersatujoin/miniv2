import type { NutritionLabel } from '../../sim/protocol'

/** Status colours (dataviz status palette), always shown with a word, never alone. */
export const NUTRITION_COLOR: Record<NutritionLabel, string> = {
  baik: '#0ca30c',
  kurang: '#fab219',
  buruk: '#d03b3b',
}

/** Body status this far below adequate is a severe deficiency (sim/nutrition.go deficitRange). */
const DEFICIT_RANGE = 0.6

/** How a body status (1 adequate, 0 empty) reads: the same thresholds as the server's label. */
export function statusLevel(status: number): NutritionLabel {
  const deficit = Math.min(1, Math.max(0, (1 - status) / DEFICIT_RANGE))
  if (deficit >= 0.6) return 'buruk'
  if (deficit >= 0.25) return 'kurang'
  return 'baik'
}

/** A body status as words: "cukup", "cukup + cadangan", or how far below adequate. */
export function statusText(status: number): string {
  if (status > 1.02) return 'cukup + cadangan'
  if (status >= 0.98) return 'cukup'
  return `${Math.round(Math.max(0, status) * 100)}% dari cukup`
}

/** What someone's food gives against their need (1 = enough). */
export function dietWord(ratio: number): string {
  if (ratio >= 1) return 'cukup'
  if (ratio >= 0.85) return 'hampir cukup'
  if (ratio >= 0.6) return 'kurang'
  return 'jauh kurang'
}

/** Height against a well-fed peer's, e.g. "93% — pendek (stunting)". */
export function statureText(stature: number, stunted: boolean): string {
  const p = `${Math.round(stature * 100)}%`
  if (stunted) return `${p} — pendek (stunting)`
  if (stature < 0.985) return `${p} — sedikit tertinggal`
  return `${p} — normal`
}
