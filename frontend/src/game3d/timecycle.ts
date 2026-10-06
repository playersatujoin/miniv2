// Presentation lighting driven by the simulation, without changing its clock.
// `storm` (0–1) darkens it further under a slate sky; `flash` (0–1) is a lightning
// strike lighting the whole scene for a moment.
export function lightCycle(light: number, wet: number, storm = 0, flash = 0) {
  const day = Math.max(0, Math.min(1, light))
  const cloud = Math.max(0, Math.min(1, wet))
  const gale = Math.max(0, Math.min(1, storm))
  const strike = Math.max(0, Math.min(1, flash))
  const dusk = Math.max(0, 1 - Math.abs(day - 0.28) / 0.28)
  return {
    day,
    dusk,
    sun: (0.12 + day * 2.9) * (1 - cloud * 0.55) * (1 - 0.6 * gale),
    ambient: (0.24 + day * 0.95) * (1 - 0.3 * gale) + 2.2 * strike,
    exposure: (0.78 + day * 0.27) * (1 - 0.12 * gale) * (1 + 0.3 * strike),
    warmth: dusk * (1 - cloud * 0.7),
  }
}
