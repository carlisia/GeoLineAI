import type { ReactNode } from 'react'
import Box from '@mui/material/Box'
import Stack from '@mui/material/Stack'
import Table from '@mui/material/Table'
import TableBody from '@mui/material/TableBody'
import TableCell from '@mui/material/TableCell'
import TableHead from '@mui/material/TableHead'
import TableRow from '@mui/material/TableRow'
import Typography from '@mui/material/Typography'
import type { ReadingDetail } from '../lib/api'
import {
  advisoryCaption,
  criterionLabel,
  displayBlank,
  formatPercent,
  formatTime,
  issueLabel,
  NOT_ASSESSED,
  NOT_AVAILABLE,
  resultStateLabel,
  UNAVAILABLE,
  unresolvedSummary,
  verdictLabel,
} from '../lib/labels'
import { PASSAGES } from '../lib/documentation.gen'
import type { Column } from '../lib/vocabulary.gen'
import { colors, fonts } from '../theme'
import { AdvisoryBadge } from './Badges'
import { advisoryColor, resultStateColor, resultStateRank } from './status'

/** The selected reading's identity, measured values, advisory and the evidence behind it. */
export function ReadingPanel({ detail }: { detail: ReadingDetail }) {
  return (
    <Stack spacing={2}>
      <Identity detail={detail} />
      <Criteria detail={detail} />
      <Fold title="Measured values" open>
        <ValuesTable
          head={['Name', 'Value']}
          rows={[
            ['Apparent resistivity (ohm-m)', value(detail, detail.apparent_resistivity, 'apparent_resistivity')],
            ['Reported resistivity error (ohm-m)', value(detail, detail.apparent_resistivity_err, 'apparent_resistivity_err')],
            ['Chargeability (V/V)', value(detail, detail.chargeability, 'chargeability')],
            ['Reported chargeability error (V/V)', value(detail, detail.chargeability_err, 'chargeability_err')],
            ['Input current (A)', value(detail, detail.input_current, 'input_current')],
            ['Transmitter-pair contact resistance (ohm)', value(detail, detail.contact_resistance, 'contact_resistance')],
            ['Stack count', value(detail, detail.stacks, 'stacks')],
            ['Array type', value(detail, detail.array_type, 'array_type')],
            ['n factor', value(detail, detail.symmetry, 'symmetry')],
          ]}
        />
      </Fold>
      <Fold title="Supporting facts: electrodes, decay samples and import issues">
        <Typography variant="h3" component="h4" sx={{ mb: 0.5 }}>
          Electrodes
        </Typography>
        <ValuesTable
          head={['Slot', 'Electrode ID', 'Latitude', 'Longitude', 'Altitude (m)']}
          rows={detail.electrodes.map((e) => [
            e.slot,
            value(detail, e.electrode_id, `${e.slot}_id`),
            value(detail, e.lat, `${e.slot}_lat`),
            value(detail, e.lon, `${e.slot}_lon`),
            value(detail, e.alt, `${e.slot}_alt`),
          ])}
        />
        <Typography variant="h3" component="h4" sx={{ mt: 2, mb: 0.5 }}>
          Decay samples (V/V)
        </Typography>
        {detail.decay.samples === null ? (
          <Typography variant="body2">{unavailable(detail, 'decay_curve')}</Typography>
        ) : (
          <ValuesTable head={['Gate', 'Value']} rows={detail.decay.samples.map((s, i) => [String(i + 1), String(s)])} />
        )}
        <Typography variant="h3" component="h4" sx={{ mt: 2, mb: 0.5 }}>
          Import issues
        </Typography>
        {detail.import_issues.length === 0 ? (
          <Typography variant="body2">No import issues are recorded for this reading.</Typography>
        ) : (
          <ValuesTable head={['Field', 'Issue', 'Source text']} rows={detail.import_issues.map((i) => [i.field, issueLabel(i.code), displayBlank(i.source_text)])} />
        )}
      </Fold>
      <Fold title="Repeat context">
        <Repeats detail={detail} />
      </Fold>
      <Fold title="Documentation the explanations may cite">
        <Documentation />
      </Fold>
    </Stack>
  )
}

