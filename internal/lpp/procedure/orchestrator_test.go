package procedure

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/transaction"
)

func TestFixtureEnvelopesRemainBoundedAtProcedureBoundary(t *testing.T) {
	root, err := filepath.Abs("../../../tools/specs/lpp/fixtures/r16.4.0")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Fixtures []struct {
			Name   string `json:"name"`
			Binary string `json:"binary_file"`
			Bits   int    `json:"bit_length"`
		} `json:"fixtures"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if len(m.Fixtures) != 11 {
		t.Fatalf("fixtures %d", len(m.Fixtures))
	}
	for _, f := range m.Fixtures {
		t.Run(f.Name, func(t *testing.T) {
			data, e := os.ReadFile(filepath.Join(root, f.Binary))
			if e != nil {
				t.Fatal(e)
			}
			msg, e := lpp.DecodeMessage(data, f.Bits)
			if e != nil {
				t.Fatal(e)
			}
			_, _ = newOrch(t).HandleInbound(msg, time.Time{})
		})
	}
}

func newOrch(t *testing.T) *Orchestrator {
	t.Helper()
	s, e := transaction.NewStore(transaction.DefaultLimits())
	if e != nil {
		t.Fatal(e)
	}
	o, e := New(s, DefaultConfig())
	if e != nil {
		t.Fatal(e)
	}
	return o
}
func request(n uint8, k lpp.BodyKind) lpp.Message {
	return lpp.Message{TransactionID: &lpp.TransactionID{Initiator: lpp.InitiatorTargetDevice, TransactionNumber: n}, Body: &lpp.Body{Kind: k}}
}
func has(events []Event, k EventKind) bool {
	for _, e := range events {
		if e.Kind == k {
			return true
		}
	}
	return false
}
func TestCapabilitiesInboundProvideAndDuplicate(t *testing.T) {
	o := newOrch(t)
	now := time.Time{}
	r, e := o.HandleInbound(request(1, lpp.BodyRequestCapabilities), now)
	if e != nil || !has(r.Events, CapabilitiesRequested) || !has(r.Events, AwaitingApplicationResult) {
		t.Fatalf("request %#v %v", r, e)
	}
	r, e = o.HandleInbound(request(1, lpp.BodyRequestCapabilities), now)
	if e != nil || !has(r.Events, DuplicateIgnored) {
		t.Fatalf("duplicate %#v %v", r, e)
	}
	k := transaction.Key{Initiator: lpp.InitiatorTargetDevice, Number: 1}
	r, e = o.ProvideCapabilities(k, now)
	if e != nil || len(r.Actions) != 1 || !has(r.Events, ProcedureCompleted) {
		t.Fatalf("provide %#v %v", r, e)
	}
	if _, e = o.ProvideCapabilities(k, now); !errors.Is(e, ErrApplicationNotPending) {
		t.Fatalf("second result %v", e)
	}
}
func TestLocationStartAndTypedProvide(t *testing.T) {
	o := newOrch(t)
	r, e := o.StartLocationInformation(StartLocationInformationOptions{}, time.Time{})
	if e != nil || len(r.Actions) != 1 {
		t.Fatalf("start %#v %v", r, e)
	}
	k := *r.Actions[0].Key
	data, _ := hex.DecodeString("920f2809007c0200100848d159c0192aac317ff8")
	m, e := lpp.DecodeMessage(data, 157)
	if e != nil {
		t.Fatal(e)
	}
	m.TransactionID = &lpp.TransactionID{Initiator: k.Initiator, TransactionNumber: k.Number}
	r, e = o.HandleInbound(m, time.Time{})
	if e != nil || !has(r.Events, LocationInformationEnvelopeProvided) || !has(r.Events, ProcedureCompleted) {
		t.Fatalf("provide %#v %v", r, e)
	}
	r, e = o.HandleInbound(m, time.Time{})
	if e != nil || !has(r.Events, DuplicateIgnored) {
		t.Fatalf("duplicate %#v %v", r, e)
	}
}

func TestInboundLocationRequestCanBeCompletedWithTypedProvide(t *testing.T) {
	o := newOrch(t)
	now := time.Time{}
	r, err := o.HandleInbound(request(17, lpp.BodyRequestLocationInformation), now)
	if err != nil || !has(r.Events, LocationInformationRequested) || !has(r.Events, AwaitingApplicationResult) {
		t.Fatalf("request %#v %v", r, err)
	}
	data, _ := hex.DecodeString("920f2809007c0200100848d159c0192aac317ff8")
	m, err := lpp.DecodeMessage(data, 157)
	if err != nil || m.Body == nil || m.Body.ProvideLocationInformation == nil {
		t.Fatal(err)
	}
	k := transaction.Key{Initiator: lpp.InitiatorTargetDevice, Number: 17}
	r, err = o.ProvideLocationInformation(k, ProvideLocationInformationOptions{LocationInformation: *m.Body.ProvideLocationInformation}, now)
	if err != nil || len(r.Actions) != 1 || r.Actions[0].Kind != SendProcedureResponse || !has(r.Events, LocationInformationEnvelopeProvided) {
		t.Fatalf("provide %#v %v", r, err)
	}
	if _, err = o.ProvideLocationInformation(k, ProvideLocationInformationOptions{}, now); !errors.Is(err, ErrApplicationNotPending) {
		t.Fatal(err)
	}
}
func TestInboundAcknowledgementAction(t *testing.T) {
	o := newOrch(t)
	m := request(3, lpp.BodyRequestCapabilities)
	m.Acknowledgement = &lpp.Acknowledgement{Requested: true}
	x := uint8(9)
	m.SequenceNumber = &x
	r, e := o.HandleInbound(m, time.Time{})
	if e != nil || len(r.Actions) != 1 || r.Actions[0].Kind != SendAcknowledgement {
		t.Fatalf("ack %#v %v", r, e)
	}
	a := r.Actions[0].Message
	if a.Acknowledgement == nil || a.Acknowledgement.Indicator == nil || a.SequenceNumber == nil || *a.SequenceNumber != 9 {
		t.Fatalf("bad ack %#v", a)
	}
}
func TestAbortErrorAndPrune(t *testing.T) {
	o := newOrch(t)
	_, _ = o.HandleInbound(request(4, lpp.BodyRequestLocationInformation), time.Time{})
	k := transaction.Key{Initiator: lpp.InitiatorTargetDevice, Number: 4}
	r, e := o.Abort(k, time.Time{})
	if e != nil || !has(r.Events, ProcedureAborted) {
		t.Fatalf("abort %#v %v", r, e)
	}
	if _, e = o.ReportError(k, time.Time{}); e == nil {
		t.Fatal("error after abort accepted")
	}
	// A short store lifetime is separately tested in transaction; pruning here
	// must at least be explicit and safe.
	_ = o.Prune(time.Now().Add(24 * time.Hour))
}
func TestConcurrentStarts(t *testing.T) {
	o := newOrch(t)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, _ = o.StartCapabilities(StartOptions{}, time.Time{}) }()
	}
	wg.Wait()
}
