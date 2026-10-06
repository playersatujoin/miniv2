import test from 'node:test'
import assert from 'node:assert/strict'
import { createCloth, skirtShape, stepCloth } from '../src/game3d/cloth.ts'

const COLS = 12,
  ROWS = 7
const SCALE = 1.6

/** A skirt hung at the origin, scaled to world units, with its waist pinned where it rests. */
function hang() {
  const shape = skirtShape(COLS, ROWS, 0.32, 0.07, 0.05, 0.088)
  const world = shape.map((v) => v * SCALE)
  const cloth = createCloth(COLS, ROWS, shape, world)
  const pins = world.slice(0, COLS * 3)
  return { cloth, pins, world }
}

const forces = (pins, wind = [0, 0], capsules = [], extra = {}) => ({
  pins,
  capsules,
  capsuleCount: capsules.length / 7,
  windX: wind[0],
  windZ: wind[1],
  scale: SCALE,
  floor: -10,
  ...extra,
})

const hemCentre = (c) => {
  let x = 0,
    z = 0
  for (let k = 0; k < COLS; k++) {
    const o = ((ROWS - 1) * COLS + k) * 3
    x += c.pos[o]
    z += c.pos[o + 2]
  }
  return [x / COLS, z / COLS]
}

test('pinned points stay pinned, even when the pins move', () => {
  const { cloth, pins } = hang()
  const moved = Float64Array.from(pins)
  for (let step = 0; step < 120; step++) {
    for (let k = 0; k < COLS; k++) moved[k * 3] = pins[k * 3] + Math.sin(step * 0.1) * 0.2
    stepCloth(cloth, forces(moved, [1, 0]), 1 / 60)
    for (let i = 0; i < COLS * 3; i++) assert.equal(cloth.pos[i], moved[i])
  }
})

test('wind pushes the hem downwind', () => {
  for (const [wx, wz] of [
    [2, 0],
    [0, -2],
    [-1.4, 1.4],
  ]) {
    const { cloth, pins } = hang()
    for (let step = 0; step < 240; step++) stepCloth(cloth, forces(pins, [wx, wz]), 1 / 60)
    const [x, z] = hemCentre(cloth)
    assert.ok(x * wx + z * wz > 0.01, `hem moved ${x.toFixed(3)}, ${z.toFixed(3)} for wind ${wx}, ${wz}`)
  }
})

test('without wind it hangs: the hem stays under the waist and below it', () => {
  const { cloth, pins } = hang()
  for (let step = 0; step < 240; step++) stepCloth(cloth, forces(pins), 1 / 60)
  const [x, z] = hemCentre(cloth)
  assert.ok(Math.hypot(x, z) < 0.01)
  for (let k = 0; k < COLS; k++) assert.ok(cloth.pos[((ROWS - 1) * COLS + k) * 3 + 1] < pins[k * 3 + 1])
})

test('stable with large time steps, teleports and a leg inside it: no explosion', () => {
  const { cloth, pins } = hang()
  const leg = [0.06 * SCALE, 0.25 * SCALE, 0.05 * SCALE, 0.06 * SCALE, 0.03 * SCALE, 0.12 * SCALE, 0.03 * SCALE]
  const far = Float64Array.from(pins)
  for (let step = 0; step < 50; step++) {
    if (step === 25) for (let k = 0; k < COLS; k++) far[k * 3] += 40
    stepCloth(cloth, forces(far, [5, 5], leg), step % 2 ? 1 : 0.5)
    assert.ok(cloth.pos.every(Number.isFinite))
  }
  // Every vertex stays within reach of its pin: the cloth never flies apart.
  for (let i = COLS; i < COLS * ROWS; i++) {
    const p = (i % COLS) * 3
    const d = Math.hypot(cloth.pos[i * 3] - far[p], cloth.pos[i * 3 + 1] - far[p + 1], cloth.pos[i * 3 + 2] - far[p + 2])
    assert.ok(d < 0.6 * SCALE, `vertex ${i} drifted ${d.toFixed(2)}`)
  }
})

test('legs push the cloth out of their capsules', () => {
  const { cloth, pins } = hang()
  const r = 0.04 * SCALE
  // A thigh swung forward through the front of the skirt.
  const leg = [0.038 * SCALE, 0.25 * SCALE, 0, 0.038 * SCALE, 0.16 * SCALE, 0.09 * SCALE, r]
  for (let step = 0; step < 120; step++) stepCloth(cloth, forces(pins, [0, 0], leg), 1 / 60)
  for (let i = COLS; i < COLS * ROWS; i++) {
    const o = i * 3
    const ax = leg[0],
      ay = leg[1],
      az = leg[2]
    const ex = leg[3] - ax,
      ey = leg[4] - ay,
      ez = leg[5] - az
    let t = ((cloth.pos[o] - ax) * ex + (cloth.pos[o + 1] - ay) * ey + (cloth.pos[o + 2] - az) * ez) / (ex * ex + ey * ey + ez * ez)
    t = Math.max(0, Math.min(1, t))
    const d = Math.hypot(cloth.pos[o] - ax - ex * t, cloth.pos[o + 1] - ay - ey * t, cloth.pos[o + 2] - az - ez * t)
    assert.ok(d > r * 0.9, `vertex ${i} inside the leg (${d.toFixed(4)} < ${r.toFixed(4)})`)
  }
})
