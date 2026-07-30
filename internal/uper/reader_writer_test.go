package uper

import (
	"errors"
	"testing"
)

func TestReaderWriterBitOrderAndPadding(t *testing.T) {
	w := NewWriter()
	if err := w.WriteBits(0b101100011, 9); err != nil {
		t.Fatal(err)
	}
	e := w.Encoded()
	if e.BitLength != 9 || e.UnusedBits != 7 || len(e.Bytes) != 2 {
		t.Fatalf("%+v", e)
	}
	if e.Bytes[0] != 0xb1 || e.Bytes[1] != 0x80 {
		t.Fatalf("%x", e.Bytes)
	}
	r, err := NewReader(e.Bytes, e.BitLength)
	if err != nil {
		t.Fatal(err)
	}
	v, err := r.ReadBits(9)
	if err != nil || v != 0b101100011 {
		t.Fatalf("%b %v", v, err)
	}
	if err := r.ValidateFinalPadding(); err != nil {
		t.Fatal(err)
	}
}

func TestReaderRejectsMalformedBoundaries(t *testing.T) {
	if _, err := NewReader([]byte{0}, 9); !errors.Is(err, ErrInvalidBitCount) {
		t.Fatal(err)
	}
	r, _ := NewReader([]byte{0}, 0)
	if _, err := r.ReadBit(); !errors.Is(err, ErrUnexpectedEOF) {
		t.Fatal(err)
	}
	r, _ = NewReader([]byte{0x01}, 5)
	if err := r.ValidateFinalPadding(); !errors.Is(err, ErrUnconsumedBits) {
		t.Fatal(err)
	}
	r, _ = NewReader([]byte{0x04}, 5)
	if _, err := r.ReadBits(5); err != nil {
		t.Fatal(err)
	}
	if err := r.ValidateFinalPadding(); !errors.Is(err, ErrNonZeroPadding) {
		t.Fatal(err)
	}
}

func TestWriterOwnership(t *testing.T) {
	w := NewWriter()
	_ = w.WriteBit(true)
	first := w.Encoded()
	_ = w.WriteBit(true)
	if len(first.Bytes) != 1 || first.Bytes[0] != 0x80 {
		t.Fatalf("writer mutated returned bytes: %x", first.Bytes)
	}
}
