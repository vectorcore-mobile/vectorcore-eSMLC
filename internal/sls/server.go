// Package sls owns SCTP association lifecycle and LCS-AP dispatch.
package sls

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ishidawataru/sctp"
	"github.com/vectorcore/esmlc/internal/aper"
	"github.com/vectorcore/esmlc/internal/config"
	"github.com/vectorcore/esmlc/internal/lcsap"
	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/location"
	"github.com/vectorcore/esmlc/internal/lpp/procedure"
	"github.com/vectorcore/esmlc/internal/lpp/transaction"
	"github.com/vectorcore/esmlc/internal/lppa"
	"github.com/vectorcore/esmlc/internal/metrics"
	"github.com/vectorcore/esmlc/internal/positioning"
	"github.com/vectorcore/esmlc/internal/session"
	"github.com/vectorcore/esmlc/internal/uper"
	"log/slog"
	"net"
	"sync"
	"time"
)

// debugHexPreviewLimit bounds how many wire-format bytes -d debug logging
// dumps per PDU. LCS-AP messages can carry GNSS assistance data payloads up
// to sls.max_message_size (default 1MiB); logging the raw bytes unbounded
// would make the console unusable during a live diagnostic session.
const debugHexPreviewLimit = 2048

// hexPreview renders b as hex, truncated to debugHexPreviewLimit with a
// trailing byte count for anything cut off.
func hexPreview(b []byte) string {
	if len(b) <= debugHexPreviewLimit {
		return hex.EncodeToString(b)
	}
	return hex.EncodeToString(b[:debugHexPreviewLimit]) + fmt.Sprintf("...(%d more bytes)", len(b)-debugHexPreviewLimit)
}

type Server struct {
	cfg      config.Config
	log      *slog.Logger
	sessions *session.Manager
	mu       sync.Mutex
	listener *sctp.SCTPListener
	closing  bool
	assocs   map[string]*association
	lpp      map[session.ID]*procedure.Orchestrator
	jobs     *positioning.Manager
	catalog  *positioning.CatalogStore
	now      func() time.Time
	wg       sync.WaitGroup
	metrics  *metrics.Registry
	outcomes *metrics.CounterVec
}
type association struct {
	id     string
	conn   *sctp.SCTPConn
	write  sync.Mutex
	server *Server
	once   sync.Once
}

func New(cfg config.Config, log *slog.Logger) *Server {
	return newServer(cfg, log, time.Now)
}

// newServer takes an explicit clock so callers that must reason about cell
// catalog freshness deterministically (tests) do not depend on the real wall
// clock alongside a fixed fixture timestamp.
func newServer(cfg config.Config, log *slog.Logger, now func() time.Time) *Server {
	if log == nil {
		log = slog.Default()
	}
	policy := positioningPolicy(cfg)
	var ecidEstimator, otdoaEstimator, agnssEstimator, lppaEstimator positioning.Estimator
	var catalog *positioning.CatalogStore
	// Unlike ECID/OTDOA, A-GNSS (UE-based mode) needs no cell catalog: the UE
	// already computed its own position, so this estimator only validates
	// and relays it.
	if cfg.Positioning.AGNSS.Enabled {
		agnssEstimator = positioning.AGNSSEstimator{}
	}
	if cfg.Positioning.ECID.CellDataFile != "" {
		catalog = positioning.NewCatalogStore(cfg.Positioning.ECID.CellDataFile, cfg.Positioning.ECID.CellDataMaxAge, now)
		reload := catalog.Reload()
		if reload.Error != "" {
			log.Error("esmlc.cell_catalog_unavailable", "error", reload.Error)
		} else {
			log.Info("esmlc.cell_catalog_loaded", "version", reload.ActiveVersion, "records", reload.RecordCount)
		}
		ecidEstimator = positioning.ServingCellCatalogEstimator{Store: catalog}
		// The same operator-maintained catalog gives OTDOA the
		// reference/neighbour cell positions its multilateration solver
		// needs; there is no separate OTDOA catalog file.
		if cfg.Positioning.OTDOA.Enabled {
			otdoaEstimator = positioning.OTDOAEstimator{Store: catalog}
		}
	} else if cfg.Positioning.Simulation.Enabled {
		simFailure := positioning.EstimationFailure(0)
		if cfg.Positioning.Simulation.FailureCause != nil {
			simFailure = positioning.InsufficientNetworkData
		}
		ecidEstimator = positioning.SimulationEstimator{Latitude: cfg.Positioning.Simulation.Latitude, Longitude: cfg.Positioning.Simulation.Longitude, Uncertainty: positioning.UncertaintyCodeFromMeters(cfg.Positioning.Simulation.UncertaintyMeters), Failure: simFailure}
	}
	if cfg.Positioning.LPPaECID.Enabled {
		// Reuses the same operator catalog as a fallback for when the eNB
		// does not report its own antenna position; catalog may be nil, in
		// which case LPPaECIDEstimator only ever succeeds from a reported
		// E-UTRANAccessPointPosition.
		lppaEstimator = positioning.LPPaECIDEstimator{Store: catalog}
	}
	// A nil positioning.Estimator (neither branch above configured) must
	// stay nil, not become a non-nil CombinedEstimator with both fields
	// nil: the two produce observably different job outcomes (Manager's own
	// nil check vs. an estimator-returned failure).
	var estimator positioning.Estimator
	if ecidEstimator != nil || otdoaEstimator != nil || agnssEstimator != nil || lppaEstimator != nil {
		estimator = positioning.CombinedEstimator{ECID: ecidEstimator, OTDOA: otdoaEstimator, AGNSS: agnssEstimator, LPPaECID: lppaEstimator}
	}
	s := &Server{cfg: cfg, log: log, sessions: session.New(cfg.SLs.MaxSessions), assocs: map[string]*association{}, lpp: map[session.ID]*procedure.Orchestrator{}, jobs: positioning.NewWithEstimator(policy, estimator), catalog: catalog, now: now, metrics: metrics.NewRegistry()}
	s.registerMetrics()
	return s
}

