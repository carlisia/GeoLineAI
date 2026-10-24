# K1: comment on modulewise/composable-runtime #45

Status: hold. Post only after evidence E22 (a TinyGo `wasip2` component runs in composable-runtime) is verified, and fill in the bracketed parts from that proof.

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

What I have: a Go function built with TinyGo [version] for the `wasip2` target, exporting a WIT interface with [a record argument / list result], composed through a `config.toml` and called through `composable-http-server`. It runs at [commit].

What I found along the way: [anything that needed a workaround, or "nothing beyond the TinyGo flags in the README"].

Two questions before I open a PR:

1. Would a single-language Go example under `examples/` help, or would you rather it arrive as one component of the multi-language domain example this issue describes?
2. Is there a layout or build-script convention you'd want it to follow beyond what the existing examples do?

Happy to adjust to whatever fits the project.
