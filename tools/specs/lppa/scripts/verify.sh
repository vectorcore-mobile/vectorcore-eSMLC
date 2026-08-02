#!/bin/sh
set -eu
cd "$(dirname "$0")/../../../.."
before=$(sha256sum tools/specs/lppa/fixtures/r16.0.0/manifest.json 2>/dev/null || true)
tools/specs/lppa/scripts/generate-fixtures.sh
after=$(sha256sum tools/specs/lppa/fixtures/r16.0.0/manifest.json)
test -z "$before" || test "$before" = "$after"
if [ -x tools/specs/lppa/reference-codec/.venv/bin/python ]; then
  tools/specs/lppa/reference-codec/.venv/bin/python tools/specs/lppa/reference-codec/compile.py
fi
GOCACHE=${GOCACHE:-/tmp/vectorcore-esmlc-go-cache} go test -buildvcs=false ./internal/lppa
