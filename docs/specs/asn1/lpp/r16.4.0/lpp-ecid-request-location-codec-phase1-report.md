# Bounded R9 ECID RequestLocationInformation codec

## Scope and boundary

`internal/lpp/location` owns the request-side R9 payload model and UPER
payload codec. `internal/lpp` owns the outer LPP message-body CHOICE and
delegates only `RequestLocationInformation` to this package. This one-way
dependency (`lpp -> location -> uper`) has no import cycle. No
`ProvideLocationInformation` payload codec is present.

## Normative bounded path

```
RequestLocationInformation
  criticalExtensions.c1[0]
  requestLocationInformation-r9[0]
  RequestLocationInformation-r9-IEs
  ecid-RequestLocationInformation
  requestedMeasurements BIT STRING (SIZE(1..8))
```

Both critical-extension CHOICE indexes are zero. The R9 IE sequence writes an
extension-present bit of zero, then its five root optional bits in this exact
order: common, A-GNSS, OTDOA, ECID, EPDU. Only ECID is accepted. Its sequence
also writes an extension-present bit of zero, followed by the bounded BIT
STRING.

`requestedMeasurements` is a named constrained BIT STRING: bit 0 (MSB-first)
is `rsrpReq`, bit 1 is `rsrqReq`, and bit 2 is `ueRxTxReq`. Bits 3–7 are not
interpreted by this bounded codec, but the original bit length and content are
preserved. This is not canonicalized: a one-bit `1` and a three-bit `100` are
distinct values.

## Fixtures and validation

The independent request fixtures are read by tests only:

| Fixture | Bytes | Meaningful bits | Decoded value |
| --- | --- | ---: | --- |
| request-ecid-rsrp | `0104` | 14 | ECID RSRP bit, length 1 |
| request-ecid-all-root | `0117` | 16 | ECID RSRP/RSRQ/UE Rx-Tx bits, length 3 |

Tests decode, validate final padding, and re-encode both values byte- and
bit-identically. Negative tests reject each unsupported root field, R9
extension presence, an empty mandatory requested-measurements value, malformed
or truncated bit strings through `internal/uper`, and malformed choice/index
inputs through the bounded reader.

## Bounds and unsupported paths

The codec has no lists, recursion, open types, or attacker-sized allocations.
The only payload data is at most eight meaningful bits and one copied byte.
Common, A-GNSS, OTDOA, EPDU, all extension additions, future critical
extensions, and every ProvideLocationInformation structure fail closed or are
unrepresentable. No CellGlobalId, PLMN, measurement result, error, procedure,
transport, timer, or positioning logic is introduced.

## Transaction identity regression

`internal/lpp/transaction` includes ECID request measurement presence, exact
bit length, and content in its bounded comparable duplicate fingerprint.
Thus otherwise identical request envelopes with differing support bits or bit
lengths cannot be classified as exact duplicates.

## Next boundary

Integrate this typed request payload into `internal/lpp/procedure` only after
the codec regression suite passes. Do not begin `ProvideLocationInformation`.
