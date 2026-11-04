# Laya shadow units

Shadow decision tests live with the Go adapter. `go test ./...` under `adapters/engram` runs them. There is no `laya_live` build tag, and that tag is not a PR gate. `cmd/laya-shadow-report` is not a PR gate. `scripts/check.sh` checks the lock, the run script, the routine-hygiene skill, and `docs/laya-shadow.md`.

## Sub-features

- `shadow-units` — client, refusal, rule, fixture, decide, and report tests pass under `go test ./...`.
- `check-files` — `scripts/check.sh` requires the Laya lock, run script, skill, contract doc, and this feature file.

## How to get to it (user POV)

- Read `docs/laya-shadow.md` for the four paths, refusals, and report exits.
- Unit coverage does not start laya-serve.

## Driving it with verify-context-efficiency

Preconditions:

- Doctor exits `0`.
- Go toolchain on PATH for the unit step.

Steps:

1. Run `(cd adapters/engram && go test ./...)` — exit `0`.
2. Run `./scripts/check.sh` — exit `0`. The stdout includes `check: all gates passed`.

Observable result: the same commit passes `go-adapter` and `markdown-and-links`.

## Gotchas

- A live miss writes the agreement files and exits 2. That is not a PR gate. Do not edit fixtures or the rule to lift a score.
- CI does not execute `adapters/laya/run-laya-serve.sh` and does not pass `-tags laya_live`.
- The agreement files are tracked after a real run. This change does not invent their numbers.
