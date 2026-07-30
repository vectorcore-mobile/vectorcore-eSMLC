#!/usr/bin/env python3
"""Symbol inventory and conservative direct-reference extraction."""
from __future__ import annotations

from parser import Definition, Module


BUILTIN_WORDS = {
    "ABSENT", "ALL", "ANY", "APPLICATION", "AUTOMATIC", "BEGIN", "BIT", "BOOLEAN",
    "BY", "CHARACTER", "CHOICE", "CLASS", "COMPONENT", "COMPONENTS", "CONTAINING",
    "DEFAULT", "DEFINITIONS", "END", "ENUMERATED", "EXPLICIT", "EXPORTS", "FALSE",
    "FROM", "IMPLICIT", "IMPORTS", "INCLUDES", "INTEGER", "MAX", "MIN", "NULL", "OBJECT",
    "OCTET", "OF", "OPTIONAL", "PRESENT", "REAL", "SEQUENCE", "SET", "SIZE", "STRING",
    "SYNTAX", "TAGS", "TRUE", "UNIQUE", "WITH",
}


def catalogue(modules: list[Module]) -> tuple[dict[str, list[str]], dict[str, set[str]]]:
    providers: dict[str, list[str]] = {}
    imports: dict[str, set[str]] = {}
    for module in modules:
        imports[module.name] = {entry.symbol for entry in module.imports}
        for definition in module.definitions:
            providers.setdefault(definition.name, []).append(module.name)
    return providers, imports


def direct_references(definition: Definition, known: set[str], imported: set[str]) -> list[str]:
    """References are only identifiers declared by a module or explicitly imported.

    This intentionally excludes ASN.1 field labels, ENUMERATED labels and named
    numbers/bits: such labels are not definitions in the global symbol table.
    """
    result: list[str] = []
    for token in definition.tokens:
        if token.kind != "identifier":
            continue
        if token.text == definition.name or token.text.upper() in BUILTIN_WORDS:
            continue
        if token.text in known or token.text in imported:
            if token.text not in result:
                result.append(token.text)
    return result
