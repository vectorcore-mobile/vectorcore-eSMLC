#!/bin/sh
set -eu
cd "$(dirname "$0")/../../../.."
before=$(sha256sum docs/specs/asn1/lcsap/r16.4.0/manifest.json tools/specs/lcsap/fixtures/r16.4.0/manifest.json)
tools/specs/lcsap/scripts/extract.sh
tools/specs/lcsap/scripts/generate-fixtures.sh
after=$(sha256sum docs/specs/asn1/lcsap/r16.4.0/manifest.json tools/specs/lcsap/fixtures/r16.4.0/manifest.json)
test "$before" = "$after"
if [ -x tools/specs/lcsap/reference-codec/.venv/bin/python ]; then
  tools/specs/lcsap/reference-codec/.venv/bin/python tools/specs/lcsap/reference-codec/compile.py
fi
GOCACHE=${GOCACHE:-/tmp/vectorcore-esmlc-go-cache} go test -buildvcs=false ./internal/lcsap
