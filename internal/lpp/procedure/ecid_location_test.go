package procedure

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/location"
	"github.com/vectorcore/esmlc/internal/lpp/transaction"
	"github.com/vectorcore/esmlc/internal/uper"
)

func locationFixture(t *testing.T, name string, bitLength int) location.RequestLocationInformationR9IEs {
	t.Helper()
	root, err := filepath.Abs("../../../tools/specs/lpp/fixtures/r16.4.0/ecid-location/valid")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, name+".uper"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := uper.NewReader(data, bitLength)
	if err != nil {
		t.Fatal(err)
	}
	v, err := location.DecodeRequestLocationInformation(r)
	if err != nil || r.ValidateFinalPadding() != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return v
}

func TestStartECIDLocationInformationMatchesFixtures(t *testing.T) {
	for _, tc := range []struct {
		name string
		bits int
		hex  []byte
	}{
		{"request-ecid-rsrp", 14, []byte{0x01, 0x04}},
		{"request-ecid-all-root", 16, []byte{0x01, 0x17}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload := locationFixture(t, tc.name, tc.bits)
			o := newOrch(t)
			r, err := o.StartLocationInformation(StartLocationInformationOptions{ECID: payload.ECID}, time.Time{})
			if err != nil || len(r.Actions) != 1 || len(r.Events) != 1 {
				t.Fatalf("start %#v: %v", r, err)
			}
			m := r.Actions[0].Message
			if m.Body == nil || m.Body.Kind != lpp.BodyRequestLocationInformation || m.Body.RequestLocationInformation == nil || m.Body.RequestLocationInformation.ECID == nil {
				t.Fatal("missing typed request")
			}
			if !m.Body.RequestLocationInformation.ECID.RequestedMeasurements.Equal(payload.ECID.RequestedMeasurements) {
				t.Fatal("request bitmap changed")
			}
			w := uper.NewWriter()
			if err := location.EncodeRequestLocationInformation(w, *m.Body.RequestLocationInformation); err != nil {
				t.Fatal(err)
			}
			encoded := w.Encoded()
			if string(encoded.Bytes) != string(tc.hex) || encoded.BitLength != tc.bits {
				t.Fatalf("got %x/%d", encoded.Bytes, encoded.BitLength)
			}
			if r.Events[0].LocationRequest == nil || !r.Events[0].LocationRequest.ECID.RequestedMeasurements.Equal(payload.ECID.RequestedMeasurements) {
				t.Fatal("event bitmap changed")
			}
		})
	}
}

func TestInboundECIDLocationRequestPendingAndDuplicate(t *testing.T) {
	o := newOrch(t)
	payload := locationFixture(t, "request-ecid-all-root", 16)
	m := request(50, lpp.BodyRequestLocationInformation)
	m.Body.RequestLocationInformation = &payload
	r, err := o.HandleInbound(m, time.Time{})
	if err != nil || len(r.Events) != 2 || r.Events[0].Kind != LocationInformationRequested || r.Events[1].Kind != AwaitingApplicationResult {
		t.Fatalf("inbound %#v: %v", r, err)
	}
	if r.Events[0].LocationRequest == nil || r.Events[0].LocationRequest.ECID.RequestedMeasurements.BitLen() != 3 || !r.Events[0].LocationRequest.ECID.RequestsRSRQ() || !r.Events[0].LocationRequest.ECID.RequestsUERxTxTimeDifference() {
		t.Fatal("missing typed event")
	}
	k := transaction.Key{Initiator: lpp.InitiatorTargetDevice, Number: 50}
	s, ok := o.Snapshot(k)
	if !ok || s.Waiting != WaitLocationInformation || s.LocationRequest == nil || !s.LocationRequest.ECID.RequestedMeasurements.Equal(payload.ECID.RequestedMeasurements) {
		t.Fatalf("bad snapshot %#v", s)
	}
	// The event owns a clone; changing it cannot alter the retained request.
	r.Events[0].LocationRequest.ECID = nil
	s, _ = o.Snapshot(k)
	if s.LocationRequest == nil || s.LocationRequest.ECID == nil {
		t.Fatal("snapshot aliased event")
	}
	r, err = o.HandleInbound(m, time.Time{})
	if err != nil || len(r.Events) != 1 || r.Events[0].Kind != DuplicateIgnored {
		t.Fatalf("duplicate %#v: %v", r, err)
	}
	changed := locationFixture(t, "request-ecid-rsrp", 14)
	m.Body.RequestLocationInformation = &changed
	if _, err = o.HandleInbound(m, time.Time{}); !errors.Is(err, ErrApplicationPending) {
		t.Fatalf("changed request accepted as duplicate: %v", err)
	}
	sameLeading, err := uper.NewBitString([]byte{0x80}, 3)
	if err != nil {
		t.Fatal(err)
	}
	m.Body.RequestLocationInformation = &location.RequestLocationInformationR9IEs{ECID: &location.ECIDRequestLocationInformation{RequestedMeasurements: sameLeading}}
	if _, err = o.HandleInbound(m, time.Time{}); !errors.Is(err, ErrApplicationPending) {
		t.Fatalf("changed bit length accepted as duplicate: %v", err)
	}
}

