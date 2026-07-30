#!/usr/bin/env python3
"""Run the V2 analysis twice and record canonical report hashes."""
from __future__ import annotations

import hashlib
from pathlib import Path
import subprocess
import sys


ROOT = Path(__file__).resolve().parents[4]
ANALYZER = ROOT / "tools/specs/lpp/analyze/analyze.py"
OUT = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/analysis"
FILES = ["module-symbols.json", "imports-exports.json", "reference-graph.json",
         "unresolved-symbols.json", "external-dependencies.json", "bounded-outer-closure.json"]


def hashes() -> dict[str, str]:
    return {name: hashlib.sha256((OUT / name).read_bytes()).hexdigest() for name in FILES}


def main() -> None:
    subprocess.run([sys.executable, str(ANALYZER)], cwd=ROOT, check=True)
    first = hashes()
    subprocess.run([sys.executable, str(ANALYZER)], cwd=ROOT, check=True)
    second = hashes()
    if first != second:
        raise SystemExit("analysis outputs are not deterministic")
    rows = "\n".join(f"| `{name}` | `{digest}` |" for name, digest in first.items())
    (OUT / "analysis-determinism.md").write_text(
        "# Analysis determinism\n\n"
        "Two clean successive runs emitted byte-identical canonical JSON. "
        "Canonical reports omit timestamps.\n\n| Report | SHA-256 |\n|---|---|\n" + rows + "\n",
        encoding="utf-8")
    print("analysis determinism verified")


if __name__ == "__main__":
    main()
