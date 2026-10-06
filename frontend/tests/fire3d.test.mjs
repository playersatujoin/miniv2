import test from 'node:test'
import assert from 'node:assert/strict'
import { FirePlanner, FIRE_REACH, layoutFlames, due, flicker, Particles, SMOKE, EMBERS } from '../src/game3d/fire3d.ts'
import { WindTracker, Lightning } from '../src/game3d/weather3d.ts'
import { lightCycle } from '../src/game3d/timecycle.ts'

const budget = { flames: 200, smoke: 300, embers: 60, lights: 3 }

test('a big fire front never exceeds the budget, and every effect is counted', () => {
  const fires = []
  for (let i = 0; i < 600; i++) fires.push({ x: i % 40, y: Math.floor(i / 40), intensity: 0.3 + ((i * 37) % 70) / 100 })
  const p = new FirePlanner().plan(fires, 20, 7, budget)
  const sum = (a) => a.slice(0, fires.length).reduce((s, v) => s + v, 0)
  assert.ok(sum(p.flames) <= budget.flames && sum(p.smoke) <= budget.smoke && sum(p.embers) <= budget.embers)
  assert.equal(sum(p.flames), p.total.flames)
  assert.equal(p.total.flames, budget.flames, 'a big fire uses the whole flame budget')
  assert.ok(p.lights.length <= budget.lights)
})

test('nearer and stronger fires get more; out of reach and spent ones get nothing', () => {
  const fires = [
    { x: 10, y: 10, intensity: 1 },
    { x: 60, y: 10, intensity: 1 },
    { x: 11, y: 10, intensity: 0.2 },
    { x: 10 + FIRE_REACH + 5, y: 10, intensity: 1 },
    { x: 12, y: 12, intensity: 0 },
  ]
  const p = new FirePlanner().plan(fires, 10, 10, { flames: 30, smoke: 30, embers: 10, lights: 2 })
  assert.ok(p.flames[0] >= p.flames[1], 'nearer')
  assert.ok(p.flames[0] > p.flames[2], 'stronger')
  assert.equal(p.flames[3], 0)
  assert.equal(p.flames[4], 0)
  assert.equal(p.smoke[3] + p.embers[3], 0)
  assert.ok(p.flames[1] >= 2, 'every fire in reach gets a minimum')
})

test('a tight budget goes to the most important fires first', () => {
  const fires = [
    { x: 50, y: 50, intensity: 0.5 },
    { x: 0, y: 0, intensity: 1 },
    { x: 80, y: 80, intensity: 0.5 },
  ]
  const p = new FirePlanner().plan(fires, 0, 0, { flames: 2, smoke: 1, embers: 0, lights: 1 })
  assert.equal(p.flames[1], 2)
  assert.equal(p.flames[0] + p.flames[2], 0)
  assert.equal(p.smoke[1], 1)
  assert.deepEqual(p.lights, [1])
})

test('lights go to the strongest fires, spread apart', () => {
  const fires = [
    { x: 0, y: 0, intensity: 1 },
    { x: 1, y: 0, intensity: 1 },
    { x: 2, y: 0, intensity: 1 },
    { x: 20, y: 0, intensity: 0.8 },
    { x: 40, y: 0, intensity: 0.6 },
  ]
  const p = new FirePlanner().plan(fires, 0, 0, { flames: 50, smoke: 50, embers: 0, lights: 3 })
  assert.equal(p.lights.length, 3)
  for (const a of p.lights) for (const b of p.lights) if (a !== b) assert.ok(Math.abs(fires[a].x - fires[b].x) >= 5)
  assert.equal(p.lights[0], 0)
})

test('planning again reuses its buffers and handles an empty sky', () => {
  const planner = new FirePlanner()
  planner.plan([{ x: 1, y: 1, intensity: 1 }], 0, 0, budget)
  const buf = planner.flames
  planner.plan([{ x: 2, y: 2, intensity: 0.5 }], 0, 0, budget)
  assert.equal(planner.flames, buf)
  planner.plan([], 0, 0, budget)
  assert.equal(planner.total.flames, 0)
  assert.equal(planner.lights.length, 0)
})

test('flames are laid out within their tile, no more than the buffer holds', () => {
  const fires = [
    { x: 3, y: 4, intensity: 1 },
    { x: 9, y: 9, intensity: 0.5 },
  ]
  const p = new FirePlanner().plan(fires, 0, 0, budget)
  const out = new Float32Array(4 * 8)
  const n = layoutFlames(fires, p, out)
  assert.ok(n <= 8)
  for (let k = 0; k < n; k++) {
    const x = out[k * 4]
    const z = out[k * 4 + 1]
    const inA = x >= 3 && x <= 4 && z >= 4 && z <= 5
    const inB = x >= 9 && x <= 10 && z >= 9 && z <= 10
    assert.ok(inA || inB)
  }
})

test('emission without state: a steady rate adds up over time', () => {
  let n = 0
  for (let t = 1 / 60; t <= 10; t += 1 / 60) n += due(t, 1 / 60, 3, 0.37)
  assert.ok(Math.abs(n - 30) <= 1)
  assert.equal(due(1, 0, 3, 0), 0)
  for (let t = 0; t < 5; t += 0.1) {
    const f = flicker(t, 0.3)
    assert.ok(f >= 0.6 && f <= 1)
  }
})

