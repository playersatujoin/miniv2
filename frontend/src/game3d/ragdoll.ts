// A lightweight Verlet ragdoll for bodies and knockdowns, after RAGE's NaturalMotion tasks
// (TaskNMRelax for a body going limp, TaskNMFallDown, TaskNMFlinch's hit reaction and its
// balance recovery) and the ragdoll joint limits of physics/RagdollConstraints. Not a rigid-body
// solver: 17 particles (head, neck, shoulders, elbows, wrists, pelvis, hips, knees, ankles and a
// point on the chest's front and back) joined by distance constraints. The torso is rigid,
// limbs keep their lengths, joints are limited by minimum distances (elbows and knees cannot fold
// flat, the head cannot sink into the chest, thighs cannot fold into the belly) and knees and
// elbows bend only one way. Gravity, ground contact with friction, then the body settles and
// freezes. Pure math (no three.js), deterministic for the same start and time steps.
//
// Everything is in body units (the crowd's skeleton: an adult stands about 0.53 tall, faces +z,
// left hand at +x, feet on y = 0), so one solver serves every size; the caller maps the ground
// into body space.

/** Particle indices. */
export const RP = {
  head: 0,
  neck: 1,
  shL: 2,
  shR: 3,
  elL: 4,
  elR: 5,
  haL: 6,
  haR: 7,
  pelvis: 8,
  hipL: 9,
  hipR: 10,
  knL: 11,
  knR: 12,
  ftL: 13,
  ftR: 14,
  front: 15,
  back: 16,
} as const
export const RAG_N = 17

/** Where each particle sits in the rest pose (standing, arms hanging): the crowd skeleton's joints (people.ts PIVOT). */
// prettier-ignore
export const RAG_REST: readonly number[] = [
  0, 0.47, 0.003, // head (centre)
  0, 0.43, 0, // neck (the head's pivot)
  0.078, 0.4, 0, // shoulders
  -0.078, 0.4, 0,
  0.083, 0.315, 0, // elbows
  -0.083, 0.315, 0,
  0.085, 0.238, 0.004, // wrists
  -0.085, 0.238, 0.004,
  0, 0.27, 0, // pelvis (the hips' pivot)
  0.038, 0.25, 0, // hip joints
  -0.038, 0.25, 0,
  0.038, 0.135, 0, // knees
  -0.038, 0.135, 0,
  0.038, 0.026, 0, // ankles
  -0.038, 0.026, 0,
  0, 0.35, 0.045, // chest front
  0, 0.35, -0.045, // back
]

/** How far each particle's surface stands from its centre (ground contact). */
const RADIUS = [0.041, 0.02, 0.026, 0.026, 0.02, 0.02, 0.017, 0.017, 0.038, 0.032, 0.032, 0.026, 0.026, 0.02, 0.02, 0.006, 0.006]

/** Gravity in body units per second² (9.8 m/s² for a 1.65 m person 0.53 units tall). */
export const RAG_GRAVITY = 3.15

const EQUAL = 0
const ATLEAST = 1
const RANGE = 2

type Link = { a: number; b: number; kind: number; lo: number; hi: number; stiff: number }

const restDist = (a: number, b: number) =>
  Math.hypot(RAG_REST[a * 3] - RAG_REST[b * 3], RAG_REST[a * 3 + 1] - RAG_REST[b * 3 + 1], RAG_REST[a * 3 + 2] - RAG_REST[b * 3 + 2])

