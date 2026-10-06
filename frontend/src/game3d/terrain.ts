// The island as a low-poly relief. Heights come from the map's geology,
// reconciled with the painted tiles, on a grid twice as fine as the tiles (a
// vertex at every tile corner, edge middle and centre), so rivers can run in
// rounded channels instead of square trenches. Water gets its own surface over
// the same grid: the sea at level 0, rivers and lakes a little below their
// banks, smoothed along their course so a river slopes rather than steps, with
// a downstream flow for the shader. One tile is one world unit; x runs east,
// z south, y up.

import * as THREE from 'three'
import type { GameMap, MapRelief, TileSet } from '../api/client'

export type Terrain = {
  width: number
  height: number
  /** Ground key per tile ('grass', 'water', …). */
  keys: string[]
  /** 1 sea, 2 fresh water (rivers, lakes, crater lakes), 0 land, per tile. */
  water: Uint8Array
  /** Heights of the fine grid's (2W + 1) × (2H + 1) vertices, row-major. */
  grid: Float32Array
  /** Water surface at each fine vertex in or near water (NaN where dry). */
  surface: Float32Array
  /** Whether a fine vertex lies in water (rather than just near it). */
  wet: Uint8Array
  /** Downstream flow at each fine vertex (x, z pairs), fresh water only. */
  flow: Float32Array
  /** Fresh rather than salt water at each fine vertex. */
  fresh: Uint8Array
  /** Deck height of a bridge tile (NaN elsewhere). */
  deck: Float32Array
  /** Ground height at any point (on a bridge, its deck). */
  heightAt: (x: number, z: number) => number
  /** The water's surface near a point, NaN away from water. */
  waterAt: (x: number, z: number) => number
}

const SEA_BED = -0.55
const DEEP_BED = -1.7
const BEACH = 0.16
/** How far under the water the bank meets it, and how deep rivers run. */
const SHORE_DIP = 0.05
const RIVER_DEPTH = 0.42

/** Ground colours, a little richer than the 2D palette under warm light. */
const GROUND_COLOR: Record<string, string> = {
  grass: '#6fae4c',
  forest_floor: '#4a8a3a',
  sand: '#e6d59a',
  dirt: '#a8794a',
  stone_floor: '#8f8b84',
  mountain: '#6d6660',
  bridge: '#9b6b3c',
  batuan_vulkanik: '#4b4441',
  kawah: '#5d5148',
  batu_gamping: '#d3cebf',
  water: '#c9b582',
  deep_water: '#5c6f6b',
}
const BED = new THREE.Color('#b9a476')
const DEEP = new THREE.Color('#6f6a52')

/** How much each ground shows the season (vegetated ground the most). */
const VEGETATION: Record<string, number> = { grass: 1, forest_floor: 0.75, dirt: 0.4, sand: 0.1 }

export const hash = (x: number, y: number, s: number) => {
  let h = (x * 374761393 + y * 668265263 + s * 2147483647) | 0
  h = Math.imul(h ^ (h >>> 13), 1274126177)
  return ((h ^ (h >>> 16)) >>> 0) / 4294967296
}

/** Height scale: taller relief on bigger islands. */
export const reliefScale = (w: number, h: number) => Math.min(16, Math.max(5, Math.min(w, h) * 0.07))

const N4 = [
  [1, 0],
  [-1, 0],
  [0, 1],
  [0, -1],
]

