package sls

import (
	"testing"

	"github.com/vectorcore/esmlc/internal/aper"
	"github.com/vectorcore/esmlc/internal/config"
	"github.com/vectorcore/esmlc/internal/lcsap"
	"github.com/vectorcore/esmlc/internal/lppa"
)

// TestLPPaECIDDeliversEstimateViaAccessPointPosition drives a full
// Handle()-level cycle: a Location Request with LPPaECID enabled must send
// an LPPa E-CID Measurement Initiation Request (Connection-Oriented
// Information payload type 1) instead of any LPP capability exchange; once
// the eNB's Initiation Response reports its own antenna position, the
// server must both deliver the LCS-AP estimate and terminate the eNB
// measurement it started.
func TestLPPaECIDDeliversEstimateViaAccessPointPosition(t *testing.T) {
	c := config.Default()
	c.Positioning.LPPaECID.Enabled = true
	s := New(c, nil)
	out, err := s.Handle("mme-a", locRequest())
	if err != nil || len(out) != 1 {
		t.Fatalf("start %d %v", len(out), err)
	}
	p, err := lcsap.Decode(out[0])
	if err != nil || p.Procedure != lcsap.ProcedureConnectionOrientedInformation {
		t.Fatalf("bad initial procedure: %#v %v", p, err)
	}
	carrier, err := lcsap.DecodeConnectionOriented(p, c.SLs.MaxMessageSize)
	if err != nil || carrier.PayloadType != 1 {
		t.Fatalf("expected LPPa payload type 1: %#v %v", carrier, err)
	}
	initiation, err := lppa.Decode(carrier.Payload)
	if err != nil || initiation.Category != lppa.Initiating || initiation.ProcedureCode != lppa.ProcedureECIDMeasurementInitiation || initiation.Criticality != aper.Reject {
		t.Fatalf("bad LPPa initiation request: %#v %v", initiation, err)
	}

	result := lppa.ECIDMeasurementResult{
		ServingCellID:       lppa.ECGI{PLMNIdentity: [3]byte{0x00, 0xf1, 0x10}, CellIdentity: 1},
		ServingCellTAC:      [2]byte{0, 1},
		AccessPointPosition: &lppa.AccessPointPosition{Latitude: 38, Longitude: -90, Confidence: 68, UncertaintySemiMajor: 10},
	}
	resultBytes, err := lppa.EncodeECIDMeasurementResult(result)
	if err != nil {
		t.Fatal(err)
	}
	mid1, _ := lppa.EncodeMeasurementID(1)
	mid9, _ := lppa.EncodeMeasurementID(9)
	respPDU := lppa.PDU{Category: lppa.Successful, ProcedureCode: lppa.ProcedureECIDMeasurementInitiation, Criticality: aper.Reject, TransactionID: initiation.TransactionID, IEs: []lppa.IE{
		{ID: lppa.IEESMLCMeasurementID, Criticality: aper.Reject, Value: mid1},
		{ID: lppa.IEENBMeasurementID, Criticality: aper.Reject, Value: mid9},
		{ID: lppa.IECIDMeasurementResult, Criticality: aper.Ignore, Value: resultBytes},
	}}
	respWire, err := lppa.Encode(respPDU)
	if err != nil {
		t.Fatal(err)
	}
	w, err := lcsap.EncodeConnectionOriented(lcsap.ConnectionOriented{Correlation: carrier.Correlation, PayloadType: 1, Payload: respWire}, c.SLs.MaxMessageSize)
	if err != nil {
		t.Fatal(err)
	}
	out, err = s.Handle("mme-a", w)
	if err != nil || len(out) != 2 {
		t.Fatalf("response %d %v", len(out), err)
	}
	termP, err := lcsap.Decode(out[0])
	if err != nil {
		t.Fatal(err)
	}
	termCarrier, err := lcsap.DecodeConnectionOriented(termP, c.SLs.MaxMessageSize)
	if err != nil || termCarrier.PayloadType != 1 {
		t.Fatalf("expected termination carrier: %#v %v", termCarrier, err)
	}
	termination, err := lppa.Decode(termCarrier.Payload)
	if err != nil || termination.ProcedureCode != lppa.ProcedureECIDMeasurementTermination {
		t.Fatalf("expected LPPa termination command: %#v %v", termination, err)
	}
	finalP, err := lcsap.Decode(out[1])
	if err != nil || finalP.Category != lcsap.Successful || finalP.Procedure != lcsap.ProcedureLocationRequest {
		t.Fatalf("expected delivered LCS-AP estimate: %#v %v", finalP, err)
	}
}

