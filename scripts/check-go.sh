#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${REPO_ROOT}"

if [[ ! -f adapters/engram/go.mod ]]; then
  echo "check-go: skip (no adapters/engram/go.mod)"
  exit 0
fi

echo "check-go: unit tests"
(cd adapters/engram && go test ./...)

ran_integration=0
if [[ -n "${ENGRAM_BIN:-}" && -x "${ENGRAM_BIN}" ]]; then
  echo "check-go: integration tests ENGRAM_BIN=${ENGRAM_BIN}"
  (cd adapters/engram && go test -tags=integration -count=1 .)
  ran_integration=1
  echo "check-go: MCP startup pin (binary_sha256 via context-efficiency-mcp)"
  (cd adapters/engram && go build -o context-efficiency-mcp ./cmd/context-efficiency-mcp/)
  mcp_data_dir="$(mktemp -d)"
  set +e
  timeout 5s env ENGRAM_DATA_DIR="${mcp_data_dir}" ENGRAM_PROJECT="pin-check" ENGRAM_BIN="${ENGRAM_BIN}" \
    "${REPO_ROOT}/adapters/engram/context-efficiency-mcp" </dev/null >/dev/null 2>&1
  mcp_ec=$?
  set -e
  rm -rf "${mcp_data_dir}"
  if [[ "${mcp_ec}" -eq 124 ]]; then
    : # timed out after successful pin check and server start
  elif [[ "${mcp_ec}" -ne 0 ]]; then
    echo "check-go: FAIL MCP startup pin check (context-efficiency-mcp exit ${mcp_ec})" >&2
    exit 1
  fi
else
  echo "check-go: skip integration (ENGRAM_BIN unset)"
fi

if [[ "${GITHUB_ACTIONS:-}" == "true" && "${ran_integration}" -eq 0 ]]; then
  echo "check-go: FAIL integration required in CI (ENGRAM_BIN unset or not executable)" >&2
  exit 1
fi

echo "check-go: OK"
