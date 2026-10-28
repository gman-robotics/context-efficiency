---
name: engram-save-topic-key
description: >-
  Use retain (mem_save without topic_key) on the Engram adapter: titles, JSON body,
  size limits, and when topic_key is reserved for handoffs.
version: 1.0.0
---

# engram-save-topic-key

**Usage skill** for storing distilled facts via **`retain`**. The Engram child persists through `mem_save`; **`retain` never sets `topic_key`**. Handoffs use `topic_key` under `handoff_write` (see `engram-handoff`).

## When to load

Before calling `retain` after a decision, outcome, or short fact that should survive transcript resets.

## Tool

| Tool | Required args | Optional args | Returns |
| --- | --- | --- | --- |
| `retain` | `text`, `importance` (number) | `tags` (string array), `scope` | observation id (string) |

## `mem_save` mapping for `retain`

| Field | Value |
| --- | --- |
| `title` | `<scope-token>` + space + first **40 bytes** of trimmed `text` (Go `len` / byte clip, not Unicode runes) |
| `content` | JSON **`RetainBody`**: `text`, `tags`, `importance`, `scope` (domain scope string) |
| `type` | `manual` |
| `scope` (Engram field) | always **`project`** (not the domain scope name) |
| `capture_prompt` | `false` |
| **`topic_key`** | **omitted** — do not pass; upsert-by-topic is only for handoffs |

After save, the adapter re-reads the observation and errors if stored JSON differs from what was sent.

## Pre-checks (fail before `mem_save`)

Content must survive Engram storage rules **unchanged**:

1. **Private blocks:** Any **single-line** `<private>...</private>` (case-insensitive) in `text` → error (`content would be redacted by storage rules`). The check uses a pattern where `.` does **not** cross newlines, so multiline private tags are not detected here — still never store secrets in memory.
2. **50 000-byte cap:** After the adapter’s **40 000-byte** marshaled JSON guard (and retain **`text` > 8000** guard), prepared body/title must not require truncation or redaction at Engram’s **50 000**-byte stored limit, or the call errors (`content would be redacted or truncated by storage rules`).
3. **Adapter size guards (run before the 50 000-byte path):** `text` over **8000** bytes or marshaled retain JSON over **40000** bytes → error (`text too large` / `retain body too large`).

Keep `text` to **1–3 sentences** of distilled fact; put long material in files and retain a pointer line.

## Agent rules

- **`importance`:** Use **1–3** (plan convention: higher = more worth recalling).
- **`tags`:** Short stable labels (e.g. `decision`, `api`, routine name); optional but helps future `recall` queries.
- **`scope`:** Set explicitly when the fact belongs to `shared`, `routine:<name>`, or `task:<id>`; default is `bot`.
- Do **not** call `mem_save` directly from skills; use domain **`retain`** so titles, JSON shape, and checks stay consistent.
- Do **not** use `retain` for structured handoffs; use **`handoff_write`**, which sets `topic_key`.

## Related

- `docs/tool-mapping.md` — `retain` row
- `plugin/skills/engram-scope/SKILL.md` — scope tokens in titles and body
- `plugin/skills/engram-handoff/SKILL.md` — `topic_key` `handoff/<token>`
