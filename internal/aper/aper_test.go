package aper

import (
	"encoding/hex"
	"testing"
)

// TestConstrainedThreeRegimes locks in the X.691 10.5.7 aligned constrained
// whole number rule across its three span regimes: a non-aligned bit-field
// for span<=255, a fixed whole-octet width for 255<span<=65536, and an
// octet-aligned length-determinant-prefixed encoding for span>65536. Each
// expected byte string was independently verified against asn1tools 0.167.0
// (see tools/specs/lcsap/reference-codec); regressing this rounds a
// constrained integer to its raw bit width again, which silently corrupted
// every LCS-AP Location-Estimate coordinate on the wire.
func TestConstrainedThreeRegimes(t *testing.T) {
	cases := []struct {
		name    string
		lo, hi  int64
		v       int64
		wantHex string
	}{
		{"small span=101 no alignment", 0, 100, 68, "88"},
		{"mid span=1000 fixed 2 octets zero", 0, 999, 0, "0000"},
		{"mid span=1000 fixed 2 octets", 0, 999, 999, "03e7"},
		{"mid boundary span=65536 exact 2 octets", 0, 65535, 256, "0100"},
		{"large span=8388608 selector + 1 octet, zero", 0, 8388607, 0, "0000"},
		{"large span=8388608 selector + 1 octet, max", 0, 8388607, 255, "00ff"},
		{"large span=8388608 selector + 2 octets", 0, 8388607, 256, "400100"},
		{"large span=8388608 selector + 3 octets", 0, 8388607, 3541856, "80360b60"},
		{"large span=8388608 selector + 3 octets, max", 0, 8388607, 8388607, "807fffff"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := NewWriter()
			if err := PutConstrained(w, c.v, c.lo, c.hi); err != nil {
				t.Fatalf("encode: %v", err)
			}
			got := w.Bytes()
			want, err := hex.DecodeString(c.wantHex)
			if err != nil {
				t.Fatalf("bad fixture hex: %v", err)
			}
			if string(got) != string(want) {
				t.Fatalf("got %x want %x", got, want)
			}
			v, err := GetConstrained(NewReader(got), c.lo, c.hi)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if v != c.v {
				t.Fatalf("round trip: got %d want %d", v, c.v)
			}
		})
	}
}

// TestFixedBitStringAlignment locks in the X.691 aligned-PER rule that a
// fixed-length BIT STRING is octet-aligned before its content only when
// longer than 16 bits (used by LPPa's 28-bit EUTRANCellIdentifier). A leading
// 1-bit field forces an odd bit offset so the two regimes produce different
// wire bytes if the alignment rule is implemented backwards.
func TestFixedBitStringAlignment(t *testing.T) {
	cases := []struct {
		name    string
		size    int
		v       uint64
		wantHex string
	}{
		{"16 bits not aligned", 16, 0xabcd, "d5e680"},
		{"28 bits aligned", 28, 0x0abcdef, "800abcdef0"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := NewWriter()
			w.bits(1, 1)
			if err := PutFixedBitString(w, c.v, c.size); err != nil {
				t.Fatalf("encode: %v", err)
			}
			got := w.Bytes()
			want, err := hex.DecodeString(c.wantHex)
			if err != nil {
				t.Fatalf("bad fixture hex: %v", err)
			}
			if string(got) != string(want) {
				t.Fatalf("got %x want %x", got, want)
			}
			r := NewReader(got)
			if _, err := r.bits(1); err != nil {
				t.Fatalf("read leading bit: %v", err)
			}
			v, err := GetFixedBitString(r, c.size)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if v != c.v {
				t.Fatalf("round trip: got %x want %x", v, c.v)
			}
		})
	}
}
