# Local end-to-end test (GMA-543)

`scripts/e2e-local.sh` is the MVP step 6 local test. It installs nothing on the host and does not touch Grok or Cursor plugin config.

1. `docker run --rm --platform linux/amd64 golang:1.26` with the repo mounted read-only. `linux/amd64` because `scripts/ci-prepare-engram.sh` always fetches the `linux_amd64` archive from `adapters/engram/pin.json`, and the pin's one `binary_sha256` is that build's binary. The image needs `python3` (the official `golang` image has it) and builds with `GOFLAGS=-buildvcs=false`, since a worktree's `.git` points outside the mount.
2. Inside the container: fetch Engram 3.6.0 with `scripts/ci-prepare-engram.sh` (sha256-checked), build `context-efficiency-mcp`, and make a throwaway `ENGRAM_DATA_DIR` under `/tmp`.
3. `adapters/engram/cmd/ce-e2e` starts the server through `plugin/bin/run-context-efficiency-mcp.sh`, the same launcher `mcp.json` uses, and speaks MCP over stdio.
4. Steps: `retain` then `recall` finds the fact; `handoff_write` then `handoff_read` returns it; `decide` on all three plain paths returns the expected label with `effect` `none`; an unknown path is `refused:`; `send_email`, `merge`, `deploy` and `pay` are `refused:` (they are not decision-layer actions).
5. Writes `reports/local-e2e.{json,md}`. Exit 0 all pass, 2 a step failed, 1 error.

Not covered here:

- The five skills under `plugin/skills/` (the four adapter skills plus `routine-hygiene`) are agent instructions, not executables. The run proves the tools they call work.
- The server lists exactly five tools: `decide`, `handoff_read`, `handoff_write`, `recall`, `retain`. The initialize step fails on any other set (GMA-567 removed the shadow tool). No model key is passed in.
- The CE-6 checks (GMA-430) stay with Kit.
- Publishing, rename, and repo move are out of scope.
