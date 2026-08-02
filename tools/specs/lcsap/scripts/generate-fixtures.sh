#!/bin/sh
set -eu
cd "$(dirname "$0")/../../../.."
python3 tools/specs/lcsap/reference-codec/generate.py
