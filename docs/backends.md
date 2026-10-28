# Memory backends (setup overview)

Pluggable memory MCP for context-efficiency skills.

## Default backend (decision record)

**Engram v3.6.0** is the default memory backend for this plugin, exposed through the in-repo **Go adapter** (`adapters/engram/`, binary `context-efficiency-mcp`). Upstream: [Gentleman-Programming/engram](https://github.com/Gentleman-Programming/engram) tag `v3.6.0`.

| Item | Value |
| --- | --- |
| Adapter module | `github.com/gman-robotics/context-efficiency/adapters/engram` |
| Domain MCP tools | `recall`, `retain`, `handoff_write`, `handoff_read` |
| Child process | `engram mcp --project "<name>" --tools=mem_save,mem_search,mem_get_observation` |
| Pin | `adapters/engram/pin.json` (archive URLs + sha256; `binary_sha256` on release artifacts) |
| Manifest / `mcp.json` | GMA-428 (not in GMA-429) |

**Runtime policy:** Rust, Go, or Bun TypeScript servers only — **never Node.js**. Legacy Node SQLite MCPs remain documented as out of policy in `docs/tool-mapping.md`.

---

## Engram adapter (local default)

**Operator env (required at runtime):**

| Variable | Rule |
| --- | --- |
| `ENGRAM_DATA_DIR` | Absolute path to SQLite data directory |
| `ENGRAM_PROJECT` | Project name (not a path; no `/` or `\`) |
| `ENGRAM_BIN` | Optional override; otherwise sibling `engram` next to `context-efficiency-mcp` |

The adapter spawns the child with an **explicit** environment: `ENGRAM_DATA_DIR`, sentinel `HOME` / `USERPROFILE` (not the operator home), and **no** `ENGRAM_PROJECT` / `ENGRAM_CLOUD_AUTOSYNC` / `ENGRAM_DATABASE_URL` in the child. Project name is only `--project` on the CLI.

**Auth:** Local SQLite — no API key.

**Mapping:** `docs/tool-mapping.md` (domain ↔ `mem_save` / `mem_search` / `mem_get_observation`).

**Verify:** `./scripts/check-go.sh` (unit tests always; integration when `ENGRAM_BIN` points at pinned v3.6.0).

---

## Mem0 (hosted alternate)

**URL:** `https://mcp.mem0.ai/mcp` ([Mem0 MCP](https://docs.mem0.ai/platform/mem0-mcp))

**Auth:** Browser **OAuth** through the client, or **API key** bearer token for headless clients.

**Data:** Hosted in Mem0 account. See `docs/tool-mapping.md` for tool names.

---

## Honcho (hosted alternate)

**URL:** `https://mcp.honcho.dev` ([MCP guide](https://honcho.dev/docs/v3/guides/integrations/mcp))

**Auth:** OAuth connect flow or API key on custom connector.

**Data:** Honcho cloud (or self-hosted). See `docs/tool-mapping.md`.

---

## Security

- No API keys in git for hosted backends.
- Local engram: protect `ENGRAM_DATA_DIR` on shared machines.

## Related docs

- `docs/tool-mapping.md` — per-backend tool tables
- `docs/plan.md` — domain model
