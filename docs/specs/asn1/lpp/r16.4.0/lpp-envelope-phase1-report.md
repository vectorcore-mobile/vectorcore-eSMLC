# Bounded LPP outer envelope — phase 1

Source: `LPP-PDU-Definitions.asn1` V16.4.0, lines 7–265.

`internal/lpp` maps `LPP-Message` in normative field order: optional bitmap
`transactionID, sequenceNumber, acknowledgement, lpp-MessageBody`; mandatory
`endTransaction`; then present optional fields. `TransactionNumber` and
`SequenceNumber` are constrained `0..255`. `Initiator` root indexes are
`locationServer=0` and `targetDevice=1`.

Supported body indexes under `LPP-MessageBody.c1` are RequestCapabilities 0,
ProvideCapabilities 1, RequestLocationInformation 4,
ProvideLocationInformation 5, Abort 6, and Error 7. The four critical
extension wrappers use root `c1=0`, R9 selector `0`, extension-present `0`,
and an all-false root optional bitmap (five bits; Abort has one). Error uses
root `error-r9=0`, extension-present `0`, and one false optional bit.

All other body alternatives, critical-extension branches, extension additions,
and any present R9 payload IE fail closed. Fixture coverage is the eleven
committed reference fixtures. No positioning data, procedure state, transport,
or transaction ownership is implemented.
