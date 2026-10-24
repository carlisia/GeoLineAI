# GeoLineAI calls converted code through composable-http-server on localhost

On the `modulewise` branch the Go server reaches every ModuleWise component through `composable-http-server`, run as a second local process on a loopback port, because ModuleWise has no Go bindings (evidence E25) and the HTTP server is the only host entry point that keeps one runtime alive across requests while being configured, like the runtime's own examples, by a TOML file. `scripts/setup.sh` builds the server from the pinned upstream commit into the git-ignored `var/tools/` with `cargo install --git --rev --root`, so anyone who clones the branch runs the same build and no machine-wide install is read or changed; `scripts/start.sh` and `scripts/verify.sh` start and stop it with the Go server.

## Considered options

- **`composable invoke` per call:** no second process, but every check starts a process and a runtime, and the CLI's argument passing for records is unknown.
- **Embedding a Wasm runtime in Go (wazero or wasmtime-go):** no second process and no Rust toolchain, but no ModuleWise in the path, which defeats the branch's purpose.
- **The `composable-http-server` already on PATH:** works on the author's machine (E26) and nowhere else reproducibly; crates.io has no such crate (E03).

## Consequences

- Setup needs a Rust toolchain and compiles the runtime once.
- A request crosses a process and an HTTP hop the baseline does not have; each chapter records that cost.
- How a WIT record argument travels in a request body is unknown (E24); chapter 1 settles it or files it upstream.
- Moving the pin is a one-line change in `setup.sh`, recorded in [evidence.md](../evidence.md).
