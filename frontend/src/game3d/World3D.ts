// The living island in 3D for watch mode: a low-poly diorama under a moving
// sky. It only draws what the simulation streams (people, animals, buildings,
// fields, weather) and never changes it. Edit and play modes stay in 2D.

import * as THREE from 'three'
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js'
import { EffectComposer } from 'three/examples/jsm/postprocessing/EffectComposer.js'
import { GTAOPass } from 'three/examples/jsm/postprocessing/GTAOPass.js'
import { OutputPass } from 'three/examples/jsm/postprocessing/OutputPass.js'
import { RenderPass } from 'three/examples/jsm/postprocessing/RenderPass.js'
import { ShaderPass } from 'three/examples/jsm/postprocessing/ShaderPass.js'
import { HorizontalTiltShiftShader } from 'three/examples/jsm/shaders/HorizontalTiltShiftShader.js'
import { VerticalTiltShiftShader } from 'three/examples/jsm/shaders/VerticalTiltShiftShader.js'
import { VignetteShader } from 'three/examples/jsm/shaders/VignetteShader.js'
import type { GameMap, MapRelief, TileSet } from '../api/client'
import type { MapView } from '../game/engine'
import { LiveWorld } from '../sim/live'
import {
  ANIMAL_FLAG,
  FLAG,
  PLOT_FLAG,
  WATER,
  type FieldPlot,
  type MinedOut,
  type MosquitoMessage,
  type SimFrame,
  type StructureFrame,
  type WaterMessage,
} from '../sim/protocol'
import {
  animalGeometries,
  buildingModel,
  birdGeometry,
  bushGeometry,
  flowerGeometry,
  leafyGeometry,
  palmGeometry,
  grassTuftGeometry,
  pineGeometry,
  reedGeometry,
  riceGeometry,
  rockGeometry,
  stumpGeometry,
  treeGeometry,
  wallGeometry,
} from './models'
import { Director, findSubjects, frameShot, smoothDamp, type Framing, type Subject, type Target } from './cinematic'
import { FireLayer } from './fireLayer'
import { Crowd, type Placement } from './people'
import { easeQuad, footSpots, quadPose, solveQuad, type QuadBuild, type QuadPose } from './quadruped'
import { lightCycle } from './timecycle'
import { bridgeGeometry, buildTerrain, oceanGeometry, reliefScale, terrainGeometry, waterGeometry, type Terrain } from './terrain'
import { VillageLayer } from './villages3d'
import { WaterLife, freshAt, waterSurface, waveAt, type WaterPlacement } from './water3d'
import { Lightning, WindTracker } from './weather3d'

export type World3DEvents = {
  /** The user clicked a person, or a building (→ its owner), or empty ground (→ null). */
  onSelectCreature?: (id: number | null) => void
  /** Following stopped because the user moved the camera. */
  onFollowChange?: (follow: boolean) => void
  /** The automatic camera turned to something (null when it is off), e.g. { id: 12, label: 'Rumah terbakar' }. */
  onCinematicSubject?: (subject: { id: number | null; label: string } | null) => void
}

const COMMUNAL = new Set(['saluran_irigasi', 'lumbung', 'kandang', 'ladang'])
const FLAT = new Set(['ladang', 'saluran_irigasi'])
/** How much each kind of plant bends in the wind (stiffness 0 = none). */
const SWAYING = new Set(['tree', 'pine', 'bush', 'palm', 'rice', 'leafy'])

const hash = (x: number, y: number, s: number) => {
  let h = (x * 374761393 + y * 668265263 + s * 2147483647) | 0
  h = Math.imul(h ^ (h >>> 13), 1274126177)
  return ((h ^ (h >>> 16)) >>> 0) / 4294967296
}

/** A smoothed copy of the stream's weather (storm 0–1). */
type Sky = { moisture: number; rain: number; enso: number; light: number; storm: number }

/** A late-morning sun from the south-west: long, soft shadows that model the relief. */
const SUN_DIR = new THREE.Vector3(-0.75, 0.72, 0.5).normalize()
/** People and animals are drawn larger than life so a crowd still reads, and grow a little more when seen from afar. */
const PERSON_SCALE = 1.6
const ANIMAL_SCALE = 1.25
/** Camera tilt (from straight down): the 2D map's view, and the diorama's. */
const FLAT_PHI = 0.0015
const VIEW_PHI = 0.9
/**
 * Field of view (degrees) of the diorama, and of the "flat" camera that stands in for the 2D map:
 * far away with a long lens, the island looks almost orthographic, like the map. Going between
 * them is a dolly zoom (the Jaws effect): distance × tan(fov / 2) stays constant, so the ground
 * under the camera keeps its size on screen while the perspective deepens.
 */
const VIEW_FOV = 42
const FLAT_FOV = 4
/** From this far away each person also gets a marker above the head, so a village never vanishes into the trees. */
const PIN_FROM = 24

export class World3D {
  private readonly renderer: THREE.WebGLRenderer
  private readonly scene = new THREE.Scene()
  private readonly camera: THREE.PerspectiveCamera
  private readonly controls: OrbitControls
  private readonly sun: THREE.DirectionalLight
  private readonly hemi: THREE.HemisphereLight
  private readonly sunDirection = SUN_DIR.clone()
  private readonly skyDome: THREE.Mesh
  private readonly timer = new THREE.Timer()
  private readonly composer: EffectComposer
  private readonly tiltH: ShaderPass
  private readonly tiltV: ShaderPass
  private readonly vignette: ShaderPass
  /** Ambient occlusion: soft contact shade where things meet the ground and each other. */
  private readonly gtao: GTAOPass
  /**
   * Quality watch: what a frame really costs (the CPU's work and, where the browser can time it,
   * the GPU's), not the time between frames, which a hidden or covered window stretches. Level 2
   * is everything, 1 drops ambient occlusion, 0 also renders at one pixel per CSS pixel.
   */
  private perf = { cost: 8, slowFor: 0, fastFor: 0, level: 2, age: 0, raisedAt: -Infinity, locked: false }
  /** Times the GPU's work on a frame, a few frames late (EXT_disjoint_timer_query_webgl2). */
  private gpuTimer: { ext: { TIME_ELAPSED_EXT: number; GPU_DISJOINT_EXT: number }; query: WebGLQuery | null; ms: number } | null = null
  /** The sky (with the sun) as a blurred environment map, for the water to reflect. */
  private readonly pmrem: THREE.PMREMGenerator
  private readonly envSky = makeEnvSky()
  private envMap: THREE.WebGLRenderTarget | null = null
  private envKey = ''
  private envAt = -Infinity
  private waterMats: THREE.MeshStandardMaterial[] = []
  private autoCentered = false
  /** How much larger people are drawn at the current distance, and their markers' size (0 = hidden). */
  private crowdScale = 1
  private pinSize = 0
  /** A scripted camera move: tilting up from the 2D map, or flattening back down to it. */
  private tween: {
    /** Seconds; negative while waiting to start. */
    t: number
    dur: number
    phi: [number, number]
    theta: [number, number]
    fov: [number, number]
    /** Tiles across the view at the target. */
    across: [number, number]
    done?: () => void
  } | null = null
  /** Called right after the frame that finished a flatten is drawn, while its pixels can still be copied. */
  private finished: (() => void) | null = null
  /** Following just started: move the camera in close to the person. */
  private closeIn = false
  /** Fire, smoke and scorched land; villages' land and names; ripples and rafts. */
  private readonly fire = new FireLayer(1, 1)
  private readonly villages: VillageLayer
  private readonly waterLife = new WaterLife()
  private readonly uniforms = {
    uTime: { value: 0 },
    uWind: { value: 0.3 },
    /** Where the wind blows towards (x, z), a unit vector. */
    uWindDir: { value: new THREE.Vector2(Math.cos(0.42), Math.sin(0.42)) },
    /** Per tile: R how scorched, G how hard it burns (FireLayer's texture). */
    uBurnt: { value: this.fire.texture },
    uDry: { value: 0 },
    uLush: { value: 0 },
    // The water cycle, one texel per tile: R how full, G how much it runs
    // (255 running, about 80 pools, 0 dry). Land by a river follows the river.
    uWater: { value: waterTexture(1, 1) as THREE.DataTexture },
    uMapSize: { value: new THREE.Vector2(1, 1) },
  }
  private lastWater: WaterMessage | null = null
  /** The water in the irrigation channels. */
  private readonly channelMat = waterMaterial(this.uniforms)
  private readonly reducedMotion = typeof window.matchMedia === 'function' ? window.matchMedia('(prefers-reduced-motion: reduce)') : null
  private raf = 0
  private resize: ResizeObserver

  private terrain: Terrain
  private relief: MapRelief | null = null
  private logged = new Set<number>()
  private world = new THREE.Group() // everything that depends on the terrain
  private landChunks = new Map<string, THREE.Mesh>()
  private plants = new THREE.Group()
  /** Grass tufts on open ground, kept off the planted plots. */
  private grass: THREE.InstancedMesh | null = null
  private grassPlots = ''
  private buildings = new THREE.Group()
  private fields = new THREE.Group()
  private clouds = new THREE.Group()
  private cloudMat: THREE.MeshStandardMaterial | null = null
  /** Flocks wheeling over the island: white egrets and dark swallows. */
  private birds: THREE.InstancedMesh
  private flocks: { ax: number; az: number; fx: number; fz: number; ph: number; n: number; spread: number; color: THREE.Color }[] = []
  private rain: THREE.LineSegments
  private smoke: THREE.Points
  /** Mosquito swarms over where they breed, and the stream message they were built from. */
  private mosquitoes: THREE.Points
  private mosquitoMsg: MosquitoMessage | null = null
  private smokeSources: { x: number; y: number; z: number; heavy: boolean; seed: number }[] = []

  /** The islanders (posed bodies) and, from afar, a marker over each. */
  private crowd = new Crowd()
  private frameDt = 0
  /** Geometries and materials made once and reused by every rebuild (fields change each second). */
  private readonly made = new Map<string, THREE.BufferGeometry | THREE.Material>()
  private pins: THREE.InstancedMesh
  private pinIds: number[] = []
  private ring: THREE.Mesh
  private structureList: StructureFrame[] = []
  private plotList: FieldPlot[] = []
  private herds: { mesh: THREE.InstancedMesh; ids: number[] }[]
  /** People and animals as streamed and smoothed; the page shares one copy with the 2D map. */
  private live = new LiveWorld()
  private sky: Sky = { moisture: 0.69, rain: 1, enso: 0, light: 1, storm: 0 }
  private skyTarget: Sky | null = null
  private simRate = 1
  private lastSim = -1
  private lastSimAt = 0
  /** The wind as drawn (with gusts), 0 – ~1.4, and the tracker that smooths the stream's. */
  private wind = 0.3
  private readonly windTracker = new WindTracker()
  /** Whether this server streams wind at all (old ones send none: the drawn wind is made up locally). */
  private streamWind = false
  private windTarget: { dir: number; strength: number; storm: boolean } | null = null
  private readonly lightning = new Lightning()
  private flash = 0
  private bolt: THREE.LineSegments
  private boltStrikes = 0
  /** What the shared live world held when last drawn, to notice new scorched land and villages. */
  private seenBurnt: unknown = undefined
  private seenVillages: unknown = undefined
  /** Where ripples and rafts go (reused every frame). */
  private readonly waterPlacement: WaterPlacement = {
    water: (x, z) => this.waterAt(x, z),
    scale: PERSON_SCALE,
    time: 0,
    still: false,
    inView: (x, y, z, size) => FRUSTUM.intersectsSphere(SPHERE.set(SPOT.set(x, y + 0.3 * size, z), 0.6 * size + 0.2)),
    camX: 0,
    camZ: 0,
  }
  /** The wind for fire and pennants, reused every frame. */
  private readonly fireWind = { x: 1, z: 0, strength: 0.3 }
  /** How big the leaders' pennants are drawn (0: hidden). */
  private villageBanners = 0
  /**
   * Placement for the crowd, reused every frame. `water` is the drawn water surface (world y, with
   * its swell) under a point, or null on dry land: swimmers float at it, rafters kneel on the raft
   * at water + RAFT_DECK × scale (see water3d.ts). Optional for people.ts, harmless if unused.
   */
  private readonly placement: Placement & { water: (x: number, z: number) => number | null }
  /** Close-up animals' feet on the ground (QuadLegSolver), eased per animal. */
  private readonly gaits = new Map<number, QuadPose>()
  private readonly quadTarget = quadPose()
  private readonly feet = new Float32Array(8)
  private readonly feetY = new Float32Array(4)
  /** The automatic camera, while it has the camera. */
  private cinematic: {
    director: Director
    /** Seconds since it started (its clock for decisions and shots). */
    t: number
    nextDecision: number
    seed: number
    shotStart: number
    /** Eye x, y, z and look x, y, z, each followed by its velocity. */
    state: Float64Array
    /** A quick fade to a far subject: 0 → 1 out, then the cut, then back. */
    fade: number
    fading: 0 | 1 | -1
    hints: { subject: Subject; until: number }[]
  } | null = null
  private readonly framing: Framing = { eyeX: 0, eyeY: 0, eyeZ: 0, lookX: 0, lookY: 0, lookZ: 0 }

  private selectedId: number | null = null
  private following = false
  private hoverId: number | null = null
  private press: { x: number; y: number; button: number } | null = null
  private readonly raycaster = new THREE.Raycaster()
  private readonly pointer = new THREE.Vector2()
  private lastHover = 0
  private labels: { selected: HTMLDivElement; hover: HTMLDivElement }