// registerMetrics wires the observability registry to this server's live
// state. Every gauge reads its source of truth at scrape time (GaugeFunc/
// CounterFunc) rather than being pushed, so there is nothing here that can
// drift out of sync with internal/session, internal/positioning, or the
// catalog store.
func (s *Server) registerMetrics() {
	s.metrics.NewGaugeFunc("esmlc_sls_associations_active", "Currently open SCTP associations.", func() int64 {
		s.mu.Lock()
		defer s.mu.Unlock()
		return int64(len(s.assocs))
	})
	// session.Manager.Create is not yet called anywhere in the request path
	// (see docs/architecture.md); this always reads zero today. Exposed
	// anyway so the gauge appears the moment that changes, rather than
	// needing a second metrics change alongside whatever wires it up.
	s.metrics.NewGaugeFunc("esmlc_sls_sessions_active", "Currently tracked LCS-AP sessions (always 0: session.Manager.Create is not yet called in the request path).", func() int64 {
		return int64(s.sessions.Count())
	})
	s.metrics.NewGaugeFunc("esmlc_positioning_jobs_active", "Currently active positioning jobs.", func() int64 {
		return int64(s.jobs.ActiveJobs())
	})
	s.outcomes = s.metrics.NewCounterVec("esmlc_positioning_job_outcomes_total", "Completed positioning jobs by terminal outcome kind.", "outcome",
		[]string{"estimate_available", "measurements_without_estimator", "estimation_failed", "quality_not_met", "no_eligible_method", "lpp_unsupported", "procedure_failure", "deadline_expired", "cancelled", "unknown"})
	if s.catalog != nil {
		s.metrics.NewGaugeFunc("esmlc_catalog_records", "Records in the active cell catalog snapshot (0 if none loaded).", func() int64 {
			return int64(s.catalog.Status().RecordCount)
		})
		s.metrics.NewCounterFunc("esmlc_catalog_reload_successes_total", "Successful cell catalog reloads.", func() uint64 {
			return s.catalog.Status().ReloadSuccesses
		})
		s.metrics.NewCounterFunc("esmlc_catalog_reload_failures_total", "Failed cell catalog reloads.", func() uint64 {
			return s.catalog.Status().ReloadFailures
		})
		s.metrics.NewCounterFunc("esmlc_catalog_authoritative_estimates_total", "Estimates served from the authoritative cell catalog.", func() uint64 {
			return s.catalog.Status().AuthoritativeEstimates
		})
		s.metrics.NewCounterFunc("esmlc_catalog_missing_cell_total", "Catalog lookups for a serving cell absent from the catalog.", func() uint64 {
			return s.catalog.Status().MissingCell
		})
		s.metrics.NewCounterFunc("esmlc_catalog_stale_data_total", "Catalog lookups rejected for exceeding the configured maximum age.", func() uint64 {
			return s.catalog.Status().StaleData
		})
	}
}

// Metrics exposes the Prometheus-format observability registry for this
// server. Callers typically mount it at an HTTP /metrics endpoint.
func (s *Server) Metrics() *metrics.Registry { return s.metrics }

// debugEnabled reports whether DEBUG-level logging is currently active, so
// callers can skip building expensive attrs (hex dumps of raw PDUs) on the
// hot path when they would just be discarded by the handler.
func (s *Server) debugEnabled() bool { return s.log.Enabled(context.Background(), slog.LevelDebug) }

// finalOutcomeLabel maps a FinalKind to its stable metric/log label. Every
// currently-defined FinalKind is mapped explicitly; "unknown" is a safety
// net for a future value this function has not been updated for, not a
// silently-accepted default.
func finalOutcomeLabel(kind positioning.FinalKind) string {
	switch kind {
	case positioning.FinalEstimateAvailable:
		return "estimate_available"
	case positioning.FinalMeasurementsWithoutEstimator:
		return "measurements_without_estimator"
	case positioning.FinalEstimationFailed:
		return "estimation_failed"
	case positioning.FinalQualityNotMet:
		return "quality_not_met"
	case positioning.FinalNoEligibleMethod:
		return "no_eligible_method"
	case positioning.FinalLPPUnsupported:
		return "lpp_unsupported"
	case positioning.FinalProcedureFailure:
		return "procedure_failure"
	case positioning.FinalDeadlineExpired:
		return "deadline_expired"
	case positioning.FinalCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// recordFinalOutcome logs and counts a job's terminal outcome. Called from
// every path that can observe a positioning.FinalOutcome (LPP-driven,
// LPPa-driven, and immediate Start-time failures alike).
func (s *Server) recordFinalOutcome(association string, final *positioning.FinalOutcome) {
	if final == nil {
		return
	}
	label := finalOutcomeLabel(final.Kind)
	s.outcomes.WithLabelValue(label).Inc()
	s.log.Info("esmlc.positioning.job_outcome", "association", association, "outcome", label)
}

// Live reports whether the process is running normally, i.e. not in the
// middle of shutting down. It never depends on SCTP or catalog state.
func (s *Server) Live() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return !s.closing
}

// Ready reports whether the server can currently accept SCTP associations:
// its listener must be bound if SLs is enabled. If SLs is disabled there is
// no listener to wait for, so readiness is trivially true (matching Listen's
// own no-op behavior in that case).
func (s *Server) Ready() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closing {
		return false
	}
	return !s.cfg.SLs.Enabled || s.listener != nil
}

