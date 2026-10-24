import { useCallback, useEffect, useRef, useState } from 'react'
import { errorMessage } from './api'

/** Where the latest request stands; only a failure carries an error message. */
type Outcome<T> =
  | { status: 'loading'; data: null; error: null }
  /** The result of the latest request, once it has succeeded. */
  | { status: 'ready'; data: T; error: null }
  | { status: 'error'; data: null; error: string }

export type LoadStatus = Outcome<unknown>['status']

const LOADING = { status: 'loading', data: null, error: null } satisfies Outcome<unknown>

type SupersedingLoad<T> = Outcome<T> & {
  /** The key's result, kept while it reloads; null once another key has been asked for. */
  shown: T | null
  retry: () => void
}

/**
 * The one loader every load goes through, the rule for superseded responses: a
 * response is applied only if it answers the latest request, so a slow earlier
 * response never overwrites the current view. A new key or a retry starts a new
 * request and aborts the previous one.
 */
export function useSupersedingLoad<T>(key: string, load: (signal: AbortSignal) => Promise<T>): SupersedingLoad<T> {
  // Every request has a token of its own: a retry takes the next one and so does a new
  // key, so an outcome stored for an earlier visit to a key never passes for the request
  // now in flight for it. `since` is the token the key was first asked for at.
  const [request, setRequest] = useState({ key, token: 0, since: 0 })
  if (request.key !== key) {
    setRequest({ key, token: request.token + 1, since: request.token + 1 })
  }
  const { key: requestKey, token, since } = request
  const [state, setState] = useState<{ token: number; outcome: Outcome<T> }>({ token, outcome: LOADING })
  const [last, setLast] = useState<{ key: string; token: number; data: T } | null>(null)
  const loadRef = useRef(load)
  useEffect(() => {
    loadRef.current = load
  })

  useEffect(() => {
    const controller = new AbortController()
    loadRef.current(controller.signal).then(
      (data) => {
        if (controller.signal.aborted) return
        setState({ token, outcome: { status: 'ready', data, error: null } })
        setLast({ key: requestKey, token, data })
      },
      (error: unknown) => {
        if (controller.signal.aborted) return
        setState({ token, outcome: { status: 'error', data: null, error: errorMessage(error) } })
        setLast(null)
      },
    )
    return () => controller.abort()
  }, [requestKey, token])

  const retry = useCallback(() => setRequest((r) => ({ ...r, token: r.token + 1 })), [])
  // A result belongs to the current request only if it answers this key and token.
  const current = request.key === key && state.token === token
  const outcome: Outcome<T> = current ? state.outcome : LOADING
  // A reload of the same key keeps its result on screen; a key returned to does not:
  // its earlier visit's result predates `since`.
  const shown = outcome.data ?? (last !== null && last.key === key && last.token >= since ? last.data : null)
  return { ...outcome, shown, retry }
}
