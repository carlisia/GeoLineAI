#!/usr/bin/env bash
# Builds GeoLineAI, runs every automated check, then exercises the user journey over
# HTTP against a fresh temporary database on an unused port. It never reads .env, so
# the server runs without a provider key and explanations must answer unavailable.
# Everything it builds and runs lives in one temporary directory, so the repository's
# own build output and database are never touched. It stops only the server it started,
# and discards only the directory it created.
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)
cd "$root"

echo "verify: automated checks"
scripts/check.sh

work=$(mktemp -d "${TMPDIR:-/tmp}/geolineai-verify.XXXXXX")
pid=""
cleanup() {
  if [[ -n $pid ]] && kill -0 "$pid" 2>/dev/null; then
    kill "$pid"
    wait "$pid" 2>/dev/null || true
  fi
  if command -v trash >/dev/null; then
    trash "$work"
  else
    python3 -I -c 'import shutil, sys; shutil.rmtree(sys.argv[1])' "$work"
  fi
}
trap cleanup EXIT

# The server finds frontend/dist and data/scenarios.json relative to its working
# directory, so it runs from a temporary root laid out like the repository.
app=$work/root
mkdir -p "$app/data"
cp data/line-1.csv data/scenarios.json "$app/data/"
echo "verify: building into $work"
(cd frontend && npm run -s build -- --outDir "$app/frontend/dist" --emptyOutDir >"$work/frontend-build.log" 2>&1) || { cat "$work/frontend-build.log" >&2; exit 1; }
(cd backend && go build -o "$work/geolineai" .)
cd "$app"

export GEOLINEAI_DB_PATH="$work/geolineai.db"
unset GEOLINEAI_OPENROUTER_API_KEY
echo "verify: importing data/line-1.csv into a fresh database"
"$work/geolineai" import data/line-1.csv >"$work/import.log" 2>&1 || { cat "$work/import.log" >&2; exit 1; }
"$work/geolineai" scenarios

# The port is free when picked but released before the server binds it, so another
# process can take it in between; that one failure retries on a new port.
for _ in 1 2 3; do
  port=$(python3 -I -c 'import socket; s = socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1])')
  GIN_MODE=release GEOLINEAI_HTTP_PORT=$port "$work/geolineai" >"$work/server.log" 2>&1 &
  pid=$!
  for _ in $(seq 50); do
    curl -sf "http://127.0.0.1:$port/_/health" >/dev/null && break 2
    kill -0 "$pid" 2>/dev/null || break
    sleep 0.2
  done
  kill "$pid" 2>/dev/null || true
  wait "$pid" 2>/dev/null || true
  pid=""
  grep -q 'is the port in use' "$work/server.log" || break
done
[[ -n $pid ]] || { echo "verify: the server did not start; see $work/server.log" >&2; cat "$work/server.log" >&2; exit 1; }

echo "verify: exercising the flow on http://127.0.0.1:$port"
python3 -I - "http://127.0.0.1:$port" "$GEOLINEAI_DB_PATH" <<'PY'
import json, sqlite3, sys, urllib.error, urllib.request

base, db = sys.argv[1], sys.argv[2]
failures = []


