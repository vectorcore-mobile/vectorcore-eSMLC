# Pinned APER reference compiler

This directory pins `asn1tools` 0.167.0 and `bitstruct` 8.21.0.  The compiler
is development-only and is deliberately isolated in `.venv`; nothing in Go
builds, production images, or runtime invokes Python or a native ASN.1
runtime.  `generate.py` independently composes the TS 29.171 complete-message
APER fixtures and verifies the scalar APER encodings against the pinned tool.
