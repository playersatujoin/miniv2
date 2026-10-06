// Procedural animals, drawn like the creatures: canvas primitives, feet at the
// given world pixel, scaled by species and age, mirrored to face their heading.

import { ANIMAL_FLAG, SPECIES } from '../sim/protocol'

type Ctx = CanvasRenderingContext2D

export type AnimalLook = {
  /** Index into SPECIES. */
  species: number
  heading: number
  flags: number
  step: number
  moving: boolean
  /** Stable per-animal number (antlers, coat shade). */
  variant: number
  time: number
}

/** Draw scale of an animal: young ones are smaller. Body size is baked into each painter. */
export function animalScale(flags: number) {
  return flags & ANIMAL_FLAG.young ? 0.6 : 1
}

/** Height of an animal's art in world pixels at scale 1, for labels and hit tests. */
export const ANIMAL_HEIGHT = [36, 18, 14, 27, 22]

/** "Rusa", "Anak harimau", "Ayam (jinak)". */
export function animalLabel(species: number, flags: number) {
  const sp = SPECIES[species]
  if (!sp) return 'Hewan'
  const tame = (flags & ANIMAL_FLAG.tame) !== 0
  const name = tame && sp.tame ? sp.tame : sp.name
  const young = flags & ANIMAL_FLAG.young ? `Anak ${name.toLowerCase()}` : name
  return tame ? `${young} (jinak)` : young
}

function ellipse(ctx: Ctx, x: number, y: number, rx: number, ry: number, color: string, rot = 0) {
  ctx.fillStyle = color
  ctx.beginPath()
  ctx.ellipse(x, y, rx, ry, rot, 0, Math.PI * 2)
  ctx.fill()
}

function circle(ctx: Ctx, x: number, y: number, r: number, color: string) {
  ctx.fillStyle = color
  ctx.beginPath()
  ctx.arc(x, y, r, 0, Math.PI * 2)
  ctx.fill()
}

/** Legs of a quadruped seen from the side; the animal faces +x. */
type Pose = {
  /** Stride phase, −1..1 (0 = standing). */
  swing: number
  /** Body lift while trotting. */
  bob: number
  /** Head lowered to graze. */
  grazing: boolean
  running: boolean
}

/** Four legs from hip (hy) to ground (0); the far pair darker and out of step. */
function legs(ctx: Ctx, xs: [back: number, front: number], hy: number, w: number, color: string, far: string, p: Pose) {
  const reach = p.running ? 3.2 : 2
  const leg = (x: number, phase: number, c: string) => {
    const dx = phase * reach
    ctx.strokeStyle = c
    ctx.lineWidth = w
    ctx.beginPath()
    ctx.moveTo(x, hy)
    ctx.lineTo(x + dx, -0.5)
    ctx.stroke()
  }
  ctx.lineCap = 'round'
  leg(xs[0] + 1.5, -p.swing, far)
  leg(xs[1] + 1.5, p.swing, far)
  leg(xs[0], p.swing, color)
  leg(xs[1], -p.swing, color)
  ctx.lineCap = 'butt'
}

/** A collar and a little bell: household livestock. */
function collar(ctx: Ctx, x: number, y: number, r: number) {
  ctx.strokeStyle = '#c0392b'
  ctx.lineWidth = 1.4
  ctx.beginPath()
  ctx.arc(x, y, r, Math.PI * 0.2, Math.PI * 0.85)
  ctx.stroke()
  circle(ctx, x - r * 0.15, y + r * 0.95, 1.1, '#ffd166')
}

