// The islanders in 3D. Every person is one instance of a single stylised body
// that carries both outfits (a woman's sarong, kemben and long hair; a man's
// cawat, ikat kepala and short hair) and a few props (bakul, tombak, palu,
// tugal, karung, dayung, a bundle to trade), switched per person with a bit mask.
// A small skeleton of 13 bones is posed on the CPU each frame from what the
// person is doing (walking, running, wading, swimming, paddling, climbing,
// gathering, planting, building, resting, eating, teaching, talking, trading…)
// and how they feel, and its matrices go to the GPU in a texture, one row per
// person, followed by their colours, face and hair sway: "skinned instancing",
// the way games draw crowds of animated characters.
//
// Close up, faces show moods (expressions.ts), blows make people flinch, a slip
// knocks them down as a ragdoll that gets back up (ragdoll.ts), the bodies of the
// dead fall and lie where they died, and the few nearest people's sarongs and kains
// are simulated cloth blowing in the wind (cloth.ts). Far away none of that costs
// anything.
//
// Units are tiles (before the world's person scale); the body faces +z with its
// left hand at +x, feet on y = 0.

import * as THREE from 'three'
import { mergeGeometries } from 'three/examples/jsm/utils/BufferGeometryUtils.js'
import type { LiveCreature } from '../sim/live'
import { FLAG, type CorpseFrame } from '../sim/protocol'
import { bodyProportions, solveLeg } from './anatomy'
import { createCloth, resetCloth, skirtShape, stepCloth, type Cloth, type ClothForces } from './cloth'
import { RAFT_DECK } from './water3d'
import {
  FACE_CHANNELS,
  GESTURE,
  MOODS,
  POSTURE,
  POSTURE_CHANNELS,
  blendFactor,
  blinkAt,
  faceFrom,
  gestureAt,
  postureFrom,
  stepMood,
  type Gesture,
} from './expressions'
import {
  FLINCH_TIME,
  RAG_N,
  RAG_POSE,
  RAG_POSE_SIZE,
  RAG_REST,
  createRagdoll,
  flinch,
  guard,
  ragdollPose,
  settleRagdoll,
  stepRagdoll,
  type BodyGround,
  type Ragdoll,
} from './ragdoll'

// --- Skeleton -----------------------------------------------------------------------

const HIPS = 0
const CHEST = 1
const HEAD = 2
const ARM_L = 3
const FORE_L = 4
const ARM_R = 5
const FORE_R = 6
const THIGH_L = 7
const SHIN_L = 8
const THIGH_R = 9
const SHIN_R = 10
const BELLY = 11
/** The sarong or kain: tilts with the legs' average, so it drapes the lap when sitting but not when walking. */
const SKIRT = 12
const NB = 13
const PARENT = [-1, HIPS, CHEST, CHEST, ARM_L, CHEST, ARM_R, HIPS, THIGH_L, HIPS, THIGH_R, HIPS, HIPS]

/** Joint positions in the rest pose (standing, arms hanging). ragdoll.ts RAG_REST uses the same joints. */
const PIVOT: [number, number, number][] = [
  [0, 0.27, 0], // hips
  [0, 0.3, 0], // chest (bends at the waist)
  [0, 0.43, 0], // head (neck)
  [0.078, 0.4, 0], // left shoulder
  [0.083, 0.315, 0], // left elbow
  [-0.078, 0.4, 0],
  [-0.083, 0.315, 0],
  [0.038, 0.25, 0], // left hip joint
  [0.038, 0.135, 0], // left knee
  [-0.038, 0.25, 0],
  [-0.038, 0.135, 0],
  [0, 0.285, 0.03], // belly
  [0, 0.27, 0], // sarong
]

/**
 * Pose vector: root lift and shift, an XYZ rotation per bone, then head and belly scale, the root's
 * sideways shift and how much the fixed skirt drapes onto the legs (1 lying down, 0 standing).
 */
const P_ROOT_Y = 0
const P_ROOT_Z = 1
const rot = (b: number) => 2 + b * 3
const P_HEAD_SCALE = 2 + NB * 3
const P_BELLY = P_HEAD_SCALE + 1
const P_ROOT_X = P_BELLY + 1
const P_DRAPE = P_ROOT_X + 1
const POSE_SIZE = P_DRAPE + 1

// --- Parts ----------------------------------------------------------------------------

/** What a vertex takes its colour from: its own, or the person's skin, hair, clothes or the clothes' trim. */
const FIXED = 0
const SKIN = 1
const HAIR = 2
const CLOTH = 3
const TRIM = 4

/** Face features (and hair that sways), moved per person in the vertex shader. */
const F_BROW = 1
const F_LID = 2
const F_EYE = 3
const F_IRIS = 4
const F_MOUTH = 5
const F_CHEEK = 6
const F_CHIN = 7
const F_HAIR = 8
const F_TAIL = 9
const F_SKIRT = 10

/** Parts shown only for some people (0 = everyone). */
export const SHOW = {
  male: 1,
  female: 2,
  /** Kepala keluarga: a band of gold with feathers. */
  chief: 4,
  /** A frangipani behind the ear. */
  flower: 8,
  /** A bakul of fruit carried on the head. */
  basket: 16,
  /** A karung slung on the back. */
  sack: 32,
  spear: 64,
  hammer: 128,
  /** A tugal, the planting stick. */
  stick: 256,
  /** A woman's sarong as a fixed shape (hidden while it is simulated cloth). */
  sarong: 512,
  /** A man's kain as a fixed shape (hidden while it is simulated cloth). */
  kain: 1024,
  /** A bundle held out in a trade. */
  bundle: 2048,
  /** A dayung, paddling a raft. */
  paddle: 4096,
} as const

type Part = { geo: THREE.BufferGeometry; bone: number; region: number; color?: string; show?: number; feat?: number }

const v3 = (x: number, y: number, z: number) => new THREE.Vector3(x, y, z)

/** Segment scale while building a body: 1 for close up, lower for the far-away version. */
let detail = 1
const seg = (n: number, min = 3) => Math.max(min, Math.round(n * detail))

/** An ellipsoid, optionally leaning sideways (about z) by `lean`. */
function ball(c: THREE.Vector3, rx: number, ry: number, rz: number, w = 8, h = 6, lean = 0) {
  const g = new THREE.SphereGeometry(1, seg(w), seg(h, 2))
  g.scale(rx, ry, rz)
  if (lean) g.rotateZ(lean)
  g.translate(c.x, c.y, c.z)
  return g
}

/** A cap of hair: the top of a sphere down to `reach` (radians from the crown), tipped back so the hairline clears the brow. */
function hairCap(r: number, reach: number, at: THREE.Vector3) {
  const g = new THREE.SphereGeometry(r, seg(14), seg(7), 0, Math.PI * 2, 0, reach)
  g.rotateX(-0.38)
  g.translate(at.x, at.y, at.z)
  return g
}

/** A tapered rod from a (radius ra) to b (radius rb), with rounded ends. */
function rod(a: THREE.Vector3, b: THREE.Vector3, ra: number, rb: number, sides = 7): THREE.BufferGeometry[] {
  const d = b.clone().sub(a)
  const g = new THREE.CylinderGeometry(rb, ra, d.length(), seg(sides), 1, true)
  g.applyQuaternion(new THREE.Quaternion().setFromUnitVectors(new THREE.Vector3(0, 1, 0), d.clone().normalize()))
  const mid = a.clone().add(b).multiplyScalar(0.5)
  g.translate(mid.x, mid.y, mid.z)
  return [g, ball(a, ra, ra, ra, sides, 4), ball(b, rb, rb, rb, sides, 4)]
}

/** A body of revolution around y from (radius, height) points, flattened front to back. */
function lathe(points: [number, number][], depth = 0.8, segments = 12) {
  const g = new THREE.LatheGeometry(
    points.map(([r, y]) => new THREE.Vector2(r, y)),
    seg(segments),
  )
  g.scale(1, 1, depth)
  return g
}

function ring(y: number, r: number, tube: number, depth = 0.8, at = v3(0, 0, 0)) {
  const g = new THREE.TorusGeometry(r, tube, seg(4), seg(16))
  g.rotateX(Math.PI / 2)
  g.scale(1, 1, depth)
  g.translate(at.x, y, at.z)
  return g
}

function parts(
  list: THREE.BufferGeometry[],
  bone: number,
  region: number,
  opts: { color?: string; show?: number; feat?: number } = {},
): Part[] {
  return list.map((geo) => ({ geo, bone, region, ...opts }))
}

/**
 * Merges the parts into one geometry with per-vertex colour, `part` (bone, colour source,
 * visibility bit) and `feat` (a face feature's centre — the top, for swaying hair — and kind).
 */
function assemble(list: Part[]) {
  const box = new THREE.Box3()
  const geos = list.map(({ geo: g, bone, region, color = '#ffffff', show = 0, feat = 0 }) => {
    if (g.getAttribute('uv')) g.deleteAttribute('uv')
    const n = g.getAttribute('position').count
    const c = new THREE.Color(color)
    const col = new Float32Array(n * 3)
    for (let i = 0; i < n; i++) col.set([c.r, c.g, c.b], i * 3)
    g.setAttribute('color', new THREE.BufferAttribute(col, 3))
    const part = new Float32Array(n * 3)
    for (let i = 0; i < n; i++) part.set([bone, region, show], i * 3)
    g.setAttribute('part', new THREE.BufferAttribute(part, 3))
    const f = new Float32Array(n * 4)
    if (feat) {
      box.setFromBufferAttribute(g.getAttribute('position') as THREE.BufferAttribute)
      const at = box.getCenter(new THREE.Vector3())
      if (feat === F_HAIR || feat === F_TAIL) at.y = box.max.y
      // Skirts and their hems all drape about the waist.
      if (feat === F_SKIRT) at.set(0, 0.3, 0)
      for (let i = 0; i < n; i++) f.set([at.x, at.y, at.z, feat], i * 4)
    }
    g.setAttribute('feat', new THREE.BufferAttribute(f, 4))
    return g
  })
  return mergeGeometries(geos, false)!
}

/** The one body every islander is drawn with (outfits and props toggled per person); `lod` < 1 for a coarser one. */
export function personGeometry(lod = 1) {
  detail = lod
  try {
    return buildPerson()
  } finally {
    detail = 1
  }
}

