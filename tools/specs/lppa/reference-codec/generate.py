#!/usr/bin/env python3
"""Independent bounded APER encoder for TS 36.455 (LPPa) E-CID measurement fixtures."""
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
OUT = ROOT / "tools/specs/lppa/fixtures/r16.0.0"


class Bits:
    def __init__(self): self.bits = []
    def put(self, value, width):
        self.bits.extend((value >> i) & 1 for i in range(width - 1, -1, -1))
    def align(self):
        self.bits.extend([0] * ((-len(self.bits)) % 8))
    def octets(self, value):
        self.align()
        for b in value: self.put(b, 8)
    def bytes(self):
        self.align()
        return bytes(sum(self.bits[i + j] << (7 - j) for j in range(8)) for i in range(0, len(self.bits), 8))


def width(span): return (span - 1).bit_length()
def octets_for_width(bit_width): return (bit_width + 7) // 8
def min_octets(off):
    n = 1
    while off >> (8 * n): n += 1
    return n


# X.691 10.5.7 aligned constrained whole number; must stay in lockstep with
# internal/aper.PutConstrained (see tools/specs/lcsap/reference-codec/generate.py
# for the regime derivation and the historical bug this guards against).
def constrained(w, value, lo, hi):
    if not lo <= value <= hi: raise ValueError("constraint")
    span = hi - lo + 1
    bit_width = width(span)
    if bit_width == 0: return
    off = value - lo
    if span <= 255:
        w.put(off, bit_width)
        return
    if span <= 65536:
        w.align()
        w.put(off, octets_for_width(bit_width) * 8)
        return
    max_octets = octets_for_width(bit_width)
    need = min_octets(off)
    constrained(w, need, 1, max_octets)
    w.align()
    w.put(off, need * 8)


# Mirrors internal/aper.PutFixedBitString: aligned only when size>16.
def fixed_bits(w, value, size):
    if size > 16: w.align()
    w.put(value, size)


def ext_bit(w, v=0): constrained(w, v, 0, 1)
def index(w, v, n): constrained(w, v, 0, n - 1)  # CHOICE/ENUMERATED root index over n root values
def criticality(w, v): constrained(w, v, 0, 2)   # non-extensible Criticality ENUMERATED{reject,ignore,notify}


def open_type(w, value):
    w.align(); n = len(value)
    if n >= 16384: raise ValueError("fixture open type too large")
    w.octets(bytes([n]) if n < 128 else bytes([0x80 | (n >> 8), n & 0xff]))
    w.octets(value)


# --- leaf IE encoders (also independently cross-checked in compile.py) ---

def cause(branch, value):
    maxima = [2, 6, 0]  # CauseRadioNetwork(3 root), CauseProtocol(7 root), CauseMisc(1 root)
    w = Bits()
    ext_bit(w)
    index(w, branch, 3)
    ext_bit(w)
    constrained(w, value, 0, maxima[branch])
    return w.bytes()


def cell_portion_id(value):
    w = Bits(); ext_bit(w); constrained(w, value, 0, 255); return w.bytes()


def measurement_id(value):
    w = Bits(); ext_bit(w); constrained(w, value, 1, 15); return w.bytes()


def report_characteristics(value):  # 0=onDemand, 1=periodic
    w = Bits(); ext_bit(w); index(w, value, 2); return w.bytes()


def measurement_periodicity(value):  # 0..12 (ms120..min60)
    w = Bits(); ext_bit(w); index(w, value, 13); return w.bytes()


def measurement_quantities_value(value):  # 0..5 (cell-ID..rSRQ)
    w = Bits(); ext_bit(w); index(w, value, 6); return w.bytes()


def plmn_identity(b):
    w = Bits(); w.octets(b); return w.bytes()


def eutran_cell_identifier(value):
    w = Bits(); fixed_bits(w, value, 28); return w.bytes()


def tac(b):
    w = Bits(); w.octets(b); return w.bytes()


def ecgi_inline(w, plmn, cell_id):
    ext_bit(w)      # ECGI SEQUENCE extension bit
    ext_bit(w, 0)   # iE-Extensions optional-presence bit (absent)
    w.octets(plmn)
    fixed_bits(w, cell_id, 28)


def ecgi(plmn, cell_id):
    w = Bits(); ecgi_inline(w, plmn, cell_id); return w.bytes()


def access_point_position_inline(w, lat_deg, lon_deg, alt_depth, altitude,
                                  semi_major, semi_minor, orientation, uncertainty_altitude, confidence):
    # lat_deg/lon_deg are decimal degrees, scaled with the same formula as
    # internal/lppa.encodeAccessPointPosition (and as TS 23.032 Geographical-
    # Coordinates) so Go-side test fixtures can be expressed in degrees too.
    latitude = round(abs(lat_deg) * ((1 << 23) - 1) / 90)
    longitude = round(lon_deg * ((1 << 23) - 1) / 180)
    ext_bit(w)  # E-UTRANAccessPointPosition SEQUENCE extension bit; no optional root fields
    constrained(w, int(lat_deg < 0), 0, 1)
    constrained(w, latitude, 0, 8388607)
    constrained(w, longitude, -8388608, 8388607)
    constrained(w, int(alt_depth), 0, 1)
    constrained(w, altitude, 0, 32767)
    constrained(w, semi_major, 0, 127)
    constrained(w, semi_minor, 0, 127)
    constrained(w, orientation, 0, 179)
    constrained(w, uncertainty_altitude, 0, 127)
    constrained(w, confidence, 0, 100)


