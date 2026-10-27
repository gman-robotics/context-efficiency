# Grok project notes

Grok discovers project rules from `AGENTS.md` at the repo root and skills under `.grok/skills/` when exported.

Run `bash /path/to/agent-bootstrap/scripts/install-grok.sh` from a sibling agent-bootstrap checkout to sync hub skills into `.grok/skills/`.

<!-- agent-bootstrap-hub:BEGIN harness-workflow -->
## Agent harness workflow (agent-bootstrap)

Do **not** skip or override this block when other agent files exist.

1. **Load `skills/standards-context-efficiency/SKILL.md`** from harness skill folders this repo already uses (for example `.cursor/skills/`, `.claude/skills/`, `.grok/skills/`). Read it at session start for PR, issue, planning, and verification formats.
2. **Before opening a PR:** run the project verify skill (`verify-*` / `bin/verify-*-ci`) and the **pre-PR ub checks** documented in `standards-context-efficiency` and `docs/shared/ub-mode-practice.md` when bootstrapped from agent-bootstrap (crapper, dryer, mutator on merge-base scope; record results honestly).
3. **PR body:** use `.github/pull_request_template.md`.
4. **PR review:** for non-trivial diffs, apply `skills/blast-radius/SKILL.md` to the changed surface before APPROVE.

Maintained by `scripts/sync_agent_harness_instructions.py` (`create-verification-skill` / `maintain-verification-skill`).
<!-- agent-bootstrap-hub:END harness-workflow -->
