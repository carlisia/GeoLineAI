import { http, HttpResponse, type JsonBodyType } from 'msw'
import { setupServer } from 'msw/node'
import type { ReadingDetail, ScenariosResponse } from '../lib/api'
import { detailResponse, errorResponse as errorBody } from './responses.golden'

type Answer = unknown | Response | Promise<unknown | Response>

/** A 500 response: the golden error body with its message replaced, the one field the frontend reads. */
export function errorResponse(message: string, status = 500): Response {
  return HttpResponse.json({ ...errorBody, message }, { status })
}

/**
 * A reading's detail built from the golden detail, unreviewed with no history; the
 * golden is read after a save, so the tests start clean and make their own history.
 */
export function detailFor(id: string, acquisitionId: string, overrides: Partial<ReadingDetail> = {}): ReadingDetail {
  return { ...detailResponse, reading_id: id, acquisition_id: acquisitionId, verdict: 'unreviewed', history: [], reason: null, ...overrides } as ReadingDetail
}

/** Three scenarios over readings 1, 2 and 3. */
export const scenarios: ScenariosResponse = {
  scenarios: [
    { label: 'Passing reading', reading_id: '1' },
    { label: 'Violated criterion', reading_id: '2' },
    { label: 'Disagreeing repeats', reading_id: '3' },
  ],
}

/** The backend every test talks to, over the real fetch. */
export const server = setupServer()

/** Requests no answer was given for; the test setup fails a test that leaves any here. */
export const strayRequests: string[] = []

/**
 * Serves the answers over fetch, keyed by method and path (`GET /api/readings/1`),
 * each a value, a Response, a pending promise, or a function called per request for
 * routes asked more than once. The scenarios and the three readings have defaults.
 */
export function installBackend(answers: Record<string, Answer | (() => Answer)> = {}) {
  const sent: { request: string; body: unknown }[] = []
  const defaults: Record<string, Answer> = {
    'GET /api/scenarios': scenarios,
    'GET /api/readings/1': detailFor('1', 'line-1-r00001'),
    'GET /api/readings/2': detailFor('2', 'line-1-r00261'),
    'GET /api/readings/3': detailFor('3', 'line-1-r01064'),
  }
  server.use(
    http.all('*', async ({ request }) => {
      const url = new URL(request.url)
      const key = `${request.method} ${url.pathname}`
      const text = await request.text()
      sent.push({ request: key, body: text ? JSON.parse(text) : null })
      if (!(key in answers) && !(key in defaults)) {
        strayRequests.push(key)
        return errorResponse(`unexpected request ${key}`)
      }
      const given = key in answers ? answers[key] : defaults[key]
      const body = await (typeof given === 'function' ? (given as () => Answer)() : given)
      return body instanceof Response ? body : HttpResponse.json(body as JsonBodyType)
    }),
  )
  return sent
}

/** A response the test releases when it chooses, to order responses against each other. */
export function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((r) => (resolve = r))
  return { promise, resolve }
}
