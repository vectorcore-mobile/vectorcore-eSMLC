package uper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// This is a test-only mapping of the eleven independent envelope fixtures. It
// is deliberately not exported and is not an LPP production model.
type fixtureMessage struct {
	tx       *fixtureTransaction
	end      bool
	sequence *uint64
	ack      *fixtureAck
	body     *fixtureBody
}
type fixtureTransaction struct {
	initiator uint64
	number    uint64
}
type fixtureAck struct {
	requested bool
	indicator *uint64
}
type fixtureBody struct{ kind uint64 }

const (
	bodyRequestCapabilities uint64 = 0
	bodyProvideCapabilities uint64 = 1
	bodyRequestLocationInfo uint64 = 4
	bodyProvideLocationInfo uint64 = 5
	bodyAbort               uint64 = 6
	bodyError               uint64 = 7
)

func decodeFixtureMessage(r *Reader) (fixtureMessage, error) {
	bits, err := r.ReadOptionalBitmap(4)
	if err != nil {
		return fixtureMessage{}, err
	}
	m := fixtureMessage{}
	if bits[0] {
		ext, err := r.ReadExtensionPresent()
		if err != nil {
			return m, err
		}
		if err = RequireNoExtension(ext); err != nil {
			return m, err
		}
		ext, err = r.ReadExtensionPresent()
		if err != nil {
			return m, err
		}
		if err = RequireNoExtension(ext); err != nil {
			return m, err
		}
		initiator, err := r.ReadRootEnumerated(2)
		if err != nil {
			return m, err
		}
		number, err := r.ReadConstrainedWholeNumber(0, 255)
		if err != nil {
			return m, err
		}
		m.tx = &fixtureTransaction{initiator, number}
	}
	m.end, err = r.ReadBoolean()
	if err != nil {
		return m, err
	}
	if bits[1] {
		value, err := r.ReadConstrainedWholeNumber(0, 255)
		if err != nil {
			return m, err
		}
		m.sequence = &value
	}
	if bits[2] {
		optional, err := r.ReadOptionalBitmap(1)
		if err != nil {
			return m, err
		}
		requested, err := r.ReadBoolean()
		if err != nil {
			return m, err
		}
		m.ack = &fixtureAck{requested: requested}
		if optional[0] {
			value, err := r.ReadConstrainedWholeNumber(0, 255)
			if err != nil {
				return m, err
			}
			m.ack.indicator = &value
		}
	}
	if bits[3] {
		outer, err := r.ReadRootChoiceIndex(2)
		if err != nil {
			return m, err
		}
		if outer != 0 {
			return m, ErrExtensionUnsupported
		}
		kind, err := r.ReadRootChoiceIndex(16)
		if err != nil {
			return m, err
		}
		switch kind {
		case bodyRequestCapabilities, bodyProvideCapabilities, bodyRequestLocationInfo, bodyProvideLocationInfo, bodyAbort:
			critical, err := r.ReadRootChoiceIndex(2)
			if err != nil {
				return m, err
			}
			if critical != 0 {
				return m, ErrExtensionUnsupported
			}
			release, err := r.ReadRootChoiceIndex(4)
			if err != nil {
				return m, err
			}
			if release != 0 {
				return m, ErrExtensionUnsupported
			}
			rootOptionals := 5
			if kind == bodyAbort {
				rootOptionals = 1
			}
			ext, err := r.ReadExtensionPresent()
			if err != nil {
				return m, err
			}
			if err = RequireNoExtension(ext); err != nil {
				return m, err
			}
			optional, err := r.ReadOptionalBitmap(rootOptionals)
			if err != nil {
				return m, err
			}
			for _, present := range optional {
				if present {
					return m, ErrExtensionUnsupported
				}
			}
		case bodyError:
			release, err := r.ReadRootChoiceIndex(2)
			if err != nil {
				return m, err
			}
			if release != 0 {
				return m, ErrExtensionUnsupported
			}
			ext, err := r.ReadExtensionPresent()
			if err != nil {
				return m, err
			}
			if err = RequireNoExtension(ext); err != nil {
				return m, err
			}
			optional, err := r.ReadOptionalBitmap(1)
			if err != nil {
				return m, err
			}
			if optional[0] {
				return m, ErrExtensionUnsupported
			}
		default:
			return m, ErrExtensionUnsupported
		}
		m.body = &fixtureBody{kind: kind}
	}
	return m, nil
}

