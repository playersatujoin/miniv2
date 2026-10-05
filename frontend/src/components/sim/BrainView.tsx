import { useEffect, useMemo, useRef, useState, type PointerEvent } from 'react'
import type { Brain } from '../../sim/protocol'
import { useElementWidth } from './useElementWidth'

const C = {
  text: '#e6edf7',
  muted: '#8ea0bb',
  dim: '#62759a',
  ring: 'rgba(255,255,255,0.18)',
  neutral: [61, 64, 72] as const, // diverging midpoint
  pos: [237, 161, 0] as const, // amber: positive / excitatory
  neg: [57, 135, 229] as const, // blue: negative / inhibitory
  learned: [27, 175, 122] as const, // aqua ring: changed by lifetime learning (not a weight sign)
}
const FONT = '10px Inter, system-ui, sans-serif'
const FONT_BOLD = '600 10px Inter, system-ui, sans-serif'
const FONT_HEAD = '600 9px Inter, system-ui, sans-serif'

const TOP = 18 // column titles
const GROUP_GAP = 14 // room for a group heading
const ROW = 10
const MAX_CELL = 8 // recurrent heatmap cell, shrunk to fit wide memories
const MIN_CELL = 3
const MAX_IN_EDGES = 60
const MAX_OUT_EDGES = 30

type Kind = 'in' | 'hid' | 'out'
type Node = { kind: Kind; i: number; x: number; y: number; r: number }
type Layout = {
  width: number
  height: number
  xIn: number
  xHid: number
  xOut: number
  inputs: Node[]
  hidden: Node[]
  outputs: Node[]
  groups: { title: string; y: number }[]
  short: string[]
  heat: { x: number; y: number; n: number; cell: number }
}

const clamp = (v: number, lo: number, hi: number) => Math.min(hi, Math.max(lo, v))

/** Blue (negative) ↔ neutral grey ↔ amber (positive). */
function diverging(v: number, alpha = 1) {
  const t = clamp(Math.abs(v), 0, 1)
  const pole = v >= 0 ? C.pos : C.neg
  const [r, g, b] = C.neutral.map((n, k) => Math.round(n + (pole[k] - n) * t))
  return `rgba(${r},${g},${b},${alpha})`
}

/** "makanan -30°" → group "makanan", short "-30°"; everything else is "kondisi diri". */
function groupInputs(labels: string[]) {
  const groups: { title: string; indices: number[] }[] = []
  const short: string[] = []
  labels.forEach((label, i) => {
    const m = /^(.*\S)\s+(-?\d+°)$/.exec(label)
    const title = m ? m[1] : 'kondisi diri'
    short[i] = m ? m[2] : label
    let g = groups.at(-1)
    if (!g || g.title !== title) {
      g = { title, indices: [] }
      groups.push(g)
    }
    g.indices.push(i)
  })
  return { groups, short }
}

function spread(n: number, top: number, bottom: number) {
  return Array.from({ length: n }, (_, k) => (n === 1 ? (top + bottom) / 2 : top + (k * (bottom - top)) / (n - 1)))
}

function computeLayout(brain: Brain, width: number): Layout {
  const { groups, short } = groupInputs(brain.inputLabels)
  const xIn = 112
  const xOut = width - 82
  const xHid = (xIn + xOut) / 2

  const inputs: Node[] = []
  const heads: Layout['groups'] = []
  let y = TOP
  for (const g of groups) {
    y += GROUP_GAP
    heads.push({ title: g.title, y: y - 4 })
    for (const i of g.indices) {
      inputs[i] = { kind: 'in', i, x: xIn, y: y + ROW / 2, r: 3.5 }
      y += ROW
    }
  }
  const top = TOP + GROUP_GAP + ROW / 2
  const bottom = y - ROW / 2
  const span = bottom - top

  // Brains evolve between ~8 and 64 hidden neurons; shrink the dots so they never overlap.
  const nh = brain.hidden.length
  const pitch = nh > 1 ? (bottom - top - 16) / (nh - 1) : 20
  const rHid = clamp(pitch / 2 - 1, 2, 5.5)
  const hidden = spread(nh, top + 8, bottom - 8).map((hy, i): Node => ({
    kind: 'hid',
    i,
    x: xHid,
    y: hy,
    r: rHid,
  }))
  const outputs = spread(brain.output.length, top + span * 0.18, bottom - span * 0.18).map((oy, i): Node => ({
    kind: 'out',
    i,
    x: xOut,
    y: oy,
    r: 6.5,
  }))

  const n = brain.wRec.length
  const cell = Math.max(MIN_CELL, Math.min(MAX_CELL, Math.floor((width - 8) / Math.max(1, n))))
  const heatY = y + 30
  return {
    width,
    height: heatY + n * cell + 8,
    xIn,
    xHid,
    xOut,
    inputs,
    hidden,
    outputs,
    groups: heads,
    short,
    heat: { x: Math.round((width - n * cell) / 2), y: heatY, n, cell },
  }
}

