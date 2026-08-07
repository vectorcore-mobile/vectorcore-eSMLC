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
	p.IEs = append(p.IEs, IE{IELCSPriority, aper.Reject, []byte{0}})
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
	p.IEs = append(p.IEs, IE{IELCSQoS, aper.Reject, qos}, IE{IEUEPositioningCapability, aper.Reject, capability})
	v, err := DecodeLocationRequest(p)
	if err != nil || v.QoS == nil || v.QoS.HorizontalAccuracy == nil || *v.QoS.HorizontalAccuracy != 4 || v.QoS.ResponseTime == nil || *v.QoS.ResponseTime != 0 || v.LPPSupported == nil || *v.LPPSupported {
		t.Fatalf("decoded %#v: %v", v, err)
	}
}

func TestLocationRequestPreservesConditionalClientType(t *testing.T) {
	for _, ct := range []ClientType{ClientTypeEmergencyServices, ClientTypeValueAddedServices, ClientTypePLMNOperatorTargetMSServiceSupport} {
		wire, err := EncodeLCSClientType(ct)
		if err != nil {
			t.Fatal(err)
		}
		p := request()
		p.IEs = append(p.IEs, IE{IELCSClientType, aper.Reject, wire})
		v, err := DecodeLocationRequest(p)
		if err != nil || v.ClientType == nil || *v.ClientType != ct {
			t.Fatalf("client type %d: decoded %#v: %v", ct, v, err)
		}
	}
	p := request()
	if _, err := EncodeLCSClientType(ClientTypePLMNOperatorTargetMSServiceSupport + 1); err == nil {
		t.Fatal("out-of-range client type accepted")
	}
	p.IEs = append(p.IEs, IE{IELCSClientType, aper.Reject, []byte{0x80}})
	if _, err := DecodeLocationRequest(p); err == nil {
		t.Fatal("extension-marked client type accepted")
	}
	if v, err := DecodeLocationRequest(request()); err != nil || v.ClientType != nil {
		t.Fatalf("client type should be nil when IE absent: %#v: %v", v, err)
	}
}

func TestSupportedLCSCauseEncodings(t *testing.T) {
	for cause, wire := range map[Cause]byte{CauseRadioNetworkUnspecified: 0x00, CauseProtocolUnspecified: 0x94, CauseMiscUnspecified: 0xd8} {
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

func TestLCSCauseRootBranches(t *testing.T) {
	for _, cause := range []LCSCause{
		{LCSCauseRadioNetwork, RadioNetworkUnspecified},
		{LCSCauseTransport, TransportResourceUnavailable},
		{LCSCauseTransport, TransportUnspecified},
		{LCSCauseProtocol, ProtocolSemanticError},
		{LCSCauseMisc, MiscHardwareFailure},
	} {
		encoded, err := EncodeLCSCause(cause)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := DecodeLCSCause(encoded)
		if err != nil || decoded != cause {
			t.Fatalf("%#v => %x => %#v: %v", cause, encoded, decoded, err)
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

func TestIndependentRootResultMetadataFixtures(t *testing.T) {
	data := NewECIDPositioningData()
	encoded, err := data.EncodeAPER()
	if err != nil || !bytes.Equal(encoded, []byte{0x40, 0x13}) {
		t.Fatalf("positioning data %x %v", encoded, err)
	}
	decoded, err := DecodePositioningData(encoded)
	if err != nil || !bytes.Equal(decoded.Methods(), []byte{0x13}) {
		t.Fatalf("positioning data decode %#v %v", decoded, err)
	}
	for value, fixture := range map[AccuracyFulfillmentIndicator]byte{AccuracyFulfilled: 0x00, AccuracyNotFulfilled: 0x40} {
		encoded, err = EncodeAccuracyFulfillmentIndicator(value)
		if err != nil || len(encoded) != 1 || encoded[0] != fixture {
			t.Fatalf("accuracy %d: %x %v", value, encoded, err)
		}
		if decodedValue, err := DecodeAccuracyFulfillmentIndicator(encoded); err != nil || decodedValue != value {
			t.Fatalf("accuracy decode %d %v", decodedValue, err)
		}
	}
	if _, err := DecodeAccuracyFulfillmentIndicator([]byte{0x01}); err == nil {
		t.Fatal("accepted non-zero APER padding")
	}
	response, err := LocationResponseWithMetadata([4]byte{1}, 38, -90, 40, &data, ptrAccuracy(AccuracyFulfilled))
	if err != nil {
		t.Fatal(err)
	}
	p, err := Decode(response)
	if err != nil || len(p.IEs) != 4 || p.IEs[2].ID != IEPositioningData || p.IEs[3].ID != IEAccuracyFulfillmentIndicator {
		t.Fatalf("response %#v %v", p, err)
	}
}

func ptrAccuracy(v AccuracyFulfillmentIndicator) *AccuracyFulfillmentIndicator { return &v }
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

func TestMalformedLocationRequestRejectedBeforeDispatch(t *testing.T) {
	for _, mutate := range []func(*PDU){
		func(p *PDU) { p.IEs = p.IEs[:2] },                                      // missing mandatory ECGI
		func(p *PDU) { p.IEs = append(p.IEs, p.IEs[0]) },                        // duplicate mandatory IE
		func(p *PDU) { p.IEs[1].Criticality = aper.Ignore },                     // reject-criticality mismatch
		func(p *PDU) { p.IEs = append(p.IEs, IE{999, aper.Reject, []byte{1}}) }, // unknown reject IE
	} {
		p := request()
		p.IEs = append([]IE(nil), p.IEs...)
		mutate(&p)
		wire, err := Encode(p)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := Decode(wire)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = DecodeLocationRequest(decoded); err == nil {
			t.Fatal("malformed Location Request accepted")
		}
	}
	valid, err := Encode(request())
	if err != nil {
		t.Fatal(err)
	}
	for n := 0; n < len(valid); n++ {
		if _, err := Decode(valid[:n]); err == nil {
			t.Fatalf("accepted truncated Location Request at %d bytes", n)
		}
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
