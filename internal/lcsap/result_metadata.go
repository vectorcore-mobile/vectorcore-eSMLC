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

// PositioningData is the supported root non-GNSS data set. The one-octet
// method/usage encoding is specified by TS 29.171 clause 7.4.13.
type PositioningData struct{ methods []byte }

func NewECIDPositioningData() PositioningData {
	// method 00010 (E-CID), usage 011 (results used to generate location).
	return PositioningData{methods: []byte{0x13}}
}

func (v PositioningData) Methods() []byte { return append([]byte(nil), v.methods...) }

func (v PositioningData) Validate() error {
	if len(v.methods) < 1 || len(v.methods) > 9 {
		return fmt.Errorf("lcsap: positioning data set must contain 1..9 methods")
	}
	for _, method := range v.methods {
		if method>>3 != 2 || method&7 > 4 {
			return fmt.Errorf("lcsap: unsupported positioning method or usage")
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
	for _, bit := range []int64{0, 1, 0, 0} {
		if err := aper.PutConstrained(w, bit, 0, 1); err != nil {
			return nil, err
		}
	}
	if err := aper.PutConstrained(w, int64(len(v.methods)), 1, 9); err != nil {
		return nil, err
	}
	for _, method := range v.methods {
		if err := aper.PutFixedOctets(w, []byte{method}, 1); err != nil {
			return nil, err
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
	if !present[0] || present[1] || present[2] {
		return PositioningData{}, fmt.Errorf("lcsap: unsupported positioning data fields")
	}
	n, err := aper.GetConstrained(r, 1, 9)
	if err != nil {
		return PositioningData{}, err
	}
	v := PositioningData{methods: make([]byte, n)}
	for i := range v.methods {
		octets, err := aper.GetFixedOctets(r, 1)
		if err != nil {
			return PositioningData{}, err
		}
		v.methods[i] = octets[0]
	}
	if r.Remaining() > 7 || !r.RemainingZero() {
		return PositioningData{}, fmt.Errorf("lcsap: trailing positioning data")
	}
	if err := v.Validate(); err != nil {
		return PositioningData{}, err
	}
	return v, nil
}
