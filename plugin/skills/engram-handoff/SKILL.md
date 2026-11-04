---
name: engram-handoff
description: >-
  Use handoff_write and handoff_read on the Engram adapter: literal handoff title,
  topic_key upsert, search/read path, and shared content rules.
version: 1.0.0
---

# engram-handoff

**Usage skill** for **`handoff_write`** and **`handoff_read`**. Handoffs are stored as a single upserted observation per scope via Engram **`topic_key`**, not via `recall`.

## When to load

When persisting or resuming a structured session handoff (goal, open items, decisions, paths, next steps) for a given scope.

## Tools

| Tool | Required args | Optional args | Returns |
| --- | --- | --- | --- |
| `handoff_write` | `handoff` (object) | `scope` | observation id (string) |
| `handoff_read` | — | `scope` | JSON `Handoff` or the text `null` |

**`Handoff` shape** (all string fields; arrays of strings): `goal`, `openItems`, `decisions`, `filePaths`, `nextSteps` — see `docs/plan.md`.

## Title rules (critical)

| Rule | Detail |
| --- | --- |
| **`mem_save` `title`** | Literal **`handoff`** (lowercase). Not the scope token, not a free-form title. |
| **Scope in title** | Scope is **not** in the title. **`Handoff` JSON has no scope field** — scope is only on the tool argument and in **`topic_key`** (`handoff/<40-char token>`). |
| **`recall` visibility** | Rows whose title starts with **`handoff`** are **excluded** from `recall` results. Load handoffs only with **`handoff_read`**. |

The adapter tests enforce the literal title; agents must not invent alternate titles for handoffs.

## `topic_key` and upsert

| Field | Value |
| --- | --- |
| **`topic_key`** | `handoff/` + **full 40-character scope token** (no extra slashes) |
| **`content`** | JSON **`Handoff` only** (not `RetainBody`) |
| Engram `scope` on save | **`project`** |
| `capture_prompt` | `false` |

Writing again with the same scope **replaces** the prior handoff for that `topic_key`.

## `handoff_read` path

1. `mem_search` with query **exactly** `handoff/<scope-token>` and limit 20.
2. Select the row whose **`topic_key`** equals that query string.
3. `mem_get_observation` on that id; unmarshal JSON to **`Handoff`**.
4. **`unknown_project`** → **`null`** (no handoff), not an error.
5. No matching topic → **`null`**.

Use the **same `scope` argument** as `handoff_write` when reading back.

## Pre-checks (same family as `retain`)

Content checks run in this order. A body that fails the 40 000-byte guard never reaches truncation.

1. **Private blocks:** Any **single-line** `<private>...</private>` (case-insensitive; pattern `(?i)<private>.*?</private>` with no `(?s)`, so `.` does not cross newlines) in any handoff string field → error (`content would be redacted by storage rules`).
2. **Marshaled JSON over 40 000 bytes** → error (`handoff body too large`).
3. **`PrepareForStore` at 50 000 bytes:** content that would be redacted or truncated → error (`content would be redacted or truncated by storage rules`).

## Agent rules

- Prefer **file paths** in `filePaths` over pasting large content into handoff strings.
- After `handoff_write`, use **`handoff_read`** (or a markdown file you also wrote) in a fresh chat — not `recall`.
- Default **`scope`** when omitted is **`bot`**; use `shared` / `routine:<name>` / `task:<id>` when the handoff belongs elsewhere.

## Related

- `docs/tool-mapping.md` — handoff rows
- `plugin/skills/engram-scope/SKILL.md` — token and `handoff/<token>` format
- `plugin/skills/engram-save-topic-key/SKILL.md` — retain without `topic_key`
