#!/usr/bin/env bash
# Evidence E27: a stock Go component runs in composable-runtime once granted wasi:p2.
# Prints "Hello from Go, World!".
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
work="$("$here/../build.sh" "$here" greeter)"
source "$here/../cleanup.sh"
trap 'cleanup_work "$work"' EXIT
cd "$work"
"${COMPOSABLE:-composable}" invoke config.toml -- greeter.greet World
