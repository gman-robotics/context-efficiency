# Context Efficiency — agent instructions

**Purpose:** Grok Bot plugin and skills that keep live transcripts short by moving durable state into external memory and files. See `README.md` and `docs/plan.md` before implementing features.

**Stack preferences**

- Skills: Markdown (`SKILL.md`) only.
- Optional custom MCP server: **Rust**, **Go**, or **Bun TypeScript** — never Node.js for TypeScript servers.
- Do not pick or ship a memory backend in bootstrap-only work; compose OSS adapters per `docs/plan.md`.

**Verify before a PR**

```bash
./scripts/check.sh
```

Or the full verify bundle:

```bash
skills/verify-context-efficiency/bin/verify-context-efficiency-ci
```

**PR body header order (required)**

1. `Tier:`
2. `Models:`
3. `Linear:`
4. `Summary:` (include `Refs GMA-###` when tied to Linear)
5. `Gates:`

Use `.github/pull_request_template.md`. Deeper PR/issue/planning rules: `skills/standards-context-efficiency/SKILL.md`.

<!-- agent-bootstrap-hub:BEGIN harness-workflow -->
## Agent harness workflow (agent-bootstrap)

Do **not** skip or override this block when other agent files exist.

1. **Load `skills/standards-context-efficiency/SKILL.md`** from harness skill folders this repo already uses (for example `.cursor/skills/`, `.claude/skills/`, `.grok/skills/`). Read it at session start for PR, issue, planning, and verification formats.
2. **Before opening a PR:** run the project verify skill (`verify-*` / `bin/verify-*-ci`) and the **pre-PR ub checks** documented in `standards-context-efficiency` and `docs/shared/ub-mode-practice.md` when bootstrapped from agent-bootstrap (crapper, dryer, mutator on merge-base scope; record results honestly).
3. **PR body:** use `.github/pull_request_template.md` (header order in item 5 above).
4. **PR review:** for non-trivial diffs, apply `skills/blast-radius/SKILL.md` to the changed surface before APPROVE.

Maintained by `scripts/sync_agent_harness_instructions.py` (`create-verification-skill` / `maintain-verification-skill`).
<!-- agent-bootstrap-hub:END harness-workflow -->