  constructor(
    private readonly container: HTMLElement,
    labelLayer: HTMLElement,
    private map: GameMap,
    private readonly tiles: TileSet,
    private readonly events: World3DEvents,
  ) {
    const r = new THREE.WebGLRenderer({ antialias: true, powerPreference: 'high-performance' })
    r.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2))
    r.shadowMap.enabled = true
    r.shadowMap.type = THREE.PCFShadowMap
    r.toneMapping = THREE.ACESFilmicToneMapping
    r.toneMappingExposure = 1.05
    r.outputColorSpace = THREE.SRGBColorSpace
    r.domElement.className = 'stage-canvas stage-canvas-3d'
    r.domElement.tabIndex = 0
    container.prepend(r.domElement)
    this.renderer = r
    this.pmrem = new THREE.PMREMGenerator(r)
    const timerExt = r.getContext().getExtension('EXT_disjoint_timer_query_webgl2')
    if (timerExt) this.gpuTimer = { ext: timerExt, query: null, ms: 0 }

    this.camera = new THREE.PerspectiveCamera(VIEW_FOV, 1, 0.1, 3000)
    this.controls = new OrbitControls(this.camera, r.domElement)
    this.controls.enableDamping = true
    this.controls.dampingFactor = 0.08
    this.controls.screenSpacePanning = false
    this.controls.maxPolarAngle = 1.32
    this.controls.minDistance = 3
    this.controls.mouseButtons = { LEFT: THREE.MOUSE.ROTATE, MIDDLE: THREE.MOUSE.DOLLY, RIGHT: THREE.MOUSE.PAN }

    // Sky and light.
    this.skyDome = makeSkyDome()
    this.scene.add(this.skyDome)
    this.scene.fog = new THREE.Fog('#cfe6f5', 60, 400)
    this.hemi = new THREE.HemisphereLight('#cde8ff', '#4d6a3a', 0.85)
    this.scene.add(this.hemi)
    this.sun = new THREE.DirectionalLight('#fff1d6', 3.1)
    this.sun.castShadow = true
    this.sun.shadow.mapSize.set(4096, 4096)
    this.sun.shadow.bias = -0.00035
    this.sun.shadow.normalBias = 0.025
    this.scene.add(this.sun, this.sun.target)

    this.scene.add(this.world, this.clouds)
    this.world.add(this.plants, this.buildings, this.fields)

    this.villages = new VillageLayer(labelLayer)
    this.fire.resize(map.width, map.height)
    this.uniforms.uBurnt.value = this.fire.texture
    this.terrain = buildTerrain(map, tiles, null)
    this.buildLand()
    this.scene.add(this.fire.group, this.villages.group, this.waterLife.group)
    this.placement = {
      ground: (x, z) => this.terrain.heightAt(x, z),
      water: this.waterAt,
      scale: PERSON_SCALE,
      time: 0,
      dt: 0,
      still: false,
      inView: (x, y, z, size) => FRUSTUM.intersectsSphere(SPHERE.set(SPOT.set(x, y + 0.3 * size, z), 0.6 * size + 0.2)),
      contact: false,
      wind: { dirX: 1, dirZ: 0, strength: 0.3 },
    }

    // People, and a marker in the family colour over each when seen from afar.
    this.pins = instanced(new THREE.OctahedronGeometry(0.5, 0).scale(0.7, 1, 0.7), new THREE.MeshBasicMaterial({ color: '#ffffff' }), 64)
    this.pins.castShadow = this.pins.receiveShadow = false
    this.scene.add(this.crowd.mesh, this.pins)
    // A gold ring at the feet of the person being watched, as on the 2D map.
    this.ring = new THREE.Mesh(
      new THREE.RingGeometry(0.34, 0.46, 28).rotateX(-Math.PI / 2),
      new THREE.MeshBasicMaterial({ color: '#f5c542', transparent: true, opacity: 0.9, depthWrite: false }),
    )
    this.ring.visible = false
    this.scene.add(this.ring)
    this.herds = animalGeometries().map((g) => {
      const mesh = instanced(g, trottingMaterial(this.uniforms), 32)
      setGait(mesh, 32)
      return { mesh, ids: [] }
    })
    for (const h of this.herds) this.scene.add(h.mesh)

    this.birds = this.makeBirds()
    this.scene.add(this.birds)
    this.rain = makeRain()
    this.scene.add(this.rain)
    this.smoke = makeSmoke()
    this.scene.add(this.smoke)
    this.mosquitoes = makeMosquitoes()
    this.scene.add(this.mosquitoes)
    this.bolt = makeBolt()
    this.scene.add(this.bolt)
    this.buildClouds()

    const selected = document.createElement('div')
    selected.className = 'label3d label3d-selected'
    const hover = document.createElement('div')
    hover.className = 'label3d'
    labelLayer.append(selected, hover)
    this.labels = { selected, hover }

    // North is up, as on the 2D map.
    const center = new THREE.Vector3(map.spawn?.x ?? map.width / 2, 0, map.spawn?.y ?? map.height / 2)
    center.y = this.terrain.heightAt(center.x, center.z)
    this.controls.target.copy(center)
    this.camera.position.copy(center).add(new THREE.Vector3().setFromSphericalCoords(30, VIEW_PHI, 0))
    this.controls.maxDistance = Math.max(map.width, map.height) * 1.4

    // Reaching for the camera takes it back from the automatic director, before the controls see the press.
    r.domElement.addEventListener('pointerdown', this.takeCamera, { capture: true })
    r.domElement.addEventListener('pointerdown', this.onPointerDown)
    r.domElement.addEventListener('pointerup', this.onPointerUp)
    r.domElement.addEventListener('pointermove', this.onPointerMove)
    r.domElement.addEventListener('wheel', this.landTween, { passive: true })
    r.domElement.addEventListener('wheel', this.takeCamera, { passive: true, capture: true })
    r.domElement.addEventListener('contextmenu', (e) => e.preventDefault())
    window.addEventListener('keydown', this.onKey)

    // Post-processing: a tilt-shift blur at the top and bottom makes the island read as a
    // miniature diorama, and a vignette frames it.
    this.composer = new EffectComposer(r)
    this.composer.addPass(new RenderPass(this.scene, this.camera))
    this.gtao = new GTAOPass(this.scene, this.camera, 512, 512)
    this.gtao.output = GTAOPass.OUTPUT.Default
    this.gtao.blendIntensity = 0.9
    this.gtao.updateGtaoMaterial({
      radius: 0.65,
      distanceExponent: 1,
      thickness: 1.2,
      scale: 1,
      samples: 12,
      distanceFallOff: 1,
      screenSpaceRadius: false,
    })
    this.gtao.updatePdMaterial({ lumaPhi: 10, depthPhi: 2, normalPhi: 3, radius: 5, rings: 2, samples: 16 })
    // At half resolution: the shade is soft anyway, and it costs a quarter.
    const aoSize = this.gtao.setSize.bind(this.gtao)
    this.gtao.setSize = (w: number, h: number) => aoSize(Math.max(1, w >> 1), Math.max(1, h >> 1))
    // The occlusion pass redraws the scene without our vertex animation (people would stand in
    // their rest pose holding every prop, far grass at full height), so the animated crowds sit it out.
    const aoRender = this.gtao.render.bind(this.gtao)
    this.gtao.render = (...args: Parameters<GTAOPass['render']>) => {
      // (So do the see-through effects: flames, smoke, ripples, the villages' bands.)
      const out = [
        this.crowd.mesh,
        this.pins,
        this.birds,
        this.grass,
        ...this.fire.effects,
        ...this.waterLife.effects,
        ...this.villages.effects,
        this.bolt,
      ].filter((o): o is THREE.Object3D => !!o && o.visible)
      for (const o of out) o.visible = false
      aoRender(...args)
      for (const o of out) o.visible = true
    }
    this.composer.addPass(this.gtao)
    this.tiltH = new ShaderPass(HorizontalTiltShiftShader)
    this.tiltV = new ShaderPass(VerticalTiltShiftShader)
    this.tiltH.uniforms.r.value = this.tiltV.uniforms.r.value = 0.52
    this.composer.addPass(this.tiltH)
    this.composer.addPass(this.tiltV)
    this.vignette = new ShaderPass(VignetteShader)
    this.vignette.uniforms.offset.value = 0.95
    this.vignette.uniforms.darkness.value = 1.1
    this.composer.addPass(this.vignette)
    this.composer.addPass(new OutputPass())

    this.resize = new ResizeObserver(() => this.fit())
    this.resize.observe(container)
    this.fit()
    this.raf = requestAnimationFrame(this.loop)
  }

  destroy() {
    cancelAnimationFrame(this.raf)
    this.resize.disconnect()
    window.removeEventListener('keydown', this.onKey)
    // The automatic camera goes with the view.
    if (this.cinematic) {
      this.cinematic = null
      this.events.onCinematicSubject?.(null)
    }
    this.controls.dispose()
    this.fire.dispose()
    this.villages.dispose()
    this.waterLife.dispose()
    this.scene.traverse((o) => {
      const m = o as THREE.Mesh
      m.geometry?.dispose()
      const mat = m.material
      if (Array.isArray(mat)) mat.forEach((x) => x.dispose())
      else mat?.dispose()
    })
    this.crowd.dispose()
    for (const v of this.made.values()) v.dispose()
    if (this.gpuTimer?.query) (this.renderer.getContext() as WebGL2RenderingContext).deleteQuery(this.gpuTimer.query)
    this.composer.dispose()
    for (const pass of [this.gtao, this.tiltH, this.tiltV, this.vignette]) pass.dispose()
    this.envMap?.dispose()
    this.pmrem.dispose()
    this.envSky.traverse((o) => {
      const m = o as THREE.Mesh
      m.geometry?.dispose()
      ;(m.material as THREE.Material | undefined)?.dispose()
    })
    this.sun.shadow.dispose()
    this.renderer.dispose()
    // Give the WebGL context back now rather than whenever it is collected (switching views makes new ones).
    this.renderer.forceContextLoss()
    this.renderer.domElement.remove()
    this.labels.selected.remove()
    this.labels.hover.remove()
  }

  private get still() {
    return this.reducedMotion?.matches ?? false
  }

  // --- Inputs from the page -----------------------------------------------------

  setRelief(relief: MapRelief | null | undefined) {
    if (!relief || relief === this.relief) return
    this.relief = relief
    this.reshape()
  }

  setMap(map: GameMap) {
    if (map === this.map) return
    this.map = map
    this.reshape()
  }

  /** Rebuilds the land, then sets the buildings, fields and camera back down on it. */
  private reshape() {
    this.terrain = buildTerrain(this.map, this.tiles, this.relief)
    if (this.fire.resize(this.map.width, this.map.height)) this.uniforms.uBurnt.value = this.fire.texture
    this.fire.reground(this.terrain)
    this.villages.setVillages(this.live.villages, this.terrain, true)
    this.buildLand()
    this.setWater(this.lastWater)
    this.setStructures(this.structureList)
    this.setFields(this.plotList)
    const t = this.controls.target
    const lift = this.terrain.heightAt(t.x, t.z) - t.y
    t.y += lift
    this.camera.position.y += lift
  }

  setMinedOut(mined: MinedOut | null | undefined) {
    const next = new Set<number>((mined?.logged ?? []).map(([x, y]) => y * this.map.width + x))
    if (next.size === this.logged.size && [...next].every((i) => this.logged.has(i))) return
    this.logged = next
    this.buildPlants()
  }

  /**
   * Hands the camera to the automatic director (after RAGE's cinematic director), or takes it back.
   * While it has the camera it reports what it shows through onCinematicSubject; any camera input
   * from the user stops it (and reports null), as it stops following someone.
   */
  setCinematic(on: boolean) {
    if (!on) {
      this.stopCinematic()
      return
    }
    if (this.cinematic) return
    this.stopFollow()
    this.closeIn = false
    const state = new Float64Array(12)
    const t = this.controls.target
    const eye = this.camera.position
    state[0] = eye.x
    state[2] = eye.y
    state[4] = eye.z
    state[6] = t.x
    state[8] = t.y
    state[10] = t.z
    this.cinematic = {
      director: new Director(),
      t: 0,
      nextDecision: 0,
      seed: Math.random(),
      shotStart: 0,
      state,
      fade: 0,
      fading: 0,
      hints: [],
    }
    if (!this.tween) this.controls.enabled = false
  }

  /**
   * Something the automatic camera might show that only the page knows of (an event such as a birth):
   * offered for the next half minute at the given importance (1 low … 3 high).
   */
  cinematicHint(hint: { id: number | null; x: number; y: number; label: string; importance?: number }) {
    const c = this.cinematic
    if (!c) return
    const key = `event:${hint.id ?? `${Math.round(hint.x)}:${Math.round(hint.y)}`}`
    c.hints = c.hints.filter((h) => h.subject.key !== key)
    c.hints.push({
      subject: {
        key,
        kind: 'event',
        id: hint.id,
        partner: null,
        x: hint.x,
        z: hint.y,
        radius: 1,
        score: 14 * (hint.importance ?? 2),
        label: hint.label,
      },
      until: c.t + 30,
    })
  }

  private stopCinematic() {
    if (!this.cinematic) return
    this.cinematic = null
    if (!this.tween) this.controls.enabled = true
    this.events.onCinematicSubject?.(null)
  }

  /** The user reached for the camera: the automatic director lets go. */
  private takeCamera = () => {
    if (this.cinematic) this.stopCinematic()
  }

  setSelected(id: number | null) {
    this.selectedId = id
  }

  setFollow(follow: boolean) {
    // Following someone is taking the camera.
    if (follow) this.stopCinematic()
    if (follow && !this.following) this.closeIn = true
    this.following = follow
  }

  /** Where the camera looks (tiles) and how many tiles fit across the view there, for the 2D map. */
  getView(): MapView {
    const t = this.controls.target
    const s = new THREE.Spherical().setFromVector3(this.camera.position.clone().sub(t))
    return { x: t.x, y: t.z, across: this.acrossAt(s.radius, this.camera.fov), phi: s.phi, theta: s.theta }
  }

  /**
   * Takes over the 2D map's view: the same spot, as wide, seen straight from above through a long
   * lens (so it looks just like the map), then tilts up and widens into the diorama, turning to
   * the angle the 3D camera last had.
   */
  setView(view: MapView) {
    this.stopCinematic()
    const x = THREE.MathUtils.clamp(view.x, 0, this.map.width)
    const z = THREE.MathUtils.clamp(view.y, 0, this.map.height)
    this.controls.target.set(x, this.terrain.heightAt(x, z), z)
    this.autoCentered = true
    this.closeIn = false
    const across = this.acrossAt(this.clampDistance(this.distanceAt(view.across, VIEW_FOV)), VIEW_FOV)
    const phi = THREE.MathUtils.clamp(view.phi ?? VIEW_PHI, 0.25, this.controls.maxPolarAngle)
    const theta = view.theta ?? 0
    if (this.still) {
      this.tween = null
      this.placeCamera(phi, theta, VIEW_FOV, across)
      return
    }
    this.startTween({
      t: -0.4,
      dur: 1.6,
      phi: [FLAT_PHI, phi],
      theta: [0, theta],
      fov: [FLAT_FOV, VIEW_FOV],
      across: [view.across, across],
    })
  }

  /**
   * Lowers the camera to straight above, north up, through a long lens, `across` tiles wide: what
   * the 2D map will show. `done` runs right after that last frame is drawn.
   */
  flatten(across: number, done: () => void) {
    this.stopCinematic()
    if (this.still) {
      done()
      return
    }
    const t = this.controls.target
    const s = new THREE.Spherical().setFromVector3(this.camera.position.clone().sub(t))
    const now = this.acrossAt(s.radius, this.camera.fov)
    this.closeIn = false
    this.startTween({
      t: 0,
      dur: 0.75,
      phi: [s.phi, FLAT_PHI],
      theta: [s.theta, 0],
      fov: [this.camera.fov, FLAT_FOV],
      across: [now, across],
      done,
    })
  }

  private startTween(tween: NonNullable<typeof this.tween>) {
    this.tween = tween
    // The camera is scripted until it lands.
    this.controls.enabled = false
    this.placeCamera(tween.phi[0], tween.theta[0], tween.fov[0], tween.across[0])
  }

  /** Puts the camera at a tilt, heading and lens around the target, showing `across` tiles there. */
  private placeCamera(phi: number, theta: number, fov: number, across: number) {
    const dist = this.distanceAt(across, fov)
    this.camera.fov = fov
    // A long lens stands far off: keep the depth range tight around the island so it stays precise.
    this.camera.near = fov < VIEW_FOV ? Math.max(0.1, dist * 0.25) : 0.1
    this.camera.far = dist + 3000
    this.camera.updateProjectionMatrix()
    this.camera.position.copy(this.controls.target).add(new THREE.Vector3().setFromSphericalCoords(dist, phi, theta))
    this.camera.lookAt(this.controls.target)
  }

  /** Tiles across the view at the target, from that distance with that lens. */
  private acrossAt(dist: number, fov: number) {
    return dist * 2 * Math.tan(THREE.MathUtils.degToRad(fov / 2)) * this.camera.aspect
  }

  private distanceAt(across: number, fov: number) {
    return across / (2 * Math.tan(THREE.MathUtils.degToRad(fov / 2)) * this.camera.aspect)
  }

  private clampDistance(dist: number) {
    return THREE.MathUtils.clamp(dist, this.controls.minDistance, this.controls.maxDistance)
  }

  /** How far the camera would be with the diorama's lens for the same view: what sizes and blur go by. */
  private viewDistance() {
    const dist = this.camera.position.distanceTo(this.controls.target)
    return (dist * Math.tan(THREE.MathUtils.degToRad(this.camera.fov / 2))) / Math.tan(THREE.MathUtils.degToRad(VIEW_FOV / 2))
  }

  /** 0 looking straight down (like the map), 1 at the diorama's tilt and below. */
  private tilt() {
    const offset = this.camera.position.clone().sub(this.controls.target)
    const phi = Math.acos(THREE.MathUtils.clamp(offset.y / (offset.length() || 1), -1, 1))
    return THREE.MathUtils.smoothstep(phi, 0.1, 0.6)
  }

  zoomBy(dir: number) {
    this.stopCinematic()
    if (this.tween) return
    const offset = this.camera.position.clone().sub(this.controls.target)
    const len = THREE.MathUtils.clamp(offset.length() * (dir > 0 ? 0.8 : 1.25), this.controls.minDistance, this.controls.maxDistance)
    this.camera.position.copy(this.controls.target).add(offset.setLength(len))
  }

  /** Draws people and animals from this shared, smoothed copy of the stream (the page feeds it). */
  setLive(live: LiveWorld) {
    this.live = live
  }

  /** The rivers' and lakes' water from the stream, into the texture the water and land shaders read. */
  setWater(msg: WaterMessage | null) {
    this.lastWater = msg
    const W = this.map.width
    const H = this.map.height
    let tex = this.uniforms.uWater.value
    if (tex.image.width !== W || tex.image.height !== H) {
      tex.dispose()
      tex = waterTexture(W, H)
      this.uniforms.uWater.value = tex
      this.uniforms.uMapSize.value.set(W, H)
    }
    const data = tex.image.data as Uint8Array
    data.fill(255)
    if (msg) {
      const run = [0, 30, 80, 255]
      const level = new Uint8Array(W * H).fill(255)
      const running = new Uint8Array(W * H).fill(255)
      const fresh = new Uint8Array(W * H)
      const foul = new Uint8Array(W * H)
      msg.tiles.forEach((t, k) => {
        level[t] = msg.state[k] >= WATER.pools ? msg.level[k] : 0
        running[t] = run[msg.state[k]] ?? 255
        foul[t] = msg.foul?.[k] ?? 0
        fresh[t] = 1
      })
      for (let y = 0; y < H; y++) {
        for (let x = 0; x < W; x++) {
          const i = y * W + x
          let r = level[i]
          let g = running[i]
          let b = foul[i]
          if (!fresh[i]) {
            // Banks: the water drawn over them stands as high as the river beside them.
            let best = -1
            for (let dy = -1; dy <= 1; dy++) {
              for (let dx = -1; dx <= 1; dx++) {
                const nx = x + dx
                const ny = y + dy
                if (nx < 0 || ny < 0 || nx >= W || ny >= H) continue
                const j = ny * W + nx
                if (fresh[j] && level[j] > best) {
                  best = level[j]
                  g = running[j]
                  b = foul[j]
                }
              }
            }
            if (best >= 0) r = best
            else g = 255
          }
          data[i * 4] = r
          data[i * 4 + 1] = g
          // Blue: how clean the water is (255 clean … 0 foul with filth).
          data[i * 4 + 2] = 255 - b
        }
      }
    }
    tex.needsUpdate = true
  }

  private get creatures() {
    return this.live.creatures
  }

  /** A new simulation frame reached the shared LiveWorld: follow the weather and the pace of time. */
  onFrame(frame: SimFrame) {
    // The one being followed has died (or left the stream): stop, as the 2D map does.
    if (this.following && this.selectedId !== null && !this.live.creatures.has(this.selectedId)) this.stopFollow()
    const now = this.live.frameAt || performance.now()
    if (frame.weather) {
      if (this.lastSim >= 0 && now - this.lastSimAt > 30) {
        const rate = Math.max(0, (frame.time - this.lastSim) / ((now - this.lastSimAt) / 1000))
        this.simRate = this.simRate * 0.7 + rate * 0.3
      }
      this.lastSim = frame.time
      this.lastSimAt = now
      const first = !this.skyTarget
      const w = frame.weather
      this.skyTarget = { moisture: w.moisture, rain: w.rain, enso: w.enso, light: w.light, storm: w.storm ? 1 : 0 }
      if (first) this.sky = { ...this.skyTarget, rain: 0 }
      // Old servers send no wind (0 both ways): the drawn wind is then made up locally.
      if (w.wind > 0 || w.windDir !== 0 || w.storm) this.streamWind = true
      this.windTarget = this.streamWind ? { dir: w.windDir, strength: w.wind, storm: w.storm } : null
    }
    this.fire.setFires(frame.fires ?? [])
  }

  setStructures(list: StructureFrame[]) {
    this.structureList = list
    // A house gone where fire was leaves its ash.
    this.fire.setStructures(list, this.terrain)
    for (const child of [...this.buildings.children]) {
      this.buildings.remove(child)
      ;(child as THREE.Mesh).geometry?.dispose()
    }
    this.smokeSources = []
    this.buildChannels(list)
    for (const st of list) {
      const ruined = !st.ownerId && !COMMUNAL.has(st.kind)
      const x = st.x + 0.5
      const z = st.y + 0.5
      const y = this.groundAt(st.x, st.y)
      if (st.kind === 'saluran_irigasi') continue
      if (FLAT.has(st.kind)) {
        const mesh = new THREE.Mesh(FIELD_GEO, FIELD_MAT)
        mesh.position.set(x, y + 0.02, z)
        mesh.receiveShadow = true
        mesh.userData = { ownerId: st.ownerId }
        this.buildings.add(mesh)
        continue
      }
      const m = cachedBuilding(st.kind, st.level)
      const body = new THREE.Mesh(m.body, ruined ? RUIN_MAT : BODY_MAT)
      body.castShadow = true
      body.receiveShadow = true
      body.userData = { ownerId: st.ownerId }
      const group = new THREE.Group()
      group.add(body)
      if (m.roof) {
        const roof = new THREE.Mesh(m.roof, ruined ? RUIN_MAT : roofMaterial(st.hue))
        roof.castShadow = true
        roof.userData = { ownerId: st.ownerId }
        group.add(roof)
      }
      group.position.set(x, y, z)
      group.rotation.y = (hash(st.x, st.y, 31) < 0.5 ? 0 : Math.PI / 2) * (st.level > 0 ? 1 : 0)
      this.buildings.add(group)
      if ((st.level > 0 && st.ownerId) || st.kind === 'tungku') {
        const top = st.kind === 'tungku' ? 0.52 : st.level >= 3 ? 0.95 : st.level === 2 ? 0.8 : 0.85
        this.smokeSources.push({ x: x + (st.kind === 'tungku' ? 0.12 : 0.05), y: y + top, z, heavy: st.kind === 'tungku', seed: st.id })
      }
    }
  }

  /**
   * Irrigation channels as on the 2D map: an earth-banked ditch from the tile's middle towards
   * each neighbour it joins (another channel, a field, the river), draped over the ground, with
   * water running in it (the same water as the rivers). On a tile that is already water, the
   * river itself is the channel.
   */
  private buildChannels(list: StructureFrame[]) {
    const W = this.map.width
    const t = this.terrain
    const at = new Map(list.map((st) => [st.y * W + st.x, st]))
    const bank: number[] = []
    const bankCol: number[] = []
    const pos: number[] = []
    const ground: number[] = []
    const flows: number[] = []
    const earth = new THREE.Color('#6e4f31')
    const quad = (out: number[], a: number[], b: number[], c: number[], d: number[]) => out.push(...a, ...c, ...b, ...b, ...c, ...d)
    for (const st of list) {
      if (st.kind !== 'saluran_irigasi') continue
      const i = st.y * W + st.x
      if (t.water[i]) continue
      const cx = st.x + 0.5
      const cz = st.y + 0.5
      const cy = t.heightAt(cx, cz)
      let links = 0
      SIDES.forEach(([dx, dz], k) => {
        const nx = st.x + dx
        const nz = st.y + dz
        if (nx < 0 || nz < 0 || nx >= W || nz >= this.map.height) return
        const other = at.get(nz * W + nx)
        if (t.water[nz * W + nx] || other?.kind === 'saluran_irigasi' || other?.kind === 'ladang') links |= 1 << k
      })
      const arms = links || 2 | 8 // a lone channel still shows as a short ditch
      const ditch = (w: number, lift: number, out: number[], water: boolean) => {
        SIDES.forEach(([dx, dz], k) => {
          if (!(arms & (1 << k))) return
          const ex = cx + dx * 0.5
          const ez = cz + dz * 0.5
          const ey = t.heightAt(ex, ez)
          const px = -dz * w
          const pz = dx * w
          const verts = [
            [cx + px, cy + lift, cz + pz],
            [cx - px, cy + lift, cz - pz],
            [ex + px, ey + lift, ez + pz],
            [ex - px, ey + lift, ez - pz],
          ]
          quad(out, verts[0], verts[1], verts[2], verts[3])
          if (water) {
            // Downhill along the arm.
            const fall = Math.sign(cy - ey) || 1
            for (let v = 0; v < 6; v++) {
              // quad() lays the vertices out as centre, edge, centre, centre, edge, edge.
              ground.push((v === 0 || v === 2 || v === 3 ? cy : ey) - 0.22)
              flows.push(dx * 0.6 * fall, dz * 0.6 * fall)
            }
          }
        })
        // The middle, where the arms meet.
        quad(out, [cx - w, cy + lift, cz - w], [cx + w, cy + lift, cz - w], [cx - w, cy + lift, cz + w], [cx + w, cy + lift, cz + w])
        if (water)
          for (let v = 0; v < 6; v++) {
            ground.push(cy - 0.22)
            flows.push(0.2, 0.1)
          }
      }
      ditch(0.23, 0.015, bank, false)
      ditch(0.12, 0.03, pos, true)
    }
    for (let k = 0; k < bank.length / 3; k++) bankCol.push(earth.r, earth.g, earth.b)
    if (!pos.length) return
    const banks = new THREE.BufferGeometry()
    banks.setAttribute('position', new THREE.Float32BufferAttribute(bank, 3))
    banks.setAttribute('color', new THREE.Float32BufferAttribute(bankCol, 3))
    banks.computeVertexNormals()
    const banksMesh = new THREE.Mesh(banks, STILL_MAT)
    banksMesh.receiveShadow = true
    const water = new THREE.BufferGeometry()
    water.setAttribute('position', new THREE.Float32BufferAttribute(pos, 3))
    water.setAttribute('ground', new THREE.Float32BufferAttribute(ground, 1))
    water.setAttribute('flow', new THREE.Float32BufferAttribute(flows, 2))
    water.setAttribute('fresh', new THREE.Float32BufferAttribute(new Array(pos.length / 3).fill(1), 1))
    water.computeVertexNormals()
    const waterMesh = new THREE.Mesh(water, this.channelMat)
    waterMesh.receiveShadow = true
    this.buildings.add(banksMesh, waterMesh)
  }

  setFields(plots: FieldPlot[]) {
    this.plotList = plots
    const plotKey = plots
      .map((p) => p.y * this.map.width + p.x)
      .sort((a, b) => a - b)
      .join(',')
    if (plotKey !== this.grassPlots) {
      this.grassPlots = plotKey
      this.buildGrass()
    }
    for (const child of [...this.fields.children]) {
      this.fields.remove(child)
      if ((child as THREE.InstancedMesh).isInstancedMesh) (child as THREE.InstancedMesh).dispose()
    }
    if (plots.length === 0) return
    // Soil patches, then each crop as its own instanced model.
    const soil = new THREE.InstancedMesh(SOIL_GEO, SOIL_MAT, plots.length)
    soil.receiveShadow = true
    const kinds: Record<string, FieldPlot[]> = {}
    const m = new THREE.Matrix4()
    const c = new THREE.Color()
    plots.forEach((p, i) => {
      m.makeTranslation(p.x + 0.5, this.groundAt(p.x, p.y) + 0.015, p.y + 0.5)
      soil.setMatrixAt(i, m)
      soil.setColorAt(i, c.set(p.flags & PLOT_FLAG.irrigated ? '#45372a' : p.flags & PLOT_FLAG.farmland ? '#6b4a2e' : '#7d5a39'))
      const kind =
        p.crop === 0 ? 'rice' : p.crop === 1 || p.crop === 2 ? 'leafy' : p.crop === 3 ? 'banana' : p.crop === 4 ? 'coconut' : 'sago'
      ;(kinds[kind] ??= []).push(p)
    })
    this.fields.add(soil)
    for (const [kind, list] of Object.entries(kinds)) {
      const mesh = new THREE.InstancedMesh(
        this.once(`crop:${kind}`, CROP_GEO[kind]),
        this.once('sway:crop', () => swayMaterial(this.uniforms, 1.4)),
        list.length,
      )
      mesh.castShadow = true
      mesh.receiveShadow = true
      const q = new THREE.Quaternion()
      const s = new THREE.Vector3()
      list.forEach((p, i) => {
        const grow = [0.35, 0.6, 0.85, 1][p.stage] ?? 1
        s.setScalar(grow * (0.9 + 0.2 * hash(p.x, p.y, 61)))
        q.setFromAxisAngle(UP, hash(p.x, p.y, 62) * Math.PI * 2)
        m.compose(new THREE.Vector3(p.x + 0.5, this.groundAt(p.x, p.y), p.y + 0.5), q, s)
        mesh.setMatrixAt(i, m)
        const withered = p.flags & PLOT_FLAG.withered
        const ripe = p.stage === 3
        c.set(withered ? '#a08a50' : kind === 'rice' ? (ripe ? '#e2bb3c' : '#79c24d') : '#5fb04a')
        if (kind === 'banana' || kind === 'coconut' || kind === 'sago') c.set(withered ? '#b8a070' : '#ffffff')
        mesh.setColorAt(i, c)
      })
      this.fields.add(mesh)
    }
  }

  // --- Building the land ------------------------------------------------------------

  private buildLand() {
    for (const child of [...this.world.children]) {
      if (child === this.plants || child === this.buildings || child === this.fields) continue
      this.world.remove(child)
      ;(child as THREE.Mesh).geometry?.dispose()
    }
    const t = this.terrain
    this.landChunks.clear()
    this.waterMats = [
      this.once('water:inland', () => waterMaterial(this.uniforms)),
      this.once('water:sea', () => waterMaterial(this.uniforms)),
    ]
    for (const m of this.waterMats) m.envMap = this.envMap?.texture ?? null
    const water = new THREE.Mesh(waterGeometry(t), this.waterMats[0])
    water.receiveShadow = true
    const ocean = new THREE.Mesh(oceanGeometry(t), this.waterMats[1])
    ocean.receiveShadow = true
    this.world.add(ocean, water)
    const planks = bridgeGeometry(t)
    if (planks) {
      const bridges = new THREE.Mesh(planks, BRIDGE_MAT)
      bridges.castShadow = bridges.receiveShadow = true
      this.world.add(bridges)
    }
    this.buildPlants()
  }

  /** Resident terrain follows the camera, with a padded boundary to hide
   * transitions. Two chunks per frame bound upload work when zooming out. */
  private updateTerrainChunks() {
    this.camera.updateMatrixWorld()
    FRUSTUM.setFromProjectionMatrix(MAT.multiplyMatrices(this.camera.projectionMatrix, this.camera.matrixWorldInverse))
    const t = this.terrain,
      wanted = new Set<string>()
    const box = new THREE.Box3()
    let created = 0
    for (let z = 0; z < t.height; z += 32)
      for (let x = 0; x < t.width; x += 32) {
        const key = `${x}:${z}`
        box.min.set(x - 8, -3, z - 8)
        box.max.set(x + 40, reliefScale(t.width, t.height) * 1.5 + 4, z + 40)
        if (!FRUSTUM.intersectsBox(box)) continue
        wanted.add(key)
        if (this.landChunks.has(key) || created >= 2) continue
        const geometry = terrainGeometry(t, { x, z, width: Math.min(32, t.width - x), height: Math.min(32, t.height - z) })
        geometry.computeBoundingSphere()
        const mesh = new THREE.Mesh(
          geometry,
          this.once('terrain', () => terrainMaterial(this.uniforms)),
        )
        mesh.name = `terrain:${key}`
        mesh.receiveShadow = mesh.castShadow = true
        this.world.add(mesh)
        this.landChunks.set(key, mesh)
        created++
      }
    for (const [key, mesh] of this.landChunks)
      if (!wanted.has(key)) {
        mesh.removeFromParent()
        mesh.geometry.dispose()
        this.landChunks.delete(key)
      }
  }

  private buildPlants() {
    for (const child of [...this.plants.children]) {
      this.plants.remove(child)
      ;(child as THREE.InstancedMesh).dispose?.()
    }
    const objectKey = new Map(this.tiles.objects.map((o) => [o.id, o.key]))
    const spots: Record<string, number[]> = {}
    const W = this.map.width
    this.map.layers.objects.forEach((id, i) => {
      let key = objectKey.get(id)
      if (!key || key === 'none') return
      if ((key === 'tree' || key === 'pine') && this.logged.has(i)) key = 'stump'
      ;(spots[key] ??= []).push(i)
    })
    const m = new THREE.Matrix4()
    const q = new THREE.Quaternion()
    const s = new THREE.Vector3()
    const c = new THREE.Color()
    for (const [key, list] of Object.entries(spots)) {
      const make = PLANT_GEO[key]
      if (!make) continue
      const stiffness = key === 'bush' ? 0.5 : key === 'pine' ? 0.7 : 1
      // Flowers and stumps do not sway, but they char where fire passes like everything green.
      const mat = SWAYING.has(key)
        ? this.once(`sway:${stiffness}`, () => swayMaterial(this.uniforms, stiffness))
        : key === 'flowers' || key === 'stump'
          ? this.once('sway:0', () => swayMaterial(this.uniforms, 0))
          : STILL_MAT
      const mesh = new THREE.InstancedMesh(this.once(`plant:${key}`, make), mat, list.length)
      mesh.castShadow = key !== 'flowers'
      mesh.receiveShadow = true
      list.forEach((i, k) => {
        const x = i % W
        const y = Math.floor(i / W)
        const big = key === 'wall' ? 1 : 0.8 + 0.45 * hash(x, y, 41)
        const jx = key === 'wall' ? 0.5 : 0.3 + 0.4 * hash(x, y, 42)
        const jz = key === 'wall' ? 0.5 : 0.3 + 0.4 * hash(x, y, 43)
        s.set(big, big * (0.9 + 0.25 * hash(x, y, 44)), big)
        q.setFromAxisAngle(UP, key === 'wall' ? 0 : hash(x, y, 45) * Math.PI * 2)
        m.compose(new THREE.Vector3(x + jx, this.terrain.heightAt(x + jx, y + jz) - 0.02, y + jz), q, s)
        mesh.setMatrixAt(k, m)
        const tint = 0.85 + 0.3 * hash(x, y, 46)
        mesh.setColorAt(k, c.setRGB(tint, tint * (0.95 + 0.1 * hash(x, y, 47)), tint))
      })
      this.plants.add(mesh)
    }
    this.buildReeds()
    // (The loop above took the old meadow down with the rest.)
    this.grass = null
    this.buildGrass()
  }

  /** Reeds where rivers and lakes meet their banks. */
  private buildReeds() {
    const t = this.terrain
    const GW = 2 * t.width + 1
    const spots: [number, number, number][] = []
    for (let g = 0; g < t.grid.length; g++) {
      // A shore vertex of fresh water: in the water's edge, with dry land beside it.
      if (!t.wet[g] || !t.fresh[g] || t.grid[g] < t.surface[g] - 0.08) continue
      const gx = g % GW
      const gz = Math.floor(g / GW)
      const r = hash(gx, gz, 120)
      if (r < 0.45) continue
      const tile = Math.floor(gz / 2) * t.width + Math.floor(gx / 2)
      if (t.keys[tile] === 'bridge') continue
      spots.push([gx / 2 + (hash(gx, gz, 121) - 0.5) * 0.4, gz / 2 + (hash(gx, gz, 122) - 0.5) * 0.4, r])
    }
    if (!spots.length) return
    const mesh = new THREE.InstancedMesh(
      REED_GEO,
      this.once('sway:reed', () => swayMaterial(this.uniforms, 1.6)),
      spots.length,
    )
    mesh.castShadow = true
    mesh.receiveShadow = true
    const m = new THREE.Matrix4()
    const q = new THREE.Quaternion()
    const s = new THREE.Vector3()
    const c = new THREE.Color()
    spots.forEach(([x, z, r], k) => {
      s.setScalar(0.8 + 0.6 * r)
      q.setFromAxisAngle(UP, r * 50)
      m.compose(new THREE.Vector3(x, t.heightAt(x, z) - 0.03, z), q, s)
      mesh.setMatrixAt(k, m)
      mesh.setColorAt(k, c.setRGB(0.9 + 0.2 * r, 0.95 + 0.1 * r, 0.85 + 0.2 * r))
    })
    this.plants.add(mesh)
  }

  /** A meadow: tufts of grass over open grass and forest floor, swaying in the wind. */
  private buildGrass() {
    if (this.grass) {
      this.plants.remove(this.grass)
      this.grass.dispose()
      this.grass = null
    }
    const W = this.map.width
    const H = this.map.height
    const t = this.terrain
    const plots = new Set(this.plotList.map((p) => p.y * W + p.x))
    const objectKey = new Map(this.tiles.objects.map((o) => [o.id, o.key]))
    const spots: [number, number, number][] = []
    for (let y = 0; y < H; y++) {
      for (let x = 0; x < W; x++) {
        const i = y * W + x
        const k = t.keys[i]
        const per = k === 'grass' ? 4 : k === 'forest_floor' ? 2 : 0
        if (!per || plots.has(i)) continue
        const obj = objectKey.get(this.map.layers.objects[i])
        if (obj === 'wall' || obj === 'rock') continue
        for (let j = 0; j < per; j++) {
          const sx = x + 0.1 + 0.8 * hash(x, y, 90 + j)
          const sz = y + 0.1 + 0.8 * hash(x, y, 95 + j)
          // Not on a bank under the water's edge.
          if (t.heightAt(sx, sz) < t.waterAt(sx, sz) + 0.04) continue
          spots.push([sx, sz, hash(x, y, 99 + j)])
        }
      }
    }
    if (!spots.length) return
    const mesh = new THREE.InstancedMesh(
      GRASS_GEO,
      this.once('grass', () => grassMaterial(this.uniforms)),
      spots.length,
    )
    mesh.receiveShadow = true
    const m = new THREE.Matrix4()
    const q = new THREE.Quaternion()
    const s = new THREE.Vector3()
    const c = new THREE.Color()
    spots.forEach(([x, z, r], k) => {
      s.setScalar(0.75 + 0.7 * r)
      q.setFromAxisAngle(UP, r * 40)
      m.compose(new THREE.Vector3(x, t.heightAt(x, z) - 0.01, z), q, s)
      mesh.setMatrixAt(k, m)
      const v = 0.8 + 0.4 * hash(Math.floor(x * 3), Math.floor(z * 3), 3)
      mesh.setColorAt(k, c.setRGB(v * (0.92 + 0.16 * r), v, v * (0.85 + 0.2 * r)))
    })
    this.grass = mesh
    this.plants.add(mesh)
  }

  private makeBirds() {
    const W = this.map.width
    const H = this.map.height
    const flocks = Math.round(THREE.MathUtils.clamp((W * H) / 2600, 3, 8))
    let total = 0
    for (let f = 0; f < flocks; f++) {
      const egrets = f % 3 === 0
      const n = egrets ? 5 + Math.floor(hash(f, 1, 81) * 4) : 7 + Math.floor(hash(f, 2, 81) * 6)
      this.flocks.push({
        ax: W * (0.2 + 0.25 * hash(f, 3, 81)),
        az: H * (0.2 + 0.25 * hash(f, 4, 81)),
        fx: 0.018 + 0.02 * hash(f, 5, 81),
        fz: 0.014 + 0.02 * hash(f, 6, 81),
        ph: hash(f, 7, 81) * 100,
        n,
        spread: egrets ? 1.4 : 2.4,
        color: new THREE.Color(egrets ? '#f4f1ea' : '#2c3138'),
      })
      total += n
    }
    const mat = new THREE.MeshStandardMaterial({ vertexColors: true, flatShading: true, roughness: 0.7, side: THREE.DoubleSide })
    mat.onBeforeCompile = (shader) => {
      shader.uniforms.uTime = this.uniforms.uTime
      shader.vertexShader = shader.vertexShader.replace('#include <common>', '#include <common>\nuniform float uTime;').replace(
        '#include <begin_vertex>',
        `#include <begin_vertex>
          float beat = sin(uTime * 11.0 + float(gl_InstanceID) * 2.3);
          transformed.y += beat * abs(transformed.z) * 0.9;`,
      )
    }
    const mesh = new THREE.InstancedMesh(birdGeometry(), mat, total)
    mesh.castShadow = true
    mesh.frustumCulled = false
    return mesh
  }

  /** Each flock drifts on a slow figure over the island; its birds wheel around its centre. */
  private updateBirds(time: number) {
    const W = this.map.width
    const H = this.map.height
    const top = reliefScale(W, H) * 1.05 + 4
    const show = this.tilt()
    this.birds.visible = show > 0.05
    if (!this.birds.visible) return
    const m = new THREE.Matrix4()
    const q = new THREE.Quaternion()
    const s = new THREE.Vector3().setScalar(1.6 * show)
    const p = new THREE.Vector3()
    const ahead = new THREE.Vector3()
    // Bird k of a flock (seed: its number on the island) at time t.
    const at = (f: (typeof this.flocks)[number], k: number, seed: number, t: number, out: THREE.Vector3) => {
      const cx = W / 2 + f.ax * Math.sin(t * f.fx + f.ph)
      const cz = H / 2 + f.az * Math.sin(t * f.fz + f.ph * 1.7)
      const a = t * (0.35 + 0.1 * hash(seed, 1, 83)) + (k / f.n) * Math.PI * 2
      const r = f.spread * (0.6 + 0.5 * hash(seed, 2, 83))
      return out.set(cx + Math.cos(a) * r, top + 1.2 * Math.sin(t * 0.4 + seed) + f.spread * 0.4 * hash(seed, 3, 83), cz + Math.sin(a) * r)
    }
    let i = 0
    for (const f of this.flocks) {
      for (let k = 0; k < f.n; k++, i++) {
        const t = this.still ? 0 : time
        at(f, k, i, t, p)
        at(f, k, i, t + 0.15, ahead)
        q.setFromAxisAngle(UP, -Math.atan2(ahead.z - p.z, ahead.x - p.x))
        m.compose(p, q, s)
        this.birds.setMatrixAt(i, m)
        this.birds.setColorAt(i, f.color)
      }
    }
    this.birds.instanceMatrix.needsUpdate = true
    if (this.birds.instanceColor) this.birds.instanceColor.needsUpdate = true
  }

  private buildClouds() {
    const W = this.map.width
    const H = this.map.height
    const top = reliefScale(W, H) * 2.2 + 12
    const mat = new THREE.MeshStandardMaterial({
      color: '#ffffff',
      flatShading: true,
      roughness: 1,
      transparent: true,
      opacity: 0.9,
      depthWrite: false,
    })
    this.cloudMat = mat
    const count = Math.round(8 + (W * H) / 1600)
    for (let k = 0; k < count; k++) {
      const parts: THREE.BufferGeometry[] = []
      const blobs = 3 + Math.floor(hash(k, 0, 71) * 4)
      for (let b = 0; b < blobs; b++) {
        const g = new THREE.IcosahedronGeometry(1 + hash(k, b, 72) * 1.3, 0)
        g.scale(1.35, 0.7, 1)
        g.translate((b - blobs / 2) * 1.4 + hash(k, b, 73), hash(k, b, 74) * 0.6, (hash(k, b, 75) - 0.5) * 1.6)
        parts.push(g)
      }
      for (const g of parts) {
        const mesh = new THREE.Mesh(g, mat)
        mesh.castShadow = true
        const cloud = new THREE.Group()
        cloud.add(mesh)
        cloud.position.set(hash(k, 1, 76) * W, top + hash(k, 2, 77) * 5, hash(k, 3, 78) * H)
        cloud.userData = { speed: 0.6 + hash(k, 4, 79) * 0.8, k }
        this.clouds.add(cloud)
      }
    }
  }

  private groundAt(x: number, y: number) {
    return this.terrain.heightAt(x + 0.5, y + 0.5)
  }

  // --- Interaction -------------------------------------------------------------------

  /** Reaching for the camera during the tilt-up lands it at once. */
  private landTween = () => {
    if (this.tween && !this.tween.done) this.tween.t = this.tween.dur
  }

  private onPointerDown = (e: PointerEvent) => {
    this.landTween()
    this.press = { x: e.clientX, y: e.clientY, button: e.button }
    if (e.button === 2 && this.following) this.stopFollow() // panning leaves the followed person
  }

  private onPointerUp = (e: PointerEvent) => {
    const p = this.press
    this.press = null
    if (!p || p.button !== 0 || Math.hypot(e.clientX - p.x, e.clientY - p.y) > 5) return
    this.events.onSelectCreature?.(this.pick(e))
  }

  private onPointerMove = (e: PointerEvent) => {
    const now = performance.now()
    if (now - this.lastHover < 70 || this.press) return
    this.lastHover = now
    const hit = this.pickCreature(e)
    this.hoverId = hit
    this.renderer.domElement.style.cursor = hit !== null ? 'pointer' : ''
  }

  private onKey = (e: KeyboardEvent) => {
    if (e.ctrlKey || e.metaKey || e.altKey) return // shortcuts (Ctrl+S…) are not camera moves
    if ((e.target as HTMLElement | null)?.isContentEditable) return
    const tag = (e.target as HTMLElement | null)?.tagName
    if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return
    const step = 2.5
    const fwd = new THREE.Vector3()
    this.camera.getWorldDirection(fwd)
    fwd.y = 0
    fwd.normalize()
    const right = new THREE.Vector3(-fwd.z, 0, fwd.x)
    let move: THREE.Vector3 | null = null
    if (e.code === 'KeyW' || e.code === 'ArrowUp') move = fwd.multiplyScalar(step)
    if (e.code === 'KeyS' || e.code === 'ArrowDown') move = fwd.multiplyScalar(-step)
    if (e.code === 'KeyD' || e.code === 'ArrowRight') move = right.multiplyScalar(step)
    if (e.code === 'KeyA' || e.code === 'ArrowLeft') move = right.multiplyScalar(-step)
    if (!move) return
    this.stopFollow()
    this.stopCinematic()
    this.controls.target.add(move)
    this.camera.position.add(move)
  }

  private stopFollow() {
    if (!this.following) return
    this.following = false
    this.events.onFollowChange?.(false)
  }

  private setPointer(e: PointerEvent) {
    const rect = this.renderer.domElement.getBoundingClientRect()
    this.pointer.set(((e.clientX - rect.left) / rect.width) * 2 - 1, -((e.clientY - rect.top) / rect.height) * 2 + 1)
    this.raycaster.setFromCamera(this.pointer, this.camera)
  }

  private pickCreature(e: PointerEvent): number | null {
    this.setPointer(e)
    // Instanced meshes cache their bounds on the first raycast; who is drawn changes every frame.
    this.crowd.mesh.computeBoundingSphere()
    this.pins.computeBoundingSphere()
    const hits = this.raycaster.intersectObjects([this.crowd.mesh, this.pins], false)
    for (const h of hits) {
      if (h.instanceId === undefined) continue
      // Bodies of the dead are drawn by the same mesh but belong to no one: look past them.
      const id = (h.object === this.pins ? this.pinIds : this.crowd.ids)[h.instanceId]
      if (id != null) return id
    }
    return null
  }

  private pick(e: PointerEvent): number | null {
    const person = this.pickCreature(e)
    if (person !== null) return person
    const hits = this.raycaster.intersectObjects(this.buildings.children, true)
    for (const h of hits) {
      const owner = h.object.userData?.ownerId
      if (owner) return owner
    }
    return null
  }

  // --- Per frame ------------------------------------------------------------------------

  private fit() {
    const w = this.container.clientWidth
    const h = this.container.clientHeight
    if (!w || !h) return
    this.renderer.setSize(w, h, false)
    this.composer.setSize(w, h)
    this.camera.aspect = w / h
    this.camera.updateProjectionMatrix()
  }

  private loop = () => {
    this.raf = requestAnimationFrame(this.loop)
    const workStart = performance.now()
    this.timer.update()
    const dt = Math.min(0.05, this.timer.getDelta())
    this.frameDt = dt
    const time = this.timer.getElapsed()
    this.updateWeather(dt, time)
    this.live.advance(performance.now())
    this.updateFollow(dt)
    this.updateCinematic(dt)
    // While the camera is scripted the controls keep out (they would clamp the long-lens distance).
    if (this.controls.enabled) this.controls.update()
    this.updateTween(dt)
    this.updateTerrainChunks()
    // People are placed (and culled) with the camera where this frame will see them from.
    this.updateActors()
    this.updateSun()
    this.updateClouds(dt)
    this.updateBirds(time)
    this.updateRain(dt)
    this.updateSmoke(time)
    this.updateMosquitoes(time)
    this.updateFire(dt, time)
    this.updateVillages(time)
    this.updateTiltShift()
    this.timedRender()
    this.updateLabels()
    this.updateQuality(dt, performance.now() - workStart)
    const finished = this.finished
    this.finished = null
    finished?.()
  }

  private updateWeather(dt: number, time: number) {
    const target = this.skyTarget
    const k = 1 - Math.exp(-dt / (this.simRate >= 4 ? 8 : 1.5))
    // At fast playback a storm can come and go within a second: average it like the daylight.
    const calm = this.simRate > 2.5 || this.still
    if (target) {
      this.sky.moisture += (target.moisture - this.sky.moisture) * k
      this.sky.enso = target.enso
      // A day is two sim-seconds: average fast playback to avoid flashing.
      const daylight = calm ? 0.68 : target.light
      this.sky.light += (daylight - this.sky.light) * (1 - Math.exp(-dt / 0.65))
      const raining = !calm ? Math.max(THREE.MathUtils.clamp((target.rain - 1.2) / 0.5, 0, 1), target.storm) : 0
      this.sky.rain += (raining - this.sky.rain) * (1 - Math.exp(-dt / 0.8))
      this.sky.storm += (target.storm - this.sky.storm) * (1 - Math.exp(-dt / (calm ? 8 : 2.5)))
    }
    const dry = THREE.MathUtils.clamp((0.69 - this.sky.moisture) / 0.5, 0, 1)
    const lush = THREE.MathUtils.clamp((this.sky.moisture - 0.69) / 0.31, 0, 1)
    this.uniforms.uDry.value = dry
    this.uniforms.uLush.value = lush

    // The wind: the stream's, turned and eased towards (with gusts on top, stronger in a storm);
    // from old servers, the local breeze it always was.
    const w = this.windTracker
    if (this.windTarget) {
      const t = this.windTarget
      w.update(dt, { dir: t.dir, strength: Math.min(1, t.strength + 0.25 * this.sky.storm), storm: t.storm }, this.still)
      this.wind = w.blowing
    } else {
      const gust = 0.12 * Math.sin(time * 0.31) + 0.08 * Math.sin(time * 0.87 + 1.3)
      const windTarget = THREE.MathUtils.clamp(0.2 + 0.6 * this.sky.rain + 0.15 * this.sky.moisture + gust, 0, 1)
      this.wind += (windTarget - this.wind) * (1 - Math.exp(-dt / 3))
      w.update(dt, null, true)
    }
    this.uniforms.uWind.value = this.still ? 0 : this.wind
    this.uniforms.uWindDir.value.set(w.dirX, w.dirZ)
    this.uniforms.uTime.value = this.still ? 0 : time
    const pw = this.placement.wind!
    pw.dirX = w.dirX
    pw.dirZ = w.dirZ
    pw.strength = this.still ? 0 : Math.min(1, this.wind)

    // Lightning in a storm: never under reduced motion, nor at fast playback (it would strobe).
    this.flash = this.lightning.update(dt, this.sky.storm, !calm && this.tilt() > 0.5)
    this.updateBolt()

    // Overcast and rain dim and cool the light; the dry season is warm and hazy; a storm is darkest.
    const overcast = Math.max(Math.min(1, 0.6 * this.sky.rain + 0.25 * lush), 0.95 * this.sky.storm)
    const flash = this.flash
    const cycle = lightCycle(this.sky.light, overcast, this.sky.storm, flash)
    this.sun.intensity = cycle.sun
    this.sun.color.setRGB(1, 0.95 - 0.04 * dry, 0.84 - 0.1 * dry + 0.12 * overcast)
    this.sun.color.lerp(new THREE.Color('#f6a166'), cycle.warmth * 0.5).lerp(new THREE.Color('#8faee5'), 1 - cycle.day)
    this.hemi.intensity = cycle.ambient + 0.1 * overcast * (1 - 0.3 * this.sky.storm)
    this.renderer.toneMappingExposure = cycle.exposure * this.fadeLevel()
    const horizon = new THREE.Color('#d6ecf7').lerp(new THREE.Color('#e9dcc0'), dry * 0.6).lerp(new THREE.Color('#aeb8c2'), overcast)
    const zenith = new THREE.Color('#4f9be0').lerp(new THREE.Color('#78a8d0'), dry * 0.4).lerp(new THREE.Color('#7c8a99'), overcast)
    horizon.lerp(new THREE.Color('#e6a484'), cycle.warmth * 0.5).lerp(new THREE.Color('#26364f'), 1 - cycle.day)
    zenith.lerp(new THREE.Color('#111d36'), 1 - cycle.day)
    // A storm's sky: slate, heavier overhead.
    horizon.lerp(new THREE.Color('#5d6670'), 0.55 * this.sky.storm)
    zenith.lerp(new THREE.Color('#2f363f'), 0.65 * this.sky.storm)
    const phase = (this.lastSim / 2) * Math.PI * 2
    const movingSun =
      this.simRate <= 2.5 && !this.still
        ? new THREE.Vector3(-Math.cos(phase), 0.25 + Math.abs(Math.sin(phase)) * 0.75, 0.5).normalize()
        : SUN_DIR
    this.sunDirection.lerp(movingSun, 1 - Math.exp(-dt / 1.2)).normalize()
    const dome = this.skyDome.material as THREE.ShaderMaterial
    // The flash lights the clouds from within: the whole sky pales for a moment.
    dome.uniforms.uHorizon.value.copy(horizon).lerp(FLASH_SKY, 0.6 * flash)
    dome.uniforms.uZenith.value.copy(zenith).lerp(FLASH_SKY, 0.45 * flash)
    ;(this.scene.fog as THREE.Fog).color.copy(horizon)
    this.updateEnvironment(horizon, zenith, overcast, time)
  }

  /** How bright the picture is through a cinematic cut's fade (1 = fully). */
  private fadeLevel() {
    const c = this.cinematic
    return c ? 1 - c.fade : 1
  }

  /** A jagged bolt in the distance, downwind, for the moment of each strike. */
  private updateBolt() {
    const l = this.lightning
    const mat = this.bolt.material as THREE.LineBasicMaterial
    this.bolt.visible = this.flash > 0.05
    mat.opacity = Math.min(1, this.flash * 1.4)
    if (l.strikes === this.boltStrikes) return
    this.boltStrikes = l.strikes
    // Somewhere out past the island in front of the camera, from the clouds to the ground.
    const t = this.controls.target
    const fwd = new THREE.Vector3()
    this.camera.getWorldDirection(fwd)
    fwd.y = 0
    if (fwd.lengthSq() < 1e-6) fwd.set(0, 0, -1)
    fwd.normalize()
    const side = new THREE.Vector3(-fwd.z, 0, fwd.x)
    const dist = Math.max(40, this.viewDistance() * 1.5)
    const base = t
      .clone()
      .addScaledVector(fwd, dist)
      .addScaledVector(side, (l.where - 0.5) * dist)
    const top = reliefScale(this.map.width, this.map.height) * 2.2 + 14
    const pos = this.bolt.geometry.getAttribute('position') as THREE.BufferAttribute
    const segs = pos.count / 2
    let x = base.x
    let z = base.z
    let y = t.y + top
    const step = (top + 2) / segs
    for (let i = 0; i < segs; i++) {
      const r1 = hash(l.strikes, i, 91) - 0.5
      const r2 = hash(l.strikes, i, 92) - 0.5
      const nx = x + r1 * 2.2
      const nz = z + r2 * 2.2
      const ny = y - step
      pos.setXYZ(i * 2, x, y, z)
      pos.setXYZ(i * 2 + 1, nx, ny, nz)
      x = nx
      z = nz
      y = ny
    }
    pos.needsUpdate = true
  }

  /** Re-renders the sky the water reflects when the weather has changed it (at most every second and a half). */
  private updateEnvironment(horizon: THREE.Color, zenith: THREE.Color, overcast: number, time: number) {
    const key = [horizon.getHexString(), zenith.getHexString(), Math.round(overcast * 10)].join()
    if (key === this.envKey || time - this.envAt < 1.5) return
    this.envKey = key
    this.envAt = time
    const u = (this.envSky.children[0] as THREE.Mesh<THREE.BufferGeometry, THREE.ShaderMaterial>).material.uniforms
    u.uHorizon.value.copy(horizon)
    u.uZenith.value.copy(zenith)
    u.uSun.value = 1 - 0.85 * overcast
    const next = this.pmrem.fromScene(this.envSky, 0.015)
    this.envMap?.dispose()
    this.envMap = next
    for (const m of [...this.waterMats, this.channelMat]) {
      if (!m.envMap) m.needsUpdate = true
      m.envMap = next.texture
    }
  }

  /** The drawn water surface under a point, riding a gentle swell (null on dry land): what swimmers and rafts float on. */
  private readonly waterAt = (x: number, z: number) => {
    const t = this.terrain
    const y = waterSurface(t, this.lastWater ? (this.uniforms.uWater.value.image.data as Uint8Array) : null, x, z)
    if (y === null) return null
    return y + waveAt(x, z, y - t.heightAt(x, z), freshAt(t, x, z), this.uniforms.uTime.value, this.uniforms.uWind.value)
  }

  private updateActors() {
    const m = new THREE.Matrix4()
    const q = new THREE.Quaternion()
    const s = new THREE.Vector3()
    const p = new THREE.Vector3()
    const c = new THREE.Color()

    // People, a little larger the further the camera is so they never shrink to specks.
    const dist = this.viewDistance()
    const far = THREE.MathUtils.clamp((dist - 12) / 45, 0, 1)
    // Markers only for the diorama: seen straight down, the island should look like the map.
    const pinSize = THREE.MathUtils.clamp((dist - PIN_FROM) / 12, 0, 1) * Math.min(0.85, dist * 0.011) * this.tilt()
    this.crowdScale = 1 + 0.8 * far
    this.pinSize = pinSize
    const n = this.creatures.size
    this.camera.updateMatrixWorld()
    FRUSTUM.setFromProjectionMatrix(MAT.multiplyMatrices(this.camera.projectionMatrix, this.camera.matrixWorldInverse))
    const at = this.placement
    at.scale = PERSON_SCALE * this.crowdScale
    at.time = this.timer.getElapsed()
    at.dt = this.frameDt
    at.still = this.still
    at.contact = dist < 26
    this.crowd.update(this.creatures.values(), n, at)
    // Bodies where people died (the characters' ragdolls), after the living.
    this.crowd.updateBodies(this.live.frame?.corpses ?? [], at)
    // Their shadows only show up close, and from afar a coarser body does.
    this.crowd.mesh.castShadow = dist < 45
    this.crowd.setDetail(dist < 32)
    this.ensurePins(n)
    const pin = this.pins
    this.ring.visible = false
    let i = 0
    this.pinIds.length = 0
    const leaders = LEADERS
    leaders.length = 0
    // Leaders' pennants only close up, where the pins are gone.
    const banner = (1 - THREE.MathUtils.smoothstep(dist, 22, 30)) * this.tilt()
    for (const cr of this.creatures.values()) {
      this.pinIds.push(cr.id)
      const scale = cr.size * (cr.flags & FLAG.child ? 0.62 : 1) * PERSON_SCALE * this.crowdScale
      const ground = this.terrain.heightAt(cr.rx, cr.ry)
      const marked = cr.id === this.selectedId
      p.set(cr.rx, ground + scale * 0.62 + pinSize * 0.9, cr.ry)
      s.setScalar(pinSize * (marked ? 1.5 : 1))
      m.compose(p, IDENTITY, s)
      pin.setMatrixAt(i, m)
      pin.setColorAt(i, marked ? c.set('#f5c542') : c.setHSL(cr.hue / 360, 0.85, 0.5))
      if (marked) {
        this.ring.visible = true
        this.ring.position.set(cr.rx, ground + 0.03, cr.ry)
        this.ring.scale.setScalar(scale / PERSON_SCALE)
      }
      if (cr.flags & FLAG.leader && banner > 0.01 && leaders.length < 16) {
        const top = scale * 0.64
        if (at.inView!(cr.rx, ground, cr.ry, scale)) leaders.push({ x: cr.rx, y: ground + top, z: cr.ry, id: cr.id })
      }
      i++
    }
    pin.count = pinSize === 0 ? 0 : n
    pin.instanceMatrix.needsUpdate = true
    if (pin.instanceColor) pin.instanceColor.needsUpdate = true
    this.villageBanners = banner * this.crowdScale

    // People in the water: ripples round them, rafts under them.
    const t = this.controls.target
    const wl = this.waterPlacement
    wl.scale = at.scale
    wl.time = at.time
    wl.still = this.still
    wl.camX = t.x
    wl.camZ = t.z
    this.waterLife.update(this.creatures.values(), wl)

    // Animals, one mesh per species. Close up their feet find the ground (QuadLegSolver): the body
    // pitches and rolls with the slope and each leg reaches its own foothold.
    const close = dist < 34
    for (const h of this.herds) h.ids.length = 0
    const counts = this.herds.map(() => 0)
    for (const a of this.live.animals.values()) {
      const herd = this.herds[a.species]
      if (!herd) continue
      const idx = counts[a.species]++
      this.ensureHerd(a.species, idx + 1)
      const young = (a.flags & ANIMAL_FLAG.young) !== 0
      const bob = a.moving && !this.still ? Math.abs(Math.sin(a.walked * 2.4 * Math.PI)) * 0.025 : 0
      const size = (young ? 0.6 : 1) * ANIMAL_SCALE * (1 + 0.5 * far)
      const build = QUAD_BUILDS[a.species]
      const legs = herd.mesh.geometry.getAttribute('aLegs') as THREE.InstancedBufferAttribute
      let y = this.terrain.heightAt(a.rx, a.ry)
      let pitch = 0
      let roll = 0
      if (build && close && Math.abs(a.rx - t.x) < 30 && Math.abs(a.ry - t.z) < 30 && at.inView!(a.rx, y, a.ry, size)) {
        footSpots(a.rx, a.ry, a.rh, build, size, this.feet)
        for (let k = 0; k < 4; k++) this.feetY[k] = this.terrain.heightAt(this.feet[k * 2], this.feet[k * 2 + 1])
        solveQuad(this.feetY, build, size, this.quadTarget)
        let pose = this.gaits.get(a.id)
        if (!pose) {
          pose = quadPose()
          pose.y = this.quadTarget.y
          this.gaits.set(a.id, pose)
        }
        if (this.still) {
          pose.y = this.quadTarget.y
          pose.pitch = this.quadTarget.pitch
          pose.roll = this.quadTarget.roll
          pose.legs.set(this.quadTarget.legs)
        } else easeQuad(pose, this.quadTarget, this.frameDt)
        y = pose.y
        pitch = pose.pitch
        roll = pose.roll
        legs.setXYZW(idx, pose.legs[0], pose.legs[1], pose.legs[2], pose.legs[3])
      } else {
        this.gaits.delete(a.id)
        legs.setXYZW(idx, 1, 1, 1, 1)
      }
      p.set(a.rx, y + bob, a.ry)
      EULER.set(roll, -a.rh, pitch, 'YZX')
      q.setFromEuler(EULER)
      s.setScalar(size)
      m.compose(p, q, s)
      herd.mesh.setMatrixAt(idx, m)
      herd.mesh.setColorAt(idx, a.flags & ANIMAL_FLAG.tame ? c.setRGB(1.25, 1.12, 1.05) : c.setRGB(1, 1, 1))
      ;(herd.mesh.geometry.getAttribute('gait') as THREE.InstancedBufferAttribute).setXY(
        idx,
        a.walked * (ANIMAL_STRIDE[a.species] ?? 1.2) * Math.PI * 2,
        a.moving && !this.still ? 1 : 0,
      )
      herd.ids.push(a.id)
    }
    // Forget poses of animals that have wandered off or died.
    if (this.gaits.size > 256) for (const id of this.gaits.keys()) if (!this.live.animals.has(id)) this.gaits.delete(id)
    this.herds.forEach((h, sp) => {
      h.mesh.count = counts[sp]
      h.mesh.geometry.getAttribute('gait').needsUpdate = true
      h.mesh.geometry.getAttribute('aLegs').needsUpdate = true
      h.mesh.instanceMatrix.needsUpdate = true
      if (h.mesh.instanceColor) h.mesh.instanceColor.needsUpdate = true
    })
  }

  /** Fires, smoke and scorched land, with the wind as drawn. */
  private updateFire(dt: number, time: number) {
    if (this.live.burnt !== this.seenBurnt) {
      this.seenBurnt = this.live.burnt
      this.fire.setBurnt(this.live.burnt)
    }
    // (From the shared live world too, so a view opened mid-stream shows the fires at once.)
    if (this.live.frame) this.fire.setFires(this.live.frame.fires ?? [])
    const w = this.fireWind
    w.x = this.windTracker.dirX
    w.z = this.windTracker.dirZ
    // (Under reduced motion the smoke column stands still, but still leans the way the wind blows.)
    w.strength = this.wind
    this.fire.update(dt, time, this.controls.target, this.terrain, w, this.perf.level, this.crowdScale, this.sky.light, this.still)
  }

  /** Villages' land and names fade in as the camera pulls back; leaders' pennants show close up. */
  private updateVillages(time: number) {
    if (this.live.villages !== this.seenVillages) {
      this.seenVillages = this.live.villages
      this.villages.setVillages(this.live.villages, this.terrain)
    }
    const fade = THREE.MathUtils.smoothstep(this.viewDistance(), 16, 32)
    this.villages.update(
      fade,
      this.camera,
      this.container.clientWidth,
      this.container.clientHeight,
      this.terrain,
      LEADERS,
      this.villageBanners * 0.8,
      this.fireWind,
      this.still ? 0 : time,
    )
  }

  /**
   * The automatic camera: once a second the director weighs what is going on near the camera and
   * may turn to something else; every frame the shot's camera is followed with a critically
   * damped spring, so moves are smooth. A subject far away is reached by a quick fade instead
   * of a sweep across the island (a cut, under reduced motion).
   */
  private updateCinematic(dt: number) {
    const c = this.cinematic
    if (!c || this.tween) return
    c.t += dt
    const t = this.controls.target
    if (c.t >= c.nextDecision) {
      c.nextDecision = c.t + 1
      c.hints = c.hints.filter((h) => h.until > c.t)
      const subjects = findSubjects(
        {
          people: this.creatures.values(),
          fires: this.live.frame?.fires ?? [],
          corpses: this.live.frame?.corpses ?? [],
          villages: this.live.villages ?? [],
          hints: c.hints.map((h) => h.subject),
        },
        FLAG,
      )
      const d = c.director.update(c.t, subjects, t.x, t.z)
      if (d.changedSubject) {
        const sub = d.subject
        this.events.onCinematicSubject?.(sub ? { id: sub.id, label: sub.label } : { id: null, label: 'Mengamati pulau' })
        c.seed = Math.random()
        c.shotStart = c.t
        const far = sub !== null && Math.hypot(sub.x - c.state[6], sub.z - c.state[10]) > 40
        if (far || this.still) c.fading = this.still ? -1 : 1
        if (this.still) c.fade = 0
      } else if (d.changedShot) c.shotStart = c.t
    }

    // Where the shot wants the camera now.
    const f = this.framing
    const sub = c.director.current
    const shotTime = this.still ? 0 : c.t - c.shotStart
    if (sub) {
      let x = sub.x
      let z = sub.z
      let heading = 0
      const cr = sub.id !== null ? this.creatures.get(sub.id) : undefined
      if (cr) {
        x = cr.rx
        z = cr.ry
        heading = cr.rh
      }
      const mate = sub.partner !== null ? this.creatures.get(sub.partner) : undefined
      const y = this.standAt(x, z)
      const target = TARGET
      target.x = x
      target.y = y
      target.z = z
      target.heading = heading
      target.radius = sub.radius
      target.partnerX = mate?.rx
      target.partnerZ = mate?.ry
      target.partnerY = mate ? this.standAt(mate.rx, mate.ry) : undefined
      frameShot(c.director.shot, target, shotTime, c.seed, 1, f)
    } else {
      const target = TARGET
      target.x = t.x
      target.y = t.y
      target.z = t.z
      target.heading = 0
      target.radius = 10
      target.partnerX = target.partnerZ = target.partnerY = undefined
      frameShot('establishing', target, shotTime, c.seed, 1, f)
    }
    // Never under the ground or the water, and never lower than the controls allow (so handing
    // the camera back does not jolt it).
    f.eyeY = Math.max(f.eyeY, this.standAt(f.eyeX, f.eyeZ) + 0.8)
    const level = Math.hypot(f.eyeX - f.lookX, f.eyeZ - f.lookZ) / Math.tan(this.controls.maxPolarAngle - 0.04)
    f.eyeY = Math.max(f.eyeY, f.lookY + level)

    const st = c.state
    if (c.fading === 1) {
      // Fading out on the old subject; at black, cut.
      c.fade = Math.min(1, c.fade + dt / 0.35)
      if (c.fade >= 1) c.fading = -1
    }
    if (c.fading === -1 && (c.fade >= 1 || this.still)) {
      // The cut itself: straight to the new shot.
      st.set([f.eyeX, 0, f.eyeY, 0, f.eyeZ, 0, f.lookX, 0, f.lookY, 0, f.lookZ, 0])
    }
    if (c.fading !== 1) {
      if (c.fading === -1) {
        c.fade = Math.max(0, c.fade - dt / 0.5)
        if (c.fade <= 0) c.fading = 0
      }
      const glide = c.t - c.shotStart < 3 ? 1.4 : 0.9
      smoothDamp(st, 0, f.eyeX, glide, dt)
      smoothDamp(st, 2, f.eyeY, glide, dt)
      smoothDamp(st, 4, f.eyeZ, glide, dt)
      smoothDamp(st, 6, f.lookX, glide * 0.7, dt)
      smoothDamp(st, 8, f.lookY, glide * 0.7, dt)
      smoothDamp(st, 10, f.lookZ, glide * 0.7, dt)
    }
    if (this.camera.fov !== VIEW_FOV) {
      this.camera.fov = VIEW_FOV
      this.camera.near = 0.1
      this.camera.updateProjectionMatrix()
    }
    this.camera.position.set(st[0], st[2], st[4])
    t.set(st[6], st[8], st[10])
    this.camera.lookAt(t)
  }

  /** Where something stands: the ground, or the water's surface over it. */
  private standAt(x: number, z: number) {
    const ground = this.terrain.heightAt(x, z)
    const water = this.waterAt(x, z)
    return water === null ? ground : Math.max(ground, water)
  }

  private ensurePins(n: number) {
    const old = this.pins
    if (old.instanceMatrix.count >= n) return
    const next = instanced(old.geometry, old.material as THREE.Material, Math.max(n, old.instanceMatrix.count * 2))
    next.castShadow = next.receiveShadow = false
    this.scene.remove(old)
    old.dispose()
    this.scene.add(next)
    this.pins = next
  }

  private ensureHerd(species: number, n: number) {
    const h = this.herds[species]
    if (h.mesh.instanceMatrix.count >= n) return
    const next = instanced(h.mesh.geometry, h.mesh.material as THREE.Material, Math.max(n, h.mesh.instanceMatrix.count * 2))
    setGait(next, next.instanceMatrix.count)
    // Keep the instances already placed this frame.
    for (let i = 0; i < h.mesh.instanceMatrix.count; i++) {
      const m = new THREE.Matrix4()
      h.mesh.getMatrixAt(i, m)
      next.setMatrixAt(i, m)
    }
    this.scene.remove(h.mesh)
    h.mesh.dispose()
    this.scene.add(next)
    h.mesh = next
  }

  private updateFollow(dt: number) {
    // The first time people appear, look at where most of them are (unless following someone alive).
    const followed = this.following && this.selectedId !== null ? this.creatures.get(this.selectedId) : undefined
    if (!this.autoCentered && this.creatures.size > 0 && !followed) {
      this.autoCentered = true
      const xs = [...this.creatures.values()].map((c) => c.rx).sort((a, b) => a - b)
      const zs = [...this.creatures.values()].map((c) => c.ry).sort((a, b) => a - b)
      const x = xs[xs.length >> 1]
      const z = zs[zs.length >> 1]
      const goal = new THREE.Vector3(x, this.terrain.heightAt(x, z), z)
      this.camera.position.add(goal.clone().sub(this.controls.target))
      this.controls.target.copy(goal)
    }
    const c = followed
    if (!c) return
    this.autoCentered = true
    // Half a tile north of the feet, as the 2D map frames them, so handing over is exact.
    const goal = new THREE.Vector3(c.rx, this.terrain.heightAt(c.rx, c.ry - 0.5), c.ry - 0.5)
    const delta = goal.sub(this.controls.target).multiplyScalar(1 - Math.exp(-dt * 4))
    this.controls.target.add(delta)
    this.camera.position.add(delta)
    if (this.closeIn) {
      const offset = this.camera.position.clone().sub(this.controls.target)
      const len = offset.length()
      const want = 11
      if (Math.abs(len - want) < 0.3) this.closeIn = false
      else this.camera.position.copy(this.controls.target).add(offset.setLength(len + (want - len) * (1 - Math.exp(-dt * 3))))
    }
  }

  private updateTween(dt: number) {
    const tw = this.tween
    if (!tw) return
    tw.t = Math.min(tw.dur, tw.t + dt)
    const k = Math.max(0, tw.t) / tw.dur
    const e = k * k * (3 - 2 * k)
    const lerp = (r: [number, number]) => r[0] + (r[1] - r[0]) * e
    const turn = Math.atan2(Math.sin(tw.theta[1] - tw.theta[0]), Math.cos(tw.theta[1] - tw.theta[0]))
    // Zoom evenly in scale (geometric), so neither end rushes.
    const across = tw.across[0] * Math.pow(tw.across[1] / tw.across[0], e)
    this.placeCamera(lerp(tw.phi), tw.theta[0] + turn * e, lerp(tw.fov), across)
    if (k < 1) return
    this.tween = null
    if (tw.done) {
      this.finished = tw.done
      return
    }
    // (Unless the automatic camera took over meanwhile.)
    this.controls.enabled = !this.cinematic
  }

  private updateSun() {
    // The shadow box follows what the camera looks at, sized to the view.
    const t = this.controls.target
    const dist = this.camera.position.distanceTo(t)
    const half = THREE.MathUtils.clamp(this.viewDistance() * 0.9, 12, 90)
    const cam = this.sun.shadow.camera
    if (cam.right !== half) {
      cam.left = -half
      cam.right = half
      cam.top = half
      cam.bottom = -half
      cam.near = 1
      cam.far = 400
      cam.updateProjectionMatrix()
    }
    this.sun.position.copy(t).addScaledVector(this.sunDirection, 120)
    this.sun.target.position.copy(t)
    this.sun.target.updateMatrixWorld()
    const fog = this.scene.fog as THREE.Fog
    fog.near = dist * 1.6
    fog.far = dist * 5 + 120
  }

  private updateClouds(dt: number) {
    const W = this.map.width
    const H = this.map.height
    const cloudiness = THREE.MathUtils.clamp(0.25 + 0.6 * this.sky.moisture + 0.4 * this.sky.rain, 0, 1)
    const visible = Math.round(this.clouds.children.length * cloudiness)
    // Seen from high up the clouds thin out so the island shows through; their shadows stay.
    // Looking straight down (as the 2D map does) they clear away entirely.
    if (this.cloudMat) {
      const camHeight =
        (this.camera.position.y - this.controls.target.y) *
        (this.viewDistance() / (this.camera.position.distanceTo(this.controls.target) || 1))
      this.cloudMat.opacity = THREE.MathUtils.clamp(1.15 - camHeight / 60, 0.25, 0.9) * this.tilt()
      this.clouds.visible = this.cloudMat.opacity > 0.01
    }
    // Storm clouds: all of them, low and dark.
    if (this.cloudMat) this.cloudMat.color.setRGB(1, 1, 1).lerp(STORM_CLOUD, 0.75 * this.sky.storm)
    const dx = this.windTracker.dirX
    const dz = this.windTracker.dirZ
    this.clouds.children.forEach((cloud, i) => {
      cloud.visible = i < Math.max(visible, Math.round(this.clouds.children.length * this.sky.storm))
      if (this.still) return
      // Carried downwind, wrapping round the island.
      const speed = cloud.userData.speed * (0.4 + 1.2 * this.wind) * dt
      cloud.position.x += dx * speed
      cloud.position.z += dz * speed
      if (cloud.position.x > W + 20) cloud.position.x = -20
      else if (cloud.position.x < -20) cloud.position.x = W + 20
      if (cloud.position.z > H + 20) cloud.position.z = -20
      else if (cloud.position.z < -20) cloud.position.z = H + 20
    })
  }

  private updateRain(dt: number) {
    const on = this.sky.rain > 0.04
    this.rain.visible = on
    if (!on) return
    const storm = this.sky.storm
    const mat = this.rain.material as THREE.LineBasicMaterial
    mat.opacity = 0.15 + 0.3 * this.sky.rain + 0.15 * storm
    const pos = this.rain.geometry.attributes.position as THREE.BufferAttribute
    const t = this.controls.target
    const span = 40
    // Driven by the wind: it falls slanting downwind, harder and faster in a storm.
    const fall = 22 + 12 * storm
    const drift = 1 + 10 * Math.min(1.2, this.wind)
    const vx = this.windTracker.dirX * drift
    const vz = this.windTracker.dirZ * drift
    const streak = (0.32 + 0.2 * storm) / fall
    // Lighter rain draws fewer drops.
    const drops = Math.round((pos.count / 2) * Math.min(1, 0.35 + 0.65 * this.sky.rain + storm))
    for (let i = 0; i < drops * 2; i += 2) {
      let y = pos.getY(i) - fall * dt
      let x = pos.getX(i) + vx * dt
      let z = pos.getZ(i) + vz * dt
      if (y < t.y - 2 || Math.abs(x - t.x) > span || Math.abs(z - t.z) > span) {
        x = t.x + (Math.random() - 0.5) * span * 2
        z = t.z + (Math.random() - 0.5) * span * 2
        y = t.y + 6 + Math.random() * 18
      }
      pos.setXYZ(i, x, y, z)
      pos.setXYZ(i + 1, x + vx * streak, y - fall * streak, z + vz * streak)
    }
    this.rain.geometry.setDrawRange(0, drops * 2)
    pos.needsUpdate = true
  }

  private updateSmoke(time: number) {
    const geo = this.smoke.geometry
    const pos = geo.attributes.position as THREE.BufferAttribute
    const alpha = geo.attributes.alpha as THREE.BufferAttribute
    const size = geo.attributes.size as THREE.BufferAttribute
    const per = 6
    const n = Math.min(this.smokeSources.length * per, pos.count)
    const lean = 0.4 + 1.6 * this.wind
    const dx = this.windTracker.dirX
    const dz = this.windTracker.dirZ
    for (let i = 0; i < n; i++) {
      const src = this.smokeSources[Math.floor(i / per)]
      const k = i % per
      const age = this.still ? (k + 0.5) / per : (time * (src.heavy ? 0.4 : 0.3) + k / per + hash(src.seed, k, 3)) % 1
      const wobble = Math.sin(time + k + src.seed) * 0.05 * age
      pos.setXYZ(
        i,
        src.x + dx * lean * age * age - dz * wobble,
        src.y + (age * (src.heavy ? 1.6 : 1.2)) / (1 + 0.5 * this.wind),
        src.z + dz * lean * age * age + dx * wobble,
      )
      alpha.setX(i, Math.min(1, age * 4) * (1 - age) ** 1.3 * (src.heavy ? 0.85 : 0.7))
      size.setX(i, (src.heavy ? 0.35 : 0.25) + age * (src.heavy ? 0.9 : 0.6))
    }
    geo.setDrawRange(0, n)
    pos.needsUpdate = true
    alpha.needsUpdate = true
    size.needsUpdate = true
    const smokeMat = this.smoke.material as THREE.ShaderMaterial
    smokeMat.uniforms.uScale.value = this.renderer.domElement.height / 2
    // Grey against the night sky, not glowing.
    smokeMat.uniforms.uLight.value = 0.25 + 0.75 * this.sky.light
  }

  /**
   * Swarms of mosquitoes dancing over their breeding water, rebuilt whenever
   * the stream sends a new map; more of them are out from dusk to dawn, when
   * Anopheles bite. Which carry malaria can't be seen.
   */
  private updateMosquitoes(time: number) {
    const msg = this.live.mosquitoes
    const pts = this.mosquitoes
    const mat = pts.material as THREE.ShaderMaterial
    if (msg !== this.mosquitoMsg) {
      this.mosquitoMsg = msg
      const pos: number[] = []
      const seed: number[] = []
      if (msg) {
        const c = msg.cell
        for (let cy = 0; cy < msg.rows; cy++) {
          for (let cx = 0; cx < msg.cols; cx++) {
            const d = msg.density[cy * msg.cols + cx] / 255
            if (d < 0.08) continue
            const swarms = 1 + Math.floor(d * 2.5)
            for (let w = 0; w < swarms; w++) {
              const x = cx * c + 0.5 + hash(cx, cy, 700 + w) * (c - 1)
              const z = cy * c + 0.5 + hash(cx, cy, 710 + w) * (c - 1)
              const ground = this.terrain.heightAt(x, z)
              const water = this.terrain.waterAt(x, z)
              // A breeding cell can straddle dry land; waterAt is NaN there.
              const y = (Number.isFinite(water) ? Math.max(ground, water) : ground) + 0.45
              const n = Math.round(4 + 18 * d)
              for (let i = 0; i < n; i++) {
                pos.push(x, y, z)
                seed.push(hash(cx * 31 + w, cy, 720 + i), hash(cx, cy * 31 + w, 740 + i), (i + 0.5) / n)
              }
            }
          }
        }
      }
      const g = pts.geometry
      g.setAttribute('position', new THREE.BufferAttribute(new Float32Array(pos), 3))
      g.setAttribute('seed', new THREE.BufferAttribute(new Float32Array(seed), 3))
      g.setDrawRange(0, pos.length / 3)
      g.computeBoundingSphere()
    }
    const light = this.live.frame?.weather?.light ?? 0.5
    mat.uniforms.uTime.value = this.still ? 0 : time
    mat.uniforms.uActive.value = 0.35 + 0.65 * (1 - light)
    mat.uniforms.uScale.value = this.renderer.domElement.height / 2
  }

  private once<T extends THREE.BufferGeometry | THREE.Material>(key: string, make: () => T): T {
    let v = this.made.get(key) as T | undefined
    if (!v) this.made.set(key, (v = make()))
    return v
  }

  /** Renders the frame, timing the GPU's part when the browser allows it. */
  private timedRender() {
    const g = this.gpuTimer
    if (!g) {
      this.composer.render()
      return
    }
    const gl = this.renderer.getContext() as WebGL2RenderingContext
    if (g.query) {
      const ready = gl.getQueryParameter(g.query, gl.QUERY_RESULT_AVAILABLE)
      const disjoint = gl.getParameter(g.ext.GPU_DISJOINT_EXT)
      if (ready || disjoint) {
        if (ready && !disjoint) g.ms = g.ms * 0.8 + (gl.getQueryParameter(g.query, gl.QUERY_RESULT) / 1e6) * 0.2
        gl.deleteQuery(g.query)
        g.query = null
      }
    }
    if (g.query) {
      this.composer.render()
      return
    }
    g.query = gl.createQuery()
    gl.beginQuery(g.ext.TIME_ELAPSED_EXT, g.query)
    this.composer.render()
    gl.endQuery(g.ext.TIME_ELAPSED_EXT)
  }

  /**
   * Keeps the frame rate up. If a frame costs more than ~16 ms (the 60 fps budget) for three
   * seconds, ambient occlusion goes, then the pixel ratio drops to 1; with plenty of headroom for
   * five seconds a level comes back, unless that just cost too much once already (then it stays).
   * Not judged in the first seconds (shaders compiling) or during a camera move. Set
   * localStorage 'miniv2.quality' to 'high' to keep everything.
   */
  private updateQuality(dt: number, cpuMs: number) {
    const p = this.perf
    // Ambient occlusion reads the camera's lens, which a scripted move keeps changing.
    this.gtao.enabled = p.level >= 2 && !this.tween && this.tilt() > 0.99
    p.age += dt
    if (dt <= 0 || document.hidden || this.tween || p.age < 3) return
    // Without a GPU timer, the time between frames is the only clue (stretched when covered).
    const cost = this.gpuTimer ? Math.max(cpuMs, this.gpuTimer.ms) : dt * 1000
    p.cost = p.cost * 0.9 + cost * 0.1
    p.slowFor = p.cost > 16 ? p.slowFor + dt : 0
    p.fastFor = p.cost < (p.level === 1 ? 6 : 8) ? p.fastFor + dt : 0
    if (p.slowFor > 3 && p.level > 0 && localStorage.getItem('miniv2.quality') !== 'high') {
      // A level just regained that cannot be afforded after all: settle one below it for good.
      if (p.age - p.raisedAt < 10) p.locked = true
      this.setQualityLevel(p.level - 1)
    } else if (p.fastFor > 5 && p.level < 2 && !p.locked) {
      p.raisedAt = p.age
      this.setQualityLevel(p.level + 1)
    }
  }

  private setQualityLevel(level: number) {
    const p = this.perf
    const ratio = level === 0 ? 1 : Math.min(window.devicePixelRatio || 1, 2)
    p.level = level
    p.slowFor = p.fastFor = 0
    if (this.renderer.getPixelRatio() !== ratio) {
      this.renderer.setPixelRatio(ratio)
      this.composer.setPixelRatio(ratio)
      this.fit()
    }
  }

  /** The blur strength follows the zoom: strongest close up, like a macro lens. */
  private updateTiltShift() {
    const tilt = this.tilt()
    const blur = THREE.MathUtils.clamp(5 / this.viewDistance(), 0.06, 0.4) * tilt
    // (Darkness 0 would whiten the corners; a zero offset is what turns it off.)
    this.vignette.uniforms.offset.value = 0.95 * tilt
    const w = this.renderer.domElement.width || 1
    const h = this.renderer.domElement.height || 1
    this.tiltH.uniforms.h.value = blur / w
    this.tiltV.uniforms.v.value = blur / h
  }

  private updateLabels() {
    const place = (el: HTMLDivElement, id: number | null) => {
      const c = id !== null ? this.creatures.get(id) : undefined
      if (!c) {
        el.style.display = 'none'
        return
      }
      const top = 0.62 * PERSON_SCALE * this.crowdScale * c.size + this.pinSize * 1.9 + 0.1
      const v = new THREE.Vector3(c.rx, this.terrain.heightAt(c.rx, c.ry) + top, c.ry).project(this.camera)
      if (v.z > 1) {
        el.style.display = 'none'
        return
      }
      const w = this.container.clientWidth
      const h = this.container.clientHeight
      el.style.display = 'block'
      el.style.transform = `translate(${((v.x + 1) / 2) * w}px, ${((1 - v.y) / 2) * h}px) translate(-50%, -100%)`
      el.textContent = `${c.name}`
    }
    place(this.labels.selected, this.selectedId)
    place(this.labels.hover, this.hoverId !== this.selectedId ? this.hoverId : null)
  }
}

