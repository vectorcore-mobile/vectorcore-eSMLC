// Package positioning owns method-neutral positioning-job lifecycle above LPP
// codecs and procedures. It does not encode ASN.1, calculate locations, or
// assume that an implemented method is permitted for every request.
package positioning

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp/capability"
	"github.com/vectorcore/esmlc/internal/lpp/location"
	"github.com/vectorcore/esmlc/internal/lpp/procedure"
	"github.com/vectorcore/esmlc/internal/lpp/transaction"
	"github.com/vectorcore/esmlc/internal/lppa"
	"github.com/vectorcore/esmlc/internal/uper"
)

var (
	ErrDuplicateJob = errors.New("positioning: active job already exists")
	ErrNotActive    = errors.New("positioning: job is not active")
)

type Scope struct {
	Association string
	Correlation [4]byte
}
type State uint8

const (
	AwaitingCapabilities State = iota
	AwaitingLocationInformation
	MeasurementsAvailable
	EstimateAvailable
	EstimationUnavailable
	QualityNotMet
	NoEligibleMethod
	Cancelled
	Expired
	ProcedureFailed
)

// ECIDPolicy is local operator policy, not an assertion of UE capability.
// RequestedMeasurements is retained exactly and is checked against an actual
// LPP ProvideCapabilities result before ECID is initiated.
type ECIDPolicy struct {
	Enabled               bool
	RequestedMeasurements location.ECIDRequestLocationInformation
}

// OTDOAPolicy is local operator policy, not an assertion of UE capability.
// Unlike ECIDPolicy there is no requested-measurements bitmap to check
// against capability: OTDOA-RequestLocationInformation's bounded root form
// carries only assistanceAvailability, which this package always sends as
// false (see Manager.locationOptionsForCapabilitiesLocked) because no
// assistance-data source is implemented.
type OTDOAPolicy struct {
	Enabled bool
}

// AGNSSPolicy is local operator policy, not an assertion of UE capability.
// This implementation only ever requests UE-based (MS-based) GPS: the UE
// reports its own already-computed position via the common location
// estimate, so there is no requested-measurements field to configure, and
// no assistance data is ever offered (AssistanceAvailability stays false;
// see Manager.locationOptionsForCapabilitiesLocked).
type AGNSSPolicy struct {
	Enabled bool
}
type Policy struct {
	ECID     ECIDPolicy
	OTDOA    OTDOAPolicy
	AGNSS    AGNSSPolicy
	LPPaECID LPPaECIDPolicy
}
type Request struct {
	Scope        Scope
	LocationType uint8
	ServingECGI  [7]byte
	Priority     *uint8
	QoS          *QoS
	LPPSupported *bool
	Deadline     time.Time
}

// QoS contains only constraints whose standards encoding and current result
// metadata are available to this workflow. Nil remains distinct from false or
// zero-valued constraints.
type QoS struct {
	HorizontalAccuracy *uint8
	VerticalRequested  *bool
	VerticalAccuracy   *uint8
	ResponseTime       *uint8
}
type Snapshot struct {
	ID                    uint64
	Request               Request
	State                 State
	Method                Method
	CapabilityTransaction *transaction.Key
	LocationTransaction   *transaction.Key
	Final                 *FinalOutcome
}
type Outcome struct {
	Snapshot Snapshot
	Actions  []procedure.Action
	// LPPa is non-nil when the SLs boundary must send an LPPa message
	// (Initiation Request or Termination Command); see LPPaAction.
	LPPa *LPPaAction
}

type job struct {
	Snapshot
	procedure *procedure.Orchestrator
	// LPPa-specific bookkeeping; unused for LPP-based methods.
	lppaTransactionID    uint16
	lppaENBMeasurementID *uint8 // learned from the eNB's Initiation Response
}
type Manager struct {
	mu          sync.Mutex
	next        uint64
	nextLPPaTxn uint16
	policy      Policy
	estimator   Estimator
	jobs        map[Scope]*job
}

func New(policy Policy) *Manager { return &Manager{policy: policy, jobs: map[Scope]*job{}} }

func NewWithEstimator(policy Policy, estimator Estimator) *Manager {
	return &Manager{policy: policy, estimator: estimator, jobs: map[Scope]*job{}}
}

