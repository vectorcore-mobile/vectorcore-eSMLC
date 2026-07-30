package lpp

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/vectorcore/esmlc/internal/lpp/location"
	"github.com/vectorcore/esmlc/internal/uper"
)

type manifest struct {
	Fixtures []struct {
		Name   string `json:"name"`
		Binary string `json:"binary_file"`
		Bits   int    `json:"bit_length"`
		Unused int    `json:"unused_trailing_bits"`
	} `json:"fixtures"`
	Negative []struct {
		Name     string `json:"name"`
		Rejected bool   `json:"rejected"`
	} `json:"negative_cases"`
}

func TestFixtureOracle(t *testing.T) {
	root := filepath.Join("..", "..", "tools", "specs", "lpp", "fixtures", "r16.4.0")
	raw, e := os.ReadFile(filepath.Join(root, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	var m manifest
	if e = json.Unmarshal(raw, &m); e != nil {
		t.Fatal(e)
	}
	if len(m.Fixtures) != 11 || len(m.Negative) != 4 {
		t.Fatal("oracle count")
	}
	for _, n := range m.Negative {
		if !n.Rejected {
			t.Fatal(n.Name)
		}
	}
	for _, f := range m.Fixtures {
		t.Run(f.Name, func(t *testing.T) {
			data, e := os.ReadFile(filepath.Join(root, f.Binary))
			if e != nil {
				t.Fatal(e)
			}
			got, e := DecodeMessage(data, f.Bits)
			if e != nil {
				t.Fatal(e)
			}
			assertMeaning(t, f.Name, got)
			encoded, e := EncodeMessage(got)
			if e != nil {
				t.Fatal(e)
			}
			if string(encoded.Bytes) != string(data) || encoded.BitLength != f.Bits || int(encoded.UnusedBits) != f.Unused {
				t.Fatalf("encoded=%x %d/%d", encoded.Bytes, encoded.BitLength, encoded.UnusedBits)
			}
			again, e := DecodeMessage(encoded.Bytes, encoded.BitLength)
			if e != nil || !equalMessage(got, again) {
				t.Fatalf("round trip %v %+v", e, again)
			}
		})
	}
}
func assertMeaning(t *testing.T, n string, m Message) {
	t.Helper()
	switch n {
	case "minimal-no-transaction":
		if m.TransactionID != nil || m.EndTransaction || m.Body != nil {
			t.Fatal(m)
		}
	case "transaction-location-server-0":
		if m.TransactionID == nil || m.TransactionID.Initiator != InitiatorLocationServer || m.TransactionID.TransactionNumber != 0 || m.EndTransaction {
			t.Fatal(m)
		}
	case "transaction-target-device-255-end":
		if m.TransactionID == nil || m.TransactionID.Initiator != InitiatorTargetDevice || m.TransactionID.TransactionNumber != 255 || !m.EndTransaction {
			t.Fatal(m)
		}
	case "sequence-zero-ack-requested":
		if m.SequenceNumber == nil || *m.SequenceNumber != 0 || m.Acknowledgement == nil || !m.Acknowledgement.Requested {
			t.Fatal(m)
		}
	case "sequence-255-ack-indicator":
		if m.SequenceNumber == nil || *m.SequenceNumber != 255 || m.Acknowledgement == nil || m.Acknowledgement.Requested || m.Acknowledgement.Indicator == nil || *m.Acknowledgement.Indicator != 255 {
			t.Fatal(m)
		}
	default:
		want := map[string]BodyKind{"request-capabilities-r9-empty": 0, "provide-capabilities-r9-empty": 1, "request-location-information-r9-empty": 4, "provide-location-information-r9-empty": 5, "abort-r9-empty": 6, "error-r9-empty": 7}[n]
		if m.Body == nil || m.Body.Kind != want {
			t.Fatal(m)
		}
	}
}
func equalMessage(a, b Message) bool {
	if a.EndTransaction != b.EndTransaction || (a.TransactionID == nil) != (b.TransactionID == nil) || (a.SequenceNumber == nil) != (b.SequenceNumber == nil) || (a.Acknowledgement == nil) != (b.Acknowledgement == nil) || (a.Body == nil) != (b.Body == nil) {
		return false
	}
	if a.TransactionID != nil && *a.TransactionID != *b.TransactionID {
		return false
	}
	if a.SequenceNumber != nil && *a.SequenceNumber != *b.SequenceNumber {
		return false
	}
	if a.Acknowledgement != nil {
		if a.Acknowledgement.Requested != b.Acknowledgement.Requested || (a.Acknowledgement.Indicator == nil) != (b.Acknowledgement.Indicator == nil) {
			return false
		}
		if a.Acknowledgement.Indicator != nil && *a.Acknowledgement.Indicator != *b.Acknowledgement.Indicator {
			return false
		}
	}
	return a.Body == nil || a.Body.Kind == b.Body.Kind
}
func TestMalformed(t *testing.T) {
	if _, e := DecodeMessage([]byte{0}, 9); e == nil {
		t.Fatal("bit length")
	}
	if _, e := DecodeMessage([]byte{0x04}, 5); !errors.Is(e, ErrMalformed) {
		t.Fatal(e)
	}
	if _, e := EncodeMessage(Message{Body: &Body{Kind: 2}}); !errors.Is(e, ErrUnsupportedBody) {
		t.Fatal(e)
	}
	if _, e := DecodeMessage([]byte{0x10}, 5); e == nil {
		t.Fatal("truncation")
	}
}

func TestMessageECIDRequestLocationPayload(t *testing.T) {
	requested, err := uper.NewBitString([]byte{0xe0}, 3)
	if err != nil {
		t.Fatal(err)
	}
	in := Message{Body: &Body{Kind: BodyRequestLocationInformation, RequestLocationInformation: &location.RequestLocationInformationR9IEs{ECID: &location.ECIDRequestLocationInformation{RequestedMeasurements: requested}}}}
	encoded, err := EncodeMessage(in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeMessage(encoded.Bytes, encoded.BitLength)
	if err != nil {
		t.Fatal(err)
	}
	if out.Body == nil || out.Body.RequestLocationInformation == nil || out.Body.RequestLocationInformation.ECID == nil {
		t.Fatal("missing location payload")
	}
	if !out.Body.RequestLocationInformation.ECID.RequestedMeasurements.Equal(requested) {
		t.Fatal("requested measurements changed")
	}
}

func TestMessageECIDProvideLocationFixture(t *testing.T) {
	// Independent whole-LPP fixture generated under tools/specs/lpp/analysis;
	// it contains a target-device transaction and all root ECID optionals.
	data, err := hex.DecodeString("920f2809007c0200100848d159c0192aac317ff8")
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeMessage(data, 157)
	if err != nil {
		t.Fatal(err)
	}
	if out.TransactionID == nil || out.TransactionID.Initiator != InitiatorTargetDevice || out.TransactionID.TransactionNumber != 7 || !out.EndTransaction || out.Body == nil || out.Body.Kind != BodyProvideLocationInformation || out.Body.ProvideLocationInformation == nil || out.Body.ProvideLocationInformation.ECID == nil {
		t.Fatalf("unexpected decoded message: %#v", out)
	}
	encoded, err := EncodeMessage(out)
	if err != nil {
		t.Fatal(err)
	}
	if encoded.BitLength != 157 || string(encoded.Bytes) != string(data) {
		t.Fatalf("encoded=%x/%d", encoded.Bytes, encoded.BitLength)
	}
}

func TestDecodeMessageOctetsFindsUniqueUPERPadding(t *testing.T) {
	// This independently generated LPP message has 19 meaningful bits in three
	// transport octets. LCS-AP carries the octets, not that bit count.
	data, err := hex.DecodeString("110000")
	if err != nil {
		t.Fatal(err)
	}
	m, err := DecodeMessageOctets(data)
	if err != nil {
		t.Fatal(err)
	}
	if m.Body == nil || m.Body.Kind != BodyRequestLocationInformation {
		t.Fatalf("%#v", m)
	}
}