function buildLinks(): Link[] {
  const L: Link[] = []
  const eq = (a: number, b: number, stiff = 1) => L.push({ a, b, kind: EQUAL, lo: restDist(a, b), hi: 0, stiff })
  const atLeast = (a: number, b: number, lo: number) => L.push({ a, b, kind: ATLEAST, lo, hi: 0, stiff: 1 })
  const range = (a: number, b: number, lo: number, hi: number) => L.push({ a, b, kind: RANGE, lo, hi, stiff: 1 })
  // A rigid torso: every pair of its points.
  const torso = [RP.neck, RP.shL, RP.shR, RP.pelvis, RP.hipL, RP.hipR, RP.front, RP.back]
  for (let i = 0; i < torso.length; i++) for (let j = i + 1; j < torso.length; j++) eq(torso[i], torso[j])
  // The head on its neck, unable to bow or roll into the chest.
  eq(RP.head, RP.neck)
  atLeast(RP.head, RP.front, restDist(RP.head, RP.front) * 0.84)
  atLeast(RP.head, RP.back, restDist(RP.head, RP.back) * 0.84)
  atLeast(RP.head, RP.shL, restDist(RP.head, RP.shL) * 0.8)
  atLeast(RP.head, RP.shR, restDist(RP.head, RP.shR) * 0.8)
  // Arms: fixed lengths, an elbow that folds to about 130°.
  for (const [sh, el, ha] of [
    [RP.shL, RP.elL, RP.haL],
    [RP.shR, RP.elR, RP.haR],
  ]) {
    eq(sh, el)
    eq(el, ha)
    const reach = restDist(sh, el) + restDist(el, ha)
    range(sh, ha, reach * 0.42, reach)
  }
  // Legs: fixed lengths, knees that fold to about 140°, hips that flex to ~120° and extend ~30°.
  for (const [hip, kn, ft] of [
    [RP.hipL, RP.knL, RP.ftL],
    [RP.hipR, RP.knR, RP.ftR],
  ]) {
    eq(hip, kn)
    eq(kn, ft)
    const reach = restDist(hip, kn) + restDist(kn, ft)
    range(hip, ft, reach * 0.34, reach)
    atLeast(kn, RP.front, 0.08)
    atLeast(kn, RP.back, 0.21)
  }
  // Legs don't pass through each other.
  atLeast(RP.knL, RP.knR, 0.05)
  atLeast(RP.ftL, RP.ftR, 0.045)
  atLeast(RP.haL, RP.pelvis, 0.03)
  atLeast(RP.haR, RP.pelvis, 0.03)
  return L
}

const LINKS = buildLinks()
const LA = Int32Array.from(LINKS, (l) => l.a)
const LB = Int32Array.from(LINKS, (l) => l.b)
const LK = Int32Array.from(LINKS, (l) => l.kind)
const LLO = Float64Array.from(LINKS, (l) => l.lo)
const LHI = Float64Array.from(LINKS, (l) => l.hi)
const LS = Float64Array.from(LINKS, (l) => l.stiff)
export const RAG_LINKS = LINKS.length

/** Ground height in body units under body-space (x, z). */
export type BodyGround = (x: number, z: number) => number

export type Ragdoll = {
  pos: Float64Array
  prev: Float64Array
  /** Unspent time (fixed steps). */
  acc: number
  /** Simulated seconds so far. */
  time: number
  /** Seconds it has been almost still. */
  calm: number
  /** Settled and frozen: no more steps. */
  asleep: boolean
}

const STEP = 1 / 60
const DAMP = 0.992
const FRICTION = 0.55
/** Below this speed (units a second) a body counts as still; after SLEEP_AFTER seconds of it, it sleeps. */
const SLEEP_SPEED = 0.03
const SLEEP_AFTER = 0.35
/** No body keeps moving for longer than this. */
const MAX_TIME = 5

const hash = (n: number, s: number) => {
  let h = (Math.floor(n) * 374761393 + Math.floor(s * 1e6) * 668265263) | 0
  h = Math.imul(h ^ (h >>> 13), 1274126177)
  return ((h ^ (h >>> 16)) >>> 0) / 4294967296
}

/**
 * A ragdoll starting from `points` (RAG_N body-space positions, e.g. the pose the person was in)
 * and set going: it tips over towards `push` (a body-space x/z direction; from `seed` if absent)
 * with its knees buckling, the way a body crumples (TaskNMRelax rather than a plank toppling).
 */
