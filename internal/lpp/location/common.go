package location

import (
	"errors"
	"fmt"
	"math"

	"github.com/vectorcore/esmlc/internal/uper"
)

var (
	ErrInvalidCommonRequest               = errors.New("lpp location: invalid common request-location payload")
	ErrInvalidCommonProvide               = errors.New("lpp location: invalid common provide-location payload")
	ErrUnsupportedLocationCoordinateShape = errors.New("lpp location: unsupported LocationCoordinates shape")
)

// LocationInformationType is the bounded root TS 37.355
// CommonIEsRequestLocationInformation.locationInformationType. Only the two
// "estimate" values matter to this implementation, which never requests raw
// measurements from a method that reports via LocationCoordinates.
type LocationInformationType uint8

const (
	LocationEstimateRequired LocationInformationType = iota
	LocationMeasurementsRequired
	LocationEstimatePreferred
	LocationMeasurementsPreferred
)

// CommonRequestLocationInformation is the bounded root of
// CommonIEsRequestLocationInformation: only locationInformationType.
// triggeredReporting, periodicalReporting, additionalInformation, qos,
// environment, locationCoordinateTypes, and velocityTypes are all OPTIONAL
// in the root and are not represented here.
type CommonRequestLocationInformation struct {
	LocationInformationType LocationInformationType
}

func (v CommonRequestLocationInformation) Validate() error {
	if v.LocationInformationType > LocationMeasurementsPreferred {
		return fmt.Errorf("%w: invalid locationInformationType", ErrInvalidCommonRequest)
	}
	return nil
}
func (v CommonRequestLocationInformation) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{false, false, false, false, false, false, false}); err != nil {
		return err
	}
	// LocationInformationType is itself an extensible ENUMERATED; its
	// extension-presence bit is separate from the SEQUENCE's own marker above.
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	return w.WriteRootEnumerated(uint64(v.LocationInformationType), 4)
}
func DecodeCommonRequestLocationInformation(r *uper.Reader) (CommonRequestLocationInformation, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return CommonRequestLocationInformation{}, err
	}
	if ext {
		return CommonRequestLocationInformation{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(7)
	if err != nil {
		return CommonRequestLocationInformation{}, err
	}
	for _, p := range present {
		if p {
			return CommonRequestLocationInformation{}, fmt.Errorf("%w: optional common request fields unsupported", ErrInvalidCommonRequest)
		}
	}
	typeExt, err := r.ReadExtensionPresent()
	if err != nil {
		return CommonRequestLocationInformation{}, err
	}
	if typeExt {
		return CommonRequestLocationInformation{}, ErrUnsupportedExtension
	}
	t, err := r.ReadRootEnumerated(4)
	if err != nil {
		return CommonRequestLocationInformation{}, err
	}
	v := CommonRequestLocationInformation{LocationInformationType: LocationInformationType(t)}
	return v, v.Validate()
}

// Coordinates is a decimal-degree geographic point, TS 37.355's own
// Ellipsoid-Point fields (distinct from, but structurally identical to,
// LCS-AP's Geographical-Coordinates in internal/lcsap): a 1-bit
// latitudeSign, 23-bit degreesLatitude magnitude, 24-bit signed
// degreesLongitude. Unlike the LCS-AP SEQUENCEs, TS 37.355's Ellipsoid-Point
// family is plain and non-extensible, so there is no header to write.
type Coordinates struct{ Latitude, Longitude float64 }

// signedRangeWidth returns hi-lo, the offset upper bound WriteConstrainedWholeNumber
// needs for a signed range shifted to start at zero; internal/uper's constrained
// whole number primitives only take non-negative uint64 bounds, so any signed
// ASN.1 range must be pre-offset by the caller (lo is always <= 0 here).
func writeSignedConstrained(w *uper.Writer, value, lo, hi int64) error {
	return w.WriteConstrainedWholeNumber(uint64(value-lo), 0, uint64(hi-lo))
}
func readSignedConstrained(r *uper.Reader, lo, hi int64) (int64, error) {
	offset, err := r.ReadConstrainedWholeNumber(0, uint64(hi-lo))
	if err != nil {
		return 0, err
	}
	return lo + int64(offset), nil
}

func writeCoordinates(w *uper.Writer, lat, lon float64) error {
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return fmt.Errorf("%w: coordinates out of range", ErrInvalidCommonProvide)
	}
	la := int64(math.Round(math.Abs(lat) * ((1 << 23) - 1) / 90))
	lo := int64(math.Round(lon * ((1 << 23) - 1) / 180))
	sign := uint64(0)
	if lat < 0 {
		sign = 1
	}
	if err := w.WriteRootEnumerated(sign, 2); err != nil {
		return err
	}
	if err := w.WriteConstrainedWholeNumber(uint64(la), 0, (1<<23)-1); err != nil {
		return err
	}
	return writeSignedConstrained(w, lo, -(1 << 23), (1<<23)-1)
}