function buildPerson() {
  const P = (b: number) => v3(...PIVOT[b])
  const list: Part[] = []
  const add = (...p: Part[][]) => p.forEach((x) => list.push(...x))

  // Head and face.
  add(
    parts([ball(v3(0, 0.492, 0.004), 0.058, 0.064, 0.054, 20, 14)], HEAD, SKIN),
    parts([ball(v3(0.059, 0.49, 0), 0.011, 0.017, 0.009), ball(v3(-0.059, 0.49, 0), 0.011, 0.017, 0.009)], HEAD, SKIN),
    parts([ball(v3(0, 0.48, 0.058), 0.01, 0.013, 0.011)], HEAD, SKIN),
    parts(
      [ball(v3(0.022, 0.498, 0.05), 0.01, 0.005, 0.005, 10, 6), ball(v3(-0.022, 0.498, 0.05), 0.01, 0.005, 0.005, 10, 6)],
      HEAD,
      FIXED,
      { color: '#e9ddc9', feat: F_EYE },
    ),
    parts(
      [ball(v3(0.022, 0.498, 0.054), 0.004, 0.0045, 0.0018, 8, 5), ball(v3(-0.022, 0.498, 0.054), 0.004, 0.0045, 0.0018, 8, 5)],
      HEAD,
      FIXED,
      { color: '#382b21', feat: F_IRIS },
    ),
    parts(
      [ball(v3(0.024, 0.516, 0.051), 0.015, 0.0035, 0.005, 8, 4), ball(v3(-0.024, 0.516, 0.051), 0.015, 0.0035, 0.005, 8, 4)],
      HEAD,
      HAIR,
      { feat: F_BROW },
    ),
    parts([ball(v3(0, 0.462, 0.052), 0.013, 0.0035, 0.005, 10, 4)], HEAD, FIXED, { color: '#7a3a2c', feat: F_MOUTH }),
    parts(rod(v3(0, 0.41, 0), v3(0, 0.45, 0.002), 0.022, 0.02), CHEST, SKIN),
  )
  if (detail >= 0.8) {
    // A bridge and nostrils, chin, eyelids and cheek planes read at portrait distance.
    add(
      parts([ball(v3(0, 0.493, 0.053), 0.006, 0.019, 0.009, 12, 8)], HEAD, SKIN),
      parts([ball(v3(0, 0.45, 0.031), 0.027, 0.015, 0.024, 14, 8)], HEAD, SKIN, { feat: F_CHIN }),
      parts(
        [ball(v3(0.025, 0.478, 0.039), 0.019, 0.012, 0.013, 12, 8), ball(v3(-0.025, 0.478, 0.039), 0.019, 0.012, 0.013, 12, 8)],
        HEAD,
        SKIN,
        { feat: F_CHEEK },
      ),
      parts(
        [ball(v3(0.022, 0.504, 0.051), 0.011, 0.003, 0.004, 10, 5), ball(v3(-0.022, 0.504, 0.051), 0.011, 0.003, 0.004, 10, 5)],
        HEAD,
        SKIN,
        { feat: F_LID },
      ),
      parts(
        [ball(v3(0.005, 0.472, 0.063), 0.0028, 0.002, 0.002, 6, 4), ball(v3(-0.005, 0.472, 0.063), 0.0028, 0.002, 0.002, 6, 4)],
        HEAD,
        FIXED,
        { color: '#5f382a' },
      ),
    )
  }

  // Arms and hands, legs and bare feet.
  for (const [arm, fore, side] of [
    [ARM_L, FORE_L, 1],
    [ARM_R, FORE_R, -1],
  ] as const) {
    const shoulder = P(arm)
    const elbow = P(fore)
    const wrist = v3(side * 0.085, 0.238, 0.004)
    add(
      parts(rod(shoulder, elbow, 0.024, 0.019), arm, SKIN),
      parts(rod(elbow, wrist, 0.018, 0.015), fore, SKIN),
      parts([ball(v3(side * 0.086, 0.218, 0.006), 0.017, 0.023, 0.013)], fore, SKIN),
      parts([ball(v3(side * 0.07, 0.225, 0.014), 0.007, 0.014, 0.007, 8, 5)], fore, SKIN),
    )
  }
  for (const [thigh, shin, side] of [
    [THIGH_L, SHIN_L, 1],
    [THIGH_R, SHIN_R, -1],
  ] as const) {
    const hip = P(thigh)
    const knee = P(shin)
    const ankle = v3(side * 0.038, 0.026, 0)
    add(
      parts(rod(hip, knee, 0.032, 0.025), thigh, SKIN),
      parts(rod(knee, ankle, 0.024, 0.017), shin, SKIN),
      parts([ball(v3(side * 0.038, 0.013, 0.016), 0.02, 0.012, 0.034)], shin, SKIN),
    )
  }

  // Man: bare chest and broad shoulders, a short kain to above the knee, an ikat kepala, short hair, a shell necklace.
  const M = { show: SHOW.male }
  add(
    parts(
      [
        lathe(
          [
            [0, 0.25],
            [0.05, 0.252],
            [0.054, 0.28],
            [0.053, 0.31],
          ],
          0.72,
        ),
      ],
      HIPS,
      SKIN,
      M,
    ),
    parts(
      [
        lathe(
          [
            [0.051, 0.295],
            [0.055, 0.33],
            [0.064, 0.37],
            [0.072, 0.395],
            [0.064, 0.414],
            [0.03, 0.428],
            [0, 0.43],
          ],
          0.68,
        ),
      ],
      CHEST,
      SKIN,
      M,
    ),
    parts([ball(v3(0.068, 0.402, 0), 0.028, 0.024, 0.026), ball(v3(-0.068, 0.402, 0), 0.028, 0.024, 0.026)], CHEST, SKIN, M),
    parts(
      [
        lathe(
          [
            [0.08, 0.188],
            [0.077, 0.21],
            [0.068, 0.245],
            [0.06, 0.28],
            [0.056, 0.298],
          ],
          0.92,
          14,
        ),
      ],
      SKIRT,
      CLOTH,
      { show: SHOW.kain, feat: F_SKIRT },
    ),
    parts(
      [
        lathe(
          [
            [0.0815, 0.186],
            [0.0805, 0.2],
          ],
          0.92,
          14,
        ),
      ],
      SKIRT,
      TRIM,
      { show: SHOW.kain, feat: F_SKIRT },
    ),
    parts([ring(0.29, 0.057, 0.0065, 0.85)], HIPS, TRIM, M),
    parts([ring(0.414, 0.034, 0.005, 0.8)], CHEST, FIXED, { color: '#efe2c4', ...M }),
    parts([ring(0.514, 0.0632, 0.0055, 0.95, v3(0, 0, 0.001))], HEAD, TRIM, M),
    parts(
      [ball(v3(0.012, 0.49, -0.068), 0.006, 0.028, 0.004, 8, 6, 0.2), ball(v3(-0.012, 0.487, -0.068), 0.006, 0.026, 0.004, 8, 6, -0.25)],
      HEAD,
      TRIM,
      { ...M, feat: F_TAIL },
    ),
    parts([hairCap(0.0655, 1.4, v3(0, 0.494, -0.002))], HEAD, HAIR, M),
  )

  // Woman: a sarong to below the knee, a kemben, long hair down the back, sometimes a frangipani.
  const F = { show: SHOW.female }
  add(
    parts(
      [
        lathe(
          [
            [0.048, 0.26],
            [0.05, 0.3],
            [0.056, 0.335],
            [0.062, 0.37],
            [0.064, 0.4],
            [0.056, 0.415],
            [0.028, 0.428],
            [0, 0.43],
          ],
          0.7,
        ),
      ],
      CHEST,
      SKIN,
      F,
    ),
    parts(
      [
        lathe(
          [
            [0.056, 0.334],
            [0.062, 0.352],
            [0.066, 0.37],
            [0.063, 0.388],
            [0.06, 0.395],
          ],
          0.86,
        ),
      ],
      CHEST,
      TRIM,
      F,
    ),
    parts(
      [
        lathe(
          [
            [0.088, 0.072],
            [0.083, 0.12],
            [0.072, 0.2],
            [0.062, 0.26],
            [0.054, 0.305],
            [0.051, 0.322],
          ],
          0.92,
          14,
        ),
      ],
      SKIRT,
      CLOTH,
      { show: SHOW.sarong, feat: F_SKIRT },
    ),
    parts(
      [
        lathe(
          [
            [0.09, 0.07],
            [0.0895, 0.092],
          ],
          0.92,
          14,
        ),
      ],
      SKIRT,
      TRIM,
      { show: SHOW.sarong, feat: F_SKIRT },
    ),
    parts([ring(0.312, 0.053, 0.007, 0.85)], HIPS, TRIM, F),
    parts([hairCap(0.0665, 1.62, v3(0, 0.494, -0.004))], HEAD, HAIR, F),
    parts([ball(v3(0, 0.425, -0.042), 0.05, 0.1, 0.026)], HEAD, HAIR, { ...F, feat: F_HAIR }),
    parts([ball(v3(0, 0.335, -0.04), 0.04, 0.03, 0.02)], HEAD, HAIR, { ...F, feat: F_HAIR }),
    parts([ball(v3(0, 0.285, 0.022), 0.05, 0.048, 0.042)], BELLY, CLOTH, F),
  )
  const flower: THREE.BufferGeometry[] = []
  for (let k = 0; k < 5; k++) {
    const a = (k / 5) * Math.PI * 2
    flower.push(ball(v3(0.056 + Math.cos(a) * 0.009, 0.525 + Math.sin(a) * 0.009, 0.018), 0.008, 0.008, 0.004, 5, 3))
  }
  add(
    parts(flower, HEAD, FIXED, { color: '#fff7ec', show: SHOW.flower }),
    parts([ball(v3(0.056, 0.525, 0.021), 0.005, 0.005, 0.003, 6, 4)], HEAD, FIXED, { color: '#f4c03f', show: SHOW.flower }),
  )

  // A household accent is woven cloth, not a crown implying a political rank.
  add(parts([ring(0.525, 0.057, 0.005, 0.95)], HEAD, TRIM, { show: SHOW.chief }))

  // Props.
  const hand = v3(-0.086, 0.218, 0.006)
  const handL = v3(0.086, 0.214, 0.02)
  add(
    parts([new THREE.CylinderGeometry(0.075, 0.056, 0.042, seg(14)).translate(0, 0.574, 0)], HEAD, FIXED, {
      color: '#b58a4e',
      show: SHOW.basket,
    }),
    parts([ring(0.595, 0.075, 0.006, 1)], HEAD, FIXED, { color: '#8a6533', show: SHOW.basket }),
    parts([ball(v3(0.024, 0.6, 0.012), 0.022, 0.02, 0.022)], HEAD, FIXED, { color: '#e9a23b', show: SHOW.basket }),
    parts([ball(v3(-0.02, 0.602, -0.015), 0.02, 0.02, 0.02)], HEAD, FIXED, { color: '#7fbf3f', show: SHOW.basket }),
    parts([ball(v3(-0.012, 0.598, 0.026), 0.017, 0.017, 0.017)], HEAD, FIXED, { color: '#c0392b', show: SHOW.basket }),
    parts([ball(v3(0, 0.355, -0.072), 0.056, 0.07, 0.04)], CHEST, FIXED, { color: '#9b7a4c', show: SHOW.sack }),
    // The spear runs along the forearm, its stone point beyond the hand.
    parts(rod(v3(hand.x, -0.12, 0.006), v3(hand.x, 0.36, 0.006), 0.006, 0.006, 5), FORE_R, FIXED, { color: '#7a5534', show: SHOW.spear }),
    parts([new THREE.ConeGeometry(0.014, 0.06, seg(6)).rotateX(Math.PI).translate(hand.x, -0.15, 0.006)], FORE_R, FIXED, {
      color: '#5d6066',
      show: SHOW.spear,
    }),
    parts(rod(v3(hand.x, 0.214, -0.01), v3(hand.x, 0.214, 0.11), 0.006, 0.006, 5), FORE_R, FIXED, { color: '#7a5534', show: SHOW.hammer }),
    parts([ball(v3(hand.x, 0.214, 0.12), 0.02, 0.026, 0.022, 7, 5)], FORE_R, FIXED, { color: '#80838a', show: SHOW.hammer }),
    parts(rod(v3(hand.x, 0.0, 0.006), v3(hand.x, 0.33, 0.006), 0.004, 0.007, 5), FORE_R, FIXED, { color: '#8a6238', show: SHOW.stick }),
    // A dayung: a shaft through the hand, the blade below it in the water.
    parts(rod(v3(hand.x, -0.13, 0.006), v3(hand.x, 0.32, 0.006), 0.005, 0.005, 5), FORE_R, FIXED, { color: '#8a6238', show: SHOW.paddle }),
    parts([ball(v3(hand.x, -0.15, 0.006), 0.006, 0.05, 0.022, 6, 5)], FORE_R, FIXED, { color: '#9c7446', show: SHOW.paddle }),
    // A bundle wrapped in leaf, tied with fibre.
    parts([ball(handL, 0.024, 0.018, 0.022, 8, 6)], FORE_L, FIXED, { color: '#b9a35e', show: SHOW.bundle }),
    parts([ring(handL.y, 0.022, 0.003, 1, v3(handL.x, 0, handL.z))], FORE_L, FIXED, { color: '#6d5530', show: SHOW.bundle }),
  )
  const geometry = assemble(list)
  const positions = geometry.getAttribute('position')
  const normals = geometry.getAttribute('normal')
  const part = geometry.getAttribute('part')
  const feat = geometry.getAttribute('feat')
  // Adult proportions: reduce the former diorama head around the neck pivot.
  const shrink = (x: number, y: number, z: number) => [x * 0.74, 0.43 + (y - 0.43) * 0.64, z * 0.76] as const
  for (let i = 0; i < positions.count; i++)
    if (part.getX(i) === HEAD) {
      positions.setXYZ(i, ...shrink(positions.getX(i), positions.getY(i), positions.getZ(i)))
      if (feat.getW(i)) feat.setXYZ(i, ...shrink(feat.getX(i), feat.getY(i), feat.getZ(i)))
      const n = v3(normals.getX(i) / 0.74, normals.getY(i) / 0.64, normals.getZ(i) / 0.76).normalize()
      normals.setXYZ(i, n.x, n.y, n.z)
    }
  return geometry
}

// --- Poses ----------------------------------------------------------------------------

type Action = 'idle' | 'gather' | 'plant' | 'hammer' | 'craft' | 'sit' | 'eat' | 'drink' | 'teach' | 'give' | 'trade' | 'strike' | 'aim'

function actionOf(flags: number): Action {
  if (flags & FLAG.resting) return 'sit'
  if (flags & FLAG.attacking) return 'strike'
  if (flags & FLAG.hunting) return 'aim'
  if (flags & FLAG.building) return 'hammer'
  if (flags & FLAG.crafting) return 'craft'
  if (flags & FLAG.planting) return 'plant'
  if (flags & FLAG.gathering) return 'gather'
  if (flags & FLAG.drinking) return 'drink'
  if (flags & FLAG.eating) return 'eat'
  if (flags & FLAG.teaching) return 'teach'
  if (flags & FLAG.trading) return 'trade'
  if (flags & FLAG.giving) return 'give'
  return 'idle'
}

/** Ways of moving that replace walking (FLAG.swimming, rafting, climbing). */
const SWIM = 1
const RAFT = 2
const CLIMB = 3
const locomotionOf = (flags: number) => (flags & FLAG.swimming ? SWIM : flags & FLAG.rafting ? RAFT : flags & FLAG.climbing ? CLIMB : 0)

const set = (p: Float32Array, b: number, x: number, y = 0, z = 0) => {
  p[rot(b)] = x
  p[rot(b) + 1] = y
  p[rot(b) + 2] = z
}

/** 0→1→0 once per period, eased: for repeated motions like hammering or picking. */
const cycle = (t: number, period: number, seed: number) => 0.5 - 0.5 * Math.cos(((t / period + seed) % 1) * Math.PI * 2)

