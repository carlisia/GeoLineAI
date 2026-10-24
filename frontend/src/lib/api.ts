/**
 * Backend client. Paths are relative so the dev server's proxy forwards `/_/*` and
 * `/api/*` to the Go server. The response types are generated from Go and checked
 * against real responses by the tests.
 */
import type { Answer, Source } from './grounded.gen'
import type { Detail } from './readings.gen'
import type { List } from './scenarios.gen'
import type { ErrorBody } from './server.gen'
import type { Outcome } from './verdict.gen'
import type { Verdict } from './vocabulary.gen'

// Every response shape is generated from the Go type the server answers with; these
// are the names the page knows them by. Go's are scoped by package, so a bare List or
// Detail says nothing on the page.
export type { ErrorBody, Source }
export type ScenariosResponse = List
export type ReadingDetail = Detail
export type SaveResponse = Outcome
export type ExplanationResponse = Answer

/** What a failed request says, for showing to the reviewer. */
export function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message
  throw new Error(`a request failed with a value that is not an Error: ${String(error)}`)
}

// An aborted request fails here like any other. A request is only ever aborted because
// a newer one replaced it; a replaced request's outcome never reaches the page.
async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  let response: Response
  try {
    response = await fetch(path, { ...init, headers: { Accept: 'application/json', ...init.headers } })
  } catch (cause) {
    throw new Error(`Could not reach the server at ${path}`, { cause })
  }
  if (!response.ok) {
    const body = (await response.json().catch(() => null)) as ErrorBody | null
    // The server's message is written for the reviewer, word for word.
    if (body !== null) throw new Error(body.message)
    throw new Error(`${path} responded ${response.status} ${response.statusText}`)
  }
  try {
    return (await response.json()) as T
  } catch (cause) {
    throw new Error(`${path} did not answer with JSON`, { cause })
  }
}

export function fetchScenarios(signal?: AbortSignal): Promise<ScenariosResponse> {
  return request('/api/scenarios', { signal })
}

export function fetchReading(id: string, signal?: AbortSignal): Promise<ReadingDetail> {
  return request(`/api/readings/${encodeURIComponent(id)}`, { signal })
}

/**
 * Saves one verdict for one reading. The answer names the reading in exactly one of
 * its lists; one that does not confirms nothing about the save, so it fails here like
 * any other failure and no caller reads it as saved.
 */
export async function saveVerdict(readingId: string, verdict: Verdict): Promise<SaveResponse> {
  const answer = await request<SaveResponse>('/api/reviews', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ targets: [{ reading_id: readingId }], verdict }),
  })
  const named = [...answer.affected_reading_ids, ...answer.unchanged_reading_ids]
  if (named.length !== 1 || named[0] !== readingId) throw new Error(`The server did not say whether reading ${readingId} was saved.`)
  return answer
}

/** Asks a question about one reading's criteria; the server writes the answer, checks it, and stores nothing. */
export function askExplanation(readingId: string, question: string): Promise<ExplanationResponse> {
  return request(`/api/readings/${encodeURIComponent(readingId)}/explanations`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ question }),
  })
}
