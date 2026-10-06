/** Presentation anatomy. These proportions never change simulation fitness. */
export function bodyProportions(age: number, seed: number, female: boolean) {
  const maturity = Math.min(1, Math.max(0, age / 16))
  const growth = 0.34 + 0.66 * Math.pow(maturity, 0.62)
  return {
    height: growth * (0.96 + seed * 0.08),
    width: growth * (female ? 0.92 : 1) * (0.94 + seed * 0.12),
    head: 1 + 0.65 * (1 - maturity),
    stoop: Math.max(0, Math.min(0.18, (age - 55) * 0.006)),
    grey: Math.max(0, Math.min(0.8, (age - 40) / 45)),
  }
}

/** Two links pointing down at rest; positive knee flexion. Target is in the
 * hip's sagittal plane. Clamp unreachable targets to prevent NaN and snapping. */
export function solveLeg(y: number, z: number, thigh = 0.115, shin = 0.109) {
  const distance = Math.min(thigh + shin - 0.0001, Math.max(Math.abs(thigh - shin) + 0.0001, Math.hypot(y, z)))
  const knee = Math.acos(Math.max(-1, Math.min(1, (distance * distance - thigh * thigh - shin * shin) / (2 * thigh * shin))))
  const hip = Math.atan2(-z, -y) - Math.atan2(shin * Math.sin(knee), thigh + shin * Math.cos(knee))
  return { hip, knee }
}
