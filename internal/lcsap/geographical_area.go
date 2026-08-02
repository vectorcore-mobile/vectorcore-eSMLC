package lcsap

import (
	"fmt"

	"github.com/vectorcore/esmlc/internal/aper"
)

// Geographical-Area root CHOICE indexes, TS 29.171 V16.4.0 clause 9 (recovered
// LCS-AP-IEs module). The four Release-15/16 High-Accuracy extension shapes
// are not implemented; encoding or decoding them is rejected rather than
// guessed, matching the fail-closed extension handling used throughout this
// package.
const (
	geoAreaPoint                                             = 0
	geoAreaPointWithUncertainty                              = 1
	geoAreaEllipsoidPointWithUncertaintyEllipse              = 2
	geoAreaPolygon                                           = 3
	geoAreaEllipsoidPointWithAltitude                        = 4
	geoAreaEllipsoidPointWithAltitudeAndUncertaintyEllipsoid = 5
	geoAreaEllipsoidArc                                      = 6
)

// maxPolygonPoints is TS 29.171 LCS-AP-Constants max-No-Of-Points.
const maxPolygonPoints = 15

// Coordinates is a decimal-degree geographic point, encoded as TS 29.171
// Geographical-Coordinates: a 1-bit LatitudeSign, a 23-bit DegreesLatitude
// magnitude, and a 24-bit signed DegreesLongitude.
type Coordinates struct{ Latitude, Longitude float64 }

// UncertaintyEllipse is TS 29.171 Uncertainty-Ellipse: two Uncertainty-Code
// semi-axes (0..127) and an Orientation-Major-Axis (0..89).
type UncertaintyEllipse struct{ SemiMajor, SemiMinor, Orientation uint8 }

// AltitudeAndDirection is TS 29.171 Altitude-And-Direction. The spec notes
// that only Altitude 0..32767 is currently valid and larger received values
// must be mapped to 32767; this package encodes exactly what it is given and
// leaves that mapping to the caller.
type AltitudeAndDirection struct {
	Depth    bool   // Direction-Of-Altitude: false = height, true = depth
	Altitude uint16 // 0..65535
}

func (v UncertaintyEllipse) validate() error {
	if v.Orientation > 89 {
		return fmt.Errorf("lcsap: orientation-major-axis out of range")
	}
	return nil
}

// writeUncertaintyEllipse encodes Uncertainty-Ellipse. It is a plain,
// non-extensible SEQUENCE with no OPTIONAL members, so it has no header bits.
func writeUncertaintyEllipse(w *aper.Writer, v UncertaintyEllipse) error {
	if err := v.validate(); err != nil {
		return err
	}
	if err := aper.PutConstrained(w, int64(v.SemiMajor), 0, 127); err != nil {
		return err
	}
	if err := aper.PutConstrained(w, int64(v.SemiMinor), 0, 127); err != nil {
		return err
	}
	return aper.PutConstrained(w, int64(v.Orientation), 0, 89)
}

// writeAltitudeAndDirection encodes Altitude-And-Direction, an extensible
// SEQUENCE with no OPTIONAL root members: one extension-marker bit only.
func writeAltitudeAndDirection(w *aper.Writer, v AltitudeAndDirection) error {
	wBit(w, 0)
	if err := aper.PutConstrained(w, boolInt(v.Depth), 0, 1); err != nil {
		return err
	}
	return aper.PutConstrained(w, int64(v.Altitude), 0, 65535)
}

func writeConfidence(w *aper.Writer, v uint8) error {
	return aper.PutConstrained(w, int64(v), 0, 100)
}

func writeAngle(w *aper.Writer, v uint8) error {
	return aper.PutConstrained(w, int64(v), 0, 179)
}

