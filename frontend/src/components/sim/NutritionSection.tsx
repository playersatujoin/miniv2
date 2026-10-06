import { NUTRITION_LABELS, NUTRITION_NEED_LABELS, type NutritionView } from '../../sim/protocol'
import { NUTRITION_COLOR, dietWord, statureText, statusLevel, statusText } from './nutrition'

const DIET_HINT = 'Kadar protein / zat gizi mikro makanan yang sedang dihidupinya, dibanding kebutuhan tubuhnya sendiri'
const STATURE_HINT = 'Tinggi badan dibanding orang sebaya yang cukup gizi; yang hilang di masa kecil sebagian besar tidak kembali'

/** One body store: protein or micronutrients, 1 = adequate (a reserve above that shows as a full bar). */
function StoreMeter({ label, status }: { label: string; status: number }) {
  const level = statusLevel(status)
  const fill = NUTRITION_COLOR[level]
  const v = Math.min(1, Math.max(0, status))
  return (
    <div className="obs-meter">
      <div className="obs-meter-head">
        <span>{label}</span>
        <span className="obs-meter-value">
          {statusText(status)}
          {level !== 'baik' && (
            <em className={`obs-level obs-level-${level === 'buruk' ? 'critical' : 'warning'}`}> · {level}</em>
          )}
        </span>
      </div>
      <div className="obs-meter-track" style={{ background: `color-mix(in srgb, ${fill} 22%, transparent)` }}>
        <div className="obs-meter-fill" style={{ width: `${v * 100}%`, background: fill }} />
      </div>
    </div>
  )
}

/**
 * A person's nutrition beyond calories (Fase 3d): the gizi label, the body's protein and
 * micronutrient status, what their food gives against their need, and height lost in childhood.
 * Renders nothing when the server sends no nutrition (switched off, or an older server).
 */
export function NutritionSection({ nutrition: n }: { nutrition?: NutritionView | null }) {
  if (!n) return null
  const color = NUTRITION_COLOR[n.label]
  return (
    <div style={{ display: 'grid', gap: 8 }}>
      <p className="small obs-note">
        <span className="obs-pill" style={{ borderColor: color }}>
          {n.label === 'baik' ? '●' : n.label === 'kurang' ? '▲' : '■'} {NUTRITION_LABELS[n.label]}
        </span>
        {n.thin && <span className="muted"> · kurus (cadangan energi rendah)</span>}
      </p>
      <div className="obs-meters">
        <StoreMeter label="Protein" status={n.protein} />
        <StoreMeter label="Zat gizi mikro (vitamin & mineral)" status={n.micro} />
      </div>
      <dl className="obs-kv">
        <dt title={DIET_HINT}>Makanannya</dt>
        <dd title={DIET_HINT}>
          protein {dietWord(n.dietProtein)} ({Math.round(n.dietProtein * 100)}%) · mikro {dietWord(n.dietMicro)} (
          {Math.round(n.dietMicro * 100)}%)
        </dd>
        <dt>Kebutuhan</dt>
        <dd>{NUTRITION_NEED_LABELS[n.need] ?? n.need}</dd>
        <dt title={STATURE_HINT}>Tinggi badan</dt>
        <dd title={STATURE_HINT}>{statureText(n.stature, n.stunted)}</dd>
      </dl>
    </div>
  )
}
