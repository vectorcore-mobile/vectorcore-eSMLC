// Package capability implements the bounded TS 37.355 R9 capability payload
// subset: the ECID request selector and ECID measurement-support BIT STRING.
// It owns capability critical-extension and R9 IE encoding, while internal/lpp
// owns the outer LPP message body. All unsupported families and extensions
// fail closed. It contains no method selection or positioning behavior.
package capability
