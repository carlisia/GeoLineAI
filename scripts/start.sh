#!/usr/bin/env bash
# Serves the built app and its API on 127.0.0.1:GEOLINEAI_HTTP_PORT (default 18080),
# over the database at GEOLINEAI_DB_PATH (default var/geolineai.db). Loads .env when it
# exists. It never resets or reimports data; run scripts/setup.sh first.
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)
cd "$root"

# shellcheck source=load-env.sh
source scripts/load-env.sh
db=${GEOLINEAI_DB_PATH:-var/geolineai.db}

[[ -x backend/geolineai && -f frontend/dist/index.html ]] || { echo "start: the app is not built; run scripts/setup.sh first." >&2; exit 1; }
[[ -e $db ]] || { echo "start: $db does not exist; run scripts/setup.sh first." >&2; exit 1; }

exec env GIN_MODE=release backend/geolineai
