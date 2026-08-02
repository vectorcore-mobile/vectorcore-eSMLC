package capability

import (
	"fmt"

	"github.com/vectorcore/esmlc/internal/uper"
)

// AGNSSRequestCapabilities is the bounded TS 37.355 root of
// A-GNSS-RequestCapabilities: all three mandatory root booleans.
type AGNSSRequestCapabilities struct {
	GNSSSupportListReq           bool
	AssistanceDataSupportListReq bool
	LocationVelocityTypesReq     bool
}

func (v AGNSSRequestCapabilities) Validate() error { return nil }
func (v AGNSSRequestCapabilities) EncodeUPER(w *uper.Writer) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteBoolean(v.GNSSSupportListReq); err != nil {
		return err
	}
	if err := w.WriteBoolean(v.AssistanceDataSupportListReq); err != nil {
		return err
	}
	return w.WriteBoolean(v.LocationVelocityTypesReq)
}
func DecodeAGNSSRequestCapabilities(r *uper.Reader) (AGNSSRequestCapabilities, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return AGNSSRequestCapabilities{}, err
	}
	if ext {
		return AGNSSRequestCapabilities{}, ErrUnsupportedExtension
	}
	supportList, err := r.ReadBoolean()
	if err != nil {
		return AGNSSRequestCapabilities{}, err
	}
	assistance, err := r.ReadBoolean()
	if err != nil {
		return AGNSSRequestCapabilities{}, err
	}
	velocity, err := r.ReadBoolean()
	if err != nil {
		return AGNSSRequestCapabilities{}, err
	}
	return AGNSSRequestCapabilities{GNSSSupportListReq: supportList, AssistanceDataSupportListReq: assistance, LocationVelocityTypesReq: velocity}, nil
}

// GNSSID is the bounded root TS 37.355 GNSS-ID (an extensible SEQUENCE
// wrapping an extensible root ENUMERATED). Only the five Release 9 root
// values are representable; bds and navic-v1610 are Release 14/16
// extension additions and are rejected on decode.
type GNSSID uint8

const (
	GNSSIDGPS GNSSID = iota
	GNSSIDSBAS
	GNSSIDQZSS
	GNSSIDGalileo
	GNSSIDGLONASS
)

func (v GNSSID) Validate() error {
	if v > GNSSIDGLONASS {
		return fmt.Errorf("%w: gnss-id %d", ErrInvalidAGNSS, v)
	}
	return nil
}
func (v GNSSID) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	return w.WriteRootEnumerated(uint64(v), 5)
}
func DecodeGNSSID(r *uper.Reader) (GNSSID, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return 0, err
	}
	if ext {
		return 0, ErrUnsupportedExtension
	}
	idExt, err := r.ReadExtensionPresent()
	if err != nil {
		return 0, err
	}
	if idExt {
		return 0, ErrUnsupportedExtension
	}
	id, err := r.ReadRootEnumerated(5)
	if err != nil {
		return 0, err
	}
	v := GNSSID(id)
	return v, v.Validate()
}

// GNSSSupportElement is the bounded root TS 37.355 GNSS-SupportElement.
// sbas-IDs and fta-MeasSupport are both conditional/OPTIONAL and are not
// represented here; a UE that includes either is rejected as unsupported
// rather than silently ignored.
type GNSSSupportElement struct {
	ID                         GNSSID
	Modes                      uper.BitString // PositioningModes.posModes
	Signals                    uper.BitString // GNSS-SignalIDs.gnss-SignalIDs, fixed 8 bits
	ADRSupport                 bool
	VelocityMeasurementSupport bool
}

// SupportsUEBased reports PositioningModes bit 1 (TS 37.355 calls MS-based
// mode "ue-based"). This implementation only ever requests ue-based mode.
func (v GNSSSupportElement) SupportsUEBased() bool { return bit(v.Modes, 1) }

