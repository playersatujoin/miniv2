import { useMemo, useState } from 'react'
import type { MapGeology } from '../api/client'

const STORAGE_KEY = 'miniv2.geologyOverlay'

/** The geological-map toggle, remembered per viewer (best effort: storage may be unavailable). */
export function useGeologyOverlay(): [boolean, (on: boolean) => void] {
  const [on, setOn] = useState(() => {
    try {
      return localStorage.getItem(STORAGE_KEY) === '1'
    } catch {
      return false
    }
  })
  const set = (next: boolean) => {
    setOn(next)
    try {
      localStorage.setItem(STORAGE_KEY, next ? '1' : '0')
    } catch {
      // Private mode or blocked storage: the toggle still works for this visit.
    }
  }
  return [on, set]
}

/** Legend for the geological map: the rock units and deposit models present on this map. */
export function GeologyLegend({ geology }: { geology: MapGeology }) {
  const [open, setOpen] = useState(true)

  const { rocks, models } = useMemo(() => {
    const rockIds = new Set(geology.rocks)
    const modelIds = new Set(geology.deposits.map((d) => d[4]))
    return {
      rocks: geology.rockTypes.filter((r) => r.id !== 0 && rockIds.has(r.id)),
      models: geology.models.filter((_, i) => modelIds.has(i)),
    }
  }, [geology])

  return (
    <section className="geo-legend" aria-label="Legenda peta geologi">
      <button type="button" className="geo-legend-head" aria-expanded={open} onClick={() => setOpen(!open)}>
        🪨 Peta geologi <span aria-hidden>{open ? '▾' : '▸'}</span>
      </button>
      {open && (
        <div className="geo-legend-body">
          <h4>Batuan</h4>
          <ul className="geo-rocks">
            {rocks.map((r) => (
              <li key={r.id} title={r.description}>
                <i style={{ background: r.color }} aria-hidden />
                {r.name}
              </li>
            ))}
          </ul>
          {models.length > 0 && (
            <>
              <h4>Endapan</h4>
              <ul className="geo-models">
                {models.map((m) => (
                  <li key={m.key} title={m.description}>
                    {m.name}
                    {m.example && <small>contoh: {m.example}</small>}
                  </li>
                ))}
              </ul>
            </>
          )}
        </div>
      )}
    </section>
  )
}
