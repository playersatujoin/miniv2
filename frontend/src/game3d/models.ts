// Low-poly models built from a few primitives each, with colours baked into
// the vertices so one InstancedMesh can draw every tree (or deer, or person)
// of a kind in a single call. Sizes are in tiles: a person stands about half
// a tile tall, a tree a little over one.

import * as THREE from 'three'
import { mergeGeometries } from 'three/examples/jsm/utils/BufferGeometryUtils.js'

/** `leg`: ±1 for a leg (which pair of a trot it swings with) hinged at (`hip`[0], `hip`[1]) in x and y; 0 for the rest. */
type Part = { geo: THREE.BufferGeometry; color: string; leg?: number; hip?: [number, number] }

/** One primitive in a colour, moved into place. */
function part(geo: THREE.BufferGeometry, color: string, at: [number, number, number] = [0, 0, 0], rot: [number, number, number] = [0, 0, 0], scale: [number, number, number] = [1, 1, 1]): Part {
  const g = geo.index ? geo.toNonIndexed() : geo
  const m = new THREE.Matrix4().compose(
    new THREE.Vector3(...at),
    new THREE.Quaternion().setFromEuler(new THREE.Euler(...rot)),
    new THREE.Vector3(...scale),
  )
  g.applyMatrix4(m)
  return { geo: g, color }
}

/** Merges coloured parts into one flat-shaded geometry. */
function model(parts: Part[]) {
  const legged = parts.some((p) => p.leg)
  const geos = parts.map(({ geo, color, leg = 0, hip = [0, 0] }) => {
    const c = new THREE.Color(color)
    const n = geo.attributes.position.count
    const cols = new Float32Array(n * 3)
    for (let i = 0; i < n; i++) {
      cols[i * 3] = c.r
      cols[i * 3 + 1] = c.g
      cols[i * 3 + 2] = c.b
    }
    const g = geo.clone()
    g.deleteAttribute('uv')
    g.setAttribute('color', new THREE.BufferAttribute(cols, 3))
    if (legged) {
      g.setAttribute('leg', new THREE.BufferAttribute(new Float32Array(n).fill(leg), 1))
      const hips = new Float32Array(n * 2)
      for (let i = 0; i < n; i++) hips.set(hip, i * 2)
      g.setAttribute('hip', new THREE.BufferAttribute(hips, 2))
    }
    return g
  })
  const merged = mergeGeometries(geos, false)!
  merged.computeVertexNormals()
  return merged
}

// --- Plants and rocks ---------------------------------------------------------

export function treeGeometry() {
  return model([
    part(new THREE.CylinderGeometry(0.05, 0.08, 0.6, 6), '#6b4a2b', [0, 0.3, 0]),
    part(new THREE.IcosahedronGeometry(0.42, 0), '#3f8f3a', [0, 0.85, 0], [0.3, 0.5, 0], [1, 0.85, 1]),
    part(new THREE.IcosahedronGeometry(0.3, 0), '#4fa043', [0.18, 1.1, 0.06], [0.7, 0.1, 0.4]),
    part(new THREE.IcosahedronGeometry(0.26, 0), '#367f33', [-0.2, 0.98, -0.1], [0.2, 0.9, 0.1]),
  ])
}

export function pineGeometry() {
  return model([
    part(new THREE.CylinderGeometry(0.04, 0.07, 0.45, 6), '#5d3f24', [0, 0.22, 0]),
    part(new THREE.ConeGeometry(0.42, 0.62, 7), '#2b6a43', [0, 0.65, 0]),
    part(new THREE.ConeGeometry(0.32, 0.52, 7), '#327a4c', [0, 0.98, 0], [0, 0.4, 0]),
    part(new THREE.ConeGeometry(0.2, 0.4, 7), '#3a8a55', [0, 1.26, 0], [0, 0.8, 0]),
  ])
}

