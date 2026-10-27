#!/usr/bin/env bash
# Repo check: Markdown structure, SKILL frontmatter, internal links, bootstrap paths.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${REPO_ROOT}"

echo "check: repo root ${REPO_ROOT}"

required=(
  AGENTS.md
  README.md
  docs/plan.md
  scripts/check.sh
  .github/pull_request_template.md
  .github/workflows/check.yml
  skills/standards-context-efficiency/SKILL.md
  skills/verify-context-efficiency/SKILL.md
)

for path in "${required[@]}"; do
  if [[ ! -f "${path}" ]]; then
    echo "check: FAIL missing required file: ${path}"
    exit 1
  fi
done

python3 - "${REPO_ROOT}" <<'PY'
from __future__ import annotations

import re
import sys
from pathlib import Path

repo = Path(sys.argv[1])
errors: list[str] = []

FRONTMATTER = re.compile(r"^---\s*\n(.*?)\n---\s*\n", re.DOTALL)


def fail(msg: str) -> None:
    errors.append(msg)


def strip_html_comments(text: str) -> str:
    return re.sub(r"<!--.*?-->", "", text, flags=re.DOTALL)


def strip_fenced_code(text: str) -> str:
    return re.sub(r"```.*?```", "", text, flags=re.DOTALL)


def check_placeholders(path: Path, text: str) -> None:
    stripped = strip_html_comments(strip_fenced_code(text))
    for match in re.finditer(r"<[a-zA-Z][a-zA-Z0-9-]*>", stripped):
        fail(f"{path}: angle-bracket placeholder {match.group(0)}")


def parse_frontmatter(path: Path, text: str) -> None:
    m = FRONTMATTER.match(text)
    if not m:
        fail(f"{path}: missing YAML frontmatter")
        return
    body = m.group(1)
    if "name:" not in body or "description:" not in body:
        fail(f"{path}: frontmatter needs name and description")


def check_internal_links(path: Path, text: str) -> None:
    # [label](./file.md) or [label](docs/foo.md) — skip http(s) and anchors-only
    for m in re.finditer(r"\]\(([^)]+)\)", text):
        target = m.group(1).strip()
        if not target or target.startswith(("http://", "https://", "mailto:", "#")):
            continue
        target = target.split("#", 1)[0]
        if not target:
            continue
        resolved = (path.parent / target).resolve()
        try:
            resolved.relative_to(repo.resolve())
        except ValueError:
            fail(f"{path}: link escapes repo: {target}")
            continue
        if not resolved.exists():
            fail(f"{path}: broken link {target} (from {path})")


md_files = sorted(repo.rglob("*.md"))
if not md_files:
    fail("no Markdown files found")

for md in md_files:
    rel = md.relative_to(repo)
    if any(part in {".git"} for part in rel.parts):
        continue
    text = md.read_text(encoding="utf-8")
    if not text.strip():
        fail(f"{rel}: empty file")
    if rel.name == "SKILL.md":
        parse_frontmatter(rel, text)
    check_placeholders(rel, text)
    check_internal_links(rel, text)

# PR template header order
pr = (repo / ".github/pull_request_template.md").read_text(encoding="utf-8")
headers = ["Tier:", "Models:", "Linear:", "Summary:", "Gates:"]
pos = -1
for h in headers:
    idx = pr.find(h)
    if idx < 0:
        fail(f".github/pull_request_template.md: missing {h}")
        continue
    if idx < pos:
        fail(".github/pull_request_template.md: headers out of order")
    pos = idx

if errors:
    print("check: FAIL")
    for e in errors:
        print(f"  - {e}")
    sys.exit(1)

print(f"check: OK ({len(md_files)} markdown files)")
PY

echo "check: all gates passed"
