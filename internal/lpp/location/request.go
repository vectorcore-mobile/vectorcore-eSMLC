package location

import (
	"fmt"

	"github.com/vectorcore/esmlc/internal/uper"
)

// RequestLocationInformationR9IEs represents the supported root form of
// RequestLocationInformation-r9-IEs. ECID is the only supported optional
// field; nil represents the normative empty R9 payload.
type RequestLocationInformationR9IEs struct {
	ECID *ECIDRequestLocationInformation
}

// ECIDRequestLocationInformation contains the TS 37.355 named BIT STRING
// requestedMeasurements. Bit zero is RSRP, bit one is RSRQ, and bit two is UE
// Rx-Tx time difference. The original bit length is significant and retained.
type ECIDRequestLocationInformation struct {
	RequestedMeasurements uper.BitString
}

func (v ECIDRequestLocationInformation) RequestsRSRP() bool {
	return requestedBit(v.RequestedMeasurements, 0)
}
func (v ECIDRequestLocationInformation) RequestsRSRQ() bool {
	return requestedBit(v.RequestedMeasurements, 1)
}
func (v ECIDRequestLocationInformation) RequestsUERxTxTimeDifference() bool {
	return requestedBit(v.RequestedMeasurements, 2)
}

func requestedBit(v uper.BitString, n int) bool {
	b := v.Bytes()
	return n >= 0 && n < v.BitLen() && b[n/8]&(1<<(7-(n%8))) != 0
}

func (v RequestLocationInformationR9IEs) Validate() error {
	if v.ECID == nil {
		return nil
	}
	n := v.ECID.RequestedMeasurements.BitLen()
	if n == 0 {
		return ErrMissingRequestedMeasurements
	}
	if n > 8 {
		return fmt.Errorf("%w: requestedMeasurements length %d", ErrInvalidECIDRequest, n)
	}
	return nil
}

// EncodeRequestLocationInformation writes only the supported R9 critical
// extension path: c1[0], requestLocationInformation-r9[0].
func EncodeRequestLocationInformation(w *uper.Writer, v RequestLocationInformationR9IEs) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteRootChoiceIndex(0, 2); err != nil {
		return err
	}
	if err := w.WriteRootChoiceIndex(0, 4); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{false, false, false, v.ECID != nil, false}); err != nil {
		return err
	}
	if v.ECID == nil {
		return nil
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	return w.WriteBitString(v.ECID.RequestedMeasurements, 1, 8)
}

// DecodeRequestLocationInformation decodes the same bounded path and rejects
// every unsupported root field and extension before returning a value.
func DecodeRequestLocationInformation(r *uper.Reader) (RequestLocationInformationR9IEs, error) {
	critical, err := r.ReadRootChoiceIndex(2)
	if err != nil {
		return RequestLocationInformationR9IEs{}, err
	}
	if critical != 0 {
		return RequestLocationInformationR9IEs{}, ErrUnsupportedCriticalExtension
	}
	r9, err := r.ReadRootChoiceIndex(4)
	if err != nil {
		return RequestLocationInformationR9IEs{}, err
	}
	if r9 != 0 {
		return RequestLocationInformationR9IEs{}, ErrUnsupportedCriticalExtension
	}
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return RequestLocationInformationR9IEs{}, err
	}
	if err = uper.RequireNoExtension(ext); err != nil {
		return RequestLocationInformationR9IEs{}, fmt.Errorf("%w: %w", ErrUnsupportedExtension, err)
	}
	bits, err := r.ReadOptionalBitmap(5)
	if err != nil {
		return RequestLocationInformationR9IEs{}, err
	}
	if bits[0] {
		return RequestLocationInformationR9IEs{}, ErrUnsupportedCommon
	}
	if bits[1] {
		return RequestLocationInformationR9IEs{}, ErrUnsupportedAGNSS
	}
	if bits[2] {
		return RequestLocationInformationR9IEs{}, ErrUnsupportedOTDOA
	}
	if bits[4] {
		return RequestLocationInformationR9IEs{}, ErrUnsupportedEPDU
	}
	v := RequestLocationInformationR9IEs{}
	if !bits[3] {
		return v, nil
	}
	ext, err = r.ReadExtensionPresent()
	if err != nil {
		return v, err
	}
	if err = uper.RequireNoExtension(ext); err != nil {
		return v, fmt.Errorf("%w: ECID request: %w", ErrUnsupportedExtension, err)
	}
	measurements, err := r.ReadBitString(1, 8)
	if err != nil {
		return v, fmt.Errorf("%w: %w", ErrInvalidECIDRequest, err)
	}
	v.ECID = &ECIDRequestLocationInformation{RequestedMeasurements: measurements}
	return v, nil
}
