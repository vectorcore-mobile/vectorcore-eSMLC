package transaction

import (
	"encoding/hex"
	"testing"

	"github.com/vectorcore/esmlc/internal/lpp"
	"github.com/vectorcore/esmlc/internal/lpp/location"
	"github.com/vectorcore/esmlc/internal/uper"
)

func locationMessage(t *testing.T, bits byte, length int) lpp.Message {
	t.Helper()
	v, err := uper.NewBitString([]byte{bits}, length)
	if err != nil {
		t.Fatal(err)
	}
	return lpp.Message{Body: &lpp.Body{Kind: lpp.BodyRequestLocationInformation, RequestLocationInformation: &location.RequestLocationInformationR9IEs{ECID: &location.ECIDRequestLocationInformation{RequestedMeasurements: v}}}}
}

func TestProvideLocationFingerprintCoversExactTypedPayload(t *testing.T) {
	decode := func(hexValue string, bits int) lpp.Message {
		t.Helper()
		data, err := hex.DecodeString(hexValue)
		if err != nil {
			t.Fatal(err)
		}
		r, err := uper.NewReader(data, bits)
		if err != nil {
			t.Fatal(err)
		}
		payload, err := location.DecodeProvideLocationInformation(r)
		if err != nil {
			t.Fatal(err)
		}
		return lpp.Message{Body: &lpp.Body{Kind: lpp.BodyProvideLocationInformation, ProvideLocationInformation: &payload}}
	}
	// These independent fixtures differ in list cardinality and optional result
	// content. Transaction duplicate detection must not collapse them.
	one := decode("0120020040190f00", 57)
	two := decode("0124000000004000803207f7ffff8bffc0", 130)
	if makeFingerprint(one) == makeFingerprint(two) {
		t.Fatal("distinct provide payloads collided")
	}
}

func TestLocationPayloadFingerprintPreservesBitLengthAndContent(t *testing.T) {
	one := locationMessage(t, 0x80, 1)
	three := locationMessage(t, 0xe0, 3)
	different := locationMessage(t, 0x40, 2)
	if makeFingerprint(one) == makeFingerprint(three) {
		t.Fatal("distinct ECID bit lengths collided")
	}
	if makeFingerprint(three) == makeFingerprint(different) {
		t.Fatal("distinct ECID bits collided")
	}
	if makeFingerprint(one) != makeFingerprint(one) {
		t.Fatal("identical ECID request fingerprint changed")
	}
}
