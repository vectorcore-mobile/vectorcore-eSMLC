package location

import (
	"errors"
	"fmt"

	"github.com/vectorcore/esmlc/internal/uper"
)

var (
	ErrInvalidAGNSSProvide               = errors.New("lpp location: invalid A-GNSS provide")
	ErrUnsupportedGNSSSignalMeasurements = errors.New("lpp location: GNSS-SignalMeasurementInformation (MS-assisted) unsupported")
	ErrInvalidAGNSSError                 = errors.New("lpp location: invalid A-GNSS error")
)

// writeGNSSIDGPSOnly and readGNSSIDRequireGPS encode/decode TS 37.355
// GNSS-ID, an extensible SEQUENCE wrapping an extensible root ENUMERATED
// {gps,sbas,qzss,galileo,glonass,...,bds,navic-v1610}. This implementation
// is GPS-only scope: encode always writes gps (root index 0); decode
// requires it and fails closed on anything else, matching the bounded-root
// discipline used throughout this package.
func writeGNSSIDGPSOnly(w *uper.Writer) error {
	if err := w.WriteExtensionPresent(false); err != nil { // GNSS-ID SEQUENCE
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil { // gnss-id ENUMERATED
		return err
	}
	return w.WriteRootEnumerated(0, 5)
}
func readGNSSIDRequireGPS(r *uper.Reader) error {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return err
	}
	if ext {
		return ErrUnsupportedExtension
	}
	idExt, err := r.ReadExtensionPresent()
	if err != nil {
		return err
	}
	if idExt {
		return ErrUnsupportedExtension
	}
	id, err := r.ReadRootEnumerated(5)
	if err != nil {
		return err
	}
	if id != 0 {
		return fmt.Errorf("%w: gnss-id %d, only gps (0) is supported", ErrInvalidAGNSSProvide, id)
	}
	return nil
}

// MeasurementReferenceTime is the bounded root TS 37.355
// MeasurementReferenceTime: gnss-TOD-msec and a required-GPS gnss-TimeID.
// gnss-TOD-frac, gnss-TOD-unc, and networkTime are all OPTIONAL in the root
// and are not represented here.
type MeasurementReferenceTime struct {
	GNSSTODMsec uint32 // 0..3599999
}

func (v MeasurementReferenceTime) Validate() error {
	if v.GNSSTODMsec > 3599999 {
		return fmt.Errorf("%w: gnss-TOD-msec %d out of range", ErrInvalidAGNSSProvide, v.GNSSTODMsec)
	}
	return nil
}
func (v MeasurementReferenceTime) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{false, false, false}); err != nil {
		return err
	}
	if err := w.WriteConstrainedWholeNumber(uint64(v.GNSSTODMsec), 0, 3599999); err != nil {
		return err
	}
	return writeGNSSIDGPSOnly(w)
}
func DecodeMeasurementReferenceTime(r *uper.Reader) (MeasurementReferenceTime, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return MeasurementReferenceTime{}, err
	}
	if ext {
		return MeasurementReferenceTime{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(3)
	if err != nil {
		return MeasurementReferenceTime{}, err
	}
	msec, err := r.ReadConstrainedWholeNumber(0, 3599999)
	if err != nil {
		return MeasurementReferenceTime{}, err
	}
	if present[0] {
		if _, err := r.ReadConstrainedWholeNumber(0, 3999); err != nil {
			return MeasurementReferenceTime{}, err
		}
		return MeasurementReferenceTime{}, fmt.Errorf("%w: gnss-TOD-frac unsupported", ErrInvalidAGNSSProvide)
	}
	if present[1] {
		return MeasurementReferenceTime{}, fmt.Errorf("%w: gnss-TOD-unc unsupported", ErrInvalidAGNSSProvide)
	}
	if err := readGNSSIDRequireGPS(r); err != nil {
		return MeasurementReferenceTime{}, err
	}
	if present[2] {
		return MeasurementReferenceTime{}, fmt.Errorf("%w: networkTime unsupported", ErrInvalidAGNSSProvide)
	}
	v := MeasurementReferenceTime{GNSSTODMsec: uint32(msec)}
	return v, v.Validate()
}

// GNSSLocationInformation is the bounded root TS 37.355
// GNSS-LocationInformation: reference time plus which GNSS systems
// contributed. It is supplementary metadata about a UE-computed fix, not
// the fix itself — the actual coordinates travel in
// CommonProvideLocationInformation.LocationEstimate (see common.go).
type GNSSLocationInformation struct {
	MeasurementReferenceTime MeasurementReferenceTime
	AGNSSList                uper.BitString // GNSS-ID-Bitmap; only bit 0 (GPS) is meaningful here
}

func (v GNSSLocationInformation) Validate() error {
	if err := v.MeasurementReferenceTime.Validate(); err != nil {
		return err
	}
	n := v.AGNSSList.BitLen()
	if n < 1 || n > 16 {
		return fmt.Errorf("%w: agnss-List length %d", ErrInvalidAGNSSProvide, n)
	}
	return nil
}
func (v GNSSLocationInformation) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := v.MeasurementReferenceTime.EncodeUPER(w); err != nil {
		return err
	}
	return writeGNSSIDBitmap(w, v.AGNSSList)
}
func DecodeGNSSLocationInformation(r *uper.Reader) (GNSSLocationInformation, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSLocationInformation{}, err
	}
	if ext {
		return GNSSLocationInformation{}, ErrUnsupportedExtension
	}
	ref, err := DecodeMeasurementReferenceTime(r)
	if err != nil {
		return GNSSLocationInformation{}, err
	}
	list, err := readGNSSIDBitmap(r)
	if err != nil {
		return GNSSLocationInformation{}, err
	}
	v := GNSSLocationInformation{MeasurementReferenceTime: ref, AGNSSList: list}
	return v, v.Validate()
}

