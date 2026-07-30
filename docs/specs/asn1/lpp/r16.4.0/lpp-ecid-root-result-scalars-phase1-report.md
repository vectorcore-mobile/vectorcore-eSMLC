# ECID root measured-result scalars — phase 1

## Scope

`internal/lpp/location/result` implements only the independently validated
Release 9 root scalars needed before a future ECID measured-result element can
be composed. It does not define an element, list, optional measurements,
extensions, or a `ProvideLocationInformation` payload.

| ASN.1 field | Production type | Root constraint | UPER width |
| --- | --- | --- | --- |
| `physCellId` | `PhysicalCellID` | `INTEGER (0..503)` | 9 bits |
| `arfcnEUTRA` | `EUTRAARFCN` | `ARFCN-ValueEUTRA`, `INTEGER (0..65535)` | 16 bits |

The future enclosing `MeasuredResultsElement` owns field order and all
structural semantics. When it composes these helpers, root packing is
`physCellId || arfcnEUTRA`, with no alignment: 25 meaningful bits.

## API and validation

`NewPhysicalCellID(uint16)` accepts `0..503`; `NewEUTRAARFCN(uint32)` accepts
only `0..65535` and validates before converting to `uint16`. This explicitly
prevents the observed reference-tool wraparound for `65536`. Both values have
`Validate`, `EncodeUPER`, and package-level decode helpers. Encoding delegates
only to `internal/uper` constrained whole numbers and wraps both scalar and
underlying UPER errors.

`ARFCN-ValueEUTRA-v9a0 (65536..262143)` is an extension addition of the future
enclosing element, not a scalar extension. It is deliberately not represented.

## Fixture results

The tests consume the committed analysis-only fixture manifest. They reproduce:

| Value | Hex | Meaningful bits |
| --- | --- | --- |
| PCI 0 | `0000` | 9 |
| PCI 1 | `0080` | 9 |
| PCI 503 | `fb80` | 9 |
| ARFCN 0 | `0000` | 16 |
| ARFCN 100 | `0064` | 16 |
| ARFCN 65535 | `ffff` | 16 |
| `{1,100}` | `00803200` | 25 |
| `{503,65535}` | `fbffff80` | 25 |

Tests also cover every truncation prefix, non-octet-aligned composition,
boundaries, value semantics, and bounded fuzzing. The package has constant
memory use and no fixture or tooling access in production code.

## Next boundary

Analyze only the root optional bitmap and remaining extension-adjacent fields
of `MeasuredResultsElement`: `rsrp-Result`, `rsrq-Result`,
`ue-RxTxTimeDiff`, `systemFrameNumber`, and `cellGlobalId`. Do not implement an
element, list, or provide-side envelope before that analysis is independently
validated.