export function bushGeometry() {
  return model([
    part(new THREE.IcosahedronGeometry(0.24, 0), '#3f8f3a', [0, 0.18, 0], [0.2, 0.3, 0], [1.2, 0.8, 1.1]),
    part(new THREE.IcosahedronGeometry(0.17, 0), '#4c9e3f', [0.18, 0.15, 0.08]),
    part(new THREE.IcosahedronGeometry(0.15, 0), '#c0392b', [-0.05, 0.32, 0.12], [0, 0, 0], [0.35, 0.35, 0.35]),
  ])
}

export function rockGeometry() {
  return model([
    part(new THREE.DodecahedronGeometry(0.3, 0), '#8c8a86', [0, 0.14, 0], [0.4, 0.7, 0.2], [1.2, 0.7, 1]),
    part(new THREE.DodecahedronGeometry(0.15, 0), '#7a7874', [0.25, 0.08, 0.1], [0.9, 0.2, 0.5]),
  ])
}

export function flowerGeometry() {
  const parts: Part[] = []
  const colors = ['#e98ad0', '#f6d743', '#ffffff', '#b38ae9', '#ff8a65']
  for (let k = 0; k < 5; k++) {
    const a = (k / 5) * Math.PI * 2
    parts.push(part(new THREE.CylinderGeometry(0.01, 0.01, 0.14, 3), '#3e8e3a', [Math.cos(a) * 0.2, 0.07, Math.sin(a) * 0.2]))
    parts.push(part(new THREE.OctahedronGeometry(0.05, 0), colors[k], [Math.cos(a) * 0.2, 0.16, Math.sin(a) * 0.2]))
  }
  return model(parts)
}

/** Reeds at the water's edge: tall leaning blades and a couple of brown cattails. */
export function reedGeometry() {
  const parts: Part[] = []
  for (let k = 0; k < 7; k++) {
    const a = (k / 7) * Math.PI * 2 + k
    const h = 0.32 + 0.16 * (((k * 37) % 7) / 7)
    const lean = 0.12 + 0.1 * (((k * 13) % 5) / 5)
    parts.push(part(new THREE.ConeGeometry(0.012, h, 3), k % 3 ? '#5f8f3a' : '#7da64a', [Math.cos(a) * 0.05, h / 2, Math.sin(a) * 0.05], [Math.sin(a) * lean, 0, -Math.cos(a) * lean]))
  }
  for (const [x, z, h] of [
    [0.02, -0.01, 0.42],
    [-0.03, 0.02, 0.36],
  ]) {
    parts.push(part(new THREE.CylinderGeometry(0.004, 0.004, h, 3), '#6f8a3a', [x, h / 2, z]))
    parts.push(part(new THREE.CapsuleGeometry(0.016, 0.06, 2, 5), '#6b4126', [x, h - 0.02, z]))
  }
  return model(parts)
}

/** A bird gliding along +x, wings spread along ±z (a shader flaps them). */
export function birdGeometry() {
  const pos = [
    // Body: a slim diamond.
    0.16, 0, 0, -0.12, 0.02, 0, 0, 0, 0.035,
    0.16, 0, 0, 0, 0, -0.035, -0.12, 0.02, 0,
    // Wings, swept back a little.
    0.05, 0.01, 0, -0.04, 0.01, 0, -0.06, 0.02, 0.3,
    0.05, 0.01, 0, -0.06, 0.02, -0.3, -0.04, 0.01, 0,
    // Tail.
    -0.1, 0.02, 0, -0.2, 0.02, 0.05, -0.2, 0.02, -0.05,
  ]
  const g = new THREE.BufferGeometry()
  g.setAttribute('position', new THREE.Float32BufferAttribute(pos, 3))
  g.setAttribute('color', new THREE.Float32BufferAttribute(new Array(pos.length).fill(1), 3))
  g.computeVertexNormals()
  return g
}

