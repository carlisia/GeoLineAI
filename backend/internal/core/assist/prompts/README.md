# Prompts and settings of the model help

Everything the explanation feature sends a model, and every value that tunes it, in one place, by the model it goes to: Jev, TypeSafe's judgment model, which answers typed questions with probabilities, and the LLM, the text model that writes answers. The files are built into the binary with `go:embed`; nothing here is read at run time from outside the binary. Their text is what a model receives: a prompt file is sent without its final newline, and a question file is the question as Jev reads it.

## Accuracy

No accuracy has been measured for this configuration. The page says "Accuracy not measured for this configuration." and claims no figure. Regression tests establish what the application does with model answers; they do not establish how often a shown answer is right.

## Configuration fingerprint

The fingerprint identifies the configuration. It is a SHA-256 over the prompt, the judgment questions, the passages, both models and the LLM's temperature, the two Jev thresholds, and the sentence, question and blank-answer retry limits, keyed by logical names rather than by file, folder or settings section, in an encoding that orders every map by key: moving a file, renaming a section, reordering a file's keys or reindenting it leaves it unchanged, while any change in what a model receives changes it. The health check reports it, and so does:

```bash
go -C backend run ./internal/core/assist/prompts/fingerprint
```

## Jev

### Files, in `jev/`

| File | What it asks |
| --- | --- |
| `judge-supported.json` | Of every written sentence, over only the sources it cites: is every claim supported |
| `judge-verdict.json` | Of every written sentence: does it tell the reviewer a verdict |

### Settings, in `[jev]`

| Setting | Value | What it controls |
| --- | --- | --- |
| `model` | `typesafe/jev-1.13` | The judgment model |
| `supported_min` | `0.5` | A sentence passes only with a supported probability at least this |
| `verdict_max` | `0.5` | A sentence passes only with a tells-a-verdict probability below this |

## LLM

### Files, in `llm/`

| File | What it controls |
| --- | --- |
| `explain-system.txt` | The system prompt: what an answer about one reading's criteria may say, at most how many sentences, citing which sources |

### Settings, in `[llm]`

| Setting | Value | What it controls |
| --- | --- | --- |
| `model` | `anthropic/claude-sonnet-5.5` | The text model |
| `temperature` | `0` | The text model's sampling temperature |
| `question_characters` | `300` | The longest question about a reading, after trimming; also the page's input limit |
| `explain_sentences` | `4` | The most sentences a shown explanation may have |
| `blank_answer_retries` | `1` | How many more times the text model is asked when it claims an answer but writes no sentence; part of the fingerprint, since it changes which answers are shown |
| `timeout_seconds` | `60` | How long one model call may take, Jev's included |
| `retry_statuses` | `[429, 529]` | Answers retried: rate-limited and overloaded, Jev's included |
| `retries` | `1` | How many times such an answer is retried |
| `retry_wait_seconds` | `1` | The wait before each retry |

The four call settings (`timeout_seconds` and the three retry settings) tune how the one client reaches both models and are deliberately not part of the fingerprint.

## Sources

| File | What it holds |
| --- | --- |
| `sources/explain-passages.json` | The documentation passages an explanation may cite, each quoted verbatim from the document and heading it names; the LLM is given them to cite, and Jev judges a sentence against the ones it cites. A test fails when a document no longer holds its passage. |

## What else needs an evaluation review

These live in code, beside what they check or build, and are not covered by the fingerprint; changing one needs an evaluation review all the same:

- the code checks a written sentence must pass (the identifier and number patterns, the sensitive word stems), in the grounded package;
- the JSON schema the LLM is held to, in the same package;
- the sources an explanation is offered, built by the explain package.
