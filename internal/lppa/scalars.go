package lppa

import (
	"fmt"

	"github.com/vectorcore/esmlc/internal/aper"
)

// EncodeMeasurementID encodes Measurement-ID ::= INTEGER (1..15, ...).
func EncodeMeasurementID(v uint8) ([]byte, error) {
	if v < 1 || v > 15 {
		return nil, fmt.Errorf("lppa: measurement-id out of range")
	}
	w := aper.NewWriter()
	if err := extBit(w, 0); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, int64(v), 1, 15); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func DecodeMeasurementID(b []byte) (uint8, error) {
	r := aper.NewReader(b)
	ext, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return 0, err
	}
	if ext != 0 {
		return 0, fmt.Errorf("lppa: measurement-id extension unsupported")
	}
	v, err := aper.GetConstrained(r, 1, 15)
	if err != nil {
		return 0, err
	}
	if r.Remaining() > 7 || !r.RemainingZero() {
		return 0, fmt.Errorf("lppa: trailing measurement-id data")
	}
	return uint8(v), nil
}

// ReportCharacteristics ::= ENUMERATED { onDemand, periodic, ... }.
type ReportCharacteristics uint8

const (
	ReportOnDemand ReportCharacteristics = iota
	ReportPeriodic
)

func EncodeReportCharacteristics(v ReportCharacteristics) ([]byte, error) {
	if v > ReportPeriodic {
		return nil, fmt.Errorf("lppa: unsupported report characteristics")
	}
	w := aper.NewWriter()
	if err := extBit(w, 0); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, int64(v), 0, 1); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func DecodeReportCharacteristics(b []byte) (ReportCharacteristics, error) {
	r := aper.NewReader(b)
	ext, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return 0, err
	}
	if ext != 0 {
		return 0, fmt.Errorf("lppa: report characteristics extension unsupported")
	}
	v, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return 0, err
	}
	if r.Remaining() > 7 || !r.RemainingZero() {
		return 0, fmt.Errorf("lppa: trailing report characteristics data")
	}
	return ReportCharacteristics(v), nil
}

// MeasurementPeriodicity ::= ENUMERATED { ms120, ms240, ms480, ms640, ms1024,
// ms2048, ms5120, ms10240, min1, min6, min12, min30, min60, ... }.
type MeasurementPeriodicity uint8

const (
	PeriodicityMS120 MeasurementPeriodicity = iota
	PeriodicityMS240
	PeriodicityMS480
	PeriodicityMS640
	PeriodicityMS1024
	PeriodicityMS2048
	PeriodicityMS5120
	PeriodicityMS10240
	PeriodicityMin1
	PeriodicityMin6
	PeriodicityMin12
	PeriodicityMin30
	PeriodicityMin60
)

func EncodeMeasurementPeriodicity(v MeasurementPeriodicity) ([]byte, error) {
	if v > PeriodicityMin60 {
		return nil, fmt.Errorf("lppa: unsupported measurement periodicity")
	}
	w := aper.NewWriter()
	if err := extBit(w, 0); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, int64(v), 0, 12); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func DecodeMeasurementPeriodicity(b []byte) (MeasurementPeriodicity, error) {
	r := aper.NewReader(b)
	ext, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return 0, err
	}
	if ext != 0 {
		return 0, fmt.Errorf("lppa: measurement periodicity extension unsupported")
	}
	v, err := aper.GetConstrained(r, 0, 12)
	if err != nil {
		return 0, err
	}
	if r.Remaining() > 7 || !r.RemainingZero() {
		return 0, fmt.Errorf("lppa: trailing measurement periodicity data")
	}
	return MeasurementPeriodicity(v), nil
}

// MeasurementQuantityValue ::= ENUMERATED { cell-ID, angleOfArrival,
// timingAdvanceType1, timingAdvanceType2, rSRP, rSRQ, ... }.
type MeasurementQuantityValue uint8

const (
	QuantityCellID MeasurementQuantityValue = iota
	QuantityAngleOfArrival
	QuantityTimingAdvanceType1
	QuantityTimingAdvanceType2
	QuantityRSRP
	QuantityRSRQ
)

