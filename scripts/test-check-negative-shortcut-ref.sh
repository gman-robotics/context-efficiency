#!/usr/bin/env bash
# Negative test: shortcut reference links [id] with [id]: path must be validated.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${REPO_ROOT}"

BAD_DOC="${REPO_ROOT}/docs/_check_negative_shortcut_ref.md"
trap 'rm -f "${BAD_DOC}"' EXIT

cat > "${BAD_DOC}" <<'EOF'
# shortcut ref fixture

See [missing] for details.

[missing]: ./does-not-exist-shortcut.md
EOF

if ./scripts/check.sh; then
  echo "negative shortcut-ref test: FAIL (check.sh should exit 1)"
  exit 1
fi

echo "negative shortcut-ref test: OK (check.sh exited non-zero for broken shortcut reference)"
