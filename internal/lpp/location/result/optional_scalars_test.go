package result

import (
	"errors"
	"github.com/vectorcore/esmlc/internal/uper"
	"testing"
)

func TestOptionalScalarFixtures(t *testing.T) {
	for _, tc := range []struct {
		name   string
		bits   int
		bytes  []byte
		encode func(*uper.Writer) error
		decode func(*uper.Reader) (uint64, error)
	}{
		{"rsrp-0", 7, []byte{0}, func(w *uper.Writer) error { return RSRPResult(0).EncodeUPER(w) }, func(r *uper.Reader) (uint64, error) { v, e := DecodeRSRPResult(r); return uint64(v), e }},
		{"rsrp-49", 7, []byte{0x62}, func(w *uper.Writer) error { return RSRPResult(49).EncodeUPER(w) }, func(r *uper.Reader) (uint64, error) { v, e := DecodeRSRPResult(r); return uint64(v), e }},
		{"rsrp-97", 7, []byte{0xc2}, func(w *uper.Writer) error { return RSRPResult(97).EncodeUPER(w) }, func(r *uper.Reader) (uint64, error) { v, e := DecodeRSRPResult(r); return uint64(v), e }},
		{"rsrq-0", 6, []byte{0}, func(w *uper.Writer) error { return RSRQResult(0).EncodeUPER(w) }, func(r *uper.Reader) (uint64, error) { v, e := DecodeRSRQResult(r); return uint64(v), e }},
		{"rsrq-17", 6, []byte{0x44}, func(w *uper.Writer) error { return RSRQResult(17).EncodeUPER(w) }, func(r *uper.Reader) (uint64, error) { v, e := DecodeRSRQResult(r); return uint64(v), e }},
		{"rsrq-34", 6, []byte{0x88}, func(w *uper.Writer) error { return RSRQResult(34).EncodeUPER(w) }, func(r *uper.Reader) (uint64, error) { v, e := DecodeRSRQResult(r); return uint64(v), e }},
		{"ue-0", 12, []byte{0, 0}, func(w *uper.Writer) error { return UERxTxTimeDiff(0).EncodeUPER(w) }, func(r *uper.Reader) (uint64, error) { v, e := DecodeUERxTxTimeDiff(r); return uint64(v), e }},
		{"ue-2048", 12, []byte{0x80, 0}, func(w *uper.Writer) error { return UERxTxTimeDiff(2048).EncodeUPER(w) }, func(r *uper.Reader) (uint64, error) { v, e := DecodeUERxTxTimeDiff(r); return uint64(v), e }},
		{"ue-4095", 12, []byte{0xff, 0xf0}, func(w *uper.Writer) error { return UERxTxTimeDiff(4095).EncodeUPER(w) }, func(r *uper.Reader) (uint64, error) { v, e := DecodeUERxTxTimeDiff(r); return uint64(v), e }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := uper.NewWriter()
			if e := tc.encode(w); e != nil {
				t.Fatal(e)
			}
			x := w.Encoded()
			if x.BitLength != tc.bits || string(x.Bytes) != string(tc.bytes) {
				t.Fatalf("%x/%d", x.Bytes, x.BitLength)
			}
			r, _ := uper.NewReader(x.Bytes, x.BitLength)
			if _, e := tc.decode(r); e != nil || r.ValidateFinalPadding() != nil {
				t.Fatal(e)
			}
		})
	}
}

