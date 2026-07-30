# Positioning job and procedure initiation milestone

## Selected workflow

The recovered SLs Location Service Request directly provides a four-octet LCS
Correlation-ID, one-byte Location-Type, and seven-octet ECGI. Association
identity is supplied by the SLs carrier. It does not provide a UE identity,
requested accuracy, response time, priority, permitted method, reporting mode,
age, UE capability, eNB capability, assistance availability, or LPPa context.

The milestone creates a bounded method-neutral job from those verified inputs.
Unknown information remains unknown; it is never converted into a positive
capability or a request constraint.

## Eligibility and initiation

Local `positioning.ecid` configuration is operator policy only. When it is
disabled, a valid LCS Location Request reaches the concrete `NoEligibleMethod`
outcome and the existing LCS failure response. It does not silently select ECID
or the legacy GNSS simulation.

When policy explicitly enables ECID and names at least one requested root
measurement, a job starts `RequestCapabilities` with the ECID selector. A
received `ProvideCapabilities` must contain an ECID support bitmap containing
every configured requested bit before the job starts
`RequestLocationInformation`. This is the first point at which ECID is
eligible. A missing or insufficient capability produces `NoEligibleMethod` and
the current LCS failure response.

The executable path is:

```text
LCS Location Request
  -> Positioning Job (association + LCS correlation, unique internal ID)
  -> ECID capability request LPP action
  -> LPP-over-SLs carrier
  -> UE ECID capability result
  -> capability check
  -> ECID RequestLocationInformation LPP action
```

LCS correlation, LPP transaction keys, the monotonic internal job ID, and
future LPPa identifiers remain separate. A job records its capability and
location LPP transaction keys and accepts events only for its active key.

## Lifecycle and results

Jobs have caller-supplied deadlines and explicit states for capability wait,
location-information wait, measurements available, no eligible method,
cancellation, expiry, and procedure failure. A cancellation or observed
deadline aborts the active LPP transaction and returns any resulting LPP Abort
action. Terminal outcomes remove their correlation-scoped job immediately, so
late events fail with `ErrNotActive` and terminal jobs do not accumulate.
Association loss, Reset, and Location Abort also remove scoped procedure state.

A received ECID report changes the job state to `MeasurementsAvailable`; it is
not a geographic estimate and does not create a successful LCS Location
Response. The estimator and final response mapping are separate dependencies.

## Dependency map

| Input / dependency | Source | Current status |
|---|---|---|
| Correlation, Location-Type, ECGI | recovered LCS-AP request | used |
| Association scope | SLs | used |
| UE ECID support | LPP ProvideCapabilities | discovered before initiation |
| eNB/LPPa capability | LPPa | unavailable |
| QoS, accuracy, response time, priority, reporting | not in recovered request model | unavailable and not guessed |
| ECID policy / requested measurements | local configuration | explicit operator policy |
| A-GNSS / OTDOA support | LPP branches | unsupported |
| GNSS simulation | local test mode | never treated as A-GNSS support |

## Validation

Tests cover eligible ECID capability discovery and initiation, no-enabled-method
failure, distinct job/LCS/LPP identifiers, duplicate job prevention,
cancellation and late-event rejection, observed-deadline abort, and the full
SLs Location Request → LPP capability request → LPP capability result → LPP
location request carrier sequence. Focused race coverage, project-wide tests,
builds, vet, Make targets and bounded fuzzing are run for the completed
milestone.

## Unsupported boundary and next goal

No A-GNSS, OTDOA, LPPa, real UE/eNB context, QoS selection, assistance data,
measurement acquisition, estimator, or final geographic response is added.
The next highest-value coherent goal is a typed method-result/estimation
boundary that can consume `MeasurementsAvailable` without fabricating a
location, together with standards-backed LCS failure/result cause mapping and
the additional request inputs needed for policy.

## Decision

**Decision A.** A verified inbound request now creates a lifecycle-managed,
method-neutral job with both an explicit no-method result and a capability-gated
ECID initiation path.
