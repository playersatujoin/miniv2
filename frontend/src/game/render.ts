// Procedural tile art. Everything is drawn with canvas primitives so the game
// needs no image assets; tiles are looked up by their `key` from the backend.

import { FLAG, type Sex } from '../sim/protocol'

export const TILE = 32

/** Returns the ground key at a tile, or null outside the map. */
export type KeyAt = (x: number, y: number) => string | null

/** Deterministic hash of a tile coordinate to [0, 1). */
export function hash(x: number, y: number, salt = 0): number {
  let h = Math.imul(x, 374761393) ^ Math.imul(y, 668265263) ^ Math.imul(salt, 2246822519)
  h = Math.imul(h ^ (h >>> 13), 1274126177)
  h ^= h >>> 16
  return (h >>> 0) / 4294967296
}

const shadeCache = new Map<string, string>()

/** Lightens (amount > 0) or darkens (amount < 0) a #rrggbb colour. */
export function shade(hex: string, amount: number): string {
  const cacheKey = hex + amount.toFixed(3)
  const hit = shadeCache.get(cacheKey)
  if (hit) return hit
  const n = parseInt(hex.slice(1, 7), 16)
  const mix = (c: number) =>
    Math.round(amount >= 0 ? c + (255 - c) * amount : c * (1 + amount))
  const out = `rgb(${mix(n >> 16)},${mix((n >> 8) & 255)},${mix(n & 255)})`
  shadeCache.set(cacheKey, out)
  return out
}

export const GROUND_COLORS: Record<string, string> = {
  deep_water: '#1f4e8c',
  water: '#2f74c0',
  sand: '#e2cf8f',
  grass: '#6aab4f',
  forest_floor: '#4b8a3d',
  dirt: '#a07a4f',
  stone_floor: '#8d8a83',
  mountain: '#5b5650',
  bridge: '#9b6b3c',
  batuan_vulkanik: '#4a4440',
  kawah: '#4fb8b0',
  batu_gamping: '#c9c3b4',
}

/** Ground tiles that are bare rock, so the lithology tint shows on them. */
export const ROCKY_GROUND = new Set(['mountain', 'stone_floor', 'batuan_vulkanik', 'batu_gamping'])

const WATERY = new Set(['water', 'deep_water', 'bridge'])

export function isWater(key: string | null) {
  return key === 'water' || key === 'deep_water'
}

function speckles(
  ctx: CanvasRenderingContext2D,
  tx: number,
  ty: number,
  px: number,
  py: number,
  count: number,
  colors: string[],
  w: number,
  h: number,
  salt: number,
) {
  for (let i = 0; i < count; i++) {
    const a = hash(tx, ty, salt + i * 3)
    const b = hash(tx, ty, salt + i * 3 + 1)
    const c = hash(tx, ty, salt + i * 3 + 2)
    ctx.fillStyle = colors[Math.floor(c * colors.length)]
    ctx.fillRect(px + Math.floor(a * (TILE - w)), py + Math.floor(b * (TILE - h)), w, h)
  }
}

const SIDES = [
  [0, -1],
  [1, 0],
  [0, 1],
  [-1, 0],
] as const

/** Draws one ground tile at (px, py) in an un-scaled TILE-sized cell. */
export function drawGround(
  ctx: CanvasRenderingContext2D,
  keyAt: KeyAt,
  tx: number,
  ty: number,
  px: number,
  py: number,
  fallbackColor = '#ff00ff',
) {
  const key = keyAt(tx, ty) ?? ''
  const base = GROUND_COLORS[key] ?? fallbackColor
  const v = (hash(tx, ty, 1) - 0.5) * 0.06
  ctx.fillStyle = shade(base, v)
  ctx.fillRect(px, py, TILE, TILE)

  switch (key) {
    case 'grass':
      speckles(ctx, tx, ty, px, py, 7, [shade(base, 0.14), shade(base, -0.12)], 2, 4, 10)
      break
    case 'forest_floor':
      speckles(ctx, tx, ty, px, py, 9, [shade(base, -0.18), '#6d7f35', shade(base, 0.1)], 2, 2, 20)
      break
    case 'sand':
      speckles(ctx, tx, ty, px, py, 7, [shade(base, -0.1), shade(base, 0.25)], 2, 2, 30)
      break
    case 'dirt':
      speckles(ctx, tx, ty, px, py, 5, [shade(base, -0.2), shade(base, 0.15)], 3, 2, 40)
      break
    case 'stone_floor': {
      const off = ty % 2 === 0 ? TILE / 2 : TILE / 4
      ctx.fillStyle = shade(base, -0.18)
      ctx.fillRect(px, py + TILE / 2, TILE, 1)
      ctx.fillRect(px, py + TILE - 1, TILE, 1)
      ctx.fillRect(px + off, py, 1, TILE / 2)
      ctx.fillRect(px + ((off + TILE / 2) % TILE), py + TILE / 2, 1, TILE / 2)
      speckles(ctx, tx, ty, px, py, 3, [shade(base, 0.12)], 2, 2, 50)
      break
    }
    case 'mountain': {
      ctx.fillStyle = shade(base, 0.12)
      for (let i = 0; i < 3; i++) {
        const cx = px + 4 + hash(tx, ty, 60 + i) * (TILE - 8)
        const cy = py + 6 + hash(tx, ty, 70 + i) * (TILE - 14)
        ctx.beginPath()
        ctx.moveTo(cx, cy - 5)
        ctx.lineTo(cx + 5, cy + 4)
        ctx.lineTo(cx - 5, cy + 4)
        ctx.fill()
      }
      // A cliff face where the mountain drops to lower ground.
      if (keyAt(tx, ty + 1) !== 'mountain') {
        ctx.fillStyle = shade(base, -0.3)
        ctx.fillRect(px, py + TILE - 9, TILE, 9)
        ctx.fillStyle = shade(base, -0.45)
        for (let x = 3; x < TILE; x += 7) ctx.fillRect(px + x, py + TILE - 8, 1, 7)
      }
      if (keyAt(tx, ty - 1) !== 'mountain') {
        ctx.fillStyle = shade(base, 0.22)
        ctx.fillRect(px, py, TILE, 2)
      }
      break
    }
    case 'deep_water':
    case 'water': {
      for (const [dx, dy] of SIDES) {
        const n = keyAt(tx + dx, ty + dy)
        if (n === null || WATERY.has(n)) {
          // Soften the deep/shallow boundary.
          if (key === 'deep_water' && n === 'water') {
            ctx.fillStyle = 'rgba(47,116,192,0.45)'
            edge(ctx, px, py, dx, dy, 5)
          }
          continue
        }
        ctx.fillStyle = 'rgba(255,255,255,0.45)'
        edge(ctx, px, py, dx, dy, 3)
        ctx.fillStyle = 'rgba(255,255,255,0.18)'
        edge(ctx, px, py, dx, dy, 6)
      }
      break
    }
    case 'bridge':
      drawBridge(ctx, keyAt, tx, ty, px, py)
      break
    case 'batuan_vulkanik': {
      // Basaltic scoria: dark vesicles and a few glassy glints.
      ctx.fillStyle = shade(base, -0.35)
      for (let i = 0; i < 6; i++) {
        const cx = px + 3 + hash(tx, ty, 80 + i) * (TILE - 6)
        const cy = py + 3 + hash(tx, ty, 90 + i) * (TILE - 6)
        ctx.beginPath()
        ctx.arc(cx, cy, 1 + hash(tx, ty, 100 + i) * 1.2, 0, Math.PI * 2)
        ctx.fill()
      }
      speckles(ctx, tx, ty, px, py, 4, [shade(base, 0.25), '#6b5446'], 2, 1, 110)
      // Cooling cracks.
      ctx.strokeStyle = shade(base, -0.45)
      ctx.lineWidth = 1
      ctx.beginPath()
      const sx = px + hash(tx, ty, 120) * TILE
      ctx.moveTo(sx, py)
      ctx.lineTo(px + hash(tx, ty, 121) * TILE, py + TILE / 2)
      ctx.lineTo(px + hash(tx, ty, 122) * TILE, py + TILE)
      ctx.stroke()
      break
    }
    case 'kawah': {
      // Acid crater lake with a sulfur crust where it meets the rim.
      ctx.fillStyle = shade(base, 0.12)
      for (let i = 0; i < 3; i++) {
        const y = py + 6 + hash(tx, ty, 130 + i) * (TILE - 12)
        ctx.fillRect(px + 4 + hash(tx, ty, 140 + i) * 10, y, 10, 1)
      }
      for (const [dx, dy] of SIDES) {
        if (keyAt(tx + dx, ty + dy) === 'kawah') continue
        ctx.fillStyle = '#e9d43f'
        edge(ctx, px, py, dx, dy, 4)
        ctx.fillStyle = 'rgba(120,96,20,0.55)'
        edge(ctx, px, py, dx, dy, 1)
      }
      break
    }
    case 'batu_gamping': {
      // Karst limestone: pale clints split by dark grikes, a few solution pits.
      ctx.strokeStyle = shade(base, -0.28)
      ctx.lineWidth = 1
      ctx.beginPath()
      const gx = px + 8 + hash(tx, ty, 150) * 16
      ctx.moveTo(gx, py)
      ctx.lineTo(gx + (hash(tx, ty, 151) - 0.5) * 8, py + TILE / 2)
      ctx.lineTo(gx + (hash(tx, ty, 152) - 0.5) * 8, py + TILE)
      const gy = py + 8 + hash(tx, ty, 153) * 16
      ctx.moveTo(px, gy)
      ctx.lineTo(px + TILE, gy + (hash(tx, ty, 154) - 0.5) * 6)
      ctx.stroke()
      speckles(ctx, tx, ty, px, py, 3, [shade(base, -0.2)], 2, 2, 160)
      speckles(ctx, tx, ty, px, py, 4, [shade(base, 0.2)], 2, 1, 170)
      break
    }
  }
}

/** Washes a rock tile with its rock unit's colour so the bedrock shows through. */
export function tintLithology(ctx: CanvasRenderingContext2D, color: string, px: number, py: number) {
  ctx.globalAlpha = 0.16
  ctx.fillStyle = color
  ctx.fillRect(px, py, TILE, TILE)
  ctx.globalAlpha = 1
}

/** Fills a strip of `size` pixels along one side of a tile. */
function edge(ctx: CanvasRenderingContext2D, px: number, py: number, dx: number, dy: number, size: number) {
  if (dy === -1) ctx.fillRect(px, py, TILE, size)
  else if (dy === 1) ctx.fillRect(px, py + TILE - size, TILE, size)
  else if (dx === -1) ctx.fillRect(px, py, size, TILE)
  else ctx.fillRect(px + TILE - size, py, size, TILE)
}

