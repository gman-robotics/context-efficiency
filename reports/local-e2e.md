# Local end-to-end (GMA-543)

Commit: `cd7b2220cd516764ba4b4e21a3c5ae24b0a177ec`. Server: `context-efficiency-mcp 0.1.0`.

Tools listed: decide, handoff_read, handoff_write, recall, retain.

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
