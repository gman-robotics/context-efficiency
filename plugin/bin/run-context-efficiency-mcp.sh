#!/usr/bin/env bash
# Launch context-efficiency-mcp (Go adapter). Does not download engram at startup.
# A missing Laya key does not block startup. The keychain read is bounded and never echoed.
set -euo pipefail

PLUGIN_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "${PLUGIN_ROOT}/.." && pwd)"
ADAPTER_DIR="${REPO_ROOT}/adapters/engram"
MCP_BIN="${CONTEXT_EFFICIENCY_MCP_BIN:-${ADAPTER_DIR}/context-efficiency-mcp}"

if [[ ! -x "${MCP_BIN}" ]]; then
  echo "context-efficiency-mcp: missing executable at ${MCP_BIN}" >&2
  echo "Build: (cd adapters/engram && go build -o context-efficiency-mcp ./cmd/context-efficiency-mcp/)" >&2
  exit 1
fi

if [[ -z "${ENGRAM_BIN:-}" ]]; then
  if [[ -x "${ADAPTER_DIR}/engram" ]]; then
    export ENGRAM_BIN="${ADAPTER_DIR}/engram"
  fi
fi

load_laya_key() {
  if [[ -n "${LAYA_API_KEY:-}" ]]; then
    return 0
  fi
  if ! command -v security >/dev/null 2>&1; then
    return 0
  fi
  local tmp pid i status value
  tmp="$(mktemp)"
  chmod 600 "${tmp}"
  security find-generic-password -s gman-laya-api-key -w >"${tmp}" 2>/dev/null &
  pid=$!
  i=0
  while [[ "${i}" -lt 20 ]]; do
    if ! kill -0 "${pid}" 2>/dev/null; then
      break
    fi
    sleep 0.1
    i=$((i + 1))
  done
  if kill -0 "${pid}" 2>/dev/null; then
    kill "${pid}" 2>/dev/null || true
    wait "${pid}" 2>/dev/null || true
    rm -f "${tmp}"
    return 0
  fi
  status=0
  wait "${pid}" 2>/dev/null || status=$?
  if [[ "${status}" -eq 0 ]]; then
    value="$(tr -d '\n\r' <"${tmp}" 2>/dev/null || true)"
    if [[ -n "${value}" ]]; then
      LAYA_API_KEY="${value}"
      export LAYA_API_KEY
    fi
  fi
  rm -f "${tmp}"
  return 0
}

load_laya_key

exec "${MCP_BIN}"
