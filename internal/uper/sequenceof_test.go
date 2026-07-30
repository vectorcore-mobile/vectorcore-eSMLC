package uper

import (
	"errors"
	"sync"
	"testing"
)

func TestSequenceOfFixedAndVariableBounds(t *testing.T) {
	for _, tc := range []struct {
		name       string
		count, min int
		max        int
	}{
		{"fixed-one", 1, 1, 1}, {"fixed-four", 4, 4, 4},
		{"zero-one", 0, 0, 1}, {"one-two", 2, 1, 2},
		{"one-eight", 5, 1, 8}, {"one-thirty-two", 16, 1, 32},
		{"zero-sixty-four", 64, 0, 64},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := NewWriter()
			var wrote []int
			if err := w.WriteSequenceOf(tc.count, tc.min, tc.max, func(i int, w *Writer) error {
				wrote = append(wrote, i)
				return w.WriteBit(i%2 == 1)
			}); err != nil {
				t.Fatal(err)
			}
			width, err := constrainedWidth(uint64(tc.min), uint64(tc.max))
			if err != nil || w.BitLength() != width+tc.count {
				t.Fatalf("bits=%d width=%d count=%d err=%v", w.BitLength(), width, tc.count, err)
			}
			if len(wrote) != tc.count {
				t.Fatalf("wrote %v", wrote)
			}
			e := w.Encoded()
			r, _ := NewReader(e.Bytes, e.BitLength)
			var got []int
			count, err := r.ReadSequenceOf(tc.min, tc.max, func(i int, r *Reader) error {
				got = append(got, i)
				v, err := r.ReadBit()
				if err != nil || v != (i%2 == 1) {
					t.Fatalf("element %d: %v %v", i, v, err)
				}
				return nil
			})
			if err != nil || count != tc.count || len(got) != tc.count || r.ValidateFinalPadding() != nil {
				t.Fatalf("count=%d got=%v err=%v", count, got, err)
			}
		})
	}
}

func TestSequenceOfEmptyAndInvalidArguments(t *testing.T) {
	w := NewWriter()
	if err := w.WriteSequenceOf(0, 0, 4, nil); err != nil || w.BitLength() != 3 {
		t.Fatalf("empty: %v bits=%d", err, w.BitLength())
	}
	e := w.Encoded()
	r, _ := NewReader(e.Bytes, e.BitLength)
	if n, err := r.ReadSequenceOf(0, 4, nil); err != nil || n != 0 {
		t.Fatalf("empty decode: %d %v", n, err)
	}
	if !errors.Is(w.WriteSequenceOf(0, 1, 2, nil), ErrSequenceOfCountBelow) {
		t.Fatal("missing below-min rejection")
	}
	if !errors.Is(w.WriteSequenceOf(3, 1, 2, nil), ErrSequenceOfCountAbove) {
		t.Fatal("missing above-max rejection")
	}
	if !errors.Is(w.WriteSequenceOf(1, 1, 2, nil), ErrSequenceOfEncodeCallback) {
		t.Fatal("missing encode callback accepted")
	}
	if !errors.Is(w.WriteSequenceOf(1, -1, 2, nil), ErrInvalidSequenceOfBounds) || !errors.Is(w.WriteSequenceOf(1, 2, 1, nil), ErrInvalidSequenceOfBounds) {
		t.Fatal("invalid bounds accepted")
	}
	r, _ = NewReader([]byte{0x80}, 1)
	if _, err := r.ReadSequenceOf(1, 2, nil); !errors.Is(err, ErrSequenceOfDecodeCallback) {
		t.Fatal(err)
	}
}

func TestSequenceOfSizeOneToThirtyTwo(t *testing.T) {
	for _, count := range []int{1, 2, 16, 31, 32} {
		t.Run(string(rune('A'+count%26)), func(t *testing.T) {
			w := NewWriter()
			if err := w.WriteSequenceOf(count, 1, 32, func(i int, w *Writer) error { return w.WriteBit(i%2 == 0) }); err != nil {
				t.Fatal(err)
			}
			e := w.Encoded()
			r, _ := NewReader(e.Bytes, e.BitLength)
			offset, err := r.ReadBits(5)
			if err != nil || offset != uint64(count-1) || e.BitLength != 5+count {
				t.Fatalf("offset=%05b bits=%d err=%v", offset, e.BitLength, err)
			}
			for i := 0; i < count; i++ {
				v, err := r.ReadBit()
				if err != nil || v != (i%2 == 0) {
					t.Fatalf("element %d: %v %v", i, v, err)
				}
			}
		})
	}
}

