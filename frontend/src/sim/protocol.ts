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
  /** A two-person exchange (engine adaptation II): only when both want it. */
  | 'talk'
  | 'trade'
/**
 * `animal`: killed by a wild animal (a tiger, a charging boar or buffalo).
 * `neonatal`: a baby who died in its first weeks. The last three are infectious diseases (Fase 3b).
 */
export type DeathCause =
  | 'starvation'
  | 'thirst'
  | 'oldAge'
  | 'killed'
  | 'animal'
  | 'childbirth'
  | 'neonatal'
  | 'diarrhea'
  | 'malaria'
  | 'respiratory'
  /** Engine adaptation II: fire, water, steep ground. */
  | 'burned'
  | 'drowned'
  | 'fall'

/** Infectious diseases (keys of DiseaseView). */
export type DiseaseKey = 'diarrhea' | 'malaria' | 'respiratory'

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
  /** Ill with an infectious disease. */
  ill: 131072,
  /** Struck in the last moments (a blow, an animal): stagger. */
  hurt: 1 << 18,
  /** In the shallows on foot. */
  wading: 1 << 19,
  /** Out of their depth, swimming. */
  swimming: 1 << 20,
  /** Afloat on a raft. */
  rafting: 1 << 21,
  /** Climbing steep ground or a boulder. */
  climbing: 1 << 22,
  /** Just fell (a slip on a slope): down on the ground for a moment. */
  fallen: 1 << 23,
  /** Talking with someone this second. */
  talking: 1 << 24,
  /** Trading with someone this second. */
  trading: 1 << 25,
  /** Leads a village: the person its people trust most. */
  leader: 1 << 26,
} as const

/** Moods shown on faces (CreatureFrame.mood). */
export const MOOD = { calm: 0, fear: 1, anger: 2, joy: 3, grief: 4 } as const
export type MoodCode = (typeof MOOD)[keyof typeof MOOD]

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
  /** Biological age in years, absent on old servers. */
  age?: number
  lookX?: number | null
  lookY?: number | null
  /** The mood on their face (MOOD) and how strongly, 0–1; absent on old servers. */
  mood?: MoodCode
  moodStrength?: number
}

/** A body where someone died, shown for a while (a lasting stimulus). */
export type CorpseFrame = {
  /** The dead person's id. */
  id: number
  x: number
  y: number
  heading: number
  sex: Sex
  hue: number
  size: number
  /** Age in years at death. */
  age: number
  cause: DeathCause
  /** Simulated seconds since they died. */
  seconds: number
}

/** A tile on fire. */
export type FireFrame = { x: number; y: number; intensity: number }

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
  /** Where the wind blows towards, radians (0 = +x); 0 on old servers. */
  windDir: number
  /** Wind strength 0–1. */
  wind: number
  /** A storm is raging. */
  storm: boolean
}

export type SimFrame = {
  tick: number
  /** Simulated seconds since the world began. */
  time: number
  creatures: CreatureFrame[]
  animals: AnimalFrame[]
  /** Missing from older servers. */
  weather: Weather | null
  /** Bodies lying where people died; empty on old servers. */
  corpses: CorpseFrame[]
  /** Burning tiles; empty on old servers. */
  fires: FireFrame[]
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
  age?: number,
  lookX?: number | null,
  lookY?: number | null,
  mood?: number,
  moodStrength?: number,
]
type RawAnimal = [id: number, species: number, x: number, y: number, heading: number, flags: number]
type RawWeather = [phase: number, light: number, moisture: number, enso: number, rain: number, windDir?: number, wind?: number, storm?: number]
type RawCorpse = [id: number, x: number, y: number, heading: number, sex: 0 | 1, hue: number, size: number, age: number, cause: DeathCause, seconds: number]
type RawFire = [x: number, y: number, intensity: number]
type RawFrame = { t: number; s: number; w?: RawWeather; a?: RawAnimal[]; d?: RawCorpse[]; f?: RawFire[]; c: RawCreature[] }

