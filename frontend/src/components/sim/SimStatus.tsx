import { useQuery } from '@tanstack/react-query'
import { simInfoQuery } from '../../sim/api'
import { MONTH_NAMES } from '../../sim/protocol'
import { SeasonBadges } from './Ecology'
import { SpeedControl } from './SpeedControl'
import { DEFAULT_SECONDS_PER_YEAR, nf, yearAt } from './format'
import './sim.css'

/** The world's clock and the time control, for the map toolbar while watching. */
export function SimStatus({ mapId }: { mapId: string }) {
  const info = useQuery(simInfoQuery(mapId))
  const data = info.data
  const spy = data?.secondsPerYear && data.secondsPerYear > 0 ? data.secondsPerYear : DEFAULT_SECONDS_PER_YEAR
  const month = data ? MONTH_NAMES[Math.floor(((Math.max(0, data.time) / spy) % 1) * 12)] : undefined

  return (
    <div className="sim-status">
      <span className="sim-clock">
        {data?.era != null && <span className="obs-era">Era {nf.format(data.era)}</span>}
        <strong>{data ? `Tahun ${nf.format(data.year ?? yearAt(data.time, spy))}` : '—'}</strong>
        {month && <span className="sim-month muted">{month}</span>}
        <SeasonBadges season={data?.season} enso={data?.enso} title={month && `Bulan ${month}`} />
        {data?.speed === 0 && <span className="obs-paused">dijeda</span>}
      </span>
      {data?.tierName && (
        <span
          className="obs-tier-badge"
          title={
            data.elementsDiscovered != null
              ? `${nf.format(data.elementsDiscovered)}/${nf.format(data.elementsTotal ?? 118)} unsur ditemukan`
              : undefined
          }
        >
          {data.tierName}
        </span>
      )}
      <SpeedControl mapId={mapId} speed={data?.speed} />
    </div>
  )
}
