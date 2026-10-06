// Fire on the island in 3D: flickering flames over each burning tile, embers
// and smoke columns that rise and drift with the simulation's wind, the
// light of the few fires that matter most, ash where a house burned down, and
// a texture the land and plants read to char where fire has passed (and glow
// where it burns). The budget and the particles themselves are in fire3d.ts.

import * as THREE from 'three'
import type { BurntMessage, FireFrame, StructureFrame } from '../sim/protocol'
import { EMBERS, FirePlanner, Particles, SMOKE, due, flicker, hash3, layoutFlames, type FireBudget } from './fire3d'
import type { Terrain } from './terrain'

/** Budgets per quality level (World3D's 0 lowest … 2 everything). */
const BUDGETS: FireBudget[] = [
  { flames: 160, smoke: 260, embers: 40, lights: 0 },
  { flames: 320, smoke: 520, embers: 100, lights: 2 },
  { flames: 480, smoke: 760, embers: 160, lights: 3 },
]
const MAX = BUDGETS[2]
const SMOKE_LIFE = 6.5
const EMBER_LIFE = 1.7
const ASH_CAP = 64

/** The wind as drawn: unit direction it blows towards (x, z) and strength (0 – ~1.4). */
export type FireWind = { x: number; z: number; strength: number }

export class FireLayer {
  readonly group = new THREE.Group()
  /** One texel per tile: R how scorched, G how hard it burns right now. The land and plants read it. */
  texture: THREE.DataTexture
  private readonly flames: THREE.Mesh<THREE.InstancedBufferGeometry, THREE.ShaderMaterial>
  private readonly smoke: THREE.Mesh<THREE.InstancedBufferGeometry, THREE.ShaderMaterial>
  private readonly embers: THREE.Mesh<THREE.InstancedBufferGeometry, THREE.ShaderMaterial>
  private readonly ash: THREE.InstancedMesh
  private readonly lights: THREE.PointLight[] = []
  private readonly planner = new FirePlanner()
  private readonly layout = new Float32Array(MAX.flames * 4)
  private readonly smokeSim = new Particles(MAX.smoke)
  private readonly emberSim = new Particles(MAX.embers)
  private fires: readonly FireFrame[] = []
  private fireTiles = new Int32Array(0)
  private fireCount = 0
  private planned: { fires: readonly FireFrame[] | null; x: number; z: number; level: number; scale: number } = {
    fires: null,
    x: 0,
    z: 0,
    level: -1,
    scale: 0,
  }
  private burnt: BurntMessage | null = null
  private structures = new Map<number, { x: number; y: number }>()
  private ashSpots = new Map<number, { x: number; z: number; y: number; seed: number }>()
  private ashDirty = false
  private stillAt = -Infinity
  private stillKey = ''

  constructor(
    private width: number,
    private height: number,
  ) {
    this.texture = burntTexture(width, height)
    const quad = new THREE.PlaneGeometry(1, 1)
    this.flames = billboards(quad, MAX.flames, ['aFlame', 4], ['aSeed', 1], flameMaterial())
    this.smoke = billboards(quad, MAX.smoke, ['aPuff', 4], ['aLook', 3], smokeMaterial())
    this.embers = billboards(quad, MAX.embers, ['aPuff', 4], ['aLook', 3], emberMaterial())
    // Smoke drawn after the flames it rises from; embers over both.
    this.flames.renderOrder = 2
    this.smoke.renderOrder = 3
    this.embers.renderOrder = 4
    this.ash = new THREE.InstancedMesh(
      ashGeometry(),
      new THREE.MeshStandardMaterial({ vertexColors: true, flatShading: true, roughness: 1 }),
      ASH_CAP,
    )
    this.ash.count = 0
    this.ash.receiveShadow = true
    this.ash.frustumCulled = false
    this.group.add(this.flames, this.smoke, this.embers, this.ash)
  }

  /** Transparent effects the ambient occlusion pass must not see as solid. */
  get effects(): THREE.Object3D[] {
    return [this.flames, this.smoke, this.embers]
  }

  /** The map changed size: a new texture (the caller hands it to its materials). */
  resize(width: number, height: number) {
    if (width === this.width && height === this.height) return false
    this.width = width
    this.height = height
    this.texture.dispose()
    this.texture = burntTexture(width, height)
    this.fireCount = 0
    this.setBurnt(this.burnt)
    return true
  }

