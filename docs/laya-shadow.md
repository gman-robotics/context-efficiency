# Laya shadow

Shadow-only decision layer beside the Engram adapter. Laya 0.4.22 runs in its own process. The bot keeps the caller-supplied `actual` label. `decide_shadow` stores Laya's label next to it. A miss does not change the routine.

This work adds `laya-serve`, which is a new service, so the standards scale calls it T3, a major architectural move.

## Paths

Each path calls `decide_shadow` and keeps `actual`. The product tool does not recompute `actual`.

### routine_go_nogo

Go only when the name is ok, the action is `write_file` or `retain_summary`, and blocked is false. Otherwise `no_go`. The routine process skill is [routine-hygiene](../plugin/skills/routine-hygiene/SKILL.md). The agreement gate is 90 percent by integer math.

### pr_class

First match wins. Infra, migration, or new service is T3, a major architectural move. Else new module or cross cutting is T2. Else existing module is T1. Else docs only, check only, or config tweak is T0, including typos. The agreement gate is 95 percent.

### tool_failure_stop

Stop when the same tool, the same error class, and the count is 3 or more. Otherwise continue. This path is report-only. `threshold_pct` and `pass` are JSON null. A disagreement on this path does not clear overall pass. An error on this path does.

### audit_sort

Blocker for security, data loss, or test failure. `should_fix` for a behavior bug. `nit` for naming. `note` for a typo question. Report-only. `threshold_pct` and `pass` are JSON null.

## Call

`POST /v1/systemone` only. No batch call. The client has no health method. The question id is `verdict`. The model is `english`. The instructions are `Pick one label.` Every successful tool result has `effect` `none`.

The default base URL is `http://127.0.0.1:8792`. Set `LAYA_BASE_URL` to that origin. The Go client does not read `LAYA_PORT`. If you change the port, change `LAYA_BASE_URL` to the same origin.

The base URL must be `http` or `https`, with a loopback IP host (`127.0.0.0/8` or `::1`), no userinfo, and no path other than `/`. Query and fragment are rejected.

## Refusals

`decide_shadow` returns an error that starts with `refused:` and does not call Laya or retain when any of these are true:

- The path or `actual` is outside the allow-list.
- A screened field hits the denylist.
- The body is empty or longer than 2000 bytes.
- The body contains `@` or a run of 7 digits.
- A private tag is present.

State text is not token-scanned. A body that contains the word merge is still sent. The saved row does not store that body.

Denied prefixes are `send`, `merge`, `deploy`, and `pay`. Whole tokens `charge`, `refund`, `wire`, `invoice`, and `money` are denied. `sender` and `payload` are allowed. `wireless` and `charged` are allowed.

On `laya api key not set`, `laya transport:`, `laya http`, `laya usage truncated`, `laya verdict choice missing`, `laya verdict choice rejected`, or `laya response unreadable`, Decide does not retain. The routine continues from `actual`. `laya log:` means retain failed. That is not a refusal. The routine still continues.

Truncation is `usage.truncated` JSON true, or a number greater than 0. Absent, false, 0, null, or a non-boolean non-number is success.

## Log

Retain uses scope `routine:layashadow`, tags `laya-shadow` and the path name, and importance 1. There is no topic key. The saved JSON has kind `laya_shadow`, path, request_id, state_hash, actual, shadow, agree, confidence, answer_confidence, model, and decided_at. The state body is not stored. The title clip cannot hold it.

Read the log with `recall` and scope `routine:layashadow`. The public recall cap stays 20. There is no `shadow_report` tool. The recall list is not the pass bit.

An invalid request id is replaced with 32 hex characters. The raw value is not returned.

## Report

From `adapters/engram`, with `LAYA_API_KEY` set and laya-serve up:

```bash
go run ./cmd/laya-shadow-report -out ../../reports
```

Exit 0 when the report is written and overall pass is true. Exit 2 when the report is written and overall pass is false. Exit 1 when Laya is down, the key is missing, the base URL is rejected, a fixture is inconsistent, or the file cannot be written. The command does not write Engram and does not read `ENGRAM_DATA_DIR`. `engram_log` is `skipped`.

Pass uses integer math: agree times 100 is at least the gate percent times total, errors are 0, and total is not 0. The displayed percent is one decimal and is not the pass bit. 18 agreements out of 20 on `pr_class` display 90.0 and fail the 95 percent gate. 19 of 20 pass that gate. 18 of 20 on `routine_go_nogo` pass the 90 percent gate. 17 of 20 fail it. One error fails a gated path. Overall pass needs both gated paths and zero errors on all four paths.

A live miss writes pass false and exits 2. Do not edit fixtures or the rule to lift a score.

This tree does not invent agreement numbers. Generate the report against a live server, then commit `reports/laya-shadow-agreement.json` and `reports/laya-shadow-agreement.md` as normal tracked files.

Expected version `0.4.22`. Model pin `english`. Expected weights sha256 `891102d372688fc2a094dac56a384bc537b87c63f21f9f3dac0be2b7cbc8d86c`. That hash is the pin, not a measurement taken at report time.

## Install and run

Python 3.10 or newer, from `$LAYA_VENV`. The run script checks `"$LAYA_VENV/bin/python"`, not a `python3` on `PATH`.

From the repo root, once:

```bash
"$LAYA_VENV/bin/python" -m pip install --require-hashes -r adapters/laya/laya-requirements.lock
```

Prefetch the weights once with `HF_HUB_OFFLINE` unset. After that, `adapters/laya/run-laya-serve.sh` forces offline mode. The run script does not install packages.

`LAYA_SHA256_DIGESTS` is a JSON object from artifact name to sha256. The run script exports:

```text
{"model.safetensors":"891102d372688fc2a094dac56a384bc537b87c63f21f9f3dac0be2b7cbc8d86c"}
```

`LAYA_REVISION=reviewed`. The script forces `LAYA_HOST=127.0.0.1` and defaults `LAYA_PORT` to 8792. It loads `LAYA_API_KEY` from the environment or the keychain service `gman-laya-api-key`. It does not print the key. A keychain prompt there may block. The MCP launcher must not.

The MCP launcher bounds that keychain read to about 2 seconds and still starts when the read fails or times out. `mcp.json` does not contain the key.

## Related

- [Routine skill](../plugin/skills/routine-hygiene/SKILL.md)
- [Plan](plan.md)
- [Tool mapping](tool-mapping.md)
- [Backends](backends.md)
