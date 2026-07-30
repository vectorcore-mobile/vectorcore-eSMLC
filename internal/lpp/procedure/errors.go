package procedure

import "errors"

var (
	ErrInvalidConfig          = errors.New("lpp procedure: invalid configuration")
	ErrApplicationPending     = errors.New("lpp procedure: application result already pending")
	ErrApplicationNotPending  = errors.New("lpp procedure: application result not pending")
	ErrWrongResult            = errors.New("lpp procedure: wrong application result kind")
	ErrCapacity               = errors.New("lpp procedure: pending application capacity exhausted")
	ErrUnsupportedMessage     = errors.New("lpp procedure: unsupported message")
	ErrUnrequestedECID        = errors.New("lpp procedure: ECID capability was not requested")
	ErrInvalidLocationRequest = errors.New("lpp procedure: invalid ECID request-location payload")
	ErrInvalidLocationProvide = errors.New("lpp procedure: invalid ECID provide-location payload")
)
