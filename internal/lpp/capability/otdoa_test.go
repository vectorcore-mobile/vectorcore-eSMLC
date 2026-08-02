package capability

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

func TestOTDOACapabilityFixtures(t *testing.T) {
	t.Run("request-capabilities-r9-otdoa-empty", func(t *testing.T) {
		data := otdoaFixture(t, "request-capabilities-r9-otdoa-empty")
		r, err := uper.NewReader(data, len(data)*8)
		if err != nil {
			t.Fatal(err)
		}
		v, err := DecodeRequestCapabilities(r)
		if err != nil {
			t.Fatal(err)
		}
		if v.OTDOA == nil {
			t.Fatal("expected OTDOA request capability selector")
		}
		w := uper.NewWriter()
		if err := EncodeRequestCapabilities(w, v); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(w.Encoded().Bytes, data) {
			t.Fatalf("got %x want %x", w.Encoded().Bytes, data)
		}
	})
	t.Run("provide-capabilities-r9-otdoa", func(t *testing.T) {
		data := otdoaFixture(t, "provide-capabilities-r9-otdoa")
		r, err := uper.NewReader(data, len(data)*8)
		if err != nil {
			t.Fatal(err)
		}
		v, err := DecodeProvideCapabilities(r)
		if err != nil {
			t.Fatal(err)
		}
		if v.OTDOA == nil || !v.OTDOA.SupportsUEAssisted() || v.OTDOA.Mode.BitLen() != 1 {
			t.Fatalf("bad OTDOA provide capabilities: %#v", v.OTDOA)
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
