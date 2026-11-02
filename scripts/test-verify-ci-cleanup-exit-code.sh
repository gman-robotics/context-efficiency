#!/usr/bin/env bash
# Assert verify-context-efficiency-ci cleanup trap preserves failure exit codes (not 139).
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERIFY="${REPO_ROOT}/skills/verify-context-efficiency/bin/verify-context-efficiency-ci"
artifacts="/tmp/verify-context-efficiency/cleanup-exit-test-$$"
mkdir -p "${artifacts}"

set +e
VERIFY_CE_TEST_FORCE_FAIL=1 VERIFY_CE_ARTIFACTS_DIR="${artifacts}" "${VERIFY}"
ec=$?
set -e

if [[ "${ec}" -eq 139 ]]; then
  echo "verify-ci cleanup exit-code test: FAIL (got SIGSEGV exit 139; trap re-entered on TERM)"
  exit 1
fi
if [[ "${ec}" -ne 7 ]]; then
  echo "verify-ci cleanup exit-code test: FAIL (expected exit 7, got ${ec})"
  exit 1
fi

echo "verify-ci cleanup exit-code test: OK (failure exit ${ec}, not 139)"
