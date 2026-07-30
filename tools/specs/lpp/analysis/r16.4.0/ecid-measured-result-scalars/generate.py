#!/usr/bin/env python3
"""Generate deterministic scalar-only UPER fixture metadata (development only)."""
import hashlib
import json
from pathlib import Path

import asn1tools
from asn1tools.codecs.per import Encoder

ROOT = Path(__file__).resolve().parent
HARNESS = ROOT / "scalar-harness.asn1"
OUT = ROOT / "fixtures.json"


def encode(specification, name, value):
    encoded = bytes(specification.encode(name, value))
    encoder = Encoder()
    specification.types[name]._type.encode(value, encoder)
    return encoded, encoder.chunks_number_of_bits + encoder.number_of_bits


def record(specification, name, value):
    encoded, bits = encode(specification, name, value)
    decoded = specification.decode(name, encoded)
    again, again_bits = encode(specification, name, decoded)
    assert encoded == again and bits == again_bits
    return {
        "bit_length": bits,
        "hex": encoded.hex(),
        "round_trip": True,
        "sha256": hashlib.sha256(encoded).hexdigest(),
        "type": name,
        "value": value,
    }


def main():
    specification = asn1tools.compile_files([str(HARNESS)], codec="uper")
    fixtures = [
        record(specification, "PhysicalCellID", value)
        for value in (0, 1, 503)
    ] + [
        record(specification, "EUTRACarrierFrequencyRoot", value)
        for value in (0, 100, 65535)
    ] + [
        record(specification, "EUTRACarrierFrequencyV9a0", 65536)
    ] + [
        record(specification, "ScalarPair", value)
        for value in (
            {"physCellId": 1, "arfcnEUTRA": 100},
            {"physCellId": 503, "arfcnEUTRA": 65535},
        )
    ]
    document = {
        "compiler": "asn1tools",
        "compiler_version": asn1tools.__version__,
        "encoding_rule": "uper",
        "harness_sha256": hashlib.sha256(HARNESS.read_bytes()).hexdigest(),
        "fixtures": fixtures,
    }
    OUT.write_text(json.dumps(document, indent=2, sort_keys=True) + "\n")


if __name__ == "__main__":
    main()
