#!/usr/bin/env python3
"""Generate deterministic, development-only ECID provide-location fixtures."""
import hashlib
import json
from pathlib import Path

import asn1tools
from asn1tools.codecs.uper import Encoder

ROOT = Path(__file__).resolve().parents[6]
PDU = ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/normalized/LPP-PDU-Definitions.asn1"
OUT = Path(__file__).resolve().parent / "fixtures.json"
SPEC = asn1tools.compile_files([str(PDU)], codec="uper")


def bit_length(value):
    encoder = Encoder()
    SPEC.types["ProvideLocationInformation"]._type.encode(value, encoder)
    return encoder.chunks_number_of_bits + encoder.number_of_bits


def message_bit_length(value):
    encoder = Encoder()
    SPEC.types["LPP-Message"]._type.encode(value, encoder)
    return encoder.chunks_number_of_bits + encoder.number_of_bits


def element(pci, arfcn, **optional):
    out = {"physCellId": pci, "arfcnEUTRA": arfcn}
    out.update(optional)
    return out


def global_id(mcc, mnc, alternative, value, bytes_required):
    width = 28 if alternative == "eutra" else 32
    # asn1tools represents BIT STRING bytes left-aligned in the final octet.
    encoded_value = value << (32 - width)
    return {
        "plmn-Identity": {"mcc": mcc, "mnc": mnc},
        "cellIdentity": (alternative, (encoded_value.to_bytes(bytes_required, "big"), width)),
    }


def provide(primary, values):
    signal = {"measuredResultsList": values}
    if primary is not None:
        signal["primaryCellMeasuredResults"] = primary
    return {
        "criticalExtensions": (
            "c1",
            ("provideLocationInformation-r9", {
                "ecid-ProvideLocationInformation": {
                    "ecid-SignalMeasurementInformation": signal
                }
            }),
        )
    }


vectors = {
    "one-rsrp": provide(None, [element(1, 100, **{"rsrp-Result": 30})]),
    "list-two-mixed": provide(
        element(0, 0),
        [element(1, 100), element(503, 65535, **{"rsrq-Result": 34, "ue-RxTxTimeDiff": 4095})],
    ),
    "all-optionals-eutra": provide(None, [element(
        1, 100,
        cellGlobalId=global_id([0, 0, 1], [0, 1], "eutra", 0x1234567, 4),
        systemFrameNumber=(bytes.fromhex("aa80"), 10),
        **{"rsrp-Result": 97, "rsrq-Result": 34, "ue-RxTxTimeDiff": 4095},
    )]),
    "utravariant": provide(None, [element(
        503, 65535,
        cellGlobalId=global_id([3, 1, 0], [2, 6, 0], "utra", 0x12345678, 4),
    )]),
}

records = []
for name, value in vectors.items():
    encoded = SPEC.encode("ProvideLocationInformation", value)
    decoded = SPEC.decode("ProvideLocationInformation", encoded)
    assert encoded == SPEC.encode("ProvideLocationInformation", decoded)
    bits = bit_length(value)
    records.append({
        "name": name,
        "hex": encoded.hex(),
        "bit_length": bits,
        "unused_trailing_bits": len(encoded) * 8 - bits,
        "sha256": hashlib.sha256(encoded).hexdigest(),
        "round_trip": True,
    })

message_value = {
    "transactionID": {"initiator": "targetDevice", "transactionNumber": 7},
    "endTransaction": True,
    "lpp-MessageBody": ("c1", ("provideLocationInformation", vectors["all-optionals-eutra"])),
}
message_encoded = SPEC.encode("LPP-Message", message_value)
assert message_encoded == SPEC.encode("LPP-Message", SPEC.decode("LPP-Message", message_encoded))
records.append({
    "name": "message-all-optionals-eutra",
    "hex": message_encoded.hex(),
    "bit_length": message_bit_length(message_value),
    "unused_trailing_bits": len(message_encoded) * 8 - message_bit_length(message_value),
    "sha256": hashlib.sha256(message_encoded).hexdigest(),
    "round_trip": True,
})

OUT.write_text(json.dumps({
    "specification": "3GPP TS 37.355 V16.4.0",
    "encoding_rule": "uper",
    "compiler": "asn1tools",
    "compiler_version": asn1tools.__version__,
    "module_sha256": hashlib.sha256(PDU.read_bytes()).hexdigest(),
    "fixtures": records,
}, indent=2, sort_keys=True) + "\n", encoding="utf-8")