func TestSystemFrameNumber(t *testing.T) {
	for _, v := range []uint16{0, 1, 512, 682, 1023} {
		s, e := NewSystemFrameNumberFromUint16(v)
		if e != nil || s.Uint16() != v {
			t.Fatal(v, e)
		}
		w := uper.NewWriter()
		if e = s.EncodeUPER(w); e != nil {
			t.Fatal(e)
		}
		x := w.Encoded()
		if x.BitLength != 10 {
			t.Fatal(x.BitLength)
		}
		r, _ := uper.NewReader(x.Bytes, 10)
		g, e := DecodeSystemFrameNumber(r)
		if e != nil || g.Uint16() != v || !g.BitString().Equal(s.BitString()) {
			t.Fatal(v, e)
		}
	}
	b, e := uper.NewBitString([]byte{0xaa, 0x80}, 10)
	if e != nil {
		t.Fatal(e)
	}
	s, e := NewSystemFrameNumber(b)
	if e != nil || s.Uint16() != 682 {
		t.Fatal(e)
	}
	for _, n := range []int{0, 1, 9, 11, 16} {
		x, _ := uper.NewBitString(make([]byte, (n+7)/8), n)
		if _, e := NewSystemFrameNumber(x); !errors.Is(e, ErrSystemFrameNumberInvalidBitLength) {
			t.Fatal(n, e)
		}
	}
}

func TestOptionalScalarBoundsAndTruncation(t *testing.T) {
	if _, e := NewRSRPResult(98); !errors.Is(e, ErrRSRPResultOutOfRange) {
		t.Fatal(e)
	}
	if _, e := NewRSRQResult(35); !errors.Is(e, ErrRSRQResultOutOfRange) {
		t.Fatal(e)
	}
	if _, e := NewUERxTxTimeDiff(4096); !errors.Is(e, ErrUERxTxTimeDiffOutOfRange) {
		t.Fatal(e)
	}
	if _, e := NewSystemFrameNumberFromUint16(1024); !errors.Is(e, ErrSystemFrameNumberOutOfRange) {
		t.Fatal(e)
	}
	for n := 0; n < 7; n++ {
		r, _ := uper.NewReader([]byte{0xff}, n)
		_, e := DecodeRSRPResult(r)
		if !errors.Is(e, uper.ErrUnexpectedEOF) {
			t.Fatal(n, e)
		}
	}
	for n := 0; n < 6; n++ {
		r, _ := uper.NewReader([]byte{0xff}, n)
		_, e := DecodeRSRQResult(r)
		if !errors.Is(e, uper.ErrUnexpectedEOF) {
			t.Fatal(n, e)
		}
	}
	for n := 0; n < 12; n++ {
		r, _ := uper.NewReader([]byte{0xff, 0xff}, n)
		_, e := DecodeUERxTxTimeDiff(r)
		if !errors.Is(e, uper.ErrUnexpectedEOF) {
			t.Fatal(n, e)
		}
	}
	for n := 0; n < 10; n++ {
		r, _ := uper.NewReader([]byte{0xff, 0xff}, n)
		_, e := DecodeSystemFrameNumber(r)
		if !errors.Is(e, uper.ErrUnexpectedEOF) {
			t.Fatal(n, e)
		}
	}
}

func TestOptionalScalarComposition(t *testing.T) {
	s, _ := NewSystemFrameNumberFromUint16(682)
	w := uper.NewWriter()
	_ = w.WriteBit(true)
	_ = RSRPResult(49).EncodeUPER(w)
	_ = RSRQResult(17).EncodeUPER(w)
	_ = UERxTxTimeDiff(2048).EncodeUPER(w)
	_ = s.EncodeUPER(w)
	_ = w.WriteConstrainedWholeNumber(2, 0, 3)
	x := w.Encoded()
	if x.BitLength != 38 {
		t.Fatal(x.BitLength)
	}
	r, _ := uper.NewReader(x.Bytes, x.BitLength)
	_, _ = r.ReadBit()
	a, _ := DecodeRSRPResult(r)
	b, _ := DecodeRSRQResult(r)
	c, _ := DecodeUERxTxTimeDiff(r)
	d, _ := DecodeSystemFrameNumber(r)
	z, _ := r.ReadConstrainedWholeNumber(0, 3)
	if a != 49 || b != 17 || c != 2048 || d.Uint16() != 682 || z != 2 {
		t.Fatal("composition")
	}
}
