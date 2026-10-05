export type TileDef = {
  id: number
  key: string
  name: string
  layer: 'ground' | 'objects'
  color: string
  solid: boolean
}

export type TileSet = { ground: TileDef[]; objects: TileDef[] }

export type Point = { x: number; y: number }

/** Row-major tile IDs: index = y * width + x. */
export type MapLayers = { ground: number[]; objects: number[] }

export type MapSummary = {
  id: string
  name: string
  width: number
  height: number
  seed: number
  createdAt: string
  updatedAt: string
}

export type GameMap = MapSummary & {
  tileSize: number
  spawn: Point
  layers: MapLayers
}

/** An item that occurs as a deposit, for the map's resource layer. */
export type DepositItem = { id: string; name: string; kind: string; formula?: string; elements: string[] }

/** A rock unit of the geological map, indexed by id. */
export type RockType = { id: number; key: string; name: string; color: string; description: string }

/** A named geological feature, for labels on the geological map. */
export type GeoFeature = {
  kind: 'volcano' | 'pluton' | 'pegmatite' | 'ophiolite' | 'basin' | 'evaporite' | 'carbonatite' | 'karst' | 'river' | 'delta'
  name: string
  x: number
  y: number
}

/** A real-world ore-deposit model, e.g. "Porfiri Cu-Au-Mo" like Grasberg. */
export type DepositModel = { key: string; name: string; description: string; example: string }

/** A map's geology: rock units per tile, features, and notable resource deposits. */
export type MapGeology = {
  width: number
  height: number
  rockTypes: RockType[]
  /** width*height rock ids (index into rockTypes), row-major. */
  rocks: number[]
  features: GeoFeature[]
  models: DepositModel[]
  items: DepositItem[]
  /** [x, y, index into items, 1 if lying on the tile else 0 (in the ground), index into models] */
  deposits: [x: number, y: number, item: number, surface: 0 | 1, model: number][]
}

export type CreateMapInput = { name: string; width: number; height: number; seed?: number }
export type UpdateMapInput = { name: string; spawn: Point; layers: MapLayers }

export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
  ) {
    super(message)
  }
}

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`/api${path}`, {
    ...init,
    headers: init?.body ? { 'Content-Type': 'application/json' } : undefined,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => null)
    throw new ApiError(res.status, body?.error ?? res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  tiles: () => request<TileSet>('/tiles'),
  listMaps: () => request<MapSummary[]>('/maps'),
  getMap: (id: string) => request<GameMap>(`/maps/${id}`),
  createMap: (input: CreateMapInput) =>
    request<GameMap>('/maps', { method: 'POST', body: JSON.stringify(input) }),
  updateMap: (id: string, input: UpdateMapInput) =>
    request<GameMap>(`/maps/${id}`, { method: 'PUT', body: JSON.stringify(input) }),
  deleteMap: (id: string) => request<void>(`/maps/${id}`, { method: 'DELETE' }),
  geology: (id: string) => request<MapGeology>(`/maps/${id}/geology`),
}

export function previewUrl(m: MapSummary, scale = 2) {
  return `/api/maps/${m.id}/preview.png?scale=${scale}&v=${Date.parse(m.updatedAt)}`
}
