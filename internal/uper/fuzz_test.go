package uper

import "testing"

func FuzzReader(f *testing.F) {
	f.Add([]byte{0}, uint8(5))
	f.Fuzz(func(t *testing.T, data []byte, declared uint8) {
		bits := int(declared)
		if bits > len(data)*8 {
			bits = len(data) * 8
		}
		r, err := NewReader(data, bits)
		if err != nil {
			return
		}
		for r.Remaining() > 0 {
			_, _ = r.ReadBit()
		}
		_ = r.ValidateFinalPadding()
	})
}

func FuzzConstrainedWholeNumber(f *testing.F) {
	f.Add([]byte{0}, uint8(0), uint8(2))
	f.Fuzz(func(t *testing.T, data []byte, lower8, upper8 uint8) {
		lower, upper := uint64(lower8), uint64(upper8)
		if lower > upper {
			lower, upper = upper, lower
		}
		r, err := NewReader(data, len(data)*8)
		if err != nil {
			return
		}
		value, err := r.ReadConstrainedWholeNumber(lower, upper)
		if err == nil && (value < lower || value > upper) {
			t.Fatalf("out of range %d", value)
		}
	})
}

func FuzzBitString(f *testing.F) {
	f.Add([]byte{0x80}, uint8(1))
	f.Fuzz(func(t *testing.T, data []byte, size uint8) {
		n := int(size % 65)
		bytes := (n + 7) / 8
		if len(data) < bytes {
			return
		}
		data = append([]byte(nil), data[:bytes]...)
		if n > 0 && n%8 != 0 {
			data[bytes-1] &= byte(0xff << (8 - n%8))
		}
		v, err := NewBitString(data, n)
		if err != nil {
			return
		}
		min, max := 1, 64
		if n == 0 {
			return
		}
		w := NewWriter()
		if w.WriteBitString(v, min, max) != nil {
			return
		}
		e := w.Encoded()
		r, err := NewReader(e.Bytes, e.BitLength)
		if err != nil {
			t.Fatal(err)
		}
		got, err := r.ReadBitString(min, max)
		if err != nil || !got.Equal(v) {
			t.Fatalf("round trip: %v", err)
		}
	})
}