/** A tuft of grass: a few blades fanning out, dark at the root and sunlit at the tip. */
export function grassTuftGeometry() {
  const pos: number[] = []
  const col: number[] = []
  const root = new THREE.Color('#4a8c34')
  const tip = new THREE.Color('#c4ea7a')
  const blades = 5
  for (let k = 0; k < blades; k++) {
    const a = (k / blades) * Math.PI * 2 + k * 0.7
    const r = 0.035 + 0.025 * ((k * 37) % 5) / 5
    const h = 0.16 + 0.07 * ((k * 53) % 7) / 7
    const cx = Math.cos(a) * r
    const cz = Math.sin(a) * r
    // Base across the blade, the tip leaning outwards.
    const w = 0.04
    const px = -Math.sin(a) * w
    const pz = Math.cos(a) * w
    pos.push(cx - px, 0, cz - pz, cx + px, 0, cz + pz, cx * 2.6, h, cz * 2.6)
    col.push(root.r, root.g, root.b, root.r, root.g, root.b, tip.r, tip.g, tip.b)
  }
  const g = new THREE.BufferGeometry()
  g.setAttribute('position', new THREE.Float32BufferAttribute(pos, 3))
  g.setAttribute('color', new THREE.Float32BufferAttribute(col, 3))
  g.computeVertexNormals()
  return g
}

export function stumpGeometry() {
  return model([
    part(new THREE.CylinderGeometry(0.08, 0.1, 0.12, 7), '#7a5534', [0, 0.06, 0]),
    part(new THREE.CylinderGeometry(0.075, 0.075, 0.01, 7), '#c9a26b', [0, 0.125, 0]),
  ])
}

export function wallGeometry() {
  return model([part(new THREE.BoxGeometry(1, 0.6, 1), '#6b625a', [0, 0.3, 0])])
}

// --- Animals (in species order: rusa, babi hutan, ayam hutan, kerbau, harimau) --
// Each faces +x.

/** Legs at the given spots, hinged at the body; diagonal pairs swing together (a trot), a bird's two alternate. */
function legs(color: string, len: number, w: number, xs: number[], zs: number[]) {
  const out: Part[] = []
  for (const x of xs) {
    for (const z of zs) {
      const pair = xs.length === 1 ? Math.sign(z) : x > 0 === z > 0 ? 1 : -1
      out.push({ ...part(new THREE.BoxGeometry(w, len, w), color, [x, len / 2, z]), leg: pair, hip: [x, len] })
    }
  }
  return out
}

