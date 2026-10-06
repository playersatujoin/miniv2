// Planted fields: one cached sprite per crop, stage and look. Low crops lie
// flat on their tile under everything else; grown banana, coconut and sago
// palms stand upright and are depth-sorted like trees.

import { PLOT_FLAG, type FieldPlot } from '../sim/protocol'
import { SPRITE_SIZE, TILE, hash } from './render'

type Ctx = CanvasRenderingContext2D

const RES = 2 // render at 2x so crops stay sharp when zoomed in
const VARIANTS = 2

const PADI = 0
const TALAS = 1
const UBI = 2
const PISANG = 3
const KELAPA = 4
const SAGU = 5

/** Grown banana, coconut and sago plants are trees: drawn upright, 2×2 tiles like the map's trees. */
export function plotIsUpright(p: FieldPlot) {
  return p.stage >= 2 && (p.crop === PISANG || p.crop === KELAPA || p.crop === SAGU)
}

function ellipse(ctx: Ctx, x: number, y: number, rx: number, ry: number, color: string, rot = 0) {
  ctx.fillStyle = color
  ctx.beginPath()
  ctx.ellipse(x, y, rx, ry, rot, 0, Math.PI * 2)
  ctx.fill()
}

function circle(ctx: Ctx, x: number, y: number, r: number, color: string) {
  ctx.fillStyle = color
  ctx.beginPath()
  ctx.arc(x, y, r, 0, Math.PI * 2)
  ctx.fill()
}

/** A heart-shaped leaf pointing at angle `a`, attached at (x, y). */
function heartLeaf(ctx: Ctx, x: number, y: number, size: number, a: number, color: string, vein: string) {
  ctx.save()
  ctx.translate(x, y)
  ctx.rotate(a)
  ctx.fillStyle = color
  ctx.beginPath()
  ctx.moveTo(0, 0)
  ctx.bezierCurveTo(size * 0.2, -size * 0.75, size * 1.15, -size * 0.6, size, 0)
  ctx.bezierCurveTo(size * 1.15, size * 0.6, size * 0.2, size * 0.75, 0, 0)
  ctx.fill()
  ctx.strokeStyle = vein
  ctx.lineWidth = 0.6
  ctx.beginPath()
  ctx.moveTo(0, 0)
  ctx.lineTo(size * 0.9, 0)
  ctx.stroke()
  ctx.restore()
}

/** A palm frond: a curved rib with leaflets, from (x, y) towards angle `a`. */
function frond(ctx: Ctx, x: number, y: number, len: number, a: number, droop: number, color: string, leaflet: number) {
  const ex = x + Math.cos(a) * len
  const ey = y + Math.sin(a) * len + droop
  const cx = x + Math.cos(a) * len * 0.5
  const cy = y + Math.sin(a) * len * 0.5 - droop * 0.3
  ctx.strokeStyle = color
  ctx.lineWidth = 1.2
  ctx.beginPath()
  ctx.moveTo(x, y)
  ctx.quadraticCurveTo(cx, cy, ex, ey)
  ctx.stroke()
  ctx.lineWidth = 0.9
  ctx.beginPath()
  for (let t = 0.2; t <= 0.95; t += 0.13) {
    const px = (1 - t) * (1 - t) * x + 2 * (1 - t) * t * cx + t * t * ex
    const py = (1 - t) * (1 - t) * y + 2 * (1 - t) * t * cy + t * t * ey
    const l = leaflet * (1 - t * 0.5)
    const n = a + Math.PI / 2
    ctx.moveTo(px, py)
    ctx.lineTo(px + Math.cos(n) * l + Math.cos(a) * l * 0.5, py + Math.sin(n) * l + 1.5)
    ctx.moveTo(px, py)
    ctx.lineTo(px - Math.cos(n) * l + Math.cos(a) * l * 0.5, py - Math.sin(n) * l + 1.5)
  }
  ctx.stroke()
}

// --- Flat (tile-sized) crops -------------------------------------------------

