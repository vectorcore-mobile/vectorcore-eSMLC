package lppa

import (
	"fmt"

	"github.com/vectorcore/esmlc/internal/aper"
)

// BuildInitiationRequest constructs an E-CIDMeasurementInitiationRequest PDU.
// MeasurementPeriodicity must be supplied iff report is ReportPeriodic (TS
// 36.455: "The IE shall be present if the Report Characteristics IE is set
// to 'periodic'"). quantities is always encoded as at least one item; this
// package's callers only ever request QuantityCellID.
func BuildInitiationRequest(transactionID uint16, measurementID uint8, report ReportCharacteristics, periodicity *MeasurementPeriodicity, quantities []MeasurementQuantityValue) (PDU, error) {
	if (report == ReportPeriodic) != (periodicity != nil) {
		return PDU{}, fmt.Errorf("lppa: measurement periodicity must be present iff report characteristics is periodic")
	}
	midValue, err := EncodeMeasurementID(measurementID)
	if err != nil {
		return PDU{}, err
	}
	reportValue, err := EncodeReportCharacteristics(report)
	if err != nil {
		return PDU{}, err
	}
	quantitiesValue, err := EncodeMeasurementQuantities(quantities)
	if err != nil {
		return PDU{}, err
	}
	ies := []IE{
		{IEESMLCMeasurementID, aper.Reject, midValue},
		{IEReportCharacteristics, aper.Reject, reportValue},
	}
	if periodicity != nil {
		periodicityValue, err := EncodeMeasurementPeriodicity(*periodicity)
		if err != nil {
			return PDU{}, err
		}
		ies = append(ies, IE{IEMeasurementPeriodicity, aper.Reject, periodicityValue})
	}
	ies = append(ies, IE{IEMeasurementQuantities, aper.Reject, quantitiesValue})
	return PDU{Category: Initiating, ProcedureCode: ProcedureECIDMeasurementInitiation, Criticality: aper.Reject, TransactionID: transactionID, IEs: ies}, nil
}

// InitiationResponse is the decoded E-CIDMeasurementInitiationResponse.
// Result, Diagnostics, and CellPortionID are nil when their optional IE is
// absent. InterRATMeasurementResult and WLANMeasurementResult are out of
// bounded scope and rejected fail-closed if the eNB sends them.
type InitiationResponse struct {
	ESMLCMeasurementID uint8
	ENBMeasurementID   uint8
	Result             *ECIDMeasurementResult
	Diagnostics        *CriticalityDiagnostics
	CellPortionID      *uint16
}

func DecodeInitiationResponse(p PDU) (InitiationResponse, error) {
	if p.Category != Successful || p.ProcedureCode != ProcedureECIDMeasurementInitiation || p.Criticality != aper.Reject {
		return InitiationResponse{}, fmt.Errorf("lppa: unsupported e-cid measurement initiation response")
	}
	seen := map[uint16]bool{}
	haveESMLC, haveENB := false, false
	var out InitiationResponse
	for _, ie := range p.IEs {
		if seen[ie.ID] {
			return InitiationResponse{}, fmt.Errorf("lppa: duplicate IE %d", ie.ID)
		}
		seen[ie.ID] = true
		switch ie.ID {
		case IEESMLCMeasurementID:
			v, err := DecodeMeasurementID(ie.Value)
			if err != nil {
				return InitiationResponse{}, err
			}
			out.ESMLCMeasurementID = v
			haveESMLC = true
		case IEENBMeasurementID:
			v, err := DecodeMeasurementID(ie.Value)
			if err != nil {
				return InitiationResponse{}, err
			}
			out.ENBMeasurementID = v
			haveENB = true
		case IECIDMeasurementResult:
			v, err := DecodeECIDMeasurementResult(ie.Value)
			if err != nil {
				return InitiationResponse{}, err
			}
			out.Result = &v
		case IECriticalityDiagnostics:
			v, err := DecodeCriticalityDiagnostics(ie.Value)
			if err != nil {
				return InitiationResponse{}, err
			}
			out.Diagnostics = &v
		case IECellPortionID:
			v, err := DecodeCellPortionID(ie.Value)
			if err != nil {
				return InitiationResponse{}, err
			}
			out.CellPortionID = &v
		default:
			if ie.Criticality == aper.Reject {
				return InitiationResponse{}, fmt.Errorf("lppa: unknown reject IE %d", ie.ID)
			}
		}
	}
	if !haveESMLC || !haveENB {
		return InitiationResponse{}, fmt.Errorf("lppa: missing mandatory IE")
	}
	return out, nil
}

// InitiationFailure is the decoded E-CIDMeasurementInitiationFailure.
type InitiationFailure struct {
	ESMLCMeasurementID uint8
	Cause              Cause
	Diagnostics        *CriticalityDiagnostics
}

func DecodeInitiationFailure(p PDU) (InitiationFailure, error) {
	if p.Category != Unsuccessful || p.ProcedureCode != ProcedureECIDMeasurementInitiation || p.Criticality != aper.Reject {
		return InitiationFailure{}, fmt.Errorf("lppa: unsupported e-cid measurement initiation failure")
	}
	seen := map[uint16]bool{}
	haveESMLC, haveCause := false, false
	var out InitiationFailure
	for _, ie := range p.IEs {
		if seen[ie.ID] {
			return InitiationFailure{}, fmt.Errorf("lppa: duplicate IE %d", ie.ID)
		}
		seen[ie.ID] = true
		switch ie.ID {
		case IEESMLCMeasurementID:
			v, err := DecodeMeasurementID(ie.Value)
			if err != nil {
				return InitiationFailure{}, err
			}
			out.ESMLCMeasurementID = v
			haveESMLC = true
		case IECause:
			v, err := DecodeCause(ie.Value)
			if err != nil {
				return InitiationFailure{}, err
			}
			out.Cause = v
			haveCause = true
		case IECriticalityDiagnostics:
			v, err := DecodeCriticalityDiagnostics(ie.Value)
			if err != nil {
				return InitiationFailure{}, err
			}
			out.Diagnostics = &v
		default:
			if ie.Criticality == aper.Reject {
				return InitiationFailure{}, fmt.Errorf("lppa: unknown reject IE %d", ie.ID)
			}
		}
	}
	if !haveESMLC || !haveCause {
		return InitiationFailure{}, fmt.Errorf("lppa: missing mandatory IE")
	}
	return out, nil
}

