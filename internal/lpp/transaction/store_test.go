package transaction

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp"
)

func TestFixtureMessagesApplyAccordingToTransactionlessPolicy(t *testing.T) {
	root, err := filepath.Abs("../../../tools/specs/lpp/fixtures/r16.4.0")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Fixtures []struct {
			Name       string `json:"name"`
			BinaryFile string `json:"binary_file"`
			BitLength  int    `json:"bit_length"`
		} `json:"fixtures"`
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Fixtures) != 11 {
		t.Fatalf("fixtures: %d", len(manifest.Fixtures))
	}
	for _, f := range manifest.Fixtures {
		t.Run(f.Name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, f.BinaryFile))
			if err != nil {
				t.Fatal(err)
			}
			m, err := lpp.DecodeMessage(data, f.BitLength)
			if err != nil {
				t.Fatal(err)
			}
			s, _ := NewStore(DefaultLimits())
			_, err = s.Apply(Inbound, RemotelyInitiated, m, time.Time{})
			if m.TransactionID == nil && m.Body == nil {
				if err != nil {
					t.Fatalf("transactionless envelope: %v", err)
				}
			} else if err == nil {
				t.Fatal("fixture unexpectedly accepted as a new remotely initiated procedure")
			}
		})
	}
}

func msg(n uint8, kind lpp.BodyKind) lpp.Message {
	return lpp.Message{TransactionID: &lpp.TransactionID{Initiator: lpp.InitiatorLocationServer, TransactionNumber: n}, Body: &lpp.Body{Kind: kind}}
}
func TestProcedureDuplicateAndTerminal(t *testing.T) {
	s, _ := NewStore(DefaultLimits())
	now := time.Unix(1, 0)
	req := msg(4, lpp.BodyRequestCapabilities)
	r, err := s.Apply(Inbound, RemotelyInitiated, req, now)
	if err != nil || r.State != Active {
		t.Fatalf("request: %#v %v", r, err)
	}
	r, err = s.Apply(Inbound, RemotelyInitiated, req, now)
	if err != nil || r.Classification != Duplicate {
		t.Fatalf("duplicate: %#v %v", r, err)
	}
	provide := msg(4, lpp.BodyProvideCapabilities)
	r, err = s.Apply(Outbound, RemotelyInitiated, provide, now)
	if err != nil || r.State != Active {
		t.Fatalf("provide: %#v %v", r, err)
	}
	end := msg(4, lpp.BodyProvideCapabilities)
	end.EndTransaction = true
	r, err = s.Apply(Outbound, RemotelyInitiated, end, now)
	if err != nil || !r.Completed {
		t.Fatalf("end: %#v %v", r, err)
	}
	_, err = s.Apply(Inbound, RemotelyInitiated, req, now)
	if !errors.Is(err, ErrCompleted) {
		t.Fatalf("post completion %v", err)
	}
}
func TestSequenceAckAndAbort(t *testing.T) {
	s, _ := NewStore(DefaultLimits())
	now := time.Unix(1, 0)
	req := msg(9, lpp.BodyRequestCapabilities)
	z := uint8(0)
	req.SequenceNumber = &z
	req.Acknowledgement = &lpp.Acknowledgement{Requested: true}
	r, err := s.Apply(Inbound, RemotelyInitiated, req, now)
	if err != nil || !r.AcknowledgementPending {
		t.Fatalf("request %v %#v", err, r)
	}
	ack := msg(9, lpp.BodyProvideCapabilities)
	ack.SequenceNumber = &z
	i := uint8(0)
	ack.Acknowledgement = &lpp.Acknowledgement{Indicator: &i}
	r, err = s.Apply(Outbound, RemotelyInitiated, ack, now)
	if err != nil || !r.AcknowledgementMatched {
		t.Fatalf("ack %v %#v", err, r)
	}
	abort := msg(9, lpp.BodyAbort)
	r, err = s.Apply(Inbound, RemotelyInitiated, abort, now)
	if err != nil || !r.Aborted {
		t.Fatalf("abort %v %#v", err, r)
	}
	r, err = s.Apply(Inbound, RemotelyInitiated, abort, now)
	if err != nil || r.Classification != Duplicate {
		t.Fatalf("duplicate abort %v %#v", err, r)
	}
}
func TestLocalAllocatorAndLimits(t *testing.T) {
	l := DefaultLimits()
	l.MaxActive = 2
	s, _ := NewStore(l)
	now := time.Time{}
	a, e := s.CreateLocal(lpp.InitiatorLocationServer, Capabilities, now)
	if e != nil {
		t.Fatal(e)
	}
	b, e := s.CreateLocal(lpp.InitiatorLocationServer, Capabilities, now)
	if e != nil || a.Key == b.Key {
		t.Fatalf("allocation %v %#v %#v", e, a, b)
	}
	if _, e = s.CreateLocal(lpp.InitiatorLocationServer, Capabilities, now); !errors.Is(e, ErrCapacity) {
		t.Fatalf("capacity %v", e)
	}
}