// EncodeMeasurementQuantities encodes MeasurementQuantities: a SEQUENCE
// (SIZE(1..maxNoMeas)) OF ProtocolIE-Single-Container wrapping a
// MeasurementQuantities-Item per requested quantity. internal/lppa only ever
// requests QuantityCellID (the E-CID positioning method needs nothing else),
// but decodes any root enumerated value.
func EncodeMeasurementQuantities(values []MeasurementQuantityValue) ([]byte, error) {
	if len(values) < 1 || len(values) > 63 {
		return nil, fmt.Errorf("lppa: measurement quantities count out of range")
	}
	w := aper.NewWriter()
	if err := aper.PutConstrained(w, int64(len(values)), 1, 63); err != nil {
		return nil, err
	}
	for _, v := range values {
		if v > QuantityRSRQ {
			return nil, fmt.Errorf("lppa: unsupported measurement quantity value")
		}
		if err := aper.PutConstrained(w, int64(IEMeasurementQuantitiesItem), 0, maxProtocolIEs); err != nil {
			return nil, err
		}
		if err := aper.PutCriticality(w, aper.Reject); err != nil {
			return nil, err
		}
		item := aper.NewWriter()
		if err := extBit(item, 0); err != nil { // MeasurementQuantities-Item SEQUENCE extension bit
			return nil, err
		}
		if err := extBit(item, 0); err != nil { // iE-Extensions optional-presence bit, absent
			return nil, err
		}
		if err := extBit(item, 0); err != nil { // MeasurementQuantitiesValue's own extension bit
			return nil, err
		}
		if err := aper.PutConstrained(item, int64(v), 0, 5); err != nil {
			return nil, err
		}
		if err := aper.PutOpenType(w, item.Bytes()); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

func DecodeMeasurementQuantities(b []byte) ([]MeasurementQuantityValue, error) {
	r := aper.NewReader(b)
	n, err := aper.GetConstrained(r, 1, 63)
	if err != nil {
		return nil, err
	}
	out := make([]MeasurementQuantityValue, 0, n)
	for i := int64(0); i < n; i++ {
		id, err := aper.GetConstrained(r, 0, maxProtocolIEs)
		if err != nil {
			return nil, err
		}
		if id != int64(IEMeasurementQuantitiesItem) {
			return nil, fmt.Errorf("lppa: unexpected measurement quantities item id %d", id)
		}
		if _, err := aper.GetCriticality(r); err != nil {
			return nil, err
		}
		itemBytes, err := aper.GetOpenType(r)
		if err != nil {
			return nil, err
		}
		item := aper.NewReader(itemBytes)
		ext, err := aper.GetConstrained(item, 0, 1)
		if err != nil {
			return nil, err
		}
		if ext != 0 {
			return nil, fmt.Errorf("lppa: measurement quantities item extension unsupported")
		}
		hasExt, err := aper.GetConstrained(item, 0, 1)
		if err != nil {
			return nil, err
		}
		if hasExt != 0 {
			return nil, fmt.Errorf("lppa: measurement quantities item extensions unsupported")
		}
		vExt, err := aper.GetConstrained(item, 0, 1)
		if err != nil {
			return nil, err
		}
		if vExt != 0 {
			return nil, fmt.Errorf("lppa: measurement quantities value extension unsupported")
		}
		v, err := aper.GetConstrained(item, 0, 5)
		if err != nil {
			return nil, err
		}
		if item.Remaining() > 7 || !item.RemainingZero() {
			return nil, fmt.Errorf("lppa: trailing measurement quantities item data")
		}
		out = append(out, MeasurementQuantityValue(v))
	}
	if r.Remaining() > 7 || !r.RemainingZero() {
		return nil, fmt.Errorf("lppa: trailing measurement quantities data")
	}
	return out, nil
}

// Cell-Portion-ID ::= INTEGER (0..255, ..., 256..4095). The extension range
// is out of bounded scope: a set extension bit is rejected fail-closed.
func EncodeCellPortionID(v uint16) ([]byte, error) {
	if v > 255 {
		return nil, fmt.Errorf("lppa: cell-portion-id extension range unsupported")
	}
	w := aper.NewWriter()
	if err := extBit(w, 0); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, int64(v), 0, 255); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func DecodeCellPortionID(b []byte) (uint16, error) {
	r := aper.NewReader(b)
	ext, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return 0, err
	}
	if ext != 0 {
		return 0, fmt.Errorf("lppa: cell-portion-id extension range unsupported")
	}
	v, err := aper.GetConstrained(r, 0, 255)
	if err != nil {
		return 0, err
	}
	if r.Remaining() > 7 || !r.RemainingZero() {
		return 0, fmt.Errorf("lppa: trailing cell-portion-id data")
	}
	return uint16(v), nil
}
