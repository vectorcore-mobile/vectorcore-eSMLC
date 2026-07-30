#!/usr/bin/env python3
"""Fail closed before analysis: verify only canonical corrected V2 inputs."""
from __future__ import annotations

import hashlib
from pathlib import Path
import zipfile


ROOT = Path(__file__).resolve().parents[4]
ZIP = ROOT / "docs/specs/37355-g40.zip"
ZIP_SHA = "cb864a208c0ac63eb39fc2c51f9b98dc506314890d4af3f5534c99a816115ed6"
DOCX = "37355-g40.docx"
DOCX_SHA = "23de490200f1bf7449c149172b7d45c3c8dd8af651ce0d17ca2d8ffdc66617f0"
REQUIRED = [
    ROOT / "docs/specs/asn1/lpp/r16.4.0/source_v2/manifest.json",
    ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/normalization-manifest.json",
    ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/normalized/LPP-PDU-Definitions.asn1",
    ROOT / "docs/specs/asn1/lpp/r16.4.0/modules_v2/normalized/LPP-Broadcast-Definitions.asn1",
]


def digest(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def main() -> None:
    if digest(ZIP.read_bytes()) != ZIP_SHA:
        raise SystemExit("TS 37.355 ZIP checksum mismatch")
    with zipfile.ZipFile(ZIP) as archive:
        if archive.namelist() != [DOCX]:
            raise SystemExit("unexpected TS 37.355 ZIP members")
        if digest(archive.read(DOCX)) != DOCX_SHA:
            raise SystemExit("TS 37.355 DOCX checksum mismatch")
    missing = [str(path.relative_to(ROOT)) for path in REQUIRED if not path.is_file()]
    if missing:
        raise SystemExit("missing canonical V2 input(s): " + ", ".join(missing))
    print("TS 37.355 archive, DOCX, corrected V2 provenance, and normalized inputs verified")


if __name__ == "__main__":
    main()