export function buildTerrain(map: GameMap, tiles: TileSet, relief: MapRelief | null | undefined): Terrain {
  const W = map.width
  const H = map.height
  const N = W * H
  const keyById = new Map(tiles.ground.map((t) => [t.id, t.key]))
  const keys = map.layers.ground.map((id) => keyById.get(id) ?? 'grass')
  const usable = relief && relief.width === W && relief.height === H
  const sea = usable ? relief.seaLevel / 1000 : 0.3
  const S = reliefScale(W, H)
  const inMap = (x: number, y: number) => x >= 0 && y >= 0 && x < W && y < H

  // Which tiles hold water. A crater holds a lake; a bridge stands over its river.
  const water = new Uint8Array(N)
  for (let i = 0; i < N; i++) {
    const k = keys[i]
    if (k === 'water' || k === 'deep_water') water[i] = usable && relief.fresh[i] ? 2 : 1
    else if (k === 'kawah') water[i] = 2
  }
  for (let i = 0; i < N; i++) {
    if (keys[i] !== 'bridge') continue
    const x = i % W
    const y = (i / W) | 0
    let kind = 0
    for (const [dx, dy] of N4) if (inMap(x + dx, y + dy)) kind = Math.max(kind, water[(y + dy) * W + x + dx])
    water[i] = kind || 2
  }

  // Land heights, softened once so the relief rolls instead of stepping.
  const land = new Float32Array(N)
  for (let i = 0; i < N; i++) {
    const k = keys[i]
    const e = usable ? relief.elevation[i] / 1000 : 0.45
    const t = Math.min(1, Math.max(0, (e - sea) / (1 - sea)))
    let h = BEACH + S * Math.pow(t, 1.55)
    if (k === 'sand') h = Math.min(h, BEACH + S * 0.025)
    if (k === 'mountain') h = Math.max(h, BEACH + S * 0.45) + hash(i % W, (i / W) | 0, 5) * 0.6
    land[i] = h
  }
  const smooth = new Float32Array(land)
  for (let y = 0; y < H; y++) {
    for (let x = 0; x < W; x++) {
      const i = y * W + x
      if (water[i] && keys[i] !== 'kawah') continue
      let sum = land[i] * 2
      let n = 2
      for (const [dx, dy] of N4) {
        if (!inMap(x + dx, y + dy)) continue
        const j = (y + dy) * W + x + dx
        if (water[j]) continue
        sum += land[j]
        n++
      }
      smooth[i] = keys[i] === 'sand' ? Math.min(land[i], sum / n) : sum / n
    }
  }

  // The water's level per tile: the sea at 0; fresh water a little under its lowest bank
  // (a crater lake under its rim), filled inwards across wide lakes.
  const level = new Float32Array(N).fill(NaN)
  const bed = new Float32Array(N)
  for (let y = 0; y < H; y++) {
    for (let x = 0; x < W; x++) {
      const i = y * W + x
      if (water[i] === 1) {
        level[i] = 0
        bed[i] = keys[i] === 'deep_water' ? DEEP_BED : SEA_BED
        continue
      }
      if (water[i] !== 2) continue
      if (keys[i] === 'kawah') {
        level[i] = smooth[i] - 0.3
        continue
      }
      let bank = Infinity
      for (let dy = -1; dy <= 1; dy++) {
        for (let dx = -1; dx <= 1; dx++) {
          if (!inMap(x + dx, y + dy)) continue
          const j = (y + dy) * W + x + dx
          if (!water[j]) bank = Math.min(bank, smooth[j])
        }
      }
      if (Number.isFinite(bank)) level[i] = Math.max(0.03, bank - 0.14)
    }
  }
  for (let pass = 0; pass < 24; pass++) {
    let missing = 0
    for (let i = 0; i < N; i++) {
      if (water[i] !== 2 || !Number.isNaN(level[i])) continue
      const x = i % W
      const y = (i / W) | 0
      let sum = 0
      let n = 0
      for (const [dx, dy] of N4) {
        if (!inMap(x + dx, y + dy)) continue
        const v = level[(y + dy) * W + x + dx]
        if (!Number.isNaN(v)) {
          sum += v
          n++
        }
      }
      if (n) level[i] = sum / n
      else missing++
    }
    if (!missing) break
  }
  for (let i = 0; i < N; i++) {
    if (water[i] !== 2) continue
    if (Number.isNaN(level[i])) level[i] = BEACH
    bed[i] = level[i] - (keys[i] === 'kawah' ? 0.7 : RIVER_DEPTH)
  }

  // The fine grid. Each vertex touches 1 (a tile's centre), 2 (an edge's middle) or 4 tiles (a corner).
  const GW = 2 * W + 1
  const GH = 2 * H + 1
  const G = GW * GH
  const touching = (gx: number, gz: number) => {
    const xs = gx % 2 ? [(gx - 1) / 2] : [gx / 2 - 1, gx / 2]
    const zs = gz % 2 ? [(gz - 1) / 2] : [gz / 2 - 1, gz / 2]
    const out: number[] = []
    for (const tz of zs) for (const tx of xs) out.push(inMap(tx, tz) ? tz * W + tx : -1) // -1: the open sea beyond the map
    return out
  }
  const surface = new Float32Array(G).fill(NaN)
  const wet = new Uint8Array(G)
  const fresh = new Uint8Array(G)
  const pureSea = new Uint8Array(G)
  for (let gz = 0; gz < GH; gz++) {
    for (let gx = 0; gx < GW; gx++) {
      let sum = 0
      let n = 0
      let salt = 0
      let sweet = 0
      for (const t of touching(gx, gz)) {
        const w = t < 0 ? 1 : water[t]
        if (!w) continue
        sum += t < 0 ? 0 : level[t]
        n++
        if (w === 1) salt++
        else sweet++
      }
      const g = gz * GW + gx
      if (!n) continue
      surface[g] = sum / n
      wet[g] = 1
      fresh[g] = sweet > 0 ? 1 : 0
      pureSea[g] = sweet === 0 ? 1 : 0
    }
  }
  // Smooth fresh water along its course, so a river slopes instead of stepping from tile to tile.
  for (let pass = 0; pass < 4; pass++) {
    const next = new Float32Array(surface)
    for (let gz = 0; gz < GH; gz++) {
      for (let gx = 0; gx < GW; gx++) {
        const g = gz * GW + gx
        if (!wet[g] || pureSea[g]) continue
        let sum = surface[g] * 2
        let n = 2
        for (const [dx, dz] of N4) {
          const x = gx + dx
          const z = gz + dz
          if (x < 0 || z < 0 || x >= GW || z >= GH) continue
          const v = surface[z * GW + x]
          if (!wet[z * GW + x]) continue
          sum += v
          n++
        }
        next[g] = sum / n
      }
    }
    surface.set(next)
  }

  // Ground on the fine grid: tile centres at the tile's height (or its bed under water), shared
  // vertices at the average of the land around them, banks meeting the water. The water's edge
  // is rounded off: a corner with land on three sides stands just clear of the water (cutting the
  // corner), one with water on three sides sinks under it (filling the bend), and the line
  // between wavers a little.
  const grid = new Float32Array(G)
  for (let gz = 0; gz < GH; gz++) {
    for (let gx = 0; gx < GW; gx++) {
      const g = gz * GW + gx
      let landSum = 0
      let landN = 0
      let bedSum = 0
      let bedN = 0
      for (const t of touching(gx, gz)) {
        if (t >= 0 && !water[t]) {
          landSum += smooth[t]
          landN++
        } else {
          bedSum += t < 0 ? DEEP_BED : bed[t]
          bedN++
        }
      }
      if (!bedN) grid[g] = landSum / landN
      else if (!landN) grid[g] = bedSum / bedN
      else {
        const s = surface[g]
        const wobble = (hash(gx, gz, 17) - 0.5) * 0.07
        const edge = landN === 3 ? s + 0.05 : bedN === 3 ? s - 0.22 : s - SHORE_DIP + wobble
        grid[g] = Math.min(landSum / landN, edge)
      }
    }
  }

  // Let the water's surface reach a little past its edge, so it meets the banks wherever they rise
  // out of it (the shoreline is where the two cross).
  for (let pass = 0; pass < 3; pass++) {
    const next = new Float32Array(surface)
    for (let gz = 0; gz < GH; gz++) {
      for (let gx = 0; gx < GW; gx++) {
        const g = gz * GW + gx
        if (!Number.isNaN(surface[g])) continue
        let sum = 0
        let n = 0
        let sweet = 0
        for (let dz = -1; dz <= 1; dz++) {
          for (let dx = -1; dx <= 1; dx++) {
            const x = gx + dx
            const z = gz + dz
            if (x < 0 || z < 0 || x >= GW || z >= GH) continue
            const v = surface[z * GW + x]
            if (Number.isNaN(v)) continue
            sum += v
            n++
            sweet += fresh[z * GW + x]
          }
        }
        if (n) {
          next[g] = sum / n
          if (sweet) fresh[g] = 1
        }
      }
    }
    surface.set(next)
  }

  // Downstream: down the water's slope, faster where it is steeper.
  const flow = new Float32Array(G * 2)
  const at = (x: number, z: number, g: number) => {
    if (x < 0 || z < 0 || x >= GW || z >= GH) return surface[g]
    const v = surface[z * GW + x]
    return Number.isNaN(v) ? surface[g] : v
  }
  for (let gz = 0; gz < GH; gz++) {
    for (let gx = 0; gx < GW; gx++) {
      const g = gz * GW + gx
      if (!fresh[g] || Number.isNaN(surface[g])) continue
      const ddx = at(gx + 1, gz, g) - at(gx - 1, gz, g)
      const ddz = at(gx, gz + 1, g) - at(gx, gz - 1, g)
      const len = Math.hypot(ddx, ddz)
      const speed = Math.min(2.2, Math.max(0.3, len * 3.5))
      flow[g * 2] = len > 1e-4 ? (-ddx / len) * speed : speed * 0.3
      flow[g * 2 + 1] = len > 1e-4 ? (-ddz / len) * speed : speed * 0.2
    }
  }

  // Bridges: a deck just above the water, level with the banks it joins.
  const deck = new Float32Array(N).fill(NaN)
  for (let i = 0; i < N; i++) {
    if (keys[i] !== 'bridge') continue
    const x = i % W
    const y = (i / W) | 0
    let top = level[i] + 0.12
    for (const [dx, dy] of N4)
      if (inMap(x + dx, y + dy) && !water[(y + dy) * W + x + dx]) top = Math.max(top, smooth[(y + dy) * W + x + dx])
    deck[i] = top + 0.03
  }

  /** Bilinear over the fine grid. */
  const sample = (field: Float32Array, x: number, z: number) => {
    const fx = Math.min(GW - 1.001, Math.max(0, x * 2))
    const fz = Math.min(GH - 1.001, Math.max(0, z * 2))
    const ix = Math.floor(fx)
    const iz = Math.floor(fz)
    const u = fx - ix
    const v = fz - iz
    const c = (a: number, b: number) => field[b * GW + a]
    const top = c(ix, iz) * (1 - u) + c(ix + 1, iz) * u
    const bottom = c(ix, iz + 1) * (1 - u) + c(ix + 1, iz + 1) * u
    return top * (1 - v) + bottom * v
  }
  const heightAt = (x: number, z: number) => {
    const tx = Math.floor(x)
    const tz = Math.floor(z)
    if (inMap(tx, tz) && !Number.isNaN(deck[tz * W + tx])) return deck[tz * W + tx]
    return sample(grid, x, z)
  }
  const waterAt = (x: number, z: number) => sample(surface, x, z)

  return { width: W, height: H, keys, water, grid, surface, wet, flow, fresh, deck, heightAt, waterAt }
}

