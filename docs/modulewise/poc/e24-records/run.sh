#!/usr/bin/env bash
# Evidence E24, through the CLI: a string, a list of records and a list of strings in,
# a list of strings out. Prints the list naming S9 as uncited.
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
work="$("$here/../build.sh" "$here" checks)"
source "$here/../cleanup.sh"
trap 'cleanup_work "$work"' EXIT
cd "$work"
"${COMPOSABLE:-composable}" invoke config.toml -- checks.problems \
  '"Line 1 at E07"' '[{"id":"S1","text":"a rule"}]' '["S1","S9"]'
