// Package sls owns SCTP association lifecycle and LCS-AP dispatch.
package sls

import (
	"context"
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
		[]string{"estimate_available", "measurements_without_estimator", "estimation_failed", "quality_not_met", "no_eligible_method", "procedure_failure", "deadline_expired", "cancelled", "unknown"})
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
	go func() { <-ctx.Done(); _ = s.Close() }()
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
		if n < 1 || n > len(buf) || info == nil || info.PPID != a.server.cfg.SLs.ExpectedPPID {
			a.server.log.Warn("esmlc.malformed_message", "reason", "ppid_or_size")
			continue
		}
		out, e := a.server.Handle(a.id, append([]byte(nil), buf[:n]...))
		if e != nil {
			a.server.log.Warn("esmlc.lcsap.error", "association", a.id, "error", e)
			continue
		}
		for _, w := range out {
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
		return nil, e
	}
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
		s.log.Info("esmlc.lcsap.connection_oriented_received", "association", association, "payload_type", v.PayloadType, "payload_length", len(v.Payload))
		switch v.PayloadType {
		case 0:
			return s.handleLPP(association, v)
		case 1:
			return s.handleLPPa(association, v)
		default:
			return nil, fmt.Errorf("lcsap: unsupported payload type %d", v.PayloadType)
		}
	case lcsap.ProcedureLocationAbort:
		s.sessions.DropAssociation(association)
		s.dropLPPAssociation(association)
		s.jobs.DropAssociation(association)
		return nil, nil
	default:
		return nil, fmt.Errorf("lcsap: unsupported procedure %d", p.Procedure)
	}
}

// handleLPP is the SLs carrier boundary for payload type 0. LCS-AP keeps the
// APDU opaque; this server is the owning boundary that decodes it, applies the
// transport-neutral procedure, and wraps any resulting LPP actions back into
// Connection-Oriented Information using the distinct LCS correlation ID.
func (s *Server) handleLPP(association string, carrier lcsap.ConnectionOriented) ([][]byte, error) {
	m, err := lpp.DecodeMessageOctets(carrier.Payload)
	if err != nil {
		return nil, fmt.Errorf("sls: decode LPP APDU: %w", err)
	}
	o, err := s.lppProcedure(session.ID{Association: association, Correlation: carrier.Correlation})
	if err != nil {
		return nil, err
	}
	r, err := o.HandleInbound(m, s.now())
	if err != nil {
		return nil, fmt.Errorf("sls: apply LPP procedure: %w", err)
	}
	actions := append([]procedure.Action(nil), r.Actions...)
	jobResult, jobErr := s.jobs.Apply(positioning.Scope{Association: association, Correlation: carrier.Correlation}, r.Events, s.now())
	var final *positioning.FinalOutcome
	if jobErr == nil {
		actions = append(actions, jobResult.Actions...)
		final = jobResult.Snapshot.Final
	} else if jobErr != positioning.ErrNotActive {
		return nil, fmt.Errorf("sls: apply positioning job: %w", jobErr)
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
		s.log.Info("esmlc.lpp.event", "association", association, "correlation", fmt.Sprintf("%x", carrier.Correlation), "kind", event.Kind)
	}
	return out, nil
}

// handleLPPa is the SLs carrier boundary for payload type 1. LPPa rides the
// same Connection-Oriented Information procedure as LPP but is transparent to
// the MME (see TS 36.305): the MME relays it to the eNB without decoding it,
// so this server owns the same decode/apply/wrap responsibility it already
// owns for LPP, just against a different codec and Manager entry points.
func (s *Server) handleLPPa(association string, carrier lcsap.ConnectionOriented) ([][]byte, error) {
	p, err := lppa.Decode(carrier.Payload)
	if err != nil {
		return nil, fmt.Errorf("sls: decode LPPa APDU: %w", err)
	}
	scope := positioning.Scope{Association: association, Correlation: carrier.Correlation}
	outcome, err := s.applyLPPaMessage(scope, p, s.now())
	if err != nil {
		if errors.Is(err, positioning.ErrNotActive) {
			s.log.Warn("esmlc.lppa.no_active_job", "association", association, "correlation", fmt.Sprintf("%x", carrier.Correlation))
			return nil, nil
		}
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
		if final.Estimate.Source == positioning.EstimateSourceAuthoritativeServingCell {
			v := lcsap.NewECIDPositioningData()
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
func (s *Server) locationRequest(association string, p lcsap.PDU) ([][]byte, error) {
	id, e := lcsap.ValidateLocationRequest(p)
	if e != nil {
		return nil, e
	}
	s.log.Info("esmlc.lcsap.location_request_received", "association", association)
	return s.startPositioningJob(association, id, p)
}

func (s *Server) startPositioningJob(association string, correlation [4]byte, p lcsap.PDU) ([][]byte, error) {
	request, err := lcsap.DecodeLocationRequest(p)
	if err != nil {
		return nil, err
	}
	scope := positioning.Scope{Association: association, Correlation: correlation}
	o, err := s.lppProcedure(session.ID{Association: association, Correlation: correlation})
	if err != nil {
		return nil, err
	}
	now := s.now()
	result, err := s.jobs.Start(positioning.Request{Scope: scope, LocationType: request.LocationType, ServingECGI: request.ECGI, Priority: request.Priority, QoS: positioningQoS(request.QoS), LPPSupported: request.LPPSupported, Deadline: now.Add(s.cfg.SLs.SessionTimeout)}, o, now)
	if err != nil {
		return nil, fmt.Errorf("sls: start positioning job: %w", err)
	}
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
