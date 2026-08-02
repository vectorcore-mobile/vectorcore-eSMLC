package result

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/vectorcore/esmlc/internal/uper"
)

var (
	ErrRSTDOutOfRange = errors.New("lpp location result: RSTD out of range")
	ErrRSTDEncode     = errors.New("lpp location result: RSTD encode failed")
	ErrRSTDDecode     = errors.New("lpp location result: RSTD decode failed")
)

// RSTD is TS 37.355's coded neighbour-cell Reference Signal Time Difference,
// INTEGER(0..12711). The wire value is a quantization-bucket index, not a
// physical time difference directly: TS 37.355's own field description says
// only "Mapping of the measured quantity is defined as in TS 36.133 [18]
// clause 9.1.10.3" and does not restate the table. That clause (staged at
// docs/specs/36133-g40.zip, section 0-11 document, Table 9.1.10.3-1) defines
// a reporting range of -15391Ts to 15391Ts, with 1Ts resolution for
// |RSTD|<=4096Ts and 5Ts resolution beyond that up to the reporting limit,
// saturating outside it. Ts = 1/(15000*2048) seconds is the LTE basic time
// unit used throughout TS 36.133 clause 9.1.10.
//
// The source table's own boundary convention is not uniformly half-open the
// same way on both sides of zero (its printed rows mix "lower <= RSTD <
// upper" and "lower < RSTD <= upper" depending on sign, consistent with a
// symmetric round-half-away-from-zero quantizer). This package instead uses
// one uniform convention, bucket k covers [lower_edge(k), lower_edge(k) +
// width), verified numerically against the table's explicit lower-edge
// values at every checkable anchor row (RSTD_0001=-15391, RSTD_2259=-4101,
// RSTD_2260=-4096, RSTD_10451=4095, RSTD_10452=4096, RSTD_12710=15386). The
// two conventions disagree only at exact bucket-boundary values, worth at
// most one bucket width (<=9.8m at 1Ts, <=49m at 5Ts) — far below realistic
// OTDOA measurement uncertainty — so this is a documented implementation
// choice, not a source of positioning error.
type RSTD uint16

// MaxRSTD is the wire-format upper bound, TS 37.355 rstd INTEGER(0..12711).
const MaxRSTD RSTD = 12711

// Ts is the LTE basic time unit, TS 36.133 clause 9.1.10 ("Ts=1/(15000*2048)
// seconds").
const Ts = time.Second / (15000 * 2048)

const (
	rstdFineHalfWidth  = 4096                                  // Ts; |RSTD|<=this uses 1Ts resolution
	rstdCoarseStep     = 5                                     // Ts; resolution beyond rstdFineHalfWidth
	rstdReportingLimit = 15391                                 // Ts; reporting range boundary
	rstdZeroIndex      = 6356                                  // index representing the [0,1)Ts fine bucket
	rstdFineLowIndex   = rstdZeroIndex - rstdFineHalfWidth     // 2260
	rstdFineHighIndex  = rstdZeroIndex + rstdFineHalfWidth - 1 // 10451
)

func NewRSTD(value uint16) (RSTD, error) {
	v := RSTD(value)
	if err := v.Validate(); err != nil {
		return 0, err
	}
	return v, nil
}

func (v RSTD) Validate() error {
	if v > MaxRSTD {
		return fmt.Errorf("%w: %d outside 0..%d", ErrRSTDOutOfRange, v, MaxRSTD)
	}
	return nil
}

// Duration returns this index's representative physical time difference: the
// midpoint of its TS 36.133 Table 9.1.10.3-1 bucket. Index 0 and MaxRSTD are
// saturating extremes whose true magnitude is unbounded; Duration returns
// the +-15391Ts reporting boundary for those, not a midpoint.
func (v RSTD) Duration() time.Duration {
	switch {
	case v == 0:
		return -rstdReportingLimit * Ts
	case v == MaxRSTD:
		return rstdReportingLimit * Ts
	case int(v) >= rstdFineLowIndex && int(v) <= rstdFineHighIndex:
		lower := int(v) - rstdZeroIndex
		return time.Duration(lower)*Ts + Ts/2
	case int(v) < rstdFineLowIndex:
		k := int(v)
		lower := -rstdReportingLimit + (k-1)*rstdCoarseStep
		return time.Duration(lower)*Ts + rstdCoarseStep*Ts/2
	default: // rstdFineHighIndex < v < MaxRSTD
		k := int(v)
		lower := rstdFineHalfWidth + (k-(rstdFineHighIndex+1))*rstdCoarseStep
		return time.Duration(lower)*Ts + rstdCoarseStep*Ts/2
	}
}

// DurationToRSTD quantizes a physical time difference into the TS 36.133
// clause 9.1.10.3 coded index, saturating at the +-15391Ts reporting limit.
// It is the Duration inverse used to build test/simulation measurements from
// known geometry; a live UE performs this quantization itself.
func DurationToRSTD(d time.Duration) RSTD {
	ts := float64(d) / float64(Ts)
	switch {
	case ts < -rstdReportingLimit:
		return 0
	case ts >= rstdReportingLimit:
		return MaxRSTD
	case ts >= -rstdFineHalfWidth && ts < rstdFineHalfWidth:
		return RSTD(clampInt(rstdZeroIndex+int(math.Floor(ts)), rstdFineLowIndex, rstdFineHighIndex))
	case ts < -rstdFineHalfWidth:
		k := 1 + int(math.Floor((ts+rstdReportingLimit)/rstdCoarseStep))
		return RSTD(clampInt(k, 1, rstdFineLowIndex-1))
	default: // ts >= rstdFineHalfWidth
		k := (rstdFineHighIndex + 1) + int(math.Floor((ts-rstdFineHalfWidth)/rstdCoarseStep))
		return RSTD(clampInt(k, rstdFineHighIndex+1, int(MaxRSTD)-1))
	}
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (v RSTD) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrRSTDEncode, err)
	}
	if err := w.WriteConstrainedWholeNumber(uint64(v), 0, uint64(MaxRSTD)); err != nil {
		return fmt.Errorf("%w: %w", ErrRSTDEncode, err)
	}
	return nil
}

func DecodeRSTD(r *uper.Reader) (RSTD, error) {
	n, err := r.ReadConstrainedWholeNumber(0, uint64(MaxRSTD))
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrRSTDDecode, err)
	}
	v, err := NewRSTD(uint16(n))
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrRSTDDecode, err)
	}
	return v, nil
}