export function createRagdoll(points: ArrayLike<number>, seed: number, push?: { x: number; z: number; strength?: number }): Ragdoll {
  const pos = Float64Array.from({ length: RAG_N * 3 }, (_, i) => points[i])
  const prev = Float64Array.from(pos)
  let dx: number, dz: number
  if (push && Math.hypot(push.x, push.z) > 1e-6) {
    const l = Math.hypot(push.x, push.z)
    dx = push.x / l
    dz = push.z / l
  } else {
    const a = hash(seed * 977, seed) * Math.PI * 2
    dx = Math.sin(a)
    dz = Math.cos(a)
  }
  const strength = push?.strength ?? 0.5 + 0.3 * hash(seed * 131, seed)
  const lowest = Math.min(pos[RP.ftL * 3 + 1], pos[RP.ftR * 3 + 1])
  for (let i = 0; i < RAG_N; i++) {
    const h = Math.max(0, pos[i * 3 + 1] - lowest) / 0.45
    // Velocity from tipping about the feet, plus a sag.
    const vx = dx * strength * h
    const vz = dz * strength * h
    const vy = -0.25 * h
    prev[i * 3] -= vx * STEP
    prev[i * 3 + 1] -= vy * STEP
    prev[i * 3 + 2] -= vz * STEP
  }
  // Knees buckle forward a touch so the legs fold instead of staying locked.
  const fx = pos[RP.front * 3] - pos[RP.back * 3]
  const fz = pos[RP.front * 3 + 2] - pos[RP.back * 3 + 2]
  const fl = Math.hypot(fx, fz) || 1
  for (const k of [RP.knL, RP.knR]) {
    prev[k * 3] -= (fx / fl) * 0.35 * STEP
    prev[k * 3 + 2] -= (fz / fl) * 0.35 * STEP
  }
  return { pos, prev, acc: 0, time: 0, calm: 0, asleep: false }
}

/** The rest pose as a starting point (standing, arms hanging), turned by `heading` about y (0 = facing +z). */
export function restPoints(out = new Float64Array(RAG_N * 3)) {
  for (let i = 0; i < RAG_N * 3; i++) out[i] = RAG_REST[i]
  return out
}

/**
 * Advances the ragdoll by `dt` seconds in fixed steps (at most four a call; `coarse` takes
 * half as many, larger steps with fewer iterations, for bodies far from the camera).
 */
export function stepRagdoll(r: Ragdoll, ground: BodyGround, dt: number, coarse = false) {
  if (r.asleep) return
  const h = coarse ? STEP * 2 : STEP
  const iterations = coarse ? 4 : 8
  r.acc += Math.min(Math.max(0, dt), 0.25)
  let n = 0
  while (r.acc >= h && n < 4) {
    substep(r, ground, h, iterations)
    r.acc -= h
    n++
    if (r.asleep) break
  }
  if (n === 4) r.acc = Math.min(r.acc, h)
}

/** Runs the ragdoll until it has settled (or MAX_TIME passed): for bodies first seen long after they fell, and reduced motion. */
export function settleRagdoll(r: Ragdoll, ground: BodyGround) {
  while (!r.asleep) substep(r, ground, STEP * 2, 6)
}

function substep(r: Ragdoll, ground: BodyGround, h: number, iterations: number) {
  const { pos, prev } = r
  const g = RAG_GRAVITY * h * h
  for (let i = 0; i < RAG_N; i++) {
    const o = i * 3
    const x = pos[o],
      y = pos[o + 1],
      z = pos[o + 2]
    pos[o] += (x - prev[o]) * DAMP
    pos[o + 1] += (y - prev[o + 1]) * DAMP - g
    pos[o + 2] += (z - prev[o + 2]) * DAMP
    prev[o] = x
    prev[o + 1] = y
    prev[o + 2] = z
  }
  touching.fill(0)
  for (let it = 0; it < iterations; it++) {
    solveLinks(pos)
    hinges(pos)
    collide(r, ground)
  }
  // Friction, once a step: contact eats most of the sliding.
  for (let i = 0; i < RAG_N; i++) {
    if (!touching[i]) continue
    const o = i * 3
    prev[o] += (pos[o] - prev[o]) * FRICTION
    prev[o + 2] += (pos[o + 2] - prev[o + 2]) * FRICTION
  }
  let fastest = 0
  for (let i = 0; i < RAG_N * 3; i++) fastest = Math.max(fastest, Math.abs(pos[i] - prev[i]))
  r.time += h
  r.calm = fastest < SLEEP_SPEED * h ? r.calm + h : 0
  if (r.calm >= SLEEP_AFTER || r.time >= MAX_TIME) {
    r.asleep = true
    prev.set(pos)
  }
}