// --- Helpers ----------------------------------------------------------------------------

const UP = new THREE.Vector3(0, 1, 0)

/** The uniforms every animated material shares (see World3D.uniforms). */
type Shared = {
  uTime: { value: number }
  uWind: { value: number }
  uWindDir: { value: THREE.Vector2 }
  uBurnt: { value: THREE.DataTexture }
  uDry: { value: number }
  uLush: { value: number }
  uWater: { value: THREE.DataTexture }
  uMapSize: { value: THREE.Vector2 }
}

/** One RGBA texel per tile, filtered so the water's edge eases from tile to tile. */
function waterTexture(w: number, h: number) {
  const tex = new THREE.DataTexture(new Uint8Array(w * h * 4).fill(255), w, h, THREE.RGBAFormat)
  tex.magFilter = THREE.LinearFilter
  tex.minFilter = THREE.LinearFilter
  tex.needsUpdate = true
  return tex
}
const FRUSTUM = new THREE.Frustum()
const MAT = new THREE.Matrix4()
const SPHERE = new THREE.Sphere()
const SPOT = new THREE.Vector3()
const IDENTITY = new THREE.Quaternion()
const EULER = new THREE.Euler()
/** Leaders seen close up this frame (filled by updateActors, drawn by the village layer). */
const LEADERS: { x: number; y: number; z: number; id: number }[] = []
/** The cinematic subject being framed (reused). */
const TARGET: Target = { x: 0, y: 0, z: 0, heading: 0, radius: 1 }
const FLASH_SKY = new THREE.Color('#dfe4ff')
const STORM_CLOUD = new THREE.Color('#4d535b')
/** Each species' legs as models.ts builds them (the fowl has two: it keeps the flat pose). */
const QUAD_BUILDS: (QuadBuild | null)[] = [
  { halfLength: 0.14, halfWidth: 0.06, legLength: 0.24 },
  { halfLength: 0.12, halfWidth: 0.06, legLength: 0.13 },
  null,
  { halfLength: 0.2, halfWidth: 0.09, legLength: 0.28 },
  { halfLength: 0.17, halfWidth: 0.06, legLength: 0.2 },
]