// AGNSSErrorSource distinguishes the two A-GNSS-Error CHOICE branches.
type AGNSSErrorSource uint8

const (
	AGNSSErrorLocationServer AGNSSErrorSource = iota
	AGNSSErrorTargetDevice
)

type AGNSSLocationServerErrorCause uint8

const (
	AGNSSLocationServerCauseUndefined AGNSSLocationServerErrorCause = iota
	AGNSSLocationServerCauseAssistanceDataNotSupported
	AGNSSLocationServerCauseAssistanceDataCurrentlyUnavailable
	AGNSSLocationServerCauseAssistanceDataPartlyUnavailable
)

type AGNSSTargetDeviceErrorCause uint8

const (
	AGNSSTargetDeviceCauseUndefined AGNSSTargetDeviceErrorCause = iota
	AGNSSTargetDeviceCauseNotEnoughSatellites
	AGNSSTargetDeviceCauseAssistanceDataMissing
	AGNSSTargetDeviceCauseNotAllRequestedMeasurementsPossible
)

// AGNSSError is the bounded root TS 37.355 A-GNSS-Error CHOICE. The three
// OPTIONAL NULL diagnostic fields on the target-device branch
// (fineTimeAssistanceMeasurementsNotPossible, adrMeasurementsNotPossible,
// multiFrequencyMeasurementsNotPossible) are represented as booleans, since
// a NULL-typed OPTIONAL field's only content is its own presence bit.
type AGNSSError struct {
	Source                                    AGNSSErrorSource
	LocationServerCause                       AGNSSLocationServerErrorCause
	TargetDeviceCause                         AGNSSTargetDeviceErrorCause
	FineTimeAssistanceMeasurementsNotPossible bool
	ADRMeasurementsNotPossible                bool
	MultiFrequencyMeasurementsNotPossible     bool
}

