import type {
  DepositItem,
  DepositModel,
  GameMap,
  GeoFeature,
  MapGeology,
  Point,
  RockType,
  TileDef,
  TileSet,
} from '../api/client'
import { FLAG, SEX_SYMBOL, type Sex, type SimFrame, type StructureFrame } from '../sim/protocol'
import {
  FEATURE_ICONS,
  FLAT_OBJECTS,
  FLAT_STRUCTURES,
  ROCKY_GROUND,
  SPRITE_SIZE,
  STRUCTURE_H,
  STRUCTURE_REACH,
  STRUCTURE_W,
  TILE,
  VARIANTS,
  creatureScale,
  createSprites,
  drawCreature,
  drawFallbackObject,
  drawFarm,
  drawDeposit,
  drawFlat,
  drawGround,
  drawPlayer,
  drawSpawnFlag,
  hash,
  isWater,
  structureLabel,
  structureSprite,
  tintLithology,
  type Facing,
  type KeyAt,
  type SpriteSheet,
} from './render'

export type Mode = 'play' | 'edit' | 'watch'
export type Layer = 'ground' | 'objects'
export type Tool = { kind: 'paint'; layer: Layer; id: number } | { kind: 'spawn' }
export type Brush = 1 | 3 | 5 | 'fill'

export type HoverDeposit = { item: DepositItem; model?: DepositModel }

export type HoverInfo = {
  x: number
  y: number
  ground: TileDef
  object: TileDef
  deposits: HoverDeposit[]
  /** Bedrock unit under the tile, once the map's geology is loaded. */
  rock?: RockType
}

export type EngineEvents = {
  onDirtyChange?: (dirty: boolean) => void
  onHistoryChange?: (canUndo: boolean, canRedo: boolean) => void
  onHover?: (info: HoverInfo | null) => void
  onPlayerTile?: (p: Point) => void
  /** Watch mode: the user clicked a creature, or a building (→ its owner), or empty ground (→ null). */
  onSelectCreature?: (id: number | null) => void
  /** Watch mode: following stopped because the user moved the camera. */
  onFollowChange?: (follow: boolean) => void
}

export type MapSnapshot = {
  /** Pass back to markSaved() once this snapshot is persisted. */
  marker: unknown
  spawn: Point
  layers: { ground: number[]; objects: number[] }
}

type View = { w: number; h: number; scale: number; left: number; top: number }

type Change = { i: number; layer: Layer; before: number; after: number }
type Stroke = { changes: Change[]; spawn?: { before: Point; after: Point } }

/** A streamed creature, interpolated from (px, py, ph) to (x, y, heading) between frames. */
type Creature = {
  id: number
  name: string
  sex: Sex
  hue: number
  size: number
  flags: number
  energy: number
  health: number
  houseId: number
  px: number
  py: number
  ph: number
  x: number
  y: number
  heading: number
  /** Interpolated render state. */
  rx: number
  ry: number
  rh: number
  step: number
  moving: boolean
}

const CHUNK = 16 // tiles per side of a cached ground chunk
const MAX_CHUNKS = 96
const OVERLAY_RES = 4 // geology overlay pixels per tile (room for unit boundaries)
const ZOOMS = [0.5, 0.75, 1, 1.5, 2, 3]
const WALK_SPEED = 5 // tiles per second
const RUN_SPEED = 8.5
const PAN_SPEED = 14
const HALF_W = 0.28 // player hitbox half-extents in tiles, centred on the feet
const HALF_H = 0.18
const EPS = 1e-4
const MAX_FILL = 20_000
const MAX_UNDO = 200
const BULK_REDRAW = 64 // above this many changed tiles, drop the cache instead of patching it
const SNAP_DISTANCE = 10 // tiles; creatures that jump further between frames are teleported, not slid
const DRAG_THRESHOLD = 4 // px before a watch-mode click becomes a pan

const KEYS = {
  up: ['KeyW', 'ArrowUp'],
  down: ['KeyS', 'ArrowDown'],
  left: ['KeyA', 'ArrowLeft'],
  right: ['KeyD', 'ArrowRight'],
  run: ['ShiftLeft', 'ShiftRight'],
}
const CAPTURED_KEYS = new Set([...KEYS.up, ...KEYS.down, ...KEYS.left, ...KEYS.right, 'Space'])

function isTyping(target: EventTarget | null) {
  return target instanceof HTMLElement && (target.isContentEditable || /^(INPUT|TEXTAREA|SELECT)$/.test(target.tagName))
}

function hexToRgb(hex: string): [number, number, number] {
  const n = parseInt(hex.slice(1, 7), 16)
  return [n >> 16, (n >> 8) & 255, n & 255]
}

export class GameEngine {
  /** updatedAt of the server version the engine's data is based on. */
  version = ''

  private readonly ctx: CanvasRenderingContext2D
  private readonly sprites: SpriteSheet = createSprites()
  private readonly groundDefs: TileDef[] = []
  private readonly objectDefs: TileDef[] = []
  private readonly groundRgb: [number, number, number][] = []
  private readonly objectRgb: [number, number, number][] = []

  private width = 0
  private height = 0
  private ground = new Uint8Array()
  private objects = new Uint8Array()
  private spawn: Point = { x: 0, y: 0 }

  private chunks = new Map<number, { canvas: HTMLCanvasElement; lastUsed: number }>()
  /** Resource deposits by tile index, baked into the ground chunks. */
  private deposits = new Map<number, (HoverDeposit & { surface: boolean })[]>()
  private geologyVersion = 0
  private rocks: ArrayLike<number> | null = null
  private rockTypes: RockType[] = []
  private features: GeoFeature[] = []
  private geologyOverlay = false
  /** Rock units at OVERLAY_RES px per tile with unit boundaries; built when first shown. */
  private overlayCanvas: HTMLCanvasElement | null = null
  private readonly minimapBase = document.createElement('canvas')

  private mode: Mode = 'play'
  private tool: Tool = { kind: 'paint', layer: 'ground', id: 0 }
  private brush: Brush = 1

  private player = { x: 0, y: 0, facing: { x: 0, y: 1 } as Facing, step: 0, moving: false }
  private placed = false
  private camera = { x: 0, y: 0, zoomIndex: 3 }

  private keys = new Set<string>()
  private pointer: { sx: number; sy: number } | null = null
  private painting = false
  private lastPaint: Point | null = null
  private panning: { sx: number; sy: number; cx: number; cy: number } | null = null
  private wheelAccum = 0

  private stroke: Stroke | null = null
  private undoStack: Stroke[] = []
  private redoStack: Stroke[] = []
  private savedMarker: Stroke | null = null
  private lastDirty = false

  private creatures = new Map<number, Creature>()
  /** Upright structures by tile row, for depth sorting; flat ones (farms) apart. */
  private structureRows = new Map<number, StructureFrame[]>()
  private flatStructures: StructureFrame[] = []
  private structureByTile = new Map<number, StructureFrame>()
  private structures: StructureFrame[] = []
  private hoverStructure: StructureFrame | null = null
  private frameAt = 0
  private frameInterval = 100
  private selectedId: number | null = null
  private following = false
  private hoverCreature: Creature | null = null
  private press: { sx: number; sy: number; cx: number; cy: number } | null = null

