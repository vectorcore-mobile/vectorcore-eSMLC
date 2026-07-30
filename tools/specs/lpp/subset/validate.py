#!/usr/bin/env python3
"""Validate the generated subset using the same structural parser as analysis."""
from __future__ import annotations

import hashlib
import json
from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parents[4]
sys.path.insert(0, str(ROOT / "tools/specs/lpp/analyze"))
from lexer import lex_file  # noqa: E402
from parser import parse  # noqa: E402
from symbols import catalogue, direct_references  # noqa: E402

OUT = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/subsets/minimum-envelope"
PDU = OUT / "LPP-PDU-Definitions.asn1"


def main() -> None:
    module = parse(lex_file(PDU))
    providers, imports = catalogue([module])
    known = set(providers)
    unresolved = []
    for definition in module.definitions:
        for ref in direct_references(definition, known, imports[module.name]):
            if ref not in known:
                unresolved.append((definition.name, ref))
    manifest = json.loads((OUT / "manifest.json").read_text(encoding="utf-8"))
    expected = manifest["generated_modules"][PDU.name]
    actual = hashlib.sha256(PDU.read_bytes()).hexdigest()
    if actual != expected:
        raise SystemExit("generated module hash does not match its manifest")
    if unresolved:
        raise SystemExit("unresolved subset references: " + repr(unresolved))
    if len(providers) != len(module.definitions):
        raise SystemExit("duplicate subset definitions")
    print(f"subset valid: {len(module.definitions)} definitions, 0 unresolved references, 0 duplicate definitions")


if __name__ == "__main__":
    main()
