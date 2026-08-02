package result

import (
	"testing"
	"time"

	"github.com/vectorcore/esmlc/internal/uper"
)

// TestRSTDAnchorLowerEdges locks the index-to-lower-edge formula against
// every numerically checkable row of TS 36.133 Table 9.1.10.3-1 that survived
// docx table extraction (the middle of each region is elided by "..." in the
// source, but the rule is fully determined by the explicit prose plus these
// anchors).
func TestRSTDAnchorLowerEdges(t *testing.T) {
	cases := []struct {
		index       RSTD
		lowerEdgeTs int
	}{
		{1, -15391},    // RSTD_0001: "-15391 <= RSTD < -15386"
		{2259, -4101},  // RSTD_2259: "-4101 <= RSTD < -4096"
		{2260, -4096},  // RSTD_2260: start of 1Ts fine region
		{10451, 4095},  // RSTD_10451: end of 1Ts fine region
		{10452, 4096},  // RSTD_10452: start of positive 5Ts coarse region
		{12710, 15386}, // RSTD_12710: "15386 < RSTD <= 15391"
	}
	for _, c := range cases {
		got := c.index.Duration()
		wantMid := float64(c.lowerEdgeTs)
		if c.index >= rstdFineLowIndex && c.index <= rstdFineHighIndex {
			wantMid += 0.5
		} else {
			wantMid += 2.5
		}
		want := time.Duration(wantMid * float64(Ts))
		if got != want {
			t.Fatalf("index %d: got %v want %v (%.1f Ts)", c.index, got, want, wantMid)
		}
	}
}

// TestRSTDRoundTrip checks every index except the two saturating extremes:
// those represent an unbounded magnitude, so Duration's boundary
// representative does not itself decode back to the extreme index (it
// legitimately falls just inside the adjacent bucket instead). Saturation
// behavior for genuinely out-of-range durations is covered separately by
// TestRSTDSaturation.
func TestRSTDRoundTrip(t *testing.T) {
	for i := RSTD(1); i < MaxRSTD; i++ {
		got := DurationToRSTD(i.Duration())
		if got != i {
			t.Fatalf("index %d: round trip gave %d (duration %v)", i, got, i.Duration())
		}
	}
}

func TestRSTDMonotonic(t *testing.T) {
	prev := RSTD(0).Duration()
	for i := RSTD(1); i <= MaxRSTD; i++ {
		cur := i.Duration()
		if cur <= prev {
			t.Fatalf("index %d duration %v not strictly greater than index %d duration %v", i, cur, i-1, prev)
		}
		prev = cur
	}
}

func TestRSTDSaturation(t *testing.T) {
	if got := DurationToRSTD(-1000 * rstdReportingLimit * Ts); got != 0 {
		t.Fatalf("expected saturation to 0, got %d", got)
	}
	if got := DurationToRSTD(1000 * rstdReportingLimit * Ts); got != MaxRSTD {
		t.Fatalf("expected saturation to MaxRSTD, got %d", got)
	}
}

func TestRSTDEncodeDecodeRoundTrip(t *testing.T) {
	for _, raw := range []uint16{0, 1, 2260, 6356, 10451, 12710, 12711} {
		v, err := NewRSTD(raw)
		if err != nil {
			t.Fatal(err)
		}
		w := uper.NewWriter()
		if err := v.EncodeUPER(w); err != nil {
			t.Fatal(err)
		}
		r, err := uper.NewReader(w.Encoded().Bytes, w.BitLength())
		if err != nil {
			t.Fatal(err)
		}
		got, err := DecodeRSTD(r)
		if err != nil {
			t.Fatal(err)
		}
		if got != v {
			t.Fatalf("raw %d: round trip gave %d", raw, got)
		}
	}
}

func TestRSTDRejectsOutOfRange(t *testing.T) {
	if _, err := NewRSTD(12712); err == nil {
		t.Fatal("expected error for RSTD above range")
	}
}
