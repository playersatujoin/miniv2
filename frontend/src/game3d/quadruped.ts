// Four feet on uneven ground, after RAGE's quadruped leg solver
// (game/ik/solvers/QuadLegSolver.h): the body pitches to the slope between its
// fore and hind feet and rolls (partly) with the slope across them, then each
// leg stretches or shortens within limits so its foot meets the ground, and if
// a leg cannot reach, the whole body comes down. Pure math (no three.js) so it
// can be tested in node; World3D samples the ground and the animals' shader
// scales each leg about its hip.
//
// Model space is the animals' (models.ts): x forward, y up, z to the side; feet
// at (±halfLength, 0, ±halfWidth), hips straight above them at `legLength`.
// Leg order everywhere: fore +z, fore −z, hind +z, hind −z.

/** The legs as built in models.ts, per species: half the distance between fore and hind feet, half the track, leg length. */
export type QuadBuild = { halfLength: number; halfWidth: number; legLength: number }

export type QuadLimits = {
  /** Largest pitch and roll (radians). */
  pitch: number
  roll: number
  /** How much of the slope across the body it rolls with (the rest the legs take up). */
  rollShare: number
  /** How short and how long a leg may become (× its length). */
  minScale: number
  maxScale: number
}

export const QUAD_LIMITS: QuadLimits = { pitch: 0.5, roll: 0.3, rollShare: 0.6, minScale: 0.6, maxScale: 1.45 }

/** The solved pose: body height (world y of the model's origin), pitch (nose up +), roll (about +x) and each leg's length scale. */
export type QuadPose = { y: number; pitch: number; roll: number; legs: Float32Array }

export function quadPose(): QuadPose {
  return { y: 0, pitch: 0, roll: 0, legs: new Float32Array([1, 1, 1, 1]) }
}

const SIDE_X = [1, 1, -1, -1]
const SORTED = new Float64Array(4)
const TILT = new Float64Array(4)
const SIDE_Z = [1, -1, 1, -1]

/**
 * Where an animal's four feet stand in the world (x, z pairs, leg order) when it is at (x, z)
 * facing `heading` (radians, 0 = +x, as in the stream) and drawn `scale` times its model size.
 */
export function footSpots(x: number, z: number, heading: number, build: QuadBuild, scale: number, out: Float32Array) {
  const c = Math.cos(heading)
  const s = Math.sin(heading)
  for (let i = 0; i < 4; i++) {
    const lx = SIDE_X[i] * build.halfLength * scale
    const lz = SIDE_Z[i] * build.halfWidth * scale
    // The model's +x turns to the heading; its +z to the heading's right-hand side in the map (x, z).
    out[i * 2] = x + lx * c - lz * s
    out[i * 2 + 1] = z + lx * s + lz * c
  }
  return out
}

/** World y of a model-space point after the body's roll then pitch (yaw does not change heights), relative to the origin. */
export function liftOf(px: number, py: number, pz: number, pitch: number, roll: number) {
  return px * Math.sin(pitch) + (py * Math.cos(roll) - pz * Math.sin(roll)) * Math.cos(pitch)
}

/**
 * Solves the body over four ground heights (world y under each foot, leg order). `scale` is the
 * model's drawn size; the pose's legs are length factors (1 = as modelled) for the shader.
 */
