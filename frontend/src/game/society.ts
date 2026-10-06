// What engine adaptation II adds to the 2D map: fires (flames, glow, smoke
// drifting with the streamed wind) and the scorched land they leave, bodies
// where people died, the glyph between two people talking or bartering, and
// villages (the land they hold, outlined in their colour, and their names).
// Drawing only: everything here reads the stream (LiveWorld) and never feeds
// anything back. Bounded work per frame: only what is in view, capped counts.

import type { BurntMessage, CorpseFrame, FireFrame, Village } from '../sim/protocol'
import { FLAG } from '../sim/protocol'
import { hullTop } from '../components/sim/villageList'
import { TILE, drawLying, drawTalkGlyph, drawTradeGlyph, hash } from './render'
import { smokeDrift, type WindVector } from './wind'

type Ctx = CanvasRenderingContext2D

const clamp01 = (v: number) => Math.min(1, Math.max(0, v))

// --- Fire --------------------------------------------------------------------

/** At most this many burning tiles get a smoke column (the hottest in view). */
const MAX_SMOKE = 48
/** At most this many burning tiles are drawn (a big wildfire in view). */
const MAX_FLAMES = 400

let glowSprite: HTMLCanvasElement | null = null

/** A soft orange glow, drawn once and stretched under every fire. */
function glow() {
  if (glowSprite) return glowSprite
  const size = 64
  const c = document.createElement('canvas')
  c.width = c.height = size
  const g = c.getContext('2d')!
  const grad = g.createRadialGradient(size / 2, size / 2, 0, size / 2, size / 2, size / 2)
  grad.addColorStop(0, 'rgba(255,170,60,0.75)')
  grad.addColorStop(0.45, 'rgba(255,110,30,0.3)')
  grad.addColorStop(1, 'rgba(255,90,20,0)')
  g.fillStyle = grad
  g.fillRect(0, 0, size, size)
  glowSprite = c
  return c
}

/** Burning tiles grouped by row, for drawing between the rows of trees and people. */
export function firesByRow(fires: readonly FireFrame[] | undefined, ty0: number, ty1: number, tx0: number, tx1: number) {
  const rows = new Map<number, FireFrame[]>()
  let n = 0
  for (const f of fires ?? []) {
    if (f.y < ty0 || f.y > ty1 || f.x < tx0 - 1 || f.x > tx1 + 1) continue
    if (++n > MAX_FLAMES) break
    let row = rows.get(f.y)
    if (!row) rows.set(f.y, (row = []))
    row.push(f)
  }
  return rows
}

/** The glow a fire throws on the ground round it; stronger by night. Call under the world transform. */
export function drawFireGlow(ctx: Ctx, fires: Iterable<FireFrame>, time: number, night: number, still: boolean) {
  const g = glow()
  ctx.save()
  ctx.globalCompositeOperation = 'lighter'
  for (const f of fires) {
    const flicker = still ? 1 : 0.85 + 0.15 * Math.sin(time * 9 + hash(f.x, f.y, 41) * 20)
    const r = (22 + 26 * f.intensity) * flicker
    ctx.globalAlpha = clamp01((0.3 + 0.5 * night) * (0.4 + 0.6 * f.intensity))
    ctx.drawImage(g, (f.x + 0.5) * TILE - r, (f.y + 0.75) * TILE - r * 0.7, r * 2, r * 1.4)
  }
  ctx.restore()
}

/** One tongue of flame: a teardrop from (x, base) up to a tip bent by the wind. */
function tongue(ctx: Ctx, x: number, base: number, w: number, h: number, lean: number) {
  ctx.beginPath()
  ctx.moveTo(x - w, base)
  ctx.quadraticCurveTo(x - w * 1.1, base - h * 0.55, x + lean, base - h)
  ctx.quadraticCurveTo(x + w * 1.1, base - h * 0.55, x + w, base)
  ctx.closePath()
  ctx.fill()
}

