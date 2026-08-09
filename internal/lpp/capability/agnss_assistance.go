package capability

import (
	"fmt"

	"github.com/vectorcore/esmlc/internal/uper"
)

// This file implements the bounded root TS 37.355 A-GNSS-ProvideCapabilities
// fields beyond gnss-SupportList: assistanceDataSupportList,
// locationCoordinateTypes, and velocityTypes. Every type here follows the
// same fail-closed convention as the rest of this package: an extension
// marker bit set to true (a Release 10+ extension addition actually in use)
// is rejected rather than decoded, since none of those additions are
// represented. This is a direct, deliberate scope boundary — not an
// oversight — matching how ECID/OTDOA-ProvideCapabilities already bound
// themselves to their own Release 9 roots.

// ---- LocationCoordinateTypes ----

// LocationCoordinateTypes is the bounded TS 37.355 LocationCoordinateTypes
// root: all seven Release 9 boolean flags, in declaration order. The
// Release 15 extension addition
// (highAccuracyEllipsoidPointWithUncertaintyEllipse, ...) is out of bounded
// scope.
type LocationCoordinateTypes struct {
	EllipsoidPoint                                    bool
	EllipsoidPointWithUncertaintyCircle               bool
	EllipsoidPointWithUncertaintyEllipse              bool
	Polygon                                            bool
	EllipsoidPointWithAltitude                         bool
	EllipsoidPointWithAltitudeAndUncertaintyEllipsoid  bool
	EllipsoidArc                                       bool
}

func (v LocationCoordinateTypes) EncodeUPER(w *uper.Writer) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	for _, b := range []bool{v.EllipsoidPoint, v.EllipsoidPointWithUncertaintyCircle, v.EllipsoidPointWithUncertaintyEllipse, v.Polygon, v.EllipsoidPointWithAltitude, v.EllipsoidPointWithAltitudeAndUncertaintyEllipsoid, v.EllipsoidArc} {
		if err := w.WriteBoolean(b); err != nil {
			return err
		}
	}
	return nil
}
func DecodeLocationCoordinateTypes(r *uper.Reader) (LocationCoordinateTypes, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return LocationCoordinateTypes{}, err
	}
	if ext {
		return LocationCoordinateTypes{}, ErrUnsupportedExtension
	}
	vals := make([]bool, 7)
	for i := range vals {
		b, err := r.ReadBoolean()
		if err != nil {
			return LocationCoordinateTypes{}, err
		}
		vals[i] = b
	}
	return LocationCoordinateTypes{
		EllipsoidPoint:                                    vals[0],
		EllipsoidPointWithUncertaintyCircle:               vals[1],
		EllipsoidPointWithUncertaintyEllipse:              vals[2],
		Polygon:                                            vals[3],
		EllipsoidPointWithAltitude:                         vals[4],
		EllipsoidPointWithAltitudeAndUncertaintyEllipsoid:  vals[5],
		EllipsoidArc:                                       vals[6],
	}, nil
}

// ---- VelocityTypes ----

// VelocityTypes is the full TS 37.355 VelocityTypes root: all four booleans,
// in declaration order. There is no extension addition group for this type
// as of the release this codec targets, only the root extension marker.
type VelocityTypes struct {
	HorizontalVelocity                           bool
	HorizontalWithVerticalVelocity                bool
	HorizontalVelocityWithUncertainty             bool
	HorizontalWithVerticalVelocityAndUncertainty  bool
}

func (v VelocityTypes) EncodeUPER(w *uper.Writer) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	for _, b := range []bool{v.HorizontalVelocity, v.HorizontalWithVerticalVelocity, v.HorizontalVelocityWithUncertainty, v.HorizontalWithVerticalVelocityAndUncertainty} {
		if err := w.WriteBoolean(b); err != nil {
			return err
		}
	}
	return nil
}
func DecodeVelocityTypes(r *uper.Reader) (VelocityTypes, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return VelocityTypes{}, err
	}
	if ext {
		return VelocityTypes{}, ErrUnsupportedExtension
	}
	vals := make([]bool, 4)
	for i := range vals {
		b, err := r.ReadBoolean()
		if err != nil {
			return VelocityTypes{}, err
		}
		vals[i] = b
	}
	return VelocityTypes{
		HorizontalVelocity:                          vals[0],
		HorizontalWithVerticalVelocity:               vals[1],
		HorizontalVelocityWithUncertainty:            vals[2],
		HorizontalWithVerticalVelocityAndUncertainty: vals[3],
	}, nil
}

