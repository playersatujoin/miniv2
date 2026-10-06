import { queryOptions, useMutation, useQueryClient } from '@tanstack/react-query'
import { request } from '../api/client'
import {
  parseFrame,
  parseStructures,
  type BrainsSummary,
  type CreatureDetail,
  type Demography,
  type Ecology,
  type HealthView,
  type Knowledge,
  type MinedOut,
  type ReplayIndex,
  type SimFrame,
  type SimInfo,
  type SimSpeed,
  type StructuresMessage,
  type Village,
} from './protocol'

export const simInfoQuery = (mapId: string) =>
  queryOptions({
    queryKey: ['sim', mapId, 'info'],
    queryFn: () => request<SimInfo>(`/maps/${mapId}/sim`),
    refetchInterval: 1000,
  })

/** Live creature detail; 404 (ApiError.status) means the creature has died. */
export const creatureQuery = (mapId: string, id: number) =>
  queryOptions({
    queryKey: ['sim', mapId, 'creature', id],
    queryFn: () => request<CreatureDetail>(`/maps/${mapId}/sim/creatures/${id}`),
    refetchInterval: (query) => (query.state.status === 'error' ? false : 250),
    retry: false,
  })

/** Civilisation knowledge: elements (periodic table), techs and building kinds. */
export const knowledgeQuery = (mapId: string) =>
  queryOptions({
    queryKey: ['sim', mapId, 'knowledge'],
    queryFn: () => request<Knowledge>(`/maps/${mapId}/sim/knowledge`),
    refetchInterval: 3000,
  })

/** Demographic indicators (life table, fertility, inequality) and reference ranges. */
export const demographyQuery = (mapId: string) =>
  queryOptions({
    queryKey: ['sim', mapId, 'demography'],
    queryFn: () => request<Demography>(`/maps/${mapId}/sim/demography`),
    refetchInterval: 3000,
  })

/** The land: seasons and weather, animals, fields and food, with their history. */
export const ecologyQuery = (mapId: string) =>
  queryOptions({
    queryKey: ['sim', mapId, 'ecology'],
    queryFn: () => request<Ecology>(`/maps/${mapId}/sim/ecology`),
    refetchInterval: 3000,
  })

/** Disease: who is ill, the epidemic curve, and the mosquito, worm and fouled-water maps. */
export const healthQuery = (mapId: string) =>
  queryOptions({
    queryKey: ['sim', mapId, 'health'],
    queryFn: () => request<HealthView>(`/maps/${mapId}/sim/health`),
    refetchInterval: 2000,
    retry: false,
  })

/** The island's brains: sizes, growth and pruning, the biggest alive, and their history. */
export const brainsQuery = (mapId: string) =>
  queryOptions({
    queryKey: ['sim', mapId, 'brains'],
    queryFn: () => request<BrainsSummary>(`/maps/${mapId}/sim/brains`),
    refetchInterval: 2000,
    retry: false,
  })

/** Mined-out deposits (old pits) and felled trees (stumps) of a living world. */
export const minedOutQuery = (mapId: string) =>
  queryOptions({
    queryKey: ['sim', mapId, 'mined'],
    queryFn: () => request<MinedOut>(`/maps/${mapId}/sim/mined`),
    refetchInterval: 10_000,
    retry: false,
  })

export function useSetSimSpeed(mapId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (speed: SimSpeed) =>
      request<SimInfo>(`/maps/${mapId}/sim/speed`, { method: 'PUT', body: JSON.stringify({ speed }) }),
    onSuccess: (info) => qc.setQueryData(simInfoQuery(mapId).queryKey, info),
  })
}

export const simStreamUrl = (mapId: string) => `/api/maps/${mapId}/sim/stream`

/** Villages with their land, people and leaders (the stream also pushes these). */
export const villagesQuery = (mapId: string) =>
  queryOptions({
    queryKey: ['sim', mapId, 'villages'],
    queryFn: () => request<Village[]>(`/maps/${mapId}/sim/villages`),
    refetchInterval: 3000,
    retry: false,
  })

/** The replay's timeline: recorded frames, building versions and markers. */
export const replayIndexQuery = (mapId: string) =>
  queryOptions({
    queryKey: ['sim', mapId, 'replay'],
    queryFn: () => request<ReplayIndex>(`/maps/${mapId}/sim/replay`),
    refetchInterval: 2000,
    retry: false,
  })

/** A recorded frame at or just before tick, and the building version it shows. */
export async function fetchReplayFrame(mapId: string, tick: number): Promise<{ frame: SimFrame; structures: number }> {
  const res = await fetch(`/api/maps/${mapId}/sim/replay/frame?tick=${Math.floor(tick)}`)
  if (!res.ok) throw new Error(`replay frame: ${res.status}`)
  return { frame: parseFrame(await res.text()), structures: Number(res.headers.get('X-Structures-Version') ?? 0) }
}

/** Buildings as they stood in a recorded version. */
export async function fetchReplayStructures(mapId: string, version: number): Promise<StructuresMessage> {
  const res = await fetch(`/api/maps/${mapId}/sim/replay/structures/${version}`)
  if (!res.ok) throw new Error(`replay structures: ${res.status}`)
  return parseStructures(await res.text())
}
