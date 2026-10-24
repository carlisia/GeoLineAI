import { useId } from 'react'
import Box from '@mui/material/Box'
import Link from '@mui/material/Link'
import Typography from '@mui/material/Typography'
import type { ExplanationResponse, Source } from '../lib/api'
import { citedSources } from '../lib/assist'
import { sourceKindLabel, unavailableMessage } from '../lib/labels'
import { colors, fonts } from '../theme'
import { Note } from './ReasonNote'

/**
 * A written answer after the checks. Shown, its sentences link to the sources they cite;
 * withheld, only the checks' findings show, never the text that failed them;
 * unanswerable or unavailable, a line says which and why. Every source offered stays
 * readable below: the cited ones first, the rest folded, so each citation lands on a
 * card the page shows.
 */
export function WrittenAnswer({ answer }: { answer: ExplanationResponse }) {
  const anchor = useId()
  const ref = (id: string) => `${anchor}-${id}`
  const cited = answer.outcome === 'shown' || answer.outcome === 'withheld' ? citedSources(answer) : []
  const citedIds = new Set(cited.map((s) => s.id))
  const others = answer.sources.filter((s) => !citedIds.has(s.id))
  return (
    <Box sx={{ mt: 1 }}>
      {answer.outcome === 'shown' && (
        // The model violet marks text a model wrote.
        <Note label="Shown: written by a model, every sentence checked against its sources" name="Written answer" accent={colors.model} labelColor={colors.modelDark}>
          {answer.sentences.map((s, i) => (
            <Typography key={i} variant="body2" sx={{ mb: 0.5 }}>
              {s.text}{' '}
              {s.sources.map((id) => (
                <Link key={id} href={`#${ref(id)}`} variant="caption" sx={{ mr: 0.5, fontFamily: fonts.code }} aria-label={`Source ${id}`}>
                  [{id}]
                </Link>
              ))}
            </Typography>
          ))}
          <Typography variant="caption" color="textSecondary" component="p" sx={{ mt: 1 }}>
            Nothing here is saved; no verdict or reason changes.
          </Typography>
        </Note>
      )}
      {answer.outcome === 'withheld' && (
        <Box role="status">
          <Typography variant="body2" sx={{ fontWeight: 600 }}>
            Withheld: a sentence failed a check, so the whole answer is withheld and its text is not shown.
          </Typography>
          {answer.sentences.length === 0 && (
            <Typography variant="body2" color="textSecondary">
              The model wrote no sentence.
            </Typography>
          )}
          {answer.sentences.map(
            (s, i) =>
              s.problems.length > 0 && (
                <Typography key={i} variant="body2" color="textSecondary">
                  Sentence {i + 1}: {s.problems.join(' ')}
                </Typography>
              ),
          )}
        </Box>
      )}
      {answer.outcome === 'unanswerable' && (
        <Typography variant="body2" role="status" sx={{ fontWeight: 600 }}>
          Unanswerable: the model found nothing in the sources that answers this question. The sources are below.
        </Typography>
      )}
      {answer.outcome === 'unavailable' && answer.reason !== null && (
        <Typography variant="body2" role="status" sx={{ fontWeight: 600 }}>
          Unavailable: {unavailableMessage(answer.reason)}
        </Typography>
      )}
      {cited.length > 0 && (
        <Box sx={{ mt: 1.5 }}>
          <Typography variant="overline" color="textSecondary" component="h4">
            {answer.outcome === 'shown' ? 'Sources cited' : 'Sources the withheld answer cited'}
          </Typography>
          {cited.map((s) => (
            <SourceCard key={s.id} source={s} id={ref(s.id)} />
          ))}
        </Box>
      )}
      {others.length > 0 && (
        <Box component="details" sx={{ mt: 1.5, '& summary': { cursor: 'pointer' } }}>
          <Typography component="summary" variant="overline" color="textSecondary">
            {cited.length > 0 ? `Other sources offered (${others.length})` : `All sources offered (${others.length})`}
          </Typography>
          {others.map((s) => (
            <SourceCard key={s.id} source={s} id={ref(s.id)} />
          ))}
        </Box>
      )}
    </Box>
  )
}

/** One source, whole, with where its text comes from. */
function SourceCard({ source, id }: { source: Source; id: string }) {
  return (
    <Box id={id} sx={{ mb: 1, pl: 1, borderLeft: `2px solid ${colors.border}`, scrollMarginTop: 16, '&:target': { borderLeftColor: colors.model, bgcolor: colors.surface } }}>
      <Typography variant="caption" color="textSecondary" component="div">
        [{source.id}] {sourceKindLabel(source.kind)}: {source.title}
      </Typography>
      <Typography variant="body2">{source.text}</Typography>
      {source.doc !== null && (
        <Typography variant="caption" color="textSecondary" component="div" sx={{ fontFamily: fonts.code }}>
          Quoted from {source.doc}
        </Typography>
      )}
    </Box>
  )
}
