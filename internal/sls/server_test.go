package sls

import (
	"github.com/vectorcore/esmlc/internal/aper"
	"github.com/vectorcore/esmlc/internal/config"
	"github.com/vectorcore/esmlc/internal/lcsap"
	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/capability"
	"github.com/vectorcore/esmlc/internal/lpp/location"
	"github.com/vectorcore/esmlc/internal/lpp/location/result"
	"github.com/vectorcore/esmlc/internal/uper"
	"testing"
	"time"
)

func locRequest() []byte {
	w, _ := lcsap.Encode(lcsap.PDU{Category: lcsap.Initiating, Procedure: lcsap.ProcedureLocationRequest, Criticality: aper.Reject, IEs: []lcsap.IE{{ID: lcsap.IECorrelationID, Criticality: aper.Reject, Value: []byte{0, 0, 0, 7}}, {ID: lcsap.IELocationType, Criticality: aper.Reject, Value: []byte{0}}, {ID: lcsap.IEECGI, Criticality: aper.Ignore, Value: []byte{0, 0xf1, 0x10, 0, 0, 0, 1}}}})
	return w
}
func TestSimulationResponseAndReset(t *testing.T) {
	c := config.Default()
	c.Positioning.ECID.Enabled = true
	c.Positioning.ECID.RequestRSRP = true
	c.Positioning.Simulation.Enabled = true
	c.Positioning.Simulation.Latitude = 38
	c.Positioning.Simulation.Longitude = -90
	s := New(c, nil)
	out, e := s.Handle("mme-a", locRequest())
	if e != nil || len(out) != 1 {
		t.Fatalf("response %v %v", len(out), e)
	}
	p, e := lcsap.Decode(out[0])
	if e != nil || p.Procedure != lcsap.ProcedureConnectionOrientedInformation {
		t.Fatalf("bad initial procedure %v", e)
	}
	reset, _ := lcsap.Encode(lcsap.PDU{Category: lcsap.Initiating, Procedure: lcsap.ProcedureReset, Criticality: aper.Ignore, IEs: []lcsap.IE{{ID: lcsap.IELCSCause, Criticality: aper.Ignore, Value: []byte{0}}}})
	out, e = s.Handle("mme-a", reset)
	if e != nil || len(out) != 1 {
		t.Fatalf("reset %v", e)
	}
}

