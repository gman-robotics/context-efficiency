# Memory backends (setup overview)

Pluggable memory MCP for context-efficiency skills. This doc does **not** ship `mcp.json` or pick a default package.

> **Default backend: pending ([GMA-425](https://linear.app/gman-personal-projects/issue/GMA-425)).** Until that issue closes, no recommended pin — especially not Node-based servers (see below).

## Shared requirements

- **Write** short memories (`retain`)
- **Search** by query (`recall`)
- Clear **auth** and **data residency**
- **Server runtime:** Rust, Go, or Bun TypeScript only for servers this project adopts — **never Node.js**

Handoffs may be file-first; see `docs/tool-mapping.md`.

---

## Local open-source (SQLite / FTS)

**Role:** Local data, no hosted API key. Preferred **class** after GMA-425, subject to runtime policy.

### Named packages (verified; not ranked)

| Package | Upstream | Runtime | Policy |
| --- | --- | --- | --- |
| `simple-memory-mcp` (npm) | [chrisribe/simple-memory-mcp](https://github.com/chrisribe/simple-memory-mcp) | **Node** (`npx -y simple-memory-mcp`) | **Out of policy** |
| `simple-memory-mcp` (Bun fork) | [mrorigo/simple-memory-mcp](https://github.com/mrorigo/simple-memory-mcp) | **Bun** (`bunx --bun github:mrorigo/simple-memory-mcp#main`) | **In-policy candidate** |
| `mini-memory-mcp` | [utk09-ncl/mini-memory-mcp](https://github.com/utk09-ncl/mini-memory-mcp) | **Node** (≥ 20; `npm run start:mcp`) | **Out of policy** |

Do not treat npm `simple-memory-mcp` and [mrorigo/simple-memory-mcp](https://github.com/mrorigo/simple-memory-mcp) as the same server — different runtimes, different tools ([tool-mapping.md](./tool-mapping.md)).

**After GMA-425 selects a backend:**

1. Install per upstream README (Bun/Rust/Go only for in-policy choices).
2. Add MCP entry in plugin `mcp.json` or Bot settings.
3. Confirm tools in the Bot UI match `docs/tool-mapping.md`.
4. Smoke-test one retain-equivalent write and one recall-equivalent search.

---

## Mem0 (hosted MCP)

**URL:** `https://mcp.mem0.ai/mcp` ([Mem0 MCP](https://docs.mem0.ai/platform/mem0-mcp))

**Auth (either):**

- **OAuth** — client opens browser sign-in on first tool use ([Signing in](https://docs.mem0.ai/platform/mem0-mcp#signing-in))
- **API key** — bearer token for headless clients (`Authorization: Token …` per Mem0 dashboard)

**Tools:** `add_memory`, `search_memories`, plus list/get/update/delete helpers ([tool table](https://docs.mem0.ai/platform/mem0-mcp#available-tools)).

**Scope:** Map domain scope to Mem0 `user_id` / `agent_id` / `app_id` / `run_id` — not a single `scope` enum.

**Data:** Hosted in Mem0 account.

---

## Honcho

**URL:** `https://mcp.honcho.dev` ([MCP integration](https://honcho.dev/docs/v3/guides/integrations/mcp), [Grok Bot](https://honcho.dev/docs/v3/guides/integrations/grok-bot))

**Auth:** OAuth through the Grok plugin connect flow, or API key on a custom MCP connector.

**Tools:** Workspace/session/peer tools such as `search`, `add_messages_to_session`, `chat` ([MCP tree](https://github.com/plastic-labs/honcho/tree/main/mcp)). Map domain operations per `docs/tool-mapping.md`; no dedicated handoff tools.

**Data:** Honcho cloud (or self-hosted if you operate Honcho yourself).

---

## Decision record

| Decision | Status |
| --- | --- |
| Default local MCP package | **Pending** — GMA-425 |
| Node-based `simple-memory-mcp` / `mini-memory-mcp` as project default | **Rejected** by runtime policy |
| Mem0 / Honcho as documented alternates | Documented |

Do not merge a pinned default `mcp.json` until GMA-425 resolves.

---

## Security

- No API keys in git; use Bot secrets / env.
- Least-privilege Mem0/Honcho keys.
- Local SQLite: scope tags if multiple Bots share a DB file.

## Related docs

- `docs/tool-mapping.md` — per-tool mapping with sources
- `docs/plan.md` — architecture
