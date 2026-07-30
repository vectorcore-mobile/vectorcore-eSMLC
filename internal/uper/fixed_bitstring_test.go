package uper

import (
	"errors"
	"testing"
)

func TestFixedBitStringTenBitVectors(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  []byte
	}{
		{"all-zero", []byte{0x00, 0x00}},
		{"low-bit", []byte{0x00, 0x40}},
		{"high-bit", []byte{0x80, 0x00}},
		{"alternating", []byte{0xaa, 0x80}},
		{"all-one", []byte{0xff, 0xc0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := mustBitString(t, tc.raw, 10)
			w := NewWriter()
			if err := w.WriteBitString(v, 10, 10); err != nil {
				t.Fatal(err)
			}
			e := w.Encoded()
			if e.BitLength != 10 || string(e.Bytes) != string(tc.raw) || e.UnusedBits != 6 {
				t.Fatalf("got %x/%d/%d", e.Bytes, e.BitLength, e.UnusedBits)
			}
			r, _ := NewReader(e.Bytes, e.BitLength)
			got, err := r.ReadBitString(10, 10)
			if err != nil || got.BitLen() != 10 || !got.Equal(v) || r.Position() != 10 || r.ValidateFinalPadding() != nil {
				t.Fatalf("got=%#v err=%v pos=%d", got, err, r.Position())
			}
		})
	}
}

func TestFixedBitStringSizesAndEquality(t *testing.T) {
	for _, n := range []int{1, 2, 7, 8, 9, 10, 16, 31, 32, 64} {
		t.Run(string(rune('a'+n%26)), func(t *testing.T) {
			bytes := make([]byte, (n+7)/8)
			for i := range bytes {
				bytes[i] = 0xa5
			}
			if n%8 != 0 {
				bytes[len(bytes)-1] &= byte(0xff << (8 - n%8))
			}
			v := mustBitString(t, bytes, n)
			w := NewWriter()
			if err := w.WriteBitString(v, n, n); err != nil {
				t.Fatal(err)
			}
			if w.BitLength() != n {
				t.Fatalf("fixed size %d wrote %d bits", n, w.BitLength())
			}
			r, _ := NewReader(w.Encoded().Bytes, n)
			got, err := r.ReadBitString(n, n)
			if err != nil || !got.Equal(v) || got.BitLen() != n {
				t.Fatalf("size %d: %v", n, err)
			}
		})
	}
	one := mustBitString(t, []byte{0x80}, 1)
	tenTrailingZero := mustBitString(t, []byte{0x80, 0x00}, 10)
	if one.Equal(tenTrailingZero) {
		t.Fatal("different fixed and variable lengths compared equal")
	}
}

func TestFixedBitStringNonAlignedComposition(t *testing.T) {
	v := mustBitString(t, []byte{0xaa, 0x80}, 10)
	w := NewWriter()
	if err := w.WriteBoolean(true); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteBitString(v, 10, 10); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteConstrainedWholeNumber(2, 0, 3); err != nil {
		t.Fatal(err)
	}
	e := w.Encoded()
	if e.BitLength != 13 || string(e.Bytes) != string([]byte{0xd5, 0x50}) {
		t.Fatalf("got %x/%d", e.Bytes, e.BitLength)
	}
	r, _ := NewReader(e.Bytes, e.BitLength)
	b, err := r.ReadBoolean()
	if err != nil || !b {
		t.Fatal(err)
	}
	got, err := r.ReadBitString(10, 10)
	if err != nil || !got.Equal(v) {
		t.Fatal(err)
	}
	n, err := r.ReadConstrainedWholeNumber(0, 3)
	if err != nil || n != 2 || r.ValidateFinalPadding() != nil {
		t.Fatalf("following field: %d %v", n, err)
	}

	w = NewWriter()
	if err = w.WriteSequenceOf(16, 1, 32, func(int, *Writer) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if err = w.WriteBitString(v, 10, 10); err != nil {
		t.Fatal(err)
	}
	if err = w.WriteBoolean(true); err != nil {
		t.Fatal(err)
	}
	e = w.Encoded()
	r, _ = NewReader(e.Bytes, e.BitLength)
	if count, err := r.ReadSequenceOf(1, 32, func(int, *Reader) error { return nil }); err != nil || count != 16 {
		t.Fatalf("count %d %v", count, err)
	}
	got, err = r.ReadBitString(10, 10)
	if err != nil || !got.Equal(v) {
		t.Fatal(err)
	}
	b, err = r.ReadBoolean()
	if err != nil || !b || r.ValidateFinalPadding() != nil {
		t.Fatal(err)
	}
}

func TestFixedBitStringInvalidAndTruncated(t *testing.T) {
	v9 := mustBitString(t, []byte{0x80, 0x00}, 9)
	v11 := mustBitString(t, []byte{0x80, 0x00}, 11)
	w := NewWriter()
	if !errors.Is(w.WriteBitString(v9, 10, 10), ErrInvalidBitString) || !errors.Is(w.WriteBitString(v11, 10, 10), ErrInvalidBitString) {
		t.Fatal("mismatched fixed length accepted")
	}
	if !errors.Is(w.WriteBitString(v9, 0, 0), ErrInvalidBitString) || !errors.Is(w.WriteBitString(v9, 65, 65), ErrInvalidBitString) {
		t.Fatal("invalid fixed bounds accepted")
	}
	for bits := 0; bits < 10; bits++ {
		r, _ := NewReader([]byte{0xaa, 0x80}, bits)
		if _, err := r.ReadBitString(10, 10); !errors.Is(err, ErrUnexpectedEOF) {
			t.Fatalf("truncation %d: %v", bits, err)
		}
	}
	r, _ := NewReader([]byte{0xaa, 0x81}, 10)
	if _, err := r.ReadBitString(10, 10); err != nil {
		t.Fatal(err)
	}
	if err := r.ValidateFinalPadding(); !errors.Is(err, ErrNonZeroPadding) {
		t.Fatal(err)
	}
}