function solveLinks(p: Float64Array) {
  for (let c = 0; c < LA.length; c++) {
    const a = LA[c] * 3,
      b = LB[c] * 3
    const dx = p[b] - p[a],
      dy = p[b + 1] - p[a + 1],
      dz = p[b + 2] - p[a + 2]
    const d = Math.sqrt(dx * dx + dy * dy + dz * dz) || 1e-9
    let target: number
    const kind = LK[c]
    if (kind === EQUAL) target = LLO[c]
    else if (kind === ATLEAST) {
      if (d >= LLO[c]) continue
      target = LLO[c]
    } else {
      if (d >= LLO[c] && d <= LHI[c]) continue
      target = d < LLO[c] ? LLO[c] : LHI[c]
    }
    const k = ((d - target) / d) * 0.5 * LS[c]
    p[a] += dx * k
    p[a + 1] += dy * k
    p[a + 2] += dz * k
    p[b] -= dx * k
    p[b + 1] -= dy * k
    p[b + 2] -= dz * k
  }
}

/** Knees bend forward and elbows back, relative to the chest's facing. */
function hinges(p: Float64Array) {
  let fx = p[RP.front * 3] - p[RP.back * 3],
    fy = p[RP.front * 3 + 1] - p[RP.back * 3 + 1],
    fz = p[RP.front * 3 + 2] - p[RP.back * 3 + 2]
  const fl = Math.hypot(fx, fy, fz) || 1
  fx /= fl
  fy /= fl
  fz /= fl
  for (const [a, m, b, sign] of HINGES) {
    const ao = a * 3,
      mo = m * 3,
      bo = b * 3
    const d =
      ((p[mo] - (p[ao] + p[bo]) * 0.5) * fx +
        (p[mo + 1] - (p[ao + 1] + p[bo + 1]) * 0.5) * fy +
        (p[mo + 2] - (p[ao + 2] + p[bo + 2]) * 0.5) * fz) *
      sign
    if (d >= 0) continue
    // Move the middle joint over to the right side, its ends a quarter as much the other way.
    const k = -d * sign
    p[mo] += fx * k * 0.6
    p[mo + 1] += fy * k * 0.6
    p[mo + 2] += fz * k * 0.6
    for (const e of [ao, bo]) {
      p[e] -= fx * k * 0.2
      p[e + 1] -= fy * k * 0.2
      p[e + 2] -= fz * k * 0.2
    }
  }
}

const touching = new Uint8Array(RAG_N)

const HINGES: [number, number, number, number][] = [
  [RP.hipL, RP.knL, RP.ftL, 1],
  [RP.hipR, RP.knR, RP.ftR, 1],
  [RP.shL, RP.elL, RP.haL, -1],
  [RP.shR, RP.elR, RP.haR, -1],
]

function collide(r: Ragdoll, ground: BodyGround) {
  const { pos, prev } = r
  for (let i = 0; i < RAG_N; i++) {
    const o = i * 3
    const floor = ground(pos[o], pos[o + 2]) + RADIUS[i]
    if (!(pos[o + 1] < floor)) continue
    pos[o + 1] = floor
    // No bounce.
    if (prev[o + 1] < floor) prev[o + 1] = floor
    touching[i] = 1
  }
}

/** The lowest surface point above the ground (≥ 0 when nothing is buried): for tests and checks. */
export function clearance(r: Ragdoll, ground: BodyGround) {
  let low = Infinity
  for (let i = 0; i < RAG_N; i++) {
    const o = i * 3
    low = Math.min(low, r.pos[o + 1] - RADIUS[i] - ground(r.pos[o], r.pos[o + 2]))
  }
  return low
}

