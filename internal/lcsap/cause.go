package lcsap

import (
	"fmt"

	"github.com/vectorcore/esmlc/internal/aper"
)

type LCSCauseBranch uint8

const (
	LCSCauseRadioNetwork LCSCauseBranch = iota
	LCSCauseTransport
	LCSCauseProtocol
	LCSCauseMisc
)

// LCSCause is the complete Release-16 root CHOICE. Value is the root
// ENUMERATED index of the selected branch; extensions are fail-closed.
type LCSCause struct {
	Branch LCSCauseBranch
	Value  uint8
}

const (
	RadioNetworkUnspecified            uint8 = 0
	TransportResourceUnavailable       uint8 = 0
	TransportUnspecified               uint8 = 1
	ProtocolTransferSyntaxError        uint8 = 0
	ProtocolAbstractSyntaxReject       uint8 = 1
	ProtocolAbstractSyntaxIgnoreNotify uint8 = 2
	ProtocolMessageIncompatible        uint8 = 3
	ProtocolSemanticError              uint8 = 4
	ProtocolUnspecified                uint8 = 5
	ProtocolAbstractSyntaxError        uint8 = 6
	MiscProcessingOverload             uint8 = 0
	MiscHardwareFailure                uint8 = 1
	MiscOMIntervention                 uint8 = 2
	MiscUnspecified                    uint8 = 3
)

func (v LCSCause) Validate() error {
	max := map[LCSCauseBranch]uint8{LCSCauseRadioNetwork: 0, LCSCauseTransport: 1, LCSCauseProtocol: 6, LCSCauseMisc: 3}
	m, ok := max[v.Branch]
	if !ok || v.Value > m {
		return fmt.Errorf("lcsap: invalid LCS cause")
	}
	return nil
}

func EncodeLCSCause(v LCSCause) ([]byte, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	w := aper.NewWriter()
	if err := aper.PutConstrained(w, int64(v.Branch), 0, 3); err != nil {
		return nil, err
	}
	// All four cause ENUMERATED definitions are extensible: root extension bit.
	if err := aper.PutConstrained(w, 0, 0, 1); err != nil {
		return nil, err
	}
	max := map[LCSCauseBranch]int64{LCSCauseRadioNetwork: 0, LCSCauseTransport: 1, LCSCauseProtocol: 6, LCSCauseMisc: 3}[v.Branch]
	if err := aper.PutConstrained(w, int64(v.Value), 0, max); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func DecodeLCSCause(b []byte) (LCSCause, error) {
	r := aper.NewReader(b)
	branch, err := aper.GetConstrained(r, 0, 3)
	if err != nil {
		return LCSCause{}, err
	}
	ext, err := aper.GetConstrained(r, 0, 1)
	if err != nil || ext != 0 {
		return LCSCause{}, fmt.Errorf("lcsap: LCS cause extensions unsupported")
	}
	v := LCSCause{Branch: LCSCauseBranch(branch)}
	max := map[LCSCauseBranch]int64{LCSCauseRadioNetwork: 0, LCSCauseTransport: 1, LCSCauseProtocol: 6, LCSCauseMisc: 3}[v.Branch]
	value, err := aper.GetConstrained(r, 0, max)
	if err != nil || r.Remaining() > 7 || !r.RemainingZero() {
		return LCSCause{}, fmt.Errorf("lcsap: invalid LCS cause")
	}
	v.Value = uint8(value)
	return v, v.Validate()
}
