# Local end-to-end test (GMA-543)

`scripts/e2e-local.sh` is the MVP step 6 local test. It installs nothing on the host and does not touch Grok or Cursor plugin config.

1. `docker run --rm --platform linux/amd64 golang:1.26` with the repo mounted read-only. `linux/amd64` because `adapters/engram/pin.json` pins one linux_amd64 Engram digest.
2. Inside the container: fetch Engram 3.6.0 with `scripts/ci-prepare-engram.sh` (sha256-checked), build `context-efficiency-mcp`, and make a throwaway `ENGRAM_DATA_DIR` under `/tmp`.
3. `adapters/engram/cmd/ce-e2e` starts the server through `plugin/bin/run-context-efficiency-mcp.sh`, the same launcher `mcp.json` uses, and speaks MCP over stdio.
4. Steps: `retain` then `recall` finds the fact; `handoff_write` then `handoff_read` returns it; `decide` on all three plain paths returns the expected label with `effect` `none`; an unknown path is `refused:`; `send_email`, `merge`, `deploy` and `pay` are `refused:` (they are not decision-layer actions).
5. Writes `reports/local-e2e.{json,md}`. Exit 0 all pass, 2 a step failed, 1 error.

Not covered here:

- The four skills under `plugin/skills/` are agent instructions, not executables. The run proves the tools they call work.
- `decide_shadow` (Laya) is not run. Laya is not in the plugin (GMA-551). No `LAYA_API_KEY` is passed in.
- The CE-6 checks (GMA-430) stay with Kit.
- Publishing, rename, and repo move are out of scope.
