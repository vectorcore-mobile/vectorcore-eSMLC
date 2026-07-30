// Package lcsap implements the recovered LCS-AP subset used on SLs.
package lcsap

import (
	"fmt"
	"github.com/vectorcore/esmlc/internal/aper"
	"math"
)

const (
	PPID                                   uint32 = 29
	ProcedureLocationRequest               uint8  = 0
	ProcedureConnectionOrientedInformation uint8  = 1
	ProcedureLocationAbort                 uint8  = 3
	ProcedureReset                         uint8  = 4
	ProcedureErrorIndication               uint8  = 5
	IECorrelationID                        uint16 = 2
	IEECGI                                 uint16 = 4
	IELCSPriority                          uint16 = 9
	IELCSQoS                               uint16 = 10
	IELCSCause                             uint16 = 11
	IELocationEstimate                     uint16 = 12
	IELocationType                         uint16 = 13
	IEPayload                              uint16 = 1
	IEPayloadType                          uint16 = 15
	IEUEPositioningCapability              uint16 = 20
)

type Category uint8

const (
	Initiating Category = iota
	Successful
	Unsuccessful
)

type IE struct {
	ID          uint16
	Criticality aper.Criticality
	Value       []byte
}

// Cause is the supported root subset of the extensible LCS-Cause CHOICE. The
// values are complete aligned-PER encodings of the corresponding root cause.
// Unknown cause alternatives remain unsupported rather than being mislabelled.
type Cause uint8

const (
	CauseRadioNetworkUnspecified Cause = iota
	CauseProtocolUnspecified
	CauseMiscUnspecified
)

type PDU struct {
	Category    Category
	Procedure   uint8
	Criticality aper.Criticality
	IEs         []IE
}

// LocationRequest is the verified subset currently consumed by positioning
// orchestration. Priority is nil when the conditional LCS-Priority IE is absent.
type LocationRequest struct {
	Correlation  [4]byte
	LocationType uint8
	ECGI         [7]byte
	Priority     *uint8
	QoS          *QoS
	LPPSupported *bool
}

// QoS is the root Release-16 LCS-QoS subset. Accuracy values are the
// standards-defined 0..127 accuracy codes, not locally invented metre values.
type QoS struct {
	HorizontalAccuracy *uint8
	VerticalRequested  *bool
	VerticalAccuracy   *uint8
	ResponseTime       *uint8 // 0: low-delay, 1: delay-tolerant
}

// ConnectionOriented carries an opaque LPP or LPPa APDU.  The current MME
// uses payload type 0 for LPP and does not include Routing-ID on this SLs
// procedure; Routing-ID belongs to its separate UE-associated LPPa path.
type ConnectionOriented struct {
	Correlation [4]byte
	Payload     []byte
	PayloadType uint8 // 0: LPP, 1: LPPa
}

func known(id uint16) bool {
	switch id {
	case IECorrelationID, IEECGI, IELCSPriority, IELCSQoS, IELCSCause, IELocationEstimate, IELocationType, IEPayload, IEPayloadType, IEUEPositioningCapability:
		return true
	}
	return false
}
func Encode(p PDU) ([]byte, error) {
	if p.Category > Unsuccessful || len(p.IEs) > 1024 {
		return nil, fmt.Errorf("lcsap: invalid PDU")
	}
	body := aper.NewWriter()
	if e := aper.PutConstrained(body, int64(len(p.IEs)), 0, 65535); e != nil {
		return nil, e
	}
	for _, ie := range p.IEs {
		if e := aper.PutConstrained(body, int64(ie.ID), 0, 65535); e != nil {
			return nil, e
		}
		if e := aper.PutCriticality(body, ie.Criticality); e != nil {
			return nil, e
		}
		if e := aper.PutOpenType(body, ie.Value); e != nil {
			return nil, e
		}
	}
	w := aper.NewWriter()
	wbits(w, uint64(p.Category), 2)
	wOctet(w, p.Procedure)
	if e := aper.PutCriticality(w, p.Criticality); e != nil {
		return nil, e
	}
	if e := aper.PutOpenType(w, body.Bytes()); e != nil {
		return nil, e
	}
	return w.Bytes(), nil
}

