#!/usr/bin/env bash
# Repo check: Markdown structure, SKILL frontmatter, internal links, hub PR template.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${REPO_ROOT}"

echo "check: repo root ${REPO_ROOT}"

PR_TEMPLATE=".github/PULL_REQUEST_TEMPLATE.md"

required=(
  AGENTS.md
  README.md
  docs/plan.md
  scripts/check.sh
  "${PR_TEMPLATE}"
  scripts/fixtures/hub-pr-template-tier-body.md
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

python3 - "${REPO_ROOT}" "${PR_TEMPLATE}" <<'PY'
from __future__ import annotations

import re
import sys
from pathlib import Path

repo = Path(sys.argv[1])
pr_rel = sys.argv[2]
pr_path = repo / pr_rel
errors: list[str] = []

FRONTMATTER = re.compile(r"^---\s*\n(.*?)\n---\s*\n", re.DOTALL)
GITHUB_CANONICAL_SOURCE_PREFIX = "<!-- Canonical source:"
HUB_PR_LITERALS = ("standards-<repo>",)


def fail(msg: str) -> None:
    errors.append(msg)


def strip_html_comments(text: str) -> str:
    return re.sub(r"<!--.*?-->", "", text, flags=re.DOTALL)


def strip_fenced_code(text: str) -> str:
    return re.sub(r"```.*?```", "", text, flags=re.DOTALL)


def normalize_github_pr_template(text: str) -> str:
    lines = [
        line
        for line in text.splitlines()
        if not line.strip().startswith(GITHUB_CANONICAL_SOURCE_PREFIX)
    ]
    while lines and not lines[-1].strip():
        lines.pop()
    return "\n".join(lines).rstrip() + "\n"


pr_rel_path = Path(pr_rel)
HUB_TEMPLATE_PATHS = frozenset(
    {
        pr_rel_path,
        Path("scripts/fixtures/hub-pr-template-tier-body.md"),
    }
)


def check_placeholders(path: Path, text: str) -> None:
    stripped = strip_html_comments(strip_fenced_code(text))
    if path in HUB_TEMPLATE_PATHS:
        for lit in HUB_PR_LITERALS:
            stripped = stripped.replace(lit, "")
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


def resolve_link_target(path: Path, target: str) -> Path | None:
    target = target.strip()
    if not target or target.startswith(("http://", "https://", "mailto:", "#")):
        return None
    target = target.split("#", 1)[0]
    if not target:
        return None
    resolved = (path.parent / target).resolve()
    try:
        resolved.relative_to(repo.resolve())
    except ValueError:
        fail(f"{path}: link escapes repo: {target}")
        return None
    return resolved


def reference_definitions(text: str) -> dict[str, str]:
    defs: dict[str, str] = {}
    for m in re.finditer(
        r"^[ \t]*\[([^\]]+)\]:\s+<?([^>\s]+)>?(?:[ \t]+[\"(].*)?$",
        text,
        re.MULTILINE,
    ):
        defs[m.group(1).strip().lower()] = m.group(2).strip()
    return defs


def check_internal_links(path: Path, text: str) -> None:
    ref_defs = reference_definitions(text)

    for m in re.finditer(r"\]\(([^)]+)\)", text):
        target = m.group(1).strip()
        resolved = resolve_link_target(path, target)
        if resolved is None:
            continue
        if not resolved.exists():
            fail(f"{path}: broken link {target} (from {path})")

    for m in re.finditer(r"\[([^\]]+)\]\[([^\]]*)\]", text):
        label, ref_id = m.group(1), m.group(2).strip()
        key = (ref_id or label).lower()
        if key not in ref_defs:
            fail(f"{path}: undefined reference link [{label}][{ref_id or label}]")
            continue
        resolved = resolve_link_target(path, ref_defs[key])
        if resolved is None:
            continue
        if not resolved.exists():
            fail(f"{path}: broken reference link [{key}] -> {ref_defs[key]} (from {path})")


def check_pr_template_headers(text: str) -> None:
    visible = strip_html_comments(text)
    headers = ["Tier:", "Models:", "Linear:", "Summary:", "Gates:"]
    positions: list[int] = []
    for h in headers:
        m = re.search(rf"^{re.escape(h)}", visible, re.MULTILINE)
        if not m:
            fail(f"{pr_path}: missing line-start header {h}")
            continue
        positions.append(m.start())
    if len(positions) == len(headers) and sorted(positions) != positions:
        fail(f"{pr_path}: PR headers out of order (use line-start Tier/Models/Linear/Summary/Gates)")

    for section in (
        "## pstack (Poteto) principles cited",
        "## Verification evidence",
        "## Checklist",
    ):
        if section not in visible:
            fail(f"{pr_path}: missing hub section {section}")


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

check_pr_template_headers(pr_path.read_text(encoding="utf-8"))

fixture = repo / "scripts/fixtures/hub-pr-template-tier-body.md"
expected = normalize_github_pr_template(fixture.read_text(encoding="utf-8"))
actual = normalize_github_pr_template(pr_path.read_text(encoding="utf-8"))
if actual != expected:
    fail(
        f"{pr_path}: body does not match scripts/fixtures/hub-pr-template-tier-body.md "
        "(hub Tier: through checklist)"
    )

if errors:
    print("check: FAIL")
    for e in errors:
        print(f"  - {e}")
    sys.exit(1)

print(f"check: OK ({len(md_files)} markdown files)")
PY

echo "check: all gates passed"
