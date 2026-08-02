#!/usr/bin/env python3
"""Compile and cross-check the independently generated TS 29.171 APER leaves."""
import importlib.metadata
import runpy
from pathlib import Path

if importlib.metadata.version("asn1tools") != "0.167.0":
    raise SystemExit("expected pinned asn1tools 0.167.0")
import asn1tools

ASN1 = """
LCSAPFixture DEFINITIONS AUTOMATIC TAGS ::= BEGIN
Accuracy-Fulfillment-Indicator ::= ENUMERATED { fulfilled, not-fulfilled, ... }
Positioning-Data ::= SEQUENCE {
 positioning-Data-Set SEQUENCE (SIZE (1..9)) OF OCTET STRING (SIZE (1)) OPTIONAL,
 gNSS-Positioning-Data-Set SEQUENCE (SIZE (1..9)) OF OCTET STRING (SIZE (1)) OPTIONAL,
 extensions OCTET STRING OPTIONAL, ... }
LCS-Cause ::= CHOICE {
 radio ENUMERATED { unspecified, ... },
 transport ENUMERATED { unavailable, unspecified, ... },
 protocol ENUMERATED { transfer, reject, ignore-notify, incompatible, semantic, unspecified, abstract, ... },
 misc ENUMERATED { overload, hardware, om, unspecified, ... } }
Geographical-Coordinates ::= SEQUENCE {
 sign ENUMERATED { north, south },
 latitude INTEGER (0..8388607),
 longitude INTEGER (-8388608..8388607),
 extensions OCTET STRING OPTIONAL, ... }
Uncertainty-Ellipse ::= SEQUENCE {
 semi-major INTEGER (0..127),
 semi-minor INTEGER (0..127),
 orientation INTEGER (0..89) }
Altitude-And-Direction ::= SEQUENCE {
 direction ENUMERATED { height, depth },
 altitude INTEGER (0..65535), ... }
Point ::= SEQUENCE {
 coordinates Geographical-Coordinates,
 extensions OCTET STRING OPTIONAL, ... }
Point-With-Uncertainty ::= SEQUENCE {
 coordinates Geographical-Coordinates,
 uncertainty INTEGER (0..127),
 extensions OCTET STRING OPTIONAL, ... }
Ellipsoid-Point-With-Uncertainty-Ellipse ::= SEQUENCE {
 coordinates Geographical-Coordinates,
 uncertainty Uncertainty-Ellipse,
 confidence INTEGER (0..100),
 extensions OCTET STRING OPTIONAL, ... }
Polygon-Point ::= SEQUENCE {
 coordinates Geographical-Coordinates,
 extensions OCTET STRING OPTIONAL, ... }
Polygon ::= SEQUENCE (SIZE (1..15)) OF Polygon-Point
Ellipsoid-Point-With-Altitude ::= SEQUENCE {
 coordinates Geographical-Coordinates,
 altitude Altitude-And-Direction,
 extensions OCTET STRING OPTIONAL, ... }
Ellipsoid-Point-With-Altitude-And-Uncertainty-Ellipsoid ::= SEQUENCE {
 coordinates Geographical-Coordinates,
 altitude Altitude-And-Direction,
 uncertainty Uncertainty-Ellipse,
 uncertainty-altitude INTEGER (0..127),
 confidence INTEGER (0..100),
 extensions OCTET STRING OPTIONAL, ... }
Ellipsoid-Arc ::= SEQUENCE {
 coordinates Geographical-Coordinates,
 inner-radius INTEGER (0..65535),
 uncertainty-radius INTEGER (0..127),
 offset-angle INTEGER (0..179),
 included-angle INTEGER (0..179),
 confidence INTEGER (0..100),
 extensions OCTET STRING OPTIONAL, ... }
Horizontal-Speed-And-Bearing ::= SEQUENCE {
 bearing INTEGER (0..359),
 horizontal-speed INTEGER (0..2047) }
Vertical-Velocity ::= SEQUENCE {
 vertical-speed INTEGER (0..255),
 direction ENUMERATED { upward, downward } }
Horizontal-Velocity ::= SEQUENCE {
 speed Horizontal-Speed-And-Bearing,
 extensions OCTET STRING OPTIONAL, ... }
Horizontal-With-Vertical-Velocity ::= SEQUENCE {
 speed Horizontal-Speed-And-Bearing,
 vertical Vertical-Velocity,
 extensions OCTET STRING OPTIONAL, ... }
Horizontal-Velocity-With-Uncertainty ::= SEQUENCE {
 speed Horizontal-Speed-And-Bearing,
 uncertainty INTEGER (0..255),
 extensions OCTET STRING OPTIONAL, ... }
Horizontal-With-Vertical-Velocity-And-Uncertainty ::= SEQUENCE {
 speed Horizontal-Speed-And-Bearing,
 vertical Vertical-Velocity,
 horizontal-uncertainty INTEGER (0..255),
 vertical-uncertainty INTEGER (0..255),
 extensions OCTET STRING OPTIONAL, ... }
Geographical-Area ::= CHOICE {
 point Point,
 point-with-uncertainty Point-With-Uncertainty,
 ellipse Ellipsoid-Point-With-Uncertainty-Ellipse,
 polygon Polygon,
 altitude Ellipsoid-Point-With-Altitude,
 altitude-ellipsoid Ellipsoid-Point-With-Altitude-And-Uncertainty-Ellipsoid,
 arc Ellipsoid-Arc, ... }
Velocity-Estimate ::= CHOICE {
 horizontal Horizontal-Velocity,
 horizontal-with-vertical Horizontal-With-Vertical-Velocity,
 horizontal-with-uncertainty Horizontal-Velocity-With-Uncertainty,
 horizontal-with-vertical-and-uncertainty Horizontal-With-Vertical-Velocity-And-Uncertainty, ... }
END
"""

