package location

import (
	"fmt"

	"github.com/vectorcore/esmlc/internal/lpp/location/result"
	"github.com/vectorcore/esmlc/internal/uper"
)

// ProvideLocationInformationR9IEs is the bounded Release 9 provide-location
// payload. ECID is the only supported method branch; the common, A-GNSS,
// OTDOA and EPDU branches fail closed.
type ProvideLocationInformationR9IEs struct {
	ECID *ECIDProvideLocationInformation
}

// ECIDProvideLocationInformation supports measurement information only.
// ECID-Error and all extension additions remain deliberately unsupported.
type ECIDProvideLocationInformation struct {
	signal *ECIDSignalMeasurementInformation
}

// ECIDSignalMeasurementInformation owns one optional primary-cell result and
// a mandatory bounded list of one through 32 measured results.
type ECIDSignalMeasurementInformation struct {
	primary *result.MeasuredResultsElement
	results []result.MeasuredResultsElement
}

func NewECIDSignalMeasurementInformation(primary *result.MeasuredResultsElement, results []result.MeasuredResultsElement) (ECIDSignalMeasurementInformation, error) {
	if len(results) < 1 || len(results) > 32 {
		return ECIDSignalMeasurementInformation{}, fmt.Errorf("%w: count %d outside 1..32", ErrMissingMeasuredResults, len(results))
	}
	v := ECIDSignalMeasurementInformation{results: append([]result.MeasuredResultsElement(nil), results...)}
	if primary != nil {
		x := *primary
		v.primary = &x
	}
	if err := v.Validate(); err != nil {
		return ECIDSignalMeasurementInformation{}, err
	}
	return v, nil
}

func NewECIDProvideLocationInformation(signal ECIDSignalMeasurementInformation) (ECIDProvideLocationInformation, error) {
	if err := signal.Validate(); err != nil {
		return ECIDProvideLocationInformation{}, err
	}
	copy := signal.clone()
	return ECIDProvideLocationInformation{signal: &copy}, nil
}

func (v ECIDSignalMeasurementInformation) PrimaryCellMeasuredResults() (result.MeasuredResultsElement, bool) {
	if v.primary == nil {
		return result.MeasuredResultsElement{}, false
	}
	return *v.primary, true
}

func (v ECIDSignalMeasurementInformation) MeasuredResults() []result.MeasuredResultsElement {
	return append([]result.MeasuredResultsElement(nil), v.results...)
}

func (v ECIDSignalMeasurementInformation) Validate() error {
	if len(v.results) < 1 || len(v.results) > 32 {
		return fmt.Errorf("%w: count %d outside 1..32", ErrMissingMeasuredResults, len(v.results))
	}
	if v.primary != nil {
		if err := v.primary.Validate(); err != nil {
			return fmt.Errorf("%w: primary cell: %w", ErrInvalidECIDProvide, err)
		}
	}
	for i := range v.results {
		if err := v.results[i].Validate(); err != nil {
			return fmt.Errorf("%w: measured result %d: %w", ErrInvalidECIDProvide, i, err)
		}
	}
	return nil
}

func (v ECIDProvideLocationInformation) SignalMeasurementInformation() (ECIDSignalMeasurementInformation, bool) {
	if v.signal == nil {
		return ECIDSignalMeasurementInformation{}, false
	}
	return v.signal.clone(), true
}

func (v ECIDProvideLocationInformation) Validate() error {
	if v.signal != nil {
		return v.signal.Validate()
	}
	return nil
}

func (v ProvideLocationInformationR9IEs) Validate() error {
	if v.ECID != nil {
		return v.ECID.Validate()
	}
	return nil
}

func (v ECIDSignalMeasurementInformation) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{v.primary != nil}); err != nil {
		return err
	}
	if v.primary != nil {
		if err := v.primary.EncodeUPER(w); err != nil {
			return fmt.Errorf("%w: primary cell: %w", ErrInvalidECIDProvide, err)
		}
	}
	return w.WriteSequenceOf(len(v.results), 1, 32, func(index int, writer *uper.Writer) error {
		if err := v.results[index].EncodeUPER(writer); err != nil {
			return fmt.Errorf("%w: measured result %d: %w", ErrInvalidECIDProvide, index, err)
		}
		return nil
	})
}

