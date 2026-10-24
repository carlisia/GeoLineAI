#!/usr/bin/env python3
"""Fails when test code outgrows its budget, and says what to cut.

The budget is test lines per production line, from TEST_BUDGET (default 0.5, which is the
target). A suite already over it can pass a higher TEST_BUDGET as a ratchet and
lower it as tests are cut. Counts the files that git tracks or would track, so new
unstaged tests count; generated files count on neither side.
"""
import os
import pathlib
import subprocess
import sys

DEFAULT_BUDGET = 0.5
SOURCE = (".go", ".ts", ".tsx")
TEST = ("_test.go", ".test.ts", ".test.tsx")
# Fakes, fixtures and helpers that only tests import are test code.
TEST_DIRS = ("backend/internal/testutil/", "backend/internal/core/intake/sourcetest/", "frontend/src/test/")
GENERATED = (".gen.ts", ".golden.ts")

root = pathlib.Path(__file__).resolve().parent.parent
try:
    budget = float(os.environ.get("TEST_BUDGET", DEFAULT_BUDGET))
except ValueError:
    sys.exit(f"test budget: TEST_BUDGET={os.environ['TEST_BUDGET']!r} is not a number")

out = subprocess.run(
    ["git", "ls-files", "-z", "--cached", "--others", "--exclude-standard"],
    cwd=root, capture_output=True, check=True,
).stdout.decode()
files = [f for f in out.split("\0")
         if f.endswith(SOURCE) and not f.endswith(GENERATED) and (root / f).exists()]


def is_test(f):
    return f.endswith(TEST) or f.startswith(TEST_DIRS)


def count(fs):
    return sum(len((root / f).read_text().splitlines()) for f in fs)


test = count(f for f in files if is_test(f))
prod = count(f for f in files if not is_test(f))
ratio = test / prod if prod else 0.0
summary = f"test budget: {test} test / {prod} production lines = {ratio:.2f} (budget {budget}, target 0.5)"
if ratio > budget:
    sys.exit(summary + " -- over: fold test functions into table rows, then cut tests "
             "from the deepest rung up (CODING_STANDARDS.md, Testing depth)")
print(summary)