// --- Back to the skeleton ------------------------------------------------------------

/**
 * The pose layout ragdollPose writes: the root's offset (x, y, z; the pelvis moved from its
 * rest spot) and then local XYZ Euler angles for hips, chest, head, upper arm and forearm
 * (left, right), thigh and shin (left, right) — the crowd's bones and rotation order.
 */
export const RAG_POSE = {
  root: 0,
  hips: 3,
  chest: 6,
  head: 9,
  armL: 12,
  foreL: 15,
  armR: 18,
  foreR: 21,
  thighL: 24,
  shinL: 27,
  thighR: 30,
  shinR: 33,
  /** The sarong: a little under half way from hanging straight to the thighs' average. */
  skirt: 36,
} as const
export const RAG_POSE_SIZE = 39

// 3×3 matrices, row-major.
type M3 = Float64Array
const m3 = () => new Float64Array(9)
const torsoM = m3(),
  headM = m3(),
  local = m3(),
  limbM = m3(),
  tmp = m3()
const va = new Float64Array(3),
  vb = new Float64Array(3)

/** Columns x, y, z. */
function fromAxes(out: M3, x: number[], y: number[], z: number[]) {
  out[0] = x[0]
  out[3] = x[1]
  out[6] = x[2]
  out[1] = y[0]
  out[4] = y[1]
  out[7] = y[2]
  out[2] = z[0]
  out[5] = z[1]
  out[8] = z[2]
  return out
}

/** out = aᵀ·v */
function mulTv(a: M3, v: ArrayLike<number>, out: Float64Array) {
  const x = v[0],
    y = v[1],
    z = v[2]
  out[0] = a[0] * x + a[3] * y + a[6] * z
  out[1] = a[1] * x + a[4] * y + a[7] * z
  out[2] = a[2] * x + a[5] * y + a[8] * z
  return out
}

function mul(a: M3, b: M3, out: M3) {
  for (let i = 0; i < 3; i++)
    for (let j = 0; j < 3; j++) tmp[i * 3 + j] = a[i * 3] * b[j] + a[i * 3 + 1] * b[3 + j] + a[i * 3 + 2] * b[6 + j]
  out.set(tmp)
  return out
}

/** The shortest rotation taking direction a onto direction b (both need not be unit). */
function swing(out: M3, a: ArrayLike<number>, b: ArrayLike<number>) {
  const al = Math.hypot(a[0], a[1], a[2]) || 1,
    bl = Math.hypot(b[0], b[1], b[2]) || 1
  const ax = a[0] / al,
    ay = a[1] / al,
    az = a[2] / al
  const bx = b[0] / bl,
    by = b[1] / bl,
    bz = b[2] / bl
  let kx = ay * bz - az * by,
    ky = az * bx - ax * bz,
    kz = ax * by - ay * bx
  const s = Math.hypot(kx, ky, kz)
  const c = ax * bx + ay * by + az * bz
  if (s < 1e-9) {
    if (c > 0) return identity(out)
    // Opposite: half a turn about any axis perpendicular to a.
    kx = Math.abs(ax) < 0.9 ? 0 : -az
    ky = Math.abs(ax) < 0.9 ? az : 0
    kz = Math.abs(ax) < 0.9 ? -ay : ax
    const l = Math.hypot(kx, ky, kz) || 1
    return axisAngle(out, kx / l, ky / l, kz / l, 0, -1)
  }
  return axisAngle(out, kx / s, ky / s, kz / s, s, c)
}

function axisAngle(out: M3, x: number, y: number, z: number, s: number, c: number) {
  const t = 1 - c
  out[0] = t * x * x + c
  out[1] = t * x * y - s * z
  out[2] = t * x * z + s * y
  out[3] = t * x * y + s * z
  out[4] = t * y * y + c
  out[5] = t * y * z - s * x
  out[6] = t * x * z - s * y
  out[7] = t * y * z + s * x
  out[8] = t * z * z + c
  return out
}

