// Season and weather over the land in watch mode: a tint that follows the
// soil moisture (yellowed grass in the dry season, deeper green after the
// rains) and light rain streaks in a downpour. Both are smoothed so the
// screen never flickers: one simulated year lasts 8 s, so at 20× the seasons
// would flip two or three times a second.

import type { Weather } from '../sim/protocol'
import { TILE, hash, type KeyAt } from './render'

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
  private readonly reducedMotion =
    typeof window.matchMedia === 'function' ? window.matchMedia('(prefers-reduced-motion: reduce)') : null

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
    if (w && !this.primed) {
      this.primed = true
      this.moisture = w.moisture
    }
  }

  /** Forgets the weather (leaving watch mode), so the next stream starts fresh. */
  reset() {
    this.target = null
    this.primed = false
    this.moisture = MEAN_MOISTURE
    this.rain = 0
    this.lastSim = -1
  }

  /** The land changed; the tint masks are rebuilt on next use. */
  invalidateLand() {
    this.dry = null
    this.lush = null
  }

  update(dt: number) {
    const w = this.target
    if (!w) return
    const tau = this.simRate >= FAST_RATE ? TAU_FAST : TAU_NORMAL
    this.moisture += (w.moisture - this.moisture) * (1 - Math.exp(-dt / tau))
    const raining =
      this.simRate <= RAIN_MAX_RATE && !this.reducedMotion?.matches
        ? clamp01((w.rain - RAIN_FROM) / (RAIN_FULL - RAIN_FROM))
        : 0
    this.rain += (raining - this.rain) * (1 - Math.exp(-dt / 0.8))
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
    const count = Math.min(420, Math.round((this.rain * w * h) / 2600))
    ctx.strokeStyle = `rgba(205,222,245,${0.25 + 0.2 * this.rain})`
    ctx.lineWidth = 1
    ctx.beginPath()
    for (let i = 0; i < count; i++) {
      const speed = 420 + hash(i, 2, 702) * 260
      const y = ((hash(i, 1, 701) * (h + 24) + time * speed) % (h + 24)) - 12
      const x = ((hash(i, 0, 700) * (w + 40) - y * 0.22) % (w + 40) + w + 40) % (w + 40) - 20
      ctx.moveTo(x, y)
      ctx.lineTo(x - 2.4, y + 11)
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
