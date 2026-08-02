#!/usr/bin/env python3
"""Compile and cross-check the independently generated TS 36.455 (LPPa) APER leaves."""
import importlib.metadata
import runpy
from pathlib import Path

if importlib.metadata.version("asn1tools") != "0.167.0":
    raise SystemExit("expected pinned asn1tools 0.167.0")
import asn1tools

ASN1 = """
LPPAFixture DEFINITIONS AUTOMATIC TAGS ::= BEGIN
Cause ::= CHOICE {
 radioNetwork ENUMERATED { unspecified, requested-item-not-supported, requested-item-temporarily-not-available, ... },
 protocol ENUMERATED { transfer-syntax-error, abstract-syntax-error-reject, abstract-syntax-error-ignore-and-notify, message-not-compatible-with-receiver-state, semantic-error, unspecified, abstract-syntax-error-falsely-constructed-message, ... },
 misc ENUMERATED { unspecified, ... },
 ... }
Cell-Portion-ID ::= INTEGER (0..255, ..., 256..4095)
Measurement-ID ::= INTEGER (1..15, ...)
ReportCharacteristics ::= ENUMERATED { onDemand, periodic, ... }
MeasurementPeriodicity ::= ENUMERATED { ms120, ms240, ms480, ms640, ms1024, ms2048, ms5120, ms10240, min1, min6, min12, min30, min60, ... }
MeasurementQuantitiesValue ::= ENUMERATED { cell-ID, angleOfArrival, timingAdvanceType1, timingAdvanceType2, rSRP, rSRQ, ... }
PLMN-Identity ::= OCTET STRING (SIZE(3))
EUTRANCellIdentifier ::= BIT STRING (SIZE (28))
ECGI ::= SEQUENCE {
 pLMN-Identity PLMN-Identity,
 eUTRANcellIdentifier EUTRANCellIdentifier,
 iE-Extensions OCTET STRING OPTIONAL, ... }
TAC ::= OCTET STRING (SIZE(2))
E-UTRANAccessPointPosition ::= SEQUENCE {
 latitudeSign ENUMERATED {north, south},
 latitude INTEGER (0..8388607),
 longitude INTEGER (-8388608..8388607),
 directionOfAltitude ENUMERATED {height, depth},
 altitude INTEGER (0..32767),
 uncertaintySemi-major INTEGER (0..127),
 uncertaintySemi-minor INTEGER (0..127),
 orientationOfMajorAxis INTEGER (0..179),
 uncertaintyAltitude INTEGER (0..127),
 confidence INTEGER (0..100),
 ... }
END
"""

def main():
    codec = asn1tools.compile_string(ASN1, "per")
    generated = runpy.run_path(str(Path(__file__).with_name("generate.py")))
    checks = [
        ("Cause", ("misc", "unspecified"), generated["cause"](2, 0)),
        ("Cause", ("radioNetwork", "unspecified"), generated["cause"](0, 0)),
        ("Cause", ("protocol", "semantic-error"), generated["cause"](1, 4)),
        ("Cell-Portion-ID", 12, generated["cell_portion_id"](12)),
        ("Measurement-ID", 7, generated["measurement_id"](7)),
        ("ReportCharacteristics", "onDemand", generated["report_characteristics"](0)),
        ("ReportCharacteristics", "periodic", generated["report_characteristics"](1)),
        ("MeasurementPeriodicity", "ms1024", generated["measurement_periodicity"](4)),
        ("MeasurementQuantitiesValue", "cell-ID", generated["measurement_quantities_value"](0)),
        ("PLMN-Identity", bytes([0x00, 0xf1, 0x10]), generated["plmn_identity"](bytes([0x00, 0xf1, 0x10]))),
        ("EUTRANCellIdentifier", (bytes([0x0a, 0xbc, 0xde, 0xf0]), 28), generated["eutran_cell_identifier"](0x0abcdef)),
        ("TAC", bytes([0x10, 0x01]), generated["tac"](bytes([0x10, 0x01]))),
        ("ECGI", {"pLMN-Identity": bytes([0x00, 0xf1, 0x10]), "eUTRANcellIdentifier": (bytes([0x0a, 0xbc, 0xde, 0xf0]), 28)}, generated["ecgi"](bytes([0x00, 0xf1, 0x10]), 0x0abcdef)),
        ("E-UTRANAccessPointPosition",
         {"latitudeSign": "north", "latitude": round(38 * ((1 << 23) - 1) / 90), "longitude": round(-90 * ((1 << 23) - 1) / 180),
          "directionOfAltitude": "height", "altitude": 120, "uncertaintySemi-major": 10, "uncertaintySemi-minor": 5,
          "orientationOfMajorAxis": 45, "uncertaintyAltitude": 3, "confidence": 68},
         generated["e_utran_access_point_position"](38, -90, False, 120, 10, 5, 45, 3, 68)),
    ]
    for type_name, value, expected in checks:
        actual = codec.encode(type_name, value)
        if actual != expected:
            raise SystemExit(f"APER compiler mismatch {type_name}: {actual.hex()} != {expected.hex()}")
    print("asn1tools 0.167.0 leaf APER checks passed")

if __name__ == "__main__": main()