function drawBridge(ctx: CanvasRenderingContext2D, keyAt: KeyAt, tx: number, ty: number, px: number, py: number) {
  const solidGround = (k: string | null) => k !== null && !isWater(k)
  const h = Number(solidGround(keyAt(tx - 1, ty))) + Number(solidGround(keyAt(tx + 1, ty)))
  const v = Number(solidGround(keyAt(tx, ty - 1))) + Number(solidGround(keyAt(tx, ty + 1)))
  const horizontal = h >= v

  ctx.fillStyle = GROUND_COLORS.water
  ctx.fillRect(px, py, TILE, TILE)

  const plank = '#a0703f'
  const gap = '#5e3f22'
  const rail = '#6b4626'
  if (horizontal) {
    ctx.fillStyle = gap
    ctx.fillRect(px, py + 4, TILE, TILE - 8)
    ctx.fillStyle = plank
    for (let x = 0; x < TILE; x += 6) ctx.fillRect(px + x, py + 5, 5, TILE - 10)
    ctx.fillStyle = rail
    ctx.fillRect(px, py + 3, TILE, 3)
    ctx.fillRect(px, py + TILE - 6, TILE, 3)
  } else {
    ctx.fillStyle = gap
    ctx.fillRect(px + 4, py, TILE - 8, TILE)
    ctx.fillStyle = plank
    for (let y = 0; y < TILE; y += 6) ctx.fillRect(px + 5, py + y, TILE - 10, 5)
    ctx.fillStyle = rail
    ctx.fillRect(px + 3, py, 3, TILE)
    ctx.fillRect(px + TILE - 6, py, 3, TILE)
  }
}

// Objects that sit flat on the ground are baked into the ground layer;
// everything else is a sprite that gets depth-sorted with the player.
export const FLAT_OBJECTS = new Set(['flowers'])

const FLOWER_COLORS = ['#f7a1d6', '#fff3a1', '#ffffff', '#c9a7ff', '#ff9e7a']

export function drawFlat(ctx: CanvasRenderingContext2D, key: string, tx: number, ty: number, px: number, py: number) {
  if (key !== 'flowers') return
  for (let i = 0; i < 5; i++) {
    const x = px + 4 + Math.floor(hash(tx, ty, 100 + i) * (TILE - 8))
    const y = py + 4 + Math.floor(hash(tx, ty, 110 + i) * (TILE - 8))
    ctx.fillStyle = '#3c7a34'
    ctx.fillRect(x, y + 2, 1, 3)
    ctx.fillStyle = FLOWER_COLORS[Math.floor(hash(tx, ty, 120 + i) * FLOWER_COLORS.length)]
    ctx.fillRect(x - 1, y - 1, 3, 1)
    ctx.fillRect(x - 1, y + 1, 3, 1)
    ctx.fillRect(x - 1, y, 1, 1)
    ctx.fillRect(x + 1, y, 1, 1)
    ctx.fillStyle = '#f2c94c'
    ctx.fillRect(x, y, 1, 1)
  }
}

// --- Sprites ---------------------------------------------------------------

/** Sprites are 2x2 tiles; the object's own tile is the bottom-centre cell. */
export const SPRITE_SIZE = TILE * 2
const SPRITE_RES = 2 // render at 2x so sprites stay sharp when zoomed in
export const VARIANTS = 3

export type SpriteSheet = Map<string, HTMLCanvasElement[]>

type Ctx = CanvasRenderingContext2D

function circle(ctx: Ctx, x: number, y: number, r: number, color: string) {
  ctx.fillStyle = color
  ctx.beginPath()
  ctx.arc(x, y, r, 0, Math.PI * 2)
  ctx.fill()
}

function ellipse(ctx: Ctx, x: number, y: number, rx: number, ry: number, color: string, rot = 0) {
  ctx.fillStyle = color
  ctx.beginPath()
  ctx.ellipse(x, y, rx, ry, rot, 0, Math.PI * 2)
  ctx.fill()
}

function triangle(ctx: Ctx, x: number, top: number, halfWidth: number, base: number, color: string) {
  ctx.fillStyle = color
  ctx.beginPath()
  ctx.moveTo(x, top)
  ctx.lineTo(x + halfWidth, base)
  ctx.lineTo(x - halfWidth, base)
  ctx.closePath()
  ctx.fill()
}

const SHADOW = 'rgba(0,0,0,0.25)'
const T = TILE
const B = TILE * 2 // bottom edge of the sprite == bottom of the object's tile

const SPRITE_PAINTERS: Record<string, (ctx: Ctx, v: number) => void> = {
  tree(ctx, v) {
    ellipse(ctx, T, B - 5, 13, 5, SHADOW)
    ctx.fillStyle = '#6b4423'
    ctx.fillRect(T - 3, B - 15, 6, 12)
    const cy = B - 23 - v * 2
    const [dark, mid, light] = [
      ['#2f6b2a', '#3f8a35', '#5aa847'],
      ['#2c6427', '#3a8031', '#62ad4b'],
      ['#356f25', '#4a8f30', '#6cb64a'],
    ][v]
    circle(ctx, T - 7, cy + 3, 9, dark)
    circle(ctx, T + 7, cy + 3, 9, dark)
    circle(ctx, T, cy - 4, 11, dark)
    circle(ctx, T - 5, cy + 1, 7, mid)
    circle(ctx, T + 5, cy, 7, mid)
    circle(ctx, T, cy - 5, 8, mid)
    circle(ctx, T - 3, cy - 8, 4, light)
    circle(ctx, T + 4, cy - 3, 3, light)
  },
  pine(ctx, v) {
    ellipse(ctx, T, B - 4, 11, 4, SHADOW)
    ctx.fillStyle = '#5a3a1e'
    ctx.fillRect(T - 2, B - 10, 4, 8)
    const top = B - 42 + v * 3
    triangle(ctx, T, top, 9, B - 24, '#1f4d33')
    triangle(ctx, T, top + 8, 12, B - 15, '#24583a')
    triangle(ctx, T, top + 16, 14, B - 7, '#2d6a45')
    triangle(ctx, T - 3, top + 18, 5, B - 9, '#3b8257')
    triangle(ctx, T - 2, top + 9, 4, B - 17, '#336f4a')
  },
  rock(ctx, v) {
    ellipse(ctx, T, B - 5, 12, 4, SHADOW)
    const pts = Array.from({ length: 7 }, (_, i) => {
      const a = (i / 7) * Math.PI * 2
      const r = 9 + hash(i, v, 7) * 4
      return [T + Math.cos(a) * r, B - 11 + Math.sin(a) * r * 0.75] as const
    })
    ctx.fillStyle = '#7f7f7f'
    ctx.strokeStyle = '#555'
    ctx.lineWidth = 1
    ctx.beginPath()
    pts.forEach(([x, y], i) => (i ? ctx.lineTo(x, y) : ctx.moveTo(x, y)))
    ctx.closePath()
    ctx.fill()
    ctx.stroke()
    ellipse(ctx, T - 2, B - 15, 6, 3.5, '#a3a3a3')
    circle(ctx, T + 5, B - 9, 1.5, '#6a6a6a')
  },
  bush(ctx, v) {
    ellipse(ctx, T, B - 4, 11, 4, SHADOW)
    circle(ctx, T - 6, B - 9, 7, '#2e6e2b')
    circle(ctx, T + 6, B - 9, 7, '#2e6e2b')
    circle(ctx, T, B - 13, 8, '#2e6e2b')
    circle(ctx, T - 4, B - 11, 5, '#3f8f3a')
    circle(ctx, T + 4, B - 12, 5, '#3f8f3a')
    circle(ctx, T - 2, B - 15, 2.5, '#5cad4f')
    if (v === 1) {
      for (const [x, y] of [[-5, -8], [3, -14], [6, -7], [-1, -11]]) circle(ctx, T + x, B + y, 1.5, '#d94a4a')
    }
  },
  wall(ctx) {
    const H = 12 // height of the front face
    ctx.fillStyle = '#9a9086'
    ctx.fillRect(T / 2, T - H, T, T)
    ctx.fillStyle = '#b1a79c'
    ctx.fillRect(T / 2, T - H, T, 2)
    ctx.fillStyle = '#857b71'
    ctx.fillRect(T / 2, T - H + T / 2, T, 1)
    ctx.fillRect(T, T - H, 1, T / 2)
    ctx.fillRect(T * 0.75, T - H + T / 2, 1, T / 2)
    ctx.fillStyle = '#6b625a'
    ctx.fillRect(T / 2, B - H, T, H)
    ctx.fillStyle = '#554d46'
    ctx.fillRect(T / 2, B - H / 2, T, 1)
    ctx.fillRect(T / 2 + 10, B - H, 1, H / 2)
    ctx.fillRect(T / 2 + 22, B - H, 1, H / 2)
    ctx.fillRect(T / 2 + 5, B - H / 2, 1, H / 2)
    ctx.fillRect(T / 2 + 16, B - H / 2, 1, H / 2)
    ctx.fillRect(T / 2 + 27, B - H / 2, 1, H / 2)
  },
}

export function createSprites(): SpriteSheet {
  const sheet: SpriteSheet = new Map()
  for (const [key, paint] of Object.entries(SPRITE_PAINTERS)) {
    const variants: HTMLCanvasElement[] = []
    for (let v = 0; v < VARIANTS; v++) {
      const c = document.createElement('canvas')
      c.width = c.height = SPRITE_SIZE * SPRITE_RES
      const ctx = c.getContext('2d')!
      ctx.scale(SPRITE_RES, SPRITE_RES)
      paint(ctx, v)
      variants.push(c)
    }
    sheet.set(key, variants)
  }
  return sheet
}

/** Generic fallback for object keys without a painter. */
export function drawFallbackObject(ctx: Ctx, color: string, px: number, py: number) {
  ctx.fillStyle = color
  ctx.fillRect(px + 6, py + 6, TILE - 12, TILE - 12)
}

// --- Player ----------------------------------------------------------------

export type Facing = { x: number; y: number }

/** Draws the player with feet at world pixel (px, py). `step` advances while walking. */
export function drawPlayer(ctx: Ctx, px: number, py: number, facing: Facing, step: number, moving: boolean) {
  const s = moving ? Math.sin(step * Math.PI * 2) * 2 : 0
  const bob = moving ? Math.abs(Math.sin(step * Math.PI * 2)) * 1.5 : 0

  ellipse(ctx, px, py, 9, 4, 'rgba(0,0,0,0.3)')

  ctx.fillStyle = '#2b2d42'
  ctx.fillRect(px - 5, py - 8 - Math.max(0, s), 4, 7)
  ctx.fillRect(px + 1, py - 8 - Math.max(0, -s), 4, 7)

  const top = py - 21 - bob
  ctx.fillStyle = '#3d5a80'
  ctx.beginPath()
  ctx.roundRect(px - 7, top, 14, 14, 4)
  ctx.fill()
  ctx.fillStyle = '#2b2d42'
  ctx.fillRect(px - 7, top + 9, 14, 2)
  ctx.fillStyle = '#e0a458'
  ctx.fillRect(px - 1, top + 9, 2, 2)

  const hy = top - 5
  circle(ctx, px, hy, 7, '#f1c27d')
  ctx.fillStyle = '#5b3a29'
  ctx.beginPath()
  if (facing.y < 0 && facing.x === 0) {
    ctx.arc(px, hy, 7, 0, Math.PI * 2) // back of the head
  } else {
    ctx.arc(px, hy, 7, Math.PI, 0)
  }
  ctx.fill()

  if (!(facing.y < 0 && facing.x === 0)) {
    ctx.fillStyle = '#222'
    const ex = px + facing.x * 2.5
    if (facing.x === 0) {
      ctx.fillRect(ex - 3.5, hy + 1, 2, 2.5)
      ctx.fillRect(ex + 1.5, hy + 1, 2, 2.5)
    } else {
      ctx.fillRect(ex + facing.x * 1.5 - 1, hy + 1, 2, 2.5)
    }
  }
}

