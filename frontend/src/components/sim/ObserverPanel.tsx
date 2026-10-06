import { useQuery } from '@tanstack/react-query'
import { useEffect, useRef, useState } from 'react'
import { knowledgeQuery, simInfoQuery } from '../../sim/api'
import { CivilizationView } from './CivilizationView'
import { Demography } from './Demography'
import { Ecology, SeasonBadges } from './Ecology'
import { EventLog } from './EventLog'
import { Inspector } from './Inspector'
import { PeriodicTable } from './PeriodicTable'
import { PopulationChart } from './PopulationChart'
import { PopulationStats } from './PopulationStats'
import { SpeedControl } from './SpeedControl'
import { MONTH_NAMES } from '../../sim/protocol'
import { DEFAULT_SECONDS_PER_YEAR, SecondsPerYearContext, nf, yearAt } from './format'
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

/** Read-only window into the simulation: the observer watches, never steers. */
export function ObserverPanel({ mapId, selectedId, following, onSelect, onFollow }: ObserverPanelProps) {
  const [tab, setTab] = useState<Tab>('population')
  const info = useQuery(simInfoQuery(mapId))
  const knowledge = useQuery({ ...knowledgeQuery(mapId), enabled: tab === 'civilization' || tab === 'elements' })
  const data = info.data
  const spy = data?.secondsPerYear && data.secondsPerYear > 0 ? data.secondsPerYear : DEFAULT_SECONDS_PER_YEAR
  const month = data ? MONTH_NAMES[Math.floor(((Math.max(0, data.time) / spy) % 1) * 12)] : undefined
  const ref = useRef<HTMLElement>(null)

  // Bring the inspector into view when a creature gets picked (e.g. from the event log).
  useEffect(() => {
    if (selectedId !== null) ref.current?.scrollTo({ top: 0, behavior: 'smooth' })
  }, [selectedId])

  return (
    <SecondsPerYearContext.Provider value={spy}>
    <aside ref={ref} className={`observer-panel ${tab === 'elements' ? 'obs-wide' : ''}`}>
      <header className="obs-header">
        <div className="obs-header-row">
          <h2>
            Akuarium
            {data?.era != null && <span className="obs-era">Era {nf.format(data.era)}</span>}
          </h2>
          <span className="obs-clock">
            {data ? `Tahun ${nf.format(data.year ?? yearAt(data.time, spy))}` : '—'}
            <SeasonBadges season={data?.season} enso={data?.enso} title={month && `Bulan ${month}`} />
            {data?.speed === 0 && <span className="obs-paused">dijeda</span>}
          </span>
        </div>
        {data?.tierName && (
          <p className="obs-tier small">
            <span className="obs-tier-badge">{data.tierName}</span>
            {data.elementsDiscovered != null && (
              <span className="muted">
                {nf.format(data.elementsDiscovered)}/{nf.format(data.elementsTotal ?? 118)} unsur
              </span>
            )}
          </p>
        )}
        <SpeedControl mapId={mapId} speed={data?.speed} />
      </header>

      {info.isError && !data && (
        <p className="obs-error small">Simulasi belum tersedia: {info.error.message}</p>
      )}

      {selectedId !== null ? (
        <Inspector mapId={mapId} id={selectedId} following={following} onSelect={onSelect} onFollow={onFollow} />
      ) : (
        <p className="obs-hint small">
          Klik salah satu makhluk atau rumah di peta untuk melihat isi otak dan keluarganya. Mereka memutuskan sendiri —
          kamu hanya pengamat.
        </p>
      )}

      <div className="obs-tabs" role="tablist" aria-label="Tampilan panel">
        {TABS.map((t) => (
          <button
            key={t.id}
            type="button"
            role="tab"
            id={`obs-tab-${t.id}`}
            aria-selected={tab === t.id}
            aria-controls={`obs-panel-${t.id}`}
            className={tab === t.id ? 'active' : ''}
            onClick={() => setTab(t.id)}
          >
            {t.label}
          </button>
        ))}
      </div>

      <div className="obs-tabpanel" role="tabpanel" id={`obs-panel-${tab}`} aria-labelledby={`obs-tab-${tab}`}>
        {tab === 'population' && (
          <>
            {data ? (
              <>
                <PopulationStats info={data} />
                <PopulationChart history={data.history ?? []} />
                <Demography mapId={mapId} />
              </>
            ) : (
              info.isPending && <p className="small muted">Menghubungkan ke simulasi…</p>
            )}
            <EventLog events={data?.events ?? []} onSelect={onSelect} />
          </>
        )}
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
    </aside>
    </SecondsPerYearContext.Provider>
  )
}
