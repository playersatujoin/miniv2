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
  | 'teach'
  | 'plant'
  | 'hunt'
  | 'harvest'
  | 'fish'
/** `animal`: killed by a wild animal (a tiger, a charging boar or buffalo). */
export type DeathCause = 'starvation' | 'thirst' | 'oldAge' | 'killed' | 'animal'

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
  /** Teaching someone (or writing at a library) this second. */
  teaching: 16384,
  /** Planting a crop this second. */
  planting: 32768,
  /** Hunting an animal this second. */
  hunting: 65536,
} as const

/** Bit flags in AnimalFrame.flags. */
export const ANIMAL_FLAG = {
  /** Fleeing, or charging at someone. */
  running: 1,
  eating: 2,
  young: 4,
  /** Household livestock. */
  tame: 8,
  hurt: 16,
  /** A tiger stalking its prey. */
  hunting: 32,
} as const

/** Bit flags in FieldPlot.flags. */
export const PLOT_FLAG = {
  /** Wilting in parched soil. */
  withered: 1,
  /** Watered by an irrigation channel (sawah). */
  irrigated: 2,
  /** Inside a ladang: cleared and weeded. */
  farmland: 4,
  /** Near a pen with livestock. */
  manured: 8,
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

/** A wild or domestic animal in a stream frame. Positions in tiles, heading in radians. */
export type AnimalFrame = {
  id: number
  /** Index into SPECIES. */
  species: number
  x: number
  y: number
  heading: number
  flags: number
}

export type Enso = 'netral' | 'el_nino' | 'la_nina'
export type Season = 'hujan' | 'kemarau'

/** The island's weather in a stream frame. */
export type Weather = {
  /** Fraction of the simulated year, 0 = 1 January. */
  phase: number
  /** Abstract day–night rhythm, 0–1 (period 2 simulated seconds). */
  light: number
  /** Soil moisture, 0–1. */
  moisture: number
  /** -1 La Niña, 0 netral, 1 El Niño. */
  enso: -1 | 0 | 1
  /** Rainfall relative to the yearly average (1). */
  rain: number
}

export type SimFrame = {
  tick: number
  /** Simulated seconds since the world began. */
  time: number
  creatures: CreatureFrame[]
  animals: AnimalFrame[]
  /** Missing from older servers. */
  weather: Weather | null
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
type RawAnimal = [id: number, species: number, x: number, y: number, heading: number, flags: number]
type RawWeather = [phase: number, light: number, moisture: number, enso: number, rain: number]
type RawFrame = { t: number; s: number; w?: RawWeather; a?: RawAnimal[]; c: RawCreature[] }

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
    animals: (raw.a ?? []).map(([id, species, x, y, heading, flags]) => ({ id, species, x, y, heading, flags })),
    weather: raw.w
      ? {
          phase: raw.w[0],
          light: raw.w[1],
          moisture: raw.w[2],
          enso: Math.sign(raw.w[3]) as Weather['enso'],
          rain: raw.w[4],
        }
      : null,
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

/** One planted tile. */
export type FieldPlot = {
  x: number
  y: number
  /** Index into CROPS. */
  crop: number
  /** 0 seedling, 1 young, 2 grown, 3 ripe. */
  stage: 0 | 1 | 2 | 3
  flags: number
}

export type FieldsMessage = { version: number; plots: FieldPlot[] }

type RawPlot = [x: number, y: number, crop: number, stage: number, flags: number]

/** Parses the `fields` SSE payload (sent on connect and when plots change, at most once a second). */
export function parseFields(data: string): FieldsMessage {
  const raw = JSON.parse(data) as { v: number; p: RawPlot[] }
  return {
    version: raw.v,
    plots: raw.p.map(([x, y, crop, stage, flags]) => ({
      x,
      y,
      crop,
      stage: Math.min(3, Math.max(0, stage)) as FieldPlot['stage'],
      flags,
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
  /** Learned a skill from someone, wrote it down, or knowledge was lost. */
  | 'learning'
  /** El Niño / La Niña onset, floods, crops lost to drought. */
  | 'climate'
  /** A species dies out on the island, or a pair arrives from overseas. */
  | 'ecology'
  /** Planting and harvesting. */
  | 'farming'
  /** Hunting, taming and slaughtering animals. */
  | 'hunt'

export type SimEvent = {
  id: number
  time: number
  kind: SimEventKind
  /** Human-readable Indonesian text, built by the backend. */
  text: string
  creatureId?: number
  /** Assaults and killings, so observers can hide violence in the log. */
  violent?: boolean
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
  /** Mean number of hidden neurons among the living. */
  avgBrainSize: number
  /** Mean of each adult's best skill level, 0–1. */
  avgSkill: number
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
  /**
   * Technical ceiling only (walkable area / 6, 300–2000), not a carrying
   * capacity: food limits the population now.
   */
  capacity: number
  /** Conceptions the technical ceiling stopped; 0 in a healthy world. */
  capacityHits: number
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
  /** Rates per simulated year over the last 50 years (easier to read than totals). */
  crimesPerYear: number
  kindnessPerYear: number
  killsPerYear: number
  /** Mean number of hidden neurons among the living (brain size evolves). */
  avgBrainSize: number
  /** How many times a skill died out with its last holder (all-time). */
  knowledgeLost: number
  season: Season
  enso: Enso
  /** Wild animals alive on the island. */
  animals: number
  /** Domestic animals kept by households. */
  livestock: number
  /** Planted tiles. */
  plots: number
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
  inputLabels: string[] // grows with new senses, ~60
  outputLabels: string[] // ~13 (adds "ajar")
  /** Number of hidden neurons; inherited and evolving (≈ 8–64). */
  size: number
  /** Inherited learning rate (how fast experience changes this brain), 0 = none. */
  learningRate: number
  /** Per hidden neuron: how much its weights changed through lifetime learning, 0–1 (relative). */
  learned: number[]
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
  /** The family granary's food; [] if it has none. */
  granary: Stack[]
  /** Livestock by kind: item = species id, name = domestic name ("Ayam"), qty = head. */
  livestock: Stack[]
}

/** A learned skill (practical know-how of one technology), 0–1; ≥ 0.3 can practise it. */
export type Skill = { tech: string; name: string; level: number }

export type Deeds = {
  kindness: number
  crimes: number
  kills: number
  built: number
  crafted: number
  discoveries: number
  planted: number
  harvested: number
  hunted: number
  tamed: number
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
  /** Skills it knows, strongest first. */
  skills: Skill[]
  /** Who it is currently learning from (teacher or library holder), if anyone. */
  teacher: CreatureRef | null
  /** How many different people it has taught (all-time). */
  taught: number
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
  /** Living people who can practise it (skill ≥ 0.3). */
  holders: number
  /** It was known once but every holder died and nothing was written down. */
  lost: boolean
  /** Stored in at least one library, so it can be relearned by reading. */
  written: boolean
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

/** Deposits that have been mined out (old pits) and trees felled to stumps. */
export type MinedOut = {
  version: number
  tiles: [x: number, y: number][]
  /** Trees cut down to stumps; their tile still holds a tree object in the map. Missing from older servers. */
  logged?: [x: number, y: number][]
}

// --- Ecology -----------------------------------------------------------------

export type Diet = 'grazer' | 'forager' | 'predator'

export type SpeciesView = {
  id: string
  name: string
  /** Its name once tamed ("Ayam"); absent if it can't be tamed. */
  domestic?: string
  diet: Diet
  wild: number
  tame: number
  /** How many lived when the world began. */
  start: number
  /** Killed by people, all-time. */
  hunted: number
  extinct: boolean
}

export type CropView = {
  item: string
  name: string
  /** Years from planting to the first harvest. */
  growYears: number
  /** Bears again and again (banana, coconut, sago) instead of once. */
  perennial: boolean
  plots: number
  ripe: number
}

/** Food people hold, in units. */
export type FoodStock = { carried: number; stored: number; granary: number }

/** Five counts, one per species in SPECIES order. */
export type PerSpecies = [number, number, number, number, number]

/** The land every 5 simulated seconds. */
export type EcoPoint = {
  time: number
  /** Wild and tame, per species. */
  animals: PerSpecies
  tame: number
  /** Share of the trees still standing, 0–1. */
  forest: number
  plots: number
  ripe: number
  /** Food units people carry and store. */
  food: number
  /** Wild food standing on the land, units. */
  wild: number
  moisture: number
  enso: -1 | 0 | 1
}

/** One closed simulated year. */
export type EcoYear = {
  year: number
  enso: -1 | 0 | 1
  /** Mean rainfall, 1 = an average year. */
  rain: number
  flood?: boolean
  planted: number
  /** Units harvested. */
  harvest: number
  /** Plots lost to drought, floods or raiders. */
  cropLoss: number
  /** … of which to drought. */
  withered: number
  /** Plots left to rot unharvested. */
  wasted: number
  fished: number
  hunted: PerSpecies
  /** Killed by tigers. */
  predated: PerSpecies
  slaughtered: number
  tamed: number
  animalBirths: PerSpecies
  animalStarved: PerSpecies
  /** Alive at the end of the year. */
  animals: PerSpecies
  /** Species ids that died out / arrived from overseas this year. */
  extinct?: string[]
  arrived?: string[]
  population: number
  births: number
  /** People who starved. */
  starved: number
  deaths: number
  /** Food units that went off. */
  rotten: number
}

export type Ecology = {
  secondsPerYear: number
  year: number
  phase: number
  /** 1–12 */
  month: number
  season: Season
  enso: Enso
  rain: number
  moisture: number
  light: number
  /** Share of the trees still standing, 0–1. */
  forest: number
  species: SpeciesView[]
  crops: CropView[]
  plots: number
  ripe: number
  food: FoodStock
  /** Every 5 simulated seconds, oldest first, at most 720 points. */
  history: EcoPoint[]
  /** Closed years, oldest first, at most 400. */
  years: EcoYear[]
}

/** Animal species in stream-frame order. `size` is body size relative to a person. */
export const SPECIES = [
  { id: 'rusa', name: 'Rusa', tame: null, size: 0.9 },
  { id: 'babi_hutan', name: 'Babi Hutan', tame: 'Babi', size: 0.8 },
  { id: 'ayam_hutan', name: 'Ayam Hutan', tame: 'Ayam', size: 0.35 },
  { id: 'kerbau', name: 'Kerbau Liar', tame: 'Kerbau', size: 1.4 },
  { id: 'harimau', name: 'Harimau', tame: null, size: 1.1 },
] as const

/** Crops in `fields` order. */
export const CROPS = [
  { id: 'padi', name: 'Padi' },
  { id: 'talas', name: 'Talas' },
  { id: 'ubi', name: 'Ubi' },
  { id: 'pisang', name: 'Pisang' },
  { id: 'kelapa', name: 'Kelapa' },
  { id: 'sagu', name: 'Sagu' },
] as const

export const SEASON_LABELS: Record<Season, string> = { hujan: 'Musim hujan', kemarau: 'Kemarau' }

export const ENSO_LABELS: Record<Enso, string> = { netral: 'Netral', el_nino: 'El Niño', la_nina: 'La Niña' }

/** ENSO as streamed (-1, 0, 1) to its name. */
export const ensoOf = (v: number): Enso => (v > 0 ? 'el_nino' : v < 0 ? 'la_nina' : 'netral')

export const MONTH_NAMES = [
  'Januari',
  'Februari',
  'Maret',
  'April',
  'Mei',
  'Juni',
  'Juli',
  'Agustus',
  'September',
  'Oktober',
  'November',
  'Desember',
]

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
  teach: 'Mengajar',
  plant: 'Menanam',
  hunt: 'Berburu',
  harvest: 'Memanen',
  fish: 'Memancing',
}

export const DEATH_LABELS: Record<DeathCause, string> = {
  starvation: 'Kelaparan',
  thirst: 'Kehausan',
  oldAge: 'Usia tua',
  killed: 'Dibunuh',
  animal: 'Diterkam hewan',
}

export const SEX_SYMBOL: Record<Sex, string> = { female: '♀', male: '♂' }

export const ROLE_LABELS: Record<CreatureDetail['role'], string> = {
  head: 'Kepala keluarga',
  member: 'Anggota keluarga',
  none: 'Tanpa rumah',
}
