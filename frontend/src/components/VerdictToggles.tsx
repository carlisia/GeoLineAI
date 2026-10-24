import Box from '@mui/material/Box'
import ToggleButton from '@mui/material/ToggleButton'
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup'
import { verdictLabel } from '../lib/labels'
import { VERDICTS, type Verdict } from '../lib/vocabulary.gen'
import { colors } from '../theme'
import { VerdictDot } from './Badges'
import { verdictColor } from './status'

interface VerdictTogglesProps {
  chosen: Verdict | null
  /** The verdict already saved, marked as such. */
  saved: Verdict | null
  disabled: boolean
  onChoose: (verdict: Verdict) => void
}

/**
 * Pass, Fail and Unreviewed: the chosen verdict fills in teal, and the saved one is
 * outlined in its status colour with a "saved" mark. Choosing never saves; only the
 * Save button does.
 */
export function VerdictToggles({ chosen, saved, disabled, onChoose }: VerdictTogglesProps) {
  return (
    <ToggleButtonGroup
      exclusive
      size="small"
      value={chosen}
      disabled={disabled}
      aria-label="Verdict"
      onChange={(_, value: Verdict | null) => value !== null && onChoose(value)}
      sx={{ gap: 1, flexWrap: 'wrap', '& .MuiToggleButtonGroup-grouped': { border: '1px solid', borderRadius: 1 } }}
    >
      {VERDICTS.map((v) => {
        const isSaved = v === saved
        return (
          <ToggleButton
            key={v}
            value={v}
            aria-current={isSaved ? 'true' : undefined}
            sx={{
              gap: 0.75,
              px: 1.5,
              color: colors.text,
              borderColor: isSaved ? verdictColor(v) : colors.inputBorder,
              '&.Mui-selected, &.Mui-selected:hover': { bgcolor: colors.primary, color: colors.panel, borderColor: colors.primary },
            }}
          >
            {isSaved && <VerdictDot verdict={v} />}
            {verdictLabel(v)}
            {isSaved && (
              <Box component="span" sx={{ fontSize: 11 }}>
                (saved)
              </Box>
            )}
          </ToggleButton>
        )
      })}
    </ToggleButtonGroup>
  )
}
