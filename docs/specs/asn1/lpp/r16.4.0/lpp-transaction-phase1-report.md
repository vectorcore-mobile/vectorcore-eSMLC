# Bounded LPP transaction control, phase 1

`internal/lpp/transaction` is a transport-neutral, caller-scoped in-memory
store above the bounded codec. Its key is `(initiator, transactionNumber)`;
therefore one store must be scoped to one LPP peer, not globally.

| State | Accepted bounded transition |
| --- | --- |
| Created / Active | matching request or provide; optional acknowledgement pending |
| AwaitingAcknowledgement | matching acknowledgement clears pending |
| Active | `endTransaction` -> Completed; Abort -> Aborted; Error -> Failed |
| Completed / Aborted / Failed | exact retransmission is Duplicate; other input is rejected |

Capabilities and location-information requests/provides must travel in the
direction implied by local/remote ownership. Sequence numbers are optional;
when present, the first value is accepted and subsequent values in the same
direction must increment modulo 256. Exact envelope fingerprints are
idempotent duplicates; equal sequence numbers with changed fields conflict.

At most one acknowledgement request can be pending per transaction. An
indicator must match its pending sequence when one was present. Abort and Error
are terminal; Error conservatively marks the transaction Failed. Transactionless
envelopes without a procedure body (including sequence/ack-only envelopes) are
reported but never stored; transactionless procedure bodies
are rejected because a key must not be invented.

The default limits are 128 active and 128 retained terminal records, with a
30-minute active lifetime and 10-minute terminal retention. `Prune(now)` is
explicit; no timer, goroutine, response generation, transport action, or
positioning payload processing exists in this phase.
