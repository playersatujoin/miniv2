// The automatic camera's mind, after RAGE's cinematic director
// (game/camera/cinematic/CinematicDirector.h with its contexts and shots): once
// a second it looks over what is happening near the camera, scores each
// candidate subject, and keeps the one it is showing for a minimum time unless
// something clearly better comes along (hysteresis); each subject is framed by
// a shot that suits it (behind the shoulder, orbiting, a wide establishing
// view, a two-shot of a pair talking), and recently shown subjects make way for
// new ones (the shot history). Pure (no three.js, no DOM) for node tests;
// World3D gathers the scene, moves the camera and reports the subject.

import type { DeathCause } from '../sim/protocol'

export type SubjectKind =
  'fire' | 'drowning' | 'corpse' | 'fight' | 'fall' | 'raft' | 'trade' | 'talk' | 'swim' | 'leader' | 'hurt' | 'village' | 'event'

export type Shot = 'follow' | 'orbit' | 'establishing' | 'twoShot'

/** Something worth watching. Positions in tiles (x east, z south). */
export type Subject = {
  /** Stable while it lasts, e.g. 'talk:12-40', 'fire:8:3', 'corpse:91'. */
  key: string
  kind: SubjectKind
  /** The person to report and follow (null for a fire or a village). */
  id: number | null
  /** The other one of a pair (a two-shot frames both). */
  partner: number | null
  x: number
  z: number
  /** How far it spreads (tiles): a fire front, a village. */
  radius: number
  score: number
  /** For the observer, in Indonesian. */
  label: string
}

/** The bits of CreatureFrame.flags the director reads (pass FLAG from sim/protocol). */
export type CinematicFlags = {
  attacking: number
  hurt: number
  swimming: number
  rafting: number
  fallen: number
  talking: number
  trading: number
  leader: number
}

export type ScenePerson = { id: number; rx: number; ry: number; flags: number; health: number }
export type SceneFire = { x: number; y: number; intensity: number }
export type SceneCorpse = { id: number; x: number; y: number; seconds: number; cause: DeathCause }
export type SceneVillage = {
  id: number
  name: string
  x: number
  y: number
  people: number
  hull: [number, number][]
  leader: { id: number } | null
}

export type Scene = {
  people: Iterable<ScenePerson>
  fires: readonly SceneFire[]
  corpses: readonly SceneCorpse[]
  villages: readonly SceneVillage[]
  /** Moments told from outside (e.g. a birth from the event log), already scored. */
  hints?: readonly Subject[]
}

const CORPSE_LABEL: Partial<Record<DeathCause, string>> = {
  drowned: 'Tenggelam',
  burned: 'Korban kebakaran',
  fall: 'Jatuh dari ketinggian',
  killed: 'Korban perkelahian',
  animal: 'Diterkam binatang buas',
}

/** Fires within one cell of this many tiles are one subject. */
const FIRE_CELL = 8

/**
 * Everything worth showing in a scene, scored (higher is more interesting). Fires gather into
 * fronts; people talking or trading pair with the nearest partner doing the same; a fight takes
 * in the hurt around it.
 */
