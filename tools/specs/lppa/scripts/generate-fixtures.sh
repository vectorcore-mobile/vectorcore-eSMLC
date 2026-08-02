#!/bin/sh
set -eu
cd "$(dirname "$0")/../../../.."
python3 tools/specs/lppa/reference-codec/generate.py