/** Little flag marking the spawn point, drawn in world pixels. */
export function drawSpawnFlag(ctx: Ctx, tx: number, ty: number) {
  const x = tx * TILE + TILE / 2
  const y = ty * TILE + TILE - 4
  ellipse(ctx, x, y, 6, 2.5, SHADOW)
  ctx.fillStyle = '#ddd'
  ctx.fillRect(x - 1, y - 26, 2, 26)
  ctx.fillStyle = '#e63946'
  ctx.beginPath()
  ctx.moveTo(x + 1, y - 26)
  ctx.lineTo(x + 14, y - 21)
  ctx.lineTo(x + 1, y - 16)
  ctx.fill()
}

// --- Creatures -------------------------------------------------------------

const SKINS = ['#f1c27d', '#e0ac69', '#c68642', '#8d5524', '#ffdbac']
const HAIRS = ['#2b1d14', '#5b3a29', '#8a5a2b', '#c9a26b', '#1c1c1c', '#7a2e1d']
const HURT_SKIN = '#e8968a'

export type CreatureLook = {
  heading: number
  sex: Sex
  hue: number
  size: number
  flags: number
  energy: number
  /** 0–1; below 0.6 a health bar shows, below 0.25 the creature looks hurt. */
  health: number
  step: number
  moving: boolean
  /** Stable per-creature number used to pick skin and hair colours. */
  variant: number
  time: number
}

/** Draw scale of a creature: its body size, shrunk while it is still a child. */
export function creatureScale(size: number, flags: number) {
  return size * (flags & FLAG.child ? 0.65 : 1)
}

/** Draws a creature with feet at world pixel (px, py), plus effects and an icon for what it is doing. */
export function drawCreature(ctx: Ctx, px: number, py: number, c: CreatureLook) {
  const s = creatureScale(c.size, c.flags)
  const female = c.sex === 'female'
  const resting = (c.flags & FLAG.resting) !== 0
  const carrying = (c.flags & FLAG.carrying) !== 0
  const isHead = (c.flags & FLAG.head) !== 0
  const hurt = c.health < 0.25
  const fx = Math.cos(c.heading)
  const fy = Math.sin(c.heading)
  const facing = { x: Math.abs(fx) > 0.38 ? Math.sign(fx) : 0, y: Math.abs(fy) > 0.38 ? Math.sign(fy) : 0 }
  const back = facing.y < 0 && facing.x === 0

  // Hungry creatures look washed out.
  const sat = Math.round(18 + 42 * Math.min(1, c.energy * 3))
  const cloth = `hsl(${c.hue} ${sat}% 52%)`
  const clothDark = `hsl(${c.hue} ${sat}% 34%)`
  const skin = hurt ? HURT_SKIN : SKINS[c.variant % SKINS.length]
  const hair = HAIRS[Math.floor(c.variant / SKINS.length) % HAIRS.length]

  const swing = c.moving ? Math.sin(c.step * Math.PI * 2) * 2 : 0
  const bob = c.moving ? Math.abs(Math.sin(c.step * Math.PI * 2)) * 1.2 : 0
  const top = (resting ? -15 : -20) - bob // sitting lowers the body
  const hy = top - 5

  ctx.save()
  ctx.translate(px, py)
  ctx.scale(s, s)
  if (hurt) ellipse(ctx, 0, 0, 10, 4.5, `rgba(230,57,70,${0.3 + 0.2 * Math.sin(c.time * 6)})`)
  ellipse(ctx, 0, 0, 8, 3.5, 'rgba(0,0,0,0.28)')

  ctx.fillStyle = female ? skin : clothDark
  if (resting) {
    ctx.fillRect(-5, -4, 10, 3) // legs folded forward
  } else {
    ctx.fillRect(-4.5, -8 - Math.max(0, swing), 3.5, 7)
    ctx.fillRect(1, -8 - Math.max(0, -swing), 3.5, 7)
    ctx.fillStyle = '#2b2d42'
    ctx.fillRect(-4.5, -2 - Math.max(0, swing), 3.5, 2)
    ctx.fillRect(1, -2 - Math.max(0, -swing), 3.5, 2)
  }

  // A bundle on the back: behind the body, unless we see the creature from behind.
  const bundle = () => {
    const bx = back ? 0 : -facing.x * 5
    ellipse(ctx, bx, top + 2, 7, 6, '#8a6a3d')
    ctx.strokeStyle = '#5e4524'
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.moveTo(bx - 5, top - 1)
    ctx.lineTo(bx + 5, top + 6)
    ctx.stroke()
  }
  if (carrying && !back) bundle()

  // Long hair hangs behind the shoulders, or over the back when seen from behind.
  const longHair = () => {
    ctx.fillStyle = hair
    ctx.beginPath()
    ctx.roundRect(-7, hy - 2, 14, back ? 17 : 13, 5)
    ctx.fill()
  }
  if (female && !back) longHair()

  ctx.fillStyle = cloth
  ctx.beginPath()
  if (female) {
    // Dress flaring out into a skirt.
    ctx.moveTo(-5, top)
    ctx.lineTo(5, top)
    ctx.lineTo(8, top + 15)
    ctx.lineTo(-8, top + 15)
    ctx.closePath()
    ctx.fill()
    ctx.fillStyle = clothDark
    ctx.fillRect(-8, top + 13, 16, 2)
  } else {
    ctx.roundRect(-6.5, top, 13, 13, 3)
    ctx.fill()
    ctx.fillStyle = clothDark
    ctx.fillRect(-6.5, top + 9, 13, 2)
  }

  ctx.fillStyle = skin
  ctx.fillRect(-8, top + 2 + swing * 0.5, 2, 7)
  ctx.fillRect(6, top + 2 - swing * 0.5, 2, 7)
  if (female && back) longHair()
  if (carrying && back) bundle()

  if (c.flags & FLAG.pregnant && !back) {
    ellipse(ctx, facing.x * 2, top + 9, 4.5, 3.5, cloth)
    ellipse(ctx, facing.x * 2 - 1, top + 8, 1.8, 1.2, 'rgba(255,255,255,0.35)')
  }

  circle(ctx, 0, hy, 6.5, skin)
  ctx.fillStyle = hair
  ctx.beginPath()
  if (back) ctx.arc(0, hy, 6.8, 0, Math.PI * 2)
  else ctx.arc(0, hy, 6.8, Math.PI, 0)
  ctx.fill()
  if (female && !back) {
    ctx.fillRect(-7, hy - 1, 2.5, 8)
    ctx.fillRect(4.5, hy - 1, 2.5, 8)
  }
  if (c.flags & FLAG.stealing) {
    // A dark hood pulled over the head.
    ctx.fillStyle = 'rgba(28,28,38,0.9)'
    ctx.beginPath()
    ctx.arc(0, hy, 7.6, Math.PI * 0.9, Math.PI * 2.1)
    ctx.lineTo(7.6, hy + 4)
    ctx.lineTo(-7.6, hy + 4)
    ctx.closePath()
    ctx.fill()
  }

  if (!back) {
    ctx.fillStyle = '#222'
    const ex = facing.x * 2.5
    const eyeH = resting ? 0.8 : 2.2 // eyes closed while resting
    if (facing.x === 0) {
      ctx.fillRect(ex - 3.2, hy + 1, 1.8, eyeH)
      ctx.fillRect(ex + 1.4, hy + 1, 1.8, eyeH)
    } else {
      ctx.fillRect(ex + facing.x * 1.5 - 0.9, hy + 1, 1.8, eyeH)
    }
  }

  let above = hy - 8 // next free spot above the head
  if (isHead) {
    drawCrown(ctx, 0, hy - 7.5)
    above -= 4
  }
  ctx.restore()

  if (c.flags & FLAG.attacking) drawSlash(ctx, px, py - 12 * s, c.heading, s, c.time)
  if (c.flags & FLAG.crafting) drawSparks(ctx, px, py - 11 * s, s, c.time)

  if (c.health < 0.6) {
    const w = 16
    const y = py + above * s - 2
    ctx.fillStyle = 'rgba(0,0,0,0.55)'
    ctx.fillRect(px - w / 2 - 0.5, y - 0.5, w + 1, 3)
    ctx.fillStyle = c.health > 0.3 ? '#f4c542' : '#e63946'
    ctx.fillRect(px - w / 2, y, w * Math.max(0, c.health), 2)
    above -= 5 / s
  }

  drawStatusIcon(ctx, px, py + (above - 4) * s, c.flags, c.time)
}

/** Small gold crown for a kepala keluarga, centred on (x, y) in body units. */
function drawCrown(ctx: Ctx, x: number, y: number) {
  ctx.fillStyle = '#ffd166'
  ctx.beginPath()
  ctx.moveTo(x - 4.5, y + 2.5)
  ctx.lineTo(x - 4.5, y - 1.8)
  ctx.lineTo(x - 2.2, y + 0.4)
  ctx.lineTo(x, y - 3)
  ctx.lineTo(x + 2.2, y + 0.4)
  ctx.lineTo(x + 4.5, y - 1.8)
  ctx.lineTo(x + 4.5, y + 2.5)
  ctx.closePath()
  ctx.fill()
  ctx.strokeStyle = '#b8860b'
  ctx.lineWidth = 0.6
  ctx.stroke()
  ctx.fillStyle = '#e63946'
  ctx.fillRect(x - 0.7, y + 0.4, 1.4, 1.4)
}

/** Red sweeping arcs on the side the creature faces. */
function drawSlash(ctx: Ctx, x: number, y: number, heading: number, s: number, time: number) {
  const ph = (time * 3) % 1
  const cx = x + Math.cos(heading) * 8 * s
  const cy = y + Math.sin(heading) * 5 * s
  ctx.lineCap = 'round'
  for (let i = 0; i < 2; i++) {
    ctx.strokeStyle = `rgba(255,${70 + i * 60},${70 + i * 40},${0.95 - ph * 0.6})`
    ctx.lineWidth = 2.2 - i * 0.8
    ctx.beginPath()
    ctx.arc(cx, cy, (6 + i * 3) * s, heading - 1.2 + ph * 0.8, heading + 0.2 + ph * 0.8)
    ctx.stroke()
  }
  ctx.lineCap = 'butt'
}

/** Sparks dancing around the hands of someone crafting. */
function drawSparks(ctx: Ctx, x: number, y: number, s: number, time: number) {
  for (let i = 0; i < 5; i++) {
    const a = time * 5 + i * 1.26
    const r = (7 + Math.sin(time * 9 + i * 2) * 2.5) * s
    ctx.globalAlpha = 0.5 + 0.5 * Math.sin(time * 12 + i)
    circle(ctx, x + Math.cos(a) * r, y + Math.sin(a) * r * 0.6, 0.9 * s + 0.3, i % 2 ? '#ffd166' : '#ff9b3d')
  }
  ctx.globalAlpha = 1
}

