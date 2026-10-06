// Life on the water in 3D: rings spreading from the feet of people wading or
// swimming (a wake behind them as they move), and a small bamboo raft under
// anyone rafting, riding the surface. After RAGE's ripple manager
// (game/vfx/systems/VfxRipple.h): a fixed pool of ripple points, each with its
// own life, size and peak opacity, registered by whoever disturbs the water.

import * as THREE from 'three'
import { FLAG } from '../sim/protocol'
import type { LiveCreature } from '../sim/live'
import type { Terrain } from './terrain'

/** How much lower fresh water is drawn when its river has fallen (the water shader's drop). */
const FALL = 0.34

/**
 * The drawn water surface at a point (world y), or null where there is no water to be in. Fresh
 * water sits lower as its river runs low (`level`: the stream's water level per tile, 0–255, as
 * in World3D's water texture; null when unknown); the sea stays at its level.
 */
export function waterSurface(t: Terrain, level: Uint8Array | null, x: number, z: number) {
  const w = t.waterAt(x, z)
  if (!Number.isFinite(w)) return null
  const GW = 2 * t.width + 1
  const gx = Math.min(GW - 1, Math.max(0, Math.round(x * 2)))
  const gz = Math.min(2 * t.height, Math.max(0, Math.round(z * 2)))
  let y = w
  if (level && t.fresh[gz * GW + gx]) {
    // Bilinear between tile centres, as the texture is sampled.
    const W = t.width
    const H = t.height
    const fx = Math.min(W - 1, Math.max(0, x - 0.5))
    const fz = Math.min(H - 1, Math.max(0, z - 0.5))
    const ix = Math.min(W - 2, Math.floor(fx))
    const iz = Math.min(H - 2, Math.floor(fz))
    const u = W > 1 ? fx - Math.max(0, ix) : 0
    const v = H > 1 ? fz - Math.max(0, iz) : 0
    const at = (a: number, b: number) => level[(Math.max(0, Math.min(H - 1, b)) * W + Math.max(0, Math.min(W - 1, a))) * 4] / 255
    const top = at(ix, iz) * (1 - u) + at(ix + 1, iz) * u
    const bottom = at(ix, iz + 1) * (1 - u) + at(ix + 1, iz + 1) * u
    const lv = top * (1 - v) + bottom * v
    if (lv < 0.03) return null // a dry bed
    y -= (1 - lv) * FALL
  }
  if (y < t.heightAt(x, z) + 0.01) return null
  return y
}

/**
 * The waves the water shader draws at a point (World3D's waterMaterial: a swell that grows with
 * the wind and with depth, much calmer on fresh water), so whatever floats rides the very
 * surface that is drawn. `depth`: water surface minus ground; `time`, `wind`: the shared uTime and uWind.
 */
export function waveAt(x: number, z: number, depth: number, fresh: boolean, time: number, wind: number) {
  const t = Math.min(1, Math.max(0, (depth - 0.05) / 0.85))
  const swell = (0.02 + 0.035 * wind) * (fresh ? 0.15 : 1) * t * t * (3 - 2 * t)
  return (Math.sin(x * 1.3 + time * 1.4) + Math.sin(z * 1.7 - time * 1.1) * 0.7) * swell
}

/** Whether the water at a point is fresh (a river or lake) rather than the sea. */
export function freshAt(t: Terrain, x: number, z: number) {
  const GW = 2 * t.width + 1
  const gx = Math.min(GW - 1, Math.max(0, Math.round(x * 2)))
  const gz = Math.min(2 * t.height, Math.max(0, Math.round(z * 2)))
  return t.fresh[gz * GW + gx] === 1
}

/**
 * How high a raft's deck stands above the water per unit of Placement.scale: someone rafting
 * kneels at `placement.water(x, z) + RAFT_DECK * placement.scale`.
 */
export const RAFT_DECK = 0.05 / 1.6

const RIPPLES = 320
/** People further than this from where the camera looks make no ripples. */
const RIPPLE_REACH = 42

export type WaterPlacement = {
  /** The drawn surface (with the swell), or null. */
  water: (x: number, z: number) => number | null
  /** Placement.scale: how big a size-1 adult is drawn at this distance. */
  scale: number
  time: number
  still: boolean
  inView: (x: number, y: number, z: number, size: number) => boolean
  /** Where the camera looks. */
  camX: number
  camZ: number
}

export class WaterLife {
  readonly group = new THREE.Group()
  private readonly ripples: THREE.Mesh<THREE.InstancedBufferGeometry, THREE.ShaderMaterial>
  private rafts: THREE.InstancedMesh
  private readonly next = new Map<number, number>()
  private slot = 0
  private emitters = 0
  private readonly m = new THREE.Matrix4()
  private readonly q = new THREE.Quaternion()
  private readonly e = new THREE.Euler()
  private readonly v = new THREE.Vector3()
  private readonly s = new THREE.Vector3()