export function animalGeometries() {
  const deer = model([
    ...legs('#6e4a2c', 0.24, 0.04, [-0.14, 0.14], [-0.06, 0.06]),
    part(new THREE.BoxGeometry(0.42, 0.18, 0.16), '#a8703f', [0, 0.32, 0]),
    part(new THREE.BoxGeometry(0.1, 0.22, 0.09), '#a8703f', [0.2, 0.46, 0], [0, 0, -0.5]),
    part(new THREE.BoxGeometry(0.16, 0.09, 0.09), '#9a6436', [0.3, 0.56, 0]),
    part(new THREE.BoxGeometry(0.06, 0.06, 0.08), '#f2e6d0', [-0.22, 0.36, 0]),
    part(new THREE.ConeGeometry(0.02, 0.16, 4), '#e8dcc0', [0.24, 0.66, 0.04], [0.3, 0, -0.3]),
    part(new THREE.ConeGeometry(0.02, 0.16, 4), '#e8dcc0', [0.24, 0.66, -0.04], [-0.3, 0, -0.3]),
  ])
  const boar = model([
    ...legs('#3b302a', 0.13, 0.05, [-0.12, 0.12], [-0.06, 0.06]),
    part(new THREE.IcosahedronGeometry(0.18, 0), '#4e4038', [0, 0.24, 0], [0, 0, 0], [1.5, 0.95, 0.9]),
    part(new THREE.BoxGeometry(0.14, 0.12, 0.12), '#4a3c34', [0.26, 0.24, 0]),
    part(new THREE.BoxGeometry(0.05, 0.06, 0.07), '#b9837a', [0.34, 0.22, 0]),
    part(new THREE.ConeGeometry(0.012, 0.05, 3), '#f4ecd8', [0.33, 0.2, 0.05], [0, 0, 1.2]),
  ])
  const fowl = model([
    ...legs('#d9a441', 0.08, 0.015, [0], [-0.025, 0.025]),
    part(new THREE.IcosahedronGeometry(0.07, 0), '#b5442a', [0, 0.13, 0], [0, 0, 0], [1.3, 1, 0.9]),
    part(new THREE.IcosahedronGeometry(0.04, 0), '#d76b2e', [0.08, 0.2, 0]),
    part(new THREE.BoxGeometry(0.02, 0.04, 0.012), '#e02020', [0.08, 0.25, 0]),
    part(new THREE.ConeGeometry(0.06, 0.12, 4), '#1f4f3c', [-0.09, 0.19, 0], [0, 0, 0.9]),
  ])
  const buffalo = model([
    ...legs('#2c2a28', 0.28, 0.07, [-0.2, 0.2], [-0.09, 0.09]),
    part(new THREE.BoxGeometry(0.62, 0.3, 0.3), '#3f3c39', [0, 0.42, 0]),
    part(new THREE.BoxGeometry(0.18, 0.18, 0.18), '#373431', [0.38, 0.42, 0]),
    part(new THREE.TorusGeometry(0.11, 0.022, 4, 8, Math.PI), '#d9cfba', [0.38, 0.52, 0], [Math.PI / 2, 0, 0]),
  ])
  const tiger = model([
    ...legs('#c96a1c', 0.2, 0.05, [-0.17, 0.17], [-0.06, 0.06]),
    part(new THREE.BoxGeometry(0.5, 0.17, 0.17), '#e07a22', [0, 0.28, 0]),
    part(new THREE.BoxGeometry(0.03, 0.18, 0.18), '#1d1a18', [-0.08, 0.28, 0]),
    part(new THREE.BoxGeometry(0.03, 0.18, 0.18), '#1d1a18', [0.06, 0.28, 0]),
    part(new THREE.BoxGeometry(0.15, 0.14, 0.15), '#e07a22', [0.3, 0.34, 0]),
    part(new THREE.BoxGeometry(0.06, 0.05, 0.1), '#f6efe0', [0.38, 0.31, 0]),
    part(new THREE.CylinderGeometry(0.018, 0.012, 0.3, 4), '#e07a22', [-0.32, 0.36, 0], [0, 0, -1]),
  ])
  return [deer, boar, fowl, buffalo, tiger]
}

// --- Crops -------------------------------------------------------------------------

/** Rice: a clump of blades; the colour (green or ripe gold) comes from the instance. */
export function riceGeometry() {
  const parts: Part[] = []
  for (let k = 0; k < 9; k++) {
    const a = (k / 9) * Math.PI * 2
    const r = 0.12 + 0.16 * ((k * 7) % 3) / 2
    parts.push(part(new THREE.ConeGeometry(0.025, 0.42, 3), '#ffffff', [Math.cos(a) * r, 0.21, Math.sin(a) * r], [Math.sin(a) * 0.25, 0, Math.cos(a) * 0.25]))
  }
  return model(parts)
}

/** Taro and yams: broad leaves low on the ground. */
export function leafyGeometry() {
  const parts: Part[] = []
  for (let k = 0; k < 5; k++) {
    const a = (k / 5) * Math.PI * 2
    parts.push(part(new THREE.CylinderGeometry(0.01, 0.01, 0.22, 3), '#ffffff', [Math.cos(a) * 0.08, 0.11, Math.sin(a) * 0.08]))
    parts.push(part(new THREE.ConeGeometry(0.13, 0.03, 5), '#ffffff', [Math.cos(a) * 0.16, 0.23, Math.sin(a) * 0.16], [0.3 * Math.sin(a), 0, -0.3 * Math.cos(a)]))
  }
  return model(parts)
}

