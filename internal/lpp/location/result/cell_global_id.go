package result

import (
	"fmt"
	"github.com/vectorcore/esmlc/internal/uper"
)

var (
	ErrPLMNDigitOutOfRange               = fmt.Errorf("lpp result: PLMN digit out of range")
	ErrMNCInvalidLength                  = fmt.Errorf("lpp result: MNC invalid length")
	ErrCellIdentityUnset                 = fmt.Errorf("lpp result: cell identity unset")
	ErrCellGlobalIDExtensionsUnsupported = fmt.Errorf("lpp result: cell global ID extensions unsupported")
)

type PLMNDigit uint8

func digit(v uint8) (PLMNDigit, error) {
	if v > 9 {
		return 0, ErrPLMNDigitOutOfRange
	}
	return PLMNDigit(v), nil
}
func (d PLMNDigit) enc(w *uper.Writer) error {
	if d > 9 {
		return ErrPLMNDigitOutOfRange
	}
	return w.WriteConstrainedWholeNumber(uint64(d), 0, 9)
}
func decDigit(r *uper.Reader) (PLMNDigit, error) {
	v, e := r.ReadConstrainedWholeNumber(0, 9)
	if e != nil {
		return 0, e
	}
	return PLMNDigit(v), nil
}

type MCC struct{ d [3]PLMNDigit }

func NewMCC(a, b, c uint8) (MCC, error) {
	x, e := digit(a)
	if e != nil {
		return MCC{}, e
	}
	y, e := digit(b)
	if e != nil {
		return MCC{}, e
	}
	z, e := digit(c)
	return MCC{[3]PLMNDigit{x, y, z}}, e
}
func (m MCC) String() string { return fmt.Sprintf("%d%d%d", m.d[0], m.d[1], m.d[2]) }

type MNC struct {
	d [3]PLMNDigit
	n uint8
}

func NewMNC2(a, b uint8) (MNC, error) {
	x, e := digit(a)
	if e != nil {
		return MNC{}, e
	}
	y, e := digit(b)
	return MNC{d: [3]PLMNDigit{x, y}, n: 2}, e
}
func NewMNC3(a, b, c uint8) (MNC, error) {
	x, e := digit(a)
	if e != nil {
		return MNC{}, e
	}
	y, e := digit(b)
	if e != nil {
		return MNC{}, e
	}
	z, e := digit(c)
	return MNC{d: [3]PLMNDigit{x, y, z}, n: 3}, e
}
func (m MNC) Length() int { return int(m.n) }
func (m MNC) String() string {
	if m.n < 2 || m.n > 3 {
		return ""
	}
	s := ""
	for i := 0; i < int(m.n); i++ {
		s += fmt.Sprint(m.d[i])
	}
	return s
}

type PLMNIdentity struct {
	mcc MCC
	mnc MNC
}

func NewPLMNIdentity(a MCC, b MNC) (PLMNIdentity, error) {
	if b.n < 2 || b.n > 3 {
		return PLMNIdentity{}, ErrMNCInvalidLength
	}
	return PLMNIdentity{a, b}, nil
}

type EUTRACellIdentity struct{ b uper.BitString }
type UTRACellIdentity struct{ b uper.BitString }

func eutra(v uint32) (EUTRACellIdentity, error) {
	if v > 0xfffffff {
		return EUTRACellIdentity{}, fmt.Errorf("EUTRA out of range")
	}
	b, e := uper.NewBitString([]byte{byte(v >> 20), byte(v >> 12), byte(v >> 4), byte(v << 4)}, 28)
	return EUTRACellIdentity{b}, e
}
func utra(v uint32) (UTRACellIdentity, error) {
	b, e := uper.NewBitString([]byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}, 32)
	return UTRACellIdentity{b}, e
}

type CellIdentity struct {
	k uint8
	e EUTRACellIdentity
	u UTRACellIdentity
}

func NewEUTRACellIdentityChoice(v EUTRACellIdentity) CellIdentity { return CellIdentity{k: 1, e: v} }
func NewUTRACellIdentityChoice(v UTRACellIdentity) CellIdentity   { return CellIdentity{k: 2, u: v} }

type CellGlobalIdEUTRAAndUTRA struct {
	p PLMNIdentity
	c CellIdentity
}

