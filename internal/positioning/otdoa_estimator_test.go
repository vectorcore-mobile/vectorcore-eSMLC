package positioning

import (
	"math"
	"testing"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/capability"
	"github.com/vectorcore/esmlc/internal/lpp/location"
	locationresult "github.com/vectorcore/esmlc/internal/lpp/location/result"
	"github.com/vectorcore/esmlc/internal/uper"
)

func mustECGI(t *testing.T, cellID uint32) locationresult.ECGI {
	t.Helper()
	mcc, err := locationresult.NewMCC(0, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	mnc, err := locationresult.NewMNC2(0, 1)
	if err != nil {
		t.Fatal(err)
	}
	b := []byte{byte(cellID >> 24), byte(cellID >> 16), byte(cellID >> 8), byte(cellID << 4)}
	bits, err := uper.NewBitString(b, 28)
	if err != nil {
		t.Fatal(err)
	}
	ecgi, err := locationresult.NewECGI(mcc, mnc, bits)
	if err != nil {
		t.Fatal(err)
	}
	return ecgi
}

func mustQuality(t *testing.T) locationresult.OTDOAMeasQuality {
	t.Helper()
	resolution, err := uper.NewBitString([]byte{0x40}, 2)
	if err != nil {
		t.Fatal(err)
	}
	value, err := uper.NewBitString([]byte{0x00}, 5)
	if err != nil {
		t.Fatal(err)
	}
	q, err := locationresult.NewOTDOAMeasQuality(resolution, value, nil)
	if err != nil {
		t.Fatal(err)
	}
	return q
}

// TestOTDOAEstimatorRecoversKnownPosition builds a synthetic scenario with
// known cell positions and a known true UE position, computes the exact
// range differences that geometry implies, quantizes them through the real
// TS 36.133 RSTD codec (round-tripping through the wire representation a UE
// would actually report), and checks the estimator recovers the true
// position within a small tolerance. This is the primary correctness
// evidence for the Gauss-Newton solver: independently re-deriving Chan's
// method by hand to cross-check it was judged more error-prone than this
// end-to-end geometric check.
func TestOTDOAEstimatorRecoversKnownPosition(t *testing.T) {
	refLat, refLon := 38.0, -90.0
	cells := []struct {
		lat, lon float64
	}{
		{38.0, -90.0},     // reference
		{38.02, -89.97},   // neighbour 1
		{37.985, -89.965}, // neighbour 2
		{38.01, -90.03},   // neighbour 3
	}
	trueLat, trueLon := 38.006, -89.995

	now := time.Unix(1000, 0)
	catalog := &CellCatalog{version: "test", cells: map[[7]byte]ServingCellRecord{}, maxAge: time.Hour, loadedAt: now}
	ecgis := make([]locationresult.ECGI, len(cells))
	for i, c := range cells {
		ecgi := mustECGI(t, uint32(i+1))
		ecgis[i] = ecgi
		catalog.cells[catalogECGIKey(ecgi)] = ServingCellRecord{ECGI: catalogECGIKey(ecgi), Latitude: c.lat, Longitude: c.lon, CoverageUncertainty: 10, Source: "test", UpdatedAt: now}
	}
	store := &CatalogStore{active: catalog}

	// Exact range differences implied by the true UE position, using the
	// same local-plane projection the estimator itself uses.
	trueX, trueY := projectToLocalMeters(refLat, refLon, trueLat, trueLon)
	d0 := math.Hypot(trueX, trueY)
	quality := mustQuality(t)
	var neighbours []location.NeighbourMeasurementElement
	for i := 1; i < len(cells); i++ {
		cx, cy := projectToLocalMeters(refLat, refLon, cells[i].lat, cells[i].lon)
		di := math.Hypot(trueX-cx, trueY-cy)
		rangeDiff := di - d0
		rstdDuration := time.Duration(rangeDiff / speedOfLight * float64(time.Second))
		rstd := locationresult.DurationToRSTD(rstdDuration)
		pci, err := locationresult.NewPhysicalCellID(uint16(i))
		if err != nil {
			t.Fatal(err)
		}
		ecgi := ecgis[i]
		nb, err := location.NewNeighbourMeasurementElement(pci, &ecgi, nil, rstd, quality)
		if err != nil {
			t.Fatal(err)
		}
		neighbours = append(neighbours, nb)
	}
	sfn, err := locationresult.NewSystemFrameNumberFromUint16(0)
	if err != nil {
		t.Fatal(err)
	}
	refPCI, err := locationresult.NewPhysicalCellID(0)
	if err != nil {
		t.Fatal(err)
	}
	refECGI := ecgis[0]
	signal, err := location.NewOTDOASignalMeasurementInformation(sfn, refPCI, &refECGI, nil, nil, neighbours)
	if err != nil {
		t.Fatal(err)
	}

	estimator := OTDOAEstimator{Store: store}
	result := estimator.Estimate(Request{}, MethodResult{Method: MethodOTDOA, OTDOA: &RawOTDOAMeasurements{Signal: signal}}, now)
	if result.Estimate == nil {
		t.Fatalf("expected estimate, got failure %v", result.Failure)
	}
	latErr := math.Abs(result.Estimate.Latitude - trueLat)
	lonErr := math.Abs(result.Estimate.Longitude - trueLon)
	// One degree of latitude is ~111km; require recovery within ~200m.
	if latErr > 200.0/111000.0 || lonErr > 200.0/111000.0 {
		t.Fatalf("recovered (%.6f,%.6f) too far from true (%.6f,%.6f): latErr=%.1fm lonErr=%.1fm",
			result.Estimate.Latitude, result.Estimate.Longitude, trueLat, trueLon, latErr*111000, lonErr*111000)
	}
	if result.Estimate.Source != EstimateSourceOTDOAMultilateration {
		t.Fatalf("wrong estimate source: %v", result.Estimate.Source)
	}
}

func TestOTDOAEstimatorRequiresECGIIdentifiedCells(t *testing.T) {
	now := time.Unix(1000, 0)
	catalog := &CellCatalog{version: "test", cells: map[[7]byte]ServingCellRecord{}, maxAge: time.Hour, loadedAt: now}
	store := &CatalogStore{active: catalog}
	quality := mustQuality(t)
	rstd, _ := locationresult.NewRSTD(6356)
	pci, _ := locationresult.NewPhysicalCellID(1)
	nb, err := location.NewNeighbourMeasurementElement(pci, nil, nil, rstd, quality)
	if err != nil {
		t.Fatal(err)
	}
	sfn, _ := locationresult.NewSystemFrameNumberFromUint16(0)
	refPCI, _ := locationresult.NewPhysicalCellID(0)
	// No cellGlobalIdRef: the reference cell cannot be resolved.
	signal, err := location.NewOTDOASignalMeasurementInformation(sfn, refPCI, nil, nil, nil, []location.NeighbourMeasurementElement{nb})
	if err != nil {
		t.Fatal(err)
	}
	estimator := OTDOAEstimator{Store: store}
	result := estimator.Estimate(Request{}, MethodResult{Method: MethodOTDOA, OTDOA: &RawOTDOAMeasurements{Signal: signal}}, now)
	if result.Estimate != nil || result.Failure != InsufficientNetworkData {
		t.Fatalf("expected InsufficientNetworkData failure, got %#v", result)
	}
}

func TestSolveHyperbolicTDOARejectsCollinearGeometry(t *testing.T) {
	ref := station{0, 0}
	neighbours := []station{{100, 0}, {200, 0}}
	_, _, _, ok := solveHyperbolicTDOA(ref, neighbours, []float64{10, 20})
	if ok {
		t.Fatal("expected collinear geometry to fail closed")
	}
}

// TestJobDeliversOTDOAEstimateThroughCombinedEstimator drives a full
// Manager.Start/Apply cycle (the same layer sls.Server uses) with a real
// CombinedEstimator{OTDOA: OTDOAEstimator{...}}, closing the gap between
// TestOTDOAEstimatorRecoversKnownPosition (the estimator in isolation) and
// the job-wiring tests in otdoa_job_test.go (which use no estimator at all).
func TestJobDeliversOTDOAEstimateThroughCombinedEstimator(t *testing.T) {
	refLat, refLon := 38.0, -90.0
	cells := []struct{ lat, lon float64 }{
		{38.0, -90.0},
		{38.02, -89.97},
		{37.985, -89.965},
	}
	now := time.Unix(0, 0)
	catalog := &CellCatalog{version: "test", cells: map[[7]byte]ServingCellRecord{}, maxAge: time.Hour, loadedAt: now}
	ecgis := make([]locationresult.ECGI, len(cells))
	for i, c := range cells {
		ecgi := mustECGI(t, uint32(i+10))
		ecgis[i] = ecgi
		catalog.cells[catalogECGIKey(ecgi)] = ServingCellRecord{ECGI: catalogECGIKey(ecgi), Latitude: c.lat, Longitude: c.lon, CoverageUncertainty: 10, Source: "test", UpdatedAt: now}
	}
	store := &CatalogStore{active: catalog}
	m := NewWithEstimator(otdoaPolicy(t, true), CombinedEstimator{OTDOA: OTDOAEstimator{Store: store}})

	o := proc(t)
	start, err := m.Start(req(), o, now)
	if err != nil || start.Snapshot.Method != MethodOTDOA {
		t.Fatalf("start %#v %v", start, err)
	}
	modeBits, _ := uper.NewBitString([]byte{0x80}, 1)
	capMsg := lpp.Message{TransactionID: &lpp.TransactionID{Initiator: start.Actions[0].Key.Initiator, TransactionNumber: start.Actions[0].Key.Number}, Body: &lpp.Body{Kind: lpp.BodyProvideCapabilities, ProvideCapabilities: &capability.ProvideCapabilitiesR9IEs{OTDOA: &capability.OTDOAProvideCapabilities{Mode: modeBits}}}}
	pr, err := o.HandleInbound(capMsg, now)
	if err != nil {
		t.Fatal(err)
	}
	locStart, err := m.Apply(req().Scope, pr.Events, now)
	if err != nil {
		t.Fatal(err)
	}

	trueLat, trueLon := 38.006, -89.995
	trueX, trueY := projectToLocalMeters(refLat, refLon, trueLat, trueLon)
	d0 := math.Hypot(trueX, trueY)
	quality := mustQuality(t)
	var neighbours []location.NeighbourMeasurementElement
	for i := 1; i < len(cells); i++ {
		cx, cy := projectToLocalMeters(refLat, refLon, cells[i].lat, cells[i].lon)
		di := math.Hypot(trueX-cx, trueY-cy)
		rstdDuration := time.Duration((di - d0) / speedOfLight * float64(time.Second))
		rstd := locationresult.DurationToRSTD(rstdDuration)
		pci, _ := locationresult.NewPhysicalCellID(uint16(i))
		ecgi := ecgis[i]
		nb, err := location.NewNeighbourMeasurementElement(pci, &ecgi, nil, rstd, quality)
		if err != nil {
			t.Fatal(err)
		}
		neighbours = append(neighbours, nb)
	}
	sfn, _ := locationresult.NewSystemFrameNumberFromUint16(0)
	refPCI, _ := locationresult.NewPhysicalCellID(0)
	refECGI := ecgis[0]
	signal, err := location.NewOTDOASignalMeasurementInformation(sfn, refPCI, &refECGI, nil, nil, neighbours)
	if err != nil {
		t.Fatal(err)
	}
	otdoaProvide, err := location.NewOTDOAProvideLocationInformation(&signal, nil)
	if err != nil {
		t.Fatal(err)
	}
	provide := lpp.Message{TransactionID: &lpp.TransactionID{Initiator: locStart.Actions[0].Key.Initiator, TransactionNumber: locStart.Actions[0].Key.Number}, Body: &lpp.Body{Kind: lpp.BodyProvideLocationInformation, ProvideLocationInformation: &location.ProvideLocationInformationR9IEs{OTDOA: &otdoaProvide}}}
	pr, err = o.HandleInbound(provide, now)
	if err != nil {
		t.Fatal(err)
	}
	final, err := m.Apply(req().Scope, pr.Events, now)
	if err != nil {
		t.Fatal(err)
	}
	if final.Snapshot.State != EstimateAvailable || final.Snapshot.Final == nil || final.Snapshot.Final.Kind != FinalEstimateAvailable || final.Snapshot.Final.Estimate == nil {
		t.Fatalf("expected a delivered OTDOA estimate, got %#v", final.Snapshot.Final)
	}
	if final.Snapshot.Final.Estimate.Source != EstimateSourceOTDOAMultilateration {
		t.Fatalf("wrong estimate source: %v", final.Snapshot.Final.Estimate.Source)
	}
	latErrM := math.Abs(final.Snapshot.Final.Estimate.Latitude-trueLat) * 111000
	if latErrM > 200 {
		t.Fatalf("recovered estimate too far from truth: %.1fm", latErrM)
	}
}
