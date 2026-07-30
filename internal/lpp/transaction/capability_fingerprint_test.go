package transaction

import (
	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/capability"
	"github.com/vectorcore/esmlc/internal/uper"
	"testing"
)

func TestCapabilityPayloadContributesToFingerprint(t *testing.T) {
	makeMessage := func(raw byte, n int) lpp.Message {
		v, _ := uper.NewBitString([]byte{raw}, n)
		return lpp.Message{Body: &lpp.Body{Kind: lpp.BodyProvideCapabilities, ProvideCapabilities: &capability.ProvideCapabilitiesR9IEs{ECID: &capability.ECIDProvideCapabilities{MeasurementSupport: v}}}}
	}
	a := makeFingerprint(makeMessage(0x80, 1))
	b := makeFingerprint(makeMessage(0x80, 3))
	c := makeFingerprint(makeMessage(0xc0, 2))
	if a == b || a == c {
		t.Fatal("capability differences collapsed")
	}
}