/** Parses the compact `frame` SSE payload. */
export function parseFrame(data: string): SimFrame {
  const raw = JSON.parse(data) as RawFrame
  return {
    tick: raw.t,
    time: raw.s,
    creatures: raw.c.map(([id, name, x, y, heading, sex, hue, size, flags, energy, health, houseId, age, lookX, lookY, mood, moodStrength]) => ({
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
      age,
      lookX,
      lookY,
      mood: mood === undefined ? undefined : (Math.min(4, Math.max(0, mood)) as MoodCode),
      moodStrength,
    })),
    animals: (raw.a ?? []).map(([id, species, x, y, heading, flags]) => ({ id, species, x, y, heading, flags })),
    weather: raw.w
      ? {
          phase: raw.w[0],
          light: raw.w[1],
          moisture: raw.w[2],
          enso: Math.sign(raw.w[3]) as Weather['enso'],
          rain: raw.w[4],
          windDir: raw.w[5] ?? 0,
          wind: raw.w[6] ?? 0,
          storm: (raw.w[7] ?? 0) > 0,
        }
      : null,
    corpses: (raw.d ?? []).map(([id, x, y, heading, sex, hue, size, age, cause, seconds]) => ({
      id,
      x,
      y,
      heading,
      sex: sex === 1 ? 'male' : 'female',
      hue,
      size,
      age,
      cause,
      seconds,
    })),
    fires: (raw.f ?? []).map(([x, y, intensity]) => ({ x, y, intensity })),
  }
}

/** Scorched land: tiles a fire has passed over and how burnt (0–1), fading as plants return. */
export type BurntMessage = { version: number; tiles: { x: number; y: number; level: number }[] }

/** Parses the `burnt` SSE payload (on connect and when scorched land changes, at most once a second). */
export function parseBurnt(data: string): BurntMessage {
  const raw = JSON.parse(data) as { v: number; t: [number, number, number][] }
  return { version: raw.v, tiles: (raw.t ?? []).map(([x, y, level]) => ({ x, y, level })) }
}

/** A village: houses close together, the land around them, its people and the person they trust most. */
export type Village = {
  id: number
  name: string
  /** Colour for its land, 0–359. */
  hue: number
  /** Convex hull of its houses and fields, in tiles, counter-clockwise. */
  hull: [number, number][]
  /** Centre, in tiles. */
  x: number
  y: number
  people: number
  houses: number
  leader: CreatureRef | null
  /** Simulated seconds when it was first recognised. */
  founded: number
}

export type VillagesMessage = { version: number; villages: Village[] }

/** An event on the replay timeline; importance 1 low, 2 normal, 3 high. */
export type ReplayMark = { tick: number; time: number; kind: SimEventKind; text: string; creatureId?: number; importance: 1 | 2 | 3 }

/** What the server's replay holds: the last few real minutes as streamed (oldest first). */
export type ReplayIndex = {
  /** [tick, simulated seconds] of each recorded frame, about two per real second. */
  frames: [number, number][]
  /** Building versions the recorded frames show. */
  structures: number[]
  marks: ReplayMark[]
}

