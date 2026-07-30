package lpp

import "errors"

var (
	ErrInvalidMessage       = errors.New("lpp: invalid message model")
	ErrUnsupportedBody      = errors.New("lpp: unsupported message body")
	ErrUnsupportedExtension = errors.New("lpp: unsupported extension")
	ErrMalformed            = errors.New("lpp: malformed encoding")
	ErrAmbiguousPadding     = errors.New("lpp: ambiguous transport-octet padding")
)
