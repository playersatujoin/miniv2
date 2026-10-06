import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { simStreamUrl } from './api'
import {
  parseBurnt,
  parseFields,
  parseFrame,
  parseMosquitoes,
  parseStructures,
  parseVillages,
  parseWater,
  type BurntMessage,
  type FieldsMessage,
  type MosquitoMessage,
  type SimFrame,
  type StructuresMessage,
  type VillagesMessage,
  type WaterMessage,
} from './protocol'

export type StreamStatus = 'connecting' | 'live' | 'error'

export type StreamHandlers = {
  /** Padded camera coverage; changing cells reconnects the read-only stream. */
  viewport?: () => { x: number; y: number; across: number; follow?: number | null } | null
  /** ~10 times a second. */
  onFrame: (frame: SimFrame) => void
  /** On connect and whenever buildings change. */
  onStructures: (message: StructuresMessage) => void
  /** On connect and whenever planted plots change (at most once a second). */
  onFields?: (message: FieldsMessage) => void
  /** On connect and whenever rivers and lakes rise or fall noticeably (at most every 0.4 s). */
  onWater?: (message: WaterMessage) => void
  /** Every two seconds while the world runs. */
  onMosquitoes?: (message: MosquitoMessage) => void
  /** On connect and whenever a village changes. */
  onVillages?: (message: VillagesMessage) => void
  /** On connect and whenever scorched land changes (at most once a second). */
  onBurnt?: (message: BurntMessage) => void
}

/**
 * Subscribes to a map's simulation stream while `enabled`. Messages go straight
 * to the handlers without touching React state; only the connection status
 * re-renders.
 */
export function useSimStream(mapId: string, enabled: boolean, handlers: StreamHandlers): StreamStatus {
  const [status, setStatus] = useState<StreamStatus>('connecting')
  const handlersRef = useRef(handlers)
  useLayoutEffect(() => {
    handlersRef.current = handlers
  })

  useEffect(() => {
    if (!enabled) return
    let source: EventSource | null = null
    let retry: ReturnType<typeof setTimeout> | undefined
    let delay = 1000
    let viewKey = ''
    const region = () => {
      const v = handlersRef.current.viewport?.()
      if (!v || !Number.isFinite(v.across) || v.across <= 0) return ''
      // Broad coverage also includes the oblique 3D view's horizon. Quantizing
      // avoids reconnecting on every camera movement.
      const radius = Math.ceil(v.across / 16) * 16 + 24
      const x = Math.floor(v.x / 16) * 16,
        y = Math.floor(v.y / 16) * 16
      return new URLSearchParams({
        x0: String(x - radius),
        y0: String(y - radius),
        x1: String(x + radius),
        y1: String(y + radius),
        follow: String(v.follow ?? 0),
      }).toString()
    }

    const connect = () => {
      setStatus('connecting')
      viewKey = region()
      const es = new EventSource(simStreamUrl(mapId) + (viewKey ? `?${viewKey}` : ''))
      source = es
      let live = false
      es.addEventListener('frame', (e) => {
        if (!live) {
          live = true
          delay = 1000
          setStatus('live')
        }
        handlersRef.current.onFrame(parseFrame((e as MessageEvent<string>).data))
      })
      es.addEventListener('structures', (e) => {
        handlersRef.current.onStructures(parseStructures((e as MessageEvent<string>).data))
      })
      es.addEventListener('fields', (e) => {
        handlersRef.current.onFields?.(parseFields((e as MessageEvent<string>).data))
      })
      es.addEventListener('water', (e) => {
        handlersRef.current.onWater?.(parseWater((e as MessageEvent<string>).data))
      })
      es.addEventListener('mosquitoes', (e) => {
        handlersRef.current.onMosquitoes?.(parseMosquitoes((e as MessageEvent<string>).data))
      })
      es.addEventListener('villages', (e) => {
        handlersRef.current.onVillages?.(parseVillages((e as MessageEvent<string>).data))
      })
      es.addEventListener('burnt', (e) => {
        handlersRef.current.onBurnt?.(parseBurnt((e as MessageEvent<string>).data))
      })
      es.onerror = () => {
        live = false
        if (es.readyState !== EventSource.CLOSED) {
          setStatus('connecting') // the browser retries on its own
          return
        }
        // An HTTP error (e.g. the proxy while the server restarts) makes
        // EventSource give up, so reconnect ourselves with backoff.
        setStatus('error')
        retry = setTimeout(connect, delay)
        delay = Math.min(delay * 2, 15000)
      }
    }
    connect()
    const camera = setInterval(() => {
      if (region() === viewKey) return
      clearTimeout(retry)
      source?.close()
      connect()
    }, 750)
    return () => {
      clearInterval(camera)
      clearTimeout(retry)
      source?.close()
    }
  }, [mapId, enabled])

  return status
}
