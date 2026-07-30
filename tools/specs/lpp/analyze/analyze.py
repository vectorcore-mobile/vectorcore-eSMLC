#!/usr/bin/env python3
"""Generate deterministic structural reports from corrected TS 37.355 V2 ASN.1."""
from __future__ import annotations

import argparse
import hashlib
from pathlib import Path
import sys

from closure import INITIAL_MINIMUM_FORM, ROOTS, classify, reachable
from lexer import lex_file
from parser import parse
from report import sha256, write_json, write_markdown
from symbols import catalogue, direct_references


ROOT = Path(__file__).resolve().parents[4]
INPUT = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/normalized"
OUTPUT = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/analysis"
EXPECTED = {
    "LPP-PDU-Definitions.asn1": "ef5f33c49cb69ba599db59776d7e58b71da5e00aa788ec57a1333f22e69e2724",
    "LPP-Broadcast-Definitions.asn1": "b740f1ce01929f7b80126a4a080b0ae1590fb0f5c1ad04bcf22dfdd85049ff65",
}


def canonical_modules() -> list[Path]:
    files = [INPUT / name for name in sorted(EXPECTED)]
    for path in files:
        if not path.is_file():
            raise SystemExit(f"missing canonical V2 module: {path}")
        got = sha256(path)
        if got != EXPECTED[path.name]:
            raise SystemExit(f"hash mismatch for {path.name}: got {got}, expected {EXPECTED[path.name]}")
    return files


def md_table(rows: list[list[str]], headings: list[str]) -> str:
    body = ["| " + " | ".join(headings) + " |", "|" + "|".join("---" for _ in headings) + "|"]
    body.extend("| " + " | ".join(value.replace("|", "\\|") for value in row) + " |" for row in rows)
    return "\n".join(body)


