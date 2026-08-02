package lppa

import (
	"testing"

	"github.com/vectorcore/esmlc/internal/aper"
)

func TestInitiationRequestRoundTrip(t *testing.T) {
	pdu, err := BuildInitiationRequest(7, 3, ReportOnDemand, nil, []MeasurementQuantityValue{QuantityCellID})
	if err != nil {
		t.Fatal(err)
	}
	wire, err := Encode(pdu)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(wire)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Category != Initiating || decoded.ProcedureCode != ProcedureECIDMeasurementInitiation || decoded.Criticality != aper.Reject || decoded.TransactionID != 7 {
		t.Fatalf("envelope mismatch: %+v", decoded)
	}
	if len(decoded.IEs) != 3 {
		t.Fatalf("expected 3 IEs, got %d", len(decoded.IEs))
	}
	mid, err := DecodeMeasurementID(decoded.IEs[0].Value)
	if err != nil || mid != 3 {
		t.Fatalf("measurement id: %d %v", mid, err)
	}
}

func TestInitiationRequestPeriodicityConditional(t *testing.T) {
	periodicity := PeriodicityMS1024
	if _, err := BuildInitiationRequest(1, 3, ReportOnDemand, &periodicity, []MeasurementQuantityValue{QuantityCellID}); err == nil {
		t.Fatal("accepted periodicity with onDemand")
	}
	if _, err := BuildInitiationRequest(1, 3, ReportPeriodic, nil, []MeasurementQuantityValue{QuantityCellID}); err == nil {
		t.Fatal("accepted periodic without periodicity")
	}
	if _, err := BuildInitiationRequest(1, 3, ReportPeriodic, &periodicity, []MeasurementQuantityValue{QuantityCellID}); err != nil {
		t.Fatalf("valid periodic request rejected: %v", err)
	}
}

func TestDecodeInitiationResponseRoundTrip(t *testing.T) {
	result := ECIDMeasurementResult{
		ServingCellID:  ECGI{PLMNIdentity: [3]byte{0x00, 0xf1, 0x10}, CellIdentity: 0x0abcdef},
		ServingCellTAC: [2]byte{0x10, 0x01},
	}
	resultBytes, err := EncodeECIDMeasurementResult(result)
	if err != nil {
		t.Fatal(err)
	}
	mid3, _ := EncodeMeasurementID(3)
	mid9, _ := EncodeMeasurementID(9)
	pdu := PDU{Category: Successful, ProcedureCode: ProcedureECIDMeasurementInitiation, Criticality: aper.Reject, TransactionID: 1, IEs: []IE{
		{IEESMLCMeasurementID, aper.Reject, mid3},
		{IEENBMeasurementID, aper.Reject, mid9},
		{IECIDMeasurementResult, aper.Ignore, resultBytes},
	}}
	resp, err := DecodeInitiationResponse(pdu)
	if err != nil {
		t.Fatal(err)
	}
	if resp.ESMLCMeasurementID != 3 || resp.ENBMeasurementID != 9 || resp.Result == nil {
		t.Fatalf("decoded response: %+v", resp)
	}
	if resp.Result.ServingCellID.CellIdentity != 0x0abcdef || resp.Result.AccessPointPosition != nil {
		t.Fatalf("decoded result: %+v", resp.Result)
	}
}