func TestSimulationEstimateDeliveredAfterECIDMeasurements(t *testing.T) {
	c := config.Default()
	c.Positioning.ECID.Enabled = true
	c.Positioning.ECID.RequestRSRP = true
	c.Positioning.Simulation.Enabled = true
	c.Positioning.Simulation.Latitude = 38
	c.Positioning.Simulation.Longitude = -90
	s := New(c, nil)
	out, err := s.Handle("mme-a", locRequest())
	if err != nil || len(out) != 1 {
		t.Fatalf("start %d %v", len(out), err)
	}
	p, _ := lcsap.Decode(out[0])
	carrier, _ := lcsap.DecodeConnectionOriented(p, c.SLs.MaxMessageSize)
	capRequest, _ := lpp.DecodeMessageOctets(carrier.Payload)
	bits, _ := uper.NewBitString([]byte{0x80}, 1)
	capProvide := lpp.Message{TransactionID: capRequest.TransactionID, Body: &lpp.Body{Kind: lpp.BodyProvideCapabilities, ProvideCapabilities: &capability.ProvideCapabilitiesR9IEs{ECID: &capability.ECIDProvideCapabilities{MeasurementSupport: bits}}}}
	encoded, _ := lpp.EncodeMessage(capProvide)
	w, _ := lcsap.EncodeConnectionOriented(lcsap.ConnectionOriented{Correlation: carrier.Correlation, PayloadType: 0, Payload: encoded.Bytes}, c.SLs.MaxMessageSize)
	out, err = s.Handle("mme-a", w)
	if err != nil || len(out) != 1 {
		t.Fatalf("capabilities %d %v", len(out), err)
	}
	p, _ = lcsap.Decode(out[0])
	carrier, _ = lcsap.DecodeConnectionOriented(p, c.SLs.MaxMessageSize)
	locRequest, _ := lpp.DecodeMessageOctets(carrier.Payload)
	pci, _ := result.NewPhysicalCellID(1)
	arfcn, _ := result.NewEUTRAARFCN(100)
	rsrp, _ := result.NewRSRPResult(30)
	measured, _ := result.NewMeasuredResultsElement(pci, arfcn, result.MeasuredResultsElementOptions{RSRPResult: &rsrp})
	signal, _ := location.NewECIDSignalMeasurementInformation(nil, []result.MeasuredResultsElement{measured})
	ecid, _ := location.NewECIDProvideLocationInformation(signal)
	provide := lpp.Message{TransactionID: locRequest.TransactionID, Body: &lpp.Body{Kind: lpp.BodyProvideLocationInformation, ProvideLocationInformation: &location.ProvideLocationInformationR9IEs{ECID: &ecid}}}
	encoded, _ = lpp.EncodeMessage(provide)
	w, _ = lcsap.EncodeConnectionOriented(lcsap.ConnectionOriented{Correlation: carrier.Correlation, PayloadType: 0, Payload: encoded.Bytes}, c.SLs.MaxMessageSize)
	out, err = s.Handle("mme-a", w)
	if err != nil || len(out) != 1 {
		t.Fatalf("measurements %d %v", len(out), err)
	}
	p, err = lcsap.Decode(out[0])
	if err != nil || p.Category != lcsap.Successful || p.Procedure != lcsap.ProcedureLocationRequest {
		t.Fatalf("estimate %#v %v", p, err)
	}
}

func TestAuthoritativeServingCellEstimateDeliveredAfterECIDMeasurements(t *testing.T) {
	c := config.Default()
	c.Positioning.ECID.Enabled = true
	c.Positioning.ECID.RequestRSRP = true
	c.Positioning.ECID.CellDataFile = "../positioning/testdata/serving-cells.yaml"
	c.Positioning.ECID.CellDataMaxAge = 32 * 24 * time.Hour
	// The fixture's updated_at is fixed at 2026-07-01T00:00:00Z; the catalog
	// clock must be fixed too, or this test becomes a time bomb that fails
	// once real time drifts past CellDataMaxAge from that fixed timestamp.
	fixedNow := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	s := newServer(c, nil, func() time.Time { return fixedNow })
	out, err := s.Handle("mme-a", locRequest())
	if err != nil || len(out) != 1 {
		t.Fatalf("start %d %v", len(out), err)
	}
	p, _ := lcsap.Decode(out[0])
	carrier, _ := lcsap.DecodeConnectionOriented(p, c.SLs.MaxMessageSize)
	capRequest, _ := lpp.DecodeMessageOctets(carrier.Payload)
	bits, _ := uper.NewBitString([]byte{0x80}, 1)
	capProvide := lpp.Message{TransactionID: capRequest.TransactionID, Body: &lpp.Body{Kind: lpp.BodyProvideCapabilities, ProvideCapabilities: &capability.ProvideCapabilitiesR9IEs{ECID: &capability.ECIDProvideCapabilities{MeasurementSupport: bits}}}}
	encoded, _ := lpp.EncodeMessage(capProvide)
	w, _ := lcsap.EncodeConnectionOriented(lcsap.ConnectionOriented{Correlation: carrier.Correlation, PayloadType: 0, Payload: encoded.Bytes}, c.SLs.MaxMessageSize)
	out, err = s.Handle("mme-a", w)
	if err != nil || len(out) != 1 {
		t.Fatalf("capabilities %d %v", len(out), err)
	}
	p, _ = lcsap.Decode(out[0])
	carrier, _ = lcsap.DecodeConnectionOriented(p, c.SLs.MaxMessageSize)
	locationRequest, _ := lpp.DecodeMessageOctets(carrier.Payload)
	pci, _ := result.NewPhysicalCellID(1)
	arfcn, _ := result.NewEUTRAARFCN(100)
	rsrp, _ := result.NewRSRPResult(30)
	measured, _ := result.NewMeasuredResultsElement(pci, arfcn, result.MeasuredResultsElementOptions{RSRPResult: &rsrp})
	signal, _ := location.NewECIDSignalMeasurementInformation(nil, []result.MeasuredResultsElement{measured})
	ecid, _ := location.NewECIDProvideLocationInformation(signal)
	provide := lpp.Message{TransactionID: locationRequest.TransactionID, Body: &lpp.Body{Kind: lpp.BodyProvideLocationInformation, ProvideLocationInformation: &location.ProvideLocationInformationR9IEs{ECID: &ecid}}}
	encoded, _ = lpp.EncodeMessage(provide)
	w, _ = lcsap.EncodeConnectionOriented(lcsap.ConnectionOriented{Correlation: carrier.Correlation, PayloadType: 0, Payload: encoded.Bytes}, c.SLs.MaxMessageSize)
	out, err = s.Handle("mme-a", w)
	if err != nil || len(out) != 1 {
		t.Fatalf("measurements %d %v", len(out), err)
	}
	p, err = lcsap.Decode(out[0])
	if err != nil || p.Category != lcsap.Successful || p.Procedure != lcsap.ProcedureLocationRequest {
		t.Fatalf("authoritative estimate %#v %v", p, err)
	}
}

