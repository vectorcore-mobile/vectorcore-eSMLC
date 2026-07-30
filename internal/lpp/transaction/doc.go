// Package transaction provides bounded, transport-neutral control state for
// the fixture-supported LPP envelope. A Store is deliberately scoped by its
// caller to one LPP peer: LPP transaction numbers are not globally unique.
//
// It neither encodes messages nor sends them. Callers supply time to Apply and
// Prune; the package starts no goroutines or timers. Only the empty Release 9
// capability, location-information, Abort, and Error wrappers supported by
// internal/lpp are represented.
package transaction