/** A sharp work stroke: slow up for most of the period, then quickly down. */
const stroke = (t: number, period: number, seed: number) => {
  const u = (t / period + seed) % 1
  return u < 0.7 ? Math.sin((u / 0.7) * Math.PI * 0.5) : Math.cos(((u - 0.7) / 0.3) * Math.PI * 0.5)
}

const ease = (x: number) => (x <= 0 ? 0 : x >= 1 ? 1 : x * x * (3 - 2 * x))

function standing(p: Float32Array, t: number, seed: number) {
  p.fill(0)
  p[P_HEAD_SCALE] = 1
  // Relaxed arms, a breath, a glance around, weight shifting from foot to foot.
  set(p, ARM_L, 0.04, 0, 0.13)
  set(p, ARM_R, 0.04, 0, -0.13)
  set(p, FORE_L, -0.18)
  set(p, FORE_R, -0.18)
  set(p, CHEST, 0.02 * Math.sin(t * 2.1 + seed * 9), 0, 0)
  set(p, HEAD, 0.06 * Math.sin(t * 0.23 + seed * 5), 0.45 * Math.sin(t * 0.31 + seed * 7) * Math.sin(t * 0.11 + seed * 3), 0)
  set(p, HIPS, 0, 0, 0.035 * Math.sin(t * 0.45 + seed * 4))
}

/** Hips low, knees up, arms round them. */
function sitting(p: Float32Array) {
  p[P_ROOT_Y] = -0.175
  set(p, THIGH_L, -1.95, 0, 0.18)
  set(p, THIGH_R, -1.95, 0, -0.18)
  set(p, SHIN_L, 2.25)
  set(p, SHIN_R, 2.25)
  set(p, CHEST, 0.32)
}

function crouching(p: Float32Array) {
  p[P_ROOT_Y] = -0.1
  set(p, THIGH_L, -1.3, 0, 0.12)
  set(p, THIGH_R, -1.3, 0, -0.12)
  set(p, SHIN_L, 1.9)
  set(p, SHIN_R, 1.9)
}

function actionPose(a: Action, p: Float32Array, t: number, seed: number) {
  standing(p, t, seed)
  switch (a) {
    case 'gather': {
      // Bent at the waist, picking with one hand and then the other.
      const k = cycle(t, 1.3, seed)
      set(p, THIGH_L, -0.25)
      set(p, THIGH_R, -0.25)
      set(p, SHIN_L, 0.35)
      set(p, SHIN_R, 0.35)
      p[P_ROOT_Y] = -0.012
      set(p, CHEST, 0.95)
      set(p, HEAD, -0.25)
      set(p, ARM_R, -0.75 - 0.45 * k, 0, -0.08)
      set(p, FORE_R, -0.25)
      set(p, ARM_L, -0.6 - 0.4 * (1 - k), 0, 0.08)
      set(p, FORE_L, -0.35)
      break
    }
    case 'plant': {
      // Crouched, the tugal stabbing the soil, the other hand dropping seed.
      crouching(p)
      const k = stroke(t, 0.9, seed)
      set(p, CHEST, 0.45)
      set(p, HEAD, 0.15)
      set(p, ARM_R, -0.9 - 0.5 * k, 0, -0.1)
      set(p, FORE_R, -0.6 + 0.3 * k)
      set(p, ARM_L, -0.8, 0, 0.15)
      set(p, FORE_L, -0.5 - 0.3 * cycle(t, 0.9, seed + 0.3))
      break
    }
    case 'hammer': {
      const k = stroke(t, 0.75, seed)
      set(p, THIGH_L, -0.1, 0, 0.12)
      set(p, THIGH_R, 0.1, 0, -0.12)
      set(p, CHEST, 0.25 + 0.15 * (1 - k), -0.15, 0)
      set(p, HEAD, 0.25)
      set(p, ARM_R, -0.5 - 1.9 * k, 0, -0.15)
      set(p, FORE_R, -0.4 - 0.5 * k)
      set(p, ARM_L, -1.0, 0, 0.2)
      set(p, FORE_L, -0.7)
      break
    }
    case 'craft': {
      // Sitting with the work in the lap, hands busy.
      sitting(p)
      set(p, THIGH_L, -1.55, 0.3, 0.45)
      set(p, THIGH_R, -1.55, -0.3, -0.45)
      set(p, SHIN_L, 2.35)
      set(p, SHIN_R, 2.35)
      p[P_ROOT_Y] = -0.2
      set(p, CHEST, 0.42)
      set(p, HEAD, 0.35)
      const k = cycle(t, 0.45, seed)
      set(p, ARM_R, -0.75, 0, -0.15)
      set(p, FORE_R, -1.0 - 0.35 * k)
      set(p, ARM_L, -0.7, 0, 0.15)
      set(p, FORE_L, -1.05 - 0.2 * (1 - k))
      break
    }
    case 'sit': {
      sitting(p)
      const breathe = 0.03 * Math.sin(t * 1.4 + seed * 6)
      set(p, CHEST, 0.42 + breathe)
      set(p, HEAD, 0.25 + 0.1 * Math.sin(t * 0.2 + seed), 0.3 * Math.sin(t * 0.17 + seed * 3))
      set(p, ARM_L, -1.0, 0, 0.2)
      set(p, ARM_R, -1.0, 0, -0.2)
      set(p, FORE_L, -1.15, 0.4, 0)
      set(p, FORE_R, -1.15, -0.4, 0)
      break
    }
    case 'eat': {
      const k = cycle(t, 1.6, seed)
      set(p, ARM_R, -0.55 - 0.15 * k, 0, -0.25)
      set(p, FORE_R, -1.3 - 0.85 * k)
      set(p, ARM_L, -0.45, 0, 0.12)
      set(p, FORE_L, -1.0)
      set(p, HEAD, 0.1 - 0.12 * k)
      break
    }
    case 'drink': {
      // Kneeling at the water, scooping it up in both hands.
      crouching(p)
      const k = cycle(t, 1.8, seed)
      set(p, CHEST, 0.85 - 0.6 * k)
      set(p, HEAD, 0.1 - 0.35 * k)
      set(p, ARM_L, -0.9 - 0.3 * k, 0, 0.1)
      set(p, ARM_R, -0.9 - 0.3 * k, 0, -0.1)
      set(p, FORE_L, -0.3 - 1.4 * k)
      set(p, FORE_R, -0.3 - 1.4 * k)
      break
    }
    case 'teach': {
      const k = cycle(t, 2.2, seed)
      set(p, ARM_R, -1.35 - 0.25 * k, 0, -0.2)
      set(p, FORE_R, -0.2)
      set(p, ARM_L, -0.3, 0, 0.25)
      set(p, FORE_L, -1.2)
      set(p, HEAD, 0.05 * Math.sin(t * 3 + seed), 0.25 * (k - 0.5))
      break
    }
    case 'give': {
      set(p, ARM_L, -1.15, 0, 0.05)
      set(p, ARM_R, -1.15, 0, -0.05)
      set(p, FORE_L, -0.35)
      set(p, FORE_R, -0.35)
      set(p, CHEST, 0.15)
      set(p, HEAD, 0.2)
      break
    }
    case 'trade': {
      // Holding a bundle out to the other, then drawing it back as theirs comes across.
      const k = cycle(t, 2.6, seed)
      set(p, ARM_L, -0.45 - 0.7 * k, 0, 0.06)
      set(p, FORE_L, -1.15 + 0.8 * k, 0.5, 0)
      set(p, ARM_R, -0.35 - 0.55 * (1 - k), 0, -0.1)
      set(p, FORE_R, -0.9 + 0.3 * (1 - k), -0.5, 0)
      set(p, CHEST, 0.1 + 0.1 * k)
      set(p, HEAD, 0.18, 0.1 * Math.sin(t * 0.8 + seed * 4))
      break
    }
    case 'strike': {
      const k = stroke(t, 0.5, seed)
      set(p, THIGH_L, -0.35, 0, 0.1)
      set(p, THIGH_R, 0.3, 0, -0.1)
      set(p, SHIN_R, 0.3)
      set(p, CHEST, 0.3 * (1 - k), -0.5 + 0.9 * (1 - k), 0)
      set(p, ARM_R, -0.6 - 2.0 * k, 0, -0.3)
      set(p, FORE_R, -0.3 - 0.6 * k)
      set(p, ARM_L, -0.9, 0, 0.3)
      set(p, FORE_L, -1.4)
      break
    }
    case 'aim': {
      // Lunging, the spear levelled at the prey and thrust.
      const k = stroke(t, 1.1, seed)
      set(p, THIGH_L, -0.45, 0, 0.12)
      set(p, THIGH_R, 0.3, 0, -0.12)
      set(p, SHIN_L, 0.35)
      set(p, SHIN_R, 0.3)
      p[P_ROOT_Y] = -0.02
      set(p, CHEST, 0.22, -0.25 + 0.2 * k, 0)
      set(p, ARM_R, -0.55 - 0.45 * k, 0, -0.25)
      set(p, FORE_R, -1.25 + 0.55 * k)
      set(p, ARM_L, -0.95 - 0.2 * k, 0, 0.3)
      set(p, FORE_L, -0.9)
      set(p, HEAD, 0.1, 0.2)
      break
    }
    default:
  }
}

/**
 * One stride: legs and arms swing in opposition, knees fold through the swing, the body bobs and twists.
 * Wading (`wade` 1) lifts the knees high out of the water, holds the arms up and out, and leans in.
 */
function walkPose(p: Float32Array, phi: number, run: number, female: boolean, wade: number) {
  const s = Math.sin(phi)
  const c = Math.cos(phi)
  // A sarong keeps a woman's steps short.
  const stride = (female ? 0.4 + 0.2 * run : 0.52 + 0.45 * run) * (1 - 0.25 * wade)
  const fold = 0.85 + 0.75 * run + 0.55 * wade
  const swing = (0.42 + 0.45 * run) * (1 - 0.6 * wade)
  const lift = 0.45 * wade
  set(p, THIGH_L, -stride * s - lift * Math.max(0, c), 0, 0.02 + 0.05 * wade)
  set(p, THIGH_R, stride * s - lift * Math.max(0, -c), 0, -0.02 - 0.05 * wade)
  set(p, SHIN_L, 0.08 + fold * Math.max(0, c))
  set(p, SHIN_R, 0.08 + fold * Math.max(0, -c))
  set(p, ARM_L, swing * s - 0.25 * wade, 0, 0.1 + 0.35 * wade)
  set(p, ARM_R, -swing * s - 0.25 * wade, 0, -0.1 - 0.35 * wade)
  set(p, FORE_L, -(0.2 + 0.4 * run) - (0.45 + 0.6 * run) * Math.max(0, -s) - 0.55 * wade)
  set(p, FORE_R, -(0.2 + 0.4 * run) - (0.45 + 0.6 * run) * Math.max(0, s) - 0.55 * wade)
  set(p, HIPS, 0, 0.12 * s, (female ? 0.07 * s : 0.02 * s) * (1 + wade))
  set(p, CHEST, -0.04 + 0.22 * run + 0.14 * wade, -0.17 * s, 0)
  set(p, HEAD, 0.03 - 0.12 * run + 0.06 * wade, 0.06 * s, 0)
  p[P_ROOT_Y] = (0.012 + 0.025 * run) * (Math.abs(c) - 0.6) - 0.012 * wade
  p[P_ROOT_Z] = 0
}

/** Standing in the shallows: arms held up out of the water, a slow sway against the current. */
function wadeStill(p: Float32Array, t: number, seed: number) {
  set(p, ARM_L, -0.25, 0, 0.42)
  set(p, ARM_R, -0.25, 0, -0.42)
  set(p, FORE_L, -0.75)
  set(p, FORE_R, -0.75)
  set(p, CHEST, 0.06 + 0.03 * Math.sin(t * 0.9 + seed * 5))
  set(p, HIPS, 0, 0, 0.05 * Math.sin(t * 0.7 + seed * 3))
}

/**
 * Ways of moving that replace walking. `phase` advances with the distance travelled, `go` is how
 * much they are moving (0 still … 1 under way). Swimmers are drawn from the water's surface (y = 0):
 * only head and shoulders stay above it; paddlers kneel on the raft's deck (y = 0).
 */
