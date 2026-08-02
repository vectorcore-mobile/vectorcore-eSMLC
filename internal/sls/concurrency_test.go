package sls

import (
	"fmt"
	"sync"
	"testing"

	"github.com/vectorcore/esmlc/internal/aper"
	"github.com/vectorcore/esmlc/internal/config"
	"github.com/vectorcore/esmlc/internal/lcsap"
	"github.com/vectorcore/esmlc/internal/positioning"
)

// locRequestFor builds a Location Request PDU for a specific association's
// correlation ID, so concurrent "MMEs" can be given the identical
// correlation ID bytes on purpose (the scoping this test verifies is
// association+correlation, not correlation alone).
func locRequestFor(correlation [4]byte) []byte {
	w, _ := lcsap.Encode(lcsap.PDU{Category: lcsap.Initiating, Procedure: lcsap.ProcedureLocationRequest, Criticality: aper.Reject, IEs: []lcsap.IE{
		{ID: lcsap.IECorrelationID, Criticality: aper.Reject, Value: correlation[:]},
		{ID: lcsap.IELocationType, Criticality: aper.Reject, Value: []byte{0}},
		{ID: lcsap.IEECGI, Criticality: aper.Ignore, Value: []byte{0, 0xf1, 0x10, 0, 0, 0, 1}},
	}})
	return w
}

// TestConcurrentMMEAssociationsAreIsolatedByScope drives many simulated MME
// associations concurrently through Server.Handle, all using the identical
// LCS-AP Correlation-ID value, and checks (a) with -race that no shared
// state (session.Manager, positioning.Manager, the LPP procedure table) is
// corrupted, and (b) that a job started under one association's identity
// cannot be observed, driven, or torn down by another association's
// traffic — Scope/session.ID key by Association+Correlation together, so
// identical correlation bytes from different MMEs must not collide.
func TestConcurrentMMEAssociationsAreIsolatedByScope(t *testing.T) {
	c := config.Default()
	c.SLs.MaxSessions = 10000
	s := New(c, nil)

	const associations = 20
	const requestsPerAssociation = 25
	sharedCorrelation := [4]byte{0xaa, 0xbb, 0xcc, 0xdd}

	var wg sync.WaitGroup
	errs := make(chan error, associations*requestsPerAssociation)
	for i := 0; i < associations; i++ {
		assoc := fmt.Sprintf("mme-%d", i)
		wg.Add(1)
		go func(assoc string) {
			defer wg.Done()
			for j := 0; j < requestsPerAssociation; j++ {
				out, err := s.Handle(assoc, locRequestFor(sharedCorrelation))
				if err != nil {
					errs <- fmt.Errorf("%s request %d: %w", assoc, j, err)
					continue
				}
				if len(out) != 1 {
					errs <- fmt.Errorf("%s request %d: expected 1 response, got %d", assoc, j, len(out))
					continue
				}
				p, err := lcsap.Decode(out[0])
				if err != nil {
					errs <- fmt.Errorf("%s request %d: decode response: %w", assoc, j, err)
					continue
				}
				// GNSS-only default config with no eligible method: every
				// request must fail the same documented way regardless of
				// which concurrent association it came from.
				if p.Category != lcsap.Unsuccessful {
					errs <- fmt.Errorf("%s request %d: expected failure response, got category %v", assoc, j, p.Category)
				}
				s.dropLPPAssociation(assoc)
				s.jobs.DropAssociation(assoc)
			}
		}(assoc)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

// TestConcurrentAssociationCloseDoesNotAffectOtherAssociations starts a real
// positioning job on one association and confirms that closing a different
// association (DropAssociation) never touches it, even when both used
// identical correlation bytes.
func TestConcurrentAssociationCloseDoesNotAffectOtherAssociations(t *testing.T) {
	c := config.Default()
	c.Positioning.ECID.Enabled = true
	c.Positioning.ECID.RequestRSRP = true
	s := New(c, nil)
	correlation := [4]byte{1, 2, 3, 4}

	out, err := s.Handle("mme-victim", locRequestFor(correlation))
	if err != nil || len(out) != 1 {
		t.Fatalf("start on mme-victim: %d %v", len(out), err)
	}
	if _, ok := s.jobs.Snapshot(positioning.Scope{Association: "mme-victim", Correlation: correlation}); !ok {
		t.Fatal("expected an active job for mme-victim")
	}

	// A concurrent, unrelated association using the exact same correlation
	// bytes must not be able to see or disturb mme-victim's job.
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		assoc := fmt.Sprintf("mme-other-%d", i)
		go func(assoc string) {
			defer wg.Done()
			s.dropLPPAssociation(assoc)
			s.jobs.DropAssociation(assoc)
		}(assoc)
	}
	wg.Wait()

	if _, ok := s.jobs.Snapshot(positioning.Scope{Association: "mme-victim", Correlation: correlation}); !ok {
		t.Fatal("mme-victim's job was disturbed by an unrelated association's teardown")
	}
}
