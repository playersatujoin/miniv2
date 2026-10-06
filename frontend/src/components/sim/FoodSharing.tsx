import type { EngineInfo, SharingInfo, VillageStores as VillageStoresInfo } from '../../sim/protocol'
import { nf } from './format'

const n = (v: number | undefined | null) => (v == null ? '—' : nf.format(v))

function Stat({ label, value, title }: { label: string; value: string; title?: string }) {
  return (
    <div className="obs-stat" title={title}>
      <span className="obs-stat-label">{label}</span>
      <strong>{value}</strong>
    </div>
  )
}

/** The island's food held and shared (SimInfo.engine.sharing); renders nothing for a server that doesn't report it. */
export function FoodSharing({ engine }: { engine: EngineInfo | undefined }) {
  const s = (engine as (EngineInfo & { sharing?: SharingInfo }) | undefined)?.sharing
  if (!s) return null
  const allTime = [
    s.fromKin ? `${nf.format(s.fromKin)} kali makan dari bekal keluarga` : '',
    s.fromOthers ? `${nf.format(s.fromOthers)} kali dari bekal tetangga yang berlimpah` : '',
    s.bawon ? `${nf.format(s.bawon)} satuan panen bawon masuk lumbung pemiliknya (${nf.format(s.bawonShare)} dibawa pulang pemanen)` : '',
    s.granaryMeals ? `${nf.format(s.granaryMeals)} bekal diambil dari lumbung desa` : '',
    s.meatLeft ? `${nf.format(s.meatLeft)} daging buruan ditinggal di tempat (${nf.format(s.meatTaken)} dimakan, ${nf.format(s.meatRotted)} busuk)` : '',
  ].filter(Boolean)
  return (
    <section className="obs-section" aria-label="Pangan bersama">
      <h3 className="obs-title">Pangan bersama</h3>
      <div className="obs-stat-grid">
        <Stat label="🎒 Dibawa" value={n(s.carriedFood)} title="Satuan makanan yang dibawa orang sekarang" />
        <Stat label="🏠 Di rumah" value={n(s.houseFood)} title="Satuan makanan di simpanan rumah" />
        <Stat
          label="🌾 Di lumbung"
          value={n(s.granaryFood)}
          title={`Satuan makanan di ${nf.format(s.granaries)} lumbung; warga desa yang lapar boleh mengambilnya`}
        />
        <Stat
          label="🍖 Bangkai buruan"
          value={n(s.carcasses)}
          title={`${nf.format(s.carcassMeat)} satuan daging tergeletak di tempat buruan, boleh diambil siapa saja`}
        />
      </div>
      {allTime.length > 0 && <p className="small muted obs-note">Sepanjang masa: {allTime.join(' · ')}.</p>}
    </section>
  )
}

/** One village's stores, for its card in the Desa tab (VillageView.houseFood / granaryFood). */
export function VillageStores({ v }: { v: VillageStoresInfo }) {
  if (v.houseFood == null && v.granaryFood == null) return null
  return (
    <p className="small muted" title="Satuan makanan yang tersimpan di rumah-rumah dan lumbung desa ini">
      Simpanan pangan: lumbung <strong>{n(v.granaryFood)}</strong> · rumah <strong>{n(v.houseFood)}</strong>
    </p>
  )
}