/** Parses the `villages` SSE payload (on connect and whenever a village changes). */
export function parseVillages(data: string): VillagesMessage {
  const raw = JSON.parse(data) as { v: number; villages: Village[] | null }
  return { version: raw.v, villages: raw.villages ?? [] }
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

/** States of a river or lake tile in the water cycle. */
export const WATER = { dry: 0, under: 1, pools: 2, flowing: 3 } as const

/**
 * The rivers' and lakes' water: per fresh water tile (by index) its state
 * (WATER) and level (0–255: how full; for pools, how much is left).
 */
/** foul: germs in each tile's water (0–255); absent from servers without disease. */
export type WaterMessage = { version: number; tiles: Int32Array; state: Uint8Array; level: Uint8Array; foul?: Uint8Array }

const fromBase64 = (s: string) => Uint8Array.from(atob(s), (c) => c.charCodeAt(0))

export function parseWater(data: string): WaterMessage {
  const raw = JSON.parse(data) as { v: number; t: number[]; s: string; l: string; g?: string }
  return {
    version: raw.v,
    tiles: Int32Array.from(raw.t),
    state: fromBase64(raw.s),
    level: fromBase64(raw.l),
    foul: raw.g ? fromBase64(raw.g) : undefined,
  }
}

/** Mosquitoes per cell (`cell` tiles square, cols × rows): how many (0–255) and the share carrying malaria (0–255). */
export type MosquitoMessage = { cell: number; cols: number; rows: number; density: Uint8Array; infected: Uint8Array }

export function parseMosquitoes(data: string): MosquitoMessage {
  const raw = JSON.parse(data) as { c: number; w: number; h: number; m: string; i: string }
  return { cell: raw.c, cols: raw.w, rows: raw.h, density: fromBase64(raw.m), infected: fromBase64(raw.i) }
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
  /** An outbreak of an infectious disease. */
  | 'disease'
  /** Engine adaptation II: villages founded, merged, abandoned; leaders. */
  | 'village'
  /** A fire breaking out or burning a building. */
  | 'fire'
  /** Barter between two people. */
  | 'trade'
  /** News or an opinion passed on in a talk. */
  | 'rumor'

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
  /** Neurons grown during life, per person on average; the biggest brain; running totals. */
  avgGrown?: number
  maxBrain?: number
  grown?: number
  pruned?: number
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
  /** Engine adaptation II summary; absent on old servers. */
  engine?: EngineInfo
}

/** One barter: a gave gaveN of item `gave` to b for gotN of `got`. */
export type TradeRecord = { time: number; a: CreatureRef; gave: string; gaveN: number; b: CreatureRef; got: string; gotN: number }

export type EngineInfo = {
  /** Lasting stimuli now, bodies among them, and how many of each kind. */
  stimuli: { active: number; corpses?: number; byKind?: Record<string, number> }
  exchange: {
    talks: number
    trades: number
    /** Memories passed on in talks, and opinions of third people. */
    rumors: number
    opinions?: number
    /** Item units that changed hands. */
    units?: number
    /** People waiting for an answer, talking, trading right now. */
    waiting?: number
    talking?: number
    trading?: number
    /** The latest trades (newest last) and the most traded items. */
    recent?: TradeRecord[]
    items?: { item: string; name: string; units: number }[]
  }
  /** Fire now and all-time (started by lightning or hearths, tiles, buildings, crops, people burned), and the wind. */
  fire: {
    burning: number
    scorched?: number
    started?: number
    lightning?: number
    hearth?: number
    tiles?: number
    buildings?: number
    crops?: number
    burned?: number
    windDir?: number
    wind?: number
    storm?: boolean
    storms?: number
  }
  /** People in the water or on slopes now; drowned and falls are all-time deaths. */
  water: {
    swimming: number
    rafting: number
    climbing: number
    wading?: number
    struggling?: number
    fallen?: number
    rafts?: number
    drowned?: number
    falls?: number
  }
  villages: {
    villages: number
    villagers?: number
    led?: number
    largest?: number
    founded?: number
    abandoned?: number
    leaderChanges?: number
  }
}

/** A person's moods, 0–1 each; dominant is '' when calm. */
export type MoodView = {
  fear: number
  anger: number
  joy: number
  grief: number
  dominant: '' | 'fear' | 'anger' | 'joy' | 'grief'
  /** Temperament genes, about 1 each: how strongly events move the moods, how fast they ebb, how cheerful at rest. */
  reactivity?: number
  recovery?: number
  cheer?: number
}

/** A two-person exchange under way, and how many each has had. */
export type ExchangeView = {
  /** ended: tick it was answered or given up (absent while waiting); ok: the exchange took place. */
  chat?: { with: CreatureRef; kind: 'talk' | 'trade'; since: number; ended?: number; ok?: boolean }
  talks: number
  trades: number
  /** Memories they only have from what others told them. */
  told?: number
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
  /** Per inherited neuron: bias, and time constant in ticks (1 = instant, more = holds a memory). */
  bias?: number[]
  tau?: number[]
  /** Neurons grown during this life, in the order they grew (missing on older servers). */
  grown?: number
  /** Age in years when each grew. */
  grownBorn?: number[]
  grownAct?: number[]
  /** How much each has been used lately (activity × reach to the outputs). */
  grownUse?: number[]
  grownTau?: number[]
  grownIn?: number[][] // [grown][input]
  grownRec?: number[][] // [grown][inherited neuron]
  grownOut?: number[][] // [grown][output]
  /** Inherited gene: how readily neurons grow when life surprises. */
  neurogenesis?: number
  /** Felt surprise lately (drives growth). */
  novelty?: number
  /** Neurons gained and pruned in this life so far. */
  grew?: number
  pruned?: number
}

/** The island's brains at a glance, for the Neuron view. */
export type BrainsSummary = {
  population: number
  avgTotal: number
  avgInherited: number
  avgGrown: number
  max: number
  histogram: number[]
  histogramBin: number
  grown: number
  pruned: number
  grownPerYear: number
  prunedPerYear: number
  avgNeurogenesis: number
  avgNovelty: number
  avgTau: number
  /** Share of inherited neurons slow enough to hold a memory (tau ≥ 3). */
  slowNeurons: number
  biggest: { id: number; name: string; sex: Sex; age: number; inherited: number; grown: number; novelty: number }[] | null
  history: { time: number; avgTotal: number; avgGrown: number; max: number }[] | null
  secondsPerYear: number
  /** 0: no limit but energy (every neuron costs food). */
  limit: number
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
  /** Two-person exchanges (engine adaptation II); absent when none. */
  talks?: number
  trades?: number
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
  body: BodyView
	/** Private, bounded memories; a heard event deliberately has no actor. */
	/** 'told': heard about from someone else (gossip, engine adaptation II). */
	memories: { kind: 'give' | 'steal' | 'attack'; actor?: CreatureRef; mode: 'direct' | 'seen' | 'heard' | 'told'; x: number; y: number; tick: number; confidence: number }[]
	relations: { person: CreatureRef; trust: number; tick: number; met: number }[]
	execution: { action: Action; phase: 'approach' | 'blocked' | 'reached' | 'perform'; target?: { x: number; y: number }; remaining?: number; waypoints: number } | null
  /** Engine adaptation II; absent on old servers. */
  mood?: MoodView | null
  exchange?: ExchangeView | null
  /** '' on foot, or wading, swimming, rafting, climbing. */
  locomotion?: '' | 'wading' | 'swimming' | 'rafting' | 'climbing'
  village?: CreatureRef | null
  leader?: boolean
  /** Food places they remember (where, how plentiful 0–1, tick last seen), and water in their tubes. */
  foodPlaces?: { x: number; y: number; rich: number; tick: number }[]
  waterCarried?: number
  waterRoom?: number
  /** Fase 3c: inbreeding, carried variants and disorders; absent on old servers. */
  genetics?: GeneticsView | null
}

/** A person's body and health (Fase 3). */
export type BodyView = {
  stage: 'bayi' | 'anak' | 'remaja' | 'dewasa' | 'lansia'
  /** The illness they have now, or null when well. */
  ill: IllView | null
  /** Immunity to each disease, 0–1. */
  immunity: Record<DiseaseKey, number>
  /** Still carrying malaria parasites without being ill. */
  carrier: boolean
  /** Worm load 0–1. */
  worms: number
  /** Inherited strength of the immune defences (about 1). */
  gene: number
  /** Women: nursing a baby, age (years) fertility ends, a month's chance of conceiving now. */
  nursing: boolean
  menopause: number
  fertility: number
  /** Who looks after a young child. */
  carer: CreatureRef | null
  /** Strength left with age, 0.5–1. */
  vigor: number
  /** Fase 3d: protein and micronutrient status; absent on old servers or with nutrition off. */
  nutrition?: NutritionView | null
}

export type IllView = {
  disease: DiseaseKey
  name: string
  /** 0 just begun … 1 over. */
  progress: number
  /** Share of their remaining health the rest of the illness would take (≥ 1 kills). */
  danger: number
}

/** The island's health at one moment (GET /sim/health history). */
export type EpiPoint = {
  time: number
  population: number
  diarrhea: number
  malaria: number
  respiratory: number
  carriers: number
  worms: number
  mosquitoes: number
  deaths: number
}

export type DiseaseView = {
  key: DiseaseKey
  name: string
  ill: number
  deaths: number
  incidence: number | null
}

/**
 * Map overlays: mosquito, infected and soil are per cell (`cell` tiles square,
 * cols × rows, row-major, 0–255, base64 on the wire); foul lists fresh water
 * tiles with germs and foulLevel how foul (0–255, base64).
 */
export type HealthMap = {
  cell: number
  cols: number
  rows: number
  mosquito: string
  infected: string
  soil: string
  foul: number[]
  foulLevel: string
}

export type HealthView = {
  secondsPerYear: number
  year: number
  now: EpiPoint
  history: EpiPoint[]
  diseases: DiseaseView[]
  latrines: number
  wells: number
  strains: number
  immunity: number
  map: HealthMap
  off?: boolean
}

/** Decodes a base64 byte string from the wire. */
export function bytesOf(b64: string): Uint8Array {
  const bin = atob(b64 ?? '')
  const out = new Uint8Array(bin.length)
  for (let i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i)
  return out
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
  /** Chance of dying before five (5q0). */
  under5Mortality: number | null
  /** Deaths of children under five in the window, by cause. */
  under5ByCause: Partial<Record<DeathCause, number>>
  /** New bouts per person-year of diarrhoea, malaria and respiratory infection. */
  incidence: (number | null)[]
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
  /** Fase 3c: inbreeding by generation and over time; absent on old servers. */
  genetics?: GeneticsInfo
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
  /** Fresh water: shares of river and lake tiles running, in pools, and dry; median groundwater. */
  running?: number
  pools?: number
  dryBeds?: number
  groundwater?: number
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
  /** The water cycle: river and lake tiles by state, median groundwater (1 ≈ a wet season), lakes' fill (-1: none). */
  water?: { flowing: number; pools: number; under: number; dry: number; groundwater: number; lakes: number }
  wells?: number
  wellsDry?: number
  /** Twice a simulated month for the last twenty years (the seasons show). */
  waterHistory?: WaterPoint[]
}

export type WaterPoint = { time: number; running: number; pools: number; dryBeds: number; groundwater: number }

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
  talk: 'Berbicara',
  trade: 'Bertukar barang',
}

