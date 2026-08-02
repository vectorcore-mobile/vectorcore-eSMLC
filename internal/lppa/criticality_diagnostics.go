package lppa

import (
	"fmt"

	"github.com/vectorcore/esmlc/internal/aper"
)

// TriggeringMessage ::= ENUMERATED { initiating-message, successful-outcome,
// unsuccessful-outcome }. Unlike most LPPa enumerations this one is declared
// without an extension marker in TS 36.455 LPPA-CommonDataTypes.
type TriggeringMessage uint8

const (
	TriggerInitiatingMessage TriggeringMessage = iota
	TriggerSuccessfulOutcome
	TriggerUnsuccessfulOutcome
)

// TypeOfError ::= ENUMERATED { not-understood, missing, ... }.
type TypeOfError uint8

const (
	ErrorNotUnderstood TypeOfError = iota
	ErrorMissing
)

// CriticalityDiagnosticsIEItem is one entry of CriticalityDiagnostics-IE-List.
type CriticalityDiagnosticsIEItem struct {
	Criticality aper.Criticality
	IEID        uint16
	TypeOfError TypeOfError
}

// CriticalityDiagnostics is the bounded root subset of the diagnostic
// SEQUENCE an eNB may return alongside a reject/ignore outcome; every field
// is independently OPTIONAL per spec, so each is a pointer/nil slice rather
// than a mandatory value. iE-Extensions (on this SEQUENCE or on any IE-List
// item) is out of bounded scope and rejected fail-closed if present.
type CriticalityDiagnostics struct {
	ProcedureCode        *uint8
	TriggeringMessage    *TriggeringMessage
	ProcedureCriticality *aper.Criticality
	TransactionID        *uint16
	IEs                  []CriticalityDiagnosticsIEItem
}

func DecodeCriticalityDiagnostics(b []byte) (CriticalityDiagnostics, error) {
	r := aper.NewReader(b)
	ext, err := aper.GetConstrained(r, 0, 1)
	if err != nil {
		return CriticalityDiagnostics{}, err
	}
	if ext != 0 {
		return CriticalityDiagnostics{}, fmt.Errorf("lppa: criticality diagnostics extension unsupported")
	}
	var present [6]int64
	for i := range present {
		v, err := aper.GetConstrained(r, 0, 1)
		if err != nil {
			return CriticalityDiagnostics{}, err
		}
		present[i] = v
	}
	var out CriticalityDiagnostics
	if present[0] != 0 {
		v, err := aper.GetConstrained(r, 0, 255)
		if err != nil {
			return CriticalityDiagnostics{}, err
		}
		x := uint8(v)
		out.ProcedureCode = &x
	}
	if present[1] != 0 {
		v, err := aper.GetConstrained(r, 0, 2)
		if err != nil {
			return CriticalityDiagnostics{}, err
		}
		x := TriggeringMessage(v)
		out.TriggeringMessage = &x
	}
	if present[2] != 0 {
		c, err := aper.GetCriticality(r)
		if err != nil {
			return CriticalityDiagnostics{}, err
		}
		out.ProcedureCriticality = &c
	}
	if present[3] != 0 {
		v, err := aper.GetConstrained(r, 0, 32767)
		if err != nil {
			return CriticalityDiagnostics{}, err
		}
		x := uint16(v)
		out.TransactionID = &x
	}
	if present[4] != 0 {
		n, err := aper.GetConstrained(r, 1, 256)
		if err != nil {
			return CriticalityDiagnostics{}, err
		}
		items := make([]CriticalityDiagnosticsIEItem, 0, n)
		for i := int64(0); i < n; i++ {
			itemExt, err := aper.GetConstrained(r, 0, 1)
			if err != nil {
				return CriticalityDiagnostics{}, err
			}
			if itemExt != 0 {
				return CriticalityDiagnostics{}, fmt.Errorf("lppa: criticality diagnostics ie-list item extension unsupported")
			}
			hasExt, err := aper.GetConstrained(r, 0, 1)
			if err != nil {
				return CriticalityDiagnostics{}, err
			}
			crit, err := aper.GetCriticality(r)
			if err != nil {
				return CriticalityDiagnostics{}, err
			}
			id, err := aper.GetConstrained(r, 0, maxProtocolIEs)
			if err != nil {
				return CriticalityDiagnostics{}, err
			}
			toeExt, err := aper.GetConstrained(r, 0, 1)
			if err != nil {
				return CriticalityDiagnostics{}, err
			}
			if toeExt != 0 {
				return CriticalityDiagnostics{}, fmt.Errorf("lppa: type-of-error extension unsupported")
			}
			toe, err := aper.GetConstrained(r, 0, 1)
			if err != nil {
				return CriticalityDiagnostics{}, err
			}
			if hasExt != 0 {
				return CriticalityDiagnostics{}, fmt.Errorf("lppa: criticality diagnostics ie-list item extensions unsupported")
			}
			items = append(items, CriticalityDiagnosticsIEItem{crit, uint16(id), TypeOfError(toe)})
		}
		out.IEs = items
	}
	if present[5] != 0 {
		return CriticalityDiagnostics{}, fmt.Errorf("lppa: criticality diagnostics extensions unsupported")
	}
	if r.Remaining() > 7 || !r.RemainingZero() {
		return CriticalityDiagnostics{}, fmt.Errorf("lppa: trailing criticality diagnostics data")
	}
	return out, nil
}