// ReloadCellCatalog reloads only the configured operator catalog. It is an
// in-process administrative boundary; callers must apply their own control
// plane authorization rather than exposing arbitrary paths remotely.
func (s *Server) ReloadCellCatalog() positioning.CatalogReloadResult {
	if s.catalog == nil {
		return positioning.CatalogReloadResult{Error: "catalog is not configured"}
	}
	result := s.catalog.Reload()
	if result.Error != "" {
		s.log.Warn("esmlc.cell_catalog_reload_failed", "error", result.Error, "active_version", result.ActiveVersion)
	} else {
		s.log.Info("esmlc.cell_catalog_reloaded", "version", result.ActiveVersion, "records", result.RecordCount)
	}
	return result
}

func (s *Server) CatalogStatus() positioning.CatalogStatus {
	if s.catalog == nil {
		return positioning.CatalogStatus{}
	}
	return s.catalog.Status()
}
func (s *Server) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}
func (s *Server) Listen(ctx context.Context) error {
	if !s.cfg.SLs.Enabled {
		return nil
	}
	addr := &sctp.SCTPAddr{IPAddrs: []net.IPAddr{{IP: net.ParseIP(s.cfg.SLs.ListenAddress)}}, Port: s.cfg.SLs.Port}
	ln, e := (&sctp.SocketConfig{}).Listen("sctp", addr)
	if e != nil {
		return fmt.Errorf("sls listen: %w", e)
	}
	s.mu.Lock()
	s.listener = ln
	s.mu.Unlock()
	s.log.Info("esmlc.sls.listening", "address", ln.Addr().String(), "ppid", lcsap.PPID)
	s.wg.Add(1)
	go func() { defer s.wg.Done(); s.pruneLoop(ctx) }()

	// AcceptSCTP (github.com/ishidawataru/sctp) issues a raw blocking
	// Accept4 syscall that is never registered with Go's runtime netpoller
	// and exposes no deadline on the listener, so closing the listener fd
	// during shutdown is not guaranteed to unblock a goroutine parked in
	// it — in practice this showed up as the process hanging forever right
	// after logging "esmlc.shutdown", needing a kill -9. Running the
	// accept loop on its own goroutine lets Listen return as soon as ctx
	// is done and Close() has finished draining existing associations,
	// rather than staying blocked on that goroutine (which may never
	// return) for the life of the process.
	acceptDone := make(chan error, 1)
	go func() { acceptDone <- s.acceptLoop(ln) }()

	select {
	case <-ctx.Done():
		_ = s.Close()
		return nil
	case e := <-acceptDone:
		return e
	}
}

func (s *Server) acceptLoop(ln *sctp.SCTPListener) error {
	for {
		c, e := ln.AcceptSCTP()
		if e != nil {
			if s.closed() {
				return nil
			}
			return fmt.Errorf("sls accept: %w", e)
		}
		s.mu.Lock()
		full := s.closing || len(s.assocs) >= s.cfg.SLs.MaxAssociations
		s.mu.Unlock()
		if full {
			_ = c.Close()
			continue
		}
		a := &association{id: c.RemoteAddr().String(), conn: c, server: s}
		s.mu.Lock()
		s.assocs[a.id] = a
		s.mu.Unlock()
		s.log.Info("esmlc.sls.association_up", "association", a.id)
		s.wg.Add(1)
		go func() { defer s.wg.Done(); a.read() }()
	}
}
func (s *Server) closed() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.closing }
func (a *association) read() {
	if e := a.conn.SubscribeEvents(sctp.SCTP_EVENT_DATA_IO); e != nil {
		a.close(e)
		return
	}
	buf := make([]byte, a.server.cfg.SLs.MaxMessageSize)
	for {
		n, info, e := a.conn.SCTPRead(buf)
		if e != nil {
			a.close(e)
			return
		}
		ppid := uint32(0)
		if info != nil {
			ppid = info.PPID
		}
		a.server.log.Debug("esmlc.sls.pdu_read", "association", a.id, "bytes", n, "ppid", ppid)
		if n < 1 || n > len(buf) || info == nil || info.PPID != a.server.cfg.SLs.ExpectedPPID {
			a.server.log.Warn("esmlc.malformed_message", "association", a.id, "reason", "ppid_or_size", "bytes", n, "ppid", ppid, "expected_ppid", a.server.cfg.SLs.ExpectedPPID)
			continue
		}
		wire := append([]byte(nil), buf[:n]...)
		if a.server.debugEnabled() {
			a.server.log.Debug("esmlc.sls.pdu_hex_in", "association", a.id, "hex", hexPreview(wire))
		}
		out, e := a.server.Handle(a.id, wire)
		if e != nil {
			attrs := []any{"association", a.id, "error", e}
			if a.server.debugEnabled() {
				attrs = append(attrs, "hex", hexPreview(wire))
			}
			a.server.log.Warn("esmlc.lcsap.error", attrs...)
			continue
		}
		if len(out) == 0 {
			a.server.log.Debug("esmlc.sls.no_outbound_pdu", "association", a.id, "reason", "handler produced no response for this inbound PDU")
		}
		for _, w := range out {
			if a.server.debugEnabled() {
				a.server.log.Debug("esmlc.sls.pdu_write", "association", a.id, "bytes", len(w), "hex", hexPreview(w))
			}
			if e = a.send(w); e != nil {
				a.close(e)
				return
			}
		}
	}
}
func (a *association) send(b []byte) error {
	a.write.Lock()
	defer a.write.Unlock()
	_, e := a.conn.SCTPWrite(b, &sctp.SndRcvInfo{PPID: lcsap.PPID, Stream: 0})
	return e
}
func (a *association) close(e error) {
	a.once.Do(func() {
		_ = a.conn.Close()
		s := a.server
		s.sessions.DropAssociation(a.id)
		s.dropLPPAssociation(a.id)
		s.jobs.DropAssociation(a.id)
		s.mu.Lock()
		delete(s.assocs, a.id)
		s.mu.Unlock()
		s.log.Info("esmlc.sls.association_down", "association", a.id, "error", e)
	})
}
func (s *Server) Close() error {
	s.mu.Lock()
	if s.closing {
		s.mu.Unlock()
		return nil
	}
	s.closing = true
	ln := s.listener
	as := make([]*association, 0, len(s.assocs))
	for _, a := range s.assocs {
		as = append(as, a)
	}
	s.mu.Unlock()
	if ln != nil {
		_ = ln.Close()
	}
	for _, a := range as {
		a.close(errors.New("shutdown"))
	}
	s.wg.Wait()
	s.log.Info("esmlc.shutdown")
	return nil
}

