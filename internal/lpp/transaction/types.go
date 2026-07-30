package transaction

import (
	"fmt"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp"
)

// Key is unique only within a caller-scoped Store.
type Key struct {
	Initiator lpp.Initiator
	Number    uint8
}

func (k Key) Validate() error {
	if k.Initiator != lpp.InitiatorLocationServer && k.Initiator != lpp.InitiatorTargetDevice {
		return fmt.Errorf("%w: initiator", ErrInvalidKey)
	}
	return nil
}
func (k Key) String() string { return fmt.Sprintf("%d/%d", k.Initiator, k.Number) }

type Ownership uint8

const (
	LocallyInitiated Ownership = iota
	RemotelyInitiated
)

func (v Ownership) valid() bool { return v == LocallyInitiated || v == RemotelyInitiated }

type Direction uint8

const (
	Inbound Direction = iota
	Outbound
)

func (v Direction) valid() bool { return v == Inbound || v == Outbound }

type State uint8

const (
	Created State = iota
	Active
	AwaitingAcknowledgement
	Completed
	Aborted
	Failed
)

func (s State) terminal() bool { return s == Completed || s == Aborted || s == Failed }

type Procedure uint8

const (
	EnvelopeOnly Procedure = iota
	Capabilities
	LocationInformation
	AbortProcedure
	ErrorProcedure
)

type Classification uint8

const (
	New Classification = iota
	Duplicate
	Conflict
	Stale
	ReplayAfterCompletion
	Transactionless
)

type Limits struct {
	MaxActive         int
	MaxRetained       int
	ActiveLifetime    time.Duration
	RetentionLifetime time.Duration
}

func DefaultLimits() Limits {
	return Limits{MaxActive: 128, MaxRetained: 128, ActiveLifetime: 30 * time.Minute, RetentionLifetime: 10 * time.Minute}
}

type Snapshot struct {
	Key                       Key
	Ownership                 Ownership
	State                     State
	Procedure                 Procedure
	CreatedAt, LastActivity   time.Time
	LastInbound, LastOutbound *uint8
	AcknowledgementPending    bool
}
type Result struct {
	HasKey                     bool
	Key                        Key
	Classification             Classification
	PreviousState              State
	State                      State
	AcknowledgementMatched     bool
	AcknowledgementPending     bool
	Completed, Aborted, Failed bool
	Transactionless            bool
}
type PruneResult struct{ ExpiredActive, RemovedRetained int }