// EncodePoint encodes Geographical-Area root shape 0 (Point): a bare
// coordinate with no uncertainty.
func EncodePoint(c Coordinates) ([]byte, error) {
	w := aper.NewWriter()
	wBit(w, 0)
	wbits(w, geoAreaPoint, 3)
	wbits(w, 0, 2) // Point: extensible SEQUENCE, one absent optional (iE-Extensions).
	if err := writeGeographicalCoordinates(w, c.Latitude, c.Longitude); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// EncodeEllipsoidPointWithUncertaintyEllipse encodes Geographical-Area root
// shape 2 (Ellipsoid-Point-With-Uncertainty-Ellipse): a coordinate with an
// elliptical uncertainty region and a confidence percentage.
func EncodeEllipsoidPointWithUncertaintyEllipse(c Coordinates, u UncertaintyEllipse, confidence uint8) ([]byte, error) {
	w := aper.NewWriter()
	wBit(w, 0)
	wbits(w, geoAreaEllipsoidPointWithUncertaintyEllipse, 3)
	wbits(w, 0, 2) // outer SEQUENCE header: extensible, one absent optional
	if err := writeGeographicalCoordinates(w, c.Latitude, c.Longitude); err != nil {
		return nil, err
	}
	if err := writeUncertaintyEllipse(w, u); err != nil {
		return nil, err
	}
	if err := writeConfidence(w, confidence); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// EncodePolygon encodes Geographical-Area root shape 3: 1..15 points
// (TS 29.171 LCS-AP-Constants max-No-Of-Points). The size constraint itself
// is not extensible, so the count carries no extension-marker bit.
func EncodePolygon(points []Coordinates) ([]byte, error) {
	if len(points) < 1 || len(points) > maxPolygonPoints {
		return nil, fmt.Errorf("lcsap: polygon must contain 1..%d points", maxPolygonPoints)
	}
	w := aper.NewWriter()
	wBit(w, 0)
	wbits(w, geoAreaPolygon, 3)
	if err := aper.PutConstrained(w, int64(len(points)), 1, maxPolygonPoints); err != nil {
		return nil, err
	}
	for _, p := range points {
		wbits(w, 0, 2) // Polygon-Point: extensible SEQUENCE, one absent optional
		if err := writeGeographicalCoordinates(w, p.Latitude, p.Longitude); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// EncodeEllipsoidPointWithAltitude encodes Geographical-Area root shape 4.
func EncodeEllipsoidPointWithAltitude(c Coordinates, alt AltitudeAndDirection) ([]byte, error) {
	w := aper.NewWriter()
	wBit(w, 0)
	wbits(w, geoAreaEllipsoidPointWithAltitude, 3)
	wbits(w, 0, 2) // outer SEQUENCE header
	if err := writeGeographicalCoordinates(w, c.Latitude, c.Longitude); err != nil {
		return nil, err
	}
	if err := writeAltitudeAndDirection(w, alt); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// EncodeEllipsoidPointWithAltitudeAndUncertaintyEllipsoid encodes
// Geographical-Area root shape 5: a 3D coordinate with horizontal elliptical
// uncertainty, vertical uncertainty, and a confidence percentage.
func EncodeEllipsoidPointWithAltitudeAndUncertaintyEllipsoid(c Coordinates, alt AltitudeAndDirection, u UncertaintyEllipse, uncertaintyAltitude, confidence uint8) ([]byte, error) {
	w := aper.NewWriter()
	wBit(w, 0)
	wbits(w, geoAreaEllipsoidPointWithAltitudeAndUncertaintyEllipsoid, 3)
	wbits(w, 0, 2) // outer SEQUENCE header
	if err := writeGeographicalCoordinates(w, c.Latitude, c.Longitude); err != nil {
		return nil, err
	}
	if err := writeAltitudeAndDirection(w, alt); err != nil {
		return nil, err
	}
	if err := writeUncertaintyEllipse(w, u); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, int64(uncertaintyAltitude), 0, 127); err != nil {
		return nil, err
	}
	if err := writeConfidence(w, confidence); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// EncodeEllipsoidArc encodes Geographical-Area root shape 6.
func EncodeEllipsoidArc(c Coordinates, innerRadius uint16, uncertaintyRadius, offsetAngle, includedAngle, confidence uint8) ([]byte, error) {
	w := aper.NewWriter()
	wBit(w, 0)
	wbits(w, geoAreaEllipsoidArc, 3)
	wbits(w, 0, 2) // outer SEQUENCE header
	if err := writeGeographicalCoordinates(w, c.Latitude, c.Longitude); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, int64(innerRadius), 0, 65535); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, int64(uncertaintyRadius), 0, 127); err != nil {
		return nil, err
	}
	if err := writeAngle(w, offsetAngle); err != nil {
		return nil, err
	}
	if err := writeAngle(w, includedAngle); err != nil {
		return nil, err
	}
	if err := writeConfidence(w, confidence); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}