def generate() -> None:
    paths = canonical_modules()
    modules = [parse(lex_file(path)) for path in paths]
    providers, imports_by_module = catalogue(modules)
    known = set(providers)
    by_symbol = {definition.name: (module, definition)
                 for module in modules for definition in module.definitions}
    duplicates = {name: value for name, value in providers.items() if len(value) > 1}
    symbols = []
    graph: dict[str, list[str]] = {}
    for module in modules:
        imported = imports_by_module[module.name]
        for definition in module.definitions:
            refs = direct_references(definition, known, imported)
            graph[definition.name] = refs
            symbols.append({
                "name": definition.name, "module": module.name, "assignment_kind": definition.kind,
                "line": definition.line, "column": definition.column, "exported": module.exports != "EXPORTS ;",
                "imported": definition.name in imported, "parameterized": definition.parameterized,
                "direct_references": refs,
            })
    imports = []
    for module in modules:
        for entry in module.imports:
            providers_for_symbol = providers.get(entry.symbol, [])
            internally_resolved = entry.source_module in providers_for_symbol
            imports.append({"importing_module": module.name, "symbol": entry.symbol,
                            "source_module": entry.source_module, "line": entry.line, "column": entry.column,
                            "internal": internally_resolved, "external": not internally_resolved,
                            "resolution": "resolved" if internally_resolved else "unresolved"})
    unresolved_refs = []
    for module in modules:
        imported = imports_by_module[module.name]
        for definition in module.definitions:
            for ref in direct_references(definition, known, imported):
                if ref not in known:
                    unresolved_refs.append({"symbol": ref, "module": module.name,
                                            "definition": definition.name, "line": definition.line,
                                            "category": "imported-but-missing"})
    strict_symbols, strict_unresolved = reachable(graph, ROOTS)
    initial_symbols = sorted(INITIAL_MINIMUM_FORM)
    initial_unresolved = sorted(symbol for symbol in initial_symbols if symbol not in graph)
    root_set = set(ROOTS) | set(INITIAL_MINIMUM_FORM)
    closure_entries = [{"symbol": name, "classification": classify(name, root_set, set(initial_unresolved)),
                        "module": by_symbol[name][0].name if name in by_symbol else None}
                       for name in initial_symbols]
    external_modules: dict[str, set[str]] = {}
    for entry in imports:
        if entry["external"]:
            external_modules.setdefault(entry["source_module"], set()).add(entry["symbol"])
    external = [{"module": module, "referenced_symbols": sorted(values),
                 "likely_origin": "not resolved by the supplied TS 37.355 V2 modules",
                 "required_for_bounded_runtime": any(name in initial_symbols for name in values)}
                for module, values in sorted(external_modules.items())]
    summary = {
        "input_hashes": {path.name: sha256(path) for path in paths}, "module_count": len(modules),
        "definition_count": len(symbols),
        "type_assignment_count": sum(s["assignment_kind"] == "type" for s in symbols),
        "value_assignment_count": sum(s["assignment_kind"] == "value" for s in symbols),
        "object_assignment_count": sum(s["assignment_kind"] in {"object", "object-set", "object-class"} for s in symbols),
        "parameterized_assignment_count": sum(s["parameterized"] for s in symbols),
        "import_count": len(imports), "external_import_count": sum(item["external"] for item in imports),
        "duplicate_definition_count": len(duplicates), "unresolved_reference_count": len(unresolved_refs),
    }
    write_json(OUTPUT / "module-symbols.json", {"summary": summary, "modules": [
        {"name": module.name, "tagging": module.tagging, "exports": module.exports,
         "definition_count": len(module.definitions)} for module in modules], "symbols": sorted(symbols, key=lambda x: (x["module"], x["line"], x["name"])),
        "duplicates": duplicates})
    write_json(OUTPUT / "imports-exports.json", {"summary": summary, "imports": sorted(imports, key=lambda x: (x["importing_module"], x["line"], x["symbol"])),
                                                    "exports": [{"module": m.name, "declaration": m.exports} for m in modules]})
    write_json(OUTPUT / "reference-graph.json", {"summary": summary, "edges": [{"from": key, "to": target,
            "kind": "local-or-imported-definition"} for key in sorted(graph) for target in graph[key]], "nodes": sorted(graph)})
    write_json(OUTPUT / "unresolved-symbols.json", {"summary": summary, "unresolved": unresolved_refs})
    write_json(OUTPUT / "external-dependencies.json", {"summary": summary, "dependencies": external})
    write_json(OUTPUT / "bounded-outer-closure.json", {
        "roots": ROOTS, "strict_structural_closure": {"symbols": strict_symbols, "unresolved": strict_unresolved},
        "initial_runtime_closure": {"symbols": initial_symbols, "unresolved": initial_unresolved,
                                    "entries": closure_entries},
        "notes": ["The strict closure contains every definition structurally reachable through all selected message branches.",
                  "The initial runtime closure is a minimum-form envelope list: it covers selected R9 wrapper branches whose root IE members are OPTIONAL, not arbitrary positioning payload decoding.",
                  "A future compiler-subset phase must prove every deferred boundary with ASN.1 encoding semantics."]})
    write_markdown(OUTPUT / "module-symbols.md", "# TS 37.355 V2 symbol inventory\n\n" +
                   f"Two normalized modules contain **{summary['definition_count']}** assignments. "
                   f"There are {summary['duplicate_definition_count']} duplicate definitions and "
                   f"{summary['unresolved_reference_count']} unresolved structural references.\n\n" +
                   md_table([[s["module"], s["name"], s["assignment_kind"], str(s["line"]), ", ".join(s["direct_references"])] for s in sorted(symbols, key=lambda x: (x["module"], x["line"]))],
                            ["Module", "Symbol", "Kind", "Line", "Direct known references"]))
    write_markdown(OUTPUT / "imports-exports.md", "# Imports and exports\n\n" +
                   md_table([[i["importing_module"], i["symbol"], i["source_module"], i["resolution"]] for i in imports],
                            ["Importer", "Symbol", "Source module", "Resolution"]) + "\n\n" +
                   md_table([[m.name, m.exports] for m in modules], ["Module", "Exports"]))
    write_markdown(OUTPUT / "reference-graph.md", "# Reference graph\n\n" +
                   f"{len(graph)} definition nodes and {sum(len(v) for v in graph.values())} conservative edges. "
                   "Edges contain only declared local definitions or declared imports; ASN.1 field labels and enumerated labels are deliberately excluded.\n")
    write_markdown(OUTPUT / "unresolved-symbols.md", "# Unresolved symbols\n\n" +
                   ("No unresolved identifier survived conservative resolution.\n" if not unresolved_refs else md_table([[u["symbol"], u["module"], u["definition"], str(u["line"]), u["category"]] for u in unresolved_refs], ["Symbol", "Module", "Definition", "Line", "Category"])))
    write_markdown(OUTPUT / "external-dependencies.md", "# External dependencies\n\n" +
                   ("No import in the supplied V2 module pair resolves outside that pair.\n" if not external else md_table([[d["module"], ", ".join(d["referenced_symbols"]), str(d["required_for_bounded_runtime"])] for d in external], ["Module", "Symbols", "Required now"])))
    rows = [[entry["symbol"], entry["module"] or "unresolved", entry["classification"]] for entry in closure_entries]
    write_markdown(OUTPUT / "bounded-outer-closure.md", "# Bounded outer LPP structural closure\n\n" +
                   f"The strict closure has **{len(strict_symbols)}** definitions; the selected minimum-form initial envelope closure has **{len(initial_symbols)}**. "
                   "The latter is not a compiler subset or a runtime implementation.\n\n" +
                   md_table(rows, ["Symbol", "Module", "Classification"]))
    print(f"analysed {summary['definition_count']} definitions in {summary['module_count']} modules")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--verify-inputs", action="store_true")
    args = parser.parse_args()
    if args.verify_inputs:
        canonical_modules()
        print("canonical V2 inputs verified")
        return
    generate()


if __name__ == "__main__":
    main()
