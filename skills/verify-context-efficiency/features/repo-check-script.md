# Repo check script

`scripts/check.sh` is the required local and CI gate for Markdown, SKILL frontmatter, internal links, and bootstrap paths.

## Sub-features

- `check-exits-zero` — script completes with `check: all gates passed`.
- `ci-workflow` — `.github/workflows/check.yml` invokes the same script.

## How to get to it (user POV)

- From repo root, run `./scripts/check.sh` before pushing or opening a PR.

## Driving it with verify-context-efficiency

Preconditions:

- `chmod +x scripts/check.sh` if needed.
- Doctor exits `0`.

Steps:

1. Run `./scripts/check.sh` — exit `0`; stdout includes `check: all gates passed`.
2. Optional: run `skills/verify-context-efficiency/bin/verify-context-efficiency-ci` — exit `0`.

Observable result: CI job `markdown-and-links` would pass on the same commit.

## Gotchas

- Adding a `SKILL.md` without YAML `name` and `description` fails the check.
- Broken relative Markdown links fail the check.
