#!/usr/bin/env bash
# Every automated check the repository has, backend then frontend. Run it before
# handing work off; scripts/verify.sh runs it first.
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)

python3 -I "$root/scripts/check-doc-identifiers.py"
python3 -I "$root/scripts/check-doc-links.py"
python3 -I "$root/scripts/check-comments.py"
python3 -I "$root/scripts/test-budget.py"

cd "$root/backend"
unformatted=$(gofmt -l .)
if [[ -n $unformatted ]]; then
  printf 'gofmt: these files need formatting:\n%s\n' "$unformatted" >&2
  exit 1
fi
go vet ./...
go tool staticcheck ./...
go test ./...

cd "$root/frontend"
npx tsc -b
npm run -s lint
npm test