function locomotionPose(mode: number, p: Float32Array, t: number, phase: number, go: number, seed: number) {
  p.fill(0)
  p[P_HEAD_SCALE] = 1
  if (mode === SWIM) {
    // Breaststroke under way, treading water at rest: the body leans into the water, the head kept level.
    const lean = 0.3 + 0.85 * go
    const s = Math.sin(phase),
      c = Math.cos(phase)
    const tread = Math.sin(t * 3 + seed * 6)
    set(p, HIPS, lean)
    set(p, HEAD, -lean * 0.85, 0.15 * Math.sin(t * 0.4 + seed * 3))
    const reach = go * (-1.9 - 0.7 * s) + (1 - go) * -0.55
    const spread = go * (0.35 + 0.6 * Math.max(0, c)) + (1 - go) * (0.95 + 0.25 * tread)
    const bend = go * (-0.35 - 0.95 * Math.max(0, -s)) + (1 - go) * (-0.55 - 0.2 * tread)
    set(p, ARM_L, reach, 0, spread)
    set(p, ARM_R, reach, 0, -spread)
    set(p, FORE_L, bend)
    set(p, FORE_R, bend)
    // Frog kick under way, slow pedalling at rest.
    const kick = go * Math.max(0, s)
    const pedalL = (1 - go) * Math.sin(t * 2.4 + seed * 4),
      pedalR = (1 - go) * Math.sin(t * 2.4 + seed * 4 + Math.PI)
    set(p, THIGH_L, -0.25 - 0.55 * kick - 0.35 * (1 - go) + 0.25 * pedalL, 0, 0.12 + 0.35 * kick)
    set(p, THIGH_R, -0.25 - 0.55 * kick - 0.35 * (1 - go) + 0.25 * pedalR, 0, -0.12 - 0.35 * kick)
    set(p, SHIN_L, 0.25 + 1.35 * kick + (1 - go) * (0.7 + 0.3 * pedalL))
    set(p, SHIN_R, 0.25 + 1.35 * kick + (1 - go) * (0.7 + 0.3 * pedalR))
    // Shoulders at the surface, the body below it; centred over where they are.
    p[P_ROOT_Y] = -0.27 - 0.13 * Math.cos(lean) + 0.012 + 0.006 * Math.sin(t * 2.2 + seed * 3)
    p[P_ROOT_Z] = -0.08 * go
  } else if (mode === RAFT) {
    // Kneeling on the raft, paddling on the right: reach forward, dig in, pull back past the hip.
    const k = stroke(t, 1.5, seed)
    set(p, THIGH_L, -0.3, 0, 0.14)
    set(p, THIGH_R, -0.3, 0, -0.14)
    set(p, SHIN_L, 1.95)
    set(p, SHIN_R, 1.95)
    // Knees and shins resting on the deck (y = 0).
    p[P_ROOT_Y] = -0.115
    const pull = go > 0.05 ? 1 - k : 0.5 + 0.1 * Math.sin(t * 0.8 + seed)
    set(p, CHEST, 0.3 + 0.18 * (1 - pull), -0.2 + 0.4 * pull, 0)
    set(p, HEAD, -0.1, 0.25 - 0.3 * pull)
    set(p, ARM_R, -1.25 + 1.3 * pull, 0, -0.32)
    set(p, FORE_R, -0.25 - 0.2 * pull)
    set(p, ARM_L, -1.75 + 0.7 * pull, 0, -0.35)
    set(p, FORE_L, -1.3 + 0.3 * pull, -0.4, 0)
  } else if (mode === CLIMB) {
    // Leaning into the slope, hands reaching up it in turn, feet stepping high.
    const s = Math.sin(phase)
    const upL = 0.5 + 0.5 * s,
      upR = 0.5 - 0.5 * s
    set(p, HIPS, 0.22)
    set(p, CHEST, 0.45)
    set(p, HEAD, -0.55, 0.1 * Math.sin(t * 0.5 + seed))
    set(p, ARM_L, -2.15 - 0.5 * upL, 0, 0.25)
    set(p, ARM_R, -2.15 - 0.5 * upR, 0, -0.25)
    set(p, FORE_L, -0.3 - 0.45 * upR)
    set(p, FORE_R, -0.3 - 0.45 * upL)
    set(p, THIGH_L, -0.55 - 0.65 * upL, 0, 0.08)
    set(p, THIGH_R, -0.55 - 0.65 * upR, 0, -0.08)
    set(p, SHIN_L, 0.55 + 0.8 * upL)
    set(p, SHIN_R, 0.55 + 0.8 * upR)
    p[P_ROOT_Y] = -0.04
    p[P_ROOT_Z] = -0.02
  }
}

/** Arms that stay busy while walking: holding the bakul on the head, the karung's strap, the spear. */
function carryArms(p: Float32Array, flags: number, female: boolean): number {
  let mask = 0
  if (flags & FLAG.carrying) {
    if (female) {
      set(p, ARM_L, -0.2, 0, 2.65)
      set(p, FORE_L, -0.2, 0, 1.15)
      mask |= 1
    } else {
      set(p, ARM_L, -0.55, 0, 0.12)
      set(p, FORE_L, -2.0)
      set(p, ARM_R, -0.55, 0, -0.12)
      set(p, FORE_R, -2.0)
      mask |= 3
    }
  }
  if (flags & FLAG.hunting) {
    set(p, ARM_R, -0.35, 0, -0.15)
    set(p, FORE_R, -1.0)
    mask |= 2
  }
  return mask
}

/** Talking (after GestureManager): hand beats, a hand to the chest or pointing; listeners nod. Returns how much they speak (0–1). */
function talkOverlay(p: Float32Array, t: number, seed: number, g: Gesture, armsFree: number) {
  gestureAt(t, seed, g)
  const e = g.env
  const left = g.side > 0
  const arm = left ? ARM_L : ARM_R,
    fore = left ? FORE_L : FORE_R,
    out = left ? 1 : -1
  const useArm = (b: number) => (b === ARM_L || b === FORE_L ? armsFree & 1 : armsFree & 2)
  const add = (b: number, x: number, y = 0, z = 0) => {
    if (!useArm(b)) return
    p[rot(b)] += x
    p[rot(b) + 1] += y
    p[rot(b) + 2] += z
  }
  switch (g.kind) {
    case GESTURE.open:
      add(arm, -0.55 * e, 0, out * 0.3 * e)
      add(fore, -0.95 * e, out * 0.7 * e)
      break
    case GESTURE.both:
      add(ARM_L, -0.5 * e, 0, 0.28 * e)
      add(ARM_R, -0.5 * e, 0, -0.28 * e)
      add(FORE_L, -1.0 * e, 0.6 * e)
      add(FORE_R, -1.0 * e, -0.6 * e)
      break
    case GESTURE.point:
      add(arm, -1.15 * e, 0, -out * 0.05 * e)
      add(fore, -0.1 * e)
      break
    case GESTURE.chest:
      add(arm, -0.3 * e, 0, -out * 0.12 * e)
      add(fore, -1.75 * e, -out * 0.5 * e)
      break
    default:
  }
  p[rot(HEAD)] += 0.13 * g.nod
  p[rot(HEAD) + 2] += 0.05 * out * e
  return g.speak
}

// --- The crowd --------------------------------------------------------------------

const SKINS = ['#8d5524', '#a0662f', '#c68642', '#d9a066', '#e0ac69', '#b0703a'].map((c) => new THREE.Color(c))
const HAIRS = ['#1d1612', '#2e2018', '#3b2a1e', '#4a3020', '#141414'].map((c) => new THREE.Color(c))
const HURT = new THREE.Color('#d98a7a')
const SOOT = new THREE.Color('#3a302a')
const EARTH = new THREE.Color('#5e5143')

const hash = (n: number, s: number) => {
  let h = (n * 374761393 + s * 668265263) | 0
  h = Math.imul(h ^ (h >>> 13), 1274126177)
  return ((h ^ (h >>> 16)) >>> 0) / 4294967296
}

/**
 * Per person, one texture row: the bone matrices (4 texels each), then EXTRA texels — skin
 * (+ visible parts), hair, clothes (+ hair sway x), trim (+ hair sway z), and two of face.
 */
const EXTRA = 6
const ROW = NB * 4 + EXTRA
const ROW_FLOATS = ROW * 4
const X_SKIN = NB * 16
const X_HAIR = X_SKIN + 4
const X_CLOTH = X_SKIN + 8
const X_TRIM = X_SKIN + 12
const X_FACE = X_SKIN + 16

/** Room kept for bodies after the living, so drawing them never has to grow the mesh. */
const BODY_ROOM = 48
/** At most this many ragdolls are simulated at once (bodies and knockdowns); more settle at once. */
const MAX_RAGDOLLS = 24
/** A body first seen later than this after death (simulated seconds) is placed already lying. */
const FRESH_BODY = 1
/** How long the server shows a body (simulated seconds), and when it starts to sink and fade. */
const BODY_LIFE = 4
const BODY_FADE = 3
/** Faces are worked out within this many body heights of the camera; cloth within CLOTH_REACH. */
const FACE_REACH = 9
const CLOTH_REACH = 10
/** How long getting back up takes (seconds). */
const RISE_TIME = 1.4

type Anim = {
  pose: Float32Array
  walk: number
  run: number
  walked: number
  speed: number
  seed: number
  look: number
  /** Crowd clock when last streamed (a body starts from the pose its person was last seen in). */
  seen: number
  flags: number
  /** Mood weights (MOOD order), eased. */
  mood: Float32Array
  /** Crowd clock of the last blow, and which way it rocked them. */
  hitAt: number
  hitSide: number
  /** Swimming, rafting, climbing or on foot (0) last frame. */
  mode: number
  /** Knocked down: the pose while down (from the ragdoll), its ragdoll while it falls, when the fall ended. */
  down: Float32Array | null
  rag: Ragdoll | null
  upAt: number
}

type Body = { rag: Ragdoll | null; pose: Float32Array; seed: number }

export type Placement = {
  /** Ground height under a point. */
  ground: (x: number, z: number) => number
  /** Size of an adult of size 1 (the world's person scale, grown a little when seen from afar). */
  scale: number
  time: number
  dt: number
  /** Reduced motion: hold still poses, no falls (bodies lie settled), no cloth or hair in the wind. */
  still: boolean
  /** Whether a body at this spot (feet) and size could be on screen; others are skipped. */
  inView?: (x: number, y: number, z: number, size: number) => boolean
  /** Nearby visible people get contact IK, faces, body language and cloth; distant crowds use their simpler gait. */
  contact?: boolean
  /** The wind from the simulation (direction it blows towards in world x/z, strength 0–1), for cloth and hair. */
  wind?: { dirX: number; dirZ: number; strength: number }
  /**
   * The drawn water surface under a point (world y, with its swell), null (or NaN) on dry land:
   * swimmers float with their shoulders at it, rafters kneel at it + RAFT_DECK × scale, waders
   * keep their feet on the bed (the ground) with the surface around their legs.
   */
  water?: (x: number, z: number) => number | null
}

// Scratch, reused every frame.
const M = new THREE.Matrix4()
const Q = new THREE.Quaternion()
const S = new THREE.Vector3()
const POS = new THREE.Vector3()
const COL = new THREE.Color()
const TARGET = new Float32Array(POSE_SIZE)
const WALK = new Float32Array(POSE_SIZE)
const FINAL = new Float32Array(POSE_SIZE)
const FACE = new Float32Array(FACE_CHANNELS)
const POSTURES = new Float32Array(POSTURE_CHANNELS)
const GEST: Gesture = { kind: 0, side: 1, env: 0, nod: 0, speak: 0 }
const POINTS = new Float64Array(RAG_N * 3)
const RAG_OUT = new Float64Array(RAG_POSE_SIZE)
const ZERO_FACE = new Float32Array(8)

/** The ground under a body, in its own body units (for ragdolls): set before stepping each one. */
const groundMap = { ox: 0, oy: 0, oz: 0, cs: 1, sn: 0, sx: 1, sy: 1, sz: 1, ground: (_x: number, _z: number) => 0 }
const bodyGround: BodyGround = (bx, bz) => {
  const g = groundMap
  const x = bx * g.sx,
    z = bz * g.sz
  return (g.ground(g.ox + x * g.cs + z * g.sn, g.oz - x * g.sn + z * g.cs) - g.oy) / g.sy
}
function mapGround(ground: Placement['ground'], ox: number, oy: number, oz: number, heading: number, sx: number, sy: number, sz: number) {
  const g = groundMap
  const angle = Math.PI / 2 - heading
  g.ground = ground
  g.ox = ox
  g.oy = oy
  g.oz = oz
  g.cs = Math.cos(angle)
  g.sn = Math.sin(angle)
  g.sx = sx
  g.sy = sy
  g.sz = sz
}

/** The ragdoll's particles for a pose (body space). */
function posePoints(p: Float32Array, out: Float64Array) {
  writeBones(p, SCRATCH_BONES, 0)
  for (let k = 0; k < RAG_N; k++) {
    POS.set(RAG_REST[k * 3], RAG_REST[k * 3 + 1], RAG_REST[k * 3 + 2]).applyMatrix4(world[RAG_BONE[k]])
    out[k * 3] = POS.x
    out[k * 3 + 1] = POS.y
    out[k * 3 + 2] = POS.z
  }
  return out
}

/** The skeleton's pose for a ragdoll's particles. */
function ragdollToPose(points: ArrayLike<number>, out: Float32Array, head: number, belly: number) {
  ragdollPose(points, RAG_OUT)
  out.fill(0)
  out[P_ROOT_X] = RAG_OUT[0]
  out[P_ROOT_Y] = RAG_OUT[1]
  out[P_ROOT_Z] = RAG_OUT[2]
  const copy = (b: number, o: number) => set(out, b, RAG_OUT[o], RAG_OUT[o + 1], RAG_OUT[o + 2])
  copy(HIPS, RAG_POSE.hips)
  copy(CHEST, RAG_POSE.chest)
  copy(HEAD, RAG_POSE.head)
  copy(ARM_L, RAG_POSE.armL)
  copy(FORE_L, RAG_POSE.foreL)
  copy(ARM_R, RAG_POSE.armR)
  copy(FORE_R, RAG_POSE.foreR)
  copy(THIGH_L, RAG_POSE.thighL)
  copy(SHIN_L, RAG_POSE.shinL)
  copy(THIGH_R, RAG_POSE.thighR)
  copy(SHIN_R, RAG_POSE.shinR)
  copy(SKIRT, RAG_POSE.skirt)
  out[P_DRAPE] = 1
  out[P_HEAD_SCALE] = head
  out[P_BELLY] = belly
  return out
}

