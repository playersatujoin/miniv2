// Fire for the 3D view, the pure half: how a bounded budget of flames, smoke
// puffs, embers and lights is shared among the burning tiles the stream sends,
// and the smoke and ember particles themselves, drifting with the simulation's
// wind. After RAGE's fire manager (game/vfx/misc/Fire.h): a fixed pool, each
// fire's strength scaling its effects, and lights only for the few that matter
// most. No three.js or DOM here; fireLayer.ts draws it.

/** A burning tile (as in the stream's frame.fires), intensity 0–1. */
export type FireSpot = { x: number; y: number; intensity: number }

/** How many of each effect may exist at once across every fire. */
export type FireBudget = { flames: number; smoke: number; embers: number; lights: number }

/** Fires further than this (tiles) from the camera get nothing. */
export const FIRE_REACH = 110

/** Per-fire limits: a minimum share for every fire in reach (while the budget lasts) and a cap scaled by intensity. */
const FLAME_MIN = 2
const SMOKE_MIN = 1
const LIGHT_SPACING = 5

export const hash3 = (x: number, y: number, s: number) => {
  let h = (x * 374761393 + y * 668265263 + s * 2147483647) | 0
  h = Math.imul(h ^ (h >>> 13), 1274126177)
  return ((h ^ (h >>> 16)) >>> 0) / 4294967296
}

const flameCap = (intensity: number) => 3 + Math.round(11 * intensity)
const smokeCap = (intensity: number) => 4 + Math.round(26 * intensity)
const emberCap = (intensity: number) => Math.round(10 * intensity)

/**
 * Shares the budget among fires. Each fire's weight is its intensity, softened, times how close
 * it is to the camera; the strongest and closest each get a minimum first, then what is left goes
 * out in proportion to weight (never past a fire's cap, which grows with its intensity). Lights
 * go to the heaviest fires that are not too close to one already lit. Buffers are reused between
 * calls, so planning every few frames allocates nothing once warm.
 */
export class FirePlanner {
  flames = new Int32Array(0)
  smoke = new Int32Array(0)
  embers = new Int32Array(0)
  /** Indices (into the fires planned) of the fires that get a light, strongest first. */
  lights: number[] = []
  /** Each fire's weight in the last plan (0: out of reach). */
  weight = new Float32Array(0)
  total = { flames: 0, smoke: 0, embers: 0 }
  private order = new Int32Array(0)

  plan(fires: readonly FireSpot[], camX: number, camZ: number, budget: FireBudget) {
    const n = fires.length
    if (this.weight.length < n) {
      const cap = Math.max(n, this.weight.length * 2, 16)
      this.flames = new Int32Array(cap)
      this.smoke = new Int32Array(cap)
      this.embers = new Int32Array(cap)
      this.weight = new Float32Array(cap)
      this.order = new Int32Array(cap)
    }
    let live = 0
    for (let i = 0; i < n; i++) {
      const f = fires[i]
      const d = Math.hypot(f.x + 0.5 - camX, f.y + 0.5 - camZ)
      const strength = Math.min(1, Math.max(0, f.intensity))
      const w = d > FIRE_REACH || strength < 0.02 ? 0 : Math.pow(strength, 0.8) / (1 + (d / 22) ** 2)
      this.weight[i] = w
      this.flames[i] = this.smoke[i] = this.embers[i] = 0
      if (w > 0) this.order[live++] = i
    }
    // Heaviest first (an insertion sort: fire lists are short, and mostly sorted from last time).
    const order = this.order
    const weight = this.weight
    for (let a = 1; a < live; a++) {
      const v = order[a]
      let b = a - 1
      while (b >= 0 && weight[order[b]] < weight[v]) {
        order[b + 1] = order[b]
        b--
      }
      order[b + 1] = v
    }
    this.total.flames = this.share(fires, live, Math.max(0, budget.flames | 0), this.flames, FLAME_MIN, flameCap)
    this.total.smoke = this.share(fires, live, Math.max(0, budget.smoke | 0), this.smoke, SMOKE_MIN, smokeCap)
    this.total.embers = this.share(fires, live, Math.max(0, budget.embers | 0), this.embers, 0, emberCap)
    this.lights.length = 0
    for (let a = 0; a < live && this.lights.length < budget.lights; a++) {
      const i = order[a]
      const f = fires[i]
      let near = false
      for (const j of this.lights) if (Math.hypot(fires[j].x - f.x, fires[j].y - f.y) < LIGHT_SPACING) near = true
      if (!near) this.lights.push(i)
    }
    return this
  }

