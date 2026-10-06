import test from 'node:test'
import assert from 'node:assert/strict'
import { footSpots, liftOf, quadPose, solveQuad, easeQuad, QUAD_LIMITS } from '../src/game3d/quadruped.ts'

const deer = { halfLength: 0.14, halfWidth: 0.06, legLength: 0.24 }
const near = (a, b, e = 1e-6) => Math.abs(a - b) < e

/** Where each foot ends up in world y for a solved pose (as the shader scales the legs). */
function feetY(pose, build, scale) {
  const hx = build.halfLength * scale,
    hz = build.halfWidth * scale,
    len = build.legLength * scale
  return [
    [hx, hz],
    [hx, -hz],
    [-hx, hz],
    [-hx, -hz],
  ].map(([x, z], i) => pose.y + liftOf(x, len - pose.legs[i] * len, z, pose.pitch, pose.roll))
}

test('flat ground: no tilt, legs as modelled, standing on the ground', () => {
  const p = solveQuad([2, 2, 2, 2], deer, 1.25, quadPose())
  assert.ok(near(p.pitch, 0) && near(p.roll, 0))
  assert.ok(near(p.y, 2))
  for (const k of p.legs) assert.ok(near(k, 1))
})

test('a slope along the body pitches it (nose up uphill) and every foot meets the ground', () => {
  const scale = 1.25
  const hx = deer.halfLength * scale
  const rise = 0.3 // per tile
  const ground = [hx * rise, hx * rise, -hx * rise, -hx * rise].map((v) => v + 1)
  const p = solveQuad(ground, deer, scale, quadPose())
  assert.ok(near(p.pitch, Math.atan(rise), 1e-9))
  const feet = feetY(p, deer, scale)
  feet.forEach((y, i) => assert.ok(near(y, ground[i], 1e-6), `foot ${i}`))
})

test('a slope across the body rolls it partly, the +z side up when the ground rises to +z, legs take the rest', () => {
  const scale = 1
  const hz = deer.halfWidth
  const ground = [hz * 0.5, -hz * 0.5, hz * 0.5, -hz * 0.5]
  const p = solveQuad(ground, deer, scale, quadPose())
  assert.ok(p.roll < 0, 'a negative roll about +x lifts the +z side')
  const len = deer.legLength
  assert.ok(liftOf(0, len, hz, p.pitch, p.roll) > liftOf(0, len, -hz, p.pitch, p.roll))
  // The uphill legs shorten and the downhill ones lengthen.
  assert.ok(p.legs[0] < 1 && p.legs[1] > 1)
  feetY(p, deer, scale).forEach((y, i) => assert.ok(near(y, ground[i], 1e-6)))
})

test('a foot over a hole stretches its leg, within the limit; the body never floats off every foot', () => {
  const p = solveQuad([0, 0, 0, -0.05], deer, 1, quadPose())
  assert.ok(p.legs[3] > 1 && p.legs[3] <= QUAD_LIMITS.maxScale + 1e-6)
  const deep = solveQuad([0, 0, 0, -5], deer, 1, quadPose())
  for (const k of deep.legs) assert.ok(k >= QUAD_LIMITS.minScale - 1e-6 && k <= QUAD_LIMITS.maxScale + 1e-6 && Number.isFinite(k))
  assert.ok(Math.abs(deep.pitch) <= QUAD_LIMITS.pitch && Math.abs(deep.roll) <= QUAD_LIMITS.roll)
  const feet = feetY(deep, deer, 1)
  assert.equal(feet.filter((y, i) => Math.abs(y - [0, 0, 0, -5][i]) < 1e-6).length, 3, 'it stands on the other three')
  assert.ok(Math.abs(deep.pitch) < 0.05 && Math.abs(deep.roll) < 0.05, 'one ditch does not tip the body')
  assert.ok(Math.abs(deep.y) < 0.05, 'nor sink it')
})

test('feet are placed under the body for any heading', () => {
  const out = new Float32Array(8)
  footSpots(10, 20, Math.PI / 2, deer, 1, out)
  // Facing +z: the fore feet are further along z.
  assert.ok(out[1] > 20 && out[3] > 20 && out[5] < 20 && out[7] < 20)
  // Model +z is the heading's right in the map (x, z): facing +z, that is −x.
  assert.ok(out[0] < 10 && out[2] > 10)
  footSpots(0, 0, 0, deer, 2, out)
  assert.ok(near(out[0], 0.28) && near(out[1], 0.12))
})

test('easing moves a pose smoothly towards its target', () => {
  const a = quadPose()
  const t = solveQuad([0.1, 0.1, 0, 0], deer, 1, quadPose())
  easeQuad(a, t, 1 / 60)
  assert.ok(a.pitch > 0 && a.pitch < t.pitch)
  for (let i = 0; i < 600; i++) easeQuad(a, t, 1 / 60)
  assert.ok(near(a.pitch, t.pitch, 1e-6))
})
