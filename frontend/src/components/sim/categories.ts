// Periodic-table category colours. Validated with the dataviz validator
// (dark mode, surface #172234, adjacent pairs in this order): all checks pass.
// The order follows how the groups touch on the table (alkali → … → noble
// gases), so every neighbouring pair on the grid is also an adjacent pair here;
// the off-chain neighbours (H above Li) were checked separately.
export type ElementGroup = { id: string; label: string; color: string }

export const ELEMENT_GROUPS: ElementGroup[] = [
  { id: 'alkali', label: 'Logam alkali', color: '#3987e5' },
  { id: 'alkaline', label: 'Logam alkali tanah', color: '#d95926' },
  { id: 'transition', label: 'Logam transisi', color: '#199e70' },
  { id: 'post', label: 'Logam pasca-transisi', color: '#c98500' },
  { id: 'metalloid', label: 'Metaloid', color: '#d55181' },
  { id: 'nonmetal', label: 'Nonlogam & halogen', color: '#008300' },
  { id: 'noble', label: 'Gas mulia', color: '#9085e9' },
  { id: 'fblock', label: 'Lantanida & aktinida', color: '#e66767' },
]

const OTHER: ElementGroup = { id: 'other', label: 'Lainnya', color: '#898781' }

const byId = new Map(ELEMENT_GROUPS.map((g) => [g.id, g]))

/** Maps the backend's free-text category (Indonesian or English) to a colour group. */
export function groupOf(category: string | undefined): ElementGroup {
  const c = (category ?? '').toLowerCase()
  const pick = (id: string) => byId.get(id)!
  if (/lantan|aktin|actin/.test(c)) return pick('fblock')
  if (/alkali tanah|alkaline/.test(c)) return pick('alkaline')
  if (/alkali/.test(c)) return pick('alkali')
  if (/pasca|post/.test(c)) return pick('post')
  if (/transisi|transition/.test(c)) return pick('transition')
  if (/metaloid|metalloid/.test(c)) return pick('metalloid')
  if (/mulia|noble/.test(c)) return pick('noble')
  if (/halogen|nonlogam|non-logam|nonmetal/.test(c)) return pick('nonmetal')
  return OTHER
}