/**
 * One small icon above the head, by priority: attacking, stealing, giving,
 * teaching, building, crafting, gathering, wants a mate, drinking, eating, resting.
 */
function drawStatusIcon(ctx: Ctx, x: number, y: number, flags: number, time: number) {
  y += Math.sin(time * 4 + x) * 0.8
  if (flags & FLAG.attacking) {
    // Crossed blades.
    ctx.strokeStyle = '#ff5c5c'
    ctx.lineWidth = 1.6
    ctx.lineCap = 'round'
    ctx.beginPath()
    ctx.moveTo(x - 3.5, y - 3.5)
    ctx.lineTo(x + 3.5, y + 3.5)
    ctx.moveTo(x + 3.5, y - 3.5)
    ctx.lineTo(x - 3.5, y + 3.5)
    ctx.stroke()
    ctx.lineCap = 'butt'
  } else if (flags & FLAG.stealing) {
    // A dark loot sack.
    circle(ctx, x, y + 1, 3.6, '#3d3426')
    ctx.fillStyle = '#3d3426'
    ctx.fillRect(x - 1.3, y - 3.5, 2.6, 2.5)
    ctx.fillStyle = '#ffd166'
    ctx.fillRect(x - 0.6, y, 1.2, 2)
  } else if (flags & FLAG.giving) {
    // A gift box with a ribbon.
    ctx.fillStyle = '#ff7aa8'
    ctx.fillRect(x - 3.5, y - 1.5, 7, 5.5)
    ctx.fillStyle = '#fff3a1'
    ctx.fillRect(x - 0.7, y - 1.5, 1.4, 5.5)
    ctx.fillRect(x - 3.5, y - 2.8, 7, 1.4)
    circle(ctx, x - 1.4, y - 3.6, 1.2, '#fff3a1')
    circle(ctx, x + 1.4, y - 3.6, 1.2, '#fff3a1')
  } else if (flags & FLAG.teaching) {
    // An open book, pages lifting gently.
    const lift = Math.sin(time * 5) * 0.6
    ctx.fillStyle = '#7a4f2c'
    ctx.fillRect(x - 5, y + 2.4, 10, 1.4)
    ctx.fillStyle = '#f4ecd8'
    ctx.beginPath()
    ctx.moveTo(x, y + 2.6)
    ctx.lineTo(x - 4.8, y + 2 - lift)
    ctx.lineTo(x - 4.8, y - 3 - lift)
    ctx.lineTo(x, y - 2.2)
    ctx.closePath()
    ctx.moveTo(x, y + 2.6)
    ctx.lineTo(x + 4.8, y + 2 - lift)
    ctx.lineTo(x + 4.8, y - 3 - lift)
    ctx.lineTo(x, y - 2.2)
    ctx.closePath()
    ctx.fill()
    ctx.strokeStyle = '#9aa3ad'
    ctx.lineWidth = 0.5
    ctx.beginPath()
    for (const dy of [-1.2, 0.4]) {
      ctx.moveTo(x - 3.8, y + dy - lift * 0.5)
      ctx.lineTo(x - 1, y + dy + 0.4)
      ctx.moveTo(x + 1, y + dy + 0.4)
      ctx.lineTo(x + 3.8, y + dy - lift * 0.5)
    }
    ctx.stroke()
  } else if (flags & FLAG.building) {
    // A hammer, swinging.
    ctx.save()
    ctx.translate(x, y + 3)
    ctx.rotate(Math.sin(time * 10) * 0.5 - 0.3)
    ctx.fillStyle = '#a0703f'
    ctx.fillRect(-0.7, -6, 1.4, 7)
    ctx.fillStyle = '#9aa3ad'
    ctx.fillRect(-3, -7.5, 6, 2.6)
    ctx.restore()
  } else if (flags & FLAG.crafting) {
    // An anvil.
    ctx.fillStyle = '#7d8590'
    ctx.beginPath()
    ctx.moveTo(x - 4.5, y - 2)
    ctx.lineTo(x + 3.5, y - 2)
    ctx.lineTo(x + 4.5, y - 0.5)
    ctx.lineTo(x + 1.5, y)
    ctx.lineTo(x + 1.5, y + 2)
    ctx.lineTo(x + 3, y + 3.5)
    ctx.lineTo(x - 3, y + 3.5)
    ctx.lineTo(x - 1.5, y + 2)
    ctx.lineTo(x - 1.5, y)
    ctx.closePath()
    ctx.fill()
  } else if (flags & FLAG.gathering) {
    // A pickaxe, digging.
    ctx.save()
    ctx.translate(x, y + 3)
    ctx.rotate(Math.sin(time * 8) * 0.4)
    ctx.fillStyle = '#a0703f'
    ctx.fillRect(-0.6, -6.5, 1.2, 7.5)
    ctx.strokeStyle = '#c3cbd3'
    ctx.lineWidth = 1.5
    ctx.beginPath()
    ctx.arc(0, -3.5, 4, Math.PI * 1.15, Math.PI * 1.85)
    ctx.stroke()
    ctx.restore()
  } else if (flags & FLAG.wantsMate) {
    circle(ctx, x - 1.8, y - 1, 2.1, '#ff5d8f')
    circle(ctx, x + 1.8, y - 1, 2.1, '#ff5d8f')
    ctx.beginPath()
    ctx.moveTo(x - 3.9, y - 0.4)
    ctx.lineTo(x + 3.9, y - 0.4)
    ctx.lineTo(x, y + 4)
    ctx.closePath()
    ctx.fill()
  } else if (flags & FLAG.drinking) {
    ctx.fillStyle = '#5ab0ff'
    ctx.beginPath()
    ctx.moveTo(x, y - 4.5)
    ctx.lineTo(x + 2.8, y + 0.8)
    ctx.arc(x, y + 0.8, 2.8, 0, Math.PI)
    ctx.closePath()
    ctx.fill()
  } else if (flags & FLAG.eating) {
    ctx.fillStyle = '#6cc24a'
    ctx.beginPath()
    ctx.ellipse(x, y, 4, 2.2, -0.6, 0, Math.PI * 2)
    ctx.fill()
    ctx.strokeStyle = '#3f7d2a'
    ctx.lineWidth = 0.7
    ctx.beginPath()
    ctx.moveTo(x - 3, y + 1.8)
    ctx.lineTo(x + 3, y - 1.8)
    ctx.stroke()
  } else if (flags & FLAG.resting) {
    ctx.fillStyle = '#dfe7ff'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.font = 'bold 8px sans-serif'
    ctx.fillText('z', x, y + 1)
    ctx.font = 'bold 6px sans-serif'
    ctx.fillText('z', x + 4, y - 3)
  }
}

// --- Structures ------------------------------------------------------------

/** Structure sprites are 2×3 tiles; the structure's own tile is the bottom-centre cell. */
export const STRUCTURE_W = TILE * 2
export const STRUCTURE_H = TILE * 3
/** How far (in tiles) a structure's art reaches above its tile, for hit-testing. */
export const STRUCTURE_REACH = 1.8

/** Structures that lie flat on the ground and are drawn under everything else. */
export const FLAT_STRUCTURES = new Set(['ladang', 'saluran_irigasi'])

/**
 * Farm works stay in use after their builder dies (only houses are inherited),
 * so they never look ruined just for having no living owner.
 */
export const COMMUNAL_STRUCTURES = new Set(['saluran_irigasi', 'lumbung', 'kandang'])

const STRUCTURE_NAMES: Record<string, string> = {
  gubuk: 'Gubuk',
  rumah_kayu: 'Rumah Kayu',
  rumah_bata: 'Rumah Bata',
  ladang: 'Ladang',
  saluran_irigasi: 'Saluran Irigasi',
  lumbung: 'Lumbung',
  kandang: 'Kandang',
  sumur: 'Sumur',
  jerat: 'Jerat',
  tungku: 'Tungku',
  laboratorium: 'Laboratorium',
  pembangkit_listrik: 'Pembangkit Listrik',
  lab_spektroskopi: 'Lab Spektroskopi',
  lab_radiasi: 'Lab Radiasi',
  reaktor_nuklir: 'Reaktor Nuklir',
  akselerator: 'Akselerator Partikel',
  perpustakaan: 'Perpustakaan',
}

export function structureName(kind: string) {
  return STRUCTURE_NAMES[kind] ?? kind.replace(/_/g, ' ').replace(/\b\w/g, (ch) => ch.toUpperCase())
}

/** Hover text: "Gubuk keluarga Adam", "Tungku · Hawa", "Gubuk kosong". */
export function structureLabel(kind: string, level: number, ownerName: string, ownerId: number) {
  const name = structureName(kind)
  if (!ownerId) return level > 0 ? `${name} kosong (ditinggalkan)` : `${name} (tak bertuan)`
  return level > 0 ? `${name} keluarga ${ownerName}` : `${name} · ${ownerName}`
}

const CX = TILE // horizontal centre of a structure sprite
const BY = TILE * 3 // bottom edge of a structure sprite == bottom of its tile

type Paint = (ctx: Ctx, hue: number, level: number) => void

/** Small pennant in the owner family's colour. */
function pennant(ctx: Ctx, x: number, y: number, hue: number) {
  ctx.fillStyle = '#5e4a36'
  ctx.fillRect(x, y, 1.2, 10)
  ctx.fillStyle = `hsl(${hue} 70% 55%)`
  ctx.beginPath()
  ctx.moveTo(x + 1.2, y)
  ctx.lineTo(x + 8, y + 2.2)
  ctx.lineTo(x + 1.2, y + 4.4)
  ctx.fill()
}

function hLines(ctx: Ctx, x0: number, x1: number, y0: number, y1: number, step: number, color: string) {
  ctx.fillStyle = color
  for (let y = y0; y < y1; y += step) ctx.fillRect(x0, y, x1 - x0, 1)
}

const paintHut: Paint = (ctx, hue) => {
  ellipse(ctx, CX, BY - 3, 18, 5, SHADOW)
  ctx.fillStyle = '#a7834f'
  ctx.fillRect(CX - 13, BY - 17, 26, 15)
  hLines(ctx, CX - 13, CX + 13, BY - 15, BY - 2, 3, '#8c6a3c')
  ctx.fillStyle = '#3b2a1a'
  ctx.beginPath()
  ctx.roundRect(CX - 4, BY - 12, 8, 10, [4, 4, 0, 0])
  ctx.fill()
  triangle(ctx, CX, BY - 40, 19, BY - 15, `hsl(${hue} 32% 46%)`)
  ctx.strokeStyle = `hsl(${hue} 32% 34%)`
  ctx.lineWidth = 0.8
  ctx.beginPath()
  for (let i = -3; i <= 3; i++) {
    ctx.moveTo(CX, BY - 39)
    ctx.lineTo(CX + i * 5.5, BY - 15.5)
  }
  ctx.stroke()
  ctx.fillStyle = `hsl(${hue} 32% 58%)`
  ctx.fillRect(CX - 1.5, BY - 42, 3, 4)
}

