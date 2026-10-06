// Ambient life over the island in watch mode: wind that sways trees and
// crops, cloud shadows drifting over the land, smoke from hearths and
// furnaces, the odd flock of birds, and a soft vignette. The wind is the
// server's (Weather.windDir/wind, the same wind that spreads fires) when it
// streams one, else a gentle local breeze from the rain and moisture. The rest
// is decoration driven by real time; nothing here changes the simulation.
// With "reduce motion" set in the OS everything stands still.

import { hash } from './render'
import { easeAngle, windVector, type WindVector } from './wind'

type Ctx = CanvasRenderingContext2D

const clamp01 = (v: number) => Math.min(1, Math.max(0, v))

/** A soft round shadow, drawn once and stretched for every cloud. */
function blobSprite() {
  const size = 128
  const c = document.createElement('canvas')
  c.width = c.height = size
  const g = c.getContext('2d')!
  const grad = g.createRadialGradient(size / 2, size / 2, 0, size / 2, size / 2, size / 2)
  grad.addColorStop(0, 'rgba(10,18,32,1)')
  grad.addColorStop(0.55, 'rgba(10,18,32,0.55)')
  grad.addColorStop(1, 'rgba(10,18,32,0)')
  g.fillStyle = grad
  g.fillRect(0, 0, size, size)
  return c
}

const CLOUDS = 14
const CLOUD_PARTS = 4
/** Tiles per second the clouds drift at a moderate wind. */
const CLOUD_SPEED = 0.55
/** Seconds between flocks, on average. */
const FLOCK_EVERY = 26
const FLOCK_SECONDS = 11

type Flock = { start: number; y0: number; y1: number; dir: 1 | -1; birds: number; seed: number }

/** The wind as streamed: where it blows towards (radians, 0 = +x), strength 0–1, and whether a storm rages. */
export type StreamedWind = { dir: number; strength: number; storm: boolean }

export class AmbientFx {
  /** Wind strength 0–1 and its direction (radians), shared by everything that moves with it. */
  wind = 0.3
  windDir = 0.35
  private cloudiness = 0.4
  private target = { wind: 0.3, cloudiness: 0.4 }
  /** The server's wind, or null to make up a breeze (older servers stream none). */
  private streamed: StreamedWind | null = null
  private blob: HTMLCanvasElement | null = null
  private vignette: { w: number; h: number; canvas: HTMLCanvasElement } | null = null
  private flock: Flock | null = null
  private nextFlock = 8
  private readonly reducedMotion = typeof window.matchMedia === 'function' ? window.matchMedia('(prefers-reduced-motion: reduce)') : null

  get still() {
    return this.reducedMotion?.matches ?? false
  }

  /** Feeds the weather: wetter, rainier days are cloudier; the wind is the server's when it sends one, else made up. */
  setWeather(moisture: number, rain: number, wind: StreamedWind | null = null) {
    this.streamed = wind
    const storm = wind?.storm ? 0.35 : 0
    this.target.cloudiness = clamp01(0.15 + 0.55 * moisture + 0.25 * Math.max(0, rain - 1) + storm)
    this.target.wind = wind ? clamp01(wind.strength) : clamp01(0.2 + 0.35 * Math.max(0, rain - 0.8) + 0.15 * moisture)
  }

  /** The wind now as a vector on the map. */
  get vector(): WindVector {
    return windVector(this.windDir, this.wind)
  }

  update(dt: number, time: number) {
    const k = 1 - Math.exp(-dt / 4)
    this.cloudiness += (this.target.cloudiness - this.cloudiness) * k
    // Gusts: a slow wobble on top of the weather's wind (smaller on the server's own wind).
    const gusty = this.streamed ? 0.4 : 1
    const gust = gusty * (0.12 * Math.sin(time * 0.31) + 0.08 * Math.sin(time * 0.87 + 1.3))
    this.wind += (clamp01(this.target.wind + gust) - this.wind) * k
    if (this.streamed) this.windDir = easeAngle(this.windDir, this.streamed.dir, 1 - Math.exp(-dt / 1.5))
    else this.windDir = 0.35 + 0.25 * Math.sin(time * 0.013)
    if (this.flock && time - this.flock.start > FLOCK_SECONDS) this.flock = null
    if (!this.flock && time > this.nextFlock && !this.still) {
      const seed = Math.floor(time)
      this.flock = {
        start: time,
        y0: 0.15 + 0.5 * hash(seed, 1, 901),
        y1: 0.1 + 0.6 * hash(seed, 2, 902),
        dir: hash(seed, 3, 903) < 0.5 ? 1 : -1,
        birds: 3 + Math.floor(hash(seed, 4, 904) * 6),
        seed,
      }
      this.nextFlock = time + FLOCK_SECONDS + FLOCK_EVERY * (0.5 + hash(seed, 5, 905))
    }
  }

  /**
   * How far (as a shear) a plant at tile (tx, ty) leans in the wind right now;
   * `stiffness` 1 for a tree, less for stiffer plants.
   */
  sway(tx: number, ty: number, time: number, stiffness = 1) {
    if (this.still) return 0
    const phase = hash(tx, ty, 77) * Math.PI * 2
    const lean = 0.02 + 0.07 * this.wind
    return Math.cos(this.windDir) * lean * stiffness * (0.55 + 0.45 * Math.sin(time * (1.3 + this.wind) + phase))
  }

