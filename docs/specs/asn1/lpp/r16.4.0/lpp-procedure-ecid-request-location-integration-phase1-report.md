# ECID RequestLocationInformation procedure integration

## Boundary

`internal/lpp/procedure` accepts and returns typed
`internal/lpp/location` request values; it does not encode ASN.1 or interpret
the requested-measurement bitmap. `internal/lpp/location` validates and owns
the payload. `internal/lpp/transaction` owns duplicate fingerprints,
acknowledgements, sequence state, and terminal transitions.

## Supported path and API

The sole added typed procedure path is:

```
RequestLocationInformation
  criticalExtensions.c1[0]
  requestLocationInformation-r9[0]
  RequestLocationInformation-r9-IEs.ecid-RequestLocationInformation
  requestedMeasurements BIT STRING (SIZE(1..8))
```

`StartLocationInformation(StartLocationInformationOptions, now)` takes an
optional `*location.ECIDRequestLocationInformation`. A nil value preserves the
existing empty envelope behavior; a non-nil value is validated by the location
package. It returns an initial send action and a typed
`LocationInformationRequested` event. No bits are added, removed, or
canonicalized.

Inbound new ECID requests emit `LocationInformationRequested` followed by
`AwaitingApplicationResult`; acknowledgement-required input prepends
`AcknowledgementRequested` and a send-acknowledgement action. The pending
application entry owns a deep copy of the typed request container. Events and
snapshots receive independent container copies, while `uper.BitString` itself
has immutable value semantics.

## Duplicate and cleanup rules

An exact duplicate emits only `DuplicateIgnored` and creates no extra wait.
Changed requested-measurement bytes or bit length are not exact duplicates and
are rejected while a request is pending. Abort, Error, and explicit pruning
remove pending location work. No location result API, placeholder result type,
automatic response, method selection, or measurement execution was added.

## Fixture coverage and bounds

Procedure tests decode the committed request fixtures through the location
codec, then submit the resulting typed envelope:

| Fixture | Bytes/bits | Preserved typed value |
| --- | --- | --- |
| request-ecid-rsrp | `0104` / 14 | one-bit RSRP request |
| request-ecid-all-root | `0117` / 16 | three-bit RSRP/RSRQ/UE Rx-Tx request |

Outbound test actions re-encode their payload values to these exact fixture
bytes and meaningful lengths. Each pending request retains at most one
immutable eight-bit bitmap plus its small typed container. There are no lists,
encoded-payload blobs, timers, goroutines, or unbounded allocations.

## Next boundary

Return to the blocked provide-side primitive work: add schema-neutral bounded
`SEQUENCE OF SIZE(1..32)` support in `internal/uper`, with no LPP or ECID
semantics.
