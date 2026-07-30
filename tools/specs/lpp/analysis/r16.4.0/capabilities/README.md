# R16.4.0 capability analysis

Run with the pinned development-only compiler environment:

```bash
tools/specs/lpp/reference-codec/.venv/bin/python tools/specs/lpp/analysis/r16.4.0/capabilities/generate.py
```

The script verifies no runtime package, ASN.1 source, or existing envelope
fixture is rewritten. It produces deterministic inventories and independent
UPER fixtures for the recommended ECID-only capability subset.
