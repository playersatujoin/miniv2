// Small maps of the island's health for the Kesehatan tab: where mosquitoes
// breed and carry malaria, which rivers and pools are fouled, and where worm
// eggs lie in the soil, over a dimmed picture of the island. Observation
// only: the world shows these things itself (murky water, swarms).

import type { GameMap, TileSet } from '../../api/client'
import { bytesOf, type HealthMap } from '../../sim/protocol'

export type HealthLayer = 'mosquito' | 'water' | 'worms'

/** The island, dimmed, one pixel per tile. */
export function islandBackdrop(map: GameMap, tiles: TileSet): HTMLCanvasElement {
  const canvas = document.createElement('canvas')
  canvas.width = map.width
  canvas.height = map.height
  const ctx = canvas.getContext('2d')!
  const img = ctx.createImageData(map.width, map.height)
  const rgb = (hex: string) => {
    const v = parseInt(hex.replace('#', '').slice(0, 6), 16)
    return [(v >> 16) & 255, (v >> 8) & 255, v & 255]
  }
  const ground = tiles.ground.map((t) => rgb(t.color || '#444444'))
  const objects = tiles.objects.map((t) => (t.id === 0 ? null : rgb(t.color || '#444444')))
  for (let i = 0; i < map.width * map.height; i++) {
    const c = objects[map.layers.objects[i]] ?? ground[map.layers.ground[i]] ?? [40, 40, 40]
    // Greyed and darkened, so the layer on top carries the colour.
    const grey = (c[0] + c[1] + c[2]) / 3
    img.data[i * 4] = Math.round((c[0] * 0.35 + grey * 0.65) * 0.45)
    img.data[i * 4 + 1] = Math.round((c[1] * 0.35 + grey * 0.65) * 0.45)
    img.data[i * 4 + 2] = Math.round((c[2] * 0.35 + grey * 0.65) * 0.5)
    img.data[i * 4 + 3] = 255
  }
  ctx.putImageData(img, 0, 0)
  return canvas
}

/**
 * Paints a layer: per-cell layers as one pixel per cell (drawn smoothed into
 * a soft heat map), fouled water as one pixel per tile.
 */
export function layerCanvas(layer: HealthLayer, map: HealthMap, width: number, height: number): HTMLCanvasElement | null {
  const canvas = document.createElement('canvas')
  if (layer === 'water') {
    canvas.width = width
    canvas.height = height
    const ctx = canvas.getContext('2d')!
    const img = ctx.createImageData(width, height)
    const level = bytesOf(map.foulLevel)
    map.foul.forEach((tile, k) => {
      if (tile < 0 || tile >= width * height) return
      const v = (level[k] ?? 0) / 255
      // From a murky green to a sewage brown as the germs rise.
      const o = tile * 4
      img.data[o] = Math.round(170 - 30 * v)
      img.data[o + 1] = Math.round(190 - 110 * v)
      img.data[o + 2] = Math.round(50 - 20 * v)
      img.data[o + 3] = Math.round(140 + 115 * Math.sqrt(v))
    })
    ctx.putImageData(img, 0, 0)
    return canvas
  }
  const { cols, rows } = map
  if (!cols || !rows) return null
  canvas.width = cols
  canvas.height = rows
  const ctx = canvas.getContext('2d')!
  const img = ctx.createImageData(cols, rows)
  const values = bytesOf(layer === 'mosquito' ? map.mosquito : map.soil)
  const infected = layer === 'mosquito' ? bytesOf(map.infected) : null
  for (let k = 0; k < cols * rows; k++) {
    const v = (values[k] ?? 0) / 255
    if (v <= 0.01) continue
    const o = k * 4
    if (infected) {
      // Yellow where mosquitoes breed, turning red where they carry malaria.
      const z = Math.min(1, ((infected[k] ?? 0) / 255) * 2.5)
      img.data[o] = 240
      img.data[o + 1] = Math.round(205 - 165 * z)
      img.data[o + 2] = Math.round(60 - 30 * z)
    } else {
      img.data[o] = 200
      img.data[o + 1] = 150
      img.data[o + 2] = 90
    }
    img.data[o + 3] = Math.round(230 * Math.sqrt(v))
  }
  ctx.putImageData(img, 0, 0)
  return canvas
}
