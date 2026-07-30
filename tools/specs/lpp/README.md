# TS 37.355 development-only tooling

This Python 3 standard-library workflow extracts marker-delimited ASN.1 from
TS 37.355 V16.4.0. Run `tools/specs/lpp/scripts/extract.sh` from any directory.
It verifies the source ZIP checksum and writes committed outputs under
`docs/specs/asn1/lpp/r16.4.0`. Python is not needed by runtime, Go tests, or
default Make targets. `/tmp` is not used.

The analysis stage consumes only `modules_v2/normalized`, verifies their
canonical hashes, and never rewrites them:

```
tools/specs/lpp/scripts/analyze.sh --verify-inputs
tools/specs/lpp/scripts/verify-v2-inputs.sh
tools/specs/lpp/scripts/analyze.sh
python3 -m unittest discover -s tools/specs/lpp/analyze/tests -v
tools/specs/lpp/scripts/verify-analysis-determinism.sh
```

It is a structural parser, not a codec or code generator. Its reports form the
evidence for a later compiler-subset phase. No default Go or Make target runs
these commands.

The compiler/fixture phase continues from corrected V2 inputs only:

```
corrected V2 extraction → normalized modules → structural analysis
→ compiler-complete subset → pinned UPER reference compiler
→ independent fixtures → future pure-Go internal/uper
```

Run explicitly:

```
tools/specs/lpp/scripts/generate-subset.sh
tools/specs/lpp/scripts/verify-analysis-state.sh
tools/specs/lpp/reference-codec/install.sh
tools/specs/lpp/scripts/compile-reference-uper.sh
tools/specs/lpp/scripts/generate-reference-fixtures.sh
```

The 19-definition closure is only a minimum envelope. The compiler-complete
target expands to the byte-identical 646-definition PDU module because the
normative `LPP-MessageBody` CHOICE is not an opaque boundary. Fixtures validate
UPER mechanics, not positioning accuracy or GNSS behavior.
