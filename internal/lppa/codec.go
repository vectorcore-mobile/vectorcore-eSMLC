// Package lppa implements the recovered TS 36.455 LPPa subset used for E-CID
// measurement between the E-SMLC and an eNB. LPPa rides the same SLs/SCTP
// connection as LPP (TS 29.171 Connection-Oriented Information, payload type
// 1); the MME relays it transparently to the eNB over S1AP without decoding
// it, so this package needs no new transport stack of its own.
//
// Only the E-CID Measurement Initiation/Report procedure family is
// implemented, matching the bounded root-only scope used throughout this
// repository: unimplemented optional IEs and extension data are rejected
// fail-closed on decode rather than silently ignored. See
// docs/specs/asn1/lppa/r16.0.0/lppa-ecid-subset.asn1 for the recovered ASN.1
// and the explicit list of what is out of scope.
package lppa

import (
	"fmt"

	"github.com/vectorcore/esmlc/internal/aper"
)

// Elementary procedure codes (LPPA-Constants), restricted to the procedures
// this package implements.
const (
	ProcedureECIDMeasurementInitiation        uint8 = 2
	ProcedureECIDMeasurementFailureIndication uint8 = 3
	ProcedureECIDMeasurementReport            uint8 = 4
	ProcedureECIDMeasurementTermination       uint8 = 5
)

// IE ids (LPPA-Constants) used by the E-CID measurement message set.
const (
	IECause                     uint16 = 0
	IECriticalityDiagnostics    uint16 = 1
	IEESMLCMeasurementID        uint16 = 2 // id-E-SMLC-UE-Measurement-ID
	IEReportCharacteristics     uint16 = 3
	IEMeasurementPeriodicity    uint16 = 4
	IEMeasurementQuantities     uint16 = 5
	IEENBMeasurementID          uint16 = 6 // id-eNB-UE-Measurement-ID
	IECIDMeasurementResult      uint16 = 7
	IEMeasurementQuantitiesItem uint16 = 11
	IECellPortionID             uint16 = 14
)

// maxProtocolIEs is LPPA-CommonDataTypes maxProtocolIEs, the width of every
// ProtocolIE-Container id/count field.
const maxProtocolIEs = 65535

// Category is the LPPA-PDU root CHOICE selector.
type Category uint8

const (
	Initiating Category = iota
	Successful
	Unsuccessful
)

// IE is one ProtocolIE-Field: an id/criticality pair and its already-encoded
// open-type value, matching internal/lcsap's IE shape.
type IE struct {
	ID          uint16
	Criticality aper.Criticality
	Value       []byte
}

// PDU is one LPPA-PDU: InitiatingMessage/SuccessfulOutcome/UnsuccessfulOutcome
// share this shape, distinguished by Category.
type PDU struct {
	Category      Category
	ProcedureCode uint8
	Criticality   aper.Criticality
	TransactionID uint16
	IEs           []IE
}

func extBit(w *aper.Writer, b int64) error { return aper.PutConstrained(w, b, 0, 1) }

func boolInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// Encode serialises a PDU. InitiatingMessage/SuccessfulOutcome/
// UnsuccessfulOutcome are non-extensible SEQUENCEs with no OPTIONAL members,
// so they carry no preamble bits of their own; every E-CID message body is
// an extensible SEQUENCE with a single mandatory protocolIEs field (a
// SEQUENCE OF ProtocolIE-Field), so its only preamble bit is the extension
// marker.
func Encode(p PDU) ([]byte, error) {
	if p.Category > Unsuccessful || p.TransactionID > 32767 || len(p.IEs) > maxProtocolIEs {
		return nil, fmt.Errorf("lppa: invalid PDU")
	}
	body := aper.NewWriter()
	if err := extBit(body, 0); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(body, int64(len(p.IEs)), 0, maxProtocolIEs); err != nil {
		return nil, err
	}
	for _, ie := range p.IEs {
		if err := aper.PutConstrained(body, int64(ie.ID), 0, maxProtocolIEs); err != nil {
			return nil, err
		}
		if err := aper.PutCriticality(body, ie.Criticality); err != nil {
			return nil, err
		}
		if err := aper.PutOpenType(body, ie.Value); err != nil {
			return nil, err
		}
	}
	w := aper.NewWriter()
	// LPPA-PDU is an extensible CHOICE over 3 root alternatives.
	if err := extBit(w, 0); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, int64(p.Category), 0, 2); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, int64(p.ProcedureCode), 0, 255); err != nil {
		return nil, err
	}
	if err := aper.PutCriticality(w, p.Criticality); err != nil {
		return nil, err
	}
	if err := aper.PutConstrained(w, int64(p.TransactionID), 0, 32767); err != nil {
		return nil, err
	}
	if err := aper.PutOpenType(w, body.Bytes()); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// Decode parses the LPPA-PDU envelope and its protocolIEs list without
// interpreting individual IE values; message-specific Decode functions
// validate the IE set for a given procedure.
func Decode(b []byte) (PDU, error) {
	if len(b) == 0 || len(b) > aper.MaxOpenType {
		return PDU{}, fmt.Errorf("lppa: invalid size")
	}
	r := aper.NewReader(b)
	ext, err := aper.GetConstrained(r, 0, 1)
	if err != nil || ext != 0 {
		return PDU{}, fmt.Errorf("lppa: PDU extensions unsupported")
	}
	cat, err := aper.GetConstrained(r, 0, 2)
	if err != nil {
		return PDU{}, err
	}
	proc, err := aper.GetConstrained(r, 0, 255)
	if err != nil {
		return PDU{}, err
	}
	crit, err := aper.GetCriticality(r)
	if err != nil {
		return PDU{}, err
	}
	txn, err := aper.GetConstrained(r, 0, 32767)
	if err != nil {
		return PDU{}, err
	}
	body, err := aper.GetOpenType(r)
	if err != nil {
		return PDU{}, err
	}
	if r.Remaining() != 0 {
		return PDU{}, fmt.Errorf("lppa: trailing PDU data")
	}
	br := aper.NewReader(body)
	ext, err = aper.GetConstrained(br, 0, 1)
	if err != nil || ext != 0 {
		return PDU{}, fmt.Errorf("lppa: message extensions unsupported")
	}
	n, err := aper.GetConstrained(br, 0, maxProtocolIEs)
	if err != nil {
		return PDU{}, err
	}
	p := PDU{Category: Category(cat), ProcedureCode: uint8(proc), Criticality: crit, TransactionID: uint16(txn), IEs: make([]IE, 0, n)}
	for i := int64(0); i < n; i++ {
		id, err := aper.GetConstrained(br, 0, maxProtocolIEs)
		if err != nil {
			return PDU{}, err
		}
		c, err := aper.GetCriticality(br)
		if err != nil {
			return PDU{}, err
		}
		v, err := aper.GetOpenType(br)
		if err != nil {
			return PDU{}, err
		}
		p.IEs = append(p.IEs, IE{uint16(id), c, v})
	}
	if br.Remaining() > 7 || !br.RemainingZero() {
		return PDU{}, fmt.Errorf("lppa: trailing message data")
	}
	return p, nil
}