func (v GNSSSupportElement) Validate() error {
	if err := v.ID.Validate(); err != nil {
		return err
	}
	if v.Modes.BitLen() < 1 || v.Modes.BitLen() > 8 {
		return fmt.Errorf("%w: agnss-Modes length %d", ErrInvalidAGNSS, v.Modes.BitLen())
	}
	if v.Signals.BitLen() != 8 {
		return fmt.Errorf("%w: gnss-Signals length %d", ErrInvalidAGNSS, v.Signals.BitLen())
	}
	return nil
}
func (v GNSSSupportElement) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{false, false}); err != nil {
		return err
	}
	if err := v.ID.EncodeUPER(w); err != nil {
		return err
	}
	if err := writePositioningModes(w, v.Modes); err != nil {
		return err
	}
	if err := writeGNSSSignalIDs(w, v.Signals); err != nil {
		return err
	}
	if err := w.WriteBoolean(v.ADRSupport); err != nil {
		return err
	}
	return w.WriteBoolean(v.VelocityMeasurementSupport)
}
func DecodeGNSSSupportElement(r *uper.Reader) (GNSSSupportElement, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return GNSSSupportElement{}, err
	}
	if ext {
		return GNSSSupportElement{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(2)
	if err != nil {
		return GNSSSupportElement{}, err
	}
	id, err := DecodeGNSSID(r)
	if err != nil {
		return GNSSSupportElement{}, err
	}
	if present[0] {
		return GNSSSupportElement{}, fmt.Errorf("%w: sbas-IDs unsupported", ErrInvalidAGNSS)
	}
	modes, err := readPositioningModes(r)
	if err != nil {
		return GNSSSupportElement{}, err
	}
	signals, err := readGNSSSignalIDs(r)
	if err != nil {
		return GNSSSupportElement{}, err
	}
	if present[1] {
		return GNSSSupportElement{}, fmt.Errorf("%w: fta-MeasSupport unsupported", ErrInvalidAGNSS)
	}
	adr, err := r.ReadBoolean()
	if err != nil {
		return GNSSSupportElement{}, err
	}
	velocity, err := r.ReadBoolean()
	if err != nil {
		return GNSSSupportElement{}, err
	}
	v := GNSSSupportElement{ID: id, Modes: modes, Signals: signals, ADRSupport: adr, VelocityMeasurementSupport: velocity}
	return v, v.Validate()
}

func writePositioningModes(w *uper.Writer, v uper.BitString) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	return w.WriteBitString(v, 1, 8)
}
func readPositioningModes(r *uper.Reader) (uper.BitString, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return uper.BitString{}, err
	}
	if ext {
		return uper.BitString{}, ErrUnsupportedExtension
	}
	return r.ReadBitString(1, 8)
}
func writeGNSSSignalIDs(w *uper.Writer, v uper.BitString) error {
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	return w.WriteBitString(v, 8, 8)
}
func readGNSSSignalIDs(r *uper.Reader) (uper.BitString, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return uper.BitString{}, err
	}
	if ext {
		return uper.BitString{}, ErrUnsupportedExtension
	}
	return r.ReadBitString(8, 8)
}

// AGNSSProvideCapabilities is the bounded root TS 37.355
// A-GNSS-ProvideCapabilities: only gnss-SupportList. assistanceDataSupportList,
// locationCoordinateTypes, and velocityTypes are all OPTIONAL in the root
// and are not represented; a UE that includes any of them is rejected.
type AGNSSProvideCapabilities struct {
	GNSSSupportList []GNSSSupportElement // 1..16
}

// SupportsGPSUEBased reports whether the list contains a GPS entry
// advertising ue-based (MS-based) mode support — the only capability this
// implementation's job-layer gating checks.
func (v AGNSSProvideCapabilities) SupportsGPSUEBased() bool {
	for _, e := range v.GNSSSupportList {
		if e.ID == GNSSIDGPS && e.SupportsUEBased() {
			return true
		}
	}
	return false
}
func (v AGNSSProvideCapabilities) Validate() error {
	if len(v.GNSSSupportList) < 1 || len(v.GNSSSupportList) > 16 {
		return fmt.Errorf("%w: gnss-SupportList count %d", ErrInvalidAGNSS, len(v.GNSSSupportList))
	}
	for i := range v.GNSSSupportList {
		if err := v.GNSSSupportList[i].Validate(); err != nil {
			return fmt.Errorf("%w: element %d: %w", ErrInvalidAGNSS, i, err)
		}
	}
	return nil
}
func (v AGNSSProvideCapabilities) EncodeUPER(w *uper.Writer) error {
	if err := v.Validate(); err != nil {
		return err
	}
	if err := w.WriteExtensionPresent(false); err != nil {
		return err
	}
	if err := w.WriteOptionalBitmap([]bool{true, false, false, false}); err != nil {
		return err
	}
	return w.WriteSequenceOf(len(v.GNSSSupportList), 1, 16, func(index int, writer *uper.Writer) error {
		return v.GNSSSupportList[index].EncodeUPER(writer)
	})
}
func DecodeAGNSSProvideCapabilities(r *uper.Reader) (AGNSSProvideCapabilities, error) {
	ext, err := r.ReadExtensionPresent()
	if err != nil {
		return AGNSSProvideCapabilities{}, err
	}
	if ext {
		return AGNSSProvideCapabilities{}, ErrUnsupportedExtension
	}
	present, err := r.ReadOptionalBitmap(4)
	if err != nil {
		return AGNSSProvideCapabilities{}, err
	}
	if present[1] || present[2] || present[3] {
		return AGNSSProvideCapabilities{}, fmt.Errorf("%w: assistanceDataSupportList/locationCoordinateTypes/velocityTypes unsupported", ErrInvalidAGNSS)
	}
	v := AGNSSProvideCapabilities{}
	if !present[0] {
		return v, nil
	}
	list := make([]GNSSSupportElement, 0, 16)
	_, err = r.ReadSequenceOf(1, 16, func(index int, reader *uper.Reader) error {
		x, err := DecodeGNSSSupportElement(reader)
		if err != nil {
			return err
		}
		list = append(list, x)
		return nil
	})
	if err != nil {
		return AGNSSProvideCapabilities{}, err
	}
	v.GNSSSupportList = list
	return v, v.Validate()
}
