package uper

import (
	"errors"
	"testing"
)

func FuzzSequenceOfRoundTrip(f *testing.F) {
	f.Add(uint8(1), uint8(8), []byte{0x80})
	f.Add(uint8(1), uint8(32), []byte{0xaa, 0x80})
	f.Fuzz(func(t *testing.T, min8, max8 uint8, values []byte) {
		min, max := int(min8%5), int(max8%33)
		if min > max {
			min, max = max, min
		}
		count := min
		if max > min {
			count += len(values) % (max - min + 1)
		}
		w := NewWriter()
		if err := w.WriteSequenceOf(count, min, max, func(i int, w *Writer) error { return w.WriteBit(i < len(values) && values[i]&1 != 0) }); err != nil {
			t.Fatal(err)
		}
		e := w.Encoded()
		r, err := NewReader(e.Bytes, e.BitLength)
		if err != nil {
			t.Fatal(err)
		}
		n, err := r.ReadSequenceOf(min, max, func(i int, r *Reader) error {
			v, err := r.ReadBit()
			if err != nil || v != (i < len(values) && values[i]&1 != 0) {
				t.Fatal("element changed")
			}
			return nil
		})
		if err != nil || n != count || r.ValidateFinalPadding() != nil {
			t.Fatalf("count=%d want=%d err=%v", n, count, err)
		}
	})
}

func FuzzSequenceOfDecode(f *testing.F) {
	f.Add([]byte{0x00}, uint8(1), uint8(8))
	f.Add([]byte{0xff}, uint8(1), uint8(32))
	f.Fuzz(func(t *testing.T, data []byte, min8, max8 uint8) {
		min, max := int(min8%5), int(max8%33)
		if min > max {
			min, max = max, min
		}
		r, err := NewReader(data, len(data)*8)
		if err != nil {
			return
		}
		calls := 0
		n, err := r.ReadSequenceOf(min, max, func(int, *Reader) error {
			calls++
			_, err := r.ReadBit()
			return err
		})
		if err == nil && (n < min || n > max || calls != n) {
			t.Fatal("invalid decoded count")
		}
	})
}

func FuzzSequenceOfCallbackFailure(f *testing.F) {
	f.Add(uint8(3), uint8(1))
	f.Fuzz(func(t *testing.T, count8, fail8 uint8) {
		count := int(count8%8) + 1
		fail := int(fail8 % uint8(count))
		cause := errors.New("fuzz callback failure")
		w := NewWriter()
		called := 0
		err := w.WriteSequenceOf(count, 1, 8, func(i int, w *Writer) error {
			called++
			if i == fail {
				return cause
			}
			return nil
		})
		if !errors.Is(err, cause) || called != fail+1 {
			t.Fatal("callback failure propagation")
		}
	})
}