  private hoverKey = ''
  private playerTileKey = ''
  private raf = 0
  private lastTime = 0
  private frame = 0

  constructor(
    private readonly canvas: HTMLCanvasElement,
    private readonly minimap: HTMLCanvasElement | null,
    tiles: TileSet,
    map: GameMap,
    private readonly events: EngineEvents = {},
  ) {
    this.ctx = canvas.getContext('2d')!
    for (const t of tiles.ground) {
      this.groundDefs[t.id] = t
      this.groundRgb[t.id] = hexToRgb(t.color)
    }
    for (const t of tiles.objects) {
      this.objectDefs[t.id] = t
      this.objectRgb[t.id] = hexToRgb(t.color)
    }
    this.setMap(map)

    window.addEventListener('keydown', this.onKeyDown)
    window.addEventListener('keyup', this.onKeyUp)
    window.addEventListener('blur', this.onBlur)
    canvas.addEventListener('pointerdown', this.onPointerDown)
    canvas.addEventListener('pointermove', this.onPointerMove)
    canvas.addEventListener('pointerup', this.onPointerUp)
    canvas.addEventListener('pointercancel', this.onPointerUp)
    canvas.addEventListener('pointerleave', this.onPointerLeave)
    canvas.addEventListener('wheel', this.onWheel, { passive: false })
    canvas.addEventListener('contextmenu', this.onContextMenu)
    minimap?.addEventListener('pointerdown', this.onMinimapPointer)
    this.raf = requestAnimationFrame(this.loop)
  }

  destroy() {
    cancelAnimationFrame(this.raf)
    window.removeEventListener('keydown', this.onKeyDown)
    window.removeEventListener('keyup', this.onKeyUp)
    window.removeEventListener('blur', this.onBlur)
    this.canvas.removeEventListener('pointerdown', this.onPointerDown)
    this.canvas.removeEventListener('pointermove', this.onPointerMove)
    this.canvas.removeEventListener('pointerup', this.onPointerUp)
    this.canvas.removeEventListener('pointercancel', this.onPointerUp)
    this.canvas.removeEventListener('pointerleave', this.onPointerLeave)
    this.canvas.removeEventListener('wheel', this.onWheel)
    this.canvas.removeEventListener('contextmenu', this.onContextMenu)
    this.minimap?.removeEventListener('pointerdown', this.onMinimapPointer)
    this.chunks.clear()
  }

  // --- Public API ------------------------------------------------------------

  /**
   * Shows the map's geology: rock units (tint and overlay), features and
   * resource deposits. null/undefined (not loaded, or failed) hides it all.
   */
  setGeology(geology: MapGeology | null | undefined) {
    this.deposits.clear()
    this.rocks = null
    this.rockTypes = []
    this.features = []
    this.overlayCanvas = null
    if (geology && geology.width === this.width && geology.height === this.height) {
      if (geology.rocks?.length === this.width * this.height) {
        this.rocks = geology.rocks
        this.rockTypes = geology.rockTypes ?? []
      }
      this.features = geology.features ?? []
      for (const [x, y, item, surface, model] of geology.deposits ?? []) {
        const def = geology.items[item]
        if (!def || !this.inBounds(x, y)) continue
        const i = y * this.width + x
        const list = this.deposits.get(i) ?? []
        list.push({ item: def, surface: surface === 1, model: geology.models?.[model] })
        this.deposits.set(i, list)
      }
    }
    this.geologyVersion++
    this.chunks.clear()
  }

  /** Toggles the geological map: rock-unit colours over the land plus feature labels. */
  setGeologyOverlay(on: boolean) {
    this.geologyOverlay = on
  }

  /** Replaces the map data, discarding local edits and history. */
  setMap(map: GameMap) {
    this.version = map.updatedAt
    this.width = map.width
    this.height = map.height
    this.ground = Uint8Array.from(map.layers.ground)
    this.objects = Uint8Array.from(map.layers.objects)
    this.spawn = { ...map.spawn }
    this.stroke = null
    this.undoStack = []
    this.redoStack = []
    this.savedMarker = null
    this.chunks.clear()
    this.renderMinimapBase()
    if (!this.placed || this.boxBlocked(this.player.x, this.player.y)) {
      this.respawn()
      this.placed = true
    }
    this.emitState()
  }

  setMode(mode: Mode) {
    if (mode === this.mode) return
    this.endStroke()
    this.mode = mode
    this.panning = null
    this.press = null
    this.keys.clear()
    if (mode === 'play' && this.boxBlocked(this.player.x, this.player.y)) this.respawn()
    // Stale creatures would slide across the map when the stream resumes.
    if (mode !== 'watch') {
      this.creatures.clear()
      this.frameAt = 0
      this.hoverCreature = null
      this.setStructures([])
    }
    this.canvas.style.cursor = mode === 'edit' ? 'crosshair' : 'default'
  }

  /** Feeds the latest simulation frame; creatures glide towards it until the next one. */
  setCreatureFrame(frame: SimFrame) {
    const now = performance.now()
    if (this.frameAt) {
      const gap = Math.min(250, Math.max(50, now - this.frameAt))
      this.frameInterval = this.frameInterval * 0.8 + gap * 0.2
    }
    this.frameAt = now

    const next = new Map<number, Creature>()
    for (const f of frame.creatures) {
      const c = this.creatures.get(f.id)
      if (c) {
        // Start the new segment from where the creature is drawn right now.
        c.px = c.rx
        c.py = c.ry
        c.ph = c.rh
      }
      const cur: Creature = c ?? {
        ...f,
        px: f.x,
        py: f.y,
        ph: f.heading,
        rx: f.x,
        ry: f.y,
        rh: f.heading,
        step: 0,
        moving: false,
      }
      cur.name = f.name
      cur.hue = f.hue
      cur.size = f.size
      cur.flags = f.flags
      cur.energy = f.energy
      cur.health = f.health
      cur.houseId = f.houseId
      cur.x = f.x
      cur.y = f.y
      cur.heading = f.heading
      if (Math.hypot(cur.x - cur.px, cur.y - cur.py) > SNAP_DISTANCE) {
        cur.px = cur.rx = cur.x
        cur.py = cur.ry = cur.y
      }
      next.set(f.id, cur)
    }
    this.creatures = next
    if (this.hoverCreature && !next.has(this.hoverCreature.id)) this.hoverCreature = null
    if (this.selectedId !== null && !next.has(this.selectedId)) this.stopFollow()
  }

  /** Replaces the buildings on the map (sent on connect and whenever they change). */
  setStructures(list: StructureFrame[]) {
    this.structures = list
    this.structureRows.clear()
    this.structureByTile.clear()
    this.flatStructures = []
    for (const st of list) {
      this.structureByTile.set(st.y * this.width + st.x, st)
      if (FLAT_STRUCTURES.has(st.kind)) {
        this.flatStructures.push(st)
        continue
      }
      let row = this.structureRows.get(st.y)
      if (!row) this.structureRows.set(st.y, (row = []))
      row.push(st)
    }
    for (const row of this.structureRows.values()) row.sort((a, b) => a.x - b.x)
    const hovered = this.hoverStructure
    this.hoverStructure = hovered ? (this.structureByTile.get(hovered.y * this.width + hovered.x) ?? null) : null
  }