func TestDecodeInitiationResponseRejectsWrongEnvelope(t *testing.T) {
	mid3, _ := EncodeMeasurementID(3)
	mid9, _ := EncodeMeasurementID(9)
	base := func() PDU {
		return PDU{Category: Successful, ProcedureCode: ProcedureECIDMeasurementInitiation, Criticality: aper.Reject, TransactionID: 1, IEs: []IE{
			{IEESMLCMeasurementID, aper.Reject, mid3},
			{IEENBMeasurementID, aper.Reject, mid9},
		}}
	}
	for _, mutate := range []func(*PDU){
		func(p *PDU) { p.Category = Initiating },
		func(p *PDU) { p.ProcedureCode = ProcedureECIDMeasurementReport },
		func(p *PDU) { p.Criticality = aper.Ignore },
		func(p *PDU) { p.IEs = p.IEs[:1] },                                      // missing mandatory eNB measurement id
		func(p *PDU) { p.IEs = append(p.IEs, p.IEs[0]) },                        // duplicate
		func(p *PDU) { p.IEs = append(p.IEs, IE{999, aper.Reject, []byte{0}}) }, // unknown reject IE
	} {
		p := base()
		p.IEs = append([]IE(nil), p.IEs...)
		mutate(&p)
		if _, err := DecodeInitiationResponse(p); err == nil {
			t.Fatalf("accepted malformed response %+v", p)
		}
	}
	// Unknown ignore/notify IEs are tolerated.
	p := base()
	p.IEs = append(p.IEs, IE{999, aper.Ignore, []byte{0}})
	if _, err := DecodeInitiationResponse(p); err != nil {
		t.Fatalf("rejected unknown ignore IE: %v", err)
	}
}

func TestDecodeInitiationFailureAndFailureIndication(t *testing.T) {
	mid, _ := EncodeMeasurementID(4)
	mid2, _ := EncodeMeasurementID(5)
	cause, err := EncodeCause(Cause{CauseMisc, 0})
	if err != nil {
		t.Fatal(err)
	}
	failure := PDU{Category: Unsuccessful, ProcedureCode: ProcedureECIDMeasurementInitiation, Criticality: aper.Reject, TransactionID: 1, IEs: []IE{
		{IEESMLCMeasurementID, aper.Reject, mid},
		{IECause, aper.Ignore, cause},
	}}
	f, err := DecodeInitiationFailure(failure)
	if err != nil || f.ESMLCMeasurementID != 4 || f.Cause != (Cause{CauseMisc, 0}) {
		t.Fatalf("decoded failure %+v: %v", f, err)
	}

	indication := PDU{Category: Initiating, ProcedureCode: ProcedureECIDMeasurementFailureIndication, Criticality: aper.Ignore, TransactionID: 2, IEs: []IE{
		{IEESMLCMeasurementID, aper.Reject, mid},
		{IEENBMeasurementID, aper.Reject, mid2},
		{IECause, aper.Ignore, cause},
	}}
	fi, err := DecodeFailureIndication(indication)
	if err != nil || fi.ESMLCMeasurementID != 4 || fi.ENBMeasurementID != 5 {
		t.Fatalf("decoded failure indication %+v: %v", fi, err)
	}
}

func TestDecodeReportRequiresResult(t *testing.T) {
	mid, _ := EncodeMeasurementID(3)
	p := PDU{Category: Initiating, ProcedureCode: ProcedureECIDMeasurementReport, Criticality: aper.Ignore, TransactionID: 1, IEs: []IE{
		{IEESMLCMeasurementID, aper.Reject, mid},
		{IEENBMeasurementID, aper.Reject, mid},
	}}
	if _, err := DecodeReport(p); err == nil {
		t.Fatal("accepted report without mandatory E-CID-MeasurementResult")
	}
}

func TestTerminationCommandRoundTrip(t *testing.T) {
	pdu, err := BuildTerminationCommand(9, 3, 9)
	if err != nil {
		t.Fatal(err)
	}
	wire, err := Encode(pdu)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(wire)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.ProcedureCode != ProcedureECIDMeasurementTermination || decoded.Criticality != aper.Reject || len(decoded.IEs) != 2 {
		t.Fatalf("decoded termination %+v", decoded)
	}
}