// ---- SBAS-ID ----

// SBASID is TS 37.355 SBAS-ID ::= SEQUENCE { sbas-id ENUMERATED { waas,
// egnos, msas, gagan, ... }, ... } — a SEQUENCE wrapping an extensible
// ENUMERATED, the same shape GNSSID already uses (two extension-marker
// bits: the wrapping SEQUENCE's, then the ENUMERATED's own).
type SBASID uint8

const (
	SBASIDWAAS SBASID = iota
	SBASIDEGNOS
	SBASIDMSAS
	SBASIDGAGAN
)

func (v SBASID) Validate() error {
	if v > SBASIDGAGAN {
		return fmt.Errorf("%w: sbas-id %d", ErrInvalidAGNSS, v)
	}
	return nil
}
func (v SBASID) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	return w.WriteRootEnumerated(uint64(v), 4)
}
func DecodeSBASID(r *uper.Reader) (SBASID, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return 0, err
	}
	if ext {
		return 0, ErrUnsupportedExtension
	}
	idExt, err := r.ReadExtensionPresent()
	if err != nil {
		return 0, err
	}
	if idExt {
		return 0, ErrUnsupportedExtension
	}
	v, err := r.ReadRootEnumerated(4)
	if err != nil {
		return 0, err
	}
	x := SBASID(v)
	return x, x.Validate()
}

// ---- GNSS-ID-Bitmap / AccessTypes (simple BIT STRING wrappers) ----

// GNSSIDBitmap is TS 37.355 GNSS-ID-Bitmap ::= SEQUENCE { gnss-ids BIT
// STRING (SIZE(1..16)), ... }.
type GNSSIDBitmap struct{ IDs uper.BitString }

func (v GNSSIDBitmap) EncodeUPER(w *uper.Writer) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	return w.WriteBitString(v.IDs, 1, 16)
}
func DecodeGNSSIDBitmap(r *uper.Reader) (GNSSIDBitmap, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSIDBitmap{}, err
	}
	if ext {
		return GNSSIDBitmap{}, ErrUnsupportedExtension
	}
	s, err := r.ReadBitString(1, 16)
	if err != nil {
		return GNSSIDBitmap{}, err
	}
	return GNSSIDBitmap{IDs: s}, nil
}

// AccessTypes is TS 37.355 AccessTypes ::= SEQUENCE { accessTypes BIT
// STRING (SIZE(1..8)), ... }.
type AccessTypes struct{ Types uper.BitString }

func (v AccessTypes) EncodeUPER(w *uper.Writer) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	return w.WriteBitString(v.Types, 1, 8)
}
func DecodeAccessTypes(r *uper.Reader) (AccessTypes, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return AccessTypes{}, err
	}
	if ext {
		return AccessTypes{}, ErrUnsupportedExtension
	}
	s, err := r.ReadBitString(1, 8)
	if err != nil {
		return AccessTypes{}, err
	}
	return AccessTypes{Types: s}, nil
}

// ---- GNSS-ReferenceTimeSupport ----

// GNSSReferenceTimeSupport is TS 37.355 GNSS-ReferenceTimeSupport ::=
// SEQUENCE { gnss-SystemTime GNSS-ID-Bitmap, fta-Support AccessTypes
// OPTIONAL, ... }.
type GNSSReferenceTimeSupport struct {
	SystemTime  GNSSIDBitmap
	FTASupport  *AccessTypes
}