  setSelected(id: number | null) {
    this.selectedId = id
    if (id === null) this.following = false
  }

  setFollow(follow: boolean) {
    this.following = follow && this.selectedId !== null
  }

  setTool(tool: Tool) {
    this.tool = tool
  }

  setBrush(brush: Brush) {
    this.brush = brush
  }

  respawn() {
    this.player.x = this.spawn.x + 0.5
    this.player.y = this.spawn.y + 0.5
    this.player.facing = { x: 0, y: 1 }
    this.camera.x = this.player.x
    this.camera.y = this.player.y
  }

  zoomBy(dir: number, anchor?: { sx: number; sy: number }) {
    const next = Math.min(ZOOMS.length - 1, Math.max(0, this.camera.zoomIndex + dir))
    if (next === this.camera.zoomIndex) return
    const before = anchor && this.screenToTile(anchor.sx, anchor.sy)
    this.camera.zoomIndex = next
    // Keep the tile under the cursor fixed while zooming a free camera.
    if (before && (this.mode === 'edit' || (this.mode === 'watch' && !this.following))) {
      const after = this.screenToTile(anchor.sx, anchor.sy)
      this.camera.x += before.x - after.x
      this.camera.y += before.y - after.y
    }
  }

  isDirty() {
    return (this.undoStack.at(-1) ?? null) !== this.savedMarker
  }

  snapshot(): MapSnapshot {
    this.endStroke()
    return {
      marker: this.undoStack.at(-1) ?? null,
      spawn: { ...this.spawn },
      layers: { ground: Array.from(this.ground), objects: Array.from(this.objects) },
    }
  }

  markSaved(marker: unknown, version: string) {
    this.savedMarker = marker as Stroke | null
    this.version = version
    this.emitState()
  }

  undo() {
    this.endStroke()
    const s = this.undoStack.pop()
    if (!s) return
    this.applyStroke(s, true)
    this.redoStack.push(s)
    this.emitState()
  }

  redo() {
    this.endStroke()
    const s = this.redoStack.pop()
    if (!s) return
    this.applyStroke(s, false)
    this.undoStack.push(s)
    this.emitState()
  }

  // --- Map queries -----------------------------------------------------------

  private inBounds(x: number, y: number) {
    return x >= 0 && y >= 0 && x < this.width && y < this.height
  }

  private keyAt: KeyAt = (x, y) =>
    this.inBounds(x, y) ? (this.groundDefs[this.ground[y * this.width + x]]?.key ?? null) : null

  private solidAt(x: number, y: number) {
    if (!this.inBounds(x, y)) return true
    const i = y * this.width + x
    return !!(this.groundDefs[this.ground[i]]?.solid || this.objectDefs[this.objects[i]]?.solid)
  }

  private boxBlocked(x: number, y: number) {
    for (let ty = Math.floor(y - HALF_H); ty <= Math.floor(y + HALF_H); ty++) {
      for (let tx = Math.floor(x - HALF_W); tx <= Math.floor(x + HALF_W); tx++) {
        if (this.solidAt(tx, ty)) return true
      }
    }
    return false
  }

  // --- Loop ------------------------------------------------------------------

  private loop = (now: number) => {
    this.raf = requestAnimationFrame(this.loop)
    const dt = this.lastTime ? Math.min(0.05, (now - this.lastTime) / 1000) : 0
    this.lastTime = now
    this.frame++
    this.update(dt, now)
    this.render(now / 1000)
  }

  private axis(neg: string[], pos: string[]) {
    const held = (codes: string[]) => codes.some((c) => this.keys.has(c))
    return Number(held(pos)) - Number(held(neg))
  }

  private update(dt: number, now: number) {
    const ax = this.axis(KEYS.left, KEYS.right)
    const ay = this.axis(KEYS.up, KEYS.down)
    const len = Math.hypot(ax, ay)
    const running = KEYS.run.some((c) => this.keys.has(c))

    if (this.mode === 'play') {
      const p = this.player
      p.moving = len > 0
      if (p.moving) {
        const dist = ((running ? RUN_SPEED : WALK_SPEED) * dt) / len
        this.moveAxis(ax * dist, 0)
        this.moveAxis(0, ay * dist)
        p.facing = { x: Math.sign(ax), y: Math.sign(ay) }
        p.step += dt * (running ? 3 : 2)
      }
      const follow = 1 - Math.exp(-dt * 10)
      this.camera.x += (p.x - this.camera.x) * follow
      this.camera.y += (p.y - this.camera.y) * follow

      const key = `${Math.floor(p.x)},${Math.floor(p.y)}`
      if (key !== this.playerTileKey) {
        this.playerTileKey = key
        this.events.onPlayerTile?.({ x: Math.floor(p.x), y: Math.floor(p.y) })
      }
    } else if (len > 0 && !this.panning) {
      const dist = ((PAN_SPEED * (running ? 2.5 : 1)) / ZOOMS[this.camera.zoomIndex]) * dt
      this.camera.x += (ax / len) * dist
      this.camera.y += (ay / len) * dist
      this.stopFollow()
    }

    if (this.mode === 'watch') {
      this.updateCreatures(dt, now)
      const target = this.selectedId !== null ? this.creatures.get(this.selectedId) : undefined
      if (this.following && target) {
        const follow = 1 - Math.exp(-dt * 6)
        this.camera.x += (target.rx - this.camera.x) * follow
        this.camera.y += (target.ry - 0.5 - this.camera.y) * follow
      }
    }

    this.clampCamera()
    this.updateHover()
    if (this.mode === 'watch') this.updateHoverTargets()
  }

  private updateCreatures(dt: number, now: number) {
    const t = Math.min(1, (now - this.frameAt) / this.frameInterval)
    for (const c of this.creatures.values()) {
      const rx = c.px + (c.x - c.px) * t
      const ry = c.py + (c.y - c.py) * t
      // Turn the short way round.
      let dh = (c.heading - c.ph) % (Math.PI * 2)
      if (dh > Math.PI) dh -= Math.PI * 2
      else if (dh < -Math.PI) dh += Math.PI * 2
      c.rh = c.ph + dh * t

      const moved = Math.hypot(rx - c.rx, ry - c.ry)
      c.moving = dt > 0 && moved / dt > 0.15
      c.step += moved * 1.4 // stride cycles per tile walked
      c.rx = rx
      c.ry = ry
    }
  }

  private stopFollow() {
    if (!this.following) return
    this.following = false
    this.events.onFollowChange?.(false)
  }

  /** Nearest creature to a screen point, measured from the middle of its body. */
  private creatureAt(sx: number, sy: number): Creature | null {
    const p = this.screenToTile(sx, sy)
    const radius = Math.max(0.6, 10 / this.view().scale)
    let best: Creature | null = null
    let bestDist = radius
    for (const c of this.creatures.values()) {
      const d = Math.hypot(c.rx - p.x, c.ry - 0.45 * creatureScale(c.size, c.flags) - p.y)
      if (d < bestDist) {
        best = c
        bestDist = d
      }
    }
    return best
  }

