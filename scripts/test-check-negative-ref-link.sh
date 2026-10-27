#!/usr/bin/env bash
# Negative test: reference-style broken links must fail check.sh
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${REPO_ROOT}"

TMP="$(mktemp -d)"
trap 'rm -rf "${TMP}"' EXIT

cat > "${TMP}/bad.md" <<'EOF'
# bad link fixture

[missing][nope]

[nope]: ./does-not-exist.md
EOF

export CHECK_EXTRA_MD="${TMP}/bad.md"
if CHECK_EXTRA_MD="${TMP}/bad.md" python3 - "${REPO_ROOT}" <<'PY'
import os
import subprocess
import sys
from pathlib import Path

repo = Path(sys.argv[1])
extra = Path(os.environ["CHECK_EXTRA_MD"])
# Inject broken ref link into a temp copy of README check by running inline validator
text = extra.read_text(encoding="utf-8")
import re

errors = []
ref_defs = {}
for m in re.finditer(
    r"^[ \t]*\[([^\]]+)\]:\s+<?([^>\s]+)>?(?:[ \t]+[\"(].*)?$",
    text,
    re.MULTILINE,
):
    ref_defs[m.group(1).strip().lower()] = m.group(2).strip()
for m in re.finditer(r"\[([^\]]+)\]\[([^\]]*)\]", text):
    label, ref_id = m.group(1), m.group(2).strip()
    key = (ref_id or label).lower()
    if key not in ref_defs:
        errors.append(f"undefined ref {key}")
        continue
    target = (extra.parent / ref_defs[key]).resolve()
    if not target.exists():
        errors.append(f"broken ref {ref_defs[key]}")

if not errors:
    print("negative ref-link test: FAIL (expected errors, got none)")
    sys.exit(1)
print("negative ref-link test: OK (detected:", "; ".join(errors), ")")
PY
then
  :
else
  exit 1
fi

# Full check.sh must reject a repo tree with bad.md under docs/
BAD_DOC="${REPO_ROOT}/docs/_check_negative_ref.md"
cp "${TMP}/bad.md" "${BAD_DOC}"
if ./scripts/check.sh; then
  rm -f "${BAD_DOC}"
  echo "negative ref-link test: FAIL (check.sh should exit 1)"
  exit 1
fi
rm -f "${BAD_DOC}"
echo "negative ref-link test: OK (check.sh exited non-zero for broken reference link)"