// Handle is transport-independent and is used by tests and the SCTP read loop.
func (s *Server) Handle(association string, wire []byte) ([][]byte, error) {
	p, e := lcsap.Decode(wire)
	if e != nil {
		return nil, fmt.Errorf("lcsap: decode: %w", e)
	}
	s.log.Debug("esmlc.lcsap.dispatch", "association", association, "procedure", p.Procedure, "category", p.Category, "criticality", p.Criticality, "ie_count", len(p.IEs))
	switch p.Procedure {
	case lcsap.ProcedureLocationRequest:
		return s.locationRequest(association, p)
	case lcsap.ProcedureReset:
		if p.Category != lcsap.Initiating || (p.Criticality != aper.Reject && p.Criticality != aper.Ignore) {
			return nil, fmt.Errorf("lcsap: invalid reset")
		}
		haveCause := false
		for _, ie := range p.IEs {
			if ie.ID == lcsap.IELCSCause && ie.Criticality == aper.Ignore && len(ie.Value) > 0 {
				haveCause = true
			}
		}
		if !haveCause {
			return nil, fmt.Errorf("lcsap: reset missing LCS Cause")
		}
		s.sessions.DropAssociation(association)
		s.dropLPPAssociation(association)
		s.jobs.DropAssociation(association)
		ack, e := lcsap.Encode(lcsap.PDU{Category: lcsap.Successful, Procedure: lcsap.ProcedureReset, Criticality: aper.Reject})
		if e != nil {
			return nil, e
		}
		return [][]byte{ack}, nil
	case lcsap.ProcedureConnectionOrientedInformation:
		v, err := lcsap.DecodeConnectionOriented(p, s.cfg.SLs.MaxMessageSize)
		if err != nil {
			return nil, err
		}
		s.log.Info("esmlc.lcsap.connection_oriented_received", "association", association, "correlation", fmt.Sprintf("%x", v.Correlation), "payload_type", v.PayloadType, "payload_length", len(v.Payload))
		switch v.PayloadType {
		case 0:
			return s.handleLPP(association, v)
		case 1:
			return s.handleLPPa(association, v)
		default:
			return nil, fmt.Errorf("lcsap: unsupported payload type %d", v.PayloadType)
		}
	case lcsap.ProcedureLocationAbort:
		return s.locationAbort(association, p)
	default:
		return nil, fmt.Errorf("lcsap: unsupported procedure %d", p.Procedure)
	}
}

// handleLPP is the SLs carrier boundary for payload type 0. LCS-AP keeps the
// APDU opaque; this server is the owning boundary that decodes it, applies the
// transport-neutral procedure, and wraps any resulting LPP actions back into
// Connection-Oriented Information using the distinct LCS correlation ID.
func (s *Server) handleLPP(association string, carrier lcsap.ConnectionOriented) ([][]byte, error) {
	corr := fmt.Sprintf("%x", carrier.Correlation)
	m, err := lpp.DecodeMessageOctets(carrier.Payload)
	if err != nil {
		s.log.Warn("esmlc.lpp.decode_failed", "association", association, "correlation", corr, "error", err)
		return nil, fmt.Errorf("sls: decode LPP APDU: %w", err)
	}
	bodyKind := "none"
	if m.Body != nil {
		bodyKind = fmt.Sprintf("%v", m.Body.Kind)
	}
	s.log.Debug("esmlc.lpp.message_received", "association", association, "correlation", corr, "body_kind", bodyKind)
	o, err := s.lppProcedure(session.ID{Association: association, Correlation: carrier.Correlation})
	if err != nil {
		return nil, err
	}
	r, err := o.HandleInbound(m, s.now())
	if err != nil {
		s.log.Warn("esmlc.lpp.procedure_apply_failed", "association", association, "correlation", corr, "body_kind", bodyKind, "error", err)
		return nil, fmt.Errorf("sls: apply LPP procedure: %w", err)
	}
	actions := append([]procedure.Action(nil), r.Actions...)
	jobResult, jobErr := s.jobs.Apply(positioning.Scope{Association: association, Correlation: carrier.Correlation}, r.Events, s.now())
	var final *positioning.FinalOutcome
	if jobErr == nil {
		actions = append(actions, jobResult.Actions...)
		final = jobResult.Snapshot.Final
	} else if jobErr != positioning.ErrNotActive {
		s.log.Warn("esmlc.positioning.apply_failed", "association", association, "correlation", corr, "error", jobErr)
		return nil, fmt.Errorf("sls: apply positioning job: %w", jobErr)
	} else {
		s.log.Debug("esmlc.positioning.apply_not_active", "association", association, "correlation", corr, "events", len(r.Events))
	}
	out, err := s.wrapLPPActions(carrier.Correlation, actions)
	if err != nil {
		return nil, err
	}
	if final != nil {
		s.recordFinalOutcome(association, final)
		response, err := encodeFinalOutcome(carrier.Correlation, *final)
		if err != nil {
			return nil, err
		}
		out = append(out, response)
	}
	for _, event := range r.Events {
		s.log.Info("esmlc.lpp.event", "association", association, "correlation", corr, "kind", event.Kind)
	}
	s.log.Debug("esmlc.lpp.handled", "association", association, "correlation", corr, "actions", len(actions), "outbound_pdus", len(out))
	return out, nil
}

