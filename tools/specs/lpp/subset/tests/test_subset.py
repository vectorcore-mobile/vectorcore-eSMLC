#!/usr/bin/env python3
import hashlib
import json
import sys
from pathlib import Path
import unittest

HERE = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(HERE))
import generate


class SubsetTests(unittest.TestCase):
    def test_strict_closure_is_exactly_pdu_assignments(self):
        pdu = generate.parse(generate.lex_file(generate.PDU))
        closure = json.loads((generate.ANALYSIS / "bounded-outer-closure.json").read_text())
        self.assertEqual(set(closure["strict_structural_closure"]["symbols"]), {d.name for d in pdu.definitions})

    def test_canonical_input_hashes(self):
        for path, expected in generate.EXPECTED.items():
            self.assertEqual(hashlib.sha256(path.read_bytes()).hexdigest(), expected)


if __name__ == "__main__":
    unittest.main()
