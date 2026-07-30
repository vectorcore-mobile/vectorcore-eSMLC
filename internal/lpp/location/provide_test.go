package location

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/vectorcore/esmlc/internal/lpp/location/result"
	"github.com/vectorcore/esmlc/internal/uper"
)

func measured(t *testing.T, pci uint16, arfcn uint32, opt result.MeasuredResultsElementOptions) result.MeasuredResultsElement {
	t.Helper()
	p, err := result.NewPhysicalCellID(pci)
	if err != nil {
		t.Fatal(err)
	}
	a, err := result.NewEUTRAARFCN(arfcn)
	if err != nil {
		t.Fatal(err)
	}
	v, err := result.NewMeasuredResultsElement(p, a, opt)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func global(t *testing.T, mcc string, mnc string, utra bool, identity uint32) result.CellGlobalIdEUTRAAndUTRA {
	t.Helper()
	m, err := result.NewMCC(mcc[0]-'0', mcc[1]-'0', mcc[2]-'0')
	if err != nil {
		t.Fatal(err)
	}
	var n result.MNC
	if len(mnc) == 2 {
		n, err = result.NewMNC2(mnc[0]-'0', mnc[1]-'0')
	} else {
		n, err = result.NewMNC3(mnc[0]-'0', mnc[1]-'0', mnc[2]-'0')
	}
	if err != nil {
		t.Fatal(err)
	}
	p, err := result.NewPLMNIdentity(m, n)
	if err != nil {
		t.Fatal(err)
	}
	if utra {
		x, err := result.NewUTRACellIdentityFromUint32(identity)
		if err != nil {
			t.Fatal(err)
		}
		v, err := result.NewCellGlobalIdEUTRAAndUTRA(p, result.NewUTRACellIdentityChoice(x))
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	x, err := result.NewEUTRACellIdentityFromUint32(identity)
	if err != nil {
		t.Fatal(err)
	}
	v, err := result.NewCellGlobalIdEUTRAAndUTRA(p, result.NewEUTRACellIdentityChoice(x))
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func providePayload(t *testing.T, name string) ProvideLocationInformationR9IEs {
	t.Helper()
	var primary *result.MeasuredResultsElement
	var values []result.MeasuredResultsElement
	switch name {
	case "one-rsrp":
		r := result.RSRPResult(30)
		values = []result.MeasuredResultsElement{measured(t, 1, 100, result.MeasuredResultsElementOptions{RSRPResult: &r})}
	case "list-two-mixed":
		p := measured(t, 0, 0, result.MeasuredResultsElementOptions{})
		primary = &p
		rq, ue := result.RSRQResult(34), result.UERxTxTimeDiff(4095)
		values = []result.MeasuredResultsElement{measured(t, 1, 100, result.MeasuredResultsElementOptions{}), measured(t, 503, 65535, result.MeasuredResultsElementOptions{RSRQResult: &rq, UERxTxTimeDiff: &ue})}
	case "all-optionals-eutra":
		cell := global(t, "001", "01", false, 0x1234567)
		sfn, err := result.NewSystemFrameNumberFromUint16(0x2aa)
		if err != nil {
			t.Fatal(err)
		}
		rp, rq, ue := result.RSRPResult(97), result.RSRQResult(34), result.UERxTxTimeDiff(4095)
		values = []result.MeasuredResultsElement{measured(t, 1, 100, result.MeasuredResultsElementOptions{CellGlobalID: &cell, SystemFrameNumber: &sfn, RSRPResult: &rp, RSRQResult: &rq, UERxTxTimeDiff: &ue})}
	case "utravariant":
		cell := global(t, "310", "260", true, 0x12345678)
		values = []result.MeasuredResultsElement{measured(t, 503, 65535, result.MeasuredResultsElementOptions{CellGlobalID: &cell})}
	default:
		t.Fatalf("unknown fixture %q", name)
	}
	signal, err := NewECIDSignalMeasurementInformation(primary, values)
	if err != nil {
		t.Fatal(err)
	}
	ecid, err := NewECIDProvideLocationInformation(signal)
	if err != nil {
		t.Fatal(err)
	}
	return ProvideLocationInformationR9IEs{ECID: &ecid}
}

func TestProvideLocationInformationIndependentFixtures(t *testing.T) {
	for _, tc := range []struct {
		name, hex string
		bits      int
	}{
		{"one-rsrp", "0120020040190f00", 57},
		{"list-two-mixed", "0124000000004000803207f7ffff8bffc0", 130},
		{"all-optionals-eutra", "01200f80400201091a2b38032555862fff", 136},
		{"utravariant", "0120087dc621260891a2b3c7fff8", 109},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want, _ := hex.DecodeString(tc.hex)
			in := providePayload(t, tc.name)
			w := uper.NewWriter()
			if err := EncodeProvideLocationInformation(w, in); err != nil {
				t.Fatal(err)
			}
			got := w.Encoded()
			if got.BitLength != tc.bits || string(got.Bytes) != string(want) {
				t.Fatalf("%x/%d want %x/%d", got.Bytes, got.BitLength, want, tc.bits)
			}
			r, err := uper.NewReader(want, tc.bits)
			if err != nil {
				t.Fatal(err)
			}
			out, err := DecodeProvideLocationInformation(r)
			if err != nil || r.ValidateFinalPadding() != nil {
				t.Fatalf("decode: %v", err)
			}
			w = uper.NewWriter()
			if err = EncodeProvideLocationInformation(w, out); err != nil {
				t.Fatal(err)
			}
			again := w.Encoded()
			if again.BitLength != tc.bits || string(again.Bytes) != string(want) {
				t.Fatalf("re-encode %x/%d", again.Bytes, again.BitLength)
			}
		})
	}
}

func TestECIDSignalMeasurementInformationBoundsAndCopies(t *testing.T) {
	v := measured(t, 0, 0, result.MeasuredResultsElementOptions{})
	if _, err := NewECIDSignalMeasurementInformation(nil, nil); !errors.Is(err, ErrMissingMeasuredResults) {
		t.Fatal(err)
	}
	tooMany := make([]result.MeasuredResultsElement, 33)
	for i := range tooMany {
		tooMany[i] = v
	}
	if _, err := NewECIDSignalMeasurementInformation(nil, tooMany); !errors.Is(err, ErrMissingMeasuredResults) {
		t.Fatal(err)
	}
	maximum := make([]result.MeasuredResultsElement, 32)
	for i := range maximum {
		maximum[i] = v
	}
	maxSignal, err := NewECIDSignalMeasurementInformation(nil, maximum)
	if err != nil {
		t.Fatal(err)
	}
	w := uper.NewWriter()
	if err = maxSignal.EncodeUPER(w); err != nil {
		t.Fatal(err)
	}
	r, _ := uper.NewReader(w.Encoded().Bytes, w.Encoded().BitLength)
	decoded, err := DecodeECIDSignalMeasurementInformation(r)
	if err != nil || len(decoded.MeasuredResults()) != 32 {
		t.Fatalf("max list: %v", err)
	}
	signal, err := NewECIDSignalMeasurementInformation(nil, []result.MeasuredResultsElement{v})
	if err != nil {
		t.Fatal(err)
	}
	copy := signal.MeasuredResults()
	copy[0] = result.MeasuredResultsElement{}
	if len(signal.MeasuredResults()) != 1 || signal.MeasuredResults()[0].PhysicalCellID() != 0 {
		t.Fatal("result slice aliases")
	}
}

func TestProvideLocationInformationExtensionAndTruncation(t *testing.T) {
	for _, stream := range [][]byte{{0x80}, {0x01, 0x18}} {
		r, _ := uper.NewReader(stream, len(stream)*8)
		_, _ = DecodeProvideLocationInformation(r)
	}
	in := providePayload(t, "utravariant")
	w := uper.NewWriter()
	if err := EncodeProvideLocationInformation(w, in); err != nil {
		t.Fatal(err)
	}
	x := w.Encoded()
	for n := 0; n < x.BitLength; n++ {
		r, _ := uper.NewReader(x.Bytes, n)
		if _, err := DecodeProvideLocationInformation(r); err == nil {
			t.Fatalf("truncation at %d", n)
		}
	}
	// The r9 ECID container's second root optional bit is ecid-Error.
	r, _ := uper.NewReader([]byte{0x00, 0x18}, 13)
	_, _ = r.ReadRootChoiceIndex(2)
	_, _ = r.ReadRootChoiceIndex(4)
	_, _ = r.ReadExtensionPresent()
	_, _ = r.ReadOptionalBitmap(5)
	// Explicit extension-bearing ECID data must be rejected by the direct decoder.
	r, _ = uper.NewReader([]byte{0x80}, 1)
	if _, err := DecodeECIDProvideLocationInformation(r); !errors.Is(err, ErrUnsupportedExtension) {
		t.Fatal(err)
	}
}

func TestProvideLocationInformationNonAligned(t *testing.T) {
	in := providePayload(t, "one-rsrp")
	w := uper.NewWriter()
	_ = w.WriteBits(5, 3)
	if err := EncodeProvideLocationInformation(w, in); err != nil {
		t.Fatal(err)
	}
	_ = w.WriteBits(17, 5)
	x := w.Encoded()
	if x.BitLength != 65 {
		t.Fatalf("bits %d", x.BitLength)
	}
	r, _ := uper.NewReader(x.Bytes, x.BitLength)
	if n, _ := r.ReadBits(3); n != 5 {
		t.Fatal(n)
	}
	if _, err := DecodeProvideLocationInformation(r); err != nil {
		t.Fatal(err)
	}
	if n, _ := r.ReadBits(5); n != 17 {
		t.Fatal(n)
	}
}

func FuzzECIDProvideLocationRoundTrip(f *testing.F) {
	f.Add(uint8(1), uint16(100), uint8(30))
	f.Fuzz(func(t *testing.T, pci uint8, arfcn uint16, rsrp uint8) {
		if pci > 255 || rsrp > 97 {
			return
		}
		r := result.RSRPResult(rsrp)
		v := measured(t, uint16(pci), uint32(arfcn), result.MeasuredResultsElementOptions{RSRPResult: &r})
		s, err := NewECIDSignalMeasurementInformation(nil, []result.MeasuredResultsElement{v})
		if err != nil {
			t.Fatal(err)
		}
		e, err := NewECIDProvideLocationInformation(s)
		if err != nil {
			t.Fatal(err)
		}
		w := uper.NewWriter()
		if err = EncodeProvideLocationInformation(w, ProvideLocationInformationR9IEs{ECID: &e}); err != nil {
			t.Fatal(err)
		}
		x := w.Encoded()
		reader, _ := uper.NewReader(x.Bytes, x.BitLength)
		if _, err = DecodeProvideLocationInformation(reader); err != nil {
			t.Fatal(err)
		}
	})
}

func FuzzDecodeECIDProvideLocation(f *testing.F) {
	f.Add([]byte{0x01, 0x20, 0x02, 0x00, 0x40, 0x19, 0x0f, 0x00}, uint8(57))
	f.Add([]byte{0x80}, uint8(1))
	f.Fuzz(func(t *testing.T, data []byte, bitLength uint8) {
		if len(data) > 512 {
			return
		}
		bits := int(bitLength)
		if bits > len(data)*8 {
			bits = len(data) * 8
		}
		r, err := uper.NewReader(data, bits)
		if err != nil {
			return
		}
		_, _ = DecodeProvideLocationInformation(r)
	})
}