const paintWoodHouse: Paint = (ctx, hue) => {
  ellipse(ctx, CX, BY - 3, 22, 6, SHADOW)
  ctx.fillStyle = '#8a5a32'
  ctx.fillRect(CX - 17, BY - 24, 34, 22)
  hLines(ctx, CX - 17, CX + 17, BY - 21, BY - 2, 4, '#6e4625')
  ctx.fillStyle = '#4a2f1a'
  ctx.fillRect(CX - 4, BY - 14, 8, 12)
  circle(ctx, CX + 2, BY - 8, 0.8, '#d9b36c')
  for (const wx of [CX - 14, CX + 7]) {
    ctx.fillStyle = '#5e3c22'
    ctx.fillRect(wx - 1, BY - 20, 9, 8)
    ctx.fillStyle = '#ffe2a0'
    ctx.fillRect(wx, BY - 19, 7, 6)
    ctx.fillStyle = '#5e3c22'
    ctx.fillRect(wx + 3, BY - 19, 1, 6)
  }
  ctx.fillStyle = `hsl(${hue} 48% 40%)`
  ctx.beginPath()
  ctx.moveTo(CX - 21, BY - 23)
  ctx.lineTo(CX + 21, BY - 23)
  ctx.lineTo(CX + 13, BY - 42)
  ctx.lineTo(CX - 13, BY - 42)
  ctx.closePath()
  ctx.fill()
  hLines(ctx, CX - 18, CX + 18, BY - 39, BY - 24, 4, `hsl(${hue} 48% 30%)`)
  ctx.fillStyle = `hsl(${hue} 48% 52%)`
  ctx.fillRect(CX - 13, BY - 43, 26, 2)
}

const paintBrickHouse: Paint = (ctx, hue) => {
  ellipse(ctx, CX, BY - 3, 26, 6.5, SHADOW)
  ctx.fillStyle = '#a4553a'
  ctx.fillRect(CX - 21, BY - 30, 42, 28)
  ctx.fillStyle = '#7d3d28'
  for (let y = BY - 28, row = 0; y < BY - 2; y += 4, row++) {
    ctx.fillRect(CX - 21, y, 42, 1)
    for (let x = CX - 21 + (row % 2 ? 4 : 0); x < CX + 21; x += 8) ctx.fillRect(x, y + 1, 1, 3)
  }
  ctx.fillStyle = '#3d2a1e'
  ctx.beginPath()
  ctx.roundRect(CX - 5, BY - 16, 10, 14, [5, 5, 0, 0])
  ctx.fill()
  for (const wx of [CX - 17, CX + 9]) {
    ctx.fillStyle = '#e8e2d6'
    ctx.fillRect(wx - 1, BY - 25, 10, 10)
    ctx.fillStyle = '#ffd98a'
    ctx.fillRect(wx, BY - 24, 8, 8)
    ctx.fillStyle = '#e8e2d6'
    ctx.fillRect(wx + 3.5, BY - 24, 1, 8)
    ctx.fillRect(wx, BY - 20.5, 8, 1)
  }
  ctx.fillStyle = '#7d3d28'
  ctx.fillRect(CX + 10, BY - 52, 6, 16)
  circle(ctx, CX + 13, BY - 56, 3, 'rgba(210,210,210,0.5)')
  circle(ctx, CX + 16, BY - 61, 3.8, 'rgba(210,210,210,0.35)')
  ctx.fillStyle = `hsl(${hue} 55% 38%)`
  ctx.beginPath()
  ctx.moveTo(CX - 26, BY - 29)
  ctx.lineTo(CX + 26, BY - 29)
  ctx.lineTo(CX, BY - 54)
  ctx.closePath()
  ctx.fill()
  ctx.strokeStyle = `hsl(${hue} 55% 28%)`
  ctx.lineWidth = 1
  ctx.beginPath()
  for (let y = BY - 33; y > BY - 52; y -= 4) {
    const half = ((y - (BY - 54)) / 25) * 26
    ctx.moveTo(CX - half, y)
    ctx.lineTo(CX + half, y)
  }
  ctx.stroke()
}

const paintWell: Paint = (ctx, hue) => {
  ellipse(ctx, CX, BY - 5, 14, 5, SHADOW)
  ctx.fillStyle = '#6b4626'
  ctx.fillRect(CX - 12, BY - 30, 3, 22)
  ctx.fillRect(CX + 9, BY - 30, 3, 22)
  triangle(ctx, CX, BY - 39, 16, BY - 28, `hsl(${hue} 40% 40%)`)
  ctx.fillStyle = '#5a3a1e'
  ctx.fillRect(CX - 12, BY - 28, 24, 2)
  ctx.fillStyle = '#c9b48a'
  ctx.fillRect(CX - 0.5, BY - 26, 1, 9)
  ctx.fillStyle = '#8a6a3d'
  ctx.fillRect(CX - 2.5, BY - 18, 5, 4)
  ellipse(ctx, CX, BY - 9, 12, 5, '#8d8a83')
  ellipse(ctx, CX, BY - 10, 9, 3.3, '#1f4e8c')
  ctx.fillStyle = '#77736c'
  ctx.fillRect(CX - 12, BY - 9, 24, 6)
  ctx.fillStyle = '#5f5b55'
  for (let x = CX - 9; x < CX + 12; x += 6) ctx.fillRect(x, BY - 9, 1, 6)
  ctx.fillRect(CX - 12, BY - 6, 24, 1)
}

/** A snare: a sapling bent over a game trail with a cord noose. */
const paintSnare: Paint = (ctx, hue) => {
  ellipse(ctx, CX, BY - 3, 10, 3, SHADOW)
  ctx.strokeStyle = '#6b4626'
  ctx.lineWidth = 2
  ctx.beginPath()
  ctx.moveTo(CX - 8, BY - 2)
  ctx.quadraticCurveTo(CX - 8, BY - 22, CX + 6, BY - 18)
  ctx.stroke()
  ctx.strokeStyle = `hsl(${hue} 30% 70%)`
  ctx.lineWidth = 1
  ctx.beginPath()
  ctx.moveTo(CX + 6, BY - 18)
  ctx.lineTo(CX + 6, BY - 9)
  ctx.stroke()
  ctx.beginPath()
  ctx.ellipse(CX + 6, BY - 6, 4, 2.5, 0, 0, Math.PI * 2)
  ctx.stroke()
  ctx.fillStyle = '#5a3a1e'
  ctx.fillRect(CX + 9, BY - 6, 2, 5)
}

const paintFurnace: Paint = (ctx, hue) => {
  ellipse(ctx, CX, BY - 3, 18, 5, SHADOW)
  ctx.fillStyle = '#7a5238'
  ctx.fillRect(CX + 5, BY - 38, 6, 16)
  circle(ctx, CX + 9, BY - 42, 3, 'rgba(200,200,200,0.55)')
  circle(ctx, CX + 12, BY - 48, 4, 'rgba(200,200,200,0.35)')
  ctx.fillStyle = '#9c6b4a'
  ctx.beginPath()
  ctx.ellipse(CX, BY - 6, 15, 20, 0, Math.PI, 0)
  ctx.fill()
  ctx.fillStyle = '#7a5238'
  ctx.fillRect(CX - 15, BY - 8, 30, 6)
  ctx.fillStyle = '#86593d'
  for (let i = 0; i < 4; i++) ctx.fillRect(CX - 10 + i * 6, BY - 20 + (i % 2) * 5, 4, 1.5)
  ctx.fillStyle = '#2a140a'
  ctx.beginPath()
  ctx.roundRect(CX - 6, BY - 17, 12, 12, [6, 6, 0, 0])
  ctx.fill()
  ctx.fillStyle = '#ff9b3d'
  ctx.beginPath()
  ctx.roundRect(CX - 4, BY - 12, 8, 7, [4, 4, 0, 0])
  ctx.fill()
  ellipse(ctx, CX, BY - 7, 3, 2, '#ffd166')
  pennant(ctx, CX - 14, BY - 30, hue)
}

const paintLab: Paint = (ctx, hue) => {
  ellipse(ctx, CX, BY - 3, 24, 6, SHADOW)
  ctx.fillStyle = '#d8d2c4'
  ctx.fillRect(CX - 20, BY - 26, 40, 24)
  ctx.fillStyle = '#b9b2a2'
  ctx.fillRect(CX - 20, BY - 6, 40, 4)
  ctx.fillStyle = '#56657a'
  ctx.fillRect(CX - 23, BY - 30, 46, 5)
  for (const wx of [CX - 16, CX + 8]) {
    ctx.fillStyle = '#8fa3b8'
    ctx.fillRect(wx - 1, BY - 21, 10, 9)
    ctx.fillStyle = '#bfe3f5'
    ctx.fillRect(wx, BY - 20, 8, 7)
  }
  ctx.fillStyle = '#56657a'
  ctx.fillRect(CX - 4, BY - 15, 8, 13)
  // Flask sign on the roof.
  ctx.fillStyle = '#e8f4f0'
  ctx.fillRect(CX - 1.5, BY - 44, 3, 6)
  circle(ctx, CX, BY - 35, 5.5, '#e8f4f0')
  ctx.fillStyle = '#5fd38d'
  ctx.beginPath()
  ctx.arc(CX, BY - 35, 4.5, 0.1, Math.PI - 0.1)
  ctx.fill()
  pennant(ctx, CX + 17, BY - 41, hue)
}

const paintPowerPlant: Paint = (ctx, hue) => {
  ellipse(ctx, CX, BY - 3, 25, 6, SHADOW)
  ctx.fillStyle = '#6b7782'
  ctx.fillRect(CX + 8, BY - 52, 9, 50)
  ctx.fillStyle = '#e63946'
  ctx.fillRect(CX + 8, BY - 52, 9, 3)
  ctx.fillRect(CX + 8, BY - 44, 9, 3)
  circle(ctx, CX + 13, BY - 57, 3.5, 'rgba(200,200,200,0.45)')
  ctx.fillStyle = '#7f8c97'
  ctx.fillRect(CX - 21, BY - 24, 29, 22)
  ctx.fillStyle = '#5d6872'
  ctx.fillRect(CX - 21, BY - 27, 29, 3)
  // Lightning bolt.
  ctx.fillStyle = '#ffd166'
  ctx.beginPath()
  ctx.moveTo(CX - 6, BY - 22)
  ctx.lineTo(CX - 12, BY - 11)
  ctx.lineTo(CX - 7, BY - 11)
  ctx.lineTo(CX - 10, BY - 3)
  ctx.lineTo(CX - 1, BY - 15)
  ctx.lineTo(CX - 6, BY - 15)
  ctx.lineTo(CX - 3, BY - 22)
  ctx.closePath()
  ctx.fill()
  // Copper coils.
  ctx.strokeStyle = '#c87533'
  ctx.lineWidth = 1.5
  for (let i = 0; i < 3; i++) {
    ctx.beginPath()
    ctx.ellipse(CX + 3, BY - 20 + i * 5, 3, 1.6, 0, 0, Math.PI * 2)
    ctx.stroke()
  }
  pennant(ctx, CX - 20, BY - 38, hue)
}

