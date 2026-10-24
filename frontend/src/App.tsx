import type { ReactNode } from 'react'
import Box from '@mui/material/Box'
import FormControlLabel from '@mui/material/FormControlLabel'
import Radio from '@mui/material/Radio'
import RadioGroup from '@mui/material/RadioGroup'
import Stack from '@mui/material/Stack'
import Typography from '@mui/material/Typography'
import { fetchScenarios, type ReadingDetail } from './lib/api'
import type { Scenario } from './lib/scenarios.gen'
import { useReadingSession } from './lib/useReadingSession'
import type { useReview } from './lib/useReview'
import { useSupersedingLoad } from './lib/useSupersedingLoad'
import { DiscardPrompt } from './components/DiscardPrompt'
import { ExplainPanel } from './components/ExplainPanel'
import { ReadingPanel } from './components/ReadingPanel'
import { ReviewPanel } from './components/ReviewPanel'
import { StateMessage } from './components/StateMessage'
import { colors, panel } from './theme'

/**
 * The one page: pick a sample reading, inspect it, ask about its checks, and save a
 * verdict. A save in flight locks the picker; an unsaved choice asks before a new
 * reading replaces it, and the browser warns before the page is left with one.
 */
export function App() {
  const scenarios = useSupersedingLoad('scenarios', fetchScenarios)
  return (
    <Box sx={{ minHeight: '100vh', bgcolor: colors.surface }}>
      <Box component="header" sx={{ bgcolor: colors.panel, borderBottom: `1px solid ${colors.border}`, px: { xs: 2, sm: 3 }, py: 2 }}>
        <Typography variant="h1">GeoLineAI</Typography>
        <Typography variant="body2" color="textSecondary">
          Evidence-grounded assistance for geophysical review.
        </Typography>
      </Box>
      <Box component="main" sx={{ px: { xs: 2, sm: 3 }, py: 2, maxWidth: 1400, mx: 'auto' }}>
        {scenarios.status === 'ready' && scenarios.data.scenarios.length > 0 ? (
          <Reviewer scenarios={scenarios.data.scenarios} />
        ) : (
          <Picker>
            {scenarios.status === 'error' ? (
              <StateMessage kind="error" title={`The sample readings could not be loaded: ${scenarios.error}`} action={{ label: 'Retry', run: scenarios.retry }} />
            ) : (
              <StateMessage kind="loading" title="Loading the sample readings..." />
            )}
          </Picker>
        )}
      </Box>
    </Box>
  )
}

/** The sample-reading panel, around whatever it holds. */
function Picker({ children }: { children: ReactNode }) {
  return (
    <Box component="section" aria-labelledby="picker-title" sx={{ ...panel, mb: 2 }}>
      <Typography id="picker-title" variant="h3" component="h2">
        Sample reading
      </Typography>
      {children}
    </Box>
  )
}

/** The picker, the selected reading and its review, drawn from the reading session. */
function Reviewer({ scenarios }: { scenarios: Scenario[] }) {
  const session = useReadingSession(scenarios)
  return (
    <>
      <Picker>
        <RadioGroup row aria-labelledby="picker-title" value={session.selected} onChange={(e) => session.pick(e.target.value)}>
          {scenarios.map((s) => (
            <FormControlLabel key={s.reading_id} value={s.reading_id} control={<Radio />} label={s.label} disabled={session.locked} />
          ))}
        </RadioGroup>
        {session.locked && (
          <Typography variant="body2" color="textSecondary" role="status">
            A save is in progress; the sample readings unlock when it finishes.
          </Typography>
        )}
      </Picker>
      <ReadingView detail={session.detail} review={session.review} />
      <DiscardPrompt open={session.prompt.open} onDiscard={session.prompt.discard} onKeep={session.prompt.keep} />
    </>
  )
}

/**
 * The selected reading: its evidence, the question panel and the review. The question
 * panel is mounted afresh for each reading, so no answer outlives its reading.
 */
function ReadingView({ detail, review }: { detail: ReturnType<typeof useSupersedingLoad<ReadingDetail>>; review: ReturnType<typeof useReview> }) {
  const shown = detail.shown
  if (shown === null) {
    return detail.status === 'error' ? (
      <Box sx={panel}>
        <StateMessage kind="error" title={`This reading could not be loaded: ${detail.error}`} action={{ label: 'Retry', run: detail.retry }} />
      </Box>
    ) : (
      <Box sx={panel}>
        <StateMessage kind="loading" title="Loading the reading..." />
      </Box>
    )
  }
  return (
    <Stack direction={{ xs: 'column', lg: 'row' }} spacing={2} sx={{ alignItems: { lg: 'flex-start' } }}>
      <Box component="section" aria-label="Reading and evidence" sx={{ ...panel, flex: 3, minWidth: 0 }}>
        {detail.status === 'error' && (
          <Typography variant="body2" role="alert" color="error" sx={{ mb: 1 }}>
            Refreshing this reading failed: {detail.error} The values shown are from the last successful load.
          </Typography>
        )}
        <ReadingPanel detail={shown} />
      </Box>
      <Stack spacing={2} sx={{ flex: 2, minWidth: 0, width: '100%' }}>
        <Box component="section" aria-labelledby="ask-title" sx={{ ...panel, borderColor: colors.model }}>
          <Typography id="ask-title" variant="h3" component="h2" sx={{ color: colors.modelDark, mb: 1 }}>
            Ask about these checks
          </Typography>
          <ExplainPanel key={shown.reading_id} detail={shown} />
        </Box>
        <Box component="section" aria-labelledby="review-title" sx={panel}>
          <Typography id="review-title" variant="h3" component="h2" sx={{ mb: 1 }}>
            Human verdict
          </Typography>
          <ReviewPanel review={review} detail={shown} />
        </Box>
      </Stack>
    </Stack>
  )
}