func NewCellGlobalIdEUTRAAndUTRA(p PLMNIdentity, c CellIdentity) (CellGlobalIdEUTRAAndUTRA, error) {
	if c.k < 1 || c.k > 2 {
		return CellGlobalIdEUTRAAndUTRA{}, ErrCellIdentityUnset
	}
	return CellGlobalIdEUTRAAndUTRA{p, c}, nil
}
func (v CellGlobalIdEUTRAAndUTRA) EncodeUPER(w *uper.Writer) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	for _, d := range v.p.mcc.d {
		if err := d.enc(w); err != nil {
			return err
		}
	}
	if v.p.mnc.n < 2 || v.p.mnc.n > 3 {
		return ErrMNCInvalidLength
	}
	if err := w.WriteConstrainedWholeNumber(uint64(v.p.mnc.n), 2, 3); err != nil {
		return err
	}
	for i := 0; i < int(v.p.mnc.n); i++ {
		if err := v.p.mnc.d[i].enc(w); err != nil {
			return err
		}
	}
	if v.c.k == 1 {
		if err := w.WriteRootChoiceIndex(0, 2); err != nil {
			return err
		}
		return w.WriteBitString(v.c.e.b, 28, 28)
	}
	if v.c.k == 2 {
		if err := w.WriteRootChoiceIndex(1, 2); err != nil {
			return err
		}
		return w.WriteBitString(v.c.u.b, 32, 32)
	}
	return ErrCellIdentityUnset
}

