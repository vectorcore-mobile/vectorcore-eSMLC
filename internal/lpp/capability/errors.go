package capability

import "errors"

var (
	ErrUnsupportedCriticalExtension = errors.New("lpp capability: unsupported critical extension")
	ErrUnsupportedExtension         = errors.New("lpp capability: unsupported extension addition")
	ErrUnsupportedCommon            = errors.New("lpp capability: unsupported common capabilities")
	ErrUnsupportedAGNSS             = errors.New("lpp capability: unsupported A-GNSS capabilities")
	ErrUnsupportedOTDOA             = errors.New("lpp capability: unsupported OTDOA capabilities")
	ErrUnsupportedEPDU              = errors.New("lpp capability: unsupported EPDU capabilities")
	ErrMissingECIDSupport           = errors.New("lpp capability: missing ECID measurement support")
	ErrInvalidECID                  = errors.New("lpp capability: invalid ECID capabilities")
)
