import { useState } from 'react'
import type { SimEvent, SimEventKind } from '../../sim/protocol'
import { formatClockShort, useSecondsPerYear } from './format'

// Dots are a secondary cue; the kind label next to each dot carries the meaning.
// The eight palette hues are taken, so "learning" shares discovery's blue (both
// are knowledge) but is drawn as a ring instead of a filled dot. The land's
// events (Alam) are rings too, each on a different hue: climate aqua, ecology
// green, farming yellow (ripe grain), hunting orange.
const KIND: Record<SimEventKind | 'immigrant', { label: string; color: string; ring?: boolean }> = {
  birth: { label: 'Lahir', color: '#199e70' },
  death: { label: 'Wafat', color: '#898781' },
  genesis: { label: 'Awal mula', color: '#9085e9' },
  milestone: { label: 'Tonggak', color: '#c98500' },
  discovery: { label: 'Penemuan', color: '#3987e5' },
  build: { label: 'Pembangunan', color: '#d95926' },
  crime: { label: 'Kejahatan', color: '#e66767' },
  kindness: { label: 'Kebaikan', color: '#008300' },
  family: { label: 'Keluarga', color: '#d55181' },
  learning: { label: 'Belajar', color: '#3987e5', ring: true },
  climate: { label: 'Iklim', color: '#199e70', ring: true },
  ecology: { label: 'Ekologi', color: '#008300', ring: true },
  farming: { label: 'Pertanian', color: '#c98500', ring: true },
  hunt: { label: 'Perburuan', color: '#d95926', ring: true },
  immigrant: { label: 'Pendatang', color: '#9085e9' },
}

const FILTERS: { id: string; label: string; kinds: string[] | null }[] = [
  { id: 'all', label: 'Semua', kinds: null },
  { id: 'life', label: 'Hidup & mati', kinds: ['birth', 'death', 'genesis', 'family'] },
  { id: 'progress', label: 'Kemajuan', kinds: ['discovery', 'build', 'milestone', 'learning'] },
  { id: 'moral', label: 'Moral', kinds: ['crime', 'kindness'] },
  { id: 'nature', label: 'Alam', kinds: ['climate', 'ecology', 'farming', 'hunt'] },
]

const HIDE_VIOLENCE_KEY = 'miniv2.hideViolence'

/** Assaults, killings and deaths by animals. Older servers don't flag events, so fall back to their wording. */
function isViolent(e: SimEvent) {
  if (e.violent != null) return e.violent
  return (
    (e.kind === 'crime' && /menyerang/.test(e.text)) ||
    (e.kind === 'death' && /dibunuh|diterkam|diseruduk|ditanduk/.test(e.text))
  )
}

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
      {hideViolence && hidden > 0 && (
        <p className="small muted obs-hidden-note">{hidden} peristiwa kekerasan disembunyikan.</p>
      )}
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
                  <i
                    className={`obs-dot${kind.ring ? ' obs-dot-ring' : ''}`}
                    style={kind.ring ? { borderColor: kind.color } : { background: kind.color }}
                  />
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
