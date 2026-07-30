# Authoritative serving-cell ECID milestone

## Selected operational capability

The repository contains ECID primary/neighbor measurements (physical cell ID,
root E-UTRA ARFCN, optional RSRP/RSRQ/Rx-Tx/SFN and CellGlobalId) but no cell
geometry, antenna survey, timing calibration, propagation model, or LPPa/eNB
assistance. RSRP/RSRQ therefore cannot honestly be converted to range and
Rx-Tx cannot be used for timing ranging.

This milestone implements the smallest defensible real estimate: an
operator-supplied authoritative serving-cell reference point with an
operator-supplied conservative GAD uncertainty code. It uses the serving ECGI
from the LCS request and confirms that a real ECID result was received. It does
not claim the UE is at the cell coordinate, infer azimuth, derive altitude, or
use radio level as distance.

## Authoritative data boundary

`internal/positioning` loads a versioned YAML catalog once at server creation.
Each record requires exact fourteen-hex-character ECGI, latitude, longitude,
coverage uncertainty code, non-empty provenance source, and RFC3339 update
time. The loader rejects malformed identifiers, invalid coordinates, duplicate
records, future values, stale records, and missing provenance. The immutable
in-memory map is the consistent snapshot used by every active job; no reload
can change an in-flight estimate.

```yaml
version: operator-survey-2026-07
cells:
  - ecgi: "00f11000000001"
    latitude: 38.0
    longitude: -90.0
    coverage_uncertainty_code: 40
    source: operator-survey
    updated_at: "2026-07-01T00:00:00Z"
```

Configuration uses `positioning.ecid.cell_data_file` plus explicit
`cell_data_max_age`. If loading fails, a typed insufficient-network-data
estimator is installed and no coordinate is emitted. Simulation is selected
only when no authoritative catalog is configured and remains explicitly marked
as simulation.

## Estimate and QoS semantics

`ServingCellEstimator` produces `EstimateSourceAuthoritativeServingCell` with
the operator-supplied conservative uncertainty code. Existing horizontal QoS
comparison accepts or rejects the coarse estimate. Vertical QoS remains
unsatisfied because no altitude is produced. `Positioning-Data`, Accuracy
Fulfilment Indicator, confidence, velocity, sector geometry, and detailed
LCS-Cause alternatives remain deferred because this estimator has no
standards-justified values for them.

## Verification

Focused catalog tests cover a valid record, exact ECGI lookup, authoritative
coordinates and uncertainty, missing data, stale records, duplicate records,
and invalid coordinates. A correlated SLs/LPP/LCS integration test drives a
real ECID provide message through a configured catalog and verifies a
successful `LocationRequest` response. The same workflow remains covered for
the explicit simulation source.

`go test ./...`, `go test -race ./...`, `go vet ./...`, normal and
`CGO_ENABLED=0` builds, and the repository `make test`, `make vet`, and
`make build` targets pass. The normalized ASN.1 and existing fixture artifacts
are unchanged; production code reads only configured operator data, never the
analysis fixtures or ASN.1 compiler tooling.

## Decision

**Decision A.** The E-SMLC can now turn a real received ECID measurement into
a coarse but authoritative, provenance-backed serving-cell geographic estimate
when an operator supplies fresh surveyed data. Missing or stale data produces
a typed failure instead of fabricated geography.

The next coherent goal is standards-backed result enrichment and operations:
full LCS response semantics (Accuracy Fulfilment Indicator, Positioning-Data,
and detailed LCS-Cause/diagnostics) together with explicit catalog reload
policy. A more accurate ECID estimator remains blocked on surveyed sector/
antenna data and a validated model, while timing-based work requires verified
LPPa/eNB calibration.
