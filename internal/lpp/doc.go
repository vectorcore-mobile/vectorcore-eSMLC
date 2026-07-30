// Package lpp implements a deliberately bounded TS 37.355 Release 16 LPP
// outer envelope. It supports the fixture-validated transaction fields and
// bounded Release 9 capability, ECID request-location, and root-only ECID
// provide-location payloads. A-GNSS, OTDOA, common location estimates,
// extensions, transport integration and positioning policy remain outside this
// codec package.
//
// Decoding requires the exact meaningful UPER bit length. All bit operations
// are delegated to internal/uper; this package has no runtime fixture or tool
// dependency.
package lpp
