import test from 'node:test'
import assert from 'node:assert/strict'
import {
  ancestorRows,
  collapse,
  descendantRows,
  formatF,
  generationLabel,
  inbreedingWord,
  relationWord,
} from '../src/components/sim/familyTree.ts'

// Kabil and Iqlima, both children of Adam (1) and Hawa (2), have a child.
const tree = {
  root: 5,
  depth: 3,
  ancestors: [
    [3, 4],
    [1, 2, 1, 2],
  ],
  descendants: [[7, 8], [9]],
  people: [],
  genetics: null,
  parentsRelated: 0.5,
}

test('ancestor rows count pedigree collapse', () => {
  const rows = ancestorRows(tree)
  assert.equal(rows.length, 2)
  assert.equal(rows[0].label, 'Orang tua')
  assert.deepEqual(rows[1].people, [
    { id: 1, count: 2 },
    { id: 2, count: 2 },
  ])
  assert.equal(rows[1].places, 4)
  assert.deepEqual(collapse(tree), { distinct: 4, places: 6 })
})

test('unknown places are skipped', () => {
  const rows = ancestorRows({ ...tree, ancestors: [[3, 0], [0, 0, 0, 0]] })
  assert.deepEqual(rows[0].people, [{ id: 3, count: 1 }])
  assert.equal(rows[1].people.length, 0)
})

test('descendant rows and generation names', () => {
  const rows = descendantRows(tree)
  assert.deepEqual(
    rows.map((r) => r.label),
    ['Anak', 'Cucu'],
  )
  assert.equal(generationLabel(3), 'Buyut')
  assert.equal(generationLabel(7), 'Gantung siwur')
  assert.equal(generationLabel(9), 'Leluhur ke-9')
  assert.equal(generationLabel(-5), 'Keturunan ke-5')
})

test('inbreeding and relationship words', () => {
  assert.equal(inbreedingWord(0), 'orang tua tidak sekerabat')
  assert.equal(inbreedingWord(0.0625), 'setara anak sepupu')
  assert.equal(inbreedingWord(0.25), 'setara anak saudara kandung')
  assert.equal(inbreedingWord(0.375), 'lebih dari anak saudara kandung')
  assert.equal(relationWord(0.5), 'sedekat saudara kandung atau orang tua–anak')
  assert.equal(relationWord(0.125), 'sedekat sepupu')
  assert.equal(formatF(0.0625), '0,0625')
})