// tiny wrappers retain bit-level outer CHOICE without exporting internals.
func wbits(w *aper.Writer, v uint64, n int) { // outer choice has exactly two bits
	for i := n - 1; i >= 0; i-- {
		wBit(w, uint8(v>>uint(i)&1))
	}
}
func wBit(w *aper.Writer, b uint8) { // encode root CHOICE through constrained 0..2
	// Two-bit values map to the same primitive representation.
	if b == 0 {
		_ = aper.PutConstrained(w, 0, 0, 1)
	} else {
		_ = aper.PutConstrained(w, 1, 0, 1)
	}
}
func wOctet(w *aper.Writer, b byte) { _ = aper.PutConstrained(w, int64(b), 0, 255) }
func Decode(b []byte) (PDU, error) {
	if len(b) == 0 || len(b) > aper.MaxOpenType {
		return PDU{}, fmt.Errorf("lcsap: invalid size")
	}
	r := aper.NewReader(b)
	cat, e := aper.GetConstrained(r, 0, 3)
	if e != nil || cat > 2 {
		return PDU{}, fmt.Errorf("lcsap: invalid PDU choice")
	}
	proc, e := aper.GetConstrained(r, 0, 255)
	if e != nil {
		return PDU{}, e
	}
	crit, e := aper.GetCriticality(r)
	if e != nil {
		return PDU{}, e
	}
	body, e := aper.GetOpenType(r)
	if e != nil {
		return PDU{}, e
	}
	if r.Remaining() != 0 {
		return PDU{}, fmt.Errorf("lcsap: trailing PDU data")
	}
	br := aper.NewReader(body)
	n, e := aper.GetConstrained(br, 0, 65535)
	if e != nil || n > 1024 {
		return PDU{}, fmt.Errorf("lcsap: invalid IE count")
	}
	p := PDU{Category: Category(cat), Procedure: uint8(proc), Criticality: crit, IEs: make([]IE, 0, n)}
	for i := int64(0); i < n; i++ {
		id, e := aper.GetConstrained(br, 0, 65535)
		if e != nil {
			return PDU{}, e
		}
		c, e := aper.GetCriticality(br)
		if e != nil {
			return PDU{}, e
		}
		v, e := aper.GetOpenType(br)
		if e != nil {
			return PDU{}, e
		}
		p.IEs = append(p.IEs, IE{uint16(id), c, v})
	}
	if br.Remaining() > 7 {
		return PDU{}, fmt.Errorf("lcsap: trailing procedure data")
	}
	return p, nil
}
func Correlation(p PDU) ([4]byte, error) {
	var out [4]byte
	found := false
	for _, ie := range p.IEs {
		if ie.ID == IECorrelationID {
			if found || len(ie.Value) != 4 {
				return out, fmt.Errorf("lcsap: invalid correlation ID")
			}
			copy(out[:], ie.Value)
			found = true
		}
	}
	if !found {
		return out, fmt.Errorf("lcsap: missing correlation ID")
	}
	return out, nil
}
func ValidateLocationRequest(p PDU) ([4]byte, error) {
	v, e := DecodeLocationRequest(p)
	return v.Correlation, e
}

// DecodeLocationRequest validates and retains the mandatory request fields and
// the conditional LCS-Priority octet. Other conditional request semantics are
// deliberately not inferred from opaque IEs.
func DecodeLocationRequest(p PDU) (LocationRequest, error) {
	var out LocationRequest
	id, e := Correlation(p)
	out.Correlation = id
	if e != nil {
		return out, e
	}
	if p.Category != Initiating || p.Procedure != ProcedureLocationRequest || p.Criticality != aper.Reject {
		return out, fmt.Errorf("lcsap: unsupported location request")
	}
	seen := map[uint16]bool{}
	hasType, hasECGI := false, false
	for _, ie := range p.IEs {
		if seen[ie.ID] {
			return out, fmt.Errorf("lcsap: duplicate IE")
		}
		seen[ie.ID] = true
		if !known(ie.ID) && ie.Criticality == aper.Reject {
			return out, fmt.Errorf("lcsap: unknown reject IE")
		}
		switch ie.ID {
		case IECorrelationID:
			if ie.Criticality != aper.Reject || len(ie.Value) != 4 {
				return out, fmt.Errorf("lcsap: invalid correlation")
			}
		case IELocationType:
			if ie.Criticality != aper.Reject || len(ie.Value) != 1 || ie.Value[0] > 1 {
				return out, fmt.Errorf("lcsap: invalid location type")
			}
			out.LocationType = ie.Value[0]
			hasType = true
		case IEECGI:
			if ie.Criticality != aper.Ignore || len(ie.Value) != 7 {
				return out, fmt.Errorf("lcsap: invalid ECGI")
			}
			copy(out.ECGI[:], ie.Value)
			hasECGI = true
		case IELCSPriority:
			if ie.Criticality != aper.Ignore || len(ie.Value) != 1 {
				return out, fmt.Errorf("lcsap: invalid LCS priority")
			}
			priority := ie.Value[0]
			out.Priority = &priority
		case IELCSQoS:
			if ie.Criticality != aper.Ignore {
				return out, fmt.Errorf("lcsap: invalid LCS QoS criticality")
			}
			qos, err := DecodeQoS(ie.Value)
			if err != nil {
				return out, fmt.Errorf("lcsap: invalid LCS QoS: %w", err)
			}
			out.QoS = &qos
		case IEUEPositioningCapability:
			if ie.Criticality != aper.Ignore {
				return out, fmt.Errorf("lcsap: invalid UE positioning capability criticality")
			}
			supported, err := DecodeUEPositioningCapability(ie.Value)
			if err != nil {
				return out, fmt.Errorf("lcsap: invalid UE positioning capability: %w", err)
			}
			out.LPPSupported = &supported
		}
	}
	if !hasType || !hasECGI {
		return out, fmt.Errorf("lcsap: missing mandatory IE")
	}
	return out, nil
}

