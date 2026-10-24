import { useState } from 'react'
import { errorMessage, saveVerdict } from './api'
import { saveFailed, verdictLabel } from './labels'
import type { LoadStatus } from './useSupersedingLoad'
import type { Verdict } from './vocabulary.gen'

type SaveState =
  | { kind: 'idle' }
  | { kind: 'saving'; verdict: Verdict }
  | { kind: 'done'; verdict: Verdict; unchanged: boolean }
  | { kind: 'failed'; verdict: Verdict; message: string }

/** One reading's review state; a new reading starts from a fresh one. */
interface ReviewState {
  readingId: string
  chosen: Verdict | null
  committed: Verdict | null
  save: SaveState
}

const fresh = (readingId: string): ReviewState => ({ readingId, chosen: null, committed: null, save: { kind: 'idle' } })

/**
 * The review state of the shown reading: the chosen verdict, its save and the status
 * line. A chosen verdict belongs to its reading until a successful save clears it; a
 * failure keeps it for retry. Only an explicit save changes the saved verdict.
 */
export function useReview(readingId: string, stored: Verdict | null, status: LoadStatus, onSaved: () => void) {
  const [state, setState] = useState(() => fresh(readingId))
  // Another reading starts its own review: nothing chosen for one carries to the next.
  const current = state.readingId === readingId ? state : fresh(readingId)
  if (current !== state) setState(current)
  const { chosen, committed, save } = current
  const update = (patch: Partial<ReviewState>) => setState((s) => (s.readingId === readingId ? { ...s, ...patch } : s))

  const saving = save.kind === 'saving'
  // Until a refetch confirms it, the last committed save outranks older evidence that
  // a slower response may carry.
  const saved = status === 'ready' ? stored : (committed ?? stored)
  // The controls act only on the current detail, and nothing competes with a save.
  const locked = status !== 'ready' || saving

  // Choosing the saved verdict is never a proposal: it clears the choice.
  const choose = (v: Verdict) => {
    if (locked) return
    update({ chosen: v === saved ? null : v, save: { kind: 'idle' } })
  }

  const canSave = !locked && chosen !== null

  const submit = () => {
    if (!canSave || chosen === null) return
    const verdict = chosen
    update({ save: { kind: 'saving', verdict } })
    saveVerdict(readingId, verdict).then(
      (res) => {
        update({ chosen: null, committed: verdict, save: { kind: 'done', verdict, unchanged: res.unchanged_reading_ids.includes(readingId) } })
        onSaved()
      },
      (error: unknown) => update({ save: { kind: 'failed', verdict, message: errorMessage(error) } }),
    )
  }

  const saveLabel = saving ? 'Saving' : save.kind === 'failed' ? 'Retry save' : 'Save'
  return { saved, chosen, saving, locked, canSave, saveLabel, status: statusLine(save, chosen, saved, status), failed: save.kind === 'failed', choose, submit }
}

/** The status line for each review state. */
function statusLine(save: SaveState, chosen: Verdict | null, saved: Verdict | null, status: LoadStatus): string {
  if (save.kind === 'saving') return `Saving ${verdictLabel(save.verdict)}...`
  if (save.kind === 'failed') return saveFailed(save.message)
  // A committed write's outcome stays shown whatever the refetch does: the server's state has moved.
  const afterOutcome = (line: string) => (save.kind === 'done' ? `${outcome(save)} ${line}` : line)
  if (status === 'error') return afterOutcome('The verdict controls are unavailable until this reading loads.')
  if (chosen) {
    // Choosing needs a loaded detail, which carries the saved verdict.
    if (saved === null) throw new Error(`verdict ${chosen} chosen with no saved verdict`)
    return `Unsaved: ${verdictLabel(saved)} -> ${verdictLabel(chosen)}. Press Save to record it.`
  }
  if (save.kind === 'done') return outcome(save)
  return 'Choose a verdict, then press Save.'
}

/** What a committed save did, as the status line reports it. */
function outcome(save: Extract<SaveState, { kind: 'done' }>): string {
  const verdict = verdictLabel(save.verdict)
  return save.unchanged ? `No change: the verdict was already ${verdict}.` : `Saved: ${verdict}.`
}
