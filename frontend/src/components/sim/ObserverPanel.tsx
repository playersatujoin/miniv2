import { useQuery } from '@tanstack/react-query'
import { useCallback, useEffect, useRef, useState, type PointerEvent as ReactPointerEvent } from 'react'
import { knowledgeQuery, simInfoQuery } from '../../sim/api'
import { CivilizationView } from './CivilizationView'
import { Demography } from './Demography'
import { Ecology } from './Ecology'
import { EventLog } from './EventLog'
import { Inspector } from './Inspector'
import { PeriodicTable } from './PeriodicTable'
import { PopulationChart } from './PopulationChart'
import { PopulationStats } from './PopulationStats'
import { DEFAULT_SECONDS_PER_YEAR, SecondsPerYearContext } from './format'
import './sim.css'

export type ObserverPanelProps = {
  mapId: string
  selectedId: number | null
  following: boolean
  onSelect: (id: number | null) => void
  onFollow: (follow: boolean) => void
}

type Tab = 'population' | 'ecology' | 'civilization' | 'elements'

const TABS: { id: Tab; label: string }[] = [
  { id: 'population', label: 'Populasi' },
  { id: 'ecology', label: 'Ekologi' },
  { id: 'civilization', label: 'Peradaban' },
  { id: 'elements', label: 'Unsur' },
]

const HEIGHT_KEY = 'miniv2.dock.height'
const COLLAPSED_KEY = 'miniv2.dock.collapsed'
const MIN_HEIGHT = 160
/** About a third of the window, at most 340 px, so the map keeps most of a small screen. */
const defaultHeight = () => Math.max(MIN_HEIGHT, Math.min(340, Math.round(window.innerHeight * 0.36)))

const maxHeight = () => Math.max(MIN_HEIGHT, Math.round(window.innerHeight * 0.75))

function storedHeight() {
  const h = Number(localStorage.getItem(HEIGHT_KEY))
  return Number.isFinite(h) && h >= MIN_HEIGHT ? Math.min(h, maxHeight()) : defaultHeight()
}

/**
 * Read-only window into the simulation, docked under the map so the map gets
 * the full width: the observer watches, never steers. The clock and the time
 * control live in the toolbar (SimStatus).
 */
