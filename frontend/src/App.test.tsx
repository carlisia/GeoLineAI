import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ThemeProvider } from '@mui/material/styles'
import { describe, expect, it } from 'vitest'
import { App } from './App'
import type { ExplanationResponse } from './lib/api'
import { NOT_MEASURED } from './lib/labels'
import { deferred, detailFor, errorResponse, installBackend } from './test/fakeBackend'
import { shownExplanationResponse, unavailableExplanationResponse } from './test/responses.golden'
import { theme } from './theme'

function renderApp() {
  const user = userEvent.setup()
  render(
    <ThemeProvider theme={theme}>
      <App />
    </ThemeProvider>,
  )
  return user
}

/** The reading the page shows, by its acquisition ID heading. */
const readingHeading = (acquisitionId: string) => screen.findByRole('heading', { name: acquisitionId })

/** The question field. */
const questionField = () => screen.getByRole('textbox', { name: 'Question about these checks' })

const shown = shownExplanationResponse as ExplanationResponse
const sentence = shown.sentences[0].text

describe('the page', () => {
  it('offers the three sample readings and shows the first one with its evidence', async () => {
    installBackend()
    renderApp()
    expect(screen.getByRole('heading', { name: 'GeoLineAI', level: 1 })).toBeInTheDocument()
    const picker = await screen.findByRole('radiogroup', { name: 'Sample reading' })
    expect(within(picker).getAllByRole('radio').map((r) => r.closest('label')?.textContent)).toEqual(['Passing reading', 'Violated criterion', 'Disagreeing repeats'])
    expect(within(picker).getByRole('radio', { name: 'Passing reading' })).toBeChecked()
    await readingHeading('line-1-r00001')
    expect(screen.getByRole('group', { name: 'Advisory' })).toHaveTextContent('Fail')
    expect(screen.getByText(/Per-criterion results/)).toHaveTextContent('2 unresolved results (1 violated, 1 review)')
    expect(screen.getByText('Resistivity sign')).toBeInTheDocument()
    expect(screen.getByText('Supporting facts: electrodes, decay samples and import issues')).toBeInTheDocument()
    expect(screen.getByText('Repeat context')).toBeInTheDocument()
    expect(screen.getByText('Documentation the explanations may cite')).toBeInTheDocument()
    expect(screen.getByText(NOT_MEASURED)).toBeInTheDocument()
  })

  it('has no control of an out-of-scope feature', async () => {
    installBackend()
    renderApp()
    await readingHeading('line-1-r00001')
    for (const name of [/undo/i, /export/i, /handover/i, /investigat/i, /filter/i, /sort/i, /bulk/i, /import/i, /upload/i, /select all/i]) {
      expect(screen.queryByRole('button', { name })).toBeNull()
    }
    expect(screen.queryByRole('grid')).toBeNull()
  })

  it('shows a failed reading load as an error with a retry, never as a reading', async () => {
    let calls = 0
    installBackend({ 'GET /api/readings/1': () => (++calls === 1 ? errorResponse('The server could not complete the request.') : detailFor('1', 'line-1-r00001')) })
    const user = renderApp()
    expect(await screen.findByRole('alert')).toHaveTextContent('This reading could not be loaded: The server could not complete the request.')
    expect(screen.queryByRole('heading', { name: 'line-1-r00001' })).toBeNull()
    await user.click(screen.getByRole('button', { name: 'Retry' }))
    await readingHeading('line-1-r00001')
  })

  it('never lets a superseded reading load replace the newer reading', async () => {
    const slow = deferred<unknown>()
    installBackend({ 'GET /api/readings/2': slow.promise })
    const user = renderApp()
    await readingHeading('line-1-r00001')
    await user.click(screen.getByRole('radio', { name: 'Violated criterion' }))
    await user.click(screen.getByRole('radio', { name: 'Disagreeing repeats' }))
    await readingHeading('line-1-r01064')
    slow.resolve(detailFor('2', 'line-1-r00261'))
    await new Promise((r) => setTimeout(r, 20))
    expect(screen.queryByRole('heading', { name: 'line-1-r00261' })).toBeNull()
    expect(screen.getByRole('heading', { name: 'line-1-r01064' })).toBeInTheDocument()
  })
})

