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
else
  echo "check-go: skip integration (ENGRAM_BIN unset)"
fi

if [[ "${GITHUB_ACTIONS:-}" == "true" && "${ran_integration}" -eq 0 ]]; then
  echo "check-go: FAIL integration required in CI (ENGRAM_BIN unset or not executable)" >&2
  exit 1
fi

echo "check-go: OK"
