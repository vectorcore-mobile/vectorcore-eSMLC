#!/usr/bin/env bash
set -euo pipefail
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)
venv="$repo_root/tools/specs/lpp/reference-codec/.venv"
if [[ ! -x "$venv/bin/python" ]]; then
    echo "reference compiler is not installed; run tools/specs/lpp/reference-codec/install.sh" >&2
    exit 2
fi
cd "$repo_root"
"$venv/bin/python" tools/specs/lpp/reference-codec/verify.py
"$venv/bin/python" tools/specs/lpp/reference-codec/compile.py
