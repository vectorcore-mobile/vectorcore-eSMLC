#!/usr/bin/env python3
"""Structural, deliberately non-semantic ASN.1 parser for analysis reports."""
from __future__ import annotations

from dataclasses import dataclass
from lexer import Token


@dataclass
class Import:
    symbol: str
    source_module: str
    line: int
    column: int


@dataclass
class Definition:
    name: str
    kind: str
    line: int
    column: int
    start: int
    end: int
    parameterized: bool
    tokens: list[Token]


@dataclass
class Module:
    name: str
    tagging: str
    exports: str
    imports: list[Import]
    definitions: list[Definition]


def significant(tokens: list[Token]) -> list[Token]:
    return [token for token in tokens if token.kind != "comment"]


def _find_keyword(tokens: list[Token], name: str, start: int = 0) -> int:
    name = name.upper()
    for i in range(start, len(tokens)):
        if tokens[i].text.upper() == name:
            return i
    return -1


def _match(tokens: list[Token], start: int, left: str, right: str) -> int:
    depth = 0
    for i in range(start, len(tokens)):
        if tokens[i].text == left:
            depth += 1
        elif tokens[i].text == right:
            depth -= 1
            if depth == 0:
                return i
    raise ValueError(f"{tokens[start].module}:{tokens[start].line}: unclosed {left}")


def _split_import_groups(tokens: list[Token]) -> list[tuple[list[Token], str]]:
    groups: list[tuple[list[Token], str]] = []
    current: list[Token] = []
    i = 0
    while i < len(tokens):
        token = tokens[i]
        if token.text.upper() == "FROM":
            if i + 1 >= len(tokens) or tokens[i + 1].kind != "identifier":
                raise ValueError(f"{token.module}:{token.line}: IMPORTS FROM lacks module")
            groups.append((current, tokens[i + 1].text))
            current = []
            i += 2
            # Skip an optional object identifier in braces.
            if i < len(tokens) and tokens[i].text == "{":
                i = _match(tokens, i, "{", "}") + 1
            if i < len(tokens) and tokens[i].text == ",":
                i += 1
            continue
        if token.text != ",":
            current.append(token)
        i += 1
    if current:
        raise ValueError(f"{current[0].module}:{current[0].line}: IMPORTS group lacks FROM")
    return groups


def _definition_kind(lhs: list[Token], rhs: list[Token]) -> tuple[str, bool]:
    parameterized = any(token.text == "{" for token in lhs)
    words = {token.text.upper() for token in rhs[:10]}
    if "CLASS" in words:
        return "object-class", parameterized
    if "OBJECT" in words and "SET" in words:
        return "object-set", parameterized
    if "OBJECT" in words:
        return "object", parameterized
    if rhs and rhs[0].kind == "integer":
        return "value", parameterized
    if lhs and lhs[-1].text.upper() not in {"TYPE", "CLASS"} and len(lhs) > 1:
        return "value", parameterized
    return "type", parameterized


def parse(tokens: list[Token]) -> Module:
    tokens = significant(tokens)
    if not tokens:
        raise ValueError("empty ASN.1 module")
    definitions_at = _find_keyword(tokens, "DEFINITIONS")
    begin_at = _find_keyword(tokens, "BEGIN")
    end_at = _find_keyword(tokens, "END", begin_at + 1)
    if definitions_at < 1 or begin_at < definitions_at or end_at < begin_at:
        raise ValueError(f"{tokens[0].module}: incomplete module envelope")
    name = tokens[0].text
    if tokens[0].kind != "identifier":
        raise ValueError(f"{tokens[0].module}:{tokens[0].line}: module name missing")
    tagging = " ".join(token.text for token in tokens[definitions_at + 1:begin_at] if token.text != "::=")
    body_start, body_end = begin_at + 1, end_at
    cursor = body_start
    exports = "EXPORTS ALL"
    if cursor < body_end and tokens[cursor].text.upper() == "EXPORTS":
        semi = cursor
        while semi < body_end and tokens[semi].text != ";":
            semi += 1
        if semi == body_end:
            raise ValueError(f"{name}:{tokens[cursor].line}: unterminated EXPORTS")
        exports = " ".join(token.text for token in tokens[cursor:semi + 1])
        cursor = semi + 1
    imports: list[Import] = []
    if cursor < body_end and tokens[cursor].text.upper() == "IMPORTS":
        semi = cursor
        depth = 0
        while semi < body_end:
            if tokens[semi].text == "{": depth += 1
            if tokens[semi].text == "}": depth -= 1
            if tokens[semi].text == ";" and depth == 0: break
            semi += 1
        if semi == body_end:
            raise ValueError(f"{name}:{tokens[cursor].line}: unterminated IMPORTS")
        for symbols, source_module in _split_import_groups(tokens[cursor + 1:semi]):
            for symbol in symbols:
                if symbol.kind == "identifier":
                    imports.append(Import(symbol.text, source_module, symbol.line, symbol.column))
        cursor = semi + 1
    assignments = [i for i in range(cursor, body_end) if tokens[i].kind == "assign"]

    def assignment_name_index(assign_at: int) -> int:
        """Find the final identifier in an ASN.1 assignment left-hand side.

        The normalized TS 37.355 modules use ordinary type assignments for the
        analysis roots. Parameterised assignments may put a brace group between
        the name and `::=`; skip that group before locating the name.
        """
        i = assign_at - 1
        if i >= cursor and tokens[i].text == "}":
            depth = 0
            while i >= cursor:
                if tokens[i].text == "}": depth += 1
                elif tokens[i].text == "{":
                    depth -= 1
                    if depth == 0:
                        i -= 1
                        break
                i -= 1
        while i >= cursor and tokens[i].kind != "identifier":
            i -= 1
        if i < cursor:
            raise ValueError(f"{name}:{tokens[assign_at].line}: assignment name missing")
        return i

    name_indexes = [assignment_name_index(assign_at) for assign_at in assignments]
    definitions: list[Definition] = []
    for index, assign_at in enumerate(assignments):
        lhs_start = name_indexes[index]
        lhs = tokens[lhs_start:assign_at]
        if not lhs:
            raise ValueError(f"{name}:{tokens[assign_at].line}: assignment without left side")
        lhs_names = [token for token in lhs if token.kind == "identifier"]
        if not lhs_names:
            raise ValueError(f"{name}:{tokens[assign_at].line}: assignment name missing")
        lhs_name = lhs_names[0]
        next_start = name_indexes[index + 1] if index + 1 < len(assignments) else body_end
        rhs = tokens[assign_at + 1:next_start]
        kind, parameterized = _definition_kind(lhs, rhs)
        definitions.append(Definition(lhs_name.text, kind, lhs_name.line, lhs_name.column,
                                      lhs_start, next_start, parameterized,
                                      tokens[lhs_start:next_start]))
    return Module(name, tagging.strip(), exports, imports, definitions)
