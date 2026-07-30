package transaction

import (
	"github.com/vectorcore/esmlc/internal/lpp"
	"testing"
	"time"
)

func FuzzApplyMessage(f *testing.F) {
	f.Add(uint8(0), uint8(0), false)
	f.Fuzz(func(t *testing.T, n, seq uint8, end bool) {
		s, _ := NewStore(DefaultLimits())
		m := lpp.Message{TransactionID: &lpp.TransactionID{Initiator: lpp.InitiatorLocationServer, TransactionNumber: n}, EndTransaction: end, Body: &lpp.Body{Kind: lpp.BodyRequestCapabilities}}
		if seq&1 == 1 {
			x := seq
			m.SequenceNumber = &x
		}
		_, _ = s.Apply(Inbound, RemotelyInitiated, m, time.Time{})
	})
}
func FuzzSequenceProgression(f *testing.F) {
	f.Add(uint8(0), uint8(1))
	f.Fuzz(func(t *testing.T, a, b uint8) {
		s, _ := NewStore(DefaultLimits())
		m := lpp.Message{TransactionID: &lpp.TransactionID{Initiator: lpp.InitiatorLocationServer}, Body: &lpp.Body{Kind: lpp.BodyRequestCapabilities}}
		m.SequenceNumber = &a
		_, _ = s.Apply(Inbound, RemotelyInitiated, m, time.Time{})
		m.SequenceNumber = &b
		_, _ = s.Apply(Inbound, RemotelyInitiated, m, time.Time{})
	})
}
func FuzzStoreOperations(f *testing.F) {
	f.Add(uint8(3))
	f.Fuzz(func(t *testing.T, n uint8) {
		l := DefaultLimits()
		l.MaxActive = 4
		s, _ := NewStore(l)
		for i := 0; i < int(n%8); i++ {
			_, _ = s.CreateLocal(lpp.InitiatorLocationServer, Capabilities, time.Time{})
			_ = s.Prune(time.Time{})
		}
	})
}
