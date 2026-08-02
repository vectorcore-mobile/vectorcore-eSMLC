#!/usr/bin/env python3
"""Independent UPER fixtures for the bounded root-only OTDOA vertical slice.

Compiles the real, complete, corrected V2 normalized LPP-PDU-Definitions
module (not a hand-written harness) with the pinned asn1tools UPER codec, the
same pattern already used for the ECID capability/location phases. Selected
scope: RequestCapabilities/ProvideCapabilities otdoa-* branches,
RequestLocationInformation/ProvideLocationInformation otdoa-* branches, and
OTDOA-SignalMeasurementInformation/NeighbourMeasurementElement/
OTDOA-MeasQuality/OTDOA-Error at their R9 root only. Release 10/14/15
extension additions (PRS config, additional paths, motion measurements,
NB-IoT variants) and OTDOA assistance data (OTDOA-ProvideAssistanceData /
OTDOA-RequestAssistanceData / OTDOA-ReferenceCellInfo /
OTDOA-NeighbourCellInfoList) are out of scope for this phase.
"""
from __future__ import annotations

import hashlib
import json
from pathlib import Path

import asn1tools
from asn1tools.codecs.uper import Encoder

ROOT = Path(__file__).resolve().parents[6]
PDU = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/normalized/LPP-PDU-Definitions.asn1"
OUT = ROOT / "tools/specs/lpp/analysis/r16.4.0/otdoa"
FIX = ROOT / "tools/specs/lpp/fixtures/r16.4.0/otdoa"


def sha(p: Path) -> str:
    return hashlib.sha256(p.read_bytes()).hexdigest()


def dump(p: Path, value) -> None:
    p.parent.mkdir(parents=True, exist_ok=True)
    p.write_text(json.dumps(value, indent=2, sort_keys=True) + "\n")


def jsonable(v):
    if isinstance(v, tuple):
        return [jsonable(x) for x in v]
    if isinstance(v, bytes):
        return {"hex": v.hex()}
    if isinstance(v, dict):
        return {k: jsonable(x) for k, x in v.items()}
    if isinstance(v, list):
        return [jsonable(x) for x in v]
    return v


def bits(spec, typ, value):
    encoded = bytes(spec.encode(typ, value, check_constraints=True))
    enc = Encoder()
    spec.types[typ]._type.encode(value, enc)
    n = enc.chunks_number_of_bits + enc.number_of_bits
    if bytes(enc.as_bytearray()) != encoded:
        raise RuntimeError("public/instrumented encoder mismatch")
    return encoded, n


ECGI_A = {"mcc": [0, 0, 1], "mnc": [0, 1], "cellidentity": (bytes([0, 0, 0, 0]), 28)}
QUALITY = {"error-Resolution": (bytes([0x40]), 2), "error-Value": (bytes([0x00]), 5)}
QUALITY_WITH_SAMPLES = {"error-Resolution": (bytes([0x40]), 2), "error-Value": (bytes([0xA0]), 5), "error-NumSamples": (bytes([0x60]), 3)}