/** A body lying on its back, settled on flat ground: for knockdowns drawn without physics. */
let lying: Float32Array | null = null
function lyingPose() {
  if (!lying) {
    const p = new Float32Array(POSE_SIZE)
    standing(p, 0, 0)
    const r = createRagdoll(posePoints(p, POINTS), 0.5, { x: 0.3, z: -1, strength: 0.6 })
    settleRagdoll(r, () => 0)
    lying = ragdollToPose(r.pos, new Float32Array(POSE_SIZE), 1, 0.01)
  }
  return lying
}

/** Halfway up from the ground: kneeling on one knee, hands on the thigh. */
const RISING = (() => {
  const p = new Float32Array(POSE_SIZE)
  standing(p, 0, 0)
  crouching(p)
  set(p, THIGH_R, -0.3, 0, -0.1)
  set(p, SHIN_R, 1.9)
  set(p, CHEST, 0.6)
  set(p, ARM_L, -0.9, 0, 0.1)
  set(p, FORE_L, -0.6)
  set(p, ARM_R, -0.5, 0, -0.15)
  set(p, FORE_R, -0.4)
  p[rot(SKIRT)] = 0.45 * (p[rot(THIGH_L)] + p[rot(THIGH_R)])
  return p
})()

export class Crowd {
  mesh: THREE.InstancedMesh
  /** Person id per instance, in drawing order (for picking). Bodies of the dead come after and have none. */
  readonly ids: number[] = []
  /** Close-up and far-away bodies; they share the per-person texture. */
  private readonly near = personGeometry()
  private readonly far = personGeometry(0.55)
  private geometry = this.near
  private readonly material: THREE.MeshStandardMaterial
  private readonly depthMaterial: THREE.MeshDepthMaterial
  private readonly bones = { value: null as THREE.DataTexture | null }
  private boneData = new Float32Array(0)
  private anims = new Map<number, Anim>()
  private bodies = new Map<number, Body>()
  /** Ragdolls still moving (bodies and knockdowns), within MAX_RAGDOLLS. */
  private readonly moving = new Set<Ragdoll>()
  private readonly bodiesSeen = new Set<number>()
  private readonly skirts = new Skirts()
  private cap = 0
  /** Seconds of drawing so far (keeps going when reduced motion freezes `time`). */
  private clock = 0
  private live = 0
  /** Where the camera was when the crowd was last drawn (for faces and cloth up close). */
  private readonly eye = new THREE.Vector3()
  private hasEye = false

  constructor() {
    this.material = new THREE.MeshStandardMaterial({ vertexColors: true, roughness: 0.72, metalness: 0 })
    this.material.onBeforeCompile = (shader) => skinShader(shader, this.bones, true)
    this.depthMaterial = new THREE.MeshDepthMaterial({ depthPacking: THREE.RGBADepthPacking })
    this.depthMaterial.onBeforeCompile = (shader) => skinShader(shader, this.bones, false)
    this.mesh = this.grow(64)
  }

  /** Uses the coarse body when people are only a few pixels tall. */
  setDetail(close: boolean) {
    const g = close ? this.near : this.far
    if (g === this.geometry) return
    this.geometry = g
    this.mesh.geometry = g
  }

  dispose() {
    this.mesh.dispose()
    this.near.dispose()
    this.far.dispose()
    this.material.dispose()
    this.depthMaterial.dispose()
    this.bones.value?.dispose()
    this.skirts.dispose()
  }

  private grow(cap: number) {
    this.cap = cap
    const old = this.boneData
    this.boneData = new Float32Array(cap * ROW_FLOATS)
    this.boneData.set(old.subarray(0, Math.min(old.length, this.boneData.length)))
    const tex = new THREE.DataTexture(this.boneData, ROW, cap, THREE.RGBAFormat, THREE.FloatType)
    tex.minFilter = tex.magFilter = THREE.NearestFilter
    this.bones.value?.dispose()
    this.bones.value = tex
    const mesh = new THREE.InstancedMesh(this.geometry, this.material, cap)
    mesh.customDepthMaterial = this.depthMaterial
    mesh.castShadow = true
    mesh.receiveShadow = true
    mesh.frustumCulled = false
    mesh.count = 0
    mesh.onBeforeRender = (_renderer, _scene, camera) => {
      this.eye.setFromMatrixPosition(camera.matrixWorld)
      this.hasEye = true
    }
    mesh.add(this.skirts.mesh)
    return mesh
  }

  /** Poses and places everyone; returns true if the mesh was replaced (it grew). */
  update(people: Iterable<LiveCreature>, size: number, at: Placement): boolean {
    let replaced = false
    if (size + BODY_ROOM > this.cap) {
      const old = this.mesh
      this.mesh = this.grow(Math.max(size + BODY_ROOM, this.cap * 2))
      old.parent?.add(this.mesh)
      old.removeFromParent()
      old.dispose()
      replaced = true
    }
    const dt = Math.max(0, at.dt)
    this.clock += dt
    const clock = this.clock
    const data = this.boneData
    const m = M,
      q = Q,
      s = S,
      pos = POS,
      c = COL
    const target = TARGET,
      walk = WALK,
      final = FINAL
    const k = 1 - Math.exp(-dt * 9)
    const kMood = blendFactor(dt)
    const t = at.still ? 0 : at.time
    // Up close (and drawn in detail): faces, body language, hit reactions, knockdowns as ragdolls, cloth.
    const close = at.contact === true
    const detailed = close && this.geometry === this.near
    const eye = this.hasEye ? this.eye : null
    const wind = at.wind && !at.still ? at.wind : null
    const skirts = this.skirts
    skirts.begin(detailed && !at.still)
    this.ids.length = 0
    let i = 0
    for (const cr of people) {
      let a = this.anims.get(cr.id)
      if (a) a.seen = clock
      const groundY = at.ground(cr.rx, cr.ry)
      const anatomy = bodyProportions(cr.age ?? (cr.flags & FLAG.child ? 8 : 25), hash(cr.id, 7), cr.sex === 'female')
      const size = cr.size * anatomy.height * at.scale
      if (at.inView && !at.inView(cr.rx, groundY, cr.ry, size)) {
        // Off screen: keep the walk counter current so the stride carries on when they return.
        if (a) {
          a.walked = cr.walked
          a.flags = cr.flags
          if (!(cr.flags & FLAG.fallen)) this.standUp(a)
        }
        continue
      }
      const female = cr.sex === 'female'
      const child = (cr.flags & FLAG.child) !== 0
      const flags = cr.flags
      if (!a) {
        a = {
          pose: new Float32Array(POSE_SIZE),
          walk: 0,
          run: 0,
          walked: cr.walked,
          speed: 0,
          seed: hash(cr.id, 1),
          look: 0,
          seen: clock,
          flags,
          mood: new Float32Array(MOODS),
          hitAt: -10,
          hitSide: 1,
          mode: locomotionOf(flags),
          down: null,
          rag: null,
          upAt: -1,
        }
        actionPose(actionOf(flags), a.pose, t, a.seed)
        this.anims.set(cr.id, a)
      }
      // How fast they go (tiles a second), from the distance walked.
      if (dt > 0) a.speed += ((cr.walked - a.walked) / dt - a.speed) * (1 - Math.exp(-dt * 4))
      a.walked = cr.walked
      const moving = cr.moving && !at.still
      a.walk += ((moving ? 1 : 0) - a.walk) * (1 - Math.exp(-dt * 8))
      a.run += ((moving ? THREE.MathUtils.smoothstep(a.speed, 1.6, 3.2) : 0) - a.run) * (1 - Math.exp(-dt * 4))
      const blow = (flags & FLAG.hurt) !== 0 && (a.flags & FLAG.hurt) === 0
      a.flags = flags
      if (blow) {
        a.hitAt = clock
        a.hitSide = hash(cr.id, Math.floor(clock * 10)) < 0.5 ? 1 : -1
      }
      if (close) stepMood(a.mood, cr.mood, cr.moodStrength, kMood)

      // Where the body stands: swimmers in the water, rafters on the raft's deck, everyone else (waders too) on the ground.
      const mode = locomotionOf(flags)
      let originY = groundY
      if (mode === SWIM || mode === RAFT) {
        const w = at.water?.(cr.rx, cr.ry)
        const wet = w != null && Number.isFinite(w) && w > groundY
        if (mode === RAFT) originY = wet ? w + RAFT_DECK * at.scale : groundY
        // Without water to float in, tread on the bottom instead of sinking into it.
        else originY = wet ? w : groundY + 0.4 * size
      }
      const sx = cr.size * anatomy.width * at.scale
      const sz = size * (0.96 + hash(cr.id, 9) * 0.08)

      // The pose: what they are doing (eased), blended with the stride while they walk.
      let action: Action = 'idle'
      let speak = 0
      if (mode !== a.mode) {
        // Into the water or onto a slope mid-stride: start from the stride; back on foot: ease the stride in.
        if (mode && !a.mode && a.walk > 0.01) {
          walk.set(a.pose)
          walkPose(walk, cr.walked * 3.2 * Math.PI, a.run, female, flags & FLAG.wading ? 1 : 0)
          mix(a.pose, walk, a.walk, a.pose)
        }
        if (!mode) a.walk = 0
        a.mode = mode
      }
      if (mode) {
        locomotionPose(mode, target, t, cr.walked * (mode === CLIMB ? 2.4 : 2.2) * Math.PI, a.walk, a.seed)
        for (let j = 0; j < POSE_SIZE; j++) a.pose[j] += (target[j] - a.pose[j]) * k
        final.set(a.pose)
      } else {
        action = moving ? 'idle' : actionOf(flags)
        const wade = flags & FLAG.wading ? 1 : 0
        actionPose(action, target, t, a.seed)
        if (wade && action === 'idle') wadeStill(target, t, a.seed)
        for (let j = 0; j < POSE_SIZE; j++) a.pose[j] += (target[j] - a.pose[j]) * k
        walk.set(a.pose)
        walkPose(walk, cr.walked * 3.2 * Math.PI, a.run, female, wade)
        for (let j = 0; j < POSE_SIZE; j++) final[j] = a.pose[j] + (walk[j] - a.pose[j]) * a.walk
        const busy = carryArms(target.fill(0), flags, female)
        if (busy & 1) for (const b of [ARM_L, FORE_L]) for (let d = 0; d < 3; d++) final[rot(b) + d] = target[rot(b) + d]
        if (busy & 2) for (const b of [ARM_R, FORE_R]) for (let d = 0; d < 3; d++) final[rot(b) + d] = target[rot(b) + d]
        // Talking: gestures with free hands (standing or sitting), nods, speech.
        if (flags & FLAG.talking && (action === 'idle' || action === 'sit') && !at.still)
          speak = talkOverlay(final, t, a.seed, GEST, moving || action === 'sit' ? 0 : 3 & ~busy) * (close ? 1 : 0)
      }
      // The kain follows the legs' average lean, softened.
      for (let d = 0; d < 3; d++) final[rot(SKIRT) + d] = 0.45 * (final[rot(THIGH_L) + d] + final[rot(THIGH_R) + d]) * (d === 0 ? 1 : 0)
      // The sick stoop, head hanging.
      if (flags & FLAG.ill) {
        final[rot(CHEST)] += 0.24
        final[rot(HEAD)] += 0.2
      }
      final[rot(CHEST)] += anatomy.stoop
      if (flags & FLAG.leader && !mode) {
        final[rot(CHEST)] -= 0.04
        final[rot(HEAD)] -= 0.05
      }
      // How they feel, in their bearing (up close only).
      let fear = 0
      if (close) {
        postureFrom(a.mood, POSTURES)
        fear = POSTURES[POSTURE.hunch]
        if (at.still) POSTURES[POSTURE.glance] = 0
        if (!mode) bodyLanguage(final, POSTURES, a.walk, t, a.seed, cr.walked * 3.2 * Math.PI)
      }
      // A blow: a jolt away from it and a guard, after NaturalMotion's flinch and balance.
      const since = clock - a.hitAt
      const hit = close && since < FLINCH_TIME && !at.still
      if (hit) hitReaction(final, since, a.hitSide, mode !== 0)
      let gaze = 0
      if (cr.lookX != null && cr.lookY != null && Math.hypot(cr.lookX - cr.rx, cr.lookY - cr.ry) > 0.2) {
        const angle = Math.atan2(cr.lookY - cr.ry, cr.lookX - cr.rx) - cr.rh
        gaze = -THREE.MathUtils.clamp(Math.atan2(Math.sin(angle), Math.cos(angle)), -0.8, 0.8)
      }
      a.look += (gaze - a.look) * (1 - Math.exp(-dt * 6))
      final[rot(HEAD) + 1] += a.look
      final[P_HEAD_SCALE] = anatomy.head
      final[P_BELLY] = female && flags & FLAG.pregnant ? 1 : 0.01

      // Knocked down: a ragdoll fall (or, far away and under reduced motion, straight to lying), then back up.
      const fallen = (flags & FLAG.fallen) !== 0
      if (fallen && (!a.down || a.upAt >= 0)) this.knockDown(a, cr, at, groundY, sx, size, sz, detailed && !at.still)
      if (!fallen && a.down && a.upAt < 0) a.upAt = clock
      if (a.down) {
        if (a.rag) {
          mapGround(at.ground, cr.rx, groundY, cr.ry, cr.rh, sx, size, sz)
          stepRagdoll(a.rag, bodyGround, dt)
          ragdollToPose(a.rag.pos, a.down, anatomy.head, final[P_BELLY])
          if (a.rag.asleep) {
            this.moving.delete(a.rag)
            a.rag = null
          }
        }
        if (a.upAt < 0) final.set(a.down)
        else {
          const r = (clock - a.upAt) / RISE_TIME
          if (r >= 1) this.standUp(a)
          else if (r < 0.45) mix(a.down, RISING, ease(r / 0.45), final)
          else mix(RISING, final, ease((r - 0.45) / 0.55), final)
        }
        final[P_HEAD_SCALE] = anatomy.head
        final[P_BELLY] = female && flags & FLAG.pregnant ? 1 : 0.01
      }
      if (at.contact && action !== 'sit' && action !== 'craft' && !child && !mode && !a.down) groundFeet(final, cr, at, groundY, size)
      const row = i * ROW_FLOATS
      writeBones(final, data, row)

      // Place the body.
      pos.set(cr.rx, originY, cr.ry)
      q.setFromAxisAngle(UP, Math.PI / 2 - cr.rh)
      s.set(sx, size, sz)
      m.compose(pos, q, s)
      this.mesh.setMatrixAt(i, m)

      // Colours: skin and hair by family line, clothes in the family's colour (paler when hungry).
      const hurt = cr.health < 0.25
      c.copy(SKINS[cr.id % SKINS.length])
      if (hurt) c.lerp(HURT, 0.5)
      data[row + X_SKIN] = c.r
      data[row + X_SKIN + 1] = c.g
      data[row + X_SKIN + 2] = c.b
      c.copy(HAIRS[Math.floor(cr.id / SKINS.length) % HAIRS.length])
      c.lerp(GREY, anatomy.grey)
      data[row + X_HAIR] = c.r
      data[row + X_HAIR + 1] = c.g
      data[row + X_HAIR + 2] = c.b
      data[row + X_HAIR + 3] = final[P_DRAPE]
      const sat = 0.25 + 0.45 * Math.min(1, cr.energy * 3)
      c.setHSL(cr.hue / 360, sat, 0.48)
      data[row + X_CLOTH] = c.r
      data[row + X_CLOTH + 1] = c.g
      data[row + X_CLOTH + 2] = c.b
      c.setHSL((cr.hue / 360 + 0.04) % 1, sat * 0.9, female ? 0.3 : 0.26)
      data[row + X_TRIM] = c.r
      data[row + X_TRIM + 1] = c.g
      data[row + X_TRIM + 2] = c.b

      // Cloth up close: the few nearest get a simulated sarong or kain instead of the fixed one.
      const dist = eye ? Math.hypot(cr.rx - eye.x, groundY - eye.y, cr.ry - eye.z) / size : 0
      const canWear = !mode && !a.down
      const clothed = canWear && dist < CLOTH_REACH * 1.3 ? skirts.wear(cr.id, female, dist) : null
      if (clothed) skirts.drive(clothed, m, flags & FLAG.pregnant && female ? 1 : 0, data, row, wind, dt, size, sx, groundY)
      else if (canWear && dist < CLOTH_REACH) skirts.offer(cr.id, dist, female)

      let bits = female ? SHOW.female : SHOW.male
      if (!clothed) bits |= female ? SHOW.sarong : SHOW.kain
      if (flags & FLAG.head) bits |= SHOW.chief
      if (female && hash(cr.id, 2) < 0.55) bits |= SHOW.flower
      if (flags & FLAG.carrying && mode !== SWIM) bits |= female ? SHOW.basket : SHOW.sack
      if (mode === RAFT) bits |= SHOW.paddle
      else if (mode) {
        // Hands busy swimming or climbing.
      } else if (flags & FLAG.hunting) bits |= SHOW.spear
      else if (!moving && flags & (FLAG.building | FLAG.attacking)) bits |= SHOW.hammer
      else if (!moving && flags & FLAG.planting) bits |= SHOW.stick
      else if (!moving && action === 'trade') bits |= SHOW.bundle
      data[row + X_SKIN + 3] = bits

      // The face, and hair in the wind (up close only).
      const faceOn = detailed && dist < FACE_REACH
      if (faceOn || (a.down && detailed)) {
        const blink = at.still ? 0 : blinkAt(clock, a.seed, fear)
        faceFrom(a.mood, blink, speak, FACE)
        if (hit) {
          // Pain: eyes squeezed, brows knotted (FacialData's pain clips).
          const g = guard(since)
          FACE[0] -= 0.9 * g
          FACE[2] -= 1.2 * g
          FACE[5] -= 0.6 * g
          FACE[4] += 0.25 * g
        }
        if (a.down && a.upAt < 0) {
          FACE[0] = -0.6
          FACE[2] = -0.9
        }
        for (let j = 0; j < 8; j++) data[row + X_FACE + j] = FACE[j]
        // Long hair and the headband's tails stream downwind and trail behind a walker.
        let hx = 0,
          hz = -0.006 * a.walk * (1 + a.run)
        if (wind) {
          const angle = Math.PI / 2 - cr.rh
          const flutter = 0.8 + 0.2 * Math.sin(at.time * 4.1 + a.seed * 9) + 0.1 * Math.sin(at.time * 9.3 + a.seed * 3)
          const w = Math.min(1, wind.strength) * 0.02 * flutter
          hx += (wind.dirX * Math.cos(angle) - wind.dirZ * Math.sin(angle)) * w
          hz += (wind.dirX * Math.sin(angle) + wind.dirZ * Math.cos(angle)) * w
        }
        if (at.still) hx = hz = 0
        data[row + X_CLOTH + 3] = hx
        data[row + X_TRIM + 3] = hz
      } else {
        data.set(ZERO_FACE, row + X_FACE)
        data[row + X_CLOTH + 3] = 0
        data[row + X_TRIM + 3] = 0
      }
      this.ids.push(cr.id)
      i++
    }
    // Forget people a while after they leave the stream (a body may still start from their pose).
    for (const [id, a] of this.anims)
      if (clock - a.seen > 2) {
        if (a.rag) this.moving.delete(a.rag)
        this.anims.delete(id)
      }
    skirts.end()
    this.live = i
    this.mesh.count = i
    this.mesh.instanceMatrix.needsUpdate = true
    if (this.bones.value) this.bones.value.needsUpdate = true
    return replaced
  }

