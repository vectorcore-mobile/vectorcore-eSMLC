package location

import (
	"fmt"

	"github.com/vectorcore/esmlc/internal/lpp/location/result"
	"github.com/vectorcore/esmlc/internal/uper"
)

// NeighbourMeasurementElement is the bounded root TS 37.355
// NeighbourMeasurementElement: one neighbour cell's RSTD measurement
// relative to the RSTD reference cell. Release 14/15 extension additions
// (transmission-point IDs, delta-rstd, additional paths, delta-SFN) are not
// represented.
type NeighbourMeasurementElement struct {
	physCellID   result.PhysicalCellID
	cellGlobalID *result.ECGI
	earfcn       *result.EUTRAARFCN
	rstd         result.RSTD
	rstdQuality  result.OTDOAMeasQuality
}

func NewNeighbourMeasurementElement(pci result.PhysicalCellID, cellGlobalID *result.ECGI, earfcn *result.EUTRAARFCN, rstd result.RSTD, quality result.OTDOAMeasQuality) (NeighbourMeasurementElement, error) {
	v := NeighbourMeasurementElement{physCellID: pci, rstd: rstd, rstdQuality: quality}
	if cellGlobalID != nil {
		x := *cellGlobalID
		v.cellGlobalID = &x
	}
	if earfcn != nil {
		x := *earfcn
		v.earfcn = &x
	}
	if err := v.Validate(); err != nil {
		return NeighbourMeasurementElement{}, err
	}
	return v, nil
}
func (v NeighbourMeasurementElement) PhysicalCellID() result.PhysicalCellID { return v.physCellID }
func (v NeighbourMeasurementElement) CellGlobalID() (result.ECGI, bool) {
	if v.cellGlobalID == nil {
		return result.ECGI{}, false
	}
	return *v.cellGlobalID, true
}
func (v NeighbourMeasurementElement) EARFCN() (result.EUTRAARFCN, bool) {
	if v.earfcn == nil {
		return 0, false
	}
	return *v.earfcn, true
}
func (v NeighbourMeasurementElement) RSTD() result.RSTD                    { return v.rstd }
func (v NeighbourMeasurementElement) RSTDQuality() result.OTDOAMeasQuality { return v.rstdQuality }
func (v NeighbourMeasurementElement) Validate() error {
	if err := v.physCellID.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidOTDOAProvide, err)
	}
	if v.cellGlobalID != nil {
		if err := v.cellGlobalID.Validate(); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidOTDOAProvide, err)
		}
	}
	if err := v.rstd.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidOTDOAProvide, err)
	}
	if err := v.rstdQuality.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidOTDOAProvide, err)
	}
	return nil
}
func (v NeighbourMeasurementElement) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{v.cellGlobalID != nil, v.earfcn != nil}); err != nil {
		return err
	}
	if err := v.physCellID.EncodeUPER(w); err != nil {
		return err
	}
	if v.cellGlobalID != nil {
		if err := v.cellGlobalID.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.earfcn != nil {
		if err := v.earfcn.EncodeUPER(w); err != nil {
			return err
		}
	}
	if err := v.rstd.EncodeUPER(w); err != nil {
		return err
	}
	return v.rstdQuality.EncodeUPER(w)
}
func DecodeNeighbourMeasurementElement(r *uper.Reader) (NeighbourMeasurementElement, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return NeighbourMeasurementElement{}, err
	}
	if ext {
		return NeighbourMeasurementElement{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(2)
	if err != nil {
		return NeighbourMeasurementElement{}, err
	}
	pci, err := result.DecodePhysicalCellID(r)
	if err != nil {
		return NeighbourMeasurementElement{}, err
	}
	var cellGlobalID *result.ECGI
	if present[0] {
		x, err := result.DecodeECGI(r)
		if err != nil {
			return NeighbourMeasurementElement{}, err
		}
		cellGlobalID = &x
	}
	var earfcn *result.EUTRAARFCN
	if present[1] {
		x, err := result.DecodeEUTRAARFCN(r)
		if err != nil {
			return NeighbourMeasurementElement{}, err
		}
		earfcn = &x
	}
	rstd, err := result.DecodeRSTD(r)
	if err != nil {
		return NeighbourMeasurementElement{}, err
	}
	quality, err := result.DecodeOTDOAMeasQuality(r)
	if err != nil {
		return NeighbourMeasurementElement{}, err
	}
	return NewNeighbourMeasurementElement(pci, cellGlobalID, earfcn, rstd, quality)
}

// OTDOASignalMeasurementInformation is the bounded root TS 37.355
// OTDOA-SignalMeasurementInformation: the RSTD reference cell identity plus
// 1..24 neighbour cell RSTD measurements. Release 14/15 extension additions
// (transmission point/PRS IDs, additional paths, hyper-SFN, motion time
// source) are not represented.
type OTDOASignalMeasurementInformation struct {
	systemFrameNumber result.SystemFrameNumber
	physCellIDRef     result.PhysicalCellID
	cellGlobalIDRef   *result.ECGI
	earfcnRef         *result.EUTRAARFCN
	referenceQuality  *result.OTDOAMeasQuality
	neighbours        []NeighbourMeasurementElement
}

func NewOTDOASignalMeasurementInformation(sfn result.SystemFrameNumber, physCellIDRef result.PhysicalCellID, cellGlobalIDRef *result.ECGI, earfcnRef *result.EUTRAARFCN, referenceQuality *result.OTDOAMeasQuality, neighbours []NeighbourMeasurementElement) (OTDOASignalMeasurementInformation, error) {
	v := OTDOASignalMeasurementInformation{systemFrameNumber: sfn, physCellIDRef: physCellIDRef, neighbours: append([]NeighbourMeasurementElement(nil), neighbours...)}
	if cellGlobalIDRef != nil {
		x := *cellGlobalIDRef
		v.cellGlobalIDRef = &x
	}
	if earfcnRef != nil {
		x := *earfcnRef
		v.earfcnRef = &x
	}
	if referenceQuality != nil {
		x := *referenceQuality
		v.referenceQuality = &x
	}
	if err := v.Validate(); err != nil {
		return OTDOASignalMeasurementInformation{}, err
	}
	return v, nil
}
func (v OTDOASignalMeasurementInformation) SystemFrameNumber() result.SystemFrameNumber {
	return v.systemFrameNumber
}
func (v OTDOASignalMeasurementInformation) PhysCellIDRef() result.PhysicalCellID {
	return v.physCellIDRef
}
func (v OTDOASignalMeasurementInformation) CellGlobalIDRef() (result.ECGI, bool) {
	if v.cellGlobalIDRef == nil {
		return result.ECGI{}, false
	}
	return *v.cellGlobalIDRef, true
}
func (v OTDOASignalMeasurementInformation) EARFCNRef() (result.EUTRAARFCN, bool) {
	if v.earfcnRef == nil {
		return 0, false
	}
	return *v.earfcnRef, true
}
func (v OTDOASignalMeasurementInformation) ReferenceQuality() (result.OTDOAMeasQuality, bool) {
	if v.referenceQuality == nil {
		return result.OTDOAMeasQuality{}, false
	}
	return *v.referenceQuality, true
}
func (v OTDOASignalMeasurementInformation) NeighbourMeasurements() []NeighbourMeasurementElement {
	return append([]NeighbourMeasurementElement(nil), v.neighbours...)
}
func (v OTDOASignalMeasurementInformation) clone() OTDOASignalMeasurementInformation {
	out := v
	out.neighbours = append([]NeighbourMeasurementElement(nil), v.neighbours...)
	return out
}
func (v OTDOASignalMeasurementInformation) Validate() error {
	if err := v.systemFrameNumber.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidOTDOAProvide, err)
	}
	if err := v.physCellIDRef.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidOTDOAProvide, err)
	}
	if len(v.neighbours) < 1 || len(v.neighbours) > 24 {
		return fmt.Errorf("%w: count %d outside 1..24", ErrMissingNeighbourMeasurements, len(v.neighbours))
	}
	for i := range v.neighbours {
		if err := v.neighbours[i].Validate(); err != nil {
			return fmt.Errorf("%w: neighbour %d: %w", ErrInvalidOTDOAProvide, i, err)
		}
	}
	return nil
}
func (v OTDOASignalMeasurementInformation) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{v.cellGlobalIDRef != nil, v.earfcnRef != nil, v.referenceQuality != nil}); err != nil {
		return err
	}
	if err := v.systemFrameNumber.EncodeUPER(w); err != nil {
		return err
	}
	if err := v.physCellIDRef.EncodeUPER(w); err != nil {
		return err
	}
	if v.cellGlobalIDRef != nil {
		if err := v.cellGlobalIDRef.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.earfcnRef != nil {
		if err := v.earfcnRef.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.referenceQuality != nil {
		if err := v.referenceQuality.EncodeUPER(w); err != nil {
			return err
		}
	}
	return w.WriteSequenceOf(len(v.neighbours), 1, 24, func(index int, writer *uper.Writer) error {
		return v.neighbours[index].EncodeUPER(writer)
	})
}
func DecodeOTDOASignalMeasurementInformation(r *uper.Reader) (OTDOASignalMeasurementInformation, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return OTDOASignalMeasurementInformation{}, err
	}
	if ext {
		return OTDOASignalMeasurementInformation{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(3)
	if err != nil {
		return OTDOASignalMeasurementInformation{}, err
	}
	sfn, err := result.DecodeSystemFrameNumber(r)
	if err != nil {
		return OTDOASignalMeasurementInformation{}, err
	}
	physCellIDRef, err := result.DecodePhysicalCellID(r)
	if err != nil {
		return OTDOASignalMeasurementInformation{}, err
	}
	var cellGlobalIDRef *result.ECGI
	if present[0] {
		x, err := result.DecodeECGI(r)
		if err != nil {
			return OTDOASignalMeasurementInformation{}, err
		}
		cellGlobalIDRef = &x
	}
	var earfcnRef *result.EUTRAARFCN
	if present[1] {
		x, err := result.DecodeEUTRAARFCN(r)
		if err != nil {
			return OTDOASignalMeasurementInformation{}, err
		}
		earfcnRef = &x
	}
	var referenceQuality *result.OTDOAMeasQuality
	if present[2] {
		x, err := result.DecodeOTDOAMeasQuality(r)
		if err != nil {
			return OTDOASignalMeasurementInformation{}, err
		}
		referenceQuality = &x
	}
	neighbours := make([]NeighbourMeasurementElement, 0, 24)
	_, err = r.ReadSequenceOf(1, 24, func(index int, reader *uper.Reader) error {
		x, err := DecodeNeighbourMeasurementElement(reader)
		if err != nil {
			return err
		}
		neighbours = append(neighbours, x)
		return nil
	})
	if err != nil {
		return OTDOASignalMeasurementInformation{}, err
	}
	return NewOTDOASignalMeasurementInformation(sfn, physCellIDRef, cellGlobalIDRef, earfcnRef, referenceQuality, neighbours)
}

// OTDOAErrorSource distinguishes the two OTDOA-Error CHOICE branches.
type OTDOAErrorSource uint8

const (
	OTDOAErrorLocationServer OTDOAErrorSource = iota
	OTDOAErrorTargetDevice
)

type OTDOALocationServerErrorCause uint8

const (
	OTDOALocationServerCauseUndefined OTDOALocationServerErrorCause = iota
	OTDOALocationServerCauseAssistanceDataNotSupportedByServer
	OTDOALocationServerCauseAssistanceDataCurrentlyUnavailable
)

type OTDOATargetDeviceErrorCause uint8

const (
	OTDOATargetDeviceCauseUndefined OTDOATargetDeviceErrorCause = iota
	OTDOATargetDeviceCauseAssistanceDataMissing
	OTDOATargetDeviceCauseUnableToMeasureReferenceCell
	OTDOATargetDeviceCauseUnableToMeasureAnyNeighbourCell
	OTDOATargetDeviceCauseAttemptedButUnableToMeasureSomeNeighbourCells
)

// OTDOAError is the bounded root TS 37.355 OTDOA-Error CHOICE.
type OTDOAError struct {
	Source              OTDOAErrorSource
	LocationServerCause OTDOALocationServerErrorCause
	TargetDeviceCause   OTDOATargetDeviceErrorCause
}

func (v OTDOAError) Validate() error {
	switch v.Source {
	case OTDOAErrorLocationServer:
		if v.LocationServerCause > OTDOALocationServerCauseAssistanceDataCurrentlyUnavailable {
			return fmt.Errorf("%w: invalid location server cause", ErrInvalidOTDOAError)
		}
	case OTDOAErrorTargetDevice:
		if v.TargetDeviceCause > OTDOATargetDeviceCauseAttemptedButUnableToMeasureSomeNeighbourCells {
			return fmt.Errorf("%w: invalid target device cause", ErrInvalidOTDOAError)
		}
	default:
		return fmt.Errorf("%w: invalid source", ErrInvalidOTDOAError)
	}
	return nil
}
func (v OTDOAError) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteRootChoiceIndex(uint64(v.Source), 2); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil { // outer SEQUENCE: extensible, no optional members
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil { // cause ENUMERATED's own extension marker
		return err
	}
	if v.Source == OTDOAErrorLocationServer {
		return w.WriteRootEnumerated(uint64(v.LocationServerCause), 3)
	}
	return w.WriteRootEnumerated(uint64(v.TargetDeviceCause), 5)
}
func DecodeOTDOAError(r *uper.Reader) (OTDOAError, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return OTDOAError{}, err
	}
	if ext {
		return OTDOAError{}, ErrUnsupportedExtension
	}
	idx, err := r.ReadRootChoiceIndex(2)
	if err != nil {
		return OTDOAError{}, err
	}
	seqExt, err := r.ReadExtensionPresent()
	if err != nil {
		return OTDOAError{}, err
	}
	if seqExt {
		return OTDOAError{}, ErrUnsupportedExtension
	}
	causeExt, err := r.ReadExtensionPresent()
	if err != nil {
		return OTDOAError{}, err
	}
	if causeExt {
		return OTDOAError{}, ErrUnsupportedExtension
	}
	if idx == uint64(OTDOAErrorLocationServer) {
		c, err := r.ReadRootEnumerated(3)
		if err != nil {
			return OTDOAError{}, err
		}
		return OTDOAError{Source: OTDOAErrorLocationServer, LocationServerCause: OTDOALocationServerErrorCause(c)}, nil
	}
	c, err := r.ReadRootEnumerated(5)
	if err != nil {
		return OTDOAError{}, err
	}
	return OTDOAError{Source: OTDOAErrorTargetDevice, TargetDeviceCause: OTDOATargetDeviceErrorCause(c)}, nil
}

// OTDOAProvideLocationInformation is the bounded root TS 37.355
// OTDOA-ProvideLocationInformation: an optional measurement report and an
// optional error, mutually informative rather than mutually exclusive per
// the ASN.1 (both OPTIONAL, independently). The NB-IoT extension addition is
// not represented.
type OTDOAProvideLocationInformation struct {
	signal *OTDOASignalMeasurementInformation
	error  *OTDOAError
}

func NewOTDOAProvideLocationInformation(signal *OTDOASignalMeasurementInformation, err *OTDOAError) (OTDOAProvideLocationInformation, error) {
	v := OTDOAProvideLocationInformation{}
	if signal != nil {
		x := *signal
		v.signal = &x
	}
	if err != nil {
		x := *err
		v.error = &x
	}
	if e := v.Validate(); e != nil {
		return OTDOAProvideLocationInformation{}, e
	}
	return v, nil
}
func (v OTDOAProvideLocationInformation) SignalMeasurementInformation() (OTDOASignalMeasurementInformation, bool) {
	if v.signal == nil {
		return OTDOASignalMeasurementInformation{}, false
	}
	return *v.signal, true
}
func (v OTDOAProvideLocationInformation) Error() (OTDOAError, bool) {
	if v.error == nil {
		return OTDOAError{}, false
	}
	return *v.error, true
}
func (v OTDOAProvideLocationInformation) clone() OTDOAProvideLocationInformation {
	out := v
	if v.signal != nil {
		s := v.signal.clone()
		out.signal = &s
	}
	if v.error != nil {
		e := *v.error
		out.error = &e
	}
	return out
}
func (v OTDOAProvideLocationInformation) Validate() error {
	if v.signal != nil {
		if err := v.signal.Validate(); err != nil {
			return err
		}
	}
	if v.error != nil {
		if err := v.error.Validate(); err != nil {
			return err
		}
	}
	return nil
}
func (v OTDOAProvideLocationInformation) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{v.signal != nil, v.error != nil}); err != nil {
		return err
	}
	if v.signal != nil {
		if err := v.signal.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.error != nil {
		return v.error.EncodeUPER(w)
	}
	return nil
}
func DecodeOTDOAProvideLocationInformation(r *uper.Reader) (OTDOAProvideLocationInformation, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return OTDOAProvideLocationInformation{}, err
	}
	if ext {
		return OTDOAProvideLocationInformation{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(2)
	if err != nil {
		return OTDOAProvideLocationInformation{}, err
	}
	var signal *OTDOASignalMeasurementInformation
	if present[0] {
		x, err := DecodeOTDOASignalMeasurementInformation(r)
		if err != nil {
			return OTDOAProvideLocationInformation{}, err
		}
		signal = &x
	}
	var errv *OTDOAError
	if present[1] {
		x, err := DecodeOTDOAError(r)
		if err != nil {
			return OTDOAProvideLocationInformation{}, err
		}
		errv = &x
	}
	return NewOTDOAProvideLocationInformation(signal, errv)
}