test('smoke rises and drifts downwind, and the pool never overflows', () => {
  const p = new Particles(50)
  for (let i = 0; i < 80; i++) p.emit(0, 0, 0, 1, 6, 0.3, 0.5, i / 80)
  assert.equal(p.count, 50)
  for (let i = 0; i < 120; i++) p.step(1 / 30, 0, 1, 0.8, SMOKE, i / 30)
  let mx = 0,
    mz = 0,
    my = 0
  for (let i = 0; i < p.count; i++) {
    mx += p.x[i]
    mz += p.z[i]
    my += p.y[i]
  }
  assert.ok(my / p.count > 1, 'rises')
  assert.ok(mz / p.count > 3 && Math.abs(mx / p.count) < 1, 'carried towards +z')
  for (let i = 0; i < 400; i++) p.step(1 / 30, 0, 1, 0.8, SMOKE, i / 30)
  assert.equal(p.count, 0, 'spent particles are removed')
})

test('a turning wind swings a smoke column round; embers fall back', () => {
  const p = new Particles(10)
  p.emit(0, 0, 0, 1, 10, 0.3, 0.5, 0.1)
  for (let i = 0; i < 60; i++) p.step(1 / 30, 1, 0, 1, SMOKE, 0)
  const vx = p.vx[0]
  for (let i = 0; i < 90; i++) p.step(1 / 30, -1, 0, 1, SMOKE, 0)
  assert.ok(vx > 0 && p.vx[0] < 0)
  const e = new Particles(1)
  e.emit(0, 0, 0, 2, 5, 0.05, 1, 0.5)
  let top = 0
  for (let i = 0; i < 120; i++) {
    e.step(1 / 30, 1, 0, 0.5, EMBERS, 0)
    top = Math.max(top, e.y[0])
  }
  assert.ok(top > e.y[0], 'it rose, then fell back')
})

test('the drawn wind turns towards the stream at a bounded rate and its gusts stay bounded', () => {
  const w = new WindTracker(3)
  w.update(1 / 60, { dir: 0, strength: 0.4, storm: false })
  assert.equal(w.dir, 0, 'the first sample is taken as it is')
  w.update(1 / 60, { dir: Math.PI / 2, strength: 0.4, storm: false })
  assert.ok(w.dir > 0 && w.dir < 0.02, 'no snapping')
  for (let i = 0; i < 60 * 30; i++) w.update(1 / 60, { dir: Math.PI / 2, strength: 0.9, storm: true })
  assert.ok(Math.abs(w.dir - Math.PI / 2) < 0.01)
  assert.ok(Math.abs(w.dirX) < 0.02 && w.dirZ > 0.99)
  let gusty = 0
  for (let i = 0; i < 60 * 60; i++) {
    w.update(1 / 60, { dir: Math.PI / 2, strength: 0.9, storm: true })
    assert.ok(w.blowing >= 0 && w.blowing <= 1.4)
    if (w.gust > 0.05) gusty++
  }
  assert.ok(gusty > 0, 'a gale gusts')
  // The short way round across ±π.
  const v = new WindTracker(1)
  v.update(1 / 60, { dir: 3, strength: 0.5, storm: false })
  v.update(1, { dir: -3, strength: 0.5, storm: false })
  assert.ok(v.dir > 3 || v.dir < -3 + 2 * Math.PI, 'turns through π, not back through 0')
})

test('lightning: only in storms, never more than two flashes a second, off when not allowed', () => {
  const l = new Lightning(5)
  for (let i = 0; i < 600; i++) assert.equal(l.update(1 / 60, 0, true), 0)
  let rising = 0
  let last = 0
  let strikes = 0
  const onsets = []
  for (let i = 0; i < 60 * 120; i++) {
    const f = l.update(1 / 60, 1, true)
    assert.ok(f >= 0 && f <= 1)
    if (f > 0.2 && last <= 0.2) onsets.push(i / 60)
    last = f
    strikes = l.strikes
    rising++
  }
  assert.ok(strikes >= 8, 'a two-minute storm strikes several times')
  for (let k = 2; k < onsets.length; k++) assert.ok(onsets[k] - onsets[k - 2] > 1, 'at most two flashes in any second')
  const m = new Lightning(5)
  for (let i = 0; i < 60 * 30; i++) assert.equal(m.update(1 / 60, 1, false), 0)
  void rising
})

test('a storm darkens the light and a lightning flash brightens it, without changing calm weather', () => {
  const calm = lightCycle(1, 0.3)
  assert.deepEqual(lightCycle(1, 0.3, 0, 0), calm)
  const storm = lightCycle(1, 0.95, 1)
  assert.ok(storm.sun < lightCycle(1, 0.95).sun && storm.ambient < calm.ambient && storm.exposure < calm.exposure)
  const flash = lightCycle(1, 0.95, 1, 1)
  assert.ok(flash.ambient > storm.ambient && flash.exposure > storm.exposure)
  assert.deepEqual(lightCycle(1, 0.3, 5, -2), lightCycle(1, 0.3, 1, 0), 'inputs are clamped')
})