function instanced(geo: THREE.BufferGeometry, mat: THREE.Material, cap: number) {
  const mesh = new THREE.InstancedMesh(geo, mat, cap)
  mesh.count = 0
  mesh.castShadow = true
  mesh.receiveShadow = true
  mesh.frustumCulled = false
  mesh.setColorAt(0, new THREE.Color(1, 1, 1))
  return mesh
}

/** Land: vertex colours, the season washing over the vegetated ground, light dancing on river and sea beds. */
function terrainMaterial(u: Shared) {
  const mat = new THREE.MeshStandardMaterial({ vertexColors: true, flatShading: true, roughness: 0.95, metalness: 0 })
  mat.onBeforeCompile = (shader) => {
    shader.uniforms.uDry = u.uDry
    shader.uniforms.uLush = u.uLush
    shader.uniforms.uTime = u.uTime
    shader.uniforms.uWater = u.uWater
    shader.uniforms.uMapSize = u.uMapSize
    shader.uniforms.uBurnt = u.uBurnt
    shader.vertexShader = shader.vertexShader
      .replace(
        '#include <common>',
        '#include <common>\nattribute float vegetation;\nattribute float waterLevel;\nvarying float vVeg;\nvarying float vLevel;\nvarying vec3 vGroundPos;',
      )
      .replace(
        '#include <begin_vertex>',
        '#include <begin_vertex>\nvVeg = vegetation;\nvLevel = waterLevel;\nvGroundPos = (modelMatrix * vec4(transformed, 1.0)).xyz;',
      )
    shader.fragmentShader = shader.fragmentShader
      .replace(
        '#include <common>',
        `#include <common>
        uniform float uDry;
        uniform float uLush;
        uniform float uTime;
        uniform sampler2D uWater;
        uniform sampler2D uBurnt;
        uniform vec2 uMapSize;
        varying float vVeg;
        varying float vLevel;
        varying vec3 vGroundPos;
        // Bright, wavering lines where sunlight focused by the ripples falls on the bed.
        float caustic(vec2 p, float t) {
          float v = sin(p.x * 1.7 + sin(p.y * 2.3 + t) * 1.3) + sin(p.y * 1.9 + sin(p.x * 2.1 - t * 1.2) * 1.3);
          return 1.0 - smoothstep(0.0, 0.4, abs(v));
        }
        float grain(vec2 p) {
          vec3 q = fract(vec3(p.xyx) * 0.1031);
          q += dot(q, q.yzx + 33.33);
          return fract((q.x + q.y) * q.z);
        }
        float patches(vec2 p) {
          vec2 i = floor(p);
          vec2 f = fract(p);
          vec2 s = f * f * (3.0 - 2.0 * f);
          return mix(mix(grain(i), grain(i + vec2(1.0, 0.0)), s.x), mix(grain(i + vec2(0.0, 1.0)), grain(i + vec2(1.0, 1.0)), s.x), s.y);
        }`,
      )
      .replace(
        '#include <color_fragment>',
        `#include <color_fragment>
        diffuseColor.rgb = mix(diffuseColor.rgb, vec3(0.80, 0.66, 0.30), uDry * vVeg * 0.45);
        diffuseColor.rgb = mix(diffuseColor.rgb, vec3(0.16, 0.48, 0.20), uLush * vVeg * 0.22);
        float under = vLevel - vGroundPos.y;
        // Where the river has dried, its bed lies bare: pale cracked mud, darker
        // and damp where water still lies under the sand.
        vec4 wt = texture2D(uWater, vGroundPos.xz / uMapSize);
        float bare = (1.0 - smoothstep(0.02, 0.2, wt.r)) * smoothstep(0.0, 0.05, under);
        float damp = smoothstep(0.05, 0.15, wt.g) * (1.0 - smoothstep(0.2, 0.32, wt.g));
        float crack = smoothstep(0.82, 0.9, abs(sin(vGroundPos.x * 9.0 + sin(vGroundPos.z * 7.0) * 1.7) * sin(vGroundPos.z * 8.0 + sin(vGroundPos.x * 6.0))));
        vec3 dryBed = mix(vec3(0.58, 0.5, 0.37), vec3(0.38, 0.3, 0.2), damp) * (1.0 - 0.3 * crack * (1.0 - damp));
        diffuseColor.rgb = mix(diffuseColor.rgb, dryBed, bare);
        under *= 1.0 - bare;
        // Where fire has passed: charred black with patches of grey ash, fading as plants return
        // (R); where it burns now, embers glowing in the ground (G).
        vec4 burnt = texture2D(uBurnt, vGroundPos.xz / uMapSize);
        float n = patches(vGroundPos.xz * 2.7) * 0.6 + patches(vGroundPos.xz * 7.3 + 5.0) * 0.4;
        vec3 charred = mix(vec3(0.05, 0.045, 0.04), vec3(0.36, 0.34, 0.31), smoothstep(0.55, 0.85, n) * 0.7);
        // A ragged edge, and patches the fire spared as it fades.
        float scorch = smoothstep(0.02, 0.5, burnt.r + (n - 0.5) * 0.35) * step(under, 0.0);
        diffuseColor.rgb = mix(diffuseColor.rgb, charred, scorch * 0.9);`,
      )
      .replace(
        '#include <emissivemap_fragment>',
        `#include <emissivemap_fragment>
        if (burnt.g > 0.0) {
          float ember = 0.55 + 0.45 * sin(uTime * 7.0 + vGroundPos.x * 3.1 + vGroundPos.z * 2.3);
          totalEmissiveRadiance += vec3(1.0, 0.32, 0.06) * burnt.g * ember * 0.9;
        }
        if (under > 0.0) {
          vec2 cp = vGroundPos.xz * 2.4;
          float c = caustic(cp, uTime * 0.9) * caustic(cp * 0.7 + 3.1, -uTime * 0.7);
          totalEmissiveRadiance += vec3(0.75, 0.95, 0.9) * c * 0.22 * smoothstep(0.0, 0.04, under) * (1.0 - smoothstep(0.1, 0.9, under));
        }`,
      )
  }
  return mat
}