def main():
    codec = asn1tools.compile_string(ASN1, "per")
    generated = runpy.run_path(str(Path(__file__).with_name("generate.py")))
    coords = {"sign": "north", "latitude": round(38 * ((1 << 23) - 1) / 90), "longitude": round(-90 * ((1 << 23) - 1) / 180)}
    checks = [
        ("Accuracy-Fulfillment-Indicator", "fulfilled", generated["accuracy"](0)),
        ("Accuracy-Fulfillment-Indicator", "not-fulfilled", generated["accuracy"](1)),
        ("Positioning-Data", {"positioning-Data-Set": [b"\x13"]}, generated["positioning_data"]()),
        ("LCS-Cause", ("misc", "unspecified"), generated["cause"](3, 3)),
        ("LCS-Cause", ("misc", "om"), generated["cause"](3, 2)),
        ("Geographical-Area", ("point-with-uncertainty", {"coordinates": coords, "uncertainty": 4}), generated["location_estimate"]()),
        ("Geographical-Area", ("point", {"coordinates": coords}), generated["geo_area_point"]()),
        ("Geographical-Area",
         ("ellipse", {"coordinates": coords, "uncertainty": {"semi-major": 10, "semi-minor": 5, "orientation": 45}, "confidence": 68}),
         generated["geo_area_ellipse"]()),
        ("Geographical-Area",
         ("polygon", [{"coordinates": coords}, {"coordinates": {"sign": "north", "latitude": round(39 * ((1 << 23) - 1) / 90), "longitude": round(-91 * ((1 << 23) - 1) / 180)}}, {"coordinates": {"sign": "north", "latitude": round(37 * ((1 << 23) - 1) / 90), "longitude": round(-89 * ((1 << 23) - 1) / 180)}}]),
         generated["geo_area_polygon"]()),
        ("Geographical-Area", ("altitude", {"coordinates": coords, "altitude": {"direction": "height", "altitude": 120}}), generated["geo_area_altitude"]()),
        ("Geographical-Area",
         ("altitude-ellipsoid", {"coordinates": coords, "altitude": {"direction": "height", "altitude": 120}, "uncertainty": {"semi-major": 10, "semi-minor": 5, "orientation": 45}, "uncertainty-altitude": 3, "confidence": 68}),
         generated["geo_area_altitude_ellipsoid"]()),
        ("Geographical-Area", ("arc", {"coordinates": coords, "inner-radius": 100, "uncertainty-radius": 10, "offset-angle": 30, "included-angle": 90, "confidence": 68}), generated["geo_area_arc"]()),
        ("Velocity-Estimate", ("horizontal", {"speed": {"bearing": 45, "horizontal-speed": 30}}), generated["velocity_horizontal"]()),
        ("Velocity-Estimate", ("horizontal-with-vertical", {"speed": {"bearing": 45, "horizontal-speed": 30}, "vertical": {"vertical-speed": 5, "direction": "upward"}}), generated["velocity_horizontal_with_vertical"]()),
        ("Velocity-Estimate", ("horizontal-with-uncertainty", {"speed": {"bearing": 45, "horizontal-speed": 30}, "uncertainty": 7}), generated["velocity_horizontal_with_uncertainty"]()),
        ("Velocity-Estimate",
         ("horizontal-with-vertical-and-uncertainty", {"speed": {"bearing": 45, "horizontal-speed": 30}, "vertical": {"vertical-speed": 5, "direction": "upward"}, "horizontal-uncertainty": 7, "vertical-uncertainty": 3}),
         generated["velocity_horizontal_with_vertical_and_uncertainty"]()),
    ]
    for type_name, value, expected in checks:
        actual = codec.encode(type_name, value)
        if actual != expected:
            raise SystemExit(f"APER compiler mismatch {type_name}: {actual.hex()} != {expected.hex()}")
    print("asn1tools 0.167.0 leaf APER checks passed")

if __name__ == "__main__": main()
