// Package aper contains the bounded APER primitives shared by LCS-AP and LPPa.
package aper

import (
	"fmt"
	"io"
	"math/bits"
)

const MaxOpenType = 1 << 20

type Criticality uint8

const (
	Reject Criticality = iota
	Ignore
	Notify
)

type Reader struct {
	data []byte
	pos  int
}

func NewReader(b []byte) *Reader { return &Reader{data: b} }
func (r *Reader) Remaining() int { return len(r.data)*8 - r.pos }

// RemainingZero reports whether unread alignment padding contains only zero
// bits. APER encoders must write zero padding; accepting non-zero padding can
// turn malformed protocol data into a valid PDU.
func (r *Reader) RemainingZero() bool {
	for i := r.pos; i < len(r.data)*8; i++ {
		if (r.data[i/8]>>uint(7-i%8))&1 != 0 {
			return false
		}
	}
	return true
}
func (r *Reader) align() {
	if m := r.pos % 8; m != 0 {
		r.pos += 8 - m
	}
}
func (r *Reader) bit() (uint8, error) {
	if r.pos >= len(r.data)*8 {
		return 0, io.ErrUnexpectedEOF
	}
	b := (r.data[r.pos/8] >> uint(7-r.pos%8)) & 1
	r.pos++
	return b, nil
}
func (r *Reader) bits(n int) (uint64, error) {
	if n < 1 || n > 64 {
		return 0, fmt.Errorf("aper: invalid bit length %d", n)
	}
	if r.pos+n > len(r.data)*8 {
		return 0, io.ErrUnexpectedEOF
	}
	var v uint64
	for i := 0; i < n; i++ {
		b, _ := r.bit()
		v = v<<1 | uint64(b)
	}
	return v, nil
}
func (r *Reader) octets(n int) ([]byte, error) {
	r.align()
	start := r.pos / 8
	if n < 0 || n > len(r.data)-start {
		return nil, io.ErrUnexpectedEOF
	}
	r.pos += n * 8
	return append([]byte(nil), r.data[start:start+n]...), nil
}
func (r *Reader) octet() (byte, error) {
	b, e := r.octets(1)
	if e != nil {
		return 0, e
	}
	return b[0], nil
}

type Writer struct {
	data []byte
	pos  int
}

func NewWriter() *Writer { return &Writer{} }
func (w *Writer) bit(b uint8) {
	i := w.pos / 8
	if i >= len(w.data) {
		w.data = append(w.data, 0)
	}
	if b != 0 {
		w.data[i] |= 1 << uint(7-w.pos%8)
	}
	w.pos++
}
func (w *Writer) bits(v uint64, n int) {
	for i := n - 1; i >= 0; i-- {
		w.bit(uint8(v >> uint(i) & 1))
	}
}
func (w *Writer) align() {
	if m := w.pos % 8; m != 0 {
		w.bits(0, 8-m)
	}
}
func (w *Writer) octet(b byte)    { w.align(); w.data = append(w.data, b); w.pos += 8 }
func (w *Writer) octets(b []byte) { w.align(); w.data = append(w.data, b...); w.pos += 8 * len(b) }
func (w *Writer) Bytes() []byte   { w.align(); return append([]byte(nil), w.data...) }

func width(n int64) int {
	if n <= 1 {
		return 0
	}
	return bits.Len64(uint64(n - 1))
}

// octetsForWidth is ceil(bitWidth/8): the whole-octet count X.691 aligned PER
// uses for any constrained whole number field wider than 255 values.
func octetsForWidth(bitWidth int) int { return (bitWidth + 7) / 8 }

// minOctets is the minimum number of octets needed to hold off in unsigned
// big-endian form, with a floor of one octet (off == 0 still takes 1 octet).
func minOctets(off uint64) int {
	n := 1
	for off>>uint(8*n) != 0 {
		n++
	}
	return n
}

// PutConstrained writes a constrained whole number per X.691 10.5.7 (aligned
// variant). There are three regimes, keyed on the number of values the range
// spans:
//   - span <= 255: a non-aligned bit-field of the minimum width.
//   - 255 < span <= 65536: octet-aligned, a fixed whole-octet width
//     (ceil(bitWidth/8) octets, zero-extended) — not the raw bit width.
//   - span > 65536: octet-aligned, but the octet count varies with the
//     specific value. A length determinant (itself a small, therefore
//     non-aligned, constrained whole number selecting 1..maxOctets) precedes
//     the value, which is then written in exactly that many octets.
func PutConstrained(w *Writer, v, lo, hi int64) error {
	if v < lo || v > hi {
		return fmt.Errorf("aper: integer out of range")
	}
	span := hi - lo + 1
	bitWidth := width(span)
	if bitWidth == 0 {
		return nil
	}
	off := uint64(v - lo)
	if span <= 255 {
		w.bits(off, bitWidth)
		return nil
	}
	if span <= 65536 {
		w.align()
		w.bits(off, octetsForWidth(bitWidth)*8)
		return nil
	}
	maxOctets := int64(octetsForWidth(bitWidth))
	need := int64(minOctets(off))
	if err := PutConstrained(w, need, 1, maxOctets); err != nil {
		return err
	}
	w.align()
	w.bits(off, int(need)*8)
	return nil
}

