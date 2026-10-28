# Tool mapping (domain → MCP backends)

Skills call a **stable domain surface** (`docs/plan.md`). Each backend exposes different MCP tool names. Adapters (this doc + user `mcp.json`, or a thin Rust/Go/Bun wrapper later) translate between them.

**Default backend:** **Engram v3.6.0** via Go adapter `context-efficiency-mcp` ([engram](https://github.com/Gentleman-Programming/engram) tag `v3.6.0`, MCP tools in `internal/mcp/mcp.go`). Child: `mem_save`, `mem_search`, `mem_get_observation` only.

**Server policy:** Rust, Go, or Bun TypeScript — **never Node.js**. Node SQLite MCPs below are comparison-only / **out of policy**.

---

## Engram v3.6.0 + Go adapter (default)

Source: [Gentleman-Programming/engram v3.6.0](https://github.com/Gentleman-Programming/engram/tree/v3.6.0). Domain scope omitted on tools defaults to **`bot`**. Scope maps to a fixed 40-character token (kind + payload, zero-padded); Engram `scope` on writes/searches is always `project`.

| Domain | Child tool | Mapping |
| --- | --- | --- |
| `recall` | `mem_search` | Query `<token> <user query>`, `match_mode` default `all`, `limit` 1–20; filter rows where title’s first field **equals** `<token>`; skip `handoff` titles; `mem_get_observation` per id; `unknown_project` → `[]` |
| `retain` | `mem_save` | Title `<token> <clip>`; body JSON `RetainBody`; `capture_prompt: false`; no `topic_key`; pre-check private tags and 50k truncation |
| `handoff_write` | `mem_save` | Title literal `handoff`; `topic_key` `handoff/<token>`; body JSON `Handoff` only; same pre-checks + post `mem_get_observation` compare |
| `handoff_read` | `mem_search` + `mem_get_observation` | Query exactly `handoff/<token>`; keep row where `topic_key` equals query; `unknown_project` → null |

`mem_session_summary`: **none** (unused). Tool-level `project`: **none** (unset; `--project` CLI only).

## Domain tools (what skills expect)

| Domain tool | Purpose | Arguments (conceptual) | Returns |
| --- | --- | --- | --- |
| `recall` | Search durable memory | `query`, optional `scope`, optional `limit` | `MemoryItem[]` |
| `retain` | Store a distilled fact | `text`, `tags`, `importance`, `scope` | `id` |
| `handoff_write` | Persist a structured handoff | `handoff`, `scope` | `pathOrId` |
| `handoff_read` | Load latest handoff for a scope | `scope` | `Handoff` or null |

Types: **Scope** `bot` \| `shared` \| `routine:<name>` \| `task:<id>`; **MemoryItem** and **Handoff** as in `docs/plan.md`.

---

## Named local packages (verified from upstream README)

### `simple-memory-mcp` (npm / [chrisribe/simple-memory-mcp](https://github.com/chrisribe/simple-memory-mcp)) — **Node — out of policy**

- **Runtime:** Node.js; start with `npx -y simple-memory-mcp` ([README](https://github.com/chrisribe/simple-memory-mcp#-quick-start)).
- **Not in policy** for context-efficiency defaults (Node server). Listed so docs do not confuse it with the Bun fork below.

**Published MCP tools** ([README § MCP Tools](https://github.com/chrisribe/simple-memory-mcp#%EF%B8%8F-mcp-tools)):

| Domain | Upstream tool | Notes |
| --- | --- | --- |
| `recall` | `memory-graphql` | GraphQL `Query.memories(query, tags, limit)` inside the tool |
| `retain` | `memory-graphql` | GraphQL `Mutation.store(content, tags)` |
| `handoff_write` | **none** | No handoff tool; use file under `/workspace/handoffs/` + `store` summary, or tags on `store` |
| `handoff_read` | **none** | No handoff tool; `memories`/`memory` queries by tag convention only |

Also: `export-memory`, `import-memory` (backup; not domain tools).

### `simple-memory-mcp` (Bun — [mrorigo/simple-memory-mcp](https://github.com/mrorigo/simple-memory-mcp)) — **Bun — in-policy candidate**

- **Runtime:** Bun; `bunx --bun github:mrorigo/simple-memory-mcp#main` ([README](https://github.com/mrorigo/simple-memory-mcp#run-from-github-with-bunx)).
- **Distinct repo** from the npm Node package above (same name, different implementation).

**Published MCP tools** ([README § Tools](https://github.com/mrorigo/simple-memory-mcp#tools)):

| Domain | Upstream tool | Notes |
| --- | --- | --- |
| `recall` | `search_memory` | `query`, optional `limit` |
| `retain` | `ingest_document` | Requires `filename` (path to UTF-8 file), not inline `text` — awkward for `retain`; wrapper or write-temp-file convention |
| `handoff_write` | **none** | File-first handoff + `ingest_document` if needed |
| `handoff_read` | **none** | No read-by-scope; `search_memory` only |

Also: `scrub_document` (delete from DB by filename; not domain).

### `mini-memory-mcp` ([utk09-ncl/mini-memory-mcp](https://github.com/utk09-ncl/mini-memory-mcp)) — **Node — out of policy**

- **Runtime:** Node.js ≥ 20; `npm start` / `npm run start:mcp` ([README](https://github.com/utk09-ncl/mini-memory-mcp#prerequisites)).
- **Not in policy** (Node server).

**Published MCP tools** ([README § Available MCP tools](https://github.com/utk09-ncl/mini-memory-mcp#available-mcp-tools)):

| Domain | Upstream tool | Notes |
| --- | --- | --- |
| `recall` | `memory_search` | Full-text search |
| `retain` | `memory_create` | `title`, `content`, optional `tags` — map `text` → `content`, encode `scope`/`importance` in tags |
| `handoff_write` | **none** | Use `memory_create` with handoff tags or file-first |
| `handoff_read` | **none** | `memory_read` needs `id`; no scope-level handoff type |

Also: `memory_list`, `memory_read`, `memory_update`, `memory_delete`.

### In-policy direction (GMA-425)

Prefer a **Rust**, **Go**, or **Bun TypeScript** MCP with add + search (and optional list). The Bun [mrorigo/simple-memory-mcp](https://github.com/mrorigo/simple-memory-mcp) fork is the only named candidate above that matches the runtime rule; its `ingest_document` shape may still require a thin adapter for `retain`.

---

## Mem0 (hosted MCP)

- **URL:** `https://mcp.mem0.ai/mcp` ([Mem0 MCP docs](https://docs.mem0.ai/platform/mem0-mcp)).
- **Auth:** Browser **OAuth** sign-in through the client, **or** API key as bearer token for headless/CI ([Signing in](https://docs.mem0.ai/platform/mem0-mcp#signing-in)).

**Published MCP tools** ([Available tools](https://docs.mem0.ai/platform/mem0-mcp#available-tools)):

| Domain | Mem0 tool | Notes |
| --- | --- | --- |
| `recall` | `search_memories` | Semantic search; scope maps to `user_id` / `agent_id` / `app_id` / `run_id` filters — not a generic `scope` string |
| `retain` | `add_memory` | Text/conversation; use `metadata` object for extra fields — **no** `tags` or `importance` parameter names on the tool |
| `handoff_write` | **none** | No handoff tool; convention: `add_memory` with structured JSON in text/metadata, or file-first |
| `handoff_read` | **none** | No handoff tool; `get_memory` / `search_memories` by metadata convention only |

Other tools (`get_memories`, `update_memory`, `delete_memory`, `list_entities`, …) are administrative; not domain aliases.

---

## Honcho (hosted MCP)

- **URL:** `https://mcp.honcho.dev` ([Grok Bot integration](https://honcho.dev/docs/v3/guides/integrations/grok-bot), [MCP guide](https://honcho.dev/docs/v3/guides/integrations/mcp)).
- **Auth:** OAuth via client connect flow, or API key bearer for custom connector ([Grok Bot doc](https://honcho.dev/docs/v3/guides/integrations/grok-bot#connecting-without-the-plugin)).

Tool list from [Honcho MCP server README](https://github.com/plastic-labs/honcho/tree/main/mcp) and [MCP instructions](https://raw.githubusercontent.com/plastic-labs/honcho/refs/heads/main/mcp/instructions.md). Reconcile with the live tool list in your Bot after connect.

| Domain | Honcho tool(s) | Notes |
| --- | --- | --- |
| `recall` | `search` | Semantic search; optional peer/session params. Also `chat`, `query_conclusions`, `get_representation` for read paths |
| `retain` | `add_messages_to_session` | Append user/assistant messages to a session; durable facts may use `create_conclusions` |
| `handoff_write` | **none** | No handoff tool; file-first under `/workspace/handoffs/` + session messages or conclusions |
| `handoff_read` | **none** | No handoff tool; `inspect_session` / `get_session_context` are session introspection, not `Handoff` shape |

Map plugin **scope** to Honcho workspace / peer / session naming (see Grok Bot doc table).

---

## File-first handoffs (all backends)

When `handoff_write` / `handoff_read` map to **none** on a backend:

1. Write markdown to `/workspace/handoffs/<scope-slug>.md`.
2. `retain` a one-line pointer (path + goal).
3. `handoff_read` → read that file (and optional `recall` for the pointer).

---

## Adapter strategies

| Strategy | When | Tradeoff |
| --- | --- | --- |
| Mapping doc + user config | Tool names close enough | No extra binary |
| Thin wrapper (Rust / Go / Bun TS) | Need domain tool names or inline `retain` | Maintained adapter |

GMA-425 selects the default backend; this mapping stays valid for alternates.

## Related docs

- `docs/backends.md` — setup per backend class
- `docs/plan.md` — architecture
