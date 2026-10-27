#!/usr/bin/env bash
# Negative test: reference definition inside a fence must not satisfy shortcut links.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${REPO_ROOT}"

BAD_DOC="${REPO_ROOT}/docs/_check_negative_shortcut_fence.md"
trap 'rm -f "${BAD_DOC}"' EXIT

cat > "${BAD_DOC}" <<'EOF'
# shortcut ref hidden in fence

See [missing] for details.

[missing]: ./does-not-exist-outside-fence.md

~~~
[missing]: ./plan.md
~~~
EOF

if ./scripts/check.sh; then
  echo "negative shortcut-ref-in-fence test: FAIL (check.sh should exit 1)"
  exit 1
fi

echo "negative shortcut-ref-in-fence test: OK (check.sh exited non-zero)"
