#!/usr/bin/env bash
# GMA-543 local end-to-end: build the MCP server and fetch the pinned Engram
# inside `docker run --rm` (linux/amd64, because pin.json pins one linux_amd64
# binary digest), then drive recall, retain, handoff and decide over stdio
# through plugin/bin/run-context-efficiency-mcp.sh against a throwaway data
# dir. Nothing is installed on the host and no Grok or Cursor config is touched.
# The repo is mounted read-only; only reports/ is written back.
# No LAYA_API_KEY is passed in. Exit: 0 all pass, 2 a step failed, 1 error.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${CE_E2E_IMAGE:-golang:1.26}"
COMMIT="$(git -C "${REPO_ROOT}" rev-parse HEAD)"
mkdir -p "${REPO_ROOT}/reports"

docker run --rm --platform linux/amd64 \
  -v "${REPO_ROOT}:/src:ro" \
  -v "${REPO_ROOT}/reports:/out" \
  -e COMMIT="${COMMIT}" \
  "${IMAGE}" bash -euo pipefail -c '
    command -v python3 >/dev/null || { echo "e2e-local: image needs python3 for ci-prepare-engram.sh" >&2; exit 1; }
    export GOFLAGS=-buildvcs=false
    cp -a /src /tmp/ce && cd /tmp/ce
    rm -rf adapters/engram/.ci adapters/engram/context-efficiency-mcp
    source scripts/ci-prepare-engram.sh >/dev/null
    (cd adapters/engram && go build -o context-efficiency-mcp ./cmd/context-efficiency-mcp/)
    data="$(mktemp -d /tmp/ce-data.XXXXXX)"
    export ENGRAM_DATA_DIR="${data}" ENGRAM_PROJECT=ce-e2e
    export CONTEXT_EFFICIENCY_MCP_BIN=/tmp/ce/adapters/engram/context-efficiency-mcp
    status=0
    (cd adapters/engram && go run ./cmd/ce-e2e -launcher /tmp/ce/plugin/bin/run-context-efficiency-mcp.sh -out /out -commit "${COMMIT}") || status=$?
    rm -rf "${data}"
    exit "${status}"
  '
