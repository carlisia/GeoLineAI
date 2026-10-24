#!/usr/bin/env python3
"""No comment in Go or TypeScript names a spec, an ADR, a standards document, a test, a
script or a file path (CODING_STANDARDS.md, Comments), nor a hook, a symbol of one of the
repository's own Go packages or a parenthesised list of identifiers (Stated assumptions).
Those move and get renamed, and a comment naming one becomes a pointer that leads
nowhere; what a rule is for, and what must hold, is said in the comment itself. Generated
files are skipped: their header must name what regenerates them. Compiler directives
(`//go:`, `/// <reference`) are not comments.
"""
import pathlib, re, subprocess, sys

root = pathlib.Path(__file__).resolve().parent.parent
SOURCE = (".go", ".ts", ".tsx")
GENERATED = (".gen.ts", ".golden.ts")
# The repository's own Go packages, by directory name; the standard library is not
# renamed, so a comment may name its symbols.
PACKAGES = sorted(p.name for p in (root / "backend" / "internal").rglob("*") if p.is_dir())
NAMES = re.compile(
    r"\b[\w-]+\.md\b"                       # a markdown document
    r"|\bADR ?\d"                           # an ADR by number
    r"|\bTest[A-Z]\w*"                      # a Go test
    r"|\b[\w.-]+\.(?:go|ts|tsx|sh|py|sql|json|yaml)\b"  # a source file, script or config
    r"|\b(?:backend|frontend|docs|scripts)/"  # a repository path
    r"|\buse[A-Z]\w*\b"                     # a React hook
    r"|\b(?:" + "|".join(PACKAGES) + r")\.[A-Z]\w*"  # a repository package's symbol
    r"|\([A-Z][a-z]\w*(?:, [A-Z][a-z]\w*)*\)"  # a parenthesised list of identifiers
)
COMMENT = re.compile(r"^\s*(?://|/\*|\*|\{/\*)")  # a JSX comment line starts with {/*
DIRECTIVE = re.compile(r"^\s*(?://go:|///\s*<reference)")

files = [f for f in subprocess.run(["git", "ls-files", "--cached", "--others", "--exclude-standard"],
                                   cwd=root, capture_output=True, text=True, check=True).stdout.split()
         if f.endswith(SOURCE) and not f.endswith(GENERATED) and (root / f).exists()]

named = []
for f in files:
    for number, line in enumerate((root / f).read_text().splitlines(), 1):
        if COMMENT.match(line) and not DIRECTIVE.match(line):
            m = NAMES.search(line)
            if m:
                named.append(f"{f}:{number}: comment names {m.group(0)!r}")
if named:
    print("\n".join(named), file=sys.stderr)
    sys.exit(1)
