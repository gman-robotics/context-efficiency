---
name: engram-scope
description: >-
  Domain scope strings and 40-character Engram tokens for recall, retain, and handoff
  on the context-efficiency Engram adapter; project vs bot scope and content limits.
version: 1.0.0
---

# engram-scope

**Usage skill** for the **`scope`** argument on `recall`, `retain`, `handoff_write`, and `handoff_read`. Domain scopes are **not** passed as Engram’s `project` field on child writes; they map to a fixed **40-character token** used in titles and search queries.

## When to load

Whenever you choose or omit `scope` on a domain memory tool, or when debugging “wrong scope” empty results.

## Supported domain scopes

| `scope` value | Meaning |
| --- | --- |
| **`bot`** | Default when the argument is omitted or blank. Bot-wide durable facts. |
| **`shared`** | Shared across bots/users in the same Engram project (still filtered by token). |
| **`routine:<name>`** | `<name>` must match **`^[a-z0-9]{1,40}$`** (lowercase alphanumeric). |
| **`task:<id>`** | `<id>` must match **`^[a-z0-9]{1,40}$`**. |

Any other string → tool error (`unsupported scope` or invalid routine/task segment).

## Scope token (40 characters)

The adapter derives a **40-character token** from the domain scope:

- One-byte **kind** + **payload**, then padded with **`-`** to length 40.
- Padding character **`-`** is forbidden inside routine/task name segments, so tokens stay unambiguous.
- Examples (payload only; full token is always length 40):
  - `bot` → kind `b`, payload `cesbot`
  - `shared` → kind `s`, payload `cesshared`
  - `routine:daily` → kind `r`, payload `cesroutine` + `daily`
  - `task:gma426` → kind `t`, payload `cestask` + `gma426`

Agents pass the **domain scope string** on tools; the adapter computes the token. Do not hand-craft tokens unless verifying adapter behavior.

## Where the token appears

| Operation | Token usage |
| --- | --- |
| **`recall`** | Child search query: `<token> <user query>`; results require title first field **==** `<token>`. |
| **`retain`** | `mem_save` title: `<token> <clip>`; `RetainBody.scope` stores the domain scope string. |
| **`handoff_write` / `handoff_read`** | `topic_key` = `handoff/<token>`; title remains literal **`handoff`**. |

## Engram `project` vs domain `scope`

| Layer | Behavior |
| --- | --- |
| **Operator env `ENGRAM_PROJECT`** | Names the Engram SQLite project (CLI `--project` on the child). Not the domain `scope` argument. |
| **Child `mem_save` `scope` field** | Always **`project`** for adapter writes. Domain isolation is via **token + JSON body**, not this field. |
| **Domain `scope` on tools** | Selects which token/filter bucket facts and handoffs use. |

Misunderstanding this layer causes “I used scope bot but searched the wrong project” — fix **`ENGRAM_PROJECT` / data dir** in MCP config, not the domain scope string alone.

## Private content and size (all scoped writes)

Before any retain or handoff write, the adapter applies Engram-aligned rules:

1. **`<private>...</private>`** in any stored string → **reject** (no silent redaction).
2. **50 000-byte** stored-content limit → **reject** if Engram would truncate or redact.
3. Retain additionally rejects oversized **`text`** or JSON as documented in `engram-save-topic-key`.

Scoped **routine** rows can be numerous; **`recall`** widens FTS to 20 hits then trims to `limit` so bot-scope facts are less likely to be crowded out — still prefer appropriate `scope` and concise retains.

## Agent rules

- Omit `scope` only when **`bot`** is intended.
- Use **`routine:<name>`** for recurring workflows and **`task:<id>`** for one-off workstreams.
- Use **`shared`** sparingly for facts that must outlive a single bot profile within the project.
- Pass the **same `scope`** to `handoff_write` and `handoff_read`.

## Related

- `docs/plan.md` — `Scope` type
- `docs/tool-mapping.md` — token filtering table
- `plugin/skills/engram-recall/SKILL.md` — search window and filtering