function paintRusa(ctx: Ctx, p: Pose, v: number, tame: boolean) {
  const coat = v % 3 === 0 ? '#9a6a3c' : '#8a5e35'
  const dark = '#5e3e22'
  const by = -14 - p.bob
  legs(ctx, [-6.5, 6], by + 2, 1.8, dark, '#4a301a', p)
  ellipse(ctx, 0, by, 10, 5, coat)
  ellipse(ctx, 0, by + 2.5, 7, 2, '#c7a27a') // pale belly
  ellipse(ctx, -9.5, by - 2, 1.8, 2.4, '#f2e6d0') // white tail
  // Neck and head, lowered while grazing.
  const hx = p.grazing ? 12 : 10
  const hy = p.grazing ? by + 6 : by - 9
  ctx.strokeStyle = coat
  ctx.lineWidth = 4
  ctx.lineCap = 'round'
  ctx.beginPath()
  ctx.moveTo(6, by - 1)
  ctx.lineTo(hx - 1, hy + 1)
  ctx.stroke()
  ctx.lineCap = 'butt'
  ellipse(ctx, hx + 1.5, hy, 3.6, 2.4, coat, p.grazing ? 0.5 : 0.25)
  circle(ctx, hx + 4.6, hy + (p.grazing ? 1.6 : 0.8), 0.9, '#2a1a0e')
  ellipse(ctx, hx - 1, hy - 2.6, 1, 2, dark, -0.5) // ear
  circle(ctx, hx + 1.6, hy - 0.6, 0.7, '#111')
  // Stags (half of the adults) carry antlers.
  if (v % 2 === 0) {
    ctx.strokeStyle = '#d8c7a6'
    ctx.lineWidth = 1
    ctx.beginPath()
    const ax = hx + 0.5
    const ay = hy - 2
    ctx.moveTo(ax, ay)
    ctx.lineTo(ax - 2, ay - 7)
    ctx.lineTo(ax - 4.5, ay - 10)
    ctx.moveTo(ax - 1.4, ay - 5)
    ctx.lineTo(ax + 1.5, ay - 8)
    ctx.moveTo(ax - 2.8, ay - 8.2)
    ctx.lineTo(ax - 1.2, ay - 11)
    ctx.stroke()
  }
  if (tame) collar(ctx, 7.5, by - 3, 2.4)
}

function paintBabi(ctx: Ctx, p: Pose, v: number, tame: boolean) {
  // Wild boar are bristly grey-brown; village pigs pinker and smoother.
  const coat = tame ? (v % 2 ? '#c99a8a' : '#4e443e') : v % 3 === 0 ? '#5a4a3e' : '#4a3d33'
  const dark = tame ? '#8a675c' : '#2e2520'
  const by = -8 - p.bob
  legs(ctx, [-5.5, 5], by + 2, 2, dark, '#241c18', p)
  ellipse(ctx, 0, by, 9, 5.5, coat)
  // Bristly mane along the back.
  if (!tame) {
    ctx.strokeStyle = dark
    ctx.lineWidth = 0.8
    ctx.beginPath()
    for (let x = -5; x <= 5; x += 2) {
      ctx.moveTo(x, by - 5)
      ctx.lineTo(x - 0.8, by - 7)
    }
    ctx.stroke()
  }
  const hx = p.grazing ? 9.5 : 9
  const hy = p.grazing ? by + 2.5 : by - 0.5
  ellipse(ctx, hx, hy, 4, 3.4, coat, 0.2)
  ellipse(ctx, hx + 4, hy + 1, 1.4, 1.7, tame ? '#e8b4a6' : '#6e5a4c') // snout
  circle(ctx, hx + 4.4, hy + 0.8, 0.45, '#111')
  ellipse(ctx, hx - 1.5, hy - 3, 1.2, 1.8, dark, -0.4)
  circle(ctx, hx + 1.4, hy - 1, 0.6, '#111')
  if (!tame) {
    // Tusks.
    ctx.strokeStyle = '#efe6d2'
    ctx.lineWidth = 0.8
    ctx.beginPath()
    ctx.moveTo(hx + 2.8, hy + 1.8)
    ctx.quadraticCurveTo(hx + 3.8, hy + 0.4, hx + 3.2, hy - 0.6)
    ctx.stroke()
  }
  // Curly tail.
  ctx.strokeStyle = dark
  ctx.lineWidth = 0.9
  ctx.beginPath()
  ctx.arc(-9.6, by - 1.5, 1.4, 0, Math.PI * 1.6)
  ctx.stroke()
  if (tame) collar(ctx, 6, by - 1, 3.2)
}