describe('asking', () => {
  it('fills the field from a suggestion and sends nothing until Ask', async () => {
    const sent = installBackend()
    const user = renderApp()
    await readingHeading('line-1-r00001')
    const suggestion = screen.getByRole('button', { name: 'What does the repeat difference measure?' })
    await user.click(suggestion)
    expect(questionField()).toHaveValue('What does the repeat difference measure?')
    expect(sent.filter((s) => s.request.endsWith('/explanations'))).toEqual([])
  })

  it('shows a checked answer with each citation linked to a source card on the page', async () => {
    const sent = installBackend({ 'POST /api/readings/1/explanations': shown })
    const user = renderApp()
    await readingHeading('line-1-r00001')
    await user.type(questionField(), 'Why did it fail?{Enter}')
    const answer = await screen.findByRole('note', { name: 'Written answer' })
    expect(answer).toHaveTextContent(sentence)
    expect(screen.getByText(NOT_MEASURED)).toBeInTheDocument()
    const link = within(answer).getByRole('link', { name: 'Source S1' })
    const target = document.getElementById(link.getAttribute('href')!.slice(1))
    expect(target).not.toBeNull()
    expect(target).toHaveTextContent(shown.sources[0].text)
    expect(sent.find((s) => s.request.endsWith('/explanations'))?.body).toEqual({ question: 'Why did it fail?' })
  })

  it('withholds the whole answer, showing the problems and never the rejected text', async () => {
    const withheld: ExplanationResponse = {
      ...shown,
      outcome: 'withheld',
      sentences: [
        { ...shown.sentences[0], problems: [] },
        { text: 'A rejected sentence with 9 readings.', sources: ['S1'], problems: ['The number 9 is not in its sources.'], supported: 0.9, tells_verdict: 0.1 },
      ],
    }
    installBackend({ 'POST /api/readings/1/explanations': withheld })
    const user = renderApp()
    await readingHeading('line-1-r00001')
    await user.type(questionField(), 'Why?{Enter}')
    expect(await screen.findByText(/Withheld: a sentence failed a check/)).toBeInTheDocument()
    expect(screen.getByText('Sentence 2: The number 9 is not in its sources.')).toBeInTheDocument()
    expect(screen.queryByText(/A rejected sentence/)).toBeNull()
    expect(screen.queryByText(sentence)).toBeNull()
    expect(screen.queryByRole('note', { name: 'Written answer' })).toBeNull()
  })

  it('tells unanswerable apart from each kind of unavailable, keeping the sources readable', async () => {
    let answer: ExplanationResponse = { ...shown, outcome: 'unanswerable', sentences: [] }
    installBackend({ 'POST /api/readings/1/explanations': () => answer })
    const user = renderApp()
    await readingHeading('line-1-r00001')
    const field = questionField()
    const ask = async () => {
      await user.clear(field)
      await user.type(field, 'Why?{Enter}')
      await screen.findByText('Asked: Why?')
    }
    await ask()
    expect(await screen.findByText(/^Unanswerable: the model found nothing/)).toBeInTheDocument()
    expect(screen.getByText(`All sources offered (${shown.sources.length})`)).toBeInTheDocument()
    for (const [reason, text] of [
      ['no_key', /Unavailable: Model help is off: the server has no provider key/],
      ['provider_error', /Unavailable: The model service did not answer/],
      ['invalid_response', /Unavailable: The model answered in a form that could not be read/],
    ] as const) {
      answer = { ...(unavailableExplanationResponse as ExplanationResponse), reason }
      await ask()
      expect(await screen.findByText(text)).toBeInTheDocument()
      expect(screen.queryByText(/^Unanswerable/)).toBeNull()
    }
  })

  it('shows a failed request as an error, with no answer', async () => {
    installBackend({ 'POST /api/readings/1/explanations': errorResponse('question must be 1 to 300 characters.', 400) })
    const user = renderApp()
    await readingHeading('line-1-r00001')
    await user.type(questionField(), 'Why?{Enter}')
    expect(await screen.findByRole('alert')).toHaveTextContent('The question could not be asked: question must be 1 to 300 characters.')
    expect(screen.queryByRole('note', { name: 'Written answer' })).toBeNull()
  })

  it('clears the question when the reading changes, and an old answer never appears under the new reading', async () => {
    const slow = deferred<unknown>()
    installBackend({ 'POST /api/readings/1/explanations': slow.promise })
    const user = renderApp()
    await readingHeading('line-1-r00001')
    await user.type(questionField(), 'Why?{Enter}')
    await screen.findByText('Writing and checking an answer...')
    await user.click(screen.getByRole('radio', { name: 'Violated criterion' }))
    await readingHeading('line-1-r00261')
    expect(questionField()).toHaveValue('')
    slow.resolve(shown)
    await new Promise((r) => setTimeout(r, 20))
    expect(screen.queryByText(sentence)).toBeNull()
    expect(screen.queryByText('Asked: Why?')).toBeNull()
  })
})