// EncodeCriticalityDiagnostics exists for deterministic round-trip tests;
// production code only decodes this eNB-originated diagnostic IE.
func EncodeCriticalityDiagnostics(v CriticalityDiagnostics) ([]byte, error) {
	w := aper.NewWriter()
	if err := extBit(w, 0); err != nil {
		return nil, err
	}
	present := []bool{v.ProcedureCode != nil, v.TriggeringMessage != nil, v.ProcedureCriticality != nil, v.TransactionID != nil, len(v.IEs) > 0, false}
	for _, p := range present {
		if err := aper.PutConstrained(w, boolInt(p), 0, 1); err != nil {
			return nil, err
		}
	}
	if v.ProcedureCode != nil {
		if err := aper.PutConstrained(w, int64(*v.ProcedureCode), 0, 255); err != nil {
			return nil, err
		}
	}
	if v.TriggeringMessage != nil {
		if *v.TriggeringMessage > TriggerUnsuccessfulOutcome {
			return nil, fmt.Errorf("lppa: unsupported triggering message")
		}
		if err := aper.PutConstrained(w, int64(*v.TriggeringMessage), 0, 2); err != nil {
			return nil, err
		}
	}
	if v.ProcedureCriticality != nil {
		if err := aper.PutCriticality(w, *v.ProcedureCriticality); err != nil {
			return nil, err
		}
	}
	if v.TransactionID != nil {
		if err := aper.PutConstrained(w, int64(*v.TransactionID), 0, 32767); err != nil {
			return nil, err
		}
	}
	if len(v.IEs) > 0 {
		if len(v.IEs) > 256 {
			return nil, fmt.Errorf("lppa: too many criticality diagnostics ie-list items")
		}
		if err := aper.PutConstrained(w, int64(len(v.IEs)), 1, 256); err != nil {
			return nil, err
		}
		for _, item := range v.IEs {
			if item.TypeOfError > ErrorMissing {
				return nil, fmt.Errorf("lppa: unsupported type of error")
			}
			if err := extBit(w, 0); err != nil {
				return nil, err
			}
			if err := extBit(w, 0); err != nil { // iE-Extensions absent
				return nil, err
			}
			if err := aper.PutCriticality(w, item.Criticality); err != nil {
				return nil, err
			}
			if err := aper.PutConstrained(w, int64(item.IEID), 0, maxProtocolIEs); err != nil {
				return nil, err
			}
			if err := extBit(w, 0); err != nil {
				return nil, err
			}
			if err := aper.PutConstrained(w, int64(item.TypeOfError), 0, 1); err != nil {
				return nil, err
			}
		}
	}
	return w.Bytes(), nil
}
