# ECID measured-result scalar analysis artifacts

`scalar-harness.asn1` deliberately copies only the verified root scalar
constraints from `LPP-PDU-Definitions`; it is not a production type and not a
complete `MeasuredResultsElement`. Generate `fixtures.json` with the pinned
development-only compiler:

```bash
tools/specs/lpp/reference-codec/.venv/bin/python generate.py
```

The generated fixture metadata is deterministic. The root scalar values are
non-extensible; `arfcnEUTRA-v9a0` is an extension addition of the enclosing
`MeasuredResultsElement`, outside this scalar-only root closure.