func NewPLMNDigit(v uint8) (PLMNDigit, error) { return digit(v) }
func (d PLMNDigit) Validate() error {
	if d > 9 {
		return ErrPLMNDigitOutOfRange
	}
	return nil
}
func (d PLMNDigit) EncodeUPER(w *uper.Writer) error     { return d.enc(w) }
func DecodePLMNDigit(r *uper.Reader) (PLMNDigit, error) { return decDigit(r) }
func (m MCC) Digits() [3]PLMNDigit                      { return m.d }
func (m MCC) Validate() error {
	for _, d := range m.d {
		if err := d.Validate(); err != nil {
			return err
		}
	}
	return nil
}
func (m MNC) Digits() []PLMNDigit { return append([]PLMNDigit(nil), m.d[:m.n]...) }
func (m MNC) Validate() error {
	if m.n < 2 || m.n > 3 {
		return ErrMNCInvalidLength
	}
	for i := 0; i < int(m.n); i++ {
		if err := m.d[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}
func (p PLMNIdentity) MCC() MCC { return p.mcc }
func (p PLMNIdentity) MNC() MNC { return p.mnc }
func (p PLMNIdentity) Validate() error {
	if e := p.mcc.Validate(); e != nil {
		return e
	}
	return p.mnc.Validate()
}
func NewEUTRACellIdentityFromUint32(v uint32) (EUTRACellIdentity, error) { return eutra(v) }
func NewUTRACellIdentityFromUint32(v uint32) (UTRACellIdentity, error)   { return utra(v) }
func (v EUTRACellIdentity) BitString() uper.BitString                    { return v.b }
func (v UTRACellIdentity) BitString() uper.BitString                     { return v.b }
func (c CellIdentity) Kind() uint8                                       { return c.k }
func (c CellIdentity) EUTRA() (EUTRACellIdentity, bool)                  { return c.e, c.k == 1 }
func (c CellIdentity) UTRA() (UTRACellIdentity, bool)                    { return c.u, c.k == 2 }
func (v CellGlobalIdEUTRAAndUTRA) PLMNIdentity() PLMNIdentity            { return v.p }
func (v CellGlobalIdEUTRAAndUTRA) CellIdentity() CellIdentity            { return v.c }
func DecodeCellGlobalIdEUTRAAndUTRA(r *uper.Reader) (CellGlobalIdEUTRAAndUTRA, error) {
	ext, e := r.ReadExtensionPresent()
	if e != nil {
		return CellGlobalIdEUTRAAndUTRA{}, e
	}
	if ext {
		return CellGlobalIdEUTRAAndUTRA{}, ErrCellGlobalIDExtensionsUnsupported
	}
	var ds [3]PLMNDigit
	for i := range ds {
		ds[i], e = decDigit(r)
		if e != nil {
			return CellGlobalIdEUTRAAndUTRA{}, e
		}
	}
	mcc := MCC{ds}
	n, e := r.ReadConstrainedWholeNumber(2, 3)
	if e != nil {
		return CellGlobalIdEUTRAAndUTRA{}, e
	}
	var md [3]PLMNDigit
	for i := 0; i < int(n); i++ {
		md[i], e = decDigit(r)
		if e != nil {
			return CellGlobalIdEUTRAAndUTRA{}, e
		}
	}
	p := PLMNIdentity{mcc, MNC{md, uint8(n)}}
	idx, e := r.ReadRootChoiceIndex(2)
	if e != nil {
		return CellGlobalIdEUTRAAndUTRA{}, e
	}
	if idx == 0 {
		b, e := r.ReadBitString(28, 28)
		if e != nil {
			return CellGlobalIdEUTRAAndUTRA{}, e
		}
		return CellGlobalIdEUTRAAndUTRA{p, NewEUTRACellIdentityChoice(EUTRACellIdentity{b})}, nil
	}
	b, e := r.ReadBitString(32, 32)
	if e != nil {
		return CellGlobalIdEUTRAAndUTRA{}, e
	}
	return CellGlobalIdEUTRAAndUTRA{p, NewUTRACellIdentityChoice(UTRACellIdentity{b})}, nil
}

func (m MCC) EncodeUPER(w *uper.Writer) error {
	if err := m.Validate(); err != nil {
		return err
	}
	for _, d := range m.d {
		if err := d.enc(w); err != nil {
			return err
		}
	}
	return nil
}
func DecodeMCC(r *uper.Reader) (MCC, error) {
	var d [3]PLMNDigit
	for i := range d {
		v, e := decDigit(r)
		if e != nil {
			return MCC{}, e
		}
		d[i] = v
	}
	return MCC{d}, nil
}
func (m MNC) EncodeUPER(w *uper.Writer) error {
	if e := m.Validate(); e != nil {
		return e
	}
	if e := w.WriteConstrainedWholeNumber(uint64(m.n), 2, 3); e != nil {
		return e
	}
	for i := 0; i < int(m.n); i++ {
		if e := m.d[i].enc(w); e != nil {
			return e
		}
	}
	return nil
}
func DecodeMNC(r *uper.Reader) (MNC, error) {
	n, e := r.ReadConstrainedWholeNumber(2, 3)
	if e != nil {
		return MNC{}, e
	}
	var d [3]PLMNDigit
	for i := 0; i < int(n); i++ {
		d[i], e = decDigit(r)
		if e != nil {
			return MNC{}, e
		}
	}
	return MNC{d, uint8(n)}, nil
}
func (p PLMNIdentity) EncodeUPER(w *uper.Writer) error {
	if e := p.Validate(); e != nil {
		return e
	}
	if e := p.mcc.EncodeUPER(w); e != nil {
		return e
	}
	return p.mnc.EncodeUPER(w)
}
func DecodePLMNIdentity(r *uper.Reader) (PLMNIdentity, error) {
	m, e := DecodeMCC(r)
	if e != nil {
		return PLMNIdentity{}, e
	}
	n, e := DecodeMNC(r)
	if e != nil {
		return PLMNIdentity{}, e
	}
	return PLMNIdentity{m, n}, nil
}
func (v EUTRACellIdentity) Validate() error {
	if v.b.BitLen() != 28 {
		return fmt.Errorf("EUTRA identity requires 28 bits")
	}
	return nil
}
func (v UTRACellIdentity) Validate() error {
	if v.b.BitLen() != 32 {
		return fmt.Errorf("UTRA identity requires 32 bits")
	}
	return nil
}
func (v EUTRACellIdentity) EncodeUPER(w *uper.Writer) error {
	if e := v.Validate(); e != nil {
		return e
	}
	return w.WriteBitString(v.b, 28, 28)
}
func (v UTRACellIdentity) EncodeUPER(w *uper.Writer) error {
	if e := v.Validate(); e != nil {
		return e
	}
	return w.WriteBitString(v.b, 32, 32)
}
func DecodeEUTRACellIdentity(r *uper.Reader) (EUTRACellIdentity, error) {
	b, e := r.ReadBitString(28, 28)
	return EUTRACellIdentity{b}, e
}
func DecodeUTRACellIdentity(r *uper.Reader) (UTRACellIdentity, error) {
	b, e := r.ReadBitString(32, 32)
	return UTRACellIdentity{b}, e
}
func (v EUTRACellIdentity) Uint32() uint32 {
	b := v.b.Bytes()
	if len(b) != 4 {
		return 0
	}
	return uint32(b[0])<<20 | uint32(b[1])<<12 | uint32(b[2])<<4 | uint32(b[3])>>4
}
func (v UTRACellIdentity) Uint32() uint32 {
	b := v.b.Bytes()
	if len(b) != 4 {
		return 0
	}
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}
func (c CellIdentity) Validate() error {
	if c.k == 1 {
		return c.e.Validate()
	}
	if c.k == 2 {
		return c.u.Validate()
	}
	return ErrCellIdentityUnset
}
func (c CellIdentity) EncodeUPER(w *uper.Writer) error {
	if e := c.Validate(); e != nil {
		return e
	}
	if c.k == 1 {
		if e := w.WriteRootChoiceIndex(0, 2); e != nil {
			return e
		}
		return c.e.EncodeUPER(w)
	}
	if e := w.WriteRootChoiceIndex(1, 2); e != nil {
		return e
	}
	return c.u.EncodeUPER(w)
}
func DecodeCellIdentity(r *uper.Reader) (CellIdentity, error) {
	i, e := r.ReadRootChoiceIndex(2)
	if e != nil {
		return CellIdentity{}, e
	}
	if i == 0 {
		v, e := DecodeEUTRACellIdentity(r)
		return NewEUTRACellIdentityChoice(v), e
	}
	v, e := DecodeUTRACellIdentity(r)
	return NewUTRACellIdentityChoice(v), e
}
func (v CellGlobalIdEUTRAAndUTRA) Validate() error {
	if e := v.p.Validate(); e != nil {
		return e
	}
	return v.c.Validate()
}
