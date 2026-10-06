// Season and weather over the land in watch mode: a tint that follows the
// soil moisture (yellowed grass in the dry season, deeper green after the
// rains), rain streaks in a downpour slanting with the streamed wind, and
// storms: a darker sky, heavy rain and now and then a lightning flash (soft,
// never a strobe, and none with reduced motion). All of it is smoothed so the
// screen never flickers: one simulated year lasts 8 s, so at 20× the seasons
// would flip two or three times a second.

import type { Weather } from '../sim/protocol'
import { TILE, hash, type KeyAt } from './render'
import { hasServerWind, rainSlant, windVector, type WindVector } from './wind'

/** How strongly each ground kind shows the season; vegetated ground the most. */
const VEGETATION: Record<string, number> = { grass: 1, forest_floor: 0.7, dirt: 0.45, sand: 0.12 }

/** Soil moisture in an ordinary year on average (rain 1 over saturation at 1.45). */
const MEAN_MOISTURE = 0.69

const DRY_TINT: [number, number, number] = [201, 164, 71]
const LUSH_TINT: [number, number, number] = [29, 107, 44]
const DRY_ALPHA = 0.28
const LUSH_ALPHA = 0.16

/** Seconds (real time) for the tint to follow the weather; much slower when the sim runs fast. */
const TAU_NORMAL = 1.5
const TAU_FAST = 8
/** At or above this many simulated seconds per real second, the tint shows only the long-run mean. */
const FAST_RATE = 4
/** Rain streaks only at 1×–2×; faster, the showers would flash on and off. */
const RAIN_MAX_RATE = 2.5
const RAIN_FROM = 1.2 // relative rainfall where streaks start
const RAIN_FULL = 1.7
/** Lightning only up to this pace: faster, storms come and go in a blink. */
const LIGHTNING_MAX_RATE = 5
/** Real seconds between strikes, at least and at most. */
const STRIKE_GAP = [3.5, 10] as const
/** How long a strike lights the sky, seconds; the bolt itself shows for less. */
const FLASH_SECONDS = 0.5
const BOLT_SECONDS = 0.22

const clamp01 = (v: number) => Math.min(1, Math.max(0, v))

export class WeatherFx {
  private target: Weather | null = null
  /** Smoothed values that are actually shown. */
  private moisture = MEAN_MOISTURE
  private rain = 0
  private primed = false
  /** Simulated seconds per real second, estimated from the frames. */
  private simRate = 1
  private lastSim = -1
  private lastAt = 0
  private dry: HTMLCanvasElement | null = null
  private lush: HTMLCanvasElement | null = null
  /** The streamed wind (a light westerly breeze for older servers), and how stormy it looks, 0–1. */
  private wind: WindVector = windVector(0.35, 0.3)
  private storm = 0
  /** Real seconds shown so far, the next strike and the last one. */
  private clock = 0
  private nextStrike = 2
  private strike: { at: number; seed: number } | null = null
  private strikes = 0
  private readonly reducedMotion = typeof window.matchMedia === 'function' ? window.matchMedia('(prefers-reduced-motion: reduce)') : null

  /** Feeds the weather from a stream frame taken at simulated time `simTime`. */
  setWeather(w: Weather | null, simTime: number) {
    const now = performance.now()
    if (this.lastSim >= 0 && now - this.lastAt > 30) {
      const rate = Math.max(0, (simTime - this.lastSim) / ((now - this.lastAt) / 1000))
      this.simRate = this.simRate * 0.7 + rate * 0.3
    }
    this.lastSim = simTime
    this.lastAt = now
    this.target = w
    if (w && hasServerWind(w)) this.wind = windVector(w.windDir, w.wind)
    if (w && !this.primed) {
      this.primed = true
      this.moisture = w.moisture
    }
  }

  /** How hard it is raining as shown on screen, 0–1 (0 when the sim runs fast). */
  get rainLevel() {
    return this.rain
  }

  /** Forgets the weather (leaving watch mode), so the next stream starts fresh. */
  reset() {
    this.target = null
    this.primed = false
    this.moisture = MEAN_MOISTURE
    this.rain = 0
    this.storm = 0
    this.strike = null
    this.lastSim = -1
  }

  /** The land changed; the tint masks are rebuilt on next use. */
  invalidateLand() {
    this.dry = null
    this.lush = null
  }

  update(dt: number) {
    this.clock += dt
    const w = this.target
    if (!w) return
    const tau = this.simRate >= FAST_RATE ? TAU_FAST : TAU_NORMAL
    this.moisture += (w.moisture - this.moisture) * (1 - Math.exp(-dt / tau))
    const calm = this.simRate <= RAIN_MAX_RATE && !this.reducedMotion?.matches
    const raining = calm ? Math.max(clamp01((w.rain - RAIN_FROM) / (RAIN_FULL - RAIN_FROM)), w.storm ? 0.85 : 0) : 0
    this.rain += (raining - this.rain) * (1 - Math.exp(-dt / 0.8))
    // A storm darkens the sky even at speed (smoothed over a few seconds there).
    this.storm += ((w.storm ? 1 : 0) - this.storm) * (1 - Math.exp(-dt / (this.simRate >= FAST_RATE ? 4 : 1.2)))
    if (this.storm > 0.6 && this.simRate <= LIGHTNING_MAX_RATE && !this.reducedMotion?.matches) {
      if (this.clock >= this.nextStrike) {
        const seed = this.strikes++
        this.strike = { at: this.clock, seed }
        this.nextStrike = this.clock + STRIKE_GAP[0] + (STRIKE_GAP[1] - STRIKE_GAP[0]) * hash(seed, 7, 991)
      }
    } else {
      this.nextStrike = Math.max(this.nextStrike, this.clock + STRIKE_GAP[0] / 2)
    }
  }