/** Which way a fine cell is split into triangles (the land and the water over it agree). */
const flips = (cx: number, cz: number) => hash(cx, cz, 3) < 0.5

/**
 * The faceted land mesh on the fine grid, coloured by ground (river and sea beds darker with
 * depth, banks darker where wet), with a vegetation weight for the season and the water level
 * above each vertex for the light dancing on the bed.
 */
export function terrainGeometry(t: Terrain, bounds = { x: 0, z: 0, width: t.width, height: t.height }) {
  const { width: W, keys, grid, surface } = t
  const GW = 2 * W + 1
  const { x: startX, z: startZ, width, height } = bounds
  const cells = 4 * width * height
  const pos = new Float32Array(cells * 18)
  const col = new Float32Array(cells * 18)
  const veg = new Float32Array(cells * 6)
  const lvl = new Float32Array(cells * 6)
  const c = new THREE.Color()
  let p = 0
  let v = 0
  for (let cz = 2 * startZ; cz < 2 * (startZ + height); cz++) {
    for (let cx = 2 * startX; cx < 2 * (startX + width); cx++) {
      const i = (cz >> 1) * W + (cx >> 1)
      const k = keys[i]
      const g00 = cz * GW + cx
      const g10 = g00 + 1
      const g01 = g00 + GW
      const g11 = g01 + 1
      const h = [grid[g00], grid[g10], grid[g01], grid[g11]]
      const ws = [surface[g00], surface[g10], surface[g01], surface[g11]].map((s) => (Number.isNaN(s) ? -100 : s))
      const mid = (h[0] + h[1] + h[2] + h[3]) / 4
      const level = Math.max(...ws)
      c.set(GROUND_COLOR[k] ?? '#6fae4c')
      if (k === 'bridge') c.set(GROUND_COLOR.water)
      c.multiplyScalar(0.92 + 0.14 * hash(cx, cz, 11))
      const under = level - mid
      if (under > 0) c.lerp(BED, Math.min(1, under * 4)).lerp(DEEP, Math.min(1, Math.max(0, under - 0.25) * 1.2))
      else if (under > -0.12) c.multiplyScalar(0.82 + 1.5 * -under) // damp earth at the water's edge
      const x0 = cx / 2
      const z0 = cz / 2
      const x1 = x0 + 0.5
      const z1 = z0 + 0.5
      const corner = [
        [x0, h[0], z0, ws[0]],
        [x1, h[1], z0, ws[1]],
        [x0, h[2], z1, ws[2]],
        [x1, h[3], z1, ws[3]],
      ]
      const order = flips(cx, cz) ? [0, 2, 1, 1, 2, 3] : [0, 3, 1, 0, 2, 3]
      const wetGround = under > -0.05 ? 0 : (VEGETATION[k] ?? 0)
      for (const o of order) {
        const [x, y, z, w] = corner[o]
        pos[p] = x
        pos[p + 1] = y
        pos[p + 2] = z
        col[p] = c.r
        col[p + 1] = c.g
        col[p + 2] = c.b
        p += 3
        veg[v] = wetGround
        lvl[v] = w
        v++
      }
    }
  }
  const g = new THREE.BufferGeometry()
  g.setAttribute('position', new THREE.BufferAttribute(pos, 3))
  g.setAttribute('color', new THREE.BufferAttribute(col, 3))
  g.setAttribute('vegetation', new THREE.BufferAttribute(veg, 1))
  g.setAttribute('waterLevel', new THREE.BufferAttribute(lvl, 1))
  g.computeVertexNormals()
  return g
}

