#!/usr/bin/env bash
# Download and verify engram v3.6.0 linux_amd64 for CI (no network in tests).
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CI_DIR="${REPO_ROOT}/adapters/engram/.ci"
ARCHIVE="${CI_DIR}/engram_3.6.0_linux_amd64.tar.gz"
URL="https://github.com/Gentleman-Programming/engram/releases/download/v3.6.0/engram_3.6.0_linux_amd64.tar.gz"
SHA="945fb471188826c251160947f8dda9fe148ba0dd095fa3453174fe7ab5c6479e"

mkdir -p "${CI_DIR}"
if [[ ! -f "${ARCHIVE}" ]]; then
  curl -fsSL -o "${ARCHIVE}" "${URL}"
fi
echo "${SHA}  ${ARCHIVE}" | sha256sum -c
tar -xzf "${ARCHIVE}" -C "${CI_DIR}" engram
chmod +x "${CI_DIR}/engram"
export ENGRAM_BIN="${CI_DIR}/engram"
echo "ENGRAM_BIN=${ENGRAM_BIN}"
