import { useEffect, useRef } from 'react'
import * as THREE from 'three'
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js'
import { Crowd } from './people'
import { bodyProportions } from './anatomy'
import { FLAG, MOOD, type CreatureDetail } from '../sim/protocol'
import type { LiveCreature } from '../sim/live'

/** The flags the portrait can show from a person's details: what they do with others, wading or climbing. */
function portraitFlags(c: CreatureDetail) {
  let flags = (!c.adult ? FLAG.child : 0) | (c.pregnant ? FLAG.pregnant : 0) | (c.body?.ill ? FLAG.ill : 0) | (c.leader ? FLAG.leader : 0)
  const chat = c.exchange?.chat
  if (chat && chat.ended === undefined) flags |= chat.kind === 'trade' ? FLAG.trading : FLAG.talking
  // Swimming and rafting need the world's water; the portrait shows them standing.
  if (c.locomotion === 'wading') flags |= FLAG.wading
  else if (c.locomotion === 'climbing') flags |= FLAG.climbing
  return flags
}

/** An observational portrait of the selected inhabitant, using the same mesh,
 * colours, age, skeleton and face as the world: their mood shows on it. It never issues simulation commands. */
export default function CharacterPortrait({ person }: { person: CreatureDetail }) {
  const ref = useRef<HTMLDivElement>(null)
  const latest = useRef(person)
  latest.current = person
  useEffect(() => {
    const host = ref.current!
    host.replaceChildren()
    let renderer: THREE.WebGLRenderer
    try {
      renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true })
    } catch {
      host.textContent = 'Tampilan 3D tidak tersedia.'
      return
    }
    renderer.setPixelRatio(Math.min(1.5, window.devicePixelRatio))
    renderer.toneMapping = THREE.ACESFilmicToneMapping
    renderer.toneMappingExposure = 1.1
    host.appendChild(renderer.domElement)
    const scene = new THREE.Scene()
    const specimen = latest.current
    const framing = specimen.traits.size * bodyProportions(specimen.age / specimen.secondsPerYear, 0.5, specimen.sex === 'female').height
    const camera = new THREE.PerspectiveCamera(30, 1, 0.01, 10)
    camera.position.set(0.46, 0.39, 1.05).multiplyScalar(framing)
    const controls = new OrbitControls(camera, renderer.domElement)
    controls.target.set(0, 0.26 * framing, 0)
    controls.enablePan = false
    // Zooming goes where the pointer is, so the face (and its mood) can be looked at closely.
    controls.zoomToCursor = true
    controls.minDistance = 0.45 * framing
    controls.maxDistance = 2 * framing
    controls.update()
    scene.add(new THREE.HemisphereLight('#dce8ff', '#6a5546', 2))
    const key = new THREE.DirectionalLight('#ffedcf', 3)
    key.position.set(1, 2, 2)
    scene.add(key)
    const rim = new THREE.DirectionalLight('#8bc5d8', 2)
    rim.position.set(-1, 1, -1)
    scene.add(rim)
    const crowd = new Crowd()
    scene.add(crowd.mesh)
    const floor = new THREE.Mesh(new THREE.CircleGeometry(0.21, 48), new THREE.MeshStandardMaterial({ color: '#445b5e', roughness: 0.9 }))
    floor.rotation.x = -Math.PI / 2
    floor.position.y = -0.006
    scene.add(floor)
    const resize = new ResizeObserver(() => {
      const w = host.clientWidth,
        h = host.clientHeight
      if (!w || !h) return
      renderer.setSize(w, h)
      camera.aspect = w / h
      camera.updateProjectionMatrix()
    })
    resize.observe(host)
    const reduced = window.matchMedia?.('(prefers-reduced-motion: reduce)')
    let frame = 0,
      previous = performance.now()
    const start = previous
    const loop = (now: number) => {
      frame = requestAnimationFrame(loop)
      const dt = Math.min(0.05, (now - previous) / 1000)
      previous = now
      const c = latest.current
      const flags = portraitFlags(c)
      const dominant = c.mood?.dominant || ''
      const cr: LiveCreature = {
        id: c.id,
        name: c.name,
        sex: c.sex,
        hue: c.traits.hue,
        size: c.traits.size,
        age: c.age / c.secondsPerYear,
        flags,
        energy: c.energy,
        health: c.health,
        houseId: c.house?.id ?? 0,
        x: 0,
        y: 0,
        px: 0,
        py: 0,
        ph: Math.PI / 2,
        heading: Math.PI / 2,
        rx: 0,
        ry: 0,
        rh: Math.PI / 2,
        walked: 0,
        moving: false,
        mood: dominant ? MOOD[dominant] : MOOD.calm,
        moodStrength: dominant ? (c.mood?.[dominant] ?? 0.5) : 0,
      }
      // Breathing, blinking and glancing about, unless the observer asked for reduced motion.
      const still = reduced?.matches ?? false
      crowd.update([cr], 1, { ground: () => 0, scale: 1, time: (now - start) / 1000, dt, still, contact: true })
      controls.update()
      renderer.render(scene, camera)
    }
    frame = requestAnimationFrame(loop)
    return () => {
      cancelAnimationFrame(frame)
      resize.disconnect()
      controls.dispose()
      crowd.dispose()
      floor.geometry.dispose()
      floor.material.dispose()
      renderer.dispose()
      renderer.domElement.remove()
    }
  }, [person.id])
  return (
    <div
      className="obs-character-portrait"
      ref={ref}
      role="img"
      aria-label={`Karakter 3D ${person.name}. Seret untuk memutar, gulir untuk memperbesar.`}
    />
  )
}
