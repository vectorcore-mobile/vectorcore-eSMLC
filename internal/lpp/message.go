package lpp

import (
	"fmt"

	"github.com/vectorcore/esmlc/internal/lpp/capability"
	"github.com/vectorcore/esmlc/internal/lpp/location"
	"github.com/vectorcore/esmlc/internal/uper"
)

type Initiator uint8

const (
	InitiatorLocationServer Initiator = iota
	InitiatorTargetDevice
)

type TransactionID struct {
	Initiator         Initiator
	TransactionNumber uint8
}
type Acknowledgement struct {
	Requested bool
	Indicator *uint8
}

type BodyKind uint8

const (
	BodyRequestCapabilities BodyKind = iota
	BodyProvideCapabilities
	BodyRequestLocationInformation BodyKind = 4
	BodyProvideLocationInformation BodyKind = 5
	BodyAbort                      BodyKind = 6
	BodyError                      BodyKind = 7
)

// Body is a tagged union for the bounded Release 9 message paths supported by
// this codec. Typed payloads are currently available for capabilities and the
// ECID request/provide location-information branches.
type Body struct {
	Kind                       BodyKind
	RequestCapabilities        *capability.RequestCapabilitiesR9IEs
	ProvideCapabilities        *capability.ProvideCapabilitiesR9IEs
	RequestLocationInformation *location.RequestLocationInformationR9IEs
	ProvideLocationInformation *location.ProvideLocationInformationR9IEs
}
type Message struct {
	TransactionID   *TransactionID
	EndTransaction  bool
	SequenceNumber  *uint8
	Acknowledgement *Acknowledgement
	Body            *Body
}

