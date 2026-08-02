package procedure

import (
	"errors"
	"testing"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/capability"
	"github.com/vectorcore/esmlc/internal/lpp/location"
	"github.com/vectorcore/esmlc/internal/lpp/transaction"
	"github.com/vectorcore/esmlc/internal/uper"
)

func TestAGNSSCapabilityProcedureFlow(t *testing.T) {
	o := newOrch(t)
	r, e := o.StartCapabilities(StartOptions{RequestAGNSS: true}, time.Time{})
	if e != nil || r.Actions[0].Message.Body.RequestCapabilities.AGNSS == nil || !r.Actions[0].Message.Body.RequestCapabilities.AGNSS.GNSSSupportListReq {
		t.Fatalf("start %#v %v", r, e)
	}
	m := request(30, lpp.BodyRequestCapabilities)
	m.Body.RequestCapabilities = &capability.RequestCapabilitiesR9IEs{AGNSS: &capability.AGNSSRequestCapabilities{GNSSSupportListReq: true}}
	r, e = o.HandleInbound(m, time.Time{})
	if e != nil || r.Events[0].CapabilityRequest == nil || r.Events[0].CapabilityRequest.AGNSS == nil {
		t.Fatalf("inbound %#v %v", r, e)
	}
	modes, _ := uper.NewBitString([]byte{0x40}, 2)
	signals, _ := uper.NewBitString([]byte{0x80}, 8)
	k := transaction.Key{Initiator: lpp.InitiatorTargetDevice, Number: 30}
	provided := capability.ProvideCapabilitiesR9IEs{AGNSS: &capability.AGNSSProvideCapabilities{GNSSSupportList: []capability.GNSSSupportElement{{ID: capability.GNSSIDGPS, Modes: modes, Signals: signals}}}}
	r, e = o.ProvideCapabilitiesResult(k, ProvideCapabilitiesOptions{Capabilities: provided}, time.Time{})
	if e != nil || !r.Actions[0].Message.Body.ProvideCapabilities.AGNSS.SupportsGPSUEBased() {
		t.Fatalf("provide %#v %v", r, e)
	}
}

func TestAGNSSUnrequestedRejected(t *testing.T) {
	o := newOrch(t)
	m := request(31, lpp.BodyRequestCapabilities)
	_, e := o.HandleInbound(m, time.Time{})
	if e != nil {
		t.Fatal(e)
	}
	modes, _ := uper.NewBitString([]byte{0x40}, 2)
	signals, _ := uper.NewBitString([]byte{0x80}, 8)
	provided := capability.ProvideCapabilitiesR9IEs{AGNSS: &capability.AGNSSProvideCapabilities{GNSSSupportList: []capability.GNSSSupportElement{{ID: capability.GNSSIDGPS, Modes: modes, Signals: signals}}}}
	_, e = o.ProvideCapabilitiesResult(transaction.Key{Initiator: lpp.InitiatorTargetDevice, Number: 31}, ProvideCapabilitiesOptions{Capabilities: provided}, time.Time{})
	if !errors.Is(e, ErrUnrequestedAGNSS) {
		t.Fatal(e)
	}
}

func TestStartAGNSSLocationInformation(t *testing.T) {
	o := newOrch(t)
	common := &location.CommonRequestLocationInformation{LocationInformationType: location.LocationEstimateRequired}
	gnssMethods, _ := uper.NewBitString([]byte{0x80}, 1)
	agnss := &location.AGNSSRequestLocationInformation{GNSSMethods: gnssMethods}
	r, err := o.StartLocationInformation(StartLocationInformationOptions{Common: common, AGNSS: agnss}, time.Time{})
	if err != nil || len(r.Actions) != 1 {
		t.Fatalf("start %#v: %v", r, err)
	}
	m := r.Actions[0].Message
	if m.Body == nil || m.Body.RequestLocationInformation == nil || m.Body.RequestLocationInformation.Common == nil || m.Body.RequestLocationInformation.Common.LocationInformationType != location.LocationEstimateRequired {
		t.Fatal("missing typed common request")
	}
	if m.Body.RequestLocationInformation.AGNSS == nil {
		t.Fatal("missing typed A-GNSS request")
	}
	if r.Events[0].LocationRequest == nil || r.Events[0].LocationRequest.Common == nil || r.Events[0].LocationRequest.AGNSS == nil {
		t.Fatal("event missing typed common/A-GNSS request")
	}
}