// DecodeQoS decodes the non-extension root of LCS-QoS. Extension data is
// rejected fail-closed because no extension semantics are implemented.
func DecodeQoS(b []byte) (QoS, error) {
	r := aper.NewReader(b)
	ext, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return QoS{}, err
	}
	if ext != 0 {
		return QoS{}, fmt.Errorf("extensions unsupported")
	}
	present := [4]bool{}
	for i := range present {
		v, err := aper.GetConstrained(r, 0, 1)
		if err != nil {
			return QoS{}, err
		}
		present[i] = v != 0
	}
	var out QoS
	if present[0] {
		v, err := aper.GetConstrained(r, 0, 127)
		if err != nil {
			return QoS{}, err
		}
		x := uint8(v)
		out.HorizontalAccuracy = &x
	}
	if present[1] {
		v, err := aper.GetConstrained(r, 0, 1)
		if err != nil {
			return QoS{}, err
		}
		x := v != 0
		out.VerticalRequested = &x
	}
	if present[2] {
		v, err := aper.GetConstrained(r, 0, 127)
		if err != nil {
			return QoS{}, err
		}
		x := uint8(v)
		out.VerticalAccuracy = &x
	}
	if present[3] {
		ext, err := aper.GetConstrained(r, 0, 1)
		if err != nil {
			return QoS{}, err
		}
		if ext != 0 {
			return QoS{}, fmt.Errorf("response-time extensions unsupported")
		}
		v, err := aper.GetConstrained(r, 0, 1)
		if err != nil {
			return QoS{}, err
		}
		x := uint8(v)
		out.ResponseTime = &x
	}
	if r.Remaining() > 7 {
		return QoS{}, fmt.Errorf("trailing QoS data")
	}
	return out, nil
}

// EncodeQoS exists for deterministic LCS-AP fixtures and tests. Production
// request handling only consumes this structure.
func EncodeQoS(v QoS) ([]byte, error) {
	w := aper.NewWriter()
	if err := aper.PutConstrained(w, 0, 0, 1); err != nil {
		return nil, err
	}
	present := []bool{v.HorizontalAccuracy != nil, v.VerticalRequested != nil, v.VerticalAccuracy != nil, v.ResponseTime != nil}
	for _, p := range present {
		if err := aper.PutConstrained(w, boolInt(p), 0, 1); err != nil {
			return nil, err
		}
	}
	if v.HorizontalAccuracy != nil {
		if err := aper.PutConstrained(w, int64(*v.HorizontalAccuracy), 0, 127); err != nil {
			return nil, err
		}
	}
	if v.VerticalRequested != nil {
		if err := aper.PutConstrained(w, boolInt(*v.VerticalRequested), 0, 1); err != nil {
			return nil, err
		}
	}
	if v.VerticalAccuracy != nil {
		if err := aper.PutConstrained(w, int64(*v.VerticalAccuracy), 0, 127); err != nil {
			return nil, err
		}
	}
	if v.ResponseTime != nil {
		if *v.ResponseTime > 1 {
			return nil, fmt.Errorf("invalid response time")
		}
		if err := aper.PutConstrained(w, 0, 0, 1); err != nil {
			return nil, err
		}
		if err := aper.PutConstrained(w, int64(*v.ResponseTime), 0, 1); err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

func DecodeUEPositioningCapability(b []byte) (bool, error) {
	r := aper.NewReader(b)
	ext, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return false, err
	}
	if ext != 0 {
		return false, fmt.Errorf("extensions unsupported")
	}
	v, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return false, err
	}
	if r.Remaining() > 7 {
		return false, fmt.Errorf("trailing capability data")
	}
	return v != 0, nil
}

func EncodeUEPositioningCapability(lpp bool) ([]byte, error) {
	w := aper.NewWriter()
	if err := aper.PutConstrained(w, 0, 0, 1); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, boolInt(lpp), 0, 1); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}
