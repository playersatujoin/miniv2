import { useState } from 'react'
import type { SimEvent, SimEventKind } from '../../sim/protocol'
import { formatClockShort, useSecondsPerYear } from './format'

// Dots are a secondary cue; the kind label next to each dot carries the meaning.
const KIND: Record<SimEventKind | 'immigrant', { label: string; color: string }> = {
  birth: { label: 'Lahir', color: '#199e70' },
  death: { label: 'Wafat', color: '#898781' },
  genesis: { label: 'Awal mula', color: '#9085e9' },
  milestone: { label: 'Tonggak', color: '#c98500' },
  discovery: { label: 'Penemuan', color: '#3987e5' },
  build: { label: 'Pembangunan', color: '#d95926' },
  crime: { label: 'Kejahatan', color: '#e66767' },
  kindness: { label: 'Kebaikan', color: '#008300' },
  family: { label: 'Keluarga', color: '#d55181' },
  immigrant: { label: 'Pendatang', color: '#9085e9' },
}

const FILTERS: { id: string; label: string; kinds: string[] | null }[] = [
  { id: 'all', label: 'Semua', kinds: null },
  { id: 'life', label: 'Hidup & mati', kinds: ['birth', 'death', 'genesis', 'family'] },
  { id: 'progress', label: 'Kemajuan', kinds: ['discovery', 'build', 'milestone'] },
  { id: 'moral', label: 'Moral', kinds: ['crime', 'kindness'] },
]

type Props = { events: SimEvent[]; onSelect: (id: number) => void }

export function EventLog({ events, onSelect }: Props) {
  const [filter, setFilter] = useState('all')
  const spy = useSecondsPerYear()
  const kinds = FILTERS.find((f) => f.id === filter)?.kinds
  const items = [...events].reverse().filter((e) => !kinds || kinds.includes(e.kind))

  return (
    <section className="obs-section">
      <h3 className="obs-title">Peristiwa</h3>
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
      {items.length === 0 ? (
        <p className="small muted">Belum ada peristiwa{filter === 'all' ? '' : ' jenis ini'}.</p>
      ) : (
        <ol className="obs-events">
          {items.map((e) => {
            const kind = KIND[e.kind] ?? { label: e.kind, color: '#898781' }
            // The dead can't be inspected, so only living subjects are clickable.
            const clickable = e.creatureId != null && e.kind !== 'death'
            const body = (
              <>
                <span className="obs-ev-meta">
                  <i className="obs-dot" style={{ background: kind.color }} />
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
