# Laya source and license check (GMA-541)

Decision note for MVP step 3. Checked 2026-11-04 from public metadata and a read-only source review. Laya was not installed or run.

## Decision

**Go** for a pinned, shadow-only install in GMA-542, with the conditions below.

## Identity

| Item | Value |
| --- | --- |
| Source repo | https://github.com/NandhaKishorM/laya (Apache-2.0, not archived) |
| Maintainer | Nandha Kishor M, publishing as Convai Innovations (https://convaiinnovations.com/) |
| Package | PyPI `laya` (https://pypi.org/project/laya/), author Convai Innovations |
| Weights | Hugging Face `convaiinnovations/laya` (ModernBERT-large backbone, 421M params, 512-token context) |

## License

- Code: Apache-2.0 (SPDX `Apache-2.0`) in the repo `LICENSE` and in the PyPI metadata.
- Weights: the Hugging Face card is tagged `license:apache-2.0` and `commercial-use`. The model is not gated.
- No extra terms found. Apache-2.0 is compatible with an open-source plugin release. If we redistribute Laya code or weights, keep the `LICENSE` and any `NOTICE`. The plugin only calls Laya as a separate process, so it does not need to redistribute it.
- Base dependencies: torch (BSD-3), transformers, safetensors and huggingface_hub (Apache-2.0), numpy (BSD-3). No copyleft in the base install. Do not add the optional extras (`fast`, `crewai`, `agno`, `langchain` and others) without a separate check.

## Pin

| Item | Value |
| --- | --- |
| Release | `v0.4.22` (published 2026-11-03) |
| Source commit | `462c2644e654add83516e53c7a65909f8072d529` |
| Wheel sha256 | `7e8cad8ca4cb76588ef608e70ad6753bae612885e015aeeeae7b43268ec5fd5a` (`laya-0.4.22-py3-none-any.whl`) |
| Sdist sha256 | `eaed70b09d6555357ef49493d23a80678218159ced7b8d1808b6061b3d2ba529` (`laya-0.4.22.tar.gz`) |
| Provenance | PyPI attestation from GitHub trusted publishing (`NandhaKishorM/laya`, `release.yml`, environment `pypi`) |
| Weights revision | `55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851`. This is the package's own reviewed pin, selected with `LAYA_REVISION=reviewed`. |
| Weights sha256 | `891102d372688fc2a094dac56a384bc537b87c63f21f9f3dac0be2b7cbc8d86c` (`model.safetensors`, 842,609,210 bytes) |

Install with `pip install --require-hashes` from a lock file that holds these hashes, in a venv inside the project. Do not do a system-wide install.

## Runtime and interface

- Python 3.10 or newer. Torch runs on the CPU or on Apple MPS.
- To run it as a separate process, use `laya-serve` (extra `laya[serve]`: FastAPI and uvicorn). Endpoints: `GET /health`, `POST /v1/route`, `POST /v1/systemone` and `POST /v1/systemone/batch`. All settings come from environment variables.
- The GMA-429 Go adapter can call this HTTP API. The upstream `laya-go` module is a pure-Go port and needs cgo ONNX to run the model. Keep it out of the MVP.
- Other entry points: `laya` CLI (stdin requests) and `laya-mcp-server` (extra `mcp`).

## Network access

- The model downloads from huggingface.co on first use. After that, it runs offline with `HF_HUB_OFFLINE=1`.
- The source review found no telemetry or analytics calls. `urllib` shows up only in the remote MCP client and in the CrewAI and Agno integrations, which call a `laya-serve` URL that you configure.
- Set `HF_HUB_DISABLE_TELEMETRY=1` to switch off Hub library telemetry.

## Maintenance health

- The last commit was 2026-11-03. The repo has about 35k stars and 91 open issues plus PRs. The top contributors have 1,036, 517 and 318 commits.
- Releases come out very often: v0.4.20, v0.4.21 and v0.4.22 shipped between 2026-11-01 and 2026-11-03.

## Red flags and conditions

1. **`laya-serve` binds to `0.0.0.0` by default.** Always set `LAYA_HOST=127.0.0.1`.
2. **Fast release cadence.** Pin the exact version, the wheel hash and the weights revision. Do not auto-upgrade.
3. **The weights repo moves.** Its `main` is at `7b928d8`. Load only the reviewed revision above.
4. **Accuracy and cost claims are author claims.** Check them in shadow mode before any path goes active.
5. **Shadow only.** Nothing may send, merge, deploy or touch money, as the P-GMA-13 rules say.