type Edge = { a: Node; b: Node; c: number }

function strongest(edges: Edge[], max: number) {
  return edges
    .filter((e) => Math.abs(e.c) > 0.02)
    .sort((p, q) => Math.abs(q.c) - Math.abs(p.c))
    .slice(0, max)
}

function drawEdges(ctx: CanvasRenderingContext2D, edges: Edge[], dim: number) {
  const maxC = edges.reduce((m, e) => Math.max(m, Math.abs(e.c)), 0) || 1
  for (const e of edges) {
    const t = Math.abs(e.c) / maxC
    ctx.strokeStyle = e.c >= 0 ? `rgba(${C.pos.join(',')},${(0.12 + 0.68 * t) * dim})` : `rgba(${C.neg.join(',')},${(0.12 + 0.68 * t) * dim})`
    ctx.lineWidth = 0.5 + 1.5 * t
    ctx.beginPath()
    ctx.moveTo(e.a.x, e.a.y)
    ctx.lineTo(e.b.x, e.b.y)
    ctx.stroke()
  }
}

/** Shortens text with an ellipsis so it fits in maxWidth pixels. */
function fit(ctx: CanvasRenderingContext2D, text: string, maxWidth: number) {
  if (ctx.measureText(text).width <= maxWidth) return text
  let t = text
  while (t.length > 1 && ctx.measureText(`${t}…`).width > maxWidth) t = t.slice(0, -1)
  return `${t.trimEnd()}…`
}

function isActiveOutput(i: number, v: number) {
  return i === 0 ? Math.abs(v) > 0.25 : v > 0.5
}

function draw(ctx: CanvasRenderingContext2D, brain: Brain, L: Layout, hoverKey: string | null) {
  const { input, hidden, output, wIn, wOut, wRec } = brain

  const allIn: Edge[] = []
  L.inputs.forEach((a, i) =>
    L.hidden.forEach((b, h) => allIn.push({ a, b, c: (wIn[i]?.[h] ?? 0) * (input[i] ?? 0) })),
  )
  const allOut: Edge[] = []
  L.hidden.forEach((a, h) =>
    L.outputs.forEach((b, o) => allOut.push({ a, b, c: (wOut[h]?.[o] ?? 0) * (hidden[h] ?? 0) })),
  )

  const focus = hoverKey ? (e: Edge) => `${e.a.kind}:${e.a.i}` === hoverKey || `${e.b.kind}:${e.b.i}` === hoverKey : null
  const dim = focus ? 0.3 : 1
  drawEdges(ctx, strongest(allIn, MAX_IN_EDGES), dim)
  drawEdges(ctx, strongest(allOut, MAX_OUT_EDGES), dim)
  if (focus) {
    drawEdges(ctx, strongest(allIn.filter(focus), 16), 1)
    drawEdges(ctx, strongest(allOut.filter(focus), 16), 1)
  }

  // Column titles.
  ctx.font = FONT_BOLD
  ctx.fillStyle = C.muted
  // The hidden and output columns sit close together, so anchor titles away from each other.
  ctx.textAlign = 'right'
  ctx.fillText('Indra', L.xIn + 4, 10)
  ctx.textAlign = 'center'
  ctx.fillText('Saraf', L.xHid, 10)
  ctx.textAlign = 'left'
  ctx.fillText('Keputusan', L.xOut - 6, 10)

  ctx.font = FONT_HEAD
  ctx.fillStyle = C.dim
  ctx.textAlign = 'left'
  for (const g of L.groups) ctx.fillText(g.title.toUpperCase(), 2, g.y)

  const node = (n: Node, v: number) => {
    ctx.fillStyle = diverging(v)
    ctx.beginPath()
    ctx.arc(n.x, n.y, n.r, 0, Math.PI * 2)
    ctx.fill()
    const hovered = `${n.kind}:${n.i}` === hoverKey
    ctx.strokeStyle = hovered ? C.text : C.ring
    ctx.lineWidth = hovered ? 2 : 1
    ctx.stroke()
  }

  ctx.font = FONT
  ctx.textAlign = 'right'
  L.inputs.forEach((n, i) => {
    const v = input[i] ?? 0
    node(n, v)
    ctx.fillStyle = Math.abs(v) > 0.5 ? C.text : C.muted
    // Long labels are ellipsized; the hover tooltip shows them in full.
    ctx.fillText(fit(ctx, L.short[i], n.x - n.r - 6), n.x - n.r - 4, n.y + 3.5)
  })

  L.hidden.forEach((n, h) => node(n, hidden[h] ?? 0))
  // Lifetime learning: a ring whose strength shows how much this neuron's weights moved.
  if (brain.learned?.length) {
    // Keep rings inside the gap between neighbours when the brain is large.
    const gap = L.hidden.length > 1 ? (L.hidden[1].y - L.hidden[0].y) / 2 - L.hidden[0].r : 2.5
    const off = clamp(gap, 1, 2.5)
    L.hidden.forEach((n, h) => {
      const t = clamp(brain.learned[h] ?? 0, 0, 1)
      if (t < 0.05) return
      ctx.strokeStyle = `rgba(${C.learned.join(',')},${0.35 + 0.65 * t})`
      ctx.lineWidth = 1 + 1.5 * t
      ctx.beginPath()
      ctx.arc(n.x, n.y, n.r + off, 0, Math.PI * 2)
      ctx.stroke()
    })
  }

  ctx.textAlign = 'left'
  L.outputs.forEach((n, o) => {
    const v = output[o] ?? 0
    // Sigmoid outputs are drawn centred on their 0.5 threshold so "off" reads blue.
    node(n, o === 0 ? v : (v - 0.5) * 2)
    const active = isActiveOutput(o, v)
    ctx.font = active ? FONT_BOLD : FONT
    ctx.fillStyle = active ? C.text : C.muted
    ctx.fillText(brain.outputLabels[o] ?? `o${o + 1}`, n.x + n.r + 5, n.y - 1)
    ctx.font = FONT
    ctx.fillStyle = active ? C.text : C.dim
    ctx.fillText(o === 0 ? (v >= 0 ? '+' : '') + v.toFixed(2) : v.toFixed(2), n.x + n.r + 5, n.y + 10)
  })

  // Recurrent weights: the brain's short-term memory wiring.
  const { x, y, n, cell } = L.heat
  ctx.font = FONT_BOLD
  ctx.fillStyle = C.muted
  ctx.textAlign = 'center'
  ctx.fillText(`Memori: bobot rekuren h→h (${n}×${n})`, L.width / 2, y - 9)
  const maxW = wRec.reduce((m, row) => row.reduce((mm, w) => Math.max(mm, Math.abs(w)), m), 0) || 1
  for (let from = 0; from < n; from++) {
    for (let to = 0; to < n; to++) {
      ctx.fillStyle = diverging((wRec[from]?.[to] ?? 0) / maxW)
      ctx.fillRect(x + to * cell, y + from * cell, cell - 1, cell - 1)
    }
  }
}

type Hover = { key: string; x: number; y: number; text: string }

function describe(brain: Brain, kind: Kind, i: number) {
  if (kind === 'in') return `${brain.inputLabels[i]}: ${(brain.input[i] ?? 0).toFixed(2)}`
  if (kind === 'hid') {
    const learned = brain.learned?.[i]
    return `Neuron h${i + 1}: ${(brain.hidden[i] ?? 0).toFixed(2)}${
      learned != null ? ` · berubah karena belajar ${Math.round(learned * 100)}%` : ''
    }`
  }
  const v = brain.output[i] ?? 0
  return `${brain.outputLabels[i] ?? `o${i + 1}`}: ${v.toFixed(2)}${isActiveOutput(i, v) ? ' (aktif)' : ''}`
}