// handleLPPa is the SLs carrier boundary for payload type 1. LPPa rides the
// same Connection-Oriented Information procedure as LPP but is transparent to
// the MME (see TS 36.305): the MME relays it to the eNB without decoding it,
// so this server owns the same decode/apply/wrap responsibility it already
// owns for LPP, just against a different codec and Manager entry points.
func (s *Server) handleLPPa(association string, carrier lcsap.ConnectionOriented) ([][]byte, error) {
	corr := fmt.Sprintf("%x", carrier.Correlation)
	p, err := lppa.Decode(carrier.Payload)
	if err != nil {
		s.log.Warn("esmlc.lppa.decode_failed", "association", association, "correlation", corr, "error", err)
		return nil, fmt.Errorf("sls: decode LPPa APDU: %w", err)
	}
	s.log.Debug("esmlc.lppa.message_received", "association", association, "correlation", corr, "category", p.Category, "procedure", p.ProcedureCode)
	scope := positioning.Scope{Association: association, Correlation: carrier.Correlation}
	outcome, err := s.applyLPPaMessage(scope, p, s.now())
	if err != nil {
		if errors.Is(err, positioning.ErrNotActive) {
			s.log.Warn("esmlc.lppa.no_active_job", "association", association, "correlation", corr)
			return nil, nil
		}
		s.log.Warn("esmlc.lppa.apply_failed", "association", association, "correlation", corr, "error", err)
		return nil, err
	}
	var out [][]byte
	if outcome.LPPa != nil {
		w, err := s.wrapLPPaAction(carrier.Correlation, outcome.LPPa)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	if outcome.Snapshot.Final != nil {
		s.recordFinalOutcome(association, outcome.Snapshot.Final)
		response, err := encodeFinalOutcome(carrier.Correlation, *outcome.Snapshot.Final)
		if err != nil {
			return nil, err
		}
		out = append(out, response)
	}
	s.log.Debug("esmlc.lppa.handled", "association", association, "correlation", corr, "outbound_pdus", len(out))
	return out, nil
}

// applyLPPaMessage dispatches a decoded LPPa PDU to the Manager entry point
// matching its category and procedure code. Only the E-CID Measurement
// procedure family is implemented; anything else is rejected fail-closed.
func (s *Server) applyLPPaMessage(scope positioning.Scope, p lppa.PDU, now time.Time) (positioning.Outcome, error) {
	switch {
	case p.Category == lppa.Successful && p.ProcedureCode == lppa.ProcedureECIDMeasurementInitiation:
		resp, err := lppa.DecodeInitiationResponse(p)
		if err != nil {
			return positioning.Outcome{}, fmt.Errorf("sls: decode LPPa initiation response: %w", err)
		}
		return s.jobs.ApplyLPPaInitiationResponse(scope, p.TransactionID, resp, now)
	case p.Category == lppa.Unsuccessful && p.ProcedureCode == lppa.ProcedureECIDMeasurementInitiation:
		fail, err := lppa.DecodeInitiationFailure(p)
		if err != nil {
			return positioning.Outcome{}, fmt.Errorf("sls: decode LPPa initiation failure: %w", err)
		}
		return s.jobs.ApplyLPPaInitiationFailure(scope, p.TransactionID, fail, now)
	case p.Category == lppa.Initiating && p.ProcedureCode == lppa.ProcedureECIDMeasurementReport:
		report, err := lppa.DecodeReport(p)
		if err != nil {
			return positioning.Outcome{}, fmt.Errorf("sls: decode LPPa report: %w", err)
		}
		return s.jobs.ApplyLPPaReport(scope, report, now)
	case p.Category == lppa.Initiating && p.ProcedureCode == lppa.ProcedureECIDMeasurementFailureIndication:
		indication, err := lppa.DecodeFailureIndication(p)
		if err != nil {
			return positioning.Outcome{}, fmt.Errorf("sls: decode LPPa failure indication: %w", err)
		}
		return s.jobs.ApplyLPPaFailureIndication(scope, indication, now)
	default:
		return positioning.Outcome{}, fmt.Errorf("sls: unsupported LPPa message (category=%d procedure=%d)", p.Category, p.ProcedureCode)
	}
}

// wrapLPPaAction builds the LPPa message a positioning.LPPaAction requests
// and carries it in Connection-Oriented Information payload type 1. The
// Initiation Request always asks for on-demand reporting of the cell-ID
// quantity only — the bounded scope this method implements — never the
// InterRAT/WLAN quantities or periodic reporting.
func (s *Server) wrapLPPaAction(correlation [4]byte, action *positioning.LPPaAction) ([]byte, error) {
	var pdu lppa.PDU
	var err error
	switch action.Kind {
	case positioning.LPPaSendInitiationRequest:
		pdu, err = lppa.BuildInitiationRequest(action.TransactionID, action.ESMLCMeasurementID, lppa.ReportOnDemand, nil, []lppa.MeasurementQuantityValue{lppa.QuantityCellID})
	case positioning.LPPaSendTerminationCommand:
		pdu, err = lppa.BuildTerminationCommand(action.TransactionID, action.ESMLCMeasurementID, action.ENBMeasurementID)
	default:
		return nil, fmt.Errorf("sls: unsupported LPPa action")
	}
	if err != nil {
		return nil, fmt.Errorf("sls: build LPPa message: %w", err)
	}
	encoded, err := lppa.Encode(pdu)
	if err != nil {
		return nil, fmt.Errorf("sls: encode LPPa message: %w", err)
	}
	wire, err := lcsap.EncodeConnectionOriented(lcsap.ConnectionOriented{Correlation: correlation, PayloadType: 1, Payload: encoded}, s.cfg.SLs.MaxMessageSize)
	if err != nil {
		return nil, fmt.Errorf("sls: encode LPPa carrier action: %w", err)
	}
	return wire, nil
}

// encodeFinalOutcome is the sole LCS delivery mapping. The recovered LCS-AP
// subset represents the standards root radio-network-layer/unspecified cause
// as zero; detailed cause alternatives require the remaining LCS-Cause CHOICE
// codec and are intentionally not fabricated here.
func encodeFinalOutcome(correlation [4]byte, final positioning.FinalOutcome) ([]byte, error) {
	if final.Kind == positioning.FinalEstimateAvailable && final.Estimate != nil {
		var data *lcsap.PositioningData
		switch final.Estimate.Source {
		case positioning.EstimateSourceAuthoritativeServingCell, positioning.EstimateSourceLPPaAccessPointPosition:
			v := lcsap.NewECIDPositioningData()
			data = &v
		case positioning.EstimateSourceOTDOAMultilateration:
			v := lcsap.NewOTDOAPositioningData()
			data = &v
		case positioning.EstimateSourceAGNSSUEReported:
			v := lcsap.NewAGNSSPositioningData()
			data = &v
		}
		var accuracy *lcsap.AccuracyFulfillmentIndicator
		if final.Accuracy == positioning.AccuracyFulfilled {
			v := lcsap.AccuracyFulfilled
			accuracy = &v
		}
		response, err := lcsap.LocationResponseWithMetadata(correlation, final.Estimate.Latitude, final.Estimate.Longitude, final.Estimate.HorizontalUncertainty, data, accuracy)
		if err != nil {
			return nil, fmt.Errorf("sls: encode positioning estimate: %w", err)
		}
		return response, nil
	}
	failure, err := lcsap.FailureWithDetailedCause(correlation, lcsCauseForOutcome(final))
	if err != nil {
		return nil, fmt.Errorf("sls: encode positioning failure: %w", err)
	}
	return failure, nil
}

// lcsCauseForOutcome maps a terminal job outcome to the recovered TS 29.171
// root LCS-Cause. The root enumeration itself is complete for this release
// (Radio-Network-Layer-Cause has only "unspecified"; Transport, Protocol, and
// Misc are otherwise exhausted too — see docs/lpp-spec-audit.md), so the only
// available refinement is which of the existing root values best fits each
// outcome. FinalNoEligibleMethod and FinalMeasurementsWithoutEstimator are
// both a direct consequence of local operator configuration (ECID disabled,
// or no estimator/catalog wired up) rather than a protocol or radio failure,
// so they get Misc/o-And-M-Intervention; everything else without a more
// specific standards value stays at Misc/unspecified.
func lcsCauseForOutcome(final positioning.FinalOutcome) lcsap.LCSCause {
	switch final.Kind {
	case positioning.FinalProcedureFailure:
		return lcsap.LCSCause{Branch: lcsap.LCSCauseProtocol, Value: lcsap.ProtocolUnspecified}
	case positioning.FinalNoEligibleMethod, positioning.FinalMeasurementsWithoutEstimator:
		return lcsap.LCSCause{Branch: lcsap.LCSCauseMisc, Value: lcsap.MiscOMIntervention}
	case positioning.FinalEstimationFailed, positioning.FinalQualityNotMet, positioning.FinalDeadlineExpired, positioning.FinalCancelled:
		return lcsap.LCSCause{Branch: lcsap.LCSCauseMisc, Value: lcsap.MiscUnspecified}
	default:
		return lcsap.LCSCause{Branch: lcsap.LCSCauseRadioNetwork, Value: lcsap.RadioNetworkUnspecified}
	}
}

func (s *Server) wrapLPPActions(correlation [4]byte, actions []procedure.Action) ([][]byte, error) {
	out := make([][]byte, 0, len(actions))
	for _, action := range actions {
		encoded, err := lpp.EncodeMessage(action.Message)
		if err != nil {
			return nil, fmt.Errorf("sls: encode LPP action: %w", err)
		}
		wire, err := lcsap.EncodeConnectionOriented(lcsap.ConnectionOriented{Correlation: correlation, PayloadType: 0, Payload: encoded.Bytes}, s.cfg.SLs.MaxMessageSize)
		if err != nil {
			return nil, fmt.Errorf("sls: encode LPP carrier action: %w", err)
		}
		out = append(out, wire)
	}
	return out, nil
}

func (s *Server) lppProcedure(id session.ID) (*procedure.Orchestrator, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if o := s.lpp[id]; o != nil {
		return o, nil
	}
	if len(s.lpp) >= s.cfg.SLs.MaxSessions {
		return nil, fmt.Errorf("sls: LPP procedure capacity exhausted")
	}
	store, err := transaction.NewStore(transaction.DefaultLimits())
	if err != nil {
		return nil, fmt.Errorf("sls: LPP transaction store: %w", err)
	}
	o, err := procedure.New(store, procedure.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("sls: LPP procedure: %w", err)
	}
	s.lpp[id] = o
	return o, nil
}

func (s *Server) dropLPPAssociation(association string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id := range s.lpp {
		if id.Association == association {
			delete(s.lpp, id)
		}
	}
}

// dropLPPCorrelation removes only the one LPP procedure entry matching id,
// unlike dropLPPAssociation's association-wide sweep — used by Location
// Abort, which TS 29.171 scopes to a single Correlation-ID, not every
// outstanding correlation on the association.
func (s *Server) dropLPPCorrelation(id session.ID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.lpp, id)
}
// pruneLoop periodically expires positioning jobs whose deadline has passed
// without a triggering inbound event (see positioning.Manager.Prune) until
// ctx is done. It runs for the life of the SLs listener, tracked by s.wg
// alongside the accept loop and every association's read loop, so Close
// waits for it to stop cleanly.
func (s *Server) pruneLoop(ctx context.Context) {
	t := time.NewTicker(s.cfg.SLs.PruneInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.prune(s.now())
		}
	}
}

