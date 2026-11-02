# Go engram adapter (v3.6.0)

The **context-efficiency-mcp** Go adapter in `adapters/engram/` wraps Engram v3.6.0 for recall, retain, handoff, and scope tools exposed through MCP. CI pins the upstream binary via `adapters/engram/pin.json` and `scripts/ci-prepare-engram.sh`.

## Sub-features

- `go-unit-tests` — `go test ./...` under `adapters/engram` passes.
- `engram-integration` — with `ENGRAM_BIN` set to a v3.6.0 binary, integration tests exercise real Engram I/O.
- `pin-metadata` — `pin.json` documents v3.6.0 archive URLs and SHA256 sums per platform.

## How to get to it (user POV)

- Install or build Engram v3.6.0 (see `docs/backends.md` and `adapters/engram/pin.json`).
- Build the adapter: `(cd adapters/engram && go build -o context-efficiency-mcp ./cmd/context-efficiency-mcp/)`.
- Launch via `plugin/bin/run-context-efficiency-mcp.sh` once `CONTEXT_EFFICIENCY_MCP_BIN` or the default build output exists.

## Driving it with verify-context-efficiency

Preconditions:

- Doctor exits `0`.
- Go toolchain on PATH (`go version` succeeds).
- For integration: export `ENGRAM_BIN` to an executable Engram v3.6.0 binary (CI uses `scripts/ci-prepare-engram.sh`).

Steps:

1. Run `(cd adapters/engram && go test ./...)` — exit `0`.
2. With `ENGRAM_BIN` set, run `./scripts/check-go.sh` — stdout includes `check-go: OK`; integration line runs when `ENGRAM_BIN` is executable.
3. Optional full bundle: `skills/verify-context-efficiency/bin/verify-context-efficiency-ci` — includes the above when `adapters/engram/go.mod` exists.

Observable result: CI job `go-adapter` (and verify bundle) pass on the same commit.

## Gotchas

- Integration tests skip when `ENGRAM_BIN` is unset; local verify without Engram only proves unit tests.
- `run-context-efficiency-mcp.sh` fails fast if the MCP binary was not built — build before expecting stdio MCP startup.
- Engram version must match `pin.json` (`3.6.0`); newer upstream releases require a deliberate pin bump and integration re-run.