export function findSubjects(scene: Scene, flags: CinematicFlags): Subject[] {
  const out: Subject[] = []
  // Fires, gathered into fronts.
  const fronts = new Map<
    string,
    { x: number; z: number; heat: number; n: number; minX: number; maxX: number; minZ: number; maxZ: number }
  >()
  for (const f of scene.fires) {
    if (!(f.intensity > 0)) continue
    const key = `${Math.floor(f.x / FIRE_CELL)}:${Math.floor(f.y / FIRE_CELL)}`
    let c = fronts.get(key)
    if (!c) fronts.set(key, (c = { x: 0, z: 0, heat: 0, n: 0, minX: f.x, maxX: f.x, minZ: f.y, maxZ: f.y }))
    c.x += (f.x + 0.5) * f.intensity
    c.z += (f.y + 0.5) * f.intensity
    c.heat += f.intensity
    c.n++
    c.minX = Math.min(c.minX, f.x)
    c.maxX = Math.max(c.maxX, f.x)
    c.minZ = Math.min(c.minZ, f.y)
    c.maxZ = Math.max(c.maxZ, f.y)
  }
  for (const [key, c] of fronts) {
    out.push({
      key: `fire:${key}`,
      kind: 'fire',
      id: null,
      partner: null,
      x: c.x / c.heat,
      z: c.z / c.heat,
      radius: Math.max(2, Math.hypot(c.maxX - c.minX, c.maxZ - c.minZ) / 2 + 1),
      score: 60 + 30 * Math.min(1, c.heat / 4),
      label: 'Kebakaran',
    })
  }

  // Bodies, the freshest first.
  for (const b of scene.corpses) {
    if (b.seconds > 30) continue
    const cause = CORPSE_LABEL[b.cause]
    out.push({
      key: `corpse:${b.id}`,
      kind: 'corpse',
      id: b.id,
      partner: null,
      x: b.x,
      z: b.y,
      radius: 1,
      score: 12 + 40 * (1 - b.seconds / 30) + (cause ? 5 : 0),
      label: cause ?? 'Seseorang meninggal',
    })
  }

  // People: fights, falls, water, pairs, leaders.
  const pairing: ScenePerson[] = []
  const hurt: ScenePerson[] = []
  const fighters: ScenePerson[] = []
  for (const p of scene.people) {
    const f = p.flags
    if (f & flags.attacking) fighters.push(p)
    else if (f & flags.hurt) hurt.push(p)
    if (f & (flags.talking | flags.trading)) pairing.push(p)
    if (f & flags.swimming) {
      const sinking = p.health < 0.4
      out.push(person(p, sinking ? 'drowning' : 'swim', sinking ? 58 : 18, sinking ? 'Hampir tenggelam' : 'Berenang'))
    } else if (f & flags.rafting) out.push(person(p, 'raft', 24, 'Berakit'))
    if (f & flags.fallen) out.push(person(p, 'fall', 38, 'Terjatuh'))
    if (f & flags.leader) {
      const v = scene.villages.find((v) => v.leader?.id === p.id)
      out.push(person(p, 'leader', 16, v ? `Pemimpin Desa ${v.name}` : 'Pemimpin desa'))
    }
  }
  for (const p of fighters) {
    let near = 0
    for (const h of hurt) if (Math.hypot(h.rx - p.rx, h.ry - p.ry) < 3) near++
    for (const o of fighters) if (o !== p && Math.hypot(o.rx - p.rx, o.ry - p.ry) < 3) near++
    out.push(person(p, 'fight', 42 + 4 * Math.min(3, near), 'Perkelahian'))
  }
  for (const p of hurt) out.push(person(p, 'hurt', 12, 'Terluka'))
  // Pairs: each with its nearest partner within reach, once.
  const paired = new Set<number>()
  for (const p of pairing) {
    if (paired.has(p.id)) continue
    const trade = (p.flags & flags.trading) !== 0
    let best: ScenePerson | null = null
    let bestD = 2.5
    for (const o of pairing) {
      if (o === p || paired.has(o.id)) continue
      const d = Math.hypot(o.rx - p.rx, o.ry - p.ry)
      if (d < bestD) {
        bestD = d
        best = o
      }
    }
    if (!best) continue
    paired.add(p.id)
    paired.add(best.id)
    const a = Math.min(p.id, best.id)
    const b = Math.max(p.id, best.id)
    const trading = trade || (best.flags & flags.trading) !== 0
    out.push({
      key: `${trading ? 'trade' : 'talk'}:${a}-${b}`,
      kind: trading ? 'trade' : 'talk',
      id: a === p.id ? p.id : best.id,
      partner: a === p.id ? best.id : p.id,
      x: (p.rx + best.rx) / 2,
      z: (p.ry + best.ry) / 2,
      radius: bestD / 2 + 0.5,
      score: trading ? 26 : 20,
      label: trading ? 'Dua orang bertukar barang' : 'Dua orang bercakap',
    })
  }

  // Villages, for a wide view now and then.
  for (const v of scene.villages) {
    let r = 4
    for (const [hx, hy] of v.hull) r = Math.max(r, Math.hypot(hx - v.x, hy - v.y))
    out.push({
      key: `village:${v.id}`,
      kind: 'village',
      id: null,
      partner: null,
      x: v.x,
      z: v.y,
      radius: r,
      score: 10 + Math.min(10, v.people / 50),
      label: `Desa ${v.name}`,
    })
  }
  if (scene.hints) out.push(...scene.hints)
  return out
}

function person(p: ScenePerson, kind: SubjectKind, score: number, label: string): Subject {
  return { key: `${kind}:${p.id}`, kind, id: p.id, partner: null, x: p.rx, z: p.ry, radius: 1, score, label }
}

/** The shots that suit each kind of subject, in the order they are used. */
export const SHOTS: Record<SubjectKind, Shot[]> = {
  fire: ['establishing', 'orbit'],
  drowning: ['follow', 'orbit'],
  corpse: ['orbit'],
  fight: ['orbit', 'follow'],
  fall: ['orbit'],
  raft: ['follow', 'orbit'],
  trade: ['twoShot', 'orbit'],
  talk: ['twoShot', 'orbit'],
  swim: ['follow', 'orbit'],
  leader: ['follow', 'orbit'],
  hurt: ['orbit'],
  village: ['establishing'],
  event: ['orbit'],
}