// TestLPPaECIDAwaitsUnsolicitedReport covers the case where the Initiation
// Response merely acknowledges the request (no result yet): the server must
// stay silent until an unsolicited Report arrives, then deliver the estimate
// and terminate the measurement exactly as it would after an immediate result.
func TestLPPaECIDAwaitsUnsolicitedReport(t *testing.T) {
	c := config.Default()
	c.Positioning.LPPaECID.Enabled = true
	s := New(c, nil)
	out, err := s.Handle("mme-a", locRequest())
	if err != nil || len(out) != 1 {
		t.Fatalf("start %d %v", len(out), err)
	}
	p, _ := lcsap.Decode(out[0])
	carrier, _ := lcsap.DecodeConnectionOriented(p, c.SLs.MaxMessageSize)
	initiation, _ := lppa.Decode(carrier.Payload)

	mid1, _ := lppa.EncodeMeasurementID(1)
	mid9, _ := lppa.EncodeMeasurementID(9)
	ackPDU := lppa.PDU{Category: lppa.Successful, ProcedureCode: lppa.ProcedureECIDMeasurementInitiation, Criticality: aper.Reject, TransactionID: initiation.TransactionID, IEs: []lppa.IE{
		{ID: lppa.IEESMLCMeasurementID, Criticality: aper.Reject, Value: mid1},
		{ID: lppa.IEENBMeasurementID, Criticality: aper.Reject, Value: mid9},
	}}
	ackWire, err := lppa.Encode(ackPDU)
	if err != nil {
		t.Fatal(err)
	}
	w, err := lcsap.EncodeConnectionOriented(lcsap.ConnectionOriented{Correlation: carrier.Correlation, PayloadType: 1, Payload: ackWire}, c.SLs.MaxMessageSize)
	if err != nil {
		t.Fatal(err)
	}
	out, err = s.Handle("mme-a", w)
	if err != nil || len(out) != 0 {
		t.Fatalf("expected the job to stay active with no output: %d %v", len(out), err)
	}

	result := lppa.ECIDMeasurementResult{ServingCellID: lppa.ECGI{PLMNIdentity: [3]byte{0x00, 0xf1, 0x10}, CellIdentity: 1}, ServingCellTAC: [2]byte{0, 1},
		AccessPointPosition: &lppa.AccessPointPosition{Latitude: 38, Longitude: -90, Confidence: 68, UncertaintySemiMajor: 10}}
	resultBytes, err := lppa.EncodeECIDMeasurementResult(result)
	if err != nil {
		t.Fatal(err)
	}
	reportPDU := lppa.PDU{Category: lppa.Initiating, ProcedureCode: lppa.ProcedureECIDMeasurementReport, Criticality: aper.Ignore, TransactionID: 999, IEs: []lppa.IE{
		{ID: lppa.IEESMLCMeasurementID, Criticality: aper.Reject, Value: mid1},
		{ID: lppa.IEENBMeasurementID, Criticality: aper.Reject, Value: mid9},
		{ID: lppa.IECIDMeasurementResult, Criticality: aper.Ignore, Value: resultBytes},
	}}
	reportWire, err := lppa.Encode(reportPDU)
	if err != nil {
		t.Fatal(err)
	}
	w, err = lcsap.EncodeConnectionOriented(lcsap.ConnectionOriented{Correlation: carrier.Correlation, PayloadType: 1, Payload: reportWire}, c.SLs.MaxMessageSize)
	if err != nil {
		t.Fatal(err)
	}
	out, err = s.Handle("mme-a", w)
	if err != nil || len(out) != 2 {
		t.Fatalf("report %d %v", len(out), err)
	}
	finalP, err := lcsap.Decode(out[1])
	if err != nil || finalP.Category != lcsap.Successful {
		t.Fatalf("expected delivered estimate after report: %#v %v", finalP, err)
	}
}

// TestLPPaMessageWithoutActiveJobIsIgnored proves a stray LPPa message (no
// matching job — already finished, or never started) is logged and dropped
// rather than treated as a fatal association error.
func TestLPPaMessageWithoutActiveJobIsIgnored(t *testing.T) {
	c := config.Default()
	s := New(c, nil)
	mid1, _ := lppa.EncodeMeasurementID(1)
	mid9, _ := lppa.EncodeMeasurementID(9)
	cause, _ := lppa.EncodeCause(lppa.Cause{Branch: lppa.CauseMisc})
	pdu := lppa.PDU{Category: lppa.Initiating, ProcedureCode: lppa.ProcedureECIDMeasurementFailureIndication, Criticality: aper.Ignore, TransactionID: 1, IEs: []lppa.IE{
		{ID: lppa.IEESMLCMeasurementID, Criticality: aper.Reject, Value: mid1},
		{ID: lppa.IEENBMeasurementID, Criticality: aper.Reject, Value: mid9},
		{ID: lppa.IECause, Criticality: aper.Ignore, Value: cause},
	}}
	wire, err := lppa.Encode(pdu)
	if err != nil {
		t.Fatal(err)
	}
	w, err := lcsap.EncodeConnectionOriented(lcsap.ConnectionOriented{Correlation: [4]byte{9, 9, 9, 9}, PayloadType: 1, Payload: wire}, c.SLs.MaxMessageSize)
	if err != nil {
		t.Fatal(err)
	}
	out, err := s.Handle("mme-a", w)
	if err != nil || len(out) != 0 {
		t.Fatalf("expected silent ignore, got %d %v", len(out), err)
	}
}
