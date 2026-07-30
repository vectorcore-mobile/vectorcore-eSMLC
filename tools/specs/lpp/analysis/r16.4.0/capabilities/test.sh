#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../../../" && pwd)
py="$root/tools/specs/lpp/reference-codec/.venv/bin/python"
"$py" "$root/tools/specs/lpp/analysis/r16.4.0/capabilities/generate.py"
first=$(sha256sum "$root/tools/specs/lpp/analysis/r16.4.0/capabilities/closures.json" "$root/tools/specs/lpp/fixtures/r16.4.0/capabilities/manifest.json")
"$py" "$root/tools/specs/lpp/analysis/r16.4.0/capabilities/generate.py"
second=$(sha256sum "$root/tools/specs/lpp/analysis/r16.4.0/capabilities/closures.json" "$root/tools/specs/lpp/fixtures/r16.4.0/capabilities/manifest.json")
test "$first" = "$second"
echo "capability analysis deterministic"
