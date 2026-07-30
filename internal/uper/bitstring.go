package uper

import "fmt"

// BitString is an immutable bounded ASN.1 BIT STRING value. Bits are packed
// MSB-first in each byte. Bytes returns a copy; bits beyond BitLen are always
// zero and are not semantically significant.
type BitString struct {
	data   [8]byte
	bitLen uint8
}

// NewBitString copies exactly ceil(bitLen/8) bytes. It accepts at most 64
// meaningful bits and rejects non-zero unused bits in the final storage byte.
func NewBitString(data []byte, bitLen int) (BitString, error) {
	if bitLen < 0 || bitLen > 64 {
		return BitString{}, fmt.Errorf("%w: bit length %d", ErrInvalidBitString, bitLen)
	}
	n := (bitLen + 7) / 8
	if len(data) != n {
		return BitString{}, fmt.Errorf("%w: %d bytes for %d bits", ErrInvalidBitString, len(data), bitLen)
	}
	if n > 0 && bitLen%8 != 0 {
		unused := 8 - bitLen%8
		if data[n-1]&((1<<unused)-1) != 0 {
			return BitString{}, fmt.Errorf("%w: non-zero storage padding", ErrInvalidBitString)
		}
	}
	var value BitString
	copy(value.data[:], data)
	value.bitLen = uint8(bitLen)
	return value, nil
}

func (v BitString) BitLen() int   { return int(v.bitLen) }
func (v BitString) Bytes() []byte { return append([]byte(nil), v.data[:(int(v.bitLen)+7)/8]...) }
func (v BitString) Equal(other BitString) bool {
	if v.bitLen != other.bitLen {
		return false
	}
	n := (int(v.bitLen) + 7) / 8
	for i := 0; i < n; i++ {
		a, b := v.data[i], other.data[i]
		if i == n-1 && v.bitLen%8 != 0 {
			mask := byte(0xff << (8 - v.bitLen%8))
			a &= mask
			b &= mask
		}
		if a != b {
			return false
		}
	}
	return true
}
func (v BitString) bit(i int) bool { return v.data[i/8]&(1<<(7-(i%8))) != 0 }
func validBitStringBounds(minBits, maxBits int) error {
	if minBits < 1 || minBits > maxBits || maxBits > 64 {
		return fmt.Errorf("%w: SIZE(%d..%d)", ErrInvalidBitString, minBits, maxBits)
	}
	return nil
}

// WriteBitString encodes SIZE(minBits..maxBits) using a constrained length,
// followed by exactly BitLen MSB-first content bits (X.691 unaligned PER).
func (w *Writer) WriteBitString(value BitString, minBits, maxBits int) error {
	if err := validBitStringBounds(minBits, maxBits); err != nil {
		return err
	}
	n := value.BitLen()
	if n < minBits || n > maxBits {
		return fmt.Errorf("%w: length %d outside %d..%d", ErrInvalidBitString, n, minBits, maxBits)
	}
	if err := w.WriteConstrainedWholeNumber(uint64(n), uint64(minBits), uint64(maxBits)); err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		if err := w.WriteBit(value.bit(i)); err != nil {
			return err
		}
	}
	return nil
}

// ReadBitString decodes a bounded UPER BIT STRING and returns an independent
// immutable value; it never aliases the reader input buffer.
func (r *Reader) ReadBitString(minBits, maxBits int) (BitString, error) {
	if err := validBitStringBounds(minBits, maxBits); err != nil {
		return BitString{}, err
	}
	n, err := r.ReadConstrainedWholeNumber(uint64(minBits), uint64(maxBits))
	if err != nil {
		return BitString{}, fmt.Errorf("%w: %w", ErrInvalidBitString, err)
	}
	bits := int(n)
	var raw [8]byte
	for i := 0; i < bits; i++ {
		b, e := r.ReadBit()
		if e != nil {
			return BitString{}, e
		}
		if b {
			raw[i/8] |= 1 << (7 - (i % 8))
		}
	}
	return NewBitString(raw[:(bits+7)/8], bits)
}
