// Wind and storms from the simulation, smoothed for drawing. After RAGE's CWind
// (game/game/wind.h): the drawn direction turns towards the simulated one at a
// bounded rate instead of snapping, and gusts come more often and harder the
// stronger it blows. Storm lightning follows the burst sequences of
// VfxLightning: a strike is a couple of short pulses, then quiet for seconds.
// Pure (no three.js, no DOM) so it can be tested in node.

/** What a stream frame says about the wind: where it blows towards (radians, 0 = +x), how hard (0–1), and a storm. */
export type WindSample = { dir: number; strength: number; storm: boolean }

/** A small seeded random generator (mulberry32), so gusts and strikes replay the same for the same seed. */
export function random(seed: number) {
  let a = seed >>> 0 || 1
  return () => {
    a = (a + 0x6d2b79f5) | 0
    let t = Math.imul(a ^ (a >>> 15), 1 | a)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

const TAU = Math.PI * 2
/** The signed short way from angle a to angle b. */
export function turn(a: number, b: number) {
  let d = (b - a) % TAU
  if (d > Math.PI) d -= TAU
  else if (d < -Math.PI) d += TAU
  return d
}

const clamp = (v: number, lo: number, hi: number) => Math.min(hi, Math.max(lo, v))
const lerp = (a: number, b: number, t: number) => a + (b - a) * t

/**
 * The wind as drawn. `dir` turns towards the simulation's at most `maxTurn` radians a second
 * (faster in a storm), `strength` eases towards it, and gusts ride on top: `blowing` is what the
 * vegetation, smoke and rain use (0 – about 1.4).
 */
export class WindTracker {
  dir = 0.42
  strength = 0.3
  /** 0–1, eased in and out over a few seconds. */
  storm = 0
  /** Extra strength of the gust blowing right now (0 when none). */
  gust = 0
  dirX = Math.cos(0.42)
  dirZ = Math.sin(0.42)
  private started = false
  private gustAt = 0
  private gustFor = 0
  private gustPeak = 0
  private readonly rand: () => number

  constructor(seed = 7) {
    this.rand = random(seed)
  }

  /** The wind with its current gust. */
  get blowing() {
    return clamp(this.strength * (1 + this.gust), 0, 1.4)
  }

  /** Advances by dt seconds towards `target` (null keeps the last wind); `calm` holds everything still. */
  update(dt: number, target: WindSample | null, calm = false) {
    if (target && !this.started) {
      // The first sample is taken as it is: nothing to turn from.
      this.started = true
      this.dir = target.dir
      this.strength = target.strength
      this.storm = target.storm ? 1 : 0
    }
    if (target && dt > 0) {
      const maxTurn = (0.25 + 0.55 * this.storm) * dt
      this.dir += clamp(turn(this.dir, target.dir) * (1 - Math.exp(-dt / 1.5)), -maxTurn, maxTurn)
      this.dir = ((this.dir % TAU) + TAU) % TAU
      this.strength += (clamp(target.strength, 0, 1) - this.strength) * (1 - Math.exp(-dt / 2.5))
      this.storm += ((target.storm ? 1 : 0) - this.storm) * (1 - Math.exp(-dt / 3))
    }
    this.dirX = Math.cos(this.dir)
    this.dirZ = Math.sin(this.dir)
    if (calm || dt <= 0) {
      this.gust = 0
      return
    }
    // Gusts: rare and soft in a breeze, frequent and strong in a gale (CWind's lo/hi gust tables).
    const hard = clamp((this.strength - 0.5) / 0.5, 0, 1) * (1 - this.storm) + this.storm
    if (this.gustFor > 0) {
      this.gustAt += dt
      const k = this.gustAt / this.gustFor
      if (k >= 1) {
        this.gustFor = 0
        this.gust = 0
      } else this.gust = this.gustPeak * Math.sin(Math.PI * k)
      return
    }
    const chancePerSecond = lerp(0.06, 0.45, hard)
    if (this.rand() < chancePerSecond * dt) {
      this.gustAt = 0
      this.gustFor = lerp(2 + this.rand(), 1 + this.rand(), hard)
      this.gustPeak = lerp(0.25 + 0.25 * this.rand(), 0.4 + 0.3 * this.rand(), hard)
    }
  }
}

/** One pulse of light in a strike: when it starts after the strike (s), how long it lasts, how bright. */
type Pulse = { at: number; length: number; peak: number }

/**
 * Lightning in a storm: every few seconds (never more often than every 3.5 s) a strike of at
 * most two pulses, each a fast flare that dies away. `flash` (0–1) is how much it lights the
 * scene right now; `strikes` counts strikes, so a bolt can be redrawn when it changes. Kept
 * to at most two flashes a second, well under the three that photosensitive viewers must not
 * see, and the caller keeps it off entirely under reduced motion.
 */
export class Lightning {
  flash = 0
  strikes = 0
  /** Where along the horizon the current bolt falls (0–1), for drawing it. */
  where = 0
  private readonly rand: () => number
  private wait = 4
  private t = 0
  private readonly pulses: Pulse[] = [
    { at: 0, length: 0, peak: 0 },
    { at: 0, length: 0, peak: 0 },
  ]
  private pulseCount = 0

  constructor(seed = 11) {
    this.rand = random(seed)
  }

  /** Advances by dt; `storming` 0–1 (strikes only above 0.5), `allowed` false fades any flash out at once. */
  update(dt: number, storming: number, allowed: boolean) {
    if (!allowed || dt <= 0) {
      this.flash = 0
      this.pulseCount = 0
      return 0
    }
    this.t += dt
    if (this.pulseCount === 0) {
      if (storming < 0.5) {
        this.wait = Math.max(this.wait, 1.5)
        this.flash = 0
        return 0
      }
      this.wait -= dt
      if (this.wait <= 0) this.strike()
    }
    let f = 0
    let live = false
    for (let i = 0; i < this.pulseCount; i++) {
      const p = this.pulses[i]
      const local = this.t - p.at
      if (local < 0) {
        live = true
        continue
      }
      if (local < p.length) {
        live = true
        // A quick flare (a fifth of the pulse), then an exponential fade.
        const rise = p.length * 0.2
        const v = local < rise ? local / rise : Math.exp(-((local - rise) / (p.length - rise)) * 3)
        f = Math.max(f, p.peak * v)
      }
    }
    if (!live) this.pulseCount = 0
    this.flash = clamp(f, 0, 1)
    return this.flash
  }

  private strike() {
    this.t = 0
    this.strikes++
    this.where = this.rand()
    this.wait = 3.5 + this.rand() * 7.5
    this.pulseCount = this.rand() < 0.55 ? 2 : 1
    this.pulses[0].at = 0
    this.pulses[0].length = 0.12 + this.rand() * 0.08
    this.pulses[0].peak = 0.75 + this.rand() * 0.25
    // A second, weaker pulse no sooner than half a second later.
    this.pulses[1].at = 0.5 + this.rand() * 0.25
    this.pulses[1].length = 0.1 + this.rand() * 0.08
    this.pulses[1].peak = 0.35 + this.rand() * 0.3
  }
}
