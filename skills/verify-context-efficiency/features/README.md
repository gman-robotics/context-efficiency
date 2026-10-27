# context-efficiency verification map

User-facing “features” for this repo are **documentation and agent workflows** until the Grok plugin ships. Read the index, then drive the matching feature file.

## Baseline preconditions

- Git checkout at repo root with `README.md` and `docs/plan.md`.
- `python3` on PATH (used by `scripts/check.sh`).
- Run `skills/verify-context-efficiency/bin/verify-context-efficiency-doctor` before the first drive.

## Driving conventions

- Run commands from repository root.
- Capture transcripts under `$VERIFY_CE_ARTIFACTS_DIR` for PR evidence.
- Do not report a feature verified if `scripts/check.sh` exited non-zero.

## Features

- [Repository docs](./repo-docs.md) — README and plan are present and linked.
- [Repo check script](./repo-check-script.md) — `scripts/check.sh` passes (CI parity).
