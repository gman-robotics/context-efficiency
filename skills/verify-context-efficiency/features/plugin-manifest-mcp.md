# Plugin manifest and MCP config

Cursor and Grok install this repo as a plugin: manifests under `.cursor-plugin/` and `.grok-plugin/` point at usage skills and `mcp.json`, which declares the stdio MCP server launched by `plugin/bin/run-context-efficiency-mcp.sh`.

## Sub-features

- `plugin-json-valid` — `.cursor-plugin/plugin.json` and `.grok-plugin/plugin.json` parse as JSON and list `./plugin/skills/` and `./mcp.json`.
- `mcp-json-valid` — root `mcp.json` parses and references the bash wrapper for `context-efficiency` stdio MCP.
- `launcher-script` — `plugin/bin/run-context-efficiency-mcp.sh` is present and executable (adapter build is separate; see go-engram-adapter feature).

## How to get to it (user POV)

- Open the repo as a Cursor or Grok plugin (manifest paths above).
- Agents load the four Engram usage skills from `plugin/skills/`.
- MCP clients read `mcp.json` and spawn `plugin/bin/run-context-efficiency-mcp.sh` with `ENGRAM_BIN` / data-dir env vars configured for the host.

## Driving it with verify-context-efficiency

Preconditions:

- Doctor exits `0`.

Steps:

1. Run `./scripts/check.sh` — exit `0`; required-file gate includes both plugin manifests, `mcp.json`, and `plugin/bin/run-context-efficiency-mcp.sh`.
2. Confirm JSON manually if debugging: `python3 -c "import json; json.load(open('mcp.json'))"` — exit `0`.
3. Inspect manifests: `skills` and `mcpServers` fields must be relative paths that exist in the checkout.

Observable result: `check: all gates passed` and CI job `markdown-and-links` passes.

## Gotchas

- Placeholder absolute paths in `mcp.json` (`/absolute/path/to/engram`) are intentional samples; hosts must replace them — `./scripts/check.sh` only checks that manifest files exist and parse as JSON (via `json.load`); it does not assert `skills` / `mcpServers` values or launcher mode.
- Both Grok and Cursor manifests should stay in sync when adding skills or changing MCP wiring.
- Changing `mcp.json` without updating `docs/tool-mapping.md` may confuse agents; keep docs aligned in product PRs (this feature only proves on-disk layout).
