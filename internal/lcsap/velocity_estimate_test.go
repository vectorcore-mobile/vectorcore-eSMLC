package lcsap

import (
	"bytes"
	"testing"
)

// TestVelocityEstimateShapeFixtures compares every implemented
// Velocity-Estimate root shape against the independent asn1tools-cross-checked
// fixture corpus.
func TestVelocityEstimateShapeFixtures(t *testing.T) {
	h := HorizontalSpeedAndBearing{Bearing: 45, Speed: 30}
	v := VerticalVelocity{Speed: 5, Downward: false}
	cases := map[string]func() ([]byte, error){
		"velocity-horizontal": func() ([]byte, error) { return EncodeHorizontalVelocity(h) },
		"velocity-horizontal-with-vertical": func() ([]byte, error) {
			return EncodeHorizontalWithVerticalVelocity(h, v)
		},
		"velocity-horizontal-with-uncertainty": func() ([]byte, error) {
			return EncodeHorizontalVelocityWithUncertainty(h, 7)
		},
		"velocity-horizontal-with-vertical-and-uncertainty": func() ([]byte, error) {
			return EncodeHorizontalWithVerticalVelocityAndUncertainty(h, v, 7, 3)
		},
	}
	for name, encode := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := encode()
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			if want := fixture(t, name); !bytes.Equal(got, want) {
				t.Fatalf("wire mismatch\n got %x\nwant %x", got, want)
			}
		})
	}
}

func TestVelocityEstimateRejectsInvalidInput(t *testing.T) {
	if _, err := EncodeHorizontalVelocity(HorizontalSpeedAndBearing{Bearing: 360}); err == nil {
		t.Fatal("expected error for out-of-range bearing")
	}
	if _, err := EncodeHorizontalVelocity(HorizontalSpeedAndBearing{Speed: 2048}); err == nil {
		t.Fatal("expected error for out-of-range speed")
	}
}
