# Future pure-Go UPER runtime requirements

Evidence source: the compiler-accepted TS 37.355 V16.4.0 PDU module and the
independent fixtures in `tools/specs/lpp/fixtures/r16.4.0`. This is a design
input only; no runtime UPER code is included in this phase.

## Required first

- MSB-first unaligned bit reader and writer, including exact pre-octet-padding
  bit lengths and zero transport padding handling.
- BOOLEAN.
- Constrained whole numbers, demonstrated by `TransactionNumber` and
  `SequenceNumber` `(0..255)`.
- Root ENUMERATED encoding for `Initiator`.
- Root CHOICE indexes for `LPP-MessageBody`, `c1`, and critical-extension
  wrapper choices.
- SEQUENCE optional-field bitmaps for `LPP-Message` and R9 IE containers.
- Root extension-presence bits for extensible SEQUENCE/ENUMERATED/CHOICE forms
  reached by the fixtures, with rejection or bounded skipping of unsupported
  extension payloads.

## Phase 1 implementation traceability

| Primitive | Runtime file | Test coverage | Fixture coverage | Status / limitation |
|---|---|---|---|---|
| Unaligned MSB-first bit I/O | `internal/uper/{reader,writer}.go` | reader/writer and fuzz tests | all 11 | complete for bounded widths |
| Exact bit length and zero padding | `reader.go`, `writer.go` | padding negatives | all 11 | complete; caller supplies meaningful length |
| Constrained whole numbers | `primitives.go` | range/unit/fuzz tests | transaction and sequence boundaries | `uint64`; full `0..MaxUint64` unsupported |
| BOOLEAN | `primitives.go` | unit and fixture tests | all 11 | root BOOLEAN only |
| Root ENUMERATED and CHOICE | `primitives.go` | unit and fixture tests | transaction initiator and wrapper alternatives | root alternatives only |
| Optional bitmaps | `primitives.go` | unit and fixture tests | outer and R9 wrapper optionals | explicit caller-provided order |
| Extension presence | `primitives.go` | extension rejection test | extensible fixture roots with bit `0` | payload additions intentionally rejected |
| Bounded variable-size BIT STRING | `bitstring.go` | 1..8, 16/32/64-bit, copy-safety and fuzz tests | ECID capability fixtures | `SIZE(1..64)` only; no fragmentation or unconstrained form |

The phase-1 bounded `internal/lpp` envelope now consumes these primitives; it
does not duplicate bit, padding, integer, bitmap, or CHOICE handling.

## Required later

- Length determinants and normally-small forms.
- OCTET STRING, constrained strings, SEQUENCE OF and open types.
- Extension additions and extension open-type payload decoding.
- Unconstrained and semi-constrained whole numbers.
- Positioning-method payload containers, GNSS estimates, assistance data, and
  all method-specific branches.

## Not reached by the minimum fixtures

No fixture in this phase exercises length-delimited payloads, open types,
method-specific extension additions, assistance data, coordinates, or actual
positioning semantics. Those must not be treated as implemented by the future
first UPER primitive milestone.