const paintSpectroLab: Paint = (ctx, hue) => {
  ellipse(ctx, CX, BY - 3, 23, 6, SHADOW)
  ctx.fillStyle = '#e9ecef'
  ctx.beginPath()
  ctx.arc(CX, BY - 21, 15, Math.PI, 0)
  ctx.fill()
  ctx.fillStyle = '#c3c9cf'
  ctx.beginPath()
  ctx.arc(CX, BY - 21, 15, Math.PI * 1.5, 0)
  ctx.lineTo(CX, BY - 21)
  ctx.fill()
  ctx.fillStyle = '#2b3440'
  ctx.fillRect(CX - 2, BY - 35, 4, 13)
  ctx.fillStyle = '#cfc8bb'
  ctx.fillRect(CX - 19, BY - 22, 38, 20)
  const rainbow = ['#ff595e', '#ff924c', '#ffca3a', '#8ac926', '#1982c4', '#6a4c93']
  rainbow.forEach((c, i) => {
    ctx.fillStyle = c
    ctx.fillRect(CX - 19, BY - 18 + i * 2, 38, 2)
  })
  ctx.fillStyle = '#4a4f57'
  ctx.fillRect(CX - 4, BY - 6, 8, 4)
  pennant(ctx, CX + 13, BY - 44, hue)
}

const paintRadiationLab: Paint = (ctx, hue) => {
  ellipse(ctx, CX, BY - 3, 25, 6, SHADOW)
  ctx.fillStyle = '#8e9196'
  ctx.fillRect(CX - 21, BY - 28, 42, 26)
  ctx.fillStyle = '#5d6168'
  ctx.fillRect(CX - 21, BY - 28, 42, 4)
  ctx.fillRect(CX - 21, BY - 10, 42, 2)
  // Hazard-striped door.
  ctx.fillStyle = '#ffd166'
  ctx.fillRect(CX - 6, BY - 12, 12, 10)
  ctx.fillStyle = '#1c1c1c'
  for (let i = 0; i < 4; i++) ctx.fillRect(CX - 6 + i * 3, BY - 12, 1.5, 10)
  // Trefoil sign.
  circle(ctx, CX, BY - 19, 5.5, '#ffd166')
  ctx.fillStyle = '#1c1c1c'
  for (let i = 0; i < 3; i++) {
    const a = -Math.PI / 2 + (i * Math.PI * 2) / 3
    ctx.beginPath()
    ctx.moveTo(CX, BY - 19)
    ctx.arc(CX, BY - 19, 4.6, a - 0.5, a + 0.5)
    ctx.closePath()
    ctx.fill()
  }
  circle(ctx, CX, BY - 19, 1.4, '#ffd166')
  circle(ctx, CX, BY - 19, 0.9, '#1c1c1c')
  pennant(ctx, CX + 15, BY - 40, hue)
}

const paintReactor: Paint = (ctx, hue) => {
  ellipse(ctx, CX, BY - 3, 28, 6.5, SHADOW)
  // Containment dome.
  ctx.fillStyle = '#e2e4e8'
  ctx.beginPath()
  ctx.arc(CX - 15, BY - 3, 11, Math.PI, 0)
  ctx.fill()
  ctx.fillStyle = '#b8bcc4'
  ctx.fillRect(CX - 26, BY - 6, 22, 2)
  // Cooling tower (hyperboloid).
  const tx = CX + 6
  ctx.fillStyle = '#c9ccd1'
  ctx.beginPath()
  ctx.moveTo(tx - 15, BY - 2)
  ctx.quadraticCurveTo(tx - 7, BY - 30, tx - 11, BY - 52)
  ctx.lineTo(tx + 11, BY - 52)
  ctx.quadraticCurveTo(tx + 7, BY - 30, tx + 15, BY - 2)
  ctx.closePath()
  ctx.fill()
  ctx.fillStyle = '#a9adb4'
  ctx.beginPath()
  ctx.moveTo(tx + 4, BY - 2)
  ctx.quadraticCurveTo(tx + 2, BY - 30, tx + 4, BY - 52)
  ctx.lineTo(tx + 11, BY - 52)
  ctx.quadraticCurveTo(tx + 7, BY - 30, tx + 15, BY - 2)
  ctx.closePath()
  ctx.fill()
  ellipse(ctx, tx, BY - 52, 11, 2.5, '#6b7078')
  circle(ctx, tx - 3, BY - 60, 6, 'rgba(240,240,240,0.65)')
  circle(ctx, tx + 4, BY - 66, 7, 'rgba(240,240,240,0.5)')
  circle(ctx, tx - 1, BY - 74, 6, 'rgba(240,240,240,0.35)')
  pennant(ctx, CX - 16, BY - 26, hue)
}

const paintAccelerator: Paint = (ctx, hue) => {
  ellipse(ctx, CX, BY - 8, 30, 10, SHADOW)
  ctx.fillStyle = '#9aa5b1'
  ctx.fillRect(CX - 10, BY - 30, 20, 16)
  ctx.fillStyle = '#7d8896'
  ctx.fillRect(CX - 10, BY - 32, 20, 3)
  ctx.fillStyle = '#7ad7ff'
  ctx.fillRect(CX - 6, BY - 25, 12, 4)
  ctx.strokeStyle = '#7d8fa6'
  ctx.lineWidth = 5
  ctx.beginPath()
  ctx.ellipse(CX, BY - 10, 27, 8, 0, 0, Math.PI * 2)
  ctx.stroke()
  ctx.strokeStyle = '#7ad7ff'
  ctx.lineWidth = 1.2
  ctx.beginPath()
  ctx.ellipse(CX, BY - 10, 27, 8, 0, 0, Math.PI * 2)
  ctx.stroke()
  ctx.fillStyle = '#3d4654'
  for (let i = 0; i < 8; i++) {
    const a = (i / 8) * Math.PI * 2
    ctx.fillRect(CX + Math.cos(a) * 27 - 2, BY - 10 + Math.sin(a) * 8 - 2.5, 4, 5)
  }
  pennant(ctx, CX + 8, BY - 44, hue)
}

/** A library of clay tablets: a colonnaded hall with shelves of tablets inside. */
const paintLibrary: Paint = (ctx, hue) => {
  ellipse(ctx, CX, BY - 3, 25, 6, SHADOW)
  // Steps and floor.
  ctx.fillStyle = '#b8ab92'
  ctx.fillRect(CX - 23, BY - 6, 46, 4)
  ctx.fillStyle = '#cdbfa5'
  ctx.fillRect(CX - 21, BY - 9, 42, 3)
  // Dark interior with shelves of tablets behind the columns.
  ctx.fillStyle = '#4a3b2c'
  ctx.fillRect(CX - 19, BY - 30, 38, 21)
  for (let row = 0; row < 3; row++) {
    const y = BY - 27 + row * 6
    ctx.fillStyle = '#6b5640'
    ctx.fillRect(CX - 18, y + 4, 36, 1)
    for (let k = 0; k < 9; k++) {
      ctx.fillStyle = (k + row) % 3 ? '#c99a62' : '#b0643c'
      ctx.fillRect(CX - 17 + k * 4, y, 3, 4)
    }
  }
  // Columns.
  ctx.fillStyle = '#e4dccb'
  for (const cx of [CX - 19, CX - 7, CX + 5, CX + 17]) {
    ctx.fillRect(cx - 1.5, BY - 31, 3, 22)
    ctx.fillRect(cx - 2.5, BY - 32, 5, 2)
  }
  // Pediment with an open-book emblem.
  ctx.fillStyle = '#d6ccb6'
  ctx.fillRect(CX - 23, BY - 35, 46, 4)
  triangle(ctx, CX, BY - 49, 24, BY - 35, `hsl(${hue} 30% 62%)`)
  ctx.fillStyle = '#f4ecd8'
  ctx.beginPath()
  ctx.moveTo(CX, BY - 39)
  ctx.lineTo(CX - 5, BY - 40.5)
  ctx.lineTo(CX - 5, BY - 44.5)
  ctx.lineTo(CX, BY - 43)
  ctx.lineTo(CX + 5, BY - 44.5)
  ctx.lineTo(CX + 5, BY - 40.5)
  ctx.closePath()
  ctx.fill()
  pennant(ctx, CX + 20, BY - 56, hue)
}

/** Lumbung: a granary on stilts with rat guards, under a tall curved thatch roof. */
const paintGranary: Paint = (ctx, hue) => {
  ellipse(ctx, CX, BY - 3, 18, 5, SHADOW)
  // Four stilts, each with a round wooden disc the rats can't climb past.
  ctx.fillStyle = '#5e3c22'
  for (const x of [CX - 11, CX - 4, CX + 3, CX + 10]) ctx.fillRect(x - 1, BY - 20, 2.4, 18)
  for (const x of [CX - 11, CX - 4, CX + 3, CX + 10]) ellipse(ctx, x + 0.2, BY - 13, 3.6, 1.2, '#8a6a48')
  // Raised store room of woven bamboo.
  ctx.fillStyle = '#a7834f'
  ctx.fillRect(CX - 13, BY - 34, 26, 14)
  ctx.fillStyle = '#8c6a3c'
  for (let x = CX - 12; x < CX + 13; x += 3) ctx.fillRect(x, BY - 34, 1, 14)
  ctx.fillStyle = '#6b4626'
  ctx.fillRect(CX - 14, BY - 21, 28, 2)
  ctx.fillStyle = '#3b2a1a'
  ctx.fillRect(CX - 3, BY - 31, 6, 8)
  // Tall thatch that bows outward at the eaves.
  ctx.fillStyle = '#b89a5a'
  ctx.beginPath()
  ctx.moveTo(CX - 19, BY - 30)
  ctx.quadraticCurveTo(CX - 13, BY - 42, CX - 3, BY - 60)
  ctx.lineTo(CX + 3, BY - 60)
  ctx.quadraticCurveTo(CX + 13, BY - 42, CX + 19, BY - 30)
  ctx.closePath()
  ctx.fill()
  ctx.strokeStyle = '#8f7442'
  ctx.lineWidth = 0.8
  ctx.beginPath()
  for (let i = -3; i <= 3; i++) {
    ctx.moveTo(CX + i * 1, BY - 58)
    ctx.lineTo(CX + i * 5.5, BY - 31)
  }
  ctx.stroke()
  ctx.fillStyle = '#9a7d45'
  ctx.fillRect(CX - 19, BY - 31, 38, 2)
  pennant(ctx, CX + 1, BY - 70, hue)
}

