package positioning

import (
	"errors"
	"testing"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/capability"
	"github.com/vectorcore/esmlc/internal/lpp/location"
	locationresult "github.com/vectorcore/esmlc/internal/lpp/location/result"
	"github.com/vectorcore/esmlc/internal/lpp/procedure"
	"github.com/vectorcore/esmlc/internal/lpp/transaction"
	"github.com/vectorcore/esmlc/internal/uper"
)

func policy(t *testing.T, enabled bool) Policy {
	t.Helper()
	bits, err := uper.NewBitString([]byte{0x80}, 1)
	if err != nil {
		t.Fatal(err)
	}
	return Policy{ECID: ECIDPolicy{Enabled: enabled, RequestedMeasurements: location.ECIDRequestLocationInformation{RequestedMeasurements: bits}}}
}
func proc(t *testing.T) *procedure.Orchestrator {
	t.Helper()
	store, err := transaction.NewStore(transaction.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	o, err := procedure.New(store, procedure.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	return o
}
func req() Request {
	return Request{Scope: Scope{Association: "mme", Correlation: [4]byte{1}}, LocationType: 0, Deadline: time.Unix(100, 0)}
}

func TestJobStartsCapabilityThenEligibleECID(t *testing.T) {
	now := time.Unix(0, 0)
	o := proc(t)
	m := New(policy(t, true))
	r, err := m.Start(req(), o, now)
	if err != nil || r.Snapshot.ID == 0 || r.Snapshot.State != AwaitingCapabilities || len(r.Actions) != 1 || r.Actions[0].Message.Body.Kind != lpp.BodyRequestCapabilities {
		t.Fatalf("start %#v %v", r, err)
	}
	key := *r.Actions[0].Key
	capBits, _ := uper.NewBitString([]byte{0x80}, 1)
	in := lpp.Message{TransactionID: &lpp.TransactionID{Initiator: key.Initiator, TransactionNumber: key.Number}, Body: &lpp.Body{Kind: lpp.BodyProvideCapabilities, ProvideCapabilities: &capability.ProvideCapabilitiesR9IEs{ECID: &capability.ECIDProvideCapabilities{MeasurementSupport: capBits}}}}
	pr, err := o.HandleInbound(in, now)
	if err != nil {
		t.Fatal(err)
	}
	r, err = m.Apply(req().Scope, pr.Events, now)
	if err != nil || r.Snapshot.State != AwaitingLocationInformation || len(r.Actions) != 1 || r.Actions[0].Message.Body.Kind != lpp.BodyRequestLocationInformation {
		t.Fatalf("apply %#v %v", r, err)
	}
	if r.Snapshot.ID == 0 || r.Snapshot.CapabilityTransaction == nil || r.Snapshot.LocationTransaction == nil || *r.Snapshot.CapabilityTransaction == *r.Snapshot.LocationTransaction {
		t.Fatal("identities not distinct")
	}
}
func TestJobNoMethodCancellationAndLateEvent(t *testing.T) {
	now := time.Unix(0, 0)
	o := proc(t)
	m := New(policy(t, false))
	r, err := m.Start(req(), o, now)
	if err != nil || r.Snapshot.State != NoEligibleMethod || len(r.Actions) != 0 {
		t.Fatalf("no method %#v %v", r, err)
	}
	if _, err = m.Apply(req().Scope, nil, now); !errors.Is(err, ErrNotActive) {
		t.Fatal(err)
	}
	m = New(policy(t, true))
	r, err = m.Start(req(), o, now)
	if err != nil {
		t.Fatal(err)
	}
	if outcome, ok := m.Cancel(req().Scope, now); !ok || outcome.Snapshot.State != Cancelled || len(outcome.Actions) != 1 {
		t.Fatal(outcome)
	}
	if _, err = m.Apply(req().Scope, nil, now); !errors.Is(err, ErrNotActive) {
		t.Fatal(err)
	}
	_ = r
}

func TestJobDeadlineExpiresAndAbortsActiveProcedure(t *testing.T) {
	now := time.Unix(0, 0)
	o := proc(t)
	m := New(policy(t, true))
	request := req()
	request.Deadline = now.Add(time.Second)
	if _, err := m.Start(request, o, now); err != nil {
		t.Fatal(err)
	}
	r, err := m.Apply(request.Scope, nil, request.Deadline)
	if err != nil || r.Snapshot.State != Expired || len(r.Actions) != 1 {
		t.Fatalf("deadline %#v %v", r, err)
	}
}

// TestManagerPruneExpiresJobsWithoutTriggeringEvent is the regression case
// for the leak this fixes: a job whose UE/eNB never sends another LPP/LPPa
// event for its correlation was previously only ever expired reactively by
// Apply, which nothing calls without a triggering event — so the job (and
// its per-Scope LPP session/transaction state, owned by the caller) stayed
// in m.jobs for the life of the process. Prune must expire it on its own,
// with no Apply call involved at all.
func TestManagerPruneExpiresJobsWithoutTriggeringEvent(t *testing.T) {
	now := time.Unix(0, 0)
	o := proc(t)
	m := New(policy(t, true))
	request := req()
	request.Deadline = now.Add(time.Second)
	if _, err := m.Start(request, o, now); err != nil {
		t.Fatal(err)
	}
	if got := m.ActiveJobs(); got != 1 {
		t.Fatalf("active jobs before deadline = %d, want 1", got)
	}
	// Before the deadline, Prune must leave the still-active job alone.
	if results := m.Prune(now); len(results) != 0 {
		t.Fatalf("prune before deadline = %#v, want none", results)
	}
	if got := m.ActiveJobs(); got != 1 {
		t.Fatalf("active jobs after early prune = %d, want 1 (untouched)", got)
	}
	results := m.Prune(request.Deadline)
	if len(results) != 1 {
		t.Fatalf("prune at deadline = %#v, want exactly one expired job", results)
	}
	r := results[0]
	if r.Scope != request.Scope {
		t.Fatalf("expired scope = %#v, want %#v", r.Scope, request.Scope)
	}
	if r.Outcome.Snapshot.State != Expired || len(r.Outcome.Actions) != 1 {
		t.Fatalf("expired outcome = %#v", r.Outcome)
	}
	if got := m.ActiveJobs(); got != 0 {
		t.Fatalf("active jobs after prune = %d, want 0 (job must be removed from m.jobs)", got)
	}
	// A second sweep must find nothing left to expire — Prune deletes as it
	// goes, matching finishLocked's own delete(m.jobs, ...) on every other
	// terminal path.
	if results := m.Prune(request.Deadline); len(results) != 0 {
		t.Fatalf("second prune = %#v, want none (already removed)", results)
	}
}

func TestECIDMeasurementsBecomeTypedEstimatorUnavailableOutcome(t *testing.T) {
	now := time.Unix(0, 0)
	o := proc(t)
	m := New(policy(t, true))
	start, err := m.Start(req(), o, now)
	if err != nil {
		t.Fatal(err)
	}
	capBits, _ := uper.NewBitString([]byte{0x80}, 1)
	cap := lpp.Message{TransactionID: &lpp.TransactionID{Initiator: start.Actions[0].Key.Initiator, TransactionNumber: start.Actions[0].Key.Number}, Body: &lpp.Body{Kind: lpp.BodyProvideCapabilities, ProvideCapabilities: &capability.ProvideCapabilitiesR9IEs{ECID: &capability.ECIDProvideCapabilities{MeasurementSupport: capBits}}}}
	pr, err := o.HandleInbound(cap, now)
	if err != nil {
		t.Fatal(err)
	}
	locationStart, err := m.Apply(req().Scope, pr.Events, now)
	if err != nil {
		t.Fatal(err)
	}
	pci, _ := locationresult.NewPhysicalCellID(1)
	arfcn, _ := locationresult.NewEUTRAARFCN(100)
	measured, _ := locationresult.NewMeasuredResultsElement(pci, arfcn, locationresult.MeasuredResultsElementOptions{})
	signal, _ := location.NewECIDSignalMeasurementInformation(nil, []locationresult.MeasuredResultsElement{measured})
	ecid, _ := location.NewECIDProvideLocationInformation(signal)
	provide := lpp.Message{TransactionID: &lpp.TransactionID{Initiator: locationStart.Actions[0].Key.Initiator, TransactionNumber: locationStart.Actions[0].Key.Number}, Body: &lpp.Body{Kind: lpp.BodyProvideLocationInformation, ProvideLocationInformation: &location.ProvideLocationInformationR9IEs{ECID: &ecid}}}
	pr, err = o.HandleInbound(provide, now)
	if err != nil {
		t.Fatal(err)
	}
	final, err := m.Apply(req().Scope, pr.Events, now)
	if err != nil || final.Snapshot.State != EstimationUnavailable || final.Snapshot.Final == nil || final.Snapshot.Final.Kind != FinalMeasurementsWithoutEstimator || final.Snapshot.Final.MethodResult == nil || final.Snapshot.Final.MethodResult.ECID == nil {
		t.Fatalf("final %#v %v", final, err)
	}
}

func TestSimulationEstimatorIsExplicitlyLabelled(t *testing.T) {
	r := (SimulationEstimator{Latitude: 38, Longitude: -90, Uncertainty: 4}).Estimate(req(), MethodResult{Method: MethodECID}, time.Unix(1, 0))
	if r.Estimate == nil || r.Estimate.Source != EstimateSourceSimulation || r.Failure != 0 {
		t.Fatalf("simulation %#v", r)
	}
}

func TestRequestConstraintsExcludeLPPAndRejectUnsuitableEstimate(t *testing.T) {
	now := time.Unix(0, 0)
	supported := false
	request := req()
	request.LPPSupported = &supported
	r, err := New(policy(t, true)).Start(request, proc(t), now)
	if err != nil || r.Snapshot.State != NoEligibleMethod || r.Snapshot.Final == nil || r.Snapshot.Final.Kind != FinalLPPUnsupported || len(r.Actions) != 0 {
		t.Fatalf("LPP exclusion %#v %v", r, err)
	}
	accuracy := uint8(3)
	if acceptsQoS(&QoS{HorizontalAccuracy: &accuracy}, GeographicEstimate{HorizontalUncertainty: 4}) {
		t.Fatal("inaccurate estimate accepted")
	}
	vertical := true
	if acceptsQoS(&QoS{VerticalRequested: &vertical}, GeographicEstimate{}) {
		t.Fatal("vertical request accepted without vertical estimate")
	}
	if !acceptsQoS(nil, GeographicEstimate{}) {
		t.Fatal("absent QoS was not treated as unevaluated")
	}
	if got := evaluateQoS(nil, GeographicEstimate{}); got != AccuracyUnevaluated {
		t.Fatalf("absent QoS fulfilment %d", got)
	}
	accuracy = 4
	if got := evaluateQoS(&QoS{HorizontalAccuracy: &accuracy}, GeographicEstimate{HorizontalUncertainty: 4}); got != AccuracyFulfilled {
		t.Fatalf("fulfilled QoS %d", got)
	}
	if got := evaluateQoS(&QoS{HorizontalAccuracy: &accuracy}, GeographicEstimate{HorizontalUncertainty: 5}); got != AccuracyNotFulfilled {
		t.Fatalf("unfulfilled QoS %d", got)
	}
}

func TestSimulationEstimateCanBeRejectedByQoS(t *testing.T) {
	now := time.Unix(0, 0)
	accuracy := uint8(3)
	request := req()
	request.QoS = &QoS{HorizontalAccuracy: &accuracy}
	o := proc(t)
	m := NewWithEstimator(policy(t, true), SimulationEstimator{Latitude: 38, Longitude: -90, Uncertainty: 4})
	start, err := m.Start(request, o, now)
	if err != nil {
		t.Fatal(err)
	}
	capBits, _ := uper.NewBitString([]byte{0x80}, 1)
	cap := lpp.Message{TransactionID: &lpp.TransactionID{Initiator: start.Actions[0].Key.Initiator, TransactionNumber: start.Actions[0].Key.Number}, Body: &lpp.Body{Kind: lpp.BodyProvideCapabilities, ProvideCapabilities: &capability.ProvideCapabilitiesR9IEs{ECID: &capability.ECIDProvideCapabilities{MeasurementSupport: capBits}}}}
	events, err := o.HandleInbound(cap, now)
	if err != nil {
		t.Fatal(err)
	}
	locationStart, err := m.Apply(request.Scope, events.Events, now)
	if err != nil {
		t.Fatal(err)
	}
	pci, _ := locationresult.NewPhysicalCellID(1)
	arfcn, _ := locationresult.NewEUTRAARFCN(100)
	measured, _ := locationresult.NewMeasuredResultsElement(pci, arfcn, locationresult.MeasuredResultsElementOptions{})
	signal, _ := location.NewECIDSignalMeasurementInformation(nil, []locationresult.MeasuredResultsElement{measured})
	ecid, _ := location.NewECIDProvideLocationInformation(signal)
	provide := lpp.Message{TransactionID: &lpp.TransactionID{Initiator: locationStart.Actions[0].Key.Initiator, TransactionNumber: locationStart.Actions[0].Key.Number}, Body: &lpp.Body{Kind: lpp.BodyProvideLocationInformation, ProvideLocationInformation: &location.ProvideLocationInformationR9IEs{ECID: &ecid}}}
	events, err = o.HandleInbound(provide, now)
	if err != nil {
		t.Fatal(err)
	}
	final, err := m.Apply(request.Scope, events.Events, now)
	if err != nil || final.Snapshot.State != QualityNotMet || final.Snapshot.Final == nil || final.Snapshot.Final.Kind != FinalQualityNotMet {
		t.Fatalf("final %#v %v", final, err)
	}
}
