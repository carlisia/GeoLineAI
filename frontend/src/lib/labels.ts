import type { WouldSaveReasons } from './readings.gen'
import {
  ADVISORY_LABELS,
  ADVISORY_MEANINGS,
  CRITERION_LABELS,
  ISSUE_LABELS,
  RECORDS_REASON,
  RESULT_STATE_LABELS,
  VERDICT_LABELS,
  type Advisory,
  type Criterion,
  type IssueCode,
  type ResultState,
  type SourceKind,
  type UnavailableReason,
  type Verdict,
} from './vocabulary.gen'

// The server sends only these values: Go generates the vocabulary file from its lists,
// and the schema's CHECK constraints hold storage to them. Every label the server also
// prints, inside explanations and saved reasons, is generated with the vocabulary.

/**
 * A wire value's entry. The types promise every key, but a response is cast to its
 * type, never checked, so a value outside its set fails here, named, rather than
 * rendering as nothing.
 */
export function entryOf<K extends string, V>(entries: Record<K, V>, value: K, set: string): V {
  if (!Object.hasOwn(entries, value)) throw new Error(`${JSON.stringify(value)} is not a known ${set}`)
  return entries[value]
}

export const verdictLabel = (verdict: Verdict) => entryOf(VERDICT_LABELS, verdict, 'verdict')
export const issueLabel = (code: IssueCode) => entryOf(ISSUE_LABELS, code, 'import issue code')
export const advisoryLabel = (advisory: Advisory) => entryOf(ADVISORY_LABELS, advisory, 'advisory')
export const criterionLabel = (criterion: Criterion) => entryOf(CRITERION_LABELS, criterion, 'criterion')
export const resultStateLabel = (state: ResultState) => entryOf(RESULT_STATE_LABELS, state, 'result state')

/** Where a source's text comes from, as its card labels it. */
const SOURCE_KIND_LABELS: Record<SourceKind, string> = { rule: 'Rule in code', doc: 'Documentation', result: 'Result recorded at import', fact: 'Fact from the stored data' }
export const sourceKindLabel = (k: SourceKind) => entryOf(SOURCE_KIND_LABELS, k, 'source kind')

/** No accuracy has been measured for this configuration, and every shown answer says so. */
export const NOT_MEASURED = 'Accuracy not measured for this configuration.'

/** Why an answer is unavailable, as the page says it; each ends by saying nothing was saved. */
const UNAVAILABLE_MESSAGES: Record<UnavailableReason, string> = {
  no_key: 'Model help is off: the server has no provider key. The evidence below is still available, and reviewing works as usual.',
  provider_error: 'The model service did not answer, so no explanation was written. Nothing was saved; try again.',
  invalid_response: 'The model answered in a form that could not be read, so nothing is shown. Nothing was saved; try again.',
}
export const unavailableMessage = (reason: UnavailableReason) => entryOf(UNAVAILABLE_MESSAGES, reason, 'unavailable reason')

/** What an advisory value means, in the words a saved reason uses, as a caption. */
export const advisoryMeaning = (advisory: Advisory) => entryOf(ADVISORY_MEANINGS, advisory, 'advisory')
export const advisoryCaption = (advisory: Advisory) => {
  const meaning = advisoryMeaning(advisory)
  return meaning.charAt(0).toUpperCase() + meaning.slice(1)
}

/** The aspects no criterion assesses, fixed text the criteria end with. */
export const NOT_ASSESSED = 'Not assessed: reciprocity, receiver contact, raw waveform noise, inversion residual, coordinate consistency between readings, instrument-specific operating ranges.'

/** The labels of the reason notes. */
export const reasonToBeSavedWith = (verdict: Verdict) => `Reason to be saved with ${verdictLabel(verdict)}`
export const reasonSavedWith = (verdict: Verdict) => `Reason saved with ${verdictLabel(verdict)}`

/**
 * Whether a verdict saves no reason, by the save's own rule; the verdicts that do are
 * exactly the keys of the would-save reasons, which the predicate's type says.
 */
export const recordsNoReason = (verdict: Verdict): verdict is Exclude<Verdict, keyof WouldSaveReasons> => !entryOf(RECORDS_REASON, verdict, 'verdict')

/** The note for a verdict that records no reason. */
export const REASON_CLEARED = 'The saved reason will be cleared: Unreviewed records none.'

/** A history entry's reason when it saved none. */
export const NO_REASON = 'None (Unreviewed records no reason)'

/**
 * "{n} unresolved results (v violated, r review)", or "1 unresolved result (1 violated)":
 * the count, then how it splits between the two states, naming only the states present.
 */
export function unresolvedSummary(violated: number, review: number): string {
  const n = violated + review
  const split = [violated > 0 ? `${violated} violated` : '', review > 0 ? `${review} review` : ''].filter(Boolean).join(', ')
  return `${n === 1 ? '1 unresolved result' : `${n} unresolved results`} (${split})`
}

/** Fixed labels for absent values. */
export const NOT_AVAILABLE = 'n/a'
export const UNAVAILABLE = 'Unavailable'

/** Blank or whitespace-only text, such as an acquisition ID or an issue's source text, shows as (blank). */
export function displayBlank(text: string): string {
  return text.trim() === '' ? '(blank)' : text
}

/** Shows a stored UTC time as `YYYY-MM-DD HH:MM:SS`, keeping any non-zero fraction. */
export function formatTime(utc: string): string {
  // The server sends normalized UTC, `YYYY-MM-DDTHH:MM:SS.fffZ` or with nine digits;
  // the golden responses pin that form.
  const parts = /^(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2})\.(\d+)Z$/.exec(utc)
  if (!parts) throw new Error(`${JSON.stringify(utc)} is not a normalized UTC time`)
  const [, date, clock, fraction] = parts
  const kept = fraction.replace(/0+$/, '')
  return `${date} ${clock}${kept ? `.${kept}` : ''}`
}

/** A repeat difference as a percentage with one decimal: 0.5 is "50.0%". */
export function formatPercent(fraction: number): string {
  return `${(fraction * 100).toFixed(1)}%`
}

/** A server message inside a UI sentence ends with exactly one full stop. */
function asSentence(message: string): string {
  return /[.!?]$/.test(message.trim()) ? message.trim() : `${message.trim()}.`
}

/** The status line of a save that failed. */
export function saveFailed(message: string): string {
  return `Save failed: ${asSentence(message)} Your verdict is kept; retry when ready.`
}
