---
name: routine-hygiene
description: >-
  Process skill for a routine. Call the plain decide tool and follow its
  label.
version: 1.0.0
---

# routine-hygiene

**Process skill** for a routine the bot creates or runs. The other four skills under `plugin/skills/` are adapter usage skills. This one is the routine process.

The gate is plain code. No model or server is needed.

## Step 1. Gate

Call `decide` with path `routine_go_nogo` and `features` `{name_ok, action, blocked}`.

- Continue only when `label` is `go`.
- Stop when `label` is `no_go`.
- Stop when the error starts with `refused:`.
- Every result has `effect` `none`.

Go only when the name is ok, the action is `write_file` or `retain_summary`, and blocked is false. Otherwise the label is `no_go`.

Tool retry uses `decide` with path `tool_failure_stop`. That gate is separate from this step. Stop there only when the same tool, the same error class, and the count is 3 or more.

## Step 2. Recall

Short `recall` for `routine:<name>`.

## Step 3. Work

Do the routine. Put long output in a file. The chat keeps the path and a one-line summary.

## Step 4. Retain

`retain` 1 to 3 sentences for the outcome. Do not retain the transcript.

## Related

- [Plain decide](../../../docs/plain-decide.md)
- [Plan](../../../docs/plan.md)
- [Tool mapping](../../../docs/tool-mapping.md)
- [Recall usage](../engram-recall/SKILL.md)
- [Scope usage](../engram-scope/SKILL.md)
