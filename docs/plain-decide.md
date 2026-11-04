# Plain decide (GMA-551)

The `decide` MCP tool is the plain deterministic decision layer. It needs no model, no server, and no key.

## One source of truth

`adapters/engram/internal/decide` calls the existing `RuleGo`, `RulePR`, and `RuleStop` in `adapters/engram/internal/rules/rules.go`. Those functions encode the tier table in `skills/standards-context-efficiency/SKILL.md` and the routine and retry rules in `plugin/skills/routine-hygiene/SKILL.md`. Change a rule there, not here.

## Paths

| Path | Features | Labels |
|---|---|---|
| `routine_go_nogo` | `name_ok`, `action`, `blocked` | `go`, `no_go` |
| `pr_class` | `infra`, `migration`, `new_service`, `new_module`, `cross_cutting`, `existing_module`, `docs_only`, `check_only`, `config_tweak` | `T0`–`T3`, first match |
| `tool_failure_stop` | `same_tool`, `same_error_class`, `count`, plus optional names | `continue`, `stop` |

`audit_sort` is not a plain path; it stays as it is. Unknown paths, unknown fields, trailing data, and unmatched shapes return an error starting with `refused:`. Every result has `effect` `none`.

## Agreement report

`go run ./cmd/plain-decide-report -out ../../reports` (from `adapters/engram`) scores the embedded GMA-542 fixtures, unchanged, and writes `reports/plain-decide-agreement.{json,md}`. Exit 0 when the gates pass (go/no-go >= 90%, PR class >= 95%), 2 when a gate fails, 1 on error.

The result is 100% by construction. The fixtures' actual labels are already validated against the same rule functions, so this is not an independent result. It only shows the tool wires the rules correctly.