  /** Front-most building whose art covers a screen point. */
  private structureAt(sx: number, sy: number): StructureFrame | null {
    const p = this.screenToTile(sx, sy)
    let best: StructureFrame | null = null
    // Art reaches up to STRUCTURE_REACH tiles above its own tile and a bit to the sides.
    for (let ty = Math.floor(p.y); ty <= Math.floor(p.y + STRUCTURE_REACH); ty++) {
      for (let tx = Math.floor(p.x) - 1; tx <= Math.floor(p.x) + 1; tx++) {
        if (!this.inBounds(tx, ty)) continue
        const st = this.structureByTile.get(ty * this.width + tx)
        if (!st) continue
        const flat = FLAT_STRUCTURES.has(st.kind)
        const pad = flat ? 0 : 0.3
        const top = flat ? ty : ty + 1 - STRUCTURE_REACH
        if (p.x < tx - pad || p.x > tx + 1 + pad || p.y < top || p.y > ty + 1) continue
        if (!best || st.y > best.y) best = st
      }
    }
    return best
  }

  private updateHoverTargets() {
    const p = this.pointer && !this.panning ? this.pointer : null
    const c = p ? this.creatureAt(p.sx, p.sy) : null
    const st = p && !c ? this.structureAt(p.sx, p.sy) : null
    if (c === this.hoverCreature && st === this.hoverStructure) return
    this.hoverCreature = c
    this.hoverStructure = st
    this.canvas.style.cursor = c || st?.ownerId ? 'pointer' : 'default'
  }

  /** Moves along one axis; on collision, slides flush against the blocking tile. */
  private moveAxis(dx: number, dy: number) {
    const p = this.player
    const nx = p.x + dx
    const ny = p.y + dy
    if (!this.boxBlocked(nx, ny)) {
      p.x = nx
      p.y = ny
      return
    }
    if (dx > 0) p.x = Math.floor(nx + HALF_W) - HALF_W - EPS
    else if (dx < 0) p.x = Math.floor(nx - HALF_W) + 1 + HALF_W + EPS
    if (dy > 0) p.y = Math.floor(ny + HALF_H) - HALF_H - EPS
    else if (dy < 0) p.y = Math.floor(ny - HALF_H) + 1 + HALF_H + EPS
  }

  private clampCamera() {
    const { w, h, scale } = this.view()
    const halfW = w / 2 / scale
    const halfH = h / 2 / scale
    const clamp = (v: number, half: number, size: number) =>
      size <= half * 2 ? size / 2 : Math.min(size - half, Math.max(half, v))
    this.camera.x = clamp(this.camera.x, halfW, this.width)
    this.camera.y = clamp(this.camera.y, halfH, this.height)
  }

  /** Viewport in CSS pixels; `left`/`top` are in tiles. */
  private view(): View {
    const w = this.canvas.clientWidth
    const h = this.canvas.clientHeight
    const scale = TILE * ZOOMS[this.camera.zoomIndex]
    return { w, h, scale, left: this.camera.x - w / 2 / scale, top: this.camera.y - h / 2 / scale }
  }

  private screenToTile(sx: number, sy: number) {
    const v = this.view()
    return { x: v.left + sx / v.scale, y: v.top + sy / v.scale }
  }

  // --- Rendering -------------------------------------------------------------

  private render(time: number) {
    const { canvas, ctx } = this
    const dpr = window.devicePixelRatio || 1
    const cw = Math.round(canvas.clientWidth * dpr)
    const ch = Math.round(canvas.clientHeight * dpr)
    if (canvas.width !== cw || canvas.height !== ch) {
      canvas.width = cw
      canvas.height = ch
    }

    const v = this.view()
    const zoom = ZOOMS[this.camera.zoomIndex]
    const k = dpr * zoom // world pixels -> device pixels
    const originX = Math.round(-v.left * v.scale * dpr)
    const originY = Math.round(-v.top * v.scale * dpr)

    ctx.setTransform(1, 0, 0, 1, 0, 0)
    ctx.fillStyle = '#10213a'
    ctx.fillRect(0, 0, cw, ch)

    const tx0 = Math.max(0, Math.floor(v.left))
    const ty0 = Math.max(0, Math.floor(v.top))
    const tx1 = Math.min(this.width - 1, Math.floor(v.left + v.w / v.scale))
    const ty1 = Math.min(this.height - 1, Math.floor(v.top + v.h / v.scale))

    // Ground: cached chunks, placed on whole device pixels so neighbours never leave seams.
    ctx.imageSmoothingEnabled = false
    const chunkPx = CHUNK * TILE * k
    for (let cy = Math.floor(ty0 / CHUNK); cy <= Math.floor(ty1 / CHUNK); cy++) {
      for (let cx = Math.floor(tx0 / CHUNK); cx <= Math.floor(tx1 / CHUNK); cx++) {
        const x0 = Math.round(originX + cx * chunkPx)
        const y0 = Math.round(originY + cy * chunkPx)
        const x1 = Math.round(originX + (cx + 1) * chunkPx)
        const y1 = Math.round(originY + (cy + 1) * chunkPx)
        ctx.drawImage(this.getChunk(cx, cy), x0, y0, x1 - x0, y1 - y0)
      }
    }
    this.evictChunks()

    ctx.setTransform(k, 0, 0, k, originX, originY)
    if (v.scale >= 16) this.drawWaterShimmer(tx0, ty0, tx1, ty1, time)
    if (this.geologyOverlay) this.drawGeologyOverlay()
    this.drawFlatStructures(tx0, ty0, tx1, ty1)

    ctx.imageSmoothingEnabled = true
    // Buildings reach up to two tiles above their own tile, so look a little below the view.
    this.drawObjectsAndActors(tx0, Math.max(0, ty0 - 1), tx1, Math.min(this.height - 1, ty1 + 2), time)

    if (this.mode === 'edit') this.drawEditOverlay(tx0, ty0, tx1, ty1, k)
    if (this.mode === 'watch') this.drawLabels(dpr, v, zoom)
    if (this.geologyOverlay) this.drawFeatureLabels(dpr, v)
    this.drawMinimap(dpr)
  }

  private getChunk(cx: number, cy: number) {
    const id = cy * 1024 + cx
    let chunk = this.chunks.get(id)
    if (!chunk) {
      const canvas = document.createElement('canvas')
      canvas.width = canvas.height = CHUNK * TILE
      const ctx = canvas.getContext('2d')!
      for (let y = 0; y < CHUNK; y++) {
        for (let x = 0; x < CHUNK; x++) {
          const tx = cx * CHUNK + x
          const ty = cy * CHUNK + y
          if (this.inBounds(tx, ty)) this.paintTile(ctx, tx, ty, x * TILE, y * TILE)
        }
      }
      chunk = { canvas, lastUsed: 0 }
      this.chunks.set(id, chunk)
    }
    chunk.lastUsed = this.frame
    return chunk.canvas
  }

  private evictChunks() {
    if (this.chunks.size <= MAX_CHUNKS) return
    const old = [...this.chunks].filter(([, c]) => c.lastUsed < this.frame).sort((a, b) => a[1].lastUsed - b[1].lastUsed)
    for (const [id] of old.slice(0, this.chunks.size - MAX_CHUNKS)) this.chunks.delete(id)
  }

