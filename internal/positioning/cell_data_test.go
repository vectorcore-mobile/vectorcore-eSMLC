package positioning

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	path := filepath.Join(t.TempDir(), "unsupported.yaml")
	valid, err := os.ReadFile(filepath.Join("testdata", "serving-cells.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(string(valid), "schema_version: 1", "schema_version: 2", 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCellCatalog(path, 31*24*time.Hour, now); err == nil {
		t.Fatal("unsupported schema accepted")
	}
	if err := os.WriteFile(path, append(valid, []byte("\nunknown: true\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCellCatalog(path, 31*24*time.Hour, now); err == nil {
		t.Fatal("unknown field accepted")
	}
}

func TestCatalogStoreAtomicReloadAndSnapshot(t *testing.T) {
	now := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)
	path := filepath.Join(t.TempDir(), "cells.yaml")
	initial, err := os.ReadFile(filepath.Join("testdata", "serving-cells.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, initial, 0o600); err != nil {
		t.Fatal(err)
	}
	store := NewCatalogStore(path, 31*24*time.Hour, func() time.Time { return now })
	if result := store.Reload(); result.Error != "" || result.ActiveVersion != "operator-survey-2026-07" || result.RecordCount != 1 {
		t.Fatalf("initial reload %#v", result)
	}
	old := store.snapshot()
	replacement := strings.Replace(string(initial), "operator-survey-2026-07", "operator-survey-2026-08", 1)
	replacement = strings.Replace(replacement, "latitude: 38.000000", "latitude: 39.000000", 1)
	if err := os.WriteFile(path, []byte(replacement), 0o600); err != nil {
		t.Fatal(err)
	}
	if result := store.Reload(); result.Error != "" || result.ActiveVersion != "operator-survey-2026-08" {
		t.Fatalf("replacement reload %#v", result)
	}
	request := req()
	request.ServingECGI = [7]byte{0, 0xf1, 0x10, 0, 0, 0, 1}
	if cell, ok := old.Lookup(request.ServingECGI, now); !ok || cell.Latitude != 38 {
		t.Fatalf("old snapshot %#v %t", cell, ok)
	}
	result := (ServingCellCatalogEstimator{Store: store}).Estimate(request, MethodResult{Method: MethodECID, ECID: &RawECIDMeasurements{}}, now)
	if result.Estimate == nil || result.Estimate.Latitude != 39 || result.Estimate.CatalogVersion != "operator-survey-2026-08" {
		t.Fatalf("new snapshot estimate %#v", result)
	}
	if err := os.WriteFile(path, []byte("not: [valid"), 0o600); err != nil {
		t.Fatal(err)
	}
	if failed := store.Reload(); failed.Error == "" || failed.ActiveVersion != "operator-survey-2026-08" {
		t.Fatalf("invalid reload %#v", failed)
	}
	status := store.Status()
	if status.ActiveVersion != "operator-survey-2026-08" || status.ReloadSuccesses != 2 || status.ReloadFailures != 1 || status.AuthoritativeEstimates != 1 {
		t.Fatalf("status %#v", status)
	}
	now = now.Add(32 * 24 * time.Hour)
	if stale := (ServingCellCatalogEstimator{Store: store}).Estimate(request, MethodResult{Method: MethodECID, ECID: &RawECIDMeasurements{}}, now); stale.Estimate != nil || stale.Failure != InsufficientNetworkData {
		t.Fatalf("stale estimate %#v", stale)
	}
	if status = store.Status(); status.StaleData != 1 {
		t.Fatalf("stale status %#v", status)
	}
}

func TestCatalogStoreConcurrentReloadAndEstimate(t *testing.T) {
	now := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)
	store := NewCatalogStore(filepath.Join("testdata", "serving-cells.yaml"), 31*24*time.Hour, func() time.Time { return now })
	if result := store.Reload(); result.Error != "" {
		t.Fatal(result)
	}
	request := req()
	request.ServingECGI = [7]byte{0, 0xf1, 0x10, 0, 0, 0, 1}
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if result := store.Reload(); result.Error != "" {
				t.Errorf("reload %#v", result)
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := (ServingCellCatalogEstimator{Store: store}).Estimate(request, MethodResult{Method: MethodECID, ECID: &RawECIDMeasurements{}}, now)
			if result.Estimate == nil || result.Estimate.CatalogVersion != "operator-survey-2026-07" {
				t.Errorf("estimate %#v", result)
			}
		}()
	}
	wg.Wait()
	if status := store.Status(); status.ReloadSuccesses != 17 || status.AuthoritativeEstimates != 16 {
		t.Fatalf("status %#v", status)
	}
}
