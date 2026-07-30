# LPP-over-SLs carrier milestone

## Selected milestone

This milestone implements bounded payload-type 0 LPP dispatch through the
existing SLs/LCS-AP Connection-Oriented Information carrier.

It was selected over method selection, QoS modeling, capability expansion,
OTDOA, and A-GNSS because it removes the immediate integration blocker: before
this work the SLs server decoded a valid LCS-AP carrier then returned `LPP is
not implemented`, leaving the completed LPP codec and procedure unreachable
from the intended MME–E-SMLC path. A policy would be speculative because the
current LCS Location Request abstraction does not yet provide the LPP QoS,
permitted-method, or real capability inputs required for a safe choice.

## Evidence and dependency map

TS 29.171 defines Connection-Oriented Information with Correlation-ID,
Payload-Type and APDU. The recovered codec already validates those IEs and
identifies payload type 0 as LPP and type 1 as LPPa. TS 37.355 provides the
unaligned LPP envelope and existing bounded ECID request/provide path.

| Area | Status |
|---|---|
| LCS-AP carrier and correlation | LPP payload type 0 supported |
| LPP octet-to-UPER boundary | Unique final-padding resolution supported |
| LPP transaction/procedure | Now reached from SLs |
| ECID request/provide | Existing root-only branch supported |
| A-GNSS / OTDOA / LPPa | Unsupported and fail closed |
| QoS, permitted methods and policy | Not modeled or inferred |
| LCS Location Request to LPP initiation | Not yet implemented |

## Implementation and new path

`lpp.DecodeMessageOctets` resolves the absent meaningful-bit count at the APDU
octet boundary. It tests the eight final-octet padding lengths and accepts only
one complete valid LPP decode; no candidate or multiple candidates fails
closed.

`internal/sls` now holds a bounded `procedure.Orchestrator` per association
plus LCS Correlation-ID. On an inbound LPP carrier it decodes the APDU, applies
the transport-neutral procedure, encodes every resulting action, and wraps it
in Connection-Oriented Information with the original correlation ID. Association
teardown, reset, and Location Abort discard scoped procedure instances. The LCS
correlation ID remains separate from the LPP transaction ID.

An MME-side test can now send an LCS-AP Connection-Oriented Information PDU
with a padded ECID RequestLocationInformation APDU requesting acknowledgement.
The server returns a correlated LCS-AP PDU carrying the LPP acknowledgement.
The carrier can decode all currently supported LPP bodies, but an inbound ECID
provide requires a prior locally started LPP transaction; creating that job from
an LCS Location Request is deliberately the next milestone.

## Validation and scope

Focused tests cover padded-octet LPP decoding, malformed APDU rejection, and a
complete LCS-AP → LPP ECID request → procedure acknowledgement → LCS-AP path.
Arbitrary octet LPP decoding has bounded fuzz coverage. Existing LPP fixtures,
procedure tests and non-aligned codec tests remain in force.

No runtime ASN.1 tooling, cgo, native codec, LPPa, OTDOA, A-GNSS, open type,
extension support, timer or goroutine in pure codec packages, positioning
estimate, or method-selection policy was added.

## Remaining blockers and next goal

This does not produce a location estimate. The LCS Location Request must first
be represented as a method-neutral positioning job with QoS, permitted methods,
priority, correlation, real UE/eNB/network capabilities and policy inputs. A
selected method then needs measurement acquisition and an estimator.

The next coherent goal is that method-neutral positioning-job boundary from an
LCS Location Request into the LPP procedure. It must use only inputs proven by
the committed TS 29.171 schema and must keep ECID as a candidate, never a
hard-coded default.

## Decision

**Decision A.** LPP-over-SLs carrier dispatch is integrated and validated. It
makes the existing LPP procedure reachable from the intended MME carrier while
retaining fail-closed boundaries for unsupported methods and policy.
