package uper

import "fmt"

// Encoded is a finalized UPER bit stream. Bytes is a copy and can be retained
// or modified by the caller without changing the writer.
type Encoded struct {
	Bytes      []byte
	BitLength  int
	UnusedBits uint8
}

type Writer struct {
	data      []byte
	bitLength int
}

func NewWriter() *Writer         { return &Writer{} }
func (w *Writer) BitLength() int { return w.bitLength }

func (w *Writer) WriteBit(value bool) error {
	if w.bitLength == int(^uint(0)>>1) {
		return ErrIntegerWidth
	}
	byteIndex, bitInByte := w.bitLength/8, w.bitLength%8
	if byteIndex == len(w.data) {
		w.data = append(w.data, 0)
	}
	if value {
		w.data[byteIndex] |= 1 << (7 - bitInByte)
	}
	w.bitLength++
	return nil
}

func (w *Writer) WriteBits(value uint64, count int) error {
	if count < 0 || count > 64 {
		return fmt.Errorf("%w: %d", ErrInvalidBitCount, count)
	}
	if count < 64 && value >= (uint64(1)<<count) {
		return fmt.Errorf("%w: value %d needs more than %d bits", ErrIntegerWidth, value, count)
	}
	for shift := count - 1; shift >= 0; shift-- {
		if err := w.WriteBit((value & (uint64(1) << shift)) != 0); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) Encoded() Encoded {
	data := append([]byte(nil), w.data...)
	unused := uint8((8 - (w.bitLength % 8)) % 8)
	return Encoded{Bytes: data, BitLength: w.bitLength, UnusedBits: unused}
}
