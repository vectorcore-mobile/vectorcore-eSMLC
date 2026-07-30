package lcsap

import (
	"bytes"
	"github.com/vectorcore/esmlc/internal/aper"
	"testing"
)

func request() PDU {
	return PDU{Category: Initiating, Procedure: ProcedureLocationRequest, Criticality: aper.Reject, IEs: []IE{{IECorrelationID, aper.Reject, []byte{0, 0, 0, 1}}, {IELocationType, aper.Reject, []byte{0}}, {IEECGI, aper.Ignore, []byte{0, 0xf1, 0x10, 0, 0, 0, 1}}}}
}
func TestMMECompatibleLocationRequestRoundTrip(t *testing.T) {
	w, e := Encode(request())
	if e != nil {
		t.Fatal(e)
	}
	p, e := Decode(w)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = ValidateLocationRequest(p); e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(p.IEs[2].Value, request().IEs[2].Value) {
		t.Fatal("ECGI changed")
	}
}

func TestLocationRequestPreservesConditionalPriority(t *testing.T) {
	p := request()
	p.IEs = append(p.IEs, IE{IELCSPriority, aper.Ignore, []byte{0}})
	v, err := DecodeLocationRequest(p)
	if err != nil || v.Priority == nil || *v.Priority != 0 || v.LocationType != 0 {
		t.Fatalf("decoded %#v: %v", v, err)
	}
	p.IEs[len(p.IEs)-1].Value = []byte{0, 1}
	if _, err := DecodeLocationRequest(p); err == nil {
		t.Fatal("invalid priority accepted")
	}
}

func TestLocationRequestPreservesQoSAndLPPAbility(t *testing.T) {
	accuracy, delay := uint8(4), uint8(0)
	qos, err := EncodeQoS(QoS{HorizontalAccuracy: &accuracy, ResponseTime: &delay})
	if err != nil {
		t.Fatal(err)
	}
	capability, err := EncodeUEPositioningCapability(false)
	if err != nil {
		t.Fatal(err)
	}
	p := request()
	p.IEs = append(p.IEs, IE{IELCSQoS, aper.Ignore, qos}, IE{IEUEPositioningCapability, aper.Ignore, capability})
	v, err := DecodeLocationRequest(p)
	if err != nil || v.QoS == nil || v.QoS.HorizontalAccuracy == nil || *v.QoS.HorizontalAccuracy != 4 || v.QoS.ResponseTime == nil || *v.QoS.ResponseTime != 0 || v.LPPSupported == nil || *v.LPPSupported {
		t.Fatalf("decoded %#v: %v", v, err)
	}
}

func TestSupportedLCSCauseEncodings(t *testing.T) {
	for cause, wire := range map[Cause]byte{CauseRadioNetworkUnspecified: 0x00, CauseProtocolUnspecified: 0x94, CauseMiscUnspecified: 0xcc} {
		encoded, err := FailureWithCause([4]byte{1}, cause)
		if err != nil {
			t.Fatal(err)
		}
		p, err := Decode(encoded)
		if err != nil {
			t.Fatal(err)
		}
		if len(p.IEs) != 2 || p.IEs[1].ID != IELCSCause || len(p.IEs[1].Value) != 1 || p.IEs[1].Value[0] != wire {
			t.Fatalf("cause %d encoded %#v", cause, p)
		}
	}
}
func TestLocationResponseAndMalformedInput(t *testing.T) {
	id := [4]byte{0, 0, 0, 1}
	w, e := LocationResponse(id, 38, -90, 4)
	if e != nil {
		t.Fatal(e)
	}
	p, e := Decode(w)
	if e != nil {
		t.Fatal(e)
	}
	got, e := Correlation(p)
	if e != nil || got != id || p.Category != Successful || len(p.IEs) != 2 {
		t.Fatalf("bad response %#v %v", p, e)
	}
	for _, b := range [][]byte{{}, {0xff}, {0, 0, 0xff}} {
		if _, e := Decode(b); e == nil {
			t.Fatalf("accepted malformed %x", b)
		}
	}
}
func TestUnknownRejectAndDuplicateRejected(t *testing.T) {
	p := request()
	p.IEs = append(p.IEs, IE{999, aper.Reject, []byte{1}})
	if _, e := ValidateLocationRequest(p); e == nil {
		t.Fatal("unknown reject IE accepted")
	}
	p = request()
	p.IEs = append(p.IEs, p.IEs[0])
	if _, e := ValidateLocationRequest(p); e == nil {
		t.Fatal("duplicate accepted")
	}
}

func TestConnectionOrientedMMEContract(t *testing.T) {
	in := ConnectionOriented{Correlation: [4]byte{0, 0, 0, 9}, PayloadType: 0, Payload: []byte{0xaa, 0x55}}
	w, err := EncodeConnectionOriented(in, 32)
	if err != nil {
		t.Fatal(err)
	}
	p, err := Decode(w)
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeConnectionOriented(p, 32)
	if err != nil {
		t.Fatal(err)
	}
	if out.Correlation != in.Correlation || out.PayloadType != 0 || !bytes.Equal(out.Payload, in.Payload) {
		t.Fatalf("round trip: %#v", out)
	}
	if p.Procedure != ProcedureConnectionOrientedInformation || p.Criticality != aper.Reject {
		t.Fatal("TS 29.171 Connection-Oriented criticality changed")
	}
	// The deployed MME currently emits ignore. Decode remains explicitly
	// compatible while its encoder is corrected independently.
	p.Criticality = aper.Ignore
	if _, err := DecodeConnectionOriented(p, 32); err != nil {
		t.Fatalf("MME compatibility: %v", err)
	}
	for _, mutate := range []func(*PDU){
		func(p *PDU) { p.IEs = p.IEs[:2] },
		func(p *PDU) { p.IEs = append(p.IEs, p.IEs[0]) },
		func(p *PDU) { p.IEs[1].Value = []byte{2} },
	} {
		bad := p
		bad.IEs = append([]IE(nil), p.IEs...)
		mutate(&bad)
		if _, err := DecodeConnectionOriented(bad, 32); err == nil {
			t.Fatal("invalid connection-oriented PDU accepted")
		}
	}
}
