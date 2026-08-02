package location

import (
	"bytes"
	"testing"

	"github.com/vectorcore/esmlc/internal/lpp/location/result"
	"github.com/vectorcore/esmlc/internal/uper"
)

func TestOTDOAProvideLocationInformationTwoNeighboursFixture(t *testing.T) {
	data := otdoaFixture(t, "provide-location-information-r9-otdoa-two-neighbours")
	r, err := uper.NewReader(data, len(data)*8)
	if err != nil {
		t.Fatal(err)
	}
	v, err := DecodeProvideLocationInformation(r)
	if err != nil {
		t.Fatal(err)
	}
	if v.OTDOA == nil {
		t.Fatal("expected OTDOA provide location information")
	}
	signal, ok := v.OTDOA.SignalMeasurementInformation()
	if !ok {
		t.Fatal("expected signal measurement information")
	}
	if signal.PhysCellIDRef() != 5 {
		t.Fatalf("physCellIdRef: got %d want 5", signal.PhysCellIDRef())
	}
	if _, ok := signal.CellGlobalIDRef(); !ok {
		t.Fatal("expected cellGlobalIdRef")
	}
	earfcn, ok := signal.EARFCNRef()
	if !ok || earfcn != 100 {
		t.Fatalf("earfcnRef: got %v ok=%v want 100", earfcn, ok)
	}
	neighbours := signal.NeighbourMeasurements()
	if len(neighbours) != 2 {
		t.Fatalf("expected 2 neighbours, got %d", len(neighbours))
	}
	if neighbours[0].PhysicalCellID() != 12 || neighbours[0].RSTD() != 6356 {
		t.Fatalf("neighbour 0: %#v", neighbours[0])
	}
	if _, ok := neighbours[0].EARFCN(); !ok {
		t.Fatal("neighbour 0: expected earfcnNeighbour present")
	}
	if neighbours[1].PhysicalCellID() != 300 || neighbours[1].RSTD() != 2260 {
		t.Fatalf("neighbour 1: %#v", neighbours[1])
	}
	if _, ok := neighbours[1].EARFCN(); ok {
		t.Fatal("neighbour 1: expected earfcnNeighbour absent")
	}
	if _, ok := neighbours[1].RSTDQuality().NumSamples(); !ok {
		t.Fatal("neighbour 1: expected error-NumSamples present")
	}

	w := uper.NewWriter()
	if err := EncodeProvideLocationInformation(w, v); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(w.Encoded().Bytes, data) {
		t.Fatalf("got %x want %x", w.Encoded().Bytes, data)
	}
}

func TestOTDOAProvideLocationInformationErrorFixture(t *testing.T) {
	data := otdoaFixture(t, "provide-location-information-r9-otdoa-error")
	r, err := uper.NewReader(data, len(data)*8)
	if err != nil {
		t.Fatal(err)
	}
	v, err := DecodeProvideLocationInformation(r)
	if err != nil {
		t.Fatal(err)
	}
	if v.OTDOA == nil {
		t.Fatal("expected OTDOA provide location information")
	}
	e, ok := v.OTDOA.Error()
	if !ok || e.Source != OTDOAErrorTargetDevice || e.TargetDeviceCause != OTDOATargetDeviceCauseAssistanceDataMissing {
		t.Fatalf("expected target-device assistance-data-missing error, got %#v ok=%v", e, ok)
	}
	w := uper.NewWriter()
	if err := EncodeProvideLocationInformation(w, v); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(w.Encoded().Bytes, data) {
		t.Fatalf("got %x want %x", w.Encoded().Bytes, data)
	}
}

func TestOTDOASignalMeasurementInformationRejectsEmptyNeighbourList(t *testing.T) {
	sfn, err := result.NewSystemFrameNumberFromUint16(0)
	if err != nil {
		t.Fatal(err)
	}
	pci, err := result.NewPhysicalCellID(5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewOTDOASignalMeasurementInformation(sfn, pci, nil, nil, nil, nil); err == nil {
		t.Fatal("expected error for empty neighbour measurement list")
	}
}
