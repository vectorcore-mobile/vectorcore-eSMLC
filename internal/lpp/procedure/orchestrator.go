package procedure

import (
	"fmt"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/capability"
	"github.com/vectorcore/esmlc/internal/lpp/location"
	"github.com/vectorcore/esmlc/internal/lpp/transaction"
)

func New(store *transaction.Store, cfg Config) (*Orchestrator, error) {
	if store == nil || cfg.MaxPendingApplications <= 0 || (cfg.LocalInitiator != lpp.InitiatorLocationServer && cfg.LocalInitiator != lpp.InitiatorTargetDevice) {
		return nil, ErrInvalidConfig
	}
	return &Orchestrator{tx: store, cfg: cfg, waits: make(map[transaction.Key]pendingWait)}, nil
}
func (o *Orchestrator) StartCapabilities(options StartOptions, now time.Time) (Result, error) {
	snap, err := o.tx.CreateLocal(o.cfg.LocalInitiator, transaction.Capabilities, now)
	if err != nil {
		return Result{}, err
	}
	ack := o.cfg.DefaultRequestAcknowledgement
	if options.RequestAcknowledgement != nil {
		ack = *options.RequestAcknowledgement
	}
	payload := capability.RequestCapabilitiesR9IEs{}
	if options.RequestECID {
		payload.ECID = &capability.ECIDRequestCapabilities{}
	}
	m := message(snap.Key, lpp.BodyRequestCapabilities, false, ack, nil)
	m.Body.RequestCapabilities = &payload
	tr, err := o.tx.Apply(transaction.Outbound, transaction.LocallyInitiated, m, now)
	if err != nil {
		return Result{}, err
	}
	return resultFor(tr, []Action{{Kind: SendInitialRequest, Message: m, Key: keyPtr(snap.Key)}}), nil
}
func (o *Orchestrator) StartLocationInformation(options StartLocationInformationOptions, now time.Time) (Result, error) {
	payload := location.RequestLocationInformationR9IEs{}
	if options.ECID != nil {
		payload.ECID = cloneECIDRequest(options.ECID)
	}
	if err := payload.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrInvalidLocationRequest, err)
	}
	snap, err := o.tx.CreateLocal(o.cfg.LocalInitiator, transaction.LocationInformation, now)
	if err != nil {
		return Result{}, err
	}
	ack := o.cfg.DefaultRequestAcknowledgement
	if options.RequestAcknowledgement != nil {
		ack = *options.RequestAcknowledgement
	}
	m := message(snap.Key, lpp.BodyRequestLocationInformation, false, ack, nil)
	m.Body.RequestLocationInformation = cloneLocationRequest(&payload)
	tr, err := o.tx.Apply(transaction.Outbound, transaction.LocallyInitiated, m, now)
	if err != nil {
		return Result{}, err
	}
	r := resultFor(tr, []Action{{Kind: SendInitialRequest, Message: m, Key: keyPtr(snap.Key)}})
	r.Events = append(r.Events, Event{Kind: LocationInformationRequested, Key: keyPtr(snap.Key), Wait: WaitLocationInformation, LocationRequest: cloneLocationRequest(&payload), Snapshot: r.Snapshot})
	return r, nil
}
func (o *Orchestrator) start(p transaction.Procedure, body lpp.BodyKind, options StartOptions, now time.Time) (Result, error) {
	snap, err := o.tx.CreateLocal(o.cfg.LocalInitiator, p, now)
	if err != nil {
		return Result{}, err
	}
	ack := o.cfg.DefaultRequestAcknowledgement
	if options.RequestAcknowledgement != nil {
		ack = *options.RequestAcknowledgement
	}
	m := message(snap.Key, body, false, ack, nil)
	tr, err := o.tx.Apply(transaction.Outbound, transaction.LocallyInitiated, m, now)
	if err != nil {
		return Result{}, err
	}
	return resultFor(tr, []Action{{Kind: SendInitialRequest, Message: m, Key: keyPtr(snap.Key)}}), nil
}
func (o *Orchestrator) ProvideCapabilities(key transaction.Key, now time.Time) (Result, error) {
	return o.ProvideCapabilitiesResult(key, ProvideCapabilitiesOptions{}, now)
}
func (o *Orchestrator) ProvideCapabilitiesResult(key transaction.Key, options ProvideCapabilitiesOptions, now time.Time) (Result, error) {
	o.mu.Lock()
	pending := o.waits[key]
	o.mu.Unlock()
	if pending.kind != WaitCapabilities {
		return Result{}, ErrApplicationNotPending
	}
	if options.Capabilities.ECID != nil && !pending.requestedECID {
		return Result{}, ErrUnrequestedECID
	}
	if err := options.Capabilities.Validate(); err != nil {
		return Result{}, err
	}
	snap, ok := o.tx.Snapshot(key)
	if !ok {
		return Result{}, transaction.ErrNotFound
	}
	m := message(key, lpp.BodyProvideCapabilities, o.cfg.ResponseEndsTransaction, false, nil)
	m.Body.ProvideCapabilities = &options.Capabilities
	tr, err := o.tx.Apply(transaction.Outbound, snap.Ownership, m, now)
	if err != nil {
		return Result{}, err
	}
	o.clearWait(key)
	r := resultFor(tr, []Action{{Kind: SendProcedureResponse, Message: m, Key: keyPtr(key)}})
	r.Events = append([]Event{{Kind: CapabilitiesEnvelopeProvided, Key: keyPtr(key), Capabilities: &options.Capabilities, Snapshot: r.Snapshot}}, r.Events...)
	return r, nil
}

