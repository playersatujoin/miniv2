/** What the automatic camera is looking at (World3DEvents.onCinematicSubject). */
export type CinematicSubject = { id: number | null; label: string }

type Props = {
  /** Only the 3D view has a director; in 2D the button explains why it can't be used. */
  available: boolean
  on: boolean
  onToggle: () => void
}

/**
 * "Kamera otomatis": hands the 3D camera to the world's director, which turns
 * to whatever is happening. Watching only — the camera never touches anyone.
 * Stays focusable while unavailable (aria-disabled), so its tooltip can be read.
 */
export function CinematicButton({ available, on, onToggle }: Props) {
  const title = !available
    ? 'Kamera otomatis hanya ada di tampilan 3D'
    : on
      ? 'Matikan kamera otomatis (menggeser kamera juga mematikannya)'
      : 'Kamera otomatis: kamera berpindah sendiri ke kejadian yang menarik'
  return (
    <button
      type="button"
      className={`geo-toggle cine-toggle ${on && available ? 'active' : ''}`}
      aria-pressed={on && available}
      aria-disabled={!available}
      aria-label="Kamera otomatis"
      title={title}
      onClick={() => available && onToggle()}
    >
      🎥 <span className="btn-label">Kamera otomatis</span>
    </button>
  )
}

/** The automatic camera's current subject over the stage; the person (if any) can be opened in the Inspector. */
export function CinematicSubjectChip({ subject, onSelect }: { subject: CinematicSubject | null; onSelect: (id: number) => void }) {
  return (
    <div className="cine-subject" role="status" aria-live="polite">
      <span aria-hidden="true">🎥</span> Kamera otomatis
      {subject ? (
        <>
          {' · '}
          {subject.id != null && subject.id > 0 ? (
            <button type="button" className="cine-subject-link" title="Amati orang ini" onClick={() => onSelect(subject.id!)}>
              {subject.label}
            </button>
          ) : (
            <strong>{subject.label}</strong>
          )}
        </>
      ) : (
        <span className="muted"> · mencari kejadian…</span>
      )}
    </div>
  )
}