func (v GNSSReferenceTimeSupport) EncodeUPER(w *uper.Writer) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{v.FTASupport != nil}); err != nil {
		return err
	}
	if err := v.SystemTime.EncodeUPER(w); err != nil {
		return err
	}
	if v.FTASupport != nil {
		return v.FTASupport.EncodeUPER(w)
	}
	return nil
}
func DecodeGNSSReferenceTimeSupport(r *uper.Reader) (GNSSReferenceTimeSupport, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSReferenceTimeSupport{}, err
	}
	if ext {
		return GNSSReferenceTimeSupport{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(1)
	if err != nil {
		return GNSSReferenceTimeSupport{}, err
	}
	st, err := DecodeGNSSIDBitmap(r)
	if err != nil {
		return GNSSReferenceTimeSupport{}, err
	}
	v := GNSSReferenceTimeSupport{SystemTime: st}
	if present[0] {
		fta, err := DecodeAccessTypes(r)
		if err != nil {
			return GNSSReferenceTimeSupport{}, err
		}
		v.FTASupport = &fta
	}
	return v, nil
}

// ---- Empty extensible-SEQUENCE "support" markers ----
//
// GNSS-ReferenceLocationSupport, GNSS-EarthOrientationParametersSupport,
// GNSS-TimeModelListSupport, GNSS-RealTimeIntegritySupport,
// GNSS-DataBitAssistanceSupport, GNSS-AcquisitionAssistanceSupport, and
// GNSS-AuxiliaryInformationSupport are all `SEQUENCE {...}` in the real
// spec: no root members at all (their content, if any, only exists as
// Release 10+ extension additions), so their whole encoding is the single
// SEQUENCE extension-marker bit.

type GNSSReferenceLocationSupport struct{}

func (v GNSSReferenceLocationSupport) EncodeUPER(w *uper.Writer) error {
	return w.WriteExtensionPresent(false)
}
func DecodeGNSSReferenceLocationSupport(r *uper.Reader) (GNSSReferenceLocationSupport, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSReferenceLocationSupport{}, err
	}
	if ext {
		return GNSSReferenceLocationSupport{}, ErrUnsupportedExtension
	}
	return GNSSReferenceLocationSupport{}, nil
}

type GNSSEarthOrientationParametersSupport struct{}

func (v GNSSEarthOrientationParametersSupport) EncodeUPER(w *uper.Writer) error {
	return w.WriteExtensionPresent(false)
}
func DecodeGNSSEarthOrientationParametersSupport(r *uper.Reader) (GNSSEarthOrientationParametersSupport, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSEarthOrientationParametersSupport{}, err
	}
	if ext {
		return GNSSEarthOrientationParametersSupport{}, ErrUnsupportedExtension
	}
	return GNSSEarthOrientationParametersSupport{}, nil
}

type GNSSTimeModelListSupport struct{}

func (v GNSSTimeModelListSupport) EncodeUPER(w *uper.Writer) error {
	return w.WriteExtensionPresent(false)
}
func DecodeGNSSTimeModelListSupport(r *uper.Reader) (GNSSTimeModelListSupport, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSTimeModelListSupport{}, err
	}
	if ext {
		return GNSSTimeModelListSupport{}, ErrUnsupportedExtension
	}
	return GNSSTimeModelListSupport{}, nil
}

type GNSSRealTimeIntegritySupport struct{}

func (v GNSSRealTimeIntegritySupport) EncodeUPER(w *uper.Writer) error {
	return w.WriteExtensionPresent(false)
}
func DecodeGNSSRealTimeIntegritySupport(r *uper.Reader) (GNSSRealTimeIntegritySupport, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSRealTimeIntegritySupport{}, err
	}
	if ext {
		return GNSSRealTimeIntegritySupport{}, ErrUnsupportedExtension
	}
	return GNSSRealTimeIntegritySupport{}, nil
}

type GNSSDataBitAssistanceSupport struct{}

func (v GNSSDataBitAssistanceSupport) EncodeUPER(w *uper.Writer) error {
	return w.WriteExtensionPresent(false)
}
func DecodeGNSSDataBitAssistanceSupport(r *uper.Reader) (GNSSDataBitAssistanceSupport, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSDataBitAssistanceSupport{}, err
	}
	if ext {
		return GNSSDataBitAssistanceSupport{}, ErrUnsupportedExtension
	}
	return GNSSDataBitAssistanceSupport{}, nil
}

// GNSSAcquisitionAssistanceSupport has no Release 9 root members at all
// (confidenceSupport-r10/dopplerUncertaintyExtSupport-r10 are both
// extension additions), matching the empty-marker shape above.
type GNSSAcquisitionAssistanceSupport struct{}

func (v GNSSAcquisitionAssistanceSupport) EncodeUPER(w *uper.Writer) error {
	return w.WriteExtensionPresent(false)
}
func DecodeGNSSAcquisitionAssistanceSupport(r *uper.Reader) (GNSSAcquisitionAssistanceSupport, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSAcquisitionAssistanceSupport{}, err
	}
	if ext {
		return GNSSAcquisitionAssistanceSupport{}, ErrUnsupportedExtension
	}
	return GNSSAcquisitionAssistanceSupport{}, nil
}

