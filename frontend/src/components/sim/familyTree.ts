import type { FamilyPerson, FamilyTree } from '../../sim/protocol'

/** Indonesian (Javanese-derived) names of the generations up and down. */
const UP = ['Orang tua', 'Kakek & nenek', 'Buyut', 'Canggah', 'Wareng', 'Udeg-udeg', 'Gantung siwur']
const DOWN = ['Anak', 'Cucu', 'Cicit', 'Piut']

/** The name of the generation `level` steps up (positive) or down (negative). */
export function generationLabel(level: number): string {
  if (level > 0) return UP[level - 1] ?? `Leluhur ke-${level}`
  if (level < 0) return DOWN[-level - 1] ?? `Keturunan ke-${-level}`
  return 'Diri sendiri'
}

export type TreeSlot = { id: number; count: number }
export type TreeRow = { level: number; label: string; people: TreeSlot[]; places: number }

/**
 * Ancestor rows, nearest first: each person once per level in Ahnentafel
 * order, with how many places of that level they fill (more than one is
 * pedigree collapse: their descendants married each other).
 */
export function ancestorRows(tree: FamilyTree): TreeRow[] {
  return (tree.ancestors ?? []).map((slots, i) => {
    const people: TreeSlot[] = []
    const at = new Map<number, TreeSlot>()
    let places = 0
    for (const id of slots) {
      if (!id) continue
      places++
      const seen = at.get(id)
      if (seen) seen.count++
      else {
        const slot = { id, count: 1 }
        at.set(id, slot)
        people.push(slot)
      }
    }
    return { level: i + 1, label: generationLabel(i + 1), people, places }
  })
}

/** Descendant rows, nearest first. */
export function descendantRows(tree: FamilyTree): TreeRow[] {
  return (tree.descendants ?? []).map((ids, i) => ({
    level: -(i + 1),
    label: generationLabel(-(i + 1)),
    people: ids.map((id) => ({ id, count: 1 })),
    places: ids.length,
  }))
}

/** How many distinct ancestors a tree shows against the places it has (2 + 4 + …). */
export function collapse(tree: FamilyTree): { distinct: number; places: number } {
  const distinct = new Set<number>()
  let places = 0
  for (const slots of tree.ancestors ?? []) {
    for (const id of slots) {
      if (!id) continue
      places++
      distinct.add(id)
    }
  }
  return { distinct: distinct.size, places }
}

const nf4 = new Intl.NumberFormat('id-ID', { maximumFractionDigits: 4 })

/** F with up to four decimals, Indonesian style: "0,0625". */
export const formatF = (f: number) => nf4.format(f)

/** What an inbreeding coefficient is like, in words. */
export function inbreedingWord(f: number): string {
  if (f <= 0) return 'orang tua tidak sekerabat'
  if (f >= 0.25 - 1e-6) return f > 0.26 ? 'lebih dari anak saudara kandung' : 'setara anak saudara kandung'
  if (f >= 0.125 - 1e-6) return 'setara anak saudara tiri atau paman–keponakan'
  if (f >= 0.0625 - 1e-6) return 'setara anak sepupu'
  if (f >= 1 / 64 - 1e-6) return 'setara anak sepupu dua kali'
  return 'kerabat jauh'
}

/** What a coefficient of relationship is like, in words. */
export function relationWord(r: number): string {
  if (r <= 0) return 'tidak sekerabat'
  if (r >= 0.5 - 1e-6) return 'sedekat saudara kandung atau orang tua–anak'
  if (r >= 0.25 - 1e-6) return 'sedekat saudara tiri, kakek–cucu, atau paman–keponakan'
  if (r >= 0.125 - 1e-6) return 'sedekat sepupu'
  if (r >= 1 / 32 - 1e-6) return 'sedekat sepupu dua kali'
  return 'kerabat jauh'
}

/** A person's tooltip line. */
export function personTitle(p: FamilyPerson | undefined): string {
  if (!p) return 'Tidak dikenal'
  if (p.forgotten) return `${p.name || 'Tanpa nama'} · sudah terlupakan dari silsilah`
  const parts = [`Generasi ${p.generation}`, `lahir tahun ${p.born}`]
  if (p.died != null) parts.push(`wafat tahun ${p.died}`)
  else if (!p.alive) parts.push('sudah wafat')
  if (p.f > 0) parts.push(`F ${formatF(p.f)}`)
  if (p.founder) parts.push('manusia pertama')
  return `${p.name} · ${parts.join(' · ')}`
}