// prune expires every job positioning.Manager.Prune finds past its
// deadline, sending the same wire actions (LPP actions to the UE, an LCS-AP
// failure to the MME) Apply would have sent had a reactive event triggered
// this same expiry, then drops the now-unneeded LPP session state for that
// Scope (see dropLPPCorrelation) so it does not accumulate for the life of
// the process.
func (s *Server) prune(now time.Time) {
	for _, r := range s.jobs.Prune(now) {
		corr := fmt.Sprintf("%x", r.Scope.Correlation)
		s.log.Info("esmlc.positioning.job_expired", "association", r.Scope.Association, "correlation", corr,
			"note", "expired by periodic sweep; no further LPP/LPPa event ever arrived for this correlation")
		out, err := s.wrapLPPActions(r.Scope.Correlation, r.Outcome.Actions)
		if err != nil {
			s.log.Warn("esmlc.positioning.prune_wrap_failed", "association", r.Scope.Association, "correlation", corr, "error", err)
			out = nil
		}
		if r.Outcome.LPPa != nil {
			w, err := s.wrapLPPaAction(r.Scope.Correlation, r.Outcome.LPPa)
			if err != nil {
				s.log.Warn("esmlc.positioning.prune_wrap_lppa_failed", "association", r.Scope.Association, "correlation", corr, "error", err)
			} else {
				out = append(out, w)
			}
		}
		if final := r.Outcome.Snapshot.Final; final != nil {
			s.recordFinalOutcome(r.Scope.Association, final)
			response, err := encodeFinalOutcome(r.Scope.Correlation, *final)
			if err != nil {
				s.log.Warn("esmlc.positioning.prune_encode_failed", "association", r.Scope.Association, "correlation", corr, "error", err)
			} else {
				out = append(out, response)
			}
		}
		s.dropLPPCorrelation(session.ID{Association: r.Scope.Association, Correlation: r.Scope.Correlation})
		if len(out) == 0 {
			continue
		}
		s.mu.Lock()
		a := s.assocs[r.Scope.Association]
		s.mu.Unlock()
		if a == nil {
			continue // association already gone; nothing to write to
		}
		for _, w := range out {
			if s.debugEnabled() {
				s.log.Debug("esmlc.sls.pdu_write", "association", r.Scope.Association, "bytes", len(w), "hex", hexPreview(w))
			}
			if e := a.send(w); e != nil {
				a.close(e)
				break
			}
		}
	}
}

