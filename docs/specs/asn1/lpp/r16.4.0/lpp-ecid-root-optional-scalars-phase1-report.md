# ECID root optional result scalars — phase 1

`internal/lpp/location/result` now provides only standalone root optional
scalar helpers: `RSRPResult` (`0..97`, 7 bits), `RSRQResult` (`0..34`, 6 bits),
`UERxTxTimeDiff` (`0..4095`, 12 bits), and immutable `SystemFrameNumber`
(`BIT STRING SIZE(10)`, 10 bits). Constructors validate before narrowing;
encode/decode delegates solely to existing constrained whole-number and fixed
BIT STRING UPER primitives and preserves their errors.

System frame numbers use MSB-first ten-bit numeric conversion, preserve
leading/trailing zeros, store immutable `uper.BitString` data, and expose only
value copies. Focused fixtures cover min/mid/max numeric scalars and 10-bit
zero, one, high-bit, alternating, and all-one SFN values. Tests cover all
truncation prefixes and non-aligned composition; fuzz targets cover each
round-trip and arbitrary decoding.

No optional bitmap, MeasuredResultsElement, CellGlobalId, list, extension/open
type, provide-side envelope, procedure, transport, timer, or tooling runtime
dependency was added. Next boundary: analyze CellGlobalId prerequisites before
any aggregate measured-result structure.
