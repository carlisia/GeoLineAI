# Features

What can be built on GeoLineAI to learn what ModuleWise provides. Conversions come first, in chapter order; extensions come after and may stand on a conversion. Every claim about ModuleWise cites a row in [evidence.md](evidence.md); terms are defined in [GLOSSARY.md](GLOSSARY.md). Paths are relative to `backend/` and refer to the `baseline` tag.

Every chapter records what changed, why, what stayed equivalent, what it cost, and when plain Go would be the better choice. A conversion is done when the function's existing tests pass against both paths, a rapid property test finds no input on which the component and the oracle disagree, and `scripts/verify.sh` passes with the component in the request path.

## Conversions

### Chapter 1: the sentence code checks as a component

- **What moves:** `codeProblems` (`internal/core/assist/grounded/checks.go:123`), called once per sentence from `internal/core/assist/grounded/answer.go:202`. It is pure: a sentence, the sources it cites, every source by ID, and the IDs it names in, the problems out. It reads `survey.Lines` and the criterion names at package level.
- **Capabilities:** C02 typed interface (E31): a WIT world exporting the check, with a record for a source. C01 and C03 stay at their defaults here.
- **Host:** `composable-http-server` on localhost (ADR [0001](adr/0001-runtime-behind-http-server.md)), a route onto the exported function.
- **Guest language:** Go through TinyGo `wasip2` (E21). If it does not load (E22 unknown), the blocker goes upstream with a reproduction and the component is written in Rust.
- **Unknowns the chapter must settle first:** E22 (TinyGo component runs in the runtime) and E24 (a record argument in a request body). Each is a proof of concept before any GeoLineAI code changes.
- **Proof:** the existing tests in `internal/core/assist/grounded` run against both paths; a rapid test generates sentences and sources and compares the component with the oracle; `scripts/verify.sh` passes.
- **Contributions it produces:** K1 (Go guest example), and K3 if the install path is still broken when the chapter starts.

### Chapter 2: composition and configuration

- **What moves:** the check component is split along what varies: the line names (`survey.Lines`) and the sensitive stems (`checks.go:39`) become instance configuration of the check component instead of compiled-in values, and the check is composed with them at startup.
- **Capabilities:** C01 (E30), C03 (E32), C15 (E44).
- **Equivalence:** the configured values are the baseline's values, read from the same Go source by a test, so a drift fails. Changing one is still a check change that needs an evaluation review (`internal/core/assist/prompts/README.md`).
- **Contributions it produces:** none planned; findings about config ergonomics become issues if ModuleWise would want them.

### Chapter 3: the explanation writer through restricted outbound HTTP

- **What moves:** the provider call behind `assist.Writer` (`internal/core/assist/assist.go:45`), today the OpenRouter client in `internal/integrations/openrouter/client.go`, becomes a component that makes the request through a granted `wasi:http` capability. `explain.Ask` (`internal/core/assist/explain/explain.go:50`) already takes the interface, so the seam exists.
- **Capabilities:** C14 (E43, partial): the component should reach only the provider's host. The runtime forwards every granted request today, so the chapter either adds the host check (K2) or records that it cannot be enforced. C04 interchangeable model clients stays an ask (E33).
- **Constraints:** the provider key stays server-side and is never logged; how it reaches the component without being readable elsewhere is part of C14's credential isolation, an ask.
- **Contributions it produces:** K2 (an outbound authority allowlist).

### Chapter 4: interception and tracing

- **What moves:** retries around the provider call, today inside the Go client, become an interceptor around the writer component; the runtime's OpenTelemetry support traces the calls.
- **Capabilities:** C11 (E40), C19 (E48, ask).
- **Contributions it produces:** K4 (a retry and call-limit interceptor); K7 if per-component interceptor wiring (#53, E11) is still open.

### Chapter 5: loading components from an OCI registry

- **What moves:** the branch's components are published to a registry and loaded by `oci://` reference instead of a local path.
- **Capabilities:** C16 (E45), C24 (E53, partial: pulled components are not signature-checked).
- **Contributions it produces:** K5 (signature verification on pull).

## Extensions

### Chapter 6: the checks as a tool and an endpoint

- **What it adds:** the code checks offered to other clients: as an MCP tool through composable-tools, and as a REST route on `composable-http-server`. The baseline's specification rules this out, so it exists only on this branch.
- **Capabilities:** C06 (E35), C07 (E36), C08 (E37), C12 (E41), C13 (E42).
- **Depends on:** chapter 1.

## Asks

Capabilities GeoLineAI could use that no public ModuleWise source provides. Each is prototyped only when a chapter runs into it.

| Capability | Evidence | GeoLineAI scenario |
| --- | --- | --- |
| C04 Interchangeable model clients | E33 | Swap the writer and judge models behind `assist.Writer` and `assist.Judge` without touching `explain.Ask`. |
| C05 Interchangeable agent implementations | E34 | None today: GeoLineAI has no agent, and its specification rules one out. |
| C09 Delegate between agents | E38 | None today. |
| C10 Host agents behind A2A | E39 | None today. |
| C18 Context-sensitive authorization interceptors | E47 | Allow the writer component to send only the closed source set it was given. |
| C19 Retry, circuit breakers, call limits, kill switch | E48 | Chapter 4. |
| C20 Human escalation with pause and resume | E49 | None today: the human verdict is a separate explicit save. |
| C21 Pluggable agent memory | E50 | None: the specification forbids storing model output. |
| C22 Generated SQL and message-template adapters | E51 | The review store (`internal/integrations/db/store.go`). |
| C23 Prompts and skills as components | E52 | The prompts and judgment questions in `internal/core/assist/prompts`, versioned with the configuration fingerprint. |
| C25 Install definitions from an artifact catalog | E54 | Select a writer and judge configuration by reference. |
| C26 An agent that builds tools and environments | E55 | None today. |
| C27 Constrained interpreter or code mode | E56 | None today. |
