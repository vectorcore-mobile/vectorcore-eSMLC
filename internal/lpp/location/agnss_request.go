package location

import (
	"errors"
	"fmt"

	"github.com/vectorcore/esmlc/internal/uper"
)

var ErrInvalidAGNSSRequest = errors.New("lpp location: invalid A-GNSS request")

// AGNSSRequestLocationInformation is the bounded TS 37.355 root of
// A-GNSS-RequestLocationInformation: gnss-PositioningInstructions'
// gnss-Methods bitmap and its four mandatory booleans. The Release 15
// ha-GNSS-Req extension addition is not representable here.
type AGNSSRequestLocationInformation struct {
	// GNSSMethods is the GNSS-ID-Bitmap requested measurement set. Only bit 0
	// (GPS) is meaningful to this implementation; see
	// docs/lpp-spec-audit.md's A-GNSS phase section for scope.
	GNSSMethods               uper.BitString
	FineTimeAssistanceMeasReq bool
	ADRMeasReq                bool
	MultiFreqMeasReq          bool
	// AssistanceAvailability indicates whether the target device may request
	// additional GNSS assistance data from the server. This implementation
	// has no assistance-data source, so callers should not set this true.
	AssistanceAvailability bool
}

func (v AGNSSRequestLocationInformation) Validate() error {
	n := v.GNSSMethods.BitLen()
	if n < 1 || n > 16 {
		return fmt.Errorf("%w: gnss-Methods length %d", ErrInvalidAGNSSRequest, n)
	}
	return nil
}

func writeGNSSIDBitmap(w *uper.Writer, v uper.BitString) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	return w.WriteBitString(v, 1, 16)
}
func readGNSSIDBitmap(r *uper.Reader) (uper.BitString, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return uper.BitString{}, err
	}
	if ext {
		return uper.BitString{}, ErrUnsupportedExtension
	}
	return r.ReadBitString(1, 16)
}

func (v AGNSSRequestLocationInformation) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	// A-GNSS-RequestLocationInformation: extensible SEQUENCE, no OPTIONAL
	// root members (gnss-PositioningInstructions is mandatory).
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	// GNSS-PositioningInstructions: extensible SEQUENCE, no OPTIONAL root
	// members either (the r15 addition lives in an extension group).
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := writeGNSSIDBitmap(w, v.GNSSMethods); err != nil {
		return err
	}
	if err := w.WriteBoolean(v.FineTimeAssistanceMeasReq); err != nil {
		return err
	}
	if err := w.WriteBoolean(v.ADRMeasReq); err != nil {
		return err
	}
	if err := w.WriteBoolean(v.MultiFreqMeasReq); err != nil {
		return err
	}
	return w.WriteBoolean(v.AssistanceAvailability)
}

func DecodeAGNSSRequestLocationInformation(r *uper.Reader) (AGNSSRequestLocationInformation, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return AGNSSRequestLocationInformation{}, err
	}
	if ext {
		return AGNSSRequestLocationInformation{}, ErrUnsupportedExtension
	}
	instrExt, err := r.ReadExtensionPresent()
	if err != nil {
		return AGNSSRequestLocationInformation{}, err
	}
	if instrExt {
		return AGNSSRequestLocationInformation{}, ErrUnsupportedExtension
	}
	methods, err := readGNSSIDBitmap(r)
	if err != nil {
		return AGNSSRequestLocationInformation{}, err
	}
	fineTime, err := r.ReadBoolean()
	if err != nil {
		return AGNSSRequestLocationInformation{}, err
	}
	adr, err := r.ReadBoolean()
	if err != nil {
		return AGNSSRequestLocationInformation{}, err
	}
	multiFreq, err := r.ReadBoolean()
	if err != nil {
		return AGNSSRequestLocationInformation{}, err
	}
	assistance, err := r.ReadBoolean()
	if err != nil {
		return AGNSSRequestLocationInformation{}, err
	}
	v := AGNSSRequestLocationInformation{GNSSMethods: methods, FineTimeAssistanceMeasReq: fineTime, ADRMeasReq: adr, MultiFreqMeasReq: multiFreq, AssistanceAvailability: assistance}
	return v, v.Validate()
}
