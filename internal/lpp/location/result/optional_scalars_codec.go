package result

import (
	"fmt"
	"github.com/vectorcore/esmlc/internal/uper"
)

func (v RSRPResult) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrRSRPResultEncode, err)
	}
	if err := w.WriteConstrainedWholeNumber(uint64(v), 0, 97); err != nil {
		return fmt.Errorf("%w: %w", ErrRSRPResultEncode, err)
	}
	return nil
}
func DecodeRSRPResult(r *uper.Reader) (RSRPResult, error) {
	n, e := r.ReadConstrainedWholeNumber(0, 97)
	if e != nil {
		return 0, fmt.Errorf("%w: %w", ErrRSRPResultDecode, e)
	}
	v, e := NewRSRPResult(uint16(n))
	if e != nil {
		return 0, fmt.Errorf("%w: %w", ErrRSRPResultDecode, e)
	}
	return v, nil
}
func (v RSRQResult) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrRSRQResultEncode, err)
	}
	if err := w.WriteConstrainedWholeNumber(uint64(v), 0, 34); err != nil {
		return fmt.Errorf("%w: %w", ErrRSRQResultEncode, err)
	}
	return nil
}
func DecodeRSRQResult(r *uper.Reader) (RSRQResult, error) {
	n, e := r.ReadConstrainedWholeNumber(0, 34)
	if e != nil {
		return 0, fmt.Errorf("%w: %w", ErrRSRQResultDecode, e)
	}
	v, e := NewRSRQResult(uint16(n))
	if e != nil {
		return 0, fmt.Errorf("%w: %w", ErrRSRQResultDecode, e)
	}
	return v, nil
}
func (v UERxTxTimeDiff) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrUERxTxTimeDiffEncode, err)
	}
	if err := w.WriteConstrainedWholeNumber(uint64(v), 0, 4095); err != nil {
		return fmt.Errorf("%w: %w", ErrUERxTxTimeDiffEncode, err)
	}
	return nil
}
func DecodeUERxTxTimeDiff(r *uper.Reader) (UERxTxTimeDiff, error) {
	n, e := r.ReadConstrainedWholeNumber(0, 4095)
	if e != nil {
		return 0, fmt.Errorf("%w: %w", ErrUERxTxTimeDiffDecode, e)
	}
	v, e := NewUERxTxTimeDiff(uint32(n))
	if e != nil {
		return 0, fmt.Errorf("%w: %w", ErrUERxTxTimeDiffDecode, e)
	}
	return v, nil
}
func (v SystemFrameNumber) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrSystemFrameNumberEncode, err)
	}
	if err := w.WriteBitString(v.bits, 10, 10); err != nil {
		return fmt.Errorf("%w: %w", ErrSystemFrameNumberEncode, err)
	}
	return nil
}
func DecodeSystemFrameNumber(r *uper.Reader) (SystemFrameNumber, error) {
	bits, e := r.ReadBitString(10, 10)
	if e != nil {
		return SystemFrameNumber{}, fmt.Errorf("%w: %w", ErrSystemFrameNumberDecode, e)
	}
	v, e := NewSystemFrameNumber(bits)
	if e != nil {
		return SystemFrameNumber{}, fmt.Errorf("%w: %w", ErrSystemFrameNumberDecode, e)
	}
	return v, nil
}
