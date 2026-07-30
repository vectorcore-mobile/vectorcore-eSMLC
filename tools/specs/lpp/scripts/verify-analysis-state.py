#!/usr/bin/env python3
"""Fail closed unless the canonical V2 analysis state matches this handoff."""
from __future__ import annotations

import json
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[4]
ANALYZE = ROOT / "tools/specs/lpp/analyze/analyze.py"
BASE = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/analysis"


def main() -> None:
    subprocess.run([sys.executable, str(ANALYZE)], cwd=ROOT, check=True)
    summary = json.loads((BASE / "module-symbols.json").read_text(encoding="utf-8"))["summary"]
    imports = json.loads((BASE / "imports-exports.json").read_text(encoding="utf-8"))["imports"]
    closure = json.loads((BASE / "bounded-outer-closure.json").read_text(encoding="utf-8"))
    external = json.loads((BASE / "external-dependencies.json").read_text(encoding="utf-8"))["dependencies"]
    expected = {"module_count": 2, "definition_count": 652, "type_assignment_count": 615,
                "value_assignment_count": 37, "import_count": 5, "external_import_count": 0,
                "duplicate_definition_count": 0, "unresolved_reference_count": 0}
    if any(summary[key] != value for key, value in expected.items()):
        raise SystemExit("analysis summary does not match the canonical V2 handoff")
    if len(imports) != 5 or not all(item["internal"] for item in imports) or external:
        raise SystemExit("canonical V2 import/external-dependency state changed")
    if len(closure["strict_structural_closure"]["symbols"]) != 646:
        raise SystemExit("strict closure count changed")
    if len(closure["initial_runtime_closure"]["symbols"]) != 19:
        raise SystemExit("minimum envelope closure count changed")
    print("canonical V2 analysis state verified")


if __name__ == "__main__":
    main()