  /** How stormy it looks, 0–1. */
  get stormLevel() {
    return this.storm
  }

  /**
   * A storm's darker sky and, now and then, lightning: a soft double flash and
   * a jagged bolt. Call with a CSS-pixel screen transform.
   */
  drawStorm(ctx: CanvasRenderingContext2D, w: number, h: number) {
    if (this.storm < 0.02) return
    ctx.fillStyle = `rgba(12,18,32,${0.2 * this.storm})`
    ctx.fillRect(0, 0, w, h)
    const s = this.strike
    if (!s) return
    const since = this.clock - s.at
    if (since > FLASH_SECONDS) return
    // Two pulses, the second fainter: a lightning flash never fills the screen.
    const pulse = since < 0.07 ? 1 : since < 0.14 ? 0.25 : since < 0.22 ? 0.7 : 0.7 * (1 - (since - 0.22) / (FLASH_SECONDS - 0.22))
    ctx.fillStyle = `rgba(220,230,255,${0.16 * pulse * this.storm})`
    ctx.fillRect(0, 0, w, h)
    if (since > BOLT_SECONDS) return
    let x = w * (0.15 + 0.7 * hash(s.seed, 1, 992))
    let y = 0
    const end = h * (0.35 + 0.35 * hash(s.seed, 2, 993))
    ctx.save()
    ctx.strokeStyle = `rgba(240,245,255,${0.85 * pulse})`
    ctx.shadowColor = 'rgba(180,200,255,0.9)'
    ctx.shadowBlur = 12
    ctx.lineWidth = 2
    ctx.lineJoin = 'round'
    ctx.beginPath()
    ctx.moveTo(x, y)
    for (let i = 0; y < end; i++) {
      y += 14 + hash(s.seed, i, 994) * 22
      x += (hash(s.seed, i, 995) - 0.5) * 34 + this.wind.x * 6
      ctx.lineTo(x, Math.min(y, end))
    }
    ctx.stroke()
    ctx.restore()
  }

  /** Washes the vegetated land with the season's colour; call under the world transform. */
  drawTint(ctx: CanvasRenderingContext2D, width: number, height: number, keyAt: KeyAt) {
    if (!this.target) return
    const dryness = clamp01((MEAN_MOISTURE - this.moisture) / 0.5)
    const lushness = clamp01((this.moisture - MEAN_MOISTURE) / (1 - MEAN_MOISTURE))
    if (dryness < 0.02 && lushness < 0.02) return
    this.dry ??= buildMask(width, height, keyAt, DRY_TINT)
    this.lush ??= buildMask(width, height, keyAt, LUSH_TINT)
    const smoothing = ctx.imageSmoothingEnabled
    ctx.imageSmoothingEnabled = false
    if (dryness >= 0.02) {
      ctx.globalAlpha = dryness * DRY_ALPHA
      ctx.drawImage(this.dry, 0, 0, width * TILE, height * TILE)
    }
    if (lushness >= 0.02) {
      ctx.globalAlpha = lushness * LUSH_ALPHA
      ctx.drawImage(this.lush, 0, 0, width * TILE, height * TILE)
    }
    ctx.globalAlpha = 1
    ctx.imageSmoothingEnabled = smoothing
  }

  /** Light rain streaks over the view; call with a CSS-pixel screen transform. */
  drawRain(ctx: CanvasRenderingContext2D, w: number, h: number, time: number) {
    if (this.rain < 0.02) return
    ctx.fillStyle = `rgba(20,32,52,${0.1 * this.rain})`
    ctx.fillRect(0, 0, w, h)
    const count = Math.min(520, Math.round((this.rain * (1 + 0.4 * this.storm) * w * h) / 2600))
    // The streaks slant with the wind: downwind as they fall.
    const slant = rainSlant(this.wind)
    const len = 11 + 4 * this.storm
    ctx.strokeStyle = `rgba(205,222,245,${0.25 + 0.2 * this.rain})`
    ctx.lineWidth = 1
    ctx.beginPath()
    for (let i = 0; i < count; i++) {
      const speed = 420 + hash(i, 2, 702) * 260
      const y = ((hash(i, 1, 701) * (h + 24) + time * speed) % (h + 24)) - 12
      const x = ((((hash(i, 0, 700) * (w + 40) + y * slant) % (w + 40)) + w + 40) % (w + 40)) - 20
      ctx.moveTo(x, y)
      ctx.lineTo(x + len * slant, y + len)
    }
    ctx.stroke()
  }
}

/** One pixel per tile: the tint colour, as opaque as the ground is vegetated. */
function buildMask(width: number, height: number, keyAt: KeyAt, [r, g, b]: [number, number, number]) {
  const canvas = document.createElement('canvas')
  canvas.width = Math.max(1, width)
  canvas.height = Math.max(1, height)
  const ctx = canvas.getContext('2d')!
  const img = ctx.createImageData(canvas.width, canvas.height)
  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      const o = (y * width + x) * 4
      img.data[o] = r
      img.data[o + 1] = g
      img.data[o + 2] = b
      img.data[o + 3] = Math.round((VEGETATION[keyAt(x, y) ?? ''] ?? 0) * 255)
    }
  }
  ctx.putImageData(img, 0, 0)
  return canvas
}