  constructor() {
    const ring = new THREE.RingGeometry(0.84, 1, 40, 1).rotateX(-Math.PI / 2)
    const g = new THREE.InstancedBufferGeometry()
    g.index = ring.index
    g.setAttribute('position', ring.getAttribute('position'))
    g.setAttribute('aSpot', new THREE.InstancedBufferAttribute(new Float32Array(RIPPLES * 4), 4).setUsage(THREE.DynamicDrawUsage))
    // Birth time far in the past: every slot starts spent.
    g.setAttribute(
      'aRipple',
      new THREE.InstancedBufferAttribute(new Float32Array(RIPPLES * 4).fill(-1e6), 4).setUsage(THREE.DynamicDrawUsage),
    )
    g.instanceCount = RIPPLES
    this.ripples = new THREE.Mesh(g, rippleMaterial())
    this.ripples.frustumCulled = false
    this.ripples.renderOrder = 1
    this.rafts = this.makeRafts(8)
    this.group.add(this.ripples, this.rafts)
  }

  /** Transparent effects the ambient occlusion pass must not see as solid. */
  get effects(): THREE.Object3D[] {
    return [this.ripples]
  }

  private makeRafts(cap: number) {
    const mesh = new THREE.InstancedMesh(
      raftGeometry(),
      new THREE.MeshStandardMaterial({ vertexColors: true, flatShading: true, roughness: 0.85 }),
      cap,
    )
    mesh.count = 0
    mesh.castShadow = true
    mesh.receiveShadow = true
    mesh.frustumCulled = false
    return mesh
  }

  update(people: Iterable<LiveCreature>, at: WaterPlacement) {
    const mat = this.ripples.material
    mat.uniforms.uTime.value = at.time
    mat.uniforms.uScale.value = at.scale
    const busy = Math.max(1, (this.emitters * 2.2 * 1.6) / RIPPLES)
    let emitters = 0
    let rafts = 0
    let spawned = 0
    // The raft and the rings are modelled for people drawn 1.6 tiles per unit of size.
    const s = at.scale / 1.6
    for (const cr of people) {
      const f = cr.flags
      if (!(f & (FLAG.wading | FLAG.swimming | FLAG.rafting))) continue
      const raft = (f & FLAG.rafting) !== 0
      const near = Math.abs(cr.rx - at.camX) < RIPPLE_REACH && Math.abs(cr.ry - at.camZ) < RIPPLE_REACH
      if (!near && !raft) continue
      const y = at.water(cr.rx, cr.ry)
      if (y === null) continue
      if (!at.inView(cr.rx, y, cr.ry, s)) continue
      if (raft) {
        if (rafts >= this.rafts.instanceMatrix.count) this.growRafts(rafts + 1)
        this.e.set(
          at.still ? 0 : 0.04 * Math.sin(at.time * 1.7 + cr.id),
          -cr.rh,
          at.still ? 0 : 0.05 * Math.sin(at.time * 1.3 + cr.id * 1.7),
          'YXZ',
        )
        this.q.setFromEuler(this.e)
        this.m.compose(this.v.set(cr.rx, y, cr.ry), this.q, this.s.setScalar(s))
        this.rafts.setMatrixAt(rafts++, this.m)
      }
      if (!near || at.still) continue
      emitters++
      const swimming = (f & FLAG.swimming) !== 0
      const interval = (cr.moving ? (swimming ? 0.3 : raft ? 0.55 : 0.42) : swimming ? 0.9 : raft ? 1.6 : 1.5) * busy
      const due = this.next.get(cr.id) ?? 0
      if (at.time < due || spawned >= 24) continue
      this.next.set(cr.id, at.time + interval * (0.85 + 0.3 * Math.random()))
      const radius = (raft ? 0.75 : swimming ? 0.5 : 0.36) * s
      // A ring where they are; moving, a stretched wake ring a little behind.
      const back = cr.moving ? 0.14 * s : 0
      this.spawn(
        cr.rx - Math.cos(cr.rh) * back,
        y + 0.012,
        cr.ry - Math.sin(cr.rh) * back,
        cr.rh,
        at.time,
        cr.moving ? 1.3 : 1.8,
        radius,
        cr.moving ? 1.6 : 1,
      )
      spawned++
    }
    this.emitters = emitters
    this.rafts.count = rafts
    this.rafts.instanceMatrix.needsUpdate = true
    if (spawned) {
      const g = this.ripples.geometry
      g.getAttribute('aSpot').needsUpdate = true
      g.getAttribute('aRipple').needsUpdate = true
    }
    // Forget people who left the water long ago.
    if (this.next.size > 2048) for (const [id, t] of this.next) if (t < at.time - 10) this.next.delete(id)
  }

