package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vectorcore/esmlc/internal/metrics"
	"github.com/vectorcore/esmlc/internal/positioning"
)

type fakeServer struct {
	metrics      *metrics.Registry
	live, ready  bool
	reloadResult positioning.CatalogReloadResult
	reloadCalled *bool
}

func (f fakeServer) Metrics() *metrics.Registry { return f.metrics }
func (f fakeServer) Live() bool                 { return f.live }
func (f fakeServer) Ready() bool                { return f.ready }
func (f fakeServer) ReloadCellCatalog() positioning.CatalogReloadResult {
	if f.reloadCalled != nil {
		*f.reloadCalled = true
	}
	return f.reloadResult
}

func TestHandlerMetricsEndpoint(t *testing.T) {
	r := metrics.NewRegistry()
	r.NewCounter("esmlc_test_total", "help").Inc()
	h := Handler(fakeServer{metrics: r, live: true, ready: true})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status: %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "esmlc_test_total 1") {
		t.Fatalf("body missing metric: %s", w.Body.String())
	}
}

func TestHandlerHealthAndReadyEndpoints(t *testing.T) {
	cases := []struct {
		name        string
		live, ready bool
		path        string
		wantStatus  int
	}{
		{"live ok", true, true, "/healthz", http.StatusOK},
		{"live not ok", false, true, "/healthz", http.StatusServiceUnavailable},
		{"ready ok", true, true, "/readyz", http.StatusOK},
		{"ready not ok", true, false, "/readyz", http.StatusServiceUnavailable},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := Handler(fakeServer{metrics: metrics.NewRegistry(), live: c.live, ready: c.ready})
			req := httptest.NewRequest(http.MethodGet, c.path, nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			if w.Code != c.wantStatus {
				t.Fatalf("got %d want %d", w.Code, c.wantStatus)
			}
		})
	}
}

func TestHandlerReloadCatalogEndpoint(t *testing.T) {
	called := false
	h := Handler(fakeServer{metrics: metrics.NewRegistry(), reloadCalled: &called, reloadResult: positioning.CatalogReloadResult{ActiveChanged: true, ActiveVersion: "v2", RecordCount: 3}})
	req := httptest.NewRequest(http.MethodPost, "/admin/reload-catalog", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if !called {
		t.Fatal("expected ReloadCellCatalog to be called")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status: %d body: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"ActiveVersion":"v2"`) {
		t.Fatalf("expected result JSON in body: %s", w.Body.String())
	}
}

func TestHandlerReloadCatalogReportsFailureAsConflict(t *testing.T) {
	h := Handler(fakeServer{metrics: metrics.NewRegistry(), reloadResult: positioning.CatalogReloadResult{Error: "catalog is not configured"}})
	req := httptest.NewRequest(http.MethodPost, "/admin/reload-catalog", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("status: %d body: %s", w.Code, w.Body.String())
	}
}

func TestHandlerReloadCatalogRejectsNonPost(t *testing.T) {
	h := Handler(fakeServer{metrics: metrics.NewRegistry()})
	req := httptest.NewRequest(http.MethodGet, "/admin/reload-catalog", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status: %d", w.Code)
	}
}
