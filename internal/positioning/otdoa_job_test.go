package positioning

import (
	"testing"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/capability"
	"github.com/vectorcore/esmlc/internal/lpp/location"
	locationresult "github.com/vectorcore/esmlc/internal/lpp/location/result"
	"github.com/vectorcore/esmlc/internal/uper"
)

func otdoaPolicy(t *testing.T, enabled bool) Policy {
	t.Helper()
	return Policy{OTDOA: OTDOAPolicy{Enabled: enabled}}
}

func TestJobStartsCapabilityThenEligibleOTDOA(t *testing.T) {
	now := time.Unix(0, 0)
	o := proc(t)
	m := New(otdoaPolicy(t, true))
	r, err := m.Start(req(), o, now)
	if err != nil || r.Snapshot.State != AwaitingCapabilities || r.Snapshot.Method != MethodOTDOA || len(r.Actions) != 1 || r.Actions[0].Message.Body.Kind != lpp.BodyRequestCapabilities || r.Actions[0].Message.Body.RequestCapabilities.OTDOA == nil {
		t.Fatalf("start %#v %v", r, err)
	}
	key := *r.Actions[0].Key
	modeBits, _ := uper.NewBitString([]byte{0x80}, 1)
	in := lpp.Message{TransactionID: &lpp.TransactionID{Initiator: key.Initiator, TransactionNumber: key.Number}, Body: &lpp.Body{Kind: lpp.BodyProvideCapabilities, ProvideCapabilities: &capability.ProvideCapabilitiesR9IEs{OTDOA: &capability.OTDOAProvideCapabilities{Mode: modeBits}}}}
	pr, err := o.HandleInbound(in, now)
	if err != nil {
		t.Fatal(err)
	}
	r, err = m.Apply(req().Scope, pr.Events, now)
	if err != nil || r.Snapshot.State != AwaitingLocationInformation || len(r.Actions) != 1 || r.Actions[0].Message.Body.Kind != lpp.BodyRequestLocationInformation || r.Actions[0].Message.Body.RequestLocationInformation.OTDOA == nil {
		t.Fatalf("apply %#v %v", r, err)
	}
	if r.Actions[0].Message.Body.RequestLocationInformation.OTDOA.AssistanceAvailability {
		t.Fatal("assistanceAvailability must stay false: no assistance-data source exists")
	}
}

func TestOTDOACapabilityWithoutUEAssistedModeIsNoEligibleMethod(t *testing.T) {
	now := time.Unix(0, 0)
	o := proc(t)
	m := New(otdoaPolicy(t, true))
	r, err := m.Start(req(), o, now)
	if err != nil {
		t.Fatal(err)
	}
	key := *r.Actions[0].Key
	// bit 1 (ue-assisted-NB) set, bit 0 (ue-assisted) clear: unsupported mode.
	modeBits, _ := uper.NewBitString([]byte{0x40}, 2)
	in := lpp.Message{TransactionID: &lpp.TransactionID{Initiator: key.Initiator, TransactionNumber: key.Number}, Body: &lpp.Body{Kind: lpp.BodyProvideCapabilities, ProvideCapabilities: &capability.ProvideCapabilitiesR9IEs{OTDOA: &capability.OTDOAProvideCapabilities{Mode: modeBits}}}}
	pr, err := o.HandleInbound(in, now)
	if err != nil {
		t.Fatal(err)
	}
	final, err := m.Apply(req().Scope, pr.Events, now)
	if err != nil || final.Snapshot.State != NoEligibleMethod {
		t.Fatalf("expected NoEligibleMethod, got %#v %v", final, err)
	}
}

func TestOTDOAMeasurementsBecomeTypedEstimatorUnavailableOutcome(t *testing.T) {
	now := time.Unix(0, 0)
	o := proc(t)
	m := New(otdoaPolicy(t, true))
	start, err := m.Start(req(), o, now)
	if err != nil {
		t.Fatal(err)
	}
	modeBits, _ := uper.NewBitString([]byte{0x80}, 1)
	capMsg := lpp.Message{TransactionID: &lpp.TransactionID{Initiator: start.Actions[0].Key.Initiator, TransactionNumber: start.Actions[0].Key.Number}, Body: &lpp.Body{Kind: lpp.BodyProvideCapabilities, ProvideCapabilities: &capability.ProvideCapabilitiesR9IEs{OTDOA: &capability.OTDOAProvideCapabilities{Mode: modeBits}}}}
	pr, err := o.HandleInbound(capMsg, now)
	if err != nil {
		t.Fatal(err)
	}
	locationStart, err := m.Apply(req().Scope, pr.Events, now)
	if err != nil {
		t.Fatal(err)
	}

	sfn, err := locationresult.NewSystemFrameNumberFromUint16(0)
	if err != nil {
		t.Fatal(err)
	}
	pci, err := locationresult.NewPhysicalCellID(5)
	if err != nil {
		t.Fatal(err)
	}
	resolution, _ := uper.NewBitString([]byte{0x40}, 2)
	value, _ := uper.NewBitString([]byte{0x00}, 5)
	quality, err := locationresult.NewOTDOAMeasQuality(resolution, value, nil)
	if err != nil {
		t.Fatal(err)
	}
	neighbourPCI, err := locationresult.NewPhysicalCellID(12)
	if err != nil {
		t.Fatal(err)
	}
	rstd, err := locationresult.NewRSTD(6356)
	if err != nil {
		t.Fatal(err)
	}
	neighbour, err := location.NewNeighbourMeasurementElement(neighbourPCI, nil, nil, rstd, quality)
	if err != nil {
		t.Fatal(err)
	}
	signal, err := location.NewOTDOASignalMeasurementInformation(sfn, pci, nil, nil, nil, []location.NeighbourMeasurementElement{neighbour})
	if err != nil {
		t.Fatal(err)
	}
	otdoaProvide, err := location.NewOTDOAProvideLocationInformation(&signal, nil)
	if err != nil {
		t.Fatal(err)
	}
	provide := lpp.Message{TransactionID: &lpp.TransactionID{Initiator: locationStart.Actions[0].Key.Initiator, TransactionNumber: locationStart.Actions[0].Key.Number}, Body: &lpp.Body{Kind: lpp.BodyProvideLocationInformation, ProvideLocationInformation: &location.ProvideLocationInformationR9IEs{OTDOA: &otdoaProvide}}}
	pr, err = o.HandleInbound(provide, now)
	if err != nil {
		t.Fatal(err)
	}
	final, err := m.Apply(req().Scope, pr.Events, now)
	if err != nil || final.Snapshot.State != EstimationUnavailable || final.Snapshot.Final == nil || final.Snapshot.Final.Kind != FinalMeasurementsWithoutEstimator || final.Snapshot.Final.MethodResult == nil || final.Snapshot.Final.MethodResult.OTDOA == nil {
		t.Fatalf("final %#v %v", final, err)
	}
	if len(final.Snapshot.Final.MethodResult.OTDOA.Signal.NeighbourMeasurements()) != 1 {
		t.Fatal("expected one preserved neighbour measurement")
	}
}

func TestECIDTakesPriorityWhenBothPoliciesEnabled(t *testing.T) {
	now := time.Unix(0, 0)
	o := proc(t)
	p := policy(t, true)
	p.OTDOA = OTDOAPolicy{Enabled: true}
	m := New(p)
	r, err := m.Start(req(), o, now)
	if err != nil || r.Snapshot.Method != MethodECID {
		t.Fatalf("expected ECID priority, got %#v %v", r, err)
	}
}
