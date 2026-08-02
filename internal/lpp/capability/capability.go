package capability

import (
	"fmt"
	"github.com/vectorcore/esmlc/internal/uper"
)

// RequestCapabilitiesR9IEs has five root optional fields in ASN.1 order:
// common, A-GNSS, OTDOA, ECID, EPDU. Only A-GNSS, OTDOA, and ECID are
// representable here.
type RequestCapabilitiesR9IEs struct {
	AGNSS *AGNSSRequestCapabilities
	OTDOA *OTDOARequestCapabilities
	ECID  *ECIDRequestCapabilities
}
type ECIDRequestCapabilities struct{}

// OTDOARequestCapabilities is TS 37.355 OTDOA-RequestCapabilities, an
// extensible SEQUENCE with no root fields at all.
type OTDOARequestCapabilities struct{}

type ProvideCapabilitiesR9IEs struct {
	AGNSS *AGNSSProvideCapabilities
	OTDOA *OTDOAProvideCapabilities
	ECID  *ECIDProvideCapabilities
}
type ECIDProvideCapabilities struct{ MeasurementSupport uper.BitString }

// OTDOAProvideCapabilities is the bounded TS 37.355 OTDOA-ProvideCapabilities
// root: only otdoa-Mode BIT STRING(SIZE(1..8)). Every Release 10/14/15
// optional root field after it (PRS config, band lists, antenna count, ...)
// is deliberately not represented.
type OTDOAProvideCapabilities struct{ Mode uper.BitString }

func (v ECIDProvideCapabilities) SupportsRSRP() bool { return bit(v.MeasurementSupport, 0) }
func (v ECIDProvideCapabilities) SupportsRSRQ() bool { return bit(v.MeasurementSupport, 1) }
func (v ECIDProvideCapabilities) SupportsUERxTxTimeDifference() bool {
	return bit(v.MeasurementSupport, 2)
}

// SupportsUEAssisted reports otdoa-Mode bit 0: UE-assisted OTDOA with LTE PRS.
func (v OTDOAProvideCapabilities) SupportsUEAssisted() bool { return bit(v.Mode, 0) }

func bit(v uper.BitString, n int) bool {
	b := v.Bytes()
	return n < v.BitLen() && b[n/8]&(1<<(7-(n%8))) != 0
}
func (v RequestCapabilitiesR9IEs) Validate() error {
	if v.AGNSS != nil {
		return v.AGNSS.Validate()
	}
	return nil
}
func (v ProvideCapabilitiesR9IEs) Validate() error {
	if v.AGNSS != nil {
		if err := v.AGNSS.Validate(); err != nil {
			return err
		}
	}
	if v.OTDOA != nil {
		n := v.OTDOA.Mode.BitLen()
		if n < 1 || n > 8 {
			return fmt.Errorf("%w: length %d", ErrInvalidOTDOA, n)
		}
	}
	if v.ECID != nil {
		n := v.ECID.MeasurementSupport.BitLen()
		if n < 1 || n > 8 {
			return fmt.Errorf("%w: length %d", ErrInvalidECID, n)
		}
	}
	return nil
}

// EncodeRequestCapabilities writes RequestCapabilities. The critical extension
// path is c1[0] then requestCapabilities-r9[0].
func EncodeRequestCapabilities(w *uper.Writer, v RequestCapabilitiesR9IEs) error {
	if e := v.Validate(); e != nil {
		return e
	}
	if e := w.WriteRootChoiceIndex(0, 2); e != nil {
		return e
	}
	if e := w.WriteRootChoiceIndex(0, 4); e != nil {
		return e
	}
	return encodeRequestR9(w, v)
}
func DecodeRequestCapabilities(r *uper.Reader) (RequestCapabilitiesR9IEs, error) {
	a, e := r.ReadRootChoiceIndex(2)
	if e != nil {
		return RequestCapabilitiesR9IEs{}, e
	}
	if a != 0 {
		return RequestCapabilitiesR9IEs{}, ErrUnsupportedCriticalExtension
	}
	b, e := r.ReadRootChoiceIndex(4)
	if e != nil {
		return RequestCapabilitiesR9IEs{}, e
	}
	if b != 0 {
		return RequestCapabilitiesR9IEs{}, ErrUnsupportedCriticalExtension
	}
	return decodeRequestR9(r)
}
func encodeRequestR9(w *uper.Writer, v RequestCapabilitiesR9IEs) error {
	if e := w.WriteExtensionPresent(false); e != nil {
		return e
	}
	if e := w.WriteOptionalBitmap([]bool{false, v.AGNSS != nil, v.OTDOA != nil, v.ECID != nil, false}); e != nil {
		return e
	}
	if v.AGNSS != nil {
		if e := v.AGNSS.EncodeUPER(w); e != nil {
			return e
		}
	}
	if v.OTDOA != nil {
		if e := w.WriteExtensionPresent(false); e != nil {
			return e
		}
	}
	if v.ECID != nil {
		return w.WriteExtensionPresent(false)
	}
	return nil
}
func decodeRequestR9(r *uper.Reader) (RequestCapabilitiesR9IEs, error) {
	ext, e := r.ReadExtensionPresent()
	if e != nil {
		return RequestCapabilitiesR9IEs{}, e
	}
	if e = uper.RequireNoExtension(ext); e != nil {
		return RequestCapabilitiesR9IEs{}, fmt.Errorf("%w: %w", ErrUnsupportedExtension, e)
	}
	bits, e := r.ReadOptionalBitmap(5)
	if e != nil {
		return RequestCapabilitiesR9IEs{}, e
	}
	if bits[0] {
		return RequestCapabilitiesR9IEs{}, ErrUnsupportedCommon
	}
	if bits[4] {
		return RequestCapabilitiesR9IEs{}, ErrUnsupportedEPDU
	}
	v := RequestCapabilitiesR9IEs{}
	if bits[1] {
		x, e := DecodeAGNSSRequestCapabilities(r)
		if e != nil {
			return v, e
		}
		v.AGNSS = &x
	}
	if bits[2] {
		ext, e = r.ReadExtensionPresent()
		if e != nil {
			return v, e
		}
		if e = uper.RequireNoExtension(ext); e != nil {
			return v, fmt.Errorf("%w: OTDOA request: %w", ErrUnsupportedExtension, e)
		}
		v.OTDOA = &OTDOARequestCapabilities{}
	}
	if bits[3] {
		ext, e = r.ReadExtensionPresent()
		if e != nil {
			return v, e
		}
		if e = uper.RequireNoExtension(ext); e != nil {
			return v, fmt.Errorf("%w: ECID request: %w", ErrUnsupportedExtension, e)
		}
		v.ECID = &ECIDRequestCapabilities{}
	}
	return v, nil
}

