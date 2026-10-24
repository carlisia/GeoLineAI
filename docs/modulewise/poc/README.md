# Proofs of concept

Each directory reproduces evidence rows in [evidence.md](../evidence.md) with one command. A script builds the Go component in a temporary copy, so the generated bindings and `main.wasm` never land in the repository, runs it, and removes the copy.

| Directory | Command | Proves | Expected output |
| --- | --- | --- | --- |
| [e27-go-greeter](e27-go-greeter/) | `docs/modulewise/poc/e27-go-greeter/run.sh` | E27, E29 | `"Hello from Go, World!"` |
| [e24-records](e24-records/) | `docs/modulewise/poc/e24-records/run.sh` | E24 through the CLI | a JSON list naming S9 as uncited |
| [e24-records](e24-records/) | `docs/modulewise/poc/e24-records/run-http.sh` | E23, E24 through `composable-http-server` | the same list and `200`, then `400` |

Needs Go 1.27.1 or later and the `composable` and `composable-http-server` binaries on PATH, or their paths in `COMPOSABLE` and `COMPOSABLE_HTTP_SERVER`. The first build downloads componentize-go v0.5.1 through `go tool`. `run-http.sh` starts a server that listens on every interface until the script exits (E60).

These are throwaway: chapter 1's component replaces them.
