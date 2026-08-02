package lppa

import (
	"fmt"
	"math"

	"github.com/vectorcore/esmlc/internal/aper"
)

// ECGI is TS 36.455 ECGI: a PLMN identity and a 28-bit E-UTRAN cell identity.
type ECGI struct {
	PLMNIdentity [3]byte
	CellIdentity uint32 // low 28 bits significant
}

// encodeECGI/decodeECGI operate on a shared writer/reader because ECGI is a
// plain nested SEQUENCE field (not an IE value), so it must not be
// independently octet-aligned or open-type wrapped.
func encodeECGI(w *aper.Writer, e ECGI) error {
	if e.CellIdentity > 1<<28-1 {
		return fmt.Errorf("lppa: cell identity out of range")
	}
	if err := extBit(w, 0); err != nil { // ECGI SEQUENCE extension bit
		return err
	}
	if err := extBit(w, 0); err != nil { // iE-Extensions optional-presence bit, absent
		return err
	}
	if err := aper.PutFixedOctets(w, e.PLMNIdentity[:], 3); err != nil {
		return err
	}
	return aper.PutFixedBitString(w, uint64(e.CellIdentity), 28)
}

func decodeECGI(r *aper.Reader) (ECGI, error) {
	ext, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return ECGI{}, err
	}
	if ext != 0 {
		return ECGI{}, fmt.Errorf("lppa: ecgi extension unsupported")
	}
	hasExt, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return ECGI{}, err
	}
	if hasExt != 0 {
		return ECGI{}, fmt.Errorf("lppa: ecgi extensions unsupported")
	}
	plmn, err := aper.GetFixedOctets(r, 3)
	if err != nil {
		return ECGI{}, err
	}
	cellID, err := aper.GetFixedBitString(r, 28)
	if err != nil {
		return ECGI{}, err
	}
	var out ECGI
	copy(out.PLMNIdentity[:], plmn)
	out.CellIdentity = uint32(cellID)
	return out, nil
}

// AccessPointPosition is TS 36.455 E-UTRANAccessPointPosition: a
// decimal-degree coordinate encoded exactly like TS 23.032
// Geographical-Coordinates, plus altitude and horizontal/vertical
// uncertainty, all flattened into one non-nested SEQUENCE with a single
// trailing extension marker (unlike the LCS-AP/LPP GAD shapes, which nest
// separately-extensible sub-SEQUENCEs).
type AccessPointPosition struct {
	Latitude, Longitude    float64 // decimal degrees
	Depth                  bool    // directionOfAltitude: false=height, true=depth
	Altitude               uint16  // 0..32767
	UncertaintySemiMajor   uint8   // 0..127
	UncertaintySemiMinor   uint8   // 0..127
	OrientationOfMajorAxis uint8   // 0..179
	UncertaintyAltitude    uint8   // 0..127
	Confidence             uint8   // 0..100
}

func encodeAccessPointPosition(w *aper.Writer, v AccessPointPosition) error {
	if math.IsNaN(v.Latitude) || math.IsNaN(v.Longitude) || math.IsInf(v.Latitude, 0) || math.IsInf(v.Longitude, 0) ||
		math.Abs(v.Latitude) > 90 || math.Abs(v.Longitude) > 180 {
		return fmt.Errorf("lppa: invalid access point coordinates")
	}
	if v.OrientationOfMajorAxis > 179 || v.Confidence > 100 {
		return fmt.Errorf("lppa: access point position field out of range")
	}
	if err := extBit(w, 0); err != nil { // SEQUENCE extension bit; no optional root members
		return err
	}
	lat := int64(math.Round(math.Abs(v.Latitude) * ((1 << 23) - 1) / 90))
	lon := int64(math.Round(v.Longitude * ((1 << 23) - 1) / 180))
	if err := aper.PutConstrained(w, boolInt(v.Latitude < 0), 0, 1); err != nil {
		return err
	}
	if err := aper.PutConstrained(w, lat, 0, (1<<23)-1); err != nil {
		return err
	}
	if err := aper.PutConstrained(w, lon, -(1 << 23), (1<<23)-1); err != nil {
		return err
	}
	if err := aper.PutConstrained(w, boolInt(v.Depth), 0, 1); err != nil {
		return err
	}
	if err := aper.PutConstrained(w, int64(v.Altitude), 0, 32767); err != nil {
		return err
	}
	if err := aper.PutConstrained(w, int64(v.UncertaintySemiMajor), 0, 127); err != nil {
		return err
	}
	if err := aper.PutConstrained(w, int64(v.UncertaintySemiMinor), 0, 127); err != nil {
		return err
	}
	if err := aper.PutConstrained(w, int64(v.OrientationOfMajorAxis), 0, 179); err != nil {
		return err
	}
	if err := aper.PutConstrained(w, int64(v.UncertaintyAltitude), 0, 127); err != nil {
		return err
	}
	return aper.PutConstrained(w, int64(v.Confidence), 0, 100)
}