export function solveQuad(ground: ArrayLike<number>, build: QuadBuild, scale: number, out: QuadPose, limits: QuadLimits = QUAD_LIMITS) {
  const hx = build.halfLength * scale
  const hz = build.halfWidth * scale
  const len = build.legLength * scale
  // The tilt follows the ground's lie, not a single foot over a ditch or up on a rock: four feet
  // on a plane satisfy g0 + g3 = g1 + g2, so when they are far from it, the foot furthest from
  // the middle two is taken as lying on the plane of the other three.
  const g = TILT
  for (let i = 0; i < 4; i++) g[i] = ground[i]
  const warp = g[0] + g[3] - g[1] - g[2]
  if (Math.abs(warp) > len * 0.3) {
    const sorted = SORTED
    sorted.set(g)
    sorted.sort()
    const mid = (sorted[1] + sorted[2]) / 2
    let odd = 0
    for (let i = 1; i < 4; i++) if (Math.abs(g[i] - mid) > Math.abs(g[odd] - mid)) odd = i
    g[odd] = odd === 0 || odd === 3 ? g[1] + g[2] - g[3 - odd] : g[0] + g[3] - g[3 - odd]
  }
  const g0 = g[0]
  const g1 = g[1]
  const g2 = g[2]
  const g3 = g[3]
  const fore = (g0 + g1) / 2
  const hind = (g2 + g3) / 2
  const plusZ = (g0 + g2) / 2
  const minusZ = (g1 + g3) / 2
  const pitch = clamp(Math.atan2(fore - hind, 2 * hx), -limits.pitch, limits.pitch)
  // A positive roll about +x lowers the +z side, so ground rising to +z rolls it the other way.
  const roll = clamp(-Math.atan2(plusZ - minusZ, 2 * hz) * limits.rollShare, -limits.roll, limits.roll)
  // Each leg's hip after the tilt, and how far its foot reaches below it per unit of leg scale.
  const reach = len * Math.cos(roll) * Math.cos(pitch)
  let y = stance(ground, hx, hz, len, pitch, roll, reach, limits, -1)
  if (Number.isNaN(y)) {
    // Not every foot can reach (one over a ditch or up on a rock): stand on the other three, the odd one out dangles or folds.
    let mean = 0
    for (let i = 0; i < 4; i++) mean += ground[i] / 4
    let odd = 0
    for (let i = 1; i < 4; i++) if (Math.abs(ground[i] - mean) > Math.abs(ground[odd] - mean)) odd = i
    y = stance(ground, hx, hz, len, pitch, roll, reach, limits, odd)
    if (Number.isNaN(y)) y = mean
  }
  for (let i = 0; i < 4; i++) {
    const hip = liftOf(SIDE_X[i] * hx, len, SIDE_Z[i] * hz, pitch, roll)
    out.legs[i] = reach > 1e-6 ? clamp((y + hip - ground[i]) / reach, limits.minScale, limits.maxScale) : 1
  }
  out.y = y
  out.pitch = pitch
  out.roll = roll
  return out
}

/** Eases a pose towards a target (rate per second), for a steady body over bumpy ground. */
export function easeQuad(pose: QuadPose, target: QuadPose, dt: number, rate = 10) {
  const k = 1 - Math.exp(-dt * rate)
  pose.y += (target.y - pose.y) * k
  pose.pitch += (target.pitch - pose.pitch) * k
  pose.roll += (target.roll - pose.roll) * k
  for (let i = 0; i < 4; i++) pose.legs[i] += (target.legs[i] - pose.legs[i]) * k
  return pose
}

/**
 * The body's height where every leg (but `skip`) reaches the ground within its limits, as near
 * as possible to where they all stand at their modelled length; NaN when no height works.
 */
function stance(
  ground: ArrayLike<number>,
  hx: number,
  hz: number,
  len: number,
  pitch: number,
  roll: number,
  reach: number,
  limits: QuadLimits,
  skip: number,
) {
  let natural = 0
  let n = 0
  let lo = -Infinity
  let hi = Infinity
  for (let i = 0; i < 4; i++) {
    if (i === skip) continue
    const hip = liftOf(SIDE_X[i] * hx, len, SIDE_Z[i] * hz, pitch, roll)
    // The origin height at which this leg, at scale k, just touches: ground − hip + k·reach.
    const base = ground[i] - hip
    natural += base + reach
    n++
    lo = Math.max(lo, base + limits.minScale * reach)
    hi = Math.min(hi, base + limits.maxScale * reach)
  }
  return lo <= hi + 1e-9 ? clamp(natural / n, lo, hi) : NaN
}

function clamp(v: number, lo: number, hi: number) {
  return Math.min(hi, Math.max(lo, v))
}
