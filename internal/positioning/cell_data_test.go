package positioning

import (
	"path/filepath"
	"testing"
	"time"
)

func TestServingCellCatalogAndEstimator(t *testing.T) {
	now := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)
	catalog, err := LoadCellCatalog(filepath.Join("testdata", "serving-cells.yaml"), 31*24*time.Hour, now)
	if err != nil {
		t.Fatal(err)
	}
	request := req()
	request.ServingECGI = [7]byte{0, 0xf1, 0x10, 0, 0, 0, 1}
	result := ServingCellEstimator{Catalog: catalog}.Estimate(request, MethodResult{Method: MethodECID, ECID: &RawECIDMeasurements{}}, now)
	if result.Estimate == nil || result.Estimate.Source != EstimateSourceAuthoritativeServingCell || result.Estimate.HorizontalUncertainty != 40 || result.Estimate.Latitude != 38 || result.Estimate.Longitude != -90 {
		t.Fatalf("estimate %#v", result)
	}
	request.ServingECGI[6] = 2
	if result := (ServingCellEstimator{Catalog: catalog}).Estimate(request, MethodResult{Method: MethodECID, ECID: &RawECIDMeasurements{}}, now); result.Estimate != nil || result.Failure != InsufficientNetworkData {
		t.Fatalf("missing cell %#v", result)
	}
}

func TestCellCatalogRejectsStaleAndDuplicateData(t *testing.T) {
	now := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)
	if _, err := LoadCellCatalog(filepath.Join("testdata", "serving-cells.yaml"), 24*time.Hour, now); err == nil {
		t.Fatal("stale catalog accepted")
	}
	if _, err := LoadCellCatalog(filepath.Join("testdata", "duplicate-serving-cells.yaml"), 31*24*time.Hour, now); err == nil {
		t.Fatal("duplicate catalog accepted")
	}
	if _, err := LoadCellCatalog(filepath.Join("testdata", "invalid-serving-cells.yaml"), 31*24*time.Hour, now); err == nil {
		t.Fatal("invalid coordinate accepted")
	}
}
