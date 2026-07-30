# Bounded UPER SEQUENCE OF primitive

## API and X.691 rule

`internal/uper` now exposes:

```go
func (w *Writer) WriteSequenceOf(count, minCount, maxCount int,
    encodeElement func(index int, w *Writer) error) error

func (r *Reader) ReadSequenceOf(minCount, maxCount int,
    decodeElement func(index int, r *Reader) error) (int, error)
```

This is the bounded, non-extensible UPER form of `SEQUENCE (SIZE(min..max))
OF T`: a constrained whole-number count, immediately followed by elements in
ascending index order. It adds no alignment, length determinant, extension
bit, reflection, schema metadata, or internal element allocation.

For `SIZE(1..32)`, the count is represented as the offset `count - 1` in five
bits. Tests verify 1, 2, 16, 31, and 32 as offsets `00000`, `00001`, `01111`,
`11110`, and `11111` respectively.

## Validation and failure contract

Negative minimum bounds and reversed bounds are rejected. Counts below or
above the declared range are rejected. Nil callbacks are accepted only for an
empty permitted list; non-empty lists require their appropriate callback.

Writers encode the count before calling an element callback. On callback
failure the count and earlier elements remain written. Readers retain the bit
position reached by the failing callback; no rollback is attempted. Element
failure errors report the zero-based index and preserve the underlying error
for `errors.Is`. EOF while decoding the count or an element is additionally
classified as truncated count or truncated element.

## Validation coverage and bounds

Tests cover fixed sizes, `SIZE(0..1)`, `SIZE(1..2)`, `SIZE(1..8)`,
`SIZE(1..32)`, `SIZE(0..64)`, synthetic high bound `SIZE(0..1024)`, empty
lists, multi-bit contiguous elements, nested lists, callback failure at
multiple positions, truncation, deterministic re-encoding, and independent
concurrent readers/writers. The primitive itself never allocates an element
slice; callbacks own all element storage.

Fuzz coverage exercises round trips, arbitrary decodes, and callback failure
stopping behavior. Extensible lists, unconstrained/fragmented lengths, and
schema-specific list types remain unsupported.

## Next boundary

Evaluate a schema-neutral fixed-size BIT STRING helper for `SIZE(10)`; do not
implement an ECID provide-side type or codec in that phase.