  /**
   * Bodies lying where people died (frame.corpses): each new one falls as a ragdoll onto the ground and
   * stays as it settled, sinking and fading at the end. Called after update() each frame.
   */
  updateBodies(corpses: readonly CorpseFrame[] | null | undefined, at: Placement) {
    const seen = this.bodiesSeen
    seen.clear()
    const data = this.boneData
    const m = M,
      q = Q,
      s = S,
      pos = POS,
      c = COL
    const final = FINAL
    const eye = this.hasEye ? this.eye : null
    let i = this.live
    for (const cp of corpses ?? []) {
      if (i >= this.cap) break
      seen.add(cp.id)
      const female = cp.sex === 'female'
      const anatomy = bodyProportions(cp.age, hash(cp.id, 7), female)
      const size = cp.size * anatomy.height * at.scale
      const sx = cp.size * anatomy.width * at.scale
      const sz = size * (0.96 + hash(cp.id, 9) * 0.08)
      const groundY = at.ground(cp.x, cp.y)
      const visible = !at.inView || at.inView(cp.x, groundY, cp.y, size)
      mapGround(at.ground, cp.x, groundY, cp.y, cp.heading, sx, size, sz)
      let b = this.bodies.get(cp.id)
      if (!b) {
        b = this.newBody(cp, at, visible, anatomy.head)
        this.bodies.set(cp.id, b)
      }
      if (!visible) {
        // Out of sight there is nothing to watch: let it lie down at once.
        if (b.rag) this.settle(b, anatomy.head)
        continue
      }
      if (b.rag) {
        const far = eye ? Math.hypot(cp.x - eye.x, groundY - eye.y, cp.y - eye.z) / size > 25 : false
        stepRagdoll(b.rag, bodyGround, at.dt, far)
        ragdollToPose(b.rag.pos, b.pose, anatomy.head, 0.01)
        if (b.rag.asleep) {
          this.moving.delete(b.rag)
          b.rag = null
        }
      }
      final.set(b.pose)
      // At the end, the body sinks a little into the ground and fades towards its colour.
      const fade = THREE.MathUtils.clamp((cp.seconds - BODY_FADE) / (BODY_LIFE - BODY_FADE), 0, 1)
      final[P_ROOT_Y] -= 0.06 * fade
      const row = i * ROW_FLOATS
      writeBones(final, data, row)
      pos.set(cp.x, groundY, cp.y)
      q.setFromAxisAngle(UP, Math.PI / 2 - cp.heading)
      s.set(sx, size, sz)
      m.compose(pos, q, s)
      this.mesh.setMatrixAt(i, m)

      const pale = 0.12 + 0.5 * fade
      c.copy(SKINS[cp.id % SKINS.length])
      if (cp.cause === 'burned') c.lerp(SOOT, 0.55)
      c.lerp(EARTH, pale)
      data[row + X_SKIN] = c.r
      data[row + X_SKIN + 1] = c.g
      data[row + X_SKIN + 2] = c.b
      c.copy(HAIRS[Math.floor(cp.id / SKINS.length) % HAIRS.length])
      c.lerp(GREY, anatomy.grey)
      c.lerp(EARTH, pale)
      data[row + X_HAIR] = c.r
      data[row + X_HAIR + 1] = c.g
      data[row + X_HAIR + 2] = c.b
      data[row + X_HAIR + 3] = 1
      c.setHSL(cp.hue / 360, 0.3, 0.44)
      if (cp.cause === 'burned') c.lerp(SOOT, 0.55)
      c.lerp(EARTH, pale)
      data[row + X_CLOTH] = c.r
      data[row + X_CLOTH + 1] = c.g
      data[row + X_CLOTH + 2] = c.b
      c.setHSL((cp.hue / 360 + 0.04) % 1, 0.27, female ? 0.3 : 0.26)
      c.lerp(EARTH, pale)
      data[row + X_TRIM] = c.r
      data[row + X_TRIM + 1] = c.g
      data[row + X_TRIM + 2] = c.b
      let bits = female ? SHOW.female | SHOW.sarong : SHOW.male | SHOW.kain
      if (female && hash(cp.id, 2) < 0.55) bits |= SHOW.flower
      data[row + X_SKIN + 3] = bits
      // Eyes closed, the face at rest.
      data.set(ZERO_FACE, row + X_FACE)
      data[row + X_FACE + 7] = 1
      data[row + X_CLOTH + 3] = 0
      data[row + X_TRIM + 3] = 0
      i++
    }
    for (const [id, b] of this.bodies)
      if (!seen.has(id)) {
        if (b.rag) this.moving.delete(b.rag)
        this.bodies.delete(id)
      }
    this.mesh.count = i
    this.mesh.instanceMatrix.needsUpdate = true
    if (this.bones.value) this.bones.value.needsUpdate = true
  }

  /** A body seen for the first time: from the pose its person was last drawn in (or standing), falling or already lying. */
  private newBody(cp: CorpseFrame, at: Placement, visible: boolean, head: number): Body {
    const seed = hash(cp.id, 11)
    const start = new Float32Array(POSE_SIZE)
    const person = this.anims.get(cp.id)
    if (person && this.clock - person.seen < 1.5) start.set(person.down ?? person.pose)
    else standing(start, 0, seed)
    start[P_HEAD_SCALE] = head
    start[P_BELLY] = 0.01
    const b: Body = { rag: createRagdoll(posePoints(start, POINTS), seed), pose: start, seed }
    if (person?.rag) this.moving.delete(person.rag)
    const fresh = cp.seconds <= FRESH_BODY && !at.still && visible && this.moving.size < MAX_RAGDOLLS
    if (fresh) this.moving.add(b.rag!)
    else this.settle(b, head)
    if (b.rag) ragdollToPose(b.rag.pos, b.pose, head, 0.01)
    return b
  }

  /** Runs a body's ragdoll to rest at once (the ground must be mapped for it). */
  private settle(b: Body, head: number) {
    if (!b.rag) return
    settleRagdoll(b.rag, bodyGround)
    ragdollToPose(b.rag.pos, b.pose, head, 0.01)
    this.moving.delete(b.rag)
    b.rag = null
  }

