---
name: standards-context-efficiency
description: >-
  PR, issue, planning, and verification standards for gman-robotics/context-efficiency.
  Load at session start before planning or opening a PR.
version: 1.0.0
---

# standards-context-efficiency

Project-local standards for **context-efficiency** (Grok Bot context-efficiency plugin). Agents load this skill with `AGENTS.md`.

## Tier scale

| Tier | Meaning |
| --- | --- |
| **T0** | Small change, docs-only, or check-only (hub Markdown, typos, config tweak) |
| **T1** | Medium feature or fix in an existing module |
| **T2** | Larger feature, new module, or cross-cutting behavior change |
| **T3** | Infra, migration, or new service / major architectural move |

## PR format

Install `.github/PULL_REQUEST_TEMPLATE.md` from agent-bootstrap (`hub_pull_request_template.py copy`). Body order:

1. `Tier:`
2. `Models:`
3. `Linear:` — `https://linear.app/...` issue URL or `N/A`
4. `Summary:` — include `Refs GMA-###` when tied to Linear
5. `Gates:` — CI / verify outcomes at head SHA, or `N/A (docs-only)`
6. `## pstack (Poteto) principles cited`
7. `## Verification evidence` — for this repo record `./scripts/check.sh` and/or `verify-context-efficiency-ci` (exit 0)
8. `## Checklist` — includes `standards-context-efficiency` when that skill exists

Verify template drift:

```bash
python3 /path/to/agent-bootstrap/scripts/hub_pull_request_template.py check --file .github/PULL_REQUEST_TEMPLATE.md
```

## Issue format

- **Problem** — what is wrong or missing
- **Acceptance criteria** — testable bullets
- **Non-goals** — explicit out-of-scope
- **Verification plan** — `./scripts/check.sh` and/or `verify-context-efficiency-ci` until product skills exist

## Planning

Read `docs/plan.md` fully before implementing plugin features. Do not ship a custom memory server unless the domain interface cannot map to existing OSS MCPs. Sequence work in verifiable units (docs → skills → manifest → backends).

## Implementation

- Skills: Markdown only under `skills/`.
- Custom MCP (if ever needed): Rust, Go, or Bun TypeScript — not Node.js.
- Prefer composing existing SQLite MCPs per `docs/plan.md` over new servers.
- Commit messages: imperative summary line; reference Linear id in body when applicable.

## Verification

**Local pre-push (CI parity)**

```bash
set -euo pipefail
./scripts/check.sh 2>&1 | tee /tmp/context-efficiency-check.log
```

**Verify skill:** `skills/verify-context-efficiency/SKILL.md` — doctor + drive via `skills/verify-context-efficiency/bin/verify-context-efficiency-ci`.

**CI pass bar:** workflow `.github/workflows/check.yml`, job `markdown-and-links`, step `Run repo check` — `./scripts/check.sh` must exit `0`.

**Gates:** ub-mode crapper/dryer/mutator are not required for this Markdown-first repo; record `N/A` or CI check results on the PR `Gates:` line.

## Maintenance

When CI, templates, or verify steps change, update this skill in the same PR. Hub drift skills (`maintain-repo-standards`, `maintain-verification-skill`) apply when bootstrapped from agent-bootstrap.
