# LCS request constraints and outcome semantics milestone

## Audited Release 16 semantics

TS 29.171 Location Request has mandatory Correlation-ID, Location-Type and
E-UTRAN Cell Identifier, plus conditional LCS-Priority, LCS-QoS and UE
Positioning Capability. This milestone implements the root, non-extension
forms that have immediate consumers: priority; LCS-QoS horizontal accuracy,
vertical request, vertical accuracy and response-time; and the UE `lPP`
capability Boolean. QoS and capability extensions fail closed.

LCS-QoS accuracy values are the specified `0..127` codes. They are not local
metre values. The recovered Location Estimate already carries the compatible
GAD uncertainty code, enabling a conservative supported check: an estimate
satisfies requested horizontal accuracy only when its uncertainty code is no
greater than the requested code. Required vertical information is rejected
because the current estimate model has no altitude. Response-time is preserved
but is QoS-unknown because its low-delay/delay-tolerant enum does not define a
duration for the E-SMLC deadline in this subset.

## Executable changes

An explicit `UE-Positioning-Capability.lPP = FALSE` now prevents ECID/LPP
initiation. A requested horizontal accuracy can reject a simulated estimate;
vertical requirements also reject it. Absent QoS is deliberately not treated
as a successful QoS evaluation. The job retains typed method results and final
outcomes independently of the LCS response mapping.

The supported root LCS-Cause mapping is deterministic:

| Internal final outcome | LCS-Cause |
|---|---|
| Procedure/protocol failure | protocol / unspecified |
| No eligible method, estimator unavailable, estimation failure, QoS failure, deadline, cancellation | misc / unspecified |
| Defensive fallback | radio-network-layer / unspecified |

The implementation keeps the detailed internal final outcome for diagnostics;
it does not expose implementation strings or invent a UE/radio-specific cause.
`Positioning-Data`, velocity, criticality diagnostics, and detailed cause
alternatives are not yet encoded because the current outcome has no
standards-justified values for them.

## Capability and lifecycle boundary

LCS request capability is scoped to the request/correlation. UE ECID
measurement capability remains scoped to the matching LPP transaction and is
discarded with terminal jobs, reset, abort, or association loss. Local policy
permits a method but never stands in for UE capability. Simulation is explicit,
labelled, and only exercises result delivery after ECID measurement exchange.

## Decision

**Decision A.** Request constraints now change executable behavior: LPP can be
excluded before initiation and estimate acceptance can fail QoS checks. The
E-SMLC now emits a deterministic typed LCS outcome for those decisions.

The next highest-value coherent goal is the remaining TS 29.171 request and
response closure that has direct consumers: service/QoS deadline policy,
Positioning-Data and Accuracy Fulfilment Indicator, full LCS-Cause CHOICE, and
Criticality Diagnostics. Real ECID geography remains blocked by authoritative
network/cell data and a validated estimator.
