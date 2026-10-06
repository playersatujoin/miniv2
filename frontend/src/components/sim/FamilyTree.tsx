import { queryOptions, useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { request } from '../../api/client'
import { SEX_SYMBOL, type FamilyPerson, type FamilyTree as Tree, type GeneticsView } from '../../sim/protocol'
import { SEX_COLOR } from './format'
import {
  ancestorRows,
  collapse,
  descendantRows,
  formatF,
  inbreedingWord,
  personTitle,
  relationWord,
  type TreeRow,
} from './familyTree'
import './FamilyTree.css'

const DEPTH_DEFAULT = 3
const DEPTH_MAX = 10

/** Someone's family tree, up and down (GET …/creatures/{id}/family?depth=N). */
export const familyQuery = (mapId: string, id: number, depth: number) =>
  queryOptions({
    queryKey: ['sim', mapId, 'family', id, depth],
    queryFn: () => request<Tree>(`/maps/${mapId}/sim/creatures/${id}/family?depth=${depth}`),
    refetchInterval: (query) => (query.state.status === 'error' ? false : 5000),
    retry: false,
  })

type Props = {
  mapId: string
  /** The person the Inspector shows. */
  id: number
  /** Opens a living person in the Inspector. */
  onSelect: (id: number) => void
}

/**
 * Pohon keluarga (Fase 3c): the person's genes (inbreeding, recessive
 * variants carried, disorders) and their ancestors and descendants, back to
 * Adam & Hawa when the depth allows. One ancestor in several places of a
 * generation is pedigree collapse, which is what inbreeding looks like.
 * Clicking someone dead shows their own tree; someone alive opens them.
 */
export function FamilyTree({ mapId, id, onSelect }: Props) {
  const [depth, setDepth] = useState(DEPTH_DEFAULT)
  const [rooted, setRooted] = useState<{ for: number; root: number } | null>(null)
  const root = rooted?.for === id ? rooted.root : id
  const q = useQuery(familyQuery(mapId, root, depth))

  if (q.isError) return <p className="small muted">Silsilah tidak tersedia.</p>
  const tree = q.data
  if (!tree) return <p className="small muted">Memuat silsilah…</p>

  const people = new Map<number, FamilyPerson>(tree.people.map((p) => [p.id, p]))
  const self = people.get(tree.root)
  const up = ancestorRows(tree)
  const down = descendantRows(tree)
  const { distinct, places } = collapse(tree)
  const reachedFounders = up.some((row) => row.people.some((s) => people.get(s.id)?.founder))
  const deeper = depth < DEPTH_MAX && (up.length === depth || down.length === depth)

  const open = (p: FamilyPerson | undefined) => {
    if (!p || p.forgotten) return
    if (p.alive) onSelect(p.id)
    else setRooted({ for: id, root: p.id })
  }

  return (
    <div className="ft">
      {tree.genetics && <GeneticsSummary g={tree.genetics} parentsRelated={tree.parentsRelated} />}
      {root !== id && (
        <p className="ft-rooted small">
          Silsilah {self?.name ?? 'leluhur'} (sudah wafat) ·{' '}
          <button type="button" className="obs-link" onClick={() => setRooted(null)}>
            kembali
          </button>
        </p>
      )}

      <div className="ft-rows" role="tree" aria-label={`Pohon keluarga ${self?.name ?? ''}`}>
        {[...up].reverse().map((row) => (
          <Row key={row.level} row={row} people={people} onOpen={open} />
        ))}
        <div className="ft-row ft-self">
          <span className="ft-level">{root === id ? 'Dia' : 'Leluhur ini'}</span>
          <div className="ft-people">
            <Chip person={self} count={1} onOpen={() => {}} self />
          </div>
        </div>
        {down.map((row) => (
          <Row key={row.level} row={row} people={people} onOpen={open} />
        ))}
      </div>

      <p className="small muted ft-foot">
        {up.length === 0
          ? self?.founder
            ? 'Manusia pertama: tidak punya orang tua.'
            : 'Orang tuanya tidak diketahui.'
          : `${distinct} leluhur berbeda mengisi ${places} tempat dalam ${up.length} generasi${
              distinct < places ? ' — sebagian leluhur muncul lebih dari sekali (perkawinan antar-kerabat)' : ''
            }.`}
        {reachedFounders && ' Sampai ke manusia pertama.'}
      </p>
      <div className="ft-depth">
        <span className="small muted">Kedalaman {depth} generasi</span>
        <button type="button" className="ghost" disabled={depth <= 1} onClick={() => setDepth(depth - 1)} aria-label="Kurangi generasi">
          −
        </button>
        <button type="button" className="ghost" disabled={!deeper} onClick={() => setDepth(depth + 1)} aria-label="Tambah generasi">
          +
        </button>
        {!reachedFounders && deeper && (
          <button type="button" className="ghost" onClick={() => setDepth(DEPTH_MAX)}>
            Sejauh mungkin
          </button>
        )}
      </div>
    </div>
  )
}

function GeneticsSummary({ g, parentsRelated }: { g: GeneticsView; parentsRelated: number }) {
  return (
    <dl className="obs-kv ft-genes">
      <dt title="Koefisien inbreeding: peluang dua salinan sebuah gen berasal dari leluhur yang sama">Inbreeding (F)</dt>
      <dd>
        {formatF(g.f)} <span className="muted small">· {inbreedingWord(g.f)}</span>
      </dd>
      {parentsRelated > 0 && (
        <>
          <dt>Kekerabatan orang tua</dt>
          <dd>
            r {formatF(parentsRelated)} <span className="muted small">· {relationWord(parentsRelated)}</span>
          </dd>
        </>
      )}
      <dt title="Varian resesif yang dibawa satu salinan: tidak menimbulkan gejala, tetapi bisa diwariskan">Pembawa</dt>
      <dd>
        {g.carried} varian resesif
        {!g.known && <span className="muted small"> · perkiraan (lahir sebelum genetika dicatat)</span>}
      </dd>
      <dt>Kelainan bawaan</dt>
      <dd className={g.defects.length ? 'ft-defect' : undefined}>{g.defects.length ? g.defects.join(', ') : 'tidak ada'}</dd>
      {g.malariaShield && (
        <>
          <dt>Talasemia</dt>
          <dd>pembawa · malaria lebih ringan</dd>
        </>
      )}
    </dl>
  )
}

function Row({
  row,
  people,
  onOpen,
}: {
  row: TreeRow
  people: Map<number, FamilyPerson>
  onOpen: (p: FamilyPerson | undefined) => void
}) {
  return (
    <div className={`ft-row ${row.level > 0 ? 'ft-up' : 'ft-down'}`} role="group" aria-label={row.label}>
      <span className="ft-level" title={`${row.places} orang`}>
        {row.label}
      </span>
      <div className="ft-people">
        {row.people.map((s) => (
          <Chip key={s.id} person={people.get(s.id)} count={s.count} onOpen={onOpen} />
        ))}
      </div>
    </div>
  )
}

function Chip({
  person,
  count,
  onOpen,
  self = false,
}: {
  person: FamilyPerson | undefined
  count: number
  onOpen: (p: FamilyPerson | undefined) => void
  self?: boolean
}) {
  const sex = person?.sex || undefined
  const dead = person && !person.alive
  const cls = ['ft-chip', self && 'ft-chip-self', dead && 'ft-dead', person?.forgotten && 'ft-forgotten', person?.founder && 'ft-founder']
    .filter(Boolean)
    .join(' ')
  const label = person?.forgotten ? person.name || '?' : (person?.name ?? '?')
  return (
    <button type="button" className={cls} title={personTitle(person)} onClick={() => onOpen(person)} disabled={self || person?.forgotten}>
      {sex && (
        <span className="ft-sex" style={{ color: SEX_COLOR[sex] }} aria-hidden>
          {SEX_SYMBOL[sex]}
        </span>
      )}
      <span className="ft-name">{label}</span>
      {dead && !person?.forgotten && <span aria-label="sudah wafat"> †</span>}
      {count > 1 && <span className="ft-count" title={`Muncul ${count} kali di generasi ini`}>×{count}</span>}
    </button>
  )
}
