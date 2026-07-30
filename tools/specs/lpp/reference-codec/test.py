#!/usr/bin/env python3
"""Development-only validation of generated independent fixture files."""
from __future__ import annotations

import json
from pathlib import Path

import asn1tools

ROOT = Path(__file__).resolve().parents[4]
ASN1 = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/subsets/minimum-envelope/LPP-PDU-Definitions.asn1"
FIXTURES = ROOT / "tools/specs/lpp/fixtures/r16.4.0"


def main() -> None:
    manifest = json.loads((FIXTURES / "manifest.json").read_text(encoding="utf-8"))
    specification = asn1tools.compile_files([str(ASN1)], codec="uper")
    for entry in manifest["fixtures"]:
        encoded = (FIXTURES / entry["binary_file"]).read_bytes()
        if encoded.hex() != entry["hex"]:
            raise SystemExit(f"fixture hex mismatch: {entry['name']}")
        decoded = specification.decode(entry["top_level_type"], encoded)
        reencoded = bytes(specification.encode(entry["top_level_type"], decoded, check_constraints=True))
        if reencoded != encoded:
            raise SystemExit(f"fixture round-trip mismatch: {entry['name']}")
    if not all(item["rejected"] for item in manifest["negative_cases"]):
        raise SystemExit("a negative fixture was accepted")
    print(f"fixture validation passed: {len(manifest['fixtures'])} valid, {len(manifest['negative_cases'])} rejected negative cases")


if __name__ == "__main__":
    main()
