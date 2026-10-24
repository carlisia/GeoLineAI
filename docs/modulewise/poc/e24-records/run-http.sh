#!/usr/bin/env bash
# Evidence E24, through composable-http-server: the same call as a JSON POST body, then
# the same body without ids, which the server rejects with 400.
# The server listens on every interface while this runs (evidence E60).
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
work="$("$here/../build.sh" "$here" checks)"
port="$(python3 -I -c 'import socket; s = socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1])')"
cd "$work"
printf '[server.api]\ntype = "http"\nport = %s\n\n[server.api.route.problems]\nmethod = "POST"\npath = "/problems"\ncomponent = "checks"\nfunction = "problems"\n\n' "$port" > server.toml
cat config.toml >> server.toml
"${COMPOSABLE_HTTP_SERVER:-composable-http-server}" server.toml > server.log 2>&1 &
server=$!
source "$here/../cleanup.sh"
trap 'kill "$server" 2>/dev/null; wait "$server" 2>/dev/null; cleanup_work "$work"' EXIT
for _ in $(seq 1 40); do
  grep -q 'HTTP server listening' server.log && break
  sleep 0.25
done
url="http://127.0.0.1:$port/problems"
curl -s -w ' %{http_code}\n' -X POST "$url" -H 'content-type: application/json' \
  -d '{"sentence":"Line 1 at E07","cited":[{"id":"S1","text":"a rule"}],"ids":["S1","S9"]}'
curl -s -o /dev/null -w '%{http_code}\n' -X POST "$url" -H 'content-type: application/json' \
  -d '{"sentence":"x","cited":[]}'
