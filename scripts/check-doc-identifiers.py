#!/usr/bin/env python3
"""Every identifier and path a doc about existing code names in backticks must exist.

Three kinds are checked: snake_case names (columns, fields, parameters), Go names
written as package.Name, and repository paths. A rename that leaves a doc naming
the old one fails here, instead of steering the next reader wrong. Every doc that
describes the code is checked: the README, the specification, the domain notes, the
research notes, the ADRs and the coding standards; each names real code.
"""
import pathlib, re, subprocess, sys

root = pathlib.Path(__file__).resolve().parent.parent
DOCS = [root / "README.md", root / "CODING_STANDARDS.md", *sorted((root / "docs").rglob("*.md"))]
# A doc the repository does not have yet is checked once it is written.
DOCS = [p for p in DOCS if p.exists()]
# Words the docs write in backticks that are not identifiers.
NOT_IDENTIFIERS = {"snake_case"}
# Names a doc cites on purpose although the code no longer has them, with why.
HISTORICAL = {}

# The working tree's code: tracked and new files, less any deleted but not committed.
files = [f for f in subprocess.run(["git", "ls-files", "--cached", "--others", "--exclude-standard"],
                                   cwd=root, capture_output=True, text=True, check=True).stdout.split()
         if (root / f).exists()]
source = {f: (root / f).read_text() for f in files
          if f.startswith(("backend/", "frontend/src")) and f.endswith((".go", ".sql", ".ts", ".tsx"))}
code = "".join(source.values())

# Top-level Go declarations by package name, test files included (docs cite tests).
declared = {}
for f, text in source.items():
    if not f.endswith(".go"):
        continue
    m = re.search(r"^package (\w+)", text, re.M)
    names = declared.setdefault(m.group(1).removesuffix("_test"), set())
    names.update(re.findall(r"^(?:func|type|var|const) (\w+)", text, re.M))
    names.update(re.findall(r"^func \([^)]*\) (\w+)", text, re.M))
    for block in re.findall(r"^(?:var|const|type) \((.*?)^\)", text, re.M | re.S):
        names.update(re.findall(r"^\t(\w+)", block, re.M))

# Standard library package names, so `io.Reader` is not mistaken for a deleted package.
stdlib = {p.rsplit("/", 1)[-1] for p in subprocess.run(["go", "list", "std"], cwd=root / "backend",
                                                     capture_output=True, text=True, check=True).stdout.split()}

stale = []
for doc in DOCS:
    text, name = doc.read_text(), doc.relative_to(root)
    for token in sorted(set(re.findall(r"`([a-z][a-z0-9]*(?:_[a-z0-9]+)+)`", text))):
        if token not in NOT_IDENTIFIERS and token not in HISTORICAL and token not in code:
            stale.append(f"{name}: `{token}` is not in the code")
    # Exported names only: a lowercase one is a table's column (readings.qc_review)
    # or a file (main.go), not the package's API.
    for pkg, ident in sorted(set(re.findall(r"`([a-z]\w*)\.([A-Z]\w*)(?:\(\))?`", text))):
        if f"{pkg}.{ident}" in HISTORICAL or pkg in stdlib:
            continue
        if pkg not in declared:
            stale.append(f"{name}: `{pkg}.{ident}` names no Go package in the repo")
        elif ident not in declared[pkg]:
            stale.append(f"{name}: `{pkg}.{ident}` is not declared in package {pkg}")
    for path in sorted(set(re.findall(r"`((?:backend|frontend|scripts|docs|internal|core|review|intake|integrations)/[\w./-]*\w)`", text))):
        candidates = [root / path, root / "backend" / path, root / "backend/internal" / path, root / "backend/internal/core" / path]
        # A path git ignores is one a run creates (the database, .env), so a fresh
        # clone lacks it and the doc naming it is still right.
        runtime = subprocess.run(["git", "check-ignore", "-q", path], cwd=root).returncode == 0
        if not runtime and not any(c.exists() for c in candidates):
            stale.append(f"{name}: `{path}` does not exist")
if stale:
    print("\n".join(stale), file=sys.stderr)
    sys.exit(1)