  private share(fires: readonly FireSpot[], live: number, budget: number, out: Int32Array, min: number, capOf: (i: number) => number) {
    const order = this.order
    let left = budget
    let sum = 0
    for (let a = 0; a < live; a++) {
      const i = order[a]
      const give = Math.min(min, capOf(fires[i].intensity), left)
      out[i] = give
      left -= give
      sum += this.weight[i]
    }
    if (left > 0 && sum > 0) {
      const pool = left
      for (let a = 0; a < live && left > 0; a++) {
        const i = order[a]
        const room = capOf(fires[i].intensity) - out[i]
        const give = Math.min(room, left, Math.floor((pool * this.weight[i]) / sum))
        if (give > 0) {
          out[i] += give
          left -= give
        }
      }
      // Rounding leftovers, one at a time, heaviest first.
      let progress = true
      while (left > 0 && progress) {
        progress = false
        for (let a = 0; a < live && left > 0; a++) {
          const i = order[a]
          if (out[i] < capOf(fires[i].intensity)) {
            out[i]++
            left--
            progress = true
          }
        }
      }
    }
    return budget - left
  }
}

/**
 * Lays out the flame tongues of a plan: for each, the tile spot it rises from (jittered within the
 * tile, so a burning field reads as many fires), its fire's intensity and a seed for its rhythm.
 * Writes [x, z, intensity, seed] per flame into `out` and returns how many.
 */
export function layoutFlames(fires: readonly FireSpot[], plan: FirePlanner, out: Float32Array) {
  let k = 0
  const max = out.length >> 2
  for (let i = 0; i < fires.length && k < max; i++) {
    const f = fires[i]
    const n = plan.flames[i]
    for (let j = 0; j < n && k < max; j++, k++) {
      // The first tongues stand in the middle, the rest spread out.
      const spread = j === 0 ? 0.08 : 0.36
      out[k * 4] = f.x + 0.5 + (hash3(f.x, f.y, 11 + j) - 0.5) * 2 * spread
      out[k * 4 + 1] = f.y + 0.5 + (hash3(f.x, f.y, 31 + j) - 0.5) * 2 * spread
      out[k * 4 + 2] = Math.min(1, Math.max(0, f.intensity))
      out[k * 4 + 3] = hash3(f.x, f.y, 51 + j)
    }
  }
  return k
}

/** How many of a steady stream (`rate` a second, offset by `phase`) are due in (t - dt, t]: no state to keep. */
export function due(t: number, dt: number, rate: number, phase: number) {
  if (rate <= 0 || dt <= 0) return 0
  return Math.max(0, Math.floor((t + phase) * rate) - Math.floor((t - dt + phase) * rate))
}

/** A fire light's flicker, 0.6–1: a few incommensurate waves, different for every seed. */
export function flicker(t: number, seed: number) {
  const s = seed * 17.3
  const v = Math.sin(t * 13.1 + s) * 0.45 + Math.sin(t * 7.7 + s * 1.7) * 0.35 + Math.sin(t * 23.3 + s * 2.9) * 0.2
  return 0.8 + 0.2 * v
}

/** How a kind of particle moves: rising, slowing, following the wind, growing. */
export type ParticleKind = {
  /** Upward speed lost per second (fraction); embers also fall back with `gravity`. */
  lift: number
  gravity: number
  /** How quickly the particle takes on the wind's speed (per second). */
  follow: number
  /** Tiles a second the wind carries things at full strength (plus `still` at none). */
  carry: number
  still: number
  /** Side to side wander, tiles a second. */
  wander: number
  /** Size growth, tiles a second. */
  growth: number
}

