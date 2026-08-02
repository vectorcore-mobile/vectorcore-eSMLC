package location

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/vectorcore/esmlc/internal/uper"
)

func otdoaFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "tools", "specs", "lpp", "fixtures", "r16.4.0", "otdoa", name+".uper"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestOTDOARequestLocationInformationFixture(t *testing.T) {
	data := otdoaFixture(t, "request-location-information-r9-otdoa-not-allowed")
	r, err := uper.NewReader(data, len(data)*8)
	if err != nil {
		t.Fatal(err)
	}
	v, err := DecodeRequestLocationInformation(r)
	if err != nil {
		t.Fatal(err)
	}
	if v.OTDOA == nil || v.OTDOA.AssistanceAvailability {
		t.Fatalf("expected OTDOA request with assistanceAvailability=false, got %#v", v.OTDOA)
	}
	w := uper.NewWriter()
	if err := EncodeRequestLocationInformation(w, v); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(w.Encoded().Bytes, data) {
		t.Fatalf("got %x want %x", w.Encoded().Bytes, data)
	}
}
