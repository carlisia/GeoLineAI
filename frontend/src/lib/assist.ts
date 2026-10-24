/** What the explanation view needs that is not a component. */
import type { ExplanationResponse, ReadingDetail, Source } from './api'
import { advisoryLabel, advisoryMeaning, criterionLabel } from './labels'

/**
 * The questions offered for one reading: why each of its first two unresolved criteria
 * triggered, what the repeat difference measures, and what its advisory establishes.
 */
export function suggestedQuestions(detail: ReadingDetail): string[] {
  const unresolved = [...new Set(detail.advisory.results.filter((r) => r.state !== 'met').map((r) => r.criterion))].slice(0, 2)
  const advisory =
    detail.advisory.value === 'pass' ? `What does "${advisoryMeaning('pass')}" actually establish?` : `Why is the advisory ${advisoryLabel(detail.advisory.value)}?`
  return [...unresolved.map((c) => `Why did ${criterionLabel(c)} trigger?`), 'What does the repeat difference measure?', advisory]
}

/**
 * The sources the sentences cite, each once, in the order first cited. A withheld
 * answer may cite an ID it was not offered, which is one reason it was withheld and
 * which its problems name, so that citation has no source to show. A shown answer
 * passed every citation check, so one pointing nowhere is impossible and fails here.
 */
export function citedSources(answer: ExplanationResponse): Source[] {
  const byId = new Map(answer.sources.map((s) => [s.id, s]))
  const ids = [...new Set(answer.sentences.flatMap((s) => s.sources))]
  return ids.flatMap((id) => {
    const source = byId.get(id)
    if (source !== undefined) return [source]
    if (answer.outcome === 'shown') throw new Error(`a shown answer cites ${id}, which is not one of its sources`)
    return []
  })
}