/** Flames on a burning tile, as tall as it burns hot, flickering and leaning with the wind. */
export function drawFlames(ctx: Ctx, f: FireFrame, time: number, wind: WindVector, still: boolean) {
  const cx = (f.x + 0.5) * TILE
  const base = (f.y + 1) * TILE - 5
  const heat = clamp01(f.intensity)
  for (let i = 0; i < 3; i++) {
    const seed = hash(f.x, f.y, 50 + i)
    const flicker = still ? 1 : 0.78 + 0.16 * Math.sin(time * (9 + seed * 5) + seed * 30) + 0.08 * Math.sin(time * 23 + seed * 11)
    const h = (10 + 22 * heat) * flicker * (i === 1 ? 1 : 0.72)
    const x = cx + (i - 1) * 7 + (seed - 0.5) * 4
    const lean = wind.x * h * 0.55 + (still ? 0 : Math.sin(time * 7 + seed * 9) * 1.5)
    const w = 4 + heat * 2.5
    ctx.fillStyle = 'rgba(214,72,24,0.88)'
    tongue(ctx, x, base, w, h, lean)
    ctx.fillStyle = 'rgba(255,150,40,0.92)'
    tongue(ctx, x, base, w * 0.68, h * 0.72, lean * 0.7)
    ctx.fillStyle = 'rgba(255,232,150,0.95)'
    tongue(ctx, x, base, w * 0.36, h * 0.42, lean * 0.4)
  }
  if (!still) {
    // A few embers rising off the flames.
    for (let i = 0; i < 3; i++) {
      const seed = hash(f.x, f.y, 60 + i)
      const age = (time * (0.6 + seed * 0.5) + seed) % 1
      const ex = cx + (seed - 0.5) * 18 + wind.x * 26 * age
      const ey = base - 12 - age * (26 + 20 * heat)
      ctx.globalAlpha = (1 - age) * 0.9
      ctx.fillStyle = '#ffd27a'
      ctx.fillRect(ex, ey, 1.6, 1.6)
    }
    ctx.globalAlpha = 1
  }
}

/** Smoke columns over the hottest burning tiles in view, drifting with the wind. Call under the world transform. */
export function drawFireSmoke(ctx: Ctx, fires: readonly FireFrame[], time: number, wind: WindVector, still: boolean) {
  const shown = fires.length > MAX_SMOKE ? [...fires].sort((a, b) => b.intensity - a.intensity).slice(0, MAX_SMOKE) : fires
  ctx.save()
  const puffs = 6
  for (const f of shown) {
    const x = (f.x + 0.5) * TILE
    const y = (f.y + 1) * TILE - 22
    const rise = 52 + 40 * f.intensity
    const seed = hash(f.x, f.y, 70)
    for (let i = 0; i < puffs; i++) {
      const age = still ? (i + 0.5) / puffs : (time * 0.32 + i / puffs + seed) % 1
      const d = smokeDrift(wind, age, 70)
      const r = 5 + age * (14 + 10 * f.intensity)
      // Dark and dense over the flames, paler and thinner as it rises and spreads.
      ctx.fillStyle = age < 0.35 ? '#2f2b28' : age < 0.7 ? '#55504b' : '#7d7873'
      ctx.globalAlpha = 0.6 * Math.min(1, age * 5) * (1 - age) ** 1.1 * (0.55 + 0.45 * f.intensity)
      ctx.beginPath()
      ctx.arc(x + d.dx + Math.sin(time * 1.1 + i * 2.3 + seed * 6) * 2.5 * age, y - rise * age + d.dy, r, 0, Math.PI * 2)
      ctx.fill()
    }
  }
  ctx.restore()
}

// --- Scorched land -----------------------------------------------------------

/** One pixel per tile, as dark as the tile is burnt; drawn stretched (and softened) over the land. */
export function buildScorchMask(burnt: BurntMessage, width: number, height: number) {
  const canvas = document.createElement('canvas')
  canvas.width = Math.max(1, width)
  canvas.height = Math.max(1, height)
  const ctx = canvas.getContext('2d')!
  const img = ctx.createImageData(canvas.width, canvas.height)
  for (const t of burnt.tiles) {
    if (t.x < 0 || t.y < 0 || t.x >= width || t.y >= height) continue
    const o = (t.y * width + t.x) * 4
    img.data[o] = 27
    img.data[o + 1] = 23
    img.data[o + 2] = 20
    img.data[o + 3] = Math.round(clamp01(t.level) * 0.78 * 255)
  }
  ctx.putImageData(img, 0, 0)
  return canvas
}

