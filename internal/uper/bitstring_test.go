package uper

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func mustBitString(t *testing.T, b []byte, n int) BitString {
	t.Helper()
	v, e := NewBitString(b, n)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func TestBitStringOneThroughEight(t *testing.T) {
	for n := 1; n <= 8; n++ {
		t.Run(string(rune('0'+n)), func(t *testing.T) {
			raw := byte(0xaa)
			if n == 8 {
				raw = 0xff
			}
			if n < 8 {
				raw &= byte(0xff << (8 - n))
			}
			v := mustBitString(t, []byte{raw}, n)
			w := NewWriter()
			if e := w.WriteBitString(v, 1, 8); e != nil {
				t.Fatal(e)
			}
			x := w.Encoded()
			r, e := NewReader(x.Bytes, x.BitLength)
			if e != nil {
				t.Fatal(e)
			}
			got, e := r.ReadBitString(1, 8)
			if e != nil || !got.Equal(v) {
				t.Fatalf("%v %#v", e, got)
			}
			if e = r.ValidateFinalPadding(); e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestBitStringLargeAndCopySafety(t *testing.T) {
	for _, n := range []int{16, 32, 64} {
		raw := make([]byte, n/8)
		for i := range raw {
			raw[i] = byte(0xa5)
		}
		v := mustBitString(t, raw, n)
		raw[0] = 0
		if v.Bytes()[0] != 0xa5 {
			t.Fatal("constructor aliases input")
		}
		out := v.Bytes()
		out[0] = 0
		if v.Bytes()[0] != 0xa5 {
			t.Fatal("Bytes aliases storage")
		}
		w := NewWriter()
		if e := w.WriteBitString(v, 1, 64); e != nil {
			t.Fatal(e)
		}
		r, _ := NewReader(w.Encoded().Bytes, w.Encoded().BitLength)
		got, e := r.ReadBitString(1, 64)
		if e != nil || !got.Equal(v) {
			t.Fatal(e)
		}
	}
}
func TestBitStringErrors(t *testing.T) {
	if _, e := NewBitString(nil, -1); !errors.Is(e, ErrInvalidBitString) {
		t.Fatal(e)
	}
	if _, e := NewBitString([]byte{0}, 0); !errors.Is(e, ErrInvalidBitString) {
		t.Fatal(e)
	}
	if _, e := NewBitString([]byte{1}, 1); !errors.Is(e, ErrInvalidBitString) {
		t.Fatal(e)
	}
	v := mustBitString(t, []byte{0x80}, 1)
	w := NewWriter()
	if e := w.WriteBitString(v, 0, 8); !errors.Is(e, ErrInvalidBitString) {
		t.Fatal(e)
	}
	if e := w.WriteBitString(v, 1, 0); !errors.Is(e, ErrInvalidBitString) {
		t.Fatal(e)
	}
	if e := w.WriteBitString(v, 2, 8); !errors.Is(e, ErrInvalidBitString) {
		t.Fatal(e)
	}
	r, _ := NewReader([]byte{0}, 1)
	if _, e := r.ReadBitString(1, 8); !errors.Is(e, ErrUnexpectedEOF) {
		t.Fatal(e)
	}
}
func TestBitStringFixtureValues(t *testing.T) {
	root, e := filepath.Abs("../../tools/specs/lpp/fixtures/r16.4.0/capabilities")
	if e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(filepath.Join(root, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	var m struct {
		Fixtures []struct {
			Name    string         `json:"name"`
			Decoded map[string]any `json:"decoded"`
		} `json:"fixtures"`
	}
	if e = json.Unmarshal(b, &m); e != nil {
		t.Fatal(e)
	}
	want := map[string]struct {
		raw byte
		n   int
	}{"provide-ecid-rsrp": {0x80, 1}, "provide-ecid-rsrp-rsrq-uerxtx": {0xe0, 3}}
	for _, f := range m.Fixtures {
		if x, ok := want[f.Name]; ok {
			v := mustBitString(t, []byte{x.raw}, x.n)
			w := NewWriter()
			if e := w.WriteBitString(v, 1, 8); e != nil {
				t.Fatal(e)
			}
			r, _ := NewReader(w.Encoded().Bytes, w.Encoded().BitLength)
			got, e := r.ReadBitString(1, 8)
			if e != nil || !got.Equal(v) {
				t.Fatalf("%s: %v", f.Name, e)
			}
		}
	}
}