/**
 * Water, in the stylised way (after Ghibli-like rivers): the depth below each point colours it
 * from clear turquoise shallows, where the bed shows, to deep teal and blue; a ripple pattern
 * slides downstream with the current (two copies half a cycle apart, cross-faded so neither
 * stretches) and bends the reflections; flow lines streak the faster reaches, white water
 * breaks where it runs steep, foam rings every shore and surf rolls up the beaches.
 */
function waterMaterial(u: Shared) {
  // The sky's reflection kept modest, so the water keeps its colour even at a glancing angle.
  const mat = new THREE.MeshStandardMaterial({
    color: '#ffffff',
    roughness: 0.1,
    metalness: 0.02,
    transparent: true,
    envMapIntensity: 0.35,
  })
  mat.onBeforeCompile = (shader) => {
    shader.uniforms.uTime = u.uTime
    shader.uniforms.uWind = u.uWind
    shader.uniforms.uWindDir = u.uWindDir
    shader.uniforms.uWater = u.uWater
    shader.uniforms.uMapSize = u.uMapSize
    shader.vertexShader = shader.vertexShader
      .replace(
        '#include <common>',
        `#include <common>
        attribute float ground;
        attribute vec2 flow;
        attribute float fresh;
        uniform float uTime;
        uniform float uWind;
        uniform sampler2D uWater;
        uniform vec2 uMapSize;
        varying float vGround;
        varying vec2 vFlow;
        varying float vFresh;
        varying vec3 vWPos;
        varying float vLevel;
        varying float vRun;
        varying float vFoul;`,
      )
      .replace(
        '#include <begin_vertex>',
        `#include <begin_vertex>
        vGround = ground;
        vFlow = flow;
        vFresh = fresh;
        // The water cycle: a river that has fallen sits lower in its bed.
        vec4 wt = texture2D(uWater, (modelMatrix * vec4(transformed, 1.0)).xz / uMapSize);
        vLevel = wt.r;
        vRun = wt.g;
        vFoul = (1.0 - wt.b) * fresh;
        transformed.y -= (1.0 - vLevel) * 0.34 * fresh;
        vec4 wp = modelMatrix * vec4(transformed, 1.0);
        vWPos = wp.xyz;
        float swell = (0.02 + 0.035 * uWind) * (1.0 - 0.85 * fresh) * smoothstep(0.05, 0.9, wp.y - ground);
        transformed.y += (sin(wp.x * 1.3 + uTime * 1.4) + sin(wp.z * 1.7 - uTime * 1.1) * 0.7) * swell;`,
      )
    shader.fragmentShader = shader.fragmentShader
      .replace(
        '#include <common>',
        `#include <common>
        uniform float uTime;
        uniform float uWind;
        uniform vec2 uWindDir;
        varying float vGround;
        varying vec2 vFlow;
        varying float vFresh;
        varying vec3 vWPos;
        varying float vLevel;
        varying float vRun;
        varying float vFoul;
        float hash12(vec2 p) {
          vec3 q = fract(vec3(p.xyx) * 0.1031);
          q += dot(q, q.yzx + 33.33);
          return fract((q.x + q.y) * q.z);
        }
        float vnoise(vec2 p) {
          vec2 i = floor(p);
          vec2 f = fract(p);
          vec2 s = f * f * (3.0 - 2.0 * f);
          return mix(mix(hash12(i), hash12(i + vec2(1.0, 0.0)), s.x), mix(hash12(i + vec2(0.0, 1.0)), hash12(i + vec2(1.0, 1.0)), s.x), s.y);
        }
        float ripples(vec2 p) {
          return vnoise(p) * 0.62 + vnoise(p * 2.3 + 7.1) * 0.38;
        }
        float flowing(vec2 p, vec2 drift, float t) {
          float a = fract(t);
          float b = fract(t + 0.5);
          return mix(ripples(p - drift * a), ripples(p - drift * b + 0.37), abs(1.0 - 2.0 * a));
        }`,
      )
      .replace(
        '#include <color_fragment>',
        `#include <color_fragment>
        // A dry bed shows through; pools stand only in the hollows.
        if (vFresh > 0.5 && vLevel < 0.03) discard;
        float running = mix(1.0, smoothstep(0.35, 0.9, vRun), vFresh);
        if (vFresh > 0.5 && running < 0.5 && ripples(vWPos.xz * 0.8 + 11.0) > vLevel * 2.4 + 0.3) discard;
        float depth = max(0.0, vWPos.y - vGround);
        float deepness = 1.0 - exp(-depth * mix(2.4, 6.0, vFresh));
        // The sea drifts with the wind; rivers run downstream (pools lie still).
        vec2 drift = mix(uWindDir * 0.25 * (0.6 + uWind), vFlow * running, vFresh) * 1.6;
        vec2 rp = vWPos.xz * 1.9;
        float rt = uTime * 0.32;
        float n = flowing(rp, drift, rt);
        float speed = length(vFlow) * vFresh * running * (0.4 + 0.6 * vLevel);
        vec3 shallowCol = mix(vec3(0.34, 0.86, 0.80), vec3(0.1, 0.52, 0.56), vFresh);
        vec3 deepCol = mix(vec3(0.02, 0.19, 0.40), vec3(0.02, 0.16, 0.3), vFresh);
        diffuseColor.rgb = mix(shallowCol, deepCol, deepness);
        // Water fouled with filth turns a murky green-brown, scummy where it lies still.
        float murk = clamp(vFoul * 1.4, 0.0, 1.0);
        diffuseColor.rgb = mix(diffuseColor.rgb, mix(vec3(0.30, 0.31, 0.13), vec3(0.42, 0.40, 0.2), smoothstep(0.55, 0.8, n) * (1.0 - running)), murk * 0.8);
        float shore = 1.0 - smoothstep(0.0, 0.03 + 0.05 * n, depth);
        float lines = smoothstep(0.64, 0.72, n) * smoothstep(0.3, 0.9, speed) * 0.7;
        float rapids = smoothstep(1.3, 2.1, speed) * smoothstep(0.42, 0.62, n);
        float surf = (1.0 - vFresh) * smoothstep(0.72, 0.96, sin(depth * 15.0 - uTime * 2.1 + n * 3.0)) * (1.0 - smoothstep(0.03, 0.42, depth));
        float foam = clamp(max(max(shore, lines), max(rapids, surf)), 0.0, 1.0);
        diffuseColor.rgb = mix(diffuseColor.rgb, vec3(0.95, 0.99, 1.0), foam * 0.88);
        diffuseColor.a = max(max(mix(mix(0.3, 0.6, vFresh), 0.94, deepness), murk * 0.85), foam * 0.92) * smoothstep(0.0, 0.02, depth);`,
      )
      .replace(
        '#include <roughnessmap_fragment>',
        `#include <roughnessmap_fragment>
        roughnessFactor = mix(roughnessFactor, 0.85, foam);`,
      )
      .replace(
        '#include <normal_fragment_begin>',
        `#include <normal_fragment_begin>
        float e = 0.08;
        float nx = flowing(rp + vec2(e, 0.0), drift, rt) - n;
        float nz = flowing(rp + vec2(0.0, e), drift, rt) - n;
        float bump = 0.55 + 0.6 * speed + 0.4 * (1.0 - vFresh) * uWind;
        vec3 wn = normalize(vec3(-nx / e * 0.06 * bump, 1.0, -nz / e * 0.06 * bump));
        normal = normalize((viewMatrix * vec4(wn, 0.0)).xyz);`,
      )
  }
  return mat
}

