import type { ReactNode } from 'react'
import Box from '@mui/material/Box'
import Typography from '@mui/material/Typography'
import { colors } from '../theme'

interface NoteProps {
  label: string
  /** The note's accessible name, its label when not given. */
  name?: string
  /** The left bar's colour: teal for what the reviewer acts on, violet for what a model wrote. */
  accent?: string
  /** The overline's colour, the secondary text when not given. */
  labelColor?: string
  children: ReactNode
}

/** A labelled note: an overline over its content, marked by a coloured left bar. */
export function Note({ label, name = label, accent = colors.primary, labelColor = colors.textSecondary, children }: NoteProps) {
  return (
    <Box role="note" aria-label={name} sx={{ mt: 1.5, px: 1.5, py: 1, border: `1px solid ${colors.border}`, borderLeft: `3px solid ${accent}`, borderRadius: 1 }}>
      <Typography variant="overline" sx={{ color: labelColor }} component="div">
        {label}
      </Typography>
      {children}
    </Box>
  )
}

/** A reason, saved or to be saved, as a note. */
export function ReasonNote({ label, text }: { label: string; text: string }) {
  return (
    <Note label={label}>
      <Typography variant="body2">{text}</Typography>
    </Note>
  )
}
