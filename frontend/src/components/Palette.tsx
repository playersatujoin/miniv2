import { useEffect, useRef } from 'react'
import type { TileDef, TileSet } from '../api/client'
import type { Brush, Tool } from '../game/engine'
import {
  FLAT_OBJECTS,
  TILE,
  createSprites,
  drawFallbackObject,
  drawFlat,
  drawGround,
  drawSpawnFlag,
  type SpriteSheet,
} from '../game/render'

let swatchSprites: SpriteSheet | null = null

function TileSwatch({ def }: { def: TileDef }) {
  const ref = useRef<HTMLCanvasElement>(null)

  useEffect(() => {
    const canvas = ref.current!
    const dpr = window.devicePixelRatio || 1
    canvas.width = canvas.height = Math.round(40 * dpr)
    const ctx = canvas.getContext('2d')!
    ctx.scale((40 * dpr) / TILE, (40 * dpr) / TILE)

    const groundKey = def.layer === 'ground' ? def.key : 'grass'
    drawGround(ctx, () => groundKey, 0, 0, 0, 0, def.color)
    if (def.layer !== 'objects') return

    if (def.key === 'none') {
      ctx.strokeStyle = '#e63946'
      ctx.lineWidth = 3
      ctx.beginPath()
      ctx.moveTo(8, 8)
      ctx.lineTo(TILE - 8, TILE - 8)
      ctx.moveTo(TILE - 8, 8)
      ctx.lineTo(8, TILE - 8)
      ctx.stroke()
    } else if (FLAT_OBJECTS.has(def.key)) {
      drawFlat(ctx, def.key, 0, 0, 0, 0)
    } else {
      swatchSprites ??= createSprites()
      const sprite = swatchSprites.get(def.key)?.[0]
      // Fit the whole 2x2-tile sprite into the swatch.
      if (sprite) ctx.drawImage(sprite, 0, 0, TILE, TILE)
      else drawFallbackObject(ctx, def.color, 0, 0)
    }
  }, [def])

  return <canvas ref={ref} className="swatch-canvas" />
}

function SpawnSwatch() {
  const ref = useRef<HTMLCanvasElement>(null)
  useEffect(() => {
    const canvas = ref.current!
    const dpr = window.devicePixelRatio || 1
    canvas.width = canvas.height = Math.round(40 * dpr)
    const ctx = canvas.getContext('2d')!
    ctx.scale((40 * dpr) / TILE, (40 * dpr) / TILE)
    drawGround(ctx, () => 'grass', 0, 0, 0, 0)
    drawSpawnFlag(ctx, 0, 0)
  }, [])
  return <canvas ref={ref} className="swatch-canvas" />
}

const BRUSHES: { value: Brush; label: string }[] = [
  { value: 1, label: '1×1' },
  { value: 3, label: '3×3' },
  { value: 5, label: '5×5' },
  { value: 'fill', label: 'Isi' },
]

type Props = {
  tiles: TileSet
  tool: Tool
  brush: Brush
  onTool: (tool: Tool) => void
  onBrush: (brush: Brush) => void
}

export function Palette({ tiles, tool, brush, onTool, onBrush }: Props) {
  const isActive = (def: TileDef) => tool.kind === 'paint' && tool.layer === def.layer && tool.id === def.id

  const group = (title: string, defs: TileDef[]) => (
    <section className="palette-group">
      <h3>{title}</h3>
      <div className="swatches">
        {defs.map((def) => (
          <button
            key={def.id}
            type="button"
            className={`swatch ${isActive(def) ? 'active' : ''}`}
            title={`${def.name}${def.solid ? ' (tidak bisa dilewati)' : ''}`}
            onClick={() => onTool({ kind: 'paint', layer: def.layer, id: def.id })}
          >
            <TileSwatch def={def} />
            <span>{def.name}</span>
            {def.solid && <i className="solid-dot" />}
          </button>
        ))}
      </div>
    </section>
  )

  return (
    <aside className="palette">
      {group('Tanah', tiles.ground)}
      {group('Objek', tiles.objects)}

      <section className="palette-group">
        <h3>Alat</h3>
        <div className="swatches">
          <button
            type="button"
            className={`swatch ${tool.kind === 'spawn' ? 'active' : ''}`}
            title="Klik tile yang bisa dilewati untuk memindahkan titik spawn"
            onClick={() => onTool({ kind: 'spawn' })}
          >
            <SpawnSwatch />
            <span>Titik Spawn</span>
          </button>
        </div>
      </section>

      <section className="palette-group">
        <h3>Kuas</h3>
        <div className="segmented">
          {BRUSHES.map((b) => (
            <button
              key={b.value}
              type="button"
              className={brush === b.value ? 'active' : ''}
              disabled={tool.kind === 'spawn'}
              onClick={() => onBrush(b.value)}
            >
              {b.label}
            </button>
          ))}
        </div>
      </section>

      <p className="legend">
        <i className="solid-dot" /> = tidak bisa dilewati
      </p>
    </aside>
  )
}