def e_utran_access_point_position(lat_deg, lon_deg, alt_depth, altitude,
                                   semi_major, semi_minor, orientation, uncertainty_altitude, confidence):
    w = Bits()
    access_point_position_inline(w, lat_deg, lon_deg, alt_depth, altitude,
                                  semi_major, semi_minor, orientation, uncertainty_altitude, confidence)
    return w.bytes()


def measurement_quantities_item(value_idx):
    w = Bits()
    ext_bit(w)      # MeasurementQuantities-Item SEQUENCE extension bit
    ext_bit(w, 0)   # iE-Extensions optional-presence bit (absent)
    ext_bit(w)      # MeasurementQuantitiesValue's own extension bit
    index(w, value_idx, 6)
    return w.bytes()


ID_MEASUREMENT_QUANTITIES_ITEM = 11


def measurement_quantities(items):
    w = Bits()
    constrained(w, len(items), 1, 63)  # non-extensible SIZE(1..maxNoMeas)
    for value_idx in items:
        constrained(w, ID_MEASUREMENT_QUANTITIES_ITEM, 0, 65535)
        criticality(w, CRIT_REJECT)
        open_type(w, measurement_quantities_item(value_idx))
    return w.bytes()


def e_cid_measurement_result(plmn, cell_id, tac_bytes, access_point_position=None):
    w = Bits()
    ext_bit(w)                                                  # E-CID-MeasurementResult SEQUENCE extension bit
    ext_bit(w, 1 if access_point_position is not None else 0)   # e-UTRANAccessPointPosition optional-presence
    ext_bit(w, 0)                                                # measuredResults optional-presence (out of bounded scope)
    ecgi_inline(w, plmn, cell_id)
    w.octets(tac_bytes)
    if access_point_position is not None:
        access_point_position_inline(w, *access_point_position)
    return w.bytes()


# --- envelope + message assembly ---

CAT_INITIATING, CAT_SUCCESSFUL, CAT_UNSUCCESSFUL = 0, 1, 2
CRIT_REJECT, CRIT_IGNORE, CRIT_NOTIFY = 0, 1, 2

PROC_INITIATION = 2
PROC_FAILURE_INDICATION = 3
PROC_REPORT = 4
PROC_TERMINATION = 5

ID_CAUSE = 0
ID_E_SMLC_UE_MEASUREMENT_ID = 2
ID_REPORT_CHARACTERISTICS = 3
ID_MEASUREMENT_PERIODICITY = 4
ID_MEASUREMENT_QUANTITIES = 5
ID_ENB_UE_MEASUREMENT_ID = 6
ID_E_CID_MEASUREMENT_RESULT = 7
ID_CELL_PORTION_ID = 14


def message_body(ies):
    # ies: list of (id, criticality, value_bytes)
    body = Bits()
    ext_bit(body)  # message-level SEQUENCE extension bit; single mandatory protocolIEs field, no optionals
    constrained(body, len(ies), 0, 65535)
    for ident, crit, value in ies:
        constrained(body, ident, 0, 65535)
        criticality(body, crit)
        open_type(body, value)
    return body.bytes()


def pdu(category, procedure_code, crit, transaction_id, body_bytes):
    w = Bits()
    ext_bit(w)                          # LPPA-PDU CHOICE extension bit
    index(w, category, 3)               # initiatingMessage/successfulOutcome/unsuccessfulOutcome
    constrained(w, procedure_code, 0, 255)
    criticality(w, crit)
    constrained(w, transaction_id, 0, 32767)
    open_type(w, body_bytes)
    return w.bytes()


def initiation_request(txn, mid, report_char, periodicity, quantities):
    ies = [
        (ID_E_SMLC_UE_MEASUREMENT_ID, CRIT_REJECT, measurement_id(mid)),
        (ID_REPORT_CHARACTERISTICS, CRIT_REJECT, report_characteristics(report_char)),
    ]
    if periodicity is not None:
        ies.append((ID_MEASUREMENT_PERIODICITY, CRIT_REJECT, measurement_periodicity(periodicity)))
    ies.append((ID_MEASUREMENT_QUANTITIES, CRIT_REJECT, measurement_quantities(quantities)))
    return pdu(CAT_INITIATING, PROC_INITIATION, CRIT_REJECT, txn, message_body(ies))


