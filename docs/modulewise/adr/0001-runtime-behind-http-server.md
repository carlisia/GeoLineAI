# GeoLineAI calls converted code through composable-http-server on localhost

Status: accepted as the target, deferred. Until `composable-http-server` can bind loopback only (E60, contribution K9), the Go server calls each component through `composable invoke`, one process per call, which handles the same record arguments (E24) and opens no port. When K9 lands, the branch moves to the server as decided below.

On the `modulewise` branch the Go server reaches every ModuleWise component through `composable-http-server`, run as a second local process, because ModuleWise has no Go bindings (evidence E25) and the HTTP server is the only host entry point that keeps one runtime alive across requests while being configured, like the runtime's own examples, by a TOML file. `scripts/setup.sh` builds the server from the pinned upstream commit into the git-ignored `var/tools/` with `cargo install --git --rev --root`, so anyone who clones the branch runs the same build and no machine-wide install is read or changed; `scripts/start.sh` and `scripts/verify.sh` start and stop it with the Go server.

## Considered options

- **`composable invoke` per call:** no second process, but every check starts a process and a runtime. It passes record arguments as JSON as the server does (E24).
- **Embedding a Wasm runtime in Go (wazero or wasmtime-go):** no second process and no Rust toolchain, but no ModuleWise in the path, which defeats the branch's purpose.
- **The `composable-http-server` already on PATH:** works on the author's machine (E26) and nowhere else reproducibly; crates.io has no such crate (E03).

## Consequences

- Setup needs a Rust toolchain and compiles the runtime once.
- The server binds every interface and cannot be told otherwise (E60), unlike GeoLineAI's own server, which binds 127.0.0.1 only. Until a listen address can be configured (contribution K9), the component it serves is reachable from the local network.
- A request crosses a process and an HTTP hop the baseline does not have; each chapter records that cost.
- A WIT record argument travels as a field of the JSON request body, and a missing parameter is a 400 (E24).
- Moving the pin is a one-line change in `setup.sh`, recorded in [evidence.md](../evidence.md).