/** Ground under the crop: flooded sawah, a weeded ladang bed, or a hoed patch per plant. */
function paintBed(ctx: Ctx, flags: number, crop: number) {
  if (flags & PLOT_FLAG.irrigated && crop === PADI) {
    // Sawah: standing water inside low earth bunds.
    ctx.fillStyle = '#6b5236'
    ctx.fillRect(1, 1, TILE - 2, TILE - 2)
    ctx.fillStyle = '#4f7f9c'
    ctx.fillRect(3, 3, TILE - 6, TILE - 6)
    ctx.fillStyle = 'rgba(255,255,255,0.28)'
    ctx.fillRect(5, 6, 9, 1)
    ctx.fillRect(15, 19, 11, 1)
    return
  }
  if (flags & PLOT_FLAG.farmland) {
    ctx.fillStyle = 'rgba(94,62,36,0.85)'
    ctx.beginPath()
    ctx.roundRect(2, 2, TILE - 4, TILE - 4, 4)
    ctx.fill()
    ctx.fillStyle = 'rgba(60,38,20,0.6)'
    for (let y = 7; y < TILE - 3; y += 7) ctx.fillRect(4, y, TILE - 8, 1.5)
    return
  }
  // Planted in the wild: a dug, darker patch.
  ellipse(ctx, TILE / 2, TILE / 2 + 2, 12, 9, 'rgba(90,62,38,0.55)')
}

function paintPadi(ctx: Ctx, stage: number, v: number) {
  const ripe = stage === 3
  const h = [3, 6, 9, 9][stage]
  const color = ripe ? '#d9b443' : stage === 2 ? '#5fae3e' : '#79c24e'
  const dark = ripe ? '#a8862c' : '#3f7d2a'
  const gap = stage === 0 ? 7 : 5
  for (let y = 9, row = 0; y < TILE - 1; y += 6, row++) {
    for (let x = 4 + ((row + v) % 2) * 2; x < TILE - 3; x += gap) {
      // A tuft: three blades fanning out.
      ctx.strokeStyle = dark
      ctx.lineWidth = 1
      ctx.beginPath()
      ctx.moveTo(x, y)
      ctx.lineTo(x - 1.6, y - h * 0.8)
      ctx.moveTo(x, y)
      ctx.lineTo(x + 1.6, y - h * 0.8)
      ctx.stroke()
      ctx.strokeStyle = color
      ctx.beginPath()
      ctx.moveTo(x, y)
      ctx.lineTo(x + (ripe ? 1.5 : 0), y - h)
      ctx.stroke()
      if (ripe) circle(ctx, x + 2, y - h + 1.5, 1.2, '#f0cf5a') // drooping panicle
    }
  }
}

function paintTalas(ctx: Ctx, stage: number, v: number) {
  const n = [2, 3, 4, 4][stage]
  const size = [4, 6, 8, 8.5][stage]
  const leaf = stage === 3 ? '#4f8f3a' : '#3f8f45'
  const plants: [number, number][] =
    stage < 2 ? [[10, 13], [22, 21]] : [[TILE / 2, TILE / 2 + 2]]
  for (const [px, py] of plants) {
    if (stage === 3) ellipse(ctx, px, py + 3, 4, 2.5, '#7a4f5e') // the corm shows at the base
    for (let k = 0; k < n; k++) {
      const a = -Math.PI / 2 + (k - (n - 1) / 2) * 1.1 + (v - 0.5) * 0.3
      ctx.strokeStyle = '#6a8f3a'
      ctx.lineWidth = 1
      ctx.beginPath()
      ctx.moveTo(px, py + 2)
      ctx.lineTo(px + Math.cos(a) * size * 0.4, py + Math.sin(a) * size * 0.4)
      ctx.stroke()
      heartLeaf(ctx, px + Math.cos(a) * size * 0.3, py + Math.sin(a) * size * 0.3, size, a, leaf, '#2e6b30')
    }
  }
}

