# Tool mapping (domain → MCP backends)

Skills in this plugin call a **small domain surface**. Each memory backend exposes different MCP tool names and argument shapes. Adapters (documentation + user `mcp.json` configuration, or a thin wrapper server later) translate between them.

**Default backend choice:** **Pending** — tracked in [GMA-425](https://linear.app/gman-personal-projects/issue/GMA-425). Do not assume a pinned local MCP package until that issue closes.

## Domain tools (what skills expect)

These names are the stable contract from `docs/plan.md`:

| Domain tool | Purpose | Arguments (conceptual) | Returns |
| --- | --- | --- | --- |
| `recall` | Search durable memory | `query`, optional `scope`, optional `limit` | `MemoryItem[]` |
| `retain` | Store a distilled fact | `text`, `tags`, `importance`, `scope` | `id` |
| `handoff_write` | Persist a structured handoff | `handoff`, `scope` | `pathOrId` |
| `handoff_read` | Load latest handoff for a scope | `scope` | `Handoff` or null |

Types:

- **Scope:** `bot` \| `shared` \| `routine:<name>` \| `task:<id>`
- **MemoryItem:** `id`, short `text`, `tags`, `importance` (1–3), `scope`, `createdAt`
- **Handoff:** `goal`, `openItems`, `decisions`, `filePaths`, `nextSteps`

Skills must not call backend-specific tool names directly. Only the MCP layer (or mapping doc the user follows) changes per backend.

## Mapping by backend class

### Local open-source SQLite MCP (candidate — not selected)

Typical OSS servers expose **add**, **search**, and **list** style tools rather than `recall` / `retain`. Handoffs are often **not** first-class; store JSON in a tagged memory row or a file path reference.

| Domain | Typical OSS MCP pattern | Notes |
| --- | --- | --- |
| `recall` | `search_memories` / `search` / `query` with text + limit | Map `scope` to tags or metadata filters if supported; otherwise filter in skill policy after search |
| `retain` | `add_memory` / `add` / `store` with text + metadata | Map `tags` + `scope` to server metadata; map `importance` to priority field or tag |
| `handoff_write` | `add_memory` with tag `handoff` + JSON body, or write file under `/workspace/handoffs/` and `retain` the path | Prefer file + `retain` when server has no structured handoff type |
| `handoff_read` | `search` for tag `handoff` + scope, or read known file path from last `retain` | May require convention: one handoff id per scope |

**Candidates (neutral list, no default):** community SQLite + FTS MCPs often named along the lines of `simple-memory-mcp`, `mini-memory-mcp`, or similar; evaluate against GMA-425 criteria (zero API key, local data, search + add, Bun/Rust/Go packaging).

### Mem0 (hosted MCP)

Documented endpoint: `https://mcp.mem0.ai/mcp` (API key in headers per Mem0 docs).

| Domain | Mem0-oriented MCP tool | Notes |
| --- | --- | --- |
| `recall` | `search_memories` (or equivalent search tool) | Pass `query`; apply `limit`; map `scope` to user/session/metadata filters if exposed |
| `retain` | `add_memory` | Map `text`, `tags`, `importance`, `scope` into Mem0 metadata fields |
| `handoff_write` | `add_memory` with structured JSON + type tag, or external file + `add_memory` summary | Same file-first fallback as local |
| `handoff_read` | `search_memories` filtered by handoff tag / metadata | Verify exact tool names in Mem0 MCP schema at install time |

### Honcho

Use the Honcho Grok Bot plugin or self-hosted Honcho MCP (see `docs/backends.md`). Tool names vary by deployment; treat the table below as **intent**, then reconcile with the live tool list in the Bot’s MCP panel.

| Domain | Honcho-oriented pattern | Notes |
| --- | --- | --- |
| `recall` | Session/user memory search APIs exposed as MCP tools | Map `scope` to Honcho session or peer identifiers |
| `retain` | Add or append memory on a session/user | Keep items short; align with Honcho’s message/memory model |
| `handoff_write` | Persist handoff as memory blob or linked resource | Prefer file path + short `retain` if structured handoff is awkward |
| `handoff_read` | Fetch by scope/session | Same conventions as local for tag `handoff` |

### Other MCP servers

Any MCP that offers **search + add** (and optionally list) can back `recall` / `retain` with the same adapter ideas. Handoffs may always be **file-first** (`handoff_write` → markdown under `/workspace/handoffs/` + `retain` pointer) if the server has no handoff type.

## Adapter strategies (no custom server vs thin wrapper)

| Strategy | When to use | Tradeoff |
| --- | --- | --- |
| **Mapping doc only** | Backend tools are close; skills use documented aliases in `mcp.json` tool exposure | User/Bot must configure correct server; skill text references logical names |
| **Thin wrapper MCP** (Rust / Go / Bun TS) | Need exact domain tool names or unified `scope` handling | Small binary; still compose OSS storage where possible |

Whether this repo needs a wrapper is **open** (see `docs/plan.md` open questions). GMA-425 decides the default backend; mapping stays valid regardless.

## Verification hooks

When a backend is selected (GMA-425):

1. Call `recall` with a known seeded string; assert ≤ `limit` items and short `text`.
2. Call `retain` then `recall`; assert round-trip.
3. `handoff_write` / `handoff_read` round-trip for one scope (or file path + `retain` convention).

Until then, `./scripts/check.sh` validates these docs and repo layout only — not live MCP calls.

## Related docs

- `docs/backends.md` — setup paths per backend class (no default pin)
- `docs/plan.md` — domain model and plugin structure
