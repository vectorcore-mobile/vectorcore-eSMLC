#!/usr/bin/env python3
"""Independent UPER fixtures for the bounded root-only A-GNSS UE-based slice.

Compiles the real, complete, corrected V2 normalized LPP-PDU-Definitions
module (not a hand-written harness), mirroring
tools/specs/lpp/analysis/r16.4.0/otdoa/generate.py. Selected scope, GPS and
UE-based (MS-based) mode only:
 - CommonIEsRequestLocationInformation.locationInformationType (the request
   signal for "please report your own computed position", not raw
   measurements)
 - A-GNSS-RequestLocationInformation.gnss-PositioningInstructions
 - A-GNSS-RequestCapabilities (all three root booleans)
 - A-GNSS-ProvideCapabilities.gnss-SupportList (needed to gate on the UE
   actually supporting GPS + ue-based mode)
 - CommonIEsProvideLocationInformation.locationEstimate, restricted to the
   ellipsoidPointWithUncertaintyCircle LocationCoordinates branch (the same
   shape TS 29.171's Point-With-Uncertainty already uses on the SLs side)
 - A-GNSS-ProvideLocationInformation.gnss-LocationInformation (supplementary
   metadata: reference time + which GNSS systems contributed), optional
 - A-GNSS-Error (root only)

GNSS-SignalMeasurementInformation (raw pseudorange/carrier measurements, the
MS-assisted path), assistance data, and every non-GPS GNSS-ID are out of
scope for this phase.
"""
from __future__ import annotations

import hashlib
import json
from pathlib import Path

import asn1tools
from asn1tools.codecs.uper import Encoder

ROOT = Path(__file__).resolve().parents[6]
PDU = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/normalized/LPP-PDU-Definitions.asn1"
OUT = ROOT / "tools/specs/lpp/analysis/r16.4.0/a-gnss"
FIX = ROOT / "tools/specs/lpp/fixtures/r16.4.0/a-gnss"


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


COORDS_A = {"latitudeSign": "north", "degreesLatitude": round(38 * ((1 << 23) - 1) / 90), "degreesLongitude": round(-90 * ((1 << 23) - 1) / 180)}
GPS_SUPPORT_UE_BASED = {
    "gnss-ID": {"gnss-id": "gps"},
    "agnss-Modes": {"posModes": (bytes([0x40]), 2)},  # bit1 (ue-based) set
    "gnss-Signals": {"gnss-SignalIDs": (bytes([0x80]), 8)},
    "adr-Support": False,
    "velocityMeasurementSupport": False,
}


def main() -> None:
    spec = asn1tools.compile_files([str(PDU)], codec="uper")

    values = {
        "request-capabilities-r9-agnss": (
            "RequestCapabilities",
            {"criticalExtensions": ("c1", ("requestCapabilities-r9", {"a-gnss-RequestCapabilities": {
                "gnss-SupportListReq": True, "assistanceDataSupportListReq": False, "locationVelocityTypesReq": False,
            }}))},
            "requests the A-GNSS capability group, asking only for the GNSS support list",
        ),
        "provide-capabilities-r9-agnss-gps-ue-based": (
            "ProvideCapabilities",
            {"criticalExtensions": ("c1", ("provideCapabilities-r9", {"a-gnss-ProvideCapabilities": {
                "gnss-SupportList": [GPS_SUPPORT_UE_BASED],
            }}))},
            "target device supports GPS in ue-based mode only",
        ),
        "request-location-information-r9-agnss-ue-based": (
            "RequestLocationInformation",
            {"criticalExtensions": ("c1", ("requestLocationInformation-r9", {
                "commonIEsRequestLocationInformation": {"locationInformationType": "locationEstimateRequired"},
                "a-gnss-RequestLocationInformation": {"gnss-PositioningInstructions": {
                    "gnss-Methods": {"gnss-ids": (bytes([0x80]), 1)},
                    "fineTimeAssistanceMeasReq": False, "adrMeasReq": False, "multiFreqMeasReq": False, "assistanceAvailability": False,
                }},
            }))},
            "server requests a UE-computed GPS position estimate, no assistance data offered",
        ),
        "provide-location-information-r9-agnss-estimate": (
            "ProvideLocationInformation",
            {"criticalExtensions": ("c1", ("provideLocationInformation-r9", {
                "commonIEsProvideLocationInformation": {"locationEstimate": ("ellipsoidPointWithUncertaintyCircle", {**COORDS_A, "uncertainty": 30})},
                "a-gnss-ProvideLocationInformation": {"gnss-LocationInformation": {
                    "measurementReferenceTime": {"gnss-TOD-msec": 1000, "gnss-TimeID": {"gnss-id": "gps"}},
                    "agnss-List": {"gnss-ids": (bytes([0x80]), 1)},
                }},
            }))},
            "target device reports its own computed GPS position via the common location estimate",
        ),
        "provide-location-information-r9-agnss-error": (
            "ProvideLocationInformation",
            {"criticalExtensions": ("c1", ("provideLocationInformation-r9", {
                "a-gnss-ProvideLocationInformation": {"gnss-Error": ("targetDeviceErrorCauses", {"cause": "assistanceDataMissing"})},
            }))},
            "target device reports it is missing GNSS assistance data",
        ),
    }

    negative = {
        "agnss-common-location-estimate-unsupported-shape": (
            "ProvideLocationInformation",
            {"criticalExtensions": ("c1", ("provideLocationInformation-r9", {
                "commonIEsProvideLocationInformation": {"locationEstimate": ("ellipsoidPoint", COORDS_A)},
            }))},
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
        data, n = bits(spec, typ, value)
        neg_records.append({"name": name, "hex": data.hex(), "bit_length": n, "note": "valid per spec; rejected by this implementation's bounded shape scope, not by the compiler"})

    manifest = {"specification": "3GPP TS 37.355 V16.4.0", "encoding_rule": "uper", "compiler": "asn1tools",
                "compiler_version": asn1tools.__version__, "module": str(PDU.relative_to(ROOT)), "module_sha256": sha(PDU),
                "scope": "A-GNSS root-only, GPS + UE-based (MS-based) mode: common location-estimate request/provide, A-GNSS capability/instructions/error",
                "fixtures": records, "negative_cases": neg_records}
    dump(OUT / "fixtures.json", manifest)
    dump(FIX / "manifest.json", manifest)
    print(f"generated {len(records)} valid fixtures and {len(neg_records)} bounded-scope negative cases")


if __name__ == "__main__":
    main()
