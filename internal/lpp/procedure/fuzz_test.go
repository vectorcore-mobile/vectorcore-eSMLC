package procedure

import (
	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/capability"
	"github.com/vectorcore/esmlc/internal/lpp/location"
	"github.com/vectorcore/esmlc/internal/lpp/transaction"
	"github.com/vectorcore/esmlc/internal/uper"
	"testing"
	"time"
)

func FuzzHandleInbound(f *testing.F) {
	f.Add(uint8(1), uint8(0))
	f.Fuzz(func(t *testing.T, n, b uint8) {
		o := newOrch(t)
		kinds := []lpp.BodyKind{lpp.BodyRequestCapabilities, lpp.BodyRequestLocationInformation, lpp.BodyAbort, lpp.BodyError}
		m := request(n, kinds[int(b)%len(kinds)])
		_, _ = o.HandleInbound(m, time.Time{})
	})
}

func FuzzCapabilityProcedureInput(f *testing.F) {
	f.Add(uint8(1))
	f.Fuzz(func(t *testing.T, n uint8) {
		o := newOrch(t)
		m := request(n, lpp.BodyRequestCapabilities)
		if n&1 == 1 {
			m.Body.RequestCapabilities = &capability.RequestCapabilitiesR9IEs{ECID: &capability.ECIDRequestCapabilities{}}
		}
		_, _ = o.HandleInbound(m, time.Time{})
	})
}
func FuzzCapabilityResultOperations(f *testing.F) {
	f.Add(uint8(0x80), uint8(1))
	f.Fuzz(func(t *testing.T, b, n uint8) {
		o := newOrch(t)
		m := request(1, lpp.BodyRequestCapabilities)
		m.Body.RequestCapabilities = &capability.RequestCapabilitiesR9IEs{ECID: &capability.ECIDRequestCapabilities{}}
		_, _ = o.HandleInbound(m, time.Time{})
		bits := int(n%8) + 1
		v, e := uper.NewBitString([]byte{b & byte(0xff<<(8-bits))}, bits)
		if e == nil {
			_, _ = o.ProvideCapabilitiesResult(transaction.Key{Initiator: lpp.InitiatorTargetDevice, Number: 1}, ProvideCapabilitiesOptions{Capabilities: capability.ProvideCapabilitiesR9IEs{ECID: &capability.ECIDProvideCapabilities{MeasurementSupport: v}}}, time.Time{})
		}
	})
}
func FuzzOperationSequence(f *testing.F) {
	f.Add(uint8(3))
	f.Fuzz(func(t *testing.T, n uint8) {
		o := newOrch(t)
		for i := 0; i < int(n%8); i++ {
			r, _ := o.StartCapabilities(StartOptions{}, time.Time{})
			if len(r.Actions) > 0 {
				_, _ = o.Abort(*r.Actions[0].Key, time.Time{})
			}
		}
	})
}
func FuzzEventDeterminism(f *testing.F) {
	f.Add(uint8(4))
	f.Fuzz(func(t *testing.T, n uint8) {
		m := request(n, lpp.BodyRequestCapabilities)
		a := newOrch(t)
		b := newOrch(t)
		ra, ea := a.HandleInbound(m, time.Time{})
		rb, eb := b.HandleInbound(m, time.Time{})
		if (ea == nil) != (eb == nil) || len(ra.Actions) != len(rb.Actions) || len(ra.Events) != len(rb.Events) {
			t.Fatal("non deterministic")
		}
	})
}

func FuzzRequestLocationProcedureInput(f *testing.F) {
	f.Add(uint8(1), byte(0x80), byte(1))
	f.Add(uint8(2), byte(0xe0), byte(3))
	f.Fuzz(func(t *testing.T, n uint8, b, length byte) {
		bits := int(length%8) + 1
		v, err := uper.NewBitString([]byte{b & byte(0xff<<(8-bits))}, bits)
		if err != nil {
			return
		}
		o := newOrch(t)
		m := request(n, lpp.BodyRequestLocationInformation)
		m.Body.RequestLocationInformation = &location.RequestLocationInformationR9IEs{ECID: &location.ECIDRequestLocationInformation{RequestedMeasurements: v}}
		r, err := o.HandleInbound(m, time.Time{})
		if err != nil {
			return
		}
		if len(r.Events) < 2 || r.Events[0].LocationRequest == nil || !r.Events[0].LocationRequest.ECID.RequestedMeasurements.Equal(v) {
			t.Fatal("location request lost")
		}
		duplicate, err := o.HandleInbound(m, time.Time{})
		if err != nil || len(duplicate.Events) != 1 || duplicate.Events[0].Kind != DuplicateIgnored {
			t.Fatal("duplicate changed")
		}
	})
}

func FuzzRequestLocationOperationSequences(f *testing.F) {
	f.Add(uint8(4))
	f.Fuzz(func(t *testing.T, operations uint8) {
		o := newOrch(t)
		for i := 0; i < int(operations%8); i++ {
			v, _ := uper.NewBitString([]byte{0x80}, 1)
			m := request(uint8(i), lpp.BodyRequestLocationInformation)
			m.Body.RequestLocationInformation = &location.RequestLocationInformationR9IEs{ECID: &location.ECIDRequestLocationInformation{RequestedMeasurements: v}}
			_, _ = o.HandleInbound(m, time.Time{})
			if operations&(1<<uint(i%8)) != 0 {
				_, _ = o.Abort(transaction.Key{Initiator: lpp.InitiatorTargetDevice, Number: uint8(i)}, time.Time{})
			}
		}
	})
}

func FuzzRequestLocationBitLengthPreservation(f *testing.F) {
	f.Add(byte(0x80), byte(1))
	f.Add(byte(0x80), byte(3))
	f.Fuzz(func(t *testing.T, b, length byte) {
		bits := int(length%8) + 1
		v, err := uper.NewBitString([]byte{b & byte(0xff<<(8-bits))}, bits)
		if err != nil {
			return
		}
		o := newOrch(t)
		r, err := o.StartLocationInformation(StartLocationInformationOptions{ECID: &location.ECIDRequestLocationInformation{RequestedMeasurements: v}}, time.Time{})
		if err != nil || r.Events[0].LocationRequest == nil || !r.Events[0].LocationRequest.ECID.RequestedMeasurements.Equal(v) {
			t.Fatal("outbound location request changed")
		}
	})
}
