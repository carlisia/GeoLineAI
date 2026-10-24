# GeoLineAI

Evidence-grounded assistance for geophysical review: one page to pick a sample DCIP reading, inspect its deterministic checks and evidence, ask a model-written, checked question about them, and save a human verdict. [docs/spec.md](docs/spec.md) is the specification; read it in full before changing behavior.

## Priorities

When choices conflict, the higher item wins; equal items resolve to the smaller change.

1. **Correctness.** A save lands on exactly its reading; values, results and evidence are accurate; an explanation is shown only when every sentence passes its checks.
2. **Usability.** The reviewer always knows what is happening and loses nothing: what a save did, why an answer was withheld or unavailable, how to recover.
3. **Simplicity.** The scope in docs/spec.md is the whole scope, plus, on the `modulewise` branch, the chapter being worked in [docs/modulewise/](docs/modulewise/README.md). Do not add features, endpoints, extension points or configuration neither names.

## Invariants

- Deterministic code owns parsing, validation, calculations, the advisory and every saved reason. Models only word explanations from the closed sources and judge sentences; they never calculate, choose an advisory, save a verdict or write the review record.
- Only an explicit human Save changes a verdict. A save names exactly one reading. The review record is append-only, written in one transaction; there is no undo and no bulk write.
- An explanation is shown only when every sentence passes; one failing sentence withholds the whole answer. Unanswerable, withheld and unavailable stay distinct, and a reply missing a required field is unavailable, never a refusal.
- Nothing a model produces is stored or sent back to a model as a source. No accuracy is claimed; the page says "Accuracy not measured for this configuration."
- The dataset is synthetic and generated (scripts/generate-line). Never hand-edit data/line-1.csv; regenerate it and update every number docs/research.md and docs/domain.md state (scripts/dataset-facts.py prints them) and the acceptance test pins.
- A documentation passage the models cite must stay verbatim in its document; a test enforces it.

## Source documents

- **Terms and domain rules:** docs/domain.md. Use its terms in code, UI text, API fields and docs.
- **Evidence for each limit:** docs/research.md.
- **Decisions:** docs/adr/: recreating the database on schema change (0001), the append-only review record (0002), the deterministic advisory (0003), server-built reasons (0004), model text behind checks (0005). Read an ADR end to end before changing what it decides.
- **Model configuration:** backend/internal/core/assist/prompts/README.md. Changing a model, prompt, question, passage, threshold or limit changes the configuration fingerprint; changing the checks, the response schema or the source construction needs an evaluation review too.
- **Code and tests:** apply CODING_STANDARDS.md when writing or reviewing them.

## Commands

| Command | What it does |
| --- | --- |
| `scripts/setup.sh` | Load .env, install, build, import the line into a new database if none exists, resolve the scenarios |
| `scripts/start.sh` | Serve on 127.0.0.1:18080 (loads .env) |
| `scripts/check.sh` | Every static check and test, backend then frontend |
| `scripts/verify.sh` | check.sh, then the HTTP flow on a fresh temporary database and a free port |
| `cd backend && go test ./internal/server -run TestFrontend -update` | Regenerate the frontend's types and golden responses from Go |
| `cd backend && go test ./internal/server -run BrowserHarness -harness-port N` | Serve the real routes with scripted models, for browser checks |

## Hand-off

Report each touched behavior as works, naming the test, command or browser check that proved it, or as incomplete, naming what is missing. Run scripts/verify.sh before hand-off. Check what only the page shows in a browser of your own, against the browser harness or a server on a fresh database and a free port; the user's database at var/geolineai.db and their own browser are theirs.
