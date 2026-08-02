// Package observability serves the metrics and health/readiness HTTP
// endpoints on a small, separate HTTP listener from the SCTP transport
// internal/sls owns. It depends on a minimal interface rather than
// internal/sls's concrete type, mirroring the layering discipline used
// throughout this codebase (nothing above a boundary knows the concrete
// type below it).
package observability

import (
	"encoding/json"
	"net/http"

	"github.com/vectorcore/esmlc/internal/metrics"
	"github.com/vectorcore/esmlc/internal/positioning"
)

// Server is the subset of *sls.Server this package needs.
type Server interface {
	Metrics() *metrics.Registry
	Live() bool
	Ready() bool
	// ReloadCellCatalog reloads only the already-configured operator catalog
	// file; there is no way to point it at a different path through this
	// endpoint. It is a no-op (CatalogReloadResult.Error set) if no catalog
	// is configured.
	ReloadCellCatalog() positioning.CatalogReloadResult
}

// Handler builds the /metrics, /healthz, /readyz, and /admin/reload-catalog
// endpoints. /healthz reports whether the process is running normally (not
// shutting down); /readyz additionally reports whether the SCTP transport
// can currently accept associations. There is deliberately no automatic
// filesystem watcher for the cell catalog (see docs/architecture.md): reload
// stays an explicit, operator-triggered action, now actually reachable
// through this endpoint rather than only from Go code. This service applies
// no authentication of its own to /admin/reload-catalog; operators exposing
// this listener beyond localhost must restrict access at the network layer.
func Handler(s Server) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = s.Metrics().WriteTo(w)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeStatus(w, s.Live())
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		writeStatus(w, s.Ready())
	})
	mux.HandleFunc("/admin/reload-catalog", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		result := s.ReloadCellCatalog()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if result.Error != "" {
			w.WriteHeader(http.StatusConflict)
		}
		_ = json.NewEncoder(w).Encode(result)
	})
	return mux
}

func writeStatus(w http.ResponseWriter, ok bool) {
	if ok {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte("unavailable\n"))
}
