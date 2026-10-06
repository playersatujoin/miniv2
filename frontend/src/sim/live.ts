// What the stream says is alive on the island right now, smoothed between
// frames. The page keeps a single copy and both the 2D map and the 3D view draw
// from it, so switching between them never loses anyone or moves them: they are
// the very same people and animals, mid-stride.

import type {
  AnimalFrame,
  BurntMessage,
  CreatureFrame,
  FieldPlot,
  MosquitoMessage,
  SimFrame,
  StructureFrame,
  Village,
  WaterMessage,
} from './protocol'

/** Something streamed that moves: interpolated from (px, py, ph) to (x, y, heading) between frames. */
export type Glide = {
  px: number
  py: number
  ph: number
  x: number
  y: number
  heading: number
  /** Where it is drawn right now (tiles, radians). */
  rx: number
  ry: number
  rh: number
  /** Tiles walked so far; each view turns this into leg swings at its own stride. */
  walked: number
  moving: boolean
}

export type LiveCreature = Omit<CreatureFrame, 'x' | 'y' | 'heading'> & Glide
export type LiveAnimal = Omit<AnimalFrame, 'x' | 'y' | 'heading'> & Glide

/** Tiles; anything that jumps further between frames is teleported, not slid across the map. */
const SNAP_DISTANCE = 10

const still = (x: number, y: number, heading: number): Glide => ({
  px: x,
  py: y,
  ph: heading,
  x,
  y,
  heading,
  rx: x,
  ry: y,
  rh: heading,
  walked: 0,
  moving: false,
})

/** Points a glide at its next position, starting from where it is drawn right now. */
function retarget(g: Glide, x: number, y: number, heading: number) {
  g.px = g.rx
  g.py = g.ry
  g.ph = g.rh
  g.x = x
  g.y = y
  g.heading = heading
  if (Math.hypot(g.x - g.px, g.y - g.py) > SNAP_DISTANCE) {
    g.px = g.rx = g.x
    g.py = g.ry = g.y
  }
}

/** Moves a glide fraction t of the way along its segment. */
function glide(g: Glide, t: number, dt: number) {
  const rx = g.px + (g.x - g.px) * t
  const ry = g.py + (g.y - g.py) * t
  // Turn the short way round.
  let dh = (g.heading - g.ph) % (Math.PI * 2)
  if (dh > Math.PI) dh -= Math.PI * 2
  else if (dh < -Math.PI) dh += Math.PI * 2
  g.rh = g.ph + dh * t
  const moved = Math.hypot(rx - g.rx, ry - g.ry)
  g.moving = dt > 0 && moved / dt > 0.15
  g.walked += moved
  g.rx = rx
  g.ry = ry
}

export class LiveWorld {
  creatures = new Map<number, LiveCreature>()
  animals = new Map<number, LiveAnimal>()
  /** The latest frame as streamed, and buildings and fields (null until they arrive). */
  frame: SimFrame | null = null
  structures: StructureFrame[] | null = null
  plots: FieldPlot[] | null = null
  water: WaterMessage | null = null
  mosquitoes: MosquitoMessage | null = null
  /** Villages and scorched land (null until they arrive, or from servers without them). Bodies and fires come in each frame. */
  villages: Village[] | null = null
  burnt: BurntMessage | null = null
  /** When the latest frame arrived (performance.now()). */
  frameAt = 0
  /** Smoothed time between frames, ms. */
  private frameInterval = 100
  private advancedAt = 0

  /**
   * Takes a new frame: everyone glides from where they are drawn now towards it until the next one.
   * `interval` (ms) says when the next frame is due, for a feed with its own pace (the replay pushes
   * recorded frames every half second or so, scaled by its speed); the live stream's pace is measured.
   */
  push(frame: SimFrame, now = performance.now(), interval?: number) {
    if (interval !== undefined && interval > 0) {
      this.frameInterval = Math.min(4000, Math.max(16, interval))
    } else if (this.frameAt) {
      const gap = Math.min(250, Math.max(50, now - this.frameAt))
      this.frameInterval = this.frameInterval * 0.8 + gap * 0.2
    }
    this.frameAt = now
    this.frame = frame

    const people = new Map<number, LiveCreature>()
    for (const f of frame.creatures) {
      const c = this.creatures.get(f.id) ?? { ...f, ...still(f.x, f.y, f.heading) }
      c.name = f.name
      c.sex = f.sex
      c.hue = f.hue
      c.size = f.size
      c.flags = f.flags
      c.energy = f.energy
      c.health = f.health
      c.houseId = f.houseId
      c.age = f.age
      c.lookX = f.lookX
      c.lookY = f.lookY
      c.mood = f.mood
      c.moodStrength = f.moodStrength
      retarget(c, f.x, f.y, f.heading)
      people.set(f.id, c)
    }
    this.creatures = people

    const herd = new Map<number, LiveAnimal>()
    for (const f of frame.animals) {
      const a = this.animals.get(f.id) ?? { id: f.id, species: f.species, flags: f.flags, ...still(f.x, f.y, f.heading) }
      a.species = f.species
      a.flags = f.flags
      retarget(a, f.x, f.y, f.heading)
      herd.set(f.id, a)
    }
    this.animals = herd
  }

  /** Moves everyone along to `now` (ms, performance.now() time); asking again for the same moment does nothing. */
  advance(now: number) {
    if (now <= this.advancedAt) return
    const dt = this.advancedAt ? Math.min(0.1, (now - this.advancedAt) / 1000) : 0
    this.advancedAt = now
    const t = Math.min(1, Math.max(0, (now - this.frameAt) / this.frameInterval))
    for (const c of this.creatures.values()) glide(c, t, dt)
    for (const a of this.animals.values()) glide(a, t, dt)
  }

  /**
   * Forgets where everyone was, so the next frame places them where it says instead of sliding them
   * there: after a jump in time (seeking in the replay, or back to the live stream). The pace of
   * frames is measured afresh.
   */
  forgetMotion() {
    this.creatures = new Map()
    this.animals = new Map()
    this.frameAt = 0
    this.frameInterval = 100
  }

  /** Forgets everything, e.g. when the stream stops: stale people would otherwise slide when it resumes. */
  clear() {
    this.creatures = new Map()
    this.animals = new Map()
    this.frame = null
    this.structures = null
    this.plots = null
    this.water = null
    this.mosquitoes = null
    this.villages = null
    this.burnt = null
    this.frameAt = 0
    this.frameInterval = 100
  }
}
