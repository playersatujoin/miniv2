// Verlet cloth for the sarong and kain of the few people nearest the camera, after RAGE's
// character cloth (cloth/characterclothcontroller, pheffects/cloth_verlet): the top row of
// vertices is pinned to the skeleton, the rest integrate with gravity and air drag towards the
// wind, edge constraints keep the weave from stretching (but let it bunch up), the legs push it
// aside as capsules, and a fresh or teleported cloth is pinned into its rest shape first
// (characterClothController's force-pin). Pure math, no three.js. Positions are in world units;
// rest lengths in body units, scaled by the wearer's size each step.

export type Cloth = {
  cols: number
  rows: number
  /** Positions and the previous step's, world units, row-major (row 0 is pinned). */
  pos: Float64Array
  prev: Float64Array
  /** Constraint ends, rest lengths (body units) and how stiff (1 rigid … 0 slack). */
  ca: Int32Array
  cb: Int32Array
  rest: Float64Array
  stiff: Float64Array
  /** How much a constraint may shrink before it pushes back (cloth bunches up freely). */
  give: Float64Array
  acc: number
  time: number
}

/** A skirt's rest shape in body space: `cols` around, `rows` from the waist (row 0) to the hem, slightly flared. */
export function skirtShape(cols: number, rows: number, top: number, hem: number, topR: number, hemR: number, depth = 0.92) {
  const out = new Float64Array(cols * rows * 3)
  for (let r = 0; r < rows; r++) {
    const f = r / (rows - 1)
    const y = top + (hem - top) * f
    const rad = topR + (hemR - topR) * Math.pow(f, 0.8)
    for (let c = 0; c < cols; c++) {
      const a = (c / cols) * Math.PI * 2
      const o = (r * cols + c) * 3
      out[o] = Math.sin(a) * rad
      out[o + 1] = y
      out[o + 2] = Math.cos(a) * rad * depth
    }
  }
  return out
}

/** A tube of cloth with rest lengths from `shape` (body units) and starting at `start` (world units, same layout). */
export function createCloth(cols: number, rows: number, shape: ArrayLike<number>, start: ArrayLike<number>): Cloth {
  const ca: number[] = [],
    cb: number[] = [],
    rest: number[] = [],
    stiff: number[] = [],
    give: number[] = []
  const id = (r: number, c: number) => r * cols + (((c % cols) + cols) % cols)
  const link = (a: number, b: number, k: number, g: number) => {
    ca.push(a)
    cb.push(b)
    rest.push(Math.hypot(shape[a * 3] - shape[b * 3], shape[a * 3 + 1] - shape[b * 3 + 1], shape[a * 3 + 2] - shape[b * 3 + 2]))
    stiff.push(k)
    give.push(g)
  }
  for (let r = 0; r < rows; r++)
    for (let c = 0; c < cols; c++) {
      // Around the tube: firm against stretching, loose enough to gather.
      link(id(r, c), id(r, c + 1), 0.8, 0.6)
      if (r + 1 < rows) {
        // Down the tube: the cloth hangs at its length.
        link(id(r, c), id(r + 1, c), 1, 0.5)
        // Shear, both ways.
        link(id(r, c), id(r + 1, c + 1), 0.35, 0.3)
        link(id(r, c + 1), id(r + 1, c), 0.35, 0.3)
      }
      // Bending, every other row: a little body to the weave.
      if (r + 2 < rows) link(id(r, c), id(r + 2, c), 0.25, 0.2)
    }
  const n = cols * rows
  const pos = Float64Array.from({ length: n * 3 }, (_, i) => start[i])
  return {
    cols,
    rows,
    pos,
    prev: Float64Array.from(pos),
    ca: Int32Array.from(ca),
    cb: Int32Array.from(cb),
    rest: Float64Array.from(rest),
    stiff: Float64Array.from(stiff),
    give: Float64Array.from(give),
    acc: 0,
    time: 0,
  }
}

/** What moves the cloth this frame. */
export type ClothForces = {
  /** The pinned row (cols × xyz, world units). */
  pins: ArrayLike<number>
  /** Collision capsules: ax, ay, az, bx, by, bz, radius per capsule (world units). */
  capsules: ArrayLike<number>
  capsuleCount: number
  /** The wind's velocity (world units a second, x and z). */
  windX: number
  windZ: number
  /** World units per body unit (the wearer's size). */
  scale: number
  /** The ground under the wearer (world y): the hem never goes below it. */
  floor: number
}

const STEP = 1 / 60
/** Body units per second² (as the ragdoll). */
const GRAVITY = 3.15
/** How quickly the cloth takes up the air's speed (per second). */
const DRAG = 3.2
const DAMP = 0.985
const ITERATIONS = 4

/**
 * Advances the cloth by `dt` seconds in fixed steps (at most three a call; time beyond that is
 * dropped, so a long frame or a large dt never blows it up).
 */
export function stepCloth(c: Cloth, f: ClothForces, dt: number) {
  c.acc += Math.min(Math.max(0, dt), 0.2)
  let n = 0
  while (c.acc >= STEP && n < 3) {
    substep(c, f, STEP)
    c.acc -= STEP
    n++
  }
  if (n === 3) c.acc = 0
  pin(c, f.pins)
}