// LocationCoordinatesShape identifies which TS 37.355 LocationCoordinates
// root CHOICE alternative is present. Only ShapePointWithUncertaintyCircle
// is implemented; the other six root shapes and both R15 high-accuracy
// extension shapes are rejected on encode and decode.
type LocationCoordinatesShape uint8

const (
	// shapeEllipsoidPoint is root CHOICE index 0 (ellipsoidPoint). It is not
	// implemented; the constant exists only to keep the implemented shapes at
	// their correct wire indices.
	shapeEllipsoidPoint LocationCoordinatesShape = iota
	ShapePointWithUncertaintyCircle
	shapeEllipsoidPointWithUncertaintyEllipse // index 2, not implemented
	shapePolygon                              // index 3, not implemented
	shapeEllipsoidPointWithAltitude           // index 4, not implemented
	ShapePointWithAltitudeAndUncertaintyEllipsoid
	shapeEllipsoidArc // index 6, not implemented
)

// AltitudeDirection is TS 37.355's non-extensible root ENUMERATED
// {height, depth} used by EllipsoidPointWithAltitudeAndUncertaintyEllipsoid.
type AltitudeDirection uint8

const (
	AltitudeHeight AltitudeDirection = iota
	AltitudeDepth
)

// LocationCoordinates is the bounded root TS 37.355 LocationCoordinates
// CHOICE, restricted to the two shapes this implementation produces and
// accepts: a coordinate with a circular uncertainty (the same shape ECID's
// authoritative serving-cell estimate already uses on the SLs side), and a
// coordinate with altitude and an uncertainty ellipsoid (the shape real
// UE-based A-GNSS fixes were observed to use on a live SLs association).
type LocationCoordinates struct {
	Shape             LocationCoordinatesShape
	Point             Coordinates
	UncertaintyCircle uint8 // Uncertainty-Code 0..127, GAD scale; ShapePointWithUncertaintyCircle only

	// The following are ShapePointWithAltitudeAndUncertaintyEllipsoid only.
	AltitudeDirection    AltitudeDirection
	Altitude             uint16 // metres, 0..32767
	UncertaintySemiMajor uint8  // Uncertainty-Code 0..127, GAD scale
	UncertaintySemiMinor uint8  // Uncertainty-Code 0..127, GAD scale
	OrientationMajorAxis uint8  // degrees from north, 0..179
	UncertaintyAltitude  uint8  // Uncertainty-Code 0..127, GAD scale
	Confidence           uint8  // percent, 0..100
}

