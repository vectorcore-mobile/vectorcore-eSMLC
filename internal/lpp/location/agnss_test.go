package location

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/vectorcore/esmlc/internal/uper"
)

func agnssFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "tools", "specs", "lpp", "fixtures", "r16.4.0", "a-gnss", name+".uper"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestAGNSSRequestLocationInformationDecodesCommonAndAGNSS(t *testing.T) {
	data := agnssFixture(t, "request-location-information-r9-agnss-ue-based")
	r, err := uper.NewReader(data, len(data)*8)
	if err != nil {
		t.Fatal(err)
	}
	v, err := DecodeRequestLocationInformation(r)
	if err != nil {
		t.Fatal(err)
	}
	if v.Common == nil || v.Common.LocationInformationType != LocationEstimateRequired {
		t.Fatalf("expected common locationEstimateRequired, got %#v", v.Common)
	}
	if v.AGNSS == nil || v.AGNSS.AssistanceAvailability {
		t.Fatalf("expected A-GNSS instructions with assistanceAvailability=false, got %#v", v.AGNSS)
	}
	w := uper.NewWriter()
	if err := EncodeRequestLocationInformation(w, v); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(w.Encoded().Bytes, data) {
		t.Fatalf("got %x want %x", w.Encoded().Bytes, data)
	}
}

func TestAGNSSProvideLocationInformationEstimateFixture(t *testing.T) {
	data := agnssFixture(t, "provide-location-information-r9-agnss-estimate")
	r, err := uper.NewReader(data, len(data)*8)
	if err != nil {
		t.Fatal(err)
	}
	v, err := DecodeProvideLocationInformation(r)
	if err != nil {
		t.Fatal(err)
	}
	if v.Common == nil || v.Common.LocationEstimate == nil {
		t.Fatal("expected a common location estimate")
	}
	est := v.Common.LocationEstimate
	if est.Shape != ShapePointWithUncertaintyCircle || est.UncertaintyCircle != 30 {
		t.Fatalf("unexpected estimate shape: %#v", est)
	}
	if latErr := est.Point.Latitude - 38.0; latErr > 0.001 || latErr < -0.001 {
		t.Fatalf("latitude off: %v", est.Point.Latitude)
	}
	if v.AGNSS == nil || v.AGNSS.GNSSLocationInformation == nil {
		t.Fatal("expected A-GNSS supplementary metadata")
	}
	w := uper.NewWriter()
	if err := EncodeProvideLocationInformation(w, v); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(w.Encoded().Bytes, data) {
		t.Fatalf("got %x want %x", w.Encoded().Bytes, data)
	}
}

func TestAGNSSProvideLocationInformationErrorFixture(t *testing.T) {
	data := agnssFixture(t, "provide-location-information-r9-agnss-error")
	r, err := uper.NewReader(data, len(data)*8)
	if err != nil {
		t.Fatal(err)
	}
	v, err := DecodeProvideLocationInformation(r)
	if err != nil {
		t.Fatal(err)
	}
	if v.AGNSS == nil || v.AGNSS.Error == nil || v.AGNSS.Error.Source != AGNSSErrorTargetDevice || v.AGNSS.Error.TargetDeviceCause != AGNSSTargetDeviceCauseAssistanceDataMissing {
		t.Fatalf("expected target-device assistanceDataMissing error, got %#v", v.AGNSS)
	}
	w := uper.NewWriter()
	if err := EncodeProvideLocationInformation(w, v); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(w.Encoded().Bytes, data) {
		t.Fatalf("got %x want %x", w.Encoded().Bytes, data)
	}
}

func TestLocationCoordinatesRejectsUnsupportedShape(t *testing.T) {
	w := uper.NewWriter()
	if err := w.WriteExtensionPresent(false); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteRootChoiceIndex(0, 7); err != nil { // shape 0: ellipsoidPoint, unsupported
		t.Fatal(err)
	}
	if err := writeCoordinates(w, 38, -90); err != nil {
		t.Fatal(err)
	}
	e := w.Encoded()
	r, err := uper.NewReader(e.Bytes, e.BitLength)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeLocationCoordinates(r); err == nil {
		t.Fatal("expected rejection of unsupported shape")
	}
}
