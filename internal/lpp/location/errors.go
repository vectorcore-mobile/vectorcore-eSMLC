package location

import "errors"

var (
	ErrUnsupportedCriticalExtension = errors.New("lpp location: unsupported critical extension")
	ErrUnsupportedExtension         = errors.New("lpp location: unsupported extension addition")
	ErrUnsupportedEPDU              = errors.New("lpp location: unsupported EPDU request")
	ErrMissingRequestedMeasurements = errors.New("lpp location: missing ECID requested measurements")
	ErrInvalidECIDRequest           = errors.New("lpp location: invalid ECID request")
	ErrInvalidECIDProvide           = errors.New("lpp location: invalid ECID provide")
	ErrMissingMeasuredResults       = errors.New("lpp location: missing ECID measured results")
	ErrUnsupportedECIDError         = errors.New("lpp location: unsupported ECID error")
	ErrInvalidOTDOAProvide          = errors.New("lpp location: invalid OTDOA provide")
	ErrMissingNeighbourMeasurements = errors.New("lpp location: missing OTDOA neighbour measurements")
	ErrInvalidOTDOAError            = errors.New("lpp location: invalid OTDOA error")
)
