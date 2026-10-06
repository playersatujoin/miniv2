import test from 'node:test'
import assert from 'node:assert/strict'
import { NUTRITION_COLOR, dietWord, statureText, statusLevel, statusText } from '../src/components/sim/nutrition.ts'

test('body status reads with the server label thresholds', () => {
  assert.equal(statusLevel(1.2), 'baik')
  assert.equal(statusLevel(1), 'baik')
  assert.equal(statusLevel(0.86), 'baik')
  assert.equal(statusLevel(0.8), 'kurang')
  assert.equal(statusLevel(0.64), 'buruk')
  assert.equal(statusLevel(0), 'buruk')
  for (const level of ['baik', 'kurang', 'buruk']) assert.match(NUTRITION_COLOR[level], /^#[0-9a-f]{6}$/)
})

test('status, diet and stature in words', () => {
  assert.equal(statusText(1.2), 'cukup + cadangan')
  assert.equal(statusText(1), 'cukup')
  assert.equal(statusText(0.5), '50% dari cukup')
  assert.equal(dietWord(1.3), 'cukup')
  assert.equal(dietWord(0.9), 'hampir cukup')
  assert.equal(dietWord(0.7), 'kurang')
  assert.equal(dietWord(0.3), 'jauh kurang')
  assert.equal(statureText(1, false), '100% — normal')
  assert.equal(statureText(0.9, true), '90% — pendek (stunting)')
  assert.equal(statureText(0.96, false), '96% — sedikit tertinggal')
})
