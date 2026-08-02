#!/usr/bin/env python3
"""Independent bounded APER encoder for TS 29.171 Location Response fixtures."""
import hashlib
import json
import math
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
OUT = ROOT / "tools/specs/lcsap/fixtures/r16.4.0"

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
        return bytes(sum(self.bits[i+j] << (7-j) for j in range(8)) for i in range(0, len(self.bits), 8))

def width(span): return (span - 1).bit_length()
def octets_for_width(bit_width): return (bit_width + 7) // 8
def min_octets(off):
    n = 1
    while off >> (8 * n): n += 1
    return n

# X.691 10.5.7 aligned constrained whole number: span<=255 is a non-aligned
# minimum-width bit-field; 255<span<=65536 is octet-aligned at a *fixed*
# whole-octet width (not the raw bit width); span>65536 is octet-aligned with
# a non-aligned length-determinant (1..maxOctets) selecting the value's own
# minimum octet count. Must stay in lockstep with internal/aper.PutConstrained.
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
def criticality(w, value): w.put(value, 2)
def open_type(w, value):
    w.align(); n = len(value)
    if n >= 16384: raise ValueError("fixture open type too large")
    w.octets(bytes([n]) if n < 128 else bytes([0x80 | (n >> 8), n & 0xff]))
    w.octets(value)

def location_estimate(lat=38, lon=-90, uncertainty=4):
    latv = round(abs(lat) * ((1 << 23) - 1) / 90)
    lonv = round(lon * ((1 << 23) - 1) / 180)
    w = Bits()
    w.put(0, 1); w.put(1, 3)               # Geographical-Area root / point-With-Uncertainty
    w.put(0, 2); w.put(0, 2)               # two extensible SEQUENCE extension/presence bits
    constrained(w, int(lat < 0), 0, 1)
    constrained(w, latv, 0, (1 << 23) - 1)
    constrained(w, lonv, -(1 << 23), (1 << 23) - 1)
    constrained(w, uncertainty, 0, 127)
    return w.bytes()

def geo_coordinates(w, lat, lon):
    latv = round(abs(lat) * ((1 << 23) - 1) / 90)
    lonv = round(lon * ((1 << 23) - 1) / 180)
    w.put(0, 2)                            # Geographical-Coordinates: extensible, one absent optional
    constrained(w, int(lat < 0), 0, 1)
    constrained(w, latv, 0, (1 << 23) - 1)
    constrained(w, lonv, -(1 << 23), (1 << 23) - 1)

def uncertainty_ellipse(w, semi_major, semi_minor, orientation):
    constrained(w, semi_major, 0, 127)
    constrained(w, semi_minor, 0, 127)
    constrained(w, orientation, 0, 89)

def altitude_and_direction(w, depth, altitude):
    w.put(0, 1)                            # extensible SEQUENCE, no optional root members
    constrained(w, int(depth), 0, 1)
    constrained(w, altitude, 0, 65535)

def geo_area_point(lat=38, lon=-90):
    w = Bits(); w.put(0, 1); w.put(0, 3)
    w.put(0, 2)
    geo_coordinates(w, lat, lon)
    return w.bytes()

def geo_area_ellipse(lat=38, lon=-90, semi_major=10, semi_minor=5, orientation=45, confidence=68):
    w = Bits(); w.put(0, 1); w.put(2, 3)
    w.put(0, 2)
    geo_coordinates(w, lat, lon)
    uncertainty_ellipse(w, semi_major, semi_minor, orientation)
    constrained(w, confidence, 0, 100)
    return w.bytes()

def geo_area_polygon(points=((38, -90), (39, -91), (37, -89))):
    w = Bits(); w.put(0, 1); w.put(3, 3)
    constrained(w, len(points), 1, 15)
    for lat, lon in points:
        w.put(0, 2)
        geo_coordinates(w, lat, lon)
    return w.bytes()

def geo_area_altitude(lat=38, lon=-90, depth=False, altitude=120):
    w = Bits(); w.put(0, 1); w.put(4, 3)
    w.put(0, 2)
    geo_coordinates(w, lat, lon)
    altitude_and_direction(w, depth, altitude)
    return w.bytes()

def geo_area_altitude_ellipsoid(lat=38, lon=-90, depth=False, altitude=120, semi_major=10, semi_minor=5, orientation=45, uncertainty_altitude=3, confidence=68):
    w = Bits(); w.put(0, 1); w.put(5, 3)
    w.put(0, 2)
    geo_coordinates(w, lat, lon)
    altitude_and_direction(w, depth, altitude)
    uncertainty_ellipse(w, semi_major, semi_minor, orientation)
    constrained(w, uncertainty_altitude, 0, 127)
    constrained(w, confidence, 0, 100)
    return w.bytes()

def geo_area_arc(lat=38, lon=-90, inner_radius=100, uncertainty_radius=10, offset_angle=30, included_angle=90, confidence=68):
    w = Bits(); w.put(0, 1); w.put(6, 3)
    w.put(0, 2)
    geo_coordinates(w, lat, lon)
    constrained(w, inner_radius, 0, 65535)
    constrained(w, uncertainty_radius, 0, 127)
    constrained(w, offset_angle, 0, 179)
    constrained(w, included_angle, 0, 179)
    constrained(w, confidence, 0, 100)
    return w.bytes()

def horizontal_speed_and_bearing(w, bearing, speed):
    constrained(w, bearing, 0, 359)
    constrained(w, speed, 0, 2047)

