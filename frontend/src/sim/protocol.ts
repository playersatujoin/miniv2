// Wire contract with backend/internal/sim. Keep in sync with the Go side.

export type Sex = 'female' | 'male'
export type Action =
  | 'explore'
  | 'eat'
  | 'drink'
  | 'mate'
  | 'rest'
  | 'gather'
  | 'craft'
  | 'build'
  | 'give'
  | 'steal'
  | 'attack'
export type DeathCause = 'starvation' | 'thirst' | 'oldAge' | 'killed'

/** Bit flags in CreatureFrame.flags. */
export const FLAG = {
  pregnant: 1,
  eating: 2,
  drinking: 4,
  wantsMate: 8,
  resting: 16,
  child: 32,
  attacking: 64,
  stealing: 128,
  giving: 256,
  building: 512,
  crafting: 1024,
  gathering: 2048,
  /** Kepala keluarga: owns a house. */
  head: 4096,
  /** Carrying a lot (inventory at least half full). */
  carrying: 8192,
} as const

/** One creature in a stream frame. Positions are in tiles, heading in radians (0 = +x, π/2 = +y/down). */
export type CreatureFrame = {
  id: number
  name: string
  x: number
  y: number
  heading: number
  sex: Sex
  /** Lineage colour, 0–359. Children inherit a blend of their parents' hue. */
  hue: number
  /** Body size multiplier, roughly 0.7–1.3. */
  size: number
  flags: number
  /** 0–1 */
  energy: number
  /** 0–1; attacks lower it, rest restores it. */
  health: number
  /** House the creature belongs to, 0 = none. */
  houseId: number
}

export type SimFrame = {
  tick: number
  /** Simulated seconds since the world began. */
  time: number
  creatures: CreatureFrame[]
}

type RawCreature = [
  id: number,
  name: string,
  x: number,
  y: number,
  heading: number,
  sex: 0 | 1,
  hue: number,
  size: number,
  flags: number,
  energy: number,
  health: number,
  houseId: number,
]
type RawFrame = { t: number; s: number; c: RawCreature[] }

/** Parses the compact `frame` SSE payload. */
export function parseFrame(data: string): SimFrame {
  const raw = JSON.parse(data) as RawFrame
  return {
    tick: raw.t,
    time: raw.s,
    creatures: raw.c.map(([id, name, x, y, heading, sex, hue, size, flags, energy, health, houseId]) => ({
      id,
      name,
      x,
      y,
      heading,
      sex: sex === 1 ? 'male' : 'female',
      hue,
      size,
      flags,
      energy,
      health,
      houseId,
    })),
  }
}

/** A building on the map. It occupies tile (x, y); creatures can walk over it. */
export type StructureFrame = {
  id: number
  /** Structure kind id from the knowledge endpoint, e.g. "gubuk", "rumah_kayu", "ladang", "tungku". */
  kind: string
  x: number
  y: number
  /** Kepala keluarga / builder; 0 = abandoned (no living owner). */
  ownerId: number
  ownerName: string
  /** House level 1–3; 0 for non-houses. */
  level: number
  /** Owner family's colour (0–359) for roofs/flags. */
  hue: number
}

export type StructuresMessage = { version: number; structures: StructureFrame[] }

type RawStructure = [id: number, kind: string, x: number, y: number, ownerId: number, ownerName: string, level: number, hue: number]

/** Parses the `structures` SSE payload (sent on connect and whenever buildings change). */
export function parseStructures(data: string): StructuresMessage {
  const raw = JSON.parse(data) as { v: number; s: RawStructure[] }
  return {
    version: raw.v,
    structures: raw.s.map(([id, kind, x, y, ownerId, ownerName, level, hue]) => ({
      id,
      kind,
      x,
      y,
      ownerId,
      ownerName,
      level,
      hue,
    })),
  }
}

export type SimEventKind =
  | 'birth'
  | 'death'
  | 'genesis'
  | 'milestone'
  | 'discovery'
  | 'build'
  | 'crime'
  | 'kindness'
  | 'family'

