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

if [[ -n "${ENGRAM_BIN:-}" && -x "${ENGRAM_BIN}" ]]; then
  echo "check-go: integration tests ENGRAM_BIN=${ENGRAM_BIN}"
  (cd adapters/engram && go test -tags=integration -count=1 .)
else
  echo "check-go: skip integration (ENGRAM_BIN unset)"
fi

echo "check-go: OK"
