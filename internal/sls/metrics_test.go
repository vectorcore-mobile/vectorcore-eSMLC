package sls

import (
	"strings"
	"testing"
	"time"

	"github.com/vectorcore/esmlc/internal/config"
)

// TestMetricsExposesAssociationAndOutcomeCounters proves the observability
// registry reflects live server state (association lifecycle) and records a
// terminal job outcome, rather than just existing unused.
func TestMetricsExposesAssociationAndOutcomeCounters(t *testing.T) {
	s := New(config.Default(), nil)
	var before strings.Builder
	if _, err := s.Metrics().WriteTo(&before); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(before.String(), "esmlc_sls_associations_active 0") {
		t.Fatalf("expected zero associations initially:\n%s", before.String())
	}

	// GNSS-only default policy with no eligible method configured: the
	// Location Request fails immediately with FinalNoEligibleMethod, which
	// must be counted.
	out, err := s.Handle("mme-a", locRequest())
	if err != nil || len(out) != 1 {
		t.Fatalf("start %d %v", len(out), err)
	}
	var after strings.Builder
	if _, err := s.Metrics().WriteTo(&after); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(after.String(), `esmlc_positioning_job_outcomes_total{outcome="no_eligible_method"} 1`) {
		t.Fatalf("expected no_eligible_method outcome counted:\n%s", after.String())
	}
}

func TestMetricsExposesCatalogCountersWhenConfigured(t *testing.T) {
	c := config.Default()
	c.Positioning.ECID.Enabled = true
	c.Positioning.ECID.RequestRSRP = true
	c.Positioning.ECID.CellDataFile = "../positioning/testdata/serving-cells.yaml"
	c.Positioning.ECID.CellDataMaxAge = 32 * 24 * time.Hour
	// The fixture's updated_at is fixed at 2026-07-01T00:00:00Z; CatalogStore
	// validates freshness at reload time too (not just lookup time), so this
	// must use a fixed clock inside that window or become the same
	// real-clock time bomb Phase 0 fixed (see docs/roadmap.md).
	fixedNow := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	s := newServer(c, nil, func() time.Time { return fixedNow })
	var buf strings.Builder
	if _, err := s.Metrics().WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "esmlc_catalog_reload_successes_total 1") {
		t.Fatalf("expected a successful reload within the fixture's freshness window:\n%s", out)
	}
	for _, want := range []string{"esmlc_catalog_records", "esmlc_catalog_reload_successes_total", "esmlc_catalog_reload_failures_total"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q present when a catalog is configured:\n%s", want, out)
		}
	}
}

func TestMetricsOmitsCatalogCountersWhenNotConfigured(t *testing.T) {
	s := New(config.Default(), nil)
	var buf strings.Builder
	if _, err := s.Metrics().WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "esmlc_catalog_records") {
		t.Fatalf("did not expect catalog metrics without a configured catalog:\n%s", buf.String())
	}
}