  private paintTile(ctx: CanvasRenderingContext2D, tx: number, ty: number, px: number, py: number) {
    const i = ty * this.width + tx
    const ground = this.groundDefs[this.ground[i]]
    drawGround(ctx, this.keyAt, tx, ty, px, py, ground?.color)
    const rock = this.rocks ? this.rockTypes[this.rocks[i]] : undefined
    if (rock && rock.id !== 0 && ground && ROCKY_GROUND.has(ground.key)) tintLithology(ctx, rock.color, px, py)
    for (const d of this.deposits.get(i) ?? []) {
      drawDeposit(ctx, d.item.id, d.item.elements, d.surface, tx, ty, px, py, d.model?.key)
    }
    const obj = this.objectDefs[this.objects[i]]
    if (obj && FLAT_OBJECTS.has(obj.key)) drawFlat(ctx, obj.key, tx, ty, px, py)
  }

  /** Repaints a tile and its neighbours (shorelines and cliffs depend on neighbours). */
  private repaintAround(x: number, y: number) {
    for (let ty = y - 1; ty <= y + 1; ty++) {
      for (let tx = x - 1; tx <= x + 1; tx++) {
        if (!this.inBounds(tx, ty)) continue
        const chunk = this.chunks.get(Math.floor(ty / CHUNK) * 1024 + Math.floor(tx / CHUNK))
        if (!chunk) continue
        this.paintTile(chunk.canvas.getContext('2d')!, tx, ty, (tx % CHUNK) * TILE, (ty % CHUNK) * TILE)
      }
    }
  }

  private drawWaterShimmer(tx0: number, ty0: number, tx1: number, ty1: number, time: number) {
    const ctx = this.ctx
    ctx.fillStyle = '#ffffff'
    for (let ty = ty0; ty <= ty1; ty++) {
      for (let tx = tx0; tx <= tx1; tx++) {
        if (!isWater(this.keyAt(tx, ty))) continue
        for (let j = 0; j < 2; j++) {
          const phase = (time * 0.3 + hash(tx, ty, 200 + j)) % 1
          ctx.globalAlpha = Math.sin(phase * Math.PI) * 0.25
          ctx.fillRect(tx * TILE + 2 + phase * (TILE - 12), ty * TILE + 5 + hash(tx, ty, 210 + j) * (TILE - 10), 8, 1.5)
        }
      }
    }
    ctx.globalAlpha = 1
  }

  private drawFlatStructures(tx0: number, ty0: number, tx1: number, ty1: number) {
    for (const st of this.flatStructures) {
      if (st.x < tx0 - 1 || st.x > tx1 + 1 || st.y < ty0 - 1 || st.y > ty1 + 1) continue
      drawFarm(this.ctx, st.x * TILE, st.y * TILE, st.hue, !st.ownerId, st.x, st.y)
      if (st === this.hoverStructure) this.outlineTile(st, 'rgba(255,255,255,0.8)')
    }
  }

  private outlineTile(st: StructureFrame, color: string) {
    const ctx = this.ctx
    ctx.strokeStyle = color
    ctx.lineWidth = 1.5
    ctx.strokeRect(st.x * TILE + 1, st.y * TILE + 1, TILE - 2, TILE - 2)
  }

  /** Draws upright objects row by row so the player and creatures are depth-sorted between them. */
  private drawObjectsAndActors(tx0: number, ty0: number, tx1: number, ty1: number, time: number) {
    const ctx = this.ctx
    const p = this.player
    let playerDrawn = this.mode !== 'play'
    const drawP = () => {
      drawPlayer(ctx, p.x * TILE, p.y * TILE, p.facing, p.step, p.moving)
      playerDrawn = true
    }

    const actors =
      this.mode === 'watch'
        ? [...this.creatures.values()]
            .filter((c) => c.rx >= tx0 - 1 && c.rx <= tx1 + 2 && c.ry >= ty0 && c.ry <= ty1 + 3)
            .sort((a, b) => a.ry - b.ry)
        : []
    let next = 0
    const drawActorsBefore = (limit: number) => {
      for (; next < actors.length && actors[next].ry < limit; next++) this.drawOneCreature(actors[next], time)
    }

    const selectedHouse = this.selectedId !== null ? (this.creatures.get(this.selectedId)?.houseId ?? 0) : 0

    // Sprites overhang half a tile sideways, so include one extra column each side.
    const sx0 = Math.max(0, tx0 - 1)
    const sx1 = Math.min(this.width - 1, tx1 + 1)
    for (let ty = ty0; ty <= ty1; ty++) {
      if (!playerDrawn && p.y < ty + 0.8) drawP()
      drawActorsBefore(ty + 0.8)
      for (let tx = sx0; tx <= sx1; tx++) {
        const id = this.objects[ty * this.width + tx]
        if (id === 0) continue
        const def = this.objectDefs[id]
        if (!def || FLAT_OBJECTS.has(def.key)) continue
        const variants = this.sprites.get(def.key)
        if (variants) {
          const sprite = variants[Math.floor(hash(tx, ty, 7) * VARIANTS)]
          ctx.drawImage(sprite, tx * TILE - TILE / 2, ty * TILE - TILE, SPRITE_SIZE, SPRITE_SIZE)
        } else {
          drawFallbackObject(ctx, def.color, tx * TILE, ty * TILE)
        }
      }
      for (const st of this.structureRows.get(ty) ?? []) {
        if (st.x < sx0 || st.x > sx1) continue
        this.drawStructure(st, selectedHouse)
      }
    }
    if (!playerDrawn) drawP()
    drawActorsBefore(Infinity)
  }

  private drawStructure(st: StructureFrame, selectedHouse: number) {
    const ctx = this.ctx
    const cx = st.x * TILE + TILE / 2
    const base = (st.y + 1) * TILE - 3
    // The observed creature's home gets a gold ring at its foot; the hovered building a white one.
    if (st.id === selectedHouse || st === this.hoverStructure) {
      ctx.strokeStyle = st.id === selectedHouse ? '#ffd166' : 'rgba(255,255,255,0.75)'
      ctx.lineWidth = 1.5
      ctx.setLineDash(st.id === selectedHouse ? [4, 3] : [])
      ctx.beginPath()
      ctx.ellipse(cx, base, 24, 7, 0, 0, Math.PI * 2)
      ctx.stroke()
      ctx.setLineDash([])
    }
    const sprite = structureSprite(st.kind, st.level, st.hue, !st.ownerId)
    ctx.drawImage(sprite, st.x * TILE - TILE / 2, (st.y + 1) * TILE - STRUCTURE_H, STRUCTURE_W, STRUCTURE_H)
  }

