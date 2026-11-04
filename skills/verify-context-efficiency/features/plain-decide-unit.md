# Plain decide units

The plain decision layer lives in `adapters/engram/internal/decide` and the `decide` tool in `internal/front`. `go test ./...` under `adapters/engram` runs them. No model or server is needed.

## Sub-features

- `decide-units` — label, first-match, refusal, and fixture scoring tests pass under `go test ./...`.
- `report` — `cmd/plain-decide-report` writes `reports/plain-decide-agreement.{json,md}`.

## Driving it with verify-context-efficiency

1. Run `(cd adapters/engram && go test ./...)` — exit `0`.
2. Run `(cd adapters/engram && go run ./cmd/plain-decide-report -out ../../reports)` — exit `0`.
3. Run `./scripts/check.sh` — exit `0`.

## Gotchas

- Agreement is 100% by construction (see `docs/plain-decide.md`). It is not an independent result.