/** Puts the cloth back into its rest drape under the pins (a force-pin), e.g. after a teleport. */
export function resetCloth(c: Cloth, shapeWorld: ArrayLike<number>) {
  c.pos.set(Array.from({ length: c.pos.length }, (_, i) => shapeWorld[i]))
  c.prev.set(c.pos)
  c.acc = 0
}

function pin(c: Cloth, pins: ArrayLike<number>) {
  for (let i = 0; i < c.cols * 3; i++) {
    c.pos[i] = pins[i]
    c.prev[i] = pins[i]
  }
}

function substep(c: Cloth, f: ClothForces, h: number) {
  const { pos, prev, cols } = c
  const n = pos.length / 3
  c.time += h
  // Gusts: the wind rises and falls, a little differently around the tube.
  const gust = 0.75 + 0.25 * Math.sin(c.time * 2.3) + 0.15 * Math.sin(c.time * 5.9)
  const g = GRAVITY * f.scale * h * h
  // Furthest a vertex may move in a step: a teleport or a spike never throws the cloth away.
  const maxMove = 0.08 * f.scale
  pin(c, f.pins)
  for (let i = cols; i < n; i++) {
    const o = i * 3
    const x = pos[o],
      y = pos[o + 1],
      z = pos[o + 2]
    let vx = (x - prev[o]) * DAMP,
      vy = (y - prev[o + 1]) * DAMP,
      vz = (z - prev[o + 2]) * DAMP
    // Air drag towards the wind's velocity (per step: speed·h).
    const flutter = gust * (1 + 0.2 * Math.sin(c.time * 7 + i * 1.7))
    vx += (f.windX * flutter * h - vx) * DRAG * h
    vz += (f.windZ * flutter * h - vz) * DRAG * h
    vy -= g
    const m = Math.hypot(vx, vy, vz)
    if (m > maxMove) {
      vx *= maxMove / m
      vy *= maxMove / m
      vz *= maxMove / m
    }
    prev[o] = x
    prev[o + 1] = y
    prev[o + 2] = z
    pos[o] = x + vx
    pos[o + 1] = y + vy
    pos[o + 2] = z + vz
  }
  for (let it = 0; it < ITERATIONS; it++) {
    links(c, f.scale)
    collide(c, f)
  }
  if (!pos.every(Number.isFinite)) {
    // Never: but a cloth that broke is better hung back up than drawn as spikes.
    for (let i = cols; i < n; i++) {
      const o = i * 3,
        top = (i % cols) * 3
      pos[o] = prev[o] = f.pins[top]
      pos[o + 1] = prev[o + 1] = f.pins[top + 1] - (Math.floor(i / cols) / c.rows) * 0.2 * f.scale
      pos[o + 2] = prev[o + 2] = f.pins[top + 2]
    }
  }
}

function links(c: Cloth, scale: number) {
  const { pos, ca, cb, rest, stiff, give, cols } = c
  for (let k = 0; k < ca.length; k++) {
    const a = ca[k],
      b = cb[k]
    const ao = a * 3,
      bo = b * 3
    const dx = pos[bo] - pos[ao],
      dy = pos[bo + 1] - pos[ao + 1],
      dz = pos[bo + 2] - pos[ao + 2]
    const d = Math.sqrt(dx * dx + dy * dy + dz * dz)
    if (d < 1e-9) continue
    const r = rest[k] * scale
    // Stretching is resisted fully, shrinking only past the give (the cloth bunches).
    let diff = d - r
    if (diff < 0) {
      diff = d < r * (1 - give[k]) ? d - r * (1 - give[k]) : 0
      if (!diff) continue
    }
    const wa = a < cols ? 0 : 1,
      wb = b < cols ? 0 : 1
    const w = wa + wb
    if (!w) continue
    const s = (diff / d) * stiff[k]
    pos[ao] += dx * s * (wa / w)
    pos[ao + 1] += dy * s * (wa / w)
    pos[ao + 2] += dz * s * (wa / w)
    pos[bo] -= dx * s * (wb / w)
    pos[bo + 1] -= dy * s * (wb / w)
    pos[bo + 2] -= dz * s * (wb / w)
  }
}

function collide(c: Cloth, f: ClothForces) {
  const { pos, cols } = c
  const n = pos.length / 3
  const caps = f.capsules
  const floor = f.floor + 0.004 * f.scale
  for (let i = cols; i < n; i++) {
    const o = i * 3
    let x = pos[o],
      y = pos[o + 1],
      z = pos[o + 2]
    for (let k = 0; k < f.capsuleCount; k++) {
      const q = k * 7
      const ax = caps[q],
        ay = caps[q + 1],
        az = caps[q + 2]
      const ex = caps[q + 3] - ax,
        ey = caps[q + 4] - ay,
        ez = caps[q + 5] - az
      const rad = caps[q + 6]
      const len2 = ex * ex + ey * ey + ez * ez || 1e-12
      let t = ((x - ax) * ex + (y - ay) * ey + (z - az) * ez) / len2
      t = t < 0 ? 0 : t > 1 ? 1 : t
      const px = x - (ax + ex * t),
        py = y - (ay + ey * t),
        pz = z - (az + ez * t)
      const d2 = px * px + py * py + pz * pz
      if (d2 >= rad * rad) continue
      const d = Math.sqrt(d2)
      if (d < 1e-9) continue
      const push = (rad - d) / d
      x += px * push
      y += py * push
      z += pz * push
    }
    if (y < floor) y = floor
    pos[o] = x
    pos[o + 1] = y
    pos[o + 2] = z
  }
}
