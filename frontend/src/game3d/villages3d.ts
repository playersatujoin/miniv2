// Villages in 3D, after RAGE's map zones (game/game/MapZones.h: a named area
// bounded by the convex hull of its points). Each village's land is a soft
// band in its colour laid along the inside of its hull, draped over the
// ground, with its name over the middle; both fade in as the camera pulls
// back (close up the houses and people speak for themselves). Close up, the
// village's leader carries a small pennant in its colour, fluttering
// downwind.

import * as THREE from 'three'
import type { Village } from '../sim/protocol'
import type { Terrain } from './terrain'

/** Smooths a closed polygon (two rounds of Chaikin's corner cutting) and resamples it every `step` tiles. */
export function softHull(hull: readonly [number, number][], step = 0.5): [number, number][] {
  let pts = hull.map(([x, y]) => [x, y] as [number, number])
  for (let round = 0; round < 2 && pts.length >= 3; round++) {
    const next: [number, number][] = []
    for (let i = 0; i < pts.length; i++) {
      const [ax, ay] = pts[i]
      const [bx, by] = pts[(i + 1) % pts.length]
      next.push([ax * 0.75 + bx * 0.25, ay * 0.75 + by * 0.25], [ax * 0.25 + bx * 0.75, ay * 0.25 + by * 0.75])
    }
    pts = next
  }
  const out: [number, number][] = []
  for (let i = 0; i < pts.length; i++) {
    const [ax, ay] = pts[i]
    const [bx, by] = pts[(i + 1) % pts.length]
    const n = Math.max(1, Math.ceil(Math.hypot(bx - ax, by - ay) / step))
    for (let k = 0; k < n; k++) out.push([ax + ((bx - ax) * k) / n, ay + ((by - ay) * k) / n])
  }
  return out
}

/** Unit normals pointing into the polygon at each point (whichever way it winds). */
export function inwardNormals(pts: readonly [number, number][]) {
  let cx = 0
  let cy = 0
  for (const [x, y] of pts) {
    cx += x / pts.length
    cy += y / pts.length
  }
  return pts.map(([x, y], i) => {
    const [px, py] = pts[(i - 1 + pts.length) % pts.length]
    const [nx, ny] = pts[(i + 1) % pts.length]
    let tx = nx - px
    let ty = ny - py
    const len = Math.hypot(tx, ty) || 1
    tx /= len
    ty /= len
    let ox = -ty
    let oy = tx
    if (ox * (cx - x) + oy * (cy - y) < 0) {
      ox = -ox
      oy = -oy
    }
    return [ox, oy] as [number, number]
  })
}

type Label = { el: HTMLDivElement; name: HTMLSpanElement; count: HTMLSpanElement; hue: number }

export class VillageLayer {
  readonly group = new THREE.Group()
  private mesh: THREE.Mesh<THREE.BufferGeometry, THREE.ShaderMaterial> | null = null
  private readonly material = outlineMaterial()
  private key = ''
  private villages: readonly Village[] = []
  private readonly labels = new Map<number, Label>()
  private banners: THREE.InstancedMesh
  private readonly bannerUniforms = { uTime: { value: 0 }, uWind: { value: 0.3 } }
  private readonly m = new THREE.Matrix4()
  private readonly q = new THREE.Quaternion()
  private readonly v = new THREE.Vector3()
  private readonly s = new THREE.Vector3()
  private readonly c = new THREE.Color()

  constructor(private readonly labelLayer: HTMLElement) {
    this.banners = this.makeBanners(4)
    this.group.add(this.banners)
  }

  /** Transparent effects the ambient occlusion pass must not see as solid. */
  get effects(): THREE.Object3D[] {
    return this.mesh ? [this.mesh] : []
  }

  /** The villages as streamed; the outlines are rebuilt only when a hull changes. */
  setVillages(villages: readonly Village[] | null, terrain: Terrain, force = false) {
    this.villages = villages ?? []
    const key = this.villages.map((v) => `${v.id}:${v.hue}:${v.hull.map((p) => p.join(',')).join(';')}`).join('|')
    if (key === this.key && !force) return
    this.key = key
    if (this.mesh) {
      this.group.remove(this.mesh)
      this.mesh.geometry.dispose()
      this.mesh = null
    }
    const pos: number[] = []
    const edge: number[] = []
    const col: number[] = []
    const index: number[] = []
    const c = this.c
    for (const v of this.villages) {
      if (v.hull.length < 3) continue
      const pts = softHull(v.hull)
      const normals = inwardNormals(pts)
      let r = 0
      for (const [x, y] of v.hull) r = Math.max(r, Math.hypot(x - v.x, y - v.y))
      const band = Math.min(2.6, Math.max(1.2, r * 0.12))
      c.setHSL(v.hue / 360, 0.85, 0.6)
      const base = pos.length / 3
      pts.forEach(([x, z], i) => {
        const [nx, nz] = normals[i]
        for (const [k, w] of [
          [0, -0.12],
          [1, band],
        ] as const) {
          const px = x + nx * w
          const pz = z + nz * w
          const ground = terrain.heightAt(px, pz)
          const water = terrain.waterAt(px, pz)
          pos.push(px, Math.max(ground, Number.isFinite(water) ? water : -Infinity) + 0.09, pz)
          edge.push(k)
          col.push(c.r, c.g, c.b)
        }
      })
      const n = pts.length
      for (let i = 0; i < n; i++) {
        const a = base + i * 2
        const b = base + ((i + 1) % n) * 2
        index.push(a, a + 1, b, b, a + 1, b + 1)
      }
    }
    if (!pos.length) return
    const g = new THREE.BufferGeometry()
    g.setAttribute('position', new THREE.Float32BufferAttribute(pos, 3))
    g.setAttribute('aEdge', new THREE.Float32BufferAttribute(edge, 1))
    g.setAttribute('color', new THREE.Float32BufferAttribute(col, 3))
    g.setIndex(index)
    g.computeBoundingSphere()
    this.mesh = new THREE.Mesh(g, this.material)
    this.mesh.renderOrder = 1
    this.group.add(this.mesh)
  }

