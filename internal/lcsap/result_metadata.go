package lcsap

import (
	"fmt"

	"github.com/vectorcore/esmlc/internal/aper"
)

// AccuracyFulfillmentIndicator is the root TS 29.171 IE. Extensions are not
// implemented and are rejected by the decoder.
type AccuracyFulfillmentIndicator uint8

const (
	AccuracyFulfilled AccuracyFulfillmentIndicator = iota
	AccuracyNotFulfilled
)

func EncodeAccuracyFulfillmentIndicator(v AccuracyFulfillmentIndicator) ([]byte, error) {
	if v > AccuracyNotFulfilled {
		return nil, fmt.Errorf("lcsap: invalid accuracy fulfilment indicator")
	}
	w := aper.NewWriter()
	// Extensible ENUMERATED root: extension bit followed by the root index.
	if err := aper.PutConstrained(w, 0, 0, 1); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, int64(v), 0, 1); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func DecodeAccuracyFulfillmentIndicator(b []byte) (AccuracyFulfillmentIndicator, error) {
	r := aper.NewReader(b)
	ext, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return 0, err
	}
	if ext != 0 {
		return 0, fmt.Errorf("lcsap: accuracy fulfilment extensions unsupported")
	}
	v, err := aper.GetConstrained(r, 0, 1)
	if err != nil || r.Remaining() > 7 || !r.RemainingZero() {
		return 0, fmt.Errorf("lcsap: invalid accuracy fulfilment indicator")
	}
	return AccuracyFulfillmentIndicator(v), nil
}

// PositioningData is the supported Positioning-Data-Set (root, non-GNSS) and
// GNSS-Positioning-Data-Set. Both are one-octet-per-entry method/usage
// encodings specified by TS 29.171 clause 7.4.13.
type PositioningData struct{ methods, gnssMethods []byte }

func NewECIDPositioningData() PositioningData {
	// method 00010 (E-CID), usage 011 (results used to generate location).
	return PositioningData{methods: []byte{0x13}}
}

func NewOTDOAPositioningData() PositioningData {
	// method 00100 (OTDOA), usage 011 (results used to generate location).
	return PositioningData{methods: []byte{0x23}}
}

func NewAGNSSPositioningData() PositioningData {
	// GNSS method 00 (UE-based), GNSS ID 000 (GPS), usage 011 (results used
	// to generate location).
	return PositioningData{gnssMethods: []byte{0x03}}
}

func (v PositioningData) Methods() []byte { return append([]byte(nil), v.methods...) }

func (v PositioningData) GNSSMethods() []byte { return append([]byte(nil), v.gnssMethods...) }

// validMethod reports whether the root Positioning-Method-And-Usage octet's
// method (bits 8-4) and usage (bits 3-1) fields are both defined per TS
// 29.171 clause 7.4.13: method one of Cell-ID (00000), E-CID (00010), OTDOA
// (00100), U-TDOA (01000); usage 0..4.
func validMethod(b byte) bool {
	switch b >> 3 {
	case 0, 2, 4, 8:
	default:
		return false
	}
	return b&7 <= 4
}

// validGNSSMethod reports whether the GNSS-Positioning-Method-And-Usage
// octet's method (bits 8-7), GNSS ID (bits 6-4), and usage (bits 3-1) fields
// are all defined per TS 29.171 clause 7.4.13: method UE-Based/UE-Assisted/
// Conventional (00/01/10); GNSS ID GPS..GLONASS (000..101); usage 0..4.
func validGNSSMethod(b byte) bool {
	return b>>6 <= 2 && (b>>3)&7 <= 5 && b&7 <= 4
}

func (v PositioningData) Validate() error {
	if len(v.methods) == 0 && len(v.gnssMethods) == 0 {
		return fmt.Errorf("lcsap: positioning data must contain at least one method or GNSS method")
	}
	if len(v.methods) > 9 || len(v.gnssMethods) > 9 {
		return fmt.Errorf("lcsap: positioning data set must contain 1..9 methods")
	}
	for _, method := range v.methods {
		if !validMethod(method) {
			return fmt.Errorf("lcsap: unsupported positioning method or usage")
		}
	}
	for _, method := range v.gnssMethods {
		if !validGNSSMethod(method) {
			return fmt.Errorf("lcsap: unsupported GNSS positioning method or usage")
		}
	}
	return nil
}

func (v PositioningData) EncodeAPER() ([]byte, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	w := aper.NewWriter()
	// Extensible SEQUENCE root and three root OPTIONAL fields: positioning
	// data set, GNSS positioning data set, IE extensions.
	presence := []int64{0, 0, 0, 0}
	if len(v.methods) > 0 {
		presence[1] = 1
	}
	if len(v.gnssMethods) > 0 {
		presence[2] = 1
	}
	for _, bit := range presence {
		if err := aper.PutConstrained(w, bit, 0, 1); err != nil {
			return nil, err
		}
	}
	if len(v.methods) > 0 {
		if err := aper.PutConstrained(w, int64(len(v.methods)), 1, 9); err != nil {
			return nil, err
		}
		for _, method := range v.methods {
			if err := aper.PutFixedOctets(w, []byte{method}, 1); err != nil {
				return nil, err
			}
		}
	}
	if len(v.gnssMethods) > 0 {
		if err := aper.PutConstrained(w, int64(len(v.gnssMethods)), 1, 9); err != nil {
			return nil, err
		}
		for _, method := range v.gnssMethods {
			if err := aper.PutFixedOctets(w, []byte{method}, 1); err != nil {
				return nil, err
			}
		}
	}
	return w.Bytes(), nil
}

func DecodePositioningData(b []byte) (PositioningData, error) {
	r := aper.NewReader(b)
	ext, err := aper.GetConstrained(r, 0, 1)
	if err != nil || ext != 0 {
		return PositioningData{}, fmt.Errorf("lcsap: positioning data extensions unsupported")
	}
	present := [3]bool{}
	for i := range present {
		v, err := aper.GetConstrained(r, 0, 1)
		if err != nil {
			return PositioningData{}, err
		}
		present[i] = v != 0
	}
	if present[2] {
		return PositioningData{}, fmt.Errorf("lcsap: unsupported positioning data fields")
	}
	var v PositioningData
	if present[0] {
		n, err := aper.GetConstrained(r, 1, 9)
		if err != nil {
			return PositioningData{}, err
		}
		v.methods = make([]byte, n)
		for i := range v.methods {
			octets, err := aper.GetFixedOctets(r, 1)
			if err != nil {
				return PositioningData{}, err
			}
			v.methods[i] = octets[0]
		}
	}
	if present[1] {
		n, err := aper.GetConstrained(r, 1, 9)
		if err != nil {
			return PositioningData{}, err
		}
		v.gnssMethods = make([]byte, n)
		for i := range v.gnssMethods {
			octets, err := aper.GetFixedOctets(r, 1)
			if err != nil {
				return PositioningData{}, err
			}
			v.gnssMethods[i] = octets[0]
		}
	}
	if !present[0] && !present[1] {
		return PositioningData{}, fmt.Errorf("lcsap: positioning data must contain at least one method or GNSS method")
	}
	if r.Remaining() > 7 || !r.RemainingZero() {
		return PositioningData{}, fmt.Errorf("lcsap: trailing positioning data")
	}
	if err := v.Validate(); err != nil {
		return PositioningData{}, err
	}
	return v, nil
}
