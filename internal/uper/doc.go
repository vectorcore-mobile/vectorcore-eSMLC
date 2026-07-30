// Package uper implements a bounded, schema-neutral subset of X.691
// unaligned PER (UPER) primitives. It is intentionally not a general ASN.1
// codec and does not implement aligned PER, open extension payloads,
// unconstrained lengths, OCTET STRINGs, or reflection. Bounded BIT STRING
// SIZE(min..max), for 1 <= min <= max <= 64, is supported: a constrained
// length precedes MSB-first content bits. BitString has an explicit meaningful
// length and rejects non-zero storage padding.
// A fixed size is expressed by min == max; that constrained range has one
// value, so it emits zero length bits and exactly the fixed number of content
// bits. Fixed trailing zero bits remain meaningful and are never shortened.
//
// Bounded, non-extensible SEQUENCE OF is also supported through callback-based
// WriteSequenceOf and ReadSequenceOf. The count uses constrained whole-number
// encoding, followed immediately by elements in wire order. The primitive adds
// neither alignment nor internal list allocation. Callback failures preserve
// partial writer/reader progress; callers own element storage and schemas.
//
// Callers supply the exact meaningful bit length when decoding. Bits stored in
// a final transport octet beyond that length are required to be zero and are
// validated by ValidateFinalPadding. The package is pure Go and has no runtime
// dependency on development fixtures or the reference compiler.
package uper