export function ObserverPanel({ mapId, selectedId, following, onSelect, onFollow }: ObserverPanelProps) {
  const [tab, setTab] = useState<Tab>('population')
  const [height, setHeight] = useState(storedHeight)
  const heightRef = useRef(height)
  heightRef.current = height
  const [collapsed, setCollapsed] = useState(() => localStorage.getItem(COLLAPSED_KEY) === '1')
  const info = useQuery(simInfoQuery(mapId))
  const knowledge = useQuery({ ...knowledgeQuery(mapId), enabled: tab === 'civilization' || tab === 'elements' })
  const data = info.data
  const spy = data?.secondsPerYear && data.secondsPerYear > 0 ? data.secondsPerYear : DEFAULT_SECONDS_PER_YEAR
  const inspectorRef = useRef<HTMLDivElement>(null)

  // A creature picked on the map or in the log opens the dock at its inspector.
  useEffect(() => {
    if (selectedId === null) return
    setCollapsed(false)
    inspectorRef.current?.scrollTo({ top: 0, behavior: 'smooth' })
  }, [selectedId])

  useEffect(() => localStorage.setItem(COLLAPSED_KEY, collapsed ? '1' : '0'), [collapsed])

  // Drag the top edge to resize; the height is remembered.
  const startResize = useCallback(
    (e: ReactPointerEvent<HTMLDivElement>) => {
      if (collapsed) return
      e.preventDefault()
      const startY = e.clientY
      const startH = height
      const handle = e.currentTarget
      handle.setPointerCapture(e.pointerId)
      const move = (ev: PointerEvent) => {
        setHeight(Math.min(maxHeight(), Math.max(MIN_HEIGHT, startH + startY - ev.clientY)))
      }
      const up = (ev: PointerEvent) => {
        handle.releasePointerCapture(ev.pointerId)
        handle.removeEventListener('pointermove', move)
        handle.removeEventListener('pointerup', up)
        localStorage.setItem(HEIGHT_KEY, String(heightRef.current))
      }
      handle.addEventListener('pointermove', move)
      handle.addEventListener('pointerup', up)
    },
    [collapsed, height],
  )

  return (
    <SecondsPerYearContext.Provider value={spy}>
      <section
        className={`observer-panel obs-dock ${collapsed ? 'collapsed' : ''}`}
        style={collapsed ? undefined : { height }}
        aria-label="Panel pengamat"
      >
        <div
          className="dock-resize"
          role="separator"
          aria-orientation="horizontal"
          aria-label="Ubah tinggi panel"
          title="Tarik untuk mengubah tinggi panel"
          onPointerDown={startResize}
        />
        <div className="dock-bar">
          <div className="obs-tabs" role="tablist" aria-label="Tampilan panel">
            {TABS.map((t) => (
              <button
                key={t.id}
                type="button"
                role="tab"
                id={`obs-tab-${t.id}`}
                aria-selected={tab === t.id}
                aria-controls={`obs-panel-${t.id}`}
                className={tab === t.id && !collapsed ? 'active' : ''}
                onClick={() => {
                  setTab(t.id)
                  setCollapsed(false)
                }}
              >
                {t.label}
              </button>
            ))}
          </div>
          {selectedId === null && (
            <span className="dock-hint small muted">
              Klik makhluk atau rumah di peta untuk melihat otak dan keluarganya. Mereka memutuskan sendiri — kamu hanya
              pengamat.
            </span>
          )}
          <button
            type="button"
            className="dock-toggle"
            aria-expanded={!collapsed}
            title={collapsed ? 'Buka panel' : 'Ciutkan panel'}
            onClick={() => setCollapsed((c) => !c)}
          >
            {collapsed ? '▴' : '▾'}
          </button>
        </div>

        {!collapsed && (
          <div className="dock-body">
            {selectedId !== null && (
              <div className="dock-inspector" ref={inspectorRef}>
                <Inspector
                  mapId={mapId}
                  id={selectedId}
                  following={following}
                  onSelect={onSelect}
                  onFollow={onFollow}
                />
              </div>
            )}
            <div
              className={`dock-content obs-tabpanel dock-${tab}`}
              role="tabpanel"
              id={`obs-panel-${tab}`}
              aria-labelledby={`obs-tab-${tab}`}
            >
              {info.isError && !data && (
                <p className="obs-error small">Simulasi belum tersedia: {info.error.message}</p>
              )}
              {tab === 'population' &&
                (data ? (
                  <>
                    <PopulationStats info={data} />
                    <PopulationChart history={data.history ?? []} />
                    <EventLog events={data.events ?? []} onSelect={onSelect} />
                    <Demography mapId={mapId} />
                  </>
                ) : (
                  info.isPending && <p className="small muted">Menghubungkan ke simulasi…</p>
                ))}
              {tab === 'ecology' && <Ecology mapId={mapId} />}
              {tab === 'civilization' && (
                <CivilizationView
                  knowledge={knowledge.data}
                  info={data}
                  error={knowledge.isError ? knowledge.error.message : null}
                  onSelect={onSelect}
                />
              )}
              {tab === 'elements' &&
                (knowledge.data ? (
                  <PeriodicTable knowledge={knowledge.data} onSelect={onSelect} />
                ) : (
                  <p className={`small ${knowledge.isError ? 'obs-error' : 'muted'}`}>
                    {knowledge.isError ? `Gagal memuat unsur: ${knowledge.error.message}` : 'Memuat tabel periodik…'}
                  </p>
                ))}
            </div>
          </div>
        )}
      </section>
    </SecondsPerYearContext.Provider>
  )
}