  /**
   * Per frame: the outlines' and names' strength (0 hides them), names placed over each village's
   * middle, and pennants over the leaders who are close enough to see.
   */
  update(
    fade: number,
    camera: THREE.PerspectiveCamera,
    width: number,
    height: number,
    terrain: Terrain,
    leaders: readonly { x: number; y: number; z: number; id: number }[],
    bannerSize: number,
    wind: { x: number; z: number; strength: number },
    time: number,
  ) {
    this.material.uniforms.uFade.value = fade
    if (this.mesh) this.mesh.visible = fade > 0.01
    const seen = new Set<number>()
    for (const v of this.villages) {
      seen.add(v.id)
      let label = this.labels.get(v.id)
      if (!label) {
        const el = document.createElement('div')
        el.className = 'label3d'
        el.setAttribute('aria-hidden', 'true')
        el.style.fontWeight = '600'
        el.style.fontSize = '13px'
        const name = document.createElement('span')
        const count = document.createElement('span')
        count.style.fontWeight = '400'
        count.style.opacity = '0.75'
        count.style.marginLeft = '6px'
        el.append(name, count)
        this.labelLayer.append(el)
        label = { el, name, count, hue: -1 }
        this.labels.set(v.id, label)
      }
      const show = fade > 0.02
      if (!show) {
        label.el.style.display = 'none'
        continue
      }
      const p = this.v.set(v.x, terrain.heightAt(v.x, v.y) + 1.2, v.y).project(camera)
      if (p.z > 1 || p.x < -1.2 || p.x > 1.2 || p.y < -1.2 || p.y > 1.2) {
        label.el.style.display = 'none'
        continue
      }
      // (Only what changed: these are DOM writes, every frame.)
      const count = `${v.people} jiwa`
      if (label.name.textContent !== v.name) label.name.textContent = v.name
      if (label.count.textContent !== count) label.count.textContent = count
      if (label.hue !== v.hue) {
        label.hue = v.hue
        label.el.style.borderColor = `hsl(${v.hue} 70% 60%)`
      }
      label.el.style.display = 'block'
      label.el.style.opacity = fade.toFixed(2)
      label.el.style.transform = `translate(${((p.x + 1) / 2) * width}px, ${((1 - p.y) / 2) * height}px) translate(-50%, -50%)`
    }
    for (const [id, label] of this.labels) {
      if (seen.has(id)) continue
      label.el.remove()
      this.labels.delete(id)
    }

    // Pennants: on a short pole over the head, the cloth streaming downwind.
    const n = bannerSize > 0.01 ? leaders.length : 0
    if (n > this.banners.instanceMatrix.count) {
      const old = this.banners
      this.banners = this.makeBanners(Math.max(n, old.instanceMatrix.count * 2))
      this.group.remove(old)
      old.dispose()
      this.group.add(this.banners)
    }
    this.bannerUniforms.uTime.value = time
    this.bannerUniforms.uWind.value = wind.strength
    this.q.setFromAxisAngle(UP, -Math.atan2(wind.z, wind.x))
    this.s.setScalar(bannerSize)
    for (let i = 0; i < n; i++) {
      const l = leaders[i]
      this.m.compose(this.v.set(l.x, l.y, l.z), this.q, this.s)
      this.banners.setMatrixAt(i, this.m)
      const village = this.villages.find((v) => v.leader?.id === l.id)
      this.banners.setColorAt(i, this.c.setHSL((village?.hue ?? 45) / 360, 0.75, 0.55))
    }
    this.banners.count = n
    this.banners.instanceMatrix.needsUpdate = true
    if (this.banners.instanceColor) this.banners.instanceColor.needsUpdate = true
  }

