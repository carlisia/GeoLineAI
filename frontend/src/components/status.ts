import { entryOf } from '../lib/labels'
import { RESULT_STATE_ORDER, type Advisory, type ResultState, type Verdict } from '../lib/vocabulary.gen'
import { colors } from '../theme'

// Status colours. Every status is also written out as text beside its colour, so
// colour never carries the meaning alone.

const ADVISORY_COLORS: Record<Advisory, string> = { pass: colors.pass, fail: colors.fail, needs_review: colors.review }
export const advisoryColor = (advisory: Advisory) => entryOf(ADVISORY_COLORS, advisory, 'advisory')

const VERDICT_COLORS: Record<Verdict, string> = { pass: colors.pass, fail: colors.fail, unreviewed: colors.textSecondary }
export const verdictColor = (verdict: Verdict) => entryOf(VERDICT_COLORS, verdict, 'verdict')

/** A violation takes the Fail red, a review the Needs review amber, a met result the secondary text. */
const RESULT_STATE_COLORS: Record<ResultState, string> = { violated: colors.fail, review: colors.review, met: colors.textSecondary }
export const resultStateColor = (state: ResultState) => entryOf(RESULT_STATE_COLORS, state, 'result state')

/** Each state's place in the criteria list, in the order the criteria rank them. */
export function resultStateRank(state: ResultState): number {
  const rank = RESULT_STATE_ORDER.indexOf(state)
  if (rank < 0) throw new Error(`${JSON.stringify(state)} is not a known result state`)
  return rank
}