function paintAyam(ctx: Ctx, p: Pose, v: number, tame: boolean, time: number) {
  // Red junglefowl: a rooster's orange hackles and green-black sickle tail; hens are plain brown.
  const rooster = v % 2 === 0
  const body = rooster ? '#b8441f' : tame ? '#b58a55' : '#8a5a32'
  const by = -6 - p.bob * 0.6
  // Two thin legs.
  ctx.strokeStyle = '#d9a24a'
  ctx.lineWidth = 0.9
  ctx.beginPath()
  ctx.moveTo(-0.5, by + 2)
  ctx.lineTo(-0.5 - p.swing * 1.5, 0)
  ctx.moveTo(1.5, by + 2)
  ctx.lineTo(1.5 + p.swing * 1.5, 0)
  ctx.stroke()
  // Tail.
  if (rooster) {
    ctx.strokeStyle = '#1f3d2a'
    ctx.lineWidth = 1.6
    ctx.beginPath()
    ctx.moveTo(-3, by - 1)
    ctx.quadraticCurveTo(-7, by - 9, -8.5, by - 3)
    ctx.moveTo(-3, by)
    ctx.quadraticCurveTo(-6, by - 6, -7.5, by - 1)
    ctx.stroke()
  } else {
    ellipse(ctx, -4, by - 2, 2, 3, '#6e4526', -0.6)
  }
  ellipse(ctx, 0, by, 4.4, 3.2, body)
  ellipse(ctx, -0.5, by + 0.6, 2.6, 1.6, rooster ? '#2a3d2e' : '#6e4526') // wing
  // Head pecking up and down while eating.
  const peck = p.grazing ? Math.max(0, Math.sin(time * 9 + v)) * 3 : 0
  const hx = 3.6
  const hy = by - 3.6 + peck
  if (rooster) ellipse(ctx, hx - 0.6, by - 1.6, 1.8, 2.6, '#e08a2c', 0.3) // hackles
  circle(ctx, hx, hy, 1.9, body)
  ctx.fillStyle = '#e63946'
  ctx.beginPath()
  ctx.moveTo(hx - 1.3, hy - 1.3)
  ctx.lineTo(hx - 0.4, hy - (rooster ? 3.6 : 2.6))
  ctx.lineTo(hx + 0.6, hy - (rooster ? 2.4 : 2))
  ctx.lineTo(hx + 1.4, hy - (rooster ? 3.2 : 2.2))
  ctx.lineTo(hx + 1.6, hy - 1)
  ctx.closePath()
  ctx.fill()
  ellipse(ctx, hx + 0.9, hy + 1.8, 0.6, 1, '#e63946') // wattle
  ctx.fillStyle = '#e8c15a'
  ctx.beginPath()
  ctx.moveTo(hx + 1.6, hy - 0.4)
  ctx.lineTo(hx + 3.2, hy + 0.2)
  ctx.lineTo(hx + 1.6, hy + 0.6)
  ctx.fill()
  circle(ctx, hx + 0.6, hy - 0.4, 0.45, '#111')
  if (tame) circle(ctx, hx - 1, hy + 2, 0.9, '#ffd166') // a little bell
}

function paintKerbau(ctx: Ctx, p: Pose, v: number, tame: boolean) {
  const coat = v % 3 === 0 ? '#4a4c52' : '#3d3f45'
  const dark = '#26272b'
  const by = -14 - p.bob * 0.7
  legs(ctx, [-9, 8], by + 3, 2.8, dark, '#1b1c1f', p)
  ellipse(ctx, 0, by, 13.5, 7, coat)
  ellipse(ctx, -1, by - 3.5, 9, 2.6, '#5a5c63') // light along the back
  // Thin tail with a tuft.
  ctx.strokeStyle = dark
  ctx.lineWidth = 1
  ctx.beginPath()
  ctx.moveTo(-13, by - 2)
  ctx.quadraticCurveTo(-15, by + 3, -14, by + 7)
  ctx.stroke()
  ellipse(ctx, -14, by + 7.5, 1, 1.6, dark)
  const hx = p.grazing ? 14 : 13
  const hy = p.grazing ? by + 6 : by - 1
  ellipse(ctx, hx, hy, 5, 4, coat, 0.35)
  ellipse(ctx, hx + 4, hy + 2.2, 2.2, 1.8, '#6a6c73') // muzzle
  circle(ctx, hx + 1.2, hy - 1.2, 0.7, '#111')
  // Long horns sweeping back in a crescent.
  ctx.strokeStyle = '#cfc6b4'
  ctx.lineWidth = 2
  ctx.lineCap = 'round'
  ctx.beginPath()
  ctx.moveTo(hx - 1, hy - 3)
  ctx.quadraticCurveTo(hx - 9, hy - 6, hx - 7.5, hy - 1)
  ctx.stroke()
  ctx.strokeStyle = '#a99f8c'
  ctx.lineWidth = 1.4
  ctx.beginPath()
  ctx.moveTo(hx + 0.5, hy - 3.5)
  ctx.quadraticCurveTo(hx - 5, hy - 9, hx - 4.5, hy - 4)
  ctx.stroke()
  ctx.lineCap = 'butt'
  if (tame) collar(ctx, hx - 4, hy + 0.5, 3.4)
}

