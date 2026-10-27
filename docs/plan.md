# Plan: Context Efficiency Plugin for Grok Bot

Architected using poteto-mode / architect skill principles.

## Principles Applied

- **Model the Domain**: Memory operations are a small domain (recall / retain / handoff / scope) rather than backend-specific calls.
- **Boundary Discipline**: Skills only talk to a stable tool surface. Validation and backend differences stay at the MCP boundary.
- **Laziness Protocol / Subtract Before You Add**: Do not ship a custom memory server. Compose existing open-source local MCPs + thin skill policy.
- **Guard the Context Window**: Primary goal. Every skill exists to keep the live transcript short.
- **Exhaust the Design Space**: Evaluated pure file handoffs, local SQLite MCP, Mem0, Honcho. Chose domain interface + pluggable backends.
- **Foundational Thinking**: Defined core types and scopes before writing skill text.

## Grounded Constraints

- Grok Bot plugins install MCP servers + skills only. Hooks do not run.
- Full conversation history is re-sent every turn.
- No direct transcript rewrite.
- Duplicate / new single-Bot chat is the reliable way to clear transcript weight while keeping profile, skills, and routines.
- External memory + files + skills are the available levers.

## Domain Model

```ts
type Scope = "bot" | "shared" | "routine:<name>" | "task:<id>"

type MemoryItem = {
  id: string
  text: string          // short distilled fact/decision/summary
  tags: string[]
  importance: 1 | 2 | 3
  scope: Scope
  createdAt: string
}

type Handoff = {
  goal: string
  openItems: string[]
  decisions: string[]
  filePaths: string[]
  nextSteps: string[]
}
```

Skills call tools with these shapes. Backend adapters map them.

## Plugin Package Structure

```
context-efficiency/
├── .cursor-plugin/plugin.json      # or .grok-plugin/plugin.json
├── mcp.json                        # default local MCP (user can override)
├── skills/
│   ├── context-hygiene/SKILL.md
│   ├── pre-task-recall/SKILL.md
│   ├── handoff-advisor/SKILL.md
│   └── routine-hygiene/SKILL.md
├── docs/
│   ├── backends.md                 # local / Mem0 / Honcho setup
│   └── tool-mapping.md
└── README.md
```

### plugin.json sketch

```json
{
  "name": "context-efficiency",
  "version": "0.1.0",
  "description": "Keeps Grok Bot conversations lean by moving durable state out of the transcript. Skills + pluggable memory (local OSS, Mem0, Honcho).",
  "author": { "name": "gman-robotics" },
  "skills": "./skills/",
  "mcpServers": "./mcp.json",
  "keywords": ["memory", "context", "token", "handoff", "efficiency"],
  "category": "productivity"
}
```

## The Five Components

### 1. MCP Memory Server (pluggable backbone)

Default: local open-source SQLite MCP (e.g. `simple-memory-mcp` or equivalent with add/search/list).

Optional backends (documented, not shipped):
- Mem0 hosted MCP (`https://mcp.mem0.ai/mcp`)
- Honcho (existing Grok Bot plugin or self-hosted)
- Any other MCP that exposes search + add

**Tools the skills expect** (adapters translate):
- `recall(query: string, scope?: Scope, limit?: number) → MemoryItem[]`
- `retain(text: string, tags: string[], importance: number, scope: Scope) → id`
- `handoff_write(handoff: Handoff, scope: Scope) → pathOrId`
- `handoff_read(scope: Scope) → Handoff | null`

**Tech note**: Prefer Rust or Go for any custom thin wrapper (single static binary, runs natively on Bot computer). Prefer composing existing OSS first. If implementing a custom MCP server in TypeScript, prefer **Bun** over Node.js (faster startup, single-binary option via `bun build --compile`, better native performance).

### 2. Context Hygiene Skill (`skills/context-hygiene/SKILL.md`)

Always-on behavioral rules:
- Long content (docs, logs, code, plans) → write to a file on the shared computer; chat only holds the path + one-line summary.
- After any multi-step outcome, call `retain` with a 1–3 sentence distilled fact.
- Prefer `recall` or reading a known file over re-asking or re-pasting.
- If the conversation is growing, explicitly recommend the handoff-advisor skill.
- Never paste large tool outputs into the reply; summarize and store the path.

### 3. Pre-Task Recall Skill (`skills/pre-task-recall/SKILL.md`)

Triggered at the start of non-trivial work:
1. Derive a short query from the user message.
2. Call `recall` (limit 3–5).
3. Inject only the returned items as working context.
4. State what was recalled so the user can correct it.

### 4. Handoff & Reset Advisor Skill (`skills/handoff-advisor/SKILL.md`)

Used at task end, end of day, or when context feels heavy:
1. Build a structured `Handoff`.
2. Call `handoff_write` + optionally write a markdown file under `/workspace/handoffs/`.
3. Call `retain` for the key decisions.
4. Recommend: “Duplicate me (or start a new single-Bot group chat) and have the copy read the handoff.”

### 5. Routine Hygiene Skill (`skills/routine-hygiene/SKILL.md`)

For any routine the Bot creates or runs:
- Start: short `recall` for that routine’s scope.
- End: `retain` outcome + write results to a file rather than leaving them in the turn.
- Prefer short routine prompts that point at files/memory.

## Backend Support

- **Local OSS default**: `mcp.json` uses a command for a lightweight SQLite MCP. Zero config, data stays on the Bot’s computer. Prefer Bun-based packages where available (`bunx` or compiled binary).
- **Mem0**: Replace the server entry with the hosted URL + API key header. Mapping table shows `add_memory` ↔ `retain`, `search_memories` ↔ `recall`.
- **Honcho**: Use the existing Honcho Grok Bot plugin or its MCP; same mapping.
- Skills contain no backend-specific code. Only the mapping doc and the user’s MCP config change.

## Implementation Order (for continuing agents)

1. Write the four SKILL.md files from the sketches above (keep them short and principle-aligned).
2. Add `docs/tool-mapping.md` and `docs/backends.md`.
3. Create the plugin manifest and default `mcp.json`.
4. Add a verification checklist (does a Bot follow the hygiene rules? does recall inject only short items?).
5. Prepare marketplace listing notes.

## Open Questions

- Exact local MCP package to pin as default (simple-memory-mcp vs mini-memory-mcp vs other). Prefer Bun-compatible if writing TypeScript.
- Whether a thin Rust/Go/Bun adapter is needed for perfect domain tool names, or if mapping docs are sufficient.
- License and org contribution guidelines.

This design keeps the live transcript short by design, moves continuity into selective external storage, and works within current Grok Bot constraints.
