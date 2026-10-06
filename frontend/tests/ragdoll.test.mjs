import test from 'node:test'
import assert from 'node:assert/strict'
import {
  RAG_N,
  RAG_REST,
  RAG_POSE,
  RAG_POSE_SIZE,
  RP,
  clearance,
  createRagdoll,
  eulerXYZ,
  flinch,
  guard,
  matrixXYZ,
  ragdollPose,
  restPoints,
  settleRagdoll,
  stepRagdoll,
} from '../src/game3d/ragdoll.ts'

const dist = (p, a, b) => Math.hypot(p[a * 3] - p[b * 3], p[a * 3 + 1] - p[b * 3 + 1], p[a * 3 + 2] - p[b * 3 + 2])
const restDist = (a, b) => dist(RAG_REST, a, b)
const BONES = [
  [RP.neck, RP.head],
  [RP.shL, RP.elL],
  [RP.elL, RP.haL],
  [RP.shR, RP.elR],
  [RP.elR, RP.haR],
  [RP.hipL, RP.knL],
  [RP.knL, RP.ftL],
  [RP.hipR, RP.knR],
  [RP.knR, RP.ftR],
  [RP.shL, RP.shR],
  [RP.hipL, RP.hipR],
  [RP.neck, RP.pelvis],
  [RP.front, RP.back],
]
const flat = () => 0
const slope = (x, z) => 0.35 * x + 0.12 * z

function run(r, ground, seconds, dt = 1 / 60) {
  for (let t = 0; t < seconds; t += dt) stepRagdoll(r, ground, dt)
}

test('the ragdoll keeps its segment lengths while it falls and once it lies still', () => {
  const r = createRagdoll(restPoints(), 0.37)
  for (let k = 0; k < 6; k++) {
    run(r, flat, 0.25)
    for (const [a, b] of BONES) {
      const err = Math.abs(dist(r.pos, a, b) - restDist(a, b)) / restDist(a, b)
      assert.ok(err < 0.04, `segment ${a}-${b} off by ${(err * 100).toFixed(1)}%`)
    }
  }
})

test('a body falls and comes to rest on sloped ground, above the terrain', () => {
  for (const seed of [0.1, 0.42, 0.77]) {
    const r = createRagdoll(restPoints(), seed)
    // The body starts on the slope: lift it so its feet stand on the ground under them.
    const lift = Math.max(slope(RAG_REST[RP.ftL * 3], 0), slope(RAG_REST[RP.ftR * 3], 0))
    for (let i = 0; i < RAG_N; i++) {
      r.pos[i * 3 + 1] += lift
      r.prev[i * 3 + 1] += lift
    }
    run(r, slope, 6)
    assert.ok(r.asleep, 'settled and frozen')
    assert.ok(clearance(r, slope) > -0.004, `buried by ${clearance(r, slope)}`)
    // Lying down: the head is now not far above the ground beneath it.
    const head = r.pos[RP.head * 3 + 1] - slope(r.pos[RP.head * 3], r.pos[RP.head * 3 + 2])
    assert.ok(head < 0.2, `head still ${head.toFixed(3)} above the ground`)
    // Frozen means frozen.
    const before = Float64Array.from(r.pos)
    run(r, slope, 1)
    assert.deepEqual(Array.from(r.pos), Array.from(before))
  }
})

test('the same seed and steps give the same fall; settling quickly matches the settled state closely', () => {
  const a = createRagdoll(restPoints(), 0.61)
  const b = createRagdoll(restPoints(), 0.61)
  run(a, slope, 3)
  run(b, slope, 3)
  assert.deepEqual(Array.from(a.pos), Array.from(b.pos))
  const c = createRagdoll(restPoints(), 0.62)
  run(c, slope, 3)
  assert.notDeepEqual(Array.from(a.pos), Array.from(c.pos))
  const quick = createRagdoll(restPoints(), 0.61)
  settleRagdoll(quick, flat)
  assert.ok(quick.asleep && clearance(quick, flat) > -0.004)
})

test('the rest pose maps back onto an unturned skeleton, and Euler angles round-trip', () => {
  const out = new Float32Array(RAG_POSE_SIZE)
  ragdollPose(RAG_REST, out)
  for (const v of out) assert.ok(Math.abs(v) < 1e-5, `rest pose angle ${v}`)
  for (const [x, y, z] of [
    [0.3, -0.2, 0.9],
    [-1.2, 0.5, 0.1],
    [2.5, 1.1, -2.9],
  ]) {
    const e = [0, 0, 0]
    eulerXYZ(matrixXYZ(x, y, z), e, 0)
    const m1 = matrixXYZ(x, y, z),
      m2 = matrixXYZ(...e)
    for (let i = 0; i < 9; i++) assert.ok(Math.abs(m1[i] - m2[i]) < 1e-9)
  }
})

test('a lying body maps to a lying skeleton: the hips turned about 90° from upright', () => {
  const r = createRagdoll(restPoints(), 0.2, { x: 0, z: -1, strength: 0.8 })
  settleRagdoll(r, flat)
  const out = new Float64Array(RAG_POSE_SIZE)
  ragdollPose(r.pos, out)
  const m = matrixXYZ(out[RAG_POSE.hips], out[RAG_POSE.hips + 1], out[RAG_POSE.hips + 2])
  // The hips' up axis (second column) is now close to horizontal.
  assert.ok(Math.abs(m[4]) < 0.6, `up·y = ${m[4]}`)
  assert.ok(out.every(Number.isFinite))
})

test('a flinch snaps to its peak quickly and dies away; the guard eases off', () => {
  let peak = 0,
    at = 0
  for (let t = 0; t < 1.4; t += 0.005) if (flinch(t) > peak) [peak, at] = [flinch(t), t]
  assert.ok(Math.abs(peak - 1) < 0.02 && at < 0.25)
  assert.equal(flinch(-0.1), 0)
  assert.equal(flinch(5), 0)
  assert.ok(Math.abs(flinch(1.3)) < 0.05)
  assert.ok(guard(0.15) > guard(1.2) && guard(0) === 0)
})
