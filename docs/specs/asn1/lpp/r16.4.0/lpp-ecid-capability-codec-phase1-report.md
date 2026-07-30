# Bounded ECID capability codec, phase 1

Architecture: `internal/lpp/capability` depends only on `internal/uper` and
owns the capability critical-extension/R9 IE path; `internal/lpp` owns the
outer message-body CHOICE and imports capability. There is no import cycle.

The exact Request path is `criticalExtensions.c1[0]` then
`requestCapabilities-r9[0]`; the Provide path is the corresponding
`provideCapabilities-r9[0]`. Both R9 IE SEQUENCE values write an extension bit
then a five-bit root optional bitmap in this order: common, A-GNSS, OTDOA,
ECID, EPDU. Only bit 3 is accepted. All extension additions, future critical
extensions, common, A-GNSS, OTDOA, and EPDU fields fail closed.

`RequestCapabilitiesR9IEs` represents ECID selector presence with an explicit
empty `ECIDRequestCapabilities`. `ProvideCapabilitiesR9IEs` represents
`ECIDProvideCapabilities.MeasurementSupport` as `uper.BitString SIZE(1..8)`.
The ASN.1 named positions are RSRP=0, RSRQ=1, UE Rx-Tx=2 (MSB-first). It is a
named BIT STRING with a size constraint; the codec preserves the exact bit
length rather than removing trailing zero named positions. Thus 1-bit `1` and
3-bit `100` remain distinct values and encodings.

The five independent top-level fixtures decode and re-encode exactly:
`0000/9`, `0100/10`, `0000/9`, `0104/14`, and `0117/16`. Empty capability
wrappers remain byte-identical in the outer-envelope fixtures. Transaction
fingerprints now include the selected ECID bit length and content so distinct
capability values cannot collapse into duplicates.

Resource bounds are constant: one eight-byte BitString maximum, five optional
bits, no lists, recursion, open types, or allocation from decoded lengths.
Procedure integration and method selection remain explicitly absent.
