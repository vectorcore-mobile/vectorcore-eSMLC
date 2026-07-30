# ECID location-information analysis

## Paths and inventories

Both outer procedures select `criticalExtensions.c1[0]` and the R9 alternative
at index 0. Their R9 IE SEQUENCE values have root optional fields in order:
common, A-GNSS, OTDOA, ECID, EPDU, followed by R13/R16 extension additions.
Only ECID is in scope; all other root fields and extensions must fail closed.

`ECID-RequestLocationInformation` has mandatory
`requestedMeasurements BIT STRING { rsrpReq(0), rsrqReq(1), ueRxTxReq(2),
nrsrpReq-r14(3), nrsrqReq-r14(4) } (SIZE(1..8))`, then an extension marker.
The bounded initial request can therefore use existing optional-bitmap,
extension-bit, and bounded variable BIT STRING primitives.

`ECID-ProvideLocationInformation` has optional signal measurement information
and optional ECID Error. Signal information has optional primary-cell results
and mandatory `MeasuredResultsList SIZE(1..32)`. Each list item mandates
`physCellId INTEGER(0..503)` and E-UTRA ARFCN, and may include cell global ID,
10-bit system frame number, RSRP `(0..97)`, RSRQ `(0..34)`, and UE Rx-Tx
`(0..4095)`. Later extensions add NB/NR measurements.

## Dependency and primitive boundary

The initial request graph is bounded and needs no new primitive. Full ECID
provide requires a SIZE-constrained SEQUENCE OF (1..32), fixed BIT STRING
SIZE(10), nested CellGlobalIdEUTRA-AndUTRA/PLMN digit lists, and error
CHOICE/ENUMERATED extension handling. No open type is required by the selected
root path, but those primitives and negative fixtures must be established before
a response codec is implemented.

The maximum response list is 32 entries; the proposed first request codec has
no list allocation and at most an eight-bit request bitmap. A future response
codec must set explicit list and nested allocation limits and defer cell global
ID and Error until their fixture-backed primitives exist.

## Independent fixtures

The pinned `asn1tools 0.167.0` UPER compiler round-tripped the fixtures in
`tools/specs/lpp/fixtures/r16.4.0/ecid-location/`: RSRP request (`0104`, 14
bits), three-root-bit request (`0117`, 16 bits), and representative one-cell
provide (`0120020040190f00`, 57 bits). Negative fixture requirements are
recorded in that manifest but are not yet generated.

## Recommended implementation order

1. Independently validate and implement only R9 ECID request payload.
2. Add fixture-backed bounded SEQUENCE OF and fixed BIT STRING primitives.
3. Analyze/implement one-cell measurement result without CellGlobalId/Error.
4. Add list support and then failure containers only after additional fixtures.

Recommended future model boundary is `internal/lpp/location`, with separate
request, measurement, and error types; no procedure, transport, or positioning
policy belongs there.

## Decision

**Decision B for the complete ECID location-information path.** The request
side is well bounded, but the complete response closure and negative fixture
matrix are not ready. The next handoff may implement only the R9 ECID request
codec after treating its two compiler fixtures as the oracle; it must not begin
ProvideLocationInformation.
