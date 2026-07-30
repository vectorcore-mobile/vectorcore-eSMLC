# Method result, estimation, and LCS outcome milestone

## Audit and selected capability

The recovered TS 29.171 Release 16.4.0 Location Request supplies mandatory
Correlation-ID, Location-Type, and E-UTRAN Cell Identifier. The implemented
subset now also preserves the conditional one-octet LCS-Priority IE. It does
not yet decode LCS-QoS, UE Positioning Capability, IMSI/IMEI, service type,
velocity, RAT, coverage, or APDU request semantics; their absence is retained
as unavailable rather than replaced with policy defaults.

The repository contains ECID radio measurements but no authoritative cell
coordinates, antenna data, timing calibration, propagation model, or topology
database. It therefore cannot derive a real ECID location estimate. The
selected milestone is an honest complete result boundary: ECID measurements
become typed method output, then either an estimator result or a typed failure,
and finally one correlated LCS outcome.

## Models and workflow

`internal/positioning` now separates `RawECIDMeasurements`, `MethodResult`,
`GeographicEstimate`, `EstimationResult`, and terminal `FinalOutcome`.
Measurements are never coordinates. An estimator receives only the scoped job
request and its verified method result. With no estimator installed, a valid
ECID report terminates as `FinalMeasurementsWithoutEstimator` with
`EstimatorUnavailable`.

The only current estimator is `SimulationEstimator`. It is installed only when
`positioning.simulation.enabled` is true and labels every result
`EstimateSourceSimulation`; it does not assert A-GNSS, OTDOA, or real ECID
support. It is invoked only after the normal policy-gated ECID capability and
measurement workflow. A configured legacy simulation failure is converted to a
typed estimation failure; asynchronous simulation delay is rejected rather than
silently bypassing the job lifecycle.

```text
LCS request -> positioning job -> UE ECID capability -> ECID measurements
  -> raw ECID method result -> estimator or EstimatorUnavailable
  -> one correlated LCS Location Response
```

An estimate uses the recovered LCS-AP geographic point-with-uncertainty codec.
All non-estimate terminal outcomes use the existing, standards root
radio-network-layer/unspecified LCS Cause representation. Detailed LCS-Cause
alternatives are not claimed until the full LCS-Cause CHOICE codec is added.

## Safeguards and unsupported scope

Final delivery is issued only from a terminal job outcome; terminal jobs are
removed before late events can complete them. Cancellation, deadline expiry,
reset, Location Abort, and association loss remain fail-closed. There is no
real ECID estimator, QoS acceptance calculation, cell database, LPPa, OTDOA,
A-GNSS, velocity estimate, positioning data, or full LCS Cause codec.

## Validation and decision

Focused tests cover priority preservation, raw ECID to
estimator-unavailable conversion, explicitly labelled simulation estimate
delivery, no-method delivery, reset cleanup, and correlation. Focused race,
bounded fuzz smoke, project-wide tests and race tests, vet, normal and pure-Go
builds, and Make targets are run for this milestone.

**Decision A.** A received method result now reaches an honest typed final
outcome and a correlated LCS success (explicit simulation only) or failure.
The next coherent goal is the standards-backed LCS request/QoS and full
LCS-Cause/Positioning-Data codec closure needed for defensible production
eligibility and outcome causes.