func TestInboundECIDLocationRequestAcknowledgementAndTerminalCleanup(t *testing.T) {
	o := newOrch(t)
	payload := locationFixture(t, "request-ecid-rsrp", 14)
	m := request(51, lpp.BodyRequestLocationInformation)
	m.Body.RequestLocationInformation = &payload
	seq := uint8(7)
	m.SequenceNumber = &seq
	m.Acknowledgement = &lpp.Acknowledgement{Requested: true}
	r, err := o.HandleInbound(m, time.Time{})
	if err != nil || len(r.Actions) != 1 || r.Actions[0].Kind != SendAcknowledgement || len(r.Events) != 3 || r.Events[0].Kind != AcknowledgementRequested || r.Events[1].Kind != LocationInformationRequested || r.Events[2].Kind != AwaitingApplicationResult {
		t.Fatalf("ack request %#v: %v", r, err)
	}
	k := transaction.Key{Initiator: lpp.InitiatorTargetDevice, Number: 51}
	if _, err = o.Abort(k, time.Time{}); err != nil {
		t.Fatal(err)
	}
	s, ok := o.Snapshot(k)
	if !ok || s.Waiting != WaitNone || s.LocationRequest != nil {
		t.Fatalf("abort did not clear pending request %#v", s)
	}
	// No ProvideLocationInformation API exists in the bounded request-side
	// procedure layer, so cleared work cannot be completed with a placeholder.
}

func TestInvalidOutboundECIDLocationRequest(t *testing.T) {
	o := newOrch(t)
	if _, err := o.StartLocationInformation(StartLocationInformationOptions{ECID: &location.ECIDRequestLocationInformation{}}, time.Time{}); !errors.Is(err, ErrInvalidLocationRequest) {
		t.Fatal(err)
	}
}

func TestECIDLocationPendingClearsOnErrorAndPrune(t *testing.T) {
	o := newOrch(t)
	payload := locationFixture(t, "request-ecid-rsrp", 14)
	m := request(52, lpp.BodyRequestLocationInformation)
	m.Body.RequestLocationInformation = &payload
	if _, err := o.HandleInbound(m, time.Time{}); err != nil {
		t.Fatal(err)
	}
	k := transaction.Key{Initiator: lpp.InitiatorTargetDevice, Number: 52}
	if _, err := o.ReportError(k, time.Time{}); err != nil {
		t.Fatal(err)
	}
	s, ok := o.Snapshot(k)
	if !ok || s.Waiting != WaitNone || s.LocationRequest != nil {
		t.Fatal("error did not clear pending location work")
	}
	m = request(53, lpp.BodyRequestLocationInformation)
	m.Body.RequestLocationInformation = &payload
	if _, err := o.HandleInbound(m, time.Time{}); err != nil {
		t.Fatal(err)
	}
	o.Prune(time.Now().Add(24 * time.Hour))
	k = transaction.Key{Initiator: lpp.InitiatorTargetDevice, Number: 53}
	if _, ok := o.Snapshot(k); ok {
		t.Fatal("prune retained location pending state")
	}
}