export type SimEvent = {
  id: number
  time: number
  kind: SimEventKind
  /** Human-readable Indonesian text, built by the backend. */
  text: string
  creatureId?: number
}

export type SimHistoryPoint = {
  time: number
  population: number
  females: number
  males: number
  avgGeneration: number
  maxGeneration: number
  /** Elements discovered so far. */
  elements: number
  houses: number
}

export type SimInfo = {
  mapId: string
  tick: number
  time: number
  /** Simulated seconds per simulated year (1 year = 8 s). Ages and rates in the UI are shown in years. */
  secondsPerYear: number
  /** Current simulated year, floor(time / secondsPerYear) + 1. */
  year: number
  speed: SimSpeed
  /** How many times humanity has started over from an Adam and Hawa (1 = the first). */
  era: number
  /** Civilisation tier 0–7 and its name, e.g. "Zaman Batu". */
  tier: number
  tierName: string
  population: number
  females: number
  males: number
  /** Population cap derived from the map's walkable area. */
  capacity: number
  births: number
  deaths: number
  deathsByCause: Record<DeathCause, number>
  maxGeneration: number
  avgGeneration: number
  elementsDiscovered: number
  elementsTotal: number
  houses: number
  structures: number
  /** All-time counts of deeds. */
  crimes: number
  kindness: number
  kills: number
  /** Sampled every 5 simulated seconds, oldest first, at most 720 points. */
  history: SimHistoryPoint[]
  /** Newest last, at most 80. */
  events: SimEvent[]
}

export type CreatureRef = { id: number; name: string }

export type Traits = {
  hue: number
  size: number
  /** tiles per second */
  maxSpeed: number
  /** tiles */
  vision: number
  metabolism: number
  mutationRate: number
}

export type Brain = {
  inputLabels: string[] // 55
  outputLabels: string[] // 12
  input: number[] // current activations
  hidden: number[] // tanh
  output: number[] // turn is tanh (-1..1), the rest sigmoid (0..1)
  wIn: number[][] // [input][hidden]
  wRec: number[][] // [hiddenFrom][hiddenTo], previous hidden state -> next
  wOut: number[][] // [hidden][output]
  bOut: number[] // [output]
}

export type Stack = { item: string; name: string; qty: number }

export type HouseDetail = {
  id: number
  kind: string
  name: string
  level: number
  x: number
  y: number
  head: CreatureRef | null
  members: number
  storage: Stack[]
  capacity: number
}

export type Deeds = {
  kindness: number
  crimes: number
  kills: number
  built: number
  crafted: number
  discoveries: number
}

export type CreatureDetail = {
  /** Simulated seconds per simulated year, to show age/lifespan in years. */
  secondsPerYear: number
  id: number
  name: string
  sex: Sex
  generation: number
  /** null for Adam and Hawa. */
  mother: CreatureRef | null
  father: CreatureRef | null
  spouse: CreatureRef | null
  /** Simulated seconds. */
  bornAt: number
  age: number
  lifespan: number
  adult: boolean
  energy: number
  hydration: number
  health: number
  /** -1 (feared/hated) .. 1 (trusted/loved), from their deeds. */
  reputation: number
  /** head = kepala keluarga (owns a house), member = lives in a family house, none = homeless. */
  role: 'head' | 'member' | 'none'
  house: HouseDetail | null
  inventory: Stack[]
  deeds: Deeds
  pregnant: boolean
  /** Pregnancy progress 0–1 (0 when not pregnant). */
  gestation: number
  children: number
  x: number
  y: number
  action: Action
  traits: Traits
  brain: Brain
}

export type ElementInfo = {
  z: number
  symbol: string
  name: string
  category: string
  period: number
  /** 1–18; 0 for lanthanides (57–71) and actinides (89–103). */
  group: number
  phase: 'padat' | 'cair' | 'gas'
  /** ppm in Earth's crust; 0 for synthetic. */
  abundance: number
  natural: boolean
  tier: number
  discovered: boolean
  discoveredAt?: number
  discoveredBy?: CreatureRef
  /** Era in which it was discovered. */
  era?: number
  /** Item it was isolated from, e.g. "Hematit", or "Sintesis". */
  source?: string
}

