package uper

import (
	"errors"
	"testing"
)

func TestConstrainedWholeNumber(t *testing.T) {
	for _, tc := range []struct{ value, lower, upper uint64 }{{0, 0, 0}, {0, 0, 1}, {1, 0, 1}, {2, 0, 2}, {7, 0, 7}, {255, 0, 255}, {11, 10, 12}} {
		w := NewWriter()
		if err := w.WriteConstrainedWholeNumber(tc.value, tc.lower, tc.upper); err != nil {
			t.Fatal(err)
		}
		e := w.Encoded()
		r, err := NewReader(e.Bytes, e.BitLength)
		if err != nil {
			t.Fatal(err)
		}
		got, err := r.ReadConstrainedWholeNumber(tc.lower, tc.upper)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.value {
			t.Fatalf("got %d want %d", got, tc.value)
		}
		if err := r.ValidateFinalPadding(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestConstrainedUnusedPatternRejected(t *testing.T) {
	// Range 0..2 uses two bits; 11 is not a valid offset.
	r, _ := NewReader([]byte{0xc0}, 2)
	_, err := r.ReadConstrainedWholeNumber(0, 2)
	if !errors.Is(err, ErrDecodedOutOfRange) {
		t.Fatalf("got %v", err)
	}
}

func TestPrimitiveErrors(t *testing.T) {
	w := NewWriter()
	if !errors.Is(w.WriteConstrainedWholeNumber(3, 0, 2), ErrValueAboveUpper) {
		t.Fatal("range error")
	}
	if !errors.Is(w.WriteRootChoiceIndex(0, 0), ErrInvalidChoice) {
		t.Fatal("choice error")
	}
	if !errors.Is(w.WriteRootEnumerated(0, 0), ErrInvalidEnumerated) {
		t.Fatal("enum error")
	}
	if !errors.Is(RequireNoExtension(true), ErrExtensionUnsupported) {
		t.Fatal("extension error")
	}
	r, _ := NewReader([]byte{0xc0}, 2)
	if _, err := r.ReadRootEnumerated(3); !errors.Is(err, ErrInvalidEnumerated) {
		t.Fatalf("enumerated decode error: %v", err)
	}
	r, _ = NewReader([]byte{0xc0}, 2)
	if _, err := r.ReadRootChoiceIndex(3); !errors.Is(err, ErrInvalidChoice) {
		t.Fatalf("choice decode error: %v", err)
	}
}

func TestOptionalBitmap(t *testing.T) {
	w := NewWriter()
	if err := w.WriteOptionalBitmap([]bool{true, false, true}); err != nil {
		t.Fatal(err)
	}
	e := w.Encoded()
	r, _ := NewReader(e.Bytes, e.BitLength)
	got, err := r.ReadOptionalBitmap(3)
	if err != nil {
		t.Fatal(err)
	}
	if got[0] != true || got[1] != false || got[2] != true {
		t.Fatal(got)
	}
	if _, err := r.ReadOptionalBitmap(-1); !errors.Is(err, ErrInvalidBitmap) {
		t.Fatal(err)
	}
}
