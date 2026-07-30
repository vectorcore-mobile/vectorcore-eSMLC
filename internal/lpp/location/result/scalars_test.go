package result

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/vectorcore/esmlc/internal/uper"
)

type scalarFixtureManifest struct {
	Fixtures []struct {
		Type      string          `json:"type"`
		Value     json.RawMessage `json:"value"`
		Hex       string          `json:"hex"`
		BitLength int             `json:"bit_length"`
	} `json:"fixtures"`
}

func loadScalarFixtures(t *testing.T) scalarFixtureManifest {
	t.Helper()
	path, err := filepath.Abs("../../../../tools/specs/lpp/analysis/r16.4.0/ecid-measured-result-scalars/fixtures.json")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var manifest scalarFixtureManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	return manifest
}

func TestPhysicalCellIDFixtures(t *testing.T) {
	manifest := loadScalarFixtures(t)
	seen := 0
	for _, fixture := range manifest.Fixtures {
		if fixture.Type != "PhysicalCellID" {
			continue
		}
		seen++
		var value uint16
		if err := json.Unmarshal(fixture.Value, &value); err != nil {
			t.Fatal(err)
		}
		want, err := hex.DecodeString(fixture.Hex)
		if err != nil {
			t.Fatal(err)
		}
		v, err := NewPhysicalCellID(value)
		if err != nil {
			t.Fatal(err)
		}
		w := uper.NewWriter()
		if err := v.EncodeUPER(w); err != nil {
			t.Fatal(err)
		}
		encoded := w.Encoded()
		if string(encoded.Bytes) != string(want) || encoded.BitLength != fixture.BitLength {
			t.Fatalf("PCI %d: got %x/%d, want %x/%d", value, encoded.Bytes, encoded.BitLength, want, fixture.BitLength)
		}
		r, err := uper.NewReader(want, fixture.BitLength)
		if err != nil {
			t.Fatal(err)
		}
		got, err := DecodePhysicalCellID(r)
		if err != nil || got != v || r.ValidateFinalPadding() != nil {
			t.Fatalf("PCI %d round trip: got %d, err %v", value, got, err)
		}
	}
	if seen != 3 {
		t.Fatalf("found %d PCI fixtures, want 3", seen)
	}
}

func TestEUTRAARFCNFixtures(t *testing.T) {
	manifest := loadScalarFixtures(t)
	seen := 0
	for _, fixture := range manifest.Fixtures {
		if fixture.Type != "EUTRACarrierFrequencyRoot" {
			continue
		}
		seen++
		var value uint32
		if err := json.Unmarshal(fixture.Value, &value); err != nil {
			t.Fatal(err)
		}
		want, err := hex.DecodeString(fixture.Hex)
		if err != nil {
			t.Fatal(err)
		}
		v, err := NewEUTRAARFCN(value)
		if err != nil {
			t.Fatal(err)
		}
		w := uper.NewWriter()
		if err := v.EncodeUPER(w); err != nil {
			t.Fatal(err)
		}
		encoded := w.Encoded()
		if string(encoded.Bytes) != string(want) || encoded.BitLength != fixture.BitLength {
			t.Fatalf("ARFCN %d: got %x/%d, want %x/%d", value, encoded.Bytes, encoded.BitLength, want, fixture.BitLength)
		}
		r, err := uper.NewReader(want, fixture.BitLength)
		if err != nil {
			t.Fatal(err)
		}
		got, err := DecodeEUTRAARFCN(r)
		if err != nil || got != v || r.ValidateFinalPadding() != nil {
			t.Fatalf("ARFCN %d round trip: got %d, err %v", value, got, err)
		}
	}
	if seen != 3 {
		t.Fatalf("found %d root ARFCN fixtures, want 3", seen)
	}
}

func TestRootScalarPackingFixtures(t *testing.T) {
	for _, tc := range []struct {
		pci PhysicalCellID
		arf EUTRAARFCN
		hex string
	}{
		{1, 100, "00803200"},
		{503, 65535, "fbffff80"},
	} {
		w := uper.NewWriter()
		if err := tc.pci.EncodeUPER(w); err != nil {
			t.Fatal(err)
		}
		if err := tc.arf.EncodeUPER(w); err != nil {
			t.Fatal(err)
		}
		encoded := w.Encoded()
		want, _ := hex.DecodeString(tc.hex)
		if string(encoded.Bytes) != string(want) || encoded.BitLength != 25 {
			t.Fatalf("got %x/%d, want %x/25", encoded.Bytes, encoded.BitLength, want)
		}
		r, _ := uper.NewReader(want, 25)
		pci, err := DecodePhysicalCellID(r)
		if err != nil {
			t.Fatal(err)
		}
		arf, err := DecodeEUTRAARFCN(r)
		if err != nil || pci != tc.pci || arf != tc.arf || r.ValidateFinalPadding() != nil {
			t.Fatalf("got {%d,%d}, err %v", pci, arf, err)
		}
	}
}

