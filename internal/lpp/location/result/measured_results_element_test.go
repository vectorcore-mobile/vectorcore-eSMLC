package result

import (
	"encoding/hex"
	"errors"
	"github.com/vectorcore/esmlc/internal/uper"
	"testing"
)

func TestMeasuredResultsElementAllOptionalMaps(t *testing.T) {
	c := cell(t, "001", "01", false, 1)
	s, _ := NewSystemFrameNumberFromUint16(0)
	rp := RSRPResult(0)
	rq := RSRQResult(0)
	u := UERxTxTimeDiff(0)
	for mask := 0; mask < 32; mask++ {
		o := MeasuredResultsElementOptions{}
		if mask&16 != 0 {
			o.CellGlobalID = &c
		}
		if mask&8 != 0 {
			o.SystemFrameNumber = &s
		}
		if mask&4 != 0 {
			o.RSRPResult = &rp
		}
		if mask&2 != 0 {
			o.RSRQResult = &rq
		}
		if mask&1 != 0 {
			o.UERxTxTimeDiff = &u
		}
		v, e := NewMeasuredResultsElement(1, 100, o)
		if e != nil {
			t.Fatal(mask, e)
		}
		w := uper.NewWriter()
		if e = v.EncodeUPER(w); e != nil {
			t.Fatal(mask, e)
		}
		x := w.Encoded()
		r, _ := uper.NewReader(x.Bytes, x.BitLength)
		g, e := DecodeMeasuredResultsElement(r)
		if e != nil || r.ValidateFinalPadding() != nil {
			t.Fatal(mask, e)
		}
		_, a := g.CellGlobalID()
		_, b := g.SystemFrameNumber()
		_, d := g.RSRPResult()
		_, f := g.RSRQResult()
		_, h := g.UERxTxTimeDiff()
		got := 0
		if a {
			got |= 16
		}
		if b {
			got |= 8
		}
		if d {
			got |= 4
		}
		if f {
			got |= 2
		}
		if h {
			got |= 1
		}
		if got != mask {
			t.Fatalf("mask %d got %d", mask, got)
		}
	}
}
func TestMeasuredResultsElementExtensionAndTruncation(t *testing.T) {
	r, _ := uper.NewReader([]byte{0x80}, 1)
	if _, e := DecodeMeasuredResultsElement(r); !errors.Is(e, ErrMeasuredResultsElementExtensions) {
		t.Fatal(e)
	}
	v, _ := NewMeasuredResultsElement(1, 100, MeasuredResultsElementOptions{})
	w := uper.NewWriter()
	_ = v.EncodeUPER(w)
	x := w.Encoded()
	for n := 0; n < x.BitLength; n++ {
		r, _ := uper.NewReader(x.Bytes, n)
		if _, e := DecodeMeasuredResultsElement(r); e == nil {
			t.Fatal(n)
		}
	}
}

func TestMeasuredResultsElementIndependentMandatoryFixture(t *testing.T) {
	// tools/specs/.../ecid-measured-results-element-fields: mandatory-only.
	want, _ := hex.DecodeString("000200c8")
	v, err := NewMeasuredResultsElement(1, 100, MeasuredResultsElementOptions{})
	if err != nil {
		t.Fatal(err)
	}
	w := uper.NewWriter()
	if err = v.EncodeUPER(w); err != nil {
		t.Fatal(err)
	}
	e := w.Encoded()
	if e.BitLength != 31 || string(e.Bytes) != string(want) {
		t.Fatalf("%x/%d", e.Bytes, e.BitLength)
	}
}
func FuzzMeasuredResultsElementDecode(f *testing.F) {
	f.Add([]byte{0}, uint8(1))
	f.Fuzz(func(t *testing.T, b []byte, n uint8) {
		if len(b) > 32 {
			return
		}
		bits := int(n)
		if bits > len(b)*8 {
			bits = len(b) * 8
		}
		r, _ := uper.NewReader(b, bits)
		_, _ = DecodeMeasuredResultsElement(r)
	})
}