export type DirectorOptions = {
  /** Seconds a subject is kept before anything may replace it. */
  minHold: number
  /** After this long, the subject's appeal halves so others get a turn. */
  maxHold: number
  /** A challenger must beat the current subject's score × margin (+ a little). */
  margin: number
  /** Seconds a subject may vanish (a flag flickering off) before it is given up. */
  grace: number
  /** Seconds a shown subject (and others of its kind right beside it) stays less appealing. */
  recent: number
  /** Seconds before a subject's next shot. */
  shotLength: number
}

export const DIRECTOR: DirectorOptions = { minHold: 6, maxHold: 24, margin: 1.3, grace: 2.5, recent: 90, shotLength: 11 }

/** The decision of one update: whether the subject or just the shot changed. */
export type Decision = { subject: Subject | null; shot: Shot; changedSubject: boolean; changedShot: boolean }

/**
 * Keeps one subject at a time. Call update() about once a second with the subjects found and
 * where the camera looks; it switches only when the current one has been held long enough and a
 * challenger clearly beats it, or when the current one has gone for longer than the grace.
 */
export class Director {
  current: Subject | null = null
  shot: Shot = 'orbit'
  since = 0
  shotSince = 0
  private shotIndex = 0
  private missingSince: number | null = null
  private history: { kind: SubjectKind; key: string; x: number; z: number; at: number }[] = []
  private readonly opts: DirectorOptions

  constructor(opts: Partial<DirectorOptions> = {}) {
    this.opts = { ...DIRECTOR, ...opts }
  }

  /** How appealing a subject is from where the camera looks now: nearer is a little better, recent much worse. */
  appeal(s: Subject, now: number, camX: number, camZ: number) {
    let v = s.score / (1 + Math.hypot(s.x - camX, s.z - camZ) / 120)
    for (const h of this.history) {
      if (now - h.at > this.opts.recent) continue
      if (h.key === s.key || (h.kind === s.kind && Math.hypot(h.x - s.x, h.z - s.z) < 8)) {
        v *= 0.45
        break
      }
    }
    return v
  }

  update(now: number, subjects: readonly Subject[], camX: number, camZ: number): Decision {
    const o = this.opts
    let changedSubject = false
    let changedShot = false
    const cur = this.current
    let found: Subject | undefined
    if (cur) {
      found = subjects.find((s) => s.key === cur.key)
      if (found) {
        this.missingSince = null
        cur.x = found.x
        cur.z = found.z
        cur.score = found.score
        cur.radius = found.radius
        cur.label = found.label
        cur.partner = found.partner
      } else if (this.missingSince === null) this.missingSince = now
    }
    const lost = cur !== null && this.missingSince !== null && now - this.missingSince > o.grace
    let curAppeal = -Infinity
    if (cur && !lost) {
      curAppeal = cur.score / (1 + Math.hypot(cur.x - camX, cur.z - camZ) / 120)
      if (!found) curAppeal *= 0.5
      if (now - this.since > o.maxHold) curAppeal *= 0.5
    }
    let best: Subject | null = null
    let bestAppeal = -Infinity
    for (const s of subjects) {
      if (cur && s.key === cur.key) continue
      const a = this.appeal(s, now, camX, camZ)
      if (a > bestAppeal) {
        bestAppeal = a
        best = s
      }
    }
    const held = now - this.since
    const take = best !== null && (cur === null || lost || (held >= o.minHold && bestAppeal > curAppeal * o.margin + 2))
    if (take && best) {
      if (cur) this.history.push({ kind: cur.kind, key: cur.key, x: cur.x, z: cur.z, at: now })
      if (this.history.length > 24) this.history.splice(0, this.history.length - 24)
      this.current = { ...best }
      this.since = now
      this.missingSince = null
      this.shotIndex = 0
      this.shot = SHOTS[best.kind][0]
      this.shotSince = now
      changedSubject = true
      changedShot = true
    } else if (lost) {
      if (cur) this.history.push({ kind: cur.kind, key: cur.key, x: cur.x, z: cur.z, at: now })
      this.current = null
      this.missingSince = null
      changedSubject = true
    } else if (this.current && now - this.shotSince >= o.shotLength) {
      // The same subject from another angle (the shot list's next).
      const list = SHOTS[this.current.kind]
      if (list.length > 1) {
        this.shotIndex = (this.shotIndex + 1) % list.length
        this.shot = list[this.shotIndex]
        changedShot = true
      }
      this.shotSince = now
    }
    return { subject: this.current, shot: this.shot, changedSubject, changedShot }
  }
}