func TestScalarBounds(t *testing.T) {
	for _, value := range []uint16{0, 1, 502, 503} {
		if _, err := NewPhysicalCellID(value); err != nil {
			t.Fatalf("PCI %d: %v", value, err)
		}
	}
	for _, value := range []uint16{504, 65535} {
		if _, err := NewPhysicalCellID(value); !errors.Is(err, ErrPhysicalCellIDOutOfRange) {
			t.Fatalf("PCI %d: %v", value, err)
		}
	}
	for _, value := range []uint32{0, 1, 65534, 65535} {
		if _, err := NewEUTRAARFCN(value); err != nil {
			t.Fatalf("ARFCN %d: %v", value, err)
		}
	}
	for _, value := range []uint32{65536, 262143, ^uint32(0)} {
		if _, err := NewEUTRAARFCN(value); !errors.Is(err, ErrEUTRAARFCNOutOfRange) {
			t.Fatalf("ARFCN %d: %v", value, err)
		}
	}
	if err := PhysicalCellID(504).EncodeUPER(uper.NewWriter()); !errors.Is(err, ErrPhysicalCellIDOutOfRange) {
		t.Fatal(err)
	}
}

func TestScalarTruncation(t *testing.T) {
	for n := 0; n < 9; n++ {
		r, _ := uper.NewReader([]byte{0xff, 0xff}, n)
		if _, err := DecodePhysicalCellID(r); !errors.Is(err, ErrPhysicalCellIDDecode) || !errors.Is(err, uper.ErrUnexpectedEOF) {
			t.Fatalf("PCI %d bits: %v", n, err)
		}
	}
	for n := 0; n < 16; n++ {
		r, _ := uper.NewReader([]byte{0xff, 0xff}, n)
		if _, err := DecodeEUTRAARFCN(r); !errors.Is(err, ErrEUTRAARFCNDecode) || !errors.Is(err, uper.ErrUnexpectedEOF) {
			t.Fatalf("ARFCN %d bits: %v", n, err)
		}
	}
}

func TestRootScalarPackingTruncation(t *testing.T) {
	w := uper.NewWriter()
	if err := PhysicalCellID(503).EncodeUPER(w); err != nil {
		t.Fatal(err)
	}
	if err := EUTRAARFCN(65535).EncodeUPER(w); err != nil {
		t.Fatal(err)
	}
	encoded := w.Encoded()
	for n := 0; n < 25; n++ {
		r, err := uper.NewReader(encoded.Bytes, n)
		if err != nil {
			t.Fatal(err)
		}
		_, pciErr := DecodePhysicalCellID(r)
		if pciErr != nil {
			if !errors.Is(pciErr, ErrPhysicalCellIDDecode) || !errors.Is(pciErr, uper.ErrUnexpectedEOF) {
				t.Fatalf("%d bits PCI error: %v", n, pciErr)
			}
			continue
		}
		if _, err := DecodeEUTRAARFCN(r); !errors.Is(err, ErrEUTRAARFCNDecode) || !errors.Is(err, uper.ErrUnexpectedEOF) {
			t.Fatalf("%d bits ARFCN error: %v", n, err)
		}
	}
}

func TestScalarNonAlignedComposition(t *testing.T) {
	w := uper.NewWriter()
	_ = w.WriteBoolean(true)
	if err := PhysicalCellID(1).EncodeUPER(w); err != nil {
		t.Fatal(err)
	}
	if err := EUTRAARFCN(100).EncodeUPER(w); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteConstrainedWholeNumber(2, 0, 3); err != nil {
		t.Fatal(err)
	}
	encoded := w.Encoded()
	if encoded.BitLength != 28 {
		t.Fatalf("got %d bits, want 28", encoded.BitLength)
	}
	r, _ := uper.NewReader(encoded.Bytes, encoded.BitLength)
	b, _ := r.ReadBoolean()
	pci, err := DecodePhysicalCellID(r)
	if err != nil {
		t.Fatal(err)
	}
	arf, err := DecodeEUTRAARFCN(r)
	if err != nil {
		t.Fatal(err)
	}
	suffix, err := r.ReadConstrainedWholeNumber(0, 3)
	if err != nil || !b || pci != 1 || arf != 100 || suffix != 2 || r.ValidateFinalPadding() != nil {
		t.Fatalf("composition failure: %v", err)
	}
}

func TestScalarValueSemantics(t *testing.T) {
	pci, err := NewPhysicalCellID(1)
	if err != nil {
		t.Fatal(err)
	}
	arf, err := NewEUTRAARFCN(100)
	if err != nil {
		t.Fatal(err)
	}
	pciCopy, arfCopy := pci, arf
	if pciCopy != pci || arfCopy != arf || pci.Validate() != nil || arf.Validate() != nil {
		t.Fatal("scalar value semantics are not stable")
	}
	if (PhysicalCellID(0)).Validate() != nil || (EUTRAARFCN(0)).Validate() != nil {
		t.Fatal("zero values must be valid")
	}
}
