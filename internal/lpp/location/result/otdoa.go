package result

import (
	"errors"
	"fmt"

	"github.com/vectorcore/esmlc/internal/uper"
)

var (
	ErrECGIInvalid               = errors.New("lpp location result: invalid ECGI")
	ErrOTDOAMeasQualityInvalid   = errors.New("lpp location result: invalid OTDOA measurement quality")
	ErrOTDOAMeasQualityExtension = errors.New("lpp location result: OTDOA measurement quality extensions unsupported")
)

// ECGI is the plain, non-extensible TS 37.355 common IE used by OTDOA
// (distinct from the extensible, CHOICE-based CellGlobalIdEUTRAAndUTRA used
// by ECID): MCC, MNC, and a fixed 28-bit E-UTRA cell identity, with no
// CHOICE and no extension container.
type ECGI struct {
	mcc          MCC
	mnc          MNC
	cellIdentity uper.BitString
}

func NewECGI(mcc MCC, mnc MNC, cellIdentity uper.BitString) (ECGI, error) {
	v := ECGI{mcc: mcc, mnc: mnc, cellIdentity: cellIdentity}
	if err := v.Validate(); err != nil {
		return ECGI{}, err
	}
	return v, nil
}
func (v ECGI) MCC() MCC                     { return v.mcc }
func (v ECGI) MNC() MNC                     { return v.mnc }
func (v ECGI) CellIdentity() uper.BitString { return v.cellIdentity }
func (v ECGI) Validate() error {
	if err := v.mcc.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrECGIInvalid, err)
	}
	if err := v.mnc.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrECGIInvalid, err)
	}
	if v.cellIdentity.BitLen() != 28 {
		return fmt.Errorf("%w: cell identity must be 28 bits, got %d", ErrECGIInvalid, v.cellIdentity.BitLen())
	}
	return nil
}
func (v ECGI) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := v.mcc.EncodeUPER(w); err != nil {
		return err
	}
	if err := v.mnc.EncodeUPER(w); err != nil {
		return err
	}
	return w.WriteBitString(v.cellIdentity, 28, 28)
}
func DecodeECGI(r *uper.Reader) (ECGI, error) {
	mcc, err := DecodeMCC(r)
	if err != nil {
		return ECGI{}, err
	}
	mnc, err := DecodeMNC(r)
	if err != nil {
		return ECGI{}, err
	}
	ci, err := r.ReadBitString(28, 28)
	if err != nil {
		return ECGI{}, err
	}
	return NewECGI(mcc, mnc, ci)
}

// OTDOAMeasQuality is TS 37.355 OTDOA-MeasQuality: an extensible SEQUENCE
// with one OPTIONAL root member (error-NumSamples). error-Resolution and
// error-Value together describe an uncertainty magnitude in metres;
// error-NumSamples is a coded sample-count bucket. Both encodings are
// defined directly in the recovered TS 37.355 field-description table (no
// external cross-reference needed, unlike RSTD).
type OTDOAMeasQuality struct {
	resolution uper.BitString // error-Resolution, 2 bits
	value      uper.BitString // error-Value, 5 bits
	numSamples *uper.BitString
}

func NewOTDOAMeasQuality(resolution, value uper.BitString, numSamples *uper.BitString) (OTDOAMeasQuality, error) {
	v := OTDOAMeasQuality{resolution: resolution, value: value}
	if numSamples != nil {
		x := *numSamples
		v.numSamples = &x
	}
	if err := v.Validate(); err != nil {
		return OTDOAMeasQuality{}, err
	}
	return v, nil
}
func (v OTDOAMeasQuality) Resolution() uper.BitString { return v.resolution }
func (v OTDOAMeasQuality) Value() uper.BitString      { return v.value }
func (v OTDOAMeasQuality) NumSamples() (uper.BitString, bool) {
	if v.numSamples == nil {
		return uper.BitString{}, false
	}
	return *v.numSamples, true
}

// ResolutionMetres decodes error-Resolution's two-bit enumeration into the
// metre value TS 37.355 Table "OTDOA-MeasQuality field descriptions" defines:
// 00->5, 01->10, 10->20, 11->30.
func (v OTDOAMeasQuality) ResolutionMetres() float64 {
	b := v.resolution.Bytes()
	code := (b[0] >> 6) & 0x3
	return []float64{5, 10, 20, 30}[code]
}

// UncertaintyMetres decodes error-Value (five bits, 0..31) against
// ResolutionMetres per the same table: code c covers [R*c, R*(c+1)-1] metres,
// c=31 is unbounded ("R*31 metres or more"); this returns the bucket's lower
// bound.
func (v OTDOAMeasQuality) UncertaintyMetres() float64 {
	b := v.value.Bytes()
	code := (b[0] >> 3) & 0x1f
	return float64(code) * v.ResolutionMetres()
}

func (v OTDOAMeasQuality) Validate() error {
	if v.resolution.BitLen() != 2 {
		return fmt.Errorf("%w: error-Resolution must be 2 bits, got %d", ErrOTDOAMeasQualityInvalid, v.resolution.BitLen())
	}
	if v.value.BitLen() != 5 {
		return fmt.Errorf("%w: error-Value must be 5 bits, got %d", ErrOTDOAMeasQualityInvalid, v.value.BitLen())
	}
	if v.numSamples != nil && v.numSamples.BitLen() != 3 {
		return fmt.Errorf("%w: error-NumSamples must be 3 bits, got %d", ErrOTDOAMeasQualityInvalid, v.numSamples.BitLen())
	}
	return nil
}
func (v OTDOAMeasQuality) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{v.numSamples != nil}); err != nil {
		return err
	}
	if err := w.WriteBitString(v.resolution, 2, 2); err != nil {
		return err
	}
	if err := w.WriteBitString(v.value, 5, 5); err != nil {
		return err
	}
	if v.numSamples != nil {
		return w.WriteBitString(*v.numSamples, 3, 3)
	}
	return nil
}
func DecodeOTDOAMeasQuality(r *uper.Reader) (OTDOAMeasQuality, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return OTDOAMeasQuality{}, err
	}
	if ext {
		return OTDOAMeasQuality{}, ErrOTDOAMeasQualityExtension
	}
	present, err := r.ReadOptionalBitmap(1)
	if err != nil {
		return OTDOAMeasQuality{}, err
	}
	resolution, err := r.ReadBitString(2, 2)
	if err != nil {
		return OTDOAMeasQuality{}, err
	}
	value, err := r.ReadBitString(5, 5)
	if err != nil {
		return OTDOAMeasQuality{}, err
	}
	var numSamples *uper.BitString
	if present[0] {
		x, err := r.ReadBitString(3, 3)
		if err != nil {
			return OTDOAMeasQuality{}, err
		}
		numSamples = &x
	}
	return NewOTDOAMeasQuality(resolution, value, numSamples)
}
