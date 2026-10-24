#!/usr/bin/env python3
"""Mutation check: does some test fail when this code is wrong?

    scripts/mutate.py FILE OLD NEW [--escapes]

Replaces OLD with NEW in FILE (OLD must occur exactly once), builds, runs the
tests of FILE's side (backend or frontend), and restores FILE from memory in
every case; no backup file is written. The outcome is one of:

    CAUGHT     the mutant builds and these tests fail   (exit 0)
    SURVIVED   the mutant builds and every test passes  (exit 1)
    INVALID    the mutant does not build: it proves nothing about the tests (exit 2)

A mutant that breaks the build fails every dependent package and reads as caught
though no test ran, so the build runs first and its failure is never counted. The one
exception is a mutant aimed at a type contract, where the build is the check under
test: there INVALID with the expected type error is the proof, so read the error.
--escapes decodes \\n and \\t in OLD and NEW, for multi-line replacements.
"""
import argparse, pathlib, re, subprocess, sys

root = pathlib.Path(__file__).resolve().parent.parent
SIDES = {
    "backend": ([["go", "vet", "./..."]], ["go", "test", "-count=1", "./..."]),
    "frontend": ([["npx", "tsc", "-b"]], ["npx", "vitest", "run"]),
}


def failing_tests(side, out):
    if side == "backend":
        tests = sorted({m.group(1) for m in re.finditer(r"^--- FAIL: (\w+) ", out, re.M)})
        return tests or sorted(set(re.findall(r"^FAIL\t(\S+)", out, re.M)))
    return sorted({re.sub(r"\s+\d+ms$", "", l.split("×", 1)[1].strip()) for l in out.splitlines() if "×" in l})


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("file")
    ap.add_argument("old")
    ap.add_argument("new")
    ap.add_argument("--escapes", action="store_true")
    a = ap.parse_args()
    old, new = a.old, a.new
    if a.escapes:
        old, new = (s.replace("\\n", "\n").replace("\\t", "\t") for s in (old, new))

    path = (pathlib.Path.cwd() / a.file).resolve()
    side = path.relative_to(root).parts[0]
    if side not in SIDES:
        sys.exit(f"mutate: {a.file} is in neither backend/ nor frontend/")
    builds, tests = SIDES[side]
    cwd = root / side

    original = path.read_text()
    if original.count(old) != 1:
        sys.exit(f"mutate: OLD occurs {original.count(old)} times in {a.file}; it must occur exactly once")
    try:
        path.write_text(original.replace(old, new))
        for build in builds:
            r = subprocess.run(build, cwd=cwd, capture_output=True, text=True)
            if r.returncode != 0:
                lines = [l for l in (r.stdout + r.stderr).splitlines() if l.strip() and not l.startswith("#")]
                errors = list(dict.fromkeys(lines))[:3]
                print("INVALID: the mutant does not build, so it says nothing about the tests:")
                print("\n".join("  " + l for l in errors))
                return 2
        r = subprocess.run(tests, cwd=cwd, capture_output=True, text=True)
    finally:
        path.write_text(original)
    if r.returncode == 0:
        print("SURVIVED: the mutant builds and every test passes")
        return 1
    print("CAUGHT by:")
    print("\n".join("  " + t for t in failing_tests(side, r.stdout + r.stderr)))
    return 0


if __name__ == "__main__":
    sys.exit(main())