// ProvideLocationInformation completes a remotely initiated location request
// with a typed Release 9 provide-location payload. It deliberately performs no
// method selection, measurement execution, or estimate calculation.
func (o *Orchestrator) ProvideLocationInformation(key transaction.Key, options ProvideLocationInformationOptions, now time.Time) (Result, error) {
	o.mu.Lock()
	pending := o.waits[key]
	o.mu.Unlock()
	if pending.kind != WaitLocationInformation {
		return Result{}, ErrApplicationNotPending
	}
	if err := options.LocationInformation.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrInvalidLocationProvide, err)
	}
	snap, ok := o.tx.Snapshot(key)
	if !ok {
		return Result{}, transaction.ErrNotFound
	}
	payload := options.LocationInformation.Clone()
	m := message(key, lpp.BodyProvideLocationInformation, o.cfg.ResponseEndsTransaction, false, nil)
	m.Body.ProvideLocationInformation = &payload
	tr, err := o.tx.Apply(transaction.Outbound, snap.Ownership, m, now)
	if err != nil {
		return Result{}, err
	}
	o.clearWait(key)
	r := resultFor(tr, []Action{{Kind: SendProcedureResponse, Message: m, Key: keyPtr(key)}})
	r.Events = append([]Event{{Kind: LocationInformationEnvelopeProvided, Key: keyPtr(key), LocationInformation: &payload, Snapshot: r.Snapshot}}, r.Events...)
	return r, nil
}
func (o *Orchestrator) Abort(key transaction.Key, now time.Time) (Result, error) {
	return o.terminal(key, lpp.BodyAbort, SendAbort, ProcedureAborted, now)
}
func (o *Orchestrator) ReportError(key transaction.Key, now time.Time) (Result, error) {
	return o.terminal(key, lpp.BodyError, SendProtocolError, ProcedureFailed, now)
}
func (o *Orchestrator) terminal(key transaction.Key, body lpp.BodyKind, action ActionKind, event EventKind, now time.Time) (Result, error) {
	snap, ok := o.tx.Snapshot(key)
	if !ok {
		return Result{}, transaction.ErrNotFound
	}
	m := message(key, body, true, false, nil)
	tr, err := o.tx.Apply(transaction.Outbound, snap.Ownership, m, now)
	if err != nil {
		return Result{}, err
	}
	o.clearWait(key)
	r := resultFor(tr, []Action{{Kind: action, Message: m, Key: keyPtr(key)}})
	r.Events = append([]Event{{Kind: event, Key: keyPtr(key), Snapshot: r.Snapshot}}, r.Events...)
	return r, nil
}
func (o *Orchestrator) HandleInbound(m lpp.Message, now time.Time) (Result, error) {
	if err := m.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrUnsupportedMessage, err)
	}
	if m.TransactionID == nil {
		tr, err := o.tx.Apply(transaction.Inbound, transaction.RemotelyInitiated, m, now)
		if err != nil {
			return Result{}, err
		}
		return resultFor(tr, nil), nil
	}
	key := transaction.Key{Initiator: m.TransactionID.Initiator, Number: m.TransactionID.TransactionNumber}
	ownership := transaction.RemotelyInitiated
	if s, ok := o.tx.Snapshot(key); ok {
		ownership = s.Ownership
	}
	if isRequest(m.Body) {
		o.mu.Lock()
		_, existing := o.waits[key]
		full := !existing && len(o.waits) >= o.cfg.MaxPendingApplications
		o.mu.Unlock()
		if full {
			return Result{}, ErrCapacity
		}
	}
	tr, err := o.tx.Apply(transaction.Inbound, ownership, m, now)
	if err != nil {
		return Result{}, err
	}
	if tr.Classification == transaction.Duplicate {
		return resultFor(tr, nil).withEvent(Event{Kind: DuplicateIgnored, Key: keyPtr(key)}), nil
	}
	r := resultFor(tr, nil)
	if a := m.Acknowledgement; a != nil && a.Indicator != nil && tr.AcknowledgementMatched {
		r.Events = append(r.Events, Event{Kind: AcknowledgementReceived, Key: keyPtr(key), Snapshot: r.Snapshot})
	}
	if a := m.Acknowledgement; a != nil && a.Requested {
		ind := uint8(0)
		ack := lpp.Message{TransactionID: &lpp.TransactionID{Initiator: key.Initiator, TransactionNumber: key.Number}, Acknowledgement: &lpp.Acknowledgement{Indicator: &ind}, SequenceNumber: m.SequenceNumber}
		_, e := o.tx.Apply(transaction.Outbound, ownership, ack, now)
		if e != nil {
			return Result{}, e
		}
		r.Events = append(r.Events, Event{Kind: AcknowledgementRequested, Key: keyPtr(key)})
		r.Actions = append(r.Actions, Action{Kind: SendAcknowledgement, Message: ack, Key: keyPtr(key)})
		r.Snapshot = snapshotPtr(o, key)
	}
	if isRequest(m.Body) {
		wait := waitFor(m.Body)
		reqECID := wait == WaitCapabilities && m.Body.RequestCapabilities != nil && m.Body.RequestCapabilities.ECID != nil
		pending := pendingWait{kind: wait, requestedECID: reqECID}
		if wait == WaitLocationInformation && m.Body.RequestLocationInformation != nil {
			pending.locationRequest = cloneLocationRequest(m.Body.RequestLocationInformation)
		}
		if err := o.addWait(key, pending); err != nil {
			return Result{}, err
		}
		e := Event{Kind: requestEvent(wait), Key: keyPtr(key), Wait: wait, Snapshot: r.Snapshot}
		if wait == WaitCapabilities && m.Body.RequestCapabilities != nil {
			x := *m.Body.RequestCapabilities
			e.CapabilityRequest = &x
		}
		if wait == WaitLocationInformation && m.Body.RequestLocationInformation != nil {
			e.LocationRequest = cloneLocationRequest(m.Body.RequestLocationInformation)
		}
		r.Events = append(r.Events, e, Event{Kind: AwaitingApplicationResult, Key: keyPtr(key), Wait: wait, Snapshot: r.Snapshot})
	}
	if m.Body != nil && m.Body.Kind == lpp.BodyProvideCapabilities && m.Body.ProvideCapabilities != nil {
		x := *m.Body.ProvideCapabilities
		r.Events = append(r.Events, Event{Kind: CapabilitiesEnvelopeProvided, Key: keyPtr(key), Capabilities: &x, Snapshot: r.Snapshot})
	}
	if m.Body != nil && m.Body.Kind == lpp.BodyProvideLocationInformation && m.Body.ProvideLocationInformation != nil {
		x := m.Body.ProvideLocationInformation.Clone()
		r.Events = append(r.Events, Event{Kind: LocationInformationEnvelopeProvided, Key: keyPtr(key), LocationInformation: &x, Snapshot: r.Snapshot})
	}
	if m.Body != nil && m.Body.Kind == lpp.BodyAbort {
		o.clearWait(key)
		r.Events = append(r.Events, Event{Kind: ProcedureAborted, Key: keyPtr(key), Snapshot: r.Snapshot})
	}
	if m.Body != nil && m.Body.Kind == lpp.BodyError {
		o.clearWait(key)
		r.Events = append(r.Events, Event{Kind: ProcedureFailed, Key: keyPtr(key), Snapshot: r.Snapshot})
	}
	if tr.Completed {
		o.clearWait(key)
		r.Events = append(r.Events, Event{Kind: ProcedureCompleted, Key: keyPtr(key), Snapshot: r.Snapshot})
	}
	return r, nil
}
func (o *Orchestrator) Prune(now time.Time) PruneResult {
	p := o.tx.Prune(now)
	o.mu.Lock()
	n := 0
	for k := range o.waits {
		if _, ok := o.tx.Snapshot(k); !ok {
			delete(o.waits, k)
			n++
		}
	}
	o.mu.Unlock()
	return PruneResult{Transaction: p, ClearedApplicationWaits: n}
}
func (o *Orchestrator) Snapshot(key transaction.Key) (ProcedureSnapshot, bool) {
	s, ok := o.tx.Snapshot(key)
	if !ok {
		return ProcedureSnapshot{}, false
	}
	o.mu.Lock()
	w := o.waits[key]
	o.mu.Unlock()
	return ProcedureSnapshot{Transaction: s, Waiting: w.kind, RequestedECID: w.requestedECID, LocationRequest: cloneLocationRequest(w.locationRequest)}, true
}
func (o *Orchestrator) addWait(k transaction.Key, w pendingWait) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if old := o.waits[k]; old.kind != WaitNone {
		return ErrApplicationPending
	}
	if len(o.waits) >= o.cfg.MaxPendingApplications {
		return ErrCapacity
	}
	o.waits[k] = w
	return nil
}
func (o *Orchestrator) clearWait(k transaction.Key) { o.mu.Lock(); delete(o.waits, k); o.mu.Unlock() }
func message(k transaction.Key, b lpp.BodyKind, end, requestAck bool, seq *uint8) lpp.Message {
	m := lpp.Message{TransactionID: &lpp.TransactionID{Initiator: k.Initiator, TransactionNumber: k.Number}, EndTransaction: end, Body: &lpp.Body{Kind: b}, SequenceNumber: seq}
	if requestAck {
		m.Acknowledgement = &lpp.Acknowledgement{Requested: true}
	}
	return m
}
func isRequest(b *lpp.Body) bool {
	return b != nil && (b.Kind == lpp.BodyRequestCapabilities || b.Kind == lpp.BodyRequestLocationInformation)
}
func waitFor(b *lpp.Body) ApplicationWait {
	if b.Kind == lpp.BodyRequestCapabilities {
		return WaitCapabilities
	}
	return WaitLocationInformation
}
func requestEvent(w ApplicationWait) EventKind {
	if w == WaitCapabilities {
		return CapabilitiesRequested
	}
	return LocationInformationRequested
}
func resultFor(t transaction.Result, a []Action) Result {
	var s *transaction.Snapshot
	if t.HasKey {
		x := snapshotPtrResult(t)
		s = &x
	}
	r := Result{Actions: append([]Action(nil), a...), Snapshot: s}
	if t.Completed {
		r.Events = append(r.Events, Event{Kind: ProcedureCompleted, Key: keyPtr(t.Key), Snapshot: s})
	}
	return r
}
func snapshotPtr(o *Orchestrator, k transaction.Key) *transaction.Snapshot {
	s, ok := o.tx.Snapshot(k)
	if !ok {
		return nil
	}
	return &s
}
func snapshotPtrResult(t transaction.Result) transaction.Snapshot {
	return transaction.Snapshot{Key: t.Key, State: t.State}
}
func keyPtr(k transaction.Key) *transaction.Key { x := k; return &x }
func (r Result) withEvent(e Event) Result       { r.Events = append(r.Events, e); return r }

// cloneLocationRequest deliberately gives pending state and every event its
// own ECID container. uper.BitString is immutable, so its copied value cannot
// alias caller-owned byte storage.
func cloneLocationRequest(v *location.RequestLocationInformationR9IEs) *location.RequestLocationInformationR9IEs {
	if v == nil {
		return nil
	}
	out := *v
	out.ECID = cloneECIDRequest(v.ECID)
	return &out
}
func cloneECIDRequest(v *location.ECIDRequestLocationInformation) *location.ECIDRequestLocationInformation {
	if v == nil {
		return nil
	}
	out := *v
	return &out
}
