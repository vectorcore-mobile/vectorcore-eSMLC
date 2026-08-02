package capability

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

func TestAGNSSCapabilityFixtures(t *testing.T) {
	t.Run("request-capabilities-r9-agnss", func(t *testing.T) {
		data := agnssFixture(t, "request-capabilities-r9-agnss")
		r, err := uper.NewReader(data, len(data)*8)
		if err != nil {
			t.Fatal(err)
		}
		v, err := DecodeRequestCapabilities(r)
		if err != nil {
			t.Fatal(err)
		}
		if v.AGNSS == nil || !v.AGNSS.GNSSSupportListReq || v.AGNSS.AssistanceDataSupportListReq || v.AGNSS.LocationVelocityTypesReq {
			t.Fatalf("bad A-GNSS request capabilities: %#v", v.AGNSS)
		}
		w := uper.NewWriter()
		if err := EncodeRequestCapabilities(w, v); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(w.Encoded().Bytes, data) {
			t.Fatalf("got %x want %x", w.Encoded().Bytes, data)
		}
	})
	t.Run("provide-capabilities-r9-agnss-gps-ue-based", func(t *testing.T) {
		data := agnssFixture(t, "provide-capabilities-r9-agnss-gps-ue-based")
		r, err := uper.NewReader(data, len(data)*8)
		if err != nil {
			t.Fatal(err)
		}
		v, err := DecodeProvideCapabilities(r)
		if err != nil {
			t.Fatal(err)
		}
		if v.AGNSS == nil || !v.AGNSS.SupportsGPSUEBased() {
			t.Fatalf("expected GPS ue-based support, got %#v", v.AGNSS)
		}
		if len(v.AGNSS.GNSSSupportList) != 1 || v.AGNSS.GNSSSupportList[0].ID != GNSSIDGPS {
			t.Fatalf("unexpected support list: %#v", v.AGNSS.GNSSSupportList)
		}
		w := uper.NewWriter()
		if err := EncodeProvideCapabilities(w, v); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(w.Encoded().Bytes, data) {
			t.Fatalf("got %x want %x", w.Encoded().Bytes, data)
		}
	})
}
