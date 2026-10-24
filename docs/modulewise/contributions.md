# Contributions

Changes this branch could offer ModuleWise or Componentized. Each one has to pass one test: ModuleWise would want it even if GeoLineAI did not exist. Claims cite [evidence.md](evidence.md). The modulewise repositories have one author, no CONTRIBUTING file and no merged external pull request (E06, E07, E08), so every candidate starts as an issue or a comment the maintainer can accept or decline before any code is offered; Componentized takes contributions under a DCO (E09). Drafts live in [drafts/](drafts/); the author posts them, and a posted one is linked here.

| ID | Contribution | Venue | Evidence | Comes from | Draft |
| --- | --- | --- | --- | --- | --- |
| K1 | A Go guest example | modulewise/composable-runtime, comment on #45 | E10, E20, E27, E28, E29 | Chapter 1 | [k1-issue-45-comment.md](drafts/k1-issue-45-comment.md), hold until chapter 1's component runs |
| K2 | An outbound authority allowlist for granted `wasi:http` | modulewise/composable-runtime issue, or a `latch-authority` component in componentized/http | E43, E09 | Chapter 3 | Not yet |
| K3 | Correct install instructions for the CLIs | modulewise/composable-runtime issue | E03, E04, E05 | Setup | Not yet |
| K4 | A retry and call-limit interceptor | modulewise/composable-runtime issue | E48 | Chapter 4 | Not yet |
| K5 | Signature verification when pulling from OCI | modulewise/composable-runtime issue | E53 | Chapter 5 | Not yet |
| K6 | A channel transport across processes | modulewise/composable-runtime issue | E46 | None yet | Not yet |
| K7 | Use from a real application for the `[interceptor]` category | comment on modulewise/composable-runtime #53 | E11 | Chapter 4 | Not yet |
| K8 | The repository name in the composable-tools build instructions | modulewise/composable-tools issue | E14 | Setup | Not yet |
| K9 | A configurable listen address for `composable-http-server`, so it can bind loopback only | modulewise/composable-runtime issue | E60 | Chapter 1 | Not yet |

## Why each passes the test

**K1.** #45 asks for components in languages other than Rust (E10), and no Go guest exists in any modulewise repository (E20). A Go component now runs once granted `wasi:p2` (E27, E29); an example showing that, and the capability a Go guest needs, answers part of that issue for every Go user, GeoLineAI or not.

**K2.** A component granted `wasi:http` can reach any host (E43). Limiting a component to named hosts is the least-privilege model the runtime already applies to capabilities, extended one level; any application that calls a third-party API needs it.

**K3.** The tutorial's `cargo install composable-http-server` has no crate behind it (E04), and its `cargo install composable-runtime` installs an older release than the tutorial's code (E05, inferred). Every new reader hits the first.

**K4.** The interceptor model exists (E47, E48) but no retry or call-limit interceptor does. Retrying an unreliable dependency is common to every service that calls a network API.

**K5.** Publishing signs components (E53), but pulling them does not check the signature, so the signature protects nothing at load time. Verifying on pull completes a chain ModuleWise already started.

**K6.** Channels are in-process only (E46). Not needed by GeoLineAI today; listed so the catalog is complete, and offered only if a chapter needs it.

**K7.** #53 is open and designed (E11); a real use case with several intercepted components is input the maintainer can weigh. A comment, not code.

**K8.** A reader following the README clones a repository name that differs from the repository they are reading (E14). Smallest possible fix; whether the old name redirects is still unknown.

**K9.** The server listens on every interface with no way to choose (E60), so anyone running it on a laptop exposes their components to the local network. A `host` key next to `port`, or a loopback default, is a small change any user running it locally benefits from.
