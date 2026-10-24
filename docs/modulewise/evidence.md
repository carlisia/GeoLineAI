# Evidence

Every claim the showcase makes about ModuleWise has one row here. The catalogs ([features.md](features.md), [contributions.md](contributions.md)) cite rows by ID; a statement with no row is unsupported. Statuses are defined in [GLOSSARY.md](GLOSSARY.md): **verified**, **stated**, **inferred**, **unknown**, **ask**. A row changes status only with new proof, and the old status stays in its history line.

## Pinned commits

All rows refer to these commits unless a row says otherwise. Code was read in a local clone of `composable-runtime` at `d4d49af`; upstream `14e40b3` is one commit ahead and changes only `Cargo.lock` files (E02), so a code row holds at both.

| Repository | Commit |
| --- | --- |
| [modulewise/composable-runtime](https://github.com/modulewise/composable-runtime) | `14e40b3343f117ea6093e8d4c740ca3a7f280f68` |
| [modulewise/composable-tools](https://github.com/modulewise/composable-tools) | `425eefda12a31171a770c9f05125173d5fc2d468` |
| [modulewise/composable-factory](https://github.com/modulewise/composable-factory) | `4db9dde4f18aacd1087042141751b29dc3a510aa` |
| [modulewise/composable-agents](https://github.com/modulewise/composable-agents) | `164b464c238c4852f9869d93c3e9a44ae97140b4` |
| [modulewise/demos](https://github.com/modulewise/demos) | `7842a99ae4670a29745e894c14826ddbde603642` |
| [componentized/http](https://github.com/componentized/http) | `56882891daafa7d3485cdad8497ea0fc6882d009` |

Shorthand: `CR` is `https://github.com/modulewise/composable-runtime/blob/14e40b3343f117ea6093e8d4c740ca3a7f280f68`.

## Project and process

| ID | Claim | Status | Proof |
| --- | --- | --- | --- |
| E01 | The upstream head of composable-runtime is `14e40b3`. | verified | `git ls-remote https://github.com/modulewise/composable-runtime HEAD` printed `14e40b3343f1...`, 2026-10-24. |
| E02 | `14e40b3` differs from `d4d49af` only in `Cargo.lock` files. | verified | `gh-axi api repos/modulewise/composable-runtime/compare/d4d49af...14e40b3`: ahead by 1; files `Cargo.lock`, `components/Cargo.lock` and four `examples/*/Cargo.lock`. |
| E03 | crates.io has `composable-runtime` 0.4.0 and no `composable-http-server` crate; the workspace version is 0.5.0-alpha.5. | verified | `cargo search composable-http-server` returned nothing; `cargo search composable-runtime` returned `composable-runtime = "0.4.0"` and `composable-interceptor = "0.4.0"`; `Cargo.toml` line 23 is `version = "0.5.0-alpha.5"`. |
| E04 | The 7cs tutorial tells readers to run `cargo install composable-http-server`, which has no crate to install (E03). | verified | `CR/examples/7cs/README.md` line 12. |
| E05 | The 7cs tutorial's `cargo install composable-runtime` installs 0.4.0 from crates.io, not the 0.5.0-alpha.5 code the tutorial sits in. | inferred | From E03 and line 11 of the same README. Whether 0.4.0 runs the tutorial is unknown. |
| E06 | Every commit in the five modulewise repositories has one author. | verified | `git log --format=%an` per repository, author names counted. |
| E07 | No external pull request has been merged into composable-runtime. | verified | `gh-axi pr list --repo modulewise/composable-runtime --state merged --limit 200` listed 91 merged PRs, all by the same author. |
| E08 | The modulewise repositories have no CONTRIBUTING file. | verified | `test -e CONTRIBUTING.md` in each of the five clones. |
| E09 | Componentized accepts contributions under a Developer Certificate of Origin, with no CLA. | verified | `componentized/http` `CONTRIBUTING.md` line 5. |
| E10 | Issue #45 "add polyglot example" is open and asks for a multi-component domain example in languages other than Rust, with a Rust interceptor applied. | verified | `gh-axi issue view 45 -R modulewise/composable-runtime`. |
| E11 | Issue #53 "add interceptor category in toml" is open. | verified | `gh-axi issue view 53 -R modulewise/composable-runtime`. |
| E12 | composable-tools issue #27 "add observability interceptor" is open. | verified | `gh-axi issue list --repo modulewise/composable-tools --state all`. |
| E13 | composable-runtime describes itself as "in an early design phase"; composable-tools as "an early-stage non-production prototype"; composable-agents as "Work in Progress". | verified | The three READMEs at their pinned commits. |
| E14 | The composable-tools README tells readers to clone `composable-mcp`, a name that differs from the repository's. | verified | `composable-tools/README.md`, Build section. Whether the old name redirects is unknown. |

## Host and guest

| ID | Claim | Status | Proof |
| --- | --- | --- | --- |
| E20 | No Go guest component exists in any modulewise repository. | verified | `find <clone> -name '*.go' -o -name go.mod` returned 0 files in each of the five clones. |
| E21 | TinyGo 0.43.0 is installed here and lists a `wasip2` target. | verified | `tinygo version`; `tinygo targets` includes `wasip2`. |
| E22 | A TinyGo `wasip2` component loads and runs in composable-runtime. | unknown | The runtime builds on wasmtime 50 with WASI p3 features enabled (`CR/Cargo.toml` lines 47 and 80-82). Chapter 1's first proof answers this. |
| E23 | `composable-http-server` serves configured routes that call a component function, mapping path segments to arguments. | stated | `CR/examples/http-server/README.md` and `config.toml`. Not yet run here. |
| E24 | `composable-http-server` accepts a request body that maps onto a function taking a WIT record or list argument. | unknown | The example's functions take a string. Chapter 1 needs records (a cited source). |
| E25 | Neither the runtime nor its crates offer Go bindings; the host APIs are the CLI, `composable-http-server`, the MCP server and the Rust library. | inferred | No Go files (E20); the crates are `function-adapter`, `function-invoker`, `http-server`, `interceptor`, `opentelemetry`. |
| E26 | `composable` and `composable-http-server` are installed on this machine from a local clone at `d4d49af`. | verified | `cargo install --list` shows both from the local clone path. This row is about the author's machine only; the branch builds its own pinned copy (ADR 0001). |

## Capabilities

The capability list is this project's own checklist of what a composable runtime could offer an application like GeoLineAI. A row says what public ModuleWise code shows today; "stated" means a public source shows it and it has not been run here.

| ID | Capability | Status | Proof |
| --- | --- | --- | --- |
| E30 | C01 Compose components at startup | stated | `CR/README.md` example: `imports = ["translator"]`; `examples/7cs` steps 1-2. |
| E31 | C02 Explicit callable interfaces and dependency injection | stated | WIT worlds such as `CR/examples/hello-world/wit`; `imports` in config. |
| E32 | C03 Configuration per component instance | stated | `config.greeting = "Bonjour"` in `CR/examples/http-server/config.toml`; `config.locale` in `CR/README.md`. |
| E33 | C04 Interchangeable model clients | ask | No public model-client component found. GeoLineAI scenario: the explanation writer and judge behind `assist.Writer` and `assist.Judge`. |
| E34 | C05 Interchangeable agent implementations | ask | composable-agents has a README only (E13). GeoLineAI has no agent. |
| E35 | C06 Adapt a component function into a tool | stated | composable-tools README: an MCP server "that exposes Wasm Components as Tools". |
| E36 | C07 Aggregate tools into a tool set | stated | composable-tools `components/tools/toolset-factory` and the tool set described in its README. |
| E37 | C08 Expose tools through MCP or compose them directly | stated | MCP hosting: composable-tools README. Direct composition into an agent: ask (E34). |
| E38 | C09 Delegate between agents using different models | ask | No public code. |
| E39 | C10 Host agents behind A2A | ask | No public code. |
| E40 | C11 Automatic runtime telemetry interception | stated | `CR/crates/opentelemetry`, `CR/examples/otel-components`, `CR/examples/otel-service`. Tracing of model or agent calls: ask (E12). |
| E41 | C12 Generate component bytecode through a factory | stated | composable-factory README: `composable_factory::build`. |
| E42 | C13 Generate typed REST adapters and tool schemas | stated | HTTP routes onto functions (E23). An OpenAPI generator: ask. |
| E43 | C14 Restrict outbound access and isolate credentials | verified | Partial. A component gets HTTP only through a declared `wasi:http` capability (`CR/README.md`, stated); once granted, the runtime's send_request hook forwards every request to wasmtime's default_send_request with no host or path check (`CR/src/runtime/host.rs` lines 323-353, read). Componentized `http` has method and scheme latches and no authority latch (`ls components`). Credential isolation: ask. |
| E44 | C15 Separate platform bindings from application logic | stated | Capabilities declared in config, e.g. `[capability.http] type = "wasi:http"` in `CR/README.md`. |
| E45 | C16 Reuse custom-code components through an OCI registry | stated | `CR/src/composition/resolver.rs` `OciResolver` pulls `oci://` references. |
| E46 | C17 Messaging through local channels | stated | `CR/src/messaging/channel.rs`; channel route in `CR/examples/http-server`. Transport across processes: ask. |
| E47 | C18 Context-sensitive authorization interceptors | ask | Interceptor building blocks exist (`CR/crates/interceptor`); no authorization interceptor found. |
| E48 | C19 Retry policies, circuit breakers, call limits and kill switch | ask | Same building blocks; no such interceptor found. |
| E49 | C20 Human escalation with task pause and resume | ask | No public code. |
| E50 | C21 Pluggable agent memory | ask | No public code. |
| E51 | C22 Generated SQL and message-template adapters | ask | No public code. |
| E52 | C23 Distribute prompts, resources and skills as components | ask | No public code. GeoLineAI scenario: its prompts and judgment questions (`backend/internal/core/assist/prompts`). |
| E53 | C24 Signing and provenance for an artifact audit trail | verified | Partial. `grep -n -i 'sign\|verif' CR/src/composition/resolver.rs` matches nothing: pulled components are not signature-checked. Publishing signs: `CR/.github/workflows/publish-component.yml` lines 110-118 run `cosign sign` and attest an SBOM (read). |
| E54 | C25 Install model/config definitions from an artifact catalog | ask | No public code. |
| E55 | C26 An agent that constructs tools, agents and environments | ask | No public code. |
| E56 | C27 Constrained interpreter or code mode | ask | No public code. |

## History

- 2026-10-24: rows E01-E56 created.