func TestAllocatorExhaustsAllNumbers(t *testing.T) {
	l := DefaultLimits()
	l.MaxActive = 256
	s, _ := NewStore(l)
	seen := map[uint8]bool{}
	for i := 0; i < 256; i++ {
		x, err := s.CreateLocal(lpp.InitiatorLocationServer, Capabilities, time.Time{})
		if err != nil || seen[x.Key.Number] {
			t.Fatalf("allocation %d: %#v %v", i, x, err)
		}
		seen[x.Key.Number] = true
	}
	if _, err := s.CreateLocal(lpp.InitiatorLocationServer, Capabilities, time.Time{}); !errors.Is(err, ErrCapacity) {
		t.Fatalf("exhaustion: %v", err)
	}
}

func TestSequenceWrap(t *testing.T) {
	s, _ := NewStore(DefaultLimits())
	m := msg(12, lpp.BodyRequestCapabilities)
	a := uint8(255)
	m.SequenceNumber = &a
	if _, err := s.Apply(Inbound, RemotelyInitiated, m, time.Time{}); err != nil {
		t.Fatal(err)
	}
	b := uint8(0)
	m.SequenceNumber = &b
	if _, err := s.Apply(Inbound, RemotelyInitiated, m, time.Time{}); err != nil {
		t.Fatalf("255 -> 0: %v", err)
	}
}
func TestTransactionlessAndPrune(t *testing.T) {
	l := DefaultLimits()
	l.ActiveLifetime = time.Second
	l.RetentionLifetime = time.Second
	s, _ := NewStore(l)
	now := time.Unix(1, 0)
	r, e := s.Apply(Inbound, RemotelyInitiated, lpp.Message{}, now)
	if e != nil || !r.Transactionless {
		t.Fatalf("transactionless %v %#v", e, r)
	}
	if _, e = s.Apply(Inbound, RemotelyInitiated, lpp.Message{Body: &lpp.Body{Kind: lpp.BodyRequestCapabilities}}, now); e == nil {
		t.Fatal("body without key accepted")
	}
	_, e = s.Apply(Inbound, RemotelyInitiated, msg(2, lpp.BodyRequestCapabilities), now)
	if e != nil {
		t.Fatal(e)
	}
	p := s.Prune(now.Add(2 * time.Second))
	if p.ExpiredActive != 1 {
		t.Fatalf("prune %#v", p)
	}
}
func TestConcurrentCreate(t *testing.T) {
	l := DefaultLimits()
	l.MaxActive = 256
	s, _ := NewStore(l)
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, _ = s.CreateLocal(lpp.InitiatorLocationServer, Capabilities, time.Time{}) }()
	}
	wg.Wait()
	seen := map[Key]bool{}
	s.mu.Lock()
	for k := range s.records {
		if seen[k] {
			t.Fatal("duplicate key")
		}
		seen[k] = true
	}
	s.mu.Unlock()
}
