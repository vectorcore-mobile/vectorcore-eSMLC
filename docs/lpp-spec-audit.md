# LPP specification-source audit

## Decision

**Decision C — no useful TS 36.355 LPP ASN.1 material found.** No LPP codec,
transaction identifier, GNSS IE, fixture, or runtime behavior was added. The
exact missing input is an authoritative, release-identified TS 36.355/37.355
archive containing `LPP-PDU-Definitions`, `LPP-Message`, transaction,
message-body, common-IE, and A-GNSS ASN.1 modules plus a compatible independent
APER fixture.

## Reopened audit — TS 37.355 added

The prior Decision C remains historically correct for the then-available
trees. The active repository now contains `docs/specs/37355-g40.zip`, SHA-256
`cb864a208c0ac63eb39fc2c51f9b98dc506314890d4af3f5534c99a816115ed6`.
It is intact, contains `37355-g40.docx`, and identifies itself as **3GPP TS
37.355 V16.4.0 (2021-03), LTE Positioning Protocol (LPP), Release 16**.

Clause 6.2 contains `LPP-PDU-Definitions`, `LPP-Message`,
`LPP-TransactionID`, Request/ProvideCapabilities,
Request/ProvideLocationInformation, Abort, Error, common IEs, A-GNSS and
location-coordinate definitions. The bounded outer extraction records the
outer message closure: transaction ID has extensible Initiator
`{locationServer,targetDevice,...}` and TransactionNumber `(0..255)`;
LPP-Message contains optional transaction ID, mandatory endTransaction,
optional sequence number `(0..255)`, optional acknowledgement, and optional
message-body. LPP uses BASIC-PER **unaligned** variant, not the aligned APER
used by LCS-AP.

**New decision: B — authoritative but not yet implementation-sufficient.**
The document supplies the definitions, and TS 37.355 V16.4.0 aligns to the
Release-16 TS 29.171 V16.4.0 LPP APDU carrier. However, this phase has not
produced a complete normalized module/dependency graph or an independent APER
fixture. The extracted Word text also needs a reviewed full-module extraction
before code generation or a hand-written codec can be trusted. No LPP runtime
code, APER primitive, generated file, fixture, or tooling was introduced.

## Closure phase update

A repository-local Python 3 extractor at `tools/specs/lpp/extract/extract.py`
now verifies the ZIP and DOCX structure and emits every DOCX
`ASN1START`/`ASN1STOP` region deterministically. It found **366 blocks**;
their original extracted text is in `docs/specs/asn1/lpp/r16.4.0/source/original/`
and checksums/paragraph bounds are in `manifest.json`. This is source recovery,
not normalization: blocks are type-level regions and have not yet been safely
reconstructed into compiler-ready ASN.1 modules. Consequently no import graph,
dependency closure, normalized module set, reference UPER compiler, or
independent fixture exists yet, and Decision B remains in force.

The next exact step is to extract all ASN1START/ASN1STOP segments from clause
6.2 into a complete reviewed `LPP-PDU-Definitions` module, record source and
normalized hashes, then generate an independent unaligned-PER fixture with a
development-only encoder before implementing `internal/lpp`.

## Staged authoritative source

`docs/specs/29171-g40.zip` is the byte-identical copy of the archived TS
29.171 V16.4.0 Release-16 archive. It contains one DOCX, `29171-g40.docx`, and
is authoritative for SLs/LCS-AP, not LPP. Its Annex A yields the six
Release-16 modules: LCS-AP-PDU-Descriptions, LCS-AP-PDU-Contents, LCS-AP-IEs,
LCS-AP-CommonDataTypes, LCS-AP-Constants, and LCS-AP-Containers. The archived
`specs/asn1/lcsap/r16.4.0` work is deterministic DOCX extraction, module split,
representational normalization, and one reviewed import-comma patch; its
manifest records archive, DOCX, and document-XML hashes.

## Complete scoped inventory

| Paths | Type / apparent source | LCS-AP | LPP / LPPa | Reuse |
|---|---|---:|---:|---|
| `docs/specs/29171-g40.zip` | TS 29.171 V16.4.0 DOCX archive | yes | transport only | copied authoritative |
| `docs/specs/README.md` | provenance documentation | no | no | reference |
| `docs/specs/TS_23.003_Numbering_Identities_Rel16_v16.12.0.zip` | TS 23.003 | no | no | unrelated |
| `docs/specs/TS_24.301_NAS_EPS_Rel16_v16.9.0.pdf` | TS 24.301 | no | NAS LPP carrier only | reference |
| `docs/specs/TS_29.272_S6a_S13_Diameter_Rel16_v16.8.0.zip` | TS 29.272 | no | no | unrelated |
| `docs/specs/TS_29.274_GTPv2C_Rel16_v16.13.0.zip` | TS 29.274 | no | no | unrelated |
| `docs/specs/TS_33.401_LTE_Security_Rel16_v16.4.0.pdf` | TS 33.401 | no | no | unrelated |
| `docs/specs/TS_36.413_S1AP_Rel16_v16.14.0.pdf` | TS 36.413 | no | LPPa transport context only | reference |
| `docs/specs/asn1/s1ap-rel16-phase1-derived.md`; `docs/specs/nas/ts24.301-rel16-phase1-derived.md` | derived metadata | no | carrier context | unrelated |
| `specs/asn1/lcsap/r16.4.0/{README.md,SHA256SUMS}` | provenance/checksums | yes | no | reference |
| `specs/asn1/lcsap/r16.4.0/normalized/*.asn1` | six normalized TS 29.171 modules | yes | no | reference only |
| `specs/asn1/lcsap/r16.4.0/source/extracted/*.asn1` | syntax-preserving extraction | yes | no | reference only |
| `specs/asn1/lcsap/r16.4.0/source/raw/{annex.txt,extraction-manifest.json}` | DOCX Annex A/provenance | yes | no | reference only |
| `specs/asn1/lcsap/r16.4.0/patches/*` | reviewed normalization patch | yes | no | reference only |

