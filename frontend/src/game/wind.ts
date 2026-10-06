// The wind as the server streams it (Weather.windDir, wind, storm) turned into
// what the 2D map draws with: a vector on the map (x east, y south, as the
// tiles run), how far smoke drifts, how rain slants. Pure maths (no DOM), so
// node tests can check it.

/** Where the wind blows towards on the map, scaled by its strength 0–1. */
export type WindVector = { x: number; y: number; strength: number }

const clamp01 = (v: number) => Math.min(1, Math.max(0, v))

/**
 * The streamed direction (radians, 0 = +x, π/2 = +y: down the 2D map, which is
 * +z in the 3D view) and strength as a vector on the map.
 */
export function windVector(dir: number, strength: number): WindVector {
  const s = clamp01(Number.isFinite(strength) ? strength : 0)
  const d = Number.isFinite(dir) ? dir : 0
  return { x: Math.cos(d) * s, y: Math.sin(d) * s, strength: s }
}

/** Whether a frame's weather carries the server's wind: older servers send neither (both 0). */
export function hasServerWind(w: { windDir?: number; wind?: number } | null | undefined): boolean {
  return !!w && ((w.wind ?? 0) > 0 || (w.windDir ?? 0) !== 0)
}

/** The signed angle (radians, −π..π) that turns `from` onto `to` the short way round. */
export function angleDelta(from: number, to: number) {
  let d = (to - from) % (Math.PI * 2)
  if (d > Math.PI) d -= Math.PI * 2
  else if (d < -Math.PI) d += Math.PI * 2
  return d
}

/** Eases an angle a fraction k of the way towards a target, the short way round. */
export function easeAngle(from: number, to: number, k: number) {
  return from + angleDelta(from, to) * clamp01(k)
}

/**
 * How far a puff of smoke of age 0–1 has drifted (world pixels) from where it
 * left a chimney or a fire: it starts straight up and bends ever more with the
 * wind. Drift along the map's y is foreshortened (the map is seen from above at
 * a slant, so smoke rising and drifting south look alike).
 */
export function smokeDrift(wind: WindVector, age: number, reach: number) {
  const a = clamp01(age)
  const bend = reach * (0.25 + wind.strength) * a * a
  return { dx: wind.x * bend * 1.6, dy: wind.y * bend * 0.6 }
}

/** Screen pixels a falling raindrop moves sideways per pixel it falls: a gale slants the rain. */
export function rainSlant(wind: WindVector) {
  return 0.08 + wind.x * 0.75
}

const COMPASS = ['timur', 'tenggara', 'selatan', 'barat daya', 'barat', 'barat laut', 'utara', 'timur laut']
const ARROWS = ['→', '↘', '↓', '↙', '←', '↖', '↑', '↗']

/** Which of eight compass points (0 east, 2 south, 4 west, 6 north) a direction is nearest. */
function octant(dir: number) {
  if (!Number.isFinite(dir)) return 0
  return Math.round((((dir % (Math.PI * 2)) + Math.PI * 2) % (Math.PI * 2)) / (Math.PI / 4)) % 8
}

/**
 * Where the wind blows towards as a compass word: 0 rad is east (+x), π/2 south
 * (+y, down the map: north is up).
 */
export function windCompass(dir: number) {
  return COMPASS[octant(dir)]
}

/** The same as an arrow on the map. */
export function windArrow(dir: number) {
  return ARROWS[octant(dir)]
}

/** Wind strength 0–1 in words (Beaufort-ish, for people on an island). */
export function windWord(strength: number, storm = false) {
  if (storm) return 'badai'
  const s = Number.isFinite(strength) ? strength : 0
  if (s < 0.1) return 'tenang'
  if (s < 0.3) return 'semilir'
  if (s < 0.55) return 'sedang'
  if (s < 0.8) return 'kencang'
  return 'sangat kencang'
}