func (m *Manager) Start(request Request, p *procedure.Orchestrator, now time.Time) (Outcome, error) {
	if p == nil || request.Scope.Association == "" || request.Deadline.IsZero() || !request.Deadline.After(now) {
		return Outcome{}, fmt.Errorf("positioning: invalid request")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.jobs[request.Scope]; ok {
		return Outcome{}, ErrDuplicateJob
	}
	m.next++
	j := &job{Snapshot: Snapshot{ID: m.next, Request: request, State: AwaitingCapabilities}, procedure: p}
	m.jobs[request.Scope] = j
	// LPPaECID needs no UE round trip at all, so it bypasses the LPPSupported
	// gate below (which only concerns UE LPP capability) and is tried before
	// every UE-based method.
	if m.policy.LPPaECID.Enabled {
		j.Method = MethodLPPaECID
		j.State = AwaitingLocationInformation
		txn := m.nextLPPaTxn
		m.nextLPPaTxn = (m.nextLPPaTxn + 1) % 32768
		j.lppaTransactionID = txn
		return Outcome{Snapshot: j.Snapshot, LPPa: &LPPaAction{Kind: LPPaSendInitiationRequest, TransactionID: txn, ESMLCMeasurementID: lppaMeasurementID}}, nil
	}
	if request.LPPSupported != nil && !*request.LPPSupported {
		j.State = NoEligibleMethod
		j.Final = &FinalOutcome{Kind: FinalLPPUnsupported}
		return m.finishLocked(j), nil
	}
	method, capOptions, ok := m.selectMethodLocked()
	if !ok {
		j.State = NoEligibleMethod
		j.Final = &FinalOutcome{Kind: FinalNoEligibleMethod}
		return m.finishLocked(j), nil
	}
	j.Method = method
	r, err := p.StartCapabilities(capOptions, now)
	if err != nil {
		j.State = ProcedureFailed
		j.Final = &FinalOutcome{Kind: FinalProcedureFailure}
		return m.finishLocked(j), err
	}
	if len(r.Actions) != 1 || r.Actions[0].Key == nil {
		j.State = ProcedureFailed
		j.Final = &FinalOutcome{Kind: FinalProcedureFailure}
		return m.finishLocked(j), fmt.Errorf("positioning: missing capability transaction")
	}
	k := *r.Actions[0].Key
	j.CapabilityTransaction = &k
	return Outcome{Snapshot: j.Snapshot, Actions: r.Actions}, nil
}

// selectMethodLocked applies local operator policy only, not QoS,
// capability, or prior-outcome awareness (see docs/architecture.md). ECID
// takes priority when both are enabled, so existing ECID-only deployments
// are unaffected by OTDOA becoming available.
func (m *Manager) selectMethodLocked() (Method, procedure.StartOptions, bool) {
	if m.policy.ECID.Enabled {
		if err := (location.RequestLocationInformationR9IEs{ECID: &m.policy.ECID.RequestedMeasurements}).Validate(); err == nil {
			return MethodECID, procedure.StartOptions{RequestECID: true}, true
		}
	}
	if m.policy.OTDOA.Enabled {
		return MethodOTDOA, procedure.StartOptions{RequestOTDOA: true}, true
	}
	if m.policy.AGNSS.Enabled {
		return MethodAGNSS, procedure.StartOptions{RequestAGNSS: true}, true
	}
	return 0, procedure.StartOptions{}, false
}

// locationOptionsForCapabilitiesLocked checks the job's already-selected
// method against the actual reported UE capability before building the
// RequestLocationInformation options; it never switches method mid-job.
func (m *Manager) locationOptionsForCapabilitiesLocked(j *job, cap *capability.ProvideCapabilitiesR9IEs) (procedure.StartLocationInformationOptions, bool) {
	switch j.Method {
	case MethodECID:
		if !supportsECID(cap, m.policy.ECID.RequestedMeasurements) {
			return procedure.StartLocationInformationOptions{}, false
		}
		return procedure.StartLocationInformationOptions{ECID: &m.policy.ECID.RequestedMeasurements}, true
	case MethodOTDOA:
		if !supportsOTDOA(cap) {
			return procedure.StartLocationInformationOptions{}, false
		}
		return procedure.StartLocationInformationOptions{OTDOA: &location.OTDOARequestLocationInformation{AssistanceAvailability: false}}, true
	case MethodAGNSS:
		if !supportsAGNSS(cap) {
			return procedure.StartLocationInformationOptions{}, false
		}
		return procedure.StartLocationInformationOptions{
			Common: &location.CommonRequestLocationInformation{LocationInformationType: location.LocationEstimateRequired},
			AGNSS:  &location.AGNSSRequestLocationInformation{GNSSMethods: gpsOnlyBitmap()},
		}, true
	default:
		return procedure.StartLocationInformationOptions{}, false
	}
}

// Apply accepts procedure events scoped by the SLs correlation. It only
// advances the job when the event key matches the job's active LPP transaction.
func (m *Manager) Apply(scope Scope, events []procedure.Event, now time.Time) (Outcome, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j := m.jobs[scope]
	if j == nil || !active(j.State) {
		return Outcome{}, ErrNotActive
	}
	if !now.Before(j.Request.Deadline) {
		return m.terminalLocked(j, Expired, now), nil
	}
	for _, event := range events {
		if event.Key == nil {
			continue
		}
		if j.State == AwaitingCapabilities && j.CapabilityTransaction != nil && *event.Key == *j.CapabilityTransaction && event.Kind == procedure.CapabilitiesEnvelopeProvided {
			locOptions, ok := m.locationOptionsForCapabilitiesLocked(j, event.Capabilities)
			if !ok {
				j.State = NoEligibleMethod
				j.Final = &FinalOutcome{Kind: FinalNoEligibleMethod}
				return m.finishLocked(j), nil
			}
			r, err := j.procedure.StartLocationInformation(locOptions, now)
			if err != nil {
				j.State = ProcedureFailed
				j.Final = &FinalOutcome{Kind: FinalProcedureFailure}
				return m.finishLocked(j), err
			}
			if len(r.Actions) != 1 || r.Actions[0].Key == nil {
				j.State = ProcedureFailed
				j.Final = &FinalOutcome{Kind: FinalProcedureFailure}
				return m.finishLocked(j), fmt.Errorf("positioning: missing location transaction")
			}
			k := *r.Actions[0].Key
			j.LocationTransaction = &k
			j.State = AwaitingLocationInformation
			return Outcome{Snapshot: j.Snapshot, Actions: r.Actions}, nil
		}
		if j.State == AwaitingLocationInformation && j.LocationTransaction != nil && *event.Key == *j.LocationTransaction && event.Kind == procedure.LocationInformationEnvelopeProvided {
			return m.handleMethodResultLocked(j, event.LocationInformation, now), nil
		}
		if (event.Kind == procedure.ProcedureAborted || event.Kind == procedure.ProcedureFailed) && matches(j, event.Key) {
			j.State = ProcedureFailed
			j.Final = &FinalOutcome{Kind: FinalProcedureFailure}
			return m.finishLocked(j), nil
		}
	}
	return Outcome{Snapshot: j.Snapshot}, nil
}
func (m *Manager) Cancel(scope Scope, now time.Time) (Outcome, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j := m.jobs[scope]
	if j == nil || !active(j.State) {
		return Outcome{}, false
	}
	return m.terminalLocked(j, Cancelled, now), true
}
func (m *Manager) DropAssociation(association string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for scope, j := range m.jobs {
		if scope.Association == association {
			j.State = Cancelled
			delete(m.jobs, scope)
		}
	}
}
func (m *Manager) Snapshot(scope Scope) (Snapshot, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j := m.jobs[scope]
	if j == nil {
		return Snapshot{}, false
	}
	return j.Snapshot, true
}

// ActiveJobs reports the number of jobs currently tracked (not yet
// terminal). Intended for observability (a live gauge), not control flow.
func (m *Manager) ActiveJobs() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.jobs)
}

