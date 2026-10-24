import { useState } from 'react'
import { fetchReading } from './api'
import type { Scenario } from './scenarios.gen'
import { useReview } from './useReview'
import { useSupersedingLoad } from './useSupersedingLoad'
import { useUnloadWarning } from './useUnloadWarning'

/**
 * The reading the reviewer has open, and the rules for moving between readings: the
 * first scenario opens at once; a save in flight locks moving; an unsaved choice asks
 * before another reading replaces it, through the prompt; and the browser warns
 * before the page is left with a choice unsaved or a save in flight.
 */
export function useReadingSession(scenarios: Scenario[]) {
  const [selected, setSelected] = useState(scenarios[0].reading_id)
  const [pending, setPending] = useState<string | null>(null)
  const detail = useSupersedingLoad(selected, (signal) => fetchReading(selected, signal))
  const review = useReview(selected, detail.shown?.verdict ?? null, detail.status, detail.retry)
  const chosen = review.chosen !== null
  useUnloadWarning(chosen || review.saving)

  const open = (id: string) => {
    setPending(null)
    setSelected(id)
  }
  /** Asks for another reading: refused while saving, prompted with a choice unsaved. */
  const pick = (id: string) => {
    if (review.saving || id === selected) return
    if (chosen) setPending(id)
    else open(id)
  }
  const prompt = {
    open: pending !== null,
    discard: () => pending !== null && open(pending),
    keep: () => setPending(null),
  }
  return { selected, detail, review, pick, locked: review.saving, prompt }
}
