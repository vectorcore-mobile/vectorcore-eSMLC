package lppa

import (
	"fmt"

	"github.com/vectorcore/esmlc/internal/aper"
)

// CauseBranch is the root Cause CHOICE alternative.
type CauseBranch uint8

const (
	CauseRadioNetwork CauseBranch = iota
	CauseProtocol
	CauseMisc
)

// causeMaxima is the highest root enumerated index for each branch:
// CauseRadioNetwork has 3 root values, CauseProtocol has 7, CauseMisc has 1.
var causeMaxima = [3]int64{2, 6, 0}

// Cause is the supported root subset of the extensible LPPa Cause CHOICE.
type Cause struct {
	Branch CauseBranch
	Value  uint8
}

func EncodeCause(c Cause) ([]byte, error) {
	if c.Branch > CauseMisc || int64(c.Value) > causeMaxima[c.Branch] {
		return nil, fmt.Errorf("lppa: unsupported cause")
	}
	w := aper.NewWriter()
	if err := extBit(w, 0); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, int64(c.Branch), 0, 2); err != nil {
		return nil, err
	}
	if err := extBit(w, 0); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, int64(c.Value), 0, causeMaxima[c.Branch]); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func DecodeCause(b []byte) (Cause, error) {
	r := aper.NewReader(b)
	ext, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return Cause{}, err
	}
	if ext != 0 {
		return Cause{}, fmt.Errorf("lppa: cause extension unsupported")
	}
	branch, err := aper.GetConstrained(r, 0, 2)
	if err != nil {
		return Cause{}, err
	}
	ext, err = aper.GetConstrained(r, 0, 1)
	if err != nil {
		return Cause{}, err
	}
	if ext != 0 {
		return Cause{}, fmt.Errorf("lppa: cause value extension unsupported")
	}
	value, err := aper.GetConstrained(r, 0, causeMaxima[branch])
	if err != nil {
		return Cause{}, err
	}
	if r.Remaining() > 7 || !r.RemainingZero() {
		return Cause{}, fmt.Errorf("lppa: trailing cause data")
	}
	return Cause{CauseBranch(branch), uint8(value)}, nil
}
