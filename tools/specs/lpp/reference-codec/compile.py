#!/usr/bin/env python3
"""Attempt independent UPER compilation; never modify normative ASN.1."""
from __future__ import annotations

import hashlib
import json
from pathlib import Path
import sys

import asn1tools

ROOT = Path(__file__).resolve().parents[4]
FULL = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/normalized"
SUBSET = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/subsets/minimum-envelope/LPP-PDU-Definitions.asn1"
OUT = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/compiler"


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def attempt(name: str, paths: list[Path]) -> dict[str, object]:
    item: dict[str, object] = {"target": name, "input_paths": [str(p.relative_to(ROOT)) for p in paths],
                               "input_hashes": {p.name: digest(p) for p in paths}, "encoding_rule": "uper"}
    try:
        specification = asn1tools.compile_files([str(p) for p in paths], codec="uper")
        item.update({"status": "accepted", "available_types": sorted(specification.types)})
    except Exception as exc:
        item.update({"status": "rejected", "error_type": type(exc).__name__, "error": str(exc),
                     "classification": "unsupported-asn1-feature-or-source-syntax"})
    return item


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    results = {"compiler": "asn1tools", "compiler_version": asn1tools.__version__, "license": "MIT",
               "python": sys.version.split()[0], "encoding_rule": "uper",
               "attempts": [attempt("full-normalized-v2", [FULL / "LPP-PDU-Definitions.asn1", FULL / "LPP-Broadcast-Definitions.asn1"]),
                            attempt("minimum-envelope-compiler-complete", [SUBSET])]}
    (OUT / "reference-compiler-results.json").write_text(json.dumps(results, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    errors = [{key: value for key, value in attempt.items() if key in {"target", "status", "error_type", "error", "classification"}}
              for attempt in results["attempts"] if attempt["status"] == "rejected"]
    (OUT / "reference-compiler-errors.json").write_text(json.dumps({"errors": errors}, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    lines = ["# UPER reference compiler report", "", "Development-only compiler: **asn1tools " + asn1tools.__version__ + "** (MIT).", "", "Encoding rule requested: `uper` (unaligned PER).", ""]
    for attempt_result in results["attempts"]:
        lines.extend(["## " + str(attempt_result["target"]), "", "Status: **" + str(attempt_result["status"]) + "**."])
        if attempt_result["status"] == "accepted":
            lines.append("Accepted types: " + str(len(attempt_result["available_types"])) + ".")
        else:
            lines.append("Classification: `" + str(attempt_result["classification"]) + "`.")
            lines.append("Error: `" + str(attempt_result["error"]).replace("`", "'") + "`.")
        lines.append("")
    (OUT / "reference-compiler-report.md").write_text("\n".join(lines), encoding="utf-8")
    print("; ".join(f"{item['target']}: {item['status']}" for item in results["attempts"]))


if __name__ == "__main__":
    main()