def vertical_velocity(w, speed, downward):
    constrained(w, speed, 0, 255)
    constrained(w, int(downward), 0, 1)

def velocity_horizontal(bearing=45, speed=30):
    w = Bits(); w.put(0, 1); w.put(0, 2)
    w.put(0, 2)
    horizontal_speed_and_bearing(w, bearing, speed)
    return w.bytes()

def velocity_horizontal_with_vertical(bearing=45, speed=30, vspeed=5, downward=False):
    w = Bits(); w.put(0, 1); w.put(1, 2)
    w.put(0, 2)
    horizontal_speed_and_bearing(w, bearing, speed)
    vertical_velocity(w, vspeed, downward)
    return w.bytes()

def velocity_horizontal_with_uncertainty(bearing=45, speed=30, uncertainty=7):
    w = Bits(); w.put(0, 1); w.put(2, 2)
    w.put(0, 2)
    horizontal_speed_and_bearing(w, bearing, speed)
    constrained(w, uncertainty, 0, 255)
    return w.bytes()

def velocity_horizontal_with_vertical_and_uncertainty(bearing=45, speed=30, vspeed=5, downward=False, huncertainty=7, vuncertainty=3):
    w = Bits(); w.put(0, 1); w.put(3, 2)
    w.put(0, 2)
    horizontal_speed_and_bearing(w, bearing, speed)
    vertical_velocity(w, vspeed, downward)
    constrained(w, huncertainty, 0, 255)
    constrained(w, vuncertainty, 0, 255)
    return w.bytes()

def positioning_data():
    w = Bits(); w.put(0, 1); w.put(1, 1); w.put(0, 2)
    constrained(w, 1, 1, 9); w.octets(b"\x13")
    return w.bytes()
def accuracy(value):
    w = Bits(); w.put(0, 1); constrained(w, value, 0, 1); return w.bytes()
def cause(branch, value):
    maxima = [0, 1, 6, 3]
    w = Bits(); constrained(w, branch, 0, 3); w.put(0, 1); constrained(w, value, 0, maxima[branch]); return w.bytes()

def pdu(category, ies):
    body = Bits(); body.put(0, 2)          # Location-Response extension and protocolExtensions absent
    constrained(body, len(ies), 0, 65535)
    for ident, crit, value in ies:
        constrained(body, ident, 0, 65535); criticality(body, crit); open_type(body, value)
    w = Bits(); w.put(0, 1); constrained(w, category, 0, 3)
    constrained(w, 0, 0, 255); criticality(w, 0); open_type(w, body.bytes())
    return w.bytes()

ID = bytes.fromhex("00000001")
def response(*extras): return pdu(1, [(2, 0, ID), *extras])
def failure(branch, value): return pdu(2, [(2, 0, ID), (11, 1, cause(branch, value))])

FIXTURES = {
    "success-location-estimate": response((12, 0, location_estimate())),
    "success-positioning-data": response((16, 0, positioning_data())),
    "success-accuracy-fulfilled": response((0, 0, accuracy(0))),
    "success-accuracy-not-fulfilled": response((0, 0, accuracy(1))),
    "success-estimate-positioning-data": response((12, 0, location_estimate()), (16, 0, positioning_data())),
    "success-authoritative-serving-cell-ecid": response((12, 0, location_estimate()), (16, 0, positioning_data()), (0, 0, accuracy(0))),
    "failure-radio-network-unspecified": failure(0, 0),
    "failure-transport-resource-unavailable": failure(1, 0),
    "failure-protocol-semantic-error": failure(2, 4),
    "failure-misc-unspecified": failure(3, 3),
    "failure-misc-om-intervention": failure(3, 2),
    "geo-area-point": geo_area_point(),
    "geo-area-ellipse": geo_area_ellipse(),
    "geo-area-polygon": geo_area_polygon(),
    "geo-area-altitude": geo_area_altitude(),
    "geo-area-altitude-ellipsoid": geo_area_altitude_ellipsoid(),
    "geo-area-arc": geo_area_arc(),
    "velocity-horizontal": velocity_horizontal(),
    "velocity-horizontal-with-vertical": velocity_horizontal_with_vertical(),
    "velocity-horizontal-with-uncertainty": velocity_horizontal_with_uncertainty(),
    "velocity-horizontal-with-vertical-and-uncertainty": velocity_horizontal_with_vertical_and_uncertainty(),
}

def main():
    OUT.mkdir(parents=True, exist_ok=True)
    records = []
    for name, wire in FIXTURES.items():
        path = OUT / (name + ".aper")
        path.write_bytes(wire)
        records.append({"name": name, "file": path.name, "sha256": hashlib.sha256(wire).hexdigest(), "hex": wire.hex()})
    manifest = {"specification": "3GPP TS 29.171", "version": "16.4.0", "encoding": "aligned PER",
                "compiler": {"name": "asn1tools", "version": "0.167.0", "role": "pinned leaf declaration compiler; complete PDU is independently composed because object-set OPEN TYPE dispatch is unsupported"},
                "source_correct_values": {"ecid_positioning_data": "4013", "accuracy_fulfilled": "00", "accuracy_not_fulfilled": "40", "lcs_cause_misc_unspecified": "d8"},
                "fixtures": records}
    (OUT / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    assert positioning_data().hex() == "4013"
    assert accuracy(0).hex() == "00" and accuracy(1).hex() == "40"
    assert cause(3, 3).hex() == "d8"

if __name__ == "__main__": main()
