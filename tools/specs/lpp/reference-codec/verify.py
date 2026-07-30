#!/usr/bin/env python3
"""Verify the pinned development-only asn1tools installation."""
from __future__ import annotations

import asn1tools
import bitstruct

EXPECTED_ASN1TOOLS = "0.167.0"
EXPECTED_BITSTRUCT = "8.21.0"
if asn1tools.__version__ != EXPECTED_ASN1TOOLS:
    raise SystemExit(f"asn1tools version mismatch: {asn1tools.__version__}")
if bitstruct.__version__ != EXPECTED_BITSTRUCT:
    raise SystemExit(f"bitstruct version mismatch: {bitstruct.__version__}")
print(f"asn1tools {asn1tools.__version__}; bitstruct {bitstruct.__version__}; UPER mode name: uper")
