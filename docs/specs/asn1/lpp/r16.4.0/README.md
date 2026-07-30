# TS 37.355 LPP extraction status

Source: `docs/specs/37355-g40.zip`, member `37355-g40.docx`, 3GPP TS 37.355
V16.4.0 (2021-03), Release 16. `source/extracted/LPP-PDU-Definitions.outer.asn1`
is a bounded, syntax-preserving transcription of the outer definitions from
clause 6.2, not a complete module set. No normalization and no generator were
used. Its source is the Annex/section text recovered from Word XML using:

```sh
unzip -p docs/specs/37355-g40.zip 37355-g40.docx > /tmp/37355-g40.docx
unzip -p /tmp/37355-g40.docx word/document.xml
```

The full document contains the nested Release-16 common and positioning IEs,
but a complete normalized dependency closure and independent PER vector have
not yet been produced. `source/original/` contains 366 marker-delimited blocks
and `manifest.json` records their deterministic checksums and DOCX paragraph
bounds. This material is reference-only and is not read by the runtime or
normal build.
