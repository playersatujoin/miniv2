import test from 'node:test'
import assert from 'node:assert/strict'
import { bodyProportions, solveLeg } from '../src/game3d/anatomy.ts'
import { lightCycle } from '../src/game3d/timecycle.ts'

test('age continuously changes stature without changing identity', () => {
  const newborn = bodyProportions(0, 0.5, false),
    teen = bodyProportions(12, 0.5, false),
    adult = bodyProportions(25, 0.5, false),
    elder = bodyProportions(75, 0.5, false)
  assert.ok(newborn.height < teen.height && teen.height < adult.height)
  assert.ok(newborn.head > adult.head)
  assert.ok(elder.grey > adult.grey && elder.stoop > adult.stoop)
})

test('contact IK reconstructs a reachable ankle and clamps impossible terrain', () => {
  for (const [y, z] of [
    [-0.19, 0.03],
    [-0.16, -0.06],
    [-0.2, 0],
  ]) {
    const { hip, knee } = solveLeg(y, z)
    const actualY = -0.115 * Math.cos(hip) - 0.109 * Math.cos(hip + knee)
    const actualZ = -0.115 * Math.sin(hip) - 0.109 * Math.sin(hip + knee)
    assert.ok(Math.abs(actualY - y) < 1e-6 && Math.abs(actualZ - z) < 1e-6)
  }
  for (const [y, z] of [
    [0, 0],
    [-100, 100],
    [1, 1],
  ]) {
    const { hip, knee } = solveLeg(y, z)
    assert.ok(Number.isFinite(hip) && Number.isFinite(knee))
  }
})

test('night remains legible and cloud cover reduces direct light', () => {
  const night = lightCycle(0, 0),
    day = lightCycle(1, 0),
    cloud = lightCycle(1, 1)
  assert.ok(night.ambient > 0 && night.sun < day.sun)
  assert.ok(cloud.sun < day.sun)
  assert.ok(lightCycle(0.28, 0).warmth > day.warmth)
})

test('the ragdoll and the leg IK share the skeleton: limb lengths agree', async () => {
  const { RAG_REST, RP } = await import('../src/game3d/ragdoll.ts')
  const d = (a, b) =>
    Math.hypot(RAG_REST[a * 3] - RAG_REST[b * 3], RAG_REST[a * 3 + 1] - RAG_REST[b * 3 + 1], RAG_REST[a * 3 + 2] - RAG_REST[b * 3 + 2])
  // solveLeg's default thigh and shin.
  assert.ok(Math.abs(d(RP.hipL, RP.knL) - 0.115) < 1e-9 && Math.abs(d(RP.knR, RP.ftR) - 0.109) < 1e-9)
  // The rest ankle (straight below the hip) is all but in reach of a straight leg.
  const { hip, knee } = solveLeg(RAG_REST[RP.ftL * 3 + 1] - RAG_REST[RP.hipL * 3 + 1], 0)
  assert.ok(Math.abs(hip) < 0.1 && Math.abs(knee) < 0.15, `hip ${hip}, knee ${knee}`)
})