/**
 * The water: the fine grid's cells in or beside water, at the water's surface, with the ground
 * height below each vertex (so the shader knows the depth everywhere, and where the shore is),
 * the downstream flow and whether it is fresh.
 */
export function waterGeometry(t: Terrain) {
  const { width: W, height: H, grid, surface, flow, fresh } = t
  const GW = 2 * W + 1
  const pos: number[] = []
  const ground: number[] = []
  const flows: number[] = []
  const sweet: number[] = []
  for (let cz = 0; cz < 2 * H; cz++) {
    for (let cx = 0; cx < 2 * W; cx++) {
      const g = [cz * GW + cx, cz * GW + cx + 1, (cz + 1) * GW + cx, (cz + 1) * GW + cx + 1]
      if (g.some((k) => Number.isNaN(surface[k]))) continue
      // Skip cells whose ground stands well clear of the water: they would be hidden anyway.
      if (g.every((k) => grid[k] > surface[k] + 0.2)) continue
      const x0 = cx / 2
      const z0 = cz / 2
      const corner = [
        [x0, z0],
        [x0 + 0.5, z0],
        [x0, z0 + 0.5],
        [x0 + 0.5, z0 + 0.5],
      ]
      const order = flips(cx, cz) ? [0, 2, 1, 1, 2, 3] : [0, 3, 1, 0, 2, 3]
      for (const o of order) {
        const k = g[o]
        pos.push(corner[o][0], surface[k], corner[o][1])
        ground.push(grid[k])
        flows.push(flow[k * 2], flow[k * 2 + 1])
        sweet.push(fresh[k])
      }
    }
  }
  const geo = new THREE.BufferGeometry()
  geo.setAttribute('position', new THREE.Float32BufferAttribute(pos, 3))
  geo.setAttribute('ground', new THREE.Float32BufferAttribute(ground, 1))
  geo.setAttribute('flow', new THREE.Float32BufferAttribute(flows, 2))
  geo.setAttribute('fresh', new THREE.Float32BufferAttribute(sweet, 1))
  geo.computeVertexNormals()
  return geo
}