  private drawOneCreature(c: Creature, time: number) {
    const ctx = this.ctx
    const s = creatureScale(c.size, c.flags)
    if (c.id === this.selectedId || c === this.hoverCreature) {
      ctx.strokeStyle = c.id === this.selectedId ? '#ffd166' : 'rgba(255,255,255,0.7)'
      ctx.lineWidth = 1.5
      ctx.beginPath()
      ctx.ellipse(c.rx * TILE, c.ry * TILE, 11 * s, 5 * s, 0, 0, Math.PI * 2)
      ctx.stroke()
    }
    drawCreature(ctx, c.rx * TILE, c.ry * TILE, {
      heading: c.rh,
      sex: c.sex,
      hue: c.hue,
      size: c.size,
      flags: c.flags,
      energy: c.energy,
      health: c.health,
      step: c.step,
      moving: c.moving,
      variant: c.id,
      time,
    })
  }

  private buildOverlay() {
    const rocks = this.rocks
    if (!rocks) return null
    const { width: w, height: h } = this
    const R = OVERLAY_RES
    const canvas = document.createElement('canvas')
    canvas.width = w * R
    canvas.height = h * R
    const ctx = canvas.getContext('2d')!
    const img = ctx.createImageData(w * R, h * R)
    const rgb = this.rockTypes.map((t) => hexToRgb(t.color))
    for (let ty = 0; ty < h; ty++) {
      for (let tx = 0; tx < w; tx++) {
        const r = rocks[ty * w + tx]
        if (!r || !rgb[r]) continue
        const [cr, cg, cb] = rgb[r]
        // Unit boundaries are drawn darker on the tile's right and bottom edges.
        const east = tx + 1 < w && rocks[ty * w + tx + 1] !== r
        const south = ty + 1 < h && rocks[(ty + 1) * w + tx] !== r
        for (let y = 0; y < R; y++) {
          for (let x = 0; x < R; x++) {
            const edge = (east && x === R - 1) || (south && y === R - 1)
            const o = ((ty * R + y) * w * R + tx * R + x) * 4
            const f = edge ? 0.45 : 1
            img.data[o] = cr * f
            img.data[o + 1] = cg * f
            img.data[o + 2] = cb * f
            img.data[o + 3] = 255
          }
        }
      }
    }
    ctx.putImageData(img, 0, 0)
    return canvas
  }

  private drawGeologyOverlay() {
    this.overlayCanvas ??= this.buildOverlay()
    if (!this.overlayCanvas) return
    const ctx = this.ctx
    ctx.imageSmoothingEnabled = false
    ctx.globalAlpha = 0.55
    ctx.drawImage(this.overlayCanvas, 0, 0, this.width * TILE, this.height * TILE)
    ctx.globalAlpha = 1
  }

  /** Names of volcanoes, plutons, rivers… on the geological map, in screen space. */
  private drawFeatureLabels(dpr: number, v: View) {
    if (!this.features.length) return
    const ctx = this.ctx
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.font = '600 11px Inter, system-ui, sans-serif'
    ctx.textAlign = 'left'
    ctx.textBaseline = 'middle'
    for (const f of this.features) {
      const x = (f.x + 0.5 - v.left) * v.scale
      const y = (f.y + 0.5 - v.top) * v.scale
      if (x < -150 || y < -20 || x > v.w + 150 || y > v.h + 20) continue
      const text = `${FEATURE_ICONS[f.kind] ?? '•'} ${f.name}`
      const w = ctx.measureText(text).width + 12
      ctx.fillStyle = 'rgba(10,18,30,0.78)'
      ctx.beginPath()
      ctx.roundRect(x - w / 2, y - 9, w, 18, 5)
      ctx.fill()
      ctx.fillStyle = '#f3ead2'
      ctx.fillText(text, x - w / 2 + 6, y + 0.5)
    }
  }

  /** Name tags for the hovered and selected creatures and the hovered building, in screen space. */
  private drawLabels(dpr: number, v: View, zoom: number) {
    const ctx = this.ctx
    const shown = new Set<Creature>()
    if (this.hoverCreature) shown.add(this.hoverCreature)
    const selected = this.selectedId !== null ? this.creatures.get(this.selectedId) : undefined
    if (selected) shown.add(selected)
    const st = this.hoverStructure
    if (!shown.size && !st) return

    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.font = '600 12px Inter, system-ui, sans-serif'
    ctx.textAlign = 'left'
    ctx.textBaseline = 'middle'

    if (st) {
      const text = structureLabel(st.kind, st.level, st.ownerName, st.ownerId)
      const flat = FLAT_STRUCTURES.has(st.kind)
      const x = (st.x + 0.5 - v.left) * v.scale
      const y = (st.y + 1 - (flat ? 1 : STRUCTURE_REACH) - v.top) * v.scale - 10
      const w = ctx.measureText(text).width + 14
      ctx.fillStyle = 'rgba(10,18,30,0.85)'
      ctx.beginPath()
      ctx.roundRect(x - w / 2, y - 10, w, 20, 6)
      ctx.fill()
      ctx.fillStyle = st.ownerId ? `hsl(${st.hue} 80% 75%)` : '#b8c0cc'
      ctx.fillText(text, x - w / 2 + 7, y + 0.5)
    }

    for (const c of shown) {
      const head = 40 * creatureScale(c.size, c.flags) * zoom
      const x = (c.rx - v.left) * v.scale
      const y = (c.ry - v.top) * v.scale - head - 10
      const symbol = SEX_SYMBOL[c.sex]
      const text = `${symbol} ${c.name}${c.flags & FLAG.head ? ' 👑' : ''}`
      const w = ctx.measureText(text).width + 14
      ctx.fillStyle = c === selected ? 'rgba(40,32,8,0.88)' : 'rgba(10,18,30,0.82)'
      ctx.beginPath()
      ctx.roundRect(x - w / 2, y - 10, w, 20, 6)
      ctx.fill()
      if (c === selected) {
        ctx.strokeStyle = '#ffd166'
        ctx.lineWidth = 1
        ctx.stroke()
      }
      ctx.fillStyle = c.sex === 'female' ? '#ff9fcf' : '#8fc4ff'
      ctx.fillText(symbol, x - w / 2 + 7, y + 0.5)
      ctx.fillStyle = '#fff'
      ctx.fillText(text.slice(symbol.length + 1), x - w / 2 + 7 + ctx.measureText(symbol + ' ').width, y + 0.5)
    }
  }

  private drawEditOverlay(tx0: number, ty0: number, tx1: number, ty1: number, k: number) {
    const ctx = this.ctx
    if (TILE * ZOOMS[this.camera.zoomIndex] >= 16) {
      ctx.strokeStyle = 'rgba(0,0,0,0.18)'
      ctx.lineWidth = 1 / k
      ctx.beginPath()
      for (let x = tx0; x <= tx1 + 1; x++) {
        ctx.moveTo(x * TILE, ty0 * TILE)
        ctx.lineTo(x * TILE, (ty1 + 1) * TILE)
      }
      for (let y = ty0; y <= ty1 + 1; y++) {
        ctx.moveTo(tx0 * TILE, y * TILE)
        ctx.lineTo((tx1 + 1) * TILE, y * TILE)
      }
      ctx.stroke()
    }
    ctx.strokeStyle = 'rgba(255,255,255,0.5)'
    ctx.lineWidth = 2 / k
    ctx.strokeRect(0, 0, this.width * TILE, this.height * TILE)

    drawSpawnFlag(ctx, this.spawn.x, this.spawn.y)

    const hover = this.hoverTile()
    if (hover) {
      const [a, b] = this.footprint()
      const x = (hover.x + a) * TILE
      const y = (hover.y + a) * TILE
      const size = (b - a + 1) * TILE
      ctx.fillStyle = 'rgba(255,255,255,0.18)'
      ctx.fillRect(x, y, size, size)
      ctx.strokeStyle = '#fff'
      ctx.lineWidth = 2 / k
      ctx.strokeRect(x, y, size, size)
    }
  }