function paintUbi(ctx: Ctx, stage: number, v: number) {
  if (stage === 3) {
    // Tubers pushing up through the soil.
    ellipse(ctx, 11, 22, 4, 2.4, '#8a5a3a', 0.4)
    ellipse(ctx, 21, 18, 3.6, 2.2, '#9a6644', -0.3)
  }
  const reach = [5, 9, 13, 13][stage]
  const leaves = [2, 4, 7, 6][stage]
  ctx.strokeStyle = '#4a7d2a'
  ctx.lineWidth = 1.1
  const cx = TILE / 2
  const cy = TILE / 2 + 1
  for (let k = 0; k < 3; k++) {
    const a = (k / 3) * Math.PI * 2 + v
    ctx.beginPath()
    ctx.moveTo(cx, cy)
    ctx.bezierCurveTo(
      cx + Math.cos(a) * reach * 0.6 + 3,
      cy + Math.sin(a) * reach * 0.6 - 3,
      cx + Math.cos(a + 0.6) * reach,
      cy + Math.sin(a + 0.6) * reach * 0.7,
      cx + Math.cos(a + 0.9) * reach,
      cy + Math.sin(a + 0.9) * reach * 0.7,
    )
    ctx.stroke()
  }
  for (let k = 0; k < leaves; k++) {
    const a = (k / leaves) * Math.PI * 2 + v * 0.7
    const r = reach * (0.45 + 0.5 * ((k * 37) % 10) / 10)
    const color = stage === 3 && k % 3 === 0 ? '#b5a23f' : '#58a03a'
    heartLeaf(ctx, cx + Math.cos(a) * r, cy + Math.sin(a) * r * 0.7, 3.4, a + 1.2, color, '#2e6b30')
  }
}

/** Young banana, coconut and sago plants, before they become trees. */
function paintSapling(ctx: Ctx, crop: number, stage: number, v: number) {
  const cx = TILE / 2
  const cy = TILE / 2 + 5
  if (crop === PISANG) {
    const h = stage === 0 ? 6 : 10
    ctx.fillStyle = '#7a9a4a'
    ctx.fillRect(cx - 1.2, cy - h, 2.4, h)
    for (const [a, len] of [
      [-2.4, h],
      [-0.7, h + 1],
      [-1.6 + v * 0.2, h + 3],
    ] as const) {
      ctx.save()
      ctx.translate(cx, cy - h)
      ctx.rotate(a)
      ellipse(ctx, len / 2, 0, len / 2, 2.6, '#58a843')
      ctx.strokeStyle = '#3f7d2a'
      ctx.lineWidth = 0.6
      ctx.beginPath()
      ctx.moveTo(0, 0)
      ctx.lineTo(len, 0)
      ctx.stroke()
      ctx.restore()
    }
    return
  }
  if (crop === KELAPA && stage === 0) {
    // A sprouting coconut lying on the ground.
    ellipse(ctx, cx, cy, 5, 4, '#8a5a2b')
    ellipse(ctx, cx - 1.5, cy - 1.2, 2, 1.4, '#a8743c')
    frond(ctx, cx + 1, cy - 3, 6, -1.9, -1, '#5aa043', 1.6)
    frond(ctx, cx + 1, cy - 3, 6, -1.1, -1, '#4f9a3c', 1.6)
    return
  }
  // Palm rosettes: coconut and sago fronds fanning from the ground.
  const color = crop === SAGU ? '#3f8a3a' : '#58a043'
  const n = crop === SAGU ? 6 : 5
  const len = (stage === 0 ? 7 : 11) + (crop === SAGU ? 1 : 0)
  for (let k = 0; k < n; k++) {
    const a = -Math.PI + (k / (n - 1)) * Math.PI + (v - 0.5) * 0.2
    frond(ctx, cx, cy, len, a, 2, color, crop === SAGU ? 2.2 : 1.8)
  }
}

function paintFlatPlot(ctx: Ctx, crop: number, stage: number, flags: number, v: number) {
  paintBed(ctx, flags, crop)
  switch (crop) {
    case PADI:
      paintPadi(ctx, stage, v)
      break
    case TALAS:
      paintTalas(ctx, stage, v)
      break
    case UBI:
      paintUbi(ctx, stage, v)
      break
    default:
      paintSapling(ctx, crop, stage, v)
  }
  if (flags & PLOT_FLAG.manured) {
    for (let i = 0; i < 4; i++) circle(ctx, 5 + i * 7 + v * 2, TILE - 5 - (i % 2) * 3, 0.9, '#3b2a18')
  }
}

