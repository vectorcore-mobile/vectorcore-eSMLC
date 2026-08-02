package lcsap

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "tools", "specs", "lcsap", "fixtures", "r16.4.0", name+".aper"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestGeographicalAreaShapeFixtures compares every implemented Geographical-Area
// root shape against the independent asn1tools-cross-checked fixture corpus.
func TestGeographicalAreaShapeFixtures(t *testing.T) {
	coords := Coordinates{Latitude: 38, Longitude: -90}
	cases := map[string]func() ([]byte, error){
		"geo-area-point": func() ([]byte, error) { return EncodePoint(coords) },
		"geo-area-ellipse": func() ([]byte, error) {
			return EncodeEllipsoidPointWithUncertaintyEllipse(coords, UncertaintyEllipse{SemiMajor: 10, SemiMinor: 5, Orientation: 45}, 68)
		},
		"geo-area-polygon": func() ([]byte, error) {
			return EncodePolygon([]Coordinates{{38, -90}, {39, -91}, {37, -89}})
		},
		"geo-area-altitude": func() ([]byte, error) {
			return EncodeEllipsoidPointWithAltitude(coords, AltitudeAndDirection{Depth: false, Altitude: 120})
		},
		"geo-area-altitude-ellipsoid": func() ([]byte, error) {
			return EncodeEllipsoidPointWithAltitudeAndUncertaintyEllipsoid(coords, AltitudeAndDirection{Depth: false, Altitude: 120}, UncertaintyEllipse{SemiMajor: 10, SemiMinor: 5, Orientation: 45}, 3, 68)
		},
		"geo-area-arc": func() ([]byte, error) {
			return EncodeEllipsoidArc(coords, 100, 10, 30, 90, 68)
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

func TestGeographicalAreaRejectsInvalidInput(t *testing.T) {
	if _, err := EncodePoint(Coordinates{Latitude: 91, Longitude: 0}); err == nil {
		t.Fatal("expected error for out-of-range latitude")
	}
	if _, err := EncodeEllipsoidPointWithUncertaintyEllipse(Coordinates{}, UncertaintyEllipse{Orientation: 90}, 0); err == nil {
		t.Fatal("expected error for out-of-range orientation")
	}
	if _, err := EncodePolygon(nil); err == nil {
		t.Fatal("expected error for empty polygon")
	}
	if _, err := EncodePolygon(make([]Coordinates, 16)); err == nil {
		t.Fatal("expected error for oversized polygon")
	}
}
