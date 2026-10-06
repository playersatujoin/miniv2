import test from 'node:test'
import assert from 'node:assert/strict'
import { Director, findSubjects, frameShot, smoothDamp, SHOTS } from '../src/game3d/cinematic.ts'
import { FLAG } from '../src/sim/protocol.ts'

const person = (id, x, y, flags = 0, health = 1) => ({ id, rx: x, ry: y, flags, health })
const scene = (over = {}) => ({ people: [], fires: [], corpses: [], villages: [], ...over })
const sub = (key, score, x = 0, z = 0, kind = 'talk') => ({ key, kind, id: 1, partner: null, x, z, radius: 1, score, label: key })

test('fires outrank conversations, fronts gather, labels are Indonesian', () => {
  const s = findSubjects(
    scene({
      people: [person(1, 10, 10, FLAG.talking), person(2, 11, 10, FLAG.talking)],
      fires: [
        { x: 40, y: 40, intensity: 0.9 },
        { x: 41, y: 40, intensity: 0.8 },
        { x: 42, y: 41, intensity: 0.5 },
      ],
    }),
    FLAG,
  )
  const fires = s.filter((x) => x.kind === 'fire')
  const talk = s.find((x) => x.kind === 'talk')
  assert.equal(fires.length, 1, 'neighbouring burning tiles are one front')
  assert.equal(fires[0].label, 'Kebakaran')
  assert.ok(talk && talk.label === 'Dua orang bercakap' && talk.partner !== null)
  assert.ok(fires[0].score > talk.score)
})

test('pairs only form between nearby people talking, each person once', () => {
  const s = findSubjects(
    scene({
      people: [
        person(1, 0, 0, FLAG.talking),
        person(2, 1, 0, FLAG.talking),
        person(3, 1.5, 0, FLAG.talking),
        person(4, 30, 0, FLAG.talking),
      ],
    }),
    FLAG,
  )
  const pairs = s.filter((x) => x.kind === 'talk')
  assert.equal(pairs.length, 1)
  assert.deepEqual([pairs[0].id, pairs[0].partner].sort(), [1, 2])
})

test('a swimmer in trouble, a fight and a fresh body are found and scored above calm ones', () => {
  const s = findSubjects(
    scene({
      people: [
        person(1, 0, 0, FLAG.swimming, 0.2),
        person(2, 5, 5, FLAG.swimming, 1),
        person(3, 9, 9, FLAG.attacking),
        person(4, 9.5, 9, FLAG.hurt),
      ],
      corpses: [
        { id: 7, x: 3, y: 3, seconds: 2, cause: 'drowned' },
        { id: 8, x: 3, y: 3, seconds: 400, cause: 'oldAge' },
      ],
      villages: [
        {
          id: 1,
          name: 'Basosejati',
          x: 20,
          y: 20,
          people: 300,
          hull: [
            [10, 10],
            [30, 10],
            [30, 30],
          ],
          leader: { id: 2 },
        },
      ],
    }),
    FLAG,
  )
  const by = (kind) => s.find((x) => x.kind === kind)
  assert.equal(by('drowning').label, 'Hampir tenggelam')
  assert.ok(by('drowning').score > by('swim').score)
  assert.ok(by('fight').score > 42, 'the hurt beside a fight raise it')
  assert.equal(by('corpse').label, 'Tenggelam')
  assert.equal(s.filter((x) => x.kind === 'corpse').length, 1, 'old bodies are not news')
  assert.equal(by('village').label, 'Desa Basosejati')
  assert.ok(by('village').radius >= 10)
})

test('leaders are named after their village', () => {
  const s = findSubjects(
    scene({
      people: [person(5, 1, 1, FLAG.leader)],
      villages: [{ id: 1, name: 'Retasejati', x: 0, y: 0, people: 10, hull: [], leader: { id: 5 } }],
    }),
    FLAG,
  )
  assert.equal(s.find((x) => x.kind === 'leader').label, 'Pemimpin Desa Retasejati')
})

test('hysteresis: holds a subject for the minimum time, then only a clearly better one takes over', () => {
  const d = new Director({ minHold: 6, margin: 1.3, grace: 2.5 })
  let r = d.update(0, [sub('a', 20)], 0, 0)
  assert.equal(r.subject.key, 'a')
  assert.ok(r.changedSubject)
  // Better, but too soon.
  r = d.update(3, [sub('a', 20), sub('b', 60)], 0, 0)
  assert.equal(r.subject.key, 'a')
  // Long enough, but only slightly better: stays.
  r = d.update(8, [sub('a', 20), sub('b', 24)], 0, 0)
  assert.equal(r.subject.key, 'a')
  // Long enough and clearly better.
  r = d.update(9, [sub('a', 20), sub('b', 60)], 0, 0)
  assert.equal(r.subject.key, 'b')
  assert.ok(r.changedSubject)
})