/** Kandang: a low pen of posts and rails with a straw floor and a lean-to shelter. */
const paintPen: Paint = (ctx, hue) => {
  ellipse(ctx, CX, BY - 6, 24, 7, SHADOW)
  // Straw-strewn floor.
  ctx.fillStyle = '#9c8452'
  ctx.beginPath()
  ctx.moveTo(CX - 22, BY - 6)
  ctx.lineTo(CX - 15, BY - 20)
  ctx.lineTo(CX + 15, BY - 20)
  ctx.lineTo(CX + 22, BY - 6)
  ctx.closePath()
  ctx.fill()
  ctx.fillStyle = '#c4a96a'
  for (let i = 0; i < 9; i++) ctx.fillRect(CX - 14 + i * 3.4, BY - 15 + (i % 3) * 3, 3, 0.8)
  // Lean-to shelter at the back.
  ctx.fillStyle = '#5e3c22'
  ctx.fillRect(CX - 13, BY - 30, 2, 12)
  ctx.fillRect(CX + 1, BY - 30, 2, 12)
  ctx.fillStyle = `hsl(${hue} 30% 42%)`
  ctx.beginPath()
  ctx.moveTo(CX - 16, BY - 26)
  ctx.lineTo(CX + 6, BY - 26)
  ctx.lineTo(CX + 4, BY - 33)
  ctx.lineTo(CX - 14, BY - 33)
  ctx.closePath()
  ctx.fill()
  // Fence: back rail, then posts and front rails.
  const post = '#6b4626'
  const rail = '#8a5a32'
  ctx.fillStyle = rail
  ctx.fillRect(CX - 15, BY - 24, 30, 1.6)
  ctx.fillStyle = post
  for (const x of [CX - 15, CX - 5, CX + 5, CX + 14]) ctx.fillRect(x, BY - 26, 1.8, 7)
  ctx.strokeStyle = rail
  ctx.lineWidth = 1.6
  ctx.beginPath()
  ctx.moveTo(CX - 22, BY - 12)
  ctx.lineTo(CX - 15, BY - 23)
  ctx.moveTo(CX + 22, BY - 12)
  ctx.lineTo(CX + 15, BY - 23)
  ctx.stroke()
  for (const x of [CX - 22, CX - 11, CX, CX + 11, CX + 20]) {
    ctx.fillStyle = post
    ctx.fillRect(x, BY - 14, 2.2, 10)
    ctx.fillStyle = '#8a6a48'
    ctx.fillRect(x, BY - 14, 2.2, 1)
  }
  ctx.fillStyle = rail
  ctx.fillRect(CX - 22, BY - 12, 44, 1.8)
  ctx.fillRect(CX - 22, BY - 8, 44, 1.8)
  // A water trough.
  ctx.fillStyle = '#6b4626'
  ctx.fillRect(CX + 6, BY - 17, 9, 3)
  ctx.fillStyle = '#5a9bd4'
  ctx.fillRect(CX + 7, BY - 17, 7, 1)
}

const paintGeneric: Paint = (ctx, hue) => {
  ellipse(ctx, CX, BY - 3, 20, 5.5, SHADOW)
  ctx.fillStyle = '#a89f91'
  ctx.fillRect(CX - 16, BY - 22, 32, 20)
  ctx.fillStyle = '#5e5850'
  ctx.fillRect(CX - 4, BY - 12, 8, 10)
  triangle(ctx, CX, BY - 36, 20, BY - 21, `hsl(${hue} 40% 42%)`)
}

const PAINTERS: Record<string, Paint> = {
  gubuk: paintHut,
  rumah_kayu: paintWoodHouse,
  rumah_bata: paintBrickHouse,
  sumur: paintWell,
  jerat: paintSnare,
  tungku: paintFurnace,
  laboratorium: paintLab,
  pembangkit_listrik: paintPowerPlant,
  lab_spektroskopi: paintSpectroLab,
  lab_radiasi: paintRadiationLab,
  reaktor_nuklir: paintReactor,
  akselerator: paintAccelerator,
  perpustakaan: paintLibrary,
  lumbung: paintGranary,
  kandang: paintPen,
}

/** Houses of unknown kinds are drawn by level. */
const HOUSE_BY_LEVEL = [paintGeneric, paintHut, paintWoodHouse, paintBrickHouse]

/** Cracks and holes over an abandoned (greyed) building. */
function paintRuin(ctx: Ctx) {
  ctx.strokeStyle = 'rgba(20,20,20,0.6)'
  ctx.lineWidth = 1
  ctx.beginPath()
  ctx.moveTo(CX - 10, BY - 22)
  ctx.lineTo(CX - 6, BY - 15)
  ctx.lineTo(CX - 9, BY - 8)
  ctx.moveTo(CX + 8, BY - 26)
  ctx.lineTo(CX + 11, BY - 18)
  ctx.stroke()
  ctx.fillStyle = 'rgba(15,15,15,0.55)'
  ctx.beginPath()
  ctx.moveTo(CX - 3, BY - 36)
  ctx.lineTo(CX + 6, BY - 33)
  ctx.lineTo(CX + 2, BY - 27)
  ctx.lineTo(CX - 5, BY - 29)
  ctx.closePath()
  ctx.fill()
  ctx.save()
  ctx.translate(CX + 14, BY - 3)
  ctx.rotate(-0.35)
  ctx.fillStyle = '#5a554e'
  ctx.fillRect(-9, -1.5, 18, 3)
  ctx.restore()
}

const structureCache = new Map<string, HTMLCanvasElement>()

function makeCanvas() {
  const c = document.createElement('canvas')
  c.width = STRUCTURE_W * SPRITE_RES
  c.height = STRUCTURE_H * SPRITE_RES
  return c
}

/** Cached 2×3-tile sprite for a structure; hue is quantised so the cache stays small. */
export function structureSprite(kind: string, level: number, hue: number, abandoned: boolean): HTMLCanvasElement {
  const qh = abandoned ? 0 : (Math.round(hue / 15) * 15) % 360
  const key = `${kind}|${level}|${abandoned ? 'ruin' : qh}`
  const hit = structureCache.get(key)
  if (hit) return hit

  const paint = PAINTERS[kind] ?? HOUSE_BY_LEVEL[level] ?? paintGeneric
  let sprite = makeCanvas()
  const ctx = sprite.getContext('2d')!
  ctx.scale(SPRITE_RES, SPRITE_RES)
  paint(ctx, qh, level)
  if (abandoned) {
    // Grey it out, then add damage.
    const ruin = makeCanvas()
    const rctx = ruin.getContext('2d')!
    rctx.filter = 'grayscale(0.9) brightness(0.72)'
    rctx.drawImage(sprite, 0, 0)
    rctx.filter = 'none'
    rctx.scale(SPRITE_RES, SPRITE_RES)
    paintRuin(rctx)
    sprite = ruin
  }
  structureCache.set(key, sprite)
  return sprite
}

/**
 * Farm field drawn flat on its tile at world pixel (px, py). A `planted` field
 * shows only its tilled rows: the crop growing on it is drawn on top.
 */
export function drawFarm(
  ctx: Ctx,
  px: number,
  py: number,
  hue: number,
  abandoned: boolean,
  tx: number,
  ty: number,
  planted = false,
) {
  ctx.fillStyle = abandoned ? '#6d5a45' : '#7a5332'
  ctx.fillRect(px + 1, py + 1, TILE - 2, TILE - 2)
  ctx.fillStyle = abandoned ? '#5a4a39' : '#5e3e24'
  for (let y = py + 5; y < py + TILE - 2; y += 6) ctx.fillRect(px + 2, y, TILE - 4, 2)
  if (planted) {
    pennant(ctx, px + TILE - 5, py - 6, hue)
    return
  }
  ctx.fillStyle = abandoned ? '#a8935a' : '#6cc24a'
  for (let y = py + 3, row = 0; y < py + TILE - 4; y += 6, row++) {
    for (let x = px + 4 + (row % 2) * 3; x < px + TILE - 3; x += 6) {
      if (abandoned && hash(tx * 7 + x, ty * 5 + y, 3) < 0.5) continue
      ctx.fillRect(x, y, 1, 3)
      ctx.fillRect(x - 1, y + 1, 1, 1)
      ctx.fillRect(x + 1, y, 1, 1)
    }
  }
  if (!abandoned) pennant(ctx, px + TILE - 5, py - 6, hue)
}

/**
 * Irrigation channel flat on its tile at world pixel (px, py): an earth-banked
 * ditch of water, running towards the neighbours in `links` (bits N=1, E=2,
 * S=4, W=8: other channels, rivers and fields it feeds).
 */
export function drawIrrigation(ctx: Ctx, px: number, py: number, links: number) {
  const c = TILE / 2
  const bank = '#7a5a38'
  const water = '#3f86c4'
  // A lone channel still shows as a short east–west ditch.
  const arms = links || (2 | 8)
  const arm = (bit: number, w: number, color: string) => {
    if (!(arms & bit)) return
    ctx.fillStyle = color
    if (bit === 1) ctx.fillRect(px + c - w / 2, py, w, c)
    else if (bit === 4) ctx.fillRect(px + c - w / 2, py + c, w, c)
    else if (bit === 2) ctx.fillRect(px + c, py + c - w / 2, c, w)
    else ctx.fillRect(px, py + c - w / 2, c, w)
  }
  for (const bit of [1, 2, 4, 8]) arm(bit, 12, bank)
  ellipse(ctx, px + c, py + c, 7, 7, bank)
  for (const bit of [1, 2, 4, 8]) arm(bit, 6, water)
  ellipse(ctx, px + c, py + c, 4, 4, water)
  ctx.fillStyle = 'rgba(255,255,255,0.35)'
  ctx.fillRect(px + c - 2, py + c - 1, 3, 1)
  // Stones lining the banks.
  ctx.fillStyle = '#8d8a83'
  for (let i = 0; i < 4; i++) {
    const a = (i / 4) * Math.PI * 2 + 0.6
    ctx.fillRect(px + c + Math.cos(a) * 7 - 1, py + c + Math.sin(a) * 7 - 1, 2, 1.6)
  }
}

/** A tree cut down to a stump, baked into the ground chunk (its sprite is no longer drawn). */
export function drawStump(ctx: Ctx, tx: number, ty: number, px: number, py: number) {
  const cx = px + TILE / 2 + (hash(tx, ty, 620) - 0.5) * 6
  const cy = py + TILE - 9
  ellipse(ctx, cx + 1, cy + 4, 9, 3, 'rgba(0,0,0,0.25)')
  // Roots.
  ctx.fillStyle = '#5a3a1e'
  ctx.beginPath()
  ctx.moveTo(cx - 9, cy + 4)
  ctx.lineTo(cx - 4, cy - 1)
  ctx.lineTo(cx + 4, cy - 1)
  ctx.lineTo(cx + 9, cy + 4)
  ctx.closePath()
  ctx.fill()
  ctx.fillStyle = '#6b4423'
  ctx.fillRect(cx - 5, cy - 6, 10, 8)
  ellipse(ctx, cx, cy + 2, 5, 1.8, '#6b4423')
  // Cut face with growth rings.
  ellipse(ctx, cx, cy - 6, 5, 2.2, '#c99a62')
  ctx.strokeStyle = '#a0703f'
  ctx.lineWidth = 0.6
  ctx.beginPath()
  ctx.ellipse(cx, cy - 6, 3, 1.3, 0, 0, Math.PI * 2)
  ctx.stroke()
  // A sapling or two sprouting beside it.
  if (hash(tx, ty, 621) < 0.6) {
    const sx = cx + 7
    ctx.fillStyle = '#4f9a3c'
    ctx.fillRect(sx, cy - 4, 1, 6)
    ellipse(ctx, sx - 1.5, cy - 4, 2, 1, '#5aa847', -0.5)
    ellipse(ctx, sx + 2, cy - 5, 2, 1, '#5aa847', 0.5)
  }
}

// --- Resource deposits -----------------------------------------------------