export type TechInfo = {
  id: string
  name: string
  description: string
  tier: number
  requires: string[]
  known: boolean
  learnedAt?: number
  learnedBy?: CreatureRef
}

export type StructureKindInfo = {
  id: string
  name: string
  house: boolean
  level: number
  tier: number
  cost: Stack[]
  /** How many exist right now. */
  count: number
}

export type Knowledge = {
  era: number
  tier: number
  tierName: string
  /** Names of tiers 0..7. */
  tierNames: string[]
  elements: ElementInfo[]
  techs: TechInfo[]
  structureKinds: StructureKindInfo[]
}

/** One 5-year age band of the living population (the last band is open-ended, e.g. "80+"). */
export type AgeBand = { label: string; from: number; to: number | null; female: number; male: number }

/** A pre-modern reference range to compare a metric against, with its source. */
export type MetricRef = { low: number; high: number; source: string; note?: string }

/**
 * Demographic indicators over a rolling window of recent simulated years,
 * computed with a period life table (like demographers do). null = not enough data yet.
 */
export type DemographyMetrics = {
  /** Years of simulated time the window covers. */
  windowYears: number
  /** Person-years of exposure in the window. */
  personYears: number
  /** e0: life expectancy at birth, years. */
  lifeExpectancy: number | null
  /** e15: remaining life expectancy at age 15, years. */
  lifeExpectancy15: number | null
  /** l15: probability a newborn survives to age 15, 0–1. */
  survivalTo15: number | null
  /** q0: probability of dying before age 1, 0–1. */
  infantMortality: number | null
  /** Most common age at death among adults (15+), years. */
  modalAgeAdultDeath: number | null
  /** Total fertility rate: children per woman over a lifetime at current rates. */
  tfr: number | null
  meanAgeFirstBirth: number | null
  /** Mean years between a mother's consecutive births. */
  meanBirthInterval: number | null
  /** Living males per 100 females. */
  sexRatio: number | null
  /** Mean members per inhabited house. */
  householdSize: number | null
  /** Killings per 100,000 person-years. */
  homicideRate: number | null
  /** Wealth inequality among adults (inventory + share of house storage), 0–1. */
  gini: number | null
  /** Deaths in the window by cause. */
  deathsByCause: Record<DeathCause, number>
}

export type DemographyPoint = {
  year: number
  population: number
  lifeExpectancy: number | null
  survivalTo15: number | null
  tfr: number | null
  gini: number | null
  homicideRate: number | null
}

export type Demography = {
  secondsPerYear: number
  year: number
  /** Living population by 5-year band: 0–4, 5–9, …, 80+. */
  pyramid: AgeBand[]
  current: DemographyMetrics
  /** Sampled every 10 simulated years, oldest first, at most 400 points. */
  history: DemographyPoint[]
  /** Pre-modern reference ranges keyed by metric name (only metrics that have one). */
  reference: Partial<Record<keyof DemographyMetrics, MetricRef>>
}

export const SIM_SPEEDS = [0, 1, 2, 5, 10, 20] as const
export type SimSpeed = (typeof SIM_SPEEDS)[number]

export const ACTION_LABELS: Record<Action, string> = {
  explore: 'Menjelajah',
  eat: 'Makan',
  drink: 'Minum',
  mate: 'Mencari pasangan',
  rest: 'Istirahat',
  gather: 'Mengumpulkan',
  craft: 'Membuat',
  build: 'Membangun',
  give: 'Berbagi',
  steal: 'Mencuri',
  attack: 'Menyerang',
}

export const DEATH_LABELS: Record<DeathCause, string> = {
  starvation: 'Kelaparan',
  thirst: 'Kehausan',
  oldAge: 'Usia tua',
  killed: 'Dibunuh',
}

export const SEX_SYMBOL: Record<Sex, string> = { female: '♀', male: '♂' }

export const ROLE_LABELS: Record<CreatureDetail['role'], string> = {
  head: 'Kepala keluarga',
  member: 'Anggota keluarga',
  none: 'Tanpa rumah',
}
