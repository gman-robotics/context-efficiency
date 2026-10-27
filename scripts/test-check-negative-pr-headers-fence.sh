#!/usr/bin/env bash
# Negative test: fenced copy of PR headers must not satisfy header-order check.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${REPO_ROOT}"

PR=".github/PULL_REQUEST_TEMPLATE.md"
BACKUP="$(mktemp)"
trap 'cp "${BACKUP}" "${PR}"; rm -f "${BACKUP}"' EXIT

cp "${PR}" "${BACKUP}"

python3 - "${PR}" <<'PY'
import sys
from pathlib import Path

path = Path(sys.argv[1])
body = path.read_text(encoding="utf-8")
# Decoy fence in hub order, then reversed real headers (must fail check).
inject = """```markdown
Tier:
Models:
Linear:
Summary:
Gates:
```

Gates: x
Summary: x
Linear: x
Models: x
Tier: x

"""
path.write_text(inject + body, encoding="utf-8")
PY

if ./scripts/check.sh; then
  echo "negative pr-headers-fence test: FAIL (check.sh should exit 1)"
  exit 1
fi

echo "negative pr-headers-fence test: OK (check.sh rejected fenced-header decoy)"