/** The open sea around the island, out to the horizon. */
export function oceanGeometry(t: Terrain) {
  const size = Math.max(t.width, t.height) * 8
  const g = new THREE.PlaneGeometry(size, size, 1, 1)
  g.rotateX(-Math.PI / 2)
  g.translate(t.width / 2, -0.02, t.height / 2)
  const n = g.attributes.position.count
  g.setAttribute('ground', new THREE.Float32BufferAttribute(new Array(n).fill(-6), 1))
  g.setAttribute('flow', new THREE.Float32BufferAttribute(new Array(n * 2).fill(0), 2))
  g.setAttribute('fresh', new THREE.Float32BufferAttribute(new Array(n).fill(0), 1))
  return g
}

/** Plank decks over bridge tiles, with a rail along each side, spanning between the banks. */
export function bridgeGeometry(t: Terrain) {
  const { width: W, height: H, deck, water } = t
  const parts: THREE.BufferGeometry[] = []
  for (let i = 0; i < W * H; i++) {
    if (Number.isNaN(deck[i])) continue
    const x = i % W
    const y = (i / W) | 0
    const dryAlongX = [x - 1, x + 1].some((nx) => nx >= 0 && nx < W && !water[y * W + nx])
    const plank = new THREE.BoxGeometry(dryAlongX ? 1.02 : 0.8, 0.06, dryAlongX ? 0.8 : 1.02)
    plank.translate(x + 0.5, deck[i] - 0.03, y + 0.5)
    parts.push(plank)
    for (const side of [-1, 1]) {
      const rail = new THREE.BoxGeometry(dryAlongX ? 1 : 0.04, 0.12, dryAlongX ? 0.04 : 1)
      rail.translate(x + 0.5 + (dryAlongX ? 0 : side * 0.38), deck[i] + 0.06, y + 0.5 + (dryAlongX ? side * 0.38 : 0))
      parts.push(rail)
    }
  }
  if (!parts.length) return null
  const merged = mergeBoxes(parts)
  return merged
}

function mergeBoxes(parts: THREE.BufferGeometry[]) {
  const pos: number[] = []
  const nrm: number[] = []
  for (const p of parts) {
    const g = p.toNonIndexed()
    pos.push(...(g.getAttribute('position').array as Float32Array))
    nrm.push(...(g.getAttribute('normal').array as Float32Array))
  }
  const g = new THREE.BufferGeometry()
  g.setAttribute('position', new THREE.Float32BufferAttribute(pos, 3))
  g.setAttribute('normal', new THREE.Float32BufferAttribute(nrm, 3))
  return g
}
