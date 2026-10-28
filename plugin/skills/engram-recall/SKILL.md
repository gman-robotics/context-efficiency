---
name: engram-recall
description: >-
  Use the context-efficiency recall MCP tool against the Engram adapter: limits,
  FTS search window, scope token prefix, and handoff exclusion.
version: 1.0.0
---

# engram-recall

**Usage skill** for the domain tool `recall` on `context-efficiency-mcp` (Engram v3.6.0 child). Process skills (hygiene, handoff advisor, etc.) live elsewhere; this file only covers correct tool calls.

## When to load

Before calling `recall` to pull durable facts into working context. Pair with `engram-scope` when choosing or omitting `scope`.

## Tool

| Tool | Required args | Optional args | Returns |
| --- | --- | --- | --- |
| `recall` | `query` (string) | `scope`, `limit` (number) | JSON array of `MemoryItem` |

`MemoryItem` fields: `id`, `text`, `tags`, `importance`, `scope`, `createdAt` (see `docs/plan.md`).

## Limit and search window

The adapter enforces limits before returning rows to the agent:

1. **`limit` default:** If omitted or below 1, the adapter uses **5**.
2. **`limit` ceiling:** Values above **20** are clamped to **20**.
3. **FTS window:** The child `mem_search` always requests at least **20** hits (`max(limit, 20)`), then the adapter filters and returns **at most `limit`** items.

Use a small `limit` (3–5) in chat to keep context lean. Use a higher limit (up to 20) only when you need broader coverage and accept the token cost.

## What the adapter does (do not reimplement)

- Builds the child query as **`<scope-token> ` + trimmed user `query`**. The token comes from `scope` (default `bot`); see `engram-scope`.
- Calls `mem_search` with that query and the widened limit above.
- Keeps rows whose **title’s first whitespace-separated field equals the scope token**.
- **Skips** rows whose title starts with **`handoff`** (handoffs are loaded via `handoff_read`, not `recall`).
- Fetches each kept id with `mem_get_observation` and unmarshals the stored `RetainBody` JSON into `MemoryItem`.
- **`unknown_project`** from the child → empty array `[]`, not an error.

## Agent rules

- Pass a **short, task-specific `query`** derived from the user message; do not paste long logs into `query`.
- Set **`scope`** explicitly when facts may live under `shared`, `routine:<name>`, or `task:<id>`; omitting `scope` is **`bot`** only.
- Do not use `recall` to load handoffs; use `handoff_read` with the same scope.
- Treat an empty array as “nothing durable matched,” not as MCP failure.

## Related

- `docs/tool-mapping.md` — Engram row for `recall`
- `plugin/skills/engram-scope/SKILL.md` — scope tokens
- `plugin/skills/engram-handoff/SKILL.md` — why handoff rows are excluded