No scoped file name or content matched LPP-PDU-Definitions, LPP-Message,
RequestCapabilities, ProvideCapabilities, RequestLocationInformation,
ProvideLocationInformation, GNSS-LocationInformation, or A-GNSS ASN.1. There
are no LPP/LPPa generated Go tables, binary fixtures, packet fixtures,
extraction scripts, Makefiles, or APER tools in the scoped trees.

## LCS-AP validation

| Procedure | Code | PDU / procedure criticality | Relevant IEs | Active status |
|---|---:|---|---|---|
| Location Service Request | 0 | initiating/successful/unsuccessful; reject | Correlation-ID 2 reject M; Location-Type 13 reject M; ECGI 4 ignore M | implemented subset |
| Connection-Oriented Information | 1 | initiating; reject | Correlation-ID 2 reject M; Payload-Type 15 reject M; APDU 1 reject M | validated container |
| Reset | 4 | initiating/successful; reject | Reset Request LCS-Cause 11 ignore M; acknowledgement none | acknowledgement only |

The audit corrected outbound Connection-Oriented Information criticality from
`ignore` to `reject`. The decoder still accepts `ignore` solely for current
MME compatibility; this is a documented MME interoperability deviation. The
current generic APER codec remains scoped compatibility code, not a complete
generated TS 29.171 implementation.

Reset validation now requires its mandatory LCS-Cause and sends a
`reject`-criticality Reset Acknowledge, while accepting the deployed MME's
inbound `ignore` compatibility deviation.

The ASN.1 object set labels these IEs optional to permit reuse, but normative
Table 7.3.4-1 makes all three mandatory for this message; the active validator
therefore correctly rejects an absent IE.

## Build boundary

Runtime language: Go. Normal build dependencies: Go modules only. The
development-only analysis tooling uses Python 3 standard library under
`tools/specs/lpp`; it is not called by `go build`, `go test`, `make build`,
`make test`, or `make vet`. The staged ZIP, corrected provenance, analysis
reports, and tooling are not runtime inputs.
## Structural analysis — V2 normalized modules

The V2 normalized modules are now analysed by repository-local Python 3
standard-library tooling under `tools/specs/lpp/analyze`. The pipeline is
`normalized ASN.1 → lexer → structural parser → symbol inventory → reference
graph → bounded dependency closure`. It performs no token repair and only
recognises references that are declared local definitions or declared imports;
this prevents SEQUENCE field labels, ENUMERATED labels, named numbers and named
bits from being reported as external ASN.1 dependencies.

The reports are in `docs/specs/asn1/lpp/r16.4.0/modules_v2/analysis/`. This is
an analysis milestone only: it does not provide a compiler subset, an
independent UPER fixture, `internal/uper`, `internal/lpp`, or GNSS support.

## Analysis decision — V2 normalized modules

The analysis stage verifies the exact corrected V2 normalized hashes before it
reads either module. It tokenizes without source repair, records locations, and
only emits direct reference edges for a local definition or a declared import.
That rule excludes ASN.1 field labels, CHOICE labels, ENUMERATED labels, named
numbers, and named bits by construction.

The two modules contain 652 assignments (615 type assignments and 37 value
assignments), no duplicate definitions, and five imports. All five imports are
from `LPP-Broadcast-Definitions` to symbols present in
`LPP-PDU-Definitions`; there are no unresolved imports or structural symbol
references in the supplied pair. The strict selected-procedure closure has 646
definitions because it follows optional and method-specific branches. The
separate 19-definition minimum-form envelope closure covers only the selected
R9 wrappers whose root IE members are optional; it is explicitly not a claim
that arbitrary LPP positioning payloads can be decoded.

**Decision A — ready for a standards-complete bounded compiler-subset phase.**
The symbol graph and closure are complete enough to construct and validate a
compiler input without guessing. This decision does *not* authorize a runtime
codec: no compiler subset, independent UPER fixture, `internal/uper`,
`internal/lpp`, capability negotiation, GNSS payload processing, coordinate
extraction, or live UE positioning has been implemented.

