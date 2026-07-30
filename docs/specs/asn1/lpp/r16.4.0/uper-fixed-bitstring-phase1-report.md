# Fixed-size UPER BIT STRING assessment

## Decision

No fixed-size helper was added. The existing schema-neutral API is the correct
representation of fixed `SIZE(N)`:

```go
w.WriteBitString(value, n, n)
r.ReadBitString(n, n)
```

The shared constrained-count implementation calculates a zero-bit width for a
single possible length. Thus a fixed size emits no length determinant and
writes/reads exactly `n` MSB-first content bits.

## SIZE(10) vectors

| Meaningful bits | Stored bytes | Meaningful length |
| --- | --- | ---: |
| `0000000000` | `0000` | 10 |
| `0000000001` | `0040` | 10 |
| `1000000000` | `8000` | 10 |
| `1010101010` | `aa80` | 10 |
| `1111111111` | `ffc0` | 10 |

The last six bits of the final storage octet are zero transport padding, not
wire content. A fixed ten-bit value such as `1000000000` remains ten bits and
cannot compare equal to the one-bit value `1`; no trailing-zero shortening or
canonicalization occurs.

## Validation

Tests cover fixed sizes 1, 2, 7, 8, 9, 10, 16, 31, 32, and 64; invalid fixed
bounds and mismatched value length; truncation at every ten-bit prefix;
non-zero transport padding; immutability; and non-octet-aligned composition.
For example, a BOOLEAN, ten-bit alternating value, and a two-bit constrained
integer pack as `d550` with a meaningful length of 13 bits.

The runtime remains bounded to 64-bit values. A ten-bit value requires two
bytes of backing storage and exactly ten wire bits. No schema-specific type,
list allocation, reflection, extension behavior, or provide-side codec was
introduced.

## Next boundary

Analyze the constrained scalar closure for one ECID measured-result element:
`physCellId INTEGER(0..503)` and the selected E-UTRA ARFCN type. Do not
implement a measured-result element or ProvideLocationInformation codec.