// FailureIndication is the decoded E-CIDMeasurementFailureIndication, an
// unsolicited eNB->E-SMLC message for an already-initiated measurement.
type FailureIndication struct {
	ESMLCMeasurementID uint8
	ENBMeasurementID   uint8
	Cause              Cause
}

func DecodeFailureIndication(p PDU) (FailureIndication, error) {
	if p.Category != Initiating || p.ProcedureCode != ProcedureECIDMeasurementFailureIndication || p.Criticality != aper.Ignore {
		return FailureIndication{}, fmt.Errorf("lppa: unsupported e-cid measurement failure indication")
	}
	seen := map[uint16]bool{}
	haveESMLC, haveENB, haveCause := false, false, false
	var out FailureIndication
	for _, ie := range p.IEs {
		if seen[ie.ID] {
			return FailureIndication{}, fmt.Errorf("lppa: duplicate IE %d", ie.ID)
		}
		seen[ie.ID] = true
		switch ie.ID {
		case IEESMLCMeasurementID:
			v, err := DecodeMeasurementID(ie.Value)
			if err != nil {
				return FailureIndication{}, err
			}
			out.ESMLCMeasurementID = v
			haveESMLC = true
		case IEENBMeasurementID:
			v, err := DecodeMeasurementID(ie.Value)
			if err != nil {
				return FailureIndication{}, err
			}
			out.ENBMeasurementID = v
			haveENB = true
		case IECause:
			v, err := DecodeCause(ie.Value)
			if err != nil {
				return FailureIndication{}, err
			}
			out.Cause = v
			haveCause = true
		default:
			if ie.Criticality == aper.Reject {
				return FailureIndication{}, fmt.Errorf("lppa: unknown reject IE %d", ie.ID)
			}
		}
	}
	if !haveESMLC || !haveENB || !haveCause {
		return FailureIndication{}, fmt.Errorf("lppa: missing mandatory IE")
	}
	return out, nil
}

// Report is the decoded E-CIDMeasurementReport, an unsolicited eNB->E-SMLC
// measurement report (used for periodic and on-demand reporting alike).
type Report struct {
	ESMLCMeasurementID uint8
	ENBMeasurementID   uint8
	Result             ECIDMeasurementResult
	CellPortionID      *uint16
}

func DecodeReport(p PDU) (Report, error) {
	if p.Category != Initiating || p.ProcedureCode != ProcedureECIDMeasurementReport || p.Criticality != aper.Ignore {
		return Report{}, fmt.Errorf("lppa: unsupported e-cid measurement report")
	}
	seen := map[uint16]bool{}
	haveESMLC, haveENB, haveResult := false, false, false
	var out Report
	for _, ie := range p.IEs {
		if seen[ie.ID] {
			return Report{}, fmt.Errorf("lppa: duplicate IE %d", ie.ID)
		}
		seen[ie.ID] = true
		switch ie.ID {
		case IEESMLCMeasurementID:
			v, err := DecodeMeasurementID(ie.Value)
			if err != nil {
				return Report{}, err
			}
			out.ESMLCMeasurementID = v
			haveESMLC = true
		case IEENBMeasurementID:
			v, err := DecodeMeasurementID(ie.Value)
			if err != nil {
				return Report{}, err
			}
			out.ENBMeasurementID = v
			haveENB = true
		case IECIDMeasurementResult:
			v, err := DecodeECIDMeasurementResult(ie.Value)
			if err != nil {
				return Report{}, err
			}
			out.Result = v
			haveResult = true
		case IECellPortionID:
			v, err := DecodeCellPortionID(ie.Value)
			if err != nil {
				return Report{}, err
			}
			out.CellPortionID = &v
		default:
			if ie.Criticality == aper.Reject {
				return Report{}, fmt.Errorf("lppa: unknown reject IE %d", ie.ID)
			}
		}
	}
	if !haveESMLC || !haveENB || !haveResult {
		return Report{}, fmt.Errorf("lppa: missing mandatory IE")
	}
	return out, nil
}

// BuildTerminationCommand constructs an E-CIDMeasurementTerminationCommand
// PDU, ending an already-initiated E-CID measurement.
func BuildTerminationCommand(transactionID uint16, esmlcMeasurementID, enbMeasurementID uint8) (PDU, error) {
	esmlcValue, err := EncodeMeasurementID(esmlcMeasurementID)
	if err != nil {
		return PDU{}, err
	}
	enbValue, err := EncodeMeasurementID(enbMeasurementID)
	if err != nil {
		return PDU{}, err
	}
	ies := []IE{
		{IEESMLCMeasurementID, aper.Reject, esmlcValue},
		{IEENBMeasurementID, aper.Reject, enbValue},
	}
	return PDU{Category: Initiating, ProcedureCode: ProcedureECIDMeasurementTermination, Criticality: aper.Reject, TransactionID: transactionID, IEs: ies}, nil
}