/** Leg cycles per tile walked, by species (as on the 2D map): chickens patter, buffalo stride. */
const ANIMAL_STRIDE = [1.1, 1.5, 3.2, 0.9, 0.9]

/** Per animal: step phase (radians) and whether it is walking. */
function setGait(mesh: THREE.InstancedMesh, cap: number) {
  const old = mesh.geometry.getAttribute('gait') as THREE.InstancedBufferAttribute | undefined
  if (old && old.count >= cap) return
  mesh.geometry.setAttribute('gait', new THREE.InstancedBufferAttribute(new Float32Array(cap * 2), 2))
  // Each leg's length (× as modelled), 1 until the feet are solved.
  mesh.geometry.setAttribute('aLegs', new THREE.InstancedBufferAttribute(new Float32Array(cap * 4).fill(1), 4))
}

/** Animals' legs swing from the hip as they walk, diagonal pairs together. */
function trottingMaterial(u: Shared) {
  const mat = new THREE.MeshStandardMaterial({ vertexColors: true, flatShading: true, roughness: 0.85 })
  mat.onBeforeCompile = (shader) => {
    shader.uniforms.uTime = u.uTime
    shader.vertexShader = shader.vertexShader
      .replace(
        '#include <common>',
        '#include <common>\nattribute float leg;\nattribute vec2 hip;\nattribute vec2 gait;\nattribute vec4 aLegs;',
      )
      .replace(
        '#include <begin_vertex>',
        `#include <begin_vertex>
        if (leg != 0.0) {
          // Its foothold (QuadLegSolver): the leg stretches or shortens from the hip. Fore legs
          // stand forward (+x), the +z pair on one side; aLegs holds fore +z, fore -z, hind +z, hind -z.
          float reach = hip.x > 0.0 ? (position.z > 0.0 ? aLegs.x : aLegs.y) : (position.z > 0.0 ? aLegs.z : aLegs.w);
          if (hip.x == 0.0) reach = 1.0;
          transformed.y = hip.y + (transformed.y - hip.y) * reach;
          float swing = sin(gait.x + (leg > 0.0 ? 0.0 : 3.14159)) * 0.55 * gait.y;
          vec2 q = transformed.xy - hip;
          float cs = cos(swing);
          float sn = sin(swing);
          transformed.xy = hip + vec2(q.x * cs - q.y * sn, q.x * sn + q.y * cs);
        }`,
      )
  }
  return mat
}

