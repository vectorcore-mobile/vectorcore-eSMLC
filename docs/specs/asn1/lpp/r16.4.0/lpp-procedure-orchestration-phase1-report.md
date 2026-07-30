# Bounded LPP procedure orchestration, phase 1

`internal/lpp/procedure` is scoped to one LPP peer and sits above the codec and
transaction store. It returns `Action` values containing validated `lpp.Message`
models and `Event` values; callers encode, send, schedule, and provide results.

| Flow | Orchestrator result |
| --- | --- |
| local request | initial-request action |
| inbound request | request + awaiting-application events |
| explicit application result | response action; completion when configured |
| inbound acknowledgement request | acknowledgement-request event + acknowledgement action |
| Abort / Error | terminal action or terminal event |

Only empty R9 Request/ProvideCapabilities and Request/ProvideLocationInformation
wrappers are constructed. Event names referring to a provided envelope do not
claim capabilities were exchanged or a location was calculated.

The ordering for a new inbound request is acknowledgement-required event and
action (when requested), then request event, then awaiting-application event.
An exact duplicate produces only `DuplicateIgnored`; it does not add another
pending application result. The orchestrator owns this bounded pending-result
map, while transaction identity, sequence progression, duplicate fingerprinting,
ack correlation, and terminal states remain in `internal/lpp/transaction`.

There are no timeout transition APIs in this phase: the caller may explicitly
prune the store, or Abort/ReportError a transaction after its own policy. This
avoids inventing retransmission or LPP procedure timeout semantics. `Prune(now)`
cleans pending application work whose transaction record no longer exists.

No transport, timer, goroutine, payload semantics, automatic response, or
positioning behavior is present. The next boundary is compiler-backed,
fixture-backed Release 9 capability payload selection.
