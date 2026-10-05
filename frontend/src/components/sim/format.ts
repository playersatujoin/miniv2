import { createContext, useContext } from 'react'
import type { Sex } from '../../sim/protocol'

/** Fallback when the server doesn't report it yet: 1 simulated year = 8 s. */
export const DEFAULT_SECONDS_PER_YEAR = 8

/** Seconds per simulated year for everything under the observer panel. */
export const SecondsPerYearContext = createContext(DEFAULT_SECONDS_PER_YEAR)

export function useSecondsPerYear() {
  return useContext(SecondsPerYearContext)
}

const spyOr = (spy?: number) => (spy && spy > 0 ? spy : DEFAULT_SECONDS_PER_YEAR)

/** Calendar year (1-based) at simulated time t. */
export function yearAt(t: number, spy?: number) {
  return Math.floor(Math.max(0, t) / spyOr(spy)) + 1
}

/** "Tahun 123" */
export function formatClock(t: number, spy?: number) {
  return `Tahun ${nf.format(yearAt(t, spy))}`
}

/** "Thn 123" for axes and log rows. */
export function formatClockShort(t: number, spy?: number) {
  return `Thn ${nf.format(yearAt(t, spy))}`
}

/** A span of simulated seconds as an age: "34 tahun", or "5 bulan" under a year. */
export function formatYears(seconds: number, spy?: number) {
  const years = Math.max(0, seconds) / spyOr(spy)
  if (years < 1) return `${Math.floor(years * 12)} bulan`
  return `${nf.format(Math.floor(years))} tahun`
}

export function hueColor(hue: number, lightness = 58) {
  return `hsl(${Math.round(hue)} 62% ${lightness}%)`
}

export const SEX_COLOR: Record<Sex, string> = { female: '#d55181', male: '#3987e5' }
export const SEX_LABEL: Record<Sex, string> = { female: 'Perempuan', male: 'Laki-laki' }

export const nf = new Intl.NumberFormat('id-ID')
export const nf1 = new Intl.NumberFormat('id-ID', { maximumFractionDigits: 1 })
const nf2 = new Intl.NumberFormat('id-ID', { minimumFractionDigits: 2, maximumFractionDigits: 2 })

/** A 0–1 share as a whole percentage. */
export const pct = (v: number) => `${Math.round(v * 100)}%`

/** A Gini coefficient with two decimals, e.g. "0,25". */
export const fmtGini = (v: number) => nf2.format(v)

/** Crustal abundance in ppm, readable across ten orders of magnitude. */
export function formatAbundance(ppm: number) {
  if (!ppm) return 'tidak ada di alam (sintetis)'
  if (ppm >= 10_000) return `${nf1.format(ppm / 10_000)}% kerak bumi`
  if (ppm >= 1) return `${nf.format(Math.round(ppm))} ppm`
  return `${ppm.toPrecision(2).replace('.', ',')} ppm`
}

/** Words for a reputation in −1..1, so the meter never relies on colour alone. */
export function reputationWord(r: number) {
  if (r <= -0.6) return 'dibenci'
  if (r <= -0.2) return 'dicurigai'
  if (r < 0.2) return 'biasa'
  if (r < 0.6) return 'disukai'
  return 'dihormati'
}