func (m Message) Validate() error {
	if m.TransactionID != nil && m.TransactionID.Initiator != InitiatorLocationServer && m.TransactionID.Initiator != InitiatorTargetDevice {
		return fmt.Errorf("%w: initiator", ErrInvalidMessage)
	}
	if m.Body != nil && !supportedBody(m.Body.Kind) {
		return fmt.Errorf("%w: %d", ErrUnsupportedBody, m.Body.Kind)
	}
	if m.Body != nil {
		if m.Body.Kind != BodyRequestCapabilities && m.Body.RequestCapabilities != nil {
			return fmt.Errorf("%w: request capability payload on body %d", ErrInvalidMessage, m.Body.Kind)
		}
		if m.Body.Kind != BodyProvideCapabilities && m.Body.ProvideCapabilities != nil {
			return fmt.Errorf("%w: provide capability payload on body %d", ErrInvalidMessage, m.Body.Kind)
		}
		if m.Body.Kind != BodyRequestLocationInformation && m.Body.RequestLocationInformation != nil {
			return fmt.Errorf("%w: request location payload on body %d", ErrInvalidMessage, m.Body.Kind)
		}
		if m.Body.Kind != BodyProvideLocationInformation && m.Body.ProvideLocationInformation != nil {
			return fmt.Errorf("%w: provide location payload on body %d", ErrInvalidMessage, m.Body.Kind)
		}
		if m.Body.RequestCapabilities != nil {
			if err := m.Body.RequestCapabilities.Validate(); err != nil {
				return err
			}
		}
		if m.Body.ProvideCapabilities != nil {
			if err := m.Body.ProvideCapabilities.Validate(); err != nil {
				return err
			}
		}
		if m.Body.RequestLocationInformation != nil {
			if err := m.Body.RequestLocationInformation.Validate(); err != nil {
				return err
			}
		}
		if m.Body.ProvideLocationInformation != nil {
			if err := m.Body.ProvideLocationInformation.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}
func supportedBody(kind BodyKind) bool {
	return kind == BodyRequestCapabilities || kind == BodyProvideCapabilities || kind == BodyRequestLocationInformation || kind == BodyProvideLocationInformation || kind == BodyAbort || kind == BodyError
}

func EncodeMessage(m Message) (uper.Encoded, error) {
	if err := m.Validate(); err != nil {
		return uper.Encoded{}, err
	}
	w := uper.NewWriter()
	if err := w.WriteOptionalBitmap([]bool{m.TransactionID != nil, m.SequenceNumber != nil, m.Acknowledgement != nil, m.Body != nil}); err != nil {
		return uper.Encoded{}, err
	}
	if tx := m.TransactionID; tx != nil {
		if err := w.WriteExtensionPresent(false); err != nil {
			return uper.Encoded{}, err
		}
		if err := w.WriteExtensionPresent(false); err != nil {
			return uper.Encoded{}, err
		}
		if err := w.WriteRootEnumerated(uint64(tx.Initiator), 2); err != nil {
			return uper.Encoded{}, err
		}
		if err := w.WriteConstrainedWholeNumber(uint64(tx.TransactionNumber), 0, 255); err != nil {
			return uper.Encoded{}, err
		}
	}
	if err := w.WriteBoolean(m.EndTransaction); err != nil {
		return uper.Encoded{}, err
	}
	if m.SequenceNumber != nil {
		if err := w.WriteConstrainedWholeNumber(uint64(*m.SequenceNumber), 0, 255); err != nil {
			return uper.Encoded{}, err
		}
	}
	if a := m.Acknowledgement; a != nil {
		if err := w.WriteOptionalBitmap([]bool{a.Indicator != nil}); err != nil {
			return uper.Encoded{}, err
		}
		if err := w.WriteBoolean(a.Requested); err != nil {
			return uper.Encoded{}, err
		}
		if a.Indicator != nil {
			if err := w.WriteConstrainedWholeNumber(uint64(*a.Indicator), 0, 255); err != nil {
				return uper.Encoded{}, err
			}
		}
	}
	if m.Body != nil {
		if err := encodeBody(w, m.Body); err != nil {
			return uper.Encoded{}, err
		}
	}
	return w.Encoded(), nil
}

func DecodeMessage(data []byte, bitLength int) (Message, error) {
	r, err := uper.NewReader(data, bitLength)
	if err != nil {
		return Message{}, fmt.Errorf("%w: %w", ErrMalformed, err)
	}
	bits, err := r.ReadOptionalBitmap(4)
	if err != nil {
		return Message{}, fmt.Errorf("%w: envelope bitmap: %w", ErrMalformed, err)
	}
	m := Message{}
	if bits[0] {
		ext, err := r.ReadExtensionPresent()
		if err != nil {
			return m, err
		}
		if err = uper.RequireNoExtension(ext); err != nil {
			return m, fmt.Errorf("%w: transaction: %w", ErrUnsupportedExtension, err)
		}
		ext, err = r.ReadExtensionPresent()
		if err != nil {
			return m, err
		}
		if err = uper.RequireNoExtension(ext); err != nil {
			return m, fmt.Errorf("%w: initiator: %w", ErrUnsupportedExtension, err)
		}
		v, err := r.ReadRootEnumerated(2)
		if err != nil {
			return m, err
		}
		n, err := r.ReadConstrainedWholeNumber(0, 255)
		if err != nil {
			return m, err
		}
		m.TransactionID = &TransactionID{Initiator: Initiator(v), TransactionNumber: uint8(n)}
	}
	m.EndTransaction, err = r.ReadBoolean()
	if err != nil {
		return m, err
	}
	if bits[1] {
		v, e := r.ReadConstrainedWholeNumber(0, 255)
		if e != nil {
			return m, e
		}
		x := uint8(v)
		m.SequenceNumber = &x
	}
	if bits[2] {
		optional, e := r.ReadOptionalBitmap(1)
		if e != nil {
			return m, e
		}
		requested, e := r.ReadBoolean()
		if e != nil {
			return m, e
		}
		m.Acknowledgement = &Acknowledgement{Requested: requested}
		if optional[0] {
			v, e := r.ReadConstrainedWholeNumber(0, 255)
			if e != nil {
				return m, e
			}
			x := uint8(v)
			m.Acknowledgement.Indicator = &x
		}
	}
	if bits[3] {
		body, e := decodeBody(r)
		if e != nil {
			return m, e
		}
		m.Body = body
	}
	if err := r.ValidateFinalPadding(); err != nil {
		return m, fmt.Errorf("%w: %w", ErrMalformed, err)
	}
	return m, nil
}

// DecodeMessageOctets decodes an LPP APDU carried as octets, where the UPER
// meaningful-bit length is not carried separately. It accepts the unique
// meaningful length in the final octet that produces a fully valid message;
// ambiguous padding fails closed rather than guessing a different PDU.
func DecodeMessageOctets(data []byte) (Message, error) {
	if len(data) == 0 {
		return Message{}, ErrMalformed
	}
	var value Message
	found := false
	for pad := 0; pad < 8 && pad <= len(data)*8; pad++ {
		candidate, err := DecodeMessage(data, len(data)*8-pad)
		if err != nil {
			continue
		}
		if found {
			return Message{}, ErrAmbiguousPadding
		}
		value, found = candidate, true
	}
	if !found {
		return Message{}, ErrMalformed
	}
	return value, nil
}

func encodeBody(w *uper.Writer, body *Body) error {
	kind := body.Kind
	if !supportedBody(kind) {
		return ErrUnsupportedBody
	}
	if err := w.WriteRootChoiceIndex(0, 2); err != nil {
		return err
	}
	if err := w.WriteRootChoiceIndex(uint64(kind), 16); err != nil {
		return err
	}
	if kind == BodyRequestCapabilities {
		v := capability.RequestCapabilitiesR9IEs{}
		if body.RequestCapabilities != nil {
			v = *body.RequestCapabilities
		}
		return capability.EncodeRequestCapabilities(w, v)
	}
	if kind == BodyProvideCapabilities {
		v := capability.ProvideCapabilitiesR9IEs{}
		if body.ProvideCapabilities != nil {
			v = *body.ProvideCapabilities
		}
		return capability.EncodeProvideCapabilities(w, v)
	}
	if kind == BodyRequestLocationInformation {
		v := location.RequestLocationInformationR9IEs{}
		if body.RequestLocationInformation != nil {
			v = *body.RequestLocationInformation
		}
		return location.EncodeRequestLocationInformation(w, v)
	}
	if kind == BodyProvideLocationInformation {
		v := location.ProvideLocationInformationR9IEs{}
		if body.ProvideLocationInformation != nil {
			v = *body.ProvideLocationInformation
		}
		return location.EncodeProvideLocationInformation(w, v)
	}
	if kind == BodyError {
		if err := w.WriteRootChoiceIndex(0, 2); err != nil {
			return err
		}
		if err := w.WriteExtensionPresent(false); err != nil {
			return err
		}
		return w.WriteOptionalBitmap([]bool{false})
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
	count := 5
	if kind == BodyAbort {
		count = 1
	}
	return w.WriteOptionalBitmap(make([]bool, count))
}
func decodeBody(r *uper.Reader) (*Body, error) {
	outer, e := r.ReadRootChoiceIndex(2)
	if e != nil {
		return nil, e
	}
	if outer != 0 {
		return nil, ErrUnsupportedBody
	}
	v, e := r.ReadRootChoiceIndex(16)
	if e != nil {
		return nil, e
	}
	kind := BodyKind(v)
	if !supportedBody(kind) {
		return nil, ErrUnsupportedBody
	}
	if kind == BodyRequestCapabilities {
		payload, e := capability.DecodeRequestCapabilities(r)
		if e != nil {
			return nil, e
		}
		return &Body{Kind: kind, RequestCapabilities: &payload}, nil
	}
	if kind == BodyProvideCapabilities {
		payload, e := capability.DecodeProvideCapabilities(r)
		if e != nil {
			return nil, e
		}
		return &Body{Kind: kind, ProvideCapabilities: &payload}, nil
	}
	if kind == BodyRequestLocationInformation {
		payload, e := location.DecodeRequestLocationInformation(r)
		if e != nil {
			return nil, e
		}
		return &Body{Kind: kind, RequestLocationInformation: &payload}, nil
	}
	if kind == BodyProvideLocationInformation {
		payload, e := location.DecodeProvideLocationInformation(r)
		if e != nil {
			return nil, e
		}
		return &Body{Kind: kind, ProvideLocationInformation: &payload}, nil
	}
	if kind == BodyError {
		rel, e := r.ReadRootChoiceIndex(2)
		if e != nil {
			return nil, e
		}
		if rel != 0 {
			return nil, ErrUnsupportedExtension
		}
		ext, e := r.ReadExtensionPresent()
		if e != nil {
			return nil, e
		}
		if e = uper.RequireNoExtension(ext); e != nil {
			return nil, ErrUnsupportedExtension
		}
		bits, e := r.ReadOptionalBitmap(1)
		if e != nil {
			return nil, e
		}
		if bits[0] {
			return nil, ErrUnsupportedBody
		}
		return &Body{Kind: kind}, nil
	}
	critical, e := r.ReadRootChoiceIndex(2)
	if e != nil {
		return nil, e
	}
	if critical != 0 {
		return nil, ErrUnsupportedExtension
	}
	rel, e := r.ReadRootChoiceIndex(4)
	if e != nil {
		return nil, e
	}
	if rel != 0 {
		return nil, ErrUnsupportedExtension
	}
	ext, e := r.ReadExtensionPresent()
	if e != nil {
		return nil, e
	}
	if e = uper.RequireNoExtension(ext); e != nil {
		return nil, ErrUnsupportedExtension
	}
	count := 5
	if kind == BodyAbort {
		count = 1
	}
	bits, e := r.ReadOptionalBitmap(count)
	if e != nil {
		return nil, e
	}
	for _, b := range bits {
		if b {
			return nil, ErrUnsupportedBody
		}
	}
	return &Body{Kind: kind}, nil
}
