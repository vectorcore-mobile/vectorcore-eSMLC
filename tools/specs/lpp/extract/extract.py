#!/usr/bin/env python3
"""Extract ordered ASN1START/ASN1STOP paragraphs from TS 37.355 DOCX."""
import hashlib, json, sys, zipfile
from pathlib import Path
from xml.etree import ElementTree as ET

ROOT = Path(__file__).resolve().parents[4]
ZIP = ROOT / "docs/specs/37355-g40.zip"
OUT = ROOT / "docs/specs/asn1/lpp/r16.4.0/source/original"
MANIFEST = ROOT / "docs/specs/asn1/lpp/r16.4.0/manifest.json"
SHA = "cb864a208c0ac63eb39fc2c51f9b98dc506314890d4af3f5534c99a816115ed6"
NS = {"w": "http://schemas.openxmlformats.org/wordprocessingml/2006/main"}

def digest(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def paragraph(p): return "".join(t.text or "" for t in p.findall(".//w:t", NS)).strip()

def main():
    if not ZIP.exists() or digest(ZIP) != SHA: raise SystemExit("source ZIP missing or checksum mismatch")
    with zipfile.ZipFile(ZIP) as outer:
        names = outer.namelist()
        if names != ["37355-g40.docx"]: raise SystemExit("unexpected outer ZIP members")
        docx = outer.read(names[0])
    docx_sha = hashlib.sha256(docx).hexdigest()
    with zipfile.ZipFile(__import__('io').BytesIO(docx)) as d:
        xml = d.read("word/document.xml")
    paras = [paragraph(p) for p in ET.fromstring(xml).findall(".//w:p", NS)]
    blocks, active = [], None
    for n, text in enumerate(paras, 1):
        if "ASN1START" in text:
            if active is not None: raise SystemExit("nested ASN1START at paragraph %d" % n)
            active = {"start_paragraph": n, "lines": []}
        elif "ASN1STOP" in text:
            if active is None: raise SystemExit("unmatched ASN1STOP at paragraph %d" % n)
            active["stop_paragraph"] = n; blocks.append(active); active = None
        elif active is not None:
            active["lines"].append(text)
    if active is not None or not blocks: raise SystemExit("unmatched marker or no ASN.1 blocks")
    OUT.mkdir(parents=True, exist_ok=True)
    records = []
    for i, b in enumerate(blocks, 1):
        text = "\n".join(x for x in b.pop("lines") if x) + "\n"
        name = "block-%03d.asn1" % i; path = OUT / name; path.write_text(text)
        records.append({"sequence":i,"original_filename":name,"original_sha256":digest(path),**b})
    MANIFEST.parent.mkdir(parents=True, exist_ok=True)
    MANIFEST.write_text(json.dumps({"specification":"3GPP TS 37.355","version":"16.4.0","release":16,"source_archive":"docs/specs/37355-g40.zip","source_archive_sha256":SHA,"source_docx":"37355-g40.docx","source_docx_sha256":docx_sha,"clause":"6.2 and all DOCX ASN1START/ASN1STOP blocks","blocks":records}, indent=2)+"\n")
    print("extracted", len(records), "blocks")
if __name__ == "__main__": main()