// GetConstrained is the PutConstrained inverse; see its doc comment for the
// three regimes.
func GetConstrained(r *Reader, lo, hi int64) (int64, error) {
	span := hi - lo + 1
	bitWidth := width(span)
	if bitWidth == 0 {
		return lo, nil
	}
	if span <= 255 {
		v, e := r.bits(bitWidth)
		if e != nil {
			return 0, e
		}
		out := lo + int64(v)
		if out > hi {
			return 0, fmt.Errorf("aper: constrained value out of range")
		}
		return out, nil
	}
	if span <= 65536 {
		r.align()
		v, e := r.bits(octetsForWidth(bitWidth) * 8)
		if e != nil {
			return 0, e
		}
		out := lo + int64(v)
		if out > hi {
			return 0, fmt.Errorf("aper: constrained value out of range")
		}
		return out, nil
	}
	maxOctets := int64(octetsForWidth(bitWidth))
	need, e := GetConstrained(r, 1, maxOctets)
	if e != nil {
		return 0, e
	}
	r.align()
	v, e := r.bits(int(need) * 8)
	if e != nil {
		return 0, e
	}
	out := lo + int64(v)
	if out > hi {
		return 0, fmt.Errorf("aper: constrained value out of range")
	}
	return out, nil
}

// PutFixedOctets writes a fixed-size OCTET STRING. APER aligns the contents
// but emits no length determinant for a fixed size.
func PutFixedOctets(w *Writer, b []byte, size int) error {
	if len(b) != size {
		return fmt.Errorf("aper: invalid fixed octet string size")
	}
	w.octets(b)
	return nil
}

func GetFixedOctets(r *Reader, size int) ([]byte, error) {
	if size < 0 {
		return nil, fmt.Errorf("aper: invalid fixed octet string size")
	}
	return r.octets(size)
}

// PutFixedBitString writes a fixed-length BIT STRING per X.691 aligned PER:
// strings longer than 16 bits are octet-aligned before their content; 16
// bits or fewer are not aligned. value holds the bits right-justified (the
// low size bits), written most-significant-bit first.
func PutFixedBitString(w *Writer, value uint64, size int) error {
	if size < 1 || size > 64 {
		return fmt.Errorf("aper: invalid fixed bit string size")
	}
	if size > 16 {
		w.align()
	}
	w.bits(value, size)
	return nil
}

// GetFixedBitString is the PutFixedBitString inverse.
func GetFixedBitString(r *Reader, size int) (uint64, error) {
	if size < 1 || size > 64 {
		return 0, fmt.Errorf("aper: invalid fixed bit string size")
	}
	if size > 16 {
		r.align()
	}
	return r.bits(size)
}
func PutCriticality(w *Writer, c Criticality) error {
	if c > Notify {
		return fmt.Errorf("aper: invalid criticality")
	}
	w.bits(uint64(c), 2)
	return nil
}
func GetCriticality(r *Reader) (Criticality, error) {
	v, e := r.bits(2)
	if e != nil {
		return 0, e
	}
	if v > 2 {
		return 0, fmt.Errorf("aper: invalid criticality")
	}
	return Criticality(v), nil
}
func PutOpenType(w *Writer, b []byte) error {
	if len(b) > MaxOpenType {
		return fmt.Errorf("aper: open type too large")
	}
	w.align()
	n := len(b)
	if n < 128 {
		w.octet(byte(n))
	} else if n < 16384 {
		w.octet(byte(0x80 | n>>8))
		w.octet(byte(n))
	} else {
		return fmt.Errorf("aper: fragmented open types unsupported")
	}
	w.octets(b)
	return nil
}
func GetOpenType(r *Reader) ([]byte, error) {
	r.align()
	b, e := r.octet()
	if e != nil {
		return nil, e
	}
	n := int(b)
	if b&0xc0 == 0xc0 {
		return nil, fmt.Errorf("aper: fragmented open type unsupported")
	}
	if b&0x80 != 0 {
		b2, e := r.octet()
		if e != nil {
			return nil, e
		}
		n = int(b&0x3f)<<8 | int(b2)
	}
	if n > MaxOpenType {
		return nil, fmt.Errorf("aper: open type too large")
	}
	return r.octets(n)
}
