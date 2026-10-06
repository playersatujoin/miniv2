import { useState } from 'react'
import type { SimEvent } from '../../sim/protocol'
import { HIDE_VIOLENCE_KEY, dotClass, dotStyle, isViolent, kindStyle } from './eventKinds'
import { formatClockShort, useSecondsPerYear } from './format'

const FILTERS: { id: string; label: string; kinds: string[] | null }[] = [
  { id: 'all', label: 'Semua', kinds: null },
  { id: 'life', label: 'Hidup & mati', kinds: ['birth', 'death', 'genesis', 'family', 'disease'] },
  { id: 'progress', label: 'Kemajuan', kinds: ['discovery', 'build', 'milestone', 'learning'] },
  // Living together: deeds good and bad, barter, news passed on, villages and their leaders.
  { id: 'moral', label: 'Sosial', kinds: ['crime', 'kindness', 'trade', 'rumor', 'village'] },
  { id: 'nature', label: 'Alam', kinds: ['climate', 'ecology', 'farming', 'hunt', 'fire'] },
]

/** "Sembunyikan kekerasan", remembered per viewer (best effort: storage may be unavailable). */
function useHideViolence(): [boolean, (on: boolean) => void] {
  const [on, setOn] = useState(() => {
    try {
      return localStorage.getItem(HIDE_VIOLENCE_KEY) === '1'
    } catch {
      return false
    }
  })
  const set = (next: boolean) => {
    setOn(next)
    try {
      localStorage.setItem(HIDE_VIOLENCE_KEY, next ? '1' : '0')
    } catch {
      // Private mode or blocked storage: the filter still works for this visit.
    }
  }
  return [on, set]
}

type Props = { events: SimEvent[]; onSelect: (id: number) => void }

export function EventLog({ events, onSelect }: Props) {
  const [filter, setFilter] = useState('all')
  const [hideViolence, setHideViolence] = useHideViolence()
  const spy = useSecondsPerYear()
  const kinds = FILTERS.find((f) => f.id === filter)?.kinds
  const shown = [...events].reverse().filter((e) => !kinds || kinds.includes(e.kind))
  const items = hideViolence ? shown.filter((e) => !isViolent(e)) : shown
  const hidden = shown.length - items.length

  return (
    <section className="obs-section">
      <div className="obs-section-head">
        <h3 className="obs-title">Peristiwa</h3>
        <label className="obs-check small">
          <input type="checkbox" checked={hideViolence} onChange={(e) => setHideViolence(e.target.checked)} />
          Sembunyikan kekerasan
        </label>
      </div>
      <div className="segmented obs-ev-filter" role="group" aria-label="Saring peristiwa">
        {FILTERS.map((f) => (
          <button
            key={f.id}
            type="button"
            className={filter === f.id ? 'active' : ''}
            aria-pressed={filter === f.id}
            onClick={() => setFilter(f.id)}
          >
            {f.label}
          </button>
        ))}
      </div>
      {hideViolence && hidden > 0 && <p className="small muted obs-hidden-note">{hidden} peristiwa kekerasan disembunyikan.</p>}
      {items.length === 0 ? (
        <p className="small muted">Belum ada peristiwa{filter === 'all' ? '' : ' jenis ini'}.</p>
      ) : (
        <ol className="obs-events">
          {items.map((e) => {
            const kind = kindStyle(e.kind)
            // The dead can't be inspected, so only living subjects are clickable.
            const clickable = e.creatureId != null && e.kind !== 'death'
            const body = (
              <>
                <span className="obs-ev-meta">
                  <i className={dotClass(kind)} style={dotStyle(kind)} />
                  {kind.label} · {formatClockShort(e.time, spy)}
                </span>
                <span className="obs-ev-text">{e.text}</span>
              </>
            )
            return (
              <li key={e.id}>
                {clickable ? (
                  <button type="button" className="obs-ev obs-ev-click" onClick={() => onSelect(e.creatureId!)} title="Amati makhluk ini">
                    {body}
                  </button>
                ) : (
                  <div className="obs-ev">{body}</div>
                )}
              </li>
            )
          })}
        </ol>
      )}
    </section>
  )
}