  /** Scorched land from the stream (`burnt`), fading as plants return. */
  setBurnt(msg: BurntMessage | null) {
    this.burnt = msg
    const data = this.texture.image.data as Uint8Array
    const W = this.width
    for (let i = 0; i < data.length; i += 4) data[i] = 0
    if (msg)
      for (const t of msg.tiles) {
        if (t.x < 0 || t.y < 0 || t.x >= W || t.y >= this.height) continue
        data[(t.y * W + t.x) * 4] = Math.round(255 * Math.min(1, Math.max(0, t.level)))
      }
    this.texture.needsUpdate = true
    // Ash stays only while the ground under it is still scorched.
    for (const tile of this.ashSpots.keys()) if (data[tile * 4] < 12) this.ashSpots.delete(tile)
    this.ashDirty = true
  }

  /** The tiles burning in the latest frame. */
  setFires(fires: readonly FireFrame[]) {
    if (fires === this.fires) return
    this.fires = fires
    const data = this.texture.image.data as Uint8Array
    const W = this.width
    for (let k = 0; k < this.fireCount; k++) data[this.fireTiles[k] * 4 + 1] = 0
    if (this.fireTiles.length < fires.length) this.fireTiles = new Int32Array(Math.max(fires.length, this.fireTiles.length * 2, 32))
    let n = 0
    for (const f of fires) {
      if (f.x < 0 || f.y < 0 || f.x >= W || f.y >= this.height) continue
      const tile = f.y * W + f.x
      data[tile * 4 + 1] = Math.round(255 * Math.min(1, Math.max(0, f.intensity)))
      this.fireTiles[n++] = tile
    }
    if (n || this.fireCount) this.texture.needsUpdate = true
    this.fireCount = n
  }

  /**
   * Buildings as streamed: one that vanished where fire burns or has just burned leaves a heap
   * of ash and charred beams, until the ground recovers.
   */
  setStructures(list: readonly StructureFrame[], terrain: Terrain) {
    const next = new Map<number, { x: number; y: number }>()
    for (const st of list) if (st.kind !== 'saluran_irigasi' && st.kind !== 'ladang') next.set(st.id, { x: st.x, y: st.y })
    const data = this.texture.image.data as Uint8Array
    for (const [id, at] of this.structures) {
      if (next.has(id) || at.x < 0 || at.y < 0 || at.x >= this.width || at.y >= this.height) continue
      const tile = at.y * this.width + at.x
      const fire = data[tile * 4 + 1] > 0 || this.burningNear(at.x, at.y)
      if (!fire && data[tile * 4] < 25) continue
      if (this.ashSpots.size >= ASH_CAP) break
      const x = at.x + 0.5
      const z = at.y + 0.5
      this.ashSpots.set(tile, { x, z, y: terrain.heightAt(x, z), seed: hash3(at.x, at.y, 5) })
      this.ashDirty = true
    }
    this.structures = next
  }

  private burningNear(x: number, y: number) {
    for (const f of this.fires) if (Math.abs(f.x - x) <= 1 && Math.abs(f.y - y) <= 1) return true
    return false
  }

  /** The ground moved (a new relief): ash settles on it again. */
  reground(terrain: Terrain) {
    for (const a of this.ashSpots.values()) a.y = terrain.heightAt(a.x, a.z)
    this.ashDirty = true
  }

