# Show model-written text only behind code and judgment checks; models never decide

A model in GeoLineAI explains; it never decides. [spec.md](../spec.md) holds the outcome contract and the prompts README in backend/internal/core/assist/prompts holds the thresholds; this record holds why they take this shape.

- **Closed sources.** A text model writes an answer only from a closed list of sources that code builds or quotes verbatim (rules from the criteria's own limits, documentation passages, recorded results, facts read from storage), citing them sentence by sentence.
- **Two checks per sentence.** Code checks each sentence against the sources it cites: its citations exist, and its numbers, identifiers, criterion names and sensitive words (recommending, hedging, diagnosing, causal, judging, ranking, excluding) appear in them. Jev, given only those sources, then judges whether the sentence is supported and whether it tells a verdict.
- **All or nothing.** One failing sentence withholds the whole answer, and the page says why without showing it.
- **Refusal is not failure.** A valid reply that the sources cannot answer is Unanswerable. No key, a provider that fails after its retries, or a reply that cannot be read, including one missing its answerable field, is Unavailable with its reason, never a refusal.
- **Nothing stored.** Nothing a model produces is stored or sent back to a model as a source, and the verdict, the advisory and the saved reason stay as ADR 0003 and ADR 0004 decide them.

## Considered options

- **No model; templates only.** Exact by construction, and the facts and rules the sources hold are already templates. But a template answers only the questions it was written for, while a reviewer's questions are phrased freely; the sources are still shown beside every answer, so a withheld answer leaves the exact material in view.
- **A text model with the judgment model as the only guard.** A judgment model alone can score an off-by-one count or an invented rank as supported; literal code checks catch exactly those failures.
- **Dropping the failing sentences and showing the rest:** a paragraph with a claim cut out can read as complete while it is not; withholding the whole answer keeps every shown answer as written.
- **Showing the judgment model's per-sentence probability as a confidence:** it is uncalibrated for this use.

## Consequences

- One provider key, read from the environment; without it explanations are Unavailable with reason no_key, and inspection and saving work as before.
- Some correct answers are withheld, because the code checks are literal; the reviewer sees why and still has the sources.
- No accuracy has been measured for this configuration, and the page says so. The models, prompt, questions, passages, thresholds and limits are one inventory built into the binary, with a fingerprint that identifies it; the code checks, the response schema and the building of sources live in code, outside the fingerprint, and changing one needs an evaluation review all the same.
