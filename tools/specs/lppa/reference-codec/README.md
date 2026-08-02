# Pinned APER reference compiler

This directory pins `asn1tools` 0.167.0 and `bitstruct` 8.21.0, the same
versions pinned for LCS-AP (see `tools/specs/lcsap/reference-codec`). The
compiler is development-only and isolated in `.venv`; nothing in Go builds,
production images, or runtime invokes Python or a native ASN.1 runtime.

`generate.py` independently composes the TS 36.455 (LPPa) E-CID measurement
procedure APER fixtures used by `internal/lppa`'s tests and verifies the leaf
APER encodings against the pinned tool, exactly as the LCS-AP reference codec
does for TS 29.171.