  /** A slip: down onto the ground (falling as a ragdoll when there is one to spare). */
  private knockDown(a: Anim, cr: LiveCreature, at: Placement, groundY: number, sx: number, sy: number, sz: number, physics: boolean) {
    if (a.rag) this.moving.delete(a.rag)
    a.rag = null
    a.upAt = -1
    a.down ??= new Float32Array(POSE_SIZE)
    mapGround(at.ground, cr.rx, groundY, cr.ry, cr.rh, sx, sy, sz)
    // Downhill if there is a slope, else backwards (feet slipping out forward).
    const gx = bodyGround(0.1, 0) - bodyGround(-0.1, 0),
      gz = bodyGround(0, 0.1) - bodyGround(0, -0.1)
    const push = Math.hypot(gx, gz) > 0.02 ? { x: -gx, z: -gz, strength: 0.7 } : { x: 0.2 * (a.seed - 0.5), z: -1, strength: 0.7 }
    if (physics && this.moving.size < MAX_RAGDOLLS) {
      a.rag = createRagdoll(posePoints(a.pose, POINTS), a.seed, push)
      this.moving.add(a.rag)
      ragdollToPose(a.rag.pos, a.down, a.pose[P_HEAD_SCALE] || 1, 0.01)
    } else if (at.still) {
      const r = createRagdoll(posePoints(a.pose, POINTS), a.seed, push)
      settleRagdoll(r, bodyGround)
      ragdollToPose(r.pos, a.down, 1, 0.01)
    } else a.down.set(lyingPose())
  }

  private standUp(a: Anim) {
    if (a.rag) this.moving.delete(a.rag)
    a.rag = null
    a.down = null
    a.upAt = -1
  }
}

/** Mood in the body's bearing (POSTURE weights): fear hunches and glances about, anger squares up, grief slumps, joy lifts. */
function bodyLanguage(p: Float32Array, w: Float32Array, walking: number, t: number, seed: number, phi: number) {
  const fear = w[POSTURE.hunch],
    anger = w[POSTURE.chest],
    grief = w[POSTURE.slump],
    joy = w[POSTURE.lift]
  if (fear + anger + grief + joy < 0.01) return
  // Swinging arms: bigger when happy, barely when grieving.
  const swing = 1 + (0.35 * joy - 0.6 * grief) * walking
  p[rot(ARM_L)] *= swing
  p[rot(ARM_R)] *= swing
  p[rot(CHEST)] += 0.17 * fear - 0.09 * anger + 0.16 * grief - 0.06 * joy
  p[rot(HEAD)] += 0.05 * fear + 0.12 * anger + 0.3 * grief - 0.12 * joy
  // Arms drawn in and hands up when afraid; held out, elbows bent, when angry.
  p[rot(ARM_L) + 2] += -0.07 * fear + 0.12 * anger - 0.05 * grief
  p[rot(ARM_R) + 2] -= -0.07 * fear + 0.12 * anger - 0.05 * grief
  p[rot(ARM_L)] -= 0.12 * fear
  p[rot(ARM_R)] -= 0.12 * fear
  p[rot(FORE_L)] -= 0.4 * fear + 0.45 * anger - 0.1 * grief
  p[rot(FORE_R)] -= 0.4 * fear + 0.45 * anger - 0.1 * grief
  // Knees soft and ready when afraid; a wider stance when angry.
  const still = 1 - walking
  p[rot(THIGH_L)] -= 0.07 * fear * still
  p[rot(THIGH_R)] -= 0.07 * fear * still
  p[rot(SHIN_L)] += 0.14 * fear * still
  p[rot(SHIN_R)] += 0.14 * fear * still
  p[rot(THIGH_L) + 2] += 0.05 * anger * still
  p[rot(THIGH_R) + 2] -= 0.05 * anger * still
  p[P_ROOT_Y] -= (0.007 * fear + 0.004 * grief) * still
  // A lighter step when happy.
  p[P_ROOT_Y] += 0.007 * joy * walking * Math.max(0, Math.sin(phi * 2))
  // Quick glances about when afraid: the head darts and holds.
  const dart = Math.sin(t * 2.9 + seed * 7)
  p[rot(HEAD) + 1] +=
    w[POSTURE.glance] * 0.5 * Math.sign(dart) * Math.min(1, Math.abs(dart) * 3) * (0.5 + 0.5 * Math.sin(t * 0.7 + seed * 3))
}

/** The flinch and stagger after a blow, blended over the current pose (TaskNMFlinch, then balance). */
function hitReaction(p: Float32Array, since: number, side: number, afloat: boolean) {
  const jolt = flinch(since),
    g = guard(since)
  p[rot(CHEST)] -= 0.32 * jolt
  p[rot(CHEST) + 2] += 0.16 * side * jolt
  p[rot(HEAD)] -= 0.28 * flinch(since - 0.04)
  p[rot(HEAD) + 2] -= 0.12 * side * jolt
  p[rot(HIPS) + 2] -= 0.07 * side * jolt
  // Arms up to cover, knees soft, a step back.
  p[rot(ARM_L)] -= 0.55 * g
  p[rot(ARM_R)] -= 0.55 * g
  p[rot(ARM_L) + 2] -= 0.1 * g
  p[rot(ARM_R) + 2] += 0.1 * g
  p[rot(FORE_L)] -= 1.0 * g
  p[rot(FORE_R)] -= 1.0 * g
  if (afloat) return
  p[rot(THIGH_L)] -= 0.12 * g
  p[rot(THIGH_R)] -= 0.12 * g
  p[rot(SHIN_L)] += 0.24 * g
  p[rot(SHIN_R)] += 0.24 * g
  p[P_ROOT_Y] -= 0.012 * g
  p[P_ROOT_Z] -= 0.03 * g
}

/** out = a + (b − a)·k (out may be b). */
function mix(a: Float32Array, b: Float32Array, k: number, out: Float32Array) {
  for (let j = 0; j < POSE_SIZE; j++) out[j] = a[j] + (b[j] - a[j]) * k
}

const UP = new THREE.Vector3(0, 1, 0)
const GREY = new THREE.Color('#afa99b')
const euler = new THREE.Euler()
const local = new THREE.Matrix4()
const scaleV = new THREE.Vector3()
const world = Array.from({ length: NB }, () => new THREE.Matrix4())
const ikMatrices = new Float32Array(NB * 16)
const SCRATCH_BONES = new Float32Array(NB * 16)
const ikPoint = new THREE.Vector3()
const ikInverse = new THREE.Matrix4()
/** The bone each ragdoll particle rides on (RP order). */
const RAG_BONE = [HEAD, CHEST, CHEST, CHEST, ARM_L, ARM_R, FORE_L, FORE_R, HIPS, HIPS, HIPS, THIGH_L, THIGH_R, SHIN_L, SHIN_R, CHEST, CHEST]

/** Local contact correction, after the action/locomotion blend. The gait's
 * lifted foot stays lifted; the stance foot follows actual ground height. */
function groundFeet(p: Float32Array, cr: LiveCreature, at: Placement, ground: number, size: number) {
  writeBones(p, ikMatrices, 0)
  const angle = Math.PI / 2 - cr.rh,
    cs = Math.cos(angle),
    sn = Math.sin(angle)
  const targets: { thigh: number; shin: number; x: number; y: number; z: number; correction: number }[] = []
  for (const [thigh, shin, side] of [
    [THIGH_L, SHIN_L, 1],
    [THIGH_R, SHIN_R, -1],
  ]) {
    ikPoint.set(side * 0.038, 0.026, 0).applyMatrix4(world[shin])
    const x = cr.rx + (ikPoint.x * cs + ikPoint.z * sn) * size
    const z = cr.ry + (-ikPoint.x * sn + ikPoint.z * cs) * size
    const floor = (at.ground(x, z) - ground) / size + 0.026
    const correction = THREE.MathUtils.clamp(floor - ikPoint.y, -0.06, 0.07)
    targets.push({ thigh, shin, x: ikPoint.x, y: ikPoint.y, z: ikPoint.z, correction })
  }
  // A small pelvis drop keeps both feet within the skeleton's reach.
  const drop = Math.max(-0.045, Math.min(0, ...targets.map((t) => t.correction)))
  p[P_ROOT_Y] += drop
  writeBones(p, ikMatrices, 0)
  ikInverse.copy(world[HIPS]).invert()
  for (const t of targets) {
    const weight = cr.moving ? THREE.MathUtils.clamp((0.075 - t.y) / 0.045, 0, 1) : 1
    if (weight <= 0) continue
    ikPoint.set(t.x, t.y + t.correction, t.z).applyMatrix4(ikInverse)
    const hip = PIVOT[t.thigh]
    const angles = solveLeg(ikPoint.y - hip[1], ikPoint.z - hip[2])
    p[rot(t.thigh)] += (angles.hip - p[rot(t.thigh)]) * weight
    p[rot(t.shin)] += (angles.knee - p[rot(t.shin)]) * weight
  }
}

/** Bone matrices (rest pose → posed, in body space) from a pose vector, into the bone texture row. */
function writeBones(p: Float32Array, out: Float32Array, offset: number) {
  for (let b = 0; b < NB; b++) {
    euler.set(p[rot(b)], p[rot(b) + 1], p[rot(b) + 2], 'XYZ')
    local.makeRotationFromEuler(euler)
    const sc = b === HEAD ? p[P_HEAD_SCALE] : b === BELLY ? p[P_BELLY] : 1
    if (sc !== 1) local.scale(scaleV.set(sc, sc, sc))
    // Turn about the joint: T(pivot) · R · T(-pivot).
    const e = local.elements
    const [px, py, pz] = PIVOT[b]
    e[12] = px - (e[0] * px + e[4] * py + e[8] * pz)
    e[13] = py - (e[1] * px + e[5] * py + e[9] * pz)
    e[14] = pz - (e[2] * px + e[6] * py + e[10] * pz)
    if (b === HIPS) {
      e[12] += p[P_ROOT_X]
      e[13] += p[P_ROOT_Y]
      e[14] += p[P_ROOT_Z]
      world[b].copy(local)
    } else {
      world[b].multiplyMatrices(world[PARENT[b]], local)
    }
    out.set(world[b].elements, offset + b * 16)
  }
}

// --- Cloth --------------------------------------------------------------------------

const SKIRT_COLS = 16
const SKIRT_ROWS = 8
/** At most this many people wear simulated cloth at once (the nearest). */
const SKIRT_MAX = 6
const SKIRT_VERTS = SKIRT_COLS * SKIRT_ROWS
const SARONG = skirtShape(SKIRT_COLS, SKIRT_ROWS, 0.322, 0.074, 0.052, 0.09)
const KAIN = skirtShape(SKIRT_COLS, SKIRT_ROWS, 0.298, 0.19, 0.057, 0.082)

type Skirt = { id: number; female: boolean; cloth: Cloth; used: boolean; d: number }

const capsuleBuffer = new Float64Array(7 * 6)
const pinBuffer = new Float64Array(SKIRT_COLS * 3)
const shapeBuffer = new Float64Array(SKIRT_VERTS * 3)
const forces: ClothForces = { pins: pinBuffer, capsules: capsuleBuffer, capsuleCount: 0, windX: 0, windZ: 0, scale: 1, floor: 0 }
const boneM = new THREE.Matrix4()
const toWorld = new THREE.Matrix4()
const pa = new THREE.Vector3()
const pb = new THREE.Vector3()

/**
 * The few nearest people's sarongs and kains as Verlet cloth, drawn as one mesh in world space
 * (a child of the crowd's mesh). People are offered each frame; the nearest get cloth next frame.
 */
class Skirts {
  readonly mesh: THREE.Mesh
  private readonly slots: Skirt[] = []
  private offers: { id: number; d: number; female: boolean }[] = []
  private enabled = false
  private readonly positions: THREE.BufferAttribute
  private readonly colors: THREE.BufferAttribute
  private readonly rest: THREE.BufferAttribute
  private used = 0

  constructor() {
    const geometry = new THREE.BufferGeometry()
    this.positions = new THREE.BufferAttribute(new Float32Array(SKIRT_MAX * SKIRT_VERTS * 3), 3)
    this.positions.setUsage(THREE.DynamicDrawUsage)
    this.colors = new THREE.BufferAttribute(new Float32Array(SKIRT_MAX * SKIRT_VERTS * 3), 3)
    this.colors.setUsage(THREE.DynamicDrawUsage)
    geometry.setAttribute('position', this.positions)
    geometry.setAttribute('color', this.colors)
    // Where each vertex sits on the garment at rest (body units), for the same weave as the fixed garment.
    this.rest = new THREE.BufferAttribute(new Float32Array(SKIRT_MAX * SKIRT_VERTS * 3), 3)
    this.rest.setUsage(THREE.DynamicDrawUsage)
    geometry.setAttribute('rest', this.rest)
    const index: number[] = []
    for (let s = 0; s < SKIRT_MAX; s++)
      for (let r = 0; r + 1 < SKIRT_ROWS; r++)
        for (let c = 0; c < SKIRT_COLS; c++) {
          const v = (rr: number, cc: number) => s * SKIRT_VERTS + rr * SKIRT_COLS + (cc % SKIRT_COLS)
          index.push(v(r, c), v(r + 1, c), v(r, c + 1), v(r, c + 1), v(r + 1, c), v(r + 1, c + 1))
        }
    geometry.setIndex(index)
    geometry.setDrawRange(0, 0)
    const material = new THREE.MeshStandardMaterial({
      vertexColors: true,
      roughness: 0.82,
      side: THREE.DoubleSide,
      shadowSide: THREE.BackSide,
    })
    material.onBeforeCompile = (shader) => {
      shader.vertexShader = shader.vertexShader.replace('#include <common>', '#include <common>\nattribute vec3 rest;').replace(
        '#include <color_vertex>',
        `#include <color_vertex>
          float weave = 0.96 + 0.04 * sin(rest.y * 480.0) * sin(rest.x * 410.0);
          float band = 0.9 + 0.1 * smoothstep(-0.3, 0.3, sin(rest.y * 155.0));
          vColor.rgb *= weave * band;`,
      )
    }
    this.mesh = new THREE.Mesh(geometry, material)
    this.mesh.frustumCulled = false
    this.mesh.castShadow = true
    this.mesh.receiveShadow = true
  }