/** Banana, coconut and sago: a trunk and a crown of fronds. */
export function palmGeometry(trunk = 1) {
  const parts: Part[] = [part(new THREE.CylinderGeometry(0.04, 0.07, 0.9 * trunk, 6), '#8a6a43', [0, 0.45 * trunk, 0])]
  for (let k = 0; k < 7; k++) {
    const a = (k / 7) * Math.PI * 2
    parts.push(part(new THREE.BoxGeometry(0.5, 0.02, 0.12), '#3f9a3a', [Math.cos(a) * 0.22, 0.9 * trunk + 0.02, Math.sin(a) * 0.22], [0, -a, -0.45]))
  }
  parts.push(part(new THREE.IcosahedronGeometry(0.07, 0), '#c9a23a', [0.05, 0.85 * trunk, 0.04]))
  return model(parts)
}

// --- Buildings -----------------------------------------------------------------------
// Each kind is a body (fixed colours) and a roof tinted with the family colour.

export type BuildingModel = { body: THREE.BufferGeometry; roof?: THREE.BufferGeometry }

function gableRoof(w: number, d: number, h: number, y: number) {
  const shape = new THREE.Shape()
  shape.moveTo(-w / 2, 0)
  shape.lineTo(w / 2, 0)
  shape.lineTo(0, h)
  shape.closePath()
  const g = new THREE.ExtrudeGeometry(shape, { depth: d, bevelEnabled: false })
  g.translate(0, y, -d / 2)
  return g
}

