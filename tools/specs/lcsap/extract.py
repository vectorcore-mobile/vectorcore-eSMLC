#!/usr/bin/env python3
"""Recover the five ASN.1 modules published in TS 29.171 V16.4.0."""
import hashlib
import io
import json
import shutil
import zipfile
from pathlib import Path
from xml.etree import ElementTree as ET

ROOT = Path(__file__).resolve().parents[3]
SOURCE = ROOT / "docs/specs/29171-g40.zip"
OUT = ROOT / "docs/specs/asn1/lcsap/r16.4.0"
EXPECTED_SHA256 = ""  # Filled by the checked-in manifest after first recovery.
NS = {"w": "http://schemas.openxmlformats.org/wordprocessingml/2006/main"}
MODULES = {
    "LCS-AP-PDU-Descriptions": "lcs-ap-pdu-descriptions.asn1",
    "LCS-AP-PDU-Contents": "lcs-ap-pdu-contents.asn1",
    "LCS-AP-IEs": "lcs-ap-ies.asn1",
    "LCS-AP-CommonDataTypes": "lcs-ap-common-data-types.asn1",
    "LCS-AP-Constants": "lcs-ap-constants.asn1",
    "LCS-AP-Containers": "lcs-ap-containers.asn1",
}

def sha(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()

def paragraphs():
    with zipfile.ZipFile(SOURCE) as archive:
        if archive.namelist() != ["29171-g40.docx"]:
            raise SystemExit("unexpected TS 29.171 archive member list")
        docx = archive.read("29171-g40.docx")
    with zipfile.ZipFile(io.BytesIO(docx)) as document:
        xml = document.read("word/document.xml")
    root = ET.fromstring(xml)
    return docx, ["".join(t.text or "" for t in p.findall(".//w:t", NS)).strip()
                  for p in root.findall(".//w:p", NS)]

def recover(lines):
    found = {}
    for title, filename in MODULES.items():
        try:
            start = next(i for i, line in enumerate(lines) if line == title or line.startswith(title + " {"))
            end = next(i for i in range(start, len(lines)) if lines[i] == "END")
        except StopIteration as error:
            raise SystemExit("could not recover " + title) from error
        # DOCX paragraph boundaries are formatting only; preserve each non-empty
        # paragraph in source order so source provenance remains auditable.
        found[filename] = {"start_paragraph": start + 1, "end_paragraph": end + 1,
                           "text": "\n".join(x for x in lines[start:end + 1] if x) + "\n"}
    return found

def normalize(text):
    # Preserve ASN.1 tokens while making DOCX extraction stable across host
    # newline conventions and eliminating only blank presentation paragraphs.
    return "\n".join(line.rstrip() for line in text.splitlines() if line.strip()) + "\n"

def main():
    if not SOURCE.exists():
        raise SystemExit("authoritative source archive is missing")
    source_hash = sha(SOURCE)
    docx, lines = paragraphs()
    modules = recover(lines)
    raw = OUT / "source" / "original"
    normalized = OUT / "modules" / "normalized"
    if raw.exists(): shutil.rmtree(raw)
    if normalized.exists(): shutil.rmtree(normalized)
    raw.mkdir(parents=True)
    normalized.mkdir(parents=True)
    records = []
    for filename, entry in modules.items():
        raw_path = raw / filename
        normalized_path = normalized / filename
        raw_path.write_text(entry["text"])
        normalized_path.write_text(normalize(entry["text"]))
        records.append({"module": filename.removesuffix(".asn1"), "file": filename,
                        "source_paragraphs": [entry["start_paragraph"], entry["end_paragraph"]],
                        "original_sha256": sha(raw_path), "normalized_sha256": sha(normalized_path),
                        "normalization": "strip empty DOCX presentation paragraphs; trim trailing whitespace; LF newline"})
    closure = {
        "root": "Location-Response",
        "complete_message_modules": ["lcs-ap-pdu-descriptions", "lcs-ap-pdu-contents", "lcs-ap-containers", "lcs-ap-common-data-types", "lcs-ap-constants"],
        "information_element_module": "lcs-ap-ies",
        "root_types": ["LCS-AP-PDU", "Location-Response", "Geographical-Area", "Positioning-Data", "Accuracy-Fulfillment-Indicator", "LCS-Cause"],
        "supported_geographical_area": "point-With-Uncertainty",
        "unsupported": ["protocol extensions", "Geographical-Area extensions and non-point-with-uncertainty alternatives", "GNSS positioning data set"],
    }
    (OUT / "dependency-closure.json").write_text(json.dumps(closure, indent=2) + "\n")
    manifest = {"specification": "3GPP TS 29.171", "version": "16.4.0", "release": 16,
                "source_archive": "docs/specs/29171-g40.zip", "source_archive_sha256": source_hash,
                "source_docx": "29171-g40.docx", "source_docx_sha256": hashlib.sha256(docx).hexdigest(),
                "extraction": "word/document.xml paragraphs, module title through END inclusive", "modules": records}
    (OUT / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")

if __name__ == "__main__": main()
