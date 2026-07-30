package location

import (
	"testing"

	"github.com/vectorcore/esmlc/internal/uper"
)

func FuzzDecodeRequestLocation(f *testing.F) {
	f.Add([]byte{0x01, 0x04}, 14)
	f.Add([]byte{0x01, 0x17}, 16)
	f.Fuzz(func(t *testing.T, data []byte, bitLength int) {
		if bitLength < 0 || bitLength > len(data)*8 {
			return
		}
		r, err := uper.NewReader(data, bitLength)
		if err != nil {
			return
		}
		v, err := DecodeRequestLocationInformation(r)
		if err != nil || r.ValidateFinalPadding() != nil {
			return
		}
		w := uper.NewWriter()
		if EncodeRequestLocationInformation(w, v) == nil {
			_ = w.Encoded()
		}
	})
}

func FuzzECIDRequestRoundTrip(f *testing.F) {
	f.Add(byte(0x80), byte(1))
	f.Add(byte(0xe0), byte(3))
	f.Fuzz(func(t *testing.T, b byte, length byte) {
		n := int(length%8) + 1
		mask := byte(0xff << (8 - n))
		v, err := uper.NewBitString([]byte{b & mask}, n)
		if err != nil {
			t.Fatal(err)
		}
		w := uper.NewWriter()
		if err := EncodeRequestLocationInformation(w, RequestLocationInformationR9IEs{ECID: &ECIDRequestLocationInformation{RequestedMeasurements: v}}); err != nil {
			t.Fatal(err)
		}
		e := w.Encoded()
		r, err := uper.NewReader(e.Bytes, e.BitLength)
		if err != nil {
			t.Fatal(err)
		}
		out, err := DecodeRequestLocationInformation(r)
		if err != nil || out.ECID == nil || !out.ECID.RequestedMeasurements.Equal(v) {
			t.Fatalf("round trip: %v", err)
		}
	})
}