function paintHarimau(ctx: Ctx, p: Pose, v: number, hunting: boolean) {
  const coat = v % 2 ? '#d9822b' : '#cf7424'
  const stripe = '#2a1608'
  // Stalking: belly low to the ground.
  const by = (hunting && !p.running ? -8 : -11) - p.bob
  legs(ctx, [-8, 7.5], by + 2, 2.4, coat, '#9a5418', p)
  // Long tail, curling up at the tip.
  ctx.strokeStyle = coat
  ctx.lineWidth = 2
  ctx.lineCap = 'round'
  ctx.beginPath()
  ctx.moveTo(-11, by - 1)
  ctx.quadraticCurveTo(-17, by + 1, -18, by - 5 + Math.sin(p.swing * 2) * 1.5)
  ctx.stroke()
  ctx.lineCap = 'butt'
  ellipse(ctx, 0, by, 12, 5, coat)
  ellipse(ctx, 0.5, by + 3, 8.5, 1.8, '#f3e2c4') // white belly
  // Stripes across the back and flank.
  ctx.strokeStyle = stripe
  ctx.lineWidth = 1.1
  ctx.beginPath()
  for (let x = -8; x <= 7; x += 3) {
    ctx.moveTo(x, by - 4.6)
    ctx.quadraticCurveTo(x + 1.2, by - 1, x - 0.4, by + 1.6)
  }
  ctx.moveTo(-15, by - 0.4)
  ctx.lineTo(-14, by - 2)
  ctx.moveTo(-17.4, by - 3)
  ctx.lineTo(-16, by - 4)
  ctx.stroke()
  const hx = 12.5
  const hy = hunting ? by - 0.5 : p.grazing ? by + 3 : by - 3
  ellipse(ctx, hx, hy, 4.6, 4, coat)
  ellipse(ctx, hx + 2.8, hy + 1.6, 2.2, 1.6, '#f3e2c4') // muzzle
  circle(ctx, hx + 4.4, hy + 0.9, 0.6, '#2a1608')
  circle(ctx, hx - 2.2, hy - 3.4, 1.4, coat) // ears
  circle(ctx, hx - 2.2, hy - 3.4, 0.6, stripe)
  circle(ctx, hx + 1.4, hy - 1, 0.75, hunting ? '#ffd166' : '#111')
  ctx.strokeStyle = stripe
  ctx.lineWidth = 0.8
  ctx.beginPath()
  ctx.moveTo(hx - 1, hy - 3.6)
  ctx.lineTo(hx - 0.2, hy - 1.8)
  ctx.moveTo(hx - 3.4, hy - 1)
  ctx.lineTo(hx - 1.8, hy)
  ctx.stroke()
}

/** Draws an animal with feet at world pixel (px, py). */
export function drawAnimal(ctx: Ctx, px: number, py: number, a: AnimalLook) {
  const s = animalScale(a.flags)
  const running = (a.flags & ANIMAL_FLAG.running) !== 0
  const tame = (a.flags & ANIMAL_FLAG.tame) !== 0
  const hurt = (a.flags & ANIMAL_FLAG.hurt) !== 0
  const hunting = (a.flags & ANIMAL_FLAG.hunting) !== 0
  const fx = Math.cos(a.heading)
  // Mirror to face left or right; squeeze a little when walking towards or away from the viewer.
  const dir = fx < 0 ? -1 : 1
  const squeeze = 0.62 + 0.38 * Math.abs(fx)
  const cycle = Math.sin(a.step * Math.PI * 2)
  const pose: Pose = {
    swing: a.moving ? cycle : 0,
    bob: a.moving ? Math.abs(cycle) * (running ? 1.6 : 0.6) : 0,
    grazing: (a.flags & ANIMAL_FLAG.eating) !== 0 && !a.moving,
    running,
  }

  ctx.save()
  ctx.translate(px, py)
  ctx.scale(s, s)
  if (hurt) ellipse(ctx, 0, 0, 14, 4.5, `rgba(230,57,70,${0.3 + 0.2 * Math.sin(a.time * 6)})`)
  const w = [11, 10, 5, 15, 13][a.species] ?? 10
  ellipse(ctx, 0, 0, w, w * 0.3, 'rgba(0,0,0,0.26)')
  ctx.scale(dir * squeeze, 1)
  switch (a.species) {
    case 0:
      paintRusa(ctx, pose, a.variant, tame)
      break
    case 1:
      paintBabi(ctx, pose, a.variant, tame)
      break
    case 2:
      paintAyam(ctx, pose, a.variant, tame, a.time)
      break
    case 3:
      paintKerbau(ctx, pose, a.variant, tame)
      break
    case 4:
      paintHarimau(ctx, pose, a.variant, hunting)
      break
    default:
      ellipse(ctx, 0, -6, 6, 4, '#8a7a6a')
  }
  ctx.restore()

  // Dust kicked up by a running animal.
  if (running && a.moving) {
    for (let i = 0; i < 3; i++) {
      const ph = (a.time * 3 + i / 3) % 1
      ctx.globalAlpha = 0.35 * (1 - ph)
      circle(ctx, px - dir * (6 + ph * 10) * s, py - 1 - ph * 3, (1.2 + ph * 2) * s, '#c9b48a')
    }
    ctx.globalAlpha = 1
  }
}
