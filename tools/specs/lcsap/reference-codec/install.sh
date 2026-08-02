#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/../../../.." && pwd)
python3 -m venv "$root/tools/specs/lcsap/reference-codec/.venv"
"$root/tools/specs/lcsap/reference-codec/.venv/bin/pip" install -r "$root/tools/specs/lcsap/reference-codec/requirements-lock.txt"
