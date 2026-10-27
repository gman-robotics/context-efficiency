# Repository docs

The project explains its purpose in `README.md` and the architected plan in `docs/plan.md`. Agents must read both before implementing plugin features.

## Sub-features

- `readme-present` — README describes goals and points at `docs/plan.md`.
- `plan-present` — plan documents domain model and five components.

## How to get to it (user POV)

- Open the repository on GitHub or in a local checkout.
- Read `README.md`, then `docs/plan.md`.

## Driving it with verify-context-efficiency

Preconditions:

- Doctor exits `0`.

Steps:

1. Run `test -f README.md && test -f docs/plan.md` — exit `0`.
2. Run `rg -n 'docs/plan.md' README.md` — at least one match (README links the plan).

Observable result: both files exist and README references the plan.

## Gotchas

- Bootstrap PRs must not implement plan features or choose a memory server — verification is docs/check-only.
