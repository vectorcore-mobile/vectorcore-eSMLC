#!/usr/bin/env bash
set -euo pipefail
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)
venv="$repo_root/tools/specs/lpp/reference-codec/.venv"
python3 -m venv "$venv"
"$venv/bin/python" -m pip install --disable-pip-version-check --require-hashes --no-build-isolation --requirement "$repo_root/tools/specs/lpp/reference-codec/requirements-lock.txt"
"$venv/bin/python" "$repo_root/tools/specs/lpp/reference-codec/verify.py"
