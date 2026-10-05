import type { Stack } from '../../sim/protocol'
import { nf } from './format'

/** Item stacks as compact chips: "Kayu ×12". */
export function Chips({ stacks, empty }: { stacks: Stack[] | undefined; empty: string }) {
  if (!stacks?.length) return <p className="small muted obs-chips-empty">{empty}</p>
  return (
    <ul className="obs-chips">
      {stacks.map((s) => (
        <li key={s.item}>
          {s.name || s.item} <strong>×{nf.format(Math.round(s.qty * 10) / 10)}</strong>
        </li>
      ))}
    </ul>
  )
}
