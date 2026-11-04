# Go engram adapter (v3.6.0)

The **context-efficiency-mcp** Go adapter in `adapters/engram/` wraps Engram v3.6.0. Domain MCP tools are **`recall`**, **`retain`**, **`handoff_write`**, **`handoff_read`**, and **`decide`** (see `adapters/engram/internal/front/server.go` and `docs/backends.md`). **`scope`** is a tool argument, not a separate MCP tool. Rule and fixture unit tests (`internal/rules`) are part of `go test ./...`.

CI downloads the pinned linux_amd64 Engram release in **`scripts/ci-prepare-engram.sh`**, reading archive URL and tarball SHA256 from **`adapters/engram/pin.json`** (`archives.linux_amd64`). The adapter binary verifies the child Engram executable against **`pin.json` → `binary_sha256`** at startup (linux v3.6.0 hash in-tree). The PR `go-adapter` job compares the extracted linux binary `adapters/engram/.ci/engram` to `binary_sha256`. Dispatch workflow `.github/workflows/engram-pin-check.yml` repeats that check and does not publish.

## Sub-features

- `go-unit-tests` — `go test ./...` under `adapters/engram` passes.
- `engram-integration` — with `ENGRAM_BIN` set to a v3.6.0 binary whose SHA256 matches `pin.json` `binary_sha256`, integration tests exercise real Engram I/O.
- `pin-metadata` — `pin.json` documents v3.6.0 archive URLs, per-archive tarball SHA256 sums, and the extracted linux_amd64 `binary_sha256`.

## How to get to it (user POV)

- Install or extract Engram v3.6.0 (see `docs/backends.md` and `adapters/engram/pin.json`).
- Build the adapter: `(cd adapters/engram && go build -o context-efficiency-mcp ./cmd/context-efficiency-mcp/)`.
- Set **`ENGRAM_DATA_DIR`** (absolute) and **`ENGRAM_PROJECT`**, and point **`ENGRAM_BIN`** at an Engram binary whose SHA256 equals `pin.json` `binary_sha256` (one linux_amd64 digest; other OS builds fail startup with `binary sha256 mismatch`). Root `mcp.json` sets `ENGRAM_BIN`, so while that key is set the launcher fallback does not run. An unset key makes `plugin/bin/run-context-efficiency-mcp.sh` use `adapters/engram/engram`. The Go process looks in the directory of the running binary, which differs when `CONTEXT_EFFICIENCY_MCP_BIN` points elsewhere.
- Launch via `plugin/bin/run-context-efficiency-mcp.sh` once the adapter build exists.

## Driving it with verify-context-efficiency

Preconditions:

- Doctor exits `0`.
- Go toolchain on PATH (`go version` succeeds).
- For integration: export `ENGRAM_BIN` to an executable Engram v3.6.0 binary (CI uses `scripts/ci-prepare-engram.sh`).

Steps:

1. Run `(cd adapters/engram && go test ./...)` — exit `0`.
2. With `ENGRAM_BIN` set, run `./scripts/check-go.sh` — stdout includes `check-go: integration tests` and `check-go: OK`.
3. Optional full bundle: `skills/verify-context-efficiency/bin/verify-context-efficiency-ci` — includes the above when `adapters/engram/go.mod` exists.

Observable result: CI jobs `go-adapter` and `verify-context-efficiency` (and the verify bundle) pass on the same commit.

## Gotchas

- Integration tests skip when `ENGRAM_BIN` is unset; CI fails closed if integration is skipped.
- Startup fatals with `pin missing binary_sha256` when that field is absent, or `binary sha256 mismatch` when `ENGRAM_BIN` / sibling `engram` does not match the pin.
- `run-context-efficiency-mcp.sh` fails fast if the MCP binary was not built — build before expecting stdio MCP startup.
- Engram version must match `pin.json` (`3.6.0`); newer upstream releases require a deliberate pin bump (archives, `binary_sha256`, and integration re-run).
- Rule units run under `go test ./...`. `cmd/plain-decide-report` is not a PR gate.
