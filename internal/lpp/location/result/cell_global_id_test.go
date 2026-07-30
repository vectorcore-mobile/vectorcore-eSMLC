package result

import (
	"encoding/hex"
	"errors"
	"github.com/vectorcore/esmlc/internal/uper"
	"testing"
)

func cell(t *testing.T, mcc string, mnc string, utra bool, n uint32) CellGlobalIdEUTRAAndUTRA {
	t.Helper()
	a, _ := NewMCC(mcc[0]-'0', mcc[1]-'0', mcc[2]-'0')
	var b MNC
	var e error
	if len(mnc) == 2 {
		b, e = NewMNC2(mnc[0]-'0', mnc[1]-'0')
	} else {
		b, e = NewMNC3(mnc[0]-'0', mnc[1]-'0', mnc[2]-'0')
	}
	if e != nil {
		t.Fatal(e)
	}
	p, _ := NewPLMNIdentity(a, b)
	var c CellIdentity
	if utra {
		x, e := NewUTRACellIdentityFromUint32(n)
		if e != nil {
			t.Fatal(e)
		}
		c = NewUTRACellIdentityChoice(x)
	} else {
		x, e := NewEUTRACellIdentityFromUint32(n)
		if e != nil {
			t.Fatal(e)
		}
		c = NewEUTRACellIdentityChoice(x)
	}
	v, e := NewCellGlobalIdEUTRAAndUTRA(p, c)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func TestCellGlobalIDFixtures(t *testing.T) {
	for _, x := range []struct {
		m, n string
		u    bool
		v    uint32
		h    string
		b    int
	}{{"001", "01", false, 0, "00080400000000", 51}, {"001", "01", false, 1, "00080400000020", 51}, {"310", "260", false, 0x1234567, "18849802468ace", 55}, {"310", "260", false, 0xfffffff, "1884981ffffffe", 55}, {"001", "01", true, 0, "00080600000000", 55}, {"001", "01", true, 1, "00080600000002", 55}, {"310", "260", true, 0x12345678, "18849822468acf00", 59}, {"310", "260", true, 0xffffffff, "1884983fffffffe0", 59}} {
		v := cell(t, x.m, x.n, x.u, x.v)
		w := uper.NewWriter()
		if e := v.EncodeUPER(w); e != nil {
			t.Fatal(e)
		}
		z := w.Encoded()
		want, _ := hex.DecodeString(x.h)
		if z.BitLength != x.b || string(z.Bytes) != string(want) {
			t.Fatalf("%x/%d want %x/%d", z.Bytes, z.BitLength, want, x.b)
		}
		r, _ := uper.NewReader(z.Bytes, z.BitLength)
		g, e := DecodeCellGlobalIdEUTRAAndUTRA(r)
		if e != nil || g.p.mcc.String() != x.m || g.p.mnc.String() != x.n || r.ValidateFinalPadding() != nil {
			t.Fatal(e)
		}
	}
}
func TestCellGlobalIDRejectionAndTruncation(t *testing.T) {
	if _, e := NewPLMNDigit(10); !errors.Is(e, ErrPLMNDigitOutOfRange) {
		t.Fatal(e)
	}
	v := cell(t, "001", "01", false, 1)
	w := uper.NewWriter()
	_ = v.EncodeUPER(w)
	z := w.Encoded()
	for n := 0; n < z.BitLength; n++ {
		r, _ := uper.NewReader(z.Bytes, n)
		if _, e := DecodeCellGlobalIdEUTRAAndUTRA(r); e == nil {
			t.Fatal(n)
		}
	}
	r, _ := uper.NewReader([]byte{0x80}, 1)
	if _, e := DecodeCellGlobalIdEUTRAAndUTRA(r); !errors.Is(e, ErrCellGlobalIDExtensionsUnsupported) {
		t.Fatal(e)
	}
}

func TestCellGlobalIDNonAlignedComposition(t *testing.T) {
	v := cell(t, "208", "01", true, 0x12345678)
	w := uper.NewWriter()
	_ = w.WriteBits(5, 3)
	if err := v.EncodeUPER(w); err != nil {
		t.Fatal(err)
	}
	_ = w.WriteBits(17, 5)
	e := w.Encoded()
	if e.BitLength != 3+55+5 {
		t.Fatal(e.BitLength)
	}
	r, _ := uper.NewReader(e.Bytes, e.BitLength)
	if n, _ := r.ReadBits(3); n != 5 {
		t.Fatal(n)
	}
	g, err := DecodeCellGlobalIdEUTRAAndUTRA(r)
	if err != nil || g.p.mcc.String() != "208" || g.p.mnc.String() != "01" {
		t.Fatal(err)
	}
	if n, _ := r.ReadBits(5); n != 17 {
		t.Fatal(n)
	}
}
func FuzzCellGlobalIdDecode(f *testing.F) {
	f.Add([]byte{0}, uint8(1))
	f.Fuzz(func(t *testing.T, b []byte, n uint8) {
		if len(b) > 16 {
			return
		}
		bits := int(n)
		if bits > len(b)*8 {
			bits = len(b) * 8
		}
		r, _ := uper.NewReader(b, bits)
		_, _ = DecodeCellGlobalIdEUTRAAndUTRA(r)
	})
}
