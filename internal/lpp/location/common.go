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
	// implemented; the constant exists only to keep
	// ShapePointWithUncertaintyCircle at its correct wire index (1).
	shapeEllipsoidPoint LocationCoordinatesShape = iota
	ShapePointWithUncertaintyCircle
)

// LocationCoordinates is the bounded root TS 37.355 LocationCoordinates
// CHOICE, restricted to the one shape this implementation produces and
// accepts: a coordinate with a circular uncertainty, the same shape ECID's
// authoritative serving-cell estimate already uses on the SLs side.
type LocationCoordinates struct {
	Shape             LocationCoordinatesShape
	Point             Coordinates
	UncertaintyCircle uint8 // Uncertainty-Code 0..127, GAD scale
}

func (v LocationCoordinates) Validate() error {
	if v.Shape != ShapePointWithUncertaintyCircle {
		return fmt.Errorf("%w: shape %d", ErrUnsupportedLocationCoordinateShape, v.Shape)
	}
	if v.Point.Latitude < -90 || v.Point.Latitude > 90 || v.Point.Longitude < -180 || v.Point.Longitude > 180 {
		return fmt.Errorf("%w: coordinates out of range", ErrInvalidCommonProvide)
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
	if v.Shape == ShapePointWithUncertaintyCircle {
		return w.WriteConstrainedWholeNumber(uint64(v.UncertaintyCircle), 0, 127)
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
	if idx != uint64(ShapePointWithUncertaintyCircle) {
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
	if v.Shape == ShapePointWithUncertaintyCircle {
		u, err := r.ReadConstrainedWholeNumber(0, 127)
		if err != nil {
			return LocationCoordinates{}, err
		}
		v.UncertaintyCircle = uint8(u)
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
