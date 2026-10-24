import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import type { ReadingDetail } from '../lib/api'
import { formatTime, NO_REASON, REASON_CLEARED, reasonSavedWith, reasonToBeSavedWith, recordsNoReason, verdictLabel } from '../lib/labels'
import type { useReview } from '../lib/useReview'
import { colors } from '../theme'
import { VerdictBadge } from './Badges'
import { ReasonNote } from './ReasonNote'
import { VerdictToggles } from './VerdictToggles'

interface ReviewPanelProps {
  review: ReturnType<typeof useReview>
  /** The detail shown: its reasons, its verdict and its history. */
  detail: ReadingDetail
}

/**
 * The human verdict: Pass, Fail or Unreviewed, recorded only by Save. The note shows
 * the reason a save of the chosen verdict would record, which the server builds, or
 * with nothing chosen the reason saved with the current verdict.
 */
export function ReviewPanel({ review, detail }: ReviewPanelProps) {
  return (
    <Box>
      <Stack direction="row" spacing={1} sx={{ alignItems: 'center', mb: 1.5 }}>
        <Typography variant="overline" color="textSecondary">
          Saved verdict
        </Typography>
        {review.saved !== null && <VerdictBadge verdict={review.saved} />}
      </Stack>
      <Stack direction="row" sx={{ alignItems: 'center', flexWrap: 'wrap', gap: 1 }}>
        <VerdictToggles chosen={review.chosen} saved={review.saved} disabled={review.locked} onChoose={review.choose} />
        <Button variant="contained" disabled={!review.canSave} onClick={review.submit}>
          {review.saveLabel}
        </Button>
      </Stack>
      <ReasonFor review={review} detail={detail} />
      <Typography variant="body2" role="status" sx={{ mt: 1, color: review.failed ? colors.fail : colors.textSecondary, fontWeight: review.failed ? 600 : 400 }}>
        {review.status}
      </Typography>
      <Typography variant="h3" component="h3" sx={{ mt: 2.5, mb: 0.5 }}>
        History
      </Typography>
      <History detail={detail} />
    </Box>
  )
}

function ReasonFor({ review, detail }: ReviewPanelProps) {
  const chosen = review.chosen
  if (chosen === null) return detail.reason === null ? null : <ReasonNote label={reasonSavedWith(detail.verdict)} text={detail.reason} />
  if (recordsNoReason(chosen)) return <ReasonNote label="Reason" text={REASON_CLEARED} />
  return <ReasonNote label={reasonToBeSavedWith(chosen)} text={detail.would_save_reasons[chosen]} />
}

/** The reading's verdict changes, newest first, each with its time and saved reason. */
function History({ detail }: { detail: ReadingDetail }) {
  if (detail.history.length === 0) return <Typography variant="body2">Never reviewed: no verdict has been saved for this reading.</Typography>
  return (
    <Box component="ol" sx={{ m: 0, pl: 2.5 }} aria-label="Review history, newest first">
      {detail.history.map((e) => (
        <Box component="li" key={e.action_id} sx={{ mb: 1 }}>
          <Typography variant="body2" sx={{ fontWeight: 600 }}>
            {verdictLabel(e.previous_verdict)} -&gt; {verdictLabel(e.new_verdict)}
          </Typography>
          <Typography variant="body2" color="textSecondary">
            {formatTime(e.recorded_at)} UTC by {e.reviewer}, action {e.action_id}
          </Typography>
          <Typography variant="body2">Reason: {e.reason ?? NO_REASON}</Typography>
        </Box>
      ))}
    </Box>
  )
}