  private spawn(x: number, y: number, z: number, angle: number, time: number, life: number, radius: number, stretch: number) {
    const g = this.ripples.geometry
    const spot = g.getAttribute('aSpot') as THREE.InstancedBufferAttribute
    const rip = g.getAttribute('aRipple') as THREE.InstancedBufferAttribute
    const i = this.slot
    this.slot = (this.slot + 1) % RIPPLES
    spot.setXYZW(i, x, y, z, angle)
    rip.setXYZW(i, time, life, radius, stretch)
  }

  private growRafts(n: number) {
    const old = this.rafts
    const next = this.makeRafts(Math.max(n, old.instanceMatrix.count * 2))
    for (let i = 0; i < old.count; i++) {
      old.getMatrixAt(i, this.m)
      next.setMatrixAt(i, this.m)
    }
    next.count = old.count
    this.group.remove(old)
    old.dispose()
    this.group.add(next)
    this.rafts = next
  }

  dispose() {
    this.ripples.geometry.dispose()
    this.ripples.material.dispose()
    this.rafts.geometry.dispose()
    ;(this.rafts.material as THREE.Material).dispose()
    this.rafts.dispose()
  }
}

/** Rings that grow fast then slow, fading as they spread; a wake is stretched along the heading. */
function rippleMaterial() {
  return new THREE.ShaderMaterial({
    transparent: true,
    depthWrite: false,
    // On the water's surface: never lost in it.
    polygonOffset: true,
    polygonOffsetFactor: -2,
    polygonOffsetUnits: -4,
    uniforms: { uTime: { value: 0 }, uScale: { value: 1 } },
    vertexShader: `
      uniform float uTime;
      attribute vec4 aSpot;
      attribute vec4 aRipple;
      varying float vAlpha;
      varying float vR;
      void main() {
        float age = (uTime - aRipple.x) / aRipple.y;
        if (age < 0.0 || age >= 1.0) {
          gl_Position = vec4(2.0, 2.0, 2.0, 1.0);
          return;
        }
        float grow = 1.0 - pow(1.0 - age, 2.4);
        float r = mix(0.18, 1.0, grow) * aRipple.z;
        vec3 p = position;
        vR = length(p.xz);
        p.x *= r * aRipple.w;
        p.z *= r;
        float c = cos(aSpot.w);
        float s = sin(aSpot.w);
        vec3 w = aSpot.xyz + vec3(p.x * c - p.z * s, 0.0, p.x * s + p.z * c);
        gl_Position = projectionMatrix * viewMatrix * vec4(w, 1.0);
        vAlpha = (1.0 - age) * (1.0 - age) * smoothstep(0.0, 0.08, age);
      }`,
    fragmentShader: `
      varying float vAlpha;
      varying float vR;
      void main() {
        float band = smoothstep(0.84, 0.91, vR) * (1.0 - smoothstep(0.93, 1.0, vR));
        float a = band * vAlpha * 0.6;
        if (a <= 0.004) discard;
        gl_FragColor = vec4(0.92, 0.97, 1.0, a);
      }`,
  })
}

/** A raft of six bamboo poles lashed across two ties, its deck 0.05 above the water (at scale 1). Faces +x. */
function raftGeometry() {
  const parts: { g: THREE.BufferGeometry; c: string }[] = []
  for (let k = 0; k < 6; k++) {
    const pole = new THREE.CylinderGeometry(0.034, 0.034, 0.82 - (k === 0 || k === 5 ? 0.08 : 0), 6)
    pole.rotateZ(Math.PI / 2)
    pole.translate(0, 0.016, -0.2 + k * 0.08)
    parts.push({ g: pole, c: k % 2 ? '#c9a55a' : '#b8924a' })
  }
  for (const x of [-0.3, 0.3]) {
    const tie = new THREE.BoxGeometry(0.05, 0.03, 0.5)
    tie.translate(x, 0.045, 0)
    parts.push({ g: tie, c: '#7a5a32' })
  }
  let count = 0
  const geos = parts.map(({ g, c }) => {
    const n = g.index ? g.toNonIndexed() : g
    count += n.getAttribute('position').count
    return { g: n, c: new THREE.Color(c) }
  })
  const P = new Float32Array(count * 3)
  const C = new Float32Array(count * 3)
  let o = 0
  for (const { g, c } of geos) {
    const pos = g.getAttribute('position')
    P.set(pos.array as Float32Array, o * 3)
    for (let i = 0; i < pos.count; i++) C.set([c.r, c.g, c.b], (o + i) * 3)
    o += pos.count
    g.dispose()
  }
  const out = new THREE.BufferGeometry()
  out.setAttribute('position', new THREE.BufferAttribute(P, 3))
  out.setAttribute('color', new THREE.BufferAttribute(C, 3))
  out.computeVertexNormals()
  return out
}
