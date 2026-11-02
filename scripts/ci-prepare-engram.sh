#!/usr/bin/env bash
# Download and verify engram v3.6.0 linux_amd64 for CI (no network in tests).
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PIN="${REPO_ROOT}/adapters/engram/pin.json"
CI_DIR="${REPO_ROOT}/adapters/engram/.ci"
ARCHIVE="${CI_DIR}/engram_3.6.0_linux_amd64.tar.gz"

read -r URL SHA <<<"$(python3 - "${PIN}" <<'PY'
import json, sys
pin = json.load(open(sys.argv[1], encoding="utf-8"))
arch = pin["archives"]["linux_amd64"]
print(arch["url"], arch["sha256"])
PY
)"

mkdir -p "${CI_DIR}"
if [[ ! -f "${ARCHIVE}" ]]; then
  curl -fsSL -o "${ARCHIVE}" "${URL}"
fi
echo "${SHA}  ${ARCHIVE}" | sha256sum -c
tar -xzf "${ARCHIVE}" -C "${CI_DIR}" engram
chmod +x "${CI_DIR}/engram"
export ENGRAM_BIN="${CI_DIR}/engram"
echo "ENGRAM_BIN=${ENGRAM_BIN}"