func TestConfiguredCatalogReloadAndStatus(t *testing.T) {
	c := config.Default()
	c.Positioning.ECID.CellDataFile = "../positioning/testdata/serving-cells.yaml"
	c.Positioning.ECID.CellDataMaxAge = 32 * 24 * time.Hour
	// See TestAuthoritativeServingCellEstimateDeliveredAfterECIDMeasurements
	// for why this must be a fixed clock rather than time.Now.
	fixedNow := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	s := newServer(c, nil, func() time.Time { return fixedNow })
	status := s.CatalogStatus()
	if !status.Configured || status.ActiveVersion != "operator-survey-2026-07" || status.RecordCount != 1 || status.ReloadSuccesses != 1 {
		t.Fatalf("initial catalog status %#v", status)
	}
	if result := s.ReloadCellCatalog(); result.Error != "" || !result.ActiveChanged || result.ActiveVersion != status.ActiveVersion {
		t.Fatalf("reload %#v", result)
	}
	if status = s.CatalogStatus(); status.ReloadSuccesses != 2 || status.LastReloadError != "" {
		t.Fatalf("reloaded catalog status %#v", status)
	}
}
func TestGNSSDoesNotFabricateResult(t *testing.T) {
	s := New(config.Default(), nil)
	out, e := s.Handle("mme-a", locRequest())
	if e != nil || len(out) != 1 {
		t.Fatal(e)
	}
	p, e := lcsap.Decode(out[0])
	if e != nil || p.Category != lcsap.Unsuccessful {
		t.Fatal("expected failure")
	}
}

