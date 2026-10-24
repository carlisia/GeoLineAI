# Sourced by setup.sh and start.sh from the repository root: exports every setting in
# .env when the file exists, so both scripts see the same database path and port.
if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi
