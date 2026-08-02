package procedure

import (
	"errors"
	"testing"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/capability"
	"github.com/vectorcore/esmlc/internal/lpp/transaction"
	"github.com/vectorcore/esmlc/internal/uper"
)

func TestOTDOACapabilityProcedureFlow(t *testing.T) {
	o := newOrch(t)
	r, e := o.StartCapabilities(StartOptions{RequestOTDOA: true}, time.Time{})
	if e != nil || r.Actions[0].Message.Body.RequestCapabilities.OTDOA == nil {
		t.Fatalf("start %#v %v", r, e)
	}
	m := request(24, lpp.BodyRequestCapabilities)
	m.Body.RequestCapabilities = &capability.RequestCapabilitiesR9IEs{OTDOA: &capability.OTDOARequestCapabilities{}}
	r, e = o.HandleInbound(m, time.Time{})
	if e != nil || r.Events[0].CapabilityRequest == nil || r.Events[0].CapabilityRequest.OTDOA == nil {
		t.Fatalf("inbound %#v %v", r, e)
	}
	v, _ := uper.NewBitString([]byte{0x80}, 1)
	k := transaction.Key{Initiator: lpp.InitiatorTargetDevice, Number: 24}
	r, e = o.ProvideCapabilitiesResult(k, ProvideCapabilitiesOptions{Capabilities: capability.ProvideCapabilitiesR9IEs{OTDOA: &capability.OTDOAProvideCapabilities{Mode: v}}}, time.Time{})
	if e != nil || !r.Actions[0].Message.Body.ProvideCapabilities.OTDOA.SupportsUEAssisted() {
		t.Fatalf("provide %#v %v", r, e)
	}
}

func TestOTDOAUnrequestedRejected(t *testing.T) {
	o := newOrch(t)
	m := request(25, lpp.BodyRequestCapabilities)
	_, e := o.HandleInbound(m, time.Time{})
	if e != nil {
		t.Fatal(e)
	}
	v, _ := uper.NewBitString([]byte{0x80}, 1)
	_, e = o.ProvideCapabilitiesResult(transaction.Key{Initiator: lpp.InitiatorTargetDevice, Number: 25}, ProvideCapabilitiesOptions{Capabilities: capability.ProvideCapabilitiesR9IEs{OTDOA: &capability.OTDOAProvideCapabilities{Mode: v}}}, time.Time{})
	if !errors.Is(e, ErrUnrequestedOTDOA) {
		t.Fatal(e)
	}
}
