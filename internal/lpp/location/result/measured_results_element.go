package result

import (
	"errors"
	"fmt"

	"github.com/vectorcore/esmlc/internal/uper"
)

var (
	ErrMeasuredResultsElementInvalid    = errors.New("lpp location result: invalid measured results element")
	ErrMeasuredResultsElementExtensions = errors.New("lpp location result: measured results element extensions unsupported")
)

// MeasuredResultsElement is the bounded root form only. Nil optional fields
// represent ASN.1 absence; non-nil pointers are copied at construction.
type MeasuredResultsElement struct {
	physicalCellID PhysicalCellID
	cellGlobalID   *CellGlobalIdEUTRAAndUTRA
	arfcnEUTRA     EUTRAARFCN
	sfn            *SystemFrameNumber
	rsrp           *RSRPResult
	rsrq           *RSRQResult
	ueRxTx         *UERxTxTimeDiff
}

type MeasuredResultsElementOptions struct {
	CellGlobalID      *CellGlobalIdEUTRAAndUTRA
	SystemFrameNumber *SystemFrameNumber
	RSRPResult        *RSRPResult
	RSRQResult        *RSRQResult
	UERxTxTimeDiff    *UERxTxTimeDiff
}

func NewMeasuredResultsElement(pci PhysicalCellID, arfcn EUTRAARFCN, opt MeasuredResultsElementOptions) (MeasuredResultsElement, error) {
	v := MeasuredResultsElement{physicalCellID: pci, arfcnEUTRA: arfcn}
	if err := pci.Validate(); err != nil {
		return v, fmt.Errorf("%w: %w", ErrMeasuredResultsElementInvalid, err)
	}
	if err := arfcn.Validate(); err != nil {
		return v, fmt.Errorf("%w: %w", ErrMeasuredResultsElementInvalid, err)
	}
	if opt.CellGlobalID != nil {
		x := *opt.CellGlobalID
		if err := x.Validate(); err != nil {
			return v, fmt.Errorf("%w: %w", ErrMeasuredResultsElementInvalid, err)
		}
		v.cellGlobalID = &x
	}
	if opt.SystemFrameNumber != nil {
		x := *opt.SystemFrameNumber
		if err := x.Validate(); err != nil {
			return v, fmt.Errorf("%w: %w", ErrMeasuredResultsElementInvalid, err)
		}
		v.sfn = &x
	}
	if opt.RSRPResult != nil {
		x := *opt.RSRPResult
		if err := x.Validate(); err != nil {
			return v, fmt.Errorf("%w: %w", ErrMeasuredResultsElementInvalid, err)
		}
		v.rsrp = &x
	}
	if opt.RSRQResult != nil {
		x := *opt.RSRQResult
		if err := x.Validate(); err != nil {
			return v, fmt.Errorf("%w: %w", ErrMeasuredResultsElementInvalid, err)
		}
		v.rsrq = &x
	}
	if opt.UERxTxTimeDiff != nil {
		x := *opt.UERxTxTimeDiff
		if err := x.Validate(); err != nil {
			return v, fmt.Errorf("%w: %w", ErrMeasuredResultsElementInvalid, err)
		}
		v.ueRxTx = &x
	}
	return v, nil
}
func (v MeasuredResultsElement) PhysicalCellID() PhysicalCellID { return v.physicalCellID }
func (v MeasuredResultsElement) EUTRAARFCN() EUTRAARFCN         { return v.arfcnEUTRA }
func (v MeasuredResultsElement) CellGlobalID() (CellGlobalIdEUTRAAndUTRA, bool) {
	if v.cellGlobalID == nil {
		return CellGlobalIdEUTRAAndUTRA{}, false
	}
	return *v.cellGlobalID, true
}
func (v MeasuredResultsElement) SystemFrameNumber() (SystemFrameNumber, bool) {
	if v.sfn == nil {
		return SystemFrameNumber{}, false
	}
	return *v.sfn, true
}
func (v MeasuredResultsElement) RSRPResult() (RSRPResult, bool) {
	if v.rsrp == nil {
		return 0, false
	}
	return *v.rsrp, true
}
func (v MeasuredResultsElement) RSRQResult() (RSRQResult, bool) {
	if v.rsrq == nil {
		return 0, false
	}
	return *v.rsrq, true
}
func (v MeasuredResultsElement) UERxTxTimeDiff() (UERxTxTimeDiff, bool) {
	if v.ueRxTx == nil {
		return 0, false
	}
	return *v.ueRxTx, true
}
func (v MeasuredResultsElement) Validate() error {
	_, e := NewMeasuredResultsElement(v.physicalCellID, v.arfcnEUTRA, MeasuredResultsElementOptions{v.cellGlobalID, v.sfn, v.rsrp, v.rsrq, v.ueRxTx})
	return e
}
func (v MeasuredResultsElement) EncodeUPER(w *uper.Writer) error {
	if e := v.Validate(); e != nil {
		return e
	}
	if e := w.WriteExtensionPresent(false); e != nil {
		return e
	}
	if e := w.WriteOptionalBitmap([]bool{v.cellGlobalID != nil, v.sfn != nil, v.rsrp != nil, v.rsrq != nil, v.ueRxTx != nil}); e != nil {
		return e
	}
	if e := v.physicalCellID.EncodeUPER(w); e != nil {
		return e
	}
	if v.cellGlobalID != nil {
		if e := v.cellGlobalID.EncodeUPER(w); e != nil {
			return e
		}
	}
	if e := v.arfcnEUTRA.EncodeUPER(w); e != nil {
		return e
	}
	if v.sfn != nil {
		if e := v.sfn.EncodeUPER(w); e != nil {
			return e
		}
	}
	if v.rsrp != nil {
		if e := v.rsrp.EncodeUPER(w); e != nil {
			return e
		}
	}
	if v.rsrq != nil {
		if e := v.rsrq.EncodeUPER(w); e != nil {
			return e
		}
	}
	if v.ueRxTx != nil {
		return v.ueRxTx.EncodeUPER(w)
	}
	return nil
}
func DecodeMeasuredResultsElement(r *uper.Reader) (MeasuredResultsElement, error) {
	ext, e := r.ReadExtensionPresent()
	if e != nil {
		return MeasuredResultsElement{}, e
	}
	if ext {
		return MeasuredResultsElement{}, ErrMeasuredResultsElementExtensions
	}
	p, e := r.ReadOptionalBitmap(5)
	if e != nil {
		return MeasuredResultsElement{}, e
	}
	pci, e := DecodePhysicalCellID(r)
	if e != nil {
		return MeasuredResultsElement{}, e
	}
	o := MeasuredResultsElementOptions{}
	if p[0] {
		x, e := DecodeCellGlobalIdEUTRAAndUTRA(r)
		if e != nil {
			return MeasuredResultsElement{}, e
		}
		o.CellGlobalID = &x
	}
	a, e := DecodeEUTRAARFCN(r)
	if e != nil {
		return MeasuredResultsElement{}, e
	}
	if p[1] {
		x, e := DecodeSystemFrameNumber(r)
		if e != nil {
			return MeasuredResultsElement{}, e
		}
		o.SystemFrameNumber = &x
	}
	if p[2] {
		x, e := DecodeRSRPResult(r)
		if e != nil {
			return MeasuredResultsElement{}, e
		}
		o.RSRPResult = &x
	}
	if p[3] {
		x, e := DecodeRSRQResult(r)
		if e != nil {
			return MeasuredResultsElement{}, e
		}
		o.RSRQResult = &x
	}
	if p[4] {
		x, e := DecodeUERxTxTimeDiff(r)
		if e != nil {
			return MeasuredResultsElement{}, e
		}
		o.UERxTxTimeDiff = &x
	}
	return NewMeasuredResultsElement(pci, a, o)
}