export const DEATH_LABELS: Record<DeathCause, string> = {
  starvation: 'Kelaparan',
  thirst: 'Kehausan',
  oldAge: 'Usia tua',
  killed: 'Dibunuh',
  animal: 'Diterkam hewan',
  childbirth: 'Melahirkan',
  neonatal: 'Bayi baru lahir',
  diarrhea: 'Diare',
  malaria: 'Malaria',
  respiratory: 'Radang paru (ISPA)',
  burned: 'Terbakar',
  drowned: 'Tenggelam',
  fall: 'Terjatuh',
}

export const MOOD_LABELS: Record<MoodView['dominant'], string> = {
  '': 'Tenang',
  fear: 'Takut',
  anger: 'Marah',
  joy: 'Senang',
  grief: 'Berduka',
}

export const LOCOMOTION_LABELS: Record<NonNullable<CreatureDetail['locomotion']>, string> = {
  '': 'Berjalan',
  wading: 'Mengarungi air',
  swimming: 'Berenang',
  rafting: 'Di atas rakit',
  climbing: 'Memanjat',
}

export const DISEASE_LABELS: Record<DiseaseKey, string> = {
  diarrhea: 'Diare',
  malaria: 'Malaria',
  respiratory: 'ISPA',
}

export const STAGE_LABELS: Record<BodyView['stage'], string> = {
  bayi: 'Bayi',
  anak: 'Anak',
  remaja: 'Remaja',
  dewasa: 'Dewasa',
  lansia: 'Lansia',
}

