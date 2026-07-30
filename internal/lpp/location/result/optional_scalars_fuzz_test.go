package result

import (
	"github.com/vectorcore/esmlc/internal/uper"
	"testing"
)

func FuzzRSRPResultRoundTrip(f *testing.F) {
	f.Add(uint16(49))
	f.Fuzz(func(t *testing.T, n uint16) {
		v, e := NewRSRPResult(n)
		if n > 97 {
			if e == nil {
				t.Fatal()
			}
			return
		}
		w := uper.NewWriter()
		_ = v.EncodeUPER(w)
		r, _ := uper.NewReader(w.Encoded().Bytes, 7)
		g, e := DecodeRSRPResult(r)
		if e != nil || g != v {
			t.Fatal(e)
		}
	})
}
func FuzzRSRQResultRoundTrip(f *testing.F) {
	f.Add(uint16(17))
	f.Fuzz(func(t *testing.T, n uint16) {
		v, e := NewRSRQResult(n)
		if n > 34 {
			if e == nil {
				t.Fatal()
			}
			return
		}
		w := uper.NewWriter()
		_ = v.EncodeUPER(w)
		r, _ := uper.NewReader(w.Encoded().Bytes, 6)
		g, e := DecodeRSRQResult(r)
		if e != nil || g != v {
			t.Fatal(e)
		}
	})
}
func FuzzUERxTxTimeDiffRoundTrip(f *testing.F) {
	f.Add(uint32(2048))
	f.Fuzz(func(t *testing.T, n uint32) {
		v, e := NewUERxTxTimeDiff(n)
		if n > 4095 {
			if e == nil {
				t.Fatal()
			}
			return
		}
		w := uper.NewWriter()
		_ = v.EncodeUPER(w)
		r, _ := uper.NewReader(w.Encoded().Bytes, 12)
		g, e := DecodeUERxTxTimeDiff(r)
		if e != nil || g != v {
			t.Fatal(e)
		}
	})
}
func FuzzSystemFrameNumberRoundTrip(f *testing.F) {
	f.Add(uint16(682))
	f.Fuzz(func(t *testing.T, n uint16) {
		v, e := NewSystemFrameNumberFromUint16(n)
		if n > 1023 {
			if e == nil {
				t.Fatal()
			}
			return
		}
		w := uper.NewWriter()
		_ = v.EncodeUPER(w)
		r, _ := uper.NewReader(w.Encoded().Bytes, 10)
		g, e := DecodeSystemFrameNumber(r)
		if e != nil || g.Uint16() != n {
			t.Fatal(e)
		}
	})
}
func FuzzOptionalResultScalarDecode(f *testing.F) {
	f.Add([]byte{0xff, 0xff}, uint8(12))
	f.Fuzz(func(t *testing.T, b []byte, n uint8) {
		if len(b) > 4 {
			return
		}
		bits := int(n)
		if bits > len(b)*8 {
			bits = len(b) * 8
		}
		r, _ := uper.NewReader(b, bits)
		_, _ = DecodeRSRPResult(r)
		r, _ = uper.NewReader(b, bits)
		_, _ = DecodeRSRQResult(r)
		r, _ = uper.NewReader(b, bits)
		_, _ = DecodeUERxTxTimeDiff(r)
		r, _ = uper.NewReader(b, bits)
		_, _ = DecodeSystemFrameNumber(r)
	})
}
