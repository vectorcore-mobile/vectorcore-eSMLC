package procedure

import (
	"errors"
	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/capability"
	"github.com/vectorcore/esmlc/internal/lpp/transaction"
	"github.com/vectorcore/esmlc/internal/uper"
	"testing"
	"time"
)

func TestECIDCapabilityProcedureFlow(t *testing.T) {
	o := newOrch(t)
	r, e := o.StartCapabilities(StartOptions{RequestECID: true}, time.Time{})
	if e != nil || r.Actions[0].Message.Body.RequestCapabilities.ECID == nil {
		t.Fatalf("start %#v %v", r, e)
	}
	m := request(22, lpp.BodyRequestCapabilities)
	m.Body.RequestCapabilities = &capability.RequestCapabilitiesR9IEs{ECID: &capability.ECIDRequestCapabilities{}}
	r, e = o.HandleInbound(m, time.Time{})
	if e != nil || r.Events[0].CapabilityRequest == nil || r.Events[0].CapabilityRequest.ECID == nil {
		t.Fatalf("inbound %#v %v", r, e)
	}
	v, _ := uper.NewBitString([]byte{0xe0}, 3)
	k := transaction.Key{Initiator: lpp.InitiatorTargetDevice, Number: 22}
	r, e = o.ProvideCapabilitiesResult(k, ProvideCapabilitiesOptions{Capabilities: capability.ProvideCapabilitiesR9IEs{ECID: &capability.ECIDProvideCapabilities{MeasurementSupport: v}}}, time.Time{})
	if e != nil || r.Actions[0].Message.Body.ProvideCapabilities.ECID.MeasurementSupport.BitLen() != 3 {
		t.Fatalf("provide %#v %v", r, e)
	}
}
func TestECIDUnrequestedRejected(t *testing.T) {
	o := newOrch(t)
	m := request(23, lpp.BodyRequestCapabilities)
	_, e := o.HandleInbound(m, time.Time{})
	if e != nil {
		t.Fatal(e)
	}
	v, _ := uper.NewBitString([]byte{0x80}, 1)
	_, e = o.ProvideCapabilitiesResult(transaction.Key{Initiator: lpp.InitiatorTargetDevice, Number: 23}, ProvideCapabilitiesOptions{Capabilities: capability.ProvideCapabilitiesR9IEs{ECID: &capability.ECIDProvideCapabilities{MeasurementSupport: v}}}, time.Time{})
	if !errors.Is(e, ErrUnrequestedECID) {
		t.Fatal(e)
	}
}
