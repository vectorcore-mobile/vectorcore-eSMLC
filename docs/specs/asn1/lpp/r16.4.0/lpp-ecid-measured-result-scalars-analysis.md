# ECID measured-result mandatory scalar closure

## Decision

**Decision A for the root scalar closure.** The two mandatory Release 9 root
scalars are fully resolved, use only the existing bounded constrained-whole-
number primitive, and have deterministic compiler fixtures. This does not make
the enclosing `MeasuredResultsElement` or `ProvideLocationInformation`
implementable: its extension additions, optional fields, and other dependencies
remain outside this analysis.

## Exact normalized ASN.1 path

All relevant declarations are in `LPP-PDU-Definitions`; no module import or
alias is traversed for either root scalar.

```
ProvideLocationInformation (lines 210–218)
  criticalExtensions.c1[0].provideLocationInformation-r9[0]
  ProvideLocationInformation-r9-IEs (lines 219–240)
  ecid-ProvideLocationInformation
  ECID-ProvideLocationInformation (lines 3943–3947)
  ecid-SignalMeasurementInformation
  ECID-SignalMeasurementInformation (lines 3949–3953)
  measuredResultsList MeasuredResultsList (line 3954)
  MeasuredResultsElement (lines 3955–3973)
```

`MeasuredResultsElement` puts mandatory `physCellId` first and mandatory
`arfcnEUTRA` second after its root optional-field bitmap. The scalar-only
harness intentionally excludes that bitmap and all other element fields, so it
proves scalar widths and their direct consecutive field order—not the complete
element envelope.

## Root scalar declarations

| Field | Exact declaration | Root range | Width | Extension status |
| --- | --- | --- | ---: | --- |
| `physCellId` | `INTEGER (0..503)` at line 3956 | 0..503 | 9 bits | direct non-extensible INTEGER |
| `arfcnEUTRA` | `ARFCN-ValueEUTRA` at line 3958; declared `INTEGER (0..maxEARFCN)` at line 280 | 0..65535 (`maxEARFCN`, line 5079) | 16 bits | scalar non-extensible |

For physical cell ID, cardinality is 504 and `ceil(log2(504)) = 9`; offsets
504 through 511 are invalid. For root E-UTRA ARFCN, cardinality is 65536 and
the width is exactly 16; every 16-bit offset is valid.

Existing UPER calls are sufficient:

```go
w.WriteConstrainedWholeNumber(value, 0, 503)
r.ReadConstrainedWholeNumber(0, 503)
w.WriteConstrainedWholeNumber(value, 0, 65535)
r.ReadConstrainedWholeNumber(0, 65535)
```

## Extension boundary

`ARFCN-ValueEUTRA-v9a0 ::= INTEGER (maxEARFCN-Plus1..maxEARFCN2)` is separately
declared at line 281 as 65536..262143. It is not an extension of
`ARFCN-ValueEUTRA`. Instead, line 3964 makes `arfcnEUTRA-v9a0` an optional
extension addition of the enclosing extensible `MeasuredResultsElement`.
Consequently a root-only scalar implementation must reject values above 65535;
it must not emit a scalar extension bit. A later complete element needs the
SEQUENCE extension-addition bitmap, normally-small number, and open-type
handling before it can represent the v9a0 addition.

## Independent fixtures

The pinned `asn1tools 0.167.0` UPER compiler generated and decoded the
deterministic scalar harness fixtures in
`tools/specs/lpp/analysis/r16.4.0/ecid-measured-result-scalars/fixtures.json`.

| Type/value | Hex | Meaningful bits |
| --- | --- | ---: |
| PhysicalCellID 0 | `0000` | 9 |
| PhysicalCellID 1 | `0080` | 9 |
| PhysicalCellID 503 | `fb80` | 9 |
| Root ARFCN 0 | `0000` | 16 |
| Root ARFCN 100 | `0064` | 16 |
| Root ARFCN 65535 | `ffff` | 16 |
| v9a0 ARFCN 65536 (analysis only) | `000000` | 18 |
| Pair `{1,100}` | `00803200` | 25 |
| Pair `{503,65535}` | `fbffff80` | 25 |

The pair vectors prove mandatory scalar order is `physCellId` then
`arfcnEUTRA`, with no alignment between the 9- and 16-bit values.

The pinned compiler silently wraps invalid structured root input `65536` to
the root encoding for zero. This is recorded in `compiler-observations.json`.
It is a compiler validation limitation for invalid input, not an ASN.1
extension behavior; future Go validation must enforce bounds before encoding.

## Safe future Go model

The minimum future root-only model is two validated named scalar values:

```go
type PhysicalCellID uint16       // valid 0..503
type EUTRAARFCNRoot uint16       // valid 0..65535
```

Zero is valid for both. Constructors/validators should enforce the bounds;
the ARFCN name deliberately states the root-only scope. No type, validator, or
codec was added in this analysis phase.

## Next boundary

Implement only those two root-only scalar values and their codec helpers in a
bounded location-result subpackage, using the committed scalar fixtures. Do
not add a measured-result SEQUENCE, list, optional measurement, CellGlobalId,
or ProvideLocationInformation codec.