// PruneResult pairs an expired job's Scope with the Outcome its expiry
// produced, so the caller (which owns the SLs association and the LPP
// procedure/transaction-store keyed by the same Scope) can send any
// resulting wire actions and drop its own per-Scope state.
type PruneResult struct {
	Scope   Scope
	Outcome Outcome
}

// Prune proactively expires every active job whose Deadline has passed.
// Apply already expires a job past its deadline, but only when a new event
// for that exact Scope arrives (see Apply's own deadline check) — a UE/eNB
// that goes silent after the job's last action never sends one, so without
// Prune such a job (and, via job.procedure, its associated LPP transaction
// store and the caller's per-Scope session state) would remain live
// forever. Call this periodically (e.g. from a ticker); it is not part of
// the per-request path.
func (m *Manager) Prune(now time.Time) []PruneResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	var scopes []Scope
	for scope, j := range m.jobs {
		if active(j.State) && !now.Before(j.Request.Deadline) {
			scopes = append(scopes, scope)
		}
	}
	if len(scopes) == 0 {
		return nil
	}
	out := make([]PruneResult, 0, len(scopes))
	for _, scope := range scopes {
		j := m.jobs[scope]
		out = append(out, PruneResult{Scope: scope, Outcome: m.terminalLocked(j, Expired, now)})
	}
	return out
}
func active(s State) bool { return s == AwaitingCapabilities || s == AwaitingLocationInformation }
func matches(j *job, key *transaction.Key) bool {
	return key != nil && ((j.CapabilityTransaction != nil && *key == *j.CapabilityTransaction) || (j.LocationTransaction != nil && *key == *j.LocationTransaction))
}
func (m *Manager) terminalLocked(j *job, state State, now time.Time) Outcome {
	var actions []procedure.Action
	var key *transaction.Key
	if j.LocationTransaction != nil {
		key = j.LocationTransaction
	} else if j.CapabilityTransaction != nil {
		key = j.CapabilityTransaction
	}
	if key != nil {
		if r, err := j.procedure.Abort(*key, now); err == nil {
			actions = append(actions, r.Actions...)
		}
	}
	j.State = state
	if j.Final == nil {
		switch state {
		case Expired:
			j.Final = &FinalOutcome{Kind: FinalDeadlineExpired}
		case Cancelled:
			j.Final = &FinalOutcome{Kind: FinalCancelled}
		case ProcedureFailed:
			j.Final = &FinalOutcome{Kind: FinalProcedureFailure}
		}
	}
	outcome := m.finishLocked(j)
	outcome.Actions = actions
	return outcome
}

