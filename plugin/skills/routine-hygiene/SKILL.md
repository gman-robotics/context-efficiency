---
name: routine-hygiene
description: >-
  Process skill for a routine. Call the plain decide tool and follow its
  label. Laya shadow is optional and never blocks the routine.
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

### Optional Laya shadow

Only when a Laya server is set up outside the plugin: also call `decide_shadow` with the same path and `actual` set to the `decide` label. A Laya miss does not change the routine. Follow `actual`.

- Stop when the error starts with `refused:`.
- On `laya api key not set`, `laya transport:`, `laya http`, `laya usage truncated`, `laya verdict choice missing`, `laya verdict choice rejected`, `laya response unreadable`, or `laya log:`, continue from `actual`. `laya log:` means the retain failed. It is not a refusal.
- Every successful result has `effect` `none`.

## Step 2. Recall

Short `recall` for `routine:<name>`.

Read shadow rows with `recall` and scope `routine:layashadow`. The public cap stays 20. That list is not the pass bit. There is no `shadow_report` tool.

## Step 3. Work

Do the routine. Put long output in a file. The chat keeps the path and a one-line summary.

## Step 4. Retain

`retain` 1 to 3 sentences for the outcome. Do not retain the transcript.

## Related

- [Plain decide](../../../docs/plain-decide.md)
- [Laya shadow contract](../../../docs/laya-shadow.md)
- [Plan](../../../docs/plan.md)
- [Tool mapping](../../../docs/tool-mapping.md)
- [Recall usage](../engram-recall/SKILL.md)
- [Scope usage](../engram-scope/SKILL.md)