func TestLPPConnectionOrientedDispatchesProcedureAcknowledgement(t *testing.T) {
	requested, err := uper.NewBitString([]byte{0x80}, 1)
	if err != nil {
		t.Fatal(err)
	}
	seq := uint8(4)
	in := lpp.Message{
		TransactionID:   &lpp.TransactionID{Initiator: lpp.InitiatorTargetDevice, TransactionNumber: 9},
		SequenceNumber:  &seq,
		Acknowledgement: &lpp.Acknowledgement{Requested: true},
		Body:            &lpp.Body{Kind: lpp.BodyRequestLocationInformation, RequestLocationInformation: &location.RequestLocationInformationR9IEs{ECID: &location.ECIDRequestLocationInformation{RequestedMeasurements: requested}}},
	}
	encoded, err := lpp.EncodeMessage(in)
	if err != nil {
		t.Fatal(err)
	}
	correlation := [4]byte{0, 0, 0, 23}
	w, err := lcsap.EncodeConnectionOriented(lcsap.ConnectionOriented{Correlation: correlation, PayloadType: 0, Payload: encoded.Bytes}, config.Default().SLs.MaxMessageSize)
	if err != nil {
		t.Fatal(err)
	}
	s := New(config.Default(), nil)
	out, err := s.Handle("mme-a", w)
	if err != nil || len(out) != 1 {
		t.Fatalf("dispatch: %v %v", len(out), err)
	}
	p, err := lcsap.Decode(out[0])
	if err != nil {
		t.Fatal(err)
	}
	carrier, err := lcsap.DecodeConnectionOriented(p, config.Default().SLs.MaxMessageSize)
	if err != nil {
		t.Fatal(err)
	}
	if carrier.Correlation != correlation || carrier.PayloadType != 0 {
		t.Fatalf("carrier %#v", carrier)
	}
	ack, err := lpp.DecodeMessageOctets(carrier.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if ack.TransactionID == nil || ack.TransactionID.TransactionNumber != 9 || ack.SequenceNumber == nil || *ack.SequenceNumber != 4 || ack.Acknowledgement == nil || ack.Acknowledgement.Indicator == nil || *ack.Acknowledgement.Indicator != 0 {
		t.Fatalf("ack %#v", ack)
	}
}

func TestLPPConnectionOrientedRejectsMalformedAPDU(t *testing.T) {
	w, err := lcsap.EncodeConnectionOriented(lcsap.ConnectionOriented{Correlation: [4]byte{1}, PayloadType: 0, Payload: []byte{0xff}}, config.Default().SLs.MaxMessageSize)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = New(config.Default(), nil).Handle("mme-a", w); err == nil {
		t.Fatal("malformed LPP accepted")
	}
}

func TestLocationRequestStartsECIDCapabilityDiscoveryWhenPolicyEnabled(t *testing.T) {
	c := config.Default()
	c.Positioning.ECID.Enabled = true
	c.Positioning.ECID.RequestRSRP = true
	s := New(c, nil)
	out, err := s.Handle("mme-a", locRequest())
	if err != nil || len(out) != 1 {
		t.Fatalf("start %d %v", len(out), err)
	}
	p, err := lcsap.Decode(out[0])
	if err != nil {
		t.Fatal(err)
	}
	carrier, err := lcsap.DecodeConnectionOriented(p, c.SLs.MaxMessageSize)
	if err != nil {
		t.Fatal(err)
	}
	start, err := lpp.DecodeMessageOctets(carrier.Payload)
	if err != nil || start.Body == nil || start.Body.Kind != lpp.BodyRequestCapabilities || start.Body.RequestCapabilities == nil || start.Body.RequestCapabilities.ECID == nil {
		t.Fatalf("capability start %#v %v", start, err)
	}
	bits, _ := uper.NewBitString([]byte{0x80}, 1)
	provide := lpp.Message{TransactionID: start.TransactionID, Body: &lpp.Body{Kind: lpp.BodyProvideCapabilities, ProvideCapabilities: &capability.ProvideCapabilitiesR9IEs{ECID: &capability.ECIDProvideCapabilities{MeasurementSupport: bits}}}}
	encoded, err := lpp.EncodeMessage(provide)
	if err != nil {
		t.Fatal(err)
	}
	w, err := lcsap.EncodeConnectionOriented(lcsap.ConnectionOriented{Correlation: carrier.Correlation, PayloadType: 0, Payload: encoded.Bytes}, c.SLs.MaxMessageSize)
	if err != nil {
		t.Fatal(err)
	}
	out, err = s.Handle("mme-a", w)
	if err != nil || len(out) != 1 {
		t.Fatalf("capability result %d %v", len(out), err)
	}
	p, err = lcsap.Decode(out[0])
	if err != nil {
		t.Fatal(err)
	}
	carrier, err = lcsap.DecodeConnectionOriented(p, c.SLs.MaxMessageSize)
	if err != nil {
		t.Fatal(err)
	}
	request, err := lpp.DecodeMessageOctets(carrier.Payload)
	if err != nil || request.Body == nil || request.Body.Kind != lpp.BodyRequestLocationInformation || request.Body.RequestLocationInformation == nil || request.Body.RequestLocationInformation.ECID == nil {
		t.Fatalf("location start %#v %v", request, err)
	}
}

func TestLocationRequestWithoutEnabledMethodFails(t *testing.T) {
	out, err := New(config.Default(), nil).Handle("mme-a", locRequest())
	if err != nil || len(out) != 1 {
		t.Fatalf("%d %v", len(out), err)
	}
	p, err := lcsap.Decode(out[0])
	if err != nil || p.Category != lcsap.Unsuccessful {
		t.Fatalf("%#v %v", p, err)
	}
}

func TestLocationRequestWithInsufficientECIDCapabilitiesFails(t *testing.T) {
	c := config.Default()
	c.Positioning.ECID.Enabled = true
	c.Positioning.ECID.RequestRSRP = true
	s := New(c, nil)
	out, err := s.Handle("mme-a", locRequest())
	if err != nil || len(out) != 1 {
		t.Fatalf("start %d %v", len(out), err)
	}
	p, err := lcsap.Decode(out[0])
	if err != nil {
		t.Fatal(err)
	}
	carrier, err := lcsap.DecodeConnectionOriented(p, c.SLs.MaxMessageSize)
	if err != nil {
		t.Fatal(err)
	}
	start, err := lpp.DecodeMessageOctets(carrier.Payload)
	if err != nil {
		t.Fatal(err)
	}
	provide := lpp.Message{
		TransactionID: start.TransactionID,
		Body: &lpp.Body{
			Kind:                lpp.BodyProvideCapabilities,
			ProvideCapabilities: &capability.ProvideCapabilitiesR9IEs{},
		},
	}
	encoded, err := lpp.EncodeMessage(provide)
	if err != nil {
		t.Fatal(err)
	}
	w, err := lcsap.EncodeConnectionOriented(lcsap.ConnectionOriented{Correlation: carrier.Correlation, PayloadType: 0, Payload: encoded.Bytes}, c.SLs.MaxMessageSize)
	if err != nil {
		t.Fatal(err)
	}
	out, err = s.Handle("mme-a", w)
	if err != nil || len(out) != 1 {
		t.Fatalf("capability result %d %v", len(out), err)
	}
	p, err = lcsap.Decode(out[0])
	if err != nil || p.Category != lcsap.Unsuccessful || p.Procedure != lcsap.ProcedureLocationRequest {
		t.Fatalf("failure %#v %v", p, err)
	}
}

func TestResetReleasesActivePositioningJob(t *testing.T) {
	c := config.Default()
	c.Positioning.ECID.Enabled = true
	c.Positioning.ECID.RequestRSRP = true
	s := New(c, nil)
	if out, err := s.Handle("mme-a", locRequest()); err != nil || len(out) != 1 {
		t.Fatalf("initial request %d %v", len(out), err)
	}
	reset, err := lcsap.Encode(lcsap.PDU{Category: lcsap.Initiating, Procedure: lcsap.ProcedureReset, Criticality: aper.Ignore, IEs: []lcsap.IE{{ID: lcsap.IELCSCause, Criticality: aper.Ignore, Value: []byte{0}}}})
	if err != nil {
		t.Fatal(err)
	}
	if out, err := s.Handle("mme-a", reset); err != nil || len(out) != 1 {
		t.Fatalf("reset %d %v", len(out), err)
	}
	if out, err := s.Handle("mme-a", locRequest()); err != nil || len(out) != 1 {
		t.Fatalf("request after reset %d %v", len(out), err)
	}
}
