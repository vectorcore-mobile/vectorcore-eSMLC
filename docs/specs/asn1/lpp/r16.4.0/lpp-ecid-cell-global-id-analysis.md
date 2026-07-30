# ECID CellGlobalIdEUTRA-AndUTRA analysis

## Decision

**Decision A.** The compact root-only dependency closure is complete; no
production code changed.

## Normalized declaration and wire order

`MeasuredResultsElement.cellGlobalId` resolves locally to
`CellGlobalIdEUTRA-AndUTRA` in `LPP-PDU-Definitions.asn1:299-309`:

```text
SEQUENCE { plmn-Identity SEQUENCE { mcc SIZE(3) OF INTEGER(0..9),
 mnc SIZE(2..3) OF INTEGER(0..9) },
 cellIdentity CHOICE { eutra BIT STRING SIZE(28), utra BIT STRING SIZE(32) }, ... }
```

The wrapper is extensible, has two mandatory root fields, no root optional
bitmap, and no extension additions in the Release 16.4.0 normalized source.
Root wire order is extension-present bit, PLMN, cell-identity CHOICE index,
then identity payload.

MCC is mandatory, not optional: there is no MCC-presence bit and no valid
MCC-absent encoding in this wrapper. MCC is three four-bit constrained digits
(12 bits). MNC is mandatory, contains two or three four-bit digits, and has a
one-bit constrained count first: zero denotes two digits, one denotes three.
Digits are ordinary `INTEGER (0..9)`, never TBCD filler nibbles; zeroes and
digit order are preserved.

The non-extensible CHOICE has exactly two root alternatives: E-UTRA index 0
and fixed 28-bit payload; UTRA index 1 and fixed 32-bit payload. Its index is
one bit. Fixed BIT STRINGs use existing `WriteBitString(v,n,n)` / `ReadBitString`
with zero size-determinant bits, MSB-first ordering, and no alignment.

## Fixtures

Pinned `asn1tools 0.167.0` generated deterministic root-only fixtures under
`tools/specs/lpp/analysis/r16.4.0/ecid-cell-global-id/`. Two-digit MNC E-UTRA
cases are 51 meaningful bits; three-digit cases are 55. UTRA counterparts are
55 and 59 bits. Fixtures cover MCC/MNC `001/01`, `310/260`, zero/one/pattern/
maximum identities, and leading zeros. The harness hash and every encoded
artifact hash are committed in `fixtures.json`.

## Runtime and model recommendation

Existing UPER support is complete for root-only coding: extension bit with
fail-closed true rejection, constrained digits, bounded sequences, fixed bit
strings, and two-root CHOICE. Open-type extension support remains absent but is
not needed because the wrapper currently has no additions. A future codec must
encode extension=false and reject true.

Use a narrow `internal/lpp/location/result` closure, not a generic NAS/PLMN
package: immutable digit value, fixed MCC, length-explicit MNC, PLMN, distinct
28-/32-bit identities, an enum-backed exclusive CHOICE, and wrapper. Preserve
MCC absence explicitly only if a future schema permits it; this wrapper does
not. Optional string constructors may accept only strict ASCII digit strings to
retain leading zeros; parsing is convenience outside codec internals.

Recommended next implementation: one bounded root-only CellGlobalId phase is
safe, provided it includes complete fixture tests and excludes
MeasuredResultsElement, lists, envelopes, extensions, and open types.
