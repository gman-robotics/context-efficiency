# Memory backends (setup overview)

Context-efficiency skills assume a **pluggable** memory MCP. This document describes how to wire each backend class in Grok Bot. It does **not** ship server code or a default `mcp.json` pin.

> **Default backend: pending (GMA-425).** Tom’s decision on which local OSS MCP (if any) to recommend is still open. Until GMA-425 closes, treat every local package below as a **candidate**, not a project default.

## Shared requirements

Whatever backend you use should support, at minimum:

- **Write** short text memories (for `retain`)
- **Search** by natural-language query (for `recall`)
- **Local or hosted** operation documented clearly (API keys, data residency)

Handoffs (`handoff_write` / `handoff_read`) may use memory rows **or** markdown files under `/workspace/handoffs/` plus a `retain` pointer — see `docs/tool-mapping.md`.

Skills stay backend-agnostic; only MCP configuration and mapping change.

---

## Local open-source (SQLite / FTS MCP)

**Role:** Zero API key; data stays on the Bot computer. Preferred **class** for the eventual default, pending GMA-425.

**Candidates (neutral — not ranked as default):**

| Candidate | Typical transport | Notes |
| --- | --- | --- |
| `simple-memory-mcp` | `npx` / `bunx` / compiled binary | SQLite + search; verify tool names against `docs/tool-mapping.md` |
| `mini-memory-mcp` | package-specific | Lightweight FTS variants exist in the ecosystem |
| Other SQLite MCPs | varies | Same bar: add + search (+ list optional) |

**Setup pattern (illustrative — replace command after GMA-425):**

1. Install the chosen MCP per its README (prefer **Bun** (`bunx`) or a **Rust/Go** binary when available; do not use Node.js for a custom TypeScript server in this project).
2. Add an entry to the plugin’s `mcp.json` (future) or the Bot’s MCP settings with the server `command` / `args` or `url`.
3. Confirm tools in the Bot UI; update the mapping table in `docs/tool-mapping.md` if names differ.
4. Smoke-test: one `retain`-equivalent add, one `recall`-equivalent search.

**Data:** SQLite file path is usually configured via env var or server default — document the path in your Bot notes so handoffs and memory are backed up together.

---

## Mem0 (hosted MCP)

**Role:** Hosted memory with MCP access; requires API key.

**Endpoint (from plan):** `https://mcp.mem0.ai/mcp`

**Setup pattern:**

1. Create a Mem0 account and API key per [Mem0 documentation](https://docs.mem0.ai/).
2. Configure MCP in Grok Bot with the hosted URL and authentication header Mem0 specifies.
3. Map tools using `docs/tool-mapping.md` (`add_memory` ↔ `retain`, `search_memories` ↔ `recall`).
4. Smoke-test search/add; keep stored items short to protect context window.

**Data:** Hosted; not on the Bot disk. Suitable when local SQLite is undesirable.

---

## Honcho

**Role:** Session-oriented memory; may already be available as a Grok Bot plugin.

**Setup pattern:**

1. Install or enable the Honcho plugin (or self-hosted Honcho MCP if you operate one).
2. List MCP tools in the Bot; align names with `docs/tool-mapping.md` Honcho section.
3. Define how `scope` maps to Honcho sessions, users, or peers.
4. Smoke-test recall/retain equivalents before enabling context-efficiency skills.

**Data:** Depends on Honcho deployment (plugin vs self-hosted).

---

## Choosing a backend (decision record)

| Decision | Status | Owner / tracker |
| --- | --- | --- |
| Default local MCP package | **Pending** | [GMA-425](https://linear.app/gman-personal-projects/issue/GMA-425) |
| Thin domain wrapper vs mapping-only | Open | `docs/plan.md` open questions |
| Mem0 / Honcho as documented alternates | Documented | This file + `tool-mapping.md` |

Do not merge plugin `mcp.json` with a pinned default until GMA-425 resolves.

---

## Security and ops

- Never commit API keys; use Bot secret / env configuration.
- Prefer least-privilege keys for hosted backends.
- For local SQLite, avoid sharing the DB file across untrusted Bots without scope tags.

## Related docs

- `docs/tool-mapping.md` — domain tool mapping per backend
- `docs/plan.md` — architecture and constraints
