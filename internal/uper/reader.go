package uper

import "fmt"

// Reader reads most-significant-bit first from an unaligned PER bit stream.
// It borrows data; callers must not mutate it during decoding.
type Reader struct {
	data      []byte
	bitLength int
	position  int
}

func NewReader(data []byte, bitLength int) (*Reader, error) {
	if bitLength < 0 || bitLength > len(data)*8 {
		return nil, fmt.Errorf("%w: bit length %d for %d bytes", ErrInvalidBitCount, bitLength, len(data))
	}
	return &Reader{data: data, bitLength: bitLength}, nil
}

func (r *Reader) Position() int  { return r.position }
func (r *Reader) BitLength() int { return r.bitLength }
func (r *Reader) Remaining() int { return r.bitLength - r.position }

func (r *Reader) ReadBit() (bool, error) {
	if r.position >= r.bitLength {
		return false, ErrUnexpectedEOF
	}
	byteIndex, bitInByte := r.position/8, r.position%8
	value := (r.data[byteIndex] & (1 << (7 - bitInByte))) != 0
	r.position++
	return value, nil
}

func (r *Reader) ReadBits(count int) (uint64, error) {
	if count < 0 || count > 64 {
		return 0, fmt.Errorf("%w: %d", ErrInvalidBitCount, count)
	}
	if count > r.Remaining() {
		return 0, ErrUnexpectedEOF
	}
	var value uint64
	for i := 0; i < count; i++ {
		bit, err := r.ReadBit()
		if err != nil {
			return 0, err
		}
		value <<= 1
		if bit {
			value |= 1
		}
	}
	return value, nil
}

// ValidateFinalPadding requires all meaningful bits to be consumed and all
// transport-octet padding bits to be zero.
func (r *Reader) ValidateFinalPadding() error {
	if r.position != r.bitLength {
		return fmt.Errorf("%w: %d meaningful bits remain", ErrUnconsumedBits, r.Remaining())
	}
	for i := r.bitLength; i < len(r.data)*8; i++ {
		if r.data[i/8]&(1<<(7-(i%8))) != 0 {
			return fmt.Errorf("%w at storage bit %d", ErrNonZeroPadding, i)
		}
	}
	return nil
}
