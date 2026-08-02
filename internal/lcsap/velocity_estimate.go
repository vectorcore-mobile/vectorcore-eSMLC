package lcsap

import "github.com/vectorcore/esmlc/internal/aper"

// Velocity-Estimate root CHOICE indexes, TS 29.171 V16.4.0 clause 9 (recovered
// LCS-AP-IEs module). The root CHOICE has no extension additions beyond the
// four listed alternatives as of this release.
const (
	velocityHorizontal                           = 0
	velocityHorizontalWithVertical               = 1
	velocityHorizontalWithUncertainty            = 2
	velocityHorizontalWithVerticalAndUncertainty = 3
)

// HorizontalSpeedAndBearing is TS 29.171 Horizontal-Speed-And-Bearing: a
// plain, non-extensible SEQUENCE with no OPTIONAL members.
type HorizontalSpeedAndBearing struct {
	Bearing uint16 // 0..359 degrees
	Speed   uint16 // 0..2047
}

// VerticalVelocity is TS 29.171 Vertical-Velocity: also a plain,
// non-extensible SEQUENCE with no OPTIONAL members.
type VerticalVelocity struct {
	Speed    uint8 // 0..255
	Downward bool  // Vertical-Speed-Direction: false = upward, true = downward
}

func writeHorizontalSpeedAndBearing(w *aper.Writer, v HorizontalSpeedAndBearing) error {
	if err := aper.PutConstrained(w, int64(v.Bearing), 0, 359); err != nil {
		return err
	}
	return aper.PutConstrained(w, int64(v.Speed), 0, 2047)
}

func writeVerticalVelocity(w *aper.Writer, v VerticalVelocity) error {
	if err := aper.PutConstrained(w, int64(v.Speed), 0, 255); err != nil {
		return err
	}
	return aper.PutConstrained(w, boolInt(v.Downward), 0, 1)
}

// EncodeHorizontalVelocity encodes Velocity-Estimate root shape 0
// (Horizontal-Velocity).
func EncodeHorizontalVelocity(v HorizontalSpeedAndBearing) ([]byte, error) {
	w := aper.NewWriter()
	wBit(w, 0)
	wbits(w, velocityHorizontal, 2)
	wbits(w, 0, 2) // outer SEQUENCE: extensible, one absent optional (iE-Extensions)
	if err := writeHorizontalSpeedAndBearing(w, v); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// EncodeHorizontalWithVerticalVelocity encodes Velocity-Estimate root shape 1.
func EncodeHorizontalWithVerticalVelocity(h HorizontalSpeedAndBearing, v VerticalVelocity) ([]byte, error) {
	w := aper.NewWriter()
	wBit(w, 0)
	wbits(w, velocityHorizontalWithVertical, 2)
	wbits(w, 0, 2)
	if err := writeHorizontalSpeedAndBearing(w, h); err != nil {
		return nil, err
	}
	if err := writeVerticalVelocity(w, v); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// EncodeHorizontalVelocityWithUncertainty encodes Velocity-Estimate root
// shape 2.
func EncodeHorizontalVelocityWithUncertainty(h HorizontalSpeedAndBearing, uncertaintySpeed uint8) ([]byte, error) {
	w := aper.NewWriter()
	wBit(w, 0)
	wbits(w, velocityHorizontalWithUncertainty, 2)
	wbits(w, 0, 2)
	if err := writeHorizontalSpeedAndBearing(w, h); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, int64(uncertaintySpeed), 0, 255); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// EncodeHorizontalWithVerticalVelocityAndUncertainty encodes Velocity-Estimate
// root shape 3.
func EncodeHorizontalWithVerticalVelocityAndUncertainty(h HorizontalSpeedAndBearing, v VerticalVelocity, horizontalUncertaintySpeed, verticalUncertaintySpeed uint8) ([]byte, error) {
	w := aper.NewWriter()
	wBit(w, 0)
	wbits(w, velocityHorizontalWithVerticalAndUncertainty, 2)
	wbits(w, 0, 2)
	if err := writeHorizontalSpeedAndBearing(w, h); err != nil {
		return nil, err
	}
	if err := writeVerticalVelocity(w, v); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, int64(horizontalUncertaintySpeed), 0, 255); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, int64(verticalUncertaintySpeed), 0, 255); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}
