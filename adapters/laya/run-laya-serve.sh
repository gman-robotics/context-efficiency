#!/usr/bin/env bash
# Start loopback laya-serve. Does not install packages and does not print the API key.
set -euo pipefail

LAYA_VENV="${LAYA_VENV:-${HOME}/.venvs/laya}"
PY="${LAYA_VENV}/bin/python"

if [[ ! -x "${PY}" ]]; then
  echo "set LAYA_VENV to a venv whose python is 3.10 or newer" >&2
  exit 1
fi

if ! "${PY}" -c 'import sys; raise SystemExit(0 if sys.version_info >= (3, 10) else 1)'; then
  echo "Python >= 3.10 required" >&2
  exit 1
fi

if [[ -z "${LAYA_API_KEY:-}" ]]; then
  if ! command -v security >/dev/null 2>&1; then
    echo "LAYA_API_KEY is unset and the keychain lookup failed" >&2
    exit 1
  fi
  key_status=0
  LAYA_API_KEY="$(security find-generic-password -s gman-laya-api-key -w 2>/dev/null)" || key_status=$?
  if [[ "${key_status}" -ne 0 || -z "${LAYA_API_KEY}" ]]; then
    echo "LAYA_API_KEY is unset and the keychain lookup failed" >&2
    exit 1
  fi
  export LAYA_API_KEY
fi

if [[ ! -x "${LAYA_VENV}/bin/laya-serve" ]]; then
  echo "venv has no laya-serve" >&2
  exit 1
fi

export LAYA_HOST=127.0.0.1
export LAYA_PORT="${LAYA_PORT:-8792}"
export HF_HUB_OFFLINE=1
export HF_HUB_DISABLE_TELEMETRY=1
export LAYA_REVISION=reviewed
export LAYA_MODELS=english
export LAYA_DEFAULT_MODEL=english
export LAYA_PRELOAD=1
export LAYA_SHA256_DIGESTS='{"model.safetensors":"891102d372688fc2a094dac56a384bc537b87c63f21f9f3dac0be2b7cbc8d86c"}'

exec "${LAYA_VENV}/bin/laya-serve"