func decodeAccessPointPosition(r *aper.Reader) (AccessPointPosition, error) {
	ext, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return AccessPointPosition{}, err
	}
	if ext != 0 {
		return AccessPointPosition{}, fmt.Errorf("lppa: access point position extension unsupported")
	}
	south, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return AccessPointPosition{}, err
	}
	lat, err := aper.GetConstrained(r, 0, (1<<23)-1)
	if err != nil {
		return AccessPointPosition{}, err
	}
	lon, err := aper.GetConstrained(r, -(1 << 23), (1<<23)-1)
	if err != nil {
		return AccessPointPosition{}, err
	}
	depth, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return AccessPointPosition{}, err
	}
	altitude, err := aper.GetConstrained(r, 0, 32767)
	if err != nil {
		return AccessPointPosition{}, err
	}
	semiMajor, err := aper.GetConstrained(r, 0, 127)
	if err != nil {
		return AccessPointPosition{}, err
	}
	semiMinor, err := aper.GetConstrained(r, 0, 127)
	if err != nil {
		return AccessPointPosition{}, err
	}
	orientation, err := aper.GetConstrained(r, 0, 179)
	if err != nil {
		return AccessPointPosition{}, err
	}
	uncertaintyAltitude, err := aper.GetConstrained(r, 0, 127)
	if err != nil {
		return AccessPointPosition{}, err
	}
	confidence, err := aper.GetConstrained(r, 0, 100)
	if err != nil {
		return AccessPointPosition{}, err
	}
	latDeg := float64(lat) * 90 / ((1 << 23) - 1)
	if south != 0 {
		latDeg = -latDeg
	}
	lonDeg := float64(lon) * 180 / ((1 << 23) - 1)
	return AccessPointPosition{
		Latitude: latDeg, Longitude: lonDeg, Depth: depth != 0, Altitude: uint16(altitude),
		UncertaintySemiMajor: uint8(semiMajor), UncertaintySemiMinor: uint8(semiMinor),
		OrientationOfMajorAxis: uint8(orientation), UncertaintyAltitude: uint8(uncertaintyAltitude), Confidence: uint8(confidence),
	}, nil
}

// ECIDMeasurementResult is the bounded root subset of E-CID-MeasurementResult:
// the serving cell identity and TAC are always present; AccessPointPosition
// is nil when the optional field is absent. The optional measuredResults
// field (enhanced-cell-ID AoA/TA/RSRP/RSRQ values) is out of bounded scope
// and is rejected fail-closed if the eNB sets it.
type ECIDMeasurementResult struct {
	ServingCellID       ECGI
	ServingCellTAC      [2]byte
	AccessPointPosition *AccessPointPosition
}

func EncodeECIDMeasurementResult(v ECIDMeasurementResult) ([]byte, error) {
	w := aper.NewWriter()
	if err := extBit(w, 0); err != nil { // E-CID-MeasurementResult SEQUENCE extension bit
		return nil, err
	}
	if err := extBit(w, boolInt(v.AccessPointPosition != nil)); err != nil { // e-UTRANAccessPointPosition presence
		return nil, err
	}
	if err := extBit(w, 0); err != nil { // measuredResults presence: always absent
		return nil, err
	}
	if err := encodeECGI(w, v.ServingCellID); err != nil {
		return nil, err
	}
	if err := aper.PutFixedOctets(w, v.ServingCellTAC[:], 2); err != nil {
		return nil, err
	}
	if v.AccessPointPosition != nil {
		if err := encodeAccessPointPosition(w, *v.AccessPointPosition); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

func DecodeECIDMeasurementResult(b []byte) (ECIDMeasurementResult, error) {
	r := aper.NewReader(b)
	ext, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return ECIDMeasurementResult{}, err
	}
	if ext != 0 {
		return ECIDMeasurementResult{}, fmt.Errorf("lppa: e-cid measurement result extension unsupported")
	}
	hasPosition, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return ECIDMeasurementResult{}, err
	}
	hasMeasured, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return ECIDMeasurementResult{}, err
	}
	if hasMeasured != 0 {
		return ECIDMeasurementResult{}, fmt.Errorf("lppa: e-cid measuredResults unsupported")
	}
	ecgi, err := decodeECGI(r)
	if err != nil {
		return ECIDMeasurementResult{}, err
	}
	tac, err := aper.GetFixedOctets(r, 2)
	if err != nil {
		return ECIDMeasurementResult{}, err
	}
	var out ECIDMeasurementResult
	out.ServingCellID = ecgi
	copy(out.ServingCellTAC[:], tac)
	if hasPosition != 0 {
		pos, err := decodeAccessPointPosition(r)
		if err != nil {
			return ECIDMeasurementResult{}, err
		}
		out.AccessPointPosition = &pos
	}
	if r.Remaining() > 7 || !r.RemainingZero() {
		return ECIDMeasurementResult{}, fmt.Errorf("lppa: trailing e-cid measurement result data")
	}
	return out, nil
}