describe('reviewing', () => {
  const savedDetail = detailFor('1', 'line-1-r00001', {
    verdict: 'pass',
    reason: 'Advisory: Fail, a criterion was violated. Reviewer selected Pass. This differs from the advisory.',
    history: [
      {
        action_id: '7',
        reviewer: 'reviewer',
        recorded_at: '2026-10-23T12:00:00.000Z',
        previous_verdict: 'unreviewed',
        new_verdict: 'pass',
        reason: 'Advisory: Fail, a criterion was violated. Reviewer selected Pass. This differs from the advisory.',
        entry_count: 1,
      },
    ],
  })

  it('writes only on Save, then shows the verdict, reason and history the server returns', async () => {
    let reads = 0
    const sent = installBackend({
      'GET /api/readings/1': () => (++reads === 1 ? detailFor('1', 'line-1-r00001') : savedDetail),
      'POST /api/reviews': { affected_reading_ids: ['1'], unchanged_reading_ids: [], action_id: '7' },
    })
    const user = renderApp()
    await readingHeading('line-1-r00001')
    await user.click(screen.getByRole('button', { name: 'Pass' }))
    expect(screen.getByText(/Unsaved: Unreviewed -> Pass/)).toBeInTheDocument()
    expect(sent.filter((s) => s.request === 'POST /api/reviews')).toEqual([])
    await user.click(screen.getByRole('button', { name: 'Save' }))
    expect(await screen.findByText('Reviewer selected Pass.', { exact: false, selector: 'li *' })).toBeInTheDocument()
    expect(sent.filter((s) => s.request === 'POST /api/reviews').map((s) => s.body)).toEqual([{ targets: [{ reading_id: '1' }], verdict: 'pass' }])
    expect(reads).toBe(2)
    expect(screen.getByText('Unreviewed -> Pass')).toBeInTheDocument()
    expect(screen.getByText(/2026-10-23 12:00:00 UTC by reviewer/)).toBeInTheDocument()
    expect(screen.getByRole('note', { name: 'Reason saved with Pass' })).toBeInTheDocument()
    expect(screen.getByText('Saved: Pass.')).toBeInTheDocument()
  })

  it('locks competing saves and the sample readings while a save is pending', async () => {
    const slow = deferred<unknown>()
    installBackend({ 'POST /api/reviews': slow.promise })
    const user = renderApp()
    await readingHeading('line-1-r00001')
    await user.click(screen.getByRole('button', { name: 'Fail' }))
    await user.click(screen.getByRole('button', { name: 'Save' }))
    expect(await screen.findByRole('button', { name: 'Saving' })).toBeDisabled()
    for (const radio of screen.getAllByRole('radio')) expect(radio).toBeDisabled()
    expect(screen.getByRole('button', { name: /^Pass/ })).toBeDisabled()
    slow.resolve({ affected_reading_ids: ['1'], unchanged_reading_ids: [], action_id: '1' })
    await waitFor(() => expect(screen.getByRole('radio', { name: 'Violated criterion' })).toBeEnabled())
  })

  it('keeps the chosen verdict after a failed save, with an actionable message and a retry', async () => {
    installBackend({ 'POST /api/reviews': errorResponse('The server could not complete the request.') })
    const user = renderApp()
    await readingHeading('line-1-r00001')
    await user.click(screen.getByRole('button', { name: 'Fail' }))
    await user.click(screen.getByRole('button', { name: 'Save' }))
    expect(await screen.findByText('Save failed: The server could not complete the request. Your verdict is kept; retry when ready.')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Fail' })).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('button', { name: 'Retry save' })).toBeEnabled()
  })

  it('asks before a new reading discards an unsaved verdict; asking a question keeps it', async () => {
    installBackend({ 'POST /api/readings/1/explanations': shown })
    const user = renderApp()
    await readingHeading('line-1-r00001')
    await user.click(screen.getByRole('button', { name: 'Fail' }))
    await user.type(questionField(), 'Why?{Enter}')
    await screen.findByRole('note', { name: 'Written answer' })
    expect(screen.getByText(/Unsaved: Unreviewed -> Fail/)).toBeInTheDocument()

    await user.click(screen.getByRole('radio', { name: 'Violated criterion' }))
    const dialog = await screen.findByRole('dialog', { name: 'Discard unsaved verdict?' })
    await user.click(within(dialog).getByRole('button', { name: 'Keep editing' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())
    expect(screen.getByRole('heading', { name: 'line-1-r00001' })).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: 'Passing reading' })).toBeChecked()
    expect(screen.getByText(/Unsaved: Unreviewed -> Fail/)).toBeInTheDocument()

    await user.click(screen.getByRole('radio', { name: 'Violated criterion' }))
    await user.click(within(await screen.findByRole('dialog')).getByRole('button', { name: 'Discard' }))
    await readingHeading('line-1-r00261')
    expect(screen.queryByText(/Unsaved:/)).toBeNull()
  })
})
