package lcsap

import "testing"

func FuzzDecodeCompletePDU(f *testing.F) {
	f.Add([]byte{0x20, 0, 0, 0x10, 0, 0, 2, 0, 2, 0, 4, 0, 0, 0, 1, 0, 0, 0, 1, 0})
	f.Add([]byte{0x40, 0, 0, 0x10, 0, 0, 2, 0, 2, 0, 4, 0, 0, 0, 1, 0, 0x0b, 0x40, 1, 0xd8})
	f.Fuzz(func(t *testing.T, wire []byte) {
		p, err := Decode(wire)
		if err != nil {
			return
		}
		if p.Procedure == ProcedureLocationRequest && p.Category == Initiating {
			_, _ = DecodeLocationRequest(p)
		}
		if p.Procedure == ProcedureConnectionOrientedInformation {
			_, _ = DecodeConnectionOriented(p, 1<<20)
		}
	})
}