// finishLocked ends a job. LPPaECID jobs that reached a live eNB
// measurement (learned the eNB's Measurement-ID from an Initiation Response)
// must have that measurement explicitly ended, regardless of how the job
// itself concluded, so the outcome carries a Termination Command action
// whenever that applies.
func (m *Manager) finishLocked(j *job) Outcome {
	delete(m.jobs, j.Request.Scope)
	outcome := Outcome{Snapshot: j.Snapshot}
	if j.Method == MethodLPPaECID && j.lppaENBMeasurementID != nil {
		outcome.LPPa = &LPPaAction{Kind: LPPaSendTerminationCommand, TransactionID: j.lppaTransactionID, ESMLCMeasurementID: lppaMeasurementID, ENBMeasurementID: *j.lppaENBMeasurementID}
	}
	return outcome
}
func (m *Manager) handleMethodResultLocked(j *job, provided *location.ProvideLocationInformationR9IEs, now time.Time) Outcome {
	method, ok := rawMethodResult(j.Method, provided, now)
	if !ok {
		j.State = ProcedureFailed
		j.Final = &FinalOutcome{Kind: FinalProcedureFailure}
		return m.finishLocked(j)
	}
	return m.completeWithMethodResultLocked(j, method, now)
}

// completeWithMethodResultLocked runs the estimator and QoS evaluation
// shared by every method (LPP-based or LPPaECID) once its raw measurement
// result is available, then ends the job.
func (m *Manager) completeWithMethodResultLocked(j *job, method *MethodResult, now time.Time) Outcome {
	if m.estimator == nil {
		j.State = EstimationUnavailable
		j.Final = &FinalOutcome{Kind: FinalMeasurementsWithoutEstimator, MethodResult: method, Failure: EstimatorUnavailable}
		return m.finishLocked(j)
	}
	estimated := m.estimator.Estimate(j.Request, *method, now)
	if estimated.Estimate == nil {
		j.State = EstimationUnavailable
		j.Final = &FinalOutcome{Kind: FinalEstimationFailed, MethodResult: method, Failure: estimated.Failure}
		return m.finishLocked(j)
	}
	if err := estimated.Estimate.Validate(); err != nil {
		j.State = EstimationUnavailable
		j.Final = &FinalOutcome{Kind: FinalEstimationFailed, MethodResult: method, Failure: InsufficientNetworkData}
		return m.finishLocked(j)
	}
	fulfilment := evaluateQoS(j.Request.QoS, *estimated.Estimate)
	if fulfilment == AccuracyNotFulfilled {
		j.State = QualityNotMet
		j.Final = &FinalOutcome{Kind: FinalQualityNotMet, MethodResult: method, Estimate: estimated.Estimate, Failure: RequestedQualityNotMet, Accuracy: fulfilment}
		return m.finishLocked(j)
	}
	j.State = EstimateAvailable
	j.Final = &FinalOutcome{Kind: FinalEstimateAvailable, MethodResult: method, Estimate: estimated.Estimate, Accuracy: fulfilment}
	return m.finishLocked(j)
}

