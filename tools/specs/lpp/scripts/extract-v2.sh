#!/bin/sh
set -eu
exec python3 "$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)/extract_v2/extract.py"
