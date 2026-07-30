#!/usr/bin/env python3
"""Generate the standards-preserving compiler-complete minimum-envelope target."""
from __future__ import annotations

import hashlib
import json
from pathlib import Path
import sys


ROOT = Path(__file__).resolve().parents[4]
sys.path.insert(0, str(ROOT / "tools/specs/lpp/analyze"))
from parser import parse  # noqa: E402
from lexer import lex_file  # noqa: E402
from symbols import catalogue, direct_references  # noqa: E402

PDU = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/normalized/LPP-PDU-Definitions.asn1"
BROADCAST = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/normalized/LPP-Broadcast-Definitions.asn1"
ANALYSIS = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/analysis"
OUT = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/subsets/minimum-envelope"
EXPECTED = {
    PDU: "ef5f33c49cb69ba599db59776d7e58b71da5e00aa788ec57a1333f22e69e2724",
    BROADCAST: "b740f1ce01929f7b80126a4a080b0ae1590fb0f5c1ad04bcf22dfdd85049ff65",
}


def digest(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def write(name: str, value: object) -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    (OUT / name).write_text(json.dumps(value, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    for path, expected in EXPECTED.items():
        if digest(path.read_bytes()) != expected:
            raise SystemExit(f"canonical input mismatch: {path.name}")
    pdu_module = parse(lex_file(PDU))
    broadcast_module = parse(lex_file(BROADCAST))
    providers, imports = catalogue([pdu_module, broadcast_module])
    pdu_names = {definition.name for definition in pdu_module.definitions}
    closure = json.loads((ANALYSIS / "bounded-outer-closure.json").read_text(encoding="utf-8"))
    strict = set(closure["strict_structural_closure"]["symbols"])
    minimum = set(closure["initial_runtime_closure"]["symbols"])
    if strict != pdu_names:
        raise SystemExit("strict closure does not exactly match the normative PDU assignment set")
    # LPP-MessageBody is a normal CHOICE, not an open type. Its alternatives
    # require every referenced message assignment to be declared. The strict
    # closure is therefore the smallest compiler-complete target that preserves
    # the normative LPP-MessageBody text unchanged.
    output_module = OUT / PDU.name
    output_module.write_bytes(PDU.read_bytes())
    included = []
    for definition in pdu_module.definitions:
        refs = direct_references(definition, set(providers), imports[pdu_module.name])
        included.append({
            "symbol": definition.name,
            "source_module": pdu_module.name,
            "assignment_kind": definition.kind,
            "original_source_line": definition.line,
            "normalized_source_sha256": EXPECTED[PDU],
            "reason": "strict compiler-complete closure" if definition.name not in minimum else "minimum-envelope structural closure",
            "closure_path": ["LPP-Message", definition.name] if definition.name in minimum else ["LPP-Message", "LPP-MessageBody", definition.name],
            "direct_dependencies": refs,
            "copied_text_sha256": digest(" ".join(token.text for token in definition.tokens).encode()),
        })
    omitted = []
    for definition in broadcast_module.definitions:
        omitted.append({"symbol": definition.name, "reason": "broadcast module is not reachable from LPP-Message",
                        "reachable_in_strict_closure": False, "positioning_method_specific": "OTDOA" in definition.name or "NR-" in definition.name,
                        "optional": False, "extension_only": "-r15" in definition.name or "-r16" in definition.name})
    manifest = {
        "target": "minimum-envelope compiler-complete subset",
        "minimum_structural_closure_count": len(minimum),
        "minimum_compiler_complete_closure_count": len(included),
        "why_expanded": "LPP-MessageBody is a normative CHOICE with non-open alternatives; preserving it unchanged requires all referenced PDU assignments and their dependencies.",
        "input_modules": {path.name: expected for path, expected in EXPECTED.items()},
        "generated_modules": {output_module.name: digest(output_module.read_bytes())},
        "included_symbol_count": len(included), "omitted_symbol_count": len(omitted),
        "generator": "tools/specs/lpp/subset/generate.py",
    }
    write("manifest.json", manifest)
    write("included-symbols.json", included)
    write("omitted-symbols.json", omitted)
    (OUT / "provenance.md").write_text(
        "# Minimum-envelope compiler-complete subset\n\n"
        "The 19-definition minimum-form envelope closure cannot be an ASN.1 module by itself. "
        "`LPP-MessageBody` is a normative CHOICE, so retaining its complete text requires every "
        "alternative definition; recursive structural resolution expands to all 646 assignments in "
        "`LPP-PDU-Definitions`. The generated compiler target is therefore a byte-identical copy of "
        "the normalized PDU module. `LPP-Broadcast-Definitions` is omitted because it is not reachable "
        "from `LPP-Message`. No definition, constraint, tag, extension marker, field order, or import was rewritten.\n",
        encoding="utf-8")
    print(f"generated compiler-complete subset: {len(included)} included, {len(omitted)} omitted")


if __name__ == "__main__":
    main()
