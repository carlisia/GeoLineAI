import Box from '@mui/material/Box'
import { advisoryLabel, verdictLabel } from '../lib/labels'
import type { Advisory, Verdict } from '../lib/vocabulary.gen'
import { advisoryColor, verdictColor } from './status'

/** The advisory as text in its colour. */
export function AdvisoryBadge({ advisory }: { advisory: Advisory }) {
  return (
    <Box component="span" sx={{ fontWeight: 600, color: advisoryColor(advisory) }}>
      {advisoryLabel(advisory)}
    </Box>
  )
}

/** A verdict's status dot, shared by the badge and the saved-verdict toggle. */
export function VerdictDot({ verdict }: { verdict: Verdict }) {
  return <Box component="span" aria-hidden sx={{ width: 8, height: 8, borderRadius: '50%', bgcolor: verdictColor(verdict), flexShrink: 0 }} />
}

/** Dot plus text; the text carries the meaning. */
export function VerdictBadge({ verdict }: { verdict: Verdict }) {
  return (
    <Box component="span" sx={{ display: 'inline-flex', alignItems: 'center', gap: 0.75, fontWeight: 600, color: verdictColor(verdict) }}>
      <VerdictDot verdict={verdict} />
      {verdictLabel(verdict)}
    </Box>
  )
}