func (v LocationCoordinates) Validate() error {
	if v.Shape != ShapePointWithUncertaintyCircle && v.Shape != ShapePointWithAltitudeAndUncertaintyEllipsoid {
		return fmt.Errorf("%w: shape %d", ErrUnsupportedLocationCoordinateShape, v.Shape)
	}
	if v.Point.Latitude < -90 || v.Point.Latitude > 90 || v.Point.Longitude < -180 || v.Point.Longitude > 180 {
		return fmt.Errorf("%w: coordinates out of range", ErrInvalidCommonProvide)
	}
	if v.Shape == ShapePointWithAltitudeAndUncertaintyEllipsoid {
		if v.AltitudeDirection != AltitudeHeight && v.AltitudeDirection != AltitudeDepth {
			return fmt.Errorf("%w: invalid altitudeDirection", ErrInvalidCommonProvide)
		}
		if v.OrientationMajorAxis > 179 {
			return fmt.Errorf("%w: orientationMajorAxis %d out of range", ErrInvalidCommonProvide, v.OrientationMajorAxis)
		}
		if v.Confidence > 100 {
			return fmt.Errorf("%w: confidence %d out of range", ErrInvalidCommonProvide, v.Confidence)
		}
	}
	return nil
}
func (v LocationCoordinates) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteRootChoiceIndex(uint64(v.Shape), 7); err != nil {
		return err
	}
	if err := writeCoordinates(w, v.Point.Latitude, v.Point.Longitude); err != nil {
		return err
	}
	switch v.Shape {
	case ShapePointWithUncertaintyCircle:
		return w.WriteConstrainedWholeNumber(uint64(v.UncertaintyCircle), 0, 127)
	case ShapePointWithAltitudeAndUncertaintyEllipsoid:
		if err := w.WriteRootEnumerated(uint64(v.AltitudeDirection), 2); err != nil {
			return err
		}
		if err := w.WriteConstrainedWholeNumber(uint64(v.Altitude), 0, 32767); err != nil {
			return err
		}
		if err := w.WriteConstrainedWholeNumber(uint64(v.UncertaintySemiMajor), 0, 127); err != nil {
			return err
		}
		if err := w.WriteConstrainedWholeNumber(uint64(v.UncertaintySemiMinor), 0, 127); err != nil {
			return err
		}
		if err := w.WriteConstrainedWholeNumber(uint64(v.OrientationMajorAxis), 0, 179); err != nil {
			return err
		}
		if err := w.WriteConstrainedWholeNumber(uint64(v.UncertaintyAltitude), 0, 127); err != nil {
			return err
		}
		return w.WriteConstrainedWholeNumber(uint64(v.Confidence), 0, 100)
	}
	return nil
}
func DecodeLocationCoordinates(r *uper.Reader) (LocationCoordinates, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return LocationCoordinates{}, err
	}
	if ext {
		return LocationCoordinates{}, ErrUnsupportedLocationCoordinateShape
	}
	idx, err := r.ReadRootChoiceIndex(7)
	if err != nil {
		return LocationCoordinates{}, err
	}
	if idx != uint64(ShapePointWithUncertaintyCircle) && idx != uint64(ShapePointWithAltitudeAndUncertaintyEllipsoid) {
		// Every other shape has a structurally different field layout past
		// this point (Polygon is a SEQUENCE OF, EllipsoidArc has five
		// further fields, etc.), so this must fail closed here rather than
		// attempt to keep decoding fields that do not apply.
		return LocationCoordinates{}, fmt.Errorf("%w: root index %d", ErrUnsupportedLocationCoordinateShape, idx)
	}
	sign, err := r.ReadRootEnumerated(2)
	if err != nil {
		return LocationCoordinates{}, err
	}
	la, err := r.ReadConstrainedWholeNumber(0, (1<<23)-1)
	if err != nil {
		return LocationCoordinates{}, err
	}
	lo, err := readSignedConstrained(r, -(1 << 23), (1<<23)-1)
	if err != nil {
		return LocationCoordinates{}, err
	}
	lat := float64(la) * 90 / ((1 << 23) - 1)
	if sign == 1 {
		lat = -lat
	}
	lon := float64(lo) * 180 / ((1 << 23) - 1)
	v := LocationCoordinates{Shape: LocationCoordinatesShape(idx), Point: Coordinates{Latitude: lat, Longitude: lon}}
	switch v.Shape {
	case ShapePointWithUncertaintyCircle:
		u, err := r.ReadConstrainedWholeNumber(0, 127)
		if err != nil {
			return LocationCoordinates{}, err
		}
		v.UncertaintyCircle = uint8(u)
	case ShapePointWithAltitudeAndUncertaintyEllipsoid:
		dir, err := r.ReadRootEnumerated(2)
		if err != nil {
			return LocationCoordinates{}, err
		}
		v.AltitudeDirection = AltitudeDirection(dir)
		alt, err := r.ReadConstrainedWholeNumber(0, 32767)
		if err != nil {
			return LocationCoordinates{}, err
		}
		v.Altitude = uint16(alt)
		semiMajor, err := r.ReadConstrainedWholeNumber(0, 127)
		if err != nil {
			return LocationCoordinates{}, err
		}
		v.UncertaintySemiMajor = uint8(semiMajor)
		semiMinor, err := r.ReadConstrainedWholeNumber(0, 127)
		if err != nil {
			return LocationCoordinates{}, err
		}
		v.UncertaintySemiMinor = uint8(semiMinor)
		orientation, err := r.ReadConstrainedWholeNumber(0, 179)
		if err != nil {
			return LocationCoordinates{}, err
		}
		v.OrientationMajorAxis = uint8(orientation)
		uncAlt, err := r.ReadConstrainedWholeNumber(0, 127)
		if err != nil {
			return LocationCoordinates{}, err
		}
		v.UncertaintyAltitude = uint8(uncAlt)
		confidence, err := r.ReadConstrainedWholeNumber(0, 100)
		if err != nil {
			return LocationCoordinates{}, err
		}
		v.Confidence = uint8(confidence)
	}
	return v, v.Validate()
}

// CommonProvideLocationInformation is the bounded root of
// CommonIEsProvideLocationInformation: only locationEstimate.
// velocityEstimate and locationError are OPTIONAL in the root and are not
// represented here.
type CommonProvideLocationInformation struct {
	LocationEstimate *LocationCoordinates
}

func (v CommonProvideLocationInformation) Validate() error {
	if v.LocationEstimate != nil {
		return v.LocationEstimate.Validate()
	}
	return nil
}
func (v CommonProvideLocationInformation) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{v.LocationEstimate != nil, false, false}); err != nil {
		return err
	}
	if v.LocationEstimate != nil {
		return v.LocationEstimate.EncodeUPER(w)
	}
	return nil
}
func DecodeCommonProvideLocationInformation(r *uper.Reader) (CommonProvideLocationInformation, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return CommonProvideLocationInformation{}, err
	}
	if ext {
		return CommonProvideLocationInformation{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(3)
	if err != nil {
		return CommonProvideLocationInformation{}, err
	}
	if present[1] || present[2] {
		return CommonProvideLocationInformation{}, fmt.Errorf("%w: velocityEstimate/locationError unsupported", ErrInvalidCommonProvide)
	}
	v := CommonProvideLocationInformation{}
	if present[0] {
		x, err := DecodeLocationCoordinates(r)
		if err != nil {
			return CommonProvideLocationInformation{}, err
		}
		v.LocationEstimate = &x
	}
	return v, nil
}

func (v CommonProvideLocationInformation) clone() CommonProvideLocationInformation {
	out := v
	if v.LocationEstimate != nil {
		x := *v.LocationEstimate
		out.LocationEstimate = &x
	}
	return out
}