// Mineral colours, roughly as the real ores look. Unlisted items fall back to
// their first element's tint, then to a hashed hue.
const DEPOSIT_COLORS: Record<string, string> = {
  hematit: '#b4532f',
  kalkopirit: '#c9a13a',
  kasiterit: '#5b4636',
  galena: '#8a96a3',
  sfalerit: '#7a5a32',
  sinabar: '#c0392b',
  bijih_emas: '#f2c94c',
  belerang: '#e9d43f',
  batu_bara: '#262626',
  batu_kapur: '#ece6d6',
  dolomit: '#ddd3c0',
  gipsum: '#f3f0e8',
  fluorit: '#9b7ae0',
  apatit: '#7fb38f',
  bauksit: '#c8703d',
  uraninit: '#9be15d',
  monasit: '#d2924f',
  xenotim: '#b48fd9',
  beril: '#5fd3b8',
  zirkon: '#d9a066',
  pirolusit: '#3f3a44',
  kromit: '#4d5a45',
  malakit: '#2fa58a',
}

const ELEMENT_TINTS: Record<string, string> = {
  Fe: '#b4532f', Cu: '#2fa58a', Sn: '#cfd6db', Pb: '#6b7682', Zn: '#94a7bb', Au: '#f2c94c', Ag: '#e8ebef',
  Hg: '#c0392b', S: '#e9d43f', C: '#2a2a2a', U: '#9be15d', Ti: '#7d808a', Cr: '#5a6b4f', Mn: '#4a3f52',
  Ni: '#a2b06e', Mo: '#9aa6b5', W: '#50545e', Zr: '#d9b48f', Be: '#7fd6c2', Li: '#d9c4ea', Cs: '#efe7cf',
  Rb: '#e6d3e8', Ce: '#c79be0', Y: '#b48fd9', Sc: '#a8cbe3', Ca: '#efe9dc', Ba: '#f1e6c9', Sr: '#d7e7f3',
  Na: '#f6f6f6', B: '#ebe5d3', Nb: '#60718a', Co: '#3a5fd0', Sb: '#8d9095', As: '#bcae8c', Bi: '#d9a6cb',
  Al: '#c8703d', V: '#cf5631', K: '#f0b3b3', P: '#82b892', Mg: '#e0d8c7', Ta: '#6b7b8c', Re: '#9aa0a8',
}

export function depositColor(id: string, elements: string[]): string {
  const direct = DEPOSIT_COLORS[id]
  if (direct) return direct
  for (const e of elements) {
    if (ELEMENT_TINTS[e]) return ELEMENT_TINTS[e]
  }
  let h = 0
  for (const ch of id) h = (h * 31 + ch.charCodeAt(0)) | 0
  return `hsl(${Math.abs(h) % 360} 55% 60%)`
}

function crystals(ctx: Ctx, tx: number, ty: number, px: number, py: number, color: string, count: number, salt: number) {
  for (let i = 0; i < count; i++) {
    const x = px + 5 + hash(tx, ty, salt + i * 2) * (TILE - 10)
    const y = py + 6 + hash(tx, ty, salt + i * 2 + 1) * (TILE - 12)
    const r = 2 + hash(tx, ty, salt + 40 + i) * 1.5
    ctx.fillStyle = 'rgba(0,0,0,0.35)'
    ctx.beginPath()
    ctx.moveTo(x, y - r - 1)
    ctx.lineTo(x + r + 1, y)
    ctx.lineTo(x, y + r + 1)
    ctx.lineTo(x - r - 1, y)
    ctx.fill()
    ctx.fillStyle = color
    ctx.beginPath()
    ctx.moveTo(x, y - r)
    ctx.lineTo(x + r, y)
    ctx.lineTo(x, y + r)
    ctx.lineTo(x - r, y)
    ctx.fill()
    ctx.fillStyle = 'rgba(255,255,255,0.55)'
    ctx.fillRect(x - 0.5, y - r + 0.5, 1, 1)
  }
}

/** A vein of white quartz across the tile, the host of most hard-rock ores. */
function quartzVein(ctx: Ctx, tx: number, ty: number, px: number, py: number) {
  const a = hash(tx, ty, 380)
  ctx.strokeStyle = 'rgba(240,240,232,0.75)'
  ctx.lineWidth = 2
  ctx.beginPath()
  ctx.moveTo(px + 2, py + 6 + a * 20)
  ctx.lineTo(px + TILE / 2, py + 10 + hash(tx, ty, 381) * 12)
  ctx.lineTo(px + TILE - 2, py + 6 + hash(tx, ty, 382) * 20)
  ctx.stroke()
}

function blotches(ctx: Ctx, tx: number, ty: number, px: number, py: number, colors: string[], count: number, salt: number) {
  for (let i = 0; i < count; i++) {
    const x = px + 5 + hash(tx, ty, salt + i * 3) * (TILE - 10)
    const y = py + 5 + hash(tx, ty, salt + i * 3 + 1) * (TILE - 10)
    ellipse(ctx, x, y, 3 + hash(tx, ty, salt + i * 3 + 2) * 3, 2 + hash(tx, ty, salt + 50 + i) * 2, colors[i % colors.length])
  }
}

/**
 * Draws a resource deposit into a ground chunk. The look follows the deposit
 * model: ore crystals in quartz veins for hard-rock deposits, grains of gravel
 * for placers, seams for coal, crusts for evaporites and sulfur, mottled red
 * soil for laterites. Surface deposits are piles lying on the tile.
 */
export function drawDeposit(
  ctx: Ctx,
  id: string,
  elements: string[],
  surface: boolean,
  tx: number,
  ty: number,
  px: number,
  py: number,
  model = '',
) {
  const color = depositColor(id, elements)
  if (surface) {
    drawSurfacePile(ctx, id, color, tx, ty, px, py)
    return
  }
  switch (model) {
    case 'plaser':
      // Heavy grains in river or beach gravel.
      for (let i = 0; i < 7; i++) {
        const x = px + 4 + hash(tx, ty, 390 + i * 2) * (TILE - 8)
        const y = py + 4 + hash(tx, ty, 391 + i * 2) * (TILE - 8)
        ctx.fillStyle = i % 3 === 0 ? color : 'rgba(110,100,90,0.6)'
        ctx.beginPath()
        ctx.arc(x, y, i % 3 === 0 ? 1.6 : 1.2, 0, Math.PI * 2)
        ctx.fill()
      }
      return
    case 'batubara':
      ctx.fillStyle = 'rgba(20,20,20,0.85)'
      ctx.fillRect(px, py + 12 + hash(tx, ty, 400) * 6, TILE, 5)
      ctx.fillStyle = 'rgba(255,255,255,0.18)'
      ctx.fillRect(px + 4, py + 13 + hash(tx, ty, 400) * 6, 8, 1)
      return
    case 'evaporit':
      blotches(ctx, tx, ty, px, py, [color, '#f4efe6', '#e9d6d0'], 5, 410)
      return
    case 'solfatara':
      blotches(ctx, tx, ty, px, py, ['#e9d43f', '#f2e27a', '#c9a92a'], 4, 420)
      ctx.fillStyle = 'rgba(30,25,20,0.8)'
      ctx.beginPath()
      ctx.arc(px + TILE / 2, py + TILE / 2, 2, 0, Math.PI * 2)
      ctx.fill()
      return
    case 'laterit_nikel':
    case 'bauksit':
      blotches(ctx, tx, ty, px, py, [color, '#a2502a', '#c98a4b'], 5, 430)
      return
    case 'porfiri':
      // Disseminated sulphide specks through the rock rather than one vein.
      for (let i = 0; i < 10; i++) {
        ctx.fillStyle = i % 2 ? color : '#d8c25a'
        ctx.fillRect(px + 3 + hash(tx, ty, 440 + i) * (TILE - 6), py + 3 + hash(tx, ty, 460 + i) * (TILE - 6), 1.5, 1.5)
      }
      return
    case 'epitermal':
    case 'granit_timah':
    case 'skarn':
    case 'mvt':
    case 'pegmatit':
      quartzVein(ctx, tx, ty, px, py)
      crystals(ctx, tx, ty, px, py, color, 3, 300)
      return
    default:
      crystals(ctx, tx, ty, px, py, color, 4, 300)
  }
}

/** An exhausted ground deposit: an old pit with spoil around its rim. */
export function drawPit(ctx: Ctx, tx: number, ty: number, px: number, py: number) {
  const cx = px + TILE / 2 + (hash(tx, ty, 470) - 0.5) * 4
  const cy = py + TILE / 2 + (hash(tx, ty, 471) - 0.5) * 4
  ellipse(ctx, cx, cy + 1, 11, 7.5, 'rgba(120,100,80,0.55)')
  ellipse(ctx, cx, cy, 9, 6, '#3a2f25')
  ellipse(ctx, cx + 1, cy + 1.5, 6, 3.6, '#221b15')
  ctx.fillStyle = 'rgba(255,255,255,0.12)'
  ctx.fillRect(cx - 7, cy - 4, 5, 1)
  for (let i = 0; i < 5; i++) {
    const a = hash(tx, ty, 480 + i) * Math.PI * 2
    ellipse(ctx, cx + Math.cos(a) * 12, cy + Math.sin(a) * 8, 1.6, 1.1, i % 2 ? '#7d6a55' : '#9a8670')
  }
}

function drawSurfacePile(ctx: Ctx, id: string, color: string, tx: number, ty: number, px: number, py: number) {
  switch (id) {
    case 'kayu':
      for (let i = 0; i < 3; i++) {
        const y = py + 14 + i * 4
        const x = px + 7 + hash(tx, ty, 320 + i) * 4
        ctx.fillStyle = '#6b4423'
        ctx.fillRect(x, y, 16, 4)
        ctx.fillStyle = '#c99a62'
        ctx.beginPath()
        ctx.arc(x + 16, y + 2, 2, 0, Math.PI * 2)
        ctx.fill()
      }
      return
    case 'rumput_laut':
      ctx.strokeStyle = '#3e8e41'
      ctx.lineWidth = 1.5
      for (let i = 0; i < 3; i++) {
        const x = px + 7 + i * 8 + hash(tx, ty, 350 + i) * 3
        ctx.beginPath()
        ctx.moveTo(x, py + 26)
        ctx.quadraticCurveTo(x + 4, py + 19, x, py + 13)
        ctx.quadraticCurveTo(x - 4, py + 9, x + 1, py + 5)
        ctx.stroke()
      }
      return
    default: {
      const tint = id === 'batu' ? '#8b8b8b' : id === 'tanah_liat' ? '#b0643c' : id === 'pasir' ? '#e2cf8f' : color
      for (let i = 0; i < 3; i++) {
        const x = px + 9 + hash(tx, ty, 330 + i) * 14
        const y = py + 12 + hash(tx, ty, 340 + i) * 12
        ellipse(ctx, x, y + 1.5, 5, 2.5, 'rgba(0,0,0,0.25)')
        ellipse(ctx, x, y, 5, 3.5, tint)
        ellipse(ctx, x - 1.5, y - 1, 2, 1.2, 'rgba(255,255,255,0.35)')
      }
    }
  }
}

/** Short symbols for geological features on the geology map. */
export const FEATURE_ICONS: Record<string, string> = {
  volcano: '🌋',
  pluton: '◆',
  pegmatite: '✦',
  ophiolite: '◈',
  basin: '▤',
  evaporite: '◇',
  carbonatite: '✺',
  karst: '▲',
  river: '〰',
  delta: '⋔',
}
