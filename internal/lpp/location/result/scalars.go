package result

import "fmt"

// PhysicalCellID is the Release 9 ECID physCellId root value.
type PhysicalCellID uint16

const (
	MinPhysicalCellID PhysicalCellID = 0
	MaxPhysicalCellID PhysicalCellID = 503
)

func NewPhysicalCellID(value uint16) (PhysicalCellID, error) {
	v := PhysicalCellID(value)
	if err := v.Validate(); err != nil {
		return 0, err
	}
	return v, nil
}

func (v PhysicalCellID) Validate() error {
	if v > MaxPhysicalCellID {
		return fmt.Errorf("%w: %d outside 0..503", ErrPhysicalCellIDOutOfRange, v)
	}
	return nil
}

// EUTRAARFCN is the non-extensible root ARFCN-ValueEUTRA value used by the
// validated Release 9 ECID path. Extension values are not representable here.
type EUTRAARFCN uint16

const (
	MinEUTRAARFCN EUTRAARFCN = 0
	MaxEUTRAARFCN EUTRAARFCN = 65535
)

// NewEUTRAARFCN validates the wider caller value before converting it, so a
// value outside the root domain cannot silently wrap into uint16.
func NewEUTRAARFCN(value uint32) (EUTRAARFCN, error) {
	if value > uint32(MaxEUTRAARFCN) {
		return 0, fmt.Errorf("%w: %d outside 0..65535", ErrEUTRAARFCNOutOfRange, value)
	}
	return EUTRAARFCN(value), nil
}

func (v EUTRAARFCN) Validate() error {
	// uint16 is exactly the complete root domain. Keep this method so callers
	// have the same validation contract as PhysicalCellID.
	return nil
}
