#!/usr/bin/env bash
# Launch context-efficiency-mcp (Go adapter). Does not download engram at startup.
set -euo pipefail

PLUGIN_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "${PLUGIN_ROOT}/.." && pwd)"
ADAPTER_DIR="${REPO_ROOT}/adapters/engram"
MCP_BIN="${CONTEXT_EFFICIENCY_MCP_BIN:-${ADAPTER_DIR}/context-efficiency-mcp}"

if [[ ! -x "${MCP_BIN}" ]]; then
  echo "context-efficiency-mcp: missing executable at ${MCP_BIN}" >&2
  echo "Build: (cd adapters/engram && go build -o context-efficiency-mcp ./cmd/context-efficiency-mcp/)" >&2
  echo "Install engram v3.6.0 beside the adapter or set ENGRAM_BIN (see docs/backends.md)." >&2
  exit 1
fi

if [[ -z "${ENGRAM_BIN:-}" ]]; then
  if [[ -x "${ADAPTER_DIR}/engram" ]]; then
    export ENGRAM_BIN="${ADAPTER_DIR}/engram"
  fi
fi

exec "${MCP_BIN}"
