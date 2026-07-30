# ECID MeasuredResultsElement field analysis

## Decision

**Decision A.** Analysis only; no production code changed. The next bounded
implementation may be only the four independent root optional scalar codecs.

## Path and root declaration

All definitions resolve locally in normalized `LPP-PDU-Definitions.asn1`:

```text
ProvideLocationInformation -> criticalExtensions.c1[0]
-> provideLocationInformation-r9 -> ProvideLocationInformation-r9-IEs
-> ecid-ProvideLocationInformation -> ECID-ProvideLocationInformation
-> ecid-SignalMeasurementInformation -> ECID-SignalMeasurementInformation
-> measuredResultsList -> MeasuredResultsList -> MeasuredResultsElement
```

`MeasuredResultsElement` is the extensible SEQUENCE at lines 3955-3975. Root
order is `physCellId`, optional `cellGlobalId`, `arfcnEUTRA`, optional
`systemFrameNumber`, optional `rsrp-Result`, optional `rsrq-Result`, optional
`ue-RxTxTimeDiff`; there are no other root fields.

| Index | Field | Optional map | ASN.1 constraint | Bits |
| ---: | --- | ---: | --- | ---: |
| 0 | physCellId | — | INTEGER (0..503) | 9 |
| 1 | cellGlobalId | 0 | CellGlobalIdEUTRA-AndUTRA | bounded |
| 2 | arfcnEUTRA | — | INTEGER (0..65535) | 16 |
| 3 | systemFrameNumber | 1 | BIT STRING SIZE(10) | 10 |
| 4 | rsrp-Result | 2 | INTEGER (0..97) | 7 |
| 5 | rsrq-Result | 3 | INTEGER (0..34) | 6 |
| 6 | ue-RxTxTimeDiff | 4 | INTEGER (0..4095) | 12 |

The extension-present bit precedes the five optional bits, ordered exactly as
`cellGlobalId`, `systemFrameNumber`, `rsrp-Result`, `rsrq-Result`, and
`ue-RxTxTimeDiff`. A present `cellGlobalId` is encoded before mandatory ARFCN.

## Independent fixtures

Pinned development-only `asn1tools 0.167.0` encodes, decodes, and re-encodes
the exact local extraction. Mandatory-only PCI 1 / ARFCN 100 is `000200c8`,
31 meaningful bits: `0 | 00000 | 000000001 | 0000000001100100`.

Individual fixtures prove optional headers `0|10000`, `0|01000`, `0|00100`,
`0|00010`, and `0|00001`; all root optionals proves `0|11111`. Harness,
fixture byte hashes, annotations, and source hash are committed under
`tools/specs/lpp/analysis/r16.4.0/ecid-measured-results-element-fields/`.

## Scalar and CellGlobalId closure

RSRP, RSRQ, and UE Rx-Tx are direct non-extensible constrained integers. Zero
is valid, there are no ASN.1 unavailable sentinels, and existing constrained
whole-number UPER is sufficient. `systemFrameNumber` is a root optional fixed
ten-bit BIT STRING and maps to existing `WriteBitString(v,10,10)` /
`ReadBitString(10,10)` without trailing-zero shortening.

`cellGlobalId` is local `CellGlobalIdEUTRA-AndUTRA` (lines 299-309): an
extensible SEQUENCE containing PLMN and a non-extensible CHOICE: `eutra` root
index 0, fixed 28-bit identity; `utra` root index 1, fixed 32-bit identity.
PLMN is not TBCD/NAS/Diameter: MCC is exactly three `INTEGER(0..9)` digits;
MNC is two to three such digits. MNC has one count bit; every digit takes four
bits. Two-/three-digit MNC and leading zeroes are wire-significant. Fixtures
cover `001/01`, `310/260`, and `208/01`.

## Extensions

There are three extension groups: index 0 r9a0 `arfcnEUTRA-v9a0 INTEGER
(65536..262143)`; index 1 r14 NRSRP/NRSRQ/NB-IoT offset/HyperSFN; and index 2
r1470 RSRP/RSRQ alternatives. Extension fixture v9a0=65536 is
`800200c80a01c0000000 / 80 bits`: extension bit 1, normally-small map-length
value 2, bitmap `100`, and three-octet open type containing padded 18-bit
scalar data.

`internal/uper` has extension-present-bit support only. It lacks normally-small
numbers, extension maps, open types, nested open limits, and unknown-extension
skipping. Root-only future coding is nevertheless safe if it encodes false and
rejects true immediately, without skipping or preserving extensions.

## Recommended next boundary

Implement only `RSRPResult`, `RSRQResult`, `UERxTxTimeDiff`, and
`SystemFrameNumber` scalar helpers. Then separately analyze CellGlobalId;
do not implement an element, list, or provide-side envelope yet.
