# MeasuredResultsElement field analysis artifacts

`measured-results-element-harness.asn1` is an analysis-only exact extraction
of the normalized declaration at `LPP-PDU-Definitions.asn1:3955-3975` and its
local dependencies. Generate `fixtures.json` using the pinned development-only
compiler:

```bash
tools/specs/lpp/reference-codec/.venv/bin/python generate.py
```

The script encodes, decodes, and re-encodes every fixture. It is not used by
production builds or runtime packages.