test('a subject that flickers out briefly is kept; one gone past the grace is replaced at once', () => {
  const d = new Director({ grace: 2.5, minHold: 6 })
  d.update(0, [sub('a', 30)], 0, 0)
  let r = d.update(1, [sub('z', 25)], 0, 0)
  assert.equal(r.subject.key, 'a', 'within the grace')
  r = d.update(2, [sub('a', 30), sub('z', 25)], 0, 0)
  assert.equal(r.subject.key, 'a')
  d.update(3, [sub('z', 25)], 0, 0)
  r = d.update(6, [sub('z', 25)], 0, 0)
  assert.equal(r.subject.key, 'z', 'gone too long: replaced even before the minimum hold')
})

test('recently shown subjects make way for new ones', () => {
  const d = new Director({ minHold: 1, maxHold: 3, margin: 1.1 })
  d.update(0, [sub('a', 30, 0, 0), sub('b', 28, 10, 10)], 0, 0)
  // a held past maxHold: its appeal halves, b takes over.
  let r = d.update(4, [sub('a', 30, 0, 0), sub('b', 28, 10, 10)], 0, 0)
  assert.equal(r.subject.key, 'b')
  // Back to a? Not while it is recent, even though it scores higher.
  r = d.update(10, [sub('a', 30, 0, 0), sub('b', 28, 10, 10)], 10, 10)
  assert.equal(r.subject.key, 'b')
})

test('shots suit the subject and rotate after a while', () => {
  const d = new Director({ shotLength: 10 })
  let r = d.update(0, [{ ...sub('t', 20), kind: 'talk' }], 0, 0)
  assert.equal(r.shot, 'twoShot')
  r = d.update(11, [{ ...sub('t', 20), kind: 'talk' }], 0, 0)
  assert.equal(r.shot, SHOTS.talk[1])
  assert.ok(r.changedShot && !r.changedSubject)
  const v = new Director()
  assert.equal(v.update(0, [{ ...sub('v', 20), kind: 'village' }], 0, 0).shot, 'establishing')
})

test('framing: follow stands behind, orbit keeps its radius, two-shot is square on to the pair', () => {
  const out = {}
  frameShot('follow', { x: 10, y: 1, z: 10, heading: 0, radius: 1 }, 0, 0.7, 1, out)
  assert.ok(out.eyeX < 10 && out.lookX > 10, 'behind, looking ahead')
  assert.ok(out.eyeY > 1)
  const r0 = frameShot('orbit', { x: 0, y: 0, z: 0, heading: 0, radius: 1 }, 0, 0.3, 1, {})
  const r1 = frameShot('orbit', { x: 0, y: 0, z: 0, heading: 0, radius: 1 }, 7, 0.3, 1, {})
  assert.ok(Math.abs(Math.hypot(r0.eyeX, r0.eyeZ) - Math.hypot(r1.eyeX, r1.eyeZ)) < 1e-9)
  assert.ok(Math.hypot(r0.eyeX - r1.eyeX, r0.eyeZ - r1.eyeZ) > 0.1, 'it moves round')
  const two = frameShot('twoShot', { x: 0, y: 0, z: 0, heading: 0, radius: 1, partnerX: 2, partnerZ: 0, partnerY: 0 }, 0, 0.2, 1, {})
  assert.ok(Math.abs(two.lookX - 1) < 1e-9 && Math.abs(two.lookZ) < 1e-9, 'looks at the middle')
  // The eye sits off the line between them (x near the middle, well out in z).
  assert.ok(Math.abs(two.eyeX - 1) < 0.5 && Math.abs(two.eyeZ) > 2)
  // Both people are within a 42° view cone around the line of sight.
  const dir = [two.lookX - two.eyeX, two.lookZ - two.eyeZ]
  for (const [px, pz] of [
    [0, 0],
    [2, 0],
  ]) {
    const v = [px - two.eyeX, pz - two.eyeZ]
    const cos = (dir[0] * v[0] + dir[1] * v[1]) / (Math.hypot(...dir) * Math.hypot(...v))
    assert.ok(Math.acos(cos) < (21 * Math.PI) / 180)
  }
  const wide = frameShot('establishing', { x: 0, y: 0, z: 0, heading: 0, radius: 30 }, 0, 0, 1, {})
  assert.ok(Math.hypot(wide.eyeX, wide.eyeZ) >= 30 && wide.eyeY > 20)
})

test('smoothDamp settles on its target without overshooting', () => {
  const s = new Float64Array([0, 0])
  let max = 0
  for (let i = 0; i < 300; i++) {
    smoothDamp(s, 0, 10, 0.8, 1 / 60)
    max = Math.max(max, s[0])
  }
  assert.ok(max <= 10 + 1e-9)
  assert.ok(Math.abs(s[0] - 10) < 1e-3)
})
