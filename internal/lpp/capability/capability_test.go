package capability

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/vectorcore/esmlc/internal/uper"
)

func bits(t *testing.T, b byte, n int) uper.BitString {
	t.Helper()
	v, e := uper.NewBitString([]byte{b}, n)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func TestCapabilityFixtures(t *testing.T) {
	root, e := filepath.Abs("../../../tools/specs/lpp/fixtures/r16.4.0/capabilities")
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(filepath.Join(root, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	var m struct {
		Fixtures []struct {
			Name   string `json:"name"`
			Binary string `json:"binary_file"`
			Bits   int    `json:"bit_length"`
			Hex    string `json:"hex"`
		} `json:"fixtures"`
	}
	if e = json.Unmarshal(raw, &m); e != nil {
		t.Fatal(e)
	}
	if len(m.Fixtures) != 5 {
		t.Fatal(len(m.Fixtures))
	}
	for _, f := range m.Fixtures {
		t.Run(f.Name, func(t *testing.T) {
			data, e := os.ReadFile(filepath.Join(root, f.Binary))
			if e != nil {
				t.Fatal(e)
			}
			r, e := uper.NewReader(data, f.Bits)
			if e != nil {
				t.Fatal(e)
			}
			w := uper.NewWriter()
			switch f.Name {
			case "request-empty", "request-ecid-selector":
				v, e := DecodeRequestCapabilities(r)
				if e != nil {
					t.Fatal(e)
				}
				if f.Name == "request-empty" && v.ECID != nil {
					t.Fatal("unexpected ECID")
				}
				if f.Name == "request-ecid-selector" && v.ECID == nil {
					t.Fatal("missing ECID")
				}
				e = EncodeRequestCapabilities(w, v)
				if e != nil {
					t.Fatal(e)
				}
			case "provide-empty", "provide-ecid-rsrp", "provide-ecid-rsrp-rsrq-uerxtx":
				v, e := DecodeProvideCapabilities(r)
				if e != nil {
					t.Fatal(e)
				}
				if f.Name == "provide-empty" && v.ECID != nil {
					t.Fatal("unexpected ECID")
				}
				if f.Name == "provide-ecid-rsrp" && (!v.ECID.SupportsRSRP() || v.ECID.MeasurementSupport.BitLen() != 1) {
					t.Fatal("bad rsrp")
				}
				if f.Name == "provide-ecid-rsrp-rsrq-uerxtx" && (!v.ECID.SupportsRSRP() || !v.ECID.SupportsRSRQ() || !v.ECID.SupportsUERxTxTimeDifference() || v.ECID.MeasurementSupport.BitLen() != 3) {
					t.Fatal("bad three bit support")
				}
				e = EncodeProvideCapabilities(w, v)
				if e != nil {
					t.Fatal(e)
				}
			}
			if e = r.ValidateFinalPadding(); e != nil {
				t.Fatal(e)
			}
			out := w.Encoded()
			if string(out.Bytes) != string(data) || out.BitLength != f.Bits {
				t.Fatalf("got %x/%d want %x/%d", out.Bytes, out.BitLength, data, f.Bits)
			}
		})
	}
}
func TestUnsupportedAndMalformed(t *testing.T) {
	for _, tc := range []struct {
		bits []bool
		want error
	}{
		{[]bool{true, false, false, false, false}, ErrUnsupportedCommon},
		{[]bool{false, false, false, false, true}, ErrUnsupportedEPDU},
	} {
		w := uper.NewWriter()
		_ = w.WriteRootChoiceIndex(0, 2)
		_ = w.WriteRootChoiceIndex(0, 4)
		_ = w.WriteExtensionPresent(false)
		_ = w.WriteOptionalBitmap(tc.bits)
		r, _ := uper.NewReader(w.Encoded().Bytes, w.Encoded().BitLength)
		if _, e := DecodeRequestCapabilities(r); !errors.Is(e, tc.want) {
			t.Fatalf("got %v want %v", e, tc.want)
		}
	}
	// ECID provide has an extension bit followed by a mandatory SIZE(1..8) bit string.
	w := uper.NewWriter()
	_ = w.WriteRootChoiceIndex(0, 2)
	_ = w.WriteRootChoiceIndex(0, 4)
	_ = w.WriteExtensionPresent(false)
	_ = w.WriteOptionalBitmap([]bool{false, false, false, true, false})
	_ = w.WriteExtensionPresent(true)
	r, _ := uper.NewReader(w.Encoded().Bytes, w.Encoded().BitLength)
	if _, e := DecodeProvideCapabilities(r); !errors.Is(e, ErrUnsupportedExtension) {
		t.Fatal(e)
	}
	bad := ProvideCapabilitiesR9IEs{ECID: &ECIDProvideCapabilities{MeasurementSupport: uper.BitString{}}}
	if e := bad.Validate(); !errors.Is(e, ErrInvalidECID) {
		t.Fatal(e)
	}
	good := ProvideCapabilitiesR9IEs{ECID: &ECIDProvideCapabilities{MeasurementSupport: bits(t, 0x80, 1)}}
	w = uper.NewWriter()
	if e := EncodeProvideCapabilities(w, good); e != nil {
		t.Fatal(e)
	}
}
