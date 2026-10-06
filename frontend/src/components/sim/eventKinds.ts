import type { SimEventKind } from '../../sim/protocol'

/** localStorage key of the log's "Sembunyikan kekerasan" (the replay's timeline honours it too). */
export const HIDE_VIOLENCE_KEY = 'miniv2.hideViolence'

/**
 * How each kind of event is marked in the log and on the replay's timeline.
 * Dots are a secondary cue; the kind label next to each dot carries the
 * meaning. The eight palette hues are taken, so kinds share a hue within a
 * family and differ in shape: filled dots for people's lives and progress,
 * rings for the land's events (Alam: climate aqua, ecology green, farming
 * yellow, hunting orange) and for learning (knowledge, discovery's blue), and
 * diamonds for society (engine adaptation II: villages violet, trade yellow,
 * news magenta) and fire (red, with the other dangers).
 */
export type KindStyle = { label: string; color: string; shape?: 'ring' | 'diamond' }

export const EVENT_KINDS: Record<SimEventKind | 'immigrant', KindStyle> = {
  birth: { label: 'Lahir', color: '#199e70' },
  death: { label: 'Wafat', color: '#898781' },
  genesis: { label: 'Awal mula', color: '#9085e9' },
  milestone: { label: 'Tonggak', color: '#c98500' },
  discovery: { label: 'Penemuan', color: '#3987e5' },
  build: { label: 'Pembangunan', color: '#d95926' },
  crime: { label: 'Kejahatan', color: '#e66767' },
  kindness: { label: 'Kebaikan', color: '#008300' },
  family: { label: 'Keluarga', color: '#d55181' },
  learning: { label: 'Belajar', color: '#3987e5', shape: 'ring' },
  climate: { label: 'Iklim', color: '#199e70', shape: 'ring' },
  ecology: { label: 'Ekologi', color: '#008300', shape: 'ring' },
  farming: { label: 'Pertanian', color: '#c98500', shape: 'ring' },
  hunt: { label: 'Perburuan', color: '#d95926', shape: 'ring' },
  disease: { label: 'Wabah', color: '#e66767', shape: 'ring' },
  village: { label: 'Desa', color: '#9085e9', shape: 'diamond' },
  fire: { label: 'Kebakaran', color: '#e66767', shape: 'diamond' },
  trade: { label: 'Dagang', color: '#c98500', shape: 'diamond' },
  rumor: { label: 'Kabar', color: '#d55181', shape: 'diamond' },
  immigrant: { label: 'Pendatang', color: '#9085e9' },
}

/** Unknown kinds (a newer server) still show, in grey under their own name. */
export function kindStyle(kind: string): KindStyle {
  return (EVENT_KINDS as Record<string, KindStyle>)[kind] ?? { label: kind, color: '#898781' }
}

/** Inline style for a kind's dot (className obs-dot plus obs-dot-ring / obs-dot-diamond). */
export function dotStyle(k: KindStyle) {
  return k.shape === 'ring' ? { borderColor: k.color } : { background: k.color }
}

export function dotClass(k: KindStyle) {
  return `obs-dot${k.shape ? ` obs-dot-${k.shape}` : ''}`
}

/**
 * Assaults, killings and deaths by animals. Older servers (and the replay's
 * markers) don't flag events, so fall back to their wording.
 */
export function isViolent(e: { kind: string; text: string; violent?: boolean }) {
  if (e.violent != null) return e.violent
  return (e.kind === 'crime' && /menyerang/.test(e.text)) || (e.kind === 'death' && /dibunuh|diterkam|diseruduk|ditanduk/.test(e.text))
}

/** Whether the viewer asked to hide violence (best effort: storage may be unavailable). */
export function violenceHidden() {
  try {
    return localStorage.getItem(HIDE_VIOLENCE_KEY) === '1'
  } catch {
    return false
  }
}
