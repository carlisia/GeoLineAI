import { useRef, useState } from 'react'
import Box from '@mui/material/Box'
import Button from '@mui/material/Button'
import Chip from '@mui/material/Chip'
import Stack from '@mui/material/Stack'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'
import { askExplanation, errorMessage, type ExplanationResponse, type ReadingDetail } from '../lib/api'
import { suggestedQuestions } from '../lib/assist'
import { NOT_MEASURED } from '../lib/labels'
import { MAX_QUESTION } from '../lib/vocabulary.gen'
import { modelButton } from '../theme'
import { WrittenAnswer } from './WrittenAnswer'

type Asked = { question: string } & ({ status: 'asking' } | { status: 'answered'; answer: ExplanationResponse } | { status: 'failed'; error: string })

/**
 * Ask about the reading's criteria. The server writes the answer from the reading's
 * results, its stored values, the rules and the documentation, and shows it only if
 * every sentence passes its checks. One question is asked at a time, and the page
 * mounts a new panel for each reading, so an answer never outlives its reading.
 */
export function ExplainPanel({ detail }: { detail: ReadingDetail }) {
  const [question, setQuestion] = useState('')
  const [asked, setAsked] = useState<Asked | null>(null)
  const canAsk = question.trim() !== '' && asked?.status !== 'asking'
  // Set the moment a question is sent, so a second Enter or click before the page
  // renders again finds it being answered and sends nothing: one question at a time.
  const running = useRef(false)
  const field = useRef<HTMLInputElement>(null)
  const ask = (q: string) => {
    const text = q.trim()
    if (text === '' || running.current) return
    running.current = true
    setQuestion(text)
    setAsked({ question: text, status: 'asking' })
    askExplanation(detail.reading_id, text).then(
      (answer) => {
        running.current = false
        setAsked({ question: text, status: 'answered', answer })
      },
      (error: unknown) => {
        running.current = false
        setAsked({ question: text, status: 'failed', error: errorMessage(error) })
      },
    )
  }
  // A suggestion only offers its question: it fills the field, focused with the cursor
  // at the end, to edit and then send; nothing is asked until the reviewer sends it.
  const offer = (q: string) => {
    setQuestion(q)
    requestAnimationFrame(() => {
      const input = field.current
      if (input === null) return
      input.focus()
      input.setSelectionRange(q.length, q.length)
    })
  }
  return (
    <Box>
      <Typography variant="body2" sx={{ mb: 1, fontWeight: 600 }}>
        {NOT_MEASURED}
      </Typography>
      <Typography variant="body2" color="textSecondary" id="suggestions-label" sx={{ mb: 0.75 }}>
        Suggested questions fill the field; nothing is sent until you press Ask.
      </Typography>
      <Stack direction="row" role="group" aria-labelledby="suggestions-label" sx={{ flexWrap: 'wrap', gap: 1, mb: 1.5 }}>
        {suggestedQuestions(detail).map((q) => (
          <Chip
            key={q}
            label={q}
            size="small"
            variant="outlined"
            onClick={() => offer(q)}
            disabled={asked?.status === 'asking'}
            // A long question wraps rather than hiding its end behind an ellipsis.
            sx={{ height: 'auto', py: 0.5, '& .MuiChip-label': { whiteSpace: 'normal' } }}
          />
        ))}
      </Stack>
      <Stack
        component="form"
        direction={{ xs: 'column', sm: 'row' }}
        spacing={1}
        onSubmit={(e) => {
          e.preventDefault()
          ask(question)
        }}
      >
        <TextField
          size="small"
          fullWidth
          label="Question about these checks"
          value={question}
          onChange={(e) => setQuestion(e.target.value)}
          inputRef={field}
          onKeyDown={(e) => {
            if (e.key !== 'Enter' || e.nativeEvent.isComposing) return
            e.preventDefault()
            if (canAsk) ask(question)
          }}
          slotProps={{ htmlInput: { maxLength: MAX_QUESTION } }}
        />
        {/* Filled in the model violet while it can ask; outlined and disabled otherwise. */}
        <Button type="submit" variant={canAsk ? 'contained' : 'outlined'} disabled={!canAsk} sx={canAsk ? modelButton : undefined}>
          Ask
        </Button>
      </Stack>
      {asked !== null && (
        <Box sx={{ mt: 1.5 }} aria-live="polite">
          <Typography variant="caption" color="textSecondary" component="div">
            Asked: {asked.question}
          </Typography>
          {asked.status === 'asking' && (
            <Typography variant="body2" role="status">
              Writing and checking an answer...
            </Typography>
          )}
          {asked.status === 'failed' && (
            <Typography variant="body2" role="alert" color="error">
              The question could not be asked: {asked.error}
            </Typography>
          )}
          {asked.status === 'answered' && <WrittenAnswer answer={asked.answer} />}
        </Box>
      )}
    </Box>
  )
}
