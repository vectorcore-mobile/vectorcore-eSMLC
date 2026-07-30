package location

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/vectorcore/esmlc/internal/uper"
)

func TestRequestFixtures(t *testing.T) {
	root, err := filepath.Abs("../../../tools/specs/lpp/fixtures/r16.4.0/ecid-location")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Fixtures []struct {
			Name   string `json:"name"`
			Bits   int    `json:"bit_length"`
			Binary string `json:"binary_file"`
		}
	}
	if err = json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	seen := 0
	for _, fixture := range manifest.Fixtures {
		if fixture.Name != "request-ecid-rsrp" && fixture.Name != "request-ecid-all-root" {
			continue
		}
		seen++
		t.Run(fixture.Name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, "valid", fixture.Name+".uper"))
			if err != nil {
				t.Fatal(err)
			}
			r, err := uper.NewReader(data, fixture.Bits)
			if err != nil {
				t.Fatal(err)
			}
			got, err := DecodeRequestLocationInformation(r)
			if err != nil {
				t.Fatal(err)
			}
			if err := r.ValidateFinalPadding(); err != nil {
				t.Fatal(err)
			}
			if got.ECID == nil || !got.ECID.RequestsRSRP() {
				t.Fatal("missing RSRP request")
			}
			switch fixture.Name {
			case "request-ecid-rsrp":
				if got.ECID.RequestedMeasurements.BitLen() != 1 || got.ECID.RequestsRSRQ() || got.ECID.RequestsUERxTxTimeDifference() {
					t.Fatalf("unexpected one-bit semantics: %+v", got.ECID)
				}
			case "request-ecid-all-root":
				if got.ECID.RequestedMeasurements.BitLen() != 3 || !got.ECID.RequestsRSRQ() || !got.ECID.RequestsUERxTxTimeDifference() {
					t.Fatalf("unexpected three-bit semantics: %+v", got.ECID)
				}
			}
			w := uper.NewWriter()
			if err := EncodeRequestLocationInformation(w, got); err != nil {
				t.Fatal(err)
			}
			out := w.Encoded()
			if string(out.Bytes) != string(data) || out.BitLength != fixture.Bits {
				t.Fatalf("got %x/%d, want %x/%d", out.Bytes, out.BitLength, data, fixture.Bits)
			}
		})
	}
	if seen != 2 {
		t.Fatalf("found %d request fixtures, want 2", seen)
	}
}

func TestRejectUnsupportedAndInvalidRequest(t *testing.T) {
	for _, tc := range []struct {
		name string
		bits []bool
		want error
	}{
		{"common", []bool{true, false, false, false, false}, ErrUnsupportedCommon},
		{"agnss", []bool{false, true, false, false, false}, ErrUnsupportedAGNSS},
		{"otdoa", []bool{false, false, true, false, false}, ErrUnsupportedOTDOA},
		{"epdu", []bool{false, false, false, false, true}, ErrUnsupportedEPDU},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := uper.NewWriter()
			_ = w.WriteRootChoiceIndex(0, 2)
			_ = w.WriteRootChoiceIndex(0, 4)
			_ = w.WriteExtensionPresent(false)
			_ = w.WriteOptionalBitmap(tc.bits)
			e := w.Encoded()
			r, _ := uper.NewReader(e.Bytes, e.BitLength)
			if _, err := DecodeRequestLocationInformation(r); !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
	w := uper.NewWriter()
	_ = w.WriteRootChoiceIndex(0, 2)
	_ = w.WriteRootChoiceIndex(0, 4)
	_ = w.WriteExtensionPresent(true)
	e := w.Encoded()
	r, _ := uper.NewReader(e.Bytes, e.BitLength)
	if _, err := DecodeRequestLocationInformation(r); !errors.Is(err, ErrUnsupportedExtension) {
		t.Fatal(err)
	}
	w = uper.NewWriter()
	_ = w.WriteRootChoiceIndex(0, 2)
	_ = w.WriteRootChoiceIndex(0, 4)
	_ = w.WriteExtensionPresent(false)
	_ = w.WriteOptionalBitmap([]bool{false, false, false, true, false})
	_ = w.WriteExtensionPresent(true)
	e = w.Encoded()
	r, _ = uper.NewReader(e.Bytes, e.BitLength)
	if _, err := DecodeRequestLocationInformation(r); !errors.Is(err, ErrUnsupportedExtension) {
		t.Fatal(err)
	}
	w = uper.NewWriter()
	_ = w.WriteRootChoiceIndex(1, 2)
	e = w.Encoded()
	r, _ = uper.NewReader(e.Bytes, e.BitLength)
	if _, err := DecodeRequestLocationInformation(r); !errors.Is(err, ErrUnsupportedCriticalExtension) {
		t.Fatal(err)
	}
	tooLong, err := uper.NewBitString([]byte{0xff, 0x80}, 9)
	if err != nil {
		t.Fatal(err)
	}
	if err := (RequestLocationInformationR9IEs{ECID: &ECIDRequestLocationInformation{RequestedMeasurements: tooLong}}).Validate(); !errors.Is(err, ErrInvalidECIDRequest) {
		t.Fatal(err)
	}
	if err := (RequestLocationInformationR9IEs{ECID: &ECIDRequestLocationInformation{}}).Validate(); !errors.Is(err, ErrMissingRequestedMeasurements) {
		t.Fatal(err)
	}
}

func TestECIDRequestRoundTrip(t *testing.T) {
	for n := 1; n <= 8; n++ {
		t.Run(string(rune('0'+n)), func(t *testing.T) {
			v, err := uper.NewBitString([]byte{byte(0xff << (8 - n))}, n)
			if err != nil {
				t.Fatal(err)
			}
			in := RequestLocationInformationR9IEs{ECID: &ECIDRequestLocationInformation{RequestedMeasurements: v}}
			w := uper.NewWriter()
			if err := EncodeRequestLocationInformation(w, in); err != nil {
				t.Fatal(err)
			}
			e := w.Encoded()
			r, err := uper.NewReader(e.Bytes, e.BitLength)
			if err != nil {
				t.Fatal(err)
			}
			out, err := DecodeRequestLocationInformation(r)
			if err != nil || !out.ECID.RequestedMeasurements.Equal(v) {
				t.Fatalf("got %#v, %v", out, err)
			}
		})
	}
}