// lppaJobLocked returns the active LPPaECID job for scope, or nil if there is
// none (already finished, wrong method, or never started).
func (m *Manager) lppaJobLocked(scope Scope) *job {
	j := m.jobs[scope]
	if j == nil || j.Method != MethodLPPaECID || !active(j.State) {
		return nil
	}
	return j
}

// ApplyLPPaInitiationResponse applies the eNB's E-CIDMeasurementInitiationResponse.
// transactionID must match the Initiation Request this job sent. If the
// response already carries a measurement result (on-demand reporting may
// return it immediately), the job completes now; otherwise it remains active,
// awaiting a subsequent ApplyLPPaReport.
func (m *Manager) ApplyLPPaInitiationResponse(scope Scope, transactionID uint16, resp lppa.InitiationResponse, now time.Time) (Outcome, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j := m.lppaJobLocked(scope)
	if j == nil || j.lppaTransactionID != transactionID {
		return Outcome{}, ErrNotActive
	}
	if !now.Before(j.Request.Deadline) {
		return m.terminalLocked(j, Expired, now), nil
	}
	enbID := resp.ENBMeasurementID
	j.lppaENBMeasurementID = &enbID
	if resp.Result == nil {
		return Outcome{Snapshot: j.Snapshot}, nil
	}
	method := &MethodResult{Method: MethodLPPaECID, LPPaECID: &RawLPPaECIDMeasurement{Result: *resp.Result}, CompletedAt: now}
	return m.completeWithMethodResultLocked(j, method, now), nil
}

// ApplyLPPaInitiationFailure applies the eNB's E-CIDMeasurementInitiationFailure:
// the eNB never started a measurement, so there is nothing to terminate.
func (m *Manager) ApplyLPPaInitiationFailure(scope Scope, transactionID uint16, _ lppa.InitiationFailure, now time.Time) (Outcome, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j := m.lppaJobLocked(scope)
	if j == nil || j.lppaTransactionID != transactionID {
		return Outcome{}, ErrNotActive
	}
	if !now.Before(j.Request.Deadline) {
		return m.terminalLocked(j, Expired, now), nil
	}
	j.State = ProcedureFailed
	j.Final = &FinalOutcome{Kind: FinalProcedureFailure}
	return m.finishLocked(j), nil
}

// ApplyLPPaReport applies an unsolicited E-CIDMeasurementReport, completing a
// job whose Initiation Response did not already carry a result.
func (m *Manager) ApplyLPPaReport(scope Scope, report lppa.Report, now time.Time) (Outcome, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j := m.lppaJobLocked(scope)
	if j == nil || j.lppaENBMeasurementID == nil || *j.lppaENBMeasurementID != report.ENBMeasurementID {
		return Outcome{}, ErrNotActive
	}
	if !now.Before(j.Request.Deadline) {
		return m.terminalLocked(j, Expired, now), nil
	}
	method := &MethodResult{Method: MethodLPPaECID, LPPaECID: &RawLPPaECIDMeasurement{Result: report.Result}, CompletedAt: now}
	return m.completeWithMethodResultLocked(j, method, now), nil
}

