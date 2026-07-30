package transaction

import (
	"fmt"
	"sync"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp"
)

type fingerprint struct {
	bodyPresent                                bool
	body                                       lpp.BodyKind
	end                                        bool
	seqPresent                                 bool
	seq                                        uint8
	ackPresent, ackRequested, indicatorPresent bool
	indicator                                  uint8
	capabilityPresent                          bool
	capabilityBitLen                           uint8
	capabilityBytes                            [1]byte
	locationPresent                            bool
	locationBitLen                             uint8
	locationBytes                              [1]byte
	// encoded is the exact bounded UPER message representation. It covers
	// typed provide-location payloads without making transaction control own
	// their schema. The currently supported bounded LPP surface is below this
	// limit (the largest ECID list has 32 bounded elements).
	encodedLen uint16
	encoded    [1024]byte
}
type pendingAck struct {
	sequence    *uint8
	body        Procedure
	requestedAt time.Time
}
type record struct {
	key               Key
	ownership         Ownership
	state             State
	procedure         Procedure
	created, last     time.Time
	inbound, outbound *uint8
	pending           *pendingAck
	lastFP            fingerprint
	haveFP            bool
}
type Store struct {
	mu      sync.Mutex
	limits  Limits
	records map[Key]*record
	cursor  [2]uint8
}

func NewStore(l Limits) (*Store, error) {
	if l.MaxActive <= 0 || l.MaxActive > 512 || l.MaxRetained < 0 || l.ActiveLifetime <= 0 || l.RetentionLifetime <= 0 {
		return nil, ErrInvalidConfiguration
	}
	return &Store{limits: l, records: make(map[Key]*record)}, nil
}

func (s *Store) CreateLocal(initiator lpp.Initiator, procedure Procedure, now time.Time) (Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if procedure != Capabilities && procedure != LocationInformation {
		return Snapshot{}, ErrUnsupportedProcedure
	}
	if err := (Key{Initiator: initiator}).Validate(); err != nil {
		return Snapshot{}, err
	}
	if s.activeCountLocked() >= s.limits.MaxActive {
		return Snapshot{}, ErrCapacity
	}
	i := int(initiator)
	for n := 0; n < 256; n++ {
		number := s.cursor[i]
		s.cursor[i]++
		key := Key{initiator, number}
		if _, ok := s.records[key]; !ok {
			r := &record{key: key, ownership: LocallyInitiated, state: Created, procedure: procedure, created: now, last: now}
			s.records[key] = r
			return snapshot(r), nil
		}
	}
	return Snapshot{}, ErrNumberExhausted
}

func (s *Store) Snapshot(key Key) (Snapshot, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.records[key]
	if !ok {
		return Snapshot{}, false
	}
	return snapshot(r), true
}
func (s *Store) Remove(key Key) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.records[key]; !ok {
		return false
	}
	delete(s.records, key)
	return true
}