  /**
   * Per frame. `cam` is where the camera looks (tiles); `scale` how much larger than life things
   * are drawn at this distance; `level` World3D's quality level; `light` 0 night … 1 day.
   */
  update(
    dt: number,
    time: number,
    cam: THREE.Vector3,
    terrain: Terrain,
    wind: FireWind,
    level: number,
    scale: number,
    light: number,
    still: boolean,
  ) {
    const budget = BUDGETS[Math.max(0, Math.min(2, level))]
    const fires = this.fires
    const p = this.planned
    if (p.fires !== fires || p.level !== level || Math.hypot(cam.x - p.x, cam.z - p.z) > 4 || Math.abs(scale - p.scale) > 0.1) {
      this.planner.plan(fires, cam.x, cam.z, budget)
      p.fires = fires
      p.x = cam.x
      p.z = cam.z
      p.level = level
      p.scale = scale
      this.placeFlames(terrain)
    }
    const plan = this.planner
    const flameMat = this.flames.material
    flameMat.uniforms.uTime.value = still ? 0 : time
    flameMat.uniforms.uWind.value.set(wind.x, wind.z, wind.strength)
    flameMat.uniforms.uScale.value = scale

    // Smoke and embers: emitted at each fire's share of the budget, carried by the wind.
    if (still) {
      // Reduced motion: a column already standing, held still (rebuilt only when the fires change).
      const key = `${fires.length}:${plan.total.smoke}:${plan.total.embers}`
      if (key !== this.stillKey && time - this.stillAt > 2) {
        this.stillKey = key
        this.stillAt = time
        this.smokeSim.clear()
        this.emberSim.clear()
        for (let t = 0; t < 6; t += 0.1) this.simulate(0.1, t, terrain, wind, scale)
      }
    } else this.simulate(dt, time, terrain, wind, scale)
    writePuffs(this.smoke, this.smokeSim, 'smoke')
    writePuffs(this.embers, this.emberSim, 'ember')
    this.smoke.material.uniforms.uLight.value = 0.25 + 0.75 * light
    this.embers.material.uniforms.uTime.value = still ? 0 : time

    // Light from the strongest fires near the camera (a few, made on first need and then kept,
    // so the number of lights, and with it every lit material, never changes again).
    if (plan.lights.length && !this.lights.length) {
      for (let k = 0; k < MAX.lights; k++) {
        const l = new THREE.PointLight('#ff8a3d', 0, 9, 1.6)
        l.castShadow = false
        this.lights.push(l)
        this.group.add(l)
      }
    }
    for (let k = 0; k < this.lights.length; k++) {
      const l = this.lights[k]
      const i = k < plan.lights.length ? plan.lights[k] : -1
      if (i < 0 || i >= fires.length) {
        l.intensity = 0
        continue
      }
      const f = fires[i]
      const x = f.x + 0.5
      const z = f.y + 0.5
      l.position.set(x, terrain.heightAt(x, z) + 0.7 * scale, z)
      // Brighter at night, when it matters.
      l.intensity = (6 + 10 * (1 - light)) * f.intensity * (still ? 0.9 : flicker(time, hash3(f.x, f.y, 3))) * scale
      l.distance = 6 + 5 * f.intensity * scale
    }

    if (this.ashDirty) {
      this.ashDirty = false
      const m = new THREE.Matrix4()
      const q = new THREE.Quaternion()
      const s = new THREE.Vector3()
      const v = new THREE.Vector3()
      let n = 0
      for (const a of this.ashSpots.values()) {
        q.setFromAxisAngle(UP, a.seed * Math.PI * 2)
        s.setScalar(0.85 + 0.3 * a.seed)
        m.compose(v.set(a.x, a.y + 0.01, a.z), q, s)
        this.ash.setMatrixAt(n++, m)
      }
      this.ash.count = n
      this.ash.instanceMatrix.needsUpdate = true
    }
  }

  private simulate(dt: number, time: number, terrain: Terrain, wind: FireWind, scale: number) {
    const plan = this.planner
    const fires = this.fires
    for (let i = 0; i < fires.length; i++) {
      const f = fires[i]
      const heat = Math.min(1, Math.max(0, f.intensity))
      const ns = due(time, dt, plan.smoke[i] / SMOKE_LIFE, hash3(f.x, f.y, 7))
      const ne = due(time, dt, plan.embers[i] / EMBER_LIFE, hash3(f.x, f.y, 9))
      if (!ns && !ne) continue
      const cx = f.x + 0.5
      const cz = f.y + 0.5
      const ground = terrain.heightAt(cx, cz)
      for (let k = 0; k < ns; k++) {
        const r = hash3(f.x + k * 7, f.y, Math.floor(time * 10))
        const jx = (r - 0.5) * 0.6
        const jz = (hash3(f.x, f.y + k * 7, Math.floor(time * 10)) - 0.5) * 0.6
        this.smokeSim.emit(
          cx + jx,
          ground + (0.35 + 0.4 * heat) * scale,
          cz + jz,
          (0.9 + 0.9 * heat) * scale,
          SMOKE_LIFE * (0.75 + 0.5 * r),
          (0.16 + 0.22 * heat) * scale,
          heat,
          r,
        )
      }
      for (let k = 0; k < ne; k++) {
        const r = hash3(f.x + k * 13, f.y, Math.floor(time * 30))
        this.emberSim.emit(
          cx + (r - 0.5) * 0.7,
          ground + 0.25 * scale,
          cz + (hash3(f.y, f.x + k, Math.floor(time * 30)) - 0.5) * 0.7,
          (1.4 + 1.2 * r) * scale,
          EMBER_LIFE * (0.6 + 0.8 * r),
          0.045 * scale,
          heat,
          r,
        )
      }
    }
    const strength = Math.min(1.4, wind.strength)
    this.smokeSim.step(dt, wind.x, wind.z, strength, SMOKE, time)
    this.emberSim.step(dt, wind.x, wind.z, strength, EMBERS, time)
  }

