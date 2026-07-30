package capability

import (
	"fmt"
	"github.com/vectorcore/esmlc/internal/uper"
)

// RequestCapabilitiesR9IEs has five root optional fields in ASN.1 order:
// common, A-GNSS, OTDOA, ECID, EPDU. Only ECID is representable here.
type RequestCapabilitiesR9IEs struct{ ECID *ECIDRequestCapabilities }
type ECIDRequestCapabilities struct{}
type ProvideCapabilitiesR9IEs struct{ ECID *ECIDProvideCapabilities }
type ECIDProvideCapabilities struct{ MeasurementSupport uper.BitString }

func (v ECIDProvideCapabilities) SupportsRSRP() bool { return bit(v.MeasurementSupport, 0) }
func (v ECIDProvideCapabilities) SupportsRSRQ() bool { return bit(v.MeasurementSupport, 1) }
func (v ECIDProvideCapabilities) SupportsUERxTxTimeDifference() bool {
	return bit(v.MeasurementSupport, 2)
}
func bit(v uper.BitString, n int) bool {
	b := v.Bytes()
	return n < v.BitLen() && b[n/8]&(1<<(7-(n%8))) != 0
}
func (v RequestCapabilitiesR9IEs) Validate() error { return nil }
func (v ProvideCapabilitiesR9IEs) Validate() error {
	if v.ECID == nil {
		return nil
	}
	n := v.ECID.MeasurementSupport.BitLen()
	if n < 1 || n > 8 {
		return fmt.Errorf("%w: length %d", ErrInvalidECID, n)
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
	if e := w.WriteOptionalBitmap([]bool{false, false, false, v.ECID != nil, false}); e != nil {
		return e
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
	if bits[1] {
		return RequestCapabilitiesR9IEs{}, ErrUnsupportedAGNSS
	}
	if bits[2] {
		return RequestCapabilitiesR9IEs{}, ErrUnsupportedOTDOA
	}
	if bits[4] {
		return RequestCapabilitiesR9IEs{}, ErrUnsupportedEPDU
	}
	v := RequestCapabilitiesR9IEs{}
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
	if e := w.WriteOptionalBitmap([]bool{false, false, false, v.ECID != nil, false}); e != nil {
		return e
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
	if bits[1] {
		return ProvideCapabilitiesR9IEs{}, ErrUnsupportedAGNSS
	}
	if bits[2] {
		return ProvideCapabilitiesR9IEs{}, ErrUnsupportedOTDOA
	}
	if bits[4] {
		return ProvideCapabilitiesR9IEs{}, ErrUnsupportedEPDU
	}
	v := ProvideCapabilitiesR9IEs{}
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