function identity(out: M3) {
  out.fill(0)
  out[0] = out[4] = out[8] = 1
  return out
}

/** XYZ Euler angles of a rotation matrix (as three.js composes them: Rx·Ry·Rz). */
export function eulerXYZ(m: ArrayLike<number>, out: Float32Array | Float64Array | number[], o: number) {
  const m13 = Math.max(-1, Math.min(1, m[2]))
  out[o + 1] = Math.asin(m13)
  if (Math.abs(m13) < 0.9999999) {
    out[o] = Math.atan2(-m[5], m[8])
    out[o + 2] = Math.atan2(-m[1], m[0])
  } else {
    out[o] = Math.atan2(m[7], m[4])
    out[o + 2] = 0
  }
}

/** Rotation matrix (row-major 3×3) from XYZ Euler angles, Rx·Ry·Rz: the inverse of eulerXYZ. */
export function matrixXYZ(x: number, y: number, z: number, out: Float64Array = m3()) {
  const a = Math.cos(x),
    b = Math.sin(x),
    c = Math.cos(y),
    d = Math.sin(y),
    e = Math.cos(z),
    f = Math.sin(z)
  out[0] = c * e
  out[1] = -c * f
  out[2] = d
  out[3] = a * f + b * e * d
  out[4] = a * e - b * f * d
  out[5] = -b * c
  out[6] = b * f - a * e * d
  out[7] = b * e + a * f * d
  out[8] = a * c
  return out
}

const sub = (p: ArrayLike<number>, i: number, j: number, out: Float64Array) => {
  out[0] = p[i * 3] - p[j * 3]
  out[1] = p[i * 3 + 1] - p[j * 3 + 1]
  out[2] = p[i * 3 + 2] - p[j * 3 + 2]
  return out
}

/** The torso's axes from its particles: left (shoulders and hips), forward (chest), up. */
function torsoFrame(p: ArrayLike<number>, out: M3) {
  let lx = p[RP.shL * 3] - p[RP.shR * 3] + p[RP.hipL * 3] - p[RP.hipR * 3]
  let ly = p[RP.shL * 3 + 1] - p[RP.shR * 3 + 1] + p[RP.hipL * 3 + 1] - p[RP.hipR * 3 + 1]
  let lz = p[RP.shL * 3 + 2] - p[RP.shR * 3 + 2] + p[RP.hipL * 3 + 2] - p[RP.hipR * 3 + 2]
  let l = Math.hypot(lx, ly, lz) || 1
  lx /= l
  ly /= l
  lz /= l
  let fx = p[RP.front * 3] - p[RP.back * 3],
    fy = p[RP.front * 3 + 1] - p[RP.back * 3 + 1],
    fz = p[RP.front * 3 + 2] - p[RP.back * 3 + 2]
  const dot = fx * lx + fy * ly + fz * lz
  fx -= lx * dot
  fy -= ly * dot
  fz -= lz * dot
  l = Math.hypot(fx, fy, fz) || 1
  fx /= l
  fy /= l
  fz /= l
  // up = forward × left
  const ux = fy * lz - fz * ly,
    uy = fz * lx - fx * lz,
    uz = fx * ly - fy * lx
  return fromAxes(out, [lx, ly, lz], [ux, uy, uz], [fx, fy, fz])
}

/** A limb bone's local rotation under its parent (world rotation `parent`), from its rest and current directions. */
function limb(p: ArrayLike<number>, from: number, to: number, parent: M3, out: Float32Array | Float64Array, o: number, world: M3) {
  sub(RAG_REST, to, from, va)
  sub(p, to, from, vb)
  mulTv(parent, vb, vb)
  swing(local, va, vb)
  eulerXYZ(local, out, o)
  return mul(parent, local, world)
}

const armM = m3()

/**
 * The skeleton's pose for the ragdoll's particles (RAG_POSE layout), so the body can be drawn with
 * the same skinned mesh as the living: the torso's rotation and where the pelvis went, then each
 * bone's swing from its parent.
 */