  /** Cloud shadows over the world; call under the world transform (pixels = tiles × tile). */
  drawCloudShadows(
    ctx: Ctx,
    tile: number,
    width: number,
    height: number,
    time: number,
    view: { x0: number; y0: number; x1: number; y1: number },
  ) {
    if (this.cloudiness < 0.05) return
    this.blob ??= blobSprite()
    const drift = this.still ? 0 : time * CLOUD_SPEED * (0.5 + this.wind)
    const dx = Math.cos(this.windDir) * drift
    const dy = Math.sin(this.windDir) * drift
    const spanX = width + 40
    const spanY = height + 40
    const shown = Math.round(CLOUDS * (0.35 + 0.65 * this.cloudiness))
    ctx.save()
    for (let i = 0; i < shown; i++) {
      const cx = ((((hash(i, 0, 801) * spanX + dx) % spanX) + spanX) % spanX) - 20
      const cy = ((((hash(i, 1, 802) * spanY + dy) % spanY) + spanY) % spanY) - 20
      const r = 6 + hash(i, 2, 803) * 9
      if (cx + r * 2 < view.x0 || cx - r * 2 > view.x1 || cy + r * 2 < view.y0 || cy - r * 2 > view.y1) continue
      ctx.globalAlpha = (0.07 + 0.09 * this.cloudiness) * (0.6 + 0.4 * hash(i, 3, 804))
      for (let p = 0; p < CLOUD_PARTS; p++) {
        const ox = (hash(i, p, 810) - 0.5) * r * 1.6
        const oy = (hash(i, p, 811) - 0.5) * r * 0.9
        const pr = r * (0.55 + 0.5 * hash(i, p, 812))
        ctx.drawImage(this.blob, (cx + ox - pr) * tile, (cy + oy - pr * 0.7) * tile, pr * 2 * tile, pr * 1.4 * tile)
      }
    }
    ctx.restore()
  }

  /** A flock crossing the view now and then; call with a CSS-pixel screen transform. */
  drawBirds(ctx: Ctx, w: number, h: number, time: number) {
    const f = this.flock
    if (!f || this.still) return
    const t = (time - f.start) / FLOCK_SECONDS
    const x = f.dir === 1 ? -80 + t * (w + 160) : w + 80 - t * (w + 160)
    const y = (f.y0 + (f.y1 - f.y0) * t) * h
    ctx.save()
    ctx.strokeStyle = 'rgba(25,28,36,0.75)'
    ctx.lineWidth = 1.6
    ctx.lineCap = 'round'
    for (let b = 0; b < f.birds; b++) {
      // A loose V: each bird trails the leader a little to the side.
      const side = b % 2 === 0 ? 1 : -1
      const rank = Math.ceil(b / 2)
      const bx = x - f.dir * rank * 16 + (hash(f.seed, b, 920) - 0.5) * 6
      const by = y + side * rank * 9 + Math.sin(time * 1.7 + b) * 2
      const flap = Math.sin(time * 9 + b * 1.7)
      const span = 5.5
      ctx.beginPath()
      ctx.moveTo(bx - span, by - flap * 3)
      ctx.quadraticCurveTo(bx - span / 2, by - 1.5 - flap, bx, by)
      ctx.quadraticCurveTo(bx + span / 2, by - 1.5 - flap, bx + span, by - flap * 3)
      ctx.stroke()
    }
    ctx.restore()
  }

  /** Soft darkening at the edges of the view; call with a CSS-pixel screen transform. */
  drawVignette(ctx: Ctx, w: number, h: number) {
    if (!this.vignette || this.vignette.w !== w || this.vignette.h !== h) {
      const c = document.createElement('canvas')
      c.width = Math.max(1, Math.round(w / 4))
      c.height = Math.max(1, Math.round(h / 4))
      const g = c.getContext('2d')!
      const grad = g.createRadialGradient(
        c.width / 2,
        c.height / 2,
        Math.min(c.width, c.height) * 0.35,
        c.width / 2,
        c.height / 2,
        Math.hypot(c.width, c.height) / 2,
      )
      grad.addColorStop(0, 'rgba(6,10,20,0)')
      grad.addColorStop(1, 'rgba(6,10,20,0.38)')
      g.fillStyle = grad
      g.fillRect(0, 0, c.width, c.height)
      this.vignette = { w, h, canvas: c }
    }
    ctx.drawImage(this.vignette.canvas, 0, 0, w, h)
  }
}

/**
 * Smoke rising from (x, y) in world pixels, bent by the wind: thin and grey
 * from a hearth, thicker and darker from a furnace.
 */
export function drawSmoke(
  ctx: Ctx,
  x: number,
  y: number,
  time: number,
  seed: number,
  wind: number,
  windDir: number,
  heavy = false,
  still = false,
) {
  const puffs = heavy ? 7 : 5
  const rise = heavy ? 44 : 34
  const reach = 8 + 26 * wind
  // Drift along the map's y is foreshortened: seen from above at a slant, it reads as rising.
  const leanX = Math.cos(windDir) * reach
  const leanY = Math.sin(windDir) * reach * 0.35
  ctx.save()
  for (let i = 0; i < puffs; i++) {
    // Puffs are evenly spaced in age, so the column never thins or bunches up.
    const age = still ? (i + 0.5) / puffs : (time * (heavy ? 0.42 : 0.32) + i / puffs) % 1
    const px = x + leanX * age * age + Math.sin(time * 1.3 + i * 2.1 + seed) * 2 * age
    const py = y - rise * age + leanY * age * age
    const r = (heavy ? 3.5 : 2.6) + age * (heavy ? 8 : 6.5)
    const fade = Math.min(1, age * 4) * (1 - age) ** 1.3
    ctx.globalAlpha = (heavy ? 0.75 : 0.62) * fade
    ctx.fillStyle = heavy ? '#3f3c3a' : '#f4f1ea'
    ctx.beginPath()
    ctx.arc(px, py, r, 0, Math.PI * 2)
    ctx.fill()
  }
  ctx.restore()
}