func encodeFixtureMessage(m fixtureMessage) (Encoded, error) {
	w := NewWriter()
	if err := w.WriteOptionalBitmap([]bool{m.tx != nil, m.sequence != nil, m.ack != nil, m.body != nil}); err != nil {
		return Encoded{}, err
	}
	if m.tx != nil {
		if err := w.WriteExtensionPresent(false); err != nil {
			return Encoded{}, err
		}
		if err := w.WriteExtensionPresent(false); err != nil {
			return Encoded{}, err
		}
		if err := w.WriteRootEnumerated(m.tx.initiator, 2); err != nil {
			return Encoded{}, err
		}
		if err := w.WriteConstrainedWholeNumber(m.tx.number, 0, 255); err != nil {
			return Encoded{}, err
		}
	}
	if err := w.WriteBoolean(m.end); err != nil {
		return Encoded{}, err
	}
	if m.sequence != nil {
		if err := w.WriteConstrainedWholeNumber(*m.sequence, 0, 255); err != nil {
			return Encoded{}, err
		}
	}
	if m.ack != nil {
		if err := w.WriteOptionalBitmap([]bool{m.ack.indicator != nil}); err != nil {
			return Encoded{}, err
		}
		if err := w.WriteBoolean(m.ack.requested); err != nil {
			return Encoded{}, err
		}
		if m.ack.indicator != nil {
			if err := w.WriteConstrainedWholeNumber(*m.ack.indicator, 0, 255); err != nil {
				return Encoded{}, err
			}
		}
	}
	if m.body != nil {
		if err := w.WriteRootChoiceIndex(0, 2); err != nil {
			return Encoded{}, err
		}
		if err := w.WriteRootChoiceIndex(m.body.kind, 16); err != nil {
			return Encoded{}, err
		}
		switch m.body.kind {
		case bodyRequestCapabilities, bodyProvideCapabilities, bodyRequestLocationInfo, bodyProvideLocationInfo, bodyAbort:
			if err := w.WriteRootChoiceIndex(0, 2); err != nil {
				return Encoded{}, err
			}
			if err := w.WriteRootChoiceIndex(0, 4); err != nil {
				return Encoded{}, err
			}
			if err := w.WriteExtensionPresent(false); err != nil {
				return Encoded{}, err
			}
			count := 5
			if m.body.kind == bodyAbort {
				count = 1
			}
			if err := w.WriteOptionalBitmap(make([]bool, count)); err != nil {
				return Encoded{}, err
			}
		case bodyError:
			if err := w.WriteRootChoiceIndex(0, 2); err != nil {
				return Encoded{}, err
			}
			if err := w.WriteExtensionPresent(false); err != nil {
				return Encoded{}, err
			}
			if err := w.WriteOptionalBitmap([]bool{false}); err != nil {
				return Encoded{}, err
			}
		default:
			return Encoded{}, ErrExtensionUnsupported
		}
	}
	return w.Encoded(), nil
}

func assertFixtureSemantics(t *testing.T, name string, m fixtureMessage) {
	t.Helper()
	switch name {
	case "minimal-no-transaction":
		if m.tx != nil || m.end || m.sequence != nil || m.ack != nil || m.body != nil {
			t.Fatalf("unexpected minimal value: %+v", m)
		}
	case "transaction-location-server-0":
		if m.tx == nil || m.tx.initiator != 0 || m.tx.number != 0 || m.end {
			t.Fatalf("transaction lower boundary: %+v", m)
		}
	case "transaction-target-device-255-end":
		if m.tx == nil || m.tx.initiator != 1 || m.tx.number != 255 || !m.end {
			t.Fatalf("transaction upper boundary: %+v", m)
		}
	case "sequence-zero-ack-requested":
		if m.sequence == nil || *m.sequence != 0 || m.ack == nil || !m.ack.requested || m.ack.indicator != nil {
			t.Fatalf("sequence/ack request: %+v", m)
		}
	case "sequence-255-ack-indicator":
		if m.sequence == nil || *m.sequence != 255 || m.ack == nil || m.ack.requested || m.ack.indicator == nil || *m.ack.indicator != 255 {
			t.Fatalf("sequence/ack response: %+v", m)
		}
	default:
		expected := map[string]uint64{
			"request-capabilities-r9-empty":         bodyRequestCapabilities,
			"provide-capabilities-r9-empty":         bodyProvideCapabilities,
			"request-location-information-r9-empty": bodyRequestLocationInfo,
			"provide-location-information-r9-empty": bodyProvideLocationInfo,
			"abort-r9-empty":                        bodyAbort,
			"error-r9-empty":                        bodyError,
		}[name]
		if m.body == nil || m.body.kind != expected {
			t.Fatalf("body %q: %+v", name, m)
		}
	}
}

type fixtureManifest struct {
	Fixtures []struct {
		Name       string `json:"name"`
		BinaryFile string `json:"binary_file"`
		BitLength  int    `json:"bit_length"`
		Unused     int    `json:"unused_trailing_bits"`
	} `json:"fixtures"`
	NegativeCases []struct {
		Name     string
		Rejected bool
	} `json:"negative_cases"`
}

func TestIndependentFixtureOracle(t *testing.T) {
	root := filepath.Join("..", "..", "tools", "specs", "lpp", "fixtures", "r16.4.0")
	manifestData, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest fixtureManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Fixtures) != 11 || len(manifest.NegativeCases) != 4 {
		t.Fatalf("oracle count: %d valid %d negative", len(manifest.Fixtures), len(manifest.NegativeCases))
	}
	for _, negative := range manifest.NegativeCases {
		if !negative.Rejected {
			t.Fatalf("reference negative case was not rejected: %s", negative.Name)
		}
	}
	for _, fixture := range manifest.Fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, fixture.BinaryFile))
			if err != nil {
				t.Fatal(err)
			}
			r, err := NewReader(data, fixture.BitLength)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := decodeFixtureMessage(r)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			assertFixtureSemantics(t, fixture.Name, decoded)
			if err := r.ValidateFinalPadding(); err != nil {
				t.Fatalf("padding: %v", err)
			}
			encoded, err := encodeFixtureMessage(decoded)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			if string(encoded.Bytes) != string(data) {
				t.Fatalf("bytes %x want %x", encoded.Bytes, data)
			}
			if encoded.BitLength != fixture.BitLength || int(encoded.UnusedBits) != fixture.Unused {
				t.Fatalf("length %+v fixture=%d/%d", encoded, fixture.BitLength, fixture.Unused)
			}
		})
	}
}

func TestFixtureSchemaRejectsUnsupportedExtension(t *testing.T) {
	w := NewWriter()
	_ = w.WriteOptionalBitmap([]bool{true, false, false, false})
	_ = w.WriteExtensionPresent(true)
	r, _ := NewReader(w.Encoded().Bytes, w.Encoded().BitLength)
	if _, err := decodeFixtureMessage(r); err == nil {
		t.Fatal("expected extension rejection")
	}
}

func Example_fixtureMessage() { fmt.Print("test-only") /* Output: test-only */ }
