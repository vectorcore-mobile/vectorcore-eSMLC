package result

import (
	"testing"

	"github.com/vectorcore/esmlc/internal/uper"
)

func FuzzPhysicalCellIDRoundTrip(f *testing.F) {
	for _, v := range []uint16{0, 1, 503, 504, 65535} {
		f.Add(v)
	}
	f.Fuzz(func(t *testing.T, value uint16) {
		v, err := NewPhysicalCellID(value)
		if value > 503 {
			if err == nil {
				t.Fatal("accepted invalid PCI")
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		w := uper.NewWriter()
		if err := v.EncodeUPER(w); err != nil {
			t.Fatal(err)
		}
		e := w.Encoded()
		if e.BitLength != 9 {
			t.Fatal(e.BitLength)
		}
		r, _ := uper.NewReader(e.Bytes, e.BitLength)
		got, err := DecodePhysicalCellID(r)
		if err != nil || got != v {
			t.Fatalf("%v %v", got, err)
		}
	})
}

func FuzzEUTRAARFCNRoundTrip(f *testing.F) {
	for _, v := range []uint32{0, 100, 65535, 65536, 262143} {
		f.Add(v)
	}
	f.Fuzz(func(t *testing.T, value uint32) {
		v, err := NewEUTRAARFCN(value)
		if value > 65535 {
			if err == nil {
				t.Fatal("accepted invalid ARFCN")
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		w := uper.NewWriter()
		if err := v.EncodeUPER(w); err != nil {
			t.Fatal(err)
		}
		e := w.Encoded()
		if e.BitLength != 16 {
			t.Fatal(e.BitLength)
		}
		r, _ := uper.NewReader(e.Bytes, e.BitLength)
		got, err := DecodeEUTRAARFCN(r)
		if err != nil || got != v {
			t.Fatalf("%v %v", got, err)
		}
	})
}

func FuzzRootScalarPacking(f *testing.F) {
	f.Add(uint16(1), uint16(100))
	f.Add(uint16(503), uint16(65535))
	f.Fuzz(func(t *testing.T, pciRaw, arfRaw uint16) {
		pci, err := NewPhysicalCellID(pciRaw)
		if err != nil {
			return
		}
		arf, _ := NewEUTRAARFCN(uint32(arfRaw))
		w := uper.NewWriter()
		_ = pci.EncodeUPER(w)
		_ = arf.EncodeUPER(w)
		e := w.Encoded()
		if e.BitLength != 25 {
			t.Fatal(e.BitLength)
		}
		r, _ := uper.NewReader(e.Bytes, e.BitLength)
		gotPCI, err := DecodePhysicalCellID(r)
		if err != nil {
			t.Fatal(err)
		}
		gotARF, err := DecodeEUTRAARFCN(r)
		if err != nil || gotPCI != pci || gotARF != arf {
			t.Fatalf("%v %v", gotPCI, gotARF)
		}
	})
}

func FuzzRootScalarDecode(f *testing.F) {
	f.Add([]byte{0x00, 0x80}, uint8(9))
	f.Add([]byte{0xff, 0xff}, uint8(16))
	f.Fuzz(func(t *testing.T, data []byte, bits uint8) {
		if len(data) > 8 {
			return
		}
		n := int(bits)
		if n > len(data)*8 {
			n = len(data) * 8
		}
		r, err := uper.NewReader(data, n)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = DecodePhysicalCellID(r)
		r, _ = uper.NewReader(data, n)
		_, _ = DecodeEUTRAARFCN(r)
	})
}
