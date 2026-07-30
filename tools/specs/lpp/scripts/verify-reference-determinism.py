#!/usr/bin/env python3
"""Verify deterministic subset, compiler report, and fixture outputs."""
from __future__ import annotations

import hashlib
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[4]
VENV = ROOT / "tools/specs/lpp/reference-codec/.venv/bin/python"
SUBSET = ROOT / "tools/specs/lpp/subset/generate.py"
COMPILER = ROOT / "tools/specs/lpp/reference-codec/compile.py"
FIXTURES = ROOT / "tools/specs/lpp/reference-codec/encode_fixtures.py"
FILES = [
    ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/subsets/minimum-envelope/LPP-PDU-Definitions.asn1",
    ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/subsets/minimum-envelope/manifest.json",
    ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/compiler/reference-compiler-results.json",
    ROOT / "tools/specs/lpp/fixtures/r16.4.0/manifest.json",
]


def state():
    return {str(path.relative_to(ROOT)): hashlib.sha256(path.read_bytes()).hexdigest() for path in FILES}


def run():
    subprocess.run([sys.executable, str(SUBSET)], cwd=ROOT, check=True)
    subprocess.run([str(VENV), str(COMPILER)], cwd=ROOT, check=True, stdout=subprocess.DEVNULL)
    subprocess.run([str(VENV), str(FIXTURES)], cwd=ROOT, check=True)


def main():
    if not VENV.is_file():
        raise SystemExit("reference compiler environment is missing")
    run(); first = state()
    run(); second = state()
    if first != second:
        raise SystemExit("subset/compiler/fixture workflow is not deterministic")
    report = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/compiler/determinism-report.md"
    report.parent.mkdir(parents=True, exist_ok=True)
    rows = "\n".join(f"| `{path}` | `{digest}` |" for path, digest in first.items())
    report.write_text("# Compiler and fixture determinism\n\nTwo successive runs produced identical subset, compiler-result, and fixture-manifest hashes.\n\n| Artifact | SHA-256 |\n|---|---|\n" + rows + "\n", encoding="utf-8")
    print("reference compiler determinism verified")


if __name__ == "__main__":
    main()
