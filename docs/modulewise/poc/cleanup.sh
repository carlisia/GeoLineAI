# Sourced by the run scripts: removes a build copy the way scripts/verify.sh does, to
# the trash when the trash command exists.
cleanup_work() {
  if command -v trash >/dev/null; then
    trash "$1"
  else
    python3 -I -c 'import shutil, sys; shutil.rmtree(sys.argv[1])' "$1"
  fi
}