def call(method, path, body=None):
    data = None if body is None else json.dumps(body).encode()
    req = urllib.request.Request(base + path, data=data, method=method, headers={"Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(req) as r:
            return r.status, r.headers.get("Content-Type", ""), r.read()
    except urllib.error.HTTPError as e:
        return e.code, e.headers.get("Content-Type", ""), e.read()


def check(ok, what):
    print(("  ok    " if ok else "  FAIL  ") + what)
    if not ok:
        failures.append(what)


def counts():
    c = sqlite3.connect(f"file:{db}?mode=ro", uri=True)
    try:
        return c.execute("SELECT (SELECT COUNT(*) FROM review_actions), (SELECT COUNT(*) FROM review_history), (SELECT group_concat(qc_review) FROM readings)").fetchone()
    finally:
        c.close()


status, _, body = call("GET", "/_/health")
health = json.loads(body)
check(status == 200 and health["status"] == "ok" and health["imports"] == 1 and health["fingerprint"].startswith("sha256:"), "health reads the database, one import, a configuration fingerprint")

status, _, body = call("GET", "/api/scenarios")
scenarios = json.loads(body)["scenarios"]
check(status == 200 and [s["label"] for s in scenarios] == ["Passing reading", "Violated criterion", "Disagreeing repeats"], "three scenarios, in manifest order")
ids = [s["reading_id"] for s in scenarios]
details = [json.loads(call("GET", f"/api/readings/{i}")[2]) for i in ids]
check([d["advisory"]["value"] for d in details] == ["pass", "fail", "needs_review"], "the scenarios' advisories are Pass, Fail and Needs review")
check([d["row_number"] for d in details] == [2, 262, 1065], "the scenarios resolve to rows 2, 262 and 1065")

before = counts()
status, _, body = call("POST", f"/api/readings/{ids[1]}/explanations", {"question": "Why did it fail?"})
answer = json.loads(body)
check(status == 200 and answer["outcome"] == "unavailable" and answer["reason"] == "no_key" and answer["sources"], "without a key, an explanation is unavailable (no_key) and still offers its sources")
check(all(s["doc"] is None or s["doc"].startswith("docs/") for s in answer["sources"]), "every documentation source names a repository document")
check(counts() == before, "asking wrote nothing")

target = ids[0]
status, _, body = call("POST", "/api/reviews", {"targets": [{"reading_id": target}], "verdict": "fail"})
saved = json.loads(body)
check(status == 200 and saved["affected_reading_ids"] == [target] and saved["action_id"], "a save of Fail on one reading records an action")
detail = json.loads(call("GET", f"/api/readings/{target}")[2])
check(detail["verdict"] == "fail" and len(detail["history"]) == 1 and detail["history"][0]["reason"].endswith("This differs from the advisory."), "the reading shows Fail, one history entry, and the server-built reason")
status, _, body = call("POST", "/api/reviews", {"targets": [{"reading_id": target}], "verdict": "fail"})
check(json.loads(body)["action_id"] is None, "saving the same verdict again records no action")
status, _, body = call("POST", "/api/reviews", {"targets": [{"reading_id": target}], "verdict": "unreviewed"})
detail = json.loads(call("GET", f"/api/readings/{target}")[2])
check(detail["verdict"] == "unreviewed" and detail["reason"] is None and len(detail["history"]) == 2 and detail["history"][1]["reason"] is not None, "Unreviewed appends an entry with no reason and keeps the earlier one")

before = counts()
for body, what in [
    ({"targets": [{"reading_id": ids[0]}, {"reading_id": ids[1]}], "verdict": "pass"}, "two targets"),
    ({"targets": [], "verdict": "pass"}, "no target"),
    ({"targets": [{"reading_id": ids[0]}], "verdict": "maybe"}, "an invalid verdict"),
    ({"targets": [{"reading_id": "007"}], "verdict": "pass"}, "an invalid ID"),
    ({"targets": [{"reading_id": ids[0]}], "verdict": "pass", "reason": "mine"}, "a client-written reason"),
]:
    status, _, _ = call("POST", "/api/reviews", body)
    check(status == 400, f"a save with {what} is refused")
status, _, _ = call("POST", "/api/reviews", {"targets": [{"reading_id": "99999999"}], "verdict": "pass"})
check(status == 404, "a save of a missing reading is not found")
check(counts() == before, "no refused save wrote anything")

for method, path in [("GET", "/api/lines"), ("GET", "/api/readings?line=line-1"), ("POST", "/api/review-actions/1/undo"), ("POST", "/api/handover-drafts"), ("GET", "/api/investigations"), ("GET", "/api/nope")]:
    status, kind, body = call(method, path, {} if method == "POST" else None)
    check(status == 404 and "json" in kind and json.loads(body)["error"] == "not_found", f"{method} {path} is the JSON not-found envelope")
status, kind, body = call("GET", "/")
check(status == 200 and "html" in kind and b"<title>GeoLineAI</title>" in body, "the page is served at /")

if failures:
    sys.exit(f"verify: {len(failures)} flow checks failed")
print("verify: every flow check passed")
PY