function Identity({ detail }: { detail: ReadingDetail }) {
  return (
    <Box>
      <Typography variant="h2" component="h2" sx={{ fontFamily: fonts.code, wordBreak: 'break-all' }}>
        {displayBlank(detail.acquisition_id)}
      </Typography>
      <Typography variant="body2" color="textSecondary">
        {detail.line_id} | {detail.timestamp === null ? unavailable(detail, 'timestamp') : `${formatTime(detail.timestamp)} UTC`}
      </Typography>
      <Typography variant="body2" color="textSecondary">
        {detail.source_filename}, row {detail.row_number} | internal reading ID {detail.reading_id}
      </Typography>
      <Stack direction="row" spacing={1} role="group" aria-label="Advisory" sx={{ mt: 1, alignItems: 'baseline', flexWrap: 'wrap' }}>
        <Typography variant="overline" color="textSecondary">
          Deterministic advisory
        </Typography>
        <AdvisoryBadge advisory={detail.advisory.value} />
        <Typography variant="body2" color="textSecondary">
          {advisoryCaption(detail.advisory.value)}
        </Typography>
      </Stack>
    </Box>
  )
}

/**
 * Every criterion result behind the advisory: the unresolved ones first, violations
 * before reviews, then the met ones, then what no criterion assesses.
 */
function Criteria({ detail }: { detail: ReadingDetail }) {
  const ordered = [...detail.advisory.results].sort((a, b) => resultStateRank(a.state) - resultStateRank(b.state))
  const violated = ordered.filter((r) => r.state === 'violated').length
  const review = ordered.filter((r) => r.state === 'review').length
  return (
    <Box component="section" aria-labelledby="criteria-title">
      <Typography id="criteria-title" variant="overline" component="h3" sx={{ fontWeight: 700, color: advisoryColor(detail.advisory.value) }}>
        Per-criterion results: {violated + review === 0 ? 'every criterion met' : unresolvedSummary(violated, review)}
      </Typography>
      <Box component="ul" sx={{ listStyle: 'none', m: 0, p: 0 }}>
        {ordered.map((r, i) => (
          <Box component="li" key={i} sx={{ py: 0.75, borderBottom: `1px solid ${colors.surface}` }}>
            <Typography variant="body2" component="div">
              <Box component="span" sx={{ fontWeight: 600 }}>
                {criterionLabel(r.criterion)}
              </Box>
              {': '}
              <Box component="span" sx={{ color: resultStateColor(r.state), fontWeight: r.state === 'met' ? 400 : 700 }}>
                {resultStateLabel(r.state)}
              </Box>
            </Typography>
            <Typography variant="body2" color="textSecondary">
              {r.explanation}
            </Typography>
          </Box>
        ))}
      </Box>
      <Typography variant="body2" color="textSecondary" sx={{ mt: 1 }}>
        {NOT_ASSESSED}
      </Typography>
    </Box>
  )
}

/** Every documentation passage an explanation may cite, quoted as the models receive it. */
function Documentation() {
  return (
    <Box component="ul" sx={{ listStyle: 'none', m: 0, p: 0 }}>
      {PASSAGES.map((p) => (
        <Box component="li" key={`${p.path}#${p.anchor}:${p.title}`} sx={{ py: 0.75, borderBottom: `1px solid ${colors.surface}` }}>
          <Typography variant="body2" sx={{ fontWeight: 600 }}>
            {p.title}
          </Typography>
          <Typography variant="body2">{p.text}</Typography>
          <Typography variant="caption" color="textSecondary" component="div" sx={{ fontFamily: fonts.code }}>
            Quoted from {p.path}#{p.anchor}
          </Typography>
        </Box>
      ))}
    </Box>
  )
}

