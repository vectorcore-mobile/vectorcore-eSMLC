package transaction

import "errors"

var (
	ErrInvalidConfiguration = errors.New("lpp transaction: invalid configuration")
	ErrInvalidKey           = errors.New("lpp transaction: invalid key")
	ErrInvalidDirection     = errors.New("lpp transaction: invalid direction")
	ErrInvalidOwnership     = errors.New("lpp transaction: invalid ownership")
	ErrInvalidMessage       = errors.New("lpp transaction: invalid message")
	ErrNotFound             = errors.New("lpp transaction: not found")
	ErrDuplicateTransaction = errors.New("lpp transaction: transaction key retained")
	ErrCapacity             = errors.New("lpp transaction: active capacity exhausted")
	ErrNumberExhausted      = errors.New("lpp transaction: transaction numbers exhausted")
	ErrInvalidTransition    = errors.New("lpp transaction: invalid procedure transition")
	ErrSequenceConflict     = errors.New("lpp transaction: sequence conflict")
	ErrStaleSequence        = errors.New("lpp transaction: stale sequence")
	ErrAckNotPending        = errors.New("lpp transaction: acknowledgement not pending")
	ErrAckMismatch          = errors.New("lpp transaction: acknowledgement mismatch")
	ErrAckAlreadyPending    = errors.New("lpp transaction: acknowledgement already pending")
	ErrCompleted            = errors.New("lpp transaction: transaction completed")
	ErrAborted              = errors.New("lpp transaction: transaction aborted")
	ErrFailed               = errors.New("lpp transaction: transaction failed")
	ErrUnsupportedProcedure = errors.New("lpp transaction: unsupported procedure")
)