func (s *Store) Apply(direction Direction, ownership Ownership, msg lpp.Message, now time.Time) (Result, error) {
	if !direction.valid() {
		return Result{}, ErrInvalidDirection
	}
	if !ownership.valid() {
		return Result{}, ErrInvalidOwnership
	}
	if err := msg.Validate(); err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrInvalidMessage, err)
	}
	if msg.TransactionID == nil {
		return s.applyTransactionless(msg)
	}
	key := Key{msg.TransactionID.Initiator, msg.TransactionID.TransactionNumber}
	if err := key.Validate(); err != nil {
		return Result{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.records[key]
	if !ok {
		if ownership != RemotelyInitiated || direction != Inbound || !isRequest(msg.Body) || msg.EndTransaction {
			return Result{}, ErrNotFound
		}
		if s.activeCountLocked() >= s.limits.MaxActive {
			return Result{}, ErrCapacity
		}
		r = &record{key: key, ownership: ownership, state: Created, procedure: procedureFor(msg.Body), created: now, last: now}
		s.records[key] = r
	}
	if r.ownership != ownership {
		return Result{}, ErrInvalidOwnership
	}
	return s.applyLocked(r, direction, msg, now)
}

func (s *Store) applyTransactionless(msg lpp.Message) (Result, error) {
	// The bounded control layer intentionally never invents a key. A bounded
	// envelope without a procedure body is reported but never stored.
	if msg.Body != nil {
		return Result{}, fmt.Errorf("%w: body requires transaction identity", ErrInvalidMessage)
	}
	return Result{Classification: Transactionless, Transactionless: true, State: Active}, nil
}

func (s *Store) applyLocked(r *record, direction Direction, msg lpp.Message, now time.Time) (Result, error) {
	result := Result{HasKey: true, Key: r.key, Classification: New, PreviousState: r.state, State: r.state}
	fp := makeFingerprint(msg)
	if r.haveFP && r.lastFP == fp {
		result.Classification = Duplicate
		result.AcknowledgementPending = r.pending != nil
		result.Completed = r.state == Completed
		result.Aborted = r.state == Aborted
		result.Failed = r.state == Failed
		return result, nil
	}
	if r.state.terminal() {
		result.Classification = ReplayAfterCompletion
		return result, terminalError(r.state)
	}
	if err := s.checkSequenceLocked(r, direction, msg.SequenceNumber, fp); err != nil {
		return result, err
	}
	if err := validateProcedure(r, direction, msg); err != nil {
		return result, err
	}
	if msg.Body != nil && msg.Body.Kind == lpp.BodyAbort {
		if s.terminalCountLocked() >= s.limits.MaxRetained {
			return result, ErrCapacity
		}
		r.state = Aborted
		r.procedure = AbortProcedure
		r.pending = nil
	} else if msg.Body != nil && msg.Body.Kind == lpp.BodyError {
		if s.terminalCountLocked() >= s.limits.MaxRetained {
			return result, ErrCapacity
		}
		r.state = Failed
		r.procedure = ErrorProcedure
		r.pending = nil
	} else {
		if a := msg.Acknowledgement; a != nil && a.Indicator != nil {
			if r.pending == nil {
				return result, ErrAckNotPending
			}
			if r.pending.sequence != nil && (msg.SequenceNumber == nil || *r.pending.sequence != *msg.SequenceNumber) {
				return result, ErrAckMismatch
			}
			r.pending = nil
			result.AcknowledgementMatched = true
		}
		if a := msg.Acknowledgement; a != nil && a.Requested {
			if r.pending != nil {
				return result, ErrAckAlreadyPending
			}
			r.pending = &pendingAck{sequence: copyByte(msg.SequenceNumber), body: r.procedure, requestedAt: now}
		}
		if msg.EndTransaction {
			if s.terminalCountLocked() >= s.limits.MaxRetained {
				return result, ErrCapacity
			}
			r.state = Completed
			r.pending = nil
		} else if r.pending != nil {
			r.state = AwaitingAcknowledgement
		} else {
			r.state = Active
		}
	}
	r.last = now
	r.lastFP = fp
	r.haveFP = true
	result.State = r.state
	result.AcknowledgementPending = r.pending != nil
	result.Completed = r.state == Completed
	result.Aborted = r.state == Aborted
	result.Failed = r.state == Failed
	return result, nil
}

func validateProcedure(r *record, d Direction, m lpp.Message) error {
	if m.Body == nil {
		return nil
	}
	k := m.Body.Kind
	if k == lpp.BodyAbort || k == lpp.BodyError {
		return nil
	}
	p := procedureFor(m.Body)
	if p == EnvelopeOnly {
		return ErrUnsupportedProcedure
	}
	if r.procedure != p {
		return ErrInvalidTransition
	}
	if r.ownership == LocallyInitiated {
		if d == Outbound && !isRequest(m.Body) {
			return ErrInvalidTransition
		}
		if d == Inbound && isRequest(m.Body) {
			return ErrInvalidTransition
		}
	} else {
		if d == Inbound && !isRequest(m.Body) {
			return ErrInvalidTransition
		}
		if d == Outbound && isRequest(m.Body) {
			return ErrInvalidTransition
		}
	}
	return nil
}
func (s *Store) checkSequenceLocked(r *record, d Direction, v *uint8, fp fingerprint) error {
	if v == nil {
		return nil
	}
	last := &r.inbound
	if d == Outbound {
		last = &r.outbound
	}
	if *last == nil {
		*last = copyByte(v)
		return nil
	}
	old := **last
	if *v == old {
		return ErrSequenceConflict
	}
	if *v != old+1 {
		return ErrStaleSequence
	}
	*last = copyByte(v)
	return nil
}
func (s *Store) Prune(now time.Time) PruneResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out PruneResult
	for k, r := range s.records {
		age := now.Sub(r.last)
		if !r.state.terminal() && age >= s.limits.ActiveLifetime {
			delete(s.records, k)
			out.ExpiredActive++
		}
		if r.state.terminal() && age >= s.limits.RetentionLifetime {
			delete(s.records, k)
			out.RemovedRetained++
		}
	}
	return out
}
func (s *Store) activeCountLocked() int {
	n := 0
	for _, r := range s.records {
		if !r.state.terminal() {
			n++
		}
	}
	return n
}