def main() -> None:
    spec = asn1tools.compile_files([str(PDU)], codec="uper")

    values = {
        "request-capabilities-r9-otdoa-empty": (
            "RequestCapabilities",
            {"criticalExtensions": ("c1", ("requestCapabilities-r9", {"otdoa-RequestCapabilities": {}}))},
            "requests the OTDOA capability group; OTDOA-RequestCapabilities has no root fields",
        ),
        "provide-capabilities-r9-otdoa": (
            "ProvideCapabilities",
            {"criticalExtensions": ("c1", ("provideCapabilities-r9", {"otdoa-ProvideCapabilities": {"otdoa-Mode": (bytes([0x80]), 1)}}))},
            "target device supports ue-assisted OTDOA only (bit 0)",
        ),
        "request-location-information-r9-otdoa-not-allowed": (
            "RequestLocationInformation",
            {"criticalExtensions": ("c1", ("requestLocationInformation-r9", {"otdoa-RequestLocationInformation": {"assistanceAvailability": False}}))},
            "server does not permit the target device to request further OTDOA assistance data",
        ),
        "provide-location-information-r9-otdoa-two-neighbours": (
            "ProvideLocationInformation",
            {"criticalExtensions": ("c1", ("provideLocationInformation-r9", {"otdoa-ProvideLocationInformation": {
                "otdoaSignalMeasurementInformation": {
                    "systemFrameNumber": (bytes([0, 0]), 10),
                    "physCellIdRef": 5,
                    "cellGlobalIdRef": ECGI_A,
                    "earfcnRef": 100,
                    "referenceQuality": QUALITY,
                    "neighbourMeasurementList": [
                        {"physCellIdNeighbour": 12, "earfcnNeighbour": 100, "rstd": 6356, "rstd-Quality": QUALITY},
                        {"physCellIdNeighbour": 300, "rstd": 2260, "rstd-Quality": QUALITY_WITH_SAMPLES},
                    ],
                },
            }}))},
            "reference cell plus two neighbour RSTD measurements; second neighbour omits earfcn and includes error-NumSamples",
        ),
        "provide-location-information-r9-otdoa-error": (
            "ProvideLocationInformation",
            {"criticalExtensions": ("c1", ("provideLocationInformation-r9", {"otdoa-ProvideLocationInformation": {
                "otdoa-Error": ("targetDeviceErrorCauses", {"cause": "assistance-data-missing"}),
            }}))},
            "target device reports it has no OTDOA assistance data",
        ),
    }

    negative = {
        "otdoa-neighbour-list-empty": (
            "ProvideLocationInformation",
            {"criticalExtensions": ("c1", ("provideLocationInformation-r9", {"otdoa-ProvideLocationInformation": {
                "otdoaSignalMeasurementInformation": {"systemFrameNumber": (bytes([0, 0]), 10), "physCellIdRef": 5, "neighbourMeasurementList": []},
            }}))},
        ),
        "otdoa-rstd-above-range": (
            "ProvideLocationInformation",
            {"criticalExtensions": ("c1", ("provideLocationInformation-r9", {"otdoa-ProvideLocationInformation": {
                "otdoaSignalMeasurementInformation": {"systemFrameNumber": (bytes([0, 0]), 10), "physCellIdRef": 5,
                                                        "neighbourMeasurementList": [{"physCellIdNeighbour": 1, "rstd": 12712, "rstd-Quality": QUALITY}]},
            }}))},
        ),
    }

    records = []
    for name, (typ, value, purpose) in values.items():
        data, n = bits(spec, typ, value)
        decoded = spec.decode(typ, data)
        again, n2 = bits(spec, typ, decoded)
        if data != again or n != n2 or jsonable(decoded) != jsonable(value):
            raise RuntimeError(name + " round trip mismatch")
        (FIX / f"{name}.uper").parent.mkdir(parents=True, exist_ok=True)
        (FIX / f"{name}.uper").write_bytes(data)
        (FIX / f"{name}.input.py").write_text("# Generated fixture source value; Python tuple denotes ASN.1 CHOICE.\nVALUE = " + repr(value) + "\n")
        records.append({"name": name, "top_level_type": typ, "purpose": purpose, "hex": data.hex(), "bit_length": n,
                         "unused_trailing_bits": len(data) * 8 - n, "decoded_value": jsonable(decoded)})

    neg_records = []
    for name, (typ, value) in negative.items():
        try:
            bits(spec, typ, value)
        except Exception as exc:
            neg_records.append({"name": name, "rejected": True, "error_type": type(exc).__name__, "error": str(exc)})
        else:
            neg_records.append({"name": name, "rejected": False, "error": "compiler accepted invalid test value"})

    manifest = {"specification": "3GPP TS 37.355 V16.4.0", "encoding_rule": "uper", "compiler": "asn1tools",
                "compiler_version": asn1tools.__version__, "module": str(PDU.relative_to(ROOT)), "module_sha256": sha(PDU),
                "scope": "OTDOA root-only: RequestCapabilities/ProvideCapabilities/RequestLocationInformation/ProvideLocationInformation otdoa-* branches",
                "fixtures": records, "negative_cases": neg_records}
    dump(OUT / "fixtures.json", manifest)
    dump(FIX / "manifest.json", manifest)
    print(f"generated {len(records)} valid fixtures and {len(neg_records)} negative cases")


if __name__ == "__main__":
    main()