  begin(enabled: boolean) {
    this.enabled = enabled
    this.offers.length = 0
    this.used = 0
    for (const s of this.slots) s.used = false
    if (!enabled) this.slots.length = 0
  }

  /** The cloth this person (at `d` body heights from the camera) wears this frame, if any. */
  wear(id: number, female: boolean, d: number): Skirt | null {
    if (!this.enabled) return null
    for (const s of this.slots)
      if (s.id === id && s.female === female && !s.used) {
        s.used = true
        s.d = d
        return s
      }
    return null
  }

  offer(id: number, d: number, female: boolean) {
    if (this.enabled) this.offers.push({ id, d, female })
  }

  /** Pins the cloth to the hips (bone row `row` of `data`, placed by `place`), collides it with the legs, steps it, draws it. */
  drive(
    s: Skirt,
    place: THREE.Matrix4,
    pregnant: number,
    data: Float32Array,
    row: number,
    wind: Placement['wind'] | null,
    dt: number,
    size: number,
    width: number,
    floor: number,
  ) {
    const shape = s.female ? SARONG : KAIN
    const bone = (b: number) => toWorld.multiplyMatrices(place, boneM.fromArray(data, row + b * 16))
    bone(HIPS)
    for (let k = 0; k < SKIRT_COLS; k++) {
      pa.set(shape[k * 3], shape[k * 3 + 1], shape[k * 3 + 2]).applyMatrix4(toWorld)
      pinBuffer[k * 3] = pa.x
      pinBuffer[k * 3 + 1] = pa.y
      pinBuffer[k * 3 + 2] = pa.z
    }
    // Teleported or brand new: hang it in its rest shape first (a force-pin).
    const moved = Math.hypot(s.cloth.pos[0] - pinBuffer[0], s.cloth.pos[1] - pinBuffer[1], s.cloth.pos[2] - pinBuffer[2])
    if (moved > 0.3 * size) {
      for (let v = 0; v < SKIRT_VERTS; v++) {
        pa.set(shape[v * 3], shape[v * 3 + 1], shape[v * 3 + 2]).applyMatrix4(toWorld)
        shapeBuffer[v * 3] = pa.x
        shapeBuffer[v * 3 + 1] = pa.y
        shapeBuffer[v * 3 + 2] = pa.z
      }
      resetCloth(s.cloth, shapeBuffer)
    }
    // The hips, thighs, shins (and a baby on the way) as capsules.
    let n = 0
    const capsule = (b: number, ax: number, ay: number, az: number, bx: number, by: number, bz: number, r: number) => {
      bone(b)
      pa.set(ax, ay, az).applyMatrix4(toWorld)
      pb.set(bx, by, bz).applyMatrix4(toWorld)
      capsuleBuffer.set([pa.x, pa.y, pa.z, pb.x, pb.y, pb.z, r * width], n * 7)
      n++
    }
    capsule(HIPS, 0.026, 0.262, 0, -0.026, 0.262, 0, 0.046)
    for (const [thigh, shin, x] of [
      [THIGH_L, SHIN_L, 0.038],
      [THIGH_R, SHIN_R, -0.038],
    ]) {
      capsule(thigh, x, 0.245, 0, x, 0.14, 0, 0.037)
      capsule(shin, x, 0.135, 0, x, 0.04, 0, 0.03)
    }
    if (pregnant) capsule(BELLY, 0, 0.285, 0.022, 0, 0.29, 0.022, 0.055)
    forces.capsuleCount = n
    forces.scale = size
    forces.floor = floor
    const w = wind ? Math.min(1, wind.strength) * 1.1 * size : 0
    forces.windX = wind ? wind.dirX * w : 0
    forces.windZ = wind ? wind.dirZ * w : 0
    stepCloth(s.cloth, forces, dt)
    // Into the mesh, in the wearer's colours (the hem in the trim's).
    const base = this.used * SKIRT_VERTS
    const p = this.positions.array as Float32Array,
      col = this.colors.array as Float32Array,
      rest = this.rest.array as Float32Array
    for (let v = 0; v < SKIRT_VERTS; v++) {
      const o = (base + v) * 3
      p[o] = s.cloth.pos[v * 3]
      p[o + 1] = s.cloth.pos[v * 3 + 1]
      p[o + 2] = s.cloth.pos[v * 3 + 2]
      rest[o] = shape[v * 3]
      rest[o + 1] = shape[v * 3 + 1]
      rest[o + 2] = shape[v * 3 + 2]
      const src = row + (v >= (SKIRT_ROWS - 1) * SKIRT_COLS ? X_TRIM : X_CLOTH)
      col[o] = data[src]
      col[o + 1] = data[src + 1]
      col[o + 2] = data[src + 2]
    }
    this.used++
  }

  /** Draws this frame's cloth and picks who wears it next frame: the nearest offered, keeping current wearers. */
  end() {
    const geometry = this.mesh.geometry
    geometry.setDrawRange(0, this.used * (SKIRT_ROWS - 1) * SKIRT_COLS * 6)
    if (this.used) {
      this.positions.needsUpdate = true
      this.colors.needsUpdate = true
      this.rest.needsUpdate = true
      geometry.computeVertexNormals()
    }
    if (!this.enabled) return
    // Wearers not drawn this frame give their cloth up; the nearest new offers take free places.
    for (let k = this.slots.length - 1; k >= 0; k--) if (!this.slots[k].used) this.slots.splice(k, 1)
    if (!this.offers.length) return
    this.offers.sort((a, b) => a.d - b.d)
    for (const o of this.offers) {
      if (this.slots.length >= SKIRT_MAX) {
        // Take over from a wearer only if they are much further away: no flickering between two.
        let far = 0
        for (let k = 1; k < this.slots.length; k++) if (this.slots[k].d > this.slots[far].d) far = k
        if (!(o.d < this.slots[far].d * 0.7)) break
        this.slots.splice(far, 1)
      }
      // Starts hung in its rest shape at the first drive (it is "teleported" from nowhere).
      const shape = o.female ? SARONG : KAIN
      const cloth = createCloth(SKIRT_COLS, SKIRT_ROWS, shape, new Float64Array(SKIRT_VERTS * 3).fill(1e9))
      this.slots.push({ id: o.id, female: o.female, cloth, used: true, d: o.d })
    }
  }

  dispose() {
    this.mesh.geometry.dispose()
    ;(this.mesh.material as THREE.Material).dispose()
  }
}

/** Skinning from the bone texture, outfit/prop visibility, face expressions and hair sway, colours from the person's palette. */
function skinShader(shader: THREE.WebGLProgramParametersWithUniforms, bones: { value: THREE.DataTexture | null }, lit: boolean) {
  shader.uniforms.uBones = bones
  shader.vertexShader = shader.vertexShader
    .replace(
      '#include <common>',
      `#include <common>
      uniform highp sampler2D uBones;
      attribute vec3 part;
      attribute vec4 feat;
      mat4 boneMatrix(int b) {
        int x = b * 4;
        return mat4(
          texelFetch(uBones, ivec2(x, gl_InstanceID), 0),
          texelFetch(uBones, ivec2(x + 1, gl_InstanceID), 0),
          texelFetch(uBones, ivec2(x + 2, gl_InstanceID), 0),
          texelFetch(uBones, ivec2(x + 3, gl_InstanceID), 0));
      }
      vec4 extra(int k) { return texelFetch(uBones, ivec2(${NB * 4} + k, gl_InstanceID), 0); }
      // Moves a face feature (rest pose, head space) for the person's expression:
      // f0 = brows inner/outer, eyes open, smile; f1 = jaw, lips stretched/pressed, cheeks, blink.
      vec3 expression(vec3 p, vec3 c, int kind) {
        vec3 r = p - c;
        if (kind == ${F_SKIRT}) {
          // A body lying down: the rigid skirt's hem falls in around the legs instead of standing out.
          float drape = extra(1).w * clamp(-r.y / 0.12, 0.0, 1.0);
          p.xz = c.xz + r.xz * (1.0 - 0.42 * drape);
          return p;
        }
        if (kind >= ${F_HAIR}) {
          // Hair and headband tails: further from the root, further in the wind.
          float reach = clamp((c.y - p.y) / (kind == ${F_HAIR} ? 0.085 : 0.03), 0.0, 1.6);
          vec2 sway = vec2(extra(2).w, extra(3).w) * reach;
          // Blown hair lifts a little as it streams.
          p += vec3(sway.x, length(sway) * 0.45, sway.y);
          return p;
        }
        vec4 f0 = extra(4);
        vec4 f1 = extra(5);
        float side = c.x >= 0.0 ? 1.0 : -1.0;
        float open = clamp(1.0 + 0.45 * f0.z, 0.3, 1.45) * (1.0 - f1.w);
        if (kind == ${F_BROW}) {
          float inner = clamp(0.5 - side * r.x / 0.016, 0.0, 1.0);
          p.y += mix(f0.y, f0.x, inner) * 0.0042 - max(0.0, -f0.z) * 0.0007 - max(0.0, -f0.x) * 0.0008;
          p.x -= side * max(0.0, -f0.x) * 0.0024 * inner;
        } else if (kind == ${F_LID}) {
          float shut = 1.0 - min(open, 1.0);
          p.y += (open - 1.0) * 0.0034 + r.y * shut * 1.3;
          p.z += shut * 0.0016;
        } else if (kind == ${F_EYE}) {
          // Closing from the top; raised cheeks push up from below.
          float h = 0.0033;
          float bottom = -h + f1.z * 0.0012;
          float top = max(-h + 2.0 * h * open, bottom + 0.0002);
          p.y = c.y + mix(bottom, top, clamp((r.y + h) / (2.0 * h), 0.0, 1.0));
        } else if (kind == ${F_IRIS}) {
          p.y = c.y + r.y * min(open, 1.0) * (1.0 - 0.3 * f1.z) + f1.z * 0.0004;
        } else if (kind == ${F_MOUTH}) {
          float xn = clamp(r.x / 0.0096, -1.0, 1.0);
          float press = max(0.0, -f1.y);
          p.x = c.x + r.x * (1.0 + 0.28 * max(0.0, f1.y) + 0.12 * max(0.0, f0.w) - 0.12 * press);
          p.y += f0.w * (xn * xn - 0.35) * 0.0045;
          p.y += r.y * (f1.x * 3.0 - press * 0.6) - f1.x * 0.002;
        } else if (kind == ${F_CHEEK}) {
          p.y += f1.z * 0.0017;
          p.z += f1.z * 0.0009;
        } else if (kind == ${F_CHIN}) {
          p.y -= f1.x * 0.0024;
        }
        return p;
      }`,
    )
    .replace(
      '#include <begin_vertex>',
      `#include <begin_vertex>
      int kind = int(feat.w + 0.5);
      if (kind > 0) transformed = expression(transformed, feat.xyz, kind);
      mat4 boneM = boneMatrix(int(part.x + 0.5));
      transformed = (boneM * vec4(transformed, 1.0)).xyz;
      int partBit = int(part.z + 0.5);
      if (partBit != 0 && (partBit & int(extra(0).w + 0.5)) == 0) transformed = vec3(0.0);`,
    )
  if (!lit) return
  shader.vertexShader = shader.vertexShader
    .replace(
      '#include <beginnormal_vertex>',
      `#include <beginnormal_vertex>
      objectNormal = normalize(mat3(boneMatrix(int(part.x + 0.5))) * objectNormal);`,
    )
    .replace(
      '#include <color_vertex>',
      `#include <color_vertex>
      int region = int(part.y + 0.5);
      if (region > 0) vColor.rgb = extra(region - 1).rgb;
      if (region > 2) {
        float weave = 0.96 + 0.04 * sin(position.y * 480.0) * sin(position.x * 410.0);
        float band = 0.9 + 0.1 * smoothstep(-0.3, 0.3, sin(position.y * 155.0));
        vColor.rgb *= weave * band;
      }
      // An open mouth shows dark inside.
      if (int(feat.w + 0.5) == ${F_MOUTH}) vColor.rgb *= 1.0 - 0.55 * extra(5).x;`,
    )
}
