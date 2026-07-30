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
	wg       sync.WaitGroup
}
type association struct {
	id     string
	conn   *sctp.SCTPConn
	write  sync.Mutex
	server *Server
	once   sync.Once
}

func New(cfg config.Config, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	policy := positioningPolicy(cfg)
	var estimator positioning.Estimator
	if cfg.Positioning.ECID.CellDataFile != "" {
		catalog, err := positioning.LoadCellCatalog(cfg.Positioning.ECID.CellDataFile, cfg.Positioning.ECID.CellDataMaxAge, time.Now())
		if err != nil {
			log.Error("esmlc.cell_catalog_unavailable", "error", err)
			estimator = unavailableEstimator{}
		} else {
			estimator = positioning.ServingCellEstimator{Catalog: catalog}
			log.Info("esmlc.cell_catalog_loaded", "version", catalog.Version())
		}
	} else if cfg.Positioning.Simulation.Enabled {
		simFailure := positioning.EstimationFailure(0)
		if cfg.Positioning.Simulation.FailureCause != nil {
			simFailure = positioning.InsufficientNetworkData
		}
		estimator = positioning.SimulationEstimator{Latitude: cfg.Positioning.Simulation.Latitude, Longitude: cfg.Positioning.Simulation.Longitude, Uncertainty: positioning.UncertaintyCodeFromMeters(cfg.Positioning.Simulation.UncertaintyMeters), Failure: simFailure}
	}
	return &Server{cfg: cfg, log: log, sessions: session.New(cfg.SLs.MaxSessions), assocs: map[string]*association{}, lpp: map[session.ID]*procedure.Orchestrator{}, jobs: positioning.NewWithEstimator(policy, estimator)}
}

type unavailableEstimator struct{}

func (unavailableEstimator) Estimate(positioning.Request, positioning.MethodResult, time.Time) positioning.EstimationResult {
	return positioning.EstimationResult{Failure: positioning.InsufficientNetworkData}
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
		if v.PayloadType != 0 {
			return nil, fmt.Errorf("lcsap: LPPa is unsupported")
		}
		return s.handleLPP(association, v)
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
	r, err := o.HandleInbound(m, time.Now())
	if err != nil {
		return nil, fmt.Errorf("sls: apply LPP procedure: %w", err)
	}
	actions := append([]procedure.Action(nil), r.Actions...)
	jobResult, jobErr := s.jobs.Apply(positioning.Scope{Association: association, Correlation: carrier.Correlation}, r.Events, time.Now())
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

// encodeFinalOutcome is the sole LCS delivery mapping. The recovered LCS-AP
// subset represents the standards root radio-network-layer/unspecified cause
// as zero; detailed cause alternatives require the remaining LCS-Cause CHOICE
// codec and are intentionally not fabricated here.
func encodeFinalOutcome(correlation [4]byte, final positioning.FinalOutcome) ([]byte, error) {
	if final.Kind == positioning.FinalEstimateAvailable && final.Estimate != nil {
		response, err := lcsap.LocationResponse(correlation, final.Estimate.Latitude, final.Estimate.Longitude, final.Estimate.HorizontalUncertainty)
		if err != nil {
			return nil, fmt.Errorf("sls: encode positioning estimate: %w", err)
		}
		return response, nil
	}
	failure, err := lcsap.FailureWithCause(correlation, lcsCauseForOutcome(final))
	if err != nil {
		return nil, fmt.Errorf("sls: encode positioning failure: %w", err)
	}
	return failure, nil
}

func lcsCauseForOutcome(final positioning.FinalOutcome) lcsap.Cause {
	switch final.Kind {
	case positioning.FinalProcedureFailure:
		return lcsap.CauseProtocolUnspecified
	case positioning.FinalNoEligibleMethod, positioning.FinalMeasurementsWithoutEstimator, positioning.FinalEstimationFailed, positioning.FinalQualityNotMet, positioning.FinalDeadlineExpired, positioning.FinalCancelled:
		return lcsap.CauseMiscUnspecified
	default:
		return lcsap.CauseRadioNetworkUnspecified
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
	result, err := s.jobs.Start(positioning.Request{Scope: scope, LocationType: request.LocationType, ServingECGI: request.ECGI, Priority: request.Priority, QoS: positioningQoS(request.QoS), LPPSupported: request.LPPSupported, Deadline: time.Now().Add(s.cfg.SLs.SessionTimeout)}, o, time.Now())
	if err != nil {
		return nil, fmt.Errorf("sls: start positioning job: %w", err)
	}
	if result.Snapshot.State == positioning.NoEligibleMethod {
		w, err := lcsap.FailureWithCause(correlation, lcsap.CauseMiscUnspecified)
		if err != nil {
			return nil, err
		}
		return [][]byte{w}, nil
	}
	return s.wrapLPPActions(correlation, result.Actions)
}

func positioningQoS(v *lcsap.QoS) *positioning.QoS {
	if v == nil {
		return nil
	}
	return &positioning.QoS{HorizontalAccuracy: v.HorizontalAccuracy, VerticalRequested: v.VerticalRequested, VerticalAccuracy: v.VerticalAccuracy, ResponseTime: v.ResponseTime}
}

func positioningPolicy(cfg config.Config) positioning.Policy {
	p := cfg.Positioning.ECID
	policy := positioning.Policy{ECID: positioning.ECIDPolicy{Enabled: p.Enabled}}
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
