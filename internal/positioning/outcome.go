package positioning

import (
	"fmt"
	"math"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp/location"
	"github.com/vectorcore/esmlc/internal/lpp/location/result"
)

// Method identifies a standards positioning method, independently of its wire
// encoding and of whether an estimate can be derived from its measurements.
type Method uint8

const (
	MethodECID Method = iota + 1
	MethodOTDOA
	MethodAGNSS
	MethodLPPaECID
)

// RawECIDMeasurements preserves ECID radio measurements as method output. It
// deliberately contains no inferred position.
type RawECIDMeasurements struct {
	Primary *result.MeasuredResultsElement
	Results []result.MeasuredResultsElement
}

// RawOTDOAMeasurements preserves the RSTD reference-cell and neighbour-cell
// measurement report as method output. It deliberately contains no computed
// position; converting RSTD values into a geographic estimate is the
// estimator boundary's job (see Estimator), not this package's decode path.
type RawOTDOAMeasurements struct {
	Signal location.OTDOASignalMeasurementInformation
}

// RawAGNSSMeasurements preserves the UE's own already-computed position
// report as method output. Unlike ECID/OTDOA this is not raw measurement
// data needing further estimation — MS-based (UE-based) A-GNSS means the
// target device solved its own fix and reported it via the common
// LocationCoordinates IE (see internal/lpp/location's CommonProvideLocationInformation);
// there is no E-SMLC-side computation for this implementation to do beyond
// validating and relaying it.
type RawAGNSSMeasurements struct {
	Estimate location.LocationCoordinates
}

// MethodResult records a completed protocol-method attempt. Measurements may
// be available even when no geographic estimate can be calculated.
type MethodResult struct {
	Method      Method
	ECID        *RawECIDMeasurements
	OTDOA       *RawOTDOAMeasurements
	AGNSS       *RawAGNSSMeasurements
	LPPaECID    *RawLPPaECIDMeasurement
	CompletedAt time.Time
}

// EstimateSource identifies how an estimate was derived. Simulation is never
// represented as a real A-GNSS or ECID estimate.
type EstimateSource uint8

const (
	EstimateSourceSimulation EstimateSource = iota + 1
	EstimateSourceAuthoritativeServingCell
	EstimateSourceOTDOAMultilateration
	EstimateSourceAGNSSUEReported
	EstimateSourceLPPaAccessPointPosition
)

// GeographicEstimate is the bounded point-with-uncertainty representation the
// recovered LCS-AP codec can deliver. HorizontalUncertainty is the 3GPP GAD
// uncertainty code, not an invented metre value.
type GeographicEstimate struct {
	Latitude              float64
	Longitude             float64
	HorizontalUncertainty uint8
	Source                EstimateSource
	Timestamp             time.Time
	// CatalogVersion, DataSource, and RecordUpdatedAt are internal provenance
	// only. They are deliberately not overloaded into LCS-AP result IEs.
	CatalogVersion  string
	DataSource      string
	RecordUpdatedAt time.Time
}

func (v GeographicEstimate) Validate() error {
	if math.IsNaN(v.Latitude) || math.IsNaN(v.Longitude) || math.IsInf(v.Latitude, 0) || math.IsInf(v.Longitude, 0) || v.Latitude < -90 || v.Latitude > 90 || v.Longitude < -180 || v.Longitude > 180 {
		return fmt.Errorf("positioning: invalid geographic estimate")
	}
	if v.Source == 0 || v.Timestamp.IsZero() {
		return fmt.Errorf("positioning: estimate source and timestamp are required")
	}
	return nil
}

type EstimationFailure uint8

const (
	EstimatorUnavailable EstimationFailure = iota + 1
	InsufficientNetworkData
	RequestedQualityNotMet
)

type EstimationResult struct {
	Estimate *GeographicEstimate
	Failure  EstimationFailure
}

// AccuracyFulfilment retains the result of the supported QoS evaluation even
// when the current recovered LCS-AP subset cannot yet encode the corresponding
// Accuracy-Fulfilment-Indicator IE.
type AccuracyFulfilment uint8

const (
	AccuracyUnevaluated AccuracyFulfilment = iota
	AccuracyFulfilled
	AccuracyNotFulfilled
)

// Estimator is a synchronous, job-scoped boundary. Implementations receive
// only the verified request and method output belonging to that job.
type Estimator interface {
	Estimate(Request, MethodResult, time.Time) EstimationResult
}

// CombinedEstimator dispatches to the estimator matching the job's completed
// method, so ECID and OTDOA can share one Manager while using different
// estimation logic (and, typically, the same underlying operator catalog).
// A nil branch for the completed method's kind is EstimatorUnavailable, the
// same failure a nil Manager.estimator produces for every method.
type CombinedEstimator struct {
	ECID     Estimator
	OTDOA    Estimator
	AGNSS    Estimator
	LPPaECID Estimator
}

func (e CombinedEstimator) Estimate(request Request, method MethodResult, now time.Time) EstimationResult {
	var next Estimator
	switch method.Method {
	case MethodECID:
		next = e.ECID
	case MethodOTDOA:
		next = e.OTDOA
	case MethodAGNSS:
		next = e.AGNSS
	case MethodLPPaECID:
		next = e.LPPaECID
	}
	if next == nil {
		return EstimationResult{Failure: EstimatorUnavailable}
	}
	return next.Estimate(request, method, now)
}