// --- Upright (2×2-tile) crops ------------------------------------------------

const T = TILE // horizontal centre of an upright sprite
const B = TILE * 2 // bottom edge == bottom of the plot's tile
const SHADOW = 'rgba(0,0,0,0.25)'

function paintBanana(ctx: Ctx, ripe: boolean, v: number) {
  ellipse(ctx, T, B - 6, 12, 4, SHADOW)
  ctx.fillStyle = '#8fa65a'
  ctx.fillRect(T - 2.2, B - 26, 4.4, 20)
  ctx.fillStyle = '#6f8a44'
  ctx.fillRect(T + 0.8, B - 26, 1.4, 20)
  // Big paddle leaves, some torn by the wind.
  const leaves: [number, number][] = [
    [-2.6, 15],
    [-0.5, 15],
    [-2.0, 18],
    [-1.1, 19],
    [-3.0 + v * 0.2, 12],
  ]
  for (const [a, len] of leaves) {
    ctx.save()
    ctx.translate(T, B - 27)
    ctx.rotate(a)
    ellipse(ctx, len / 2, 0, len / 2, 4.2, '#4f9e3e')
    ctx.strokeStyle = '#7cc25e'
    ctx.lineWidth = 0.8
    ctx.beginPath()
    ctx.moveTo(0, 0)
    ctx.lineTo(len, 0)
    ctx.stroke()
    ctx.restore()
  }
  if (ripe) {
    // A hanging bunch of yellow bananas and the purple heart below it.
    ctx.strokeStyle = '#6f8a44'
    ctx.lineWidth = 1.2
    ctx.beginPath()
    ctx.moveTo(T + 2, B - 25)
    ctx.quadraticCurveTo(T + 8, B - 24, T + 7, B - 14)
    ctx.stroke()
    for (let r = 0; r < 3; r++) {
      for (let k = 0; k < 3; k++) ellipse(ctx, T + 5 + k * 2, B - 22 + r * 3, 1.3, 2.4, '#f2d14a', 0.4)
    }
    ellipse(ctx, T + 7, B - 12, 1.8, 2.8, '#7a2e4a')
  }
}

function paintCoconut(ctx: Ctx, ripe: boolean, v: number) {
  ellipse(ctx, T + 4, B - 5, 12, 4, SHADOW)
  // A tall, gently curved, ringed trunk.
  const lean = v ? 5 : -4
  ctx.strokeStyle = '#8a6a48'
  ctx.lineWidth = 4
  ctx.beginPath()
  ctx.moveTo(T, B - 5)
  ctx.quadraticCurveTo(T + lean * 0.2, B - 25, T + lean, B - 46)
  ctx.stroke()
  ctx.strokeStyle = '#6b4f33'
  ctx.lineWidth = 1
  ctx.beginPath()
  for (let t = 0.1; t < 1; t += 0.14) {
    const x = T + lean * t * t
    const y = B - 5 - 41 * t
    ctx.moveTo(x - 2, y)
    ctx.lineTo(x + 2, y)
  }
  ctx.stroke()
  const tx = T + lean
  const ty = B - 46
  for (let k = 0; k < 7; k++) {
    const a = -Math.PI * 0.95 + (k / 6) * Math.PI * 0.9 + (k % 2) * 0.1
    frond(ctx, tx, ty, 16, a, 7, k % 2 ? '#3f8f38' : '#4fa043', 2.6)
  }
  if (ripe) {
    circle(ctx, tx - 2, ty + 2.5, 2.4, '#7a8a2a')
    circle(ctx, tx + 2, ty + 3, 2.4, '#8a6a2a')
    circle(ctx, tx, ty + 5, 2.4, '#6f7f28')
  }
}