func (v AGNSSError) Validate() error {
	switch v.Source {
	case AGNSSErrorLocationServer:
		if v.LocationServerCause > AGNSSLocationServerCauseAssistanceDataPartlyUnavailable {
			return fmt.Errorf("%w: invalid location server cause", ErrInvalidAGNSSError)
		}
	case AGNSSErrorTargetDevice:
		if v.TargetDeviceCause > AGNSSTargetDeviceCauseNotAllRequestedMeasurementsPossible {
			return fmt.Errorf("%w: invalid target device cause", ErrInvalidAGNSSError)
		}
	default:
		return fmt.Errorf("%w: invalid source", ErrInvalidAGNSSError)
	}
	return nil
}
func (v AGNSSError) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteRootChoiceIndex(uint64(v.Source), 2); err != nil {
		return err
	}
	if v.Source == AGNSSErrorLocationServer {
		if err := w.WriteExtensionPresent(false); err != nil { // outer SEQUENCE, no optional members
			return err
		}
		if err := w.WriteExtensionPresent(false); err != nil { // cause ENUMERATED
			return err
		}
		return w.WriteRootEnumerated(uint64(v.LocationServerCause), 4)
	}
	if err := w.WriteExtensionPresent(false); err != nil { // outer SEQUENCE, extensible, 3 optional members
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{v.FineTimeAssistanceMeasurementsNotPossible, v.ADRMeasurementsNotPossible, v.MultiFrequencyMeasurementsNotPossible}); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil { // cause ENUMERATED
		return err
	}
	return w.WriteRootEnumerated(uint64(v.TargetDeviceCause), 4)
}
func DecodeAGNSSError(r *uper.Reader) (AGNSSError, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return AGNSSError{}, err
	}
	if ext {
		return AGNSSError{}, ErrUnsupportedExtension
	}
	idx, err := r.ReadRootChoiceIndex(2)
	if err != nil {
		return AGNSSError{}, err
	}
	if idx == uint64(AGNSSErrorLocationServer) {
		seqExt, err := r.ReadExtensionPresent()
		if err != nil {
			return AGNSSError{}, err
		}
		if seqExt {
			return AGNSSError{}, ErrUnsupportedExtension
		}
		causeExt, err := r.ReadExtensionPresent()
		if err != nil {
			return AGNSSError{}, err
		}
		if causeExt {
			return AGNSSError{}, ErrUnsupportedExtension
		}
		c, err := r.ReadRootEnumerated(4)
		if err != nil {
			return AGNSSError{}, err
		}
		return AGNSSError{Source: AGNSSErrorLocationServer, LocationServerCause: AGNSSLocationServerErrorCause(c)}, nil
	}
	seqExt, err := r.ReadExtensionPresent()
	if err != nil {
		return AGNSSError{}, err
	}
	if seqExt {
		return AGNSSError{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(3)
	if err != nil {
		return AGNSSError{}, err
	}
	causeExt, err := r.ReadExtensionPresent()
	if err != nil {
		return AGNSSError{}, err
	}
	if causeExt {
		return AGNSSError{}, ErrUnsupportedExtension
	}
	c, err := r.ReadRootEnumerated(4)
	if err != nil {
		return AGNSSError{}, err
	}
	return AGNSSError{Source: AGNSSErrorTargetDevice, TargetDeviceCause: AGNSSTargetDeviceErrorCause(c), FineTimeAssistanceMeasurementsNotPossible: present[0], ADRMeasurementsNotPossible: present[1], MultiFrequencyMeasurementsNotPossible: present[2]}, nil
}

// AGNSSProvideLocationInformation is the bounded root TS 37.355
// A-GNSS-ProvideLocationInformation. GNSS-SignalMeasurementInformation (the
// MS-assisted raw-measurement branch) is out of scope and rejected if
// present; see docs/lpp-spec-audit.md's A-GNSS phase section.
type AGNSSProvideLocationInformation struct {
	GNSSLocationInformation *GNSSLocationInformation
	Error                   *AGNSSError
}

func (v AGNSSProvideLocationInformation) Validate() error {
	if v.GNSSLocationInformation != nil {
		if err := v.GNSSLocationInformation.Validate(); err != nil {
			return err
		}
	}
	if v.Error != nil {
		if err := v.Error.Validate(); err != nil {
			return err
		}
	}
	return nil
}
func (v AGNSSProvideLocationInformation) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{false, v.GNSSLocationInformation != nil, v.Error != nil}); err != nil {
		return err
	}
	if v.GNSSLocationInformation != nil {
		if err := v.GNSSLocationInformation.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.Error != nil {
		return v.Error.EncodeUPER(w)
	}
	return nil
}
func DecodeAGNSSProvideLocationInformation(r *uper.Reader) (AGNSSProvideLocationInformation, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return AGNSSProvideLocationInformation{}, err
	}
	if ext {
		return AGNSSProvideLocationInformation{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(3)
	if err != nil {
		return AGNSSProvideLocationInformation{}, err
	}
	if present[0] {
		return AGNSSProvideLocationInformation{}, ErrUnsupportedGNSSSignalMeasurements
	}
	v := AGNSSProvideLocationInformation{}
	if present[1] {
		x, err := DecodeGNSSLocationInformation(r)
		if err != nil {
			return AGNSSProvideLocationInformation{}, err
		}
		v.GNSSLocationInformation = &x
	}
	if present[2] {
		x, err := DecodeAGNSSError(r)
		if err != nil {
			return AGNSSProvideLocationInformation{}, err
		}
		v.Error = &x
	}
	return v, nil
}

func (v AGNSSProvideLocationInformation) clone() AGNSSProvideLocationInformation {
	out := v
	if v.GNSSLocationInformation != nil {
		x := *v.GNSSLocationInformation
		out.GNSSLocationInformation = &x
	}
	if v.Error != nil {
		x := *v.Error
		out.Error = &x
	}
	return out
}