// AGNSSEstimator validates and relays the UE's own already-computed
// position (MS-based/UE-based A-GNSS): unlike ECID's catalog lookup or
// OTDOA's multilateration solve, there is no E-SMLC-side computation here.
// It needs no configuration and does not depend on a cell catalog.
type AGNSSEstimator struct{}

func (AGNSSEstimator) Estimate(_ Request, method MethodResult, now time.Time) EstimationResult {
	if method.Method != MethodAGNSS || method.AGNSS == nil {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	reported := method.AGNSS.Estimate
	if err := reported.Validate(); err != nil {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	// GeographicEstimate carries a single GAD uncertainty-code scalar; for the
	// altitude/ellipsoid shape, report the larger of the two semi-axes so the
	// value stays a conservative (never-too-small) bound.
	horizontalUncertainty := reported.UncertaintyCircle
	if reported.Shape == location.ShapePointWithAltitudeAndUncertaintyEllipsoid {
		horizontalUncertainty = reported.UncertaintySemiMajor
		if reported.UncertaintySemiMinor > horizontalUncertainty {
			horizontalUncertainty = reported.UncertaintySemiMinor
		}
	}
	estimate := GeographicEstimate{
		Latitude:              reported.Point.Latitude,
		Longitude:             reported.Point.Longitude,
		HorizontalUncertainty: horizontalUncertainty,
		Source:                EstimateSourceAGNSSUEReported,
		Timestamp:             now,
	}
	if err := estimate.Validate(); err != nil {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	return EstimationResult{Estimate: &estimate}
}

// SimulationEstimator is an explicitly labelled development estimator. It is
// only installed when simulation is enabled in local configuration.
type SimulationEstimator struct {
	Latitude, Longitude float64
	Uncertainty         uint8
	Failure             EstimationFailure
}

func (s SimulationEstimator) Estimate(_ Request, _ MethodResult, now time.Time) EstimationResult {
	if s.Failure != 0 {
		return EstimationResult{Failure: s.Failure}
	}
	v := GeographicEstimate{Latitude: s.Latitude, Longitude: s.Longitude, HorizontalUncertainty: s.Uncertainty, Source: EstimateSourceSimulation, Timestamp: now}
	if err := v.Validate(); err != nil {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	return EstimationResult{Estimate: &v}
}

type FinalKind uint8

const (
	FinalEstimateAvailable FinalKind = iota + 1
	FinalMeasurementsWithoutEstimator
	FinalEstimationFailed
	FinalQualityNotMet
	FinalNoEligibleMethod
	// FinalLPPUnsupported is distinct from FinalNoEligibleMethod: it means the
	// UE-Positioning-Capability IE on the Location Request itself said the UE
	// has no LPP support at all, so the job never even attempts a capability
	// exchange — as opposed to FinalNoEligibleMethod, which covers both "no
	// positioning method is enabled in local policy" and "the UE's actually
	// reported capabilities don't support the policy-selected method" (a
	// mismatch only discoverable after a capability exchange). Kept separate
	// so an operator can distinguish "this UE can't do LPP" from those other
	// causes directly from the outcome, without inferring it from the
	// absence of capability-exchange log lines.
	FinalLPPUnsupported
	FinalProcedureFailure
	FinalDeadlineExpired
	FinalCancelled
)

// FinalOutcome is returned once, at terminal job transition, and is the only
// object the SLs boundary may translate into an LCS response.
type FinalOutcome struct {
	Kind         FinalKind
	MethodResult *MethodResult
	Estimate     *GeographicEstimate
	Failure      EstimationFailure
	Accuracy     AccuracyFulfilment
}

func rawECID(v *location.ProvideLocationInformationR9IEs) (*RawECIDMeasurements, bool) {
	if v == nil || v.ECID == nil {
		return nil, false
	}
	signal, ok := v.ECID.SignalMeasurementInformation()
	if !ok {
		return nil, false
	}
	out := &RawECIDMeasurements{Results: signal.MeasuredResults()}
	if primary, ok := signal.PrimaryCellMeasuredResults(); ok {
		out.Primary = &primary
	}
	return out, true
}

func rawOTDOA(v *location.ProvideLocationInformationR9IEs) (*RawOTDOAMeasurements, bool) {
	if v == nil || v.OTDOA == nil {
		return nil, false
	}
	signal, ok := v.OTDOA.SignalMeasurementInformation()
	if !ok {
		return nil, false
	}
	return &RawOTDOAMeasurements{Signal: signal}, true
}

// rawAGNSS extracts the UE-reported position from the common IE, not from
// v.AGNSS: TS 37.355 carries the actual coordinates in
// CommonProvideLocationInformation.LocationEstimate regardless of which
// method produced them; the A-GNSS-specific branch is supplementary
// reference-time/constellation metadata only.
func rawAGNSS(v *location.ProvideLocationInformationR9IEs) (*RawAGNSSMeasurements, bool) {
	if v == nil || v.Common == nil || v.Common.LocationEstimate == nil {
		return nil, false
	}
	return &RawAGNSSMeasurements{Estimate: *v.Common.LocationEstimate}, true
}

// UncertaintyCodeFromMeters retains the existing bounded simulation mapping
// while keeping simulated metre input out of the LCS wire model.
func UncertaintyCodeFromMeters(m float64) uint8 {
	if m <= 0 {
		return 0
	}
	for i := uint8(0); i < 127; i++ {
		if 10*(1.1*float64(i)+1) >= m {
			return i
		}
	}
	return 127
}