def initiation_response(txn, e_smlc_id, enb_id, result=None, cell_portion=None):
    ies = [
        (ID_E_SMLC_UE_MEASUREMENT_ID, CRIT_REJECT, measurement_id(e_smlc_id)),
        (ID_ENB_UE_MEASUREMENT_ID, CRIT_REJECT, measurement_id(enb_id)),
    ]
    if result is not None:
        ies.append((ID_E_CID_MEASUREMENT_RESULT, CRIT_IGNORE, result))
    if cell_portion is not None:
        ies.append((ID_CELL_PORTION_ID, CRIT_IGNORE, cell_portion_id(cell_portion)))
    return pdu(CAT_SUCCESSFUL, PROC_INITIATION, CRIT_REJECT, txn, message_body(ies))


def initiation_failure(txn, e_smlc_id, branch, cause_value):
    ies = [
        (ID_E_SMLC_UE_MEASUREMENT_ID, CRIT_REJECT, measurement_id(e_smlc_id)),
        (ID_CAUSE, CRIT_IGNORE, cause(branch, cause_value)),
    ]
    return pdu(CAT_UNSUCCESSFUL, PROC_INITIATION, CRIT_REJECT, txn, message_body(ies))


def failure_indication(txn, e_smlc_id, enb_id, branch, cause_value):
    ies = [
        (ID_E_SMLC_UE_MEASUREMENT_ID, CRIT_REJECT, measurement_id(e_smlc_id)),
        (ID_ENB_UE_MEASUREMENT_ID, CRIT_REJECT, measurement_id(enb_id)),
        (ID_CAUSE, CRIT_IGNORE, cause(branch, cause_value)),
    ]
    return pdu(CAT_INITIATING, PROC_FAILURE_INDICATION, CRIT_IGNORE, txn, message_body(ies))


def report(txn, e_smlc_id, enb_id, result, cell_portion=None):
    ies = [
        (ID_E_SMLC_UE_MEASUREMENT_ID, CRIT_REJECT, measurement_id(e_smlc_id)),
        (ID_ENB_UE_MEASUREMENT_ID, CRIT_REJECT, measurement_id(enb_id)),
        (ID_E_CID_MEASUREMENT_RESULT, CRIT_IGNORE, result),
    ]
    if cell_portion is not None:
        ies.append((ID_CELL_PORTION_ID, CRIT_IGNORE, cell_portion_id(cell_portion)))
    return pdu(CAT_INITIATING, PROC_REPORT, CRIT_IGNORE, txn, message_body(ies))


def termination(txn, e_smlc_id, enb_id):
    ies = [
        (ID_E_SMLC_UE_MEASUREMENT_ID, CRIT_REJECT, measurement_id(e_smlc_id)),
        (ID_ENB_UE_MEASUREMENT_ID, CRIT_REJECT, measurement_id(enb_id)),
    ]
    return pdu(CAT_INITIATING, PROC_TERMINATION, CRIT_REJECT, txn, message_body(ies))


PLMN = bytes([0x00, 0xf1, 0x10])
CELL_ID = 0x0abcdef
TAC_BYTES = bytes([0x10, 0x01])
ACCESS_POINT = (38, -90, False, 120, 10, 5, 45, 3, 68)  # lat=38N, lon=90W: same magic numbers as the LCS-AP fixtures

RESULT_MINIMAL = e_cid_measurement_result(PLMN, CELL_ID, TAC_BYTES)
RESULT_WITH_POSITION = e_cid_measurement_result(PLMN, CELL_ID, TAC_BYTES, ACCESS_POINT)

FIXTURES = {
    "initiation-request-ondemand": initiation_request(1, 3, 0, None, [0]),
    "initiation-request-periodic": initiation_request(2, 5, 1, 4, [0]),
    "initiation-response-minimal": initiation_response(1, 3, 9),
    "initiation-response-full": initiation_response(2, 5, 11, RESULT_WITH_POSITION, cell_portion=12),
    "initiation-failure-misc-unspecified": initiation_failure(1, 3, 2, 0),
    "initiation-failure-radio-network-unspecified": initiation_failure(3, 7, 0, 0),
    "failure-indication-protocol-semantic-error": failure_indication(4, 9, 11, 1, 4),
    "report-with-cell-portion": report(5, 9, 11, RESULT_MINIMAL, cell_portion=200),
    "report-without-cell-portion": report(6, 3, 9, RESULT_WITH_POSITION),
    "termination": termination(1, 3, 9),
}


def main():
    OUT.mkdir(parents=True, exist_ok=True)
    records = []
    for name, wire in FIXTURES.items():
        path = OUT / (name + ".aper")
        path.write_bytes(wire)
        records.append({"name": name, "file": path.name, "sha256": hashlib.sha256(wire).hexdigest(), "hex": wire.hex()})
    manifest = {
        "specification": "3GPP TS 36.455 (ATIS reprint)",
        "version": "16.0.0",
        "encoding": "aligned PER",
        "compiler": {
            "name": "asn1tools", "version": "0.167.0",
            "role": "pinned leaf declaration compiler; complete PDU is independently composed because "
                    "object-set OPEN TYPE dispatch is unsupported",
        },
        "fixtures": records,
    }
    (OUT / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")


if __name__ == "__main__": main()
