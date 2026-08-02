package procedure

import (
	"testing"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/location"
	"github.com/vectorcore/esmlc/internal/lpp/transaction"
)

func TestStartOTDOALocationInformation(t *testing.T) {
	o := newOrch(t)
	req := &location.OTDOARequestLocationInformation{AssistanceAvailability: false}
	r, err := o.StartLocationInformation(StartLocationInformationOptions{OTDOA: req}, time.Time{})
	if err != nil || len(r.Actions) != 1 || len(r.Events) != 1 {
		t.Fatalf("start %#v: %v", r, err)
	}
	m := r.Actions[0].Message
	if m.Body == nil || m.Body.Kind != lpp.BodyRequestLocationInformation || m.Body.RequestLocationInformation == nil || m.Body.RequestLocationInformation.OTDOA == nil {
		t.Fatal("missing typed OTDOA request")
	}
	if m.Body.RequestLocationInformation.OTDOA.AssistanceAvailability {
		t.Fatal("assistanceAvailability changed")
	}
	if r.Events[0].LocationRequest == nil || r.Events[0].LocationRequest.OTDOA == nil {
		t.Fatal("event missing typed OTDOA request")
	}
}

func TestInboundOTDOALocationRequestPending(t *testing.T) {
	o := newOrch(t)
	payload := location.RequestLocationInformationR9IEs{OTDOA: &location.OTDOARequestLocationInformation{AssistanceAvailability: true}}
	m := request(60, lpp.BodyRequestLocationInformation)
	m.Body.RequestLocationInformation = &payload
	r, err := o.HandleInbound(m, time.Time{})
	if err != nil || len(r.Events) != 2 || r.Events[0].Kind != LocationInformationRequested || r.Events[1].Kind != AwaitingApplicationResult {
		t.Fatalf("inbound %#v: %v", r, err)
	}
	if r.Events[0].LocationRequest == nil || r.Events[0].LocationRequest.OTDOA == nil || !r.Events[0].LocationRequest.OTDOA.AssistanceAvailability {
		t.Fatal("missing typed OTDOA event")
	}
	k := transaction.Key{Initiator: lpp.InitiatorTargetDevice, Number: 60}
	s, ok := o.Snapshot(k)
	if !ok || s.Waiting != WaitLocationInformation || s.LocationRequest == nil || s.LocationRequest.OTDOA == nil {
		t.Fatalf("bad snapshot %#v", s)
	}
}