func TestMeasuredResultsOutOfScopeRejected(t *testing.T) {
	// measuredResults optional-presence bit set (1) with no content: decoding
	// must fail closed rather than silently skip the unimplemented field.
	w := aper.NewWriter()
	_ = extBit(w, 0)
	_ = extBit(w, 0) // e-UTRANAccessPointPosition absent
	_ = extBit(w, 1) // measuredResults present
	_ = encodeECGI(w, ECGI{PLMNIdentity: [3]byte{0, 0xf1, 0x10}, CellIdentity: 1})
	_ = aper.PutFixedOctets(w, []byte{0, 1}, 2)
	if _, err := DecodeECIDMeasurementResult(w.Bytes()); err == nil {
		t.Fatal("accepted measuredResults presence bit")
	}
}

func TestCellPortionIDExtensionRangeRejected(t *testing.T) {
	w := aper.NewWriter()
	_ = extBit(w, 1) // extension bit set: value would be in the 256..4095 range
	if _, err := DecodeCellPortionID(w.Bytes()); err == nil {
		t.Fatal("accepted cell-portion-id extension range")
	}
	if _, err := EncodeCellPortionID(256); err == nil {
		t.Fatal("encoded cell-portion-id extension range value")
	}
}

func TestMalformedPDURejected(t *testing.T) {
	valid, err := BuildTerminationCommand(1, 3, 9)
	if err != nil {
		t.Fatal(err)
	}
	wire, err := Encode(valid)
	if err != nil {
		t.Fatal(err)
	}
	for n := 0; n < len(wire); n++ {
		if _, err := Decode(wire[:n]); err == nil {
			t.Fatalf("accepted truncated PDU at %d bytes", n)
		}
	}
	for _, b := range [][]byte{{}, {0xff}, {0, 0, 0xff}} {
		if _, err := Decode(b); err == nil {
			t.Fatalf("accepted malformed %x", b)
		}
	}
}

func TestCauseRootBranches(t *testing.T) {
	for _, c := range []Cause{
		{CauseRadioNetwork, 0}, {CauseRadioNetwork, 2},
		{CauseProtocol, 0}, {CauseProtocol, 6},
		{CauseMisc, 0},
	} {
		encoded, err := EncodeCause(c)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := DecodeCause(encoded)
		if err != nil || decoded != c {
			t.Fatalf("%#v => %x => %#v: %v", c, encoded, decoded, err)
		}
	}
	if _, err := EncodeCause(Cause{CauseMisc, 1}); err == nil {
		t.Fatal("accepted out-of-range cause value")
	}
}

func TestCriticalityDiagnosticsRoundTrip(t *testing.T) {
	proc := uint8(2)
	trigger := TriggerSuccessfulOutcome
	crit := aper.Reject
	txn := uint16(9)
	diag := CriticalityDiagnostics{
		ProcedureCode: &proc, TriggeringMessage: &trigger, ProcedureCriticality: &crit, TransactionID: &txn,
		IEs: []CriticalityDiagnosticsIEItem{{aper.Ignore, IECellPortionID, ErrorNotUnderstood}},
	}
	encoded, err := EncodeCriticalityDiagnostics(diag)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeCriticalityDiagnostics(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if *decoded.ProcedureCode != 2 || *decoded.TriggeringMessage != TriggerSuccessfulOutcome || *decoded.ProcedureCriticality != aper.Reject || *decoded.TransactionID != 9 {
		t.Fatalf("decoded diagnostics %+v", decoded)
	}
	if len(decoded.IEs) != 1 || decoded.IEs[0].IEID != IECellPortionID || decoded.IEs[0].TypeOfError != ErrorNotUnderstood {
		t.Fatalf("decoded diagnostics IE list %+v", decoded.IEs)
	}
}

func TestMeasurementQuantitiesRoundTrip(t *testing.T) {
	values := []MeasurementQuantityValue{QuantityCellID, QuantityRSRP, QuantityRSRQ}
	encoded, err := EncodeMeasurementQuantities(values)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeMeasurementQuantities(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != len(values) {
		t.Fatalf("decoded %v want %v", decoded, values)
	}
	for i, v := range values {
		if decoded[i] != v {
			t.Fatalf("decoded[%d]=%v want %v", i, decoded[i], v)
		}
	}
}