// EncodeProvideCapabilities writes ProvideCapabilities using c1[0] then
// provideCapabilities-r9[0].
func EncodeProvideCapabilities(w *uper.Writer, v ProvideCapabilitiesR9IEs) error {
	if e := v.Validate(); e != nil {
		return e
	}
	if e := w.WriteRootChoiceIndex(0, 2); e != nil {
		return e
	}
	if e := w.WriteRootChoiceIndex(0, 4); e != nil {
		return e
	}
	return encodeProvideR9(w, v)
}
func DecodeProvideCapabilities(r *uper.Reader) (ProvideCapabilitiesR9IEs, error) {
	a, e := r.ReadRootChoiceIndex(2)
	if e != nil {
		return ProvideCapabilitiesR9IEs{}, e
	}
	if a != 0 {
		return ProvideCapabilitiesR9IEs{}, ErrUnsupportedCriticalExtension
	}
	b, e := r.ReadRootChoiceIndex(4)
	if e != nil {
		return ProvideCapabilitiesR9IEs{}, e
	}
	if b != 0 {
		return ProvideCapabilitiesR9IEs{}, ErrUnsupportedCriticalExtension
	}
	return decodeProvideR9(r)
}
func encodeProvideR9(w *uper.Writer, v ProvideCapabilitiesR9IEs) error {
	if e := w.WriteExtensionPresent(false); e != nil {
		return e
	}
	if e := w.WriteOptionalBitmap([]bool{false, v.AGNSS != nil, v.OTDOA != nil, v.ECID != nil, false}); e != nil {
		return e
	}
	if v.AGNSS != nil {
		if e := v.AGNSS.EncodeUPER(w); e != nil {
			return e
		}
	}
	if v.OTDOA != nil {
		if e := w.WriteExtensionPresent(false); e != nil {
			return e
		}
		if e := w.WriteBitString(v.OTDOA.Mode, 1, 8); e != nil {
			return e
		}
	}
	if v.ECID != nil {
		if e := w.WriteExtensionPresent(false); e != nil {
			return e
		}
		return w.WriteBitString(v.ECID.MeasurementSupport, 1, 8)
	}
	return nil
}
func decodeProvideR9(r *uper.Reader) (ProvideCapabilitiesR9IEs, error) {
	ext, e := r.ReadExtensionPresent()
	if e != nil {
		return ProvideCapabilitiesR9IEs{}, e
	}
	if e = uper.RequireNoExtension(ext); e != nil {
		return ProvideCapabilitiesR9IEs{}, fmt.Errorf("%w: %w", ErrUnsupportedExtension, e)
	}
	bits, e := r.ReadOptionalBitmap(5)
	if e != nil {
		return ProvideCapabilitiesR9IEs{}, e
	}
	if bits[0] {
		return ProvideCapabilitiesR9IEs{}, ErrUnsupportedCommon
	}
	if bits[4] {
		return ProvideCapabilitiesR9IEs{}, ErrUnsupportedEPDU
	}
	v := ProvideCapabilitiesR9IEs{}
	if bits[1] {
		x, e := DecodeAGNSSProvideCapabilities(r)
		if e != nil {
			return v, e
		}
		v.AGNSS = &x
	}
	if bits[2] {
		ext, e = r.ReadExtensionPresent()
		if e != nil {
			return v, e
		}
		if e = uper.RequireNoExtension(ext); e != nil {
			return v, fmt.Errorf("%w: OTDOA provide: %w", ErrUnsupportedExtension, e)
		}
		s, e := r.ReadBitString(1, 8)
		if e != nil {
			return v, fmt.Errorf("%w: %w", ErrInvalidOTDOA, e)
		}
		v.OTDOA = &OTDOAProvideCapabilities{Mode: s}
	}
	if bits[3] {
		ext, e = r.ReadExtensionPresent()
		if e != nil {
			return v, e
		}
		if e = uper.RequireNoExtension(ext); e != nil {
			return v, fmt.Errorf("%w: ECID provide: %w", ErrUnsupportedExtension, e)
		}
		s, e := r.ReadBitString(1, 8)
		if e != nil {
			return v, fmt.Errorf("%w: %w", ErrInvalidECID, e)
		}
		v.ECID = &ECIDProvideCapabilities{MeasurementSupport: s}
	}
	return v, nil
}
