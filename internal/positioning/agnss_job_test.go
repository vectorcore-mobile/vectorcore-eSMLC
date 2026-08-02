package positioning

import (
	"testing"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/capability"
	"github.com/vectorcore/esmlc/internal/lpp/location"
	"github.com/vectorcore/esmlc/internal/uper"
)

func agnssPolicy(t *testing.T, enabled bool) Policy {
	t.Helper()
	return Policy{AGNSS: AGNSSPolicy{Enabled: enabled}}
}

func gpsUEBasedCapabilityMessage(t *testing.T, initiator lpp.Initiator, number uint8) lpp.Message {
	t.Helper()
	modes, err := uper.NewBitString([]byte{0x40}, 2) // bit1 (ue-based) set
	if err != nil {
		t.Fatal(err)
	}
	signals, err := uper.NewBitString([]byte{0x80}, 8)
	if err != nil {
		t.Fatal(err)
	}
	return lpp.Message{TransactionID: &lpp.TransactionID{Initiator: initiator, TransactionNumber: number}, Body: &lpp.Body{Kind: lpp.BodyProvideCapabilities, ProvideCapabilities: &capability.ProvideCapabilitiesR9IEs{AGNSS: &capability.AGNSSProvideCapabilities{
		GNSSSupportList: []capability.GNSSSupportElement{{ID: capability.GNSSIDGPS, Modes: modes, Signals: signals}},
	}}}}
}

func TestJobStartsCapabilityThenEligibleAGNSS(t *testing.T) {
	now := time.Unix(0, 0)
	o := proc(t)
	m := New(agnssPolicy(t, true))
	r, err := m.Start(req(), o, now)
	if err != nil || r.Snapshot.State != AwaitingCapabilities || r.Snapshot.Method != MethodAGNSS || len(r.Actions) != 1 || r.Actions[0].Message.Body.RequestCapabilities.AGNSS == nil {
		t.Fatalf("start %#v %v", r, err)
	}
	key := *r.Actions[0].Key
	in := gpsUEBasedCapabilityMessage(t, key.Initiator, key.Number)
	pr, err := o.HandleInbound(in, now)
	if err != nil {
		t.Fatal(err)
	}
	r, err = m.Apply(req().Scope, pr.Events, now)
	if err != nil || r.Snapshot.State != AwaitingLocationInformation || len(r.Actions) != 1 {
		t.Fatalf("apply %#v %v", r, err)
	}
	loc := r.Actions[0].Message.Body.RequestLocationInformation
	if loc == nil || loc.Common == nil || loc.Common.LocationInformationType != location.LocationEstimateRequired || loc.AGNSS == nil {
		t.Fatalf("bad location request: %#v", loc)
	}
}

// TestJobDeliversAGNSSEstimateThroughCombinedEstimator drives a full
// Manager.Start/Apply cycle where the UE reports its own already-computed
// GPS position via the common location estimate, proving the pass-through
// AGNSSEstimator is correctly wired end to end (not just correct in
// isolation).
func TestJobDeliversAGNSSEstimateThroughCombinedEstimator(t *testing.T) {
	now := time.Unix(0, 0)
	o := proc(t)
	m := NewWithEstimator(agnssPolicy(t, true), CombinedEstimator{AGNSS: AGNSSEstimator{}})
	start, err := m.Start(req(), o, now)
	if err != nil {
		t.Fatal(err)
	}
	capMsg := gpsUEBasedCapabilityMessage(t, start.Actions[0].Key.Initiator, start.Actions[0].Key.Number)
	pr, err := o.HandleInbound(capMsg, now)
	if err != nil {
		t.Fatal(err)
	}
	locStart, err := m.Apply(req().Scope, pr.Events, now)
	if err != nil {
		t.Fatal(err)
	}

	estimate := location.LocationCoordinates{Shape: location.ShapePointWithUncertaintyCircle, Point: location.Coordinates{Latitude: 38.0, Longitude: -90.0}, UncertaintyCircle: 30}
	provide := lpp.Message{TransactionID: &lpp.TransactionID{Initiator: locStart.Actions[0].Key.Initiator, TransactionNumber: locStart.Actions[0].Key.Number}, Body: &lpp.Body{Kind: lpp.BodyProvideLocationInformation, ProvideLocationInformation: &location.ProvideLocationInformationR9IEs{
		Common: &location.CommonProvideLocationInformation{LocationEstimate: &estimate},
	}}}
	pr, err = o.HandleInbound(provide, now)
	if err != nil {
		t.Fatal(err)
	}
	final, err := m.Apply(req().Scope, pr.Events, now)
	if err != nil {
		t.Fatal(err)
	}
	if final.Snapshot.State != EstimateAvailable || final.Snapshot.Final == nil || final.Snapshot.Final.Kind != FinalEstimateAvailable || final.Snapshot.Final.Estimate == nil {
		t.Fatalf("expected a delivered A-GNSS estimate, got %#v", final.Snapshot.Final)
	}
	if final.Snapshot.Final.Estimate.Source != EstimateSourceAGNSSUEReported {
		t.Fatalf("wrong estimate source: %v", final.Snapshot.Final.Estimate.Source)
	}
	if final.Snapshot.Final.Estimate.Latitude != 38.0 || final.Snapshot.Final.Estimate.Longitude != -90.0 {
		t.Fatalf("estimate coordinates changed: %#v", final.Snapshot.Final.Estimate)
	}
}