  private renderMinimapBase() {
    const c = this.minimapBase
    c.width = this.width
    c.height = this.height
    const ctx = c.getContext('2d')!
    const img = ctx.createImageData(this.width, this.height)
    for (let i = 0; i < this.ground.length; i++) {
      const [r, g, b] = this.minimapColor(i)
      img.data.set([r, g, b, 255], i * 4)
    }
    ctx.putImageData(img, 0, 0)
  }

  private minimapColor(i: number) {
    const obj = this.objectDefs[this.objects[i]]
    if (obj && obj.id !== 0 && !FLAT_OBJECTS.has(obj.key)) return this.objectRgb[obj.id]
    return this.groundRgb[this.ground[i]] ?? [255, 0, 255]
  }

  private minimapLayout() {
    const mm = this.minimap!
    const s = Math.min(mm.clientWidth / this.width, mm.clientHeight / this.height)
    return { s, ox: (mm.clientWidth - this.width * s) / 2, oy: (mm.clientHeight - this.height * s) / 2 }
  }

  private drawMinimap(dpr: number) {
    const mm = this.minimap
    if (!mm) return
    const w = Math.round(mm.clientWidth * dpr)
    const h = Math.round(mm.clientHeight * dpr)
    if (mm.width !== w || mm.height !== h) {
      mm.width = w
      mm.height = h
    }
    const ctx = mm.getContext('2d')!
    const { s, ox, oy } = this.minimapLayout()
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, mm.clientWidth, mm.clientHeight)
    ctx.imageSmoothingEnabled = false
    ctx.drawImage(this.minimapBase, ox, oy, this.width * s, this.height * s)
    if (this.geologyOverlay) {
      this.overlayCanvas ??= this.buildOverlay()
      if (this.overlayCanvas) {
        ctx.globalAlpha = 0.7
        ctx.drawImage(this.overlayCanvas, ox, oy, this.width * s, this.height * s)
        ctx.globalAlpha = 1
      }
    }

    const v = this.view()
    ctx.strokeStyle = 'rgba(255,255,255,0.9)'
    ctx.lineWidth = 1
    ctx.strokeRect(ox + v.left * s, oy + v.top * s, (v.w / v.scale) * s, (v.h / v.scale) * s)