/** Where the camera stands and what it looks at (tiles). */
export type Framing = { eyeX: number; eyeY: number; eyeZ: number; lookX: number; lookY: number; lookZ: number }

/** The subject as framed right now: where it stands (y: its ground), which way it faces, and its partner for a two-shot. */
export type Target = {
  x: number
  y: number
  z: number
  heading: number
  radius: number
  partnerX?: number
  partnerZ?: number
  partnerY?: number
}

/**
 * A shot's camera at `t` seconds into it. `seed` (0–1) varies the side and starting angle, and
 * `scale` is how tall people are drawn (tiles), so close shots keep people a good size.
 */
export function frameShot(shot: Shot, s: Target, t: number, seed: number, scale: number, out: Framing): Framing {
  const side = seed < 0.5 ? -1 : 1
  switch (shot) {
    case 'follow': {
      // Behind the shoulder, a little to one side, looking past them along their way.
      const fx = Math.cos(s.heading)
      const fz = Math.sin(s.heading)
      const back = 3.2 * scale
      const off = 0.8 * scale * side
      out.eyeX = s.x - fx * back - fz * off
      out.eyeZ = s.z - fz * back + fx * off
      out.eyeY = s.y + 1.9 * scale
      out.lookX = s.x + fx * 2 * scale
      out.lookZ = s.z + fz * 2 * scale
      out.lookY = s.y + 0.55 * scale
      return out
    }
    case 'twoShot': {
      // Square on to the line between the two, far enough back to hold both.
      const px = s.partnerX ?? s.x + 1
      const pz = s.partnerZ ?? s.z
      const mx = (s.x + px) / 2
      const mz = (s.z + pz) / 2
      const ax = px - s.x
      const az = pz - s.z
      const gap = Math.hypot(ax, az) || 1
      const nx = (-az / gap) * side
      const nz = (ax / gap) * side
      const back = 3 * scale + gap * 1.3
      // A slow drift along the line keeps a long conversation alive.
      const drift = Math.sin(t * 0.15 + seed * 6) * 0.35 * scale
      out.eyeX = mx + nx * back + (ax / gap) * drift
      out.eyeZ = mz + nz * back + (az / gap) * drift
      out.eyeY = Math.max(s.y, s.partnerY ?? s.y) + 1.5 * scale
      out.lookX = mx
      out.lookZ = mz
      out.lookY = (s.y + (s.partnerY ?? s.y)) / 2 + 0.5 * scale
      return out
    }
    case 'establishing': {
      // High and wide, drifting slowly round.
      const r = Math.max(16, s.radius * 1.7)
      const a = seed * Math.PI * 2 + t * 0.025
      out.eyeX = s.x + Math.cos(a) * r
      out.eyeZ = s.z + Math.sin(a) * r
      out.eyeY = s.y + r * 0.8
      out.lookX = s.x
      out.lookZ = s.z
      out.lookY = s.y
      return out
    }
    default: {
      // Orbit: round the subject at a walking pace.
      const r = 4.5 * scale + s.radius * 1.4
      const a = seed * Math.PI * 2 + t * 0.11 * side
      out.eyeX = s.x + Math.cos(a) * r
      out.eyeZ = s.z + Math.sin(a) * r
      out.eyeY = s.y + 2.2 * scale + s.radius * 0.6
      out.lookX = s.x
      out.lookZ = s.z
      out.lookY = s.y + 0.45 * scale
      return out
    }
  }
}

/**
 * A critically damped spring (as game cameras smooth their moves): moves `state[i]` towards
 * `target` with its velocity in `state[i + 1]`, settling in about `smoothTime` seconds without
 * overshooting.
 */
export function smoothDamp(state: Float64Array, i: number, target: number, smoothTime: number, dt: number) {
  const omega = 2 / Math.max(1e-4, smoothTime)
  const x = omega * dt
  const exp = 1 / (1 + x + 0.48 * x * x + 0.235 * x * x * x)
  const change = state[i] - target
  const temp = (state[i + 1] + omega * change) * dt
  state[i + 1] = (state[i + 1] - omega * temp) * exp
  let next = target + (change + temp) * exp
  // No overshoot past the target.
  if (target - state[i] > 0 === next > target) {
    next = target
    state[i + 1] = 0
  }
  state[i] = next
  return next
}