/** The reading's own repeat difference above its repeats, each judged against its own. */
function Repeats({ detail }: { detail: ReadingDetail }) {
  if (detail.repeat_group_id === null) return <Typography variant="body2">No repeat group: an electrode slot is blank.</Typography>
  return (
    <>
      <Typography variant="body2" sx={{ mb: 1 }}>
        Repeat difference: {detail.repeat_difference === null ? noDifference(detail) : difference(detail.repeat_difference, detail.disagrees)}
      </Typography>
      {detail.repeats.length === 0 ? (
        <Typography variant="body2">No repeats of this geometry.</Typography>
      ) : (
        <ValuesTable
          head={['Acquisition ID', 'Time (UTC)', 'Resistivity (ohm-m)', 'Chargeability (V/V)', 'Verdict', 'Difference', 'Reference']}
          rows={detail.repeats.map((r) => [
            displayBlank(r.acquisition_id),
            r.timestamp === null ? NOT_AVAILABLE : formatTime(r.timestamp),
            r.apparent_resistivity === null ? NOT_AVAILABLE : String(r.apparent_resistivity),
            r.chargeability === null ? NOT_AVAILABLE : String(r.chargeability),
            verdictLabel(r.verdict),
            difference(r.repeat_difference, r.disagrees),
            r.left_out ? 'Left out' : 'Counts',
          ])}
        />
      )}
    </>
  )
}

/** A repeat difference with its Disagrees mark in the Fail red, or n/a when the reading has none. */
function difference(fraction: number | null, disagrees: boolean): ReactNode {
  if (fraction === null) return NOT_AVAILABLE
  if (!disagrees) return formatPercent(fraction)
  return (
    <Box component="span" sx={{ color: colors.fail, fontWeight: 700 }}>
      {formatPercent(fraction)} Disagrees
    </Box>
  )
}

/**
 * Why the reading has no repeat difference, by its definition: its own resistivity, and
 * two repeats that count toward the reference.
 */
function noDifference(detail: ReadingDetail): string {
  if (detail.apparent_resistivity === null) return `${UNAVAILABLE}: this reading has no resistivity`
  const counting = detail.repeats.filter((r) => !r.left_out).length
  if (counting >= 2) return `${UNAVAILABLE}: no repeat difference was recorded, though ${counting} repeats count toward the reference`
  return `${UNAVAILABLE}: needs two repeats that count toward the reference; this reading has ${counting}`
}

/** "Unavailable", then the field's import-issue labels when it has any. */
function unavailable(detail: ReadingDetail, field: Column): string {
  const labels = detail.import_issues.filter((i) => i.field === field).map((i) => issueLabel(i.code))
  return labels.length ? `${UNAVAILABLE} - ${labels.join(', ')}` : UNAVAILABLE
}

function value(detail: ReadingDetail, v: number | string | null, field: Column): string {
  return v === null ? unavailable(detail, field) : String(v)
}

/** An expandable section; a disclosure is keyboard operable as the browser gives it. */
function Fold({ title, open = false, children }: { title: string; open?: boolean; children: ReactNode }) {
  return (
    <Box component="details" open={open} sx={{ borderTop: `1px solid ${colors.border}`, pt: 1, '& > summary': { cursor: 'pointer', py: 0.5 } }}>
      <Typography component="summary" variant="overline" sx={{ fontWeight: 700 }}>
        {title}
      </Typography>
      <Box sx={{ pt: 1 }}>{children}</Box>
    </Box>
  )
}

/** Scrolls a wide table inside its own box, so a narrow screen never scrolls the page sideways. */
function Scroll({ children }: { children: ReactNode }) {
  return <Box sx={{ overflowX: 'auto', maxWidth: '100%' }}>{children}</Box>
}

function ValuesTable({ head, rows }: { head: string[]; rows: ReactNode[][] }) {
  // Cells wrap on a phone, so a label never pushes its value out of view, and stay on one line from tablet width up.
  const cells = { py: 0.5, pr: 2, pl: 0, whiteSpace: { xs: 'normal', sm: 'nowrap' }, fontVariantNumeric: 'tabular-nums', borderBottom: 0, verticalAlign: 'top' }
  return (
    <Scroll>
      <Table size="small" sx={{ width: 'auto', '& td, & th': cells, '& th': { color: colors.textSecondary, fontWeight: 500 } }}>
        <TableHead>
          <TableRow>
            {head.map((h) => (
              <TableCell key={h}>{h}</TableCell>
            ))}
          </TableRow>
        </TableHead>
        <TableBody>
          {rows.map((r, i) => (
            <TableRow key={i}>
              {r.map((c, j) => (
                <TableCell key={j}>{c}</TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </Scroll>
  )
}
