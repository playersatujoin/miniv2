import test from 'node:test'
import assert from 'node:assert/strict'
import {
  FACE,
  FACE_CHANNELS,
  MOODS,
  POSTURE,
  POSTURE_CHANNELS,
  blendFactor,
  blinkAt,
  faceFrom,
  gestureAt,
  moodTarget,
  postureFrom,
  stepMood,
} from '../src/game3d/expressions.ts'

const settle = (mood, strength, seconds = 4, dt = 1 / 60) => {
  const w = new Float32Array(MOODS)
  for (let t = 0; t < seconds; t += dt) stepMood(w, mood, strength, blendFactor(dt))
  return w
}

test('blend weights per mood are bounded and continuous over time', () => {
  const w = new Float32Array(MOODS)
  const face = new Float32Array(FACE_CHANNELS)
  const prevFace = new Float32Array(FACE_CHANNELS)
  const dt = 1 / 60
  const k = blendFactor(dt)
  // Moods switching abruptly, with odd strengths (including out of range and missing).
  const script = [
    [1, 1],
    [2, 0.4],
    [3, 7],
    [4, -2],
    [0, 1],
    [3, undefined],
    [9, 1],
    [undefined, undefined],
  ]
  let first = true
  for (const [mood, strength] of script) {
    for (let step = 0; step < 50; step++) {
      const before = Float32Array.from(w)
      stepMood(w, mood, strength, k)
      for (let m = 0; m < MOODS; m++) {
        assert.ok(w[m] >= 0 && w[m] <= 1, `weight ${m} = ${w[m]}`)
        assert.ok(Math.abs(w[m] - before[m]) <= k + 1e-6, 'no jumps')
      }
      faceFrom(w, 0, 0, face)
      for (let c = 0; c < FACE_CHANNELS; c++) {
        assert.ok(face[c] >= -1 && face[c] <= 1)
        // A step moves the face only a little: the largest mood face component is ±1.
        if (!first) assert.ok(Math.abs(face[c] - prevFace[c]) <= 4 * k + 1e-6, `face channel ${c} jumped`)
      }
      prevFace.set(face)
      first = false
    }
  }
})

test('calm settles to the neutral face and posture', () => {
  const w = settle(1, 1, 2)
  for (let t = 0; t < 6; t += 1 / 60) stepMood(w, 0, 1, blendFactor(1 / 60))
  const face = faceFrom(w, 0, 0, new Float32Array(FACE_CHANNELS))
  for (const v of face) assert.ok(Math.abs(v) < 1e-4)
  const posture = postureFrom(w, new Float32Array(POSTURE_CHANNELS))
  for (const v of posture) assert.ok(Math.abs(v) < 1e-4)
  assert.deepEqual(Array.from(moodTarget(undefined, undefined, new Float32Array(MOODS))), [0, 0, 0, 0, 0])
})

test('each mood reads as itself', () => {
  const face = (m) => faceFrom(settle(m, 1), 0, 0, new Float32Array(FACE_CHANNELS))
  const fear = face(1),
    anger = face(2),
    joy = face(3),
    grief = face(4)
  assert.ok(fear[FACE.browInner] > 0.8 && fear[FACE.eyes] > 0.8 && fear[FACE.stretch] > 0.5)
  assert.ok(anger[FACE.browInner] < -0.8 && anger[FACE.eyes] < 0 && anger[FACE.stretch] < -0.5)
  assert.ok(joy[FACE.smile] > 0.9 && joy[FACE.cheek] > 0.9)
  assert.ok(grief[FACE.browInner] > 0.5 && grief[FACE.browOuter] < 0 && grief[FACE.smile] < -0.5)
  // A weak mood shows weakly.
  const faint = faceFrom(settle(3, 0.2), 0, 0, new Float32Array(FACE_CHANNELS))
  assert.ok(faint[FACE.smile] > 0.1 && faint[FACE.smile] < 0.3)
  const p = postureFrom(settle(4, 1), new Float32Array(POSTURE_CHANNELS))
  assert.ok(p[POSTURE.slump] > 0.9 && p[POSTURE.hunch] < 0.01)
})

test('blinks are brief, bounded and happen every few seconds', () => {
  let closed = 0,
    blinks = 0,
    was = false
  const dt = 1 / 120
  for (let t = 0; t < 60; t += dt) {
    const b = blinkAt(t, 0.37)
    assert.ok(b >= 0 && b <= 1)
    if (b > 0.5) closed += dt
    if (b > 0.5 && !was) blinks++
    was = b > 0.5
  }
  assert.ok(blinks >= 10 && blinks <= 30, `${blinks} blinks a minute`)
  assert.ok(closed < 4, `eyes shut ${closed.toFixed(1)} s a minute`)
})

test('gestures ease in and out and only speakers move their jaws', () => {
  const g = { kind: 0, side: 1, env: 0, nod: 0, speak: 0 }
  let prev = null
  for (let t = 0; t < 30; t += 1 / 60) {
    gestureAt(t, 0.21, g)
    assert.ok(g.env >= 0 && g.env <= 1 && g.speak >= 0 && g.speak <= 1)
    if (g.kind === 0) assert.equal(g.speak, 0)
    if (prev !== null && g.kind === prev.kind) assert.ok(Math.abs(g.env - prev.env) < 0.2)
    prev = { ...g }
  }
})