export const SMOKE: ParticleKind = { lift: 0.12, gravity: 0, follow: 0.6, carry: 2.6, still: 0.15, wander: 0.12, growth: 0.24 }
export const EMBERS: ParticleKind = { lift: 0.9, gravity: 0.5, follow: 1.8, carry: 3.4, still: 0.2, wander: 0.5, growth: -0.02 }

/**
 * A fixed pool of particles in flat arrays, kept packed (the live ones are 0 … count-1, a dead
 * one swaps with the last), so drawing them is a single range. Nothing is allocated after
 * construction.
 */
export class Particles {
  readonly cap: number
  count = 0
  readonly x: Float32Array
  readonly y: Float32Array
  readonly z: Float32Array
  readonly vx: Float32Array
  readonly vy: Float32Array
  readonly vz: Float32Array
  readonly age: Float32Array
  readonly life: Float32Array
  readonly size: Float32Array
  /** How hot the fire was that made it (0–1): darker smoke, brighter embers. */
  readonly heat: Float32Array
  readonly seed: Float32Array

  constructor(cap: number) {
    this.cap = cap
    this.x = new Float32Array(cap)
    this.y = new Float32Array(cap)
    this.z = new Float32Array(cap)
    this.vx = new Float32Array(cap)
    this.vy = new Float32Array(cap)
    this.vz = new Float32Array(cap)
    this.age = new Float32Array(cap)
    this.life = new Float32Array(cap)
    this.size = new Float32Array(cap)
    this.heat = new Float32Array(cap)
    this.seed = new Float32Array(cap)
  }

  /** Adds a particle; false (and nothing added) when the pool is full. */
  emit(x: number, y: number, z: number, vy: number, life: number, size: number, heat: number, seed: number) {
    if (this.count >= this.cap) return false
    const i = this.count++
    this.x[i] = x
    this.y[i] = y
    this.z[i] = z
    this.vx[i] = 0
    this.vy[i] = vy
    this.vz[i] = 0
    this.age[i] = 0
    this.life[i] = Math.max(0.05, life)
    this.size[i] = size
    this.heat[i] = heat
    this.seed[i] = seed
    return true
  }

  /** 0–1 through its life. */
  progress(i: number) {
    return this.age[i] / this.life[i]
  }

  /**
   * Moves everything on by dt in a wind blowing towards (windX, windZ) (a unit vector) at
   * `strength` 0–1+: each particle eases towards the wind's speed, so a column bends downwind and
   * swings round when the wind turns. Spent particles are removed.
   */
  step(dt: number, windX: number, windZ: number, strength: number, kind: ParticleKind, time: number) {
    if (dt <= 0) return
    const follow = 1 - Math.exp(-dt * kind.follow)
    const slow = Math.exp(-dt * kind.lift)
    const speed = kind.still + kind.carry * strength
    const wx = windX * speed
    const wz = windZ * speed
    let i = 0
    while (i < this.count) {
      this.age[i] += dt
      if (this.age[i] >= this.life[i]) {
        this.kill(i)
        continue
      }
      const s = this.seed[i] * 40
      this.vx[i] += (wx + Math.sin(time * 1.3 + s) * kind.wander - this.vx[i]) * follow
      this.vz[i] += (wz + Math.cos(time * 1.1 + s * 1.3) * kind.wander - this.vz[i]) * follow
      this.vy[i] = this.vy[i] * slow - kind.gravity * dt
      this.x[i] += this.vx[i] * dt
      this.y[i] += this.vy[i] * dt
      this.z[i] += this.vz[i] * dt
      this.size[i] = Math.max(0.01, this.size[i] + kind.growth * dt)
      i++
    }
  }

  /** Removes particle i (the last one takes its place). */
  kill(i: number) {
    const last = --this.count
    if (i === last) return
    this.x[i] = this.x[last]
    this.y[i] = this.y[last]
    this.z[i] = this.z[last]
    this.vx[i] = this.vx[last]
    this.vy[i] = this.vy[last]
    this.vz[i] = this.vz[last]
    this.age[i] = this.age[last]
    this.life[i] = this.life[last]
    this.size[i] = this.size[last]
    this.heat[i] = this.heat[last]
    this.seed[i] = this.seed[last]
  }

  clear() {
    this.count = 0
  }
}
