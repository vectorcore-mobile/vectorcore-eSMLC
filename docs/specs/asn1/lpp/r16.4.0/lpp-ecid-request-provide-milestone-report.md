# ECID request/provide LPP milestone

## Audit and selected milestone

The normalized TS 37.355 Release 16.4.0 module has common, A-GNSS, OTDOA,
ECID, EPDU and later radio-method branches. The repository's only completed
method codec is ECID. `internal/config` still exposes a historical simulated
GNSS mode, while `internal/sls` deliberately treats LPP as an opaque APDU;
neither is evidence of a real GNSS LPP implementation. `internal/lpp` and
`internal/lpp/procedure` already provide common outer envelope, transaction,
acknowledgement, duplicate and terminal-state handling.

The shortest meaningful protocol milestone is therefore the complete,
transport-neutral ECID request/provide exchange, not another isolated ASN.1
leaf. It makes the following LPP path usable by a caller that supplies and
consumes measurements:

```
RequestLocationInformation-r9.ecid-RequestLocationInformation
  -> transaction/procedure correlation
  -> ProvideLocationInformation-r9.ecid-ProvideLocationInformation
  -> ECID-SignalMeasurementInformation
  -> optional primaryCellMeasuredResults + MeasuredResultsList (1..32)
  -> MeasuredResultsElement
```

This is a protocol capability, not a positioning engine. It neither chooses
ECID nor calculates a location from ECID measurements.

## Dependency map

| Layer | Shared / method-neutral | ECID-specific |
|---|---|---|
| UPER | bit IO, constrained integers, fixed and bounded BIT STRING, bounded `SEQUENCE OF`, root CHOICE, optional maps, extension bit | none |
| LPP | outer message, transaction ID, acknowledgement, R9 critical extensions | request/provide R9 IE maps |
| Procedure | correlation, duplicate classification, acknowledgement, terminal cleanup, bounded pending application work | typed ECID request and provide events |
| Result payload | none | measured list, signal information and ECID provide container |
| Positioning policy | intentionally absent | intentionally absent |

The absorbed former micro-dependencies are `MeasuredResultsList`,
`ECID-SignalMeasurementInformation`, `ECID-ProvideLocationInformation`, the
R9 provide wrapper, the LPP message body integration, provide-payload
fingerprinting, and procedure delivery.

## Implementation

`internal/lpp/location/provide.go` adds immutable-container APIs for a bounded
root ECID report:

* `NewECIDSignalMeasurementInformation` enforces one through 32 result
  elements, copies the input slice, and retains an optional primary-cell
  result.
* `NewECIDProvideLocationInformation` supports the signal-information branch;
  `ECID-Error` is rejected rather than silently ignored.
* `ProvideLocationInformationR9IEs` encodes and decodes only its ECID root
  optional bit. Common, A-GNSS, OTDOA and EPDU bits fail closed.

The root encoding is exact unaligned PER: R9 critical-extension selectors,
extension-present=false, the five R9 method presence bits, ECID
extension-present=false, the two ECID option bits, signal
extension-present=false, primary-cell presence, constrained list count, and
the pre-existing root `MeasuredResultsElement` encodings. There is no list
allocation driven by an untrusted decoded count beyond the fixed maximum 32,
and no alignment, open type, extension map, or extension skipping.

`internal/lpp/message.go` now carries the typed provide payload. Transaction
fingerprints include the exact bounded encoded LPP message, so distinct ECID
result lists and optional contents cannot be classified as duplicates.
`internal/lpp/procedure` accepts inbound typed provides as
`LocationInformationEnvelopeProvided` events and lets an application complete
an inbound request through `ProvideLocationInformation`. It performs no method
selection or measurement validation beyond the schema-owned value checks.

## Independent fixtures

`tools/specs/lpp/analysis/r16.4.0/ecid-provide-location/generate.py` uses the
pinned development-only `asn1tools` 0.167.0 UPER compiler and the normalized
module. The committed metadata includes these independent payload fixtures:

| Fixture | Meaningful bits | Hex |
|---|---:|---|
| one-rsrp | 57 | `0120020040190f00` |
| list-two-mixed | 130 | `0124000000004000803207f7ffff8bffc0` |
| all-optionals-eutra | 136 | `01200f80400201091a2b38032555862fff` |
| utravariant | 109 | `0120087dc621260891a2b3c7fff8` |
| full LPP message / all-optionals-eutra | 157 | `920f2809007c0200100848d159c0192aac317ff8` |

The tests consume literal independently generated values, not output from the
codec under test. They cover list bounds, primary-cell placement, all optional
fields, E-UTRA and UTRA CellGlobalId variants, every-prefix truncation of a
representative UTRA payload, non-aligned composition, outer-message
round-trip, procedure response delivery, and duplicate-safe payload changes.

## Scope and next goal

Supported: transport-neutral ECID request/provide LPP procedure data flow.
Unsupported: method selection policy, UE/eNB capability collection beyond the
bounded ECID capability branch, ECID measurement execution, location estimate
calculation, result ranking, `ECID-Error`, extensions, A-GNSS, OTDOA, LPPa,
SLs/LCS-AP LPP dispatch, and all live positioning transport.

The next coherent goal is not another ECID leaf: define the method-neutral
positioning-request and capability input boundary, then connect the existing
LPP procedure actions/events to the SLs LPP APDU carrier. That goal must
establish QoS, permitted-method, UE/eNB/network-capability and operator-policy
inputs before any selection policy is implemented. It can then add ECID as one
candidate rather than making it a default.

## Decision

**Decision A — transport-neutral ECID LPP request/provide protocol milestone
complete**, subject to the validation results recorded with this change. The
repository can now encode, decode, correlate, deliver and respond with bounded
root ECID measurement reports; it still cannot derive or transport a real
position end to end.