func (s *Store) terminalCountLocked() int {
	n := 0
	for _, r := range s.records {
		if r.state.terminal() {
			n++
		}
	}
	return n
}
func snapshot(r *record) Snapshot {
	return Snapshot{Key: r.key, Ownership: r.ownership, State: r.state, Procedure: r.procedure, CreatedAt: r.created, LastActivity: r.last, LastInbound: copyByte(r.inbound), LastOutbound: copyByte(r.outbound), AcknowledgementPending: r.pending != nil}
}
func copyByte(v *uint8) *uint8 {
	if v == nil {
		return nil
	}
	x := *v
	return &x
}
func procedureFor(b *lpp.Body) Procedure {
	if b == nil {
		return EnvelopeOnly
	}
	switch b.Kind {
	case lpp.BodyRequestCapabilities, lpp.BodyProvideCapabilities:
		return Capabilities
	case lpp.BodyRequestLocationInformation, lpp.BodyProvideLocationInformation:
		return LocationInformation
	}
	return EnvelopeOnly
}
func isRequest(b *lpp.Body) bool {
	return b != nil && (b.Kind == lpp.BodyRequestCapabilities || b.Kind == lpp.BodyRequestLocationInformation)
}
func makeFingerprint(m lpp.Message) fingerprint {
	f := fingerprint{end: m.EndTransaction, seqPresent: m.SequenceNumber != nil}
	if m.SequenceNumber != nil {
		f.seq = *m.SequenceNumber
	}
	if m.Body != nil {
		f.bodyPresent = true
		f.body = m.Body.Kind
	}
	if a := m.Acknowledgement; a != nil {
		f.ackPresent = true
		f.ackRequested = a.Requested
		f.indicatorPresent = a.Indicator != nil
		if a.Indicator != nil {
			f.indicator = *a.Indicator
		}
	}
	if m.Body != nil && m.Body.ProvideCapabilities != nil && m.Body.ProvideCapabilities.ECID != nil {
		s := m.Body.ProvideCapabilities.ECID.MeasurementSupport
		b := s.Bytes()
		f.capabilityPresent = true
		f.capabilityBitLen = uint8(s.BitLen())
		if len(b) != 0 {
			f.capabilityBytes[0] = b[0]
		}
	}
	if m.Body != nil && m.Body.RequestCapabilities != nil && m.Body.RequestCapabilities.ECID != nil {
		f.capabilityPresent = true
		f.capabilityBitLen = 0
	}
	if m.Body != nil && m.Body.RequestLocationInformation != nil && m.Body.RequestLocationInformation.ECID != nil {
		s := m.Body.RequestLocationInformation.ECID.RequestedMeasurements
		b := s.Bytes()
		f.locationPresent = true
		f.locationBitLen = uint8(s.BitLen())
		if len(b) != 0 {
			f.locationBytes[0] = b[0]
		}
	}
	if encoded, err := lpp.EncodeMessage(m); err == nil && len(encoded.Bytes) <= len(f.encoded) {
		f.encodedLen = uint16(len(encoded.Bytes))
		copy(f.encoded[:], encoded.Bytes)
	}
	return f
}
func terminalError(s State) error {
	switch s {
	case Completed:
		return ErrCompleted
	case Aborted:
		return ErrAborted
	default:
		return ErrFailed
	}
}