type GNSSAuxiliaryInformationSupport struct{}

func (v GNSSAuxiliaryInformationSupport) EncodeUPER(w *uper.Writer) error {
	return w.WriteExtensionPresent(false)
}
func DecodeGNSSAuxiliaryInformationSupport(r *uper.Reader) (GNSSAuxiliaryInformationSupport, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSAuxiliaryInformationSupport{}, err
	}
	if ext {
		return GNSSAuxiliaryInformationSupport{}, ErrUnsupportedExtension
	}
	return GNSSAuxiliaryInformationSupport{}, nil
}

// ---- GNSS-IonosphericModelSupport ----

type GNSSIonosphericModelSupport struct{ IonoModel uper.BitString }

func (v GNSSIonosphericModelSupport) EncodeUPER(w *uper.Writer) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	return w.WriteBitString(v.IonoModel, 1, 8)
}
func DecodeGNSSIonosphericModelSupport(r *uper.Reader) (GNSSIonosphericModelSupport, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSIonosphericModelSupport{}, err
	}
	if ext {
		return GNSSIonosphericModelSupport{}, ErrUnsupportedExtension
	}
	s, err := r.ReadBitString(1, 8)
	if err != nil {
		return GNSSIonosphericModelSupport{}, err
	}
	return GNSSIonosphericModelSupport{IonoModel: s}, nil
}

// ---- GNSS-DifferentialCorrectionsSupport ----

// GNSSDifferentialCorrectionsSupport reuses this package's existing
// GNSS-SignalIDs helpers (readGNSSSignalIDs/writeGNSSSignalIDs, already used
// by GNSSSupportElement — a fixed 8-bit BIT STRING).
type GNSSDifferentialCorrectionsSupport struct {
	SignalIDs           uper.BitString
	ValidityTimeSupport bool
}

func (v GNSSDifferentialCorrectionsSupport) EncodeUPER(w *uper.Writer) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := writeGNSSSignalIDs(w, v.SignalIDs); err != nil {
		return err
	}
	return w.WriteBoolean(v.ValidityTimeSupport)
}
func DecodeGNSSDifferentialCorrectionsSupport(r *uper.Reader) (GNSSDifferentialCorrectionsSupport, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSDifferentialCorrectionsSupport{}, err
	}
	if ext {
		return GNSSDifferentialCorrectionsSupport{}, ErrUnsupportedExtension
	}
	sig, err := readGNSSSignalIDs(r)
	if err != nil {
		return GNSSDifferentialCorrectionsSupport{}, err
	}
	valid, err := r.ReadBoolean()
	if err != nil {
		return GNSSDifferentialCorrectionsSupport{}, err
	}
	return GNSSDifferentialCorrectionsSupport{SignalIDs: sig, ValidityTimeSupport: valid}, nil
}

// ---- GNSS-NavigationModelSupport ----

type GNSSNavigationModelSupport struct {
	ClockModel *uper.BitString
	OrbitModel *uper.BitString
}