/** Ash and charcoal flecks on scorched tiles in view (zoomed in). Call under the world transform. */
export function drawAsh(ctx: Ctx, tiles: BurntMessage['tiles'], tx0: number, ty0: number, tx1: number, ty1: number) {
  for (const t of tiles) {
    if (t.x < tx0 || t.x > tx1 || t.y < ty0 || t.y > ty1 || t.level < 0.15) continue
    for (let i = 0; i < 4; i++) {
      const light = i % 2 === 0
      ctx.fillStyle = light ? `rgba(170,165,158,${0.45 * t.level})` : `rgba(12,10,9,${0.6 * t.level})`
      const s = light ? 1.6 : 2.4
      ctx.fillRect(t.x * TILE + 3 + hash(t.x, t.y, 90 + i) * (TILE - 6), t.y * TILE + 3 + hash(t.x, t.y, 95 + i) * (TILE - 6), s, s)
    }
  }
}

const charred = new WeakMap<HTMLCanvasElement, HTMLCanvasElement>()

/** A tree sprite as fire leaves it: blackened (made once per sprite). */
export function charredSprite(sprite: HTMLCanvasElement) {
  let c = charred.get(sprite)
  if (!c) {
    c = document.createElement('canvas')
    c.width = sprite.width
    c.height = sprite.height
    const g = c.getContext('2d')!
    g.drawImage(sprite, 0, 0)
    g.globalCompositeOperation = 'source-atop'
    g.fillStyle = 'rgba(30,24,20,0.82)'
    g.fillRect(0, 0, c.width, c.height)
    charred.set(sprite, c)
  }
  return c
}

// --- Bodies --------------------------------------------------------------------

/** Simulated seconds a body stays (the backend's stimCorpse life) and when it starts to fade. */
const CORPSE_LIFE = 4
const CORPSE_FADE = 2.5

/** How visible a body is after `seconds`: whole at first, fading out by the time the backend drops it. */
export function corpseAlpha(seconds: number) {
  return clamp01(1 - (seconds - CORPSE_FADE) / (CORPSE_LIFE - CORPSE_FADE))
}

/** A body lying where someone died, in their colours, muted, fading with time. No blood: violence stays abstract. */
export function drawCorpse(ctx: Ctx, c: CorpseFrame) {
  drawLying(ctx, c.x * TILE, c.y * TILE, {
    heading: c.heading,
    sex: c.sex,
    hue: c.hue,
    size: c.size,
    child: c.age < 13,
    variant: c.id,
    alpha: corpseAlpha(c.seconds) * 0.92,
    dead: true,
    cause: c.cause,
  })
}

// --- Talking and bartering -------------------------------------------------------

type Talker = { id: number; rx: number; ry: number; flags: number; size: number }

/** Tiles: two people this close, both talking (or both trading), are taken to be talking to each other. */
const PARTNER_REACH = 3

/**
 * Pairs people talking or trading with whoever nearest is doing the same (the
 * stream doesn't say who with). Returns the pairs and the ids paired.
 */
export function pairExchanges<T extends Talker>(people: readonly T[], max = 80) {
  const busy = people.filter((c) => c.flags & (FLAG.talking | FLAG.trading)).slice(0, max)
  const paired = new Set<number>()
  const pairs: { a: T; b: T; trade: boolean }[] = []
  for (const a of busy) {
    if (paired.has(a.id)) continue
    const trade = (a.flags & FLAG.trading) !== 0
    let best: T | null = null
    let bestD = PARTNER_REACH
    for (const b of busy) {
      if (b === a || paired.has(b.id) || ((b.flags & FLAG.trading) !== 0) !== trade) continue
      const d = Math.hypot(b.rx - a.rx, b.ry - a.ry)
      if (d < bestD) {
        best = b
        bestD = d
      }
    }
    if (!best) continue
    paired.add(a.id)
    paired.add(best.id)
    pairs.push({ a, b: best, trade })
  }
  return { pairs, paired }
}

/** The glyph between two partners, above their heads, with faint dotted ties to each. Call under the world transform. */
export function drawExchange(ctx: Ctx, a: Talker, b: Talker, trade: boolean, time: number, scaleOf: (c: Talker) => number) {
  const ha = 44 * scaleOf(a)
  const hb = 44 * scaleOf(b)
  const ax = a.rx * TILE
  const ay = a.ry * TILE - ha
  const bx = b.rx * TILE
  const by = b.ry * TILE - hb
  const mx = (ax + bx) / 2
  const my = Math.min(ay, by) - 9
  ctx.save()
  ctx.strokeStyle = trade ? 'rgba(255,209,102,0.55)' : 'rgba(244,241,234,0.5)'
  ctx.lineWidth = 1
  ctx.setLineDash([1.5, 2.5])
  ctx.beginPath()
  ctx.moveTo(ax, ay)
  ctx.quadraticCurveTo((ax + mx) / 2, my, mx - 4, my + 2)
  ctx.moveTo(bx, by)
  ctx.quadraticCurveTo((bx + mx) / 2, my, mx + 4, my + 2)
  ctx.stroke()
  ctx.restore()
  if (trade) drawTradeGlyph(ctx, mx, my, time)
  else drawTalkGlyph(ctx, mx, my, time)
}

