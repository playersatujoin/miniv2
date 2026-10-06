import { queryOptions, useMutation, useQueryClient } from '@tanstack/react-query'
import { api, type UpdateMapInput } from './client'

export const tilesQuery = queryOptions({
  queryKey: ['tiles'],
  queryFn: api.tiles,
  staleTime: Infinity,
})

export const mapsQuery = queryOptions({
  queryKey: ['maps'],
  queryFn: api.listMaps,
})

export const mapQuery = (id: string) =>
  queryOptions({
    queryKey: ['maps', id],
    queryFn: () => api.getMap(id),
    staleTime: 30_000,
  })

/** Deposits only change when the map is edited, so the map version is part of the key. */
export const reliefQuery = (id: string, version: string) =>
  queryOptions({
    queryKey: ['maps', id, 'relief', version],
    queryFn: () => api.relief(id),
    staleTime: Infinity,
    retry: false,
  })

export const geologyQuery = (id: string, version: string) =>
  queryOptions({
    queryKey: ['maps', id, 'geology', version],
    queryFn: () => api.geology(id),
    staleTime: Infinity,
    // The map works without geology; don't hammer a missing endpoint.
    retry: false,
  })

export function useCreateMap() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: api.createMap,
    onSuccess: (map) => {
      qc.setQueryData(mapQuery(map.id).queryKey, map)
      qc.invalidateQueries({ queryKey: mapsQuery.queryKey, exact: true })
    },
  })
}

export function useSaveMap(id: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: UpdateMapInput) => api.updateMap(id, input),
    onSuccess: (map) => {
      qc.setQueryData(mapQuery(id).queryKey, map)
      qc.invalidateQueries({ queryKey: mapsQuery.queryKey, exact: true })
    },
  })
}

export function useDeleteMap() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: api.deleteMap,
    onSuccess: (_, id) => {
      qc.removeQueries({ queryKey: mapQuery(id).queryKey })
      qc.invalidateQueries({ queryKey: mapsQuery.queryKey, exact: true })
    },
  })
}
