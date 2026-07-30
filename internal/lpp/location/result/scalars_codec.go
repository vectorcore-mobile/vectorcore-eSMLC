package result

import (
	"fmt"

	"github.com/vectorcore/esmlc/internal/uper"
)

func (v PhysicalCellID) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrPhysicalCellIDEncode, err)
	}
	if err := w.WriteConstrainedWholeNumber(uint64(v), uint64(MinPhysicalCellID), uint64(MaxPhysicalCellID)); err != nil {
		return fmt.Errorf("%w: %w", ErrPhysicalCellIDEncode, err)
	}
	return nil
}

func DecodePhysicalCellID(r *uper.Reader) (PhysicalCellID, error) {
	value, err := r.ReadConstrainedWholeNumber(uint64(MinPhysicalCellID), uint64(MaxPhysicalCellID))
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrPhysicalCellIDDecode, err)
	}
	v, err := NewPhysicalCellID(uint16(value))
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrPhysicalCellIDDecode, err)
	}
	return v, nil
}

func (v EUTRAARFCN) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrEUTRAARFCNEncode, err)
	}
	if err := w.WriteConstrainedWholeNumber(uint64(v), uint64(MinEUTRAARFCN), uint64(MaxEUTRAARFCN)); err != nil {
		return fmt.Errorf("%w: %w", ErrEUTRAARFCNEncode, err)
	}
	return nil
}

func DecodeEUTRAARFCN(r *uper.Reader) (EUTRAARFCN, error) {
	value, err := r.ReadConstrainedWholeNumber(uint64(MinEUTRAARFCN), uint64(MaxEUTRAARFCN))
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrEUTRAARFCNDecode, err)
	}
	v, err := NewEUTRAARFCN(uint32(value))
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrEUTRAARFCNDecode, err)
	}
	return v, nil
}