// ApplyLPPaFailureIndication applies an unsolicited E-CIDMeasurementFailureIndication:
// the eNB was already measuring but can no longer continue.
func (m *Manager) ApplyLPPaFailureIndication(scope Scope, indication lppa.FailureIndication, now time.Time) (Outcome, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j := m.lppaJobLocked(scope)
	if j == nil || j.lppaENBMeasurementID == nil || *j.lppaENBMeasurementID != indication.ENBMeasurementID {
		return Outcome{}, ErrNotActive
	}
	if !now.Before(j.Request.Deadline) {
		return m.terminalLocked(j, Expired, now), nil
	}
	// The eNB reported it stopped measuring; there is nothing left to
	// terminate, unlike other terminal paths (see finishLocked).
	j.lppaENBMeasurementID = nil
	j.State = ProcedureFailed
	j.Final = &FinalOutcome{Kind: FinalProcedureFailure}
	return m.finishLocked(j), nil
}
func acceptsQoS(q *QoS, estimate GeographicEstimate) bool {
	return evaluateQoS(q, estimate) != AccuracyNotFulfilled
}

func evaluateQoS(q *QoS, estimate GeographicEstimate) AccuracyFulfilment {
	if q == nil {
		return AccuracyUnevaluated
	}
	if q.HorizontalAccuracy != nil && estimate.HorizontalUncertainty > *q.HorizontalAccuracy {
		return AccuracyNotFulfilled
	}
	if q.VerticalRequested != nil && *q.VerticalRequested {
		return AccuracyNotFulfilled
	}
	if q.VerticalAccuracy != nil {
		return AccuracyNotFulfilled
	}
	if q.HorizontalAccuracy == nil {
		return AccuracyUnevaluated
	}
	return AccuracyFulfilled
}
func supportsECID(cap *capability.ProvideCapabilitiesR9IEs, requested location.ECIDRequestLocationInformation) bool {
	if cap == nil || cap.ECID == nil {
		return false
	}
	for i := 0; i < requested.RequestedMeasurements.BitLen(); i++ {
		b := requested.RequestedMeasurements.Bytes()[i/8]&(1<<uint(7-i%8)) != 0
		if !b {
			continue
		}
		if i >= cap.ECID.MeasurementSupport.BitLen() || cap.ECID.MeasurementSupport.Bytes()[i/8]&(1<<uint(7-i%8)) == 0 {
			return false
		}
	}
	return true
}

// supportsOTDOA requires the UE-assisted OTDOA mode bit specifically: this
// package requests only that mode (see locationOptionsForCapabilitiesLocked)
// and has no assistance-data source for NB-IoT OTDOA variants.
func supportsOTDOA(cap *capability.ProvideCapabilitiesR9IEs) bool {
	return cap != nil && cap.OTDOA != nil && cap.OTDOA.SupportsUEAssisted()
}

// supportsAGNSS requires a GPS entry advertising ue-based (MS-based) mode
// specifically: this package requests only GPS in that mode (see
// locationOptionsForCapabilitiesLocked) and has no assistance-data source
// that would make ue-assisted mode useful.
func supportsAGNSS(cap *capability.ProvideCapabilitiesR9IEs) bool {
	return cap != nil && cap.AGNSS != nil && cap.AGNSS.SupportsGPSUEBased()
}

// gpsOnlyBitmap is the canonical (minimal-trailing-zero) UPER encoding of a
// GNSS-ID-Bitmap with only bit 0 (GPS) set.
func gpsOnlyBitmap() uper.BitString {
	v, err := uper.NewBitString([]byte{0x80}, 1)
	if err != nil {
		panic(err) // unreachable: a single valid bit is always constructible
	}
	return v
}

// rawMethodResult decodes the provide-location payload branch matching the
// job's already-selected method. A mismatched or absent branch is a
// procedure-level failure, not a silent fallback to another method.
func rawMethodResult(method Method, provided *location.ProvideLocationInformationR9IEs, now time.Time) (*MethodResult, bool) {
	switch method {
	case MethodECID:
		raw, ok := rawECID(provided)
		if !ok {
			return nil, false
		}
		return &MethodResult{Method: MethodECID, ECID: raw, CompletedAt: now}, true
	case MethodOTDOA:
		raw, ok := rawOTDOA(provided)
		if !ok {
			return nil, false
		}
		return &MethodResult{Method: MethodOTDOA, OTDOA: raw, CompletedAt: now}, true
	case MethodAGNSS:
		raw, ok := rawAGNSS(provided)
		if !ok {
			return nil, false
		}
		return &MethodResult{Method: MethodAGNSS, AGNSS: raw, CompletedAt: now}, true
	default:
		return nil, false
	}
}
