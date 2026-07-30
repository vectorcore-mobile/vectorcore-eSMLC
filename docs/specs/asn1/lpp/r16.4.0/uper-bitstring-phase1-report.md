# Bounded UPER BIT STRING primitive

`internal/uper/bitstring.go` adds schema-neutral X.691 unaligned-PER support
for `BIT STRING (SIZE(min..max))`, where `1 <= min <= max <= 64`.

`BitString` stores up to eight bytes privately with an explicit meaningful bit
count. `NewBitString` copies input and rejects excess/missing storage and
non-zero bits outside the meaningful count; `Bytes` returns a copy. Equality
compares only meaningful MSB-first bits.

`WriteBitString` writes the constrained length (`min..max`) followed by exactly
that number of MSB-first content bits. `ReadBitString` reverses this operation
and returns independent storage. It does not implement unconstrained,
fragmented, or larger-than-64-bit strings.

The primitive was verified against the independent ECID capability values:
one-bit `80` (RSRP) and three-bit `e0` (RSRP/RSRQ/UE-RxTx), as well as all
lengths 1 through 8 and representative 16, 32, and 64-bit inputs. No LPP
payload, envelope, transaction, or procedure code changed.
