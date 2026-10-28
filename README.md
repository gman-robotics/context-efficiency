# Context Efficiency

Grok Bot plugin + skills that maximize efficiency in conversation context to reduce token usage.

The full conversation transcript is re-sent on every turn. This project moves durable state out of the live transcript into selective external memory and files so that continuity is cheap.

## Goals

- Keep live transcripts short by design
- Provide portable skills any Grok Bot can use
- Support pluggable memory backends (local open-source MCP, Mem0, Honcho)
- Package as a marketplace-ready plugin

## Design Summary

See `docs/plan.md` for the full architected plan (domain model, five components, plugin structure, backend recommendations).

Core idea: skills speak a small domain (recall / retain / handoff). Backends are adapters. The default is a local SQLite MCP so no API key is required.

## Current Status

- [x] Repository created
- [x] High-level plan documented
- [x] Engram usage skills (`plugin/skills/`, GMA-426)
- [x] plugin.json + mcp.json scaffolding (GMA-428)
- [x] Tool mapping docs for backends (`docs/tool-mapping.md`, `docs/backends.md`)
- [x] Engram v3.6.0 Go adapter (`adapters/engram/`, GMA-429)
- [ ] Marketplace submission notes

## Tech Recommendations

- **Skills**: Pure Markdown (SKILL.md). No runtime language required.
- **Optional local MCP server**: **Rust**, **Go**, or **Bun TypeScript** only — never Node.js for TypeScript servers. Prefer single-binary Rust/Go or `bun build --compile` when using Bun.
- Use `bunx` in `mcp.json` where a Bun-based package applies.
- Existing OSS options to compose rather than rewrite where they match runtime policy (see `docs/backends.md` — Node npm packages are out of policy; Bun/Rust/Go only).
- Do **not** ship a custom server unless the domain interface cannot be mapped cleanly.

## How Other Agents Should Continue

1. Read `AGENTS.md` for verify commands and PR header order.
2. Read `docs/plan.md` fully.
3. Load the four **usage** skills under `plugin/skills/` when calling memory tools; process skills from `docs/plan.md` are not shipped yet.
4. Build the adapter and configure `mcp.json` env (see `docs/backends.md`).
5. Run `./scripts/check.sh` (see `skills/verify-context-efficiency/`).
6. Update this README status when done.

License: MIT (or match org standard).

## References

- Prior design discussion on Grok Bot transcript re-sending, memory vs history, Duplicate as reset mechanism.
- pstack / poteto-mode architect principles applied (Model the Domain, Boundary Discipline, Laziness Protocol, Guard the Context Window).
