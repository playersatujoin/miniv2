import { useSetSimSpeed } from '../../sim/api'
import { SIM_SPEEDS, type SimSpeed } from '../../sim/protocol'

type Props = { mapId: string; speed: SimSpeed | undefined }

/** Time control only — it never touches what the creatures decide. */
export function SpeedControl({ mapId, speed }: Props) {
  const setSpeed = useSetSimSpeed(mapId)
  const shown = setSpeed.isPending ? setSpeed.variables : speed

  return (
    <div className="obs-speed">
      <div className="segmented" role="group" aria-label="Kecepatan simulasi">
        {SIM_SPEEDS.map((s) => (
          <button
            key={s}
            type="button"
            className={shown === s ? 'active' : ''}
            aria-pressed={shown === s}
            disabled={speed === undefined}
            title={s === 0 ? 'Jeda' : `Kecepatan ${s}×`}
            onClick={() => setSpeed.mutate(s)}
          >
            {s === 0 ? '⏸' : `${s}×`}
          </button>
        ))}
      </div>
      {setSpeed.error && <p className="obs-error small">Gagal mengubah kecepatan: {setSpeed.error.message}</p>}
    </div>
  )
}