func (s *Server) locationRequest(association string, p lcsap.PDU) ([][]byte, error) {
	id, e := lcsap.ValidateLocationRequest(p)
	if e != nil {
		s.log.Warn("esmlc.lcsap.location_request_invalid", "association", association, "error", e)
		return nil, e
	}
	s.log.Info("esmlc.lcsap.location_request_received", "association", association, "correlation", fmt.Sprintf("%x", id))
	return s.startPositioningJob(association, id, p)
}

func (s *Server) startPositioningJob(association string, correlation [4]byte, p lcsap.PDU) ([][]byte, error) {
	corr := fmt.Sprintf("%x", correlation)
	request, err := lcsap.DecodeLocationRequest(p)
	if err != nil {
		s.log.Warn("esmlc.lcsap.location_request_decode_failed", "association", association, "correlation", corr, "error", err)
		return nil, err
	}
	s.log.Debug("esmlc.lcsap.location_request_fields", "association", association, "correlation", corr,
		"location_type", request.LocationType, "ecgi", fmt.Sprintf("%x", request.ECGI), "priority", request.Priority, "lpp_supported", request.LPPSupported, "client_type", request.ClientType)
	if request.ClientType == nil {
		s.log.Warn("esmlc.lcsap.location_request_no_client_type", "association", association, "correlation", corr,
			"note", "LCS-Client-Type IE absent; cannot distinguish emergency-services from other LCS clients for this request")
	} else if *request.ClientType != lcsap.ClientTypeEmergencyServices {
		s.log.Info("esmlc.lcsap.location_request_client_type", "association", association, "correlation", corr, "client_type", *request.ClientType,
			"note", "non-emergency LCS client; this service does not implement TS 23.271 privacy notification/verification, so the UE may decline")
	}
	scope := positioning.Scope{Association: association, Correlation: correlation}
	o, err := s.lppProcedure(session.ID{Association: association, Correlation: correlation})
	if err != nil {
		s.log.Warn("esmlc.sls.lpp_procedure_unavailable", "association", association, "correlation", corr, "error", err, "consequence", "no LCS-AP response will be sent for this location request")
		return nil, err
	}
	now := s.now()
	deadline := now.Add(s.cfg.SLs.SessionTimeout)
	result, err := s.jobs.Start(positioning.Request{Scope: scope, LocationType: request.LocationType, ServingECGI: request.ECGI, Priority: request.Priority, QoS: positioningQoS(request.QoS), LPPSupported: request.LPPSupported, Deadline: deadline}, o, now)
	if err != nil {
		// A non-nil error here means the job could not even be scheduled
		// (duplicate job, transaction-store exhaustion, malformed
		// procedure options): the MME sent a Location Request and gets
		// nothing back on the wire for it. Logged at Warn regardless of
		// -d so this specific failure mode is always visible, not just
		// when debug logging happens to be on.
		s.log.Warn("esmlc.positioning.job_start_failed", "association", association, "correlation", corr, "error", err, "consequence", "no LCS-AP response will be sent for this location request")
		return nil, fmt.Errorf("sls: start positioning job: %w", err)
	}
	s.log.Debug("esmlc.positioning.job_started", "association", association, "correlation", corr, "method", result.Snapshot.Method, "state", result.Snapshot.State, "actions", len(result.Actions), "lppa_action", result.LPPa != nil, "deadline", deadline,
		"note", "deadline is only checked reactively on the next inbound LPP/LPPa event for this correlation; a UE/eNB that never replies again will not get a proactive LCS-AP failure response before this deadline")
	s.recordFinalOutcome(association, result.Snapshot.Final)
	if result.Snapshot.State == positioning.NoEligibleMethod {
		w, err := lcsap.FailureWithCause(correlation, lcsap.CauseMiscUnspecified)
		if err != nil {
			return nil, err
		}
		return [][]byte{w}, nil
	}
	out, err := s.wrapLPPActions(correlation, result.Actions)
	if err != nil {
		return nil, err
	}
	if result.LPPa != nil {
		w, err := s.wrapLPPaAction(correlation, result.LPPa)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	if len(out) == 0 {
		s.log.Warn("esmlc.positioning.job_started_no_wire_output", "association", association, "correlation", corr, "state", result.Snapshot.State, "consequence", "no LCS-AP response will be sent for this location request")
	}
	return out, nil
}

// locationAbort handles TS 29.171's Location Abort procedure, scoped to the
// single Correlation-ID the request names (not the whole association, the
// way Reset legitimately is). Canceling the job may itself produce an LPP
// Abort to the UE and/or an LPPa Termination Command to the eNB (via
// positioning.Manager.Cancel's shared terminalLocked/finishLocked path, the
// same one every other terminal outcome uses) — both are sent alongside the
// procedure's own Successful Outcome. Per the elementary procedure table
// there is no Unsuccessful Outcome for this procedure, so the acknowledgment
// is sent whether or not a matching job was found.
func (s *Server) locationAbort(association string, p lcsap.PDU) ([][]byte, error) {
	req, err := lcsap.DecodeLocationAbortRequest(p)
	if err != nil {
		s.log.Warn("esmlc.lcsap.location_abort_decode_failed", "association", association, "error", err)
		return nil, err
	}
	corr := fmt.Sprintf("%x", req.Correlation)
	scope := positioning.Scope{Association: association, Correlation: req.Correlation}
	now := s.now()
	outcome, found := s.jobs.Cancel(scope, now)
	var out [][]byte
	if found {
		s.recordFinalOutcome(association, outcome.Snapshot.Final)
		actions, err := s.wrapLPPActions(req.Correlation, outcome.Actions)
		if err != nil {
			return nil, err
		}
		out = append(out, actions...)
		if outcome.LPPa != nil {
			w, err := s.wrapLPPaAction(req.Correlation, outcome.LPPa)
			if err != nil {
				return nil, err
			}
			out = append(out, w)
		}
	}
	s.dropLPPCorrelation(session.ID{Association: association, Correlation: req.Correlation})
	ack, err := lcsap.AbortAcknowledge(req.Correlation)
	if err != nil {
		return nil, err
	}
	out = append(out, ack)
	s.log.Info("esmlc.lcsap.location_abort", "association", association, "correlation", corr, "cause_branch", req.Cause.Branch, "cause_value", req.Cause.Value, "job_found", found)
	return out, nil
}

func positioningQoS(v *lcsap.QoS) *positioning.QoS {
	if v == nil {
		return nil
	}
	return &positioning.QoS{HorizontalAccuracy: v.HorizontalAccuracy, VerticalRequested: v.VerticalRequested, VerticalAccuracy: v.VerticalAccuracy, ResponseTime: v.ResponseTime}
}

func positioningPolicy(cfg config.Config) positioning.Policy {
	p := cfg.Positioning.ECID
	policy := positioning.Policy{ECID: positioning.ECIDPolicy{Enabled: p.Enabled}, OTDOA: positioning.OTDOAPolicy{Enabled: cfg.Positioning.OTDOA.Enabled}, AGNSS: positioning.AGNSSPolicy{Enabled: cfg.Positioning.AGNSS.Enabled}, LPPaECID: positioning.LPPaECIDPolicy{Enabled: cfg.Positioning.LPPaECID.Enabled}}
	if !p.Enabled {
		return policy
	}
	length := 0
	raw := byte(0)
	if p.RequestRSRP {
		raw |= 0x80
		length = 1
	}
	if p.RequestRSRQ {
		raw |= 0x40
		if length < 2 {
			length = 2
		}
	}
	if p.RequestUERxTxTimeDiff {
		raw |= 0x20
		if length < 3 {
			length = 3
		}
	}
	bits, err := uper.NewBitString([]byte{raw}, length)
	if err == nil {
		policy.ECID.RequestedMeasurements = location.ECIDRequestLocationInformation{RequestedMeasurements: bits}
	}
	return policy
}
