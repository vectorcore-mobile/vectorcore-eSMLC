#!/usr/bin/env python3
"""Generate deterministic analysis-only UPER fixtures for MeasuredResultsElement."""
import hashlib
import json
from pathlib import Path

import asn1tools
from asn1tools.codecs.per import Encoder

ROOT = Path(__file__).resolve().parent
HARNESS = ROOT / "measured-results-element-harness.asn1"
OUT = ROOT / "fixtures.json"


def bits(specification, value):
    encoder = Encoder()
    specification.types["MeasuredResultsElement"]._type.encode(value, encoder)
    return encoder.chunks_number_of_bits + encoder.number_of_bits


def record(name, value, annotation):
    encoded = SPEC.encode("MeasuredResultsElement", value)
    decoded = SPEC.decode("MeasuredResultsElement", encoded)
    again = SPEC.encode("MeasuredResultsElement", decoded)
    assert encoded == again
    return {
        "annotation": annotation,
        "bit_length": bits(SPEC, value),
        "hex": encoded.hex(),
        "name": name,
        "round_trip": True,
        "sha256": hashlib.sha256(encoded).hexdigest(),
        "value": canonical(value),
    }


def canonical(value):
    """Represent compiler BIT STRING tuples as deterministic JSON objects."""
    if isinstance(value, bytes):
        return {"hex": value.hex()}
    if isinstance(value, tuple) and len(value) == 2 and isinstance(value[0], bytes):
        return {"bit_length": value[1], "hex": value[0].hex()}
    if isinstance(value, tuple):
        return [canonical(item) for item in value]
    if isinstance(value, list):
        return [canonical(item) for item in value]
    if isinstance(value, dict):
        return {key: canonical(item) for key, item in value.items()}
    return value


def base():
    return {"physCellId": 1, "arfcnEUTRA": 100}


def cell(mcc, mnc, cell_bits):
    return {
        "plmn-Identity": {"mcc": mcc, "mnc": mnc},
        "cellIdentity": ("eutra", (cell_bits, 28)),
    }


SPEC = asn1tools.compile_files([str(HARNESS)], codec="uper")
FIXTURES = []
FIXTURES.append(record("mandatory-only", base(), "extension=0; root optionals=00000"))
for field, value in (
    ("cellGlobalId", cell([0, 0, 1], [0, 1], b"\x00\x00\x00\x00")),
    ("systemFrameNumber", (b"\xaa\x80", 10)),
    ("rsrp-Result", 49),
    ("rsrq-Result", 17),
    ("ue-RxTxTimeDiff", 2048),
):
    item = base(); item[field] = value
    FIXTURES.append(record("optional-" + field, item, "one root optional present"))
for field, values in (
    ("rsrp-Result", (0, 49, 97)),
    ("rsrq-Result", (0, 17, 34)),
    ("ue-RxTxTimeDiff", (0, 2048, 4095)),
):
    for value in values:
        item = base(); item[field] = value
        FIXTURES.append(record(f"{field}-{value}", item, "scalar boundary fixture"))
for name, value in (("sfn-all-zero", (b"\x00\x00", 10)), ("sfn-all-one", (b"\xff\xc0", 10))):
    item = base(); item["systemFrameNumber"] = value
    FIXTURES.append(record(name, item, "fixed 10-bit BIT STRING"))
for name, value in (
    ("cell-two-digit-mnc", cell([0, 0, 1], [0, 1], b"\x00\x00\x00\x00")),
    ("cell-three-digit-mnc", cell([3, 1, 0], [2, 6, 0], b"\xff\xff\xff\xf0")),
    ("cell-mcc-zero", cell([2, 0, 8], [0, 1], b"\x12\x34\x56\x70")),
):
    item = base(); item["cellGlobalId"] = value
    FIXTURES.append(record(name, item, "E-UTRA CellGlobalId fixture"))
all_fields = base()
all_fields.update({
    "cellGlobalId": cell([0, 0, 1], [0, 1], b"\x12\x34\x56\x70"),
    "systemFrameNumber": (b"\xaa\x80", 10),
    "rsrp-Result": 97,
    "rsrq-Result": 34,
    "ue-RxTxTimeDiff": 4095,
})
FIXTURES.append(record("all-root-optionals", all_fields, "extension=0; root optionals=11111"))
extension = base(); extension["arfcnEUTRA-v9a0"] = 65536
FIXTURES.append(record("extension-arfcn-v9a0", extension, "extension addition 0 present"))

OUT.write_text(json.dumps({
    "compiler": "asn1tools",
    "compiler_version": asn1tools.__version__,
    "encoding_rule": "uper",
    "harness_sha256": hashlib.sha256(HARNESS.read_bytes()).hexdigest(),
    "fixtures": FIXTURES,
}, indent=2, sort_keys=True) + "\n")