func (v GNSSNavigationModelSupport) EncodeUPER(w *uper.Writer) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{v.ClockModel != nil, v.OrbitModel != nil}); err != nil {
		return err
	}
	if v.ClockModel != nil {
		if err := w.WriteBitString(*v.ClockModel, 1, 8); err != nil {
			return err
		}
	}
	if v.OrbitModel != nil {
		return w.WriteBitString(*v.OrbitModel, 1, 8)
	}
	return nil
}
func DecodeGNSSNavigationModelSupport(r *uper.Reader) (GNSSNavigationModelSupport, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSNavigationModelSupport{}, err
	}
	if ext {
		return GNSSNavigationModelSupport{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(2)
	if err != nil {
		return GNSSNavigationModelSupport{}, err
	}
	v := GNSSNavigationModelSupport{}
	if present[0] {
		s, err := r.ReadBitString(1, 8)
		if err != nil {
			return GNSSNavigationModelSupport{}, err
		}
		v.ClockModel = &s
	}
	if present[1] {
		s, err := r.ReadBitString(1, 8)
		if err != nil {
			return GNSSNavigationModelSupport{}, err
		}
		v.OrbitModel = &s
	}
	return v, nil
}

// ---- GNSS-AlmanacSupport / GNSS-UTC-ModelSupport ----

type GNSSAlmanacSupport struct{ AlmanacModel *uper.BitString }

func (v GNSSAlmanacSupport) EncodeUPER(w *uper.Writer) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{v.AlmanacModel != nil}); err != nil {
		return err
	}
	if v.AlmanacModel != nil {
		return w.WriteBitString(*v.AlmanacModel, 1, 8)
	}
	return nil
}
func DecodeGNSSAlmanacSupport(r *uper.Reader) (GNSSAlmanacSupport, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSAlmanacSupport{}, err
	}
	if ext {
		return GNSSAlmanacSupport{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(1)
	if err != nil {
		return GNSSAlmanacSupport{}, err
	}
	v := GNSSAlmanacSupport{}
	if present[0] {
		s, err := r.ReadBitString(1, 8)
		if err != nil {
			return GNSSAlmanacSupport{}, err
		}
		v.AlmanacModel = &s
	}
	return v, nil
}

type GNSSUTCModelSupport struct{ UTCModel *uper.BitString }

func (v GNSSUTCModelSupport) EncodeUPER(w *uper.Writer) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{v.UTCModel != nil}); err != nil {
		return err
	}
	if v.UTCModel != nil {
		return w.WriteBitString(*v.UTCModel, 1, 8)
	}
	return nil
}
func DecodeGNSSUTCModelSupport(r *uper.Reader) (GNSSUTCModelSupport, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSUTCModelSupport{}, err
	}
	if ext {
		return GNSSUTCModelSupport{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(1)
	if err != nil {
		return GNSSUTCModelSupport{}, err
	}
	v := GNSSUTCModelSupport{}
	if present[0] {
		s, err := r.ReadBitString(1, 8)
		if err != nil {
			return GNSSUTCModelSupport{}, err
		}
		v.UTCModel = &s
	}
	return v, nil
}

// ---- GNSS-CommonAssistanceDataSupport ----

// GNSSCommonAssistanceDataSupport is TS 37.355
// GNSS-CommonAssistanceDataSupport: four root OPTIONAL members, in
// declaration order. The Release 15 RTK extension addition group is out of
// bounded scope.
type GNSSCommonAssistanceDataSupport struct {
	ReferenceTime     *GNSSReferenceTimeSupport
	ReferenceLocation *GNSSReferenceLocationSupport
	IonosphericModel  *GNSSIonosphericModelSupport
	EarthOrientation  *GNSSEarthOrientationParametersSupport
}

func (v GNSSCommonAssistanceDataSupport) EncodeUPER(w *uper.Writer) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{v.ReferenceTime != nil, v.ReferenceLocation != nil, v.IonosphericModel != nil, v.EarthOrientation != nil}); err != nil {
		return err
	}
	if v.ReferenceTime != nil {
		if err := v.ReferenceTime.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.ReferenceLocation != nil {
		if err := v.ReferenceLocation.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.IonosphericModel != nil {
		if err := v.IonosphericModel.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.EarthOrientation != nil {
		return v.EarthOrientation.EncodeUPER(w)
	}
	return nil
}
func DecodeGNSSCommonAssistanceDataSupport(r *uper.Reader) (GNSSCommonAssistanceDataSupport, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSCommonAssistanceDataSupport{}, err
	}
	if ext {
		return GNSSCommonAssistanceDataSupport{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(4)
	if err != nil {
		return GNSSCommonAssistanceDataSupport{}, err
	}
	v := GNSSCommonAssistanceDataSupport{}
	if present[0] {
		x, err := DecodeGNSSReferenceTimeSupport(r)
		if err != nil {
			return GNSSCommonAssistanceDataSupport{}, err
		}
		v.ReferenceTime = &x
	}
	if present[1] {
		x, err := DecodeGNSSReferenceLocationSupport(r)
		if err != nil {
			return GNSSCommonAssistanceDataSupport{}, err
		}
		v.ReferenceLocation = &x
	}
	if present[2] {
		x, err := DecodeGNSSIonosphericModelSupport(r)
		if err != nil {
			return GNSSCommonAssistanceDataSupport{}, err
		}
		v.IonosphericModel = &x
	}
	if present[3] {
		x, err := DecodeGNSSEarthOrientationParametersSupport(r)
		if err != nil {
			return GNSSCommonAssistanceDataSupport{}, err
		}
		v.EarthOrientation = &x
	}
	return v, nil
}

// ---- GNSS-GenericAssistDataSupportElement / GNSS-GenericAssistanceDataSupport ----

// GNSSGenericAssistDataSupportElement is TS 37.355
// GNSS-GenericAssistDataSupportElement: one mandatory field (gnss-ID) then
// ten root OPTIONAL members, in declaration order. The Release 12+
// extension addition groups (bds-DifferentialCorrectionsSupport-r12, ...)
// are out of bounded scope.
type GNSSGenericAssistDataSupportElement struct {
	ID                             GNSSID
	SBASID                         *SBASID
	TimeModelsSupport              *GNSSTimeModelListSupport
	DifferentialCorrectionsSupport *GNSSDifferentialCorrectionsSupport
	NavigationModelSupport         *GNSSNavigationModelSupport
	RealTimeIntegritySupport       *GNSSRealTimeIntegritySupport
	DataBitAssistanceSupport       *GNSSDataBitAssistanceSupport
	AcquisitionAssistanceSupport   *GNSSAcquisitionAssistanceSupport
	AlmanacSupport                 *GNSSAlmanacSupport
	UTCModelSupport                *GNSSUTCModelSupport
	AuxiliaryInformationSupport    *GNSSAuxiliaryInformationSupport
}

func (v GNSSGenericAssistDataSupportElement) EncodeUPER(w *uper.Writer) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{
		v.SBASID != nil, v.TimeModelsSupport != nil, v.DifferentialCorrectionsSupport != nil,
		v.NavigationModelSupport != nil, v.RealTimeIntegritySupport != nil, v.DataBitAssistanceSupport != nil,
		v.AcquisitionAssistanceSupport != nil, v.AlmanacSupport != nil, v.UTCModelSupport != nil,
		v.AuxiliaryInformationSupport != nil,
	}); err != nil {
		return err
	}
	if err := v.ID.EncodeUPER(w); err != nil {
		return err
	}
	if v.SBASID != nil {
		if err := v.SBASID.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.TimeModelsSupport != nil {
		if err := v.TimeModelsSupport.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.DifferentialCorrectionsSupport != nil {
		if err := v.DifferentialCorrectionsSupport.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.NavigationModelSupport != nil {
		if err := v.NavigationModelSupport.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.RealTimeIntegritySupport != nil {
		if err := v.RealTimeIntegritySupport.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.DataBitAssistanceSupport != nil {
		if err := v.DataBitAssistanceSupport.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.AcquisitionAssistanceSupport != nil {
		if err := v.AcquisitionAssistanceSupport.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.AlmanacSupport != nil {
		if err := v.AlmanacSupport.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.UTCModelSupport != nil {
		if err := v.UTCModelSupport.EncodeUPER(w); err != nil {
			return err
		}
	}
	if v.AuxiliaryInformationSupport != nil {
		return v.AuxiliaryInformationSupport.EncodeUPER(w)
	}
	return nil
}
func DecodeGNSSGenericAssistDataSupportElement(r *uper.Reader) (GNSSGenericAssistDataSupportElement, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSGenericAssistDataSupportElement{}, err
	}
	if ext {
		return GNSSGenericAssistDataSupportElement{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(10)
	if err != nil {
		return GNSSGenericAssistDataSupportElement{}, err
	}
	id, err := DecodeGNSSID(r)
	if err != nil {
		return GNSSGenericAssistDataSupportElement{}, err
	}
	v := GNSSGenericAssistDataSupportElement{ID: id}
	if present[0] {
		x, err := DecodeSBASID(r)
		if err != nil {
			return GNSSGenericAssistDataSupportElement{}, err
		}
		v.SBASID = &x
	}
	if present[1] {
		x, err := DecodeGNSSTimeModelListSupport(r)
		if err != nil {
			return GNSSGenericAssistDataSupportElement{}, err
		}
		v.TimeModelsSupport = &x
	}
	if present[2] {
		x, err := DecodeGNSSDifferentialCorrectionsSupport(r)
		if err != nil {
			return GNSSGenericAssistDataSupportElement{}, err
		}
		v.DifferentialCorrectionsSupport = &x
	}
	if present[3] {
		x, err := DecodeGNSSNavigationModelSupport(r)
		if err != nil {
			return GNSSGenericAssistDataSupportElement{}, err
		}
		v.NavigationModelSupport = &x
	}
	if present[4] {
		x, err := DecodeGNSSRealTimeIntegritySupport(r)
		if err != nil {
			return GNSSGenericAssistDataSupportElement{}, err
		}
		v.RealTimeIntegritySupport = &x
	}
	if present[5] {
		x, err := DecodeGNSSDataBitAssistanceSupport(r)
		if err != nil {
			return GNSSGenericAssistDataSupportElement{}, err
		}
		v.DataBitAssistanceSupport = &x
	}
	if present[6] {
		x, err := DecodeGNSSAcquisitionAssistanceSupport(r)
		if err != nil {
			return GNSSGenericAssistDataSupportElement{}, err
		}
		v.AcquisitionAssistanceSupport = &x
	}
	if present[7] {
		x, err := DecodeGNSSAlmanacSupport(r)
		if err != nil {
			return GNSSGenericAssistDataSupportElement{}, err
		}
		v.AlmanacSupport = &x
	}
	if present[8] {
		x, err := DecodeGNSSUTCModelSupport(r)
		if err != nil {
			return GNSSGenericAssistDataSupportElement{}, err
		}
		v.UTCModelSupport = &x
	}
	if present[9] {
		x, err := DecodeGNSSAuxiliaryInformationSupport(r)
		if err != nil {
			return GNSSGenericAssistDataSupportElement{}, err
		}
		v.AuxiliaryInformationSupport = &x
	}
	return v, nil
}

// GNSSGenericAssistanceDataSupport is TS 37.355
// GNSS-GenericAssistanceDataSupport ::= SEQUENCE (SIZE(1..16)) OF
// GNSS-GenericAssistDataSupportElement.
type GNSSGenericAssistanceDataSupport struct {
	Elements []GNSSGenericAssistDataSupportElement
}

func (v GNSSGenericAssistanceDataSupport) Validate() error {
	if len(v.Elements) < 1 || len(v.Elements) > 16 {
		return fmt.Errorf("%w: generic assistance data support count %d", ErrInvalidAGNSS, len(v.Elements))
	}
	return nil
}
func (v GNSSGenericAssistanceDataSupport) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	return w.WriteSequenceOf(len(v.Elements), 1, 16, func(index int, writer *uper.Writer) error {
		return v.Elements[index].EncodeUPER(writer)
	})
}
func DecodeGNSSGenericAssistanceDataSupport(r *uper.Reader) (GNSSGenericAssistanceDataSupport, error) {
	list := make([]GNSSGenericAssistDataSupportElement, 0, 16)
	_, err := r.ReadSequenceOf(1, 16, func(index int, reader *uper.Reader) error {
		x, err := DecodeGNSSGenericAssistDataSupportElement(reader)
		if err != nil {
			return err
		}
		list = append(list, x)
		return nil
	})
	if err != nil {
		return GNSSGenericAssistanceDataSupport{}, err
	}
	return GNSSGenericAssistanceDataSupport{Elements: list}, nil
}

// ---- AssistanceDataSupportList ----

// AssistanceDataSupportList is TS 37.355 AssistanceDataSupportList: two
// mandatory root members (no optional-presence bitmap needed), extensible.
type AssistanceDataSupportList struct {
	Common  GNSSCommonAssistanceDataSupport
	Generic GNSSGenericAssistanceDataSupport
}

func (v AssistanceDataSupportList) EncodeUPER(w *uper.Writer) error {
	if err := v.Generic.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := v.Common.EncodeUPER(w); err != nil {
		return err
	}
	return v.Generic.EncodeUPER(w)
}
func DecodeAssistanceDataSupportList(r *uper.Reader) (AssistanceDataSupportList, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return AssistanceDataSupportList{}, err
	}
	if ext {
		return AssistanceDataSupportList{}, ErrUnsupportedExtension
	}
	common, err := DecodeGNSSCommonAssistanceDataSupport(r)
	if err != nil {
		return AssistanceDataSupportList{}, err
	}
	generic, err := DecodeGNSSGenericAssistanceDataSupport(r)
	if err != nil {
		return AssistanceDataSupportList{}, err
	}
	return AssistanceDataSupportList{Common: common, Generic: generic}, nil
}