  private placeFlames(terrain: Terrain) {
    const n = layoutFlames(this.fires, this.planner, this.layout)
    const g = this.flames.geometry
    const at = g.getAttribute('aFlame') as THREE.InstancedBufferAttribute
    const seed = g.getAttribute('aSeed') as THREE.InstancedBufferAttribute
    const L = this.layout
    for (let k = 0; k < n; k++) {
      const x = L[k * 4]
      const z = L[k * 4 + 1]
      at.setXYZW(k, x, terrain.heightAt(x, z), z, L[k * 4 + 2])
      seed.setX(k, L[k * 4 + 3])
    }
    at.needsUpdate = true
    seed.needsUpdate = true
    g.instanceCount = n
    this.flames.visible = n > 0
  }

  dispose() {
    for (const m of [this.flames, this.smoke, this.embers]) {
      m.geometry.dispose()
      m.material.dispose()
    }
    this.ash.geometry.dispose()
    ;(this.ash.material as THREE.Material).dispose()
    this.ash.dispose()
    this.texture.dispose()
  }
}

const UP = new THREE.Vector3(0, 1, 0)

function burntTexture(w: number, h: number) {
  const tex = new THREE.DataTexture(new Uint8Array(Math.max(1, w * h) * 4), Math.max(1, w), Math.max(1, h), THREE.RGBAFormat)
  tex.magFilter = THREE.LinearFilter
  tex.minFilter = THREE.LinearFilter
  tex.needsUpdate = true
  return tex
}

/** Camera-facing quads, one per instance, with two per-instance attributes. */
function billboards(quad: THREE.PlaneGeometry, cap: number, a: [string, number], b: [string, number], mat: THREE.ShaderMaterial) {
  const g = new THREE.InstancedBufferGeometry()
  g.index = quad.index
  g.setAttribute('position', quad.getAttribute('position'))
  g.setAttribute('uv', quad.getAttribute('uv'))
  g.setAttribute(a[0], new THREE.InstancedBufferAttribute(new Float32Array(cap * a[1]), a[1]).setUsage(THREE.DynamicDrawUsage))
  g.setAttribute(b[0], new THREE.InstancedBufferAttribute(new Float32Array(cap * b[1]), b[1]).setUsage(THREE.DynamicDrawUsage))
  g.instanceCount = 0
  const mesh = new THREE.Mesh(g, mat)
  mesh.frustumCulled = false
  mesh.castShadow = mesh.receiveShadow = false
  return mesh
}

/** Copies live particles into a billboard mesh: position and size, then opacity, shade and heat. */
function writePuffs(mesh: THREE.Mesh<THREE.InstancedBufferGeometry, THREE.ShaderMaterial>, sim: Particles, kind: 'smoke' | 'ember') {
  const g = mesh.geometry
  const puff = g.getAttribute('aPuff') as THREE.InstancedBufferAttribute
  const look = g.getAttribute('aLook') as THREE.InstancedBufferAttribute
  const P = puff.array as Float32Array
  const L = look.array as Float32Array
  const n = sim.count
  for (let i = 0; i < n; i++) {
    const t = sim.age[i] / sim.life[i]
    P[i * 4] = sim.x[i]
    P[i * 4 + 1] = sim.y[i]
    P[i * 4 + 2] = sim.z[i]
    P[i * 4 + 3] = sim.size[i]
    if (kind === 'smoke') {
      // Fades in quickly, thins out as it spreads; the hotter the fire, the darker the smoke.
      L[i * 3] = Math.min(1, t * 6) * Math.pow(1 - t, 1.6) * (0.3 + 0.3 * sim.heat[i])
      L[i * 3 + 1] = sim.heat[i] * (1 - 0.5 * t)
      // Lit from below by the flames while it is low.
      L[i * 3 + 2] = sim.heat[i] * Math.max(0, 1 - t * 4)
    } else {
      L[i * 3] = Math.min(1, t * 8) * (1 - t)
      L[i * 3 + 1] = sim.heat[i]
      L[i * 3 + 2] = sim.seed[i]
    }
  }
  g.instanceCount = n
  mesh.visible = n > 0
  if (n) {
    puff.addUpdateRange(0, n * 4)
    look.addUpdateRange(0, n * 3)
    puff.needsUpdate = true
    look.needsUpdate = true
  }
}

