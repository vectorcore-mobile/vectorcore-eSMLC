package lpp

import "testing"

func FuzzDecodeMessage(f *testing.F) {
	f.Add([]byte{0}, uint8(5))
	f.Fuzz(func(t *testing.T, data []byte, bits uint8) {
		n := int(bits)
		if n > len(data)*8 {
			n = len(data) * 8
		}
		_, _ = DecodeMessage(data, n)
	})
}
func FuzzMessageRoundTrip(f *testing.F) {
	f.Add([]byte{0}, uint8(5))
	f.Fuzz(func(t *testing.T, data []byte, bits uint8) {
		n := int(bits)
		if n > len(data)*8 {
			n = len(data) * 8
		}
		m, e := DecodeMessage(data, n)
		if e != nil {
			return
		}
		x, e := EncodeMessage(m)
		if e != nil {
			t.Fatal(e)
		}
		again, e := DecodeMessage(x.Bytes, x.BitLength)
		if e != nil || !equalMessage(m, again) {
			t.Fatal(e)
		}
	})
}

func FuzzDecodeMessageOctets(f *testing.F) {
	f.Add([]byte{0x11, 0x00, 0x00})
	f.Add([]byte{0xff})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1024 {
			return
		}
		_, _ = DecodeMessageOctets(data)
	})
}