    const dot = (x: number, y: number, color: string) => {
      ctx.fillStyle = color
      ctx.beginPath()
      ctx.arc(ox + x * s, oy + y * s, 3, 0, Math.PI * 2)
      ctx.fill()
    }
    if (this.mode === 'watch') {
      // Houses in their family's colour, other buildings white, abandoned ones grey.
      for (const st of this.structures) {
        ctx.fillStyle = !st.ownerId ? '#8a8f98' : st.level > 0 ? `hsl(${st.hue} 75% 60%)` : '#ffffff'
        ctx.fillRect(ox + (st.x + 0.5) * s - 1.5, oy + (st.y + 0.5) * s - 1.5, 3, 3)
      }
      for (const c of this.creatures.values()) {
        ctx.fillStyle = c.sex === 'female' ? '#ff8fc7' : '#6fb6ff'
        ctx.fillRect(ox + c.rx * s - 1, oy + c.ry * s - 1, 2, 2)
      }
      const sel = this.selectedId !== null ? this.creatures.get(this.selectedId) : undefined
      if (sel) {
        ctx.strokeStyle = '#ffd166'
        ctx.lineWidth = 1.5
        ctx.beginPath()
        ctx.arc(ox + sel.rx * s, oy + sel.ry * s, 4, 0, Math.PI * 2)
        ctx.stroke()
      }
      return
    }
    dot(this.spawn.x + 0.5, this.spawn.y + 0.5, '#e63946')
    if (this.mode === 'play') dot(this.player.x, this.player.y, '#ffd166')
  }

  // --- Editing ---------------------------------------------------------------

  private hoverTile(): Point | null {
    if (!this.pointer) return null
    const t = this.screenToTile(this.pointer.sx, this.pointer.sy)
    const p = { x: Math.floor(t.x), y: Math.floor(t.y) }
    return this.inBounds(p.x, p.y) ? p : null
  }

  private updateHover() {
    const t = this.hoverTile()
    const i = t ? t.y * this.width + t.x : -1
    const key = t ? `${t.x},${t.y},${this.ground[i]},${this.objects[i]},${this.geologyVersion}` : ''
    if (key === this.hoverKey) return
    this.hoverKey = key
    this.events.onHover?.(
      t
        ? {
            x: t.x,
            y: t.y,
            ground: this.groundDefs[this.ground[i]],
            object: this.objectDefs[this.objects[i]],
            deposits: (this.deposits.get(i) ?? []).map(({ item, model }) => ({ item, model })),
            rock: this.rocks ? this.rockTypes[this.rocks[i]] : undefined,
          }
        : null,
    )
  }

  /** Offsets of the brush square relative to the hovered tile. */
  private footprint(): [number, number] {
    if (this.tool.kind === 'spawn' || this.brush === 'fill') return [0, 0]
    const r = (this.brush - 1) / 2
    return [-r, r]
  }

  private paintAt(tx: number, ty: number) {
    const tool = this.tool
    if (tool.kind === 'spawn') {
      if (this.solidAt(tx, ty)) return
      const before = this.stroke!.spawn?.before ?? { ...this.spawn }
      this.spawn = { x: tx, y: ty }
      this.stroke!.spawn = { before, after: { ...this.spawn } }
      return
    }
    const [a, b] = this.footprint()
    for (let y = ty + a; y <= ty + b; y++) {
      for (let x = tx + a; x <= tx + b; x++) {
        if (this.setTile(x, y, tool.layer, tool.id)) this.tileChanged(x, y)
      }
    }
  }

  /** Sets a tile and records it in the current stroke. Returns whether it changed. */
  private setTile(x: number, y: number, layer: Layer, id: number) {
    if (!this.inBounds(x, y)) return false
    const i = y * this.width + x
    const arr = layer === 'ground' ? this.ground : this.objects
    const before = arr[i]
    if (before === id) return false
    // The spawn tile must stay walkable.
    const defs = layer === 'ground' ? this.groundDefs : this.objectDefs
    if (x === this.spawn.x && y === this.spawn.y && defs[id]?.solid) return false
    arr[i] = id
    this.stroke!.changes.push({ i, layer, before, after: id })
    return true
  }

  private tileChanged(x: number, y: number) {
    this.repaintAround(x, y)
    const ctx = this.minimapBase.getContext('2d')!
    const [r, g, b] = this.minimapColor(y * this.width + x)
    ctx.fillStyle = `rgb(${r},${g},${b})`
    ctx.fillRect(x, y, 1, 1)
  }

  private bulkChanged(changes: Change[]) {
    if (changes.length > BULK_REDRAW) {
      this.chunks.clear()
      this.renderMinimapBase()
      return
    }
    for (const c of changes) this.tileChanged(c.i % this.width, Math.floor(c.i / this.width))
  }

  private floodFill(tx: number, ty: number) {
    if (this.tool.kind !== 'paint') return
    const { layer, id } = this.tool
    const arr = layer === 'ground' ? this.ground : this.objects
    const start = ty * this.width + tx
    const target = arr[start]
    if (target === id) return

    const seen = new Uint8Array(arr.length)
    const stack = [start]
    const from = this.stroke!.changes.length
    let filled = 0
    while (stack.length && filled < MAX_FILL) {
      const i = stack.pop()!
      if (seen[i] || arr[i] !== target) continue
      seen[i] = 1
      const x = i % this.width
      const y = Math.floor(i / this.width)
      this.setTile(x, y, layer, id)
      filled++
      if (x > 0) stack.push(i - 1)
      if (x < this.width - 1) stack.push(i + 1)
      if (y > 0) stack.push(i - this.width)
      if (y < this.height - 1) stack.push(i + this.width)
    }
    this.bulkChanged(this.stroke!.changes.slice(from))
  }

  private beginStroke() {
    this.endStroke()
    this.stroke = { changes: [] }
  }

  private endStroke() {
    const s = this.stroke
    this.stroke = null
    this.painting = false
    this.lastPaint = null
    if (!s || (s.changes.length === 0 && !s.spawn)) return
    this.undoStack.push(s)
    if (this.undoStack.length > MAX_UNDO) this.undoStack.shift()
    this.redoStack = []
    this.emitState()
  }

  private applyStroke(s: Stroke, reverse: boolean) {
    const changes = reverse ? [...s.changes].reverse() : s.changes
    for (const c of changes) {
      const arr = c.layer === 'ground' ? this.ground : this.objects
      arr[c.i] = reverse ? c.before : c.after
    }
    this.bulkChanged(s.changes)
    if (s.spawn) this.spawn = { ...(reverse ? s.spawn.before : s.spawn.after) }
  }

  private emitState() {
    const dirty = this.isDirty()
    if (dirty !== this.lastDirty) {
      this.lastDirty = dirty
      this.events.onDirtyChange?.(dirty)
    }
    this.events.onHistoryChange?.(this.undoStack.length > 0, this.redoStack.length > 0)
  }

  // --- Input -----------------------------------------------------------------

  private onKeyDown = (e: KeyboardEvent) => {
    if (isTyping(e.target)) return
    const mod = e.ctrlKey || e.metaKey
    if (mod && this.mode === 'edit' && (e.code === 'KeyZ' || e.code === 'KeyY')) {
      e.preventDefault()
      if (e.code === 'KeyY' || e.shiftKey) this.redo()
      else this.undo()
      return
    }
    if (mod || e.altKey) return
    if (e.code === 'Equal' || e.code === 'NumpadAdd') this.zoomBy(1)
    if (e.code === 'Minus' || e.code === 'NumpadSubtract') this.zoomBy(-1)
    if (CAPTURED_KEYS.has(e.code)) e.preventDefault()
    this.keys.add(e.code)
  }

  private onKeyUp = (e: KeyboardEvent) => {
    this.keys.delete(e.code)
  }

  private onBlur = () => {
    this.keys.clear()
    this.endStroke()
    this.panning = null
    this.press = null
  }

  private setPointer(e: MouseEvent) {
    const r = this.canvas.getBoundingClientRect()
    this.pointer = { sx: e.clientX - r.left, sy: e.clientY - r.top }
  }

  private onPointerDown = (e: PointerEvent) => {
    this.setPointer(e)
    if (this.mode === 'play') return
    this.canvas.setPointerCapture(e.pointerId)
    if (e.button === 1 || e.button === 2) {
      this.panning = { sx: e.clientX, sy: e.clientY, cx: this.camera.x, cy: this.camera.y }
      this.stopFollow()
      return
    }
    if (e.button !== 0) return
    if (this.mode === 'watch') {
      // Decide on pointer-up whether this was a click (select) or a drag (pan).
      this.press = { sx: e.clientX, sy: e.clientY, cx: this.camera.x, cy: this.camera.y }
      return
    }
    const t = this.hoverTile()
    if (!t) return
    this.beginStroke()
    if (this.tool.kind === 'paint' && this.brush === 'fill') {
      this.floodFill(t.x, t.y)
      this.endStroke()
      return
    }
    this.painting = true
    this.lastPaint = t
    this.paintAt(t.x, t.y)
  }

  private onPointerMove = (e: PointerEvent) => {
    this.setPointer(e)
    if (this.press && Math.hypot(e.clientX - this.press.sx, e.clientY - this.press.sy) > DRAG_THRESHOLD) {
      this.panning = this.press
      this.press = null
      this.stopFollow()
    }
    if (this.panning) {
      const scale = TILE * ZOOMS[this.camera.zoomIndex]
      this.camera.x = this.panning.cx - (e.clientX - this.panning.sx) / scale
      this.camera.y = this.panning.cy - (e.clientY - this.panning.sy) / scale
      return
    }
    if (!this.painting || this.tool.kind === 'spawn') return
    const t = this.hoverTile()
    if (!t || !this.lastPaint) return
    // Paint every tile along the drag so fast mouse moves leave no gaps.
    const from = this.lastPaint
    const steps = Math.max(Math.abs(t.x - from.x), Math.abs(t.y - from.y))
    for (let s = 1; s <= steps; s++) {
      this.paintAt(Math.round(from.x + ((t.x - from.x) * s) / steps), Math.round(from.y + ((t.y - from.y) * s) / steps))
    }
    this.lastPaint = t
  }

  private onPointerUp = () => {
    if (this.press && this.pointer) {
      const { sx, sy } = this.pointer
      // A creature wins over the building behind it; a building selects its owner.
      const id = this.creatureAt(sx, sy)?.id ?? (this.structureAt(sx, sy)?.ownerId || null)
      this.selectedId = id
      this.events.onSelectCreature?.(id)
    }
    this.press = null
    this.panning = null
    if (this.painting) this.endStroke()
  }

  private onPointerLeave = () => {
    if (!this.painting && !this.panning && !this.press) this.pointer = null
  }

  private onWheel = (e: WheelEvent) => {
    e.preventDefault()
    this.wheelAccum += e.deltaY
    if (Math.abs(this.wheelAccum) < 50) return
    this.setPointer(e)
    this.zoomBy(this.wheelAccum < 0 ? 1 : -1, this.pointer ?? undefined)
    this.wheelAccum = 0
  }

  private onContextMenu = (e: Event) => e.preventDefault()

  private onMinimapPointer = (e: PointerEvent) => {
    if (this.mode === 'play' || !this.minimap) return
    const r = this.minimap.getBoundingClientRect()
    const { s, ox, oy } = this.minimapLayout()
    this.camera.x = (e.clientX - r.left - ox) / s
    this.camera.y = (e.clientY - r.top - oy) / s
    this.stopFollow()
  }
}
