#!/usr/bin/env bash
# Moves a database and its sidecar files out of the way, into a folder beside it that
# goes to the trash when the trash command exists; nothing is deleted outright. Run
# scripts/setup.sh afterwards: it finds no database at the path, creates one with the
# current schema, and imports the line.
set -euo pipefail

if [[ ${1:-} == --help ]]; then
  echo 'Usage: scripts/recreate-db.sh [database-path]'
  echo 'Stop the server first. Every saved review moves aside with the database.'
  echo 'Defaults to GEOLINEAI_DB_PATH or var/geolineai.db, relative to the repository root.'
  exit 0
fi
if (( $# > 1 )); then
  echo 'Expected at most one database path; see --help.' >&2
  exit 1
fi

root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd -P)
db_path=${1-${GEOLINEAI_DB_PATH-var/geolineai.db}}
case "$db_path" in
  ''|file:*|:memory:)
    echo 'Provide a local database path, not an empty, URI or in-memory one.' >&2
    exit 1 ;;
  /*) ;;
  *) db_path="$root/$db_path" ;;
esac
mkdir -p "$(dirname -- "$db_path")"
db_dir=$(cd -- "$(dirname -- "$db_path")" && pwd -P)
db_path="$db_dir/$(basename -- "$db_path")"

command -v lsof >/dev/null || { echo 'Required command unavailable: lsof' >&2; exit 1; }

for candidate in "$db_path" "$db_path-wal" "$db_path-shm" "$db_path-journal"; do
  if [[ -L "$candidate" || ( -e "$candidate" && ! -f "$candidate" ) ]]; then
    echo "Refusing non-regular file or symlink: $candidate" >&2
    exit 1
  fi
  if [[ -e "$candidate" ]] && lsof -t -- "$candidate" >/dev/null 2>&1; then
    echo "File is in use; stop the server first: $candidate" >&2
    exit 1
  fi
done

# Catch an accidental source-file path before moving anything: every SQLite database
# starts with the same 16-byte header.
if [[ -s "$db_path" && $(head -c 15 -- "$db_path") != 'SQLite format 3' ]]; then
  echo "Refusing a file that is not an SQLite database: $db_path" >&2
  exit 1
fi
if [[ ! -e "$db_path" ]]; then
  echo "No database at $db_path; run scripts/setup.sh to create one."
  exit 0
fi

aside="$db_dir/.replaced-$(basename -- "$db_path")-$(date +%Y%m%d%H%M%S)"
mkdir "$aside"
for candidate in "$db_path" "$db_path-wal" "$db_path-shm" "$db_path-journal"; do
  [[ -e "$candidate" ]] && mv -- "$candidate" "$aside/"
done
printf 'Moved the database at %s aside.\n' "$db_path"
if command -v trash >/dev/null; then
  trash "$aside"
  echo 'The previous database went to the trash.'
else
  echo "The previous database is in $aside."
fi
echo 'Run scripts/setup.sh to create a new database and import the line.'
