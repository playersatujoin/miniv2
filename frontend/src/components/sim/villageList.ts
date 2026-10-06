// Villages for the Desa tab: sorting, the residents' trust in their leader,
// attitudes between villages and where to point the camera. Pure functions
// (only type imports), so node tests can check them.

import type { VillageStores, Village } from '../../sim/protocol'

/** Someone living in a village, as GET /sim/villages lists them. */
export type VillageResident = { id: number; name: string; house: number; adult: boolean; trust: number; leader?: boolean }

/** How one village's adults regard another: mean trust in the people of it they know. */
export type VillageAttitude = { id: number; name: string; trust: number; known: number; attitude: string }

/** A village with what the villages endpoint adds to the stream's fields (all optional: older servers). */
export type VillageDetail = Village &
  VillageStores & {
    /** Tiles of land. */
    area?: number
    houseIds?: number[]
    /** Planted plots of its families. */
    fields?: number
    buildings?: { id: number; kind: string; name: string; x: number; y: number }[]
    /** Simulated seconds when the current leader took over. */
    leaderSince?: number
    /** The leader's summed trust among the adults. */
    leaderTrust?: number
    supporters?: number
    /** How many people have led it. */
    leaders?: number
    residents?: VillageResident[]
    attitudes?: VillageAttitude[]
  }

/** Biggest first (people, then houses), then by name, so the list doesn't reshuffle between refreshes. */
export function sortVillages<V extends Village>(list: readonly V[] | null | undefined): V[] {
  return [...(list ?? [])].sort(
    (a, b) => (b.people ?? 0) - (a.people ?? 0) || (b.houses ?? 0) - (a.houses ?? 0) || a.name.localeCompare(b.name, 'id') || a.id - b.id,
  )
}

/** Trust above this counts as trusting the leader, below its negative as distrusting (as the Inspector's relations). */
const TRUST_SHOWN = 0.05

export type LeaderTrust = {
  /** Adults living there (the leader not counted). */
  adults: number
  /** Of them, those with an opinion of the leader. */
  opinions: number
  trusting: number
  distrusting: number
  /** Mean trust among those with an opinion, −1..1; null when nobody has one. */
  mean: number | null
}

/** How the village's adults regard their leader, from the residents list. */
export function leaderTrust(v: VillageDetail): LeaderTrust {
  const out: LeaderTrust = { adults: 0, opinions: 0, trusting: 0, distrusting: 0, mean: null }
  let sum = 0
  for (const r of v.residents ?? []) {
    if (!r.adult || r.leader || (v.leader && r.id === v.leader.id)) continue
    out.adults++
    if (!r.trust) continue
    out.opinions++
    sum += r.trust
    if (r.trust > TRUST_SHOWN) out.trusting++
    else if (r.trust < -TRUST_SHOWN) out.distrusting++
  }
  if (out.opinions) out.mean = sum / out.opinions
  return out
}

/** A word for a trust in −1..1. */
export function trustWord(t: number) {
  if (t >= 0.5) return 'sangat percaya'
  if (t > TRUST_SHOWN) return 'percaya'
  if (t > -TRUST_SHOWN) return 'netral'
  if (t > -0.5) return 'curiga'
  return 'sangat tidak percaya'
}

/** The backend's words for how one village regards another. */
export const ATTITUDE_LABELS: Record<string, string> = {
  hormat: 'Hormat',
  suka: 'Suka',
  acuh: 'Acuh',
  'tidak suka': 'Tidak suka',
  benci: 'Benci',
  asing: 'Belum saling kenal',
}

export const attitudeLabel = (a: string) => ATTITUDE_LABELS[a] ?? (a ? a[0].toUpperCase() + a.slice(1) : '—')

/** Attitudes towards other villages, warmest first (strangers last). */
export function sortAttitudes(list: readonly VillageAttitude[] | null | undefined): VillageAttitude[] {
  return [...(list ?? [])].sort(
    (a, b) => Number(a.known === 0) - Number(b.known === 0) || b.trust - a.trust || a.name.localeCompare(b.name, 'id'),
  )
}

/** Calendar year (1-based) at simulated time t, with s seconds per year. */
export function yearOf(t: number, spy: number) {
  return Math.floor(Math.max(0, t) / (spy > 0 ? spy : 8)) + 1
}

/** Whole simulated years between two times (never negative). */
export function yearsBetween(from: number, to: number, spy: number) {
  return Math.max(0, Math.floor((to - from) / (spy > 0 ? spy : 8)))
}

/**
 * Where to look to see the whole village: its hull's bounding box with a margin
 * (tiles across), at least `minAcross`; the centre when it has no hull.
 */
export function villageCamera(v: Pick<Village, 'x' | 'y' | 'hull'>, minAcross = 28): { x: number; y: number; across: number } {
  const hull = v.hull ?? []
  if (!hull.length) return { x: v.x, y: v.y, across: minAcross }
  let x0 = Infinity
  let y0 = Infinity
  let x1 = -Infinity
  let y1 = -Infinity
  for (const [x, y] of hull) {
    x0 = Math.min(x0, x)
    y0 = Math.min(y0, y)
    x1 = Math.max(x1, x)
    y1 = Math.max(y1, y)
  }
  // The view is wider than tall, so a tall village needs more across to fit.
  const across = Math.max(minAcross, (x1 - x0) * 1.3, (y1 - y0) * 1.3 * 1.6)
  return { x: (x0 + x1) / 2, y: (y0 + y1) / 2, across }
}

/** The topmost point of a hull (where a label sits clear of the crowd), or the centre. */
export function hullTop(v: Pick<Village, 'x' | 'y' | 'hull'>): [number, number] {
  let best: [number, number] | null = null
  for (const p of v.hull ?? []) if (!best || p[1] < best[1]) best = [p[0], p[1]]
  return best ?? [v.x, v.y]
}
