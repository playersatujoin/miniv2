import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { simStreamUrl } from './api'
import {
  parseFields,
  parseFrame,
  parseStructures,
  type FieldsMessage,
  type SimFrame,
  type StructuresMessage,
} from './protocol'

export type StreamStatus = 'connecting' | 'live' | 'error'

export type StreamHandlers = {
  /** ~10 times a second. */
  onFrame: (frame: SimFrame) => void
  /** On connect and whenever buildings change. */
  onStructures: (message: StructuresMessage) => void
  /** On connect and whenever planted plots change (at most once a second). */
  onFields?: (message: FieldsMessage) => void
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

    const connect = () => {
      setStatus('connecting')
      const es = new EventSource(simStreamUrl(mapId))
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
    return () => {
      clearTimeout(retry)
      source?.close()
    }
  }, [mapId, enabled])

  return status
}