export function BrainView({ brain }: { brain: Brain }) {
  const [wrapRef, width] = useElementWidth<HTMLDivElement>()
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [hover, setHover] = useState<Hover | null>(null)

  const layout = useMemo(
    () => (width >= 220 ? computeLayout(brain, width) : null),
    // Labels and sizes are fixed per creature; structural sharing keeps these references stable.
    [brain.inputLabels, brain.hidden.length, brain.output.length, brain.wRec.length, width],
  )

  const hoverKey = hover?.key ?? null
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || !layout) return
    const dpr = window.devicePixelRatio || 1
    canvas.width = Math.round(layout.width * dpr)
    canvas.height = Math.round(layout.height * dpr)
    canvas.style.width = `${layout.width}px`
    canvas.style.height = `${layout.height}px`
    const ctx = canvas.getContext('2d')!
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, layout.width, layout.height)
    draw(ctx, brain, layout, hoverKey)
  }, [brain, layout, hoverKey])

  const onPointerMove = (e: PointerEvent<HTMLCanvasElement>) => {
    if (!layout) return
    const rect = e.currentTarget.getBoundingClientRect()
    const px = e.clientX - rect.left
    const py = e.clientY - rect.top

    let best: Node | null = null
    let bestD = Infinity
    for (const n of [...layout.inputs, ...layout.hidden, ...layout.outputs]) {
      const d = Math.hypot(n.x - px, n.y - py)
      if (d < n.r + 5 && d < bestD) {
        best = n
        bestD = d
      }
    }
    if (best) {
      const key = `${best.kind}:${best.i}`
      if (key !== hover?.key) {
        setHover({ key, x: best.x + best.r + 6, y: best.y - 10, text: describe(brain, best.kind, best.i) })
      }
      return
    }

    const { x, y, n, cell } = layout.heat
    const col = Math.floor((px - x) / cell)
    const row = Math.floor((py - y) / cell)
    if (col >= 0 && row >= 0 && col < n && row < n) {
      const key = `rec:${row}:${col}`
      if (key !== hover?.key) {
        const w = brain.wRec[row]?.[col] ?? 0
        setHover({ key, x: x + col * cell + cell + 4, y: y + row * cell - 10, text: `h${row + 1} → h${col + 1}: ${w.toFixed(2)}` })
      }
      return
    }
    if (hover) setHover(null)
  }

  // Keep the tooltip text live while the pointer rests on a neuron.
  const [hoverKind, hoverIndex] = hover?.key.split(':') ?? []
  const tooltip = hover && hoverKind !== 'rec' ? describe(brain, hoverKind as Kind, Number(hoverIndex)) : hover?.text

  const activeOutputs = brain.outputLabels.filter((_, o) => isActiveOutput(o, brain.output[o] ?? 0))

  return (
    <div ref={wrapRef} className="obs-brain">
      <canvas
        ref={canvasRef}
        role="img"
        aria-label={`Jaringan saraf ${brain.input.length}→${brain.hidden.length}→${brain.output.length}. Keputusan aktif: ${activeOutputs.join(', ') || 'tidak ada'}.`}
        onPointerMove={onPointerMove}
        onPointerLeave={() => setHover(null)}
      />
      {hover && tooltip && (
        <div className="obs-tooltip obs-brain-tip" style={{ left: Math.min(hover.x, width - 170), top: Math.max(0, hover.y) }}>
          {tooltip}
        </div>
      )}
      <p className="obs-legend-row small muted">
        <i className="obs-swatch-sm" style={{ background: diverging(1) }} /> positif
        <i className="obs-swatch-sm" style={{ background: diverging(0) }} /> nol
        <i className="obs-swatch-sm" style={{ background: diverging(-1) }} /> negatif
        {brain.learned?.length ? (
          <>
            <i className="obs-swatch-sm obs-swatch-ring" style={{ borderColor: `rgb(${C.learned.join(',')})` }} /> berubah
            karena belajar
          </>
        ) : null}
      </p>
    </div>
  )
}
