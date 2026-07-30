# Staged specification sources

| Artifact | Source | SHA-256 | Version / release | Purpose |
|---|---|---|---|---|
| `29171-g40.zip` | `/usr/src/vectorcore-mme.old/docs/specs/29171-g40.zip` | `c7cde985b781acb3ad90b1cb31c1ecf085413aac2bbae28390f0cdf6765635dd` | 3GPP TS 29.171 V16.4.0, Release 16, September 2022 | Authoritative LCS-AP source audit |
| `37355-g40.zip` | active repository addition | `cb864a208c0ac63eb39fc2c51f9b98dc506314890d4af3f5534c99a816115ed6` | 3GPP TS 37.355 V16.4.0, Release 16, March 2021 | Authoritative LPP source audit |

The ZIP is copied byte-for-byte and is not unpacked or modified in this
repository. It contains `29171-g40.docx`. The active runtime does not read it;
it is documentation and development reference only. No generator is required
by `make build`, `make test`, `go build`, or `go test`.

`37355-g40.zip` is intact and contains `37355-g40.docx`. Clause 6.2 contains
the normative LPP PDU ASN.1. The authoritative corrected provenance is
`asn1/lpp/r16.4.0/source_v2`, with V2 normalized modules and deterministic
structural analysis under `asn1/lpp/r16.4.0/modules_v2`. These artifacts are
development reference only: no runtime codec or generated runtime output is
present, and no generator is required by default build or test targets.

The compiler-complete minimum-envelope target is under
`asn1/lpp/r16.4.0/modules_v2/subsets/minimum-envelope`. It preserves the
normalized PDU module byte-for-byte: the 19-definition structural envelope is
not a standalone ASN.1 module because its normative CHOICE branches pull in
the whole 646-definition PDU closure. A pinned, development-only UPER fixture
workflow lives under `tools/specs/lpp/reference-codec`; it is not a runtime
dependency.
