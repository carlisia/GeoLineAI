#!/usr/bin/env python3
"""Every relative link in the repo's markdown must reach an existing file and heading.

A link is `[text](path)`, `[text](path#anchor)` or `[text](#anchor)`; external links
(any `scheme:` target) are not fetched. An anchor is checked against the target's
headings, slugged the way GitHub renders them, so renaming a file or a heading that a
doc links to fails here instead of leaving a pointer that leads nowhere. Links inside
code fences and inline code are not links and are skipped.
"""
import pathlib, re, subprocess, sys

root = pathlib.Path(__file__).resolve().parent.parent
# Tracked and new markdown, less any deleted but not committed.
DOCS = [root / f for f in subprocess.run(["git", "ls-files", "--cached", "--others", "--exclude-standard", "*.md"],
                                         cwd=root, capture_output=True, text=True, check=True).stdout.split()
        if (root / f).exists()]
LINK = re.compile(r"\[[^\]]*\]\(([^)\s]+)(?:\s+\"[^\"]*\")?\)")


def prose(text):
    """The lines outside code fences, with inline code removed, numbered from 1."""
    fenced = False
    for number, line in enumerate(text.splitlines(), 1):
        if line.lstrip().startswith("```"):
            fenced = not fenced
        elif not fenced:
            yield number, re.sub(r"`[^`]*`", "", line)


def anchors(path, cache={}):
    """GitHub's heading slugs: lowercased, punctuation dropped, spaces to hyphens, repeats numbered."""
    if path not in cache:
        seen, slugs = {}, set()
        for _, line in prose(path.read_text()):
            m = re.match(r"#{1,6}\s+(.*?)\s*#*\s*$", line)
            if m:
                slug = re.sub(r"[^\w\- ]", "", m.group(1).lower()).replace(" ", "-")
                count = seen.get(slug, 0)
                seen[slug] = count + 1
                slugs.add(f"{slug}-{count}" if count else slug)
        cache[path] = slugs
    return cache[path]


broken = []
for doc in DOCS:
    name = doc.relative_to(root)
    for number, line in prose(doc.read_text()):
        for target in LINK.findall(line):
            if re.match(r"[a-z][a-z0-9+.-]*:", target):
                continue
            path, _, anchor = target.partition("#")
            dest = (doc.parent / path).resolve() if path else doc
            if not dest.exists():
                broken.append(f"{name}:{number}: ({target}) names no file")
            elif anchor and dest.suffix == ".md" and anchor not in anchors(dest):
                broken.append(f"{name}:{number}: ({target}) names no heading in {dest.relative_to(root)}")
if broken:
    print("\n".join(broken), file=sys.stderr)
    sys.exit(1)
