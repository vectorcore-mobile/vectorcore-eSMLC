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
type Policy struct{ ECID ECIDPolicy }
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
	CapabilityTransaction *transaction.Key
	LocationTransaction   *transaction.Key
	Final                 *FinalOutcome
}
type Outcome struct {
	Snapshot Snapshot
	Actions  []procedure.Action
}

type job struct {
	Snapshot
	procedure *procedure.Orchestrator
}
type Manager struct {
	mu        sync.Mutex
	next      uint64
	policy    Policy
	estimator Estimator
	jobs      map[Scope]*job
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
	if !m.policy.ECID.Enabled {
		j.State = NoEligibleMethod
		j.Final = &FinalOutcome{Kind: FinalNoEligibleMethod}
		return m.finishLocked(j), nil
	}
	if request.LPPSupported != nil && !*request.LPPSupported {
		j.State = NoEligibleMethod
		j.Final = &FinalOutcome{Kind: FinalNoEligibleMethod}
		return m.finishLocked(j), nil
	}
	if err := (location.RequestLocationInformationR9IEs{ECID: &m.policy.ECID.RequestedMeasurements}).Validate(); err != nil {
		j.State = NoEligibleMethod
		j.Final = &FinalOutcome{Kind: FinalNoEligibleMethod}
		return m.finishLocked(j), nil
	}
	r, err := p.StartCapabilities(procedure.StartOptions{RequestECID: true}, now)
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
			if !supportsECID(event.Capabilities, m.policy.ECID.RequestedMeasurements) {
				j.State = NoEligibleMethod
				j.Final = &FinalOutcome{Kind: FinalNoEligibleMethod}
				return m.finishLocked(j), nil
			}
			r, err := j.procedure.StartLocationInformation(procedure.StartLocationInformationOptions{ECID: &m.policy.ECID.RequestedMeasurements}, now)
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
func (m *Manager) finishLocked(j *job) Outcome {
	delete(m.jobs, j.Request.Scope)
	return Outcome{Snapshot: j.Snapshot}
}
func (m *Manager) handleMethodResultLocked(j *job, provided *location.ProvideLocationInformationR9IEs, now time.Time) Outcome {
	raw, ok := rawECID(provided)
	if !ok {
		j.State = ProcedureFailed
		j.Final = &FinalOutcome{Kind: FinalProcedureFailure}
		return m.finishLocked(j)
	}
	method := &MethodResult{Method: MethodECID, ECID: raw, CompletedAt: now}
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
	if !acceptsQoS(j.Request.QoS, *estimated.Estimate) {
		j.State = QualityNotMet
		j.Final = &FinalOutcome{Kind: FinalQualityNotMet, MethodResult: method, Estimate: estimated.Estimate, Failure: RequestedQualityNotMet}
		return m.finishLocked(j)
	}
	j.State = EstimateAvailable
	j.Final = &FinalOutcome{Kind: FinalEstimateAvailable, MethodResult: method, Estimate: estimated.Estimate}
	return m.finishLocked(j)
}
func acceptsQoS(q *QoS, estimate GeographicEstimate) bool {
	if q == nil {
		return true
	}
	if q.HorizontalAccuracy != nil && estimate.HorizontalUncertainty > *q.HorizontalAccuracy {
		return false
	}
	if q.VerticalRequested != nil && *q.VerticalRequested {
		return false
	}
	if q.VerticalAccuracy != nil {
		return false
	}
	return true
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