const BILLBOARD = `
  vec3 camRight = vec3(viewMatrix[0][0], viewMatrix[1][0], viewMatrix[2][0]);
  vec3 camUp = vec3(viewMatrix[0][1], viewMatrix[1][1], viewMatrix[2][1]);`

/**
 * Flame tongues, animated on the GPU: each rises from its spot, narrowing and fading, then
 * starts again (its own rhythm from its seed), leaning downwind as it climbs.
 */
function flameMaterial() {
  return new THREE.ShaderMaterial({
    transparent: true,
    depthWrite: false,
    blending: THREE.AdditiveBlending,
    uniforms: { uTime: { value: 0 }, uWind: { value: new THREE.Vector3(1, 0, 0.3) }, uScale: { value: 1 } },
    vertexShader: `
      uniform float uTime;
      uniform vec3 uWind;
      uniform float uScale;
      attribute vec4 aFlame;
      attribute float aSeed;
      varying vec2 vUv;
      varying float vLife;
      varying float vHeat;
      void main() {
        ${BILLBOARD}
        float heat = aFlame.w;
        float life = fract(uTime * (1.1 + aSeed * 0.9) + aSeed * 7.13);
        float tall = (0.3 + 0.7 * heat) * uScale;
        float h = life * tall;
        vec3 c = aFlame.xyz;
        c.x += sin(uTime * 7.0 + aSeed * 40.0) * 0.05 * life * uScale;
        c.z += cos(uTime * 6.3 + aSeed * 31.0) * 0.05 * life * uScale;
        c.xz += uWind.xy * min(uWind.z, 1.3) * life * life * tall * 0.9;
        c.y += h;
        float size = (0.26 + 0.34 * heat) * uScale * (1.0 - 0.6 * life);
        vec3 p = c + camRight * position.x * size + camUp * (position.y + 0.3) * size * 1.7;
        gl_Position = projectionMatrix * viewMatrix * vec4(p, 1.0);
        vUv = uv;
        vLife = life;
        vHeat = heat;
      }`,
    fragmentShader: `
      varying vec2 vUv;
      varying float vLife;
      varying float vHeat;
      void main() {
        vec2 q = vUv - vec2(0.5, 0.32);
        // A teardrop: round at the root, drawn to a point above.
        q.x *= 1.0 + max(0.0, q.y) * 2.4;
        float d = length(q * vec2(2.1, 1.45));
        float body = 1.0 - smoothstep(0.3, 0.72, d);
        if (body <= 0.001) discard;
        float core = 1.0 - smoothstep(0.0, 0.42, d + vLife * 0.35);
        vec3 col = mix(vec3(0.85, 0.12, 0.02), vec3(1.0, 0.5, 0.08), body);
        col = mix(col, vec3(1.0, 0.9, 0.55), core);
        float fade = (1.0 - vLife) * smoothstep(0.0, 0.12, vLife);
        gl_FragColor = vec4(col * (0.9 + 0.8 * vHeat), body * fade);
      }`,
  })
}

/** Smoke puffs: soft, grey to near-black with heat, warm where the flames light them from below. */
function smokeMaterial() {
  return new THREE.ShaderMaterial({
    transparent: true,
    depthWrite: false,
    uniforms: { uLight: { value: 1 } },
    vertexShader: `
      attribute vec4 aPuff;
      attribute vec3 aLook;
      varying vec2 vUv;
      varying vec3 vLook;
      void main() {
        ${BILLBOARD}
        vec3 p = aPuff.xyz + (camRight * position.x + camUp * position.y) * aPuff.w * 2.0;
        gl_Position = projectionMatrix * viewMatrix * vec4(p, 1.0);
        vUv = uv;
        vLook = aLook;
      }`,
    fragmentShader: `
      uniform float uLight;
      varying vec2 vUv;
      varying vec3 vLook;
      void main() {
        vec2 q = vUv - 0.5;
        float d = length(q);
        // A lumpy edge, different for every puff.
        float lump = 0.06 * sin(atan(q.y, q.x) * 5.0 + vLook.x * 30.0);
        float a = (1.0 - smoothstep(0.12, 0.5 + lump, d)) * vLook.x;
        if (a <= 0.003) discard;
        vec3 col = mix(vec3(0.72, 0.71, 0.69), vec3(0.16, 0.15, 0.15), vLook.y) * uLight;
        col += vec3(1.0, 0.45, 0.12) * vLook.z * 0.55;
        gl_FragColor = vec4(col, a);
      }`,
  })
}