function paintSago(ctx: Ctx, ripe: boolean, v: number) {
  ellipse(ctx, T, B - 5, 14, 4.5, SHADOW)
  // Short, thick trunk with old leaf bases, crowded with suckers at its foot.
  ctx.fillStyle = '#6b5a3a'
  ctx.fillRect(T - 3.5, B - 30, 7, 25)
  ctx.fillStyle = '#57482d'
  for (let y = B - 28; y < B - 6; y += 5) ctx.fillRect(T - 3.5, y, 7, 1.4)
  for (const dx of [-9, 8]) {
    for (let k = 0; k < 3; k++) frond(ctx, T + dx, B - 7, 7, -Math.PI / 2 + (k - 1) * 0.7, 2, '#4f9a3c', 1.4)
  }
  for (let k = 0; k < 9; k++) {
    const a = -Math.PI + (k / 8) * Math.PI + (v - 0.5) * 0.15
    frond(ctx, T, B - 30, 18, a, 6, k % 2 ? '#35803a' : '#44923f', 2.4)
  }
  if (ripe) {
    // The flowering spike that tells it is time to fell the trunk for its starch.
    ctx.strokeStyle = '#e8dcb0'
    ctx.lineWidth = 1.6
    ctx.beginPath()
    for (let k = -2; k <= 2; k++) {
      ctx.moveTo(T, B - 31)
      ctx.lineTo(T + k * 2.5, B - 41 + Math.abs(k))
    }
    ctx.stroke()
  }
}

// --- Cache -------------------------------------------------------------------

const cache = new Map<string, HTMLCanvasElement>()

/** Drought: browned, pale leaves. */
const WITHER = 'sepia(0.75) saturate(0.55) brightness(0.88)'

function render(size: number, key: string, paint: (ctx: Ctx) => void, withered: boolean) {
  const hit = cache.get(key)
  if (hit) return hit
  const c = document.createElement('canvas')
  c.width = c.height = size * RES
  const ctx = c.getContext('2d')!
  ctx.scale(RES, RES)
  paint(ctx)
  let out = c
  if (withered) {
    out = document.createElement('canvas')
    out.width = out.height = size * RES
    const wctx = out.getContext('2d')!
    wctx.filter = WITHER
    wctx.drawImage(c, 0, 0)
  }
  cache.set(key, out)
  return out
}

/** Flags that change how a plot looks. */
const LOOK = PLOT_FLAG.withered | PLOT_FLAG.irrigated | PLOT_FLAG.farmland | PLOT_FLAG.manured

/** Draws a low plot flat on its tile, or the bed under an upright one, at world pixel (px, py). */
export function drawPlotFlat(ctx: Ctx, p: FieldPlot, px: number, py: number) {
  const v = Math.floor(hash(p.x, p.y, 610) * VARIANTS)
  const upright = plotIsUpright(p)
  const flags = p.flags & LOOK
  const stage = upright ? -1 : p.stage
  const key = `f|${p.crop}|${stage}|${flags}|${v}`
  const sprite = render(
    TILE,
    key,
    (c) => (upright ? paintBed(c, flags, p.crop) : paintFlatPlot(c, p.crop, p.stage, flags, v)),
    (flags & PLOT_FLAG.withered) !== 0 && !upright,
  )
  ctx.drawImage(sprite, px, py, TILE, TILE)
}

/** Draws a grown banana, coconut or sago palm standing on its tile (world pixels, tile top-left). */
export function drawPlotUpright(ctx: Ctx, p: FieldPlot, px: number, py: number) {
  const v = Math.floor(hash(p.x, p.y, 610) * VARIANTS)
  const ripe = p.stage === 3
  const withered = (p.flags & PLOT_FLAG.withered) !== 0
  const key = `u|${p.crop}|${ripe ? 1 : 0}|${withered ? 1 : 0}|${v}`
  const sprite = render(
    SPRITE_SIZE,
    key,
    (c) => (p.crop === PISANG ? paintBanana(c, ripe, v) : p.crop === KELAPA ? paintCoconut(c, ripe, v) : paintSago(c, ripe, v)),
    withered,
  )
  ctx.drawImage(sprite, px - TILE / 2, py - TILE, SPRITE_SIZE, SPRITE_SIZE)
}
