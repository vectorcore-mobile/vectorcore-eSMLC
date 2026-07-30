package result

import (
	"fmt"

	"github.com/vectorcore/esmlc/internal/uper"
)

type RSRPResult uint8

const (
	MinRSRPResult RSRPResult = 0
	MaxRSRPResult RSRPResult = 97
)

func NewRSRPResult(v uint16) (RSRPResult, error) {
	if v > uint16(MaxRSRPResult) {
		return 0, fmt.Errorf("%w: %d outside 0..97", ErrRSRPResultOutOfRange, v)
	}
	return RSRPResult(v), nil
}
func (v RSRPResult) Validate() error {
	if v > MaxRSRPResult {
		return fmt.Errorf("%w: %d outside 0..97", ErrRSRPResultOutOfRange, v)
	}
	return nil
}

type RSRQResult uint8

const (
	MinRSRQResult RSRQResult = 0
	MaxRSRQResult RSRQResult = 34
)

func NewRSRQResult(v uint16) (RSRQResult, error) {
	if v > uint16(MaxRSRQResult) {
		return 0, fmt.Errorf("%w: %d outside 0..34", ErrRSRQResultOutOfRange, v)
	}
	return RSRQResult(v), nil
}
func (v RSRQResult) Validate() error {
	if v > MaxRSRQResult {
		return fmt.Errorf("%w: %d outside 0..34", ErrRSRQResultOutOfRange, v)
	}
	return nil
}

type UERxTxTimeDiff uint16

const (
	MinUERxTxTimeDiff UERxTxTimeDiff = 0
	MaxUERxTxTimeDiff UERxTxTimeDiff = 4095
)

func NewUERxTxTimeDiff(v uint32) (UERxTxTimeDiff, error) {
	if v > uint32(MaxUERxTxTimeDiff) {
		return 0, fmt.Errorf("%w: %d outside 0..4095", ErrUERxTxTimeDiffOutOfRange, v)
	}
	return UERxTxTimeDiff(v), nil
}
func (v UERxTxTimeDiff) Validate() error {
	if v > MaxUERxTxTimeDiff {
		return fmt.Errorf("%w: %d outside 0..4095", ErrUERxTxTimeDiffOutOfRange, v)
	}
	return nil
}

// SystemFrameNumber preserves the fixed ten-bit ASN.1 BIT STRING exactly.
// Bit zero on the wire is the most-significant bit of Uint16's ten-bit value.
type SystemFrameNumber struct{ bits uper.BitString }

func NewSystemFrameNumber(bits uper.BitString) (SystemFrameNumber, error) {
	if bits.BitLen() != 10 {
		return SystemFrameNumber{}, fmt.Errorf("%w: got %d", ErrSystemFrameNumberInvalidBitLength, bits.BitLen())
	}
	return SystemFrameNumber{bits: bits}, nil
}
func NewSystemFrameNumberFromUint16(v uint16) (SystemFrameNumber, error) {
	if v > 1023 {
		return SystemFrameNumber{}, fmt.Errorf("%w: %d outside 0..1023", ErrSystemFrameNumberOutOfRange, v)
	}
	bits, err := uper.NewBitString([]byte{byte(v >> 2), byte(v << 6)}, 10)
	if err != nil {
		return SystemFrameNumber{}, fmt.Errorf("%w: %w", ErrSystemFrameNumberInvalidBitLength, err)
	}
	return NewSystemFrameNumber(bits)
}
func (v SystemFrameNumber) Validate() error {
	if v.bits.BitLen() != 10 {
		return fmt.Errorf("%w: got %d", ErrSystemFrameNumberInvalidBitLength, v.bits.BitLen())
	}
	return nil
}
func (v SystemFrameNumber) BitString() uper.BitString { return v.bits }
func (v SystemFrameNumber) Uint16() uint16 {
	b := v.bits.Bytes()
	if len(b) != 2 {
		return 0
	}
	return uint16(b[0])<<2 | uint16(b[1])>>6
}