/** Grass: bends easily in the wind, dries to straw in the dry season and greens up in the wet. */
function grassMaterial(u: Shared) {
  const mat = swayMaterial(u, 2.6)
  mat.side = THREE.DoubleSide
  const sway = mat.onBeforeCompile
  mat.onBeforeCompile = (shader, renderer) => {
    sway(shader, renderer)
    // Far away the tufts would only speckle the ground: they shrink into it.
    shader.vertexShader = shader.vertexShader.replace(
      '#include <project_vertex>',
      `vec3 tuftAt = (modelMatrix * instanceMatrix * vec4(0.0, 0.0, 0.0, 1.0)).xyz;
      transformed *= 1.0 - smoothstep(40.0, 85.0, distance(cameraPosition, tuftAt));
      #include <project_vertex>`,
    )
    shader.uniforms.uDry = u.uDry
    shader.uniforms.uLush = u.uLush
    shader.fragmentShader = shader.fragmentShader
      .replace('#include <common>', '#include <common>\nuniform float uDry;\nuniform float uLush;')
      .replace(
        '#include <color_fragment>',
        `#include <color_fragment>
        diffuseColor.rgb = mix(diffuseColor.rgb, diffuseColor.rgb * vec3(1.35, 1.08, 0.45), uDry * 0.75);
        diffuseColor.rgb = mix(diffuseColor.rgb, diffuseColor.rgb * vec3(0.85, 1.1, 0.8), uLush * 0.5);`,
      )
  }
  return mat
}