export const SEX_SYMBOL: Record<Sex, string> = { female: '♀', male: '♂' }

export const ROLE_LABELS: Record<CreatureDetail['role'], string> = {
  head: 'Kepala keluarga',
  member: 'Anggota keluarga',
  none: 'Tanpa rumah',
}

// --- Fase 3d: nutrition beyond calories --------------------------------------

/** The gizi label: the worse of the protein/micronutrient deficit and thinness. */
export type NutritionLabel = 'baik' | 'kurang' | 'buruk'

/** Why someone needs denser food than an ordinary adult ('' = they don't). */
export type NutritionNeed = '' | 'bayi' | 'balita' | 'anak' | 'remaja' | 'hamil' | 'menyusui' | 'lansia'

/** A person's nutrition: `BodyView.nutrition` in the creature detail (absent when nutrition is switched off). */
export type NutritionView = {
  label: NutritionLabel
  /** Body status, 1 = adequate (up to 1.25 with a reserve, 0 = empty). */
  protein: number
  micro: number
  /** What the brain feels as "kurang gizi", 0 well fed … 1 severely short. */
  deficit: number
  /** What their food gives against their own need (1 = enough). */
  dietProtein: number
  dietMicro: number
  need: NutritionNeed
  /** Height against a well-fed person's of the same age (1 = full); stunted below about −2 SD. */
  stature: number
  stunted: boolean
  thin: boolean
}

