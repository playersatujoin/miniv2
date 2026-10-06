import test from 'node:test'
import assert from 'node:assert/strict'
import { registerHooks } from 'node:module'
import {
  adjacentMark,
  bucketMarks,
  evictable,
  formatClockSpan,
  frameAtTick,
  frameDelay,
  indexToPosition,
  nearestMark,
  positionToIndex,
  prefetchWindow,
  rankMarks,
  secondsBehind,
  tickToPosition,
} from '../src/components/sim/replayTimeline.ts'
import { isViolent, kindStyle } from '../src/components/sim/eventKinds.ts'
import { MOOD_COLOR, MOOD_COLORS, MOOD_KEYS, dominantMood, temperWord } from '../src/components/sim/moods.ts'
import {
  attitudeLabel,
  hullTop,
  leaderTrust,
  sortAttitudes,
  sortVillages,
  trustWord,
  villageCamera,
  yearOf,
  yearsBetween,
} from '../src/components/sim/villageList.ts'
import {
  angleDelta,
  easeAngle,
  hasServerWind,
  rainSlant,
  smokeDrift,
  windArrow,
  windCompass,
  windVector,
  windWord,
} from '../src/game/wind.ts'

// The app's modules import each other without extensions (Vite resolves them); let Node do the same
// for the replay player, which imports the timeline maths.
registerHooks({
  resolve(specifier, context, next) {
    try {
      return next(specifier, context)
    } catch (err) {
      if (/^\.\.?\//.test(specifier) && !/\.[cm]?[jt]sx?$/.test(specifier)) return next(`${specifier}.ts`, context)
      throw err
    }
  },
})
const { ReplayPlayer } = await import('../src/components/sim/replayPlayer.ts')

const close = (a, b, eps = 1e-9) => assert.ok(Math.abs(a - b) <= eps, `${a} ≉ ${b}`)

// Ten frames, ten ticks apart, with a pause (two frames on the same tick) in the middle.
const frames = [
  [100, 5],
  [110, 5.5],
  [120, 6],
  [130, 6.5],
  [130, 6.5],
  [140, 7],
  [150, 7.5],
  [160, 8],
  [170, 8.5],
  [180, 9],
]

test('replay: tick ↔ timeline position', () => {
  assert.equal(frameAtTick(frames, 99), 0)
  assert.equal(frameAtTick(frames, 100), 0)
  assert.equal(frameAtTick(frames, 125), 2)
  // A paused run maps to its first frame.
  assert.equal(frameAtTick(frames, 130), 3)
  assert.equal(frameAtTick(frames, 135), 3)
  assert.equal(frameAtTick(frames, 999), 9)
  assert.equal(frameAtTick([], 5), -1)

  assert.equal(tickToPosition(frames, 0), 0)
  assert.equal(tickToPosition(frames, 180), 1)
  close(tickToPosition(frames, 110), 1 / 9)
  close(tickToPosition(frames, 115), 1.5 / 9)
  // Between the paused run and the next frame: interpolate over the whole gap.
  close(tickToPosition(frames, 135), 4 / 9)
  for (let i = 0; i < frames.length; i++) assert.equal(positionToIndex(frames, indexToPosition(frames, i)), i)
  assert.equal(positionToIndex(frames, -1), 0)
  assert.equal(positionToIndex(frames, 2), 9)
  assert.equal(tickToPosition([[5, 1]], 5), 1)
})

test('replay: nearest marker, buckets and jumps', () => {
  const marks = [
    { tick: 110, importance: 1, text: 'a' },
    { tick: 112, importance: 3, text: 'b' },
    { tick: 160, importance: 2, text: 'c' },
    { tick: 175, importance: 3, text: 'd' },
  ]
  // Near tick 110 both a and b are within tolerance: the important one wins.
  assert.equal(nearestMark(frames, marks, 1 / 9, 0.03).text, 'b')
  assert.equal(nearestMark(frames, marks, 7 / 9, 0.01).text, 'c')
  assert.equal(nearestMark(frames, marks, 0.5, 0.01), null)

  const buckets = bucketMarks(frames, marks, 3)
  assert.equal(buckets.length, 2)
  assert.deepEqual(
    buckets.map((b) => [b.mark.text, b.count]),
    [
      ['b', 2],
      ['d', 2],
    ],
  )

  assert.equal(adjacentMark(marks, 112, 1).text, 'd')
  assert.equal(adjacentMark(marks, 175, -1).text, 'b')
  assert.equal(adjacentMark(marks, 175, 1), null)
  assert.equal(adjacentMark(marks, 150, 1, 2).text, 'c')
})

test('replay: pacing and prefetch window', () => {
  assert.equal(frameDelay(1), 500)
  assert.equal(frameDelay(4), 125)
  assert.equal(frameDelay(0.5), 1000)
  assert.equal(secondsBehind(frames, 9), 0)
  assert.equal(secondsBehind(frames, 3), 3)
  assert.equal(formatClockSpan(65), '1:05')

  const none = () => false
  // 3 s ahead at 1× is 6 frames (+ the current one), nearest first, then one behind.
  const at1 = prefetchWindow(10, 1, 100, none)
  assert.deepEqual(at1.slice(0, 3), [10, 11, 12])
  assert.equal(at1.at(-1), 9)
  // Faster playback reads further ahead.
  assert.ok(prefetchWindow(10, 4, 100, none).length > at1.length)
  // Bounded by the recording and by max.
  assert.deepEqual(prefetchWindow(98, 4, 100, none), [98, 99, 97])
  assert.ok(prefetchWindow(0, 4, 1000, none, 30, 8).length <= 8)
  // Frames already held are skipped.
  assert.deepEqual(
    prefetchWindow(10, 1, 100, (i) => i !== 12 && i !== 9),
    [12, 9],
  )
  assert.deepEqual(prefetchWindow(-1, 1, 100, none), [])

  assert.deepEqual(evictable([0, 5, 20, 40, 80], 20, 12, 48), [0, 5, 80])
})

const village = (id, name, people, houses, extra = {}) => ({
  id,
  name,
  hue: 100,
  hull: [],
  x: 0,
  y: 0,
  people,
  houses,
  leader: null,
  founded: 0,
  ...extra,
})

test('villages: sorting, trust and attitudes', () => {
  const list = [village(1, 'Wates', 20, 3), village(2, 'Arjosari', 50, 9), village(3, 'Bandar', 20, 5), village(4, 'Asem', 20, 5)]
  assert.deepEqual(
    sortVillages(list).map((v) => v.name),
    ['Arjosari', 'Asem', 'Bandar', 'Wates'],
  )
  assert.deepEqual(sortVillages(null), [])
  // The input is not reordered.
  assert.equal(list[0].name, 'Wates')

  const v = village(1, 'Wates', 6, 2, {
    leader: { id: 9, name: 'Ratu' },
    residents: [
      { id: 9, name: 'Ratu', house: 1, adult: true, trust: 0, leader: true },
      { id: 1, name: 'A', house: 1, adult: true, trust: 0.8 },
      { id: 2, name: 'B', house: 1, adult: true, trust: 0.2 },
      { id: 3, name: 'C', house: 2, adult: true, trust: -0.4 },
      { id: 4, name: 'D', house: 2, adult: true, trust: 0 },
      { id: 5, name: 'E', house: 2, adult: false, trust: 0.9 },
    ],
  })
  const t = leaderTrust(v)
  assert.equal(t.adults, 4)
  assert.equal(t.opinions, 3)
  assert.equal(t.trusting, 2)
  assert.equal(t.distrusting, 1)
  close(t.mean, (0.8 + 0.2 - 0.4) / 3)
  assert.equal(leaderTrust(village(2, 'X', 0, 0)).mean, null)
  assert.equal(trustWord(0.7), 'sangat percaya')
  assert.equal(trustWord(0), 'netral')
  assert.equal(trustWord(-0.3), 'curiga')

  const att = sortAttitudes([
    { id: 2, name: 'B', trust: 0, known: 0, attitude: 'asing' },
    { id: 3, name: 'C', trust: -0.5, known: 4, attitude: 'benci' },
    { id: 4, name: 'D', trust: 0.4, known: 2, attitude: 'suka' },
  ])
  assert.deepEqual(
    att.map((a) => a.name),
    ['D', 'C', 'B'],
  )
  assert.equal(attitudeLabel('tidak suka'), 'Tidak suka')
  assert.equal(attitudeLabel('asing'), 'Belum saling kenal')
  assert.equal(attitudeLabel('baru'), 'Baru')
})

test('villages: years and camera', () => {
  assert.equal(yearOf(0, 8), 1)
  assert.equal(yearOf(27828, 8), 3479)
  assert.equal(yearsBetween(27828, 27908, 8), 10)
  assert.equal(yearsBetween(30, 10, 8), 0)

  const cam = villageCamera({
    x: 0,
    y: 0,
    hull: [
      [50, 80],
      [80, 80],
      [80, 130],
      [50, 130],
    ],
  })
  assert.equal(cam.x, 65)
  assert.equal(cam.y, 105)
  // Tall: 50 tiles high needs more across than its 30 wide.
  assert.ok(cam.across >= 50 * 1.3)
  assert.deepEqual(villageCamera({ x: 3, y: 4, hull: [] }), { x: 3, y: 4, across: 28 })
  assert.deepEqual(
    hullTop({
      x: 0,
      y: 0,
      hull: [
        [1, 9],
        [4, 2],
        [7, 5],
      ],
    }),
    [4, 2],
  )
})

test('wind: streamed direction to a 2D vector', () => {
  const east = windVector(0, 1)
  close(east.x, 1)
  close(east.y, 0)
  // π/2 blows down the map (+y; +z in 3D).
  const south = windVector(Math.PI / 2, 0.5)
  close(south.x, 0)
  close(south.y, 0.5)
  assert.equal(windVector(1, 3).strength, 1)
  assert.equal(windVector(NaN, NaN).strength, 0)

  assert.equal(hasServerWind({ windDir: 0, wind: 0 }), false)
  assert.equal(hasServerWind(null), false)
  assert.equal(hasServerWind({ windDir: 2.1, wind: 0 }), true)
  assert.equal(hasServerWind({ windDir: 0, wind: 0.2 }), true)

  // The short way round across ±π.
  close(angleDelta(3, -3), 2 * Math.PI - 6)
  close(angleDelta(-3, 3), 6 - 2 * Math.PI)
  close(easeAngle(3, -3, 1), 3 + (2 * Math.PI - 6))
  close(easeAngle(0, 1, 0.5), 0.5)

  // Smoke bends with the wind, more with age, and not at all at the chimney.
  const d0 = smokeDrift(east, 0, 30)
  const d1 = smokeDrift(east, 1, 30)
  assert.equal(d0.dx, 0)
  assert.ok(d1.dx > 0 && Math.abs(d1.dy) < 1e-9)
  assert.ok(smokeDrift(windVector(Math.PI, 1), 1, 30).dx < 0)
  // Rain leans with the wind.
  assert.ok(rainSlant(windVector(0, 1)) > rainSlant(windVector(Math.PI, 1)))
})

test('wind: compass words and arrows (north is up the map)', () => {
  assert.equal(windCompass(0), 'timur')
  assert.equal(windCompass(Math.PI / 2), 'selatan')
  assert.equal(windCompass(-Math.PI / 2), 'utara')
  assert.equal(windCompass(Math.PI), 'barat')
  assert.equal(windCompass(-2.79), 'barat')
  assert.equal(windCompass(Math.PI / 4), 'tenggara')
  assert.equal(windCompass(-Math.PI / 4), 'timur laut')
  assert.equal(windCompass(7 * Math.PI), 'barat')
  assert.equal(windCompass(NaN), 'timur')
  assert.equal(windArrow(Math.PI / 2), '↓')
  assert.equal(windArrow(-Math.PI / 4), '↗')
  assert.equal(windWord(0.05), 'tenang')
  assert.equal(windWord(0.18), 'semilir')
  assert.equal(windWord(0.9), 'sangat kencang')
  assert.equal(windWord(0.2, true), 'badai')
})

test('moods: colours, the dominant mood and temperament words', () => {
  assert.deepEqual(MOOD_KEYS, ['fear', 'anger', 'joy', 'grief'])
  // By MOOD code: 0 calm has none, then fear … grief.
  assert.equal(MOOD_COLORS.length, 5)
  assert.equal(MOOD_COLORS[0], '')
  MOOD_KEYS.forEach((k, i) => assert.equal(MOOD_COLORS[i + 1], MOOD_COLOR[k]))
  assert.equal(new Set(MOOD_COLORS.slice(1)).size, 4)

  assert.equal(dominantMood({ fear: 0.9, anger: 0, joy: 0, grief: 0, dominant: 'joy' }), 'joy')
  // An older server that says calm: the strongest mood above the shown threshold.
  assert.equal(dominantMood({ fear: 0.2, anger: 0.5, joy: 0.4, grief: 0, dominant: '' }), 'anger')
  assert.equal(dominantMood({ fear: 0.2, anger: 0.1, joy: 0.29, grief: 0, dominant: '' }), '')
  assert.equal(dominantMood(null), '')

  assert.equal(temperWord(1, 'lambat', 'cepat'), 'biasa')
  assert.equal(temperWord(1.08, 'lambat', 'cepat'), 'cepat')
  assert.equal(temperWord(1.2, 'lambat', 'cepat'), 'sangat cepat')
  assert.equal(temperWord(0.88, 'murung', 'riang'), 'murung')
  assert.equal(temperWord(undefined, 'a', 'b'), '—')
})

test('events: kinds of engine adaptation II and violence', () => {
  for (const k of ['village', 'fire', 'trade', 'rumor']) {
    const st = kindStyle(k)
    assert.ok(st.label && st.color.startsWith('#'), k)
    assert.equal(st.shape, 'diamond')
  }
  assert.deepEqual(kindStyle('gempa'), { label: 'gempa', color: '#898781' })
  assert.equal(isViolent({ kind: 'death', text: 'Arato ♂ dibunuh oleh Budi' }), true)
  assert.equal(isViolent({ kind: 'death', text: 'Arato ♂ mati kelaparan' }), false)
  assert.equal(isViolent({ kind: 'crime', text: 'x menyerang y', violent: false }), false)
})

test('replay: common kinds step down a level on the timeline', () => {
  const marks = [
    ...Array.from({ length: 30 }, (_, i) => ({ tick: i, kind: 'death', importance: 3, text: 'd' })),
    ...Array.from({ length: 30 }, (_, i) => ({ tick: i, kind: 'birth', importance: 1, text: 'b' })),
    { tick: 5, kind: 'village', importance: 3, text: 'v' },
    { tick: 6, kind: 'trade', importance: 2, text: 't' },
  ]
  // 60 recorded seconds: more than one every 15 s (4) is common.
  const ranked = rankMarks(marks, 60)
  assert.equal(ranked.find((m) => m.kind === 'death').importance, 2)
  assert.equal(ranked.find((m) => m.kind === 'death').base, 3)
  assert.equal(ranked.find((m) => m.kind === 'birth').importance, 1)
  assert.equal(ranked.find((m) => m.kind === 'village').importance, 3)
  assert.equal(ranked.find((m) => m.kind === 'trade').importance, 2)
  // A long recording makes the same deaths rare enough to keep their level.
  assert.equal(rankMarks(marks, 600).find((m) => m.kind === 'death').importance, 3)
})

/** A clock whose timers fire only when the test says so. */
function fakeClock() {
  const timers = []
  return {
    timers,
    setTimeout: (fn, ms) => {
      const t = { fn, ms, done: false }
      timers.push(t)
      return t
    },
    clearTimeout: (t) => {
      if (t) t.done = true
    },
    /** Fires the earliest pending timer; returns its delay or -1. */
    fire() {
      const t = timers.find((x) => !x.done)
      if (!t) return -1
      t.done = true
      t.fn()
      return t.ms
    },
  }
}

const settle = () => new Promise((r) => setImmediate(r))

function fakeServer(ticks, versionOf = () => 7) {
  const asked = []
  const buildings = []
  return {
    asked,
    buildings,
    load: {
      frame: async (tick) => {
        asked.push(tick)
        return {
          frame: { tick, time: tick / 20, creatures: [], animals: [], weather: null, corpses: [], fires: [] },
          structures: versionOf(tick),
        }
      },
      structures: async (v) => {
        buildings.push(v)
        return { version: v, structures: [{ id: v, kind: 'gubuk', x: 1, y: 1, ownerId: 0, ownerName: '', level: 1, hue: 0 }] }
      },
    },
  }
}

test('replay player: seek, play at the recorded pace, step, end', async () => {
  const ticks = Array.from({ length: 30 }, (_, i) => 1000 + i * 10)
  const server = fakeServer(ticks, (t) => (t < 1150 ? 7 : 8))
  const shown = []
  const structures = []
  const clock = fakeClock()
  const p = new ReplayPlayer(
    server.load,
    { show: (f, interval, jump) => shown.push({ tick: f.tick, interval, jump }), structures: (list, v) => structures.push(v) },
    clock,
  )
  let changes = 0
  p.subscribe(() => changes++)
  p.setIndex({ frames: ticks.map((t) => [t, t / 20]) })

  p.seek(5)
  assert.equal(p.getState().waiting, true)
  await settle()
  assert.deepEqual(shown.at(-1), { tick: 1050, interval: 500, jump: true })
  assert.equal(p.getState().index, 5)
  assert.equal(p.getState().waiting, false)
  assert.deepEqual(structures, [7])
  // It read ahead of the cursor.
  assert.ok(server.asked.includes(1060) && server.asked.includes(1070))
  assert.ok(changes > 0)

  // Playing shows the next frame after one frame's delay (0.5 s at 1×), gliding (no jump).
  p.play()
  assert.equal(clock.fire(), 500)
  await settle()
  assert.deepEqual(shown.at(-1), { tick: 1060, interval: 500, jump: false })
  assert.equal(p.getState().index, 6)

  // 4× is four times quicker, and the glide matches.
  p.setSpeed(4)
  assert.equal(clock.fire(), 500) // the timer already set
  await settle()
  assert.equal(clock.fire(), 125)
  await settle()
  assert.deepEqual(shown.at(-1), { tick: 1080, interval: 125, jump: false })

  // Pausing stops the loop; stepping pauses and moves one frame.
  p.pause()
  assert.equal(clock.fire(), -1)
  p.step(-1)
  await settle()
  assert.equal(p.getState().index, 7)
  assert.equal(p.getState().playing, false)
  assert.equal(shown.at(-1).jump, true)

  // A new building version is fetched once when its frames show.
  p.seek(20)
  await settle()
  assert.deepEqual(structures, [7, 8])
  p.seek(21)
  await settle()
  assert.deepEqual(structures, [7, 8])

  // Playing into the newest frame pauses there.
  p.seek(28)
  await settle()
  p.play()
  clock.fire()
  await settle()
  assert.equal(p.getState().index, 29)
  clock.fire()
  await settle()
  assert.equal(p.getState().atEnd, true)
  assert.equal(p.getState().playing, false)
  // Play again from the end starts over at the oldest frame.
  p.play()
  await settle()
  assert.equal(p.getState().index, 0)
  assert.equal(shown.at(-1).tick, 1000)
  p.destroy()
})

test('replay player: the window slides, late frames from old seeks are dropped', async () => {
  const ticks = Array.from({ length: 20 }, (_, i) => 100 + i)
  const resolvers = new Map()
  const shown = []
  const p = new ReplayPlayer(
    {
      frame: (tick) =>
        new Promise((r) =>
          resolvers.set(tick, () =>
            r({ frame: { tick, time: tick, creatures: [], animals: [], weather: null, corpses: [], fires: [] }, structures: 0 }),
          ),
        ),
      structures: async () => ({ version: 0, structures: [] }),
    },
    { show: (f) => shown.push(f.tick), structures: () => {} },
    fakeClock(),
  )
  p.setIndex({ frames: ticks.map((t) => [t, t]) })
  p.seek(3)
  p.seek(10)
  resolvers.get(103)()
  await settle()
  assert.deepEqual(shown, [])
  resolvers.get(110)()
  await settle()
  assert.deepEqual(shown, [110])

  // Five older frames drop out and five new ones arrive: the cursor stays on its tick.
  p.setIndex({ frames: [...ticks.slice(5), 120, 121, 122, 123, 124].map((t) => [t, t]) })
  assert.equal(p.getState().index, 5)
  // A cursor older than the window moves to the oldest frame left.
  p.setIndex({ frames: [130, 131].map((t) => [t, t]) })
  assert.equal(p.getState().index, 0)
  p.destroy()
})

test('replay player: a failed frame pauses with the error', async () => {
  const p = new ReplayPlayer(
    { frame: async () => Promise.reject(new Error('replay frame: 404')), structures: async () => ({ version: 0, structures: [] }) },
    { show: () => {}, structures: () => {} },
    fakeClock(),
  )
  p.setIndex({
    frames: [
      [1, 1],
      [2, 2],
    ],
  })
  p.play()
  await settle()
  await settle()
  assert.equal(p.getState().playing, false)
  assert.match(p.getState().error, /404/)
  p.destroy()
})
