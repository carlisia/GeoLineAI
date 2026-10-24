# GeoLineAI

Evidence-grounded assistance for geophysical review.

This is experimental.

GeoLineAI is a small local application for reviewing DCIP (direct current resistivity and induced polarization) survey readings. On one page you pick a sample reading, inspect its measured values and the deterministic quality checks behind its advisory, ask a question about those checks, and save your own verdict. A model may write the answer to a question, but only from a closed set of sources, and the answer is shown only when every sentence passes code checks and Jev, a judgment model. Models never decide anything: the checks, the advisory, the saved reason and the verdict are code's and yours.

## What it does

- **Three sample readings** from one synthetic survey line: one whose advisory is Pass, one with a violated criterion, and one whose repeats disagree.
- **The evidence:** the reading's values, each criterion's result with its explanation, its electrodes, decay samples, import issues and repeats.
- **Questions about the checks.** An answer is shown, withheld (a sentence failed a check, and the problems are shown instead), unanswerable (the sources hold no answer), or unavailable (no provider key, the provider failed, or its reply could not be read). Every citation links to the source it cites, shown in full on the page.
- **Your verdict:** Pass, Fail or Unreviewed, recorded only when you press Save, with a reason the server builds from the reading's evidence, and an append-only history.

## How it uses AI

Two models take part, both reached through OpenRouter with one provider key, and only when you ask a question:

1. **A text model writes the answer.** Claude Sonnet, at temperature 0, receives the question and a closed, numbered list of sources the server builds for the reading: its criterion results, facts read from storage, rules generated from the limits the code compares against, and documentation passages quoted verbatim. Every sentence must cite a source, or the model must reply that the sources cannot answer.
2. **Code checks every sentence.** Its citations must exist, and every number, identifier, criterion name and sensitive word (recommending, hedging, diagnosing, causal, judging, ranking, excluding) in it must appear in the sources it cites.
3. **Jev judges every sentence.** Jev is TypeSafe's judgment model: instead of writing text, it answers a typed question with a probability. Each sentence goes to Jev with only the sources it cites, and two questions: is every claim in it supported, and does it tell the reviewer what verdict to record. A sentence passes only when it is likely supported and unlikely to tell a verdict.

One failing sentence, by code or by Jev, withholds the whole answer, and the page shows each sentence's problems and the sources instead of the text. The two checks catch different failures: code catches an off-by-one count or an invented identifier that a judgment model can score as supported, and Jev catches a claim the cited sources do not make even when every number in it appears there. Jev's probabilities are not shown as a confidence, because they are not calibrated for this use.

The exact models, the prompt, Jev's questions and thresholds, and the configuration fingerprint the health check reports are in [the prompts README](backend/internal/core/assist/prompts/README.md); [ADR 0005](docs/adr/0005-model-text-behind-checks.md) records why the pipeline takes this shape.

## Prerequisites

- Go 1.26 or later
- Node.js 24 or later, with npm
- Python 3.12 or later, for the repository checks
- lsof, for scripts/recreate-db.sh only

devbox.json pins these for [devbox](https://www.jetify.com/devbox) users: `devbox shell`.

## Setup and start

From the repository root:

```bash
scripts/setup.sh
scripts/start.sh
```

Then open http://127.0.0.1:18080. Setup installs the dependencies, builds the app, imports data/line-1.csv into var/geolineai.db when that file does not exist yet, and checks that the three sample readings resolve. Running it again rebuilds and keeps the existing database and every review in it.

To enable model help, copy .env.example to .env and set GEOLINEAI_OPENROUTER_API_KEY. Without a key, questions answer "unavailable" and everything else works.

## Configuration

| Variable                     | Default          | Meaning                                                                                   |
| ---------------------------- | ---------------- | ----------------------------------------------------------------------------------------- |
| GEOLINEAI_DB_PATH            | var/geolineai.db | The SQLite database, relative to the repository root                                      |
| GEOLINEAI_HTTP_PORT          | 18080            | The port, on 127.0.0.1 only; the server stops with a clear message when the port is taken |
| GEOLINEAI_OPENROUTER_API_KEY | empty            | The provider key for the explanation models; it stays on the server                       |

## A fresh database

To start over with no reviews, either move the database aside (into the trash when the trash command exists) and set up again:

```bash
scripts/recreate-db.sh
scripts/setup.sh
```

or point at a new file and keep the old one:

```bash
GEOLINEAI_DB_PATH=var/fresh.db scripts/setup.sh
GEOLINEAI_DB_PATH=var/fresh.db scripts/start.sh
```

## Verification

```bash
scripts/verify.sh
```

runs every check (documentation links and identifiers, comment rules, the test budget, gofmt, go vet, staticcheck, the Go tests, the TypeScript build, ESLint and the frontend tests), then builds the app, imports the line into a fresh temporary database, serves it on an unused port without a provider key, and checks the user journey over HTTP. It cleans up only what it created.

## Limitations

- **Local and single-reviewer.** The server listens on 127.0.0.1 only. There are no accounts, roles or authorization; every save is recorded under the fixed label "reviewer". Hiding an operation in the page is not access control.
- **A fixed, synthetic dataset.** The one line is generated (see below); there is no import screen or upload.
- **No measured accuracy.** No accuracy has been measured for the current models, prompts and thresholds, and the page says so. Regression tests establish what the application does with model answers, not how often a shown answer is right.
- **Some correct answers are withheld,** because the code checks are literal; the sources stay on the page.

## The dataset

data/line-1.csv is synthetic: a SimPEG forward model of one conductive, chargeable body, with field conditions and defects added by documented, seeded rules. scripts/generate-line holds the generator and how to rerun it. [docs/domain.md](docs/domain.md) describes the line and its defects.

## Documentation

- [docs/spec.md](docs/spec.md): the specification
- [docs/domain.md](docs/domain.md): domain rules, terms, criteria, the dataset and how citations resolve
- [docs/research.md](docs/research.md): the evidence behind each limit
- [docs/adr](docs/adr): the architecture decisions
- [backend/internal/core/assist/prompts/README.md](backend/internal/core/assist/prompts/README.md): the models, prompts and thresholds, and the configuration fingerprint
- [CLAUDE.md](CLAUDE.md) and [CODING_STANDARDS.md](CODING_STANDARDS.md): how to work on the code

## Layout

| Path     | What it holds                                                                                      |
| -------- | -------------------------------------------------------------------------------------------------- |
| backend  | The Go server: Gin, SQLite (pure Go), the criteria, the review record and the explanation pipeline |
| frontend | The React and TypeScript page, built with Vite and MUI                                             |
| data     | The survey line and the sample-reading manifest                                                    |
| docs     | The documentation above                                                                            |
| scripts  | Setup, start, verification, the database reset, the dataset generator and the repository checks    |
