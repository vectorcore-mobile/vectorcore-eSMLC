#!/usr/bin/env bash
set -euo pipefail
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)
venv="$repo_root/tools/specs/lpp/reference-codec/.venv"
"$venv/bin/python" "$repo_root/tools/specs/lpp/reference-codec/test.py"
