#!/usr/bin/env bash
# Installs the dependencies, builds the app, and imports the shipped line into a new
# database when none exists yet. Running it again rebuilds and leaves an existing
# database, and every saved review in it, untouched. Loads .env when it exists, as
# scripts/start.sh does, so both use the same database. Run from anywhere.
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)
cd "$root"
# shellcheck source=load-env.sh
source scripts/load-env.sh

for tool in go node npm; do
  command -v "$tool" >/dev/null || { echo "setup: $tool is required; see README.md, Prerequisites." >&2; exit 1; }
done

db=${GEOLINEAI_DB_PATH:-var/geolineai.db}
export GEOLINEAI_DB_PATH=$db

# A clean install runs only when the lockfile changed since the last one; the stamp
# lives in node_modules, so it goes when they do.
lock=$(cksum < frontend/package-lock.json)
stamp=frontend/node_modules/.geolineai-lock
if [[ -f $stamp && $(cat "$stamp") == "$lock" ]]; then
  echo "setup: frontend dependencies are up to date"
else
  echo "setup: installing frontend dependencies"
  (cd frontend && npm ci --no-audit --no-fund --silent)
  printf '%s\n' "$lock" >"$stamp"
fi
echo "setup: building the frontend"
(cd frontend && npm run -s build >/dev/null)
echo "setup: building the backend"
(cd backend && go build -o geolineai .)

mkdir -p "$(dirname "$db")"
if [[ -e $db ]]; then
  echo "setup: $db exists; keeping it and its reviews (nothing is reimported)"
else
  echo "setup: importing data/line-1.csv into $db"
  backend/geolineai import data/line-1.csv
fi

# Opening the database checks its schema, and resolving the scenarios checks that the
# line was imported exactly once; either failure stops here with its reason.
echo "setup: resolving the sample readings"
if ! backend/geolineai scenarios; then
  echo "setup: $db cannot serve the sample readings. Move it aside with scripts/recreate-db.sh $db, or set GEOLINEAI_DB_PATH to a new file, then run setup again." >&2
  exit 1
fi
echo "setup: done. Start the app with scripts/start.sh"
