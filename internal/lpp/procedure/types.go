package procedure

import (
	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/capability"
	"github.com/vectorcore/esmlc/internal/lpp/location"
	"github.com/vectorcore/esmlc/internal/lpp/transaction"
	"sync"
)

type Config struct {
	LocalInitiator                lpp.Initiator
	DefaultRequestAcknowledgement bool
	MaxPendingApplications        int
	ResponseEndsTransaction       bool
}

func DefaultConfig() Config {
	return Config{LocalInitiator: lpp.InitiatorLocationServer, MaxPendingApplications: 64, ResponseEndsTransaction: true}
}

type ApplicationWait uint8

const (
	WaitNone ApplicationWait = iota
	WaitCapabilities
	WaitLocationInformation
)

type ActionKind uint8

const (
	SendInitialRequest ActionKind = iota
	SendProcedureResponse
	SendAcknowledgement
	SendAbort
	SendProtocolError
)

type Action struct {
	Kind    ActionKind
	Message lpp.Message
	Key     *transaction.Key
}
type EventKind uint8

const (
	CapabilitiesRequested EventKind = iota
	CapabilitiesEnvelopeProvided
	LocationInformationRequested
	LocationInformationEnvelopeProvided
	AcknowledgementRequested
	AcknowledgementReceived
	AwaitingApplicationResult
	ProcedureCompleted
	ProcedureAborted
	ProcedureFailed
	DuplicateIgnored
	ProtocolViolation
)

type Event struct {
	Kind                EventKind
	Key                 *transaction.Key
	Wait                ApplicationWait
	Snapshot            *transaction.Snapshot
	CapabilityRequest   *capability.RequestCapabilitiesR9IEs
	Capabilities        *capability.ProvideCapabilitiesR9IEs
	LocationRequest     *location.RequestLocationInformationR9IEs
	LocationInformation *location.ProvideLocationInformationR9IEs
}
type Result struct {
	Actions  []Action
	Events   []Event
	Snapshot *transaction.Snapshot
}
type StartOptions struct {
	RequestAcknowledgement *bool
	RequestECID            bool
	RequestOTDOA           bool
	RequestAGNSS           bool
}

// StartLocationInformationOptions carries the only supported typed location
// requests. A nil Common/AGNSS/OTDOA/ECID value intentionally retains the
// established empty R9 request envelope; no measurements are inferred by the
// procedure layer. Common is separate from AGNSS because TS 37.355 carries
// the actual requested-report-type signal (locationInformationType) in the
// method-agnostic common IEs, not in any method-specific branch.
type StartLocationInformationOptions struct {
	RequestAcknowledgement *bool
	Common                 *location.CommonRequestLocationInformation
	AGNSS                  *location.AGNSSRequestLocationInformation
	ECID                   *location.ECIDRequestLocationInformation
	OTDOA                  *location.OTDOARequestLocationInformation
}
type ProvideCapabilitiesOptions struct {
	Capabilities capability.ProvideCapabilitiesR9IEs
}

// ProvideLocationInformationOptions contains a typed ECID measurement report.
// It is protocol data only: this package neither selects a positioning method
// nor evaluates the reported measurements.
type ProvideLocationInformationOptions struct {
	LocationInformation location.ProvideLocationInformationR9IEs
}
type PruneResult struct {
	Transaction             transaction.PruneResult
	ClearedApplicationWaits int
}
type ProcedureSnapshot struct {
	Transaction     transaction.Snapshot
	Waiting         ApplicationWait
	RequestedECID   bool
	RequestedOTDOA  bool
	RequestedAGNSS  bool
	LocationRequest *location.RequestLocationInformationR9IEs
}
type Orchestrator struct {
	tx    *transaction.Store
	cfg   Config
	mu    sync.Mutex
	waits map[transaction.Key]pendingWait
}
type pendingWait struct {
	kind            ApplicationWait
	requestedECID   bool
	requestedOTDOA  bool
	requestedAGNSS  bool
	locationRequest *location.RequestLocationInformationR9IEs
}
