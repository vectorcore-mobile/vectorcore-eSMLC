package capability

import (
	"github.com/vectorcore/esmlc/internal/uper"
	"testing"
)

func FuzzDecodeCapabilities(f *testing.F) {
	f.Add([]byte{0, 0}, uint8(9))
	f.Fuzz(func(t *testing.T, b []byte, n uint8) {
		bits := int(n)
		if bits > len(b)*8 {
			bits = len(b) * 8
		}
		r, e := uper.NewReader(b, bits)
		if e == nil {
			_, _ = DecodeProvideCapabilities(r)
		}
	})
}
func FuzzECIDProvideRoundTrip(f *testing.F) {
	f.Add(uint8(0x80), uint8(1))
	f.Fuzz(func(t *testing.T, b, n uint8) {
		nbits := int(n%8) + 1
		raw := []byte{b & byte(0xff<<(8-nbits))}
		s, e := uper.NewBitString(raw, nbits)
		if e != nil {
			t.Fatal(e)
		}
		v := ProvideCapabilitiesR9IEs{ECID: &ECIDProvideCapabilities{MeasurementSupport: s}}
		w := uper.NewWriter()
		if EncodeProvideCapabilities(w, v) != nil {
			t.Fatal("encode")
		}
		x := w.Encoded()
		r, _ := uper.NewReader(x.Bytes, x.BitLength)
		got, e := DecodeProvideCapabilities(r)
		if e != nil || got.ECID == nil || !got.ECID.MeasurementSupport.Equal(s) {
			t.Fatalf("%v", e)
		}
	})
}
func FuzzCapabilityMutation(f *testing.F) {
	f.Add([]byte{1, 4}, uint8(14))
	f.Fuzz(func(t *testing.T, b []byte, n uint8) {
		bits := int(n)
		if bits > len(b)*8 {
			bits = len(b) * 8
		}
		r, e := uper.NewReader(b, bits)
		if e == nil {
			_, _ = DecodeProvideCapabilities(r)
		}
	})
}