/** `BodyView` as sent since Fase 3d (the lead may fold `nutrition` into `BodyView`). */
export type BodyViewWithNutrition = BodyView & { nutrition?: NutritionView | null }

/** The island's nutrition now (sim.NutritionInfo; soak reports). */
export type NutritionInfo = {
  people: number
  deficient: number
  severe: number
  proteinShort: number
  microShort: number
  thin: number
  malnourished: number
  under5: number
  stunted5: number
  children: number
  stuntedChild: number
  adults: number
  stuntedAdults: number
  pregnantShort: number
  pregnantOrNursing: number
  dietProtein: number
  dietMicro: number
  adultStunt: number
}

export const NUTRITION_LABELS: Record<NutritionLabel, string> = {
  baik: 'Gizi baik',
  kurang: 'Gizi kurang',
  buruk: 'Gizi buruk',
}

export const NUTRITION_NEED_LABELS: Record<NutritionNeed, string> = {
  '': 'seperti orang dewasa',
  bayi: 'bayi (ASI)',
  balita: 'balita: zat gizi mikro jauh lebih padat',
  anak: 'anak: zat gizi mikro lebih padat',
  remaja: 'remaja yang sedang tumbuh',
  hamil: 'hamil: protein dan zat gizi mikro lebih banyak',
  menyusui: 'menyusui: protein dan zat gizi mikro lebih banyak',
  lansia: 'lansia: protein sedikit lebih banyak',
}
// --- Food sharing (backend sim/sharing.go) -------------------------------------

