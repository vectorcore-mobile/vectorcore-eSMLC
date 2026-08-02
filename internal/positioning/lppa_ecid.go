package positioning

import (
	"time"

	"github.com/vectorcore/esmlc/internal/lppa"
)

// LPPaECIDPolicy is local operator policy, not an assertion of eNB
// capability. Unlike ECID/OTDOA/A-GNSS this method needs no UE round trip at
// all (no LPP capability/location-information exchange with the target
// device): the E-SMLC asks the eNB directly, over LPPa, for the serving
// cell's identity and (if the eNB has it) its known antenna position. When
// enabled, this method takes priority over every UE-based method, since it
// is faster and does not depend on UE LPP support; when disabled (the
// default), job behaviour is unchanged from before this method existed.
type LPPaECIDPolicy struct{ Enabled bool }

// RawLPPaECIDMeasurement preserves the eNB-reported E-CID-MeasurementResult
// as method output, matching the raw-measurement-only convention used by
// RawECIDMeasurements/RawOTDOAMeasurements: no position is inferred here.
type RawLPPaECIDMeasurement struct {
	Result lppa.ECIDMeasurementResult
}

// LPPaActionKind identifies which LPPa message the SLs boundary must send
// next. Unlike LPP actions (procedure.Action, which carries a fully-typed
// lpp.Message), this package does not depend on internal/lppa's message
// types directly: it only decides *that* a message of a given kind must be
// sent, using the fixed request shape internal/sls builds (on-demand
// reporting, cell-ID quantity only — the bounded scope chosen for this
// method), keeping internal/positioning free of LPPa codec details.
type LPPaActionKind uint8

const (
	LPPaSendInitiationRequest LPPaActionKind = iota
	LPPaSendTerminationCommand
)

// LPPaAction is attached to an Outcome when the SLs boundary must send an
// LPPa message. ENBMeasurementID is only meaningful for
// LPPaSendTerminationCommand: it is learned from the eNB's Initiation
// Response and is absent (zero) otherwise.
type LPPaAction struct {
	Kind               LPPaActionKind
	TransactionID      uint16
	ESMLCMeasurementID uint8
	ENBMeasurementID   uint8
}

// lppaMeasurementID is always 1: Measurement-ID only needs to be unique
// among the concurrently active LPPa measurements the MME's Routing-ID
// scopes to a single UE context, and this package never runs more than one
// LPPaECID job for a given Scope at a time (method selection is mutually
// exclusive per job, exactly like ECID/OTDOA/A-GNSS already are).
const lppaMeasurementID uint8 = 1

// LPPaECIDEstimator prefers the eNB's own self-reported antenna position
// (E-UTRANAccessPointPosition) when present — the most authoritative source,
// requiring no operator-maintained catalog — and otherwise falls back to the
// same operator cell catalog the LPP-based ECID method already uses, keyed
// by the serving ECGI the original LCS-AP Location Request already carried
// (request.ServingECGI), exactly like ServingCellCatalogEstimator.
type LPPaECIDEstimator struct{ Store *CatalogStore }

func (e LPPaECIDEstimator) Estimate(request Request, method MethodResult, now time.Time) EstimationResult {
	if method.Method != MethodLPPaECID || method.LPPaECID == nil {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	if pos := method.LPPaECID.Result.AccessPointPosition; pos != nil {
		estimate := GeographicEstimate{
			Latitude: pos.Latitude, Longitude: pos.Longitude,
			HorizontalUncertainty: pos.UncertaintySemiMajor, // already a TS 23.032 Uncertainty-Code (0..127), not metres
			Source:                EstimateSourceLPPaAccessPointPosition,
			Timestamp:             now,
		}
		if err := estimate.Validate(); err != nil {
			return EstimationResult{Failure: InsufficientNetworkData}
		}
		return EstimationResult{Estimate: &estimate}
	}
	if e.Store == nil {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	catalog := e.Store.snapshot()
	if catalog == nil {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	cell, ok, stale := catalog.lookup(request.ServingECGI, now)
	if !ok {
		if stale {
			e.Store.recordStaleData()
		} else {
			e.Store.recordMissingCell()
		}
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	e.Store.recordAuthoritativeEstimate()
	estimate := GeographicEstimate{
		Latitude: cell.Latitude, Longitude: cell.Longitude, HorizontalUncertainty: cell.CoverageUncertainty,
		Source: EstimateSourceAuthoritativeServingCell, Timestamp: now,
		CatalogVersion: catalog.Version(), DataSource: cell.Source, RecordUpdatedAt: cell.UpdatedAt,
	}
	return EstimationResult{Estimate: &estimate}
}
