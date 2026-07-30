package uper

import "testing"

func FuzzFixedBitStringRoundTrip(f *testing.F) {
	f.Add(uint8(10), []byte{0xaa, 0x80})
	f.Fuzz(func(t *testing.T, size uint8, raw []byte) {
		n := int(size%64) + 1
		bytes := (n + 7) / 8
		if len(raw) < bytes {
			return
		}
		raw = append([]byte(nil), raw[:bytes]...)
		if n%8 != 0 {
			raw[bytes-1] &= byte(0xff << (8 - n%8))
		}
		v, err := NewBitString(raw, n)
		if err != nil {
			return
		}
		w := NewWriter()
		if err = w.WriteBitString(v, n, n); err != nil {
			t.Fatal(err)
		}
		e := w.Encoded()
		if e.BitLength != n {
			t.Fatal("fixed size emitted a length determinant")
		}
		r, _ := NewReader(e.Bytes, e.BitLength)
		got, err := r.ReadBitString(n, n)
		if err != nil || !got.Equal(v) {
			t.Fatalf("round trip: %v", err)
		}
	})
}

func FuzzFixedBitStringTenBitDecode(f *testing.F) {
	f.Add([]byte{0xaa, 0x80}, uint8(10))
	f.Fuzz(func(t *testing.T, data []byte, length uint8) {
		bits := int(length % 17)
		if bits > len(data)*8 {
			return
		}
		r, err := NewReader(data, bits)
		if err != nil {
			return
		}
		v, err := r.ReadBitString(10, 10)
		if err == nil && (v.BitLen() != 10 || r.Position() != 10) {
			t.Fatal("incorrect ten-bit decode")
		}
	})
}

func FuzzFixedBitStringInvalidStorage(f *testing.F) {
	f.Add([]byte{0x80}, uint8(1))
	f.Fuzz(func(t *testing.T, raw []byte, size uint8) {
		_, _ = NewBitString(raw, int(size%66))
	})
}