/** Embers: tiny bright sparks that twinkle as they rise. */
function emberMaterial() {
  return new THREE.ShaderMaterial({
    transparent: true,
    depthWrite: false,
    blending: THREE.AdditiveBlending,
    uniforms: { uTime: { value: 0 } },
    vertexShader: `
      attribute vec4 aPuff;
      attribute vec3 aLook;
      uniform float uTime;
      varying vec2 vUv;
      varying float vAlpha;
      varying float vHeat;
      void main() {
        ${BILLBOARD}
        vec3 p = aPuff.xyz + (camRight * position.x + camUp * position.y) * aPuff.w * 2.0;
        gl_Position = projectionMatrix * viewMatrix * vec4(p, 1.0);
        vUv = uv;
        vAlpha = aLook.x * (0.6 + 0.4 * sin(uTime * 23.0 + aLook.z * 50.0));
        vHeat = aLook.y;
      }`,
    fragmentShader: `
      varying vec2 vUv;
      varying float vAlpha;
      varying float vHeat;
      void main() {
        float d = length(vUv - 0.5);
        float a = (1.0 - smoothstep(0.1, 0.5, d)) * vAlpha;
        if (a <= 0.003) discard;
        gl_FragColor = vec4(mix(vec3(1.0, 0.35, 0.05), vec3(1.0, 0.85, 0.4), vHeat) * 2.0, a);
      }`,
  })
}

/** A heap of ash with a few charred beams: what is left of a house. */
function ashGeometry() {
  const parts: THREE.BufferGeometry[] = []
  const heap = new THREE.CircleGeometry(0.42, 9)
  const pos = heap.getAttribute('position') as THREE.BufferAttribute
  for (let i = 1; i < pos.count; i++) {
    // (The rim's last vertex is its first again.)
    const k = 0.75 + 0.35 * hash3((i - 1) % 9, 3, 17)
    pos.setXY(i, pos.getX(i) * k, pos.getY(i) * k)
  }
  heap.rotateX(-Math.PI / 2)
  // A low mound: the middle lifted.
  pos.setY(0, 0.06)
  parts.push(colored(heap.toNonIndexed(), '#3b3734'))
  const beams: [number, number, number, number][] = [
    [0.05, 0.03, 0.4, 0.6],
    [-0.12, 0.05, -0.05, -0.4],
    [0.14, 0.04, -0.15, 1.3],
  ]
  for (const [x, y, z, r] of beams) {
    const b = new THREE.BoxGeometry(0.5, 0.05, 0.06)
    b.rotateZ(0.12)
    b.rotateY(r)
    b.translate(x, y, z)
    parts.push(colored(b.toNonIndexed(), '#1c1816'))
  }
  const out = new THREE.BufferGeometry()
  const count = parts.reduce((n, p) => n + p.getAttribute('position').count, 0)
  const P = new Float32Array(count * 3)
  const C = new Float32Array(count * 3)
  let o = 0
  for (const p of parts) {
    P.set(p.getAttribute('position').array as Float32Array, o * 3)
    C.set(p.getAttribute('color').array as Float32Array, o * 3)
    o += p.getAttribute('position').count
    p.dispose()
  }
  out.setAttribute('position', new THREE.BufferAttribute(P, 3))
  out.setAttribute('color', new THREE.BufferAttribute(C, 3))
  out.computeVertexNormals()
  return out
}

function colored(g: THREE.BufferGeometry, color: string) {
  const c = new THREE.Color(color)
  const n = g.getAttribute('position').count
  const cols = new Float32Array(n * 3)
  for (let i = 0; i < n; i++) cols.set([c.r, c.g, c.b], i * 3)
  g.setAttribute('color', new THREE.BufferAttribute(cols, 3))
  return g
}