// --- Villages ------------------------------------------------------------------

/** Each village's land: a faint wash in its colour inside a dashed outline. Call under the world transform. */
export function drawVillageLand(
  ctx: Ctx,
  villages: readonly Village[],
  zoom: number,
  view: { x0: number; y0: number; x1: number; y1: number },
) {
  ctx.save()
  ctx.lineJoin = 'round'
  for (const v of villages) {
    const hull = v.hull ?? []
    if (hull.length < 3) continue
    let x0 = Infinity
    let y0 = Infinity
    let x1 = -Infinity
    let y1 = -Infinity
    for (const [x, y] of hull) {
      x0 = Math.min(x0, x)
      y0 = Math.min(y0, y)
      x1 = Math.max(x1, x)
      y1 = Math.max(y1, y)
    }
    if (x1 < view.x0 || x0 > view.x1 || y1 < view.y0 || y0 > view.y1) continue
    ctx.beginPath()
    hull.forEach(([x, y], i) => (i ? ctx.lineTo(x * TILE, y * TILE) : ctx.moveTo(x * TILE, y * TILE)))
    ctx.closePath()
    ctx.fillStyle = `hsla(${v.hue}, 60%, 55%, 0.07)`
    ctx.fill()
    // Two device-independent pixels wide whatever the zoom, a dark edge under the coloured dash.
    ctx.setLineDash([])
    ctx.strokeStyle = 'rgba(8,14,24,0.45)'
    ctx.lineWidth = 3.5 / zoom
    ctx.stroke()
    ctx.setLineDash([9 / zoom, 6 / zoom])
    ctx.strokeStyle = `hsla(${v.hue}, 72%, 68%, 0.9)`
    ctx.lineWidth = 2 / zoom
    ctx.stroke()
  }
  ctx.restore()
}

/**
 * Village names, in screen space (CSS pixels): over the middle of the village
 * when the map shows a wide area, at the top of its land when zoomed in (clear
 * of the crowd). `toScreen` maps tiles to CSS pixels.
 */
export function drawVillageLabels(
  ctx: Ctx,
  villages: readonly Village[],
  zoom: number,
  toScreen: (x: number, y: number) => [number, number],
  w: number,
  h: number,
) {
  ctx.textBaseline = 'middle'
  ctx.textAlign = 'left'
  for (const v of villages) {
    const [tx, ty] = zoom <= 1 ? [v.x, v.y] : hullTop(v)
    const [sx, sy0] = toScreen(tx, ty)
    const sy = zoom <= 1 ? sy0 : sy0 - 16
    if (sx < -200 || sx > w + 200 || sy < -40 || sy > h + 40) continue
    const name = `🏘 ${v.name}`
    const meta = `${v.people} jiwa${v.leader ? ` · 🚩 ${v.leader.name}` : ''}`
    ctx.font = '700 13px Inter, system-ui, sans-serif'
    const nw = ctx.measureText(name).width
    ctx.font = '500 11px Inter, system-ui, sans-serif'
    const mw = ctx.measureText(meta).width
    const bw = Math.max(nw, mw) + 18
    const bh = 36
    ctx.fillStyle = 'rgba(10,18,30,0.84)'
    ctx.strokeStyle = `hsla(${v.hue}, 65%, 60%, 0.9)`
    ctx.lineWidth = 1.2
    ctx.beginPath()
    ctx.roundRect(sx - bw / 2, sy - bh / 2, bw, bh, 8)
    ctx.fill()
    ctx.stroke()
    ctx.font = '700 13px Inter, system-ui, sans-serif'
    ctx.fillStyle = `hsl(${v.hue} 80% 80%)`
    ctx.fillText(name, sx - nw / 2, sy - 7)
    ctx.font = '500 11px Inter, system-ui, sans-serif'
    ctx.fillStyle = '#c9d3e3'
    ctx.fillText(meta, sx - mw / 2, sy + 9)
  }
}
