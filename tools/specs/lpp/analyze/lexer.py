#!/usr/bin/env python3
"""Tokenise the corrected TS 37.355 ASN.1 without repairing source text."""
from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
import re


@dataclass(frozen=True)
class Token:
    module: str
    line: int
    column: int
    kind: str
    text: str


KEYWORDS = {
    "ABSENT", "ALL", "ANY", "APPLICATION", "AUTOMATIC", "BEGIN", "BIT",
    "BOOLEAN", "BY", "CHARACTER", "CHOICE", "CLASS", "COMPONENT", "COMPONENTS",
    "CONSTRAINED", "CONTAINING", "DEFAULT", "DEFINED", "DEFINITIONS", "EMBEDDED",
    "END", "ENUMERATED", "EXCEPT", "EXPLICIT", "EXPORTS", "EXTENSIBILITY",
    "FALSE", "FROM", "GENERALIZEDTIME", "IDENTIFIER", "IMPLICIT", "IMPORTS",
    "INCLUDES", "INSTANCE", "INTEGER", "INTERSECTION", "MAX", "MIN", "NULL",
    "OBJECT", "OCTET", "OF", "OPTIONAL", "PATTERN", "PDV", "PRESENT", "PRIVATE",
    "REAL", "RELATIVE-OID", "SEQUENCE", "SET", "SIZE", "STRING", "SYNTAX", "TAGS",
    "TIME", "TRUE", "TYPE-IDENTIFIER", "UNION", "UNIQUE", "UNIVERSAL", "UTCTIME",
    "WITH", "COMPONENT", "SETTINGS",
}
BUILTINS = {
    "BIT", "BOOLEAN", "CHARACTER", "CHOICE", "EMBEDDED", "ENUMERATED", "GENERALIZEDTIME",
    "INTEGER", "NULL", "OBJECT", "OCTET", "REAL", "RELATIVE-OID", "SEQUENCE", "SET",
    "STRING", "TIME", "UTCTIME",
}


def lex(module: str, text: str) -> list[Token]:
    tokens: list[Token] = []
    i, line, column = 0, 1, 1
    n = len(text)

    def advance(value: str) -> None:
        nonlocal line, column
        breaks = value.count("\n")
        if breaks:
            line += breaks
            column = len(value) - value.rfind("\n")
        else:
            column += len(value)

    while i < n:
        ch = text[i]
        if ch.isspace():
            advance(ch)
            i += 1
            continue
        start_line, start_column = line, column
        if text.startswith("--", i):
            end = text.find("\n", i)
            end = n if end < 0 else end
            value = text[i:end]
            tokens.append(Token(module, start_line, start_column, "comment", value))
            advance(value)
            i = end
            continue
        if text.startswith("::=", i):
            tokens.append(Token(module, line, column, "assign", "::="))
            advance("::=")
            i += 3
            continue
        if text.startswith("...", i):
            tokens.append(Token(module, line, column, "extension", "..."))
            advance("...")
            i += 3
            continue
        if text.startswith("..", i):
            tokens.append(Token(module, line, column, "range", ".."))
            advance("..")
            i += 2
            continue
        if ch in "{}()[],;|":
            tokens.append(Token(module, line, column, "punctuation", ch))
            advance(ch)
            i += 1
            continue
        if ch == "'":
            match = re.match(r"'(?:[01]+|[0-9A-Fa-f]+)'[BH]", text[i:])
            if not match:
                raise ValueError(f"{module}:{line}:{column}: malformed bit or hex string")
            value = match.group(0)
            tokens.append(Token(module, line, column, "bit-string", value))
            advance(value)
            i += len(value)
            continue
        if ch == '"':
            end = i + 1
            while end < n and text[end] != '"':
                end += 1
            if end == n:
                raise ValueError(f"{module}:{line}:{column}: unterminated string")
            value = text[i:end + 1]
            tokens.append(Token(module, line, column, "string", value))
            advance(value)
            i = end + 1
            continue
        if ch.isdigit() or (ch == "-" and i + 1 < n and text[i + 1].isdigit()):
            match = re.match(r"-?[0-9]+", text[i:])
            assert match
            value = match.group(0)
            tokens.append(Token(module, line, column, "integer", value))
            advance(value)
            i += len(value)
            continue
        if ch.isalpha():
            match = re.match(r"[A-Za-z][A-Za-z0-9-]*", text[i:])
            assert match
            value = match.group(0)
            upper = value.upper()
            kind = "keyword" if upper in KEYWORDS else "identifier"
            tokens.append(Token(module, line, column, kind, value))
            advance(value)
            i += len(value)
            continue
        raise ValueError(f"{module}:{line}:{column}: unsupported character {ch!r}")
    return tokens


def lex_file(path: Path) -> list[Token]:
    return lex(path.stem, path.read_text(encoding="utf-8"))
