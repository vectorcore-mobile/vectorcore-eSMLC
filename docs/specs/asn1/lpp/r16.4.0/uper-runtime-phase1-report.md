# Pure-Go UPER runtime — phase 1

`internal/uper` is a bounded, schema-neutral X.691 unaligned-PER primitive
package. It is pure Go and has no runtime dependency on the TS 37.355 tooling
or reference compiler.

## Implemented

- MSB-first unaligned bit reader/writer with exact meaningful bit lengths.
- Zero-filled final transport padding and strict decode padding validation.
- `uint64` constrained whole numbers except the unsupported full
  `0..MaxUint64` cardinality.
- BOOLEAN, root ENUMERATED indexes, root CHOICE indexes, optional bitmaps, and
  extension-presence bits.
- Explicit rejection when an extension-addition payload would need decoding.

## Fixture oracle

The test-only fixture schema in `internal/uper/fixtures_test.go` maps the
eleven committed independent UPER fixtures to primitive calls. It is not an
exported LPP model and is excluded from production builds. Each fixture is
decoded with its manifest bit length, semantically checked for the bounded
envelope fields, consumed exactly, and re-encoded byte/bit identically.

## Unsupported

Aligned PER, general length determinants, unconstrained integers, open types,
extension-addition payloads, strings, BIT/OCTET STRING payloads, SEQUENCE OF,
and all LPP positioning semantics remain outside this phase.

## Next boundary

The next phase may implement a bounded `internal/lpp` outer-envelope package
using these primitives only for the 11 fixture-covered structures. It must not
duplicate bit manipulation or claim GNSS positioning support.