export function buildingModel(kind: string, level: number): BuildingModel {
  switch (kind) {
    case 'gubuk':
      return {
        body: model([part(new THREE.CylinderGeometry(0.3, 0.32, 0.36, 8), '#b88a53', [0, 0.18, 0]), part(new THREE.BoxGeometry(0.12, 0.2, 0.04), '#4a3320', [0, 0.1, 0.31])]),
        roof: model([part(new THREE.ConeGeometry(0.46, 0.5, 8), '#ffffff', [0, 0.6, 0])]),
      }
    case 'rumah_kayu':
      return {
        body: model([part(new THREE.BoxGeometry(0.72, 0.42, 0.56), '#8a5a33', [0, 0.21, 0]), part(new THREE.BoxGeometry(0.14, 0.26, 0.03), '#3f2a18', [0, 0.13, 0.29]), part(new THREE.BoxGeometry(0.13, 0.11, 0.03), '#e8d08f', [0.22, 0.26, 0.29])]),
        roof: model([part(gableRoof(0.86, 0.66, 0.36, 0.42), '#ffffff', [0, 0, 0])]),
      }
    case 'rumah_bata':
      return {
        body: model([part(new THREE.BoxGeometry(0.84, 0.5, 0.64), '#b5583a', [0, 0.25, 0]), part(new THREE.BoxGeometry(0.15, 0.3, 0.03), '#3f2a18', [0, 0.15, 0.33]), part(new THREE.BoxGeometry(0.14, 0.13, 0.03), '#a9d6f5', [-0.25, 0.3, 0.33]), part(new THREE.BoxGeometry(0.14, 0.13, 0.03), '#a9d6f5', [0.25, 0.3, 0.33])]),
        roof: model([part(gableRoof(0.98, 0.74, 0.42, 0.5), '#ffffff', [0, 0, 0])]),
      }
    case 'lumbung': {
      const stilts: Part[] = []
      for (const x of [-0.2, 0.2]) for (const z of [-0.16, 0.16]) stilts.push(part(new THREE.CylinderGeometry(0.03, 0.03, 0.3, 5), '#6b4a2b', [x, 0.15, z]))
      return {
        body: model([...stilts, part(new THREE.BoxGeometry(0.5, 0.3, 0.42), '#a87b4a', [0, 0.45, 0])]),
        roof: model([part(gableRoof(0.66, 0.54, 0.34, 0.6), '#ffffff', [0, 0, 0])]),
      }
    }
    case 'kandang': {
      const posts: Part[] = []
      for (let k = 0; k < 10; k++) {
        const a = (k / 10) * Math.PI * 2
        posts.push(part(new THREE.BoxGeometry(0.04, 0.22, 0.04), '#7a5534', [Math.cos(a) * 0.4, 0.11, Math.sin(a) * 0.4]))
      }
      posts.push(part(new THREE.TorusGeometry(0.4, 0.012, 3, 20), '#8a6a43', [0, 0.16, 0], [Math.PI / 2, 0, 0]))
      return { body: model(posts) }
    }
    case 'sumur':
      return {
        body: model([part(new THREE.CylinderGeometry(0.2, 0.22, 0.18, 10, 1, true), '#8d8a83', [0, 0.09, 0]), part(new THREE.CircleGeometry(0.19, 10), '#2f74c0', [0, 0.12, 0], [-Math.PI / 2, 0, 0]), part(new THREE.BoxGeometry(0.03, 0.42, 0.03), '#6b4626', [-0.18, 0.21, 0]), part(new THREE.BoxGeometry(0.03, 0.42, 0.03), '#6b4626', [0.18, 0.21, 0])]),
        roof: model([part(gableRoof(0.5, 0.32, 0.16, 0.42), '#ffffff', [0, 0, 0])]),
      }
    case 'jamban':
      // A booth of woven bamboo on a timber floor over the pit, with a lean-to thatch.
      return {
        body: model([
          part(new THREE.BoxGeometry(0.4, 0.05, 0.4), '#5a3a1e', [0, 0.025, 0]),
          part(new THREE.BoxGeometry(0.3, 0.36, 0.3), '#b89a62', [0, 0.23, 0]),
          part(new THREE.BoxGeometry(0.11, 0.24, 0.02), '#4a3320', [0, 0.17, 0.16]),
        ]),
        roof: model([part(new THREE.BoxGeometry(0.4, 0.035, 0.4), '#ffffff', [0, 0.43, 0], [0.18, 0, 0])]),
      }
    case 'tungku':
      return {
        body: model([part(new THREE.SphereGeometry(0.3, 8, 5, 0, Math.PI * 2, 0, Math.PI / 2), '#8f8578', [0, 0, 0]), part(new THREE.BoxGeometry(0.11, 0.4, 0.11), '#7a5238', [0.12, 0.3, 0]), part(new THREE.BoxGeometry(0.12, 0.08, 0.03), '#ff8a3d', [0, 0.06, 0.29])]),
      }
    case 'jerat':
      return {
        body: model([part(new THREE.CylinderGeometry(0.012, 0.018, 0.5, 4), '#6b4626', [-0.12, 0.2, 0], [0, 0, -0.6]), part(new THREE.TorusGeometry(0.07, 0.008, 3, 10), '#d8cfa8', [0.06, 0.03, 0], [Math.PI / 2, 0, 0]), part(new THREE.CylinderGeometry(0.004, 0.004, 0.3, 3), '#d8cfa8', [0.04, 0.17, 0])]),
      }
    case 'perpustakaan':
      return {
        body: model([part(new THREE.BoxGeometry(0.8, 0.5, 0.6), '#c8b48a', [0, 0.25, 0]), ...[-0.3, -0.1, 0.1, 0.3].map((x) => part(new THREE.CylinderGeometry(0.035, 0.035, 0.5, 6), '#efe6d2', [x, 0.25, 0.32]))]),
        roof: model([part(gableRoof(0.96, 0.72, 0.28, 0.5), '#ffffff', [0, 0, 0])]),
      }
    default: {
      // Stations and laboratories: a solid block with a tall roof in the family colour.
      const tall = 0.45 + 0.08 * Math.max(0, level)
      return {
        body: model([part(new THREE.BoxGeometry(0.8, tall, 0.62), '#9a9890', [0, tall / 2, 0]), part(new THREE.BoxGeometry(0.16, 0.28, 0.03), '#3f3a34', [0, 0.14, 0.32]), part(new THREE.BoxGeometry(0.12, 0.38, 0.12), '#6f6a62', [0.26, tall + 0.1, -0.12])]),
        roof: model([part(new THREE.BoxGeometry(0.86, 0.06, 0.68), '#ffffff', [0, tall + 0.03, 0])]),
      }
    }
  }
}
