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
)

// RawECIDMeasurements preserves ECID radio measurements as method output. It
// deliberately contains no inferred position.
type RawECIDMeasurements struct {
	Primary *result.MeasuredResultsElement
	Results []result.MeasuredResultsElement
}

// MethodResult records a completed protocol-method attempt. Measurements may
// be available even when no geographic estimate can be calculated.
type MethodResult struct {
	Method      Method
	ECID        *RawECIDMeasurements
	CompletedAt time.Time
}

// EstimateSource identifies how an estimate was derived. Simulation is never
// represented as a real A-GNSS or ECID estimate.
type EstimateSource uint8

const (
	EstimateSourceSimulation EstimateSource = iota + 1
	EstimateSourceAuthoritativeServingCell
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

// Estimator is a synchronous, job-scoped boundary. Implementations receive
// only the verified request and method output belonging to that job.
type Estimator interface {
	Estimate(Request, MethodResult, time.Time) EstimationResult
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
