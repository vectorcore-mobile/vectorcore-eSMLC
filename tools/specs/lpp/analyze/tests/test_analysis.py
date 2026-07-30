#!/usr/bin/env python3
import sys
from pathlib import Path
import unittest

HERE = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(HERE))
from lexer import lex
from parser import parse
from symbols import catalogue, direct_references
from closure import reachable


class AnalysisTests(unittest.TestCase):
    def parse_text(self, name, text):
        return parse(lex(name, text))

    def test_lexer_preserves_locations_and_comments(self):
        tokens = lex("X", "A ::= INTEGER (0..255), ..., 'AF'H -- comment\n")
        self.assertEqual(tokens[0].text, "A")
        self.assertEqual(tokens[0].line, 1)
        self.assertTrue(any(token.kind == "comment" for token in tokens))
        self.assertTrue(any(token.kind == "range" for token in tokens))
        self.assertTrue(any(token.kind == "extension" for token in tokens))
        self.assertTrue(any(token.kind == "bit-string" for token in tokens))

    def test_imports_and_definitions(self):
        m = self.parse_text("A", "A DEFINITIONS ::= BEGIN IMPORTS T FROM B; EXPORTS ALL; T ::= SEQUENCE { field T OPTIONAL } END")
        self.assertEqual(m.imports[0].symbol, "T")
        self.assertEqual(m.imports[0].source_module, "B")
        self.assertEqual(m.definitions[0].name, "T")

    def test_repository_local_structural_fixture(self):
        source = (HERE.parent / "testdata/structural-sample.asn1").read_text(encoding="utf-8")
        m = self.parse_text("Sample", source)
        self.assertEqual(m.name, "Sample")
        self.assertEqual([entry.symbol for entry in m.imports], ["ImportedType"])
        self.assertEqual([definition.name for definition in m.definitions], ["Example", "LocalType"])

    def test_field_and_enumerated_labels_are_not_references(self):
        a = self.parse_text("A", "A DEFINITIONS ::= BEGIN T ::= SEQUENCE { field U OPTIONAL, status ENUMERATED { good, bad } } U ::= INTEGER END")
        providers, imports = catalogue([a])
        refs = direct_references(a.definitions[0], set(providers), imports[a.name])
        self.assertEqual(refs, ["U"])

    def test_parameterized_definition_and_nested_braces(self):
        m = self.parse_text("A", "A DEFINITIONS ::= BEGIN T {X} ::= SEQUENCE { a INTEGER (0..3), ... } END")
        self.assertTrue(m.definitions[0].parameterized)

    def test_malformed_module_fails(self):
        with self.assertRaises(ValueError):
            self.parse_text("A", "A DEFINITIONS ::= BEGIN T ::= INTEGER")

    def test_closure_cycle(self):
        found, unresolved = reachable({"A": ["B"], "B": ["A"]}, ["A"])
        self.assertEqual(found, ["A", "B"])
        self.assertEqual(unresolved, [])

    def test_catalogue_exposes_duplicate_definitions(self):
        a = self.parse_text("A", "A DEFINITIONS ::= BEGIN T ::= INTEGER END")
        b = self.parse_text("B", "B DEFINITIONS ::= BEGIN T ::= BOOLEAN END")
        providers, _ = catalogue([a, b])
        self.assertEqual(providers["T"], ["A", "B"])


if __name__ == "__main__":
    unittest.main()