func TestSequenceOfMultiBitAndNested(t *testing.T) {
	w := NewWriter()
	if err := w.WriteSequenceOf(3, 1, 4, func(i int, w *Writer) error {
		switch i {
		case 0:
			return w.WriteBits(1, 1)
		case 1:
			return w.WriteBits(1, 2)
		default:
			return w.WriteBits(0x15, 5)
		}
	}); err != nil {
		t.Fatal(err)
	}
	e := w.Encoded()
	if e.BitLength != 10 || string(e.Bytes) != string([]byte{0xad, 0x40}) {
		t.Fatalf("got %x/%d", e.Bytes, e.BitLength)
	}
	r, _ := NewReader(e.Bytes, e.BitLength)
	lengths := []int{1, 2, 5}
	want := []uint64{1, 1, 0x15}
	if n, err := r.ReadSequenceOf(1, 4, func(i int, r *Reader) error {
		v, err := r.ReadBits(lengths[i])
		if err != nil || v != want[i] {
			t.Fatalf("%d: %d %v", i, v, err)
		}
		return nil
	}); err != nil || n != 3 {
		t.Fatalf("nested base: %d %v", n, err)
	}

	w = NewWriter()
	if err := w.WriteSequenceOf(2, 1, 3, func(i int, w *Writer) error {
		return w.WriteSequenceOf(i, 0, 2, func(j int, w *Writer) error { return w.WriteBit((i+j)%2 == 0) })
	}); err != nil {
		t.Fatal(err)
	}
	e = w.Encoded()
	r, _ = NewReader(e.Bytes, e.BitLength)
	if n, err := r.ReadSequenceOf(1, 3, func(i int, r *Reader) error {
		inner, err := r.ReadSequenceOf(0, 2, func(j int, r *Reader) error {
			v, err := r.ReadBit()
			if err != nil || v != ((i+j)%2 == 0) {
				t.Fatalf("nested %d/%d: %v", i, j, err)
			}
			return nil
		})
		if err != nil || inner != i {
			t.Fatalf("inner %d: %d %v", i, inner, err)
		}
		return nil
	}); err != nil || n != 2 {
		t.Fatalf("outer: %d %v", n, err)
	}
}

func TestSequenceOfCallbackFailuresAndTruncation(t *testing.T) {
	cause := errors.New("element cause")
	w := NewWriter()
	called := 0
	err := w.WriteSequenceOf(3, 1, 3, func(i int, w *Writer) error {
		called++
		if i == 1 {
			return cause
		}
		return w.WriteBit(true)
	})
	if !errors.Is(err, ErrSequenceOfElementEncode) || !errors.Is(err, cause) || called != 2 || w.BitLength() != 3 {
		t.Fatalf("writer error=%v called=%d bits=%d", err, called, w.BitLength())
	}

	w = NewWriter()
	_ = w.WriteSequenceOf(3, 1, 3, func(i int, w *Writer) error { return w.WriteBit(true) })
	e := w.Encoded()
	r, _ := NewReader(e.Bytes, e.BitLength)
	called = 0
	_, err = r.ReadSequenceOf(1, 3, func(i int, r *Reader) error {
		called++
		if i == 1 {
			return cause
		}
		_, _ = r.ReadBit()
		return nil
	})
	if !errors.Is(err, ErrSequenceOfElementDecode) || !errors.Is(err, cause) || called != 2 || r.Position() != 3 {
		t.Fatalf("reader error=%v called=%d pos=%d", err, called, r.Position())
	}

	r, _ = NewReader([]byte{0}, 3)
	if _, err = r.ReadSequenceOf(1, 32, func(int, *Reader) error { return nil }); !errors.Is(err, ErrSequenceOfTruncatedCount) {
		t.Fatal(err)
	}
	w = NewWriter()
	_ = w.WriteSequenceOf(2, 1, 2, func(int, *Writer) error { return nil })
	e = w.Encoded()
	r, _ = NewReader(e.Bytes, e.BitLength)
	called = 0
	_, err = r.ReadSequenceOf(1, 2, func(int, *Reader) error { called++; _, err := r.ReadBit(); return err })
	if !errors.Is(err, ErrSequenceOfTruncatedElement) || called != 1 {
		t.Fatalf("truncation=%v called=%d", err, called)
	}
}

func TestSequenceOfHighBoundDeterminismAndIndependentConcurrency(t *testing.T) {
	encode := func() Encoded {
		w := NewWriter()
		if err := w.WriteSequenceOf(1024, 0, 1024, func(int, *Writer) error { return nil }); err != nil {
			t.Fatal(err)
		}
		return w.Encoded()
	}
	a, b := encode(), encode()
	if string(a.Bytes) != string(b.Bytes) || a.BitLength != b.BitLength {
		t.Fatal("non-deterministic encoding")
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e := encode()
			r, err := NewReader(e.Bytes, e.BitLength)
			if err != nil {
				t.Error(err)
				return
			}
			n, err := r.ReadSequenceOf(0, 1024, func(int, *Reader) error { return nil })
			if err != nil || n != 1024 {
				t.Errorf("%d %v", n, err)
			}
		}()
	}
	wg.Wait()
}