  private makeBanners(cap: number) {
    const mat = new THREE.MeshStandardMaterial({ vertexColors: true, flatShading: true, roughness: 0.8, side: THREE.DoubleSide })
    const u = this.bannerUniforms
    mat.onBeforeCompile = (shader) => {
      shader.uniforms.uTime = u.uTime
      shader.uniforms.uWind = u.uWind
      shader.vertexShader = shader.vertexShader
        .replace('#include <common>', '#include <common>\nattribute float aFlap;\nuniform float uTime;\nuniform float uWind;')
        .replace(
          '#include <begin_vertex>',
          `#include <begin_vertex>
          // The cloth ripples more towards its free end, and harder in a stronger wind.
          transformed.z += sin(aFlap * 16.0 - uTime * (6.0 + 6.0 * uWind) + float(gl_InstanceID)) * aFlap * (0.12 + 0.25 * uWind);`,
        )
        .replace(
          '#include <color_vertex>',
          `vColor = vec4(1.0);
          vColor.rgb *= color;
          #ifdef USE_INSTANCING_COLOR
            if (aFlap > 0.0) vColor.rgb *= instanceColor.rgb;
          #endif`,
        )
    }
    const mesh = new THREE.InstancedMesh(pennantGeometry(), mat, cap)
    mesh.count = 0
    mesh.castShadow = true
    mesh.frustumCulled = false
    mesh.setColorAt(0, new THREE.Color(1, 1, 1))
    return mesh
  }

  dispose() {
    this.mesh?.geometry.dispose()
    this.material.dispose()
    this.banners.geometry.dispose()
    ;(this.banners.material as THREE.Material).dispose()
    this.banners.dispose()
    for (const l of this.labels.values()) l.el.remove()
    this.labels.clear()
  }
}

const UP = new THREE.Vector3(0, 1, 0)

/** The band: a crisp line just inside the hull, then a glow fading inwards. */
function outlineMaterial() {
  return new THREE.ShaderMaterial({
    transparent: true,
    depthWrite: false,
    polygonOffset: true,
    polygonOffsetFactor: -2,
    polygonOffsetUnits: -4,
    vertexColors: true,
    uniforms: { uFade: { value: 0 } },
    vertexShader: `
      attribute float aEdge;
      varying float vEdge;
      varying vec3 vCol;
      void main() {
        vEdge = aEdge;
        vCol = color;
        gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
      }`,
    fragmentShader: `
      uniform float uFade;
      varying float vEdge;
      varying vec3 vCol;
      void main() {
        // A dark edge outside the line so it reads on any ground, the line in the village's
        // colour, then a glow of it fading inwards.
        float shade = smoothstep(0.0, 0.03, vEdge) * (1.0 - smoothstep(0.05, 0.09, vEdge));
        float line = smoothstep(0.06, 0.1, vEdge) * (1.0 - smoothstep(0.15, 0.24, vEdge));
        float glow = pow(1.0 - vEdge, 2.4) * 0.3;
        float a = max(max(shade * 0.45, line * 0.95), glow) * uFade;
        if (a <= 0.003) discard;
        vec3 col = mix(vCol * 0.25, mix(vCol, vec3(1.0), 0.15), smoothstep(0.04, 0.08, vEdge));
        gl_FragColor = vec4(col, a);
      }`,
  })
}

/** A short pole with a swallow-tailed pennant at its top, streaming along +x. `aFlap`: 0 on the pole, distance along the cloth on it. */
function pennantGeometry() {
  const pos: number[] = []
  const col: number[] = []
  const flap: number[] = []
  const pole = new THREE.CylinderGeometry(0.012, 0.014, 0.42, 5).translate(0, 0.21, 0).toNonIndexed()
  const pp = pole.getAttribute('position')
  for (let i = 0; i < pp.count; i++) {
    pos.push(pp.getX(i), pp.getY(i), pp.getZ(i))
    col.push(0.36, 0.25, 0.16)
    flap.push(0)
  }
  pole.dispose()
  // The cloth: a strip in a few sections, so it can ripple, with a notch at its end.
  const sections = 5
  const length = 0.26
  const top = 0.42
  const drop = 0.13
  for (let k = 0; k < sections; k++) {
    const x0 = (k / sections) * length
    const x1 = ((k + 1) / sections) * length
    const narrow = (x: number) => (x / length) * 0.035
    const a = [x0, top - narrow(x0)]
    const b = [x1, top - narrow(x1)]
    const cBottom = [x1, top - drop + narrow(x1) + (k === sections - 1 ? 0.04 : 0)]
    const d = [x0, top - drop + narrow(x0)]
    for (const [x, y] of [a, d, b, b, d, cBottom]) {
      pos.push(x + 0.01, y, 0)
      col.push(1, 1, 1)
      flap.push(Math.max(0.001, x / length))
    }
  }
  const g = new THREE.BufferGeometry()
  g.setAttribute('position', new THREE.Float32BufferAttribute(pos, 3))
  g.setAttribute('color', new THREE.Float32BufferAttribute(col, 3))
  g.setAttribute('aFlap', new THREE.Float32BufferAttribute(flap, 1))
  g.computeVertexNormals()
  return g
}