The exact next step is to generate a standards-complete bounded compiler subset
from the V2 analysis, pin a development-only UPER reference compiler, obtain
independent fixtures, and only then begin a pure-Go runtime codec phase.

## Compiler subset and independent UPER fixtures

The compiler-complete subset generator is `tools/specs/lpp/subset/generate.py`.
It re-verifies V2 hashes and analysis closure before writing its output. The
minimum structural envelope has 19 definitions, but the complete normative
`LPP-MessageBody` CHOICE means a valid compiler input requires the entire
646-definition `LPP-PDU-Definitions` module. The generated subset is therefore
byte-identical to that normalized PDU module; six unreferenced Broadcast-module
assignments are omitted. The subset validator finds zero unresolved references
and zero duplicate definitions.

`asn1tools` 0.167.0 (MIT), with `bitstruct` 8.21.0 and `pyparsing` 3.3.2, is
pinned as a development-only UPER reference compiler under
`tools/specs/lpp/reference-codec`. Both the full V2 module pair and the
compiler-complete PDU target were accepted with explicit `uper` encoding. The
workflow generated independent byte and pre-padding-bit-length fixtures under
`tools/specs/lpp/fixtures/r16.4.0`; they cover transaction number boundaries,
both initiators, terminal and non-terminal envelopes, sequence and
acknowledgement fields, and empty R9 wrappers for the selected messages.

This validates codec mechanics only. It does not add a runtime codec, LPP
session state, GNSS capability negotiation, estimates, assistance data, or live
positioning behavior.

## Pure-Go UPER primitive phase

`internal/uper` now implements the bounded, schema-neutral unaligned-PER
primitive layer required by the independent fixture corpus: MSB-first bit I/O,
exact bit lengths and zero padding, constrained whole numbers, BOOLEAN, root
ENUMERATED and CHOICE indexes, optional bitmaps, and extension-presence bits.
Fixture envelope mapping is test-only and no `internal/lpp` package exists.
The phase is documented in
`docs/specs/asn1/lpp/r16.4.0/uper-runtime-phase1-report.md`.

**UPER primitive decision: A.** The 11 independent fixtures decode and
re-encode byte/bit identically with strict padding and bounded malformed-input
tests. The next scope may be a bounded LPP outer-envelope model only; GNSS and
positioning payload semantics remain deferred.

## Bounded LPP envelope and transaction-control phases

`internal/lpp` implements the fixture-covered TS 37.355 outer envelope and
empty Release 9 procedure wrappers using `internal/uper`; it remains a codec
only. `internal/lpp/transaction` now provides caller-scoped, transport-neutral
control state above that codec. Its comparable key is initiator plus transaction
number, local/remote ownership and inbound/outbound direction are explicit, and
all time is supplied by the caller. It has bounded active/retained capacity,
explicit pruning, duplicate fingerprints, optional sequence tracking, one
pending acknowledgement, and terminal Completed/Aborted/Failed retention.

The package neither opens a connection nor starts a goroutine, schedules a
timer, sends a response, or interprets positioning payloads. Exact duplicate
envelopes are idempotent; conflicting sequence reuse and normal progression
after a terminal message fail closed. This is deliberately a bounded procedure
control milestone, not LPP transport or positioning support.

**Transaction-control decision: A.** The next phase may design a bounded
procedure orchestration layer that constructs supported messages and consumes
transaction results. Transport, subscriber context, and positioning methods
remain outside scope.

## Bounded procedure orchestration phase

`internal/lpp/procedure` provides peer-scoped, deterministic application
intents for the empty capability and location-information envelopes. It emits
transport-neutral message actions and immutable procedure events, delegates all
transaction mechanics to `internal/lpp/transaction`, and keeps only bounded
pending application-result state. It has no encoding, transport, timer,
goroutine, automatic response, or positioning logic.

**Procedure orchestration decision: A.** The bounded control flow is ready for
the next standards-driven payload phase: Release 9 RequestCapabilities and
ProvideCapabilities IEs, with fresh compiler-backed fixtures.

## Capability payload analysis phase

The first payload analysis selects only the R9 ECID capability request selector
and `ecid-MeasSupported BIT STRING (SIZE(1..8))` provide field. The normalized
ASN.1 graph, pinned asn1tools UPER compiler, and five independent fixtures are
recorded under `tools/specs/lpp/analysis/r16.4.0/capabilities/` and
`tools/specs/lpp/fixtures/r16.4.0/capabilities/`.

**Capability analysis decision: A, conditional on the first missing UPER
primitive.** The selected standards closure is complete and compiler-validated;
the next runtime phase must first add fixture-backed bounded variable-length BIT
STRING support, then implement only the selected ECID capability fields.

## Bounded ECID capability codec phase

`internal/lpp/capability` now implements only the ECID R9 request selector and
the `ecid-MeasSupported SIZE(1..8)` named BIT STRING provide field. It uses
schema-neutral `internal/uper` BIT STRING operations, preserves all empty R9
wrapper encodings, and rejects every other root family and all extensions.

**ECID capability codec decision: A.** The next phase may carry these typed
values through the transport-neutral procedure API; it must not select or run a
positioning method.