func DecodeECIDSignalMeasurementInformation(r *uper.Reader) (ECIDSignalMeasurementInformation, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return ECIDSignalMeasurementInformation{}, fmt.Errorf("%w: signal extension: %w", ErrInvalidECIDProvide, err)
	}
	if ext {
		return ECIDSignalMeasurementInformation{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(1)
	if err != nil {
		return ECIDSignalMeasurementInformation{}, fmt.Errorf("%w: signal optional bitmap: %w", ErrInvalidECIDProvide, err)
	}
	var primary *result.MeasuredResultsElement
	if present[0] {
		x, err := result.DecodeMeasuredResultsElement(r)
		if err != nil {
			return ECIDSignalMeasurementInformation{}, fmt.Errorf("%w: primary cell: %w", ErrInvalidECIDProvide, err)
		}
		primary = &x
	}
	values := make([]result.MeasuredResultsElement, 0, 32)
	_, err = r.ReadSequenceOf(1, 32, func(index int, reader *uper.Reader) error {
		x, err := result.DecodeMeasuredResultsElement(reader)
		if err != nil {
			return fmt.Errorf("%w: measured result %d: %w", ErrInvalidECIDProvide, index, err)
		}
		values = append(values, x)
		return nil
	})
	if err != nil {
		return ECIDSignalMeasurementInformation{}, fmt.Errorf("%w: measured results: %w", ErrInvalidECIDProvide, err)
	}
	return NewECIDSignalMeasurementInformation(primary, values)
}

func (v ECIDProvideLocationInformation) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{v.signal != nil, false}); err != nil {
		return err
	}
	if v.signal != nil {
		return v.signal.EncodeUPER(w)
	}
	return nil
}

func DecodeECIDProvideLocationInformation(r *uper.Reader) (ECIDProvideLocationInformation, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return ECIDProvideLocationInformation{}, fmt.Errorf("%w: ECID extension: %w", ErrInvalidECIDProvide, err)
	}
	if ext {
		return ECIDProvideLocationInformation{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(2)
	if err != nil {
		return ECIDProvideLocationInformation{}, fmt.Errorf("%w: ECID optional bitmap: %w", ErrInvalidECIDProvide, err)
	}
	if present[1] {
		return ECIDProvideLocationInformation{}, ErrUnsupportedECIDError
	}
	if !present[0] {
		return ECIDProvideLocationInformation{}, nil
	}
	signal, err := DecodeECIDSignalMeasurementInformation(r)
	if err != nil {
		return ECIDProvideLocationInformation{}, err
	}
	return NewECIDProvideLocationInformation(signal)
}

func EncodeProvideLocationInformation(w *uper.Writer, v ProvideLocationInformationR9IEs) error {
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
	if v.ECID != nil {
		return v.ECID.EncodeUPER(w)
	}
	return nil
}

func DecodeProvideLocationInformation(r *uper.Reader) (ProvideLocationInformationR9IEs, error) {
	critical, err := r.ReadRootChoiceIndex(2)
	if err != nil {
		return ProvideLocationInformationR9IEs{}, err
	}
	if critical != 0 {
		return ProvideLocationInformationR9IEs{}, ErrUnsupportedCriticalExtension
	}
	release, err := r.ReadRootChoiceIndex(4)
	if err != nil {
		return ProvideLocationInformationR9IEs{}, err
	}
	if release != 0 {
		return ProvideLocationInformationR9IEs{}, ErrUnsupportedCriticalExtension
	}
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return ProvideLocationInformationR9IEs{}, err
	}
	if ext {
		return ProvideLocationInformationR9IEs{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(5)
	if err != nil {
		return ProvideLocationInformationR9IEs{}, err
	}
	if present[0] {
		return ProvideLocationInformationR9IEs{}, ErrUnsupportedCommon
	}
	if present[1] {
		return ProvideLocationInformationR9IEs{}, ErrUnsupportedAGNSS
	}
	if present[2] {
		return ProvideLocationInformationR9IEs{}, ErrUnsupportedOTDOA
	}
	if present[4] {
		return ProvideLocationInformationR9IEs{}, ErrUnsupportedEPDU
	}
	v := ProvideLocationInformationR9IEs{}
	if present[3] {
		x, err := DecodeECIDProvideLocationInformation(r)
		if err != nil {
			return v, err
		}
		v.ECID = &x
	}
	return v, nil
}

func (v ECIDSignalMeasurementInformation) clone() ECIDSignalMeasurementInformation {
	out := ECIDSignalMeasurementInformation{results: append([]result.MeasuredResultsElement(nil), v.results...)}
	if v.primary != nil {
		x := *v.primary
		out.primary = &x
	}
	return out
}

// Clone returns independent container state. Result values are immutable
// values, so copying their elements is sufficient.
func (v ProvideLocationInformationR9IEs) Clone() ProvideLocationInformationR9IEs {
	out := v
	if v.ECID != nil {
		x := *v.ECID
		if v.ECID.signal != nil {
			s := v.ECID.signal.clone()
			x.signal = &s
		}
		out.ECID = &x
	}
	return out
}