func LocationEstimate(lat, lon float64, uncertainty uint8) ([]byte, error) {
	if math.IsNaN(lat) || math.IsNaN(lon) || math.IsInf(lat, 0) || math.IsInf(lon, 0) || math.Abs(lat) > 90 || math.Abs(lon) > 180 {
		return nil, fmt.Errorf("lcsap: invalid coordinates")
	}
	la := uint32(math.Round(math.Abs(lat) * (1 << 23) / 90))
	lo := int32(math.Round(lon * (1 << 23) / 360))
	return []byte{0, byte((boolByte(lat < 0) << 7) | byte(la>>16&0x7f)), byte(la >> 8), byte(la), byte(uint32(lo) >> 16), byte(uint32(lo) >> 8), byte(lo), uncertainty}, nil
}
func boolByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}
func boolInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
func LocationResponse(id [4]byte, lat, lon float64, uncertainty uint8) ([]byte, error) {
	v, e := LocationEstimate(lat, lon, uncertainty)
	if e != nil {
		return nil, e
	}
	return Encode(PDU{Successful, ProcedureLocationRequest, aper.Reject, []IE{{IECorrelationID, aper.Reject, id[:]}, {IELocationEstimate, aper.Reject, v}}})
}
func Failure(id [4]byte, cause byte) ([]byte, error) {
	return Encode(PDU{Unsuccessful, ProcedureLocationRequest, aper.Reject, []IE{{IECorrelationID, aper.Reject, id[:]}, {IELCSCause, aper.Ignore, []byte{cause}}}})
}

func FailureWithCause(id [4]byte, cause Cause) ([]byte, error) {
	var wire byte
	switch cause {
	case CauseRadioNetworkUnspecified:
		wire = 0x00 // CHOICE radio-network-layer, extension false, unspecified
	case CauseProtocolUnspecified:
		wire = 0x94 // CHOICE protocol, extension false, root enum unspecified
	case CauseMiscUnspecified:
		wire = 0xcc // CHOICE misc, extension false, root enum unspecified
	default:
		return nil, fmt.Errorf("lcsap: unsupported LCS cause")
	}
	return Failure(id, wire)
}

// EncodeConnectionOriented creates the Release-16 initiating PDU. TS 29.171
// declares the elementary-procedure criticality as reject.
func EncodeConnectionOriented(v ConnectionOriented, maxPayload int) ([]byte, error) {
	if len(v.Payload) == 0 || len(v.Payload) > maxPayload {
		return nil, fmt.Errorf("lcsap: invalid connection-oriented payload length")
	}
	if v.PayloadType > 1 {
		return nil, fmt.Errorf("lcsap: unsupported payload type")
	}
	return Encode(PDU{Category: Initiating, Procedure: ProcedureConnectionOrientedInformation, Criticality: aper.Reject, IEs: []IE{
		{ID: IECorrelationID, Criticality: aper.Reject, Value: v.Correlation[:]},
		{ID: IEPayloadType, Criticality: aper.Reject, Value: []byte{v.PayloadType}},
		{ID: IEPayload, Criticality: aper.Reject, Value: append([]byte(nil), v.Payload...)},
	}})
}

func DecodeConnectionOriented(p PDU, maxPayload int) (ConnectionOriented, error) {
	var out ConnectionOriented
	if p.Category != Initiating || p.Procedure != ProcedureConnectionOrientedInformation || (p.Criticality != aper.Reject && p.Criticality != aper.Ignore) {
		return out, fmt.Errorf("lcsap: invalid connection-oriented procedure")
	}
	seen := map[uint16]bool{}
	haveID, haveType, havePayload := false, false, false
	for _, ie := range p.IEs {
		if seen[ie.ID] {
			return out, fmt.Errorf("lcsap: duplicate connection-oriented IE %d", ie.ID)
		}
		seen[ie.ID] = true
		if !known(ie.ID) && ie.Criticality == aper.Reject {
			return out, fmt.Errorf("lcsap: unknown reject connection-oriented IE %d", ie.ID)
		}
		switch ie.ID {
		case IECorrelationID:
			if ie.Criticality != aper.Reject || len(ie.Value) != 4 {
				return out, fmt.Errorf("lcsap: invalid connection-oriented correlation ID")
			}
			copy(out.Correlation[:], ie.Value)
			haveID = true
		case IEPayloadType:
			if ie.Criticality != aper.Reject || len(ie.Value) != 1 || ie.Value[0] > 1 {
				return out, fmt.Errorf("lcsap: invalid connection-oriented payload type")
			}
			out.PayloadType = ie.Value[0]
			haveType = true
		case IEPayload:
			if ie.Criticality != aper.Reject || len(ie.Value) == 0 || len(ie.Value) > maxPayload {
				return out, fmt.Errorf("lcsap: invalid connection-oriented payload")
			}
			out.Payload = append([]byte(nil), ie.Value...)
			havePayload = true
		}
	}
	if !haveID || !haveType || !havePayload {
		return out, fmt.Errorf("lcsap: missing connection-oriented mandatory IE")
	}
	return out, nil
}
