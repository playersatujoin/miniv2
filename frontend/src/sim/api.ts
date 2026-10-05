import { queryOptions, useMutation, useQueryClient } from '@tanstack/react-query'
import { request } from '../api/client'
import type { CreatureDetail, Demography, Knowledge, SimInfo, SimSpeed } from './protocol'

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

export function useSetSimSpeed(mapId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (speed: SimSpeed) =>
      request<SimInfo>(`/maps/${mapId}/sim/speed`, { method: 'PUT', body: JSON.stringify({ speed }) }),
    onSuccess: (info) => qc.setQueryData(simInfoQuery(mapId).queryKey, info),
  })
}

export const simStreamUrl = (mapId: string) => `/api/maps/${mapId}/sim/stream`