/**
 * Plants that bend in the wind from their foot: the higher the vertex, the more it moves, leaning
 * downwind (the stream's wind) and swaying, with gusts sweeping across the land in bands. Where fire
 * has passed they are charred, and where it burns they glow.
 */
function swayMaterial(u: Shared, stiffness: number) {
  const mat = new THREE.MeshStandardMaterial({ vertexColors: true, flatShading: true, roughness: 0.9 })
  // The stiffness is baked into the shader text, which three.js would not tell apart by itself.
  mat.customProgramCacheKey = () => `sway:${stiffness.toFixed(2)}`
  mat.onBeforeCompile = (shader) => {
    shader.uniforms.uTime = u.uTime
    shader.uniforms.uWind = u.uWind
    shader.uniforms.uWindDir = u.uWindDir
    shader.uniforms.uBurnt = u.uBurnt
    shader.uniforms.uMapSize = u.uMapSize
    shader.vertexShader = shader.vertexShader
      .replace(
        '#include <common>',
        '#include <common>\nuniform float uTime;\nuniform float uWind;\nuniform vec2 uWindDir;\nuniform sampler2D uBurnt;\nuniform vec2 uMapSize;\nvarying vec2 vBurnt;',
      )
      .replace(
        '#include <begin_vertex>',
        `#include <begin_vertex>
        #ifdef USE_INSTANCING
          vec2 base = instanceMatrix[3].xz;
          vBurnt = texture2D(uBurnt, base / uMapSize).rg;
        #else
          vec2 base = vec2(0.0);
          vBurnt = vec2(0.0);
        #endif
        // Gusts run downwind across the land in broad bands.
        float band = sin(dot(base, uWindDir) * 0.18 - uTime * (0.9 + 1.4 * uWind));
        float gust = 1.0 + 0.7 * uWind * smoothstep(0.2, 1.0, band);
        float lean = (0.03 + 0.09 * uWind) * ${stiffness.toFixed(2)} * gust;
        float sway = (0.55 + 0.45 * sin(uTime * (1.3 + uWind) + base.x * 0.7 + base.y * 1.3)) * lean * max(0.0, transformed.y);
        // A little across the wind too, so nothing moves like a metronome.
        float across = sin(uTime * (2.1 + uWind) + base.y * 0.9) * 0.18 * lean * max(0.0, transformed.y);
        transformed.x += sway * uWindDir.x - across * uWindDir.y;
        transformed.z += sway * uWindDir.y + across * uWindDir.x;`,
      )
    shader.fragmentShader = shader.fragmentShader
      .replace('#include <common>', '#include <common>\nuniform float uTime;\nvarying vec2 vBurnt;')
      .replace(
        '#include <color_fragment>',
        `#include <color_fragment>
        diffuseColor.rgb = mix(diffuseColor.rgb, vec3(0.07, 0.06, 0.05), smoothstep(0.02, 0.55, vBurnt.r) * 0.88);`,
      )
      .replace(
        '#include <emissivemap_fragment>',
        `#include <emissivemap_fragment>
        totalEmissiveRadiance += vec3(1.0, 0.36, 0.07) * vBurnt.g * (0.5 + 0.5 * sin(uTime * 9.0 + vViewPosition.x * 4.0)) * 0.8;`,
      )
  }
  return mat
}

/** A sky to bake into the environment map: the dome's gradient, a bright sun and a glow around it. */
function makeEnvSky() {
  const scene = new THREE.Scene()
  const mat = new THREE.ShaderMaterial({
    side: THREE.BackSide,
    depthWrite: false,
    uniforms: {
      uHorizon: { value: new THREE.Color('#d6ecf7') },
      uZenith: { value: new THREE.Color('#4f9be0') },
      uSunDir: { value: SUN_DIR.clone() },
      uSun: { value: 1 },
    },
    vertexShader:
      'varying vec3 vDir; void main(){ vDir = normalize(position); gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0); }',
    fragmentShader: `uniform vec3 uHorizon; uniform vec3 uZenith; uniform vec3 uSunDir; uniform float uSun; varying vec3 vDir;
      void main(){
        vec3 sky = mix(uHorizon, uZenith, smoothstep(-0.05, 0.55, vDir.y));
        sky = mix(sky * 0.55, sky, smoothstep(-0.3, 0.0, vDir.y));
        float d = max(dot(normalize(vDir), uSunDir), 0.0);
        sky += vec3(1.0, 0.92, 0.75) * (pow(d, 900.0) * 40.0 + pow(d, 24.0) * 0.6) * uSun;
        gl_FragColor = vec4(sky, 1.0);
      }`,
  })
  scene.add(new THREE.Mesh(new THREE.SphereGeometry(10, 32, 16), mat))
  return scene
}

function makeSkyDome() {
  const g = new THREE.SphereGeometry(1500, 24, 12)
  const m = new THREE.ShaderMaterial({
    side: THREE.BackSide,
    depthWrite: false,
    fog: false,
    uniforms: { uHorizon: { value: new THREE.Color('#d6ecf7') }, uZenith: { value: new THREE.Color('#4f9be0') } },
    vertexShader:
      'varying vec3 vDir; void main(){ vDir = normalize(position); gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0); }',
    fragmentShader:
      'uniform vec3 uHorizon; uniform vec3 uZenith; varying vec3 vDir; void main(){ float t = smoothstep(-0.05, 0.55, vDir.y); gl_FragColor = vec4(mix(uHorizon, uZenith, t), 1.0); }',
  })
  const mesh = new THREE.Mesh(g, m)
  mesh.renderOrder = -1
  mesh.onBeforeRender = (_r, _s, camera) => mesh.position.copy(camera.position)
  return mesh
}

function makeRain() {
  const n = 1400
  const g = new THREE.BufferGeometry()
  g.setAttribute('position', new THREE.BufferAttribute(new Float32Array(n * 6).fill(-9999), 3))
  const m = new THREE.LineBasicMaterial({ color: '#dbe8f5', transparent: true, opacity: 0.4 })
  const lines = new THREE.LineSegments(g, m)
  lines.frustumCulled = false
  lines.visible = false
  return lines
}

/** A lightning bolt: a jagged line, redrawn for each strike, seen only while it flashes. */
function makeBolt() {
  const g = new THREE.BufferGeometry()
  g.setAttribute('position', new THREE.BufferAttribute(new Float32Array(18 * 2 * 3), 3))
  const m = new THREE.LineBasicMaterial({ color: '#eef1ff', transparent: true, opacity: 0, fog: false, depthWrite: false })
  const lines = new THREE.LineSegments(g, m)
  lines.frustumCulled = false
  lines.visible = false
  return lines
}

function makeSmoke() {
  const max = 6 * 600
  const g = new THREE.BufferGeometry()
  g.setAttribute('position', new THREE.BufferAttribute(new Float32Array(max * 3), 3))
  g.setAttribute('alpha', new THREE.BufferAttribute(new Float32Array(max), 1))
  g.setAttribute('size', new THREE.BufferAttribute(new Float32Array(max), 1))
  g.setDrawRange(0, 0)
  const m = new THREE.ShaderMaterial({
    transparent: true,
    depthWrite: false,
    uniforms: { uScale: { value: 400 }, uLight: { value: 1 } },
    vertexShader:
      'attribute float alpha; attribute float size; uniform float uScale; varying float vAlpha; void main(){ vAlpha = alpha; vec4 mv = modelViewMatrix * vec4(position, 1.0); gl_PointSize = size * uScale / -mv.z; gl_Position = projectionMatrix * mv; }',
    fragmentShader:
      'uniform float uLight; varying float vAlpha; void main(){ float d = length(gl_PointCoord - 0.5); float a = smoothstep(0.5, 0.15, d) * vAlpha; gl_FragColor = vec4(vec3(0.93, 0.92, 0.9) * uLight, a); }',
  })
  const pts = new THREE.Points(g, m)
  pts.frustumCulled = false
  return pts
}

/** Mosquitoes: tiny dark specks, each wheeling round its swarm's centre (animated on the GPU). */
function makeMosquitoes() {
  const g = new THREE.BufferGeometry()
  g.setAttribute('position', new THREE.BufferAttribute(new Float32Array(0), 3))
  g.setAttribute('seed', new THREE.BufferAttribute(new Float32Array(0), 3))
  const m = new THREE.ShaderMaterial({
    transparent: true,
    depthWrite: false,
    uniforms: { uTime: { value: 0 }, uActive: { value: 0.5 }, uScale: { value: 400 } },
    vertexShader: `
      attribute vec3 seed;
      uniform float uTime;
      uniform float uActive;
      uniform float uScale;
      varying float vAlpha;
      void main() {
        vec3 p = position;
        p.x += sin(uTime * (5.0 + seed.x * 4.0) + seed.x * 40.0) * (0.12 + 0.22 * seed.y);
        p.z += cos(uTime * (4.0 + seed.y * 5.0) + seed.y * 40.0) * (0.08 + 0.18 * seed.x);
        p.y += sin(uTime * (6.0 + seed.x * 3.0) + seed.y * 20.0) * 0.07;
        vec4 mv = modelViewMatrix * vec4(p, 1.0);
        // Only the active share flies; far off they are too small to see.
        vAlpha = step(seed.z, uActive) * smoothstep(42.0, 14.0, -mv.z);
        gl_PointSize = max(1.8, 0.08 * uScale / -mv.z) * step(0.01, vAlpha);
        gl_Position = projectionMatrix * mv;
      }`,
    fragmentShader: `
      varying float vAlpha;
      void main() {
        float d = length(gl_PointCoord - 0.5);
        gl_FragColor = vec4(0.06, 0.06, 0.07, smoothstep(0.5, 0.2, d) * vAlpha * 0.85);
      }`,
  })
  const pts = new THREE.Points(g, m)
  pts.frustumCulled = false
  return pts
}

// Shared geometries and materials for buildings and fields.
const BODY_MAT = new THREE.MeshStandardMaterial({ vertexColors: true, flatShading: true, roughness: 0.85 })
const RUIN_MAT = new THREE.MeshStandardMaterial({ color: '#6a6560', flatShading: true, roughness: 1 })
const STILL_MAT = new THREE.MeshStandardMaterial({ vertexColors: true, flatShading: true, roughness: 0.95 })
const roofMats = new Map<number, THREE.MeshStandardMaterial>()
function roofMaterial(hue: number) {
  const key = Math.round(hue / 10) * 10
  let m = roofMats.get(key)
  if (!m) {
    m = new THREE.MeshStandardMaterial({ color: new THREE.Color().setHSL(key / 360, 0.38, 0.42), flatShading: true, roughness: 0.8 })
    roofMats.set(key, m)
  }
  return m
}
const buildingCache = new Map<string, ReturnType<typeof buildingModel>>()
function cachedBuilding(kind: string, level: number) {
  const key = `${kind}|${level}`
  let m = buildingCache.get(key)
  if (!m) {
    m = buildingModel(kind, level)
    buildingCache.set(key, m)
  }
  return m
}
const FIELD_GEO = new THREE.BoxGeometry(0.96, 0.04, 0.96)
const FIELD_MAT = new THREE.MeshStandardMaterial({ color: '#6b4a2e', flatShading: true, roughness: 1 })
/** Neighbours in the 2D map's link order: N, E, S, W. */
const SIDES = [
  [0, -1],
  [1, 0],
  [0, 1],
  [-1, 0],
]
const SOIL_GEO = new THREE.BoxGeometry(0.9, 0.03, 0.9)
const GRASS_GEO = grassTuftGeometry()
const REED_GEO = reedGeometry()
const BRIDGE_MAT = new THREE.MeshStandardMaterial({ color: '#8a5a32', roughness: 0.9, flatShading: true })
const SOIL_MAT = new THREE.MeshStandardMaterial({ color: '#ffffff', flatShading: true, roughness: 1 })
const PLANT_GEO: Record<string, () => THREE.BufferGeometry> = {
  tree: treeGeometry,
  pine: pineGeometry,
  bush: bushGeometry,
  rock: rockGeometry,
  flowers: flowerGeometry,
  stump: stumpGeometry,
  wall: wallGeometry,
}
const CROP_GEO: Record<string, () => THREE.BufferGeometry> = {
  rice: riceGeometry,
  leafy: leafyGeometry,
  banana: () => palmGeometry(0.6),
  coconut: () => palmGeometry(1.4),
  sago: () => palmGeometry(1),
}
