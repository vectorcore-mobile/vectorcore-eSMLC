package uper

import "errors"

var (
	ErrUnexpectedEOF              = errors.New("uper: unexpected end of input")
	ErrInvalidBitCount            = errors.New("uper: invalid bit count")
	ErrInvalidRange               = errors.New("uper: invalid constrained range")
	ErrValueBelowLower            = errors.New("uper: value below constrained lower bound")
	ErrValueAboveUpper            = errors.New("uper: value above constrained upper bound")
	ErrDecodedOutOfRange          = errors.New("uper: decoded value outside constrained range")
	ErrInvalidEnumerated          = errors.New("uper: invalid enumerated index")
	ErrInvalidChoice              = errors.New("uper: invalid choice index")
	ErrInvalidBitmap              = errors.New("uper: invalid optional bitmap")
	ErrNonZeroPadding             = errors.New("uper: non-zero trailing padding")
	ErrUnconsumedBits             = errors.New("uper: unconsumed meaningful bits")
	ErrIntegerWidth               = errors.New("uper: unsupported integer width")
	ErrExtensionUnsupported       = errors.New("uper: extension additions are unsupported")
	ErrInvalidBitString           = errors.New("uper: invalid bit string")
	ErrInvalidSequenceOfBounds    = errors.New("uper: invalid SEQUENCE OF bounds")
	ErrSequenceOfCountBelow       = errors.New("uper: SEQUENCE OF count below minimum")
	ErrSequenceOfCountAbove       = errors.New("uper: SEQUENCE OF count above maximum")
	ErrSequenceOfEncodeCallback   = errors.New("uper: missing SEQUENCE OF encode callback")
	ErrSequenceOfDecodeCallback   = errors.New("uper: missing SEQUENCE OF decode callback")
	ErrSequenceOfElementEncode    = errors.New("uper: SEQUENCE OF element encode failure")
	ErrSequenceOfElementDecode    = errors.New("uper: SEQUENCE OF element decode failure")
	ErrSequenceOfTruncatedCount   = errors.New("uper: truncated SEQUENCE OF count")
	ErrSequenceOfTruncatedElement = errors.New("uper: truncated SEQUENCE OF element")
)