export function ragdollPose(p: ArrayLike<number>, out: Float32Array | Float64Array) {
  torsoFrame(p, torsoM)
  out[0] = p[RP.pelvis * 3] - RAG_REST[RP.pelvis * 3]
  out[1] = p[RP.pelvis * 3 + 1] - RAG_REST[RP.pelvis * 3 + 1]
  out[2] = p[RP.pelvis * 3 + 2] - RAG_REST[RP.pelvis * 3 + 2]
  eulerXYZ(torsoM, out, RAG_POSE.hips)
  out[RAG_POSE.chest] = out[RAG_POSE.chest + 1] = out[RAG_POSE.chest + 2] = 0
  limb(p, RP.neck, RP.head, torsoM, out, RAG_POSE.head, headM)
  limb(p, RP.shL, RP.elL, torsoM, out, RAG_POSE.armL, armM)
  limb(p, RP.elL, RP.haL, armM, out, RAG_POSE.foreL, limbM)
  limb(p, RP.shR, RP.elR, torsoM, out, RAG_POSE.armR, armM)
  limb(p, RP.elR, RP.haR, armM, out, RAG_POSE.foreR, limbM)
  limb(p, RP.hipL, RP.knL, torsoM, out, RAG_POSE.thighL, armM)
  limb(p, RP.knL, RP.ftL, armM, out, RAG_POSE.shinL, limbM)
  limb(p, RP.hipR, RP.knR, torsoM, out, RAG_POSE.thighR, armM)
  limb(p, RP.knR, RP.ftR, armM, out, RAG_POSE.shinR, limbM)
  // The skirt hangs from the hips towards the thighs' mean direction (in the hips' frame).
  sub(p, RP.knL, RP.hipL, va)
  sub(p, RP.knR, RP.hipR, vb)
  const la = Math.hypot(va[0], va[1], va[2]) || 1,
    lb = Math.hypot(vb[0], vb[1], vb[2]) || 1
  for (let k = 0; k < 3; k++) skirtDir[k] = va[k] / la + vb[k] / lb
  mulTv(torsoM, skirtDir, skirtDir)
  const l = Math.hypot(skirtDir[0], skirtDir[1], skirtDir[2]) || 1
  for (let k = 0; k < 3; k++) skirtDir[k] = DOWN[k] * 0.55 + (skirtDir[k] / l) * 0.45
  swing(local, DOWN, skirtDir)
  eulerXYZ(local, out, RAG_POSE.skirt)
  return out
}

const DOWN = [0, -1, 0]
const skirtDir = new Float64Array(3)

// --- Hit reactions -------------------------------------------------------------------

const FLINCH_W = 11
const FLINCH_Z = 0.32
const FLINCH_WD = FLINCH_W * Math.sqrt(1 - FLINCH_Z * FLINCH_Z)
const FLINCH_PEAK = (() => {
  let m = 0
  for (let t = 0; t < 1; t += 0.001) m = Math.max(m, Math.exp(-FLINCH_Z * FLINCH_W * t) * Math.sin(FLINCH_WD * t))
  return m
})()
/** How long a flinch lasts, seconds. */
export const FLINCH_TIME = 1.4

/**
 * A blow's jolt t seconds after it landed (TaskNMFlinch then balance recovery): a damped spring
 * kicked once — snapping back to a peak of 1, overshooting a little, settling within a second.
 */
export function flinch(t: number) {
  if (!(t >= 0) || t > FLINCH_TIME) return 0
  return (Math.exp(-FLINCH_Z * FLINCH_W * t) * Math.sin(FLINCH_WD * t)) / FLINCH_PEAK
}

/** The guarding that follows a blow (arms up, knees soft, a step back): up fast, easing off. */
export function guard(t: number) {
  if (!(t >= 0) || t > FLINCH_TIME) return 0
  return (1 - Math.exp(-t * 18)) * Math.exp(-t * 2.2)
}
