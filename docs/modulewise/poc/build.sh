#!/usr/bin/env bash
# Builds one proof of concept's Go component in a temporary copy and prints the copy's
# path; the main.wasm and the generated bindings stay there, never in the repository.
# Usage: build.sh <poc directory> <wit world>
set -euo pipefail

src="$(cd "$1" && pwd)"
world="$2"
work="$(mktemp -d)"
cp -R "$src/wit" "$src/export_wit_world" "$src/go.mod" "$src/go.sum" "$src/config.toml" "$work/"
cd "$work"
go tool componentize-go -d wit -w "$world" bindings >&2
# The bindings step drops the tool directive from go.mod (evidence E61); put it back.
go get -tool github.com/bytecodealliance/componentize-go@v0.5.1 >&2
go tool componentize-go -d wit -w "$world" build >&2
echo "$work"
