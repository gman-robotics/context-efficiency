# Local end-to-end (GMA-543)

Commit: `d2b05f6cf06d64b2b7b6b8b04918f89f5641e39a`. Server: `context-efficiency-mcp 0.1.0`.

Tools listed: decide, decide_shadow, handoff_read, handoff_write, recall, retain.

Result: 15 of 15 steps pass.

| Step | Result | Detail |
|---|---|---|
| retain | pass |  |
| recall | pass |  |
| handoff_write | pass |  |
| handoff_read | pass |  |
| decide routine_go_nogo go | pass |  |
| decide routine_go_nogo no_go | pass |  |
| decide pr_class T3 | pass |  |
| decide pr_class T0 | pass |  |
| decide tool_failure_stop stop | pass |  |
| decide tool_failure_stop continue | pass |  |
| decide unknown path refused | pass |  |
| decide send_email refused | pass |  |
| decide merge refused | pass |  |
| decide deploy refused | pass |  |
| decide pay refused | pass |  |