/** Food sharing: all-time counts, and the meat lying out and the food people hold now (SimInfo.engine.sharing). */
export type SharingInfo = {
  /** Meat units left at kills that nobody could carry, then eaten or cut from them, or lost to rot and scavengers. */
  meatLeft: number
  meatTaken: number
  meatRotted: number
  /** Meals eaten from the food of family close by, or from a neighbour's load beyond what they can use. */
  fromKin: number
  fromOthers: number
  /** Units cut in another household's field and carried to its store (bawon), and the harvesters' share. */
  bawon: number
  bawonShare: number
  /** Meals drawn from a village granary in hunger. */
  granaryMeals: number
  /** Ownerless granaries taken over by the household beside them. */
  adopted?: number
  /** Times someone getting hungry learned of better land near their water. */
  landKnown?: number
  carcasses: number
  carcassMeat: number
  granaries: number
  granaryFood: number
  houseFood: number
  carriedFood: number
}

/** Food a village keeps (VillageView), in units: in its houses and in its granaries. */
export type VillageStores = { houseFood?: number; granaryFood?: number }
// --- Fase 3c: genetics (backend sim/genetics.go, geneview.go) ---------------------------
// The lead adds `genetics?: GeneticsView | null` to CreatureDetail and
// `genetics?: GeneticsInfo` to Demography when merging.

/** One person's genes (CreatureDetail.genetics, FamilyTree.genetics). */
export type GeneticsView = {
  /** Inbreeding coefficient: 0 unrelated parents, 1/16 first cousins, 1/4 siblings. */
  f: number
  /** Loci with one copy of a recessive variant: a healthy carrier. */
  carried: number
  /** Recessive disorders they have (two copies), named in Indonesian. */
  defects: string[]
  /** Carries thalassaemia: malaria goes milder. */
  malariaShield: boolean
  /** False for people born before genetics: their variants are a guess. */
  known: boolean
}

/** Someone in a family tree. */
export type FamilyPerson = {
  id: number
  /** Empty when the pedigree forgot them and no living child remembers the name. */
  name: string
  sex: Sex | ''
  generation: number
  mother?: number
  father?: number
  f: number
  alive: boolean
  /** Calendar year of birth (1-based). */
  born: number
  died?: number
  /** One of the first humans of an era (an Adam or a Hawa). */
  founder?: boolean
  /** Nobody alive descends from them within ten generations: only the id is kept. */
  forgotten?: boolean
}

/** GET /api/maps/{id}/sim/creatures/{cid}/family?depth=N */
export type FamilyTree = {
  root: number
  depth: number
  /** Level k (1-based) holds the 2^k places k generations up, Ahnentafel order (father, mother); 0 = unknown. */
  ancestors: number[][]
  /** Level k holds everyone k generations down, by id. */
  descendants: number[][]
  people: FamilyPerson[]
  /** The person's genes, when alive. */
  genetics: GeneticsView | null
  /** Coefficient of relationship of the person's parents. */
  parentsRelated: number
}

/** Demography.genetics: inbreeding on the island, by generation and over time. */
export type GeneticsInfo = {
  remembered: number
  meanF: number
  /** Share of the living born to second cousins or closer (F ≥ 1/64). */
  inbred: number
  /** Share of the living born to close kin (F ≥ 1/8). */
  closeKin: number
  /** Recessive variants carried (one copy) per living person. */
  carried: number
  /** Share of the living with a recessive disorder. */
  affected: number
  /** Babies lost to recessive disorders, all time. */
  lethal: number
  variants: { name: string; frequency: number }[]
  byGeneration: { label: string; births: number; meanF: number; consanguineous: number; closeKin: number }[]
  byF: { label: string; births: number; affected: number; lethal: number; under1: number; under15: number; q15: number | null }[]
  periods: {
    year: number
    births: number
    meanF: number
    closeKin: number
    /** Share of moments adults looking at a parent, child or sibling wanted to mate (null: too few). */
    kinDesire: number | null
    otherDesire: number | null
    kinSeen: number
    otherSeen: number
    otherR: number
  }[]
}
