# K1: comment on modulewise/composable-runtime #45

Status: hold. Post once chapter 1's component runs; E27 and E24 already prove a stock Go component built with componentize-go runs. Fill in the bracketed parts from the chapter's component.

Post with (sends only the text below the line):

```
sed '1,/^---$/d' \
  docs/modulewise/drafts/k1-issue-45-comment.md \
  | gh-axi issue comment 45 \
    -R modulewise/composable-runtime --body-file -
```

Body (everything below the line):

---

Hi! I have a Go component running in Composable Runtime and would like to contribute it toward this issue, if it's useful.

What I have: a Go function built with [Go and componentize-go versions], exporting a WIT interface with [a record argument / list result], composed through a `config.toml` and called through `composable invoke`. It runs at [commit].

What I found along the way: a stock Go component imports the full WASI 0.2.12 CLI, IO, clocks, filesystem and random set even when it uses none of them, so it needs `imports = ["wasip2"]` with a `wasi:p2` capability; without it, loading fails with `unsatisfied imports`. [anything else from the chapter]

Two questions before I open a PR:

1. Would a single-language Go example under `examples/` help, or would you rather it arrive as one component of the multi-language domain example this issue describes?
2. Is there a layout or build-script convention you'd want it to follow beyond what the existing examples do?

Happy to adjust to whatever fits the project.
