---
name: verify-context-efficiency
description: >-
  Verify context-efficiency repo docs, skills layout, and scripts/check.sh the way
  an agent does before a PR. Use for bootstrap CI parity and T0/T1 doc changes.
version: 1.0.0
---

# verify-context-efficiency

Verification skill for **gman-robotics/context-efficiency**. There is no long-lived app server: the surface is **Markdown skills**, **docs**, and **`scripts/check.sh`**.

## Launch

**Prepare (once per machine / CI job):**

```bash
cd "$(git rev-parse --show-toplevel)"
```

**Ready when:** `test -f README.md` and `test -f docs/plan.md`.

No daemon. Each drive step is a fresh subprocess.

```bash
export VERIFY_CE_RUN_ID="${VERIFY_CE_RUN_ID:-$(date -u +%Y%m%dT%H%M%SZ)-$$}"
export VERIFY_CE_ARTIFACTS_DIR="${VERIFY_CE_ARTIFACTS_DIR:-/tmp/verify-context-efficiency/${VERIFY_CE_RUN_ID}}"
mkdir -p "$VERIFY_CE_ARTIFACTS_DIR"
```

## Doctor

```bash
skills/verify-context-efficiency/bin/verify-context-efficiency-doctor
```

Requires exit code `0` and a line `verify-context-efficiency doctor: OK`.

## Drive

| Goal | Command |
| --- | --- |
| Repo check (CI parity) | `./scripts/check.sh` |
| Full bundle | `skills/verify-context-efficiency/bin/verify-context-efficiency-ci` |

Read `features/README.md` before driving; cover the mapped entry points relevant to your change.

**Example (bootstrap docs):**

```bash
export VERIFY_CE_RUN_ID="manual-$(date -u +%Y%m%dT%H%M%SZ)"
export VERIFY_CE_ARTIFACTS_DIR="/tmp/verify-context-efficiency/${VERIFY_CE_RUN_ID}"
mkdir -p "$VERIFY_CE_ARTIFACTS_DIR"
skills/verify-context-efficiency/bin/verify-context-efficiency-doctor
./scripts/check.sh 2>&1 | tee "$VERIFY_CE_ARTIFACTS_DIR/check.transcript"
```

## Evidence

- Run real commands from repo root (not mocked outcomes).
- Capture command, exit code, and stdout/stderr in `$VERIFY_CE_ARTIFACTS_DIR/*.transcript`.
- `./scripts/check.sh` must end with `check: all gates passed` and exit `0`.
- Artifacts survive cleanup.

## Cleanup

Remove only scratch dirs you created under `/tmp/verify-context-efficiency/` for this run. Do not delete proof transcripts you intend to cite in a PR.

## CI bundle

```bash
skills/verify-context-efficiency/bin/verify-context-efficiency-ci
```

Runs doctor + `./scripts/check.sh`. Writes `N/A (no ub-mode gates; Markdown repo)` to `.metrics/verify-ub-outcome.txt` when `.metrics/` exists. Always exits `0` when checks pass.

## Helpers

| Script | Role |
| --- | --- |
| `bin/verify-context-efficiency-doctor` | Read-only preflight |
| `bin/verify-context-efficiency-ci` | Doctor + check.sh for PR/CI |
