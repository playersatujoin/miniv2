import { useQuery } from '@tanstack/react-query'
import { demographyQuery } from '../../sim/api'
import type { Demography, GeneticsInfo } from '../../sim/protocol'
import { formatF } from './familyTree'
import { nf, pct } from './format'

/**
 * Inbreeding on the island (Fase 3c), for the Demography tab: how inbred the
 * living are, inbreeding by generation (high after Adam & Hawa, falling as
 * the population grows), what became of inbred children, and whether adults
 * want parents, children and siblings as mates less than others.
 */
export function InbreedingStats({ mapId }: { mapId: string }) {
  const q = useQuery(demographyQuery(mapId))
  const g = (q.data as (Demography & { genetics?: GeneticsInfo }) | undefined)?.genetics
  if (!g) return null
  const gens = g.byGeneration.filter((b) => b.births > 0)
  const bands = g.byF.filter((b) => b.births > 0)
  const periods = g.periods.slice(-6)

  return (
    <section className="obs-section">
      <h4 className="obs-subtitle">Genetika & perkawinan sedarah</h4>
      <dl className="obs-kv">
        <dt title="Koefisien inbreeding rata-rata penduduk yang hidup">F rata-rata</dt>
        <dd>{formatF(g.meanF)}</dd>
        <dt>Anak sepupu dua kali atau lebih dekat</dt>
        <dd>{pct(g.inbred)}</dd>
        <dt>Anak kerabat dekat (F ≥ 1/8)</dt>
        <dd>{pct(g.closeKin)}</dd>
        <dt>Varian resesif dibawa per orang</dt>
        <dd>{g.carried.toFixed(2).replace('.', ',')}</dd>
        <dt>Punya kelainan bawaan</dt>
        <dd>{pct(g.affected)}</dd>
        <dt>Bayi wafat karena kelainan resesif</dt>
        <dd>{nf.format(g.lethal)}</dd>
      </dl>

      {gens.length > 0 && (
        <table className="obs-table">
          <caption className="small muted">Kelahiran menurut generasi anak</caption>
          <thead>
            <tr>
              <th>Generasi</th>
              <th>Lahir</th>
              <th>F rata-rata</th>
              <th>Anak kerabat dekat</th>
            </tr>
          </thead>
          <tbody>
            {gens.map((b) => (
              <tr key={b.label}>
                <td>{b.label}</td>
                <td>{nf.format(b.births)}</td>
                <td>{formatF(b.meanF)}</td>
                <td>{pct(b.closeKin)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {bands.length > 0 && (
        <table className="obs-table">
          <caption className="small muted">Nasib anak menurut F orang tuanya</caption>
          <thead>
            <tr>
              <th>F</th>
              <th>Lahir</th>
              <th>Berkelainan</th>
              <th>Wafat &lt; 15 th</th>
            </tr>
          </thead>
          <tbody>
            {bands.map((b) => (
              <tr key={b.label}>
                <td>{b.label}</td>
                <td>{nf.format(b.births)}</td>
                <td>{pct(b.affected / b.births)}</td>
                <td title="Anak yang masih kecil belum terhitung, jadi ini batas bawah">{b.q15 == null ? '–' : pct(b.q15)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {periods.length > 0 && (
        <table className="obs-table">
          <caption className="small muted">
            Keinginan kawin saat melihat calon: orang tua, anak, atau saudara dibanding orang lain (per 25 tahun)
          </caption>
          <thead>
            <tr>
              <th>Tahun</th>
              <th>Keluarga inti</th>
              <th>Orang lain</th>
              <th title="Rata-rata koefisien kekerabatan dengan calon lain">r lain</th>
            </tr>
          </thead>
          <tbody>
            {periods.map((p) => (
              <tr key={p.year}>
                <td>
                  {nf.format(p.year + 1)}–{nf.format(p.year + 25)}
                </td>
                <td>{p.kinDesire == null ? '–' : pct(p.kinDesire)}</td>
                <td>{p.otherDesire == null ? '–' : pct(p.otherDesire)}</td>
                <td>{p.otherSeen ? formatF(p.otherR) : '–'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      <p className="obs-note small muted">
        Tidak ada aturan yang melarang kawin dengan kerabat. Otak hanya merasakan seberapa dekat kekerabatan dan seberapa sehat calon
        pasangan terlihat; penghindaran (efek Westermarck) hanya bisa muncul lewat evolusi.
      </p>
    </section>
  )
}
